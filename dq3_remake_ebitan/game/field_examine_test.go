package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldExamineDosgolemNormalInputComparison(t *testing.T) {
	runFieldItemActionNormalAt394(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-search-return-normal-r1"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "3af2954a524f521670aad308377efd2aca4c28633a971c111f7af6fc709c44eb" {
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
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 404 || len(src.Artifacts) != 1022 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(fieldSearchArtifactPath(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		inputs := []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: 0}, {Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3}}
		scans := []string{"39", "50", "50", "50", "50", "50", "39", "1c", "4b", "4d"}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		before, rng := g.snapshot(), g.prng
		stored := make([][]byte, g.fieldSaveLoad.contract.SlotCount)
		for j := range stored {
			stored[j], e = os.ReadFile(fieldSaveSlotPath(j))
			if e != nil {
				t.Fatal(e)
			}
		}
		var samples []map[string]any
		for i, in := range inputs {
			n := 395 + i
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
			if n == 401 {
				if g.fieldMessagePrompt == nil {
					g.renderFrame()
					f, err := os.Create(filepath.Join(dest, "search-empty-red-401.png"))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
					ce := f.Close()
					if err != nil || ce != nil {
						t.Fatal(err, ce)
					}
					diff := sourceCanvasDifference(t, g, source, prefix+"-packet-401-waiting.png")
					t.Fatalf("empty examine did not present native response; full640x350 RGB difference=%d", diff)
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
			if n == 403 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("empty examine changed state", n)
			}
			for j, b := range stored {
				actual, err := os.ReadFile(fieldSaveSlotPath(j))
				if err != nil || !bytes.Equal(b, actual) {
					t.Fatal("examine wrote a slot", n, j, err)
				}
			}
			if n >= 402 && (g.fieldMessagePrompt != nil || g.fieldSpell.active || g.cmd.open || g.panel != panelNone) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("search-empty-packet-%03d.png", n))
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
			t.Logf("normal search-empty packet%d full640x350 RGB difference=%d", n, diff)
			if n <= 400 && diff != 0 {
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
		report := map[string]any{"scope": "normal new-game through404 single-member empty examine, fresh-key return and next moves", "source_sha256": "3af2954a524f521670aad308377efd2aca4c28633a971c111f7af6fc709c44eb", "samples": samples, "game_state_injection": false, "rng_unchanged": true, "normal_save_load_roundtrip": true, "save_checkpoint_verified": true, "load_clock": clock, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "search-empty-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}

	})
}

func fieldSearchArtifactPath(dir, path string) string {
	if filepath.Dir(path) != "." {
		return filepath.Join(filepath.Dir(dir), path)
	}
	return filepath.Join(dir, path)
}

func TestFieldExamineFreshKeyAndScope(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	before, rng := g.snapshot(), g.prng
	g.selectCommand(cmdExamine)
	if g.fieldMessagePrompt == nil {
		t.Fatal("empty response absent")
	}
	if err := g.step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true}); err != nil {
		t.Fatal(err)
	}
	if g.fieldSaveLoad.active || g.fieldMessagePrompt == nil {
		t.Fatal("early key leaked")
	}
	for i := 0; !g.fieldMessagePrompt.dialogue.waitingForConfirm(); i++ {
		if i >= 1000 {
			t.Fatal("fresh wait")
		}
		if err := g.step(idle); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.step(InputState{DirHeld: 2, DirEdge: 2, SaveMenu: true, AnyKeyEdge: true}); err != nil {
		t.Fatal(err)
	}
	if g.fieldMessagePrompt != nil || g.fieldSaveLoad.active || !equalFieldSave(before, g.snapshot()) || g.prng != rng {
		t.Fatal("dismissal leaked movement, save or transaction")
	}
	g.selectCommand(cmdExamine)
	g.restore(before)
	if g.fieldMessagePrompt != nil {
		t.Fatal("restored transient response")
	}
	g.heroConditions = conditionPoison
	g.beginEmptyExamine()
	if g.fieldMessagePrompt != nil {
		t.Fatal("unreviewed condition branch presented")
	}
	g.heroConditions = 0
	g.shipAboard = true
	g.beginEmptyExamine()
	if g.fieldMessagePrompt != nil {
		t.Fatal("unreviewed ship response presented")
	}
}
