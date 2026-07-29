package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

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

// drawStatus:主角狀況(等級/HP/EXP/G)。glyph 22='H'、標籤用數字為主(中文標籤 glyph 之後補)。
func (g *Game) drawStatus(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	level, maxHP, atk, def, agi := g.heroStats()
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	x := 64
	// 每列:名稱 glyph(D3TXT00 對應字)+ 數值。這裡以 H 標 HP,其餘用數值列(中文標籤 glyph 之後補)。
	drawNumber(rgba, g.dlg.tx, x, 60, level, yellow) // 等級
	drawGlyph(rgba, g.dlg.tx, x, 90, 22, white)      // H
	drawNumber(rgba, g.dlg.tx, x+24, 90, g.heroHP, white)
	drawNumber(rgba, g.dlg.tx, x+120, 90, maxHP, white)
	drawNumber(rgba, g.dlg.tx, x, 120, atk, white)     // 攻
	drawNumber(rgba, g.dlg.tx, x+120, 120, def, white) // 防
	drawNumber(rgba, g.dlg.tx, x+240, 120, agi, white) // 敏
	drawNumber(rgba, g.dlg.tx, x, 150, int(g.heroExp), white)
	drawNumber(rgba, g.dlg.tx, x, 180, g.heroGold, yellow)
	// 主線進度階段(0..9;第一個未完成里程碑處停)—— 可見的破關進度指標
	drawNumber(rgba, g.dlg.tx, x, 210, g.progressStage(), yellow)
}

// drawCmdStatus:命令窗開啟時左下同時顯示的 HP/MP/等級小狀態窗(DOSBox oracle:
// dosbox/verify_open_21_after_space.png,黑底白框 3 列 H/M/勇+數值)。glyph 22='H'、
// 27='M'、106='勇'(對齊 classNames[0][0] 勇者職業首字,借代等級列標籤)。
func (g *Game) drawCmdStatus(rgba []byte, white dq3data.Color) {
	const gpx = dq3data.GlyphPx
	const rh = 18
	x, y := 48, ScreenH-3*rh-gpx-16
	w, h := 5*gpx, 3*rh+gpx
	fillBox(rgba, x-gpx/2, y-gpx/2, w, h, white)
	level, _, _, _, _ := g.heroStats()
	drawGlyph(rgba, g.dlg.tx, x, y, 22, white) // H
	drawNumber(rgba, g.dlg.tx, x+2*gpx, y, g.heroHP, white)
	drawGlyph(rgba, g.dlg.tx, x, y+rh, 27, white) // M
	drawNumber(rgba, g.dlg.tx, x+2*gpx, y+rh, g.heroMP, white)
	drawGlyph(rgba, g.dlg.tx, x, y+2*rh, 106, white) // 勇(等級列)
	drawNumber(rgba, g.dlg.tx, x+2*gpx, y+2*rh, level, white)
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
