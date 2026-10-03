package game

import (
	"errors"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// openingEscortAnimating reports the finite post-creation mother escort.
func (g *Game) openingEscortAnimating() bool {
	if g.openingEscort == nil || g.openingEscortIndex < 0 {
		return false
	}
	switch g.openingEscortPhase {
	case 0:
		return g.openingEscortIndex < len(g.openingEscort.Frames) && g.openingEscortNPC >= 0
	case 1, 3:
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
	if g.px != first.Player.X || g.py != first.Player.Y {
		return false
	}
	for _, f := range e.Frames {
		if f.Leader.X < 0 || f.Leader.Y < 0 || f.Player.X < 0 || f.Player.Y < 0 || f.Leader.X >= g.cur.w || f.Leader.Y >= g.cur.h || f.Player.X >= g.cur.w || f.Player.Y >= g.cur.h {
			return false
		}
	}
	idx := -1
	for i, n := range g.cur.npcs {
		if n.recordIndex == *e.Home.LeaderRecord && n.x == first.Leader.X && n.y == first.Leader.Y {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	g.openingEscortNPC = idx
	g.openingEscortPhase = 0
	g.openingEscortIndex = 0
	g.openingEscortTick = 0
	g.openingEscortDialogue = 0
	g.applyOpeningEscortFrame(first)
	return true
}

func (g *Game) applyOpeningEscortFrame(frame gamepack.OpeningArrivalFrame) {
	if g.cur == nil || g.openingEscortNPC < 0 || g.openingEscortNPC >= len(g.cur.npcs) {
		return
	}
	n := &g.cur.npcs[g.openingEscortNPC]
	n.facing = frame.LeaderFacing
	n.x, n.y = frame.Leader.X, frame.Leader.Y
	n.walk ^= 1
	g.facing = facingBetween(g.px, g.py, frame.Player.X, frame.Player.Y, g.facing)
	if g.px != frame.Player.X || g.py != frame.Player.Y {
		g.walk ^= 1
	}
	g.px, g.py = frame.Player.X, frame.Player.Y
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

func (g *Game) advanceOpeningEscort() error {
	holdFrames := 0
	if g.openingEscortPhase == 0 {
		holdFrames = g.openingEscort.Frames[g.openingEscortIndex].HoldFrames
	} else {
		holdFrames = g.openingEscort.ArrivalFrames[g.openingEscortIndex].HoldFrames
	}
	g.openingEscortTick++
	if g.openingEscortTick < holdFrames {
		return nil
	}
	g.openingEscortTick = 0
	g.openingEscortIndex++
	if g.openingEscortPhase == 0 {
		if g.openingEscortIndex < len(g.openingEscort.Frames) {
			g.applyOpeningEscortFrame(g.openingEscort.Frames[g.openingEscortIndex])
			return nil
		}
		g.openingEscortIndex, g.openingEscortPhase = -1, 4
		return nil
	}
	if g.openingEscortIndex < len(g.openingEscort.ArrivalFrames) {
		g.applyOpeningArrivalFrame(g.openingEscort.ArrivalFrames[g.openingEscortIndex])
		if g.openingEscortPhase == 1 && g.openingEscortIndex == g.openingEscort.DialogueFrameIndex {
			g.openingEscortPhase = 2
			g.openingEscortDialogue = 0
			if !g.openOpeningEscortText(g.openingEscort.DialogueTextIDs[0]) {
				return errors.New("opening escort text unavailable")
			}
		}
		if g.openingEscortPhase == 3 && g.openingEscortIndex == len(g.openingEscort.ArrivalFrames)-1 {
			g.completeOpeningEscort()
			g.openingEscortPhase, g.openingEscortIndex, g.openingEscortNPC = -1, -1, -1
		}
		return nil
	}
	g.openingEscortPhase = -1
	g.openingEscortIndex = -1
	g.openingEscortNPC = -1
	g.completeOpeningEscort()
	return nil
}

// resumeOpeningEscortAfterDialogue consumes the pack-owned dialogue sequence
// at one route frame. The flag transaction waits until the final movement ends.
func (g *Game) resumeOpeningEscortAfterDialogue() bool {
	if g.openingEscort == nil || g.openingEscortPhase != 2 || g.dlg.open {
		return false
	}
	g.openingEscortDialogue++
	if g.openingEscortDialogue < len(g.openingEscort.DialogueTextIDs) {
		return g.openOpeningEscortText(g.openingEscort.DialogueTextIDs[g.openingEscortDialogue])
	}
	g.clearOpeningPresentation()
	g.openingEscortPhase = 3
	g.openingEscortTick = 0
	return true
}

func (g *Game) applyOpeningArrivalFrame(frame gamepack.OpeningArrivalFrame) {
	if g.cur == nil || g.openingEscortNPC < 0 || g.openingEscortNPC >= len(g.cur.npcs) {
		return
	}
	n := &g.cur.npcs[g.openingEscortNPC]
	n.x, n.y, n.facing = frame.Leader.X, frame.Leader.Y, frame.LeaderFacing
	moved := g.px != frame.Player.X || g.py != frame.Player.Y
	g.facing = facingBetween(g.px, g.py, frame.Player.X, frame.Player.Y, g.facing)
	g.px, g.py = frame.Player.X, frame.Player.Y
	if moved {
		g.walk ^= 1
		g.deferRegionDialogueReturnAtPlayer()
	}
}

// The shared finite text renderer consumes a stable ID and a reviewed presentation.
// It does not inherit the home scene's camera when the dialogue is in another scene.
func (g *Game) openOpeningEscortText(id string) bool {
	p, ok := g.pack.OpeningPrelude()
	e, sceneOK := g.pack.OpeningScenePresentation()
	if !ok || !sceneOK || g.openingEscort == nil || g.openingEscort.DialoguePresentationID != p.ID {
		return false
	}
	codes, ok := g.pack.TextGlyphCodes(id)
	frame, frameOK := g.pack.TextGlyphCodes(p.FrameTextID)
	if !ok || !frameOK || !g.dlg.openRecord(codes) {
		return false
	}
	g.dlg.prelude, g.dlg.preludeFrame, g.dlg.layout = p, frame, p.Window
	g.dlg.shadow = &e.Shadow
	return true
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
// renderer fixtures. Production input walks home, selects pictures, approaches
// the mother manually, then follows the timed town sequence.
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
