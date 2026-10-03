package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// Registration 是 docs/188 已審查的有限登錄契約。流程由引擎的具名狀態執行。
type Registration struct {
	ID                       string              `json:"id"`
	Binding                  RegistrationBinding `json:"binding"`
	PresentationID           string              `json:"presentation_id"`
	GeometryID               string              `json:"geometry_id"`
	Shadow                   WindowShadow        `json:"shadow"`
	TextRoles                RegistrationTexts   `json:"text_roles"`
	ClassOptions             []RegistrationClass `json:"class_options"`
	ClassMenu                RegistrationMenu    `json:"class_menu"`
	RequireNonemptyName      bool                `json:"require_nonempty_name"`
	ReviewBeforeConfirmation bool                `json:"review_before_confirmation"`
	RetainTextBetweenRecords bool                `json:"retain_text_between_records"`
	RosterCapacity           int                 `json:"roster_capacity"`
	Evidence                 Evidence            `json:"evidence"`
}
type RegistrationBinding struct {
	CTYRaw        int `json:"cty_raw"`
	Section       int `json:"section"`
	NPCHandlerRaw int `json:"npc_handler_raw"`
}
type RegistrationTexts struct {
	Greeting          string `json:"greeting"`
	InitialDecline    string `json:"initial_decline"`
	NamePrompt        string `json:"name_prompt"`
	CreationCancelled string `json:"creation_cancelled"`
	Registered        string `json:"registered"`
	Farewell          string `json:"farewell"`
}
type RegistrationClass struct {
	ClassRaw int    `json:"class_raw"`
	TextID   string `json:"text_id"`
}
type RegistrationMenu struct {
	RawWindow    RawNewGameWindow `json:"raw_window"`
	WindowTextID string           `json:"window_text_id"`
	Cursor       GeometryAnchor   `json:"cursor"`
	HitRect      GeometryRect     `json:"hit_rect"`
}

func (s *Registration) UnmarshalJSON(b []byte) error {
	type plain Registration
	return requiredHome(b, (*plain)(s))
}
func (s *RegistrationBinding) UnmarshalJSON(b []byte) error {
	type plain RegistrationBinding
	return requiredHome(b, (*plain)(s))
}
func (s *RegistrationTexts) UnmarshalJSON(b []byte) error {
	type plain RegistrationTexts
	return requiredHome(b, (*plain)(s))
}
func (s *RegistrationClass) UnmarshalJSON(b []byte) error {
	type plain RegistrationClass
	return requiredHome(b, (*plain)(s))
}
func (s *RegistrationMenu) UnmarshalJSON(b []byte) error {
	type plain RegistrationMenu
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
	if err := decodeOpeningObject(fields["cursor"], &s.Cursor, []string{"x", "y", "step_x", "step_y"}); err != nil {
		return err
	}
	return decodeOpeningObject(fields["hit_rect"], &s.HitRect, []string{"x", "y", "width", "height"})
}
func (s Registration) TextIDs() []string {
	t := s.TextRoles
	ids := []string{t.Greeting, t.InitialDecline, t.NamePrompt, t.CreationCancelled, t.Registered, t.Farewell, s.ClassMenu.WindowTextID}
	for _, c := range s.ClassOptions {
		ids = append(ids, c.TextID)
	}
	return ids
}
func (p *Pack) validateRegistration() error {
	s := p.Interface.Registration
	if s == nil {
		return fmt.Errorf("registration is required")
	}
	g := p.Interface.NewGameGeometry
	e := p.Interface.OpeningPrelude
	if s.ID == "" || g == nil || g.Raster == nil || e == nil || s.GeometryID != g.ID || s.PresentationID != e.ID ||
		!s.RequireNonemptyName || !s.ReviewBeforeConfirmation || !s.RetainTextBetweenRecords ||
		s.RosterCapacity <= 0 || s.RosterCapacity > 255 || len(s.ClassOptions) == 0 || len(s.ClassOptions) > 8 ||
		s.Binding.CTYRaw < 0 || s.Binding.Section < 0 || s.Binding.NPCHandlerRaw < 0 || s.Binding.NPCHandlerRaw > 255 ||
		s.Shadow.Mode != "vga_word_latch_and" || s.Shadow.OffsetX < 0 || s.Shadow.OffsetX%8 != 0 || s.Shadow.OffsetY < 0 || s.Evidence.Level != "D3" {
		return fmt.Errorf("registration contract is invalid")
	}
	if s.Shadow.OffsetX != g.Raster.ShadowOffset.X || s.Shadow.OffsetY != g.Raster.ShadowOffset.Y {
		return fmt.Errorf("registration shadow differs from referenced raster")
	}
	for _, v := range []Evidence{s.Evidence, s.Shadow.Evidence} {
		if v.Level != "D3" {
			return fmt.Errorf("registration evidence is unreviewed")
		}
		if err := validateEvidence(v); err != nil {
			return err
		}
	}
	seen := map[int]bool{}
	for _, c := range s.ClassOptions {
		if c.ClassRaw < 0 || c.ClassRaw >= 8 || seen[c.ClassRaw] {
			return fmt.Errorf("registration class mapping is invalid")
		}
		seen[c.ClassRaw] = true
		codes, ok := p.TextGlyphCodes(c.TextID)
		if !ok || len(codes) == 0 || g.StatsHero.X+len(codes)*dq3data.GlyphPx > 640 {
			return fmt.Errorf("registration class text is outside canvas")
		}
		for _, v := range codes {
			if v >= dq3data.GlyphMax {
				return fmt.Errorf("registration class text must contain only glyphs")
			}
		}
	}
	m := s.ClassMenu
	w := m.RawWindow
	a := m.Cursor
	h := m.HitRect
	last := len(s.ClassOptions) - 1
	if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 3 || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		w.X*8+w.Width*8+s.Shadow.OffsetX > 640 || w.Y+w.Height+s.Shadow.OffsetY > 350 ||
		a.X < w.X*8 || a.Y < w.Y || a.StepY <= 0 || a.X+16 > w.X*8+w.Width*8 || a.Y+last*a.StepY+16 > w.Y+w.Height ||
		h.X < w.X*8 || h.Y < w.Y || h.Width <= 0 || h.Height <= 0 || h.X+h.Width > w.X*8+w.Width*8 || h.Y+last*a.StepY+h.Height > w.Y+w.Height {
		return fmt.Errorf("registration class menu is outside its window")
	}
	for _, old := range g.RawWindows {
		if old.ID == w.ID {
			return fmt.Errorf("registration window id duplicates geometry")
		}
	}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || len(d.GlyphCodes) == 0 || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Source.File == "" || d.Evidence.Level != "D3" {
			return fmt.Errorf("registration text %q is unreviewed", id)
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return err
		}
	}
	return nil
}
