package game

import (
	"bytes"
	"encoding/json"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

func TestDosgolemRegionDialogueReturnNormalInput(t *testing.T) {
	path := os.Getenv("DQ3_MOTHER_RETURN_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_MOTHER_RETURN_ORIGINAL to the audited cold normal39-input receipt")
	}
	prefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	pressDown := func() {
		t.Helper()
		for g.cd > 0 {
			step(idle)
		}
		step(InputState{DirHeld: 0, DirEdge: 0, AnyKeyEdge: true})
	}
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(g.snapshot())
	pressDown()
	if g.px != 21 || g.py != 18 || g.regionDialogueReturn == nil {
		t.Fatal("normal Down did not start native warning")
	}
	if err := g.Save(); err == nil {
		t.Fatal("saved an incomplete forced return")
	}
	for i := 0; i < 2000 && g.regionDialogueReturn != nil && g.regionDialogueReturn.phase != 2; i++ {
		step(idle)
	}
	if g.regionDialogueReturn == nil || g.regionDialogueReturn.phase != 2 || !g.dlg.open || g.dlg.waitingForConfirm() {
		t.Fatal("automatic text did not reach final hold")
	}
	g.renderFrame()
	full := sourceCanvasDifference(t, g, path, prefix+"-pc-10232.png")
	f, err := os.Open(filepath.Join(filepath.Dir(path), prefix+"-pc-10232.png"))
	if err != nil {
		t.Fatal(err)
	}
	original, err := png.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	w, s := g.dlg.layout, g.dlg.shadow
	panel := 0
	for y := w.Y; y < w.Y+w.Height+s.OffsetY; y++ {
		for x := w.X; x < w.X+w.Width+s.OffsetX; x++ {
			r, gg, b, _ := original.At(x, y).RGBA()
			o := (y*ScreenW + x) * 4
			if uint8(r>>8) != g.rgba[o] || uint8(gg>>8) != g.rgba[o+1] || uint8(b>>8) != g.rgba[o+2] {
				panel++
			}
		}
	}
	if panel != 0 {
		t.Fatalf("native warning window/shadow RGB differs at %d pixels; full=%d", panel, full)
	}
	out := os.Getenv("DQ3_MOTHER_RETURN_OUT")
	capture := func(suffix string) {
		t.Helper()
		if out == "" {
			return
		}
		f, err := os.Create(out + suffix + ".png")
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
		ce := f.Close()
		if err != nil || ce != nil {
			t.Fatal(err, ce)
		}
	}
	capture("-warning")
	for i := 0; i < 100 && g.regionDialogueReturn != nil; i++ {
		step(idle)
	}
	if g.regionDialogueReturn != nil || g.dlg.open || g.dlg.prelude != nil || g.px != 21 || g.py != 17 || !g.storyFlag(23) || g.storyFlag(24) || g.heroGold != 0 || len(testInventory(g.items)) != 0 {
		t.Fatal("native automatic return transaction differs")
	}
	g.renderFrame()
	afterFull := sourceCanvasDifference(t, g, path, prefix+"-pc-1024b.png")
	capture("-returned")
	// Retry from the same normal checkpoint, then load the real pre-event save.
	pressDown()
	if g.regionDialogueReturn == nil {
		t.Fatal("repeat warning did not trigger")
	}
	if err := g.Load(); err != nil {
		t.Fatal(err)
	}
	restored, _ := json.Marshal(g.snapshot())
	if g.regionDialogueReturn != nil || g.dlg.open || !bytes.Equal(before, restored) {
		t.Fatal("Load retained pending forced move or changed checkpoint")
	}
	pressDown()
	for i := 0; i < 2200 && g.regionDialogueReturn != nil; i++ {
		step(idle)
	}
	if g.py != 17 || g.regionDialogueReturn != nil {
		t.Fatal("normal repeat did not return")
	}
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	saved, _ := json.Marshal(g.snapshot())
	loaded, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []InputState{{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}} {
		if err := loaded.step(in); err != nil {
			t.Fatal(err)
		}
	}
	actual, _ := json.Marshal(loaded.snapshot())
	if loaded.showTitle || !bytes.Equal(saved, actual) {
		t.Fatal("normal title Load changed completed return")
	}
	for loaded.cd > 0 {
		if err := loaded.step(idle); err != nil {
			t.Fatal(err)
		}
	}
	if err := loaded.step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}); err != nil {
		t.Fatal(err)
	}
	if loaded.px != 21 || loaded.py != 16 {
		t.Fatal("normal north continuation blocked")
	}
	if out != "" {
		b, _ := json.MarshalIndent(map[string]any{"schema": g.pack.Schema(), "content_version": g.pack.ContentVersion(), "content_hash": g.pack.ContentHash(), "normal_inputs": 39, "warning_full_rgb_difference": full, "warning_window_shadow_rgb_difference": panel, "return_full_rgb_difference": afterFull, "repeat_trigger": true, "same_instance_load_clears_pending": true, "title_save_load": true, "north_continuation": true, "flags_inventory_gold_unchanged": true, "original_save_load_parity": false, "audio_parity": false}, "", "  ")
		if err := os.WriteFile(out+".json", append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("normal native warning window RGB=%d; full=%d; return full=%d; repeat, title Load and north continuation PASS", panel, full, afterFull)
}

func TestRegionDialogueReturnClearedFlagDoesNotTrigger(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !g.finishMotherEscort() {
		t.Fatal("scene fixture missing")
	}
	e := g.pack.Events.RegionDialogueReturnEvents[0]
	g.setStoryFlag(e.RequiredFlagRaw, false)
	g.px, g.py = 21, 18
	active, err := g.tryRegionDialogueReturn()
	if err != nil || active || g.regionDialogueReturn != nil || g.dlg.open {
		t.Fatal("cleared native gate still triggered", err)
	}
}

func TestRegionDialogueReturnRejectsBadSources(t *testing.T) {
	for _, name := range []string{"nil", "missing_scene", "truncated_scene", "wrong_handler", "outside_actor", "missing_text", "changed_text"} {
		t.Run(name, func(t *testing.T) {
			p, err := gamepack.BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			e := &p.Events.RegionDialogueReturnEvents[0]
			d, _ := p.TextDefinition(e.TextID)
			raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), ctyFile(e.CTYRaw)))
			if err != nil {
				t.Fatal(err)
			}
			text, err := os.ReadFile(filepath.Join(spineAssetsDir(t), d.Source.File))
			if err != nil {
				t.Fatal(err)
			}
			broken := fstest.MapFS{ctyFile(e.CTYRaw): &fstest.MapFile{Data: raw}, d.Source.File: &fstest.MapFile{Data: text}}
			switch name {
			case "nil":
				if validateRegionDialogueReturnSources(nil, p) == nil {
					t.Fatal("accepted nil source")
				}
				return
			case "missing_scene":
				delete(broken, ctyFile(e.CTYRaw))
			case "truncated_scene":
				broken[ctyFile(e.CTYRaw)].Data = raw[:2]
			case "wrong_handler":
				e.HandlerRaw = 255
			case "outside_actor":
				e.ActorRecordRaw = 65535
			case "missing_text":
				delete(broken, d.Source.File)
			case "changed_text":
				d.GlyphCodes[0]++
			}
			if validateRegionDialogueReturnSources(broken, p) == nil {
				t.Fatal("accepted incompatible source")
			}
		})
	}
}
