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

// dnScale 回傳晝夜相位對應的 RGB 縮放百分比(移植 dq3_scene.c dn_scale)。
// 0=白天(原色) 1=黃昏(暗偏暖) 2=黑夜(暗偏藍) 3=黎明(微暗偏冷)。
func dnScale(phase int) (r, g, b int) {
	switch phase & 3 {
	case 2:
		return 42, 44, 70 // 黑夜
	case 1:
		return 82, 62, 58 // 黃昏
	case 3:
		return 72, 74, 92 // 黎明
	default:
		return 100, 100, 100 // 白天:原色
	}
}

// DarkenPalette 依晝夜相位回傳調暗後的新色盤(移植 dq3_scene_apply_palette + dn_tint,
// dq3_scene.c:479-503)。bg 色盤(slot 0..15)以相位縮放全調暗;sprite 色盤 bank(16..31)
// 用「淺一半」的暗化以保角色夜間仍可見;其餘 slot 原樣複製。phase 0 為原色(identity)。
// 回傳恆為新 slice,不動輸入(呼叫端傳日中 base pal,可重覆套用不累積)。
func DarkenPalette(pal []Color, phase int) []Color {
	rs, gs, bs := dnScale(phase)
	srs, sgs, sbs := (rs+100)/2, (gs+100)/2, (bs+100)/2 // sprite 較淺暗化
	out := make([]Color, len(pal))
	copy(out, pal)
	for i := 0; i < len(out) && i < 32; i++ {
		r, g, b := rs, gs, bs
		if i >= 16 {
			r, g, b = srs, sgs, sbs
		}
		out[i] = Color{
			R: uint8(int(out[i].R) * r / 100),
			G: uint8(int(out[i].G) * g / 100),
			B: uint8(int(out[i].B) * b / 100),
		}
	}
	return out
}
