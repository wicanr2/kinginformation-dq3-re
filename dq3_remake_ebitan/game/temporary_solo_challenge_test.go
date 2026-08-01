package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func mustSoloChallengeEvent(t *testing.T, g *Game) gamepack.TemporarySoloChallengeEvent {
	t.Helper()
	events := g.pack.TemporarySoloChallenges()
	if len(events) != 1 {
		t.Fatalf("temporary_solo_challenge count=%d, want 1", len(events))
	}
	return events[0]
}

func closeSoloChallengeDialogue(g *Game) {
	g.dlg.open = false
	g.advanceSoloChallengeDialogue()
}

func TestTemporarySoloChallengeRejectAcceptSaveAndRestore(t *testing.T) {
	g, _ := temporaryRoleTestGame(t)
	event := mustSoloChallengeEvent(t, g)
	g.companions = []*Member{
		{Name: []int{101}, Class: 1, CurHP: 11, Inventory: []int{0x20}},
		{Name: []int{102}, Class: 2, CurHP: 12, Inventory: []int{0x21}},
		{Name: []int{103}, Class: 3, CurHP: 13, Inventory: []int{0x22}},
	}
	for _, member := range g.companions {
		member.ensureStats()
		member.syncLearnedSpells()
	}
	wantCompanions := compsToSav(g.companions)
	priest := loadTemporaryRoleScene(t, g, event.EntryNPC)

	if !g.talkTemporarySoloChallenge(priest) || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Prompt)) {
		t.Fatal("勇氣試煉入口未播放 game-pack prompt")
	}
	closeSoloChallengeDialogue(g)
	if g.soloChallengeStage != soloChallengeChoice {
		t.Fatalf("prompt 後 stage=%d, want choice", g.soloChallengeStage)
	}
	g.soloChallengeCursor = 1
	g.soloChallengeChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if !reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Reject)) ||
		g.soloChallengeActive || len(g.companions) != 3 {
		t.Fatal("拒絕試煉不應修改隊伍或啟用單人模式")
	}
	g.dlg.open = false

	priest = loadTemporaryRoleScene(t, g, event.EntryNPC)
	if !g.talkTemporarySoloChallenge(priest) {
		t.Fatal("拒絕後無法再次開始試煉")
	}
	closeSoloChallengeDialogue(g)
	g.soloChallengeChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if !reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Accept)) {
		t.Fatal("接受試煉未播放 game-pack accept")
	}
	closeSoloChallengeDialogue(g)
	if !g.soloChallengeActive || len(g.companions) != 0 ||
		!reflect.DeepEqual(compsToSav(g.soloChallengeCompanions), wantCompanions) ||
		g.inTown || g.layer != event.SoloWorldPosition.Layer ||
		g.px != event.SoloWorldPosition.X || g.py != event.SoloWorldPosition.Y {
		t.Fatalf("單人試煉 transaction 不符：active=%v comps=%d saved=%d town=%v pos=(%d,%d,L%d)",
			g.soloChallengeActive, len(g.companions), len(g.soloChallengeCompanions),
			g.inTown, g.px, g.py, g.layer)
	}
	if g.storyFlag(event.CompletedFlagRaw) {
		t.Fatal("handler37 不得捏造 completed flag writer")
	}

	saved := g.snapshot()
	g.soloChallengeActive, g.soloChallengeEventID = false, ""
	g.soloChallengeCompanions = nil
	g.companions = []*Member{{Name: []int{999}}}
	g.restore(saved)
	if !g.soloChallengeActive || len(g.companions) != 0 ||
		!reflect.DeepEqual(compsToSav(g.soloChallengeCompanions), wantCompanions) {
		t.Fatal("試煉途中 save/load 未保留離隊同伴")
	}

	returnPriest := loadTemporaryRoleScene(t, g, event.ReturnNPC)
	if !g.talkTemporarySoloChallenge(returnPriest) ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Return)) {
		t.Fatal("返回神官未播放 game-pack return")
	}
	closeSoloChallengeDialogue(g)
	d := event.ReturnTownDestination
	if g.soloChallengeActive || g.soloChallengeEventID != "" ||
		!reflect.DeepEqual(compsToSav(g.companions), wantCompanions) || len(g.soloChallengeCompanions) != 0 ||
		!g.inTown || g.curCty != d.CTYRaw || g.cur == nil || g.cur.sec != d.Section ||
		g.px != d.X || g.py != d.Y {
		t.Fatalf("返回試煉 transaction 不符：active=%v comps=%d cty=%d sec=%v pos=(%d,%d)",
			g.soloChallengeActive, len(g.companions), g.curCty, g.cur != nil && g.cur.sec == d.Section, g.px, g.py)
	}
}

func TestTemporarySoloChallengeCompletedReadDoesNotMutateFlag(t *testing.T) {
	g, _ := temporaryRoleTestGame(t)
	event := mustSoloChallengeEvent(t, g)
	g.inTown = false
	g.layer, g.px, g.py = event.SoloWorldPosition.Layer,
		event.SoloWorldPosition.X, event.SoloWorldPosition.Y
	if got := g.resolveSoloChallengeWorldEntrance(-1); got != event.ReturnNPC.CTYRaw {
		t.Fatalf("completed flag clear entrance=%d, want CTY%d", got, event.ReturnNPC.CTYRaw)
	}
	g.setStoryFlag(event.CompletedFlagRaw, true)
	if got := g.resolveSoloChallengeWorldEntrance(-1); got != event.EntryNPC.CTYRaw {
		t.Fatalf("completed flag set entrance=%d, want CTY%d", got, event.EntryNPC.CTYRaw)
	}
	g.inTown = true
	priest := loadTemporaryRoleScene(t, g, event.EntryNPC)
	if !g.talkTemporarySoloChallenge(priest) ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Completed)) {
		t.Fatal("completed flag read branch 未播放 record84")
	}
	if !g.storyFlag(event.CompletedFlagRaw) || g.soloChallengeActive || g.soloChallengeEventID != "" {
		t.Fatal("completed read branch 不得修改旗標或啟動試煉")
	}
}
