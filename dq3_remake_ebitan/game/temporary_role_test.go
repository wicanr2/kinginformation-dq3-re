package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func mustTemporaryRoleEvent(t *testing.T, g *Game) gamepack.TemporaryRoleEvent {
	t.Helper()
	events := g.pack.TemporaryRoleEvents()
	if len(events) != 1 {
		t.Fatalf("temporary_role event count=%d, want 1", len(events))
	}
	return events[0]
}

func loadTemporaryRoleScene(t *testing.T, g *Game, s gamepack.ScriptedNPCSelector) *npcInst {
	t.Helper()
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		s.CTYRaw, mapBlkNum[s.CTYRaw], s.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, s.CTYRaw
	g.dlg.tx = sc.dlgText
	for i := range sc.npcs {
		if sc.npcs[i].b4 == s.HandlerRaw && (sc.npcs[i].ctrl>>3)&7 == 2 {
			return &sc.npcs[i]
		}
	}
	t.Fatalf("CTY%02d sec%d 找不到 scripted handler%d", s.CTYRaw, s.Section, s.HandlerRaw)
	return nil
}

func temporaryRoleTestGame(t *testing.T) (*Game, gamepack.TemporaryRoleEvent) {
	t.Helper()
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	return g, mustTemporaryRoleEvent(t, g)
}

func closeTemporaryRoleDialogue(g *Game) {
	g.dlg.open = false
	g.advanceTemporaryRoleDialogue()
}

func TestTemporaryRoleQuestAndForcedItemReturn(t *testing.T) {
	g, event := temporaryRoleTestGame(t)
	g.inventory = []int{0x55}
	g.setStoryFlag(event.PendingFlagRaw, true)
	king := loadTemporaryRoleScene(t, g, event.OfferNPC)

	if !g.talkTemporaryRole(king) || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.QuestGreeting)) {
		t.Fatal("首次晉見未開 game-pack quest_greeting")
	}
	closeTemporaryRoleDialogue(g)
	if !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.QuestRequest)) {
		t.Fatal("quest_greeting 結束後未接 game-pack quest_request")
	}
	if !g.storyFlag(event.PendingFlagRaw) || !g.hasItem(0x55) ||
		g.hasItem(event.RequiredItemRawID) {
		t.Fatal("任務對白不應修改旗標或道具")
	}
	closeTemporaryRoleDialogue(g)

	g.inventory = append(g.inventory, event.RequiredItemRawID)
	if !g.talkTemporaryRole(king) || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.ReturnPraise)) {
		t.Fatal("持必要道具時未進入 return_praise")
	}
	if g.hasItem(event.RequiredItemRawID) || g.storyFlag(event.PendingFlagRaw) {
		t.Fatal("原版 return transaction 應在 return_praise 前收一件道具並清 pending flag")
	}
	closeTemporaryRoleDialogue(g)
	if !reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.ForcedOffer)) {
		t.Fatal("return_praise 後未接 forced_offer")
	}
	closeTemporaryRoleDialogue(g)
	if g.temporaryRoleStage != temporaryRoleForcedChoice {
		t.Fatalf("forced_offer 後 stage=%d", g.temporaryRoleStage)
	}

	g.temporaryRoleCursor = 1
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if g.temporaryRoleStage != temporaryRoleForcedReject ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.ForcedReject)) {
		t.Fatal("第一次選否未播放 forced_reject")
	}
	closeTemporaryRoleDialogue(g)
	closeTemporaryRoleDialogue(g)
	if g.temporaryRoleStage != temporaryRoleForcedChoice {
		t.Fatal("首次還物的拒絕必須回到 forced_offer，不能離開")
	}
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if g.temporaryRoleStage != temporaryRoleAccept ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Accept)) {
		t.Fatal("選是未播放 temporary_role accept")
	}
	closeTemporaryRoleDialogue(g)
	if g.storyFlag(event.NormalRoleFlagRaw) || !g.storyFlag(event.ActiveRoleFlagRaw) ||
		g.heroRole == nil || g.temporaryRoleStage != temporaryRoleIdle {
		t.Fatalf("接受臨時身分 transaction 錯：normal=%v active=%v sprite=%v stage=%d",
			g.storyFlag(event.NormalRoleFlagRaw), g.storyFlag(event.ActiveRoleFlagRaw),
			g.heroRole != nil, g.temporaryRoleStage)
	}
	want := dq3data.LoadCharSprite(g.manBLS, event.RoleSpriteEntries.Male)
	if !reflect.DeepEqual(g.heroRole, want) {
		t.Fatal("男性臨時身分角色圖未使用 pack entry base")
	}
}

func TestTemporaryRoleRestoreTwoStageChoiceAndSaveDerivation(t *testing.T) {
	g, event := temporaryRoleTestGame(t)
	g.setStoryFlag(event.PendingFlagRaw, false)
	g.setStoryFlag(event.NormalRoleFlagRaw, false)
	g.setStoryFlag(event.ActiveRoleFlagRaw, true)
	g.heroGender = 1
	g.syncTemporaryRoleVisual()
	if !reflect.DeepEqual(g.heroRole,
		dq3data.LoadCharSprite(g.manBLS, event.RoleSpriteEntries.Female)) {
		t.Fatal("女性臨時身分角色圖未使用 pack entry base")
	}
	formerKing := loadTemporaryRoleScene(t, g, event.RestoreNPC)
	if !g.talkTemporaryRole(formerKing) ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.RestoreIntro)) {
		t.Fatal("恢復 NPC 未播放 restore_intro")
	}
	closeTemporaryRoleDialogue(g)
	g.temporaryRoleCursor = 1 // 第一層 No：表示不想繼續當王
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if !reflect.DeepEqual(g.dlg.buf,
		mustPackTextCodes(t, g, event.DialogueTextIDs.RestoreReconsider)) {
		t.Fatal("恢復第一層 No 未播放 restore_reconsider")
	}
	closeTemporaryRoleDialogue(g)
	g.temporaryRoleCursor = 1 // 第二層 No：反悔，繼續當王
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if !reflect.DeepEqual(g.dlg.buf,
		mustPackTextCodes(t, g, event.DialogueTextIDs.RestoreReject)) {
		t.Fatal("恢復第二層 No 未播放 restore_reject")
	}
	closeTemporaryRoleDialogue(g)
	if !reflect.DeepEqual(g.dlg.buf,
		mustPackTextCodes(t, g, event.DialogueTextIDs.ContinueRole)) {
		t.Fatal("restore_reject 後未接 continue_role")
	}
	closeTemporaryRoleDialogue(g)
	if !g.storyFlag(event.ActiveRoleFlagRaw) || g.heroRole == nil {
		t.Fatal("第二層 No 不得恢復冒險者")
	}

	formerKing = loadTemporaryRoleScene(t, g, event.RestoreNPC)
	if !g.talkTemporaryRole(formerKing) {
		t.Fatal("第二次談話未開始")
	}
	closeTemporaryRoleDialogue(g)
	g.temporaryRoleCursor = 1
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	closeTemporaryRoleDialogue(g)
	g.temporaryRoleChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}) // 第二層 Yes：恢復冒險者
	if !reflect.DeepEqual(g.dlg.buf,
		mustPackTextCodes(t, g, event.DialogueTextIDs.RestoreAccept)) {
		t.Fatal("恢復第二層 Yes 未播放 restore_accept")
	}
	closeTemporaryRoleDialogue(g)
	if g.storyFlag(event.ActiveRoleFlagRaw) || !g.storyFlag(event.NormalRoleFlagRaw) ||
		g.heroRole != nil {
		t.Fatalf("恢復冒險者 transaction 錯：active=%v normal=%v sprite=%v",
			g.storyFlag(event.ActiveRoleFlagRaw), g.storyFlag(event.NormalRoleFlagRaw),
			g.heroRole != nil)
	}

	s := g.snapshot()
	g.setStoryFlag(event.ActiveRoleFlagRaw, true)
	g.setStoryFlag(event.NormalRoleFlagRaw, false)
	g.syncTemporaryRoleVisual()
	active := g.snapshot()
	g.restore(active)
	if g.heroRole == nil {
		t.Fatal("save/load 後應由 active role flag 重建角色圖")
	}
	g.restore(s)
	if g.heroRole != nil {
		t.Fatal("save/load 後應由清除的 active role flag 恢復一般角色圖")
	}
}
