package game

import "testing"

func TestDragonQueenOriginalFlags(t *testing.T) {
	g := r4Game(t)
	g.curCty = ctyDragonQueen
	g.setStoryFlag(0x4e, true)
	g.setStoryFlag(0x19, false)
	g.scriptedTalk(dragonQueenHandler)
	if !g.hasItem(itemLightOrb) {
		t.Fatal("龍女王應給光之珠0x65")
	}
	if g.storyFlag(0x4e) || !g.storyFlag(0x19) {
		t.Fatalf("sub_15E02 flags 不符：flag4e=%v flag19=%v", g.storyFlag(0x4e), g.storyFlag(0x19))
	}
	h := r4Game(t)
	h.restore(g.snapshot())
	if h.storyFlag(0x4e) || !h.storyFlag(0x19) || !h.hasItem(itemLightOrb) {
		t.Fatal("光之珠與龍女王原版 flags 存檔 round-trip 失敗")
	}
}

func TestZomaNaturalTalkEntry(t *testing.T) {
	g := r4Game(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyZomaCastle,
		mapBlkNum[ctyZomaCastle], zomaFinalSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyZomaCastle
	g.dlg.tx = sc.dlgText
	var z *npcInst
	for i := range sc.npcs {
		if sc.npcs[i].b4 == zomaHandler {
			z = &sc.npcs[i]
			break
		}
	}
	if z == nil {
		t.Fatal("CTY90 sec5 找不到原版 handler80 索瑪 NPC")
	}
	g.px, g.py, g.facing = z.x, z.y+1, 1
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 開命令窗
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 話す
	if !g.battle.active || g.battle.monID != 0x7a {
		t.Fatalf("自然交談應從 formation0x7a 開始：active=%v id=%02x", g.battle.active, g.battle.monID)
	}
}

func TestZomaSection4PlainTileTransitionsToFinalFloor(t *testing.T) {
	g := r4Game(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyZomaCastle,
		mapBlkNum[ctyZomaCastle], 4, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	// CTY90 sec4 (19,29) 是 attr==0、hiMap subid1；原版資料目的為 sec5@(12,55)。
	dc, ds, dx, dy, ok := sc.tileTransition(19, 29)
	if !ok || dc != ctyZomaCastle || ds != zomaFinalSection || dx != 12 || dy != 55 {
		t.Fatalf("sec4 plain transition=%v %d.%d@(%d,%d)", ok, dc, ds, dx, dy)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyZomaCastle
	g.px, g.py = 19, 29
	g.tryTransition()
	if g.cur == nil || g.cur.sec != zomaFinalSection || g.px != 12 || g.py != 55 {
		t.Fatalf("正式 tryTransition 未進最終層：sec=%d @(%d,%d)", g.cur.sec, g.px, g.py)
	}
}

func TestRainbowBridgeSaveRoundTrip(t *testing.T) {
	g := r4Game(t)
	g.layer, g.inTown, g.cur = 1, false, g.loadUnder()
	g.px, g.py = rainbowUseX, rainbowUseY
	g.inventory = []int{itemRainbowDrop}
	g.panelCursor = 0
	g.useSelectedItem()

	h := r4Game(t)
	h.restore(g.snapshot())
	if h.worldState&worldStateRainbowBridge == 0 ||
		h.loadUnder().tileIdx(rainbowBridgeX, rainbowBridgeY) != rainbowBridgeTile {
		t.Fatal("彩虹橋 world state / tile override 存檔 round-trip 失敗")
	}
}
