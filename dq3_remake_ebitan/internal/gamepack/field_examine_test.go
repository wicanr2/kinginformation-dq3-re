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

func TestFieldExamineOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Interface.FieldExamine
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
	if fmt.Sprintf("%x", code(0x18c93, 0x18cb0)) != "bf08019a64021b11bf09019a64021b119adb0004118d366e3ee85569c3" {
		t.Fatal("native file-byte consumer differs")
	}
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	for i, id := range []string{s.IntroTextID, s.ResultTextID} {
		a := []int{0x18c93, 0x18c9b}[i]
		record := int(binary.LittleEndian.Uint16(code(a, a+3)[1:]))
		d, _ := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		if d.Source.Record == nil || *d.Source.Record != record || !reflect.DeepEqual(codes, tx.Record(record)) {
			t.Fatal("native record differs", i)
		}
		if i == 0 && *s.ActorVariableCode != int(codes[0]) {
			t.Fatal("actor binding differs")
		}
	}
	raw := exe[0x19fae : 0x19fae+24]
	word := func(n int) int { return int(binary.LittleEndian.Uint16(raw[n:])) }
	w := s.Presentation.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) {
		t.Fatal("native message window differs")
	}
	d, _ := p.TextDefinition(s.Presentation.FrameTextID)
	frame, _ := p.TextGlyphCodes(s.Presentation.FrameTextID)
	if d.Source.Record == nil || *d.Source.Record != word(10) || !reflect.DeepEqual(frame, tx.Record(word(10))) {
		t.Fatal("native message frame differs")
	}
	t.Log("pack canonical", p.ContentHash())
}

func TestFieldExamineRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "actor_missing", "actor_null", "actor_invalid", "intro", "result", "unreviewed", "scope", "return", "join", "presentation", "flow", "unbound_control", "combined_overflow", "pagination"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing" {
				p.Interface.FieldExamine = nil
				if p.validateFieldExamine() == nil {
					t.Fatal("missing accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldExamine)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "actor_missing":
				delete(m, "actor_variable_code")
			case "actor_null":
				m["actor_variable_code"] = nil
			case "actor_invalid":
				m["actor_variable_code"] = 0
			case "intro":
				m["intro_text_id"] = "unknown"
			case "result":
				m["result_text_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "scope":
				m["scope"] = "all_members"
			case "return":
				m["return_mode"] = "automatic"
			case "join":
				m["record_join"] = "inline"
			case "presentation":
				m["presentation"] = nil
			case "flow":
				m["text_flow"].(map[string]any)["scroll_steps"] = 0
			case "unbound_control":
				d, _ := p.TextDefinition(p.Interface.FieldExamine.ResultTextID)
				d.GlyphCodes = []int{65530}
			case "combined_overflow":
				d, _ := p.TextDefinition(p.Interface.FieldExamine.ResultTextID)
				d.GlyphCodes = []int{65534, 65534, 65534, 65534}
			case "pagination":
				d, _ := p.TextDefinition(p.Interface.FieldExamine.ResultTextID)
				d.GlyphCodes = []int{int(dq3data.TxtPage)}
			}
			b, _ = json.Marshal(m)
			var s FieldExamine
			err = json.Unmarshal(b, &s)
			if err == nil {
				p.Interface.FieldExamine = &s
				err = p.validateFieldExamine()
			}
			if err == nil {
				t.Fatal("broken accepted", name)
			}
		})
	}
}
