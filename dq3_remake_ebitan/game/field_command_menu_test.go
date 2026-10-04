package game

import (
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

func TestFieldCommandConfirmKeys(t *testing.T) {
	for _, enter := range []bool{false, true} {
		t.Run(fmt.Sprint("enter-", enter), func(t *testing.T) {
			g := fieldSaveLoadComponentGame(t)
			before, rng := g.snapshot(), g.prng
			for _, in := range []InputState{
				{DirHeld: -1, DirEdge: -1, Confirm: true},
				{DirHeld: -1, DirEdge: 0},
				{DirHeld: -1, DirEdge: 3},
				{DirHeld: -1, DirEdge: -1, Enter: enter, Confirm: !enter},
			} {
				if err := g.step(in); err != nil {
					t.Fatal(err)
				}
			}
			if g.cmd.open || g.panel != panelItem || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
				t.Fatal("confirmation key changed dispatch or persistence")
			}
		})
	}
}

// Optional raw source is an independently accepted cold run, never a restored
// executable snapshot. Missing assets/source is fatal when this route is enabled.
func TestFieldCommandDosgolemNormalInputComparison(t *testing.T) {
	dir, out := os.Getenv("DQ3_COMMAND_ORACLE_DIR"), os.Getenv("DQ3_COMMAND_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem normal command oracle")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	const sourceHash = "5b2a78c152e77f1ddbfd850af0cc129837bf206cd9a15731529927d21b12978a"
	sourcePath := filepath.Join(dir, "issue4-command-navigation-r1-source-r1-receipt.json")
	raw, err := os.ReadFile(sourcePath)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("source identity differs", err)
	}
	var src struct{ Queued, States []map[string]string }
	if err = json.Unmarshal(raw, &src); err != nil || len(src.Queued) != 206 || len(src.States) != 206 {
		t.Fatal("source shape", err)
	}
	dest := filepath.Join(out, "command-menu")
	if err = os.Mkdir(dest, 0755); err != nil {
		t.Fatal(err)
	}
	prefixOut := filepath.Join(dest, "prefix")
	if err = os.Mkdir(prefixOut, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_ORACLE_DIR", dir)
	t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_RECEIPT_DIR", prefixOut)
	runRecruitmentEmptyNormalInputComparisonAtCheckpoint(t, "menu-cancel", "issue4-recruit-menu-esc-normal-r1-source-r1-receipt.json", "137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c", "issue4-recruit-menu-esc-normal-r1", 193, 189, 190, func(g *Game) {
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		step := func(in InputState) {
			t.Helper()
			if e := g.step(in); e != nil {
				t.Fatal(e)
			}
			if e := g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		var samples []map[string]any
		for n := 194; n <= 206; n++ {
			in := idle
			in.AnyKeyEdge = true
			scan, e := strconv.ParseInt(src.Queued[n-1]["scan"], 16, 64)
			if e != nil {
				t.Fatal(e)
			}
			switch scan {
			case 0x39:
				in.Confirm = true
			case 0x01:
				in.Cancel = true
			case 0x50:
				in.DirEdge = 0
			case 0x48:
				in.DirEdge = 1
			case 0x4b:
				in.DirEdge = 2
			case 0x4d:
				in.DirEdge = 3
			default:
				t.Fatal("unsupported native command input")
			}
			step(in)
			if !equalFieldSave(before, g.snapshot()) || rng != g.prng || g.dayNightClock() != 30 {
				t.Fatal("normal menu changed persistence/RNG/clock", n)
			}
			cursor := -1
			if n == 202 {
				if g.cmd.open || g.panel != panelNone {
					t.Fatal("Esc did not return to field")
				}
			} else if n == 206 {
				if g.cmd.open || g.panel != panelItem {
					t.Fatal("native Item command not dispatched")
				}
			} else {
				want, _ := strconv.Atoi(src.States[n-1]["choice_cursor"])
				cursor = g.cmd.nativeCursor() + 1
				if !g.cmd.open || cursor != want {
					t.Fatal("normal native cursor differs", n, cursor, want)
				}
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("packet-%03d.png", n))
			f, e := os.Create(p)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, sourcePath, fmt.Sprintf("issue4-command-navigation-r1-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			if n <= 205 && diff != 0 {
				t.Fatal("normal command full canvas differs", n, diff)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "native_cursor": cursor, "hero_facing": g.facing, "hero_walk": g.walk, "animation_updates": g.anim})
			t.Logf("normal packet%d complete640x350 RGB difference=%d", n, diff)
		}
		// Continue from the actual Item entry through Cancel, F5, F6, then a
		// legal adjacent move. These additional keys are remake validation;
		// no new original save or post-item receipt is claimed.
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "command-save.json"))
		step(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
		if g.panel != panelNone {
			t.Fatal("item cancel did not leave panel")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
		settleFieldSaveLoad(t, g)
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsQuestion {
			t.Fatal("normal F5 unavailable")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		settleFieldSaveLoad(t, g)
		if g.fieldSaveLoad.stage != fsSlots {
			t.Fatal("normal save slots unavailable")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		settleFieldSaveLoad(t, g)
		if g.fieldSaveLoad.stage != fsFinalWait {
			t.Fatal("normal save completion unavailable")
		}
		written, e := os.ReadFile(savePath())
		if e != nil {
			t.Fatal(e)
		}
		saved, e := decodeSave(written)
		if e != nil || !equalFieldSave(saved, g.snapshot()) {
			t.Fatal("normal saved snapshot differs", e)
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsSlots {
			t.Fatal("normal F6 unavailable")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		clock, ok := g.fieldSaveLoad.contract.LoadClock(saved.Layer)
		if !ok {
			t.Fatal("load clock contract missing")
		}
		phase := g.dayNightCycle.ClockTicks / 4
		saved.DNPhase, saved.DNStep = clock/phase, clock%phase
		if g.fieldSaveLoad.active || !equalFieldSave(saved, g.snapshot()) || rng != g.prng || g.cmd.open {
			t.Fatal("menu/save/load round trip differs")
		}
		x, y := g.px, g.py
		step(InputState{DirHeld: 3, DirEdge: 3, AnyKeyEdge: true})
		if g.px != x+1 || g.py != y {
			t.Fatal("post-load normal move blocked")
		}
		report := map[string]any{"scope": "normal command194..205; Item206 retains explicit DRAFT blocker", "source_sha256": sourceHash, "samples": samples, "snapshot_unchanged_before_save": true, "rng_unchanged": true, "pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(), "normal_save_load_and_next_step": true, "item_equipped_row_parity": false, "production_changed": true}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}, true)
}
