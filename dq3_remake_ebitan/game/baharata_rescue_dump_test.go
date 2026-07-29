package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpBaharataRescue 產出真實 CTY、game-pack 事件與 production renderer 組合的
// 視覺證據。正式玩家可達性另由 TestOpeningProductionInputTrace 驗證。
func TestDumpBaharataRescue(t *testing.T) {
	if os.Getenv("DQ3_DUMP_BAHARATA") == "" {
		t.Skip("設 DQ3_DUMP_BAHARATA=1 才執行")
	}
	out := os.Getenv("BAHARATA_OUT")
	if out == "" {
		t.Fatal("需設 BAHARATA_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	g, event := baharataTestGame(t)
	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	finishBattle := func() {
		t.Helper()
		g.battle.result, g.battle.phase = 1, phEnd
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
		if g.battle.active {
			t.Fatal("runtime dump 戰鬥未經正式 Confirm 關閉")
		}
	}

	guard := findRescueNPC(t, g, event.GuardNPCs[0])
	g.px, g.py, g.facing = guard.x, guard.y-1, 0
	if !g.talkHostageRescue(guard) || !g.dlg.open {
		t.Fatal("守衛問話 runtime 圖未觸發")
	}
	dump("baharata_guard_question")

	g.dlg.open = false
	g.advanceHostageRescueDialogue()
	g.hostageRescueChoiceInput(InputState{Cancel: true, DirEdge: -1})
	closeRescueDialogue(g)
	finishBattle()

	trigger := event.ApproachTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryHostageRescueTrigger() {
		t.Fatal("handler59 runtime 圖前置未觸發")
	}
	closeRescueDialogue(g)
	runRescueAnimation(t, g)

	trigger = event.SwitchTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryHostageRescueTrigger() {
		t.Fatal("handler60 runtime 圖前置未觸發")
	}
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeRescueDialogue(g) // switch_pressed → reunion
	if !g.dlg.open || g.hostageRescueStage != hostageRescueReunion {
		t.Fatal("俘虜重逢 runtime 圖未進入 reunion")
	}
	dump("baharata_captive_reunion")
	closeRescueDialogue(g)
	closeRescueDialogue(g)
	runRescueAnimation(t, g)
	closeRescueDialogue(g)

	boss := findRescueNPC(t, g, event.BossNPCs[0])
	if !g.talkHostageRescue(boss) {
		t.Fatal("boss runtime 圖未觸發 challenge")
	}
	closeRescueDialogue(g)
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatal("boss runtime 圖未啟動原版四敵 formation")
	}
	dump("baharata_boss_formation")

	finishBattle()
	closeRescueDialogue(g)
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeRescueDialogue(g)

	pepperScene, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.RewardNPC.CTYRaw, mapBlkNum[event.RewardNPC.CTYRaw],
		event.RewardNPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.inTown, g.curCty = true, event.RewardNPC.CTYRaw
	g.town, g.cur, g.dlg.tx = pepperScene, pepperScene, pepperScene.dlgText
	pepper := findRescueNPC(t, g, event.RewardNPC)
	g.px, g.py, g.facing = pepper.x, pepper.y+1, 1
	if !g.talkHostageRescue(pepper) || !g.dlg.open {
		t.Fatal("古布達 runtime 圖未觸發")
	}
	closeRescueDialogue(g)
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	if !g.hasItem(event.RewardItemRawID) || !g.dlg.open {
		t.Fatal("黑胡椒 runtime 圖未完成交易")
	}
	dump("baharata_pepper_received")
}
