package gamepack

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// READY native no-target response: docs/188; JSON contract: docs/84.
type FieldTalk struct {
	EmptyTextID   string               `json:"empty_text_id"`
	Presentation  *FieldItemPrompt     `json:"presentation"`
	TextFlow      RetainedTextFlow     `json:"text_flow"`
	WaitIndicator OpeningWaitIndicator `json:"wait_indicator"`
	Scope         string               `json:"scope"`
	ReturnMode    string               `json:"return_mode"`
	Evidence      Evidence             `json:"evidence"`
}

func (s *FieldTalk) UnmarshalJSON(b []byte) error {
	type plain FieldTalk
	return requiredHome(b, (*plain)(s))
}
func (p *Pack) validateFieldTalk() error {
	s := p.Interface.FieldTalk
	if s == nil || s.Scope != "healthy_single_member_town_no_target" || s.ReturnMode != "fresh_key_to_field" || s.Evidence.Level != "D3" {
		return fmt.Errorf("field talk response missing or unreviewed")
	}
	if err := p.validateFieldMessageResponse(s.EmptyTextID, s.Presentation, s.TextFlow, s.WaitIndicator, s.Evidence, nil); err != nil {
		return err
	}
	codes, _ := p.TextGlyphCodes(s.EmptyTextID)
	rows := 1
	for _, c := range codes {
		if c == dq3data.TxtNL || c == dq3data.TxtNL2 {
			rows++
		}
		if c == dq3data.TxtPage {
			return fmt.Errorf("field talk response has unreviewed pagination")
		}
	}
	if rows > s.Presentation.Window.LinesPerPage {
		return fmt.Errorf("field talk response capacity differs")
	}
	return nil
}
