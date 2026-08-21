package dq3data

import "testing"

func TestSelectPaletteBank(t *testing.T) {
	base := make([]Color, 80)
	for i := range base {
		base[i] = Color{R: uint8(i)}
	}
	bank, ok := SelectPaletteBank(base, 4, 16)
	if !ok || len(bank) != 16 || bank[0].R != 64 || bank[15].R != 79 {
		t.Fatalf("bank4=%+v ok=%v", bank, ok)
	}
	bank[0].R = 0
	if base[64].R != 64 {
		t.Fatal("SelectPaletteBank 不得回傳會修改原始 palette 的 alias")
	}
	for _, tc := range [][2]int{{-1, 16}, {5, 16}, {0, 0}} {
		if _, ok := SelectPaletteBank(base, tc[0], tc[1]); ok {
			t.Fatalf("越界 bank/entries=%v 不得成功", tc)
		}
	}
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
