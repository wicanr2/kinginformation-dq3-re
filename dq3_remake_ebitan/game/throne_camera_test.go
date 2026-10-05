package game

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runDosgolemThroneLanding(t *testing.T) (*Game, string) {
	t.Helper()
	path := os.Getenv("DQ3_KING_AUDIENCE_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_KING_AUDIENCE_ORIGINAL to verified original approach receipt")
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
		Events []string `json:"king_audience_events"`
	}
	if err = json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.Scenario != "king_audience" || original.Injected || len(original.Inputs) != 90 {
		t.Fatal("unverified source approach")
	}
	landing := false
	for _, event := range original.Events {
		if strings.Contains(event, "phase=ready input=21 ") && strings.Contains(event, "player_x=9 player_y=22 raw0b24=08c7 ") {
			landing = true
		}
	}
	if !landing {
		t.Fatal("original normal landing missing")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	up := InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}
	press := func() {
		t.Helper()
		for g.cd > 0 {
			step(idle)
		}
		step(up)
		step(idle)
		for g.cd > 0 {
			step(idle)
		}
	}
	for _, input := range original.Inputs[38:47] {
		if input.Scan != "0x48" {
			t.Fatal("castle direction changed")
		}
		press()
	}
	if g.curCty != 25 || g.cur.sec != 0 || g.px != 15 || g.py != 30 {
		t.Fatal("normal castle checkpoint differs")
	}
	for cycle := 0; cycle < 3; cycle++ {
		for i := 0; i < 300 && !g.fieldIdle.open; i++ {
			step(idle)
		}
		if !g.fieldIdle.open {
			t.Fatal("natural idle window missing")
		}
		press()
	}
	for _, input := range original.Inputs[50:70] {
		if input.Scan != "0x48" {
			t.Fatal("throne route direction changed")
		}
		press()
	}
	if g.curCty != 25 || g.cur.sec != 1 || g.px != 9 || g.py != 22 || !g.storyFlag(0x17) || g.storyFlag(0x18) || g.heroGold != 0 || len(testInventory(g.items)) != 0 || g.dlg.open {
		t.Fatal("normal throne landing or pre-audience state differs")
	}
	return g, path
}

func TestDosgolemThroneCameraNormalInput(t *testing.T) {
	g, path := runDosgolemThroneLanding(t)
	var err error
	up := InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}
	c := g.activeSceneCamera()
	if c == nil || c != g.pack.SceneCamera(25, 1) || g.px-c.AnchorX != 0 || g.py-c.AnchorY != 15 || c.ExteriorTile != 27 {
		t.Fatal("reviewed throne camera missing")
	}
	g.renderFrame()
	prefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	difference := sourceCanvasDifference(t, g, path, prefix+"-audience-north-20.png")
	if out := os.Getenv("DQ3_THRONE_CAMERA_OUT"); out != "" {
		f, e := os.Create(out + ".png")
		if e != nil {
			t.Fatal(e)
		}
		e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
		ce := f.Close()
		if e != nil || ce != nil {
			t.Fatal(e, ce)
		}
	}
	if err = g.Save(); err != nil {
		t.Fatal(err)
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
	lc := loaded.activeSceneCamera()
	if loaded.showTitle || loaded.curCty != g.curCty || loaded.cur.sec != g.cur.sec || loaded.px != g.px || loaded.py != g.py || lc == nil || *lc != *c || loaded.storyFlag(0x17) != g.storyFlag(0x17) || loaded.storyFlag(0x18) != g.storyFlag(0x18) || loaded.heroGold != g.heroGold || len(testInventory(loaded.items)) != len(testInventory(g.items)) {
		t.Fatal("normal title load lost throne camera or state")
	}
	if err = loaded.step(up); err != nil || loaded.px != 9 || loaded.py != 21 {
		t.Fatal("normal continuation after load failed", err)
	}
	if out := os.Getenv("DQ3_THRONE_CAMERA_OUT"); out != "" {
		report := map[string]any{"source_receipt": path, "schema": g.pack.Schema(), "content_hash": g.pack.ContentHash(), "normal_inputs_to_landing": 70, "state_injection": false, "camera": c, "x": g.px, "y": g.py, "full_rgb_difference": difference, "full_rgb_parity": difference == 0, "title_save_load_camera_state": true, "normal_continuation": true, "original_save_load_oracle": false, "king_audience_completed": false, "audio_parity": false, "input_timing": "logical press after cooldown; no CPU-step timing equivalence claimed"}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(out+".json", append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("normal throne camera and title load pass; complete 640x350 RGB difference=%d; king audience and audio remain unverified", difference)
}
