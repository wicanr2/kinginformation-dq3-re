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
	"testing"
)

func TestRecruitmentEntryOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.RecruitmentEntry
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
	for i, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		if *d.Source.Record != 527+i || !reflect.DeepEqual(codes, tx.Record(*d.Source.Record)) {
			t.Fatalf("text %s differs", id)
		}
	}
	w := s.Menu.RawWindow
	word := func(i int) int { return int(binary.LittleEndian.Uint16(exe[0x19f72+i*2:])) }
	if w.Flags != word(0)>>8 || w.X != word(1) || w.Y != word(2) || w.Width != word(3) || w.Height != word(4) || len(s.OptionActions) != word(10) || s.Menu.Cursor.X != word(12)*8 || s.Menu.Cursor.Y != word(13) || s.Menu.Cursor.StepY != 16 {
		t.Fatal("original menu geometry differs")
	}
	if !reflect.DeepEqual(s.OptionActions, []string{RecruitJoin, RecruitLeave, RecruitView}) || s.Binding != (RegistrationBinding{CTYRaw: 0, Section: 0, NPCHandlerRaw: 1}) {
		t.Fatal("original menu roles or NPC handler differ")
	}
	for _, a := range []struct {
		off int
		raw string
	}{{0x16dd, "e8924c"}, {0x16e0, "bf0f02"}, {0x16e8, "bf1002"}, {0x16f0, "8d36323e"}, {0x16f4, "e85cf1"}} {
		if fmt.Sprintf("%x", exe[a.off:a.off+len(a.raw)/2]) != a.raw {
			t.Fatalf("caller bytes file%x", a.off)
		}
	}
}

func TestRecruitmentEntryRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "null_binding", "missing_zero_section", "unknown", "unknown_menu_field", "null_cursor", "outside_window", "unknown_text", "unknown_presentation", "unreviewed", "unknown_action", "duplicate_action"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(p.Interface.RecruitmentEntry)
			var m map[string]any
			json.Unmarshal(raw, &m)
			switch name {
			case "missing":
				p.Interface.RecruitmentEntry = nil
			case "null_binding":
				m["binding"] = nil
			case "missing_zero_section":
				delete(m["binding"].(map[string]any), "section")
			case "unknown":
				m["guess"] = true
			case "unknown_menu_field":
				m["menu"].(map[string]any)["guess"] = true
			case "null_cursor":
				m["menu"].(map[string]any)["cursor"] = nil
			case "outside_window":
				m["menu"].(map[string]any)["cursor"].(map[string]any)["y"] = 340
			case "unknown_text":
				m["greeting_text_ids"].([]any)[0] = "unknown"
			case "unknown_presentation":
				m["presentation_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "unknown_action":
				m["option_actions"].([]any)[0] = "guess"
			case "duplicate_action":
				m["option_actions"].([]any)[1] = m["option_actions"].([]any)[0]
			}
			if name != "missing" {
				raw, _ = json.Marshal(m)
				var s RecruitmentEntry
				if e = json.Unmarshal(raw, &s); e != nil {
					return
				}
				p.Interface.RecruitmentEntry = &s
			}
			if e = p.validateRecruitmentEntry(); e == nil {
				t.Fatal("accepted broken recruitment entry")
			}
		})
	}
}
