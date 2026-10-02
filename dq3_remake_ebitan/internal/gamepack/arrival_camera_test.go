package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestArrivalCameraRejectsIncompleteContract(t *testing.T) {
	for _, name := range []string{"missing", "null", "missing_zero_tile", "null_anchor", "unknown_field", "unknown_mode", "bad_x", "bad_y", "bad_tile", "unreviewed"} {
		t.Run(name, func(t *testing.T) {
			p, err := BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			raw, err := json.Marshal(p.Interface.OpeningEscort)
			if err != nil {
				t.Fatal(err)
			}
			var data map[string]any
			if err = json.Unmarshal(raw, &data); err != nil {
				t.Fatal(err)
			}
			camera := data["arrival_camera"].(map[string]any)
			switch name {
			case "missing":
				delete(data, "arrival_camera")
			case "null":
				data["arrival_camera"] = nil
			case "missing_zero_tile":
				delete(camera, "exterior_tile")
			case "null_anchor":
				camera["anchor_x"] = nil
			case "unknown_field":
				camera["guessed"] = true
			case "unknown_mode":
				camera["mode"] = "unknown"
			case "bad_x":
				camera["anchor_x"] = 20
			case "bad_y":
				camera["anchor_y"] = -1
			case "bad_tile":
				camera["exterior_tile"] = 256
			case "unreviewed":
				camera["evidence"].(map[string]any)["level"] = "D2"
			}
			raw, err = json.Marshal(data)
			if err != nil {
				t.Fatal(err)
			}
			var changed OpeningEscort
			if err = json.Unmarshal(raw, &changed); err != nil {
				return
			}
			p.Interface.OpeningEscort = &changed
			if err = p.validateInterface(); err == nil {
				t.Fatal("accepted incomplete arrival camera")
			}
		})
	}
}

func TestArrivalCameraMatchesOriginal(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../../assets_raw"
	}
	exe, err := os.ReadFile(filepath.Join(dir, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	cty, err := os.ReadFile(filepath.Join(dir, "CTY00.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if len(exe) != 115282 || fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" ||
		len(cty) != 7546 || fmt.Sprintf("%x", sha256.Sum256(cty)) != "ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836" {
		t.Fatal("original input identity differs")
	}
	p, err := BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	e := p.Interface.OpeningEscort
	c := e.ArrivalCamera
	if e.Destination.CTY != 0 || e.Destination.Section != 0 || c.Mode != "player_anchor" {
		t.Fatal("source scene differs")
	}
	start := 0x11971 - 0xec90
	if !bytes.Equal(exe[start:start+3], []byte{0xa1, 0x33, 0x4f}) ||
		!bytes.Equal(exe[start+3:start+4], []byte{0x2d}) ||
		!bytes.Equal(exe[start+13:start+15], []byte{0x83, 0xeb}) {
		t.Fatal("original anchor writer differs")
	}
	if c.AnchorX != int(binary.LittleEndian.Uint16(exe[start+4:start+6])) || c.AnchorY != int(exe[start+15]) {
		t.Fatal("anchor differs from original")
	}
	base := int(binary.LittleEndian.Uint16(cty[e.Destination.Section*2:]))
	if c.ExteriorTile != int(cty[base+0x12]) {
		t.Fatal("exterior tile differs from CTY")
	}
	t.Logf("schema=%s content=%s hash=%s", p.Schema(), p.Manifest.ContentVersion, p.ContentHash())
}
