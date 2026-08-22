package dq3data

import "fmt"

// 怪物數值/AI(D3MNS.DAT,130 隻 × 41 byte/筆)。移植自 C dq3_monster.c(docs/16)。
const (
	MonsterCount = 130
	monsterRec   = 41
	spellCount   = 60 // DGROUP 0x37c3 的 spell descriptor 筆數；原版 effect id 0..59
)

// MonsterStat 是一隻怪的數值(欄位偏移對齊 C）。
type MonsterStat struct {
	HPBase, HPRand uint16 // +0x00 / +0x02
	MP             uint16 // +0x04:戰鬥 MP；帕魯朋特效果會由敵人吸取此值
	Atk            uint8  // +0x08
	Def, Agi       uint16 // +0x09 / +0x0b
	FleeResist     uint8  // +0x18
	Exp, Gold      uint16 // +0x21 / +0x23
	DropRate       uint8  // +0x25；勝利結算以 rng(256) <= 此值，0xff 必掉
	DropItem       uint8  // +0x26
	Unknown27      uint8  // +0x27；尚未閉合，不得誤當掉落率
	SpawnWeight    uint8  // +0x28
}

// MonsterAI 保存一隻怪的原始 action gate／mask 與逃跑參數。
type MonsterAI struct {
	CastProb   uint8    // +0x0d:rng(256) < 此 → 選 action mask；歷史欄位名保留
	FleeThresh uint8    // +0x17:我方平均等級 ≥ 此 → 考慮逃跑
	FleeRate   uint8    // +0x18:觸發後 rng(256) ≤ 此就逃(0=不逃)
	SpellMask  [6]uint8 // +0x0e..+0x13:raw action mask；並非每個 bit 都是咒文
}

// Monsters 持有 D3MNS.DAT 整檔。
type Monsters struct{ raw []byte }

// OpenMonsters 解 D3MNS.DAT。移植 dq3_monsters_load。
func OpenMonsters(d []byte) (*Monsters, error) {
	if len(d) < MonsterCount*monsterRec {
		return nil, fmt.Errorf("D3MNS.DAT too small: %d < %d", len(d), MonsterCount*monsterRec)
	}
	return &Monsters{raw: d}, nil
}

func (m *Monsters) rec(id int) []byte { return m.raw[id*monsterRec:] }

// Stat 取怪 id 的數值。移植 dq3_monster_get_stat。
func (m *Monsters) Stat(id int) (MonsterStat, bool) {
	if id < 0 || id >= MonsterCount {
		return MonsterStat{}, false
	}
	r := m.rec(id)
	return MonsterStat{
		HPBase: le16(r, 0x00), HPRand: le16(r, 0x02),
		MP:  le16(r, 0x04),
		Atk: r[0x08], Def: le16(r, 0x09), Agi: le16(r, 0x0b),
		FleeResist: r[0x18], Exp: le16(r, 0x21), Gold: le16(r, 0x23),
		DropRate: r[0x25], DropItem: r[0x26], Unknown27: r[0x27], SpawnWeight: r[0x28],
	}, true
}

// SpellChance 回原版對指定怪物施 spellID 時的成功門檻(0/68/180/255)。
//
// DQ3.EXE file 0xeb6b 以 spellID 分到 packed 2-bit 類別，再從 D3MNS.DAT
// +0x19..+0x1d 取代碼；在目前 0..59 的 descriptor domain 中，Palpunte 是 spellID 59，
// 落在類別 17（類別 16 是原始 0 sentinel 形成的空洞）。這裡只回傳 raw threshold。
// 原版 caller 的邊界不同：單體
// sub_1D2D3 以 rng <= threshold 進效果，多體 sub_1D338/sub_1D3B7 以
// rng < threshold 進效果；呼叫端不可把此函式當成固定的單一比較器。
func (m *Monsters) SpellChance(id, spellID int) (int, bool) {
	if id < 0 || id >= MonsterCount || spellID < 0 || spellID >= spellCount {
		return 0, false
	}
	// DGROUP 0x36e5。原程式是 cmp AL,[table+BX]；相等時先 inc BX，
	// 所以邊界採「spellID >= threshold 就前進」，不能改成第一個 >= 的索引。
	// 最後四個 byte 與 DGROUP 0x36f5 class table 連續，0 是原版資料，不能刪掉。
	thresholds := [...]int{3, 6, 9, 13, 16, 18, 22, 23, 25, 27, 28, 29, 31, 35, 37, 38, 0, 68, 180, 255}
	category := 0
	for category < len(thresholds) && spellID >= thresholds[category] {
		category++
	}
	packed := m.rec(id)[0x19+category/4]
	// sub_eb6b 以 rol 2/4/6/8 後取低 2 bit，故類別依序放在
	// byte 的 bits 7..6、5..4、3..2、1..0。
	shift := uint((3 - category%4) * 2)
	code := (packed >> shift) & 3
	chances := [...]int{0, 68, 180, 255} // DGROUP 0x36f5
	return chances[code], true
}

// AI 取怪 id 的行為。移植 dq3_monster_get_ai。
func (m *Monsters) AI(id int) (MonsterAI, bool) {
	if id < 0 || id >= MonsterCount {
		return MonsterAI{}, false
	}
	r := m.rec(id)
	ai := MonsterAI{CastProb: r[0x0d], FleeThresh: r[0x17], FleeRate: r[0x18]}
	copy(ai.SpellMask[:], r[0x0e:0x14])
	return ai, true
}

// HPMax 回怪 id 的最大 HP(base+rand)。移植 dq3_monster_hp_max。
func (m *Monsters) HPMax(id int) uint16 {
	s, ok := m.Stat(id)
	if !ok {
		return 0
	}
	return s.HPBase + s.HPRand
}

// 怪物 sprite 尺寸上限(對齊 C DQ3_SHP_MAX*)。八頭大蛇 416×143、索瑪 384×144。
const (
	ShpMaxW = 416
	ShpMaxH = 160
)

// MonsterSprite 是一隻怪的圖（palette 索引對 MNSBK.PAL；透明度由獨立 AND-mask 決定）。
type MonsterSprite struct {
	W, H   int
	Px     [][]uint8 // [H][W] palette 索引
	Opaque [][]bool  // [H][W] true=畫
}

// DecodeMonsterSprite 解 DQ3MNS.SHP 第 id 隻。移植 dq3_monster_sprite_decode。
// SHP:131×u32 offset 表 + 每隻 {u16 w_bytes, u16 h, 4 plane plane-major(plane 3,2,1,0),
// 逐列 RLE AND-mask}。原始基線的 boss id128/129 仍可能為空；執行版 patched SHP
// 會直接提供兩格，只有空／無效時才嘗試用復原資料回退。
// 回 nil,error 表空 sprite 且無復原資料、或越界。
func DecodeMonsterSprite(shp []byte, id int) (*MonsterSprite, error) {
	if id < 0 || id >= MonsterCount {
		return nil, fmt.Errorf("id %d 越界", id)
	}
	if (id+1)*4+4 > len(shp) {
		return nil, fmt.Errorf("offset 表越界")
	}
	off := int(le32(shp, id*4))
	nxt := int(le32(shp, (id+1)*4))

	// 先嘗試 SHP 中的資料
	if off+4 <= len(shp) && nxt > off {
		// SHP 頭合法，試解碼；若失敗則回退
		spr, err := decodeSpriteBytesInternal(shp, off, nxt, id, true)
		if err == nil {
			return spr, nil // 成功
		}
		// 解碼失敗，改用復原資料
	}

	// SHP 中無有效資料，嘗試用復原資料
	data, ok := restoredSprite[id]
	if !ok {
		return nil, fmt.Errorf("SHP 空或無效(id %d),且無復原資料", id)
	}
	return decodeSpriteBytesInternal(data, 0, len(data), id, false)
}

// decodeSpriteBytesInternal 依據資料與 offset 解碼 sprite。
// 移植 C dq3_monster_sprite_decode，複用到 restored_sprites 回退路徑。
func decodeSpriteBytesInternal(data []byte, off, end, id int, requireMask bool) (*MonsterSprite, error) {
	if off+4 > len(data) {
		return nil, fmt.Errorf("sprite 頭越界(id %d)", id)
	}
	if end > len(data) || end < off+4 {
		return nil, fmt.Errorf("sprite 尾界越界(id %d)", id)
	}
	wb := int(le16(data, off) & 0x7fff)
	h := int(le16(data, off+2))
	w := wb * 8
	if w <= 0 || h <= 0 || w > ShpMaxW || h > ShpMaxH {
		return nil, fmt.Errorf("尺寸越界 %d×%d", w, h)
	}
	base := off + 4
	planeSz := wb * h
	if base+planeSz*4 > end {
		return nil, fmt.Errorf("sprite 資料越界(id %d)", id)
	}
	spr := &MonsterSprite{W: w, H: h, Px: make([][]uint8, h), Opaque: make([][]bool, h)}
	for r := 0; r < h; r++ {
		spr.Px[r] = make([]uint8, w)
		spr.Opaque[r] = make([]bool, w)
	}
	for s := 0; s < 4; s++ {
		pl := uint(3 - s) // plane 3 先(Map Mask 0x08→0x01)
		b0 := base + s*planeSz
		for r := 0; r < h; r++ {
			for b := 0; b < wb; b++ {
				v := data[b0+r*wb+b]
				for bit := 0; bit < 8; bit++ {
					if v&(0x80>>bit) != 0 {
						spr.Px[r][b*8+bit] |= 1 << pl
					}
				}
			}
		}
	}
	maskStart := base + planeSz*4
	if maskStart < end {
		if err := decodeMonsterMask(data, maskStart, end, wb, h, spr.Opaque); err != nil {
			return nil, fmt.Errorf("sprite mask(id %d): %w", id, err)
		}
	} else if requireMask {
		return nil, fmt.Errorf("sprite mask 缺失(id %d)", id)
	} else {
		// 舊版 embedded fallback 若沒有 mask，才以非零色作遮罩；現行產生檔的
		// 128/129 已附 RLE AND-mask，正常不會走此分支。
		for r := 0; r < h; r++ {
			for x := 0; x < w; x++ {
				spr.Opaque[r][x] = spr.Px[r][x] != 0
			}
		}
	}
	return spr, nil
}

// decodeMonsterMask 解原版 sub_b31a/sub_b37c 使用的逐列 RLE AND-mask。
// mask bit 1 保留背景（透明），bit 0 清除背景後再由四個色彩 plane 寫入（不透明）。
func decodeMonsterMask(data []byte, pos, end, rowBytes, rows int, opaque [][]bool) error {
	for y := 0; y < rows; y++ {
		written := 0
		for written < rowBytes {
			if pos >= end {
				return fmt.Errorf("第 %d 列提前結束", y)
			}
			code := data[pos]
			pos++
			count, value := 1, code
			switch code & 0xc0 {
			case 0x40:
				count, value = int(code&0x3f), 0xff
			case 0x80:
				count, value = int(code&0x3f), 0x00
			case 0xc0:
				count = int(code & 0x3f)
				if pos >= end {
					return fmt.Errorf("第 %d 列 literal run 缺值", y)
				}
				value = data[pos]
				pos++
			}
			if count <= 0 || written+count > rowBytes {
				return fmt.Errorf("第 %d 列 run 越界：written=%d count=%d width=%d", y, written, count, rowBytes)
			}
			for n := 0; n < count; n++ {
				for bit := 0; bit < 8; bit++ {
					opaque[y][(written+n)*8+bit] = value&(0x80>>bit) == 0
				}
			}
			written += count
		}
	}
	return nil
}
