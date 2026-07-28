package itemuse

import "testing"

// 對拍 dq3_item_use_kind:每個已知消耗品 id → 正確種類。
func TestKindOf(t *testing.T) {
	cases := []struct {
		id   int
		want Kind
	}{
		{0x41, HealHP}, {0x42, CurePoison}, {0x43, ReturnTown}, {0x44, Repel},
		{0x45, CureParalysis}, {0x48, PrayerRing}, {0x5f, DarkLamp}, {0x61, Mirror},
		{0x5a, Awaken}, {0x0f, Gaia}, {0x5e, Drain}, {0x77, FairyFlute}, {0x75, Rainbow},
		{0x01, None}, {0x99, None}, // 裝備/未知 → 不可用
	}
	for _, c := range cases {
		if got := KindOf(c.id); got != c.want {
			t.Errorf("KindOf(0x%02x)=%d, want %d", c.id, got, c.want)
		}
	}
}

// 藥草治療量:未滿補 30、封頂於 max、陣亡/已滿回 0。對拍 dq3_item_use_heal。
func TestHealHPAmount(t *testing.T) {
	if got := HealHPAmount(10, 100); got != 30 {
		t.Errorf("HP10/100 補 %d, want 30", got)
	}
	if got := HealHPAmount(90, 100); got != 10 { // 封頂
		t.Errorf("HP90/100 補 %d, want 10(封頂)", got)
	}
	if got := HealHPAmount(100, 100); got != 0 { // 已滿
		t.Errorf("HP 已滿補 %d, want 0", got)
	}
	if got := HealHPAmount(0, 100); got != 0 { // 陣亡不可用
		t.Errorf("陣亡補 %d, want 0", got)
	}
}

// 祈禱之戒 MP 回復:未滿補 30、封頂、已滿回 0。對拍 dq3_item_use_prayer_mp。
func TestPrayerMPAmount(t *testing.T) {
	if got := PrayerMPAmount(5, 50); got != 30 {
		t.Errorf("MP5/50 補 %d, want 30", got)
	}
	if got := PrayerMPAmount(40, 50); got != 10 {
		t.Errorf("MP40/50 補 %d, want 10(封頂)", got)
	}
	if got := PrayerMPAmount(50, 50); got != 0 {
		t.Errorf("MP 已滿補 %d, want 0", got)
	}
}

// 消耗語意:立即效果消耗;位置相關/武器道具/祈禱之戒不在此消耗。
func TestConsumable(t *testing.T) {
	for _, k := range []Kind{HealHP, ReturnTown, Repel, CurePoison} {
		if !k.Consumable() {
			t.Errorf("種類 %d 應消耗", k)
		}
	}
	for _, k := range []Kind{PrayerRing, Gaia, Drain, FairyFlute, Mirror, None} {
		if k.Consumable() {
			t.Errorf("種類 %d 不應在此消耗", k)
		}
	}
}
