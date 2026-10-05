package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func (g *Game) stepStatusMenu(in InputState, tapIdx int) {
	if g.pack == nil || g.pack.Interface.FieldStatusMenu == nil {
		return
	}
	s := g.pack.Interface.FieldStatusMenu
	n := len(s.Entries)
	if n == 0 || g.panelCursor < 0 || g.panelCursor >= n {
		return
	}
	confirm := in.Confirm
	if tapIdx >= 0 && tapIdx < n {
		g.panelCursor, confirm = tapIdx, true
	}
	switch {
	case confirm:
		// Keep the existing detailed renderer reachable. Its content and the
		// other results remain outside this first-selector parity receipt.
		if s.Entries[g.panelCursor].Role == "detail" {
			g.panel = panelStatus
		}
	case in.DirEdge == 0:
		g.panelCursor = (g.panelCursor + 1) % n
	case in.DirEdge == 1:
		g.panelCursor = (g.panelCursor + n - 1) % n
	}
}

func (g *Game) drawNativeStatusMenu() {
	if g.pack == nil || g.cur == nil || g.newGame.raster == nil || g.cmd.contract == nil {
		return
	}
	s := g.pack.Interface.FieldStatusMenu
	if s == nil {
		return
	}
	codes, ok := g.pack.TextGlyphCodes(s.TextID)
	if !ok {
		return
	}
	g.drawNativeFieldCommandMenu()
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	r.palette[s.FontIndex] = g.fieldIdleForeground()
	r.windows = map[string]gamepack.RawNewGameWindow{s.RawWindow.ID: s.RawWindow, g.cmd.contract.RawWindow.ID: g.cmd.contract.RawWindow}
	r.texts = map[string][]uint16{s.TextID: codes}
	if !r.captureBackground(g.rgba) {
		return
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: g.cmd.contract.RawWindow.ID})
	r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: s.RawWindow.ID, TextID: s.TextID})
	g.panelHits.reset()
	for i, e := range s.Entries {
		g.panelHits.add(e.X, e.Y, (s.RawWindow.X+s.RawWindow.Width)*8-e.X, dq3data.GlyphPx, i)
		if i == g.panelCursor {
			r.opaqueGlyph(g.cmd.tx, e.X, e.Y, s.CursorGlyph)
		}
	}
	drawIndexedPCX(g.rgba, r.pixels, r.palette)
}
