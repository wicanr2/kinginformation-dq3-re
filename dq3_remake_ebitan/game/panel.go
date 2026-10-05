package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 資訊面板（狀況／道具）：多人道具先選持有者，再操作該角色個人物品欄。
type panelKind int

const (
	panelNone panelKind = iota
	panelStatus
	panelItem
	panelEquip
	panelStatusMenu
)

const (
	itemActionList = iota
	itemActionMenu
	itemActionTarget
	itemActionUseTarget
	itemActionGiveWait
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

// drawItems:持有者 selector 與角色局部道具清單（品名 = D3TXT00 rec=code+1）。
func (g *Game) drawItems(rgba []byte, white dq3data.Color) {
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	g.panelHits.reset()
	if g.panelActor < 0 {
		fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
		for actor := 0; actor <= len(g.companions); actor++ {
			y := 56 + actor*24
			if actor == g.panelCursor {
				drawGlyph(rgba, g.dlg.tx, 44, y, curGlyph, yellow)
			}
			for j, glyph := range g.equipActorName(actor) {
				drawGlyph(rgba, g.dlg.tx, 64+j*16, y, glyph, white)
			}
			g.panelHits.add(44, y-3, ScreenW-80-8, 20, actor)
		}
		return
	}
	g.drawNativeItems(rgba)
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
	count := 1 + len(g.companions)
	fillBox(rgba, 360, 48, 220, 32+count*24, white)
	for actor := 0; actor < count; actor++ {
		y := 64 + actor*24
		if actor == g.itemActionCursor {
			drawGlyph(rgba, g.dlg.tx, 372, y, curGlyph, white)
		}
		for j, glyph := range g.equipActorName(actor) {
			drawGlyph(rgba, g.dlg.tx, 400+j*16, y, glyph, white)
		}
		g.panelHits.add(372, y-3, 200, 20, actor)
	}
}

func (g *Game) drawItemUseTargets(rgba []byte, white dq3data.Color) {
	count := 1 + len(g.companions)
	fillBox(rgba, 360, 48, 220, 32+count*24, white)
	for actor := 0; actor < count; actor++ {
		y := 64 + actor*24
		if actor == g.itemActionCursor {
			drawGlyph(rgba, g.dlg.tx, 372, y, curGlyph, white)
		}
		for j, glyph := range g.equipActorName(actor) {
			drawGlyph(rgba, g.dlg.tx, 400+j*16, y, glyph, white)
		}
		g.panelHits.add(372, y-3, 200, 20, actor)
	}
}

func (g *Game) drawShopTargets(rgba []byte, white dq3data.Color) {
	if !g.shop.targeting {
		return
	}
	count := 1 + len(g.companions)
	fillBox(rgba, 360, 48, 220, 32+count*24, white)
	g.shop.targetHits.reset()
	for actor := 0; actor < count; actor++ {
		y := 64 + actor*24
		if actor == g.shop.targetCursor {
			drawGlyph(rgba, g.dlg.tx, 372, y, curGlyph, white)
		}
		for j, glyph := range g.equipActorName(actor) {
			drawGlyph(rgba, g.dlg.tx, 400+j*16, y, glyph, white)
		}
		g.shop.targetHits.add(372, y-3, 200, 20, actor)
	}
}

func (g *Game) purchaseShopItem(actor int) bool {
	if g.pack == nil || !g.shop.targeting || g.shop.pendingCode < 0 || g.shop.items == nil {
		return false
	}
	s := g.actorItemStore(actor)
	if s == nil {
		return false
	}
	code := g.shop.pendingCode
	price := g.shop.items.Price(code)
	if price < 0 || g.heroGold < price {
		return false
	}
	if _, ok := s.Add(code); !ok {
		return false
	}
	g.heroGold -= price
	g.shop.targeting, g.shop.pendingCode = false, -1
	g.completeSpecialShopPurchase(code)
	return true
}

type shopSellEntry struct{ code, position int }

func (g *Game) shopSellEntries(actor int) []shopSellEntry {
	var out []shopSellEntry
	for _, e := range g.actorItemEntries(actor) {
		out = append(out, shopSellEntry{e.Code, e.Position})
	}
	return out
}

func (g *Game) shopSellPrice(code int) (int, bool) {
	if g.pack == nil || g.shop.items == nil || code < 0 {
		return 0, false
	}
	return g.pack.ShopSellPrice(g.shop.kind, code, g.shop.items.Price(code))
}

func (g *Game) sellShopItem(actor, index int) bool {
	entries := g.shopSellEntries(actor)
	if index < 0 || index >= len(entries) {
		return false
	}
	entry := entries[index]
	price, ok := g.shopSellPrice(entry.code)
	if !ok {
		return false
	}
	s := g.actorItemStore(actor)
	if s == nil || !s.Remove(entry.position) {
		return false
	}
	g.heroGold += price
	g.completeSpecialShopSale(entry.code)
	return true
}

func (g *Game) drawShopSellItems(rgba []byte, white dq3data.Color) {
	entries := g.shopSellEntries(g.shop.sellActor)
	if entries == nil {
		return
	}
	fillBox(rgba, 24, 24, ScreenW-48, ScreenH-96, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	for i, entry := range entries {
		code := entry.code
		y := 40 + i*22
		if i == g.shop.sellCursor {
			drawGlyph(rgba, g.dlg.tx, 40, y, curGlyph, white)
		}
		g.shop.drawItemName(rgba, 64, y, code, white)
		if price, ok := g.shopSellPrice(code); ok {
			drawNumber(rgba, g.dlg.tx, 380, y, price, white)
		}
	}
	drawNumber(rgba, g.dlg.tx, 64, ScreenH-84, g.heroGold, yellow)
}

func (g *Game) giveSelectedItem(target int) bool {
	source, dest := g.actorItemStore(g.panelActor), g.actorItemStore(target)
	if g.pack == nil || source == nil || dest == nil || !source.Give(dest, g.itemSelected) {
		return false
	}
	if source == dest {
		g.panelCursor = len(source.Entries()) - 1
	} else {
		g.clampPanelCursor()
	}
	return true
}

// dropSelectedItem:原版 rec421「丟掉」動作。丟棄只改目前持有者的選定
// personal inventory slot，不觸發任何道具效果、旗標或消耗動畫。
func (g *Game) dropSelectedItem() bool {
	s := g.actorItemStore(g.panelActor)
	entry, valid := g.selectedItemEntry()
	if s == nil || g.pack == nil || !valid || !g.fieldActorAlive(g.panelActor) || !g.pack.ItemDropAllowed(entry) || !s.Remove(entry.Position) {
		return false
	}
	g.clampPanelCursor()
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
	store := g.actorItemStore(g.panelActor)
	if store == nil {
		return
	}
	for i, entry := range store.Inventory() {
		code := entry.Code
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

// equipSelected changes flags in place; candidates retain their physical identity.
func (g *Game) equipSelected() {
	s := g.actorItemStore(g.panelActor)
	if s == nil || g.shop.items == nil {
		return
	}
	entries := s.Inventory()
	if g.panelCursor < 0 || g.panelCursor >= len(entries) {
		return
	}
	entry := entries[g.panelCursor]
	if !g.shop.items.CanEquip(entry.Code, g.equipActorClass(g.panelActor)) {
		return
	}
	s.Wear(entry.Position)
	if g.panelCursor >= len(s.Inventory()) {
		g.panelCursor = len(s.Inventory()) - 1
	}
	if g.panelCursor < 0 {
		g.panelCursor = 0
	}
}
