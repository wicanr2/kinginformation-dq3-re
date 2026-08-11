package dq3data

import (
	"os"
	"testing"
)

// TestRestoredSprites 驗證 boss 128/129 的 sprite 正確解碼；原始唯讀基線走 fallback，
// 產生的執行版 SHP 則應可走同一個 direct loader。
func TestRestoredSprites(t *testing.T) {
	shp := findAsset(t, "DQ3MNS.SHP")

	// 測試 id128 (歐里狄加 / Oridecon)：raw 基線可能為空，patched SHP 會直接提供資料；
	// 兩種輸入都必須成功解碼。
	spr128, err := DecodeMonsterSprite(shp, 128)
	if err != nil {
		t.Fatalf("id128 復原資料解碼失敗: %v", err)
	}
	if spr128 == nil {
		t.Fatal("id128 sprite 為 nil")
	}
	// 復原資料前 4 bytes: w_bytes=12 (0x0c), h=88 (0x58) → w=12*8=96, h=88
	if spr128.W != 96 || spr128.H != 88 {
		t.Fatalf("id128 尺寸應 96×88,得 %d×%d", spr128.W, spr128.H)
	}

	// 驗證至少有不透明像素(非全透明)
	any128 := false
	for r := 0; r < spr128.H && !any128; r++ {
		for b := 0; b < spr128.W; b++ {
			if spr128.Opaque[r][b] {
				any128 = true
				break
			}
		}
	}
	if !any128 {
		t.Fatal("id128 sprite 全透明(解碼疑有誤)")
	}
	t.Logf("id128 (歐里狄加) sprite %d×%d ✓,有墨", spr128.W, spr128.H)

	// 測試 id129 (五頭龍大王 / Ryuuou)：raw 基線可能為空，patched SHP 會直接提供資料；
	// 兩種輸入都必須成功解碼。
	spr129, err := DecodeMonsterSprite(shp, 129)
	if err != nil {
		t.Fatalf("id129 復原資料解碼失敗: %v", err)
	}
	if spr129 == nil {
		t.Fatal("id129 sprite 為 nil")
	}
	// 復原資料前 4 bytes: w_bytes=16 (0x10), h=96 (0x60) → w=16*8=128, h=96
	if spr129.W != 128 || spr129.H != 96 {
		t.Fatalf("id129 尺寸應 128×96,得 %d×%d", spr129.W, spr129.H)
	}

	// 驗證至少有不透明像素
	any129 := false
	for r := 0; r < spr129.H && !any129; r++ {
		for b := 0; b < spr129.W; b++ {
			if spr129.Opaque[r][b] {
				any129 = true
				break
			}
		}
	}
	if !any129 {
		t.Fatal("id129 sprite 全透明(解碼疑有誤)")
	}
	t.Logf("id129 (五頭龍大王) sprite %d×%d ✓,有墨", spr129.W, spr129.H)
}

// TestPatchedSHPDirectLoad 只在驗證衍生執行版時啟用；它禁止測試悄悄回退到
// restoredSprite，確保 128／129 的 patched DQ3MNS.SHP 確實含可解的 RLE mask。
func TestPatchedSHPDirectLoad(t *testing.T) {
	if os.Getenv("DQ3_EXPECT_DIRECT_SHP") != "1" {
		t.Skip("設 DQ3_EXPECT_DIRECT_SHP=1 才驗證 patched SHP direct load")
	}
	shp := findAsset(t, "DQ3MNS.SHP")
	for _, tc := range []struct {
		id int
		w  int
		h  int
	}{
		{id: 128, w: 96, h: 88},
		{id: 129, w: 128, h: 96},
	} {
		if (tc.id+1)*4+4 > len(shp) {
			t.Fatalf("id%d offset 表越界", tc.id)
		}
		off, end := int(le32(shp, tc.id*4)), int(le32(shp, (tc.id+1)*4))
		spr, err := decodeSpriteBytesInternal(shp, off, end, tc.id, true)
		if err != nil {
			t.Fatalf("id%d 無法 direct decode（含 mask）: %v", tc.id, err)
		}
		if spr.W != tc.w || spr.H != tc.h {
			t.Fatalf("id%d direct 尺寸=%d×%d, want %d×%d", tc.id, spr.W, spr.H, tc.w, tc.h)
		}
		t.Logf("id%d direct SHP + RLE AND-mask ✓", tc.id)
	}
}
