// Package itemuse 是消耗品使用效果(移植 dq3_item_use.c)。純邏輯:道具 id → 效果種類 +
// 數值(藥草治療、祈禱之戒 MP、聖水步數、損壞門檻)。效果種類為 RE 事實;數值採 classic DQ3 值
// (docs/49:ITEM.DAT 無使用威力欄,同咒文威力表處境)。
package itemuse

// Kind:效果種類(對齊 dq3_item_use.h enum)。
type Kind int

const (
	None          Kind = iota // 不可使用(裝備/鑰匙/劇情物)
	HealHP                    // 回 HP(藥草)
	CurePoison                // 解毒(驅毒草)
	CureParalysis             // 解麻痺(滿月草)
	ReturnTown                // 回最後城鎮(蓋美拉翅膀)
	Repel                     // 暫時驅弱敵(聖水)
	Awaken                    // 覺醒粉(位置相關)
	Gaia                      // 蓋亞之劍(位置相關,不消耗)
	Drain                     // 乾渴壺(位置相關,不消耗)
	FairyFlute                // 妖精之笛(位置相關,不消耗)
	PrayerRing                // 祈禱之戒(回 MP + ~25.4% 損壞)
	Ranaruta                  // 拉那魯達(切晝夜)
	DarkLamp                  // 黑暗之燈(強制變夜,不消耗)
	Rainbow                   // 彩虹水滴(位置相關)
	Mirror                    // 拉之鏡(沙曼歐莎假王揭露,不消耗)
	MagicBall                 // 魔法球(誘惑洞窟指定位置破牆)
)

// 消耗品 id(商店交叉驗證,docs/49)。
const (
	ItemHerb        = 0x41 // 藥草
	ItemAntidote    = 0x42 // 驅毒草
	ItemChimeraWing = 0x43 // 蓋美拉翅膀
	ItemHolyWater   = 0x44 // 聖水
	ItemFullmoon    = 0x45 // 滿月草
	ItemPrayerRing  = 0x48 // 祈禱之戒
	ItemMagicBall   = 0x58 // 魔法球
	ItemDarkLamp    = 0x5f // 黑暗之燈
	ItemMirror      = 0x61 // 拉之鏡
)

// 數值常數(classic DQ3 值 / RE 精確值,見 dq3_item_use.h)。
const (
	HerbHeal      = 30   // 藥草固定治療量
	PrayerMPAmt   = 30   // 祈禱之戒 MP 回復量
	HolySteps     = 64   // 聖水驅敵步數
	PrayerBreakLE = 0x40 // 祈禱之戒損壞門檻:RNG(256) ≤ 0x40 → 損壞(~25.4%)
)

// KindOf:道具 id → 效果種類(移植 dq3_item_use_kind)。
func KindOf(itemID int) Kind {
	switch itemID {
	case ItemHerb:
		return HealHP
	case ItemAntidote:
		return CurePoison
	case ItemChimeraWing:
		return ReturnTown
	case ItemHolyWater:
		return Repel
	case ItemFullmoon:
		return CureParalysis
	case ItemPrayerRing:
		return PrayerRing
	case ItemMagicBall:
		return MagicBall
	case ItemDarkLamp:
		return DarkLamp
	case ItemMirror:
		return Mirror
	case 0x5a:
		return Awaken // 覺醒粉(諾阿尼魯解催眠)
	case 0x0f:
		return Gaia // 蓋亞之劍(火山開通)
	case 0x5e:
		return Drain // 乾渴壺(吸海顯祠堂)
	case 0x77:
		return FairyFlute // 妖精之笛(魯比斯解詛咒)
	case 0x75:
		return Rainbow // 彩虹水滴(架彩虹橋)
	default:
		return None
	}
}

// HealHPAmount:對 (cur,max) 套藥草 → 實際回復量。陣亡(cur==0)/已滿回 0。移植 dq3_item_use_heal。
func HealHPAmount(cur, max int) int {
	if cur == 0 || cur >= max {
		return 0
	}
	healed := HerbHeal
	if cur+healed > max {
		healed = max - cur
	}
	return healed
}

// PrayerMPAmount:對 (cur,max) 套祈禱之戒 → 實際 MP 回復量。已滿回 0。移植 dq3_item_use_prayer_mp。
func PrayerMPAmount(cur, max int) int {
	if cur >= max {
		return 0
	}
	restored := PrayerMPAmt
	if cur+restored > max {
		restored = max - cur
	}
	return restored
}

// Consumable:此種類使用後是否消耗道具(位置相關/不消耗物件回 false)。
func (k Kind) Consumable() bool {
	switch k {
	case HealHP, CurePoison, CureParalysis, ReturnTown, Repel, Rainbow, MagicBall:
		return true // 立即效果 → 消耗(彩虹水滴確認位置後消耗)
	default:
		return false // 祈禱之戒(損壞才消耗,呼叫端處理)/ 武器道具 / 不可用
	}
}
