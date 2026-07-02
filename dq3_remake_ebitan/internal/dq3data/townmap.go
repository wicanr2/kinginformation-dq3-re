package dq3data

import "fmt"

// Town 是一張城鎮版面(CTY 檔的一個 section)。移植自 C dq3_town_load 的 layout 解析。
// CTY 檔用 CTYnn.DAT;阿里阿罕(起始城)= CTY00.DAT / section 0 / blk_n=1(DQ31.BLK+BLKBM1.DAT)。
type Town struct {
	W, H           int
	Cells          []byte // 每格 tile 索引(u16 低 byte;高 byte=事件 subid,渲染不用)
	SpawnX, SpawnY int    // 版面 spawn(section header +0x13/+0x14)
}

func townU16(d []byte, o int) uint16 {
	if o < 0 || o+1 >= len(d) {
		return 0
	}
	return uint16(d[o]) | uint16(d[o+1])<<8
}

// OpenTown 解 CTY 檔的某 section → 城鎮版面。移植 dq3_town_load:
// section 偏移表無 count 前綴,word[i]=section i 的 base(0xffff=空);
// layout 指標在 section+0x0e,layout={w:u16, h:u16, tiles…(緊接 +4,stride 2)}。
func OpenTown(cty []byte, section int) (*Town, error) {
	if len(cty) < 2 {
		return nil, fmt.Errorf("CTY too small")
	}
	if section < 0 {
		return nil, fmt.Errorf("section out of range")
	}
	so := int(townU16(cty, 2*section))
	if so == 0xffff || so+0x16 > len(cty) {
		return nil, fmt.Errorf("section empty/oob")
	}
	lay := so + int(townU16(cty, so+0x0e)) // layout_ptr 相對 section
	if lay+4 > len(cty) {
		return nil, fmt.Errorf("layout oob")
	}
	w := int(townU16(cty, lay))
	h := int(townU16(cty, lay+2))
	tbase := lay + 4
	if w <= 0 || h <= 0 || tbase >= len(cty) {
		return nil, fmt.Errorf("tile array oob (w=%d h=%d)", w, h)
	}
	t := &Town{
		W: w, H: h,
		Cells:  make([]byte, w*h),
		SpawnX: int(cty[so+0x13]), // spawn_x(section header)
		SpawnY: int(cty[so+0x14]), // spawn_y
	}
	for i := 0; i < w*h; i++ {
		t.Cells[i] = byte(townU16(cty, tbase+2*i)) // 低 byte = BLK index(容忍末尾略超檔尾 → 0)
	}
	return t, nil
}

// Tile 回 (x,y) 的 tile 索引;越界回 0(對齊 C bounds-safe 讀)。
func (t *Town) Tile(x, y int) int {
	if x < 0 || y < 0 || x >= t.W || y >= t.H {
		return 0
	}
	return int(t.Cells[y*t.W+x])
}
