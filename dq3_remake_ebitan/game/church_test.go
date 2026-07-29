package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func churchTestGame(t *testing.T) *Game {
	t.Helper()
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	return &Game{pack: pack, heroGold: 100, heroHP: 1, heroInit: true}
}

func chooseRevive(t *testing.T, g *Game, target int) {
	t.Helper()
	g.church.open(nil)
	g.churchInput(InputState{DirEdge: 0})
	g.churchInput(InputState{DirEdge: 0})
	g.churchInput(InputState{Confirm: true})
	for i := 0; i < target; i++ {
		g.churchInput(InputState{DirEdge: 0})
	}
	g.churchInput(InputState{Confirm: true})
}

func TestChurchRevivePaysExactCostAndRestoresMember(t *testing.T) {
	g := churchTestGame(t)
	dead := newMember([]int{1}, 3, 0, 0)
	dead.CurHP, dead.CurMP = 0, 0
	g.companions = []*Member{dead}

	chooseRevive(t, g, 1)
	if g.church.stage != churchConfirm || g.church.pendingCost != 10 {
		t.Fatalf("revive confirmation stage=%d cost=%d", g.church.stage, g.church.pendingCost)
	}
	g.churchInput(InputState{Confirm: true})

	if g.heroGold != 90 {
		t.Fatalf("Lv1 revive should charge 10G, gold=%d", g.heroGold)
	}
	if dead.CurHP != dead.MaxHP() || dead.CurMP != dead.MaxMP() {
		t.Fatalf("revived member not full: HP=%d/%d MP=%d/%d",
			dead.CurHP, dead.MaxHP(), dead.CurMP, dead.MaxMP())
	}
	if g.church.stage != churchService {
		t.Fatalf("successful transaction should return to service menu, stage=%d", g.church.stage)
	}
}

func TestChurchReviveInsufficientGoldHasNoSideEffect(t *testing.T) {
	g := churchTestGame(t)
	g.heroGold = 9
	g.heroHP = 0
	chooseRevive(t, g, 0)
	g.churchInput(InputState{Confirm: true})

	if g.heroGold != 9 || g.heroHP != 0 {
		t.Fatalf("insufficient funds changed state: gold=%d hp=%d", g.heroGold, g.heroHP)
	}
}

func TestChurchRejectsLivingTargetAndCancelDoesNotCharge(t *testing.T) {
	g := churchTestGame(t)
	chooseRevive(t, g, 0)
	if g.church.stage != churchTarget || g.church.msg != "不需要復活" {
		t.Fatalf("living target should stay in target menu: stage=%d msg=%q",
			g.church.stage, g.church.msg)
	}
	if g.heroGold != 100 {
		t.Fatalf("living target charged gold=%d", g.heroGold)
	}

	g.heroHP = 0
	g.churchInput(InputState{Confirm: true})
	g.churchInput(InputState{Cancel: true})
	if g.church.stage != churchTarget || g.heroGold != 100 || g.heroHP != 0 {
		t.Fatalf("cancel changed transaction: stage=%d gold=%d hp=%d",
			g.church.stage, g.heroGold, g.heroHP)
	}
}

func TestChurchFacilityOpensModalInsteadOfSaving(t *testing.T) {
	g := churchTestGame(t)
	g.curCty = 0
	g.openFacility(3)
	if !g.church.active || g.church.stage != churchService {
		t.Fatalf("CTY00 church facility did not open modal: active=%v stage=%d",
			g.church.active, g.church.stage)
	}
}
