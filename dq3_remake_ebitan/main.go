// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 2/3:載入真實 DQ3CON.MAP + DQ3.BLK + DQ3.PAL(全用移植的 Go parser),
// 把地表 overworld 的一個 viewport 渲成 640×350 畫面 —— 證明「資產→Go解析→Ebiten渲染」到「畫出真實遊戲畫面」。
package main

import (
	"image/color"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

const ScreenW, ScreenH = 640, 350 // 對齊 C remake 邏輯畫布
const TileW, TileH = 32, 24       // BLK tile 尺寸

type Game struct {
	frame *ebiten.Image // 預渲的地表 viewport(phase 2 靜態;移動/捲動之後補)
}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {
	if g.frame != nil {
		screen.DrawImage(g.frame, nil)
	} else {
		screen.Fill(color.RGBA{16, 24, 48, 255})
	}
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

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// renderOverworld:用移植的 parser 把地表 viewport 渲成 640×350 RGBA → ebiten.Image。
func renderOverworld() *ebiten.Image {
	pal := dq3data.DecodePalette(mustRead("DQ3.PAL"), 256)
	blk, err := dq3data.OpenBLK(mustRead("DQ3.BLK"))
	if err != nil {
		log.Fatalf("OpenBLK: %v", err)
	}
	m, err := dq3data.OpenFieldMap(mustRead("DQ3CON.MAP"))
	if err != nil {
		log.Fatalf("OpenFieldMap: %v", err)
	}
	log.Printf("地表 %d×%d、DQ3.BLK %d tiles、DQ3.PAL %d 色", m.W, m.H, blk.Count, len(pal))

	ox := envInt("DQ3_OX", m.W/2-10) // viewport 左上的 tile 座標(可用 env 調)
	oy := envInt("DQ3_OY", m.H/2-7)
	rgba := make([]byte, ScreenW*ScreenH*4)
	for cy := 0; cy <= ScreenH/TileH; cy++ {
		for cx := 0; cx <= ScreenW/TileW; cx++ {
			tile := blk.Tile(m.Tile(ox+cx, oy+cy)) // 每 tile 只解一次
			for py := 0; py < TileH; py++ {
				for px := 0; px < TileW; px++ {
					sx, sy := cx*TileW+px, cy*TileH+py
					if sx >= ScreenW || sy >= ScreenH {
						continue
					}
					idx := int(tile[py][px])
					c := color.RGBA{}
					if idx < len(pal) {
						c = color.RGBA{pal[idx].R, pal[idx].G, pal[idx].B, 255}
					}
					o := (sy*ScreenW + sx) * 4
					rgba[o], rgba[o+1], rgba[o+2], rgba[o+3] = c.R, c.G, c.B, 255
				}
			}
		}
	}
	img := ebiten.NewImage(ScreenW, ScreenH)
	img.WritePixels(rgba)
	return img
}

func main() {
	g := &Game{frame: renderOverworld()}
	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (overworld)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
