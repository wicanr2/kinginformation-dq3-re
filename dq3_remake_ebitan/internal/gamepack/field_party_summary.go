package gamepack

import (
	"encoding/json"
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldPartySummary describes read-only horizontal actor columns. The original
// callback is evidence only; READY scope and source identities are in docs/188.
type FieldPartySummary struct {
	MoneyWindow  RawNewGameWindow `json:"money_window"`
	MoneyTextID  string           `json:"money_text_id"`
	MoneyValue   NumberField      `json:"money_value"`
	RawWindow    RawNewGameWindow `json:"raw_window"`
	LeftTextID   string           `json:"left_text_id"`
	ColumnTextID string           `json:"column_text_id"`
	RightTextID  string           `json:"right_text_id"`
	WidthBase    int              `json:"width_base"`
	ColumnStep   int              `json:"column_step"`
	MaxColumns   int              `json:"max_columns"`
	ColumnOrigin GeometryAnchor   `json:"column_origin"`
	Name         GeometryAnchor   `json:"name"`
	NameLimit    int              `json:"name_limit"`
	HP           NumberField      `json:"hp"`
	MaxHP        NumberField      `json:"max_hp"`
	MP           NumberField      `json:"mp"`
	MaxMP        NumberField      `json:"max_mp"`
	ReturnMode   string           `json:"return_mode"`
	Evidence     Evidence         `json:"evidence"`
}

func (s *FieldPartySummary) UnmarshalJSON(b []byte) error {
	type plain FieldPartySummary
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	for _, v := range []struct {
		key      string
		dest     any
		required []string
	}{
		{"money_window", &s.MoneyWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}},
		{"raw_window", &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}},
		{"money_value", &s.MoneyValue, []string{"x", "y", "digits"}},
		{"column_origin", &s.ColumnOrigin, []string{"x", "y", "step_x", "step_y"}},
		{"name", &s.Name, []string{"x", "y", "step_x", "step_y"}},
		{"hp", &s.HP, []string{"x", "y", "digits"}},
		{"max_hp", &s.MaxHP, []string{"x", "y", "digits"}},
		{"mp", &s.MP, []string{"x", "y", "digits"}},
		{"max_mp", &s.MaxMP, []string{"x", "y", "digits"}},
	} {
		if err := decodeOpeningObject(fields[v.key], v.dest, v.required); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pack) validateFieldPartySummary() error {
	s := p.Interface.FieldStatusMenu.Summary
	geo := p.Interface.NewGameGeometry
	if s == nil || geo == nil || geo.Raster == nil || s.ReturnMode != "fresh_key_to_field" || s.Evidence.Level != "D3" {
		return fmt.Errorf("party summary missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	for _, w := range []RawNewGameWindow{s.MoneyWindow, s.RawWindow} {
		if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 15 || w.X < 0 || w.X > 80 || w.Y < 0 || w.Y > 350 || w.Width <= 0 || w.Width > 80 || w.Width%2 != 0 || w.Height <= 0 || w.Height > 350 || w.Height%dq3data.GlyphPx != 0 ||
			(w.X+w.Width)*8+geo.Raster.ShadowOffset.X > 640 || w.Y+w.Height+geo.Raster.ShadowOffset.Y > 350 {
			return fmt.Errorf("party summary window outside canvas")
		}
	}
	w := s.RawWindow
	if s.MoneyWindow.ID == w.ID || s.WidthBase <= 0 || s.WidthBase%2 != 0 || s.ColumnStep <= 0 || s.ColumnStep > 640 || s.ColumnStep%dq3data.GlyphPx != 0 || s.MaxColumns <= 0 || s.MaxColumns > w.Width*8/s.ColumnStep ||
		s.WidthBase+s.ColumnStep/8*s.MaxColumns != w.Width || s.ColumnOrigin.X != w.X*8+dq3data.GlyphPx || s.ColumnOrigin.Y != w.Y || s.ColumnOrigin.StepX != s.ColumnStep || s.ColumnOrigin.StepY != 0 ||
		s.WidthBase*8 != s.ColumnOrigin.X-w.X*8+dq3data.GlyphPx ||
		s.Name.X < s.ColumnOrigin.X || s.Name.X > 640 || s.Name.Y < w.Y || s.Name.Y > 350 || s.Name.Y+dq3data.GlyphPx > w.Y+w.Height || s.Name.StepX != dq3data.GlyphPx || s.Name.StepY != 0 || s.NameLimit <= 0 || s.NameLimit > s.ColumnStep/s.Name.StepX || s.Name.X+(s.MaxColumns-1)*s.ColumnStep+s.NameLimit*s.Name.StepX > (w.X+w.Width)*8 {
		return fmt.Errorf("party summary columns invalid")
	}
	for _, v := range []struct {
		id            string
		columns, rows int
	}{
		{s.MoneyTextID, s.MoneyWindow.Width / 2, s.MoneyWindow.Height / dq3data.GlyphPx},
		{s.LeftTextID, 1, w.Height / dq3data.GlyphPx},
		{s.ColumnTextID, s.ColumnStep / dq3data.GlyphPx, w.Height / dq3data.GlyphPx},
		{s.RightTextID, 1, w.Height / dq3data.GlyphPx},
	} {
		d, ok := p.TextDefinition(v.id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || d.Layout.Columns != v.columns || d.Layout.LinesPerPage != v.rows {
			return fmt.Errorf("party summary text reference invalid")
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return err
		}
		col, rows := 0, 1
		for _, code := range d.GlyphCodes {
			if code == int(dq3data.TxtNL) {
				if col != v.columns {
					return fmt.Errorf("party summary text row invalid")
				}
				col = 0
				rows++
			} else {
				if code < 0 || code >= dq3data.GlyphMax {
					return fmt.Errorf("party summary glyph invalid")
				}
				col++
			}
		}
		if col != v.columns || rows != v.rows {
			return fmt.Errorf("party summary text shape invalid")
		}
	}
	inside := func(f NumberField, window RawNewGameWindow, last int) bool {
		return f.Digits > 0 && f.Digits <= 8 && f.X >= window.X*8 && f.Y >= window.Y && f.X+last*s.ColumnStep+f.Digits*dq3data.GlyphPx <= (window.X+window.Width)*8 && f.Y+dq3data.GlyphPx <= window.Y+window.Height
	}
	if !inside(s.MoneyValue, s.MoneyWindow, 0) {
		return fmt.Errorf("party summary money outside window")
	}
	for _, f := range []NumberField{s.HP, s.MaxHP, s.MP, s.MaxMP} {
		if !inside(f, w, s.MaxColumns-1) || f.Digits*dq3data.GlyphPx > s.ColumnStep {
			return fmt.Errorf("party summary value outside column")
		}
	}
	return nil
}
