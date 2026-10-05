package gamepack

import (
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// FieldStatusReorder describes the accepted single-member response. Party
// pointer transactions require a separate reviewed contract and receipt.
type FieldStatusReorder struct {
	SingleMemberTextID string               `json:"single_member_text_id"`
	Presentation       *FieldItemPrompt     `json:"presentation"`
	TextFlow           RetainedTextFlow     `json:"text_flow"`
	WaitIndicator      OpeningWaitIndicator `json:"wait_indicator"`
	Scope              string               `json:"scope"`
	ReturnMode         string               `json:"return_mode"`
	Evidence           Evidence             `json:"evidence"`
}

func (s *FieldStatusReorder) UnmarshalJSON(b []byte) error {
	type plain FieldStatusReorder
	return requiredHome(b, (*plain)(s))
}

func (p *Pack) validateFieldStatusReorder() error {
	s := p.Interface.FieldStatusMenu.Reorder
	if s == nil || s.Scope != "single_member" || s.ReturnMode != "fresh_key_to_field" || s.Evidence.Level != "D3" {
		return fmt.Errorf("status reorder response missing or unreviewed")
	}
	return p.validateFieldMessageResponse(s.SingleMemberTextID, s.Presentation, s.TextFlow, s.WaitIndicator, s.Evidence, nil)
}

func (p *Pack) validateFieldMessageResponse(textID string, presentation *FieldItemPrompt, flow RetainedTextFlow, indicator OpeningWaitIndicator, evidence Evidence, variables map[int]bool) error {
	if err := validateEvidence(evidence); err != nil {
		return err
	}
	if err := p.validateNativeFieldPrompt(presentation); err != nil {
		return err
	}
	f, w := flow, presentation.Window
	if f.Mode != "retained_rows" || f.Evidence.Level != "D3" ||
		f.ScrollStepPixels <= 0 || f.ScrollStepPixels > dq3data.GlyphPx ||
		f.ScrollSteps <= 0 || f.ScrollSteps > dq3data.GlyphPx ||
		f.ScrollStepPixels*f.ScrollSteps != dq3data.GlyphPx ||
		f.ScrollHoldFrames <= 0 || f.ScrollHoldFrames > 60 ||
		w.Height-2*w.TextInsetY != w.LinesPerPage*dq3data.GlyphPx || (w.Width-2*w.TextInsetX)%8 != 0 {
		return fmt.Errorf("status reorder retained text flow invalid")
	}
	if err := validateEvidence(f.Evidence); err != nil {
		return err
	}
	if err := validateOpeningWaitIndicator(indicator, w); err != nil {
		return err
	}
	d, ok := p.TextDefinition(textID)
	if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
		return fmt.Errorf("status reorder text reference missing or unreviewed")
	}
	if err := validateEvidence(d.Evidence); err != nil {
		return err
	}
	column := 0
	for _, c := range d.GlyphCodes {
		if c == int(dq3data.TxtNL) || c == int(dq3data.TxtNL2) || c == int(dq3data.TxtPage) {
			column = 0
			continue
		}
		if variables[c] {
			column++
			continue
		}
		if c < 0 || c >= dq3data.GlyphMax || column >= w.Columns ||
			w.TextInsetX+column*presentation.GlyphStepX+dq3data.GlyphPx > w.Width-w.TextInsetX {
			return fmt.Errorf("status reorder text control or row invalid")
		}
		column++
	}
	return nil
}
