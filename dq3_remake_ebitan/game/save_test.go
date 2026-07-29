package game

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// 存檔 round-trip:encode → decode 應完全一致。
func TestSaveRoundTrip(t *testing.T) {
	s := saveState{
		HeroExp: 4364, HeroHP: 42, HeroGold: 250,
		Inventory: []int{3, 0x21, 0x1e},
		PX:        12, PY: 28, InTown: true,
		StoryBits: []byte{0x80, 0x40}, DNPhase: 2, DNStep: 17,
		Cty: 44, Section: 1, Layer: 0,
	}
	b, err := encodeSave(s)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := decodeSave(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !reflect.DeepEqual(s, got) {
		t.Fatalf("round-trip 不一致:\n 存 %+v\n 讀 %+v", s, got)
	}
	t.Logf("存檔 round-trip 一致 ✓(exp%d hp%d gold%d inv%v @%d,%d town%v)",
		got.HeroExp, got.HeroHP, got.HeroGold, got.Inventory, got.PX, got.PY, got.InTown)
}

func TestSaveRecordsAndChecksGamePackIdentity(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	g := &Game{pack: pack}
	s := g.snapshot()
	if s.PackID != "dq3_cht" || s.PackSchema != gamepack.SchemaVersion ||
		s.PackContentHash == "" {
		t.Fatalf("snapshot pack metadata incomplete: %#v", s)
	}

	s.PackContentHash = "sha256:modified"
	raw, err := encodeSave(s)
	if err != nil {
		t.Fatal(err)
	}
	name := t.TempDir() + "/save.json"
	if err := os.WriteFile(name, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DQ3_SAVE", name)
	if err := g.Load(); err == nil || !strings.Contains(err.Error(), "game pack mismatch") {
		t.Fatalf("Load mismatch error=%v", err)
	}
}

func TestSavePreservesRememberedOverworldPosition(t *testing.T) {
	g := &Game{
		overPx: 47, overPy: 66,
		px: 7, py: 3, inTown: true, curCty: 2,
		equip: [4]int{-1, -1, -1, -1},
	}
	s := g.snapshot()
	if !s.OverworldPosV2 || s.OverPX != 47 || s.OverPY != 66 {
		t.Fatalf("城內 snapshot 未保存 remembered-world：%+v", s)
	}
	var restored Game
	restored.restore(s)
	if restored.overPx != 47 || restored.overPy != 66 {
		t.Fatalf("城內 round-trip 遺失 remembered-world：(%d,%d)",
			restored.overPx, restored.overPy)
	}

	// 舊 Go 存檔沒有欄位時，依已保存 CTY/layer 的原版 cty_loc 遷移，
	// 不可保留 NewGame 的阿里阿罕座標。
	legacy := saveState{
		PX: 7, PY: 3, InTown: true, Cty: 4, Layer: 0,
		EquipmentV2: true, Equip: [4]int{-1, -1, -1, -1},
	}
	restored.overPx, restored.overPy = 153, 174
	restored.restore(legacy)
	if restored.overPx != ctyLoc[4][0] || restored.overPy != ctyLoc[4][1] {
		t.Fatalf("舊城內存檔未由 CTY04 遷移 remembered-world：(%d,%d)",
			restored.overPx, restored.overPy)
	}
}
