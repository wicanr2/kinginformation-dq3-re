package dq3data

// BlockAttr 是 tile 屬性表(BLKBM.DAT / BLKBMn.DAT):每 tile 一個 u16。移植自 C dq3_scene.c 用法。
// bit0(0x0001)=阻擋不可走;bit3(0x0008)=事件格;高 3 bit(0xe000)=轉場 tile。
// 地表/城鎮碰撞只需 bit0(dq3_scene_walkable:`a & 0x0001 ? 不可走 : 可走`)。
type BlockAttr struct{ A []uint16 }

// OpenBlockAttr 解 BLKBM.DAT → 每 tile 一個 u16(little-endian)。
func OpenBlockAttr(d []byte) *BlockAttr {
	n := len(d) / 2
	a := make([]uint16, n)
	for i := 0; i < n; i++ {
		a[i] = uint16(d[i*2]) | uint16(d[i*2+1])<<8
	}
	return &BlockAttr{A: a}
}

// Raw 回某 tile 索引的原始屬性 u16(bit3=0x8 事件格、高3bit=0xe000 轉場)。越界回 0。
func (b *BlockAttr) Raw(tileIdx int) uint16 {
	if b == nil || tileIdx < 0 || tileIdx >= len(b.A) {
		return 0
	}
	return b.A[tileIdx]
}

// Blocked 回某 tile 索引是否阻擋(不可走)。移植 dq3_scene_walkable 的 `attr[idx] & 1`。
// tile 索引越界視為可走(對齊 C:無屬性 → 可走)。
func (b *BlockAttr) Blocked(tileIdx int) bool {
	if b == nil || tileIdx < 0 || tileIdx >= len(b.A) {
		return false
	}
	return b.A[tileIdx]&0x0001 != 0
}
