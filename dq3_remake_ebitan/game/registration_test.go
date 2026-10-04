package game

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"
)

func configureRegistrationFixture(t *testing.T, tv *Tavern) {
	t.Helper()
	p, e := gamepack.BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	l, _ := p.NewGameLabels()
	geo, _ := p.NewGameGeometry()
	tv.setLabels(l)
	tv.setGeometry(geo)
	if e = tv.configure(p, nil); e != nil {
		t.Fatal(e)
	}
	tv.equipment = [4]int{-1, 0x1e, -1, -1}
}
func drainRegistrationText(t *testing.T, tv *Tavern, r *rng.RNG) {
	t.Helper()
	for i := 0; i < 5000; i++ {
		if tv.stage != tavText {
			return
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		if tv.dialogue.waitingForConfirm() {
			in.Confirm = true
		}
		tv.input(in, r)
	}
	t.Fatal("registration text did not return")
}
func traceRegistrationText(t *testing.T, g *Game) {
	t.Helper()
	for i := 0; i < 5000; i++ {
		if g.tavern.stage != tavText {
			return
		}
		in := InputState{DirHeld: -1, DirEdge: -1}
		if g.tavern.dialogue.waitingForConfirm() {
			in.Confirm = true
		}
		if e := g.step(in); e != nil {
			t.Fatal(e)
		}
	}
	t.Fatal("normal registration text did not return")
}
func TestRegistrationNonemptyNameAndAcceptanceBoundary(t *testing.T) {
	g := &Game{}
	configureRegistrationFixture(t, &g.tavern)
	tv := &g.tavern
	tv.open()
	drainRegistrationText(t, tv, &g.prng)
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	drainRegistrationText(t, tv, &g.prng)
	tv.ni.nameZhu = false
	tv.ni.cursor = niCellOK
	before := g.prng
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	if tv.stage != tavName || g.prng != before || len(g.roster) != 0 {
		t.Fatal("empty name must stay without RNG or roster writes")
	}
	tv.ni.nameBuf = []int{0}
	for i := 0; i < 4; i++ {
		g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	}
	if tv.stage != tavAccept || tv.candidate == nil || len(g.roster) != 0 {
		t.Fatal("ability/acceptance boundary missing")
	}
	rolled := g.prng
	g.tavernCreate(InputState{DirEdge: -1, Cancel: true})
	if tv.candidate != nil || len(g.roster) != 0 || rolled != g.prng {
		t.Fatal("cancel wrote a member or rerolled")
	}
	drainRegistrationText(t, tv, &g.prng)
	g.tavernCreate(InputState{DirEdge: 3})
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	drainRegistrationText(t, tv, &g.prng)
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	if tv.active || len(g.roster) != 0 {
		t.Fatal("cancel/againNo/final key did not return")
	}
}
func TestRegistrationFullCapacityFailsClosed(t *testing.T) {
	g := &Game{}
	configureRegistrationFixture(t, &g.tavern)
	tv := &g.tavern
	for i := 0; i < tv.contract.RosterCapacity; i++ {
		g.roster = append(g.roster, &Member{})
	}
	tv.open()
	drainRegistrationText(t, tv, &g.prng)
	before := g.prng
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true})
	if tv.stage != tavInitialQuestion || len(g.roster) != tv.contract.RosterCapacity || g.prng != before {
		t.Fatal("unreviewed full replacement path must not create/delete/reroll")
	}
}

// 私有 oracle 是 dosgolem 自行生成的原版收據；沒有它的 CI 不宣稱 parity。
func TestRegistrationDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_REGISTRY_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem registration oracle")
	}
	out := os.Getenv("DQ3_REGISTRY_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit receipt output directory required")
	}
	for _, birth := range []bool{false, true} {
		runRegistrationNormalComparison(t, dir, out, birth, false, false)
	}
}

func runRegistrationNormalComparison(t *testing.T, dir, out string, birth, entry, selection bool, continueYes ...bool) {
	if len(continueYes) > 2 || len(continueYes) > 0 && continueYes[0] && !selection {
		t.Fatal("invalid recruitment comparison route")
	}
	yes := len(continueYes) > 0 && continueYes[0]
	join := len(continueYes) == 2 && continueYes[1]
	if join && (!selection || yes) {
		t.Fatal("invalid join route")
	}
	runRegistrationNormalComparisonRoute(t, dir, out, registrationComparisonRoute{birth: birth, entry: entry, selection: selection, yes: yes, join: join})
}

type registrationComparisonRoute struct {
	birth, entry, selection, yes, join, view, detail bool
}

func runRegistrationNormalComparisonRoute(t *testing.T, dir, out string, route registrationComparisonRoute) {
	birth, entry, selection := route.birth, route.entry, route.selection
	yes, join, view := route.yes, route.join, route.view
	detail := route.detail
	if detail && !view {
		t.Fatal("detail requires view route")
	}
	if view && (!birth || !entry || !selection || yes || join) {
		t.Fatal("invalid view route")
	}
	t.Run(fmt.Sprintf("birth=%v", birth), func(t *testing.T) {
		family := "issue4-registry-quiescent-r2"
		suffix := "-cancel-source-r1-receipt.json"
		want := 164
		if birth {
			family = "issue4-registry-birth-normal-r2"
			suffix = "-source-r1-receipt.json"
			want = 172
		}
		if entry {
			family = "issue4-recruit-normal-r4"
			suffix = "-source-r2-receipt.json"
			want = 196
		}
		if selection {
			family, suffix, want = "issue4-recruit-selection-cancel-r1", "-source-r2-receipt.json", 201
		}
		if yes {
			family, suffix, want = "issue4-recruit-yes-r1", "-source-r1-receipt.json", 204
		}
		if join {
			family, suffix, want = "issue4-recruit-party-r2", "-source-r1-receipt.json", 199
		}
		if view {
			family, suffix, want = "issue4-recruit-view-cancel-r1", "-source-r1-receipt.json", 203
		}
		if detail {
			family, suffix, want = "issue4-recruit-view-detail-r1", "-source-r1-receipt.json", 205
		}
		stem := fmt.Sprintf("registration-birth-%v", birth)
		if entry {
			stem = "recruitment-entry"
		}
		if selection {
			stem = "recruitment-selection"
		}
		if yes {
			stem = "recruitment-continue"
		}
		if join {
			stem = "recruitment-join"
		}
		if view {
			stem = "recruitment-view-cancel"
		}
		if detail {
			stem = "recruitment-view-detail"
		}
		path := filepath.Join(dir, family+suffix)
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		oracleHash := "d59315c19fc5de86cd900b7becdb8c0d3487c18bd3ed72177f4a788a21be2682"
		if yes {
			oracleHash = "cf0730f10e48673e2da6702c77a6e458269cfe0153216b8770b7d3889a08e829"
		}
		if join {
			oracleHash = "d0f6428dbc6f66b0887c3991c4bd17cb00be2825df4a53e1cf5bc049d806ed32"
		}
		if view {
			oracleHash = "cf23fcf92bbd599feb8a2bf1a2b6d7092e2450d187cb36a570798e899eb355bd"
		}
		if detail {
			oracleHash = "e61060c7db7007c5a76e3790f4d55cc2c252619fd04c5cf572f615bc8fb69ee0"
		}
		if selection && fmt.Sprintf("%x", sha256.Sum256(raw)) != oracleHash {
			t.Fatal("selection oracle receipt identity differs")
		}
		var src struct {
			Queued []map[string]string `json:"queued"`
			States []map[string]string `json:"states"`
		}
		if e = json.Unmarshal(raw, &src); e != nil {
			t.Fatal(e)
		}
		if len(src.Queued) != want || len(src.States) != want {
			t.Fatal("oracle packet shape")
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
			for i := 0; i < 5000; i++ {
				if g.recruit.active && (g.recruit.stage == rcGreeting || g.recruit.stage == rcText || g.recruit.stage >= rcJoinedText && g.recruit.stage <= rcFinishText) && !g.recruit.dialogue.waitingForConfirm() || g.tavern.active && g.tavern.stage == tavText && !g.tavern.dialogue.waitingForConfirm() || g.regionDialogueReturn != nil || !g.dlg.open && !g.tavern.active && !g.recruit.active && g.cd > 0 || g.dlg.open && !g.dlg.waitingForConfirm() {
					step(idle)
					continue
				}
				return
			}
			t.Fatal("normal input did not settle")
		}
		var samples []map[string]any
		var viewSnapshot []byte
		var viewRNG rng.RNG
		var detailCanvas []byte
		for i, q := range src.Queued {
			settle()
			scan, e := strconv.ParseInt(q["scan"], 16, 64)
			if e != nil {
				t.Fatal(e)
			}
			in := idle
			in.AnyKeyEdge = true
			if scan == 0x1c {
				in.Enter = true
				in.Confirm = true
			} else if scan == 0x01 {
				in.Cancel = true
			} else {
				d, ok := map[int64]int{0x50: 0, 0x48: 1, 0x4b: 2, 0x4d: 3}[scan]
				if !ok {
					t.Fatal("unknown scan")
				}
				in.DirEdge = d
				in.DirHeld = d
			}
			step(in)
			step(idle)
			settle()
			// 現行命令窗的 Talk 多一個正式确认；不聲稱 raw key parity。
			if (q["kind"] == "registry_approach" || q["kind"] == "recruit_approach") && scan == 0x1c {
				step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true, AnyKeyEdge: true})
				step(idle)
				settle()
			}
			s := src.States[i]
			x, _ := strconv.Atoi(s["player_x"])
			y, _ := strconv.Atoi(s["player_y"])
			section, _ := strconv.ParseInt(s["raw0b24"], 16, 64)
			gold, _ := strconv.ParseInt(s["gold_lo"], 16, 64)
			data, e := os.ReadFile(filepath.Join(spineAssetsDir(t), fmt.Sprintf("CTY%02d.DAT", g.curCty)))
			if e != nil {
				t.Fatal(e)
			}
			if g.px != x || g.py != y || int(binary.LittleEndian.Uint16(data[g.cur.sec*2:])) != int(section) || g.heroGold != int(gold) || fmt.Sprintf("%x", g.storyBits) != s["flags"] {
				t.Fatalf("packet%d native field state differs", i+1)
			}
			if i+1 >= 150 {
				expected := tavName
				switch i + 1 {
				case 150, 151:
					expected = tavText
				case 152:
					expected = tavInitialQuestion
				}
				if birth {
					switch i + 1 {
					case 165:
						expected = tavClass
					case 166:
						expected = tavGender
					case 167:
						expected = tavReview
					case 168:
						expected = tavAccept
					case 169, 170:
						expected = tavAgain
					case 171:
						expected = tavFinalWait
					case 172:
						expected = -1
					}
				} else {
					switch i + 1 {
					case 160:
						expected = tavText
					case 161, 162:
						expected = tavAgain
					case 163:
						expected = tavFinalWait
					case 164:
						expected = -1
					}
				}
				if entry && i+1 > 172 {
					expected = -1
				}
				if expected < 0 {
					if g.tavern.active {
						t.Fatal("registry did not return")
					}
				} else if !g.tavern.active || g.tavern.stage != expected {
					t.Fatalf("packet%d stage%d want%d, name=%v", i+1, g.tavern.stage, expected, g.tavern.ni.nameBuf)
				}
				n := 0
				if birth && i+1 >= 169 {
					n = 1
				}
				companions := 0
				if join && i+1 >= 198 {
					n = 0
					companions = 1
				}
				if len(g.roster) != n || len(g.companions) != companions {
					t.Fatalf("packet%d roster write boundary", i+1)
				}
				if entry && i+1 >= 173 {
					if i+1 < 195 && g.recruit.active {
						t.Fatal("recruitment opened before Talk")
					}
					if i+1 == 195 && (!g.recruit.active || g.recruit.stage != rcGreeting || !g.recruit.dialogue.waitingForConfirm()) {
						t.Fatal("native greeting inline wait differs")
					}
					if i+1 == 196 && (!g.recruit.active || g.recruit.stage != rcMenu || g.recruit.cursor != 0 || g.recruit.dialogue.open) {
						t.Fatal("first native recruitment menu differs")
					}
					if view && i+1 == 196 {
						viewRNG = g.prng
						viewSnapshot, e = encodeSave(g.snapshot())
						if e != nil {
							t.Fatal(e)
						}
					}
				}
				if selection && i+1 >= 197 {
					stages := map[int]int{197: rcJoin, 198: rcAgain, 199: rcAgain, 200: rcFinalWait}
					cursors := map[int]int{197: 0, 198: 0, 199: 1}
					if yes {
						stages = map[int]int{197: rcJoin, 198: rcAgain, 199: rcMenu, 200: rcJoin, 201: rcAgain, 202: rcAgain, 203: rcFinalWait}
						cursors = map[int]int{197: 0, 198: 0, 199: 0, 200: 0, 201: 0, 202: 1}
					}
					if join {
						stages = map[int]int{197: rcJoin, 198: rcJoinedText, 199: rcMusicWait}
						cursors = map[int]int{197: 0}
					}
					if view {
						stages = map[int]int{197: rcMenu, 198: rcMenu, 199: rcView, 200: rcAgain, 201: rcAgain, 202: rcFinalWait}
						cursors = map[int]int{197: 1, 198: 2, 199: 0, 200: 0, 201: 1}
						if detail {
							stages = map[int]int{197: rcMenu, 198: rcMenu, 199: rcView, 200: rcViewAbility, 201: rcViewClose, 202: rcAgain, 203: rcAgain, 204: rcFinalWait}
							cursors = map[int]int{197: 1, 198: 2, 199: 0, 200: 0, 201: 0, 202: 0, 203: 1}
						}
						current, err := encodeSave(g.snapshot())
						if err != nil || string(current) != string(viewSnapshot) || g.prng != viewRNG {
							t.Fatal("view changed persistent state or RNG")
						}
					}
					if expected, ok := stages[i+1]; ok && (!g.recruit.active || g.recruit.stage != expected) {
						t.Fatalf("packet%d recruitment stage%d want%d", i+1, g.recruit.stage, expected)
					}
					if !join && i+1 == want && g.recruit.active {
						t.Fatal("cancel did not return to field")
					}
					if cursor, ok := cursors[i+1]; ok && g.recruit.cursor != cursor {
						t.Fatal("native selection cursor differs")
					}
				}
				g.renderFrame()
				if detail && i+1 == 200 {
					compareRecruitmentViewActorFixture(t, g, path, out, fmt.Sprintf("%s-packet-%03d-%s.png", family, i+1, s["phase"]))
					detailCanvas = append([]byte(nil), g.rgba...)
				}
				if detail && i+1 == 201 && string(g.rgba) != string(detailCanvas) {
					t.Fatal("ability canvas changed between key waits")
				}
				f, e := os.Create(filepath.Join(out, fmt.Sprintf("%s-packet-%03d.png", stem, i+1)))
				if e != nil {
					t.Fatal(e)
				}
				e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
				f.Close()
				if e != nil {
					t.Fatal(e)
				}
				var npcVisuals []map[string]int
				for _, npc := range g.cur.npcs {
					npcVisuals = append(npcVisuals, map[string]int{"record": npc.recordIndex, "x": npc.x, "y": npc.y, "facing": npc.facing, "walk": npc.walk})
				}
				samples = append(samples, map[string]any{"packet": i + 1, "original_phase": s["phase"], "stage": g.tavern.stage, "active": g.tavern.active, "recruit_stage": g.recruit.stage, "recruit_active": g.recruit.active, "roster": len(g.roster), "candidate": g.tavern.candidate, "recruit_waiting": g.recruit.dialogue.waitingForConfirm(), "hero_facing": g.facing, "hero_walk": g.walk, "npc_visuals": npcVisuals, "full_rgb_difference": sourceCanvasDifference(t, g, path, fmt.Sprintf("%s-packet-%03d-%s.png", family, i+1, s["phase"]))})
			}
		}
		if birth {
			var m *Member
			if join {
				m = g.companions[0]
			} else {
				m = g.roster[0]
			}
			if !reflect.DeepEqual(m.Name, []int{0}) || m.Class != 1 || m.Gender != 0 || m.Level() != 1 || m.Armor != 0x1e || m.Weapon != -1 || m.Shield != -1 || m.Head != -1 {
				t.Fatalf("accepted record differs: %+v", m)
			}
		}

		if join {
			if samples[len(samples)-1]["full_rgb_difference"] != 295 || g.recruit.musicFrames <= 0 {
				t.Fatal("normal199 text/caller comparison", samples[len(samples)-1])
			}
			// Runtime-only continuation: original oracle stops at audio entry.
			for i := 0; i < 5000 && (g.recruit.stage == rcMusicWait || g.recruit.stage == rcText); i++ {
				step(idle)
			}
			if g.recruit.stage != rcAgain {
				t.Fatal("one-shot failed to continue")
			}
			step(InputState{DirEdge: 3, DirHeld: 3})
			step(idle)
			step(InputState{DirEdge: -1, DirHeld: -1, Enter: true, Confirm: true})
			step(idle)
			settle()
			if g.recruit.stage != rcFinalWait {
				t.Fatal("normal No after one-shot")
			}
			step(InputState{DirEdge: -1, DirHeld: -1, Enter: true, Confirm: true})
			step(idle)
			settle()
			if g.recruit.active || len(g.companions) != 1 || len(g.roster) != 0 {
				t.Fatal("normal join return lost member")
			}
		}
		if e = g.Save(); e != nil {
			t.Fatal(e)
		}
		saved, _ := encodeSave(g.snapshot())
		if e = g.Load(); e != nil {
			t.Fatal(e)
		}
		loaded, _ := encodeSave(g.snapshot())
		if string(saved) != string(loaded) || g.tavern.active || g.tavern.candidate != nil || g.recruit.active {
			t.Fatal("registration Save/Load roundtrip")
		}
		if entry {
			x, _ := strconv.Atoi(src.States[len(src.States)-1]["player_x"])
			y, _ := strconv.Atoi(src.States[len(src.States)-1]["player_y"])
			step(InputState{DirHeld: 2, DirEdge: 2})
			step(idle)
			settle()
			if g.px != x-1 || g.py != y || g.recruit.active {
				t.Fatal("normal step after same-version Load did not continue")
			}
		}
		if birth && !entry {
			traceWalkThroughPortal(t, g, 8, 2, 0, 0)
			traceTalkNPC(t, g, 2, 16)
			traceRecruitmentGreeting(t, g)
			if !g.recruit.active {
				t.Fatal("normal lower-floor recruit entry")
			}
			step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
			step(idle)
			settle()
			step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true})
			step(idle)
			if len(g.roster) != 0 || len(g.companions) != 1 || g.companions[0].Class != 1 {
				t.Fatal("normal recruitment lost/duplicated member")
			}
			if e = g.Save(); e != nil {
				t.Fatal(e)
			}
			saved, _ = encodeSave(g.snapshot())
			if e = g.Load(); e != nil {
				t.Fatal(e)
			}
			loaded, _ = encodeSave(g.snapshot())
			if string(saved) != string(loaded) {
				t.Fatal("recruited Save/Load")
			}
		}
		b, e := json.MarshalIndent(map[string]any{"source": path, "seed": 0x1357, "seed_method": "once at cold start both sides; global dice order not aligned", "comparison": "equivalent Talk extra confirmation; original stats not exact; all full RGB differences retained", "pack_schema": g.pack.Manifest.SchemaVersion, "pack_hash": g.pack.ContentHash(), "samples": samples}, "", "  ")
		if e != nil {
			t.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(out, stem+".json"), append(b, '\n'), 0644); e != nil {
			t.Fatal(e)
		}
	})
}

func TestRegistrationLoadClearsCandidateOnlyAfterValidation(t *testing.T) {
	g, e := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if e != nil {
		t.Fatal(e)
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "registration.json"))
	if e = g.Save(); e != nil {
		t.Fatal(e)
	}
	book, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	g.tavern.active = true
	g.tavern.stage = tavReview
	g.tavern.candidate = &Member{Name: []int{0}}
	candidate := g.tavern.candidate
	contract := g.tavern.contract
	r := g.prng
	bad := g.snapshot()
	bad.PackSchema = "invalid"
	raw, e := encodeSave(bad)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(savePath(), raw, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e == nil {
		t.Fatal("invalid schema accepted")
	}
	if !g.tavern.active || g.tavern.candidate != candidate || g.prng != r {
		t.Fatal("rejected Load changed pending candidate")
	}
	if e = os.WriteFile(savePath(), book, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	if g.tavern.active || g.tavern.candidate != nil || g.tavern.contract != contract || g.tavern.raster == nil || len(g.roster) != 0 {
		t.Fatal("valid Load did not clear only transient registration")
	}
}
