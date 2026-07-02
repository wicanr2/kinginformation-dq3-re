package main

import "testing"

func TestQuantize4(t *testing.T) {
	cases := []struct {
		dx, dy, want int
		name         string
	}{
		{0, 0, -1, "死區中心"},
		{5, 5, -1, "死區內"},
		{0, 40, 0, "下"},
		{0, -40, 1, "上"},
		{-40, 0, 2, "左"},
		{40, 0, 3, "右"},
		{40, 10, 3, "右下偏右→右(主軸)"},
		{10, 40, 0, "右下偏下→下(主軸)"},
	}
	for _, c := range cases {
		if got := quantize4(c.dx, c.dy); got != c.want {
			t.Errorf("%s quantize4(%d,%d)=%d,want %d", c.name, c.dx, c.dy, got, c.want)
		}
	}
}

func TestTouchZones(t *testing.T) {
	tu := newTouchUI()
	if z := tu.zoneOf(aCX, aCY); z != zoneA {
		t.Errorf("A 鍵中心應 zoneA,得 %d", z)
	}
	if z := tu.zoneOf(bCX, bCY); z != zoneB {
		t.Errorf("B 鍵中心應 zoneB,得 %d", z)
	}
	if z := tu.zoneOf(padCX, padCY); z != zoneDpad {
		t.Errorf("十字鍵中心應 zoneDpad,得 %d", z)
	}
	if z := tu.zoneOf(ScreenW/2+10, 10); z != zoneNone { // 右上空白
		t.Errorf("右上空白應 zoneNone,得 %d", z)
	}
}
