package game

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestOpeningHomeAssetShapeFailsClosed(t *testing.T) {
	assets := os.DirFS(spineAssetsDir(t))
	g, err := newProductionTraceGame(assets)
	if err != nil {
		t.Fatal(err)
	}
	s := g.openingEscort.Home.Picture
	imageRef, _ := g.pack.Asset(s.AssetKey)
	image, err := os.ReadFile(spineAssetsDir(t) + "/" + imageRef.Path)
	if err != nil {
		t.Fatal(err)
	}
	palRef, _ := g.pack.Asset(s.PaletteAssetKey)
	pal, err := os.ReadFile(spineAssetsDir(t) + "/" + palRef.Path)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"header", "body", "count", "dimensions", "palette"} {
		t.Run(name, func(t *testing.T) {
			raw := append([]byte(nil), image...)
			colors := pal
			switch name {
			case "header":
				raw = raw[:5]
			case "body":
				raw = raw[:len(raw)-1]
			case "count":
				raw[4] = 255
				raw[5] = 255
			case "dimensions":
				raw[0] = 3
			case "palette":
				colors = nil
			}
			if err := g.loadHomePicture(func(key string) []byte {
				if key == s.AssetKey {
					return raw
				}
				if key == s.PaletteAssetKey {
					return colors
				}
				t.Fatal("unexpected logical asset")
				return nil
			}); err == nil {
				t.Fatal("invalid archive accepted")
			}
		})
	}
}

func TestOpeningHomeInvalidSaveIsAtomic(t *testing.T) {
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "home-save.json"))
	g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	initial := g.snapshot()
	initial.OpeningHomeAwait = true
	initial.InTown = true
	initial.Cty = g.openingEscort.CTY
	initial.Section = g.openingEscort.Section
	initial.PX = 5
	initial.PY = 5
	before := g.snapshot()
	for _, name := range []string{"scene", "gate", "metadata", "bounds"} {
		t.Run(name, func(t *testing.T) {
			s := initial
			s.StoryBits = append([]byte(nil), initial.StoryBits...)
			switch name {
			case "scene":
				s.Section++
			case "gate":
				flag := g.openingEscort.Home.ApproachFlag
				s.StoryBits[flag/8] &^= 128 >> uint(flag%8)
			case "metadata":
				s.PackID = ""
			case "bounds":
				s.PX = 99999
			}
			raw, err := encodeSave(s)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(savePath(), raw, 0600); err != nil {
				t.Fatal(err)
			}
			if err := g.Load(); err == nil {
				t.Fatal("invalid home checkpoint accepted")
			}
			a, _ := encodeSave(before)
			b, _ := encodeSave(g.snapshot())
			if !bytes.Equal(a, b) {
				t.Fatal("failed load consumed current state")
			}
		})
	}
	prior, err := os.ReadFile(savePath())
	if err != nil {
		t.Fatal(err)
	}
	respawn := g.respawn
	for _, in := range []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 0}, {Confirm: true, DirHeld: -1, DirEdge: -1}} {
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	if !g.showTitle || g.newGame.stage != ngMenu || g.newGame.cursor != ngOptLoad {
		t.Fatal("failed title load left menu or entered gameplay")
	}
	g.homeSelection.active = true
	if err := g.Save(); err == nil {
		t.Fatal("selector transaction saved mid-modal")
	}
	after, err := os.ReadFile(savePath())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(prior, after) || g.respawn != respawn {
		t.Fatal("failed save overwrote book or checkpoint")
	}
}
