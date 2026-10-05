package game

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// assertArrivalCanvas verifies the full source canvas at the normally reached
// final arrival state. It does not control animation or substitute a save load.
func assertArrivalCanvas(t *testing.T, g *Game, receiptPath string) {
	t.Helper()
	name := strings.TrimSuffix(filepath.Base(receiptPath), "-receipt.json") + "-mother-entry-1020a.png"
	difference := sourceCanvasDifference(t, g, receiptPath, name)
	if difference != 0 {
		t.Fatalf("normal arrival full RGB differs at %d pixels", difference)
	}
	t.Log("normal arrival full 640x350 RGB difference=0; no crop, mask or frame override")
}

func sourceCanvasDifference(t *testing.T, g *Game, receiptPath, name string) int {
	t.Helper()
	raw, err := os.ReadFile(receiptPath)
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		OriginalHash string `json:"original_sha256"`
		Meta         struct {
			OriginalHash string `json:"original_sha256"`
		} `json:"meta"`
		Artifacts []struct {
			Path   string `json:"path"`
			Size   int    `json:"size"`
			SHA256 string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err = json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.OriginalHash == "" {
		receipt.OriginalHash = receipt.Meta.OriginalHash
	}
	if receipt.OriginalHash != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" ||
		(receipt.Meta.OriginalHash != "" && receipt.Meta.OriginalHash != receipt.OriginalHash) {
		t.Fatal("original executable identity differs")
	}
	content, err := os.ReadFile(filepath.Join(filepath.Dir(receiptPath), name))
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.Sum256(content)
	verified := false
	for _, a := range receipt.Artifacts {
		if a.Path == name {
			verified = a.Size == len(content) && a.SHA256 == hex.EncodeToString(hash[:])
		}
	}
	if !verified {
		t.Fatal("final PNG identity differs from source manifest")
	}
	f, err := os.Open(filepath.Join(filepath.Dir(receiptPath), name))
	if err != nil {
		t.Fatal(err)
	}
	original, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if original.Bounds() != image.Rect(0, 0, ScreenW, ScreenH) {
		t.Fatal("original canvas dimensions differ")
	}
	difference := 0
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			r, gg, b, _ := original.At(x, y).RGBA()
			o := (y*ScreenW + x) * 4
			if uint8(r>>8) != g.rgba[o] || uint8(gg>>8) != g.rgba[o+1] || uint8(b>>8) != g.rgba[o+2] {
				difference++
			}
		}
	}
	return difference
}
