package game

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
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
	shopActorUsedSlots := func(actor int) int {
		items := g.equipActorInventory(actor)
		equipped := g.equipActorSlots(actor)
		if items == nil || equipped == nil {
			return 1 << 30
		}
		used := len(*items)
		for _, code := range equipped {
			if code >= 0 {
				used++
			}
		}
		return used
	}
	shopActorHasSpace := func(actor int) bool {
		return shopActorUsedSlots(actor) < g.pack.ItemActions().PersonalInventorySlots
	}
	partyInventoryFreeSlots := func() int {
		free := 0
		for actor := 0; actor <= len(g.companions); actor++ {
			if used := shopActorUsedSlots(actor); used < g.pack.ItemActions().PersonalInventorySlots {
				free += g.pack.ItemActions().PersonalInventorySlots - used
			}
		}
		return free
	}
	dropHerbsForPartyInventorySlots := func(want int) int {
		t.Helper()
		for partyInventoryFreeSlots() < want {
			discarded := false
			for actor := 0; actor <= len(g.companions); actor++ {
				items := g.equipActorInventory(actor)
				if items != nil && containsInt(*items, herbCode) {
					traceDropActorInventoryItem(t, g, actor, herbCode)
					discarded = true
					break
				}
			}
			if !discarded {
				break
			}
		}
		return partyInventoryFreeSlots()
	}
	dropSpareEquipmentForPartyInventorySlots := func(want int) int {
		t.Helper()
		dropHerbsForPartyInventorySlots(want)
		for partyInventoryFreeSlots() < want {
			dropped := false
			for actor := 0; actor <= len(g.companions); actor++ {
				items := g.equipActorInventory(actor)
				if items == nil {
					continue
				}
				for _, code := range *items {
					// equipActorInventory 不含目前 equipped slots；ITEM.DAT
					// 分類為裝備者才是可證明的備品，不碰劇情 raw ID。
					if g.shop.items.EquipSlot(code) < 0 {
						continue
					}
					traceDropActorInventoryItem(t, g, actor, code)
					dropped = true
					break
				}
				if dropped {
					break
				}
			}
			if !dropped {
				break
			}
		}
		return partyInventoryFreeSlots()
	}
	ensurePartyInventorySpace := func(reason string, want int) {
		t.Helper()
		if want <= 0 {
			return
		}
		if free := dropHerbsForPartyInventorySlots(want); free < want {
			t.Fatalf("%s：需要 %d 格、目前僅 %d 格，且沒有可經正式選單丟棄的藥草",
				reason, want, free)
		}
	}
	buyCurrentShopItemForActor := func(actor int) bool {
		t.Helper()
		if !g.shop.active || g.shop.targeting || !shopActorHasSpace(actor) {
			return false
		}
		beforeGold := g.heroGold
		press(InputState{Confirm: true})
		if !g.shop.targeting {
			return false
		}
		for g.shop.targetCursor != actor {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true})
		return !g.shop.targeting && g.heroGold < beforeGold
	}
	buyCurrentShopItemAnyActor := func(preferHero bool) (int, bool) {
		t.Helper()
		if preferHero && shopActorHasSpace(0) {
			return 0, buyCurrentShopItemForActor(0)
		}
		for actor := 0; actor <= len(g.companions); actor++ {
			if shopActorHasSpace(actor) {
				return actor, buyCurrentShopItemForActor(actor)
			}
		}
		return -1, false
	}

	press(InputState{Confirm: true}) // splash → menu
	press(InputState{Confirm: true}) // menu「遊戲開始」→ 注音命名
	press(InputState{Toggle: true})  // 注音 → 英數
	press(InputState{Confirm: true}) // cursor0：「0」
	for i := 0; i < 3; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0}) // raw0→raw27
	}
	for i := 0; i < 8; i++ {
		send(InputState{DirHeld: -1, DirEdge: 3}) // raw27→raw35(cell43,進功能列)
	}
	press(InputState{Confirm: true}) // cell43：進右側功能列
	for i := 0; i < 4; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0}) // 功能列 row1→row5
	}
	press(InputState{Confirm: true}) // 功能列「完成」→ 性別
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
		for j := 0; j < 3; j++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		for j := 0; j < 8; j++ {
			send(InputState{DirHeld: -1, DirEdge: 3})
		}
		press(InputState{Confirm: true}) // cell43：進功能列
		for j := 0; j < 4; j++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true}) // 功能列「完成」
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

	// 四人旅店每次 8G；國王的 50G 全數保留作練級住宿。先買藥草會在前段練級耗盡
	// 住宿費，因此這條 deterministic 正式路線以逃離強敵、旅店補滿為前段補給策略。

	// 國王給的裝備不會自動穿上；依正式命令窗逐人分配。
	equipItem := func(actor, code int) {
		t.Helper()
		if actor > 0 && !containsInt(g.companions[actor-1].Inventory, code) {
			traceGiveInventoryItem(t, g, code, actor-1)
		}
		var items []int
		if actor == 0 {
			items = g.inventory
		} else {
			items = g.companions[actor-1].Inventory
		}
		idx := -1
		for i, item := range items {
			if item == code {
				idx = i
				break
			}
		}
		if idx < 0 {
			t.Fatalf("隊員%d沒有待裝備 item %#x：%v", actor, code, items)
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
	// 在阿里阿罕入口旁的原版低危 region 以正式移動／戰鬥／旅店輸入練到 Lv20；
	// 不寫入 EXP、HP、金幣或 RNG，所有成長與補給都必須走 production consumer。
	traceTrainNearTown(t, g, 0, 20)
	if g.dnPhase != 0 {
		traceWaitForDayNearCty(t, g, 0)
	}
	// 原版掉落 writer 會在全隊八格皆滿時拒絕後續物品；低危練級已讓
	// 勇者物品欄合法裝滿藥草。以正式 rec421「丟掉」釋放一格，保留給
	// 馬上要買並在地表使用的聖水，不直接改 slice 或擴大容量。
	if !shopActorHasSpace(0) {
		traceDropInventoryItem(t, g, herbCode)
	}

	// CTY07/08 是有遭遇表的塔區；正式玩家會先到雷貝道具店買聖水，
	// 再以道具選單驅敵穿越地道。阿里阿罕的原始貨架沒有聖水，
	// 因此先由正式地表路線到雷貝採買，避免把戰鬥 RNG 誤當成
	// CTY07→08 transition 不可達。
	traceAdventureWalkToCty(t, g, 1)
	traceTalkFacility(t, g, facItem)
	najimiHolyIdx := -1
	for i, code := range g.shop.codes {
		if code == itemuse.ItemHolyWater {
			najimiHolyIdx = i
			break
		}
	}
	if najimiHolyIdx < 0 {
		t.Fatalf("雷貝道具店沒有聖水：codes=%v", g.shop.codes)
	}
	for bought := 0; bought < 1 && g.heroGold >= g.shop.items.Price(itemuse.ItemHolyWater); bought++ {
		for g.shop.cursor != najimiHolyIdx {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		if _, ok := buyCurrentShopItemAnyActor(true); !ok {
			t.Fatal("雷貝道具店無法把聖水放入勇者個人物品欄")
		}
	}
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceAdventureWalkToCty(t, g, 7) // 阿里阿罕西側地道入口；塔本體隔海，不能由地表直走
	traceTownSectionTo(t, g, 8, 3)
	traceTalkNPC(t, g, 9, 9) // CTY08 sec3 sub2 handler12：巴可達老人
	if !g.hasPartyItem(0x55) || !g.progressDone(msThiefKey) || !g.dlg.open {
		t.Fatalf("正式對話後應取得盜賊鑰匙：item=%v milestone=%v dlg=%v",
			g.hasPartyItem(0x55), g.progressDone(msThiefKey), g.dlg.open)
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
	if !g.hasPartyItem(0x58) || !g.hasPartyItem(0x55) || !g.progressDone(msMagicBal) || !g.dlg.open {
		t.Fatalf("正式雷貝老人對話交易錯：ball=%v key=%v milestone=%v dlg=%v",
			g.hasPartyItem(0x58), g.hasPartyItem(0x55), g.progressDone(msMagicBal), g.dlg.open)
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
	if _, ok := buyCurrentShopItemAnyActor(true); !ok {
		t.Fatal("雷貝補給無法把聖水放入勇者個人物品欄")
	}
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
		if _, ok := buyCurrentShopItemAnyActor(false); !ok {
			break
		}
	}
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceAdventureWalkToCty(t, g, 9)
	traceTownSectionTo(t, g, ctyTemptationCave, magicBallSection)
	traceWalkTo(t, g, magicBallLeftX, magicBallUseY)
	traceUseInventoryItem(t, g, itemuse.ItemMagicBall)
	if g.hasPartyItem(itemuse.ItemMagicBall) || g.storyFlag(magicBallIntactFlag) {
		t.Fatalf("production 道具選單破牆後狀態錯：ball=%v flag51=%v",
			g.hasPartyItem(itemuse.ItemMagicBall), g.storyFlag(magicBallIntactFlag))
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
	if !g.storyFlag(roleEvent.PendingFlagRaw) || g.hasPartyItem(roleEvent.RequiredItemRawID) {
		t.Fatalf("金皇冠任務交付後狀態錯：flag2c=%v crown=%v",
			g.storyFlag(roleEvent.PendingFlagRaw), g.hasPartyItem(roleEvent.RequiredItemRawID))
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
	// 正確 RNG 序列下，羅馬利亞外 region 在 Lv15 可出現三隻中階敵，
	// 不是低危練級區。現在已正式造訪 CTY0／CTY2 且旅店補滿 MP；
	// 依 docs/154 用正式魯拉回阿里阿罕的已驗證低階 region 練級，
	// 再魯拉回羅馬利亞整備；不寫 visited town、MP、座標或 region。
	traceRuraToCty(t, g, 0)
	traceTrainNearTown(t, g, 0, 25)
	traceAdventureWalkToCty(t, g, 0)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	traceAdventureWalkToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	// 戰鬥掉落可能已合法填滿勇者／戰士八格；依 docs/158 走原版 owner
	// selector 各丟一株藥草，替龜殼甲胄／青銅盾保留欄位，不直接改 inventory。
	if !shopActorHasSpace(0) {
		traceDropActorInventoryItem(t, g, 0, herbCode)
	}
	if !shopActorHasSpace(1) {
		traceDropActorInventoryItem(t, g, 1, herbCode)
	}
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
	if !buyCurrentShopItemForActor(1) {
		t.Fatal("羅馬利亞武防店無法把青銅盾交給戰士")
	}
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
	if !buyCurrentShopItemForActor(0) {
		t.Fatal("羅馬利亞武防店無法把龜殼甲胄交給勇者")
	}
	press(InputState{Cancel: true})
	equipItem(1, bronzeShield) // 戰士正式換上青銅盾
	equipItem(0, tortoiseArmor)
	if g.companions[0].Shield != bronzeShield {
		t.Fatalf("戰士未經正式裝備面板換上青銅盾：%#x", g.companions[0].Shield)
	}
	if g.equip[1] != tortoiseArmor {
		t.Fatalf("勇者未經正式裝備面板換上龜殼甲胄：%#x", g.equip[1])
	}
	// 換裝會把舊甲胄退回個人物品欄；若全隊再次滿格，正式丟掉一株
	// 藥草，確保下一間店仍可購入北行所需的聖水。
	partyHasSpace := false
	for actor := 0; actor <= len(g.companions); actor++ {
		partyHasSpace = partyHasSpace || shopActorHasSpace(actor)
	}
	if !partyHasSpace {
		traceDropActorInventoryItem(t, g, 0, herbCode)
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
			preferHero := code != herbCode
			if _, ok := buyCurrentShopItemAnyActor(preferHero); !ok {
				break
			}
			bought++
		}
		return bought
	}
	topUpHolyWater := func(want int) {
		t.Helper()
		before := g.countPartyItem(itemuse.ItemHolyWater)
		if before >= want {
			return
		}
		need := want - before
		if partyInventoryFreeSlots() < need {
			if g.shop.active {
				press(InputState{Cancel: true})
			}
			free := dropHerbsForPartyInventorySlots(need)
			if free < need {
				t.Logf("聖水補給受原版個人容量限制：策略上限=%d、原有=%d、可新增=%d",
					want, before, free)
				need = free
			}
			traceTalkFacility(t, g, facItem)
		}
		if bought := buyFromOpenShop(itemuse.ItemHolyWater, need); bought != need ||
			g.countPartyItem(itemuse.ItemHolyWater) != before+need {
			t.Fatalf("正式補給聖水失敗：need=%d bought=%d before=%d after=%d gold=%d inventory=%v",
				need, bought, before, g.countPartyItem(itemuse.ItemHolyWater), g.heroGold, g.inventory)
		}
	}
	topUpAntidote := func(want int) {
		t.Helper()
		before := g.countPartyItem(itemuse.ItemAntidote)
		if before >= want {
			return
		}
		need := want - before
		if bought := buyFromOpenShop(itemuse.ItemAntidote, need); bought != need ||
			g.countPartyItem(itemuse.ItemAntidote) != want {
			t.Fatalf("正式補給驅毒草失敗：need=%d bought=%d before=%d after=%d gold=%d inventory=%v",
				need, bought, before, g.countPartyItem(itemuse.ItemAntidote), g.heroGold, g.inventory)
		}
	}
	topUpHerb := func(want int) {
		t.Helper()
		before := g.countPartyItem(itemuse.ItemHerb)
		if before >= want {
			return
		}
		need := want - before
		if free := partyInventoryFreeSlots(); free < need {
			t.Logf("藥草補給受原版個人容量限制：策略上限=%d、原有=%d、可新增=%d",
				want, before, free)
			need = free
		}
		if bought := buyFromOpenShop(itemuse.ItemHerb, need); bought != need ||
			g.countPartyItem(itemuse.ItemHerb) != before+need {
			t.Fatalf("正式補給藥草失敗：need=%d bought=%d before=%d after=%d gold=%d inventory=%v",
				need, bought, before, g.countPartyItem(itemuse.ItemHerb), g.heroGold, g.inventory)
		}
	}
	buyBestFromOpenShop := func(slot, class int) int {
		t.Helper()
		bestCode, bestValue := -1, -1
		for _, code := range g.shop.codes {
			value := g.shop.items.Defense(code)
			if slot == 0 {
				value = g.shop.items.Attack(code)
			}
			if g.shop.items.EquipSlot(code) == slot && g.shop.items.CanEquip(code, class) &&
				g.shop.items.Price(code) <= g.heroGold && value > bestValue {
				bestCode, bestValue = code, value
			}
		}
		if bestCode < 0 {
			t.Fatalf("商店沒有職業%d可負擔的部位%d裝備：gold=%d codes=%v",
				class, slot, g.heroGold, g.shop.codes)
		}
		for i, code := range g.shop.codes {
			if code != bestCode {
				continue
			}
			for g.shop.cursor != i {
				send(InputState{DirHeld: -1, DirEdge: 0})
			}
			break
		}
		if !buyCurrentShopItemForActor(0) {
			t.Fatalf("正式購買裝備 %#x 時勇者個人物品欄已滿", bestCode)
		}
		if !g.hasItem(bestCode) {
			t.Fatalf("正式購買裝備 %#x 後未進勇者物品欄：inventory=%v", bestCode, g.inventory)
		}
		return bestCode
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
	// 巴拉摩斯戰後仍要以正式魯拉返回阿里阿罕；本輪戰術保留
	// 施法資源，對應原版玩家在終戰前保留回城 MP 的必要條件。
	traceResolveBattle(t, g, false, true)
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
	if !g.hasPartyItem(roleEvent.RequiredItemRawID) {
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
		!g.hasPartyItem(roleEvent.RequiredItemRawID) {
		t.Fatalf("金皇冠讀檔後未能正式返回羅馬利亞：town=%v cty=%d crown=%v",
			g.inTown, g.curCty, g.hasPartyItem(roleEvent.RequiredItemRawID))
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
	if g.hasPartyItem(roleEvent.RequiredItemRawID) || g.storyFlag(roleEvent.PendingFlagRaw) {
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
	if !g.hasPartyItem(treasure.ItemRawID) || g.storyFlag(treasure.PresentFlag) {
		t.Fatalf("正式調查 CTY11 sec3 (21,20) 未取得夢幻紅寶石：%v", g.inventory)
	}
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, exchange.NPC.CTYRaw)
	traceTownSectionTo(t, g, exchange.NPC.CTYRaw, exchange.NPC.Section,
		exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	traceTalkNPC(t, g, exchange.NPC.Tile.X, exchange.NPC.Tile.Y)
	exchangeText, _ := g.pack.TextGlyphCodes(exchange.SuccessTextID)
	if g.hasPartyItem(exchange.RequiredItemRawID) || !g.hasPartyItem(exchange.GrantedItemRawID) ||
		!g.dlg.open || !reflect.DeepEqual(g.dlg.buf, exchangeText) {
		t.Fatalf("精靈女王未正式以夢幻紅寶石換覺醒粉：ruby=%v powder=%v dlg=%v",
			g.hasPartyItem(exchange.RequiredItemRawID), g.hasPartyItem(exchange.GrantedItemRawID), g.dlg.open)
	}
	traceCloseDialogue(t, g)
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, noanielEvent.Use.CTYRaw)
	traceUseInventoryItem(t, g, noanielEvent.Use.ItemRawID)
	awakenText, _ := g.pack.TextGlyphCodes(noanielEvent.Use.SuccessTextID)
	if g.hasPartyItem(noanielEvent.Use.ItemRawID) ||
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
		g.hasPartyItem(noanielEvent.Use.ItemRawID) ||
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
		}
		traceExitTownBoundary(t, g)
	}
	// CTY12／32／73 的原版 cty_loc 均為 `(36,105)`，但 file 0x43a6
	// 從 CTY0 向上掃描並在第一筆命中停止，因此地表正式入口是 CTY12。
	// CTY32／73 是城內王宮區，不是必須依序踩過的重疊地表入口。
	traceAdventureWalkToCty(t, g, 12)
	traceTownSectionTo(t, g, 12, 1)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	// 依席斯武防店是正常主線已經過且日夜穩定的補給點。品項由原始商店表、能力與
	// 可裝職業欄決定；正式買下並從裝備面板穿上，不把特定裝備 ID 寫成測試捷徑。
	for g.pack.ItemActions().PersonalInventorySlots-shopActorUsedSlots(0) < 2 {
		discard := -1
		for _, code := range g.inventory {
			if code == herbCode || g.shop.items.EquipSlot(code) >= 0 {
				discard = code
				break
			}
		}
		if discard < 0 {
			t.Fatalf("依席斯購買兩件裝備前無可正式丟棄的藥草／舊裝備：equip=%v inv=%v",
				g.equip, g.inventory)
		}
		traceDropActorInventoryItem(t, g, 0, discard)
	}
	traceTalkFacility(t, g, facWeapon)
	isisWeapon := buyBestFromOpenShop(0, 0)
	isisArmor := buyBestFromOpenShop(1, 0)
	press(InputState{Cancel: true})
	equipItem(0, isisWeapon)
	equipItem(0, isisArmor)
	traceTalkFacility(t, g, facItem)
	buyFromOpenShop(herbCode, 5)
	press(InputState{Cancel: true})
	traceTownSectionTo(t, g, 12, 0)
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
	ensurePartyInventorySpace("金字塔魔法鑰匙前整理背包", 1)
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
	if !g.hasPartyItem(magicKey.ItemRawID) || g.storyFlag(magicKey.PresentFlag) {
		partyInventory := [][]int{append([]int(nil), g.inventory...)}
		partyEquipment := [][]int{append([]int(nil), g.equip[:]...)}
		for _, member := range g.companions {
			partyInventory = append(partyInventory, append([]int(nil), member.Inventory...))
			partyEquipment = append(partyEquipment,
				[]int{member.Weapon, member.Armor, member.Shield, member.Head})
		}
		t.Fatalf("正式調查未取得魔法鑰匙：item=%v flag=%v inventories=%v equipment=%v companions=%v",
			g.hasPartyItem(magicKey.ItemRawID), g.storyFlag(magicKey.PresentFlag),
			partyInventory, partyEquipment, g.companions)
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
	if !g.hasPartyItem(magicKey.ItemRawID) || g.storyFlag(magicKey.PresentFlag) ||
		g.storyFlag(pyramidGate.ClearFlagRaw) || g.curCty != pyramidGate.CTYRaw ||
		sceneSection(g.cur) != pyramidGate.Section || g.cur.Blocked(13, 10) {
		t.Fatalf("魔法鑰匙 checkpoint round-trip 錯：key=%v present=%v gate=%v cty=%d sec=%d blocked=%v",
			g.hasPartyItem(magicKey.ItemRawID), g.storyFlag(magicKey.PresentFlag),
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
	traceAdventureWalkToCty(t, g, 34)
	traceTownSectionTo(t, g, 35, 0)
	traceExitTownBoundary(t, g, true)
	traceAdventureWalkToCty(t, g, 16)
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
	ensurePartyInventorySpace("波魯多加國王授信前整理背包", 1)
	traceTalkNPC(t, g, portoga.NPC.Tile.X, portoga.NPC.Tile.Y)
	if !g.dlg.open || g.vehicleExchangeStage != vehicleExchangeQuestIntro ||
		g.hasPartyItem(portoga.GrantedItemRawID) {
		t.Fatalf("波魯多加首訪未停在授信介紹：dlg=%v stage=%d item=%v",
			g.dlg.open, g.vehicleExchangeStage, g.hasPartyItem(portoga.GrantedItemRawID))
	}
	traceCloseDialogue(t, g)
	if !g.hasPartyItem(portoga.GrantedItemRawID) ||
		g.storyFlag(portoga.QuestPresentFlagRaw) || g.shipOwned {
		t.Fatalf("波魯多加首訪交易錯誤：letter=%v questFlag=%v ship=%v",
			g.hasPartyItem(portoga.GrantedItemRawID),
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
	if !g.hasPartyItem(portoga.GrantedItemRawID) ||
		g.storyFlag(portoga.QuestPresentFlagRaw) || g.shipOwned ||
		g.curCty != portoga.NPC.CTYRaw || sceneSection(g.cur) != portoga.NPC.Section {
		t.Fatalf("國王信件 save/load round-trip 錯誤：item=%v flag=%v ship=%v cty=%d sec=%d",
			g.hasPartyItem(portoga.GrantedItemRawID), g.storyFlag(portoga.QuestPresentFlagRaw),
			g.shipOwned, g.curCty, sceneSection(g.cur))
	}

	// 國王信件 checkpoint → 反向穿過羅馬利亞祠堂。CTY34 出口記住
	// cty_loc 的左格 (38,53)，往東一格仍屬同一個兩格寬入口，會正常重入
	// CTY34；因此使用已習得的魯拉，從正式命令窗回到已造訪的羅馬利亞，
	// 再徒步前往阿莎拉慕東方 CTY62。全程不注入場景或座標。
	traceTownSectionTo(t, g, 16, 0)
	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, 35)
	traceTownSectionTo(t, g, 34, 0)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 2)
	guidedEvents := g.pack.GuidedPassageEvents()
	if len(guidedEvents) != 2 {
		t.Fatalf("guided passage events=%d, want CTY62/63 variants", len(guidedEvents))
	}
	// 持信首訪由山脈西側 CTY62 spawn=(3,8) 進入；CTY63
	// spawn=(94,10) 是通道另一側，從羅馬利亞徒步不可達。
	norud := guidedEvents[0]
	traceAdventureWalkToCty(t, g, norud.GuideNPC.CTYRaw)
	start := norud.InteractionTile
	traceWalkTo(t, g, start.X, start.Y)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	talkDir := -1
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		if start.X+dx == norud.GuideNPC.Tile.X &&
			start.Y+dy == norud.GuideNPC.Tile.Y {
			talkDir = dir
			break
		}
	}
	if talkDir < 0 {
		t.Fatalf("諾魯多 interaction_tile 未與 guide_npc 相鄰：start=%+v npc=%+v",
			start, norud.GuideNPC.Tile)
	}
	send(InputState{DirHeld: talkDir, DirEdge: -1}) // 面向 NPC；碰撞不移動。
	press(InputState{Confirm: true})                // 開命令窗。
	press(InputState{Confirm: true})                // 選「對話」。
	if !g.dlg.open || g.guidedPassageStage != guidedPassageIntroduction {
		t.Fatalf("正式交談未命中諾魯多 rec87：dlg=%v stage=%d",
			g.dlg.open, g.guidedPassageStage)
	}
	traceCloseDialogue(t, g) // rec87 → rec88 → 啟動引路者移動。
	for frames := 0; frames < 2000 && g.guidedPassageStage != guidedPassageAwaitTrigger; frames++ {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if g.guidedPassageStage != guidedPassageAwaitTrigger || g.dlg.open ||
		!g.storyFlag(norud.AnimationPendingFlagRaw) {
		t.Fatalf("諾魯多引路後未交回玩家控制：stage=%d dlg=%v pending=%v",
			g.guidedPassageStage, g.dlg.open, g.storyFlag(norud.AnimationPendingFlagRaw))
	}
	trigger := norud.CompletionTrigger.Tiles[0]
	traceWalkTo(t, g, trigger.X, trigger.Y)
	if g.guidedPassageStage != guidedPassageCompletionDialogue || !g.dlg.open {
		t.Fatalf("走到 handler57 trigger 後未開 rec90：stage=%d dlg=%v pos=(%d,%d)",
			g.guidedPassageStage, g.dlg.open, g.px, g.py)
	}
	traceCloseDialogue(t, g)
	for frames := 0; frames < 2000 && g.guidedPassageStage != guidedPassageIdle; frames++ {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if g.guidedPassageStage != guidedPassageIdle ||
		g.storyFlag(norud.GuidePresentFlagRaw) ||
		g.storyFlag(norud.AnimationPendingFlagRaw) ||
		!g.storyFlag(norud.CompletedFlagRaw) ||
		!g.hasPartyItem(norud.RequiredItemRawID) ||
		g.px != trigger.X || g.py != trigger.Y {
		t.Fatalf("諾魯特正式引路交易錯：stage=%d flags=%v/%v/%v letter=%v pos=(%d,%d)",
			g.guidedPassageStage, g.storyFlag(norud.GuidePresentFlagRaw),
			g.storyFlag(norud.AnimationPendingFlagRaw), g.storyFlag(norud.CompletedFlagRaw),
			g.hasPartyItem(norud.RequiredItemRawID), g.px, g.py)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存諾魯特密道 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建諾魯特密道讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if g.storyFlag(norud.GuidePresentFlagRaw) ||
		!g.storyFlag(norud.CompletedFlagRaw) ||
		!g.hasPartyItem(norud.RequiredItemRawID) ||
		g.curCty != norud.GuideNPC.CTYRaw ||
		sceneSection(g.cur) != norud.GuideNPC.Section ||
		g.px != trigger.X || g.py != trigger.Y {
		t.Fatalf("諾魯特 checkpoint round-trip 錯：flags=%v/%v letter=%v cty=%d sec=%d pos=(%d,%d)",
			g.storyFlag(norud.GuidePresentFlagRaw), g.storyFlag(norud.CompletedFlagRaw),
			g.hasPartyItem(norud.RequiredItemRawID), g.curCty, sceneSection(g.cur), g.px, g.py)
	}

	// 諾魯德 checkpoint → 明確走 CTY transition table 的東側出口。密道同時有
	// world (81,81)/(84,81) 兩個合法出入口，不能讓通用 BFS 任選第一個西側出口。
	eastSideCTY := guidedEvents[1].GuideNPC.CTYRaw
	eastX, eastY := ctyLoc[eastSideCTY][0], ctyLoc[eastSideCTY][1]
	exitX, exitY := -1, -1
	for y := 0; y < g.cur.h && exitX < 0; y++ {
		for x := 0; x < g.cur.w; x++ {
			_, dsec, dx, dy, ok := g.cur.tileTransition(x, y)
			if ok && dsec == 0xff && dx == eastX && dy == eastY {
				exitX, exitY = x, y
				break
			}
		}
	}
	if exitX < 0 {
		t.Fatalf("CTY%d 找不到原始東側 world transition (%d,%d)",
			g.curCty, eastX, eastY)
	}
	traceWalkThroughPortal(t, g, exitX, exitY, -1, -1)
	if g.px != eastX || g.py != eastY {
		t.Fatalf("諾魯德東側離場未使用原始 world transition 座標：got (%d,%d), want (%d,%d)",
			g.px, g.py, eastX, eastY)
	}
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 3, DirEdge: -1})
	if g.inTown || g.px != eastX+1 || g.py != eastY || g.worldEntranceGrace {
		t.Fatalf("諾魯德離場後首步未通過原版入口第二格：town=%v pos=(%d,%d) grace=%v",
			g.inTown, g.px, g.py, g.worldEntranceGrace)
	}
	traceAdventureWalkToCty(t, g, 15)
	if !g.inTown || g.curCty != 15 || sceneSection(g.cur) != 0 {
		t.Fatalf("諾魯德東口後未能正式抵達巴哈拉達：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存巴哈拉達 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建巴哈拉達讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 15 || sceneSection(g.cur) != 0 ||
		!g.storyFlag(norud.CompletedFlagRaw) ||
		!g.hasPartyItem(norud.RequiredItemRawID) {
		t.Fatalf("巴哈拉達 checkpoint round-trip 錯：town=%v cty=%d sec=%d complete=%v letter=%v",
			g.inTown, g.curCty, sceneSection(g.cur),
			g.storyFlag(norud.CompletedFlagRaw), g.hasPartyItem(norud.RequiredItemRawID))
	}

	rescueEvents := g.pack.HostageRescueEvents()
	if len(rescueEvents) != 1 {
		t.Fatalf("hostage rescue events=%d, want 1", len(rescueEvents))
	}
	rescue := rescueEvents[0]
	// 巴哈拉達 checkpoint → 徒步前往東北洞窟 CTY14。所有交談、選項、
	// 場景 trigger 與戰鬥均經正式 InputState；不直接設定救援旗標。
	traceExitTownBoundary(t, g)
	// 巴哈拉達周邊有會連續施放全體不能行動異常的編隊；正常玩家以已習得的魯拉
	// 回羅馬利亞低危區準備，不靠不穩定的逃跑亂數或測試狀態注入。
	traceRuraToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	traceTrainNearTown(t, g, roleEvent.OfferNPC.CTYRaw, 18)
	traceAdventureWalkToCty(t, g, roleEvent.OfferNPC.CTYRaw)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, rescue.RewardNPC.CTYRaw)
	traceAdventureWalkToCty(t, g, rescue.RewardNPC.CTYRaw)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceTalkFacility(t, g, facItem)
	buyFromOpenShop(herbCode, 8)
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)
	traceAdventureWalkToCty(t, g, rescue.ApproachTrigger.CTYRaw)
	traceTownSectionTo(t, g, rescue.ApproachTrigger.CTYRaw,
		rescue.ApproachTrigger.Section)
	guard := rescue.GuardNPCs[0]
	traceWalkToNoPortal(t, g, guard.Tile.X, guard.Tile.Y-1)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 0, DirEdge: -1}) // 守衛北側面向下；碰撞不移動。
	press(InputState{Confirm: true})
	press(InputState{Confirm: true})
	if !g.dlg.open || g.hostageRescueStage != hostageRescueGuardQuestion {
		t.Fatalf("正式交談未命中 handler58：dlg=%v stage=%d",
			g.dlg.open, g.hostageRescueStage)
	}
	traceCloseDialogue(t, g)
	send(InputState{DirHeld: -1, DirEdge: 0}) // 選「否」並迎戰。
	if g.hostageRescueStage != hostageRescueGuardChoice || g.hostageRescueCursor != 1 {
		t.Fatalf("守衛選項無法以正式方向輸入移到「否」：stage=%d cursor=%d",
			g.hostageRescueStage, g.hostageRescueCursor)
	}
	press(InputState{Confirm: true})
	if !g.dlg.open || g.hostageRescueStage != hostageRescueGuardFight {
		t.Fatalf("拒絕加入後未播放 rec108：dlg=%v stage=%d cursor=%d buf=%d",
			g.dlg.open, g.hostageRescueStage, g.hostageRescueCursor, len(g.dlg.buf))
	}
	traceCloseDialogue(t, g)
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatalf("handler58 未啟動四名 0x39 守衛：active=%v enemies=%v",
			g.battle.active, g.battle.enemies)
	}
	traceResolveBattle(t, g, false)
	if g.storyFlag(rescue.GuardPresentFlagRaw) {
		t.Fatal("正式擊敗守衛後未 clear 原版 flag0x80")
	}

	approach := rescue.ApproachTrigger.Tiles[0]
	traceWalkTo(t, g, approach.X, approach.Y)
	if !g.dlg.open || g.hostageRescueStage != hostageRescueApproachCry {
		t.Fatalf("正式踏入 handler59 未播放 rec109：dlg=%v stage=%d pos=(%d,%d)",
			g.dlg.open, g.hostageRescueStage, g.px, g.py)
	}
	traceCloseDialogue(t, g)
	for frames := 0; frames < 2000 && g.hostageRescueAnimating(); frames++ {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if g.hostageRescueStage != hostageRescueIdle ||
		g.storyFlag(rescue.ApproachPendingFlagRaw) {
		t.Fatalf("handler59 移動後狀態錯：stage=%d flag81=%v",
			g.hostageRescueStage, g.storyFlag(rescue.ApproachPendingFlagRaw))
	}

	sw := rescue.SwitchTrigger.Tiles[0]
	traceWalkTo(t, g, sw.X, sw.Y)
	if !g.dlg.open || g.hostageRescueStage != hostageRescueSwitchPrompt {
		t.Fatalf("正式踏入 handler60 未播放按鍵 prompt：dlg=%v stage=%d",
			g.dlg.open, g.hostageRescueStage)
	}
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true}) // 按下牆上按鍵。
	if !g.dlg.open {
		t.Fatalf("按鍵接受後未開 rec87：stage=%d", g.hostageRescueStage)
	}
	// traceCloseDialogue 以正式 A 逐頁，依序關閉 rec87→110→111，
	// 最後才啟動兩名俘虜的 movement。
	traceCloseDialogue(t, g)
	for frames := 0; frames < 3000 && g.hostageRescueAnimating(); frames++ {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if g.hostageRescueStage != hostageRescueBossArrival || !g.dlg.open {
		t.Fatalf("俘虜移動後未播放 rec112：stage=%d dlg=%v",
			g.hostageRescueStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if g.storyFlag(rescue.CaptivePresentFlagRaw) ||
		!g.storyFlag(rescue.BossPresentFlagRaw) {
		t.Fatalf("救援後未 clear0x82/set0x14：captive=%v boss=%v",
			g.storyFlag(rescue.CaptivePresentFlagRaw),
			g.storyFlag(rescue.BossPresentFlagRaw))
	}

	boss := rescue.BossNPCs[0]
	traceTalkNPC(t, g, boss.Tile.X, boss.Tile.Y)
	if !g.dlg.open || g.hostageRescueStage != hostageRescueBossChallenge {
		t.Fatalf("正式交談未命中 handler61：dlg=%v stage=%d",
			g.dlg.open, g.hostageRescueStage)
	}
	traceCloseDialogue(t, g)
	if !g.battle.active || len(g.battle.enemies) != 4 {
		t.Fatalf("handler61 未啟動 0x38+0x39×3：active=%v enemies=%v",
			g.battle.active, g.battle.enemies)
	}
	traceResolveBattle(t, g, false)
	if !g.dlg.open || g.hostageRescueStage != hostageRescueBossDefeated {
		t.Fatalf("正式擊敗甘達特後未播放 rec124：dlg=%v stage=%d heroHP=%d",
			g.dlg.open, g.hostageRescueStage, g.heroHP)
	}
	traceCloseDialogue(t, g)
	traceCloseDialogue(t, g)
	if g.hostageRescueStage != hostageRescueBossChoice {
		t.Fatalf("rec114 後未進求饒選項：stage=%d", g.hostageRescueStage)
	}
	send(InputState{DirHeld: -1, DirEdge: 0}) // 先拒絕一次。
	press(InputState{Confirm: true})
	traceCloseDialogue(t, g)
	if g.hostageRescueStage != hostageRescueBossChoice {
		t.Fatalf("rec115 後未回求饒選項：stage=%d", g.hostageRescueStage)
	}
	press(InputState{Confirm: true}) // 再接受。
	traceCloseDialogue(t, g)
	if g.storyFlag(rescue.BossPresentFlagRaw) ||
		g.storyFlag(0x34) || !g.storyFlag(0x25) {
		t.Fatalf("求饒完成旗標錯：flag14=%v flag34=%v flag25=%v",
			g.storyFlag(rescue.BossPresentFlagRaw), g.storyFlag(0x34), g.storyFlag(0x25))
	}

	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, rescue.RewardNPC.CTYRaw)
	traceTownSectionTo(t, g, rescue.RewardNPC.CTYRaw, rescue.RewardNPC.Section)
	ensurePartyInventorySpace("古布達交付黑胡椒前整理背包", 1)
	traceTalkNPC(t, g, rescue.RewardNPC.Tile.X, rescue.RewardNPC.Tile.Y)
	if !g.dlg.open || g.hostageRescueStage != hostageRescueRewardOffer {
		t.Fatalf("救人後古布達未播放 rec118：dlg=%v stage=%d",
			g.dlg.open, g.hostageRescueStage)
	}
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true})
	if !g.dlg.open || g.hostageRescueStage != hostageRescueRewardReceived ||
		!g.hasPartyItem(rescue.RewardItemRawID) ||
		g.storyFlag(rescue.RewardAvailableFlagRaw) {
		t.Fatalf("黑胡椒交易錯：dlg=%v stage=%d item=%v flag36=%v",
			g.dlg.open, g.hostageRescueStage, g.hasPartyItem(rescue.RewardItemRawID),
			g.storyFlag(rescue.RewardAvailableFlagRaw))
	}
	traceCloseDialogue(t, g)
	if err := g.Save(); err != nil {
		t.Fatalf("保存黑胡椒 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatal(err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入黑胡椒 checkpoint
	if !g.hasPartyItem(rescue.RewardItemRawID) ||
		!g.storyFlag(0x25) ||
		g.storyFlag(0x34) ||
		g.storyFlag(rescue.RewardAvailableFlagRaw) {
		t.Fatal("巴哈拉塔救援與黑胡椒未通過 save/load round-trip")
	}

	// 黑胡椒 checkpoint → 波魯多加取船。從實際讀回的存檔繼續，只經正式魯拉、
	// 地表入口、transition 與 cmdTalk；交換道具、旗標與停泊座標皆取自 game pack。
	traceTownSectionTo(t, g, rescue.RewardNPC.CTYRaw, 0)
	// 接下來連續跨海到加爾那之塔、達瑪再回巴哈拉達；在具備道具店的
	// 巴哈拉達以正式購買補足聖水，避免把後續持久麻痺遭遇交給幸運 RNG。
	partyFreeSlots := func() int {
		free := 0
		for actor := 0; actor <= len(g.companions); actor++ {
			if n := g.pack.ItemActions().PersonalInventorySlots - shopActorUsedSlots(actor); n > 0 {
				free += n
			}
		}
		return free
	}
	for partyFreeSlots() < 8 {
		discardActor := -1
		for actor := 0; actor <= len(g.companions); actor++ {
			items := g.equipActorInventory(actor)
			if items != nil && containsInt(*items, herbCode) {
				discardActor = actor
				break
			}
		}
		if discardActor < 0 {
			t.Fatalf("巴哈拉塔補給前無法以正式丟棄清出八格：free=%d hero=%v",
				partyFreeSlots(), g.inventory)
		}
		traceDropActorInventoryItem(t, g, discardActor, herbCode)
	}
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn) // 巴哈拉塔迷宮後以正常旅店恢復魯拉所需 MP。
	traceTownSectionTo(t, g, -1, -1)
	traceRuraToCty(t, g, 16)
	traceAdventureWalkToCty(t, g, 16)
	traceTownSectionTo(t, g, portoga.NPC.CTYRaw, portoga.NPC.Section)
	traceTalkNPC(t, g, portoga.NPC.Tile.X, portoga.NPC.Tile.Y)
	if !g.dlg.open || !g.shipOwned || g.shipAboard ||
		g.hasPartyItem(portoga.RequiredItemRawID) ||
		!g.storyFlag(portoga.ExchangeAvailableFlagRaw) {
		t.Fatalf("黑胡椒換船交易錯：dlg=%v owned=%v aboard=%v pepper=%v flag38=%v",
			g.dlg.open, g.shipOwned, g.shipAboard,
			g.hasPartyItem(portoga.RequiredItemRawID),
			g.storyFlag(portoga.ExchangeAvailableFlagRaw))
	}
	for _, flag := range portoga.Vehicle.ClearFlagsRaw {
		if g.storyFlag(flag) {
			t.Fatalf("正式授船後未清 vehicle flag 0x%02x", flag)
		}
	}
	wantShip := portoga.Vehicle.WorldPosition
	if g.shipX != wantShip.X || g.shipY != wantShip.Y || g.layer != wantShip.Layer {
		t.Fatalf("正式取船停泊位置=(%d,%d,L%d)，want (%d,%d,L%d)",
			g.shipX, g.shipY, g.layer, wantShip.X, wantShip.Y, wantShip.Layer)
	}
	traceCloseDialogue(t, g)
	if err := g.Save(); err != nil {
		t.Fatalf("保存取船 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建取船讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入取船 checkpoint
	if !g.shipOwned || g.shipAboard ||
		g.shipX != wantShip.X || g.shipY != wantShip.Y ||
		g.hasPartyItem(portoga.RequiredItemRawID) ||
		!g.storyFlag(portoga.ExchangeAvailableFlagRaw) {
		t.Fatalf("取船 checkpoint round-trip 錯：owned=%v aboard=%v ship=(%d,%d) pepper=%v flag38=%v",
			g.shipOwned, g.shipAboard, g.shipX, g.shipY,
			g.hasPartyItem(portoga.RequiredItemRawID),
			g.storyFlag(portoga.ExchangeAvailableFlagRaw))
	}
	for _, flag := range portoga.Vehicle.ClearFlagsRaw {
		if g.storyFlag(flag) {
			t.Fatalf("取船 checkpoint 讀回後 vehicle flag 0x%02x 復原", flag)
		}
	}

	traceTownSectionTo(t, g, 16, 0)
	traceTownSectionTo(t, g, -1, -1)
	traceBoardAndSailShip(t, g)

	// 首次航行 → 加爾那之塔。船陸混合尋路只選方向鍵；航行、靠岸、
	// 地表入口、塔內 transition 與寶箱調查仍全部由 production step 處理。
	satoriEvent, ok := g.pack.TreasureEvent("dq3:event.garuna_satori_book")
	if !ok {
		t.Fatal("缺 dq3:event.garuna_satori_book")
	}
	satori := satoriEvent.Treasure
	// 遠洋沿岸含必定嘗試持久麻痺的獵人蜂；使用背包中實際購得的聖水
	// 維持正式驅敵路徑，不以重擲 RNG 或削弱原版 action3 穿越。
	traceAdventureTravelToCty(t, g, satori.CTYRaw, true, true)
	traceTownSectionTo(t, g, satori.CTYRaw, satori.Section)
	traceExaminePackTreasure(t, g, satori)
	if !g.hasPartyItem(satori.ItemRawID) || g.storyFlag(satori.PresentFlag) {
		t.Fatalf("正式調查未取得領悟之書：item=%v flag=%v inventory=%v",
			g.hasPartyItem(satori.ItemRawID), g.storyFlag(satori.PresentFlag), g.inventory)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存領悟之書 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建領悟之書讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})          // 標題 splash → 主選單
	send(InputState{DirHeld: -1, DirEdge: 0}) // 遊戲開始 → 載入進度
	press(InputState{Confirm: true})          // 正式載入領悟之書 checkpoint
	if !g.hasPartyItem(satori.ItemRawID) || g.storyFlag(satori.PresentFlag) ||
		g.curCty != satori.CTYRaw || sceneSection(g.cur) != satori.Section {
		t.Fatalf("領悟之書 checkpoint round-trip 錯：item=%v flag=%v CTY%d sec%d",
			g.hasPartyItem(satori.ItemRawID), g.storyFlag(satori.PresentFlag),
			g.curCty, sceneSection(g.cur))
	}

	// 領悟之書 checkpoint → 達瑪神殿。先以正式路徑離塔，於達瑪附近正常
	// 遭遇練到「實際要轉職的同伴」Lv20；不可只檢查勇者等級。
	traceTownSectionTo(t, g, satori.CTYRaw, 0)
	traceExitTownBoundary(t, g)
	traceAdventureTravelToCty(t, g, 17)
	traceExitTownBoundary(t, g)
	for heroTarget := stats.LevelForExp(0, g.heroExp); g.companions[0].Level() < 20; heroTarget++ {
		if heroTarget >= stats.MaxLevel {
			t.Fatalf("正式練級到 hero cap 仍未讓同伴達 Lv20：hero%d companion%d",
				stats.LevelForExp(0, g.heroExp), g.companions[0].Level())
		}
		traceTrainNearTown(t, g, 17, heroTarget+1)
	}
	traceAdventureWalkToCty(t, g, 17)

	// 原版賢者 gate 查「所選隊員的八格個人物品」，不是全隊共用背包。
	// 經正式 rec421 道具動作選單，把領悟之書給第一名同伴。
	traceGiveInventoryItem(t, g, satori.ItemRawID, 0)
	if !g.hasPartyItem(satori.ItemRawID) ||
		!containsInt(g.companions[0].Inventory, satori.ItemRawID) {
		t.Fatalf("正式給予領悟之書失敗：hero=%v companion=%v",
			g.inventory, g.companions[0].Inventory)
	}

	// CTY17 sec0 handler39：(介紹→隊員→賢者→兩次確認→Lv1 transaction)。
	traceOpenReachableDoor(t, g)
	traceTalkNPC(t, g, 10, 10)
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true}) // 是否轉職：是
	traceCloseDialogue(t, g)
	send(InputState{DirHeld: -1, DirEdge: 0}) // 主角→第一名同伴
	press(InputState{Confirm: true})
	traceCloseDialogue(t, g)
	for i := 0; i < 5; i++ { // 原版 5 基本職業後第 6 項才是賢者
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true})
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true}) // 確認目標職業
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true}) // 確認從 Lv1 開始
	traceCloseDialogue(t, g)         // success → transaction → farewell
	if g.reclassStage != reclassIdle || g.companions[0].Class != 5 ||
		g.companions[0].Level() != 1 ||
		containsInt(g.companions[0].Inventory, satori.ItemRawID) {
		t.Fatalf("正式達瑪轉賢者未閉合：stage=%d class=%d level=%d items=%v",
			g.reclassStage, g.companions[0].Class, g.companions[0].Level(),
			g.companions[0].Inventory)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存達瑪轉職 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建達瑪轉職讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if len(g.companions) == 0 || g.companions[0].Class != 5 ||
		g.companions[0].Level() != 1 ||
		containsInt(g.companions[0].Inventory, satori.ItemRawID) {
		t.Fatalf("達瑪轉職 save/load round-trip 錯：companions=%+v", g.companions)
	}

	// 達瑪 checkpoint → 提頓。由 section0 出城、登船、航行與靠岸；
	// 不直接改晝夜、座標或 vehicle state。
	traceTownSectionTo(t, g, 17, 0)
	traceOpenReachableDoor(t, g) // save/load 後原版魔法門回到關閉 tile
	traceExitTownBoundary(t, g)
	// 轉職賢者降回 Lv1；先徒步回具備旅店與教會的巴哈拉達，在該段
	// 正式海路遇敵時選逃跑，再於城外低危區練回 Lv20。這條船路可抽到
	// monster 48 的四體編隊，不能把它誤當成可讓剛轉職隊伍硬戰的低危區；
	// 逃跑仍完全經由正式戰鬥選單與原版成功判定。
	traceAdventureTravelToCty(t, g, 15, true, true)
	// 剛轉職的賢者仍是 Lv1。先在實際抵達的城鎮完成正常補給與住宿，
	// 再到入口附近練功；不可把耗盡的航路補給或主角的高等級誤當成全隊戰力。
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	// 巴哈拉達入口 region 包含 monster 48 的多體 formation；即使主角
	// 已高等，剛轉職的賢者也不該靠逃跑成功率賭過這段。已造訪的
	// 阿里阿罕入口是此前完整驗證過的低危區，因此以正式魯拉返回、
	// 正式練功，並在完成後再返回達瑪的船路。
	traceRuraToCty(t, g, 0)
	for heroTarget := stats.LevelForExp(0, g.heroExp); g.companions[0].Level() < 20; heroTarget++ {
		if heroTarget >= stats.MaxLevel {
			t.Fatalf("正式練級到 hero cap 仍未讓轉職賢者達 Lv20：hero%d sage%d",
				stats.LevelForExp(0, g.heroExp), g.companions[0].Level())
		}
		traceTrainNearTown(t, g, 0, heroTarget+1)
	}
	traceAdventureWalkToCty(t, g, 0)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, 17)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 20)
	// 航程可能跨過晝夜門檻；先正常出村，在地表走到白天後重新進村，
	// 再進武器店二樓，不能直接寫 dnPhase。
	if g.dnPhase != 0 {
		traceTownSectionTo(t, g, 20, 0)
		traceExitTownBoundary(t, g)
		traceWaitForDayNearCty(t, g, 20)
		traceAdventureWalkToCty(t, g, 20, false)
	}
	if g.dnPhase != 0 {
		t.Fatalf("重新進提頓後不是白天：phase=%d step=%d", g.dnPhase, g.dnStep)
	}
	darkLampEvent, ok := g.pack.TreasureEvent("dq3:event.teidon_dark_lamp")
	if !ok {
		t.Fatal("缺 dq3:event.teidon_dark_lamp")
	}
	darkLamp := darkLampEvent.Treasure
	darkLampEffect, ok := g.pack.ItemUseEffectByRawID(darkLamp.ItemRawID)
	if !ok || darkLampEffect.DayNightClock == nil {
		t.Fatal("缺黑暗之燈原版 day_night_clock")
	}
	traceTownSectionTo(t, g, darkLamp.CTYRaw, darkLamp.Section)
	traceExaminePackTreasure(t, g, darkLamp)
	if !g.hasPartyItem(darkLamp.ItemRawID) || g.storyFlag(darkLamp.PresentFlag) {
		t.Fatalf("正式取得黑暗燈 transaction 錯：item=%v flag=%v",
			g.hasPartyItem(darkLamp.ItemRawID), g.storyFlag(darkLamp.PresentFlag))
	}

	// 先回 section0，原版 handler logical 0x4063 才允許在地表使用。
	traceTownSectionTo(t, g, darkLamp.CTYRaw, 0)
	// 原版 handler logical 0x4063 只允許在地表使用。正常離開提頓後經
	// rec421 選「使用」，強制夜晚但保留黑暗燈。
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, darkLamp.ItemRawID)
	if g.dayNightClock() != *darkLampEffect.DayNightClock || !g.hasPartyItem(darkLamp.ItemRawID) {
		t.Fatalf("正式使用黑暗燈未閉合：phase=%d step=%d item=%v",
			g.dnPhase, g.dnStep, g.hasPartyItem(darkLamp.ItemRawID))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存提頓黑暗燈 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建提頓黑暗燈讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if g.inTown || g.dayNightClock() != *darkLampEffect.DayNightClock ||
		!g.hasPartyItem(darkLamp.ItemRawID) || g.storyFlag(darkLamp.PresentFlag) {
		t.Fatalf("提頓黑暗燈 save/load round-trip 錯：town=%v phase=%d step=%d item=%v flag=%v",
			g.inTown, g.dnPhase, g.dnStep, g.hasPartyItem(darkLamp.ItemRawID),
			g.storyFlag(darkLamp.PresentFlag))
	}

	// 剛完成賢者轉職的隊伍若直接穿越海路，可能帶著逃跑失敗的傷害與
	// 耗盡的 MP 進入八頭大蛇固定戰。CTY20 的現行原始 NPC 場景沒有
	// 可達旅店 NPC，故依正式船路先到有確認旅店的 CTY22 補給，再前往
	// CTY19；不偽造提頓設施或直接補滿數值。
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 22, true, true)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 19, true, true)
	if !g.inTown || g.curCty != 19 {
		t.Fatalf("黑暗燈後無法繼續到八頭大蛇洞窟：town=%v cty=%d", g.inTown, g.curCty)
	}

	// CTY19 合法 checkpoint → handler45 第一戰 → 原版掉落草薙大劍 → 宮殿 rec70
	// → handler36 No 分支第二戰 → 同格寶箱取得紫寶珠。全程只走 transition、NPC 對話、
	// 戰鬥及調查命令，不直接呼叫 staged boss handler 或寫旗標／座標／道具。
	orochi, ok := g.pack.StagedBossEvent("dq3:event.jipang_orochi")
	if !ok {
		t.Fatal("缺 dq3:event.jipang_orochi")
	}
	traceTownSectionTo(t, g, orochi.FirstNPC.CTYRaw, orochi.FirstNPC.Section,
		orochi.FirstNPC.Tile.X, orochi.FirstNPC.Tile.Y)
	// handler45 戰後把玩家向右推兩格進宮殿；原版互動路線必須從 NPC 左側交談，
	// 若從右側開戰，守衛自身的右移會占住玩家路徑而無法觸發同一轉場。
	traceWalkTo(t, g, orochi.FirstNPC.Tile.X-1, orochi.FirstNPC.Tile.Y)
	// 最後一格也可能開啟 CTY 的原始遭遇 modal。到達座標不代表該 frame
	// 已交回地圖輸入；先以正式戰鬥選單結束遭遇，再面向固定 NPC。
	for waits := 0; g.cd > 0 || g.battle.active; waits++ {
		if waits >= 8192 {
			t.Fatalf("大蛇 NPC 前 cooldown／戰鬥未收斂：cd=%d battle=%v pos=(%d,%d)",
				g.cd, g.battle.active, g.px, g.py)
		}
		if g.battle.active {
			traceResolveBattle(t, g, true)
			continue
		}
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 3, DirEdge: -1}) // 面向右側 NPC
	press(InputState{Confirm: true})          // 開命令窗
	press(InputState{Confirm: true})          // 選對話
	if !g.battle.active || g.stagedBossStage != stagedBossFirstBattle || g.battle.monID != 75 {
		t.Fatalf("handler45 未由正式對話開第一戰：active=%v stage=%d mon=%d",
			g.battle.active, g.stagedBossStage, g.battle.monID)
	}
	traceResolveBattle(t, g, false)
	// 原版掉落草薙大劍會留下需要玩家確認的訊息；先以正式確認鍵關閉，
	// 戰後 NPC／玩家移動腳本才會取得輸入所有權。
	for frames := 0; frames < 1000 && g.stagedBossStage == stagedBossFirstMoving; frames++ {
		// 勝利、經驗與掉落訊息可能依序取得 modal；逐一以正式確認鍵推進。
		if g.dlg.open {
			traceCloseDialogue(t, g)
			continue
		}
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if !g.hasPartyItem(0x14) || g.hasPartyItem(0x69) || g.storyFlag(0x44) || !g.storyFlag(0x20) ||
		g.curCty != 21 || sceneSection(g.cur) != 0 || g.stagedBossStage != stagedBossFirstPost || !g.dlg.open {
		t.Fatalf("第一戰後 transaction／轉場錯：item14=%v item69=%v flags44/20=%v/%v cty=%d sec=%d pos=(%d,%d) stage=%d dlg=%v",
			g.hasPartyItem(0x14), g.hasPartyItem(0x69), g.storyFlag(0x44), g.storyFlag(0x20),
			g.curCty, sceneSection(g.cur), g.px, g.py, g.stagedBossStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	// 兩階段事件不強迫玩家連戰。第一戰後先以正式出口、船、教會及旅店到 CTY22
	// 復原，再由地表正式入口回日邦格拒絕無姬；不直接補 HP／MP 或復活角色。
	traceTownSectionTo(t, g, 21, 0)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 22, true, true)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 21, true)
	traceTalkNPC(t, g, orochi.SecondNPC.Tile.X, orochi.SecondNPC.Tile.Y)
	traceCloseDialogue(t, g)
	if g.stagedBossStage != stagedBossSecondChoice {
		t.Fatalf("rec66 後未進第二戰選項：stage=%d", g.stagedBossStage)
	}
	send(InputState{DirHeld: -1, DirEdge: 0}) // 選「否」
	press(InputState{Confirm: true})
	if g.stagedBossStage != stagedBossSecondChallenge || !g.dlg.open {
		t.Fatalf("No 未顯示 rec68：stage=%d dlg=%v", g.stagedBossStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if !g.battle.active || g.stagedBossStage != stagedBossSecondBattle || !g.battle.suppressDrop {
		t.Fatalf("rec68 後未開抑制掉落的第二戰：active=%v stage=%d suppress=%v",
			g.battle.active, g.stagedBossStage, g.battle.suppressDrop)
	}
	traceResolveBattle(t, g, false)
	if g.hasPartyItem(0x69) || g.storyFlag(0x20) || !g.storyFlag(0x1f) ||
		g.stagedBossStage != stagedBossSecondVictory || !g.dlg.open {
		t.Fatalf("第二戰勝利 transaction 錯：item69=%v flags20/1f=%v/%v stage=%d dlg=%v",
			g.hasPartyItem(0x69), g.storyFlag(0x20), g.storyFlag(0x1f), g.stagedBossStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	traceExaminePackTreasure(t, g, gamepack.QuestTreasureSelector{
		CTYRaw: 21, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x69, PresentFlag: 0x89,
	})
	if !g.hasPartyItem(0x69) || !g.storyFlag(0x89) {
		t.Fatalf("正式調查無姬寶箱未取得紫寶珠：item=%v flag89=%v",
			g.hasPartyItem(0x69), g.storyFlag(0x89))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存日邦格 checkpoint：%v", err)
	}
	restored, err = NewGame(g.assets, nil)
	if err != nil {
		t.Fatalf("重建日邦格讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 21 || !g.hasPartyItem(0x14) || !g.hasPartyItem(0x69) ||
		g.storyFlag(0x20) || !g.storyFlag(0x1f) || !g.storyFlag(0x89) {
		t.Fatalf("日邦格 save/load round-trip 錯：town=%v cty=%d item14/69=%v/%v flags20/1f/89=%v/%v/%v",
			g.inTown, g.curCty, g.hasPartyItem(0x14), g.hasPartyItem(0x69), g.storyFlag(0x20),
			g.storyFlag(0x1f), g.storyFlag(0x89))
	}

	// 日邦格 checkpoint → 原始 CTY38 道具店。docs/40 的 SHP.DAT 貨架盤點證實此店
	// 含隱身草 0x5d；不把朗錫爾 CTY47／勇氣神殿態 CTY75 的地名線索誤當貨架 oracle。
	// 全程不直接 grant item 或呼叫 shop transaction。
	traceTownSectionTo(t, g, 21, 0)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 38)
	traceTalkFacility(t, g, facItem)
	if !g.shop.active {
		t.Fatal("CTY38 道具店未由正式 facility NPC 開啟")
	}
	if bought := buyFromOpenShop(0x5d, 1); bought != 1 || !g.hasPartyItem(0x5d) {
		t.Fatalf("CTY38 正式購買隱身草失敗：bought=%d gold=%d inventory=%v",
			bought, g.heroGold, g.inventory)
	}
	// CTY38 的原始貨架同時販售聖水；以正式商店交易補到八瓶，後續再在
	// 正常抵達的港口補足，而不是把高階海域的逃跑機率當成規格。
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	// 乾渴壺後的 CTY40 沒有邊界出口，原版只能靠魯拉離開。
	// 進耶進貝亞前先在 CTY38 以正式旅店交易補滿 MP，不注入魔力，
	// 也不把實況影片中修改過的 999 MP 當成原版參數。
	traceTalkFacility(t, g, facInn)

	// CTY38 → 耶進貝亞：由正式地表／船路抵達 CTY39，在守衛前才從道具選單
	// 使用隱身草。透明狀態使原始 handler29 不再把 slot0 守衛橫移到玩家同欄；
	// 玩家由右欄越過後，再由原始 transition 進 CTY76 地下室。
	traceTownSectionTo(t, g, 38, 0)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	// 這是進入無步行出口的乾涸淺灘前最後一個旅店；後續陸海路程
	// 遇故時優先用正式逃跑命令，保留魯拉所需 MP。
	traceAdventureTravelToCty(t, g, 39, true)
	traceWalkTo(t, g, 14, 38)
	traceUseInventoryItem(t, g, 0x5d)
	if g.hasItem(0x5d) || g.remoaru != 0x19 {
		t.Fatalf("正式使用隱身草 transaction 錯：item=%v timer=%d", g.hasItem(0x5d), g.remoaru)
	}
	traceWalkTo(t, g, 14, 36)
	if g.cur.npcAt(13, 36) < 0 || g.cur.npcAt(14, 36) >= 0 || g.remoaru != 0x17 {
		t.Fatalf("隱身越過耶進貝亞守衛失敗：player=(%d,%d) timer=%d npcs=%+v",
			g.px, g.py, g.remoaru, g.cur.npcs)
	}
	traceTownSectionTo(t, g, 76, 0)
	if g.curCty != 76 || g.cur.sec != 0 {
		t.Fatalf("守衛 gate 後未由原始 transition 抵達 CTY76：cty=%d sec=%d", g.curCty, g.cur.sec)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存愛丁貝亞地下室 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建愛丁貝亞讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 76 || sceneSection(g.cur) != 0 || g.hasItem(0x5d) {
		t.Fatalf("愛丁貝亞地下室 save/load round-trip 錯：town=%v cty=%d sec=%d grass=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.hasItem(0x5d))
	}

	// CTY76 原始 NPC slot0..2 依 ctrl bit0x40 接受推動；走完整 119 步正式解路，
	// 最後一推令 handler30 清 flag0x3b 並打開 passage，再由命令窗調查 event0
	// 取得乾渴壺。全程只送正式方向與命令輸入，沒有座標或事件狀態注入。
	traceEginbearPushPuzzle(t, g)
	if g.storyFlag(0x3b) || g.remoaru != 0 {
		t.Fatalf("愛丁貝亞推石完成狀態錯：flag3b=%v timer=%d", g.storyFlag(0x3b), g.remoaru)
	}
	traceExaminePackTreasure(t, g, gamepack.QuestTreasureSelector{
		CTYRaw: 76, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x5e, PresentFlag: 0x3a,
	})
	if !g.hasPartyItem(0x5e) || g.storyFlag(0x3a) {
		t.Fatalf("正式取得乾渴壺 transaction 錯：item=%v flag3a=%v",
			g.hasPartyItem(0x5e), g.storyFlag(0x3a))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存乾渴壺 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建乾渴壺讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 76 || sceneSection(g.cur) != 0 || !g.hasPartyItem(0x5e) ||
		g.storyFlag(0x3a) || g.storyFlag(0x3b) {
		t.Fatalf("乾渴壺 save/load round-trip 錯：town=%v cty=%d sec=%d item=%v flags3a/3b=%v/%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.hasPartyItem(0x5e),
			g.storyFlag(0x3a), g.storyFlag(0x3b))
	}

	// 乾渴壺 checkpoint → 正常離開地下室、登船並航行到原版唯一成功格
	// (0x92,0x35)。正式道具選單使用後，pack 的 5x4 world patch 顯出
	// (0x92,0x34) 祠堂入口；向上正常移動進 CTY40，再調查 event0 取得最終鑰匙。
	traceTownSectionTo(t, g, 39, 0)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	drain, ok := g.pack.ItemUseEffectByRawID(0x5e)
	if !ok || drain.UseTile == nil || drain.MapPatch == nil {
		t.Fatalf("乾渴壺 pack effect 缺失：%+v", drain)
	}
	// 乾涸淺灘的祠堂沒有步行出口；玩家需預留 8 MP 以魯拉離開。
	// 這段遠洋以正式戰鬥選單逃跑，不讓 trace 的自動攻擊策略耗盡魔力。
	// 遠洋段以既有正式聖水／特黑洛斯保護航路；毒氣 E2 接線後，僅靠
	// 逐場逃跑會讓正常隊伍在抵達位置式道具格前耗盡生命。這仍只送
	// production 物品／咒文／方向輸入，不修改 encounter counter 或狀態。
	traceSailToWorldCoordinate(t, g, drain.UseTile.X, drain.UseTile.Y, true)
	traceUseInventoryItem(t, g, drain.ItemRawID)
	if g.worldState&uint16(drain.SetWorldStateMask) == 0 || !g.hasPartyItem(drain.ItemRawID) ||
		g.over.tileIdx(0x92, 0x34) != 0x7c {
		t.Fatalf("乾渴壺正式使用 transaction 錯：state=%#x item=%v entrance=%#x",
			g.worldState, g.hasPartyItem(drain.ItemRawID), g.over.tileIdx(0x92, 0x34))
	}
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 1, DirEdge: -1})
	if !g.inTown || g.curCty != 40 || sceneSection(g.cur) != 0 {
		t.Fatalf("祠堂 world patch 未由正式上移進 CTY40：town=%v cty=%d sec=%d pos=(%d,%d)",
			g.inTown, g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	finalKey, ok := g.pack.TreasureEvent("dq3:event.shoal_final_key")
	if !ok {
		t.Fatal("缺 dq3:event.shoal_final_key")
	}
	traceExaminePackTreasure(t, g, finalKey.Treasure)
	if !g.hasPartyItem(0x57) || g.storyFlag(0x47) {
		t.Fatalf("正式取得最終鑰匙 transaction 錯：item=%v flag47=%v",
			g.hasPartyItem(0x57), g.storyFlag(0x47))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存最終鑰匙 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建最終鑰匙讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 40 || !g.hasPartyItem(0x57) || g.storyFlag(0x47) ||
		g.worldState&uint16(drain.SetWorldStateMask) == 0 || g.over.tileIdx(0x92, 0x34) != 0x7c {
		t.Fatalf("最終鑰匙 save/load round-trip 錯：town=%v cty=%d item=%v flag47=%v state=%#x entrance=%#x",
			g.inTown, g.curCty, g.hasPartyItem(0x57), g.storyFlag(0x47), g.worldState,
			g.over.tileIdx(0x92, 0x34))
	}

	// CTY40 checkpoint → 依原版影片由正式咒文選單施放魯拉回巴哈拉達；
	// 魯拉 handler 同步把船移到 pack 的原版落點 (104,126)。在地表以黑暗之燈
	// 確保黑夜後，正式走到船、登船與航行進 CTY20。原版牢門不是「調查」自動開啟；必須從 rec421 正式使用
	// tier3 最終鑰匙。開門後與夜間 subtype2/handler35 NPC 對話取得綠寶珠。
	traceRuraToCty(t, g, 15)
	// 巴哈拉達是乾涸淺灘後、前往提頓與蘭西爾前最後能正常補給的
	// 道具店；進城、購買、出城與登船皆走 production path。
	traceAdventureWalkToCty(t, g, 15)
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn) // 補給後回滿魯拉與後續航路所需 MP。
	traceExitTownBoundary(t, g)
	darkLampEffect, found := g.pack.ItemUseEffectByRawID(0x5f)
	if !found {
		t.Fatal("缺 dq3:item_use.dark_lamp")
	}
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 20, true, true)
	// 遠洋航行會正常推進日夜；靠岸後先由 section0 正式出村，
	// 再在提頓入口外使用黑暗之燈並重新進入，不直接寫 day/night state。
	traceTownSectionTo(t, g, 20, 0)
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, darkLampEffect.ItemRawID)
	if g.dnPhase != darkLampEffect.DayNightPhase {
		t.Fatalf("返回提頓前未以正式道具進入黑夜：phase=%d want=%d",
			g.dnPhase, darkLampEffect.DayNightPhase)
	}
	traceAdventureWalkToCty(t, g, 20, false)
	if !g.inTown || g.curCty != 20 || g.dnPhase != 2 || !g.cur.night {
		t.Fatalf("未以正式夜間航路進提頓：town=%v cty=%d phase=%d sceneNight=%v",
			g.inTown, g.curCty, g.dnPhase, g.cur.night)
	}
	traceOpenReachableDoor(t, g, 17, 4)
	var greenOrb *gamepack.NPCItemRewardEvent
	for _, event := range g.pack.NPCItemRewardEvents() {
		if event.ID == "dq3:event.teidon_green_orb" {
			e := event
			greenOrb = &e
			break
		}
	}
	if greenOrb == nil {
		t.Fatal("缺 dq3:event.teidon_green_orb")
	}
	partyHasGreenOrb := func(game *Game) bool {
		if game.hasItem(greenOrb.GrantedItemRaw) {
			return true
		}
		for _, member := range game.companions {
			if containsInt(member.Inventory, greenOrb.GrantedItemRaw) {
				return true
			}
		}
		return false
	}
	traceTalkNPC(t, g, greenOrb.NPC.Tile.X, greenOrb.NPC.Tile.Y)
	if !partyHasGreenOrb(g) || g.storyFlag(greenOrb.PresentFlagRaw) {
		t.Fatalf("提頓夜間綠寶珠 transaction 錯：item=%v flag=%v",
			partyHasGreenOrb(g), g.storyFlag(greenOrb.PresentFlagRaw))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存提頓綠寶珠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建提頓綠寶珠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 20 || g.dnPhase != 2 || !partyHasGreenOrb(g) ||
		g.storyFlag(greenOrb.PresentFlagRaw) {
		t.Fatalf("提頓綠寶珠 save/load round-trip 錯：town=%v cty=%d phase=%d item=%v flag=%v",
			g.inTown, g.curCty, g.dnPhase, partyHasGreenOrb(g),
			g.storyFlag(greenOrb.PresentFlagRaw))
	}
	// 綠寶珠合法 checkpoint → 正常離開提頓、登船航行到蘭西爾。最終鑰匙
	// 開門後從正式「對話」選擇接受勇氣試煉；handler37 的原版 transaction
	// 暫存三名同伴並把勇者送到 (82,165)。再由正常地表移動進 CTY23，穿越
	// 勇氣洞窟取得藍寶珠，回 CTY75 與 handler62 交談後原樣復隊。
	traceTownSectionTo(t, g, 20, 0)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 47, true, true)
	if g.dnPhase != 0 {
		traceTalkFacility(t, g, facInn)
		if g.dnPhase != 0 {
			t.Fatalf("蘭西爾正式住宿後未回白天：phase=%d", g.dnPhase)
		}
	}
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	challenge, ok := g.pack.TemporarySoloChallenge("dq3:event.lancel_courage_trial")
	if !ok {
		t.Fatal("缺 dq3:event.lancel_courage_trial")
	}
	traceOpenReachableDoor(t, g)
	if !g.inTown || g.curCty != challenge.EntryNPC.CTYRaw ||
		g.cur.npcAt(challenge.EntryNPC.Tile.X, challenge.EntryNPC.Tile.Y) < 0 {
		t.Fatalf("白天開門後找不到勇氣試煉神官：town=%v cty=%d phase=%d npc=%d",
			g.inTown, g.curCty, g.dnPhase,
			g.cur.npcAt(challenge.EntryNPC.Tile.X, challenge.EntryNPC.Tile.Y))
	}
	wantParty := compsToSav(g.companions)
	if len(wantParty) != 3 {
		t.Fatalf("勇氣試煉前隊伍同伴=%d, want 3", len(wantParty))
	}
	if !shopActorHasSpace(0) {
		targetActor := -1
		for actor := 1; actor <= len(g.companions); actor++ {
			if shopActorHasSpace(actor) {
				targetActor = actor
				break
			}
		}
		if targetActor < 0 {
			for actor := 1; actor <= len(g.companions); actor++ {
				items := g.equipActorInventory(actor)
				if items != nil && containsInt(*items, herbCode) {
					traceDropActorInventoryItem(t, g, actor, herbCode)
					targetActor = actor
					break
				}
			}
		}
		if targetActor < 0 && g.hasPartyItem(itemuse.ItemHolyWater) {
			traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
			for actor := 1; actor <= len(g.companions); actor++ {
				if shopActorHasSpace(actor) {
					targetActor = actor
					break
				}
			}
		}
		heroItems := g.equipActorInventory(0)
		moveIdx := -1
		if heroItems != nil {
			for i, code := range *heroItems {
				if code != 0x55 && code != 0x56 && code != 0x57 {
					moveIdx = i
					break
				}
			}
		}
		if targetActor < 0 || heroItems == nil || moveIdx < 0 {
			t.Fatalf("勇氣試煉前無法以正式給予替勇者保留空格：hero=%v companions=%v",
				heroItems, g.companions)
		}
		traceGiveActorInventoryItem(t, g, 0, moveIdx, targetActor)
	}
	if !shopActorHasSpace(0) {
		t.Fatal("勇氣試煉前正式重分配後勇者仍無空格")
	}
	// handler37 應保存正式重分配完成後的完整同伴 records；不能拿整理前快照
	// 誤判合法的 inventory 移交為隊伍資料損壞。
	wantParty = compsToSav(g.companions)
	traceTalkNPC(t, g, challenge.EntryNPC.Tile.X, challenge.EntryNPC.Tile.Y, true)
	traceCloseDialogue(t, g)
	if g.soloChallengeStage != soloChallengeChoice {
		t.Fatalf("勇氣試煉 prompt 後 stage=%d, want choice", g.soloChallengeStage)
	}
	press(InputState{Confirm: true}) // 正式 Yes
	traceCloseDialogue(t, g)
	if !g.soloChallengeActive || len(g.companions) != 0 ||
		!reflect.DeepEqual(compsToSav(g.soloChallengeCompanions), wantParty) ||
		g.inTown || g.px != challenge.SoloWorldPosition.X || g.py != challenge.SoloWorldPosition.Y {
		t.Fatalf("正式接受勇氣試煉 transaction 錯：active=%v comps=%d saved=%d town=%v pos=(%d,%d)",
			g.soloChallengeActive, len(g.companions), len(g.soloChallengeCompanions),
			g.inTown, g.px, g.py)
	}
	if g.storyFlag(challenge.CompletedFlagRaw) {
		t.Fatal("勇氣試煉入口不得捏造 unknown 的 completed flag writer")
	}

	traceAdventureWalkToCty(t, g, 23, true)
	traceTownSectionTo(t, g, 23, 2)
	blueOrb, ok := g.pack.TreasureEvent("dq3:event.courage_cave_blue_orb")
	if !ok {
		t.Fatal("缺 dq3:event.courage_cave_blue_orb")
	}
	traceExaminePackTreasure(t, g, blueOrb.Treasure)
	if !g.hasItem(blueOrb.Treasure.ItemRawID) || g.storyFlag(blueOrb.Treasure.PresentFlag) ||
		!g.soloChallengeActive || len(g.companions) != 0 {
		t.Fatalf("單人取得藍寶珠 transaction 錯：item=%v flag=%v active=%v comps=%d",
			g.hasItem(blueOrb.Treasure.ItemRawID), g.storyFlag(blueOrb.Treasure.PresentFlag),
			g.soloChallengeActive, len(g.companions))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存單人藍寶珠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建單人藍寶珠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 23 || sceneSection(g.cur) != 2 ||
		!g.soloChallengeActive || len(g.companions) != 0 || len(g.soloChallengeCompanions) != 3 ||
		!g.hasItem(blueOrb.Treasure.ItemRawID) || g.storyFlag(blueOrb.Treasure.PresentFlag) {
		t.Fatalf("單人藍寶珠 save/load 錯：town=%v cty=%d sec=%d active=%v comps=%d saved=%d item=%v flag=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.soloChallengeActive,
			len(g.companions), len(g.soloChallengeCompanions), g.hasItem(blueOrb.Treasure.ItemRawID),
			g.storyFlag(blueOrb.Treasure.PresentFlag))
	}
	// 洞窟毒氣可能讓單人主角帶著持久 poison 回程。背包補給由先前正式商店交易取得；
	// 若確實中毒，經正常道具清單→使用→選主角操作驅毒草，不直接改 condition。
	if g.heroConditions&conditionPoison != 0 {
		traceUseInventoryItem(t, g, itemuse.ItemAntidote, 0)
		traceCloseDialogue(t, g)
		if g.heroConditions&conditionPoison != 0 {
			t.Fatal("正式使用驅毒草後主角仍中毒")
		}
	}

	traceTownSectionTo(t, g, -1, -1)
	traceAdventureWalkToCty(t, g, 75, true)
	traceTalkNPC(t, g, challenge.ReturnNPC.Tile.X, challenge.ReturnNPC.Tile.Y, true)
	traceCloseDialogue(t, g)
	destination := challenge.ReturnTownDestination
	if g.soloChallengeActive || len(g.soloChallengeCompanions) != 0 ||
		!reflect.DeepEqual(compsToSav(g.companions), wantParty) || !g.inTown ||
		g.curCty != destination.CTYRaw || sceneSection(g.cur) != destination.Section ||
		g.px != destination.X || g.py != destination.Y {
		t.Fatalf("勇氣試煉返回復隊錯：active=%v comps=%d saved=%d cty=%d sec=%d pos=(%d,%d)",
			g.soloChallengeActive, len(g.companions), len(g.soloChallengeCompanions),
			g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存勇氣試煉復隊 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建勇氣試煉復隊讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != destination.CTYRaw || sceneSection(g.cur) != destination.Section ||
		g.soloChallengeActive || len(g.soloChallengeCompanions) != 0 ||
		!reflect.DeepEqual(compsToSav(g.companions), wantParty) ||
		!g.hasItem(blueOrb.Treasure.ItemRawID) || g.storyFlag(blueOrb.Treasure.PresentFlag) {
		t.Fatalf("勇氣試煉復隊 save/load 錯：town=%v cty=%d sec=%d active=%v comps=%d item=%v flag=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.soloChallengeActive,
			len(g.companions), g.hasItem(blueOrb.Treasure.ItemRawID),
			g.storyFlag(blueOrb.Treasure.PresentFlag))
	}

	// 藍寶珠復隊 checkpoint → 正常離開蘭西爾、回船航行至海盜村。CTY27
	// 一般入口只通往孤立房間；正式路徑必須把 transition subid1 同格、ctrl bit0x40
	// 的物件向上推開，玩家踏入原格後才進 sec1 寶箱區。
	traceRuraToCty(t, g, 15) // 原船不可由 (82,165) 出口區域徒步取回；魯拉同步重定位船。
	// 試煉與魯拉已消耗勇者 MP；在可達的巴哈拉達正常住宿、補給後再上船，
	// 以免下一個無旅店的海盜村把「尚有 8 MP」誤當成隱含前提。
	traceAdventureWalkToCty(t, g, 15)
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 27, true)
	if !g.inTown || g.curCty != 27 || sceneSection(g.cur) != 0 {
		t.Fatalf("未由正式航線抵達海盜村：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	redOrb, ok := g.pack.TreasureEvent("dq3:event.pirates_den_red_orb")
	if !ok {
		t.Fatal("缺 dq3:event.pirates_den_red_orb")
	}
	entranceX, entranceY := 26, 9
	objectIndex := g.cur.npcAt(entranceX, entranceY)
	if objectIndex < 0 || g.cur.npcs[objectIndex].ctrl&g.pack.NPCPushRule().CtrlMask == 0 {
		t.Fatalf("CTY27 密道入口缺可推物件：index=%d", objectIndex)
	}
	dcty, dsec, dx, dy, transitionOK := g.cur.tileTransition(entranceX, entranceY)
	if !transitionOK || dcty != 27 || dsec != redOrb.Treasure.Section || dx != 5 || dy != 9 {
		t.Fatalf("CTY27 密道 transition 不符：ok=%v -> CTY%d sec%d (%d,%d)",
			transitionOK, dcty, dsec, dx, dy)
	}
	entranceObject := &g.cur.npcs[objectIndex]
	traceWalkToNoPortal(t, g, entranceX, entranceY+1)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 1, DirEdge: -1}) // 向上推物件並踏入原格
	if entranceObject.x != entranceX || entranceObject.y != entranceY-1 ||
		!g.inTown || g.curCty != 27 || sceneSection(g.cur) != redOrb.Treasure.Section ||
		g.px != dx || g.py != dy {
		t.Fatalf("推開入口物件後未進紅寶珠區：object=(%d,%d) town=%v cty=%d sec=%d pos=(%d,%d)",
			entranceObject.x, entranceObject.y, g.inTown, g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	traceExaminePackTreasure(t, g, redOrb.Treasure)
	if !g.hasPartyItem(redOrb.Treasure.ItemRawID) || g.storyFlag(redOrb.Treasure.PresentFlag) {
		t.Fatalf("紅寶珠 transaction 錯：item=%v flag=%v",
			g.hasPartyItem(redOrb.Treasure.ItemRawID), g.storyFlag(redOrb.Treasure.PresentFlag))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存紅寶珠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建紅寶珠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 27 || sceneSection(g.cur) != redOrb.Treasure.Section ||
		!g.hasPartyItem(redOrb.Treasure.ItemRawID) || g.storyFlag(redOrb.Treasure.PresentFlag) {
		t.Fatalf("紅寶珠 save/load 錯：town=%v cty=%d sec=%d item=%v flag=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.hasPartyItem(redOrb.Treasure.ItemRawID),
			g.storyFlag(redOrb.Treasure.PresentFlag))
	}
	traceTownSectionTo(t, g, 27, 0)
	if g.cur.npcAt(entranceX, entranceY) >= 0 || g.cur.npcAt(entranceX, entranceY-1) >= 0 {
		t.Fatal("紅寶珠 present flag 清除後，密道物件不應在重載 CTY27 sec0 時重新出現")
	}

	// 紅寶珠 checkpoint → 用正式酒場選單寄放一名隊員、登錄商人、
	// 再指定商人入隊。不直接塞 roster/companions，也不注入建城旗標。
	// 魯拉是地表咒文；CTY27 的城內 mapFlags 不會開啟目的地清單，必須先
	// 由 section0 的正式邊界出口離城。
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, 0)
	traceAdventureWalkToCty(t, g, 0, false)
	traceTalkNPC(t, g, 2, 16)
	if !g.recruit.active || g.recruit.stage != rcMenu || len(g.companions) != 3 {
		t.Fatalf("商人登錄前露易達酒場狀態錯：active=%v stage=%d companions=%d",
			g.recruit.active, g.recruit.stage, len(g.companions))
	}
	send(InputState{DirHeld: -1, DirEdge: 0}) // 與同伴分離
	press(InputState{Confirm: true})
	leaveIndex := traceCompanionWithoutPhoenixOrb(g)
	if leaveIndex < 0 {
		t.Fatalf("三名同伴都持有祭壇寶珠，不能安全寄放以登錄建城商人：companions=%v",
			compsToSav(g.companions))
	}
	for i := 0; i < leaveIndex; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true}) // 由正式清單寄放第一名未持珠同伴
	press(InputState{Cancel: true})  // 清單→主選單
	press(InputState{Cancel: true})  // 關閉酒場
	if len(g.companions) != 2 || len(g.roster) != 1 {
		t.Fatalf("正式寄放隊員後清單錯：companions=%d roster=%d",
			len(g.companions), len(g.roster))
	}
	if traceMemberHasPhoenixOrb(g.roster[0]) {
		t.Fatalf("建城商人登錄前誤把持珠同伴寄放到 roster：%+v", compsToSav(g.roster))
	}

	traceWalkThroughPortal(t, g, 8, 14, 0, 2)
	traceTalkNPC(t, g, 2, 3)
	for i := 0; i < 6; i++ { // class raw 6：商人
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true})
	press(InputState{Toggle: true})
	press(InputState{Confirm: true}) // 名稱「0」
	for i := 0; i < 3; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	for i := 0; i < 8; i++ {
		send(InputState{DirHeld: -1, DirEdge: 3})
	}
	press(InputState{Confirm: true}) // cell43：進功能列
	for i := 0; i < 4; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true}) // 功能列「完成」
	press(InputState{Confirm: true}) // 男性
	if len(g.roster) != 2 || g.roster[1].Class != 6 {
		t.Fatalf("正式登錄商人失敗：roster=%d lastClass=%d",
			len(g.roster), g.roster[len(g.roster)-1].Class)
	}
	merchantName := append([]int(nil), g.roster[1].Name...)
	traceWalkThroughPortal(t, g, 8, 2, 0, 0)
	traceTalkNPC(t, g, 2, 16)
	press(InputState{Confirm: true}) // 找同伴參加
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true}) // roster[1] 商人
	press(InputState{Cancel: true})
	press(InputState{Cancel: true})
	if len(g.companions) != 3 || g.companions[2].Class != 6 || len(g.roster) != 1 {
		t.Fatalf("正式招募商人後隊伍錯：companions=%d lastClass=%d roster=%d",
			len(g.companions), g.companions[len(g.companions)-1].Class, len(g.roster))
	}

	// 魯拉只回到本 trace 早先已正常訪問的 CTY38，並由 production
	// 重定位船隻。先在阿里阿罕處理可能的陣亡與 MP 補給，再由正式
	// section0 出口離城施法；魯拉在城內及 MP 不足時都必須 fail closed。
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, 38)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 58, true)
	founderEvent, ok := g.pack.SettlementFounderEvent("dq3:event.merchant_settlement_founder")
	if !ok {
		t.Fatal("缺 dq3:event.merchant_settlement_founder")
	}
	if g.curCty != founderEvent.NPC.CTYRaw || sceneSection(g.cur) != founderEvent.NPC.Section {
		t.Fatalf("未由正式入口進入商人城初階段：cty=%d sec=%d",
			g.curCty, sceneSection(g.cur))
	}
	traceTalkNPC(t, g, founderEvent.NPC.Tile.X, founderEvent.NPC.Tile.Y)
	traceCloseDialogue(t, g) // 自我介紹、商人候選，停在第一次選擇
	if g.settlementFounderStage != settlementFounderFirstChoice {
		t.Fatalf("商人城第一次選擇階段錯：%d", g.settlementFounderStage)
	}
	press(InputState{Confirm: true})
	traceCloseDialogue(t, g) // 最後確認文字，停在第二次選擇
	if g.settlementFounderStage != settlementFounderFinalChoice {
		t.Fatalf("商人城第二次選擇階段錯：%d", g.settlementFounderStage)
	}
	press(InputState{Confirm: true})
	traceCloseDialogue(t, g)
	if !g.storyFlag(founderEvent.CompletionFlagRaw) || len(g.companions) != 2 ||
		g.settlementFounder == nil || g.settlementFounder.Class != 6 ||
		!reflect.DeepEqual(g.settlementFounder.Name, merchantName) {
		t.Fatalf("商人交付 transaction 錯：flag=%v companions=%d founder=%+v",
			g.storyFlag(founderEvent.CompletionFlagRaw), len(g.companions), g.settlementFounder)
	}
	for _, m := range g.roster {
		if m.Class == 6 && reflect.DeepEqual(m.Name, merchantName) {
			t.Fatal("已交付的建城商人不得回到酒場名冊")
		}
	}

	wantFounder := compsToSav([]*Member{g.settlementFounder})[0]
	wantStorage := append([]int(nil), g.sharedStorage...)
	if err := g.Save(); err != nil {
		t.Fatalf("保存商人城 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建商人城讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != founderEvent.NPC.CTYRaw ||
		!g.storyFlag(founderEvent.CompletionFlagRaw) || g.settlementFounder == nil ||
		!reflect.DeepEqual(compsToSav([]*Member{g.settlementFounder})[0], wantFounder) ||
		!reflect.DeepEqual(g.sharedStorage, wantStorage) {
		t.Fatalf("商人城 save/load 錯：town=%v cty=%d flag=%v founder=%+v storage=%v",
			g.inTown, g.curCty, g.storyFlag(founderEvent.CompletionFlagRaw),
			g.settlementFounder, g.sharedStorage)
	}

	// 建城商人交付後，離開再由同一世界座標重進，production selector 應依
	// founder flag 自然載入 CTY60；不得直接指定城鎮階段。
	traceExitTownBoundary(t, g, true)
	traceReenterWorldEntrance(t, g, 60)
	if !g.inTown || g.curCty != 60 {
		t.Fatalf("商人交付後重進未自然選到 CTY60：town=%v cty=%d", g.inTown, g.curCty)
	}

	// 建城交出商人後隊伍已有空位；先前為登錄商人而寄放的原同伴仍以
	// 原版個人物品保管語意留在 roster。必須立刻經露易達正式選單復隊，
	// 否則其暗黑神燈不在 active party，後續沙曼歐莎夜間事件無法進行。
	// 不直接搬 roster、道具、MP 或座標。
	if len(g.roster) != 1 || len(g.companions) != 2 {
		t.Fatalf("建城後立即復隊前清單錯：companions=%d roster=%d",
			len(g.companions), len(g.roster))
	}
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 0)
	traceAdventureWalkToCty(t, g, 0, false)
	traceTalkNPC(t, g, 2, 16)
	if !g.recruit.active || g.recruit.stage != rcMenu {
		t.Fatalf("建城後立即復隊未開啟露易達主選單：active=%v stage=%d",
			g.recruit.active, g.recruit.stage)
	}
	press(InputState{Confirm: true}) // 找同伴參加
	press(InputState{Confirm: true}) // 唯一的 roster[0]
	press(InputState{Cancel: true})  // 清單→主選單
	press(InputState{Cancel: true})  // 關閉酒場
	if len(g.companions) != 3 || len(g.roster) != 0 ||
		!g.hasPartyItem(darkLampEffect.ItemRawID) {
		t.Fatalf("建城後立即正式復隊錯：companions=%d roster=%d darkLamp=%v",
			len(g.companions), len(g.roster), g.hasPartyItem(darkLampEffect.ItemRawID))
	}

	// 原版攻略路線：由商人城返回船隻，航行至雪島 CTY41，經右側旅人之門
	// 抵達 CTY42，再由 transition record 的原始世界座標徒步前往沙曼歐莎。
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 38)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 41, true)
	traceTownSectionTo(t, g, 42, 0)
	traceTownSectionTo(t, g, -1, -1)
	if g.px != 213 || g.py != 123 {
		t.Fatalf("CTY42 旅人之門原始落點錯：(%d,%d)，want (213,123)", g.px, g.py)
	}
	// 地表西行會先踏入 CTY43；原始 transition graph 指定由 sec0 (5,31)
	// 跨 CTY 到沙曼歐莎 CTY44 sec0，不能把 CTY43 入口當成可穿越地形。
	traceAdventureWalkToCty(t, g, 43, false)
	// 建城與航行後的隊伍可能帶傷或中毒。先用 CTY43 正式旅店恢復
	// 存活者，再由已造訪的 CTY2 教會逐員復活／解毒；否則 CTY44
	// 北出口途中會正確觸發原版全滅 modal，不能誤報成 portal 壞掉。
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 2)
	traceAdventureWalkToCty(t, g, 2, false)
	traceReviveDeadAtChurch(t, g)
	traceCurePartyPoisonAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 43)
	traceAdventureWalkToCty(t, g, 43, false)
	traceTownSectionTo(t, g, ctySamanosa, 0)
	// 沙曼歐莎本城 sec0 沒有直接 world boundary；離場需反向穿回 CTY43
	// 關卡，再由關卡邊界回地表。
	traceTownSectionTo(t, g, 43, 0)
	traceExitTownBoundary(t, g, true)
	traceAdventureWalkToCty(t, g, 24, false)

	mirrorTreasure, ok := g.pack.TreasureEvent("dq3:event.shamanoasa_mod_change_mirror")
	if !ok {
		t.Fatal("缺 dq3:event.shamanoasa_mod_change_mirror")
	}
	traceTownSectionTo(t, g, mirrorTreasure.Treasure.CTYRaw,
		mirrorTreasure.Treasure.Section)
	traceExaminePackTreasure(t, g, mirrorTreasure.Treasure)
	if !g.hasPartyItem(mirrorTreasure.Treasure.ItemRawID) ||
		g.storyFlag(mirrorTreasure.Treasure.PresentFlag) {
		t.Fatalf("拉之鏡 transaction 錯：item=%v flag=%v",
			g.hasPartyItem(mirrorTreasure.Treasure.ItemRawID),
			g.storyFlag(mirrorTreasure.Treasure.PresentFlag))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存拉之鏡 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建拉之鏡讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != mirrorTreasure.Treasure.CTYRaw ||
		!g.hasPartyItem(mirrorTreasure.Treasure.ItemRawID) ||
		g.storyFlag(mirrorTreasure.Treasure.PresentFlag) {
		t.Fatalf("拉之鏡 save/load 錯：town=%v cty=%d item=%v flag=%v",
			g.inTown, g.curCty, g.hasPartyItem(mirrorTreasure.Treasure.ItemRawID),
			g.storyFlag(mirrorTreasure.Treasure.PresentFlag))
	}

	traceTownSectionTo(t, g, -1, -1)
	// CTY43 的 raw index0 旅店有日／夜 facility NPC consumer；先由正常入口
	// 住宿恢復存活勇者的魯拉 MP。孤立的 index2 教會 block 沒有 NPC consumer，
	// 所以再以正式魯拉回已造訪的 CTY2 復活並補滿全隊，最後魯拉回 CTY43。
	traceAdventureWalkToCty(t, g, 43, false)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 2)
	traceAdventureWalkToCty(t, g, 2, false)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 43)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceUseInventoryItem(t, g, darkLampEffect.ItemRawID)
	if !g.isNight() {
		t.Fatalf("前往假王事件前黑暗燈未切到夜晚：phase=%d", g.dnPhase)
	}
	traceAdventureWalkToCty(t, g, 43, false)
	traceTownSectionTo(t, g, ctySamanosa, 0)
	traceTownSectionTo(t, g, ctySamanosa, mirrorSection, mirrorX, mirrorY)
	traceWalkTo(t, g, mirrorX, mirrorY)
	ensurePartyInventorySpace("怪力魔勝利獎勵", 1)
	traceUseInventoryItem(t, g, mirrorTreasure.Treasure.ItemRawID)
	if g.mirrorStage != 1 || !g.dlg.open || g.storyFlag(0x42) ||
		!g.storyFlag(0x10) {
		t.Fatalf("正式使用拉之鏡未揭露假王：stage=%d dlg=%v f42=%v f10=%v night=%v cty=%d sec=%d pos=(%d,%d) item=%v",
			g.mirrorStage, g.dlg.open, g.storyFlag(0x42), g.storyFlag(0x10),
			g.isNight(), g.curCty, sceneSection(g.cur), g.px, g.py,
			g.hasPartyItem(mirrorTreasure.Treasure.ItemRawID))
	}
	traceCloseDialogue(t, g)
	preBossParty := make([][3]int, len(g.companions))
	for i, member := range g.companions {
		preBossParty[i] = [3]int{member.Level(), member.CurHP, member.MaxHP()}
	}
	_, preBossHeroMax, _, _, _ := g.heroStats()
	t.Logf("怪力魔戰前資源：hero=%d/%d MP=%d companions(lv,hp,max)=%v",
		g.heroHP, preBossHeroMax, g.heroMP, preBossParty)
	traceCloseDialogue(t, g)
	if g.mirrorStage != 3 || !g.battle.active || g.battle.monID != monsterBossTroll {
		t.Fatalf("拉之鏡對話後未進怪力魔戰：stage=%d active=%v mon=%d",
			g.mirrorStage, g.battle.active, g.battle.monID)
	}
	traceResolveBattle(t, g, false)
	traceCloseDialogue(t, g)
	if g.battle.active || g.mirrorStage != 0 || !g.hasPartyItem(itemModChangeStaff) ||
		g.storyFlag(0x42) || !g.storyFlag(0x22) || g.dnPhase != 0 {
		t.Fatalf("怪力魔勝利 transaction 錯：active=%v stage=%d staff=%v f42=%v f22=%v phase=%d",
			g.battle.active, g.mirrorStage, g.hasItem(itemModChangeStaff),
			g.storyFlag(0x42), g.storyFlag(0x22), g.dnPhase)
	}

	exchanges := g.pack.ChoiceItemExchangeEvents()
	if len(exchanges) != 1 {
		t.Fatalf("choice item exchange event count=%d, want 1", len(exchanges))
	}
	staffExchange := exchanges[0]
	traceTownSectionTo(t, g, 44, 0)
	traceTownSectionTo(t, g, 43, 0)
	traceExitTownBoundary(t, g, true)
	traceAdventureWalkToCty(t, g, 43, false) // 從城堡側出口重進可達旅店的城鎮區塊。
	traceTalkFacility(t, g, facInn)          // 假王戰後以正式旅店補足魯拉 MP。
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 2)
	traceAdventureWalkToCty(t, g, 2, false)
	traceReviveDeadAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	traceRuraToCty(t, g, 39) // 正式重定位船到可航向格陵蘭的已造訪港口。
	traceBoardAndSailShip(t, g)
	// 這段從旅店補滿後出發，且下一個安全補給點在遠洋彼端；明確允許
	// 路線採用特黑洛斯，而不是把這個策略外溢到所有需保留魯拉 MP 的航段。
	traceAdventureTravelToCty(t, g, staffExchange.NPC.CTYRaw, true, true, true)
	traceTownSectionTo(t, g, staffExchange.NPC.CTYRaw, staffExchange.NPC.Section)
	traceTalkNPC(t, g, staffExchange.NPC.Tile.X, staffExchange.NPC.Tile.Y)
	if g.choiceItemExchangeStage != choiceItemExchangeOffer || !g.dlg.open {
		t.Fatalf("正式持杖交談未進交換提議：stage=%d dlg=%v",
			g.choiceItemExchangeStage, g.dlg.open)
	}
	traceCloseDialogue(t, g)
	if g.choiceItemExchangeStage != choiceItemExchangeOfferChoice {
		t.Fatalf("交換提議關閉後未進 Yes/No：stage=%d", g.choiceItemExchangeStage)
	}
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if g.choiceItemExchangeStage != choiceItemExchangeReject || !g.dlg.open ||
		!g.hasPartyItem(staffExchange.RequiredItemRawID) ||
		!g.storyFlag(staffExchange.AvailableFlagRaw) {
		t.Fatalf("正式拒絕交換修改狀態：stage=%d dlg=%v staff=%v flag=%v",
			g.choiceItemExchangeStage, g.dlg.open,
			g.hasPartyItem(staffExchange.RequiredItemRawID),
			g.storyFlag(staffExchange.AvailableFlagRaw))
	}
	traceCloseDialogue(t, g)
	traceTalkNPC(t, g, staffExchange.NPC.Tile.X, staffExchange.NPC.Tile.Y)
	traceCloseDialogue(t, g)
	press(InputState{Confirm: true})
	ghostPos, ghostActive := g.trackedWorldPositions[staffExchange.ActivateWorldObject.ObjectID]
	if g.choiceItemExchangeStage != choiceItemExchangeSuccess || !g.dlg.open ||
		g.hasPartyItem(staffExchange.RequiredItemRawID) ||
		!g.hasPartyItem(staffExchange.GrantedItemRawID) ||
		g.storyFlag(staffExchange.AvailableFlagRaw) ||
		g.worldState&uint16(staffExchange.SetWorldStateMaskRaw) == 0 || !ghostActive ||
		ghostPos != (trackedWorldPosition{X: staffExchange.ActivateWorldObject.Position.X,
			Y: staffExchange.ActivateWorldObject.Position.Y, Layer: staffExchange.ActivateWorldObject.Position.Layer}) {
		t.Fatalf("正式接受交換 transaction 錯：stage=%d dlg=%v staff=%v bones=%v flag=%v world=%#x player=(%d,%d) object=%+v",
			g.choiceItemExchangeStage, g.dlg.open,
			g.hasPartyItem(staffExchange.RequiredItemRawID),
			g.hasPartyItem(staffExchange.GrantedItemRawID),
			g.storyFlag(staffExchange.AvailableFlagRaw), g.worldState, g.overPx, g.overPy, ghostPos)
	}
	traceCloseDialogue(t, g)
	traceCloseDialogue(t, g)
	if g.choiceItemExchangeStage != choiceItemExchangeIdle {
		t.Fatalf("交換後話未閉合：stage=%d", g.choiceItemExchangeStage)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存船員骨頭 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建船員骨頭讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	ghostPos, ghostActive = g.trackedWorldPositions[staffExchange.ActivateWorldObject.ObjectID]
	if !g.inTown || g.curCty != staffExchange.NPC.CTYRaw ||
		!g.hasPartyItem(staffExchange.GrantedItemRawID) ||
		g.hasPartyItem(staffExchange.RequiredItemRawID) ||
		g.storyFlag(staffExchange.AvailableFlagRaw) ||
		g.worldState&uint16(staffExchange.SetWorldStateMaskRaw) == 0 || !ghostActive ||
		ghostPos != (trackedWorldPosition{X: staffExchange.ActivateWorldObject.Position.X,
			Y: staffExchange.ActivateWorldObject.Position.Y, Layer: staffExchange.ActivateWorldObject.Position.Layer}) {
		t.Fatalf("船員骨頭 save/load 錯：town=%v cty=%d bones=%v staff=%v flag=%v world=%#x player=(%d,%d) object=%+v",
			g.inTown, g.curCty, g.hasPartyItem(staffExchange.GrantedItemRawID),
			g.hasPartyItem(staffExchange.RequiredItemRawID),
			g.storyFlag(staffExchange.AvailableFlagRaw), g.worldState, g.overPx, g.overPy, ghostPos)
	}

	// 船員骨頭由正式道具選單連續播放 rec740、X 距離記錄、Y 距離記錄；
	// 接著從格陵蘭正常出城、徒步登船、逐格航行至動態座標進入 CTY36。
	traceUseInventoryItem(t, g, staffExchange.GrantedItemRawID)
	if !g.dlg.open || g.trackedWorldLocatorStage != 1 {
		t.Fatalf("船員骨頭前導記錄未開啟：dlg=%v stage=%d",
			g.dlg.open, g.trackedWorldLocatorStage)
	}
	traceCloseDialogue(t, g) // helper 會穿過三個依序開啟的原版記錄
	if g.trackedWorldLocatorStage != 0 {
		t.Fatalf("船員骨頭三段記錄未閉合：stage=%d", g.trackedWorldLocatorStage)
	}
	if g.panel != panelNone {
		press(InputState{Cancel: true})
	}
	traceExitTownBoundary(t, g, true)
	// 格陵蘭的原始 facility 表沒有旅店；交換後勇者只剩足夠施放一次
	// 魯拉的 MP。先以正式魯拉回已造訪、確有可達旅店且與幽靈船同一
	// 可航水域的 CTY38 補滿，再從該港口出航尋找幽靈船；避免把後續
	// 數段航路的 MP 耗盡誤當成尼羅肯特／蓋亞事件的功能缺陷。
	traceRuraToCty(t, g, 38)
	if g.isNight() {
		// 原版 phase3 仍使用夜間 NPC 表；在城外以正式步行等到日表，
		// 不把舊 remake 的「黎明=白天」錯誤固化成 trace 前提。
		traceWaitForDayNearCty(t, g, 38)
	}
	traceAdventureWalkToCty(t, g, 38, false)
	traceTalkFacility(t, g, facItem)
	press(InputState{Cancel: true})
	if free := dropSpareEquipmentForPartyInventorySlots(32); free < 10 {
		t.Fatalf("幽靈船前正式丟完備用裝備後容量仍不足：free=%d minimum=10", free)
	}
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(2)
	topUpAntidote(3)
	topUpHerb(16)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	if !g.shipAboard {
		traceBoardAndSailShip(t, g)
	}
	ghostObject, ok := g.pack.TrackedWorldObject(staffExchange.ActivateWorldObject.ObjectID)
	if !ok {
		t.Fatal("讀檔後缺幽靈船 world object")
	}
	traceSailIntoTrackedWorldObject(t, g, ghostObject)
	traceWalkThroughPortal(t, g, 18, 37, ghostObject.EntranceCTYRaw, 1, 18, 38)
	loveMemory, ok := g.pack.TreasureEvent("dq3:event.ghost_ship_loves_memory")
	if !ok {
		t.Fatal("缺幽靈船愛的回憶 treasure event")
	}
	traceExaminePackTreasure(t, g, loveMemory.Treasure)
	if !g.hasItem(loveMemory.Treasure.ItemRawID) ||
		g.storyFlag(loveMemory.Treasure.PresentFlag) {
		t.Fatalf("正式幽靈船寶箱錯：memory=%v present=%v",
			g.hasItem(loveMemory.Treasure.ItemRawID),
			g.storyFlag(loveMemory.Treasure.PresentFlag))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存愛的回憶 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建愛的回憶讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != ghostObject.EntranceCTYRaw || sceneSection(g.cur) != 1 ||
		!g.hasItem(loveMemory.Treasure.ItemRawID) ||
		g.storyFlag(loveMemory.Treasure.PresentFlag) {
		t.Fatalf("愛的回憶 save/load 錯：town=%v cty=%d sec=%d memory=%v present=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.hasItem(loveMemory.Treasure.ItemRawID),
			g.storyFlag(loveMemory.Treasure.PresentFlag))
	}

	// 從愛的回憶 checkpoint 只走原始 CTY36 轉場離船，重新登船後航向
	// `(76,54)`。座標事件必須自動播放 records597/598、保留道具並 clear flag0x35。
	traceTransitionToSection(t, g, ghostObject.EntranceCTYRaw, 0)
	traceTransitionToOverworld(t, g)
	if !g.shipAboard {
		traceBoardAndSailShip(t, g)
	}
	oliviaGate, ok := g.pack.CoordinateItemGateAt(76, 54, 0)
	if !ok || !g.storyFlag(oliviaGate.ActiveStoryFlagRaw) {
		t.Fatalf("奧莉薇亞海岬前置錯：event=%v flag35=%v", ok,
			ok && g.storyFlag(oliviaGate.ActiveStoryFlagRaw))
	}
	traceSailToWorldCoordinate(t, g, oliviaGate.Coordinate.X, oliviaGate.Coordinate.Y)
	if !g.dlg.open || g.coordinateItemGateStage != coordinateGateSuccessApproach {
		t.Fatalf("海岬正式入口未顯示 record597：dlg=%v stage=%d pos=(%d,%d)",
			g.dlg.open, g.coordinateItemGateStage, g.px, g.py)
	}
	traceCloseDialogue(t, g)
	if g.storyFlag(oliviaGate.ClearStoryFlagRaw) || !g.hasItem(oliviaGate.RequiredItemRawID) ||
		g.coordinateForcedSteps != 0 {
		t.Fatalf("海岬成功交易錯：flag35=%v memory=%v forced=%d",
			g.storyFlag(oliviaGate.ClearStoryFlagRaw), g.hasItem(oliviaGate.RequiredItemRawID),
			g.coordinateForcedSteps)
	}

	// 舊 defeat 近似會在這段長航路全滅時暗中補滿整隊，掩蓋三人皆
	// 中毒的正式玩家狀態。使用 CTY38 預先買好的驅毒草，透過正式
	// rec421／選人 modal 在再次移動前逐員解毒；若已有死者才走教會。
	for target := 0; target <= len(g.companions); target++ {
		poisoned := target == 0 && g.heroHP > 0 && g.heroConditions&conditionPoison != 0
		if target > 0 {
			poisoned = g.companions[target-1].CurHP > 0 &&
				g.companions[target-1].Conditions&conditionPoison != 0
		}
		if !poisoned {
			continue
		}
		traceUseInventoryItem(t, g, itemuse.ItemAntidote, target)
		traceCloseDialogue(t, g)
	}
	if g.heroConditions&conditionPoison != 0 {
		t.Fatal("奧莉薇亞事件後正式解毒未清除勇者 poison")
	}
	for _, member := range g.companions {
		if member.CurHP > 0 && member.Conditions&conditionPoison != 0 {
			t.Fatal("奧莉薇亞事件後正式解毒未清除同伴 poison")
		}
	}
	traceHealLivingParty(t, g)
	if tracePartyLivingNeedsHealing(g) {
		t.Fatal("正式野外補血咒文與剩餘藥草仍不足以治療海岬後存活隊員")
	}
	if !g.shipAboard {
		traceBoardAndSailShip(t, g)
	}
	if g.repel <= 8 {
		if !g.hasItem(itemuse.ItemHolyWater) {
			t.Fatal("CTY55 返航前缺少在 CTY38 正式購買的聖水")
		}
		traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	}
	if !g.shipAboard {
		t.Fatal("CTY55 返航前正式使用聖水後不應離船")
	}

	gaiaTreasure, ok := g.pack.TreasureEvent("dq3:event.shrine_jail_gaia_sword")
	if !ok {
		t.Fatal("缺 CTY55 蓋亞之劍 treasure event")
	}
	traceSailToTown(t, g, gaiaTreasure.Treasure.CTYRaw, true)
	traceExaminePackTreasure(t, g, gaiaTreasure.Treasure)
	if !g.hasItem(gaiaTreasure.Treasure.ItemRawID) ||
		g.storyFlag(gaiaTreasure.Treasure.PresentFlag) {
		t.Fatalf("正式取得蓋亞之劍錯：item=%v present=%v",
			g.hasItem(gaiaTreasure.Treasure.ItemRawID),
			g.storyFlag(gaiaTreasure.Treasure.PresentFlag))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存蓋亞之劍 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建蓋亞之劍讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != gaiaTreasure.Treasure.CTYRaw ||
		!g.hasItem(gaiaTreasure.Treasure.ItemRawID) ||
		g.storyFlag(oliviaGate.ClearStoryFlagRaw) ||
		g.storyFlag(gaiaTreasure.Treasure.PresentFlag) {
		t.Fatalf("蓋亞之劍 save/load 錯：town=%v cty=%d sword=%v flag35=%v present=%v",
			g.inTown, g.curCty, g.hasItem(gaiaTreasure.Treasure.ItemRawID),
			g.storyFlag(oliviaGate.ClearStoryFlagRaw),
			g.storyFlag(gaiaTreasure.Treasure.PresentFlag))
	}

	// 幽靈船→海岬→CTY55 已是很長的資源鏈；在最後一個合法城市 checkpoint
	// 以正式魯拉回已造訪 CTY38，補足聖水並住宿，再由重定位後的船直航火山。
	// 這讓火山／尼羅肯特路段不依賴剛好少抽一場 encounter，也不直接寫 MP、
	// repel 或位置。蓋亞之劍已跨 save/load，沒有事件理由再繞回 CTY55；船位與
	// 後續航行都經 production handler。
	traceRuraToCty(t, g, 38)
	if g.isNight() {
		traceWaitForDayNearCty(t, g, 38)
	}
	traceAdventureWalkToCty(t, g, 38, false)
	traceTalkFacility(t, g, facItem)
	topUpHolyWater(8)
	press(InputState{Cancel: true})
	traceTalkFacility(t, g, facInn)
	traceExitTownBoundary(t, g, true)
	needsRevive := false
	for _, member := range g.companions {
		needsRevive = needsRevive || member.CurHP <= 0
	}
	if needsRevive {
		// CTY38 沒有已閉合的教會 consumer；住宿恢復勇者魯拉 MP 後，
		// 走正式 CTY2 教會／旅店復活並補滿，再魯拉回同一港口。
		traceRuraToCty(t, g, 2)
		traceAdventureWalkToCty(t, g, 2)
		traceReviveDeadAtChurch(t, g)
		traceTalkFacility(t, g, facInn)
		traceExitTownBoundary(t, g, true)
		traceRuraToCty(t, g, 38)
	}
	traceBoardAndSailShip(t, g)

	// CTY55 蓋亞之劍 checkpoint → CTY38 正式補給 → 蓋亞之劍原版使用格
	// → 火山 patch → 尼羅肯特洞窟
	// → CTY64。原版不是從火山直接航到 CTY64：patch 開出的陸路先到
	// CTY56 南洞口，穿過 section 1/2/3 的正式 transition，再由 sec0 北洞口
	// 出現在 (43,138)，最後步行進 CTY64。這段只送正式命令窗、道具清單、
	// 方向鍵與 NPC 對話輸入；patch 不是測試注入，且 item 0x0f 在原版 handler
	// 不消耗。
	gaiaUse, ok := g.pack.ItemUseEffectByRawID(0x0f)
	if !ok || gaiaUse.UseTile == nil || gaiaUse.MapPatch == nil || gaiaUse.DirectMapPatch == nil {
		t.Fatalf("蓋亞之劍 item-use pack effect 缺失：%+v", gaiaUse)
	}
	// 蓋亞使用格位於陸地內側；原始水路可到達 (63,107) 岸格，最後由
	// tryMove 正式下船，再以正常陸路走到 (65,109)，不能把內陸段當船路。
	// 火山後的尼羅肯特出口必須再以魯拉返回港口；航程中可用正式
	// 特黑洛斯避開高階遭遇，但要為同一位勇者保留 8 MP。
	traceSailToWorldCoordinate(t, g, 63, 107, true)
	traceWalkToWorldCoordinate(t, g, gaiaUse.UseTile.X, gaiaUse.UseTile.Y)
	traceUseInventoryItem(t, g, gaiaUse.ItemRawID)
	if g.worldState&uint16(gaiaUse.SetWorldStateMask) == 0 ||
		!g.hasItem(gaiaUse.ItemRawID) || g.panel != panelNone {
		t.Fatalf("正式使用蓋亞之劍 transaction 錯：state=%#x item=%v panel=%d",
			g.worldState, g.hasItem(gaiaUse.ItemRawID), g.panel)
	}
	for y := 0; y < gaiaUse.DirectMapPatch.Height; y++ {
		for x := 0; x < gaiaUse.DirectMapPatch.Width; x++ {
			want := gaiaUse.DirectMapPatch.TilesRaw[y*gaiaUse.DirectMapPatch.Width+x]
			if got := g.over.tileIdx(gaiaUse.DirectMapPatch.Origin.X+x, gaiaUse.DirectMapPatch.Origin.Y+y); got != want {
				t.Fatalf("正式蓋亞火山 patch (%d,%d)=%#x，want %#x",
					gaiaUse.DirectMapPatch.Origin.X+x, gaiaUse.DirectMapPatch.Origin.Y+y, got, want)
			}
		}
	}
	if g.dlg.open || g.flags[0x32] {
		t.Fatalf("蓋亞使用不應開對話或寫舊 remake flag：dlg=%v flag32=%v", g.dlg.open, g.flags[0x32])
	}

	// 共用 portal helper 會保留最後一瓶聖水，但本固定路段的下一個資源 gate
	// 是 CTY64 出口的魯拉，不是物品。進洞前經正式道具選單用掉最後一瓶，
	// 讓驅敵效果涵蓋尼羅肯特；不直接寫 repel counter。
	if g.hasItem(itemuse.ItemHolyWater) {
		traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	}
	ruraReserve := spell.MPCost(fieldRura)
	traceGaiaSwordToNirokenta(t, g)
	if g.heroMP < ruraReserve {
		t.Fatalf("尼羅肯特正式路線未保留魯拉 MP：mp=%d reserve=%d repel=%d holyWater=%d",
			g.heroMP, ruraReserve, g.repel, g.countItem(itemuse.ItemHolyWater))
	}
	var silver *gamepack.NPCItemRewardEvent
	for _, event := range g.pack.NPCItemRewardEvents() {
		if event.ID == "dq3:event.nirokenta_silver_orb" {
			e := event
			silver = &e
			break
		}
	}
	if silver == nil {
		t.Fatal("缺 CTY64 尼羅肯特銀寶珠 NPC pack event")
	}
	partyHasSilver := func(game *Game) bool {
		if game.hasItem(silver.GrantedItemRaw) {
			return true
		}
		for _, member := range game.companions {
			if containsInt(member.Inventory, silver.GrantedItemRaw) {
				return true
			}
		}
		return false
	}
	if !g.storyFlag(silver.PresentFlagRaw) {
		t.Fatalf("進入尼羅肯特前銀寶珠 present flag %#x 未成立", silver.PresentFlagRaw)
	}
	// 野外回復咒文會保留原先預備的藥草；handler49 的原版共用 writer
	// 在全隊八格皆滿時必須失敗且不清 flag。經正式 rec421 優先丟藥草，
	// 否則只丟 ITEM.DAT 證明為裝備且未穿戴的備品，為銀寶珠預留一格；
	// 不放寬 engine 容量或直接改 inventory。
	if free := dropSpareEquipmentForPartyInventorySlots(1); free < 1 {
		t.Fatalf("尼羅肯特銀寶珠 handler49：需要1格，正式丟棄藥草／未裝備備品後仍只有%d格", free)
	}
	traceTalkNPC(t, g, silver.NPC.Tile.X, silver.NPC.Tile.Y)
	if !g.dlg.open {
		t.Fatal("尼羅肯特銀寶珠正式對話未開啟")
	}
	traceCloseDialogue(t, g)
	if !partyHasSilver(g) || g.storyFlag(silver.PresentFlagRaw) {
		t.Fatalf("正式取得銀寶珠 transaction 錯：item=%v present=%v",
			partyHasSilver(g), g.storyFlag(silver.PresentFlagRaw))
	}
	silverCheckpointMP := g.heroMP
	if silverCheckpointMP < ruraReserve {
		t.Fatalf("銀寶珠存檔前未保留魯拉 MP：mp=%d reserve=%d", silverCheckpointMP, ruraReserve)
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存銀寶珠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建銀寶珠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if g.heroMP != silverCheckpointMP {
		t.Fatalf("銀寶珠正式載入改變勇者 MP：before=%d after=%d max=%d stats=%v",
			silverCheckpointMP, g.heroMP, g.heroMaxMP(), g.heroStat)
	}
	if !g.inTown || g.curCty != 64 || sceneSection(g.cur) != 0 ||
		!partyHasSilver(g) || g.storyFlag(silver.PresentFlagRaw) {
		t.Fatalf("銀寶珠 save/load round-trip 錯：town=%v cty=%d sec=%d item=%v present=%v",
			g.inTown, g.curCty, sceneSection(g.cur), partyHasSilver(g),
			g.storyFlag(silver.PresentFlagRaw))
	}
	// 銀寶珠 checkpoint → 以正式魯拉／登船返回商人城。蓋亞之劍已將
	// 世界入口的 ordered gate 推進至 CTY83；不得直接指定 CTY 或把黃寶珠
	// 寫入隊伍。先與牢中建城者完成原版兩段提示，再調查椅後事件物件。
	// CTY64 sec0 的原始出口不是一般 town boundary；地圖記錄的門
	// `(14,23) -> section 0xff @ (51,133)` 會回到世界層。
	traceWalkThroughPortal(t, g, 14, 23, -1, 0)
	// 這段航路的正式戰鬥策略保留魯拉 MP；從 CTY64 世界出口先以
	// 魯拉返回已造訪港口，讓原版 handler 正常重定位停泊船。
	traceRuraToCty(t, g, 38)
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, 83, true)
	yellowOrb, ok := g.pack.TreasureEvent("dq3:event.merchant_town_yellow_orb")
	if !ok {
		t.Fatal("缺 CTY83 商人城黃寶珠 treasure event")
	}
	var followup *gamepack.SettlementFounderFollowupEvent
	for _, event := range g.pack.SettlementFounderFollowups() {
		if event.ID == "dq3:event.merchant_settlement_imprisoned_with_orb" {
			e := event
			followup = &e
			break
		}
	}
	if followup == nil {
		t.Fatal("缺 CTY83 建城者黃寶珠前置對話 event")
	}
	if g.curCty != yellowOrb.Treasure.CTYRaw || sceneSection(g.cur) != yellowOrb.Treasure.Section ||
		!g.storyFlag(yellowOrb.Treasure.PresentFlag) {
		t.Fatalf("蓋亞後 CTY83 黃寶珠入口錯：cty=%d sec=%d present=%v",
			g.curCty, sceneSection(g.cur), g.storyFlag(yellowOrb.Treasure.PresentFlag))
	}
	if g.cur.npcAt(followup.NPC.Tile.X, followup.NPC.Tile.Y) < 0 {
		t.Fatalf("CTY83 缺少建城者後續 NPC：(%d,%d)", followup.NPC.Tile.X, followup.NPC.Tile.Y)
	}
	traceTalkNPC(t, g, followup.NPC.Tile.X, followup.NPC.Tile.Y)
	traceCloseDialogue(t, g)
	if len(g.settlementFounderFollowup) != 0 || g.dlg.open {
		t.Fatalf("建城者黃寶珠提示未閉合：pending=%v dlg=%v",
			g.settlementFounderFollowup, g.dlg.open)
	}
	// 場外回復咒文會保留原先預備的藥草；從銀寶珠返航的戰鬥掉落也可能
	// 再次填滿剛釋出的格子。CTY83 type1 事件沿用原版全隊八格 writer，
	// 滿格時必須保留 present flag。經正式 rec421 優先丟藥草，否則只丟
	// ITEM.DAT 證明為裝備且未穿戴的備品；不放寬容量或直接改 inventory。
	if free := dropSpareEquipmentForPartyInventorySlots(1); free < 1 {
		t.Fatalf("CTY83 黃寶珠：需要1格，正式丟棄藥草／未裝備備品後仍只有%d格", free)
	}
	traceExaminePackTreasure(t, g, yellowOrb.Treasure)
	if !g.hasPartyItem(yellowOrb.Treasure.ItemRawID) || g.storyFlag(yellowOrb.Treasure.PresentFlag) {
		t.Fatalf("黃寶珠 transaction 錯：item=%v present=%v",
			g.hasPartyItem(yellowOrb.Treasure.ItemRawID), g.storyFlag(yellowOrb.Treasure.PresentFlag))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存黃寶珠 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建黃寶珠讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != yellowOrb.Treasure.CTYRaw || sceneSection(g.cur) != yellowOrb.Treasure.Section ||
		!g.hasPartyItem(yellowOrb.Treasure.ItemRawID) || g.storyFlag(yellowOrb.Treasure.PresentFlag) {
		t.Fatalf("黃寶珠 save/load round-trip 錯：town=%v cty=%d sec=%d item=%v present=%v",
			g.inTown, g.curCty, sceneSection(g.cur), g.hasPartyItem(yellowOrb.Treasure.ItemRawID),
			g.storyFlag(yellowOrb.Treasure.PresentFlag))
	}
	traceTalkNPC(t, g, followup.NPC.Tile.X, followup.NPC.Tile.Y)
	traceCloseDialogue(t, g)
	if !g.hasPartyItem(yellowOrb.Treasure.ItemRawID) || g.storyFlag(yellowOrb.Treasure.PresentFlag) {
		t.Fatalf("黃寶珠後重複建城者對話改變交易狀態：item=%v present=%v",
			g.hasPartyItem(yellowOrb.Treasure.ItemRawID), g.storyFlag(yellowOrb.Treasure.PresentFlag))
	}

	// 六珠來源完成後，正式離開 CTY83 並航向 CTY70 不死鳥祠堂；原版
	// party_has_item consumer 會掃目前隊伍的八格記錄，不能把珠子從同伴
	// 欄位搬運成測試前置，也不能直接寫入祭壇旗標。
	partyHasOrb := func(code int) bool {
		if g.hasItem(code) {
			return true
		}
		for _, member := range g.companions {
			if containsInt(member.Inventory, code) {
				return true
			}
		}
		return false
	}
	missingOrb := -1
	for code := itemGreenOrb; code <= itemSilverOrb; code++ {
		if !partyHasOrb(code) {
			missingOrb = code
			break
		}
	}
	if missingOrb >= 0 {
		t.Logf("黃寶珠 checkpoint 缺少祭壇珠子 %#x：hero=%v companions=%v roster=%v founder=%+v storage=%v",
			missingOrb, g.inventory, compsToSav(g.companions), compsToSav(g.roster),
			g.settlementFounder, g.sharedStorage)
		// 先前的正式離隊流程可能讓持珠同伴留在名冊；玩家可由正式酒場
		// 重新入隊，不能讓測試把名冊資料直接搬到背包。精確來源由本段
		// runtime ledger 記錄後再追其 transaction，不歸因於建城 founder。
		rosterIndex := -1
		for i, member := range g.roster {
			if containsInt(member.Inventory, missingOrb) {
				rosterIndex = i
				break
			}
		}
		if rosterIndex < 0 {
			t.Fatalf("黃寶珠 checkpoint 後目前隊伍／名冊缺少祭壇珠子 %#x：hero=%v companions=%v roster=%v",
				missingOrb, g.inventory, compsToSav(g.companions), compsToSav(g.roster))
		}
		traceExitTownBoundary(t, g, true)
		// 商人城長途航行可能已耗盡魯拉 MP；這不是產品事件 gate。
		// 有足夠 MP 時走正式魯拉，否則從同一個正式世界出口登船回阿里阿罕，
		// 不直接改 MP、座標或造訪表來維持 trace。
		if g.heroMP >= spell.MPCost(fieldRura) {
			traceRuraToCty(t, g, 0)
		} else {
			traceBoardAndSailShip(t, g)
			traceAdventureTravelToCty(t, g, 0, true)
		}
		// 魯拉的正式落點是目的城西側地表格；仍須由玩家步行踏入
		// 城鎮入口，不能把城外落點當成已在酒場內。
		traceAdventureWalkToCty(t, g, 0)
		traceTalkNPC(t, g, 2, 16)
		press(InputState{Confirm: true}) // 找同伴參加
		for i := 0; i < rosterIndex; i++ {
			send(InputState{DirHeld: -1, DirEdge: 0})
		}
		press(InputState{Confirm: true})
		press(InputState{Cancel: true})
		press(InputState{Cancel: true})
		if !partyHasOrb(missingOrb) {
			t.Fatalf("正式酒場重新招募珠子持有者後仍缺 %#x：companions=%v roster=%v",
				missingOrb, compsToSav(g.companions), compsToSav(g.roster))
		}
	}
	if g.inTown {
		traceExitTownBoundary(t, g, true)
	}
	traceBoardAndSailShip(t, g)
	traceAdventureTravelToCty(t, g, ctyPhoenixShrine, true)
	traceTownSectionTo(t, g, ctyPhoenixShrine, 0)
	if !g.inTown || g.curCty != ctyPhoenixShrine || sceneSection(g.cur) != 0 {
		t.Fatalf("黃寶珠後未正式進入不死鳥祠堂：town=%v cty=%d sec=%d",
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	approachPhoenixAltar := func(xy [2]int) {
		standX, standY, face := -1, -1, -1
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			sx, sy := xy[0]+dx, xy[1]+dy
			if sx < 0 || sy < 0 || sx >= g.cur.w || sy >= g.cur.h || g.cur.Blocked(sx, sy) {
				continue
			}
			path := tracePortalPath(g.cur, g.px, g.py, sx, sy, g.keyTier())
			if len(path) == 0 && (g.px != sx || g.py != sy) {
				continue
			}
			for f := 0; f < 4; f++ {
				fx, fy := frontTile(sx, sy, f)
				if fx == xy[0] && fy == xy[1] {
					standX, standY, face = sx, sy, f
					break
				}
			}
			if face >= 0 {
				break
			}
		}
		if face < 0 {
			t.Fatalf("祭壇(%d,%d) 沒有由目前位置 (%d,%d) 可抵達的相鄰站位",
				xy[0], xy[1], g.px, g.py)
		}
		traceWalkToNoPortal(t, g, standX, standY)
		for g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("祭壇(%d,%d) 等待面向輸入：%v", xy[0], xy[1], err)
			}
		}
		if err := g.step(InputState{DirHeld: face, DirEdge: -1}); err != nil {
			t.Fatalf("祭壇(%d,%d) 面向輸入：%v", xy[0], xy[1], err)
		}
	}
	for i, xy := range phoenixAltarTop {
		approachPhoenixAltar(xy)
		press(InputState{Confirm: true})
		send(InputState{DirHeld: -1, DirEdge: 2}) // 對話→咒文
		send(InputState{DirHeld: -1, DirEdge: 1}) // 咒文→調查
		press(InputState{Confirm: true})
		traceCloseDialogue(t, g)
		if g.storyFlag(phoenixAltarFlagFirst+i) || g.hasItem(itemGreenOrb+i) {
			t.Fatalf("祭壇%d transaction 錯：flag=%v item=%v inv=%v",
				i, g.storyFlag(phoenixAltarFlagFirst+i), g.hasItem(itemGreenOrb+i), g.inventory)
		}
	}
	if len(g.placedPhoenixOrbs()) != 6 {
		t.Fatalf("六座祭壇後 placed orb 數=%d，want 6", len(g.placedPhoenixOrbs()))
	}
	// inventory 可有其他物品；這裡只要求六顆珠子都已從主角欄位移除。
	for code := itemGreenOrb; code <= itemSilverOrb; code++ {
		if g.hasItem(code) {
			t.Fatalf("六座祭壇後仍有珠子 %#x：inv=%v", code, g.inventory)
		}
	}
	traceTalkNPC(t, g, 7, 10)
	traceCloseDialogue(t, g)
	for i := 0; g.phoenixStage == 3 && i < 64; i++ {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if !g.phoenixOwned || g.phoenixAboard || g.storyFlag(flagPhoenixUnrevived) ||
		g.phoenixX != phoenixParkX || g.phoenixY != phoenixParkY {
		t.Fatalf("不死鳥復活 transaction 錯：owned=%v aboard=%v flag=%v park=(%d,%d)",
			g.phoenixOwned, g.phoenixAboard, g.storyFlag(flagPhoenixUnrevived), g.phoenixX, g.phoenixY)
	}
	traceExitTownBoundary(t, g, true)
	traceWalkToWorldCoordinate(t, g, phoenixParkX, phoenixParkY+1)
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatalf("正式走上拉米亞：%v", err)
	}
	if !g.phoenixAboard || g.px != phoenixParkX || g.py != phoenixParkY {
		t.Fatalf("正式搭乘拉米亞失敗：aboard=%v pos=(%d,%d)", g.phoenixAboard, g.px, g.py)
	}

	// P5：拉米亞飛抵巴拉摩斯外城。飛行中的原版 mode2 不走地表入口、
	// 不遇敵；玩家必須以正式方向鍵飛到可降落格，再按確認鍵降落，
	// 不能把座標或 phoenixAboard 直接寫成測試前置。
	flyPhoenixTo := func(tx, ty int) {
		for i := 0; i < 2000 && (g.px != tx || g.py != ty); i++ {
			if g.cd > 0 {
				send(InputState{DirHeld: -1, DirEdge: -1})
				continue
			}
			dir := -1
			if g.px < tx {
				dir = 3
			} else if g.px > tx {
				dir = 2
			} else if g.py < ty {
				dir = 0
			} else if g.py > ty {
				dir = 1
			}
			if dir < 0 {
				break
			}
			send(InputState{DirHeld: dir, DirEdge: -1})
		}
		if g.px != tx || g.py != ty {
			t.Fatalf("拉米亞正式飛行無法抵達 (%d,%d)：(%d,%d)", tx, ty, g.px, g.py)
		}
	}
	boardPhoenix := func() {
		if g.inTown || !g.phoenixOwned || g.phoenixAboard {
			t.Fatalf("正式登上拉米亞起點錯：town=%v owned=%v aboard=%v", g.inTown, g.phoenixOwned, g.phoenixAboard)
		}
		// 地表／下層通往拉米亞的路徑仍會進行正式遭遇判定。遭遇一旦
		// 開啟，移動 cooldown 會保留到戰鬥結束；不能把它誤當成單純
		// cooldown 而無限送空白 InputState。
		waitTravelReady := func(where string) {
			for steps := 0; g.battle.active || g.cd > 0; steps++ {
				if steps >= 8192 {
					t.Fatalf("%s 等待移動就緒逾上限：battle=%v mon=%d cd=%d pos=(%d,%d)",
						where, g.battle.active, g.battle.monID, g.cd, g.px, g.py)
				}
				if g.battle.active {
					traceResolveBattle(t, g)
					continue
				}
				send(InputState{DirHeld: -1, DirEdge: -1})
			}
		}
		// 若離場落點正好在拉米亞格，先以方向鍵走開；登乘仍必須由
		// 下一次正式移動踏回 0xfd 格，不直接寫 phoenixAboard。
		if g.px == g.phoenixX && g.py == g.phoenixY {
			waitTravelReady("走離拉米亞前")
			moved := false
			for dir := 0; dir < 4; dir++ {
				dx, dy := dirDelta(dir)
				nx, ny := g.px+dx, g.py+dy
				if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
					g.cur.Blocked(nx, ny) || (nx == g.phoenixX && ny == g.phoenixY) {
					continue
				}
				if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
					t.Fatalf("走離拉米亞：%v", err)
				}
				moved = g.px != g.phoenixX || g.py != g.phoenixY
				if moved {
					break
				}
			}
			if !moved {
				t.Fatal("離場後站在拉米亞格且沒有可走開的正式鄰格")
			}
		}
		boardDir := -1
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny := g.phoenixX-dx, g.phoenixY-dy
			if g.px == nx && g.py == ny {
				boardDir = dir
				break
			}
			if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
				g.cur.Blocked(nx, ny) || findCtyAtLayer(nx, ny, g.layer) >= 0 {
				continue
			}
			if g.layer == 0 {
				traceWalkToWorldCoordinate(t, g, nx, ny)
			} else {
				path := traceWorldPath(g, nx, ny, -1)
				for _, dir := range path {
					waitTravelReady("前往拉米亞相鄰格")
					send(InputState{DirHeld: dir, DirEdge: -1})
				}
			}
			boardDir = dir
			break
		}
		if boardDir < 0 {
			t.Fatalf("拉米亞 (%d,%d) 沒有可由玩家抵達的相鄰陸格", g.phoenixX, g.phoenixY)
		}
		waitTravelReady("登乘拉米亞前")
		if err := g.step(InputState{DirHeld: boardDir, DirEdge: -1}); err != nil {
			t.Fatalf("正式登上拉米亞 dir%d：%v", boardDir, err)
		}
		if !g.phoenixAboard {
			t.Fatalf("正式方向鍵未登上拉米亞：pos=(%d,%d) park=(%d,%d)", g.px, g.py, g.phoenixX, g.phoenixY)
		}
	}
	// 先在已造訪且同時有教會、旅店與聖水貨架的羅馬利亞 CTY2 恢復長途後隊伍，再重新登上
	// 停泊在旅店外的拉米亞；CTY66 本身沒有旅店資料。
	flyPhoenixTo(ctyLoc[2][0], ctyLoc[2][1]-1)
	press(InputState{Confirm: true})
	if g.phoenixAboard || !g.phoenixOwned || g.phoenixX != ctyLoc[2][0] || g.phoenixY != ctyLoc[2][1]-1 {
		t.Fatalf("羅馬利亞前正式降落錯：owned=%v aboard=%v park=(%d,%d)",
			g.phoenixOwned, g.phoenixAboard, g.phoenixX, g.phoenixY)
	}
	traceAdventureWalkToCty(t, g, 2, true)
	traceReviveDeadAtChurch(t, g)
	traceCurePartyPoisonAtChurch(t, g)
	traceTalkFacility(t, g, facInn)
	// 建城前為商人空出隊伍格時，正式酒場流程留下了一名無珠同伴；六珠已在
	// CTY70 祭壇完成交易，現在可於巴拉摩斯戰前把該角色重新招回。這裡只能
	// 經已造訪城鎮、露依達清單與 production InputState，不能直接搬 roster，
	// 也不能把缺少的多拉瑪那 rec175 複製給其他角色。
	if len(g.roster) > 0 {
		if len(g.companions) >= rcPartyMax-1 {
			t.Fatalf("終盤酒場復隊前隊伍已滿：companions=%d roster=%d",
				len(g.companions), len(g.roster))
		}
		traceExitTownBoundary(t, g, true)
		traceRuraToCty(t, g, 0)
		traceAdventureWalkToCty(t, g, 0, false)
		traceTalkNPC(t, g, 2, 16)
		if !g.recruit.active || g.recruit.stage != rcMenu {
			t.Fatalf("終盤復隊未開啟露依達主選單：active=%v stage=%d",
				g.recruit.active, g.recruit.stage)
		}
		press(InputState{Confirm: true}) // 找同伴參加
		press(InputState{Confirm: true}) // roster[0]
		press(InputState{Cancel: true})  // 清單→主選單
		press(InputState{Cancel: true})  // 關閉酒場
		if len(g.companions) != rcPartyMax-1 || len(g.roster) != 0 {
			t.Fatalf("終盤正式酒場復隊錯：companions=%d roster=%d",
				len(g.companions), len(g.roster))
		}
		traceExitTownBoundary(t, g, true)
		traceRuraToCty(t, g, 2)
		traceAdventureWalkToCty(t, g, 2, true)
		traceReviveDeadAtChurch(t, g)
		traceCurePartyPoisonAtChurch(t, g)
		traceTalkFacility(t, g, facInn)
	}
	if !g.hasItem(itemuse.ItemHolyWater) {
		traceTalkFacility(t, g, facItem)
		if bought := buyFromOpenShop(itemuse.ItemHolyWater, 4); bought < 4 {
			t.Fatalf("阿里阿罕後段道具店買不到終盤聖水：bought=%d gold=%d", bought, g.heroGold)
		}
		press(InputState{Cancel: true})
	}
	// 終盤連戰的藥草先在上層正式購買，再用道具面板分配給同伴；
	// 下層沒有可依賴的道具店，戰鬥仍只使用實際隊伍物品，不注入
	// battle.heroHerbs。補給目標是全隊合計八株，不是額外購買八株；
	// 既有藥草與每人八格容量都必須納入交易。
	traceTalkFacility(t, g, facItem)
	topUpHerb(8)
	press(InputState{Cancel: true})
	traceExitTownBoundary(t, g)
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	if g.px == g.phoenixX && g.py == g.phoenixY {
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny := g.px+dx, g.py+dy
			if nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h && !g.cur.Blocked(nx, ny) {
				send(InputState{DirHeld: dir, DirEdge: -1})
				break
			}
		}
	}
	traceWalkToWorldCoordinate(t, g, g.phoenixX, g.phoenixY)
	if !g.phoenixAboard {
		t.Fatalf("羅馬利亞住宿後未重新登上拉米亞：pos=(%d,%d) park=(%d,%d)",
			g.px, g.py, g.phoenixX, g.phoenixY)
	}
	// CTY66 的原始世界入口為 (43,130)；入口上方 (43,129) 是玩家可降落格。
	flyPhoenixTo(43, 129)
	press(InputState{Confirm: true})
	if g.phoenixAboard || !g.phoenixOwned || g.phoenixX != 43 || g.phoenixY != 129 {
		t.Fatalf("巴拉摩斯外城前正式降落錯：owned=%v aboard=%v park=(%d,%d)",
			g.phoenixOwned, g.phoenixAboard, g.phoenixX, g.phoenixY)
	}
	traceAdventureWalkToCty(t, g, 66, true)
	traceUseFieldSpell(t, g, fieldToramana)
	if !g.toramana {
		t.Fatal("進入巴拉摩斯外城後正式施放多拉瑪那失敗")
	}
	traceTownSectionTo(t, g, 65, 0)
	traceTalkNPC(t, g, 8, 3, true)
	if !g.dlg.open || !g.bossIntro || g.battle.active {
		t.Fatalf("巴拉摩斯正式前置錯：dlg=%v intro=%v battle=%v", g.dlg.open, g.bossIntro, g.battle.active)
	}
	traceCloseDialogue(t, g)
	if !g.battle.active || g.battle.monID != 0x79 {
		t.Fatalf("巴拉摩斯正式對話後未開單戰：active=%v mon=%#x", g.battle.active, g.battle.monID)
	}
	traceResolveBattle(t, g, false)
	traceCloseDialogue(t, g)
	if g.battle.active || !g.flags[0x213] || g.storyFlag(0x29) {
		t.Fatalf("巴拉摩斯勝利交易錯：battle=%v f213=%v f29=%v",
			g.battle.active, g.flags[0x213], g.storyFlag(0x29))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存巴拉摩斯 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建巴拉摩斯讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 65 || !g.flags[0x213] || g.storyFlag(0x29) {
		t.Fatalf("巴拉摩斯 save/load 錯：town=%v cty=%d f213=%v f29=%v",
			g.inTown, g.curCty, g.flags[0x213], g.storyFlag(0x29))
	}
	// CTY66 的入口區與南側城外區被 section1 的多個原始 transition
	// 分隔；正式玩家先回到 section0，再走到有地表邊界的 (24,20)
	// portal 區。這裡要把 (24,20) 當作「抵達後站在 portal 上」的
	// exact position，不能誤踏入 section2 內部再尋找出口。
	traceTownSectionTo(t, g, 66, 0, 24, 20)
	traceExitTownBoundary(t, g, true)
	boardPhoenix()
	// 由巴拉摩斯外城入口附近飛回阿里阿罕城鎮入口；城堡與城鎮
	// 共用同一個地表座標，先進 CTY0，再由原始城內 transition 進王城。
	flyPhoenixTo(ctyLoc[0][0], ctyLoc[0][1]-1)
	press(InputState{Confirm: true})
	if g.phoenixAboard || g.phoenixX != ctyLoc[0][0] || g.phoenixY != ctyLoc[0][1]-1 {
		t.Fatalf("阿里阿罕前拉米亞降落錯：aboard=%v park=(%d,%d)", g.phoenixAboard, g.phoenixX, g.phoenixY)
	}
	traceAdventureWalkToCty(t, g, 0, true)
	traceTownSectionTo(t, g, ctyAliahanCastle, 0)
	traceTownSectionTo(t, g, ctyAliahanCastle, aliahanThroneSection)
	traceWalkTo(t, g, aliahanKingX, aliahanKingY+1)
	traceCloseDialogue(t, g)
	if g.storyFlag(0x4d) || g.baramosReturn != 0 {
		t.Fatalf("阿里阿罕王座索瑪現身後狀態錯：flag4d=%v stage=%d",
			g.storyFlag(0x4d), g.baramosReturn)
	}
	traceTownSectionTo(t, g, ctyAliahanCastle, 0)
	// 王城 section0 沒有地表邊界；先走原始王城→CTY0 transition，
	// 再由 CTY0 的正式邊界離開阿里阿罕。
	traceTownSectionTo(t, g, 0, 0)
	traceReviveDeadAtChurch(t, g)   // 巴拉摩斯戰後可能有陣亡隊員；先經正式教會復活
	traceTalkFacility(t, g, facInn) // 下層終盤前先由正式旅店恢復施法資源
	traceExitTownBoundary(t, g, true)
	// 巴拉摩斯後的龍女王事件位於上世界 CTY67：由正式拉米亞
	// 抵達 handler52，取得光之珠並閉合原版 CLEAR flag4e / SET flag19。
	// 光之珠之後以正式「給予」移交同伴，保留主角八格容量給終盤交易；
	// 索瑪弱化消費者仍會掃描整個隊伍的實際物品記錄。
	boardPhoenix()
	flyPhoenixTo(ctyLoc[ctyDragonQueen][0]-1, ctyLoc[ctyDragonQueen][1])
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, ctyDragonQueen, true)
	traceOpenReachableDoor(t, g)
	// handler52 與其他劇情獎勵共用原版全隊八格 writer；終盤補給後
	// 全隊可能再次滿格。談話前只以正式道具選單丟棄藥草或未裝備
	// 備品預留一格，不直接寫光之珠或改 story flags。
	if free := dropSpareEquipmentForPartyInventorySlots(1); free < 1 {
		t.Fatalf("龍女王光之珠：需要1格，正式丟棄藥草／未裝備備品後仍只有%d格", free)
	}
	traceTalkNPC(t, g, 14, 24)
	traceCloseDialogue(t, g)
	if !g.hasPartyItem(itemLightOrb) || g.storyFlag(0x4e) || !g.storyFlag(0x19) {
		t.Fatalf("龍女王光之珠 transaction 錯：item=%v flag4e=%v flag19=%v",
			g.hasPartyItem(itemLightOrb), g.storyFlag(0x4e), g.storyFlag(0x19))
	}
	// 原版 writer 會把道具交給隊伍順序中的第一個空格。只有落在勇者
	// 時才需以正式「給予」移交；若本來就在同伴，已達成保留勇者容量的目的。
	if g.hasItem(itemLightOrb) {
		traceGiveInventoryItem(t, g, itemLightOrb, 0)
	}
	if g.hasItem(itemLightOrb) || !g.hasPartyItem(itemLightOrb) {
		t.Fatalf("正式轉交光之珠後隊伍持有狀態錯：hero=%v party=%v",
			g.hasItem(itemLightOrb), g.hasPartyItem(itemLightOrb))
	}
	traceExitTownBoundary(t, g, true)
	boardPhoenix()
	flyPhoenixTo(ctyLoc[72][0], ctyLoc[72][1]-1)
	press(InputState{Confirm: true})
	if g.phoenixAboard || g.phoenixX != ctyLoc[72][0] || g.phoenixY != ctyLoc[72][1]-1 {
		t.Fatalf("蓋亞洞口前拉米亞降落錯：aboard=%v park=(%d,%d)", g.phoenixAboard, g.phoenixX, g.phoenixY)
	}
	traceAdventureWalkToCty(t, g, 72, true)
	traceTownSectionTo(t, g, 77, 0)
	traceTownSectionTo(t, g, -1, 0)
	if g.inTown || g.layer != 1 || g.px != 85 || g.py != 67 || !g.progressDone(msDescend) {
		t.Fatalf("P5 自然下降未閉合：town=%v layer=%d @(%d,%d) progress=%v",
			g.inTown, g.layer, g.px, g.py, g.progressDone(msDescend))
	}

	// P6：下層彩虹水滴鏈。先由正式飛行進入拉達多姆外城，經 CTY80
	// section2 原始寶箱取得太陽之石，再走妖精之笛→魯比斯→精靈祠堂
	// 的完整道具交易；每一格都由正式 transition／命令窗輸入驅動。
	boardPhoenix()
	flyPhoenixTo(ctyLoc[79][0], ctyLoc[79][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 79, true)
	traceTownSectionTo(t, g, 80, 2)
	sunStone := gamepack.QuestTreasureSelector{
		CTYRaw: 80, Section: 2, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x72, PresentFlag: 0xc2,
	}
	// 太陽之石 type1 寶箱沿用原版全隊八格 writer；下層抵達前的
	// 終盤補給與戰鬥掉落可能填滿所有角色。只由正式道具選單丟棄
	// 藥草或未裝備備品預留一格，不直接改 treasure transaction。
	if free := dropSpareEquipmentForPartyInventorySlots(1); free < 1 {
		t.Fatalf("CTY80 太陽之石：需要1格，正式丟棄藥草／未裝備備品後仍只有%d格", free)
	}
	traceExaminePackTreasure(t, g, sunStone)
	if !g.hasPartyItem(sunStone.ItemRawID) || g.storyFlag(sunStone.PresentFlag) {
		t.Fatalf("太陽之石原始寶箱 transaction 錯：item=%v present=%v",
			g.hasPartyItem(sunStone.ItemRawID), g.storyFlag(sunStone.PresentFlag))
	}
	traceTownSectionTo(t, g, 80, 0)
	// CTY80 sec0 的原始城門先回 CTY79 外城 (44,5)；CTY79
	// section0 再由一般地圖邊界離開，不能把兩段合成一個
	// 「直接到下世界」的 transition。
	// 城門 (19,7) 同時是鎖門與開門後的 CTY79 transition；正式使用
	// 鑰匙後，玩家踏入它會落在 CTY79 (19,6)，因此要以這個原始
	// consumer 結果作為抵達點，而不是把尚未開門時的靜態邊界 graph
	// 當成唯一入口。
	traceWalkThroughPortal(t, g, 11, 0, 79, 0, 19, 6)
	traceExitTownBoundary(t, g, true)
	// 下層船座標必須由原版魯拉 consumer 寫回；稍後造訪 CTY85 後
	// 會使用其 layer1 船表，不能把上層降落時留下的座標沿用到下層。
	boardPhoenix()

	// CTY81 (22,20) 的原始 type1 寶箱是妖精之笛 0x77；不把道具
	// 預塞到背包，並以同一個 legacy selector 走正式調查流程。
	flyPhoenixTo(ctyLoc[81][0], ctyLoc[81][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 81, true)
	fairyFlute := gamepack.QuestTreasureSelector{
		CTYRaw: 81, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x77, PresentFlag: 0x9b,
	}
	// 妖精之笛同樣是 type1／全隊八格交易；太陽之石後的正式航行
	// 遭遇可能重新填滿預留格。只從已證實安全集合正式丟棄一件。
	if free := dropSpareEquipmentForPartyInventorySlots(1); free < 1 {
		t.Fatalf("CTY81 妖精之笛：需要1格，正式丟棄藥草／未裝備備品後仍只有%d格", free)
	}
	traceExaminePackTreasure(t, g, fairyFlute)
	if !g.hasPartyItem(fairyFlute.ItemRawID) || g.storyFlag(fairyFlute.PresentFlag) {
		t.Fatalf("妖精之笛原始寶箱 transaction 錯：item=%v present=%v",
			g.hasPartyItem(fairyFlute.ItemRawID), g.storyFlag(fairyFlute.PresentFlag))
	}
	traceTownSectionTo(t, g, -1, 0)
	boardPhoenix()

	// 魯比斯之塔使用妖精之笛：只由 pack 的 grant_quest_item primitive
	// 實作，妖精之笛不消耗且不再寫舊 remake flags。前段原版商店／
	// 獎勵會讓主角的物品欄可能超過八格；先用正式 rec421「丟掉」
	// 清理不再需要的物品，讓八格容量 gate 與實機一致，而不是直接
	// 改寫背包或呼叫事件 handler。
	flyPhoenixTo(ctyLoc[82][0], ctyLoc[82][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 82, true)
	slots := g.pack.ItemActions().PersonalInventorySlots
	// 先用仍持有的魔法鑰匙開啟魯比斯之塔目前連通區門；
	// 門的執行期覆蓋完成後才正式丟棄鑰匙，後續退出不再依賴它。
	traceOpenReachableDoor(t, g)
	keepP6 := map[int]bool{
		sunStone.ItemRawID:    true,
		fairyFlute.ItemRawID:  true,
		0x56:                  true, // 魔法鑰匙：索瑪城內仍有 tier-2 門
		itemuse.ItemHolyWater: true, // 彩虹橋後正式下船段仍需一瓶聖水
		// 光之珠已正式轉交同伴；保留魔法鑰匙與必要的終盤道具，
		// 由容量 gate 自動丟棄一件非關鍵聖水。
	}
	for g.countItem(itemuse.ItemHolyWater) > 1 {
		traceDropInventoryItem(t, g, itemuse.ItemHolyWater)
	}
	// 主角欄位先以正式「給予」把終盤藥草分配給同伴，讓後續容量
	// 清理只處理主角非關鍵物品；Battle 會依 party inventory 提供
	// 藥草，並在戰後從實際持有者扣除。
	for g.countItem(herbCode) > 0 {
		target := -1
		for i, m := range g.companions {
			if m.itemCount() < slots {
				target = i
				break
			}
		}
		if target < 0 {
			// 同伴欄位已滿時，剩餘主角藥草交由下方正式丟棄
			// transaction 清理；保留已分配的 party reserve。
			break
		}
		traceGiveInventoryItem(t, g, herbCode, target)
	}
	for {
		heroCount := len(g.inventory)
		for _, equipped := range g.equip {
			if equipped >= 0 {
				heroCount++
			}
		}
		if heroCount < slots {
			break
		}
		candidate := -1
		for _, code := range g.inventory {
			if !keepP6[code] {
				candidate = code
				break
			}
		}
		if candidate < 0 {
			t.Fatalf("魯比斯前主角物品欄無法以正式丟棄清出容量：slots=%d equip=%v inv=%v",
				slots, g.equip, g.inventory)
		}
		traceDropInventoryItem(t, g, candidate)
	}
	traceUseInventoryItem(t, g, fairyFlute.ItemRawID)
	if !g.hasPartyItem(0x74) || !g.hasPartyItem(fairyFlute.ItemRawID) || g.flags[0x34] {
		t.Fatalf("魯比斯妖精之笛 transaction 錯：guard=%v flute=%v legacyFlag34=%v",
			g.hasPartyItem(0x74), g.hasPartyItem(fairyFlute.ItemRawID), g.flags[0x34])
	}
	traceTownSectionTo(t, g, -1, 0)
	boardPhoenix()

	// CTY92 精靈：原始地圖 NPC 是 sub0/handler76；pack 保留 handler75
	// 的 IDA 語意與 raw selector，正式話す才將精靈的守護 0x74
	// 原位替換成雲雨之杖 0x73。
	flyPhoenixTo(ctyLoc[92][0], ctyLoc[92][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 92, true)
	staffEvent, ok := g.pack.NPCItemRewardEvent("dq3:event.staff_of_rain_exchange")
	if !ok || staffEvent.RequiredItemRawID == nil {
		t.Fatal("缺雲雨之杖精靈 pack event")
	}
	traceTalkNPC(t, g, staffEvent.NPC.Tile.X, staffEvent.NPC.Tile.Y)
	traceCloseDialogue(t, g)
	if !g.hasItem(staffEvent.GrantedItemRaw) || g.hasItem(*staffEvent.RequiredItemRawID) ||
		g.storyFlag(staffEvent.PresentFlagRaw) {
		t.Fatalf("正式取得雲雨之杖錯：inv=%v present=%v",
			g.inventory, g.storyFlag(staffEvent.PresentFlagRaw))
	}
	if err := g.Save(); err != nil {
		t.Fatalf("保存雲雨之杖 checkpoint：%v", err)
	}
	restored, err = NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("重建雲雨之杖讀檔 Game：%v", err)
	}
	g = restored
	g.frame = nil
	press(InputState{Confirm: true})
	send(InputState{DirHeld: -1, DirEdge: 0})
	press(InputState{Confirm: true})
	if !g.inTown || g.curCty != 92 || !g.hasItem(staffEvent.GrantedItemRaw) ||
		g.storyFlag(staffEvent.PresentFlagRaw) {
		t.Fatalf("雲雨之杖 save/load 錯：town=%v cty=%d staff=%v present=%v",
			g.inTown, g.curCty, g.hasItem(staffEvent.GrantedItemRaw),
			g.storyFlag(staffEvent.PresentFlagRaw))
	}
	traceTalkNPC(t, g, staffEvent.NPC.Tile.X, staffEvent.NPC.Tile.Y)
	traceCloseDialogue(t, g)
	if len(g.inventory) == 0 || !g.hasItem(staffEvent.GrantedItemRaw) ||
		g.hasItem(*staffEvent.RequiredItemRawID) {
		t.Fatalf("雲雨之杖重複對話改變道具：%v", g.inventory)
	}
	traceTownSectionTo(t, g, -1, 0)
	boardPhoenix()
	// 先造訪有原始下層船表的 CTY85，再以正式魯拉寫入其
	// layer1 停泊座標；CTY79 的船格與彩虹水滴海域不相連，不能
	// 把另一筆目的地近似成同一港口。
	flyPhoenixTo(ctyLoc[85][0], ctyLoc[85][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 85, true)
	// CTY85 的 section0 原始出口是外框邊界，不是 section transition
	// table 的 dsec=0xfe 節點；沿玩家可走的邊界送正式方向鍵離城。
	traceExitTownBoundary(t, g)
	traceRuraToCty(t, g, 85)
	boardPhoenix()

	// CTY93 神聖祠堂的合成事件由正式「調查」命令觸發；太陽之石與
	// 雲雨之杖均來自上面兩個原始 consumer，成功才生成彩虹水滴。
	flyPhoenixTo(ctyLoc[93][0], ctyLoc[93][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 93, true)
	traceExamineCurrent(t, g)
	if !g.hasItem(itemRainbowDrop) || g.hasItem(itemSunStone) || g.hasItem(itemRaincloudRod) ||
		!g.progressDone(msRainbow) {
		t.Fatalf("神聖祠堂彩虹合成錯：inv=%v rainbow=%v progress=%v",
			g.inventory, g.hasItem(itemRainbowDrop), g.progressDone(msRainbow))
	}
	traceTownSectionTo(t, g, -1, 0)

	// 彩虹水滴使用格是下層海面，飛行中的拉米亞不能降落；依原版
	// vehicle mode 先正式登船、航行到 (127,117)，再由命令窗使用。
	// 魯拉的下層船座標位在獨立海岸，先由拉米亞正式飛到其相鄰
	// 陸格，再踏上停泊船；不把船或玩家位置直接注入測試狀態。
	portX, portY := -1, -1
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		x, y := g.shipX-dx, g.shipY-dy
		if x >= 0 && y >= 0 && x < g.cur.w && y < g.cur.h &&
			!g.cur.Blocked(x, y) && g.cur.attr.Raw(g.cur.tileIdx(x, y))&0x0002 != 0 {
			portX, portY = x, y
			break
		}
	}
	if portX < 0 {
		t.Fatalf("下層停泊船 (%d,%d) 找不到可降落相鄰陸格", g.shipX, g.shipY)
	}
	// 船上仍會執行原版弱敵遭遇判定；保持 CTY0 旅店後的施法資源，
	// 不在無法回城的下層水路額外消耗主角 MP。
	boardPhoenix()
	flyPhoenixTo(portX, portY)
	press(InputState{Confirm: true})
	if g.phoenixAboard {
		t.Fatalf("下層船港陸格 (%d,%d) 拉米亞無法正式降落", portX, portY)
	}
	traceBoardAndSailShip(t, g)
	traceSailToWorldCoordinate(t, g, rainbowUseX, rainbowUseY)
	traceUseInventoryItem(t, g, itemRainbowDrop)
	if g.worldState&worldStateRainbowBridge == 0 || g.hasItem(itemRainbowDrop) {
		t.Fatalf("彩虹水滴正式使用錯：worldState=%#x item=%v", g.worldState, g.hasItem(itemRainbowDrop))
	}

	// 彩虹橋後進 CTY90：隱藏樓梯、歐魯迪卡橋事件、索瑪三連戰。
	// 原始下層海圖的彩虹格是船路瓶頸；建橋後船仍停在南側水格，
	// 不能把船路直接延伸到索瑪城護城河。依正式 vehicle mode，先向左
	// 踏上新生成的橋，再徒步回到停泊的拉米亞，最後由拉米亞飛到
	// 索瑪城西側可降落格。這保留了「使用彩虹水滴→橋 tile→飛行→
	// 進城」的玩家可見事件鏈，不注入船或玩家座標。
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 2, DirEdge: -1}) // 船 (127,117) → 彩虹橋 (126,117)
	if g.shipAboard || g.px != rainbowBridgeX || g.py != rainbowBridgeY {
		t.Fatalf("彩虹橋正式下船錯：aboard=%v player=(%d,%d) ship=(%d,%d)",
			g.shipAboard, g.px, g.py, g.shipX, g.shipY)
	}
	if !g.hasItem(itemuse.ItemHolyWater) {
		t.Fatal("彩虹橋後正式徒步段缺少聖水")
	}
	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	traceWalkLowerWorldCoordinate(t, g, g.phoenixX, g.phoenixY)
	if !g.phoenixAboard {
		t.Fatalf("彩虹橋後徒步回拉米亞停泊格未正式登乘：pos=(%d,%d) park=(%d,%d)",
			g.px, g.py, g.phoenixX, g.phoenixY)
	}
	flyPhoenixTo(ctyLoc[ctyZomaCastle][0]-1, ctyLoc[ctyZomaCastle][1])
	press(InputState{Confirm: true})
	if g.phoenixAboard || g.phoenixX != ctyLoc[ctyZomaCastle][0]-1 ||
		g.phoenixY != ctyLoc[ctyZomaCastle][1] {
		t.Fatalf("索瑪城前拉米亞正式降落錯：aboard=%v park=(%d,%d)",
			g.phoenixAboard, g.phoenixX, g.phoenixY)
	}
	traceAdventureWalkToCty(t, g, ctyZomaCastle, true)
	traceTownSectionTo(t, g, ctyZomaCastle, 0, 23, 4)
	traceWalkToNoPortal(t, g, 23, 4) // helper 只保證 section；正式調查仍需站在 event tile 下方
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 1, DirEdge: -1})
	traceExamineCurrent(t, g)
	traceCloseDialogue(t, g)
	traceTownSectionTo(t, g, ctyZomaCastle, 1)
	traceTownSectionTo(t, g, ctyZomaCastle, ortegaSection, 19, 29)
	traceWalkToNoPortal(t, g, 19, 29) // section graph 到達後，仍須正式走到橋東側觸發格
	for g.cd > 0 {
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	send(InputState{DirHeld: 2, DirEdge: -1})
	if !g.dlg.open || g.ortegaStage != 1 {
		t.Fatalf("歐魯迪卡正式橋事件未觸發：dlg=%v stage=%d", g.dlg.open, g.ortegaStage)
	}
	traceCloseDialogue(t, g)
	// sec5 由多個不相連的原始 transition component 組成；先從目前
	// section5 的 pack scene 取得 handler80 座標，再讓 transition graph
	// 選到能真正走到索瑪 NPC 的 component，不能把入口 spawn 當成全區可達。
	zomaScene, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyZomaCastle, mapBlkNum[ctyZomaCastle], zomaFinalSection, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatalf("載入 CTY90 sec5 pack scene：%v", err)
	}
	zomaX, zomaY := -1, -1
	for _, npc := range zomaScene.npcs {
		if npc.b4 == zomaHandler {
			zomaX, zomaY = npc.x, npc.y
			break
		}
	}
	if zomaX < 0 {
		t.Fatal("CTY90 sec5 正式路線找不到索瑪 NPC")
	}
	traceTownSectionTo(t, g, ctyZomaCastle, zomaFinalSection, zomaX, zomaY)
	if !g.hasPartyItem(itemLightOrb) {
		t.Fatal("索瑪前隊伍缺少龍女王給予的光之珠")
	}
	traceTalkNPC(t, g, zomaX, zomaY)
	for i := 0; i < 16 && !g.cleared; i++ {
		if g.battle.active {
			traceResolveBattle(t, g, false, false)
			continue
		}
		if g.dlg.open {
			traceCloseDialogue(t, g)
			continue
		}
		send(InputState{DirHeld: -1, DirEdge: -1})
	}
	if !g.cleared || g.battle.active || g.dlg.open || g.layer != 1 || g.inTown {
		t.Fatalf("索瑪終盤正式連戰未閉合：cleared=%v battle=%v dlg=%v layer=%d town=%v",
			g.cleared, g.battle.active, g.dlg.open, g.layer, g.inTown)
	}

	// 結局前必須由玩家返回拉達多姆王座；正式飛行→進 CTY79→CTY80
	// section1→國王 handler74，關閉冊封對話後才開始 ENDTXT。
	boardPhoenix()
	flyPhoenixTo(ctyLoc[79][0], ctyLoc[79][1]-1)
	press(InputState{Confirm: true})
	traceAdventureWalkToCty(t, g, 79, true)
	traceTownSectionTo(t, g, ctyRadatomeCastle, radatomeThroneSec)
	kingX, kingY := -1, -1
	for _, npc := range g.cur.npcs {
		if npc.b4 == radatomeKingHandle {
			kingX, kingY = npc.x, npc.y
			break
		}
	}
	if kingX < 0 {
		t.Fatal("CTY80 sec1 找不到國王 handler74")
	}
	traceTalkNPC(t, g, kingX, kingY)
	traceCloseDialogue(t, g)
	for i := 0; i < 128 && g.endSeq >= 0; i++ {
		press(InputState{Confirm: true})
	}
	if !g.lotoBlessed || !g.flags[0x217] || g.endSeq >= 0 {
		t.Fatalf("THE END 正式冊封／捲動未閉合：loto=%v flag217=%v endSeq=%d",
			g.lotoBlessed, g.flags[0x217], g.endSeq)
	}
	if out := os.Getenv("DQ3_PRODUCTION_DUMP_FINAL"); out != "" {
		if len(g.endingPix) != ScreenW*ScreenH || len(g.endingPal) == 0 {
			t.Fatalf("終盤 pack asset 尚未載入：pix=%d pal=%d", len(g.endingPix), len(g.endingPal))
		}
		// 主要 trace 期間刻意停用 frame；只在正式 THE END 狀態重建一次
		// Ebiten image，避免長流程累積圖形命令，也不改玩家狀態。
		g.frame = ebiten.NewImage(ScreenW, ScreenH)
		g.renderFrame()
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			t.Fatalf("建立終盤 PNG 目錄：%v", err)
		}
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		copy(img.Pix, g.rgba)
		f, err := os.Create(out)
		if err != nil {
			t.Fatalf("建立終盤 PNG：%v", err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			t.Fatalf("寫入終盤 PNG：%v", err)
		}
		if err := f.Close(); err != nil {
			t.Fatalf("關閉終盤 PNG：%v", err)
		}
		t.Logf("正式 THE END runtime PNG：%s", out)
	}
}

// traceWaitForDayNearCty 在城鎮外只送正常方向鍵並處理正式隨機遭遇，
// 直到地表步數週期回到白天。它不修改晝夜 counter、座標或 RNG。
func traceWaitForDayNearCty(t *testing.T, g *Game, cty int) {
	t.Helper()
	if g.inTown || cty < 0 || cty >= len(ctyLoc) {
		t.Fatalf("等待白天起點錯：town=%v cty=%d", g.inTown, cty)
	}
	prevX, prevY := -1, -1
	for steps := 0; steps < 2000 && g.dnPhase != 0; steps++ {
		if g.battle.active {
			// 等待晝夜只需要正常走動；遇到目前隊伍不適合硬打的後期怪物時，
			// 以正式戰鬥選單逃跑，不用測試注入改 HP、RNG 或遭遇表。
			traceResolveBattle(t, g, g.battle.monID >= 80)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待白天 cooldown：%v", err)
			}
			continue
		}
		chosen := -1
		fallback := -1
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny := g.px+dx, g.py+dy
			if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
				g.cur.Blocked(nx, ny) || findCtyAtLayer(nx, ny, g.layer) >= 0 ||
				nx == g.shipX && ny == g.shipY {
				continue
			}
			if fallback < 0 {
				fallback = dir
			}
			if nx != prevX || ny != prevY {
				chosen = dir
				break
			}
		}
		if chosen < 0 {
			chosen = fallback
		}
		if chosen < 0 {
			t.Fatalf("CTY%d 外 (%d,%d) 找不到可用於等待白天的陸格",
				cty, g.px, g.py)
		}
		oldX, oldY := g.px, g.py
		if err := g.step(InputState{DirHeld: chosen, DirEdge: -1}); err != nil {
			t.Fatalf("等待白天移動 dir%d：%v", chosen, err)
		}
		prevX, prevY = oldX, oldY
	}
	if g.dnPhase != 0 {
		t.Fatalf("CTY%d 外 2000 個 production step 後仍未回白天：phase=%d step=%d",
			cty, g.dnPhase, g.dnStep)
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
		if g.heroHP <= 0 || g.heroHP*4 <= heroMax*3 {
			return true
		}
		for _, m := range g.companions {
			if m.CurHP <= 0 || m.CurHP*4 <= m.MaxHP()*3 {
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
		// 有些城鎮（提頓）夜間會換掉旅店 NPC；在城外以正常步行等到白天，
		// 再走正式入口，不直接寫晝夜 phase。
		if g.dnPhase != 0 {
			traceWaitForDayNearCty(t, g, cty)
		}
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

	for step := 0; step < 500000; step++ {
		level, _, _, _, _ := g.heroStats()
		if level >= wantLevel {
			if needsRest() {
				rest()
			}
			t.Logf("正式低危區練級完成：hero Lv%d exp=%d gold=%d", level, g.heroExp, g.heroGold)
			return
		}
		if g.battle.active {
			if g.battle.phase == phCommand && g.battle.commandActor == g.battle.firstCommandActor() {
				statuses := make([]int, 1+len(g.battle.companions))
				statuses[0] = g.battle.heroStatus
				for i, actor := range g.battle.companions {
					statuses[i+1] = actor.status
				}
				for _, status := range statuses {
					if status&statusPersistentParalysis != 0 {
						t.Logf("練級戰進入命令階段時仍有持久麻痺：mon=%d statuses=%v field=%#x companions=%v",
							g.battle.monID, statuses, g.heroConditions, compsToSav(g.companions))
						break
					}
				}
			}
			// 以全隊最低等級而非主角等級判定訓練戰力：達瑪轉職後主角雖已
			// 高等，賢者仍為 Lv1。原版正常玩家面對超出該隊員承受力的編隊
			// 會由正式戰鬥選單逃跑，而不是把一次全滅當成資料或路由錯誤。
			level, _, _, _, _ := g.heroStats()
			weakestLevel := level
			for _, member := range g.companions {
				if member.Level() < weakestLevel {
					weakestLevel = member.Level()
				}
			}
			// raw monster ID 不是戰力排序。docs/154 要求依目前隊伍與
			// 敵群的實際 HP／攻防／速度與已閉合持久狀態作保守判斷。
			flee := cty == 0 || !traceCanSafelyFinishEncounter(g)
			traceResolveBattle(t, g, flee)
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
	t.Fatalf("正式低危區練級 500000 step 仍未達 Lv%d：exp=%d", wantLevel, g.heroExp)
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

// traceCurePartyPoisonAtChurch 只在目前隊伍有人中毒時，經正式教會
// service0→target→yes/no 逐人解毒。旅店不清 poison；這個 helper 不直接
// 修改 Conditions 或金錢。
func traceCurePartyPoisonAtChurch(t *testing.T, g *Game) {
	t.Helper()
	poisoned := func(target int) bool {
		if target == 0 {
			return g.heroConditions&conditionPoison != 0
		}
		return target > 0 && target <= len(g.companions) &&
			g.companions[target-1].Conditions&conditionPoison != 0
	}
	hasPoison := false
	for target := 0; target < g.churchPartyLen(); target++ {
		if poisoned(target) {
			hasPoison = true
			break
		}
	}
	if !hasPoison {
		return
	}
	beforeGold := g.heroGold
	wantCost := 0
	traceTalkFacility(t, g, facChurch)
	if !g.church.active || g.church.stage != churchService {
		t.Fatalf("中毒後未進入正式教會 modal：active=%v stage=%d",
			g.church.active, g.church.stage)
	}
	for target := 0; target < g.churchPartyLen(); target++ {
		if !poisoned(target) {
			continue
		}
		wantCost += g.pack.CurePoisonCost()
		for g.church.serviceCursor != 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: 1}); err != nil {
				t.Fatalf("教會選解毒服務：%v", err)
			}
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會確認解毒服務：%v", err)
		}
		for g.church.targetCursor != target {
			if err := g.step(InputState{DirHeld: -1, DirEdge: 0}); err != nil {
				t.Fatalf("教會選中毒角色%d：%v", target, err)
			}
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會確認中毒角色%d：%v", target, err)
		}
		if g.church.stage != churchConfirm {
			t.Fatalf("教會中毒角色%d 未進 yes/no：stage=%d msg=%q",
				target, g.church.stage, g.church.msg)
		}
		if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("教會支付角色%d解毒：%v", target, err)
		}
		if poisoned(target) {
			t.Fatalf("教會交易後角色%d仍中毒", target)
		}
	}
	if err := g.step(InputState{Cancel: true, DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatalf("離開教會解毒：%v", err)
	}
	if g.heroGold != beforeGold-wantCost {
		t.Fatalf("教會解毒扣款錯：before=%d after=%d wantCost=%d",
			beforeGold, g.heroGold, wantCost)
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

// traceRuraToCty 只經正式命令窗、咒文清單與魯拉目的地清單操作。
// 目的城必須已由玩家正常造訪；找不到目的地或施法者 MP 不足時 fail closed。
func traceRuraToCty(t *testing.T, g *Game, wantCty int) {
	t.Helper()
	press := func(in InputState) {
		if err := g.step(in); err != nil {
			t.Fatalf("魯拉正式輸入：%v", err)
		}
	}
	if g.battle.active {
		traceResolveBattle(t, g)
	}
	for g.cd > 0 {
		press(InputState{DirHeld: -1, DirEdge: -1})
	}
	press(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 開命令窗
	if !g.cmd.open {
		t.Fatalf("魯拉前未能開啟正式命令窗：town=%v cty=%d battle=%v",
			g.inTown, g.curCty, g.battle.active)
	}
	for moves := 0; moves < 2 && g.cmd.cursor != int(cmdSpell); moves++ {
		press(InputState{DirHeld: -1, DirEdge: 3})
	}
	if g.cmd.cursor != int(cmdSpell) {
		t.Fatalf("正式命令窗兩次水平輸入後仍未選到咒文：cursor=%d", g.cmd.cursor)
	}
	press(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.active {
		t.Fatalf("正式命令窗未開啟野外咒文：hero=%d MP=%d conditions=%#x companions=%+v roster=%+v town=%v cty=%d sec=%d",
			g.heroHP, g.heroMP, g.heroConditions, compsToSav(g.companions), compsToSav(g.roster),
			g.inTown, g.curCty, sceneSection(g.cur))
	}
	caster := traceFieldSpellCaster(g, fieldRura)
	if caster < 0 {
		t.Fatal("目前隊伍沒有存活且 MP 足夠的魯拉施法者")
	}
	for g.fieldSpell.cursor != caster {
		press(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	spellIndex := -1
	for i, choice := range g.fieldSpell.choices {
		if choice.rec == fieldRura {
			spellIndex = i
			break
		}
	}
	if spellIndex < 0 {
		t.Fatal("目前隊伍尚未習得魯拉")
	}
	for g.fieldSpell.cursor != spellIndex {
		press(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.dest {
		caster, mp, cost := -1, -1, -1
		if spellIndex >= 0 && spellIndex < len(g.fieldSpell.choices) {
			choice := g.fieldSpell.choices[spellIndex]
			caster, mp, cost = choice.caster, g.fieldCasterMP(choice.caster), choice.mp
		}
		mapFlags := byte(0)
		if g.cur != nil {
			mapFlags = g.cur.mapFlags
		}
		t.Fatalf("魯拉未進入已造訪城鎮清單：town=%v cty=%d sec=%d mapFlags=%#x caster=%d mp=%d heroMaxMP=%d heroStat=%v cost=%d active=%v visits=%+v",
			g.inTown, g.curCty, sceneSection(g.cur), mapFlags, caster, mp, g.heroMaxMP(),
			g.heroStat, cost, g.fieldSpell.active, g.visitedTowns)
	}
	destIndex := -1
	for i, visit := range g.visitedTowns {
		if visit.Cty == wantCty {
			destIndex = i
			break
		}
	}
	if destIndex < 0 {
		t.Fatalf("魯拉目的地 CTY%d 尚未由玩家造訪：%+v", wantCty, g.visitedTowns)
	}
	for g.fieldSpell.cursor != destIndex {
		press(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.inTown || g.curCty >= 0 || g.layer != ctyLoc[wantCty][2] ||
		g.px != ctyLoc[wantCty][0]-1 || g.py != ctyLoc[wantCty][1] {
		t.Fatalf("魯拉未抵達 CTY%d 西側地表：town=%v cty=%d layer=%d @(%d,%d)",
			wantCty, g.inTown, g.curCty, g.layer, g.px, g.py)
	}
}

func traceFieldSpellCaster(g *Game, wantRec int) int {
	if g.pack == nil {
		return -1
	}
	def, ok := g.pack.FieldSpellByRecord(wantRec)
	if !ok {
		return -1
	}
	for actor := 0; actor <= len(g.companions); actor++ {
		if g.fieldActorAlive(actor) && g.fieldCasterMP(actor) >= def.MPCost &&
			containsRec(g.fieldActorSpells(actor), wantRec) {
			return actor
		}
	}
	return -1
}

// traceUseFieldSpell 只經正式命令窗、施法者、咒文與必要的單體目標頁。
// targetActor 只供 pack 宣告 party_member 的咒文；場景咒文不得傳入。
func traceUseFieldSpell(t *testing.T, g *Game, wantRec int, targetActor ...int) {
	t.Helper()
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatalf("場景咒文正式輸入 rec%d：%v", wantRec, err)
		}
	}
	for g.cd > 0 {
		step(InputState{DirHeld: -1, DirEdge: -1})
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.cmd.open {
		t.Fatalf("場景咒文 rec%d 前未能開啟命令窗", wantRec)
	}
	for moves := 0; moves < 4 && g.cmd.cursor != int(cmdSpell); moves++ {
		step(InputState{DirHeld: -1, DirEdge: 3})
	}
	if g.cmd.cursor != int(cmdSpell) {
		t.Fatalf("正式命令窗無法選到咒文：cursor=%d rec%d", g.cmd.cursor, wantRec)
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.active {
		t.Fatalf("正式命令窗未開啟場景咒文清單 rec%d", wantRec)
	}
	caster := traceFieldSpellCaster(g, wantRec)
	if caster < 0 {
		t.Fatalf("目前隊伍無存活且 MP 足夠的施法者可用 rec%d", wantRec)
	}
	for g.fieldSpell.cursor != caster {
		step(InputState{DirHeld: -1, DirEdge: 0})
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	spellIndex := -1
	for i, choice := range g.fieldSpell.choices {
		if choice.rec == wantRec {
			spellIndex = i
			break
		}
	}
	if spellIndex < 0 {
		t.Fatalf("目前隊伍無存活施法者可用場景咒文 rec%d：choices=%v cty=%d sec=%d pos=(%d,%d) heroHP=%d companions=%+v roster=%+v",
			wantRec, g.fieldSpell.choices, g.curCty, sceneSection(g.cur), g.px, g.py,
			g.heroHP, compsToSav(g.companions), compsToSav(g.roster))
	}
	for g.fieldSpell.cursor != spellIndex {
		step(InputState{DirHeld: -1, DirEdge: 0})
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.targeting {
		if len(targetActor) != 1 || targetActor[0] < 0 || targetActor[0] > len(g.companions) {
			t.Fatalf("rec%d 需要唯一合法 party target，got %v", wantRec, targetActor)
		}
		for g.fieldSpell.cursor != targetActor[0] {
			step(InputState{DirHeld: -1, DirEdge: 0})
		}
		step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	} else if len(targetActor) != 0 {
		t.Fatalf("rec%d 不是單體目標咒文，不應傳 target %v", wantRec, targetActor)
	}
	if g.fieldSpell.active {
		choice := fieldSpellChoice{}
		if g.fieldSpell.cursor >= 0 && g.fieldSpell.cursor < len(g.fieldSpell.choices) {
			choice = g.fieldSpell.choices[g.fieldSpell.cursor]
		}
		t.Fatalf("場景咒文 rec%d 施放後 modal 仍開啟：dest=%v cursor=%d choice={rec=%d caster=%d cost=%d casterMP=%d} choices=%+v",
			wantRec, g.fieldSpell.dest, g.fieldSpell.cursor, choice.rec, choice.caster, choice.mp,
			g.fieldCasterMP(choice.caster), g.fieldSpell.choices)
	}
}

func tracePartyActorNeedsHealing(g *Game, actor int) bool {
	if actor == 0 {
		_, maxHP, _, _, _ := g.heroStats()
		return g.heroHP > 0 && g.heroHP < maxHP
	}
	if actor < 1 || actor > len(g.companions) {
		return false
	}
	m := g.companions[actor-1]
	return m != nil && m.CurHP > 0 && m.CurHP < m.MaxHP()
}

func tracePartyLivingNeedsHealing(g *Game) bool {
	for actor := 0; actor <= len(g.companions); actor++ {
		if tracePartyActorNeedsHealing(g, actor) {
			return true
		}
	}
	return false
}

// traceHealLivingParty is a deterministic player strategy over the formal
// field-spell UI. It prefers an affordable full-party heal, then the strongest
// affordable single-target heal; herbs remain a legal last resort only after
// no living caster has enough MP.
func traceHealLivingParty(t *testing.T, g *Game) {
	t.Helper()
	for casts := 0; casts < 64 && tracePartyLivingNeedsHealing(g); casts++ {
		cast := false
		for _, rec := range []int{165, 164} {
			if traceFieldSpellCaster(g, rec) >= 0 {
				traceUseFieldSpell(t, g, rec)
				cast = true
				break
			}
		}
		if cast {
			continue
		}
		for actor := 0; actor <= len(g.companions) && !cast; actor++ {
			if !tracePartyActorNeedsHealing(g, actor) {
				continue
			}
			for _, rec := range []int{163, 162, 161} {
				if traceFieldSpellCaster(g, rec) >= 0 {
					traceUseFieldSpell(t, g, rec, actor)
					cast = true
					break
				}
			}
		}
		if cast {
			continue
		}
		if g.hasPartyItem(itemuse.ItemHerb) {
			traceUseInventoryItem(t, g, itemuse.ItemHerb)
			continue
		}
		break
	}
}

func traceProtectBaramosHazard(t *testing.T, g *Game, nx, ny int) {
	t.Helper()
	if (g.curCty != 65 && g.curCty != 66 && g.curCty != ctyZomaCastle) || g.cur == nil || g.cur.attr == nil ||
		nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h {
		return
	}
	ah := byte(g.cur.attr.Raw(g.cur.tileIdx(nx, ny)) >> 8)
	if ah&0x0c != 0 && !g.toramana && !g.hazardGuard {
		traceUseFieldSpell(t, g, fieldToramana)
	}
}

// traceUseInventoryItem 只送 production 命令窗／道具面板輸入，從隊伍個人物品欄
// 中依 owner 順序選到指定 id 後使用。
// 效果另需選目標的道具必須傳入 target actor（0=主角，1..=同伴）；持有者由
// 隊伍中第一個實際持有該 raw id 的角色決定。
func traceUseInventoryItem(t *testing.T, g *Game, code int, targetActor ...int) {
	t.Helper()
	owner, idx := -1, -1
	for actor := 0; actor <= len(g.companions) && owner < 0; actor++ {
		items := g.equipActorInventory(actor)
		if items == nil {
			continue
		}
		for i, got := range *items {
			if got == code {
				owner, idx = actor, i
				break
			}
		}
	}
	if idx < 0 {
		t.Fatalf("隊伍沒有要使用的道具 0x%02x", code)
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
	if len(g.companions) > 0 {
		for i := 0; i < owner; i++ {
			step(InputState{DirEdge: 0})
		}
		step(InputState{Confirm: true}) // owner selector
	}
	for i := 0; i < idx; i++ {
		step(InputState{DirEdge: 0})
	}
	if g.panel != panelItem || g.panelCursor != idx {
		t.Fatalf("正式道具選單未移到 index%d：panel=%d cursor=%d", idx, g.panel, g.panelCursor)
	}
	step(InputState{Confirm: true})
	step(InputState{Confirm: true}) // 原版 rec421「使用／給予／丟掉」預設選使用
	if g.panel == panelItem && g.itemActionStage == itemActionUseTarget {
		if len(targetActor) != 1 || targetActor[0] < 0 || targetActor[0] > len(g.companions) {
			t.Fatalf("道具 0x%02x 需要合法選人目標：%v", code, targetActor)
		}
		for i := 0; i < targetActor[0]; i++ {
			step(InputState{DirEdge: 0})
		}
		step(InputState{Confirm: true})
	} else if len(targetActor) != 0 {
		t.Fatalf("道具 0x%02x 未進選人 modal，卻指定 actor=%v", code, targetActor)
	}
	if g.panel != panelNone { // 聖水等一般消耗品使用後仍留在清單，正式按 B 關閉
		step(InputState{Cancel: true})
		if g.panel != panelNone { // 多人隊伍：清單→owner selector→關閉
			step(InputState{Cancel: true})
		}
	}
}

// traceCureHeroPoison 只在主角仍存活、中毒且背包已有驅毒草時，走正式道具選人流程。
// 回傳 true 表示本輪已消費輸入；呼叫端應重新計算路徑。
func traceCureHeroPoison(t *testing.T, g *Game) bool {
	t.Helper()
	if g.battle.active || g.dlg.open || g.panel != panelNone || g.heroHP <= 0 ||
		g.heroConditions&conditionPoison == 0 || !g.hasItem(itemuse.ItemAntidote) {
		return false
	}
	traceUseInventoryItem(t, g, itemuse.ItemAntidote, 0)
	traceCloseDialogue(t, g)
	if g.heroConditions&conditionPoison != 0 {
		t.Fatalf("正式使用驅毒草後主角仍中毒：HP=%d item=%v", g.heroHP,
			g.hasItem(itemuse.ItemAntidote))
	}
	return true
}

// traceDropInventoryItem 只用正式命令窗與 rec421 動作選單丟掉一件勇者
// 物品；呼叫端需自行保留後續事件所需的道具。
func traceDropInventoryItem(t *testing.T, g *Game, code int) {
	traceDropActorInventoryItem(t, g, 0, code)
}

// traceDropActorInventoryItem 經正式 owner selector 丟棄指定角色的一件物品。
func traceDropActorInventoryItem(t *testing.T, g *Game, actor, code int) {
	t.Helper()
	items := g.equipActorInventory(actor)
	if items == nil {
		t.Fatalf("丟棄持有者不存在：actor=%d party=%d", actor, len(g.companions)+1)
	}
	idx := -1
	for i, got := range *items {
		if got == code {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("actor%d 沒有要丟棄的道具 0x%02x：%v", actor, code, *items)
	}
	before := len(*items)
	beforeCount := countOccurrences(*items, code)
	step := func(in InputState) {
		t.Helper()
		if in.DirHeld == 0 {
			in.DirHeld = -1
		}
		if in.DirEdge == 0 && in.Confirm {
			in.DirEdge = -1
		}
		if err := g.step(in); err != nil {
			t.Fatalf("丟棄道具 production input: %v", err)
		}
	}
	step(InputState{Confirm: true}) // 開命令窗
	step(InputState{DirEdge: 3})    // 對話→咒文
	step(InputState{DirEdge: 0})    // 咒文→道具
	step(InputState{Confirm: true}) // 開道具面板
	for i := 0; i < actor; i++ {
		step(InputState{DirEdge: 0})
	}
	if len(g.companions) > 0 {
		step(InputState{Confirm: true}) // 選定持有者
	}
	for i := 0; i < idx; i++ {
		step(InputState{DirEdge: 0})
	}
	step(InputState{Confirm: true}) // 選取道具→rec421
	step(InputState{DirEdge: 0})    // 使用→給予
	step(InputState{DirEdge: 0})    // 給予→丟掉
	step(InputState{Confirm: true})
	if len(*items) != before-1 || countOccurrences(*items, code) != beforeCount-1 {
		t.Fatalf("正式丟棄未移除 actor%d 道具 0x%02x：before=%d after=%d inv=%v", actor, code, before, len(*items), *items)
	}
	if g.panel != panelNone {
		step(InputState{Cancel: true})
		if g.panel != panelNone {
			step(InputState{Cancel: true})
		}
	}
}

func countOccurrences(items []int, code int) int {
	n := 0
	for _, item := range items {
		if item == code {
			n++
		}
	}
	return n
}

// traceGiveActorInventoryItem 只使用正式 owner selector、rec421 與全隊目標選單，
// 從指定角色的指定未裝備欄移交一件物品。
func traceGiveActorInventoryItem(t *testing.T, g *Game, owner, idx, targetActor int) {
	t.Helper()
	source := g.equipActorInventory(owner)
	dest := g.equipActorInventory(targetActor)
	if source == nil || dest == nil || idx < 0 || idx >= len(*source) || owner == targetActor ||
		g.actorItemCount(targetActor) >= g.pack.ItemActions().PersonalInventorySlots {
		t.Fatalf("給予起點錯：owner=%d idx=%d target=%d source=%v dest=%v",
			owner, idx, targetActor, source, dest)
	}
	code := (*source)[idx]
	sourceBefore := countOccurrences(*source, code)
	destBefore := countOccurrences(*dest, code)
	step := func(in InputState) {
		t.Helper()
		if in.DirHeld == 0 {
			in.DirHeld = -1
		}
		if in.DirEdge == 0 && in.Confirm {
			in.DirEdge = -1
		}
		if err := g.step(in); err != nil {
			t.Fatalf("給予 production input: %v", err)
		}
	}
	step(InputState{Confirm: true})
	step(InputState{DirEdge: 3})
	step(InputState{DirEdge: 0})
	step(InputState{Confirm: true})
	for i := 0; i < owner; i++ {
		step(InputState{DirEdge: 0})
	}
	step(InputState{Confirm: true}) // 選定實際持有者
	for i := 0; i < idx; i++ {
		step(InputState{DirEdge: 0})
	}
	step(InputState{Confirm: true}) // 選道具 → rec421 動作
	step(InputState{DirEdge: 0})    // 使用 → 給予
	step(InputState{Confirm: true})
	for i := 0; i < targetActor; i++ {
		step(InputState{DirEdge: 0})
	}
	step(InputState{Confirm: true})
	if g.itemActionStage != itemActionList {
		t.Fatalf("給予後未回道具清單：stage=%d", g.itemActionStage)
	}
	if countOccurrences(*g.equipActorInventory(owner), code) != sourceBefore-1 ||
		countOccurrences(*g.equipActorInventory(targetActor), code) != destBefore+1 {
		t.Fatalf("正式給予未完成 owner%d→target%d：source=%v dest=%v",
			owner, targetActor, *g.equipActorInventory(owner), *g.equipActorInventory(targetActor))
	}
	step(InputState{Cancel: true})
	step(InputState{Cancel: true})
}

// traceGiveInventoryItem 把隊伍中第一個實際持有 raw ID 的角色移交給指定同伴。
// 目標滿格時只丟棄其他角色已證實可丟的藥草，再用正式給予重分配一格。
func traceGiveInventoryItem(t *testing.T, g *Game, code, target int) {
	t.Helper()
	findOwner := func() (int, int) {
		for actor := 0; actor <= len(g.companions); actor++ {
			items := g.equipActorInventory(actor)
			if items == nil {
				continue
			}
			for idx, got := range *items {
				if got == code {
					return actor, idx
				}
			}
		}
		return -1, -1
	}
	owner, idx := findOwner()
	targetActor := target + 1
	if owner < 0 || target < 0 || target >= len(g.companions) {
		t.Fatalf("給予起點錯：code=%#x owner=%d idx=%d target=%d party=%d",
			code, owner, idx, target, len(g.companions))
	}
	if owner == targetActor {
		return
	}
	if g.actorItemCount(targetActor) >= g.pack.ItemActions().PersonalInventorySlots {
		freeActor := -1
		for actor := 0; actor <= len(g.companions); actor++ {
			if actor == targetActor {
				continue
			}
			items := g.equipActorInventory(actor)
			if items != nil && containsInt(*items, herbCode) {
				traceDropActorInventoryItem(t, g, actor, herbCode)
				freeActor = actor
				break
			}
		}
		targetItems := g.equipActorInventory(targetActor)
		if freeActor < 0 || targetItems == nil || len(*targetItems) == 0 {
			t.Fatalf("給予目標 actor%d 已滿，無法以已證實的藥草丟棄重分配：%v",
				targetActor, targetItems)
		}
		traceGiveActorInventoryItem(t, g, targetActor, 0, freeActor)
		owner, idx = findOwner()
		if owner < 0 || owner == targetActor {
			t.Fatalf("重分配後找不到待給予道具 %#x：owner=%d idx=%d", code, owner, idx)
		}
	}
	traceGiveActorInventoryItem(t, g, owner, idx, targetActor)
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
				need := g.cur.doorTier(x, y)
				keyCode, keyTier := -1, 4
				for _, effect := range g.pack.ItemUseEffects() {
					if effect.EffectID == "open_facing_locked_door" &&
						effect.DoorKeyTier >= need && effect.DoorKeyTier < keyTier &&
						g.hasPartyItem(effect.ItemRawID) {
						keyCode, keyTier = effect.ItemRawID, effect.DoorKeyTier
					}
				}
				if keyCode < 0 {
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
					traceUseInventoryItem(t, g, keyCode)
					if g.cur.doorTier(x, y) != 0 || !g.hasPartyItem(keyCode) {
						t.Fatalf("正式使用鑰匙未開門或錯誤消耗：doorTier=%d key=%#x held=%v",
							g.cur.doorTier(x, y), keyCode, g.hasItem(keyCode))
					}
					opened++
					t.Logf("正式道具使用開門 CTY%d sec%d door(%d,%d) key=%#x",
						g.curCty, sceneSection(g.cur), x, y, keyCode)
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
	t.Logf("正式地表導航：(%d,%d,L%d) → CTY%d，強敵逃跑=%v",
		g.px, g.py, g.layer, wantCty, shouldFleeStrong)
	if wantCty < 0 || wantCty >= len(ctyLoc) || ctyLoc[wantCty][2] != g.layer {
		t.Fatalf("無效 CTY%d / layer%d", wantCty, g.layer)
	}
	// 城鎮邊界出口可能把玩家放回入口判定格本身；production 只在「踏入」
	// 該格時進城，因此先以正式方向鍵走離入口，再由尋路踏回。
	if !g.inTown && findCtyAtLayer(g.px, g.py, g.layer) == wantCty {
		for g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待離開 CTY%d 入口 cooldown：%v", wantCty, err)
			}
		}
		moved := false
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny := g.px+dx, g.py+dy
			if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
				g.cur.Blocked(nx, ny) || findCtyAtLayer(nx, ny, g.layer) >= 0 {
				continue
			}
			oldX, oldY := g.px, g.py
			if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
				t.Fatalf("離開 CTY%d 入口格：%v", wantCty, err)
			}
			moved = g.px != oldX || g.py != oldY
			break
		}
		if !moved {
			t.Fatalf("CTY%d 入口格 (%d,%d) 無法以正式移動走開", wantCty, g.px, g.py)
		}
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
		if traceCureHeroPoison(t, g) {
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
		oldX, oldY := g.px, g.py
		var walkableRura []int
		for _, visit := range g.visitedTowns {
			if visit.Cty < 0 || visit.Cty >= len(ctyLoc) || ctyLoc[visit.Cty][2] != g.layer {
				continue
			}
			g.px, g.py = ctyLoc[visit.Cty][0]-1, ctyLoc[visit.Cty][1]
			if len(traceWorldPath(g, lx, ly, wantCty)) > 0 {
				walkableRura = append(walkableRura, visit.Cty)
			}
		}
		g.px, g.py = oldX, oldY
		rawPath := tracePath(g.cur, g.px, g.py, lx, ly)
		firstCty, firstX, firstY := -1, -1, -1
		x, y := g.px, g.py
		for _, dir := range rawPath {
			dx, dy := dirDelta(dir)
			x, y = x+dx, y+dy
			if cty := findCtyAtLayer(x, y, g.layer); cty >= 0 && cty != wantCty {
				firstCty, firstX, firstY = cty, x, y
				break
			}
		}
		t.Fatalf("無法由地表抵達 CTY%d；目前 town=%v cty=%d @(%d,%d)，入口 (%d,%d) blocked=%v / (%d,%d) blocked=%v，不避開其他入口的路徑長度=%d、首個衝突入口=CTY%d@(%d,%d)、已造訪且可徒步抵達的魯拉點=%v",
			wantCty, g.inTown, g.curCty, g.px, g.py,
			lx, ly, g.cur.Blocked(lx, ly), lx+1, ly, g.cur.Blocked(lx+1, ly),
			len(rawPath), firstCty, firstX, firstY, walkableRura)
	}
}

// traceAdventureTravelToCty 規劃一段最多一次靠岸的船陸混合路徑；只把規劃出的方向送給
// production Game.step。可選第一個旗標表示高危遭遇採逃跑，第二個旗標才允許啟用
// 驅敵策略（聖水優先、無水時特黑洛斯），第三個旗標則讓已補滿 MP 的特定遠洋段優先
// 使用特黑洛斯以保存後續陸路的聖水；避免把有限補給或 MP 策略暗中套到所有航段。
// 船的登船、航行、靠岸與 CTY 入口判定皆由 tryMove/step 決定。
func traceMemberHasPhoenixOrb(member *Member) bool {
	if member == nil {
		return false
	}
	for code := itemGreenOrb; code <= itemSilverOrb; code++ {
		if containsInt(member.Inventory, code) {
			return true
		}
	}
	return false
}

func traceCompanionWithoutPhoenixOrb(g *Game) int {
	if g == nil {
		return -1
	}
	for i, member := range g.companions {
		if !traceMemberHasPhoenixOrb(member) {
			return i
		}
	}
	return -1
}

func TestTraceCompanionWithoutPhoenixOrb(t *testing.T) {
	g := &Game{companions: []*Member{
		{Inventory: []int{0x1e, itemGreenOrb}},
		{Inventory: []int{0x1f}},
		{Inventory: []int{itemGreenOrb + 2}},
	}}
	if got := traceCompanionWithoutPhoenixOrb(g); got != 1 {
		t.Fatalf("應跳過持綠／紅寶珠同伴並選 index1，got %d", got)
	}
	g.companions[1].Inventory = []int{itemGreenOrb + 1}
	if got := traceCompanionWithoutPhoenixOrb(g); got != -1 {
		t.Fatalf("全員持珠時必須失敗即關閉，got %d", got)
	}
}

func traceCanSafelyFinishEncounter(g *Game) bool {
	if g == nil || !g.battle.active {
		return false
	}
	alive := make([]*enemyUnit, 0, len(g.battle.enemies))
	for i := range g.battle.enemies {
		if g.battle.enemies[i].alive() {
			alive = append(alive, &g.battle.enemies[i])
		}
	}
	if len(alive) == 0 {
		return false
	}
	// 這是 deterministic trace 的保守玩家判斷，不是原版 AI：只有勇者
	// 對每隻存活敵都不慢、攻擊至少是敵人最大 HP 兩倍，且守備至少是敵攻兩倍，
	// 才以逐隻普通攻擊收尾。這也允許只會自療的低攻群體；任一敵不符合就
	// 整群維持逃跑策略。怪物 ID 不是強度排序，不參與判斷。
	for _, e := range alive {
		if g.battle.mons != nil && g.pack != nil {
			ai, ok := g.battle.mons.AI(e.monID)
			if ok && ai.CastProb != 0 {
				for bit := 0; bit < len(ai.SpellMask)*8; bit++ {
					if ai.SpellMask[bit/8]&(0x80>>uint(bit%8)) == 0 {
						continue
					}
					action, wired := g.battle.monsterActions[bit]
					condition, known := g.pack.ConditionDefinition(action.ConditionID)
					if wired && known && condition.Persistence == "persistent" {
						return false
					}
				}
			}
		}
		if g.battle.heroAgi < e.agi || g.battle.heroAtk < e.max*2 ||
			g.battle.heroDef < e.atk*2 {
			return false
		}
	}
	return true
}

func TestTraceCanSafelyFinishLowStatHighIDEncounter(t *testing.T) {
	g := &Game{battle: Battle{
		active: true, heroAtk: 151, heroDef: 124, heroAgi: 113,
		enemies: []enemyUnit{{monID: 125, hp: 54, max: 54, atk: 30, def: 51, agi: 50}},
	}}
	if !traceCanSafelyFinishEncounter(g) {
		t.Fatal("單隻低 HP／低攻／較慢敵人應可由正式普通攻擊安全收尾")
	}
	g.battle.heroAtk, g.battle.heroDef, g.battle.heroAgi = 155, 127, 113
	g.battle.enemies = make([]enemyUnit, 5)
	for i := range g.battle.enemies {
		g.battle.enemies[i] = enemyUnit{monID: 4, hp: 50, max: 50, atk: 18, def: 45, agi: 50}
	}
	if !traceCanSafelyFinishEncounter(g) {
		t.Fatal("五隻只會自療的低攻貝荷瑪史萊姆應可由正式普通攻擊逐隻收尾")
	}
	g.battle.enemies[4].atk = 80
	if traceCanSafelyFinishEncounter(g) {
		t.Fatal("任一敵人攻擊超過安全界線時不可放行整群")
	}
}

func traceAdventureTravelToCty(t *testing.T, g *Game, wantCty int, fleeAll ...bool) {
	t.Helper()
	if wantCty < 0 || wantCty >= len(ctyLoc) || ctyLoc[wantCty][2] != g.layer {
		t.Fatalf("無效船陸目的 CTY%d / layer%d", wantCty, g.layer)
	}
	for i := 0; i < 20000 && (!g.inTown || g.curCty != wantCty); i++ {
		if g.battle.active {
			// 中後期遠洋可能抽到遠超目前隊伍承受力的高階區域怪；正常玩家會選逃跑，
			// 不能讓 campaign trace 依賴剛好沒抽到怪87等幸運 RNG。反之，若整群敵人
			// 都可由目前勇者安全逐隻擊倒，就正式普攻收尾；不可把 raw monster ID
			// 當成戰力排序，否則 monster125 這類低階追加怪會因反覆逃跑磨死隊伍。
			preserveMP := len(fleeAll) > 0 && fleeAll[0]
			// 這裡的安全判斷本來只在 true 時覆寫 raw-ID heuristic，導致
			// 「判定不安全」的低 ID 遭遇仍被強迫戰鬥。直接以保守判斷
			// 決定逃跑；怪物 ID 不再冒充戰力或狀態風險證據。
			shouldFlee := !traceCanSafelyFinishEncounter(g)
			traceResolveBattle(t, g, shouldFlee, preserveMP)
			continue
		}
		// 高危船陸路段沿用下層長途路線的正式驅敵策略：效果將盡時，
		// 優先由命令窗使用背包聖水；沒有物品則由已習得且 MP 足夠的
		// 特黑洛斯接續。兩者都不可用時才以正式逃跑作為 fallback。
		// 不修改遭遇計數器或 RNG，也不把「最後一瓶保留」當成讓四人隊
		// 賭高階遭遇的理由；後續固定路段會在可補給城鎮正式補足。
		explicitRepel := len(fleeAll) > 1 && fleeAll[1]
		if g.repel <= 8 && (g.countPartyItem(itemuse.ItemHolyWater) > 1 || explicitRepel) {
			castToheros := func() bool {
				caster := g.fieldSpellCaster(fieldToheros)
				reserveRuraMP := 0
				if caster == 0 {
					reserveRuraMP = spell.MPCost(fieldRura)
				}
				if caster < 0 || g.fieldCasterMP(caster) < spell.MPCost(fieldToheros)+reserveRuraMP {
					return false
				}
				traceUseFieldSpell(t, g, fieldToheros)
				return true
			}
			if len(fleeAll) > 2 && fleeAll[2] && castToheros() {
				continue
			}
			if g.hasPartyItem(itemuse.ItemHolyWater) {
				traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
				continue
			}
			if castToheros() {
				continue
			}
		}
		if g.inTown {
			t.Fatalf("前往 CTY%d 時誤入 CTY%d", wantCty, g.curCty)
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待船陸路徑 cooldown：%v", err)
			}
			continue
		}
		path := traceVehicleWorldPath(g, wantCty)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("船陸移動 dir%d：%v", path[0], err)
		}
	}
	if !g.inTown || g.curCty != wantCty {
		lx, ly := ctyLoc[wantCty][0], ctyLoc[wantCty][1]
		oldX, oldY, oldAboard := g.px, g.py, g.shipAboard
		var reachableShipRura []int
		for _, destination := range g.pack.RuraDestinations() {
			visited := false
			for _, visit := range g.visitedTowns {
				visited = visited || visit.Cty == destination.CTYRaw
			}
			if !visited || destination.ShipWorldPosition.Layer != g.layer {
				continue
			}
			g.px, g.py, g.shipAboard = destination.ShipWorldPosition.X,
				destination.ShipWorldPosition.Y, true
			if len(traceVehicleWorldPath(g, wantCty)) > 0 {
				reachableShipRura = append(reachableShipRura, destination.CTYRaw)
			}
		}
		g.px, g.py, g.shipAboard = oldX, oldY, oldAboard
		t.Fatalf("無法以正式航行／靠岸抵達 CTY%d；town=%v cty=%d @(%d,%d) aboard=%v ship=(%d,%d) entrance=(%d,%d) raw=%#x blocked=%v / (%d,%d) raw=%#x blocked=%v；已造訪且船區可達的魯拉點=%v",
			wantCty, g.inTown, g.curCty, g.px, g.py, g.shipAboard, g.shipX, g.shipY,
			lx, ly, g.cur.attr.Raw(g.cur.tileIdx(lx, ly)), g.cur.Blocked(lx, ly),
			lx+1, ly, g.cur.attr.Raw(g.cur.tileIdx(lx+1, ly)), g.cur.Blocked(lx+1, ly),
			reachableShipRura)
	}
}

// traceReenterWorldEntrance 處理多個 CTY 共用同一地表入口的 ordered variant。
// 出城會回到 remembered entrance tile；production 只在成功踏入地表格時判定進城，
// 因此先以方向鍵走離整個入口 footprint，再由正式船陸 helper 踏回。
func traceReenterWorldEntrance(t *testing.T, g *Game, wantCty int) {
	t.Helper()
	if g.inTown || wantCty < 0 || wantCty >= len(ctyLoc) {
		t.Fatalf("重進 CTY%d 起點錯：town=%v cty=%d", wantCty, g.inTown, g.curCty)
	}
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待離開共用入口 cooldown：%v", err)
		}
	}
	moved := false
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		nx, ny := g.px+dx, g.py+dy
		if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h ||
			g.cur.Blocked(nx, ny) || findCtyAtLayer(nx, ny, g.layer) >= 0 {
			continue
		}
		oldX, oldY := g.px, g.py
		if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
			t.Fatalf("走離共用入口：%v", err)
		}
		moved = g.px != oldX || g.py != oldY
		if moved {
			break
		}
	}
	if !moved {
		t.Fatalf("CTY%d 共用入口 (%d,%d) 無法以正式移動走離", wantCty, g.px, g.py)
	}
	traceAdventureTravelToCty(t, g, wantCty, true)
}

func traceVehicleWorldPath(g *Game, wantCty int) []int {
	type node struct {
		x, y   int
		aboard bool
	}
	start := node{x: g.px, y: g.py, aboard: g.shipAboard}
	q := []node{start}
	seen := map[node]bool{start: true}
	prev := map[node]node{}
	prevDir := map[node]int{}
	var goal node
	found := false
	for len(q) > 0 && !found {
		p := q[0]
		q = q[1:]
		for dir := 0; dir < 4; dir++ {
			dx, dy := dirDelta(dir)
			nx, ny, valid := p.x+dx, p.y+dy, true
			if wantCty == 58 {
				nx, ny, valid = g.normalizeSurfaceTarget(nx, ny)
			} else {
				valid = nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h
			}
			n := node{x: nx, y: ny, aboard: p.aboard}
			if !valid {
				continue
			}
			if p.aboard {
				if g.cur.attr.Raw(g.cur.tileIdx(n.x, n.y))&0x0002 != 0 {
					if g.cur.Blocked(n.x, n.y) {
						continue
					}
					n.aboard = false
				}
			} else if g.cur.Blocked(n.x, n.y) {
				continue
			}
			if seen[n] {
				continue
			}
			cty := findCtyAtLayer(n.x, n.y, g.layer)
			if resolved := worldEntranceResolve(n.x, n.y, g.layer,
				g.pack.WorldEntranceVariants(), g.storyFlag); resolved >= 0 {
				cty = resolved
			}
			if cty >= 0 && cty != wantCty {
				continue
			}
			seen[n] = true
			prev[n], prevDir[n] = p, dir
			if cty == wantCty {
				goal, found = n, true
				break
			}
			q = append(q, n)
		}
	}
	if !found {
		return nil
	}
	var rev []int
	for cur := goal; cur != start; cur = prev[cur] {
		rev = append(rev, prevDir[cur])
	}
	path := make([]int, len(rev))
	for i := range rev {
		path[len(rev)-1-i] = rev[i]
	}
	return path
}

func traceExaminePackTreasure(t *testing.T, g *Game, tr gamepack.QuestTreasureSelector) {
	t.Helper()
	if g.cur == nil || g.curCty != tr.CTYRaw || sceneSection(g.cur) != tr.Section {
		t.Fatalf("寶箱場景錯：CTY%d sec%d，want CTY%d sec%d",
			g.curCty, sceneSection(g.cur), tr.CTYRaw, tr.Section)
	}
	type target struct {
		x, y, standX, standY, face int
		onTile                     bool
	}
	var found *target
	var routeTarget *target
	for y := 0; y < g.cur.h && found == nil; y++ {
		for x := 0; x < g.cur.w && found == nil; x++ {
			ev, subid, ok := g.cur.tileEvent(x, y)
			if !ok || subid != tr.TileSubID || ev[0] != tr.EventTypeRaw ||
				ev[1] != tr.ItemRawID || ev[2] != tr.PresentFlag {
				continue
			}
			if !g.cur.Blocked(x, y) &&
				(x == g.px && y == g.py || len(tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier())) > 0) {
				found = &target{x: x, y: y, standX: x, standY: y, onTile: true}
				break
			}
			if !g.cur.Blocked(x, y) && routeTarget == nil {
				routeTarget = &target{x: x, y: y, standX: x, standY: y, onTile: true}
			}
			for face := 0; face < 4; face++ {
				dx, dy := dirDelta(face)
				sx, sy := x-dx, y-dy
				if sx < 0 || sy < 0 || sx >= g.cur.w || sy >= g.cur.h ||
					g.cur.Blocked(sx, sy) {
					continue
				}
				if sx == g.px && sy == g.py ||
					len(tracePortalPath(g.cur, g.px, g.py, sx, sy, g.keyTier())) > 0 {
					found = &target{x: x, y: y, standX: sx, standY: sy, face: face}
					break
				}
				if routeTarget == nil {
					routeTarget = &target{x: x, y: y, standX: sx, standY: sy, face: face}
				}
			}
		}
	}
	if found == nil && routeTarget != nil {
		// 同一 section 可能有多個由不同樓梯進入的連通區（加爾那之塔 sec1
		// 即需先繞 sec2→sec3→sec4 再回 sec1）。要求 transition graph 尋到
		// 寶箱操作格所在的精確 component，不能因 section 編號相同就提早停止。
		traceTownSectionTo(t, g, tr.CTYRaw, tr.Section,
			routeTarget.standX, routeTarget.standY, 0)
		found = routeTarget
	}
	if found == nil {
		t.Fatalf("CTY%d sec%d 找不到可達的 pack treasure subid%d",
			tr.CTYRaw, tr.Section, tr.TileSubID)
	}
	traceWalkToNoPortal(t, g, found.standX, found.standY)
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatalf("寶箱 production input：%v", err)
		}
	}
	for g.cd > 0 {
		step(InputState{DirHeld: -1, DirEdge: -1})
	}
	if !found.onTile && g.facing != found.face {
		step(InputState{DirHeld: found.face, DirEdge: -1})
		for g.cd > 0 {
			step(InputState{DirHeld: -1, DirEdge: -1})
		}
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.cmd.open {
		t.Fatal("寶箱前未能開啟正式命令窗")
	}
	for _, dir := range []int{3, 0, 0} {
		if g.cmd.cursor == int(cmdExamine) {
			break
		}
		step(InputState{DirHeld: -1, DirEdge: dir})
	}
	if g.cmd.cursor != int(cmdExamine) {
		t.Fatalf("寶箱前無法選到調查：cursor=%d", g.cmd.cursor)
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
}

// traceExamineCurrent sends the production command-window inputs for a
// location-wide examine event (for example CTY93 rainbow synthesis). It does
// not call examine directly or inject the event state.
func traceExamineCurrent(t *testing.T, g *Game) {
	t.Helper()
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatalf("調查 production input：%v", err)
		}
	}
	for g.cd > 0 {
		step(InputState{DirHeld: -1, DirEdge: -1})
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	for _, dir := range []int{3, 0, 0} {
		if g.cmd.cursor == int(cmdExamine) {
			break
		}
		step(InputState{DirHeld: -1, DirEdge: dir})
	}
	if g.cmd.cursor != int(cmdExamine) {
		t.Fatalf("目前場景無法選到調查：cursor=%d", g.cmd.cursor)
	}
	step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
}

// traceBoardAndSailShip 從正式地表位置尋路到 game-pack 指定停泊船旁，送一次正常方向
// 輸入上船，再航行一格水域。它不修改船、玩家或地圖狀態。
func traceBoardAndSailShip(t *testing.T, g *Game) {
	t.Helper()
	if g.inTown || !g.shipOwned || g.shipAboard {
		t.Fatalf("首次上船前狀態錯：town=%v owned=%v aboard=%v",
			g.inTown, g.shipOwned, g.shipAboard)
	}
	type approach struct {
		x, y int
		dir  int
		path []int
	}
	var best *approach
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		a := &approach{x: g.shipX - dx, y: g.shipY - dy, dir: dir}
		if a.x < 0 || a.y < 0 || a.x >= g.cur.w || a.y >= g.cur.h ||
			g.cur.Blocked(a.x, a.y) {
			continue
		}
		if g.px != a.x || g.py != a.y {
			a.path = traceWorldPath(g, a.x, a.y, -1)
			if len(a.path) == 0 {
				continue
			}
		}
		if best == nil || len(a.path) < len(best.path) {
			best = a
		}
	}
	if best == nil {
		t.Fatalf("停泊船 (%d,%d) 沒有可由玩家抵達的相鄰陸格；目前 pos=(%d,%d) layer=%d cty=%d", g.shipX, g.shipY, g.px, g.py, g.layer, g.curCty)
	}
	for i := 0; i < 3000 && (g.px != best.x || g.py != best.y); i++ {
		if g.battle.active {
			traceResolveBattle(t, g, false)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待上船路徑 cooldown：%v", err)
			}
			continue
		}
		path := traceWorldPath(g, best.x, best.y, -1)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("走向停泊船 dir%d：%v", path[0], err)
		}
	}
	if g.px != best.x || g.py != best.y {
		t.Fatalf("無法走到停泊船相鄰格 (%d,%d)，目前 (%d,%d)",
			best.x, best.y, g.px, g.py)
	}
	for g.cd > 0 || g.battle.active {
		if g.battle.active {
			traceResolveBattle(t, g, false)
			continue
		}
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待正式上船 cooldown：%v", err)
		}
	}
	if err := g.step(InputState{DirHeld: best.dir, DirEdge: -1}); err != nil {
		t.Fatalf("正式上船 dir%d：%v", best.dir, err)
	}
	if !g.shipAboard || g.px != g.shipX || g.py != g.shipY {
		t.Fatalf("走上船格後未登船：aboard=%v player=(%d,%d) ship=(%d,%d)",
			g.shipAboard, g.px, g.py, g.shipX, g.shipY)
	}

	// 登船本身也是一個地表移動；若正好耗盡遭遇計數器，原版會先開
	// 戰鬥 modal，直到正式戰鬥輸入結束前 cd 不會遞減。不能只空等
	// cooldown，否則追跡會把可正常處理的遭遇誤判成卡死。
	for waits := 0; g.cd > 0 || g.battle.active; waits++ {
		if waits >= 8192 {
			t.Fatalf("登船後 cooldown／戰鬥未收斂：cd=%d battle=%v pos=(%d,%d) ship=(%d,%d)",
				g.cd, g.battle.active, g.px, g.py, g.shipX, g.shipY)
		}
		if g.battle.active {
			traceResolveBattle(t, g, false)
			continue
		}
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待航行 cooldown：%v", err)
		}
	}
	startX, startY := g.px, g.py
	sailDir := -1
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		nx, ny := startX+dx, startY+dy
		if nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h &&
			g.cur.attr.Raw(g.cur.tileIdx(nx, ny))&0x20 != 0 {
			sailDir = dir
			break
		}
	}
	if sailDir < 0 {
		t.Fatalf("船停泊格 (%d,%d) 周圍沒有可航水格", startX, startY)
	}
	if err := g.step(InputState{DirHeld: sailDir, DirEdge: -1}); err != nil {
		t.Fatalf("首次正式航行 dir%d：%v", sailDir, err)
	}
	if !g.shipAboard || g.px == startX && g.py == startY {
		t.Fatalf("首次航行未移動：aboard=%v from=(%d,%d) to=(%d,%d)",
			g.shipAboard, startX, startY, g.px, g.py)
	}
}

// traceShipToWorldTile 只在原始水域上規劃船路，並把方向逐步送入 production
// Game.step。它避開所有 CTY 入口，讓位置式道具一定由玩家在指定格主動使用。
func traceShipToWorldTile(t *testing.T, g *Game, tx, ty int, fleeStrong ...bool) {
	t.Helper()
	if g.inTown || !g.shipAboard {
		t.Fatalf("航向位置式道具格起點錯：town=%v aboard=%v", g.inTown, g.shipAboard)
	}
	for steps := 0; steps < 12000 && (g.px != tx || g.py != ty); steps++ {
		if g.battle.active {
			preserveMP := len(fleeStrong) > 0 && fleeStrong[0]
			traceResolveBattle(t, g, preserveMP || g.battle.monID >= 80, preserveMP)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待船路 cooldown：%v", err)
			}
			continue
		}
		path := traceShipWaterPath(g, tx, ty)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("正式船路 dir%d：%v", path[0], err)
		}
	}
	if g.inTown || !g.shipAboard || g.px != tx || g.py != ty {
		t.Fatalf("無法正式航行到 (%d,%d)：town=%v aboard=%v pos=(%d,%d)",
			tx, ty, g.inTown, g.shipAboard, g.px, g.py)
	}
}

// traceSailIntoTrackedWorldObject 以正式船移動逐格抵達 pack 動態入口；最後一步
// 必須由 production Update 自動切入目的 CTY，不能直接呼叫事件函式或注入座標。
func traceSailIntoTrackedWorldObject(t *testing.T, g *Game, obj *gamepack.TrackedWorldObject) {
	t.Helper()
	p, active := g.trackedWorldObjectPosition(obj)
	if !active || g.inTown || !g.shipAboard || p.Layer != g.layer {
		t.Fatalf("航向動態世界物件起點錯：active=%v town=%v aboard=%v layer=%d/%d",
			active, g.inTown, g.shipAboard, g.layer, p.Layer)
	}
	for steps := 0; steps < 12000 && !g.inTown; steps++ {
		if g.battle.active {
			// 這段只是航路；保留魯拉所需 MP，與玩家可選擇全隊
			// 普攻／逃跑的正式戰鬥輸入一致。
			traceResolveBattle(t, g, g.battle.monID >= 80, true)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待幽靈船航路 cooldown：%v", err)
			}
			continue
		}
		var stillActive bool
		p, stillActive = g.trackedWorldObjectPosition(obj)
		if !stillActive {
			t.Fatal("航行途中動態世界物件失效")
		}
		path := traceShipWaterPath(g, p.X, p.Y)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("幽靈船正式航路 dir%d：%v", path[0], err)
		}
	}
	p, _ = g.trackedWorldObjectPosition(obj)
	if !g.inTown || g.curCty != obj.EntranceCTYRaw || sceneSection(g.cur) != obj.EntranceSection {
		// 若魯拉重定位過船，列出已造訪港口中實際可航向動態物件者；這是
		// 路由診斷，不能把「可施放魯拉」誤當成「同一片海域可到達」。
		oldX, oldY, oldAboard := g.px, g.py, g.shipAboard
		var reachableRuraPorts []int
		for _, destination := range g.pack.RuraDestinations() {
			visited := false
			for _, visit := range g.visitedTowns {
				visited = visited || visit.Cty == destination.CTYRaw
			}
			if !visited || destination.ShipWorldPosition.Layer != g.layer {
				continue
			}
			g.px, g.py, g.shipAboard = destination.ShipWorldPosition.X,
				destination.ShipWorldPosition.Y, true
			if len(traceShipWaterPath(g, p.X, p.Y)) > 0 {
				reachableRuraPorts = append(reachableRuraPorts, destination.CTYRaw)
			}
		}
		g.px, g.py, g.shipAboard = oldX, oldY, oldAboard
		t.Fatalf("未由動態座標進入 CTY%d：town=%v cty=%d sec=%d player=(%d,%d) object=(%d,%d) reachableRuraPorts=%v",
			obj.EntranceCTYRaw, g.inTown, g.curCty, sceneSection(g.cur), g.px, g.py, p.X, p.Y, reachableRuraPorts)
	}
}

func traceShipWaterPath(g *Game, tx, ty int) []int {
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
				g.cur.attr.Raw(g.cur.tileIdx(n.x, n.y))&0x0002 != 0 &&
					(n != goal || g.cur.Blocked(n.x, n.y)) {
				continue
			}
			cty := findCtyAtLayer(n.x, n.y, g.layer)
			if resolved := worldEntranceResolve(n.x, n.y, g.layer,
				g.pack.WorldEntranceVariants(), g.storyFlag); resolved >= 0 {
				cty = resolved
			}
			if cty >= 0 && n != goal {
				continue
			}
			if n != goal && g.pack != nil {
				tracked := false
				for _, obj := range g.pack.TrackedWorldObjects() {
					p, active := g.trackedWorldObjectPosition(&obj)
					if active && p.Layer == g.layer && p.X == n.x && p.Y == n.y {
						tracked = true
						break
					}
				}
				if tracked {
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

// traceSailToWorldCoordinate 只在原始水域以正式方向鍵前往指定格。
// 可選 protectWithRepel 只供明確的高危航段使用；它先用聖水，無水時才由
// 特黑洛斯接續，且若勇者施法會保留一發魯拉的 MP。
func traceSailToWorldCoordinate(t *testing.T, g *Game, tx, ty int, protectWithRepel ...bool) {
	t.Helper()
	for steps := 0; steps < 12000 && !g.inTown && (g.px != tx || g.py != ty); steps++ {
		if g.dlg.open {
			return
		}
		if g.battle.active {
			traceResolveBattle(t, g, g.battle.monID >= 80, true)
			continue
		}
		if len(protectWithRepel) > 0 && protectWithRepel[0] && g.repel <= 8 {
			if g.hasItem(itemuse.ItemHolyWater) {
				traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
				continue
			}
			caster := g.fieldSpellCaster(fieldToheros)
			reserveRuraMP := 0
			if caster == 0 {
				reserveRuraMP = spell.MPCost(fieldRura)
			}
			if caster >= 0 && g.fieldCasterMP(caster) >= spell.MPCost(fieldToheros)+reserveRuraMP {
				traceUseFieldSpell(t, g, fieldToheros)
				continue
			}
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待航行 cooldown：%v", err)
			}
			continue
		}
		path := traceShipWaterPath(g, tx, ty)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("航向 (%d,%d) dir%d：%v", tx, ty, path[0], err)
		}
	}
	// 抵達目標格的同一步仍可能耗盡原版遭遇計數器並開啟戰鬥。
	// 座標迴圈會因已到達而結束，但玩家必須先以正式戰鬥輸入結算，
	// 才能在該格開命令窗使用位置道具。
	if g.battle.active {
		traceResolveBattle(t, g, g.battle.monID >= 80, true)
	}
	if g.px != tx || g.py != ty {
		t.Fatalf("未由正式船路抵達 (%d,%d)：town=%v cty=%d sec=%d aboard=%v pos=(%d,%d)",
			tx, ty, g.inTown, g.curCty, sceneSection(g.cur), g.shipAboard, g.px, g.py)
	}
}

// traceSailAndDisembarkNearCty 由船正式航到目標城入口附近的水格，再以
// 一次正常方向鍵下船。水格、陸格與入口均從目前 layer 的原始地圖掃描，
// 不把版本專屬座標寫進終盤 trace；下船後由 traceAdventureWalkToCty
// 依 production world entrance resolver 進城。
func traceSailAndDisembarkNearCty(t *testing.T, g *Game, wantCty int) {
	t.Helper()
	if g.inTown || !g.shipAboard || g.cur == nil {
		t.Fatalf("靠岸前船狀態錯：town=%v aboard=%v cty=%d layer=%d",
			g.inTown, g.shipAboard, g.curCty, g.layer)
	}
	type landing struct {
		waterX, waterY int
		landDir        int
		score          int
	}
	var best *landing
	for ey := 0; ey < g.cur.h; ey++ {
		for ex := 0; ex < g.cur.w; ex++ {
			if findCtyAtLayer(ex, ey, g.layer) != wantCty {
				continue
			}
			ly0, ly1 := ey-3, ey+3
			if ly0 < 0 {
				ly0 = 0
			}
			if ly1 >= g.cur.h {
				ly1 = g.cur.h - 1
			}
			for ly := ly0; ly <= ly1; ly++ {
				lx0, lx1 := ex-3, ex+3
				if lx0 < 0 {
					lx0 = 0
				}
				if lx1 >= g.cur.w {
					lx1 = g.cur.w - 1
				}
				for lx := lx0; lx <= lx1; lx++ {
					if findCtyAtLayer(lx, ly, g.layer) >= 0 || g.cur.Blocked(lx, ly) ||
						g.cur.attr.Raw(g.cur.tileIdx(lx, ly))&0x0002 == 0 {
						continue
					}
					landPath := tracePath(g.cur, lx, ly, ex, ey)
					if lx != ex || ly != ey {
						if len(landPath) == 0 {
							continue
						}
					}
					for dir := 0; dir < 4; dir++ {
						dx, dy := dirDelta(dir)
						wx, wy := lx-dx, ly-dy
						if wx < 0 || wy < 0 || wx >= g.cur.w || wy >= g.cur.h ||
							g.cur.attr.Raw(g.cur.tileIdx(wx, wy))&0x0002 != 0 {
							continue
						}
						waterPath := traceShipWaterPath(g, wx, wy)
						if wx != g.px || wy != g.py {
							if len(waterPath) == 0 {
								continue
							}
						}
						score := len(waterPath) + len(landPath)
						if best == nil || score < best.score {
							best = &landing{waterX: wx, waterY: wy, landDir: dir, score: score}
						}
					}
				}
			}
		}
	}
	if best == nil {
		t.Fatalf("CTY%d 附近找不到可由目前船路抵達的水格／下船陸格", wantCty)
	}
	traceSailToWorldCoordinate(t, g, best.waterX, best.waterY)
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("CTY%d 靠岸等待 cooldown：%v", wantCty, err)
		}
	}
	if err := g.step(InputState{DirHeld: best.landDir, DirEdge: -1}); err != nil {
		t.Fatalf("CTY%d 正式下船：%v", wantCty, err)
	}
	if g.shipAboard {
		t.Fatalf("CTY%d 正式下船後仍在船上：pos=(%d,%d)", wantCty, g.px, g.py)
	}
}

// traceWalkToWorldEntrance 走到指定 world 座標並讓 production movement 消費入口；
// traceWorldPath 會把其他 CTY 入口視為障礙，只允許 wantCty 的最後一格。
func traceWalkToWorldEntrance(t *testing.T, g *Game, tx, ty, wantCty int) {
	t.Helper()
	if g.inTown || g.cur == nil || g.layer != 0 {
		t.Fatalf("地表入口尋路起點錯：town=%v scene=%v layer=%d", g.inTown, g.cur != nil, g.layer)
	}
	for steps := 0; steps < 20000 && (!g.inTown || g.curCty != wantCty); steps++ {
		if g.battle.active {
			traceResolveBattle(t, g, g.battle.monID >= 80, true)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待地表入口 cooldown：%v", err)
			}
			continue
		}
		path := traceWorldPath(g, tx, ty, wantCty)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("走向 CTY%d 入口 dir%d：%v", wantCty, path[0], err)
		}
	}
	if !g.inTown || g.curCty != wantCty {
		t.Fatalf("未由正式地表輸入進 CTY%d：town=%v cty=%d pos=(%d,%d) target=(%d,%d)",
			wantCty, g.inTown, g.curCty, g.px, g.py, tx, ty)
	}
}

// traceGaiaSwordToNirokenta 閉合原版「火山→尼羅肯特洞窟→北出口」的有限
// transition graph。所有座標、section 與 portal 均來自 CTY56/57.DAT；
// 不呼叫事件函式、不注入位置，也不把 CTY64 當成可直航的地表目標。
func traceGaiaSwordToNirokenta(t *testing.T, g *Game) {
	t.Helper()
	if g.inTown || g.layer != 0 {
		t.Fatalf("蓋亞後洞窟路徑起點錯：town=%v layer=%d", g.inTown, g.layer)
	}
	traceWalkToWorldEntrance(t, g, 43, 140, 56)
	if sceneSection(g.cur) != 0 {
		t.Fatalf("尼羅肯特南洞口 section=%d", sceneSection(g.cur))
	}
	traceTownSectionToPreservingRura(t, g, 56, 1)
	traceTownSectionToPreservingRura(t, g, 56, 2)
	// 原版 section graph 不是單向直線：先由 sec2 的 (27,34) 到 sec3
	// (10,20)，再回 sec2 的 (40,38)，最後由 (45,32) 進 sec3 北側
	// (23,3)。若省略中間回返，sec0 只會落在南洞口。
	traceTownSectionToPreservingRura(t, g, 56, 3)
	traceWalkThroughPortalPreservingRura(t, g, 33, 16, 56, 2, 40, 38)
	traceWalkThroughPortalPreservingRura(t, g, 45, 32, 56, 3, 23, 3)
	traceWalkThroughPortalPreservingRura(t, g, 37, 20, 56, 0, 3, 2)
	// sec0 的 (3,2) 與北出口 (64,3) 屬同一原始連通區；精確指定 portal，
	// 避免 traceTransitionToOverworld 選回南洞口。
	traceWalkThroughPortalPreservingRura(t, g, 64, 3, -1, 0)
	if g.inTown || g.px != 43 || g.py != 138 {
		t.Fatalf("尼羅肯特北洞口出口錯：town=%v pos=(%d,%d)", g.inTown, g.px, g.py)
	}
	traceWalkToWorldEntrance(t, g, ctyLoc[64][0], ctyLoc[64][1], 64)
}

// traceWalkToWorldCoordinate 在地表只送正式方向鍵走到指定陸地座標；
// random encounter 仍由 production battle input 完成，不改寫座標或 RNG。
func traceWalkToWorldCoordinate(t *testing.T, g *Game, tx, ty int) {
	t.Helper()
	if g.inTown || g.cur == nil || g.layer != 0 {
		t.Fatalf("地表步行到座標起點錯：town=%v scene=%v layer=%d", g.inTown, g.cur != nil, g.layer)
	}
	for steps := 0; steps < 20000 && (g.px != tx || g.py != ty); steps++ {
		if g.battle.active {
			traceResolveBattle(t, g, g.battle.monID >= 80, true)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待地表步行 cooldown：%v", err)
			}
			continue
		}
		path := traceWorldPath(g, tx, ty, -1)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("地表步行到 (%d,%d) dir%d：%v", tx, ty, path[0], err)
		}
	}
	if g.inTown || g.px != tx || g.py != ty {
		t.Fatalf("未由正式地表步行抵達 (%d,%d)：town=%v cty=%d aboard=%v pos=(%d,%d)",
			tx, ty, g.inTown, g.curCty, g.shipAboard, g.px, g.py)
	}
}

// traceWalkLowerWorldCoordinate 是下層地表的正式方向鍵尋路；與上層
// traceWalkToWorldCoordinate 分開，避免把 layer1 座標誤套上層環世界邊界。
func traceWalkLowerWorldCoordinate(t *testing.T, g *Game, tx, ty int) {
	t.Helper()
	if g.inTown || g.layer != 1 || g.shipAboard || g.cur == nil {
		t.Fatalf("下層徒步起點錯：town=%v layer=%d aboard=%v cur=%v",
			g.inTown, g.layer, g.shipAboard, g.cur != nil)
	}
	for steps := 0; steps < 20000 && (g.px != tx || g.py != ty); steps++ {
		// 正式輸入重施驅魔水：只在效果即將到期時使用背包物品，
		// 避免下層長距離航路被隨機遭遇戰改變終盤基線。
		if g.repel <= 8 && g.hasItem(itemuse.ItemHolyWater) {
			traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
		}
		if g.battle.active {
			traceResolveBattle(t, g, true, true)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待下層徒步 cooldown：%v", err)
			}
			continue
		}
		path := traceWorldPath(g, tx, ty, -1)
		if len(path) == 0 {
			break
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("下層徒步到 (%d,%d) dir%d：%v", tx, ty, path[0], err)
		}
	}
	if g.inTown || g.layer != 1 || g.shipAboard || g.px != tx || g.py != ty {
		t.Fatalf("未由正式下層徒步抵達 (%d,%d)：town=%v layer=%d aboard=%v cty=%d pos=(%d,%d)",
			tx, ty, g.inTown, g.layer, g.shipAboard, g.curCty, g.px, g.py)
	}
}

func traceSailToTown(t *testing.T, g *Game, cty int, refreshRepel ...bool) {
	t.Helper()
	if cty < 0 || cty >= len(ctyLoc) {
		t.Fatalf("無效航行目的 CTY%d", cty)
	}
	tx, ty := ctyLoc[cty][0], ctyLoc[cty][1]
	for steps := 0; steps < 12000 && !g.inTown; steps++ {
		if g.battle.active {
			// 這條航路的目的只是到達下一個劇情入口；保留魯拉與
			// 特黑洛斯所需 MP，遭遇仍經正式逃跑／普通攻擊輸入結束。
			if len(refreshRepel) == 1 && refreshRepel[0] {
				t.Logf("CTY%d 長航路遭遇：monster=%d count=%d heroHP=%d repel=%d",
					cty, g.battle.monID, len(g.battle.enemies), g.heroHP, g.repel)
			}
			traceResolveBattle(t, g, g.battle.monID >= 80, true)
			continue
		}
		if g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("等待 CTY%d 航行 cooldown：%v", cty, err)
			}
			continue
		}
		if len(refreshRepel) == 1 && refreshRepel[0] && g.repel <= 8 &&
			g.hasItem(itemuse.ItemHolyWater) {
			traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
			continue
		}
		var path []int
		if g.shipAboard {
			path = traceShipWaterPath(g, tx, ty)
		} else {
			path = traceWorldPath(g, tx, ty, cty)
		}
		if len(path) == 0 {
			// 海岬航道的事件格帶原版特殊 attr；一般 BFS 會保守拒絕，
			// 但 production tryMove 仍依 mode1 的 bit1/bit0 規則判斷。
			dir := -1
			switch {
			case tx < g.px:
				dir = 2
			case tx > g.px:
				dir = 3
			case ty < g.py:
				dir = 1
			case ty > g.py:
				dir = 0
			}
			if dir < 0 {
				break
			}
			path = []int{dir}
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("航向 CTY%d dir%d：%v", cty, path[0], err)
		}
	}
	if !g.inTown || g.curCty != cty {
		t.Fatalf("未由正式海岬路徑進 CTY%d target=(%d,%d)：town=%v cty=%d aboard=%v pos=(%d,%d) dlg=%v defeat=%d heroHP=%d attrW=%#x attrHere=%#x attrE=%#x targetAttr=%#x targetCty=%d",
			cty, tx, ty, g.inTown, g.curCty, g.shipAboard, g.px, g.py,
			g.dlg.open, g.defeatDialogueStage, g.heroHP,
			g.cur.attr.Raw(g.cur.tileIdx(g.px-1, g.py)), g.cur.attr.Raw(g.cur.tileIdx(g.px, g.py)),
			g.cur.attr.Raw(g.cur.tileIdx(g.px+1, g.py)), g.cur.attr.Raw(g.cur.tileIdx(tx, ty)),
			findCtyAtLayer(tx, ty, g.layer))
	}
}

func traceTransitionToSection(t *testing.T, g *Game, cty, section int) {
	t.Helper()
	for y := 0; y < g.cur.h; y++ {
		for x := 0; x < g.cur.w; x++ {
			destCTY, destSec, _, _, ok := g.cur.tileTransition(x, y)
			if ok && destCTY == cty && destSec == section {
				traceWalkThroughPortal(t, g, x, y, cty, section)
				return
			}
		}
	}
	t.Fatalf("CTY%d sec%d 找不到通往 sec%d 的原始 transition", g.curCty, sceneSection(g.cur), section)
}

func traceTransitionToOverworld(t *testing.T, g *Game) {
	t.Helper()
	for y := 0; y < g.cur.h; y++ {
		for x := 0; x < g.cur.w; x++ {
			_, destSec, _, _, ok := g.cur.tileTransition(x, y)
			if ok && destSec == 0xff &&
				(g.px == x && g.py == y || len(tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier())) > 0) {
				traceWalkThroughPortal(t, g, x, y, -1, 0)
				return
			}
		}
	}
	t.Fatalf("CTY%d sec%d 找不到原始地表出口", g.curCty, sceneSection(g.cur))
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
			if !traceWorldTileAllowed(g, n.x, n.y, wantCty) {
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

// traceWorldTileAllowed mirrors Game.step's pack-owned ordered entrance
// resolver.  Shared footprints such as CTY71/72 must be judged by the active
// story flag, not by the first numeric ctyLoc entry.
func traceWorldTileAllowed(g *Game, x, y, wantCty int) bool {
	if resolved := worldEntranceResolve(x, y, g.layer, g.pack.WorldEntranceVariants(), g.storyFlag); resolved >= 0 {
		return resolved == wantCty
	}
	if cty := findCtyAtLayer(x, y, g.layer); cty >= 0 {
		return cty == wantCty
	}
	return true
}

// dumpProductionOrochiBattlePhase 只在既有的完整 production trace 抵達日邦格兩戰時
// 取樣；它不改寫隊伍、戰鬥、RNG 或輸入。輸出目錄必須在 gitignore 的暫存區，供原版
// 同 phase 畫面核對，不能當成 debug shortcut 或發佈素材。
func dumpProductionOrochiBattlePhase(t *testing.T, g *Game, captured map[string]bool) {
	t.Helper()
	out := os.Getenv("DQ3_PRODUCTION_DUMP_OROCHI")
	if out == "" || g == nil || !g.battle.active {
		return
	}

	stage := ""
	switch g.stagedBossStage {
	case stagedBossFirstBattle:
		stage = "orochi_first"
	case stagedBossSecondBattle:
		stage = "orochi_second"
	default:
		return
	}
	phase := ""
	switch g.battle.phase {
	case phCommand:
		phase = "command"
	case phMessage:
		phase = "message"
	case phEnd:
		phase = "end"
	default:
		return
	}
	name := stage + "_" + phase
	if captured[name] {
		return
	}
	captured[name] = true

	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatalf("建立日邦格 production dump 目錄：%v", err)
	}
	// 主線 trace 平時刻意不配置 Ebiten frame，避免數千步輸入累積圖形命令；只有
	// 這個已經由正式輸入抵達的瞬間暫時配置一張畫布，隨即還原原狀。
	previousFrame := g.frame
	if previousFrame == nil {
		g.frame = ebiten.NewImage(ScreenW, ScreenH)
	}
	defer func() { g.frame = previousFrame }()
	// 正式遊戲在每個 Update 之間會持續 Draw；但這條高效率 trace 故意只重播
	// InputState，若剛好停在受擊後便會把尚未過期的純紅 flash 當成靜態畫面。
	// 不改寫命令、HP、RNG 或輸入；只消耗 renderer 自己的暫態 flash 計數，補足
	// 可見幀後取樣玩家按下下一個按鍵前會看到的 settled phase，而不是 trace 的
	// 繪圖缺口。
	// 以取樣瞬間的值界定次數；若未來 renderer 因前置條件跳過 draw，也不能讓
	// 證據 hook 卡在無界迴圈。
	for frames := g.battle.flashCol; frames > 0; frames-- {
		g.renderFrame()
	}
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	copy(img.Pix, g.rgba)
	path := filepath.Join(out, name+".png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("建立日邦格 production PNG：%v", err)
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		t.Fatalf("寫入日邦格 production PNG：%v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("關閉日邦格 production PNG：%v", err)
	}
	t.Logf("日邦格正式 trace %s：%s", name, path)
}

func traceResolveBattle(t *testing.T, g *Game, fleeStrong ...bool) {
	t.Helper()
	shouldFleeStrong := len(fleeStrong) == 0 || fleeStrong[0]
	preserveMP := len(fleeStrong) > 1 && fleeStrong[1]
	// 0x79 巴拉摩斯與 0x7a..0x7c 下層三連戰都使用終盤正式補給／
	// 封咒／回復策略；舊下界 0x7a 使下面明寫的巴拉摩斯分支永遠不可達。
	finalBoss := g.battle.monID >= 0x79 && g.battle.monID <= 0x7c
	usableHerbs := g.battle.heroHerbs
	// 一般遭遇先用回復咒文；MP 已耗盡時再使用實際攜帶的藥草。先前把一般
	// 遭遇的 usableHerbs 強制歸零，會在原版 bit3 致命攻擊恢復後讓正式路線
	// 明明帶著大量藥草卻拒絕治療，這不是正常玩家策略，也不能作完成證據。
	healTarget := -1
	wantedSpell := -1
	sealAttempted := false
	attackBuffAttempted := false
	blindAttempted := false
	defenseCasts := 0
	strategyBossID := g.battle.monID
	// 巴拉摩斯戰後必須經 CTY65→66 的原始 transition 與邊界出口；兩個
	// section 的烈米特 gate 都關閉。該路由含多個獨立傷害地板區段，因此
	// 為多拉瑪那施法者保留兩次成本，避免在最後一段才因少 1 MP 失敗。
	reservedHazardCaster := -1
	if g.battle.monID == 0x79 {
		reservedHazardCaster = g.fieldSpellCaster(fieldToramana)
	}
	availableMP := func(actor int) int {
		mp := g.battle.actorMP(actor)
		if g.battle.monID == 0x79 && actor == reservedHazardCaster {
			mp -= 2 * spell.MPCost(fieldToramana)
		}
		if mp < 0 {
			return 0
		}
		return mp
	}
	bestSpell := func(actor int, kind spell.Kind) int {
		best, bestBase := -1, -1
		for _, rec := range g.battle.actorSpells(actor) {
			def, ok := spell.GetDef(rec)
			if ok && def.Kind == kind && def.MP <= availableMP(actor) && def.Base > bestBase {
				best, bestBase = rec, def.Base
			}
		}
		return best
	}
	bestHealSpell := func(actor int) int {
		return traceSelectHealSpell(g.battle.actorSpells(actor), availableMP(actor), false)
	}
	battleHealSpell := func(actor int) int {
		// 巴拉摩斯是獨立高傷 Boss；最低 MP 的低階回復無法覆蓋其每回合
		// 傷害。其他終盤連戰仍使用低耗策略，避免第一場耗盡 MP。
		strongest := g.battle.monID == 0x79
		return traceSelectHealSpell(g.battle.actorSpells(actor), availableMP(actor), strongest)
	}
	hasAffordableSpell := func(actor, rec int) bool {
		for _, known := range g.battle.actorSpells(actor) {
			if known == rec {
				def, ok := spell.GetDef(rec)
				return ok && def.MP <= availableMP(actor)
			}
		}
		return false
	}
	herbSlot := func(actor int) int {
		for i, item := range g.battle.actorItems(actor) {
			if item.rawID == herbCode && !item.equipped {
				return i
			}
		}
		return -1
	}
	// 舊預算 8192 已涵蓋沒有逐 cue wait 的長戰。docs/150、151 接入後，玩家物理 miss
	// 是最長單一結果：cue6／cue11 各 11 frame，再接 cue3 的 14 frame。因此以每次語意
	// 確認最大 1+36 frame 膨脹原預算；正常命中為 22+11，敵方命中為 11+11，均在此界內。
	// 仍保留硬上限，且不改角色／Boss 數值。
	const (
		legacyBattleInputBudget = 8192
		physicalCueWaitFrames   = 11 + 11 + 14
		maxInputs               = legacyBattleInputBudget * (1 + physicalCueWaitFrames)
	)
	capturedOrochi := map[string]bool{}
	for i := 0; i < maxInputs && g.battle.active; i++ {
		dumpProductionOrochiBattlePhase(t, g, capturedOrochi)
		if g.battle.monID != strategyBossID {
			// bossQueue 會在同一個正式 phEnd 輸入後立即開下一場；
			// 每一場都重新封咒／拜基魯多，不能把上一場的 trace
			// 暫態旗標帶進下一個敵人。
			strategyBossID = g.battle.monID
			sealAttempted = false
			attackBuffAttempted = false
			blindAttempted = false
			defenseCasts = 0
		}
		if g.battle.result == 2 {
			heroLv, _, _, _, _ := g.heroStats()
			comp := make([][4]int, len(g.battle.companions))
			for i, c := range g.battle.companions {
				comp[i] = [4]int{c.level, c.hp, c.maxHP, c.status}
			}
			t.Fatalf("正常前段路線發生全滅：mon=%d enemies=%v heroLv=%d hero=%d/%d MP=%d conditions=%#x battleStatus=%#x inventory=%v repel=%d holyWater=%d atk=%d def=%d companions(lv,hp,max,status)=%v",
				g.battle.monID, g.battle.enemies, heroLv, g.battle.heroHP,
				g.battle.heroMax, g.heroMP, g.heroConditions, g.battle.heroStatus, g.inventory, g.repel,
				g.countPartyItem(itemuse.ItemHolyWater), g.battle.heroAtk, g.battle.heroDef, comp)
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		switch g.battle.phase {
		case phCommand:
			menu := g.battle.commandMenu(g.battle.commandActor)
			healTarget, wantedSpell = -1, -1
			persistentTarget := -1
			for _, actor := range g.battle.aliveActorIndices() {
				if g.battle.actorStatus(actor)&statusPersistentParalysis != 0 {
					persistentTarget = actor
					break
				}
			}
			canCastHeal := false
			for _, actor := range g.battle.aliveActorIndices() {
				if battleHealSpell(actor) >= 0 {
					canCastHeal = true
					break
				}
			}
			if g.battle.usedHerbs < usableHerbs || (finalBoss && canCastHeal) {
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
			case persistentTarget >= 0 &&
				bestSpell(g.battle.commandActor, spell.CureStatus) >= 0:
				// 原版 battle handler 以 DX=0x10／record399 清持久麻痺；
				// 未失能且學會基阿里克／薩梅哈的角色，先救回隊友的命令權。
				healTarget = persistentTarget
				wantedSpell = bestSpell(g.battle.commandActor, spell.CureStatus)
				wantCommand = bcSpell
			case needHeal && !preserveMP && bestHealSpell(g.battle.commandActor) >= 0:
				// 逃跑可失敗；練級戰若每回合都不救治就重試，
				// 即使低階怪物也會把滿資源隊伍磨死。docs/154 要求
				// 先以正式咒文救治，下一回合再由首位送出逃跑。
				wantedSpell = bestHealSpell(g.battle.commandActor)
				wantCommand = bcSpell
			case needHeal && !preserveMP && herbSlot(g.battle.commandActor) >= 0:
				wantCommand = bcItem
			case needFlee:
				// 高危遭遇的正式玩家策略是先嘗試逃跑；若先治療，
				// 敵方會在逃跑結算前取得一輪傷害，可能把整隊打光。
				wantCommand = bcFlee
			case g.battle.monID == 0x79 && g.battle.commandActor == reservedHazardCaster:
				// 巴拉摩斯戰後的原始出口有多段高傷害地板；這位唯一
				// 多拉瑪那施法者採正式普通攻擊而不施咒，保留 MP 給
				// 玩家可見的場景路徑，而非在戰後補寫資源或延長咒文效果。
				wantCommand = bcWar
			case finalBoss && !sealAttempted && hasAffordableSpell(g.battle.commandActor, 156):
				// 巴拉摩斯、殭屍索瑪與索瑪都必須先完成原版封咒
				// 效果，否則其回復／特殊行動會讓戰鬥資源無法閉合。
				wantedSpell, wantCommand = 156, bcSpell
			case finalBoss && !attackBuffAttempted && g.battle.commandActor != 0 &&
				hasAffordableSpell(g.battle.commandActor, 151):
				// 賢者的正式拜基魯多先強化主角，讓終盤在高防敵人
				// 的低 MP 階段仍保有原版可見的物理傷害。
				healTarget = 0
				wantedSpell, wantCommand = 151, bcSpell
			case needHeal && finalBoss && g.battle.commandActor != 0 && battleHealSpell(g.battle.commandActor) >= 0:
				// 僧侶／賢者的正式荷依米只耗 3 MP；讓同伴先用自身
				// 咒文補血，保留隊伍藥草給主角與真正無 MP 的情況。
				wantedSpell = battleHealSpell(g.battle.commandActor)
				wantCommand = bcSpell
			case needHeal && finalBoss && g.battle.commandActor != 0 && canCastHeal:
				// 沒有補血咒文的魔法使者不要搶用共享藥草；讓有
				// 荷依米的同伴在本回合處理最低 HP 目標。
				wantCommand = bcWar
			case needHeal && finalBoss && herbSlot(g.battle.commandActor) >= 0:
				// 正式隊伍藥草是可跨戰保留的資源；終盤先用道具，
				// 將 MP 留給封咒與後續敵人。
				wantCommand = bcItem
			case preserveMP:
				// 逃跑指令在全隊命令輸入完後才結算；此模式連補血也不選，
				// 所有成員只普攻，避免航路 trace 在等待逃跑時耗盡 MP。
				wantCommand = bcWar
			case g.bossSurrenderStage == bossSurrenderBattle &&
				!blindAttempted && hasAffordableSpell(g.battle.commandActor, 158):
				wantedSpell, wantCommand = 158, bcSpell
			case g.bossSurrenderStage == bossSurrenderBattle &&
				defenseCasts < 2 && hasAffordableSpell(g.battle.commandActor, 154):
				wantedSpell, wantCommand = 154, bcSpell
			case finalBoss && g.battle.monID == 0x7a && g.battle.commandActor == 0:
				// 怨靈巴拉摩斯封咒後由同伴輸出；主角保留 MP
				// 給後續殭屍／索瑪連戰，避免第一戰耗盡資源。
				wantCommand = bcWar
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
				if wantedSpell == 156 {
					sealAttempted = true
				}
				if wantedSpell == 151 {
					attackBuffAttempted = true
				}
				in.Confirm = true
			}
		case phItem:
			want := herbSlot(g.battle.commandActor)
			if want < 0 {
				t.Fatalf("正式戰鬥進入道具選單但 actor %d 沒有藥草", g.battle.commandActor)
			}
			if g.battle.itemCursor != want {
				in.DirEdge = 0
			} else {
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
		t.Fatalf("戰鬥 %d 次正式輸入後仍未結束：phase=%d actor=%d cursor=%d target=%d healTarget=%d pending=%+v herbs=%d/%d targets=%v hits=%+v mon=%d result=%d hero=%d enemyHP=%v msg=%q queue=%d resume=%d",
			maxInputs,
			g.battle.phase, g.battle.commandActor, g.battle.cursor, g.battle.targetCursor,
			healTarget, g.battle.pending, g.battle.usedHerbs, g.battle.heroHerbs,
			g.battle.aliveActorIndices(), g.battle.hits,
			g.battle.monID, g.battle.result, g.battle.heroHP, enemyHP,
			g.battle.msg, len(g.battle.messageQueue), g.battle.messageResume)
	}
	// 最後一個正式輸入可能同時把 active 清掉並寫入敗北 result；迴圈頂端
	// 的診斷看不到這個 frame。結束後仍須失敗即關閉，不能讓事件的敗／逃
	// transaction 只在下一個旗標斷言中被誤報成 consumer 缺失。
	if g.battle.result == 2 {
		comp := make([][4]int, len(g.battle.companions))
		for i, c := range g.battle.companions {
			comp[i] = [4]int{c.level, c.hp, c.maxHP, c.status}
		}
		t.Fatalf("正式戰鬥最後輸入造成全滅：mon=%d hero=%d/%d MP=%d conditions=%#x inventory=%v companions(lv,hp,max,status)=%v usedHerbs=%d",
			g.battle.monID, g.battle.heroHP, g.battle.heroMax, g.heroMP,
			g.heroConditions, g.inventory, comp, g.battle.usedHerbs)
	}
}

// traceSelectHealSpell 是 deterministic 玩家策略，不是原版 spell table：一般長連戰
// 選最低 MP，單場高傷 Boss 可選最高 Base；兩者都只能從 actor 已學且可負擔的正式
// 咒文中選擇。
func traceSelectHealSpell(records []int, mpBudget int, strongest bool) int {
	best, bestMP, bestBase := -1, int(^uint(0)>>1), -1
	for _, rec := range records {
		def, ok := spell.GetDef(rec)
		if !ok || def.Kind != spell.Heal || def.MP > mpBudget {
			continue
		}
		if strongest {
			if def.Base > bestBase || (def.Base == bestBase && def.MP < bestMP) {
				best, bestMP, bestBase = rec, def.MP, def.Base
			}
			continue
		}
		if def.MP < bestMP || (def.MP == bestMP && def.Base > bestBase) {
			best, bestMP, bestBase = rec, def.MP, def.Base
		}
	}
	return best
}

func TestTraceSelectHealSpellPolicy(t *testing.T) {
	records := spell.KnownAll(5, 40)
	cheap := traceSelectHealSpell(records, 200, false)
	strong := traceSelectHealSpell(records, 200, true)
	cheapDef, cheapOK := spell.GetDef(cheap)
	strongDef, strongOK := spell.GetDef(strong)
	if !cheapOK || !strongOK || cheapDef.Kind != spell.Heal || strongDef.Kind != spell.Heal {
		t.Fatalf("回復策略未選到合法咒文：cheap=%d strong=%d", cheap, strong)
	}
	if cheapDef.MP >= strongDef.MP || cheapDef.Base >= strongDef.Base {
		t.Fatalf("最低 MP／最高 Base 策略未分離：cheap=%+v strong=%+v", cheapDef, strongDef)
	}
	if got := traceSelectHealSpell(records, cheapDef.MP-1, false); got >= 0 {
		t.Fatalf("MP 預算不足時應失敗即關閉，got rec%d", got)
	}
}

// traceTownSectionTo 從目前 section 依真實 transition table 找路，再逐格走過每個 portal。
func traceTownSectionTo(t *testing.T, g *Game, wantCty, wantSec int, wantNPC ...int) {
	traceTownSectionToWithRepelPolicy(t, g, wantCty, wantSec, false, wantNPC...)
}

// traceTownSectionToPreservingRura 只供已證實的長 encounter section 鏈；效果將盡時
// 可正式施放特黑洛斯，但必須為同一施法者保留一發魯拉。
func traceTownSectionToPreservingRura(t *testing.T, g *Game, wantCty, wantSec int, wantNPC ...int) {
	traceTownSectionToWithRepelPolicy(t, g, wantCty, wantSec, true, wantNPC...)
}

func traceTownSectionToWithRepelPolicy(t *testing.T, g *Game, wantCty, wantSec int,
	preserveRura bool, wantNPC ...int) {
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
				if len(wantNPC) >= 3 && wantNPC[2] == -999 {
					// Internal callers use the sentinel to require the exact
					// transition destination, not merely a walkable point near
					// an NPC/treasure.
					reachable = sc != nil && cur.x == wantNPC[0] && cur.y == wantNPC[1]
				} else {
					reachable = sc != nil && (cur.x == wantNPC[0] && cur.y == wantNPC[1] ||
						len(tracePortalPath(sc, cur.x, cur.y, wantNPC[0], wantNPC[1], g.keyTier())) > 0)
				}
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
		traceWalkThroughPortalWithRepelPolicy(t, g, h.portalX, h.portalY,
			h.to.cty, h.to.sec, preserveRura, h.to.x, h.to.y)
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
	traceWalkThroughPortalWithRepelPolicy(t, g, x, y, wantCty, wantSec, false, wantPos...)
}

func traceWalkThroughPortalPreservingRura(t *testing.T, g *Game, x, y, wantCty, wantSec int,
	wantPos ...int) {
	traceWalkThroughPortalWithRepelPolicy(t, g, x, y, wantCty, wantSec, true, wantPos...)
}

func traceWalkThroughPortalWithRepelPolicy(t *testing.T, g *Game, x, y, wantCty, wantSec int,
	preserveRura bool, wantPos ...int) {
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
		if g.battle.active {
			// CTY07/08 是有遭遇表的塔區；踏格時可能先進入正式戰鬥，
			// 路由驗證必須用同一條 production battle input 完成，
			// 不能把戰鬥留在背景而誤報成 portal 不可達。隨機高危遭遇
			// 先依正式逃跑指令離開，避免把路由驗證變成資源／RNG 賭局；
			// 事件型 boss 仍由各自的正式劇情段落處理。
			shouldFlee := true
			if preserveRura && traceCanSafelyFinishEncounter(g) {
				shouldFlee = false
			}
			traceResolveBattle(t, g, shouldFlee, preserveRura)
			continue
		}
		if traceCureHeroPoison(t, g) {
			continue
		}
		if g.encounterEnabled() && g.repel <= 8 && g.countItem(itemuse.ItemHolyWater) > 1 {
			traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
			continue
		}
		if preserveRura && g.encounterEnabled() && g.repel <= 8 {
			caster := g.fieldSpellCaster(fieldToheros)
			reserveRuraMP := 0
			if caster == 0 {
				reserveRuraMP = spell.MPCost(fieldRura)
			}
			if caster >= 0 && g.fieldCasterMP(caster) >= spell.MPCost(fieldToheros)+reserveRuraMP {
				traceUseFieldSpell(t, g, fieldToheros)
				continue
			}
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
		if g.px != x || g.py != y {
			if _, _, _, _, ok := g.cur.tileTransition(g.px, g.py); ok {
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
					traceProtectBaramosHazard(t, g, nx, ny)
					if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
						t.Fatalf("離開目前回程 portal dir%d：%v", dir, err)
					}
					movedOff = g.px == nx && g.py == ny
					if movedOff {
						break
					}
				}
				if !movedOff {
					t.Fatalf("目前位置 (%d,%d) 是 portal 但找不到可走開的鄰格", g.px, g.py)
				}
				continue
			}
		}
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
		// CTY66/65 的原始 ah&0x0c 傷害地板分成多個不連續區段；精訊版
		// 多拉瑪那只在第一次踏入當前連續區段時消費 0x2000，因此離開
		// 一段後若下一格又是新區段，玩家必須再次透過正式命令窗施法。
		// 這裡只在玩家正式走向下一格前補上該輸入，不修改 HP 或 hazard 狀態。
		traceProtectBaramosHazard(t, g, g.px+dx, g.py+dy)
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("走向 portal dir%d: %v", path[0], err)
		}
	}
	t.Fatalf("走到 portal(%d,%d)後仍未抵達 CTY%d sec%d；目前 CTY%d sec%d @(%d,%d)，title=%v dlg=%v dlgPos=%d/%d cmd=%v panel=%d gateStage=%d defeatStage=%d heroHP=%d cond=%#x cd=%d path=%v",
		x, y, wantCty, wantSec, g.curCty, sceneSection(g.cur), g.px, g.py,
		g.showTitle, g.dlg.open, g.dlg.pos, len(g.dlg.buf), g.cmd.open, g.panel,
		g.sequenceGateStage, g.defeatDialogueStage, g.heroHP, g.heroConditions, g.cd,
		tracePortalPath(g.cur, g.px, g.py, x, y, g.keyTier()))
}

func traceWalkTo(t *testing.T, g *Game, x, y int) {
	t.Helper()
	// 踏入目標格的同一個正式移動 frame 仍可能開啟 encounter modal；座標
	// 已到不表示玩家已重新取得地圖輸入。把 battle 納入終止條件，避免
	// 呼叫端只等 cd 時留下 modal 而無限等待。
	for i := 0; i < 3000 && (g.px != x || g.py != y || g.battle.active); i++ {
		traceWalkOne(t, g, x, y)
	}
	if g.px != x || g.py != y || g.battle.active {
		t.Fatalf("無法穩定走到 (%d,%d)，停在 (%d,%d)，battle=%v dlg=%v gateStage=%d cmd=%v path=%v",
			x, y, g.px, g.py, g.battle.active, g.dlg.open, g.sequenceGateStage, g.cmd.open,
			tracePath(g.cur, g.px, g.py, x, y))
	}
}

func traceWalkToNoPortal(t *testing.T, g *Game, x, y int) {
	t.Helper()
	for i := 0; i < 3000 && (g.px != x || g.py != y || g.battle.active); i++ {
		if g.battle.active {
			// CTY12/13 金字塔 section 仍保留原始遭遇 gate；正式走向
			// 開關時若先觸發遭遇，必須以 production battle input 收尾，
			// 不能讓 modal battle 吞掉後續方向鍵而誤報地圖不連通。
			traceResolveBattle(t, g, true)
			continue
		}
		if traceCureHeroPoison(t, g) {
			continue
		}
		if g.encounterEnabled() && g.repel <= 8 && g.countItem(itemuse.ItemHolyWater) > 1 {
			traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
			continue
		}
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
		traceProtectBaramosHazard(t, g, g.px+dx, g.py+dy)
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("無轉場路徑移動 dir%d: %v", path[0], err)
		}
	}
	if g.px != x || g.py != y || g.battle.active {
		compState := make([][3]int, len(g.companions))
		for i, m := range g.companions {
			compState[i] = [3]int{m.CurHP, int(m.Conditions), m.MaxHP()}
		}
		_, heroMaxHP, _, _, _ := g.heroStats()
		t.Fatalf("無轉場路徑無法穩定走到 (%d,%d)，CTY%d sec%d 停在 (%d,%d)，battle=%v dlg=%v defeat=%d hero=(%d,%#x/%d) companions=%v",
			x, y, g.curCty, sceneSection(g.cur), g.px, g.py, g.battle.active, g.dlg.open,
			g.defeatDialogueStage, g.heroHP, g.heroConditions, heroMaxHP, compState)
	}
}

func traceWalkOne(t *testing.T, g *Game, tx, ty int) {
	t.Helper()
	if g.battle.active {
		traceResolveBattle(t, g, true)
		return
	}
	if traceCureHeroPoison(t, g) {
		return
	}
	if g.encounterEnabled() && g.repel <= 8 && g.countItem(itemuse.ItemHolyWater) > 1 {
		traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
		return
	}
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
	dx, dy := dirDelta(path[0])
	traceProtectBaramosHazard(t, g, g.px+dx, g.py+dy)
	if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
		t.Fatalf("移動 dir%d: %v", path[0], err)
	}
}

func traceTalkNPC(t *testing.T, g *Game, nx, ny int, avoidPortals ...bool) {
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
		path := tracePath(g.cur, g.px, g.py, p.x, p.y)
		if len(avoidPortals) > 0 && avoidPortals[0] {
			path = tracePortalPath(g.cur, g.px, g.py, p.x, p.y, g.keyTier())
		}
		if len(path) > 0 ||
			g.px == p.x && g.py == p.y {
			if len(avoidPortals) > 0 && avoidPortals[0] {
				traceWalkToNoPortal(t, g, p.x, p.y)
			} else {
				traceWalkTo(t, g, p.x, p.y)
			}
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
