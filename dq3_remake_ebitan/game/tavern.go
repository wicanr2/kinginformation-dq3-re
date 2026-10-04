package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

// docs/188 有限 READY：姓名 → 職業 → 性別 → 能力 → 接受 → 登錄。
const (
	tavClass = iota
	tavName
	tavGender
	tavText
	tavInitialQuestion
	tavReview
	tavAccept
	tavAgain
	tavFinalWait
	tavSpells
)

type Tavern struct {
	tx                     *dq3data.Text
	active                 bool
	stage, cursor, pendCls int
	ni                     NameInput
	gs                     GenderSelect
	labels                 *gamepack.NewGameLabels
	geometry               *gamepack.NewGameGeometry
	classHits              hitList
	equipment              [4]int
	contract               *gamepack.Registration
	texts                  map[string][]uint16
	dialogue               Dialogue
	afterText              int
	candidate              *Member
	occupied               int
	items                  *dq3data.Items
	raster                 *indexedNewGameRenderer
}

func (tv *Tavern) setLabels(labels gamepack.NewGameLabels) {
	tv.labels = &labels
	tv.ni.setLabels(tv.labels)
	tv.gs.setLabels(tv.labels)
}
func (tv *Tavern) setGeometry(geometry gamepack.NewGameGeometry) { tv.geometry = &geometry }
func (tv *Tavern) configure(pack *gamepack.Pack, tx *dq3data.Text) error {
	if pack == nil || pack.Interface.Registration == nil || pack.Interface.OpeningPrelude == nil {
		return fmt.Errorf("registration contract missing")
	}
	tv.contract = pack.Interface.Registration
	tv.texts = map[string][]uint16{}
	for _, id := range tv.contract.TextIDs() {
		codes, ok := pack.TextGlyphCodes(id)
		if !ok {
			return fmt.Errorf("registration text missing: %q", id)
		}
		tv.texts[id] = codes
	}
	p := pack.Interface.OpeningPrelude
	frame, ok := pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return fmt.Errorf("registration frame missing")
	}
	tv.tx = tx
	tv.dialogue = Dialogue{tx: tx, layout: p.Window, prelude: p, preludeFrame: frame}
	return nil
}
func (tv *Tavern) reset() {
	tv.active = false
	tv.candidate = nil
	tv.dialogue.open = false
	tv.dialogue.retained = nil
	tv.classHits.reset()
	tv.ni.hits.reset()
	tv.gs.hits.reset()
}
func (tv *Tavern) open() {
	tv.reset()
	if tv.contract == nil {
		return
	}
	tv.active = true
	tv.cursor = 0
	tv.ni = NameInput{}
	tv.ni.setLabels(tv.labels)
	tv.gs = GenderSelect{}
	tv.gs.setLabels(tv.labels)
	tv.startText(tv.contract.TextRoles.Greeting, tavInitialQuestion)
}
func (tv *Tavern) startText(id string, next int) {
	codes, ok := tv.texts[id]
	if !ok {
		tv.reset()
		return
	}
	tv.dialogue.appendRetainedRecord(codes)
	tv.dialogue.shadow = &tv.contract.Shadow
	tv.stage = tavText
	tv.afterText = next
	tv.cursor = 0
}

// 選單消費 Enter/Space；文字 EOF 與後續選單不共用同一個按鍵。
func (tv *Tavern) input(in InputState, rs ...*rng.RNG) (*Member, bool) {
	if !tv.active {
		return nil, true
	}
	if tv.contract == nil && tv.stage != tavName {
		return nil, false
	}
	r := rng.New(0x1357)
	if len(rs) > 0 && rs[0] != nil {
		r = rs[0]
	}
	confirm := in.Confirm || in.Enter
	tap := -1
	if in.Tapped {
		switch tv.stage {
		case tavClass:
			tap = tv.classHits.at(in.TapX, in.TapY)
		case tavName:
			tap = tv.ni.hits.at(in.TapX, in.TapY)
		case tavGender:
			tap = tv.gs.hits.at(in.TapX, in.TapY)
		case tavInitialQuestion, tavAgain, tavAccept:
			tap = tv.classHits.at(in.TapX, in.TapY)
		}
	}
	switch tv.stage {
	case tavText:
		tv.dialogue.Tick()
		if !tv.dialogue.open {
			tv.stage = tv.afterText
			return nil, false
		}
		if tv.dialogue.waitingForConfirm() && (confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped) {
			tv.dialogue.Advance()
		}
	case tavInitialQuestion, tavAgain, tavAccept:
		if tap >= 0 {
			tv.cursor = tap
			confirm = true
		}
		if in.Cancel || confirm {
			yes := confirm && !in.Cancel && tv.cursor == 0
			switch tv.stage {
			case tavInitialQuestion:
				if !yes {
					tv.startText(tv.contract.TextRoles.InitialDecline, tavFinalWait)
				} else if tv.occupied < tv.contract.RosterCapacity {
					tv.ni.Init()
					tv.startText(tv.contract.TextRoles.NamePrompt, tavName)
				}
			case tavAgain:
				if !yes {
					tv.startText(tv.contract.TextRoles.Farewell, tavFinalWait)
				} else if tv.occupied < tv.contract.RosterCapacity {
					tv.ni.Init()
					tv.startText(tv.contract.TextRoles.NamePrompt, tavName)
				}
			case tavAccept:
				if !yes {
					tv.candidate = nil
					tv.startText(tv.contract.TextRoles.CreationCancelled, tavAgain)
				} else if tv.candidate != nil && tv.occupied < tv.contract.RosterCapacity {
					m := tv.candidate
					tv.candidate = nil
					tv.startText(tv.contract.TextRoles.Registered, tavAgain)
					return m, false
				}
			}
		} else if in.DirEdge >= 0 {
			tv.cursor ^= 1
		}
	case tavName:
		done, cancel := tv.ni.input(in, tap)
		if cancel && tv.contract != nil {
			tv.candidate = nil
			tv.startText(tv.contract.TextRoles.CreationCancelled, tavAgain)
		} else if done && len(tv.ni.nameBuf) > 0 {
			tv.stage = tavClass
			tv.cursor = 0
		}
	case tavClass:
		if tap >= 0 {
			tv.cursor = tap
			confirm = true
		}
		if in.Cancel {
			tv.startText(tv.contract.TextRoles.CreationCancelled, tavAgain)
		} else if confirm {
			if tv.cursor < 0 || tv.cursor >= len(tv.contract.ClassOptions) {
				return nil, false
			}
			tv.pendCls = tv.contract.ClassOptions[tv.cursor].ClassRaw
			tv.gs.Init()
			tv.stage = tavGender
		} else if in.DirEdge >= 0 {
			delta := 1
			if in.DirEdge == 1 || in.DirEdge == 2 {
				delta = -1
			}
			tv.cursor = (tv.cursor + delta + len(tv.contract.ClassOptions)) % len(tv.contract.ClassOptions)
		}
	case tavGender:
		if in.Cancel {
			tv.startText(tv.contract.TextRoles.CreationCancelled, tavAgain)
			break
		}
		if in.DirEdge == 2 {
			in.DirEdge = 1
		}
		if in.DirEdge == 3 {
			in.DirEdge = 0
		}
		gender, done := tv.gs.input(in, tap)
		if done {
			tv.candidate = newLevelOneMember(tv.ni.nameBuf, tv.pendCls, gender, r, tv.equipment)
			tv.stage = tavReview
			tv.cursor = 0
		}
	case tavReview:
		if confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped {
			if tv.candidate != nil && len(tv.candidate.LearnedSpells) > 0 {
				if tv.raster == nil {
					return nil, false
				}
				if _, ok := tv.raster.spells.OrderedTextIDs(tv.candidate.LearnedSpells); !ok {
					return nil, false
				}
				tv.stage = tavSpells
				return nil, false
			}
			tv.stage = tavAccept
			tv.cursor = 0
		}
	case tavSpells:
		if confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped {
			tv.stage = tavAccept
			tv.cursor = 0
		}
	case tavFinalWait:
		if confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped {
			tv.reset()
			return nil, true
		}
	}
	return nil, false
}
func (tv *Tavern) draw(rgba []byte, white dq3data.Color) {
	if !tv.active || tv.geometry == nil {
		return
	}
	// EOF 後仍保留原版文字畫布。這個副本只供重畫，不能消費輸入。
	if tv.dialogue.retained != nil {
		d := tv.dialogue
		d.open = true
		d.draw(rgba, white)
	}
	if tv.raster == nil {
		// 明確安裝 pack 幾何的 widget fixture；production 必須有索引色 renderer。
		if tv.stage == tavName {
			tv.ni.draw(rgba, tv.tx, white, white, *tv.geometry)
		}
		return
	}
	r := tv.raster
	if !r.captureBackground(rgba) {
		return
	}
	nf := NewGameFlow{ni: tv.ni, gs: tv.gs, labels: tv.labels, geometry: tv.geometry}
	switch tv.stage {
	case tavName:
		nf.stage = ngName
		r.drawContent(tv.tx, &nf)
		tv.ni.hits = nf.ni.hits
	case tavClass, tavGender:
		// 姓名輸入器返回後撤窗；底部文字由登錄 caller 保留。
		if tv.stage == tavClass {
			m := tv.contract.ClassMenu
			r.window(tv.tx, gamepack.RasterWindowRef{RawWindowID: m.RawWindow.ID, TextID: m.WindowTextID})
			r.opaqueGlyph(tv.tx, m.Cursor.X, m.Cursor.Y+tv.cursor*m.Cursor.StepY, tv.labels.ChoiceCursor[0])
			tv.classHits.reset()
			for i := range tv.contract.ClassOptions {
				h := m.HitRect
				tv.classHits.add(h.X, h.Y+i*m.Cursor.StepY, h.Width, h.Height, i)
			}
		} else {
			nf.stage = ngGender
			r.drawContent(tv.tx, &nf)
			tv.gs.hits = nf.gs.hits
		}
	case tavInitialQuestion, tavAgain:
		s := r.style
		r.window(tv.tx, s.ConfirmationChoice)
		a := tv.geometry.StatsChoiceCursor
		r.opaqueGlyph(tv.tx, a.X, a.Y+tv.cursor*a.StepY, tv.labels.ChoiceCursor[0])
		tv.choiceHits()
	case tavReview, tavAccept, tavSpells:
		if tv.candidate == nil {
			return
		}
		m := tv.candidate
		labels := *tv.labels
		for _, c := range tv.contract.ClassOptions {
			if c.ClassRaw == m.Class {
				labels.Hero = nil
				for _, v := range tv.texts[c.TextID] {
					labels.Hero = append(labels.Hero, int(v))
				}
				break
			}
		}
		nf.labels = &labels
		nf.preview = m.Stats
		nf.previewGender = m.Gender
		nf.previewLevel = m.Level()
		nf.previewHP = m.CurHP
		nf.previewMP = m.CurMP
		nf.previewDef = tv.previewDefense()
		nf.stage = ngReview
		nf.confirmCursor = tv.cursor
		if tv.stage == tavAccept {
			nf.stage = ngConfirm
			tv.choiceHits()
		}
		r.drawContent(tv.tx, &nf)
		if tv.stage == tavSpells {
			if ids, ok := r.spells.OrderedTextIDs(m.LearnedSpells); ok {
				r.characterSpells(tv.tx, ids)
			}
		}
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}
func (tv *Tavern) choiceHits() {
	tv.classHits.reset()
	s := tv.raster.style
	w := tv.raster.windows[s.ConfirmationChoice.RawWindowID]
	a := tv.geometry.StatsChoiceCursor
	for i := 0; i < 2; i++ {
		tv.classHits.add(a.X, a.Y+i*a.StepY, (w.Width-4)*8, dq3data.GlyphPx, i)
	}
}
func (tv *Tavern) previewDefense() int { return tv.candidate.Def(tv.items) }
