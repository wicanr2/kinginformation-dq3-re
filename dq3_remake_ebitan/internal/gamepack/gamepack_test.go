package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestBuiltinDQ3FinalePackReferences(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	for key, want := range map[string]AssetRef{
		"title_image":   {Path: "TITG.P", Size: 39610, SHA256: "127fbed5603b8c0f6481b54e44aac0413384c1581c901780fbf0ffe947ecafa8"},
		"ending_image":  {Path: "TIT3.P", Size: 28948, SHA256: "f4c0b82c0df531ff7e8e10ce05f18b22e3e87a7d9a21237666b98722f132378e"},
		"world_palette": {Path: "DQ3.PAL", Size: 240, SHA256: "178a9ca809a33108d8e2a6430796eae8909d98e514fe543fc1923139110389c4"},
	} {
		got, ok := p.Asset(key)
		if !ok || got != want {
			t.Fatalf("pack asset %s = %+v/%v, want %+v", key, got, ok, want)
		}
	}
	if track, ok := p.AudioTrack("ending"); !ok || track != 17 {
		t.Fatalf("ending audio cue = %d/%v, want 17/true", track, ok)
	}
}

func TestBuiltinDQ3BattleFleeSoundCuesMatchOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := p.BattleSoundCues()
	if cues.PlayerFled.CueRaw != 13 || !cues.PlayerFled.WaitForCompletion ||
		cues.EnemyFled.CueRaw != 21 || !cues.EnemyFled.WaitForCompletion {
		t.Fatalf("battle flee sound cues=%+v", cues)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0xc843, []byte{0xbd, 0x0d, 0x00}}, // IDA 0x1b4d3: player flee cue13
		{0xbda7, []byte{0xbd, 0x15, 0x00}}, // IDA 0x1aa37: enemy flee cue21
	} {
		if anchor.off+len(anchor.raw) > len(exe) || !bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE flee cue anchor file0x%x mismatch", anchor.off)
		}
	}
	fvoc, err := os.ReadFile(filepath.Join(dir, "FVOC.VCX"))
	if err != nil {
		t.Skipf("original FVOC.VCX unavailable: %v", err)
	}
	bank := dq3data.DecodeVOCBank(fvoc, 44100)
	if len(bank) <= 21 || bank[13].FileOffset != 0x8498 || bank[13].SourceSamples != 3519 ||
		bank[21].FileOffset != 0x11146 || bank[21].SourceSamples != 2566 {
		t.Fatalf("FVOC flee descriptors cue13=%+v cue21=%+v", bank[13], bank[21])
	}
}

func TestBuiltinDQ3PlayerPhysicalSoundSequenceMatchesOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	steps := p.BattleSoundCues().PlayerPhysical.Steps
	if len(steps) != 2 || steps[0].CueRaw != 6 || !steps[0].WaitForCompletion ||
		steps[1].CueRaw != 11 || !steps[1].WaitForCompletion {
		t.Fatalf("player physical sound sequence=%+v, want cue6 then cue11 with completion waits", steps)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0xc8d8, []byte{0xbd, 0x06, 0x00}}, // IDA 0x1b568: first player physical cue
		{0xc8e0, []byte{0xbf, 0x4a, 0x01}}, // IDA 0x1b570: D3TXT record0x14a
		{0xc8ed, []byte{0xbd, 0x0b, 0x00}}, // IDA 0x1b57d: second player physical cue
	} {
		if anchor.off+len(anchor.raw) > len(exe) || !bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE player physical anchor file0x%x mismatch", anchor.off)
		}
	}
	fvoc, err := os.ReadFile(filepath.Join(dir, "FVOC.VCX"))
	if err != nil {
		t.Skipf("original FVOC.VCX unavailable: %v", err)
	}
	bank := dq3data.DecodeVOCBank(fvoc, 44100)
	if len(bank) <= 11 || bank[6].FileOffset != 0x215e || bank[6].BlockLength != 1224 ||
		bank[6].SourceRate != 6711 || bank[6].SourceSamples != 1222 ||
		bank[11].FileOffset != 0x7ad9 || bank[11].BlockLength != 1167 ||
		bank[11].SourceRate != 6711 || bank[11].SourceSamples != 1165 {
		t.Fatalf("FVOC player physical descriptors cue6=%+v cue11=%+v", bank[6], bank[11])
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	got, ok := p.TextGlyphCodes("common:text.battle.actor.attack")
	if !ok {
		t.Fatal("missing pack actor.attack text")
	}
	want := dq3data.LoadText(nil, txtRaw).Record(330)
	if !reflect.DeepEqual(got, want) || !reflect.DeepEqual(want, []uint16{0xfffb, 149, 623, 624, 57}) {
		t.Fatalf("actor.attack glyphs=%v want D3TXT00 record330=%v", got, want)
	}
}

func TestBuiltinDQ3CommonPhysicalResultMatchesOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := p.BattleSoundCues()
	for role, got := range map[string]BattleSoundCue{
		"enemy_physical_attack": cues.EnemyPhysicalAttack,
		"physical_hit":          cues.PhysicalHit,
		"physical_miss":         cues.PhysicalMiss,
	} {
		want := map[string]int{"enemy_physical_attack": 4, "physical_hit": 1, "physical_miss": 3}[role]
		if got.CueRaw != want || !got.WaitForCompletion {
			t.Fatalf("%s cue=%+v, want cue%d with completion wait", role, got, want)
		}
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0xbf7b, []byte{0xbd, 0x04, 0x00}}, // IDA 0x1ac0b: enemy physical cue4
		{0xbf83, []byte{0xbf, 0x4b, 0x01}}, // IDA 0x1ac13: enemy attack record331
		{0xc11e, []byte{0xbd, 0x01, 0x00}}, // IDA 0x1adae: successful damage cue1
		{0xc132, []byte{0xbf, 0x4c, 0x01}}, // IDA 0x1adc2: enemy-victim damage record332
		{0xc14b, []byte{0xbf, 0x4d, 0x01}}, // IDA 0x1addb: player-victim damage record333
		{0xc314, []byte{0xbf, 0x65, 0x01}}, // IDA 0x1afa4: player death record357
		{0xc323, []byte{0xbf, 0x50, 0x01}}, // IDA 0x1afb3: enemy defeated record336
		{0xc336, []byte{0xbd, 0x03, 0x00}}, // IDA 0x1afc6: physical miss cue3
		{0xc33e, []byte{0xbf, 0x4f, 0x01}}, // IDA 0x1afce: physical miss record335
	} {
		if anchor.off+len(anchor.raw) > len(exe) || !bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE common physical anchor file0x%x mismatch", anchor.off)
		}
	}
	fvoc, err := os.ReadFile(filepath.Join(dir, "FVOC.VCX"))
	if err != nil {
		t.Fatalf("read FVOC.VCX: %v", err)
	}
	bank := dq3data.DecodeVOCBank(fvoc, 44100)
	for cue, want := range map[int]struct {
		off, block, rate, samples int
	}{
		1: {0x4e6, 1163, 6493, 1161},
		3: {0x12b6, 1511, 6493, 1509},
		4: {0x18a2, 1155, 6493, 1153},
	} {
		got := bank[cue]
		if got.FileOffset != want.off || got.BlockLength != want.block ||
			got.SourceRate != want.rate || got.SourceSamples != want.samples {
			t.Fatalf("FVOC cue%d descriptor=%+v, want %+v", cue, got, want)
		}
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	table := dq3data.LoadText(nil, txtRaw)
	for id, record := range map[string]int{
		"common:text.battle.enemy.attack":   331,
		"common:text.battle.actor.damage":   332,
		"common:text.battle.party.damage":   333,
		"common:text.battle.actor.missed":   335,
		"common:text.battle.enemy.defeated": 336,
		"common:text.battle.actor.died":     357,
	} {
		got, ok := p.TextGlyphCodes(id)
		if !ok || !reflect.DeepEqual(got, table.Record(record)) {
			t.Fatalf("%s glyphs=%v/%v, want D3TXT00 record%d=%v", id, got, ok, record, table.Record(record))
		}
	}
}

func TestDQ3SpecialPhysicalParalysisPackMatchesOriginal(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	var action MonsterActionDefinition
	for _, candidate := range p.MonsterActionDefinitions() {
		if candidate.MaskBit == 3 {
			action = candidate
		}
	}
	if action.ActionRaw != 3 || action.Kind != "special_physical_condition" ||
		action.ConditionID != ParalysisCondition || action.TargetScope != "party_one_alive" ||
		action.DamageFormula != "ignore_defense_half_plus_random_quarter" ||
		!action.ConditionOnSurvival || action.CastTextRole != "fatal_strike" ||
		action.SuccessTextRole != "actor_paralyzed" {
		t.Fatalf("特殊物理 pack action=%+v", action)
	}
	condition, ok := p.ConditionDefinition(ParalysisCondition)
	if !ok || condition.Persistence != "persistent" || condition.LegacyStatusMask != 0x10 ||
		condition.FieldClearAfterSteps != 0x28 || condition.FieldDamagePerStep != 0 {
		t.Fatalf("麻痺 condition=%+v ok=%v", condition, ok)
	}
	if cue := p.BattleSoundCues().FatalStrike; cue.CueRaw != 9 || !cue.WaitForCompletion {
		t.Fatalf("致命一擊 cue=%+v", cue)
	}

	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0xc060, []byte{0x83, 0x3e, 0x11, 0x25, 0x03, 0x75, 0x06, 0xba, 0x00, 0x00}},
		{0xc22e, []byte{0xd1, 0xeb, 0x8b, 0xc3, 0xd1, 0xe8, 0x53, 0x8b, 0xd8}},
		{0xc1cc, []byte{0x83, 0x3e, 0x11, 0x25, 0x03, 0x75, 0x21, 0xbf, 0xc9, 0x00}},
		{0xc290, []byte{0xbd, 0x09, 0x00}},
		{0xd304, []byte{0xf7, 0x44, 0x38, 0xb0, 0x00}},
		{0xd559, []byte{0xf7, 0x44, 0x38, 0xb0, 0x00}},
		{0xdebe, []byte{0xf7, 0x45, 0x38, 0x10, 0x00}},
		{0xa8c3, []byte{0xf7, 0x06, 0x44, 0x4f, 0x10, 0x00}},
	} {
		if anchor.off+len(anchor.raw) > len(exe) ||
			!bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE 麻痺錨點 file0x%x 不一致", anchor.off)
		}
	}

	mons, err := os.ReadFile(filepath.Join(dir, "D3MNS.DAT"))
	if err != nil {
		t.Fatalf("read D3MNS.DAT: %v", err)
	}
	var users []int
	for id := 0; id < len(mons)/0x29; id++ {
		record := mons[id*0x29 : (id+1)*0x29]
		if record[0x0e]&0x10 != 0 {
			users = append(users, id)
			if !bytes.Equal(record[0x0e:0x14], []byte{0x10, 0, 0, 0, 0, 0}) {
				t.Fatalf("monster%d action mask=%x", id, record[0x0e:0x14])
			}
		}
	}
	if !reflect.DeepEqual(users, []int{3, 48, 51}) {
		t.Fatalf("mask bit3 users=%v，預期 [3 48 51]", users)
	}

	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	table := dq3data.LoadText(nil, txtRaw)
	for id, record := range map[string]int{
		"common:text.battle.fatal_strike":    334,
		"common:text.battle.actor.paralyzed": 201,
		"dq3:text.item_use.fullmoon.success": 364,
		"dq3:text.item_use.no_effect":        341,
	} {
		got, ok := p.TextGlyphCodes(id)
		if !ok || !reflect.DeepEqual(got, table.Record(record)) {
			t.Fatalf("%s glyphs=%v/%v，預期 record%d=%v", id, got, ok, record, table.Record(record))
		}
	}

	fvoc, err := os.ReadFile(filepath.Join(dir, "FVOC.VCX"))
	if err != nil {
		t.Fatalf("read FVOC.VCX: %v", err)
	}
	bank := dq3data.DecodeVOCBank(fvoc, 44100)
	if len(bank) <= 9 || bank[9].FileOffset != 0x4cd2 || bank[9].BlockLength != 2978 ||
		bank[9].SourceRate != 6711 || bank[9].SourceSamples != 2976 {
		t.Fatalf("FVOC cue9 descriptor=%+v", bank[9])
	}
	effect, ok := p.ItemUseEffectByRawID(0x45)
	if !ok || effect.EffectID != "clear_condition" || effect.ConditionID != ParalysisCondition ||
		effect.TargetScope != "party_member" || !effect.Consume {
		t.Fatalf("滿月草 effect=%+v ok=%v", effect, ok)
	}
}

func TestBuiltinDQ3DayNightCycleMatchesOriginalEXEAndPalette(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cycle := p.DayNightCycle()
	wantBanks := []int{1, 0, 0, 0, 1, 2, 3, 4, 4, 4, 3, 2}
	if cycle.ClockTicks != 0xf0 || cycle.NightStartTick != 0x78 ||
		cycle.PaletteSegmentTicks != 0x14 || cycle.PaletteEntriesPerBank != 16 ||
		!reflect.DeepEqual(cycle.PaletteBankIndices, wantBanks) || cycle.PaletteAssetKey != "world_palette" {
		t.Fatalf("day-night cycle=%+v", cycle)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	palRaw, err := os.ReadFile(filepath.Join(dir, "DQ3.PAL"))
	if err != nil {
		t.Skipf("original DQ3.PAL unavailable: %v", err)
	}
	for i, raw := range exe[0x18705 : 0x18705+12] {
		if cycle.PaletteBankIndices[i] != int(raw) {
			t.Fatalf("palette selector[%d]=%d, EXE=%d", i, cycle.PaletteBankIndices[i], raw)
		}
	}
	if len(palRaw) != 5*0x30 || fmt.Sprintf("%x", sha256.Sum256(palRaw)) !=
		"178a9ca809a33108d8e2a6430796eae8909d98e514fe543fc1923139110389c4" {
		t.Fatal("DQ3.PAL five-bank contract drift")
	}
}

func TestBuiltinDQ3BattlePackMatchesOriginalRawTables(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b, ok := p.BattlePack()
	if !ok {
		t.Fatal("missing E2 battle pack")
	}
	background, ok := p.BattleBackground()
	if !ok {
		t.Fatal("missing battle background contract")
	}
	if background.ArchiveAsset != "battle_background" || background.PaletteAsset != "battle_palette" ||
		background.PageStrideBytes != 0x13d80 || background.FieldOffsetBytes != 0x6e00 ||
		background.FieldChunkBytes != 0xcf80 || background.WidthPixels != 640 || background.FieldRows != 166 ||
		background.PlaneCount != 4 || background.PageCount != 46 || background.PaletteBankCount != 10 ||
		background.PaletteEntriesPerBank != 16 || background.DefaultSelector != (BattleBackgroundSelector{PageRaw: 0, PaletteBankRaw: 0}) {
		t.Fatalf("battle background=%+v, want original PACKBG/MNSBK contract", background)
	}
	region, candidates, ok := p.BattleEncounterRaw()
	if !ok {
		t.Fatal("battle encounter raw accessor failed")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	if !bytes.Equal(region, exe[0x16140+0x4966:0x16140+0x4a66]) {
		t.Fatal("battle pack region lookup differs from DQ3.EXE")
	}
	if !bytes.Equal(candidates, exe[0x16140+0x4a56:0x16140+0x4a56+0x360]) {
		t.Fatal("battle pack candidate table differs from DQ3.EXE")
	}
	threshold, err := hex.DecodeString(b.Resistance.EffectThresholds.RawHex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(threshold, exe[0x16140+0x36e5:0x16140+0x36e5+20]) {
		t.Fatal("battle pack resistance thresholds differ from DQ3.EXE")
	}
	descriptors, err := hex.DecodeString(b.Resistance.DescriptorTable.RawHex)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(descriptors, exe[0x16140+0x37c3:0x16140+0x37c3+60*3]) {
		t.Fatal("battle pack resistance descriptors differ from DQ3.EXE")
	}
	if len(b.Encounter.FixedRecords) != 13 || len(b.Resistance.EffectNamesZH) != 60 {
		t.Fatalf("battle pack cardinality fixed=%d names=%d", len(b.Encounter.FixedRecords), len(b.Resistance.EffectNamesZH))
	}
	position, ok := p.BattleFormationPosition()
	if !ok || position.OriginRaw != 0x26 || position.FirstMemberOffsetRaw != 2 || position.WeightStepMultiplier != 2 ||
		position.EGAStrideBytes != 0x54 || position.EGABottomRaw != 0x102 ||
		position.PixelsPerRawByte != 8 {
		t.Fatalf("battle formation position=%+v/%v, want origin=0x26 first=2 step=2 stride=0x54 bottom=0x102 px=8", position, ok)
	}
	if position.Evidence.Level != "D2" || position.Evidence.Source != "DQ3.EXE" {
		t.Fatalf("battle formation position evidence=%+v, want D2/DQ3.EXE", position.Evidence)
	}
	for key, want := range map[string]AssetRef{
		"battle_background": {Path: "PACKBG.SCR", Size: 3738880, SHA256: "4e226cd23c38bbbce9db974da559933a2e84669b3abb1559ab9755c85ed800d0"},
		"battle_palette":    {Path: "MNSBK.PAL", Size: 480, SHA256: "e2985503b3eb5256bcaf44f2f1b0a0b1ab124e66873c2de210e0d25ffd75110d"},
	} {
		got, ok := p.Asset(key)
		if !ok || got != want {
			t.Fatalf("battle asset %s=%+v/%v, want %+v", key, got, ok, want)
		}
		raw, err := os.ReadFile(filepath.Join(dir, got.Path))
		if err != nil {
			t.Fatalf("read battle asset %s: %v", key, err)
		}
		if int64(len(raw)) != got.Size || fmt.Sprintf("%x", sha256.Sum256(raw)) != got.SHA256 {
			t.Fatalf("battle asset %s integrity mismatch", key)
		}
	}
	orochi, ok := p.StagedBossEvent("dq3:event.jipang_orochi")
	if !ok {
		t.Fatal("missing Jipang Orochi event")
	}
	for _, tc := range []struct {
		name     string
		fileOff  int
		selector BattleBackgroundSelector
	}{
		{"first", 0x16140 + 0x4ed0, orochi.FirstFormation.Background},
		{"second", 0x16140 + 0x4ed5, orochi.SecondFormation.Background},
	} {
		if tc.fileOff+3 > len(exe) {
			t.Fatalf("%s formation offset %#x outside DQ3.EXE", tc.name, tc.fileOff)
		}
		if tc.selector.PageRaw != int(exe[tc.fileOff+1]) ||
			tc.selector.PaletteBankRaw != int(exe[tc.fileOff+2]) {
			t.Fatalf("%s selector=%+v, raw=%02x%02x", tc.name, tc.selector, exe[tc.fileOff+1], exe[tc.fileOff+2])
		}
	}
	for _, tc := range []struct {
		name    string
		monster int
		fileOff int
	}{
		{"samanosa_boss_troll", 0x59, 0x16140 + 0x4ee4},
		{"baramos", 0x79, 0x16140 + 0x4edf},
		{"baramos_ghost", 0x7a, 0x16140 + 0x4efa},
		{"baramos_zombie", 0x7b, 0x16140 + 0x4eff},
		{"zoma", 0x7c, 0x16140 + 0x4ef0},
	} {
		formation, ok := p.FixedBattleFormationForSingleMonster(tc.monster)
		if !ok {
			t.Fatalf("%s fixed single-monster formation missing", tc.name)
		}
		if tc.fileOff+5 > len(exe) {
			t.Fatalf("%s formation offset %#x outside DQ3.EXE", tc.name, tc.fileOff)
		}
		raw := exe[tc.fileOff : tc.fileOff+5]
		if formation.Background.PageRaw != int(raw[1]) ||
			formation.Background.PaletteBankRaw != int(raw[2]) ||
			len(formation.Groups) != 1 || formation.Groups[0].MonsterRawID != int(raw[3]) ||
			formation.Groups[0].Count != int(raw[4]) {
			t.Fatalf("%s fixed formation=%+v, raw=%x", tc.name, formation, raw)
		}
	}
	if _, ok := p.FixedBattleFormationForSingleMonster(0x4b); ok {
		t.Fatal("two Orochi fixed records must not select an arbitrary background")
	}
}

func TestDQ3RuraNavigationMatchesOriginalEXE(t *testing.T) {
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
		t.Fatal(err)
	}
	destinations := p.RuraDestinations()
	if len(destinations) != 20 {
		t.Fatalf("魯拉目的地數=%d, want 20", len(destinations))
	}
	if len(exe) < 0x16c64 {
		t.Fatalf("DQ3.EXE too small: %d", len(exe))
	}
	ctyRaw := exe[0x16c00:0x16c14]  // DGROUP 0x0ac0
	shipRaw := exe[0x16c14:0x16c64] // DGROUP 0x0ad4
	for i, d := range destinations {
		if d.CTYRaw != int(ctyRaw[i]) || d.NameRecordRaw != 0x1ef+i {
			t.Fatalf("魯拉 index%d CTY/name 不符：%+v rawCTY=%d", i, d, ctyRaw[i])
		}
		x := int(binary.LittleEndian.Uint16(shipRaw[i*4:]))
		rawY := int(binary.LittleEndian.Uint16(shipRaw[i*4+2:]))
		layer, y := 0, rawY
		if rawY >= 300 {
			layer, y = 1, rawY-300
		}
		if d.ShipWorldPosition.X != x || d.ShipWorldPosition.Y != y ||
			d.ShipWorldPosition.Layer != layer {
			t.Fatalf("魯拉 index%d 船落點不符：got %+v raw=(%d,%d)",
				i, d.ShipWorldPosition, x, rawY)
		}
		wantOffsetX := -1
		if i == 10 { // logical 0x4c69..0x4c75：index 10 另加 X+2。
			wantOffsetX = 1
		}
		if d.ArrivalOffset.X != wantOffsetX || d.ArrivalOffset.Y != 0 {
			t.Fatalf("魯拉 index%d arrival offset=%+v want (%d,0)",
				i, d.ArrivalOffset, wantOffsetX)
		}
	}
}

func TestDQ3WorldEntranceVariantsMatchOriginalEXE(t *testing.T) {
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
		t.Fatal(err)
	}
	variants := p.WorldEntranceVariants()
	if len(variants) != 2 {
		t.Fatalf("world entrance variants=%d, want 2", len(variants))
	}
	merchant := variants[0]
	if merchant.WorldTile != (TileCoordinate{X: 210, Y: 64}) || merchant.Layer != 0 ||
		merchant.DefaultCTYRaw != 83 {
		t.Fatalf("merchant entrance header=%+v", merchant)
	}
	wantBranches := []WorldEntranceVariantBranch{
		{FlagRaw: 0x23, WhenFlagSet: false, CTYRaw: 58},
		{FlagRaw: 0x47, WhenFlagSet: true, CTYRaw: 59},
		{FlagRaw: 0x42, WhenFlagSet: true, CTYRaw: 60},
		{FlagRaw: 0x48, WhenFlagSet: true, CTYRaw: 61},
	}
	if !reflect.DeepEqual(merchant.Branches, wantBranches) {
		t.Fatalf("merchant branches=%+v, want %+v", merchant.Branches, wantBranches)
	}
	// cty_loc[58] at DGROUP0x748 + 58*4, file DGROUP base0x16140.
	if len(exe) < 0x16974 || !bytes.Equal(exe[0x16970:0x16974], []byte{0xd2, 0x00, 0x40, 0x00}) {
		t.Fatal("DQ3.EXE cty_loc[58] is not world tile (210,64)")
	}
	// file0x396e function: each mov bx is the tested flag and each mov bp is
	// the selected CTY. The first jz skips CTY58 when flag0x23 is set, proving
	// that branch's polarity differs from the three following set branches.
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0x3988, []byte{0xbb, 0x23, 0x00}},
		{0x3990, []byte{0x74, 0x06}},
		{0x3992, []byte{0xbd, 0x3a, 0x00}},
		{0x3998, []byte{0xbb, 0x47, 0x00}},
		{0x39a2, []byte{0xbd, 0x3b, 0x00}},
		{0x39a8, []byte{0xbb, 0x42, 0x00}},
		{0x39b2, []byte{0xbd, 0x3c, 0x00}},
		{0x39b8, []byte{0xbb, 0x48, 0x00}},
		{0x39c2, []byte{0xbd, 0x3d, 0x00}},
		{0x39c8, []byte{0xbd, 0x53, 0x00}},
	} {
		if len(exe) < anchor.off+len(anchor.raw) ||
			!bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE world entrance anchor mismatch at file %#x", anchor.off)
		}
	}
}

func TestDQ3MerchantSettlementFounderMatchesOriginalBinaryAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY58.DAT"), read("D3TXT07.TXT")
	txt00 := read("D3TXT00.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	event, ok := p.SettlementFounderEvent("dq3:event.merchant_settlement_founder")
	if !ok {
		t.Fatal("missing merchant settlement founder event")
	}
	wantNPC := ScriptedNPCSelector{CTYRaw: 58, Section: 0,
		Tile: TileCoordinate{X: 7, Y: 15}, HandlerRaw: 40}
	if event.NPC != wantNPC || event.RequiredClassRaw != 6 || !event.RequireAlive ||
		event.CompletionFlagRaw != 0x23 ||
		!reflect.DeepEqual(event.FounderVisibilityFlagsRaw, []int{0x15, 0x23}) ||
		event.SharedStorageCapacity != 0x78 {
		t.Fatalf("merchant founder pack fields drifted: %+v", event)
	}
	sec := int(binary.LittleEndian.Uint16(cty[0:2]))
	npcTable := sec + int(binary.LittleEndian.Uint16(cty[sec:sec+2]))
	found := false
	for i := 0; i < int(cty[npcTable]); i++ {
		rec := cty[npcTable+1+i*7 : npcTable+1+(i+1)*7]
		if int(rec[0]) == wantNPC.Tile.X && int(rec[1]) == wantNPC.Tile.Y &&
			(rec[3]>>3)&7 == 2 && int(rec[4]) == wantNPC.HandlerRaw {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("CTY58 sec0 does not contain scripted NPC (7,15) handler40")
	}

	// Non-destructive IDA 9.4 trace, file0x6e12..0x6f29: flag0x23
	// repeat branch, class6 party scan, dead-status gate, two choices, storage
	// loop, 18-byte founder copy and active-party removal.
	for _, anchor := range []struct {
		off int
		raw []byte
	}{
		{0x6e12, []byte{0xe8, 0x5d, 0xf5, 0xbb, 0x23, 0x00, 0xe8, 0x5e, 0x14, 0x3c, 0x01}},
		{0x6e32, []byte{0x8a, 0x0e, 0x77, 0x50, 0x32, 0xed, 0xbb, 0x00, 0x00, 0x8b, 0xb7, 0x15, 0x4f, 0x80, 0x7c, 0x01, 0x06}},
		{0x6e52, []byte{0x24, 0xd1, 0xeb, 0x43, 0x89, 0x1e, 0x9c, 0x25, 0xf7, 0x44, 0x38, 0x80, 0x00}},
		{0x6ea0, []byte{0xbb, 0x23, 0x00, 0xe8, 0xa9, 0x13}},
		{0x6ec0, []byte{0xc6, 0x87, 0x64, 0x4f, 0x00, 0x56, 0x83, 0xc6, 0x3a, 0xb9, 0x08, 0x00}},
		{0x6ede, []byte{0x5e, 0x8d, 0x3e, 0x58, 0x50, 0xb9, 0x12, 0x00}},
	} {
		if anchor.off+len(anchor.raw) > len(exe) ||
			!bytes.Equal(exe[anchor.off:anchor.off+len(anchor.raw)], anchor.raw) {
			t.Fatalf("DQ3.EXE merchant handler raw anchor mismatch at file %#x", anchor.off)
		}
	}

	refs := event.DialogueTextIDs
	for id, record := range map[string]int{
		refs.Introduction: 0, refs.NoEligible: 1, refs.FirstOffer: 2,
		refs.Reject: 3, refs.FinalOffer: 4, refs.Accepted: 5, refs.After: 6,
	} {
		start := int(binary.LittleEndian.Uint16(txt[record*2:]))
		end := int(binary.LittleEndian.Uint16(txt[(record+1)*2:]))
		want := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			want = append(want, binary.LittleEndian.Uint16(txt[off:off+2]))
		}
		got, exists := p.TextGlyphCodes(id)
		if !exists || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s does not exactly match D3TXT07 record%d: got=%v want=%v",
				id, record, got, want)
		}
	}
	record := 309
	start := int(binary.LittleEndian.Uint16(txt00[record*2:]))
	end := int(binary.LittleEndian.Uint16(txt00[(record+1)*2:]))
	want := make([]uint16, 0, (end-start)/2)
	for off := start; off+2 <= end; off += 2 {
		want = append(want, binary.LittleEndian.Uint16(txt00[off:off+2]))
	}
	got, exists := p.TextGlyphCodes(event.DialogueTextIDs.StorageFull)
	if !exists || !reflect.DeepEqual(got, want) {
		t.Fatalf("%s does not exactly match D3TXT00 record309: got=%v want=%v",
			event.DialogueTextIDs.StorageFull, got, want)
	}
}

func TestDQ3MerchantSettlementConditionalShopMatchesOriginalEXEAndText(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT07.TXT"))
	if err != nil {
		t.Skipf("original D3TXT07.TXT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	event, ok := p.ConditionalFacilityEvent("dq3:event.merchant_settlement_shop_gate")
	if !ok {
		t.Fatal("missing CTY59 conditional shop event")
	}
	wantNPC := ScriptedNPCSelector{CTYRaw: 59, Section: 0,
		Tile: TileCoordinate{X: 10, Y: 2}, HandlerRaw: 41}
	if event.Kind != "conditional_facility" || event.NPC != wantNPC ||
		event.RequiredPlayerY != 4 || event.FacilityK != 0 {
		t.Fatalf("conditional shop event drifted: %+v", event)
	}
	// sub_15BBA (IDA linear 0x15bba; file 0x6f2a) branches on DS:4f35==4, then calls
	// the generic facility dispatcher with BX=0; otherwise it selects di=0xbc0.
	const handlerFile = 0x6f2a
	if len(exe) < handlerFile+7 ||
		!bytes.Equal(exe[handlerFile:handlerFile+7], []byte{0x83, 0x3e, 0x35, 0x4f, 0x04, 0x74, 0x0e}) {
		t.Fatalf("handler41 Y=4 gate bytes changed: %x", exe[handlerFile:handlerFile+7])
	}
	start := int(binary.LittleEndian.Uint16(txt[8*2:]))
	end := int(binary.LittleEndian.Uint16(txt[9*2:]))
	want := make([]uint16, 0, (end-start)/2)
	for off := start; off+2 <= end; off += 2 {
		want = append(want, binary.LittleEndian.Uint16(txt[off:off+2]))
	}
	got, exists := p.TextGlyphCodes(event.FallbackTextID)
	if !exists || !reflect.DeepEqual(got, want) {
		t.Fatalf("%s does not exactly match D3TXT07 record8: got=%v want=%v",
			event.FallbackTextID, got, want)
	}
}

func TestDQ3JipangOrochiMatchesOriginalEXEAndText(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("original %s unavailable: %v", name, err)
		}
		return raw
	}
	exe, txt := read("DQ3.EXE"), read("D3TXT04.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	e, ok := p.StagedBossEvent("dq3:event.jipang_orochi")
	if !ok {
		t.Fatal("dq3:event.jipang_orochi missing")
	}
	if e.FirstNPC != (ScriptedNPCSelector{CTYRaw: 19, Section: 1, Tile: TileCoordinate{X: 35, Y: 12}, HandlerRaw: 45}) ||
		e.SecondNPC != (ScriptedNPCSelector{CTYRaw: 21, Section: 0, Tile: TileCoordinate{X: 18, Y: 7}, HandlerRaw: 36}) {
		t.Fatalf("Orochi NPC selectors mismatch: first=%+v second=%+v", e.FirstNPC, e.SecondNPC)
	}
	for _, anchor := range []struct {
		off  int
		want string
	}{
		{0x1b010, e.FirstFormation.RawBytesHex},
		{0x1b015, e.SecondFormation.RawBytesHex},
		{0x19e2e, "010300010301ff"},
	} {
		want, err := hex.DecodeString(anchor.want)
		if err != nil {
			t.Fatal(err)
		}
		if anchor.off+len(want) > len(exe) || !bytes.Equal(exe[anchor.off:anchor.off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x mismatch: got=%x want=%x", anchor.off,
				exe[anchor.off:anchor.off+len(want)], want)
		}
	}
	if !reflect.DeepEqual(e.FirstNPCPostBattleDirections, []string{"right", "right"}) ||
		!reflect.DeepEqual(e.FirstPostBattleDirections, []string{"right", "right"}) {
		t.Fatalf("scripted movement was not preserved: npc=%v player=%v",
			e.FirstNPCPostBattleDirections, e.FirstPostBattleDirections)
	}
	records := map[string]int{
		e.DialogueTextIDs.SecondOffer: 66, e.DialogueTextIDs.SecondAccept: 67,
		e.DialogueTextIDs.SecondChallenge: 68, e.DialogueTextIDs.SecondVictory: 69,
		e.DialogueTextIDs.FirstPost: 70,
	}
	for id, record := range records {
		start := int(binary.LittleEndian.Uint16(txt[record*2:]))
		end := int(binary.LittleEndian.Uint16(txt[(record+1)*2:]))
		var want []uint16
		for off := start; off+2 <= end; off += 2 {
			want = append(want, binary.LittleEndian.Uint16(txt[off:off+2]))
		}
		got, found := p.TextGlyphCodes(id)
		if !found || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s does not match D3TXT04 rec%d", id, record)
		}
	}
}

func TestDQ3TeidonDarkLampMatchesOriginalEXEAndCTY(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("original %s unavailable: %v", name, err)
		}
		return raw
	}
	exe, cty := read("DQ3.EXE"), read("CTY20.DAT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	treasure, ok := p.TreasureEvent("dq3:event.teidon_dark_lamp")
	if !ok {
		t.Fatal("dq3:event.teidon_dark_lamp missing")
	}
	use, ok := p.ItemUseEffectByRawID(treasure.Treasure.ItemRawID)
	if !ok || use.ID != "dq3:item_use.dark_lamp" {
		t.Fatalf("黑暗燈寶箱未接到 item-use effect：%+v", use)
	}
	if treasure.Treasure != (QuestTreasureSelector{
		CTYRaw: 20, Section: 1, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x5f, PresentFlag: 0x39,
	}) {
		t.Fatalf("提頓黑暗燈 treasure JSON 不符：%+v", treasure.Treasure)
	}
	if use.ItemRawID != 0x5f || use.EffectID != "force_day_night_phase" ||
		use.LocationKind != "overworld" || use.DayNightPhase != 2 ||
		use.DayNightClock == nil || *use.DayNightClock != 0x8c ||
		!use.ResetDayNightSteps || use.Consume {
		t.Fatalf("黑暗燈 item-use JSON 不符：%+v", use)
	}

	section := int(binary.LittleEndian.Uint16(cty[2:4]))
	eventTable := section + int(binary.LittleEndian.Uint16(cty[section+8:section+10]))
	if eventTable+5 > len(cty) || cty[eventTable] < 1 {
		t.Fatalf("CTY20 sec1 event table 無效：section=%#x table=%#x", section, eventTable)
	}
	if raw := cty[eventTable+1 : eventTable+5]; !reflect.DeepEqual(raw,
		[]byte{0x01, 0x5f, 0x00, 0x39}) {
		t.Fatalf("CTY20 sec1 subid0=%x，want 01 5f 00 39", raw)
	}

	for off, want := range map[int][]byte{
		// selected record is decremented to raw item ID; IDs >=0x41 index the
		// DGROUP 0x366a near-handler table.
		0x503f:  {0x4b, 0x83, 0xfb, 0x41, 0x7d, 0x2b},
		0x506f:  {0x90, 0x83, 0xeb, 0x41, 0xd1, 0xe3, 0x83, 0xbf, 0x6a, 0x36},
		0x197e6: {0x63, 0x40}, // raw item 0x5f -> logical handler 0x4063
		// overworld gate, day-byte gate, write night, then clock target 0x008c.
		0x53d3: {0x83, 0x3e, 0x2d, 0x4f, 0x00, 0x74, 0x03},
		0x53dd: {0x80, 0x3e, 0x6c, 0x52, 0x01, 0x74, 0x03},
		0x53e7: {0xc6, 0x06, 0x6c, 0x52, 0x00},
		0x53ed: {0x81, 0x3e, 0x1d, 0x25, 0x8c, 0x00},
		0x5413: {0x1d, 0x25, 0x8c, 0x00},
	} {
		if off+len(want) > len(exe) ||
			!reflect.DeepEqual(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x 不符：got %x want %x",
				off, exe[off:off+len(want)], want)
		}
	}
}

func TestDQ3MerchantTownYellowOrbMatchesOriginalCTY(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY83.DAT"))
	if err != nil {
		t.Skipf("original CTY83.DAT unavailable: %v", err)
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT07.TXT"))
	if err != nil {
		t.Skipf("original D3TXT07.TXT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	event, ok := p.TreasureEvent("dq3:event.merchant_town_yellow_orb")
	if !ok {
		t.Fatal("缺 dq3:event.merchant_town_yellow_orb")
	}
	if event.Treasure != (QuestTreasureSelector{
		CTYRaw: 83, Section: 0, TileSubID: 2, EventTypeRaw: 1,
		ItemRawID: 0x6a, PresentFlag: 0x4a,
	}) {
		t.Fatalf("商人城黃寶珠 treasure JSON 不符：%+v", event.Treasure)
	}
	section := int(binary.LittleEndian.Uint16(cty[0:2]))
	eventTable := section + int(binary.LittleEndian.Uint16(cty[section+8:section+10]))
	off := eventTable + 1 + event.Treasure.TileSubID*4
	if off+4 > len(cty) || cty[eventTable] <= byte(event.Treasure.TileSubID) {
		t.Fatalf("CTY83 sec0 event table 無效：section=%#x table=%#x", section, eventTable)
	}
	if raw := cty[off : off+4]; !reflect.DeepEqual(raw, []byte{0x01, 0x6a, 0x00, 0x4a}) {
		t.Fatalf("CTY83 sec0 subid2=%x，want 01 6a 00 4a", raw)
	}
	if raw := exe[0x70c0:0x70e0]; !reflect.DeepEqual(raw, []byte{
		0xe8, 0xaf, 0xf2, 0xbf, 0xe2, 0x0b, 0x9a, 0x64,
		0x02, 0x1b, 0x11, 0xbb, 0x4a, 0x00, 0xe8, 0xa8,
		0x11, 0x3c, 0x01, 0x75, 0x08, 0xbf, 0xe3, 0x0b,
		0x9a, 0x64, 0x02, 0x1b, 0x11, 0xe9, 0xa0, 0xf2,
	}) {
		t.Fatalf("DQ3.EXE handler48 file0x70c0 raw bytes drifted: %x", raw)
	}
	followups := p.SettlementFounderFollowups()
	if len(followups) != 2 {
		t.Fatalf("CTY83 founder followup count=%d, want 2", len(followups))
	}
	for id, record := range map[string]int{
		"dq3:text.merchant_settlement.imprisoned": 42,
		"dq3:text.merchant_settlement.seat_hint":  43,
	} {
		start := int(binary.LittleEndian.Uint16(txt[record*2:]))
		end := int(binary.LittleEndian.Uint16(txt[(record+1)*2:]))
		want := make([]uint16, 0, (end-start)/2)
		for pos := start; pos+2 <= end; pos += 2 {
			want = append(want, binary.LittleEndian.Uint16(txt[pos:pos+2]))
		}
		got, exists := p.TextGlyphCodes(id)
		if !exists || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s 與 D3TXT07 record%d 不符：got=%v want=%v", id, record, got, want)
		}
	}
}

func TestDQ3InvisibilityGrassMatchesOriginalEXE(t *testing.T) {
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
	use, ok := p.ItemUseEffectByRawID(0x5d)
	if !ok || use.ID != "dq3:item_use.invisibility_grass" {
		t.Fatalf("隱身草 item-use effect missing: %+v", use)
	}
	if use.EffectID != "temporary_invisibility" || use.LocationKind != "any" ||
		use.StepCount != 0x19 || !use.Consume || use.ResetDayNightSteps {
		t.Fatalf("隱身草 item-use JSON 不符：%+v", use)
	}

	// DGROUP 0x366a entry (0x5d-0x41) points to logical handler 0x3fff.
	if got := binary.LittleEndian.Uint16(exe[0x197e2:0x197e4]); got != 0x3fff {
		t.Fatalf("item 0x5d handler=%#x，want logical 0x3fff", got)
	}
	// file 0x536f: consume selected slot, then timer=0x19 and world flags|=0xc000.
	want := []byte{0xe8, 0xf7, 0x0c, 0xe8, 0x3f, 0xa6, 0xc6, 0x06, 0xf7,
		0x52, 0x19, 0x90, 0x81, 0x0e, 0x46, 0x4f, 0x00, 0xc0, 0xc3}
	if got := exe[0x536f : 0x536f+len(want)]; !reflect.DeepEqual(got, want) {
		t.Fatalf("item 0x5d handler bytes=%x，want %x", got, want)
	}
}

func TestDQ3EginbearPushPuzzleMatchesOriginalEXEAndCTY(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY76.DAT"))
	if err != nil {
		t.Skipf("original CTY76.DAT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	events := p.PushPuzzleEvents()
	if len(events) != 1 {
		t.Fatalf("push_puzzle_events=%d，want 1", len(events))
	}
	e := events[0]
	pushRule := p.NPCPushRule()
	if e.CTYRaw != 76 || e.Section != 0 || e.CompletionAxis != "y" ||
		e.CompletionValue != 5 || e.ClearStoryFlagRaw != 0x3b ||
		len(e.NPCs) != 3 || len(e.TriggerTiles) != 8 ||
		pushRule.CtrlMask != 0x40 || pushRule.CtrlValue != 0x40 ||
		pushRule.BlockedByEffectID != "temporary_invisibility" {
		t.Fatalf("CTY76 push puzzle JSON 不符：%+v", e)
	}
	treasure, ok := p.TreasureEvent("dq3:event.eginbear_thirsty_pitcher")
	if !ok || treasure.Treasure.CTYRaw != 76 || treasure.Treasure.Section != 0 ||
		treasure.Treasure.TileSubID != 0 || treasure.Treasure.EventTypeRaw != 1 ||
		treasure.Treasure.ItemRawID != 0x5e || treasure.Treasure.PresentFlag != 0x3a {
		t.Fatalf("CTY76 乾渴壺 treasure JSON 不符：%+v", treasure)
	}
	town, err := dq3data.OpenTown(cty, 0, false)
	if err != nil {
		t.Fatalf("OpenTown CTY76: %v", err)
	}
	if !reflect.DeepEqual(town.SpecialHandlers, []int{30}) ||
		!reflect.DeepEqual(town.Events, [][3]int{{1, 0x5e, 0x3a}, {1, 0x5e, 0x3b}}) ||
		len(town.NPCs) != 3 {
		t.Fatalf("CTY76 原始 tables 不符：handlers=%v events=%v npcs=%+v",
			town.SpecialHandlers, town.Events, town.NPCs)
	}
	for i, want := range [][2]int{{3, 10}, {5, 10}, {7, 10}} {
		npc := town.NPCs[i]
		selector := e.NPCs[i]
		if selector.Slot != i || selector.InitialTile.X != want[0] ||
			selector.InitialTile.Y != want[1] || selector.CtrlRaw != 0xc0 ||
			npc.X != want[0] || npc.Y != want[1] ||
			npc.Ctrl != 0xc0 {
			t.Fatalf("CTY76 slot%d anchor 不符：json=%+v raw=%+v", i, selector, npc)
		}
	}
	for _, trigger := range e.TriggerTiles {
		i := trigger.Tile.Y*town.W + trigger.Tile.X
		if trigger.TileSubID != 1 || trigger.HandlerRaw != 30 || town.HiMap[i]&0x1f != 1 {
			t.Fatalf("CTY76 trigger 不符：%+v hi=%02x", trigger, town.HiMap[i])
		}
	}

	// logical 0x586b / file 0x6bdb: test flag0x3b, require the first
	// three runtime NPC records' Y byte to equal 5, then clear flag0x3b.
	handler := []byte{0xbb, 0x3b, 0x00, 0xe8, 0x98, 0x16, 0x3c, 0x01, 0x74, 0x01, 0xc3,
		0xb9, 0x03, 0x00, 0x8d, 0x3e, 0x66, 0x0b, 0x80, 0x7d, 0x01, 0x05, 0x75, 0xf2,
		0x83, 0xc7, 0x08, 0xe2, 0xf5, 0xbb, 0x3b, 0x00, 0xe8, 0x66, 0x16, 0xe8, 0x47,
		0xdb, 0xbd, 0x09, 0x00, 0x9a, 0x40, 0x02, 0x53, 0x10, 0x9a, 0xb2, 0x03, 0x53,
		0x10, 0xc3}
	if got := exe[0x6bdb : 0x6bdb+len(handler)]; !reflect.DeepEqual(got, handler) {
		t.Fatalf("handler30 bytes=%x，want %x", got, handler)
	}
	// logical 0x1de2a / file 0xf19a: invisibility bits short-circuit,
	// then the collided runtime NPC must have ctrl bit0x40 before pushing.
	pushAnchor := []byte{0xf7, 0x06, 0x46, 0x4f, 0x00, 0xc0, 0x74, 0x03, 0xe9, 0x89, 0x00,
		0x8a, 0xdd, 0x80, 0xe3, 0x1f, 0x32, 0xff, 0x88, 0x2e, 0x7c, 0x06, 0xd1, 0xe3,
		0xd1, 0xe3, 0xd1, 0xe3, 0x8d, 0x3e, 0x66, 0x0b, 0x03, 0xfb, 0xb0, 0x40,
		0x84, 0x45, 0x03}
	if got := exe[0xf19a : 0xf19a+len(pushAnchor)]; !reflect.DeepEqual(got, pushAnchor) {
		t.Fatalf("push consumer bytes=%x，want %x", got, pushAnchor)
	}
}

func TestDQ3ThirstyPitcherAndFinalKeyMatchOriginalEXEAndCTY(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY40.DAT"))
	if err != nil {
		t.Skipf("original CTY40.DAT unavailable: %v", err)
	}
	attrRaw, err := os.ReadFile(filepath.Join(dir, "BLKBM1.DAT"))
	if err != nil {
		t.Skipf("original BLKBM1.DAT unavailable: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	effect, ok := p.ItemUseEffectByRawID(0x5e)
	if !ok || effect.EffectID != "reveal_world_map_patch" || effect.UseTile == nil ||
		effect.RequiredLayer != 0 || effect.RequiredVehicle != "ship" ||
		*effect.UseTile != (TileCoordinate{X: 0x92, Y: 0x35}) ||
		effect.RequiredStoryFlagRaw != 0x12 || effect.RequiredStoryFlagSet ||
		effect.SetWorldStateMask != 0x08 || effect.Consume ||
		effect.AnimationPaletteModeRaw != 4 || effect.AnimationCycles != 4 {
		t.Fatalf("乾渴壺 item-use JSON 不符：%+v", effect)
	}
	patch := effect.MapPatch
	wantTiles := []int{
		0x58, 0x58, 0x60, 0x58, 0x58,
		0x58, 0x68, 0x42, 0x66, 0x58,
		0x5e, 0x42, 0x7c, 0x42, 0x5c,
		0x58, 0x5a, 0x5a, 0x5a, 0x58,
	}
	if patch == nil || patch.Layer != 0 ||
		patch.Origin != (TileCoordinate{X: 0x90, Y: 0x32}) ||
		patch.Width != 5 || patch.Height != 4 || !reflect.DeepEqual(patch.TilesRaw, wantTiles) ||
		!reflect.DeepEqual(effect.ClearStoryFlagsRaw, []int{0x12}) {
		t.Fatalf("乾渴壺 world patch JSON 不符：%+v", patch)
	}
	if got := binary.LittleEndian.Uint16(exe[0x197e4:0x197e6]); got != 0x4012 {
		t.Fatalf("item0x5e handler=%#x，want logical 0x4012", got)
	}
	wantHandler, _ := hex.DecodeString(
		"bb1200e8f12e3c017416803e3b4f01750f813e2f4f92007507833e314f357403" +
			"e9d202e87d0a830e444f08c706d3250400b90400e84caf8d36543abf9000b832" +
			"00e8db0cbb1200e8982ee8b2e4e826a5c3")
	if got := exe[0x5382 : 0x5382+len(wantHandler)]; !reflect.DeepEqual(got, wantHandler) {
		t.Fatalf("乾渴壺 handler bytes=%x，want %x", got, wantHandler)
	}
	wantPatch := make([]byte, len(wantTiles))
	for i, tile := range wantTiles {
		wantPatch[i] = byte(tile)
	}
	if got := exe[0x19b94 : 0x19b94+len(wantPatch)]; !reflect.DeepEqual(got, wantPatch) {
		t.Fatalf("乾渴壺 map patch bytes=%x，want %x", got, wantPatch)
	}

	town, err := dq3data.OpenTown(cty, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(town.Events, [][3]int{{1, 0x57, 0x47}}) {
		t.Fatalf("CTY40 event table=%v，want [[1 0x57 0x47]]", town.Events)
	}
	i := 17*town.W + 18
	attr := dq3data.OpenBlockAttr(attrRaw)
	if i < 0 || i >= len(town.HiMap) || town.HiMap[i]&0x1f != 0 ||
		attr.Raw(int(town.Cells[i]))&0x0008 == 0 {
		t.Fatalf("CTY40 最終鑰匙 tile(18,17)：subid=%d tile=%#x attr=%#x",
			town.HiMap[i]&0x1f, town.Cells[i], attr.Raw(int(town.Cells[i])))
	}
	treasure, ok := p.TreasureEvent("dq3:event.shoal_final_key")
	if !ok || treasure.Treasure != (QuestTreasureSelector{
		CTYRaw: 40, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x57, PresentFlag: 0x47,
	}) {
		t.Fatalf("CTY40 最終鑰匙 treasure JSON 不符：%+v", treasure)
	}
}

func TestDQ3DoorKeysAndTeidonGreenOrbMatchOriginal(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("original %s unavailable: %v", name, err)
		}
		return raw
	}
	exe, cty := read("DQ3.EXE"), read("CTY20.DAT")
	attrRaw, fon, txt := read("BLKBM1.DAT"), read("D3TXT00.FON"), read("D3TXT04.TXT")
	pack, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}

	for i, rawID := range []int{0x55, 0x56, 0x57} {
		effect, ok := pack.ItemUseEffectByRawID(rawID)
		if !ok || effect.EffectID != "open_facing_locked_door" ||
			effect.LocationKind != "town" || effect.DoorKeyTier != i+1 || effect.Consume {
			t.Fatalf("key %#x pack effect mismatch: %+v", rawID, effect)
		}
		wordOff := 0x16140 + 0x366a + (rawID-0x41)*2
		if got, want := binary.LittleEndian.Uint16(exe[wordOff:wordOff+2]), uint16(0x3f65+i*9); got != want {
			t.Fatalf("key %#x handler=%#x, want %#x", rawID, got, want)
		}
	}
	wantKeyStubs := []byte{
		0xc7, 0x06, 0x93, 0x25, 0x01, 0x00, 0xeb, 0x10, 0x90,
		0xc7, 0x06, 0x93, 0x25, 0x02, 0x00, 0xeb, 0x07, 0x90,
		0xc7, 0x06, 0x93, 0x25, 0x03, 0x00, 0xe8, 0x86, 0x09, 0xc3,
	}
	if got := exe[0x52d5 : 0x52d5+len(wantKeyStubs)]; !bytes.Equal(got, wantKeyStubs) {
		t.Fatalf("key handler stubs mismatch: %x", got)
	}

	var reward *NPCItemRewardEvent
	for _, event := range pack.NPCItemRewardEvents() {
		if event.ID == "dq3:event.teidon_green_orb" {
			e := event
			reward = &e
			break
		}
	}
	if reward == nil || reward.NPC != (ScriptedNPCSelector{
		CTYRaw: 20, Section: 0, Tile: TileCoordinate{X: 16, Y: 2}, HandlerRaw: 35,
	}) || reward.PresentFlagRaw != 0x3e || reward.GrantedItemRaw != 0x66 {
		t.Fatalf("teidon reward mismatch: %+v", reward)
	}
	if got := binary.LittleEndian.Uint16(exe[0x16140+0x3bb4+35*2:]); got != 0x593a {
		t.Fatalf("scripted handler table entry35=%#x, want 0x593a", got)
	}
	if got := exe[0x6cad : 0x6cad+3]; !bytes.Equal(got, []byte{0xbb, 0x3e, 0x00}) {
		t.Fatalf("handler35 present-flag selector bytes=%x", got)
	}
	if got := exe[0x6cca : 0x6cca+6]; !bytes.Equal(got, []byte{0xc7, 0x06, 0x93, 0x25, 0x66, 0x00}) {
		t.Fatalf("handler35 granted-item writer bytes=%x", got)
	}
	if got := exe[0x7be8 : 0x7be8+3]; !bytes.Equal(got, []byte{0x80, 0xfa, 0x08}) {
		t.Fatalf("party inventory eight-slot consumer bytes=%x", got)
	}

	day, err := dq3data.OpenTown(cty, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	night, err := dq3data.OpenTown(cty, 0, true)
	if err != nil {
		t.Fatal(err)
	}
	hasNPC := func(town *dq3data.Town, x, y, subtype, handler int) bool {
		for _, npc := range town.NPCs {
			if npc.X == x && npc.Y == y && int((npc.Ctrl>>3)&7) == subtype && int(npc.B4) == handler {
				return true
			}
		}
		return false
	}
	if !hasNPC(day, 16, 2, 0, 29) || !hasNPC(night, 16, 2, 2, 35) {
		t.Fatal("CTY20 day corpse/night scripted reward NPC variants mismatch")
	}
	i := 4*night.W + 17
	attr := dq3data.OpenBlockAttr(attrRaw).Raw(int(night.Cells[i]))
	if tier := (attr >> 6) & 3; tier != 3 {
		t.Fatalf("CTY20 locked door (17,4) tier=%d attr=%#x", tier, attr)
	}
	decoded := dq3data.LoadText(fon, txt)
	for rec, textID := range map[int]string{
		38: reward.DialogueTextIDs.Success,
		39: reward.DialogueTextIDs.After,
	} {
		want, ok := pack.TextGlyphCodes(textID)
		got := append(decoded.Record(rec), 0xffff)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("D3TXT04 rec%d != %s: got=%v want=%v", rec, textID, got, want)
		}
	}
}

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

func TestDQ3GreenlandTransformStaffExchangeMatchesOriginalEXECTYAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY54.DAT"), read("D3TXT06.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	e, ok := p.ChoiceItemExchangeEvent("dq3:event.greenland_transform_staff_exchange")
	if !ok {
		t.Fatal("dq3:event.greenland_transform_staff_exchange missing")
	}
	if e.NPC != (ScriptedNPCSelector{CTYRaw: 54, Section: 0,
		Tile: TileCoordinate{X: 8, Y: 2}, HandlerRaw: 44}) ||
		e.AvailableFlagRaw != 0x43 || e.RequiredItemRawID != 0x62 ||
		e.GrantedItemRawID != 0x63 ||
		e.ActivateWorldObject != (WorldObjectActivation{ObjectID: "dq3:world_object.ghost_ship",
			Position: VehicleWorldPosition{X: 150, Y: 90, Layer: 0}}) ||
		e.SetWorldStateMaskRaw != 4 || !reflect.DeepEqual(e.ClearStoryFlagsRaw, []int{0x43}) {
		t.Fatalf("格陵蘭交換 JSON 不符原版：%+v", e)
	}

	sec := int(binary.LittleEndian.Uint16(cty[e.NPC.Section*2:]))
	npcOff := sec + int(binary.LittleEndian.Uint16(cty[sec:]))
	found := false
	for i := 0; i < int(cty[npcOff]); i++ {
		rec := cty[npcOff+1+i*7 : npcOff+1+(i+1)*7]
		if int(rec[0]) == e.NPC.Tile.X && int(rec[1]) == e.NPC.Tile.Y &&
			(rec[3]>>3)&7 == 2 && int(rec[4]) == e.NPC.HandlerRaw {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("CTY54 找不到 (8,2) sub2 handler44")
	}

	for off, want := range map[int][]byte{
		0x6f93: {0xbb, 0x43, 0x00},
		0x6fa8: {0xc7, 0x06, 0x93, 0x25, 0x62, 0x00},
		0x6ffb: {0xc7, 0x04, 0x63, 0x00},
		0x6fff: {0xc7, 0x06, 0x53, 0x50, 0x96, 0x00},
		0x7005: {0xc7, 0x06, 0x55, 0x50, 0x5a, 0x00},
		0x700b: {0x83, 0x0e, 0x44, 0x4f, 0x04},
		0x7013: {0xbb, 0x43, 0x00},
	} {
		if off+len(want) > len(exe) || !bytes.Equal(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x bytes=%x, want %x", off, exe[off:off+len(want)], want)
		}
	}

	records := map[string]int{
		e.DialogueTextIDs.Introduction: 56, e.DialogueTextIDs.Interested: 57,
		e.DialogueTextIDs.Hint: 58, e.DialogueTextIDs.Offer: 59,
		e.DialogueTextIDs.Success: 60, e.DialogueTextIDs.After: 61,
		e.DialogueTextIDs.Reject: 62,
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
			t.Fatalf("%s 未與 D3TXT06 rec%d 完整一致", id, record)
		}
	}
}

func TestDQ3GhostShipMatchesOriginalEXECTYAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY36.DAT"), read("D3TXT00.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	obj, ok := p.TrackedWorldObject("dq3:world_object.ghost_ship")
	if !ok {
		t.Fatal("dq3:world_object.ghost_ship missing")
	}
	if obj.ActiveWorldStateMask != 4 || obj.Layer != 0 || obj.EntranceCTYRaw != 36 ||
		obj.EntranceSection != 0 || obj.EntryVehicleMode != "disembark_ship_if_aboard" ||
		obj.TrackerItemRawID != 0x63 || !reflect.DeepEqual(obj.TileRawByRNGLow2, []int{0x33, 0x35, 0x37, 0x39}) {
		t.Fatalf("幽靈船 JSON 固定欄位不符原版：%+v", obj)
	}

	wantCandidates := exe[0x19b4c : 0x19b4c+18*4]
	gotCandidates := make([]byte, 0, len(wantCandidates))
	for _, candidate := range obj.Relocation.Candidates {
		gotCandidates = binary.LittleEndian.AppendUint16(gotCandidates, uint16(candidate.X))
		gotCandidates = binary.LittleEndian.AppendUint16(gotCandidates, uint16(candidate.Y))
		if candidate.Layer != 0 {
			t.Fatalf("幽靈船候選點 layer=%d, want 0", candidate.Layer)
		}
	}
	if !bytes.Equal(gotCandidates, wantCandidates) {
		t.Fatalf("幽靈船 relocation candidates=%x, want DQ3.EXE file0x19b4c=%x", gotCandidates, wantCandidates)
	}
	itemJumpOff := 0x19728 + obj.TrackerItemRawID*2
	if got := binary.LittleEndian.Uint16(exe[itemJumpOff:]); got != 0x413f {
		t.Fatalf("item0x63 jump=%#x, want logical 0x413f", got)
	}

	eventOff := 0x13cb
	if got, want := cty[eventOff:eventOff+4], []byte{0x01, 0x64, 0x00, 0x8b}; !bytes.Equal(got, want) {
		t.Fatalf("CTY36 sec1 event0=%x, want %x", got, want)
	}

	records := map[string]int{
		obj.TrackerDialogueTextIDs.Preamble: 740,
		obj.TrackerDialogueTextIDs.West:     741,
		obj.TrackerDialogueTextIDs.East:     742,
		obj.TrackerDialogueTextIDs.North:    743,
		obj.TrackerDialogueTextIDs.South:    744,
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
			t.Fatalf("%s 未與 D3TXT00 rec%d 完整一致：got=%#v raw=%#v", id, record, got, raw)
		}
	}
}

func TestDQ3OliviaCapeAndGaiaSwordMatchOriginalEXECTYAndText(t *testing.T) {
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
	exe, cty, txt := read("DQ3.EXE"), read("CTY55.DAT"), read("D3TXT00.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	event, ok := p.CoordinateItemGateAt(76, 54, 0)
	if !ok {
		t.Fatal("奧莉薇亞海岬 coordinate_item_gate missing")
	}
	if event.ActiveStoryFlagRaw != 0x35 || event.RequiredItemRawID != 0x64 ||
		event.ClearStoryFlagRaw != 0x35 || event.ConsumeRequiredItem ||
		event.FailureDirection != 3 || event.FailureSteps != 5 ||
		event.FailureDayNightSteps != 20 {
		t.Fatalf("奧莉薇亞海岬 JSON 不符原版：%+v", event)
	}
	for off, want := range map[int][]byte{
		0x3a2d: {0x83, 0x3e, 0x2f, 0x4f, 0x4c},
		0x3a34: {0x83, 0x3e, 0x31, 0x4f, 0x36},
		0x3a3b: {0xbb, 0x35, 0x00, 0xe8, 0x38, 0x48},
		0x3a45: {0xc7, 0x06, 0x93, 0x25, 0x64, 0x00},
		0x3a58: {0xbf, 0x55, 0x02},
		0x3a60: {0xbf, 0x56, 0x02},
		0x3a6b: {0xbb, 0x35, 0x00, 0xe8, 0xf3, 0x47},
		0x3a7a: {0x80, 0x06, 0xf4, 0x52, 0x14},
		0x3a7f: {0xb9, 0x05, 0x00},
		0x3a89: {0xc7, 0x06, 0x1f, 0x4f, 0x03, 0x00},
		0x3a8f: {0xe8, 0xa1, 0x6d},
	} {
		if off+len(want) > len(exe) || !bytes.Equal(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x bytes=%x, want %x", off, exe[off:off+len(want)], want)
		}
	}
	if got, want := cty[0x56:0x5a], []byte{0x01, 0x0f, 0x00, 0x48}; !bytes.Equal(got, want) {
		t.Fatalf("CTY55 event0=%x, want %x", got, want)
	}
	for id, record := range map[string]int{
		event.DialogueTextIDs.Approach: 597,
		event.DialogueTextIDs.Success:  598,
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
			t.Fatalf("%s 未與 D3TXT00 rec%d 完整一致", id, record)
		}
	}
}

// 蓋亞之劍的 item dispatcher、唯一座標、5×4 原始 patch 與 world-state writer
// 必須直接對回同一份 DQ3.EXE；logical/file 換算見 docs/106。
func TestDQ3GaiaSwordMatchesOriginalEXEAndPackPatch(t *testing.T) {
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
		t.Fatal(err)
	}
	effect, ok := p.ItemUseEffectByRawID(0x0f)
	if !ok || effect.EffectID != "reveal_world_map_patch" || effect.UseTile == nil || effect.MapPatch == nil || effect.DirectMapPatch == nil {
		t.Fatalf("蓋亞之劍 item_use effect missing/incomplete: %+v", effect)
	}
	if effect.UseTile.X != 0x41 || effect.UseTile.Y != 0x6d || effect.RequiredLayer != 0 ||
		effect.RequiredVehicle != "" || effect.RequiredStoryFlagRaw != -1 || effect.Consume ||
		effect.SetWorldStateMask != 0x20 || effect.AnimationCycles != 30 ||
		effect.AnimationPaletteModeRaw != -1 {
		t.Fatalf("蓋亞之劍 JSON gate/transaction 不符原版: %+v", effect)
	}
	if effect.MapPatch.Origin.X != 0x3d || effect.MapPatch.Origin.Y != 0x6f ||
		effect.MapPatch.Width != 5 || effect.MapPatch.Height != 4 {
		t.Fatalf("蓋亞 patch geometry=%+v", effect.MapPatch)
	}
	wantPatch := []int{
		0x58, 0x58, 0x60, 0x58, 0x58,
		0x58, 0x68, 0x42, 0x66, 0x58,
		0x5e, 0x42, 0x7c, 0x42, 0x5c,
		0x58, 0x5a, 0x5a, 0x5a, 0x58,
	}
	if !reflect.DeepEqual(effect.MapPatch.TilesRaw, wantPatch) {
		t.Fatalf("蓋亞 patch tiles=%v, want %v", effect.MapPatch.TilesRaw, wantPatch)
	}
	wantDirectPatch := []int{
		0x27, 0x09, 0x09, 0x03, 0x00,
		0x09, 0x09, 0x03, 0x00, 0x42,
		0x03, 0x50, 0x00, 0x26, 0x42,
		0x6e, 0x26, 0x26, 0x25, 0x42,
	}
	if !reflect.DeepEqual(effect.DirectMapPatch.TilesRaw, wantDirectPatch) {
		t.Fatalf("蓋亞 direct patch tiles=%v, want %v", effect.DirectMapPatch.TilesRaw, wantDirectPatch)
	}
	for off, want := range map[int][]byte{
		0x50d5: {0x83, 0x3e, 0x2f, 0x4f, 0x41, 0x75, 0x0e},
		0x50dc: {0x83, 0x3e, 0x31, 0x4f, 0x6d, 0x75, 0x07},
		0x50e3: {0xe8, 0x3f, 0x0d, 0xe8, 0x50, 0x06, 0xc3},
		0x57d1: {0x83, 0x0e, 0x44, 0x4f, 0x20},
		0x19b94: {0x58, 0x58, 0x60, 0x58, 0x58, 0x58, 0x68, 0x42, 0x66, 0x58,
			0x5e, 0x42, 0x7c, 0x42, 0x5c, 0x58, 0x5a, 0x5a, 0x5a, 0x58},
		0x19cd6: {0x27, 0x09, 0x09, 0x03, 0x00, 0x09, 0x09, 0x03, 0x00, 0x42,
			0x03, 0x50, 0x00, 0x26, 0x42, 0x6e, 0x26, 0x26, 0x25, 0x42},
	} {
		if off+len(want) > len(exe) || !bytes.Equal(exe[off:off+len(want)], want) {
			got := []byte(nil)
			if off < len(exe) {
				end := off + len(want)
				if end > len(exe) {
					end = len(exe)
				}
				got = exe[off:end]
			}
			t.Fatalf("DQ3.EXE file %#x bytes=%x, want %x", off, got, want)
		}
	}
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

func TestDQ3DialogueWindowMatchesOriginalEXE(t *testing.T) {
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
	got := p.DialogueWindowLayout()
	want := WindowLayout{ID: "dq3:window.dialogue", X: 152, Y: 238, Width: 352, Height: 96,
		TextInsetX: 16, TextInsetY: 16, Columns: 20, LinesPerPage: 4, Frame: got.Frame, Evidence: got.Evidence}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("dialogue layout=%+v, want %+v", got, want)
	}
	// DGROUP 0x3e6e，file 0x19fae：kind、x byte、y、content width byte、content height。
	// sub_1fd30 畫出的可見框直接使用 +6/+8；sub_1fb36 的 +1/+0x10
	// 只是關閉時備份背景的保留範圍，不可當作玩家看見的外框。
	const off = 0x19fae
	wantRaw := []byte{0x0b, 0x01, 0x13, 0x00, 0xee, 0x00, 0x2c, 0x00, 0x60, 0x00}
	if len(exe) < off+len(wantRaw) || !bytes.Equal(exe[off:off+len(wantRaw)], wantRaw) {
		t.Fatalf("DQ3.EXE dialogue window structure at file %#x does not match", off)
	}
}

func TestDQ3BattleMessageWindowMatchesOriginalEXE(t *testing.T) {
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
	got, ok := p.BattleMessageWindowLayout()
	if !ok {
		t.Fatal("builtin pack missing battle message layout")
	}
	want := WindowLayout{ID: "dq3:window.battle_message", X: 152, Y: 238, Width: 352, Height: 96,
		TextInsetX: 16, TextInsetY: 16, Columns: 20, LinesPerPage: 4, Frame: got.Frame, Evidence: got.Evidence}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("battle message layout=%+v, want %+v", got, want)
	}
	// The shared win_rect at DGROUP 0x3e6e is consumed by both scripted
	// dialogue and battle text_draw; preserve the raw file bytes as parity.
	const off = 0x19fae
	wantRaw := []byte{0x0b, 0x01, 0x13, 0x00, 0xee, 0x00, 0x2c, 0x00, 0x60, 0x00}
	if len(exe) < off+len(wantRaw) || !bytes.Equal(exe[off:off+len(wantRaw)], wantRaw) {
		t.Fatalf("DQ3.EXE battle message window structure at file %#x does not match", off)
	}
}

func TestDQ3BattleHUDRectsMatchOriginalEXE(t *testing.T) {
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
	command, ok := p.BattleCommandLayout()
	if !ok || command.X != 144 || command.Y != 248 || command.Width != 128 || command.Height != 96 ||
		command.BaseRows != 2 || command.RowHeight != 16 || command.CursorInsetX != 16 ||
		command.RowInsetY != 16 || command.LabelInsetX != 32 || command.SecondaryLabelInsetX != 80 ||
		command.ValueInsetX != 112 || command.Frame == nil || command.Frame.BorderRGB != [3]uint8{255, 255, 255} ||
		command.Frame.InteriorRGB != [3]uint8{0, 0, 0} {
		t.Fatalf("battle command layout=%+v", command)
	}
	enemy, ok := p.BattleEnemyLayout()
	if !ok || enemy.X != 288 || enemy.Y != 248 || enemy.Width != 256 || enemy.Height != 48 ||
		enemy.BaseRows != 2 || enemy.RowHeight != 16 || enemy.NameInsetX != 32 || enemy.CountInsetX != 144 || enemy.CountInsetY == nil || *enemy.CountInsetY != 0 ||
		enemy.Frame == nil || enemy.Frame.ID != "dq3:frame.ega_window_white_black" {
		t.Fatalf("battle enemy layout=%+v", enemy)
	}
	labels, ok := p.BattleCommandLabels()
	if !ok {
		t.Fatal("battle command labels missing")
	}
	wantLabels := map[string][2]int{
		"war": {107, 207}, "flee": {629, 630}, "defend": {203, 204},
		"item": {402, 1354}, "spell": {429, 430},
	}
	for role, want := range wantLabels {
		got, exists := labels.Entries()[role]
		if !exists || got.PrimaryGlyph != want[0] || got.SecondaryGlyph != want[1] {
			t.Fatalf("battle command label %s=%+v, want glyphs %v", role, got, want)
		}
	}
	// DS:0x4112 is linear 0x28ee2 and maps to file 0x1a252; byte_28f00 is
	// linear 0x28f00/file 0x1a270. Keep both raw rects as parity anchors.
	for off, wantRaw := range map[int][]byte{
		0x1a252: {0x1a, 0x03, 0x12, 0x00, 0xf8, 0x00, 0x10, 0x00, 0x60, 0x00},
		0x1a270: {0x11, 0x0b, 0x24, 0x00, 0xf8, 0x00, 0x20, 0x00, 0x60, 0x00},
	} {
		if len(exe) < off+len(wantRaw) || !bytes.Equal(exe[off:off+len(wantRaw)], wantRaw) {
			t.Fatalf("DQ3.EXE battle HUD rect at file %#x does not match", off)
		}
	}
}

func TestBattleEnemyCountInsetYIsRequired(t *testing.T) {
	p := BattlePanelLayout{
		WindowLayout: WindowLayout{
			X: 1, Y: 1, Width: 64, Height: 64, TextInsetX: 8, TextInsetY: 8,
			Columns: 1, LinesPerPage: 1,
			Evidence: Evidence{Level: "D3", SourceKind: "exe", Source: "DQ3.EXE",
				AddressSpace: "file", Address: "0x1", Consumer: "test", Doc: "docs/x.md"},
		},
		HeightMode: "rows_plus_base", BaseRows: 2, RowHeight: 16,
	}
	if err := validateBattlePanelLayout("battle enemy", p, true); err == nil || !strings.Contains(err.Error(), "count_inset_y") {
		t.Fatalf("missing count_inset_y error=%v", err)
	}
	zero := 0
	p.CountInsetY = &zero
	if err := validateBattlePanelLayout("battle enemy", p, true); err != nil {
		t.Fatalf("valid count_inset_y rejected: %v", err)
	}
}

func TestDQ3FieldCommandLabelsMatchOriginalGlyphs(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	font, err := os.ReadFile(filepath.Join(dir, "D3TXT00.FON"))
	if err != nil {
		t.Skipf("original D3TXT00.FON unavailable: %v", err)
	}
	if len(font) != 47232 {
		t.Fatalf("D3TXT00.FON size=%d, want 47232", len(font))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(font)); got != "c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02" {
		t.Fatalf("D3TXT00.FON SHA-256=%s", got)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	labels, ok := p.FieldCommandLabels()
	if !ok {
		t.Fatal("field command labels missing")
	}
	want := map[string][2]int{
		"title": {274, 664}, "talk": {443, 449}, "spell": {429, 430},
		"status": {665, 666}, "item": {402, 237}, "equip": {197, 409},
		"examine": {544, 545},
	}
	for role, pair := range want {
		got, exists := labels.Entries()[role]
		if !exists || got.PrimaryGlyph != pair[0] || got.SecondaryGlyph != pair[1] {
			t.Fatalf("field command label %s=%+v, want glyphs %v", role, got, pair)
		}
		if got.Evidence.Level != "D2" || got.Evidence.Source != "D3TXT00.FON" {
			t.Fatalf("field command label %s evidence=%+v, want D2/D3TXT00.FON", role, got.Evidence)
		}
	}
}

func TestDQ3BattleSceneLayoutMatchesOriginalContract(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	scene, ok := p.BattleSceneLayout()
	if !ok {
		t.Fatal("battle scene layout missing")
	}
	if scene.ID != "dq3:battle_scene" || scene.FieldY0 != 80 || scene.FieldY1 != 246 ||
		scene.GroundY != 232 || scene.CursorGlyph != 0x77 {
		t.Fatalf("battle scene layout=%+v, want field 80..246 ground 232 cursor 0x77", scene)
	}
	if scene.Evidence.Level != "D2" || scene.Evidence.Doc != "docs/111-battle-scene-layout-re.md" {
		t.Fatalf("battle scene evidence=%+v, want D2 docs/111", scene.Evidence)
	}
}

func TestDQ3NewGameLabelsMatchOriginalGlyphs(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	font, err := os.ReadFile(filepath.Join(dir, "D3TXT00.FON"))
	if err != nil {
		t.Skipf("original D3TXT00.FON unavailable: %v", err)
	}
	if len(font) != 47232 {
		t.Fatalf("D3TXT00.FON size=%d, want 47232", len(font))
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(font)); got != "c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02" {
		t.Fatalf("D3TXT00.FON SHA-256=%s", got)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	labels, ok := p.NewGameLabels()
	if !ok {
		t.Fatal("new-game labels missing")
	}
	want := map[string][]int{
		"title":             {106, 187, 207, 710, 179},
		"start":             {113, 689, 488, 711},
		"load":              {712, 519, 696, 283},
		"male":              {775, 674},
		"female":            {234, 674},
		"level":             {12, 419, 12, 420, 667},
		"hp":                {12, 12, 22, 30, 667},
		"mp":                {12, 12, 27, 30, 667},
		"strength":          {170, 280, 667},
		"agility":           {282, 283, 667},
		"vitality":          {284, 170, 667},
		"intelligence":      {285, 286, 283, 667},
		"luck":              {276, 426, 425, 427, 667},
		"max_hp":            {291, 163, 22, 30, 667},
		"max_mp":            {291, 163, 27, 30, 667},
		"attack":            {623, 624, 170, 667},
		"defense":           {340, 409, 170, 667},
		"experience":        {525, 526, 667},
		"sex":               {12, 674, 12, 417, 667},
		"hero":              {106, 187},
		"cloth":             {190, 149, 191, 192},
		"prompt":            {398, 546, 194, 229, 456, 534, 58},
		"yes":               {399},
		"no":                {678},
		"choice_cursor":     {11},
		"backspace":         {11},
		"ok":                {29, 25},
		"name_title":        {690, 519, 691, 692},
		"name_left_arrow":   {693, 693},
		"name_right_arrow":  {694, 694},
		"name_zhuyin_grid":  {65, 70, 75, 80, 85, 90, 95, 100, 104, 66, 71, 76, 81, 86, 91, 96, 101, 105, 67, 72, 77, 82, 87, 92, 97, 102, 697, 68, 73, 78, 83, 88, 93, 98, 700, 121, 69, 74, 79, 84, 89, 94, 99, 103, 122},
		"name_alnum_grid":   {0, 5, 15, 20, 25, 30, 35, 40, 118, 1, 6, 16, 21, 26, 31, 36, 114, 119, 2, 7, 17, 22, 27, 32, 37, 115, 120, 3, 8, 18, 23, 28, 33, 38, 116, 121, 4, 9, 19, 24, 29, 34, 39, 117, 122},
		"input_mode_zhuyin": {690, 519, 702, 313},
	}
	if !reflect.DeepEqual(labels.Entries(), want) {
		t.Fatalf("new-game labels=%v, want %v", labels.Entries(), want)
	}
	if labels.Evidence.Level != "D2" || labels.Evidence.Source != "D3TXT00.TXT records 407,451-456 + D3TXT00.FON" || labels.Evidence.Doc != "docs/112-newgame-labels-re.md" {
		t.Fatalf("new-game labels evidence=%+v, want D2/FON+map/docs/112", labels.Evidence)
	}
}

func TestDQ3NewGameGeometryMatchesOriginalWindows(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	g, ok := p.NewGameGeometry()
	if !ok {
		t.Fatal("new-game geometry missing")
	}
	if g.ID != "dq3:new_game_geometry" ||
		g.Frame == nil || g.Frame.ID != "dq3:frame.ega_window_lavender_black" ||
		g.Frame.BorderPattern != "beveled_2px" ||
		g.Frame.BorderRGB != [3]uint8{255, 223, 255} ||
		g.Frame.InteriorRGB != [3]uint8{0, 0, 0} ||
		g.NamePanel != (GeometryRect{X: 159, Y: 52, Width: 242, Height: 130}) ||
		g.NameGrid != (GeometryGrid{X: 169, Y: 94, Columns: 9, Rows: 5, StepX: 16, StepY: 16}) ||
		g.NameTitle != (GeometryAnchor{X: 233, Y: 46, StepX: 16}) ||
		g.NameText != (GeometryAnchor{X: 249, Y: 62, StepX: 16}) ||
		g.GenderPanel != (GeometryRect{X: 344, Y: 46, Width: 96, Height: 64}) ||
		g.ConfirmChoiceContent != (GeometryRect{X: 376, Y: 78, Width: 80, Height: 32}) ||
		g.Stats == nil ||
		g.StatsLeft != (GeometryRect{X: 159, Y: 52, Width: 145, Height: 98, FrameEdgeWidths: FrameEdgeWidths{Right: 1}}) ||
		g.StatsEquipment != (GeometryRect{X: 159, Y: 148, Width: 145, Height: 82, FrameEdgeWidths: FrameEdgeWidths{Right: 1}}) ||
		g.StatsRight != (GeometryRect{X: 303, Y: 52, Width: 194, Height: 178}) ||
		g.StatsName != (GeometryAnchor{X: 200, Y: 46}) ||
		g.StatsCloth != (GeometryAnchor{X: 200, Y: 158}) ||
		g.Stats.Strength != (NewGameStatField{Label: GeometryAnchor{X: 376, Y: 62}, Value: NumberField{X: 408, Y: 62, Digits: 5}}) ||
		g.Stats.Luck != (NewGameStatField{Label: GeometryAnchor{X: 344, Y: 126}, Value: NumberField{X: 408, Y: 126, Digits: 5}}) ||
		g.Stats.Experience != (NewGameStatField{Label: GeometryAnchor{X: 328, Y: 206}, Value: NumberField{X: 360, Y: 206, Digits: 8}}) ||
		len(g.RawWindows) != 7 {
		t.Fatalf("new-game geometry=%+v, want raw windows and measured panels", g)
	}
	if g.RawWindows[0].Address != "linear:0x29088" || g.RawWindows[0].X != 19 ||
		g.RawWindows[0].Y != 46 || g.RawWindows[0].Width != 32 || g.RawWindows[0].Height != 144 {
		t.Fatalf("name raw window=%+v, want linear 0x29088 19,46,32,144", g.RawWindows[0])
	}
	if len(g.WindowBackdrops) != 3 ||
		g.WindowBackdrops[0].RawWindowID != "creation_panel" ||
		g.WindowBackdrops[0].Primitive != "ega_window_black_backdrop" ||
		g.WindowBackdrops[1].RawWindowID != "confirm_prompt" ||
		g.WindowBackdrops[2].RawWindowID != "confirm_choice" {
		t.Fatalf("new-game window backdrops=%+v, want three ordered raw EGA black backdrops", g.WindowBackdrops)
	}
	for i, want := range []struct{ order, right, bottom int }{{0, 8, 8}, {1, 0, 8}, {1, 0, 0}} {
		backdrop := g.WindowBackdrops[i]
		if backdrop.DrawOrder == nil || backdrop.TerminalRightPixels == nil || backdrop.TerminalBottomPixels == nil ||
			*backdrop.DrawOrder != want.order || *backdrop.TerminalRightPixels != want.right || *backdrop.TerminalBottomPixels != want.bottom {
			t.Fatalf("window_backdrops[%d]=%+v, want draw_order/terminal=%+v", i, backdrop, want)
		}
	}
	if g.Evidence.Level != "D2" || g.Evidence.Doc != "docs/113-newgame-geometry-re.md" {
		t.Fatalf("new-game geometry evidence=%+v, want D2/docs/113", g.Evidence)
	}
	if g.ConfirmChoiceBackdrop == nil ||
		g.ConfirmChoiceBackdrop.ID != "dq3:new_game.confirm_choice_blue_checkerboard" ||
		g.ConfirmChoiceBackdrop.Pattern != "checkerboard_1px" ||
		g.ConfirmChoiceBackdrop.Rect != (GeometryRect{X: 360, Y: 62, Width: 112, Height: 64}) ||
		g.ConfirmChoiceBackdrop.RGB != [3]uint8{0, 85, 223} {
		t.Fatalf("confirm choice backdrop=%+v, want original blue checkerboard rect", g.ConfirmChoiceBackdrop)
	}
	if g.ConfirmChoiceBackdrop.Evidence.Level != "D2" ||
		g.ConfirmChoiceBackdrop.Evidence.Doc != "docs/118-newgame-choice-backdrop-re.md" ||
		g.ConfirmChoiceBackdrop.Evidence.Address != "0x28bc6" {
		t.Fatalf("confirm choice backdrop evidence=%+v, want D2/linear 0x28bc6/docs/118", g.ConfirmChoiceBackdrop.Evidence)
	}
	if g.ConfirmChoiceFrame == nil ||
		g.ConfirmChoiceFrame.ID != "dq3:new_game.confirm_choice_lavender_skin_frame" ||
		g.ConfirmChoiceFrame.BorderPattern != "checkerboard_frame_2px" ||
		g.ConfirmChoiceFrame.BorderRGB != [3]uint8{255, 223, 255} ||
		g.ConfirmChoiceFrame.BorderAccentRGB == nil ||
		*g.ConfirmChoiceFrame.BorderAccentRGB != [3]uint8{235, 178, 130} {
		t.Fatalf("confirm choice frame=%+v, want named lavender/skin 2px frame", g.ConfirmChoiceFrame)
	}
	if g.ConfirmChoiceFrame.Evidence.Level != "D2" ||
		g.ConfirmChoiceFrame.Evidence.Doc != "docs/118-newgame-choice-backdrop-re.md" ||
		g.ConfirmChoiceFrame.Evidence.AddressSpace != "linear" ||
		g.ConfirmChoiceFrame.Evidence.Address != "0x28bc6" {
		t.Fatalf("confirm choice frame evidence=%+v, want D2/linear 0x28bc6/docs/118", g.ConfirmChoiceFrame.Evidence)
	}
	if g.Frame.Evidence.Level != "D2" || g.Frame.Evidence.Doc != "docs/117-newgame-frame-re.md" ||
		g.Frame.Evidence.AddressSpace != "dgroup" {
		t.Fatalf("new-game frame evidence=%+v, want D2/dgroup/docs/117", g.Frame.Evidence)
	}
}

func TestDQ3PartySpriteAndHUDContract(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	layout, ok := p.PartyHUDLayout()
	if !ok || layout.ID != "dq3:window.field_party_hud" ||
		layout.X != 152 || layout.Y != 238 || layout.Width != 352 || layout.Height != 80 ||
		layout.TextInsetX != 16 || layout.TextInsetY != 0 ||
		layout.Columns != 4 || layout.LinesPerPage != 4 ||
		layout.HPLabelGlyph == nil || *layout.HPLabelGlyph != 22 ||
		layout.MPLabelGlyph == nil || *layout.MPLabelGlyph != 27 ||
		layout.NameInsetX != 16 || layout.ValueInsetX != 16 || layout.NameMaxGlyphs != 4 ||
		len(layout.ClassGlyphs) != 8 || len(layout.ClassGlyphs[0]) != 1 || layout.ClassGlyphs[0][0] != 106 ||
		len(layout.ClassGlyphs[7]) != 1 || layout.ClassGlyphs[7][0] != 113 ||
		layout.Frame == nil || layout.Frame.BorderRGB != [3]uint8{255, 255, 255} ||
		layout.Frame.InteriorRGB != [3]uint8{0, 0, 0} {
		t.Fatalf("party HUD layout=%+v, want original DQ3 four-column contract", layout)
	}
	if layout.Evidence.Level != "D3" || layout.Evidence.Doc != "docs/130-party-field-hud-re.md" ||
		layout.Evidence.AddressSpace != "linear" {
		t.Fatalf("party HUD evidence=%+v, want D3 IDA/player-visible closure", layout.Evidence)
	}
	wantEntries := []int{0, 4, 8, 12, 16, 20, 24, 28, 32, 36, 40, 44, 48, 52, 56, 60}
	for class := 0; class < 8; class++ {
		for gender := 0; gender < 2; gender++ {
			asset, entry, ok := p.PartySprite(class, gender)
			if !ok || asset != "DQ3MST.BLS" || entry != wantEntries[class*2+gender] {
				t.Fatalf("party sprite class=%d gender=%d = %q/%d/%v", class, gender, asset, entry, ok)
			}
		}
	}
}

func TestDQ3AttractContract(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	seq, ok := p.AttractSequence()
	if !ok || seq.ID != "dq3:attract.class_showcase" || seq.StartDelayFrames != 1200 || len(seq.Frames) != 8 {
		t.Fatalf("attract sequence=%+v, want eight-card idle contract", seq)
	}
	want := []string{
		"attract_warrior", "attract_priest", "attract_magician", "attract_fighter",
		"attract_merchant", "attract_worthy", "attract_player", "attract_hero",
	}
	for i, frame := range seq.Frames {
		if frame.AssetKey != want[i] || frame.HoldFrames != 1200 {
			t.Fatalf("attract frame[%d]=%+v, want key=%q hold=1200", i, frame, want[i])
		}
		if ref, ok := p.Asset(frame.AssetKey); !ok || ref.Path == "" || ref.Size <= 0 || ref.SHA256 == "" {
			t.Fatalf("attract asset %q missing manifest integrity", frame.AssetKey)
		}
	}
}

func TestDQ3OpeningCutsceneContract(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	seq, ok := p.OpeningSequence()
	if !ok || seq.ID != "dq3:opening.boot_cutscene" || !seq.SkipOnInput || len(seq.Frames) != 5 {
		t.Fatalf("opening sequence=%+v, want five-frame skip-on-input contract", seq)
	}
	want := []string{"opening_year", "opening_dragon", "opening_party", "opening_year_1993", "opening_crest"}
	for i, frame := range seq.Frames {
		if frame.AssetKey != want[i] || frame.HoldFrames != 120 {
			t.Fatalf("opening frame[%d]=%+v, want key=%q hold=120", i, frame, want[i])
		}
		if ref, ok := p.Asset(frame.AssetKey); !ok || ref.Size <= 0 || ref.SHA256 == "" {
			t.Fatalf("opening asset %q missing manifest integrity", frame.AssetKey)
		}
	}
}

func TestDQ3NewGameConfirmationContract(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	screen, ok := p.NewGameConfirmation()
	if !ok || screen.ID != "dq3:screen.new_game_confirmation" ||
		screen.AssetKey != "new_game_confirmation" || screen.Width != 640 || screen.Height != 350 ||
		screen.Format != "row_interleaved_4bpp" || len(screen.Palette) != 16 {
		t.Fatalf("new-game confirmation contract=%+v, want FIRST.SCR screen", screen)
	}
	if ref, ok := p.Asset(screen.AssetKey); !ok || ref.Path != "FIRST.SCR" || ref.Size != 112000 || ref.SHA256 == "" {
		t.Fatalf("new-game confirmation manifest ref missing integrity: %+v", ref)
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

func TestDQ3PoisonPackMatchesOriginalActionAndConsumers(t *testing.T) {
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
	actions := p.MonsterActionDefinitions()
	byBit := make(map[int]MonsterActionDefinition, len(actions))
	for _, action := range actions {
		byBit[action.MaskBit] = action
	}
	poison, ok := byBit[38]
	if !ok || poison.ActionRaw != 0x32 || poison.SuccessRollMax != 0x64 ||
		poison.ConditionID != PoisonCondition {
		t.Fatalf("poison monster action=%+v", actions)
	}
	// DGROUP 0x3930 remap：bit38 的 index 是 38-18=20；DGROUP 0x394e
	// handler table 的 action0x32 entry 是 logical0xa307。
	if got := exe[0x19a70+20]; got != byte(poison.ActionRaw) {
		t.Fatalf("poison remap JSON=%#x EXE=%#x", poison.ActionRaw, got)
	}
	if got := binary.LittleEndian.Uint16(exe[0x19aca:]); got != 0xa307 {
		t.Fatalf("poison handler pointer=%#x want logical0xa307", got)
	}
	condition, ok := p.PoisonConditionDefinition()
	if !ok || condition.LegacyStatusMask != 0x40 || condition.FieldDamagePerStep != 1 ||
		condition.SuppressWorldStateMask != 3 {
		t.Fatalf("poison condition=%+v ok=%v", condition, ok)
	}
	for off, want := range map[int][]byte{
		0x84d2: {0xa9, 0x40, 0x00},                   // church test +0x38 bit0x40
		0x84e6: {0xc7, 0x06, 0x93, 0x25, 0x05, 0x00}, // fixed 5G writer
		0xa9f5: {0xa9, 0x40, 0x00},                   // movement poison test
		0xaa02: {0x80, 0x87, 0x62, 0x06, 0x01},       // per-step damage +1
		0xb6a5: {0x80, 0x4d, 0x38, 0x40},             // poison-gas status writer
	} {
		if got := exe[off : off+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("DQ3.EXE file %#x=%x want %x", off, got, want)
		}
	}
	if p.CurePoisonCost() != 5 {
		t.Fatalf("CurePoisonCost=%d want 5", p.CurePoisonCost())
	}
	if p.RemoveCurseCost(7) != 700 {
		t.Fatalf("RemoveCurseCost(7)=%d want 700", p.RemoveCurseCost(7))
	}
	wantCursed := []int{0x1b, 0x1d, 0x37, 0x38, 0x3d}
	if got := p.CursedEquipmentRawIDs(); !reflect.DeepEqual(got, wantCursed) {
		t.Fatalf("cursed equipment ids=%v want %v", got, wantCursed)
	}
	itemsRaw, err := os.ReadFile(filepath.Join(dir, "ITEM.DAT"))
	if err != nil {
		t.Fatalf("read ITEM.DAT: %v", err)
	}
	items, err := dq3data.OpenItems(itemsRaw)
	if err != nil {
		t.Fatalf("OpenItems: %v", err)
	}
	for rawID := 0; rawID < dq3data.ItemCount; rawID++ {
		if got, want := p.IsCursedEquipment(rawID), items.CursedWhenEquipped(rawID); got != want {
			t.Fatalf("item %#x pack cursed=%v ITEM writer oracle=%v", rawID, got, want)
		}
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	decoded := dq3data.LoadText(nil, txtRaw)
	for id, record := range map[string]int{
		"common:text.battle.actor.poisoned": 371,
		"common:text.battle.poison.gas":     375,
	} {
		got, ok := p.TextGlyphCodes(id)
		if !ok {
			t.Fatalf("missing poison text %q", id)
		}
		raw := decoded.Record(record)
		if !reflect.DeepEqual(got, raw) {
			t.Fatalf("%s glyphs=%v want D3TXT00 rec%d=%v", id, got, record, raw)
		}
	}
}

func TestDQ3DefeatRespawnTextsAndConsumerMatchOriginal(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	refs, ok := p.BattleTextRefs()
	if !ok {
		t.Fatal("missing battle_texts")
	}
	decoded := dq3data.LoadText(nil, txtRaw)
	for id, record := range map[string]int{
		refs.PartyDefeated: 361,
		refs.RevivalVoice:  362,
	} {
		got, ok := p.TextGlyphCodes(id)
		if !ok {
			t.Fatalf("missing defeat text %q", id)
		}
		if want := decoded.Record(record); !reflect.DeepEqual(got, want) {
			t.Fatalf("%s glyphs=%v want D3TXT00 rec%d=%v", id, got, record, want)
		}
	}
	// IDA 9.4 linear 0x1962f selects record 361 before the all-down
	// consumer; 0x1c7fd selects record 362, then 0x1c815 addresses the
	// fixed first character and restores current HP from max HP.  File
	// offsets use file=(linear-0x10000)+0x1370 for this exact executable.
	for off, want := range map[int][]byte{
		0xa99f: {0xbf, 0x69, 0x01},
		0xdb6d: {0xbf, 0x6a, 0x01},
		0xdb85: {0x8d, 0x36, 0x7f, 0x50, 0x81, 0x64, 0x38, 0x7f, 0xff, 0x8b, 0x44, 0x2a, 0x89, 0x44, 0x16},
	} {
		if got := exe[off : off+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("DQ3.EXE file %#x=%x want %x", off, got, want)
		}
	}
}

func TestDQ3Monster46SleepActionAndMPMatchOriginal(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	mons, err := os.ReadFile(filepath.Join(dir, "D3MNS.DAT"))
	if err != nil {
		t.Fatalf("read D3MNS.DAT: %v", err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	var sleep MonsterActionDefinition
	for _, action := range p.MonsterActionDefinitions() {
		if action.MaskBit == 41 {
			sleep = action
		}
	}
	if sleep.ActionRaw != 0x35 || sleep.SuccessRollMax != 0xb4 ||
		sleep.ConditionID != SleepCondition || sleep.CastTextRole != "sleep_breath" {
		t.Fatalf("monster46 sleep action=%+v", sleep)
	}
	record := mons[46*0x29 : 47*0x29]
	if !bytes.Equal(record[14:20], []byte{0, 0x40, 0, 0, 0, 0x40}) ||
		binary.LittleEndian.Uint16(record[4:6]) != 10 {
		t.Fatalf("monster46 MP/mask=%d/%x", binary.LittleEndian.Uint16(record[4:6]), record[14:20])
	}
	if got := exe[0x19a70+23]; got != byte(sleep.ActionRaw) {
		t.Fatalf("sleep remap JSON=%#x EXE=%#x", sleep.ActionRaw, got)
	}
	if got := binary.LittleEndian.Uint16(exe[0x19ad0:]); got != 0xa3ee {
		t.Fatalf("sleep handler pointer=%#x want logical0xa3ee", got)
	}
	for off, want := range map[int][]byte{
		0xb776: {0xf7, 0x44, 0x38, 0xa0, 0x00}, // skip dead/already sleeping
		0xb780: {0x3c, 0xb4},                   // success roll <=180
		0xb784: {0x80, 0x4c, 0x38, 0x20},       // +0x38 bit0x20 writer
	} {
		if got := exe[off : off+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("DQ3.EXE file %#x=%x want %x", off, got, want)
		}
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatalf("read D3TXT00.TXT: %v", err)
	}
	decoded := dq3data.LoadText(nil, txtRaw)
	got, ok := p.TextGlyphCodes("common:text.battle.sleep.breath")
	if !ok || !reflect.DeepEqual(got, decoded.Record(369)) {
		t.Fatalf("sleep breath glyphs=%v ok=%v want rec369=%v", got, ok, decoded.Record(369))
	}
}

func TestDQ3AntidotePackMatchesOriginalDispatcherHandlerAndText(t *testing.T) {
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
	effect, ok := p.ItemUseEffectByRawID(0x42)
	if !ok || effect.EffectID != "clear_condition" || effect.ConditionID != PoisonCondition ||
		effect.TargetScope != "party_member" || !effect.Consume {
		t.Fatalf("antidote effect=%+v ok=%v", effect, ok)
	}
	if got := binary.LittleEndian.Uint16(exe[0x197ac:]); got != 0x3dc3 {
		t.Fatalf("item0x42 handler pointer=%#x want logical0x3dc3", got)
	}
	for off, want := range map[int][]byte{
		0x5133: {0xe8, 0x33, 0x0f, 0xba, 0x40, 0x00, 0xbd, 0x6c, 0x01},
		0x6069: {0xe8, 0x38, 0x4b, 0x83, 0xc6, 0x3a, 0xb9, 0x08, 0x00},
		0x5a2c: {0x8b, 0x44, 0x38, 0x85, 0xc2},
		0x5a41: {0x21, 0x54, 0x38},
	} {
		if got := exe[off : off+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("DQ3.EXE file %#x=%x want %x", off, got, want)
		}
	}
	txtRaw, err := os.ReadFile(filepath.Join(dir, "D3TXT00.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	decoded := dq3data.LoadText(nil, txtRaw)
	for id, record := range map[string]int{
		effect.NoEffectTextID: 341,
		effect.SuccessTextID:  364,
	} {
		got, ok := p.TextGlyphCodes(id)
		if !ok || !reflect.DeepEqual(got, decoded.Record(record)) {
			t.Fatalf("%s glyphs=%v ok=%v want rec%d=%v", id, got, ok, record, decoded.Record(record))
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

func TestDQ3LancelCourageTrialMatchesOriginalEXECTYAndText(t *testing.T) {
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
	cty23, cty47, cty75 := read("CTY23.DAT"), read("CTY47.DAT"), read("CTY75.DAT")
	txt06, txt07 := read("D3TXT06.TXT"), read("D3TXT07.TXT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	event, ok := p.TemporarySoloChallenge("dq3:event.lancel_courage_trial")
	if !ok {
		t.Fatal("missing dq3:event.lancel_courage_trial")
	}
	if event.EntryNPC != (ScriptedNPCSelector{CTYRaw: 47, Section: 0, Tile: TileCoordinate{X: 25, Y: 16}, HandlerRaw: 37}) ||
		event.ReturnNPC != (ScriptedNPCSelector{CTYRaw: 75, Section: 0, Tile: TileCoordinate{X: 26, Y: 7}, HandlerRaw: 62}) ||
		event.CompletedFlagRaw != 0x13 || event.ModeMaskRaw != 0x80 ||
		event.SoloWorldPosition != (VehicleWorldPosition{X: 82, Y: 165, Layer: 0}) ||
		event.ReturnTownDestination != (SceneDestination{CTYRaw: 47, Section: 0, X: 25, Y: 17}) {
		t.Fatalf("勇氣試煉 JSON 設定不符：%+v", event)
	}

	findNPC := func(raw []byte, section int, want ScriptedNPCSelector) bool {
		base := int(binary.LittleEndian.Uint16(raw[section*2:]))
		table := base + int(binary.LittleEndian.Uint16(raw[base:]))
		for i := 0; i < int(raw[table]); i++ {
			rec := raw[table+1+i*7 : table+1+(i+1)*7]
			if int(rec[0]) == want.Tile.X && int(rec[1]) == want.Tile.Y &&
				(rec[3]>>3)&7 == 2 && int(rec[4]) == want.HandlerRaw {
				return true
			}
		}
		return false
	}
	if !findNPC(cty47, 0, event.EntryNPC) || !findNPC(cty75, 0, event.ReturnNPC) {
		t.Fatal("CTY47/75 找不到 game-pack 指定的 handler37/62 scripted NPC")
	}

	sec23 := int(binary.LittleEndian.Uint16(cty23[2*2:]))
	eventTable := sec23 + int(binary.LittleEndian.Uint16(cty23[sec23+8:]))
	if eventTable+5 > len(cty23) || cty23[eventTable] < 1 ||
		!reflect.DeepEqual(cty23[eventTable+1:eventTable+5], []byte{0x01, 0x67, 0x00, 0xad}) {
		t.Fatalf("CTY23 sec2 event0 不是藍寶珠 raw 01 67 00 ad")
	}
	treasure, ok := p.TreasureEvent("dq3:event.courage_cave_blue_orb")
	if !ok || treasure.Treasure != (QuestTreasureSelector{CTYRaw: 23, Section: 2, TileSubID: 0,
		EventTypeRaw: 1, ItemRawID: 0x67, PresentFlag: 0xad}) {
		t.Fatalf("藍寶珠 treasure JSON 不符：%+v", treasure)
	}

	// Exact raw-instruction anchors from the non-destructive IDA 9.4 trace.
	// logical0x59e4/file0x6d54 writes count 1, world (82,165) and ORs mode0x80;
	// logical0x608a/file0x73fa restores the count and ANDs mode with 0xff7f.
	if !bytes.Equal(exe[0x6dce:0x6de2], []byte{0xc6, 0x06, 0x77, 0x50, 0x01, 0x90,
		0xc7, 0x06, 0x8a, 0x25, 0x01, 0x00, 0xc7, 0x06, 0xb0, 0x26, 0x00, 0x00, 0xc6, 0x06}) ||
		!bytes.Equal(exe[0x6df5:0x6e0b], []byte{0xc7, 0x06, 0x48, 0x4f, 0x52, 0x00,
			0xc7, 0x06, 0x4a, 0x4f, 0xa5, 0x00, 0x81, 0x0e, 0x46, 0x4f, 0x80, 0x00, 0xc3, 0xbf, 0xd8, 0x0b}) ||
		!bytes.Equal(exe[0x73fa:0x740e], []byte{0xbf, 0xc4, 0x0b, 0xe8, 0x93, 0xef,
			0xa0, 0x57, 0x50, 0xa2, 0x77, 0x50, 0x81, 0x26, 0x46, 0x4f, 0x7f, 0xff, 0xe8, 0x53}) {
		t.Fatal("DQ3.EXE handler37/62 raw anchor drift")
	}

	checkRecord := func(raw []byte, id string, record int) {
		t.Helper()
		start := int(binary.LittleEndian.Uint16(raw[record*2:]))
		end := int(binary.LittleEndian.Uint16(raw[(record+1)*2:]))
		codes := make([]uint16, 0, (end-start)/2)
		for off := start; off+2 <= end; off += 2 {
			codes = append(codes, binary.LittleEndian.Uint16(raw[off:off+2]))
		}
		got, found := p.TextGlyphCodes(id)
		if !found || !reflect.DeepEqual(got, codes) {
			t.Fatalf("%s 未與原始 record%d 完整一致", id, record)
		}
	}
	for id, record := range map[string]int{
		event.DialogueTextIDs.Prompt: 9, event.DialogueTextIDs.Accept: 10,
		event.DialogueTextIDs.Reject: 11, event.DialogueTextIDs.Return: 12,
		event.DialogueTextIDs.Completed: 84,
	} {
		checkRecord(txt06, id, record)
	}
	checkRecord(txt07, event.DialogueTextIDs.CaveBack, 66)
	checkRecord(txt07, event.DialogueTextIDs.CaveEmpty, 67)
}

func TestDQ3ModChangeMirrorMatchesOriginalEXEAndCTY(t *testing.T) {
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
	exe, cty := read("DQ3.EXE"), read("CTY24.DAT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	treasure, ok := p.TreasureEvent("dq3:event.shamanoasa_mod_change_mirror")
	if !ok || treasure.Treasure != (QuestTreasureSelector{CTYRaw: 24, Section: 2, TileSubID: 0,
		EventTypeRaw: 1, ItemRawID: 0x61, PresentFlag: 0x9f}) {
		t.Fatalf("拉之鏡 treasure JSON 不符：%+v", treasure)
	}

	section := int(binary.LittleEndian.Uint16(cty[4:6]))
	eventTable := section + int(binary.LittleEndian.Uint16(cty[section+8:section+10]))
	off := eventTable + 1 + treasure.Treasure.TileSubID*4
	if section != 0x2c77 || eventTable != 0x2c90 || off+4 > len(cty) ||
		cty[eventTable] <= byte(treasure.Treasure.TileSubID) {
		t.Fatalf("CTY24 sec2 event table 無效：section=%#x table=%#x", section, eventTable)
	}
	if raw := cty[off : off+4]; !bytes.Equal(raw, []byte{0x01, 0x61, 0x00, 0x9f}) {
		t.Fatalf("CTY24 sec2 subid0=%x，want 01 61 00 9f", raw)
	}
	// 通用 examine consumer：事件 type1 將 raw item 交給 inventory writer；
	// writer 成功後才清除 p2 指定的 present flag。
	if raw := exe[0x9d86:0x9da0]; !bytes.Equal(raw, []byte{
		0x8b, 0xcd, 0x80, 0xf9, 0x03, 0x74, 0x10, 0x80,
		0xf9, 0x00, 0x75, 0x06, 0xe8, 0x2c, 0x01, 0xeb,
		0x1f, 0x90, 0x80, 0xf9, 0x01, 0x75, 0x06, 0xe8,
		0x49, 0x00,
	}) {
		t.Fatalf("DQ3.EXE common examine consumer file0x9d86 raw drifted: %x", raw)
	}
}

func TestDQ3PiratesRedOrbMatchesOriginalEXEAndCTY(t *testing.T) {
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
	exe, cty := read("DQ3.EXE"), read("CTY27.DAT")
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	treasure, ok := p.TreasureEvent("dq3:event.pirates_den_red_orb")
	if !ok || treasure.Treasure != (QuestTreasureSelector{CTYRaw: 27, Section: 1, TileSubID: 0,
		EventTypeRaw: 1, ItemRawID: 0x68, PresentFlag: 0x3f}) {
		t.Fatalf("紅寶珠 treasure JSON 不符：%+v", treasure)
	}

	for off, want := range map[int][]byte{
		0x25:  {0x1a, 0x09, 0x28, 0xc1, 0x0e, 0x3f, 0x00}, // daytime entrance object
		0x34:  {0x1a, 0x09, 0x28, 0xc1, 0x0e, 0x3f, 0x00}, // nighttime entrance object
		0x81:  {0x1b, 0x01, 0x06, 0x03},                   // ordinary isolated entrance
		0x85:  {0x1b, 0x01, 0x05, 0x09},                   // object-covered treasure entrance
		0x2ef: {0x21, 0x01},                               // CTY27 sec0 tile (26,9), subid1
		0xa5e: {0x03, 0x01, 0x68, 0x00, 0x3f, 0x01, 0x50, 0x00, 0x40, 0x01, 0x43, 0x00, 0x41},
	} {
		if off+len(want) > len(cty) || !bytes.Equal(cty[off:off+len(want)], want) {
			t.Fatalf("CTY27 file %#x raw drift: got=%x want=%x", off, cty[off:off+len(want)], want)
		}
	}
	// IDA linear 0x1353a/file 0x48aa: section+0xc transition table and
	// (tile high byte & 0x1f)*4 selector consumer.
	for off, want := range map[int][]byte{
		0x48aa: {0x8b, 0x16, 0x24, 0x0b, 0xbf, 0x0c, 0x00, 0x03, 0xfa,
			0x26, 0x8b, 0x35, 0x03, 0xf2, 0x8a, 0x1e, 0x55, 0x0b, 0x80, 0xe3, 0x1f,
			0x32, 0xff, 0xd1, 0xe3, 0xd1, 0xe3, 0x03, 0xf3},
		// IDA linear 0x1de2a/file 0xf19a and 0x1de4c/file 0xf1bc:
		// invisibility blocks pushing; otherwise NPC ctrl bit0x40 is required.
		0xf19a: {0xf7, 0x06, 0x46, 0x4f, 0x00, 0xc0, 0x74, 0x03},
		0xf1bc: {0xb0, 0x40, 0x84, 0x45, 0x03, 0x74, 0x6b},
	} {
		if off+len(want) > len(exe) || !bytes.Equal(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x raw drift", off)
		}
	}
}

func TestLoadRejectsUnknownAndInvalidData(t *testing.T) {
	validManifest := `{
	  "schema_version":"0.1.44","pack_id":"test","game":"dq3","edition":"cht_jingxun",
	  "content_version":"0.1.0","engine_api":">=0.1.0 <0.2.0",
	  "title_text_id":"x:title","entry_event_id":"x:new","save_namespace":"test",
	  "capabilities":[],"data":{"facilities":"facilities.json","events":"events.json","interface":"interface.json",
	  "characters":"characters.json","texts":"texts.json","spells":"spells.json"},"assets":{"pal":{"path":"PAL","size":0,"sha256":""}}
	}`
	validFacilities := `{
	  "schema_version":"0.1.44","service_definitions":[{
	    "id":"common:service.cure_poison","pricing":{"formula_id":"common:formula.fixed","fixed_gold":5},
	    "evidence":{"level":"D2","source_kind":"exe","source":"x","address_space":"file","address":"1","consumer":"x","doc":"x"}}, {
	    "id":"common:service.remove_curse","pricing":{"formula_id":"common:formula.level_multiplier","gold_per_level":100},
	    "affected_equipped_item_raw_ids":[27],
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE+ITEM.DAT","address_space":"logical","address":"0x1","consumer":"curse writer","doc":"docs/x.md"}}, {
	    "id":"common:service.revive",
	    "pricing":{"formula_id":"common:formula.level_table","level_cap":1,"costs_gold":[10]},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"reader","doc":"docs/x.md"}
	  }]
	}`
	validEvents := `{
	  "schema_version":"0.1.44",
	  "day_night_cycle":{"clock_ticks":4,"night_start_tick":2,"palette_segment_ticks":1,
	    "palette_entries_per_bank":1,"palette_bank_indices":[0,0,0,0],"palette_asset_key":"pal",
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"palette","doc":"docs/x.md"}},
	  "item_actions":{"personal_inventory_slots":8,
	    "text_ids":{"use":"x:text","give":"x:text","drop":"x:text"},
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"record","address":"421","consumer":"item menu","doc":"docs/x.md"}},
	  "rura_navigation":{"destinations":[],
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"logical","address":"0x1","consumer":"teleport","doc":"docs/x.md"}},
	  "npc_push_rule":{"ctrl_mask":64,"ctrl_value":64,"blocked_by_effect_id":"temporary_invisibility",
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"collision","doc":"docs/x.md"}},
	  "boss_surrender_events":[],"temporary_role_events":[],"temporary_solo_challenges":[],"conditional_facility_events":[],
	    "world_entrance_variants":[],"settlement_founder_events":[],"settlement_founder_followups":[],"quest_item_chain_events":[],"choice_item_exchange_events":[],"tracked_world_objects":[],"coordinate_item_gate_events":[],"treasure_events":[],"npc_item_reward_events":[],"item_use_effects":[],"tracking_guard_events":[],
	    "push_puzzle_events":[],
	    "two_step_floor_switch_gates":[],
	  "staged_vehicle_exchange_events":[],"guided_passage_events":[],
  "hostage_rescue_events":[],"reclass_events":[],"staged_boss_events":[],"story_flag_runtime_events":[]
	}`
	validCharacters := `{
	  "schema_version":"0.1.44",
	  "default_refs":{"new_game_player":"test:character.player"},
	  "defaults":[
	    {"id":"test:character.player",
	     "equipment":{"weapon":null,"armor":30,"shield":null,"head":null},
	     "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	       "address_space":"file","address":"0x1","consumer":"writer","doc":"docs/x.md"}}
	  ]
	}`
	validTexts := `{
	  "schema_version":"0.1.44","definitions":[{
	    "id":"x:text","value":"字","glyph_codes":[1],
	    "layout":{"kind":"menu_label"},
	    "source":{"kind":"glyph_map","file":"font.bin"},
	    "evidence":{"level":"D3","source_kind":"data_file","source":"font.bin",
	      "address_space":"record","address":"glyph 1","consumer":"renderer","doc":"docs/x.md"}
	  }]
	}`
	validInterface := `{
	  "schema_version":"0.1.44","dialogue":{"id":"x:dialogue","x":1,"y":1,
	    "width":64,"height":64,"text_inset_x":8,"text_inset_y":8,
	    "columns":3,"lines_per_page":3,
	    "evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE",
	      "address_space":"file","address":"0x1","consumer":"renderer","doc":"docs/x.md"}}
	}`
	validSpells := `{"schema_version":"0.1.44","field":[]}`
	tests := []struct {
		name, manifest, facilities, events, characters, want string
	}{
		{"unknown manifest field", strings.Replace(validManifest, `"save_namespace":"test",`, `"save_namespace":"test","typo":1,`, 1), validFacilities, validEvents, validCharacters, "unknown field"},
		{"path escape", strings.Replace(validManifest, `"facilities.json"`, `"../facilities.json"`, 1), validFacilities, validEvents, validCharacters, "pack-relative"},
		{"cost length", validManifest, strings.Replace(validFacilities, `"level_cap":1`, `"level_cap":2`, 1), validEvents, validCharacters, "must equal"},
		{"unknown facilities field", validManifest, strings.Replace(validFacilities, `"schema_version":"0.1.44"`, `"schema_version":"0.1.44","typo":1`, 1), validEvents, validCharacters, "unknown field"},
		{"unknown events field", validManifest, validFacilities, strings.Replace(validEvents, `"boss_surrender_events":[]`, `"boss_surrender_events":[],"typo":1`, 1), validCharacters, "unknown field"},
		{"invalid push puzzle", validManifest, validFacilities, strings.Replace(validEvents,
			`"push_puzzle_events":[]`,
			`"push_puzzle_events":[{"id":"x:push","kind":"push_puzzle","cty_raw":76,"section":0,"trigger_tiles":[],"npcs":[],"completion_axis":"x","completion_value":5,"clear_story_flag_raw":59,"evidence":{"level":"D3","source_kind":"exe","source":"DQ3.EXE","address_space":"file","address":"0x1","consumer":"reader","doc":"docs/x.md"}}]`, 1), validCharacters, "invalid event"},
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
				"interface.json":  {Data: []byte(validInterface)},
				"spells.json":     {Data: []byte(validSpells)},
				"PAL":             {Data: []byte{0}},
			}
			_, err := Load(fsys)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Load error=%v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestFieldHealSpellPackMatchesOriginalDescriptors(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	want := map[int]string{
		161: "031e0b", 162: "05550b", 163: "07ff0b", 164: "123213", 165: "3eff13",
	}
	for rec, raw := range want {
		d, ok := p.FieldSpellByRecord(rec)
		if !ok || d.DescriptorHex != raw || d.EffectID != FieldHealHP {
			t.Fatalf("field spell rec%d mismatch: ok=%v def=%+v want raw=%s", rec, ok, d, raw)
		}
		if rec <= 163 && d.TargetScope != "party_member" || rec >= 164 && d.TargetScope != "party_all" {
			t.Fatalf("field spell rec%d target scope=%q", rec, d.TargetScope)
		}
	}
}

func TestFieldSpellPackRejectsDescriptorAndFormulaDrift(t *testing.T) {
	tests := []struct {
		name string
		edit func(*FieldSpellDefinition)
	}{
		{"raw mismatch", func(d *FieldSpellDefinition) { d.DescriptorHex = "001e0b" }},
		{"unknown formula", func(d *FieldSpellDefinition) { d.AmountFormula = "battle_half_base" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tt.edit(&p.Spells.Field[0])
			if err := p.validateSpells(); err == nil {
				t.Fatal("drifted field spell definition must fail closed")
			}
		})
	}
}

func TestTemporarySoloChallengeRejectsUnknownTextReference(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	p.Events.TemporarySoloChallenges[0].CaveMessages[0].TextID = "dq3:text.missing"
	if err := p.validateEventTextRefs(); err == nil || !strings.Contains(err.Error(), "references unknown text") {
		t.Fatalf("validateEventTextRefs error=%v, want unknown text reference", err)
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
