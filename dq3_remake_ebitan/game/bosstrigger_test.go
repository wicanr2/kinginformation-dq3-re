package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// TestBossTriggerAt:純表查詢(座標/城/section 三者皆須符合;不命中回 nil)。
func TestBossTriggerAt(t *testing.T) {
	if got := bossTriggerAt(19, 1, 35, 12); got != nil {
		t.Fatalf("八頭大蛇是 game-pack 兩階段事件，不得命中舊 bossTrigger：%+v", got)
	}
	for _, c := range [][2]int{{14, 12}, {14, 13}, {15, 12}, {15, 13}} {
		if got := bossTriggerAt(14, 1, c[0], c[1]); got != nil {
			t.Fatalf("CTY14 sec1 %v 是 scripted NPC/scene handler，不得再命中舊座標 boss：%+v", c, got)
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
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	return g
}

// TestTalkStagedOrochiFirstBattle:第一戰由固定 NPC handler45 的正式「對話」入口進入；
// 勝利只從原始怪物掉落取得草薙大劍；
// 紫寶珠與第二戰後 CTY21 寶箱綁定，不可在此提前給予。
func TestTalkStagedOrochiFirstBattle(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(1), 19, true
	n := &npcInst{x: 35, y: 12, b4: 45}

	if !g.talkStagedBoss(n) {
		t.Fatal("handler45 的正式對話入口應觸發第一戰")
	}
	if !g.battle.active || g.battle.monID != 75 {
		t.Fatalf("examine 觸發點應開戰 怪75,得 active=%v monID=%d", g.battle.active, g.battle.monID)
	}
	if g.stagedBossStage != stagedBossFirstBattle || g.pendingTrigger != nil {
		t.Fatalf("應進入 pack 第一戰且不使用舊 pendingTrigger：stage=%d pending=%+v", g.stagedBossStage, g.pendingTrigger)
	}

	g.battle.active = false
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 1, 1, 0x14
	g.onBattleEnd()
	for frames := 0; frames < 200 && g.stagedBossMovingNPC; frames++ {
		g.advanceStagedBossMovement()
	}

	if g.storyFlag(0x44) || !g.storyFlag(0x20) {
		t.Error("NPC slot0 原版移動後應 clear flag0x44 並 set flag0x20")
	}
	if !g.hasItem(0x14) || g.hasItem(0x69) {
		t.Errorf("第一戰只應取得草薙大劍0x14，不得提前取得紫寶珠0x69：%v", g.inventory)
	}
	if g.stagedBossStage != stagedBossFirstMoving {
		t.Fatalf("第一戰後應進入原版兩步移動／轉場階段，stage=%d", g.stagedBossStage)
	}
}

// TestTalkStagedOrochiAbort:戰敗／逃跑不交易旗標，且可由正式對話入口重試。
func TestTalkStagedOrochiAbort(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(1), 19, true
	n := &npcInst{x: 35, y: 12, b4: 45}

	g.talkStagedBoss(n)
	if !g.battle.active {
		t.Fatal("應開戰")
	}
	g.battle.result = 2 // 敗
	g.onBattleEnd()
	if !g.storyFlag(0x44) || g.storyFlag(0x20) {
		t.Error("戰敗不得改動第一階段旗標")
	}
	if g.stagedBossStage != stagedBossIdle || g.stagedBossEventID != "" {
		t.Error("戰敗應清空 staged boss 狀態並允許重試")
	}

	// 重試：再次正式對話應重新開戰（旗標未改）。
	g.battle.active = false
	g.talkStagedBoss(n)
	if !g.battle.active || g.battle.monID != 75 {
		t.Error("戰敗後重新 examine 應可再次觸發同一 boss")
	}
}

// TestStagedOrochiSecondChoiceAndVictory 鎖定原版「Yes 可重談、No 才開第二戰」及
// 第二戰抑制重複掉落、勝利後才解除紫寶珠寶箱 gate。
func TestStagedOrochiSecondChoiceAndVictory(t *testing.T) {
	g := bossTestGame(t)
	g.cur, g.curCty, g.inTown = bossTestScene(0), 21, true
	g.setStoryFlag(0x20, true)
	n := &npcInst{x: 18, y: 7, b4: 36}

	if !g.talkStagedBoss(n) || g.stagedBossStage != stagedBossSecondOffer {
		t.Fatalf("CTY21 無姬交談應顯示第二階段提問，stage=%d", g.stagedBossStage)
	}
	g.advanceStagedBossDialogue()
	if g.stagedBossStage != stagedBossSecondChoice {
		t.Fatalf("rec66 後應進入 Yes/No，stage=%d", g.stagedBossStage)
	}
	g.stagedBossChoiceInput(InputState{Confirm: true})
	if g.stagedBossStage != stagedBossSecondAccept {
		t.Fatalf("Yes 應只顯示 rec67，stage=%d", g.stagedBossStage)
	}
	g.advanceStagedBossDialogue()
	if !g.storyFlag(0x20) || g.battle.active {
		t.Fatal("Yes 不得開戰或消耗第二階段旗標")
	}

	if !g.talkStagedBoss(n) {
		t.Fatal("Yes 後原版允許再次交談")
	}
	g.advanceStagedBossDialogue()
	g.stagedBossCursor = 1
	g.stagedBossChoiceInput(InputState{Confirm: true})
	if g.stagedBossStage != stagedBossSecondChallenge {
		t.Fatalf("No 應顯示 rec68，stage=%d", g.stagedBossStage)
	}
	g.advanceStagedBossDialogue()
	if !g.battle.active || g.stagedBossStage != stagedBossSecondBattle || !g.battle.suppressDrop {
		t.Fatalf("rec68 後應開第二戰且抑制掉落：active=%v stage=%d suppress=%v",
			g.battle.active, g.stagedBossStage, g.battle.suppressDrop)
	}

	g.battle.active = false
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 1, 1, -1
	g.onBattleEnd()
	if g.storyFlag(0x20) || g.storyFlag(0x45) || !g.storyFlag(0x1f) {
		t.Fatal("第二戰勝利旗標交易不符原版")
	}
	if g.hasItem(0x69) {
		t.Fatal("第二戰勝利不得直接給紫寶珠，必須由 CTY21 寶箱取得")
	}
	if g.treasureBlockedByPack(21, 0, 18, 7, 0x69) {
		t.Fatal("第二戰勝利後應解除紫寶珠寶箱 gate")
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

	// 真實載入 CTY19 sec1，使用 handler45 固定 NPC 的正式對話入口。
	blkn := 1
	if 19 < len(mapBlkNum) {
		blkn = mapBlkNum[19]
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 19, blkn, 1, 0, nil)
	if err != nil {
		t.Fatalf("載入 CTY19 sec1: %v", err)
	}
	g.town, g.cur, g.inTown, g.curCty = sc, sc, true, 19
	g.px, g.py = sc.startPos()
	traceWalkTo(t, g, 34, 12)
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
	for _, in := range []InputState{
		{DirHeld: 3, DirEdge: -1},
		{DirHeld: -1, DirEdge: -1, Confirm: true},
		{DirHeld: -1, DirEdge: -1, Confirm: true},
	} {
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	firstPX, firstPY := g.px, g.py
	if !g.battle.active {
		t.Fatalf("handler45 對話應觸發第一戰：monID=%d active=%v", g.battle.monID, g.battle.active)
	}
	dump("jipang_orochi_first_battle")

	g.battle.active = false
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 1, 1, 0x14
	g.onBattleEnd()
	for i := 0; i < 300 && g.stagedBossStage == stagedBossFirstMoving; i++ {
		g.advanceStagedBossMovement()
	}
	if g.stagedBossStage != stagedBossFirstPost || g.curCty != 21 {
		t.Fatalf("第一戰後未抵達宮殿 rec70：talk=(%d,%d) now=(%d,%d) cty=%d stage=%d",
			firstPX, firstPY, g.px, g.py, g.curCty, g.stagedBossStage)
	}
	dump("jipang_orochi_first_post")
	g.dlg.open = false
	g.advanceStagedBossDialogue()

	traceTalkNPC(t, g, 18, 7)
	g.dlg.open = false
	g.advanceStagedBossDialogue()
	g.stagedBossCursor = 1
	g.stagedBossChoiceInput(InputState{Confirm: true})
	g.dlg.open = false
	g.advanceStagedBossDialogue()
	if !g.battle.active || !g.battle.suppressDrop {
		t.Fatal("第二戰未啟動或未抑制掉落")
	}
	dump("jipang_orochi_second_battle")

	g.battle.active = false
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 1, 1, -1
	g.onBattleEnd()
	g.dlg.open = false
	g.advanceStagedBossDialogue()
	traceExaminePackTreasure(t, g, gamepack.QuestTreasureSelector{
		CTYRaw: 21, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x69, PresentFlag: 0x89,
	})
	dump("jipang_purple_orb_obtained")
}
