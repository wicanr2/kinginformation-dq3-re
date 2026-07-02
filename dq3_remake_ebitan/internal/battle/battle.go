// Package battle 是戰鬥結算公式(移植自 C dq3_battle.c,docs/13)。純邏輯、無引擎相依。
// 反組譯真公式(file 0xc03e):物理傷害、傷害套用、逃跑判定、勝負結算。
package battle

// Outcome 是一回合結算後的戰鬥狀態。
type Outcome int

const (
	Ongoing Outcome = iota
	Victory
	Defeat
)

// Alive 回存活數(HP>0 且未被移出/吹飛)。移植 dq3_battle_alive。
func Alive(hp []uint16, removed []bool, n int) int {
	c := 0
	for i := 0; i < n && i < len(hp); i++ {
		if hp[i] > 0 && (removed == nil || i >= len(removed) || !removed[i]) {
			c++
		}
	}
	return c
}

// Resolve 判勝負(#1 修正:先判我方全滅 → 敗,杜絕「全隊被吹飛卻誤勝」)。移植 dq3_battle_resolve。
func Resolve(partyHP []uint16, partyRemoved []bool, partyN int, enemyHP []uint16, enemyN int) Outcome {
	if Alive(partyHP, partyRemoved, partyN) == 0 {
		return Defeat
	}
	if Alive(enemyHP, nil, enemyN) == 0 {
		return Victory
	}
	return Ongoing
}

// ResolveBuggy 重現原版缺陷:漏判我方全滅 → 全隊被吹飛時誤判勝。移植 dq3_battle_resolve_buggy(保留供對拍/文件)。
func ResolveBuggy(partyHP []uint16, partyRemoved []bool, partyN int, enemyHP []uint16, enemyN int) Outcome {
	if Alive(partyHP, partyRemoved, partyN) == 0 {
		return Victory // ★ bug:全隊被吹飛 → 誤判勝
	}
	if Alive(enemyHP, nil, enemyN) == 0 {
		return Victory
	}
	return Ongoing
}

// ApplyDamage 扣血(uint16 不下溢)。移植 dq3_battle_apply_damage。
func ApplyDamage(hp uint16, dmg int) uint16 {
	v := int(hp) - dmg
	if v < 0 {
		v = 0
	}
	return uint16(v)
}

// PhysDamage 物理傷害。移植 dq3_battle_phys_damage(file 0xc03e):
//
//	atk > def:dmg = (atk−def)/2 + rng(0..(atk−def)/4);會心 → ×2。
//	atk ≤ def(弱攻):roll<0x80(50%)→ 1 點,否則 miss(0)。roll 0..255。
func PhysDamage(atk, def, roll, crit int) int {
	if roll < 0 {
		roll = 0
	}
	if roll > 255 {
		roll = 255
	}
	diff := atk - def
	if diff <= 0 {
		if roll < 0x80 {
			return 1
		}
		return 0
	}
	base := diff / 2
	quarter := diff / 4
	variance := 0
	if quarter > 0 {
		variance = quarter * roll / 256
	}
	dmg := base + variance
	if crit != 0 {
		dmg *= 2
	}
	return dmg
}

// FleeOK 逃跑成功判定(agi 越高越易、敵抗性越高越難;roll 0..255)。移植 dq3_battle_flee_ok。
func FleeOK(partyAgi, enemyFleeResist, roll int) bool {
	chance := 128 + partyAgi - enemyFleeResist*2
	if chance < 16 {
		chance = 16
	}
	if chance > 240 {
		chance = 240
	}
	return roll < chance
}
