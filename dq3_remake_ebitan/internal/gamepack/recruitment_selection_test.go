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

func TestRecruitmentSelectionOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.RecruitmentSelection
	dir := filepath.Join("..", "..", "..", "assets_raw")
	exe, e := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if e != nil {
		t.Fatal(e)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original identity")
	}
	word := func(i int) int { return int(binary.LittleEndian.Uint16(exe[0x19f54+i*2:])) }
	w := s.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(1) || w.Y != word(2) || w.Width != word(3) || w.Height != word(4) || s.Cursor.X != word(12)*8 || s.Cursor.Y != word(13) {
		t.Fatal("selection window raw data differs")
	}
	if s.ExtraRows != 3 || s.NameCapacity != 4 || s.Name != (GeometryAnchor{X: 168, Y: 78, StepX: 16, StepY: 16}) || s.Level != (NumberField{X: 216, Y: 78, Digits: 5}) || s.Class.X != 312 || s.Gender.X != 408 {
		t.Fatal("selection consumer geometry differs")
	}
	data, e := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(data)) != "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b" {
		t.Fatal("original text identity")
	}
	tx := dq3data.LoadText(nil, data)
	records := []int{530, word(5), word(7), word(8), 540, 541, 457, 458, 459, 460, 461, 462, 463, 464, 534, 535, 528, 316}
	for i, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		if *d.Source.Record != records[i] || !reflect.DeepEqual(codes, tx.Record(records[i])) {
			t.Fatalf("text %s differs", id)
		}
	}
	for _, raw := range []struct {
		off   int
		bytes string
	}{
		{0x199f, "eb65"}, {0x1a06, "bf3c01"},
		{0x176d, "050300"}, {0x1770, "b104"}, {0x1772, "d3e0"}, {0x1774, "a31c3e"},
		{0x1d55, "c70616071500"}, {0x1d5b, "c70618074e00"}, {0x1d86, "83c506"},
		{0x1da9, "8306160712"}, {0x1dce, "830616071e"}, {0x177e, "803e260701"},
	} {
		if fmt.Sprintf("%x", exe[raw.off:raw.off+len(raw.bytes)/2]) != raw.bytes {
			t.Fatalf("original consumer file%x differs", raw.off)
		}
	}
}

func TestRecruitmentSelectionRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "null_name", "missing_zero_step", "unknown", "unknown_window", "unknown_text", "unreviewed", "outside", "row_step", "header_rows", "unknown_class", "missing_class", "unknown_gender", "frame_control", "missing_empty_view", "null_empty_view", "unknown_empty_view"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			raw, _ := json.Marshal(p.Interface.RecruitmentSelection)
			var m map[string]any
			json.Unmarshal(raw, &m)
			switch name {
			case "missing":
				p.Interface.RecruitmentSelection = nil
			case "null_name":
				m["name"] = nil
			case "missing_zero_step":
				delete(m["cursor"].(map[string]any), "step_x")
			case "unknown":
				m["guess"] = true
			case "unknown_window":
				m["raw_window"].(map[string]any)["guess"] = true
			case "unknown_text":
				m["prompt_text_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "outside":
				m["name"].(map[string]any)["y"] = 340
			case "row_step":
				m["name"].(map[string]any)["step_y"] = 17
			case "header_rows":
				m["extra_rows"] = 4
			case "unknown_class":
				m["class_options"].([]any)[0].(map[string]any)["text_id"] = "unknown"
			case "missing_class":
				m["class_options"] = []any{}
			case "unknown_gender":
				m["gender_text_ids"].([]any)[0] = "unknown"
			case "frame_control":
				m["row_text_id"] = m["prompt_text_id"]
			case "missing_empty_view":
				delete(m, "empty_view_text_id")
			case "null_empty_view":
				m["empty_view_text_id"] = nil
			case "unknown_empty_view":
				m["empty_view_text_id"] = "unknown"
			}
			if name != "missing" {
				raw, _ = json.Marshal(m)
				var s RecruitmentSelection
				if e = json.Unmarshal(raw, &s); e != nil {
					return
				}
				p.Interface.RecruitmentSelection = &s
			}
			if e = p.validateRecruitmentSelection(); e == nil {
				t.Fatal("broken selection accepted")
			}
		})
	}
}
