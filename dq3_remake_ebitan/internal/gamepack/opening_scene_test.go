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

func TestOpeningArrivalMatchesOriginalData(t *testing.T) {
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	e := p.Interface.OpeningEscort
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "..", "assets_raw")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("EXE 身分不符")
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY00.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if len(cty) != 7546 || fmt.Sprintf("%x", sha256.Sum256(cty)) != "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836" {
		t.Fatal("CTY 身分不符")
	}
	home, err := dq3data.OpenTown(cty, e.Section, false)
	if err != nil {
		t.Fatal(err)
	}
	landing := home.Transitions[0] // 原版 selector1 的第一個四-byte transition。
	if e.Destination != (SceneCoordinate{CTY: landing[0], Section: landing[1]}) {
		t.Fatal("原始轉場引用不符")
	}
	town, err := dq3data.OpenTown(cty, e.Destination.Section, false)
	if err != nil {
		t.Fatal(err)
	}
	if e.ArrivalLeaderRecord == nil || *e.ArrivalLeaderRecord != 0 {
		t.Fatal("原版 DX=0 actor 身分不符")
	}
	npc := town.NPCs[*e.ArrivalLeaderRecord]
	state := OpeningArrivalFrame{Player: TileCoordinate{X: landing[2], Y: landing[3]}, Leader: TileCoordinate{X: npc.X, Y: npc.Y}, LeaderFacing: 0}
	expected := []OpeningArrivalFrame{state}
	move := func(pos *TileCoordinate, raw int) {
		switch raw {
		case 0:
			pos.Y++
		case 1:
			pos.X--
		case 2:
			pos.Y--
		case 3:
			pos.X++
		default:
			t.Fatalf("未知原始方向 %d", raw)
		}
	}
	engineFacing := []int{0, 2, 1, 3}
	// IDA 9.4 已審查 caller：CX count、AX direction、DX actor0、CL mode0。
	// 數量與方向由原始 bytes 讀出，沒有從 production 路徑反推 oracle。
	for _, linear := range []int{0x10130, 0x1014b, 0x10166, 0x10181, 0x1019c} {
		o := linear - 0xec90
		if exe[o] != 0xb9 || exe[o+3] != 0x51 || exe[o+4] != 0xb8 || !bytes.Equal(exe[o+7:o+12], []byte{0xba, 0, 0, 0xb1, 0}) || !bytes.Equal(exe[o+15:o+19], []byte{0xc7, 6, 0x1f, 0x4f}) {
			t.Fatalf("loop bytes 不符 IDA linear %#x", linear)
		}
		count := int(binary.LittleEndian.Uint16(exe[o+1 : o+3]))
		raw := int(binary.LittleEndian.Uint16(exe[o+5 : o+7]))
		if raw != int(binary.LittleEndian.Uint16(exe[o+19:o+21])) {
			t.Fatal("NPC 與主角 direction writer 不符")
		}
		for i := 0; i < count; i++ {
			move(&state.Player, raw)
			move(&state.Leader, raw)
			state.LeaderFacing = engineFacing[raw]
			expected = append(expected, state)
		}
	}
	o := 0x101b7 - 0xec90
	if !bytes.Equal(exe[o:o+8], []byte{0xb8, 0, 0, 0xba, 0, 0, 0xb1, 2}) {
		t.Fatal("原始 NPC turn-only caller 不符")
	}
	state.LeaderFacing = engineFacing[int(binary.LittleEndian.Uint16(exe[o+1:o+3]))]
	expected = append(expected, state)
	if e.DialogueFrameIndex != len(expected)-1 {
		t.Fatal("對話應在原始 turn-only 後開啟")
	}
	for _, linear := range []int{0x101dd, 0x101e6, 0x101ef} {
		o := linear - 0xec90
		if !bytes.Equal(exe[o:o+4], []byte{0xc7, 6, 0x1f, 0x4f}) {
			t.Fatal("後三步原始方向 writer 不符")
		}
		move(&state.Player, int(binary.LittleEndian.Uint16(exe[o+4:o+6])))
		expected = append(expected, state)
	}
	if len(expected) != len(e.ArrivalFrames) {
		t.Fatal("帶路 frame count 不符原始迴圈")
	}
	for i, want := range expected {
		got := e.ArrivalFrames[i]
		if got.Player != want.Player || got.Leader != want.Leader || got.LeaderFacing != want.LeaderFacing {
			t.Fatalf("frame%d 狀態=%+v want%+v", i, got, want)
		}
	}
	o = 0x101f8 - 0xec90
	if !reflect.DeepEqual(e.SetStoryFlags, []int{int(binary.LittleEndian.Uint16(exe[o+1 : o+3]))}) || !reflect.DeepEqual(e.ClearStoryFlags, []int{int(binary.LittleEndian.Uint16(exe[o+7 : o+9]))}) {
		t.Fatal("旗標 caller 不符")
	}
	o = 0x101cb - 0xec90
	if exe[o] != 0xbf {
		t.Fatal("文字 selector caller 不符")
	}
	record := int(binary.LittleEndian.Uint16(exe[o+1:o+3])) - 0x0bb8
	if !reflect.DeepEqual(e.DialogueRecords, []int{record}) || len(e.DialogueTextIDs) != 1 {
		t.Fatal("原版只有單一文字 consumer")
	}
	txt, err := os.ReadFile(filepath.Join(dir, "D3TXT01.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	codes, ok := p.TextGlyphCodes(e.DialogueTextIDs[0])
	if !ok || !reflect.DeepEqual(codes, dq3data.LoadText(nil, txt).Record(record)) {
		t.Fatal("城門文字與原始 record 不符")
	}
	t.Log("42個狀態、原始NPC身分、單一文字與旗標引用吻合；hold_frames及動畫不在此驗證範圍")
}

func TestOpeningArrivalRejectsInvalidContract(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*OpeningEscort)
	}{
		{"missing actor", func(e *OpeningEscort) { e.ArrivalLeaderRecord = nil }},
		{"negative actor", func(e *OpeningEscort) { v := -1; e.ArrivalLeaderRecord = &v }},
		{"unknown presentation", func(e *OpeningEscort) { e.DialoguePresentationID = "unknown" }},
		{"unreviewed arrival", func(e *OpeningEscort) { e.ArrivalEvidence.Level = "D2" }},
		{"missing text ids", func(e *OpeningEscort) { e.DialogueTextIDs = nil }},
		{"invalid facing", func(e *OpeningEscort) { e.ArrivalFrames[0].LeaderFacing = 4 }},
		{"leader teleport", func(e *OpeningEscort) { e.ArrivalFrames[1].Leader.X += 4 }},
		{"idle frame", func(e *OpeningEscort) { e.ArrivalFrames[1] = e.ArrivalFrames[0] }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			tc.edit(p.Interface.OpeningEscort)
			if p.validateInterface() == nil {
				t.Fatal("接受錯誤帶路契約")
			}
		})
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	p.Interface.OpeningEscort.DialogueTextIDs[0] = "unknown:text"
	if p.validateOpeningEscortTextRefs() == nil {
		t.Fatal("接受未知文字引用")
	}
	raw, err := json.Marshal(p.Interface.OpeningEscort.ArrivalFrames[0])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err = json.Unmarshal(raw, &fields); err != nil {
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
				b, _ := json.Marshal(copyFields)
				var f OpeningArrivalFrame
				if json.Unmarshal(b, &f) == nil {
					t.Fatal("接受缺少或null欄位")
				}
			})
		}
	}
	for _, actor := range []string{"player", "leader"} {
		for _, axis := range []string{"x", "y"} {
			for _, mode := range []string{"missing", "null"} {
				t.Run(mode+" "+actor+"."+axis, func(t *testing.T) {
					var copyFields map[string]json.RawMessage
					if err := json.Unmarshal(raw, &copyFields); err != nil {
						t.Fatal(err)
					}
					var coordinate map[string]json.RawMessage
					if err := json.Unmarshal(copyFields[actor], &coordinate); err != nil {
						t.Fatal(err)
					}
					if mode == "missing" {
						delete(coordinate, axis)
					} else {
						coordinate[axis] = json.RawMessage("null")
					}
					copyFields[actor], _ = json.Marshal(coordinate)
					b, _ := json.Marshal(copyFields)
					var f OpeningArrivalFrame
					if json.Unmarshal(b, &f) == nil {
						t.Fatal("接受缺少或null座標")
					}
				})
			}
		}
	}
}
