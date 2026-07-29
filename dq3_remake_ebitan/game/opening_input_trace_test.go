package game

import (
	"os"
	"path/filepath"
	"testing"
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
	for i := 0; i < 12000 && (!g.inTown || g.curCty != wantCty); i++ {
		if g.battle.active {
			traceResolveBattle(t, g)
			continue
		}
		traceWalkOne(t, g, tx, ty)
	}
	if !g.inTown || g.curCty != wantCty {
		t.Fatalf("無法由地表抵達 CTY%d；目前 town=%v cty=%d @(%d,%d)",
			wantCty, g.inTown, g.curCty, g.px, g.py)
	}
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
			if needHeal && menu[g.battle.cursor] != bcItem {
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
