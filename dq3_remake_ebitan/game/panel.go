package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 資訊面板(狀況 / 道具):命令窗選 狀況/道具 開啟,B 關。用已建的升級/背包/品名系統。
type panelKind int

const (
	panelNone panelKind = iota
	panelStatus
	panelItem
	panelEquip
)

const (
	itemActionList = iota
	itemActionMenu
	itemActionTarget
)

// drawStatus:依原版 D3TXT00 record407 的 22×12 格狀況窗，將主角角色 record
// 的七個持久能力與裝備衍生值填入 pack-owned anchors。沒有完整 pack 契約時
// fail closed，不畫猜測版黑色大框。
func (g *Game) drawStatus(rgba []byte, white dq3data.Color) {
	if g.pack == nil || g.dlg.tx == nil {
		return
	}
	layout, ok := g.pack.FieldStatusLayout()
	if !ok {
		return
	}
	glyphs, ok := g.pack.TextGlyphCodes(layout.TextID)
	if !ok || len(glyphs) == 0 {
		return
	}
	win := layout.Window
	fillPackBox(rgba, win, win.X, win.Y, win.Width, win.Height)
	drawGlyphGrid(rgba, g.dlg.tx, win.X, win.Y, win.Columns, win.LinesPerPage, glyphs, white)

	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	value := dq3data.Color{R: 255, G: 224, B: 32}
	drawStatusGlyphs(rgba, g.dlg.tx, layout.Name, g.heroName, white)
	drawStatusGlyphs(rgba, g.dlg.tx, layout.Class, layout.HeroClassGlyphs, white)
	sex := layout.MaleGlyphs
	if g.heroGender != 0 {
		sex = layout.FemaleGlyphs
	}
	drawStatusGlyphs(rgba, g.dlg.tx, layout.Sex, sex, white)
	for _, field := range []struct {
		anchor gamepack.GeometryAnchor
		value  int
	}{
		{layout.Level, level}, {layout.CurrentHP, g.heroHP}, {layout.CurrentMP, g.heroMP},
		{layout.Strength, int(g.heroStat[stats.STR])}, {layout.Agility, agi},
		{layout.Vitality, int(g.heroStat[stats.VIT])}, {layout.Intelligence, int(g.heroStat[stats.INT])},
		{layout.Luck, int(g.heroStat[stats.LUCK])}, {layout.MaxHP, maxHP},
		{layout.MaxMP, maxMP}, {layout.Attack, atk}, {layout.Defense, def},
		{layout.Experience, int(g.heroExp)},
	} {
		drawNumber(rgba, g.dlg.tx, field.anchor.X, field.anchor.Y, field.value, value)
	}
}

// drawGlyphGrid 畫 pack 文字定義中的固定欄列與原始控制碼。record407 的換行
// 是資料的一部分；renderer 不以字串內容猜測欄寬，也不把中文標籤寫回 Go。
func drawGlyphGrid(rgba []byte, tx *dq3data.Text, x, y, columns, lines int, glyphs []uint16, fg dq3data.Color) {
	if columns <= 0 || lines <= 0 {
		return
	}
	col, line := 0, 0
	for _, v := range glyphs {
		if line >= lines {
			return
		}
		switch v {
		case dq3data.TxtEnd, dq3data.TxtPage:
			return
		case dq3data.TxtNL, dq3data.TxtNL2:
			col, line = 0, line+1
			continue
		default:
			if v < dq3data.GlyphMax {
				drawGlyph(rgba, tx, x+col*dq3data.GlyphPx, y+line*dq3data.GlyphPx, int(v), fg)
			}
			col++
			if col >= columns {
				col, line = 0, line+1
			}
		}
	}
}

func drawStatusGlyphs(rgba []byte, tx *dq3data.Text, anchor gamepack.GeometryAnchor, glyphs []int, fg dq3data.Color) {
	for i, glyph := range glyphs {
		if glyph < 0 || glyph >= dq3data.GlyphMax {
			continue
		}
		drawGlyph(rgba, tx, anchor.X+i*dq3data.GlyphPx, anchor.Y, glyph, fg)
	}
}

// drawCmdStatus:命令窗開啟時顯示 pack 定義的縱列隊伍 H/M/等級狀態窗。
// 幾何與標籤 glyph 由 game pack 提供；沒有該資料時不畫猜測版 UI。
func (g *Game) drawCmdStatus(rgba []byte, white dq3data.Color) {
	g.drawPartyHUD(rgba, white)
}

// drawItems:持有道具清單(品名 = D3TXT00 rec=code+1),游標可選 → A 使用。空 → 不列。
func (g *Game) drawItems(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	g.panelHits.reset()
	for i, code := range g.inventory {
		if i >= 10 {
			break
		}
		y := 56 + i*22
		if i == g.panelCursor {
			drawGlyph(rgba, g.dlg.tx, 44, y, 11, yellow) // ► 游標
		}
		g.shop.drawItemName(rgba, 64, y, code, white)
		g.panelHits.add(44, y-3, ScreenW-80-8, 20, i)
	}
	switch g.itemActionStage {
	case itemActionMenu:
		g.drawItemActionMenu(rgba, white)
	case itemActionTarget:
		g.drawItemGiveTargets(rgba, white)
	}
}

func (g *Game) drawItemActionMenu(rgba []byte, white dq3data.Color) {
	if g.pack == nil {
		return
	}
	actions := g.pack.ItemActions()
	ids := [3]string{actions.TextIDs.Use, actions.TextIDs.Give, actions.TextIDs.Drop}
	fillBox(rgba, 400, 48, 160, 92, white)
	for i, id := range ids {
		glyphs, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return
		}
		y := 62 + i*24
		if i == g.itemActionCursor {
			drawGlyph(rgba, g.dlg.tx, 412, y, curGlyph, white)
		}
		for j, glyph := range glyphs {
			if glyph < 0xffed {
				drawGlyph(rgba, g.dlg.tx, 440+j*16, y, int(glyph), white)
			}
		}
	}
}

func (g *Game) drawItemGiveTargets(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 360, 48, 220, 32+len(g.companions)*24, white)
	for i, m := range g.companions {
		y := 64 + i*24
		if i == g.itemActionCursor {
			drawGlyph(rgba, g.dlg.tx, 372, y, curGlyph, white)
		}
		for j, glyph := range m.Name {
			drawGlyph(rgba, g.dlg.tx, 400+j*16, y, glyph, white)
		}
	}
}

func (g *Game) giveSelectedItem(target int) bool {
	if g.pack == nil || g.itemSelected < 0 || g.itemSelected >= len(g.inventory) ||
		target < 0 || target >= len(g.companions) {
		return false
	}
	m := g.companions[target]
	if m.itemCount() >= g.pack.ItemActions().PersonalInventorySlots {
		return false
	}
	code := g.inventory[g.itemSelected]
	g.inventory = append(g.inventory[:g.itemSelected], g.inventory[g.itemSelected+1:]...)
	m.Inventory = append(m.Inventory, code)
	if len(g.inventory) == 0 {
		g.panelCursor = 0
	} else if g.itemSelected >= len(g.inventory) {
		g.panelCursor = len(g.inventory) - 1
	} else {
		g.panelCursor = g.itemSelected
	}
	return true
}

// dropSelectedItem:原版 rec421「丟掉」動作。丟棄只改主角目前選取的
// inventory slot，不觸發任何道具效果、旗標或消耗動畫。
func (g *Game) dropSelectedItem() bool {
	if g.itemSelected < 0 || g.itemSelected >= len(g.inventory) {
		return false
	}
	g.inventory = append(g.inventory[:g.itemSelected], g.inventory[g.itemSelected+1:]...)
	if len(g.inventory) == 0 {
		g.panelCursor = 0
	} else if g.itemSelected >= len(g.inventory) {
		g.panelCursor = len(g.inventory) - 1
	} else {
		g.panelCursor = g.itemSelected
	}
	g.itemSelected = -1
	g.itemActionCursor = 0
	g.itemActionStage = itemActionList
	return true
}

func (g *Game) equipActorName(actor int) []int {
	if actor == 0 {
		if len(g.heroName) > 0 {
			return g.heroName
		}
		return classNames[0]
	}
	if actor > 0 && actor <= len(g.companions) {
		return g.companions[actor-1].Name
	}
	return nil
}

func (g *Game) equipActorClass(actor int) int {
	if actor == 0 {
		return 0
	}
	if actor > 0 && actor <= len(g.companions) {
		return g.companions[actor-1].Class
	}
	return -1
}

func (g *Game) equipActorSlots(actor int) *[4]int {
	if actor == 0 {
		return &g.equip
	}
	if actor <= 0 || actor > len(g.companions) {
		return nil
	}
	m := g.companions[actor-1]
	return &[4]int{m.Weapon, m.Armor, m.Shield, m.Head}
}

func (g *Game) setEquipActorSlot(actor, slot, code int) {
	if actor == 0 {
		g.equip[slot] = code
		return
	}
	m := g.companions[actor-1]
	switch slot {
	case 0:
		m.Weapon = code
	case 1:
		m.Armor = code
	case 2:
		m.Shield = code
	case 3:
		m.Head = code
	}
}

func (g *Game) equipActorInventory(actor int) *[]int {
	if actor == 0 {
		return &g.inventory
	}
	if actor <= 0 || actor > len(g.companions) {
		return nil
	}
	return &g.companions[actor-1].Inventory
}

func (g *Game) equipCandidateCount(actor int) int {
	items := g.equipActorInventory(actor)
	if items == nil {
		return 0
	}
	return len(*items)
}

// drawEquip:先選隊員，再顯示其四個裝備槽與該角色自己的未裝備物品。
func (g *Game) drawEquip(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	g.panelHits.reset()
	if g.panelActor < 0 {
		for actor := 0; actor <= len(g.companions); actor++ {
			y := 56 + actor*24
			if actor == g.panelCursor {
				drawGlyph(rgba, g.dlg.tx, 44, y, curGlyph, yellow)
			}
			x := 64
			for _, gi := range g.equipActorName(actor) {
				drawGlyph(rgba, g.dlg.tx, x, y, gi, white)
				x += dq3data.GlyphPx
			}
			g.panelHits.add(44, y-3, ScreenW-80-8, 22, actor)
		}
		return
	}

	x := 64
	for _, gi := range g.equipActorName(g.panelActor) {
		drawGlyph(rgba, g.dlg.tx, x, 52, gi, yellow)
		x += dq3data.GlyphPx
	}
	if slots := g.equipActorSlots(g.panelActor); slots != nil {
		for s, code := range slots {
			if code >= 0 {
				g.shop.drawItemName(rgba, 64+(s&1)*260, 76+(s>>1)*22, code, yellow)
			}
		}
	}
	items := g.equipActorInventory(g.panelActor)
	if items == nil {
		return
	}
	for i, code := range *items {
		if i >= 8 {
			break
		}
		y := 132 + i*22
		if i == g.panelCursor {
			drawGlyph(rgba, g.dlg.tx, 44, y, 11, yellow) // ►
		}
		g.shop.drawItemName(rgba, 64, y, code, white)
		g.panelHits.add(44, y-3, ScreenW-80-8, 20, i)
	}
}

// equipSelected:把所選角色自己的未裝備物品穿上；舊裝備回到同一角色物品欄。
func (g *Game) equipSelected() {
	items := g.equipActorInventory(g.panelActor)
	if g.panelActor < 0 || items == nil ||
		g.panelCursor < 0 || g.panelCursor >= len(*items) {
		return
	}
	code := (*items)[g.panelCursor]
	if g.shop.items == nil {
		return
	}
	slot := g.shop.items.EquipSlot(code)
	cls := g.equipActorClass(g.panelActor)
	if slot < 0 || slot >= 4 || !g.shop.items.CanEquip(code, cls) {
		return
	}
	slots := g.equipActorSlots(g.panelActor)
	if slots == nil {
		return
	}
	old := slots[slot]
	*items = append((*items)[:g.panelCursor], (*items)[g.panelCursor+1:]...)
	g.setEquipActorSlot(g.panelActor, slot, code)
	if old >= 0 {
		*items = append(*items, old)
	}
	if len(*items) == 0 {
		g.panelCursor = 0
	} else if g.panelCursor >= len(*items) {
		g.panelCursor = len(*items) - 1
	}
}
