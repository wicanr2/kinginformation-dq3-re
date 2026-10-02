package game

import (
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestOpeningHomeNormalInput(t *testing.T) {
	for _, navigation := range []bool{false, true} {
		t.Run(fmt.Sprintf("navigation=%v", navigation), func(t *testing.T) {
			t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "home-save.json"))
			g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
			if err != nil {
				t.Fatal(err)
			}
			g.prng.Seed(0x1357)
			seed := uint16(0x151b)
			g.homeSelection.seed = &seed
			idle := InputState{DirHeld: -1, DirEdge: -1}
			step := func(in InputState) {
				t.Helper()
				if err := g.step(in); err != nil {
					t.Fatal(err)
				}
			}
			confirm := func() { in := idle; in.Enter = true; step(in) }
			wait := func(done func() bool) {
				t.Helper()
				for i := 0; i < 3000 && !done(); i++ {
					step(idle)
				}
				if !done() {
					t.Fatalf("normal input timeout: opening=%d escort=%d:%d dlg=%v picture=%v", g.openingIdx, g.openingEscortPhase, g.openingEscortIndex, g.dlg.open, g.homeSelection.active)
				}
			}
			for _, scan := range []int{0x1c, 0x1c, 0x48, 0x4b, 0x1c, 0x1c, 0x4d, 0x50, 0x1c, 0x48, 0x4b, 0x1c, 0x48, 0x1c, 0x1c, 0x1c, 0x1c} {
				in := idle
				switch scan {
				case 0x1c:
					in.Enter = true
				case 0x48:
					in.DirEdge = 1
				case 0x50:
					in.DirEdge = 0
				case 0x4b:
					in.DirEdge = 2
				case 0x4d:
					in.DirEdge = 3
				}
				step(in)
			}
			wait(func() bool { return g.dlg.waitingForConfirm() })
			confirm()
			wait(func() bool { return g.openingIdx == 1 && g.dlg.waitingForConfirm() })
			confirm()
			seen := []string{}
			last := -1
			for updates := 0; updates < 3000 && g.openingIdx != 2; updates++ {
				if g.openingEscortPhase == 0 && g.openingEscortIndex != last {
					n := g.cur.npcs[g.openingEscortNPC]
					if n.recordIndex != 0 || g.px != 5 || g.py != 5 || g.cur.npcAt(8, 3) < 0 {
						t.Fatal("home actor identity or stationary player mismatch")
					}
					seen = append(seen, fmt.Sprintf("%d,%d,%d", n.x, n.y, n.facing))
					last = g.openingEscortIndex
				}
				step(idle)
			}
			want := []string{"5,4,0", "4,4,2", "3,4,2", "3,5,0", "3,6,0", "3,7,0", "3,8,0", "3,9,0", "3,10,0", "4,10,3", "5,10,3", "6,10,3", "7,10,3", "8,10,3", "9,10,3", "10,10,3", "10,10,2"}
			if !reflect.DeepEqual(seen, want) || !reflect.DeepEqual(g.dlg.buf, g.dlg.tx.Record(81)) {
				t.Fatalf("home source sequence mismatch: %v", seen)
			}
			wait(func() bool { return g.homeSelection.active })
			h := &g.homeSelection
			if h.question != 13 || !reflect.DeepEqual(h.options, [][]int{{3, 4, 5, 57, 58, 59}, {33, 34, 35, 48, 49, 50}, {6, 7, 8, 24, 25, 26}}) || !reflect.DeepEqual(h.expected, []int{1, 1, 4}) || g.prng.State() != 0x356d {
				t.Fatal("natural clock seed source fixture differs or ability RNG consumed")
			}
			capture := func(name string) {
				t.Helper()
				if !navigation {
					verifyHomeWindow(t, g, name)
				}
				out := os.Getenv("DQ3_HOME_TEST_OUT")
				if out == "" {
					return
				}
				img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
				copy(img.Pix, g.rgba)
				f, err := os.Create(filepath.Join(out, "issue4-remake-home-"+name+".png"))
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				if err := png.Encode(f, img); err != nil {
					t.Fatal(err)
				}
			}
			capture("preview")
			confirm()
			capture("round-1")
			if navigation {
				in := idle
				in.Cancel = true
				step(in)
				if h.round != 1 || h.choices[0] != 1 {
					t.Fatal("source Escape commits current option")
				}
				for i, dir := range []int{2, 1, 3, 0} {
					in := idle
					in.DirEdge = dir
					step(in)
					if h.cursor != []int{5, 4, 5, 0}[i] {
						t.Fatal("source direction wrap differs")
					}
				}
				confirm()
			} else {
				confirm()
				capture("round-2")
				confirm()
			}
			capture("round-3")
			confirm()
			if h.stage != 2 || !reflect.DeepEqual(h.choices, []int{1, 1, 1}) || !h.active {
				t.Fatal("source unchecked third choice did not reach result")
			}
			capture("result")
			confirm()
			if h.active || !g.homeAwait || g.cur.sec != 4 || g.px != 5 || g.py != 5 || g.cur.npcAt(10, 10) < 0 || !g.storyFlag(80) {
				t.Fatal("control not returned in original home state")
			}
			capture("control")
			if err := g.Save(); err != nil {
				t.Fatal(err)
			}
			restored, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
			if err != nil {
				t.Fatal(err)
			}
			g = restored
			confirm()
			loadChoice := idle
			loadChoice.DirEdge = 0
			step(loadChoice)
			confirm()
			if restored.showTitle {
				t.Fatal("normal title load entry did not return control")
			}
			if !restored.homeAwait || restored.cur.npcAt(10, 10) < 0 || restored.px != 5 || restored.py != 5 {
				t.Fatal("same-version home save failed")
			}
			g = restored
			for _, dir := range []int{0, 0, 2, 2, 0, 0, 0, 3, 3, 3, 3, 3, 3} {
				for g.cd > 0 {
					step(idle)
				}
				in := idle
				in.DirHeld = dir
				in.DirEdge = dir
				step(in)
			}
			if g.px != 9 || g.py != 10 || g.openingEscortPhase != 5 || g.cur.npcAt(10, 11) < 0 || g.cur.sec != 4 {
				t.Fatalf("normal home approach missing: player=%d,%d phase=%d", g.px, g.py, g.openingEscortPhase)
			}
			wait(func() bool { return g.openingEscortPhase == 2 && g.dlg.waitingForConfirm() })
			if g.px != 22 || g.py != 19 || g.cur.npcAt(22, 18) < 0 || !g.storyFlag(80) || g.storyFlag(23) {
				t.Fatal("town dialogue gate mismatch")
			}
			confirm()
			wait(func() bool { return g.openingEscortPhase < 0 })
			if g.px != 21 || g.py != 17 || g.storyFlag(80) || !g.storyFlag(23) {
				t.Fatal("town final transaction mismatch")
			}
			capture("castle")
			if err := g.Save(); err != nil {
				t.Fatal(err)
			}
			if err := g.Load(); err != nil {
				t.Fatal(err)
			}
			if g.homeAwait || g.px != 21 || g.py != 17 || !g.storyFlag(23) {
				t.Fatal("town save failed")
			}
			if out := os.Getenv("DQ3_HOME_TEST_OUT"); out != "" {
				raw, _ := json.MarshalIndent(map[string]any{"normal_input": true, "ability_seed": "1357", "picture_seed": "151b", "home_states": seen, "question": h.question, "options": h.options, "choices": h.choices, "home_save_round_trip": true, "town_save_round_trip": true, "full_rgb_parity": false}, "", "  ")
				if err := os.WriteFile(filepath.Join(out, fmt.Sprintf("issue4-remake-home-navigation-%v-receipt.json", navigation)), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}
