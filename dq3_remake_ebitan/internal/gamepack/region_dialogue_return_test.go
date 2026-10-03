package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"os"
	"path/filepath"
	"testing"
)

func TestRegionDialogueReturnRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing_collection", "missing_flag", "null_shadow", "unknown_field", "bad_text", "inline_wait", "bad_scene", "bad_presentation", "bad_facing", "bad_direction", "negative_hold", "bad_shadow", "unreviewed", "duplicate"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing_collection" {
				p.Events.RegionDialogueReturnEvents = nil
			} else {
				b, _ := json.Marshal(p.Events.RegionDialogueReturnEvents[0])
				var m map[string]any
				json.Unmarshal(b, &m)
				switch name {
				case "missing_flag":
					delete(m, "required_flag_raw")
				case "null_shadow":
					m["shadow"] = nil
				case "unknown_field":
					m["guess"] = true
				case "bad_text":
					m["text_id"] = "unknown"
				case "inline_wait":
					m["text_id"] = p.Events.RegionDialogueRewardEvents[0].TextID
				case "bad_scene":
					m["section"] = 99
				case "bad_presentation":
					m["presentation_id"] = "unknown"
				case "bad_facing":
					m["actor_facing"] = 4
				case "bad_direction":
					m["return_direction"] = 4
				case "negative_hold":
					m["return_hold_frames"] = -1
				case "bad_shadow":
					m["shadow"].(map[string]any)["offset_x"] = 640
				case "unreviewed":
					m["evidence"].(map[string]any)["level"] = "D2"
				}
				b, _ = json.Marshal(m)
				var e RegionDialogueReturnEvent
				if json.Unmarshal(b, &e) != nil {
					return
				}
				p.Events.RegionDialogueReturnEvents = []RegionDialogueReturnEvent{e}
				if name == "duplicate" {
					p.Events.RegionDialogueReturnEvents = append(p.Events.RegionDialogueReturnEvents, e)
				}
			}
			if p.validateRegionDialogueReturns() == nil && p.validateRegionDialogueReturnRefs() == nil {
				t.Fatal("accepted broken return contract")
			}
		})
	}
}

func TestRegionDialogueReturnMatchesOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Events.RegionDialogueReturnEvents) != 1 {
		t.Fatal("unexpected count")
	}
	e := p.Events.RegionDialogueReturnEvents[0]
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original identity differs")
	}
	u16 := func(ea int) int { return int(binary.LittleEndian.Uint16(exe[ea-0xec90:])) }
	// Raw native 0=down,1=left,2=up,3=right maps to the engine's directions.
	nativeDirections := []int{0, 2, 1, 3}
	if e.RequiredFlagRaw != u16(0x1020c) || e.ActorRecordRaw != u16(0x1021a) || e.ActorFacing != nativeDirections[u16(0x10217)] ||
		e.ReturnDirection != nativeDirections[u16(0x10240)] || exe[0x1021d-0xec90] != 2 || e.TurnHoldFrames != 5 || e.ReturnHoldFrames != 5 {
		t.Fatal("native operands or timer approximation differ")
	}
	if u16(0x24dd0+0x3bb4+e.HandlerRaw*2) != 0x020b {
		t.Fatal("handler pointer differs")
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY00.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	town, err := dq3data.OpenTown(cty, e.Section, false)
	if err != nil {
		t.Fatal(err)
	}
	if e.CTYRaw != 0 || e.Section != 0 || town.SpecialHandlers[int(town.HiMap[18*town.W+21]&31)-1] != e.HandlerRaw {
		t.Fatal("normal native trigger differs")
	}
	d, ok := p.TextDefinition(e.TextID)
	if !ok || d.Source.File != "D3TXT01.TXT" || d.Source.Record == nil || *d.Source.Record != u16(0x1022b)-3000 {
		t.Fatal("native text selector differs")
	}
	raw, err := os.ReadFile(filepath.Join(dir, d.Source.File))
	if err != nil {
		t.Fatal(err)
	}
	native := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
	if len(native) != len(d.GlyphCodes) {
		t.Fatal("text length differs")
	}
	for i, c := range native {
		if int(c) != d.GlyphCodes[i] || c == dq3data.TxtPage {
			t.Fatal("native automatic text differs")
		}
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.ContentVersion(), p.ContentHash())
}
