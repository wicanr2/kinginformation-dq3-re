package game

import (
	"io/fs"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// overworld (x,y,map) → CTY 號(=index);BLK 號 per CTY。自 dq3_exedata dq3x_cty_loc / dq3x_map_blknum baked。
var ctyLoc = [100][3]int{
	{153, 174, 0},
	{141, 149, 0},
	{47, 66, 0},
	{58, 27, 0},
	{49, 12, 0},
	{31, 16, 0},
	{75, 85, 0},
	{141, 175, 0},
	{146, 171, 0},
	{168, 164, 0},
	{28, 45, 0},
	{30, 20, 0},
	{36, 105, 0},
	{32, 91, 0},
	{106, 114, 0},
	{101, 119, 0},
	{26, 72, 0},
	{107, 101, 0},
	{113, 88, 0},
	{142, 115, 0},
	{36, 145, 0},
	{140, 116, 0},
	{142, 79, 0},
	{78, 166, 0},
	{201, 135, 0},
	{153, 174, 0},
	{182, 80, 0},
	{197, 164, 0},
	{148, 156, 0},
	{166, 169, 0},
	{168, 164, 0},
	{45, 70, 0},
	{36, 105, 0},
	{50, 110, 0},
	{38, 53, 0},
	{36, 53, 0},
	{0, 0, 255},
	{26, 72, 0},
	{190, 61, 0},
	{20, 32, 0},
	{146, 52, 0},
	{186, 37, 0},
	{213, 123, 0},
	{190, 123, 0},
	{190, 123, 0},
	{136, 110, 0},
	{26, 121, 0},
	{83, 165, 0},
	{183, 189, 0},
	{107, 101, 0},
	{27, 78, 0},
	{78, 57, 0},
	{109, 39, 0},
	{210, 64, 0},
	{174, 20, 0},
	{66, 54, 0},
	{43, 140, 0},
	{43, 138, 0},
	{210, 64, 0},
	{210, 64, 0},
	{210, 64, 0},
	{210, 64, 0},
	{81, 81, 0},
	{84, 81, 0},
	{51, 133, 0},
	{53, 130, 0},
	{43, 130, 0},
	{89, 36, 0},
	{127, 125, 0},
	{126, 128, 0},
	{48, 188, 0},
	{54, 129, 0},
	{54, 129, 0},
	{36, 105, 0},
	{0, 0, 255},
	{82, 165, 0},
	{20, 32, 0},
	{84, 67, 1},
	{64, 26, 1},
	{106, 66, 1},
	{106, 66, 1},
	{166, 34, 1},
	{142, 25, 1},
	{210, 64, 0},
	{87, 113, 1},
	{135, 126, 1},
	{163, 96, 1},
	{89, 36, 1},
	{91, 81, 1},
	{166, 68, 1},
	{110, 73, 1},
	{110, 73, 1},
	{139, 136, 1},
	{170, 133, 1},
	{0, 0, 255},
	{0, 0, 255},
	{0, 0, 255},
	{0, 0, 255},
	{0, 0, 255},
	{0, 0, 255},
}

var mapBlkNum = [100]int{1, 1, 1, 1, 1, 1, 1, 2, 4, 1, 4, 2, 1, 5, 5, 1, 1, 4, 4, 2, 1, 3, 1, 4, 2, 1, 4, 3, 1, 1, 5, 1, 1, 1, 1, 1, 3, 1, 3, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2, 2, 1, 1, 1, 1, 5, 5, 1, 4, 1, 1, 1, 1, 3, 5, 5, 3, 1, 1, 3, 3, 1, 1, 1, 1, 4, 1, 1, 1, 1, 2, 2, 2, 4, 4, 1, 1, 2, 2, 1, 2, 2, 2}

// findCtyAt:地面 (px,py) → CTY 號;無回 -1(向後相容,= layer 0)。
func findCtyAt(px, py int) int { return findCtyAtLayer(px, py, 0) }

// findCtyAtLayer:某地表層 (px,py) → CTY 號(移植 find_cty_at_map);無回 -1。
// 對齊 RE:只配對同層(ctyLoc[i][2]==layer,255=空槽)、py==loc.Y 且 (px==loc.X 或 loc.X+1)。
func findCtyAtLayer(px, py, layer int) int {
	for i := 0; i < len(ctyLoc); i++ {
		if ctyLoc[i][2] != layer { // 只配對同層 + 跳過空槽(0xff)
			continue
		}
		lx, ly := ctyLoc[i][0], ctyLoc[i][1]
		if py == ly && (px == lx || px == lx+1) {
			return i
		}
	}
	return -1
}

// loadTownScene:載入任一 CTY 城鎮(section 0)為 Scene(通用,移植 dq3_town_load 用法)。
// blkn 由 mapBlkNum[cty] 決定 → DQ3<blkn>.BLK + BLKBM<blkn>.DAT。含 NPC(sprite 自 DQ3MAN.BLS)。
// phase = 進城當下晝夜相位(0白天/1黃昏/2黑夜/3黎明):決定 NPC 日/夜表 + palette 調暗。
func loadTownScene(assets fs.FS, pal []dq3data.Color, manBLS []byte, cty, blkn, phase int, flagSet func(int) bool) (*Scene, error) {
	return loadTownSceneSec(assets, pal, manBLS, cty, blkn, 0, phase, flagSet)
}

// loadTownSceneSec 載入某城的指定 section(轉場用)。section 0 = 城鎮外圍入口層。
// pal 須為日中 base 色盤;函式內依 phase 調暗(不動輸入),NPC 依 phase==2 選日/夜表。
// flagSet(id)=某 story-flag 是否已設(通常 g.flags 查詢);nil = 不過濾(測試用)。
// 條件 NPC(NPCStoryGate)其 byte5 旗標未設時不載入(對齊原版 file 0x4560 的 je skip,docs/71)。
func loadTownSceneSec(assets fs.FS, pal []dq3data.Color, manBLS []byte, cty, blkn, section, phase int, flagSet func(int) bool) (*Scene, error) {
	if blkn < 1 || blkn > 5 {
		blkn = 1
	}
	rd := func(name string) []byte { d, _ := fs.ReadFile(assets, name); return d }
	ctyName := ctyFile(cty)
	blkName := blkFile(blkn)
	attrName := blkbmFile(blkn)

	night := (phase & 3) == 2 // 2=黑夜 → NPC 用夜表;白天/黃昏/黎明皆當白天(對齊 dq3_scene.c:117)
	tw, err := dq3data.OpenTown(rd(ctyName), section, night)
	if err != nil {
		return nil, err
	}
	blk, err := dq3data.OpenBLK(rd(blkName))
	if err != nil {
		return nil, err
	}
	sc := &Scene{
		blk: blk, attr: dq3data.OpenBlockAttr(rd(attrName)), pal: dq3data.DarkenPalette(pal, phase),
		w: tw.W, h: tw.H, tileAt: tw.Tile, spawnX: tw.SpawnX, spawnY: tw.SpawnY,
		hiMap: tw.HiMap, events: tw.Events, transitions: tw.Transitions, sec: section, night: night,
		hiTransition: cty == 72 && section == 0,
	}
	sc.npcRng.Seed(uint16(section*2654 + 1)) // 確定性種子(依 section;對齊 C dq3_scene.c npc_rng)
	// 該城對話 bank(D3TXT0<bank>.TXT,section +0x17;共用 D3TXT00.FON)
	bank := tw.DlgBank
	if bank < 1 || bank > 9 {
		bank = 1
	}
	sc.dlgText = dq3data.LoadText(rd("D3TXT00.FON"), rd("D3TXT0"+string(rune('0'+bank))+".TXT"))
	sprCache := map[int]*dq3data.CharSprite{}
	for _, n := range tw.NPCs {
		// NPC 可見性過濾(對齊原版 file 0x4560:test [byte5→0x4f70]; je skip):byte5=story-flag id,
		// 該旗標未設就不載入。新遊戲初值 flag 0-39 清、40-255 設(dq3data.NPCStoryInitFlags),
		// 故 byte5<40 的條件 NPC 起始隱藏、≥40 的基礎人口顯示。flagSet==nil 時不過濾(測試)。
		if flagSet != nil && !flagSet(n.Flags) {
			continue
		}
		// b2<4:原版 sub_0xffc3(file 0xffc3)對 NPC 一律走 bp=2 分支、無條件 `ax-=4`
		// 再 seek (ax)*0xf00+6 到 DQ3MAN.BLS——b2<4 時 ax-4 對 u16 underflow,
		// seek 落在檔案(222726B)外 240MB+,INT21h/AH=3Fh 讀 0 byte,頁緩衝維持
		// stale 內容(未定義)。反組譯已證實**不是**走 DQ3LIN.BLS(該檔另有專用
		// loader,file 0x13387,固定 18-subframe 表,與這條 NPC path 無關,見
		// docs/68)。原版行為 = undefined,remake 不硬猜、不丟棄:同 b2<4 NPC
		// 共用 entry 0(DQ3MAN.BLS 第一個角色)當忠實但誠實標示的 fallback。
		b2 := n.B2
		if b2 < 4 {
			b2 = 4 // fallback → entry 0
		}
		spr, ok := sprCache[b2]
		if !ok {
			spr = dq3data.LoadCharSprite(manBLS, (b2-4)*4)
			sprCache[b2] = spr
		}
		sc.npcs = append(sc.npcs, npcInst{x: n.X, y: n.Y, ctrl: n.Ctrl, b4: n.B4, spr: spr})
	}
	return sc, nil
}

// CTYnn.DAT / DQ3n.BLK / BLKBMn.DAT 檔名(避免 fmt 依賴,手拼兩位數)。
func ctyFile(cty int) string { return "CTY" + two(cty) + ".DAT" }
func blkFile(n int) string   { return "DQ3" + string(rune('0'+n)) + ".BLK" }
func blkbmFile(n int) string { return "BLKBM" + string(rune('0'+n)) + ".DAT" }
func two(n int) string {
	if n < 10 {
		return "0" + string(rune('0'+n))
	}
	return string(rune('0'+n/10)) + string(rune('0'+n%10))
}
