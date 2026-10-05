package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestMairaSpecialShopOriginalEXEBytes(t *testing.T) {
	exe, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "DQ3.EXE"))
	if err != nil {
		t.Skipf("original DQ3.EXE unavailable: %v", err)
	}
	for off, want := range map[int][]byte{
		0x7685: {0xc7, 0x06, 0x62, 0x0b, 0x01, 0x00},                   // special mode set
		0x8855: {0x83, 0x3e, 0x62, 0x0b, 0x01},                         // gated stock branch
		0x8a28: {0x83, 0x3e, 0x62, 0x0b, 0x01, 0x75, 0x10},             // one-off buy tail
		0x8bb5: {0x83, 0x3e, 0x62, 0x0b, 0x01, 0x75, 0x15, 0x83, 0x3e}, // special sell tail
	} {
		if off < 0 || off+len(want) > len(exe) || !bytes.Equal(exe[off:off+len(want)], want) {
			t.Fatalf("DQ3.EXE file %#x special-shop bytes drifted", off)
		}
	}
}

func mairaSpecialShopTestGame(t *testing.T) (*Game, int, int) {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	event, ok := g.pack.SpecialShopEvent("dq3:event.maira_kings_sword_shop")
	if !ok {
		t.Fatal("missing Maira special shop event")
	}
	// 索瑪前的愛列夫加特使用黑暗／夜間表；CTY81 sec1 因此含原始
	// `(4,3)` sub2 handler72 NPC。
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.NPC.CTYRaw, mapBlkNum[event.NPC.CTYRaw], event.NPC.Section, 2, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	idx := sc.npcAt(event.NPC.Tile.X, event.NPC.Tile.Y)
	if idx < 0 || (sc.npcs[idx].ctrl>>3)&7 != 2 || sc.npcs[idx].b4 != event.NPC.HandlerRaw {
		t.Fatalf("CTY81 sec1 special shop NPC mismatch: idx=%d npcs=%+v", idx, sc.npcs)
	}
	g.showTitle, g.inTown, g.curCty = false, true, event.NPC.CTYRaw
	g.town, g.cur, g.dlg.tx = sc, sc, sc.dlgText
	g.px, g.py, g.facing = event.NPC.Tile.X, event.NPC.Tile.Y+1, 1
	setTestInventory(&g.items, []int{event.UnlockSellItemRawID})
	g.heroGold = 20000
	return g, event.UnlockFlagRaw, event.ConditionalItemFlagRaw
}

func TestMairaSpecialShopFormalInputSellUnlockBuyAndSaveRoundTrip(t *testing.T) {
	g, unlockFlag, stockFlag := mairaSpecialShopTestGame(t)
	event, _ := g.pack.SpecialShopEvent("dq3:event.maira_kings_sword_shop")
	if !g.storyFlag(unlockFlag) || !g.storyFlag(stockFlag) {
		t.Fatal("original initial story bits must keep both special-shop flags set")
	}

	// A 開命令窗，再按 A 選擇預設的「交談」。
	if err := g.step(InputState{Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if err := g.step(InputState{Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if !g.shop.active || g.shop.kind != "special" {
		t.Fatalf("handler72 did not open special shop: active=%v kind=%q", g.shop.active, g.shop.kind)
	}
	for _, code := range g.shop.codes {
		if code == event.ConditionalItemRawID {
			t.Fatal("King's Sword must remain hidden before Orichalcum sale")
		}
	}

	// B 進入賣出，再以 A 選勇者、選道具並確認交易。
	for _, in := range []InputState{{Cancel: true}, {Confirm: true}, {Confirm: true}, {Confirm: true}} {
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	if g.storyFlag(unlockFlag) || len(testInventory(g.items)) != 0 || g.heroGold != 42500 {
		t.Fatalf("Orichalcum sale transaction mismatch: flag=%v inv=%v gold=%d",
			g.storyFlag(unlockFlag), testInventory(g.items), g.heroGold)
	}
	// B 由物品清單回到選人，再按 B 關店；勇者仍可能有已裝備項目，不能把
	// 賣掉背包中的金屬誤判成整份可售清單已空。
	for range 2 {
		if err := g.step(InputState{Cancel: true}); err != nil {
			t.Fatal(err)
		}
	}

	// 由同一玩家入口重開；flag0x135 仍 set 時，隱藏的第五項已出現。
	for range 2 {
		if err := g.step(InputState{Confirm: true}); err != nil {
			t.Fatal(err)
		}
	}
	if !g.shop.active || len(g.shop.codes) == 0 ||
		g.shop.codes[len(g.shop.codes)-1] != event.ConditionalItemRawID {
		t.Fatalf("王者之劍賣出後未上架：active=%v kind=%q stage=%d flags=%v/%v codes=%v",
			g.shop.active, g.shop.kind, g.shop.stage, g.storyFlag(unlockFlag),
			g.storyFlag(stockFlag), g.shop.codes)
	}
	for g.shop.cursor != len(g.shop.codes)-1 {
		if err := g.step(InputState{DirEdge: 0}); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.step(InputState{Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if err := g.step(InputState{Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if g.storyFlag(stockFlag) || g.heroGold != 7500 || len(testInventory(g.items)) != 1 ||
		testInventory(g.items)[0] != event.ConditionalItemRawID {
		t.Fatalf("King's Sword purchase mismatch: flag=%v inv=%v gold=%d",
			g.storyFlag(stockFlag), testInventory(g.items), g.heroGold)
	}
	for _, code := range g.shop.codes {
		if code == event.ConditionalItemRawID {
			t.Fatal("purchased one-off stock remained buyable in the same session")
		}
	}

	saved := g.snapshot()
	g2, _, _ := mairaSpecialShopTestGame(t)
	if err := g2.restore(saved); err != nil {
		t.Fatal(err)
	}
	if g2.storyFlag(unlockFlag) || g2.storyFlag(stockFlag) {
		t.Fatalf("special-shop flags did not survive save round-trip: unlock=%v stock=%v",
			g2.storyFlag(unlockFlag), g2.storyFlag(stockFlag))
	}
}
