package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
	"image"
	"os"
	"path/filepath"
	"strings"
)

const (
	fsText = iota
	fsQuestion
	fsSlots
	fsSoundWait
	fsFinalWait
)

type saveSlot struct {
	index         int
	name          []int
	level, gender int
}
type saveExperience struct {
	name      []int
	remaining uint32
}

type backgroundAudio interface {
	PauseBackground()
	ResumeBackground()
}

// FieldSaveLoad owns transient UI only. All slot writes go through saveTo;
// loadFrom validates a snapshot before restore clears this modal.
type FieldSaveLoad struct {
	contract                              *gamepack.FieldSaveLoad
	texts                                 map[string][]uint16
	dialogue                              Dialogue
	shadow                                gamepack.WindowShadow
	raster                                *indexedNewGameRenderer
	background                            []byte
	active, loading                       bool
	stage, afterText, cursor, soundFrames int
	experiences                           []saveExperience
	slots                                 []saveSlot
	hits                                  hitList
	cursorGlyph                           int
}

func (m *FieldSaveLoad) configure(pack *gamepack.Pack, tx *dq3data.Text) error {
	if pack == nil || pack.Interface.FieldSaveLoad == nil || pack.Interface.OpeningPrelude == nil || pack.Interface.NewGameLabels == nil {
		return fmt.Errorf("field save/load presentation missing")
	}
	m.contract = pack.Interface.FieldSaveLoad
	m.shadow = m.contract.Shadow
	m.texts = map[string][]uint16{}
	for _, id := range m.contract.TextIDs() {
		c, ok := pack.TextGlyphCodes(id)
		if !ok {
			return fmt.Errorf("save/load text missing: %q", id)
		}
		m.texts[id] = c
	}
	p := pack.Interface.OpeningPrelude
	frame, ok := pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return fmt.Errorf("save/load frame missing")
	}
	m.dialogue = Dialogue{tx: tx, layout: p.Window, prelude: p, preludeFrame: frame}
	m.cursorGlyph = pack.Interface.NewGameLabels.ChoiceCursor[0]
	return nil
}

func (m *FieldSaveLoad) reset() {
	m.active = false
	m.background = nil
	m.slots = nil
	m.experiences = nil
	m.soundFrames = 0
	m.dialogue.open = false
	m.dialogue.retained = nil
	m.dialogue.varGlyph = nil
	m.hits.reset()
}

func fieldSaveSlotPath(index int) string {
	base := savePath()
	if index == 0 {
		return base
	}
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext) + fmt.Sprintf(".slot-%02d", index+1) + ext
}

func (g *Game) openFieldSaveLoad(loading bool) error {
	m := &g.fieldSaveLoad
	if m.contract == nil || g.cur == nil || g.newGame.raster == nil {
		return fmt.Errorf("field save/load unavailable")
	}
	slots := make([]saveSlot, 0, m.contract.SlotCount)
	for i := 0; i < m.contract.SlotCount; i++ {
		entry := saveSlot{index: i}
		b, err := os.ReadFile(fieldSaveSlotPath(i))
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil {
			s, err := decodeSave(b)
			if err != nil {
				return err
			}
			entry.name = append([]int(nil), s.HeroName...)
			entry.level = stats.LevelForExp(m.contract.PrimaryClassRaw, s.HeroExp)
			entry.gender = s.HeroGender
			if len(entry.name) == 0 || entry.gender < 0 || entry.gender >= len(m.contract.GenderTextIDs) {
				return fmt.Errorf("save slot metadata invalid")
			}
		}
		if !loading || entry.level > 0 {
			slots = append(slots, entry)
		}
	}
	if loading && len(slots) == 0 {
		return nil
	}
	m.reset()
	m.slots = slots
	m.active = true
	m.loading = loading
	m.cursor = 0
	m.background = append([]byte(nil), g.rgba...)
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	f := m.dialogue.prelude.ForegroundRGB
	r.palette[*r.style.FontIndex] = dq3data.Color{R: f[0], G: f[1], B: f[2]}
	r.windows = map[string]gamepack.RawNewGameWindow{}
	for id, w := range g.newGame.raster.windows {
		r.windows[id] = w
	}
	r.windows[m.contract.RawWindow.ID] = m.contract.RawWindow
	r.texts = map[string][]uint16{}
	for id, c := range g.newGame.raster.texts {
		r.texts[id] = c
	}
	for id, c := range m.texts {
		r.texts[id] = c
	}
	m.raster = &r
	if loading {
		m.stage = fsSlots
		return nil
	}
	add := func(name []int, class int, exp uint32) {
		level := stats.LevelForExp(class, exp)
		if level <= m.contract.ExperienceMaxLevel {
			m.experiences = append(m.experiences, saveExperience{append([]int(nil), name...), stats.ExpForLevel(class, level) - exp})
		}
	}
	add(g.heroName, m.contract.PrimaryClassRaw, g.heroExp)
	for _, member := range g.companions {
		if member != nil {
			add(member.Name, member.Class, member.Exp)
		}
	}
	m.nextExperience(g.heroName)
	return nil
}

func (m *FieldSaveLoad) text(id string, next int, name []int) {
	m.dialogue.appendRetainedRecord(m.texts[id])
	m.dialogue.shadow = &m.shadow
	m.dialogue.varGlyph = map[uint16][]int{}
	for _, code := range m.contract.NameControlCodes {
		m.dialogue.varGlyph[code] = name
	}
	m.stage, m.afterText = fsText, next
}

func (m *FieldSaveLoad) nextExperience(primaryName []int) {
	if len(m.experiences) == 0 {
		m.text(m.contract.QuestionTextID, fsQuestion, primaryName)
		return
	}
	x := m.experiences[0]
	m.experiences = m.experiences[1:]
	m.text(m.contract.ExperienceTextID, fsText, x.name)
	m.dialogue.varGlyph[m.contract.ExperienceControlCode] = digitGlyphs(int(x.remaining))
}

func (g *Game) fieldSaveLoadInput(in InputState) error {
	m := &g.fieldSaveLoad
	confirm := in.Enter || in.Confirm
	anyKey := confirm || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped
	switch m.stage {
	case fsText:
		m.dialogue.Tick()
		if !m.dialogue.open {
			if m.afterText == fsText {
				m.nextExperience(g.heroName)
			} else {
				m.stage = m.afterText
			}
			return nil
		}
		if m.dialogue.waitingForConfirm() && anyKey {
			m.dialogue.Advance()
		}
	case fsQuestion:
		if in.Tapped {
			if row := m.hits.at(in.TapX, in.TapY); row >= 0 {
				m.cursor = row
				confirm = true
			}
		}
		if confirm || in.Cancel {
			if m.cursor == 0 {
				m.cursor = 0
				m.text(m.contract.PromptTextID, fsSlots, g.heroName)
			} else {
				m.text(m.contract.FarewellTextID, fsFinalWait, g.heroName)
			}
		} else if in.DirEdge >= 0 {
			m.cursor ^= 1
		}
	case fsSlots:
		if in.Tapped {
			if row := m.hits.at(in.TapX, in.TapY); row >= 0 {
				m.cursor = row
				confirm = true
			}
		}
		if in.Cancel {
			if m.loading {
				m.reset()
			} else {
				m.text(m.contract.FarewellTextID, fsFinalWait, g.heroName)
			}
			return nil
		}
		if in.DirEdge >= 0 {
			delta := 1
			if in.DirEdge == 1 || in.DirEdge == 2 {
				delta = -1
			}
			m.cursor = (m.cursor + delta + len(m.slots)) % len(m.slots)
		} else if confirm {
			if m.loading {
				return g.loadSnapshotFile(fieldSaveSlotPath(m.slots[m.cursor].index), false, m.contract)
			}
			m.soundFrames = 0
			if a, ok := g.music.(backgroundAudio); ok {
				a.PauseBackground()
			}
			if g.music != nil {
				g.music.PlaySFX(m.contract.Sound.CueRaw)
				if m.contract.Sound.WaitForCompletion {
					m.soundFrames = sfxWaitFrames(g.music.SFXDurationNanos(m.contract.Sound.CueRaw))
				}
			}
			if err := g.saveTo(fieldSaveSlotPath(m.slots[m.cursor].index)); err != nil {
				if a, ok := g.music.(backgroundAudio); ok {
					a.ResumeBackground()
				}
				m.soundFrames = 0
				return err
			}
			m.stage = fsSoundWait
		}
	case fsSoundWait:
		if m.soundFrames > 0 {
			m.soundFrames--
			return nil
		}
		if a, ok := g.music.(backgroundAudio); ok {
			a.ResumeBackground()
		}
		m.text(m.contract.FarewellTextID, fsFinalWait, g.heroName)
	case fsFinalWait:
		if anyKey {
			m.reset()
		}
	}
	return nil
}

func (g *Game) drawFieldSaveLoad() {
	m := &g.fieldSaveLoad
	if m.dialogue.retained != nil {
		d := m.dialogue
		d.open = true
		f := d.prelude.ForegroundRGB
		d.draw(g.rgba, dq3data.Color{R: f[0], G: f[1], B: f[2]})
	}
	r, s := m.raster, m.contract
	if r == nil || (m.stage != fsQuestion && m.stage != fsSlots) || !r.captureBackground(g.rgba) {
		return
	}
	m.hits.reset()
	if m.stage == fsQuestion {
		r.window(m.dialogue.tx, r.style.ConfirmationChoice)
		a := r.geo.StatsChoiceCursor
		r.opaqueGlyph(m.dialogue.tx, a.X, a.Y+m.cursor*a.StepY, m.cursorGlyph)
		w := r.windows[r.style.ConfirmationChoice.RawWindowID]
		for i := 0; i < 2; i++ {
			m.hits.add(a.X, a.Y+i*a.StepY, (w.Width-4)*8, dq3data.GlyphPx, i)
		}
	} else {
		w := s.RawWindow
		w.Height = (len(m.slots) + s.ExtraRows) * s.Name.StepY
		r.windows[w.ID] = w
		x, y, width := w.X*8, w.Y, w.Width*8
		offset := image.Pt(s.Shadow.OffsetX, s.Shadow.OffsetY)
		r.shadow(image.Rect(x+offset.X, y+offset.Y, x+offset.X+width, y+offset.Y+w.Height))
		r.text(m.dialogue.tx, s.HeaderTextID, image.Pt(x, y))
		for row, slot := range m.slots {
			y := s.Name.Y + row*s.Name.StepY
			r.text(m.dialogue.tx, s.RowTextID, image.Pt(x, y))
			num := s.Number
			num.Y = y
			r.number(m.dialogue.tx, num, slot.index+1)
			if slot.level == 0 {
				r.text(m.dialogue.tx, s.EmptyTextID, image.Pt(s.Name.X, y))
			} else {
				for i, c := range slot.name {
					if i >= s.NameCapacity {
						break
					}
					r.opaqueGlyph(m.dialogue.tx, s.Name.X+i*s.Name.StepX, y, c)
				}
				level := s.Level
				level.Y = y
				r.number(m.dialogue.tx, level, slot.level)
				r.text(m.dialogue.tx, s.GenderTextIDs[slot.gender], image.Pt(s.Gender.X, y))
			}
			h := s.HitRect
			m.hits.add(h.X, h.Y+row*s.Name.StepY, h.Width, h.Height, row)
		}
		r.text(m.dialogue.tx, s.FooterTextID, image.Pt(x, s.Name.Y+len(m.slots)*s.Name.StepY))
		r.frame(gamepack.RasterWindowRef{RawWindowID: w.ID})
		r.opaqueGlyph(m.dialogue.tx, s.Cursor.X, s.Cursor.Y+m.cursor*s.Cursor.StepY, m.cursorGlyph)
	}
	drawIndexedPCX(g.rgba, r.pixels, r.palette)
}
