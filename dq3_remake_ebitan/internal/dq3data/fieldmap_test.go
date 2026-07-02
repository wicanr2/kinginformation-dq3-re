package dq3data

import "testing"

func TestOpenFieldMap(t *testing.T) {
	raw := findAsset(t, "DQ3CON.MAP")
	m, err := OpenFieldMap(raw)
	if err != nil {
		t.Fatalf("OpenFieldMap: %v", err)
	}
	if m.W <= 0 || m.H <= 0 || len(m.Cells) != m.W*m.H {
		t.Fatalf("尺寸不符:w=%d h=%d cells=%d", m.W, m.H, len(m.Cells))
	}
	t.Logf("DQ3CON.MAP: %d×%d = %d 格", m.W, m.H, len(m.Cells))

	// 交叉驗證:所有 tile 索引都在 DQ3.BLK 的 tile 數內(地表用 DQ3.BLK)。
	blk, err := OpenBLK(findAsset(t, "DQ3.BLK"))
	if err != nil {
		t.Fatalf("OpenBLK: %v", err)
	}
	over := 0
	for _, c := range m.Cells {
		if int(c) >= blk.Count {
			over++
		}
	}
	if over > 0 {
		t.Fatalf("%d 格的 tile 索引超出 DQ3.BLK 的 %d tiles", over, blk.Count)
	}
	// 角落 tile 可取
	_ = m.Tile(0, 0)
	_ = m.Tile(m.W-1, m.H-1)
}
