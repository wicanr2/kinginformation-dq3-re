package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"io/fs"
)

func validateFieldItemSources(assets fs.FS, p *gamepack.Pack, tx *dq3data.Text) error {
	s := p.Interface.FieldItems
	if s == nil || tx == nil {
		return fmt.Errorf("field item sources missing")
	}
	for _, id := range []string{s.TextIDs.Header, s.TextIDs.Row, s.TextIDs.Footer, s.TextIDs.Actions, s.TextIDs.GivePrompt} {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Record == nil {
			return fmt.Errorf("item text reference missing")
		}
		raw, err := fs.ReadFile(assets, d.Source.File)
		if err != nil {
			return err
		}
		codes := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if len(codes) != len(d.GlyphCodes) {
			return fmt.Errorf("item text shape differs")
		}
		for i, c := range codes {
			if int(c) != d.GlyphCodes[i] {
				return fmt.Errorf("item text source differs")
			}
			if c < dq3data.GlyphMax {
				if _, ok := tx.Glyph(int(c)); !ok {
					return fmt.Errorf("item glyph missing")
				}
			}
		}
	}
	for _, glyph := range []int{s.CursorGlyph, s.WornGlyph} {
		if _, ok := tx.Glyph(glyph); !ok {
			return fmt.Errorf("item marker glyph missing")
		}
	}
	return nil
}

// drawNativeItems consumes physical entries and the pack-owned native windows.
func (g *Game) drawNativeItems(rgba []byte) {
	if g.pack == nil || g.cur == nil || g.newGame.raster == nil || g.cmd.contract == nil {
		return
	}
	s := g.pack.Interface.FieldItems
	if s == nil {
		return
	}
	entries := g.actorItemEntries(g.panelActor)
	if len(entries) == 0 {
		return
	}
	g.drawNativeFieldCommandMenu()
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	r.palette[g.cmd.contract.FontIndex] = g.fieldIdleForeground()
	w := s.RawWindow
	w.Height = (len(entries) + s.FrameRows) * s.RowStep
	r.windows = map[string]gamepack.RawNewGameWindow{w.ID: w, s.ActionWindow.ID: s.ActionWindow, g.cmd.contract.RawWindow.ID: g.cmd.contract.RawWindow}
	header, ok := g.pack.TextGlyphCodes(s.TextIDs.Header)
	if !ok {
		return
	}
	row, ok := g.pack.TextGlyphCodes(s.TextIDs.Row)
	if !ok {
		return
	}
	footer, ok := g.pack.TextGlyphCodes(s.TextIDs.Footer)
	if !ok {
		return
	}
	actions, ok := g.pack.TextGlyphCodes(s.TextIDs.Actions)
	if !ok {
		return
	}
	frame := append([]uint16(nil), header...)
	for range entries {
		frame = append(frame, dq3data.TxtNL)
		frame = append(frame, row...)
	}
	frame = append(frame, dq3data.TxtNL)
	frame = append(frame, footer...)
	r.texts = map[string][]uint16{s.TextIDs.Header: frame, s.TextIDs.Actions: actions}
	if !r.captureBackground(rgba) {
		return
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: g.cmd.contract.RawWindow.ID})
	r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: w.ID, TextID: s.TextIDs.Header})
	g.panelHits.reset()
	for i, entry := range entries {
		if entry.Word&uint16(*s.MarkerMask) != 0 {
			r.opaqueGlyph(g.cmd.tx, s.Worn.X, s.Worn.Y+i*s.RowStep, s.WornGlyph)
		}
		for j, glyph := range itemNameGlyphs(g.shop.nameText, entry.Code) {
			r.opaqueGlyph(g.cmd.tx, s.Name.X+j*dq3data.GlyphPx, s.Name.Y+i*s.RowStep, glyph)
		}
		g.panelHits.add(s.Cursor.X, s.Cursor.Y+i*s.RowStep, (w.X+w.Width)*8-s.Cursor.X, s.RowStep, i)
	}
	r.opaqueGlyph(g.cmd.tx, s.Cursor.X, s.Cursor.Y+g.panelCursor*s.RowStep, s.CursorGlyph)
	switch g.itemActionStage {
	case itemActionMenu, itemActionTarget, itemActionUseTarget:
		r.frame(gamepack.RasterWindowRef{RawWindowID: w.ID})
		r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: s.ActionWindow.ID, TextID: s.TextIDs.Actions})
		r.opaqueGlyph(g.cmd.tx, s.ActionCursor.X, s.ActionCursor.Y+g.itemActionCursor*s.RowStep, s.CursorGlyph)
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
	if g.itemActionStage == itemActionTarget {
		g.drawItemGiveTargets(rgba, g.fieldIdleForeground())
	} else if g.itemActionStage == itemActionUseTarget {
		g.drawItemUseTargets(rgba, g.fieldIdleForeground())
	}
}

func (g *Game) beginSingleOwnerGift() bool {
	if g.pack == nil || len(g.companions) != 0 || g.panelActor != 0 {
		return false
	}
	s := g.pack.Interface.FieldItems
	if s == nil {
		return false
	}
	if _, ok := g.pack.TextGlyphCodes(s.TextIDs.GivePrompt); !ok {
		return false
	}
	if !g.giveSelectedItem(0) {
		return false
	}
	g.panel = panelNone
	g.itemActionStage = itemActionGiveWait
	g.cmd.open = false
	return g.openPackText(s.TextIDs.GivePrompt)
}
