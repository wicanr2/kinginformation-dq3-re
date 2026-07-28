package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func mirrorEventGame(t *testing.T) *Game {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.dnPhase, g.dnStep = 2, 0
	g.setStoryFlag(0x42, true)
	g.setStoryFlag(0x10, false)
	g.setStoryFlag(0x21, true)
	g.setStoryFlag(0x22, false)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctySamanosa, mapBlkNum[ctySamanosa], mirrorSection, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatalf("load CTY44 sec1 night: %v", err)
	}
	g.town, g.cur, g.inTown, g.curCty = sc, sc, true, ctySamanosa
	g.px, g.py = mirrorX, mirrorY
	g.dlg.tx = sc.dlgText
	g.inventory = []int{0x61}
	return g
}

func closeCurrentDialogue(t *testing.T, g *Game) {
	t.Helper()
	stage := g.mirrorStage
	for i := 0; g.dlg.open && g.mirrorStage == stage && i < 16; i++ {
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
}

// 從合法事件 checkpoint 開始，所有玩家操作都走 production Game.step：
// 命令→道具→拉之鏡→rec97→rec98→怪力魔89。
func TestMirrorProductionInputTrace(t *testing.T) {
	g := mirrorEventGame(t)
	send := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	send(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	send(InputState{DirHeld: -1, DirEdge: 3})
	send(InputState{DirHeld: -1, DirEdge: 0})
	send(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	send(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})

	if g.panel != panelNone || g.mirrorStage != 1 || !g.dlg.open ||
		g.storyFlag(0x42) || !g.storyFlag(0x10) || !g.hasItem(0x61) {
		t.Fatalf("揭露交易錯: panel=%d stage=%d dlg=%v f42=%v f10=%v inv=%v",
			g.panel, g.mirrorStage, g.dlg.open, g.storyFlag(0x42), g.storyFlag(0x10), g.inventory)
	}
	closeCurrentDialogue(t, g)
	if g.mirrorStage != 2 || !g.dlg.open {
		t.Fatalf("rec97 關閉後應接 rec98: stage=%d dlg=%v", g.mirrorStage, g.dlg.open)
	}
	// 事件 checkpoint 使用足以穩定擊敗 boss 的角色數值；戰鬥指令本身仍全部經 Game.step。
	g.heroStat[stats.STR], g.heroStat[stats.VIT], g.heroStat[stats.AGI] = 1000, 1000, 1000
	g.heroStat[stats.HP], g.heroHP, g.heroInit = 999, 999, true
	closeCurrentDialogue(t, g)
	if g.mirrorStage != 3 || !g.battle.active || g.battle.monID != monsterBossTroll {
		t.Fatalf("rec98 後應進怪力魔89: stage=%d active=%v mon=%d",
			g.mirrorStage, g.battle.active, g.battle.monID)
	}
	for i := 0; g.battle.active && i < 24; i++ {
		send(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if g.battle.active || g.mirrorStage != 0 || !g.hasItem(itemModChangeStaff) ||
		g.storyFlag(0x42) || !g.storyFlag(0x22) || g.dnPhase != 0 || !g.dlg.open {
		t.Fatalf("正式戰鬥輸入後勝利交易錯: active=%v stage=%d inv=%v f42=%v f22=%v phase=%d dlg=%v",
			g.battle.active, g.mirrorStage, g.inventory, g.storyFlag(0x42),
			g.storyFlag(0x22), g.dnPhase, g.dlg.open)
	}
}

func TestMirrorBattleTransactions(t *testing.T) {
	t.Run("loss restores fake king", func(t *testing.T) {
		g := mirrorEventGame(t)
		g.useMirror()
		g.mirrorStage, g.battle.monID, g.battle.result = 3, monsterBossTroll, 3
		g.settleMirrorBattle()
		if !g.storyFlag(0x42) || g.storyFlag(0x10) || g.hasItem(itemModChangeStaff) || !g.isNight() {
			t.Fatalf("敗北回滾錯: f42=%v f10=%v staff=%v night=%v",
				g.storyFlag(0x42), g.storyFlag(0x10), g.hasItem(itemModChangeStaff), g.isNight())
		}
	})
	t.Run("win changes king and grants staff", func(t *testing.T) {
		g := mirrorEventGame(t)
		g.useMirror()
		g.mirrorStage, g.battle.monID, g.battle.result = 3, monsterBossTroll, 1
		g.settleMirrorBattle()
		if g.storyFlag(0x42) || !g.storyFlag(0x10) || g.storyFlag(0x21) ||
			!g.storyFlag(0x22) || !g.hasItem(itemModChangeStaff) || g.dnPhase != 0 || !g.dlg.open {
			t.Fatalf("勝利交易錯: f42=%v f10=%v f21=%v f22=%v staff=%v phase=%d dlg=%v",
				g.storyFlag(0x42), g.storyFlag(0x10), g.storyFlag(0x21), g.storyFlag(0x22),
				g.hasItem(itemModChangeStaff), g.dnPhase, g.dlg.open)
		}
	})
}

func TestMirrorWinSaveRestore(t *testing.T) {
	g := mirrorEventGame(t)
	g.useMirror()
	g.mirrorStage, g.battle.monID, g.battle.result = 3, monsterBossTroll, 1
	g.settleMirrorBattle()
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "samanosa-win.json"))
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}

	restored, err := NewGame(g.assets, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := restored.Load(); err != nil {
		t.Fatal(err)
	}
	if !restored.inTown || restored.curCty != ctySamanosa || restored.cur == nil ||
		restored.cur.sec != mirrorSection || restored.px != mirrorX || restored.py != mirrorY ||
		restored.dnPhase != 0 || restored.storyFlag(0x42) || !restored.storyFlag(0x22) ||
		!restored.hasItem(itemModChangeStaff) {
		t.Fatalf("讀檔未保留事件世界狀態: town=%v cty=%d sec=%d @%d,%d phase=%d f42=%v f22=%v inv=%v",
			restored.inTown, restored.curCty, sceneSection(restored.cur), restored.px, restored.py,
			restored.dnPhase, restored.storyFlag(0x42), restored.storyFlag(0x22), restored.inventory)
	}
}
