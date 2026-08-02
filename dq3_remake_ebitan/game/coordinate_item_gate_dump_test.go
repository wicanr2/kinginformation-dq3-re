package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpOliviaCapeAndGaiaSword(t *testing.T) {
	if os.Getenv("DQ3_DUMP_OLIVIA_GAIA") == "" {
		t.Skip("設 DQ3_DUMP_OLIVIA_GAIA=1 才產出海岬／蓋亞之劍核對圖")
	}
	out := os.Getenv("OLIVIA_GAIA_OUT")
	if out == "" {
		t.Fatal("需設 OLIVIA_GAIA_OUT")
	}
	dump := func(g *Game, name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{
				R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255,
			})
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

	g := oliviaCapeTestGame(t)
	g.showTitle = false
	g.inventory = []int{0x64}
	if !g.tryCoordinateItemGateEvent() {
		t.Fatal("無法開啟奧莉薇亞海岬事件")
	}
	g.dlg.open = false
	g.advanceCoordinateItemGateDialogue()
	dump(g, "olivia_cape_memory_reunion")

	section, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 55,
		mapBlkNum[55], 0, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.setStoryFlag(0x48, true)
	g.inTown, g.shipAboard, g.curCty = true, false, 55
	g.town, g.cur, g.px, g.py = section, section, 14, 22
	g.dlg.open = false
	g.examine()
	dump(g, "gaia_sword_obtained")
}
