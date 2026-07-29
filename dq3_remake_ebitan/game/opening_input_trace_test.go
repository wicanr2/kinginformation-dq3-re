package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
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
	for i := 0; i < 3; i++ {
		traceTalkNPC(t, g, 2, 3)
		if !g.tavern.active {
			t.Fatalf("第%d次對冒險者登錄所 NPC 應開 tavern modal", i+1)
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

	// 正常前段迷宮會連續遇敵；用國王給的 50G 到阿里阿罕道具店買 5 株藥草。
	traceTalkNPC(t, g, 15, 24) // CTY00 sec0 facility k2，道具店第一項是 0x41
	if !g.shop.active {
		t.Fatal("對阿里阿罕道具店 NPC 應開商店")
	}
	for i := 0; i < 5; i++ {
		press(InputState{Confirm: true})
	}
	press(InputState{Cancel: true})
	if g.countItem(herbCode) != 5 {
		t.Fatalf("正式購買後應有 5 株藥草，got %d (gold=%d)", g.countItem(herbCode), g.heroGold)
	}

	// 國王給的裝備不會自動穿上；依正式命令窗選「裝備」，把背包第4格銅劍裝給隊長。
	press(InputState{Confirm: true})          // 開命令
	send(InputState{DirHeld: -1, DirEdge: 0}) // 對話→狀況
	send(InputState{DirHeld: -1, DirEdge: 0}) // 狀況→裝備
	press(InputState{Confirm: true})          // 開裝備面板
	for i := 0; i < 3; i++ {
		send(InputState{DirHeld: -1, DirEdge: 0})
	}
	press(InputState{Confirm: true}) // 裝備 inventory[3] = 銅劍 0x03
	press(InputState{Cancel: true})
	if g.equip[0] != 0x03 {
		t.Fatalf("正式裝備輸入未穿上國王給的銅劍：equip=%v", g.equip)
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
	for i := 0; i < 5 && g.heroGold >= g.shop.items.Price(herbCode); i++ {
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
	if err := g.Save(); err != nil {
		t.Fatalf("保存羅馬利亞 checkpoint: %v", err)
	}
	if err := restored.Load(); err != nil {
		t.Fatalf("讀取羅馬利亞 checkpoint: %v", err)
	}
	if restored.hasItem(itemuse.ItemMagicBall) || restored.storyFlag(magicBallIntactFlag) ||
		!restored.inTown || restored.curCty != 2 || sceneSection(restored.cur) != 0 {
		t.Fatalf("羅馬利亞 checkpoint round-trip 錯：ball=%v flag51=%v town=%v cty=%d sec=%d",
			restored.hasItem(itemuse.ItemMagicBall), restored.storyFlag(magicBallIntactFlag),
			restored.inTown, restored.curCty, sceneSection(restored.cur))
	}
}

func traceExitTownBoundary(t *testing.T, g *Game) {
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
	traceWalkTo(t, g, best.x, best.y)
	for g.cd > 0 {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatalf("等待出城 cooldown: %v", err)
		}
	}
	if err := g.step(InputState{DirHeld: best.dir, DirEdge: -1}); err != nil {
		t.Fatalf("走出城鎮邊界: %v", err)
	}
	if g.inTown {
		t.Fatalf("從 CTY%d 邊界 (%d,%d) dir%d 未正常出城", g.curCty, best.x, best.y, best.dir)
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
func traceAdventureWalkToCty(t *testing.T, g *Game, wantCty int) {
	t.Helper()
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
			traceResolveBattle(t, g)
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
		t.Fatalf("無法由地表抵達 CTY%d；目前 town=%v cty=%d @(%d,%d)",
			wantCty, g.inTown, g.curCty, g.px, g.py)
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

func traceResolveBattle(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 512 && g.battle.active; i++ {
		if g.battle.result == 2 {
			t.Fatalf("正常前段路線發生全滅：mon=%d heroHP=%d", g.battle.monID, g.battle.heroHP)
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		switch g.battle.phase {
		case phCommand:
			menu := g.battle.commandMenu(g.battle.commandActor)
			needHeal := g.battle.commandActor == 0 &&
				g.battle.heroHP*2 < g.battle.heroMax &&
				g.battle.usedHerbs < g.battle.heroHerbs
			needFlee := g.battle.commandActor == g.battle.firstCommandActor() &&
				g.battle.monID >= 10
			if needHeal && menu[g.battle.cursor] != bcItem {
				in.DirEdge = 0
			} else if !needHeal && needFlee && menu[g.battle.cursor] != bcFlee {
				in.DirEdge = 0
			} else {
				in.Confirm = true
			}
		case phTargetEnemy, phTargetAlly, phMessage, phEnd:
			in.Confirm = true
		default:
			t.Fatalf("自動戰鬥遇到未處理 phase=%d", g.battle.phase)
		}
		if err := g.step(in); err != nil {
			t.Fatalf("推進正式戰鬥輸入: %v", err)
		}
	}
	if g.battle.active {
		t.Fatal("戰鬥 512 次正式輸入後仍未結束")
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
			if len(wantNPC) < 2 || traceNPCReachable(g, cur.cty, cur.sec, cur.x, cur.y, wantNPC[0], wantNPC[1]) {
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
				if dsec >= 0xfe && wantCty >= 0 {
					continue
				}
				if (x != cur.x || y != cur.y) && len(tracePortalPath(sc, cur.x, cur.y, x, y)) == 0 {
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
		t.Fatalf("transition graph 無法由 CTY%d sec%d 到 CTY%d sec%d",
			start.cty, start.sec, wantCty, wantSec)
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
		path := tracePortalPath(g.cur, g.px, g.py, x, y)
		if len(path) == 0 {
			t.Fatalf("無法在不踩其他 transition 下抵達 portal(%d,%d)，目前 @(%d,%d)",
				x, y, g.px, g.py)
		}
		if err := g.step(InputState{DirHeld: path[0], DirEdge: -1}); err != nil {
			t.Fatalf("走向 portal dir%d: %v", path[0], err)
		}
	}
	t.Fatalf("走到 portal(%d,%d)後仍未抵達 CTY%d sec%d；目前 CTY%d sec%d @(%d,%d)",
		x, y, wantCty, wantSec, g.curCty, sceneSection(g.cur), g.px, g.py)
}

func traceWalkTo(t *testing.T, g *Game, x, y int) {
	t.Helper()
	for i := 0; i < 3000 && (g.px != x || g.py != y); i++ {
		traceWalkOne(t, g, x, y)
	}
	if g.px != x || g.py != y {
		t.Fatalf("無法走到 (%d,%d)，停在 (%d,%d)", x, y, g.px, g.py)
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
	t.Fatalf("無法抵達 NPC(%d,%d) 的相鄰格", nx, ny)
}

// tracePortalPath 把目標以外的 transition tile 視為阻擋；否則最短路徑可能先踩到
// 另一座樓梯，執行時場景已切換，靜態 graph 卻仍誤判可達。
func tracePortalPath(sc *Scene, sx, sy, tx, ty int) []int {
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
				sc.Blocked(n.x, n.y) {
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
