package game

import (
	"os"
	"reflect"
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
	dialogueFrame := g.openingEscort.ArrivalFrames[g.openingEscort.DialogueFrameIndex].Player
	if g.curCty != 0 || g.cur == nil || g.cur.sec != 0 || g.px != dialogueFrame.X || g.py != dialogueFrame.Y {
		t.Fatalf("opening 對話序列觸發位置錯：cty=%d sec=%v pos=(%d,%d)，want=(%d,%d)",
			g.curCty, currentSceneSection(g.cur), g.px, g.py, dialogueFrame.X, dialogueFrame.Y)
	}
	if !g.dlg.open || g.openingEscortDialogue != 0 {
		t.Fatal("抵達 pack 指定格後應開第一段對話")
	}
	if !reflect.DeepEqual(g.dlg.buf, g.cur.dlgText.Record(80)) {
		t.Fatal("第一段不是 D3TXT01 record 80")
	}
	if g.dlg.layout.GlyphHoldFrames != 3 || g.dlg.revealCells != 0 {
		t.Fatalf("逐字初值錯：hold=%d visible=%d", g.dlg.layout.GlyphHoldFrames, g.dlg.revealCells)
	}
	g.dlg.Tick()
	g.dlg.Tick()
	if g.dlg.revealCells != 0 {
		t.Fatal("未滿 pack hold frames 不得提早顯示 glyph")
	}
	g.dlg.Tick()
	if g.dlg.revealCells != 1 {
		t.Fatal("滿 pack hold frames 應顯示一個 glyph cell")
	}
	if g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatalf("rec80 關閉前不得交易旗標：17=%v 50=%v", g.storyFlag(0x17), g.storyFlag(0x50))
	}
	g.dlg.open = false
	if !g.resumeOpeningEscortAfterDialogue() {
		t.Fatal("第一段關閉後應在同一格開第二段對話")
	}
	if !g.dlg.open || g.openingEscortDialogue != 1 || g.storyFlag(0x17) || !g.storyFlag(0x50) {
		t.Fatalf("第二段對話狀態錯：open=%v index=%d flag17=%v flag50=%v",
			g.dlg.open, g.openingEscortDialogue, g.storyFlag(0x17), g.storyFlag(0x50))
	}
	if !reflect.DeepEqual(g.dlg.buf, g.cur.dlgText.Record(79)) {
		t.Fatal("第二段不是 D3TXT01 record 79")
	}
	g.dlg.open = false
	if !g.resumeOpeningEscortAfterDialogue() {
		t.Fatal("最後一段關閉後應交易旗標並恢復剩餘自動步行")
	}
	for g.openingEscortAnimating() && steps < 2000 {
		g.advanceOpeningEscort()
		steps++
	}
	if g.px != 21 || g.py != 9 || g.storyFlag(0x17) == false || g.storyFlag(0x50) {
		t.Fatalf("剩餘步行完成狀態錯：pos=(%d,%d) flag17=%v flag50=%v",
			g.px, g.py, g.storyFlag(0x17), g.storyFlag(0x50))
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
