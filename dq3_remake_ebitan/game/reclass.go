package game

import (
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

const (
	reclassIdle = iota
	reclassIntroduction
	reclassInitialChoice
	reclassDeclined
	reclassChooseMemberPrompt
	reclassChooseMember
	reclassRejected
	reclassChooseClassPrompt
	reclassChooseClass
	reclassSameClass
	reclassConfirmClass
	reclassConfirmClassChoice
	reclassConfirmLevel
	reclassConfirmLevelChoice
	reclassSuccess
	reclassFarewell
)

func (g *Game) activeReclass() (*gamepack.ReclassEvent, bool) {
	if g.pack == nil || g.reclassEventID == "" {
		return nil, false
	}
	return g.pack.ReclassEvent(g.reclassEventID)
}

func (g *Game) talkReclass(n *npcInst) bool {
	if g.pack == nil || g.cur == nil || g.reclassStage != reclassIdle {
		return false
	}
	for _, event := range g.pack.ReclassEvents() {
		if event.NPC.CTYRaw != g.curCty || event.NPC.Section != g.cur.sec ||
			event.NPC.Tile.X != n.x || event.NPC.Tile.Y != n.y ||
			event.NPC.HandlerRaw != n.b4 {
			continue
		}
		if !g.openPackText(event.DialogueTextIDs.Introduction) {
			return false
		}
		g.reclassEventID = event.ID
		g.reclassStage = reclassIntroduction
		g.reclassCursor, g.reclassMember, g.reclassTarget = 0, -1, -1
		return true
	}
	return false
}

func (g *Game) reclassChoosing() bool {
	switch g.reclassStage {
	case reclassInitialChoice, reclassChooseMember, reclassChooseClass,
		reclassConfirmClassChoice, reclassConfirmLevelChoice:
		return true
	}
	return false
}

func (g *Game) reclassMemberAt(index int) (*Member, int, []int, bool) {
	if index == 0 {
		name := g.heroName
		if len(name) == 0 {
			name = classNames[0]
		}
		return nil, 0, name, true
	}
	if index < 1 || index > len(g.companions) {
		return nil, -1, nil, false
	}
	m := g.companions[index-1]
	return m, m.Class, m.Name, true
}

func containsInt(values []int, want int) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func (g *Game) reclassClassGlyphs(event *gamepack.ReclassEvent, class int) []int {
	if event == nil || g.pack == nil {
		return nil
	}
	id := event.ClassNameTextIDsRaw[fmt.Sprintf("%d", class)]
	glyphs, ok := g.pack.TextGlyphCodes(id)
	if !ok {
		return nil
	}
	out := make([]int, 0, len(glyphs))
	for _, glyph := range glyphs {
		if glyph < 0xffed {
			out = append(out, int(glyph))
		}
	}
	return out
}

func (g *Game) openReclassText(id string, memberName, className []int) bool {
	if !g.openPackText(id) {
		return false
	}
	if memberName != nil {
		g.setDlgVarGlyphs(dq3data.TxtVarEnt, memberName)
	}
	if className != nil {
		g.setDlgVarGlyphs(dq3data.TxtVarNum, className)
	}
	return true
}

func (g *Game) reclassTargets(event *gamepack.ReclassEvent, m *Member) []int {
	if event == nil || m == nil {
		return nil
	}
	targets := append([]int(nil), event.BasicTargetClasses...)
	if m.Class == event.AdvancedFreeSourceClass ||
		containsInt(m.Inventory, event.AdvancedRequiredItemRaw) {
		targets = append(targets, event.AdvancedTargetClass)
	}
	return targets
}

func (g *Game) reclassChoiceInput(in InputState) {
	event, ok := g.activeReclass()
	if !ok {
		g.resetReclass()
		return
	}
	switch g.reclassStage {
	case reclassInitialChoice:
		switch {
		case in.Confirm && g.reclassCursor == 0:
			if g.openPackText(event.DialogueTextIDs.ChooseMember) {
				g.reclassStage = reclassChooseMemberPrompt
			}
		case in.Confirm || in.Cancel:
			if g.openPackText(event.DialogueTextIDs.Decline) {
				g.reclassStage = reclassDeclined
			}
		case in.DirEdge == 0 || in.DirEdge == 1:
			g.reclassCursor ^= 1
		}
	case reclassChooseMember:
		count := 1 + len(g.companions)
		switch {
		case in.Cancel:
			g.openReclassFarewell(event)
		case in.Confirm && count > 0:
			m, class, name, valid := g.reclassMemberAt(g.reclassCursor)
			if !valid {
				return
			}
			g.reclassMember = g.reclassCursor
			switch {
			case containsInt(event.ForbiddenSourceClasses, class):
				if g.openPackText(event.DialogueTextIDs.HeroForbidden) {
					g.reclassStage = reclassRejected
				}
			case m == nil || m.Level() < event.MinimumLevel:
				classGlyphs := g.reclassClassGlyphs(event, class)
				if len(classGlyphs) == 0 {
					g.resetReclass()
					return
				}
				if g.openReclassText(event.DialogueTextIDs.LevelInsufficient, name, classGlyphs) {
					g.reclassStage = reclassRejected
				}
			default:
				if g.openReclassText(event.DialogueTextIDs.ChooseClass, name, nil) {
					g.reclassStage = reclassChooseClassPrompt
				}
			}
		case in.DirEdge == 0 && count > 0:
			g.reclassCursor = (g.reclassCursor + 1) % count
		case in.DirEdge == 1 && count > 0:
			g.reclassCursor = (g.reclassCursor + count - 1) % count
		}
	case reclassChooseClass:
		m, _, name, valid := g.reclassMemberAt(g.reclassMember)
		targets := g.reclassTargets(event, m)
		switch {
		case !valid || len(targets) == 0:
			g.openReclassFarewell(event)
		case in.Cancel:
			g.openReclassFarewell(event)
		case in.Confirm:
			target := targets[g.reclassCursor]
			if target == m.Class {
				if g.openPackText(event.DialogueTextIDs.SameClass) {
					g.reclassStage = reclassSameClass
				}
				return
			}
			g.reclassTarget = target
			if g.openReclassText(event.DialogueTextIDs.ConfirmClass, name,
				g.reclassClassGlyphs(event, target)) {
				g.reclassStage = reclassConfirmClass
			}
		case in.DirEdge == 0:
			g.reclassCursor = (g.reclassCursor + 1) % len(targets)
		case in.DirEdge == 1:
			g.reclassCursor = (g.reclassCursor + len(targets) - 1) % len(targets)
		}
	case reclassConfirmClassChoice:
		switch {
		case in.Confirm && g.reclassCursor == 0:
			if g.openPackText(event.DialogueTextIDs.ConfirmLevelReset) {
				g.reclassStage = reclassConfirmLevel
			}
		case in.Confirm || in.Cancel:
			g.openReclassFarewell(event)
		case in.DirEdge == 0 || in.DirEdge == 1:
			g.reclassCursor ^= 1
		}
	case reclassConfirmLevelChoice:
		switch {
		case in.Confirm && g.reclassCursor == 0:
			_, _, name, valid := g.reclassMemberAt(g.reclassMember)
			if !valid || !g.openReclassText(event.DialogueTextIDs.Success, name,
				g.reclassClassGlyphs(event, g.reclassTarget)) {
				g.resetReclass()
				return
			}
			g.reclassStage = reclassSuccess
		case in.Confirm || in.Cancel:
			g.openReclassFarewell(event)
		case in.DirEdge == 0 || in.DirEdge == 1:
			g.reclassCursor ^= 1
		}
	}
}

func (g *Game) advanceReclassDialogue() {
	event, ok := g.activeReclass()
	if !ok {
		g.resetReclass()
		return
	}
	switch g.reclassStage {
	case reclassIntroduction:
		g.reclassStage, g.reclassCursor = reclassInitialChoice, 0
	case reclassDeclined, reclassRejected:
		g.openReclassFarewell(event)
	case reclassChooseMemberPrompt:
		g.reclassStage, g.reclassCursor = reclassChooseMember, 0
	case reclassChooseClassPrompt:
		g.reclassStage, g.reclassCursor = reclassChooseClass, 0
	case reclassSameClass:
		g.reclassStage, g.reclassCursor = reclassChooseClass, 0
	case reclassConfirmClass:
		g.reclassStage, g.reclassCursor = reclassConfirmClassChoice, 0
	case reclassConfirmLevel:
		g.reclassStage, g.reclassCursor = reclassConfirmLevelChoice, 0
	case reclassSuccess:
		if !g.applyReclass(event) {
			g.resetReclass()
			return
		}
		g.openReclassFarewell(event)
	case reclassFarewell:
		g.resetReclass()
	}
}

func (g *Game) openReclassFarewell(event *gamepack.ReclassEvent) {
	if event != nil && g.openPackText(event.DialogueTextIDs.Farewell) {
		g.reclassStage = reclassFarewell
		return
	}
	g.resetReclass()
}

func (g *Game) resetReclass() {
	g.reclassEventID = ""
	g.reclassStage = reclassIdle
	g.reclassCursor, g.reclassMember, g.reclassTarget = 0, -1, -1
}

func removeOne(values []int, want int) ([]int, bool) {
	for i, value := range values {
		if value == want {
			return append(values[:i], values[i+1:]...), true
		}
	}
	return values, false
}

// applyReclass mirrors DQ3.EXE logical 0x0be9..0x0c73. It intentionally does
// not halve VIT, clear prior spells, or consume a party-wide item.
func (g *Game) applyReclass(event *gamepack.ReclassEvent) bool {
	if event == nil || g.reclassMember <= 0 || g.reclassMember > len(g.companions) {
		return false
	}
	m := g.companions[g.reclassMember-1]
	targets := g.reclassTargets(event, m)
	if !containsInt(targets, g.reclassTarget) || m.Class == g.reclassTarget ||
		m.Level() < event.MinimumLevel || m.itemCount() > event.PersonalInventorySlots {
		return false
	}
	m.syncLearnedSpells()
	if g.reclassTarget == event.AdvancedTargetClass &&
		m.Class != event.AdvancedFreeSourceClass {
		var consumed bool
		m.Inventory, consumed = removeOne(m.Inventory, event.AdvancedRequiredItemRaw)
		if !consumed {
			return false
		}
	}
	m.Inventory = append(m.Inventory, m.equippedItems()...)
	m.Weapon, m.Armor, m.Shield, m.Head = -1, -1, -1, -1
	m.Class, m.Exp = g.reclassTarget, 0
	m.CurHP >>= 1
	m.CurMP >>= 1
	for _, kind := range []stats.StatKind{
		stats.STR, stats.AGI, stats.HP, stats.MP, stats.INT, stats.LUCK,
	} {
		m.Stats[kind] >>= 1
	}
	m.syncLearnedSpells()
	return true
}

func (g *Game) drawReclass(rgba []byte, white dq3data.Color) {
	event, ok := g.activeReclass()
	if !ok || !g.reclassChoosing() {
		return
	}
	switch g.reclassStage {
	case reclassInitialChoice, reclassConfirmClassChoice, reclassConfirmLevelChoice:
		g.drawReclassYesNo(rgba, white, event)
	case reclassChooseMember:
		fillBox(rgba, 392, 40, 200, 32+(1+len(g.companions))*24, white)
		for i := 0; i <= len(g.companions); i++ {
			_, _, name, valid := g.reclassMemberAt(i)
			if !valid {
				continue
			}
			y := 56 + i*24
			if i == g.reclassCursor {
				drawGlyph(rgba, g.dlg.tx, 404, y, curGlyph, white)
			}
			for j, glyph := range name {
				drawGlyph(rgba, g.dlg.tx, 432+j*16, y, glyph, white)
			}
		}
	case reclassChooseClass:
		m, _, _, valid := g.reclassMemberAt(g.reclassMember)
		if !valid {
			return
		}
		targets := g.reclassTargets(event, m)
		fillBox(rgba, 392, 32, 200, 32+len(targets)*24, white)
		for i, class := range targets {
			y := 48 + i*24
			if i == g.reclassCursor {
				drawGlyph(rgba, g.dlg.tx, 404, y, curGlyph, white)
			}
			for j, glyph := range g.reclassClassGlyphs(event, class) {
				drawGlyph(rgba, g.dlg.tx, 432+j*16, y, glyph, white)
			}
		}
	}
}

func (g *Game) drawReclassYesNo(rgba []byte, white dq3data.Color, event *gamepack.ReclassEvent) {
	fillBox(rgba, 430, 220, 120, 68, white)
	ids := [2]string{event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo}
	for i, id := range ids {
		label, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		y := 232 + i*24
		if i == g.reclassCursor {
			drawGlyph(rgba, g.dlg.tx, 442, y, curGlyph, white)
		}
		for j, glyph := range label {
			if glyph >= 0xffed {
				continue
			}
			drawGlyph(rgba, g.dlg.tx, 470+j*16, y, int(glyph), white)
		}
	}
}
