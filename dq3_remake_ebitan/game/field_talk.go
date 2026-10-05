package game

// Only the reviewed single healthy member on foot in a town is accepted.
// Targeted NPC and counter interactions retain their existing consumers.
func (g *Game) beginEmptyTalk() {
	if g.pack == nil || g.pack.Interface.FieldTalk == nil || g.cur == nil || !g.inTown || len(g.companions) != 0 || g.heroHP <= 0 || g.heroConditions != 0 || g.shipAboard {
		return
	}
	s := g.pack.Interface.FieldTalk
	g.beginFieldMessageRecords([]string{s.EmptyTextID}, s.Presentation, s.TextFlow, s.WaitIndicator, nil)
}
