package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpTrackedWorldObject(t *testing.T) {
	if os.Getenv("DQ3_DUMP_TRACKED_WORLD_OBJECT") == "" {
		t.Skip("設 DQ3_DUMP_TRACKED_WORLD_OBJECT=1 才產出幽靈船核對圖")
	}
	out := os.Getenv("TRACKED_WORLD_OBJECT_OUT")
	if out == "" {
		t.Fatal("需設 TRACKED_WORLD_OBJECT_OUT")
	}
	g, objectID := trackedWorldObjectTestGame(t)
	obj, _ := g.pack.TrackedWorldObject(objectID)
	p := g.trackedWorldPositions[objectID]
	dump := func(name string) {
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

	g.showTitle, g.inTown, g.layer, g.curCty = false, false, p.Layer, -1
	g.cur, g.px, g.py = g.overworldScene(), p.X-1, p.Y
	g.shipOwned, g.shipAboard, g.shipX, g.shipY = true, true, g.px, g.py
	dump("ghost_ship_overworld")

	section, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, obj.EntranceCTYRaw,
		mapBlkNum[obj.EntranceCTYRaw], 1, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.inTown, g.shipAboard, g.curCty = true, false, obj.EntranceCTYRaw
	g.town, g.cur, g.px, g.py = section, section, 18, 56
	dump("ghost_ship_loves_memory_chest")
}
