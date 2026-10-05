package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"

// actorItemStore returns the sole writable owner. All displayed lists are copies.
func (g *Game) actorItemStore(actor int) *itemstore.Store {
	if actor == 0 {
		return &g.items
	}
	if actor < 1 || actor > len(g.companions) {
		return nil
	}
	return &g.companions[actor-1].Items
}
func (g *Game) actorItemEntries(actor int) []itemstore.Entry {
	s := g.actorItemStore(actor)
	if s == nil {
		return nil
	}
	return s.Entries()
}

// equipActorInventory remains a read-only derived view for UI adapters.
func (g *Game) equipActorInventory(actor int) *[]int {
	s := g.actorItemStore(actor)
	if s == nil {
		return nil
	}
	codes := s.Codes()
	return &codes
}
func (g *Game) equipActorSlots(actor int) *[4]int {
	s := g.actorItemStore(actor)
	if s == nil {
		return nil
	}
	equipment := s.Equipment()
	return &equipment
}
func (g *Game) actorItemCount(actor int) int { return len(g.actorItemEntries(actor)) }
func (g *Game) equipCandidateCount(actor int) int {
	s := g.actorItemStore(actor)
	if s == nil {
		return 0
	}
	return len(s.Inventory())
}
func (g *Game) selectedItemEntry() (itemstore.Entry, bool) {
	s := g.actorItemStore(g.panelActor)
	if s == nil {
		return itemstore.Entry{}, false
	}
	return s.At(g.itemSelected)
}
func (g *Game) selectPanelItem() bool {
	entries := g.actorItemEntries(g.panelActor)
	if g.panelCursor < 0 || g.panelCursor >= len(entries) {
		return false
	}
	g.itemSelected = entries[g.panelCursor].Position
	return true
}
