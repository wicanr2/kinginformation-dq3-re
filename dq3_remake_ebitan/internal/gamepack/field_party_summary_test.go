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

func TestFieldPartySummaryOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.FieldStatusMenu.Summary
	read := func(name, hash string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("summary original identity", name, err)
		}
		return b
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	tx := dq3data.LoadText(nil, read("D3TXT00.TXT", "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"))
	for _, v := range []struct {
		offset        int
		window        RawNewGameWindow
		ids           []string
		recordOffsets []int
	}{
		{0x19fc4, s.MoneyWindow, []string{s.MoneyTextID}, []int{10}},
		{0x1a03c, s.RawWindow, []string{s.LeftTextID, s.ColumnTextID, s.RightTextID}, []int{10, 14, 16}},
	} {
		raw := exe[v.offset : v.offset+24]
		word := func(o int) int { return int(binary.LittleEndian.Uint16(raw[o:])) }
		w := v.window
		if w.Flags != word(0)>>8 || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) {
			t.Fatal("summary raw window differs")
		}
		for i, id := range v.ids {
			d, _ := p.TextDefinition(id)
			codes, _ := p.TextGlyphCodes(id)
			record := word(v.recordOffsets[i])
			if d.Source.Record == nil || *d.Source.Record != record || !reflect.DeepEqual(codes, tx.Record(record)) {
				t.Fatal("summary raw text differs", id)
			}
		}
	}
	code := func(a, b int) []byte { return exe[a-0xec90 : b-0xec90] }
	if fmt.Sprintf("%x", code(0x17cdf, 0x17cec)) != "b30af6e3050400a3a23ea3023f" || fmt.Sprintf("%x", code(0x1fc57, 0x1fc5f)) != "8a4401a8087401c3" || fmt.Sprintf("%x", code(0x1867f, 0x18685)) != "9adb000411c3" {
		t.Fatal("summary width, shadow or fresh-key consumer differs")
	}
	if s.WidthBase != int(code(0x17ce3, 0x17ce6)[1]) || s.ColumnStep != int(binary.LittleEndian.Uint16(exe[0x1a052:]))*8 || s.MaxColumns != (s.RawWindow.Width-s.WidthBase)/(s.ColumnStep/8) {
		t.Fatal("summary width or columns differ")
	}
	w := s.RawWindow
	if s.Name.X != (w.X+2)*8 || s.Name.Y != w.Y+16 || s.NameLimit != 4 || s.MoneyValue != (NumberField{X: (s.MoneyWindow.X + 4) * 8, Y: s.MoneyWindow.Y + 16, Digits: 8}) {
		t.Fatal("summary name or money consumer differs")
	}
	for _, v := range []struct {
		field NumberField
		dy    int
	}{{s.HP, 32}, {s.MaxHP, 64}, {s.MP, 80}, {s.MaxMP, 112}} {
		if v.field != (NumberField{X: w.X * 8, Y: w.Y + v.dy, Digits: 5}) {
			t.Fatal("summary actor numerical consumer differs")
		}
	}
}

func TestFieldPartySummaryRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "unknown", "missing_zero", "missing_window", "unknown_text", "unreviewed", "width", "step", "capacity", "name", "name_overflow", "money", "hp", "return"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			if name == "missing" {
				p.Interface.FieldStatusMenu.Summary = nil
				if p.validateFieldPartySummary() == nil {
					t.Fatal("missing accepted")
				}
				return
			}
			b, _ := json.Marshal(p.Interface.FieldStatusMenu.Summary)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "unknown":
				m["guess"] = 1
			case "missing_zero":
				delete(m["name"].(map[string]any), "step_y")
			case "missing_window":
				delete(m, "money_window")
			case "unknown_text":
				m["column_text_id"] = "unknown"
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D1"
			case "width":
				m["width_base"] = 0
			case "step":
				m["column_step"] = 0
			case "capacity":
				m["max_columns"] = 1000
			case "name":
				m["name_limit"] = 1000
			case "name_overflow":
				m["name"].(map[string]any)["x"] = 1000
			case "money":
				m["money_value"].(map[string]any)["digits"] = 1000
			case "hp":
				m["hp"].(map[string]any)["x"] = 0
			case "return":
				m["return_mode"] = "guess"
			}
			b, _ = json.Marshal(m)
			var s FieldPartySummary
			e = json.Unmarshal(b, &s)
			if e == nil {
				p.Interface.FieldStatusMenu.Summary = &s
				e = p.validateFieldPartySummary()
			}
			if e == nil {
				t.Fatal("broken contract accepted", name)
			}
		})
	}
}
