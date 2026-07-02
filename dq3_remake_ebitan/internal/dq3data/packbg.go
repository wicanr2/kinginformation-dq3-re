package dq3data

// 戰鬥背景 PACKBG.SCR 解碼(移植 dq3_packbg.c,file 0xda55)。
// 每頁 0x6e00 = 88 row × 4 plane × 80 byte(640 寬 row-interleaved planar,MSB-first);
// 索引 0..15 對 MNSBK.PAL 前 16 色。草原 = page 22(對 references/game3.png)。
const (
	PackBGW    = 640
	PackBGH    = 88
	packBGSize = 0x6e00
)

// DecodePackBG 解第 page 頁背景到 [88][640] indexed。越界回 (nil, false)。
func DecodePackBG(scr []byte, page int) (*[PackBGH][PackBGW]uint8, bool) {
	off := page * packBGSize
	if off < 0 || off+packBGSize > len(scr) {
		return nil, false
	}
	buf := scr[off:]
	var out [PackBGH][PackBGW]uint8
	o := 0
	pb := PackBGW / 8 // 80
	for r := 0; r < PackBGH; r++ {
		for pl := 0; pl < 4; pl++ {
			for b := 0; b < pb; b++ {
				var v byte
				if o < packBGSize {
					v = buf[o]
				}
				o++
				for bit := 0; bit < 8; bit++ {
					if v&(0x80>>bit) != 0 {
						out[r][b*8+bit] |= 1 << uint(pl)
					}
				}
			}
		}
	}
	return &out, true
}
