package dq3data

import "testing"

func TestDecodePCX(t *testing.T) {
	pix, pal, w, h, err := DecodePCX(findAsset(t, "TITG.P"))
	if err != nil {
		t.Fatalf("解 TITG.P: %v", err)
	}
	if w != 640 || h != 350 {
		t.Fatalf("標題應 640×350,得 %d×%d", w, h)
	}
	if len(pal) != 16 || len(pix) != w*h {
		t.Fatalf("pal %d / pix %d 不合", len(pal), len(pix))
	}
	// 非全空(有畫面)+ 用到多種索引
	seen := map[uint8]bool{}
	for _, p := range pix {
		seen[p] = true
	}
	if len(seen) < 2 {
		t.Fatal("標題全同一色(疑解碼有誤)")
	}
	t.Logf("TITG.P 640×350、16 色、用到 %d 種索引 ✓", len(seen))
}
