package dq3data

import "errors"

// FieldMap 是地表 overworld tilemap(DQ3CON.MAP / 下層 DQ3CO2.MAP)。移植自 C dq3_field.c。
// 格式:header 4B(w:u16, h:u16),之後 w*h bytes 各 = 該格的 BLK tile 索引。
type FieldMap struct {
	W, H  int
	Cells []byte // 長度 W*H,每 byte = tile 索引
}

// OpenFieldMap 解 overworld map。移植 dq3_field_load_map 的 header 解析。
func OpenFieldMap(d []byte) (*FieldMap, error) {
	if len(d) < 4 {
		return nil, errors.New("map too small")
	}
	w := int(d[0]) | int(d[1])<<8
	h := int(d[2]) | int(d[3])<<8
	if w <= 0 || h <= 0 || w*h+4 > len(d) {
		return nil, errors.New("map size mismatch")
	}
	return &FieldMap{W: w, H: h, Cells: d[4 : 4+w*h]}, nil
}

// Tile 回 (x,y) 格的 tile 索引;越界回 0。
func (m *FieldMap) Tile(x, y int) int {
	if x < 0 || y < 0 || x >= m.W || y >= m.H {
		return 0
	}
	return int(m.Cells[y*m.W+x])
}
