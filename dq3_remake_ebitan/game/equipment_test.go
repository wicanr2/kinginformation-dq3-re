package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
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

func TestEquipSelectedTransfersInventoryAndSupportsItemZero(t *testing.T) {
	g := &Game{
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
		panelActor:  1,
		panelCursor: 0,
		inventory:   []int{0x03, 0x41},
		equip:       [4]int{-1, 0x1e, -1, -1},
		companions:  []*Member{m},
	}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if m.Weapon != 0x03 {
		t.Fatalf("戰士應換上銅劍：weapon=%#x", m.Weapon)
	}
	if len(g.inventory) != 2 || g.inventory[0] != 0x41 || g.inventory[1] != 0x00 {
		t.Fatalf("新裝備應移出、舊檜木棒應回背包：%v", g.inventory)
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
