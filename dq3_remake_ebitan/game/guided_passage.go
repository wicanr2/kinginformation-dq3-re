package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

const (
	guidedPassageIdle = iota
	guidedPassageIntroduction
	guidedPassageAcceptance
	guidedPassageGuideWalk
	guidedPassageAwaitTrigger
	guidedPassageCompletionDialogue
	guidedPassageGuideReturn
)

func (g *Game) activeGuidedPassage() (*gamepack.GuidedPassageEvent, bool) {
	if g.pack == nil || g.guidedPassageEventID == "" {
		return nil, false
	}
	return g.pack.GuidedPassageEvent(g.guidedPassageEventID)
}

// talkGuidedPassage starts a finite required-item/guided-passage event. The
// required item is resolved only after the introduction closes, matching the
// original record order; missing data never mutates state.
func (g *Game) talkGuidedPassage(n *npcInst) bool {
	if g.pack == nil || g.guidedPassageStage != guidedPassageIdle {
		return false
	}
	for _, event := range g.pack.GuidedPassageEvents() {
		if g.scriptedNPCMatches(n, event.CompletedNPC) &&
			g.storyFlag(event.CompletedFlagRaw) {
			g.openPackText(event.DialogueTextIDs.After)
			return true
		}
		if !g.scriptedNPCMatches(n, event.GuideNPC) {
			continue
		}
		if !g.storyFlag(event.GuidePresentFlagRaw) {
			g.openPackText(event.DialogueTextIDs.After)
			return true
		}
		if g.px != event.InteractionTile.X || g.py != event.InteractionTile.Y {
			// The original script owns a fixed route from the legal talk tile.
			// A mismatched entry must not teleport the player into that route.
			return true
		}
		if !g.openPackText(event.DialogueTextIDs.Introduction) {
			return true
		}
		g.guidedPassageEventID = event.ID
		g.guidedPassageStage = guidedPassageIntroduction
		g.guidedPassageNPC = g.sceneNPCIndex(n)
		return true
	}
	return false
}

func (g *Game) sceneNPCIndex(want *npcInst) int {
	if g.cur == nil || want == nil {
		return -1
	}
	for i := range g.cur.npcs {
		if &g.cur.npcs[i] == want {
			return i
		}
	}
	return -1
}

func (g *Game) advanceGuidedPassageDialogue() {
	event, ok := g.activeGuidedPassage()
	if !ok {
		g.finishGuidedPassage()
		return
	}
	switch g.guidedPassageStage {
	case guidedPassageIntroduction:
		if !g.hasItem(event.RequiredItemRawID) {
			g.finishGuidedPassage()
			return
		}
		if !g.openPackText(event.DialogueTextIDs.Acceptance) {
			g.finishGuidedPassage()
			return
		}
		g.guidedPassageStage = guidedPassageAcceptance
	case guidedPassageAcceptance:
		g.guidedPassageStage = guidedPassageGuideWalk
		g.guidedPassageWaypoint = 1
		g.guidedPassageTick = 0
	case guidedPassageCompletionDialogue:
		g.guidedPassageStage = guidedPassageGuideReturn
		g.guidedPassageWaypoint = 1
		g.guidedPassageTick = 0
	}
}

func (g *Game) guidedPassageAnimating() bool {
	return g.guidedPassageStage == guidedPassageGuideWalk ||
		g.guidedPassageStage == guidedPassageGuideReturn
}

func passageStep(x, y *int, target gamepack.TileCoordinate) (facing int, arrived bool) {
	switch {
	case *x < target.X:
		(*x)++
		facing = 3
	case *x > target.X:
		(*x)--
		facing = 2
	case *y < target.Y:
		(*y)++
		facing = 0
	case *y > target.Y:
		(*y)--
		facing = 1
	default:
		return 0, true
	}
	return facing, *x == target.X && *y == target.Y
}

func passageFacing(facing string) (int, bool) {
	switch facing {
	case "down":
		return 0, true
	case "up":
		return 1, true
	case "left":
		return 2, true
	case "right":
		return 3, true
	default:
		return 0, false
	}
}

func (g *Game) advanceGuidedPassageAnimation() {
	event, ok := g.activeGuidedPassage()
	if !ok || g.cur == nil || g.curCty != event.GuideNPC.CTYRaw ||
		g.cur.sec != event.GuideNPC.Section {
		g.finishGuidedPassage()
		return
	}
	g.guidedPassageTick++
	if g.guidedPassageTick < event.StepFrames {
		return
	}
	g.guidedPassageTick = 0

	switch g.guidedPassageStage {
	case guidedPassageGuideWalk:
		if g.guidedPassageNPC < 0 || g.guidedPassageNPC >= len(g.cur.npcs) {
			g.finishGuidedPassage()
			return
		}
		n := &g.cur.npcs[g.guidedPassageNPC]
		target := event.GuidePath[g.guidedPassageWaypoint]
		facing, arrived := passageStep(&n.x, &n.y, target)
		n.facing, n.walk = facing, n.walk^1
		if arrived {
			g.guidedPassageWaypoint++
		}
		if g.guidedPassageWaypoint >= len(event.GuidePath) {
			if facing, ok := passageFacing(event.GuideEndFacing); ok {
				n.facing = facing
			}
			g.setStoryFlag(event.AnimationPendingFlagRaw, true)
			// handler50 ends here and returns normal player control. The player
			// must follow the opened passage and step on the configured scene
			// trigger before handler57 may run.
			g.guidedPassageStage = guidedPassageAwaitTrigger
		}
	case guidedPassageGuideReturn:
		if g.guidedPassageNPC < 0 || g.guidedPassageNPC >= len(g.cur.npcs) {
			g.finishGuidedPassage()
			return
		}
		n := &g.cur.npcs[g.guidedPassageNPC]
		target := event.GuideReturnPath[g.guidedPassageWaypoint]
		facing, arrived := passageStep(&n.x, &n.y, target)
		n.facing, n.walk = facing, n.walk^1
		if arrived {
			g.guidedPassageWaypoint++
		}
		if g.guidedPassageWaypoint >= len(event.GuideReturnPath) {
			if facing, ok := passageFacing(event.GuideReturnEndFacing); ok {
				n.facing = facing
			}
			if event.ConsumeRequiredItem {
				g.removeItems(event.RequiredItemRawID, 1)
			}
			g.setStoryFlag(event.AnimationPendingFlagRaw, false)
			g.setStoryFlag(event.GuidePresentFlagRaw, false)
			g.setStoryFlag(event.CompletedFlagRaw, true)
			g.finishGuidedPassage()
			g.reloadTownDaynight()
		}
	}
}

// tryGuidedPassageTrigger dispatches the pack-owned equivalent of the CTY
// section+4 special-handler table. It is called only after a real player move;
// a pending flag alone cannot teleport or auto-complete the passage.
func (g *Game) tryGuidedPassageTrigger() bool {
	if g.pack == nil || g.cur == nil || !g.inTown ||
		g.guidedPassageStage != guidedPassageAwaitTrigger {
		return false
	}
	event, ok := g.activeGuidedPassage()
	if !ok || g.curCty != event.CompletionTrigger.CTYRaw ||
		g.cur.sec != event.CompletionTrigger.Section ||
		!g.storyFlag(event.AnimationPendingFlagRaw) {
		return false
	}
	onTile := false
	for _, tile := range event.CompletionTrigger.Tiles {
		if g.px == tile.X && g.py == tile.Y {
			onTile = true
			break
		}
	}
	if !onTile {
		return false
	}
	i := g.py*g.cur.w + g.px
	if i < 0 || i >= len(g.cur.hiMap) ||
		int(g.cur.hiMap[i]&0x1f) != event.CompletionTrigger.TileSubID ||
		event.CompletionTrigger.TileSubID > len(g.cur.specialHandlers) ||
		g.cur.specialHandlers[event.CompletionTrigger.TileSubID-1] != event.SceneHandlerRaw {
		return false
	}
	if !g.openPackText(event.DialogueTextIDs.GuideReady) {
		return false
	}
	g.guidedPassageStage = guidedPassageCompletionDialogue
	return true
}

func (g *Game) finishGuidedPassage() {
	g.guidedPassageStage = guidedPassageIdle
	g.guidedPassageEventID = ""
	g.guidedPassageNPC = -1
	g.guidedPassageWaypoint = 0
	g.guidedPassageTick = 0
}
