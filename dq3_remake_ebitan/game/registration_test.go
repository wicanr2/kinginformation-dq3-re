package game

import (
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
		runRegistrationNormalComparison(t, dir, out, birth, false)
	}
}

func runRegistrationNormalComparison(t *testing.T, dir, out string, birth, entry bool) {
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
		stem := fmt.Sprintf("registration-birth-%v", birth)
		if entry {
			stem = "recruitment-entry"
		}
		path := filepath.Join(dir, family+suffix)
		raw, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
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
				if g.recruit.active && g.recruit.stage == rcGreeting && !g.recruit.dialogue.waitingForConfirm() || g.tavern.active && g.tavern.stage == tavText && !g.tavern.dialogue.waitingForConfirm() || g.regionDialogueReturn != nil || !g.dlg.open && !g.tavern.active && g.cd > 0 || g.dlg.open && !g.dlg.waitingForConfirm() {
					step(idle)
					continue
				}
				return
			}
			t.Fatal("normal input did not settle")
		}
		var samples []map[string]any
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
				if len(g.roster) != n || len(g.companions) != 0 {
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
				}
				g.renderFrame()
				f, e := os.Create(filepath.Join(out, fmt.Sprintf("%s-packet-%03d.png", stem, i+1)))
				if e != nil {
					t.Fatal(e)
				}
				e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
				f.Close()
				if e != nil {
					t.Fatal(e)
				}
				samples = append(samples, map[string]any{"packet": i + 1, "original_phase": s["phase"], "stage": g.tavern.stage, "active": g.tavern.active, "recruit_stage": g.recruit.stage, "recruit_active": g.recruit.active, "roster": len(g.roster), "candidate": g.tavern.candidate, "full_rgb_difference": sourceCanvasDifference(t, g, path, fmt.Sprintf("%s-packet-%03d-%s.png", family, i+1, s["phase"]))})
			}
		}
		if birth {
			m := g.roster[0]
			if !reflect.DeepEqual(m.Name, []int{0}) || m.Class != 1 || m.Gender != 0 || m.Level() != 1 || m.Armor != 0x1e || m.Weapon != -1 || m.Shield != -1 || m.Head != -1 {
				t.Fatalf("accepted record differs: %+v", m)
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
