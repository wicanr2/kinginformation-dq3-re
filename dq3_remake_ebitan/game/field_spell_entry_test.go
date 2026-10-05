package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldSpellEntryDosgolemNormalInputComparison(t *testing.T) {
	runFieldStatusReorderNormalAt298(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-spell-empty-normal-r1"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "c096702865ce694706ba2aac2f9333408b895bf45d856bae4a14334d1275cfcf" {
			t.Fatal("source identity", e)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 304 || len(src.Artifacts) != 720 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(filepath.Join(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		inputs := []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 3}, {Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3}}
		scans := []string{"39", "4d", "39", "1c", "4b", "4d"}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		before, rng := g.snapshot(), g.prng
		var samples []map[string]any
		for i, in := range inputs {
			n := 299 + i
			if src.Queued[n-1]["scan"] != scans[i] {
				t.Fatal("input", n)
			}
			in.AnyKeyEdge = true
			for g.cd > 0 {
				if e = g.step(idle); e != nil {
					t.Fatal(e)
				}
			}
			if e = g.step(in); e != nil {
				t.Fatal(e)
			}
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
			if n == 301 {
				if g.fieldMessagePrompt == nil || g.fieldSpell.active {
					t.Fatal("empty entry did not present native response")
				}
				for j := 0; !g.fieldMessagePrompt.dialogue.waitingForConfirm(); j++ {
					if j >= 1000 {
						t.Fatal("wait")
					}
					if e = g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
			}
			want := before
			if n == 303 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("empty spell changed state", n)
			}
			if n >= 302 && (g.fieldMessagePrompt != nil || g.fieldSpell.active || g.cmd.open || g.panel != panelNone) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("spell-empty-packet-%03d.png", n))
			f, e := os.Create(p)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			t.Logf("normal spell-empty packet%d RGB difference=%d caster_selector=%t", n, diff, g.fieldSpell.selectingCaster)
			if diff != 0 {
				t.Fatal("normal complete640x350 differs", n, diff)
			}
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))})
		}
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "field-save.json"))
		expectedSave := before
		expectedSave.Respawn = respawnToSave(g.currentRespawnPoint())
		step := func(in InputState) {
			t.Helper()
			if e := g.step(in); e != nil {
				t.Fatal(e)
			}
			settleFieldSaveLoad(t, g)
		}
		enter := InputState{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}
		step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true})
		step(enter)
		step(enter)
		if g.fieldSaveLoad.stage != fsFinalWait {
			t.Fatal("normal save final wait")
		}
		step(enter)
		if g.fieldSaveLoad.active {
			t.Fatal("save did not return")
		}
		written, err := os.ReadFile(savePath())
		if err != nil {
			t.Fatal(err)
		}
		saved, err := decodeSave(written)
		if err != nil || !equalFieldSave(expectedSave, saved) || !equalFieldSave(expectedSave, g.snapshot()) {
			t.Fatal("normal save checkpoint transaction", err)
		}
		clock, ok := g.pack.Interface.FieldSaveLoad.LoadClock(saved.Layer)
		phaseTicks := g.dayNightCycle.ClockTicks / 4
		if !ok || phaseTicks <= 0 {
			t.Fatal("normal load clock contract")
		}
		expectedLoad := expectedSave
		expectedLoad.DNPhase, expectedLoad.DNStep = clock/phaseTicks, clock%phaseTicks
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true, AnyKeyEdge: true})
		step(enter)
		if g.fieldSaveLoad.active || !equalFieldSave(expectedLoad, g.snapshot()) || g.prng != rng || g.fieldMessagePrompt != nil {
			x, _ := encodeSave(expectedLoad)
			y, _ := encodeSave(g.snapshot())
			var a, b map[string]json.RawMessage
			json.Unmarshal(x, &a)
			json.Unmarshal(y, &b)
			for k, v := range a {
				if string(v) != string(b[k]) {
					t.Logf("load difference %s expected=%s actual=%s", k, v, b[k])
				}
			}
			t.Fatalf("normal save/load roundtrip: active=%t state_equal=%t rng_equal=%t prompt=%t", g.fieldSaveLoad.active, equalFieldSave(expectedLoad, g.snapshot()), g.prng == rng, g.fieldMessagePrompt != nil)
		}
		for g.cd > 0 {
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		if e = g.step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true}); e != nil {
			t.Fatal(e)
		}
		if g.px != 2 || g.py != before.PY {
			t.Fatal("movement after load")
		}
		report := map[string]any{"scope": "normal new-game through304 single-member empty spell, fresh-key return and next moves", "source_sha256": "c096702865ce694706ba2aac2f9333408b895bf45d856bae4a14334d1275cfcf", "samples": samples, "game_state_injection": false, "rng_unchanged": true, "normal_save_load_roundtrip": true, "save_checkpoint_verified": true, "load_clock": clock, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "spell-empty-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}

	})
}

func TestFieldSpellEntryFreshKeyAndScope(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	before, rng := g.snapshot(), g.prng
	g.selectCommand(cmdSpell)
	if g.fieldMessagePrompt == nil || g.fieldSpell.active {
		t.Fatal("empty message absent")
	}
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	if g.fieldSaveLoad.active || g.fieldMessagePrompt == nil {
		t.Fatal("early key leaked")
	}
	for i := 0; !g.fieldMessagePrompt.dialogue.waitingForConfirm(); i++ {
		if i >= 1000 {
			t.Fatal("fresh wait")
		}
		if e := g.step(idle); e != nil {
			t.Fatal(e)
		}
	}
	if e := g.step(InputState{DirHeld: 2, DirEdge: 2, SaveMenu: true, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	if g.fieldMessagePrompt != nil || g.fieldSaveLoad.active || !equalFieldSave(before, g.snapshot()) || g.prng != rng {
		t.Fatal("dismissal action or transaction")
	}
	g.selectCommand(cmdSpell)
	g.restore(before)
	if g.fieldMessagePrompt != nil {
		t.Fatal("restored transient message")
	}
	g.heroConditions = conditionPoison
	g.openFieldSpellMenu()
	if !g.fieldSpell.selectingCaster {
		t.Fatal("unreviewed branch altered")
	}
}
