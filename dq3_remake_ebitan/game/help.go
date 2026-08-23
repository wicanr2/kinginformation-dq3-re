package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// HelpOverlay 是跨版本 HELP renderer；可見內容與視窗幾何都由 game pack 提供。
type HelpOverlay struct {
	open   bool
	layout gamepack.HelpOverlay
	tx     *dq3data.Text
}

// utilityInput handles HELP and system settings without depending on a renderer.
// The caller decides which gameplay states permit these overlays.
func (g *Game) utilityInput(in InputState) bool {
	switch {
	case g.help.open:
		if in.Help || in.Cancel || in.Confirm {
			g.help.open = false
		}
		return true
	case g.settings.open:
		g.settingsInput(in)
		return true
	case in.Help:
		g.help.open = true
		return true
	case in.Settings:
		g.settings.Open()
		return true
	}
	return false
}

func (h *HelpOverlay) draw(rgba []byte, fg dq3data.Color) {
	if !h.open || h.layout.Window.Frame == nil {
		return
	}
	w := h.layout.Window
	fillPackBox(rgba, w, w.X, w.Y, w.Width, w.Height)
	for row, line := range h.layout.Lines {
		y := w.Y + w.TextInsetY + row*dq3data.GlyphPx
		for col, glyph := range line {
			drawGlyph(rgba, h.tx, w.X+w.TextInsetX+col*dq3data.GlyphPx, y, glyph, fg)
		}
	}
}
