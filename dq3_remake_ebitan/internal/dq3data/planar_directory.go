package dq3data

import (
	"encoding/binary"
	"fmt"
)

// PlanarSprite 保留原始圖塊的色號；透明與座標由 consumer 決定。
type PlanarSprite struct {
	Width, Height int
	Pixels        []uint8
}

// DecodePlanarDirectory 解碼 u32 offset directory、u16 byte-width/height 與四平面。
// 不解釋 consumer 未讀取的段尾資料，也不推測平面排列。
func DecodePlanarDirectory(raw []byte, layouts []string) ([]PlanarSprite, error) {
	n := len(layouts)
	if n == 0 || n > 256 || len(raw) < n*4 {
		return nil, fmt.Errorf("invalid planar directory")
	}
	result := make([]PlanarSprite, n)
	for i, layout := range layouts {
		start := int(binary.LittleEndian.Uint32(raw[i*4:]))
		end := len(raw)
		if i+1 < n {
			end = int(binary.LittleEndian.Uint32(raw[(i+1)*4:]))
		}
		if start < n*4 || end > len(raw) || start > end-4 {
			return nil, fmt.Errorf("planar entry %d range", i)
		}
		w := int(binary.LittleEndian.Uint16(raw[start:]))
		h := int(binary.LittleEndian.Uint16(raw[start+2:]))
		if w == 0 || w > 1024 || h == 0 || h > 4096 || w*h*4 > end-start-4 {
			return nil, fmt.Errorf("planar entry %d dimensions", i)
		}
		if layout != "byte_interleaved" && layout != "row_planar" {
			return nil, fmt.Errorf("planar entry %d layout", i)
		}
		data := raw[start+4 : end]
		s := PlanarSprite{Width: w * 8, Height: h, Pixels: make([]uint8, w*8*h)}
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				for p := 0; p < 4; p++ {
					o := y*w*4 + x*4 + p
					if layout == "row_planar" {
						o = y*w*4 + p*w + x
					}
					for bit := 0; bit < 8; bit++ {
						s.Pixels[y*s.Width+x*8+bit] |= ((data[o] >> (7 - bit)) & 1) << p
					}
				}
			}
		}
		result[i] = s
	}
	return result, nil
}
