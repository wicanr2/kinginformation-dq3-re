package dq3data

import "testing"

func TestDecodePackBGOriginalPageStrideAndFieldChunk(t *testing.T) {
	scr := findAsset(t, "PACKBG.SCR")
	format := PackBGFormat{
		PageStrideBytes:  0x13d80,
		FieldOffsetBytes: 0x6e00,
		FieldChunkBytes:  0xcf80,
		WidthPixels:      640,
		FieldRows:        166,
		PlaneCount:       4,
	}
	if got, want := len(scr), 46*format.PageStrideBytes; got != want {
		t.Fatalf("PACKBG.SCR bytes=%d, want %d (46 original pages)", got, want)
	}
	for _, tc := range []struct {
		page int
		hist [16]int
	}{
		{0, [16]int{5761, 0, 0, 0, 48833, 22639, 11768, 0, 0, 2720, 0, 1535, 0, 0, 0, 12984}},
		{26, [16]int{5760, 0, 0, 0, 23325, 11120, 0, 24787, 0, 36492, 4756}},
		{35, [16]int{43695, 0, 0, 0, 0, 0, 0, 22039, 0, 4510, 35996}},
	} {
		bg, ok := DecodePackBG(scr, tc.page, format)
		if !ok {
			t.Fatalf("page %d decode failed", tc.page)
		}
		if bg.Width != 640 || bg.Height != 166 || len(bg.Pixels) != 640*166 {
			t.Fatalf("page %d surface=%dx%d pixels=%d", tc.page, bg.Width, bg.Height, len(bg.Pixels))
		}
		var got [16]int
		for _, px := range bg.Pixels {
			if px >= 16 {
				t.Fatalf("page %d index=%d out of 4-plane range", tc.page, px)
			}
			got[px]++
		}
		if got != tc.hist {
			t.Fatalf("page %d histogram=%v, want %v", tc.page, got, tc.hist)
		}
	}
	if _, ok := DecodePackBG(scr, 46, format); ok {
		t.Fatal("page 46 should fail closed: archive has pages 0..45")
	}
	broken := format
	broken.FieldOffsetBytes = broken.PageStrideBytes
	if _, ok := DecodePackBG(scr, 0, broken); ok {
		t.Fatal("invalid field offset should fail closed")
	}
}
