// Package dq3data holds Go ports of the RE'd DQ3 data parsers (from the C remake
// dq3_remake/src/*.c). Pure Go, no engine dependency — buildable & testable headless.
// These translate the *already understood* formats/logic; the C version + docs/ are the oracle.
package dq3data

// Color is an 8-bit RGB palette entry (matches C dq3_color).
type Color struct{ R, G, B uint8 }

// DecodePalette parses a DQ3 .PAL (DQ3.PAL / MNSBK.PAL): consecutive 6-bit VGA RGB
// triplets, expanded to 8-bit via (v<<2)|(v>>4). Port of dq3_pal_decode (dq3_pal.c).
// Returns up to max colors.
func DecodePalette(d []byte, max int) []Color {
	if max <= 0 {
		max = 256
	}
	out := make([]Color, 0, max)
	for i := 0; i+2 < len(d) && len(out) < max; i += 3 {
		r, g, b := d[i], d[i+1], d[i+2]
		out = append(out, Color{
			R: (r << 2) | (r >> 4),
			G: (g << 2) | (g >> 4),
			B: (b << 2) | (b >> 4),
		})
	}
	return out
}
