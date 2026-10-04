package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// CharacterSpells is a read-only character-page contract, reviewed in docs/188.
// Catalog order supplies the display order; records and geometry belong to the pack.
type CharacterSpells struct {
	RawWindow    RawNewGameWindow     `json:"raw_window"`
	Columns      int                  `json:"columns"`
	ExtraRows    int                  `json:"extra_rows"`
	Names        GeometryAnchor       `json:"names"`
	HeaderTextID string               `json:"header_text_id"`
	RowTextID    string               `json:"row_text_id"`
	FooterTextID string               `json:"footer_text_id"`
	Catalog      []CharacterSpellName `json:"catalog"`
	Evidence     Evidence             `json:"evidence"`
}

type CharacterSpellName struct {
	RecordRaw int      `json:"record_raw"`
	TextID    string   `json:"text_id"`
	Evidence  Evidence `json:"evidence"`
}

func (s *CharacterSpells) UnmarshalJSON(b []byte) error {
	type plain CharacterSpells
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	if err := decodeOpeningObject(fields["raw_window"], &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}); err != nil {
		return err
	}
	if err := decodeOpeningObject(fields["names"], &s.Names, []string{"x", "y", "step_x", "step_y"}); err != nil {
		return err
	}
	return nil
}
func (s *CharacterSpellName) UnmarshalJSON(b []byte) error {
	type plain CharacterSpellName
	return requiredHome(b, (*plain)(s))
}

func (s *CharacterSpells) TextIDs() []string {
	ids := []string{s.HeaderTextID, s.RowTextID, s.FooterTextID}
	for _, name := range s.Catalog {
		ids = append(ids, name.TextID)
	}
	return ids
}

// OrderedTextIDs validates the entire input before returning a deduplicated page.
// It never calls a character's learning rules or changes the supplied slice.
func (s *CharacterSpells) OrderedTextIDs(records []int) ([]string, bool) {
	if s == nil {
		return nil, false
	}
	wanted := make(map[int]bool, len(records))
	for _, record := range records {
		wanted[record] = true
	}
	var ids []string
	for _, name := range s.Catalog {
		if wanted[name.RecordRaw] {
			ids = append(ids, name.TextID)
			delete(wanted, name.RecordRaw)
		}
	}
	return ids, len(wanted) == 0
}

func (p *Pack) validateCharacterSpells() error {
	s, geometry := p.Interface.CharacterSpells, p.Interface.NewGameGeometry
	if s == nil || geometry == nil || geometry.Raster == nil || s.Evidence.Level != "D3" {
		return fmt.Errorf("character spell page is missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	w, a := s.RawWindow, s.Names
	if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 3 || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		s.Columns <= 0 || s.Columns > 640/dq3data.GlyphPx || s.ExtraRows != 2 || len(s.Catalog) == 0 || len(s.Catalog) > 256 || a.StepX <= 0 || a.StepY != dq3data.GlyphPx {
		return fmt.Errorf("character spell page geometry is invalid")
	}
	rows := (len(s.Catalog) + s.Columns - 1) / s.Columns
	height := (rows + s.ExtraRows) * a.StepY
	shadow := geometry.Raster.ShadowOffset
	if w.X*8+w.Width*8+shadow.X > 640 || w.Y+height+shadow.Y > 350 || a.X < w.X*8 || a.Y != w.Y+a.StepY ||
		a.X+(s.Columns-1)*a.StepX+dq3data.GlyphPx > (w.X+w.Width)*8 || a.Y+(rows-1)*a.StepY+dq3data.GlyphPx > w.Y+height-a.StepY {
		return fmt.Errorf("character spell page exceeds its canvas")
	}
	for _, old := range geometry.RawWindows {
		if old.ID == w.ID {
			return fmt.Errorf("duplicate character spell window")
		}
	}
	for _, old := range []string{p.Interface.Registration.ClassMenu.RawWindow.ID, p.Interface.RecruitmentEntry.Menu.RawWindow.ID, p.Interface.RecruitmentSelection.RawWindow.ID} {
		if old == w.ID {
			return fmt.Errorf("duplicate character spell window")
		}
	}
	plain := func(id string, max int, level string) error {
		text, ok := p.TextDefinition(id)
		if !ok || text.Source.Kind != "legacy_record" || text.Source.Record == nil || text.Evidence.Level != level || len(text.GlyphCodes) == 0 || len(text.GlyphCodes) > max {
			return fmt.Errorf("character spell text %q is invalid", id)
		}
		for _, code := range text.GlyphCodes {
			if code < 0 || code >= dq3data.GlyphMax {
				return fmt.Errorf("character spell text %q contains control words", id)
			}
		}
		return validateEvidence(text.Evidence)
	}
	for _, id := range []string{s.HeaderTextID, s.RowTextID, s.FooterTextID} {
		if err := plain(id, w.Width*8/dq3data.GlyphPx, "D3"); err != nil {
			return err
		}
		codes, _ := p.TextGlyphCodes(id)
		if len(codes)*dq3data.GlyphPx != w.Width*8 {
			return fmt.Errorf("character spell frame width differs")
		}
	}
	seen := map[int]bool{}
	for _, name := range s.Catalog {
		if name.RecordRaw < 0 || seen[name.RecordRaw] || name.Evidence.Level != "D2" {
			return fmt.Errorf("character spell catalog is invalid")
		}
		seen[name.RecordRaw] = true
		if err := validateEvidence(name.Evidence); err != nil {
			return err
		}
		if err := plain(name.TextID, a.StepX/dq3data.GlyphPx, "D2"); err != nil {
			return err
		}
		text, _ := p.TextDefinition(name.TextID)
		if *text.Source.Record != name.RecordRaw {
			return fmt.Errorf("character spell catalog text record differs")
		}
		if a.X+(s.Columns-1)*a.StepX+len(text.GlyphCodes)*dq3data.GlyphPx > (w.X+w.Width)*8 {
			return fmt.Errorf("character spell name exceeds window")
		}
	}
	return nil
}
