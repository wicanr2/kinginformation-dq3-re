package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

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
	if len(g.fieldSpell.choices) != 1 || g.fieldSpell.choices[0].rec != fieldRura {
		t.Fatalf("Lv7 野外清單應只有魯拉，得 %+v", g.fieldSpell.choices)
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
	if len(g.fieldSpell.choices) < 2 || g.fieldSpell.choices[1].rec != fieldRemitto {
		t.Fatalf("Lv14 第二個野外咒應為烈米特，得 %+v", g.fieldSpell.choices)
	}
	testStep(t, g, InputState{DirHeld: -1, DirEdge: 0})
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
	g.openFieldSpellMenu()
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Cancel: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0-8 || g.fieldSpell.active || g.fieldSpell.dest {
		t.Fatalf("取消目的地仍應耗 8 MP 並結束施法：mp=%d active=%v dest=%v",
			g.heroMP, g.fieldSpell.active, g.fieldSpell.dest)
	}

	g.heroMP = 7
	g.openFieldSpellMenu()
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.dest || g.heroMP != 7 {
		t.Fatalf("MP 不足不可進目的地或扣 MP：mp=%d dest=%v", g.heroMP, g.fieldSpell.dest)
	}
}

func TestRemittoOutsideDungeonStillSpends(t *testing.T) {
	g := fieldSpellGame(t, 14)
	g.inTown, g.curCty = false, -1
	g.openFieldSpellMenu()
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
			g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
	mp0 := g.heroMP
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldRura)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.dest || g.heroMP != mp0-8 || g.fieldSpell.active {
		t.Fatalf("mapFlags bit0 清時魯拉 handler 拒絕，但 caster 已扣 MP：dest=%v mp=%d active=%v",
			g.fieldSpell.dest, g.heroMP, g.fieldSpell.active)
	}

	g.exitTown()
	g.openFieldSpellMenu()
	g.fieldSpell.cursor = fieldSpellChoiceIndex(g, fieldToheros)
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.repel != 0x28 || g.heroMP != mp0-12 {
		t.Fatalf("特黑洛斯應依 handler 設 40 步並扣 4 MP：repel=%d mp=%d", g.repel, g.heroMP)
	}

	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
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

	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
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
	g.openFieldSpellMenu()
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.px != ctyLoc[47][0]+1 || g.py != ctyLoc[47][1] {
		t.Fatalf("CTY47 原版特例應為 X+1,Y：got (%d,%d)", g.px, g.py)
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
