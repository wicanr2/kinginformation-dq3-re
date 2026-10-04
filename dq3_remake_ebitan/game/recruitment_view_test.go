package game

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func TestRecruitmentViewDetailDosgolemNormalInputComparison(t *testing.T) {
	dir := os.Getenv("DQ3_RECRUIT_DETAIL_ORACLE_DIR")
	if dir == "" {
		t.Skip("optional private dosgolem view detail oracle")
	}
	out := os.Getenv("DQ3_RECRUIT_DETAIL_RECEIPT_DIR")
	if out == "" {
		t.Fatal("explicit detail receipt output directory required")
	}
	runRegistrationNormalComparisonRoute(t, dir, out, registrationComparisonRoute{birth: true, entry: true, selection: true, view: true, detail: true})
}

// 只在正常路徑取樣點另驗同角色 renderer。原版97bytes不是正式遊戲狀態，
// 此元件畫面另外命名，不取代保留各自HP的正常205包。
func compareRecruitmentViewActorFixture(t *testing.T, g *Game, receipt, out, originalName string) {
	t.Helper()
	var source struct {
		Detail []map[string]string `json:"view_detail"`
	}
	b, e := os.ReadFile(receipt)
	if e != nil {
		t.Fatal(e)
	}
	if e = json.Unmarshal(b, &source); e != nil || len(source.Detail) != 4 {
		t.Fatal("detail source missing original actor")
	}
	raw, e := hex.DecodeString(source.Detail[2]["actor520b"])
	if e != nil || len(raw) != 97 || source.Detail[2]["ida_linear"] != "1834e" {
		t.Fatal("original actor identity differs")
	}
	word := func(offset int) int { return int(binary.LittleEndian.Uint16(raw[offset:])) }
	current := g.recruit.viewFlow
	fixture := *current
	fixture.preview = stats.Values{uint16(word(0x1a)), uint16(word(0x1e)), uint16(word(0x24)), uint16(word(0x2a)), uint16(word(0x2c)), uint16(word(0x26)), uint16(word(0x28))}
	fixture.previewLevel, fixture.previewHP, fixture.previewMP = int(raw[0x15]), word(0x16), word(0x18)
	fixture.previewDef, fixture.previewExp = word(0x20), int(binary.LittleEndian.Uint32(raw[0x32:]))
	if word(0x1c) != word(0x1a) || raw[0x2e] != 0 || raw[0x2f] != 0 || raw[0x30] != 0 || raw[0x31] != 0 {
		t.Fatal("fixture is outside reviewed unarmed no-spell scope")
	}
	g.recruit.viewFlow = &fixture
	defer func() { g.recruit.viewFlow = current; g.renderFrame() }()
	g.renderFrame()
	if diff := sourceCanvasDifference(t, g, receipt, originalName); diff != 7 {
		t.Fatalf("same actor component full RGB difference %d, expected existing exposed NPC pixels only", diff)
	}
	f, e := os.Create(filepath.Join(out, "recruitment-view-detail-same-actor-packet-200.png"))
	if e != nil {
		t.Fatal(e)
	}
	e = png.Encode(f, &image.RGBA{Pix: g.rgba, Stride: ScreenW * 4, Rect: image.Rect(0, 0, ScreenW, ScreenH)})
	closeErr := f.Close()
	if e != nil || closeErr != nil {
		t.Fatal(fmt.Sprint(e, closeErr))
	}
}

func recruitmentViewGame(t *testing.T) *Game {
	t.Helper()
	g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	// 元件fixture沒有經過正常NPC交談；顯式安裝同一個pack繪圖器。
	if err = g.recruit.installRaster(g.newGame.raster, g.worldPal); err != nil {
		t.Fatal(err)
	}
	g.roster = []*Member{newLevelOneMember([]int{0}, g.tavern.contract.ClassOptions[0].ClassRaw, 0, &g.prng, g.tavern.equipment)}
	g.recruit.active, g.recruit.stage = true, rcView
	return g
}

func TestRecruitmentViewDetailTwoWaitsPreserveState(t *testing.T) {
	g := recruitmentViewGame(t)
	rc := &g.recruit
	before, e := encodeSave(g.snapshot())
	if e != nil {
		t.Fatal(e)
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcViewAbility || rc.viewFlow == nil {
		t.Fatalf("selection failed to open ability wait: member=%+v initial=%v stage=%d raster=%v items=%v", g.roster[0], g.tavern.equipment, rc.stage, rc.raster != nil, g.shop.items != nil)
	}
	for i := 0; i < 100; i++ {
		g.recruitInput(InputState{DirEdge: -1, DirHeld: 0})
	}
	if rc.stage != rcViewAbility {
		t.Fatal("held direction consumed a fresh-key wait")
	}
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if rc.stage != rcViewClose || rc.viewFlow == nil {
		t.Fatal("first key skipped the separate close wait")
	}
	g.recruitInput(InputState{DirEdge: -1, AnyKeyEdge: true})
	if rc.stage != rcViewClose {
		t.Fatal("unknown key guessed the rename branch")
	}
	g.recruitInput(InputState{DirEdge: -1, Cancel: true})
	drainRecruitmentSelectionText(t, rc)
	if rc.stage != rcAgain || rc.viewFlow != nil {
		t.Fatal("close must restore caller and ask continue")
	}
	g.recruitInput(InputState{DirEdge: 3})
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	drainRecruitmentSelectionText(t, rc)
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	after, e := encodeSave(g.snapshot())
	if e != nil || string(before) != string(after) || rc.active {
		t.Fatal("view changed persistent state, RNG or return path")
	}
}

func TestRecruitmentViewDetailUnknownStateStaysInList(t *testing.T) {
	for name, change := range map[string]func(*Member){
		"spells":    func(m *Member) { m.LearnedSpells = []int{1} },
		"equipment": func(m *Member) { m.Weapon = 0 },
		"condition": func(m *Member) { m.Conditions = conditionPoison },
		"class":     func(m *Member) { m.Class = -1 },
		"gender":    func(m *Member) { m.Gender = -1 },
	} {
		t.Run(name, func(t *testing.T) {
			g := recruitmentViewGame(t)
			member := *g.roster[0]
			g.roster = append(g.roster, &member)
			g.recruit.cursor = 1
			change(g.roster[1])
			before := *g.roster[1]
			rng := g.prng
			g.recruitInput(InputState{DirEdge: -1, Enter: true})
			if g.recruit.stage != rcView || g.recruit.cursor != 1 || g.recruit.viewFlow != nil || g.prng != rng || !reflect.DeepEqual(before, *g.roster[1]) {
				t.Fatal("unknown state entered a guessed page or changed member")
			}
		})
	}
}

func TestRecruitmentViewDetailLoadClearsOnlyValidatedUI(t *testing.T) {
	for _, secondWait := range []bool{false, true} {
		t.Run(map[bool]string{false: "ability", true: "close"}[secondWait], func(t *testing.T) {
			g := recruitmentViewGame(t)
			t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "view.json"))
			g.recruitInput(InputState{DirEdge: -1, Enter: true})
			if secondWait {
				g.recruitInput(InputState{DirEdge: -1, Enter: true})
			}
			if g.recruit.viewFlow == nil {
				t.Fatal("pending ability missing")
			}
			if e := g.Save(); e != nil {
				t.Fatal(e)
			}
			valid, e := os.ReadFile(savePath())
			if e != nil {
				t.Fatal(e)
			}
			flow, stage := g.recruit.viewFlow, g.recruit.stage
			rng := g.prng
			bad := g.snapshot()
			bad.PackSchema = "invalid"
			raw, e := encodeSave(bad)
			if e != nil {
				t.Fatal(e)
			}
			if e = os.WriteFile(savePath(), raw, 0644); e != nil {
				t.Fatal(e)
			}
			if e = g.Load(); e == nil || g.recruit.viewFlow != flow || g.recruit.stage != stage {
				t.Fatal("rejected Load changed the pending wait")
			}
			if e = os.WriteFile(savePath(), valid, 0644); e != nil {
				t.Fatal(e)
			}
			if e = g.Load(); e != nil {
				t.Fatal(e)
			}
			expected, e := decodeSave(valid)
			if e != nil {
				t.Fatal(e)
			}
			loaded := g.snapshot()
			// 標題fixture的未建角主角會沿既有Load正規化；正常205包另驗完整snapshot。
			if !reflect.DeepEqual(expected.Roster, loaded.Roster) || !reflect.DeepEqual(expected.Comps, loaded.Comps) ||
				expected.HeroGold != loaded.HeroGold || !reflect.DeepEqual(expected.StoryBits, loaded.StoryBits) || g.prng != rng ||
				g.recruit.active || g.recruit.viewFlow != nil {
				t.Fatal("valid Load failed to preserve member and clear UI")
			}
		})
	}
}
