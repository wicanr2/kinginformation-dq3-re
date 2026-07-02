package dq3data

import "testing"

// 對拍 C test_monster.c / docs/16 的數值。
func TestMonsters(t *testing.T) {
	raw := findAsset(t, "D3MNS.DAT")
	m, err := OpenMonsters(raw)
	if err != nil {
		t.Fatalf("OpenMonsters: %v", err)
	}
	// id5 史萊姆:HP6 / exp4 / gold2 / HP_max=base+rand=9
	s5, ok := m.Stat(5)
	if !ok {
		t.Fatal("取 id5 失敗")
	}
	if s5.HPBase != 6 || s5.Exp != 4 || s5.Gold != 2 {
		t.Fatalf("史萊姆 應 HP6/exp4/gold2,得 HP%d/exp%d/gold%d", s5.HPBase, s5.Exp, s5.Gold)
	}
	if m.HPMax(5) != 9 {
		t.Fatalf("史萊姆 HP_max 應 9,得 %d", m.HPMax(5))
	}
	// id2 金屬史萊姆:HP3 / exp4140
	s2, _ := m.Stat(2)
	if s2.HPBase != 3 || s2.Exp != 4140 {
		t.Fatalf("金屬史萊姆 應 HP3/exp4140,得 HP%d/exp%d", s2.HPBase, s2.Exp)
	}
	// 越界防護
	if _, ok := m.Stat(MonsterCount); ok {
		t.Fatal("越界 id 應回 false")
	}
	// 全 130 隻可解、AI 可取
	for id := 0; id < MonsterCount; id++ {
		if _, ok := m.Stat(id); !ok {
			t.Fatalf("id%d Stat 失敗", id)
		}
		if _, ok := m.AI(id); !ok {
			t.Fatalf("id%d AI 失敗", id)
		}
	}
	t.Logf("130 隻怪解析 ✓;史萊姆 HP6/exp4/gold2/max9、金屬 HP3/exp4140 對拍 C")
}
