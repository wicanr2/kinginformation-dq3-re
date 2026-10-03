package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestRegionDialogueRewardRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing_collection", "missing_gold", "null_tile", "unknown_field", "bad_text", "bad_presentation", "bad_scene", "bad_item", "bad_flags", "bad_shadow", "unreviewed", "duplicate"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing_collection" {
				p.Events.RegionDialogueRewardEvents = nil
			} else {
				b, _ := json.Marshal(p.Events.RegionDialogueRewardEvents[0])
				var m map[string]any
				json.Unmarshal(b, &m)
				switch name {
				case "missing_gold":
					delete(m, "gold")
				case "null_tile":
					m["tile"] = nil
				case "unknown_field":
					m["guess"] = true
				case "bad_text":
					m["text_id"] = "unknown"
				case "bad_presentation":
					m["presentation_id"] = "unknown"
				case "bad_scene":
					m["section"] = 99
				case "bad_item":
					m["item_raw_ids"] = []int{255}
				case "bad_flags":
					m["set_flag_raw"] = m["clear_flag_raw"]
				case "bad_shadow":
					m["shadow"].(map[string]any)["offset_x"] = 640
				case "unreviewed":
					m["evidence"].(map[string]any)["level"] = "D2"
				}
				b, _ = json.Marshal(m)
				var e RegionDialogueRewardEvent
				if json.Unmarshal(b, &e) != nil {
					return
				}
				p.Events.RegionDialogueRewardEvents = []RegionDialogueRewardEvent{e}
				if name == "duplicate" {
					p.Events.RegionDialogueRewardEvents = append(p.Events.RegionDialogueRewardEvents, e)
				}
			}
			if p.validateRegionDialogueRewards() == nil && p.validateRegionDialogueRewardRefs() == nil {
				t.Fatal("accepted broken region dialogue reward")
			}
		})
	}
}

func TestRegionDialogueRewardMatchesOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Events.RegionDialogueRewardEvents) != 1 {
		t.Fatal("unexpected event count")
	}
	e := p.Events.RegionDialogueRewardEvents[0]
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	read := func(name, hash string, size int) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if len(b) != size || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("original identity differs", name)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c", 115282)
	u16 := func(linear int) int { return int(binary.LittleEndian.Uint16(exe[linear-0xec90:])) }
	if e.RequiredFlagRaw != u16(0x1024d) || e.ClearFlagRaw != u16(0x102b2) || e.SetFlagRaw != u16(0x102b8) || e.Gold != u16(0x102a9) {
		t.Fatal("native flag/gold operands differ")
	}
	var items []int
	for _, linear := range []int{0x10276, 0x1027f, 0x10288, 0x10291, 0x1029a, 0x102a3} {
		items = append(items, u16(linear))
	}
	if !reflect.DeepEqual(items, e.ItemRawIDs) {
		t.Fatal("native ordered reward operands differ")
	}
	// Calls immediately follow each item writer; no conditional branch checks AX.
	for _, linear := range []int{0x10278, 0x10281, 0x1028a, 0x10293, 0x1029c, 0x102a5} {
		if exe[linear-0xec90] != 0xe8 {
			t.Fatal("grant caller bytes differ")
		}
	}
	cty := read("CTY25.DAT", "11d5c60377c6a98bbb9cfc9532652c5e23f4e5c939e769397e8fa5b2f22f9e2b", 3756)
	town, err := dq3data.OpenTown(cty, e.Section, false)
	if err != nil {
		t.Fatal(err)
	}
	if e.CTYRaw != 25 || e.Tile.X >= town.W || e.Tile.Y >= town.H {
		t.Fatal("native scene binding differs")
	}
	subid := int(town.HiMap[e.Tile.Y*town.W+e.Tile.X] & 31)
	if subid != 1 || len(town.SpecialHandlers) < subid || town.SpecialHandlers[subid-1] != 56 {
		t.Fatal("native step handler binding differs")
	}
	d, ok := p.TextDefinition(e.TextID)
	if !ok || d.Source.File != "D3TXT01.TXT" || d.Source.Record == nil || *d.Source.Record != u16(0x10261)-3000 {
		t.Fatal("native reader record differs")
	}
	raw, err := os.ReadFile(filepath.Join(dir, d.Source.File))
	if err != nil {
		t.Fatal(err)
	}
	native := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
	if len(native) != len(d.GlyphCodes) {
		t.Fatal("native text length differs")
	}
	waits := 0
	for i, c := range native {
		if int(c) != d.GlyphCodes[i] {
			t.Fatal("native text word differs", i)
		}
		if c == dq3data.TxtPage {
			waits++
		}
	}
	if waits != 9 || e.ProgressFlagRaw != msCompatibilityStartForTest {
		t.Fatal("wait count or remake save compatibility differs")
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}

const msCompatibilityStartForTest = 0x200 // Existing remake save ID, not a native story flag.
