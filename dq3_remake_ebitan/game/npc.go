package game

// NPC 隨機走動步進器(移植 dq3_npc.c / dq3_scene_npc_tick,鏡射 EXE mover L02025)。
// Go 版用 npcInst 位置 + Scene.npcAt/Blocked 取代 C 的 hi_map OCCUPIED stamp(語意等價)。

var npcDDX = [4]int{0, -1, 0, 1} // 0下 1左 2上 3右(EXE [0xb35] 表)
var npcDDY = [4]int{1, 0, -1, 0}

const (
	npcMoveBit   = 0x04 // ctrl bit2:可移動(清 = 靜止 NPC)
	npcFrozenBit = 0x80 // ctrl bit7:凍結
)

// npcTick:每幀步進當前城鎮所有 NPC(移植 dq3_scene_npc_tick)。地表/戰鬥/對話中不呼叫。
func (g *Game) npcTick() {
	sc := g.cur
	if sc == nil || !g.inTown {
		return
	}
	half := len(sc.npcs) / 2
	for i := range sc.npcs {
		g.npcStep(i, half)
	}
}

// npcStep:單一 NPC 步進判定(移植 dq3_npc_step)。
func (g *Game) npcStep(idx, half int) {
	sc := g.cur
	n := &sc.npcs[idx]
	if sc.npcRng.Next(4) != 1 { // 節流:每幀 1/4 才評估(L02040)
		return
	}
	if n.ctrl&npcFrozenBit != 0 { // 凍結(L02048)
		return
	}
	if n.ctrl&npcMoveBit == 0 { // bit2 清 = 靜止 NPC(L0204e)
		return
	}
	cur := n.ctrl & 3
	rdir := sc.npcRng.Next(4) // 隨機方向(L02062)
	if rdir == cur {          // 同朝向 → 直接走(L02057)
		g.npcTryStep(idx)
		return
	}
	if sc.npcRng.Next(20) == 1 { // 1/20 才轉向(L02071)
		nd := (cur + 3) & 3 // ±1 by index(L0207f):前半 +1、後半 -1
		if idx < half {
			nd = (cur + 1) & 3
		}
		n.ctrl = (n.ctrl &^ 3) | nd
		g.npcTryStep(idx)
	}
}

// npcTryStep:沿當前朝向試走一步(移植 try_step)。成功回 true。
func (g *Game) npcTryStep(idx int) bool {
	sc := g.cur
	n := &sc.npcs[idx]
	dir := n.ctrl & 3
	tx, ty := n.x+npcDDX[dir], n.y+npcDDY[dir]
	if tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h { // 界外
		return false
	}
	// 近玩家閘(L020e1/L02111,分軸):左右看 X、上下看 Y;距離 <3 不走(留空間讓玩家靠近對話)。
	if dir&1 != 0 {
		if d := g.px - tx; d < 3 && d > -3 {
			return false
		}
	} else {
		if d := g.py - ty; d < 3 && d > -3 {
			return false
		}
	}
	if tx == g.px && ty == g.py { // 踏到玩家格 → 反向、不走(L02146)
		n.ctrl = (n.ctrl &^ 3) | ((dir + 2) & 3)
		return false
	}
	if sc.npcAt(tx, ty) >= 0 { // 已有 NPC(L0217e)
		return false
	}
	if sc.attr.Blocked(sc.tileIdx(tx, ty)) { // 牆(attr bit0,L0218c)
		return false
	}
	n.x, n.y = tx, ty
	return true
}
