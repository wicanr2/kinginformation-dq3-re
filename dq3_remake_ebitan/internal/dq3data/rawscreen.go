package dq3data

import "fmt"

// DecodeRowInterleaved4BPP 解 legacy raw screen：每一掃描列依序放四個
// 1-bit plane（plane 0..3），不含 header 或 palette。寬度必須是 8 的倍數，
// caller 另依 game-pack contract 提供版本專屬 palette。
func DecodeRowInterleaved4BPP(d []byte, width, height int) ([]uint8, error) {
	if width <= 0 || height <= 0 || width%8 != 0 {
		return nil, fmt.Errorf("raw screen dimensions must be positive and width multiple of 8: %dx%d", width, height)
	}
	bytesPerPlane := width / 8
	rowBytes := bytesPerPlane * 4
	want := rowBytes * height
	if len(d) != want {
		return nil, fmt.Errorf("raw screen size=%d, want %d for %dx%d", len(d), want, width, height)
	}
	pix := make([]uint8, width*height)
	for y := 0; y < height; y++ {
		row := d[y*rowBytes : (y+1)*rowBytes]
		for plane := 0; plane < 4; plane++ {
			pb := row[plane*bytesPerPlane : (plane+1)*bytesPerPlane]
			for xb, b := range pb {
				for bit := 0; bit < 8; bit++ {
					if b&(1<<uint(7-bit)) != 0 {
						pix[y*width+xb*8+bit] |= 1 << uint(plane)
					}
				}
			}
		}
	}
	return pix, nil
}
