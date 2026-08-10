package game

import (
	"os"
	"testing"
)

func TestOpeningCutsceneBootAndSkip(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.openingSeq == nil || len(g.openingSeq.Frames) != 5 || len(g.openingPix) != 5 {
		t.Fatalf("opening assets=%v/%d, want five pack frames", g.openingSeq, len(g.openingPix))
	}
	g.StartOpeningCutscene()
	if !g.openingActive || g.openingIndex != 0 || g.openingFrame != 0 {
		t.Fatalf("opening should start at first frame: active=%v index=%d frame=%d",
			g.openingActive, g.openingIndex, g.openingFrame)
	}
	if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if !g.openingActive || g.openingFrame != 1 {
		t.Fatalf("no-input frame should advance opening: active=%v frame=%d", g.openingActive, g.openingFrame)
	}
	if err := g.step(InputState{Confirm: true}); err != nil {
		t.Fatal(err)
	}
	if g.openingActive || g.newGame.stage != ngMenu {
		t.Fatalf("confirm should skip opening and open title menu: active=%v stage=%d", g.openingActive, g.newGame.stage)
	}
}
