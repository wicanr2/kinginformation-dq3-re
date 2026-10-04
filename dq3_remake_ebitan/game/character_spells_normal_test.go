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

func TestCharacterSpellsDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_CHARACTER_SPELLS_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem character-spell oracle")
	}
	path := filepath.Join(dir, "issue4-view-spells-class3-normal-r2-source-r2-receipt.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%x", sha256.Sum256(raw)) != "4dcb99c80e5cddc17940b09772c37169d5348cc40249996827029812365b8c08" {
		t.Fatal("original accepted receipt identity differs")
	}
	var src struct{ Queued, States []map[string]string }
	if err = json.Unmarshal(raw, &src); err != nil || len(src.States) != 209 {
		t.Fatal(err, "source shape")
	}
	g := runDosgolemMotherArrivalStateComparison(t)
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
	}
	settle := func() {
		t.Helper()
		for n := 0; n < 5000; n++ {
			if g.recruit.active && (g.recruit.stage == rcGreeting || g.recruit.stage == rcText || g.recruit.stage >= rcJoinedText && g.recruit.stage <= rcFinishText) && !g.recruit.dialogue.waitingForConfirm() || g.tavern.active && g.tavern.stage == tavText && !g.tavern.dialogue.waitingForConfirm() || g.regionDialogueReturn != nil || !g.dlg.open && !g.tavern.active && !g.recruit.active && g.cd > 0 || g.dlg.open && !g.dlg.waitingForConfirm() {
				step(idle)
				continue
			}
			return
		}
		t.Fatal("normal input did not settle")
	}
	out := os.Getenv("DQ3_CHARACTER_SPELLS_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit output required")
	}
	var samples []map[string]any
	var persistent []byte
	var holdRNG = g.prng
	var abilityCanvas []byte
	var candidateJSON []byte
	var creationRNG = g.prng
	for i, q := range src.Queued {
		if i == 198 {
			persistent, _ = encodeSave(g.snapshot())
			holdRNG = g.prng
		}
		settle()
		scan, e := strconv.ParseInt(q["scan"], 16, 64)
		if e != nil {
			t.Fatal(e)
		}
		in := idle
		in.AnyKeyEdge = true
		switch scan {
		case 0x1c:
			in.Enter = true
			in.Confirm = true
		case 1:
			in.Cancel = true
		default:
			d, ok := map[int64]int{0x50: 0, 0x48: 1, 0x4b: 2, 0x4d: 3}[scan]
			if !ok {
				t.Fatal("scan")
			}
			in.DirEdge = d
			in.DirHeld = d
		}
		step(in)
		step(idle)
		settle()
		if (q["kind"] == "registry_approach" || q["kind"] == "recruit_approach") && scan == 0x1c {
			step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true, AnyKeyEdge: true})
			step(idle)
			settle()
		}
		s := src.States[i]
		x, _ := strconv.Atoi(s["player_x"])
		y, _ := strconv.Atoi(s["player_y"])
		if g.px != x || g.py != y || fmt.Sprintf("%x", g.storyBits) != s["flags"] {
			t.Fatalf("normal field state differs at%d", i+1)
		}

		switch i + 1 {
		case 169:
			if g.tavern.stage != tavReview || len(g.roster) != 0 {
				t.Fatal("ability wait or early registration")
			}
			candidateJSON, _ = json.Marshal(g.tavern.candidate)
			creationRNG = g.prng
		case 170, 171:
			expected := tavSpells
			if i+1 == 171 {
				expected = tavAccept
			}
			now, _ := json.Marshal(g.tavern.candidate)
			if g.tavern.stage != expected || len(g.roster) != 0 || string(now) != string(candidateJSON) || g.prng != creationRNG {
				t.Fatal("spell wait changed candidate or registered early")
			}
		case 203, 204, 205:
			expected := map[int]int{203: rcViewAbility, 204: rcViewSpells, 205: rcViewClose}[i+1]
			if g.recruit.stage != expected || g.recruit.viewFlow == nil {
				t.Fatal("view wait sequence differs")
			}
		}
		if i+1 >= 199 {
			now, e := encodeSave(g.snapshot())
			if e != nil || string(now) != string(persistent) || g.prng != holdRNG {
				t.Fatal("view changed persistent state or RNG")
			}
		}
		if i+1 >= 165 {
			g.renderFrame()
			if i+1 == 203 {
				abilityCanvas = append([]byte(nil), g.rgba...)
			}
			if i+1 == 205 && string(abilityCanvas) != string(g.rgba) {
				t.Fatal("ability canvas changed after closing spell page")
			}
			name := fmt.Sprintf("packet-%03d.png", i+1)
			f, e := os.Create(filepath.Join(out, name))
			if e != nil {
				t.Fatal(e)
			}
			e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
			closeErr := f.Close()
			if e != nil || closeErr != nil {
				t.Fatal(e)
			}
			sample := map[string]any{"packet": i + 1, "original_phase": s["phase"], "tavern_stage": g.tavern.stage, "recruit_stage": g.recruit.stage, "candidate": g.tavern.candidate, "full_rgb_difference": sourceCanvasDifference(t, g, path, fmt.Sprintf("issue4-view-spells-class3-normal-r2-packet-%03d-%s.png", i+1, s["phase"]))}
			samples = append(samples, sample)
		}
	}
	if g.recruit.active || len(g.roster) != 1 || len(g.companions) != 0 || g.roster[0].Class != 3 || len(g.roster[0].LearnedSpells) != 1 || g.roster[0].LearnedSpells[0] != 161 {
		t.Fatal("normal full return differs")
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "spell-page.json"))
	if e := g.Save(); e != nil {
		t.Fatal(e)
	}
	saved, e := encodeSave(g.snapshot())
	if e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	loaded, e := encodeSave(g.snapshot())
	if e != nil || string(saved) != string(loaded) || g.tavern.active || g.recruit.active {
		t.Fatal("normal Save/Load")
	}
	step(InputState{DirHeld: 2, DirEdge: 2})
	step(idle)
	settle()
	if g.px != 1 || g.py != 18 || g.recruit.active {
		t.Fatal("next step after Load")
	}
	report := map[string]any{"source_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)), "pack_schema": g.pack.Manifest.SchemaVersion, "pack_hash": g.pack.ContentHash(), "scope": "normal production input; finite state and spell-layer parity; full RGB parity remains RED", "original_packets": 209, "save_load_next_step": true, "samples": samples}
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(out, "receipt.json"), append(data, '\n'), 0644); e != nil {
		t.Fatal(e)
	}
	t.Log("正式209包、三次觀看等待、存讀檔及下一步通過；完整RGB差異保留")
}
