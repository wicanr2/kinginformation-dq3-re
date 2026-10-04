package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/opl2"
)

const (
	rcJoinedText = rcFinalWait + 1 + iota
	rcLeaderText
	rcFinishText
	rcMusicWait
)

type oneShotAudio interface {
	PlayOneShot([]byte, int64, int64, string, int64, string)
	FinishOneShot()
}

func (rc *Recruit) configureJoinStream(data []byte) error {
	s := rc.join.Sound
	if s.Start < 0 || s.End > len(data) || s.End <= s.Start {
		return fmt.Errorf("recruitment join source range is invalid")
	}
	stream := data[s.Start:s.End]
	events, ticks, err := opl2.ParseEventStream(stream)
	if err != nil || len(events) != s.EventCount || ticks != uint64(s.DeltaTicks) {
		return fmt.Errorf("recruitment join event identity differs")
	}
	rc.musicStream = append([]byte(nil), stream...)
	return nil
}

func (g *Game) startRecruitmentJoin(name []int) {
	rc := &g.recruit
	rc.joinedName = append([]int(nil), name...)
	rc.leaderName = append([]int(nil), g.heroName...)
	rc.startJoinText(rc.join.JoinedTextID, rc.joinedName, rcJoinedText)
}

func (rc *Recruit) startJoinText(id string, name []int, stage int) {
	rc.dialogue.appendRetainedRecord(rc.texts[id])
	rc.dialogue.shadow = &rc.contract.Shadow
	rc.dialogue.varGlyph = map[uint16][]int{rc.join.NameControlCode: name}
	rc.stage = stage
}

func (g *Game) recruitJoinInput(in InputState) {
	rc := &g.recruit
	if rc.musicFrames > 0 {
		rc.musicFrames--
	}
	if rc.stage == rcMusicWait {
		if rc.musicFrames == 0 {
			g.finishRecruitmentMusic()
			rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
		}
		return
	}
	rc.dialogue.Tick()
	if !rc.dialogue.open {
		switch rc.stage {
		case rcJoinedText:
			rc.startJoinText(rc.join.LeaderTextID, rc.leaderName, rcLeaderText)
		case rcLeaderText:
			rc.musicFrames = rc.join.Sound.HoldFrames()
			g.startRecruitmentMusic()
			rc.startJoinText(rc.join.FinishTextID, rc.joinedName, rcFinishText)
		case rcFinishText:
			rc.stage = rcMusicWait
		}
		return
	}
	if rc.dialogue.waitingForConfirm() && (in.Confirm || in.Enter || in.AnyKeyEdge || in.Cancel || in.DirEdge >= 0 || in.Tapped) {
		rc.dialogue.Advance()
	}
}

func (g *Game) startRecruitmentMusic() {
	a, ok := g.music.(oneShotAudio)
	if !ok {
		return
	}
	s := g.recruit.join.Sound
	a.PlayOneShot(g.recruit.musicStream, s.ClockDivisor, s.ReferenceHz, s.RenderFile, s.RenderSize, s.RenderSHA256)
}

func (g *Game) finishRecruitmentMusic() {
	if a, ok := g.music.(oneShotAudio); ok {
		a.FinishOneShot()
	}
}
