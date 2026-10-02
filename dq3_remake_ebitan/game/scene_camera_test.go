package game

import (
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestSceneCameraSourcesRejectUnknownReferences(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if err := validateSceneCameraSources(nil, pack); err == nil {
		t.Fatal("accepted missing scene camera assets")
	}
	for _, name := range []string{"missing_scene", "missing_section", "wrong_exterior"} {
		t.Run(name, func(t *testing.T) {
			pack, err := gamepack.BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "missing_scene":
				pack.Interface.SceneCameras[0].CTY = 10000
			case "missing_section":
				pack.Interface.SceneCameras[0].Section = 10000
			case "wrong_exterior":
				pack.Interface.SceneCameras[0].Camera.ExteriorTile = 255
			}
			if err = validateSceneCameraSources(os.DirFS(spineAssetsDir(t)), pack); err == nil {
				t.Fatal("accepted unknown or mismatched scene camera source")
			}
		})
	}
}

func TestDosgolemCastleCameraNormalInput(t *testing.T) {
	path := os.Getenv("DQ3_KING_APPROACH_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_KING_APPROACH_ORIGINAL to the verified normal original receipt")
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
		Ready []string `json:"king_ready_events"`
	}
	if err = json.Unmarshal(raw, &original); err != nil {
		t.Fatal(err)
	}
	if original.Scenario != "king_approach" || original.Injected || len(original.Inputs) != 47 || len(original.Ready) != 1 || !strings.Contains(original.Ready[0], "player_x=15 player_y=30") {
		t.Fatal("unverified original normal castle state")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	for _, in := range original.Inputs[38:] {
		if in.Scan != "0x48" {
			t.Fatal("original input route changed")
		}
		for g.cd > 0 {
			step(idle)
		}
		step(InputState{DirHeld: 1, DirEdge: -1})
		for g.cd > 0 {
			step(idle)
		}
	}
	if g.curCty != 25 || g.cur.sec != 0 || g.px != 15 || g.py != 30 {
		t.Fatalf("normal castle route differs: %d/%d (%d,%d)", g.curCty, g.cur.sec, g.px, g.py)
	}
	c := g.activeSceneCamera()
	if c == nil || c != g.pack.SceneCamera(g.curCty, g.cur.sec) || g.px-c.AnchorX != 6 || g.py-c.AnchorY != 23 {
		t.Fatal("declared original castle viewport missing")
	}
	g.renderFrame()
	prefix := strings.TrimSuffix(filepath.Base(path), "-receipt.json")
	timedDiff := sourceCanvasDifference(t, g, path, prefix+"-castle-north-09.png")
	waitDiff := sourceCanvasDifference(t, g, path, prefix+"-king-ready-09.png")
	t.Logf("normal castle camera applied; full RGB diagnostic timed=%d wait=%d, full parity remains false", timedDiff, waitDiff)
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
	if loaded.showTitle || loaded.curCty != g.curCty || loaded.cur.sec != g.cur.sec || loaded.px != g.px || loaded.py != g.py || lc == nil || *lc != *c || loaded.storyFlag(0x17) != g.storyFlag(0x17) || loaded.storyFlag(0x18) != g.storyFlag(0x18) {
		t.Fatal("normal title load lost castle camera or state")
	}
	if out := os.Getenv("DQ3_KING_CAMERA_OUT"); out != "" {
		f, err := os.Create(out + ".png")
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
		closeErr := f.Close()
		if err != nil || closeErr != nil {
			t.Fatalf("PNG: %v %v", err, closeErr)
		}
		report := map[string]any{"original_receipt": path, "normal_inputs": 47, "state_injection": false,
			"schema": g.pack.Schema(), "content_hash": g.pack.ContentHash(), "camera": c, "title_save_load": true,
			"timed_full_rgb_difference": timedDiff, "wait_full_rgb_difference": waitDiff, "full_rgb_parity": false, "king_audience_parity": false, "audio_parity": false}
		data, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(out+".json", append(data, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
