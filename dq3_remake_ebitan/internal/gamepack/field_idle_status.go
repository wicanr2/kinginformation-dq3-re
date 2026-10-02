package gamepack

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldIdleStatus is the finite, source-reviewed field modal; see docs/188.
type FieldIdleStatus struct {
	ID               string               `json:"id"`
	Scenes           []FieldIdleScene     `json:"scenes"`
	PartyHUDID       string               `json:"party_hud_id"`
	FontAsset        string               `json:"font_asset"`
	LeftTextID       string               `json:"left_text_id"`
	ColumnTextID     string               `json:"column_text_id"`
	RightTextID      string               `json:"right_text_id"`
	DelayTicks       int                  `json:"delay_ticks"`
	RateNumerator    int                  `json:"rate_numerator"`
	RateDenominator  int                  `json:"rate_denominator"`
	Digits           int                  `json:"digits"`
	BlankGlyph       int                  `json:"blank_glyph"`
	StatusInsetX     int                  `json:"status_inset_x"`
	HPInsetY         int                  `json:"hp_inset_y"`
	MPInsetY         int                  `json:"mp_inset_y"`
	StatusInsetY     int                  `json:"status_inset_y"`
	Shadow           WindowShadow         `json:"shadow"`
	BackdropRGB      []uint8              `json:"backdrop_rgb"`
	StatusRows       []FieldIdleStatusRow `json:"status_rows"`
	DeadStatusMask   int                  `json:"dead_status_mask"`
	HealthDivisors   []int                `json:"health_divisors"`
	HealthSumCap     int                  `json:"health_sum_cap"`
	DeadPaletteIndex int                  `json:"dead_palette_index"`
	PaletteIndex     int                  `json:"palette_index"`
	HealthPaletteRaw [][]uint8            `json:"health_palette_raw"`
	Evidence         Evidence             `json:"evidence"`
	TimingEvidence   Evidence             `json:"timing_evidence"`
}
type FieldIdleScene struct {
	CTY     int `json:"cty"`
	Section int `json:"section"`
}
type FieldIdleStatusRow struct {
	MaskRaw int    `json:"mask_raw"`
	TextID  string `json:"text_id"`
}

func (s *FieldIdleStatus) UnmarshalJSON(b []byte) error {
	type plain FieldIdleStatus
	return requiredHome(b, (*plain)(s))
}
func (s *FieldIdleScene) UnmarshalJSON(b []byte) error {
	type plain FieldIdleScene
	return requiredHome(b, (*plain)(s))
}
func (s *FieldIdleStatusRow) UnmarshalJSON(b []byte) error {
	type plain FieldIdleStatusRow
	return requiredHome(b, (*plain)(s))
}
func (s FieldIdleStatus) HoldFrames() int {
	return int((int64(s.DelayTicks)*60*int64(s.RateDenominator) + int64(s.RateNumerator) - 1) / int64(s.RateNumerator))
}
func (s FieldIdleStatus) Enabled(cty, section int) bool {
	for _, b := range s.Scenes {
		if b.CTY == cty && b.Section == section {
			return true
		}
	}
	return false
}
func (s FieldIdleStatus) TextIDs() []string {
	ids := []string{s.LeftTextID, s.ColumnTextID, s.RightTextID}
	for _, row := range s.StatusRows {
		ids = append(ids, row.TextID)
	}
	return ids
}
func (p *Pack) validateFieldIdleStatus() error {
	s := p.Interface.FieldIdleStatus
	if s == nil {
		return nil
	} // Data fixtures may omit the modal; production requires it.
	h := p.Interface.PartyHUD
	if s.ID == "" || s.PartyHUDID != h.ID || s.FontAsset == "" || h.Columns <= 0 ||
		s.DelayTicks <= 0 || s.DelayTicks > 65535 || s.RateNumerator <= 0 || s.RateNumerator > 1000000000 ||
		s.RateDenominator <= 0 || s.RateDenominator > 1000000000 || s.Digits <= 0 || s.Digits > 9 ||
		s.BlankGlyph < 0 || s.BlankGlyph >= dq3data.GlyphMax || s.StatusInsetX < 0 || s.StatusInsetX%dq3data.GlyphPx != 0 ||
		len(s.BackdropRGB) != 3 || s.Shadow.Mode != "vga_word_latch_and" || s.Shadow.OffsetX < 0 || s.Shadow.OffsetX%8 != 0 || s.Shadow.OffsetY < 0 ||
		s.DeadStatusMask <= 0 || s.DeadStatusMask > 65535 || s.PaletteIndex < 0 || s.PaletteIndex >= 16 ||
		s.HealthSumCap < 0 || s.DeadPaletteIndex <= s.HealthSumCap || len(s.HealthPaletteRaw) != s.DeadPaletteIndex+1 || len(s.HealthDivisors) == 0 ||
		len(s.StatusRows) == 0 || len(s.Scenes) == 0 || s.Evidence.Level != "D3" || s.HoldFrames() > 3600 {
		return fmt.Errorf("field idle status contract is invalid")
	}
	if _, ok := p.Asset(s.FontAsset); !ok {
		return fmt.Errorf("field idle font reference is unknown")
	}
	for _, e := range []Evidence{s.Evidence, s.TimingEvidence, s.Shadow.Evidence} {
		if e.Level != "D2" && e.Level != "D3" {
			return fmt.Errorf("field idle evidence is unreviewed")
		}
		if err := validateEvidence(e); err != nil {
			return err
		}
	}
	seen := map[[2]int]bool{}
	for _, b := range s.Scenes {
		key := [2]int{b.CTY, b.Section}
		if b.CTY < 0 || b.Section < 0 || seen[key] || p.SceneCamera(b.CTY, b.Section) == nil {
			return fmt.Errorf("field idle scene reference is invalid")
		}
		seen[key] = true
	}
	masks := map[int]bool{}
	for _, row := range s.StatusRows {
		if row.MaskRaw <= 0 || row.MaskRaw > 65535 || masks[row.MaskRaw] || row.TextID == "" {
			return fmt.Errorf("field idle status row is invalid")
		}
		masks[row.MaskRaw] = true
	}
	for _, d := range s.HealthDivisors {
		if d <= 0 || d > 65535 {
			return fmt.Errorf("field idle health divisor is invalid")
		}
	}
	for _, rgb := range s.HealthPaletteRaw {
		if len(rgb) != 3 {
			return fmt.Errorf("field idle palette is incomplete")
		}
		for _, v := range rgb {
			if v > 63 {
				return fmt.Errorf("field idle palette exceeds VGA range")
			}
		}
	}
	return nil
}
func (p *Pack) validateFieldIdleStatusRefs() error {
	s := p.Interface.FieldIdleStatus
	if s == nil {
		return nil
	}
	h := p.Interface.PartyHUD
	for _, y := range []int{h.TextInsetY, s.HPInsetY, s.MPInsetY, s.StatusInsetY} {
		if y < 0 || y%dq3data.GlyphPx != 0 || y+dq3data.GlyphPx > h.Height {
			return fmt.Errorf("field idle row is outside window")
		}
	}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Source.File == "" || d.Layout.Kind != "menu_record" || (d.Evidence.Level != "D2" && d.Evidence.Level != "D3") {
			return fmt.Errorf("field idle text %q has invalid source", id)
		}
		columns, rows := 0, 1
		for _, code := range d.GlyphCodes {
			if code == dq3data.TxtNL {
				if columns != d.Layout.Columns {
					return fmt.Errorf("field idle text row shape mismatch")
				}
				columns = 0
				rows++
			} else if code < dq3data.GlyphMax {
				columns++
			} else {
				return fmt.Errorf("field idle text has unsupported control")
			}
		}
		if columns != d.Layout.Columns || rows != d.Layout.LinesPerPage {
			return fmt.Errorf("field idle text shape mismatch")
		}
	}
	l, _ := p.TextDefinition(s.LeftTextID)
	c, _ := p.TextDefinition(s.ColumnTextID)
	r, _ := p.TextDefinition(s.RightTextID)
	width := (l.Layout.Columns + c.Layout.Columns*h.Columns + r.Layout.Columns) * dq3data.GlyphPx
	if l.Layout.LinesPerPage*dq3data.GlyphPx != h.Height || c.Layout.LinesPerPage != l.Layout.LinesPerPage || r.Layout.LinesPerPage != l.Layout.LinesPerPage ||
		h.X < 0 || h.Y < 0 || h.X+width+s.Shadow.OffsetX > 640 || h.Y+h.Height+s.Shadow.OffsetY > 350 || width%16 != 0 ||
		h.TextInsetX+h.NameInsetX+h.NameMaxGlyphs*dq3data.GlyphPx > (l.Layout.Columns+c.Layout.Columns)*dq3data.GlyphPx ||
		h.TextInsetX+h.ValueInsetX+s.Digits*dq3data.GlyphPx > (l.Layout.Columns+c.Layout.Columns)*dq3data.GlyphPx {
		return fmt.Errorf("field idle window geometry is outside source shape")
	}
	for _, row := range s.StatusRows {
		d, _ := p.TextDefinition(row.TextID)
		if d.Layout.LinesPerPage != 1 || s.StatusInsetX+d.Layout.Columns*dq3data.GlyphPx > (l.Layout.Columns+c.Layout.Columns)*dq3data.GlyphPx {
			return fmt.Errorf("field idle status text exceeds column")
		}
	}
	return nil
}
