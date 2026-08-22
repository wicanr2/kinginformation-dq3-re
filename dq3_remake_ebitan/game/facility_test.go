package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/itemuse"
)

func TestShopPurchaseUsesSelectedPersonalEightSlotInventory(t *testing.T) {
	companion := newMember([]int{1}, 1, 0, 0)
	g := &Game{
		pack:       loadTestPack(t),
		heroGold:   1000,
		inventory:  []int{1, 2, 3, 4},
		equip:      [4]int{5, 6, 7, 8},
		companions: []*Member{companion},
	}
	g.shop.items = loadTestItems(t)
	g.shop.active, g.shop.targeting, g.shop.pendingCode = true, true, itemuse.ItemHerb
	beforeGold := g.heroGold
	if g.purchaseShopItem(0) || g.heroGold != beforeGold || len(g.inventory) != 4 ||
		!g.shop.targeting {
		t.Fatalf("勇者八格已滿不得購買或扣款：gold=%d inv=%v targeting=%v",
			g.heroGold, g.inventory, g.shop.targeting)
	}
	if !g.purchaseShopItem(1) || g.heroGold != beforeGold-g.shop.items.Price(itemuse.ItemHerb) ||
		len(companion.Inventory) != 1 || companion.Inventory[0] != itemuse.ItemHerb ||
		g.shop.targeting {
		t.Fatalf("改選同伴後交易錯：gold=%d companion=%v targeting=%v",
			g.heroGold, companion.Inventory, g.shop.targeting)
	}
}

func TestShopEightSlotWriterMatchesOriginal(t *testing.T) {
	exe, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for off, want := range map[int][]byte{
		0x8ab7: {0x8b, 0x1e, 0x22, 0x07, 0x4b, 0xd1, 0xe3},
		0x8ac2: {0x83, 0xc6, 0x3a, 0xb9, 0x08, 0x00},
		0x8ac8: {0x8b, 0x04, 0x3d, 0xff, 0x00},
		0x8ad3: {0xb0, 0xff, 0xc3, 0xa1, 0x91, 0x25, 0x48, 0x89, 0x04},
		0x8f0d: {0x8f, 0x06, 0x22, 0x07, 0xe8, 0xa3, 0xfb, 0x3c, 0xff, 0x74, 0x1f},
	} {
		if got := exe[off : off+len(want)]; !bytes.Equal(got, want) {
			t.Fatalf("DQ3.EXE file %#x=%x want %x", off, got, want)
		}
	}
}

func TestCTY43ChurchBlockHasNoFacilityNPCConsumer(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	for _, phase := range []int{0, 2} {
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
			43, mapBlkNum[43], 0, phase, nil)
		if err != nil {
			t.Fatalf("CTY43 sec0 phase%d: %v", phase, err)
		}
		hasInn := false
		for _, npc := range sc.npcs {
			if (npc.ctrl>>3)&7 < 3 {
				continue
			}
			switch npc.b4 {
			case 0:
				hasInn = true
			case 2:
				t.Fatalf("CTY43 phase%d 不得把孤立 church block index2 誤接成 facility NPC：(%d,%d)",
					phase, npc.x, npc.y)
			}
		}
		if !hasInn {
			t.Fatalf("CTY43 phase%d 缺原始 index0 旅店 NPC consumer", phase)
		}
	}
}

func TestInnChargesAndHealsLivingPartyOnly(t *testing.T) {
	living := newMember([]int{1}, 1, 0, 0)
	living.CurHP, living.CurMP = 1, 0
	dead := newMember([]int{2}, 3, 0, 0)
	dead.CurHP, dead.CurMP = 0, 0
	g := &Game{
		curCty:     0,
		heroGold:   20,
		heroHP:     1,
		heroMP:     0,
		heroInit:   true,
		companions: []*Member{living, dead},
	}

	// CTY00 facility 0 的單價是 2；原版只計仍存活的勇者與一名同伴。
	g.openFacility(0)

	_, heroMaxHP, _, _, _ := g.heroStats()
	if g.heroGold != 16 {
		t.Fatalf("旅店應收單價2×存活2人=4，gold=%d", g.heroGold)
	}
	if g.heroHP != heroMaxHP || g.heroMP != g.heroMaxMP() {
		t.Fatalf("勇者未補滿：HP=%d/%d MP=%d/%d",
			g.heroHP, heroMaxHP, g.heroMP, g.heroMaxMP())
	}
	if living.CurHP != living.MaxHP() || living.CurMP != living.MaxMP() {
		t.Fatalf("存活同伴未補滿：HP=%d/%d MP=%d/%d",
			living.CurHP, living.MaxHP(), living.CurMP, living.MaxMP())
	}
	if dead.CurHP != 0 || dead.CurMP != 0 {
		t.Fatalf("旅店不可復活死亡同伴：HP=%d MP=%d", dead.CurHP, dead.CurMP)
	}
}

func TestInnInsufficientGoldDoesNotPartiallyHeal(t *testing.T) {
	companion := newMember([]int{1}, 1, 0, 0)
	companion.CurHP, companion.CurMP = 1, 0
	g := &Game{
		curCty:     0,
		heroGold:   3,
		heroHP:     1,
		heroMP:     0,
		heroInit:   true,
		companions: []*Member{companion},
	}

	g.openFacility(0) // 需要 2×2=4。

	if g.heroGold != 3 || g.heroHP != 1 || g.heroMP != 0 ||
		companion.CurHP != 1 || companion.CurMP != 0 {
		t.Fatalf("金幣不足不可扣款或部分恢復：gold=%d hero=%d/%d companion=%d/%d",
			g.heroGold, g.heroHP, g.heroMP, companion.CurHP, companion.CurMP)
	}
}
