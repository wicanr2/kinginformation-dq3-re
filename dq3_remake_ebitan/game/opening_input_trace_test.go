package game

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

// TestOpeningProductionInputTrace 從標題開始，只送入 production InputState：
// 標題→開始→英數命名→性別→四段開場→王座→酒場四人隊→出城→拿吉米之塔→盜賊鑰匙。
// 路徑搜尋只負責選下一個方向鍵，不直接修改 Game 狀態；每一步仍經 Game.step、
// 碰撞、NPC、轉場與 runner region。這是第一條真正的玩家可達 E3 開場 trace。
func TestOpeningProductionInputTrace(t *testing.T) {
	dir := spineAssetsDir(t)
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "opening-input-trace.json"))

	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	roleEvent := mustTemporaryRoleEvent(t, g)
	questEvents := g.pack.QuestItemChainEvents()
	if len(questEvents) != 1 {
		t.Fatalf("quest item chain event count=%d, want 1", len(questEvents))
	}
	noanielEvent := questEvents[0]
	sequenceGates := g.pack.TwoStepFloorSwitchGates()
	if len(sequenceGates) != 1 {
		t.Fatalf("two-step floor switch gate count=%d, want 1", len(sequenceGates))
	}
	pyramidGate := sequenceGates[0]
	// 此測試驗證數千次正式 InputState 的連續狀態路線；逐步 WritePixels 會在沒有
	// ebiten 顯示迴圈的 test binary 累積圖形命令。畫面另由 framedump 測試驗證。
	g.frame = nil
	send := func(in InputState) {
		t.Helper()
		if in.DirHeld == 0 && in.DirEdge == 0 && !in.Confirm && !in.Cancel &&
			!in.Enter && !in.Toggle && !in.Settings && !in.CtxTap && !in.Tapped {
			// InputState 零值的方向是 0(下)，無輸入必須明確使用 -1。
			in.DirHeld, in.DirEdge = -1, -1
		}
		if err := g.step(in); err != nil {
			t.Fatalf("step: %v", err)
		}
	}
	press := func(in InputState) {
		if in.DirHeld == 0 {
			in.DirHeld = -1
		}
		if in.DirEdge == 0 {
			in.DirEdge = -1
		}
		send(in)
	}

	press(InputState{Confirm: true}) // splash → menu
	press(InputState{Confirm: true}) // menu「遊戲開始」→ 注音命名
	press(InputState{Toggle: true})  // 注音 → 英數
	press(InputState{Confirm: true}) // cursor0：「0」
	send(InputState{DirHeld: -1, DirEdge: 2})
	press(InputState{Confirm: true}) // cursor37：OK → 性別
	press(InputState{Confirm: true}) // 男性 → 能力確認
	press(InputState{Confirm: true}) // 「是」→ startOpening

	if g.showTitle || g.newGame.stage != ngConfirm || !g.inTown || g.cur.sec != 4 ||
		g.px != openingHomeX || g.py != openingHomeY {
		t.Fatalf("創角輸入後應落在家中：title=%v stage=%d cty=%d sec=%d @(%d,%d)",
			g.showTitle, g.newGame.stage, g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	if len(g.heroName) != 1 || g.heroName[0] != niCellGlyph(0) {
		t.Fatalf("輸入名稱未經 production flow 寫回：%v", g.heroName)
	}

	for _, wantIdx := range []int{0, 1, 2, 3} {
		if g.openingIdx != wantIdx || !g.dlg.open {
			t.Fatalf("開場段 %d 未開啟：idx=%d open=%v", wantIdx, g.openingIdx, g.dlg.open)
		}
		traceCloseDialogue(t, g)
		send(InputState{DirHeld: -1, DirEdge: -1}) // opening runner 開下一段
	}
	if g.openingIdx != -1 || g.dlg.open || g.curCty != 0 || g.cur.sec != 0 ||
		g.px != 8 || g.py != 38 || !g.storyFlag(0x17) {
		t.Fatalf("母親演出後狀態錯：idx=%d dlg=%v cty=%d sec=%d @(%d,%d) flag17=%v",
			g.openingIdx, g.dlg.open, g.curCty, sceneSection(g.cur), g.px, g.py, g.storyFlag(0x17))
	}

	traceWalkThroughPortal(t, g, 21, 8, ctyAliahanCastle, 0)
	traceWalkThroughPortal(t, g, 15, 10, ctyAliahanCastle, aliahanThroneSection)
	traceWalkTo(t, g, aliahanKingX, aliahanKingY+1)

	if !g.progressDone(msStart) || g.heroGold != 50 ||
		len(g.inventory) != len(aliahanKingRewardItems) || !g.dlg.open {
		t.Fatalf("正式步行到王座後未完成謁見：ms=%v gold=%d items=%v dlg=%v",
			g.progressDone(msStart), g.heroGold, g.inventory, g.dlg.open)
	}

	// 謁見後正常步行回城鎮，從西側樓梯進 2F 冒險者登錄所。
	traceCloseDialogue(t, g)
	traceWalkThroughPortal(t, g, 9, 22, ctyAliahanCastle, 0)
	traceWalkThroughPortal(t, g, 15, 31, 0, 0)
	traceWalkThroughPortal(t, g, 8, 14, 0, 2)

	// 用正式命令窗對 b4=3 NPC 登錄三人；每次完成後 modal 關閉，再重新對話。
	for i, class := range []int{1, 3, 4} { // 戰士、僧侶、魔法使者
		traceTalkNPC(t, g, 2, 3)
		if !g.tavern.active {
			t.Fatalf("第%d次對冒險者登錄所 NPC 應開 tavern modal", i+1)
		}
		for j := 0; j < class; j++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true}) // 職業
		press(InputState{Toggle: true})  // 英數
		press(InputState{Confirm: true}) // 輸入「0」
		send(InputState{DirHeld: -1, DirEdge: 2})
		press(InputState{Confirm: true}) // OK
		press(InputState{Confirm: true}) // 男性→完成
	}
	if len(g.roster) != 3 || len(g.companions) != 0 {
		t.Fatalf("登錄三人後應 roster=3/companions=0，得 %d/%d", len(g.roster), len(g.companions))
	}

	// 回樓下，對 b4=1 露依達，從名冊連續找三名同伴加入。
	traceWalkThroughPortal(t, g, 8, 2, 0, 0)
	traceTalkNPC(t, g, 2, 16)
	if !g.recruit.active {
		t.Fatal("對露依達 NPC 應開 recruit modal")
	}
	press(InputState{Confirm: true}) // 找同伴參加
	for i := 0; i < 3; i++ {
		press(InputState{Confirm: true})
	}
	if len(g.companions) != 3 || len(g.roster) != 0 {
		t.Fatalf("招募後應成為主角+3 四人隊，得 companions=%d roster=%d",
			len(g.companions), len(g.roster))
	}
	press(InputState{Cancel: true}) // 清單→主選單
	press(InputState{Cancel: true}) // 關酒場

	// 四人旅店每次 8G；國王的 50G 全數保留作練級住宿。先買藥草會在 Lv4 前耗盡
	// 住宿費，因此這條 deterministic 正式路線以逃離強敵、旅店補滿為前段補給策略。

	// 國王給的裝備不會自動穿上；依正式命令窗逐人分配。
	equipItem := func(actor, code int) {
		t.Helper()
		idx := -1
		for i, item := range g.inventory {
			if item == code {
				idx = i
				break
			}
		}
		if idx < 0 {
			t.Fatalf("背包沒有待裝備 item %#x：%v", code, g.inventory)
		}
		press(InputState{Confirm: true}) // 開命令
		for g.cmd.cursor != int(cmdEquip) {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true}) // 開裝備面板
		for i := 0; i < actor; i++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true}) // 選隊員
		for i := 0; i < idx; i++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true}) // 裝上；舊品回背包
		press(InputState{Cancel: true})  // 回隊員選擇
		press(InputState{Cancel: true})  // 關裝備面板
	}
	equipItem(0, 0x03) // 勇者：銅劍
	equipItem(1, 0x01) // 戰士：木棒
	equipItem(1, 0x1f) // 戰士：旅人的衣服
	equipItem(2, 0x01) // 僧侶：木棒
	equipItem(2, 0x1f) // 僧侶：旅人的衣服
	equipItem(3, 0x00) // 魔法使者：檜木棒
	if g.equip[0] != 0x03 {
		t.Fatalf("正式裝備輸入未穿上國王給的銅劍：equip=%v", g.equip)
	}
	if got := [3][2]int{
		{g.companions[0].Weapon, g.companions[0].Armor},
		{g.companions[1].Weapon, g.companions[1].Armor},
		{g.companions[2].Weapon, g.companions[2].Armor},
	}; got != [3][2]int{{0x01, 0x1f}, {0x01, 0x1f}, {0x00, 0x1e}} {
		t.Fatalf("正式分配國王裝備錯：%v", got)
	}

	// 走回原始 section0 邊界 spawn，再向外一步正常出城；不可使用 Esc/debug exit。
	traceWalkTo(t, g, g.cur.spawnX, g.cur.spawnY)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 2, DirEdge: -1})
	if g.inTown || len(g.companions) != 3 {
		t.Fatalf("四人隊正常出城失敗：town=%v companions=%d @(%d,%d)",
			g.inTown, len(g.companions), g.px, g.py)
	}

	// 最短任務路線會讓四人一直停在 Lv1，抵達羅馬利亞北方後連單隻中期怪都無法承受。
	// 在阿里阿罕入口旁的原版低危 region 以正式移動／戰鬥／旅店輸入練到 Lv4；
	// 不寫入 EXP、HP、金幣或 RNG，所有成長與補給都必須走 production consumer。
	traceTrainNearTown(t, g, 0, 4)

	// 從阿里阿罕出口沿真實地表走到拿吉米之塔；途中若發生隨機遭遇，也只送正式戰鬥輸入。
	traceAdventureWalkToCty(t, g, 7) // 阿里阿罕西側地道入口；塔本體隔海，不能由地表直走
	traceTownSectionTo(t, g, 8, 3)
	traceTalkNPC(t, g, 9, 9) // CTY08 sec3 sub2 handler12：巴可達老人
	if !g.hasItem(0x55) || !g.progressDone(msThiefKey) || !g.dlg.open {
		t.Fatalf("正式對話後應取得盜賊鑰匙：item=%v milestone=%v dlg=%v",
			g.hasItem(0x55), g.progressDone(msThiefKey), g.dlg.open)
	}
	traceCloseDialogue(t, g)

	// 取得任務道具後的 checkpoint 必須可保存；不能只靠當前記憶體內 scripted 狀態。
	if err := g.Save(); err != nil {
		t.Fatalf("保存盜賊鑰匙 checkpoint: %v", err)
	}
	restored, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建讀檔 Game: %v", err)
	}
	if err := restored.Load(); err != nil {
		t.Fatalf("讀取盜賊鑰匙 checkpoint: %v", err)
	}
	if !restored.hasItem(0x55) || !restored.progressDone(msThiefKey) ||
		!restored.inTown || restored.curCty != 8 || sceneSection(restored.cur) != 3 {
		t.Fatalf("盜賊鑰匙 checkpoint round-trip 錯：item=%v milestone=%v cty=%d sec=%d",
			restored.hasItem(0x55), restored.progressDone(msThiefKey),
			restored.curCty, sceneSection(restored.cur))
	}

	// 從已取得 0x55 的合法狀態走回地表，再往雷貝 CTY01 的二樓老人取得魔法球。
	traceTownSectionTo(t, g, 7, 0) // 先回阿里阿罕側地道；塔本體的直接出口會落在隔海孤島
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, 1)
	traceOpenReachableDoor(t, g, 28, 8) // 原版雷貝右上屋的 tier-1 門
	traceTownSectionTo(t, g, 1, 1, 3, 3)
	traceTalkNPC(t, g, 3, 3) // CTY01 sec1 sub2 handler7
	if !g.hasItem(0x58) || !g.hasItem(0x55) || !g.progressDone(msMagicBal) || !g.dlg.open {
		t.Fatalf("正式雷貝老人對話交易錯：ball=%v key=%v milestone=%v dlg=%v",
			g.hasItem(0x58), g.hasItem(0x55), g.progressDone(msMagicBal), g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if err := g.Save(); err != nil {
		t.Fatalf("保存魔法球 checkpoint: %v", err)
	}
	if err := restored.Load(); err != nil {
		t.Fatalf("讀取魔法球 checkpoint: %v", err)
	}
	if !restored.hasItem(0x58) || !restored.hasItem(0x55) ||
		!restored.progressDone(msMagicBal) || restored.curCty != 1 || sceneSection(restored.cur) != 1 {
		t.Fatalf("魔法球 checkpoint round-trip 錯：ball=%v key=%v milestone=%v cty=%d sec=%d",
			restored.hasItem(0x58), restored.hasItem(0x55), restored.progressDone(msMagicBal),
			restored.curCty, sceneSection(restored.cur))
	}

	// 從雷貝正常出城，經 CTY09 祠堂入口進誘惑洞窟；在 IDA file 0x58cd 指定的
	// CTY30 sec0 @(8/9,13) 從正式道具面板使用 0x58。
	traceTownSectionTo(t, g, 1, 0)
	traceOpenReachableDoor(t, g, 28, 8) // section 重載後門回到原始關閉狀態，持 0x55 再正式開門
	traceTalkNPC(t, g, 5, 15)           // 雷貝旅社 k1：正式付費補滿長途迷宮前的 HP/MP
	traceTalkNPC(t, g, 5, 6)            // 雷貝道具店 k0：貨架第二項是藥草
	if !g.shop.active {
		t.Fatal("雷貝道具店未由正式 NPC 對話開啟")
	}
	// 買一瓶聖水(貨架 0x44)抑制前段已不足以承受的高區遭遇，再用餘款補藥草。
	holyIdx := -1
	for i, code := range g.shop.codes {
		if code == itemuse.ItemHolyWater {
			holyIdx = i
			break
		}
	}
	if holyIdx < 0 || g.heroGold < g.shop.items.Price(itemuse.ItemHolyWater) {
		t.Fatalf("雷貝補給不足以買聖水：idx=%d gold=%d codes=%v", holyIdx, g.heroGold, g.shop.codes)
	}
	for g.shop.cursor != holyIdx {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true})
	herbIdx := -1
	for i, code := range g.shop.codes {
		if code == herbCode {
			herbIdx = i
			break
		}
	}
	for g.shop.cursor != herbIdx {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	// 不把前段戰鬥收入全換成藥草；保留羅馬利亞住宿與下一段聖水預算。
	for i := 0; i < 2 && g.heroGold >= g.shop.items.Price(herbCode); i++ {
		press(InputState{Confirm: true})
	}
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceAdventureWalkToCty(t, g, 9)
	traceTownSectionTo(t, g, ctyTemptationCave, magicBallSection)
	traceWalkTo(t, g, magicBallLeftX, magicBallUseY)
	traceUseInventoryItem(t, g, itemuse.ItemMagicBall)
	if g.hasItem(itemuse.ItemMagicBall) || g.storyFlag(magicBallIntactFlag) {
		t.Fatalf("production 道具選單破牆後狀態錯：ball=%v flag51=%v",
			g.hasItem(itemuse.ItemMagicBall), g.storyFlag(magicBallIntactFlag))
	}
	for _, p := range [][3]int{{8, 14, 25}, {9, 14, 25}, {8, 15, 23}, {9, 15, 23}} {
		if got := g.cur.tileIdx(p[0], p[1]); got != p[2] {
			t.Fatalf("production 破牆 tile(%d,%d)=%d, want %d", p[0], p[1], got, p[2])
		}
	}

	// 穿越已打通的誘惑洞窟。CTY30 sec2 的入口與出口被原版 tier-1 門分隔；
	// 以仍持有的盜賊鑰匙從正式「調查」命令開左側門，再正常走 CTY31 到羅馬利亞。
	traceWalkThroughPortal(t, g, 9, 19, ctyTemptationCave, 1, 19, 28)
	traceTownSectionTo(t, g, ctyTemptationCave, 2)
	traceOpenReachableDoor(t, g, 6, 11)
	traceWalkThroughPortal(t, g, 4, 36, 31, 1, 2, 4)
	traceOpenReachableDoor(t, g, 4, 4)
	traceWalkThroughPortal(t, g, 8, 4, 31, 0, 4, 5)
	traceExitTownBoundary(t, g)
	traceAdventureWalkToCty(t, g, 2)
	if !g.inTown || g.curCty != 2 || sceneSection(g.cur) != 0 {
		t.Fatalf("魔法球後未自然抵達羅馬利亞：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	traceTownSectionTo(t, g, roleEvent.OfferNPC.CTYRaw, roleEvent.OfferNPC.Section,
		roleEvent.OfferNPC.Tile.X, roleEvent.OfferNPC.Tile.Y)
	traceTalkNPC(t, g, roleEvent.OfferNPC.Tile.X, roleEvent.OfferNPC.Tile.Y)
	if !g.dlg.open || g.temporaryRoleStage != temporaryRoleQuestGreeting ||
		!reflect.DeepEqual(g.dlg.buf,
			mustPackTextCodes(t, g, roleEvent.DialogueTextIDs.QuestGreeting)) {
		t.Fatalf("正式晉見羅馬利亞王未開始 quest_greeting：dlg=%v stage=%d",
			g.dlg.open, g.temporaryRoleStage)
	}
	traceCloseDialogue(t, g)
	if !g.storyFlag(roleEvent.PendingFlagRaw) || g.hasItem(roleEvent.RequiredItemRawID) {
		t.Fatalf("金皇冠任務交付後狀態錯：flag2c=%v crown=%v",
			g.storyFlag(roleEvent.PendingFlagRaw), g.hasItem(roleEvent.RequiredItemRawID))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存羅馬利亞 checkpoint: %v", err)
	}
	if err := restored.Load(); err != nil {
		t.Fatalf("讀取羅馬利亞 checkpoint: %v", err)
	}
	if restored.hasItem(itemuse.ItemMagicBall) || restored.storyFlag(magicBallIntactFlag) ||
		!restored.inTown || restored.curCty != roleEvent.OfferNPC.CTYRaw ||
		sceneSection(restored.cur) != roleEvent.OfferNPC.Section ||
		!restored.storyFlag(roleEvent.PendingFlagRaw) {
		t.Fatalf("羅馬利亞 checkpoint round-trip 錯：ball=%v flag51=%v town=%v cty=%d sec=%d",
			restored.hasItem(itemuse.ItemMagicBall), restored.storyFlag(magicBallIntactFlag),
			restored.inTown, restored.curCty, sceneSection(restored.cur))
	}
	// save/load 已驗證；釋放重建出的第二套大型圖形快取，後半沿原本玩家狀態繼續。
	restored = nil
	runtime.GC()

	// 從合法王座 checkpoint 正常回到羅馬利亞 sec0。死亡者先走教會、存活者再住旅店；
	// 之後以現有金錢在道具店購買一瓶聖水與可負擔藥草，正式北行 CTY10。
	traceTownSectionTo(t, g, roleEvent.OfferNPC.CTYRaw, 0)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceTrainNearTown(t, g, roleEvent.OfferNPC.CTYRaw, 10)
	traceAdventureWalkToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceTalkFacility(t, g, facWeapon)
	if !g.shop.active {
		t.Fatal("羅馬利亞武防店未由正式 facility NPC 開啟")
	}
	const bronzeShield = 0x3a
	shieldIdx := -1
	for i, code := range g.shop.codes {
		if code == bronzeShield {
			shieldIdx = i
			break
		}
	}
	if shieldIdx < 0 || g.heroGold < g.shop.items.Price(bronzeShield) {
		t.Fatalf("羅馬利亞武防店買不起青銅盾：idx=%d gold=%d codes=%v",
			shieldIdx, g.heroGold, g.shop.codes)
	}
	for g.shop.cursor != shieldIdx {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true})
	const tortoiseArmor = 0x22
	armorIdx := -1
	for i, code := range g.shop.codes {
		if code == tortoiseArmor {
			armorIdx = i
			break
		}
	}
	if armorIdx < 0 || g.heroGold < g.shop.items.Price(tortoiseArmor) {
		t.Fatalf("羅馬利亞武防店買不起龜殼甲胄：idx=%d gold=%d codes=%v",
			armorIdx, g.heroGold, g.shop.codes)
	}
	for g.shop.cursor != armorIdx {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true})
	press(InputState{Cancel: true})
	equipItem(1, bronzeShield) // 戰士正式換上青銅盾
	equipItem(0, tortoiseArmor)
	if g.companions[0].Shield != bronzeShield {
		t.Fatalf("戰士未經正式裝備面板換上青銅盾：%#x", g.companions[0].Shield)
	}
	if g.equip[1] != tortoiseArmor {
		t.Fatalf("勇者未經正式裝備面板換上龜殼甲胄：%#x", g.equip[1])
	}
	traceTalkFacility(t, g, facItem)
	if !g.shop.active {
		t.Fatal("羅馬利亞道具店未由正式 facility NPC 開啟")
	}
	buyFromOpenShop := func(code, limit int) int {
		t.Helper()
		idx := -1
		for i, candidate := range g.shop.codes {
			if candidate == code {
				idx = i
				break
			}
		}
		if idx < 0 {
			t.Fatalf("羅馬利亞道具店缺 item %#x：codes=%v", code, g.shop.codes)
		}
		for g.shop.cursor != idx {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		bought := 0
		for bought < limit && g.heroGold >= g.shop.items.Price(code) {
			press(InputState{Confirm: true})
			bought++
		}
		return bought
	}
	if bought := buyFromOpenShop(itemuse.ItemHolyWater, 1); bought != 1 {
		t.Fatalf("羅馬利亞北行前買不到聖水：bought=%d gold=%d", bought, g.heroGold)
	}
	buyFromOpenShop(herbCode, 5)
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)

	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceAdventureWalkToCty(t, g, 10)
	if !g.inTown || g.curCty != 10 {
		t.Fatalf("Lv4 正式北行未抵達 CTY10：town=%v cty=%d heroHP=%d gold=%d",
			g.inTown, g.curCty, g.heroHP, g.heroGold)
	}
	// 逐層走完香巴尼塔，最後由 CTY10 sec5 原始可走 trigger (6,8) 自動進入
	// handler14；不可直接呼叫 startFormation，也不可用調查寶箱代替。
	kandarEvent := g.pack.BossSurrenderEvents()[0]
	traceTownSectionTo(t, g, kandarEvent.Trigger.CTYRaw, kandarEvent.Trigger.Section)
	traceWalkTo(t, g, 6, 9)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 1, DirEdge: 1})
	if g.px != 6 || g.py != 8 || g.bossSurrenderStage != bossSurrenderIntro || !g.dlg.open {
		t.Fatalf("香巴尼塔原始 trigger 未自然啟動：pos=(%d,%d) stage=%d dlg=%v",
			g.px, g.py, g.bossSurrenderStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if !g.battle.active || g.bossSurrenderStage != bossSurrenderBattle || len(g.battle.enemies) != 4 {
		t.Fatalf("香巴尼塔 rec84 後未開原版四敵混合編隊：active=%v stage=%d enemies=%v",
			g.battle.active, g.bossSurrenderStage, g.battle.enemies)
	}
	traceResolveBattle(t, g, false)
	if g.bossSurrenderStage != bossSurrenderApology || !g.dlg.open {
		t.Fatalf("正式擊敗甘達特後未進 rec85：stage=%d dlg=%v hero=%d",
			g.bossSurrenderStage, g.dlg.open, g.heroHP)
	}
	traceCloseDialogue(t, g)
	if g.bossSurrenderStage != bossSurrenderChoice {
		t.Fatalf("rec85 後未進求饒選擇：stage=%d", g.bossSurrenderStage)
	}
	send(InputState{DirHeld: -1, DirEdge: 0}) // 選「否」
	press(InputState{Confirm: true})
	if g.bossSurrenderStage != bossSurrenderReject || !g.dlg.open {
		t.Fatalf("正式拒絕求饒未進 rec87：stage=%d dlg=%v", g.bossSurrenderStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if g.bossSurrenderStage != bossSurrenderChoice {
		t.Fatalf("rec87 後未回求饒選擇：stage=%d", g.bossSurrenderStage)
	}
	press(InputState{Confirm: true}) // 預設「是」
	if g.bossSurrenderStage != bossSurrenderFarewell || !g.dlg.open {
		t.Fatalf("正式接受求饒未進 rec86：stage=%d dlg=%v", g.bossSurrenderStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if g.storyFlag(kandarEvent.ClearFlagRaw) || g.bossSurrenderStage != bossSurrenderIdle {
		t.Fatalf("rec86 後未清除甘達特：flag=%v stage=%d",
			g.storyFlag(kandarEvent.ClearFlagRaw), g.bossSurrenderStage)
	}

	// 先走到寶箱下方第二格，再正常向上走到 (4,5)，使角色自然面向 (4,4) 寶箱。
	traceWalkTo(t, g, 4, 6)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 1, DirEdge: 1})
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if g.px != 4 || g.py != 5 || g.facing != 1 {
		t.Fatalf("未由正式行走抵達金皇冠寶箱下方：pos=(%d,%d) facing=%d",
			g.px, g.py, g.facing)
	}
	press(InputState{Confirm: true}) // 開命令
	if !g.cmd.open {
		t.Fatalf("金皇冠寶箱前未正常開啟命令窗：pos=(%d,%d) facing=%d cd=%d dlg=%v",
			g.px, g.py, g.facing, g.cd, g.dlg.open)
	}
	// 命令窗是 2 欄×3 列；先移到右欄，再往下到「調查」。
	for _, dir := range []int{3, 0, 0} {
		if g.cmd.cursor == int(cmdExamine) {
			break
		}
		send(InputState{DirHeld: -1, DirEdge: dir})
	}
	if g.cmd.cursor != int(cmdExamine) {
		t.Fatalf("命令窗無法以正式方向輸入選到調查：cursor=%d", g.cmd.cursor)
	}
	press(InputState{Confirm: true})
	if !g.hasItem(roleEvent.RequiredItemRawID) {
		t.Fatalf("正式調查甘達特事件寶箱未取得金皇冠：inventory=%v", g.inventory)
	}

	// 金皇冠與甘達特清除旗標必須跨存讀檔；讀檔後再以正式轉場離塔、返回羅馬利亞，
	// 證明事件不是只能在當前記憶體中成立的孤立 handler。
	if err := g.Save(); err != nil {
		t.Fatalf("保存金皇冠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建金皇冠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入冒險之書
	if !restored.hasItem(roleEvent.RequiredItemRawID) ||
		restored.storyFlag(kandarEvent.ClearFlagRaw) ||
		restored.showTitle ||
		!restored.inTown || restored.curCty != kandarEvent.Trigger.CTYRaw ||
		sceneSection(restored.cur) != kandarEvent.Trigger.Section {
		t.Fatalf("金皇冠 checkpoint round-trip 錯：crown=%v flag=%v title=%v town=%v cty=%d sec=%d",
			restored.hasItem(roleEvent.RequiredItemRawID),
			restored.storyFlag(kandarEvent.ClearFlagRaw), restored.showTitle, restored.inTown,
			restored.curCty, sceneSection(restored.cur))
	}
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	if !g.inTown || g.curCty != roleEvent.OfferNPC.CTYRaw ||
		!g.hasItem(roleEvent.RequiredItemRawID) {
		t.Fatalf("金皇冠讀檔後未能正式返回羅馬利亞：town=%v cty=%d crown=%v",
			g.inTown, g.curCty, g.hasItem(roleEvent.RequiredItemRawID))
	}
	// CTY02 夜表以兩名靜止守衛封住王宮中央通道；原版旅店成功交易會把
	// [0x526c] 寫回白天並重設時刻。正式路線若夜間抵達，先住宿再晉見。
	if g.isNight() {
		traceTalkFacility(t, g, facInn)
	}
	traceTownSectionTo(t, g, roleEvent.OfferNPC.CTYRaw, roleEvent.OfferNPC.Section,
		roleEvent.OfferNPC.Tile.X, roleEvent.OfferNPC.Tile.Y)
	traceTalkNPC(t, g, roleEvent.OfferNPC.Tile.X, roleEvent.OfferNPC.Tile.Y)
	if !g.dlg.open || g.temporaryRoleStage != temporaryRoleReturnPraise {
		t.Fatalf("持皇冠正式晉見未開始 return_praise：dlg=%v stage=%d",
			g.dlg.open, g.temporaryRoleStage)
	}
	if g.hasItem(roleEvent.RequiredItemRawID) || g.storyFlag(roleEvent.PendingFlagRaw) {
		t.Fatal("return_praise 前未完成原版還冠 transaction")
	}
	traceCloseDialogue(t, g)
	if g.temporaryRoleStage != temporaryRoleForcedChoice {
		t.Fatalf("還冠對白後未進 forced choice：stage=%d", g.temporaryRoleStage)
	}
	send(InputState{DirHeld: -1, DirEdge: 0}) // 選「否」
	press(InputState{Confirm: true})
	if g.temporaryRoleStage != temporaryRoleForcedReject || !g.dlg.open {
		t.Fatal("第一次選否未進 forced_reject")
	}
	traceCloseDialogue(t, g)
	if g.temporaryRoleStage != temporaryRoleForcedChoice {
		t.Fatal("第一次還冠選否後未被迫回到王位提議")
	}
	press(InputState{Confirm: true}) // 選「是」
	if g.temporaryRoleStage != temporaryRoleAccept || !g.dlg.open {
		t.Fatal("接受王位未播放 accept")
	}
	traceCloseDialogue(t, g)
	if g.storyFlag(roleEvent.NormalRoleFlagRaw) ||
		!g.storyFlag(roleEvent.ActiveRoleFlagRaw) || g.heroRole == nil {
		t.Fatalf("接受王位後狀態錯：normal=%v active=%v sprite=%v",
			g.storyFlag(roleEvent.NormalRoleFlagRaw),
			g.storyFlag(roleEvent.ActiveRoleFlagRaw), g.heroRole != nil)
	}

	// 臨時身分只保存 canonical story flag；從標題正式載入後必須自行重建角色圖。
	if err := g.Save(); err != nil {
		t.Fatalf("保存臨時王位 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建臨時王位讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入冒險之書
	if g.showTitle || !g.storyFlag(roleEvent.ActiveRoleFlagRaw) || g.heroRole == nil ||
		!g.inTown || g.curCty != roleEvent.OfferNPC.CTYRaw {
		t.Fatalf("臨時王位 checkpoint round-trip 錯：title=%v active=%v sprite=%v town=%v cty=%d",
			g.showTitle, g.storyFlag(roleEvent.ActiveRoleFlagRaw), g.heroRole != nil,
			g.inTown, g.curCty)
	}

	// 依原版影片由正式轉場走到地下競技場左上前國王。第一層選否表示不想
	// 繼續當王，第二層選是確認恢復冒險者；restore_accept 關閉後才交換旗標。
	traceTownSectionTo(t, g, roleEvent.RestoreNPC.CTYRaw, roleEvent.RestoreNPC.Section,
		roleEvent.RestoreNPC.Tile.X, roleEvent.RestoreNPC.Tile.Y)
	traceTalkNPC(t, g, roleEvent.RestoreNPC.Tile.X, roleEvent.RestoreNPC.Tile.Y)
	if !g.dlg.open || g.temporaryRoleStage != temporaryRoleRestoreIntro {
		t.Fatalf("前國王未開始 restore_intro：dlg=%v stage=%d",
			g.dlg.open, g.temporaryRoleStage)
	}
	traceCloseDialogue(t, g)
	send(InputState{DirHeld: -1, DirEdge: 0}) // 第一層「否」：不繼續當王
	press(InputState{Confirm: true})
	if g.temporaryRoleStage != temporaryRoleRestoreReconsider || !g.dlg.open {
		t.Fatal("恢復第一層選否未進 restore_reconsider")
	}
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true}) // 第二層「是」：確定退出王位
	if g.temporaryRoleStage != temporaryRoleRestoreAccept || !g.dlg.open {
		t.Fatal("恢復第二層選是未進 restore_accept")
	}
	traceCloseDialogue(t, g)
	if g.storyFlag(roleEvent.ActiveRoleFlagRaw) ||
		!g.storyFlag(roleEvent.NormalRoleFlagRaw) || g.heroRole != nil {
		t.Fatalf("恢復冒險者後狀態錯：active=%v normal=%v sprite=%v",
			g.storyFlag(roleEvent.ActiveRoleFlagRaw),
			g.storyFlag(roleEvent.NormalRoleFlagRaw), g.heroRole != nil)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存恢復冒險者 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := restored.Load(); err != nil {
		t.Fatalf("讀取恢復冒險者 checkpoint：%v", err)
	}
	if restored.storyFlag(roleEvent.ActiveRoleFlagRaw) ||
		!restored.storyFlag(roleEvent.NormalRoleFlagRaw) || restored.heroRole != nil {
		t.Fatal("恢復冒險者狀態未通過 save/load round-trip")
	}
	traceTownSectionTo(t, g, roleEvent.OfferNPC.CTYRaw, 0)
	traceExitTownBoundary(t, g)
	if g.inTown {
		t.Fatal("恢復冒險者後未能由正式路線離開羅馬利亞")
	}
	// 攻略下一節點：由羅馬利亞合法 checkpoint 北行，完整閉合
	// 沉睡村 → 夢幻紅寶石 → 精靈女王換物 → 覺醒粉甦醒。
	traceAdventureWalkToCty(t, g, noanielEvent.Use.CTYRaw)
	if !g.inTown || g.curCty != noanielEvent.Use.CTYRaw || sceneSection(g.cur) != 0 {
		t.Fatalf("辭位後未能正式抵達諾阿尼魯：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	traceExitTownBoundary(t, g)
	exchange := noanielEvent.Exchange
	traceAdventureWalkToCty(t, g, exchange.NPC.CTYRaw)
	traceTownSectionTo(t, g, exchange.NPC.CTYRaw, exchange.NPC.Section,
		exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	traceTalkNPC(t, g, exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	beforeText, _ := g.pack.TextGlyphCodes(exchange.BeforeTextID)
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, beforeText) {
		t.Fatal("未持夢幻紅寶石時，精靈女王未由正式話す入口播放 pack before text")
	}
	traceCloseDialogue(t, g)
	traceTownSectionTo(t, g, -1, -1)
	treasure := noanielEvent.Treasure
	traceAdventureWalkToCty(t, g, treasure.CTYRaw)
	if !g.inTown || g.curCty != treasure.CTYRaw {
		t.Fatalf("精靈村後未能正式抵達地底湖洞窟：town=%v cty=%d",
			g.inTown, g.curCty)
	}
	traceTownSectionTo(t, g, treasure.CTYRaw, treasure.Section, 21, 20)
	traceWalkTo(t, g, 21, 20)
	press(InputState{Confirm: true})
	if !g.cmd.open {
		t.Fatal("夢幻紅寶石事件格未能開啟正式命令窗")
	}
	for _, dir := range []int{3, 0, 0} {
		if g.cmd.cursor == int(cmdExamine) {
			break
		}
		send(InputState{DirHeld: -1, DirEdge: dir})
	}
	if g.cmd.cursor != int(cmdExamine) {
		t.Fatalf("夢幻紅寶石事件格無法選到調查：cursor=%d", g.cmd.cursor)
	}
	press(InputState{Confirm: true})
	if !g.hasItem(treasure.ItemRawID) || g.storyFlag(treasure.PresentFlag) {
		t.Fatalf("正式調查 CTY11 sec3 (21,20) 未取得夢幻紅寶石：%v", g.inventory)
	}
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, exchange.NPC.CTYRaw)
	traceTownSectionTo(t, g, exchange.NPC.CTYRaw, exchange.NPC.Section,
		exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	traceTalkNPC(t, g, exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	exchangeText, _ := g.pack.TextGlyphCodes(exchange.SuccessTextID)
	if g.hasItem(exchange.RequiredItemRawID) || !g.hasItem(exchange.GrantedItemRawID) ||
		!g.dlg.open || !reflect.DeepEqual(g.dlg.buf, exchangeText) {
		t.Fatalf("精靈女王未正式以夢幻紅寶石換覺醒粉：ruby=%v powder=%v dlg=%v",
			g.hasItem(exchange.RequiredItemRawID), g.hasItem(exchange.GrantedItemRawID), g.dlg.open)
	}
	traceCloseDialogue(t, g)
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, noanielEvent.Use.CTYRaw)
	traceUseInventoryItem(t, g, noanielEvent.Use.ItemRawID)
	awakenText, _ := g.pack.TextGlyphCodes(noanielEvent.Use.SuccessTextID)
	if g.hasItem(noanielEvent.Use.ItemRawID) ||
		!g.storyFlag(noanielEvent.Use.SetFlagsRaw[0]) ||
		g.storyFlag(noanielEvent.Use.ClearFlagsRaw[0]) ||
		!g.dlg.open || !reflect.DeepEqual(g.dlg.buf, awakenText) {
		t.Fatal("在諾阿尼魯正式使用覺醒粉後，未完成原版旗標切換、消耗、文字與場景重載")
	}
	traceCloseDialogue(t, g)
	// 長途往沙漠前先在剛甦醒的村莊以正式旅店交易補滿隊伍，
	// 不以測試狀態注入回復 HP/MP。
	traceTalkFacility(t, g, facInn)
	if err := g.Save(); err != nil {
		t.Fatalf("保存諾亞尼爾甦醒 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入冒險之書
	if !g.inTown || g.curCty != noanielEvent.Use.CTYRaw ||
		sceneSection(g.cur) != 0 ||
		g.hasItem(noanielEvent.Use.ItemRawID) ||
		!g.storyFlag(noanielEvent.Use.SetFlagsRaw[0]) ||
		g.storyFlag(noanielEvent.Use.ClearFlagsRaw[0]) {
		t.Fatalf("諾亞尼爾甦醒狀態未通過 save/load round-trip：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	traceExitTownBoundary(t, g)
	if g.inTown {
		t.Fatal("甦醒並讀檔後未能由正式邊界離開諾亞尼爾")
	}
	// 下一段攻略路線：阿莎拉慕 → 沙漠祠堂 → 依席斯 → 金字塔。
	for _, cty := range []int{6, 33} {
		traceAdventureWalkToCty(t, g, cty)
		if !g.inTown || g.curCty != cty || sceneSection(g.cur) != 0 {
			t.Fatalf("諾亞尼爾後未能正式抵達 CTY%d：town=%v cty=%d sec=%d",
				cty, g.inTown, g.curCty, sceneSection(g.cur))
		}
		if cty == 6 {
			traceReviveDeadAtChurch(t, g)
			traceTalkFacility(t, g, facInn)
			traceExitTownBoundary(t, g)
			traceTrainNearTown(t, g, cty, 13)
			traceAdventureWalkToCty(t, g, cty)
			traceReviveDeadAtChurch(t, g)
			traceTalkFacility(t, g, facInn)
		}
		traceExitTownBoundary(t, g)
	}
	// CTY12／32／73 的原版 cty_loc 均為 `(36,105)`，但 file 0x43a6
	// 從 CTY0 向上掃描並在第一筆命中停止，因此地表正式入口是 CTY12。
	// CTY32／73 是城內王宮區，不是必須依序踩過的重疊地表入口。
	traceAdventureWalkToCty(t, g, 12)
	traceExitTownBoundary(t, g)
	traceAdventureWalkToCty(t, g, 13)
	if !g.inTown || g.curCty != 13 || sceneSection(g.cur) != 0 {
		t.Fatalf("依席斯後未能正式抵達金字塔：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	traceTownSectionTo(t, g, 13, 2)
	pressSwitch := func(x, y int) {
		t.Helper()
		// 先走到開關正上方再向下踏入，避免路徑搜尋沿 y=25 穿過其他開關。
		traceWalkToNoPortal(t, g, x, y-1)
		traceWalkToNoPortal(t, g, x, y)
		if !g.dlg.open || g.sequenceGateStage != sequenceGatePrompt {
			t.Fatalf("正式踩開關 (%d,%d) 未開 prompt：stage=%d dlg=%v",
				x, y, g.sequenceGateStage, g.dlg.open)
		}
		traceCloseDialogue(t, g)
		if !g.sequenceGateChoosing() {
			t.Fatalf("開關 (%d,%d) prompt 後未進 Yes/No", x, y)
		}
		press(InputState{Confirm: true})
		if !g.dlg.open || g.sequenceGateStage != sequenceGatePressed {
			t.Fatalf("開關 (%d,%d) 選是後未進 pressed：stage=%d dlg=%v",
				x, y, g.sequenceGateStage, g.dlg.open)
		}
		for i := 0; i < 64 && g.sequenceGateStage == sequenceGatePressed && g.dlg.open; i++ {
			press(InputState{Confirm: true})
		}
		if g.sequenceGateStage == sequenceGatePressed {
			t.Fatalf("開關 (%d,%d) pressed 64 次 Confirm 後仍未推進", x, y)
		}
	}
	pressSwitch(25, 25)
	if !g.sequenceGateArmed || !g.storyFlag(pyramidGate.ClearFlagRaw) {
		t.Fatal("金字塔外右開關未只武裝第一階段")
	}
	pressSwitch(2, 25)
	if g.storyFlag(pyramidGate.ClearFlagRaw) || g.cur.Blocked(13, 10) ||
		g.sequenceGateStage != sequenceGateSuccess || !g.dlg.open {
		t.Fatalf("金字塔外右→外左未開石門：flag=%v blocked=%v stage=%d dlg=%v",
			g.storyFlag(pyramidGate.ClearFlagRaw), g.cur.Blocked(13, 10),
			g.sequenceGateStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)

	// 經已開石門走到魔法鑰匙下方，使用正式命令窗「調查」面向寶箱。
	traceWalkTo(t, g, 13, 6)
	if g.facing != 1 {
		for g.cd > 0 {
			send(InputState{DirHeld: -1, DirEdge: -1})
		}
		send(InputState{DirHeld: 1, DirEdge: -1}) // 對阻擋寶箱轉向，不會移動
	}
	press(InputState{Confirm: true})
	for _, dir := range []int{3, 0, 0} {
		if g.cmd.cursor == int(cmdExamine) {
			break
		}
		send(InputState{DirHeld: -1, DirEdge: dir})
	}
	if g.cmd.cursor != int(cmdExamine) {
		t.Fatalf("魔法鑰匙前無法以正式方向輸入選到調查：cursor=%d", g.cmd.cursor)
	}
	press(InputState{Confirm: true})
	magicKey := pyramidGate.UnlockedTreasure
	if !g.hasItem(magicKey.ItemRawID) || g.storyFlag(magicKey.PresentFlag) {
		t.Fatalf("正式調查未取得魔法鑰匙：item=%v flag=%v inventory=%v",
			g.hasItem(magicKey.ItemRawID), g.storyFlag(magicKey.PresentFlag), g.inventory)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存魔法鑰匙 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建魔法鑰匙讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入魔法鑰匙 checkpoint
	if !g.hasItem(magicKey.ItemRawID) || g.storyFlag(magicKey.PresentFlag) ||
		g.storyFlag(pyramidGate.ClearFlagRaw) || g.curCty != pyramidGate.CTYRaw ||
		sceneSection(g.cur) != pyramidGate.Section || g.cur.Blocked(13, 10) {
		t.Fatalf("魔法鑰匙 checkpoint round-trip 錯：key=%v present=%v gate=%v cty=%d sec=%d blocked=%v",
			g.hasItem(magicKey.ItemRawID), g.storyFlag(magicKey.PresentFlag),
			g.storyFlag(pyramidGate.ClearFlagRaw), g.curCty, sceneSection(g.cur),
			g.cur.Blocked(13, 10))
	}
	// 魔法鑰匙後依正式 transition table 離開金字塔，回依席斯王宮。
	// CTY32 是與地表 CTY12 共用入口座標的宮殿區，只能由城內轉場抵達。
	traceTownSectionTo(t, g, pyramidGate.CTYRaw, 0, 1, 41, 1)
	traceExitTownBoundary(t, g, true)
	traceAdventureWalkToCty(t, g, 12)
	traceTownSectionTo(t, g, 32, 4)
	traceTownSectionTo(t, g, 73, 0)

	// 依席斯王宮 → 羅馬利亞祠堂 → 波魯多加，全部使用正式地表與
	// transition table；祠堂紅門由背包中的魔法鑰匙開啟。
	traceTownSectionTo(t, g, 12, 1, 10, 34)
	traceTalkFacility(t, g, facInn)
	traceTownSectionTo(t, g, 12, 0)
	traceExitTownBoundary(t, g)
	traceAdventureWalkToCty(t, g, 34, false)
	traceTownSectionTo(t, g, 35, 0)
	traceExitTownBoundary(t, g, true)
	traceAdventureWalkToCty(t, g, 16, false)
	traceTalkFacility(t, g, facInn)
	if g.dnPhase != 0 || g.dnStep != 0 {
		t.Fatalf("波魯多加住宿後未回到白天：phase=%d step=%d", g.dnPhase, g.dnStep)
	}

	vehicleEvents := g.pack.StagedVehicleExchangeEvents()
	if len(vehicleEvents) != 1 {
		t.Fatalf("staged vehicle exchange events=%d, want 1", len(vehicleEvents))
	}
	portoga := vehicleEvents[0]
	traceTownSectionTo(t, g, portoga.NPC.CTYRaw, portoga.NPC.Section)
	traceTalkNPC(t, g, portoga.NPC.Tile.X, portoga.NPC.Tile.Y)
	if !g.dlg.open || g.vehicleExchangeStage != vehicleExchangeQuestIntro ||
		g.hasItem(portoga.GrantedItemRawID) {
		t.Fatalf("波魯多加首訪未停在授信介紹：dlg=%v stage=%d item=%v",
			g.dlg.open, g.vehicleExchangeStage, g.hasItem(portoga.GrantedItemRawID))
	}
	traceCloseDialogue(t, g)
	if !g.hasItem(portoga.GrantedItemRawID) ||
		g.storyFlag(portoga.QuestPresentFlagRaw) || g.shipOwned {
		t.Fatalf("波魯多加首訪交易錯誤：letter=%v questFlag=%v ship=%v",
			g.hasItem(portoga.GrantedItemRawID),
			g.storyFlag(portoga.QuestPresentFlagRaw), g.shipOwned)
	}

	if err := g.Save(); err != nil {
		t.Fatalf("保存國王信件 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建國王信件讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.hasItem(portoga.GrantedItemRawID) ||
		g.storyFlag(portoga.QuestPresentFlagRaw) || g.shipOwned ||
		g.curCty != portoga.NPC.CTYRaw || sceneSection(g.cur) != portoga.NPC.Section {
		t.Fatalf("國王信件 save/load round-trip 錯誤：item=%v flag=%v ship=%v cty=%d sec=%d",
			g.hasItem(portoga.GrantedItemRawID), g.storyFlag(portoga.QuestPresentFlagRaw),
			g.shipOwned, g.curCty, sceneSection(g.cur))
	}
}

// traceTrainNearTown 在指定城鎮入口附近的同一個低危 region 來回走動，以正式遭遇輸入
// 取得 EXP；隊伍有人降到 1/2 HP 就正常進城找旅店，再由 section0 邊界出城。
func traceTrainNearTown(t *testing.T, g *Game, cty, wantLevel int) {
	t.Helper()
	if g.inTown || cty < 0 || cty >= len(ctyLoc) || ctyLoc[cty][2] != g.layer {
		t.Fatalf("練級起點錯：town=%v cty=%d layer=%d", g.inTown, cty, g.layer)
	}
	startRegion := g.encounters.Region(g.px, g.py)
	prevX, prevY := -1, -1

	needsRest := func() bool {
		_, heroMax, _, _, _ := g.heroStats()
		if g.heroHP <= 0 || g.heroHP*2 <= heroMax {
			return true
		}
		for _, m := range g.companions {
			if m.CurHP <= 0 || m.CurHP*2 <= m.MaxHP() {
				return true
			}
		}
		return false
	}
	allAlive := func() bool {
		if g.heroHP <= 0 {
			return false
		}
		for _, m := range g.companions {
			if !m.Alive() {
				return false
			}
		}
		return true
	}
	rest := func() {
		t.Helper()
		traceAdventureWalkToCty(t, g, cty)
		if !allAlive() {
			traceReviveDeadAtChurch(t, g)
		}
		before := g.heroGold
		traceTalkFacility(t, g, facInn)
		wantCost := facilityForCty(cty, 0).innCost * (1 + len(g.companions))
		if g.heroGold != before-wantCost {
			t.Fatalf("練級住宿扣款錯：before=%d after=%d want=%d",
				before, g.heroGold, wantCost)
		}
		_, heroMax, _, _, _ := g.heroStats()
		if g.heroHP != heroMax || g.heroMP != g.heroMaxMP() {
			t.Fatalf("練級住宿後勇者未補滿：HP=%d/%d MP=%d/%d",
				g.heroHP, heroMax, g.heroMP, g.heroMaxMP())
		}
		for i, m := range g.companions {
			if m.CurHP != m.MaxHP() || m.CurMP != m.MaxMP() {
				t.Fatalf("練級住宿後同伴%d未補滿：HP=%d/%d MP=%d/%d",
					i, m.CurHP, m.MaxHP(), m.CurMP, m.MaxMP())
			}
		}
		traceExitTownBoundary(t, g)
		prevX, prevY = -1, -1
	}

	for step := 0; step < 50000; step++ {
		level, _, _, _, _ := g.heroStats()
		if level >= wantLevel {
			if needsRest() {
				rest()
			}
			t.Logf("正式低危區練級完成：hero Lv%d exp=%d gold=%d", level, g.heroExp, g.heroGold)
			return
		}
		if g.battle.active {
			// 原版低階隊伍遇到超出目前區域戰力的怪物應使用正式「逃跑」指令；
			// 阿里阿罕只與弱敵交戰；Lv4 後的羅馬利亞訓練才正式迎戰當地怪物。
			traceResolveBattle(t, g, cty == 0)
			continue
		}
		if needsRest() {
			rest()
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待練級移動 cooldown: %v", err)
			}
			continue
		}

		moved := false
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny := g.px+dx, g.py+dy
			if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
				g.cur.Blocked(nx, ny) || findCtyAtLayer(nx, ny, g.layer) >= 0 ||
				g.encounters.Region(nx, ny) != startRegion {
				continue
			}
			if nx == prevX && ny == prevY && dir < 3 {
				continue
			}
			prevX, prevY = g.px, g.py
			if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
				t.Fatalf("低危區練級移動 dir%d: %v", dir, err)
			}
			moved = true
			break
		}
		if !moved {
			t.Fatalf("CTY%d 入口附近找不到同 region 可走練級格 @(%d,%d)",
				cty, g.px, g.py)
		}
	}
	t.Fatalf("正式低危區練級 50000 step 仍未達 Lv%d：exp=%d", wantLevel, g.heroExp)
}

// traceTalkFacility 依當前原始 NPC 表找 facility type，而不是把店員座標寫成 debug shortcut；
// 找到後仍由 traceTalkNPC 逐步走到櫃台並送正式交談輸入。
func traceTalkFacility(t *testing.T, g *Game, typ int) {
	t.Helper()
	for i := range g.cur.npcs {
		n := &g.cur.npcs[i]
		f := facilityForCty(g.curCty, n.b4)
		if f == nil || f.typ != typ || (n.ctrl>>3)&7 < 3 {
			continue
		}
		if traceNPCReachable(g, g.curCty, sceneSection(g.cur), g.px, g.py, n.x, n.y) {
			traceTalkNPC(t, g, n.x, n.y)
			return
		}
	}
	t.Fatalf("CTY%d sec%d 找不到可達 facility type%d NPC",
		g.curCty, sceneSection(g.cur), typ)
}

// traceReviveDeadAtChurch 只在隊伍有死者時，經目前城鎮的正式 facility NPC 與完整
// service→target→yes/no input state machine 復活。它不注入 HP 或金錢。
func traceReviveDeadAtChurch(t *testing.T, g *Game) {
	t.Helper()
	hasDead := false
	for target := 0; target < g.churchPartyLen(); target++ {
		if !g.churchMemberAlive(target) {
			hasDead = true
			break
		}
	}
	if !hasDead {
		return
	}
	beforeChurch := g.heroGold
	wantChurchCost := 0
	traceTalkFacility(t, g, facChurch)
	if !g.church.active || g.church.stage != churchService {
		t.Fatalf("陣亡後未進入正式教會 modal：active=%v stage=%d",
			g.church.active, g.church.stage)
	}
	for target := 0; target < g.churchPartyLen(); target++ {
		if g.churchMemberAlive(target) {
			continue
		}
		wantChurchCost += g.pack.ReviveCost(g.churchMemberLevel(target))
		for g.church.serviceCursor != 2 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: 0}); err != nil {
				t.Fatalf("教會選復活服務：%v", err)
			}
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會確認復活服務：%v", err)
		}
		for g.church.targetCursor != target {
			if err := g.step(InputState{DirHeld: -1, DirEdge: 0}); err != nil {
				t.Fatalf("教會選陣亡角色%d：%v", target, err)
			}
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會確認角色%d：%v", target, err)
		}
		if g.church.stage != churchConfirm {
			t.Fatalf("教會角色%d 未進 yes/no：stage=%d msg=%q",
				target, g.church.stage, g.church.msg)
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會支付角色%d復活：%v", target, err)
		}
		if !g.churchMemberAlive(target) {
			t.Fatalf("教會交易後角色%d仍陣亡", target)
		}
	}
	if err := g.step(InputState{Cancel: true, DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatalf("離開教會：%v", err)
	}
	if g.heroGold != beforeChurch-wantChurchCost {
		t.Fatalf("教會復活扣款錯：before=%d after=%d wantCost=%d",
			beforeChurch, g.heroGold, wantChurchCost)
	}
}

func traceExitTownBoundary(t *testing.T, g *Game, avoidPortals ...bool) {
	t.Helper()
	type candidate struct {
		x, y, dir int
		path      []int
	}
	var best *candidate
	for y := 0; y < g.cur.h; y++ {
		for x := 0; x < g.cur.w; x++ {
			dir := -1
			switch {
			case y == g.cur.h-1:
				dir = 0
			case y == 0:
				dir = 1
			case x == 0:
				dir = 2
			case x == g.cur.w-1:
				dir = 3
			}
			if dir < 0 || g.cur.Blocked(x, y) {
				continue
			}
			if _, _, _, _, ok := g.cur.tileTransition(x, y); ok {
				continue
			}
			path := tracePath(g.cur, g.px, g.py, x, y)
			if len(avoidPortals) > 0 && avoidPortals[0] {
				path = tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier())
			}
			if (x != g.px || y != g.py) && len(path) == 0 {
				continue
			}
			if best == nil || len(path) < len(best.path) {
				best = &candidate{x: x, y: y, dir: dir, path: path}
			}
		}
	}
	if best == nil {
		t.Fatalf("CTY%d sec%d 找不到可達的 section0 邊界出口", g.curCty, sceneSection(g.cur))
	}
	if len(avoidPortals) > 0 && avoidPortals[0] {
		traceWalkToNoPortal(t, g, best.x, best.y)
	} else {
		traceWalkTo(t, g, best.x, best.y)
	}
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待出城 cooldown: %v", err)
		}
	}
	if err := g.step(InputState{DirHeld: best.dir, DirEdge: -1}); err != nil {
		t.Fatalf("走出城鎮邊界: %v", err)
	}
	if g.inTown {
		t.Fatalf("從 CTY%d sec%d 邊界 (%d,%d) dir%d 未正常出城：player=(%d,%d) cd=%d cmd=%v panel=%d dlg=%v title=%v",
			g.curCty, sceneSection(g.cur), best.x, best.y, best.dir,
			g.px, g.py, g.cd, g.cmd.open, g.panel, g.dlg.open, g.showTitle)
	}
}

// traceUseInventoryItem 只送 production 命令窗／道具面板輸入，從背包中選到指定 id 後使用。
func traceUseInventoryItem(t *testing.T, g *Game, code int) {
	t.Helper()
	idx := -1
	for i, got := range g.inventory {
		if got == code {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("背包沒有要使用的道具 0x%02x：%v", code, g.inventory)
	}
	step := func(in InputState) {
		t.Helper()
		if in.DirHeld == 0 {
			in.DirHeld = -1
		}
		if in.DirEdge == 0 && in.Confirm {
			in.DirEdge = -1
		}
		if err := g.step(in); err != nil {
			t.Fatalf("道具 production input: %v", err)
		}
	}
	step(InputState{Confirm: true}) // 開命令窗，cursor=對話
	step(InputState{DirEdge: 3})    // 對話→咒文
	step(InputState{DirEdge: 0})    // 咒文→道具
	step(InputState{Confirm: true}) // 開道具面板
	for i := 0; i < idx; i++ {
		step(InputState{DirEdge: 0})
	}
	if g.panel != panelItem || g.panelCursor != idx {
		t.Fatalf("正式道具選單未移到 index%d：panel=%d cursor=%d", idx, g.panel, g.panelCursor)
	}
	step(InputState{Confirm: true})
	if g.panel != panelNone { // 聖水等一般消耗品使用後仍留在清單，正式按 B 關閉
		step(InputState{Cancel: true})
	}
}

// traceOpenReachableDoor 找目前連通區邊界上可由持有鑰匙開啟的門，並經正式「調查」命令開門。
func traceOpenReachableDoor(t *testing.T, g *Game, wantDoor ...int) {
	t.Helper()
	opened := 0
	for {
		found := false
	findDoor:
		for y := 0; y < g.cur.h; y++ {
			for x := 0; x < g.cur.w; x++ {
				if len(wantDoor) >= 2 && (x != wantDoor[0] || y != wantDoor[1]) {
					continue
				}
				if need := g.cur.doorTier(x, y); need == 0 || g.keyTier() < need {
					continue
				}
				for dir := 0; dir < 4; dir++ {
					dx, dy := dirDelta(dir)
					sx, sy := x-dx, y-dy
					if sx < 0 || sy < 0 || sx >= g.cur.w || sy >= g.cur.h || g.cur.Blocked(sx, sy) {
						continue
					}
					if (sx != g.px || sy != g.py) && len(tracePath(g.cur, g.px, g.py, sx, sy)) == 0 {
						continue
					}
					traceWalkTo(t, g, sx, sy)
					for g.cd > 0 {
						if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
							t.Fatalf("等待開門 cooldown: %v", err)
						}
					}
					if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
						t.Fatalf("面向門: %v", err)
					}
					press := func(in InputState) {
						t.Helper()
						if in.DirHeld == 0 {
							in.DirHeld = -1
						}
						if in.DirEdge == 0 && in.Confirm {
							in.DirEdge = -1
						}
						if err := g.step(in); err != nil {
							t.Fatalf("開門命令輸入: %v", err)
						}
					}
					press(InputState{Confirm: true}) // 開命令
					press(InputState{DirEdge: 3})    // 對話→咒文
					press(InputState{DirEdge: 0})    // 咒文→道具
					press(InputState{DirEdge: 0})    // 道具→調查
					press(InputState{Confirm: true}) // 調查面向門
					if g.cur.doorTier(x, y) != 0 || !g.hasItem(0x55) {
						t.Fatalf("正式調查未開門或錯誤消耗鑰匙：doorTier=%d key=%v",
							g.cur.doorTier(x, y), g.hasItem(0x55))
					}
					opened++
					t.Logf("正式調查開門 CTY%d sec%d door(%d,%d)", g.curCty, sceneSection(g.cur), x, y)
					found = true
					break findDoor
				}
			}
		}
		if !found {
			break
		}
	}
	if opened == 0 {
		t.Fatal("目前 section 找不到可由玩家抵達並開啟的門")
	}
}

// traceAdventureWalkToCty 在同一層地表尋路到 CTY 入口；隨機遭遇由正式 battle input 解決。
// fightStrong=false 供已完成該區戰力訓練的路段使用，避免把「反覆逃跑直到成功」誤當成
// 可重現的玩家流程；省略時保留前段低等級隊伍遇強敵即逃的策略。
func traceAdventureWalkToCty(t *testing.T, g *Game, wantCty int, fleeStrong ...bool) {
	t.Helper()
	shouldFleeStrong := len(fleeStrong) == 0 || fleeStrong[0]
	if wantCty < 0 || wantCty >= len(ctyLoc) || ctyLoc[wantCty][2] != g.layer {
		t.Fatalf("無效 CTY%d / layer%d", wantCty, g.layer)
	}
	tx, ty := ctyLoc[wantCty][0]+1, ctyLoc[wantCty][1]
	// 原版 findCtyAt 接受 loc.X 與 loc.X+1 兩格；兩格不保證都在同一個可走連通區
	// （CTY9 誘惑洞窟入口只有左格能由雷貝方向抵達）。
	if (g.px != tx || g.py != ty) && len(traceWorldPath(g, tx, ty, wantCty)) == 0 {
		tx = ctyLoc[wantCty][0]
	}
	for i := 0; i < 12000 && (!g.inTown || g.curCty != wantCty); i++ {
		if g.battle.active {
			traceResolveBattle(t, g, shouldFleeStrong)
			continue
		}
		if g.inTown {
			t.Fatalf("前往 CTY%d 時誤入 CTY%d；world path 未避開其他入口", wantCty, g.curCty)
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待地表移動 cooldown: %v", err)
			}
			continue
		}
		path := traceWorldPath(g, tx, ty, wantCty)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("地表移動 dir%d: %v", path[0], err)
		}
	}
	if !g.inTown || g.curCty != wantCty {
		lx, ly := ctyLoc[wantCty][0], ctyLoc[wantCty][1]
		t.Fatalf("無法由地表抵達 CTY%d；目前 town=%v cty=%d @(%d,%d)，入口 (%d,%d) blocked=%v / (%d,%d) blocked=%v",
			wantCty, g.inTown, g.curCty, g.px, g.py,
			lx, ly, g.cur.Blocked(lx, ly), lx+1, ly, g.cur.Blocked(lx+1, ly))
	}
}

// traceWorldPath 與 tracePath 相同，但把目的以外的 CTY 入口格視為障礙；
// 否則最短路徑可能先踩中 CTY28 等重疊入口，自動切場景後便不再是同一條地表路徑。
func traceWorldPath(g *Game, tx, ty, wantCty int) []int {
	type node struct{ x, y int }
	start, goal := node{g.px, g.py}, node{tx, ty}
	if start == goal {
		return nil
	}
	q := []node{start}
	seen := map[node]bool{start: true}
	prev := map[node]node{}
	prevDir := map[node]int{}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			n := node{p.x + dx, p.y + dy}
			if seen[n] || n.x < 0 || n.y < 0 || n.x >= g.cur.w || n.y >= g.cur.h ||
				g.cur.Blocked(n.x, n.y) {
				continue
			}
			if cty := findCtyAtLayer(n.x, n.y, g.layer); cty >= 0 && cty != wantCty {
				continue
			}
			seen[n], prev[n], prevDir[n] = true, p, dir
			if n == goal {
				var rev []int
				for cur := goal; cur != start; cur = prev[cur] {
					rev = append(rev, prevDir[cur])
				}
				for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
					rev[i], rev[j] = rev[j], rev[i]
				}
				return rev
			}
			q = append(q, n)
		}
	}
	return nil
}

func traceResolveBattle(t *testing.T, g *Game, fleeStrong ...bool) {
	t.Helper()
	shouldFleeStrong := len(fleeStrong) == 0 || fleeStrong[0]
	healTarget := -1
	wantedSpell := -1
	blindAttempted := false
	defenseCasts := 0
	bestSpell := func(actor int, kind spell.Kind) int {
		best, bestBase := -1, -1
		for _, rec := range g.battle.actorSpells(actor) {
			def, ok := spell.GetDef(rec)
			if ok && def.Kind == kind && def.MP <= g.battle.actorMP(actor) && def.Base > bestBase {
				best, bestBase = rec, def.Base
			}
		}
		return best
	}
	hasAffordableSpell := func(actor, rec int) bool {
		for _, known := range g.battle.actorSpells(actor) {
			if known == rec {
				def, ok := spell.GetDef(rec)
				return ok && def.MP <= g.battle.actorMP(actor)
			}
		}
		return false
	}
	const maxInputs = 1536
	for i := 0; i < maxInputs && g.battle.active; i++ {
		if g.battle.result == 2 {
			heroLv, _, _, _, _ := g.heroStats()
			comp := make([][3]int, len(g.battle.companions))
			for i, c := range g.battle.companions {
				comp[i] = [3]int{c.level, c.hp, c.maxHP}
			}
			t.Fatalf("正常前段路線發生全滅：mon=%d enemies=%v heroLv=%d hero=%d/%d atk=%d def=%d companions(lv,hp,max)=%v",
				g.battle.monID, g.battle.enemies, heroLv, g.battle.heroHP,
				g.battle.heroMax, g.battle.heroAtk, g.battle.heroDef, comp)
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		switch g.battle.phase {
		case phCommand:
			menu := g.battle.commandMenu(g.battle.commandActor)
			healTarget, wantedSpell = -1, -1
			if g.battle.usedHerbs < g.battle.heroHerbs {
				bestHP, bestMax := 0, 1
				for _, actor := range g.battle.aliveActorIndices() {
					hp, maxHP := g.battle.actorHP(actor)
					if hp*4 <= maxHP*3 &&
						(healTarget < 0 || hp*bestMax < bestHP*maxHP) {
						healTarget, bestHP, bestMax = actor, hp, maxHP
					}
				}
			}
			needHeal := healTarget >= 0
			needFlee := g.battle.commandActor == g.battle.firstCommandActor() &&
				shouldFleeStrong && g.battle.monID >= 10
			wantCommand := bcWar
			switch {
			case needHeal && bestSpell(g.battle.commandActor, spell.Heal) >= 0:
				wantedSpell = bestSpell(g.battle.commandActor, spell.Heal)
				wantCommand = bcSpell
			case needHeal:
				wantCommand = bcItem
			case needFlee:
				wantCommand = bcFlee
			case g.bossSurrenderStage == bossSurrenderBattle &&
				!blindAttempted && hasAffordableSpell(g.battle.commandActor, 158):
				wantedSpell, wantCommand = 158, bcSpell
			case g.bossSurrenderStage == bossSurrenderBattle &&
				defenseCasts < 2 && hasAffordableSpell(g.battle.commandActor, 154):
				wantedSpell, wantCommand = 154, bcSpell
			case bestSpell(g.battle.commandActor, spell.Dmg) >= 0:
				wantedSpell = bestSpell(g.battle.commandActor, spell.Dmg)
				wantCommand = bcSpell
			}
			if menu[g.battle.cursor] != wantCommand {
				in.DirEdge = 0
			} else {
				in.Confirm = true
			}
		case phSpell:
			spells := g.battle.actorSpells(g.battle.commandActor)
			if wantedSpell < 0 {
				t.Fatalf("正式戰鬥進入咒文選單但沒有選定咒文：actor=%d", g.battle.commandActor)
			}
			if spells[g.battle.spellCursor] != wantedSpell {
				in.DirEdge = 0
			} else {
				if wantedSpell == 158 {
					blindAttempted = true
				}
				if wantedSpell == 154 {
					defenseCasts++
				}
				in.Confirm = true
			}
		case phTargetAlly:
			targets := g.battle.aliveActorIndices()
			if len(targets) == 0 {
				t.Fatal("治療目標選單沒有存活隊員")
			}
			wantPos := 0
			for i, actor := range targets {
				if actor == healTarget {
					wantPos = i
					break
				}
			}
			tapped := false
			for _, hit := range g.battle.hits {
				if hit.idx != wantPos {
					continue
				}
				in.Tapped, in.TapX, in.TapY = true, hit.x+hit.w/2, hit.y+hit.h/2
				tapped = true
				break
			}
			if !tapped {
				if g.battle.targetCursor == wantPos {
					in.Confirm = true
				} else {
					in.DirEdge = 0
				}
			}
		case phTargetEnemy:
			targets := g.battle.aliveEnemyIndices()
			want := 0
			for i := 1; i < len(targets); i++ {
				if g.battle.enemies[targets[i]].max < g.battle.enemies[targets[want]].max {
					want = i
				}
			}
			if g.battle.targetCursor != want {
				in.DirEdge = 0
			} else {
				in.Confirm = true
			}
		case phMessage, phEnd:
			in.Confirm = true
		default:
			t.Fatalf("自動戰鬥遇到未處理 phase=%d", g.battle.phase)
		}
		if err := g.step(in); err != nil {
			t.Fatalf("推進正式戰鬥輸入: %v", err)
		}
	}
	if g.battle.active {
		enemyHP := make([]int, len(g.battle.enemies))
		for i := range g.battle.enemies {
			enemyHP[i] = g.battle.enemies[i].hp
		}
		t.Fatalf("戰鬥 %d 次正式輸入後仍未結束：phase=%d actor=%d cursor=%d target=%d healTarget=%d pending=%+v herbs=%d/%d targets=%v hits=%+v mon=%d result=%d hero=%d enemyHP=%v",
			maxInputs,
			g.battle.phase, g.battle.commandActor, g.battle.cursor, g.battle.targetCursor,
			healTarget, g.battle.pending, g.battle.usedHerbs, g.battle.heroHerbs,
			g.battle.aliveActorIndices(), g.battle.hits,
			g.battle.monID, g.battle.result, g.battle.heroHP, enemyHP)
	}
}

// traceTownSectionTo 從目前 section 依真實 transition table 找路，再逐格走過每個 portal。
func traceTownSectionTo(t *testing.T, g *Game, wantCty, wantSec int, wantNPC ...int) {
	t.Helper()
	type node struct{ cty, sec, x, y int }
	type hop struct {
		from, to         node
		portalX, portalY int
	}
	start := node{g.curCty, sceneSection(g.cur), g.px, g.py}
	q := []node{start}
	seen := map[node]bool{start: true}
	prev := map[node]hop{}
	var goal node
	found := false
	for len(q) > 0 && !found {
		cur := q[0]
		q = q[1:]
		if cur.cty == wantCty && (wantCty < 0 || cur.sec == wantSec) {
			reachable := len(wantNPC) < 2
			if len(wantNPC) == 2 {
				reachable = traceNPCReachable(g, cur.cty, cur.sec, cur.x, cur.y,
					wantNPC[0], wantNPC[1])
			} else if len(wantNPC) >= 3 {
				sc := g.cur
				if cur != start {
					var err error
					sc, err = loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
						cur.cty, mapBlkNum[cur.cty], cur.sec, g.dnPhase, g.storyFlag)
					if err != nil {
						sc = nil
					}
				}
				reachable = sc != nil && (cur.x == wantNPC[0] && cur.y == wantNPC[1] ||
					len(tracePortalPath(sc, cur.x, cur.y, wantNPC[0], wantNPC[1], g.keyTier())) > 0)
			}
			if reachable {
				goal, found = cur, true
				break
			}
		}
		if cur.cty < 0 {
			continue
		}
		sc := g.cur
		if cur != start {
			var err error
			sc, err = loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
				cur.cty, mapBlkNum[cur.cty], cur.sec, g.dnPhase, g.storyFlag)
			if err != nil {
				continue
			}
		}
		for y := 0; y < sc.h; y++ {
			for x := 0; x < sc.w; x++ {
				dcty, dsec, dx, dy, ok := sc.tileTransition(x, y)
				if !ok {
					continue
				}
				pathLen := 0
				if x == cur.x && y == cur.y {
					pathLen = -1
				} else {
					pathLen = len(tracePortalPath(sc, cur.x, cur.y, x, y, g.keyTier()))
				}
				if dsec >= 0xfe && wantCty >= 0 {
					continue
				}
				if (x != cur.x || y != cur.y) &&
					pathLen == 0 {
					continue // transition table 有 edge，但此入口所在的連通區走不到該 portal
				}
				next := node{-1, -1, dx, dy}
				if dsec < 0xfe {
					next = node{cur.cty, dsec, dx, dy}
					if dcty >= 0 && dcty < 100 && dcty != cur.cty {
						next.cty = dcty
					}
				}
				if next == cur || seen[next] {
					continue
				}
				seen[next] = true
				prev[next] = hop{from: cur, to: next, portalX: x, portalY: y}
				q = append(q, next)
			}
		}
	}
	if !found {
		var portals [][7]int
		for y := 0; y < g.cur.h; y++ {
			for x := 0; x < g.cur.w; x++ {
				dcty, dsec, dx, dy, ok := g.cur.tileTransition(x, y)
				if !ok {
					continue
				}
				portals = append(portals, [7]int{x, y, dcty, dsec, dx, dy,
					len(tracePortalPath(g.cur, start.x, start.y, x, y, g.keyTier()))})
			}
		}
		npcs := make([][3]int, len(g.cur.npcs))
		for i, n := range g.cur.npcs {
			npcs[i] = [3]int{n.x, n.y, n.ctrl}
		}
		t.Fatalf("transition graph 無法由 CTY%d sec%d @(%d,%d) 到 CTY%d sec%d；portal[x y cty sec dx dy pathlen]=%v NPC[x y ctrl]=%v",
			start.cty, start.sec, start.x, start.y, wantCty, wantSec, portals, npcs)
	}
	var route []hop
	for cur := goal; cur != start; cur = prev[cur].from {
		route = append(route, prev[cur])
	}
	for i := len(route) - 1; i >= 0; i-- {
		h := route[i]
		traceWalkThroughPortal(t, g, h.portalX, h.portalY, h.to.cty, h.to.sec, h.to.x, h.to.y)
	}
}

func traceNPCReachable(g *Game, cty, sec, sx, sy, nx, ny int) bool {
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		cty, mapBlkNum[cty], sec, g.dnPhase, g.storyFlag)
	if err != nil {
		return false
	}
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		for dist := 1; dist <= 2; dist++ {
			x, y := nx-dx*dist, ny-dy*dist
			if x < 0 || y < 0 || x >= sc.w || y >= sc.h || sc.Blocked(x, y) {
				continue
			}
			if dist == 2 && !sc.attr.Blocked(sc.tileIdx(nx-dx, ny-dy)) {
				continue
			}
			if x == sx && y == sy || len(tracePath(sc, sx, sy, x, y)) > 0 {
				return true
			}
		}
	}
	return false
}

func traceCloseDialogue(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 64 && g.dlg.open; i++ {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
			t.Fatalf("推進對話: %v", err)
		}
	}
	if g.dlg.open {
		t.Fatal("對話 64 次 Confirm 後仍未關閉")
	}
}

func traceWalkThroughPortal(t *testing.T, g *Game, x, y, wantCty, wantSec int, wantPos ...int) {
	t.Helper()
	arrived := func() bool {
		if wantCty < 0 {
			return !g.inTown && g.curCty < 0
		}
		if g.curCty != wantCty || g.cur == nil || g.cur.sec != wantSec {
			return false
		}
		return len(wantPos) < 2 || g.px == wantPos[0] && g.py == wantPos[1]
	}
	for i := 0; i < 3000; i++ {
		if arrived() {
			return
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待 portal cooldown: %v", err)
			}
			continue
		}
		// 轉場目的座標可能正好也是回程 portal。原版只在「移動踏入」時消費
		// transition；站在該格不會自動連續折返，所以用正式方向輸入先走開，
		// 下一輪再依路徑踏回。
		if g.px == x && g.py == y {
			movedOff := false
			for dir := 0; dir < 4; dir++ {
				dx, dy := dirDelta(dir)
				nx, ny := g.px+dx, g.py+dy
				if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
					g.cur.Blocked(nx, ny) {
					continue
				}
				if _, _, _, _, ok := g.cur.tileTransition(nx, ny); ok {
					continue
				}
				if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
					t.Fatalf("離開回程 portal(%d,%d) dir%d: %v", x, y, dir, err)
				}
				movedOff = g.px == nx && g.py == ny
				if movedOff {
					break
				}
			}
			if !movedOff {
				t.Fatalf("站在回程 portal(%d,%d) 但找不到可走開的相鄰格", x, y)
			}
			continue
		}
		path := tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier())
		if len(path) == 0 {
			// 遊走 NPC 可能暫時封住唯一通道；與 traceWalkOne 一樣送空白
			// frame，讓 production NPC tick 後重新尋路。永久不連通仍會由
			// 3000 次上限報錯。
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待 portal 路徑 NPC: %v", err)
			}
			continue
		}
		dx, dy := dirDelta(path[0])
		if g.cur.doorTier(g.px+dx, g.py+dy) != 0 {
			traceOpenReachableDoor(t, g, g.px+dx, g.py+dy)
			continue
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("走向 portal dir%d: %v", path[0], err)
		}
	}
	t.Fatalf("走到 portal(%d,%d)後仍未抵達 CTY%d sec%d；目前 CTY%d sec%d @(%d,%d)，title=%v dlg=%v cmd=%v panel=%d gateStage=%d cd=%d path=%v",
		x, y, wantCty, wantSec, g.curCty, sceneSection(g.cur), g.px, g.py,
		g.showTitle, g.dlg.open, g.cmd.open, g.panel, g.sequenceGateStage, g.cd,
		tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier()))
}

func traceWalkTo(t *testing.T, g *Game, x, y int) {
	t.Helper()
	for i := 0; i < 3000 && (g.px != x || g.py != y); i++ {
		traceWalkOne(t, g, x, y)
	}
	if g.px != x || g.py != y {
		t.Fatalf("無法走到 (%d,%d)，停在 (%d,%d)，dlg=%v gateStage=%d cmd=%v path=%v",
			x, y, g.px, g.py, g.dlg.open, g.sequenceGateStage, g.cmd.open,
			tracePath(g.cur, g.px, g.py, x, y))
	}
}

func traceWalkToNoPortal(t *testing.T, g *Game, x, y int) {
	t.Helper()
	for i := 0; i < 3000 && (g.px != x || g.py != y); i++ {
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待無轉場路徑 cooldown: %v", err)
			}
			continue
		}
		path := tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier())
		if len(path) == 0 {
			t.Fatalf("無法在不踩 transition 下由 (%d,%d) 走到 (%d,%d)",
				g.px, g.py, x, y)
		}
		dx, dy := dirDelta(path[0])
		if g.cur.doorTier(g.px+dx, g.py+dy) != 0 {
			traceOpenReachableDoor(t, g, g.px+dx, g.py+dy)
			continue
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("無轉場路徑移動 dir%d: %v", path[0], err)
		}
	}
	if g.px != x || g.py != y {
		t.Fatalf("無轉場路徑無法走到 (%d,%d)，停在 (%d,%d)", x, y, g.px, g.py)
	}
}

func traceWalkOne(t *testing.T, g *Game, tx, ty int) {
	t.Helper()
	if g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待移動 cooldown: %v", err)
		}
		return
	}
	path := tracePath(g.cur, g.px, g.py, tx, ty)
	if len(path) == 0 {
		// 遊走 NPC 暫時擋路；送一個空白 frame 等它移動後重算。
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待 NPC: %v", err)
		}
		return
	}
	if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
		t.Fatalf("移動 dir%d: %v", path[0], err)
	}
}

func traceTalkNPC(t *testing.T, g *Game, nx, ny int) {
	t.Helper()
	type facePos struct{ x, y, dir int }
	var candidates []facePos
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		for dist := 1; dist <= 2; dist++ {
			x, y := nx-dx*dist, ny-dy*dist
			if x < 0 || y < 0 || x >= g.cur.w || y >= g.cur.h || g.cur.Blocked(x, y) {
				continue
			}
			if dist == 2 {
				mx, my := nx-dx, ny-dy
				if !g.cur.attr.Blocked(g.cur.tileIdx(mx, my)) {
					continue
				}
			}
			candidates = append(candidates, facePos{x, y, dir})
		}
	}
	for _, p := range candidates {
		if path := tracePath(g.cur, g.px, g.py, p.x, p.y); len(path) > 0 ||
			g.px == p.x && g.py == p.y {
			traceWalkTo(t, g, p.x, p.y)
			for g.cd > 0 {
				if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
					t.Fatalf("等待對話前 cooldown: %v", err)
				}
			}
			if err := g.step(InputState{DirHeld: p.dir, DirEdge: -1}); err != nil {
				t.Fatalf("面向 NPC: %v", err)
			}
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
				t.Fatalf("開命令窗: %v", err)
			}
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
				t.Fatalf("選對話: %v", err)
			}
			return
		}
	}
	t.Fatalf("無法由 CTY%d sec%d @(%d,%d) 抵達 NPC(%d,%d) 的相鄰格；候選=%v dn=%d keyTier=%d",
		g.curCty, sceneSection(g.cur), g.px, g.py, nx, ny, candidates, g.dnPhase, g.keyTier())
}

// tracePortalPath 把目標以外的 transition tile 視為阻擋；否則最短路徑可能先踩到
// 另一座樓梯，執行時場景已切換，靜態 graph 卻仍誤判可達。
func tracePortalPath(sc *Scene, sx, sy, tx, ty int, keyTier ...int) []int {
	if sc == nil || sx == tx && sy == ty {
		return nil
	}
	allowedDoorTier := 0
	if len(keyTier) > 0 {
		allowedDoorTier = keyTier[0]
	}
	type node struct{ x, y int }
	start, goal := node{sx, sy}, node{tx, ty}
	q := []node{start}
	seen := map[node]bool{start: true}
	prev := map[node]node{}
	prevDir := map[node]int{}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			n := node{p.x + dx, p.y + dy}
			doorTier := sc.doorTier(n.x, n.y)
			if seen[n] || n.x < 0 || n.y < 0 || n.x >= sc.w || n.y >= sc.h ||
				sc.Blocked(n.x, n.y) && !(doorTier > 0 && doorTier <= allowedDoorTier) {
				continue
			}
			if n != goal {
				if _, _, _, _, ok := sc.tileTransition(n.x, n.y); ok {
					continue
				}
			}
			seen[n], prev[n], prevDir[n] = true, p, dir
			if n == goal {
				var rev []int
				for cur := goal; cur != start; cur = prev[cur] {
					rev = append(rev, prevDir[cur])
				}
				for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
					rev[i], rev[j] = rev[j], rev[i]
				}
				return rev
			}
			q = append(q, n)
		}
	}
	return nil
}

// tracePath 對目前 production Scene 做 BFS，回方向鍵序列；Blocked 包含地形與當幀 NPC。
func tracePath(sc *Scene, sx, sy, tx, ty int) []int {
	if sc == nil || sx == tx && sy == ty {
		return nil
	}
	type node struct{ x, y int }
	start, goal := node{sx, sy}, node{tx, ty}
	q := []node{start}
	seen := map[node]bool{start: true}
	prev := map[node]node{}
	prevDir := map[node]int{}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			n := node{p.x + dx, p.y + dy}
			if seen[n] || n.x < 0 || n.y < 0 || n.x >= sc.w || n.y >= sc.h ||
				(n != goal && sc.Blocked(n.x, n.y)) {
				continue
			}
			if n == goal && sc.Blocked(n.x, n.y) {
				continue
			}
			seen[n], prev[n], prevDir[n] = true, p, dir
			if n == goal {
				var rev []int
				for cur := goal; cur != start; cur = prev[cur] {
					rev = append(rev, prevDir[cur])
				}
				for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
					rev[i], rev[j] = rev[j], rev[i]
				}
				return rev
			}
			q = append(q, n)
		}
	}
	return nil
}
