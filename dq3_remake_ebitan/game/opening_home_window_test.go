package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// The reviewed window is a bounded parity claim. Full RGB remains observable
// and RED; this helper neither masks it nor reports the whole frame as passed.
func verifyHomeWindow(t *testing.T, g *Game, label string) {
	t.Helper()
	dir := os.Getenv("DQ3_HOME_ORIGINAL_DIR")
	if dir == "" {
		return
	}
	suffix, ok := map[string]string{"preview": "continue-2-stable", "round-1": "opening-modal-01", "round-2": "opening-modal-02", "round-3": "opening-modal-03", "result": "opening-modal-04"}[label]
	if !ok {
		return
	}
	prefix := "issue4-home-contract"
	raw, err := os.ReadFile(filepath.Join(dir, prefix+"-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Scenario  string `json:"scenario"`
		Injected  bool   `json:"game_state_injection"`
		SHA256    string `json:"original_sha256"`
		Artifacts []struct {
			Path   string `json:"path"`
			Size   int    `json:"size"`
			SHA256 string `json:"sha256"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Scenario != "mother_home_contract" || receipt.Injected || receipt.SHA256 != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original home receipt identity invalid")
	}
	name := prefix + "-" + suffix + ".png"
	raw, err = os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Fatal(err)
	}
	verified := false
	for _, a := range receipt.Artifacts {
		if a.Path == name {
			if verified || a.Size != len(raw) || a.SHA256 != fmt.Sprintf("%x", sha256.Sum256(raw)) {
				t.Fatal("original image manifest mismatch")
			}
			verified = true
		}
	}
	if !verified {
		t.Fatal("original image not in manifest")
	}
	original, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if original.Bounds().Dx() != ScreenW || original.Bounds().Dy() != ScreenH {
		t.Fatal("original image dimensions invalid")
	}
	w := g.openingEscort.Home.Picture.Window
	full, window := 0, 0
	for y := 0; y < ScreenH; y++ {
		for x := 0; x < ScreenW; x++ {
			r, green, b, _ := original.At(x, y).RGBA()
			off := (y*ScreenW + x) * 4
			if uint8(r>>8) != g.rgba[off] || uint8(green>>8) != g.rgba[off+1] || uint8(b>>8) != g.rgba[off+2] {
				full++
				if x >= w.X && x < w.X+w.Width && y >= w.Y && y < w.Y+w.Height {
					window++
				}
			}
		}
	}
	t.Logf("%s: reviewed window RGB=%d; complete 640x350 RGB=%d; full-frame parity=%v", label, window, full, full == 0)
	if window != 0 {
		t.Fatalf("%s reviewed home selector window differs by %d pixels", label, window)
	}
}
