package game

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
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

func inkInRect(rgba []byte, r image.Rectangle) int {
	n := 0
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			off := (y*ScreenW + x) * 4
			if rgba[off] != 0 || rgba[off+1] != 0 || rgba[off+2] != 0 {
				n++
			}
		}
	}
	return n
}

func TestPartyHUDUsesOriginalPackColumnAnchors(t *testing.T) {
	g := partyRenderFixture(t)
	g.companions = nil
	g.heroName = []int{15, 16, 17, 18, 19}
	clear(g.rgba)
	g.drawPartyHUD(g.rgba, dq3data.Color{R: 255, G: 255, B: 255})

	// 原版 sub_18222：label 在 base+2 bytes、姓名/數值在 base+4
	// bytes；姓名最多四字。略過頂框列，只量 glyph ink。
	if inkInRect(g.rgba, image.Rect(184, 240, 248, 254)) == 0 {
		t.Fatal("原版姓名 anchor (184,238) 沒有 glyph ink")
	}
	if inkInRect(g.rgba, image.Rect(248, 240, 264, 254)) != 0 {
		t.Fatal("第五個姓名 glyph 不應越入下一欄")
	}
	if inkInRect(g.rgba, image.Rect(168, 254, 184, 270)) == 0 ||
		inkInRect(g.rgba, image.Rect(184, 254, 232, 270)) == 0 {
		t.Fatal("H label/value 未依原版 +2/+4 byte anchors 繪製")
	}
	if inkInRect(g.rgba, image.Rect(168, 286, 184, 302)) == 0 ||
		inkInRect(g.rgba, image.Rect(184, 286, 232, 302)) == 0 {
		t.Fatal("職業 glyph/等級值未依原版末列 anchors 繪製")
	}
}

func TestPartyHUDOpensThroughProductionInput(t *testing.T) {
	g := partyRenderFixture(t)
	if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if !g.cmd.open {
		t.Fatal("正式確認輸入未開啟含隊伍 HUD 的地表命令窗")
	}
	g.renderFrame()
	if inkInRect(g.rgba, image.Rect(152, 238, 504, 318)) == 0 {
		t.Fatal("正式命令入口沒有畫出 pack HUD")
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
	if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if !g.cmd.open {
		t.Fatal("正式確認輸入未開啟命令窗")
	}
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
