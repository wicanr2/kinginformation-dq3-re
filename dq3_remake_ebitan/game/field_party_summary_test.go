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

func TestFieldPartySummaryFreshKeyConsumesAction(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	for _, in := range []InputState{
		{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true},
		{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true},
		{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true},
	} {
		g.panel, g.panelCursor, g.cmd.open = panelStatusMenu, 1, false
		before, rng := g.snapshot(), g.prng
		g.stepStatusMenu(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}, -1)
		if g.panel != panelPartySummary {
			t.Fatal("summary entry did not open")
		}
		if e := g.step(InputState{DirHeld: -1, DirEdge: -1}); e != nil {
			t.Fatal(e)
		}
		if g.panel != panelPartySummary {
			t.Fatal("idle closed summary")
		}
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		if g.panel != panelNone || g.cmd.open || g.fieldSaveLoad.active || !equalFieldSave(before, g.snapshot()) || g.prng != rng {
			t.Fatal("summary dismissal leaked action or transacted state")
		}
	}
}

func TestFieldPartySummaryDosgolemNormalInputComparison(t *testing.T) {
	runFieldStatusDetailNormalAt280(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const sourceHash = "fac03247817ebd9ac9f6aa80c8bba97bc5d9d9ff0f59aba7385575a370a731e2"
		const prefix = "issue4-field-summary-normal-r2"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("normal party summary source identity", e)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 288 || len(src.Queued) != 288 || len(src.Artifacts) != 672 {
			t.Fatal("summary source shape", e)
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
		for n := 281; n <= 288; n++ {
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
			if n == 287 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("detail persistent state differs", n)
			}
			if n == 285 && (g.panel != panelPartySummary || g.cmd.open) {
				t.Fatal("summary page did not wait")
			}
			if n >= 286 && (g.panel != panelNone || g.cmd.open || g.dlg.open) {
				t.Fatal("fresh Enter did not return field", n)
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("summary-packet-%03d.png", n))
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
			if n == 285 && diff != 0 {
				t.Fatal("native party summary full640x350 differs", diff)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "words": g.items.Words(), "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))})
			t.Logf("normal party summary packet%d complete640x350 RGB difference=%d", n, diff)
		}
		// The enclosing normal route performs F5/F6 and the next field move after
		// this callback. Keep its save comparison aligned with the accepted state.
		b, e := json.MarshalIndent(map[string]any{"scope": "normal new-game through288 sole healthy unlearned hero native party summary, fresh Enter and next moves", "source_sha256": sourceHash, "samples": samples, "game_state_injection": false, "rng_unchanged": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "summary-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
