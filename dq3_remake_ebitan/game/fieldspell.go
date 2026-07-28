package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

const (
	fieldRura     = 172 // 魯拉
	fieldRemitto  = 173 // 烈米特
	fieldToheros  = 176 // 特黑洛斯
	fieldRanaruta = 177 // 拉那魯達
)

type fieldSpellDef struct {
	rec int
	mp  int
}

// MP 來自 DQ3.EXE DGROUP 0x37c3 spell descriptor 第 0 byte。
var fieldSpellDefs = [...]fieldSpellDef{
	{fieldRura, spell.MPCost(fieldRura)},
	{fieldRemitto, spell.MPCost(fieldRemitto)},
	{fieldToheros, spell.MPCost(fieldToheros)},
	{fieldRanaruta, spell.MPCost(fieldRanaruta)},
}

type fieldSpellChoice struct {
	fieldSpellDef
	caster int // 0=勇者，1..=companions index+1
}

type townVisit struct {
	Cty, Section, X, Y int
}

// DQ3.EXE DGROUP 0x0ac0：魯拉目的地 index → CTY。原版 handler file 0x5f9b
// 以此值索引 cty_loc，寫入 world X=ctyX-1、Y=ctyY；不是載入城內最後位置。
var ruraCtyOrder = [...]int{
	0, 1, 2, 3, 4, 6, 12, 16, 15, 17, 47, 21, 39, 43, 38, 79, 81, 84, 85, 86,
}

func ruraCtyAllowed(cty int) bool {
	return ruraCtyIndex(cty) >= 0
}

func ruraCtyIndex(cty int) int {
	for i, v := range ruraCtyOrder {
		if v == cty {
			return i
		}
	}
	return -1
}

func ruraNameRec(cty int) int {
	if i := ruraCtyIndex(cty); i >= 0 {
		return 0x1ef + i // handler file 0x5f29..0x5f32
	}
	return -1
}

type FieldSpellMenu struct {
	active  bool
	dest    bool
	cursor  int
	choices []fieldSpellChoice
	pending fieldSpellChoice // 魯拉進目的地頁後保留施法者／MP；cursor 此時已改作目的地 index
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
			m.dest, m.cursor = false, 0
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
	if g.fieldCasterMP(c.caster) < c.mp {
		return
	}
	switch c.rec {
	case fieldRura:
		// 原版 field handler file 0xe0da：地表可用；CTY section 僅 mapFlags bit0 可用。
		if g.inTown && (g.cur == nil || g.cur.mapFlags&1 == 0) {
			return
		}
		if len(g.visitedTowns) == 0 {
			return
		}
		m.pending = c
		m.dest, m.cursor = true, 0 // 選定目的地後才扣 MP
	case fieldRemitto:
		// 原版 file 0xe10c：scene state 必須為 CTY，且 section header +0x10 bit1。
		if !g.inTown || g.cur == nil || g.cur.mapFlags&2 == 0 {
			return
		}
		if g.spendFieldMP(c.caster, c.mp) {
			g.exitTown()
			m.active = false
		}
	case fieldToheros:
		if g.spendFieldMP(c.caster, c.mp) {
			g.repel = 0x28 // handler file 0xe19c 寫 [0x52f6]=0x28
			m.active = false
		}
	case fieldRanaruta:
		if g.inTown { // handler file 0xe1ab：scene state==CTY 時拒絕。
			return
		}
		if g.spendFieldMP(c.caster, c.mp) {
			g.toggleDaynight()
			m.active = false
		}
	}
}

func (g *Game) teleportToVisit(v townVisit) bool {
	if v.Cty < 0 || v.Cty >= len(ctyLoc) || !ruraCtyAllowed(v.Cty) {
		return false
	}
	loc := ctyLoc[v.Cty]
	if loc[2] != 0 && loc[2] != 1 {
		return false
	}
	choice := g.fieldSpell.pending
	if !g.spendFieldMP(choice.caster, choice.mp) {
		return false
	}
	g.layer = loc[2]
	g.cur = g.overworldScene()
	g.inTown, g.curCty = false, -1
	g.px, g.py = loc[0]-1, loc[1]
	if v.Cty == 47 { // handler file 0x5fd9：目的 index 10（CTY47）另加 X+2
		g.px += 2
	}
	g.shipAboard, g.phoenixAboard = false, false
	if g.layer == 1 {
		g.music.Play(trackDungeon)
	} else {
		g.music.Play(trackField)
	}
	return true
}

func (g *Game) rememberTown() {
	if !g.inTown || g.cur == nil || !ruraCtyAllowed(g.curCty) {
		return
	}
	g.addVisitedTown(g.curCty)
}

func (g *Game) addVisitedTown(cty int) {
	if !ruraCtyAllowed(cty) {
		return
	}
	for i := range g.visitedTowns {
		if g.visitedTowns[i].Cty == cty {
			return
		}
	}
	order := ruraCtyIndex(cty)
	at := len(g.visitedTowns)
	for i, v := range g.visitedTowns {
		if ruraCtyIndex(v.Cty) > order {
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
			drawTextRecord(rgba, g.shop.nameText, x, y, ruraNameRec(g.visitedTowns[i].Cty), white)
		} else {
			drawTextRecord(rgba, g.shop.nameText, x, y, m.choices[i].rec, white)
			drawNumber(rgba, g.dlg.tx, ScreenW-116, y, m.choices[i].mp, white)
		}
		m.hits.add(x-20, y-3, 244, 20, i)
	}
}
