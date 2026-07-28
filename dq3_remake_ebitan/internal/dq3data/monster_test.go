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
	// id89 怪力魔:D3MNS +0x26=0x62 變身杖、+0x27=200、+0x28=11。
	s89, _ := m.Stat(89)
	if s89.DropItem != 0x62 || s89.DropRate != 200 || s89.SpawnWeight != 11 {
		t.Fatalf("怪力魔掉落欄應 item62/rate200/weight11,得 %02x/%d/%d",
			s89.DropItem, s89.DropRate, s89.SpawnWeight)
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

func TestMonsterSprite(t *testing.T) {
	shp := findAsset(t, "DQ3MNS.SHP")
	// id5 史萊姆:尺寸合理且有不透明像素(對拍 test_monster)
	spr, err := DecodeMonsterSprite(shp, 5)
	if err != nil {
		t.Fatalf("史萊姆 sprite 解碼失敗: %v", err)
	}
	if spr.W <= 0 || spr.H <= 0 {
		t.Fatalf("史萊姆 sprite 尺寸 %d×%d", spr.W, spr.H)
	}
	any := false
	for r := 0; r < spr.H && !any; r++ {
		for b := 0; b < spr.W; b++ {
			if spr.Opaque[r][b] {
				any = true
				break
			}
		}
	}
	if !any {
		t.Fatal("史萊姆 sprite 全透明(解碼疑有誤)")
	}
	t.Logf("史萊姆 sprite %d×%d 有墨 ✓", spr.W, spr.H)
}
