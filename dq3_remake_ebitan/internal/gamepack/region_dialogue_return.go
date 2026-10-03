package gamepack

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// RegionDialogueReturnEvent binds a native scene handler to a finite actor
// turn, automatic dialogue and collision-checked return step.
type RegionDialogueReturnEvent struct {
	ID               string       `json:"id"`
	CTYRaw           int          `json:"cty_raw"`
	Section          int          `json:"section"`
	HandlerRaw       int          `json:"handler_raw"`
	RequiredFlagRaw  int          `json:"required_flag_raw"`
	ActorRecordRaw   int          `json:"actor_record_raw"`
	ActorFacing      int          `json:"actor_facing"`
	TurnHoldFrames   int          `json:"turn_hold_frames"`
	ReturnDirection  int          `json:"return_direction"`
	ReturnHoldFrames int          `json:"return_hold_frames"`
	TextID           string       `json:"text_id"`
	PresentationID   string       `json:"presentation_id"`
	Shadow           WindowShadow `json:"shadow"`
	Evidence         Evidence     `json:"evidence"`
	TimingEvidence   Evidence     `json:"timing_evidence"`
}

func (e *RegionDialogueReturnEvent) UnmarshalJSON(b []byte) error {
	type plain RegionDialogueReturnEvent
	return requiredHome(b, (*plain)(e))
}

func (p *Pack) validateRegionDialogueReturns() error {
	if p.Events.RegionDialogueReturnEvents == nil {
		return fmt.Errorf("region_dialogue_return_events must be present")
	}
	ids, bindings := map[string]bool{}, map[[3]int]bool{}
	for _, e := range p.Events.RegionDialogueReturnEvents {
		key := [3]int{e.CTYRaw, e.Section, e.HandlerRaw}
		prelude := p.Interface.OpeningPrelude
		if e.ID == "" || ids[e.ID] || bindings[key] || e.CTYRaw < 0 || e.Section < 0 || p.SceneCamera(e.CTYRaw, e.Section) == nil || e.HandlerRaw < 0 || e.HandlerRaw > 255 ||
			e.RequiredFlagRaw < 0 || e.RequiredFlagRaw >= 512 || e.ActorRecordRaw < 0 || e.ActorFacing < 0 || e.ActorFacing > 3 ||
			e.ReturnDirection < 0 || e.ReturnDirection > 3 || e.TurnHoldFrames < 0 || e.ReturnHoldFrames < 0 ||
			prelude == nil || e.PresentationID != prelude.ID || prelude.ReturnMode != "automatic_after_reveal" ||
			e.TextID == "" || e.Shadow.Mode != "vga_word_latch_and" || e.Shadow.OffsetX < 0 || e.Shadow.OffsetX%8 != 0 || e.Shadow.OffsetY < 0 ||
			prelude.Window.X+prelude.Window.Width+e.Shadow.OffsetX > 640 || prelude.Window.Y+prelude.Window.Height+e.Shadow.OffsetY > 350 ||
			e.Evidence.Level != "D3" || e.Shadow.Evidence.Level != "D3" || e.TimingEvidence.Level != "D2" {
			return fmt.Errorf("region dialogue return %q has invalid contract", e.ID)
		}
		for _, evidence := range []Evidence{e.Evidence, e.Shadow.Evidence, e.TimingEvidence} {
			if err := validateEvidence(evidence); err != nil {
				return err
			}
		}
		ids[e.ID], bindings[key] = true, true
	}
	return nil
}

func (p *Pack) validateRegionDialogueReturnRefs() error {
	for _, e := range p.Events.RegionDialogueReturnEvents {
		d, ok := p.TextDefinition(e.TextID)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("region dialogue return %q text reference is invalid", e.ID)
		}
		for _, code := range d.GlyphCodes {
			if code == int(dq3data.TxtPage) {
				return fmt.Errorf("region dialogue return %q requires automatic text without an inline input wait", e.ID)
			}
		}
	}
	return nil
}
