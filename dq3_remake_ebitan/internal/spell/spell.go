// Package spell 是咒文系統(移植 dq3_spell:習得表 + 施放 descriptor)。純邏輯 + 內建 EXE 表。
// 傷害/回復公式(file 0xc22e):val = base/2 + rng(base/2)。咒名 rec 對 D3TXT00 名表。
package spell

// Kind:咒文種類。
type Kind int

const (
	Dmg Kind = iota
	Heal
	Revive
)

// Target:傷害咒的目標範圍(移植 dq3_spelldef.c DQ3_TG_*)。Heal 不用此欄(固定治己方,見
// battle.go execSpell)。ENEMY1=單體(玩家端簡化為「指向第一個存活敵」,原版隨機選);
// Group=全體存活敵(原版 TG_GROUP/TG_ALL 在玩家咒文的敵方視角下等價,均全體命中)。
type Target int

const (
	TargetSingle Target = iota // DQ3_TG_ENEMY1
	TargetGroup                // DQ3_TG_GROUP / DQ3_TG_ALL(玩家咒對敵方視角等價,皆全體)
)

// Learn:某職業系在 level 級習得咒名 rec。
type Learn struct {
	Level, Rec int
}

// Def:咒文施放 descriptor(rec/MP/base 威力/種類/目標範圍)。移植 dq3_spelldef.c。
type Def struct {
	Rec, MP, Base int
	Kind          Kind
	Target        Target // 僅 Dmg 咒有意義(對齊 W2 多敵戰鬥:單體咒指定目標、群體咒全部命中)
}

// 勇者系習得表(dq3_school_hero,18 咒,按等級)。
var heroSchool = []Learn{
	{2, 121}, {4, 161}, {6, 143}, {7, 172}, {10, 124}, {12, 160}, {14, 173}, {16, 144},
	{18, 156}, {19, 176}, {23, 125}, {26, 137}, {29, 162}, {31, 128}, {33, 163}, {35, 169},
	{38, 165}, {41, 138},
}

// 可施放咒文 descriptor(dq3_spelldef 的傷害/回復咒;輔助咒無 def → 戰鬥不列)。
// Target 對齊 src/dq3_spelldef.c(DQ3_TG_ENEMY1→TargetSingle;DQ3_TG_GROUP/DQ3_TG_ALL→TargetGroup,
// 玩家咒對敵方視角下皆「全體存活敵」等價)。
var defs = map[int]Def{
	121: {121, 2, 10, Dmg, TargetSingle},    // 美拉
	122: {122, 6, 80, Dmg, TargetSingle},    // 美拉米
	123: {123, 10, 180, Dmg, TargetSingle},  // 美拉宙瑪
	124: {124, 4, 20, Dmg, TargetGroup},     // 吉拉
	125: {125, 6, 35, Dmg, TargetGroup},     // 比吉拉瑪
	126: {126, 10, 100, Dmg, TargetGroup},   // 比吉拉肯
	127: {127, 5, 20, Dmg, TargetGroup},     // 伊歐
	128: {128, 8, 60, Dmg, TargetGroup},     // 伊歐拉
	129: {129, 15, 140, Dmg, TargetGroup},   // 伊歐那順
	130: {130, 3, 30, Dmg, TargetSingle},    // 希亞多
	131: {131, 5, 50, Dmg, TargetGroup},     // 希亞達可
	132: {132, 8, 70, Dmg, TargetGroup},     // 希亞達依恩
	161: {161, 3, 30, Heal, TargetSingle},   // 荷依米
	162: {162, 5, 85, Heal, TargetSingle},   // 比荷依米
	163: {163, 7, 255, Heal, TargetSingle},  // 比荷瑪
	165: {165, 36, 255, Heal, TargetSingle}, // 比荷瑪順
}

// 僧侶系(24)/ 魔法系(31)習得表(自 dq3_school_priest/mage)。
var priestSchool = []Learn{{1, 161}, {2, 143}, {5, 145}, {7, 158}, {8, 146}, {9, 144}, {11, 166}, {12, 134}, {13, 156}, {14, 162}, {15, 167}, {16, 168}, {18, 147}, {20, 148}, {22, 139}, {24, 169}, {26, 135}, {28, 140}, {30, 163}, {32, 159}, {34, 164}, {36, 136}, {38, 170}, {41, 141}}
var mageSchool = []Learn{{1, 121}, {4, 154}, {5, 130}, {7, 124}, {8, 173}, {9, 155}, {11, 127}, {12, 172}, {13, 149}, {14, 125}, {15, 150}, {17, 122}, {18, 174}, {19, 175}, {20, 131}, {21, 151}, {23, 128}, {24, 157}, {25, 177}, {26, 132}, {27, 152}, {29, 126}, {30, 171}, {32, 133}, {33, 178}, {34, 142}, {35, 179}, {36, 123}, {37, 153}, {38, 129}, {40, 180}}

// Known 回職業 cls 在 level 級已學會、且可在戰鬥施放(有 def)的咒名 rec。移植 dq3_spells_known。
// 職業→系:勇者0→勇者系、僧侶3→僧侶、魔法4→魔法、賢者5→僧侶+魔法;其餘無咒。
func Known(cls, level int) []int {
	var schools [][]Learn
	switch cls {
	case 0:
		schools = [][]Learn{heroSchool}
	case 3:
		schools = [][]Learn{priestSchool}
	case 4:
		schools = [][]Learn{mageSchool}
	case 5:
		schools = [][]Learn{priestSchool, mageSchool}
	default:
		return nil
	}
	seen := map[int]bool{}
	var out []int
	for _, s := range schools {
		for _, l := range s {
			if l.Level <= level && !seen[l.Rec] {
				if _, ok := defs[l.Rec]; ok {
					out = append(out, l.Rec)
					seen[l.Rec] = true
				}
			}
		}
	}
	return out
}

// HeroKnown 回勇者(class0)已學可施放咒(向後相容)。
func HeroKnown(level int) []int { return Known(0, level) }

// GetDef 查咒文 rec 的施放 descriptor。
func GetDef(rec int) (Def, bool) {
	d, ok := defs[rec]
	return d, ok
}

// CastValue:傷害/回復量 = base/2 + rng(base/2)。roll 0..255。移植 file 0xc22e。
func CastValue(base, roll int) int {
	half := base / 2
	return half + half*roll/256
}

// 怪物咒文 bit(0..47)→ 咒名 rec(移植 dq3_monster_spell_rec,EXE 0x3930 remap,docs/37)。
var monsterSpellRec = [48]int{
	121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
	133, 134, 135, 136, 139, 140, 141, 144, 145, 146, 147, 148,
	149, 150, 152, 154, 155, 156, 157, 158, 161, 162, 163, 164,
	169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180,
}

// MonsterSpellRec:怪物咒文 bit → 咒名 rec;越界回 0。
func MonsterSpellRec(bit int) int {
	if bit < 0 || bit >= 48 {
		return 0
	}
	return monsterSpellRec[bit]
}

// MonsterSpellBits:自 spell_mask[6] 取出所有已知咒文 bit(0..47)。
func MonsterSpellBits(mask [6]uint8) []int {
	var out []int
	for b := 0; b < 48; b++ {
		if mask[b/8]&(0x80>>(b%8)) != 0 {
			out = append(out, b)
		}
	}
	return out
}
