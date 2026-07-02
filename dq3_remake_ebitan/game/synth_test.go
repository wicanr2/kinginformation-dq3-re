package game

import "testing"

// 彩虹合成:太陽之石+雲雨之杖 → 消耗太陽之石、雲雨之杖格變彩虹水滴、設 RAINBOW 里程碑。
func TestSynthRainbow(t *testing.T) {
	g := &Game{flags: map[int]bool{}, noticeCode: -1}
	g.inventory = []int{itemSunStone, itemRaincloudRod, 0x41}
	if !g.synthRainbowAtShrine() {
		t.Fatal("有齊材料應合成成功")
	}
	if g.hasItem(itemSunStone) {
		t.Error("太陽之石未被消耗")
	}
	if g.hasItem(itemRaincloudRod) || !g.hasItem(itemRainbowDrop) {
		t.Error("雲雨之杖未變成彩虹水滴")
	}
	if !g.progressDone(msRainbow) {
		t.Error("合成後應設 RAINBOW 里程碑(flag 0x139)")
	}
	// 再合成一次 → 已設旗標 → false(不重複)
	g.inventory = append(g.inventory, itemSunStone, itemRaincloudRod)
	if g.synthRainbowAtShrine() {
		t.Error("已合成過不應重複")
	}
}

// 材料不足不合成、不設旗標。
func TestSynthRainbowNoMaterials(t *testing.T) {
	g := &Game{flags: map[int]bool{}, noticeCode: -1}
	g.inventory = []int{itemSunStone} // 缺雲雨之杖
	if g.synthRainbowAtShrine() {
		t.Error("缺材料不應合成")
	}
	if g.progressDone(msRainbow) {
		t.Error("未合成不應設旗標")
	}
}
