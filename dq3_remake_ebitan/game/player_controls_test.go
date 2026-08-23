package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gaudio"
)

func TestTalkTurnsNPCTowardPlayer(t *testing.T) {
	for playerFacing, wantNPC := range map[int]int{0: 1, 1: 0, 2: 3, 3: 2} {
		g := &Game{
			facing: playerFacing,
			cur:    &Scene{w: 3, h: 3, npcs: []npcInst{{x: 1, y: 1}}},
			px:     1, py: 1,
		}
		dx, dy := dirDelta(playerFacing)
		g.px, g.py = 1-dx, 1-dy
		g.selectCommand(cmdTalk)
		if got := g.cur.npcs[0].facing; got != wantNPC {
			t.Fatalf("player facing %d -> NPC facing %d, want %d", playerFacing, got, wantNPC)
		}
	}
}

func TestHelpAndSystemShortcutsOpenAndClose(t *testing.T) {
	g := &Game{cfg: config.Default(), music: gaudio.NewMusic(nil)}
	if !g.utilityInput(InputState{Help: true}) || !g.help.open {
		t.Fatal("F1 should open HELP")
	}
	if !g.utilityInput(InputState{Cancel: true}) || g.help.open {
		t.Fatal("ESC should close HELP")
	}
	if !g.utilityInput(InputState{Settings: true}) || !g.settings.open {
		t.Fatal("F2/S should open system settings")
	}
	if !g.utilityInput(InputState{Cancel: true}) || g.settings.open {
		t.Fatal("ESC should close system settings")
	}
}
