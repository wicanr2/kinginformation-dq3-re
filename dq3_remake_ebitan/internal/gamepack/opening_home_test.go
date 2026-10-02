package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

func TestOpeningHomeRejectsIncompleteContract(t *testing.T) {
	source, err := fs.Sub(builtin, "packs/dq3_cht")
	if err != nil {
		t.Fatal(err)
	}
	clone := func() fstest.MapFS {
		out := fstest.MapFS{}
		if err := fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			raw, err := fs.ReadFile(source, path)
			if err == nil {
				out[path] = &fstest.MapFile{Data: raw}
			}
			return err
		}); err != nil {
			t.Fatal(err)
		}
		return out
	}
	for _, name := range []string{"missing_home", "missing_leader", "null_picture", "unknown_field", "missing_coordinate", "null_facing", "missing_blank", "wrong_shape", "missing_text", "missing_asset", "bad_palette", "negative_background"} {
		t.Run(name, func(t *testing.T) {
			files := clone()
			var data map[string]any
			if err := json.Unmarshal(files["data/interface.json"].Data, &data); err != nil {
				t.Fatal(err)
			}
			escort := data["opening_escort"].(map[string]any)
			home := escort["home"].(map[string]any)
			picture := home["picture"].(map[string]any)
			switch name {
			case "missing_home":
				delete(escort, "home")
			case "missing_leader":
				delete(home, "leader_record")
			case "null_picture":
				home["picture"] = nil
			case "unknown_field":
				picture["unreviewed"] = true
			case "missing_coordinate":
				delete(picture["option_origin"].(map[string]any), "x")
			case "null_facing":
				escort["frames"].([]any)[0].(map[string]any)["leader_facing"] = nil
			case "missing_blank":
				delete(picture, "blank_glyph")
			case "wrong_shape":
				picture["questions"].([]any)[0] = []any{1}
			case "missing_text":
				picture["top_text_id"] = "unknown:text"
			case "missing_asset":
				picture["asset_key"] = "unknown:asset"
			case "bad_palette":
				picture["palette_overrides"].([]any)[0].(map[string]any)["rgb"] = []any{1}
			case "negative_background":
				picture["background_tile"] = -1
			}
			raw, err := json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			files["data/interface.json"].Data = raw
			if _, err := Load(files); err == nil {
				t.Fatal("incomplete or unreviewed home contract accepted")
			}
		})
	}
}

func TestOpeningHomeDQ3SourceParity(t *testing.T) {
	assets := os.Getenv("DQ3_ASSETS")
	if assets == "" {
		assets = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(assets, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("DQ3.EXE identity differs")
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	e, _ := p.OpeningEscort()
	h := e.Home
	s := h.Picture
	var table []byte
	for _, question := range s.Questions {
		for _, id := range question {
			table = append(table, byte(id))
		}
	}
	if !bytes.Equal(table, exe[0x16a00:0x16a00+300]) || s.RandomAdd != 0x9014 || s.RandomRotate != 3 || s.RandomMask != 127 {
		t.Fatal("independent picture generator or raw question table differs")
	}
	sequence := exe[0x19dfc:0x19e09]
	if !bytes.Equal(sequence, []byte{2, 1, 0, 6, 0, 0, 7, 3, 0, 1, 1, 2, 255}) {
		t.Fatal("raw mother sequence differs")
	}
	x, y := e.Frames[0].Leader.X, e.Frames[0].Leader.Y
	frame := 1
	for i := 0; i < len(sequence)-1; i += 3 {
		for repeat := 0; repeat < int(sequence[i]); repeat++ {
			rawDir := int(sequence[i+1])
			if sequence[i+2] != 2 {
				deltas := [][2]int{{0, 1}, {-1, 0}, {0, -1}, {1, 0}}
				x += deltas[rawDir][0]
				y += deltas[rawDir][1]
			}
			got := e.Frames[frame]
			if got.Leader.X != x || got.Leader.Y != y || got.LeaderFacing != []int{0, 2, 1, 3}[rawDir] || got.Player != e.Frames[0].Player {
				t.Fatalf("raw mother action %d differs", frame)
			}
			frame++
		}
	}
	if frame != len(e.Frames) {
		t.Fatal("raw mother sequence length differs")
	}
	window := exe[0x1a488 : 0x1a488+30]
	word := func(i int) int { return int(window[i]) | int(window[i+1])<<8 }
	if s.Window.X != word(2)*8 || s.Window.Y != word(4) || s.Window.Width != word(6)*8 || s.Window.Height != word(8) || s.BodyRows != word(12) || s.OptionCount != word(20) || word(22) != 4 {
		t.Fatal("raw picture window differs")
	}
	fon, err := os.ReadFile(filepath.Join(assets, "D3TXT00.FON"))
	if err != nil {
		t.Fatal(err)
	}
	bank, err := os.ReadFile(filepath.Join(assets, "D3TXT00.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	oracle := dq3data.LoadText(fon, bank)
	for _, pair := range []struct {
		id     string
		record int
	}{{s.TopTextID, word(10)}, {s.BodyTextID, word(14)}, {s.BottomTextID, word(16)}} {
		text, ok := p.TextDefinition(pair.id)
		if !ok || text.Source.Record == nil || *text.Source.Record != pair.record || text.Source.File != "D3TXT00.TXT" {
			t.Fatal("raw frame record reference differs")
		}
		raw := oracle.Record(pair.record)
		if len(raw) > 0 && raw[len(raw)-1] == 65535 {
			raw = raw[:len(raw)-1]
		}
		codes, ok := p.TextGlyphCodes(pair.id)
		if !ok || len(raw) != len(codes) {
			t.Fatal("raw frame glyph count differs")
		}
		for i, c := range raw {
			if c != codes[i] {
				t.Fatal("raw frame glyph differs")
			}
		}
	}
}
