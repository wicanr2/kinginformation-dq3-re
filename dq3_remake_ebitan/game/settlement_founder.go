package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	settlementFounderIdle = iota
	settlementFounderIntroduction
	settlementFounderNoEligible
	settlementFounderFirstOffer
	settlementFounderFirstChoice
	settlementFounderReject
	settlementFounderFinalOffer
	settlementFounderFinalChoice
	settlementFounderAccepted
	settlementFounderStorageFull
)

func (g *Game) activeSettlementFounder() (*gamepack.SettlementFounderEvent, bool) {
	if g.pack == nil || g.settlementFounderEventID == "" {
		return nil, false
	}
	return g.pack.SettlementFounderEvent(g.settlementFounderEventID)
}

func (g *Game) talkSettlementFounder(n *npcInst) bool {
	if g.pack == nil || g.cur == nil || g.settlementFounderStage != settlementFounderIdle {
		return false
	}
	for _, event := range g.pack.SettlementFounderEvents() {
		if event.NPC.CTYRaw != g.curCty || event.NPC.Section != g.cur.sec ||
			event.NPC.Tile.X != n.x || event.NPC.Tile.Y != n.y || event.NPC.HandlerRaw != n.b4 {
			continue
		}
		if g.storyFlag(event.CompletionFlagRaw) {
			return g.openPackText(event.DialogueTextIDs.After)
		}
		if !g.openPackText(event.DialogueTextIDs.Introduction) {
			return false
		}
		g.settlementFounderEventID = event.ID
		g.settlementFounderStage = settlementFounderIntroduction
		g.settlementFounderCursor, g.settlementFounderMember = 0, -1
		for i, member := range g.companions {
			if member.Class == event.RequiredClassRaw {
				// The original stops at the first class match; a dead first merchant
				// does not make a later merchant eligible.
				if !event.RequireAlive || member.Alive() {
					g.settlementFounderMember = i
				}
				break
			}
		}
		return true
	}
	return false
}

func (g *Game) settlementFounderChoosing() bool {
	return g.settlementFounderStage == settlementFounderFirstChoice ||
		g.settlementFounderStage == settlementFounderFinalChoice
}

func (g *Game) openSettlementFounderText(id string) bool {
	if !g.openPackText(id) {
		return false
	}
	if i := g.settlementFounderMember; i >= 0 && i < len(g.companions) {
		g.setDlgVarGlyphs(dq3data.TxtVarEnt, g.companions[i].Name)
	}
	return true
}

func (g *Game) settlementFounderChoiceInput(in InputState) {
	event, ok := g.activeSettlementFounder()
	if !ok {
		g.resetSettlementFounder()
		return
	}
	switch {
	case in.DirEdge == 0 || in.DirEdge == 1:
		g.settlementFounderCursor ^= 1
	case in.Confirm && g.settlementFounderCursor == 0:
		if g.settlementFounderStage == settlementFounderFirstChoice {
			if g.openSettlementFounderText(event.DialogueTextIDs.FinalOffer) {
				g.settlementFounderStage = settlementFounderFinalOffer
			}
		} else if g.openSettlementFounderText(event.DialogueTextIDs.Accepted) {
			g.settlementFounderStage = settlementFounderAccepted
		}
	case in.Confirm || in.Cancel:
		if g.openPackText(event.DialogueTextIDs.Reject) {
			g.settlementFounderStage = settlementFounderReject
		}
	}
}

func (g *Game) advanceSettlementFounderDialogue() {
	event, ok := g.activeSettlementFounder()
	if !ok {
		return
	}
	switch g.settlementFounderStage {
	case settlementFounderIntroduction:
		if g.settlementFounderMember < 0 {
			if g.openPackText(event.DialogueTextIDs.NoEligible) {
				g.settlementFounderStage = settlementFounderNoEligible
				return
			}
		} else if g.openSettlementFounderText(event.DialogueTextIDs.FirstOffer) {
			g.settlementFounderStage = settlementFounderFirstOffer
			return
		}
		g.resetSettlementFounder()
	case settlementFounderFirstOffer:
		g.settlementFounderStage, g.settlementFounderCursor = settlementFounderFirstChoice, 0
	case settlementFounderFinalOffer:
		g.settlementFounderStage, g.settlementFounderCursor = settlementFounderFinalChoice, 0
	case settlementFounderAccepted:
		if !g.beginSettlementFounder(event) {
			g.resetSettlementFounder()
		}
	case settlementFounderStorageFull:
		g.settlementFounderOverflow--
		if g.settlementFounderOverflow > 0 {
			g.openPackText(event.DialogueTextIDs.StorageFull)
			return
		}
		g.finishSettlementFounder(event)
	case settlementFounderNoEligible, settlementFounderReject:
		g.resetSettlementFounder()
	}
}

func (g *Game) beginSettlementFounder(event *gamepack.SettlementFounderEvent) bool {
	i := g.settlementFounderMember
	if event == nil || i < 0 || i >= len(g.companions) ||
		g.storyFlag(event.CompletionFlagRaw) {
		return false
	}
	member := g.companions[i]
	if member.Class != event.RequiredClassRaw || event.RequireAlive && !member.Alive() {
		return false
	}
	items := append(member.equippedItems(), member.Inventory...)
	available := event.SharedStorageCapacity - len(g.sharedStorage)
	if available > len(items) {
		available = len(items)
	}
	if available > 0 {
		g.sharedStorage = append(g.sharedStorage, items[:available]...)
	}
	g.setStoryFlag(event.CompletionFlagRaw, true)
	g.settlementFounderOverflow = len(items) - available
	if g.settlementFounderOverflow > 0 {
		if !g.openPackText(event.DialogueTextIDs.StorageFull) {
			return false
		}
		g.settlementFounderStage = settlementFounderStorageFull
		return true
	}
	g.finishSettlementFounder(event)
	return true
}

func (g *Game) finishSettlementFounder(event *gamepack.SettlementFounderEvent) {
	i := g.settlementFounderMember
	if event == nil || i < 0 || i >= len(g.companions) {
		g.resetSettlementFounder()
		return
	}
	member := g.companions[i]
	founder := *member
	founder.Name = append([]int(nil), member.Name...)
	founder.Inventory = nil
	founder.LearnedSpells = append([]int(nil), member.LearnedSpells...)
	founder.Weapon, founder.Armor, founder.Shield, founder.Head = -1, -1, -1, -1
	g.settlementFounder = &founder
	g.companions = append(g.companions[:i], g.companions[i+1:]...)
	g.reloadTownDaynight()
	g.resetSettlementFounder()
}

func (g *Game) resetSettlementFounder() {
	g.settlementFounderStage = settlementFounderIdle
	g.settlementFounderCursor = 0
	g.settlementFounderEventID = ""
	g.settlementFounderMember = -1
	g.settlementFounderOverflow = 0
}

func (g *Game) drawSettlementFounderChoice(rgba []byte, white dq3data.Color) {
	if !g.settlementFounderChoosing() {
		return
	}
	event, ok := g.activeSettlementFounder()
	if !ok {
		return
	}
	g.drawBinaryChoice(rgba, white, g.settlementFounderCursor,
		event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo)
}
