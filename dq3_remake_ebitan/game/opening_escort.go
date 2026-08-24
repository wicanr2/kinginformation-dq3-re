package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// openingEscortAnimating reports the finite post-creation mother escort.
func (g *Game) openingEscortAnimating() bool {
	return g.openingEscort != nil && g.openingEscortIndex >= 0 &&
		g.openingEscortIndex < len(g.openingEscort.Frames) && g.openingEscortNPC >= 0
}

func (g *Game) startMotherEscort() bool {
	e := g.openingEscort
	if e == nil || g.cur == nil || g.curCty != e.CTY || g.cur.sec != e.Section || len(e.Frames) < 2 {
		return false
	}
	first := e.Frames[0]
	idx := g.cur.npcAt(first.Leader.X, first.Leader.Y)
	if idx < 0 {
		return false
	}
	g.openingEscortNPC = idx
	g.openingEscortIndex = 0
	g.openingEscortTick = 0
	g.applyOpeningEscortFrame(first)
	return true
}

func (g *Game) applyOpeningEscortFrame(frame gamepack.OpeningEscortFrame) {
	if g.cur == nil || g.openingEscortNPC < 0 || g.openingEscortNPC >= len(g.cur.npcs) {
		return
	}
	n := &g.cur.npcs[g.openingEscortNPC]
	n.facing = facingBetween(n.x, n.y, frame.Leader.X, frame.Leader.Y, n.facing)
	n.x, n.y = frame.Leader.X, frame.Leader.Y
	n.walk ^= 1
	g.facing = facingBetween(g.px, g.py, frame.Player.X, frame.Player.Y, g.facing)
	g.px, g.py = frame.Player.X, frame.Player.Y
	g.walk ^= 1
}

func facingBetween(x0, y0, x1, y1, fallback int) int {
	switch {
	case y1 > y0:
		return 0
	case y1 < y0:
		return 1
	case x1 < x0:
		return 2
	case x1 > x0:
		return 3
	default:
		return fallback
	}
}

func (g *Game) advanceOpeningEscort() {
	frame := g.openingEscort.Frames[g.openingEscortIndex]
	g.openingEscortTick++
	if g.openingEscortTick < frame.HoldFrames {
		return
	}
	g.openingEscortTick = 0
	g.openingEscortIndex++
	if g.openingEscortIndex < len(g.openingEscort.Frames) {
		g.applyOpeningEscortFrame(g.openingEscort.Frames[g.openingEscortIndex])
		return
	}
	g.openingEscortIndex = -1
	g.openingEscortNPC = -1
	g.finishMotherEscort()
	g.dlg.Open(openingMotherDirectionsRec)
}

// motherEscort remains the direct state helper used by focused state tests.
// Production input uses startMotherEscort/advanceOpeningEscort instead.
func (g *Game) motherEscort() { g.finishMotherEscort() }
