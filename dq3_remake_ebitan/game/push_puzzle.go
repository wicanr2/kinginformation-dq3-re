package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// tryPushableNPC applies the pack-defined original control-bit rule to any
// collided runtime NPC. The original collision consumer moves the NPC one cell
// first; only then may the player enter its former cell.
func (g *Game) tryPushableNPC(nx, ny int) bool {
	if g.pack == nil || !g.inTown || g.cur == nil {
		return false
	}
	idx := g.cur.npcAt(nx, ny)
	if idx < 0 {
		return false
	}
	rule := g.pack.NPCPushRule()
	if g.cur.npcs[idx].ctrl&rule.CtrlMask != rule.CtrlValue {
		return false
	}
	if rule.BlockedByEffectID == "temporary_invisibility" && g.remoaru > 0 {
		return false
	}
	dx, dy := nx-g.px, ny-g.py
	if (dx == 0) == (dy == 0) || dx < -1 || dx > 1 || dy < -1 || dy > 1 {
		return false
	}
	tx, ty := nx+dx, ny+dy
	if tx < 0 || ty < 0 || tx >= g.cur.w || ty >= g.cur.h ||
		g.cur.attr.Blocked(g.cur.tileIdx(tx, ty)) || g.cur.npcAt(tx, ty) >= 0 {
		return false
	}
	n := &g.cur.npcs[idx]
	n.x, n.y = tx, ty
	return true
}

// tryPushPuzzleCompletionEvent runs after a successful player step. It mirrors
// the original finite handler: a validated trigger invokes the slot/axis
// completion check, then clears the pack-defined story flag and refreshes the
// affected event tiles.
func (g *Game) tryPushPuzzleCompletionEvent() bool {
	if g.pack == nil || !g.inTown || g.cur == nil {
		return false
	}
	for _, event := range g.pack.PushPuzzleEvents() {
		if event.CTYRaw != g.curCty || event.Section != g.cur.sec ||
			!g.storyFlag(event.ClearStoryFlagRaw) {
			continue
		}
		if !pushPuzzleTriggerMatches(g.cur, event, g.px, g.py) {
			continue
		}
		for _, participant := range event.NPCs {
			if participant.Slot < 0 || participant.Slot >= len(g.cur.npcs) ||
				event.CompletionAxis != "y" ||
				g.cur.npcs[participant.Slot].y != event.CompletionValue {
				return true
			}
		}
		g.setStoryFlag(event.ClearStoryFlagRaw, false)
		g.cur.applyClearedEventTiles(g.storyFlag)
		return true
	}
	return false
}

func pushPuzzleTriggerMatches(sc *Scene, event gamepack.PushPuzzleEvent, x, y int) bool {
	for _, trigger := range event.TriggerTiles {
		if trigger.Tile.X == x && trigger.Tile.Y == y &&
			trigger.TileSubID >= 1 && trigger.TileSubID <= len(sc.specialHandlers) &&
			sc.specialHandlers[trigger.TileSubID-1] == trigger.HandlerRaw {
			return true
		}
	}
	return false
}
