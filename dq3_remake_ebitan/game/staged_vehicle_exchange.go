package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

const (
	vehicleExchangeIdle = iota
	vehicleExchangeQuestIntro
	vehicleExchangeItemGrant
)

func (g *Game) activeStagedVehicleExchange() (*gamepack.StagedVehicleExchangeEvent, bool) {
	if g.pack == nil || g.vehicleExchangeEventID == "" {
		return nil, false
	}
	return g.pack.StagedVehicleExchangeEvent(g.vehicleExchangeEventID)
}

// talkStagedVehicleExchange runs a finite pack-declared quest-item then vehicle
// exchange. The engine knows only the common transaction; selectors, flags,
// items, dialogue and world coordinates belong to the selected game pack.
func (g *Game) talkStagedVehicleExchange(n *npcInst) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.StagedVehicleExchangeEvents() {
		if !g.scriptedNPCMatches(n, event.NPC) {
			continue
		}
		switch {
		case g.storyFlag(event.QuestPresentFlagRaw):
			if !g.openPackText(event.DialogueTextIDs.QuestIntro) {
				return true
			}
			g.vehicleExchangeEventID = event.ID
			g.vehicleExchangeStage = vehicleExchangeQuestIntro
		case g.shipOwned || !g.storyFlag(event.ExchangeAvailableFlagRaw):
			g.openPackText(event.DialogueTextIDs.After)
		case !g.hasItem(event.RequiredItemRawID):
			g.openPackText(event.DialogueTextIDs.NeedItem)
		default:
			// Resolve every visible success reference before mutating state.
			if !g.openPackText(event.DialogueTextIDs.Success) {
				return true
			}
			g.removeItems(event.RequiredItemRawID, 1)
			for _, flag := range event.Vehicle.ClearFlagsRaw {
				g.setStoryFlag(flag, false)
			}
			g.shipOwned, g.shipAboard = true, false
			g.shipX = event.Vehicle.WorldPosition.X
			g.shipY = event.Vehicle.WorldPosition.Y
			g.progressSet(msShip)
		}
		return true
	}
	return false
}

// advanceStagedVehicleExchangeDialogue applies the quest-item transaction only
// after the introductory record has closed, then opens the original grant
// record. Missing pack data leaves state untouched.
func (g *Game) advanceStagedVehicleExchangeDialogue() {
	event, ok := g.activeStagedVehicleExchange()
	if !ok {
		g.finishStagedVehicleExchange()
		return
	}
	switch g.vehicleExchangeStage {
	case vehicleExchangeQuestIntro:
		if !g.openPackText(event.DialogueTextIDs.ItemGrant) {
			g.finishStagedVehicleExchange()
			return
		}
		g.inventory = append(g.inventory, event.GrantedItemRawID)
		g.setStoryFlag(event.QuestPresentFlagRaw, false)
		g.noticeCode, g.noticeTimer = event.GrantedItemRawID, 120
		g.vehicleExchangeStage = vehicleExchangeItemGrant
	case vehicleExchangeItemGrant:
		g.finishStagedVehicleExchange()
	}
}

func (g *Game) finishStagedVehicleExchange() {
	g.vehicleExchangeStage = vehicleExchangeIdle
	g.vehicleExchangeEventID = ""
}
