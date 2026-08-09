package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestFramePatternCheckerboardLeavesUnderlyingBorder(t *testing.T) {
	const (
		x = 10
		y = 20
		w = 6
		h = 6
	)
	underlay := dq3data.Color{R: 7, G: 11, B: 13}
	interior := dq3data.Color{R: 0, G: 0, B: 0}
	border := dq3data.Color{R: 255, G: 223, B: 255}
	rgba := make([]byte, ScreenW*ScreenH*4)
	for i := 0; i < len(rgba); i += 4 {
		rgba[i], rgba[i+1], rgba[i+2], rgba[i+3] = underlay.R, underlay.G, underlay.B, 255
	}

	fillBoxStyle(rgba, x, y, w, h, interior, border, "checkerboard_1px")
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			o := ((y+row)*ScreenW + x + col) * 4
			want := interior
			if row == 0 || row == h-1 || col == 0 || col == w-1 {
				want = underlay
				if (row+col)&1 == 1 {
					want = border
				}
			}
			got := dq3data.Color{R: rgba[o], G: rgba[o+1], B: rgba[o+2]}
			if got != want {
				t.Fatalf("pixel (%d,%d)=%+v, want %+v", col, row, got, want)
			}
		}
	}
}
