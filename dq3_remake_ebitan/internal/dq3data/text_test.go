package dq3data

import "testing"

func TestTextAndFont(t *testing.T) {
	fon := findAsset(t, "D3TXT00.FON")
	txt := findAsset(t, "D3TXT01.TXT") // 阿里阿罕對話 bank(dlg_bank=1)
	tx := LoadText(fon, txt)
	if tx.NRecords <= 0 {
		t.Fatalf("D3TXT01 解出 %d 記錄", tx.NRecords)
	}
	t.Logf("D3TXT01:%d 記錄、FON %d 字模", tx.NRecords, len(fon)/32)

	// 掃全部記錄:每個值要嘛 <1476(字模)、要嘛 >=0xffed(控制/插值),不該落在中間的無效區。
	glyphSeen := map[int]bool{}
	nonEmpty := 0
	for rec := 0; rec < tx.NRecords; rec++ {
		codes := tx.Record(rec)
		if len(codes) > 0 {
			nonEmpty++
		}
		for _, v := range codes {
			if v >= GlyphMax && v < txtVarLo {
				t.Fatalf("rec %d 出現無效碼 0x%04x(既非字模也非控制碼)", rec, v)
			}
			if v < GlyphMax {
				glyphSeen[int(v)] = true
			}
		}
	}
	if nonEmpty == 0 {
		t.Fatal("全部記錄皆空(疑指標表解析錯)")
	}
	t.Logf("非空記錄 %d、用到 %d 個相異字模", nonEmpty, len(glyphSeen))

	// 字型:用到的字模至少要有一個是「非全空 bitmap」(真的有畫東西)。
	anyInk := false
	for idx := range glyphSeen {
		g, ok := tx.Glyph(idx)
		if !ok {
			continue
		}
		for y := 0; y < 16 && !anyInk; y++ {
			for x := 0; x < 16; x++ {
				if g[y][x] != 0 {
					anyInk = true
					break
				}
			}
		}
		if anyInk {
			break
		}
	}
	if !anyInk {
		t.Fatal("用到的字模全空白(疑 FON 解碼有誤)")
	}
	// 決定性
	a := tx.Record(1)
	b := tx.Record(1)
	if len(a) != len(b) {
		t.Fatal("Record 解碼非決定性")
	}
	t.Log("D3TXT01 記錄解碼 + FON 字模有墨 ✓")
}
