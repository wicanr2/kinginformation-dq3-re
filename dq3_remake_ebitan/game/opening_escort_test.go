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
	if g.curCty != 0 || g.cur == nil || g.cur.sec != 0 || g.px != 8 || g.py != 38 {
		t.Fatalf("帶路完成落點錯：cty=%d sec=%v pos=(%d,%d)", g.curCty, currentSceneSection(g.cur), g.px, g.py)
	}
	if !g.dlg.open {
		t.Fatal("帶路完成後應開 rec80 對話")
	}
}
