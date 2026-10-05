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

func TestFieldStatusReorderOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.FieldStatusMenu.Reorder
	read := func(name, hash string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("reorder original identity", name, err)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	code := func(a, b int) []byte { return exe[a-0xec90 : b-0xec90] }
	if fmt.Sprintf("%x", code(0x18685, 0x1869b)) != "803e77500174088d36143fe8506ec3bf0a02e889c9c3" ||
		fmt.Sprintf("%x", code(0x15010, 0x15015)) != "9adb000411" ||
		fmt.Sprintf("%x", code(0x21286, 0x21298)) != "81ffb80b7d0cd1e71ea12e258ed88b351fc3" {
		t.Fatal("reorder consumer differs")
	}
	record := int(binary.LittleEndian.Uint16(code(0x18694, 0x18697)[1:]))
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	d, _ := p.TextDefinition(s.SingleMemberTextID)
	codes, _ := p.TextGlyphCodes(s.SingleMemberTextID)
	if d.Source.Record == nil || *d.Source.Record != record || !reflect.DeepEqual(codes, tx.Record(record)) {
		t.Fatal("reorder text differs")
	}
	raw := exe[0x19fae : 0x19fae+24]
	word := func(n int) int { return int(binary.LittleEndian.Uint16(raw[n:])) }
	w := s.Presentation.RawWindow
	if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) {
		t.Fatal("reorder original window differs")
	}
	d, _ = p.TextDefinition(s.Presentation.FrameTextID)
	frame, _ := p.TextGlyphCodes(s.Presentation.FrameTextID)
	if d.Source.Record == nil || *d.Source.Record != word(10) || !reflect.DeepEqual(frame, tx.Record(word(10))) {
		t.Fatal("reorder original frame differs")
	}
	if s.Presentation.GlyphStepX != p.Interface.OpeningPrelude.GlyphStepX || s.TextFlow.ScrollSteps != p.Interface.OpeningPrelude.TextFlow.ScrollSteps || s.TextFlow.ScrollStepPixels != p.Interface.OpeningPrelude.TextFlow.ScrollStepPixels ||
		s.WaitIndicator.X != p.Interface.OpeningPrelude.WaitIndicator.X || s.WaitIndicator.VisibleGlyph != p.Interface.OpeningPrelude.WaitIndicator.VisibleGlyph || s.WaitIndicator.HiddenGlyph != p.Interface.OpeningPrelude.WaitIndicator.HiddenGlyph {
		t.Fatal("shared native consumer differs")
	}
}

func TestFieldStatusReorderRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "missing_flow", "missing_zero", "unreviewed", "unknown_text", "unknown_frame", "nil_presentation", "overflow", "control", "scroll", "indicator", "return", "scope"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			if name == "missing" {
				p.Interface.FieldStatusMenu.Reorder = nil
				if p.validateFieldStatusReorder() == nil {
					t.Fatal("missing accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldStatusMenu.Reorder)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "missing_flow":
				delete(m, "text_flow")
			case "missing_zero":
				delete(m["wait_indicator"].(map[string]any), "hidden_glyph")
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "unknown_text":
				m["single_member_text_id"] = "unknown"
			case "unknown_frame":
				m["presentation"].(map[string]any)["frame_text_id"] = "unknown"
			case "nil_presentation":
				m["presentation"] = nil
			case "overflow":
				d, _ := p.TextDefinition(p.Interface.FieldStatusMenu.Reorder.SingleMemberTextID)
				d.GlyphCodes = make([]int, 100)
			case "control":
				d, _ := p.TextDefinition(p.Interface.FieldStatusMenu.Reorder.SingleMemberTextID)
				d.GlyphCodes = []int{65520}
			case "scroll":
				m["text_flow"].(map[string]any)["scroll_steps"] = 0
			case "indicator":
				m["wait_indicator"].(map[string]any)["x"] = 1000
			case "return":
				m["return_mode"] = "guess"
			case "scope":
				m["scope"] = "all_members"
			}
			b, _ = json.Marshal(m)
			var s FieldStatusReorder
			e = json.Unmarshal(b, &s)
			if e == nil {
				p.Interface.FieldStatusMenu.Reorder = &s
				e = p.validateFieldStatusReorder()
			}
			if e == nil {
				t.Fatal("broken contract accepted", name)
			}
		})
	}
}
