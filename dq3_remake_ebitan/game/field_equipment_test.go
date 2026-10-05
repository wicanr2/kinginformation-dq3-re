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

func TestFieldEquipmentDosgolemNormalInputComparison(t *testing.T) {
	runFieldSpellEmptyNormalAt304(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-equip-wear-normal-r1"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "721d93f6de2123f03ad82ce991185f7329fa918898c1b1cfbf25d2870f3cf414" {
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
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 339 || len(src.Artifacts) != 825 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(filepath.Join(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("artifact", a.Path, e)
			}
		}
		scans := []string{"39", "50", "50", "39", "01", "01", "01", "01", "4b", "4d", "39", "50", "50", "39", "50", "39", "50", "50", "39", "01", "01", "4b", "4d", "39", "50", "50", "39", "48", "39", "48", "39", "01", "01", "4b", "4d"}
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		for i, scan := range scans {
			n := 305 + i
			st := src.States[n-1]
			if src.Queued[n-1]["scan"] != scan {
				t.Fatal("input", n)
			}
			in := InputState{DirHeld: -1, DirEdge: -1, AnyKeyEdge: true}
			switch scan {
			case "39":
				in.Confirm = true
			case "01":
				in.Cancel = true
			case "50":
				in.DirEdge = 0
			case "48":
				in.DirEdge = 1
			case "4b":
				in.DirHeld, in.DirEdge = 2, 2
			case "4d":
				in.DirHeld, in.DirEdge = 3, 3
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
				t.Fatalf("physical words packet%d got%#v want%#v", n, g.items.Words(), words)
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
			if n == 325 || n == 337 {
				_, _, atk, def, _ := g.heroStats()
				if atk != int(binary.LittleEndian.Uint16(actor[0x1c:])) || def != int(binary.LittleEndian.Uint16(actor[0x20:])) {
					t.Fatal("returned derived abilities", n, atk, def)
				}
			}
			if n == 312 || n == 325 || n == 337 {
				if g.fieldEquipment.active || g.cmd.open || g.panel != panelNone {
					t.Fatal("native return", n)
				}
			}
			if g.fieldEquipment.active {
				count, _ := strconv.Atoi(st["choice_count"])
				cursor, _ := strconv.Atoi(st["choice_cursor"])
				if count != len(g.nativeEquipmentEntries())+1 || cursor != g.panelCursor+1 {
					t.Fatal("native choices", n)
				}
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("equipment-packet-%03d.png", n))
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
			t.Logf("normal equipment packet%d full640x350 RGB difference=%d", n, diff)
			b, _ := os.ReadFile(p)
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "physical_words": words})
			// These complete canvases are verified V3. Later NPC animation
			// differences remain fully recorded and are not called parity.
			mustMatch := n <= 312 || n >= 315 && n <= 325 || n == 338
			if mustMatch && diff != 0 {
				t.Errorf("full canvas differs packet%d: %d", n, diff)
			}
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
			t.Fatal("save transaction", e)
		}
		clock, ok := g.pack.Interface.FieldSaveLoad.LoadClock(saved.Layer)
		phaseTicks := g.dayNightCycle.ClockTicks / 4
		if !ok || phaseTicks <= 0 {
			t.Fatal("load clock")
		}
		expectedLoad := expectedSave
		expectedLoad.DNPhase, expectedLoad.DNStep = clock/phaseTicks, clock%phaseTicks
		step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true, AnyKeyEdge: true})
		step(enter)
		if g.fieldSaveLoad.active || g.fieldEquipment.active || !equalFieldSave(expectedLoad, g.snapshot()) || g.prng != rng {
			t.Fatal("save/load roundtrip")
		}
		for g.cd > 0 {
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
		if e = g.step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true}); e != nil || g.px != 2 {
			t.Fatal("move after load", e)
		}
		report := map[string]any{"scope": "normal new-game through339 equipment wear/unwear and four-slot cancel", "source_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "samples": samples, "state_injection": false, "rng_unchanged": true, "normal_save_load_roundtrip": true, "move_after_load": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}
		b, e = json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "equipment-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}

func TestFieldEquipmentScopeAndRestore(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	before := g.snapshot()
	g.selectCommand(cmdEquip)
	if !g.fieldEquipment.active || g.panelActor != 0 {
		t.Fatal("single-owner entry")
	}
	g.restore(before)
	if g.fieldEquipment.active {
		t.Fatal("transient survived load")
	}
	g.heroConditions = conditionPoison
	g.selectCommand(cmdEquip)
	if g.fieldEquipment.active || g.panelActor != -1 {
		t.Fatal("unreviewed scope changed")
	}
}
