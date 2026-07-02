package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"

// useSelectedItem:對道具面板游標指的道具套用使用效果(移植 main.c apply_item_use)。
// Go 版狀態範圍:藥草(HP)、蓋美拉翅膀(回鎮)、聖水(驅敵)、祈禱之戒(MP+損壞)已接;
// 解狀態(無狀態系統)、晝夜/劇情道具(無對應系統/需特定地點)→ 不消耗(對齊原版「無對應→不消耗」)。
func (g *Game) useSelectedItem() {
	if g.panelCursor < 0 || g.panelCursor >= len(g.inventory) {
		return
	}
	code := g.inventory[g.panelCursor]
	switch itemuse.KindOf(code) {
	case itemuse.HealHP: // 藥草:回復第一個未滿且未陣亡的隊員 HP
		if g.applyHealHP(code) {
			g.removeItems(code, 1)
			g.noticeCode, g.noticeTimer = code, 90
			g.clampPanelCursor()
		}
	case itemuse.ReturnTown: // 蓋美拉翅膀:回起始城鎮(阿里阿罕)
		g.removeItems(code, 1)
		g.panel = panelNone
		g.enterTownCty(0)
	case itemuse.Repel: // 聖水:驅弱敵 64 步
		g.repel = itemuse.HolySteps
		g.removeItems(code, 1)
		g.noticeCode, g.noticeTimer = code, 90
		g.clampPanelCursor()
	case itemuse.PrayerRing: // 祈禱之戒:回勇者 MP;每次使用 ~25.4% 損壞(RNG(256)≤0x40)。損壞才消耗
		if r := itemuse.PrayerMPAmount(g.heroMP, g.heroMaxMP()); r > 0 {
			g.heroMP += r
			g.noticeCode, g.noticeTimer = code, 90
		}
		if g.prng.Next(256) <= itemuse.PrayerBreakLE { // 損壞消失(忠實 RE:無「MP 滿則不用」閘)
			g.removeItems(code, 1)
			g.clampPanelCursor()
		}
	default:
		// 解狀態(無狀態系統)/ 晝夜 / 劇情道具(需特定地點)→ 不消耗,對齊原版
	}
}

// applyHealHP:對第一個未滿且未陣亡的隊員套藥草治療(勇者優先,再同伴)。回是否有人被治療。
func (g *Game) applyHealHP(code int) bool {
	_, maxHP, _, _, _ := g.heroStats()
	if h := itemuse.HealHPAmount(g.heroHP, maxHP); h > 0 { // 勇者
		g.heroHP += h
		return true
	}
	for _, m := range g.companions { // 同伴
		if h := itemuse.HealHPAmount(m.CurHP, m.MaxHP()); h > 0 {
			m.CurHP += h
			return true
		}
	}
	return false
}

// clampPanelCursor:道具消耗後,把游標夾回清單範圍。
func (g *Game) clampPanelCursor() {
	if g.panelCursor >= len(g.inventory) {
		g.panelCursor = len(g.inventory) - 1
	}
	if g.panelCursor < 0 {
		g.panelCursor = 0
	}
}
