package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

type fieldSpellChoice struct {
	def    gamepack.FieldSpellDefinition
	rec    int
	mp     int
	caster int // 0=勇者，1..=companions index+1
}

type townVisit struct {
	Cty, Section, X, Y int
}

func (g *Game) ruraCtyIndex(cty int) int {
	if g.pack == nil {
		return -1
	}
	for i, d := range g.pack.RuraDestinations() {
		if d.CTYRaw == cty {
			return i
		}
	}
	return -1
}

func (g *Game) ruraDestination(cty int) (gamepack.RuraDestination, bool) {
	if g.pack == nil {
		return gamepack.RuraDestination{}, false
	}
	if i := g.ruraCtyIndex(cty); i >= 0 {
		return g.pack.RuraDestinations()[i], true
	}
	return gamepack.RuraDestination{}, false
}

type FieldSpellMenu struct {
	active          bool
	selectingCaster bool
	dest            bool
	targeting       bool
	cursor          int
	caster          int
	choices         []fieldSpellChoice
	pending         fieldSpellChoice // 目的地頁已扣 MP；目標頁尚未扣 MP
	hits            hitList
}

func containsRec(xs []int, want int) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func (g *Game) heroStatsLevel() int {
	level, _, _, _, _ := g.heroStats()
	return level
}

func (g *Game) openFieldSpellMenu() {
	m := &g.fieldSpell
	*m = FieldSpellMenu{active: true, selectingCaster: true}
}

func (g *Game) fieldActorSpells(actor int) []int {
	if actor == 0 {
		return spell.KnownAll(0, g.heroStatsLevel())
	}
	if actor < 1 || actor > len(g.companions) || g.companions[actor-1] == nil {
		return nil
	}
	return g.companions[actor-1].AllSpells()
}

// fieldSpellCaster is a read-only trace diagnostic. Production UI does not
// call it to choose for the player; it only reports the first living actor who
// knows a pack-declared field spell so long-running tests can plan inputs.
func (g *Game) fieldSpellCaster(rec int) int {
	if g.pack == nil {
		return -1
	}
	if _, ok := g.pack.FieldSpellByRecord(rec); !ok {
		return -1
	}
	for actor := 0; actor <= len(g.companions); actor++ {
		if g.fieldActorAlive(actor) && containsRec(g.fieldActorSpells(actor), rec) {
			return actor
		}
	}
	return -1
}

func (g *Game) selectFieldSpellCaster(actor int) bool {
	if g.pack == nil || actor < 0 || actor > len(g.companions) {
		return false
	}
	m := &g.fieldSpell
	m.choices = m.choices[:0]
	for _, rec := range g.fieldActorSpells(actor) {
		def, ok := g.pack.FieldSpellByRecord(rec)
		if !ok {
			continue
		}
		m.choices = append(m.choices, fieldSpellChoice{
			def: def, rec: def.RecordRaw, mp: def.MPCost, caster: actor,
		})
	}
	if len(m.choices) == 0 {
		return false
	}
	m.selectingCaster, m.caster, m.cursor = false, actor, 0
	return true
}

func (g *Game) fieldCasterMP(caster int) int {
	if caster == 0 {
		return g.heroMP
	}
	if caster-1 >= 0 && caster-1 < len(g.companions) {
		return g.companions[caster-1].CurMP
	}
	return 0
}

func (g *Game) spendFieldMP(caster, mp int) bool {
	if g.fieldCasterMP(caster) < mp {
		return false
	}
	if caster == 0 {
		g.heroMP -= mp
	} else {
		g.companions[caster-1].CurMP -= mp
	}
	return true
}

func (g *Game) fieldActorAlive(actor int) bool {
	if actor == 0 {
		return g.heroHP > 0
	}
	return actor > 0 && actor <= len(g.companions) &&
		g.companions[actor-1] != nil && g.companions[actor-1].Alive()
}

func (g *Game) healFieldActor(actor int, def gamepack.FieldSpellDefinition) {
	if !g.fieldActorAlive(actor) {
		return
	}
	amount := def.BaseAmount
	if def.AmountFormula == "base_to_base_plus_9" {
		amount += g.prng.Next(10)
	}
	if actor == 0 {
		_, maxHP, _, _, _ := g.heroStats()
		if def.AmountFormula == "full_hp" || g.heroHP+amount > maxHP {
			g.heroHP = maxHP
		} else {
			g.heroHP += amount
		}
		return
	}
	m := g.companions[actor-1]
	if def.AmountFormula == "full_hp" || m.CurHP+amount > m.MaxHP() {
		m.CurHP = m.MaxHP()
	} else {
		m.CurHP += amount
	}
}

func (g *Game) applyFieldHeal(choice fieldSpellChoice, target int) {
	if choice.def.TargetScope == "party_all" {
		for actor := 0; actor <= len(g.companions); actor++ {
			g.healFieldActor(actor, choice.def)
		}
		return
	}
	g.healFieldActor(target, choice.def)
}

func (g *Game) fieldSpellInput(in InputState) {
	m := &g.fieldSpell
	n := len(m.choices)
	if m.selectingCaster || m.targeting {
		n = 1 + len(g.companions)
	} else if m.dest {
		n = len(g.visitedTowns)
	}
	if in.Cancel {
		if m.dest {
			// 原版 caster 在 effect handler 前已扣 MP；目的地選單取消不會回到
			// 咒文清單或退款，而是結束這次施法。
		}
		// 原版施法者、咒文與單體目標 selector 的取消都返回上層 command；
		// 目標 selector 發生在 MP writer 前，所以該路徑不扣 MP。
		*m = FieldSpellMenu{}
		return
	}
	if n == 0 {
		return
	}
	confirm := in.Confirm
	if in.Tapped {
		if i := m.hits.at(in.TapX, in.TapY); i >= 0 && i < n {
			m.cursor, confirm = i, true
		}
	}
	if in.DirEdge == 0 {
		m.cursor = (m.cursor + 1) % n
	} else if in.DirEdge == 1 {
		m.cursor = (m.cursor + n - 1) % n
	} else if m.dest && in.DirEdge == 2 && m.cursor >= 10 {
		m.cursor -= 10
	} else if m.dest && in.DirEdge == 3 && m.cursor+10 < n {
		m.cursor += 10
	}
	if !confirm {
		return
	}
	if m.selectingCaster {
		if !g.selectFieldSpellCaster(m.cursor) {
			*m = FieldSpellMenu{}
		}
		return
	}
	if m.dest {
		if g.teleportToVisit(g.visitedTowns[m.cursor]) {
			*m = FieldSpellMenu{}
		}
		return
	}
	if m.targeting {
		choice := m.pending
		if !g.spendFieldMP(choice.caster, choice.mp) {
			*m = FieldSpellMenu{}
			return
		}
		g.applyFieldHeal(choice, m.cursor)
		*m = FieldSpellMenu{}
		return
	}
	c := m.choices[m.cursor]
	if !g.fieldActorAlive(c.caster) || g.fieldCasterMP(c.caster) < c.mp {
		*m = FieldSpellMenu{}
		return
	}
	if c.def.EffectID == gamepack.FieldHealHP && c.def.TargetScope == "party_member" {
		m.pending, m.targeting, m.cursor = c, true, 0
		return
	}
	// DQ3.EXE field caster file 0x1cbad..0x1cbd4：共通施法流程後先從
	// [member+0x18] 扣 MP，再由 0x38cc pointer table 派發 effect handler。
	// handler 的場景 gate、目的地取消或效果失敗沒有一般退款分支。
	if !g.spendFieldMP(c.caster, c.mp) {
		*m = FieldSpellMenu{}
		return
	}
	switch c.def.EffectID {
	case gamepack.FieldHealHP:
		g.applyFieldHeal(c, 0)
		*m = FieldSpellMenu{}
	case gamepack.FieldTeleport:
		// 原版 field handler file 0xe0da：地表可用；CTY section 僅 mapFlags bit0 可用。
		if g.inTown && (g.cur == nil || g.cur.mapFlags&1 == 0) {
			*m = FieldSpellMenu{}
			return
		}
		if len(g.visitedTowns) == 0 {
			*m = FieldSpellMenu{}
			return
		}
		m.pending = c
		m.dest, m.cursor = true, 0
	case gamepack.FieldExitDungeon:
		// 原版 file 0xe10c：scene state 必須為 CTY，且 section header +0x10 bit1。
		if !g.inTown || g.cur == nil || g.cur.mapFlags&2 == 0 {
			*m = FieldSpellMenu{}
			return
		}
		g.exitTown()
		*m = FieldSpellMenu{}
	case gamepack.FieldInspectFacing:
		// handler file 0xe13b 將 examine mode 設 2，使 file 0x9cf6 只回傳
		// 面向事件表的 type byte而不執行。type2→rec254，其餘合法 type→rec253；
		// 無事件／非 CTY 則走共通失敗 rec346 (0x15a)。
		typ, ok := g.inpasFacingType()
		switch {
		case !ok:
			g.dlg.OpenFrom(g.shop.nameText, 0x15a)
		case typ == 2:
			g.dlg.OpenFrom(g.shop.nameText, 0xfe)
		default:
			g.dlg.OpenFrom(g.shop.nameText, 0xfd)
		}
		*m = FieldSpellMenu{}
	case gamepack.FieldHazardGuard:
		// handler file 0xe18c：設 world flag 0x2000、timer 0x28。原版一般
		// timer consumer 卻檢查 0x0200；真正 consumer 是傷害地板 file
		// 0x197a8，第一次遇到 AH&0x0c 時把 0x2000 轉成 0x0400，
		// 離開該段地板後清除。保留這個精訊版一次性連續區行為。
		g.toramana = true
		g.hazardGuard = false
		*m = FieldSpellMenu{}
	case gamepack.FieldRepel:
		g.repel = 0x28 // handler file 0xe19c 寫 [0x52f6]=0x28
		*m = FieldSpellMenu{}
	case gamepack.FieldToggleDayNight:
		if g.inTown { // handler file 0xe1ab：scene state==CTY 時拒絕。
			*m = FieldSpellMenu{}
			return
		}
		g.toggleDaynight()
		*m = FieldSpellMenu{}
	case gamepack.FieldInvisibility:
		// handler file 0xe1df：world flag 0x8000、timer [0x52f7]=0x19。
		// movement consumer file 0xa902..0xa919 每個有效步遞減，歸零清 0xc000。
		g.remoaru = 0x19
		*m = FieldSpellMenu{}
	case gamepack.FieldOpenFacingDoor:
		// handler file 0xe1f2 設 door permission 4 後呼叫 file 0x14906。
		// 該 consumer 只在 CTY scene 檢查面向格；door tier 值域 1..3，
		// 因此 permission 4 可開全部門，失敗則顯示原版 rec0x15a。
		opened := false
		if g.inTown && g.cur != nil {
			dx, dy := dirDelta(g.facing)
			opened = g.cur.openDoor(g.px+dx, g.py+dy)
		}
		if !opened {
			g.dlg.OpenFrom(g.shop.nameText, 0x15a)
		}
		*m = FieldSpellMenu{}
	default:
		// Pack validator should reject unknown primitives. Keep runtime
		// fail-closed in case a non-canonical Pack is injected by a test.
		*m = FieldSpellMenu{}
	}
}

func (g *Game) inpasFacingType() (int, bool) {
	if !g.inTown || g.cur == nil {
		return 0, false
	}
	x, y := frontTile(g.px, g.py, g.facing)
	ev, subid, ok := g.cur.tileEvent(x, y)
	if !ok {
		return 0, false
	}
	// 原版正常 examine 會清 tile 的低 5-bit event id。Go 版目前以一次性
	// treasure flag 保存已取狀態，因此在 Inpas read-only path 補同等 gate。
	if event, ok := g.questTreasureFor(g.curCty, g.cur.sec, subid); ok &&
		!g.storyFlag(event.Treasure.PresentFlag) {
		return 0, false
	}
	if t := treasureFor(g.curCty, g.cur.sec, subid); t != nil && g.flags[t[5]] {
		return 0, false
	}
	return ev[0], true
}

// advanceFieldSpellStep 對齊原版成功移動後的暫態效果 consumer。撞牆或 modal frame
// 不算一步；城鎮與地表都會進入 file 0xa8a0 的 common movement-status routine。
func (g *Game) advanceFieldSpellStep() {
	if g.remoaru > 0 {
		g.remoaru--
	}
}

// applyHazardStep 對齊原版 movement-status file 0x195ec..0x1980f。
// tile attr 高 byte低 nibble：
//
//	bits0..1 → 全隊 1..3 傷害，多拉瑪那不保護；
//	bits2..3 → 全隊 10/20/30 傷害，0x2000 多拉瑪那轉成 0x0400 連續區 guard；
//	無 hazard → 清 0x0400 guard。
//
// 同一原版 per-step damage byte 還會由角色 poison condition 加上 pack 指定傷害；
// 船／飛行的 suppress mask 只跳過毒傷，不會吞掉地形傷害。
func (g *Game) applyHazardStep() {
	heroWasAlive := g.heroHP > 0
	companionWasAlive := make([]bool, len(g.companions))
	for i, member := range g.companions {
		companionWasAlive[i] = member != nil && member.CurHP > 0
	}
	terrainDamage := 0
	if g.cur == nil || g.cur.attr == nil {
		g.hazardGuard = false
	} else {
		ah := byte(g.cur.attr.Raw(g.cur.tileIdx(g.px, g.py)) >> 8)
		switch {
		case ah&0x03 != 0:
			terrainDamage = int(ah & 0x03)
		case ah&0x0c != 0:
			if !g.hazardGuard {
				if g.toramana {
					g.toramana = false
					g.hazardGuard = true
				} else {
					terrainDamage = int(ah>>2) * 10
				}
			}
		default:
			g.hazardGuard = false
		}
	}
	poisonDamage, poisonSuppressed := 0, false
	if g.pack != nil {
		if condition, ok := g.pack.PoisonConditionDefinition(); ok {
			poisonDamage = condition.FieldDamagePerStep
			vehicleState := 0
			if g.shipAboard {
				vehicleState |= 1
			}
			if g.phoenixAboard {
				vehicleState |= 2
			}
			poisonSuppressed = vehicleState&condition.SuppressWorldStateMask != 0
		}
	}
	heroDamage := terrainDamage
	if !poisonSuppressed && g.heroConditions&conditionPoison != 0 {
		heroDamage += poisonDamage
	}
	if g.heroHP > 0 {
		g.heroHP = max0(g.heroHP - heroDamage)
	}
	for _, m := range g.companions {
		if m.CurHP > 0 {
			damage := terrainDamage
			if !poisonSuppressed && m.Conditions&conditionPoison != 0 {
				damage += poisonDamage
			}
			m.CurHP = max0(m.CurHP - damage)
		}
	}
	newDeath := heroWasAlive && g.heroHP == 0
	for i, member := range g.companions {
		if companionWasAlive[i] && member.CurHP == 0 {
			newDeath = true
		}
	}
	if !newDeath {
		return
	}
	if g.partyAllDown() {
		g.beginFieldDefeat()
	} else {
		g.selectLivingPartyLeader()
	}
}

func (g *Game) partyHasCondition(condition conditionSet) bool {
	if g.heroConditions&condition != 0 {
		return true
	}
	for _, member := range g.companions {
		if member.Conditions&condition != 0 {
			return true
		}
	}
	return false
}

// advancePersistentConditionStep executes pack-owned field expiry contracts.
// DQ3 paralysis uses one shared 0x28-step counter and clears every affected
// party member together; poison has no expiry and remains in applyHazardStep.
func (g *Game) advancePersistentConditionStep() {
	if !g.partyHasCondition(conditionParalysis) {
		g.paralysisSteps = 0
		return
	}
	if g.paralysisSteps <= 0 {
		if g.pack == nil {
			return
		}
		definition, ok := g.pack.ConditionDefinition(gamepack.ParalysisCondition)
		if !ok || definition.FieldClearAfterSteps <= 0 {
			return
		}
		g.paralysisSteps = definition.FieldClearAfterSteps
	}
	g.paralysisSteps--
	if g.paralysisSteps > 0 {
		return
	}
	g.heroConditions &^= conditionParalysis
	for _, member := range g.companions {
		member.Conditions &^= conditionParalysis
	}
}

func (g *Game) teleportToVisit(v townVisit) bool {
	destination, ok := g.ruraDestination(v.Cty)
	if v.Cty < 0 || v.Cty >= len(ctyLoc) || !ok {
		return false
	}
	loc := ctyLoc[v.Cty]
	if loc[2] != 0 && loc[2] != 1 {
		return false
	}
	g.layer = loc[2]
	g.cur = g.overworldScene()
	g.inTown, g.curCty = false, -1
	g.px = loc[0] + destination.ArrivalOffset.X
	g.py = loc[1] + destination.ArrivalOffset.Y
	if g.shipOwned {
		// DQ3.EXE logical 0x4c7e..0x4c9c：world-state bit0 設定時，
		// 以同一魯拉 index 查 DGROUP 0x0ad4 並寫回船座標。
		g.shipX = destination.ShipWorldPosition.X
		g.shipY = destination.ShipWorldPosition.Y
	}
	g.shipAboard, g.phoenixAboard = false, false
	if g.layer == 1 {
		g.playAudioCue(audioCueDungeon)
	} else {
		g.playAudioCue(audioCueField)
	}
	return true
}

func (g *Game) rememberTown() {
	if !g.inTown || g.cur == nil || g.ruraCtyIndex(g.curCty) < 0 {
		return
	}
	g.addVisitedTown(g.curCty)
}

func (g *Game) addVisitedTown(cty int) {
	if g.ruraCtyIndex(cty) < 0 {
		return
	}
	for i := range g.visitedTowns {
		if g.visitedTowns[i].Cty == cty {
			return
		}
	}
	order := g.ruraCtyIndex(cty)
	at := len(g.visitedTowns)
	for i, v := range g.visitedTowns {
		if g.ruraCtyIndex(v.Cty) > order {
			at = i
			break
		}
	}
	g.visitedTowns = append(g.visitedTowns, townVisit{})
	copy(g.visitedTowns[at+1:], g.visitedTowns[at:])
	g.visitedTowns[at] = townVisit{Cty: cty}
}

func drawTextRecord(rgba []byte, tx *dq3data.Text, x, y, rec int, fg dq3data.Color) {
	if tx == nil || rec < 0 {
		return
	}
	col := 0
	for _, v := range tx.Record(rec) {
		if v < dq3data.GlyphMax {
			drawGlyph(rgba, tx, x+col*dq3data.GlyphPx, y, int(v), fg)
			col++
		}
	}
}

func (g *Game) drawFieldSpell(rgba []byte, white dq3data.Color) {
	m := &g.fieldSpell
	if !m.active {
		return
	}
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	m.hits.reset()
	n := len(m.choices)
	if m.selectingCaster || m.targeting {
		n = 1 + len(g.companions)
	} else if m.dest {
		n = len(g.visitedTowns)
	}
	for i := 0; i < n && i < 20; i++ {
		col, row := 0, i
		if m.dest && i >= 10 {
			col, row = 1, i-10
		}
		x := 64 + col*256
		y := 56 + row*22
		if i == m.cursor {
			drawGlyph(rgba, g.dlg.tx, x-20, y, 11, yellow)
		}
		if m.selectingCaster || m.targeting {
			for j, glyph := range g.equipActorName(i) {
				drawGlyph(rgba, g.dlg.tx, x+j*dq3data.GlyphPx, y, glyph, white)
			}
		} else if m.dest {
			// 原版直接使用 D3TXT00 rec 0x1ef+destination index 的正式地名。
			if destination, ok := g.ruraDestination(g.visitedTowns[i].Cty); ok {
				drawTextRecord(rgba, g.shop.nameText, x, y, destination.NameRecordRaw, white)
			}
		} else {
			drawTextRecord(rgba, g.shop.nameText, x, y, m.choices[i].rec, white)
			drawNumber(rgba, g.dlg.tx, ScreenW-116, y, m.choices[i].mp, white)
		}
		m.hits.add(x-20, y-3, 244, 20, i)
	}
}
