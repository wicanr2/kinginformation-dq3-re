package game

import "testing"

func TestEncounterEnabledUsesRawCTYSectionFlag(t *testing.T) {
	g := &Game{inTown: true, cur: &Scene{encounterFlag: 0}}
	if g.encounterEnabled() {
		t.Fatal("CTY section +0x11=0 應為安全區，不得推進遭遇")
	}
	g.cur.encounterFlag = 1
	if !g.encounterEnabled() {
		t.Fatal("CTY section +0x11 非零應允許推進遭遇")
	}
	g.inTown = false
	if !g.encounterEnabled() {
		t.Fatal("地表不依 CTY section flag gate")
	}
	g.phoenixAboard = true
	if g.encounterEnabled() {
		t.Fatal("飛行中不得推進遭遇")
	}
}

func TestEncounterStepScheduleBounds(t *testing.T) {
	g := &Game{dnPhase: 0}
	g.prng.Seed(0x1357)
	for i := 0; i < 64; i++ {
		got := 4 + g.prng.Next(0x10)
		if got < 4 || got > 19 {
			t.Fatalf("初始遭遇步數=%d，不在原版 4..19", got)
		}
	}
	for phase := 0; phase < 4; phase++ {
		g.dnPhase = phase
		for i := 0; i < 64; i++ {
			got := g.nextEncounterStep()
			if phase == 0 {
				if got < 10 || got > 17 {
					t.Fatalf("白天重擲步數=%d，不在原版 10..17", got)
				}
			} else if got < 8 || got > 15 {
				t.Fatalf("非白天重擲步數=%d，不在原版減 2 後 8..15", got)
			}
		}
	}
}
