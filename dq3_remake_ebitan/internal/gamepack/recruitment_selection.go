package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// RecruitmentSelection 保存 docs/188 有限 READY 的清單與取消契約。
// 空名冊View／Join與單人Leave依正常來源閉合；滿隊及非空分離另待驗。
type RecruitmentSelection struct {
	ID               string              `json:"id"`
	EntryID          string              `json:"entry_id"`
	RawWindow        RawNewGameWindow    `json:"raw_window"`
	ExtraRows        int                 `json:"extra_rows"`
	PromptTextID     string              `json:"prompt_text_id"`
	HeaderTextID     string              `json:"header_text_id"`
	RowTextID        string              `json:"row_text_id"`
	FooterTextID     string              `json:"footer_text_id"`
	AgainTextID      string              `json:"again_text_id"`
	FarewellTextID   string              `json:"farewell_text_id"`
	ContinueTextID   string              `json:"continue_text_id"`
	EmptyViewTextID  string              `json:"empty_view_text_id"`
	EmptyJoinTextID  string              `json:"empty_join_text_id"`
	EmptyLeaveTextID string              `json:"empty_leave_text_id"`
	Name             GeometryAnchor      `json:"name"`
	NameCapacity     int                 `json:"name_capacity"`
	Level            NumberField         `json:"level"`
	Class            GeometryAnchor      `json:"class"`
	Gender           GeometryAnchor      `json:"gender"`
	Cursor           GeometryAnchor      `json:"cursor"`
	HitRect          GeometryRect        `json:"hit_rect"`
	ClassOptions     []RegistrationClass `json:"class_options"`
	GenderTextIDs    []string            `json:"gender_text_ids"`
	ViewRename       *RecruitmentRename  `json:"view_rename"`
	Evidence         Evidence            `json:"evidence"`
}

// RecruitmentRename selects a reviewed finite rename primitive. The name
// widget uses the pack's existing new-game geometry and glyph definitions.
type RecruitmentRename struct {
	InputAction         string   `json:"input_action"`
	TargetScope         string   `json:"target_scope"`
	RequireNonemptyName *bool    `json:"require_nonempty_name"`
	Evidence            Evidence `json:"evidence"`
}

func (r *RecruitmentRename) UnmarshalJSON(b []byte) error {
	type plain RecruitmentRename
	return requiredHome(b, (*plain)(r))
}

func (s *RecruitmentSelection) UnmarshalJSON(b []byte) error {
	type plain RecruitmentSelection
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	for _, field := range []struct {
		key      string
		value    any
		required []string
	}{
		{"raw_window", &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}},
		{"name", &s.Name, []string{"x", "y", "step_x", "step_y"}},
		{"class", &s.Class, []string{"x", "y", "step_x", "step_y"}},
		{"gender", &s.Gender, []string{"x", "y", "step_x", "step_y"}},
		{"cursor", &s.Cursor, []string{"x", "y", "step_x", "step_y"}},
		{"level", &s.Level, []string{"x", "y", "digits"}},
		{"hit_rect", &s.HitRect, []string{"x", "y", "width", "height"}},
	} {
		if err := decodeOpeningObject(fields[field.key], field.value, field.required); err != nil {
			return err
		}
	}
	return nil
}

func (s RecruitmentSelection) TextIDs() []string {
	ids := []string{s.PromptTextID, s.HeaderTextID, s.RowTextID, s.FooterTextID, s.AgainTextID, s.FarewellTextID}
	for _, option := range s.ClassOptions {
		ids = append(ids, option.TextID)
	}
	return append(append(ids, s.GenderTextIDs...), s.ContinueTextID, s.EmptyViewTextID, s.EmptyJoinTextID, s.EmptyLeaveTextID)
}

func (p *Pack) validateRecruitmentSelection() error {
	s, entry, registration := p.Interface.RecruitmentSelection, p.Interface.RecruitmentEntry, p.Interface.Registration
	if s == nil || entry == nil || registration == nil || s.ID == "" || s.EntryID != entry.ID || s.Evidence.Level != "D3" {
		return fmt.Errorf("recruitment selection contract is missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	rename := s.ViewRename
	if rename == nil || rename.InputAction != "rename" || rename.TargetScope != "singleton_party_leader" ||
		rename.RequireNonemptyName == nil || !*rename.RequireNonemptyName || rename.Evidence.Level != "D3" {
		return fmt.Errorf("recruitment rename contract is missing or unreviewed")
	}
	if err := validateEvidence(rename.Evidence); err != nil {
		return err
	}
	w, row := s.RawWindow, s.Name
	n := registration.RosterCapacity
	height := (n + s.ExtraRows) * row.StepY
	if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 3 || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		s.ExtraRows <= 0 || row.StepY != dq3data.GlyphPx || row.StepX != dq3data.GlyphPx || s.NameCapacity <= 0 ||
		w.X*8+w.Width*8+entry.Shadow.OffsetX > 640 || w.Y+height+entry.Shadow.OffsetY > 350 {
		return fmt.Errorf("recruitment selection window is invalid")
	}
	inside := func(x, y, width, h int) bool {
		return x >= w.X*8 && y >= w.Y && width > 0 && h > 0 && x+width <= (w.X+w.Width)*8 && y+(n-1)*row.StepY+h <= w.Y+height
	}
	if !inside(row.X, row.Y, s.NameCapacity*row.StepX, dq3data.GlyphPx) || s.Level.Digits <= 0 ||
		!inside(s.Level.X, s.Level.Y, s.Level.Digits*dq3data.GlyphPx, dq3data.GlyphPx) || s.Level.Y != row.Y ||
		s.Class.Y != row.Y || s.Gender.Y != row.Y || s.Cursor.Y != row.Y ||
		s.Class.StepX != row.StepX || s.Gender.StepX != row.StepX || s.Class.StepY != row.StepY || s.Gender.StepY != row.StepY ||
		s.Cursor.StepX != 0 || s.Cursor.StepY != row.StepY || !inside(s.Cursor.X, s.Cursor.Y, dq3data.GlyphPx, dq3data.GlyphPx) ||
		s.HitRect.Y != row.Y || !inside(s.HitRect.X, s.HitRect.Y, s.HitRect.Width, s.HitRect.Height) {
		return fmt.Errorf("recruitment selection rows are invalid")
	}
	if s.RawWindow.ID == entry.Menu.RawWindow.ID || s.RawWindow.ID == registration.ClassMenu.RawWindow.ID {
		return fmt.Errorf("duplicate recruitment selection window")
	}
	for _, old := range p.Interface.NewGameGeometry.RawWindows {
		if old.ID == w.ID {
			return fmt.Errorf("duplicate recruitment selection geometry")
		}
	}
	for _, id := range s.TextIDs() {
		text, ok := p.TextDefinition(id)
		if !ok || text.Source.Kind != "legacy_record" || text.Source.Record == nil || text.Evidence.Level != "D3" || len(text.GlyphCodes) == 0 {
			return fmt.Errorf("recruitment selection text %q is unreviewed", id)
		}
		if err := validateEvidence(text.Evidence); err != nil {
			return err
		}
	}
	if row.Y != w.Y+(s.ExtraRows-1)*row.StepY || len(s.ClassOptions) == 0 || len(s.ClassOptions) > 256 {
		return fmt.Errorf("recruitment selection header or labels are invalid")
	}
	for _, frame := range []struct {
		id   string
		rows int
	}{{s.HeaderTextID, s.ExtraRows - 1}, {s.RowTextID, 1}, {s.FooterTextID, 1}} {
		codes, _ := p.TextGlyphCodes(frame.id)
		rows, width := 1, 0
		for _, code := range codes {
			if code == dq3data.TxtNL {
				if width*dq3data.GlyphPx != w.Width*8 {
					return fmt.Errorf("selection frame width differs")
				}
				rows++
				width = 0
			} else if code < dq3data.GlyphMax {
				width++
			} else {
				return fmt.Errorf("selection frame control is invalid")
			}
		}
		if rows != frame.rows || width*dq3data.GlyphPx != w.Width*8 {
			return fmt.Errorf("selection frame dimensions differ")
		}
	}
	seen := map[int]bool{}
	plainGlyphs := func(codes []uint16) bool {
		for _, code := range codes {
			if code >= dq3data.GlyphMax {
				return false
			}
		}
		return true
	}
	for _, option := range s.ClassOptions {
		codes, ok := p.TextGlyphCodes(option.TextID)
		if option.ClassRaw < 0 || option.ClassRaw > 255 || seen[option.ClassRaw] || !ok || !plainGlyphs(codes) || !inside(s.Class.X, s.Class.Y, len(codes)*dq3data.GlyphPx, dq3data.GlyphPx) {
			return fmt.Errorf("recruitment selection class mapping is invalid")
		}
		seen[option.ClassRaw] = true
	}
	for _, option := range registration.ClassOptions {
		if !seen[option.ClassRaw] {
			return fmt.Errorf("selection omits registration class")
		}
	}
	if len(s.GenderTextIDs) != 2 {
		return fmt.Errorf("recruitment selection labels are incomplete")
	}
	for _, id := range s.GenderTextIDs {
		codes, _ := p.TextGlyphCodes(id)
		if !plainGlyphs(codes) || !inside(s.Gender.X, s.Gender.Y, len(codes)*dq3data.GlyphPx, dq3data.GlyphPx) {
			return fmt.Errorf("recruitment selection gender outside row")
		}
	}
	return nil
}
