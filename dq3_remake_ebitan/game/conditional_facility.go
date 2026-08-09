package game

// talkConditionalFacility handles the original scripted-NPC branch that is
// only a facility at one player coordinate.  It deliberately matches the
// pack selector before applying the coordinate gate; no DQ3 CTY/handler
// values are embedded in the shared engine.
func (g *Game) talkConditionalFacility(n *npcInst) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.ConditionalFacilityEvents() {
		if !g.scriptedNPCMatches(n, event.NPC) {
			continue
		}
		if g.py == event.RequiredPlayerY {
			g.openFacility(event.FacilityK)
			return true
		}
		// The original non-gated branch is a repeatable dialogue and has no
		// state transaction.  Missing pack text remains fail-closed in
		// openPackText rather than silently falling through to generic tables.
		_ = g.openPackText(event.FallbackTextID)
		return true
	}
	return false
}
