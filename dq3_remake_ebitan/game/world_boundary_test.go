package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestNormalizeOriginalSurfaceSeams(t *testing.T) {
	g := &Game{cur: &Scene{w: 244, h: 205}, layer: 0}
	for _, tc := range []struct {
		x, y         int
		wantX, wantY int
	}{
		{x: -1, y: 64, wantX: 242, wantY: 64},
		{x: 244, y: 64, wantX: 2, wantY: 64},
		{x: 210, y: -1, wantX: 210, wantY: 204},
		{x: 210, y: 205, wantX: 210, wantY: 2},
	} {
		x, y, ok := g.normalizeSurfaceTarget(tc.x, tc.y)
		if !ok || x != tc.wantX || y != tc.wantY {
			t.Errorf("surface (%d,%d) -> (%d,%d,%v), want (%d,%d,true)",
				tc.x, tc.y, x, y, ok, tc.wantX, tc.wantY)
		}
	}
}

func TestNormalizeUnderworldBoundaryFailsClosed(t *testing.T) {
	g := &Game{cur: &Scene{w: 244, h: 167}, layer: 1}
	if _, _, ok := g.normalizeSurfaceTarget(-1, 80); ok {
		t.Fatal("下層世界邊界在原版 consumer 閉合前不得外推環繞規則")
	}
}

func TestOriginalShipModeDistinguishesWaterLandAndBlockedTerrain(t *testing.T) {
	tiles := []int{0, 1, 2}
	sc := &Scene{
		w: 3, h: 1,
		tileAt: func(x, _ int) int { return tiles[x] },
		attr:   &dq3data.BlockAttr{A: []uint16{0x0000, 0x0002, 0x0007}},
	}
	g := &Game{cur: sc, over: sc, layer: 0, shipAboard: true, px: 1, py: 0}
	if !g.tryMove(0, 0) || !g.shipAboard {
		t.Fatal("attr bit1 clear 應依原版 mode1 航行，不得被舊 attr0x20 規則擋下")
	}
	if !g.tryMove(1, 0) || g.shipAboard {
		t.Fatal("attr bit1 set/bit0 clear 應允許上岸")
	}
	g.shipAboard, g.px = true, 1
	if g.tryMove(2, 0) || !g.shipAboard {
		t.Fatal("attr bit1 set/bit0 set 應阻擋船隻且保持登船")
	}
}
