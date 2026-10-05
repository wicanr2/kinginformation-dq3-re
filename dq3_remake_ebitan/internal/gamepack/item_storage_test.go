package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestItemStorageAcceptedOriginalInitialWords(t *testing.T) {
	directory := os.Getenv("DQ3_ITEM_ORACLE_DIR")
	if directory == "" {
		t.Skip("enable accepted original initial-word source with DQ3_ITEM_ORACLE_DIR")
	}
	raw, err := os.ReadFile(filepath.Join(directory, "issue4-field-item-reorder-r2-source-r3-receipt.json"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4" {
		t.Fatal("accepted original source differs", err)
	}
	var source struct {
		States []struct{ Packet, Actor string } `json:"states"`
	}
	if err := json.Unmarshal(raw, &source); err != nil {
		t.Fatal(err)
	}
	if len(source.States) != 230 || source.States[0].Packet != "1" {
		t.Fatal("normal original first packet missing")
	}
	actor, err := hex.DecodeString(source.States[0].Actor)
	if err != nil || len(actor) != 128 {
		t.Fatal("original actor shape differs", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	store, ok := p.NewGamePlayerItems()
	if !ok {
		t.Fatal("missing explicit initial words")
	}
	words := store.Words()
	for position, word := range words {
		// Original actor record +3A: the first normal input follows creation,
		// before the king's item reward. No compact/equipment inference.
		if word != binary.LittleEndian.Uint16(actor[0x3a+position*2:]) {
			t.Fatal("initial physical words differ", position)
		}
	}
}

func originalItemStorageInputs(t *testing.T) ([]byte, []byte, *dq3data.Items) {
	t.Helper()
	read := func(name, hash string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != hash {
			t.Fatal("original item storage input identity differs", name, err)
		}
		return raw
	}
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	raw := read("ITEM.DAT", "7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d")
	items, err := dq3data.OpenItems(raw)
	if err != nil {
		t.Fatal(err)
	}
	return exe, raw, items
}

func TestItemStorageOriginalEncodingInitialWordsAndMetadata(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	exe, raw, items := originalItemStorageInputs(t)
	s := p.Characters.ItemStorage
	word := func(file int) int { return int(binary.LittleEndian.Uint16(exe[file:])) }
	// File offsets below correspond to the original IDA9.4 instructions in
	// docs/188; compare identities/bytes before interpreting their immediates.
	for _, check := range []struct {
		offset   int
		expected []byte
	}{
		{0x4c99, []byte{0x3d, 0xff, 0}}, {0x4cab, []byte{0x81, 0x26, 0x91, 0x25, 0xff, 0}},
		{0x4d78, []byte{0xa9, 0, 0xe0}}, {0x9408, []byte{0xb8, 0, 0x80}},
		{0x9416, []byte{0xb8, 0, 0x40}}, {0x1c19, []byte{0xb9, 8, 0, 0xc7, 4, 0xff, 0}},
		{0x1c2d, []byte{0xb8, 0x1e, 0, 0x0d, 0, 0x80, 0x89, 0x44, 0x3a}},
		{0x1cb0, []byte{0x0d, 0, 0x80}}, // Registered-member writer OR worn bit.
	} {
		if !bytes.Equal(exe[check.offset:check.offset+len(check.expected)], check.expected) {
			t.Fatalf("original writer bytes differ at file%x", check.offset)
		}
	}
	if *s.Encoding.Empty != word(0x4c9a) || *s.Encoding.CodeMask != word(0x4caf) ||
		*s.Encoding.WornMask != word(0x9409) || *s.Encoding.CurseMask != word(0x9417) ||
		*s.Encoding.TransferBlockedMask != word(0x4d79) || len(s.Items) != len(raw)/7 || items.Count() != len(s.Items) {
		t.Fatal("word encoding or actual archive count differs")
	}
	if err := p.ValidateItemStorageAgainstItems(items); err != nil {
		t.Fatal(err)
	}
	want := make([]uint16, word(0x1c1a))
	for position := range want {
		want[position] = uint16(word(0x1c1e))
	}
	want[0] = uint16(word(0x1c2e) | word(0x1c31))
	for _, ref := range []string{p.Characters.DefaultRefs.NewGamePlayer, p.Characters.DefaultRefs.RegisteredPartyMember} {
		store, ok := p.CharacterItems(ref)
		if !ok || !reflect.DeepEqual(store.Words(), want) {
			t.Fatalf("explicit initial words differ for %s", ref)
		}
		preview, ok := p.CharacterEquipment(ref)
		if !ok || preview != [4]int{-1, 30, -1, -1} {
			t.Fatalf("equipment must derive from original initial words: %v", preview)
		}
		if _, ok := store.Add(0); !ok {
			t.Fatal("independent character words")
		}
		again, ok := p.CharacterItems(ref)
		if !ok || !reflect.DeepEqual(again.Words(), want) {
			t.Fatal("getter mutated immutable pack")
		}
	}
	t.Logf("schema=%s content=%s canonical=%s", p.Schema(), p.ContentVersion(), p.ContentHash())
}

func TestItemStorageRejectsMissingMalformedAndAmbiguousDefaults(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Pack)
	}{
		{"missing_storage", func(p *Pack) { p.Characters.ItemStorage = nil }},
		{"missing_empty", func(p *Pack) { p.Characters.ItemStorage.Encoding.Empty = nil }},
		{"missing_code_mask", func(p *Pack) { p.Characters.ItemStorage.Encoding.CodeMask = nil }},
		{"missing_worn_mask", func(p *Pack) { p.Characters.ItemStorage.Encoding.WornMask = nil }},
		{"missing_curse_mask", func(p *Pack) { p.Characters.ItemStorage.Encoding.CurseMask = nil }},
		{"missing_transfer_mask", func(p *Pack) { p.Characters.ItemStorage.Encoding.TransferBlockedMask = nil }},
		{"missing_drop_mask", func(p *Pack) { p.Characters.ItemStorage.DropBlockedMask = nil }},
		{"drop_identity_overlap", func(p *Pack) { *p.Characters.ItemStorage.DropBlockedMask |= 1 }},
		{"missing_drop_metadata", func(p *Pack) { p.Characters.ItemStorage.Items[0].DropForbidden = nil }},
		{"empty_is_valid_item", func(p *Pack) { *p.Characters.ItemStorage.Encoding.Empty = 0 }},
		{"out_of_word_range", func(p *Pack) { *p.Characters.ItemStorage.Encoding.Empty = 65536 }},
		{"overlapping_masks", func(p *Pack) {
			*p.Characters.ItemStorage.Encoding.CurseMask = *p.Characters.ItemStorage.Encoding.WornMask
		}},
		{"missing_parts", func(p *Pack) { p.Characters.ItemStorage.PartCount = nil }},
		{"missing_metadata", func(p *Pack) { p.Characters.ItemStorage.Items = nil }},
		{"missing_part", func(p *Pack) { p.Characters.ItemStorage.Items[0].EquipmentPart = nil }},
		{"missing_false", func(p *Pack) { p.Characters.ItemStorage.Items[0].CursedWhenWorn = nil }},
		{"invalid_part", func(p *Pack) { *p.Characters.ItemStorage.Items[0].EquipmentPart = 4 }},
		{"unreviewed", func(p *Pack) { p.Characters.ItemStorage.Evidence.Level = "D1" }},
		{"missing_words", func(p *Pack) { p.Characters.Defaults[0].ItemWords = nil }},
		{"initial_unreviewed", func(p *Pack) { p.Characters.Defaults[0].Evidence.Level = "D1" }},
		{"initial_not_dynamic", func(p *Pack) { p.Characters.Defaults[0].Evidence.Level = "D2" }},
		{"short_words", func(p *Pack) { p.Characters.Defaults[0].ItemWords = p.Characters.Defaults[0].ItemWords[:7] }},
		{"null_word", func(p *Pack) { p.Characters.Defaults[0].ItemWords[1] = nil }},
		{"negative_word", func(p *Pack) { *p.Characters.Defaults[0].ItemWords[1] = -1 }},
		{"invalid_archive_index", func(p *Pack) { *p.Characters.Defaults[0].ItemWords[1] = 128 }},
		{"duplicate_equipment_part", func(p *Pack) { *p.Characters.Defaults[0].ItemWords[1] = *p.Characters.Defaults[0].ItemWords[0] }},
		{"nongear_worn", func(p *Pack) {
			for code, item := range p.Characters.ItemStorage.Items {
				if *item.EquipmentPart < 0 {
					*p.Characters.Defaults[0].ItemWords[0] = code | *p.Characters.ItemStorage.Encoding.WornMask
					return
				}
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(p)
			if err := p.validateCharacters(); err == nil {
				t.Fatal("broken item contract accepted")
			}
		})
	}
}

func TestFieldItemsOriginalWindowsDropGateAndMarkers(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	exe, raw, _ := originalItemStorageInputs(t)
	s := p.Interface.FieldItems
	word := func(file int) int { return int(binary.LittleEndian.Uint16(exe[file:])) }
	for _, check := range []struct {
		linear int
		bytes  []byte
	}{
		{0x13801, []byte{5, 2, 0}}, {0x1380d, []byte{5, 16, 0}},
		{0x1381c, []byte{5, 2, 0, 0xb1, 4, 0xd3, 0xe0}},
		{0x13ad5, []byte{0xa9, 0, 0xe0}}, {0x13aeb, []byte{0xa8, 2}},
	} {
		offset := check.linear - 0xec90
		if !bytes.Equal(exe[offset:offset+len(check.bytes)], check.bytes) {
			t.Fatalf("original instruction differs: IDAlinear%x", check.linear)
		}
	}
	for _, check := range []struct {
		address int
		window  RawNewGameWindow
	}{{0x3fd8, s.RawWindow}, {0x4050, s.ActionWindow}} {
		base := 0x16140 + check.address
		w := check.window
		if w.Flags != word(base)>>8 || w.X != word(base+2) || w.Y != word(base+4) || w.Width != word(base+6) || w.Height != word(base+8) {
			t.Fatal("original window differs", check.address)
		}
	}
	base := 0x16140 + 0x3fd8
	action := 0x16140 + 0x4050
	if s.FrameRows != word(0x13802-0xec90) || s.RowStep != word(0x1380e-0xec90) || s.Name != (GeometryAnchor{X: (word(base+2) + 4) * 8, Y: word(base+4) + 16}) || s.Cursor != (GeometryAnchor{X: word(base+24) * 8, Y: word(base + 26)}) || s.Worn != (GeometryAnchor{X: word(base+2) * 8, Y: word(base+4) + 16}) || s.ActionCursor != (GeometryAnchor{X: word(action+24) * 8, Y: word(action + 26)}) {
		t.Fatal("original dynamic geometry differs")
	}
	if *s.MarkerMask != word(0x138c5-0xec90)|word(0x138ca-0xec90) || s.WornGlyph != word(0x138d0-0xec90) || *p.Characters.ItemStorage.DropBlockedMask != word(0x13ad6-0xec90) {
		t.Fatal("original marker/drop mask differs")
	}
	for code, item := range p.Characters.ItemStorage.Items {
		if *item.DropForbidden != (raw[code*7+5]&exe[0x13aec-0xec90] != 0) {
			t.Fatal("original drop metadata differs", code)
		}
	}
}

func TestFieldItemsRejectsMissingGeometryAndTextShape(t *testing.T) {
	for _, edit := range []func(*Pack){
		func(p *Pack) { p.Interface.FieldItems = nil },
		func(p *Pack) { p.Interface.FieldItems.MarkerMask = nil },
		func(p *Pack) { *p.Interface.FieldItems.MarkerMask |= 1 },
		func(p *Pack) { p.Interface.FieldItems.Name = GeometryAnchor{} },
		func(p *Pack) {
			p.Interface.FieldItems.Name.Y = p.Interface.FieldItems.RawWindow.Y + p.Interface.FieldItems.RawWindow.Height - dq3data.GlyphPx
		},
		func(p *Pack) {
			p.Interface.FieldItems.Cursor.Y = p.Interface.FieldItems.RawWindow.Y + p.Interface.FieldItems.RawWindow.Height - dq3data.GlyphPx
		},
		func(p *Pack) {
			p.Interface.FieldItems.Worn.Y = p.Interface.FieldItems.RawWindow.Y + p.Interface.FieldItems.RawWindow.Height - dq3data.GlyphPx
		},
		func(p *Pack) { p.Interface.FieldItems.RawWindow.Width = 15 },
		func(p *Pack) { p.Interface.FieldItems.ActionWindow.Height = 64 },
		func(p *Pack) { p.Interface.FieldItems.TextIDs.Row = "unknown" },
		func(p *Pack) { p.Interface.FieldItems.Evidence.Level = "D2" },
	} {
		p, err := BuiltinDQ3()
		if err != nil {
			t.Fatal(err)
		}
		edit(p)
		if p.validateFieldItems() == nil {
			t.Fatal("invalid item UI accepted")
		}
	}
}

func TestFieldItemPromptOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	exe, _, _ := originalItemStorageInputs(t)
	s := p.Interface.FieldItems.GivePresentation
	raw := exe[0x19fae : 0x19fae+30]
	word := func(i int) int { return int(binary.LittleEndian.Uint16(raw[i*2:])) }
	r, w := s.RawWindow, s.Window
	if r.Flags != word(0)>>8 || r.X != word(1) || r.Y != word(2) || r.Width != word(3) || r.Height != word(4) || w.X != r.X*8 || w.Y != r.Y || w.Width != r.Width*8 || w.Height != r.Height || w.TextInsetX != 16 || w.TextInsetY != 16 || s.GlyphStepX != 24 || s.VariableCodeWords != 1 {
		t.Fatal("native give window/consumer differs")
	}
	for _, check := range []struct {
		linear int
		bytes  []byte
	}{
		{0x139ae, []byte{0xbf, 0x34, 0x01}},
		{0x21414, []byte{0xa1, 0x70, 0x3e}},
		{0x214fb, []byte{0x83, 0xc5, 0x03}},
	} {
		off := check.linear - 0xec90
		if !bytes.Equal(exe[off:off+len(check.bytes)], check.bytes) {
			t.Fatalf("native prompt instruction differs at IDAlinear%x", check.linear)
		}
	}
	for _, check := range []struct {
		id     string
		record int
	}{{s.FrameTextID, word(5)}, {p.Interface.FieldItems.TextIDs.GivePrompt, 308}} {
		d, ok := p.TextDefinition(check.id)
		if !ok || d.Source.Record == nil || *d.Source.Record != check.record {
			t.Fatal("native prompt record reference differs")
		}
		raw, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", d.Source.File))
		if err != nil {
			t.Fatal(err)
		}
		codes := dq3data.LoadText(nil, raw).Record(check.record)
		if len(codes) != len(d.GlyphCodes) {
			t.Fatal("prompt record shape differs")
		}
		for i, c := range codes {
			if int(c) != d.GlyphCodes[i] {
				t.Fatal("prompt source differs")
			}
		}
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.ContentVersion(), p.ContentHash())
}

func TestFieldItemPromptRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "null_window", "missing_zero", "unknown_field", "raw_mismatch", "bad_shadow", "bad_text", "bad_frame", "unreviewed", "missing_timing"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			b, _ := json.Marshal(p.Interface.FieldItems.GivePresentation)
			var m map[string]any
			json.Unmarshal(b, &m)
			switch name {
			case "missing":
				p.Interface.FieldItems.GivePresentation = nil
			case "null_window":
				m["window"] = nil
			case "missing_zero":
				delete(m["raw_window"].(map[string]any), "flags")
			case "unknown_field":
				m["guess"] = 1
			case "raw_mismatch":
				m["raw_window"].(map[string]any)["x"] = 0
			case "bad_shadow":
				m["shadow"].(map[string]any)["offset_x"] = 640
			case "bad_text":
				m["glyph_step_x"] = 100
			case "bad_frame":
				m["frame_text_id"] = p.Interface.FieldItems.TextIDs.Header
			case "unreviewed":
				m["evidence"].(map[string]any)["level"] = "D2"
			case "missing_timing":
				delete(m["window"].(map[string]any), "glyph_timing_evidence")
			}
			if name != "missing" {
				b, _ = json.Marshal(m)
				var s FieldItemPrompt
				if json.Unmarshal(b, &s) != nil {
					return
				}
				p.Interface.FieldItems.GivePresentation = &s
			}
			if p.validateFieldItems() == nil {
				t.Fatal("broken prompt accepted")
			}
		})
	}
}

func TestItemStorageBootRejectsActualArchiveMismatch(t *testing.T) {
	_, raw, items := originalItemStorageInputs(t)
	for _, name := range []string{"missing", "extra_record", "partial_record", "metadata_part", "metadata_curse"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			candidate := items
			switch name {
			case "missing":
				candidate = nil
			case "extra_record":
				candidate, err = dq3data.OpenItems(append(append([]byte(nil), raw...), make([]byte, 7)...))
			case "partial_record":
				candidate, err = dq3data.OpenItems(append(append([]byte(nil), raw...), 0))
			case "metadata_part":
				*p.Characters.ItemStorage.Items[0].EquipmentPart = 1
			case "metadata_curse":
				*p.Characters.ItemStorage.Items[0].CursedWhenWorn = !*p.Characters.ItemStorage.Items[0].CursedWhenWorn
			}
			if err != nil {
				t.Fatal(err)
			}
			if err := p.ValidateItemStorageAgainstItems(candidate); err == nil {
				t.Fatal("boot accepted archive mismatch")
			}
		})
	}
}

func TestItemStorageStrictDecoderRejectsLegacyAndUnknownFields(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(p.Characters)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"equipment", "unknown_mask", "null_word", "float_word"} {
		t.Run(name, func(t *testing.T) {
			var fields map[string]any
			if err := json.Unmarshal(raw, &fields); err != nil {
				t.Fatal(err)
			}
			switch name {
			case "equipment":
				fields["defaults"].([]any)[0].(map[string]any)["equipment"] = map[string]any{"armor": 30}
			case "unknown_mask":
				fields["item_storage"].(map[string]any)["encoding"].(map[string]any)["guessed_mask"] = 1
			case "null_word":
				fields["defaults"].([]any)[0].(map[string]any)["item_words"].([]any)[1] = nil
			case "float_word":
				fields["defaults"].([]any)[0].(map[string]any)["item_words"].([]any)[1] = 0.5
			}
			candidate, err := json.Marshal(fields)
			if err != nil {
				t.Fatal(err)
			}
			decoder := json.NewDecoder(bytes.NewReader(candidate))
			decoder.DisallowUnknownFields()
			err = decoder.Decode(&p.Characters)
			if err == nil {
				err = p.validateCharacters()
			}
			if err == nil {
				t.Fatal("strict item schema accepted legacy/ambiguous fields")
			}
		})
	}
}
