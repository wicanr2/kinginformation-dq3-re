package dq3data

// 自建字形(D3TXT00.FON 缺的 CJK)。移植自 dq3_remake/src/dq3_customglyph.c + .h
// (該檔由 tools/gen_custom_glyphs.py 從 CJK 字型 rasterize,同原版 16×16 MSB-first 格式)。
// 用途:remake 自有 UI 詞(如「設定」的「設」、「原版」的「版」)在原版字庫沒有時補上,不依賴原版 record。

// 自建字形索引(對齊 C dq3_customglyph.h 的 enum DQ3_CG_SHE/DQ3_CG_BAN,即產生順序)。
const (
	CGShe = 0 // 設
	CGBan = 1 // 版
)

// customGlyphBits 是自建字形點陣:16×16 MSB-first row-major,字身 row0..13(row14/15 恆 0)。
// 資料逐 word 對照 dq3_customglyph.c 的 dq3_customglyph_bits[][16],勿手改。
var customGlyphBits = [][16]uint16{
	{0x7df0, 0x0190, 0x7d10, 0x0114, 0x7f1c, 0x0200, 0x7ff8, 0x0108,
		0x7c98, 0x44d0, 0x4460, 0x4470, 0x7d9c, 0x4204, 0x0000, 0x0000}, // 設
	{0x0800, 0x29fc, 0x2900, 0x2900, 0x2900, 0x3ffc, 0x21c4, 0x2148,
		0x7d48, 0x4528, 0x4530, 0x4530, 0x4538, 0x464c, 0x0000, 0x0000}, // 版
}

// CustomGlyph 解自建字形 idx 的 16×16 bitmap。格式與行為對齊 Text.Glyph(MSB-first:
// x=0 對應該列最高位元),供既有 glyph 繪製流程(drawGlyph 系)直接沿用同一套繪製邏輯。
// 回 (bitmap, ok);ok=false 表 idx 越界。
func CustomGlyph(idx int) ([16][16]byte, bool) {
	var g [16][16]byte
	if idx < 0 || idx >= len(customGlyphBits) {
		return g, false
	}
	for y := 0; y < 16; y++ {
		row := customGlyphBits[idx][y]
		for x := 0; x < 16; x++ {
			g[y][x] = byte((row >> (15 - x)) & 1)
		}
	}
	return g, true
}
