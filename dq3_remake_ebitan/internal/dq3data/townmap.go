package dq3data

import "fmt"

// Town 是一張城鎮版面(CTY 檔的一個 section)。移植自 C dq3_town_load 的 layout 解析。
// CTY 檔用 CTYnn.DAT;阿里阿罕(起始城)= CTY00.DAT / section 0 / blk_n=1(DQ31.BLK+BLKBM1.DAT)。
type Town struct {
	W, H           int
	Cells          []byte   // 每格 tile 索引(u16 低 byte;高 byte=事件 subid,渲染不用)
	SpawnX, SpawnY int      // 版面 spawn(section header +0x13/+0x14)
	DlgBank        int      // 對話 bank(section header +0x17 → D3TXT0<bank>.TXT)
	HiMap          []byte   // 每格高 byte(低 5 bit = 事件/轉場 subid)
	Events         [][3]int // section 事件表(section+8):{type, param, p2}
	Transitions    [][4]int // section 轉場表(section+0xc):{destCty, destSec, x, y},by subid
	NPCs           []NPC
}

// NPC 是城鎮裡的一個角色。移植自 dq3_scene_load_npcs 的 7-byte record。
// B2 = sprite key(DQ3MAN.BLS entry_base=(B2-4)*4)。B2<4:原版 sub_0xffc3 對 NPC 一律
// (key-4)*0xf00+6 seek DQ3MAN.BLS,B2<4 時 u16 underflow → seek 超出檔案 240MB+ →
// 讀 0 byte → 頁緩衝殘留舊資料(undefined,非 DQ3LIN.BLS,詳見 docs/68)。
type NPC struct {
	X, Y  int
	B2    int // sprite key
	Ctrl  int // 行為/朝向控制;(Ctrl>>3)&7 = 互動子型(0/1=對話、2=劇本、3-7=設施)
	B4    int // 子型 0/1 時 = 對話記錄號;設施時 = 設施索引
	Flags int // bit3+ = 晝夜可見旗標等
}

func townU16(d []byte, o int) uint16 {
	if o < 0 || o+1 >= len(d) {
		return 0
	}
	return uint16(d[o]) | uint16(d[o+1])<<8
}

// OpenTown 解 CTY 檔的某 section → 城鎮版面。移植 dq3_town_load:
// section 偏移表無 count 前綴,word[i]=section i 的 base(0xffff=空);
// layout 指標在 section+0x0e,layout={w:u16, h:u16, tiles…(緊接 +4,stride 2)}。
// night=true 時 NPC 用黑夜表(section word[1]),對齊 C dq3_scene_load_npcs 的晝夜切換。
func OpenTown(cty []byte, section int, night bool) (*Town, error) {
	if len(cty) < 2 {
		return nil, fmt.Errorf("CTY too small")
	}
	if section < 0 {
		return nil, fmt.Errorf("section out of range")
	}
	so := int(townU16(cty, 2*section))
	if so == 0xffff || so+0x16 > len(cty) {
		return nil, fmt.Errorf("section empty/oob")
	}
	lay := so + int(townU16(cty, so+0x0e)) // layout_ptr 相對 section
	if lay+4 > len(cty) {
		return nil, fmt.Errorf("layout oob")
	}
	w := int(townU16(cty, lay))
	h := int(townU16(cty, lay+2))
	tbase := lay + 4
	if w <= 0 || h <= 0 || tbase >= len(cty) {
		return nil, fmt.Errorf("tile array oob (w=%d h=%d)", w, h)
	}
	dlgBank := 1
	if so+0x17 < len(cty) {
		dlgBank = int(cty[so+0x17]) // 對話 bank(section header +0x17,docs/42)
	}
	t := &Town{
		W: w, H: h,
		Cells:   make([]byte, w*h),
		SpawnX:  int(cty[so+0x13]), // spawn_x(section header)
		SpawnY:  int(cty[so+0x14]), // spawn_y
		DlgBank: dlgBank,
	}
	t.HiMap = make([]byte, w*h)
	for i := 0; i < w*h; i++ {
		v := townU16(cty, tbase+2*i)
		t.Cells[i] = byte(v)      // 低 byte = BLK index
		t.HiMap[i] = byte(v >> 8) // 高 byte = 事件/轉場 subid
	}
	// section 事件表(section+8:count byte + 4-byte 項{type, param u16, p2})
	if evptr := int(townU16(cty, so+8)); evptr != 0xffff && so+evptr < len(cty) {
		et := so + evptr
		cnt := int(cty[et])
		if cnt > 32 {
			cnt = 32
		}
		for k := 0; k < cnt; k++ {
			o := et + 1 + k*4
			if o+4 > len(cty) {
				break
			}
			t.Events = append(t.Events, [3]int{int(cty[o]), int(townU16(cty, o+1)), int(cty[o+3])})
		}
	}
	// section 轉場表(section+0xc:4-byte 項{destCty, destSec, x, y},by subid,無 count 前綴)。
	// 讀到檔尾,或讀到 layout 前為止(layout 緊跟表後時)。移植 dq3_town.c。
	if tptr := int(townU16(cty, so+0x0c)); tptr != 0xffff && so+tptr < len(cty) {
		tt := so + tptr
		lim := len(cty)
		if lay > tt && lay < lim { // layout 緊跟在表後 → 表止於 layout
			lim = lay
		}
		for k := 0; k < 32; k++ {
			o := tt + k*4
			if o+4 > lim {
				break
			}
			t.Transitions = append(t.Transitions, [4]int{int(cty[o]), int(cty[o+1]), int(cty[o+2]), int(cty[o+3])})
		}
	}
	t.NPCs = parseNPCs(cty, so, w, h, night)
	return t, nil
}

// parseNPCs 解 section 的 NPC 表(移植 dq3_scene_load_npcs, dq3_scene.c:105-146)。
// section 開頭 word[0]=白天表(人多)、word[1]=黑夜表(多睡覺/隱藏,只剩少數)(相對 so);
// 首 byte=count,後接 count×7-byte record。night 選主表:白天→word0、黑夜→word1;
// 主表無效(0xffff/越界)才退另一份(部分 section word[1]=none,日夜共用同一表)。
func parseNPCs(cty []byte, so, w, h int, night bool) []NPC {
	np := -1
	for k := 0; k < 2; k++ { // k=0 主表(依 night 選);k=1 退另一份
		off := 0               // word0=白天
		if night != (k == 1) { // night XOR k:夜間主表=word1、fallback=word0
			off = 2
		}
		wv := int(townU16(cty, so+off))
		if wv != 0xffff && so+wv < len(cty) {
			np = wv
			break
		}
	}
	if np < 0 {
		return nil
	}
	base := so + np
	if base >= len(cty) {
		return nil
	}
	cnt := int(cty[base])
	var out []NPC
	for i := 0; i < cnt; i++ {
		r := base + 1 + i*7
		if r+7 > len(cty) {
			break
		}
		x, y := int(cty[r]), int(cty[r+1])
		if x >= w || y >= h { // 座標越界跳過(對齊 C)
			continue
		}
		out = append(out, NPC{X: x, Y: y, B2: int(cty[r+2]), Ctrl: int(cty[r+3]), B4: int(cty[r+4]), Flags: int(cty[r+5])})
	}
	return out
}

// Tile 回 (x,y) 的 tile 索引;越界回 0(對齊 C bounds-safe 讀)。
func (t *Town) Tile(x, y int) int {
	if x < 0 || y < 0 || x >= t.W || y >= t.H {
		return 0
	}
	return int(t.Cells[y*t.W+x])
}
