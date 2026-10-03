package gamepack

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestSceneCamerasRejectIncompleteBindings(t *testing.T) {
	for _, name := range []string{"missing_collection", "missing_scene", "null_camera", "missing_tile", "unknown_field", "negative_scene", "duplicate_scene", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			if name == "missing_collection" {
				p.Interface.SceneCameras = nil
			} else {
				raw, err := json.Marshal(p.Interface.SceneCameras[0])
				if err != nil {
					t.Fatal(err)
				}
				var changed map[string]any
				if err = json.Unmarshal(raw, &changed); err != nil {
					t.Fatal(err)
				}
				camera := changed["camera"].(map[string]any)
				switch name {
				case "missing_scene":
					delete(changed, "section")
				case "null_camera":
					changed["camera"] = nil
				case "missing_tile":
					delete(camera, "exterior_tile")
				case "unknown_field":
					changed["guessed"] = true
				case "negative_scene":
					changed["cty"] = -1
				case "unreviewed":
					camera["evidence"].(map[string]any)["level"] = "D2"
				}
				raw, err = json.Marshal(changed)
				if err != nil {
					t.Fatal(err)
				}
				var binding SceneCameraBinding
				if err = json.Unmarshal(raw, &binding); err != nil {
					return
				}
				p.Interface.SceneCameras = []SceneCameraBinding{binding}
				if name == "duplicate_scene" {
					p.Interface.SceneCameras = append(p.Interface.SceneCameras, binding)
				}
			}
			if err := p.validateInterface(); err == nil {
				t.Fatal("accepted incomplete scene camera")
			}
		})
	}
}

func TestSceneCameraMatchesOriginalCastleHeader(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	raw, err := os.ReadFile(filepath.Join(dir, "CTY25.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 3756 || fmt.Sprintf("%x", sha256.Sum256(raw)) != "11d5c60377c6a98bbb9cfc9532652c5e23f4e5c939e769397e8fa5b2f22f9e2b" {
		t.Fatal("original castle identity differs")
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Interface.SceneCameras) != 4 {
		t.Fatal("unexpected reviewed scene count")
	}
	c := p.SceneCamera(25, 0)
	if c == nil || c.Mode != "player_anchor" || c.Evidence.Level != "D3" {
		t.Fatal("reviewed castle camera missing")
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original executable differs")
	}
	start := 0x11971 - 0xec90
	if c.AnchorX != int(binary.LittleEndian.Uint16(exe[start+4:])) || c.AnchorY != int(exe[start+15]) || c.ExteriorTile != int(raw[int(binary.LittleEndian.Uint16(raw))+0x12]) {
		t.Fatal("scene camera differs from original writer/header")
	}
	throne := p.SceneCamera(25, 1)
	if throne == nil || throne.Mode != "player_anchor" || throne.Evidence.Level != "D3" ||
		throne.AnchorX != c.AnchorX || throne.AnchorY != c.AnchorY ||
		throne.ExteriorTile != int(raw[0x08c7+0x12]) {
		t.Fatal("throne camera differs from original writer/header")
	}
	townRaw, err := os.ReadFile(filepath.Join(dir, "CTY00.DAT"))
	if err != nil || len(townRaw) != 7546 || fmt.Sprintf("%x", sha256.Sum256(townRaw)) != "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836" {
		t.Fatal("original town identity differs", err)
	}
	for _, section := range []int{0, 2} {
		camera := p.SceneCamera(0, section)
		base := int(binary.LittleEndian.Uint16(townRaw[section*2:]))
		if camera == nil || camera.Mode != "player_anchor" || camera.Evidence.Level != "D3" || camera.AnchorX != c.AnchorX || camera.AnchorY != c.AnchorY || camera.ExteriorTile != int(townRaw[base+0x12]) {
			t.Fatal("return camera differs from original writer/header", section)
		}
	}
	arrival := p.Interface.OpeningEscort.ArrivalCamera
	town := p.SceneCamera(0, 0)
	if arrival.Mode != town.Mode || arrival.AnchorX != town.AnchorX || arrival.AnchorY != town.AnchorY || arrival.ExteriorTile != town.ExteriorTile {
		t.Fatal("explicit town camera differs from existing arrival camera")
	}
	if p.SceneCamera(25, 2) != nil || p.SceneCamera(0, 1) != nil {
		t.Fatal("camera leaked into undeclared scenes")
	}
	p.Interface.SceneCameras = []SceneCameraBinding{}
	if err = p.validateSceneCameras(); err != nil {
		t.Fatal("explicit empty collection rejected", err)
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}
