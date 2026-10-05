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

func TestFieldEquipmentOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.FieldEquipment
	read := func(n, h string) []byte {
		t.Helper()
		b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", n))
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != h {
			t.Fatal("original identity", n, e)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	item := read("ITEM.DAT", "7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d")
	ids := []string{s.RowTextID, s.FooterTextID, s.NoneTextID, s.PreviewTextID}
	for i, v := range s.Parts {
		if v.Part != []int{0, 1, 3, 2}[i] {
			t.Fatal("native category order")
		}
		ids = append(ids, v.HeaderTextID)
	}
	for _, id := range ids {
		d, _ := p.TextDefinition(id)
		c, _ := p.TextGlyphCodes(id)
		if !reflect.DeepEqual(c, tx.Record(*d.Source.Record)) {
			t.Fatal("original text", id)
		}
	}
	for _, v := range []struct {
		w   RawNewGameWindow
		off int
		rec string
	}{{s.RawWindow, 0x1a21c, s.Parts[0].HeaderTextID}, {s.PreviewWindow, 0x1a23a, s.PreviewTextID}} {
		word := func(n int) int { return int(binary.LittleEndian.Uint16(exe[v.off+n:])) }
		w := v.w
		d, _ := p.TextDefinition(v.rec)
		if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) || *d.Source.Record != word(10) {
			t.Fatal("raw native window")
		}
	}
	if s.FrameRows != 3 || s.RowStep != 16 || s.Name != (GeometryAnchor{X: 376, Y: 46}) || s.Cursor != (GeometryAnchor{X: 360, Y: 46}) || s.Worn != (GeometryAnchor{X: 344, Y: 46}) || s.WornGlyph != 10 || s.CursorGlyph != 11 || s.Attack != (NumberField{408, 16, 5}) || s.Defense != (NumberField{408, 32, 5}) {
		t.Fatal("native window consumer differs")
	}
	for code, el := range s.Eligibility {
		gender := -1
		if binary.LittleEndian.Uint16(item[code*7+4:])&0x100 != 0 {
			gender = 1
		}
		var cs = []int{}
		for c := 0; c < 8; c++ {
			if item[code*7+6]&(1<<c) != 0 {
				cs = append(cs, c)
			}
		}
		if el.RequiredGender != gender || !reflect.DeepEqual(el.Classes, cs) {
			t.Fatal("original eligibility", code)
		}
	}
	t.Log("pack canonical", p.ContentHash())
}
func TestFieldEquipmentRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "scope", "unreviewed", "part", "text", "capacity", "number", "anchor", "gender", "class", "eligibility", "window_missing", "number_missing", "part_missing"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			if name == "missing" {
				p.Interface.FieldEquipment = nil
				if p.validateFieldEquipment() == nil {
					t.Fatal("accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldEquipment)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "scope":
				m["scope"] = "all"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "part":
				m["parts"].([]any)[1].(map[string]any)["part"] = 0
			case "text":
				m["none_text_id"] = "unknown"
			case "capacity":
				m["frame_rows"] = 99
			case "number":
				m["attack"].(map[string]any)["digits"] = 99
			case "anchor":
				m["worn"].(map[string]any)["x"] = -1
			case "gender":
				m["eligibility"].([]any)[0].(map[string]any)["required_gender"] = 3
			case "class":
				m["eligibility"].([]any)[0].(map[string]any)["classes"] = []int{99}
			case "eligibility":
				m["eligibility"] = []any{}
			case "window_missing":
				delete(m["raw_window"].(map[string]any), "flags")
			case "number_missing":
				delete(m["attack"].(map[string]any), "x")
			case "part_missing":
				delete(m["parts"].([]any)[0].(map[string]any), "part")
			}
			b, _ = json.Marshal(m)
			var s FieldEquipment
			e = json.Unmarshal(b, &s)
			if e == nil {
				p.Interface.FieldEquipment = &s
				e = p.validateFieldEquipment()
			}
			if e == nil {
				t.Fatal("broken accepted", name)
			}
		})
	}
}
