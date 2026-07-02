package dq3data

import "fmt"

// 怪物數值/AI(D3MNS.DAT,130 隻 × 41 byte/筆)。移植自 C dq3_monster.c(docs/16)。
const (
	MonsterCount = 130
	monsterRec   = 41
)

// MonsterStat 是一隻怪的數值(欄位偏移對齊 C）。
type MonsterStat struct {
	HPBase, HPRand uint16 // +0x00 / +0x02
	Atk            uint8  // +0x08
	Def, Agi       uint16 // +0x09 / +0x0b
	FleeResist     uint8  // +0x18
	Exp, Gold      uint16 // +0x21 / +0x23
	SpawnWeight    uint8  // +0x28
}

// MonsterAI 是一隻怪的行為(咒文機率/逃跑/已知咒文)。
type MonsterAI struct {
	CastProb   uint8    // +0x0d:rng(256) < 此 → 放咒
	FleeThresh uint8    // +0x17:我方平均等級 ≥ 此 → 考慮逃跑
	FleeRate   uint8    // +0x18:觸發後 rng(256) ≤ 此就逃(0=不逃)
	SpellMask  [6]uint8 // +0x0e..+0x13:已知咒文 bitmask
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
		Atk: r[0x08], Def: le16(r, 0x09), Agi: le16(r, 0x0b),
		FleeResist: r[0x18], Exp: le16(r, 0x21), Gold: le16(r, 0x23), SpawnWeight: r[0x28],
	}, true
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

// MonsterSprite 是一隻怪的圖(palette 索引對 MNSBK.PAL;色 0 = 透明)。
type MonsterSprite struct {
	W, H   int
	Px     [][]uint8 // [H][W] palette 索引
	Opaque [][]bool  // [H][W] true=畫
}

// DecodeMonsterSprite 解 DQ3MNS.SHP 第 id 隻。移植 dq3_monster_sprite_decode。
// SHP:131×u32 offset 表 + 每隻 {u16 w_bytes, u16 h, 4 plane plane-major(plane 3,2,1,0)}。
// 回 nil,error 表空 sprite(如未完成 boss id128/129)或越界。
func DecodeMonsterSprite(shp []byte, id int) (*MonsterSprite, error) {
	if id < 0 || id >= MonsterCount {
		return nil, fmt.Errorf("id %d 越界", id)
	}
	if (id+1)*4+4 > len(shp) {
		return nil, fmt.Errorf("offset 表越界")
	}
	off := int(le32(shp, id*4))
	nxt := int(le32(shp, (id+1)*4))
	if off+4 > len(shp) || nxt <= off {
		return nil, fmt.Errorf("空 sprite(id %d)", id)
	}
	wb := int(le16(shp, off) & 0x7fff)
	h := int(le16(shp, off+2))
	w := wb * 8
	if w <= 0 || h <= 0 || w > ShpMaxW || h > ShpMaxH {
		return nil, fmt.Errorf("尺寸越界 %d×%d", w, h)
	}
	base := off + 4
	planeSz := wb * h
	if base+planeSz*4 > len(shp) {
		return nil, fmt.Errorf("sprite 資料越界")
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
				v := shp[b0+r*wb+b]
				for bit := 0; bit < 8; bit++ {
					if v&(0x80>>bit) != 0 {
						spr.Px[r][b*8+bit] |= 1 << pl
					}
				}
			}
		}
	}
	for r := 0; r < h; r++ {
		for b := 0; b < w; b++ {
			spr.Opaque[r][b] = spr.Px[r][b] != 0 // 色0=透明
		}
	}
	return spr, nil
}
