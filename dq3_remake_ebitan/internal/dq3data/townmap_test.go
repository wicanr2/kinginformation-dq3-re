package dq3data

import "testing"

// 阿里阿罕(起始城)= CTY00.DAT / section 0 / DQ31.BLK。對拍 C dq3_town_load。
func TestOpenTownAliahan(t *testing.T) {
	cty := findAsset(t, "CTY00.DAT")
	town, err := OpenTown(cty, 0, false)
	if err != nil {
		t.Fatalf("OpenTown CTY00 sec0: %v", err)
	}
	t.Logf("阿里阿罕 section0:%d×%d、spawn (%d,%d)", town.W, town.H, town.SpawnX, town.SpawnY)
	if town.MapFlags != 1 {
		t.Fatalf("CTY00 sec0 header+0x10 應允許魯拉(bit0)：got %#x", town.MapFlags)
	}

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

	// NPC:阿里阿罕(起始城)應有一批 NPC(C 版 boot 驗過 24 隻)。
	if len(town.NPCs) == 0 {
		t.Fatal("阿里阿罕 section0 解出 0 NPC(疑 NPC 表解析錯)")
	}
	spr := 0
	for _, n := range town.NPCs {
		if n.X < 0 || n.X >= town.W || n.Y < 0 || n.Y >= town.H {
			t.Fatalf("NPC 座標 (%d,%d) 出界", n.X, n.Y)
		}
		if n.B2 >= 4 { // 有對應 DQ3MAN.BLS sprite
			spr++
		}
	}
	t.Logf("阿里阿罕 NPC=%d(其中 %d 有 sprite key b2>=4)✓", len(town.NPCs), spr)
}

// 多 section:阿里阿罕 CTY00 有多個 section(section0 外圍 + 室內/寶物層)。
// 驗轉場表(section+0xc)解得出、dest 合理,且 section>0 也載得起來(內容廣度前提)。
func TestOpenTownMultiSection(t *testing.T) {
	cty := findAsset(t, "CTY00.DAT")
	loaded, withTrans, totalTrans := 0, 0, 0
	for sec := 0; sec < 16; sec++ {
		town, err := OpenTown(cty, sec, false)
		if err != nil {
			continue // 空 section(0xffff)→ 跳過
		}
		loaded++
		if town.W < 1 || town.H < 1 || len(town.Cells) != town.W*town.H {
			t.Fatalf("section %d 版面壞:%d×%d cells=%d", sec, town.W, town.H, len(town.Cells))
		}
		if len(town.Transitions) > 0 {
			withTrans++
			totalTrans += len(town.Transitions)
			for i, tr := range town.Transitions {
				// dest_sec 0xff/0xfe = 出城;其餘 < 16;dest_cty < 100 或同城。座標 byte 範圍內。
				if tr[1] > 0xff || tr[2] > 255 || tr[3] > 255 {
					t.Fatalf("section %d 轉場 %d 數值異常:%v", sec, i, tr)
				}
			}
		}
	}
	if loaded < 2 {
		t.Fatalf("CTY00 只載到 %d 個 section(預期多 section)", loaded)
	}
	if withTrans == 0 {
		t.Fatal("CTY00 所有 section 都沒有轉場表(疑 section+0xc 解析錯 → 無法跨 section)")
	}
	t.Logf("阿里阿罕:載到 %d section、%d 個 section 有轉場、共 %d 筆轉場 ✓", loaded, withTrans, totalTrans)
}
