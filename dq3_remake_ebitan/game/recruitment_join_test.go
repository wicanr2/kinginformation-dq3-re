package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

type joinRecordingAudio struct {
	recordingAudio
	starts, finishes int
}

func (a *joinRecordingAudio) PlayOneShot([]byte, int64, int64, string, int64, string) { a.starts++ }
func (a *joinRecordingAudio) FinishOneShot()                                          { a.finishes++ }

func TestRecruitmentJoinNamesOneShotAndTransaction(t *testing.T) {
	m := newMember([]int{77}, 1, 0, 0)
	a := &joinRecordingAudio{}
	g := &Game{heroName: []int{42}, roster: []*Member{m}, music: a}
	configureRecruitmentEntryFixture(t, &g.recruit)
	rc := &g.recruit
	rc.active, rc.stage = true, rcJoin
	rng := g.prng
	idle := InputState{DirEdge: -1, DirHeld: -1}
	g.recruitInput(InputState{DirEdge: -1, DirHeld: -1, Enter: true})
	if rc.stage != rcJoinedText || !reflect.DeepEqual(rc.dialogue.varGlyph[rc.join.NameControlCode], []int{77}) {
		t.Fatal("joined text binding")
	}
	waits := 0
	for i := 0; i < 5000 && rc.stage != rcMusicWait; i++ {
		in := idle
		if rc.dialogue.waitingForConfirm() {
			in.Enter = true
			waits++
		}
		g.recruitInput(in)
		if rc.stage == rcLeaderText && !reflect.DeepEqual(rc.dialogue.varGlyph[rc.join.NameControlCode], []int{42}) {
			t.Fatal("leader text must use primary actor")
		}
		if rc.stage == rcFinishText && !reflect.DeepEqual(rc.dialogue.varGlyph[rc.join.NameControlCode], []int{77}) {
			t.Fatal("finish text must use joined actor")
		}
	}
	if rc.stage != rcMusicWait || rc.musicFrames <= 0 || waits != 1 || a.starts != 1 || a.finishes != 0 {
		t.Fatal("inline wait/one-shot order", rc.stage, rc.musicFrames, waits, a.starts, a.finishes)
	}
	var inserted []int
	for _, op := range rc.dialogue.retained.ops {
		if !op.scroll && (op.glyph == 42 || op.glyph == 77) {
			inserted = append(inserted, op.glyph)
		}
	}
	if !reflect.DeepEqual(inserted, []int{77, 77, 42, 77}) {
		t.Fatal("consumed name glyph order", inserted)
	}
	remaining := rc.musicFrames
	for i := 0; i < remaining-1; i++ {
		g.recruitInput(InputState{DirEdge: 0, DirHeld: 0, Enter: true, Cancel: true})
		if rc.stage != rcMusicWait {
			t.Fatal("input shortened playback")
		}
	}
	g.recruitInput(idle)
	if a.finishes != 1 || rc.stage != rcText || rc.afterText != rcAgain || g.prng != rng || len(g.roster) != 0 || !reflect.DeepEqual(g.companions, []*Member{m}) {
		t.Fatal("completion repeated or changed transaction")
	}
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcAgain || rc.cursor != 0 {
		t.Fatal("completion must continue native Yes/No")
	}
}

func TestRecruitmentJoinDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_JOIN_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem normal199 oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_JOIN_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit output required")
	}
	runRegistrationNormalComparison(t, dir, out, true, true, true, false, true)
}

func TestRecruitmentJoinLoadClearsAudioOnlyAfterValidation(t *testing.T) {
	g, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "join.json"))
	a := &joinRecordingAudio{}
	g.music = a
	g.roster = []*Member{newMember([]int{77}, 1, 0, 0)}
	g.recruit.open()
	g.recruit.stage = rcJoin
	g.recruitInput(InputState{DirEdge: -1, DirHeld: -1, Enter: true})
	for i := 0; i < 5000 && g.recruit.stage != rcMusicWait; i++ {
		in := InputState{DirEdge: -1, DirHeld: -1}
		if g.recruit.dialogue.waitingForConfirm() {
			in.Enter = true
		}
		g.recruitInput(in)
	}
	if g.recruit.stage != rcMusicWait || a.starts != 1 {
		t.Fatal("missing music wait")
	}
	if e = g.Save(); e != nil {
		t.Fatal(e)
	}
	valid, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	frames, flow := g.recruit.musicFrames, g.recruit.dialogue.retained
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
	if g.recruit.stage != rcMusicWait || g.recruit.musicFrames != frames || g.recruit.dialogue.retained != flow || a.finishes != 0 {
		t.Fatal("rejected Load changed music/UI")
	}
	if e = os.WriteFile(savePath(), valid, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	if a.finishes != 1 || g.recruit.active || g.recruit.musicFrames != 0 || g.recruit.backdrop != nil || g.recruit.dialogue.retained != nil || len(g.companions) != 1 || len(g.roster) != 0 {
		t.Fatal("valid Load must preserve member and release music/UI")
	}
}
