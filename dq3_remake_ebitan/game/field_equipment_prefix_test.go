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

func runFieldSpellEmptyNormalAt304(t *testing.T, after func(*Game)) {
	runFieldStatusReorderNormalAt298(t, func(g *Game) {
		dir, dest := os.Getenv("DQ3_ITEM_ORDERED_ORACLE_DIR"), os.Getenv("DQ3_ITEM_ORDERED_RECEIPT_DIR")
		const prefix = "issue4-spell-empty-normal-r1"
		source := filepath.Join(dir, prefix+"-source-r1-receipt.json")
		raw, e := os.ReadFile(source)
		if e != nil || fmt.Sprintf("%x", sha256.Sum256(raw)) != "c096702865ce694706ba2aac2f9333408b895bf45d856bae4a14334d1275cfcf" {
			t.Fatal("source identity", e)
		}
		var src struct {
			Queued, States []map[string]string
			Artifacts      []struct {
				Path   string
				Size   int
				SHA256 string
			}
		}
		if e = json.Unmarshal(raw, &src); e != nil || len(src.States) != 304 || len(src.Artifacts) != 720 {
			t.Fatal("source shape", e)
		}
		for _, a := range src.Artifacts {
			b, e := os.ReadFile(filepath.Join(dir, a.Path))
			if e != nil || len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA256 {
				t.Fatal("source artifact", a.Path, e)
			}
		}
		inputs := []InputState{{Confirm: true, DirHeld: -1, DirEdge: -1}, {DirHeld: -1, DirEdge: 3}, {Confirm: true, DirHeld: -1, DirEdge: -1}, {Enter: true, DirHeld: -1, DirEdge: -1}, {DirHeld: 2, DirEdge: 2}, {DirHeld: 3, DirEdge: 3}}
		scans := []string{"39", "4d", "39", "1c", "4b", "4d"}
		idle := InputState{DirHeld: -1, DirEdge: -1}
		before, rng := g.snapshot(), g.prng
		var samples []map[string]any
		for i, in := range inputs {
			n := 299 + i
			if src.Queued[n-1]["scan"] != scans[i] {
				t.Fatal("input", n)
			}
			in.AnyKeyEdge = true
			for g.cd > 0 {
				if e = g.step(idle); e != nil {
					t.Fatal(e)
				}
			}
			if e = g.step(in); e != nil {
				t.Fatal(e)
			}
			if e = g.step(idle); e != nil {
				t.Fatal(e)
			}
			if n == 301 {
				if g.fieldMessagePrompt == nil || g.fieldSpell.active {
					t.Fatal("empty entry did not present native response")
				}
				for j := 0; !g.fieldMessagePrompt.dialogue.waitingForConfirm(); j++ {
					if j >= 1000 {
						t.Fatal("wait")
					}
					if e = g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
			}
			want := before
			if n == 303 {
				want.PX = 2
			}
			if !equalFieldSave(want, g.snapshot()) || rng != g.prng {
				t.Fatal("empty spell changed state", n)
			}
			if n >= 302 && (g.fieldMessagePrompt != nil || g.fieldSpell.active || g.cmd.open || g.panel != panelNone) {
				t.Fatal("fresh-key return", n)
			}
			g.renderFrame()
			p := filepath.Join(dest, fmt.Sprintf("spell-empty-packet-%03d.png", n))
			f, e := os.Create(p)
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if e != nil || ce != nil {
				t.Fatal(e, ce)
			}
			diff := sourceCanvasDifference(t, g, source, fmt.Sprintf(prefix+"-packet-%03d-%s.png", n, src.States[n-1]["phase"]))
			t.Logf("normal spell-empty packet%d RGB difference=%d caster_selector=%t", n, diff, g.fieldSpell.selectingCaster)
			if diff != 0 {
				t.Fatal("normal complete640x350 differs", n, diff)
			}
			b, e := os.ReadFile(p)
			if e != nil {
				t.Fatal(e)
			}
			samples = append(samples, map[string]any{"packet": n, "full_rgb_difference": diff, "png_sha256": fmt.Sprintf("%x", sha256.Sum256(b))})
		}
		after(g)
	})
}
