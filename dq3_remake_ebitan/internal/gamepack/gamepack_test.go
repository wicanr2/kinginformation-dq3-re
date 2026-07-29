package gamepack

import (
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDQ3KandarEventMatchesOriginalEXEAndCTY(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY10.DAT"))
	if err != nil {
		t.Skipf("original CTY10.DAT unavailable: %v", err)
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT02.TXT"))
	if err != nil {
		t.Skipf("original D3TXT02.TXT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.BossSurrenderEvent("dq3:event.shanpane_kandar")
	if !ok {
		t.Fatal("dq3:event.shanpane_kandar missing")
	}

	rawFormation, err := hex.DecodeString(e.Formation.RawBytesHex)
	if err != nil {
		t.Fatalf("formation raw_bytes_hex: %v", err)
	}
	const formationOff = 0x1aff5
	if formationOff+len(rawFormation) > len(exe) ||
		!reflect.DeepEqual(rawFormation, exe[formationOff:formationOff+len(rawFormation)]) {
		t.Fatalf("JSON formation does not match DQ3.EXE file %#x", formationOff)
	}

	const sectionBase = 0x225d
	handlerTable := sectionBase + int(binary.LittleEndian.Uint16(cty[sectionBase+4:]))
	if cty[handlerTable] != 0x0e {
		t.Fatalf("CTY10 sec5 special handler=%#x, want 0x0e", cty[handlerTable])
	}
	if e.Trigger.CTYRaw != 10 || e.Trigger.Section != 5 || e.Trigger.TileSubID != 1 ||
		e.PresenceFlagRaw != 0x2e || e.ClearFlagRaw != 0x2e ||
		!reflect.DeepEqual(e.Trigger.Tiles, []TileCoordinate{{X: 6, Y: 8}, {X: 7, Y: 8}}) {
		t.Fatalf("JSON 甘達特事件的觸發點或旗標不符：%+v", e)
	}
	for id, want := range map[string]struct {
		record int
		value  string
	}{
		e.DialogueTextIDs.Intro:   {84, "*「來的好,但是你們仍然\n拿我沒辦法的。"},
		e.DialogueTextIDs.Apology: {85, "*「拜託金皇冠還給你們,\n原諒我好嗎?"},
		e.DialogueTextIDs.Accept:  {86, "*「謝謝,\n我絕對不會忘記你們的。"},
		e.DialogueTextIDs.Reject:  {87, "*「不要說這種話,\n原諒我吧,好嗎?"},
	} {
		def, ok := p.TextDefinition(id)
		if !ok || def.Source.Record == nil || *def.Source.Record != want.record ||
			def.Value != want.value {
			t.Fatalf("JSON 甘達特文字 %q 的 record/value 不符：%+v", id, def)
		}
		first := int(binary.LittleEndian.Uint16(txt[:2]))
		ptr := func(record int) int {
			off := record * 2
			if off+2 > first {
				t.Fatalf("D3TXT02 record %d 超出 pointer table", record)
			}
			return int(binary.LittleEndian.Uint16(txt[off : off+2]))
		}
		start, end := ptr(want.record), ptr(want.record+1)
		rawCodes := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			rawCodes = append(rawCodes, binary.LittleEndian.Uint16(txt[off:off+2]))
		}
		gotCodes, ok := p.TextGlyphCodes(id)
		if !ok || !reflect.DeepEqual(gotCodes, rawCodes) {
			t.Fatalf("JSON 甘達特文字 %q glyph_codes 與 D3TXT02 rec%d 不符",
				id, want.record)
		}
	}
}

func TestBuiltinDQ3ReviveCosts(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	for level, want := range map[int]int{0: 10, 1: 10, 3: 10, 4: 20, 40: 1610, 41: 1610} {
		if got := p.ReviveCost(level); got != want {
			t.Errorf("ReviveCost(%d)=%d, want %d", level, got, want)
		}
	}
}

func TestDQ3InitialEquipmentMatchesOriginalEXE(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	for _, tc := range []struct {
		id  string
		off int
	}{
		{"dq3:character.new_game_hero", 0x1c2d},
		{"dq3:character.registered_member", 0x1cad},
	} {
		eq, ok := p.CharacterEquipment(tc.id)
		if !ok {
			t.Fatalf("%s missing", tc.id)
		}
		if exe[tc.off] != 0xb8 || int(binary.LittleEndian.Uint16(exe[tc.off+1:])) != eq[1] ||
			!reflect.DeepEqual(exe[tc.off+3:tc.off+6], []byte{0x0d, 0x00, 0x80}) {
			t.Fatalf("%s JSON=%v does not match equipped-item writer at file %#x", tc.id, eq, tc.off)
		}
		if eq != [4]int{-1, 0x1e, -1, -1} {
			t.Fatalf("%s initial equipment=%v", tc.id, eq)
		}
	}
}

func TestDQ3ReviveCostsMatchOriginalEXE(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	const off = 0x19dac
	if len(exe) < off+40*2 {
		t.Fatalf("DQ3.EXE too small for revive table: %d", len(exe))
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	for level := 1; level <= 40; level++ {
		want := int(binary.LittleEndian.Uint16(exe[off+(level-1)*2:]))
		if got := p.ReviveCost(level); got != want {
			t.Errorf("level %d JSON=%d, DQ3.EXE file %#x=%d", level, got, off+(level-1)*2, want)
		}
	}
}

func TestLoadRejectsUnknownAndInvalidData(t *testing.T) {
	validManifest := `{
	  "schema_version":"0.1.0","pack_id":"test","game":"dq3","edition":"cht_jingxun",
	  "content_version":"0.1.0","engine_api":">=0.1.0 <0.2.0",
	  "title_text_id":"x:title","entry_event_id":"x:new","save_namespace":"test",
	  "capabilities":[],"data":{"facilities":"facilities.json","events":"events.json",
	  "characters":"characters.json","texts":"texts.json"},"assets":{}
	}`
	validFacilities := `{
	  "schema_version":"0.1.0","service_definitions":[{
	    "id":"common:service.revive",
	    "pricing":{"formula_id":"common:formula.level_table","level_cap":1,"costs_gold":[10]},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"reader","doc":"docs/x.md"}
	  }]
	}`
	validEvents := `{
	  "schema_version":"0.1.0","boss_surrender_events":[]
	}`
	validCharacters := `{
	  "schema_version":"0.1.0",
	  "default_refs":{"new_game_player":"test:character.player"},
	  "defaults":[
	    {"id":"test:character.player",
	     "equipment":{"weapon":null,"armor":30,"shield":null,"head":null},
	     "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	       "address_space":"file","address":"0x1","consumer":"writer","doc":"docs/x.md"}}
	  ]
	}`
	validTexts := `{
	  "schema_version":"0.1.0","definitions":[{
	    "id":"x:text","value":"字","glyph_codes":[1],
	    "layout":{"kind":"menu_label"},
	    "source":{"kind":"glyph_map","file":"font.bin"},
	    "evidence":{"level":"D3","source_kind":"data_file","source":"font.bin",
	      "address_space":"record","address":"glyph 1","consumer":"renderer","doc":"docs/x.md"}
	  }]
	}`
	tests := []struct {
		name, manifest, facilities, events, characters, want string
	}{
		{"unknown manifest field", strings.Replace(validManifest, `"assets":{}`, `"assets":{},"typo":1`, 1), validFacilities, validEvents, validCharacters, "unknown field"},
		{"path escape", strings.Replace(validManifest, `"facilities.json"`, `"../facilities.json"`, 1), validFacilities, validEvents, validCharacters, "pack-relative"},
		{"cost length", validManifest, strings.Replace(validFacilities, `"level_cap":1`, `"level_cap":2`, 1), validEvents, validCharacters, "must equal"},
		{"unknown facilities field", validManifest, strings.Replace(validFacilities, `"schema_version":"0.1.0"`, `"schema_version":"0.1.0","typo":1`, 1), validEvents, validCharacters, "unknown field"},
		{"unknown events field", validManifest, validFacilities, strings.Replace(validEvents, `"boss_surrender_events":[]`, `"boss_surrender_events":[],"typo":1`, 1), validCharacters, "unknown field"},
		{"unknown characters field", validManifest, validFacilities, validEvents, strings.Replace(validCharacters, `"defaults":[`, `"typo":1,"defaults":[`, 1), "unknown field"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fsys := fstest.MapFS{
				"manifest.json":   {Data: []byte(tc.manifest)},
				"facilities.json": {Data: []byte(tc.facilities)},
				"events.json":     {Data: []byte(tc.events)},
				"characters.json": {Data: []byte(tc.characters)},
				"texts.json":      {Data: []byte(validTexts)},
			}
			_, err := Load(fsys)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load error=%v, want substring %q", err, tc.want)
			}
		})
	}
}
