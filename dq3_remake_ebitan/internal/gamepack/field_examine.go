package gamepack

import (
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldExamine describes the reviewed no-result response, not event rules.
// READY evidence: docs/188; JSON contract: docs/84.
type FieldExamine struct {
	IntroTextID       string               `json:"intro_text_id"`
	ResultTextID      string               `json:"result_text_id"`
	RecordJoin        string               `json:"record_join"`
	Presentation      *FieldItemPrompt     `json:"presentation"`
	TextFlow          RetainedTextFlow     `json:"text_flow"`
	WaitIndicator     OpeningWaitIndicator `json:"wait_indicator"`
	ActorVariableCode *int                 `json:"actor_variable_code"`
	Scope             string               `json:"scope"`
	ReturnMode        string               `json:"return_mode"`
	Evidence          Evidence             `json:"evidence"`
}

func (s *FieldExamine) UnmarshalJSON(b []byte) error {
	type plain FieldExamine
	return requiredHome(b, (*plain)(s))
}

func (p *Pack) validateFieldExamine() error {
	s := p.Interface.FieldExamine
	if s == nil || s.Scope != "healthy_single_member_no_result" || s.ReturnMode != "fresh_key_to_field" || s.RecordJoin != "new_line" || s.Evidence.Level != "D3" || s.ActorVariableCode == nil || *s.ActorVariableCode < 0 || *s.ActorVariableCode > 65535 || !dq3data.IsVarInsert(uint16(*s.ActorVariableCode)) {
		return fmt.Errorf("field examine response missing or unreviewed")
	}
	for _, id := range []string{s.IntroTextID, s.ResultTextID} {
		if err := p.validateFieldMessageResponse(id, s.Presentation, s.TextFlow, s.WaitIndicator, s.Evidence, map[int]bool{*s.ActorVariableCode: true}); err != nil {
			return err
		}
	}
	intro, _ := p.TextGlyphCodes(s.IntroTextID)
	result, _ := p.TextGlyphCodes(s.ResultTextID)
	bindings, rows := 0, 2
	for i, codes := range [][]uint16{intro, result} {
		for _, code := range codes {
			if int(code) == *s.ActorVariableCode {
				if i != 0 {
					return fmt.Errorf("field examine result actor binding differs")
				}
				bindings++
			}
			if code == dq3data.TxtNL || code == dq3data.TxtNL2 {
				rows++
			}
			if code == dq3data.TxtPage {
				return fmt.Errorf("field examine response has unreviewed pagination")
			}
		}
	}
	if bindings != 1 || rows > s.Presentation.Window.LinesPerPage {
		return fmt.Errorf("field examine actor binding or combined capacity differs")
	}
	return nil
}
