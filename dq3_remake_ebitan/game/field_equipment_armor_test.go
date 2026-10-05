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

func TestFieldEquipmentArmorDosgolemNormalInputComparison(t *testing.T) {
	runFieldEquipmentNormalAt366(t, nil)
}

func runFieldEquipmentNormalAt366(t *testing.T, after func(*Game)) {
	t.Helper()
	if root := os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR"); root != "" {
		dest := filepath.Join(root, "armor")
		if err := os.Mkdir(dest, 0755); err != nil {
			t.Fatal(err)
		}
		t.Setenv("DQ3_ITEM_ORDERED_RECEIPT_DIR", dest)
	}
	runFieldEquipmentNormalAt339(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-equip-armor-normal-r1"
		const sourceHash = "377d6d3088cd1d1a71eb789ef67579b05656428bbd92bf4130cfdfeb2ea457d9"
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
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 366 || len(src.Artifacts) != 908 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			path := filepath.Join(dir, a.Path)
			if filepath.Dir(a.Path) != "." {
				path = filepath.Join(filepath.Dir(dir), a.Path)
			}
			b, e := os.ReadFile(path)
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "field-save.json"))
		before, rng := g.snapshot(), g.prng
		expectedRespawn := respawnToSave(g.currentRespawnPoint())
		metadata := prepareFieldSaveSlotMetadata(t, g, before)
		if !equalFieldSave(before, g.snapshot()) || g.prng != rng {
			t.Fatal("external fixture changed live game")
		}
		stored := make([][]byte, len(metadata))
		for i := range stored {
			stored[i], e = os.ReadFile(fieldSaveSlotPath(i))
			if e != nil {
				t.Fatal(e)
			}
		}
		scans := []string{"39", "50", "50", "39", "01", "50", "39", "01", "01", "39", "50", "39", "39", "1c", "4b", "4d", "3f", "1c", "50", "1c", "1c", "4b", "40", "50", "1c", "4b", "4d"}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		var savedFile []byte
		for i, scan := range scans {
			n := 340 + i
			st := src.States[n-1]
			if src.Queued[n-1]["scan"] != scan {
				t.Fatal("input", n)
			}
			in := idle
			in.AnyKeyEdge = true
			switch scan {
			case "39":
				in.Confirm = true
			case "1c":
				in.Enter = true
			case "01":
				in.Cancel = true
			case "50":
				in.DirEdge = 0
			case "4b":
				in.DirHeld, in.DirEdge = 2, 2
			case "4d":
				in.DirHeld, in.DirEdge = 3, 3
			case "3f":
				in.SaveMenu = true
			case "40":
				in.LoadMenu = true
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
			settleFieldSaveLoad(t, g)
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
			if n >= 359 {
				want.Respawn = expectedRespawn
			}
			if n >= 364 {
				clock, ok := g.pack.Interface.FieldSaveLoad.LoadClock(want.Layer)
				ticks := g.dayNightCycle.ClockTicks / 4
				if !ok || ticks <= 0 {
					t.Fatal("load clock")
				}
				want.DNPhase, want.DNStep = clock/ticks, clock%ticks
			}
			if !equalFieldSave(want, g.snapshot()) || g.prng != rng {
				t.Fatal("persistent transaction or RNG", n)
			}
			clocks := src.Clocks
			found := false
			for _, s := range clocks {
				if s["packet"] == strconv.Itoa(n) {
					found = true
					if s["clock"] != strconv.Itoa(g.dayNightClock()) {
						t.Fatal("clock", n)
					}
				}
			}
			if !found {
				t.Fatal("missing clock", n)
			}
			if n >= 348 {
				_, _, atk, def, _ := g.heroStats()
				if atk != 8 || def != 8 {
					t.Fatal("derived abilities", n, atk, def)
				}
			}
			if g.fieldEquipment.active {
				count, _ := strconv.Atoi(st["choice_count"])
				cursor, _ := strconv.Atoi(st["choice_cursor"])
				if count != len(g.nativeEquipmentEntries())+1 || cursor != g.panelCursor+1 {
					t.Fatal("equipment choices", n)
				}
			}
			m := &g.fieldSaveLoad
			switch n {
			case 356:
				if !m.active || m.loading || m.stage != fsQuestion || m.cursor != 0 {
					t.Fatal("save question")
				}
			case 357, 358:
				if !m.active || m.loading || m.stage != fsSlots || m.cursor != n-357 {
					t.Fatal("save cursor")
				}
			case 359:
				if !m.active || m.stage != fsFinalWait {
					t.Fatal("save wait")
				}
			case 362, 363:
				if !m.active || !m.loading || m.stage != fsSlots || m.cursor != n-362 {
					t.Fatal("load cursor")
				}
			default:
				if m.active {
					t.Fatal("unexpected save modal", n)
				}
			}
			if n == 352 && (g.panel != panelStatusDetail || g.cmd.open) {
				t.Fatal("native detail wait")
			}
			if n == 348 || n == 353 || n == 364 || n == 366 {
				if g.fieldEquipment.active || g.cmd.open || g.panel != panelNone {
					t.Fatal("field return", n)
				}
			}
			if n == 359 {
				savedFile, e = os.ReadFile(fieldSaveSlotPath(1))
				if e != nil {
					t.Fatal(e)
				}
				saved, e := decodeSave(savedFile)
				if e != nil || !equalFieldSave(want, saved) || bytes.Equal(savedFile, stored[1]) {
					t.Fatal("worn second slot save", e)
				}
			}
			for j := range stored {
				wantFile := stored[j]
				if j == 1 && n >= 359 {
					wantFile = savedFile
				}
				actual, e := os.ReadFile(fieldSaveSlotPath(j))
				if e != nil || !bytes.Equal(actual, wantFile) {
					t.Fatal("unexpected slot write", n, j, e)
				}
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("armor-packet-%03d.png", n))
			f, e := os.Create(p)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, st["phase"]))
			t.Logf("normal armor packet%d full640x350 RGB difference=%d", n, diff)
			if (n == 357 || n == 358 || n == 362 || n == 363 || n == 366) && diff != 0 {
				t.Errorf("verified full canvas differs packet%d: %d", n, diff)
			}
			b, _ := os.ReadFile(p)
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "physical_words": words, "clock": g.dayNightClock()})
		}
		if after != nil {
			after(g)
			return
		}

		for g.cd > 0 {
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		if e = g.step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true}); e != nil || g.px != 2 || g.py != 18 {
			t.Fatal("next step", e)
		}
		report := map[string]any{"scope": "normal cold new-game366 second physical armor, detail, worn F5/F6 slot2 and next step", "source_sha256": sourceHash, "samples": samples, "state_injection": false, "rng_unchanged": true, "normal_worn_save_load": true, "other_slots_unchanged": true, "move_after_load": true, "initial_slot_metadata": metadata, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "armor-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
