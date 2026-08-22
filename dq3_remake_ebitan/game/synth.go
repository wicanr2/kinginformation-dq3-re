package game

// 合成祠堂事件(移植 dq3_synth_rainbow_drop / synth_shrine_examine,file 0x77ce)。
// 「雨和太陽合而為一出現彩虹」:太陽之石 + 雲雨之杖 → 彩虹水滴。祠堂 = CTY93。

const (
	itemSunStone     = 0x72 // 太陽之石
	itemRaincloudRod = 0x73 // 雲雨之杖
	itemRainbowDrop  = 0x75 // 彩虹水滴(正確合成品;#2 修正,非銀寶珠 0x6b)
	shrineCty        = 93   // 合成祠堂(下層,利姆達爾南方)
)

// synthRainbowAtShrine:在合成祠堂嘗試「太陽之石 + 雲雨之杖 → 彩虹水滴」。
// 已合成(flagRainbowSynthed)→ 不重複;材料不足 → false。成功消耗太陽之石、把一支雲雨之杖換成彩虹水滴、
// 設 RAINBOW 里程碑。移植 synth_shrine_examine + synth_rainbow_drop。
func (g *Game) synthRainbowAtShrine() bool {
	if g.progressDone(msRainbow) { // 旗標 0x139 已設 → 不再合成(避免重複)
		return false
	}
	if !g.hasPartyItem(itemSunStone) || !g.hasPartyItem(itemRaincloudRod) { // 材料不足
		return false
	}
	g.removePartyItems(itemSunStone, 1) // 消耗太陽之石
	if !g.replacePartyItem(itemRaincloudRod, itemRainbowDrop) {
		return false
	}
	g.progressSet(msRainbow) // set flag 0x139 → 推進 RAINBOW 里程碑
	g.noticeCode, g.noticeTimer = itemRainbowDrop, 120
	return true
}
