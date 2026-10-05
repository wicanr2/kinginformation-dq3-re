package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestFieldStatusDetailFreshKeyAndBranchScope(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	g.heroHP = 15
	if !g.nativeStatusDetailEligible() {
		t.Fatal("healthy unlearned branch rejected")
	}
	g.heroConditions = 1
	if g.nativeStatusDetailEligible() {
		t.Fatal("condition branch claimed as native")
	}
	g.heroConditions = 0
	g.heroHP = 0
	if g.nativeStatusDetailEligible() {
		t.Fatal("dead branch claimed as native")
	}
	g.heroHP = 15
	for _, in := range []InputState{
		{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true},
		{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true},
		{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true},
	} {
		g.panel, g.cmd.open = panelStatusDetail, false
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		if e := g.step(idle); e != nil {
			t.Fatal(e)
		}
		if g.panel != panelStatusDetail {
			t.Fatal("idle closed waiting page")
		}
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		if g.panel != panelNone || g.cmd.open || g.fieldSaveLoad.active || !equalFieldSave(before, g.snapshot()) || rng != g.prng {
			t.Fatal("close key leaked into field action")
		}
	}
}

func TestFieldStatusDetailDosgolemNormalInputComparison(t *testing.T) {
	runFieldStatusMenuNormalAt273(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const sourceHash = "edfad33432fb8b0dc6ee5bd59536e826af9b27e4c3b91ad4eb861365983469ff"
		const prefix = "issue4-field-detail-normal-r5"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("normal detail source identity", e)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 280 || len(src.Queued) != 280 || len(src.Artifacts) != 648 {
			t.Fatal("detail source shape", e)
		}
		for _, a := range src.Artifacts {
			b, err := os.ReadFile(filepath.Join(dir, a.Path))
			if filepath.Base(a.Path) != a.Path || err != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("detail artifact identity", a.Path, err)
			}
		}
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		for n := 274; n <= 280; n++ {
			in := idle
			in.AnyKeyEdge = true
			scan, err := strconv.ParseInt(src.Queued[n-1]["scan"], 16, 64)
			if err != nil {
				t.Fatal(err)
			}
			switch scan {
			case 0x39:
				in.Confirm = true
			case 0x50:
				in.DirEdge = 0
			case 0x1c:
				in.Enter = true
			case 0x4b:
				in.DirEdge, in.DirHeld = 2, 2
			case 0x4d:
				in.DirEdge, in.DirHeld = 3, 3
			default:
				t.Fatal("unsupported detail input", n)
			}
			if in.DirHeld >= 0 {
				for updates := 0; g.cd > 0; updates++ {
					if updates >= moveCooldown {
						t.Fatal("detail next-step cooldown")
					}
					if e = g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
			}
			if e = g.step(in); e != nil {
				t.Fatal(e)
			}
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
			want := before
			if n == 279 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("detail persistent state differs", n)
			}
			if n == 277 && (g.panel != panelStatusDetail || g.cmd.open) {
				t.Fatal("detail first page did not wait")
			}
			if n >= 278 && (g.panel != panelNone || g.cmd.open || g.dlg.open) {
				t.Fatal("fresh Enter did not return field", n)
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("detail-packet-%03d.png", n))
			f, err := os.Create(path)
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if err != nil || ce != nil {
				t.Fatal(err, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			if n == 277 && diff != 0 {
				t.Fatal("native detail full640x350 differs", diff)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "words": g.items.Words(), "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))})
			t.Logf("normal detail packet%d complete640x350 RGB difference=%d", n, diff)
		}
		// The enclosing normal route performs F5/F6 and the next field move after
		// this callback. Keep its save comparison aligned with the accepted state.
		b, e := json.MarshalIndent(map[string]any{"scope": "normal new-game through280 sole healthy unlearned hero native detail, fresh Enter and next moves", "source_sha256": sourceHash, "samples": samples, "game_state_injection": false, "rng_unchanged": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "detail-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
