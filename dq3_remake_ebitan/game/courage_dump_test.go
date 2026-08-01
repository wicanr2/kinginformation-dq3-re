package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpLancelCourageTrial records current Ebitengine runtime evidence for
// the pack-driven offer and the solo blue-orb result. Normal reachability is
// independently covered by TestOpeningProductionInputTrace.
func TestDumpLancelCourageTrial(t *testing.T) {
	if os.Getenv("DQ3_DUMP_COURAGE") == "" {
		t.Skip("設 DQ3_DUMP_COURAGE=1 才執行")
	}
	out := os.Getenv("COURAGE_OUT")
	if out == "" {
		t.Fatal("需設 COURAGE_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	assets := os.Getenv("DQ3_ASSETS")
	if assets == "" {
		assets = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(assets), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	event, ok := g.pack.TemporarySoloChallenge("dq3:event.lancel_courage_trial")
	if !ok {
		t.Fatal("缺 dq3:event.lancel_courage_trial")
	}
	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
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

	priest := loadTemporaryRoleScene(t, g, event.EntryNPC)
	g.px, g.py, g.facing = event.EntryNPC.Tile.X, event.EntryNPC.Tile.Y+1, 1
	if !g.talkTemporarySoloChallenge(priest) || !g.dlg.open {
		t.Fatal("勇氣試煉畫面 fixture 未開啟 prompt")
	}
	dump("lancel_courage_trial_prompt")

	// Recreate the original post-handler37 solo state, then use the canonical
	// CTY23 event selector for the visible blue-orb acquisition result.
	g.dlg.open = false
	g.soloChallengeEventID = event.ID
	g.companions = []*Member{{Name: []int{101}}, {Name: []int{102}}, {Name: []int{103}}}
	g.beginSoloChallenge(event)
	treasure, ok := g.pack.TreasureEvent("dq3:event.courage_cave_blue_orb")
	if !ok {
		t.Fatal("缺 dq3:event.courage_cave_blue_orb")
	}
	tr := treasure.Treasure
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		tr.CTYRaw, mapBlkNum[tr.CTYRaw], tr.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, tr.CTYRaw, true
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
	}
	found := false
	for y := 0; y < sc.h && !found; y++ {
		for x := 0; x < sc.w && !found; x++ {
			ev, subid, ok := sc.tileEvent(x, y)
			if !ok || subid != tr.TileSubID || ev[0] != tr.EventTypeRaw ||
				ev[1] != tr.ItemRawID || ev[2] != tr.PresentFlag {
				continue
			}
			g.px, g.py = x, y
			found = true
		}
	}
	if !found || !g.collectPackTreasure(tr) {
		t.Fatal("勇氣洞窟畫面 fixture 找不到藍寶珠 event")
	}
	sc.revealEventTile(g.px, g.py)
	dump("courage_cave_blue_orb_obtained")
}
