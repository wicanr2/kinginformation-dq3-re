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

func chooseRemoveCurse(t *testing.T, g *Game, target int) {
	t.Helper()
	g.church.open(nil)
	g.churchInput(InputState{DirEdge: 0}) // option 1：解詛咒
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
	if g.church.stage != churchTarget || g.church.msg != churchMsgNoReviveNeed {
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

func TestChurchCurePoisonChargesPackFixedCostAndClearsOnlySelectedMember(t *testing.T) {
	g := churchTestGame(t)
	g.heroConditions = conditionPoison
	companion := newMember([]int{1}, 3, 0, 0)
	companion.Conditions = conditionPoison
	g.companions = []*Member{companion}
	g.church.open(nil)
	g.churchInput(InputState{Confirm: true}) // option 0：解毒
	g.churchInput(InputState{DirEdge: 0})    // 同伴
	g.churchInput(InputState{Confirm: true})
	if g.church.stage != churchConfirm || g.church.pendingCost != g.pack.CurePoisonCost() {
		t.Fatalf("poison confirmation stage=%d cost=%d", g.church.stage, g.church.pendingCost)
	}
	g.churchInput(InputState{Confirm: true})
	if companion.Conditions&conditionPoison != 0 || g.heroConditions&conditionPoison == 0 {
		t.Fatalf("selected poison transaction mismatch: hero=%#x companion=%#x",
			g.heroConditions, companion.Conditions)
	}
	if g.heroGold != 95 || g.church.msg != churchMsgPoisonCured {
		t.Fatalf("poison cure gold=%d msg=%q", g.heroGold, g.church.msg)
	}
}

func TestChurchCurePoisonRejectsUnaffectedTargetWithoutCharge(t *testing.T) {
	g := churchTestGame(t)
	g.church.open(nil)
	g.churchInput(InputState{Confirm: true})
	g.churchInput(InputState{Confirm: true})
	if g.church.stage != churchTarget || g.church.msg != churchMsgNotPoisoned || g.heroGold != 100 {
		t.Fatalf("unaffected poison target stage=%d msg=%q gold=%d",
			g.church.stage, g.church.msg, g.heroGold)
	}
}

func TestChurchRemoveCurseChargesLevelOnceAndDestroysAllCursedEquipment(t *testing.T) {
	g := churchTestGame(t)
	g.heroGold = 500
	g.equip = [4]int{0x1b, 0x1e, 0x3d, -1}
	g.inventory = []int{0x03}
	chooseRemoveCurse(t, g, 0)
	if g.church.stage != churchConfirm || g.church.pendingCost != 100 {
		t.Fatalf("curse confirmation stage=%d cost=%d", g.church.stage, g.church.pendingCost)
	}
	g.churchInput(InputState{Confirm: true})
	if g.heroGold != 400 || g.equip != [4]int{-1, 0x1e, -1, -1} {
		t.Fatalf("curse transaction gold=%d equip=%v", g.heroGold, g.equip)
	}
	if len(g.inventory) != 1 || g.inventory[0] != 0x03 || g.church.msg != churchMsgCurseRemoved {
		t.Fatalf("removed equipment must not enter inventory: inv=%v msg=%q", g.inventory, g.church.msg)
	}
}

func TestChurchRemoveCurseRejectsUnaffectedTargetWithoutCharge(t *testing.T) {
	g := churchTestGame(t)
	g.equip = [4]int{0x03, 0x1e, -1, -1}
	chooseRemoveCurse(t, g, 0)
	if g.church.stage != churchTarget || g.church.msg != churchMsgNotCursed || g.heroGold != 100 {
		t.Fatalf("unaffected curse target stage=%d msg=%q gold=%d",
			g.church.stage, g.church.msg, g.heroGold)
	}
}

func TestChurchRemoveCurseInsufficientGoldHasNoSideEffect(t *testing.T) {
	g := churchTestGame(t)
	g.heroGold = 99
	g.equip = [4]int{0x1d, 0x1e, -1, -1}
	chooseRemoveCurse(t, g, 0)
	g.churchInput(InputState{Confirm: true})
	if g.heroGold != 99 || g.equip[0] != 0x1d || g.church.msg != churchMsgGoldShort {
		t.Fatalf("insufficient curse transaction gold=%d equip=%v msg=%q",
			g.heroGold, g.equip, g.church.msg)
	}
}

func TestCursedEquipmentSaveRoundTripRetainsChurchConsumer(t *testing.T) {
	g := churchTestGame(t)
	g.equip = [4]int{0x1d, 0x1e, -1, -1}
	raw, err := encodeSave(g.snapshot())
	if err != nil {
		t.Fatalf("encodeSave: %v", err)
	}
	saved, err := decodeSave(raw)
	if err != nil {
		t.Fatalf("decodeSave: %v", err)
	}
	restored := churchTestGame(t)
	restored.restore(saved)
	if restored.equip[0] != 0x1d || !restored.churchMemberCursed(0) {
		t.Fatalf("cursed equipment round trip lost consumer: equip=%v", restored.equip)
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
