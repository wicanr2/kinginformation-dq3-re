package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestOpeningSceneRejectsInvalidContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*OpeningScenePresentation)
	}{
		{"empty id", func(e *OpeningScenePresentation) { e.ID = "" }},
		{"wrong presentation", func(e *OpeningScenePresentation) { e.PresentationID = "unknown" }},
		{"wrong scene", func(e *OpeningScenePresentation) { e.Section++ }},
		{"no texts", func(e *OpeningScenePresentation) { e.TextIDs = nil }},
		{"duplicate text", func(e *OpeningScenePresentation) { e.TextIDs[1] = e.TextIDs[0] }},
		{"unknown camera", func(e *OpeningScenePresentation) { e.Camera.Mode = "unknown" }},
		{"bad anchor", func(e *OpeningScenePresentation) { e.Camera.AnchorX = 20 }},
		{"bad tile", func(e *OpeningScenePresentation) { e.Camera.ExteriorTile = 256 }},
		{"unknown shadow", func(e *OpeningScenePresentation) { e.Shadow.Mode = "unknown" }},
		{"unaligned shadow", func(e *OpeningScenePresentation) { e.Shadow.OffsetX = 1 }},
		{"shadow overflow", func(e *OpeningScenePresentation) { e.Shadow.OffsetY = 350 }},
		{"unreviewed camera", func(e *OpeningScenePresentation) { e.Camera.Evidence.Level = "D2" }},
		{"unreviewed shadow", func(e *OpeningScenePresentation) { e.Shadow.Evidence.Level = "D2" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(p.Interface.OpeningScenePresentation)
			if err := p.validateOpeningScenePresentation(); err == nil {
				t.Fatal("接受錯誤契約")
			}
		})
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	p.Interface.OpeningScenePresentation.TextIDs[0] = "unknown:text"
	if err := p.validateOpeningSceneRefs(); err == nil {
		t.Fatal("接受未知文字引用")
	}
	p.Interface.OpeningScenePresentation = nil
	if err := p.validateOpeningScenePresentation(); err == nil {
		t.Fatal("缺少場景契約卻猜補預設")
	}
}

func TestOpeningWaitIndicatorRejectsBrokenContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*OpeningWaitIndicator)
	}{
		{"unknown mode", func(w *OpeningWaitIndicator) { w.Mode = "unknown" }},
		{"left outside text", func(w *OpeningWaitIndicator) { w.X = 0 }},
		{"right outside text", func(w *OpeningWaitIndicator) { w.X = 640 }},
		{"unaligned x", func(w *OpeningWaitIndicator) { w.X++ }},
		{"negative glyph", func(w *OpeningWaitIndicator) { w.VisibleGlyph = -1 }},
		{"glyph outside font", func(w *OpeningWaitIndicator) { w.HiddenGlyph = dq3data.GlyphMax }},
		{"identical phases", func(w *OpeningWaitIndicator) { w.HiddenGlyph = w.VisibleGlyph }},
		{"zero visible ticks", func(w *OpeningWaitIndicator) { w.VisibleTicks = 0 }},
		{"zero hidden ticks", func(w *OpeningWaitIndicator) { w.HiddenTicks = 0 }},
		{"oversized ticks", func(w *OpeningWaitIndicator) { w.VisibleTicks = 65536 }},
		{"zero numerator", func(w *OpeningWaitIndicator) { w.RateNumerator = 0 }},
		{"zero denominator", func(w *OpeningWaitIndicator) { w.RateDenominator = 0 }},
		{"oversized frequency", func(w *OpeningWaitIndicator) { w.RateNumerator = 1000000001 }},
		{"oversized divisor", func(w *OpeningWaitIndicator) { w.RateDenominator = 1000000001 }},
		{"unbounded duration", func(w *OpeningWaitIndicator) { w.RateNumerator = 1 }},
		{"unreviewed behavior", func(w *OpeningWaitIndicator) { w.Evidence.Level = "D2" }},
		{"missing timing source", func(w *OpeningWaitIndicator) { w.TimingEvidence.Source = "" }},
		{"unknown timing", func(w *OpeningWaitIndicator) { w.TimingEvidence.Level = "D1" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(&p.Interface.OpeningPrelude.WaitIndicator)
			if err := p.validateInterface(); err == nil {
				t.Fatal("損壞的等待指示契約未拒絕")
			}
		})
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(p.Interface.OpeningPrelude.WaitIndicator)
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatal(err)
	}
	for name := range fields {
		for _, mode := range []string{"missing", "null"} {
			t.Run(mode+" "+name, func(t *testing.T) {
				copyFields := map[string]json.RawMessage{}
				for k, v := range fields {
					copyFields[k] = v
				}
				if mode == "missing" {
					delete(copyFields, name)
				} else {
					copyFields[name] = json.RawMessage("null")
				}
				b, err := json.Marshal(copyFields)
				if err != nil {
					t.Fatal(err)
				}
				var w OpeningWaitIndicator
				if err := json.Unmarshal(b, &w); err == nil {
					t.Fatal("接受缺少或null欄位")
				}
			})
		}
	}
	fields["unknown"] = json.RawMessage("1")
	b, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	var w OpeningWaitIndicator
	if err := json.Unmarshal(b, &w); err == nil {
		t.Fatal("接受未知欄位")
	}
}

func TestOpeningWaitIndicatorMatchesOriginalBytes(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	w := p.Interface.OpeningPrelude.WaitIndicator
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("原版EXE身份不符")
	}
	// IDA linear=file+EC90。全部沿用原始運算元，不用導覽名稱代替bytes。
	for off, want := range map[int][]byte{
		0x12a35: {0xbd, 0x2a, 0x00}, 0x12a38: {0xbb, 0x0d, 0x00},
		0x12a56: {0x3d, 0x08, 0x00}, 0x12a5b: {0xbb, 0x0c, 0x00},
		0x12a79: {0x3d, 0x05, 0x00}, 0x12a80: {0xbb, 0x0c, 0x00},
	} {
		if !bytes.Equal(exe[off:off+len(want)], want) {
			t.Fatalf("file%#x原始bytes不符", off)
		}
	}
	word := func(off int) int { return int(binary.LittleEndian.Uint16(exe[off : off+2])) }
	if w.X != word(0x12a36)*8 || w.VisibleGlyph != word(0x12a39) || w.HiddenGlyph != word(0x12a5c) ||
		w.HiddenGlyph != word(0x12a81) || w.VisibleTicks != word(0x12a57) || w.HiddenTicks != word(0x12a7a) {
		t.Fatal("等待指示JSON與原始consumer不符")
	}
	fon, err := os.ReadFile(filepath.Join(dir, "D3TXT00.FON"))
	if err != nil {
		t.Fatal(err)
	}
	if len(fon) != 47232 || fmt.Sprintf("%x", sha256.Sum256(fon)) != "c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02" {
		t.Fatal("原版字模身份不符")
	}
	if !bytes.Equal(fon[w.HiddenGlyph*32:(w.HiddenGlyph+1)*32], make([]byte, 32)) {
		t.Fatal("清除字模不是空白")
	}
	n := 0
	for _, b := range fon[w.VisibleGlyph*32 : (w.VisibleGlyph+1)*32] {
		for b != 0 {
			n += int(b & 1)
			b >>= 1
		}
	}
	if n != 41 {
		t.Fatal("箭頭原始字模像素數不符")
	}
	if w.RateNumerator != 315000000 || w.RateDenominator != 264*12428 || w.HoldFrames(w.VisibleTicks) != 5 || w.HoldFrames(w.HiddenTicks) != 4 {
		t.Fatal("已審查平台時序近似不符")
	}
}

func TestOpeningSceneRejectsMissingNullAndUnknownFields(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	objects := []any{p.Interface.OpeningScenePresentation, p.Interface.OpeningScenePresentation.Camera, p.Interface.OpeningScenePresentation.Shadow}
	for n, object := range objects {
		raw, err := json.Marshal(object)
		if err != nil {
			t.Fatal(err)
		}
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(raw, &fields); err != nil {
			t.Fatal(err)
		}
		decode := func(b []byte) error {
			switch n {
			case 0:
				var v OpeningScenePresentation
				return json.Unmarshal(b, &v)
			case 1:
				var v SceneCamera
				return json.Unmarshal(b, &v)
			default:
				var v WindowShadow
				return json.Unmarshal(b, &v)
			}
		}
		for key, value := range fields {
			delete(fields, key)
			b, _ := json.Marshal(fields)
			if err := decode(b); err == nil {
				t.Fatalf("object%d接受缺少%s", n, key)
			}
			fields[key] = json.RawMessage("null")
			b, _ = json.Marshal(fields)
			if err := decode(b); err == nil {
				t.Fatalf("object%d接受null %s", n, key)
			}
			fields[key] = value
		}
		fields["unknown"] = json.RawMessage("1")
		b, _ := json.Marshal(fields)
		if err := decode(b); err == nil {
			t.Fatal("接受未知欄位")
		}
	}
}

func TestOpeningSceneMatchesOriginalData(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	e, ok := p.OpeningScenePresentation()
	if !ok {
		t.Fatal("缺少正式房間契約")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("EXE身份不符")
	}
	for _, a := range []struct {
		off int
		raw []byte
	}{
		{0x2ce1, []byte{0xa1, 0x33, 0x4f, 0x2d, 0x09, 0x00, 0xa3, 0x25, 0x4f}},
		{0x2cea, []byte{0x8b, 0x1e, 0x35, 0x4f, 0x83, 0xeb, 0x07, 0x89, 0x1e, 0x27, 0x4f}},
		{0x10fd5, []byte{0x8b, 0x7c, 0x02, 0x47, 0x03, 0x3e, 0x09, 0x4f}},
		{0x10fdd, []byte{0x8b, 0x44, 0x04, 0x05, 0x08, 0x00}},
		{0x11007, []byte{0xbd, 0xaa, 0xaa}},
		{0x1100c, []byte{0xd1, 0xcd}},
		{0x11015, []byte{0x26, 0x21, 0x2d, 0x47, 0x47}},
	} {
		if !bytes.Equal(exe[a.off:a.off+len(a.raw)], a.raw) {
			t.Fatalf("原始consumer file%#x不符", a.off)
		}
	}
	if e.Camera.AnchorX != int(binary.LittleEndian.Uint16(exe[0x2ce5:0x2ce7])) || e.Camera.AnchorY != int(exe[0x2cf0]) ||
		e.Shadow.OffsetX != 8 || e.Shadow.OffsetY != int(binary.LittleEndian.Uint16(exe[0x10fe1:0x10fe3])) {
		t.Fatal("契約與原始camera／shadow consumer不符")
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY00.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cty) != 7546 || fmt.Sprintf("%x", sha256.Sum256(cty)) != "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836" {
		t.Fatal("CTY身份不符")
	}
	base := int(binary.LittleEndian.Uint16(cty[e.Section*2 : e.Section*2+2]))
	if e.CTY != 0 || base != 0x1383 || e.Camera.ExteriorTile != int(cty[base+0x12]) {
		t.Fatal("section界外圖塊與raw欄位不符")
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT01.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	if len(e.TextIDs) != 2 {
		t.Fatal("原始caller對白數不符")
	}
	for i, rec := range []int{83, 81} {
		codes, ok := p.TextGlyphCodes(e.TextIDs[i])
		if !ok || !reflect.DeepEqual(codes, dq3data.LoadText(nil, txt).Record(rec)) {
			t.Fatalf("record%d原始詞流不符", rec)
		}
	}
}
