package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

const (
	soloChallengeIdle = iota
	soloChallengePrompt
	soloChallengeChoice
	soloChallengeAccept
	soloChallengeReturn
)

func (g *Game) activeSoloChallenge() (*gamepack.TemporarySoloChallengeEvent, bool) {
	if g.pack == nil || g.soloChallengeEventID == "" {
		return nil, false
	}
	return g.pack.TemporarySoloChallenge(g.soloChallengeEventID)
}

// resolveSoloChallengeWorldEntrance applies the original world-loader read
// branch for an overlapping normal/solo entrance. It reads the proven
// completed flag but never writes it.
func (g *Game) resolveSoloChallengeWorldEntrance(cty int) int {
	if g.pack == nil || g.inTown {
		return cty
	}
	for _, event := range g.pack.TemporarySoloChallenges() {
		p := event.SoloWorldPosition
		if g.layer != p.Layer || g.px != p.X || g.py != p.Y {
			continue
		}
		if g.storyFlag(event.CompletedFlagRaw) {
			return event.EntryNPC.CTYRaw
		}
		return event.ReturnNPC.CTYRaw
	}
	return cty
}

// talkTemporarySoloChallenge owns only the finite party-removal transaction.
// In particular it does not invent a writer for CompletedFlagRaw: IDA proves
// the read branch, but its writer remains unknown.
func (g *Game) talkTemporarySoloChallenge(n *npcInst) bool {
	if g.pack == nil || g.soloChallengeStage != soloChallengeIdle {
		return false
	}
	for _, event := range g.pack.TemporarySoloChallenges() {
		switch {
		case g.scriptedNPCMatches(n, event.EntryNPC):
			if g.storyFlag(event.CompletedFlagRaw) {
				g.openPackText(event.DialogueTextIDs.Completed)
				return true
			}
			g.soloChallengeEventID = event.ID
			if !g.openPackText(event.DialogueTextIDs.Prompt) {
				g.finishSoloChallengeDialogue()
				return true
			}
			g.soloChallengeStage = soloChallengePrompt
			return true
		case g.scriptedNPCMatches(n, event.ReturnNPC):
			if !g.soloChallengeActive || g.soloChallengeEventID != event.ID {
				return true
			}
			if !g.openPackText(event.DialogueTextIDs.Return) {
				return true
			}
			g.soloChallengeStage = soloChallengeReturn
			return true
		default:
			for _, message := range event.CaveMessages {
				if g.inTown && g.cur != nil && g.curCty == message.CTYRaw &&
					g.cur.sec == message.Section && n != nil && n.b4 == message.HandlerRaw {
					g.openPackText(message.TextID)
					return true
				}
			}
		}
	}
	return false
}

func (g *Game) advanceSoloChallengeDialogue() {
	event, ok := g.activeSoloChallenge()
	if !ok {
		g.finishSoloChallengeDialogue()
		return
	}
	switch g.soloChallengeStage {
	case soloChallengePrompt:
		g.soloChallengeStage, g.soloChallengeCursor = soloChallengeChoice, 0
	case soloChallengeAccept:
		g.beginSoloChallenge(event)
		g.soloChallengeStage = soloChallengeIdle
	case soloChallengeReturn:
		g.finishSoloChallenge(event)
		g.soloChallengeStage = soloChallengeIdle
	}
}

func (g *Game) soloChallengeChoiceInput(in InputState) {
	event, ok := g.activeSoloChallenge()
	if !ok {
		g.finishSoloChallengeDialogue()
		return
	}
	if in.DirEdge == 0 || in.DirEdge == 1 {
		g.soloChallengeCursor ^= 1
		return
	}
	if !in.Confirm && !in.Cancel {
		return
	}
	yes := g.soloChallengeCursor == 0 && !in.Cancel
	if yes {
		if g.openPackText(event.DialogueTextIDs.Accept) {
			g.soloChallengeStage = soloChallengeAccept
			return
		}
	} else if g.openPackText(event.DialogueTextIDs.Reject) {
		g.soloChallengeStage = soloChallengeIdle
		g.soloChallengeEventID = ""
		return
	}
	g.finishSoloChallengeDialogue()
}

func (g *Game) beginSoloChallenge(event *gamepack.TemporarySoloChallengeEvent) {
	if event == nil || g.soloChallengeActive {
		return
	}
	g.soloChallengeCompanions = append([]*Member(nil), g.companions...)
	g.companions = nil
	g.soloChallengeActive = true
	g.shipAboard, g.phoenixAboard = false, false
	g.layer = event.SoloWorldPosition.Layer
	g.px, g.py = event.SoloWorldPosition.X, event.SoloWorldPosition.Y
	g.overPx, g.overPy = g.px, g.py
	g.inTown, g.curCty = false, -1
	g.cur = g.overworldScene()
	g.applyDaynightPalette()
}

func (g *Game) finishSoloChallenge(event *gamepack.TemporarySoloChallengeEvent) {
	if event == nil || !g.soloChallengeActive {
		return
	}
	g.companions = append([]*Member(nil), g.soloChallengeCompanions...)
	g.soloChallengeCompanions = nil
	g.soloChallengeActive = false
	d := event.ReturnTownDestination
	g.inTown, g.curCty = true, d.CTYRaw
	g.restoreTownScene(d.CTYRaw, d.Section)
	g.px, g.py = d.X, d.Y
	g.soloChallengeEventID = ""
}

func (g *Game) finishSoloChallengeDialogue() {
	g.soloChallengeStage = soloChallengeIdle
	g.soloChallengeCursor = 0
	if !g.soloChallengeActive {
		g.soloChallengeEventID = ""
	}
}

func (g *Game) drawSoloChallengeChoice(rgba []byte, white dq3data.Color) {
	if g.soloChallengeStage != soloChallengeChoice {
		return
	}
	event, ok := g.activeSoloChallenge()
	if !ok {
		return
	}
	g.drawBinaryChoice(rgba, white, g.soloChallengeCursor,
		event.DialogueTextIDs.ChoiceYes, event.DialogueTextIDs.ChoiceNo)
}
