package dq3data

import "testing"

// DarkenPalette:白天=原色(identity);黑夜 bg slot 明顯變暗;sprite bank(16..31)較淺(仍可見)。
// 對齊 dq3_scene.c dn_scale/dn_tint(bg 42/44/70、sprite 取「淺一半」)。
func TestDarkenPalette(t *testing.T) {
	base := make([]Color, 32)
	for i := range base {
		base[i] = Color{R: 200, G: 200, B: 200}
	}

	// 白天(phase 0):identity
	day := DarkenPalette(base, 0)
	for i := range day {
		if day[i] != base[i] {
			t.Fatalf("白天 slot %d 應為原色,得 %+v", i, day[i])
		}
	}

	// 黑夜(phase 2):bg slot 全調暗(明顯 < 原色)
	night := DarkenPalette(base, 2)
	bg := night[5]
	if bg.R >= 200 || bg.G >= 200 || bg.B >= 200 {
		t.Fatalf("黑夜 bg slot 未變暗:%+v", bg)
	}
	// bg 用 42/44/70% → R=84 G=88 B=140
	if bg.R != 84 || bg.G != 88 || bg.B != 140 {
		t.Fatalf("黑夜 bg slot 縮放不符(預期 84/88/140):%+v", bg)
	}

	// sprite bank(slot 16..31)較 bg 淺(角色夜間仍可見):同色下 sprite 亮度 > bg 亮度
	spr := night[20]
	if spr.R <= bg.R || spr.G <= bg.G {
		t.Fatalf("sprite bank slot 應較 bg 淺(仍可見):sprite=%+v bg=%+v", spr, bg)
	}
	// sprite 用 (rs+100)/2 → R=(42+100)/2=71% → 142;確認非全黑
	if spr.R == 0 && spr.G == 0 && spr.B == 0 {
		t.Fatal("sprite bank slot 全黑(角色夜間不可見)")
	}

	// 不變性:輸入 base 未被就地修改(可反覆套用不累積)
	for i := range base {
		if base[i] != (Color{R: 200, G: 200, B: 200}) {
			t.Fatalf("DarkenPalette 竄改輸入 base slot %d", i)
		}
	}
	t.Logf("白天=原色、黑夜 bg=%+v、sprite=%+v(sprite 較淺仍可見)✓", bg, spr)
}

// parseNPCs 晝夜表:阿里阿罕(CTY00)section0 白天 24 隻、黑夜 6 隻(夜間多數 NPC 睡覺/隱藏)。
// 對齊 C dq3_scene_load_npcs 的 word[0]日表 / word[1]夜表切換。
func TestParseNPCsDayNight(t *testing.T) {
	cty := findAsset(t, "CTY00.DAT")
	day, err := OpenTown(cty, 0, false)
	if err != nil {
		t.Fatalf("OpenTown CTY00 sec0 白天: %v", err)
	}
	night, err := OpenTown(cty, 0, true)
	if err != nil {
		t.Fatalf("OpenTown CTY00 sec0 黑夜: %v", err)
	}
	if len(day.NPCs) != 24 {
		t.Fatalf("阿里阿罕 sec0 白天 NPC 預期 24,得 %d(日表 word0 解析錯?)", len(day.NPCs))
	}
	if len(night.NPCs) != 6 {
		t.Fatalf("阿里阿罕 sec0 黑夜 NPC 預期 6,得 %d(夜表 word1 未選用?)", len(night.NPCs))
	}
	if len(night.NPCs) >= len(day.NPCs) {
		t.Fatalf("黑夜 NPC(%d)未少於白天(%d)", len(night.NPCs), len(day.NPCs))
	}
	t.Logf("阿里阿罕 sec0:白天 %d 隻、黑夜 %d 隻 ✓", len(day.NPCs), len(night.NPCs))
}
