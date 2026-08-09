package game

// attract.go keeps the title idle loop as a data-driven presentation state.
// The sequence, asset keys, and timing belong to the versioned game pack;
// this file only supplies the shared finite-state mechanics.

// titleInputInterruptsAttract reports an edge or held action that should wake
// the title screen. DirHeld is included because a held d-pad is still player
// intent even when no new edge was produced this frame.
func titleInputInterruptsAttract(in InputState) bool {
	return in.DirHeld >= 0 || in.DirEdge >= 0 || in.Confirm || in.Cancel ||
		in.Enter || in.Toggle || in.Settings || in.CtxTap || in.Tapped
}

func (g *Game) attractReady() bool {
	return g.attractSeq != nil && len(g.attractSeq.Frames) > 0 &&
		len(g.attractPix) == len(g.attractSeq.Frames) &&
		g.attractIndex >= 0 && g.attractIndex < len(g.attractPix) &&
		len(g.attractPix[g.attractIndex]) > 0
}

func (g *Game) startAttract() {
	if g.attractSeq == nil || len(g.attractSeq.Frames) == 0 ||
		len(g.attractPix) != len(g.attractSeq.Frames) {
		return
	}
	g.attractActive = true
	g.attractIndex = 0
	g.attractFrame = 0
}

func (g *Game) stopAttract() {
	g.attractActive = false
	g.attractIndex = 0
	g.attractFrame = 0
	g.titleIdleFrames = 0
}

func (g *Game) advanceAttract() {
	if !g.attractReady() {
		g.stopAttract()
		return
	}
	frame := g.attractSeq.Frames[g.attractIndex]
	g.attractFrame++
	if g.attractFrame < frame.HoldFrames {
		return
	}
	g.attractFrame = 0
	g.attractIndex = (g.attractIndex + 1) % len(g.attractSeq.Frames)
}
