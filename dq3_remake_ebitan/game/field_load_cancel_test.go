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
	"reflect"
	"strconv"
	"testing"
)

func TestFieldLoadCancelDosgolemNormalInputComparison(t *testing.T) {
	dir, out := os.Getenv("DQ3_LOAD_CANCEL_ORACLE_DIR"), os.Getenv("DQ3_LOAD_CANCEL_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem F6 cancel source")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	const prefix = "issue4-load-cancel-r1"
	const sourceHash = "426c7623d239af6f83fa9715de861ca932a9fa6bbc2a223bb7eac2f49aa484a7"
	path := filepath.Join(dir, prefix+"-source-r2-receipt.json")
	raw, err := os.ReadFile(path)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("source identity differs", err)
	}
	var src struct{ Queued, States, Clocks []map[string]string }
	if err := json.Unmarshal(raw, &src); err != nil || len(src.Queued) != 197 || len(src.States) != 197 || len(src.Clocks) != 5 {
		t.Fatal("source shape", err)
	}
	dest := filepath.Join(out, "load-cancel")
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
		var samples []map[string]any
		for i := 193; i < 197; i++ {
			n := i + 1
			scan, err := strconv.ParseInt(src.Queued[i]["scan"], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			in := idle
			in.AnyKeyEdge = true
			switch scan {
			case 0x40:
				in.LoadMenu = true
			case 1:
				in.Cancel = true
			case 0x4b:
				in.DirEdge = 2
				in.DirHeld = 2
			case 0x4d:
				in.DirEdge = 3
				in.DirHeld = 3
			default:
				t.Fatal("unsupported normal scan")
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
			if n >= 196 {
				for updates := 0; g.px != x || g.py != y; updates++ {
					if updates >= 60 {
						t.Fatal("normal held movement did not reach source tile")
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
			if n == 194 && (!m.active || !m.loading || m.stage != fsSlots || m.cursor != 0 || len(m.slots) != len(metadata)) || n >= 195 && m.active {
				t.Fatalf("normal F6 cancel phase differs at %d", n)
			}
			want := before
			want.PX, want.PY = x, y
			if !equalFieldSave(want, g.snapshot()) || g.prng != rng || fmt.Sprintf("%x", g.storyBits) != src.States[i]["flags"] {
				t.Fatalf("normal F6 cancel persistent state differs at %d", n)
			}
			clock, err := strconv.Atoi(src.Clocks[n-193]["clock"])
			if err != nil || g.dayNightClock() != clock {
				t.Fatalf("normal F6 cancel clock differs at %d: %d vs %d", n, g.dayNightClock(), clock)
			}
			for j := range stored {
				current, err := os.ReadFile(fieldSaveSlotPath(j))
				if err != nil || !bytes.Equal(stored[j], current) {
					t.Fatal("F6 cancel changed external slot", j, err)
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
			samples = append(samples, map[string]any{"packet": n, "active": m.active, "stage": m.stage, "clock": g.dayNightClock(), "full_rgb_difference": diff, "hero_walk": g.walk, "hero_facing": g.facing, "animation_updates": g.anim, "movement_cooldown": g.cd})
		}
		for g.cd > 0 {
			if err := g.step(idle); err != nil {
				t.Fatal(err)
			}
		}
		if err := g.Save(); err != nil {
			t.Fatal(err)
		}
		// Saving establishes the current respawn checkpoint. Compare Load
		// with the successfully saved state, after that documented transaction.
		beforeLoad := g.snapshot()
		if err := g.Load(); err != nil || !equalFieldSave(beforeLoad, g.snapshot()) || g.prng != rng {
			want, got := reflect.ValueOf(beforeLoad), reflect.ValueOf(g.snapshot())
			for i := 0; i < want.NumField(); i++ {
				if !reflect.DeepEqual(want.Field(i).Interface(), got.Field(i).Interface()) {
					t.Logf("save/load field %s: before=%v after=%v", want.Type().Field(i).Name, want.Field(i).Interface(), got.Field(i).Interface())
				}
			}
			t.Fatalf("normal cancel save/load differs: err=%v rng_changed=%t", err, g.prng != rng)
		}
		if err := g.step(InputState{DirHeld: 2, DirEdge: 2}); err != nil {
			t.Fatal(err)
		}
		if g.px != 1 || g.py != 18 || g.fieldSaveLoad.active {
			t.Fatal("normal field step after cancel save/load")
		}
		report := map[string]any{"source_sha256": sourceHash, "normal_input_prefix": 193, "normal_final_packet": 197, "rng_unchanged": true, "cancel_storage_unchanged": true, "cancel_persistent_state_unchanged": true, "normal_save_load_and_next_step": true, "initial_slot_metadata": metadata, "samples": samples, "pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(), "limitation": "visible slot metadata only; other stored worlds are not equivalent; full animation timing oracle remains unknown"}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dest, "receipt.json"), append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}, true)
}
