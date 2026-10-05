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

func TestFieldStatusMenuOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Interface.FieldStatusMenu
	read := func(name, hash string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("original status identity differs", name, e)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	raw := exe[0x19ff4 : 0x19ff4+42]
	word := func(o int) int { return int(binary.LittleEndian.Uint16(raw[o:])) }
	w := s.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) || len(s.Entries) != word(20) || word(22) != 1 {
		t.Fatal("native status header differs")
	}
	for i, e := range s.Entries {
		o := 24 + i*6
		if e.X != word(o)*8 || e.Y != word(o+2) || e.CallbackRaw != word(o+4) {
			t.Fatal("native status row differs", i)
		}
	}
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	d, _ := p.TextDefinition(s.TextID)
	codes, _ := p.TextGlyphCodes(s.TextID)
	if *d.Source.Record != word(10) || !reflect.DeepEqual(codes, tx.Record(word(10))) {
		t.Fatal("native status record differs")
	}
	if fmt.Sprintf("%x", exe[0x1830b-0xec90:0x18313-0xec90]) != "8d36b43ee8d171c3" {
		t.Fatal("status caller/window bytes differ")
	}
}

func TestFieldStatusMenuRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "empty_entries", "missing_zero", "unknown_field", "unknown_text", "unknown_role", "duplicate_role", "bad_window", "bad_rows", "bad_font", "bad_navigation", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(p.Interface.FieldStatusMenu)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "missing":
				p.Interface.FieldStatusMenu = nil
			case "empty_entries":
				m["entries"] = []any{}
			case "missing_zero":
				delete(m["entries"].([]any)[0].(map[string]any), "callback_raw")
			case "unknown_field":
				m["guess"] = true
			case "unknown_text":
				m["text_id"] = "unknown"
			case "unknown_role":
				m["entries"].([]any)[0].(map[string]any)["role"] = "unknown"
			case "duplicate_role":
				m["entries"].([]any)[1].(map[string]any)["role"] = "detail"
			case "bad_window":
				m["raw_window"].(map[string]any)["height"] = 64
			case "bad_rows":
				m["entries"].([]any)[1].(map[string]any)["y"] = 0
			case "bad_font":
				m["font_index"] = 99
			case "bad_navigation":
				m["navigation"] = "guess"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			}
			if name != "missing" {
				b, _ = json.Marshal(m)
				var s FieldStatusMenu
				if e = json.Unmarshal(b, &s); e != nil {
					return
				}
				p.Interface.FieldStatusMenu = &s
			}
			if e = p.validateFieldStatusMenu(); e == nil {
				t.Fatal("broken status menu accepted")
			}
		})
	}
}
