package dq3data

import "testing"

func TestDecodeRowInterleaved4BPP(t *testing.T) {
	// 8×1：四個 plane 的最高位分別設為 1，第一個 pixel 應為 0xf。
	d := []byte{0x80, 0x80, 0x80, 0x80}
	pix, err := DecodeRowInterleaved4BPP(d, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	if pix[0] != 0xf || pix[1] != 0 || pix[7] != 0 {
		t.Fatalf("decoded pixels=%v, want first=0xf and remaining zero", pix)
	}
}

func TestDecodeRowInterleaved4BPPRejectsInvalidSize(t *testing.T) {
	if _, err := DecodeRowInterleaved4BPP(make([]byte, 3), 8, 1); err == nil {
		t.Fatal("short raw screen must fail closed")
	}
	if _, err := DecodeRowInterleaved4BPP(nil, 7, 1); err == nil {
		t.Fatal("non-byte-aligned raw screen must fail closed")
	}
}

func TestDecodeFirstSCRAsset(t *testing.T) {
	d := findAsset(t, "FIRST.SCR")
	pix, err := DecodeRowInterleaved4BPP(d, 640, 350)
	if err != nil {
		t.Fatal(err)
	}
	if len(pix) != 640*350 {
		t.Fatalf("FIRST.SCR pixels=%d, want %d", len(pix), 640*350)
	}
	// The two figures and the black backdrop are stable raw anchors; palette is
	// intentionally supplied by the versioned pack rather than this decoder.
	nonzero := 0
	for _, v := range pix {
		if v != 0 {
			nonzero++
		}
	}
	if pix[0] != 0 || nonzero < 1000 {
		t.Fatalf("FIRST.SCR raw anchors did not decode as expected: nonzero=%d", nonzero)
	}
}
