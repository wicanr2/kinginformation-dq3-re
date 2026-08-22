package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// Raw records are test evidence only. Production dispatch reads them from the
// selected game pack and must not gain a DQ3-specific Go fallback.
const (
	fieldRura     = 172
	fieldRemitto  = 173
	fieldInpas    = 174
	fieldToramana = 175
	fieldToheros  = 176
	fieldRanaruta = 177
	fieldRemoaru  = 178
	fieldAbakamu  = 179
)

func TestPersistentParalysisClearsWholePartyAfterFortySteps(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{
		pack: pack, heroConditions: conditionParalysis,
		companions:     []*Member{{CurHP: 10, Conditions: conditionParalysis}},
		paralysisSteps: 40,
	}
	for step := 1; step < 40; step++ {
		g.advancePersistentConditionStep()
		if !g.partyHasCondition(conditionParalysis) || g.paralysisSteps != 40-step {
			t.Fatalf("第 %d 步提前解除：conditions=%#x/%#x counter=%d",
				step, g.heroConditions, g.companions[0].Conditions, g.paralysisSteps)
		}
	}
	g.advancePersistentConditionStep()
	if g.partyHasCondition(conditionParalysis) || g.paralysisSteps != 0 {
		t.Fatalf("第 40 步未全隊解除：conditions=%#x/%#x counter=%d",
			g.heroConditions, g.companions[0].Conditions, g.paralysisSteps)
	}
}

func fieldSpellGame(t *testing.T, level int) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.heroExp = stats.ExpForLevel(0, level)
	g.ensureHeroStats()
	_, maxHP, _, _, _ := g.heroStats()
	g.heroHP = maxHP
	g.heroMP = 99
	return g
}

func openSpellFromProductionInput(t *testing.T, g *Game) {
	t.Helper()
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 開命令窗
	testStep(t, g, InputState{DirHeld: -1, DirEdge: 3})                 // 對話→咒文
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.active {
		t.Fatal("正式命令窗選「咒文」後應開啟野外咒文 modal")
	}
	if !g.fieldSpell.selectingCaster {
		t.Fatal("原版流程應先進入施法者選擇頁")
	}
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 選主角
}

func openFieldSpellForRecord(t *testing.T, g *Game, rec int) {
	t.Helper()
	g.openFieldSpellMenu()
	caster := g.fieldSpellCaster(rec)
	if caster < 0 {
		t.Fatalf("active party 無人可施放 pack field spell rec%d", rec)
	}
	g.fieldSpell.cursor = caster
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.selectingCaster || fieldSpellChoiceIndex(g, rec) < 0 {
		t.Fatalf("選施法者%d後未建立 rec%d 清單：%+v", caster, rec, g.fieldSpell.choices)
	}
}

func TestRuraProductionInputVisitedTownAndMP(t *testing.T) {
	g := fieldSpellGame(t, 7) // 勇者 Lv7 習得魯拉
	g.enterTownCty(0)
	g.exitTown()
	g.enterTownCty(2)
	g.exitTown()
	if len(g.visitedTowns) != 2 {
		t.Fatalf("已造訪阿里阿罕／羅馬利亞應記兩城，得 %+v", g.visitedTowns)
	}
	mp0 := g.heroMP
	openSpellFromProductionInput(t, g)
	ruraIndex := fieldSpellChoiceIndex(g, fieldRura)
	if ruraIndex < 0 {
		t.Fatalf("Lv7 野外清單應包含魯拉，得 %+v", g.fieldSpell.choices)
	}
	for g.fieldSpell.cursor != ruraIndex {
		testStep(t, g, InputState{DirHeld: -1, DirEdge: 0})
	}
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 魯拉→目的地
	if !g.fieldSpell.dest || g.heroMP != mp0-8 {
		t.Fatalf("原版 caster 應在進目的地 handler 前扣 8 MP：dest=%v mp=%d", g.fieldSpell.dest, g.heroMP)
	}
	testStep(t, g, InputState{DirHeld: -1, DirEdge: 0}) // 阿里阿罕→羅馬利亞
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.inTown || g.layer != 0 || g.px != ctyLoc[2][0]-1 || g.py != ctyLoc[2][1] ||
		g.fieldSpell.active {
		t.Fatalf("魯拉應依原版停在 CTY2 西側地表：town=%v layer=%d @(%d,%d) modal=%v",
			g.inTown, g.layer, g.px, g.py, g.fieldSpell.active)
	}
	if g.heroMP != mp0-8 {
		t.Fatalf("魯拉成功後應扣 8 MP：got %d want %d", g.heroMP, mp0-8)
	}
}

func TestRemittoProductionInputDungeonOnly(t *testing.T) {
	g := fieldSpellGame(t, 14) // 勇者已習得魯拉、烈米特
	g.enterTownCty(8)          // 拿吉米之塔，BLK2=dungeon
	g.overPx, g.overPy = 44, 55
	mp0 := g.heroMP
	openSpellFromProductionInput(t, g)
	remittoIndex := fieldSpellChoiceIndex(g, fieldRemitto)
	if remittoIndex < 0 {
		t.Fatalf("Lv14 野外清單應包含烈米特，得 %+v", g.fieldSpell.choices)
	}
	for g.fieldSpell.cursor != remittoIndex {
		testStep(t, g, InputState{DirHeld: -1, DirEdge: 0})
	}
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.inTown || g.px != 44 || g.py != 55 || g.fieldSpell.active {
		t.Fatalf("迷宮內烈米特應回入口地表：town=%v @(%d,%d) modal=%v",
			g.inTown, g.px, g.py, g.fieldSpell.active)
	}
	if g.heroMP != mp0-8 {
		t.Fatalf("烈米特成功後應依 EXE 扣 8 MP：got %d want %d", g.heroMP, mp0-8)
	}
}

func TestVisitedTownsSaveRoundTrip(t *testing.T) {
	want := []townVisit{{Cty: 0, Section: 0, X: 0, Y: 28}, {Cty: 37, Section: 0, X: 4, Y: 19}}
	s := saveState{VisitedTowns: want}
	b, err := encodeSave(s)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeSave(b)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got.VisitedTowns, want) {
		t.Fatalf("魯拉目的地存檔 round-trip 不一致：got %+v want %+v", got.VisitedTowns, want)
	}
}

func TestRuraCancelKeepsPreHandlerMPSpend(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.visitedTowns = []townVisit{{Cty: 0}}
	openFieldSpellForRecord(t, g, fieldRura)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Cancel: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0-8 || g.fieldSpell.active || g.fieldSpell.dest {
		t.Fatalf("取消目的地仍應耗 8 MP 並結束施法：mp=%d active=%v dest=%v",
			g.heroMP, g.fieldSpell.active, g.fieldSpell.dest)
	}

	g.heroMP = 7
	openFieldSpellForRecord(t, g, fieldRura)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.dest || g.heroMP != 7 {
		t.Fatalf("MP 不足不可進目的地或扣 MP：mp=%d dest=%v", g.heroMP, g.fieldSpell.dest)
	}
}

func TestRemittoOutsideDungeonStillSpends(t *testing.T) {
	g := fieldSpellGame(t, 14)
	g.inTown, g.curCty = false, -1
	openFieldSpellForRecord(t, g, fieldRemitto)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRemitto)
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0-8 || g.fieldSpell.active {
		t.Fatalf("迷宮外烈米特被 handler 拒絕，但原版 caster 已扣 MP：mp=%d active=%v",
			g.heroMP, g.fieldSpell.active)
	}
}

func fieldSpellChoiceIndex(g *Game, rec int) int {
	for i, c := range g.fieldSpell.choices {
		if c.rec == rec {
			return i
		}
	}
	return -1
}

func TestFieldHealCompanionCasterAndTargetUsesOriginalTransaction(t *testing.T) {
	g := fieldSpellGame(t, 1)
	priest := newMember([]int{1}, 3, 0, stats.ExpForLevel(3, 14))
	target := newMember([]int{2}, 1, 0, stats.ExpForLevel(1, 30))
	priest.CurMP, target.CurHP = 99, 1
	g.companions = []*Member{priest, target}

	g.openFieldSpellMenu()
	if !g.fieldSpell.selectingCaster {
		t.Fatal("野外咒文必須先選施法者")
	}
	g.fieldSpell.cursor = 1
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	idx := fieldSpellChoiceIndex(g, 162)
	if idx < 0 || containsRec(g.fieldActorSpells(0), 162) {
		t.Fatalf("僧侶清單應含 rec162，且不得混入主角未習得咒文：%+v", g.fieldSpell.choices)
	}
	g.fieldSpell.cursor = idx
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.targeting || priest.CurMP != 99 {
		t.Fatalf("單體回復應先選目標且尚未扣 MP：targeting=%v mp=%d",
			g.fieldSpell.targeting, priest.CurMP)
	}
	g.fieldSpell.cursor = 2
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	healed := target.CurHP - 1
	if healed < 85 || healed > 94 || priest.CurMP != 94 || g.fieldSpell.active {
		t.Fatalf("rec162 應回復85..94並扣5 MP：heal=%d mp=%d active=%v",
			healed, priest.CurMP, g.fieldSpell.active)
	}

	b, err := encodeSave(g.snapshot())
	if err != nil {
		t.Fatal(err)
	}
	saved, err := decodeSave(b)
	if err != nil {
		t.Fatal(err)
	}
	if saved.Comps[0].CurMP != priest.CurMP || saved.Comps[1].CurHP != target.CurHP {
		t.Fatalf("施法後 HP/MP 存檔 round-trip 漂移：got mp=%d hp=%d want mp=%d hp=%d",
			saved.Comps[0].CurMP, saved.Comps[1].CurHP, priest.CurMP, target.CurHP)
	}
}

func TestFieldHealTargetCancelDoesNotSpendMP(t *testing.T) {
	g := fieldSpellGame(t, 4)
	openFieldSpellForRecord(t, g, 161)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, 161)
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.fieldSpell.targeting {
		t.Fatal("rec161 應進單體目標頁")
	}
	g.fieldSpellInput(InputState{Cancel: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0 || g.fieldSpell.active {
		t.Fatalf("取消單體目標不應扣 MP 且應結束 modal：mp=%d active=%v", g.heroMP, g.fieldSpell.active)
	}
}

func TestFieldHealFullAndDeadTargetsStillSpendAfterConfirm(t *testing.T) {
	tests := []struct {
		name string
		dead bool
	}{
		{"full", false},
		{"dead", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := fieldSpellGame(t, 4)
			target := newMember([]int{2}, 1, 0, stats.ExpForLevel(1, 20))
			if tt.dead {
				target.CurHP = 0
			}
			g.companions = []*Member{target}
			openFieldSpellForRecord(t, g, 161)
			g.fieldSpell.cursor = fieldSpellChoiceIndex(g, 161)
			mp0, hp0 := g.heroMP, target.CurHP
			g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
			g.fieldSpell.cursor = 1
			g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
			if g.heroMP != mp0-3 || target.CurHP != hp0 {
				t.Fatalf("%s target 應維持 HP 但已扣3 MP：hp=%d/%d mp=%d/%d",
					tt.name, target.CurHP, hp0, g.heroMP, mp0)
			}
		})
	}
}

func TestFieldHealGroupUsesIndependentLivingMemberWriters(t *testing.T) {
	g := fieldSpellGame(t, 40)
	priest := newMember([]int{1}, 3, 0, stats.ExpForLevel(3, 34))
	dead := newMember([]int{2}, 1, 0, stats.ExpForLevel(1, 30))
	living := newMember([]int{3}, 4, 0, stats.ExpForLevel(4, 30))
	priest.CurMP, priest.CurHP, dead.CurHP, living.CurHP = 99, 1, 0, 1
	g.heroHP = 1
	g.companions = []*Member{priest, dead, living}
	openFieldSpellForRecord(t, g, 164)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, 164)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	for name, hp := range map[string]int{"hero": g.heroHP, "priest": priest.CurHP, "living": living.CurHP} {
		if hp < 51 || hp > 60 {
			t.Fatalf("rec164 %s 應由1回復至51..60，得%d", name, hp)
		}
	}
	if dead.CurHP != 0 || priest.CurMP != 81 || g.fieldSpell.active {
		t.Fatalf("rec164 應略過死人且只扣一次18 MP：dead=%d mp=%d active=%v",
			dead.CurHP, priest.CurMP, g.fieldSpell.active)
	}
}

func TestFieldFullHealGroupUsesPackRec165(t *testing.T) {
	g := fieldSpellGame(t, 41)
	living := newMember([]int{1}, 1, 0, stats.ExpForLevel(1, 30))
	dead := newMember([]int{2}, 1, 0, stats.ExpForLevel(1, 30))
	g.heroHP, living.CurHP, dead.CurHP = 1, 1, 0
	g.companions = []*Member{living, dead}
	openFieldSpellForRecord(t, g, 165)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, 165)
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	_, heroMax, _, _, _ := g.heroStats()
	if g.heroHP != heroMax || living.CurHP != living.MaxHP() || dead.CurHP != 0 || g.heroMP != mp0-62 {
		t.Fatalf("rec165 應補滿活人、略過死人並扣62 MP：hero=%d/%d living=%d/%d dead=%d mp=%d/%d",
			g.heroHP, heroMax, living.CurHP, living.MaxHP(), dead.CurHP, g.heroMP, mp0)
	}
}

func placeFacingEventType(t *testing.T, g *Game, cty, sec, wantType int) {
	t.Helper()
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		cty, mapBlkNum[cty], sec, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < sc.h; y++ {
		for x := 0; x < sc.w; x++ {
			ev, _, ok := sc.tileEvent(x, y)
			if !ok || ev[0] != wantType {
				continue
			}
			for d := 0; d < 4; d++ {
				dx, dy := dirDelta(d)
				px, py := x-dx, y-dy
				if px < 0 || py < 0 || px >= sc.w || py >= sc.h || sc.Blocked(px, py) {
					continue
				}
				g.cur, g.town, g.inTown, g.curCty = sc, sc, true, cty
				g.px, g.py, g.facing = px, py, d
				return
			}
		}
	}
	t.Fatalf("CTY%02d sec%d 找不到可面向的 type%d 事件", cty, sec, wantType)
}

func TestInpasOriginalEventTypeRecordsAndMP(t *testing.T) {
	tests := []struct {
		name          string
		cty, sec, typ int
		wantRec       int
	}{
		{"ordinary treasure", 0, 5, 1, 0xfd},
		{"type2 warp", 14, 0, 2, 0xfe},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := fieldSpellGame(t, 1)
			mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 18))
			mage.CurMP = 99
			g.companions = []*Member{mage}
			placeFacingEventType(t, g, tt.cty, tt.sec, tt.typ)
			openFieldSpellForRecord(t, g, fieldInpas)
			idx := fieldSpellChoiceIndex(g, fieldInpas)
			if idx < 0 {
				t.Fatalf("Lv18 魔法使的場景咒文清單缺 rec174：%+v", g.fieldSpell.choices)
			}
			g.fieldSpell.cursor = idx
			g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
			if mage.CurMP != 96 || !g.dlg.open ||
				!reflect.DeepEqual(g.dlg.buf, g.shop.nameText.Record(tt.wantRec)) {
				t.Fatalf("因帕斯 type%d 應扣3 MP並顯示 rec%d：mp=%d open=%v",
					tt.typ, tt.wantRec, mage.CurMP, g.dlg.open)
			}
		})
	}
}

func TestInpasInvalidSceneStillSpendsAndShowsFailure(t *testing.T) {
	g := fieldSpellGame(t, 1)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 18))
	mage.CurMP = 99
	g.companions = []*Member{mage}
	g.inTown, g.curCty = false, -1
	openFieldSpellForRecord(t, g, fieldInpas)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldInpas)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if mage.CurMP != 96 || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, g.shop.nameText.Record(0x15a)) {
		t.Fatalf("因帕斯無事件仍應先扣3 MP再顯示 rec346：mp=%d open=%v",
			mage.CurMP, g.dlg.open)
	}
}

func TestToramanaOriginalCostAndOneHazardRegionContract(t *testing.T) {
	g := fieldSpellGame(t, 1)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 19))
	mage.CurMP, mage.CurHP = 99, 50
	g.companions = []*Member{mage}
	openFieldSpellForRecord(t, g, fieldToramana)
	idx := fieldSpellChoiceIndex(g, fieldToramana)
	if idx < 0 {
		t.Fatalf("Lv19 魔法使的場景咒文清單缺 rec175：%+v", g.fieldSpell.choices)
	}
	g.fieldSpell.cursor = idx
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if mage.CurMP != 97 || !g.toramana || g.hazardGuard || g.fieldSpell.active {
		t.Fatalf("多拉瑪那應扣2 MP並設原版0x2000：mp=%d toramana=%v guard=%v active=%v",
			mage.CurMP, g.toramana, g.hazardGuard, g.fieldSpell.active)
	}

	attr := &dq3data.BlockAttr{A: []uint16{0x0400}} // AH&4 → 原版 10 damage 類
	g.cur = &Scene{w: 1, h: 1, tileAt: func(_, _ int) int { return 0 }, attr: attr}
	g.px, g.py, g.heroHP, mage.CurHP = 0, 0, 50, 50
	g.applyHazardStep()
	if g.heroHP != 50 || mage.CurHP != 50 || g.toramana || !g.hazardGuard {
		t.Fatalf("第一次高傷地板應把0x2000轉0x0400且免傷：hp=%d/%d toramana=%v guard=%v",
			g.heroHP, mage.CurHP, g.toramana, g.hazardGuard)
	}
	g.applyHazardStep()
	if g.heroHP != 50 || mage.CurHP != 50 {
		t.Fatalf("同一段連續傷害地板應持續免傷：hp=%d/%d", g.heroHP, mage.CurHP)
	}
	attr.A[0] = 0
	g.applyHazardStep()
	if g.hazardGuard {
		t.Fatal("離開傷害地板後原版0x0400 guard應清除")
	}
	attr.A[0] = 0x0400
	g.applyHazardStep()
	if g.heroHP != 40 || mage.CurHP != 40 {
		t.Fatalf("重新踏入高傷地板應全隊各受10傷：hp=%d/%d", g.heroHP, mage.CurHP)
	}

	g.heroHP, mage.CurHP, g.toramana = 50, 50, true
	attr.A[0] = 0x0100 // AH&1 分支早於多拉瑪那檢查
	g.applyHazardStep()
	if g.heroHP != 49 || mage.CurHP != 49 || !g.toramana {
		t.Fatalf("AH&1 類仍應受1傷且不消耗0x2000：hp=%d/%d toramana=%v",
			g.heroHP, mage.CurHP, g.toramana)
	}
}

func TestFieldSpellOriginalMapFlagsAndCosts(t *testing.T) {
	g := fieldSpellGame(t, 25) // 四種野外咒文皆習得
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 25))
	mage.CurMP = 99
	g.companions = []*Member{mage}
	g.enterTownCty(8) // CTY08 sec0 header+0x10 == 2：烈米特可、魯拉不可
	openFieldSpellForRecord(t, g, fieldRura)
	mp0 := g.heroMP
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.dest || g.heroMP != mp0-8 || g.fieldSpell.active {
		t.Fatalf("mapFlags bit0 清時魯拉 handler 拒絕，但 caster 已扣 MP：dest=%v mp=%d active=%v",
			g.fieldSpell.dest, g.heroMP, g.fieldSpell.active)
	}

	g.exitTown()
	openFieldSpellForRecord(t, g, fieldToheros)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldToheros)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.repel != 0x28 || g.heroMP != mp0-12 {
		t.Fatalf("特黑洛斯應依 handler 設 40 步並扣 4 MP：repel=%d mp=%d", g.repel, g.heroMP)
	}

	openFieldSpellForRecord(t, g, fieldRanaruta)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRanaruta)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.isNight() || g.heroMP != mp0-12 || mage.CurMP != 87 {
		t.Fatalf("地表拉那魯達應由魔法使扣 12 MP：night=%v heroMP=%d mageMP=%d",
			g.isNight(), g.heroMP, mage.CurMP)
	}
}

func TestRanarutaInTownStillSpends(t *testing.T) {
	g := fieldSpellGame(t, 25)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 25))
	mage.CurMP = 99
	g.companions = []*Member{mage}
	g.enterTownCty(0)
	openFieldSpellForRecord(t, g, fieldRanaruta)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRanaruta)
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0 || mage.CurMP != 87 || g.fieldSpell.active {
		t.Fatalf("CTY 內拉那魯達被 handler 拒絕，但原版 caster 已扣 MP：heroMP=%d mageMP=%d active=%v",
			g.heroMP, mage.CurMP, g.fieldSpell.active)
	}
}

func TestRemoaruOriginalCostTimerAndExpiry(t *testing.T) {
	g := fieldSpellGame(t, 1)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 33))
	mage.CurMP = 99
	g.companions = []*Member{mage}
	openFieldSpellForRecord(t, g, fieldRemoaru)
	idx := -1
	for i, c := range g.fieldSpell.choices {
		if c.rec == fieldRemoaru {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("Lv33 魔法使的場景咒文清單缺 rec178：%+v", g.fieldSpell.choices)
	}
	g.fieldSpell.cursor = idx
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.remoaru != 0x19 || mage.CurMP != 84 || g.fieldSpell.active {
		t.Fatalf("rec178 應扣 15 MP 並設 25 步透明：timer=%d mp=%d active=%v",
			g.remoaru, mage.CurMP, g.fieldSpell.active)
	}
	for i := 0; i < 24; i++ {
		g.advanceFieldSpellStep()
	}
	if g.remoaru != 1 {
		t.Fatalf("24 個有效步後 timer 應剩 1，得 %d", g.remoaru)
	}
	g.advanceFieldSpellStep()
	if g.remoaru != 0 {
		t.Fatalf("第 25 步應解除透明，得 timer=%d", g.remoaru)
	}
}

func TestRemoaruSuccessfulMovementAndRendererConsumeState(t *testing.T) {
	g := fieldSpellGame(t, 1)
	g.renderFrame()
	visible := append([]byte(nil), g.rgba...)
	g.remoaru = 25
	g.renderFrame()
	if reflect.DeepEqual(visible, g.rgba) {
		t.Fatal("透明狀態必須改變 production renderer，而非只保存 timer")
	}

	dir := -1
	for d := 0; d < 4; d++ {
		dx, dy := dirDelta(d)
		if !g.cur.Blocked(g.px+dx, g.py+dy) {
			dir = d
			break
		}
	}
	if dir < 0 {
		t.Fatal("測試起點周圍沒有可走格")
	}
	x0, y0 := g.px, g.py
	g.cd = 0
	testStep(t, g, InputState{DirHeld: dir, DirEdge: -1})
	if g.px == x0 && g.py == y0 {
		t.Fatal("production movement input 沒有移動")
	}
	if g.remoaru != 24 {
		t.Fatalf("一個成功 production movement 應使 timer 25→24，得 %d", g.remoaru)
	}
}

func TestRemoaruIsTransientAcrossRestore(t *testing.T) {
	g := fieldSpellGame(t, 1)
	s := g.snapshot()
	g.remoaru = 25
	g.restore(s)
	if g.remoaru != 0 {
		t.Fatalf("冒險之書讀檔不可保留 runtime 透明 timer：%d", g.remoaru)
	}
}

func TestAbakamuOpensAnyFacingDoorAtZeroMP(t *testing.T) {
	g := fieldSpellGame(t, 1)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 35))
	mage.CurMP = 0 // EXE descriptor rec179 的 MP cost 就是 0。
	g.companions = []*Member{mage}

	found := false
	for sec := 0; sec < 12 && !found; sec++ {
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], sec, 0, g.storyFlag)
		if err != nil {
			continue
		}
		for y := 0; y < sc.h && !found; y++ {
			for x := 0; x < sc.w && !found; x++ {
				if sc.doorTier(x, y) == 0 {
					continue
				}
				for d := 0; d < 4; d++ {
					dx, dy := dirDelta(d)
					px, py := x-dx, y-dy
					if px < 0 || py < 0 || px >= sc.w || py >= sc.h || sc.Blocked(px, py) {
						continue
					}
					g.cur, g.town, g.inTown, g.curCty = sc, sc, true, 0
					g.px, g.py, g.facing = px, py, d
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Fatal("CTY00 找不到可由相鄰可走格面向的門")
	}

	openFieldSpellForRecord(t, g, fieldAbakamu)
	idx := -1
	for i, c := range g.fieldSpell.choices {
		if c.rec == fieldAbakamu {
			idx = i
			break
		}
	}
	if idx < 0 {
		t.Fatalf("Lv35 魔法使的場景咒文清單缺 rec179：%+v", g.fieldSpell.choices)
	}
	dx, dy := dirDelta(g.facing)
	doorX, doorY := g.px+dx, g.py+dy
	g.fieldSpell.cursor = idx
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.cur.doorTier(doorX, doorY) != 0 || mage.CurMP != 0 || g.fieldSpell.active {
		t.Fatalf("阿巴卡姆應以 0 MP 開啟任意面向門：tier=%d mp=%d active=%v",
			g.cur.doorTier(doorX, doorY), mage.CurMP, g.fieldSpell.active)
	}
}

func TestAbakamuFailureUsesOriginalGlobalMessage(t *testing.T) {
	g := fieldSpellGame(t, 1)
	mage := newMember(nil, 4, 0, stats.ExpForLevel(4, 35))
	mage.CurMP = 0
	g.companions = []*Member{mage}
	g.inTown, g.curCty = false, -1
	openFieldSpellForRecord(t, g, fieldAbakamu)
	for i, c := range g.fieldSpell.choices {
		if c.rec == fieldAbakamu {
			g.fieldSpell.cursor = i
			break
		}
	}
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, g.shop.nameText.Record(0x15a)) {
		t.Fatal("阿巴卡姆非 CTY／非門失敗應顯示 D3TXT00 rec0x15a")
	}
}

func TestRuraSpecialDestination47Coordinate(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.visitedTowns = []townVisit{{Cty: 47}}
	openFieldSpellForRecord(t, g, fieldRura)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.px != ctyLoc[47][0]+1 || g.py != ctyLoc[47][1] {
		t.Fatalf("CTY47 原版特例應為 X+1,Y：got (%d,%d)", g.px, g.py)
	}
}

func TestRuraRelocatesOwnedShipFromPack(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.shipOwned = true
	g.shipX, g.shipY = 1, 1
	g.visitedTowns = []townVisit{{Cty: 15}}
	openFieldSpellForRecord(t, g, fieldRura)
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.px != ctyLoc[15][0]-1 || g.py != ctyLoc[15][1] ||
		g.shipX != 104 || g.shipY != 126 {
		t.Fatalf("巴哈拉達魯拉應同步重定位船：player=(%d,%d) ship=(%d,%d)",
			g.px, g.py, g.shipX, g.shipY)
	}
}

func TestVisitedTownsKeepOriginalRuraOrder(t *testing.T) {
	g := fieldSpellGame(t, 7)
	for _, cty := range []int{47, 0, 16, 2} {
		g.curCty, g.inTown, g.cur = cty, true, g.town
		g.rememberTown()
	}
	want := []int{0, 2, 16, 47}
	for i, cty := range want {
		if g.visitedTowns[i].Cty != cty {
			t.Fatalf("魯拉目的地須依 EXE table 排序：got %+v want %v", g.visitedTowns, want)
		}
	}
}

func TestPoisonFieldDamageUsesPackValueAndVehicleSuppression(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.cur = nil // 毒傷是 movement-status consumer，不依賴地形 attr。
	g.heroHP, g.heroConditions = 5, conditionPoison
	companion := newMember([]int{1}, 3, 0, 0)
	companion.CurHP, companion.Conditions = 5, conditionPoison
	g.companions = []*Member{companion}
	g.applyHazardStep()
	if g.heroHP != 4 || companion.CurHP != 4 {
		t.Fatalf("poison step damage hero=%d companion=%d, want 4/4", g.heroHP, companion.CurHP)
	}
	g.shipAboard = true
	g.applyHazardStep()
	if g.heroHP != 4 || companion.CurHP != 4 {
		t.Fatalf("ship must suppress poison damage hero=%d companion=%d", g.heroHP, companion.CurHP)
	}
	g.shipAboard, g.phoenixAboard = false, true
	g.applyHazardStep()
	if g.heroHP != 4 || companion.CurHP != 4 {
		t.Fatalf("phoenix must suppress poison damage hero=%d companion=%d", g.heroHP, companion.CurHP)
	}
}

func TestPoisonDamageAppliesOnlyToAffectedMember(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.heroHP, g.heroConditions = 10, conditionPoison
	companion := newMember([]int{1}, 3, 0, 0)
	companion.CurHP = 10
	g.companions = []*Member{companion}
	g.cur = nil
	g.applyHazardStep()
	if g.heroHP != 9 || companion.CurHP != 10 {
		t.Fatalf("poison must apply only to affected member hero=%d companion=%d", g.heroHP, companion.CurHP)
	}
}
