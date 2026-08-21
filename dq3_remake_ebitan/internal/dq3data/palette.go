// Package dq3data 保存由原始 DQ3 檔案與可追溯反組譯證據支持的資料 decoder。
// 它不依賴引擎，可獨立測試；歷史 C remake 只能作定位線索，不能當原版 oracle。
package dq3data

// Color is an 8-bit RGB palette entry (matches C dq3_color).
type Color struct{ R, G, B uint8 }

// SelectPaletteBank 回傳原始 palette 的一個完整 bank。它不修改輸入，也不以
// RGB 縮放猜補缺失 bank；任何越界都失敗即關閉。
func SelectPaletteBank(pal []Color, bank, entries int) ([]Color, bool) {
	if bank < 0 || entries <= 0 || len(pal) < entries || bank > (len(pal)-entries)/entries {
		return nil, false
	}
	start := bank * entries
	out := make([]Color, entries)
	copy(out, pal[start:start+entries])
	return out, true
}

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
