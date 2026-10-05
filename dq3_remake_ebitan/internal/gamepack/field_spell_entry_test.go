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

func TestFieldSpellEntryOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.FieldSpellEntry
	read := func(n, h string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", n))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != h {
			t.Fatal("original identity", n, e)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	code := func(a, b int) []byte { return exe[a-0xec90 : b-0xec90] }
	if fmt.Sprintf("%x", code(0x18869, 0x18870)) != "c70622070100c3" || fmt.Sprintf("%x", code(0x1c9e7, 0x1c9ee)) != "bf0601e83686c3" || fmt.Sprintf("%x", code(0x1cb39, 0x1cb3c)) != "b001c3" {
		t.Fatal("empty spell native branches differ")
	}
	rec := int(binary.LittleEndian.Uint16(code(0x1c9e7, 0x1c9ea)[1:]))
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	d, _ := p.TextDefinition(s.EmptyTextID)
	codes, _ := p.TextGlyphCodes(s.EmptyTextID)
	if d.Source.Record == nil || *d.Source.Record != rec || !reflect.DeepEqual(codes, tx.Record(rec)) || s.ActorVariableCode == nil || *s.ActorVariableCode != int(codes[0]) {
		t.Fatal("original empty-spell text or actor binding differs")
	}
	raw := exe[0x19fae : 0x19fae+24]
	word := func(n int) int { return int(binary.LittleEndian.Uint16(raw[n:])) }
	w := s.Presentation.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) {
		t.Fatal("native message window differs")
	}
	d, _ = p.TextDefinition(s.Presentation.FrameTextID)
	frame, _ := p.TextGlyphCodes(s.Presentation.FrameTextID)
	if d.Source.Record == nil || *d.Source.Record != word(10) || !reflect.DeepEqual(frame, tx.Record(word(10))) {
		t.Fatal("original message frame differs")
	}
	t.Log("pack canonical", p.ContentHash())
}

func TestFieldSpellEntryRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "missing_actor", "actor_null", "actor_invalid", "unknown_text", "unreviewed", "scope", "return", "presentation", "flow", "overflow", "unbound_control"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			if name == "missing" {
				p.Interface.FieldSpellEntry = nil
				if p.validateFieldSpellEntry() == nil {
					t.Fatal("missing accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldSpellEntry)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "missing_actor":
				delete(m, "actor_variable_code")
			case "actor_null":
				m["actor_variable_code"] = nil
			case "actor_invalid":
				m["actor_variable_code"] = 0
			case "unknown_text":
				m["empty_text_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "scope":
				m["scope"] = "all_members"
			case "return":
				m["return_mode"] = "automatic"
			case "presentation":
				m["presentation"] = nil
			case "flow":
				m["text_flow"].(map[string]any)["scroll_steps"] = 0
			case "overflow":
				d, _ := p.TextDefinition(p.Interface.FieldSpellEntry.EmptyTextID)
				d.GlyphCodes = make([]int, 100)
			case "unbound_control":
				d, _ := p.TextDefinition(p.Interface.FieldSpellEntry.EmptyTextID)
				d.GlyphCodes = []int{65531, 65530}
			}
			b, _ = json.Marshal(m)
			var s FieldSpellEntry
			e = json.Unmarshal(b, &s)
			if e == nil {
				p.Interface.FieldSpellEntry = &s
				e = p.validateFieldSpellEntry()
			}
			if e == nil {
				t.Fatal("broken accepted", name)
			}
		})
	}
}
