package game

import (
	"os"
	"reflect"
	"testing"
)

func TestRomalyKingFirstQuestRecords(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(dir + "/CTY02.DAT"); err != nil {
		t.Skipf("無素材:%v", err)
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyRomaly, mapBlkNum[ctyRomaly], romalyKingSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyRomaly
	g.dlg.tx = sc.dlgText
	g.inventory = []int{0x55}
	g.setStoryFlag(romalyCrownPendingFlag, true) // isolated fixture：production 由新遊戲原始 flag image 建立
	var king *npcInst
	for i := range sc.npcs {
		if sc.npcs[i].b4 == romalyKingHandler && (sc.npcs[i].ctrl>>3)&7 == 2 {
			king = &sc.npcs[i]
			break
		}
	}
	if king == nil {
		t.Fatal("CTY02 sec1 找不到 handler9 國王")
	}
	if !g.talkRomalyKing(king) || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, sc.dlgText.Record(romalyKingGreetingRec)) {
		t.Fatal("首次晉見未開 D3TXT02 rec45")
	}
	g.dlg.open = false
	g.advanceRomalyKing()
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, sc.dlgText.Record(romalyCrownQuestRec)) {
		t.Fatal("rec45 結束後未接 D3TXT02 rec15 金皇冠任務")
	}
	if !g.storyFlag(romalyCrownPendingFlag) || !g.hasItem(0x55) ||
		g.hasItem(romalyGoldenCrownItem) {
		t.Fatalf("任務對白不應修改交易狀態：flag2c=%v key=%v crown=%v",
			g.storyFlag(romalyCrownPendingFlag), g.hasItem(0x55), g.hasItem(romalyGoldenCrownItem))
	}

	// 尚未還原的王位 Yes/No modal 必須 fail-closed：不可因談話就自動收走皇冠。
	g.dlg.open = false
	g.inventory = append(g.inventory, romalyGoldenCrownItem)
	if !g.talkRomalyKing(king) || g.dlg.open || !g.hasItem(romalyGoldenCrownItem) ||
		!g.storyFlag(romalyCrownPendingFlag) {
		t.Fatalf("皇冠分支應 fail-closed：dlg=%v crown=%v flag2c=%v",
			g.dlg.open, g.hasItem(romalyGoldenCrownItem), g.storyFlag(romalyCrownPendingFlag))
	}
}
