package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRecruitmentSelectionLoadPreservesRejectedAndClearsAcceptedUI(t *testing.T) {
	g, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "selection.json"))
	g.roster = []*Member{newMember([]int{0}, 1, 0, 0)}
	g.recruit.open()
	// 元件 fixture 仍在標題；直接消費 modal 輸入，正常玩家路徑另由201包對拍驗收。
	for i := 0; i < 5000 && g.recruit.stage == rcGreeting; i++ {
		in := InputState{DirHeld: -1, DirEdge: -1}
		if g.recruit.dialogue.waitingForConfirm() {
			in.Confirm = true
		}
		g.recruitInput(in)
	}
	g.recruitInput(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, &g.recruit)
	if g.recruit.stage != rcJoin {
		t.Fatal("selection did not open")
	}
	if e = g.Save(); e != nil {
		t.Fatal(e)
	}
	before, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	flow, contract, raster, rng := g.recruit.dialogue.retained, g.recruit.selection, g.recruit.raster, g.prng
	bad := g.snapshot()
	bad.PackSchema = "invalid"
	raw, e := encodeSave(bad)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(savePath(), raw, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e == nil {
		t.Fatal("bad Load accepted")
	}
	if !g.recruit.active || g.recruit.stage != rcJoin || g.recruit.dialogue.retained != flow || g.prng != rng || len(g.roster) != 1 || len(g.companions) != 0 {
		t.Fatal("bad Load changed selection or member")
	}
	if e = os.WriteFile(savePath(), before, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	expected, e := decodeSave(before)
	if e != nil {
		t.Fatal(e)
	}
	loaded := g.snapshot()
	if !reflect.DeepEqual(expected.Roster, loaded.Roster) {
		t.Fatalf("valid Load changed roster: before=%+v after=%+v", expected.Roster, loaded.Roster)
	}
	if expected.HeroGold != loaded.HeroGold || !reflect.DeepEqual(expected.StoryBits, loaded.StoryBits) || g.prng != rng || g.recruit.active || g.recruit.dialogue.retained != nil || g.recruit.selection != contract || g.recruit.raster != raster {
		t.Fatal("valid Load must preserve roster and clear only UI")
	}
}

func TestRecruitmentSelectionDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_SELECTION_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem selection oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_SELECTION_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit selection receipt directory required")
	}
	runRegistrationNormalComparison(t, dir, out, true, true, true)
}

func TestRecruitmentContinueDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_YES_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem continue-Yes oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_YES_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit continue-Yes receipt directory required")
	}
	runRegistrationNormalComparison(t, dir, out, true, true, true, true)
}

func drainRecruitmentSelectionText(t *testing.T, rc *Recruit) {
	t.Helper()
	for i := 0; i < 5000 && rc.stage == rcText; i++ {
		in := InputState{DirHeld: -1, DirEdge: -1}
		if rc.dialogue.waitingForConfirm() {
			in.Confirm = true
		}
		rc.selectionInput(in)
	}
	if rc.stage == rcText {
		t.Fatal("selection text did not settle")
	}
}

func TestRecruitmentSelectionCancelNoReturnsWithoutTransaction(t *testing.T) {
	m := newMember([]int{0}, 1, 0, 0)
	g := &Game{roster: []*Member{m}}
	configureRecruitmentEntryFixture(t, &g.recruit)
	rc := &g.recruit
	rc.active, rc.stage = true, rcMenu
	rng := g.prng
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcText || rc.afterText != rcJoin {
		t.Fatal("selection prompt missing")
	}
	drainRecruitmentSelectionText(t, rc)
	g.recruitInput(InputState{DirEdge: -1, Cancel: true})
	if rc.stage != rcText || rc.afterText != rcAgain {
		t.Fatal("Esc must append continue question")
	}
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcAgain || rc.cursor != 0 {
		t.Fatal("continue question initial Yes")
	}
	g.recruitInput(InputState{DirEdge: 3})
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcFinalWait || !rc.active {
		t.Fatal("farewell requires separate confirmation")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.active || g.prng != rng || !reflect.DeepEqual(g.roster, []*Member{m}) || len(g.companions) != 0 {
		t.Fatal("cancel changed transaction or failed to return")
	}
}

func TestRecruitmentSelectionContinueYesReplaysQuestion(t *testing.T) {
	g := &Game{roster: []*Member{newMember([]int{0}, 1, 0, 0)}}
	configureRecruitmentEntryFixture(t, &g.recruit)
	rc := &g.recruit
	rc.active, rc.stage = true, rcJoin
	g.recruitInput(InputState{DirEdge: -1, Cancel: true})
	drainRecruitmentSelectionText(t, rc)
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcText || rc.afterText != rcMenu {
		t.Fatal("Yes must replay parent question before menu")
	}
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcMenu || rc.cursor != 0 || len(g.roster) != 1 || len(g.companions) != 0 {
		t.Fatal("Yes changed member")
	}
}

func TestRecruitmentSelectionAgainEscRetainsChoice(t *testing.T) {
	for _, choice := range []int{0, 1} {
		t.Run(string(rune('0'+choice)), func(t *testing.T) {
			g := &Game{roster: []*Member{newMember([]int{0}, 1, 0, 0)}}
			configureRecruitmentEntryFixture(t, &g.recruit)
			rc := &g.recruit
			rc.active, rc.stage, rc.cursor = true, rcAgain, choice
			g.recruitInput(InputState{DirEdge: -1, Cancel: true})
			want := rcMenu
			if choice == 1 {
				want = rcFinalWait
			}
			if rc.stage != rcText || rc.afterText != want || len(g.roster) != 1 || len(g.companions) != 0 {
				t.Fatal("Esc must preserve current native Yes/No choice")
			}
		})
	}
}
