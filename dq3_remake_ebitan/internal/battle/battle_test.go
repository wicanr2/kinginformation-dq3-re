package battle

import "testing"

// 對拍 C test_battle.c 的斷言值。
func TestPhysDamage(t *testing.T) {
	cases := []struct {
		atk, def, roll, crit, want int
		msg                        string
	}{
		{40, 20, 0, 0, 10, "一般 roll0 = (atk−def)/2 = 10"},
		{40, 20, 255, 0, 14, "一般 roll255 = 10 + (atk−def)/4 變異 = 14"},
		{40, 20, 0, 1, 20, "會心一擊 = 一般 ×2 = 20"},
		{10, 80, 0, 0, 1, "弱攻 roll0 → 1 點"},
		{10, 80, 255, 0, 0, "弱攻 roll255 → miss(0)"},
	}
	for _, c := range cases {
		if got := PhysDamage(c.atk, c.def, c.roll, c.crit); got != c.want {
			t.Errorf("%s:PhysDamage(%d,%d,%d,%d)=%d,want %d", c.msg, c.atk, c.def, c.roll, c.crit, got, c.want)
		}
	}
}

func TestIgnoreDefensePhysicalDamage(t *testing.T) {
	for _, tc := range []struct {
		atk, roll, want int
	}{
		{44, 0, 22},
		{44, 255, 32},
		{1, 255, 0},
		{150, 128, 93},
	} {
		if got := IgnoreDefensePhysicalDamage(tc.atk, tc.roll); got != tc.want {
			t.Fatalf("IgnoreDefensePhysicalDamage(%d,%d)=%d，預期 %d", tc.atk, tc.roll, got, tc.want)
		}
	}
}

func TestApplyDamage(t *testing.T) {
	if ApplyDamage(30, 50) != 0 {
		t.Error("HP30 受50傷 應 0(不下溢)")
	}
	if ApplyDamage(100, 30) != 70 {
		t.Error("HP100 受30傷 應 70")
	}
}

func TestFleeOK(t *testing.T) {
	if !FleeOK(50, 0, 0) {
		t.Error("高敏捷低抗性 roll0 應逃成功")
	}
	if FleeOK(0, 100, 250) {
		t.Error("低敏捷高抗性 roll250 應逃失敗")
	}
}

func TestResolve(t *testing.T) {
	party := []uint16{50, 50, 50, 50}
	// 敵 HP0 → 勝
	if Resolve(party, nil, 4, []uint16{0}, 1) != Victory {
		t.Error("打死敵人(敵HP0)應勝")
	}
	// 全隊 HP0 → 敗
	if Resolve([]uint16{0, 0, 0, 0}, nil, 4, []uint16{10}, 1) != Defeat {
		t.Error("全隊HP0 應敗")
	}
	// 雙方存活 → 續戰
	if Resolve(party, nil, 4, []uint16{10, 10}, 2) != Ongoing {
		t.Error("雙方都有存活 應續戰")
	}
	// #1:全隊被吹飛(HP>0 但 removed)→ 修正判敗、原版誤判勝
	removed := []bool{true, true, true, true}
	if Resolve(party, removed, 4, []uint16{10}, 1) != Defeat {
		t.Error("修正:全隊被吹飛 應判敗")
	}
	if ResolveBuggy(party, removed, 4, []uint16{10}, 1) != Victory {
		t.Error("原版:全隊被吹飛 應誤判勝(重現 #1)")
	}
}
