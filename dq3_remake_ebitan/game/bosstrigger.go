package game

// bossTrigger 描述一個「examine 固定座標 → 開 boss 戰」的觸發點。移植 main.c 的
// `cur_cty==X && b4==Y` 特例分支。
// 資料源:CTY 靜態資料 + IDA/Capstone 對 DQ3.EXE sub2 handler 的交叉驗證。
type bossTrigger struct {
	cty, sec       int
	coords         [][2]int // 觸發格(面向格或腳下命中任一即算;2×2 boss 房多格同事件)
	monsters       []int    // boss 戰鬥鏈,依序全勝才算完成
	doneFlag       int      // 全勝後設的旗標;examine 前先查此旗標,已設 → 不再開戰(對齊原版 examine 已清事件)
	clearStoryFlag int      // 原版 [0x4f70] 勝利後清除的旗標(0=無)
	preRec         int      // 開戰前對白(-1=無)
	postRec        int      // 全勝後對白(-1=無)
	rewardItems    []int    // 全勝後入背包的獎勵道具(nil = 無實體獎勵,只設旗標)
}

// bossTriggers:已確認的座標 boss 觸發點。
var bossTriggers = []bossTrigger{
	{ // 八頭大蛇(CTY19 sec1 (35,12),dlg=45;main.c cur_cty==19 && b4==45)。
		// 勝利給紫寶珠 0x69 + 草薙之劍 0x14(杜勝利攻略:無姬大人=八頭大蛇)。
		cty: 19, sec: 1,
		coords:      [][2]int{{35, 12}},
		monsters:    []int{75},
		doneFlag:    0x44,
		preRec:      -1,
		postRec:     -1,
		rewardItems: []int{0x69, 0x14},
	},
	{ // 巴拉摩斯房：CTY65 sec0 NPC=(8,3),sub2 handler70。
		// IDA sub_1622A → formation ds:4EDF={1,0x27,6,0x79,1}，正常只戰一次；
		// 勝利 CLR 原版 flag0x29，D3TXT06 rec85→戰、rec86→戰後光線。
		cty: 65, sec: 0,
		coords:         [][2]int{{8, 3}},
		monsters:       []int{0x79},
		doneFlag:       0x213,
		clearStoryFlag: 0x29,
		preRec:         85,
		postRec:        86,
		rewardItems:    nil,
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
