package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// newGameConfirmV3Fixture 建立與原版正式輸入截圖 verify_open_10_story_01 相同的
// 靜態確認內容。它只用於 renderer 的同狀態視覺比對：原版截圖本身由正常創角
// 操作取得，而 remake 的數值在這裡固定，避免 RNG 差異污染版面／glyph 對拍。
// 它不是用來取代 TestOpeningProductionInputTrace 的正式玩家流程證據。
func newGameConfirmV3Fixture(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.newGame.stage = ngConfirm
	g.newGame.ni.nameBuf = []int{0}
	g.newGame.preview[stats.STR] = 8
	g.newGame.preview[stats.VIT] = 4
	g.newGame.preview[stats.AGI] = 4
	g.newGame.preview[stats.HP] = 12
	g.newGame.preview[stats.MP] = 9
	g.newGame.preview[stats.INT] = 7
	g.newGame.preview[stats.LUCK] = 8
	g.newGame.previewDef = 6
	g.newGame.previewGender = 0
	return g
}

func TestNewGameConfirmationBuildsTenOriginalRightStatRowsBeforeChoiceOverlay(t *testing.T) {
	g := newGameConfirmV3Fixture(t)
	if g.newGame.labels == nil || g.newGame.geometry == nil || g.newGame.geometry.Stats == nil || g.newGame.geometry.ConfirmChoiceFrame == nil {
		t.Fatal("production new-game pack is missing labels, geometry, or confirmation frame")
	}
	// 原版 choice window 會蓋住前四列的一部分；先直接驗證其下方完整
	// record407 stat pass，避免把正確的 compositing 遮罩誤當成漏畫 glyph。
	geo := *g.newGame.geometry
	geo.ConfirmChoiceBackdrop = nil
	geo.ConfirmChoiceFrame = nil
	geo.ConfirmChoice = gamepack.GeometryRect{}
	geo.WindowBackdrops = nil
	rgba := make([]byte, ScreenW*ScreenH*4)
	drawNewGameStats(rgba, g.dlg.tx, &g.newGame, frameColor(geo.Frame.BorderRGB), dq3data.Color{}, &geo)

	statsLayout := g.newGame.geometry.Stats
	rows := []struct {
		label []int
		x, y  int
	}{
		{g.newGame.labels.Strength, statsLayout.Strength.Label.X, statsLayout.Strength.Label.Y},
		{g.newGame.labels.Agility, statsLayout.Agility.Label.X, statsLayout.Agility.Label.Y},
		{g.newGame.labels.Vitality, statsLayout.Vitality.Label.X, statsLayout.Vitality.Label.Y},
		{g.newGame.labels.Intelligence, statsLayout.Intelligence.Label.X, statsLayout.Intelligence.Label.Y},
		{g.newGame.labels.Luck, statsLayout.Luck.Label.X, statsLayout.Luck.Label.Y},
		{g.newGame.labels.MaxHP, statsLayout.MaxHP.Label.X, statsLayout.MaxHP.Label.Y},
		{g.newGame.labels.MaxMP, statsLayout.MaxMP.Label.X, statsLayout.MaxMP.Label.Y},
		{g.newGame.labels.Attack, statsLayout.Attack.Label.X, statsLayout.Attack.Label.Y},
		{g.newGame.labels.Defense, statsLayout.Defense.Label.X, statsLayout.Defense.Label.Y},
		{g.newGame.labels.Experience, statsLayout.Experience.Label.X, statsLayout.Experience.Label.Y},
	}
	textColor := frameColor(g.newGame.geometry.ConfirmChoiceFrame.BorderRGB)
	for i, row := range rows {
		if len(row.label) == 0 {
			t.Fatalf("stat row %d label is empty", i)
		}
		seen := false
		for y := row.y; y < row.y+dq3data.GlyphPx && !seen; y++ {
			for x := row.x; x < row.x+len(row.label)*dq3data.GlyphPx; x++ {
				o := (y*ScreenW + x) * 4
				if rgba[o] == textColor.R && rgba[o+1] == textColor.G && rgba[o+2] == textColor.B {
					seen = true
					break
				}
			}
		}
		if !seen {
			t.Fatalf("stat row %d (%d,%d) was not rendered", i, row.x, row.y)
		}
	}
}

func TestNewGameWindowBackdropUsesRawEGAByteProjection(t *testing.T) {
	rgba := make([]byte, ScreenW*ScreenH*4)
	for i := range rgba {
		rgba[i] = 0x7f
	}
	geo := &gamepack.NewGameGeometry{
		RawWindows: []gamepack.RawNewGameWindow{{ID: "window", X: 2, Y: 3, Width: 3, Height: 2}},
		WindowBackdrops: []gamepack.NewGameWindowBackdrop{{
			ID: "test", RawWindowID: "window", Primitive: "ega_window_black_backdrop",
		}},
	}
	drawNewGameWindowBackdrops(rgba, geo, 0)
	for y := 3; y < 5; y++ {
		for x := 16; x < 40; x++ {
			o := (y*ScreenW + x) * 4
			if rgba[o] != 0 || rgba[o+1] != 0 || rgba[o+2] != 0 || rgba[o+3] != 255 {
				t.Fatalf("raw EGA backdrop pixel (%d,%d)=%v, want opaque black", x, y, rgba[o:o+4])
			}
		}
	}
	if rgba[(2*ScreenW+16)*4] != 0x7f || rgba[(3*ScreenW+15)*4] != 0x7f || rgba[(3*ScreenW+40)*4] != 0x7f {
		t.Fatal("raw EGA backdrop modified a pixel outside its byte-projected rectangle")
	}
}

// TestDumpNewGameConfirmV3StaticFixture 輸出 640×350 的 remake 同狀態畫面，供 Docker
// 外的影像差分容器與原版正式輸入截圖比較。正常 test 不寫任何檔案。
func TestDumpNewGameConfirmV3StaticFixture(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NG_V3") == "" {
		t.Skip("設 DQ3_DUMP_NG_V3=1 才執行 V3 靜態畫面 dump")
	}
	out := os.Getenv("NG_V3_OUT")
	if out == "" {
		t.Fatal("需設 NG_V3_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g := newGameConfirmV3Fixture(t)
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	for i := 0; i < ScreenW*ScreenH; i++ {
		o := i * 4
		img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
	}
	f, err := os.Create(filepath.Join(out, "ng_confirm_v3_fixture.png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s dumped", f.Name())
}
