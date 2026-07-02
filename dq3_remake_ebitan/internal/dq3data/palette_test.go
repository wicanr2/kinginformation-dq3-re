package dq3data

import (
	"os"
	"path/filepath"
	"testing"
)

// findPAL locates a real DQ3.PAL: env DQ3_ASSETS, else the repo's assets_raw.
func findPAL(t *testing.T) []byte {
	cands := []string{}
	if a := os.Getenv("DQ3_ASSETS"); a != "" {
		cands = append(cands, filepath.Join(a, "DQ3.PAL"))
	}
	cands = append(cands, filepath.Join("..", "..", "..", "assets_raw", "DQ3.PAL"))
	for _, p := range cands {
		if d, err := os.ReadFile(p); err == nil {
			return d
		}
	}
	t.Skip("DQ3.PAL 不在(設 DQ3_ASSETS 或放 assets_raw)— 跳過")
	return nil
}

func TestDecodePalette(t *testing.T) {
	raw := findPAL(t)
	pal := DecodePalette(raw, 256)
	if len(pal) == 0 {
		t.Fatal("解出 0 色")
	}
	t.Logf("DQ3.PAL %d bytes → %d 色", len(raw), len(pal))

	// 對拍 C 版公式:每分量 6-bit → 8-bit = (v<<2)|(v>>4)。逐色驗證與原始 byte 一致。
	for i, c := range pal {
		r, g, b := raw[i*3], raw[i*3+1], raw[i*3+2]
		want := Color{R: (r << 2) | (r >> 4), G: (g << 2) | (g >> 4), B: (b << 2) | (b >> 4)}
		if c != want {
			t.Fatalf("色 %d 不符:得 %+v 期 %+v", i, c, want)
		}
		// 6-bit 值 → 8-bit 應落在合理範圍(0 或 top-bits 複製)
		if r < 64 && c.R != want.R {
			t.Fatalf("色 %d R 展開錯", i)
		}
	}
	// 第一色印出來看(通常 index0=黑或背景)
	t.Logf("色0 = %+v", pal[0])
}
