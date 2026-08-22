package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestPackTreasureGrantUsesFirstPartyInventoryWithSpace(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	g.companions = []*Member{m}
	fillActorInventoryToCapacity(t, g, 0, 0x41)
	treasure := gamepack.QuestTreasureSelector{ItemRawID: 0x74, PresentFlag: 0x91}
	g.setStoryFlag(treasure.PresentFlag, true)

	if !g.collectPackTreasure(treasure) {
		t.Fatal("pack treasure 應由共用 writer 接管")
	}
	if len(m.Inventory) != 1 || m.Inventory[0] != treasure.ItemRawID ||
		g.storyFlag(treasure.PresentFlag) {
		t.Fatalf("寶箱未寫入第一個有空位的同伴：member=%v present=%v",
			m.Inventory, g.storyFlag(treasure.PresentFlag))
	}
}

func TestPackTreasureGrantAllPartyFullFailsWithoutClearingFlag(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	g.companions = []*Member{m}
	fillActorInventoryToCapacity(t, g, 0, 0x41)
	fillActorInventoryToCapacity(t, g, 1, 0x41)
	treasure := gamepack.QuestTreasureSelector{ItemRawID: 0x74, PresentFlag: 0x91}
	g.setStoryFlag(treasure.PresentFlag, true)

	if !g.collectPackTreasure(treasure) {
		t.Fatal("滿格寶箱仍應消費互動，但不可冒稱取得")
	}
	if !g.storyFlag(treasure.PresentFlag) || g.hasPartyItem(treasure.ItemRawID) {
		t.Fatalf("全隊滿格不得清 present flag 或寫物品：present=%v",
			g.storyFlag(treasure.PresentFlag))
	}
}

func TestRemovePartyItemsConsumesOnlyOneMatchingOwnerSlot(t *testing.T) {
	g := &Game{
		inventory: []int{0x55},
		companions: []*Member{{Inventory: []int{0x55, 0x41}}},
	}
	g.removePartyItems(0x55, 1)
	if len(g.inventory) != 0 || len(g.companions[0].Inventory) != 2 {
		t.Fatalf("一次消耗不得再刪同伴同 id：hero=%v member=%v",
			g.inventory, g.companions[0].Inventory)
	}
	g.removePartyItems(0x55, 1)
	if len(g.companions[0].Inventory) != 1 || g.companions[0].Inventory[0] != 0x41 {
		t.Fatalf("勇者沒有時才應消耗同伴：%v", g.companions[0].Inventory)
	}
}

func TestCommonPartyGrantCallersMatchOriginalBytes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []struct {
		linear int
		hex    []byte
	}{
		{0x10272, []byte{0xc7, 0x06, 0x93, 0x25, 0x00, 0x00, 0xe8, 0x1d, 0x66}},
		{0x1029f, []byte{0xc7, 0x06, 0x93, 0x25, 0x1f, 0x00, 0xe8, 0xa6, 0x65}},
		{0x15e0d, []byte{0xc7, 0x06, 0x93, 0x25, 0x65, 0x00, 0xe8, 0x38, 0x0a}},
		{0x16856, []byte{0x8a, 0x0e, 0x77, 0x50}},
		{0x1686e, []byte{0x81, 0x3c, 0xff, 0x00}},
		{0x16895, []byte{0x89, 0x04}},
	} {
		file := want.linear - 0xec90
		if file < 0 || file+len(want.hex) > len(raw) ||
			!bytes.Equal(raw[file:file+len(want.hex)], want.hex) {
			t.Fatalf("DQ3.EXE linear %#x / file %#x bytes mismatch", want.linear, file)
		}
	}
}
