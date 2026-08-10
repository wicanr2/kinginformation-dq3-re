package game

// opening_cutscene.go implements the shared boot-presentation state machine.
// The frame list, timing and asset identity are pack data; this file does not
// know DQ3 filenames or story text.

func (g *Game) openingReady() bool {
	return g.openingSeq != nil && len(g.openingSeq.Frames) > 0 &&
		len(g.openingPix) == len(g.openingSeq.Frames) &&
		g.openingIndex >= 0 && g.openingIndex < len(g.openingPix) &&
		len(g.openingPix[g.openingIndex]) == ScreenW*ScreenH
}

// StartOpeningCutscene is the normal desktop/mobile boot entry. Tests that
// start directly at a title fixture can omit it; production main calls it
// immediately after NewGame and before the first Ebiten frame.
func (g *Game) StartOpeningCutscene() {
	if !g.showTitle || !g.openingReady() {
		return
	}
	g.openingActive = true
	g.openingIndex = 0
	g.openingFrame = 0
}

func (g *Game) stopOpeningCutscene() {
	g.openingActive = false
	g.openingIndex = 0
	g.openingFrame = 0
}

func (g *Game) advanceOpeningCutscene() {
	if !g.openingReady() {
		g.stopOpeningCutscene()
		return
	}
	frame := g.openingSeq.Frames[g.openingIndex]
	g.openingFrame++
	if g.openingFrame < frame.HoldFrames {
		return
	}
	g.openingFrame = 0
	g.openingIndex++
	if g.openingIndex >= len(g.openingSeq.Frames) {
		g.stopOpeningCutscene()
	}
}

func (g *Game) openingInput(in InputState) bool {
	if !g.openingActive {
		return false
	}
	if g.openingSeq != nil && g.openingSeq.SkipOnInput && titleInputInterruptsAttract(in) {
		g.stopOpeningCutscene()
		return false // process the same key as title splash input
	}
	g.advanceOpeningCutscene()
	return true
}
