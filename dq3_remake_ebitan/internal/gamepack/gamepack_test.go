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

func TestDQ3RomalyTemporaryRoleMatchesOriginalEXECTYAndText(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("original %s unavailable: %v", name, err)
		}
		return b
	}
	exe, cty, txt := read("DQ3.EXE"), read("CTY02.DAT"), read("D3TXT02.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.TemporaryRoleEvent("dq3:event.romaly_crown_kingship")
	if !ok {
		t.Fatal("dq3:event.romaly_crown_kingship missing")
	}
	if e.OfferNPC != (ScriptedNPCSelector{
		CTYRaw: 2, Section: 1, Tile: TileCoordinate{X: 7, Y: 2}, HandlerRaw: 9,
	}) || e.RestoreNPC != (ScriptedNPCSelector{
		CTYRaw: 2, Section: 3, Tile: TileCoordinate{X: 4, Y: 11}, HandlerRaw: 13,
	}) || e.PendingFlagRaw != 0x2c || e.RequiredItemRawID != 0x33 ||
		e.NormalRoleFlagRaw != 0x2b || e.ActiveRoleFlagRaw != 0x27 ||
		e.RoleSpriteEntries != (GenderedSpriteEntryBases{Male: 64, Female: 116}) {
		t.Fatalf("羅馬利亞 temporary_role JSON 設定不符：%+v", e)
	}

	// CTY02 原始 7-byte NPC record：王座 handler9 與玩家為王時才顯示的
	// 地下競技場 handler13。後者不是先前誤記的 jump-table slot 18/handler11。
	findNPC := func(section int, want []byte) bool {
		so := int(binary.LittleEndian.Uint16(cty[section*2:]))
		table := so + int(binary.LittleEndian.Uint16(cty[so:]))
		for i := 0; i < int(cty[table]); i++ {
			rec := cty[table+1+i*7 : table+1+(i+1)*7]
			if reflect.DeepEqual(rec, want) {
				return true
			}
		}
		return false
	}
	if !findNPC(1, []byte{7, 2, 4, 16, 9, 43, 0}) ||
		!findNPC(3, []byte{4, 11, 31, 16, 13, 39, 0}) {
		t.Fatal("CTY02 找不到 JSON 指定的王座／地下競技場 scripted NPC raw record")
	}
	// CTY02 sec0 的兩份 NPC table 也解釋 production trace 的實際 gate：
	// 白天守衛站 (13,6)/(16,6)，中央可通；夜間改站 (14,6)/(15,6) 封宮。
	tableHas := func(table int, want []byte) bool {
		for i := 0; i < int(cty[table]); i++ {
			rec := cty[table+1+i*7 : table+1+(i+1)*7]
			if reflect.DeepEqual(rec, want) {
				return true
			}
		}
		return false
	}
	sec0 := int(binary.LittleEndian.Uint16(cty[0:]))
	dayTable := sec0 + int(binary.LittleEndian.Uint16(cty[sec0:]))
	nightTable := sec0 + int(binary.LittleEndian.Uint16(cty[sec0+2:]))
	if !tableHas(dayTable, []byte{13, 6, 22, 0, 6, 43, 0}) ||
		!tableHas(dayTable, []byte{16, 6, 22, 0, 6, 43, 0}) ||
		!tableHas(nightTable, []byte{14, 6, 22, 0, 13, 40, 0}) ||
		!tableHas(nightTable, []byte{15, 6, 22, 0, 13, 40, 0}) {
		t.Fatal("CTY02 日／夜守衛 gate raw record 不符")
	}

	for off, want := range map[int][]byte{
		0x6623: {0xc7, 0x06, 0x93, 0x25, 0x33, 0x00}, // selected item=0x33
		0x6649: {0xc7, 0x04, 0xff, 0x00},             // consume matched slot
		0x664d: {0xbb, 0x2c, 0x00},                   // clear pending flag
		0x667f: {0xbb, 0x2b, 0x00},                   // clear normal-role flag
		0x6685: {0xbb, 0x27, 0x00},                   // set active-role flag
		0x669b: {0xb8, 0x1d, 0x00},                   // female AX before bp=1 sprite load
		0x66a1: {0xb8, 0x10, 0x00},                   // male AX before bp=1 sprite load
		0x679e: {0xbb, 0x27, 0x00},                   // restore: clear active-role
		0x67a4: {0xbb, 0x2b, 0x00},                   // restore: set normal-role
	} {
		if off+len(want) > len(exe) || !reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x 不符：got %x want %x", off, exe[off:off+len(want)], want)
		}
	}

	records := map[string]int{
		e.DialogueTextIDs.QuestGreeting: 45, e.DialogueTextIDs.QuestRequest: 15,
		e.DialogueTextIDs.ReturnPraise: 49, e.DialogueTextIDs.ForcedOffer: 50,
		e.DialogueTextIDs.ForcedReject: 51, e.DialogueTextIDs.Accept: 48,
		e.DialogueTextIDs.LaterIntro: 46, e.DialogueTextIDs.LaterOffer: 47,
		e.DialogueTextIDs.LaterReject: 52, e.DialogueTextIDs.RestoreIntro: 68,
		e.DialogueTextIDs.ContinueRole: 69, e.DialogueTextIDs.RestoreReconsider: 70,
		e.DialogueTextIDs.RestoreReject: 71, e.DialogueTextIDs.RestoreAccept: 72,
	}
	for id, record := range records {
		start := int(binary.LittleEndian.Uint16(txt[record*2:]))
		end := int(binary.LittleEndian.Uint16(txt[(record+1)*2:]))
		raw := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			raw = append(raw, binary.LittleEndian.Uint16(txt[off:off+2]))
		}
		got, found := p.TextGlyphCodes(id)
		def, defined := p.TextDefinition(id)
		if !found || !defined || def.Source.Record == nil || *def.Source.Record != record ||
			!reflect.DeepEqual(got, raw) {
			t.Fatalf("%s 未與 D3TXT02 rec%d 完整一致", id, record)
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
	  "schema_version":"0.1.0","boss_surrender_events":[],"temporary_role_events":[]
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
