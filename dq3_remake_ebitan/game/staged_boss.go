package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	stagedBossIdle = iota
	stagedBossFirstBattle
	stagedBossFirstMoving
	stagedBossFirstPost
	stagedBossSecondOffer
	stagedBossSecondChoice
	stagedBossSecondAccept
	stagedBossSecondChallenge
	stagedBossSecondBattle
	stagedBossSecondVictory
)

func (g *Game) activeStagedBoss() (*gamepack.StagedBossEvent, bool) {
	if g.pack == nil || g.stagedBossEventID == "" {
		return nil, false
	}
	return g.pack.StagedBossEvent(g.stagedBossEventID)
}

func (g *Game) startPackFormation(f gamepack.BattleFormation, suppressDrop bool) bool {
	level, maxHP, atk, def, agi := g.heroStats()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, g.heroMaxMP(), true
	}
	groups := make([]enemyGroup, len(f.Groups))
	for i, group := range f.Groups {
		groups[i] = enemyGroup{monID: group.MonsterRawID, count: group.Count}
	}
	if bg, ok := dq3data.DecodePackBG(g.battle.scr, f.BackgroundRaw); ok {
		g.battle.bg = bg
	} else {
		return false
	}
	hp := heroParams{
		level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countPartyItem(herbCode), mp: g.heroMP, maxMP: g.heroMaxMP(), spells: g.heroSpells(),
	}
	if !g.battle.startFormation(groups, int64(g.anim)*2654+1, hp, g.buildCompanionActors()) {
		return false
	}
	g.battle.suppressDrop = suppressDrop
	g.playAudioCue(audioCueBattle)
	return true
}

func (g *Game) talkStagedBoss(n *npcInst) bool {
	if g.pack == nil || !g.inTown || g.cur == nil || n == nil || g.stagedBossStage != stagedBossIdle {
		return false
	}
	for _, event := range g.pack.StagedBossEvents() {
		sel := event.FirstNPC
		if sel.CTYRaw == g.curCty && sel.Section == g.cur.sec && sel.Tile.X == n.x &&
			sel.Tile.Y == n.y && sel.HandlerRaw == n.b4 && g.storyFlag(event.FirstPresenceFlagRaw) {
			g.stagedBossEventID = event.ID
			if !g.startPackFormation(event.FirstFormation, event.FirstDropPolicy == "suppress") {
				g.stagedBossEventID = ""
				return false
			}
			g.stagedBossStage = stagedBossFirstBattle
			return true
		}

		sel = event.SecondNPC
		if sel.CTYRaw != g.curCty || sel.Section != g.cur.sec || sel.Tile.X != n.x ||
			sel.Tile.Y != n.y || sel.HandlerRaw != n.b4 || !g.storyFlag(event.SecondRequiredFlagRaw) {
			continue
		}
		g.stagedBossEventID = event.ID
		if !g.openPackText(event.DialogueTextIDs.SecondOffer) {
			g.stagedBossEventID = ""
			return false
		}
		g.stagedBossStage = stagedBossSecondOffer
		return true
	}
	return false
}

func setPackFlags(g *Game, clearFlags, setFlags []int) {
	for _, flag := range clearFlags {
		g.setStoryFlag(flag, false)
	}
	for _, flag := range setFlags {
		g.setStoryFlag(flag, true)
	}
}

func (g *Game) settleStagedBossBattle() {
	event, ok := g.activeStagedBoss()
	if !ok {
		return
	}
	switch g.stagedBossStage {
	case stagedBossFirstBattle:
		if g.battle.result != 1 {
			g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
			return
		}
		g.stagedBossStage = stagedBossFirstMoving
		g.stagedBossMoveIndex, g.stagedBossMoveTick = 0, 0
		g.stagedBossMovingNPC = true
	case stagedBossSecondBattle:
		if g.battle.result != 1 {
			g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
			return
		}
		setPackFlags(g, event.SecondVictoryClearFlagsRaw, event.SecondVictorySetFlagsRaw)
		g.reloadTownDaynight()
		if g.openPackText(event.DialogueTextIDs.SecondVictory) {
			g.stagedBossStage = stagedBossSecondVictory
		} else {
			g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
		}
	}
}

func stagedBossDirection(name string) int {
	return map[string]int{"down": 0, "up": 1, "left": 2, "right": 3}[name]
}

func (g *Game) advanceStagedBossMovement() {
	event, ok := g.activeStagedBoss()
	if !ok || g.stagedBossStage != stagedBossFirstMoving {
		return
	}
	// 原版先顯示怪物掉落，再進 scripted NPC／玩家移動與 rec70；不能讓取得道具
	// overlay 蓋住戰後對話。staged modal 會提早 return，因此在此消費通知倒數。
	if g.noticeTimer > 0 {
		g.noticeTimer--
		return
	}
	g.stagedBossMoveTick++
	if g.stagedBossMoveTick < event.StepFrames {
		return
	}
	g.stagedBossMoveTick = 0
	if g.stagedBossMovingNPC {
		if g.stagedBossMoveIndex < len(event.FirstNPCPostBattleDirections) {
			dir := stagedBossDirection(event.FirstNPCPostBattleDirections[g.stagedBossMoveIndex])
			for i := range g.cur.npcs {
				n := &g.cur.npcs[i]
				if n.b4 != event.FirstNPC.HandlerRaw {
					continue
				}
				dx, dy := dirDelta(dir)
				n.x, n.y = n.x+dx, n.y+dy
				n.ctrl = n.ctrl&^3 | dir
				break
			}
			g.stagedBossMoveIndex++
			return
		}
		setPackFlags(g, event.FirstVictoryClearFlagsRaw, event.FirstVictorySetFlagsRaw)
		g.reloadTownDaynight()
		g.stagedBossMovingNPC = false
		g.stagedBossMoveIndex = 0
		return
	}
	if g.stagedBossMoveIndex < len(event.FirstPostBattleDirections) {
		dir := stagedBossDirection(event.FirstPostBattleDirections[g.stagedBossMoveIndex])
		g.facing = dir
		dx, dy := dirDelta(dir)
		if g.tryMove(g.px+dx, g.py+dy) {
			g.px, g.py = g.px+dx, g.py+dy
			g.tryTransition()
		}
		g.stagedBossMoveIndex++
		return
	}
	if g.openPackText(event.DialogueTextIDs.FirstPost) {
		g.stagedBossStage = stagedBossFirstPost
	} else {
		g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
	}
}

func (g *Game) advanceStagedBossDialogue() {
	event, ok := g.activeStagedBoss()
	if !ok {
		g.stagedBossStage = stagedBossIdle
		return
	}
	switch g.stagedBossStage {
	case stagedBossFirstPost, stagedBossSecondAccept, stagedBossSecondVictory:
		g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
	case stagedBossSecondOffer:
		g.stagedBossStage, g.stagedBossCursor = stagedBossSecondChoice, 0
	case stagedBossSecondChallenge:
		if g.startPackFormation(event.SecondFormation, event.SecondDropPolicy == "suppress") {
			g.stagedBossStage = stagedBossSecondBattle
		} else {
			g.stagedBossStage, g.stagedBossEventID = stagedBossIdle, ""
		}
	}
}

func (g *Game) stagedBossChoiceInput(in InputState) {
	event, ok := g.activeStagedBoss()
	if !ok {
		g.stagedBossStage = stagedBossIdle
		return
	}
	switch {
	case in.Confirm && g.stagedBossCursor == 0:
		if g.openPackText(event.DialogueTextIDs.SecondAccept) {
			g.stagedBossStage = stagedBossSecondAccept
		}
	case in.Confirm || in.Cancel:
		if g.openPackText(event.DialogueTextIDs.SecondChallenge) {
			g.stagedBossStage = stagedBossSecondChallenge
		}
	case in.DirEdge == 0 || in.DirEdge == 1:
		g.stagedBossCursor ^= 1
	}
}

func (g *Game) drawStagedBossChoice(rgba []byte, white dq3data.Color) {
	if g.stagedBossStage != stagedBossSecondChoice {
		return
	}
	event, ok := g.activeStagedBoss()
	if !ok {
		return
	}
	x, y, w, h := 430, 220, 120, 68
	fillBox(rgba, x, y, w, h, white)
	ids := [2]string{event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo}
	for i, id := range ids {
		label, found := g.pack.TextGlyphCodes(id)
		if !found {
			return
		}
		yy := y + 12 + i*24
		if i == g.stagedBossCursor {
			drawGlyph(rgba, g.dlg.tx, x+12, yy, curGlyph, white)
		}
		for j, glyph := range label {
			drawGlyph(rgba, g.dlg.tx, x+40+j*16, yy, int(glyph), white)
		}
	}
}
