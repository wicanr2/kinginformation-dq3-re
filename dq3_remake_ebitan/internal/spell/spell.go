// Package spell 提供共用咒文型別、戰鬥咒文與原始 EXE 習得表 oracle。
// 正式野外咒文 descriptor 由 game pack 載入；本套件內的表不可作為缺資料時的 fallback。
// 傷害/回復公式(file 0xc22e):val = base/2 + rng(base/2)。咒名 rec 對 D3TXT00 名表。
package spell

// Kind:咒文種類。
type Kind int

const (
	Dmg Kind = iota
	Heal
	Revive
	Sleep      // 拉里荷144/美達巴尼152:施法方對敵對側單體下睡眠/混亂(remake 映射同一狀態,不能行動)
	BuffAtk    // 拜基魯多151:施法方全體物攻上升(caster 自施,W3 新增)
	BuffDef    // 史卡拉154/史克魯多155:施法方全體守備上升(caster 自施,W3 新增)
	Seal       // 瑪荷頓156:封敵對側單體咒文(不能施咒)
	Blind      // 瑪努莎158:敵對側單體陷入幻惑(物攻易失手)
	CurePoison // 基阿里166:解毒
	CureStatus // 基阿里克167/薩梅哈168：解持久麻痺及戰鬥暫態睡眠／混亂；runtime 以不同位元保存
	Palpunte   // 帕魯朋特180:16 格原版亂數表（5 格有效、4 種唯一戰鬥效果）
)

// Target 由 DQ3.EXE DS:0x37c3 的 descriptor flags 直接解出。
// 0x0d=敵單體、0x15/0x1d=敵群體、0x09/0x0b=我方單體、0x11/0x13=我方全體。
type Target int

const (
	TargetEnemyOne Target = iota
	TargetEnemyGroup
	TargetAllyOne
	TargetAllyGroup
	TargetNone
)

// Learn:某職業系在 level 級習得咒名 rec。
type Learn struct {
	Level, Rec int
}

// Def:咒文施放 descriptor(rec/MP/base 威力/種類/目標範圍)。移植 dq3_spelldef.c。
type Def struct {
	Rec, MP, Base int
	Kind          Kind
	Target        Target
}

// mpCosts 是 DQ3.EXE DGROUP 0x37c3 起、60 筆×3-byte spell descriptor 的第 0 byte。
// 原版 field/battle caster file 0x1cb61..0x1cb84 以 spellID*3 查此 byte並扣 caster MP。
// rec = spellID + 0x79。這取代舊 C remake 的 FC 推測值。
var mpCosts = [...]int{
	2, 6, 12, 4, 6, 12, 5, 9, 18, 3, 6, 9, 12, 4, 6, 9, 8, 30, 7, 7,
	1, 24, 2, 3, 3, 3, 4, 7, 3, 0, 6, 5, 12, 3, 4, 3, 8, 4, 6, 6,
	3, 5, 7, 18, 62, 3, 6, 3, 10, 20, 18, 8, 8, 3, 2, 4, 12, 15, 0, 20,
}

// MPCost 回精訊版 EXE 的原生 MP cost；非咒文 record 回 -1。
func MPCost(rec int) int {
	i := rec - 0x79
	if i < 0 || i >= len(mpCosts) {
		return -1
	}
	return mpCosts[i]
}

// BattleUsable reports whether rec has a battle-caster descriptor. Learned
// spell persistence stores both field and battle records; battle menus must
// filter the retained union through this predicate.
func BattleUsable(rec int) bool {
	_, ok := defs[rec]
	return ok
}

// 勇者系習得表(dq3_school_hero,18 咒,按等級)。
var heroSchool = []Learn{
	{2, 121}, {4, 161}, {6, 143}, {7, 172}, {10, 124}, {12, 160}, {14, 173}, {16, 144},
	{18, 156}, {19, 176}, {23, 125}, {26, 137}, {29, 162}, {31, 128}, {33, 163}, {35, 169},
	{38, 165}, {41, 138},
}

// 可施放咒文 descriptor(dq3_spelldef 的傷害/回復咒;輔助咒無 def → 戰鬥不列)。
// Target 直接對齊原版 descriptor 的敵我側別與單體／群體 flags。
var defs = map[int]Def{
	121: {121, 2, 10, Dmg, TargetEnemyOne},     // 美拉
	122: {122, 6, 80, Dmg, TargetEnemyOne},     // 美拉米
	123: {123, 12, 180, Dmg, TargetEnemyOne},   // 美拉宙瑪
	124: {124, 4, 20, Dmg, TargetEnemyGroup},   // 吉拉
	125: {125, 6, 35, Dmg, TargetEnemyGroup},   // 比吉拉瑪
	126: {126, 12, 100, Dmg, TargetEnemyGroup}, // 比吉拉肯
	127: {127, 5, 20, Dmg, TargetEnemyGroup},   // 伊歐
	128: {128, 9, 60, Dmg, TargetEnemyGroup},   // 伊歐拉
	129: {129, 18, 140, Dmg, TargetEnemyGroup}, // 伊歐那順
	130: {130, 3, 30, Dmg, TargetEnemyOne},     // 希亞多
	131: {131, 6, 50, Dmg, TargetEnemyGroup},   // 希亞達可
	132: {132, 9, 70, Dmg, TargetEnemyGroup},   // 希亞達依恩
	161: {161, 3, 30, Heal, TargetAllyOne},     // 荷依米
	162: {162, 5, 85, Heal, TargetAllyOne},     // 比荷依米
	163: {163, 7, 255, Heal, TargetAllyOne},    // 比荷瑪
	165: {165, 62, 255, Heal, TargetAllyGroup}, // 比荷瑪順

	// 輔助/狀態咒(base==0,dq3_spelldef.c 生成表未收錄——該表只列 DMG/HEAL 有威力值的咒;
	// docs/data/spell-effects-research.md 為 remake 定義的效果模型,W3 補上玩家可施放的 descriptor。
	// MP 由 EXE 0x37c3 descriptor 第 0 byte 證實；Base 恆 0,
	// 不進 CastValue(這些 Kind 不算傷害/回復量,見 game/battle.go execSpell)。
	144: {144, 3, 0, Sleep, TargetEnemyGroup},     // 拉里荷
	151: {151, 6, 0, BuffAtk, TargetAllyOne},      // 拜基魯多
	152: {152, 5, 0, Sleep, TargetEnemyOne},       // 美達巴尼
	154: {154, 3, 0, BuffDef, TargetAllyOne},      // 史卡拉
	155: {155, 4, 0, BuffDef, TargetAllyGroup},    // 史克魯多
	156: {156, 3, 0, Seal, TargetEnemyGroup},      // 瑪荷頓
	158: {158, 4, 0, Blind, TargetEnemyGroup},     // 瑪努莎
	166: {166, 3, 0, CurePoison, TargetAllyOne},   // 基阿里
	167: {167, 6, 0, CureStatus, TargetAllyOne},   // 基阿里克
	168: {168, 3, 0, CureStatus, TargetAllyGroup}, // 薩梅哈
	180: {180, 20, 0, Palpunte, TargetNone},       // 帕魯朋特
}

// 僧侶系(24)/ 魔法系(31)習得表(自 dq3_school_priest/mage)。
var priestSchool = []Learn{{1, 161}, {2, 143}, {5, 145}, {7, 158}, {8, 146}, {9, 144}, {11, 166}, {12, 134}, {13, 156}, {14, 162}, {15, 167}, {16, 168}, {18, 147}, {20, 148}, {22, 139}, {24, 169}, {26, 135}, {28, 140}, {30, 163}, {32, 159}, {34, 164}, {36, 136}, {38, 170}, {41, 141}}
var mageSchool = []Learn{{1, 121}, {4, 154}, {5, 130}, {7, 124}, {8, 173}, {9, 155}, {11, 127}, {12, 172}, {13, 149}, {14, 125}, {15, 150}, {17, 122}, {18, 174}, {19, 175}, {20, 131}, {21, 151}, {23, 128}, {24, 157}, {25, 177}, {26, 132}, {27, 152}, {29, 126}, {30, 171}, {32, 133}, {33, 178}, {34, 142}, {35, 179}, {36, 123}, {37, 153}, {38, 129}, {40, 180}}

// Known 回職業 cls 在 level 級已學會、且可在戰鬥施放(有 def)的咒名 rec。移植 dq3_spells_known。
// 職業→系:勇者0→勇者系、僧侶3→僧侶、魔法4→魔法、賢者5→僧侶+魔法;其餘無咒。
func Known(cls, level int) []int {
	all := KnownAll(cls, level)
	out := all[:0]
	for _, rec := range all {
		if _, ok := defs[rec]; ok {
			out = append(out, rec)
		}
	}
	return out
}

// KnownAll 回傳職業在 level 已學會的全部咒文，包含沒有戰鬥 descriptor 的野外工具咒。
// 野外命令窗必須用此 API；Known 只供戰鬥選單，不能再因 defs 過濾掉魯拉／烈米特。
func KnownAll(cls, level int) []int {
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
				out = append(out, l.Rec)
				seen[l.Rec] = true
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

// 歷史 lookup：怪物 action bit(0..47)→ 候選文字 rec。現行 Battle 對未由 pack 接管的
// bit 仍以此表嘗試普通 spell descriptor；這是尚待逐 bit RE 的相容近似，不是原版 exact，
// 也尚未達成未知 action 失敗即關閉。
var monsterSpellRec = [48]int{
	121, 122, 123, 124, 125, 126, 127, 128, 129, 130, 131, 132,
	133, 134, 135, 136, 139, 140, 141, 144, 145, 146, 147, 148,
	149, 150, 152, 154, 155, 156, 157, 158, 161, 162, 163, 164,
	169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179, 180,
}

// MonsterSpellRec 保留歷史 API 名稱；回傳 action 的候選文字 rec，越界回 0。
func MonsterSpellRec(bit int) int {
	if bit < 0 || bit >= 48 {
		return 0
	}
	return monsterSpellRec[bit]
}

// MonsterSpellBits 保留歷史 API 名稱；只列出 raw action mask 的 set bit。
func MonsterSpellBits(mask [6]uint8) []int {
	var out []int
	for b := 0; b < 48; b++ {
		if mask[b/8]&(0x80>>(b%8)) != 0 {
			out = append(out, b)
		}
	}
	return out
}
