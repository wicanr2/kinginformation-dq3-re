package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

const (
	fieldRura     = 172 // 魯拉
	fieldRemitto  = 173 // 烈米特
	fieldInpas    = 174 // 因帕斯
	fieldToramana = 175 // 多拉瑪那
	fieldToheros  = 176 // 特黑洛斯
	fieldRanaruta = 177 // 拉那魯達
	fieldRemoaru  = 178 // レムオル（透明）
	fieldAbakamu  = 179 // 阿巴卡姆
)

type fieldSpellDef struct {
	rec int
	mp  int
}

// MP 來自 DQ3.EXE DGROUP 0x37c3 spell descriptor 第 0 byte。
var fieldSpellDefs = [...]fieldSpellDef{
	{fieldRura, spell.MPCost(fieldRura)},
	{fieldRemitto, spell.MPCost(fieldRemitto)},
	{fieldInpas, spell.MPCost(fieldInpas)},
	{fieldToramana, spell.MPCost(fieldToramana)},
	{fieldToheros, spell.MPCost(fieldToheros)},
	{fieldRanaruta, spell.MPCost(fieldRanaruta)},
	{fieldRemoaru, spell.MPCost(fieldRemoaru)},
	{fieldAbakamu, spell.MPCost(fieldAbakamu)},
}

type fieldSpellChoice struct {
	fieldSpellDef
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
	active  bool
	dest    bool
	cursor  int
	choices []fieldSpellChoice
	pending fieldSpellChoice // 魯拉進目的地頁後保留選擇；MP 已由共通 caster 扣除
	hits    hitList
}

func containsRec(xs []int, want int) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func (g *Game) fieldSpellCaster(rec int) int {
	if g.heroHP > 0 && containsRec(spell.KnownAll(0, g.heroStatsLevel()), rec) {
		return 0
	}
	for i, m := range g.companions {
		if m.Alive() && containsRec(m.AllSpells(), rec) {
			return i + 1
		}
	}
	return -1
}

func (g *Game) heroStatsLevel() int {
	level, _, _, _, _ := g.heroStats()
	return level
}

func (g *Game) openFieldSpellMenu() {
	m := &g.fieldSpell
	*m = FieldSpellMenu{active: true}
	for _, d := range fieldSpellDefs {
		if caster := g.fieldSpellCaster(d.rec); caster >= 0 {
			m.choices = append(m.choices, fieldSpellChoice{fieldSpellDef: d, caster: caster})
		}
	}
	if len(m.choices) == 0 {
		m.active = false
	}
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

func (g *Game) fieldSpellInput(in InputState) {
	m := &g.fieldSpell
	n := len(m.choices)
	if m.dest {
		n = len(g.visitedTowns)
	}
	if in.Cancel {
		if m.dest {
			// 原版 caster 在 effect handler 前已扣 MP；目的地選單取消不會回到
			// 咒文清單或退款，而是結束這次施法。
			m.active, m.dest = false, false
		} else {
			m.active = false
		}
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
	if m.dest {
		if g.teleportToVisit(g.visitedTowns[m.cursor]) {
			m.active = false
		}
		return
	}
	c := m.choices[m.cursor]
	// DQ3.EXE field caster file 0x1cbad..0x1cbd4：共通施法流程後先從
	// [member+0x18] 扣 MP，再由 0x38cc pointer table 派發 effect handler。
	// handler 的場景 gate、目的地取消或效果失敗沒有一般退款分支。
	if !g.spendFieldMP(c.caster, c.mp) {
		return
	}
	switch c.rec {
	case fieldRura:
		// 原版 field handler file 0xe0da：地表可用；CTY section 僅 mapFlags bit0 可用。
		if g.inTown && (g.cur == nil || g.cur.mapFlags&1 == 0) {
			m.active = false
			return
		}
		if len(g.visitedTowns) == 0 {
			m.active = false
			return
		}
		m.pending = c
		m.dest, m.cursor = true, 0
	case fieldRemitto:
		// 原版 file 0xe10c：scene state 必須為 CTY，且 section header +0x10 bit1。
		if !g.inTown || g.cur == nil || g.cur.mapFlags&2 == 0 {
			m.active = false
			return
		}
		g.exitTown()
		m.active = false
	case fieldInpas:
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
		m.active = false
	case fieldToramana:
		// handler file 0xe18c：設 world flag 0x2000、timer 0x28。原版一般
		// timer consumer 卻檢查 0x0200；真正 consumer 是傷害地板 file
		// 0x197a8，第一次遇到 AH&0x0c 時把 0x2000 轉成 0x0400，
		// 離開該段地板後清除。保留這個精訊版一次性連續區行為。
		g.toramana = true
		g.hazardGuard = false
		m.active = false
	case fieldToheros:
		g.repel = 0x28 // handler file 0xe19c 寫 [0x52f6]=0x28
		m.active = false
	case fieldRanaruta:
		if g.inTown { // handler file 0xe1ab：scene state==CTY 時拒絕。
			m.active = false
			return
		}
		g.toggleDaynight()
		m.active = false
	case fieldRemoaru:
		// handler file 0xe1df：world flag 0x8000、timer [0x52f7]=0x19。
		// movement consumer file 0xa902..0xa919 每個有效步遞減，歸零清 0xc000。
		g.remoaru = 0x19
		m.active = false
	case fieldAbakamu:
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
		m.active = false
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
func (g *Game) applyHazardStep() {
	if g.cur == nil || g.cur.attr == nil {
		g.hazardGuard = false
		return
	}
	ah := byte(g.cur.attr.Raw(g.cur.tileIdx(g.px, g.py)) >> 8)
	damage := 0
	switch {
	case ah&0x03 != 0:
		damage = int(ah & 0x03)
	case ah&0x0c != 0:
		if g.hazardGuard {
			return
		}
		if g.toramana {
			g.toramana = false
			g.hazardGuard = true
			return
		}
		damage = int(ah>>2) * 10
	default:
		g.hazardGuard = false
		return
	}
	if damage <= 0 {
		return
	}
	g.heroHP = max0(g.heroHP - damage)
	for _, m := range g.companions {
		if m.CurHP > 0 {
			m.CurHP = max0(m.CurHP - damage)
		}
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
	if m.dest {
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
		if m.dest {
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
