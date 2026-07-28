package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"

// useSelectedItem:對道具面板游標指的道具套用使用效果(移植 main.c apply_item_use)。
// Go 版狀態範圍:藥草(HP)、蓋美拉翅膀(回鎮)、聖水(驅敵)、祈禱之戒(MP+損壞)、
// 覺醒粉/蓋亞之劍/乾渴壺/妖精之笛/彩虹水滴(位置相關劇情道具)已接;
// 拉那魯達/黑暗之燈(切晝夜,不消耗)已接;解狀態(無狀態系統)→ 不消耗(對齊原版「無對應→不消耗」)。
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
	case itemuse.Awaken: // 覺醒粉:諾阿尼魯村(CTY4)解全村催眠
		if g.inTown && g.curCty == 4 {
			g.removeItems(code, 1)
			g.flags[0x31] = true
			g.noticeCode, g.noticeTimer = code, 90
			g.clampPanelCursor()
		}
		// 否則此處用不上,不消耗(對齊原版)
	case itemuse.Gaia: // 蓋亞之劍:地表小火山旁 → 開通往尼羅肯特(武器,不消耗)
		if !g.inTown {
			g.flags[0x32] = true
			g.noticeCode, g.noticeTimer = code, 90
		}
	case itemuse.Drain: // 乾渴壺:地表四島礁 → 吸乾海水顯現祠堂(不消耗,可重用)
		if !g.inTown {
			g.flags[0x33] = true
			g.noticeCode, g.noticeTimer = code, 90
		}
	case itemuse.FairyFlute: // 妖精之笛:魯比斯之塔(CTY82)解詛咒 → 得精靈的守護 0x74(不消耗)
		if g.inTown && g.curCty == 82 {
			const itemSpiritGuard = 0x74
			if !g.hasItem(itemSpiritGuard) {
				g.inventory = append(g.inventory, itemSpiritGuard)
			}
			g.flags[0x34] = true
			g.noticeCode, g.noticeTimer = itemSpiritGuard, 90
		}
	case itemuse.Rainbow: // DQ3.EXE loc_14243：下層世界 (127,117) 使用，改 (126,117)=tile0x53
		if !g.inTown && g.layer == 1 && g.px == rainbowUseX && g.py == rainbowUseY {
			g.removeItems(code, 1)
			g.worldState |= worldStateRainbowBridge
			g.applyRainbowBridge()
			g.noticeCode, g.noticeTimer = code, 90
			g.clampPanelCursor()
		}
		// 否則需在下層利姆達爾西北盡頭地表使用,不消耗
	case itemuse.Ranaruta: // 拉那魯達:白天↔黑夜切換(不消耗;城內用時重載當前 section NPC)
		g.toggleDaynight()
		g.noticeCode, g.noticeTimer = code, 90
	case itemuse.DarkLamp: // 黑暗之燈:強制黑夜(不消耗;城內用時重載當前 section NPC)
		g.setDaynight(2)
		g.noticeCode, g.noticeTimer = code, 90
	case itemuse.Mirror:
		g.useMirror()
	default:
		// 解狀態(無狀態系統)→ 不消耗,對齊原版
	}
}

const (
	rainbowUseX, rainbowUseY              = 0x7f, 0x75
	rainbowBridgeX, rainbowBridgeY        = 0x7e, 0x75
	rainbowBridgeTile                     = 0x53
	worldStateRainbowBridge        uint16 = 0x40
)

func (g *Game) applyRainbowBridge() {
	if g.worldState&worldStateRainbowBridge == 0 {
		return
	}
	u := g.loadUnder()
	if u == nil {
		return
	}
	if u.override == nil {
		u.override = map[int]int{}
	}
	u.override[rainbowBridgeY*u.w+rainbowBridgeX] = rainbowBridgeTile
}

const (
	ctySamanosa        = 44
	mirrorSection      = 1
	mirrorX, mirrorY   = 14, 7
	monsterBossTroll   = 89
	itemModChangeStaff = 0x62
)

// useMirror 還原 DQ3.EXE file 0x5682 的沙曼歐莎假王事件入口。
// 原版只檢查 flag42、夜晚、CTY44 sec1、玩家座標(14,7)，沒有 facing gate；
// 拉之鏡是重要物品，成功或失敗都不消耗。
func (g *Game) useMirror() {
	if g.mirrorStage != 0 || !g.storyFlag(0x42) || !g.isNight() ||
		!g.inTown || g.curCty != ctySamanosa || g.cur == nil ||
		g.cur.sec != mirrorSection || g.px != mirrorX || g.py != mirrorY {
		return
	}
	g.panel = panelNone
	g.setStoryFlag(0x42, false)
	g.setStoryFlag(0x10, true)
	g.reloadTownDaynight()
	if g.dlg.Open(97) {
		g.mirrorStage = 1
	}
}

// advanceMirrorEvent 串接原版 rec97 → rec98 → monster89。
func (g *Game) advanceMirrorEvent() {
	switch g.mirrorStage {
	case 1:
		if g.dlg.Open(98) {
			g.mirrorStage = 2
		}
	case 2:
		if g.startBossBattle(monsterBossTroll) {
			g.mirrorStage = 3
		}
	}
}

// settleMirrorBattle 還原 0x56ed..0x5732 的勝敗交易。
// 敗/逃：clear10、restore42；勝：強制白天、clear21、set22、rec99。
func (g *Game) settleMirrorBattle() {
	if g.mirrorStage != 3 || g.battle.monID != monsterBossTroll {
		return
	}
	g.mirrorStage = 0
	if g.battle.result != 1 {
		g.setStoryFlag(0x10, false)
		g.setStoryFlag(0x42, true)
		g.reloadTownDaynight()
		return
	}
	if !g.hasItem(itemModChangeStaff) {
		g.inventory = append(g.inventory, itemModChangeStaff)
	}
	g.setStoryFlag(0x21, false)
	g.setStoryFlag(0x22, true)
	g.setDaynight(0)
	g.dlg.Open(99)
	g.noticeCode, g.noticeTimer = itemModChangeStaff, 120
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
