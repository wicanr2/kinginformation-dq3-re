package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"image"
)

// docs/188 READY: redraw the ability page, deactivate its frame, then show
// the pack-owned spell union. Closing redraws the unchanged ability page.
func (r *indexedNewGameRenderer) characterSpells(tx *dq3data.Text, ids []string) {
	s := r.spells
	if s == nil || len(ids) == 0 {
		return
	}
	for _, id := range ids {
		if len(r.texts[id]) == 0 {
			return
		}
	}
	rows := (len(ids) + s.Columns - 1) / s.Columns
	w := s.RawWindow
	w.Height = (rows + s.ExtraRows) * s.Names.StepY
	r.windows[w.ID] = w
	r.frame(r.style.Ability)
	x, y, width := w.X*8, w.Y, w.Width*8
	o := r.style.ShadowOffset
	r.shadow(image.Rect(x+o.X, y+o.Y, x+o.X+width, y+o.Y+w.Height))
	r.text(tx, s.HeaderTextID, image.Pt(x, y))
	for row := 0; row < rows; row++ {
		r.text(tx, s.RowTextID, image.Pt(x, y+(row+1)*s.Names.StepY))
	}
	r.text(tx, s.FooterTextID, image.Pt(x, y+(rows+1)*s.Names.StepY))
	for i, id := range ids {
		r.text(tx, id, image.Pt(s.Names.X+(i%s.Columns)*s.Names.StepX, s.Names.Y+(i/s.Columns)*s.Names.StepY))
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: w.ID})
}
