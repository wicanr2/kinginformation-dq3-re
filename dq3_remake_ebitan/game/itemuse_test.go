package game

import (
	"testing"

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
