package stats

import "testing"

func TestGrowthTarget(t *testing.T) {
	// #4 修正:勇者 Lv43 MaxMP = base8 + slope10*43/2 = 8+215 = 223(>=200,可放全體補血咒)
	if got := GrowthTarget(0, MP, 43); got != 223 {
		t.Fatalf("勇者 Lv43 MaxMP 應 223(#4 修正),得 %d", got)
	}
	// 勇者 Lv1 HP = base6 + slope6*1/2 = 6+3 = 9
	if got := GrowthTarget(0, HP, 1); got != 9 {
		t.Fatalf("勇者 Lv1 HP 應 9,得 %d", got)
	}
	// 勇者 Lv1 STR = base8 + slope16*1/2 = 8+8 = 16
	if got := GrowthTarget(0, STR, 1); got != 16 {
		t.Fatalf("勇者 Lv1 STR 應 16,得 %d", got)
	}
	if GrowthTarget(99, HP, 1) != 0 {
		t.Fatal("越界職業應回 0")
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
