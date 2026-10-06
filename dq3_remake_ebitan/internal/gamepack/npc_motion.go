package gamepack

import "errors"

// NPCMotionDefinition supplies parameters for the finite automatic-wander
// primitive. It has no scene IDs, arbitrary expressions or executable code.
type NPCMotionDefinition struct {
	ViewportAnchor       *GeometryAnchor  `json:"viewport_anchor"`
	ViewportColumns      int              `json:"viewport_columns"`
	ViewportRows         int              `json:"viewport_rows"`
	Evaluation           NPCMotionRoll    `json:"evaluation"`
	DirectionBound       int              `json:"direction_bound"`
	DirectionMask        int              `json:"direction_mask"`
	Turn                 NPCMotionRoll    `json:"turn"`
	Step                 NPCMotionRoll    `json:"step"`
	MoveMask             int              `json:"move_mask"`
	FrozenMask           int              `json:"frozen_mask"`
	QuotientMask         int              `json:"quotient_mask"`
	TurnDeltaByLayer     []int            `json:"turn_delta_by_layer"`
	Directions           []GeometryAnchor `json:"directions"`
	MinimumAxisDistance  int              `json:"minimum_axis_distance"`
	BlockedAttributeMask int              `json:"blocked_attribute_mask"`
	Evidence             Evidence         `json:"evidence"`
	TurnLayerEvidence    Evidence         `json:"turn_layer_evidence"`
}

type NPCMotionRoll struct {
	Bound    int  `json:"bound"`
	Accepted *int `json:"accepted"`
}

func (p *Pack) validateNPCMotion() error {
	s := p.Characters.NPCMotion
	if s == nil || s.ViewportAnchor == nil || s.ViewportColumns <= 0 || s.ViewportColumns > 128 || s.ViewportRows <= 0 || s.ViewportRows > 128 || s.ViewportAnchor.X < 0 || s.ViewportAnchor.X >= s.ViewportColumns || s.ViewportAnchor.Y < 0 || s.ViewportAnchor.Y >= s.ViewportRows {
		return errors.New("npc_motion requires a bounded viewport")
	}
	for _, r := range []NPCMotionRoll{s.Evaluation, s.Turn, s.Step} {
		if r.Bound <= 0 || r.Bound > 65535 || r.Accepted == nil || *r.Accepted < 0 || *r.Accepted >= r.Bound {
			return errors.New("npc_motion roll is incomplete")
		}
	}
	if s.DirectionBound <= 0 || s.DirectionBound > 256 || s.DirectionBound&(s.DirectionBound-1) != 0 || s.DirectionMask != s.DirectionBound-1 || len(s.Directions) != s.DirectionBound {
		return errors.New("npc_motion direction domain is invalid")
	}
	seen := map[GeometryAnchor]bool{}
	for _, d := range s.Directions {
		if !((d.X == 0 && (d.Y == -1 || d.Y == 1)) || (d.Y == 0 && (d.X == -1 || d.X == 1))) || seen[d] {
			return errors.New("npc_motion directions must be unique cardinal steps")
		}
		seen[d] = true
	}
	if s.MoveMask <= 0 || s.MoveMask > 255 || s.MoveMask&(s.MoveMask-1) != 0 || s.FrozenMask <= 0 || s.FrozenMask > 255 || s.FrozenMask&(s.FrozenMask-1) != 0 || s.MoveMask&s.FrozenMask != 0 || (s.MoveMask|s.FrozenMask)&s.DirectionMask != 0 || s.QuotientMask <= 0 || s.QuotientMask > 65535 || s.MinimumAxisDistance <= 0 || s.MinimumAxisDistance > 255 || s.BlockedAttributeMask <= 0 || s.BlockedAttributeMask > 65535 || len(s.TurnDeltaByLayer) != 4 {
		return errors.New("npc_motion masks, distance or typed layer domain are invalid")
	}
	for _, delta := range s.TurnDeltaByLayer {
		if delta != -1 && delta != 1 {
			return errors.New("npc_motion turn delta must be one cardinal direction")
		}
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	if s.Evidence.Level != "D3" {
		return errors.New("npc_motion requires native D3 evidence")
	}
	if err := validateEvidence(s.TurnLayerEvidence); err != nil {
		return err
	}
	if s.TurnLayerEvidence.Level != "D3" {
		return errors.New("npc_motion typed layer reduction requires native D3 evidence")
	}
	return nil
}
