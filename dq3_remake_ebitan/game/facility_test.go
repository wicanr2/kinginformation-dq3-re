package game

import "testing"

func TestInnChargesAndHealsLivingPartyOnly(t *testing.T) {
	living := newMember([]int{1}, 1, 0, 0)
	living.CurHP, living.CurMP = 1, 0
	dead := newMember([]int{2}, 3, 0, 0)
	dead.CurHP, dead.CurMP = 0, 0
	g := &Game{
		curCty:     0,
		heroGold:   20,
		heroHP:     1,
		heroMP:     0,
		heroInit:   true,
		companions: []*Member{living, dead},
	}

	// CTY00 facility 0 的單價是 2；原版只計仍存活的勇者與一名同伴。
	g.openFacility(0)

	_, heroMaxHP, _, _, _ := g.heroStats()
	if g.heroGold != 16 {
		t.Fatalf("旅店應收單價2×存活2人=4，gold=%d", g.heroGold)
	}
	if g.heroHP != heroMaxHP || g.heroMP != g.heroMaxMP() {
		t.Fatalf("勇者未補滿：HP=%d/%d MP=%d/%d",
			g.heroHP, heroMaxHP, g.heroMP, g.heroMaxMP())
	}
	if living.CurHP != living.MaxHP() || living.CurMP != living.MaxMP() {
		t.Fatalf("存活同伴未補滿：HP=%d/%d MP=%d/%d",
			living.CurHP, living.MaxHP(), living.CurMP, living.MaxMP())
	}
	if dead.CurHP != 0 || dead.CurMP != 0 {
		t.Fatalf("旅店不可復活死亡同伴：HP=%d MP=%d", dead.CurHP, dead.CurMP)
	}
}

func TestInnInsufficientGoldDoesNotPartiallyHeal(t *testing.T) {
	companion := newMember([]int{1}, 1, 0, 0)
	companion.CurHP, companion.CurMP = 1, 0
	g := &Game{
		curCty:     0,
		heroGold:   3,
		heroHP:     1,
		heroMP:     0,
		heroInit:   true,
		companions: []*Member{companion},
	}

	g.openFacility(0) // 需要 2×2=4。

	if g.heroGold != 3 || g.heroHP != 1 || g.heroMP != 0 ||
		companion.CurHP != 1 || companion.CurMP != 0 {
		t.Fatalf("金幣不足不可扣款或部分恢復：gold=%d hero=%d/%d companion=%d/%d",
			g.heroGold, g.heroHP, g.heroMP, companion.CurHP, companion.CurMP)
	}
}
