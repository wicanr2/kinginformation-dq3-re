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

func TestFieldCommandOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Interface.FieldCommandMenu
	read := func(name, hash string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("original identity differs", name, err)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	raw := exe[0x19eac : 0x19eac+60]
	word := func(offset int) int { return int(binary.LittleEndian.Uint16(raw[offset:])) }
	w := s.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) || len(s.Entries) != word(20) || word(22) != 2 {
		t.Fatal("native menu header differs")
	}
	for i, e := range s.Entries {
		o := 24 + i*6
		if e.X != word(o)*8 || e.Y != word(o+2) || e.CallbackRaw != word(o+4) {
			t.Fatal("native selector entry differs", i)
		}
	}
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	d, _ := p.TextDefinition(s.TextID)
	codes, _ := p.TextGlyphCodes(s.TextID)
	if *d.Source.Record != word(10) || !reflect.DeepEqual(codes, tx.Record(word(10))) {
		t.Fatal("command frame record differs")
	}
	labels := p.Interface.FieldCommandLabels.Entries()
	// Consumer 1F908 uses six-byte entries; visible labels occupy the two
	// cells after each selector, in record400's original two-column rows.
	for _, e := range s.Entries {
		row := (e.Y - w.Y) / dq3data.GlyphPx
		col := (e.X-w.X*8)/dq3data.GlyphPx + 1
		start := row*(d.Layout.Columns+1) + col
		l := labels[e.Command]
		if int(codes[start]) != l.PrimaryGlyph || int(codes[start+2]) != l.SecondaryGlyph {
			t.Fatal("selector role does not match visible native label", e.Command)
		}
	}
}

func TestFieldCommandRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "missing_zero", "unknown_field", "unknown_text", "unknown_command", "duplicate_role", "bad_window", "bad_rows", "bad_columns", "bad_font", "bad_navigation", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(p.Interface.FieldCommandMenu)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "missing":
				p.Interface.FieldCommandMenu = nil
			case "missing_zero":
				delete(m["entries"].([]any)[0].(map[string]any), "callback_raw")
			case "unknown_field":
				m["guess"] = true
			case "unknown_text":
				m["text_id"] = "unknown"
			case "unknown_command":
				m["entries"].([]any)[0].(map[string]any)["command"] = "unknown"
			case "duplicate_role":
				m["entries"].([]any)[1].(map[string]any)["command"] = "talk"
			case "bad_window":
				m["raw_window"].(map[string]any)["width"] = 20
			case "bad_rows":
				m["entries"].([]any)[1].(map[string]any)["y"] = 90
			case "bad_columns":
				m["entries"].([]any)[3].(map[string]any)["x"] = 150
			case "bad_font":
				m["font_index"] = 99
			case "bad_navigation":
				m["navigation"] = "row_wrap"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			}
			if name != "missing" {
				b, _ = json.Marshal(m)
				var s FieldCommandMenu
				if err = json.Unmarshal(b, &s); err != nil {
					return
				}
				p.Interface.FieldCommandMenu = &s
			}
			if err = p.validateFieldCommandMenu(); err == nil {
				t.Fatal("broken menu accepted")
			}
		})
	}
}
