package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

func (g *Game) npcMotion() *gamepack.NPCMotionDefinition {
	if g.pack == nil {
		return nil
	}
	return g.pack.Characters.NPCMotion
}

// Automatic movement visits live cells in row-major order. An actor moving
// right or down can therefore be visited again later in the same scan.
func (g *Game) npcTick() {
	sc, rule := g.cur, g.npcMotion()
	if sc == nil || !g.inTown || rule == nil {
		return
	}
	// Keep the existing animation approximation separate from movement order.
	for i := range sc.npcs {
		n := &sc.npcs[i]
		if n.ctrl&rule.FrozenMask == 0 {
			n.anim++
			if n.anim%18 == 0 {
				n.walk ^= 1
			}
		}
	}
	if len(sc.hiMap) != sc.w*sc.h || g.px < 0 || g.py < 0 || g.px >= sc.w || g.py >= sc.h {
		return
	}
	layer := sc.tileLayer(g.px, g.py)
	left, top := g.px-rule.ViewportAnchor.X, g.py-rule.ViewportAnchor.Y
	for y := top; y < top+rule.ViewportRows; y++ {
		for x := left; x < left+rule.ViewportColumns; x++ {
			if x < 0 || y < 0 || x >= sc.w || y >= sc.h || sc.tileLayer(x, y) != layer {
				continue
			}
			if idx := sc.npcAt(x, y); idx >= 0 {
				g.npcStep(idx)
			}
		}
	}
}

// Raw town direction order and renderer direction order are distinct formats.
func npcCtrlFacing(dir int) int { return [...]int{0, 2, 1, 3}[dir&3] }

func (g *Game) npcStep(idx int) {
	sc, rule := g.cur, g.npcMotion()
	if sc == nil || rule == nil || idx < 0 || idx >= len(sc.npcs) || len(sc.hiMap) != sc.w*sc.h {
		return
	}
	n := &sc.npcs[idx]
	if n.x < 0 || n.y < 0 || n.x >= sc.w || n.y >= sc.h {
		return
	}
	if sc.npcRng.Next(rule.Evaluation.Bound) != *rule.Evaluation.Accepted || n.ctrl&rule.FrozenMask != 0 || n.ctrl&rule.MoveMask == 0 {
		return
	}
	dir := n.ctrl & rule.DirectionMask
	if sc.npcRng.Next(rule.DirectionBound) == dir {
		if sc.npcRng.Next(rule.Step.Bound) == *rule.Step.Accepted {
			g.npcTryStep(idx)
		}
		return
	}
	if sc.npcRng.Next(rule.Turn.Bound) != *rule.Turn.Accepted {
		return
	}
	layer := sc.tileLayer(n.x, n.y)
	if layer < 0 || layer >= len(rule.TurnDeltaByLayer) {
		return
	}
	quotient := (int(sc.npcRng.State()) / rule.Turn.Bound) & rule.QuotientMask
	dir = (quotient + rule.TurnDeltaByLayer[layer]) & rule.DirectionMask
	n.ctrl = (n.ctrl &^ rule.DirectionMask) | dir
	n.facing = npcCtrlFacing(dir)
}

// Random qualification belongs to npcStep. This transaction checks bounds,
// proximity, occupancy and the complete pack-owned terrain mask.
func (g *Game) npcTryStep(idx int) bool {
	sc, rule := g.cur, g.npcMotion()
	if sc == nil || rule == nil || idx < 0 || idx >= len(sc.npcs) || sc.attr == nil {
		return false
	}
	n := &sc.npcs[idx]
	dir := n.ctrl & rule.DirectionMask
	if dir >= len(rule.Directions) {
		return false
	}
	d := rule.Directions[dir]
	tx, ty := n.x+d.X, n.y+d.Y
	if tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return false
	}
	distance := g.py - ty
	if d.X != 0 {
		distance = g.px - tx
	}
	if distance < rule.MinimumAxisDistance && distance > -rule.MinimumAxisDistance {
		return false
	}
	if sc.npcAt(tx, ty) >= 0 {
		return false
	}
	tile := sc.tileIdx(tx, ty)
	if tile < 0 || tile >= len(sc.attr.A) || int(sc.attr.A[tile])&rule.BlockedAttributeMask != 0 {
		return false
	}
	n.x, n.y, n.facing = tx, ty, npcCtrlFacing(dir)
	return true
}
