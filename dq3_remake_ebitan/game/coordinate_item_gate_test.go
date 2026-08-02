package game

import (
	"reflect"
	"testing"
)

func oliviaCapeTestGame(t *testing.T) *Game {
	t.Helper()
	g, _ := trackedWorldObjectTestGame(t)
	g.inTown, g.layer = false, 0
	g.cur, g.px, g.py = g.overworldScene(), 76, 54
	g.shipOwned, g.shipAboard, g.shipX, g.shipY = true, true, 76, 54
	g.setStoryFlag(0x35, true)
	return g
}

func TestOliviaCapeMissingMemoryForcesFiveSteps(t *testing.T) {
	g := oliviaCapeTestGame(t)
	event, ok := g.pack.CoordinateItemGateAt(76, 54, 0)
	if !ok {
		t.Fatal("奧莉薇亞海岬 coordinate_item_gate missing")
	}
	beforePhase, beforeStep := g.dnPhase, g.dnStep
	if !g.tryCoordinateItemGateEvent() {
		t.Fatal("缺愛的回憶仍應觸發悲歌")
	}
	want, _ := g.pack.TextGlyphCodes(event.DialogueTextIDs.Approach)
	if !reflect.DeepEqual(g.dlg.buf, want) {
		t.Fatal("缺道具分支未顯示原版 record597")
	}
	g.dlg.open = false
	g.advanceCoordinateItemGateDialogue()
	if g.coordinateForcedSteps != 5 || g.coordinateForcedDir != 3 {
		t.Fatalf("強制水流=%d steps dir%d, want 5 dir3", g.coordinateForcedSteps, g.coordinateForcedDir)
	}
	if g.dnPhase != beforePhase || g.dnStep != beforeStep+20 {
		t.Fatalf("日夜步數=(%d,%d), want (%d,%d)", g.dnPhase, g.dnStep, beforePhase, beforeStep+20)
	}
	for g.coordinateForcedSteps > 0 {
		g.advanceCoordinateItemGateMovement()
	}
	if g.px != 81 || g.py != 54 || !g.storyFlag(0x35) {
		t.Fatalf("水流後=(%d,%d) flag35=%v, want (81,54) set", g.px, g.py, g.storyFlag(0x35))
	}
}

func TestOliviaCapeMemoryClearsCurseWithoutConsumption(t *testing.T) {
	g := oliviaCapeTestGame(t)
	g.inventory = []int{0x64}
	event, _ := g.pack.CoordinateItemGateAt(76, 54, 0)
	if !g.tryCoordinateItemGateEvent() {
		t.Fatal("持愛的回憶應觸發海岬事件")
	}
	g.dlg.open = false
	g.advanceCoordinateItemGateDialogue()
	want, _ := g.pack.TextGlyphCodes(event.DialogueTextIDs.Success)
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, want) {
		t.Fatal("成功分支未接續原版 record598")
	}
	if !g.storyFlag(0x35) {
		t.Fatal("record598 關閉前不可先提交旗標交易")
	}
	g.dlg.open = false
	g.advanceCoordinateItemGateDialogue()
	if g.storyFlag(0x35) || !g.hasItem(0x64) || g.px != 76 || g.py != 54 || g.coordinateForcedSteps != 0 {
		t.Fatalf("成功交易錯：flag35=%v item=%v pos=(%d,%d) forced=%d",
			g.storyFlag(0x35), g.hasItem(0x64), g.px, g.py, g.coordinateForcedSteps)
	}
	if g.tryCoordinateItemGateEvent() {
		t.Fatal("flag0x35 clear 後同座標不得重播")
	}
}

func TestGaiaSwordTreasureComesFromPack(t *testing.T) {
	g := oliviaCapeTestGame(t)
	if treasureFor(55, 0, 0) != nil {
		t.Fatal("CTY55 蓋亞之劍不得殘留在 Go treasure table")
	}
	found := false
	for _, event := range g.pack.TreasureEvents() {
		if event.Treasure.CTYRaw == 55 && event.Treasure.Section == 0 &&
			event.Treasure.TileSubID == 0 && event.Treasure.ItemRawID == 0x0f &&
			event.Treasure.PresentFlag == 0x48 {
			found = true
		}
	}
	if !found {
		t.Fatal("CTY55 蓋亞之劍 game-pack treasure missing")
	}
}
