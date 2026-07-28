package stats

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

func TestGrowthTarget(t *testing.T) {
	// 原版真正 MP 是 pair4：base8+slope2*43/2=51。
	if got := GrowthTarget(0, MP, 43); got != 51 {
		t.Fatalf("勇者原版 Lv43 MP target 應 51,得 %d", got)
	}
	if got := GrowthTarget(0, HP, 1); got != 16 {
		t.Fatalf("勇者 Lv1 HP target 應 16,得 %d", got)
	}
	if got := GrowthTarget(0, STR, 1); got != 9 {
		t.Fatalf("勇者 Lv1 STR target 應 9,得 %d", got)
	}
	if GrowthTarget(99, HP, 1) != 0 {
		t.Fatal("越界職業應回 0")
	}
}

func TestInitValuesMatchesSubED3C(t *testing.T) {
	r := rng.New(1)
	v := InitValues(0, r)
	if v[MP] != 9 {
		t.Fatalf("MP delta=1 經 sub_fa57 必為 1，Lv1 MP 應固定 9，得 %d", v[MP])
	}
	if v[STR] < 7 || v[STR] > 8 || v[VIT] < 4 || v[VIT] > 4 ||
		v[AGI] < 4 || v[AGI] > 4 || v[HP] < 9 || v[HP] > 15 {
		t.Fatalf("Lv1 原版範圍錯：%v", v)
	}
}

func TestLevelForExp(t *testing.T) {
	// thresh[0][10]=4364 → 達 lv10
	if got := LevelForExp(0, 4364); got != 10 {
		t.Fatalf("勇者 exp4364 應 lv10,得 %d", got)
	}
	// exp0 → lv1
	if got := LevelForExp(0, 0); got != 1 {
		t.Fatalf("exp0 應 lv1,得 %d", got)
	}
	// 極大 exp → clamp lv43(不越界)
	if got := LevelForExp(0, 1<<30); got != MaxLevel {
		t.Fatalf("極大 exp 應夾 lv43,得 %d", got)
	}
	// 差 1 未達門檻 → 不升(thresh[0][11]=6218)
	if got := LevelForExp(0, 6217); got != 10 {
		t.Fatalf("exp6217(<6218)應 lv10,得 %d", got)
	}
}
