// Package stats 是角色數值/升級系統。成長表與欄位順序直接依 DQ3.EXE
// sub_ed3c(file 0xed3c)的 writer→角色 record offset 對應，不再沿用舊 C remake
// 對表格欄位的推測。
package stats

import "github.com/wicanr2/dq3_remake_ebitan/internal/rng"

const (
	NumClass = 8
	MaxLevel = 43 // 門檻表 lv0..43,共 44 entry
)

// StatKind 順序就是 DQ3.EXE DS:4366 每列七組 {base,slope} 的順序。
// sub_ed3c 依序寫到角色 +1a/+1e/+24/+2a/+2c/+26/+28；再由
// sub_96be 狀態畫面及 1994 存檔欄位表確認語意。
type StatKind int

const (
	STR StatKind = iota // 力量 (+1a)
	VIT                 // 耐力 (+1e)
	AGI                 // 速度 (+24)
	HP                  // 最大 HP (+2a)
	MP                  // 最大 MP (+2c)
	INT                 // 聰明 (+26)
	LUCK
	StatCount
)

// Values 是角色 record 中七個持久能力欄。HP/MP 指最大值；目前值由 Game 另存。
type Values [StatCount]uint16

// 成長表 [class][14]:每屬性 2 byte {base,slope}，順序
// STR,VIT,AGI,HP,MP,INT,LUCK。以下是 DQ3.EXE file 0x1a4a6 的原始 bytes，
// 不偷偷套用舊 docs/23 的「MP 修正」；該文件把 pair1(VIT)誤認為 MP。
var growth = [NumClass][14]uint8{
	{0x06, 0x06, 0x03, 0x05, 0x03, 0x05, 0x08, 0x10, 0x08, 0x02, 0x06, 0x06, 0x06, 0x06},
	{0x08, 0x06, 0x04, 0x06, 0x03, 0x03, 0x08, 0x14, 0x00, 0x00, 0x04, 0x04, 0x04, 0x04},
	{0x0a, 0x08, 0x04, 0x08, 0x03, 0x09, 0x08, 0x16, 0x00, 0x00, 0x06, 0x04, 0x06, 0x06},
	{0x04, 0x06, 0x02, 0x05, 0x03, 0x04, 0x07, 0x0e, 0x05, 0x0e, 0x05, 0x08, 0x08, 0x08},
	{0x04, 0x04, 0x01, 0x05, 0x03, 0x06, 0x06, 0x0c, 0x05, 0x10, 0x06, 0x0a, 0x08, 0x08},
	{0x04, 0x06, 0x02, 0x05, 0x03, 0x04, 0x07, 0x0e, 0x05, 0x0c, 0x08, 0x08, 0x08, 0x08},
	{0x04, 0x06, 0x02, 0x05, 0x03, 0x05, 0x08, 0x10, 0x00, 0x00, 0x06, 0x04, 0x06, 0x06},
	{0x04, 0x06, 0x02, 0x05, 0x03, 0x05, 0x07, 0x0e, 0x00, 0x00, 0x06, 0x06, 0x06, 0x0a},
}

// 升級門檻(累積經驗)[class][lv0..43]。自 dq3x_thresh 抽出。
var thresh = [NumClass][MaxLevel + 1]uint32{
	{0, 29, 87, 174, 304, 499, 792, 1232, 1891, 2880, 4364, 6218, 8534, 11428, 15045, 19114, 23690, 28837, 34627, 41141, 48468, 56711, 65983, 76413, 88147, 101347, 116196, 132901, 151694, 172836, 196621, 223378, 253480, 287344, 325440, 368298, 416512, 470752, 531771, 600417, 677644, 764524, 862263, 960002},
	{0, 12, 36, 84, 156, 264, 426, 669, 1033, 1579, 2398, 3627, 5163, 7083, 9483, 12483, 16233, 20920, 26779, 34102, 42304, 51608, 62034, 73763, 86957, 101801, 118500, 137286, 158421, 182197, 208954, 239036, 272888, 310972, 353816, 402015, 456239, 517241, 585868, 663073, 749928, 847639, 945350, 1043061},
	{0, 18, 54, 126, 234, 396, 639, 1003, 1549, 2369, 3598, 5441, 7745, 10625, 14225, 18725, 24350, 30678, 37797, 45805, 54814, 64949, 76350, 89176, 103605, 119837, 138098, 158641, 181751, 207749, 236996, 269898, 306912, 348522, 395397, 448097, 507384, 574081, 649115, 733528, 828492, 935326, 1055514, 1190725},
	{0, 14, 42, 98, 182, 308, 497, 780, 1205, 1842, 2798, 4232, 6024, 8264, 11064, 14564, 18939, 24407, 30559, 37479, 45263, 54020, 63872, 74955, 87423, 101450, 117229, 134981, 154952, 177419, 202694, 231128, 263116, 299102, 339585, 385128, 463364, 494004, 558849, 631799, 713869, 806194, 910061, 1026912},
	{0, 20, 60, 140, 260, 440, 710, 1115, 1722, 2633, 3999, 6047, 8607, 11807, 15807, 20807, 27057, 34869, 43657, 53543, 64664, 77175, 91250, 107083, 124895, 144933, 167475, 192835, 221365, 253461, 289568, 330188, 375885, 427293, 485126, 550188, 623383, 705726, 798362, 902577, 1019818, 1151714, 1300096, 1448478},
	{0, 20, 60, 140, 260, 440, 710, 1115, 1722, 2633, 3999, 6047, 8607, 11807, 15807, 20807, 27057, 34869, 43657, 53543, 64664, 77175, 91250, 107083, 124895, 144933, 167475, 192835, 221365, 253461, 289568, 330188, 375885, 427293, 485126, 550188, 623383, 705726, 798362, 902577, 1019818, 1151714, 1300096, 1448478},
	{0, 10, 30, 70, 130, 220, 355, 557, 860, 1315, 1998, 3022, 4302, 5902, 7902, 10402, 13527, 17433, 22315, 27807, 33985, 40935, 48754, 57550, 67445, 78557, 91100, 105188, 121037, 138867, 158925, 181490, 206876, 235435, 267563, 303707, 344368, 390112, 441573, 499467, 564597, 637868, 720298, 813032},
	{0, 11, 33, 77, 143, 242, 390, 612, 946, 1447, 2198, 3324, 4732, 6492, 8692, 11442, 14879, 19175, 24545, 31258, 38810, 47306, 56863, 67614, 79709, 93316, 108623, 125844, 145217, 167012, 191531, 219114, 250145, 285055, 324329, 368511, 418216, 474134, 537042, 607813, 687430, 776998, 877762, 991121},
}

// GrowthTarget 回某職業某屬性在 level 的成長目標值:base + (slope×level)/2(16-bit)。移植 dq3_stats_growth_target。
func GrowthTarget(cls int, kind StatKind, level int) uint16 {
	if cls < 0 || cls >= NumClass || kind < 0 || kind >= StatCount {
		return 0
	}
	off := int(kind) * 2
	base := uint16(growth[cls][off])
	slope := uint16(growth[cls][off+1])
	return base + uint16((slope*uint16(level))>>1)
}

// legacyGrowAmt 忠實移植 sub_fa57：以 delta 做除數取餘；餘數為 0 時改成 1。
// 因此 delta>1 的結果是 [1,delta-1]，delta==1 固定為 1，而不是常見的
// [1,delta]。這個細節決定原版 Lv1 能力值範圍。
func legacyGrowAmt(delta int, r *rng.RNG) uint16 {
	if delta <= 0 {
		return 0
	}
	v := r.Next(delta)
	if v == 0 {
		v = 1
	}
	return uint16(v)
}

// InitValues 重現 sub_ed3c 的 Lv1 transaction：先把七欄寫成 base，再各自朝
// level 1 target 成長一次。呼叫者提供共享 DOS RNG，讓重抽與之後升級延續同一序列。
func InitValues(cls int, r *rng.RNG) Values {
	var v Values
	if cls < 0 || cls >= NumClass || r == nil {
		return v
	}
	for k := StatKind(0); k < StatCount; k++ {
		v[k] = uint16(growth[cls][int(k)*2])
		target := GrowthTarget(cls, k, 1)
		if target > v[k] {
			v[k] += legacyGrowAmt(int(target-v[k]), r)
		}
	}
	return v
}

// GrowLevel 就像 sub_ed3c 的非 Lv1 路徑，將現有角色的七欄向指定等級目標
// 各成長一次，並回傳各欄實際增量（目前 HP/MP 也要加上同一增量）。
func GrowLevel(cls, level int, v *Values, r *rng.RNG) Values {
	var gained Values
	if cls < 0 || cls >= NumClass || level < 1 || v == nil || r == nil {
		return gained
	}
	for k := StatKind(0); k < StatCount; k++ {
		target := GrowthTarget(cls, k, level)
		if target <= (*v)[k] {
			continue
		}
		add := legacyGrowAmt(int(target-(*v)[k]), r)
		(*v)[k] += add
		gained[k] = add
	}
	return gained
}

// LevelForExp 回某職業在累積經驗 exp 的等級([1,43] 夾住;#5 修正版,不越界)。移植 dq3_stats_level_for_exp(fixed)。
func LevelForExp(cls int, exp uint32) int {
	if cls < 0 || cls >= NumClass {
		return 1
	}
	level := 1
	for level < MaxLevel && exp >= thresh[cls][level+1] {
		level++
	}
	return level
}

// ExpForLevel 回某職業達到 level 所需累積經驗(門檻表查詢)。
func ExpForLevel(cls, level int) uint32 {
	if cls < 0 || cls >= NumClass || level < 0 || level > MaxLevel {
		return 0
	}
	return thresh[cls][level]
}
