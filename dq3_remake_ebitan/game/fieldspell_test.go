package game

import (
	"os"
	"reflect"
	"testing"

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
	if !g.fieldSpell.dest || g.heroMP != mp0 {
		t.Fatalf("選咒後應先選目的地且尚不扣 MP：dest=%v mp=%d", g.fieldSpell.dest, g.heroMP)
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
	if g.heroMP != mp0-6 {
		t.Fatalf("烈米特成功後應扣 6 MP：got %d want %d", g.heroMP, mp0-6)
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

func TestRuraCancelAndInsufficientMPDoNotSpend(t *testing.T) {
	g := fieldSpellGame(t, 7)
	g.visitedTowns = []townVisit{{Cty: 0}}
	g.openFieldSpellMenu()
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	g.fieldSpellInput(InputState{Cancel: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0 || g.fieldSpell.dest {
		t.Fatalf("取消魯拉目的地不可扣 MP：mp=%d dest=%v", g.heroMP, g.fieldSpell.dest)
	}

	g.heroMP = 7
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.fieldSpell.dest || g.heroMP != 7 {
		t.Fatalf("MP 不足不可進目的地或扣 MP：mp=%d dest=%v", g.heroMP, g.fieldSpell.dest)
	}
}

func TestRemittoOutsideDungeonDoesNotSpend(t *testing.T) {
	g := fieldSpellGame(t, 14)
	g.inTown, g.curCty = false, -1
	g.openFieldSpellMenu()
	g.fieldSpell.cursor = 1
	mp0 := g.heroMP
	g.fieldSpellInput(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.heroMP != mp0 || !g.fieldSpell.active {
		t.Fatalf("迷宮外烈米特不可扣 MP 或關閉選單：mp=%d active=%v", g.heroMP, g.fieldSpell.active)
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
