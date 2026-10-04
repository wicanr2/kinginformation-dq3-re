package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"os"
	"path/filepath"
	"testing"
)

func configureRecruitmentEntryFixture(t *testing.T, rc *Recruit) {
	t.Helper()
	p, e := gamepack.BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	if e = rc.configure(p, nil); e != nil {
		t.Fatal(e)
	}
	l, _ := p.NewGameLabels()
	rc.cursorGlyph = l.ChoiceCursor[0]
}

func TestRecruitmentEntryLoadClearsTransientOnlyAfterValidation(t *testing.T) {
	g, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "recruitment.json"))
	if e = g.Save(); e != nil {
		t.Fatal(e)
	}
	original, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	g.recruit.open()
	g.recruitInput(InputState{DirHeld: -1, DirEdge: -1})
	flow, contract, raster, r := g.recruit.dialogue.retained, g.recruit.contract, g.recruit.raster, g.prng
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
		t.Fatal("invalid Load accepted")
	}
	if !g.recruit.active || g.recruit.dialogue.retained != flow || g.prng != r {
		t.Fatal("rejected Load changed greeting")
	}
	if e = os.WriteFile(savePath(), original, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	if g.recruit.active || g.recruit.dialogue.open || g.recruit.dialogue.retained != nil || g.recruit.contract != contract || g.recruit.raster != raster {
		t.Fatal("valid Load did not clear only UI transient")
	}
}

func traceRecruitmentGreeting(t *testing.T, g *Game) {
	traceRecruitmentText(t, g)
}
func traceRecruitmentText(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 5000; i++ {
		if !g.recruit.active || (g.recruit.stage != rcGreeting && g.recruit.stage != rcText) {
			return
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		if g.recruit.dialogue.waitingForConfirm() {
			in.Confirm = true
		}
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("normal recruitment greeting did not return")
}

func traceCloseRecruitmentSelection(t *testing.T, g *Game) {
	t.Helper()
	press := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		if e := g.step(InputState{DirHeld: -1, DirEdge: -1}); e != nil {
			t.Fatal(e)
		}
	}
	press(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
	traceRecruitmentText(t, g)
	if g.recruit.stage != rcAgain || g.recruit.cursor != 0 {
		t.Fatal("selection cancel must open native continue question")
	}
	press(InputState{DirHeld: 3, DirEdge: 3})
	press(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	traceRecruitmentText(t, g)
	if g.recruit.stage != rcFinalWait {
		t.Fatal("No must wait after farewell")
	}
	press(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if g.recruit.active {
		t.Fatal("normal selection cancel did not close")
	}
}

func TestRecruitmentEntryWaitAndMenuDoNotWriteRoster(t *testing.T) {
	g := &Game{roster: []*Member{{Name: []int{0}, Class: 1}}}
	configureRecruitmentEntryFixture(t, &g.recruit)
	g.recruit.open()
	waits := 0
	for i := 0; i < 5000 && g.recruit.stage == rcGreeting; i++ {
		in := InputState{DirHeld: -1, DirEdge: -1}
		if g.recruit.dialogue.waitingForConfirm() {
			waits++
			in.Enter = true
		}
		g.recruitInput(in)
	}
	if waits != 1 || g.recruit.stage != rcMenu || len(g.roster) != 1 || len(g.companions) != 0 {
		t.Fatal("native greeting boundary or roster changed")
	}
	g.recruitInput(InputState{DirEdge: 1})
	if g.recruit.cursor != 2 {
		t.Fatal("menu up wrap")
	}
	g.recruitInput(InputState{DirEdge: 3})
	if g.recruit.cursor != 0 {
		t.Fatal("menu right wrap")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, &g.recruit)
	if g.recruit.stage != rcJoin || len(g.roster) != 1 || len(g.companions) != 0 {
		t.Fatal("enter selected menu must not move member")
	}
}

func TestRecruitmentEntryMissingContractFailsClosed(t *testing.T) {
	var rc Recruit
	rc.open()
	if rc.active {
		t.Fatal("missing contract opened")
	}
}

func TestRecruitmentEntryDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_ENTRY_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem recruitment entry oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_ENTRY_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit receipt directory required")
	}
	runRegistrationNormalComparison(t, dir, out, true, true, false)
}
