package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// storyFlagRuntimeEvent resolves a named finite primitive from the versioned
// pack. Missing data is deliberately a no-op: the loader rejects the shipped
// production pack before gameplay, while lightweight fixtures remain
// fail-closed instead of inventing a DQ3 value.
func (g *Game) storyFlagRuntimeEvent(kind string) (*gamepack.StoryFlagRuntimeEvent, bool) {
	if g == nil || g.pack == nil {
		return nil, false
	}
	return g.pack.StoryFlagRuntimeEventKind(kind)
}

func storyFlagNPCMatches(e *gamepack.StoryFlagRuntimeEvent, cty, section int, n *npcInst) bool {
	return e != nil && n != nil && e.NPC.CTYRaw == cty && e.NPC.Section == section &&
		e.NPC.Tile.X == n.x && e.NPC.Tile.Y == n.y && e.NPC.HandlerRaw == n.b4
}

func (g *Game) applyStoryFlagLists(e *gamepack.StoryFlagRuntimeEvent) {
	if e == nil {
		return
	}
	for _, flag := range e.ClearStoryFlagsRaw {
		g.setStoryFlag(flag, false)
	}
	for _, flag := range e.SetStoryFlagsRaw {
		g.setStoryFlag(flag, true)
	}
}

func (g *Game) reloadStoryFlagScene(e *gamepack.StoryFlagRuntimeEvent) {
	if e == nil || !e.SceneReload {
		return
	}
	g.reloadTownDaynight()
}

// talkConditionalStoryFlag implements handler33/sub_15904. The raw dialogue
// record and all flag mutations come from the pack contract; the engine only
// supplies the finite “display → transaction → scene reload” primitive.
func (g *Game) talkConditionalStoryFlag(n *npcInst) bool {
	e, ok := g.storyFlagRuntimeEvent("conditional_story_flag")
	if !ok || !storyFlagNPCMatches(e, g.curCty, currentSceneSection(g.cur), n) {
		return false
	}
	// handler33 opens the original record before its flag transaction reloads
	// the NPC scene.  Dialogue owns a copy of the record buffer, so the
	// subsequent reload cannot invalidate the visible message.
	if !g.dlg.Open(e.DialogueRecordRaw) {
		return true
	}
	g.applyStoryFlagLists(e)
	g.reloadStoryFlagScene(e)
	return true
}

// applyBossAftermathStoryFlag implements the post-battle tail of handler70.
// The boss trigger itself remains an existing finite combat primitive; this
// function closes only the independently proven story-flag/day-night/world
// transaction after its final victory.
func (g *Game) applyBossAftermathStoryFlag(t *bossTrigger) {
	e, ok := g.storyFlagRuntimeEvent("boss_aftermath")
	if !ok || t == nil || e.NPC.CTYRaw != t.cty || e.NPC.Section != t.sec {
		return
	}
	matched := false
	for _, xy := range t.coords {
		if xy[0] == e.NPC.Tile.X && xy[1] == e.NPC.Tile.Y {
			matched = true
			break
		}
	}
	if !matched {
		return
	}
	g.applyStoryFlagLists(e)
	if e.ResetDayNightSteps {
		g.setDaynight(e.DayNightPhase)
	}
	if e.RebuildWorldScenes {
		// setStoryFlag invalidates the town cache; setDaynight refreshes the
		// active section and applyDaynightPalette updates both field layers.
		g.applyDaynightPalette()
	}
}

func (g *Game) phoenixRuntimeEvent() (*gamepack.StoryFlagRuntimeEvent, bool) {
	e, ok := g.storyFlagRuntimeEvent("phoenix_revival")
	if !ok || e.Animation == nil || e.MapCell == nil || e.ParkPosition == nil {
		return nil, false
	}
	return e, true
}

func (g *Game) isPhoenixGuardian(n *npcInst) bool {
	e, ok := g.storyFlagRuntimeEvent("phoenix_revival")
	return ok && storyFlagNPCMatches(e, g.curCty, currentSceneSection(g.cur), n)
}

// applyPhoenixMapCell applies handler69's confirmed byte write to a loaded
// CTY scene. The mutation is kept as a save bit and replayed after a reload;
// the pristine CTY file is never modified.
func (g *Game) applyPhoenixMapCell() {
	if g == nil || !g.phoenixMapCellWritten || g.cur == nil {
		return
	}
	e, ok := g.phoenixRuntimeEvent()
	if !ok || g.curCty != e.MapCell.CTYRaw || currentSceneSection(g.cur) != e.MapCell.Section {
		return
	}
	cell := e.MapCell.Tile
	if cell.X < 0 || cell.Y < 0 || cell.X >= g.cur.w || cell.Y >= g.cur.h {
		return
	}
	if g.cur.override == nil {
		g.cur.override = map[int]int{}
	}
	g.cur.override[cell.Y*g.cur.w+cell.X] = e.MapCell.ValueRaw
}
