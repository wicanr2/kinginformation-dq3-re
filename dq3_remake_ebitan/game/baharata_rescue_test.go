package game

import (
	"encoding/binary"
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestBaharataRescueMatchesOriginalEXECTYAndText(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	event, ok := pack.HostageRescueEvent("dq3:event.baharata_hostage_rescue")
	if !ok {
		t.Fatal("missing Baharata hostage rescue event")
	}
	if event.GuardPresentFlagRaw != 0x80 ||
		event.ApproachPendingFlagRaw != 0x81 ||
		event.CaptivePresentFlagRaw != 0x82 ||
		event.BossPresentFlagRaw != 0x14 ||
		!reflect.DeepEqual(event.CompletionSetFlagsRaw, []int{0x25}) ||
		!reflect.DeepEqual(event.CompletionClearFlagsRaw, []int{0x14, 0x34}) ||
		event.RewardAvailableFlagRaw != 0x36 || event.RewardItemRawID != 0x5c {
		t.Fatalf("原版旗標／黑胡椒交易不符：%+v", event)
	}
	if event.GuardFormation.RawBytesHex != "0118013904" ||
		event.BossFormation.RawBytesHex != "0418013801390139013901" {
		t.Fatalf("原版 formation bytes 不符：guard=%s boss=%s",
			event.GuardFormation.RawBytesHex, event.BossFormation.RawBytesHex)
	}

	cty14 := asset(t, "CTY14.DAT")
	town, err := dq3data.OpenTown(cty14, event.ApproachTrigger.Section, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(town.SpecialHandlers, []int{59, 60}) {
		t.Fatalf("CTY14 sec1 handlers=%v, want [59 60]", town.SpecialHandlers)
	}
	for _, trigger := range []gamepack.SceneTileTrigger{
		event.ApproachTrigger, event.SwitchTrigger,
	} {
		for _, tile := range trigger.Tiles {
			if got := int(town.HiMap[tile.Y*town.W+tile.X] & 0x1f); got != trigger.TileSubID {
				t.Fatalf("CTY14 tile(%d,%d) subid=%d, want %d",
					tile.X, tile.Y, got, trigger.TileSubID)
			}
		}
	}
	for _, selector := range append(append([]gamepack.ScriptedNPCSelector{},
		event.GuardNPCs...), event.BossNPCs...) {
		found := false
		for _, npc := range town.NPCs {
			if npc.X == selector.Tile.X && npc.Y == selector.Tile.Y &&
				npc.B4 == selector.HandlerRaw {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("CTY14 sec1 missing raw NPC selector %+v", selector)
		}
	}

	exe := asset(t, "DQ3.EXE")
	const dgroupFileBase = 0x16140
	for off, want := range map[int][]byte{
		dgroupFileBase + 0x4ec0: {0x01, 0x18, 0x01, 0x39, 0x04},
		dgroupFileBase + 0x4ec5: {0x04, 0x18, 0x01, 0x38, 0x01, 0x39, 0x01, 0x39, 0x01, 0x39, 0x01},
		dgroupFileBase + 0x3d27: {0x02, 0x00, 0x00, 0x02, 0x03, 0x00, 0xff},
		dgroupFileBase + 0x3d2e: {0x03, 0x03, 0x00, 0xff},
		dgroupFileBase + 0x3d32: {0x02, 0x00, 0x00, 0x05, 0x01, 0x00, 0xff},
	} {
		if !reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x=%x, want %x", off, exe[off:off+len(want)], want)
		}
	}

	txt := asset(t, "D3TXT03.TXT")
	textRecords := map[string]int{
		event.DialogueTextIDs.GuardQuestion:  106,
		event.DialogueTextIDs.GuardJoin:      107,
		event.DialogueTextIDs.GuardFight:     108,
		event.DialogueTextIDs.RescueCry:      109,
		event.DialogueTextIDs.SwitchPrompt:   86,
		event.DialogueTextIDs.SwitchPressed:  87,
		event.DialogueTextIDs.Reunion:        110,
		event.DialogueTextIDs.ReturnHome:     111,
		event.DialogueTextIDs.BossArrival:    112,
		event.DialogueTextIDs.BossChallenge:  113,
		event.DialogueTextIDs.BossApology:    114,
		event.DialogueTextIDs.BossReject:     115,
		event.DialogueTextIDs.BossFarewell:   116,
		event.DialogueTextIDs.RewardOffer:    118,
		event.DialogueTextIDs.RewardDecline:  119,
		event.DialogueTextIDs.RewardReceived: 120,
		event.DialogueTextIDs.BossDefeated:   124,
	}
	for id, record := range textRecords {
		start := int(binary.LittleEndian.Uint16(txt[record*2:]))
		end := int(binary.LittleEndian.Uint16(txt[(record+1)*2:]))
		raw := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			raw = append(raw, binary.LittleEndian.Uint16(txt[off:off+2]))
		}
		got, found := pack.TextGlyphCodes(id)
		if !found || !reflect.DeepEqual(got, raw) {
			t.Fatalf("%s does not match D3TXT03 rec%d", id, record)
		}
	}
}

func baharataTestGame(t *testing.T) (*Game, gamepack.HostageRescueEvent) {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	events := g.pack.HostageRescueEvents()
	if len(events) != 1 {
		t.Fatalf("hostage rescue events=%d, want 1", len(events))
	}
	event := events[0]
	g.setStoryFlag(event.GuardPresentFlagRaw, true)
	g.setStoryFlag(event.ApproachPendingFlagRaw, true)
	g.setStoryFlag(event.CaptivePresentFlagRaw, true)
	g.setStoryFlag(event.BossPresentFlagRaw, false)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.ApproachTrigger.CTYRaw, mapBlkNum[event.ApproachTrigger.CTYRaw],
		event.ApproachTrigger.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	g.inTown, g.curCty = true, event.ApproachTrigger.CTYRaw
	g.town, g.cur = sc, sc
	g.dlg.tx = sc.dlgText
	if len(tracePortalPath(sc, 26, 3, 14, 12, 1)) != 0 ||
		len(tracePortalPath(sc, 26, 3, 14, 12, 2)) == 0 {
		t.Fatal("CTY14 central room must require the original magic-key tier")
	}
	return g, event
}

func findRescueNPC(t *testing.T, g *Game, selector gamepack.ScriptedNPCSelector) *npcInst {
	t.Helper()
	for i := range g.cur.npcs {
		n := &g.cur.npcs[i]
		if n.x == selector.Tile.X && n.y == selector.Tile.Y &&
			n.b4 == selector.HandlerRaw {
			return n
		}
	}
	t.Fatalf("runtime NPC not found: %+v", selector)
	return nil
}

func closeRescueDialogue(g *Game) {
	g.dlg.open = false
	g.advanceHostageRescueDialogue()
}

func runRescueAnimation(t *testing.T, g *Game) {
	t.Helper()
	for frames := 0; frames < 3000 && g.hostageRescueAnimating(); frames++ {
		g.advanceHostageRescueAnimation()
	}
	if g.hostageRescueAnimating() {
		t.Fatal("hostage rescue animation did not finish")
	}
}

func TestBaharataRescueTransaction(t *testing.T) {
	g, event := baharataTestGame(t)
	guard := findRescueNPC(t, g, event.GuardNPCs[0])
	if !g.talkHostageRescue(guard) ||
		g.hostageRescueStage != hostageRescueGuardQuestion {
		t.Fatal("handler58 must open the guard question")
	}
	g.dlg.open = false
	g.advanceHostageRescueDialogue()
	g.hostageRescueChoiceInput(InputState{Cancel: true, DirEdge: -1})
	closeRescueDialogue(g)
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatalf("guard formation enemies=%d, want four", len(g.battle.enemies))
	}
	for _, enemy := range g.battle.enemies {
		if enemy.monID != 0x39 {
			t.Fatalf("guard monster=%#x, want 0x39", enemy.monID)
		}
	}
	g.battle.result = 1
	g.onBattleEnd()
	if g.storyFlag(event.GuardPresentFlagRaw) {
		t.Fatal("guard victory must clear original flag0x80")
	}

	trigger := event.ApproachTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryHostageRescueTrigger() {
		t.Fatal("normal movement onto handler59 tile must start rescue cry")
	}
	closeRescueDialogue(g)
	runRescueAnimation(t, g)
	if g.storyFlag(event.ApproachPendingFlagRaw) {
		t.Fatal("handler59 movement completion must clear flag0x81")
	}

	trigger = event.SwitchTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryHostageRescueTrigger() {
		t.Fatal("normal movement onto handler60 tile must open switch prompt")
	}
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeRescueDialogue(g)
	closeRescueDialogue(g)
	closeRescueDialogue(g)
	runRescueAnimation(t, g)
	if g.hostageRescueStage != hostageRescueBossArrival {
		t.Fatalf("captive movement stage=%d, want boss arrival", g.hostageRescueStage)
	}
	closeRescueDialogue(g)
	if g.storyFlag(event.CaptivePresentFlagRaw) ||
		!g.storyFlag(event.BossPresentFlagRaw) {
		t.Fatal("switch sequence must clear flag0x82 and set flag0x14")
	}

	boss := findRescueNPC(t, g, event.BossNPCs[0])
	if !g.talkHostageRescue(boss) {
		t.Fatal("handler61 returning boss must be claimed by pack")
	}
	closeRescueDialogue(g)
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatalf("boss formation enemies=%d, want four groups", len(g.battle.enemies))
	}
	wantMonsters := []int{0x38, 0x39, 0x39, 0x39}
	for i, enemy := range g.battle.enemies {
		if enemy.monID != wantMonsters[i] {
			t.Fatalf("boss formation enemy%d=%#x, want %#x", i, enemy.monID, wantMonsters[i])
		}
	}
	g.battle.result = 1
	g.onBattleEnd()
	closeRescueDialogue(g)
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Cancel: true, DirEdge: -1})
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeRescueDialogue(g)
	if g.storyFlag(event.BossPresentFlagRaw) ||
		g.storyFlag(0x34) || !g.storyFlag(0x25) {
		t.Fatal("accepted surrender must clear flags0x14/0x34 and set flag0x25")
	}
	if !g.storyFlag(event.RewardAvailableFlagRaw) {
		t.Fatal("rescue transaction must preserve initial one-time pepper flag0x36")
	}
}
