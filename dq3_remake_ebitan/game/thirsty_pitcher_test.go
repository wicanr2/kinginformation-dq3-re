package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func dumpThirstyPitcherFrame(t *testing.T, g *Game, path string) {
	t.Helper()
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	for i := 0; i < ScreenW*ScreenH; i++ {
		o := i * 4
		img.Set(i%ScreenW, i/ScreenW,
			color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
	}
	f, err := os.Create(path)
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

func TestDumpThirstyPitcherFinalKey(t *testing.T) {
	if os.Getenv("DQ3_DUMP_DRAIN") == "" {
		t.Skip("設 DQ3_DUMP_DRAIN=1 才執行")
	}
	out := os.Getenv("DRAIN_OUT")
	if out == "" {
		t.Fatal("需設 DRAIN_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, effect := thirstyPitcherGame(t)
	g.panel = panelNone
	g.shipX, g.shipY = g.px, g.py
	dumpThirstyPitcherFrame(t, g, filepath.Join(out, "thirsty_pitcher_before.png"))

	traceUseInventoryItem(t, g, effect.ItemRawID)
	dumpThirstyPitcherFrame(t, g, filepath.Join(out, "thirsty_pitcher_revealed.png"))
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if !g.inTown || g.curCty != 40 {
		t.Fatalf("未進 CTY40：town=%v cty=%d", g.inTown, g.curCty)
	}
	dumpThirstyPitcherFrame(t, g, filepath.Join(out, "final_key_shrine_before.png"))
	finalKey, ok := g.pack.TreasureEvent("dq3:event.shoal_final_key")
	if !ok {
		t.Fatal("缺最終鑰匙 treasure event")
	}
	traceExaminePackTreasure(t, g, finalKey.Treasure)
	dumpThirstyPitcherFrame(t, g, filepath.Join(out, "final_key_obtained.png"))
}
