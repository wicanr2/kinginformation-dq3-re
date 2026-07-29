package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// TestDumpDhamaReclass replaces the stale C-remake image with the current
// Ebitengine CTY17 runtime class menu and original success dialogue. Natural
// reachability is covered separately by TestOpeningProductionInputTrace.
func TestDumpDhamaReclass(t *testing.T) {
	if os.Getenv("DQ3_DUMP_DHAMA") == "" {
		t.Skip("設 DQ3_DUMP_DHAMA=1 才執行")
	}
	out := os.Getenv("DHAMA_OUT")
	if out == "" {
		t.Fatal("需設 DHAMA_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	assets := os.Getenv("DQ3_ASSETS")
	if assets == "" {
		assets = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(assets), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		17, mapBlkNum[17], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, 17, true
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
	}
	m := newMember([]int{110, 208, 210, 187}, 4, 0, stats.ExpForLevel(4, 20))
	m.Inventory = []int{0x4a}
	g.companions = []*Member{m}

	var priest *npcInst
	for i := range sc.npcs {
		n := &sc.npcs[i]
		if n.x == 10 && n.y == 10 && n.b4 == 39 {
			priest = n
			break
		}
	}
	if priest == nil {
		t.Fatal("CTY17 sec0 找不到 handler39 NPC")
	}
	g.px, g.py, g.facing = 10, 11, 1
	if !g.talkReclass(priest) {
		t.Fatal("handler39 未啟動轉職")
	}
	closeDialogue := func() {
		g.dlg.open = false
		g.advanceReclassDialogue()
	}
	closeDialogue()
	g.reclassChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeDialogue()
	g.reclassCursor = 1
	g.reclassChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeDialogue()
	g.reclassCursor = 5 // 原版 rec67 第六項賢者

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
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	dump("dhama_menu")

	g.reclassChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeDialogue()
	g.reclassChoiceInput(InputState{Confirm: true, DirEdge: -1})
	closeDialogue()
	g.reclassChoiceInput(InputState{Confirm: true, DirEdge: -1})
	dump("dhama_reclass_success")
}
