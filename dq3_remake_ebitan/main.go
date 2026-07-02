// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 3:載入真實 DQ3CON.MAP + DQ3.BLK + DQ3.PAL + DQ3MST.BLS(全用移植的 Go parser),
// 地表 viewport 隨主角捲動、方向鍵走動、主角 sprite masked blit + 走路動畫。朝「可玩」推進。
package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

const (
	ScreenW, ScreenH = 640, 350
	TileW, TileH     = 32, 24
	Cols, Rows       = ScreenW/TileW + 2, ScreenH/TileH + 2 // viewport tile 數(多一圈)
	moveCooldown     = 6                                    // 每幾幀走一格(grid walk)
)

type Game struct {
	pal      []dq3data.Color
	blk      *dq3data.BLK
	m        *dq3data.FieldMap
	hero     *dq3data.CharSprite
	px, py   int // 主角 tile 座標
	facing   int // 0..3(dir)
	walk     int // 0/1 走路動畫相位
	cd, anim int
	frame    *ebiten.Image
	rgba     []byte
}

func (g *Game) Update() error {
	moved := false
	if g.cd > 0 {
		g.cd--
	} else {
		dx, dy, face := 0, 0, g.facing
		switch {
		case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
			dy, face = 1, 0
		case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
			dy, face = -1, 1
		case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
			dx, face = -1, 2
		case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
			dx, face = 1, 3
		}
		if dx != 0 || dy != 0 {
			nx, ny := g.px+dx, g.py+dy
			if nx >= 0 && ny >= 0 && nx < g.m.W && ny < g.m.H {
				g.px, g.py = nx, ny // 註:碰撞(BLKBM.DAT 屬性)之後補
			}
			g.facing = face
			g.cd = moveCooldown
			moved = true
		}
	}
	g.anim++
	if moved || g.anim%18 == 0 { // 走動 / 待機都擺手腳(對齊 remake NPC idle 動畫)
		if g.anim%9 == 0 {
			g.walk ^= 1
		}
	}
	g.renderFrame() // ★ 在 Update 渲染 → WritePixels 於 Draw 前生效(Ebiten:WritePixels 下一幀才顯示)
	return nil
}

// renderFrame:把地表 viewport(隨主角捲動)+ 主角 sprite 渲成 640×350 → g.frame。
func (g *Game) renderFrame() {
	ox, oy := g.px-Cols/2, g.py-Rows/2 // viewport 左上 tile(主角置中)
	for cy := 0; cy < Rows; cy++ {
		for cx := 0; cx < Cols; cx++ {
			tile := g.blk.Tile(g.m.Tile(ox+cx, oy+cy))
			blitTile(g.rgba, cx*TileW, cy*TileH, tile, g.pal)
		}
	}
	blitSprite(g.rgba, (Cols/2)*TileW, (Rows/2)*TileH, g.hero.Frames[g.facing*dq3data.CharWalk+g.walk], g.pal)
	g.frame.WritePixels(g.rgba)
}

func blitTile(rgba []byte, dx, dy int, px [24][32]uint8, pal []dq3data.Color) {
	for r := 0; r < TileH; r++ {
		sy := dy + r
		if sy < 0 || sy >= ScreenH {
			continue
		}
		for x := 0; x < TileW; x++ {
			sx := dx + x
			if sx < 0 || sx >= ScreenW {
				continue
			}
			idx := int(px[r][x])
			var c dq3data.Color
			if idx < len(pal) {
				c = pal[idx]
			}
			o := (sy*ScreenW + sx) * 4
			rgba[o], rgba[o+1], rgba[o+2], rgba[o+3] = c.R, c.G, c.B, 255
		}
	}
}

func blitSprite(rgba []byte, dx, dy int, f dq3data.Frame, pal []dq3data.Color) {
	for r := 0; r < dq3data.CharH; r++ {
		sy := dy + r
		if sy < 0 || sy >= ScreenH {
			continue
		}
		for x := 0; x < dq3data.CharW; x++ {
			if !f.Opaque[r][x] { // 透明像素跳過(masked blit)
				continue
			}
			sx := dx + x
			if sx < 0 || sx >= ScreenW {
				continue
			}
			idx := int(f.Px[r][x])
			var c dq3data.Color
			if idx < len(pal) {
				c = pal[idx]
			}
			o := (sy*ScreenW + sx) * 4
			rgba[o], rgba[o+1], rgba[o+2], rgba[o+3] = c.R, c.G, c.B, 255
		}
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.DrawImage(g.frame, nil) // frame 已在 Update 渲好
}

func (g *Game) Layout(int, int) (int, int) { return ScreenW, ScreenH }

func assetsDir() string {
	if d := os.Getenv("DQ3_ASSETS"); d != "" {
		return d
	}
	return "assets_raw"
}

func mustRead(name string) []byte {
	d, err := os.ReadFile(filepath.Join(assetsDir(), name))
	if err != nil {
		log.Fatalf("讀 %s 失敗:%v(設 DQ3_ASSETS 指向原版素材夾)", name, err)
	}
	return d
}

func main() {
	g := &Game{
		pal:  dq3data.DecodePalette(mustRead("DQ3.PAL"), 256),
		rgba: make([]byte, ScreenW*ScreenH*4),
	}
	var err error
	if g.blk, err = dq3data.OpenBLK(mustRead("DQ3.BLK")); err != nil {
		log.Fatalf("OpenBLK: %v", err)
	}
	if g.m, err = dq3data.OpenFieldMap(mustRead("DQ3CON.MAP")); err != nil {
		log.Fatalf("OpenFieldMap: %v", err)
	}
	g.hero = dq3data.LoadCharSprite(mustRead("DQ3MST.BLS"), 0)
	g.px, g.py = g.m.W/2, g.m.H/2 // 起點(暫用地圖中心;真實起點座標之後接)
	g.frame = ebiten.NewImage(ScreenW, ScreenH)
	g.renderFrame() // 首幀
	log.Printf("地表 %d×%d、%d tiles、%d 色、主角 8 frame — 方向鍵走動", g.m.W, g.m.H, g.blk.Count, len(g.pal))

	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (walkable overworld)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
