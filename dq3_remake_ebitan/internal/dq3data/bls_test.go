package dq3data

import "testing"

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
