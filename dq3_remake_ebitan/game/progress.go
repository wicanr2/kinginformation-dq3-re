package game

// 主線進度里程碑鏈(移植 dq3_progress.c)。里程碑順序 = 杜勝利攻略進度;
// 完成事件設旗標(g.flags),progressStage 由已設旗標算目前階段。
// 旗標 id 用 0x200+(避開 EXE 既有旗標 < 0x200);RAINBOW/DESCEND 同步既有 EXE 旗標(eff_flag)。

const (
	msStart    = 0x200 // 阿里阿罕開場
	msThiefKey = 0x201 // 盜賊鑰匙(拿吉米之塔)
	msMagicBal = 0x202 // 雷貝魔法球
	msRomaly   = 0x203 // 羅馬利亞皇冠線
	msDhama    = 0x204 // 達瑪轉職開放
	msShip     = 0x205 // 波魯多加胡椒換船
	msRainbow  = 0x206 // 彩虹水滴合成
	msDescend  = 0x207 // 下降アレフガルド
	msZoma     = 0x208 // 大魔王索瑪(終盤)
	msCount    = 9
)

// msSeq:里程碑順序(= 攻略進度)。
var msSeq = [msCount]int{msStart, msThiefKey, msMagicBal, msRomaly, msDhama, msShip, msRainbow, msDescend, msZoma}

// msName:階段名(0..msCount)。
var msName = [msCount + 1]string{
	"未開始", "阿里阿罕開場", "盜賊鑰匙", "魔法球", "羅馬利亞", "達瑪轉職",
	"取船", "彩虹水滴", "下降アレフガルド", "索瑪(終盤)",
}

const (
	flagRainbowSynthed = 0x139 // MS_RAINBOW 對應的既有 EXE 旗標
	flagDescended      = 0x13a // MS_DESCEND 對應的既有 EXE 旗標
)

// effFlag:里程碑 → 實際旗標 id(RAINBOW/DESCEND 同步既有 EXE 旗標,避免兩套狀態)。移植 eff_flag。
func effFlag(ms int) int {
	switch ms {
	case msRainbow:
		return flagRainbowSynthed
	case msDescend:
		return flagDescended
	}
	return ms
}

// progressDone:里程碑是否已完成。移植 dq3_progress_done。
func (g *Game) progressDone(ms int) bool {
	return g.flags != nil && g.flags[effFlag(ms)]
}

// progressSet:設里程碑完成。移植 dq3_progress_set。
func (g *Game) progressSet(ms int) {
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	g.flags[effFlag(ms)] = true
}

// progressStage:目前進度階段(0..msCount)= 連續已完成里程碑數(第一個未完成處停)。移植 dq3_progress_stage。
func (g *Game) progressStage() int {
	for i := 0; i < msCount; i++ {
		if !g.progressDone(msSeq[i]) {
			return i
		}
	}
	return msCount
}

// progressName:階段名(0..msCount)。移植 dq3_progress_name。
func progressName(stage int) string {
	if stage < 0 {
		stage = 0
	}
	if stage > msCount {
		stage = msCount
	}
	return msName[stage]
}
