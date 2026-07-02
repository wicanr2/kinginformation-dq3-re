package game

import "testing"

// findCtyAtLayer:下層城(cty77-93,ctyLoc[i][2]==1)只在 layer1 配對,不在 layer0 誤配。
func TestFindCtyAtLayer(t *testing.T) {
	// cty77 是下層城之一;取其座標,驗 layer1 命中、layer0 不命中。
	lx, ly, lm := ctyLoc[77][0], ctyLoc[77][1], ctyLoc[77][2]
	if lm != 1 {
		t.Fatalf("cty77 第三欄應為 1(下層),實為 %d", lm)
	}
	if got := findCtyAtLayer(lx, ly, 1); got != 77 {
		t.Errorf("下層 (%d,%d) layer1 回 %d, want 77", lx, ly, got)
	}
	if got := findCtyAtLayer(lx, ly, 0); got == 77 {
		t.Errorf("下層城 cty77 不應在 layer0(地面)被配到")
	}
	// 地面城 cty0(阿里阿罕)只在 layer0
	gx, gy := ctyLoc[0][0], ctyLoc[0][1]
	if got := findCtyAtLayer(gx, gy, 0); got != 0 {
		t.Errorf("地面 (%d,%d) layer0 回 %d, want 0", gx, gy, got)
	}
	if got := findCtyAtLayer(gx, gy, 1); got == 0 {
		t.Errorf("地面城 cty0 不應在 layer1(下層)被配到")
	}
}
