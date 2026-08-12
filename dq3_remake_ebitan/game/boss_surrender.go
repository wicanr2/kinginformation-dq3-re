package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	bossSurrenderIdle = iota
	bossSurrenderIntro
	bossSurrenderBattle
	bossSurrenderApology
	bossSurrenderChoice
	bossSurrenderReject
	bossSurrenderFarewell
)

func (g *Game) activeBossSurrender() (*gamepack.BossSurrenderEvent, bool) {
	if g.pack == nil || g.bossSurrenderEventID == "" {
		return nil, false
	}
	return g.pack.BossSurrenderEvent(g.bossSurrenderEventID)
}

func (g *Game) tryBossSurrenderEvent() bool {
	if g.pack == nil || !g.inTown || g.cur == nil ||
		g.bossSurrenderStage != bossSurrenderIdle {
		return false
	}
	for _, event := range g.pack.BossSurrenderEvents() {
		if event.Trigger.CTYRaw != g.curCty || event.Trigger.Section != g.cur.sec ||
			!g.storyFlag(event.PresenceFlagRaw) {
			continue
		}
		for _, tile := range event.Trigger.Tiles {
			if tile.X != g.px || tile.Y != g.py {
				continue
			}
			if g.cur.hiMap == nil ||
				int(g.cur.hiMap[g.py*g.cur.w+g.px]&0x1f) != event.Trigger.TileSubID {
				continue
			}
			g.bossSurrenderEventID = event.ID
			if !g.openPackText(event.DialogueTextIDs.Intro) {
				g.bossSurrenderEventID = ""
				return false
			}
			g.bossSurrenderStage = bossSurrenderIntro
			return true
		}
	}
	return false
}

func (g *Game) startBossSurrenderBattle() bool {
	event, ok := g.activeBossSurrender()
	if !ok {
		return false
	}
	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, maxMP, true
	}
	hp := heroParams{
		level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countPartyItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells(),
	}
	groups := make([]enemyGroup, len(event.Formation.Groups))
	for i, group := range event.Formation.Groups {
		groups[i] = enemyGroup{monID: group.MonsterRawID, count: group.Count}
	}
	g.battle.lightOrb = false
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	if !g.battle.startFormationWithBackground(groups, int64(g.anim)*2654+1, hp, g.buildCompanionActors(), &event.Formation.Background) {
		return false
	}
	g.playAudioCue(audioCueBattle)
	return true
}

func (g *Game) advanceBossSurrenderDialogue() {
	event, ok := g.activeBossSurrender()
	if !ok {
		g.bossSurrenderStage = bossSurrenderIdle
		return
	}
	switch g.bossSurrenderStage {
	case bossSurrenderIntro:
		if g.startBossSurrenderBattle() {
			g.bossSurrenderStage = bossSurrenderBattle
		} else {
			g.bossSurrenderStage = bossSurrenderIdle
		}
	case bossSurrenderApology:
		g.bossSurrenderStage, g.bossSurrenderCursor = bossSurrenderChoice, 0
	case bossSurrenderReject:
		g.bossSurrenderStage, g.bossSurrenderCursor = bossSurrenderChoice, 0
	case bossSurrenderFarewell:
		g.setStoryFlag(event.ClearFlagRaw, false)
		g.bossSurrenderStage = bossSurrenderIdle
		g.bossSurrenderEventID = ""
		g.reloadTownDaynight()
	}
}

func (g *Game) settleBossSurrenderBattle() {
	if g.bossSurrenderStage != bossSurrenderBattle {
		return
	}
	event, ok := g.activeBossSurrender()
	if !ok {
		g.bossSurrenderStage = bossSurrenderIdle
		return
	}
	if g.battle.result == 1 {
		if !g.openPackText(event.DialogueTextIDs.Apology) {
			g.bossSurrenderStage = bossSurrenderIdle
			g.bossSurrenderEventID = ""
			return
		}
		g.bossSurrenderStage = bossSurrenderApology
		return
	}
	g.bossSurrenderStage = bossSurrenderIdle
	g.bossSurrenderEventID = ""
}

func (g *Game) bossSurrenderChoiceInput(in InputState) {
	event, ok := g.activeBossSurrender()
	if !ok {
		g.bossSurrenderStage = bossSurrenderIdle
		return
	}
	switch {
	case in.Confirm && g.bossSurrenderCursor == 0:
		if g.openPackText(event.DialogueTextIDs.Accept) {
			g.bossSurrenderStage = bossSurrenderFarewell
		}
	case in.Confirm || in.Cancel:
		if g.openPackText(event.DialogueTextIDs.Reject) {
			g.bossSurrenderStage = bossSurrenderReject
		}
	case in.DirEdge == 0 || in.DirEdge == 1:
		g.bossSurrenderCursor ^= 1
	}
}

func (g *Game) treasureBlockedByPack(cty, section, x, y, itemRawID int) bool {
	if g.pack == nil {
		return false
	}
	for _, event := range g.pack.BossSurrenderEvents() {
		for _, gate := range event.TreasureGates {
			if gate.CTYRaw == cty && gate.Section == section &&
				gate.Tile.X == x && gate.Tile.Y == y && gate.ItemRawID == itemRawID &&
				g.storyFlag(gate.WhileFlagSet) {
				return true
			}
		}
	}
	for _, event := range g.pack.StagedBossEvents() {
		for _, gate := range event.TreasureGates {
			if gate.CTYRaw == cty && gate.Section == section &&
				gate.Tile.X == x && gate.Tile.Y == y && gate.ItemRawID == itemRawID &&
				g.storyFlag(gate.WhileFlagSet) {
				return true
			}
		}
	}
	return false
}

func (g *Game) drawBossSurrenderChoice(rgba []byte, white dq3data.Color) {
	if g.bossSurrenderStage != bossSurrenderChoice {
		return
	}
	event, ok := g.activeBossSurrender()
	if !ok {
		return
	}
	g.drawBinaryChoice(rgba, white, g.bossSurrenderCursor,
		event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo)
}

// drawBinaryChoice is the shared renderer for the existing two-option event UI.
// Its legacy geometry still requires a separate original-layout RE slice before
// it may be promoted into the game-pack interface contract.
func (g *Game) drawBinaryChoice(rgba []byte, white dq3data.Color, cursor int, yesID, noID string) {
	x, y, w, h := 430, 220, 120, 68
	fillBox(rgba, x, y, w, h, white)
	ids := [2]string{yesID, noID}
	for i, id := range ids {
		label, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		yy := y + 12 + i*24
		if i == cursor {
			drawGlyph(rgba, g.dlg.tx, x+12, yy, curGlyph, white)
		}
		for j, glyph := range label {
			drawGlyph(rgba, g.dlg.tx, x+40+j*16, yy, int(glyph), white)
		}
	}
}
