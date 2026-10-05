package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/fstest"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestRegionDialogueRewardRejectsBadSources(t *testing.T) {
	assets := os.DirFS(spineAssetsDir(t))
	for _, name := range []string{"nil", "missing_scene", "truncated_scene", "outside_tile", "wrong_handler", "missing_text", "changed_text"} {
		t.Run(name, func(t *testing.T) {
			p, err := gamepack.BuiltinDQ3()
			if err != nil {
				t.Fatal(err)
			}
			e := &p.Events.RegionDialogueRewardEvents[0]
			raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), ctyFile(e.CTYRaw)))
			if err != nil {
				t.Fatal(err)
			}
			d, _ := p.TextDefinition(e.TextID)
			text, err := os.ReadFile(filepath.Join(spineAssetsDir(t), d.Source.File))
			if err != nil {
				t.Fatal(err)
			}
			broken := fstest.MapFS{ctyFile(e.CTYRaw): &fstest.MapFile{Data: raw}, d.Source.File: &fstest.MapFile{Data: text}}
			switch name {
			case "nil":
				if validateRegionDialogueRewardSources(nil, p) == nil {
					t.Fatal("accepted nil source")
				}
				return
			case "missing_scene":
				delete(broken, ctyFile(e.CTYRaw))
			case "truncated_scene":
				broken[ctyFile(e.CTYRaw)].Data = raw[:2]
			case "outside_tile":
				e.Tile.X = 65535
			case "wrong_handler":
				e.HandlerRaw++
			case "missing_text":
				delete(broken, d.Source.File)
			case "changed_text":
				d.GlyphCodes[0]++
			}
			if validateRegionDialogueRewardSources(broken, p) == nil {
				t.Fatal("accepted incompatible native source")
			}
		})
	}
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if err = validateRegionDialogueRewardSources(assets, p); err != nil {
		t.Fatal(err)
	}
}

func TestDosgolemRegionDialogueRewardNormalInput(t *testing.T) {
	path := os.Getenv("DQ3_KING_TEXT_ORIGINAL")
	if path == "" {
		t.Skip("set DQ3_KING_TEXT_ORIGINAL to the verified cold-boot receipt")
	}
	g, _ := runDosgolemThroneLanding(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	up := InputState{DirHeld: 1, DirEdge: 1, AnyKeyEdge: true}
	enter := InputState{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}
	step := func(in InputState) {
		t.Helper()
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	press := func(in InputState) {
		t.Helper()
		for i := 0; g.cd > 0 && !g.dlg.open && i < 120; i++ {
			step(idle)
		}
		if g.cd > 0 && !g.dlg.open {
			t.Fatal("cooldown did not finish")
		}
		step(in)
		step(idle)
		for i := 0; g.cd > 0 && !g.dlg.open && i < 120; i++ {
			step(idle)
		}
	}
	naturalIdle := func() {
		t.Helper()
		for i := 0; i < 300 && !g.fieldIdle.open; i++ {
			step(idle)
		}
		if !g.fieldIdle.open {
			t.Fatal("natural idle missing")
		}
	}
	wait := func() {
		t.Helper()
		for i := 0; i < 3000 && g.dlg.open && !g.dlg.waitingForConfirm(); i++ {
			step(idle)
		}
	}
	naturalIdle()
	press(up)
	if g.py != 22 {
		t.Fatal("idle dismissal moved player")
	}
	for i := 0; i < 14; i++ {
		press(up)
	}
	naturalIdle()
	for i := 0; i < 5; i++ {
		press(enter)
	}
	if g.px != 9 || g.py != 8 || g.dlg.open || g.heroGold != 0 || len(testInventory(g.items)) != 0 {
		t.Fatal("normal ninety-input checkpoint differs")
	}
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	beforeFile, err := os.ReadFile(savePath())
	if err != nil {
		t.Fatal(err)
	}
	e := g.pack.Events.RegionDialogueRewardEvents[0]
	unchanged := func() {
		t.Helper()
		if g.heroGold != 0 || len(testInventory(g.items)) != 0 || !g.storyFlag(e.RequiredFlagRaw) || g.storyFlag(e.SetFlagRaw) || g.progressDone(e.ProgressFlagRaw) {
			t.Fatal("reward transaction occurred before native text return")
		}
	}
	var samples []map[string]any
	capture := func(ordinal int) {
		t.Helper()
		g.renderFrame()
		name := fmt.Sprintf("issue4-king-text-text-text_waiting-%02d.png", ordinal)
		full := sourceCanvasDifference(t, g, path, name)
		f, err := os.Open(filepath.Join(filepath.Dir(path), name))
		if err != nil {
			t.Fatal(err)
		}
		original, err := png.Decode(f)
		f.Close()
		if err != nil {
			t.Fatal(err)
		}
		w := g.dlg.layout
		panel := 0
		for y := w.Y; y < w.Y+w.Height+e.Shadow.OffsetY; y++ {
			for x := w.X; x < w.X+w.Width+e.Shadow.OffsetX; x++ {
				r, gg, b, _ := original.At(x, y).RGBA()
				o := (y*ScreenW + x) * 4
				if uint8(r>>8) != g.rgba[o] || uint8(gg>>8) != g.rgba[o+1] || uint8(b>>8) != g.rgba[o+2] {
					panel++
				}
			}
		}
		if panel != 0 {
			t.Fatalf("native wait %d window/shadow differs at %d pixels", ordinal, panel)
		}
		if out := os.Getenv("DQ3_KING_TEXT_OUT"); out != "" {
			f, err := os.Create(fmt.Sprintf("%s-wait-%02d.png", out, ordinal))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			ce := f.Close()
			if err != nil || ce != nil {
				t.Fatal(err, ce)
			}
		}
		samples = append(samples, map[string]any{"wait": ordinal, "full_rgb_difference": full, "window_and_shadow_rgb_difference": panel, "gold": g.heroGold, "inventory": append([]int{}, testInventory(g.items)...), "flag17": g.storyFlag(e.RequiredFlagRaw), "flag18": g.storyFlag(e.SetFlagRaw), "seed": g.prng.State()})
	}
	press(up)
	wait()
	if g.py != 7 || g.regionDialogueReward == nil || !g.dlg.waitingForConfirm() {
		t.Fatal("normal step did not reach audience wait")
	}
	for i := 1; i <= 9; i++ {
		unchanged()
		capture(i)
		press(enter)
		wait()
		if i < 9 && !g.dlg.waitingForConfirm() {
			t.Fatal("Enter did not reach next native wait", i)
		}
	}
	if g.dlg.open || g.regionDialogueReward != nil || g.dlg.prelude != nil || g.dlg.shadow != nil || g.heroGold != e.Gold || !reflect.DeepEqual(testInventory(g.items), e.ItemRawIDs) || g.storyFlag(e.ClearFlagRaw) || !g.storyFlag(e.SetFlagRaw) || !g.progressDone(e.ProgressFlagRaw) {
		t.Fatal("native EOF reward or modal restore differs")
	}
	// Further inputs cannot grant again, and normal movement can leave the tile.
	for i := 0; i < 3; i++ {
		press(enter)
	}
	if g.heroGold != e.Gold || !reflect.DeepEqual(testInventory(g.items), e.ItemRawIDs) {
		t.Fatal("audience reward duplicated")
	}
	press(InputState{DirHeld: 0, DirEdge: 0, AnyKeyEdge: true})
	if g.py != 8 {
		t.Fatal("normal continuation blocked")
	}
	if err = g.Save(); err != nil {
		t.Fatal(err)
	}
	saved, _ := json.Marshal(g.snapshot())
	completedFile, err := os.ReadFile(savePath())
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	for _, in := range []InputState{enter, {DirHeld: -1, DirEdge: 0}, enter} {
		if err = loaded.step(in); err != nil {
			t.Fatal(err)
		}
	}
	actual, _ := json.Marshal(loaded.snapshot())
	if loaded.showTitle || loaded.dlg.open || loaded.regionDialogueReward != nil || !bytes.Equal(saved, actual) {
		t.Fatal("normal title Load lost completed reward state")
	}
	// Restore a real pre-audience save, enter again normally, then Load while
	// text is pending. The pending callback must never reward restored state.
	if err = os.WriteFile(savePath(), beforeFile, 0644); err != nil {
		t.Fatal(err)
	}
	if err = g.Load(); err != nil {
		t.Fatal(err)
	}
	press(up)
	wait()
	unchanged()
	if g.regionDialogueReward == nil {
		t.Fatal("normal retry did not start")
	}
	if err = g.Load(); err != nil || g.regionDialogueReward != nil || g.dlg.open || g.dlg.prelude != nil || g.dlg.shadow != nil {
		t.Fatal("same-instance Load retained pending transaction", err)
	}
	step(idle)
	unchanged()
	if g.py != 8 {
		t.Fatal("Load failed to restore pre-audience checkpoint")
	}
	if err = os.WriteFile(savePath(), completedFile, 0644); err != nil {
		t.Fatal(err)
	}
	if err = g.Load(); err != nil {
		t.Fatal(err)
	}
	actual, _ = json.Marshal(g.snapshot())
	if !bytes.Equal(saved, actual) {
		t.Fatal("same-instance Load changed completed transaction")
	}
	if out := os.Getenv("DQ3_KING_TEXT_OUT"); out != "" {
		b, err := json.MarshalIndent(map[string]any{"schema": g.pack.Schema(), "content_version": g.pack.Manifest.ContentVersion, "content_hash": g.pack.ContentHash(), "normal_inputs": 100, "state_injection": false, "samples": samples, "native_waits": 9, "dialogue_before_rewards": true, "once_only": true, "normal_continuation": true, "title_save_load": true, "same_instance_load_clears_pending": true, "window_parity": true, "full_rgb_parity": false, "audio_parity": false, "original_save_load_oracle": false}, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.WriteFile(out+".json", append(b, '\n'), 0644); err != nil {
			t.Fatal(err)
		}
	}
}
