package game

const (
	ctyRomaly              = 2
	romalyKingSection      = 1
	romalyKingHandler      = 9
	romalyCrownPendingFlag = 0x2c
	romalyKingGreetingRec  = 45
	romalyCrownQuestRec    = 15
	romalyGoldenCrownItem  = 0x33
)

// talkRomalyKing 還原 DQ3.EXE handler9 首次、尚未取得金皇冠的主線分支
// (file 0x65ee..0x6646)：flag0x2c 為 1 且背包找不到 item0x33 時，
// 依序播放 D3TXT02 rec45 與 rec15。此分支不修改 flag 或道具，重談會重播任務。
//
// 持皇冠後的讓位 Yes/No 分支(file 0x6649 起)另需原版選擇 modal；未完成前不可自動
// 接受王位或擅自消耗皇冠。
func (g *Game) talkRomalyKing(n *npcInst) bool {
	if g.curCty != ctyRomaly || g.cur == nil || g.cur.sec != romalyKingSection ||
		n == nil || n.b4 != romalyKingHandler {
		return false
	}
	if g.storyFlag(romalyCrownPendingFlag) && !g.hasItem(romalyGoldenCrownItem) {
		if g.dlg.Open(romalyKingGreetingRec) {
			g.romalyKingStage = 1
		}
		return true
	}
	return true // handler9 已辨識；未完成的皇冠選擇分支 fail-closed，不落入錯誤泛用表。
}

func (g *Game) advanceRomalyKing() {
	if g.romalyKingStage != 1 {
		return
	}
	g.romalyKingStage = 0
	g.dlg.Open(romalyCrownQuestRec)
}
