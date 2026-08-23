package dq3data

import "testing"

func TestLoadCharSpriteReordersRawUpAndLeftToLogicalDirections(t *testing.T) {
	raw := make([]byte, blsBody0+CharDirs*blsStride)
	// 每個 raw direction 的第一個 pixel 寫入不同 palette index，mask 預設不影響 Px。
	for dir := 0; dir < CharDirs; dir++ {
		base := blsBody0 + dir*blsStride
		value := dir + 1
		for storedPlane, logicalPlane := range [4]int{3, 2, 1, 0} {
			if value&(1<<logicalPlane) != 0 {
				raw[base+storedPlane*4*CharH] = 0x80
			}
		}
	}
	sprite := LoadCharSprite(raw, 0)
	want := [CharDirs]uint8{1, 3, 2, 4} // logical 下、上、左、右 ← raw 下、左、上、右
	for dir := 0; dir < CharDirs; dir++ {
		if got := sprite.Frames[dir*CharWalk].Px[0][0]; got != want[dir] {
			t.Fatalf("logical direction %d palette=%d, want %d", dir, got, want[dir])
		}
	}
}

func TestLoadCharSprite(t *testing.T) {
	raw := findAsset(t, "DQ3MST.BLS") // 主角
	cs := LoadCharSprite(raw, 0)      // entry_base 0

	// 8 frame(4 方向 × 2 walk);每 frame 像素 0..15;整體至少有不透明像素(不是全空/全透)。
	anyOpaque := false
	for i := 0; i < CharFrames; i++ {
		f := cs.Frames[i]
		for r := 0; r < CharH; r++ {
			for x := 0; x < CharW; x++ {
				if f.Px[r][x] > 15 {
					t.Fatalf("frame %d px[%d][%d]=%d 超 4-bit", i, r, x, f.Px[r][x])
				}
				if f.Opaque[r][x] {
					anyOpaque = true
				}
			}
		}
	}
	if !anyOpaque {
		t.Fatal("主角 8 frame 全透明(解碼疑有誤)")
	}
	// 決定性
	if LoadCharSprite(raw, 0).Frames[0] != cs.Frames[0] {
		t.Fatal("解碼非決定性")
	}
	// 走路動畫:同方向兩 walk frame 應不同(手腳擺動)
	if cs.Frames[0] == cs.Frames[1] {
		t.Log("注意:方向0 的 walk0/walk1 相同(待機不擺動或該 entry 無走路變體)")
	}
	t.Logf("主角:%d frame(4 方向×2 walk),有不透明像素 ✓", CharFrames)
}
