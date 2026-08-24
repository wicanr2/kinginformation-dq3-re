package game

import (
	"os"
	"testing"
)

func TestOpeningMotherEscortUsesVisibleFrames(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.startOpening()
	if !g.startMotherEscort() {
		t.Fatal("CTY00 sec4 應找到 pack 指定母親並開始逐格帶路")
	}
	steps := 0
	for g.openingEscortAnimating() && steps < 1000 {
		g.advanceOpeningEscort()
		steps++
	}
	if steps <= len(g.openingEscort.Frames) {
		t.Fatalf("帶路應有可見 hold frames，僅 %d 次更新", steps)
	}
	if g.curCty != 0 || g.cur == nil || g.cur.sec != 0 || g.px != 21 || g.py != 9 {
		t.Fatalf("帶路完成落點錯：cty=%d sec=%v pos=(%d,%d)", g.curCty, currentSceneSection(g.cur), g.px, g.py)
	}
	if !g.dlg.open {
		t.Fatal("帶路完成後應開 rec80 對話")
	}
	if g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatalf("rec80 關閉前不得交易旗標：17=%v 50=%v", g.storyFlag(0x17), g.storyFlag(0x50))
	}
}

func TestOpeningKingAudienceRendersHero(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = throne, throne, ctyAliahanCastle, true
	g.px, g.py, g.facing = aliahanKingX, aliahanKingY+1, 1
	g.showTitle = false
	g.renderFrame()
	visible := append([]byte(nil), g.rgba...)
	g.remoaru = 1
	g.renderFrame()

	camX := clampi(g.px-ViewCols/2, 0, max0(throne.w-ViewCols))
	camY := clampi(g.py-ViewRows/2, 0, max0(throne.h-ViewRows))
	sx, sy := (g.px-camX)*TileW, (g.py-camY)*TileH
	different := false
	for y := sy; y < sy+TileH && !different; y++ {
		for x := sx; x < sx+TileW; x++ {
			o := (y*ScreenW + x) * 4
			if visible[o] != g.rgba[o] || visible[o+1] != g.rgba[o+1] || visible[o+2] != g.rgba[o+2] {
				different = true
				break
			}
		}
	}
	if !different {
		t.Fatal("王座正前方的可見畫面與透明狀態相同：勇者 sprite 未繪出")
	}
}

// TestDumpOpeningCastleArrivalPath is an opt-in evidence helper: it derives a
// legal shortest tile path from the original CTY00 data between handler54's
// transition landing and the castle portal.  The printed coordinates are only
// a candidate until checked against the original video; production JSON must
// not be generated from this test alone.
func TestDumpOpeningCastleArrivalPath(t *testing.T) {
	if os.Getenv("DQ3_DUMP_OPENING_CASTLE_PATH") == "" {
		t.Skip("set DQ3_DUMP_OPENING_CASTLE_PATH=1 to derive the CTY00 candidate route")
	}
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !g.finishMotherEscort() {
		t.Fatal("無法載入 pack 指定的城鎮抵達段")
	}
	path := tracePath(g.cur, g.px, g.py, 21, 9)
	if len(path) == 0 {
		t.Fatal("CTY00 landing cannot reach the tile before the castle portal")
	}
	x, y := g.px, g.py
	t.Logf("route start=(%d,%d)", x, y)
	for i, dir := range path {
		dx, dy := dirDelta(dir)
		x, y = x+dx, y+dy
		t.Logf("route[%d]=(%d,%d) dir=%d", i+1, x, y, dir)
	}
}
