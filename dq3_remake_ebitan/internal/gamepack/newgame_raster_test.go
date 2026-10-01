package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 原始格式 oracle：直接讀 EXE 的 byte 座標與 TXT 的完整記錄，不能用另一份 Go 表代替。
func TestNewGameRasterOriginalDataParity(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	exe, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	txt, err := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "D3TXT00.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []struct {
		data []byte
		hash string
	}{
		{exe, "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c"},
		{txt, "38d7f9b8d79b5c7fed9dc9692c9f477bb828a2b8707b1e0cfb29a6e5e70c8a2b"},
	} {
		if fmt.Sprintf("%x", sha256.Sum256(source.data)) != source.hash {
			t.Fatal("原始輸入雜湊不符")
		}
	}
	g := p.Interface.NewGameGeometry
	r := g.Raster
	raw := map[string]RawNewGameWindow{}
	for _, w := range g.RawWindows {
		raw[w.ID] = w
	}
	for _, ref := range []RasterWindowRef{r.Menu, r.Header, r.Mode, r.Gender} {
		w := raw[ref.RawWindowID]
		linear, err := strconv.ParseInt(strings.TrimPrefix(w.Address, "linear:"), 0, 32)
		if err != nil {
			t.Fatal(err)
		}
		off := int(linear) - 0xec90 // IDA linear → MZ file；docs/113。
		if off < 0 || off+30 > len(exe) {
			t.Fatal("原始視窗範圍錯")
		}
		word := func(n int) int { return int(binary.LittleEndian.Uint16(exe[off+n:])) }
		if w.Flags != int(exe[off+1]) || w.X != word(2) || w.Y != word(4) || w.Width != word(6) || w.Height != word(8) {
			t.Fatalf("%s 與 EXE(file %#x) 不符：%+v", w.ID, off, w)
		}
		if record := p.texts[ref.TextID].Source.Record; record == nil || *record != word(10) {
			t.Fatalf("%s 的原始文字引用不符", w.ID)
		}
		if ref == r.Menu && (r.MenuCursor.X != word(24)*8 || r.MenuCursor.Y != word(26)) {
			t.Fatal("主選單游標與原始結構不符")
		}
		if ref == r.Menu && r.MenuHit.Width != (word(6)-4)*8 {
			t.Fatal("主選單命中寬度與原始結構／sub_1F908 consumer 不符")
		}
		if ref == r.Gender && (r.GenderCursor.X != word(24)*8 || r.GenderCursor.Y != word(26) ||
			r.GenderHit.X != word(24)*8 || r.GenderHit.Y != word(26) || r.GenderHit.Width != (word(6)-4)*8 ||
			r.GenderHit.Height != 16 || r.GenderCursor.StepY != 16) {
			t.Fatal("性別游標／命中列與原始結構及選項 consumer 不符")
		}
	}
	// 測試專用原始定位，保留 IDA linear／MZ file 換算；不進 production renderer。
	functionOffset := 0x290a6 - 0xec90
	functionWord := func(n int) int { return int(binary.LittleEndian.Uint16(exe[functionOffset+n:])) }
	if r.FunctionCursor.X != functionWord(24)*8 || r.FunctionCursor.Y != functionWord(26) ||
		r.FunctionHit.Width != (functionWord(6)-4)*8 {
		t.Fatal("功能列游標／命中寬度與原始結構不符")
	}
	decoded := dq3data.LoadText(nil, txt)
	for _, id := range []string{r.Menu.TextID, r.Header.TextID, r.Mode.TextID, r.Gender.TextID, r.ZhuyinTextID, r.AlnumTextID} {
		d := p.texts[id]
		codes, ok := p.TextGlyphCodes(id)
		if !ok || d.Source.Record == nil {
			t.Fatalf("%s 缺原始記錄引用", id)
		}
		if !reflect.DeepEqual(codes, decoded.Record(*d.Source.Record)) {
			t.Fatalf("%s 完整字模記錄不符", id)
		}
		if d.Evidence.Level != "D3" {
			t.Fatalf("%s 未達 D3", id)
		}
	}
}

func TestNewGameRasterRejectsInvalidPack(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Pack)
	}{
		{"missing", func(p *Pack) { p.Interface.NewGameGeometry.Raster = nil }},
		{"font", func(p *Pack) { p.Interface.NewGameGeometry.Raster.FontIndex = nil }},
		{"window", func(p *Pack) { p.Interface.NewGameGeometry.Raster.Header.RawWindowID = "unknown" }},
		{"gender_window", func(p *Pack) { p.Interface.NewGameGeometry.Raster.Gender.RawWindowID = "unknown" }},
		{"gender_text", func(p *Pack) { p.Interface.NewGameGeometry.Raster.Gender.TextID = "unknown" }},
		{"gender_cursor", func(p *Pack) { p.Interface.NewGameGeometry.Raster.GenderCursor.StepY = 0 }},
		{"gender_hit", func(p *Pack) { p.Interface.NewGameGeometry.Raster.GenderHit.Width = 0 }},
		{"gender_last_row", func(p *Pack) { p.Interface.NewGameGeometry.Raster.GenderCursor.Y = 330 }},
		{"text", func(p *Pack) { p.Interface.NewGameGeometry.Raster.AlnumTextID = "unknown" }},
		{"mask", func(p *Pack) { *p.Interface.NewGameGeometry.Raster.CursorXOR = 16 }},
		{"palette", func(p *Pack) { p.Interface.NewGameGeometry.Raster.PaletteOverrides[0].RGB = []uint8{1} }},
		{"duplicate_palette", func(p *Pack) {
			r := p.Interface.NewGameGeometry.Raster
			r.PaletteOverrides = append(r.PaletteOverrides, r.PaletteOverrides[0])
		}},
		{"shadow_bounds", func(p *Pack) { p.Interface.NewGameGeometry.Raster.ShadowOffset.X = 640 }},
		{"grid_bounds", func(p *Pack) { p.Interface.NewGameGeometry.NameGrid.X = 630 }},
		{"text_control", func(p *Pack) { p.texts[p.Interface.NewGameGeometry.Raster.Menu.TextID].GlyphCodes[0] = 0xfffd }},
		{"text_shape", func(p *Pack) { p.texts[p.Interface.NewGameGeometry.Raster.Menu.TextID].GlyphCodes = []int{1} }},
		{"evidence", func(p *Pack) { p.Interface.NewGameGeometry.Raster.Evidence.Level = "D2" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if err := p.validateNewGameRasterRefs(); err != nil {
				t.Fatalf("正對照失敗：%v", err)
			}
			tc.mutate(p)
			if err := p.validateNewGameRasterRefs(); err == nil {
				t.Fatal("損壞的繪圖契約未被拒絕")
			}
		})
	}
}
