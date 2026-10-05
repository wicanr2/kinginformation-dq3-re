package gamepack

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldStatusMenu describes the reviewed first status selector. CallbackRaw
// retains original evidence and is never executed. READY scope: docs/188.
type FieldStatusMenu struct {
	ID          string             `json:"id"`
	RawWindow   RawNewGameWindow   `json:"raw_window"`
	TextID      string             `json:"text_id"`
	Navigation  string             `json:"navigation"`
	CursorGlyph int                `json:"cursor_glyph"`
	FontIndex   int                `json:"font_index"`
	Entries     []FieldStatusEntry `json:"entries"`
	Evidence    Evidence           `json:"evidence"`
}

type FieldStatusEntry struct {
	Role        string `json:"role"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
	CallbackRaw int    `json:"callback_raw"`
}

func (s *FieldStatusEntry) UnmarshalJSON(b []byte) error {
	type plain FieldStatusEntry
	return requiredHome(b, (*plain)(s))
}

func (s *FieldStatusMenu) UnmarshalJSON(b []byte) error {
	type plain FieldStatusMenu
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	return decodeOpeningObject(fields["raw_window"], &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"})
}

func (p *Pack) validateFieldStatusMenu() error {
	s := p.Interface.FieldStatusMenu
	cmd := p.Interface.FieldCommandMenu
	if s == nil || cmd == nil || s.ID == "" || s.Navigation != "cyclic_single_column" ||
		len(s.Entries) == 0 || s.FontIndex != cmd.FontIndex ||
		s.CursorGlyph != cmd.CursorGlyph || s.Evidence.Level != "D3" {
		return fmt.Errorf("field status menu missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	w := s.RawWindow
	d, ok := p.TextDefinition(s.TextID)
	geo := p.Interface.NewGameGeometry
	if !ok || geo == nil || geo.Raster == nil || d.Source.Kind != "legacy_record" ||
		d.Source.Record == nil || d.Evidence.Level != "D3" || w.ID == "" || w.Address == "" ||
		w.Flags < 0 || w.Flags > 7 || w.Flags&3 != 3 || w.X < 0 || w.Y < 0 ||
		w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		w.X*8+w.Width*8+geo.Raster.ShadowOffset.X > 640 ||
		w.Y+w.Height+geo.Raster.ShadowOffset.Y > 350 ||
		d.Layout.Columns*dq3data.GlyphPx != w.Width*8 ||
		d.Layout.LinesPerPage*dq3data.GlyphPx != w.Height {
		return fmt.Errorf("field status window or text shape invalid")
	}
	if err := validateEvidence(d.Evidence); err != nil {
		return err
	}
	col, row := 0, 1
	for _, code := range d.GlyphCodes {
		if code == int(dq3data.TxtNL) {
			if col != d.Layout.Columns {
				return fmt.Errorf("status text row shape differs")
			}
			col = 0
			row++
		} else {
			col++
			if code < 0 || code >= dq3data.GlyphMax || col > d.Layout.Columns {
				return fmt.Errorf("status text glyph outside window")
			}
		}
	}
	if row != d.Layout.LinesPerPage || col != d.Layout.Columns {
		return fmt.Errorf("status text height differs")
	}
	known := map[string]bool{"detail": true, "party_summary": true, "reorder": true}
	seen := map[string]bool{}
	for i, e := range s.Entries {
		if !known[e.Role] || seen[e.Role] || e.CallbackRaw < 0 || e.CallbackRaw > 65535 ||
			e.X < w.X*8 || e.Y < w.Y || e.X+dq3data.GlyphPx > (w.X+w.Width)*8 ||
			e.Y+dq3data.GlyphPx > w.Y+w.Height ||
			(i > 0 && (e.X != s.Entries[i-1].X || e.Y != s.Entries[i-1].Y+dq3data.GlyphPx)) {
			return fmt.Errorf("field status selector entry invalid")
		}
		seen[e.Role] = true
	}
	return nil
}
