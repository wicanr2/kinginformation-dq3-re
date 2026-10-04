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

func TestFieldSecondSlotDosgolemNormalInputComparison(t *testing.T) {
	dir, out := os.Getenv("DQ3_SECOND_SLOT_ORACLE_DIR"), os.Getenv("DQ3_SECOND_SLOT_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem second slot")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	const prefix = "issue4-second-slot-r1"
	const sourceHash = "1a5a7c22c4599ce1bf4f6070d178a6d24ee49a8bbc4f374e84e39459a7a3f9d2"
	path := filepath.Join(dir, prefix+"-source-r1-receipt.json")
	raw, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("source identity differs", err)
	}
	var src struct{ Queued, States, Clocks []map[string]string }
	if err := json.Unmarshal(raw, &src); err != nil || len(src.Queued) != 204 || len(src.States) != 204 || len(src.Clocks) != 12 {
		t.Fatal("source shape", err)
	}
	dest := filepath.Join(out, "second-slot")
	if err := os.Mkdir(dest, 0755); err != nil {
		t.Fatal(err)
	}
	prefixOut := filepath.Join(dest, "prefix")
	if err := os.Mkdir(prefixOut, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_ORACLE_DIR", dir)
	t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_RECEIPT_DIR", prefixOut)
	runRecruitmentEmptyNormalInputComparisonAtCheckpoint(t, "menu-cancel", "issue4-recruit-menu-esc-normal-r1-source-r1-receipt.json", "137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c", "issue4-recruit-menu-esc-normal-r1", 193, 189, 190, func(g *Game) {
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "field-save.json"))
		before, rng := g.snapshot(), g.prng
		metadata := prepareFieldSaveSlotMetadata(t, g, before)
		if !equalFieldSave(before, g.snapshot()) || g.prng != rng {
			t.Fatal("external fixture changed running game")
		}
		stored := make([][]byte, len(metadata))
		for i := range stored {
			stored[i], err = os.ReadFile(fieldSaveSlotPath(i))
			if err != nil {
				t.Fatal(err)
			}
		}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var saved saveState
		var secondFile []byte
		var samples []map[string]any
		for i := 193; i < 204; i++ {
			n := i + 1
			scan, err := strconv.ParseInt(src.Queued[i]["scan"], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			in := idle
			in.AnyKeyEdge = true
			switch scan {
			case 0x3f:
				in.SaveMenu = true
			case 0x40:
				in.LoadMenu = true
			case 0x1c:
				in.Enter = true
			case 0x50:
				in.DirHeld = 0
				in.DirEdge = 0
			case 0x4b:
				in.DirHeld = 2
				in.DirEdge = 2
			case 0x4d:
				in.DirHeld = 3
				in.DirEdge = 3
			default:
				t.Fatal("unsupported source scan")
			}
			x, err := strconv.Atoi(src.States[i]["player_x"])
			if err != nil {
				t.Fatal(err)
			}
			y, err := strconv.Atoi(src.States[i]["player_y"])
			if err != nil {
				t.Fatal(err)
			}
			if err := g.step(in); err != nil {
				t.Fatal(err)
			}
			if n == 199 || n >= 203 {
				for updates := 0; g.px != x || g.py != y; updates++ {
					if updates >= 60 {
						t.Fatal("normal held move did not reach source tile")
					}
					held := idle
					held.DirHeld = in.DirHeld
					if err := g.step(held); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := g.step(idle); err != nil {
				t.Fatal(err)
			}
			settleFieldSaveLoad(t, g)
			m := &g.fieldSaveLoad
			switch n {
			case 194:
				if !m.active || m.loading || m.stage != fsQuestion || m.cursor != 0 {
					t.Fatal("normal save question")
				}
			case 195, 196:
				if !m.active || m.loading || m.stage != fsSlots || m.cursor != n-195 {
					t.Fatal("normal save slot cursor", n)
				}
			case 197:
				if !m.active || m.stage != fsFinalWait {
					t.Fatal("normal save completion wait")
				}
			case 200, 201:
				if !m.active || !m.loading || m.stage != fsSlots || m.cursor != n-200 {
					t.Fatal("normal load slot cursor", n)
				}
			default:
				if m.active {
					t.Fatal("normal field return", n)
				}
			}
			current := g.snapshot()
			if n == 197 {
				secondFile, err = os.ReadFile(fieldSaveSlotPath(1))
				if err != nil {
					t.Fatal(err)
				}
				saved, err = decodeSave(secondFile)
				if err != nil || !equalFieldSave(saved, current) || bytes.Equal(secondFile, stored[1]) {
					t.Fatal("second slot was not replaced with current snapshot", err)
				}
			}
			want := before
			if n >= 197 {
				want = saved
			}
			want.PX, want.PY = x, y
			if n >= 202 {
				clock, ok := m.contract.LoadClock(saved.Layer)
				phaseTicks := g.dayNightCycle.ClockTicks / 4
				if !ok || phaseTicks <= 0 {
					t.Fatal("load clock rule missing")
				}
				want.DNPhase, want.DNStep = clock/phaseTicks, clock%phaseTicks
			}
			if !equalFieldSave(want, current) || g.prng != rng || fmt.Sprintf("%x", g.storyBits) != src.States[i]["flags"] {
				t.Fatalf("second slot persistent state differs at %d", n)
			}
			clock, err := strconv.Atoi(src.Clocks[n-193]["clock"])
			if err != nil || g.dayNightClock() != clock {
				t.Fatal("source world clock differs", n, err)
			}
			for j := range stored {
				wantFile := stored[j]
				if j == 1 && n >= 197 {
					wantFile = secondFile
				}
				actual, err := os.ReadFile(fieldSaveSlotPath(j))
				if err != nil || !bytes.Equal(actual, wantFile) {
					t.Fatal("second slot transaction changed unexpected storage", n, j, err)
				}
			}
			g.renderFrame()
			f, err := os.Create(filepath.Join(dest, fmt.Sprintf("packet-%03d.png", n)))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			closeErr := f.Close()
			if err != nil || closeErr != nil {
				t.Fatal(err, closeErr)
			}
			diff := sourceCanvasDifference(t, g, path, fmt.Sprintf("%s-packet-%03d-%s.png", prefix, n, src.States[i]["phase"]))
			if (n >= 194 && n <= 197 || n == 202 || n == 203) && diff != 0 {
				t.Fatalf("same-state whole RGB differs at %d: %d", n, diff)
			}
			var npcVisuals []map[string]int
			for _, npc := range g.cur.npcs {
				npcVisuals = append(npcVisuals, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
			}
			samples = append(samples, map[string]any{"packet": n, "active": m.active, "stage": m.stage, "cursor": m.cursor, "clock": g.dayNightClock(), "full_rgb_difference": diff, "hero_walk": g.walk, "hero_facing": g.facing, "animation_updates": g.anim, "movement_cooldown": g.cd, "npc_visuals": npcVisuals})
		}
		for g.cd > 0 {
			if err := g.step(idle); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Save(); err != nil {
			t.Fatal(err)
		}
		beforeLoad := g.snapshot()
		if err := g.Load(); err != nil || !equalFieldSave(beforeLoad, g.snapshot()) || g.prng != rng {
			t.Fatal("save/load after second slot differs", err)
		}
		if err := g.step(InputState{DirHeld: 2, DirEdge: 2}); err != nil {
			t.Fatal(err)
		}
		if g.px != 1 || g.py != 18 || g.fieldSaveLoad.active {
			t.Fatal("normal next field step")
		}
		report := map[string]any{"source_sha256": sourceHash, "normal_input_prefix": 193, "normal_final_packet": 204, "rng_unchanged": true, "selected_slot": 2, "other_slots_unchanged_during_route": true, "normal_save_load_and_next_step": true, "initial_slot_metadata": metadata, "second_slot_json_sha256": fmt.Sprintf("%x", sha256.Sum256(secondFile)), "samples": samples, "pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(), "limitation": "visible prior-slot metadata only; stored worlds not equivalent until this route writes slot2; original global RNG and complete animation timing not aligned"}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "receipt.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}, true)
}
