package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func conditionalFacilityTestGame(t *testing.T) (*Game, int) {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	event, ok := g.pack.ConditionalFacilityEvent("dq3:event.merchant_settlement_shop_gate")
	if !ok {
		t.Fatal("missing CTY59 conditional shop event")
	}
	// Reproduce the post-founder stage that makes the CTY59 handler visible:
	// the common completion flag and one gender-specific visibility flag are set
	// by the original handler before the next scene load.
	founder, ok := g.pack.SettlementFounderEvent("dq3:event.merchant_settlement_founder")
	if !ok || founder == nil || len(founder.FounderVisibilityFlagsRaw) == 0 {
		t.Fatal("missing merchant founder visibility contract")
	}
	g.setStoryFlag(founder.CompletionFlagRaw, true)
	g.setStoryFlag(founder.FounderVisibilityFlagsRaw[0], true)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.NPC.CTYRaw, mapBlkNum[event.NPC.CTYRaw], event.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	if idx := sc.npcAt(event.NPC.Tile.X, event.NPC.Tile.Y); idx < 0 {
		for i, n := range sc.npcs {
			t.Logf("CTY59 sec0 NPC[%d] x=%d y=%d ctrl=%d b4=%d", i, n.x, n.y, n.ctrl, n.b4)
		}
		t.Fatalf("conditional facility selector NPC missing at (%d,%d) handler=%d", event.NPC.Tile.X, event.NPC.Tile.Y, event.NPC.HandlerRaw)
	}
	g.showTitle, g.inTown, g.curCty = false, true, event.NPC.CTYRaw
	g.town, g.cur, g.dlg.tx = sc, sc, sc.dlgText
	g.px, g.facing = event.NPC.Tile.X, 1
	return g, event.NPC.Tile.Y
}

func TestConditionalFacilityOriginalCoordinateGate(t *testing.T) {
	g, npcY := conditionalFacilityTestGame(t)

	// Below the original Y=4 facility gate, the scripted handler is a
	// repeatable D3TXT07 record-8 dialogue.
	g.py = npcY + 1
	g.selectCommand(cmdTalk)
	if !g.dlg.open || g.shop.active {
		t.Fatalf("handler41 fallback branch did not open dialogue: dlg=%v shop=%v",
			g.dlg.open, g.shop.active)
	}
	g.dlg.open = false

	// At Y=4 the same handler dispatches CTY59 facility index 0, the
	// two-item item shop.  The interaction still comes through InputState's
	// formal command path, not a direct event call.
	g.py = 4
	g.selectCommand(cmdTalk)
	if !g.shop.active || len(g.shop.codes) != 2 {
		t.Fatalf("handler41 facility branch did not open CTY59 shop: active=%v codes=%v",
			g.shop.active, g.shop.codes)
	}
}

func TestDumpConditionalFacilityOriginalBranches(t *testing.T) {
	if os.Getenv("DQ3_DUMP_MERCHANT_HANDLER41") == "" {
		t.Skip("設 DQ3_DUMP_MERCHANT_HANDLER41=1 才產出 handler41 核對圖")
	}
	out := os.Getenv("MERCHANT_HANDLER41_OUT")
	if out == "" {
		t.Fatal("需設 MERCHANT_HANDLER41_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, npcY := conditionalFacilityTestGame(t)
	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	g.py = npcY + 1
	g.selectCommand(cmdTalk)
	dump("merchant_settlement_handler41_dialogue")
	g.dlg.open = false
	g.py = 4
	g.selectCommand(cmdTalk)
	dump("merchant_settlement_handler41_shop")
}
