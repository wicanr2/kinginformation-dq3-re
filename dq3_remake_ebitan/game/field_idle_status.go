package game

import (
	"crypto/sha256"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"io/fs"
	"strconv"
)

// UI-only state. It never enters snapshot; restore starts a fresh wait.
type fieldIdleState struct {
	elapsed          int
	open             bool
	restoring        bool
	consumeDirection bool
	direction        int
	background       []byte
}

func validateFieldIdleSources(assets fs.FS, p *gamepack.Pack) error {
	_, err := loadFieldIdleSources(assets, p)
	return err
}

func loadFieldIdleSources(assets fs.FS, p *gamepack.Pack) (*dq3data.Text, error) {
	s := p.Interface.FieldIdleStatus
	if s == nil || assets == nil {
		return nil, fmt.Errorf("field idle status or assets are missing")
	}
	ref, ok := p.Asset(s.FontAsset)
	if !ok {
		return nil, fmt.Errorf("field idle font reference missing")
	}
	font, err := fs.ReadFile(assets, ref.Path)
	if err != nil || int64(len(font)) != ref.Size || fmt.Sprintf("%x", sha256.Sum256(font)) != ref.SHA256 {
		return nil, fmt.Errorf("field idle font differs from source")
	}
	tx := dq3data.LoadText(font, nil)
	glyphs := []int{s.BlankGlyph}
	for digit := 0; digit < 10; digit++ {
		glyphs = append(glyphs, digit)
	}
	for _, codes := range p.Interface.PartyHUD.ClassGlyphs {
		glyphs = append(glyphs, codes...)
	}
	for _, id := range s.TextIDs() {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Record == nil {
			return nil, fmt.Errorf("field idle text reference missing")
		}
		raw, err := fs.ReadFile(assets, d.Source.File)
		if err != nil {
			return nil, fmt.Errorf("field idle text source: %w", err)
		}
		codes := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if codes == nil || len(codes) != len(d.GlyphCodes) {
			return nil, fmt.Errorf("field idle text source shape differs")
		}
		for i, code := range codes {
			if int(code) != d.GlyphCodes[i] {
				return nil, fmt.Errorf("field idle text source differs")
			}
			if code < dq3data.GlyphMax {
				glyphs = append(glyphs, int(code))
			}
		}
	}
	for _, code := range glyphs {
		if _, ok := tx.Glyph(code); !ok {
			return nil, fmt.Errorf("field idle glyph outside source")
		}
	}
	return tx, nil
}

func fieldIdleHasInput(in InputState) bool {
	return in.AnyKeyEdge || in.DirHeld >= 0 || in.DirEdge >= 0 || in.Confirm || in.Cancel || in.Enter ||
		in.Toggle || in.Settings || in.Help || in.CtxTap || in.Tapped
}

// stepFieldIdle owns only the declared ordinary-field stage, after forced
// scenes and dialogues. A held dismissal cannot also become a movement.
func (g *Game) stepFieldIdle(in InputState) bool {
	if g.pack == nil {
		g.fieldIdle.elapsed = 0
		return false
	}
	s := g.pack.Interface.FieldIdleStatus
	if g.fieldIdle.open {
		if fieldIdleHasInput(in) {
			g.fieldIdle.open = false
			g.fieldIdle.restoring = true
			g.fieldIdle.elapsed = 0
			g.fieldIdle.consumeDirection = in.DirHeld >= 0
			g.fieldIdle.direction = in.DirHeld
		}
		return true
	}
	if g.fieldIdle.consumeDirection {
		if in.DirHeld == g.fieldIdle.direction {
			g.fieldIdle.elapsed = 0
			return true
		}
		g.fieldIdle.consumeDirection = false
	}
	if s == nil || !g.inTown || g.cur == nil || !s.Enabled(g.curCty, g.cur.sec) || g.cmd.open ||
		g.panel != panelNone || g.help.open || g.settings.open || g.cd > 0 {
		g.fieldIdle.elapsed = 0
		return false
	}
	if fieldIdleHasInput(in) {
		g.fieldIdle.elapsed = 0
		return false
	}
	g.fieldIdle.elapsed++
	if g.fieldIdle.elapsed < s.HoldFrames() {
		return false
	}
	// Capture the current scene without a modal overlay. Actor animation is
	// deliberately left to the existing runtime, rather than forcing a frame.
	g.renderFrame()
	g.fieldIdle.background = append(g.fieldIdle.background[:0], g.rgba...)
	g.fieldIdle.open = true
	return true
}

func (g *Game) fieldIdleStatusWord(actor int) int {
	s := g.pack.Interface.FieldIdleStatus
	conditions := g.heroConditions
	if actor > 0 {
		conditions = g.companions[actor-1].Conditions
	}
	status := 0
	if !g.actorAlive(actor) {
		status |= s.DeadStatusMask
	}
	for _, id := range []string{gamepack.PoisonCondition, gamepack.ParalysisCondition} {
		bit, ok := conditionBit(id)
		if !ok || conditions&bit == 0 {
			continue
		}
		if definition, ok := g.pack.ConditionDefinition(id); ok {
			status |= definition.LegacyStatusMask
		}
	}
	return status
}

func (g *Game) fieldIdleForeground() dq3data.Color {
	s := g.pack.Interface.FieldIdleStatus
	colorIndex := 0
	dead := false
	actors := g.partyHUDActors()
	for i, actor := range actors {
		if g.fieldIdleStatusWord(i)&s.DeadStatusMask != 0 {
			colorIndex = s.DeadPaletteIndex
			dead = true
			break
		}
		maximum := actor.hp
		if i == 0 {
			_, maximum, _, _, _ = g.heroStats()
		} else {
			maximum = g.companions[i-1].MaxHP()
		}
		for _, divisor := range s.HealthDivisors {
			if actor.hp <= maximum/divisor {
				colorIndex++
			}
		}
	}
	if !dead && colorIndex > s.HealthSumCap {
		colorIndex = s.HealthSumCap
	}
	return dq3data.DecodePalette(s.HealthPaletteRaw[colorIndex], 1)[0]
}

func (g *Game) applyFieldIdlePalette() {
	if g.pack == nil {
		return
	}
	s := g.pack.Interface.FieldIdleStatus
	if s != nil && g.inTown && g.cur != nil && s.Enabled(g.curCty, g.cur.sec) && s.PaletteIndex < len(g.cur.pal) {
		g.cur.pal[s.PaletteIndex] = g.fieldIdleForeground()
	}
}

func (g *Game) drawFieldIdleStatus() {
	s := g.pack.Interface.FieldIdleStatus
	h := g.pack.Interface.PartyHUD
	actors := g.partyHUDActors()
	if len(actors) > h.Columns {
		actors = actors[:h.Columns]
	}
	left, _ := g.pack.TextDefinition(s.LeftTextID)
	body, _ := g.pack.TextDefinition(s.ColumnTextID)
	right, _ := g.pack.TextDefinition(s.RightTextID)
	columnWidth := body.Layout.Columns * dq3data.GlyphPx
	w := h.WindowLayout
	w.Width = (left.Layout.Columns+right.Layout.Columns)*dq3data.GlyphPx + len(actors)*columnWidth
	dark := dq3data.Color{R: s.BackdropRGB[0], G: s.BackdropRGB[1], B: s.BackdropRGB[2]}
	white := g.fieldIdleForeground()
	drawWindowWordLatchShadow(g.rgba, w, s.Shadow, dark)
	for y := w.Y; y < w.Y+w.Height; y++ {
		for x := w.X; x < w.X+w.Width; x++ {
			putPx(g.rgba, x, y, dark)
		}
	}
	glyph := func(x, y, code int) {
		bitmap, ok := g.fieldIdleFont.Glyph(code)
		if !ok {
			return
		} // bootstrap rejects missing static glyphs
		for row := 0; row < dq3data.GlyphPx; row++ {
			for col := 0; col < dq3data.GlyphPx; col++ {
				c := dark
				if bitmap[row][col] != 0 {
					c = white
				}
				putPx(g.rgba, x+col, y+row, c)
			}
		}
	}
	text := func(id string, x, y int) {
		origin := x
		codes, _ := g.pack.TextGlyphCodes(id)
		for _, code := range codes {
			if code == dq3data.TxtNL {
				x, y = origin, y+dq3data.GlyphPx
			} else {
				glyph(x, y, int(code))
				x += dq3data.GlyphPx
			}
		}
	}
	text(s.LeftTextID, w.X, w.Y)
	for i := range actors {
		text(s.ColumnTextID, w.X+left.Layout.Columns*dq3data.GlyphPx+i*columnWidth, w.Y)
	}
	text(s.RightTextID, w.X+w.Width-right.Layout.Columns*dq3data.GlyphPx, w.Y)
	number := func(x, y, value int) {
		digits := strconv.Itoa(value)
		if value < 0 || len(digits) > s.Digits {
			return
		}
		for i := 0; i < s.Digits; i++ {
			code := s.BlankGlyph
			if i >= s.Digits-len(digits) {
				code = int(digits[i-(s.Digits-len(digits))] - '0')
			}
			glyph(x+i*dq3data.GlyphPx, y, code)
		}
	}
	for i, actor := range actors {
		x := w.X + i*columnWidth
		for j, code := range actor.name {
			if j >= h.NameMaxGlyphs {
				break
			}
			glyph(x+h.TextInsetX+h.NameInsetX+j*dq3data.GlyphPx, w.Y+h.TextInsetY, code)
		}
		valueX := x + h.TextInsetX + h.ValueInsetX
		number(valueX, w.Y+s.HPInsetY, actor.hp)
		number(valueX, w.Y+s.MPInsetY, actor.mp)
		status := g.fieldIdleStatusWord(i)
		statusID := ""
		for _, row := range s.StatusRows {
			if status&row.MaskRaw != 0 {
				statusID = row.TextID
				break
			}
		}
		if statusID != "" {
			text(statusID, x+s.StatusInsetX, w.Y+s.StatusInsetY)
		} else {
			number(valueX, w.Y+s.StatusInsetY, actor.level)
		}
		if actor.class >= 0 && actor.class < len(h.ClassGlyphs) {
			for j, code := range h.ClassGlyphs[actor.class] {
				glyph(x+h.TextInsetX+j*dq3data.GlyphPx, w.Y+s.StatusInsetY, code)
			}
		}
	}
}
