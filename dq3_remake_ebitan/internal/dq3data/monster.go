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
