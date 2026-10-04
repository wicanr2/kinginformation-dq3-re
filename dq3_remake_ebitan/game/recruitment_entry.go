package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const rcGreeting = rcView + 1

func (rc *Recruit) configure(pack *gamepack.Pack, tx *dq3data.Text) error {
	if pack == nil || pack.Interface.Registration == nil || pack.Interface.RecruitmentEntry == nil || pack.Interface.RecruitmentSelection == nil || pack.Interface.OpeningPrelude == nil {
		return fmt.Errorf("recruitment entry missing")
	}
	rc.contract = pack.Interface.RecruitmentEntry
	rc.selection = pack.Interface.RecruitmentSelection
	rc.rosterCapacity = pack.Interface.Registration.RosterCapacity
	rc.texts = map[string][]uint16{}
	for _, id := range append(rc.contract.TextIDs(), rc.selection.TextIDs()...) {
		codes, ok := pack.TextGlyphCodes(id)
		if !ok {
			return fmt.Errorf("recruitment entry text missing: %q", id)
		}
		rc.texts[id] = codes
	}
	p := pack.Interface.OpeningPrelude
	frame, ok := pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return fmt.Errorf("recruitment entry frame missing")
	}
	rc.tx = tx
	rc.dialogue = Dialogue{tx: tx, layout: p.Window, prelude: p, preludeFrame: frame}
	return nil
}
func (rc *Recruit) reset() {
	rc.active = false
	rc.dialogue.open = false
	rc.dialogue.retained = nil
	rc.menuHits.reset()
	rc.listHits.reset()
}
func (rc *Recruit) startGreeting() {
	rc.dialogue.appendRetainedRecord(rc.texts[rc.contract.GreetingTextIDs[rc.greetingIndex]])
	rc.dialogue.shadow = &rc.contract.Shadow
}
func (rc *Recruit) greetingInput(in InputState) {
	rc.dialogue.Tick()
	if !rc.dialogue.open {
		rc.greetingIndex++
		if rc.greetingIndex < len(rc.contract.GreetingTextIDs) {
			rc.startGreeting()
		} else {
			rc.stage, rc.cursor = rcMenu, 0
		}
		return
	}
	if rc.dialogue.waitingForConfirm() && (in.Confirm || in.Enter || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped) {
		rc.dialogue.Advance()
	}
}
func (rc *Recruit) drawEntry(rgba []byte, white dq3data.Color) {
	if rc.contract == nil {
		return
	}
	if rc.dialogue.retained != nil {
		d := rc.dialogue
		d.open = true
		d.draw(rgba, white)
	}
	if (rc.stage != rcMenu && rc.stage != rcAgain) || rc.raster == nil || !rc.raster.captureBackground(rgba) {
		return
	}
	r := rc.raster
	if rc.stage == rcAgain {
		r.window(rc.tx, r.style.ConfirmationChoice)
		a := r.geo.StatsChoiceCursor
		r.opaqueGlyph(rc.tx, a.X, a.Y+rc.cursor*a.StepY, rc.cursorGlyph)
		rc.menuHits.reset()
		w := r.windows[r.style.ConfirmationChoice.RawWindowID]
		for i := 0; i < 2; i++ {
			rc.menuHits.add(a.X, a.Y+i*a.StepY, (w.Width-4)*8, dq3data.GlyphPx, i)
		}
		drawIndexedPCX(rgba, r.pixels, r.palette)
		return
	}
	m := rc.contract.Menu
	r.window(rc.tx, gamepack.RasterWindowRef{RawWindowID: m.RawWindow.ID, TextID: m.WindowTextID})
	r.opaqueGlyph(rc.tx, m.Cursor.X, m.Cursor.Y+rc.cursor*m.Cursor.StepY, rc.cursorGlyph)
	rc.menuHits.reset()
	for i := range rc.contract.OptionActions {
		h := m.HitRect
		rc.menuHits.add(h.X, h.Y+i*m.Cursor.StepY, h.Width, h.Height, i)
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}
func (rc *Recruit) installRaster(base *indexedNewGameRenderer, palette []dq3data.Color) error {
	if rc.contract == nil || base == nil || len(palette) < 16 {
		return fmt.Errorf("recruitment entry raster missing")
	}
	r := *base
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), palette...)
	r.backgroundPalette = append([]dq3data.Color(nil), palette...)
	f := rc.dialogue.prelude.ForegroundRGB
	r.palette[*r.style.FontIndex] = dq3data.Color{R: f[0], G: f[1], B: f[2]}
	r.windows = map[string]gamepack.RawNewGameWindow{}
	for id, w := range base.windows {
		r.windows[id] = w
	}
	r.texts = map[string][]uint16{}
	for id, c := range base.texts {
		r.texts[id] = c
	}
	m := rc.contract.Menu
	r.windows[m.RawWindow.ID] = m.RawWindow
	r.texts[m.WindowTextID] = rc.texts[m.WindowTextID]
	r.windows[rc.selection.RawWindow.ID] = rc.selection.RawWindow
	for _, id := range rc.selection.TextIDs() {
		r.texts[id] = rc.texts[id]
	}
	rc.raster = &r
	return nil
}
