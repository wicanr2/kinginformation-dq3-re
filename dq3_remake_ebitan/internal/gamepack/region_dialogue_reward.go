package gamepack

import "fmt"

// RegionDialogueRewardEvent is a finite step-triggered dialogue followed by
// ordered rewards. ProgressFlagRaw belongs to remake saves, not native flags.
type RegionDialogueRewardEvent struct {
	ID                    string       `json:"id"`
	CTYRaw                int          `json:"cty_raw"`
	Section               int          `json:"section"`
	HandlerRaw            int          `json:"handler_raw"`
	Tile                  HomePoint    `json:"tile"`
	RequiredFlagRaw       int          `json:"required_flag_raw"`
	ClearFlagRaw          int          `json:"clear_flag_raw"`
	SetFlagRaw            int          `json:"set_flag_raw"`
	ProgressFlagRaw       int          `json:"progress_flag_raw"`
	ItemRawIDs            []int        `json:"item_raw_ids"`
	Gold                  int          `json:"gold"`
	TextID                string       `json:"text_id"`
	PresentationID        string       `json:"presentation_id"`
	Shadow                WindowShadow `json:"shadow"`
	Evidence              Evidence     `json:"evidence"`
	CompatibilityEvidence Evidence     `json:"compatibility_evidence"`
}

func (e *RegionDialogueRewardEvent) UnmarshalJSON(b []byte) error {
	type plain RegionDialogueRewardEvent
	return requiredHome(b, (*plain)(e))
}

func (p *Pack) validateRegionDialogueRewards() error {
	if p.Events.RegionDialogueRewardEvents == nil {
		return fmt.Errorf("region_dialogue_reward_events must be present")
	}
	ids, scenes := map[string]bool{}, map[[4]int]bool{}
	for _, e := range p.Events.RegionDialogueRewardEvents {
		key := [4]int{e.CTYRaw, e.Section, e.Tile.X, e.Tile.Y}
		prelude := p.Interface.OpeningPrelude
		if e.ID == "" || ids[e.ID] || scenes[key] || e.CTYRaw < 0 || e.Section < 0 || e.HandlerRaw < 0 || e.HandlerRaw > 255 || e.Tile.X < 0 || e.Tile.Y < 0 ||
			p.SceneCamera(e.CTYRaw, e.Section) == nil || e.RequiredFlagRaw < 0 || e.RequiredFlagRaw >= 512 ||
			e.ClearFlagRaw < 0 || e.ClearFlagRaw >= 512 || e.SetFlagRaw < 0 || e.SetFlagRaw >= 512 ||
			e.SetFlagRaw == e.ClearFlagRaw || e.ProgressFlagRaw < 512 || e.ProgressFlagRaw > 65535 ||
			len(e.ItemRawIDs) == 0 || e.Gold < 0 || prelude == nil || e.PresentationID != prelude.ID ||
			prelude.ReturnMode != "automatic_after_reveal" || e.TextID == "" ||
			e.Shadow.Mode != "vga_word_latch_and" || e.Shadow.OffsetX < 0 || e.Shadow.OffsetX%8 != 0 || e.Shadow.OffsetY < 0 ||
			prelude.Window.X+prelude.Window.Width+e.Shadow.OffsetX > 640 || prelude.Window.Y+prelude.Window.Height+e.Shadow.OffsetY > 350 ||
			e.Evidence.Level != "D3" || e.Shadow.Evidence.Level != "D3" || e.CompatibilityEvidence.Level != "D2" || e.CompatibilityEvidence.SourceKind != "engine" {
			return fmt.Errorf("region dialogue reward %q has invalid contract", e.ID)
		}
		for _, raw := range e.ItemRawIDs {
			if raw < 0 || raw >= 255 {
				return fmt.Errorf("region dialogue reward item exceeds native item namespace")
			}
		}
		for _, evidence := range []Evidence{e.Evidence, e.Shadow.Evidence, e.CompatibilityEvidence} {
			if err := validateEvidence(evidence); err != nil {
				return err
			}
		}
		ids[e.ID], scenes[key] = true, true
	}
	return nil
}

func (p *Pack) validateRegionDialogueRewardRefs() error {
	for _, e := range p.Events.RegionDialogueRewardEvents {
		d, ok := p.TextDefinition(e.TextID)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("region dialogue reward %q text reference is invalid", e.ID)
		}
	}
	return nil
}
