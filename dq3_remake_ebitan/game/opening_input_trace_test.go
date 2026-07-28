package game

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOpeningProductionInputTrace 從標題開始，只送入 production InputState：
// 標題→開始→英數命名→性別→四段開場→走出城鎮→進王城→上樓→王座。
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

func traceWalkThroughPortal(t *testing.T, g *Game, x, y, wantCty, wantSec int) {
	t.Helper()
	for i := 0; i < 3000; i++ {
		if g.curCty == wantCty && g.cur != nil && g.cur.sec == wantSec {
			return
		}
		traceWalkOne(t, g, x, y)
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
