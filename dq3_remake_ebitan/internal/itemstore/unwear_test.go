package itemstore

import "testing"

func TestUnwearPartPreservesPhysicalHolesAndOtherParts(t *testing.T) {
	s := fixture(t, []uint16{0xff, 0x9003, 3, 0xff, 0x801e, 0x2001})
	other := s
	if !s.UnwearPart(0) {
		t.Fatal("unwear weapon")
	}
	requireWords(t, s, []uint16{0xff, 0x1003, 3, 0xff, 0x801e, 0x2001})
	requireWords(t, other, []uint16{0xff, 0x9003, 3, 0xff, 0x801e, 0x2001})
	if !s.UnwearPart(1) {
		t.Fatal("unwear armor")
	}
	requireWords(t, s, []uint16{0xff, 0x1003, 3, 0xff, 30, 0x2001})
	if !s.UnwearPart(1) {
		t.Fatal("empty part must be idempotent")
	}
	for _, p := range []int{-1, 4} {
		if s.UnwearPart(p) {
			t.Fatal("invalid part accepted")
		}
	}
}
func TestUnwearPartCursedFailureIsAtomic(t *testing.T) {
	s := fixture(t, []uint16{0x8003, 0xc01b, 0xff, 0x801e})
	before := s.Words()
	if s.UnwearPart(0) {
		t.Fatal("cursed part accepted")
	}
	requireWords(t, s, before)
	if !s.UnwearPart(1) {
		t.Fatal("unrelated part blocked")
	}
	requireWords(t, s, []uint16{0x8003, 0xc01b, 0xff, 30})
	var empty Store
	if empty.UnwearPart(0) {
		t.Fatal("uninitialized accepted")
	}
}
