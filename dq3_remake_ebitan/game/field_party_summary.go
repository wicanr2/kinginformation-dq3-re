package game

import (
	"image"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func (g *Game) drawNativePartySummary() {
	if g.pack == nil || g.cur == nil || g.newGame.raster == nil || g.cmd.contract == nil || g.pack.Interface.FieldStatusMenu == nil {
		return
	}
	menu := g.pack.Interface.FieldStatusMenu
	s := menu.Summary
	if s == nil || len(g.companions)+1 > s.MaxColumns {
		return
	}
	actors := g.partyHUDActors()
	if len(actors) != len(g.companions)+1 {
		return
	}
	g.drawNativeStatusMenu()
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	r.palette[menu.FontIndex] = g.fieldIdleForeground()
	body := s.RawWindow
	body.Width = s.WidthBase + len(actors)*s.ColumnStep/8
	r.windows = map[string]gamepack.RawNewGameWindow{menu.RawWindow.ID: menu.RawWindow, s.MoneyWindow.ID: s.MoneyWindow, body.ID: body}
	r.texts = map[string][]uint16{}
	for _, id := range []string{s.MoneyTextID, s.LeftTextID, s.ColumnTextID, s.RightTextID} {
		codes, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		r.texts[id] = codes
	}
	if !r.captureBackground(g.rgba) {
		return
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: menu.RawWindow.ID})
	r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: s.MoneyWindow.ID, TextID: s.MoneyTextID})
	r.number(g.cmd.tx, s.MoneyValue, g.heroGold)
	ref := gamepack.RasterWindowRef{RawWindowID: body.ID}
	r.windowShadow(ref)
	r.text(g.cmd.tx, s.LeftTextID, image.Pt(body.X*8, body.Y))
	for i := range actors {
		r.text(g.cmd.tx, s.ColumnTextID, image.Pt(s.ColumnOrigin.X+i*s.ColumnStep, s.ColumnOrigin.Y))
	}
	r.text(g.cmd.tx, s.RightTextID, image.Pt(s.ColumnOrigin.X+len(actors)*s.ColumnStep, s.ColumnOrigin.Y))
	r.frame(ref)
	for i, a := range actors {
		for n, code := range a.name {
			if n >= s.NameLimit {
				break
			}
			r.opaqueGlyph(g.cmd.tx, s.Name.X+i*s.ColumnStep+n*s.Name.StepX, s.Name.Y, code)
		}
		maximumHP, maximumMP := int(g.heroStat[stats.HP]), int(g.heroStat[stats.MP])
		if i > 0 {
			maximumHP, maximumMP = g.companions[i-1].MaxHP(), g.companions[i-1].MaxMP()
		}
		for _, v := range []struct {
			field gamepack.NumberField
			value int
		}{{s.HP, a.hp}, {s.MaxHP, maximumHP}, {s.MP, a.mp}, {s.MaxMP, maximumMP}} {
			f := v.field
			f.X += i * s.ColumnStep
			r.number(g.cmd.tx, f, v.value)
		}
	}
	g.panelHits.reset()
	drawIndexedPCX(g.rgba, r.pixels, r.palette)
}
