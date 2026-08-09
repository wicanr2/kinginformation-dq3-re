package game

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func firstOpenOverworldTile(sc *Scene) (int, int, bool) {
	if sc == nil {
		return 0, 0, false
	}
	for y := 1; y < sc.h-1; y++ {
		for x := 1; x < sc.w-1; x++ {
			if sc.Blocked(x, y) {
				continue
			}
			for _, d := range []int{0, 1, 2, 3} {
				dx, dy := dirDelta(d)
				if !sc.Blocked(x+dx, y+dy) {
					return x, y, true
				}
			}
		}
	}
	return 0, 0, false
}

func partyRenderFixture(t *testing.T) *Game {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	g.cur, g.inTown, g.curCty = g.over, false, -1
	x, y, ok := firstOpenOverworldTile(g.over)
	if !ok {
		t.Fatal("找不到可重播隊列的地表開闊格")
	}
	g.px, g.py, g.facing, g.walk = x+1, y, 2, 0
	if g.cur.Blocked(g.px, g.py) {
		g.px, g.py = x, y
	}
	g.heroName = []int{15}
	g.companions = []*Member{
		newLevelOneMember([]int{16}, 1, 0, &g.prng, g.tavern.equipment),
		newLevelOneMember([]int{17}, 3, 0, &g.prng, g.tavern.equipment),
		newLevelOneMember([]int{18}, 4, 1, &g.prng, g.tavern.equipment),
	}
	for _, member := range g.companions {
		member.CurHP = member.MaxHP()
	}
	if g.partySprite(1, 0) == nil || g.partySprite(3, 0) == nil || g.partySprite(4, 1) == nil {
		t.Fatal("pack 定義的三個隊員 sprite 未能從 DQ3MST.BLS 解出")
	}
	g.resetPartyTrail()
	return g
}

func TestPartyRenderTrailAndPackSprites(t *testing.T) {
	g := partyRenderFixture(t)
	g.renderFrame()
	soloFrame := append([]byte(nil), g.rgba...)

	// A legal pre-move position is recorded, then the production renderer is
	// asked to draw the resulting four-person scene at the new player tile.
	oldX, oldY := g.px, g.py
	g.facing = 3
	g.recordPartyTrail()
	g.px++
	g.renderFrame()
	if g.partyTrailLen != 1 || g.partyTrail[0].x != oldX || g.partyTrail[0].y != oldY {
		t.Fatalf("隊列歷史未保留移動前格：len=%d trail=%+v old=(%d,%d)",
			g.partyTrailLen, g.partyTrail[0], oldX, oldY)
	}
	if bytes.Equal(soloFrame, g.rgba) {
		t.Fatal("加入三名同伴後 renderFrame 與單人畫面完全相同")
	}

	// The second and third follower positions are intentionally still anchored
	// until enough legal steps are recorded; they must not be guessed from the
	// current player coordinate.
	if g.partyTrail[1].x != oldX || g.partyTrail[1].y != oldY {
		t.Fatalf("短 trail 的後續同伴不應取得未證實座標：%+v", g.partyTrail[1])
	}
}

func TestDumpPartyFieldPNG(t *testing.T) {
	if os.Getenv("DQ3_DUMP_PARTY") == "" {
		t.Skip("設 DQ3_DUMP_PARTY=1 才輸出隊伍對拍圖")
	}
	out := os.Getenv("PARTY_OUT")
	if out == "" {
		t.Fatal("DQ3_DUMP_PARTY 需要 PARTY_OUT")
	}
	g := partyRenderFixture(t)
	for i := 0; i < 3; i++ {
		g.recordPartyTrail()
		g.px++
	}
	g.cmd.open = true
	g.renderFrame()
	img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
	copy(img.Pix, g.rgba)
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(out, "party_field_hud.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
	t.Logf("隊伍對拍圖：%s", path)
}
