package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestRegistrationRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "null_binding", "unknown", "missing_zero_cty", "missing_cursor_x", "null_window_width", "unknown_hit", "duplicate_class", "class_outside", "unknown_text", "unknown_geometry", "unreviewed", "empty_name_fallback", "no_review", "no_retention", "zero_capacity", "cursor_outside"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(p.Interface.Registration)
			var m map[string]any
			json.Unmarshal(raw, &m)
			switch name {
			case "missing":
				p.Interface.Registration = nil
			case "null_binding":
				m["binding"] = nil
			case "unknown":
				m["guess"] = true
			case "missing_zero_cty":
				delete(m["binding"].(map[string]any), "cty_raw")
			case "missing_cursor_x":
				delete(m["class_menu"].(map[string]any)["cursor"].(map[string]any), "x")
			case "null_window_width":
				m["class_menu"].(map[string]any)["raw_window"].(map[string]any)["width"] = nil
			case "unknown_hit":
				m["class_menu"].(map[string]any)["hit_rect"].(map[string]any)["guess"] = true
			case "duplicate_class":
				cs := m["class_options"].([]any)
				cs[1] = cs[0]
			case "class_outside":
				m["class_options"].([]any)[0].(map[string]any)["class_raw"] = 999
			case "unknown_text":
				m["text_roles"].(map[string]any)["greeting"] = "unknown"
			case "unknown_geometry":
				m["geometry_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "empty_name_fallback":
				m["require_nonempty_name"] = false
			case "no_review":
				m["review_before_confirmation"] = false
			case "no_retention":
				m["retain_text_between_records"] = false
			case "zero_capacity":
				m["roster_capacity"] = 0
			case "cursor_outside":
				m["class_menu"].(map[string]any)["cursor"].(map[string]any)["y"] = 340
			}
			if name != "missing" {
				raw, _ = json.Marshal(m)
				var s Registration
				if e = json.Unmarshal(raw, &s); e != nil {
					return
				}
				p.Interface.Registration = &s
			}
			if e = p.validateRegistration(); e == nil {
				t.Fatal("accepted invalid registration")
			}
		})
	}
}

func TestRegistrationOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.Registration
	dir := filepath.Join("..", "..", "..", "assets_raw")
	exe, e := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if e != nil {
		t.Fatal(e)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original EXE identity")
	}
	txt, e := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(txt)) != "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b" {
		t.Fatal("original TXT identity")
	}
	tx := dq3data.LoadText(nil, txt)
	for _, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		if !reflect.DeepEqual(codes, tx.Record(*d.Source.Record)) {
			t.Fatalf("text %s differs", id)
		}
	}
	roles := s.TextRoles
	for id, rec := range map[string]int{roles.Greeting: 550, roles.InitialDecline: 551, roles.NamePrompt: 554, roles.CreationCancelled: 558, roles.Registered: 559, roles.Farewell: 560, s.ClassMenu.WindowTextID: 555} {
		d, _ := p.TextDefinition(id)
		if *d.Source.Record != rec {
			t.Fatal("role record mismatch")
		}
	}
	var classes []byte
	for _, c := range s.ClassOptions {
		classes = append(classes, byte(c.ClassRaw))
		d, _ := p.TextDefinition(c.TextID)
		if *d.Source.Record != 457+c.ClassRaw {
			t.Fatal("class record mismatch")
		}
	}
	if !reflect.DeepEqual(classes, exe[0x1b095:0x1b09b]) {
		t.Fatal("class mapping differs from native one-based table")
	}
	w := s.ClassMenu.RawWindow
	linear, e := strconv.ParseInt(strings.TrimPrefix(w.Address, "linear:"), 0, 32)
	if e != nil {
		t.Fatal(e)
	}
	off := int(linear) - 0xec90 // IDA9.4 linear -> file。只作原始格式 oracle。
	if int(exe[off+1]) != w.Flags || int(binary.LittleEndian.Uint16(exe[off+2:])) != w.X || int(binary.LittleEndian.Uint16(exe[off+4:])) != w.Y || int(binary.LittleEndian.Uint16(exe[off+6:])) != w.Width || int(binary.LittleEndian.Uint16(exe[off+8:])) != w.Height || int(binary.LittleEndian.Uint16(exe[off+10:])) != 555 {
		t.Fatal("class window differs")
	}
	a := s.ClassMenu.Cursor
	h := s.ClassMenu.HitRect
	if a.X != int(binary.LittleEndian.Uint16(exe[off+24:]))*8 || a.Y != int(binary.LittleEndian.Uint16(exe[off+26:])) || a.StepY != 16 || h.X != a.X || h.Y != a.Y || h.Width != (w.Width-4)*8 || h.Height != 16 {
		t.Fatal("cursor projection differs")
	}
	// caller 將初始 count5 覆寫為6，不能靜態截掉第六職業。
	if binary.LittleEndian.Uint16(exe[off+20:]) != 5 || len(s.ClassOptions) != 6 || s.RosterCapacity != 11 {
		t.Fatal("runtime class count / nonhero slot capacity")
	}
	cty, e := os.ReadFile(filepath.Join(dir, fmt.Sprintf("CTY%02d.DAT", s.Binding.CTYRaw)))
	if e != nil {
		t.Fatal(e)
	}
	town, e := dq3data.OpenTown(cty, s.Binding.Section, false)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, n := range town.NPCs {
		if (n.Ctrl>>3)&7 == 2 && n.B4 == s.Binding.NPCHandlerRaw {
			found = true
		}
	}
	if !found {
		t.Fatal("registration binding has no original NPC")
	}
}
