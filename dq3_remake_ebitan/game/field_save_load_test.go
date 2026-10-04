package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func equalFieldSave(a, b saveState) bool {
	x, ex := encodeSave(a)
	y, ey := encodeSave(b)
	return ex == nil && ey == nil && bytes.Equal(x, y)
}

func settleFieldSaveLoad(t *testing.T, g *Game) {
	t.Helper()
	for n := 0; n < 5000; n++ {
		m := &g.fieldSaveLoad
		if !m.active || m.stage != fsSoundWait && (m.stage != fsText || m.dialogue.waitingForConfirm()) {
			return
		}
		if e := g.step(InputState{DirHeld: -1, DirEdge: -1}); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("save/load modal did not settle")
}

func TestFieldSaveLoadDosgolemNormalInputComparison(t *testing.T) {
	runFieldSaveLoadNormalInputComparison(t, false)
}

func TestFieldSaveLoadAfterLoadMoveDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_AFTER_LOAD_MOVE_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem normal movement after F6")
	}
	t.Setenv("DQ3_FIELD_SAVE_LOAD_ORACLE_DIR", dir)
	runFieldSaveLoadNormalInputComparison(t, true)
}

func runFieldSaveLoadNormalInputComparison(t *testing.T, afterLoad bool) {
	t.Helper()
	dir, out := os.Getenv("DQ3_FIELD_SAVE_LOAD_ORACLE_DIR"), os.Getenv("DQ3_FIELD_SAVE_LOAD_RECEIPT_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem F5/F6 oracle")
	}
	if out == "" {
		t.Fatal("explicit output required")
	}
	type routeSpec struct {
		name, prefix, hash string
		packets            int
	}
	routes := []routeSpec{
		{"decline", "issue4-save-f5-decline-r1", "08a3feadd29c2567a806d3bd0bda121214a5a57323bf993f969a0d2cbda963d2", 197},
		{"cancel", "issue4-save-load-cancel-r1", "30a4f5471277bb040d4dfcf62958098ac412e11d5d641235f9cf378c2fcbd32c", 197},
		{"roundtrip", "issue4-save-load-roundtrip-r1", "0a88466eab206d7a4b346b8229bbcebd10683044aa363cb346bb81e173e3d238", 200},
	}
	if afterLoad {
		routes = []routeSpec{{"after-load-move", "issue4-after-load-move-r1", "71a52768a26dabb174a3a96e53efe6aaebba78c7620b20cce39fe2f761548e10", 202}}
	}
	for _, route := range routes {
		t.Run(route.name, func(t *testing.T) {
			roundtrip := route.name == "roundtrip" || afterLoad
			dest := filepath.Join(out, route.name)
			if e := os.Mkdir(dest, 0755); e != nil {
				t.Fatal(e)
			}
			prefixOut := filepath.Join(dest, "prefix")
			if e := os.Mkdir(prefixOut, 0755); e != nil {
				t.Fatal(e)
			}
			t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_ORACLE_DIR", dir)
			t.Setenv("DQ3_RECRUIT_EMPTY_MENU_CANCEL_RECEIPT_DIR", prefixOut)
			runRecruitmentEmptyNormalInputComparisonAtCheckpoint(t, "menu-cancel", "issue4-recruit-menu-esc-normal-r1-source-r1-receipt.json", "137411c711e81684943b4cc8aac2d952b8c47c8981f70a79f0d308e369204c4c", "issue4-recruit-menu-esc-normal-r1", 193, 189, 190, func(g *Game) {
				path := filepath.Join(dir, route.prefix+"-source-r1-receipt.json")
				if afterLoad {
					path = filepath.Join(dir, route.prefix+"-source-r2-receipt.json")
				}
				raw, e := os.ReadFile(path)
				if e != nil {
					t.Fatal(e)
				}
				if fmt.Sprintf("%x", sha256.Sum256(raw)) != route.hash {
					t.Fatal("source identity differs")
				}
				var src struct{ Queued, States, Clocks []map[string]string }
				if e = json.Unmarshal(raw, &src); e != nil || len(src.Queued) != route.packets {
					t.Fatal(e, "source shape")
				}
				t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "field-save.json"))
				before := g.snapshot()
				rng := g.prng
				var saved saveState
				var samples []map[string]any
				idle := InputState{DirHeld: -1, DirEdge: -1}
				for i := 193; i < route.packets; i++ {
					n := i + 1
					q := src.Queued[i]
					scan, e := strconv.ParseInt(q["scan"], 16, 64)
					if e != nil {
						t.Fatal(e)
					}
					in := idle
					in.AnyKeyEdge = true
					switch scan {
					case 0x3f:
						in.SaveMenu = true
					case 0x40:
						in.LoadMenu = true
					case 0x1c:
						in.Enter = true
					case 1:
						in.Cancel = true
					case 0x4b:
						in.DirEdge = 2
						in.DirHeld = 2
					case 0x4d:
						in.DirEdge = 3
						in.DirHeld = 3
					default:
						t.Fatal("unsupported source scan")
					}
					if e = g.step(in); e != nil {
						t.Fatal(e)
					}
					if afterLoad && n > 200 {
						x, _ := strconv.Atoi(src.States[i]["player_x"])
						y, _ := strconv.Atoi(src.States[i]["player_y"])
						// The native packet ends after a whole move and release.
						// Replay the held direction through the engine cooldown,
						// then release at that same tile boundary. No state writes.
						for updates := 0; g.px != x || g.py != y; updates++ {
							if updates >= 60 {
								t.Fatal("normal held direction did not complete source move")
							}
							held := idle
							held.DirHeld = in.DirHeld
							if e = g.step(held); e != nil {
								t.Fatal(e)
							}
						}
					}
					if e = g.step(idle); e != nil {
						t.Fatal(e)
					}
					settleFieldSaveLoad(t, g)
					m := &g.fieldSaveLoad
					s := src.States[i]
					x, _ := strconv.Atoi(s["player_x"])
					y, _ := strconv.Atoi(s["player_y"])
					if g.px != x || g.py != y || fmt.Sprintf("%x", g.storyBits) != s["flags"] || g.prng != rng {
						t.Fatalf("normal state differs at %d: got position %d,%d cooldown %d; want %d,%d", n, g.px, g.py, g.cd, x, y)
					}
					if route.name == "decline" {
						if n <= 195 && (!m.active || m.stage != fsQuestion || m.cursor != n-194) || n == 196 && (!m.active || m.stage != fsFinalWait) || n == 197 && m.active {
							t.Fatalf("No phase differs at %d stage%d", n, m.stage)
						}
					} else if n == 194 && (!m.active || m.stage != fsQuestion || m.cursor != 0) || n == 195 && (!m.active || m.stage != fsSlots || m.cursor != 0) || n == 196 && (!m.active || m.stage != fsFinalWait) || n == 197 && m.active || roundtrip && n == 199 && (!m.active || !m.loading || m.stage != fsSlots) || n >= 200 && m.active {
						t.Fatalf("source phase differs at %d stage%d", n, m.stage)
					}
					current := g.snapshot()
					if !roundtrip {
						if !reflect.DeepEqual(before, current) {
							t.Fatal("cancel changed persistent state")
						}
						if _, e = os.Stat(savePath()); !os.IsNotExist(e) {
							t.Fatal("cancel wrote save")
						}
					}
					if roundtrip && n == 196 {
						b, e := os.ReadFile(savePath())
						if e != nil {
							t.Fatal(e)
						}
						saved, e = decodeSave(b)
						if e != nil || !equalFieldSave(saved, current) {
							t.Fatal("normal F5 write differs", e)
						}
					}
					if roundtrip && (n == 200 || afterLoad && n > 200) {
						want := saved
						want.PX, want.PY = x, y
						clock, ok := m.contract.LoadClock(saved.Layer)
						if !ok {
							t.Fatal("clock rule missing")
						}
						phaseTicks := g.dayNightCycle.ClockTicks / 4
						want.DNPhase, want.DNStep = clock/phaseTicks, clock%phaseTicks
						if !equalFieldSave(want, current) || g.dayNightClock() != clock {
							t.Fatal("normal F6 restore/clock differs")
						}
					}
					if afterLoad {
						if len(src.Clocks) != 10 || src.Clocks[n-193]["packet"] != strconv.Itoa(n) || src.Clocks[n-193]["clock"] != strconv.Itoa(g.dayNightClock()) {
							t.Fatalf("normal source clock differs at %d", n)
						}
					}
					g.renderFrame()
					name := fmt.Sprintf("packet-%03d.png", n)
					f, e := os.Create(filepath.Join(dest, name))
					if e != nil {
						t.Fatal(e)
					}
					e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
					ce := f.Close()
					if e != nil || ce != nil {
						t.Fatal(e, ce)
					}
					diff := sourceCanvasDifference(t, g, path, fmt.Sprintf("%s-packet-%03d-%s.png", route.prefix, n, s["phase"]))
					if n == 194 || n == 196 || n == 197 || n == 200 {
						if diff != 0 {
							t.Fatalf("same-state whole RGB differs at %d: %d", n, diff)
						}
					}
					samples = append(samples, map[string]any{"packet": n, "original_phase": s["phase"], "stage": m.stage, "active": m.active, "clock": g.dayNightClock(), "full_rgb_difference": diff})
				}
				for g.cd > 0 {
					if e := g.step(idle); e != nil {
						t.Fatal(e)
					}
				}
				if afterLoad {
					beforeLoad := g.snapshot()
					if e := g.Save(); e != nil {
						t.Fatal(e)
					}
					if e := g.Load(); e != nil || !equalFieldSave(beforeLoad, g.snapshot()) || g.prng != rng {
						t.Fatal("engine save/load after normal movement differs", e)
					}
				}
				if e := g.step(InputState{DirHeld: 2, DirEdge: 2}); e != nil {
					t.Fatal(e)
				}
				if g.px != 1 || g.py != 18 || g.fieldSaveLoad.active {
					t.Fatal("normal next field step")
				}
				report := map[string]any{"source_sha256": route.hash, "pack_schema": g.pack.Schema(), "pack_hash": g.pack.ContentHash(), "normal_input_prefix": 193, "normal_final_packet": route.packets, "rng_unchanged": true, "native_f5_f6": roundtrip, "normal_after_load_movement": afterLoad, "samples": samples, "initial_storage_same_state": false, "limitation": "original PLAYER.DAT contains ten prior entries; remake starts with no JSON saves; whole RGB slot differences retained; no full V3 claim"}
				b, e := json.MarshalIndent(report, "", "  ")
				if e != nil {
					t.Fatal(e)
				}
				if e = os.WriteFile(filepath.Join(dest, "receipt.json"), append(b, '\n'), 0644); e != nil {
					t.Fatal(e)
				}
			}, true)
		})
	}
}

func TestFieldSaveLoadSlotPathsPreservePrimary(t *testing.T) {
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "checkpoint.json"))
	seen := map[string]bool{}
	for i := 0; i < 10; i++ {
		path := fieldSaveSlotPath(i)
		if seen[path] || filepath.Dir(path) != filepath.Dir(savePath()) {
			t.Fatal("slot collision or directory changed")
		}
		seen[path] = true
		if i == 0 && path != savePath() {
			t.Fatal("primary path changed")
		}
	}
}

func fieldSaveLoadComponentGame(t *testing.T) *Game {
	t.Helper()
	g, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	// Component fixture only. The normal cold-start test above owns path parity.
	g.showTitle = false
	g.openingIdx = -1
	g.curCty = -1       // Normal overworld metadata, as restored by loadFrom.
	g.addVisitedTown(0) // Opening's proven starting town; avoids legacy empty-list migration.
	g.heroName = []int{0}
	_, maxHP, _, _, _ := g.heroStats()
	g.heroHP = maxHP
	g.heroMP = g.heroMaxMP()
	g.renderFrame()
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "checkpoint.json"))
	return g
}

func TestFieldSaveLoadNoAndSlotCancelNeverWrite(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	before := g.snapshot()
	rng := g.prng
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		settleFieldSaveLoad(t, g)
	}
	step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
	if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsQuestion {
		t.Fatal("normal F5 question missing")
	}
	for i := 0; i < 20; i++ {
		step(idle)
	}
	if g.fieldSaveLoad.stage != fsQuestion || g.fieldSaveLoad.cursor != 0 {
		t.Fatal("held/no edge changed choice")
	}
	step(InputState{DirHeld: -1, DirEdge: 3})
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	if g.fieldSaveLoad.stage != fsFinalWait {
		t.Fatal("No must wait for separate key")
	}
	for i := 0; i < 20; i++ {
		step(idle)
	}
	if !g.fieldSaveLoad.active {
		t.Fatal("held key dismissed farewell")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, AnyKeyEdge: true})
	step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
	step(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
	if g.fieldSaveLoad.stage != fsSlots {
		t.Fatal("Yes cursor Esc must open slots")
	}
	step(InputState{DirHeld: -1, DirEdge: 0})
	if g.fieldSaveLoad.cursor != 1 {
		t.Fatal("slot navigation")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
	if g.fieldSaveLoad.stage != fsFinalWait {
		t.Fatal("slot Esc farewell missing")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	if g.fieldSaveLoad.active || g.prng != rng || !reflect.DeepEqual(before, g.snapshot()) {
		t.Fatal("cancel mutated persistence or RNG")
	}
	entries, e := os.ReadDir(filepath.Dir(savePath()))
	if e != nil || len(entries) != 0 {
		t.Fatal("cancel wrote file", e)
	}
	step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
	if g.fieldSaveLoad.active {
		t.Fatal("empty load must return without modal")
	}
}

func TestFieldSaveLoadIndependentSlotsAndRejectedLoad(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	if e := g.Save(); e != nil {
		t.Fatal(e)
	}
	primary, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	idle := InputState{DirHeld: -1, DirEdge: -1}
	step := func(in InputState) {
		t.Helper()
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		settleFieldSaveLoad(t, g)
	}
	step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true})
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	step(InputState{DirHeld: -1, DirEdge: 0})
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	if g.fieldSaveLoad.stage != fsFinalWait {
		t.Fatal("successful save farewell")
	}
	second, e := os.ReadFile(fieldSaveSlotPath(1))
	if e != nil {
		t.Fatal(e)
	}
	unchanged, e := os.ReadFile(savePath())
	if e != nil || string(unchanged) != string(primary) {
		t.Fatal("secondary slot overwrote primary")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
	if len(g.fieldSaveLoad.slots) != 2 || g.fieldSaveLoad.slots[1].index != 1 {
		t.Fatal("load compressed slots lost actual index")
	}
	step(InputState{DirHeld: -1, DirEdge: 0})
	m := &g.fieldSaveLoad
	// Component state changed after the saved empty roster/party. Rejection
	// must preserve it; a valid load must restore both saved empty collections.
	g.roster = []*Member{newMember([]int{1}, 1, 0, 0)}
	g.companions = []*Member{newMember([]int{2}, 2, 0, 0)}
	flow, raster, background, cursor, rng, before := m.dialogue.retained, m.raster, append([]byte(nil), m.background...), m.cursor, g.prng, g.snapshot()
	s, e := decodeSave(second)
	if e != nil {
		t.Fatal(e)
	}
	s.PackSchema = "invalid"
	bad, e := encodeSave(s)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(fieldSaveSlotPath(1), bad, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true}); e == nil {
		t.Fatal("incompatible load accepted")
	}
	if !m.active || m.stage != fsSlots || m.cursor != cursor || m.raster != raster || m.dialogue.retained != flow || !reflect.DeepEqual(background, m.background) || g.prng != rng || !reflect.DeepEqual(before, g.snapshot()) {
		t.Fatal("rejected load changed modal or state")
	}
	if e = os.WriteFile(fieldSaveSlotPath(1), second, 0644); e != nil {
		t.Fatal(e)
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Enter: true})
	expected, e := decodeSave(second)
	clock, ok := m.contract.LoadClock(expected.Layer)
	if !ok {
		t.Fatal("clock rule missing")
	}
	phaseTicks := g.dayNightCycle.ClockTicks / 4
	expected.DNPhase, expected.DNStep = clock/phaseTicks, clock%phaseTicks
	if e != nil || m.active || m.dialogue.retained != nil || g.prng != rng || !equalFieldSave(expected, g.snapshot()) {
		t.Fatalf("accepted load differs: err=%v active=%v retained=%v RNG=%v bytes_equal=%v", e, m.active, m.dialogue.retained != nil, g.prng != rng, equalFieldSave(expected, g.snapshot()))
	}
	// Removing a listed save between selection and confirmation must be an error.
	step(InputState{DirHeld: -1, DirEdge: -1, LoadMenu: true})
	if e = os.Remove(savePath()); e != nil {
		t.Fatal(e)
	}
	before = g.snapshot()
	if e = g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true}); !os.IsNotExist(e) {
		t.Fatal("missing listed slot must fail", e)
	}
	if !m.active || !reflect.DeepEqual(before, g.snapshot()) {
		t.Fatal("missing slot changed state")
	}
	step(InputState{DirHeld: -1, DirEdge: -1, Cancel: true})
	if m.active {
		t.Fatal("load Esc must return directly")
	}
	_ = idle
}

func TestFieldSaveLoadFailedWritePreservesModal(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	for _, in := range []InputState{{DirHeld: -1, DirEdge: -1, SaveMenu: true}, {DirHeld: -1, DirEdge: -1, Enter: true}} {
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		settleFieldSaveLoad(t, g)
	}
	if e := os.Mkdir(savePath(), 0755); e != nil {
		t.Fatal(e)
	}
	before, rng, flow := g.snapshot(), g.prng, g.fieldSaveLoad.dialogue.retained
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true}); e == nil {
		t.Fatal("directory accepted as save file")
	}
	if !g.fieldSaveLoad.active || g.fieldSaveLoad.stage != fsSlots || g.fieldSaveLoad.dialogue.retained != flow || g.prng != rng || !reflect.DeepEqual(before, g.snapshot()) {
		t.Fatal("failed write changed modal or persistence")
	}
}

func TestFieldSaveLoadExistingModalsHavePriority(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	g.cmd.open = true
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true, LoadMenu: true}); e != nil {
		t.Fatal(e)
	}
	if g.fieldSaveLoad.active || !g.cmd.open {
		t.Fatal("F5/F6 stole command modal")
	}
	g.cmd.open = false
	g.recruit.open()
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, SaveMenu: true}); e != nil {
		t.Fatal(e)
	}
	if g.fieldSaveLoad.active || !g.recruit.active {
		t.Fatal("F5 stole recruitment modal")
	}
}

type fieldSaveAudioTrace struct {
	gameAudio
	calls             []string
	cues              []int
	duration          int64
	playedBeforeWrite bool
}

func (a *fieldSaveAudioTrace) PauseBackground()  { a.calls = append(a.calls, "pause") }
func (a *fieldSaveAudioTrace) ResumeBackground() { a.calls = append(a.calls, "resume") }
func (a *fieldSaveAudioTrace) PlaySFX(cue int) {
	a.calls = append(a.calls, "cue")
	a.cues = append(a.cues, cue)
	_, e := os.Stat(savePath())
	a.playedBeforeWrite = os.IsNotExist(e)
}
func (a *fieldSaveAudioTrace) SFXDurationNanos(int) int64 { return a.duration }

func TestFieldSaveLoadSoundCompletionRejectsEarlyInput(t *testing.T) {
	g := fieldSaveLoadComponentGame(t)
	audio := &fieldSaveAudioTrace{gameAudio: g.music, duration: 1_000_000_000}
	g.music = audio
	for _, in := range []InputState{{DirHeld: -1, DirEdge: -1, SaveMenu: true}, {DirHeld: -1, DirEdge: -1, Enter: true}} {
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
		settleFieldSaveLoad(t, g)
	}
	if e := g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true}); e != nil {
		t.Fatal(e)
	}
	m := &g.fieldSaveLoad
	if m.stage != fsSoundWait || m.soundFrames != 60 || !audio.playedBeforeWrite || !reflect.DeepEqual(audio.calls, []string{"pause", "cue"}) || !reflect.DeepEqual(audio.cues, []int{m.contract.Sound.CueRaw}) {
		t.Fatal("save cue/write order or hardware approximation differs")
	}
	written, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	rng := g.prng
	for i := 0; i < 60; i++ {
		if e = g.step(InputState{DirHeld: 0, DirEdge: -1, Enter: true, Cancel: true, SaveMenu: true, LoadMenu: true, AnyKeyEdge: true}); e != nil {
			t.Fatal(e)
		}
		if m.stage != fsSoundWait {
			t.Fatal("early input bypassed audio wait")
		}
	}
	if e = g.step(InputState{DirHeld: -1, DirEdge: -1, Enter: true, AnyKeyEdge: true}); e != nil {
		t.Fatal(e)
	}
	settleFieldSaveLoad(t, g)
	if m.stage != fsFinalWait || !m.active || g.prng != rng || !reflect.DeepEqual(audio.calls, []string{"pause", "cue", "resume"}) {
		t.Fatal("sound completion must reveal farewell and await separate key")
	}
	now, e := os.ReadFile(savePath())
	if e != nil || string(now) != string(written) {
		t.Fatal("early input rewrote save")
	}
}
