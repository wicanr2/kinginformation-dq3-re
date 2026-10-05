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
	g := &Game{pack: loadTestPack(t), items: testItemStore([]int{3, 0x21, 0x1e}, [4]int{-1, -1, -1, -1}), heroExp: 4364, heroHP: 42, heroGold: 250, heroConditions: conditionPoison | conditionParalysis, paralysisSteps: 17, px: 12, py: 28}
	g.companions = []*Member{newMember([]int{1}, 1, 0, 0)}
	g.companions[0].Conditions = conditionPoison
	s := g.snapshot()
	b, err := encodeSave(s)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	got, err := decodeSave(b)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := g.validateSavedItemStores(&got); err != nil {
		t.Fatal(err)
	}
	a, _ := encodeSave(s)
	z, _ := encodeSave(got)
	if string(a) != string(z) || !reflect.DeepEqual(s.itemStore.Words(), got.itemStore.Words()) {
		t.Fatalf("round-trip 不一致:\n 存 %+v\n 讀 %+v", s, got)
	}
	t.Logf("存檔 round-trip 一致 ✓(exp%d hp%d gold%d inv%v @%d,%d town%v)",
		got.HeroExp, got.HeroHP, got.HeroGold, testInventory(got.itemStore), got.PX, got.PY, got.InTown)
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
	g := &Game{pack: loadTestPack(t), overPx: 47, overPy: 66, px: 7, py: 3, inTown: true, curCty: 2, items: testItemStore(nil, [4]int{-1, -1, -1, -1})}
	s := g.snapshot()
	if !s.OverworldPosV2 || s.OverPX != 47 || s.OverPY != 66 {
		t.Fatalf("城內 snapshot 未保存 remembered-world：%+v", s)
	}
	restored := Game{pack: loadTestPack(t)}
	if err := restored.restore(s); err != nil {
		t.Fatal(err)
	}
	if restored.overPx != 47 || restored.overPy != 66 {
		t.Fatalf("城內 round-trip 遺失 remembered-world：(%d,%d)",
			restored.overPx, restored.overPy)
	}

	if _, err := decodeSave([]byte(`{"px":7,"py":3,"town":true,"cty":4}`)); err == nil {
		t.Fatal("unversioned save accepted")
	}
}
