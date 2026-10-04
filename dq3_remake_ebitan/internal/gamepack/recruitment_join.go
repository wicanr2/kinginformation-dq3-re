package gamepack

import (
	"encoding/json"
	"fmt"
	"io/fs"
)

// RecruitmentJoin is a fixed presentation contract, not an executable script.
// The finite evidence and timing approximation are reviewed in docs/188.
type RecruitmentJoin struct {
	ID                   string               `json:"id"`
	EntryID              string               `json:"entry_id"`
	JoinedTextID         string               `json:"joined_text_id"`
	LeaderTextID         string               `json:"leader_text_id"`
	FinishTextID         string               `json:"finish_text_id"`
	LeaderNameRole       string               `json:"leader_name_role"`
	NameControlCode      uint16               `json:"name_control_code"`
	RetainCallerBackdrop bool                 `json:"retain_caller_backdrop"`
	Sound                RecruitmentJoinSound `json:"sound"`
	Evidence             Evidence             `json:"evidence"`
}

type RecruitmentJoinSound struct {
	SourceAsset    string   `json:"source_asset"`
	Start          int      `json:"start"`
	End            int      `json:"end"`
	EventCount     int      `json:"event_count"`
	DeltaTicks     int64    `json:"delta_ticks"`
	ClockDivisor   int64    `json:"clock_divisor"`
	ReferenceHz    int64    `json:"reference_hz"`
	RenderFile     string   `json:"render_file"`
	RenderSize     int64    `json:"render_size"`
	RenderSHA256   string   `json:"render_sha256"`
	TimingEvidence Evidence `json:"timing_evidence"`
}

func (s *RecruitmentJoin) UnmarshalJSON(b []byte) error {
	type plain RecruitmentJoin
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(b, &fields); err != nil {
		return err
	}
	type sound RecruitmentJoinSound
	return requiredHome(fields["sound"], (*sound)(&s.Sound))
}

func (s RecruitmentJoin) TextIDs() []string {
	return []string{s.JoinedTextID, s.LeaderTextID, s.FinishTextID}
}

func (s RecruitmentJoinSound) HoldFrames() int {
	return int((s.DeltaTicks*s.ClockDivisor*60 + s.ReferenceHz - 1) / s.ReferenceHz)
}

func (p *Pack) validateRecruitmentJoin() error {
	s, e := p.Interface.RecruitmentJoin, p.Interface.RecruitmentEntry
	if s == nil || e == nil || s.ID == "" || s.EntryID != e.ID || s.LeaderNameRole != "primary_actor" || !s.RetainCallerBackdrop || s.Evidence.Level != "D3" {
		return fmt.Errorf("recruitment join contract is missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	// This presentation primitive binds the named entity insertion control.
	// A pack may not route an unrelated control through the same primitive.
	if s.NameControlCode != 0xfffb {
		return fmt.Errorf("unsupported entity name control")
	}
	seen := map[string]bool{}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || seen[id] || d.Evidence.Level != "D3" || d.Source.Kind != "legacy_record" || d.Source.Record == nil || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("invalid recruitment join text %q", id)
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return err
		}
		seen[id] = true
		found := false
		for _, code := range d.GlyphCodes {
			if code == int(s.NameControlCode) {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("recruitment join text omits entity name")
		}
	}
	a := s.Sound
	asset, ok := p.Manifest.Assets[a.SourceAsset]
	if !ok || a.Start < 0 || a.End <= a.Start || int64(a.End) > asset.Size || a.EventCount <= 0 || a.EventCount > 100000 || a.DeltaTicks <= 0 || a.DeltaTicks > 1000000 || a.ClockDivisor <= 0 || a.ClockDivisor > 1000000 || a.ReferenceHz <= 0 || a.ReferenceHz > 1000000000 || a.DeltaTicks*a.ClockDivisor/a.ReferenceHz > 3600 || !fs.ValidPath(a.RenderFile) || a.RenderSize <= 0 || len(a.RenderSHA256) != 64 {
		return fmt.Errorf("invalid recruitment join sound")
	}
	for _, c := range a.RenderSHA256 {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return fmt.Errorf("invalid recruitment sound hash")
		}
	}
	if a.TimingEvidence.Level != "D2" || a.HoldFrames() <= 0 {
		return fmt.Errorf("unreviewed recruitment sound timing")
	}
	return validateEvidence(a.TimingEvidence)
}
