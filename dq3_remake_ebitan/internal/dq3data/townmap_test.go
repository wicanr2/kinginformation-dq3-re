package dq3data

import "testing"

// 阿里阿罕(起始城)= CTY00.DAT / section 0 / DQ31.BLK。對拍 C dq3_town_load。
func TestOpenTownAliahan(t *testing.T) {
	cty := findAsset(t, "CTY00.DAT")
	town, err := OpenTown(cty, 0)
	if err != nil {
		t.Fatalf("OpenTown CTY00 sec0: %v", err)
	}
	t.Logf("阿里阿罕 section0:%d×%d、spawn (%d,%d)", town.W, town.H, town.SpawnX, town.SpawnY)

	// 版面尺寸合理(城鎮小圖,通常數十格見方)
	if town.W < 4 || town.H < 4 || town.W > 256 || town.H > 256 {
		t.Fatalf("版面尺寸不合理:%d×%d", town.W, town.H)
	}
	if len(town.Cells) != town.W*town.H {
		t.Fatalf("cells 數 %d != w*h %d", len(town.Cells), town.W*town.H)
	}
	// spawn 落在版面內
	if town.SpawnX < 0 || town.SpawnX >= town.W || town.SpawnY < 0 || town.SpawnY >= town.H {
		t.Fatalf("spawn (%d,%d) 出界(版面 %d×%d)", town.SpawnX, town.SpawnY, town.W, town.H)
	}
	// tile 索引須落在 DQ31.BLK 範圍內
	blk, err := OpenBLK(findAsset(t, "DQ31.BLK"))
	if err != nil {
		t.Fatalf("OpenBLK DQ31: %v", err)
	}
	for i, c := range town.Cells {
		if int(c) >= blk.Count {
			t.Fatalf("cell %d tile %d 超出 DQ31.BLK count %d", i, c, blk.Count)
		}
	}
	// 版面非全同一格(有牆/地/門等多種 tile)
	first := town.Cells[0]
	varied := false
	for _, c := range town.Cells {
		if c != first {
			varied = true
			break
		}
	}
	if !varied {
		t.Fatal("版面全同一 tile(疑 layout 解析錯)")
	}
	t.Logf("cells 全在 DQ31.BLK(%d tiles)範圍、tile 多樣 ✓", blk.Count)
}
