package game

// Only the reviewed healthy single-member on-foot response is presented.
// Existing event handlers keep their transaction and selection rules.
func (g *Game) beginEmptyExamine() {
	if g.pack == nil || g.pack.Interface.FieldExamine == nil || len(g.companions) != 0 || g.heroHP <= 0 || g.heroConditions != 0 || g.shipAboard {
		return
	}
	s := g.pack.Interface.FieldExamine
	actor := g.equipActorName(0)
	if s.ActorVariableCode == nil || len(actor) == 0 || s.RecordJoin != "new_line" {
		return
	}
	g.beginFieldMessageRecords([]string{s.IntroTextID, s.ResultTextID}, s.Presentation, s.TextFlow, s.WaitIndicator, map[uint16][]int{uint16(*s.ActorVariableCode): append([]int(nil), actor...)})
}
