package game

import (
	"os"
	"testing"
)

func TestShipDirectTownEntryParksAndExitRestoresBothLayers(t *testing.T) {
	for _, tc := range []struct {
		name  string
		layer int
		cty   int
	}{
		{"surface", 0, 38},
		{"underworld", 1, 85},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
			if err != nil {
				t.Fatalf("NewGame: %v", err)
			}
			g.inTown, g.layer, g.px, g.py = false, tc.layer, 77, 66
			g.cur, g.shipOwned, g.shipAboard = g.overworldScene(), true, true
			g.shipX, g.shipY = 1, 2
			g.enterTownCty(tc.cty)
			if !g.inTown || g.curCty != tc.cty || g.shipAboard || g.shipX != 77 || g.shipY != 66 {
				t.Fatalf("進城未停泊：town=%v cty=%d aboard=%v ship=(%d,%d)",
					g.inTown, g.curCty, g.shipAboard, g.shipX, g.shipY)
			}
			g.exitTown()
			if g.inTown || !g.shipAboard || g.px != 77 || g.py != 66 {
				t.Fatalf("出城未恢復同層乘船：town=%v aboard=%v pos=(%d,%d) layer=%d",
					g.inTown, g.shipAboard, g.px, g.py, g.layer)
			}
		})
	}
}
