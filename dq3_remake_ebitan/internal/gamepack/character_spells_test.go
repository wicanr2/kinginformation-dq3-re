package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestCharacterSpellsOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.CharacterSpells
	b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "DQ3.EXE"))
	if e != nil {
		t.Fatal(e)
	}
	if len(b) != 115282 || fmt.Sprintf("%x", sha256.Sum256(b)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original identity")
	}
	if hex.EncodeToString(b[0x1a074:0x1a08c]) != "120313002e002c006000d7010100d801d901000000000000" || s.RawWindow.X != 19 || s.RawWindow.Y != 46 || s.RawWindow.Width != 44 || s.RawWindow.Height != 96 || s.RawWindow.Flags != 3 || s.Columns != 4 || s.ExtraRows != 2 || s.Names != (GeometryAnchor{X: 168, Y: 62, StepX: 80, StepY: 16}) {
		t.Fatal("original window contract")
	}
	for _, row := range []struct {
		linear int
		raw    string
	}{{0x184a1, "8b1e2207"}, {0x185c1, "9a14021b11"}, {0x185d7, "9adb000411"}} {
		raw, e := hex.DecodeString(row.raw)
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(b[row.linear-0xec90:row.linear-0xec90+len(raw)], raw) {
			t.Fatalf("original IDA linear%x bytes differ", row.linear)
		}
	}
	b, e = os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "D3TXT00.TXT"))
	if e != nil {
		t.Fatal(e)
	}
	source := dq3data.LoadText(nil, b)
	if len(b) != 18680 || source.NRecords != 759 || len(s.Catalog) != 60 {
		t.Fatal("original archive shape")
	}
	refs := map[string]int{s.HeaderTextID: 471, s.RowTextID: 472, s.FooterTextID: 473}
	for i, row := range s.Catalog {
		if row.RecordRaw != 121+i {
			t.Fatal("original catalog order")
		}
		refs[row.TextID] = row.RecordRaw
	}
	for id, record := range refs {
		text, ok := p.TextDefinition(id)
		codes, _ := p.TextGlyphCodes(id)
		if !ok || text.Source.File != "D3TXT00.TXT" || *text.Source.Record != record || !reflect.DeepEqual(codes, source.Record(record)) {
			t.Fatalf("record%d pack glyphs differ", record)
		}
	}
	t.Log("pack canonical", p.ContentHash())
}

func TestCharacterSpellsUnionRejectsUnknownWithoutMutation(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	s := p.Interface.CharacterSpells
	input := []int{180, 161, 121, 161}
	before := append([]int(nil), input...)
	ids, ok := s.OrderedTextIDs(input)
	if !ok || !reflect.DeepEqual(ids, []string{s.Catalog[0].TextID, s.Catalog[40].TextID, s.Catalog[59].TextID}) || !reflect.DeepEqual(input, before) {
		t.Fatal("union modified input or order")
	}
	if _, ok = s.OrderedTextIDs([]int{161, -1}); ok {
		t.Fatal("unknown record accepted")
	}
}

func TestCharacterSpellsRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "columns", "rows", "bounds", "duplicate", "unknown_text", "control", "name_width", "record", "evidence"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			s := p.Interface.CharacterSpells
			switch name {
			case "missing":
				p.Interface.CharacterSpells = nil
			case "columns":
				s.Columns = 0
			case "rows":
				s.ExtraRows = 0
			case "bounds":
				s.Names.X = 639
			case "duplicate":
				s.Catalog[1].RecordRaw = s.Catalog[0].RecordRaw
			case "unknown_text":
				s.HeaderTextID = "unknown"
			case "control":
				text, _ := p.TextDefinition(s.RowTextID)
				text.GlyphCodes[0] = dq3data.TxtNL
			case "name_width":
				s.Names.StepX = 16
			case "record":
				s.Catalog[0].RecordRaw = -1
			case "evidence":
				s.Evidence.Level = "D1"
			}
			if e = p.validateCharacterSpells(); e == nil {
				t.Fatal("broken contract accepted")
			}
		})
	}
	for _, missing := range []string{"names.x", "raw_window.flags"} {
		t.Run(missing, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			b, _ := json.Marshal(p.Interface.CharacterSpells)
			var data map[string]json.RawMessage
			json.Unmarshal(b, &data)
			key, field := "names", "x"
			if missing == "raw_window.flags" {
				key, field = "raw_window", "flags"
			}
			var nested map[string]any
			json.Unmarshal(data[key], &nested)
			delete(nested, field)
			data[key], _ = json.Marshal(nested)
			b, _ = json.Marshal(data)
			var result CharacterSpells
			if json.Unmarshal(b, &result) == nil {
				t.Fatal("omitted geometry accepted")
			}
		})
	}
}
