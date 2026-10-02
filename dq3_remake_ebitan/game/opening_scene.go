package game

import (
	"fmt"
	"io/fs"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// validateSceneCameraSources validates declared camera and layer references against the
// native town loader before any game state is created.
func validateSceneCameraSources(assets fs.FS, pack *gamepack.Pack) error {
	if (len(pack.Interface.SceneCameras) > 0 || len(pack.Interface.SceneTileLayers) > 0) && assets == nil {
		return fmt.Errorf("scene camera assets are nil")
	}
	for _, b := range pack.Interface.SceneCameras {
		raw, err := fs.ReadFile(assets, ctyFile(b.CTY))
		if err != nil {
			return fmt.Errorf("scene camera source: %w", err)
		}
		if b.Section >= len(raw)/2 {
			return fmt.Errorf("scene camera section is outside source")
		}
		town, err := dq3data.OpenTown(raw, b.Section, false)
		if err != nil {
			return fmt.Errorf("scene camera section: %w", err)
		}
		if b.Camera == nil || b.Camera.ExteriorTile != int(town.ExteriorTile) {
			return fmt.Errorf("scene camera exterior tile differs from source")
		}
	}
	for _, s := range pack.Interface.SceneTileLayers {
		raw, err := fs.ReadFile(assets, ctyFile(s.CTY))
		if err != nil {
			return fmt.Errorf("scene tile layer source: %w", err)
		}
		if s.Section >= len(raw)/2 {
			return fmt.Errorf("scene tile layer section is outside source")
		}
		town, err := dq3data.OpenTown(raw, s.Section, false)
		if err != nil {
			return fmt.Errorf("scene tile layer section: %w", err)
		}
		if s.BaseTile != int(town.BaseLayerTile) || s.OtherTile != int(town.OtherLayerTile) {
			return fmt.Errorf("scene tile layers differ from source")
		}
		if s.CTY >= len(mapBlkNum) {
			return fmt.Errorf("scene tile layer block source is unknown")
		}
		block, err := fs.ReadFile(assets, blkFile(mapBlkNum[s.CTY]))
		if err != nil {
			return fmt.Errorf("scene tile layer block source: %w", err)
		}
		blk, err := dq3data.OpenBLK(block)
		if err != nil || s.BaseTile >= blk.Count || s.OtherTile >= blk.Count {
			return fmt.Errorf("scene tile layers exceed block source")
		}
	}
	return nil
}

func (sc *Scene) tileLayer(x, y int) int {
	return dq3data.TownTileLayer(sc.hiMap[y*sc.w+x])
}

func sceneLayerTile(s *gamepack.SceneTileLayers, playerLayer, tileLayer int, outside bool, tile int) int {
	if playerLayer == s.BaseLayer {
		if !outside && tileLayer != s.BaseLayer {
			return s.BaseTile
		}
	} else if outside || tileLayer != playerLayer {
		return s.OtherTile
	}
	return tile
}

func (g *Game) activeOpeningScenePresentation() *gamepack.OpeningScenePresentation {
	e, ok := g.pack.OpeningScenePresentation()
	if !ok || (!g.homeSelection.active && !g.homeAwait && (g.openingIdx <= 0 || g.openingIdx > len(e.TextIDs))) ||
		!g.inTown || g.cur == nil || g.curCty != e.CTY || g.cur.sec != e.Section {
		return nil
	}
	return e
}

func (g *Game) activeSceneCamera() *gamepack.SceneCamera {
	if e := g.activeOpeningScenePresentation(); e != nil {
		return &e.Camera
	}
	if e, ok := g.pack.OpeningEscort(); ok && g.inTown && g.cur != nil &&
		g.curCty == e.Destination.CTY && g.cur.sec == e.Destination.Section {
		return e.ArrivalCamera
	}
	if g.inTown && g.cur != nil {
		return g.pack.SceneCamera(g.curCty, g.cur.sec)
	}
	return nil
}

func (g *Game) openOpeningSceneText(id string) bool {
	e := g.activeOpeningScenePresentation()
	p, ok := g.pack.OpeningPrelude()
	if e == nil || !ok || e.PresentationID != p.ID {
		return false
	}
	codes, ok := g.pack.TextGlyphCodes(id)
	if !ok || !g.dlg.openRecord(codes) {
		return false
	}
	frame, ok := g.pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return false
	}
	g.dlg.prelude, g.dlg.preludeFrame, g.dlg.layout = p, frame, p.Window
	g.dlg.shadow = &e.Shadow
	return true
}

func (g *Game) clearOpeningPresentation() {
	g.dlg.prelude, g.dlg.preludeFrame, g.dlg.shadow = nil, nil, nil
	g.dlg.layout = g.pack.DialogueWindowLayout()
}

// drawWindowWordLatchShadow models the reviewed executor's byte-read latches:
// both writes of a 16-bit AND use the last (second-byte) VGA latch. See docs/188.
// Values and geometry come from the pack; this is not a DOS wall-clock model.
func drawWindowWordLatchShadow(rgba []byte, w gamepack.WindowLayout, s gamepack.WindowShadow, dark dq3data.Color) {
	x, y := w.X+s.OffsetX, w.Y+s.OffsetY
	for row := 0; row < w.Height; row++ {
		for col := 0; col < w.Width; col++ {
			color := dark
			if (col+row)&1 != 0 {
				sourceX := x + (col/16)*16 + 8 + (col % 8)
				o := ((y+row)*ScreenW + sourceX) * 4
				color = dq3data.Color{R: rgba[o], G: rgba[o+1], B: rgba[o+2]}
			}
			putPx(rgba, x+col, y+row, color)
		}
	}
}
