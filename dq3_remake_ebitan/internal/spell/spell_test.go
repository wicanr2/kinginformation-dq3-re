package spell

import "testing"

func TestHeroKnown(t *testing.T) {
	// lv1:未學任何咒
	if len(HeroKnown(1)) != 0 {
		t.Fatalf("lv1 應 0 咒,得 %v", HeroKnown(1))
	}
	// lv2:美拉(121)
	k2 := HeroKnown(2)
	if len(k2) != 1 || k2[0] != 121 {
		t.Fatalf("lv2 應 [121 美拉],得 %v", k2)
	}
	// lv4:美拉 + 荷依米(161)
	k4 := HeroKnown(4)
	if len(k4) != 2 || k4[1] != 161 {
		t.Fatalf("lv4 應含 荷依米161,得 %v", k4)
	}
	// lv10:+吉拉(124),但輔助咒(尼夫拉姆143/魯拉172)無 def → 不列
	found := false
	for _, r := range HeroKnown(10) {
		if r == 124 {
			found = true
		}
		if r == 143 || r == 172 {
			t.Fatalf("輔助咒 %d 不該出現在戰鬥可施放清單", r)
		}
	}
	if !found {
		t.Fatal("lv10 應含 吉拉124")
	}
}

func TestSpellDefAndCast(t *testing.T) {
	d, ok := GetDef(121)
	if !ok || d.MP != 2 || d.Base != 10 || d.Kind != Dmg {
		t.Fatalf("美拉121 應 mp2/base10/DMG,得 %+v", d)
	}
	h, ok := GetDef(161)
	if !ok || h.Kind != Heal || h.MP != 3 || h.Base != 30 {
		t.Fatalf("荷依米161 應 mp3/base30/HEAL,得 %+v", h)
	}
	// 公式 base/2 + rng(base/2):base10 roll0=5、roll255=5+5*255/256=9
	if CastValue(10, 0) != 5 || CastValue(10, 255) != 9 {
		t.Fatalf("CastValue(10,*) 應 5..9,得 %d..%d", CastValue(10, 0), CastValue(10, 255))
	}
}
