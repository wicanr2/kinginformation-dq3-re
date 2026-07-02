package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// Member 是一名隊員(移植 dq3_recruit/dq3_member)。數值由職業成長表 + 裝備推導。
type Member struct {
	Name                        []int // glyph index(對 D3TXT00.FON)
	Class, Gender               int
	Exp                         uint32
	CurHP, CurMP                int
	Weapon, Armor, Shield, Head int // item code(0=無)
}

// 職業名 glyph(dq3_class_names):勇者/戰士/武鬥家/僧侶/魔法使者/賢者/商人/遊玩者。
var classNames = [8][]int{
	{106, 187}, {107, 144}, {108, 207, 657}, {109, 704},
	{110, 208, 210, 187}, {111, 187}, {112, 194}, {113, 705, 187},
}

func (m *Member) Level() int  { return stats.LevelForExp(m.Class, m.Exp) }
func (m *Member) MaxHP() int  { return int(stats.GrowthTarget(m.Class, stats.HP, m.Level())) }
func (m *Member) MaxMP() int  { return int(stats.GrowthTarget(m.Class, stats.MP, m.Level())) }
func (m *Member) Agi() int    { return int(stats.GrowthTarget(m.Class, stats.AGI, m.Level())) }
func (m *Member) Alive() bool { return m.CurHP > 0 }

// Atk = 力量 + 武器攻;Def = 耐力/2 + 鎧盾兜防(移植 dq3_recruit 數值公式)。
func (m *Member) Atk(items *dq3data.Items) int {
	a := int(stats.GrowthTarget(m.Class, stats.STR, m.Level()))
	if items != nil {
		a += items.Attack(m.Weapon)
	}
	return a
}
func (m *Member) Def(items *dq3data.Items) int {
	d := int(stats.GrowthTarget(m.Class, stats.VIT, m.Level())) / 2
	if items != nil {
		d += items.Defense(m.Armor) + items.Defense(m.Shield) + items.Defense(m.Head)
	}
	return d
}

// Spells = 該職業該等級已學可施放咒(class-aware)。
func (m *Member) Spells() []int { return spell.Known(m.Class, m.Level()) }

// heal 補至上限;fullHeal 補滿 HP/MP。
func (m *Member) fullHeal() { m.CurHP, m.CurMP = m.MaxHP(), m.MaxMP() }

// newMember 建一名 lv(由 exp 決定)的隊員並補滿 HP/MP。
func newMember(name []int, class, gender int, exp uint32) *Member {
	m := &Member{Name: name, Class: class, Gender: gender, Exp: exp}
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
