package game

import (
	"bytes"
	"encoding/json"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFieldIdleStatusRejectsBadSources(t *testing.T) {
	for _, name := range []string{"missing_contract", "missing_assets", "bad_font", "bad_record", "bad_glyph"} {
		t.Run(name, func(t *testing.T) {
			p, err := gamepack.BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			assets := os.DirFS(spineAssetsDir(t))
			switch name {
			case "missing_contract":
				p.Interface.FieldIdleStatus = nil
			case "missing_assets":
				if validateFieldIdleSources(nil, p) == nil {
					t.Fatal("accepted missing assets")
				}
				return
			case "bad_font":
				p.Interface.FieldIdleStatus.FontAsset = "missing"
			case "bad_record":
				d, _ := p.TextDefinition(p.Interface.FieldIdleStatus.LeftTextID)
				r := 0
				d.Source.Record = &r
			case "bad_glyph":
				p.Interface.PartyHUD.ClassGlyphs[0][0] = 10000
			}
			if validateFieldIdleSources(assets, p) == nil {
				t.Fatal("accepted missing or mismatched source")
			}
		})
	}
}

func TestDosgolemFieldIdleNormalLifecycle(t *testing.T) {
	path := os.Getenv("DQ3_KING_IDLE_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_KING_IDLE_ORIGINAL to verified original receipt")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var original struct {
		Scenario string `json:"scenario"`
		Injected bool   `json:"game_state_injection"`
		Inputs   []struct {
			Scan string `json:"scan"`
		} `json:"player_input"`
		Windows []string `json:"idle_window_events"`
	}
	if err = json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.Scenario != "king_idle" || original.Injected || len(original.Inputs) != 49 || len(original.Windows) != 5 {
		t.Fatal("unverified source lifecycle")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	for _, in := range original.Inputs[38:47] {
		if in.Scan != "0x48" {
			t.Fatal("original direction route changed")
		}
		for g.cd > 0 {
			step(idle)
		}
		step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true})
		for g.cd > 0 {
			step(idle)
		}
	}
	if g.curCty != 25 || g.cur.sec != 0 || g.px != 15 || g.py != 30 {
		t.Fatal("normal castle checkpoint differs")
	}
	initial := g.snapshot()
	originalPrefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	type sample struct {
		Ordinal          int  `json:"ordinal"`
		FullDifference   int  `json:"full_rgb_difference"`
		WindowDifference int  `json:"window_rgb_difference"`
		Open             bool `json:"open"`
	}
	var samples []sample
	capture := func(ordinal int, sourceName string) {
		t.Helper()
		g.renderFrame()
		full := sourceCanvasDifference(t, g, path, sourceName)
		originalFile := filepath.Join(filepath.Dir(path), sourceName)
		f, err := os.Open(originalFile)
		if err != nil {
			t.Fatal(err)
		}
		img, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		windowDiff := 0
		for y := 238; y < 326; y++ {
			for x := 152; x < 272; x++ {
				r, b, c, _ := img.At(x, y).RGBA()
				i := (y*ScreenW + x) * 4
				if g.rgba[i] != uint8(r>>8) || g.rgba[i+1] != uint8(b>>8) || g.rgba[i+2] != uint8(c>>8) {
					windowDiff++
				}
			}
		}
		if g.fieldIdle.open && windowDiff != 0 {
			t.Fatalf("normal idle window differs at %d pixels; full=%d", windowDiff, full)
		}
		samples = append(samples, sample{ordinal, full, windowDiff, g.fieldIdle.open})
		if out := os.Getenv("DQ3_FIELD_IDLE_OUT"); out != "" {
			f, err := os.Create(out + "-" + sourceName)
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
	for cycle := 0; cycle < 3; cycle++ {
		for count := g.fieldIdle.elapsed; count < g.pack.Interface.FieldIdleStatus.HoldFrames()-1; count++ {
			step(idle)
			if g.fieldIdle.open {
				t.Fatal("idle window opened before threshold")
			}
		}
		step(idle)
		if !g.fieldIdle.open {
			t.Fatal("idle window did not open naturally")
		}
		ordinal := cycle*2 + 1
		capture(ordinal, originalPrefix+"-idle-waiting-"+[]string{"01", "03", "05"}[cycle]+".png")
		frozen := append([]byte(nil), g.rgba...)
		rng := g.prng.State()
		for i := 0; i < 5; i++ {
			step(idle)
		}
		if !bytes.Equal(frozen, g.rgba) || g.prng.State() != rng {
			t.Fatal("waiting modal mutated canvas or RNG")
		}
		if cycle == 2 {
			break
		}
		step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true})
		if g.fieldIdle.open || g.px != 15 || g.py != 30 {
			t.Fatal("dismissal also moved player")
		}
		capture(ordinal+1, originalPrefix+"-idle-restore-"+[]string{"02", "04"}[cycle]+".png")
		for i := 0; i < 3; i++ {
			step(InputState{DirHeld: 1, DirEdge: -1})
			if g.px != 15 || g.py != 30 {
				t.Fatal("held dismissal moved player")
			}
		}
		step(idle)
	}
	before, _ := json.Marshal(initial)
	after, _ := json.Marshal(g.snapshot())
	if !bytes.Equal(before, after) {
		t.Fatal("idle lifecycle changed saved game state")
	}
	if err = g.Save(); err != nil {
		t.Fatal(err)
	}
	if err = g.Load(); err != nil || g.fieldIdle.open || g.fieldIdle.consumeDirection || g.fieldIdle.elapsed != 0 || len(g.fieldIdle.background) != 0 {
		t.Fatal("same-instance load retained UI transient", err)
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
	if loaded.showTitle || loaded.fieldIdle.open || loaded.fieldIdle.consumeDirection || loaded.fieldIdle.elapsed != 0 || loaded.px != 15 || loaded.py != 30 || loaded.heroHP != g.heroHP || loaded.heroMP != g.heroMP {
		t.Fatal("normal title load lost state or retained modal")
	}
	if err = loaded.step(InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}); err != nil || loaded.py != 29 {
		t.Fatal("normal player cannot continue after load", err)
	}
	if out := os.Getenv("DQ3_FIELD_IDLE_OUT"); out != "" {
		report := map[string]any{"schema": g.pack.Schema(), "content_hash": g.pack.ContentHash(), "normal_inputs": 49, "state_injection": false, "cycles": 3, "dismissals": 2, "hold_frames": g.pack.Interface.FieldIdleStatus.HoldFrames(), "timing_accuracy": "hardware-spec approximation", "samples": samples, "title_save_load": true, "normal_continuation": true, "full_rgb_parity": false, "original_save_load_oracle": false, "audio_parity": false}
		data, _ := json.MarshalIndent(report, "", "  ")
		if err = os.WriteFile(out+".json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("normal idle cycles=3 dismissals=2 title_load=true full_differences=%+v", samples)
}

func TestFieldIdleHealthPaletteBranches(t *testing.T) {
	for _, tc := range []struct {
		name      string
		hp        int
		companion bool
		expected  int
	}{
		{"normal", 100, false, 0}, {"half", 50, false, 1}, {"quarter", 25, false, 2},
		{"sum_capped", 25, true, 3}, {"dead", 0, false, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := partyRenderFixture(t)
			g.heroStat[stats.HP] = 100
			g.heroHP = tc.hp
			g.companions = nil
			if tc.companion {
				g.companions = []*Member{{CurHP: 25, Stats: stats.Values{stats.HP: 100}}}
			}
			s := g.pack.Interface.FieldIdleStatus
			want := dq3data.DecodePalette(s.HealthPaletteRaw[tc.expected], 1)[0]
			if got := g.fieldIdleForeground(); got != want {
				t.Fatalf("health colour %v want %v", got, want)
			}
		})
	}
}
