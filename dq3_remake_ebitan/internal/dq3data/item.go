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

// EquipSlot 回裝備部位(0x2_武器/0x4_鎧/0x6_盾/0x8_兜;非裝備 -1)。移植 dq3_item_equip_slot。
func (it *Items) EquipSlot(code int) int {
	b4 := it.Category(code)
	if b4 < 0x20 {
		return -1
	}
	return (b4 >> 5) - 1
}

// CanEquip 回職業 cls(0..7)可否裝備(b6 bitmask;0x80=勇者專用)。移植 dq3_item_can_equip。
func (it *Items) CanEquip(code, cls int) bool {
	if code < 0 || code >= ItemCount || cls < 0 || cls > 7 {
		return false
	}
	return it.raw[code*itemStride+6]&(0x80>>uint(cls)) != 0
}
