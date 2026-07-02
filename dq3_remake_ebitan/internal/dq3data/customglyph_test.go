package dq3data

import "testing"

// 兩個自建字形(設/版)應解出非全空 bitmap;越界 idx 應回 ok=false。
func TestCustomGlyph(t *testing.T) {
	for idx, name := range map[int]string{CGShe: "設", CGBan: "版"} {
		g, ok := CustomGlyph(idx)
		if !ok {
			t.Fatalf("%s(idx=%d):應為合法索引", name, idx)
		}
		set := 0
		for y := 0; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if g[y][x] != 0 {
					set++
				}
			}
		}
		if set == 0 {
			t.Errorf("%s(idx=%d):bitmap 全空", name, idx)
		}
		// row14/15 恆為字身留白(對齊原版 16×16 格式,字身只用 row0..13)。
		for y := 14; y < 16; y++ {
			for x := 0; x < 16; x++ {
				if g[y][x] != 0 {
					t.Errorf("%s row%d 應留白,實 set at x=%d", name, y, x)
				}
			}
		}
	}
	if _, ok := CustomGlyph(-1); ok {
		t.Error("idx=-1 應越界")
	}
	if _, ok := CustomGlyph(2); ok {
		t.Error("idx=2 應越界(只有 2 個自建字形)")
	}
}
