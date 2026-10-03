package gamepack

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

const (
	RecruitJoin  = "common:recruit.join"
	RecruitLeave = "common:recruit.leave"
	RecruitView  = "common:recruit.view"
)

// RecruitmentEntry 描述正常交談到首次選單。入隊後的文字另待證據閉合。
type RecruitmentEntry struct {
	ID              string              `json:"id"`
	Binding         RegistrationBinding `json:"binding"`
	PresentationID  string              `json:"presentation_id"`
	GeometryID      string              `json:"geometry_id"`
	Shadow          WindowShadow        `json:"shadow"`
	GreetingTextIDs []string            `json:"greeting_text_ids"`
	Menu            RegistrationMenu    `json:"menu"`
	OptionActions   []string            `json:"option_actions"`
	Evidence        Evidence            `json:"evidence"`
}

func (s *RecruitmentEntry) UnmarshalJSON(b []byte) error {
	type plain RecruitmentEntry
	return requiredHome(b, (*plain)(s))
}

func (s RecruitmentEntry) TextIDs() []string {
	return append(append([]string(nil), s.GreetingTextIDs...), s.Menu.WindowTextID)
}

func (p *Pack) validateRecruitmentEntry() error {
	s := p.Interface.RecruitmentEntry
	g := p.Interface.NewGameGeometry
	e := p.Interface.OpeningPrelude
	if s == nil || g == nil || g.Raster == nil || e == nil {
		return fmt.Errorf("recruitment entry is required")
	}
	if s.ID == "" || s.PresentationID != e.ID || s.GeometryID != g.ID || len(s.GreetingTextIDs) == 0 || len(s.OptionActions) == 0 || len(s.OptionActions) > 255 ||
		s.Binding.CTYRaw < 0 || s.Binding.Section < 0 || s.Binding.NPCHandlerRaw < 0 || s.Binding.NPCHandlerRaw > 255 ||
		s.Shadow.Mode != "vga_word_latch_and" || s.Shadow.OffsetX != g.Raster.ShadowOffset.X || s.Shadow.OffsetY != g.Raster.ShadowOffset.Y {
		return fmt.Errorf("recruitment entry contract is invalid")
	}
	for _, evidence := range []Evidence{s.Evidence, s.Shadow.Evidence} {
		if evidence.Level != "D3" {
			return fmt.Errorf("recruitment entry is unreviewed")
		}
		if err := validateEvidence(evidence); err != nil {
			return err
		}
	}
	seen := map[string]bool{}
	for _, action := range s.OptionActions {
		if seen[action] {
			return fmt.Errorf("duplicate recruitment entry action")
		}
		switch action {
		case RecruitJoin, RecruitLeave, RecruitView:
		default:
			return fmt.Errorf("unknown recruitment entry action")
		}
		seen[action] = true
	}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || len(d.GlyphCodes) == 0 || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Source.File == "" || d.Evidence.Level != "D3" {
			return fmt.Errorf("recruitment entry text %q is unreviewed", id)
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return err
		}
	}
	m, w := s.Menu, s.Menu.RawWindow
	a, h := m.Cursor, m.HitRect
	last := len(s.OptionActions) - 1
	if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 3 || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%2 != 0 || w.Height <= 0 ||
		w.X*8+w.Width*8+s.Shadow.OffsetX > 640 || w.Y+w.Height+s.Shadow.OffsetY > 350 ||
		a.X < w.X*8 || a.Y < w.Y || a.StepX != 0 || a.StepY <= 0 || a.X+dq3data.GlyphPx > w.X*8+w.Width*8 || a.Y+last*a.StepY+dq3data.GlyphPx > w.Y+w.Height ||
		h.X < w.X*8 || h.Y < w.Y || h.Width <= 0 || h.Height <= 0 || h.X+h.Width > w.X*8+w.Width*8 || h.Y+last*a.StepY+h.Height > w.Y+w.Height {
		return fmt.Errorf("recruitment entry menu is outside its window")
	}
	for _, old := range g.RawWindows {
		if old.ID == w.ID {
			return fmt.Errorf("recruitment entry window duplicates geometry")
		}
	}
	if p.Interface.Registration != nil && p.Interface.Registration.ClassMenu.RawWindow.ID == w.ID {
		return fmt.Errorf("recruitment entry window duplicates registration")
	}
	return nil
}
