package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"strconv"
)

// FieldSaveLoad is the finite normal field save/load contract in docs/188.
// Storage remains the engine's JSON format; raw DOS save files are an oracle.
type FieldSaveLoad struct {
	LoadClockRules        []FieldLoadClockRule `json:"load_clock_rules"`
	Shadow                WindowShadow         `json:"shadow"`
	ID                    string               `json:"id"`
	SlotCount             int                  `json:"slot_count"`
	ExtraRows             int                  `json:"extra_rows"`
	PrimaryClassRaw       int                  `json:"primary_class_raw"`
	ExperienceMaxLevel    int                  `json:"experience_max_level"`
	ExperienceTextID      string               `json:"experience_text_id"`
	QuestionTextID        string               `json:"question_text_id"`
	PromptTextID          string               `json:"prompt_text_id"`
	FarewellTextID        string               `json:"farewell_text_id"`
	HeaderTextID          string               `json:"header_text_id"`
	RowTextID             string               `json:"row_text_id"`
	FooterTextID          string               `json:"footer_text_id"`
	EmptyTextID           string               `json:"empty_text_id"`
	NameControlCodes      []uint16             `json:"name_control_codes"`
	ExperienceControlCode uint16               `json:"experience_control_code"`
	RawWindow             RawNewGameWindow     `json:"raw_window"`
	Number                NumberField          `json:"number"`
	Name                  GeometryAnchor       `json:"name"`
	NameCapacity          int                  `json:"name_capacity"`
	Level                 NumberField          `json:"level"`
	Gender                GeometryAnchor       `json:"gender"`
	Cursor                GeometryAnchor       `json:"cursor"`
	HitRect               GeometryRect         `json:"hit_rect"`
	GenderTextIDs         []string             `json:"gender_text_ids"`
	Sound                 BattleSoundCue       `json:"sound"`
	Evidence              Evidence             `json:"evidence"`
}

type FieldLoadClockRule struct {
	Layer    int      `json:"layer"`
	Clock    int      `json:"clock"`
	Evidence Evidence `json:"evidence"`
}

func (r *FieldLoadClockRule) UnmarshalJSON(b []byte) error {
	type plain FieldLoadClockRule
	return requiredHome(b, (*plain)(r))
}

func (s FieldSaveLoad) LoadClock(layer int) (int, bool) {
	for _, r := range s.LoadClockRules {
		if r.Layer == layer {
			return r.Clock, true
		}
	}
	return 0, false
}

func (s *FieldSaveLoad) UnmarshalJSON(b []byte) error {
	type plain FieldSaveLoad
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	for _, v := range []struct {
		key      string
		value    any
		required []string
	}{
		{"raw_window", &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}},
		{"number", &s.Number, []string{"x", "y", "digits"}},
		{"level", &s.Level, []string{"x", "y", "digits"}},
		{"name", &s.Name, []string{"x", "y", "step_x", "step_y"}},
		{"gender", &s.Gender, []string{"x", "y", "step_x", "step_y"}},
		{"cursor", &s.Cursor, []string{"x", "y", "step_x", "step_y"}},
		{"hit_rect", &s.HitRect, []string{"x", "y", "width", "height"}},
		{"sound", &s.Sound, []string{"cue_raw", "wait_for_completion", "evidence"}},
		{"shadow", &s.Shadow, []string{"mode", "offset_x", "offset_y", "evidence"}},
	} {
		if err := decodeOpeningObject(fields[v.key], v.value, v.required); err != nil {
			return err
		}
	}
	return nil
}

func (s FieldSaveLoad) TextIDs() []string {
	return append([]string{s.ExperienceTextID, s.QuestionTextID, s.PromptTextID, s.FarewellTextID, s.HeaderTextID, s.RowTextID, s.FooterTextID, s.EmptyTextID}, s.GenderTextIDs...)
}

func (p *Pack) validateFieldSaveLoad() error {
	s := p.Interface.FieldSaveLoad
	if s == nil || s.ID == "" || s.SlotCount < 1 || s.SlotCount > 100 || s.ExtraRows < 2 || s.ExtraRows > 8 ||
		s.PrimaryClassRaw < 0 || s.PrimaryClassRaw > 255 || s.ExperienceMaxLevel < 1 || s.ExperienceMaxLevel > 255 ||
		s.NameCapacity < 1 || s.NameCapacity > 9 || len(s.GenderTextIDs) != 2 || len(s.NameControlCodes) != 2 ||
		s.NameControlCodes[0] == s.NameControlCodes[1] || s.ExperienceControlCode < dq3data.GlyphMax ||
		s.Sound.CueRaw < 0 || s.Sound.CueRaw >= 30 || !s.Sound.WaitForCompletion || s.Evidence.Level != "D3" || s.Sound.Evidence.Level != "D3" {
		return fmt.Errorf("field save/load contract is missing or unreviewed")
	}
	if s.Shadow.Mode != "vga_word_latch_and" || s.Shadow.OffsetX < 0 || s.Shadow.OffsetY < 0 || s.Shadow.Evidence.Level != "D3" {
		return fmt.Errorf("save/load shadow is invalid")
	}
	for _, e := range []Evidence{s.Evidence, s.Sound.Evidence, s.Shadow.Evidence} {
		if err := validateEvidence(e); err != nil {
			return err
		}
	}
	for _, code := range s.NameControlCodes {
		if !dq3data.IsVarInsert(code) || code == s.ExperienceControlCode {
			return fmt.Errorf("save/load interpolation contract is invalid")
		}
	}
	if !dq3data.IsVarInsert(s.ExperienceControlCode) {
		return fmt.Errorf("save/load experience control is invalid")
	}
	classKnown := false
	if p.Interface.RecruitmentSelection != nil {
		for _, option := range p.Interface.RecruitmentSelection.ClassOptions {
			if option.ClassRaw == s.PrimaryClassRaw {
				classKnown = true
			}
		}
	}
	if !classKnown {
		return fmt.Errorf("save/load primary class reference is unknown")
	}
	if len(s.LoadClockRules) == 0 {
		return fmt.Errorf("save/load clock rules missing")
	}
	seenLayers := map[int]bool{}
	for _, r := range s.LoadClockRules {
		if r.Layer < 0 || r.Layer > 255 || seenLayers[r.Layer] || r.Clock < 0 || r.Clock >= p.Events.DayNightCycle.ClockTicks || (r.Evidence.Level != "D2" && r.Evidence.Level != "D3") {
			return fmt.Errorf("save/load clock rule invalid")
		}
		if err := validateEvidence(r.Evidence); err != nil {
			return err
		}
		seenLayers[r.Layer] = true
	}
	w, row := s.RawWindow, s.Name
	if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 3 || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 ||
		row.StepX != dq3data.GlyphPx || row.StepY != dq3data.GlyphPx || w.Height != (s.SlotCount+s.ExtraRows)*row.StepY ||
		row.Y != w.Y+(s.ExtraRows-1)*row.StepY || (w.X+w.Width)*8+s.Shadow.OffsetX > 640 || w.Y+w.Height+s.Shadow.OffsetY > 350 {
		return fmt.Errorf("save/load window is invalid")
	}
	inside := func(x, y, width, height int) bool {
		return x >= w.X*8 && y == row.Y && width > 0 && height > 0 && x+width <= (w.X+w.Width)*8 && y+(s.SlotCount-1)*row.StepY+height <= w.Y+w.Height
	}
	if !inside(row.X, row.Y, s.NameCapacity*row.StepX, dq3data.GlyphPx) || s.Number.Digits <= 0 || s.Number.Digits > 9 || s.Level.Digits <= 0 || s.Level.Digits > 9 ||
		!inside(s.Number.X+(s.Number.Digits-len(strconv.Itoa(s.SlotCount)))*dq3data.GlyphPx, s.Number.Y, len(strconv.Itoa(s.SlotCount))*dq3data.GlyphPx, dq3data.GlyphPx) || !inside(s.Level.X, s.Level.Y, s.Level.Digits*dq3data.GlyphPx, dq3data.GlyphPx) ||
		s.Gender.StepY != row.StepY || s.Gender.StepX != row.StepX || s.Cursor.StepY != row.StepY || s.Cursor.StepX != 0 ||
		!inside(s.Cursor.X, s.Cursor.Y, dq3data.GlyphPx, dq3data.GlyphPx) || !inside(s.HitRect.X, s.HitRect.Y, s.HitRect.Width, s.HitRect.Height) {
		return fmt.Errorf("save/load rows are invalid")
	}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("save/load text %q is unreviewed", id)
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return err
		}
	}
	for _, f := range []struct {
		id   string
		rows int
	}{{s.HeaderTextID, s.ExtraRows - 1}, {s.RowTextID, 1}, {s.FooterTextID, 1}} {
		codes, _ := p.TextGlyphCodes(f.id)
		rows, width := 1, 0
		for _, c := range codes {
			if c == dq3data.TxtNL {
				if width*dq3data.GlyphPx != w.Width*8 {
					return fmt.Errorf("save/load frame width differs")
				}
				rows++
				width = 0
			} else if c < dq3data.GlyphMax {
				width++
			} else {
				return fmt.Errorf("save/load frame control differs")
			}
		}
		if rows != f.rows || width*dq3data.GlyphPx != w.Width*8 {
			return fmt.Errorf("save/load frame dimensions differ")
		}
	}
	for _, id := range s.GenderTextIDs {
		codes, _ := p.TextGlyphCodes(id)
		if !inside(s.Gender.X, s.Gender.Y, len(codes)*dq3data.GlyphPx, dq3data.GlyphPx) {
			return fmt.Errorf("save/load gender exceeds row")
		}
		for _, c := range codes {
			if c >= dq3data.GlyphMax {
				return fmt.Errorf("save/load gender has control")
			}
		}
	}
	for _, old := range p.Interface.NewGameGeometry.RawWindows {
		if old.ID == w.ID {
			return fmt.Errorf("duplicate save/load window")
		}
	}
	return nil
}
