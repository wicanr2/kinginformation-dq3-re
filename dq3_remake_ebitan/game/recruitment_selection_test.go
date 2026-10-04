package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func TestRecruitmentEmptyViewDosgolemNormalInputComparison(t *testing.T) {
	runRecruitmentEmptyNormalInputComparison(t, "view", "issue4-empty-view-normal-r1-source-r2-receipt.json", "ff7a8abd5e93c867e5650f08a3af607feecf7013d7a781ff9a9ad2fd44eaeba7", "issue4-empty-view-normal-r1", 195, 191, 192)
}

func TestRecruitmentEmptyJoinDosgolemNormalInputComparison(t *testing.T) {
	runRecruitmentEmptyNormalInputComparison(t, "join", "issue4-empty-join-normal-r1-source-r1-receipt.json", "47d1a11585d4d3a3d5b6a04fc8923a3ab7adc305b1d7272a5d5823b73e3733cd", "issue4-empty-join-normal-r1", 193, 189, 190)
}

func TestRecruitmentEmptyLeaveDosgolemNormalInputComparison(t *testing.T) {
	runRecruitmentEmptyNormalInputComparison(t, "leave", "issue4-empty-leave-normal-r1-source-r1-receipt.json", "df05d0d4e61ff06c3ced99c9939c180ac9d8116305e31012f33c5350ba4d0ffb", "issue4-empty-leave-normal-r1", 193, 190, 190)
}

func runRecruitmentEmptyNormalInputComparison(t *testing.T, action, sourceName, sourceHash, prefix string, packets, openPacket, againPacket int) {
	t.Helper()
	envAction := map[string]string{"view": "VIEW", "join": "JOIN", "leave": "LEAVE"}[action]
	dir := os.Getenv("DQ3_RECRUIT_EMPTY_" + envAction + "_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem empty-view oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_EMPTY_" + envAction + "_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit output required")
	}
	path := filepath.Join(dir, sourceName)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("original accepted source identity differs")
	}
	var src struct{ Queued, States []map[string]string }
	if err = json.Unmarshal(raw, &src); err != nil || len(src.Queued) != packets || len(src.States) != packets {
		t.Fatal(err, "source shape")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
	}
	settle := func() {
		t.Helper()
		for n := 0; n < 5000; n++ {
			if g.recruit.active && (g.recruit.stage == rcGreeting || g.recruit.stage == rcText) && !g.recruit.dialogue.waitingForConfirm() || g.tavern.active && g.tavern.stage == tavText && !g.tavern.dialogue.waitingForConfirm() || g.regionDialogueReturn != nil || !g.dlg.open && !g.tavern.active && !g.recruit.active && g.cd > 0 || g.dlg.open && !g.dlg.waitingForConfirm() {
				step(idle)
				continue
			}
			return
		}
		t.Fatal("normal empty-view input did not settle")
	}
	var persistent []byte
	var holdRNG = g.prng
	var samples []map[string]any
	for i, q := range src.Queued {
		settle()
		if i == 188 {
			persistent, err = encodeSave(g.snapshot())
			if err != nil {
				t.Fatal(err)
			}
			holdRNG = g.prng
		}
		scan, e := strconv.ParseInt(q["scan"], 16, 64)
		if e != nil {
			t.Fatal(e)
		}
		in := idle
		in.AnyKeyEdge = true
		if scan == 0x1c {
			in.Enter = true
			in.Confirm = true
		} else {
			d, ok := map[int64]int{0x50: 0, 0x48: 1, 0x4b: 2, 0x4d: 3}[scan]
			if !ok {
				t.Fatal("scan")
			}
			in.DirEdge = d
			in.DirHeld = d
		}
		step(in)
		step(idle)
		settle()
		if (q["kind"] == "registry_approach" || q["kind"] == "recruit_approach") && scan == 0x1c {
			step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true, AnyKeyEdge: true})
			step(idle)
			settle()
		}
		s := src.States[i]
		x, _ := strconv.Atoi(s["player_x"])
		y, _ := strconv.Atoi(s["player_y"])
		if g.px != x || g.py != y || fmt.Sprintf("%x", g.storyBits) != s["flags"] {
			t.Fatalf("normal field differs at %d", i+1)
		}
		if i+1 >= 150 && (len(g.roster) != 0 || len(g.companions) != 0) {
			t.Fatal("cancel or empty View changed roster")
		}
		if i+1 >= 189 {
			now, e := encodeSave(g.snapshot())
			if e != nil || string(now) != string(persistent) || g.prng != holdRNG {
				t.Fatal("empty View changed persistent state or RNG")
			}
		}
		switch n := i + 1; {
		case n >= 188 && n < openPacket:
			if !g.recruit.active || g.recruit.stage != rcMenu || g.recruit.cursor != i-187 {
				t.Fatal("normal menu differs")
			}
		case n == openPacket && openPacket < againPacket:
			if g.recruit.stage != rcText || !g.recruit.dialogue.waitingForConfirm() || g.recruit.viewFlow != nil {
				t.Fatal("empty selection must display record with one inline wait")
			}
		case n == againPacket || n == againPacket+1:
			if g.recruit.stage != rcAgain || g.recruit.cursor != n-againPacket {
				t.Fatal("empty selection must finish hint and ask continue")
			}
		case n == againPacket+2:
			if !g.recruit.active || g.recruit.stage != rcFinalWait {
				t.Fatal("farewell separate wait missing")
			}
		case n == againPacket+3:
			if g.recruit.active {
				t.Fatal("empty View did not return to field")
			}
		}
		if i+1 >= 150 {
			g.renderFrame()
			name := fmt.Sprintf("packet-%03d.png", i+1)
			f, e := os.Create(filepath.Join(out, name))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			closeErr := f.Close()
			if e != nil || closeErr != nil {
				t.Fatal(e, closeErr)
			}
			samples = append(samples, map[string]any{"packet": i + 1, "original_phase": s["phase"], "recruit_stage": g.recruit.stage, "full_rgb_difference": sourceCanvasDifference(t, g, path, fmt.Sprintf("%s-packet-%03d-%s.png", prefix, i+1, s["phase"]))})
		}
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "empty-view.json"))
	if e := g.Save(); e != nil {
		t.Fatal(e)
	}
	saved, e := encodeSave(g.snapshot())
	if e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	loaded, e := encodeSave(g.snapshot())
	if e != nil || string(saved) != string(loaded) || g.recruit.active || g.tavern.active {
		t.Fatal("normal Save/Load")
	}
	step(InputState{DirHeld: 2, DirEdge: 2})
	step(idle)
	settle()
	if g.px != 1 || g.py != 18 || g.recruit.active {
		t.Fatal("next step after Load")
	}
	report := map[string]any{"source_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "pack_schema": g.pack.Manifest.SchemaVersion, "pack_hash": g.pack.ContentHash(), "original_packets": packets, "action": action, "scope": "normal empty selection finite state parity; full RGB differences retained", "save_load_next_step": true, "samples": samples}
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
}

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

func TestRecruitmentViewDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_VIEW_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem view oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_VIEW_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit view receipt output directory required")
	}
	runRegistrationNormalComparisonRoute(t, dir, out, registrationComparisonRoute{birth: true, entry: true, selection: true, view: true})
}

func TestRecruitmentViewNavigationCancelDoesNotChangeMembers(t *testing.T) {
	g := &Game{roster: []*Member{newMember([]int{0}, 1, 0, 0), newMember([]int{1}, 2, 0, 0)}, companions: []*Member{newMember([]int{2}, 1, 0, 0)}}
	configureRecruitmentEntryFixture(t, &g.recruit)
	rc := &g.recruit
	rc.active, rc.stage = true, rcMenu
	for i := 0; i < 2; i++ {
		g.recruitInput(InputState{DirEdge: 0})
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcView || rc.cursor != 0 || rc.dialogue.open {
		t.Fatal("view must open roster without join prompt")
	}
	rng := g.prng
	beforeRoster := append([]*Member(nil), g.roster...)
	beforeParty := append([]*Member(nil), g.companions...)
	for _, cursor := range []int{1, 0, 1} {
		g.recruitInput(InputState{DirEdge: 0})
		if rc.cursor != cursor {
			t.Fatal("view navigation included party member")
		}
	}
	g.recruitInput(InputState{DirEdge: 1})
	if rc.cursor != 0 {
		t.Fatal("view up navigation")
	}
	g.recruitInput(InputState{DirEdge: -1, Cancel: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcAgain || rc.cursor != 0 {
		t.Fatal("view cancel must ask continue")
	}
	g.recruitInput(InputState{DirEdge: 3})
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcFinalWait {
		t.Fatal("view farewell wait missing")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.active || g.prng != rng || !reflect.DeepEqual(beforeRoster, g.roster) || !reflect.DeepEqual(beforeParty, g.companions) {
		t.Fatal("view cancel changed members or RNG")
	}
}

func TestRecruitmentEmptyViewWaitAndNoPreserveState(t *testing.T) {
	g := &Game{}
	configureRecruitmentEntryFixture(t, &g.recruit)
	rc := &g.recruit
	rc.active, rc.stage, rc.cursor = true, rcMenu, 2
	rng := g.prng
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcText || rc.afterText != rcEmptySelectionReturn || rc.viewFlow != nil {
		t.Fatal("empty View must use reviewed hint without opening a list")
	}
	for i := 0; i < 5000 && !rc.dialogue.waitingForConfirm(); i++ {
		g.recruitInput(InputState{DirEdge: -1, DirHeld: -1})
	}
	if !rc.dialogue.waitingForConfirm() {
		t.Fatal("hint inline wait missing")
	}
	for i := 0; i < 20; i++ {
		g.recruitInput(InputState{DirEdge: -1, DirHeld: 0})
	}
	if !rc.dialogue.waitingForConfirm() || rc.stage != rcText {
		t.Fatal("held key consumed inline wait")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcAgain || rc.cursor != 0 {
		t.Fatal("hint confirmation must stop at separate Yes/No")
	}
	g.recruitInput(InputState{DirEdge: 3})
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcFinalWait || !rc.active {
		t.Fatal("farewell requires separate key")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.active || len(g.roster) != 0 || len(g.companions) != 0 || g.prng != rng {
		t.Fatal("empty View changed state")
	}
}

func TestRecruitmentEmptyJoinLeaveSequencePreservesState(t *testing.T) {
	for _, action := range []string{"join", "leave"} {
		t.Run(action, func(t *testing.T) {
			g := &Game{heroName: []int{0, 1}}
			configureRecruitmentEntryFixture(t, &g.recruit)
			rc := &g.recruit
			rc.active, rc.stage = true, rcMenu
			if action == "leave" {
				rc.cursor = 1
			}
			rng := g.prng
			g.recruitInput(InputState{DirEdge: -1, Enter: true})
			if rc.stage != rcText {
				t.Fatal("empty action must show reviewed text")
			}
			if action == "join" {
				for i := 0; i < 5000 && rc.stage == rcText && !rc.dialogue.waitingForConfirm(); i++ {
					g.recruitInput(InputState{DirEdge: -1, DirHeld: -1})
				}
				if rc.stage != rcText || !rc.dialogue.waitingForConfirm() {
					t.Fatal("Join must show prompt followed by hint inline wait")
				}
				g.recruitInput(InputState{DirEdge: -1, Enter: true})
			} else if !reflect.DeepEqual(rc.dialogue.varGlyphs(0xfffb), g.heroName) {
				t.Fatal("single-party Leave must use current primary name")
			}
			drainRecruitmentSelectionText(t, rc)
			if rc.stage != rcAgain || rc.cursor != 0 {
				t.Fatal("empty action must stop at separate Yes/No")
			}
			g.recruitInput(InputState{DirEdge: 3})
			g.recruitInput(InputState{DirEdge: -1, Enter: true})
			drainRecruitmentSelectionText(t, rc)
			if rc.stage != rcFinalWait || !rc.active {
				t.Fatal("farewell separate wait missing")
			}
			g.recruitInput(InputState{DirEdge: -1, Enter: true})
			if rc.active || len(g.roster) != 0 || len(g.companions) != 0 || g.prng != rng || !reflect.DeepEqual(g.heroName, []int{0, 1}) {
				t.Fatal("empty action changed persistent state")
			}
		})
	}
}
