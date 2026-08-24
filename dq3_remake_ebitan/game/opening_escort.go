package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// openingEscortAnimating reports the finite post-creation mother escort.
func (g *Game) openingEscortAnimating() bool {
	if g.openingEscort == nil || g.openingEscortIndex < 0 {
		return false
	}
	switch g.openingEscortPhase {
	case 0:
		return g.openingEscortIndex < len(g.openingEscort.Frames) && g.openingEscortNPC >= 0
	case 1:
		return g.openingEscortIndex < len(g.openingEscort.ArrivalFrames)
	default:
		return false
	}
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
	g.openingEscortPhase = 0
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
	holdFrames := 0
	if g.openingEscortPhase == 0 {
		holdFrames = g.openingEscort.Frames[g.openingEscortIndex].HoldFrames
	} else {
		holdFrames = g.openingEscort.ArrivalFrames[g.openingEscortIndex].HoldFrames
	}
	g.openingEscortTick++
	if g.openingEscortTick < holdFrames {
		return
	}
	g.openingEscortTick = 0
	g.openingEscortIndex++
	if g.openingEscortPhase == 0 {
		if g.openingEscortIndex < len(g.openingEscort.Frames) {
			g.applyOpeningEscortFrame(g.openingEscort.Frames[g.openingEscortIndex])
			return
		}
		g.openingEscortNPC = -1
		if !g.finishMotherEscort() {
			g.openingEscortPhase = -1
			g.openingEscortIndex = -1
			return
		}
		g.openingEscortPhase = 1
		g.openingEscortIndex = 0
		g.applyOpeningArrivalFrame(g.openingEscort.ArrivalFrames[0])
		return
	}
	if g.openingEscortIndex < len(g.openingEscort.ArrivalFrames) {
		g.applyOpeningArrivalFrame(g.openingEscort.ArrivalFrames[g.openingEscortIndex])
		return
	}
	g.openingEscortPhase = -1
	g.openingEscortIndex = -1
	g.dlg.Open(g.openingEscort.CompletionDialogueRecord)
}

func (g *Game) applyOpeningArrivalFrame(frame gamepack.OpeningArrivalFrame) {
	g.facing = facingBetween(g.px, g.py, frame.Player.X, frame.Player.Y, g.facing)
	g.px, g.py = frame.Player.X, frame.Player.Y
	g.walk ^= 1
}

func (g *Game) completeOpeningEscort() {
	if g.openingEscort == nil {
		return
	}
	for _, flag := range g.openingEscort.SetStoryFlags {
		g.setStoryFlag(flag, true)
	}
	for _, flag := range g.openingEscort.ClearStoryFlags {
		g.setStoryFlag(flag, false)
	}
}

// motherEscort remains a direct full-transaction helper for focused state and
// renderer fixtures. Production input uses the timed two-phase sequence.
func (g *Game) motherEscort() {
	if g.openingEscort == nil || len(g.openingEscort.ArrivalFrames) == 0 {
		return
	}
	if !g.finishMotherEscort() {
		return
	}
	g.applyOpeningArrivalFrame(g.openingEscort.ArrivalFrames[len(g.openingEscort.ArrivalFrames)-1])
	g.completeOpeningEscort()
}
