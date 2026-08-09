package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	hostageRescueIdle = iota
	hostageRescueGuardQuestion
	hostageRescueGuardChoice
	hostageRescueGuardJoin
	hostageRescueGuardFight
	hostageRescueGuardBattle
	hostageRescueApproachCry
	hostageRescueApproachMove
	hostageRescueSwitchPrompt
	hostageRescueSwitchChoice
	hostageRescueSwitchPressed
	hostageRescueReunion
	hostageRescueReturnHome
	hostageRescueCaptiveMove
	hostageRescueBossArrival
	hostageRescueBossChallenge
	hostageRescueBossBattle
	hostageRescueBossDefeated
	hostageRescueBossApology
	hostageRescueBossChoice
	hostageRescueBossReject
	hostageRescueBossFarewell
	hostageRescueRewardOffer
	hostageRescueRewardChoice
	hostageRescueRewardDecline
	hostageRescueRewardReceived
)

func (g *Game) activeHostageRescue() (*gamepack.HostageRescueEvent, bool) {
	if g.pack == nil || g.hostageRescueEventID == "" {
		return nil, false
	}
	return g.pack.HostageRescueEvent(g.hostageRescueEventID)
}

func (g *Game) beginHostageRescue(event gamepack.HostageRescueEvent) {
	g.hostageRescueEventID = event.ID
	g.hostageRescueCursor = 0
}

func (g *Game) talkHostageRescue(n *npcInst) bool {
	if g.pack == nil || g.hostageRescueStage != hostageRescueIdle {
		return false
	}
	for _, event := range g.pack.HostageRescueEvents() {
		for _, selector := range event.GuardNPCs {
			if g.scriptedNPCMatches(n, selector) && g.storyFlag(event.GuardPresentFlagRaw) {
				g.beginHostageRescue(event)
				if g.openPackText(event.DialogueTextIDs.GuardQuestion) {
					g.hostageRescueStage = hostageRescueGuardQuestion
				}
				return true
			}
		}
		for _, selector := range event.BossNPCs {
			if g.scriptedNPCMatches(n, selector) && g.storyFlag(event.BossPresentFlagRaw) {
				g.beginHostageRescue(event)
				if g.openPackText(event.DialogueTextIDs.BossChallenge) {
					g.hostageRescueStage = hostageRescueBossChallenge
				}
				return true
			}
		}
		if g.scriptedNPCMatches(n, event.RewardNPC) {
			g.beginHostageRescue(event)
			if !g.storyFlag(event.RewardAvailableFlagRaw) {
				g.openPackText(event.DialogueTextIDs.RewardDecline)
				g.hostageRescueStage = hostageRescueRewardDecline
				return true
			}
			if g.openPackText(event.DialogueTextIDs.RewardOffer) {
				g.hostageRescueStage = hostageRescueRewardOffer
			}
			return true
		}
	}
	return false
}

func (g *Game) hostageRescueChoiceInput(in InputState) {
	event, ok := g.activeHostageRescue()
	if !ok {
		g.finishHostageRescue()
		return
	}
	switch {
	case in.DirEdge == 0 || in.DirEdge == 1:
		g.hostageRescueCursor ^= 1
		return
	case !in.Confirm && !in.Cancel:
		return
	}
	yes := in.Confirm && g.hostageRescueCursor == 0
	switch g.hostageRescueStage {
	case hostageRescueGuardChoice:
		if yes {
			g.openPackText(event.DialogueTextIDs.GuardJoin)
			g.hostageRescueStage = hostageRescueGuardJoin
		} else {
			g.openPackText(event.DialogueTextIDs.GuardFight)
			g.hostageRescueStage = hostageRescueGuardFight
		}
	case hostageRescueSwitchChoice:
		if yes {
			g.openPackText(event.DialogueTextIDs.SwitchPressed)
			g.hostageRescueStage = hostageRescueSwitchPressed
		} else {
			g.finishHostageRescue()
		}
	case hostageRescueBossChoice:
		if yes {
			g.openPackText(event.DialogueTextIDs.BossFarewell)
			g.hostageRescueStage = hostageRescueBossFarewell
		} else {
			g.openPackText(event.DialogueTextIDs.BossReject)
			g.hostageRescueStage = hostageRescueBossReject
		}
	case hostageRescueRewardChoice:
		if yes {
			g.inventory = append(g.inventory, event.RewardItemRawID)
			g.setStoryFlag(event.RewardAvailableFlagRaw, false)
			g.noticeCode, g.noticeTimer = event.RewardItemRawID, 120
			g.openPackText(event.DialogueTextIDs.RewardReceived)
			g.hostageRescueStage = hostageRescueRewardReceived
		} else {
			g.openPackText(event.DialogueTextIDs.RewardDecline)
			g.hostageRescueStage = hostageRescueRewardDecline
		}
	}
}

func (g *Game) startHostageFormation(formation gamepack.BattleFormation) bool {
	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, maxMP, true
	}
	hp := heroParams{
		level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countPartyItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells(),
	}
	groups := make([]enemyGroup, len(formation.Groups))
	for i, group := range formation.Groups {
		groups[i] = enemyGroup{monID: group.MonsterRawID, count: group.Count}
	}
	g.battle.lightOrb = false
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	if !g.battle.startFormation(groups, int64(g.anim)*2654+1, hp, g.buildCompanionActors()) {
		return false
	}
	g.playAudioCue(audioCueBattle)
	return true
}

func (g *Game) advanceHostageRescueDialogue() {
	event, ok := g.activeHostageRescue()
	if !ok {
		return
	}
	switch g.hostageRescueStage {
	case hostageRescueGuardQuestion:
		g.hostageRescueStage, g.hostageRescueCursor = hostageRescueGuardChoice, 0
	case hostageRescueGuardJoin, hostageRescueRewardDecline, hostageRescueRewardReceived:
		g.finishHostageRescue()
	case hostageRescueGuardFight:
		if g.startHostageFormation(event.GuardFormation) {
			g.hostageRescueStage = hostageRescueGuardBattle
		} else {
			g.finishHostageRescue()
		}
	case hostageRescueApproachCry:
		g.startHostageMovement(event.ApproachMovement, hostageRescueApproachMove)
	case hostageRescueSwitchPrompt:
		g.hostageRescueStage, g.hostageRescueCursor = hostageRescueSwitchChoice, 0
	case hostageRescueSwitchPressed:
		g.openPackText(event.DialogueTextIDs.Reunion)
		g.hostageRescueStage = hostageRescueReunion
	case hostageRescueReunion:
		g.openPackText(event.DialogueTextIDs.ReturnHome)
		g.hostageRescueStage = hostageRescueReturnHome
	case hostageRescueReturnHome:
		g.hostageRescueMovement = 0
		g.startHostageMovement(event.CaptiveMovements[0], hostageRescueCaptiveMove)
	case hostageRescueBossArrival:
		g.setStoryFlag(event.CaptivePresentFlagRaw, false)
		g.setStoryFlag(event.BossPresentFlagRaw, true)
		g.finishHostageRescue()
		g.reloadTownDaynight()
	case hostageRescueBossChallenge:
		if g.startHostageFormation(event.BossFormation) {
			g.hostageRescueStage = hostageRescueBossBattle
		} else {
			g.finishHostageRescue()
		}
	case hostageRescueBossDefeated:
		g.openPackText(event.DialogueTextIDs.BossApology)
		g.hostageRescueStage = hostageRescueBossApology
	case hostageRescueBossApology, hostageRescueBossReject:
		g.hostageRescueStage, g.hostageRescueCursor = hostageRescueBossChoice, 0
	case hostageRescueBossFarewell:
		for _, flag := range event.CompletionClearFlagsRaw {
			g.setStoryFlag(flag, false)
		}
		for _, flag := range event.CompletionSetFlagsRaw {
			g.setStoryFlag(flag, true)
		}
		g.finishHostageRescue()
		g.reloadTownDaynight()
	case hostageRescueRewardOffer:
		g.hostageRescueStage, g.hostageRescueCursor = hostageRescueRewardChoice, 0
	}
}

func (g *Game) settleHostageRescueBattle() {
	event, ok := g.activeHostageRescue()
	if !ok {
		return
	}
	switch g.hostageRescueStage {
	case hostageRescueGuardBattle:
		if g.battle.result == 1 {
			g.setStoryFlag(event.GuardPresentFlagRaw, false)
			g.reloadTownDaynight()
		}
		g.finishHostageRescue()
	case hostageRescueBossBattle:
		if g.battle.result == 1 && g.openPackText(event.DialogueTextIDs.BossDefeated) {
			g.hostageRescueStage = hostageRescueBossDefeated
		} else {
			g.finishHostageRescue()
		}
	}
}

func sceneTriggerMatches(g *Game, trigger gamepack.SceneTileTrigger) bool {
	if g.cur == nil || g.curCty != trigger.CTYRaw || g.cur.sec != trigger.Section {
		return false
	}
	for _, tile := range trigger.Tiles {
		if tile.X == g.px && tile.Y == g.py && g.px >= 0 && g.py >= 0 &&
			g.px < g.cur.w && g.py < g.cur.h &&
			int(g.cur.hiMap[g.py*g.cur.w+g.px]&0x1f) == trigger.TileSubID {
			return true
		}
	}
	return false
}

func (g *Game) tryHostageRescueTrigger() bool {
	if g.pack == nil || !g.inTown || g.hostageRescueStage != hostageRescueIdle {
		return false
	}
	for _, event := range g.pack.HostageRescueEvents() {
		switch {
		case g.storyFlag(event.ApproachPendingFlagRaw) &&
			sceneTriggerMatches(g, event.ApproachTrigger):
			g.beginHostageRescue(event)
			if g.openPackText(event.DialogueTextIDs.RescueCry) {
				g.hostageRescueStage = hostageRescueApproachCry
			}
			return true
		case g.storyFlag(event.CaptivePresentFlagRaw) &&
			sceneTriggerMatches(g, event.SwitchTrigger):
			g.beginHostageRescue(event)
			if g.openPackText(event.DialogueTextIDs.SwitchPrompt) {
				g.hostageRescueStage = hostageRescueSwitchPrompt
			}
			return true
		}
	}
	return false
}

func (g *Game) findHostageNPC(selector gamepack.ScriptedNPCSelector) int {
	if g.cur == nil || g.curCty != selector.CTYRaw || g.cur.sec != selector.Section {
		return -1
	}
	for i := range g.cur.npcs {
		n := &g.cur.npcs[i]
		if n.x == selector.Tile.X && n.y == selector.Tile.Y && n.b4 == selector.HandlerRaw {
			return i
		}
	}
	return -1
}

func (g *Game) startHostageMovement(movement gamepack.ScriptedMovement, stage int) {
	index := g.findHostageNPC(movement.NPC)
	if index < 0 && g.cur != nil && len(movement.Path) > 0 {
		start := movement.Path[0]
		for i := range g.cur.npcs {
			n := &g.cur.npcs[i]
			if n.x == start.X && n.y == start.Y && n.b4 == movement.NPC.HandlerRaw {
				index = i
				break
			}
		}
	}
	if index < 0 || len(movement.Path) < 2 {
		g.finishHostageRescue()
		return
	}
	g.hostageRescueNPC = index
	g.hostageRescueWaypoint = 1
	g.hostageRescueTick = 0
	g.hostageRescueStage = stage
}

func (g *Game) hostageRescueAnimating() bool {
	return g.hostageRescueStage == hostageRescueApproachMove ||
		g.hostageRescueStage == hostageRescueCaptiveMove
}

func (g *Game) advanceHostageRescueAnimation() {
	event, ok := g.activeHostageRescue()
	if !ok || g.cur == nil || g.hostageRescueNPC < 0 ||
		g.hostageRescueNPC >= len(g.cur.npcs) {
		g.finishHostageRescue()
		return
	}
	var movement gamepack.ScriptedMovement
	if g.hostageRescueStage == hostageRescueApproachMove {
		movement = event.ApproachMovement
	} else if g.hostageRescueMovement < len(event.CaptiveMovements) {
		movement = event.CaptiveMovements[g.hostageRescueMovement]
	} else {
		g.finishHostageRescue()
		return
	}
	g.hostageRescueTick++
	if g.hostageRescueTick < event.StepFrames {
		return
	}
	g.hostageRescueTick = 0
	n := &g.cur.npcs[g.hostageRescueNPC]
	target := movement.Path[g.hostageRescueWaypoint]
	facing, arrived := passageStep(&n.x, &n.y, target)
	n.facing, n.walk = facing, n.walk^1
	if arrived {
		g.hostageRescueWaypoint++
	}
	if g.hostageRescueWaypoint < len(movement.Path) {
		return
	}
	if facing, ok := passageFacing(movement.EndFacing); ok {
		n.facing = facing
	}
	if g.hostageRescueStage == hostageRescueApproachMove {
		g.setStoryFlag(event.ApproachPendingFlagRaw, false)
		g.finishHostageRescue()
		return
	}
	g.hostageRescueMovement++
	if g.hostageRescueMovement < len(event.CaptiveMovements) {
		g.startHostageMovement(event.CaptiveMovements[g.hostageRescueMovement],
			hostageRescueCaptiveMove)
		return
	}
	if g.openPackText(event.DialogueTextIDs.BossArrival) {
		g.hostageRescueStage = hostageRescueBossArrival
	}
}

func (g *Game) finishHostageRescue() {
	g.hostageRescueStage = hostageRescueIdle
	g.hostageRescueCursor = 0
	g.hostageRescueEventID = ""
	g.hostageRescueMovement = 0
	g.hostageRescueWaypoint = 0
	g.hostageRescueTick = 0
	g.hostageRescueNPC = -1
}

func (g *Game) drawHostageRescueChoice(rgba []byte, white dq3data.Color) {
	switch g.hostageRescueStage {
	case hostageRescueGuardChoice, hostageRescueSwitchChoice,
		hostageRescueBossChoice, hostageRescueRewardChoice:
	default:
		return
	}
	event, ok := g.activeHostageRescue()
	if !ok {
		return
	}
	x, y, w, h := 430, 220, 120, 68
	fillBox(rgba, x, y, w, h, white)
	for i, id := range []string{event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo} {
		label, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		yy := y + 12 + i*24
		if i == g.hostageRescueCursor {
			drawGlyph(rgba, g.dlg.tx, x+12, yy, curGlyph, white)
		}
		for j, glyph := range label {
			drawGlyph(rgba, g.dlg.tx, x+40+j*16, yy, int(glyph), white)
		}
	}
}
