package game

import (
	"reflect"
	"testing"
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
