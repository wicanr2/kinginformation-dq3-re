package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
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
	g.newGame.ni.Init() // 原版新遊戲進入姓名畫面預設注音模式(raw0=ㄅ)
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

	// 四人隊共用裝備入口：先選角色，再顯示該角色四槽與背包候選。
	g.heroName = []int{15}
	g.equip = [4]int{0x03, 0x1e, -1, -1}
	g.companions = []*Member{
		newLevelOneMember(classNames[1], 1, 0, &g.prng, g.tavern.equipment),
		newLevelOneMember(classNames[3], 3, 0, &g.prng, g.tavern.equipment),
		newLevelOneMember(classNames[4], 4, 0, &g.prng, g.tavern.equipment),
	}
	g.inventory = []int{0x01, 0x1f, 0x3a, 0x22}
	g.panel, g.panelActor, g.panelCursor = panelEquip, -1, 0
	dump("equipment_actor_select")
	g.panelActor, g.panelCursor = 1, 0
	dump("equipment_party_inventory")
	g.panel = panelNone

	// 原版 D3TXT00 record407 的詳細狀況窗：正式命令入口只顯示
	// pack-owned frame／標籤與主角角色 record 數值，不注入 debug 進度欄位。
	g.panel = panelStatus
	dump("status_detail")
	g.panel = panelNone

	// CTY10 sec5 原版 handler14：甘達特×1 + 手下×3 的混合編隊。
	kandarEvent := g.pack.BossSurrenderEvents()[0]
	kandar, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		kandarEvent.Trigger.CTYRaw, mapBlkNum[kandarEvent.Trigger.CTYRaw],
		kandarEvent.Trigger.Section,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = kandar, kandar, kandarEvent.Trigger.CTYRaw, true
	g.px, g.py, g.dlg.tx, g.dlg.open = 6, 8, kandar.dlgText, false
	g.bossSurrenderEventID = kandarEvent.ID
	g.startBossSurrenderBattle()
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatal("甘達特原版混合編隊未啟動")
	}
	dump("kandar_mixed_battle")
	g.battle.active = false
	g.bossSurrenderStage, g.bossSurrenderCursor = bossSurrenderChoice, 0
	dump("kandar_surrender_choice")
	g.bossSurrenderStage, g.bossSurrenderEventID = bossSurrenderIdle, ""

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
	g.endSeq, g.dlg.open, g.panel, g.showTitle, g.lotoBlessed = -1, false, panelNone, false, false
	g.layer, g.cur, g.inTown, g.curCty = 0, g.over, false, -1
	g.px, g.py = ctyLoc[0][0]-1, ctyLoc[0][1]
	g.heroExp, g.heroMP = stats.ExpForLevel(0, 14), 99
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 35))
	mage.CurMP = 99
	g.companions = []*Member{mage}
	g.visitedTowns = []townVisit{{Cty: 0}, {Cty: 1}, {Cty: 2}}
	g.openFieldSpellMenu()
	dump("field_spell_menu")
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	dump("rura_destinations")

	// 逐隊員戰鬥命令：隊長下令後輪到 Lv40 魔法使同伴；其選單不可出現逃跑。
	// 再開該同伴的五行捲動咒文窗，清單尾端巴魯朋特必須可見。
	g.fieldSpell = FieldSpellMenu{}
	mageSpells := spell.Known(4, 40)
	if !g.battle.startGroup(5, 1, 1, heroParams{
		level: 40, curHP: 200, maxHP: 200, atk: 80, def: 80, agi: 80,
		mp: 80, maxMP: 80, spells: spell.Known(0, 40),
	}, []*battleActor{{
		class: 4, level: 40, hp: 150, maxHP: 150, mp: 200, maxMP: 200,
		atk: 40, def: 50, agi: 90, spells: mageSpells,
	}}) {
		t.Fatal("巴魯朋特選單截圖開戰失敗")
	}
	g.battle.commandActor = 1
	g.battle.phase = phCommand
	dump("battle_party_command")
	g.battle.phase = phTargetAlly
	g.battle.targetCursor = 1
	dump("battle_ally_target")
	g.battle.phase = phSpell
	g.battle.spellCursor = len(mageSpells) - 1
	dump("battle_palpunte_menu")

	// 三敵單體選擇：游標與觸控 hit-box 對應實際怪物位置，不再固定第一隻。
	if !g.battle.startGroup(5, 3, 1, heroParams{
		level: 20, curHP: 150, maxHP: 150, atk: 60, def: 60, agi: 60,
	}, nil) {
		t.Fatal("敵方目標選擇截圖開戰失敗")
	}
	g.battle.phase = phTargetEnemy
	g.battle.targetCursor = 1
	dump("battle_enemy_target")

	// 回合結算訊息：逐筆顯示行動結果，確認後才前進到勝利／下一回合訊息。
	if !g.battle.startGroup(5, 1, 1, heroParams{
		level: 20, curHP: 150, maxHP: 150, atk: 999, def: 60, agi: 999,
	}, nil) {
		t.Fatal("戰鬥訊息截圖開戰失敗")
	}
	g.battle.enemies[0].hp, g.battle.enemies[0].max, g.battle.enemies[0].agi = 1, 1, 1
	g.battle.commands[0] = battleCommand{kind: bcWar, target: 0}
	g.battle.resolveRound()
	dump("battle_message_queue")

	// 第一個主線道具：CTY08 sec3 正式「話す」老人，顯示取得盜賊鑰匙的 runtime 狀態。
	najimi, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 8, mapBlkNum[8], 3, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.battle.active, g.inTown, g.curCty = false, true, 8
	g.town, g.cur, g.dlg.tx = najimi, najimi, najimi.dlgText
	g.inventory = nil
	delete(g.flags, msThiefKey)
	g.dlg.open, g.cmd.open = false, false
	positioned := false
	for dir := 0; dir < 4 && !positioned; dir++ {
		dx, dy := dirDelta(dir)
		for dist := 1; dist <= 2; dist++ {
			x, y := 9-dx*dist, 9-dy*dist
			if x < 0 || y < 0 || x >= najimi.w || y >= najimi.h || najimi.Blocked(x, y) {
				continue
			}
			if dist == 2 && !najimi.attr.Blocked(najimi.tileIdx(9-dx, 9-dy)) {
				continue
			}
			g.px, g.py, g.facing, g.cd = x, y, dir, 0
			positioned = true
			break
		}
	}
	if !positioned {
		t.Fatal("拿吉米老人周圍沒有原版可交談位置")
	}
	g.selectCommand(cmdTalk)
	if !g.hasItem(0x55) || !g.dlg.open {
		t.Fatal("拿吉米老人 runtime 圖未觸發盜賊鑰匙事件")
	}
	dump("najimi_thief_key")

	// 雷貝右上屋二樓：持 0x55 與 handler7 正式交談，取得且顯示魔法球 0x58。
	leve, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 1, mapBlkNum[1], 1, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.town, g.cur, g.dlg.tx = 1, leve, leve, leve.dlgText
	g.inventory = []int{0x55}
	delete(g.flags, msMagicBal)
	g.dlg.open, g.cmd.open = false, false
	positioned = false
	for dir := 0; dir < 4 && !positioned; dir++ {
		dx, dy := dirDelta(dir)
		for dist := 1; dist <= 2; dist++ {
			x, y := 3-dx*dist, 3-dy*dist
			if x < 0 || y < 0 || x >= leve.w || y >= leve.h || leve.Blocked(x, y) {
				continue
			}
			if dist == 2 && !leve.attr.Blocked(leve.tileIdx(3-dx, 3-dy)) {
				continue
			}
			g.px, g.py, g.facing, g.cd = x, y, dir, 0
			positioned = true
			break
		}
	}
	if !positioned {
		t.Fatal("雷貝老人周圍沒有原版可交談位置")
	}
	g.selectCommand(cmdTalk)
	if !g.hasItem(0x58) || !g.hasItem(0x55) || !g.dlg.open {
		t.Fatal("雷貝老人 runtime 圖未觸發魔法球事件，或錯誤消耗盜賊鑰匙")
	}
	dump("leve_magic_ball")

	// 誘惑洞窟入口：在原版精確座標使用魔法球後，0x51 清除並由通用 event
	// rebuild 將 2x2 牆 tile 24/22 改為 25/23。
	g.setStoryFlag(magicBallIntactFlag, true)
	cave, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyTemptationCave, mapBlkNum[ctyTemptationCave], magicBallSection, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.town, g.cur, g.dlg.tx = ctyTemptationCave, cave, cave, cave.dlgText
	g.inventory = []int{itemuse.ItemMagicBall}
	g.px, g.py = magicBallLeftX, magicBallUseY
	g.dlg.open, g.cmd.open = false, false
	g.useMagicBall()
	if g.hasItem(itemuse.ItemMagicBall) || g.storyFlag(magicBallIntactFlag) {
		t.Fatal("誘惑洞窟 runtime 圖未完成魔法球 transaction")
	}
	dump("temptation_magic_ball_open")

	// 魔法球後正常穿越 CTY30/31 所抵達的下一個 campaign checkpoint。
	// framedump 只固定呈現狀態；production 可達性由 TestOpeningProductionInputTrace 證明。
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1
	g.px, g.py = ctyLoc[2][0], ctyLoc[2][1]
	g.noticeTimer = 0 // production 穿越 CTY30/31 的步數早已令魔法球取得通知逾時
	g.enterTownCty(2)
	if !g.inTown || g.curCty != 2 || sceneSection(g.cur) != 0 {
		t.Fatal("羅馬利亞 runtime 圖載入失敗")
	}
	dump("romaly_arrival")

	roleEvent := mustTemporaryRoleEvent(t, g)
	romalyThrone, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		roleEvent.OfferNPC.CTYRaw, mapBlkNum[roleEvent.OfferNPC.CTYRaw],
		roleEvent.OfferNPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.dlg.tx = romalyThrone, romalyThrone, romalyThrone.dlgText
	g.setStoryFlag(roleEvent.PendingFlagRaw, true)
	g.removeItems(roleEvent.RequiredItemRawID, g.countItem(roleEvent.RequiredItemRawID))
	var romalyKing *npcInst
	for i := range romalyThrone.npcs {
		if romalyThrone.npcs[i].b4 == roleEvent.OfferNPC.HandlerRaw &&
			(romalyThrone.npcs[i].ctrl>>3)&7 == 2 {
			romalyKing = &romalyThrone.npcs[i]
			break
		}
	}
	if romalyKing == nil || !g.talkTemporaryRole(romalyKing) || !g.dlg.open {
		t.Fatal("羅馬利亞國王任務 runtime 圖未觸發")
	}
	g.px, g.py = 7, 3
	g.facing = 1
	dump("romaly_king_crown_quest")

	// 交還金皇冠後的強制王位選擇，及接受後由 canonical active-role flag
	// 派生的國王角色圖。對白、選項與角色圖 entry 均來自 game pack。
	g.dlg.open = false
	g.finishTemporaryRoleEvent()
	g.inventory = append(g.inventory, roleEvent.RequiredItemRawID)
	if !g.talkTemporaryRole(romalyKing) || !g.dlg.open {
		t.Fatal("羅馬利亞還冠 runtime 圖未觸發")
	}
	dump("romaly_crown_return")
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue() // return_praise → forced_offer
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue() // forced_offer → forced_choice
	if !g.temporaryRoleChoosing() {
		t.Fatal("羅馬利亞強制王位選項未開啟")
	}
	dump("romaly_kingship_choice")
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue()
	if !g.storyFlag(roleEvent.ActiveRoleFlagRaw) || g.heroRole == nil {
		t.Fatal("接受王位後未切換 game-pack 角色圖")
	}
	dump("romaly_temporary_king")

	// 原版影片 12:42–13:14：地下競技場左上前國王的兩階段辭位確認。
	romalyArena, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		roleEvent.RestoreNPC.CTYRaw, mapBlkNum[roleEvent.RestoreNPC.CTYRaw],
		roleEvent.RestoreNPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.dlg.tx = romalyArena, romalyArena, romalyArena.dlgText
	g.curCty, g.inTown = roleEvent.RestoreNPC.CTYRaw, true
	var formerKing *npcInst
	for i := range romalyArena.npcs {
		n := &romalyArena.npcs[i]
		if n.x == roleEvent.RestoreNPC.Tile.X &&
			n.y == roleEvent.RestoreNPC.Tile.Y &&
			n.b4 == roleEvent.RestoreNPC.HandlerRaw && (n.ctrl>>3)&7 == 2 {
			formerKing = &romalyArena.npcs[i]
			break
		}
	}
	if formerKing == nil || !g.talkTemporaryRole(formerKing) || !g.dlg.open {
		t.Fatal("地下競技場前國王 runtime 圖未觸發")
	}
	g.px, g.py, g.facing = roleEvent.RestoreNPC.Tile.X,
		roleEvent.RestoreNPC.Tile.Y+1, 1
	dump("romaly_restore_intro")
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue()
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: 0})
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue()
	if g.temporaryRoleStage != temporaryRoleRestoreConfirm {
		t.Fatal("地下競技場未進入第二層辭位確認")
	}
	dump("romaly_restore_confirm")

	// 下一個正式切片：諾亞尼爾沉睡村、精靈女王以紅寶石換覺醒粉、
	// 以及原版 SET 0x26 / CLEAR 0x31 後即時重載的清醒村莊。
	g.finishTemporaryRoleEvent()
	g.setStoryFlag(roleEvent.ActiveRoleFlagRaw, false)
	g.setStoryFlag(roleEvent.NormalRoleFlagRaw, true)
	g.syncTemporaryRoleVisual()
	questEvents := g.pack.QuestItemChainEvents()
	if len(questEvents) != 1 {
		t.Fatalf("quest item chain event count=%d, want 1", len(questEvents))
	}
	noaniel := questEvents[0]
	sleeping, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		noaniel.Use.CTYRaw, mapBlkNum[noaniel.Use.CTYRaw], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.inTown, g.town, g.cur = noaniel.Use.CTYRaw, true, sleeping, sleeping
	g.dlg.tx, g.dlg.open = sleeping.dlgText, false
	g.px, g.py, g.facing = 15, 20, 1
	dump("noaniel_sleeping")

	exchange := noaniel.Exchange
	fairyVillage, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		exchange.NPC.CTYRaw, mapBlkNum[exchange.NPC.CTYRaw], exchange.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.town, g.cur = exchange.NPC.CTYRaw, fairyVillage, fairyVillage
	g.dlg.tx, g.dlg.open = fairyVillage.dlgText, false
	g.px, g.py, g.facing = exchange.NPC.Tile.X, exchange.NPC.Tile.Y+1, 1
	g.inventory = append(g.inventory, exchange.RequiredItemRawID)
	var fairyQueen *npcInst
	for i := range fairyVillage.npcs {
		if g.scriptedNPCMatches(&fairyVillage.npcs[i], exchange.NPC) {
			fairyQueen = &fairyVillage.npcs[i]
			break
		}
	}
	if fairyQueen == nil || !g.talkQuestItemChain(fairyQueen) ||
		!g.hasItem(exchange.GrantedItemRawID) || !g.dlg.open {
		t.Fatal("精靈女王換物 runtime 圖未完成 pack transaction")
	}
	dump("fairy_queen_ruby_exchange")

	sleeping, err = loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		noaniel.Use.CTYRaw, mapBlkNum[noaniel.Use.CTYRaw], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.town, g.cur = noaniel.Use.CTYRaw, sleeping, sleeping
	g.dlg.tx, g.dlg.open = sleeping.dlgText, false
	g.px, g.py, g.facing = 15, 20, 1
	g.panel, g.panelCursor = panelItem, 0
	g.inventory = []int{noaniel.Use.ItemRawID}
	g.useSelectedItem()
	if !g.storyFlag(noaniel.Use.SetFlagsRaw[0]) ||
		g.storyFlag(noaniel.Use.ClearFlagsRaw[0]) || !g.dlg.open {
		t.Fatal("諾亞尼爾甦醒 runtime 圖未完成 pack flag transaction")
	}
	dump("noaniel_awakened")

	// 金字塔雙開關：全部 selector、文字、旗標、音效與魔法鑰匙均取自
	// two_step_floor_switch_gate pack；畫面使用正式 production renderer。
	gates := g.pack.TwoStepFloorSwitchGates()
	if len(gates) != 1 {
		t.Fatalf("two-step floor switch gate count=%d, want 1", len(gates))
	}
	gate := gates[0]
	g.setStoryFlag(gate.ClearFlagRaw, true)
	g.setStoryFlag(gate.UnlockedTreasure.PresentFlag, true)
	pyramid, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		gate.CTYRaw, mapBlkNum[gate.CTYRaw], gate.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.curCty, g.inTown, g.town, g.cur = gate.CTYRaw, true, pyramid, pyramid
	g.dlg.tx, g.dlg.open = pyramid.dlgText, false
	first := gate.Switches[0]
	var completion = gate.Switches[len(gate.Switches)-1]
	g.px, g.py, g.facing = first.Tile.X, first.Tile.Y, 0
	if !g.trySequenceGateEvent() || !g.dlg.open {
		t.Fatal("金字塔開關 runtime 圖未觸發 pack prompt")
	}
	dump("pyramid_switch_prompt")
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
	g.sequenceGateChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
	g.px, g.py = completion.Tile.X, completion.Tile.Y
	if !g.trySequenceGateEvent() {
		t.Fatal("金字塔完成開關未觸發")
	}
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
	g.sequenceGateChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
	if g.storyFlag(gate.ClearFlagRaw) || !g.dlg.open {
		t.Fatal("金字塔開門 runtime 圖未完成 pack transaction")
	}
	g.px, g.py, g.facing = 13, 12, 1
	dump("pyramid_magic_gate_open")
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
	g.px, g.py, g.facing = 13, 6, 1
	g.examine()
	if !g.hasItem(gate.UnlockedTreasure.ItemRawID) ||
		g.storyFlag(gate.UnlockedTreasure.PresentFlag) {
		t.Fatal("魔法鑰匙 runtime 圖未由 pack treasure selector 取得")
	}
	dump("pyramid_magic_key")
}

// TestDumpChurchRevive 產出 production renderer 的教會復活確認窗。它不是 raw glyph dump；
// modal、原始文字 bank、隊伍狀態與 JSON 費用都由實際 Game 組合。
func TestDumpChurchRevive(t *testing.T) {
	if os.Getenv("DQ3_DUMP_CHURCH") == "" {
		t.Skip("設 DQ3_DUMP_CHURCH=1 才執行")
	}
	out := os.Getenv("CHURCH_OUT")
	if out == "" {
		t.Fatal("需設 CHURCH_OUT")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.inTown, g.curCty = false, true, 0
	g.cur, g.town = g.towns[0], g.towns[0]
	g.px, g.py = 15, 25
	g.heroGold, g.heroHP = 34, 1
	dead := newMember(classNames[3], 3, 0, 0)
	dead.CurHP, dead.CurMP = 0, 0
	g.companions = []*Member{dead}
	g.church.open(g.shop.nameText)
	g.churchInput(InputState{DirEdge: 0})
	g.churchInput(InputState{DirEdge: 0})
	g.churchInput(InputState{Confirm: true})
	g.churchInput(InputState{DirEdge: 0})
	g.churchInput(InputState{Confirm: true})
	if g.church.stage != churchConfirm || g.church.pendingCost != 10 {
		t.Fatalf("church fixture not at confirmation: stage=%d cost=%d",
			g.church.stage, g.church.pendingCost)
	}
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	copy(img.Pix, g.rgba)
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	t.Logf("church revive runtime image: %s", out)
}
