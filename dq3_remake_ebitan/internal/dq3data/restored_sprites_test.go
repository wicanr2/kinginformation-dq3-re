package dq3data

import "testing"

// TestRestoredSprites 驗證 boss 128/129 的復原 sprite 正確解碼。
func TestRestoredSprites(t *testing.T) {
	shp := findAsset(t, "DQ3MNS.SHP")

	// 測試 id128 (歐里狄加 / Oridecon):原版 SHP 為空,應改用復原資料成功解碼
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

	// 測試 id129 (五頭龍大王 / Ryuuou):原版 SHP 為空,應改用復原資料成功解碼
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
