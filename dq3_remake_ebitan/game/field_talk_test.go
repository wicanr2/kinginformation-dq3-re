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

func TestFieldTalkDosgolemNormalInputComparison(t *testing.T) {
	runFieldExamineNormalAt404(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-talk-empty-return-normal-r1"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "753da910efcb614851eb707a8e35bd22dfc1e978a9347ba078a067208ff197eb" {
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
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 409 || len(src.Artifacts) != 1037 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(fieldSearchArtifactPath(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		inputs := []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3}}
		scans := []string{"39", "39", "1c", "4b", "4d"}
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
			n := 405 + i
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
			if n == 406 {
				if g.fieldMessagePrompt == nil {
					g.renderFrame()
					f, err := os.Create(filepath.Join(dest, "talk-empty-red-406.png"))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
					ce := f.Close()
					if err != nil || ce != nil {
						t.Fatal(err, ce)
					}
					diff := sourceCanvasDifference(t, g, source, prefix+"-packet-406-waiting.png")
					t.Fatalf("no-target talk did not present native response; full640x350 RGB difference=%d", diff)
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
			if n == 408 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("no-target talk changed state", n)
			}
			for j, b := range stored {
				actual, err := os.ReadFile(fieldSaveSlotPath(j))
				if err != nil || !bytes.Equal(b, actual) {
					t.Fatal("examine wrote a slot", n, j, err)
				}
			}
			if n >= 407 && (g.fieldMessagePrompt != nil || g.fieldSpell.active || g.cmd.open || g.panel != panelNone) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("talk-empty-packet-%03d.png", n))
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
			t.Logf("normal talk-empty packet%d full640x350 RGB difference=%d", n, diff)
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			var visuals []map[string]int
			for _, npc := range g.cur.npcs {
				visuals = append(visuals, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "hero_facing": g.facing, "hero_walk": g.walk, "npc_visuals": visuals})
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
		report := map[string]any{"scope": "normal new-game through409 healthy single-member town no-target talk, fresh-key return and next moves", "source_sha256": "753da910efcb614851eb707a8e35bd22dfc1e978a9347ba078a067208ff197eb", "samples": samples, "game_state_injection": false, "rng_unchanged": true, "normal_save_load_roundtrip": true, "save_checkpoint_verified": true, "load_clock": clock, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "talk-empty-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}

	})
}

func TestFieldTalkFreshKeyScopeAndNPC(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	// Component fixture only. Normal409 owns original-versus-remake path parity.
	if g.cur == nil {
		t.Fatal("town fixture absent")
	}
	g.inTown = true
	g.cur.npcs = nil
	g.renderFrame()
	before, rng := g.snapshot(), g.prng
	idle := InputState{DirHeld: -1, DirEdge: -1}
	g.selectCommand(cmdTalk)
	if g.fieldMessagePrompt == nil {
		t.Fatal("no-target response absent")
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
		t.Fatal("dismissal leaked movement/save/transaction")
	}
	g.selectCommand(cmdTalk)
	g.restore(before)
	if g.fieldMessagePrompt != nil {
		t.Fatal("restored transient prompt")
	}
	for _, branch := range []string{"condition", "ship", "party", "world"} {
		t.Run(branch, func(t *testing.T) {
			g.heroConditions = 0
			g.shipAboard = false
			g.companions = nil
			g.inTown = true
			switch branch {
			case "condition":
				g.heroConditions = conditionPoison
			case "ship":
				g.shipAboard = true
			case "party":
				g.companions = append(g.companions, &Member{})
			case "world":
				g.inTown = false
			}
			g.beginEmptyTalk()
			if g.fieldMessagePrompt != nil {
				t.Fatal("unreviewed response presented")
			}
		})
	}
	// Check the real target lookup and NPC-facing update with a component NPC.
	g.heroConditions = 0
	g.shipAboard = false
	g.companions = nil
	g.inTown = true
	dx, dy := dirDelta(g.facing)
	g.cur.npcs = []npcInst{{x: g.px + dx, y: g.py + dy, b4: 0, ctrl: 0}}
	g.selectCommand(cmdTalk)
	if g.fieldMessagePrompt != nil || g.cur.npcs[0].facing != g.facing^1 {
		t.Fatal("targeted NPC routed to no-target response")
	}
}
