package game

import (
	"os"
	"testing"
)

// TestE2BattlePackEncounterTrace verifies the production bootstrap consumes
// the versioned pack raw table, not the legacy DQ3.EXE path or a Go fallback.
// It is a deterministic headless contract test; it does not start DOSBox or
// assert unproven visual/audio timing.
func TestE2BattlePackEncounterTrace(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if g.encounters == nil {
		t.Fatal("production game has no encounter decoder")
	}
	if got := g.encounters.Region(153, 174); got != 1 {
		t.Fatalf("Aliahan region=%d, want raw region 1", got)
	}
	slot := g.encounters.Slot(1, 0)
	if slot.Threshold != 1 || slot.Background != 0 {
		t.Fatalf("pack encounter region1/sub0 header=%+v", slot)
	}
	want := []int{5, 6, 5, 10}
	if len(slot.Candidates) != len(want) {
		t.Fatalf("pack encounter candidates=%v, want %v", slot.Candidates, want)
	}
	for i, id := range want {
		if slot.Candidates[i] != id {
			t.Fatalf("pack encounter candidates=%v, want %v", slot.Candidates, want)
		}
	}
	if _, ok := g.pack.BattlePack(); !ok {
		t.Fatal("production game did not retain battle pack contract")
	}
}
