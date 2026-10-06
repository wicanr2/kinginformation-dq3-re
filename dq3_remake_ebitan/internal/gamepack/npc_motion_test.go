package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestNPCMotionRejectIncompleteParameters(t *testing.T) {
	for _, name := range []string{"missing", "anchor", "roll", "null_accept", "accept_range", "directions", "duplicate_direction", "direction_mask", "mask_overlap", "missing_layers", "layer_delta", "terrain_mask", "distance", "unreviewed", "layer_unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			s := p.Characters.NPCMotion
			switch name {
			case "missing":
				p.Characters.NPCMotion = nil
			case "anchor":
				s.ViewportAnchor = nil
			case "roll":
				s.Step.Bound = 0
			case "null_accept":
				s.Evaluation.Accepted = nil
			case "accept_range":
				bad := s.Turn.Bound
				s.Turn.Accepted = &bad
			case "directions":
				s.Directions = nil
			case "duplicate_direction":
				s.Directions[1] = s.Directions[0]
			case "direction_mask":
				s.DirectionMask = 0
			case "mask_overlap":
				s.FrozenMask = s.MoveMask
			case "missing_layers":
				s.TurnDeltaByLayer = nil
			case "layer_delta":
				s.TurnDeltaByLayer[2] = 0
			case "terrain_mask":
				s.BlockedAttributeMask = 0
			case "distance":
				s.MinimumAxisDistance = 0
			case "unreviewed":
				s.Evidence.Level = "D2"
			case "layer_unreviewed":
				s.TurnLayerEvidence.Level = "D2"
			}
			if p.validateNPCMotion() == nil {
				t.Fatal("accepted incomplete motion parameters")
			}
		})
	}
}

func TestNPCMotionMatchesOriginalOperandsAndDirectionTable(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original identity", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Characters.NPCMotion
	u8 := func(pc, operand int) int { return int(exe[pc-0xec90+operand]) }
	u16 := func(pc, operand int) int { return int(binary.LittleEndian.Uint16(exe[pc-0xec90+operand:])) }
	for _, r := range []struct{ got, want int }{
		{s.ViewportAnchor.X, u8(0x11f61, 2)}, {s.ViewportAnchor.Y, u16(0x11f51, 1)},
		{s.ViewportColumns, u16(0x11f7e, 1)}, {s.ViewportRows, u16(0x11f74, 1)},
		{s.Evaluation.Bound, u16(0x1203d, 1)}, {*s.Evaluation.Accepted, u8(0x12043, 2)},
		{s.DirectionBound, u16(0x1205f, 1)}, {s.DirectionMask, u8(0x12068, 1)},
		{s.Turn.Bound, u16(0x1206e, 1)}, {*s.Turn.Accepted, u8(0x12074, 2)},
		{s.Step.Bound, u16(0x120aa, 1)}, {*s.Step.Accepted, u8(0x120b0, 2)},
		{s.MoveMask, u8(0x1204e, 3)}, {s.FrozenMask, u8(0x12048, 3)},
		{s.MinimumAxisDistance, u8(0x120e7, 2)}, {s.MinimumAxisDistance, u8(0x12117, 2)},
	} {
		if r.got != r.want {
			t.Fatal("original operand differs", r)
		}
	}
	for i, d := range s.Directions {
		off := 0x16140 + 0xb35 + i*4
		if d.X != int(int16(binary.LittleEndian.Uint16(exe[off:]))) || d.Y != int(int16(binary.LittleEndian.Uint16(exe[off+2:]))) {
			t.Fatal("original direction differs", i)
		}
	}
	if s.QuotientMask != 255 || s.BlockedAttributeMask != 255 || exe[0x1218d-0xec90] != 0x3c || exe[0x1218d-0xec90+1] != 0 || !reflect.DeepEqual(s.TurnDeltaByLayer, []int{-1, -1, 1, 1}) {
		t.Fatal("AL width or reviewed signed-word reduction differs")
	}
	t.Logf("schema%s content%s hash%s", p.Schema(), p.ContentVersion(), p.ContentHash())
}
