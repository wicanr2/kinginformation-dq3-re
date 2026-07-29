package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// Member 是一名隊員(移植 dq3_recruit/dq3_member)。七項能力是原版角色 record 的持久值；
// 攻防再疊加裝備，不可每幀由職業/等級 target 反推。
type Member struct {
	Name                        []int // glyph index(對 D3TXT00.FON)
	Class, Gender               int
	Exp                         uint32
	Stats                       stats.Values // 原版角色 record 的七個持久能力欄
	CurHP, CurMP                int
	Weapon, Armor, Shield, Head int // item code；-1=無（0x00 是合法檜木棒）
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
		a += items.Attack(m.Weapon)
	}
	return a
}
func (m *Member) Def(items *dq3data.Items) int {
	m.ensureStats()
	d := int(m.Stats[stats.VIT])
	if items != nil {
		d += items.Defense(m.Armor) + items.Defense(m.Shield) + items.Defense(m.Head)
	}
	return d
}

// Spells = 該職業該等級已學可施放咒(class-aware)。
func (m *Member) Spells() []int    { return spell.Known(m.Class, m.Level()) }
func (m *Member) AllSpells() []int { return spell.KnownAll(m.Class, m.Level()) }

// heal 補至上限;fullHeal 補滿 HP/MP。
func (m *Member) fullHeal() { m.CurHP, m.CurMP = m.MaxHP(), m.MaxMP() }

// newMember 建一名 lv(由 exp 決定)的隊員並補滿 HP/MP。
func newMember(name []int, class, gender int, exp uint32) *Member {
	m := &Member{Name: name, Class: class, Gender: gender, Exp: exp,
		Weapon: -1, Armor: -1, Shield: -1, Head: -1}
	m.ensureStats() // 舊測試/debug 建角的相容路徑；正式登錄走 newLevelOneMember。
	m.fullHeal()
	return m
}

// newLevelOneMember 重現 DQ3.EXE sub_1c94 → sub_ed3c 的登錄所 Lv1 建角：
// 共用全域 DOS RNG 擲七欄，並穿上布衣(item 0x1e)。
func newLevelOneMember(name []int, class, gender int, r *rng.RNG, equipment [4]int) *Member {
	m := &Member{Name: append([]int(nil), name...), Class: class, Gender: gender,
		Weapon: equipment[0], Armor: equipment[1], Shield: equipment[2], Head: equipment[3]}
	m.Stats = stats.InitValues(class, r)
	m.fullHeal()
	return m
}

// startingCompanions:示範用預建隊伍(戰士/僧侶/魔法使,與隊長同經驗)。
// 註:忠實 DQ3 是在露易達酒館逐一創角招募(名字用注音輸入);此為 tavern UI 完成前的 placeholder。
func startingCompanions(exp uint32) []*Member {
	return []*Member{
		newMember(classNames[1], 1, 0, exp), // 戰士
		newMember(classNames[3], 3, 0, exp), // 僧侶
		newMember(classNames[4], 4, 1, exp), // 魔法使者
	}
}
