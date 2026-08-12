package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
)

func kandarTestGame(t *testing.T) *Game {
	t.Helper()
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	event := g.pack.BossSurrenderEvents()[0]
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.Trigger.CTYRaw, mapBlkNum[event.Trigger.CTYRaw], event.Trigger.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, event.Trigger.CTYRaw
	g.dlg.tx = sc.dlgText
	g.px, g.py, g.facing = 6, 9, 1
	g.showTitle, g.openingIdx, g.cd = false, -1, 0
	return g
}

func TestKandarNaturalTileTriggerAndMixedFormation(t *testing.T) {
	g := kandarTestGame(t)
	event := g.pack.BossSurrenderEvents()[0]
	if !g.storyFlag(event.PresenceFlagRaw) {
		t.Fatal("新遊戲原始 story flag0x2e 應為 set，四名盜賊在場")
	}
	if err := g.step(InputState{DirHeld: 1, DirEdge: 1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 6 || g.py != 8 || g.bossSurrenderStage != bossSurrenderIntro || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Intro)) {
		t.Fatalf("踏入 (6,8) 未自然觸發 rec84：pos=(%d,%d) stage=%d dlg=%v",
			g.px, g.py, g.bossSurrenderStage, g.dlg.open)
	}
	for i := 0; i < 16 && g.dlg.open; i++ {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
			t.Fatal(err)
		}
	}
	if !g.battle.active || g.bossSurrenderStage != bossSurrenderBattle {
		t.Fatalf("rec84 後未開戰：active=%v stage=%d", g.battle.active, g.bossSurrenderStage)
	}
	want := []int{26, 27, 27, 27}
	if len(g.battle.enemies) != len(want) {
		t.Fatalf("敵數=%d, want 4", len(g.battle.enemies))
	}
	for i, id := range want {
		if g.battle.enemies[i].monID != id {
			t.Fatalf("敵%d id=%d, want %d", i, g.battle.enemies[i].monID, id)
		}
	}
}

func TestKandarCrownGateAndOriginalSurrenderLoop(t *testing.T) {
	g := kandarTestGame(t)
	event := g.pack.BossSurrenderEvents()[0]
	crownItem := event.TreasureGates[0].ItemRawID
	// 戰前即使直接站到寶箱旁，金皇冠也不得繞過 handler14。
	g.px, g.py, g.facing = 4, 5, 1
	g.examine()
	if g.hasItem(crownItem) {
		t.Fatal("甘達特仍在時不得取得金皇冠")
	}

	g.bossSurrenderEventID = event.ID
	g.bossSurrenderStage = bossSurrenderBattle
	g.battle.result, g.battle.gotExp, g.battle.gotGold = 1, 2440, 0
	g.onBattleEnd()
	if g.bossSurrenderStage != bossSurrenderApology || !g.dlg.open {
		t.Fatalf("勝利後應播放 rec85：stage=%d dlg=%v", g.bossSurrenderStage, g.dlg.open)
	}
	g.dlg.open = false
	g.advanceBossSurrenderDialogue()
	if g.bossSurrenderStage != bossSurrenderChoice {
		t.Fatalf("rec85 後應進求饒選擇，stage=%d", g.bossSurrenderStage)
	}
	g.bossSurrenderCursor = 1
	g.bossSurrenderChoiceInput(InputState{Confirm: true})
	if g.bossSurrenderStage != bossSurrenderReject ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Reject)) {
		t.Fatal("選否應播 rec87，再回到選擇")
	}
	g.dlg.open = false
	g.advanceBossSurrenderDialogue()
	g.bossSurrenderChoiceInput(InputState{Confirm: true})
	if g.bossSurrenderStage != bossSurrenderFarewell ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Accept)) {
		t.Fatal("選是應播 rec86")
	}
	g.dlg.open = false
	g.advanceBossSurrenderDialogue()
	if g.storyFlag(event.PresenceFlagRaw) || g.bossSurrenderStage != bossSurrenderIdle {
		t.Fatalf("rec86 後應 clear flag0x2e：flag=%v stage=%d",
			g.storyFlag(event.PresenceFlagRaw), g.bossSurrenderStage)
	}
	g.px, g.py, g.facing = 4, 5, 1
	g.examine()
	if !g.hasItem(crownItem) {
		t.Fatal("接受求饒後應可由 (4,4) 寶箱取得金皇冠")
	}
}

func mustPackTextCodes(t *testing.T, g *Game, id string) []uint16 {
	t.Helper()
	codes, ok := g.pack.TextGlyphCodes(id)
	if !ok {
		t.Fatalf("game pack 缺少文字 %q", id)
	}
	return codes
}

func TestKandarTowerProductionRouteOpensThiefKeyDoor(t *testing.T) {
	g := kandarTestGame(t)
	event := g.pack.BossSurrenderEvents()[0]
	// 此 fixture 只驗證已持盜賊鑰匙時的 CTY10 transition／開門路由；
	// 隨機遭遇與隊伍戰力已由完整 E3 主線 trace 覆蓋。以正式道具面板
	// 使用聖水，避免 Lv1 單人 fixture 偶遇 monster 23 而把門路由測試
	// 錯誤地變成戰鬥難度測試。
	g.inventory = []int{
		0x55,
		itemuse.ItemHolyWater, itemuse.ItemHolyWater, itemuse.ItemHolyWater,
		itemuse.ItemHolyWater, itemuse.ItemHolyWater,
	}
	g.enterTownCty(event.Trigger.CTYRaw)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	if g.repel != itemuse.HolySteps || g.countItem(itemuse.ItemHolyWater) != 4 {
		t.Fatalf("甘達特塔 route fixture 的正式聖水 transaction 錯：repel=%d inventory=%v",
			g.repel, g.inventory)
	}
	traceTownSectionTo(t, g, event.Trigger.CTYRaw, event.Trigger.Section)
	if sceneSection(g.cur) != event.Trigger.Section {
		t.Fatalf("正式開門路線未抵達 CTY10 sec%d：sec=%d",
			event.Trigger.Section, sceneSection(g.cur))
	}
}
