package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// TestDumpNewGameScreens:把新遊戲流程各畫面(主選單/命名/性別)dump 成 PNG,
// 供與 docs/36_shots 原版實機截圖肉眼對照(glyph 形近誤標高風險 → 視覺驗證,規則 63)。
// 平常 skip;設 DQ3_DUMP_NG=1 + NG_OUT=<gitignored 夾> 才跑。
func TestDumpNewGameScreens(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NG") == "" {
		t.Skip("設 DQ3_DUMP_NG=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("NG_OUT")
	if out == "" {
		t.Fatal("需設 NG_OUT")
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
	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(out + "/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped", name)
	}
	// 依新遊戲狀態機逐畫面 dump(直接設 state,不跑 ebiten 主迴圈)
	g.newGame.stage = ngMenu
	dump("ng_menu")
	g.newGame.stage = ngName
	dump("ng_name")
	g.newGame.stage = ngGender
	dump("ng_gender")
}
