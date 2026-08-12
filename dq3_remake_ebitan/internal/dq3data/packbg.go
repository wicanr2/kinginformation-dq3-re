package dq3data

// PackBGFormat is a pack-owned planar archive layout.  It contains no
// edition-specific fallback: production passes the strictly validated values
// from data/battle.json.
type PackBGFormat struct {
	PageStrideBytes  int
	FieldOffsetBytes int
	FieldChunkBytes  int
	WidthPixels      int
	FieldRows        int
	PlaneCount       int
}

// PackBG is one indexed field surface decoded from a legacy planar archive.
// Palette-bank selection remains separate because original formations choose
// it independently of the archive page.
type PackBG struct {
	Width  int
	Height int
	Pixels []uint8
}

// DecodePackBG decodes the pack-selected field chunk for one background page.
// The source is row-interleaved planar data, MSB first.  Malformed layouts or
// out-of-range selectors fail closed instead of selecting a nearby page.
func DecodePackBG(scr []byte, page int, format PackBGFormat) (*PackBG, bool) {
	rowBytes := format.WidthPixels / 8
	wantChunk := rowBytes * format.FieldRows * format.PlaneCount
	if page < 0 || format.PageStrideBytes <= 0 || format.FieldOffsetBytes < 0 ||
		format.FieldChunkBytes <= 0 || format.WidthPixels <= 0 ||
		format.WidthPixels%8 != 0 || format.FieldRows <= 0 ||
		format.PlaneCount < 1 || format.PlaneCount > 8 ||
		format.FieldChunkBytes != wantChunk ||
		format.FieldOffsetBytes+format.FieldChunkBytes > format.PageStrideBytes {
		return nil, false
	}
	if page > (len(scr)-format.FieldOffsetBytes-format.FieldChunkBytes)/format.PageStrideBytes {
		return nil, false
	}
	off := page*format.PageStrideBytes + format.FieldOffsetBytes
	if off < 0 || off+format.FieldChunkBytes > len(scr) {
		return nil, false
	}
	out := &PackBG{
		Width:  format.WidthPixels,
		Height: format.FieldRows,
		Pixels: make([]uint8, format.WidthPixels*format.FieldRows),
	}
	buf := scr[off : off+format.FieldChunkBytes]
	p := 0
	for y := 0; y < out.Height; y++ {
		for plane := 0; plane < format.PlaneCount; plane++ {
			for b := 0; b < rowBytes; b++ {
				v := buf[p]
				p++
				for bit := 0; bit < 8; bit++ {
					if v&(0x80>>bit) != 0 {
						out.Pixels[y*out.Width+b*8+bit] |= 1 << uint(plane)
					}
				}
			}
		}
	}
	return out, true
}
