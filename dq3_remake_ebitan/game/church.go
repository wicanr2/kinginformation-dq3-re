package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

func (c *Church) open(tx *dq3data.Text) {
	c.active, c.stage = true, churchService
	c.serviceCursor, c.targetCursor, c.confirmCursor = 0, 0, 0
	c.pendingTarget, c.pendingCost, c.msg, c.tx = -1, 0, "", tx
}

func (g *Game) churchPartyLen() int { return 1 + len(g.companions) }

func (g *Game) churchMemberLevel(i int) int {
	if i == 0 {
		level, _, _, _, _ := g.heroStats()
		return level
	}
	if i > 0 && i <= len(g.companions) {
		return g.companions[i-1].Level()
	}
	return 1
}

func (g *Game) churchMemberAlive(i int) bool {
	if i == 0 {
		return g.heroHP > 0
	}
	return i > 0 && i <= len(g.companions) && g.companions[i-1].Alive()
}

func (g *Game) churchRevive(i int) {
	if i == 0 {
		_, maxHP, _, _, _ := g.heroStats()
		g.heroHP, g.heroMP, g.heroInit = maxHP, g.heroMaxMP(), true
		return
	}
	if i > 0 && i <= len(g.companions) {
		g.companions[i-1].fullHeal()
	}
}

func (g *Game) churchInput(in InputState) {
	c := &g.church
	switch c.stage {
	case churchService:
		switch {
		case in.Cancel:
			c.active = false
		case in.Confirm:
			c.msg = ""
			if c.serviceCursor != 2 {
				// Member 尚無原版 +0x38 poison/curse 持久旗標；不能假裝免費清除。
				c.msg = "此服務尚無可處理的狀態"
				return
			}
			c.stage, c.targetCursor = churchTarget, 0
		case in.DirEdge == 0:
			c.serviceCursor = (c.serviceCursor + 1) % 3
		case in.DirEdge == 1:
			c.serviceCursor = (c.serviceCursor + 2) % 3
		}
	case churchTarget:
		n := g.churchPartyLen()
		switch {
		case in.Cancel:
			c.stage, c.msg = churchService, ""
		case in.Confirm:
			if g.churchMemberAlive(c.targetCursor) {
				c.msg = "不需要復活"
				return
			}
			c.pendingTarget = c.targetCursor
			c.pendingCost = g.pack.ReviveCost(g.churchMemberLevel(c.targetCursor))
			c.confirmCursor, c.stage, c.msg = 0, churchConfirm, ""
		case in.DirEdge == 0:
			c.targetCursor = (c.targetCursor + 1) % n
		case in.DirEdge == 1:
			c.targetCursor = (c.targetCursor + n - 1) % n
		}
	case churchConfirm:
		switch {
		case in.Cancel:
			c.stage, c.msg = churchTarget, ""
		case in.Confirm:
			if c.confirmCursor == 1 {
				c.stage, c.msg = churchService, ""
				return
			}
			if g.heroGold < c.pendingCost {
				c.stage, c.msg = churchService, "金錢不足"
				return
			}
			g.heroGold -= c.pendingCost
			g.churchRevive(c.pendingTarget)
			c.stage, c.msg = churchService, "復活完成"
		case in.DirEdge == 0 || in.DirEdge == 1:
			c.confirmCursor ^= 1
		}
	}
}

func drawChurchRecord(rgba []byte, tx *dq3data.Text, rec, x, y, varNum int, col dq3data.Color) {
	if tx == nil {
		return
	}
	cx, cy := x, y
	buf := tx.Record(rec)
	for i := 0; i < len(buf); {
		v := buf[i]
		switch {
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			cx, cy, i = x, cy+dq3data.GlyphPx, i+1
		case dq3data.IsVarInsert(v):
			glyphs := []int(nil)
			if v == dq3data.TxtVarNum && varNum >= 0 {
				glyphs = digitGlyphs(varNum)
			}
			i++
			if i < len(buf) { // 原版插值控制碼後另有一個參數 word
				i++
			}
			if len(glyphs) == 0 {
				cx += dq3data.GlyphPx
				continue
			}
			for _, glyph := range glyphs {
				drawGlyph(rgba, tx, cx, cy, glyph, col)
				cx += dq3data.GlyphPx
			}
		default:
			if v < dq3data.GlyphMax {
				drawGlyph(rgba, tx, cx, cy, int(v), col)
				cx += dq3data.GlyphPx
			}
			i++
		}
	}
}

func (g *Game) drawChurch(rgba []byte, white dq3data.Color) {
	c := &g.church
	if !c.active {
		return
	}
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	fillBox(rgba, 40, 38, ScreenW-80, ScreenH-92, white)
	switch c.stage {
	case churchService:
		for i, rec := range [...]int{310, 311, 312} {
			y := 58 + i*38
			if i == c.serviceCursor {
				drawGlyph(rgba, c.tx, 58, y, curGlyph, yellow)
			}
			drawChurchRecord(rgba, c.tx, rec, 80, y, -1, white)
		}
	case churchTarget:
		drawChurchRecord(rgba, c.tx, 312, 64, 54, -1, white)
		for i := 0; i < g.churchPartyLen(); i++ {
			y := 88 + i*30
			if i == c.targetCursor {
				drawGlyph(rgba, c.tx, 58, y, curGlyph, yellow)
			}
			name := g.heroName
			if i > 0 {
				name = g.companions[i-1].Name
			}
			x := 80
			for _, gi := range name {
				drawGlyph(rgba, c.tx, x, y, gi, white)
				x += dq3data.GlyphPx
			}
			drawNumber(rgba, c.tx, 280, y, g.churchMemberLevel(i), white)
		}
	case churchConfirm:
		drawChurchRecord(rgba, c.tx, 301, 64, 58, c.pendingCost, white)
		for i, glyph := range [...]int{399, 678} { // 是／否
			y := 106 + i*34
			if i == c.confirmCursor {
				drawGlyph(rgba, c.tx, 74, y, curGlyph, yellow)
			}
			drawGlyph(rgba, c.tx, 96, y, glyph, white)
		}
	}
	if c.msg != "" {
		drawNumber(rgba, c.tx, 64, 220, g.heroGold, yellow)
	}
}
