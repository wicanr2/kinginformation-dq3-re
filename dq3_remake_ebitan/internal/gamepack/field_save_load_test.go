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

func TestFieldSaveLoadOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.FieldSaveLoad
	read := func(name, hash string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal(e, "original identity differs", name)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	for i, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		want := []int{251, 253, 250, 252, 465, 466, 467, 468, 534, 535}[i]
		if *d.Source.Record != want || !reflect.DeepEqual(codes, tx.Record(want)) {
			t.Fatal("save/load record differs", id)
		}
	}
	var w [16]int
	for i := range w {
		w[i] = int(binary.LittleEndian.Uint16(exe[0x1a0fa+i*2:]))
	}
	if s.RawWindow.Flags != w[0]>>8 || s.RawWindow.X != w[1] || s.RawWindow.Y != w[2] || s.RawWindow.Width != w[3] || s.RawWindow.Height != w[4] || s.SlotCount != w[10] || s.ExtraRows != w[11] || s.Cursor.X != w[12]*8 || s.Name.Y != w[13] || s.Number.X != (w[12]-4)*8 || s.Name.X != (w[12]+8)*8 || s.Level.X != (w[12]+16)*8 || s.Gender.X != (w[12]+8+24)*8 || s.NameCapacity != 4 || s.Name.StepX != 16 || s.Name.StepY != 16 || s.Number.Digits != 5 || s.Level.Digits != 5 {
		t.Fatal("raw save/load window or consumers differ")
	}
	for _, x := range []struct {
		offset int
		raw    string
	}{{0x2734, "bffd00"}, {0x2746, "bffa00"}, {0x2751, "bffc00"}, {0x27af, "bd1200"}, {0x283d, "8d16294f"}, {0x2843, "cd21"}, {0x2995, "c7061d250000"}, {0x297b, "c7061d257800"}} {
		if fmt.Sprintf("%x", exe[x.offset:x.offset+len(x.raw)/2]) != x.raw {
			t.Fatalf("caller file%x differs: %x", x.offset, exe[x.offset:x.offset+len(x.raw)/2])
		}
	}
	if clock, ok := s.LoadClock(0); !ok || clock != 0 {
		t.Fatal("upper Load clock differs")
	}
	if clock, ok := s.LoadClock(1); !ok || clock != 120 {
		t.Fatal("lower static Load clock differs")
	}
	if s.Sound.CueRaw != 18 || !s.Sound.WaitForCompletion || s.PrimaryClassRaw != 0 || s.ExperienceMaxLevel != 44 || !reflect.DeepEqual(s.NameControlCodes, []uint16{dq3data.TxtVarEnt, dq3data.TxtVar0}) || s.ExperienceControlCode != dq3data.TxtVarItem {
		t.Fatal("reviewed interpolation/audio differs")
	}
}

func TestFieldSaveLoadRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "missing_zero", "unknown_field", "unknown_text", "bad_height", "bad_cursor", "bad_frame", "bad_gender", "bad_slot_count", "bad_sound", "missing_wait", "unreviewed", "null_geometry"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(p.Interface.FieldSaveLoad)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "missing":
				p.Interface.FieldSaveLoad = nil
			case "missing_zero":
				delete(m, "primary_class_raw")
			case "unknown_field":
				m["guess"] = true
			case "unknown_text":
				m["question_text_id"] = "unknown"
			case "bad_height":
				m["raw_window"].(map[string]any)["height"] = 32
			case "bad_cursor":
				m["cursor"].(map[string]any)["x"] = 900
			case "bad_frame":
				m["header_text_id"] = m["question_text_id"]
			case "bad_gender":
				m["gender_text_ids"] = []string{"unknown"}
			case "bad_slot_count":
				m["slot_count"] = 0
			case "bad_sound":
				m["sound"].(map[string]any)["cue_raw"] = 999
			case "missing_wait":
				delete(m["sound"].(map[string]any), "wait_for_completion")
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "null_geometry":
				m["name"] = nil
			}
			if name != "missing" {
				b, _ = json.Marshal(m)
				var s FieldSaveLoad
				if e = json.Unmarshal(b, &s); e != nil {
					return
				}
				p.Interface.FieldSaveLoad = &s
			}
			if e = p.validateFieldSaveLoad(); e == nil {
				t.Fatal("broken save/load accepted")
			}
		})
	}
}
