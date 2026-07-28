package game

import (
	"reflect"
	"testing"
)

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

func TestZomaHiddenStairNaturalExamine(t *testing.T) {
	g := r4Game(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyZomaCastle,
		mapBlkNum[ctyZomaCastle], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyZomaCastle
	g.dlg.tx = sc.dlgText

	const x, y = 23, 3
	ev, subid, ok := sc.tileEvent(x, y)
	if !ok || subid != 3 || ev != [3]int{4, 0, 0xe5} {
		t.Fatalf("CTY90 sec0 隱藏樓梯原始事件錯：ok=%v subid=%d ev=%v", ok, subid, ev)
	}
	before := sc.tileIdx(x, y)
	g.setStoryFlag(0xe5, true)
	g.px, g.py, g.facing = x, y+1, 1
	chooseExamine(t, g)

	if g.storyFlag(0xe5) {
		t.Fatal("sub_18C01 應 CLEAR flag e5")
	}
	if sc.tileIdx(x, y) != (before+1)&0xff || sc.hiMap[y*sc.w+x]&0x1f != 0 {
		t.Fatalf("樓梯 tile mutation 錯：%02x→%02x hi=%02x",
			before, sc.tileIdx(x, y), sc.hiMap[y*sc.w+x])
	}
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, g.shop.nameText.Record(484)) {
		t.Fatal("應顯示 D3TXT00 rec484「發現一座隱藏樓梯」")
	}

	// 清掉訊息後踩上剛顯露的樓梯；清除 subid 後使用 transition[0]。
	for g.dlg.open {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	g.px, g.py = x, y
	g.tryTransition()
	if g.cur.sec != 1 || g.px != 16 || g.py != 8 {
		t.Fatalf("顯露樓梯未依原始 transition[0] 進 sec1@(16,8)：sec=%d @(%d,%d)",
			g.cur.sec, g.px, g.py)
	}
}

func TestOrtegaBridgeEventNaturalMovement(t *testing.T) {
	g := r4Game(t)
	g.setStoryFlag(ortegaFightFlag, true)
	g.setStoryFlag(ortegaDyingFlag, false)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyZomaCastle,
		mapBlkNum[ctyZomaCastle], ortegaSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyZomaCastle
	g.dlg.tx = sc.dlgText
	g.px, g.py, g.cd = ortegaTriggerX+1, ortegaTriggerY, 0

	testStep(t, g, InputState{DirHeld: 2, DirEdge: 2}) // 從橋右側向左踏入原版 trigger
	if g.px != ortegaTriggerX || g.py != ortegaTriggerY {
		t.Fatalf("未踏到歐魯迪卡 trigger：@(%d,%d)", g.px, g.py)
	}
	if g.ortegaStage != 1 || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, sc.dlgText.Record(ortegaFightRec)) {
		t.Fatalf("應自動開 handler79 rec69：stage=%d open=%v", g.ortegaStage, g.dlg.open)
	}
	if g.battle.active {
		t.Fatal("歐魯迪卡是固定演出，不可開玩家通常戰鬥 UI")
	}

	for g.ortegaStage == 1 {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if g.ortegaStage != 2 || g.storyFlag(ortegaFightFlag) ||
		!g.storyFlag(ortegaDyingFlag) || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, g.cur.dlgText.Record(ortegaDyingRec)) {
		t.Fatalf("handler79→90 flags/rec70 錯：stage=%d e0=%v e=%v open=%v",
			g.ortegaStage, g.storyFlag(ortegaFightFlag), g.storyFlag(ortegaDyingFlag), g.dlg.open)
	}
	dying := 0
	for i := range g.cur.npcs {
		if g.cur.npcs[i].b4 == 90 {
			dying++
		}
	}
	if dying != 1 {
		t.Fatalf("SET flag0e 後應只載入一個 handler90 歐魯迪卡，得 %d", dying)
	}

	for g.dlg.open {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if g.ortegaStage != 0 || g.storyFlag(ortegaDyingFlag) {
		t.Fatalf("handler90 結束應 CLEAR flag0e：stage=%d flag=%v",
			g.ortegaStage, g.storyFlag(ortegaDyingFlag))
	}
	for i := range g.cur.npcs {
		if g.cur.npcs[i].b4 == 90 {
			t.Fatal("臨終對白後不應殘留歐魯迪卡 sprite")
		}
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
