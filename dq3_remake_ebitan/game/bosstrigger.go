package game

// bossTrigger 描述一個「examine 固定座標 → 開 boss 戰」的觸發點。移植 main.c 的
// `cur_cty==X && b4==Y` 特例分支(八頭大蛇/甘達特巢穴,main.c ~L1644/L1663)。
// 資料源:docs/data/boss-trigger-points.md「已確認的 boss 觸發點」表(npc_dialogue.json
// kind=special 靜態 RE 結果),不重新反組譯。巴拉摩斯/怪力魔等座標未在該文件定案
// (見文件「巴拉摩斯城 CTY 定位(未決)」段),故不收錄——待人工補 RE 後再擴表。
type bossTrigger struct {
	cty, sec    int
	coords      [][2]int // 觸發格(面向格或腳下命中任一即算;2×2 boss 房多格同事件)
	monsters    []int    // boss 戰鬥鏈,依序全勝才算完成(單一 boss 僅 1 筆;甘達特巢穴兩段連戰)
	doneFlag    int      // 全勝後設的旗標;examine 前先查此旗標,已設 → 不再開戰(對齊原版 examine 已清事件)
	rewardItems []int    // 全勝後入背包的獎勵道具(nil = 無實體獎勵,只設旗標)
}

// bossTriggers:已確認的座標 boss 觸發點。
var bossTriggers = []bossTrigger{
	{ // 八頭大蛇(CTY19 sec1 (35,12),dlg=45;main.c cur_cty==19 && b4==45)。
		// 勝利給紫寶珠 0x69 + 草薙之劍 0x14(杜勝利攻略:無姬大人=八頭大蛇)。
		cty: 19, sec: 1,
		coords:      [][2]int{{35, 12}},
		monsters:    []int{75},
		doneFlag:    0x44,
		rewardItems: []int{0x69, 0x14},
	},
	{ // 甘達特盜賊巢穴(CTY14 sec1 2×2 (14,12)(14,13)(15,12)(15,13),dlg=58;
		// main.c cur_cty==14 && b4==58)。守衛甘達特手下(怪27)→ 返回的甘達特(怪26)。
		// 救人劇情:無道具獎勵,設 flag 0x211(與 CTY15 古布達黑胡椒鏈 scripted byte4=25 的
		// gating 閉環——該 gating 尚未接進 scriptedTalk,屬另一批次 C-4b,此處只負責設旗標)。
		cty: 14, sec: 1,
		coords:      [][2]int{{14, 12}, {14, 13}, {15, 12}, {15, 13}},
		monsters:    []int{27, 26},
		doneFlag:    0x211,
		rewardItems: nil,
	},
}

// bossTriggerAt:查 (cty,sec,x,y) 是否命中已確認的 boss 觸發點;命中回該筆指標,否則 nil。
func bossTriggerAt(cty, sec, x, y int) *bossTrigger {
	for i := range bossTriggers {
		t := &bossTriggers[i]
		if t.cty != cty || t.sec != sec {
			continue
		}
		for _, c := range t.coords {
			if c[0] == x && c[1] == y {
				return t
			}
		}
	}
	return nil
}
