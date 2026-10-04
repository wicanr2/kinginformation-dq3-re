package game

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func recruitmentRenameOracleHash(t *testing.T, run int) string {
	t.Helper()
	// 每份接受收據在原版全畫布與交易驗證完成後才固定身分。
	hashes := map[int]string{2: "7cf247fd19a34022a4507fb404eee6325d4654281ba507cb60baeffe409c512d", 4: "c139c899b83f915e19c6b395845f6d42adb207142ba77c699f70e1c49ae1d6dc", 5: "0d43e2bc421854b87dc0dd2c2fd177c9eca4bdfaa2a1609bfa5e4099ef437cef"}
	if hashes[run] == "" {
		t.Fatal("rename original receipt has not been accepted")
	}
	return hashes[run]
}

func TestRecruitmentRenameDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_RENAME_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem rename oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_RENAME_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit rename receipt output required")
	}
	for _, run := range []int{2, 4, 5} {
		t.Run("run-"+strconv.Itoa(run), func(t *testing.T) {
			dest := filepath.Join(out, strconv.Itoa(run))
			if e := os.Mkdir(dest, 0755); e != nil {
				t.Fatal(e)
			}
			runRegistrationNormalComparisonRoute(t, dir, dest, registrationComparisonRoute{birth: true, entry: true, selection: true, view: true, detail: true, rename: run})
		})
	}
}

func compareRecruitmentRenameWidget(t *testing.T, g *Game, s map[string]string) {
	t.Helper()
	ni := &g.recruit.renameFlow.ni
	cursor, e := strconv.Atoi(s["name_cursor"])
	if e != nil || ni.cursor != cursor {
		t.Fatalf("rename name cursor %d differs from %s", ni.cursor, s["name_cursor"])
	}
	length, e := strconv.Atoi(s["name_length"])
	if e != nil || len(ni.nameBuf) != length {
		t.Fatal("rename pending name length differs")
	}
	mode, e := strconv.Atoi(s["name_mode"])
	if e != nil || ni.nameZhu != (mode&1 != 0) || ni.functionFocus != (s["phase"] == "choice") {
		t.Fatal("rename widget mode or focus differs")
	}
	if ni.functionFocus {
		choice, e := strconv.Atoi(s["choice_cursor"])
		if e != nil || ni.functionCursor != choice-1 {
			t.Fatal("rename function cursor differs")
		}
	}
}

func recruitmentRenameGame(t *testing.T) *Game {
	g := recruitmentViewGame(t)
	g.heroName = []int{0}
	g.dlg.heroName = g.heroName
	g.recruitInput(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	g.recruitInput(InputState{DirHeld: -1, DirEdge: -1, Rename: true})
	if g.recruit.stage != rcViewClose {
		t.Fatal("first K skipped the first wait")
	}
	g.recruitInput(InputState{DirHeld: -1, DirEdge: -1, Rename: true})
	if g.recruit.stage != rcViewRename || g.recruit.renameFlow == nil {
		t.Fatal("second K did not open the name widget")
	}
	return g
}

func recruitmentRenameKeys(g *Game, scans []int) {
	for _, scan := range scans {
		in := InputState{DirHeld: -1, DirEdge: -1, AnyKeyEdge: true}
		if scan == 0x1c {
			in.Enter = true
		} else {
			in.DirEdge = map[int]int{0x50: 0, 0x48: 1, 0x4b: 2, 0x4d: 3}[scan]
		}
		g.recruitInput(in)
	}
}

func TestRecruitmentRenameTransactionAndEmptyGate(t *testing.T) {
	g := recruitmentRenameGame(t)
	before := g.snapshot()
	rng := g.prng
	// 原生功能列完成空姓名，然後正常選取不同的英數字格。
	recruitmentRenameKeys(g, []int{0x48, 0x4b, 0x1c, 0x48, 0x1c})
	if g.recruit.stage != rcViewRename || !reflect.DeepEqual(before, g.snapshot()) || g.prng != rng {
		t.Fatal("empty name wrote state or closed")
	}
	// 元件在上述空完成後仍停 raw35，重新進功能列切換英數。
	recruitmentRenameKeys(g, []int{0x1c, 0x1c, 0x4d, 0x50, 0x4d, 0x1c, 0x48, 0x4b, 0x4b, 0x1c, 0x48, 0x1c})
	expected := before
	expected.HeroName = []int{5}
	if g.recruit.stage != rcText || g.recruit.renameFlow != nil || g.recruit.viewFlow != nil || !reflect.DeepEqual(expected, g.snapshot()) || g.prng != rng || !reflect.DeepEqual(g.dlg.heroName, []int{5}) {
		t.Fatal("rename did not modify only hero name")
	}
}

func TestRecruitmentRenameCancellationPreservesState(t *testing.T) {
	g := recruitmentRenameGame(t)
	before := g.snapshot()
	rng := g.prng
	recruitmentRenameKeys(g, []int{0x48, 0x4b, 0x1c, 0x50, 0x50, 0x50, 0x1c})
	if g.recruit.renameFlow != nil || g.recruit.viewFlow != nil || !reflect.DeepEqual(before, g.snapshot()) || g.prng != rng {
		t.Fatal("cancel wrote or retained rename state")
	}
}

func TestRecruitmentRenameUnknownPartyStaysAtSecondWait(t *testing.T) {
	g := recruitmentViewGame(t)
	g.heroName = []int{0}
	g.companions = []*Member{{Name: []int{1}}}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	flow := g.recruit.viewFlow
	before := g.snapshot()
	rng := g.prng
	g.recruitInput(InputState{DirEdge: -1, Rename: true})
	if g.recruit.stage != rcViewClose || g.recruit.viewFlow != flow || g.recruit.renameFlow != nil || !reflect.DeepEqual(before, g.snapshot()) || g.prng != rng {
		t.Fatal("unknown party entered a guessed chooser")
	}
}

func TestRecruitmentRenameLoadValidatesBeforeClearing(t *testing.T) {
	g := recruitmentRenameGame(t)
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "rename.json"))
	if e := g.Save(); e != nil {
		t.Fatal(e)
	}
	valid, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	flow := g.recruit.renameFlow
	before := g.snapshot()
	rng := g.prng
	bad := before
	bad.PackSchema = "invalid"
	raw, e := encodeSave(bad)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(savePath(), raw, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e == nil || g.recruit.renameFlow != flow || g.prng != rng {
		t.Fatal("bad Load changed pending name")
	}
	if e = os.WriteFile(savePath(), valid, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	if g.recruit.active || g.recruit.renameFlow != nil || g.recruit.viewFlow != nil || !reflect.DeepEqual(g.heroName, before.HeroName) || !reflect.DeepEqual(g.snapshot().Roster, before.Roster) || g.prng != rng {
		t.Fatal("valid Load failed to clear rename UI and preserve names")
	}
}
