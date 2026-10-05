package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestFieldStatusReorderFreshWaitAndState(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	before, rng := g.snapshot(), g.prng
	g.panel, g.panelCursor, g.cmd.open = panelStatusMenu, 2, false
	g.stepStatusMenu(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}, -1)
	if g.statusReorderPrompt == nil {
		t.Fatal("single-member response absent")
	}
	// Selection and an early Enter cannot consume the later wait.
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	wait := func() {
		t.Helper()
		for n := 0; !g.statusReorderPrompt.dialogue.waitingForConfirm(); n++ {
			if n >= 1000 {
				t.Fatal("message did not reach fresh wait")
			}
			if e := g.step(idle); e != nil {
				t.Fatal(e)
			}
		}
	}
	wait()
	if g.statusReorderPrompt.dialogue.retained.finished {
		t.Fatal("inline wait skipped")
	}
	if e := g.step(InputState{DirHeld: 2, DirEdge: 2, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	wait()
	if !g.statusReorderPrompt.dialogue.retained.finished {
		t.Fatal("terminal fresh wait absent")
	}
	for n := 0; n < 20; n++ {
		if e := g.step(idle); e != nil {
			t.Fatal(e)
		}
	}
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	if g.statusReorderPrompt != nil || g.panel != panelNone || g.cmd.open || g.fieldSaveLoad.active ||
		!equalFieldSave(before, g.snapshot()) || rng != g.prng {
		t.Fatal("dismissal leaked action or transacted state")
	}
	g.panel, g.panelCursor = panelStatusMenu, 2
	g.stepStatusMenu(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}, -1)
	g.restore(before)
	if g.statusReorderPrompt != nil {
		t.Fatal("restore retained transient prompt")
	}
}

func TestFieldStatusReorderDosgolemNormalInputComparison(t *testing.T) {
	runFieldPartySummaryNormalAt288(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-field-reorder-normal-r2"
		const sourceHash = "e06a5e419d255303140c1516af0aa09021bbb967e94b0db1e0db89d1aca90370"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("normal reorder source identity", e)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 298 || len(src.Queued) != 298 || len(src.Artifacts) != 702 {
			t.Fatal("normal reorder source shape", e)
		}
		for _, a := range src.Artifacts {
			b, err := os.ReadFile(filepath.Join(dir, a.Path))
			if filepath.Base(a.Path) != a.Path || err != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("normal reorder artifact identity", a.Path, err)
			}
		}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		before, rng := g.snapshot(), g.prng
		inputs := []InputState{
			{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 0},
			{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 0}, {DirHeld: -1, DirEdge: 0},
			{Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1},
			{DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3},
		}
		scans := []string{"39", "50", "39", "50", "50", "39", "1c", "1c", "4b", "4d"}
		var samples []map[string]any
		for i, in := range inputs {
			n := 289 + i
			if src.Queued[n-1]["scan"] != scans[i] {
				t.Fatal("normal reorder equivalent input", n)
			}
			in.AnyKeyEdge = true
			if in.DirHeld >= 0 {
				for updates := 0; g.cd > 0; updates++ {
					if updates >= moveCooldown {
						t.Fatal("normal next-step cooldown")
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
			if n == 294 || n == 295 {
				if g.statusReorderPrompt == nil {
					t.Fatal("normal reorder message absent", n)
				}
				for updates := 0; !g.statusReorderPrompt.dialogue.waitingForConfirm(); updates++ {
					if updates >= 1000 {
						t.Fatal("normal message wait timeout", n)
					}
					if e = g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
				if g.statusReorderPrompt.dialogue.retained.finished != (n == 295) {
					t.Fatal("normal wait phase", n)
				}
			}
			want := before
			if n == 297 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("normal persistent state changed", n)
			}
			if n >= 296 && (g.statusReorderPrompt != nil || g.panel != panelNone || g.cmd.open) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("reorder-packet-%03d.png", n))
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
			if diff != 0 {
				t.Fatal("normal reorder complete640x350 differs", n, diff)
			}
			b, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "words": g.items.Words()})
			t.Logf("normal reorder packet%d complete640x350 RGB difference=%d", n, diff)
		}
		b, e := json.MarshalIndent(map[string]any{"scope": "normal new-game through298 single-member reorder two retained pages, fresh keys and next moves", "source_sha256": sourceHash, "samples": samples, "game_state_injection": false, "rng_unchanged": true, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "save_version": saveFormatVersion, "animation_timing_parity": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "reorder-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}
