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
	g := &Game{pack: loadTestPack(t), panelActor: 0, panelCursor: 0, items: testItemStore([]int{0x00}, [4]int{-1, 0x1e, -1, -1})}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if g.items.Equipment()[0] != 0x00 || len(testInventory(g.items)) != 0 {
		t.Fatalf("檜木棒 0x00 應可由背包轉入武器槽：equip=%v inv=%v", g.items.Equipment(), testInventory(g.items))
	}
}

func TestEquipSelectedCompanionReturnsOldEquipment(t *testing.T) {
	m := newMember([]int{107, 144}, 1, 0, 0)
	setTestGear(&m.Items, 0, 0x00)
	g := &Game{pack: loadTestPack(t), panelActor: 1, panelCursor: 0, companions: []*Member{m}, items: testItemStore(nil, [4]int{-1, 0x1e, -1, -1})}
	setTestInventory(&m.Items, []int{0x03, 0x41})
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if m.Items.Equipment()[0] != 0x03 {
		t.Fatalf("戰士應換上銅劍：weapon=%#x", m.Items.Equipment()[0])
	}
	if len(testInventory(m.Items)) != 2 || testInventory(m.Items)[0] != 0x41 || testInventory(m.Items)[1] != 0x00 {
		t.Fatalf("新裝備應移出、舊檜木棒應回該同伴物品欄：%v", testInventory(m.Items))
	}
	if len(testInventory(g.items)) != 0 {
		t.Fatalf("同伴換裝不得污染主角物品欄：%v", testInventory(g.items))
	}
}

func TestEquipSelectedCannotReplaceCursedSameSlot(t *testing.T) {
	g := &Game{pack: loadTestPack(t), panelActor: 0, panelCursor: 0, items: testItemStore([]int{0x03}, [4]int{0x1b, 0x1e, -1, -1})}
	g.shop.items = loadTestItems(t)
	g.equipSelected()
	if g.items.Equipment()[0] != 0x1b || len(testInventory(g.items)) != 1 || testInventory(g.items)[0] != 0x03 {
		t.Fatalf("cursed weapon must block same-slot replacement: equip=%v inv=%v",
			g.items.Equipment(), testInventory(g.items))
	}
}

func TestLegacyEquipmentSaveIsRejected(t *testing.T) {
	for _, raw := range []string{`{"eq":[0,30,0,0],"inv":[]}`, `{"save_version":1}`, `{}`} {
		if _, err := decodeSave([]byte(raw)); err == nil {
			t.Fatalf("old save accepted: %s", raw)
		}
	}
}
