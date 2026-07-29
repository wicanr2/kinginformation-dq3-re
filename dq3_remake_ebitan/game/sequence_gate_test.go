package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func sequenceGateTestGame(t *testing.T) (*Game, gamepack.TwoStepFloorSwitchGate) {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	events := g.pack.TwoStepFloorSwitchGates()
	if len(events) != 1 {
		t.Fatalf("two-step floor switch gate count=%d, want 1", len(events))
	}
	event := events[0]
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.CTYRaw, mapBlkNum[event.CTYRaw], event.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, event.CTYRaw
	g.dlg.tx = sc.dlgText
	return g, event
}

func closeSequenceGateDialogue(g *Game) {
	g.dlg.open = false
	g.advanceSequenceGateDialogue()
}

func pressSequenceGate(t *testing.T, g *Game, event gamepack.TwoStepFloorSwitchGate,
	sw gamepack.FloorSwitch, yes bool) {
	t.Helper()
	g.px, g.py = sw.Tile.X, sw.Tile.Y
	if !g.trySequenceGateEvent() || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Prompt)) {
		t.Fatalf("switch %+v 未開 pack prompt", sw)
	}
	closeSequenceGateDialogue(g)
	if !g.sequenceGateChoosing() {
		t.Fatalf("prompt 關閉後 stage=%d，未進 Yes/No", g.sequenceGateStage)
	}
	if !yes {
		g.sequenceGateChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
		return
	}
	g.sequenceGateChoiceInput(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
	if !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Pressed)) {
		t.Fatal("選是後未播放 pack pressed text")
	}
}

func TestTwoStepFloorSwitchGateDeclineDoesNotArm(t *testing.T) {
	g, event := sequenceGateTestGame(t)
	pressSequenceGate(t, g, event, event.Switches[0], false)
	if g.sequenceGateArmed || g.sequenceGateStage != sequenceGateIdle ||
		!g.storyFlag(event.ClearFlagRaw) {
		t.Fatal("選否不得武裝、清旗標或留下 modal")
	}
}

func TestTwoStepFloorSwitchGateSuccessOpensGateAndTreasure(t *testing.T) {
	g, event := sequenceGateTestGame(t)
	var first, completion gamepack.FloorSwitch
	for _, sw := range event.Switches {
		if sw.TileSubID == event.CompletionTileSubID {
			completion = sw
		} else if first == (gamepack.FloorSwitch{}) {
			first = sw
		}
	}
	if !g.cur.Blocked(13, 10) {
		t.Fatal("原版 clear flag 尚 set 時石門應阻擋")
	}
	pressSequenceGate(t, g, event, first, true)
	closeSequenceGateDialogue(g)
	if !g.sequenceGateArmed || g.sequenceGateStage != sequenceGateIdle {
		t.Fatal("第一個開關應只武裝 transient state")
	}
	pressSequenceGate(t, g, event, completion, true)
	closeSequenceGateDialogue(g)
	if g.storyFlag(event.ClearFlagRaw) || g.cur.Blocked(13, 10) ||
		!g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Success)) {
		t.Fatalf("正確順序未清旗標／開門／顯示 success：flag=%v blocked=%v",
			g.storyFlag(event.ClearFlagRaw), g.cur.Blocked(13, 10))
	}
	closeSequenceGateDialogue(g)

	tr := event.UnlockedTreasure
	if !g.collectQuestTreasure(tr.CTYRaw, tr.Section, tr.TileSubID) ||
		g.storyFlag(tr.PresentFlag) || !g.hasItem(tr.ItemRawID) {
		t.Fatal("開門後魔法鑰匙未由 pack treasure selector 取得")
	}
	before := g.countItem(tr.ItemRawID)
	if !g.collectQuestTreasure(tr.CTYRaw, tr.Section, tr.TileSubID) ||
		g.countItem(tr.ItemRawID) != before {
		t.Fatal("魔法鑰匙 present flag 清除後不得重複取得")
	}
}

func TestTwoStepFloorSwitchGateWrongSecondSwitchTraps(t *testing.T) {
	g, event := sequenceGateTestGame(t)
	first := event.Switches[0]
	wrong := event.Switches[1]
	if wrong.TileSubID == event.CompletionTileSubID {
		wrong = event.Switches[2]
	}
	pressSequenceGate(t, g, event, first, true)
	closeSequenceGateDialogue(g)
	pressSequenceGate(t, g, event, wrong, true)
	closeSequenceGateDialogue(g)
	if !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, mustPackTextCodes(t, g, event.DialogueTextIDs.Trap)) {
		t.Fatal("錯誤第二鍵未播放 pack trap text")
	}
	closeSequenceGateDialogue(g)
	d := event.TrapDestination
	if g.curCty != d.CTYRaw || g.cur.sec != d.Section || g.px != d.X || g.py != d.Y ||
		g.sequenceGateArmed || !g.storyFlag(event.ClearFlagRaw) {
		t.Fatalf("陷阱目的地／狀態錯：cty=%d sec=%d @(%d,%d) armed=%v",
			g.curCty, g.cur.sec, g.px, g.py, g.sequenceGateArmed)
	}
}
