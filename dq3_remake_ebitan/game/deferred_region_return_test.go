package game

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestDosgolemDeferredRegionReturnNormalInput(t *testing.T) {
	path := os.Getenv("DQ3_FIRST_MOVE_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_FIRST_MOVE_ORIGINAL to the audited cold first-up receipt")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		NormalInputs int                 `json:"normal_inputs"`
		States       []map[string]string `json:"states"`
	}
	if err := json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.NormalInputs != 41 || len(original.States) != 3 {
		t.Fatal("invalid original first-move scope")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	up := func() {
		t.Helper()
		for g.cd > 0 {
			step(idle)
		}
		step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true})
	}
	if g.deferredRegionDialogueReturnID == "" {
		t.Fatal("normal escort discarded its native pending selector")
	}
	before := g.snapshot()
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	saved, _ := json.Marshal(before)
	up()
	if g.px != 21 || g.py != 16 || g.regionDialogueReturn == nil || g.deferredRegionDialogueReturnID != "" {
		t.Fatal("normal first Up did not consume the native pending event")
	}
	for i := 0; i < 2200 && g.regionDialogueReturn != nil && g.regionDialogueReturn.phase != 2; i++ {
		step(idle)
	}
	if g.regionDialogueReturn == nil || g.regionDialogueReturn.phase != 2 {
		t.Fatal("automatic warning did not complete")
	}
	g.renderFrame()
	prefix := "issue4-registry-quiescent-r2"
	full := sourceCanvasDifference(t, g, path, prefix+"-event-10232.png")
	f, err := os.Open(filepath.Join(filepath.Dir(path), prefix+"-event-10232.png"))
	if err != nil {
		t.Fatal(err)
	}
	orig, err := png.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	w, s := g.dlg.layout, g.dlg.shadow
	panel := 0
	for y := w.Y; y < w.Y+w.Height+s.OffsetY; y++ {
		for x := w.X; x < w.X+w.Width+s.OffsetX; x++ {
			r, gg, b, _ := orig.At(x, y).RGBA()
			o := (y*ScreenW + x) * 4
			if uint8(r>>8) != g.rgba[o] || uint8(gg>>8) != g.rgba[o+1] || uint8(b>>8) != g.rgba[o+2] {
				panel++
			}
		}
	}
	if panel != 0 {
		t.Fatalf("first Up warning window/shadow RGB differs=%d full=%d", panel, full)
	}
	out := os.Getenv("DQ3_FIRST_MOVE_OUT")
	capture := func(label string) {
		t.Helper()
		if out == "" {
			return
		}
		f, err := os.Create(out + "-" + label + ".png")
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatal(err, closeErr)
		}
	}
	capture("warning")
	for i := 0; i < 100 && g.regionDialogueReturn != nil; i++ {
		step(idle)
	}
	differences := []int{}
	for i, state := range original.States {
		if i > 0 {
			up()
		}
		wantX, err := strconv.Atoi(state["player_x"])
		if err != nil {
			t.Fatal(err)
		}
		wantY, err := strconv.Atoi(state["player_y"])
		if err != nil {
			t.Fatal(err)
		}
		if g.px != wantX || g.py != wantY || g.regionDialogueReturn != nil || g.deferredRegionDialogueReturnID != "" || g.dlg.open {
			t.Fatalf("normal Up%d state differs: got=%d,%d want=%d,%d", i+1, g.px, g.py, wantX, wantY)
		}
		if g.heroGold != before.HeroGold || !bytes.Equal(g.storyBits[:], before.StoryBits) {
			t.Fatal("warning changed story or gold")
		}
		inv, _ := json.Marshal(g.inventory)
		oldInv, _ := json.Marshal(before.Inventory)
		if !bytes.Equal(inv, oldInv) {
			t.Fatal("warning changed inventory")
		}
		g.renderFrame()
		differences = append(differences, sourceCanvasDifference(t, g, path, prefix+"-packet-00"+strconv.Itoa(i+1)+"-ready.png"))
		capture("up" + strconv.Itoa(i+1))
	}
	// Replay the real pre-move save separately from the cold input comparison.
	if err := g.Load(); err != nil {
		t.Fatal(err)
	}
	restored, _ := json.Marshal(g.snapshot())
	if !bytes.Equal(saved, restored) {
		t.Fatal("Load changed unconsumed normal checkpoint")
	}
	up()
	for i := 0; i < 2300 && g.regionDialogueReturn != nil; i++ {
		step(idle)
	}
	if g.px != 21 || g.py != 15 || g.regionDialogueReturn != nil || g.deferredRegionDialogueReturnID != "" {
		t.Fatal("pending Load did not preserve one native return")
	}
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	if err := g.Load(); err != nil {
		t.Fatal(err)
	}
	if g.deferredRegionDialogueReturnID != "" {
		t.Fatal("Load regenerated consumed event")
	}
	up()
	if g.py != 14 || g.regionDialogueReturn != nil {
		t.Fatal("post-Load continuation repeated consumed event")
	}
	if out != "" {
		report := map[string]any{"normal_inputs": 41, "content_version": g.pack.ContentVersion(), "content_hash": g.pack.ContentHash(), "warning_full_rgb_difference": full, "warning_window_shadow_rgb_difference": panel, "up_full_rgb_differences": differences, "pending_save_load": true, "consumed_save_load": true, "no_repeat": true, "original_save_load_parity": false, "audio_parity": false}
		b, _ := json.MarshalIndent(report, "", "  ")
		if err := os.WriteFile(out+".json", append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("native first three Up, pending/consumed Load PASS; window RGB=%d full=%d; states full=%v", panel, full, differences)
}

func TestDeferredRegionReturnRejectsBadSaves(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.motherEscort()
	e := g.pack.Events.RegionDialogueReturnEvents[0]
	g.deferredRegionDialogueReturnID = e.ID
	valid := g.snapshot()
	if err := g.validateDeferredRegionDialogueReturnSave(valid); err != nil {
		t.Fatal(err)
	}
	for _, kind := range []string{"unknown_id", "wrong_scene", "missing_pack", "outside_town", "home_pending", "cleared_gate", "escort_incomplete"} {
		t.Run(kind, func(t *testing.T) {
			s := valid
			s.StoryBits = append([]byte(nil), valid.StoryBits...)
			switch kind {
			case "unknown_id":
				s.DeferredRegionDialogueReturnID = "unknown"
			case "wrong_scene":
				s.Section++
			case "missing_pack":
				s.PackID = ""
			case "outside_town":
				s.InTown = false
			case "home_pending":
				s.OpeningHomeAwait = true
			case "cleared_gate":
				s.StoryBits[e.RequiredFlagRaw/8] &= ^(128 >> uint(e.RequiredFlagRaw%8))
			case "escort_incomplete":
				escort, ok := g.pack.OpeningEscort()
				if !ok || len(escort.ClearStoryFlags) == 0 {
					t.Fatal("source escort missing")
				}
				flag := escort.ClearStoryFlags[0]
				s.StoryBits[flag/8] |= 128 >> uint(flag%8)
			}
			b, _ := encodeSave(s)
			t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "save.json"))
			if err := os.WriteFile(savePath(), b, 0644); err != nil {
				t.Fatal(err)
			}
			before, _ := json.Marshal(g.snapshot())
			if err := g.Load(); err == nil {
				t.Fatal("accepted invalid deferred event")
			}
			after, _ := json.Marshal(g.snapshot())
			if !bytes.Equal(before, after) {
				t.Fatal("failed Load changed state")
			}
		})
	}
}

func TestDeferredRegionReturnBlockedInputPreservesSelection(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.motherEscort()
	g.showTitle = false
	e := g.pack.Events.RegionDialogueReturnEvents[0]
	g.deferredRegionDialogueReturnID = e.ID
	for y := 1; y < g.cur.h-1; y++ {
		for x := 1; x < g.cur.w-1; x++ {
			g.px, g.py = x, y
			if !g.tryMove(x, y) || g.tryMove(x, y-1) {
				continue
			}
			if err := g.step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}); err != nil {
				t.Fatal(err)
			}
			if g.px != x || g.py != y || g.deferredRegionDialogueReturnID != e.ID || g.regionDialogueReturn != nil {
				t.Fatal("blocked normal input consumed selection")
			}
			return
		}
	}
	t.Fatal("source map has no blocked north neighbour")
}

func TestDeferredRegionReturnDispatchClearsSelection(t *testing.T) {
	for _, kind := range []string{"cleared_gate", "new_special"} {
		t.Run(kind, func(t *testing.T) {
			g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
			if err != nil {
				t.Fatal(err)
			}
			g.motherEscort()
			e := g.pack.Events.RegionDialogueReturnEvents[0]
			g.deferredRegionDialogueReturnID = e.ID
			if kind == "cleared_gate" {
				g.setStoryFlag(e.RequiredFlagRaw, false)
			} else {
				found := false
				for i, h := range g.cur.hiMap {
					subid := int(h & 31)
					if subid > 0 && subid <= len(g.cur.specialHandlers) && g.cur.specialHandlers[subid-1] != e.HandlerRaw {
						g.px, g.py = i%g.cur.w, i/g.cur.w
						found = true
						break
					}
				}
				if !found {
					t.Fatal("source map has no different special selector")
				}
			}
			active, err := g.tryRegionDialogueReturn()
			if err != nil || active || g.regionDialogueReturn != nil || g.deferredRegionDialogueReturnID != "" {
				t.Fatal("dispatch retained an invalid old selection", err)
			}
		})
	}
}
