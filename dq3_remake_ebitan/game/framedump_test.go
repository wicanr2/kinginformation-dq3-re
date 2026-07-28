package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// TestDumpNewGameScreens:把新遊戲流程各畫面(主選單/命名/性別/能力確認)dump 成 PNG,
// 供與 docs/36_shots 原版實機截圖肉眼對照(glyph 形近誤標高風險 → 視覺驗證,規則 63)。
// 平常 skip;設 DQ3_DUMP_NG=1 + NG_OUT=<gitignored 夾> 才跑。
func TestDumpNewGameScreens(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NG") == "" {
		t.Skip("設 DQ3_DUMP_NG=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("NG_OUT")
	if out == "" {
		t.Fatal("需設 NG_OUT")
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
	// 依新遊戲狀態機逐畫面 dump(直接設 state,不跑 ebiten 主迴圈)
	g.newGame.stage = ngMenu
	dump("ng_menu")
	g.newGame.stage = ngName
	dump("ng_name")
	g.newGame.stage = ngGender
	dump("ng_gender")
	g.newGame.ni.nameBuf = []int{15} // 英數「A」
	g.heroName = append([]int(nil), g.newGame.ni.nameBuf...)
	g.rollHeroLevelOne()
	g.newGame.preview = g.heroStat
	g.newGame.previewDef = int(g.heroStat[stats.VIT]) + g.shop.items.Defense(g.equip[1])
	g.newGame.stage = ngConfirm
	dump("ng_confirm")

	// C-2:模擬創角完成 → 進主角家(sec4)+ 原版開場 runner 序列。
	g.showTitle = false
	g.startOpening()
	dump("opening_home_rec82") // 家室內 + 旁白第一句(rec82)
	g.dlg.Open(83)
	dump("opening_home_rec83")
	g.dlg.Open(81)
	dump("opening_home_rec81")

	// handler54:母親帶出門 → sec0 (8,38) → rec80。
	g.motherEscort()
	g.dlg.Open(openingMotherDirectionsRec)
	dump("opening_town_rec80")

	// handler56:走到阿里阿罕王座正前方 → rec78 + 50G/六件裝備。
	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty = throne, throne, ctyAliahanCastle
	g.px, g.py, g.facing = aliahanKingX, aliahanKingY+1, 1
	if !g.tryOpeningRegionEvent() {
		t.Fatal("王座 opening region 未觸發")
	}
	dump("opening_king_rec78")
	t.Logf("開場:謁見後 cty=%d sec=%d @(%d,%d), gold=%d items=%v",
		g.curCty, g.cur.sec, g.px, g.py, g.heroGold, g.inventory)

	// CTY00 原版 sub2 handler 身分接線後的兩個可達酒場畫面。
	luida, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = luida, luida, 0, true
	g.px, g.py, g.facing = 2, 18, 1 // 隔 row17 櫃台面向 b4=1 露依達
	g.dlg.open = false
	g.recruit.open()
	dump("luida_recruit")
	g.recruit.active = false

	registry, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 2, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town = registry, registry
	g.px, g.py, g.facing = 2, 5, 1 // 隔 row4 櫃台面向 b4=3 登錄員
	g.tavern.open()
	dump("adventurer_registry")
	g.tavern.active = false

	// R-2:沙曼歐莎夜間在原版精確座標使用拉之鏡 → rec97 揭露 → 怪力魔89。
	g.dnPhase = 2
	g.setStoryFlag(0x42, true)
	g.setStoryFlag(0x10, false)
	samanosa, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctySamanosa, mapBlkNum[ctySamanosa], mirrorSection, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = samanosa, samanosa, ctySamanosa, true
	g.px, g.py, g.inventory = mirrorX, mirrorY, []int{0x61}
	g.dlg.tx = samanosa.dlgText
	g.useMirror()
	dump("samanosa_mirror_reveal")

	g.dlg.open, g.mirrorStage = false, 2
	g.advanceMirrorEvent()
	if !g.battle.active || g.battle.monID != monsterBossTroll {
		t.Fatal("拉之鏡揭露後未進怪力魔89戰")
	}
	dump("samanosa_boss_troll")

	// R-3:原版 CTY70 六寶珠祭壇 → 守護神復活拉米亞 → 地表搭乘飛行。
	g.battle.active = false
	g.dnPhase = 0
	for i := 0; i < 6; i++ {
		g.setStoryFlag(phoenixAltarFlagFirst+i, false)
		g.setStoryFlag(phoenixVisualFlagBase+i, false)
	}
	g.setStoryFlag(flagPhoenixUnrevived, true)
	phoenixShrine, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyPhoenixShrine, mapBlkNum[ctyPhoenixShrine], 0, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = phoenixShrine, phoenixShrine, ctyPhoenixShrine, true
	g.px, g.py, g.dlg.tx, g.dlg.open = 8, 12, phoenixShrine.dlgText, false
	dump("phoenix_six_orbs")

	g.phoenixList = g.placedPhoenixOrbs()
	g.revivePhoenix()
	g.phoenixAnim = 4
	dump("phoenix_revival_animation")
	g.phoenixAnim, g.phoenixAnimTick = 5, 7
	g.advancePhoenixAnimation()
	g.px, g.py = 8, 12
	g.dlg.Open(94)
	dump("phoenix_revived")

	g.dlg.open = false
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1
	g.px, g.py, g.phoenixAboard, g.facing = phoenixParkX, phoenixParkY, true, 1
	dump("phoenix_flight")

	// R-4:CTY65 巴拉摩斯單戰 → 阿里阿罕索瑪現身 → CTY72/77 自然下降。
	g.phoenixAboard = false
	baramos, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 65, mapBlkNum[65], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = baramos, baramos, 65, true
	g.px, g.py, g.facing, g.dlg.tx = 8, 4, 1, baramos.dlgText
	g.dlg.Open(85)
	dump("baramos_challenge")
	g.dlg.open = false
	if !g.startBossBattle(0x79) {
		t.Fatal("巴拉摩斯 0x79 戰鬥素材載入失敗")
	}
	dump("baramos_battle")

	g.battle.active = false
	throne, err = loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = throne, throne, ctyAliahanCastle, true
	g.px, g.py, g.facing, g.dlg.tx = aliahanKingX, aliahanKingY+1, 1, throne.dlgText
	g.dlg.Open(99)
	dump("zoma_appears_aliahan")

	g.dlg.open = false
	gaia, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 72, mapBlkNum[72], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = gaia, gaia, 72, true
	g.px, g.py, g.facing, g.dlg.tx = 14, 6, 0, gaia.dlgText
	dump("gaia_pit_open")

	underGate, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 77, mapBlkNum[77], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = underGate, underGate, 77, true
	g.px, g.py, g.facing, g.dlg.tx = 16, 9, 0, underGate.dlgText
	dump("alefgard_fall_landing")

	g.cur, g.inTown, g.curCty, g.layer = g.loadUnder(), false, -1, 1
	g.px, g.py = 85, 67
	g.dlg.open = false
	dump("alefgard_overworld")

	// R-5:龍女王原版 handler52 給光之珠並更新 story flags。
	queen, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyDragonQueen, mapBlkNum[ctyDragonQueen], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = queen, queen, ctyDragonQueen, true
	g.px, g.py, g.facing, g.dlg.tx = 14, 25, 1, queen.dlgText
	g.dlg.open = false
	g.removeItems(itemLightOrb, g.countItem(itemLightOrb))
	g.scriptedTalk(dragonQueenHandler)
	dump("dragon_queen_light_orb")

	// 彩虹水滴精確座標：(127,117) 使用前後，西側 (126,117) 變 tile0x53。
	g.dlg.open = false
	g.cur, g.inTown, g.curCty, g.layer = g.loadUnder(), false, -1, 1
	g.px, g.py, g.facing = rainbowUseX, rainbowUseY, 2
	g.worldState &^= worldStateRainbowBridge
	if g.cur.override != nil {
		delete(g.cur.override, rainbowBridgeY*g.cur.w+rainbowBridgeX)
	}
	dump("rainbow_bridge_before")
	g.inventory = append(g.inventory, itemRainbowDrop)
	g.panelCursor = len(g.inventory) - 1
	g.useSelectedItem()
	dump("rainbow_bridge_after")

	// R-5b:CTY90 王座 type4/e5 隱藏樓梯，sub_18C01 調查後 tile+1。
	g.setStoryFlag(0xe5, true)
	hidden, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyZomaCastle, mapBlkNum[ctyZomaCastle], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = hidden, hidden, ctyZomaCastle, true
	g.px, g.py, g.facing, g.dlg.tx = 23, 4, 1, hidden.dlgText
	g.dlg.open, g.noticeTimer = false, 0
	dump("zoma_hidden_stair_before")
	g.examine()
	dump("zoma_hidden_stair_revealed")

	// R-5b:橋上自動觸發 handler79→90；e0 的戰鬥 sprite 換成 flag0e 的垂死歐魯迪卡。
	g.dlg.open = false
	g.setStoryFlag(ortegaFightFlag, true)
	g.setStoryFlag(ortegaDyingFlag, false)
	ortega, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyZomaCastle, mapBlkNum[ctyZomaCastle], ortegaSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = ortega, ortega, ctyZomaCastle, true
	g.px, g.py, g.facing, g.dlg.tx = ortegaTriggerX, ortegaTriggerY, 2, ortega.dlgText
	g.noticeTimer = 0
	if !g.tryOrtegaEvent() {
		t.Fatal("歐魯迪卡橋上事件未觸發")
	}
	dump("ortega_bridge_rec69")
	g.dlg.open = false
	g.advanceOrtegaEvent()
	dump("ortega_dying_rec70")

	// CTY90 sec5 自然「話す」入口與原版三個 formation。
	zoma, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyZomaCastle, mapBlkNum[ctyZomaCastle], zomaFinalSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = zoma, zoma, ctyZomaCastle, true
	g.px, g.py, g.facing, g.dlg.tx = 11, 6, 1, zoma.dlgText
	g.dlg.open, g.noticeTimer = false, 0
	dump("zoma_final_room")
	if !g.startBossBattle(0x7a) {
		t.Fatal("巴拉摩斯怨靈 0x7a 戰鬥素材載入失敗")
	}
	dump("baramos_ghost_battle")
	g.battle.active = false
	if !g.startBossBattle(0x7b) {
		t.Fatal("巴拉摩斯殭屍 0x7b 戰鬥素材載入失敗")
	}
	dump("baramos_zombie_battle")
	g.battle.active = false
	g.dlg.Open(zomaDialogueRec)
	dump("zoma_final_challenge")
	g.dlg.open = false
	if !g.startBossBattle(0x7c) {
		t.Fatal("索瑪 0x7c 戰鬥素材載入失敗")
	}
	dump("zoma_final_battle")

	// handler80 勝利後回索瑪城外；下層傳送到 CTY79，再由 CTY80 handler74 國王冊封。
	g.battle.active = false
	g.runFinale()
	dump("zoma_aftermath_overworld")
	g.inventory = append(g.inventory, 0x43)
	g.panel, g.panelCursor = panelItem, len(g.inventory)-1
	g.useSelectedItem()
	dump("radatome_return")

	radatome, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyRadatomeCastle, mapBlkNum[ctyRadatomeCastle], radatomeThroneSec, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = radatome, radatome, ctyRadatomeCastle, true
	g.dlg.tx = radatome.dlgText
	for i := range radatome.npcs {
		if radatome.npcs[i].b4 == radatomeKingHandle {
			g.px, g.py, g.facing = radatome.npcs[i].x, radatome.npcs[i].y+1, 1
			break
		}
	}
	g.selectCommand(cmdTalk)
	dump("radatome_king_loto")
	g.dlg.open = false
	g.advanceEndingKing()
	dump("ending_scroll_start")

	// 野外正式「咒文」：已學工具咒清單 → 魯拉已造訪城鎮目的地。
	// 背景重置到阿里阿罕地表，避免沿用結局 modal／palette 污染新圖。
	g.endSeq, g.dlg.open, g.panel = -1, false, panelNone
	g.layer, g.cur, g.inTown, g.curCty = 0, g.over, false, -1
	g.px, g.py = ctyLoc[0][0]-1, ctyLoc[0][1]
	g.heroExp, g.heroMP = stats.ExpForLevel(0, 14), 99
	g.visitedTowns = []townVisit{{Cty: 0}, {Cty: 1}, {Cty: 2}}
	g.openFieldSpellMenu()
	dump("field_spell_menu")
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	dump("rura_destinations")
}
