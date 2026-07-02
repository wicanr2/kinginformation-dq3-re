package dq3data

import "testing"

func TestDecodePackBG(t *testing.T) {
	scr := findAsset(t, "PACKBG.SCR")
	bg, ok := DecodePackBG(scr, 22) // 草原(對 game3.png)
	if !ok {
		t.Fatal("解 page22 失敗")
	}
	// 天空帶(上半)應有非 0 索引(藍天白雲),不是全空
	nonZero := 0
	for r := 0; r < PackBGH; r++ {
		for x := 0; x < PackBGW; x++ {
			if bg[r][x] != 0 {
				nonZero++
			}
		}
	}
	if nonZero == 0 {
		t.Fatal("page22 全 0(疑解碼有誤)")
	}
	// 越界頁 → false
	if _, ok := DecodePackBG(scr, 1<<20); ok {
		t.Fatal("越界頁應回 false")
	}
	t.Logf("packbg page22 解碼 ✓,88×640,非零像素 %d", nonZero)
}
