package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// TestBossTriggerAt:純表查詢(座標/城/section 三者皆須符合;不命中回 nil)。
func TestBossTriggerAt(t *testing.T) {
	if t19 := bossTriggerAt(19, 1, 35, 12); t19 == nil || t19.monsters[0] != 75 || t19.doneFlag != 0x44 {
		t.Fatalf("CTY19 sec1 (35,12) 應命中八頭大蛇(怪75,flag0x44),得 %+v", t19)
	}
	for _, c := range [][2]int{{14, 12}, {14, 13}, {15, 12}, {15, 13}} {
		tt := bossTriggerAt(14, 1, c[0], c[1])
		if tt == nil || len(tt.monsters) != 2 || tt.monsters[0] != 27 || tt.monsters[1] != 26 || tt.doneFlag != 0x211 {
			t.Fatalf("CTY14 sec1 %v 應命中甘達特巢穴鏈(27→26,flag0x211),得 %+v", c, tt)
		}
	}
	if bossTriggerAt(19, 1, 0, 0) != nil {
		t.Error("非觸發格應回 nil")
	}
	if bossTriggerAt(19, 2, 35, 12) != nil {
		t.Error("section 不符應回 nil(19 sec2 非 sec1)")
	}
	if bossTriggerAt(99, 1, 35, 12) != nil {
		t.Error("不存在的 CTY 應回 nil")
	}
}

// bossTestScene:給 examine() 用的最小合法 Scene(不需真實地圖,只要 tileAt/attr 不 panic)。
func bossTestScene(sec int) *Scene {
	return &Scene{w: 60, h: 40, sec: sec, tileAt: func(x, y int) int { return 0 }}
}

// bossTestGame:掛好 mons/shp 供 battle.start 使用的最小 Game(不經 NewGame 完整初始化)。
func bossTestGame(t *testing.T) *Game {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{flags: map[int]bool{}}
	g.battle.mons = mons
	g.battle.shp = asset(t, "DQ3MNS.SHP")
	_, mhp, _, _, _ := g.heroStats()
	g.heroHP, g.heroMP, g.heroInit = mhp, g.heroMaxMP(), true
	return g
}

// TestExamineBossTriggerSingle:CTY19 (35,12) 八頭大蛇——面向格 examine → 開戰(怪75);
// 模擬勝出(對齊 spine_test 手動驅動慣例:result=1 + 真實 onBattleEnd())→ 獎勵道具 0x69+0x14 入背包
// + flag 0x44 設;已設旗標後再 examine 同格 → 不再開戰(對齊原版 examine 已清事件)。
func TestExamineBossTriggerSingle(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(1), 19, true
	g.px, g.py, g.facing = 35, 13, 1 // 站 boss 格正下方,面向上 → frontTile=(35,12)

	g.examine()
	if !g.battle.active || g.battle.monID != 75 {
		t.Fatalf("examine 觸發點應開戰 怪75,得 active=%v monID=%d", g.battle.active, g.battle.monID)
	}
	if g.pendingTrigger == nil || g.pendingTrigger.doneFlag != 0x44 {
		t.Fatalf("應記錄 pendingTrigger(flag0x44),得 %+v", g.pendingTrigger)
	}

	g.battle.result, g.battle.gotExp, g.battle.gotGold = 1, 1, 1 // 判定本場勝出(手動驅動)
	g.onBattleEnd()

	if !g.flags[0x44] {
		t.Error("擊敗八頭大蛇應設 flag 0x44")
	}
	if !g.hasItem(0x69) || !g.hasItem(0x14) {
		t.Errorf("擊敗八頭大蛇應得紫寶珠0x69+草薙之劍0x14,背包=%v", g.inventory)
	}
	if g.pendingTrigger != nil {
		t.Error("鏈全勝後 pendingTrigger 應清 nil")
	}

	// 已擊敗 → 再 examine 同格不應重新開戰
	g.battle.active = false
	g.examine()
	if g.battle.active {
		t.Error("已擊敗的 boss 觸發點不應重複開戰")
	}
}

// TestExamineBossTriggerChain:CTY14 2×2 甘達特巢穴——examine 任一格皆命中同一鏈(27→26)。
// 逐場真實 g.onBattleEnd() 結算(mid-chain 勝出不給獎勵、鏈全勝才給旗標),
// 對齊 spine_test 步驟9 的 Update() 驅動邏輯(won && bossQueue 非空 → advanceBossQueue)。
func TestExamineBossTriggerChain(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(1), 14, true
	g.px, g.py, g.facing = 13, 12, 3 // 面向右 → frontTile=(14,12),鏈四格之一

	g.examine()
	if !g.battle.active || g.battle.monID != 27 {
		t.Fatalf("鏈第一場應為甘達特手下(怪27),得 active=%v monID=%d", g.battle.active, g.battle.monID)
	}

	g.battle.result, g.battle.gotExp, g.battle.gotGold = 1, 1, 1
	g.onBattleEnd()
	if g.pendingTrigger == nil {
		t.Fatal("鏈未完(還剩怪26)→ pendingTrigger 不應清")
	}
	if g.flags[0x211] {
		t.Error("鏈中段(僅勝怪27)不應提早設 flag 0x211")
	}
	if len(g.bossQueue) != 1 {
		t.Fatalf("鏈剩怪26 → bossQueue 應剩1筆,得 %v", g.bossQueue)
	}
	g.battle.active = false
	g.advanceBossQueue() // 對齊 Update() 的 won&&len>0 續打下一場
	if !g.battle.active || g.battle.monID != 26 {
		t.Fatalf("鏈第二場應為甘達特本體(怪26),得 active=%v monID=%d", g.battle.active, g.battle.monID)
	}

	g.battle.result, g.battle.gotExp, g.battle.gotGold = 1, 1, 1
	g.onBattleEnd()
	if !g.flags[0x211] {
		t.Error("鏈全勝(27→26)應設 flag 0x211")
	}
	if g.pendingTrigger != nil {
		t.Error("鏈全勝後 pendingTrigger 應清 nil")
	}
}

// TestExamineBossTriggerAbort:鏈中戰敗/逃跑 → pendingTrigger 清空、旗標不設,
// 允許玩家重新 examine 觸發(不會卡在半殘鏈)。
func TestExamineBossTriggerAbort(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(1), 19, true
	g.px, g.py, g.facing = 35, 13, 1

	g.examine()
	if !g.battle.active {
		t.Fatal("應開戰")
	}
	g.battle.result = 2 // 敗
	g.onBattleEnd()
	if g.flags[0x44] {
		t.Error("戰敗不應設完成旗標")
	}
	if g.pendingTrigger != nil {
		t.Error("戰敗應清空 pendingTrigger,允許重試")
	}

	// 重試:再次 examine → 應重新開戰(旗標未設)
	g.battle.active = false
	g.examine()
	if !g.battle.active || g.battle.monID != 75 {
		t.Error("戰敗後重新 examine 應可再次觸發同一 boss")
	}
}

// TestDumpBossTrigger:視覺核驗——examine 觸發 boss 開戰的畫面(怪物+指令窗)dump 成 PNG。
// 平常 skip;設 DQ3_DUMP_BOSS=1 + BOSS_OUT=<gitignored 夾> 才跑(規則 63/64:靜態測試之外
// 留一份可肉眼核對的畫面,證明機制在真實素材/真實 renderFrame 下確實 live)。
func TestDumpBossTrigger(t *testing.T) {
	if os.Getenv("DQ3_DUMP_BOSS") == "" {
		t.Skip("設 DQ3_DUMP_BOSS=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("BOSS_OUT")
	if out == "" {
		t.Fatal("需設 BOSS_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false

	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(out + "/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped", name)
	}

	// 真實載入 CTY19 sec1(八頭大蛇巢穴),站在 boss 格前 examine 觸發開戰。
	blkn := 1
	if 19 < len(mapBlkNum) {
		blkn = mapBlkNum[19]
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 19, blkn, 1, 0)
	if err != nil {
		t.Fatalf("載入 CTY19 sec1: %v", err)
	}
	g.town, g.cur, g.inTown, g.curCty = sc, sc, true, 19
	g.px, g.py, g.facing = 35, 13, 1
	dump("boss_hydra_before") // 觸發前:站在巢穴內,面向 boss 格

	g.examine()
	if !g.battle.active {
		t.Fatalf("examine 應觸發八頭大蛇開戰(monID=%d active=%v)", g.battle.monID, g.battle.active)
	}
	dump("boss_hydra_battle") // 觸發後:戰鬥畫面(怪物 sprite + 指令窗)
	t.Logf("boss trigger dump ✓:monID=%d flag_pending=%v", g.battle.monID, g.pendingTrigger != nil)
}
