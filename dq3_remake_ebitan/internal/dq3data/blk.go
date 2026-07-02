package dq3data

import "errors"

// BLK 是 DQ3 tile 圖庫(地表 DQ3.BLK / 城鎮 DQ3N.BLK)。移植自 C dq3_blk.c(docs/04)。
// header 6B(row_bytes, height, count);每 tile = row_bytes*height*4 bytes,4-bit planar
// (4 plane 各 row_bytes*height,row×row_bytes,MSB-first);解出 32×24 indexed 像素(0..15)。
type BLK struct {
	RowBytes int
	Height   int
	Count    int
	TileSize int
	body     []byte
}

// OpenBLK 解 BLK header。移植 dq3_blk_open。
func OpenBLK(d []byte) (*BLK, error) {
	if len(d) < 6 {
		return nil, errors.New("BLK too small")
	}
	b := &BLK{
		RowBytes: int(d[0]) | int(d[1])<<8,
		Height:   int(d[2]) | int(d[3])<<8,
		Count:    int(d[4]) | int(d[5])<<8,
	}
	b.TileSize = b.RowBytes * b.Height * 4
	b.body = d[6:]
	if b.Count <= 0 || b.TileSize <= 0 {
		return nil, errors.New("bad BLK header")
	}
	if b.Count*b.TileSize > len(b.body) {
		return nil, errors.New("BLK body too short")
	}
	return b, nil
}

// Tile 解第 idx 個 tile → 24×32 indexed 像素(值 0..15,對 palette)。移植 dq3_blk_tile。
func (b *BLK) Tile(idx int) [24][32]uint8 {
	var px [24][32]uint8
	if idx < 0 || idx >= b.Count {
		return px
	}
	tile := b.body[idx*b.TileSize:]
	planesz := b.RowBytes * b.Height
	for plane := 0; plane < 4; plane++ {
		base := plane * planesz
		for row := 0; row < b.Height && row < 24; row++ {
			ro := base + row*b.RowBytes
			for bx := 0; bx < b.RowBytes && bx < 4; bx++ {
				by := tile[ro+bx]
				if by == 0 {
					continue
				}
				for bit := 0; bit < 8; bit++ {
					if by&(0x80>>bit) != 0 {
						px[row][bx*8+bit] |= 1 << plane
					}
				}
			}
		}
	}
	return px
}
