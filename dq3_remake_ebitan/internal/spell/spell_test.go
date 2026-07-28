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

func TestKnownAllKeepsFieldSpells(t *testing.T) {
	all := KnownAll(0, 14)
	for _, want := range []int{172, 173} {
		found := false
		for _, rec := range all {
			found = found || rec == want
		}
		if !found {
			t.Fatalf("勇者 Lv14 全咒文清單應保留野外咒 rec%d，得 %v", want, all)
		}
	}
}

func TestPalpunteIsBattleOnlySpell(t *testing.T) {
	d, ok := GetDef(180)
	if !ok || d.Kind != Palpunte || d.MP != 20 {
		t.Fatalf("rec180 應為 20MP Palpunte battle def，得 %+v ok=%v", d, ok)
	}
	known := Known(4, 40)
	found := false
	for _, rec := range known {
		if rec == 180 {
			found = true
		}
	}
	if !found {
		t.Fatal("Lv40 魔法使的戰鬥咒文清單應包含 rec180")
	}
}

func TestOriginalMPCostTable(t *testing.T) {
	cases := map[int]int{
		123: 12, // 美拉宙瑪：舊 remake 誤用 10
		151: 6,  // 拜基魯多：舊 remake 誤用 8
		165: 62, // 比荷瑪順：精訊版特有高 cost
		172: 8, 173: 8, 176: 4, 177: 12,
	}
	for rec, want := range cases {
		if got := MPCost(rec); got != want {
			t.Errorf("rec%d MP=%d want EXE %d", rec, got, want)
		}
		if d, ok := GetDef(rec); ok && d.MP != want {
			t.Errorf("rec%d Def.MP=%d want EXE %d", rec, d.MP, want)
		}
	}
	if MPCost(120) != -1 || MPCost(181) != -1 {
		t.Fatal("咒文表範圍外應回 -1")
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

func TestOriginalDescriptorTargetFlags(t *testing.T) {
	cases := map[int]Target{
		121: TargetEnemyOne,   // flags 0x0d
		124: TargetEnemyGroup, // flags 0x1d
		144: TargetEnemyGroup, // 拉里荷，不是單體
		152: TargetEnemyOne,   // 美達巴尼
		151: TargetAllyOne,    // 拜基魯多
		154: TargetAllyOne,    // 史卡拉
		155: TargetAllyGroup,  // 史克魯多
		161: TargetAllyOne,    // 荷依米
		165: TargetAllyGroup,  // 比荷瑪順
		168: TargetAllyGroup,  // 薩梅哈
		180: TargetNone,
	}
	for rec, want := range cases {
		d, ok := GetDef(rec)
		if !ok || d.Target != want {
			t.Errorf("rec%d target=%d ok=%v, want %d", rec, d.Target, ok, want)
		}
	}
}

// TestW3StatusSpellDefs:W3(docs/72 A1)新增的輔助/狀態咒 descriptor(Kind/MP)。
func TestW3StatusSpellDefs(t *testing.T) {
	cases := []struct {
		rec  int
		kind Kind
	}{
		{144, Sleep}, {152, Sleep},
		{151, BuffAtk},
		{154, BuffDef}, {155, BuffDef},
		{156, Seal},
		{158, Blind},
		{166, CurePoison},
		{167, CureStatus}, {168, CureStatus},
	}
	for _, c := range cases {
		d, ok := GetDef(c.rec)
		if !ok {
			t.Errorf("rec%d 應有 descriptor,得 not found", c.rec)
			continue
		}
		if d.Kind != c.kind {
			t.Errorf("rec%d Kind 應為 %d,得 %d", c.rec, c.kind, d.Kind)
		}
		if d.MP <= 0 {
			t.Errorf("rec%d MP 應 >0,得 %d", c.rec, d.MP)
		}
	}
}

// TestKnownIncludesW3StatusSpells:玩家咒文選單應能選到 W3 新增的狀態咒(依習得表等級門檻)。
// 修正前 defs 無這些 rec → Known() 過濾掉,選單看不到;修正後應出現。
func TestKnownIncludesW3StatusSpells(t *testing.T) {
	// 僧侶系(class3):158@7、144@9、166@11、156@13、167@15、168@16(priestSchool)。
	priest := Known(3, 20)
	for _, rec := range []int{144, 156, 158, 166, 167, 168} {
		found := false
		for _, r := range priest {
			if r == rec {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("僧侶 lv20 應能選到 rec%d,得清單 %v", rec, priest)
		}
	}
	// 魔法系(class4):154@4、155@9、151@21(mageSchool)。
	mage := Known(4, 25)
	for _, rec := range []int{154, 155, 151} {
		found := false
		for _, r := range mage {
			if r == rec {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("魔法使 lv25 應能選到 rec%d,得清單 %v", rec, mage)
		}
	}
}
