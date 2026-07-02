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

// Learn:某職業系在 level 級習得咒名 rec。
type Learn struct {
	Level, Rec int
}

// Def:咒文施放 descriptor(rec/MP/base 威力/種類)。target 省略(port 單人單敵)。
type Def struct {
	Rec, MP, Base int
	Kind          Kind
}

// 勇者系習得表(dq3_school_hero,18 咒,按等級)。
var heroSchool = []Learn{
	{2, 121}, {4, 161}, {6, 143}, {7, 172}, {10, 124}, {12, 160}, {14, 173}, {16, 144},
	{18, 156}, {19, 176}, {23, 125}, {26, 137}, {29, 162}, {31, 128}, {33, 163}, {35, 169},
	{38, 165}, {41, 138},
}

// 可施放咒文 descriptor(dq3_spelldef 的傷害/回復咒;輔助咒無 def → 戰鬥不列)。
var defs = map[int]Def{
	121: {121, 2, 10, Dmg},   // 美拉
	122: {122, 6, 80, Dmg},   // 美拉米
	123: {123, 10, 180, Dmg}, // 美拉宙瑪
	124: {124, 4, 20, Dmg},   // 吉拉
	125: {125, 6, 35, Dmg},   // 比吉拉瑪
	126: {126, 10, 100, Dmg}, // 比吉拉肯
	127: {127, 5, 20, Dmg},   // 伊歐
	128: {128, 8, 60, Dmg},   // 伊歐拉
	129: {129, 15, 140, Dmg}, // 伊歐那順
	130: {130, 3, 30, Dmg},   // 希亞多
	131: {131, 5, 50, Dmg},   // 希亞達可
	132: {132, 8, 70, Dmg},   // 希亞達依恩
	161: {161, 3, 30, Heal},  // 荷依米
	162: {162, 5, 85, Heal},  // 比荷依米
	163: {163, 7, 255, Heal}, // 比荷瑪
	165: {165, 36, 255, Heal}, // 比荷瑪順
}

// HeroKnown 回勇者在 level 級已學會、且可在戰鬥施放(有 def)的咒名 rec(按習得等級)。
func HeroKnown(level int) []int {
	var out []int
	for _, l := range heroSchool {
		if l.Level <= level {
			if _, ok := defs[l.Rec]; ok {
				out = append(out, l.Rec)
			}
		}
	}
	return out
}

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
