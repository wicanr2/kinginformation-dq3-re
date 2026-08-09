package game

import (
	"os"
	"testing"
)

func r4Game(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	return g
}

func closeDialogueByInput(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; g.dlg.open && i < 32; i++ {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if g.dlg.open {
		t.Fatal("對話 32 次 A 後仍未關閉")
	}
}

func TestBaramosOriginalDataIdentity(t *testing.T) {
	tr := bossTriggerAt(65, 0, 8, 3)
	if tr == nil || len(tr.monsters) != 1 || tr.monsters[0] != 0x79 ||
		tr.doneFlag != 0x213 || tr.clearStoryFlag != 0x29 || tr.preRec != 85 || tr.postRec != 86 {
		t.Fatalf("巴拉摩斯 trigger 與 EXE/CTY 不一致: %+v", tr)
	}
	g := r4Game(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 65, mapBlkNum[65], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	if sc.w != 17 || sc.h != 20 {
		t.Fatalf("CTY65 應為 17x20，得 %dx%d", sc.w, sc.h)
	}
	idx := sc.npcAt(8, 3)
	if idx < 0 || (sc.npcs[idx].ctrl>>3)&7 != 2 || sc.npcs[idx].b4 != 70 {
		t.Fatalf("CTY65 (8,3) 應為 sub2 handler70，idx=%d", idx)
	}
}

// 從正式命令選單「話す」進入巴拉摩斯，經 preRec→單場戰鬥→postRec；
// 事件入口與戰鬥關閉均只送 production InputState。
func TestBaramosProductionInputTrace(t *testing.T) {
	g := r4Game(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 65, mapBlkNum[65], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty, g.dlg.tx = sc, sc, true, 65, sc.dlgText
	g.px, g.py, g.facing = 8, 4, 1

	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 開命令
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 話す
	if !g.dlg.open || !g.bossIntro || g.battle.active {
		t.Fatalf("rec85 前置狀態錯: dlg=%v intro=%v battle=%v", g.dlg.open, g.bossIntro, g.battle.active)
	}
	closeDialogueByInput(t, g)
	if !g.battle.active || g.battle.monID != 0x79 || len(g.battle.enemies) != 1 {
		t.Fatalf("應開始單隻怪0x79，active=%v id=%x n=%d", g.battle.active, g.battle.monID, len(g.battle.enemies))
	}

	g.battle.enemies[0].hp, g.battle.enemies[0].def = 1, 0
	g.battle.enemies[0].atk, g.battle.enemies[0].agi = 0, 0
	g.battle.heroAtk, g.battle.heroAgi = 999, 999
	for n := 0; g.battle.result == 0 && n < 16; n++ { // 命令→目標，逐一收完隊員才結算
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	for i := 0; i < 8 && g.battle.active; i++ {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 逐筆確認原版戰鬥／獎勵訊息
	}
	if g.battle.active || !g.flags[0x213] || g.storyFlag(0x29) || !g.dlg.open {
		t.Fatalf("巴拉摩斯勝利狀態錯:battle=%v result=%d f213=%v f29=%v postDlg=%v",
			g.battle.active, g.battle.result, g.flags[0x213], g.storyFlag(0x29), g.dlg.open)
	}
	closeDialogueByInput(t, g)
}

func TestBaramosReturnAndNaturalDescentInputTrace(t *testing.T) {
	g := r4Game(t)
	g.flags[0x213] = true
	g.setStoryFlag(0x29, false)

	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty, g.dlg.tx = throne, throne, true, ctyAliahanCastle, throne.dlgText
	g.px, g.py, g.facing, g.cd = aliahanKingX, aliahanKingY+2, 1, 0
	testStep(t, g, InputState{DirHeld: 1, DirEdge: -1}) // 走到國王正前方
	if g.baramosReturn != 1 || !g.dlg.open {
		t.Fatalf("應觸發國王 rec98，stage=%d dlg=%v @(%d,%d) blocked=%v",
			g.baramosReturn, g.dlg.open, g.px, g.py, throne.Blocked(aliahanKingX, aliahanKingY+1))
	}
	closeDialogueByInput(t, g) // rec98 關閉時自動開 rec99；helper 也把 rec99 關完
	if g.storyFlag(0x4d) || g.baramosReturn != 0 {
		t.Fatalf("索瑪現身後應 clear flag4d，flag=%v stage=%d", g.storyFlag(0x4d), g.baramosReturn)
	}
	saved := g.snapshot()
	restored := r4Game(t)
	restored.restore(saved)
	g = restored
	if !g.flags[0x213] || g.storyFlag(0x29) || g.storyFlag(0x4d) {
		t.Fatalf("R-4 save round-trip 錯:f213=%v f29=%v f4d=%v",
			g.flags[0x213], g.storyFlag(0x29), g.storyFlag(0x4d))
	}

	// 從地表走入 (54,129)：同點應選地震後 CTY72，不再是封閉 CTY71。
	g.cur, g.inTown, g.curCty, g.layer = g.over, false, -1, 0
	g.px, g.py, g.cd = 54, 130, 0
	testStep(t, g, InputState{DirHeld: 1, DirEdge: -1})
	if !g.inTown || g.curCty != 72 {
		t.Fatalf("flag4d 後蓋亞洞窟應載 CTY72，cty=%d inTown=%v", g.curCty, g.inTown)
	}

	// CTY72 普通可走 tile + hiMap subid1 是跳坑；踏入後自然跨到 CTY77。
	g.px, g.py, g.cd = 14, 6, 0
	testStep(t, g, InputState{DirHeld: 0, DirEdge: -1})
	if g.curCty != 77 || g.cur.sec != 0 || g.px != 16 || g.py != 9 {
		t.Fatalf("CTY72 跳坑應到 CTY77@(16,9)，cty=%d sec=%d @(%d,%d)",
			g.curCty, g.cur.sec, g.px, g.py)
	}

	// CTY77 的原版 0xfe 出口自然切 DQ3UND，並落在 (85,67)。
	g.px, g.py, g.cd = 31, 13, 0
	testStep(t, g, InputState{DirHeld: 3, DirEdge: -1})
	if g.inTown || g.layer != 1 || g.px != 85 || g.py != 67 || !g.progressDone(msDescend) {
		t.Fatalf("CTY77 0xfe 應自然下降:inTown=%v layer=%d @(%d,%d) progress=%v",
			g.inTown, g.layer, g.px, g.py, g.progressDone(msDescend))
	}
}
