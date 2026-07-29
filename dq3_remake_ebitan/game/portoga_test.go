package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestPortogaRoyalMissionMapReachability(t *testing.T) {
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	event := g.pack.StagedVehicleExchangeEvents()[0]
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, event.NPC.CTYRaw,
		mapBlkNum[event.NPC.CTYRaw], event.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	const entryX, entryY = 13, 29 // CTY16 sec0 transition (2,5) 的原始目的座標。
	if !traceNPCReachable(g, event.NPC.CTYRaw, event.NPC.Section,
		entryX, entryY, event.NPC.Tile.X, event.NPC.Tile.Y) {
		t.Fatalf("白天原始入口(%d,%d)無法抵達 Portoga 國王(%d,%d)相鄰格；map=%dx%d",
			entryX, entryY, event.NPC.Tile.X, event.NPC.Tile.Y, sc.w, sc.h)
	}
}

func portogaTestGame(t *testing.T) (*Game, gamepack.StagedVehicleExchangeEvent, *npcInst) {
	t.Helper()
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	events := pack.StagedVehicleExchangeEvents()
	if len(events) != 1 {
		t.Fatalf("staged vehicle exchange events=%d, want 1", len(events))
	}
	event := events[0]
	n := &npcInst{
		x: event.NPC.Tile.X, y: event.NPC.Tile.Y,
		ctrl: 2 << 3, b4: event.NPC.HandlerRaw,
	}
	g := &Game{
		pack: pack, flags: map[int]bool{}, curCty: event.NPC.CTYRaw,
		cur:    &Scene{sec: event.NPC.Section, npcs: []npcInst{*n}},
		inTown: true, noticeCode: -1,
	}
	g.setStoryFlag(event.QuestPresentFlagRaw, true)
	g.setStoryFlag(event.ExchangeAvailableFlagRaw, true)
	return g, event, &g.cur.npcs[0]
}

func packText(t *testing.T, g *Game, id string) []uint16 {
	t.Helper()
	codes, ok := g.pack.TextGlyphCodes(id)
	if !ok {
		t.Fatalf("找不到 pack text %q", id)
	}
	return codes
}

// TestPortogaRoyalMissionTransaction locks the original handler26 transaction:
// first visit record24 -> item0x5b + clear flag0x37 -> record25; later item0x5c
// is consumed to grant the ship, clear flag0x2c and berth at pack coordinates.
func TestPortogaRoyalMissionTransaction(t *testing.T) {
	g, event, king := portogaTestGame(t)

	if !g.talkStagedVehicleExchange(king) {
		t.Fatal("pack selector should claim Portoga king")
	}
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf,
		packText(t, g, event.DialogueTextIDs.QuestIntro)) {
		t.Fatal("首訪應先顯示原版任務 record24")
	}
	if g.hasItem(event.GrantedItemRawID) || !g.storyFlag(event.QuestPresentFlagRaw) {
		t.Fatal("任務介紹關閉前不得先給信或清旗標")
	}

	g.dlg.open = false
	g.advanceStagedVehicleExchangeDialogue()
	if !g.hasItem(event.GrantedItemRawID) || g.storyFlag(event.QuestPresentFlagRaw) {
		t.Fatal("record24 關閉後應給國王的信並 clear 原版 flag0x37")
	}
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf,
		packText(t, g, event.DialogueTextIDs.ItemGrant)) {
		t.Fatal("給信後應顯示原版 record25")
	}
	if g.noticeCode != event.GrantedItemRawID || g.noticeTimer == 0 {
		t.Fatal("給信應留下道具取得通知")
	}
	g.dlg.open = false
	g.advanceStagedVehicleExchangeDialogue()
	if g.vehicleExchangeStage != vehicleExchangeIdle || g.vehicleExchangeEventID != "" {
		t.Fatal("給信對話結束後 transient event state 應清空")
	}

	g.inventory = []int{event.GrantedItemRawID}
	g.talkStagedVehicleExchange(king)
	if g.shipOwned || !reflect.DeepEqual(g.dlg.buf,
		packText(t, g, event.DialogueTextIDs.NeedItem)) {
		t.Fatal("無交換道具時不授船，應顯示原版 record26")
	}

	g.inventory = append(g.inventory, event.RequiredItemRawID)
	g.setStoryFlag(event.Vehicle.ClearFlagsRaw[0], true)
	g.talkStagedVehicleExchange(king)
	if !g.shipOwned || g.shipAboard || g.hasItem(event.RequiredItemRawID) {
		t.Fatal("持交換道具時應消耗並授予停泊中的船")
	}
	want := event.Vehicle.WorldPosition
	if g.shipX != want.X || g.shipY != want.Y {
		t.Fatalf("船停泊=(%d,%d), want (%d,%d)", g.shipX, g.shipY, want.X, want.Y)
	}
	if g.storyFlag(event.Vehicle.ClearFlagsRaw[0]) || !g.progressDone(msShip) {
		t.Fatal("授船應 clear pack flag 並設定引擎取船里程碑")
	}
	if !reflect.DeepEqual(g.dlg.buf, packText(t, g, event.DialogueTextIDs.Success)) {
		t.Fatal("授船應顯示原版 record27")
	}

	g.talkStagedVehicleExchange(king)
	if !reflect.DeepEqual(g.dlg.buf, packText(t, g, event.DialogueTextIDs.After)) {
		t.Fatal("已取船後應顯示原版 record28")
	}
}

// TestSelectCommandPortogaRoyalMission verifies the formal command dispatch,
// including the pack selector. No CTY, coordinate or handler constant exists in
// production Go.
func TestSelectCommandPortogaRoyalMission(t *testing.T) {
	g, event, _ := portogaTestGame(t)
	g.px, g.py = event.NPC.Tile.X, event.NPC.Tile.Y+1
	g.facing = 1

	g.selectCommand(cmdTalk)
	if !g.dlg.open || g.vehicleExchangeStage != vehicleExchangeQuestIntro {
		t.Fatal("正式 cmdTalk 應命中 pack 的首訪任務")
	}
}

// TestScriptedTalkPepperGate retains the upstream quest precondition: the
// rescued-person milestone gates the original pepper giver.
func TestScriptedTalkPepperGate(t *testing.T) {
	const (
		pepperGiverCTY = 15
		pepperHandler  = 25
		pepperItem     = 0x5c
		rescuedFlag    = 0x211
	)
	g := &Game{flags: map[int]bool{}, curCty: pepperGiverCTY}

	g.scriptedTalk(pepperHandler)
	if g.hasItem(pepperItem) {
		t.Error("未救出達妮亞時不應給黑胡椒")
	}
	g.flags[rescuedFlag] = true
	g.scriptedTalk(pepperHandler)
	if !g.hasItem(pepperItem) {
		t.Error("救人後應給黑胡椒")
	}
}
