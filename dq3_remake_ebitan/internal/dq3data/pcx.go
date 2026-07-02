package dq3data

import "fmt"

// TIT*.P 標題畫面解碼(ZSoft PCX,640×350,4 plane,RLE)。移植 dq3_pcx.c。
// planar 4bpp → indexed;檔內 16 色 palette(header offset 16,已 0..255)。

// DecodePCX 解 PCX → (indexed 像素 W*H, 16 色 palette, W, H)。
func DecodePCX(d []byte) (pix []uint8, pal []Color, w, h int, err error) {
	if len(d) < 128 || d[0] != 0x0A {
		return nil, nil, 0, 0, fmt.Errorf("非 PCX")
	}
	xmin := int(d[4]) | int(d[5])<<8
	ymin := int(d[6]) | int(d[7])<<8
	xmax := int(d[8]) | int(d[9])<<8
	ymax := int(d[10]) | int(d[11])<<8
	w, h = xmax-xmin+1, ymax-ymin+1
	nplanes := int(d[65])
	bpl := int(d[66]) | int(d[67])<<8
	if w <= 0 || h <= 0 || nplanes != 4 {
		return nil, nil, 0, 0, fmt.Errorf("尺寸/plane 不合(%d×%d,%d plane)", w, h, nplanes)
	}
	pal = make([]Color, 16) // header offset 16:16 色 EGA palette
	for i := 0; i < 16; i++ {
		pal[i] = Color{R: d[16+i*3], G: d[16+i*3+1], B: d[16+i*3+2]}
	}
	rowBytes := nplanes * bpl
	pix = make([]uint8, w*h)
	line := make([]uint8, rowBytes)
	ip := 128
	for row := 0; row < h; row++ {
		filled := 0
		for filled < rowBytes && ip < len(d) { // RLE 解一整列
			b := d[ip]
			ip++
			if b&0xC0 == 0xC0 {
				cnt := int(b & 0x3F)
				var v byte
				if ip < len(d) {
					v = d[ip]
					ip++
				}
				for cnt > 0 && filled < rowBytes {
					line[filled] = v
					filled++
					cnt--
				}
			} else {
				line[filled] = b
				filled++
			}
		}
		for plane := 0; plane < nplanes; plane++ { // 4 planes → 4bpp
			pb := line[plane*bpl:]
			for x := 0; x < w; x++ {
				bit := (pb[x>>3] >> (7 - (x & 7))) & 1
				pix[row*w+x] |= bit << uint(plane)
			}
		}
	}
	return pix, pal, w, h, nil
}
