package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func loadTestItems(t *testing.T) *dq3data.Items {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "ITEM.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	items, err := dq3data.OpenItems(raw)
	if err != nil {
		t.Fatal(err)
	}
	return items
}

func loadTestPack(t *testing.T) *gamepack.Pack {
	t.Helper()
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	return pack
}

func TestEquipSelectedTransfersInventoryAndSupportsItemZero(t *testing.T) {
	g := &Game{
		pack:        loadTestPack(t),
		panelActor:  0,
		panelCursor: 0,
		inventory:   []int{0x00},
		equip:       [4]int{-1, 0x1e, -1, -1},
	}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if g.equip[0] != 0x00 || len(g.inventory) != 0 {
		t.Fatalf("檜木棒 0x00 應可由背包轉入武器槽：equip=%v inv=%v", g.equip, g.inventory)
	}
}

func TestEquipSelectedCompanionReturnsOldEquipment(t *testing.T) {
	m := newMember([]int{107, 144}, 1, 0, 0)
	m.Weapon = 0x00
	g := &Game{
		pack:        loadTestPack(t),
		panelActor:  1,
		panelCursor: 0,
		equip:       [4]int{-1, 0x1e, -1, -1},
		companions:  []*Member{m},
	}
	m.Inventory = []int{0x03, 0x41}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if m.Weapon != 0x03 {
		t.Fatalf("戰士應換上銅劍：weapon=%#x", m.Weapon)
	}
	if len(m.Inventory) != 2 || m.Inventory[0] != 0x41 || m.Inventory[1] != 0x00 {
		t.Fatalf("新裝備應移出、舊檜木棒應回該同伴物品欄：%v", m.Inventory)
	}
	if len(g.inventory) != 0 {
		t.Fatalf("同伴換裝不得污染主角物品欄：%v", g.inventory)
	}
}

func TestEquipSelectedCannotReplaceCursedSameSlot(t *testing.T) {
	g := &Game{
		pack:        loadTestPack(t),
		panelActor:  0,
		panelCursor: 0,
		inventory:   []int{0x03},
		equip:       [4]int{0x1b, 0x1e, -1, -1},
	}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if g.equip[0] != 0x1b || len(g.inventory) != 1 || g.inventory[0] != 0x03 {
		t.Fatalf("cursed weapon must block same-slot replacement: equip=%v inv=%v",
			g.equip, g.inventory)
	}
}

func TestLegacySaveMigratesZeroEmptySlots(t *testing.T) {
	eq := [4]int{0, 0x1e, 0, 0}
	migrateLegacyEmptyEquipment(&eq)
	if eq != [4]int{-1, 0x1e, -1, -1} {
		t.Fatalf("舊主角空槽 migration 錯：%v", eq)
	}
	m := &Member{Class: 1, Armor: 0x1e}
	migrateLegacyMemberEquipment(m)
	if m.Weapon != -1 || m.Armor != 0x1e || m.Shield != -1 || m.Head != -1 {
		t.Fatalf("舊同伴空槽 migration 錯：%#v", m)
	}
}
