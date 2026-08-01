package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

var eginbearPushSolution = []int{
	0, 0, 2, 2, 1, 1, 1, 1, 0, 2, 2, 1, 1, 3, 3, 3, 0, 2, 0, 0,
	0, 0, 3, 3, 3, 3, 1, 1, 1, 3, 1, 1, 2, 0, 2, 2, 1, 1, 1, 1,
	0, 0, 0, 0, 2, 1, 2, 2, 2, 0, 0, 3, 3, 1, 0, 2, 2, 1, 1, 3,
	3, 3, 0, 2, 0, 0, 0, 0, 3, 3, 3, 3, 1, 1, 1, 3, 1, 2, 2, 2,
	1, 1, 1, 2, 1, 3, 0, 0, 0, 0, 2, 1, 2, 2, 2, 0, 0, 3, 3, 1,
	0, 2, 2, 1, 1, 3, 3, 3, 0, 3, 1, 1, 1, 3, 1, 2, 0, 2, 1,
}

func eginbearPushPuzzleGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.enterTownCty(76)
	if g.cur == nil || g.cur.sec != 0 || len(g.cur.specialHandlers) != 1 ||
		g.cur.specialHandlers[0] != 30 || len(g.cur.npcs) != 3 {
		t.Fatalf("CTY76 原始 puzzle scene 未解出：handlers=%v npcs=%+v",
			g.cur.specialHandlers, g.cur.npcs)
	}
	for i, want := range [][2]int{{3, 10}, {5, 10}, {7, 10}} {
		n := g.cur.npcs[i]
		if n.x != want[0] || n.y != want[1] || n.ctrl != 0xc0 {
			t.Fatalf("CTY76 NPC slot%d 原始 anchor 錯：%+v", i, n)
		}
	}
	return g
}

func TestPushPuzzleNPCUsesCtrlBitAndOpenDestination(t *testing.T) {
	g := eginbearPushPuzzleGame(t)
	g.px, g.py, g.cd = 3, 11, 0
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 3 || g.py != 10 || g.cur.npcs[0].x != 3 || g.cur.npcs[0].y != 9 {
		t.Fatalf("原始 ctrl bit0x40 推動 transaction 錯：player=(%d,%d) npc=%+v",
			g.px, g.py, g.cur.npcs[0])
	}
}

func TestNPCPushRuleAppliesOutsideDeclaredPuzzle(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.enterTownCty(1)
	if g.cur == nil || g.cur.sec != 0 || len(g.cur.npcs) <= 6 {
		t.Fatalf("CTY01 section0 NPC table 未解出：%+v", g.cur)
	}
	n := &g.cur.npcs[6]
	if n.ctrl&0x40 == 0 {
		t.Fatalf("CTY01 slot6 應保留原始 ctrl bit0x40：%+v", n)
	}
	directions := [][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range directions {
		px, py := n.x-d[0], n.y-d[1]
		tx, ty := n.x+d[0], n.y+d[1]
		if px < 0 || py < 0 || px >= g.cur.w || py >= g.cur.h ||
			tx < 0 || ty < 0 || tx >= g.cur.w || ty >= g.cur.h ||
			g.cur.attr.Blocked(g.cur.tileIdx(px, py)) ||
			g.cur.attr.Blocked(g.cur.tileIdx(tx, ty)) ||
			g.cur.npcAt(px, py) >= 0 || g.cur.npcAt(tx, ty) >= 0 {
			continue
		}
		g.px, g.py = px, py
		if !g.tryPushableNPC(n.x, n.y) {
			t.Fatal("pack 的全域 ctrl 規則未推動 puzzle event 未列出的 NPC")
		}
		if n.x != tx || n.y != ty {
			t.Fatalf("CTY01 slot6 推動終點錯：got=(%d,%d) want=(%d,%d)", n.x, n.y, tx, ty)
		}
		return
	}
	t.Fatal("CTY01 slot6 周圍找不到可驗證的原始開放推動方向")
}

func TestPushPuzzleInvisibilityBlocksPush(t *testing.T) {
	g := eginbearPushPuzzleGame(t)
	g.px, g.py, g.cd, g.remoaru = 3, 11, 0, 25
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 3 || g.py != 11 || g.cur.npcs[0].x != 3 || g.cur.npcs[0].y != 10 ||
		g.remoaru != 25 {
		t.Fatalf("隱形時不得推 NPC：player=(%d,%d) npc=%+v timer=%d",
			g.px, g.py, g.cur.npcs[0], g.remoaru)
	}
}

func TestPushPuzzleFinalStepClearsFlagAndRevealsPassage(t *testing.T) {
	g := eginbearPushPuzzleGame(t)
	g.cur.npcs[0].x, g.cur.npcs[0].y = 4, 5
	g.cur.npcs[1].x, g.cur.npcs[1].y = 5, 5
	g.cur.npcs[2].x, g.cur.npcs[2].y = 6, 6
	g.px, g.py, g.cd = 6, 7, 0
	if !g.storyFlag(0x3b) {
		t.Fatal("原始初值應令 CTY76 passage flag0x3b 為 set")
	}
	beforeGate := g.cur.tileIdx(7, 5)
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 6 || g.py != 6 || g.cur.npcs[2].y != 5 {
		t.Fatalf("最後一推未完成：player=(%d,%d) npcs=%+v", g.px, g.py, g.cur.npcs)
	}
	if g.storyFlag(0x3b) {
		t.Fatal("三個原始 slot 的 Y 全為5後應 clear flag0x3b")
	}
	if got := g.cur.tileIdx(7, 5); got != (beforeGate+1)&0xff || g.cur.hiMap[5*g.cur.w+7]&0x1f != 0 {
		t.Fatalf("完成後原始事件 tile 未刷新：before=%d after=%d hi=%02x",
			beforeGate, got, g.cur.hiMap[5*g.cur.w+7])
	}
}

func traceEginbearPushPuzzle(t *testing.T, g *Game) {
	t.Helper()
	for stepIndex, dir := range eginbearPushSolution {
		for g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatalf("推石 cooldown step%d：%v", stepIndex, err)
			}
		}
		px, py := g.px, g.py
		if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
			t.Fatalf("推石 production input step%d：%v", stepIndex, err)
		}
		if g.px == px && g.py == py {
			t.Fatalf("推石解路 step%d dir%d 未移動：player=(%d,%d) npcs=%+v",
				stepIndex, dir, g.px, g.py, g.cur.npcs)
		}
	}
}

func TestEginbearPushPuzzleProductionRouteAndTreasure(t *testing.T) {
	g := eginbearPushPuzzleGame(t)
	g.px, g.py, g.remoaru = 5, 12, 0
	traceEginbearPushPuzzle(t, g)
	if g.px != 4 || g.py != 6 || g.storyFlag(0x3b) || g.remoaru != 0 {
		t.Fatalf("正式推石解路終態錯：player=(%d,%d) flag3b=%v timer=%d",
			g.px, g.py, g.storyFlag(0x3b), g.remoaru)
	}
	for i := range g.cur.npcs {
		if g.cur.npcs[i].y != 5 {
			t.Fatalf("slot%d 未推至原始完成列：%+v", i, g.cur.npcs[i])
		}
	}
	if treasureFor(76, 0, 0) != nil || treasureFor(76, 0, 1) != nil {
		t.Fatal("CTY76 乾渴壺不得殘留在 legacy Go treasure table")
	}
	traceExaminePackTreasure(t, g, gamepack.QuestTreasureSelector{
		CTYRaw: 76, Section: 0, TileSubID: 0, EventTypeRaw: 1,
		ItemRawID: 0x5e, PresentFlag: 0x3a,
	})
	if !g.hasItem(0x5e) || g.storyFlag(0x3a) {
		t.Fatalf("正式調查未取得乾渴壺：item=%v flag3a=%v",
			g.hasItem(0x5e), g.storyFlag(0x3a))
	}
}

func TestDumpEginbearPushPuzzle(t *testing.T) {
	if os.Getenv("DQ3_DUMP_PUSH") == "" {
		t.Skip("設 DQ3_DUMP_PUSH=1 才執行")
	}
	out := os.Getenv("PUSH_OUT")
	if out == "" {
		t.Fatal("需設 PUSH_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dump := func(g *Game, name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	g := eginbearPushPuzzleGame(t)
	g.px, g.py = 5, 12
	dump(g, "eginbear_push_puzzle_before")
	traceEginbearPushPuzzle(t, g)
	dump(g, "eginbear_push_puzzle_solved")
}
