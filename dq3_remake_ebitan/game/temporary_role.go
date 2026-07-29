package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	temporaryRoleIdle = iota
	temporaryRoleQuestGreeting
	temporaryRoleReturnPraise
	temporaryRoleForcedOffer
	temporaryRoleForcedChoice
	temporaryRoleForcedReject
	temporaryRoleAccept
	temporaryRoleLaterIntro
	temporaryRoleLaterOffer
	temporaryRoleLaterChoice
	temporaryRoleLaterReject
	temporaryRoleRestoreIntro
	temporaryRoleRestoreChoice
	temporaryRoleRestoreReconsider
	temporaryRoleRestoreConfirm
	temporaryRoleRestoreReject
	temporaryRoleRestoreContinue
	temporaryRoleRestoreAccept
)

func (g *Game) activeTemporaryRole() (*gamepack.TemporaryRoleEvent, bool) {
	if g.pack == nil || g.temporaryRoleEventID == "" {
		return nil, false
	}
	return g.pack.TemporaryRoleEvent(g.temporaryRoleEventID)
}

func (g *Game) scriptedNPCMatches(n *npcInst, s gamepack.ScriptedNPCSelector) bool {
	return g.inTown && g.cur != nil && n != nil &&
		g.curCty == s.CTYRaw && g.cur.sec == s.Section &&
		n.x == s.Tile.X && n.y == s.Tile.Y && n.b4 == s.HandlerRaw
}

// talkTemporaryRole dispatches a pack-owned required-item/temporary-role event.
// Missing text or references fail closed before any inventory or flag mutation.
func (g *Game) talkTemporaryRole(n *npcInst) bool {
	if g.pack == nil || g.temporaryRoleStage != temporaryRoleIdle {
		return false
	}
	for _, event := range g.pack.TemporaryRoleEvents() {
		switch {
		case g.scriptedNPCMatches(n, event.OfferNPC):
			g.temporaryRoleEventID = event.ID
			switch {
			case g.storyFlag(event.PendingFlagRaw) &&
				g.countItem(event.RequiredItemRawID) == 0:
				if !g.openPackText(event.DialogueTextIDs.QuestGreeting) {
					g.temporaryRoleEventID = ""
					return true
				}
				g.temporaryRoleStage = temporaryRoleQuestGreeting
			case g.storyFlag(event.PendingFlagRaw):
				if !g.openPackText(event.DialogueTextIDs.ReturnPraise) {
					g.temporaryRoleEventID = ""
					return true
				}
				// DQ3.EXE handler9 performs the transaction before rec49.
				g.removeItems(event.RequiredItemRawID, 1)
				g.setStoryFlag(event.PendingFlagRaw, false)
				g.temporaryRoleStage = temporaryRoleReturnPraise
			default:
				if !g.openPackText(event.DialogueTextIDs.LaterIntro) {
					g.temporaryRoleEventID = ""
					return true
				}
				g.temporaryRoleStage = temporaryRoleLaterIntro
			}
			return true
		case g.scriptedNPCMatches(n, event.RestoreNPC):
			if !g.storyFlag(event.ActiveRoleFlagRaw) {
				return true
			}
			g.temporaryRoleEventID = event.ID
			if !g.openPackText(event.DialogueTextIDs.RestoreIntro) {
				g.temporaryRoleEventID = ""
				return true
			}
			g.temporaryRoleStage = temporaryRoleRestoreIntro
			return true
		}
	}
	return false
}

func (g *Game) advanceTemporaryRoleDialogue() {
	event, ok := g.activeTemporaryRole()
	if !ok {
		g.temporaryRoleStage = temporaryRoleIdle
		return
	}
	open := func(id string, next int) {
		if g.openPackText(id) {
			g.temporaryRoleStage = next
			return
		}
		g.finishTemporaryRoleEvent()
	}
	switch g.temporaryRoleStage {
	case temporaryRoleQuestGreeting:
		open(event.DialogueTextIDs.QuestRequest, temporaryRoleIdle)
		if g.temporaryRoleStage == temporaryRoleIdle {
			g.temporaryRoleEventID = ""
		}
	case temporaryRoleReturnPraise:
		open(event.DialogueTextIDs.ForcedOffer, temporaryRoleForcedOffer)
	case temporaryRoleForcedOffer:
		g.temporaryRoleStage, g.temporaryRoleCursor = temporaryRoleForcedChoice, 0
	case temporaryRoleForcedReject:
		open(event.DialogueTextIDs.ForcedOffer, temporaryRoleForcedOffer)
	case temporaryRoleAccept:
		g.setStoryFlag(event.NormalRoleFlagRaw, false)
		g.setStoryFlag(event.ActiveRoleFlagRaw, true)
		g.syncTemporaryRoleVisual()
		g.finishTemporaryRoleEvent()
		g.reloadTownDaynight()
	case temporaryRoleLaterIntro:
		open(event.DialogueTextIDs.LaterOffer, temporaryRoleLaterOffer)
	case temporaryRoleLaterOffer:
		g.temporaryRoleStage, g.temporaryRoleCursor = temporaryRoleLaterChoice, 0
	case temporaryRoleLaterReject, temporaryRoleRestoreContinue:
		g.finishTemporaryRoleEvent()
	case temporaryRoleRestoreIntro:
		g.temporaryRoleStage, g.temporaryRoleCursor = temporaryRoleRestoreChoice, 0
	case temporaryRoleRestoreReconsider:
		g.temporaryRoleStage, g.temporaryRoleCursor = temporaryRoleRestoreConfirm, 0
	case temporaryRoleRestoreReject:
		open(event.DialogueTextIDs.ContinueRole, temporaryRoleRestoreContinue)
	case temporaryRoleRestoreAccept:
		g.setStoryFlag(event.ActiveRoleFlagRaw, false)
		g.setStoryFlag(event.NormalRoleFlagRaw, true)
		g.syncTemporaryRoleVisual()
		g.finishTemporaryRoleEvent()
		g.reloadTownDaynight()
	}
}

func (g *Game) temporaryRoleChoiceInput(in InputState) {
	event, ok := g.activeTemporaryRole()
	if !ok {
		g.finishTemporaryRoleEvent()
		return
	}
	yes := g.temporaryRoleCursor == 0
	if in.DirEdge == 0 || in.DirEdge == 1 {
		g.temporaryRoleCursor ^= 1
		return
	}
	if !in.Confirm && !in.Cancel {
		return
	}
	if in.Cancel {
		yes = false
	}
	open := func(id string, next int) {
		if g.openPackText(id) {
			g.temporaryRoleStage = next
		} else {
			g.finishTemporaryRoleEvent()
		}
	}
	switch g.temporaryRoleStage {
	case temporaryRoleForcedChoice:
		if yes {
			open(event.DialogueTextIDs.Accept, temporaryRoleAccept)
		} else {
			open(event.DialogueTextIDs.ForcedReject, temporaryRoleForcedReject)
		}
	case temporaryRoleLaterChoice:
		if yes {
			open(event.DialogueTextIDs.Accept, temporaryRoleAccept)
		} else {
			open(event.DialogueTextIDs.LaterReject, temporaryRoleLaterReject)
		}
	case temporaryRoleRestoreChoice:
		if yes {
			open(event.DialogueTextIDs.ContinueRole, temporaryRoleRestoreContinue)
		} else {
			open(event.DialogueTextIDs.RestoreReconsider, temporaryRoleRestoreReconsider)
		}
	case temporaryRoleRestoreConfirm:
		if yes {
			open(event.DialogueTextIDs.RestoreAccept, temporaryRoleRestoreAccept)
		} else {
			open(event.DialogueTextIDs.RestoreReject, temporaryRoleRestoreReject)
		}
	}
}

func (g *Game) finishTemporaryRoleEvent() {
	g.temporaryRoleStage = temporaryRoleIdle
	g.temporaryRoleCursor = 0
	g.temporaryRoleEventID = ""
}

func (g *Game) temporaryRoleChoosing() bool {
	switch g.temporaryRoleStage {
	case temporaryRoleForcedChoice, temporaryRoleLaterChoice,
		temporaryRoleRestoreChoice, temporaryRoleRestoreConfirm:
		return true
	}
	return false
}

func (g *Game) drawTemporaryRoleChoice(rgba []byte, white dq3data.Color) {
	if !g.temporaryRoleChoosing() {
		return
	}
	event, ok := g.activeTemporaryRole()
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
		if i == g.temporaryRoleCursor {
			drawGlyph(rgba, g.dlg.tx, x+12, yy, curGlyph, white)
		}
		for j, glyph := range label {
			drawGlyph(rgba, g.dlg.tx, x+40+j*16, yy, int(glyph), white)
		}
	}
}

// syncTemporaryRoleVisual derives the runtime sprite solely from canonical
// story flags. Save files therefore need no duplicate role/sprite state.
func (g *Game) syncTemporaryRoleVisual() {
	g.heroRole = nil
	if g.pack == nil {
		return
	}
	for _, event := range g.pack.TemporaryRoleEvents() {
		if !g.storyFlag(event.ActiveRoleFlagRaw) {
			continue
		}
		entry := event.RoleSpriteEntries.Male
		if g.heroGender == 1 {
			entry = event.RoleSpriteEntries.Female
		}
		g.heroRole = dq3data.LoadCharSprite(g.manBLS, entry)
		return
	}
}

func (g *Game) heroSceneSprite() *dq3data.CharSprite {
	if g.heroRole != nil {
		return g.heroRole
	}
	return g.hero
}
