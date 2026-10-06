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
	"strconv"
	"testing"
)

type fieldTalkNormalCase struct {
	prefix, sourceHash, output, scope string
	count, artifacts                  int
	inputs                            []InputState
	scans                             []string
	nextDir, nextX, nextY             int
	hiddenRecords                     []int
}

func fieldTalkNormalInputs() []InputState {
	return []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3}}
}

func TestFieldTalkDosgolemNormalInputComparison(t *testing.T) {
	runFieldTalkNormalComparison(t, fieldTalkNormalCase{
		prefix: "issue4-talk-empty-return-normal-r1", sourceHash: "753da910efcb614851eb707a8e35bd22dfc1e978a9347ba078a067208ff197eb",
		output: "talk-empty", scope: "normal new-game through409 healthy single-member town no-target talk, fresh-key return and next moves",
		count: 409, artifacts: 1037, inputs: fieldTalkNormalInputs(), scans: []string{"39", "39", "1c", "4b", "4d"}, nextDir: 2, nextX: 2, nextY: 18,
	})
}

func TestFieldRoomDoorDosgolemNormalInputComparison(t *testing.T) {
	inputs := append(fieldTalkNormalInputs(), InputState{Enter: true, DirHeld: -1, DirEdge: -1})
	for i := 0; i < 4; i++ {
		inputs = append(inputs, InputState{DirHeld: 2, DirEdge: 2})
	}
	for i := 0; i < 3; i++ {
		inputs = append(inputs, InputState{DirHeld: 3, DirEdge: 3})
	}
	for i := 0; i < 5; i++ {
		inputs = append(inputs, InputState{DirHeld: 0, DirEdge: 0})
	}
	runFieldTalkNormalComparison(t, fieldTalkNormalCase{
		prefix: "issue4-field-room-door-normal-r1", sourceHash: "b7d0b1534a8b1055b57d78de8c343d12b736f355cf17584211b24dc2b84bcf3e",
		output: "room-door", scope: "normal new-game through422 field Enter, left wall, room door, normal save/load and next right move",
		count: 422, artifacts: 1076, inputs: inputs, scans: []string{"39", "39", "1c", "4b", "4d", "1c", "4b", "4b", "4b", "4b", "4d", "4d", "4d", "50", "50", "50", "50", "50"}, nextDir: 3, nextX: 5, nextY: 23,
		hiddenRecords: []int{14, 15},
	})
}

func TestNPCMotionDosgolemNormalInputComparison(t *testing.T) {
	inputs := append(fieldTalkNormalInputs(), InputState{Enter: true, DirHeld: -1, DirEdge: -1})
	scans := []string{"39", "39", "1c", "4b", "4d", "1c"}
	for _, leg := range []struct {
		dir, count int
		scan       string
	}{{2, 4, "4b"}, {3, 3, "4d"}, {0, 5, "50"}, {3, 13, "4d"}, {0, 5, "50"}, {2, 13, "4b"}, {1, 5, "48"}} {
		for j := 0; j < leg.count; j++ {
			inputs = append(inputs, InputState{DirHeld: leg.dir, DirEdge: leg.dir})
			scans = append(scans, leg.scan)
		}
	}
	runFieldTalkNormalComparison(t, fieldTalkNormalCase{
		prefix: "issue4-npc-move-continue-normal-r1", sourceHash: "56ef662ad02931486a60bbfd0820a518887e699320ec1c3ff131070734ba95be",
		output: "npc-motion", scope: "normal new-game through458 fixed field movement; ten slots unchanged; F5/F6 and next right movement; NPC/global RNG and full RGB parity unknown",
		count: 458, artifacts: 1184, inputs: inputs, scans: scans, nextDir: 3, nextX: 4, nextY: 28,
	})
}

func runFieldTalkNormalComparison(t *testing.T, scenario fieldTalkNormalCase) {
	t.Helper()
	if len(scenario.inputs) != scenario.count-404 || len(scenario.scans) != len(scenario.inputs) {
		t.Fatal("normal trace contract shape")
	}

	runFieldExamineNormalAt404(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		prefix := scenario.prefix
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != scenario.sourceHash {
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
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != scenario.count || len(src.Artifacts) != scenario.artifacts {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(fieldSearchArtifactPath(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		inputs, scans := scenario.inputs, scenario.scans
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
					f, err := os.Create(filepath.Join(dest, scenario.output+"-red-406.png"))
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
			want.PX, e = strconv.Atoi(src.States[n-1]["player_x"])
			if e != nil {
				t.Fatal(e)
			}
			want.PY, e = strconv.Atoi(src.States[n-1]["player_y"])
			if e != nil {
				t.Fatal(e)
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("field input changed state", n)
			}
			for j, b := range stored {
				actual, err := os.ReadFile(fieldSaveSlotPath(j))
				if err != nil || !bytes.Equal(b, actual) {
					t.Fatal("field input wrote a slot", n, j, err)
				}
			}
			if n >= 407 && (g.fieldMessagePrompt != nil || g.fieldSpell.active || g.cmd.open || g.panel != panelNone) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf(scenario.output+"-packet-%03d.png", n))
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
			if n == scenario.count && len(scenario.hiddenRecords) != 0 {
				assertNormalNPCBackground(t, g, filepath.Join(dir, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, src.States[n-1]["phase"])), scenario.hiddenRecords)
			}
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
		expectedSave := g.snapshot()
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
		if e = g.step(InputState{DirHeld: scenario.nextDir, DirEdge: scenario.nextDir, AnyKeyEdge: true}); e != nil {
			t.Fatal(e)
		}
		if g.px != scenario.nextX || g.py != scenario.nextY {
			t.Fatal("movement after load")
		}
		report := map[string]any{"scope": scenario.scope, "source_sha256": scenario.sourceHash, "samples": samples, "game_state_injection": false, "rng_unchanged": true, "normal_save_load_roundtrip": true, "save_checkpoint_verified": true, "load_clock": clock, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, scenario.output+"-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}

	})
}

// This checks every pixel in the two complete native NPC cells. The full canvas
// difference is recorded separately; no image is cropped, masked or replaced.
func assertNormalNPCBackground(t *testing.T, g *Game, original string, records []int) {
	t.Helper()
	f, err := os.Open(original)
	if err != nil {
		t.Fatal(err)
	}
	im, err := png.Decode(f)
	closeErr := f.Close()
	if err != nil || closeErr != nil || im.Bounds() != image.Rect(0, 0, ScreenW, ScreenH) {
		t.Fatal("native canvas", err, closeErr)
	}
	camera := g.activeSceneCamera()
	if camera == nil {
		t.Fatal("normal scene camera absent")
	}
	camX, camY := g.px-camera.AnchorX, g.py-camera.AnchorY
	for _, record := range records {
		found := false
		for _, npc := range g.cur.npcs {
			if npc.recordIndex != record {
				continue
			}
			found = true
			x, y := (npc.x-camX)*TileW, (npc.y-camY)*TileH
			if x < 0 || y < 0 || x+TileW > ScreenW || y+TileH > ScreenH {
				t.Fatal("native NPC cell outside sample", record)
			}
			for row := y; row < y+TileH; row++ {
				for col := x; col < x+TileW; col++ {
					r, gr, b, _ := im.At(col, row).RGBA()
					o := (row*ScreenW + col) * 4
					if g.rgba[o] != byte(r>>8) || g.rgba[o+1] != byte(gr>>8) || g.rgba[o+2] != byte(b>>8) {
						t.Fatalf("hidden NPC record%d differs from native background at %d,%d", record, col, row)
					}
				}
			}
		}
		if !found {
			t.Fatal("native NPC record absent", record)
		}
	}
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
