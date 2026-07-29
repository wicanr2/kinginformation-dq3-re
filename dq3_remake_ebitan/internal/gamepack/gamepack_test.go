package gamepack

import (
	"bytes"
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

func TestDQ3NoanielQuestItemChainMatchesOriginalEXECTYAndText(t *testing.T) {
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
	exe := read("DQ3.EXE")
	cty4, cty5, cty11 := read("CTY04.DAT"), read("CTY05.DAT"), read("CTY11.DAT")
	txt0, txt2 := read("D3TXT00.TXT"), read("D3TXT02.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.QuestItemChainEvent("dq3:event.noaniel_awakening")
	if !ok {
		t.Fatal("dq3:event.noaniel_awakening missing")
	}
	if e.Treasure != (QuestTreasureSelector{
		CTYRaw: 11, Section: 3, TileSubID: 0, EventTypeRaw: 3,
		ItemRawID: 0x59, PresentFlag: 0x2f,
	}) || e.Exchange.NPC != (ScriptedNPCSelector{
		CTYRaw: 5, Section: 0, Tile: TileCoordinate{X: 17, Y: 7}, HandlerRaw: 16,
	}) || e.Exchange.RequiredItemRawID != 0x59 || e.Exchange.GrantedItemRawID != 0x5a ||
		e.Use.ItemRawID != 0x5a || e.Use.LocationKind != "town" || e.Use.CTYRaw != 4 ||
		!reflect.DeepEqual(e.Use.SetFlagsRaw, []int{0x26}) ||
		!reflect.DeepEqual(e.Use.ClearFlagsRaw, []int{0x31}) || !e.Use.ReloadScene {
		t.Fatalf("諾亞尼爾 quest_item_chain JSON 設定不符：%+v", e)
	}

	// CTY11 sec3 examine subid0 = type3, item0x59, armed flag0x2f.
	sec11 := int(binary.LittleEndian.Uint16(cty11[e.Treasure.Section*2:]))
	ev11 := sec11 + int(binary.LittleEndian.Uint16(cty11[sec11+8:]))
	entry11 := cty11[ev11+1 : ev11+5]
	if !reflect.DeepEqual(entry11, []byte{3, 0x59, 0, 0x2f}) {
		t.Fatalf("CTY11 sec3 subid0=%x, want 0359002f", entry11)
	}

	// CTY05 raw NPC record identifies the exchange consumer, not a Go table.
	sec5 := int(binary.LittleEndian.Uint16(cty5[e.Exchange.NPC.Section*2:]))
	npc5 := sec5 + int(binary.LittleEndian.Uint16(cty5[sec5:]))
	foundQueen := false
	for i := 0; i < int(cty5[npc5]); i++ {
		rec := cty5[npc5+1+i*7 : npc5+1+(i+1)*7]
		if int(rec[0]) == e.Exchange.NPC.Tile.X && int(rec[1]) == e.Exchange.NPC.Tile.Y &&
			(rec[3]>>3)&7 == 2 && int(rec[4]) == e.Exchange.NPC.HandlerRaw {
			foundQueen = true
			break
		}
	}
	if !foundQueen {
		t.Fatal("CTY05 sec0 找不到 JSON 指定的 handler16 精靈女王 raw record")
	}

	// CTY04 pairs awake flag0x26 records with sleeping flag0x31 records.
	sec4 := int(binary.LittleEndian.Uint16(cty4[:2]))
	npc4 := sec4 + int(binary.LittleEndian.Uint16(cty4[sec4:]))
	countByFlag := map[byte]int{}
	for i := 0; i < int(cty4[npc4]); i++ {
		rec := cty4[npc4+1+i*7 : npc4+1+(i+1)*7]
		countByFlag[rec[5]]++
	}
	if countByFlag[0x26] != 11 || countByFlag[0x31] != 11 {
		t.Fatalf("CTY04 sec0 awake/sleep records=%v, want flag26=11 flag31=11", countByFlag)
	}

	for off, want := range map[int][]byte{
		0x5320: {0x83, 0x3e, 0x6c, 0x25, 0x04},       // current CTY == 4
		0x5330: {0xe8, 0x36, 0x0d},                   // consume selected item
		0x5333: {0xbb, 0x26, 0x00, 0xe8, 0x16, 0x2f}, // SET 0x26
		0x5339: {0xbb, 0x31, 0x00, 0xe8, 0x25, 0x2f}, // CLEAR 0x31
		0x533f: {0xbf, 0x57, 0x02},                   // global text rec0x257 = 599
		0x68f7: {0xc7, 0x06, 0x93, 0x25, 0x59, 0x00}, // selected item 0x59
		0x690a: {0xc7, 0x04, 0x5a, 0x00},             // replace slot with 0x5a
	} {
		if off+len(want) > len(exe) || !reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x 不符：got %x want %x", off, exe[off:off+len(want)], want)
		}
	}

	checkText := func(id string, file []byte, record int) {
		t.Helper()
		start := int(binary.LittleEndian.Uint16(file[record*2:]))
		end := int(binary.LittleEndian.Uint16(file[(record+1)*2:]))
		raw := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			raw = append(raw, binary.LittleEndian.Uint16(file[off:off+2]))
		}
		got, found := p.TextGlyphCodes(id)
		def, defined := p.TextDefinition(id)
		if !found || !defined || def.Source.Record == nil || *def.Source.Record != record ||
			!reflect.DeepEqual(got, raw) {
			t.Fatalf("%s 未與原始文字 rec%d 完整一致", id, record)
		}
	}
	checkText(e.Exchange.BeforeTextID, txt2, 90)
	checkText(e.Exchange.SuccessTextID, txt2, 96)
	checkText(e.Use.SuccessTextID, txt0, 599)
}

func TestDQ3PyramidSwitchGateMatchesOriginalEXECTYAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY13.DAT"), read("D3TXT03.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.TwoStepFloorSwitchGate("dq3:event.pyramid_magic_key_gate")
	if !ok {
		t.Fatal("dq3:event.pyramid_magic_key_gate missing")
	}
	if e.CTYRaw != 13 || e.Section != 2 || e.CompletionTileSubID != 3 ||
		e.ClearFlagRaw != 0x46 || e.TrapTransitionSubIDRaw != 3 ||
		e.TrapDestination != (SceneDestination{CTYRaw: 13, Section: 1, X: 13, Y: 28}) ||
		e.UnlockedTreasure != (QuestTreasureSelector{
			CTYRaw: 13, Section: 2, TileSubID: 1, EventTypeRaw: 1,
			ItemRawID: 0x56, PresentFlag: 0x73,
		}) || e.SuccessSFXRaw != 0x0b {
		t.Fatalf("金字塔 gate JSON 設定不符：%+v", e)
	}

	const sectionBase = 0x1a1d
	if got := int(binary.LittleEndian.Uint16(cty[sectionBase+4:])); got != 0x19 {
		t.Fatalf("CTY13 sec2 handler table relative=%#x, want 0x19", got)
	}
	handlerTable := sectionBase + 0x19
	if !reflect.DeepEqual(cty[handlerTable:handlerTable+3], []byte{21, 22, 23}) {
		t.Fatalf("CTY13 sec2 handlers=%x, want 151617", cty[handlerTable:handlerTable+3])
	}
	eventTable := sectionBase + int(binary.LittleEndian.Uint16(cty[sectionBase+8:]))
	wantEvents := []byte{
		0, 0x62, 0, 0x46,
		1, 0x56, 0, 0x73,
		1, 0x41, 0, 0x74,
	}
	if int(cty[eventTable]) < 3 ||
		!reflect.DeepEqual(cty[eventTable+1:eventTable+1+len(wantEvents)], wantEvents) {
		t.Fatalf("CTY13 sec2 event table=%x", cty[eventTable:eventTable+1+len(wantEvents)])
	}
	transitionTable := sectionBase + int(binary.LittleEndian.Uint16(cty[sectionBase+0x0c:]))
	if got := cty[transitionTable+3*4 : transitionTable+4*4]; !reflect.DeepEqual(got, []byte{13, 1, 13, 28}) {
		t.Fatalf("CTY13 sec2 transition subid3=%x, want 0d010d1c", got)
	}

	// Raw instruction anchors independently pin all three CTY handler entries
	// and handler23's clear-story-flag consumer.
	for off, want := range map[int][]byte{
		0x6969: {0xe8, 0x06, 0xfa, 0xbf, 0x0e, 0x0c},
		0x69c9: {0xe8, 0xa6, 0xf9, 0xbf, 0x0e, 0x0c},
		0x69f2: {0x8d, 0x36, 0x6e, 0x3e, 0xe8, 0x07},
	} {
		if off+len(want) > len(exe) || !reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x handler anchor=%x, want %x",
				off, exe[off:off+len(want)], want)
		}
	}
	if off := bytes.Index(exe[0x69f2:0x6a60], []byte{0xbb, 0x46, 0x00}); off < 0 {
		t.Fatal("handler23 找不到 clear flag 0x46 的原始指令 anchor")
	}

	for id, record := range map[string]int{
		e.DialogueTextIDs.Prompt: 86, e.DialogueTextIDs.Pressed: 87,
		e.DialogueTextIDs.Trap: 88, e.DialogueTextIDs.Success: 89,
	} {
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
			t.Fatalf("%s 未與 D3TXT03 rec%d 完整一致", id, record)
		}
	}
}

func TestDQ3PortogaRoyalMissionMatchesOriginalEXECTYAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY37.DAT"), read("D3TXT04.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.StagedVehicleExchangeEvent("dq3:event.portoga_royal_mission")
	if !ok {
		t.Fatal("dq3:event.portoga_royal_mission missing")
	}
	if e.NPC != (ScriptedNPCSelector{
		CTYRaw: 37, Section: 0, Tile: TileCoordinate{X: 9, Y: 6}, HandlerRaw: 26,
	}) || e.QuestPresentFlagRaw != 0x37 || e.GrantedItemRawID != 0x5b ||
		e.ExchangeAvailableFlagRaw != 0x38 || e.RequiredItemRawID != 0x5c ||
		e.Vehicle.Kind != "ship" ||
		e.Vehicle.WorldPosition != (VehicleWorldPosition{X: 25, Y: 73, Layer: 0}) ||
		!reflect.DeepEqual(e.Vehicle.ClearFlagsRaw, []int{0x2c}) {
		t.Fatalf("波魯多加 pack transaction 不符：%+v", e)
	}

	section := int(binary.LittleEndian.Uint16(cty[e.NPC.Section*2:]))
	npcTable := section + int(binary.LittleEndian.Uint16(cty[section:]))
	foundNPC := false
	for i := 0; i < int(cty[npcTable]); i++ {
		rec := cty[npcTable+1+i*7 : npcTable+1+(i+1)*7]
		if int(rec[0]) == e.NPC.Tile.X && int(rec[1]) == e.NPC.Tile.Y &&
			(rec[3]>>3)&7 == 2 && int(rec[4]) == e.NPC.HandlerRaw {
			foundNPC = true
			break
		}
	}
	if !foundNPC {
		t.Fatal("CTY37 sec0 找不到 pack 指定的 handler26 國王 raw record")
	}

	for off, want := range map[int][]byte{
		0x6ae4: {0xbb, 0x37, 0x00, 0xe8, 0x8f, 0x17},       // GET flag0x37
		0x6af1: {0xbb, 0x38, 0x00, 0xe8, 0x82, 0x17},       // GET flag0x38
		0x6afb: {0xc7, 0x06, 0x93, 0x25, 0x5c, 0x00},       // required item0x5c
		0x6b16: {0xc7, 0x04, 0xff, 0x00, 0xbb, 0x2c, 0x00}, // consume slot; flag0x2c
		0x6b25: {0xc7, 0x06, 0x3c, 0x4f, 0x19, 0x00,
			0xc7, 0x06, 0x3e, 0x4f, 0x49, 0x00}, // ship x=25,y=73
		0x6b47: {0xbf, 0xd0, 0x0b},                   // record24
		0x6b4f: {0xc7, 0x06, 0x93, 0x25, 0x5b, 0x00}, // grant item0x5b
		0x6b5f: {0xbb, 0x37, 0x00, 0xe8, 0xff, 0x16}, // CLEAR flag0x37
		0x6b65: {0xbf, 0xd1, 0x0b},                   // record25
	} {
		if off+len(want) > len(exe) || !reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x 不符：got %x want %x",
				off, exe[off:off+len(want)], want)
		}
	}

	for id, record := range map[string]int{
		e.DialogueTextIDs.QuestIntro: 24, e.DialogueTextIDs.ItemGrant: 25,
		e.DialogueTextIDs.NeedItem: 26, e.DialogueTextIDs.Success: 27,
		e.DialogueTextIDs.After: 28,
	} {
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
			t.Fatalf("%s 未與 D3TXT04 rec%d 完整一致", id, record)
		}
	}
}

func TestDQ3GarunaSatoriBookMatchesOriginalEXEAndCTY(t *testing.T) {
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
	exe, cty := read("DQ3.EXE"), read("CTY18.DAT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.TreasureEvent("dq3:event.garuna_satori_book")
	if !ok {
		t.Fatal("dq3:event.garuna_satori_book missing")
	}
	want := QuestTreasureSelector{
		CTYRaw: 18, Section: 1, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x4a, PresentFlag: 0xb7,
	}
	if e.Treasure != want {
		t.Fatalf("加爾那領悟之書 JSON selector=%+v, want %+v", e.Treasure, want)
	}

	sectionBase := int(binary.LittleEndian.Uint16(cty[2:4]))
	eventTable := sectionBase + int(binary.LittleEndian.Uint16(cty[sectionBase+8:]))
	if eventTable+5 > len(cty) || cty[eventTable] != 1 ||
		!reflect.DeepEqual(cty[eventTable+1:eventTable+5], []byte{0x01, 0x4a, 0x00, 0xb7}) {
		t.Fatalf("CTY18 sec1 subid0 raw event 不符：base=%#x table=%#x raw=%x",
			sectionBase, eventTable, cty[eventTable:eventTable+5])
	}

	for off, raw := range map[int][]byte{
		0x9d86: {0x8b, 0xcd, 0x80, 0xf9, 0x03, 0x74, 0x10, 0x80, 0xf9, 0x00},
		0x9de9: {0xc6, 0x06, 0x26, 0x07, 0x00, 0x90, 0x80, 0xf9, 0x03},
	} {
		if off+len(raw) > len(exe) || !reflect.DeepEqual(exe[off:off+len(raw)], raw) {
			t.Fatalf("DQ3.EXE treasure consumer file %#x bytes 不符", off)
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

func TestDQ3DhamaReclassMatchesOriginalTextAndConfig(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT06.TXT"))
	if err != nil {
		t.Skipf("original D3TXT06.TXT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	event, ok := p.ReclassEvent("dq3:event.dhama_reclass")
	if !ok {
		t.Fatal("missing dq3:event.dhama_reclass")
	}
	if event.NPC.CTYRaw != 17 || event.NPC.Section != 0 ||
		event.NPC.Tile.X != 10 || event.NPC.Tile.Y != 10 ||
		event.NPC.HandlerRaw != 39 || event.MinimumLevel != 20 ||
		!reflect.DeepEqual(event.BasicTargetClasses, []int{1, 2, 3, 4, 6}) ||
		event.AdvancedTargetClass != 5 || event.AdvancedFreeSourceClass != 7 ||
		event.AdvancedRequiredItemRaw != 0x4a || event.PersonalInventorySlots != 8 {
		t.Fatalf("轉職 JSON 與 EXE/CTY17 證據不符：%+v", event)
	}
	records := map[string]int{
		event.DialogueTextIDs.Introduction:      45,
		event.DialogueTextIDs.Decline:           46,
		event.DialogueTextIDs.ChooseMember:      47,
		event.DialogueTextIDs.HeroForbidden:     48,
		event.DialogueTextIDs.LevelInsufficient: 49,
		event.DialogueTextIDs.ChooseClass:       50,
		event.DialogueTextIDs.SameClass:         51,
		event.DialogueTextIDs.ConfirmClass:      52,
		event.DialogueTextIDs.ConfirmLevelReset: 53,
		event.DialogueTextIDs.Success:           54,
		event.DialogueTextIDs.Farewell:          55,
		event.DialogueTextIDs.ClassMenuBasic:    66,
		event.DialogueTextIDs.ClassMenuSage:     67,
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
		if !found || !defined || def.Source.Record == nil ||
			*def.Source.Record != record || !reflect.DeepEqual(got, raw) {
			t.Fatalf("%s 未與 D3TXT06 rec%d 完整一致", id, record)
		}
	}
	txt00, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Skipf("original D3TXT00.TXT unavailable: %v", err)
	}
	rec421Start := int(binary.LittleEndian.Uint16(txt00[421*2:]))
	rec421End := int(binary.LittleEndian.Uint16(txt00[422*2:]))
	rec421 := make([]uint16, 0, (rec421End-rec421Start)/2)
	for off := rec421Start; off+2 <= rec421End; off += 2 {
		rec421 = append(rec421, binary.LittleEndian.Uint16(txt00[off:off+2]))
	}
	actions := p.ItemActions()
	for id, want := range map[string][]uint16{
		actions.TextIDs.Use:  rec421[9:11],
		actions.TextIDs.Give: rec421[16:18],
		actions.TextIDs.Drop: rec421[23:25],
	} {
		got, found := p.TextGlyphCodes(id)
		if !found || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s 未與 D3TXT00 rec421 對應 label 一致：got=%v want=%v",
				id, got, want)
		}
	}
}

func TestLoadRejectsUnknownAndInvalidData(t *testing.T) {
	validManifest := `{
	  "schema_version":"0.1.7","pack_id":"test","game":"dq3","edition":"cht_jingxun",
	  "content_version":"0.1.0","engine_api":">=0.1.0 <0.2.0",
	  "title_text_id":"x:title","entry_event_id":"x:new","save_namespace":"test",
	  "capabilities":[],"data":{"facilities":"facilities.json","events":"events.json",
	  "characters":"characters.json","texts":"texts.json"},"assets":{}
	}`
	validFacilities := `{
	  "schema_version":"0.1.7","service_definitions":[{
	    "id":"common:service.revive",
	    "pricing":{"formula_id":"common:formula.level_table","level_cap":1,"costs_gold":[10]},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"reader","doc":"docs/x.md"}
	  }]
	}`
	validEvents := `{
	  "schema_version":"0.1.7",
	  "item_actions":{"personal_inventory_slots":8,
	    "text_ids":{"use":"x:text","give":"x:text","drop":"x:text"},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"record","address":"421","consumer":"item menu","doc":"docs/x.md"}},
	  "boss_surrender_events":[],"temporary_role_events":[],
	  "quest_item_chain_events":[],"treasure_events":[],"two_step_floor_switch_gates":[],
	  "staged_vehicle_exchange_events":[],"guided_passage_events":[],
	  "hostage_rescue_events":[],"reclass_events":[]
	}`
	validCharacters := `{
	  "schema_version":"0.1.7",
	  "default_refs":{"new_game_player":"test:character.player"},
	  "defaults":[
	    {"id":"test:character.player",
	     "equipment":{"weapon":null,"armor":30,"shield":null,"head":null},
	     "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	       "address_space":"file","address":"0x1","consumer":"writer","doc":"docs/x.md"}}
	  ]
	}`
	validTexts := `{
	  "schema_version":"0.1.7","definitions":[{
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
		{"unknown facilities field", validManifest, strings.Replace(validFacilities, `"schema_version":"0.1.7"`, `"schema_version":"0.1.7","typo":1`, 1), validEvents, validCharacters, "unknown field"},
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

func TestHostageRescueRejectsInvalidTransactionsAndPaths(t *testing.T) {
	tests := []struct {
		name string
		edit func(*HostageRescueEvent)
		want string
	}{
		{
			name: "duplicate completion flag",
			edit: func(e *HostageRescueEvent) {
				e.CompletionSetFlagsRaw[0] = e.GuardPresentFlagRaw
			},
			want: "duplicate hostage rescue flag",
		},
		{
			name: "duplicate trigger tile",
			edit: func(e *HostageRescueEvent) {
				e.ApproachTrigger.Tiles = append(e.ApproachTrigger.Tiles,
					e.ApproachTrigger.Tiles[0])
			},
			want: "duplicate approach_trigger tile",
		},
		{
			name: "diagonal movement",
			edit: func(e *HostageRescueEvent) {
				start := e.ApproachMovement.Path[0]
				e.ApproachMovement.Path[1] = TileCoordinate{X: start.X + 1, Y: start.Y + 1}
			},
			want: "not a cardinal segment",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(&p.Events.HostageRescueEvents[0])
			err = p.validateEvents()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("validateEvents error=%v, want substring %q", err, tc.want)
			}
		})
	}
}
