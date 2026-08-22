package game

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
)

func TestFieldItemCompanionDropUsesSelectedOwnerSlot(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	m.Inventory = []int{0x41, itemuse.ItemHerb, 0x42}
	g.companions = []*Member{m}
	g.inventory = []int{itemuse.ItemHerb}
	g.panelActor, g.itemSelected, g.panelCursor = 1, 1, 1

	if !g.dropSelectedItem() {
		t.Fatal("同伴選定欄位應可丟棄")
	}
	if !reflect.DeepEqual(m.Inventory, []int{0x41, 0x42}) ||
		!reflect.DeepEqual(g.inventory, []int{itemuse.ItemHerb}) {
		t.Fatalf("丟棄必須只改選定 owner：hero=%v member=%v", g.inventory, m.Inventory)
	}
}

func TestFieldItemCompanionUseConsumesExactOwnerSlot(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	m.Inventory = []int{0x41, itemuse.ItemHolyWater, 0x42}
	g.companions = []*Member{m}
	g.inventory = []int{itemuse.ItemHolyWater}
	g.panel, g.panelActor, g.panelCursor = panelItem, 1, 1

	g.useSelectedItem()
	if g.repel != itemuse.HolySteps || !reflect.DeepEqual(m.Inventory, []int{0x41, 0x42}) ||
		!reflect.DeepEqual(g.inventory, []int{itemuse.ItemHolyWater}) {
		t.Fatalf("同伴道具使用 ownership 錯：repel=%d hero=%v member=%v",
			g.repel, g.inventory, m.Inventory)
	}
}

func TestFieldSupplyTraceUsesCompanionOwnedHolyWater(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	m.Inventory = []int{itemuse.ItemHolyWater}
	g.companions = []*Member{m}
	g.inventory = []int{0x55}

	traceUseInventoryItem(t, g, itemuse.ItemHolyWater)
	if g.repel != itemuse.HolySteps || len(m.Inventory) != 0 ||
		!reflect.DeepEqual(g.inventory, []int{0x55}) || g.panel != panelNone {
		t.Fatalf("正式 owner selector 未使用同伴聖水：repel=%d hero=%v member=%v panel=%d",
			g.repel, g.inventory, m.Inventory, g.panel)
	}
}

func TestFieldItemGiveSupportsWholePartyAndSameOwnerReorder(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	m.Inventory = []int{0x41, 0x42}
	g.companions = []*Member{m}
	g.inventory = []int{0x43}
	g.panelActor, g.itemSelected = 1, 0

	if !g.giveSelectedItem(0) {
		t.Fatal("同伴應可把物品交給勇者")
	}
	if !reflect.DeepEqual(m.Inventory, []int{0x42}) || !reflect.DeepEqual(g.inventory, []int{0x43, 0x41}) {
		t.Fatalf("跨 owner 給予錯：hero=%v member=%v", g.inventory, m.Inventory)
	}

	m.Inventory = []int{0x44, 0x45, 0x46}
	g.panelActor, g.itemSelected = 1, 0
	if !g.giveSelectedItem(1) || !reflect.DeepEqual(m.Inventory, []int{0x45, 0x46, 0x44}) {
		t.Fatalf("同 owner 應只重排、不複製：%v", m.Inventory)
	}
}

func TestFieldItemOwnerActionsMatchOriginalBytes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []struct {
		linear int
		hex    []byte
	}{
		{0x1373a, []byte{0xa0, 0x77, 0x50}},       // party count
		{0x13768, []byte{0x8b, 0x1e, 0x22, 0x07}}, // chosen owner
		{0x13770, []byte{0x89, 0x1e, 0x2d, 0x06}}, // owner state
		{0x138f8, []byte{0x4b}},                   // owner-1 inventory reader
		{0x13919, []byte{0x4b}},                   // owner-local item selector
		{0x13a62, []byte{0x8b, 0x1e, 0x2d, 0x06}}, // same-owner reorder
		{0x13af8, []byte{0xc7, 0x04, 0xff, 0x00}}, // exact slot = 0x00ff
	} {
		file := want.linear - 0xec90
		if file < 0 || file+len(want.hex) > len(raw) ||
			!bytes.Equal(raw[file:file+len(want.hex)], want.hex) {
			t.Fatalf("DQ3.EXE linear %#x / file %#x bytes mismatch", want.linear, file)
		}
	}
}
