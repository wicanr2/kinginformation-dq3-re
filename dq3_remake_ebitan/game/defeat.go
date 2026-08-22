package game

const (
	defeatDialogueIdle = iota
	defeatDialogueParty
	defeatDialogueVoice
)

// respawnPoint is the typed remake equivalent of the original six raw
// saved-location words. The raw words and their address-space contract remain
// in docs/155; the engine stores only the cross-version player-visible state.
type respawnPoint struct {
	Valid               bool
	PX, PY              int
	OverPX, OverPY      int
	InTown              bool
	Cty, Section, Layer int
}

func (g *Game) currentRespawnPoint() respawnPoint {
	r := respawnPoint{
		Valid: true, PX: g.px, PY: g.py, OverPX: g.overPx, OverPY: g.overPy,
		InTown: g.inTown, Cty: g.curCty, Layer: g.layer,
	}
	if g.inTown && g.cur != nil {
		r.Section = g.cur.sec
	}
	return r
}

func (g *Game) partyAllDown() bool {
	if g.heroHP > 0 {
		return false
	}
	for _, member := range g.companions {
		if member != nil && member.CurHP > 0 {
			return false
		}
	}
	return true
}

func (g *Game) actorAlive(actor int) bool {
	if actor == 0 {
		return g.heroHP > 0
	}
	i := actor - 1
	return i >= 0 && i < len(g.companions) && g.companions[i] != nil && g.companions[i].CurHP > 0
}

// selectLivingPartyLeader mirrors sub_1BCF2's visible ordering rule: keep the
// current leader while alive; otherwise promote the first living party slot.
func (g *Game) selectLivingPartyLeader() {
	if g.actorAlive(g.partyLeader) {
		return
	}
	for actor := 0; actor <= len(g.companions); actor++ {
		if g.actorAlive(actor) {
			g.partyLeader = actor
			return
		}
	}
}

func (g *Game) defeatTextID(role string) string {
	if g.pack == nil {
		return ""
	}
	refs, ok := g.pack.BattleTextRefs()
	if !ok {
		return ""
	}
	return refs.IDs()[role]
}

// beginFieldDefeat starts the original record361 → record362 modal chain.
func (g *Game) beginFieldDefeat() {
	if g.defeatDialogueStage != defeatDialogueIdle {
		return
	}
	g.defeatDialogueStage = defeatDialogueParty
	if !g.openPackText(g.defeatTextID("party_defeated")) {
		// Pack validation normally makes this unreachable. Fail closed by keeping
		// the player in the modal stage instead of silently inventing text.
		return
	}
	g.dlg.heroName = append([]int(nil), g.heroName...)
}

// beginBattleDefeat continues after Battle already displayed record361.
func (g *Game) beginBattleDefeat() {
	if g.defeatDialogueStage != defeatDialogueIdle {
		return
	}
	g.applyDefeatRespawn()
	g.defeatDialogueStage = defeatDialogueVoice
	if !g.openPackText(g.defeatTextID("revival_voice")) {
		return
	}
}

func (g *Game) advanceDefeatDialogue() {
	switch g.defeatDialogueStage {
	case defeatDialogueParty:
		g.applyDefeatRespawn()
		g.defeatDialogueStage = defeatDialogueVoice
		if !g.openPackText(g.defeatTextID("revival_voice")) {
			return
		}
	case defeatDialogueVoice:
		g.defeatDialogueStage = defeatDialogueIdle
	}
}

func (g *Game) applyDefeatRespawn() {
	// sub_1C7D9 clears death and restores max HP only for the fixed first
	// character record. Other dead party members remain dead for church revive.
	_, maxHP, _, _, _ := g.heroStats()
	g.heroHP = maxHP
	g.partyLeader = 0

	r := g.respawn
	if !r.Valid {
		r = g.currentRespawnPoint()
		g.respawn = r
	}
	g.layer, g.overPx, g.overPy = r.Layer, r.OverPX, r.OverPY
	g.px, g.py, g.inTown, g.curCty = r.PX, r.PY, r.InTown, r.Cty
	g.shipAboard, g.phoenixAboard = false, false
	g.fieldSpell = FieldSpellMenu{}
	g.cmd.open, g.panel = false, panelNone
	g.resetPartyTrail()
	if r.InTown {
		g.restoreTownScene(r.Cty, r.Section)
	} else {
		g.cur = g.overworldScene()
		g.curCty = -1
	}
	g.applyDaynightPalette()
}
