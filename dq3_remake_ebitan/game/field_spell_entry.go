package game

// The reviewed empty entry is a message, with the native command canvas kept
// behind it. Learned spells and other actors retain their existing flow.
func (g *Game) beginSingleMemberEmptySpellEntry() bool {
	if len(g.companions) != 0 || g.heroHP <= 0 || g.heroConditions != 0 || len(g.fieldActorSpells(0)) != 0 {
		return false
	}
	g.fieldSpell = FieldSpellMenu{}
	if g.pack == nil || g.pack.Interface.FieldSpellEntry == nil {
		return true
	}
	s := g.pack.Interface.FieldSpellEntry
	actor := g.equipActorName(0)
	if s.ActorVariableCode == nil || len(actor) == 0 {
		return true
	}
	g.beginFieldMessagePrompt(s.EmptyTextID, s.Presentation, s.TextFlow, s.WaitIndicator, map[uint16][]int{uint16(*s.ActorVariableCode): append([]int(nil), actor...)})
	return true
}
