package gamepack

import (
	"bytes"
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
	if s == nil || len(p.Interface.SceneTileLayers) != 2 || s.Mode != "player_cell_layer" ||
		s.BaseTile != int(town.BaseLayerTile) || s.OtherTile != int(town.OtherLayerTile) ||
		s.BaseLayer != int(exe[0x11e0c-0xec90+4]) {
		t.Fatal("layer declaration differs from original header/consumer")
	}
	if exe[0x11e09-0xec90+2] != 0xc0 || dq3data.TownTileLayer(town.HiMap[23*town.W+24]) != 2 ||
		dq3data.TownTileLayer(town.HiMap[30*town.W+15]) != s.BaseLayer {
		t.Fatal("typed layer differs from original mask or normal entry")
	}
	townRaw := read("CTY00.DAT", "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836")
	field, err := dq3data.OpenTown(townRaw, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	fieldLayers := p.SceneTileLayers(0, 0)
	if fieldLayers == nil || fieldLayers.Mode != s.Mode || fieldLayers.BaseLayer != s.BaseLayer || fieldLayers.BaseTile != int(field.BaseLayerTile) || fieldLayers.OtherTile != int(field.OtherLayerTile) {
		t.Fatal("field layers differ from original header/consumer")
	}
	if dq3data.TownTileLayer(field.HiMap[22*field.W+5]) == fieldLayers.BaseLayer {
		t.Fatal("original normal tavern path did not sample non-base layer")
	}
	// Native layer branches precede the NPC flag test and sprite adapter. These
	// original bytes explain both hidden exterior actors while inside and hidden
	// interior actors after the normal room-door step.
	for _, original := range []struct {
		start int
		data  []byte
	}{
		{0x11e09, []byte{0x80, 0xe7, 0xc0}},
		{0x11e13, []byte{0x3a, 0x3e, 0x79, 0x25}},
		{0x11e19, []byte{0x8a, 0x1e, 0x57, 0x0b}},
		{0x11e20, []byte{0xf6, 0xc7, 0xc0}},
		{0x11e25, []byte{0x8a, 0x1e, 0x56, 0x0b}},
		{0x11e2c, []byte{0xf6, 0xc4, 0x20}},
		{0x11e33, []byte{0xe8, 0x71, 0x00}},
		{0x11ecc, []byte{0xe8, 0x01, 0x00}},
	} {
		off := original.start - 0xec90
		if !bytes.Equal(exe[off:off+len(original.data)], original.data) {
			t.Fatalf("native NPC layer consumer differs at IDA linear%X", original.start)
		}
	}
	if dq3data.TownTileLayer(field.HiMap[22*field.W+4]) != 2 || dq3data.TownTileLayer(field.HiMap[23*field.W+4]) != fieldLayers.BaseLayer {
		t.Fatal("normal421/422 player layer transition differs from source")
	}
	for _, record := range []int{14, 15} {
		npc := field.NPCs[record]
		if dq3data.TownTileLayer(field.HiMap[npc.Y*field.W+npc.X]) != 2 {
			t.Fatal("native hidden NPC cell layer differs", record)
		}
	}
	t.Logf("normal tavern layer=%d", dq3data.TownTileLayer(field.HiMap[22*field.W+5]))
	if p.SceneTileLayers(25, 1) != nil || p.SceneTileLayers(0, 2) != nil {
		t.Fatal("layer declaration leaked into another scene")
	}
	p.Interface.SceneTileLayers = []SceneTileLayers{}
	if err = p.validateSceneTileLayers(); err != nil {
		t.Fatal("explicit empty declaration rejected", err)
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}
