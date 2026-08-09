package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	sequenceGateIdle = iota
	sequenceGatePrompt
	sequenceGateChoice
	sequenceGatePressed
	sequenceGateTrap
	sequenceGateSuccess
)

func (g *Game) activeSequenceGate() (*gamepack.TwoStepFloorSwitchGate, bool) {
	if g.pack == nil || g.sequenceGateEventID == "" {
		return nil, false
	}
	return g.pack.TwoStepFloorSwitchGate(g.sequenceGateEventID)
}

// trySequenceGateEvent consumes only pack-declared walk-on tiles. The raw
// handler numbers remain evidence anchors in the pack and are not interpreted
// by the engine.
func (g *Game) trySequenceGateEvent() bool {
	if g.pack == nil || !g.inTown || g.cur == nil ||
		g.sequenceGateStage != sequenceGateIdle {
		return false
	}
	for _, event := range g.pack.TwoStepFloorSwitchGates() {
		if event.CTYRaw != g.curCty || event.Section != g.cur.sec ||
			!g.storyFlag(event.ClearFlagRaw) {
			continue
		}
		for _, sw := range event.Switches {
			if sw.Tile.X != g.px || sw.Tile.Y != g.py {
				continue
			}
			i := g.py*g.cur.w + g.px
			if i < 0 || i >= len(g.cur.hiMap) ||
				int(g.cur.hiMap[i]&0x1f) != sw.TileSubID {
				continue
			}
			if !g.openPackText(event.DialogueTextIDs.Prompt) {
				return false
			}
			g.sequenceGateEventID = event.ID
			g.sequenceGateSubID = sw.TileSubID
			g.sequenceGateStage = sequenceGatePrompt
			return true
		}
	}
	return false
}

func (g *Game) sequenceGateChoosing() bool {
	return g.sequenceGateStage == sequenceGateChoice
}

func (g *Game) sequenceGateChoiceInput(in InputState) {
	event, ok := g.activeSequenceGate()
	if !ok {
		g.finishSequenceGateEvent(false)
		return
	}
	if in.DirEdge == 0 || in.DirEdge == 1 {
		g.sequenceGateCursor ^= 1
		return
	}
	if !in.Confirm && !in.Cancel {
		return
	}
	yes := g.sequenceGateCursor == 0 && !in.Cancel
	if !yes {
		g.finishSequenceGateEvent(false)
		return
	}
	if !g.openPackText(event.DialogueTextIDs.Pressed) {
		g.finishSequenceGateEvent(false)
		return
	}
	g.sequenceGateStage = sequenceGatePressed
}

func (g *Game) advanceSequenceGateDialogue() {
	event, ok := g.activeSequenceGate()
	if !ok {
		return
	}
	switch g.sequenceGateStage {
	case sequenceGatePrompt:
		g.sequenceGateStage, g.sequenceGateCursor = sequenceGateChoice, 0
	case sequenceGatePressed:
		if !g.sequenceGateArmed {
			g.sequenceGateArmed = true
			g.finishSequenceGateEvent(true)
			return
		}
		if g.sequenceGateSubID == event.CompletionTileSubID {
			if !g.openPackText(event.DialogueTextIDs.Success) {
				g.finishSequenceGateEvent(false)
				return
			}
			g.setStoryFlag(event.ClearFlagRaw, false)
			g.cur.applyClearedEventTiles(g.storyFlag)
			g.music.PlaySFX(event.SuccessSFXRaw)
			g.sequenceGateArmed = false
			g.sequenceGateStage = sequenceGateSuccess
			return
		}
		if !g.openPackText(event.DialogueTextIDs.Trap) {
			g.finishSequenceGateEvent(false)
			return
		}
		g.sequenceGateArmed = false
		g.sequenceGateStage = sequenceGateTrap
	case sequenceGateTrap:
		d := event.TrapDestination
		g.finishSequenceGateEvent(false)
		g.loadPackSceneDestination(d)
	case sequenceGateSuccess:
		g.finishSequenceGateEvent(false)
	}
}

func (g *Game) finishSequenceGateEvent(keepArmed bool) {
	g.sequenceGateStage = sequenceGateIdle
	g.sequenceGateCursor = 0
	g.sequenceGateEventID = ""
	g.sequenceGateSubID = 0
	if !keepArmed {
		g.sequenceGateArmed = false
	}
}

// loadPackSceneDestination applies a validated declarative scene destination.
// It intentionally accepts no DQ-specific constants or callback expressions.
func (g *Game) loadPackSceneDestination(d gamepack.SceneDestination) bool {
	blkn := 1
	if d.CTYRaw >= 0 && d.CTYRaw < len(mapBlkNum) {
		blkn = mapBlkNum[d.CTYRaw]
	}
	ns, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, d.CTYRaw, blkn,
		d.Section, g.dnPhase, g.storyFlag)
	if err != nil || d.X >= ns.w || d.Y >= ns.h {
		return false
	}
	cross := d.CTYRaw != g.curCty
	g.town, g.cur, g.curCty = ns, ns, d.CTYRaw
	g.px, g.py = d.X, d.Y
	if ns.dlgText != nil {
		g.dlg.tx = ns.dlgText
	}
	if cross {
		g.playSceneMusic(d.CTYRaw)
	}
	g.cd = moveCooldown
	return true
}

func (g *Game) drawSequenceGateChoice(rgba []byte, white dq3data.Color) {
	if !g.sequenceGateChoosing() {
		return
	}
	event, ok := g.activeSequenceGate()
	if !ok {
		return
	}
	x, y, w, h := 430, 220, 120, 68
	fillBox(rgba, x, y, w, h, white)
	ids := [2]string{event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo}
	for i, id := range ids {
		label, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		yy := y + 12 + i*24
		if i == g.sequenceGateCursor {
			drawGlyph(rgba, g.dlg.tx, x+12, yy, curGlyph, white)
		}
		for j, glyph := range label {
			drawGlyph(rgba, g.dlg.tx, x+40+j*16, yy, int(glyph), white)
		}
	}
}
