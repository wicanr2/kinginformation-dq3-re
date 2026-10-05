package game

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
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
		report := map[string]any{"scope": "normal command194..205; item ownership/window covered by the separate ordered-word trace", "source_sha256": sourceHash, "samples": samples, "snapshot_unchanged_before_save": true, "rng_unchanged": true, "pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(), "normal_save_load_and_next_step": true, "item_equipped_row_parity": false, "production_changed": true}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}, true)
}

// This source verifies single-owner cancellation. Ordered words and selected
// identity are covered by the separate 230-packet original source below.
func TestFieldItemSingleOwnerCancelDosgolemNormalInputComparison(t *testing.T) {
	dir, out := os.Getenv("DQ3_ITEM_CANCEL_ORACLE_DIR"), os.Getenv("DQ3_ITEM_CANCEL_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem single-owner item cancel oracle")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	const sourceHash = "28995c8dd702f83d70c51f3f09212dc556356ed563d5b0b9d3977ff4b0d60eb4"
	sourcePath := filepath.Join(dir, "issue4-field-item-navigation-r2-source-r1-receipt.json")
	raw, err := os.ReadFile(sourcePath)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("source identity differs", err)
	}
	var src struct {
		Queued, States []map[string]string
		Artifacts      []struct {
			Path   string
			Size   int
			SHA256 string
		}
	}
	if err = json.Unmarshal(raw, &src); err != nil || len(src.Queued) != 218 || len(src.States) != 218 || len(src.Artifacts) != 462 {
		t.Fatal("source shape differs", err)
	}
	for _, artifact := range src.Artifacts {
		if filepath.Base(artifact.Path) != artifact.Path {
			t.Fatal("source artifact path differs")
		}
		b, e := os.ReadFile(filepath.Join(dir, artifact.Path))
		if e != nil || len(b) != artifact.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != artifact.SHA256 {
			t.Fatal("source artifact differs", artifact.Path, e)
		}
	}
	dest := filepath.Join(out, "item-cancel")
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
		if len(g.companions) != 0 {
			t.Fatal("single-owner route differs")
		}
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
		for n := 194; n <= 218; n++ {
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
				if n == 218 {
					in.DirHeld = 3
				}
			default:
				t.Fatal("unsupported source input", n)
			}
			step(in)
			if n < 218 && (!equalFieldSave(before, g.snapshot()) || rng != g.prng || g.dayNightClock() != 30) {
				t.Fatal("item navigation/cancel changed persistence", n)
			}
			if (n == 206 || n == 212) && (g.panel != panelItem || g.itemActionStage != itemActionList) {
				t.Fatal("normal item open/reopen unavailable", n)
			}
			if n == 207 && (g.panel != panelItem || g.itemActionStage != itemActionMenu) {
				t.Fatal("normal item action unavailable")
			}
			if (n == 208 || n == 217) && (g.panel != panelNone || g.cmd.open) {
				t.Fatal("single-owner Esc did not return to field", n, g.panel, g.itemActionStage)
			}
			if n >= 209 && n <= 211 && (!g.cmd.open || g.panel != panelNone) {
				t.Fatal("normal command reopen unavailable", n)
			}
			if n == 218 {
				want := before
				want.PX++
				if !equalFieldSave(want, g.snapshot()) || rng != g.prng || g.px != 3 || g.py != 18 {
					t.Fatal("normal post-cancel next step changed more than position")
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
			diff := sourceCanvasDifference(t, g, sourcePath, fmt.Sprintf("issue4-field-item-navigation-r2-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			// Sprite timing after item cancellation is unresolved. Retain the
			// complete RGB diagnostic without forcing a selected phase.
			if n <= 205 && diff != 0 {
				t.Fatal("command full canvas differs", n, diff)
			}
			encoded, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			var npcs []map[string]int
			if g.cur != nil {
				for _, npc := range g.cur.npcs {
					npcs = append(npcs, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
				}
			}
			samples = append(samples, map[string]any{
				"packet": n, "full_rgb_difference": diff, "panel": g.panel, "action_stage": g.itemActionStage,
				"png_sha256": fmt.Sprintf("%x", sha256.Sum256(encoded)), "png_size": len(encoded),
				"hero_facing": g.facing, "hero_walk": g.walk, "npc_visuals": npcs,
			})
			t.Logf("normal item packet%d complete640x350 RGB difference=%d", n, diff)
		}
		// Formal F5/F6 validates the actual post-cancel game, not a restore
		// fixture. Save updates Respawn, so compare with the written snapshot.
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "item-cancel-save.json"))
		step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
		settleFieldSaveLoad(t, g)
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsQuestion {
			t.Fatal("post-cancel F5 unavailable")
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
			t.Fatal("post-cancel saved snapshot differs", e)
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsSlots {
			t.Fatal("post-cancel F6 unavailable")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		clock, ok := g.fieldSaveLoad.contract.LoadClock(saved.Layer)
		if !ok {
			t.Fatal("load clock contract missing")
		}
		phase := g.dayNightCycle.ClockTicks / 4
		saved.DNPhase, saved.DNStep = clock/phase, clock%phase
		if g.fieldSaveLoad.active || !equalFieldSave(saved, g.snapshot()) || rng != g.prng || g.cmd.open || g.panel != panelNone {
			t.Fatal("post-cancel normal save/load differs")
		}
		for updates := 0; g.cd > 0; updates++ {
			if updates >= moveCooldown {
				t.Fatal("post-load movement cooldown did not finish")
			}
			if e := g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		x, y := g.px, g.py
		step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true})
		if g.px != x-1 || g.py != y {
			t.Fatal("post-load normal move blocked")
		}
		report := map[string]any{
			"scope":         "single-owner item cancel208/217; ordered list/selection covered by the separate 230-packet source",
			"source_sha256": sourceHash, "samples": samples,
			"cancel_without_transaction": true, "rng_unchanged": true,
			"normal_reopen_save_load_and_next_step": true,
			"item_storage_parity":                   false, "selected_item_parity": false,
			"cancel_full_rgb_parity": false, "animation_timing_parity": false,
			"pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(),
		}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}, true)
}

func TestFieldItemOrderedWordsDosgolemNormalInputComparison(t *testing.T) {
	dir, out := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem single-owner item cancel oracle")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	const sourceHash = "aad971bb8cbf08956384b6964b4a893251cb3b4ff316263953f7563a7e5b2fc4"
	sourcePath := filepath.Join(dir, "issue4-field-item-reorder-r2-source-r3-receipt.json")
	raw, err := os.ReadFile(sourcePath)
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
		t.Fatal("source identity differs", err)
	}
	var src struct {
		Queued, States []map[string]string
		Artifacts      []struct {
			Path   string
			Size   int
			SHA256 string
		}
	}
	if err = json.Unmarshal(raw, &src); err != nil || len(src.Queued) != 230 || len(src.States) != 230 || len(src.Artifacts) != 498 {
		t.Fatal("source shape differs", err)
	}
	for _, artifact := range src.Artifacts {
		if filepath.Base(artifact.Path) != artifact.Path {
			t.Fatal("source artifact path differs")
		}
		b, e := os.ReadFile(filepath.Join(dir, artifact.Path))
		if e != nil || len(b) != artifact.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != artifact.SHA256 {
			t.Fatal("source artifact differs", artifact.Path, e)
		}
	}
	dest := filepath.Join(out, "item-ordered")
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
		if len(g.companions) != 0 {
			t.Fatal("single-owner route differs")
		}
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
		for n := 194; n <= 230; n++ {
			in := idle
			in.AnyKeyEdge = true
			scan, e := strconv.ParseInt(src.Queued[n-1]["scan"], 16, 64)
			if e != nil {
				t.Fatal(e)
			}
			switch scan {
			case 0x39:
				in.Confirm = true
			case 0x1c:
				in.Enter = true
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
				if n == 218 {
					in.DirHeld = 3
				}
			default:
				t.Fatal("unsupported source input", n)
			}
			step(in)
			if n < 218 && (!equalFieldSave(before, g.snapshot()) || rng != g.prng || g.dayNightClock() != 30) {
				t.Fatal("item navigation/cancel changed persistence", n)
			}
			if (n == 206 || n == 212) && (g.panel != panelItem || g.itemActionStage != itemActionList) {
				t.Fatal("normal item open/reopen unavailable", n)
			}
			if n == 207 && (g.panel != panelItem || g.itemActionStage != itemActionMenu) {
				t.Fatal("normal item action unavailable")
			}
			if (n == 208 || n == 217) && (g.panel != panelNone || g.cmd.open) {
				t.Fatal("single-owner Esc did not return to field", n, g.panel, g.itemActionStage)
			}
			if n >= 209 && n <= 211 && (!g.cmd.open || g.panel != panelNone) {
				t.Fatal("normal command reopen unavailable", n)
			}
			if n == 218 {
				want := before
				want.PX++
				if !equalFieldSave(want, g.snapshot()) || rng != g.prng || g.px != 3 || g.py != 18 {
					t.Fatal("normal post-cancel next step changed more than position")
				}
			}
			actorRaw, err := hex.DecodeString(src.States[n-1]["actor"])
			if err != nil || len(actorRaw) != 128 {
				t.Fatal("original actor shape")
			}
			expectedWords := make([]uint16, 8)
			for i := range expectedWords {
				expectedWords[i] = binary.LittleEndian.Uint16(actorRaw[0x3a+i*2:])
			}
			if !reflect.DeepEqual(g.items.Words(), expectedWords) {
				t.Fatalf("normal physical words packet%d: got %04x want %04x", n, g.items.Words(), expectedWords)
			}
			if (n == 207 || n == 223) && g.itemSelected != 0 {
				t.Fatal("first displayed worn physical slot was not selected")
			}
			if n == 225 && (!g.dlg.open || g.itemActionStage != itemActionGiveWait || g.items.Worn()[0].Position != 7) {
				t.Fatal("self gift did not enter original wait with wear at final physical slot")
			}
			if n == 226 && (g.dlg.open || g.panel != panelNone || g.cmd.open) {
				t.Fatal("Enter did not finish single-owner gift")
			}
			if n == 230 && (g.panel != panelItem || g.itemActionStage != itemActionList || len(g.actorItemEntries(0)) != 7 || g.actorItemEntries(0)[6].Position != 7) {
				t.Fatal("normal reopen lost hole or physical order")
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
			diff := sourceCanvasDifference(t, g, sourcePath, fmt.Sprintf("issue4-field-item-reorder-r2-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			// Sprite timing after item cancellation is unresolved. Retain the
			// complete RGB diagnostic without forcing a selected phase.
			if n <= 207 && diff != 0 {
				t.Fatal("command full canvas differs", n, diff)
			}
			encoded, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			var npcs []map[string]int
			if g.cur != nil {
				for _, npc := range g.cur.npcs {
					npcs = append(npcs, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
				}
			}
			samples = append(samples, map[string]any{
				"packet": n, "full_rgb_difference": diff, "panel": g.panel, "action_stage": g.itemActionStage,
				"png_sha256": fmt.Sprintf("%x", sha256.Sum256(encoded)), "png_size": len(encoded),
				"hero_facing": g.facing, "hero_walk": g.walk, "npc_visuals": npcs,
			})
			t.Logf("normal item packet%d complete640x350 RGB difference=%d", n, diff)
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
		// Formal F5/F6 validates the actual post-cancel game, not a restore
		// fixture. Save updates Respawn, so compare with the written snapshot.
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "item-ordered-save.json"))
		step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
		settleFieldSaveLoad(t, g)
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsQuestion {
			t.Fatal("post-cancel F5 unavailable")
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
			t.Fatal("post-cancel saved snapshot differs", e)
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
		if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsSlots {
			t.Fatal("post-cancel F6 unavailable")
		}
		step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
		clock, ok := g.fieldSaveLoad.contract.LoadClock(saved.Layer)
		if !ok {
			t.Fatal("load clock contract missing")
		}
		phase := g.dayNightCycle.ClockTicks / 4
		saved.DNPhase, saved.DNStep = clock/phase, clock%phase
		if g.fieldSaveLoad.active || !equalFieldSave(saved, g.snapshot()) || rng != g.prng || g.cmd.open || g.panel != panelNone {
			t.Fatal("post-cancel normal save/load differs")
		}
		for updates := 0; g.cd > 0; updates++ {
			if updates >= moveCooldown {
				t.Fatal("post-load movement cooldown did not finish")
			}
			if e := g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		x, y := g.px, g.py
		step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true})
		if g.px != x-1 || g.py != y {
			t.Fatal("post-load normal move blocked")
		}
		report := map[string]any{
			"scope":         "ordered original physical words, native206/207 UI, normal self-gift225, reopen230 and F5/F6",
			"source_sha256": sourceHash, "samples": samples,
			"cancel_without_transaction": true, "rng_unchanged": true,
			"normal_reopen_save_load_and_next_step": true,
			"item_storage_parity":                   true, "selected_item_parity": true, "ui206_207_full_rgb_parity": true,
			"original_tool": "dosgolem", "original_revision": "2f44a68ebfc54b28fb15dd4a34510b0b04a5415d",
			"original_exe_sha256": "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c",
			"original_seed":       "0x1357", "remake_seed": "0x1357", "seed_configuration": "each side fixed once before normal character creation",
			"picture_seed_original": "0x151b", "picture_seed_remake": "0x151b", "picture_seed_method": "original natural BIOS value; remake controlled seed installed before home selection",
			"initial_checkpoint":   map[string]any{"packet": 193, "cty": before.Cty, "section": before.Section, "x": before.PX, "y": before.PY, "words": before.itemStore.Words()},
			"game_state_injection": false, "pack_content_version": g.pack.ContentVersion(),
			"original_words_compared_at_each_packet": true, "source_actor_words_offset": "0x3a",
			"save_version":           saveFormatVersion,
			"cancel_full_rgb_parity": false, "animation_timing_parity": false,
			"pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(),
		}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}, true)
}
