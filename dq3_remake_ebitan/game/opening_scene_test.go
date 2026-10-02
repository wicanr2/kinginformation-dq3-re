package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestOpeningScenePresentationNormalInput(t *testing.T) {
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "opening-scene-save.json"))
	g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	g.prng.Seed(0x1357)
	for _, scan := range []int{0x1c, 0x1c, 0x48, 0x4b, 0x1c, 0x1c, 0x4d, 0x50, 0x1c, 0x48, 0x4b, 0x1c, 0x48, 0x1c, 0x1c, 0x1c, 0x1c} {
		in := InputState{DirHeld: -1, DirEdge: -1}
		switch scan {
		case 0x1c:
			in.Confirm = true
		case 0x48:
			in.DirEdge = 1
		case 0x50:
			in.DirEdge = 0
		case 0x4b:
			in.DirEdge = 2
		case 0x4d:
			in.DirEdge = 3
		}
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	idle := InputState{DirHeld: -1, DirEdge: -1}
	wait := func(done func() bool) {
		t.Helper()
		for i := 0; i < 2000 && !done(); i++ {
			if err := g.step(idle); err != nil {
				t.Fatal(err)
			}
		}
		if !done() {
			t.Fatal("正常輸入等待逾限")
		}
	}
	wait(func() bool { return g.dlg.waitingForConfirm() })
	if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
		t.Fatal(err)
	}
	wait(func() bool { return g.openingIdx == 1 && g.dlg.waitingForConfirm() })
	e := g.activeOpeningScenePresentation()
	if e == nil || g.dlg.shadow == nil || g.prng.State() != 0x356d || !reflect.DeepEqual(g.dlg.buf, g.dlg.tx.Record(83)) {
		t.Fatal("房間正式入口未套用已審查資料")
	}
	before := g.dlg.retained
	if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if before != g.dlg.retained || g.dlg.retained.waiting {
		t.Fatal("房間確認重建或清除前文")
	}
	wait(func() bool { return g.openingIdx == 2 })
	if !reflect.DeepEqual(g.dlg.buf, g.dlg.tx.Record(81)) || g.dlg.retained != nil {
		t.Fatal("房間EOF未自動返回下一段，或保留上一段畫布")
	}
	wait(func() bool { return g.openingEscortAnimating() })
	if g.dlg.prelude != nil || g.dlg.shadow != nil || g.activeOpeningScenePresentation() != nil {
		t.Fatal("共用呈現洩漏到帶路流程")
	}
	t.Log("正常創角、第18／19次確認：record83保留前文與EOF捲動，record81自動返回帶路；無場景注入")
}

func TestWindowShadowUsesSecondByteLatch(t *testing.T) {
	rgba := make([]byte, ScreenW*ScreenH*4)
	for y := 0; y < 2; y++ {
		for x := 0; x < 16; x++ {
			o := (y*ScreenW + x) * 4
			rgba[o], rgba[o+3] = uint8(1+x+y*16), 255
		}
	}
	original := append([]byte(nil), rgba...)
	drawWindowWordLatchShadow(rgba, gamepack.WindowLayout{Width: 16, Height: 2}, gamepack.WindowShadow{Mode: "vga_word_latch_and"}, dq3data.Color{})
	for y := 0; y < 2; y++ {
		for x := 0; x < 16; x++ {
			want := uint8(0)
			if (x+y)&1 != 0 {
				want = original[(y*ScreenW+8+x%8)*4]
			}
			if got := rgba[(y*ScreenW+x)*4]; got != want {
				t.Fatalf("(%d,%d)=%d want%d", x, y, got, want)
			}
		}
	}
}
