package dq3data

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDumpTitP:把 TIT*.P(PCX)+ FIRST.SCR(raw 640×350 4-plane)全解成 PNG,
// 供 attract/開場素材比對(docs/65 戰役 new-RE:立繪素材定位,rulebook 64 截圖 oracle 法)。
// 平常 skip;設 DQ3_DUMP_TITP=1 + TITP_OUT=<輸出夾> 才跑。輸出含版權圖 → 只准放 gitignored 目錄。
func TestDumpTitP(t *testing.T) {
	if os.Getenv("DQ3_DUMP_TITP") == "" {
		t.Skip("設 DQ3_DUMP_TITP=1 才執行(dump 工具,非對拍測試)")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	out := os.Getenv("TITP_OUT")
	if out == "" {
		t.Fatal("需設 TITP_OUT(gitignored 輸出夾)")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}

	writePNG := func(name string, pix []uint8, pal []Color, w, h int) {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for i, p := range pix {
			c := pal[p&0x0f]
			img.Set(i%w, i/w, color.RGBA{c.R, c.G, c.B, 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	}

	// 1) 全部 TIT*.P(PCX,自帶 16 色 palette)
	matches, _ := filepath.Glob(filepath.Join(dir, "TIT*.P"))
	okCnt := 0
	for _, f := range matches {
		d, err := os.ReadFile(f)
		if err != nil {
			t.Logf("%s:讀失敗 %v", f, err)
			continue
		}
		pix, pal, w, h, err := DecodePCX(d)
		base := strings.TrimSuffix(filepath.Base(f), ".P")
		if err != nil {
			t.Logf("%s:解碼失敗 %v", base, err)
			continue
		}
		writePNG(base, pix, pal, w, h)
		okCnt++
		t.Logf("%s:%d×%d ✓", base, w, h)
	}

	// 2) FIRST.SCR:112000 bytes = 640×350 4-plane raw(plane-major,80B/列/plane 假設先試
	//    「列內 plane 交錯」與「整面 plane 分離」兩種排列,都 dump 出來肉眼挑對的)
	if d, err := os.ReadFile(filepath.Join(dir, "FIRST.SCR")); err == nil && len(d) == 112000 {
		const w, h, bpl = 640, 350, 80
		gray := make([]Color, 16)
		for i := range gray {
			v := uint8(i * 17)
			gray[i] = Color{R: v, G: v, B: v}
		}
		// 排列 A:每列 4 plane 交錯(row-major,plane-in-row)
		pixA := make([]uint8, w*h)
		for row := 0; row < h; row++ {
			for plane := 0; plane < 4; plane++ {
				pb := d[row*4*bpl+plane*bpl:]
				for x := 0; x < w; x++ {
					bit := (pb[x>>3] >> (7 - (x & 7))) & 1
					pixA[row*w+x] |= bit << uint(plane)
				}
			}
		}
		writePNG("FIRST_scr_rowinterleave", pixA, gray, w, h)
		// 排列 B:整面 plane 分離(plane-major)
		pixB := make([]uint8, w*h)
		planeSz := bpl * h
		for plane := 0; plane < 4; plane++ {
			pb := d[plane*planeSz:]
			for row := 0; row < h; row++ {
				for x := 0; x < w; x++ {
					bit := (pb[row*bpl+x>>3] >> (7 - (x & 7))) & 1
					pixB[row*w+x] |= bit << uint(plane)
				}
			}
		}
		writePNG("FIRST_scr_planemajor", pixB, gray, w, h)
		t.Log("FIRST.SCR:兩種 plane 排列都已 dump(灰階 palette,先看形狀)")
	}

	if okCnt == 0 {
		t.Fatal("沒有任何 TIT*.P 解碼成功")
	}
	t.Logf("共 %d 張 TIT*.P 解出 → %s", okCnt, out)
}
