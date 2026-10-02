package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldIdleStatusOriginalParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.Interface.FieldIdleStatus
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original EXE identity differs", err)
	}
	if s.DelayTicks != int(binary.LittleEndian.Uint16(exe[0x19940-0xec90+1:])) {
		t.Fatal("idle threshold differs")
	}
	if s.HoldFrames() != 188 || s.RateNumerator != 1193182 || s.RateDenominator != 12428 {
		t.Fatal("hardware-spec conversion differs")
	}
	// Each native SHR halves the already divided maximum. Keep the native
	// instructions and their file mapping as the oracle for the typed values.
	for _, linear := range []int{0x1ef23, 0x1ef2e} {
		if !bytes.Equal(exe[linear-0xec90:linear-0xec90+2], []byte{0xd1, 0xe8}) {
			t.Fatal("native health shift differs")
		}
	}
	if len(s.HealthDivisors) != 2 || s.HealthDivisors[0] != 2 || s.HealthDivisors[1] != 4 ||
		s.HealthSumCap != int(exe[0x1ef3e-0xec90+4]) ||
		s.DeadPaletteIndex != int(exe[0x1ef4e-0xec90+4]) ||
		s.DeadStatusMask != int(binary.LittleEndian.Uint16(exe[0x1ef1b-0xec90+1:])) ||
		s.PaletteIndex*3 != int(exe[0x1ef7a-0xec90+2]) ||
		s.BlankGlyph != int(binary.LittleEndian.Uint16(exe[0x219bf-0xec90+1:])) {
		t.Fatal("native health or leading-blank contract differs")
	}
	if s.Digits != 3 ||
		binary.LittleEndian.Uint16(exe[0x219db-0xec90+1:]) != 100 ||
		binary.LittleEndian.Uint16(exe[0x219e1-0xec90+1:]) != 10 ||
		binary.LittleEndian.Uint16(exe[0x219e7-0xec90+1:]) != 1 {
		t.Fatal("native three-digit consumer differs")
	}
	for i, row := range s.StatusRows {
		off := 0x24dd0 + 0x361a - 0xec90 + i*2
		if row.MaskRaw != int(binary.LittleEndian.Uint16(exe[off:])) {
			t.Fatal("status priority mask differs")
		}
		d, _ := p.TextDefinition(row.TextID)
		if *d.Source.Record != 588+i {
			t.Fatal("status record differs from native 24Ch base")
		}
	}
	raw := exe[0x18716:0x18725]
	for i, color := range s.HealthPaletteRaw {
		for j, c := range color {
			if c != raw[i*3+j] {
				t.Fatal("native health palette differs")
			}
		}
	}
	for _, id := range s.TextIDs() {
		d, _ := p.TextDefinition(id)
		raw, err := os.ReadFile(filepath.Join(dir, d.Source.File))
		if err != nil {
			t.Fatal(err)
		}
		codes := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if len(codes) != len(d.GlyphCodes) {
			t.Fatal("native text length differs")
		}
		for i, code := range codes {
			if int(code) != d.GlyphCodes[i] {
				t.Fatal("native text glyph differs")
			}
		}
	}
	h := p.Interface.PartyHUD
	hpStep := int(exe[0x18263-0xec90+2])
	mpStep := int(exe[0x18270-0xec90+2])
	statusStep := int(exe[0x1828d-0xec90+2])
	if s.HPInsetY != h.TextInsetY+hpStep || s.MPInsetY != s.HPInsetY+mpStep ||
		s.StatusInsetY != s.MPInsetY+statusStep || statusStep != int(exe[0x182af-0xec90+2]) {
		t.Fatal("native party row insets differ")
	}
	if h.X != int(binary.LittleEndian.Uint16(exe[0x17dc5-0xec90+4:]))*8 || h.Y != int(binary.LittleEndian.Uint16(exe[0x17dcb-0xec90+4:])) {
		t.Fatal("native window anchor differs")
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}

func TestFieldIdleStatusRejectsBrokenContracts(t *testing.T) {
	for _, name := range []string{"missing_delay", "null_blank", "missing_hp_row", "null_mp_row", "out_of_row", "unknown_field", "unknown_text", "bad_font", "duplicate_scene", "unreviewed", "bad_rate", "bad_palette", "out_of_canvas", "duplicate_mask", "bad_status_shape"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			s := p.Interface.FieldIdleStatus
			raw, _ := json.Marshal(s)
			var fields map[string]any
			json.Unmarshal(raw, &fields)
			switch name {
			case "missing_delay":
				delete(fields, "delay_ticks")
			case "null_blank":
				fields["blank_glyph"] = nil
			case "missing_hp_row":
				delete(fields, "hp_inset_y")
			case "null_mp_row":
				fields["mp_inset_y"] = nil
			case "out_of_row":
				s.StatusInsetY = p.Interface.PartyHUD.Height
			case "unknown_field":
				fields["guess"] = true
			case "unknown_text":
				fields["column_text_id"] = "missing:text"
			case "bad_font":
				fields["font_asset"] = "missing:font"
			case "bad_rate":
				fields["rate_numerator"] = 0
			case "unreviewed":
				fields["evidence"].(map[string]any)["level"] = "D2"
			case "bad_palette":
				s.HealthPaletteRaw[0][0] = 64
			case "duplicate_scene":
				s.Scenes = append(s.Scenes, s.Scenes[0])
			case "out_of_canvas":
				p.Interface.PartyHUD.Y = 340
			case "duplicate_mask":
				s.StatusRows[1].MaskRaw = s.StatusRows[0].MaskRaw
			case "bad_status_shape":
				p.texts[s.StatusRows[0].TextID].Layout.Columns = 6
			}
			if name == "missing_delay" || name == "null_blank" || name == "missing_hp_row" || name == "null_mp_row" || name == "unknown_field" || name == "unknown_text" || name == "bad_font" || name == "bad_rate" || name == "unreviewed" {
				raw, _ = json.Marshal(fields)
				var changed FieldIdleStatus
				if err = json.Unmarshal(raw, &changed); err != nil {
					return
				}
				p.Interface.FieldIdleStatus = &changed
			}
			if p.validateFieldIdleStatus() == nil && p.validateFieldIdleStatusRefs() == nil {
				t.Fatal("accepted damaged idle contract")
			}
		})
	}
}
