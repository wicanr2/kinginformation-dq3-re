package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// tryTrackingGuardEvent applies the finite pack-defined guard primitive after
// a successful player step. A visible player causes the guard to mirror the
// player's X coordinate; temporary invisibility suppresses the movement.
func (g *Game) tryTrackingGuardEvent() bool {
	if g.pack == nil || !g.inTown || g.cur == nil {
		return false
	}
	for _, event := range g.pack.TrackingGuardEvents() {
		if event.CTYRaw != g.curCty || event.Section != g.cur.sec {
			continue
		}
		matched := false
		for _, trigger := range event.TriggerTiles {
			if trigger.Tile.X != g.px || trigger.Tile.Y != g.py ||
				trigger.TileSubID < 1 || trigger.TileSubID > len(g.cur.specialHandlers) ||
				g.cur.specialHandlers[trigger.TileSubID-1] != trigger.HandlerRaw {
				continue
			}
			matched = true
			break
		}
		if !matched {
			continue
		}
		if event.BypassEffectID == "temporary_invisibility" && g.remoaru > 0 {
			return true
		}
		if event.TrackAxis != "horizontal" {
			return true
		}
		idx := trackingGuardNPC(g.cur, event)
		if idx < 0 || g.cur.npcs[idx].x == g.px {
			return true
		}
		n := &g.cur.npcs[idx]
		nx := n.x - 1
		if n.x < g.px {
			nx = n.x + 1
		}
		if nx < 0 || nx >= g.cur.w || g.cur.attr.Blocked(g.cur.tileIdx(nx, n.y)) ||
			(g.cur.npcAt(nx, n.y) >= 0 && g.cur.npcAt(nx, n.y) != idx) {
			return true
		}
		n.x = nx
		return true
	}
	return false
}

func trackingGuardNPC(sc *Scene, event gamepack.TrackingGuardEvent) int {
	for i := range sc.npcs {
		n := &sc.npcs[i]
		if n.b4 == event.Guard.HandlerRaw && n.y == event.Guard.Tile.Y {
			return i
		}
	}
	return -1
}
