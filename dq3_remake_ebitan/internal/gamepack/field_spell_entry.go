package gamepack

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// READY scope and native source: docs/188; JSON contract: docs/84.
type FieldSpellEntry struct {
	EmptyTextID       string               `json:"empty_text_id"`
	Presentation      *FieldItemPrompt     `json:"presentation"`
	TextFlow          RetainedTextFlow     `json:"text_flow"`
	WaitIndicator     OpeningWaitIndicator `json:"wait_indicator"`
	ActorVariableCode *int                 `json:"actor_variable_code"`
	Scope             string               `json:"scope"`
	ReturnMode        string               `json:"return_mode"`
	Evidence          Evidence             `json:"evidence"`
}

func (s *FieldSpellEntry) UnmarshalJSON(b []byte) error {
	type plain FieldSpellEntry
	return requiredHome(b, (*plain)(s))
}

func (p *Pack) validateFieldSpellEntry() error {
	s := p.Interface.FieldSpellEntry
	if s == nil || s.Scope != "single_member_no_spells" || s.ReturnMode != "fresh_key_to_field" || s.Evidence.Level != "D3" || s.ActorVariableCode == nil || !dq3data.IsVarInsert(uint16(*s.ActorVariableCode)) || *s.ActorVariableCode < 0 || *s.ActorVariableCode > 65535 {
		return fmt.Errorf("field spell entry missing or unreviewed")
	}
	if err := p.validateFieldMessageResponse(s.EmptyTextID, s.Presentation, s.TextFlow, s.WaitIndicator, s.Evidence, map[int]bool{*s.ActorVariableCode: true}); err != nil {
		return err
	}
	codes, _ := p.TextGlyphCodes(s.EmptyTextID)
	count := 0
	for _, c := range codes {
		if int(c) == *s.ActorVariableCode {
			count++
		}
	}
	if count != 1 {
		return fmt.Errorf("field spell entry actor binding differs")
	}
	return nil
}
