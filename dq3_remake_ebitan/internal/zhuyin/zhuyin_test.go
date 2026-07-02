package zhuyin

import "testing"

func TestZhuyin(t *testing.T) {
	if len(zhKey) != nbucket || len(zhOff) != nbucket+1 {
		t.Fatalf("表大小不合:key %d off %d nbucket %d", len(zhKey), len(zhOff), nbucket)
	}
	// 掃全表:每個 key 都能查到非空候選
	nonEmpty := 0
	for i := 0; i < nbucket; i++ {
		key := zhKey[i]
		// 反解 key → sh,ji,yu,tone
		tone := key % 5
		r := key / 5
		yu := r % 14
		r /= 14
		ji := r % 4
		sh := r / 4
		if got := Lookup(sh, ji, yu, tone); len(got) > 0 {
			nonEmpty++
		}
	}
	if nonEmpty != nbucket {
		t.Fatalf("應每 key 都有候選,得 %d/%d", nonEmpty, nbucket)
	}
	// 查無此音 → nil
	if Lookup(99, 0, 0, 0) != nil {
		t.Fatal("不存在的音應回 nil")
	}
	// CellGlyph:聲母0→65、韻母(21)→86、介音(34)→99
	if CellGlyph(0) != 65 || CellGlyph(21) != 86 || CellGlyph(34) != 99 {
		t.Fatal("CellGlyph 對映錯")
	}
	// Composer:組一個音 → 有候選
	var c Composer
	c.Init()
	// 用第一個 key 反解出的注音組一次
	key := zhKey[0]
	tone := key % 5
	r := key / 5
	yu := r % 14
	ji := (r / 14) % 4
	sh := r / 14 / 4
	c.Sh, c.Ji, c.Yu = sh, ji, yu
	c.Cursor = 37 + tone - 1 // 對應聲調格(tone1→37…);tone0→41
	if tone == 0 {
		c.Cursor = 41
	}
	c.Confirm()
	if !c.Pick || len(c.Cand) == 0 {
		t.Fatalf("組字後應進候選(key%d sh%d ji%d yu%d tone%d)", key, sh, ji, yu, tone)
	}
	t.Logf("注音表 ✓:%d key 全有候選、pool %d、Composer 組字→候選 %d 字", nbucket, len(zhPool), len(c.Cand))
}
