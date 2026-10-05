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

func TestFieldStatusMenuNavigationAndUnknownResults(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	g.panel, g.panelCursor = panelStatusMenu, 0
	before, rng := g.snapshot(), g.prng
	for _, dir := range []int{0, 0, 0, 1, 1, 1} {
		old := g.panelCursor
		g.stepStatusMenu(InputState{DirEdge: dir}, -1)
		want := (old + 1) % 3
		if dir == 1 {
			want = (old + 2) % 3
		}
		if g.panelCursor != want {
			t.Fatal("cyclic selector differs")
		}
	}
	for _, i := range []int{1, 2} {
		g.panelCursor = i
		g.stepStatusMenu(InputState{DirEdge: -1, Confirm: true}, -1)
		if g.panel != panelStatusMenu || g.panelCursor != i || !equalFieldSave(before, g.snapshot()) || g.prng != rng {
			t.Fatal("unreviewed result transacted state")
		}
	}
	g.panelCursor = 99
	g.stepStatusMenu(InputState{Confirm: true}, -1)
	if !equalFieldSave(before, g.snapshot()) {
		t.Fatal("invalid selector changed state")
	}
}

func TestFieldStatusMenuDosgolemNormalInputComparison(t *testing.T) {
	runFieldStatusMenuNormalAt273(t, nil)
}

func runFieldStatusMenuNormalAt273(t *testing.T, after func(*Game)) {
	runFieldItemDropNormalAt261(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const sourceHash = "a30e50edd8ab3a3f8f8f95532b84aff4773d5a19879385ab31158e4d57b9b9aa"
		source := filepath.Join(dir, "issue4-field-status-normal-r3-source-r1-receipt.json")
		raw, err := os.ReadFile(source)
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != sourceHash {
			t.Fatal("normal status source differs", err)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if err = json.Unmarshal(raw, &src); err != nil || len(src.States) != 273 || len(src.Queued) != 273 || len(src.Artifacts) != 627 {
			t.Fatal("status source shape differs", err)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(filepath.Join(dir, a.Path))
			if filepath.Base(a.Path) != a.Path || e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("status artifact differs", a.Path, e)
			}
		}
		before, rng := g.snapshot(), g.prng
		idle := InputState{DirHeld: -1, DirEdge: -1}
		var samples []map[string]any
		for n := 262; n <= 273; n++ {
			in := idle
			in.AnyKeyEdge = true
			scan, e := strconv.ParseInt(src.Queued[n-1]["scan"], 16, 64)
			if e != nil {
				t.Fatal(e)
			}
			switch scan {
			case 0x39:
				in.Confirm = true
			case 0x50:
				in.DirEdge = 0
			case 0x48:
				in.DirEdge = 1
			case 0x01:
				in.Cancel = true
			case 0x4b:
				in.DirEdge, in.DirHeld = 2, 2
			case 0x4d:
				in.DirEdge, in.DirHeld = 3, 3
			default:
				t.Fatal("unsupported normal status input", n)
			}
			if in.DirHeld >= 0 {
				for updates := 0; g.cd > 0; updates++ {
					if updates >= moveCooldown {
						t.Fatal("normal field movement cooldown did not finish")
					}
					if e := g.step(idle); e != nil {
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
			if n == 272 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				actual := g.snapshot()
				t.Fatalf("status state differs packet%d PX=%d/%d DNStep=%d/%d DNPhase=%d/%d EncounterStep=%d/%d RNG=%d/%d", n, actual.PX, want.PX, actual.DNStep, want.DNStep, actual.DNPhase, want.DNPhase, actual.EncounterStep, want.EncounterStep, g.prng, rng)
			}
			if n >= 264 && n <= 270 {
				cursor, e := strconv.Atoi(src.States[n-1]["choice_cursor"])
				if e != nil {
					t.Fatal(e)
				}
				if g.panel != panelStatusMenu || g.panelCursor != cursor-1 || g.cmd.open {
					t.Fatal("native first selector differs", n, g.panel, g.panelCursor)
				}
			}
			if n >= 271 && (g.panel != panelNone || g.cmd.open || g.dlg.open) {
				t.Fatal("Esc did not return to field", n)
			}
			g.renderFrame()
			path := filepath.Join(dest, fmt.Sprintf("status-packet-%03d.png", n))
			f, e := os.Create(path)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf("issue4-field-status-normal-r3-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			b, e := os.ReadFile(path)
			if e != nil {
				t.Fatal(e)
			}
			var npcs []map[string]int
			for _, npc := range g.cur.npcs {
				npcs = append(npcs, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "words": g.items.Words(), "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b)), "png_size": len(b), "npc_visuals": npcs})
			t.Logf("normal status packet%d complete640x350 RGB difference=%d", n, diff)
		}
		b, e := json.MarshalIndent(map[string]any{"scope": "normal new-game through273 sole healthy hero first status menu, navigation, Esc and next step", "source_sha256": sourceHash, "samples": samples, "game_state_injection": false, "persistent_status_unchanged": true, "rng_unchanged": true, "cancel_returns_to_field": true, "save_version": saveFormatVersion, "pack_schema": g.pack.Schema(), "pack_content_version": g.pack.ContentVersion(), "pack_hash": g.pack.ContentHash(), "animation_timing_parity": false}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(dest, "status-receipt.json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
		if after != nil {
			after(g)
		}
	})
}
