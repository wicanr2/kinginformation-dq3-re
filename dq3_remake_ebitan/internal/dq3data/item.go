package dq3data

import "fmt"

// 道具/裝備資料(ITEM.DAT,128 項 × 7 byte)。移植自 C dq3_combat.c(docs/22)。
// 7-byte:b0 攻、b1 防、b2+b3*256 價、b4 類別/部位、b5 武器旗標、b6 可裝備職業 bitmask。
const (
	ItemCount  = 128
	itemStride = 7
)

// Items 持有 ITEM.DAT 整檔。
type Items struct{ raw []byte }

// OpenItems 解 ITEM.DAT。移植 dq3_items_load。
func OpenItems(d []byte) (*Items, error) {
	if len(d) < ItemCount*itemStride {
		return nil, fmt.Errorf("ITEM.DAT too small: %d < %d", len(d), ItemCount*itemStride)
	}
	return &Items{raw: d}, nil
}

func (it *Items) b(code, off int) int {
	if code < 0 || code >= ItemCount {
		return 0
	}
	return int(it.raw[code*itemStride+off])
}

// Attack 回武器攻擊力。移植 dq3_item_attack。
func (it *Items) Attack(code int) int { return it.b(code, 0) }

// Defense 回防具防禦力。移植 dq3_item_defense。
func (it *Items) Defense(code int) int { return it.b(code, 1) }

// Price 回購買價(b2 + b3*256)。移植 dq3_item_price。
func (it *Items) Price(code int) int { return it.b(code, 2) | it.b(code, 3)<<8 }

// Category 回類別/部位(b4)。移植 dq3_item_category。
func (it *Items) Category(code int) int { return it.b(code, 4) }

// MetadataWord 回 ITEM.DAT +4/+5 的 little-endian 原始 word。這是原始格式
// oracle；各 bit 的玩家可見語意仍須由 EXE writer/consumer 個別閉合。
func (it *Items) MetadataWord(code int) int {
	return it.b(code, 4) | it.b(code, 5)<<8
}

// CursedWhenEquipped 回原版 sub_17ED9 的精確 writer gate：ITEM +4/+5
// metadata word 與 0x0e00 相交時，裝備後的角色 item word 會 OR bit0x4000。
// 此方法只作 DQ3 原始資料 parity oracle；production 清單仍由 game pack 提供。
func (it *Items) CursedWhenEquipped(code int) bool {
	return it.MetadataWord(code)&0x0e00 != 0
}

// EquipSlot 回引擎部位 0武器/1鎧/2盾/3兜。ITEM.DAT 類別群實為
// 0x2_=武器、0x4_=鎧、0x6_=兜、0x8_/0x9_=盾；最後兩群不可用線性公式。
func (it *Items) EquipSlot(code int) int {
	b4 := it.Category(code)
	if b4 < 0x20 {
		return -1
	}
	switch b4 >> 5 {
	case 1:
		return 0
	case 2:
		return 1
	case 3:
		return 3
	case 4:
		return 2
	default:
		return -1
	}
}

// CanEquip 回職業 cls(0..7)可否裝備(b6 bitmask;0x80=勇者專用)。移植 dq3_item_can_equip。
func (it *Items) CanEquip(code, cls int) bool {
	if code < 0 || code >= ItemCount || cls < 0 || cls > 7 {
		return false
	}
	return it.raw[code*itemStride+6]&(0x80>>uint(cls)) != 0
}
