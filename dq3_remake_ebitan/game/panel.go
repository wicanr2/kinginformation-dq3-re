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
	drawNumber(rgba, g.dlg.tx, x, 120, atk, white)      // 攻
	drawNumber(rgba, g.dlg.tx, x+120, 120, def, white)  // 防
	drawNumber(rgba, g.dlg.tx, x+240, 120, agi, white)  // 敏
	drawNumber(rgba, g.dlg.tx, x, 150, int(g.heroExp), white)
	drawNumber(rgba, g.dlg.tx, x, 180, g.heroGold, yellow)
}

// drawItems:持有道具清單(品名 = D3TXT00 rec=code+1)。空 → 不列。
func (g *Game) drawItems(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	for i, code := range g.inventory {
		if i >= 10 {
			break
		}
		g.shop.drawItemName(rgba, 64, 56+i*22, code, white)
	}
}

// drawEquip:上方現裝備(武器/鎧),下方持有可裝備品(游標選 → A 裝上)。
func (g *Game) drawEquip(rgba []byte, white dq3data.Color) {
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	// 現裝備(武器 + 鎧;空槽不畫名)
	for s := 0; s < 2; s++ {
		if g.equip[s] != 0 || s == 0 {
			g.shop.drawItemName(rgba, 64, 56+s*22, g.equip[s], yellow)
		}
	}
	// 持有可裝備品(游標)
	for i, code := range g.inventory {
		if i >= 8 {
			break
		}
		y := 120 + i*22
		if i == g.panelCursor {
			drawGlyph(rgba, g.dlg.tx, 44, y, 11, yellow) // ►
		}
		g.shop.drawItemName(rgba, 64, y, code, white)
	}
}

// equipSelected:把游標指的持有品裝上其部位(武器/鎧/盾/兜)。
func (g *Game) equipSelected() {
	if g.panelCursor < 0 || g.panelCursor >= len(g.inventory) {
		return
	}
	code := g.inventory[g.panelCursor]
	if g.shop.items == nil {
		return
	}
	slot := g.shop.items.EquipSlot(code)
	if slot >= 0 && slot < 4 && g.shop.items.CanEquip(code, 0) {
		g.equip[slot] = code
	}
}
