package game

import (
	"os"
	"testing"
)

func TestPiratesRedOrbPushEntranceAndSharedPresentFlag(t *testing.T) {
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	event, ok := g.pack.TreasureEvent("dq3:event.pirates_den_red_orb")
	if !ok {
		t.Fatal("缺 dq3:event.pirates_den_red_orb")
	}
	tr := event.Treasure
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		tr.CTYRaw, mapBlkNum[tr.CTYRaw], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	const entranceX, entranceY = 26, 9
	idx := sc.npcAt(entranceX, entranceY)
	rule := g.pack.NPCPushRule()
	if idx < 0 || sc.npcs[idx].ctrl&rule.CtrlMask != rule.CtrlValue {
		t.Fatalf("密道入口物件不是 pack 規則的可推物件：idx=%d", idx)
	}
	dcty, dsec, dx, dy, ok := sc.tileTransition(entranceX, entranceY)
	if !ok || dcty != tr.CTYRaw || dsec != tr.Section || dx != 5 || dy != 9 {
		t.Fatalf("密道 transition 不符：ok=%v -> CTY%d sec%d (%d,%d)", ok, dcty, dsec, dx, dy)
	}

	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, tr.CTYRaw
	g.showTitle, g.openingIdx = false, -1
	g.px, g.py, g.facing, g.cd = entranceX, entranceY+1, 1, 0
	object := &sc.npcs[idx]
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if object.x != entranceX || object.y != entranceY-1 ||
		g.curCty != tr.CTYRaw || sceneSection(g.cur) != tr.Section || g.px != dx || g.py != dy {
		t.Fatalf("推物件進密道錯：object=(%d,%d) cty=%d sec=%d pos=(%d,%d)",
			object.x, object.y, g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	if !g.collectPackTreasure(tr) || !g.hasItem(tr.ItemRawID) || g.storyFlag(tr.PresentFlag) {
		t.Fatalf("紅寶珠交易錯：item=%v flag=%v", g.hasItem(tr.ItemRawID), g.storyFlag(tr.PresentFlag))
	}
	reloaded, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		tr.CTYRaw, mapBlkNum[tr.CTYRaw], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.npcAt(entranceX, entranceY) >= 0 || reloaded.npcAt(entranceX, entranceY-1) >= 0 {
		t.Fatal("event0 清除的 flag0x3f 必須同時隱藏重載後的入口物件")
	}
}
