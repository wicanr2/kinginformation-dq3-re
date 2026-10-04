package gamepack

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldCommandMenu describes the ordered choices consumed by the shared
// two-column selector. Original addresses and callback words are evidence only.
// READY contract and source identity: docs/188; JSON fields: docs/84.
type FieldCommandMenu struct {
	ID          string              `json:"id"`
	RawWindow   RawNewGameWindow    `json:"raw_window"`
	TextID      string              `json:"text_id"`
	Navigation  string              `json:"navigation"`
	CursorGlyph int                 `json:"cursor_glyph"`
	FontIndex   int                 `json:"font_index"`
	Entries     []FieldCommandEntry `json:"entries"`
	Evidence    Evidence            `json:"evidence"`
}

type FieldCommandEntry struct {
	Command     string `json:"command"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	CallbackRaw int    `json:"callback_raw"`
}

func (s *FieldCommandEntry) UnmarshalJSON(b []byte) error {
	type plain FieldCommandEntry
	return requiredHome(b, (*plain)(s))
}

func (s *FieldCommandMenu) UnmarshalJSON(b []byte) error {
	type plain FieldCommandMenu
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	return decodeOpeningObject(fields["raw_window"], &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"})
}

func (p *Pack) validateFieldCommandMenu() error {
	s := p.Interface.FieldCommandMenu
	geo := p.Interface.NewGameGeometry
	if s == nil || geo == nil || geo.Raster == nil || geo.Raster.FontIndex == nil ||
		s.ID == "" || s.Navigation != "linear_two_columns" || len(s.Entries) != 6 ||
		s.FontIndex != *geo.Raster.FontIndex || s.CursorGlyph < 0 || s.CursorGlyph >= dq3data.GlyphMax ||
		s.Evidence.Level != "D3" || p.Interface.FieldIdleStatus == nil {
		return fmt.Errorf("field command menu missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	w := s.RawWindow
	d, ok := p.TextDefinition(s.TextID)
	if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" ||
		w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 7 || w.Flags&3 != 3 ||
		w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		w.X*8+w.Width*8+geo.Raster.ShadowOffset.X > 640 || w.Y+w.Height+geo.Raster.ShadowOffset.Y > 350 ||
		d.Layout.Columns*dq3data.GlyphPx != w.Width*8 || d.Layout.LinesPerPage*dq3data.GlyphPx != w.Height {
		return fmt.Errorf("field command window or text shape invalid")
	}
	if err := validateEvidence(d.Evidence); err != nil {
		return err
	}
	column, row := 0, 1
	for _, code := range d.GlyphCodes {
		if code == int(dq3data.TxtNL) {
			if column != d.Layout.Columns {
				return fmt.Errorf("command text row shape differs")
			}
			column = 0
			row++
		} else {
			column++
			if code < 0 || code >= dq3data.GlyphMax || column > d.Layout.Columns {
				return fmt.Errorf("command text glyph outside window")
			}
		}
	}
	if row != d.Layout.LinesPerPage || column != d.Layout.Columns {
		return fmt.Errorf("command text height differs")
	}
	seen := map[string]bool{}
	known := map[string]bool{"talk": true, "spell": true, "status": true, "item": true, "equip": true, "examine": true}
	half := len(s.Entries) / 2
	for i, entry := range s.Entries {
		if !known[entry.Command] || seen[entry.Command] || entry.CallbackRaw < 0 || entry.CallbackRaw > 65535 ||
			entry.X < w.X*8 || entry.Y < w.Y || entry.X+3*dq3data.GlyphPx > (w.X+w.Width)*8 || entry.Y+dq3data.GlyphPx > w.Y+w.Height {
			return fmt.Errorf("field command entry invalid")
		}
		seen[entry.Command] = true
		if i%half != 0 && (entry.X != s.Entries[i-1].X || entry.Y != s.Entries[i-1].Y+dq3data.GlyphPx) {
			return fmt.Errorf("field command rows are not contiguous")
		}
		if i >= half && (entry.Y != s.Entries[i-half].Y || entry.X <= s.Entries[i-half].X) {
			return fmt.Errorf("field command columns are not aligned")
		}
	}
	return nil
}
