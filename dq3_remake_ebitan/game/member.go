package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// conditionSet 是跨場景持久狀態；數值是 remake save schema，不冒充原版
// character +0x38 的 bit layout。各 game pack 仍必須提供原版 mask 與 consumer 證據。
type conditionSet uint16

const (
	conditionPoison conditionSet = 1 << iota
	conditionParalysis
)

func conditionBit(id string) (conditionSet, bool) {
	switch id {
	case gamepack.PoisonCondition:
		return conditionPoison, true
	case gamepack.ParalysisCondition:
		return conditionParalysis, true
	default:
		return 0, false
	}
}

// Member 是一名隊員(移植 dq3_recruit/dq3_member)。七項能力是原版角色 record 的持久值；
// 攻防再疊加裝備，不可每幀由職業/等級 target 反推。
type Member struct {
	Name          []int // glyph index(對 D3TXT00.FON)
	Class, Gender int
	Exp           uint32
	Stats         stats.Values // 原版角色 record 的七個持久能力欄
	CurHP, CurMP  int
	Items         itemstore.Store // 唯一持有權集合，保留原始物理格與words
	LearnedSpells []int           // 已永久學會的全咒文 record；轉職後不清除
	Conditions    conditionSet
}

// 職業名 glyph(dq3_class_names):勇者/戰士/武鬥家/僧侶/魔法使者/賢者/商人/遊玩者。
var classNames = [8][]int{
	{106, 187}, {107, 144}, {108, 207, 657}, {109, 704},
	{110, 208, 210, 187}, {111, 187}, {112, 194}, {113, 705, 187},
}

func (m *Member) Level() int { return stats.LevelForExp(m.Class, m.Exp) }
func (m *Member) ensureStats() {
	if m.Stats != (stats.Values{}) {
		return
	}
	for k := stats.StatKind(0); k < stats.StatCount; k++ {
		m.Stats[k] = stats.GrowthTarget(m.Class, k, m.Level())
	}
}
func (m *Member) MaxHP() int  { m.ensureStats(); return int(m.Stats[stats.HP]) }
func (m *Member) MaxMP() int  { m.ensureStats(); return int(m.Stats[stats.MP]) }
func (m *Member) Agi() int    { m.ensureStats(); return int(m.Stats[stats.AGI]) }
func (m *Member) Alive() bool { return m.CurHP > 0 }

// Atk = 力量 + 武器攻;Def = 耐力 + 鎧盾兜防（DQ3.EXE sub_9521）。
func (m *Member) Atk(items *dq3data.Items) int {
	m.ensureStats()
	a := int(m.Stats[stats.STR])
	if items != nil {
		a += items.Attack(m.Items.Equipment()[0])
	}
	return a
}
func (m *Member) Def(items *dq3data.Items) int {
	m.ensureStats()
	d := int(m.Stats[stats.VIT])
	if items != nil {
		d += items.Defense(m.Items.Equipment()[1]) + items.Defense(m.Items.Equipment()[2]) + items.Defense(m.Items.Equipment()[3])
	}
	return d
}

// Spells = 該職業該等級已學可施放咒(class-aware)。
func appendUnique(dst []int, src ...int) []int {
	seen := make(map[int]bool, len(dst)+len(src))
	for _, value := range dst {
		seen[value] = true
	}
	for _, value := range src {
		if !seen[value] {
			dst = append(dst, value)
			seen[value] = true
		}
	}
	return dst
}

func (m *Member) syncLearnedSpells() {
	m.LearnedSpells = appendUnique(m.LearnedSpells, spell.KnownAll(m.Class, m.Level())...)
}

func (m *Member) AllSpells() []int {
	m.syncLearnedSpells()
	return append([]int(nil), m.LearnedSpells...)
}

func (m *Member) Spells() []int {
	all := m.AllSpells()
	out := all[:0]
	for _, rec := range all {
		if spell.BattleUsable(rec) {
			out = append(out, rec)
		}
	}
	return out
}

func (m *Member) equippedItems() []int {
	equipment := m.Items.Equipment()
	out := make([]int, 0, len(equipment))
	for _, code := range equipment {
		if code >= 0 {
			out = append(out, code)
		}
	}
	return out
}

func (m *Member) itemCount() int {
	return len(m.Items.Entries())
}

// heal 補至上限;fullHeal 補滿 HP/MP。
func (m *Member) fullHeal() { m.CurHP, m.CurMP = m.MaxHP(), m.MaxMP() }

// newLevelOneMember 重現 DQ3.EXE sub_1c94 → sub_ed3c 的登錄所 Lv1 建角：
// 共用全域 DOS RNG 擲七欄，並穿上布衣(item 0x1e)。
func newLevelOneMember(name []int, class, gender int, r *rng.RNG, items itemstore.Store) *Member {
	m := &Member{Name: append([]int(nil), name...), Class: class, Gender: gender,
		Items: items.Clone()}
	m.Stats = stats.InitValues(class, r)
	m.syncLearnedSpells()
	m.fullHeal()
	return m
}
