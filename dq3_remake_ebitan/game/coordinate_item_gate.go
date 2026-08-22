package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

const (
	coordinateGateIdle = iota
	coordinateGateFailureApproach
	coordinateGateSuccessApproach
	coordinateGateSuccessDialogue
)

func (g *Game) activeCoordinateItemGate() *gamepack.CoordinateItemGateEvent {
	if g.pack == nil || g.coordinateItemGateID == "" {
		return nil
	}
	event, _ := g.pack.CoordinateItemGateEvent(g.coordinateItemGateID)
	return event
}

func (g *Game) tryCoordinateItemGateEvent() bool {
	if g.pack == nil || g.inTown || g.coordinateItemGateStage != coordinateGateIdle ||
		g.coordinateForcedSteps != 0 {
		return false
	}
	event, ok := g.pack.CoordinateItemGateAt(g.px, g.py, g.layer)
	if !ok || !g.storyFlag(event.ActiveStoryFlagRaw) {
		return false
	}
	if !g.openPackText(event.DialogueTextIDs.Approach) {
		return false
	}
	g.coordinateItemGateID = event.ID
	if g.hasPartyItem(event.RequiredItemRawID) {
		g.coordinateItemGateStage = coordinateGateSuccessApproach
	} else {
		g.coordinateItemGateStage = coordinateGateFailureApproach
	}
	return true
}

func (g *Game) advanceCoordinateItemGateDialogue() {
	event := g.activeCoordinateItemGate()
	if event == nil {
		g.coordinateItemGateID = ""
		g.coordinateItemGateStage = coordinateGateIdle
		return
	}
	switch g.coordinateItemGateStage {
	case coordinateGateFailureApproach:
		g.addCoordinateGateDayNightSteps(event.FailureDayNightSteps)
		g.coordinateForcedDir = event.FailureDirection
		g.coordinateForcedSteps = event.FailureSteps
		g.coordinateItemGateStage = coordinateGateIdle
		g.coordinateItemGateID = ""
	case coordinateGateSuccessApproach:
		if g.openPackText(event.DialogueTextIDs.Success) {
			g.coordinateItemGateStage = coordinateGateSuccessDialogue
		}
	case coordinateGateSuccessDialogue:
		g.setStoryFlag(event.ClearStoryFlagRaw, false)
		if event.ConsumeRequiredItem {
			g.removePartyItems(event.RequiredItemRawID, 1)
		}
		g.coordinateItemGateStage = coordinateGateIdle
		g.coordinateItemGateID = ""
	}
}

func (g *Game) addCoordinateGateDayNightSteps(steps int) {
	for ; steps > 0; steps-- {
		g.advanceDaynight()
	}
}

func (g *Game) advanceCoordinateItemGateMovement() {
	g.facing = g.coordinateForcedDir
	dx, dy := dirDelta(g.coordinateForcedDir)
	nx, ny, valid := g.normalizeSurfaceTarget(g.px+dx, g.py+dy)
	if !valid || !g.tryMove(nx, ny) {
		g.coordinateForcedSteps = 0
		return
	}
	g.px, g.py = nx, ny
	g.coordinateForcedSteps--
	g.updateTrackedWorldObjectWindow(g.coordinateForcedDir)
}
