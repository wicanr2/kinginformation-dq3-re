package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpPiratesRedOrb records current runtime evidence for the CTY27
// object-covered entrance and the pack-owned red-orb transaction. Full
// reachability is independently covered by TestOpeningProductionInputTrace.
func TestDumpPiratesRedOrb(t *testing.T) {
	if os.Getenv("DQ3_DUMP_RED_ORB") == "" {
		t.Skip("設 DQ3_DUMP_RED_ORB=1 才執行")
	}
	out := os.Getenv("RED_ORB_OUT")
	if out == "" {
		t.Fatal("需設 RED_ORB_OUT")
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
	g.showTitle, g.openingIdx = false, -1
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
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, tr.CTYRaw
	g.px, g.py, g.facing = 26, 10, 1

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

	dump("pirates_red_orb_hidden_entrance")
	g.cd = 0
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	found := false
	for y := 0; y < g.cur.h && !found; y++ {
		for x := 0; x < g.cur.w && !found; x++ {
			ev, subid, ok := g.cur.tileEvent(x, y)
			if ok && subid == tr.TileSubID && ev[0] == tr.EventTypeRaw &&
				ev[1] == tr.ItemRawID && ev[2] == tr.PresentFlag {
				g.px, g.py, found = x, y, true
			}
		}
	}
	if sceneSection(g.cur) != tr.Section || !found || !g.collectPackTreasure(tr) {
		t.Fatal("runtime fixture 未經可推入口進入紅寶珠區")
	}
	g.cur.revealEventTile(g.px, g.py)
	dump("pirates_red_orb_obtained")
}
