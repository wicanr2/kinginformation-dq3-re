package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
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

func TestPatternFillCheckerboardWritesOnlyBluePhase(t *testing.T) {
	const (
		x = 360
		y = 62
		w = 8
		h = 6
	)
	underlay := dq3data.Color{R: 3, G: 5, B: 7}
	blue := dq3data.Color{R: 0, G: 85, B: 223}
	rgba := make([]byte, ScreenW*ScreenH*4)
	for i := 0; i < len(rgba); i += 4 {
		rgba[i], rgba[i+1], rgba[i+2], rgba[i+3] = underlay.R, underlay.G, underlay.B, 255
	}

	fillPatternRect(rgba, &gamepack.PatternFill{
		ID:      "test:checker",
		Pattern: "checkerboard_1px",
		Rect:    gamepack.GeometryRect{X: x, Y: y, Width: w, Height: h},
		RGB:     [3]uint8{blue.R, blue.G, blue.B},
	})
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			o := ((y+row)*ScreenW + x + col) * 4
			want := underlay
			if (row+col)&1 == 1 {
				want = blue
			}
			got := dq3data.Color{R: rgba[o], G: rgba[o+1], B: rgba[o+2]}
			if got != want {
				t.Fatalf("pixel (%d,%d)=%+v, want %+v", col, row, got, want)
			}
		}
	}
}

func TestCheckerboardFrame2pxMatchesOriginalEdgePhases(t *testing.T) {
	const (
		x = 367
		y = 68
		w = 98
		h = 50
	)
	underlay := dq3data.Color{R: 0, G: 85, B: 223}
	border := dq3data.Color{R: 255, G: 223, B: 255}
	accent := dq3data.Color{R: 235, G: 178, B: 130}
	rgba := make([]byte, ScreenW*ScreenH*4)
	for i := 0; i < len(rgba); i += 4 {
		rgba[i], rgba[i+1], rgba[i+2], rgba[i+3] = underlay.R, underlay.G, underlay.B, 255
	}

	strokeCheckerboardFrame2px(rgba, x, y, w, h, border, accent)
	colorAt := func(px, py int) dq3data.Color {
		o := (py*ScreenW + px) * 4
		return dq3data.Color{R: rgba[o], G: rgba[o+1], B: rgba[o+2]}
	}
	// The stable DOSBox frame has two alternating rows/columns on each edge;
	// the underlying blue checker remains visible at the four outer corners.
	checks := []struct {
		px, py int
		want   dq3data.Color
	}{
		{x, y, underlay}, {x + 1, y, border}, {x + 2, y, accent},
		{x, y + 1, border}, {x + 1, y + 1, accent}, {x + 2, y + 1, border},
		{x, y + 2, accent}, {x + 1, y + 2, border},
		{x + w - 1, y, underlay}, {x + w - 2, y, border},
		{x + w - 1, y + 1, accent}, {x + w - 2, y + 1, border},
		{x, y + h - 1, border}, {x + 1, y + h - 1, accent},
		{x, y + h - 2, accent}, {x + 1, y + h - 2, border},
	}
	for _, tc := range checks {
		if got := colorAt(tc.px, tc.py); got != tc.want {
			t.Fatalf("pixel (%d,%d)=%+v, want %+v", tc.px, tc.py, got, tc.want)
		}
	}
	if got := colorAt(x+10, y+10); got != underlay {
		t.Fatalf("frame primitive overwrote interior=%+v, want underlay %+v", got, underlay)
	}
}
