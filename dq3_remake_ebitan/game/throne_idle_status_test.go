package game

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDosgolemThroneIdleNormalLifecycle(t *testing.T) {
	g, path := runDosgolemThroneLanding(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Windows []string `json:"idle_window_events"`
	}
	if err = json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if len(original.Windows) != 8 || !strings.Contains(original.Windows[6], "phase=waiting ordinal=7 ") || !strings.Contains(original.Windows[7], "phase=restore ordinal=8 ") {
		t.Fatal("original floor idle lifecycle missing")
	}
	for _, event := range original.Windows[6:] {
		if !strings.Contains(event, "player_x=9 player_y=22 party_count=1") {
			t.Fatal("original floor idle position differs")
		}
	}
	idle := InputState{DirHeld: -1, DirEdge: -1}
	up := InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	before, _ := json.Marshal(g.snapshot())
	hold := g.pack.Interface.FieldIdleStatus.HoldFrames()
	for g.fieldIdle.elapsed < hold-1 {
		step(idle)
		if g.fieldIdle.open {
			t.Fatal("floor idle opened before threshold")
		}
	}
	step(idle)
	if !g.fieldIdle.open || g.px != 9 || g.py != 22 {
		t.Fatal("floor idle did not open naturally at landing")
	}
	prefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	type sample struct {
		Phase            string `json:"phase"`
		FullDifference   int    `json:"full_rgb_difference"`
		WindowDifference int    `json:"window_rgb_difference"`
	}
	var samples []sample
	capture := func(phase, sourceName string) {
		t.Helper()
		g.renderFrame()
		full := sourceCanvasDifference(t, g, path, prefix+sourceName)
		window := -1
		if g.fieldIdle.open {
			f, err := os.Open(filepath.Join(filepath.Dir(path), prefix+sourceName))
			if err != nil {
				t.Fatal(err)
			}
			img, err := png.Decode(f)
			f.Close()
			if err != nil {
				t.Fatal(err)
			}
			window = 0
			for y := 238; y < 326; y++ {
				for x := 152; x < 272; x++ {
					r, b, c, _ := img.At(x, y).RGBA()
					i := (y*ScreenW + x) * 4
					if g.rgba[i] != uint8(r>>8) || g.rgba[i+1] != uint8(b>>8) || g.rgba[i+2] != uint8(c>>8) {
						window++
					}
				}
			}
			if window != 0 {
				t.Fatalf("floor idle window differs: window=%d full=%d", window, full)
			}
		}
		samples = append(samples, sample{phase, full, window})
		if out := os.Getenv("DQ3_THRONE_IDLE_OUT"); out != "" {
			f, err := os.Create(out + "-" + phase + ".png")
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			closeErr := f.Close()
			if err != nil || closeErr != nil {
				t.Fatal(err, closeErr)
			}
		}
	}
	capture("waiting", "-idle-waiting-07.png")
	frozen := append([]byte(nil), g.rgba...)
	rng := g.prng.State()
	for i := 0; i < 5; i++ {
		step(idle)
	}
	if !bytes.Equal(frozen, g.rgba) || rng != g.prng.State() {
		t.Fatal("floor idle mutated canvas or RNG")
	}
	waitingSnapshot, _ := json.Marshal(g.snapshot())
	if !bytes.Equal(before, waitingSnapshot) {
		t.Fatal("natural floor idle changed saved progress")
	}
	if err = g.Save(); err != nil {
		t.Fatal(err)
	}
	// Save intentionally updates the respawn checkpoint. Compare UI actions
	// with the saved snapshot, then require title Load to restore it in full.
	savedSnapshot, _ := json.Marshal(g.snapshot())
	step(up)
	if g.fieldIdle.open || g.px != 9 || g.py != 22 {
		t.Fatal("dismissal also moved player")
	}
	capture("restore", "-idle-restore-08.png")
	for i := 0; i < 3; i++ {
		step(InputState{DirHeld: 1, DirEdge: -1})
		if g.py != 22 {
			t.Fatal("held dismissal moved player")
		}
	}
	step(idle)
	after, _ := json.Marshal(g.snapshot())
	if !bytes.Equal(savedSnapshot, after) {
		var a, b map[string]json.RawMessage
		json.Unmarshal(savedSnapshot, &a)
		json.Unmarshal(after, &b)
		for key, value := range a {
			if !bytes.Equal(value, b[key]) {
				t.Logf("snapshot field %s before=%s after=%s", key, value, b[key])
			}
		}
		t.Fatal("floor idle changed saved progress")
	}
	for i := 0; i < hold && !g.fieldIdle.open; i++ {
		step(idle)
	}
	if !g.fieldIdle.open {
		t.Fatal("floor idle did not reopen")
	}
	if err = g.Load(); err != nil || g.fieldIdle.open || g.fieldIdle.consumeDirection || g.fieldIdle.elapsed != 0 || len(g.fieldIdle.background) != 0 {
		t.Fatal("load retained floor UI transient", err)
	}
	loaded, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []InputState{{DirHeld: -1, DirEdge: -1, Enter: true}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: -1, Enter: true}} {
		if err = loaded.step(in); err != nil {
			t.Fatal(err)
		}
	}
	if loaded.showTitle || loaded.curCty != 25 || loaded.cur.sec != 1 || loaded.px != 9 || loaded.py != 22 || loaded.fieldIdle.open || loaded.fieldIdle.consumeDirection || loaded.fieldIdle.elapsed != 0 || len(loaded.fieldIdle.background) != 0 {
		t.Fatal("normal title load lost progress or retained modal")
	}
	loadedSnapshot, _ := json.Marshal(loaded.snapshot())
	if !bytes.Equal(savedSnapshot, loadedSnapshot) {
		t.Fatal("normal title load changed floor progress")
	}
	if err = loaded.step(up); err != nil || loaded.px != 9 || loaded.py != 21 {
		t.Fatal("normal continuation failed", err)
	}
	if out := os.Getenv("DQ3_THRONE_IDLE_OUT"); out != "" {
		report := map[string]any{"schema": g.pack.Schema(), "content_version": g.pack.Manifest.ContentVersion, "content_hash": g.pack.ContentHash(), "normal_inputs_to_landing": 70, "state_injection": false, "hold_frames": hold, "timing_accuracy": "hardware-spec approximation", "samples": samples, "snapshot_unchanged": true, "modal_frozen": true, "same_instance_load_clears_transient": true, "title_save_load": true, "normal_continuation": true, "full_rgb_parity": false, "king_audience_completed": false, "original_save_load_oracle": false, "audio_parity": false}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(out+".json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("floor idle hold=%d samples=%+v; normal title load and continuation pass", hold, samples)
}
