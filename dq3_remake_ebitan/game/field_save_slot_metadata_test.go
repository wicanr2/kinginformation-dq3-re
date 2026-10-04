package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// prepareFieldSaveSlotMetadata writes external JSON fixtures before normal F5.
// Only their visible metadata is equivalent to the original PLAYER.DAT. The
// running game is untouched, and no original saved world is imported.
// Original consumer: IDA9.4 linear116D0/116FE/11707/11711, docs/188.
func prepareFieldSaveSlotMetadata(t *testing.T, g *Game, baseline saveState) []map[string]any {
	t.Helper()
	read := func(name, hash string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(spineAssetsDir(t), name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(b)) != hash {
			t.Fatal("original slot fixture identity differs", name, err)
		}
		return b
	}
	player := read("PLAYER.DAT", "a445a11f52a6711aba1433d9107d10284253e08d27aa6be2b82dc7aec91376dc")
	china := read("CHINA.FON", "42b769cc51857fc14f4e820f937c46a0d254c04374fe52870d59c6eb24f7bb49")
	fon := read("D3TXT00.FON", "c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02")
	const stride, glyphBytes = 20, 32
	if len(player) != g.fieldSaveLoad.contract.SlotCount*stride || len(fon) != dq3data.GlyphMax*glyphBytes {
		t.Fatal("original slot fixture shape differs")
	}
	var reports []map[string]any
	for slot := 0; slot < g.fieldSaveLoad.contract.SlotCount; slot++ {
		record := player[slot*stride : (slot+1)*stride]
		count := int(binary.LittleEndian.Uint16(record))
		if count < 1 || count > g.fieldSaveLoad.contract.NameCapacity {
			t.Fatal("original slot name count differs", slot)
		}
		var name []int
		var rawCodes []uint32
		for i := 0; i < count; i++ {
			code := binary.LittleEndian.Uint32(record[2+i*4:])
			rawCodes = append(rawCodes, code)
			glyph := -1
			if uint16(code) == 0xffff {
				glyph = int(code >> 16)
			} else {
				offset := int(code)
				if offset < 0 || offset+glyphBytes > len(china) {
					t.Fatal("original slot font offset outside CHINA.FON", slot)
				}
				bitmap := china[offset : offset+glyphBytes]
				for candidate := 0; candidate < len(fon)/glyphBytes; candidate++ {
					if bytes.Equal(bitmap, fon[candidate*glyphBytes:(candidate+1)*glyphBytes]) {
						if glyph >= 0 {
							t.Fatal("ambiguous original slot glyph", slot)
						}
						glyph = candidate
					}
				}
			}
			if glyph < 0 || glyph >= dq3data.GlyphMax {
				t.Fatal("original slot glyph absent from runtime font", slot)
			}
			name = append(name, glyph)
		}
		level, rawGender := int(record[18]), int(record[19])
		gender := rawGender - 1 // Original11716..1171B indexes records534/535.
		if gender < 0 || gender >= len(g.fieldSaveLoad.contract.GenderTextIDs) || stats.LevelForExp(g.fieldSaveLoad.contract.PrimaryClassRaw, baseline.HeroExp) != level {
			t.Fatal("fixture cannot represent original level/gender", slot)
		}
		fixture := baseline
		fixture.HeroName, fixture.HeroGender = name, gender
		encoded, err := encodeSave(fixture)
		if err != nil {
			t.Fatal(err)
		}
		decoded, err := decodeSave(encoded)
		if err != nil || !equalFieldSave(fixture, decoded) {
			t.Fatal("slot metadata fixture is not a valid current JSON save", err)
		}
		path := fieldSaveSlotPath(slot)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("refuse to overwrite existing fixture", path, err)
		}
		if err := os.WriteFile(path, encoded, 0644); err != nil {
			t.Fatal(err)
		}
		reports = append(reports, map[string]any{"slot": slot + 1, "player_file_offset": slot * stride, "raw_name_codes": rawCodes, "glyphs": name, "level": level, "gender_raw": rawGender, "gender_logical": gender, "json_sha256": fmt.Sprintf("%x", sha256.Sum256(encoded))})
	}
	return reports
}
