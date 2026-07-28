package game

import (
	"os"
	"path/filepath"
	"testing"
)

// spineAssetsDir:主線串接測試用的真實素材目錄(同 asset() helper 的路徑判斷,
// 但這裡要目錄本身餵給 os.DirFS,不是單一檔內容)。無素材 → skip(非 CI 缺件不算失敗)。
func spineAssetsDir(t *testing.T) string {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(filepath.Join(dir, "DQ3.PAL")); err != nil {
		t.Skipf("無素材目錄 %s: %v", dir, err)
	}
	return dir
}

// TestComponentSpine:主線元件串接測試。它用 NewGame() 走 production 初始化，
// (真實素材,非手搭最小 Game),再依攻略順序用真實 game API 依序驅動九大里程碑,
// 每步驗前一系統的輸出真的餵進下一系統(道具→事件旗標→progressStage;戰鬥結算→
// 經驗/金幣→Game 狀態;boss 連戰→onBattleEnd→runFinale→結局捲動)。
// 注意：本測試直接呼叫事件函式並人工判定終盤戰勝，不是玩家輸入的完整通關證據。
//
// 手動驅動(對齊既有 TestZomaSeq 慣例)僅用在:(1) 終盤連戰 —— 用真實成長表等級撞終盤 boss
// 會需要不確定回合數且有陣亡風險,不影響「系統銜接」本身,故直接判定每場勝出,但透過
// 真實 g.onBattleEnd() 寫回(而非手動設 cleared/lotoBlessed),讓 runFinale 由 onBattleEnd
// 內建的 monID==0x7c 判斷觸發,比 TestZomaSeq 更貼近正式 Update() 路徑。
// (2) 取船里程碑(msShip)——見步驟 6 註解,production code 目前唯一設點在 DQ3_SHIP debug env,
// 沒有真實店家/事件觸發路徑,此為已知缺口非本測試要繞過的 bug。
func TestComponentSpine(t *testing.T) {
	dir := spineAssetsDir(t)
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "spine-save.json")) // 隔離存檔,不寫進 repo/CWD

	g, err := NewGame(os.DirFS(dir), nil) // music=nil → 靜音降級,不需音軌檔
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	g.showTitle = false // 略過標題,後續直接呼叫方法驅動(同 debug 捷徑的作法)

	// ---- 1. 開場:出生時尚未見國王，不能預先完成 msStart。----
	if s := g.progressStage(); s != 0 {
		t.Fatalf("NewGame 後尚未見國王，進度應為 0,得 %d", s)
	}
	// 後續只驗元件鏈，因此在這個邊界人工補上尚未移植的國王/酒場事件結果。
	// 完整通關驗收不得引用這個 synthetic transition。
	g.progressSet(msStart)

	// ---- 2. 戰鬥 × 經濟銜接:真實遭遇 → execTurn 逐回合 → g.onBattleEnd() 寫回。
	// 驗 Battle.gotExp/gotGold 真的流回 Game.heroExp/heroGold(playthrough_test 只測 Battle 本身,
	// 沒測這條 Game 銜接線)。用 exp4364(=lv10,對齊 TestPlaythroughProgression)確保打得贏起始弱怪池。
	g.heroExp = 4364
	beforeExp, beforeGold := g.heroExp, g.heroGold
	// 決定性:直接對史萊姆(id5,不逃)開戰,取代 startEncounter 的隨機挑怪
	// (隨機池含會逃的怪,高等級主角下觸發敵逃 result=3 → 測試 flaky)。
	// 仍走真實 battle.start → execTurn → onBattleEnd 的 Game 銜接線。
	lv, mhp, atk, def, agi := g.heroStats()
	g.heroHP, g.heroMP, g.heroInit = mhp, g.heroMaxMP(), true
	hp := heroParams{level: lv, curHP: g.heroHP, maxHP: mhp, atk: atk, def: def, agi: agi,
		herbs: g.countItem(herbCode), mp: g.heroMP, maxMP: g.heroMaxMP(), spells: g.heroSpells()}
	if !g.battle.start(5, 1, hp, g.buildCompanionActors()) {
		t.Fatal("對史萊姆開戰失敗(sprite 缺失?)")
	}
	turns := 0
	for g.battle.result == 0 && turns < 50 {
		g.battle.cursor = bcWar
		g.battle.execTurn()
		turns++
	}
	if g.battle.result != 1 {
		t.Fatalf("lv10 勇者+隊伍對起始弱怪應勝,得 result=%d(%d 回合後敵HP=%d 我方HP=%d)",
			g.battle.result, turns, g.battle.enemies[0].hp, g.battle.heroHP)
	}
	gotExp, gotGold := g.battle.gotExp, g.battle.gotGold
	g.onBattleEnd()
	if g.heroExp != beforeExp+uint32(gotExp) {
		t.Errorf("onBattleEnd 應把戰鬥 EXP 寫回 heroExp:want %d, got %d", beforeExp+uint32(gotExp), g.heroExp)
	}
	if g.heroGold != beforeGold+gotGold {
		t.Errorf("onBattleEnd 應把戰鬥金幣寫回 heroGold:want %d, got %d", beforeGold+gotGold, g.heroGold)
	}
	if g.heroHP != g.battle.heroHP {
		t.Errorf("onBattleEnd 應同步 heroHP:g.heroHP=%d battle.heroHP=%d", g.heroHP, g.battle.heroHP)
	}

	// ---- 3. 劇情道具鏈:盜賊鑰匙(CTY8)→ 雷貝魔法球(CTY1,吃盜賊鑰匙當 prereq,不消耗)。
	// scriptedTalk 只讀 g.curCty/g.inventory/g.flags,不依賴場景已載入 —— 用真實 dispatch 函式驅動,
	// 不手動 progressSet(對齊 scriptedTable 資料驅動設計)。
	g.curCty = 8
	g.scriptedTalk(12) // scriptedTable row0:give item85 + milestone 0x201(msThiefKey)
	if !g.hasItem(85) {
		t.Fatal("盜賊鑰匙事件應給 item85")
	}
	if !g.progressDone(msThiefKey) {
		t.Fatal("盜賊鑰匙事件應設 msThiefKey")
	}
	if s := g.progressStage(); s != 2 {
		t.Fatalf("盜賊鑰匙後進度應到 2,得 %d", s)
	}

	g.curCty = 1
	g.scriptedTalk(7) // scriptedTable row1:prereq item85(已持有,不消耗)→ give item88 + milestone 0x202(msMagicBal)
	if !g.hasItem(88) {
		t.Fatal("雷貝魔法球事件應給 item88")
	}
	if !g.hasItem(85) {
		t.Error("盜賊鑰匙 consume=0,魔法球事件後不應被消耗")
	}
	if !g.progressDone(msMagicBal) {
		t.Fatal("雷貝魔法球事件應設 msMagicBal")
	}
	if s := g.progressStage(); s != 3 {
		t.Fatalf("魔法球後進度應到 3,得 %d", s)
	}

	// ---- 4. 羅馬利亞/達瑪:真實 enterTownCty(真的讀 CTY02.DAT/CTY49.DAT 建場景),
	// 里程碑判定內嵌在 enterTownCty(main.c cur_cty 判定移植),不是獨立函式。
	g.enterTownCty(2)
	if g.curCty != 2 {
		t.Fatalf("進羅馬利亞(CTY2)應成功換場景,curCty=%d(CTY02.DAT 載入失敗?)", g.curCty)
	}
	if !g.progressDone(msRomaly) || g.progressStage() != 4 {
		t.Fatalf("進羅馬利亞應設 msRomaly,階段=%d", g.progressStage())
	}

	g.enterTownCty(49)
	if g.curCty != 49 {
		t.Fatalf("進達瑪神殿(CTY49)應成功換場景,curCty=%d(CTY49.DAT 載入失敗?)", g.curCty)
	}
	if !g.progressDone(msDhama) || g.progressStage() != 5 {
		t.Fatalf("進達瑪神殿應設 msDhama,階段=%d", g.progressStage())
	}

	// ---- 5. 取船(msShip):目前 production code 唯一設定點是 applyDebugEnv 的 DQ3_SHIP 分支
	// (game.go:1319-1323,main.c「波魯多加胡椒換船」事件未移植),沒有可呼叫的真實商店/劇情函式可串。
	// 這是本測試發現的缺口(見任務回報),非測試繞過——如實記錄後改用 progressSet 對齊里程碑鏈往下走。
	g.progressSet(msShip)
	if s := g.progressStage(); s != 6 {
		t.Fatalf("取船後進度應到 6,得 %d", s)
	}

	// ---- 6. 彩虹水滴合成(合成祠堂 CTY93):走 g.examine() 真實 dispatch(對齊 cmdExamine 入口),
	// 不直接呼叫 synthRainbowAtShrine——多驗一層「調查指令 → 祠堂分支」的真實路由。
	g.inventory = append(g.inventory, itemSunStone, itemRaincloudRod)
	g.inTown, g.curCty = true, shrineCty
	g.examine()
	if !g.hasItem(itemRainbowDrop) {
		t.Fatal("合成祠堂調查應產出彩虹水滴")
	}
	if g.hasItem(itemSunStone) {
		t.Error("合成應消耗太陽之石")
	}
	if !g.progressDone(msRainbow) || g.progressStage() != 7 {
		t.Fatalf("合成彩虹水滴應設 msRainbow,階段=%d", g.progressStage())
	}

	// ---- 7. 下降阿雷夫加德底層 helper:descend()(真的讀 DQ3UND.MAP)。
	// 正式 CTY72→CTY77→0xfe production trace 另由 baramos_test 驗證；此處只保留 loader 單元覆蓋。
	g.descend()
	if g.layer != 1 {
		t.Fatalf("descend 應切到下層地表(layer=1),得 %d(DQ3UND.MAP 載入失敗?)", g.layer)
	}
	if !g.progressDone(msDescend) || g.progressStage() != 8 {
		t.Fatalf("下降後進度應到 8,得 %d", g.progressStage())
	}

	// ---- 8. 彩虹水滴原版座標閘：(127,117) 使用，(126,117) 改成 tile0x53。
	// 驗「劇情道具」端到端:合成(步驟6)產出的道具,真的能在正確位置被 useSelectedItem 消費。
	idx := -1
	for i, c := range g.inventory {
		if c == itemRainbowDrop {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("背包應仍持有步驟6合成的彩虹水滴")
	}
	g.panelCursor = idx
	g.px, g.py = rainbowUseX, rainbowUseY
	g.useSelectedItem()
	if g.worldState&worldStateRainbowBridge == 0 ||
		g.cur.tileIdx(rainbowBridgeX, rainbowBridgeY) != rainbowBridgeTile {
		t.Error("正確座標使用彩虹水滴應架起原版 tile0x53")
	}
	if g.hasItem(itemRainbowDrop) {
		t.Error("彩虹水滴架橋後應消耗")
	}

	// ---- 9. 索瑪神殿終盤連戰:真實 startZomaSeq()/advanceBossQueue() 驅動佇列,
	// 每場「判定勝出」後呼叫真實 g.onBattleEnd()(而非手動設 cleared)——runFinale 由
	// onBattleEnd 內建的 monID==0x7c 分支觸發,驗證的是正式 Update() 會走的同一條線。
	g.startZomaSeq()
	if !g.battle.active {
		t.Fatal("startZomaSeq 應開第一場 boss(守衛 106 sprite/數值缺失?)")
	}
	seen := []int{g.battle.monID}
	for i := 0; i < 20 && (g.battle.active || g.zomaIntro); i++ {
		if g.zomaIntro {
			g.dlg.open = false
			g.advanceZomaIntro()
			if g.battle.active {
				seen = append(seen, g.battle.monID)
			}
			continue
		}
		g.battle.result, g.battle.gotExp, g.battle.gotGold = 1, 1, 1 // 判定本場勝出(手動驅動,見函式註解)
		g.onBattleEnd()                                              // 真實寫回:exp/gold/HP +(monID==0x7c → runFinale)
		g.battle.active = false
		if len(g.bossQueue) > 0 {
			g.advanceBossQueue()
			if g.battle.active {
				seen = append(seen, g.battle.monID)
			}
		}
	}
	if seen[len(seen)-1] != 0x7c {
		t.Fatalf("連戰最後一場應為索瑪 0x7c,實際序列=%v", seen)
	}
	if !g.cleared || g.lotoBlessed {
		t.Error("破索瑪應 cleared，但國王冊封前不可先有 lotoBlessed")
	}
	if !g.progressDone(msZoma) {
		t.Error("破索瑪應設 msZoma 里程碑")
	}
	if s := g.progressStage(); s != msCount {
		t.Fatalf("破索瑪後進度應 9/9,得 %d(序列=%v)", s, seen)
	}
	if g.inTown || g.layer != 1 || g.endSeq >= 0 || g.dlg.open {
		t.Fatalf("索瑪後應回可操作的下層地表、不可立刻 ending：town=%v layer=%d end=%d open=%v",
			g.inTown, g.layer, g.endSeq, g.dlg.open)
	}

	// ---- 10. 回拉達多姆見國王：handler74 rec48 關閉後才授予洛特封號並開始 ending。
	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyRadatomeCastle,
		mapBlkNum[ctyRadatomeCastle], radatomeThroneSec, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = throne, throne, true, ctyRadatomeCastle
	g.dlg.tx = throne.dlgText
	var king *npcInst
	for i := range throne.npcs {
		if throne.npcs[i].b4 == radatomeKingHandle {
			king = &throne.npcs[i]
			break
		}
	}
	if king == nil {
		t.Fatal("CTY80 sec1 找不到 handler74 國王")
	}
	g.px, g.py, g.facing = king.x, king.y+1, 1
	g.selectCommand(cmdTalk)
	for g.endingKingStage == 1 {
		g.dlg.open = false
		g.advanceEndingKing()
	}
	if !g.lotoBlessed || !g.flags[0x217] || g.endSeq != 0 || !g.dlg.open || g.dlg.tx != g.endText {
		t.Fatalf("國王冊封後才應開始 ending：loto=%v flag=%v end=%d open=%v",
			g.lotoBlessed, g.flags[0x217], g.endSeq, g.dlg.open)
	}

	// ---- 11. 結局捲動:逐段推進(同 TestEndingScroll)到 endSeq 回 -1(播完)。
	if g.endText == nil || g.endText.NRecords == 0 {
		t.Fatal("應已載入 ENDTXT 結局文本")
	}
	steps := 0
	for g.endSeq >= 0 && steps < g.endText.NRecords+2 {
		g.dlg.open = false // 該段翻完關閉(同 Update 的手動輸入語意)
		g.endSeq++
		if g.endSeq < g.endText.NRecords {
			g.dlg.Open(g.endSeq)
		} else {
			g.endSeq = -1
		}
		steps++
	}
	if g.endSeq != -1 {
		t.Errorf("結局應播完(endSeq=-1),實 %d", g.endSeq)
	}
	if steps != g.endText.NRecords {
		t.Errorf("推進步數 %d != ENDTXT 段數 %d", steps, g.endText.NRecords)
	}

	// ---- 最終斷言:全主線 9/9 銜接完成 + 破關。 ----
	if s := g.progressStage(); s != msCount {
		t.Errorf("最終進度應 9/9,得 %d", s)
	}
	if !g.cleared {
		t.Error("最終應 cleared==true")
	}
	t.Logf("主線串接 ✓:9/9 里程碑、%d 回合擊倒起始遭遇(EXP%d G%d)、連戰序列=%v、結局 %d 段播完",
		turns, gotExp, gotGold, seen, steps)
}
