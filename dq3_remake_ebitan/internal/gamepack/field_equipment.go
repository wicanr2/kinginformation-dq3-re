package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// Reviewed finite contract: docs/188; field_equipment schema: docs/84.
type FieldEquipment struct {
	Scope         string                 `json:"scope"`
	Parts         []FieldEquipmentPart   `json:"parts"`
	RawWindow     RawNewGameWindow       `json:"raw_window"`
	PreviewWindow RawNewGameWindow       `json:"preview_window"`
	RowTextID     string                 `json:"row_text_id"`
	FooterTextID  string                 `json:"footer_text_id"`
	NoneTextID    string                 `json:"none_text_id"`
	PreviewTextID string                 `json:"preview_text_id"`
	FrameRows     int                    `json:"frame_rows"`
	RowStep       int                    `json:"row_step"`
	Name          GeometryAnchor         `json:"name"`
	Cursor        GeometryAnchor         `json:"cursor"`
	Worn          GeometryAnchor         `json:"worn"`
	CursorGlyph   int                    `json:"cursor_glyph"`
	WornGlyph     int                    `json:"worn_glyph"`
	Attack        NumberField            `json:"attack"`
	Defense       NumberField            `json:"defense"`
	Eligibility   []EquipmentEligibility `json:"eligibility"`
	Evidence      Evidence               `json:"evidence"`
}
type FieldEquipmentPart struct {
	Part         int    `json:"part"`
	HeaderTextID string `json:"header_text_id"`
}
type EquipmentEligibility struct {
	Classes        []int `json:"classes"`
	RequiredGender int   `json:"required_gender"`
}

func (s *FieldEquipmentPart) UnmarshalJSON(b []byte) error {
	type plain FieldEquipmentPart
	return requiredHome(b, (*plain)(s))
}
func (s *EquipmentEligibility) UnmarshalJSON(b []byte) error {
	type plain EquipmentEligibility
	return requiredHome(b, (*plain)(s))
}
func (s *FieldEquipment) UnmarshalJSON(b []byte) error {
	type plain FieldEquipment
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var f map[string]json.RawMessage
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	for name, d := range map[string]*RawNewGameWindow{"raw_window": &s.RawWindow, "preview_window": &s.PreviewWindow} {
		if err := decodeOpeningObject(f[name], d, []string{"id", "flags", "x", "y", "width", "height", "address"}); err != nil {
			return err
		}
	}
	for name, d := range map[string]*GeometryAnchor{"name": &s.Name, "cursor": &s.Cursor, "worn": &s.Worn} {
		if err := decodeOpeningObject(f[name], d, []string{"x", "y"}); err != nil {
			return err
		}
	}
	for name, d := range map[string]*NumberField{"attack": &s.Attack, "defense": &s.Defense} {
		if err := decodeOpeningObject(f[name], d, []string{"x", "y", "digits"}); err != nil {
			return err
		}
	}
	return nil
}
func (s *FieldEquipment) CanEquip(code, class, gender int) bool {
	if s == nil || code < 0 || code >= len(s.Eligibility) {
		return false
	}
	e := s.Eligibility[code]
	if e.RequiredGender >= 0 && gender != e.RequiredGender {
		return false
	}
	for _, c := range e.Classes {
		if c == class {
			return true
		}
	}
	return false
}
func (p *Pack) validateFieldEquipment() error {
	s := p.Interface.FieldEquipment
	if s == nil || s.Scope != "healthy_single_owner_eligible_uncursed" || s.Evidence.Level != "D3" || s.FrameRows < 2 || s.RowStep != dq3data.GlyphPx || len(s.Parts) != *p.Characters.ItemStorage.PartCount || len(s.Eligibility) != len(p.Characters.ItemStorage.Items) {
		return fmt.Errorf("field equipment missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	ids := []string{s.RowTextID, s.FooterTextID, s.NoneTextID, s.PreviewTextID}
	seen := map[int]bool{}
	for _, part := range s.Parts {
		if part.Part < 0 || part.Part >= len(s.Parts) || seen[part.Part] {
			return fmt.Errorf("equipment part permutation invalid")
		}
		seen[part.Part] = true
		ids = append(ids, part.HeaderTextID)
	}
	classes, genders := map[int]bool{}, map[int]bool{}
	for _, entry := range p.Characters.PartySprite.Entries {
		classes[entry.ClassRaw], genders[entry.GenderRaw] = true, true
	}
	for _, e := range s.Eligibility {
		if e.RequiredGender != -1 && !genders[e.RequiredGender] {
			return fmt.Errorf("equipment gender invalid")
		}
		cs := map[int]bool{}
		for _, c := range e.Classes {
			if !classes[c] || cs[c] {
				return fmt.Errorf("equipment class invalid")
			}
			cs[c] = true
		}
	}
	for _, w := range []RawNewGameWindow{s.RawWindow, s.PreviewWindow} {
		if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 7 || w.X < 0 || w.Y < 0 || w.Width < 1 || w.Width%2 != 0 || w.Height < 1 || (w.X+w.Width)*8 > 640 || w.Y+w.Height > 350 {
			return fmt.Errorf("equipment window invalid")
		}
	}
	if s.RawWindow.ID == s.PreviewWindow.ID || s.RawWindow.Flags&3 != 3 || s.PreviewWindow.Flags != 1 {
		return fmt.Errorf("equipment window activity invalid")
	}
	maxRows := p.Events.ItemActions.PersonalInventorySlots + s.FrameRows
	if s.RawWindow.Y+maxRows*s.RowStep+s.PreviewWindow.Height > 350 {
		return fmt.Errorf("equipment capacity outside canvas")
	}
	for _, a := range []GeometryAnchor{s.Name, s.Cursor, s.Worn} {
		if a.X < s.RawWindow.X*8 || a.X+dq3data.GlyphPx > (s.RawWindow.X+s.RawWindow.Width)*8 || a.Y < s.RawWindow.Y || a.Y+(p.Events.ItemActions.PersonalInventorySlots+1)*s.RowStep > s.RawWindow.Y+maxRows*s.RowStep {
			return fmt.Errorf("equipment anchor invalid")
		}
	}
	for _, g := range []int{s.CursorGlyph, s.WornGlyph} {
		if g < 0 || g >= dq3data.GlyphMax {
			return fmt.Errorf("equipment marker invalid")
		}
	}
	for _, n := range []NumberField{s.Attack, s.Defense} {
		if n.Digits < 1 || n.X < s.PreviewWindow.X*8 || n.X+n.Digits*dq3data.GlyphPx > (s.PreviewWindow.X+s.PreviewWindow.Width)*8 || n.Y < 0 || n.Y+dq3data.GlyphPx > s.PreviewWindow.Height {
			return fmt.Errorf("equipment number invalid")
		}
	}
	for _, id := range ids {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("equipment text reference invalid")
		}
		col, rows := 0, 1
		for _, c := range d.GlyphCodes {
			if c == int(dq3data.TxtNL) {
				if id != s.PreviewTextID || col != d.Layout.Columns {
					return fmt.Errorf("equipment frame newline invalid")
				}
				col = 0
				rows++
				continue
			}
			if c < 0 || c >= dq3data.GlyphMax {
				return fmt.Errorf("equipment text control invalid")
			}
			col++
		}
		if id == s.NoneTextID {
			if s.Name.X+col*dq3data.GlyphPx > (s.RawWindow.X+s.RawWindow.Width)*8 {
				return fmt.Errorf("equipment none text overflow")
			}
			continue
		}
		w := s.RawWindow
		if id == s.PreviewTextID {
			w = s.PreviewWindow
		}
		if col != d.Layout.Columns || rows != d.Layout.LinesPerPage || col*dq3data.GlyphPx != w.Width*8 || (id == s.PreviewTextID && rows*dq3data.GlyphPx != w.Height) {
			return fmt.Errorf("equipment frame shape invalid")
		}
	}
	return nil
}
