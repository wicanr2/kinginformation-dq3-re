package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func questItemChainTestGame(t *testing.T) (*Game, gamepack.QuestItemChainEvent) {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	events := g.pack.QuestItemChainEvents()
	if len(events) != 1 {
		t.Fatalf("quest item chain event count=%d, want 1", len(events))
	}
	return g, events[0]
}

func TestQuestItemChainTreasureUsesOriginalPresentFlag(t *testing.T) {
	g, event := questItemChainTestGame(t)
	q := event.Treasure
	if !g.storyFlag(q.PresentFlag) {
		t.Fatalf("原版 treasure present flag %#x 初值應為 set", q.PresentFlag)
	}
	if !g.collectQuestTreasure(q.CTYRaw, q.Section, q.TileSubID) {
		t.Fatal("game-pack treasure selector 未處理")
	}
	if g.storyFlag(q.PresentFlag) || !g.hasItem(q.ItemRawID) {
		t.Fatalf("取得寶物後 flag/item 錯：flag=%v item=%v",
			g.storyFlag(q.PresentFlag), g.hasItem(q.ItemRawID))
	}
	before := g.countItem(q.ItemRawID)
	if !g.collectQuestTreasure(q.CTYRaw, q.Section, q.TileSubID) ||
		g.countItem(q.ItemRawID) != before {
		t.Fatal("已清 present flag 的寶物不得重複取得")
	}
}

func TestQuestItemChainExchangeReplacesInventorySlot(t *testing.T) {
	g, event := questItemChainTestGame(t)
	x := event.Exchange
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		x.NPC.CTYRaw, mapBlkNum[x.NPC.CTYRaw], x.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.inTown, g.curCty, g.town, g.cur = true, x.NPC.CTYRaw, sc, sc
	var queen *npcInst
	for i := range sc.npcs {
		if g.scriptedNPCMatches(&sc.npcs[i], x.NPC) {
			queen = &sc.npcs[i]
			break
		}
	}
	if queen == nil {
		t.Fatal("找不到 game-pack exchange NPC")
	}

	if !g.talkQuestItemChain(queen) {
		t.Fatal("缺前置道具時仍應由 quest chain 處理")
	}
	before, _ := g.pack.TextGlyphCodes(x.BeforeTextID)
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, before) {
		t.Fatal("缺夢幻紅寶石未顯示 pack before text")
	}
	g.dlg.open = false

	g.inventory = []int{0x41, x.RequiredItemRawID, 0x44}
	if !g.talkQuestItemChain(queen) {
		t.Fatal("持前置道具時 exchange 未處理")
	}
	success, _ := g.pack.TextGlyphCodes(x.SuccessTextID)
	if !reflect.DeepEqual(g.inventory, []int{0x41, x.GrantedItemRawID, 0x44}) ||
		!g.dlg.open || !reflect.DeepEqual(g.dlg.buf, success) {
		t.Fatalf("原地換物／成功文字錯：inventory=%v dialogue=%v", g.inventory, g.dlg.open)
	}
}

func TestQuestItemChainLocationUseSwitchesSceneFlagsAndReloads(t *testing.T) {
	g, event := questItemChainTestGame(t)
	u := event.Use
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		u.CTYRaw, mapBlkNum[u.CTYRaw], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.inTown, g.curCty, g.town, g.cur = true, u.CTYRaw, sc, sc
	g.px, g.py = 15, 20
	if g.storyFlag(u.SetFlagsRaw[0]) || !g.storyFlag(u.ClearFlagsRaw[0]) {
		t.Fatalf("甦醒前原版旗標初值錯：set=%v clear=%v",
			g.storyFlag(u.SetFlagsRaw[0]), g.storyFlag(u.ClearFlagsRaw[0]))
	}
	sleeping := len(sc.npcs)
	g.inventory = []int{u.ItemRawID}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()

	success, _ := g.pack.TextGlyphCodes(u.SuccessTextID)
	if g.hasItem(u.ItemRawID) || !g.storyFlag(u.SetFlagsRaw[0]) ||
		g.storyFlag(u.ClearFlagsRaw[0]) || g.panel != panelNone ||
		!g.dlg.open || !reflect.DeepEqual(g.dlg.buf, success) {
		t.Fatalf("甦醒交易錯：item=%v set=%v clear=%v panel=%d dialogue=%v",
			g.hasItem(u.ItemRawID), g.storyFlag(u.SetFlagsRaw[0]),
			g.storyFlag(u.ClearFlagsRaw[0]), g.panel, g.dlg.open)
	}
	if g.cur == sc || len(g.cur.npcs) != sleeping {
		t.Fatalf("場景未依新旗標重載：same=%v npc before/after=%d/%d",
			g.cur == sc, sleeping, len(g.cur.npcs))
	}
}

func TestQuestItemChainWrongLocationDoesNotConsume(t *testing.T) {
	g, event := questItemChainTestGame(t)
	u := event.Use
	g.inTown, g.curCty = true, u.CTYRaw+1
	g.inventory = []int{u.ItemRawID}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if !g.hasItem(u.ItemRawID) || g.storyFlag(u.SetFlagsRaw[0]) ||
		!g.storyFlag(u.ClearFlagsRaw[0]) {
		t.Fatal("位置 gate 失敗不得消耗道具或改變甦醒旗標")
	}
}
