package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func choiceItemExchangeTestGame(t *testing.T) (*Game, gamepack.ChoiceItemExchangeEvent, *npcInst) {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	events := g.pack.ChoiceItemExchangeEvents()
	if len(events) != 1 {
		t.Fatalf("choice item exchange event count=%d, want 1", len(events))
	}
	event := events[0]
	g.inTown, g.curCty = true, event.NPC.CTYRaw
	g.cur = &Scene{sec: event.NPC.Section}
	npc := &npcInst{x: event.NPC.Tile.X, y: event.NPC.Tile.Y, b4: event.NPC.HandlerRaw}
	return g, event, npc
}

func assertChoiceExchangeText(t *testing.T, g *Game, id string) {
	t.Helper()
	want, ok := g.pack.TextGlyphCodes(id)
	if !ok || !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, want) {
		t.Fatalf("對話 %s 未由 pack 開啟：ok=%v open=%v", id, ok, g.dlg.open)
	}
}

func TestChoiceItemExchangeNoItemKnowledgeBranchesDoNotMutateState(t *testing.T) {
	g, event, npc := choiceItemExchangeTestGame(t)
	g.setStoryFlag(event.AvailableFlagRaw, true)
	g.inventory = []int{7, 8}
	beforeState, beforeX, beforeY := g.worldState, g.overPx, g.overPy

	if !g.talkChoiceItemExchange(npc) || g.choiceItemExchangeStage != choiceItemExchangeIntroduction {
		t.Fatal("缺道具時未進原版介紹記錄")
	}
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Introduction)
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	if !g.choiceItemExchangeChoosing() {
		t.Fatal("介紹記錄關閉後未進知情 Yes/No")
	}
	g.choiceItemExchangeChoiceInput(InputState{Confirm: true, DirEdge: -1})
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Interested)
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	if g.choiceItemExchangeStage != choiceItemExchangeIdle ||
		!reflect.DeepEqual(g.inventory, []int{7, 8}) || g.worldState != beforeState ||
		g.overPx != beforeX || g.overPy != beforeY || !g.storyFlag(event.AvailableFlagRaw) {
		t.Fatal("缺道具詢問分支不得修改持久狀態")
	}

	if !g.talkChoiceItemExchange(npc) {
		t.Fatal("缺道具重談未被事件接管")
	}
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	g.choiceItemExchangeChoiceInput(InputState{DirEdge: 0})
	g.choiceItemExchangeChoiceInput(InputState{Confirm: true, DirEdge: -1})
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Hint)
}

func TestChoiceItemExchangeRejectThenSuccessMatchesOriginalTransaction(t *testing.T) {
	g, event, npc := choiceItemExchangeTestGame(t)
	g.setStoryFlag(event.AvailableFlagRaw, true)
	g.inventory = []int{7, event.RequiredItemRawID, 8}

	if !g.talkChoiceItemExchange(npc) || g.choiceItemExchangeStage != choiceItemExchangeOffer {
		t.Fatal("持有變身杖時未進交換提議")
	}
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Offer)
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	g.choiceItemExchangeChoiceInput(InputState{DirEdge: 0})
	g.choiceItemExchangeChoiceInput(InputState{Confirm: true, DirEdge: -1})
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Reject)
	if !g.hasItem(event.RequiredItemRawID) || g.hasItem(event.GrantedItemRawID) ||
		!g.storyFlag(event.AvailableFlagRaw) || g.worldState != 0 {
		t.Fatal("拒絕交換不得消耗道具或修改世界狀態")
	}
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()

	if !g.talkChoiceItemExchange(npc) {
		t.Fatal("拒絕後重談未被事件接管")
	}
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	g.choiceItemExchangeChoiceInput(InputState{Confirm: true, DirEdge: -1})
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.Success)
	wantInventory := []int{7, event.GrantedItemRawID, 8}
	if !reflect.DeepEqual(g.inventory, wantInventory) ||
		g.overPx != event.SuccessWorldPosition.X || g.overPy != event.SuccessWorldPosition.Y ||
		g.layer != event.SuccessWorldPosition.Layer ||
		g.worldState&uint16(event.SetWorldStateMaskRaw) == 0 ||
		g.storyFlag(event.AvailableFlagRaw) {
		t.Fatalf("成功交易不符原版：inventory=%v pos=(%d,%d,%d) world=%#x flag=%v",
			g.inventory, g.overPx, g.overPy, g.layer, g.worldState,
			g.storyFlag(event.AvailableFlagRaw))
	}
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.After)
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	if g.choiceItemExchangeStage != choiceItemExchangeIdle {
		t.Fatal("成功後話關閉未結束事件")
	}
	if !g.talkChoiceItemExchange(npc) {
		t.Fatal("完成後重談未被事件接管")
	}
	assertChoiceExchangeText(t, g, event.DialogueTextIDs.After)
}
