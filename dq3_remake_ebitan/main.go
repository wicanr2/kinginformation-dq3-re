// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 1:Ebiten 開窗 + 主迴圈 + 用移植的 Go parser 載入真實 DQ3.PAL 畫成 palette,
// 證明「真實資產 → Go 解析(internal/dq3data)→ Ebiten 渲染」管線可行。
package main

import (
	"image/color"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

const ScreenW, ScreenH = 640, 350 // 對齊 C remake 邏輯畫布

type Game struct {
	pal []dq3data.Color
}

func (g *Game) Update() error { return nil }

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{16, 24, 48, 255})
	// palette swatches:真實 DQ3.PAL → dq3data.DecodePalette → Ebiten
	const cols = 16
	for i, c := range g.pal {
		x := float32(20 + (i%cols)*38)
		y := float32(24 + (i/cols)*32)
		vector.DrawFilledRect(screen, x, y, 34, 26, color.RGBA{c.R, c.G, c.B, 255}, false)
	}
}

func (g *Game) Layout(int, int) (int, int) { return ScreenW, ScreenH }

func loadPAL() []dq3data.Color {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "assets_raw"
	}
	d, err := os.ReadFile(filepath.Join(dir, "DQ3.PAL"))
	if err != nil {
		log.Printf("讀 DQ3.PAL 失敗(%v)— 空 palette(設 DQ3_ASSETS 指向原版素材夾)", err)
		return nil
	}
	pal := dq3data.DecodePalette(d, 256)
	log.Printf("DQ3.PAL → %d 色", len(pal))
	return pal
}

func main() {
	g := &Game{pal: loadPAL()}
	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (phase 1)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
