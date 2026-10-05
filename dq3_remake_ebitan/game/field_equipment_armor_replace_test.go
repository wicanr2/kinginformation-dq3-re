package game

import (
	"bytes"
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

func TestFieldEquipmentArmorReplaceDosgolemNormalInputComparison(t *testing.T) {
	runFieldEquipmentNormalAt366(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-equip-armor-replace-normal-r1"
		const sourceHash = "d13ac42535c7f8b0bf4844ef8820f1f8d9e3b25b08095478fe4237ebdd43771f"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("source identity", e)
		}
		var src struct {
			Queued, States, Clocks []map[string]string
			Artifacts              []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 381 || len(src.Artifacts) != 953 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			path := filepath.Join(dir, a.Path)
			if filepath.Dir(a.Path) != "." {
				path = filepath.Join(filepath.Dir(dir), a.Path)
			}
			b, e := os.ReadFile(path)
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("artifact", a.Path, e)
			}
		}
		before, rng := g.snapshot(), g.prng
		stored := make([][]byte, g.fieldSaveLoad.contract.SlotCount)
		for i := range stored {
			stored[i], e = os.ReadFile(fieldSaveSlotPath(i))
			if e != nil {
				t.Fatal(e)
			}
		}
		scans := []string{"39", "50", "50", "39", "01", "39", "01", "01", "39", "50", "39", "39", "1c", "4b", "4d"}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		for i, scan := range scans {
			n := 367 + i
			st := src.States[n-1]
			if src.Queued[n-1]["scan"] != scan {
				t.Fatal("input", n)
			}
			in := idle
			in.AnyKeyEdge = true
			switch scan {
			case "39":
				in.Confirm = true
			case "50":
				in.DirEdge = 0
			case "01":
				in.Cancel = true
			case "1c":
				in.Enter = true
			case "4b":
				in.DirHeld, in.DirEdge = 2, 2
			case "4d":
				in.DirHeld, in.DirEdge = 3, 3
			default:
				t.Fatal("scan")
			}
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
			actor, e := hex.DecodeString(st["actor"])
			if e != nil || len(actor) != 128 {
				t.Fatal("actor", n, e)
			}
			words := make([]uint16, 8)
			for j := range words {
				words[j] = binary.LittleEndian.Uint16(actor[0x3a+j*2:])
			}
			if !reflect.DeepEqual(g.items.Words(), words) {
				t.Fatal("physical words", n, g.items.Words(), words)
			}
			want := before
			want.Items, _ = json.Marshal(struct {
				Version int      `json:"storage_version"`
				Words   []uint16 `json:"words"`
			}{1, words})
			want.PX, _ = strconv.Atoi(st["player_x"])
			want.PY, _ = strconv.Atoi(st["player_y"])
			if !equalFieldSave(want, g.snapshot()) || g.prng != rng {
				t.Fatal("persistent transaction or RNG", n)
			}
			found := false
			for _, s := range src.Clocks {
				if s["packet"] == strconv.Itoa(n) {
					found = true
					if s["clock"] != strconv.Itoa(g.dayNightClock()) {
						t.Fatal("source clock", n)
					}
				}
			}
			if !found {
				t.Fatal("missing clock", n)
			}
			if n >= 374 {
				_, _, atk, def, _ := g.heroStats()
				if atk != int(binary.LittleEndian.Uint16(actor[0x1c:])) || def != int(binary.LittleEndian.Uint16(actor[0x20:])) {
					t.Fatal("derived abilities", n)
				}
			}
			if g.fieldEquipment.active {
				count, _ := strconv.Atoi(st["choice_count"])
				cursor, _ := strconv.Atoi(st["choice_cursor"])
				if count != len(g.nativeEquipmentEntries())+1 || cursor != g.panelCursor+1 {
					t.Fatal("native choices", n)
				}
			}
			if n == 378 && (g.panel != panelStatusDetail || g.cmd.open) {
				t.Fatal("native detail wait")
			}
			if n == 374 || n == 379 || n == 381 {
				if g.fieldEquipment.active || g.cmd.open || g.panel != panelNone {
					t.Fatal("field return", n)
				}
			}
			for j, b := range stored {
				actual, e := os.ReadFile(fieldSaveSlotPath(j))
				if e != nil || !bytes.Equal(actual, b) {
					t.Fatal("equipment unexpectedly wrote slot", n, j, e)
				}
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("armor-replace-packet-%03d.png", n))
			f, e := os.Create(path)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, st["phase"]))
			t.Logf("normal armor replacement packet%d full640x350 RGB difference=%d", n, diff)
			if n <= 378 && diff != 0 {
				t.Errorf("verified full canvas differs packet%d: %d", n, diff)
			}
			b, _ := os.ReadFile(path)
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "physical_words": words, "clock": g.dayNightClock(), "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))})
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
			t.Fatal("save wait")
		}
		step(enter)
		b, e := os.ReadFile(savePath())
		if e != nil {
			t.Fatal(e)
		}
		saved, e := decodeSave(b)
		if e != nil || !equalFieldSave(expectedSave, saved) || !equalFieldSave(expectedSave, g.snapshot()) {
			t.Fatal("replacement save", e)
		}
		clock, ok := g.pack.Interface.FieldSaveLoad.LoadClock(saved.Layer)
		ticks := g.dayNightCycle.ClockTicks / 4
		if !ok || ticks <= 0 {
			t.Fatal("load clock")
		}
		expectedLoad := expectedSave
		expectedLoad.DNPhase, expectedLoad.DNStep = clock/ticks, clock%ticks
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true, AnyKeyEdge: true})
		step(enter)
		if g.fieldSaveLoad.active || g.fieldEquipment.active || !equalFieldSave(expectedLoad, g.snapshot()) || g.prng != rng {
			t.Fatal("replacement save/load")
		}
		for g.cd > 0 {
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		if e = g.step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true}); e != nil || g.px != 2 || g.py != 18 {
			t.Fatal("next step", e)
		}
		report := map[string]any{"scope": "normal new-game381 same-part armor physical5 to4 replacement, detail and return", "source_sha256": sourceHash, "samples": samples, "state_injection": false, "rng_unchanged": true, "normal_remake_save_load_roundtrip": true, "original_replacement_save_sample": false, "all_prior_json_slots_unchanged_during_replacement": true, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e = json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "armor-replace-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
