package game

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpOpeningArrivalFrames is an opt-in visual evidence helper. It renders
// every pack-owned arrival tile with the current production renderer so a
// recorded original frame can be matched to a route index without guessing a
// scene, camera, or coordinate in production code.
func TestDumpOpeningArrivalFrames(t *testing.T) {
	out := os.Getenv("DQ3_DUMP_OPENING_ARRIVAL_DIR")
	if out == "" {
		t.Skip("set DQ3_DUMP_OPENING_ARRIVAL_DIR to dump opening arrival frames")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	if !g.finishMotherEscort() {
		t.Fatal("cannot load opening arrival scene")
	}
	for i, frame := range g.openingEscort.ArrivalFrames {
		g.applyOpeningArrivalFrame(frame)
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for p := 0; p < ScreenW*ScreenH; p++ {
			o := p * 4
			img.Set(p%ScreenW, p/ScreenW, color.RGBA{
				R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255,
			})
		}
		name := filepath.Join(out, fmt.Sprintf("arrival-%02d-%02d-%02d.png", i, frame.Player.X, frame.Player.Y))
		f, err := os.Create(name)
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
}
