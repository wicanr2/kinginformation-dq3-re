// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 3+:載入真實地表(DQ3CON.MAP+DQ3.BLK)+ 阿里阿罕城(CTY00.DAT+DQ31.BLK),全用移植的 Go parser。
// 方向鍵走動、viewport 捲動、主角 masked sprite + 走路動畫、BLKBM 碰撞(不可穿山/海),Enter 進城 / Esc 回地表。
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
	ViewCols         = ScreenW / TileW              // 20(對齊 C VIEW_COLS)
	ViewRows         = (ScreenH + TileH - 1) / TileH // 15(ceil,對齊 C VIEW_ROWS)
	moveCooldown     = 6                             // 每幾幀走一格(grid walk)
)

// npcInst 是場景裡一個已定位、已載入 sprite 的 NPC。
type npcInst struct {
	x, y int
	spr  *dq3data.CharSprite
}

// Scene 是一張可走動的地圖(地表或城鎮),統一 render + 碰撞邏輯。
type Scene struct {
	blk            *dq3data.BLK
	attr           *dq3data.BlockAttr
	pal            []dq3data.Color
	w, h           int
	tileAt         func(x, y int) int // (x,y) → BLK tile 索引
	spawnX, spawnY int                // 城鎮進入 spawn(地表不用)
	npcs           []npcInst
}

// Blocked:出界 / BLKBM 屬性 bit0=1 / 有 NPC 佔格 → 不可走(移植 dq3_scene_walkable + npc_at)。
func (sc *Scene) Blocked(x, y int) bool {
	if x < 0 || y < 0 || x >= sc.w || y >= sc.h {
		return true
	}
	if sc.attr.Blocked(sc.tileAt(x, y)) {
		return true
	}
	for i := range sc.npcs { // NPC 擋路(對齊 dq3_scene_npc_at)
		if sc.npcs[i].x == x && sc.npcs[i].y == y {
			return true
		}
	}
	return false
}

type Game struct {
	over, town     *Scene // 地表 / 阿里阿罕城
	cur            *Scene
	inTown         bool
	overPx, overPy int // 記住進城前的地表座標(Esc 回來用)
	hero           *dq3data.CharSprite
	px, py         int // 主角在 cur 內的 tile 座標
	facing         int // 0..3
	walk           int // 0/1 走路動畫相位
	cd, anim       int
	frame          *ebiten.Image
	rgba           []byte
}

func (g *Game) Update() error {
	moved := false
	// 進城 / 回地表(邊緣觸發:上一幀沒按、這幀按下)
	if g.cd == 0 {
		if !g.inTown && ebiten.IsKeyPressed(ebiten.KeyEnter) {
			g.enterTown()
			return nil
		}
		if g.inTown && ebiten.IsKeyPressed(ebiten.KeyEscape) {
			g.exitTown()
			return nil
		}
	}
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
			g.facing = face // 撞牆也轉向(對齊原版:面向先變,可走才移動)
			nx, ny := g.px+dx, g.py+dy
			if !g.cur.Blocked(nx, ny) { // 碰撞:BLKBM attr&1 = 不可走(山/海/牆)
				g.px, g.py = nx, ny
				moved = true
			}
			g.cd = moveCooldown
		}
	}
	g.anim++
	if moved || g.anim%18 == 0 { // 走動 / 待機都擺手腳
		if g.anim%9 == 0 {
			g.walk ^= 1
		}
	}
	g.renderFrame() // ★ 在 Update 渲染 → WritePixels 於 Draw 前生效(Ebiten:WritePixels 下一幀才顯示)
	return nil
}

func (g *Game) enterTown() {
	g.overPx, g.overPy = g.px, g.py
	g.cur, g.inTown = g.town, true
	g.px, g.py = g.town.spawnX, g.town.spawnY
	g.cd = moveCooldown
	g.renderFrame()
}

func (g *Game) exitTown() {
	g.cur, g.inTown = g.over, false
	g.px, g.py = g.overPx, g.overPy
	g.cd = moveCooldown
	g.renderFrame()
}

func clampi(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// renderFrame:把 cur 的 viewport(攝影機 clamp 在地圖內,對齊 C)+ NPC + 主角渲成 640×350 → g.frame。
func (g *Game) renderFrame() {
	sc := g.cur
	// 攝影機:主角置中,但夾在地圖邊界內(移植 dq3_scene 的 cam clamp)→ 邊緣不露黑
	camX := clampi(g.px-ViewCols/2, 0, max0(sc.w-ViewCols))
	camY := clampi(g.py-ViewRows/2, 0, max0(sc.h-ViewRows))
	for cy := 0; cy < ViewRows; cy++ {
		for cx := 0; cx < ViewCols; cx++ {
			tile := sc.blk.Tile(sc.tileAt(camX+cx, camY+cy))
			blitTile(g.rgba, cx*TileW, cy*TileH, tile, sc.pal)
		}
	}
	for i := range sc.npcs { // NPC(在視窗內才畫;靜態朝下 frame)
		n := &sc.npcs[i]
		if n.spr == nil || n.x < camX || n.x >= camX+ViewCols || n.y < camY || n.y >= camY+ViewRows {
			continue
		}
		blitSprite(g.rgba, (n.x-camX)*TileW, (n.y-camY)*TileH, n.spr.Frames[0], sc.pal)
	}
	blitSprite(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH,
		g.hero.Frames[g.facing*dq3data.CharWalk+g.walk], sc.pal)
	g.frame.WritePixels(g.rgba)
}

func max0(v int) int {
	if v < 0 {
		return 0
	}
	return v
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
	pal := dq3data.DecodePalette(mustRead("DQ3.PAL"), 256) // 城鎮/地表共用(tile 只用色 0..15)
	g := &Game{rgba: make([]byte, ScreenW*ScreenH*4)}

	// 地表 scene
	overBLK, err := dq3data.OpenBLK(mustRead("DQ3.BLK"))
	if err != nil {
		log.Fatalf("OpenBLK 地表: %v", err)
	}
	fm, err := dq3data.OpenFieldMap(mustRead("DQ3CON.MAP"))
	if err != nil {
		log.Fatalf("OpenFieldMap: %v", err)
	}
	g.over = &Scene{
		blk: overBLK, attr: dq3data.OpenBlockAttr(mustRead("BLKBM.DAT")), pal: pal,
		w: fm.W, h: fm.H, tileAt: fm.Tile,
	}

	// 阿里阿罕城 scene(CTY00.DAT section 0 / DQ31.BLK + BLKBM1.DAT)
	townBLK, err := dq3data.OpenBLK(mustRead("DQ31.BLK"))
	if err != nil {
		log.Fatalf("OpenBLK 城鎮: %v", err)
	}
	tw, err := dq3data.OpenTown(mustRead("CTY00.DAT"), 0)
	if err != nil {
		log.Fatalf("OpenTown 阿里阿罕: %v", err)
	}
	g.town = &Scene{
		blk: townBLK, attr: dq3data.OpenBlockAttr(mustRead("BLKBM1.DAT")), pal: pal,
		w: tw.W, h: tw.H, tileAt: tw.Tile,
		spawnX: tw.SpawnX, spawnY: tw.SpawnY,
	}
	// NPC sprite:DQ3MAN.BLS,entry_base=(b2-4)*4;同 b2 共用一份(移植 dq3_scene_load_npc_sprites)
	manBLS := mustRead("DQ3MAN.BLS")
	sprCache := map[int]*dq3data.CharSprite{}
	for _, n := range tw.NPCs {
		if n.B2 < 4 { // key<4 無對應 BLS 角色
			continue
		}
		spr, ok := sprCache[n.B2]
		if !ok {
			spr = dq3data.LoadCharSprite(manBLS, (n.B2-4)*4)
			sprCache[n.B2] = spr
		}
		g.town.npcs = append(g.town.npcs, npcInst{x: n.X, y: n.Y, spr: spr})
	}

	g.hero = dq3data.LoadCharSprite(mustRead("DQ3MST.BLS"), 0)
	g.cur = g.over
	g.px, g.py = g.over.w/2, g.over.h/2 // 地表起點(暫用中心)
	if os.Getenv("DQ3_START") == "town" {  // debug:直接起在城內(headless 截圖驗證用)
		g.cur, g.inTown = g.town, true
		g.px, g.py = g.town.spawnX, g.town.spawnY
	}
	g.frame = ebiten.NewImage(ScreenW, ScreenH)
	g.renderFrame() // 首幀
	log.Printf("地表 %d×%d / 阿里阿罕 %d×%d(spawn %d,%d、NPC %d)— 方向鍵走、Enter 進城、Esc 回地表",
		g.over.w, g.over.h, g.town.w, g.town.h, tw.SpawnX, tw.SpawnY, len(g.town.npcs))

	ebiten.SetWindowSize(ScreenW*2, ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (overworld + アリアハン)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
