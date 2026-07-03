package dq3data

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpLinBLS:把 DQ3LIN.BLS(46086 = 12×0xf00+6,同 DQ3MAN.BLS 格式)12 隻角色
// 各 4 方向 frame 全 dump,驗證「b2<4 NPC 的 sprite 源」假說(docs/65 U3 RE)。
// 平常 skip;DQ3_DUMP_LIN=1 + LIN_OUT=<gitignored 夾> 才跑。
func TestDumpLinBLS(t *testing.T) {
	if os.Getenv("DQ3_DUMP_LIN") == "" {
		t.Skip("設 DQ3_DUMP_LIN=1 才執行")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	out := os.Getenv("LIN_OUT")
	if out == "" {
		t.Fatal("需設 LIN_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	lin, err := os.ReadFile(filepath.Join(dir, "DQ3LIN.BLS"))
	if err != nil {
		t.Fatal(err)
	}
	pal := DecodePalette(mustRead(t, filepath.Join(dir, "DQ3.PAL")), 256)
	t.Logf("DQ3LIN.BLS %d bytes(12×0xf00+6=%d)", len(lin), 12*0xf00+6)

	// 12 隻 × 4 方向(walk frame 0);一隻拼成一橫條
	for ch := 0; ch < 12; ch++ {
		spr := LoadCharSprite(lin, ch*4)
		w, h := CharW, CharH
		img := image.NewRGBA(image.Rect(0, 0, w*4+6, h))
		for dir := 0; dir < 4; dir++ {
			fr := &spr.Frames[dir*CharWalk]
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					if !fr.Opaque[y][x] {
						continue
					}
					c := pal[fr.Px[y][x]]
					img.Set(dir*(w+2)+x, y, color.RGBA{c.R, c.G, c.B, 255})
				}
			}
		}
		fo, err := os.Create(filepath.Join(out, "lin_ch"+string(rune('0'+ch/10))+string(rune('0'+ch%10))+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(fo, img); err != nil {
			t.Fatal(err)
		}
		fo.Close()
		t.Logf("ch%02d:%d×%d ×4frame ✓", ch, w, h)
	}
}

func mustRead(t *testing.T, p string) []byte {
	d, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	return d
}
