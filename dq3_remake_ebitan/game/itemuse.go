package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
)

func (g *Game) selectedItemRequiresTarget() bool {
	items := g.equipActorInventory(g.panelActor)
	if g.pack == nil || items == nil || g.itemSelected < 0 || g.itemSelected >= len(*items) {
		return false
	}
	effect, ok := g.pack.ItemUseEffectByRawID((*items)[g.itemSelected])
	return ok && effect.EffectID == "clear_condition" && effect.TargetScope == "party_member"
}

// useSelectedPackItemOnTarget 執行已確認的 pack-owned 選人道具交易。原版驅毒草與
// 滿月草都先消耗 item，再檢查死亡／condition；因此無效果也不退還。
func (g *Game) useSelectedPackItemOnTarget(actor int) bool {
	items := g.equipActorInventory(g.panelActor)
	if g.pack == nil || items == nil || g.itemSelected < 0 || g.itemSelected >= len(*items) ||
		actor < 0 || actor > len(g.companions) {
		return false
	}
	code := (*items)[g.itemSelected]
	effect, ok := g.pack.ItemUseEffectByRawID(code)
	if !ok || effect.EffectID != "clear_condition" || effect.TargetScope != "party_member" ||
		!effect.Consume {
		return false
	}
	condition, ok := conditionBit(effect.ConditionID)
	if !ok {
		return false
	}
	g.panelCursor = g.itemSelected
	g.consumeSelectedItem(code)
	g.clampPanelCursor()
	alive, affected := g.heroHP > 0, g.heroConditions&condition != 0
	if actor > 0 {
		member := g.companions[actor-1]
		alive, affected = member.CurHP > 0, member.Conditions&condition != 0
		if alive && affected {
			member.Conditions &^= condition
		}
	} else if alive && affected {
		g.heroConditions &^= condition
	}
	if condition == conditionParalysis && !g.partyHasCondition(conditionParalysis) {
		g.paralysisSteps = 0
	}
	textID := effect.NoEffectTextID
	if alive && affected {
		textID = effect.SuccessTextID
	}
	name := g.equipActorName(actor)
	g.panel = panelNone
	g.itemActionStage, g.itemActionCursor, g.itemSelected = itemActionList, 0, -1
	if !g.openPackText(textID) {
		return false
	}
	if alive && affected {
		g.setDlgVarGlyphs(dq3data.TxtVarEnt, name)
	}
	return true
}

// useSelectedItem:對道具面板游標指的道具套用使用效果(移植 main.c apply_item_use)。
// Go 版狀態範圍:藥草(HP)、蓋美拉翅膀(回鎮)、聖水(驅敵)、祈禱之戒(MP+損壞)、
// 覺醒粉/蓋亞之劍/乾渴壺/妖精之笛/彩虹水滴(位置相關劇情道具)已接;
// 拉那魯達/黑暗之燈(切晝夜,不消耗)、驅毒草與滿月草(選人、先消耗再判定狀態)已接；
// 未由 pack 閉合的效果維持失敗即關閉且不消耗。
func (g *Game) useSelectedItem() {
	items := g.equipActorInventory(g.panelActor)
	if items == nil || g.panelCursor < 0 || g.panelCursor >= len(*items) {
		return
	}
	code := (*items)[g.panelCursor]
	if g.usePackItemEffect(code) {
		return
	}
	if g.useTrackedWorldObjectLocator(code) {
		return
	}
	if g.useQuestItemChain(code) {
		return
	}
	switch itemuse.KindOf(code) {
	case itemuse.HealHP: // 藥草:回復第一個未滿且未陣亡的隊員 HP
		if g.applyHealHP(code) {
			g.consumeSelectedItem(code)
			g.noticeCode, g.noticeTimer = code, 90
			g.clampPanelCursor()
		}
	case itemuse.ReturnTown: // 蓋美拉翅膀：地上回阿里阿罕，下層回拉達多姆外城 CTY79
		g.consumeSelectedItem(code)
		g.panel = panelNone
		dest := 0
		if g.layer == 1 {
			dest = 79
		}
		g.enterTownCty(dest)
	case itemuse.Repel: // 聖水:驅弱敵 64 步
		g.repel = itemuse.HolySteps
		g.consumeSelectedItem(code)
		g.noticeCode, g.noticeTimer = code, 90
		g.clampPanelCursor()
	case itemuse.PrayerRing: // 祈禱之戒:回勇者 MP;每次使用 ~25.4% 損壞(RNG(256)≤0x40)。損壞才消耗
		if r := itemuse.PrayerMPAmount(g.heroMP, g.heroMaxMP()); r > 0 {
			g.heroMP += r
			g.noticeCode, g.noticeTimer = code, 90
		}
		if g.prng.Next(256) <= itemuse.PrayerBreakLE { // 損壞消失(忠實 RE:無「MP 滿則不用」閘)
			g.consumeSelectedItem(code)
			g.clampPanelCursor()
		}
	case itemuse.Rainbow: // DQ3.EXE loc_14243：下層世界 (127,117) 使用，改 (126,117)=tile0x53
		if !g.inTown && g.layer == 1 && g.px == rainbowUseX && g.py == rainbowUseY {
			g.consumeSelectedItem(code)
			g.worldState |= worldStateRainbowBridge
			g.applyRainbowBridge()
			g.noticeCode, g.noticeTimer = code, 90
			g.clampPanelCursor()
		}
		// 否則需在下層利姆達爾西北盡頭地表使用,不消耗
	case itemuse.MagicBall:
		g.useMagicBall()
	case itemuse.Ranaruta: // 拉那魯達:白天↔黑夜切換(不消耗;城內用時重載當前 section NPC)
		g.toggleDaynight()
		g.noticeCode, g.noticeTimer = code, 90
	case itemuse.Mirror:
		g.useMirror()
	default:
		// 未接線效果不改狀態、不消耗；不得把此 fallback 當成原版精確 consumer。
	}
}

// usePackItemEffect 在舊相容派發器之前接管版本專屬道具選擇器。
// 引擎只實作有限的具名效果；未知效果或不合法 gate 均失敗即關閉且不改變狀態。
func (g *Game) usePackItemEffect(code int) bool {
	if g.pack == nil {
		return false
	}
	effect, ok := g.pack.ItemUseEffectByRawID(code)
	if !ok {
		return false
	}
	if effect.LocationKind == "overworld" && g.inTown {
		return true
	}
	switch effect.EffectID {
	case "grant_quest_item":
		if !g.inTown || g.cur == nil || g.curCty != effect.RequiredCTYRaw ||
			currentSceneSection(g.cur) != effect.RequiredSection {
			return true
		}
		if g.hasPartyItem(effect.GrantItemRawID) {
			return true
		}
		if !g.grantPartyItem(effect.GrantItemRawID) {
			return true
		}
		for _, flag := range effect.SetStoryFlagsRaw {
			g.setStoryFlag(flag, true)
		}
		g.noticeCode, g.noticeTimer = effect.GrantItemRawID, 90
	case "force_day_night_phase":
		if effect.DayNightClock == nil {
			return true
		}
		if g.isNight() {
			return true
		}
		if !effect.ResetDayNightSteps {
			return true
		}
		g.setDaynightClock(*effect.DayNightClock)
		if effect.Consume {
			g.consumeSelectedItem(code)
			g.clampPanelCursor()
		}
		g.noticeCode, g.noticeTimer = code, 90
	case "temporary_invisibility":
		if effect.StepCount <= 0 {
			return true
		}
		g.remoaru = effect.StepCount
		if effect.Consume {
			g.consumeSelectedItem(code)
			g.clampPanelCursor()
		}
		g.noticeCode, g.noticeTimer = code, 90
	case "reveal_world_map_patch":
		if g.inTown || g.layer != effect.RequiredLayer ||
			(effect.RequiredVehicle != "" && effect.RequiredVehicle != "ship") ||
			(effect.RequiredVehicle == "ship" && !g.shipAboard) || effect.UseTile == nil ||
			g.px != effect.UseTile.X || g.py != effect.UseTile.Y ||
			(effect.RequiredStoryFlagRaw >= 0 &&
				g.storyFlag(effect.RequiredStoryFlagRaw) != effect.RequiredStoryFlagSet) {
			return true
		}
		g.worldState |= uint16(effect.SetWorldStateMask)
		for _, flag := range effect.ClearStoryFlagsRaw {
			g.setStoryFlag(flag, false)
		}
		g.applyPackWorldMapPatches()
		g.applyPackDirectWorldMapPatch(effect)
		g.panel = panelNone
	case "open_facing_locked_door":
		if !g.inTown || g.cur == nil {
			return true
		}
		fx, fy := frontTile(g.px, g.py, g.facing)
		need := g.cur.doorTier(fx, fy)
		if need == 0 || effect.DoorKeyTier < need {
			return true
		}
		if g.cur.openDoor(fx, fy) {
			g.panel = panelNone
		}
	}
	return true
}

// applyPackWorldMapPatches rebuilds persistent original world-state tile
// replacements from validated pack data. It is called both on item use and
// after save/load so the map never depends on an in-memory-only mutation.
func (g *Game) applyPackWorldMapPatches() {
	if g.pack == nil {
		return
	}
	for _, effect := range g.pack.ItemUseEffects() {
		if effect.EffectID != "reveal_world_map_patch" || effect.MapPatch == nil ||
			g.worldState&uint16(effect.SetWorldStateMask) == 0 {
			continue
		}
		g.applyWorldMapPatch(effect.MapPatch)
	}
}

// applyPackDirectWorldMapPatch applies the immediate viewport table used by
// the original item handler. It is intentionally not called from save/load:
// a reload must rebuild only the persistent world-state table.
func (g *Game) applyPackDirectWorldMapPatch(effect *gamepack.ItemUseEffect) {
	if effect == nil || effect.EffectID != "reveal_world_map_patch" ||
		effect.DirectMapPatch == nil {
		return
	}
	g.applyWorldMapPatch(effect.DirectMapPatch)
}

func (g *Game) applyWorldMapPatch(patch *gamepack.WorldMapPatch) {
	if patch == nil {
		return
	}
	var scene *Scene
	switch patch.Layer {
	case 0:
		scene = g.over
	case 1:
		scene = g.loadUnder()
	}
	if scene == nil || patch.Origin.X+patch.Width > scene.w ||
		patch.Origin.Y+patch.Height > scene.h {
		return
	}
	if scene.override == nil {
		scene.override = map[int]int{}
	}
	for y := 0; y < patch.Height; y++ {
		for x := 0; x < patch.Width; x++ {
			i := y*patch.Width + x
			scene.override[(patch.Origin.Y+y)*scene.w+patch.Origin.X+x] = patch.TilesRaw[i]
		}
	}
}

const (
	ctyTemptationCave   = 30
	magicBallSection    = 0
	magicBallUseY       = 13
	magicBallLeftX      = 8
	magicBallRightX     = 9
	magicBallIntactFlag = 0x51
)

// useMagicBall 還原 DQ3.EXE file 0x58cd..0x5990。
// 原版只檢查城鎮模式、CTY30 sec0、玩家 Y=13、X=8/9，沒有 facing gate。
// 成功時 CLR flag0x51；通用 file 0x4703 重建規則會將四格牆 tile+1 並清 event subid。
func (g *Game) useMagicBall() {
	if !g.inTown || g.cur == nil || g.curCty != ctyTemptationCave ||
		g.cur.sec != magicBallSection || g.py != magicBallUseY ||
		(g.px != magicBallLeftX && g.px != magicBallRightX) ||
		!g.storyFlag(magicBallIntactFlag) {
		return
	}
	g.panel = panelNone
	g.setStoryFlag(magicBallIntactFlag, false)
	g.cur.applyClearedEventTiles(g.storyFlag)
	g.consumeSelectedItem(itemuse.ItemMagicBall)
	g.clampPanelCursor()
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
	if !g.hasPartyItem(itemModChangeStaff) && !g.grantPartyItem(itemModChangeStaff) {
		// 原版在獎勵 writer 設 DS:0726=1 後跳回敗／逃回滾分支；
		// 滿載不得留下成功旗標而永久失去變化之杖。
		g.setStoryFlag(0x10, false)
		g.setStoryFlag(0x42, true)
		g.reloadTownDaynight()
		return
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
	items := g.equipActorInventory(g.panelActor)
	if items == nil {
		g.panelCursor = 0
		return
	}
	if g.panelCursor >= len(*items) {
		g.panelCursor = len(*items) - 1
	}
	if g.panelCursor < 0 {
		g.panelCursor = 0
	}
}

// consumeSelectedItem 移除目前道具持有者的確切 panelCursor 欄位；進入選人
// modal 前會先把 itemSelected 複製回 panelCursor，因此不需按 item id 掃描。
func (g *Game) consumeSelectedItem(code int) bool {
	items := g.equipActorInventory(g.panelActor)
	if items == nil {
		return false
	}
	idx := g.panelCursor
	if idx < 0 || idx >= len(*items) || (*items)[idx] != code {
		return false
	}
	*items = append((*items)[:idx], (*items)[idx+1:]...)
	g.panelCursor = idx
	return true
}
