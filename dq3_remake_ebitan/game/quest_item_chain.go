package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

func (g *Game) questTreasureFor(cty, section, subid int) (*gamepack.QuestItemChainEvent, bool) {
	if g.pack == nil {
		return nil, false
	}
	for _, event := range g.pack.QuestItemChainEvents() {
		t := event.Treasure
		if t.CTYRaw == cty && t.Section == section && t.TileSubID == subid {
			found, ok := g.pack.QuestItemChainEvent(event.ID)
			return found, ok
		}
	}
	return nil, false
}

// collectQuestTreasure consumes the original present flag. Unlike the legacy
// baked treasure table, this follows the original convention exactly:
// story flag set = available, clear = already collected.
func (g *Game) collectQuestTreasure(cty, section, subid int) (handled bool) {
	event, ok := g.questTreasureFor(cty, section, subid)
	if !ok {
		return false
	}
	t := event.Treasure
	if !g.storyFlag(t.PresentFlag) {
		return true
	}
	g.setStoryFlag(t.PresentFlag, false)
	g.inventory = append(g.inventory, t.ItemRawID)
	g.noticeCode, g.noticeTimer = t.ItemRawID, 120
	return true
}

// talkQuestItemChain performs a finite required-item replacement selected by
// pack-owned CTY/NPC data. Missing text fails closed before inventory mutation.
func (g *Game) talkQuestItemChain(n *npcInst) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.QuestItemChainEvents() {
		x := event.Exchange
		if !g.scriptedNPCMatches(n, x.NPC) {
			continue
		}
		if !g.hasItem(x.RequiredItemRawID) {
			g.openPackText(x.BeforeTextID)
			return true
		}
		if !g.openPackText(x.SuccessTextID) {
			return true
		}
		for i, item := range g.inventory {
			if item == x.RequiredItemRawID {
				// Original handler16 replaces the matched inventory word in
				// place (DQ3.EXE file 0x690a), preserving slot order.
				g.inventory[i] = x.GrantedItemRawID
				return true
			}
		}
		return true
	}
	return false
}

// useQuestItemChain handles a pack-owned location-gated quest item. Returning
// true means the item belongs to this primitive even when its location gate
// fails, so no unrelated fallback may consume it.
func (g *Game) useQuestItemChain(item int) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.QuestItemChainEvents() {
		u := event.Use
		if u.ItemRawID != item {
			continue
		}
		if u.LocationKind != "town" || !g.inTown || g.curCty != u.CTYRaw {
			return true
		}
		codes, ok := g.pack.TextGlyphCodes(u.SuccessTextID)
		if !ok {
			return true
		}
		g.removeItems(item, 1)
		for _, flag := range u.SetFlagsRaw {
			g.setStoryFlag(flag, true)
		}
		for _, flag := range u.ClearFlagsRaw {
			g.setStoryFlag(flag, false)
		}
		g.panel = panelNone
		g.clampPanelCursor()
		if u.ReloadScene {
			g.reloadTownDaynight()
		}
		g.dlg.openRecord(codes)
		return true
	}
	return false
}
