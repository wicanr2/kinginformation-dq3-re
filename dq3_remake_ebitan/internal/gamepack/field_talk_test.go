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

func TestFieldTalkOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Interface.FieldTalk
	read := func(name, hash string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("original identity", name, e)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	code := func(a, b int) []byte { return exe[a-0xec90 : b-0xec90] }
	if fmt.Sprintf("%x", code(0x14e7f, 0x14e8c)) != "bf0401e89e01c6067b060090c3" {
		t.Fatal("native no-target caller differs")
	}
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	record := int(binary.LittleEndian.Uint16(code(0x14e7f, 0x14e82)[1:]))
	d, _ := p.TextDefinition(s.EmptyTextID)
	codes, _ := p.TextGlyphCodes(s.EmptyTextID)
	if d.Source.Record == nil || *d.Source.Record != record || !reflect.DeepEqual(codes, tx.Record(record)) {
		t.Fatal("native record differs")
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
		t.Fatal("native message frame differs")
	}
	t.Log("pack canonical", p.ContentHash())
}

func TestFieldTalkRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "text", "unreviewed", "scope", "return", "presentation", "flow", "unbound_control", "overflow", "pagination"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing" {
				p.Interface.FieldTalk = nil
				if p.validateFieldTalk() == nil {
					t.Fatal("missing accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldTalk)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "text":
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
			case "unbound_control":
				d, _ := p.TextDefinition(p.Interface.FieldTalk.EmptyTextID)
				d.GlyphCodes = []int{65530}
			case "overflow":
				d, _ := p.TextDefinition(p.Interface.FieldTalk.EmptyTextID)
				d.GlyphCodes = []int{65534, 65534, 65534, 65534}
			case "pagination":
				d, _ := p.TextDefinition(p.Interface.FieldTalk.EmptyTextID)
				d.GlyphCodes = []int{int(dq3data.TxtPage)}
			}
			b, _ = json.Marshal(m)
			var s FieldTalk
			err = json.Unmarshal(b, &s)
			if err == nil {
				p.Interface.FieldTalk = &s
				err = p.validateFieldTalk()
			}
			if err == nil {
				t.Fatal("broken accepted", name)
			}
		})
	}
}
