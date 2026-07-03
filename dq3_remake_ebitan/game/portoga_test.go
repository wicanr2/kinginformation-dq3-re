package game

import "testing"

// TestScriptedTalkPepperGate:古布達(CTY15 byte4=25)黑胡椒 gate(main.c:1736)——
// 未救出達妮亞(無 flag 0x211)不給胡椒;救人後(flag 0x211=true)才給黑胡椒 0x5c。
func TestScriptedTalkPepperGate(t *testing.T) {
	g := &Game{flags: map[int]bool{}, curCty: 15}

	g.scriptedTalk(25)
	if g.hasItem(itemPepper) {
		t.Error("未救出達妮亞(無 flag 0x211)不應給黑胡椒")
	}

	g.flags[0x211] = true
	g.scriptedTalk(25)
	if !g.hasItem(itemPepper) {
		t.Error("救出達妮亞後(flag 0x211)應給黑胡椒 0x5c")
	}
}

// TestTalkPortogaKing:波魯多加國王(CTY37 (9,6))授船特例(main.c:1607-1631)——
// 無胡椒不授船;持黑胡椒獻上 → 消耗胡椒 + shipOwned + 船停泊港口座標 + msShip 里程碑;
// 已取船後再對話仍維持授船狀態(不重複扣胡椒/不出錯)。
func TestTalkPortogaKing(t *testing.T) {
	g := &Game{flags: map[int]bool{}, curCty: ctyPortoga}

	g.talkPortogaKing()
	if g.shipOwned {
		t.Error("無黑胡椒不應授船")
	}

	g.inventory = append(g.inventory, itemPepper)
	g.talkPortogaKing()
	if !g.shipOwned {
		t.Fatal("持黑胡椒獻上應授船")
	}
	if g.shipAboard {
		t.Error("授船當下應停在港口陸地(未登船)")
	}
	if g.shipX != portogaHarborX || g.shipY != portogaHarborY {
		t.Errorf("船應停泊港口 (%d,%d),得 (%d,%d)", portogaHarborX, portogaHarborY, g.shipX, g.shipY)
	}
	if !g.progressDone(msShip) {
		t.Error("應設 msShip(取船)里程碑")
	}
	if g.hasItem(itemPepper) {
		t.Error("獻上的黑胡椒應被消耗(移除背包)")
	}

	// 再度對話(已取船):不應 panic、shipOwned 維持 true。
	g.talkPortogaKing()
	if !g.shipOwned {
		t.Error("已取船後再對話應仍保持授船狀態")
	}
}

// TestSelectCommandTalkPortogaKing:完整 NPC 對話派發路徑(main.c 位置特例優先於
// scripted 表判定)—— 玩家站國王 sub2 NPC 前 cmdTalk,經 selectCommand → sub==2
// 分支的位置判定命中波魯多加國王,而非落回泛用 scriptedTalk(byte4 查表)。
func TestSelectCommandTalkPortogaKing(t *testing.T) {
	sc := &Scene{
		w: 20, h: 20,
		npcs: []npcInst{{x: portogaKingX, y: portogaKingY, ctrl: 2 << 3, b4: 26}},
	}
	g := &Game{
		cur: sc, curCty: ctyPortoga, inTown: true, flags: map[int]bool{},
		px: portogaKingX, py: portogaKingY + 1, facing: 1, // 面向上 → frontTile=(kingX,kingY)
	}

	g.selectCommand(cmdTalk)
	if g.shipOwned {
		t.Error("無黑胡椒時對話不應授船")
	}

	g.inventory = append(g.inventory, itemPepper)
	g.selectCommand(cmdTalk)
	if !g.shipOwned {
		t.Fatal("經 selectCommand 派發,持黑胡椒對國王應授船")
	}
	if g.shipX != portogaHarborX || g.shipY != portogaHarborY {
		t.Errorf("船應停泊港口 (%d,%d),得 (%d,%d)", portogaHarborX, portogaHarborY, g.shipX, g.shipY)
	}
}
