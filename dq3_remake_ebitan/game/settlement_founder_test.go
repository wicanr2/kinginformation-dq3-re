package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func testSettlementFounder(t *testing.T) (*gamepack.Pack, *gamepack.SettlementFounderEvent) {
	t.Helper()
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	event, ok := pack.SettlementFounderEvent("dq3:event.merchant_settlement_founder")
	if !ok {
		t.Fatal("missing merchant settlement founder event")
	}
	return pack, event
}

func TestSettlementFounderOriginalTwoConfirmTransaction(t *testing.T) {
	pack, event := testSettlementFounder(t)
	merchant := newMember([]int{700, 701}, event.RequiredClassRaw, 1, 0)
	merchant.Weapon, merchant.Armor, merchant.Shield, merchant.Head = 1, 30, -1, -1
	merchant.Inventory = []int{64, 65}
	other := newMember([]int{702}, 1, 0, 0)
	g := &Game{
		pack: pack, curCty: event.NPC.CTYRaw,
		cur:           &Scene{sec: event.NPC.Section},
		companions:    []*Member{merchant, other},
		sharedStorage: []int{9}, settlementFounderMember: -1,
	}
	n := &npcInst{x: event.NPC.Tile.X, y: event.NPC.Tile.Y, b4: event.NPC.HandlerRaw}
	if !g.talkSettlementFounder(n) || g.settlementFounderStage != settlementFounderIntroduction {
		t.Fatalf("正式 NPC selector 未開啟建城事件：stage=%d", g.settlementFounderStage)
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	if g.settlementFounderStage != settlementFounderFirstOffer ||
		!reflect.DeepEqual(g.dlg.varGlyph[dq3data.TxtVarEnt], merchant.Name) {
		t.Fatalf("第一層提議／商人姓名插值錯誤：stage=%d vars=%v",
			g.settlementFounderStage, g.dlg.varGlyph)
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	g.settlementFounderChoiceInput(InputState{Confirm: true, DirEdge: -1})
	if g.settlementFounderStage != settlementFounderFinalOffer {
		t.Fatalf("第一層 Yes 未進第二次確認：stage=%d", g.settlementFounderStage)
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	g.settlementFounderChoiceInput(InputState{Confirm: true, DirEdge: -1})
	if g.settlementFounderStage != settlementFounderAccepted || len(g.companions) != 2 {
		t.Fatal("成功文字關閉前不得提前移除商人")
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	if len(g.companions) != 1 || g.companions[0] != other || g.settlementFounder == nil ||
		!reflect.DeepEqual(g.settlementFounder.Name, merchant.Name) || !g.storyFlag(event.CompletionFlagRaw) ||
		!g.storyFlag(event.FounderVisibilityFlagsRaw[merchant.Gender]) {
		t.Fatalf("建城者／隊伍／旗標交易錯誤：companions=%d founder=%v flag=%v",
			len(g.companions), g.settlementFounder, g.storyFlag(event.CompletionFlagRaw))
	}
	if !reflect.DeepEqual(g.sharedStorage, []int{9, 1, 30, 64, 65}) {
		t.Fatalf("裝備與個人物品未依序移入共用預存所：%v", g.sharedStorage)
	}
	if len(g.roster) != 0 {
		t.Fatalf("建城者不得被放回酒館名冊：%v", g.roster)
	}
}

func TestSettlementFounderStopsAtFirstDeadMerchant(t *testing.T) {
	pack, event := testSettlementFounder(t)
	dead := newMember([]int{700}, event.RequiredClassRaw, 0, 0)
	dead.CurHP = 0
	alive := newMember([]int{701}, event.RequiredClassRaw, 0, 0)
	g := &Game{pack: pack, curCty: event.NPC.CTYRaw, cur: &Scene{sec: event.NPC.Section},
		companions: []*Member{dead, alive}, settlementFounderMember: -1}
	n := &npcInst{x: event.NPC.Tile.X, y: event.NPC.Tile.Y, b4: event.NPC.HandlerRaw}
	if !g.talkSettlementFounder(n) {
		t.Fatal("handler should still open its introduction")
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	if g.settlementFounderStage != settlementFounderNoEligible {
		t.Fatalf("原版遇到第一名死亡商人應直接走 no-eligible，stage=%d",
			g.settlementFounderStage)
	}
}

func TestSettlementFounderFullStorageRepeatsOriginalNoticeBeforeRemoval(t *testing.T) {
	pack, event := testSettlementFounder(t)
	merchant := newMember([]int{700}, event.RequiredClassRaw, 0, 0)
	merchant.Inventory = []int{64, 65}
	g := &Game{pack: pack, companions: []*Member{merchant},
		settlementFounderEventID: event.ID, settlementFounderMember: 0,
		settlementFounderStage: settlementFounderAccepted,
		sharedStorage:          make([]int, event.SharedStorageCapacity)}
	if !g.beginSettlementFounder(event) {
		t.Fatal("原版預存所滿仍會完成建城交易，只在每個溢出道具顯示訊息")
	}
	if g.settlementFounderStage != settlementFounderStorageFull ||
		g.settlementFounderOverflow != 2 || len(g.companions) != 1 ||
		!g.storyFlag(event.CompletionFlagRaw) {
		t.Fatalf("第一個滿倉訊息前狀態錯：stage=%d overflow=%d comps=%d flag=%v",
			g.settlementFounderStage, g.settlementFounderOverflow,
			len(g.companions), g.storyFlag(event.CompletionFlagRaw))
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	if !g.dlg.open || g.settlementFounderOverflow != 1 || len(g.companions) != 1 {
		t.Fatal("每個無法寄存的原始物品槽都應再次顯示 record309")
	}
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	if len(g.companions) != 0 || g.settlementFounder == nil ||
		g.settlementFounderStage != settlementFounderIdle {
		t.Fatalf("最後一個滿倉訊息後才應移除商人：comps=%d founder=%v stage=%d",
			len(g.companions), g.settlementFounder, g.settlementFounderStage)
	}
}

func TestSettlementFounderSaveRoundTrip(t *testing.T) {
	founder := newMember([]int{700, 701}, 6, 1, 123)
	founder.Inventory = nil
	g := &Game{settlementFounder: founder, sharedStorage: []int{1, 30, 64},
		equip: [4]int{-1, -1, -1, -1}}
	s := g.snapshot()
	var restored Game
	restored.restore(s)
	if restored.settlementFounder == nil ||
		!reflect.DeepEqual(restored.settlementFounder.Name, founder.Name) ||
		!reflect.DeepEqual(restored.sharedStorage, g.sharedStorage) {
		t.Fatalf("建城者與共用預存所存讀檔失敗：founder=%v storage=%v",
			restored.settlementFounder, restored.sharedStorage)
	}
}

func TestSettlementFounderImprisonedDialogueTracksYellowOrb(t *testing.T) {
	pack, _ := testSettlementFounder(t)
	orb, ok := pack.TreasureEvent("dq3:event.merchant_town_yellow_orb")
	if !ok {
		t.Fatal("缺商人城黃寶珠事件")
	}
	founder := newMember([]int{700, 701}, 6, 1, 0)
	g := &Game{
		pack: pack, inTown: true, curCty: 83, cur: &Scene{sec: 0}, flags: map[int]bool{},
		settlementFounder: founder,
	}
	g.setStoryFlag(orb.Treasure.PresentFlag, true)
	n := &npcInst{x: 2, y: 23, b4: 48}
	if !g.talkSettlementFounderFollowup(n) || !g.dlg.open ||
		len(g.settlementFounderFollowup) != 1 {
		t.Fatalf("黃寶珠仍在時應先顯示牢中對話並排入座位提示：queue=%v",
			g.settlementFounderFollowup)
	}
	for _, code := range []uint16{dq3data.TxtVar0, dq3data.TxtVar7} {
		if !reflect.DeepEqual(g.dlg.varGlyph[code], founder.Name) {
			t.Fatalf("控制碼 %#x 未插入建城者姓名：%v", code, g.dlg.varGlyph)
		}
	}
	g.dlg.open = false
	g.advanceSettlementFounderFollowup()
	wantHint, _ := pack.TextGlyphCodes("dq3:text.merchant_settlement.seat_hint")
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, wantHint) ||
		len(g.settlementFounderFollowup) != 0 {
		t.Fatal("record42 關閉後未依原版接續座位後方提示")
	}

	g.dlg.open = false
	if !g.collectQuestTreasure(orb.Treasure.CTYRaw, orb.Treasure.Section,
		orb.Treasure.TileSubID) || !g.hasItem(0x6a) || g.storyFlag(0x4a) {
		t.Fatalf("黃寶珠 transaction 錯：inv=%v flag=%v", g.inventory, g.storyFlag(0x4a))
	}
	if !g.talkSettlementFounderFollowup(n) || len(g.settlementFounderFollowup) != 0 {
		t.Fatal("黃寶珠取得後 handler48 仍不得排入 record43")
	}
	wantPrison, _ := pack.TextGlyphCodes("dq3:text.merchant_settlement.imprisoned")
	if !reflect.DeepEqual(g.dlg.buf, wantPrison) {
		t.Fatal("黃寶珠取得後 handler48 應只保留牢中 record42")
	}
	saved := g.snapshot()
	restored := &Game{pack: pack}
	restored.restore(saved)
	if restored.settlementFounder == nil ||
		!reflect.DeepEqual(restored.settlementFounder.Name, founder.Name) ||
		!restored.hasItem(0x6a) || restored.storyFlag(0x4a) {
		t.Fatalf("革命／黃寶珠 save-load 錯：founder=%v inv=%v flag=%v",
			restored.settlementFounder, restored.inventory, restored.storyFlag(0x4a))
	}
}
