package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"image"
)

const (
	rcText = rcGreeting + 1 + iota
	rcAgain
	rcFinalWait
)

// 僅作文字EOF後的具名延續，不保存額外UI或資料交易。
const rcEmptyViewReturn = rcViewSpells + 1

func (rc *Recruit) startSelectionText(id string, next int) {
	codes, ok := rc.texts[id]
	if !ok {
		rc.reset()
		return
	}
	rc.dialogue.appendRetainedRecord(codes)
	rc.dialogue.shadow = &rc.contract.Shadow
	rc.stage, rc.afterText, rc.cursor = rcText, next, 0
}

func (rc *Recruit) selectionInput(in InputState) {
	confirm := in.Confirm || in.Enter
	switch rc.stage {
	case rcText:
		rc.dialogue.Tick()
		if !rc.dialogue.open {
			if rc.afterText == rcEmptyViewReturn {
				rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
				return
			}
			rc.stage = rc.afterText
			return
		}
		if rc.dialogue.waitingForConfirm() && (confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped) {
			rc.dialogue.Advance()
		}
	case rcAgain:
		if in.Tapped {
			if row := rc.menuHits.at(in.TapX, in.TapY); row >= 0 {
				rc.cursor, confirm = row, true
			}
		}
		if in.Cancel || confirm {
			if rc.cursor == 0 {
				rc.startSelectionText(rc.selection.ContinueTextID, rcMenu)
			} else {
				rc.startSelectionText(rc.selection.FarewellTextID, rcFinalWait)
			}
		} else if in.DirEdge >= 0 {
			rc.cursor ^= 1
		}
	case rcFinalWait:
		if confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped {
			rc.reset()
		}
	}
}

func (g *Game) drawRecruitSelection(rgba []byte, white dq3data.Color) {
	rc := &g.recruit
	rc.drawEntry(rgba, white)
	s, r := rc.selection, rc.raster
	n := len(g.roster)
	if s == nil || r == nil || n == 0 || n > rc.rosterCapacity || !r.captureBackground(rgba) {
		return
	}
	w := s.RawWindow
	w.Height = (n + s.ExtraRows) * s.Name.StepY
	r.windows[w.ID] = w
	x, y, width := w.X*8, w.Y, w.Width*8
	offset := r.style.ShadowOffset
	r.shadow(image.Rect(x+offset.X, y+offset.Y, x+offset.X+width, y+offset.Y+w.Height))
	r.text(rc.tx, s.HeaderTextID, image.Pt(x, y))
	for i := 0; i < n; i++ {
		r.text(rc.tx, s.RowTextID, image.Pt(x, s.Name.Y+i*s.Name.StepY))
	}
	r.text(rc.tx, s.FooterTextID, image.Pt(x, s.Name.Y+n*s.Name.StepY))
	rc.listHits.reset()
	for row, member := range g.roster {
		y := s.Name.Y + row*s.Name.StepY
		for i, code := range member.Name {
			if i >= s.NameCapacity {
				break
			}
			r.opaqueGlyph(rc.tx, s.Name.X+i*s.Name.StepX, y, code)
		}
		level := s.Level
		level.Y += row * s.Name.StepY
		r.number(rc.tx, level, member.Level())
		for _, option := range s.ClassOptions {
			if option.ClassRaw == member.Class {
				r.text(rc.tx, option.TextID, image.Pt(s.Class.X, y))
				break
			}
		}
		if member.Gender >= 0 && member.Gender < len(s.GenderTextIDs) {
			r.text(rc.tx, s.GenderTextIDs[member.Gender], image.Pt(s.Gender.X, y))
		}
		h := s.HitRect
		rc.listHits.add(h.X, h.Y+row*s.Name.StepY, h.Width, h.Height, row)
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: w.ID})
	if rc.cursor >= 0 && rc.cursor < n {
		r.opaqueGlyph(rc.tx, s.Cursor.X, s.Cursor.Y+rc.cursor*s.Cursor.StepY, rc.cursorGlyph)
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}
