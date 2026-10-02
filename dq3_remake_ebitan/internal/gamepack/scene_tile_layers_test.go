package gamepack

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func TestSceneTileLayersRejectIncompleteBindings(t *testing.T) {
	for _, name := range []string{"missing_collection", "missing_scene", "missing_base", "missing_tile", "null_tile", "unknown_field", "unknown_mode", "negative_tile", "layer_range", "duplicate_scene", "missing_camera", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing_collection" {
				p.Interface.SceneTileLayers = nil
			} else {
				raw, err := json.Marshal(p.Interface.SceneTileLayers[0])
				if err != nil {
					t.Fatal(err)
				}
				var changed map[string]any
				if err = json.Unmarshal(raw, &changed); err != nil {
					t.Fatal(err)
				}
				switch name {
				case "missing_scene":
					delete(changed, "section")
				case "missing_base":
					delete(changed, "base_layer")
				case "missing_tile":
					delete(changed, "other_tile")
				case "null_tile":
					changed["base_tile"] = nil
				case "unknown_field":
					changed["roof"] = true
				case "unknown_mode":
					changed["mode"] = "guessed"
				case "negative_tile":
					changed["base_tile"] = -1
				case "layer_range":
					changed["base_layer"] = 4
				case "unreviewed":
					changed["evidence"].(map[string]any)["level"] = "D1"
				}
				raw, err = json.Marshal(changed)
				if err != nil {
					t.Fatal(err)
				}
				var binding SceneTileLayers
				if err = json.Unmarshal(raw, &binding); err != nil {
					return
				}
				p.Interface.SceneTileLayers = []SceneTileLayers{binding}
				if name == "duplicate_scene" {
					p.Interface.SceneTileLayers = append(p.Interface.SceneTileLayers, binding)
				}
				if name == "missing_camera" {
					p.Interface.SceneCameras = []SceneCameraBinding{}
				}
			}
			if err := p.validateSceneTileLayers(); err == nil {
				t.Fatal("accepted incomplete scene tile layers")
			}
		})
	}
}

func TestSceneTileLayersMatchOriginalHeaderAndConsumer(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	read := func(name, hash string) []byte {
		t.Helper()
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != hash {
			t.Fatalf("original identity: %s %v", name, err)
		}
		return raw
	}
	raw := read("CTY25.DAT", "11d5c60377c6a98bbb9cfc9532652c5e23f4e5c939e769397e8fa5b2f22f9e2b")
	exe := read("DQ3.EXE", "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c")
	town, err := dq3data.OpenTown(raw, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	s := p.SceneTileLayers(25, 0)
	if s == nil || len(p.Interface.SceneTileLayers) != 1 || s.Mode != "player_cell_layer" ||
		s.BaseTile != int(town.BaseLayerTile) || s.OtherTile != int(town.OtherLayerTile) ||
		s.BaseLayer != int(exe[0x11e0c-0xec90+4]) {
		t.Fatal("layer declaration differs from original header/consumer")
	}
	if exe[0x11e09-0xec90+2] != 0xc0 || dq3data.TownTileLayer(town.HiMap[23*town.W+24]) != 2 ||
		dq3data.TownTileLayer(town.HiMap[30*town.W+15]) != s.BaseLayer {
		t.Fatal("typed layer differs from original mask or normal entry")
	}
	if p.SceneTileLayers(25, 1) != nil || p.SceneTileLayers(0, 0) != nil {
		t.Fatal("layer declaration leaked into another scene")
	}
	p.Interface.SceneTileLayers = []SceneTileLayers{}
	if err = p.validateSceneTileLayers(); err != nil {
		t.Fatal("explicit empty declaration rejected", err)
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}
