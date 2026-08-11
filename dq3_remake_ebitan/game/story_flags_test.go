package game

import (
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestStoryFlagRuntimePackContracts(t *testing.T) {
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]struct {
		kind    string
		handler int
		set     int
		clear   int
	}{
		"dq3:story.post_zoma_ending":                {"post_zoma_ending", 74, 0x0d, 0xf5},
		"dq3:story.phoenix_revival":                 {"phoenix_revival", 69, 0x11, 0x8e},
		"dq3:story.baramos_aftermath":               {"boss_aftermath", 70, 0x1a, 0x29},
		"dq3:story.samanosa_conditional_visibility": {"conditional_story_flag", 33, 0x24, 0x3d},
	}
	for id, expected := range want {
		event, ok := p.StoryFlagRuntimeEventKind(expected.kind)
		if !ok {
			t.Fatalf("缺少 pack event %s", id)
		}
		if event.Kind != expected.kind || event.HandlerRaw != expected.handler {
			t.Fatalf("%s primitive 不符: kind=%s handler=%d", id, event.Kind, event.HandlerRaw)
		}
		if !containsInt(event.SetStoryFlagsRaw, expected.set) || !containsInt(event.ClearStoryFlagsRaw, expected.clear) {
			t.Fatalf("%s flag transaction 不符: set=%v clear=%v", id, event.SetStoryFlagsRaw, event.ClearStoryFlagsRaw)
		}
		if expected.kind == "phoenix_revival" && (event.DialogueRecordRaw != 92 || event.PhoenixOrbDialogueRaw != 93 ||
			event.PhoenixRevivedDialogueRaw != 94 || event.PhoenixAltarDialogueRaw != 96 || event.PhoenixAltarVisualTileRaw != 149) {
			t.Fatalf("phoenix pack dialogue/tile raw 不符: %+v", event)
		}
		if event.Evidence.Level != "D3" || event.Evidence.SourceKind != "exe" || event.Evidence.AddressSpace != "linear" {
			t.Fatalf("%s 缺少可回查 IDA evidence: %+v", id, event.Evidence)
		}
	}
}

func TestPostZomaEndingStoryFlagWriter(t *testing.T) {
	g := r4Game(t)
	g.cleared = true
	g.endingKingStage = 1
	g.setStoryFlag(0xf5, true)
	g.advanceEndingKing()
	if !g.storyFlag(0x0d) || g.storyFlag(0xf5) || !g.lotoBlessed {
		t.Fatalf("post-Zoma story transaction 錯: set0d=%v clearf5=%v loto=%v",
			g.storyFlag(0x0d), g.storyFlag(0xf5), g.lotoBlessed)
	}
}

func TestPhoenixRuntimeMapMutationAndSaveRoundTrip(t *testing.T) {
	g := phoenixEventGame(t)
	e, ok := g.phoenixRuntimeEvent()
	if !ok {
		t.Fatal("缺 phoenix_revival pack primitive")
	}
	g.setStoryFlag(e.PhoenixUnrevivedFlagRaw, true)
	g.phoenixStage, g.phoenixAnim, g.phoenixAnimTick = 3, 0, 0
	g.phoenixOwned, g.phoenixAboard = false, false
	for i := 0; i < len(e.Animation.TilesRaw)*e.Animation.FrameTicks; i++ {
		g.advancePhoenixAnimation()
	}
	if g.phoenixStage != 0 || !g.phoenixOwned || g.phoenixAboard {
		t.Fatalf("拉米亞復活 transaction 未完成: stage=%d owned=%v aboard=%v",
			g.phoenixStage, g.phoenixOwned, g.phoenixAboard)
	}
	if g.storyFlag(e.PhoenixUnrevivedFlagRaw) || g.storyFlag(e.PhoenixTransientFlagRaw) || !g.phoenixMapCellWritten {
		t.Fatalf("handler69 flags/map state 錯: unrevived=%v transient=%v cell=%v",
			g.storyFlag(e.PhoenixUnrevivedFlagRaw), g.storyFlag(e.PhoenixTransientFlagRaw), g.phoenixMapCellWritten)
	}
	if got := g.cur.tileIdx(e.MapCell.Tile.X, e.MapCell.Tile.Y); got != e.MapCell.ValueRaw {
		t.Fatalf("CTY70 map cell 未寫入 0x80: got=%#x want=%#x", got, e.MapCell.ValueRaw)
	}
	if g.phoenixX != e.ParkPosition.X || g.phoenixY != e.ParkPosition.Y {
		t.Fatalf("拉米亞停放座標錯: (%d,%d) want (%d,%d)", g.phoenixX, g.phoenixY, e.ParkPosition.X, e.ParkPosition.Y)
	}

	saved := g.snapshot()
	b, err := encodeSave(saved)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := decodeSave(b)
	if err != nil {
		t.Fatal(err)
	}
	restored := phoenixEventGame(t)
	restored.restore(decoded)
	re, ok := restored.phoenixRuntimeEvent()
	if !ok || !restored.phoenixMapCellWritten || restored.cur.tileIdx(re.MapCell.Tile.X, re.MapCell.Tile.Y) != re.MapCell.ValueRaw {
		t.Fatalf("save/load 未重放 CTY70 map patch: written=%v tile=%#x", restored.phoenixMapCellWritten, restored.cur.tileIdx(re.MapCell.Tile.X, re.MapCell.Tile.Y))
	}
}

func TestBaramosAftermathStoryFlagWriter(t *testing.T) {
	g := r4Game(t)
	trigger := bossTriggerAt(65, 0, 8, 3)
	if trigger == nil {
		t.Fatal("缺巴拉摩斯 trigger")
	}
	g.dnPhase, g.dnStep = 2, 37
	g.setStoryFlag(trigger.clearStoryFlag, true)
	g.applyBossAftermathStoryFlag(trigger)
	if !g.storyFlag(0x1a) || !g.storyFlag(0x1b) || g.storyFlag(trigger.clearStoryFlag) || g.dnPhase != 0 || g.dnStep != 0 {
		t.Fatalf("handler70 aftermath 錯: f1a=%v f1b=%v f29=%v phase=%d step=%d",
			g.storyFlag(0x1a), g.storyFlag(0x1b), g.storyFlag(trigger.clearStoryFlag), g.dnPhase, g.dnStep)
	}
}

func TestSamanosaConditionalStoryFlagWriter(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	e, ok := g.storyFlagRuntimeEvent("conditional_story_flag")
	if !ok {
		t.Fatal("缺 conditional_story_flag pack primitive")
	}
	g.setStoryFlag(e.ClearStoryFlagsRaw[0], true)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, e.NPC.CTYRaw,
		mapBlkNum[e.NPC.CTYRaw], e.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	idx := sc.npcAt(e.NPC.Tile.X, e.NPC.Tile.Y)
	if idx < 0 || sc.npcs[idx].b4 != e.NPC.HandlerRaw {
		t.Fatalf("CTY44 handler33 NPC 不存在: idx=%d", idx)
	}
	g.cur, g.town, g.inTown, g.curCty, g.dlg.tx = sc, sc, true, e.NPC.CTYRaw, sc.dlgText
	if rec := sc.dlgText.Record(e.DialogueRecordRaw); len(rec) == 0 {
		t.Fatalf("CTY44 dialogue bank 缺 record %d", e.DialogueRecordRaw)
	}
	// 正式玩家入口：站在 handler33 NPC 下方，開命令窗後選「話す」；不直接呼叫
	// handler，確保 selector、facing 與 command dispatch 都在同一條路徑。
	g.px, g.py, g.facing = e.NPC.Tile.X, e.NPC.Tile.Y+1, 1
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.storyFlag(0x24) || !g.storyFlag(0x21) || g.storyFlag(0x3d) || !g.dlg.open {
		t.Fatalf("handler33 transaction 錯: f24=%v f21=%v f3d=%v dlg=%v",
			g.storyFlag(0x24), g.storyFlag(0x21), g.storyFlag(0x3d), g.dlg.open)
	}
}
