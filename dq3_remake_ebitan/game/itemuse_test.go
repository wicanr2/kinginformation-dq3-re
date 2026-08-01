package game

import (
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
)

// 藥草治療:勇者受傷 → 用藥草回 30(封頂 maxHP)、消耗 1 個。
func TestUseHerbHeals(t *testing.T) {
	g := &Game{}
	_, maxHP, _, _, _ := g.heroStats() // level1 勇者 maxHP
	g.heroHP = 1
	g.inventory = []int{itemuse.ItemHerb, itemuse.ItemHerb}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()

	want := 1 + itemuse.HealHPAmount(1, maxHP)
	if g.heroHP != want {
		t.Errorf("藥草後 HP=%d, want %d", g.heroHP, want)
	}
	if len(g.inventory) != 1 {
		t.Errorf("藥草應消耗 1 個,剩 %d", len(g.inventory))
	}
}

// 藥草滿血不消耗:HP 已滿 → 用藥草無效、不消耗(對齊原版「無人需要治療 → 不消耗」)。
func TestUseHerbFullNoConsume(t *testing.T) {
	g := &Game{}
	_, maxHP, _, _, _ := g.heroStats()
	g.heroHP = maxHP
	g.inventory = []int{itemuse.ItemHerb}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if len(g.inventory) != 1 {
		t.Errorf("滿血用藥草不應消耗,剩 %d", len(g.inventory))
	}
}

// 聖水:使用 → repel = 64 步、消耗。
func TestUseHolyWaterRepel(t *testing.T) {
	g := &Game{}
	g.inventory = []int{itemuse.ItemHolyWater}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.repel != itemuse.HolySteps {
		t.Errorf("聖水後 repel=%d, want %d", g.repel, itemuse.HolySteps)
	}
	if len(g.inventory) != 0 {
		t.Errorf("聖水應消耗,剩 %d", len(g.inventory))
	}
}

func TestPackDarkLampOriginalGateAndTransaction(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	effect, ok := pack.ItemUseEffectByRawID(0x5f)
	if !ok {
		t.Fatal("黑暗之燈 pack effect missing")
	}
	use := func(g *Game) {
		t.Helper()
		g.inventory = []int{effect.ItemRawID}
		g.panel, g.panelCursor = panelItem, 0
		g.useSelectedItem()
	}

	overworld := &Game{pack: pack, dnPhase: 0, dnStep: 47}
	use(overworld)
	if overworld.dnPhase != effect.DayNightPhase || overworld.dnStep != 0 ||
		!overworld.hasItem(effect.ItemRawID) || overworld.noticeCode != effect.ItemRawID {
		t.Fatalf("地表使用黑暗燈 transaction 錯：phase=%d step=%d item=%v notice=%d",
			overworld.dnPhase, overworld.dnStep,
			overworld.hasItem(effect.ItemRawID), overworld.noticeCode)
	}

	alreadyNight := &Game{pack: pack, dnPhase: effect.DayNightPhase, dnStep: 31}
	use(alreadyNight)
	if alreadyNight.dnPhase != effect.DayNightPhase || alreadyNight.dnStep != 31 ||
		alreadyNight.noticeTimer != 0 || !alreadyNight.hasItem(effect.ItemRawID) {
		t.Fatalf("已是夜晚不得重置 clock 或消耗：phase=%d step=%d notice=%d item=%v",
			alreadyNight.dnPhase, alreadyNight.dnStep, alreadyNight.noticeTimer,
			alreadyNight.hasItem(effect.ItemRawID))
	}

	town := &Game{pack: pack, inTown: true, dnPhase: 0, dnStep: 47}
	use(town)
	if town.dnPhase != 0 || town.dnStep != 47 || town.noticeTimer != 0 ||
		!town.hasItem(effect.ItemRawID) {
		t.Fatalf("城內使用不得繞過原版 overworld gate：phase=%d step=%d notice=%d item=%v",
			town.dnPhase, town.dnStep, town.noticeTimer, town.hasItem(effect.ItemRawID))
	}
}

func TestPackInvisibilityGrassConsumesAndSharesRemoaruTimer(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	effect, ok := pack.ItemUseEffectByRawID(0x5d)
	if !ok {
		t.Fatal("隱身草 pack effect missing")
	}
	for _, inTown := range []bool{false, true} {
		g := &Game{pack: pack, inTown: inTown, inventory: []int{0x5d},
			panel: panelItem, panelCursor: 0}
		g.useSelectedItem()
		if g.remoaru != 0x19 || g.hasItem(0x5d) || g.noticeCode != 0x5d {
			t.Fatalf("inTown=%v 隱身草 transaction 錯：timer=%d item=%v notice=%d",
				inTown, g.remoaru, g.hasItem(0x5d), g.noticeCode)
		}
	}
	if effect.StepCount != 0x19 || !effect.Consume {
		t.Fatalf("隱身草 pack effect 錯：%+v", effect)
	}
}

// 祈禱之戒:回勇者 MP(未滿補 30 封頂);損壞為機率,MP 一定被回復。
func TestUsePrayerRingMP(t *testing.T) {
	g := &Game{}
	g.prng.Seed(0x1357)
	g.heroExp = 5000 // 拉高等級讓勇者有 MP
	g.heroMP = 0
	max := g.heroMaxMP()
	if max == 0 {
		t.Skip("此等級勇者無 MP")
	}
	g.inventory = []int{itemuse.ItemPrayerRing}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.heroMP != itemuse.PrayerMPAmount(0, max) {
		t.Errorf("祈禱之戒後 MP=%d, want %d", g.heroMP, itemuse.PrayerMPAmount(0, max))
	}
}

func TestUseMagicBallBreaksTemptationCaveWall(t *testing.T) {
	const w, h = 12, 18
	cells := make([]int, w*h)
	hi := make([]byte, w*h)
	attr := make([]uint16, 26)
	attr[22], attr[24] = 0x0009, 0x0009
	for _, p := range [][3]int{{8, 14, 24}, {9, 14, 24}, {8, 15, 22}, {9, 15, 22}} {
		cells[p[1]*w+p[0]] = p[2]
	}
	sc := &Scene{
		w: w, h: h, attr: &dq3data.BlockAttr{A: attr},
		tileAt: func(x, y int) int { return cells[y*w+x] },
		hiMap:  hi, events: [][3]int{{1, 1, magicBallIntactFlag}}, sec: magicBallSection,
	}
	g := &Game{cur: sc, inTown: true, curCty: ctyTemptationCave, px: 8, py: magicBallUseY}
	g.initStoryBits()
	g.inventory = []int{itemuse.ItemMagicBall}
	g.panel = panelItem
	g.useSelectedItem()

	if g.hasItem(itemuse.ItemMagicBall) || g.storyFlag(magicBallIntactFlag) {
		t.Fatalf("成功破牆後應消耗魔法球並清 flag0x51")
	}
	for _, p := range [][3]int{{8, 14, 25}, {9, 14, 25}, {8, 15, 23}, {9, 15, 23}} {
		if got := sc.tileIdx(p[0], p[1]); got != p[2] {
			t.Errorf("破牆 tile (%d,%d)=%d, want %d", p[0], p[1], got, p[2])
		}
	}
}

func TestUseMagicBallWrongPositionDoesNotConsume(t *testing.T) {
	g := &Game{cur: &Scene{sec: magicBallSection}, inTown: true, curCty: ctyTemptationCave, px: 8, py: 12}
	g.initStoryBits()
	g.inventory = []int{itemuse.ItemMagicBall}
	g.panel = panelItem
	g.useSelectedItem()
	if !g.hasItem(itemuse.ItemMagicBall) || !g.storyFlag(magicBallIntactFlag) {
		t.Fatal("非原版指定座標使用魔法球不應消耗或清 flag0x51")
	}
}

func TestTemptationCaveRealDataRebuildsClearedWall(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(dir + "/CTY30.DAT"); err != nil {
		t.Skipf("無 CTY30 原始素材：%v", err)
	}
	assets := os.DirFS(dir)
	rd := func(n string) []byte { d, _ := os.ReadFile(dir + "/" + n); return d }
	pal := dq3data.DecodePalette(rd("DQ3.PAL"), 256)
	manBLS := rd("DQ3MAN.BLS")

	intact, err := loadTownSceneSec(assets, pal, manBLS, ctyTemptationCave, mapBlkNum[ctyTemptationCave],
		magicBallSection, 0, func(id int) bool { return id == magicBallIntactFlag })
	if err != nil {
		t.Fatal(err)
	}
	cleared, err := loadTownSceneSec(assets, pal, manBLS, ctyTemptationCave, mapBlkNum[ctyTemptationCave],
		magicBallSection, 0, func(int) bool { return false })
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range [][4]int{{8, 14, 24, 25}, {9, 14, 24, 25}, {8, 15, 22, 23}, {9, 15, 22, 23}} {
		if got := intact.tileIdx(p[0], p[1]); got != p[2] {
			t.Errorf("原始牆 (%d,%d)=%d, want %d", p[0], p[1], got, p[2])
		}
		if got := cleared.tileIdx(p[0], p[1]); got != p[3] {
			t.Errorf("重建破牆 (%d,%d)=%d, want %d", p[0], p[1], got, p[3])
		}
		if cleared.hiMap[p[1]*cleared.w+p[0]]&0x1f != 0 {
			t.Errorf("重建破牆 (%d,%d) 未清 event subid", p[0], p[1])
		}
	}
}

// 蓋亞之劍(0x0f):在地表使用 → 設 flag 0x32,不消耗(武器)。
func TestUseGaiaOnField(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown = false
	g.inventory = []int{0x0f}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if !g.flags[0x32] {
		t.Errorf("蓋亞之劍在地表應設 flag 0x32")
	}
	if len(g.inventory) != 1 {
		t.Errorf("蓋亞之劍不應消耗,剩 %d", len(g.inventory))
	}
}

// 蓋亞之劍:在城鎮內使用 → 無效果、不設 flag。
func TestUseGaiaInTownNoEffect(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown = true
	g.inventory = []int{0x0f}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.flags[0x32] {
		t.Errorf("蓋亞之劍在城鎮內不應設 flag 0x32")
	}
	if len(g.inventory) != 1 {
		t.Errorf("蓋亞之劍不應消耗,剩 %d", len(g.inventory))
	}
}

func thirstyPitcherGame(t *testing.T) (*Game, *gamepack.ItemUseEffect) {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	effect, ok := g.pack.ItemUseEffectByRawID(0x5e)
	if !ok || effect.EffectID != "reveal_world_map_patch" || effect.UseTile == nil {
		t.Fatalf("乾渴壺 pack effect 不完整：%+v", effect)
	}
	g.showTitle, g.openingIdx = false, -1
	g.inTown, g.cur, g.layer = false, g.over, effect.RequiredLayer
	g.px, g.py = effect.UseTile.X, effect.UseTile.Y
	g.shipOwned, g.shipAboard = true, true
	g.inventory = []int{effect.ItemRawID}
	g.panel, g.panelCursor = panelItem, 0
	return g, effect
}

func TestUseThirstyPitcherRequiresOriginalShipTileAndRevealsPatch(t *testing.T) {
	g, effect := thirstyPitcherGame(t)
	g.useSelectedItem()
	if g.worldState&uint16(effect.SetWorldStateMask) == 0 || g.panel != panelNone {
		t.Fatalf("乾渴壺成功後 state/panel 錯：state=%#x panel=%d", g.worldState, g.panel)
	}
	if len(g.inventory) != 1 || g.inventory[0] != effect.ItemRawID {
		t.Fatalf("原版乾渴壺不消耗：inventory=%v", g.inventory)
	}
	patch := effect.MapPatch
	for y := 0; y < patch.Height; y++ {
		for x := 0; x < patch.Width; x++ {
			want := patch.TilesRaw[y*patch.Width+x]
			if got := g.over.tileIdx(patch.Origin.X+x, patch.Origin.Y+y); got != want {
				t.Fatalf("world patch (%d,%d)=%#x，want %#x",
					patch.Origin.X+x, patch.Origin.Y+y, got, want)
			}
		}
	}
}

func TestUseThirstyPitcherFailsClosedOffTileOrOffShip(t *testing.T) {
	for _, tc := range []struct {
		name string
		edit func(*Game)
	}{
		{"off tile", func(g *Game) { g.px++ }},
		{"off ship", func(g *Game) { g.shipAboard = false }},
		{"in town", func(g *Game) { g.inTown = true }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, effect := thirstyPitcherGame(t)
			tc.edit(g)
			g.useSelectedItem()
			if g.worldState&uint16(effect.SetWorldStateMask) != 0 || len(g.inventory) != 1 {
				t.Fatalf("錯誤 gate 不得改狀態：state=%#x inventory=%v", g.worldState, g.inventory)
			}
		})
	}
}

// 妖精之笛(0x77):在魯比斯之塔(CTY82)使用 → 設 flag 0x34、給 0x74,不消耗。
func TestUseFairyFluteInRubissTower(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown, g.curCty = true, 82
	g.inventory = []int{0x77}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if !g.flags[0x34] {
		t.Errorf("妖精之笛在魯比斯之塔應設 flag 0x34")
	}
	if !g.hasItem(0x74) {
		t.Errorf("妖精之笛應給精靈的守護 0x74")
	}
	if len(g.inventory) != 2 { // 原道具 + 新給的 0x74
		t.Errorf("妖精之笛不應消耗,應共 2 個道具,剩 %d", len(g.inventory))
	}
}

// 妖精之笛:已持有 0x74 時再使用 → 不重複給。
func TestUseFairyFluteAlreadyHasSpiritGuard(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown, g.curCty = true, 82
	g.inventory = []int{0x77, 0x74}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if !g.flags[0x34] {
		t.Errorf("妖精之笛在魯比斯之塔應設 flag 0x34")
	}
	if len(g.inventory) != 2 {
		t.Errorf("已持有 0x74 不應重複給,應仍 2 個道具,剩 %d", len(g.inventory))
	}
}

// 妖精之笛:位置不符 → 無效果、不設 flag。
func TestUseFairyFluteWrongPlaceNoEffect(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown, g.curCty = true, 4
	g.inventory = []int{0x77}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.flags[0x34] {
		t.Errorf("妖精之笛位置不符不應設 flag 0x34")
	}
	if len(g.inventory) != 1 {
		t.Errorf("妖精之笛位置不符不應給道具,剩 %d", len(g.inventory))
	}
}

// 彩虹水滴(0x75):原版只接受下層 (127,117)，並把西側 (126,117) 改成橋 tile0x53。
func TestUseRainbowAtOriginalCoordinate(t *testing.T) {
	g := r4Game(t)
	g.inTown, g.layer, g.px, g.py = false, 1, rainbowUseX, rainbowUseY
	g.cur = g.loadUnder()
	g.inventory = []int{0x75}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.worldState&worldStateRainbowBridge == 0 {
		t.Errorf("彩虹水滴應設原版 world state bit0x40")
	}
	if got := g.cur.tileIdx(rainbowBridgeX, rainbowBridgeY); got != rainbowBridgeTile {
		t.Errorf("橋 tile=%02x, want %02x", got, rainbowBridgeTile)
	}
	if len(g.inventory) != 0 {
		t.Errorf("正確位置應消耗彩虹水滴,剩 %d", len(g.inventory))
	}
}

// 同在下層但不是原版座標，也不得消耗或改世界狀態。
func TestUseRainbowWrongCoordinateNoConsume(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.inTown, g.layer, g.px, g.py = false, 1, rainbowUseX-1, rainbowUseY
	g.inventory = []int{0x75}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if g.worldState != 0 {
		t.Errorf("錯誤位置不應改 world state: %04x", g.worldState)
	}
	if len(g.inventory) != 1 {
		t.Errorf("錯誤位置不應消耗,剩 %d", len(g.inventory))
	}
}

func teidonLockedDoorGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 20, mapBlkNum[20], 0, 2, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.cur, g.town, g.inTown, g.curCty, g.dlg.tx = sc, sc, true, 20, sc.dlgText
	g.px, g.py, g.facing = 17, 5, 1
	if tier := sc.doorTier(17, 4); tier != 3 {
		t.Fatalf("CTY20 door tier=%d, want 3", tier)
	}
	return g
}

func TestPackDoorKeyRequiresFormalItemUse(t *testing.T) {
	g := teidonLockedDoorGame(t)
	finalKey, ok := g.pack.ItemUseEffectByRawID(0x57)
	if !ok || finalKey.EffectID != "open_facing_locked_door" || finalKey.DoorKeyTier != 3 {
		t.Fatalf("final-key pack effect mismatch: %+v", finalKey)
	}
	g.inventory = []int{finalKey.ItemRawID}
	g.examine()
	if tier := g.cur.doorTier(17, 4); tier != 3 {
		t.Fatalf("調查不得取代原版鑰匙使用，door tier=%d", tier)
	}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if tier := g.cur.doorTier(17, 4); tier != 0 || !g.hasItem(finalKey.ItemRawID) || g.panel != panelNone {
		t.Fatalf("正式使用最終鑰匙 transaction 錯：tier=%d held=%v panel=%d",
			tier, g.hasItem(finalKey.ItemRawID), g.panel)
	}
}

func TestPackDoorKeyFailsClosedBelowRequiredTier(t *testing.T) {
	g := teidonLockedDoorGame(t)
	magicKey, ok := g.pack.ItemUseEffectByRawID(0x56)
	if !ok || magicKey.DoorKeyTier != 2 {
		t.Fatalf("magic-key pack effect mismatch: %+v", magicKey)
	}
	g.inventory = []int{magicKey.ItemRawID}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if tier := g.cur.doorTier(17, 4); tier != 3 || !g.hasItem(magicKey.ItemRawID) {
		t.Fatalf("tier2 key must not open tier3 door: tier=%d held=%v",
			tier, g.hasItem(magicKey.ItemRawID))
	}
}
