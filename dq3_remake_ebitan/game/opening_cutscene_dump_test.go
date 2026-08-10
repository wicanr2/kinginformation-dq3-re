package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpOpeningCutscene 以 production renderer 輸出五張開機過場卡片。
// 平常跳過；設 DQ3_DUMP_OPENING=1 與 OPENING_OUT=<gitignored 目錄> 才執行。
// 這個 dump 只證明 game-pack asset 已經由正常 opening state 消費，不把
// 120 幀停留或排序升格成原版逐幀 timing。
func TestDumpOpeningCutscene(t *testing.T) {
	if os.Getenv("DQ3_DUMP_OPENING") == "" {
		t.Skip("設 DQ3_DUMP_OPENING=1 才輸出開機過場 runtime 圖")
	}
	out := os.Getenv("OPENING_OUT")
	if out == "" {
		t.Fatal("DQ3_DUMP_OPENING 需要 OPENING_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.openingSeq == nil || len(g.openingSeq.Frames) == 0 {
		t.Fatal("game pack 沒有 opening sequence")
	}
	g.StartOpeningCutscene()
	if !g.openingActive {
		t.Fatal("production opening entry 未啟動")
	}
	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255})
		}
		path := filepath.Join(out, name+".png")
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
		t.Logf("%s dumped", path)
	}
	for i, frame := range g.openingSeq.Frames {
		if !g.openingActive || g.openingIndex != i {
			t.Fatalf("opening frame state active=%v index=%d want=%d", g.openingActive, g.openingIndex, i)
		}
		dump(frame.AssetKey)
		for n := 0; n < frame.HoldFrames; n++ {
			// -1 表示沒有方向 edge／held；InputState 的零值是「下」方向，
			// 不能拿來代表無輸入，否則會被 skip_on_input 正確中斷。
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatal(err)
			}
		}
	}
	if g.openingActive {
		t.Fatal("opening sequence should stop after final frame")
	}
}
