package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	choiceItemExchangeIdle = iota
	choiceItemExchangeIntroduction
	choiceItemExchangeKnowledgeChoice
	choiceItemExchangeKnowledgeAnswer
	choiceItemExchangeOffer
	choiceItemExchangeOfferChoice
	choiceItemExchangeSuccess
	choiceItemExchangeAfter
	choiceItemExchangeReject
)

func (g *Game) activeChoiceItemExchange() (*gamepack.ChoiceItemExchangeEvent, bool) {
	if g.pack == nil || g.choiceItemExchangeEventID == "" {
		return nil, false
	}
	return g.pack.ChoiceItemExchangeEvent(g.choiceItemExchangeEventID)
}

// talkChoiceItemExchange starts the shared finite conversation. Pack data owns
// every selector, item, flag, coordinate and visible text reference.
func (g *Game) talkChoiceItemExchange(n *npcInst) bool {
	if g.pack == nil || g.choiceItemExchangeStage != choiceItemExchangeIdle {
		return false
	}
	for _, event := range g.pack.ChoiceItemExchangeEvents() {
		if !g.scriptedNPCMatches(n, event.NPC) {
			continue
		}
		if !g.storyFlag(event.AvailableFlagRaw) {
			g.openPackText(event.DialogueTextIDs.After)
			return true
		}
		g.choiceItemExchangeEventID = event.ID
		if g.hasItem(event.RequiredItemRawID) {
			if g.openPackText(event.DialogueTextIDs.Offer) {
				g.choiceItemExchangeStage = choiceItemExchangeOffer
			} else {
				g.finishChoiceItemExchange()
			}
			return true
		}
		if g.openPackText(event.DialogueTextIDs.Introduction) {
			g.choiceItemExchangeStage = choiceItemExchangeIntroduction
		} else {
			g.finishChoiceItemExchange()
		}
		return true
	}
	return false
}

func (g *Game) choiceItemExchangeChoosing() bool {
	return g.choiceItemExchangeStage == choiceItemExchangeKnowledgeChoice ||
		g.choiceItemExchangeStage == choiceItemExchangeOfferChoice
}

func (g *Game) advanceChoiceItemExchangeDialogue() {
	event, ok := g.activeChoiceItemExchange()
	if !ok {
		g.finishChoiceItemExchange()
		return
	}
	switch g.choiceItemExchangeStage {
	case choiceItemExchangeIntroduction:
		g.choiceItemExchangeStage, g.choiceItemExchangeCursor = choiceItemExchangeKnowledgeChoice, 0
	case choiceItemExchangeKnowledgeAnswer, choiceItemExchangeReject, choiceItemExchangeAfter:
		g.finishChoiceItemExchange()
	case choiceItemExchangeOffer:
		g.choiceItemExchangeStage, g.choiceItemExchangeCursor = choiceItemExchangeOfferChoice, 0
	case choiceItemExchangeSuccess:
		if g.openPackText(event.DialogueTextIDs.After) {
			g.choiceItemExchangeStage = choiceItemExchangeAfter
		} else {
			g.finishChoiceItemExchange()
		}
	}
}

func (g *Game) choiceItemExchangeChoiceInput(in InputState) {
	event, ok := g.activeChoiceItemExchange()
	if !ok {
		g.finishChoiceItemExchange()
		return
	}
	switch {
	case in.DirEdge == 0 || in.DirEdge == 1:
		g.choiceItemExchangeCursor ^= 1
	case in.Confirm && g.choiceItemExchangeStage == choiceItemExchangeKnowledgeChoice:
		textID := event.DialogueTextIDs.Interested
		if g.choiceItemExchangeCursor != 0 {
			textID = event.DialogueTextIDs.Hint
		}
		if g.openPackText(textID) {
			g.choiceItemExchangeStage = choiceItemExchangeKnowledgeAnswer
		}
	case in.Confirm && g.choiceItemExchangeStage == choiceItemExchangeOfferChoice &&
		g.choiceItemExchangeCursor == 0:
		g.completeChoiceItemExchange(event)
	case in.Confirm || in.Cancel:
		textID := event.DialogueTextIDs.Hint
		stage := choiceItemExchangeKnowledgeAnswer
		if g.choiceItemExchangeStage == choiceItemExchangeOfferChoice {
			textID = event.DialogueTextIDs.Reject
			stage = choiceItemExchangeReject
		}
		if g.openPackText(textID) {
			g.choiceItemExchangeStage = stage
		}
	}
}

func (g *Game) completeChoiceItemExchange(event *gamepack.ChoiceItemExchangeEvent) {
	// Resolve the first visible success record before mutating persistent state.
	if !g.openPackText(event.DialogueTextIDs.Success) {
		return
	}
	for i, item := range g.inventory {
		if item != event.RequiredItemRawID {
			continue
		}
		g.inventory[i] = event.GrantedItemRawID
		g.overPx, g.overPy = event.SuccessWorldPosition.X, event.SuccessWorldPosition.Y
		g.layer = event.SuccessWorldPosition.Layer
		g.worldState |= uint16(event.SetWorldStateMaskRaw)
		for _, flag := range event.ClearStoryFlagsRaw {
			g.setStoryFlag(flag, false)
		}
		g.choiceItemExchangeStage = choiceItemExchangeSuccess
		return
	}
	// Inventory changed while the choice was open: fail closed and leave every
	// other field untouched.
	g.dlg.open, g.dlg.buf, g.dlg.pos = false, nil, 0
	g.finishChoiceItemExchange()
}

func (g *Game) finishChoiceItemExchange() {
	g.choiceItemExchangeStage = choiceItemExchangeIdle
	g.choiceItemExchangeCursor = 0
	g.choiceItemExchangeEventID = ""
}

func (g *Game) drawChoiceItemExchangeChoice(rgba []byte, white dq3data.Color) {
	if !g.choiceItemExchangeChoosing() {
		return
	}
	event, ok := g.activeChoiceItemExchange()
	if !ok {
		return
	}
	g.drawBinaryChoice(rgba, white, g.choiceItemExchangeCursor,
		event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo)
}
