package dq3data

import "testing"

func TestBlockAttr(t *testing.T) {
	raw := findAsset(t, "BLKBM.DAT")
	attr := OpenBlockAttr(raw)
	if len(attr.A) == 0 {
		t.Fatal("BLKBM.DAT 解出 0 項")
	}
	t.Logf("BLKBM.DAT %d bytes → %d tile 屬性", len(raw), len(attr.A))

	// 地表應同時有「可走(草地)」與「阻擋(山/海)」的 tile —— 屬性資料才有意義。
	blocked, walkable := 0, 0
	for i := range attr.A {
		if attr.Blocked(i) {
			blocked++
		} else {
			walkable++
		}
	}
	t.Logf("阻擋 %d / 可走 %d", blocked, walkable)
	if blocked == 0 || walkable == 0 {
		t.Fatal("屬性全同(疑解碼有誤 —— 地表應有山/海阻擋 + 草地可走)")
	}
	// 越界 → 可走(對齊 C dq3_scene_walkable)
	if attr.Blocked(len(attr.A)) || attr.Blocked(-1) {
		t.Fatal("越界應視為可走")
	}
}
