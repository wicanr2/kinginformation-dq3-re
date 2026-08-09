package game

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func attractFixture(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.attractSeq == nil || len(g.attractSeq.Frames) != 8 {
		t.Fatalf("DQ3 pack 應有 8 張職業 attract 卡，得 %#v", g.attractSeq)
	}
	return g
}

func TestAttractIdleLoopAndInputInterrupt(t *testing.T) {
	g := attractFixture(t)
	seq := *g.attractSeq
	seq.StartDelayFrames = 2
	seq.Frames = append([]gamepack.AttractFrame(nil), seq.Frames...)
	seq.Frames[0].HoldFrames = 2
	seq.Frames[1].HoldFrames = 2
	g.attractSeq = &seq
	g.showTitle = true
	g.newGame.stage = ngSplash
	g.attractActive = false

	noInput := InputState{DirHeld: -1, DirEdge: -1}
	if err := g.step(noInput); err != nil {
		t.Fatal(err)
	}
	if g.attractActive || g.titleIdleFrames != 1 {
		t.Fatalf("第一個閒置幀不應啟動 attract：active=%v idle=%d", g.attractActive, g.titleIdleFrames)
	}
	if err := g.step(noInput); err != nil {
		t.Fatal(err)
	}
	if !g.attractActive || g.attractIndex != 0 || g.attractFrame != 0 {
		t.Fatalf("到達 pack delay 後應顯示第一張 attract：active=%v idx=%d frame=%d",
			g.attractActive, g.attractIndex, g.attractFrame)
	}
	if err := g.step(noInput); err != nil {
		t.Fatal(err)
	}
	if g.attractIndex != 0 || g.attractFrame != 1 {
		t.Fatalf("第一張卡停留幀未依 pack 推進：idx=%d frame=%d", g.attractIndex, g.attractFrame)
	}
	if err := g.step(noInput); err != nil {
		t.Fatal(err)
	}
	if g.attractIndex != 1 || g.attractFrame != 0 {
		t.Fatalf("第一張卡結束後未進第二張：idx=%d frame=%d", g.attractIndex, g.attractFrame)
	}

	if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if g.attractActive || g.newGame.stage != ngMenu {
		t.Fatalf("玩家確認鍵應中斷 attract 並開主選單：active=%v stage=%d", g.attractActive, g.newGame.stage)
	}
}

func TestDQ3AttractPackAssets(t *testing.T) {
	g := attractFixture(t)
	for i, frame := range g.attractSeq.Frames {
		if frame.AssetKey == "" || frame.HoldFrames <= 0 || len(g.attractPix[i]) != ScreenW*ScreenH {
			t.Fatalf("attract frame[%d] 未載入完整 640x350 PCX：%+v pix=%d",
				i, frame, len(g.attractPix[i]))
		}
	}
}

func TestNewGameConfirmationBackground(t *testing.T) {
	g := attractFixture(t)
	if len(g.newGameConfirmPix) != ScreenW*ScreenH || len(g.newGameConfirmPal) != 16 {
		t.Fatalf("FIRST.SCR confirmation background not loaded: pix=%d pal=%d",
			len(g.newGameConfirmPix), len(g.newGameConfirmPal))
	}
	g.newGame.stage = ngConfirm
	g.renderFrame()
	if got := g.rgba[0:4]; got[0] != 0 || got[1] != 0 || got[2] != 0 || got[3] != 255 {
		t.Fatalf("confirmation background first pixel=%v, want opaque black", got)
	}
	// Stable region coverage from FIRST.SCR proves that rendering is not the old
	// empty black canvas; the palette itself is owned by the pack contract.
	nonzero := 0
	for y := 0; y < ScreenH; y++ {
		for x := ScreenW / 2; x < ScreenW; x++ {
			o := (y*ScreenW + x) * 4
			if g.rgba[o] != 0 || g.rgba[o+1] != 0 || g.rgba[o+2] != 0 {
				nonzero++
			}
		}
	}
	if nonzero < 1000 {
		t.Fatalf("confirmation background figure region remained mostly black: nonzero=%d", nonzero)
	}
}

func TestDumpAttractPNG(t *testing.T) {
	if os.Getenv("DQ3_DUMP_ATTRACT") == "" {
		t.Skip("設 DQ3_DUMP_ATTRACT=1 才輸出 attract 對拍圖")
	}
	out := os.Getenv("ATTRACT_OUT")
	if out == "" {
		t.Fatal("DQ3_DUMP_ATTRACT 需要 ATTRACT_OUT")
	}
	g := attractFixture(t)
	g.startAttract()
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	copy(img.Pix, g.rgba)
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(out, "title_attract_warrior.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}
