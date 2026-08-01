package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func eginbearGuardGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.enterTownCty(39)
	if g.cur == nil || g.cur.sec != 0 || len(g.cur.specialHandlers) != 1 ||
		g.cur.specialHandlers[0] != 29 {
		t.Fatalf("CTY39 sec0 原始 special handler table 未解出：%v", g.cur.specialHandlers)
	}
	return g
}

func TestEginbearGuardTracksVisiblePlayerAndBlocks(t *testing.T) {
	g := eginbearGuardGame(t)
	g.px, g.py, g.cd = 14, 38, 0
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	idx := g.cur.npcAt(14, 36)
	if g.px != 14 || g.py != 37 || idx < 0 || g.cur.npcs[idx].b4 != 15 {
		t.Fatalf("可見玩家踏 gate 後守衛未橫移攔截：player=(%d,%d) guard=%d npcs=%+v",
			g.px, g.py, idx, g.cur.npcs)
	}
	g.cd = 0
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 14 || g.py != 37 {
		t.Fatalf("可見玩家不應穿過守衛：(%d,%d)", g.px, g.py)
	}
}

func TestEginbearInvisibilitySuppressesGuardTracking(t *testing.T) {
	g := eginbearGuardGame(t)
	g.px, g.py, g.cd, g.remoaru = 14, 38, 0, 25
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.cur.npcAt(13, 36) < 0 || g.cur.npcAt(14, 36) >= 0 || g.remoaru != 24 {
		t.Fatalf("透明玩家踏 gate 不得令守衛追蹤：timer=%d npcs=%+v", g.remoaru, g.cur.npcs)
	}
	g.cd = 0
	if err := g.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.px != 14 || g.py != 36 || g.remoaru != 23 {
		t.Fatalf("透明玩家應由右側越過守衛：player=(%d,%d) timer=%d", g.px, g.py, g.remoaru)
	}
}

// TestDumpEginbearTrackingGuard 固定兩個 production renderer 狀態：可見玩家被守衛
// 橫移攔下，以及隱形玩家由右欄越過。正常可達性另由完整 production trace 驗證。
func TestDumpEginbearTrackingGuard(t *testing.T) {
	if os.Getenv("DQ3_DUMP_EGINBEAR") == "" {
		t.Skip("設 DQ3_DUMP_EGINBEAR=1 才執行")
	}
	out := os.Getenv("EGINBEAR_OUT")
	if out == "" {
		t.Fatal("需設 EGINBEAR_OUT")
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

	visible := eginbearGuardGame(t)
	visible.px, visible.py, visible.cd = 14, 38, 0
	if err := visible.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	dump(visible, "eginbear_guard_visible_block")

	invisible := eginbearGuardGame(t)
	invisible.px, invisible.py, invisible.cd, invisible.remoaru = 14, 38, 0, 25
	if err := invisible.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	invisible.cd = 0
	if err := invisible.step(InputState{DirHeld: 1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	dump(invisible, "eginbear_invisibility_pass")
}
