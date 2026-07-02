// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 3+:載入真實地表(DQ3CON.MAP+DQ3.BLK)+ 阿里阿罕城(CTY00.DAT+DQ31.BLK),全用移植的 Go parser。
// 方向鍵走動、viewport 捲動、主角 masked sprite + 走路動畫、BLKBM 碰撞(不可穿山/海),Enter 進城 / Esc 回地表。
package game

import (
	"fmt"
	"io/fs"
	"log"
	"math/rand"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gaudio"
)

// 場景配樂軌號(對齊 C dq3_audio g_scene_track):地表 FIELD=6、阿里阿罕城 CASTLE=1。
const (
	trackField  = 6
	trackCastle = 1
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
	x, y     int
	ctrl, b4 int // 互動:(ctrl>>3)&7=子型;b4=對話 rec / 設施索引
	spr      *dq3data.CharSprite
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
	return sc.npcAt(x, y) >= 0 // NPC 擋路(對齊 dq3_scene_npc_at)
}

// npcAt 回 (x,y) 上的 NPC 索引;無則 -1。移植 dq3_scene_npc_at。
func (sc *Scene) npcAt(x, y int) int {
	for i := range sc.npcs {
		if sc.npcs[i].x == x && sc.npcs[i].y == y {
			return i
		}
	}
	return -1
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
	dlg            Dialogue // 對話視窗
	cmd            CmdMenu  // 野外命令窗
	battle         Battle   // 戰鬥場景
	music          *gaudio.Music
	input          *Input // 抽象輸入(鍵盤 + 觸控)
	frame          *ebiten.Image
	rgba           []byte
}

// 場景配樂軌:BATTLE=14(地圖遇敵,對齊 C g_scene_track)。
const trackBattle = 14

// dpadEdge:方向鍵「剛按下」→ facing 碼(0下 1上 2左 3右);無則 -1。選單導覽用。
func dpadEdge() int {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
		return 0
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
		return 1
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		return 2
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		return 3
	}
	return -1
}

// selectCommand:命令窗選定一項後的動作。對話可用;其餘指令需隊伍/道具/事件系統(尚未移植)→ 暫關窗。
func (g *Game) selectCommand(cmd int) {
	g.cmd.open = false
	if cmd == cmdTalk {
		fx, fy := frontTile(g.px, g.py, g.facing)
		if idx := g.cur.npcAt(fx, fy); idx >= 0 {
			n := &g.cur.npcs[idx]
			if sub := (n.ctrl >> 3) & 7; sub <= 1 {
				g.dlg.Open(n.b4)
			}
		}
	}
}

func (g *Game) Update() error {
	in := g.input.Poll() // 抽象輸入(鍵盤 + 觸控合流)
	moved := false

	// 戰鬥 modal:結束時回地表音樂
	if g.battle.active {
		if g.battle.input(in) {
			g.music.Play(trackField)
		}
		g.renderFrame()
		return nil
	}
	// 對話 modal:只吃 A(推進/關閉),不移動
	if g.dlg.open {
		if in.Confirm {
			g.dlg.Advance()
		}
		g.renderFrame()
		return nil
	}
	// 命令窗 modal:方向移游標、A 選定、B 關窗
	if g.cmd.open {
		switch {
		case in.Cancel:
			g.cmd.open = false
		case in.Confirm:
			g.selectCommand(g.cmd.cursor)
		default:
			if in.DirEdge >= 0 {
				g.cmd.move(in.DirEdge)
			}
		}
		g.renderFrame()
		return nil
	}
	// A:城內開命令窗;地表 → 進阿里阿罕城(觸控/鍵盤共用)
	if in.Confirm {
		if g.inTown {
			g.cmd.Open()
		} else {
			g.enterTown()
		}
		g.renderFrame()
		return nil
	}
	if in.Enter && !g.inTown { // 鍵盤 Enter 也進城
		g.enterTown()
		return nil
	}
	if in.Cancel && g.inTown {
		g.exitTown()
		return nil
	}
	// 移動(方向 held + grid cooldown)
	if g.cd > 0 {
		g.cd--
	} else if in.DirHeld >= 0 {
		g.facing = in.DirHeld // 撞牆也轉向(對齊原版:面向先變,可走才移動)
		dx, dy := dirDelta(in.DirHeld)
		nx, ny := g.px+dx, g.py+dy
		if !g.cur.Blocked(nx, ny) { // 碰撞:BLKBM attr&1 + NPC 佔格
			g.px, g.py = nx, ny
			moved = true
		}
		g.cd = moveCooldown
	}
	if moved && !g.inTown && rand.Intn(16) == 0 { // 地表隨機遇敵(~1/16 步)
		g.startEncounter()
	}
	g.anim++
	if moved || g.anim%18 == 0 { // 走動 / 待機都擺手腳
		if g.anim%9 == 0 {
			g.walk ^= 1
		}
	}
	g.renderFrame() // ★ 在 Update 渲染 → WritePixels 於 Draw 前生效
	return nil
}

func (g *Game) enterTown() {
	g.overPx, g.overPy = g.px, g.py
	g.cur, g.inTown = g.town, true
	g.px, g.py = g.town.spawnX, g.town.spawnY
	g.cd = moveCooldown
	g.music.Play(trackCastle) // 阿里阿罕(CASTLE 曲)
	g.renderFrame()
}

func (g *Game) exitTown() {
	g.cur, g.inTown = g.over, false
	g.px, g.py = g.overPx, g.overPy
	g.cd = moveCooldown
	g.music.Play(trackField) // 地表(FIELD 曲)
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

// startEncounter:從阿里阿罕周邊弱怪池挑一隻有 sprite 的,開戰 + 切戰鬥音樂。
func (g *Game) startEncounter() {
	pool := []int{5, 0, 1, 3, 4, 6, 7, 8} // 低階怪(id5=史萊姆…);start 會跳過空 sprite
	for _, i := range rand.Perm(len(pool)) {
		if g.battle.start(pool[i], int64(g.anim)*2654+1) {
			g.music.Play(trackBattle)
			return
		}
	}
}

// renderFrame:戰鬥時畫戰鬥場景;否則畫地圖 viewport(攝影機 clamp)+ NPC + 主角 → g.frame。
func (g *Game) renderFrame() {
	if g.battle.active {
		g.battle.draw(g.rgba, g.cur.pal)
		g.frame.WritePixels(g.rgba)
		return
	}
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
	white := dq3data.Color{R: 255, G: 255, B: 255}
	if len(sc.pal) > 15 {
		white = sc.pal[15]
	}
	if g.cmd.open { // 命令窗(左上)
		yellow := dq3data.Color{R: 255, G: 224, B: 32}
		g.cmd.draw(g.rgba, white, yellow, 48, 32)
	}
	if g.dlg.open { // 對話框
		g.dlg.draw(g.rgba, white)
	}
	g.input.touch.draw(g.rgba) // 觸控控制疊在最上層(有觸控過才顯示)
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

// loader 依序讀資產;第一個錯誤即記下,後續讀取短路(讓 NewGame 集中回報一個 error)。
type loader struct {
	fsys fs.FS
	err  error
}

func (l *loader) read(name string) []byte {
	if l.err != nil {
		return nil
	}
	d, e := fs.ReadFile(l.fsys, name)
	if e != nil {
		l.err = fmt.Errorf("讀 %s: %w", name, e)
	}
	return d
}

// NewGame 從 assets 檔案系統載入所有資產、組出可執行的 Game(ebiten.Game)。
// assets:原版素材(桌面=os.DirFS(DQ3_ASSETS)、行動=embed 子樹)。
// music:放 track_NN.ogg 的檔案系統(nil → 靜音)。
// 回傳的 *Game 直接餵 ebiten.RunGame(桌面)或 mobile.SetGame(行動)。
func NewGame(assets fs.FS, music fs.FS) (*Game, error) {
	ld := &loader{fsys: assets}
	pal := dq3data.DecodePalette(ld.read("DQ3.PAL"), 256) // 城鎮/地表共用(tile 只用色 0..15)
	g := &Game{rgba: make([]byte, ScreenW*ScreenH*4), input: newInput()}

	// 地表 scene
	overBLK, err := dq3data.OpenBLK(ld.read("DQ3.BLK"))
	if err != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenBLK 地表: %w", err)
	}
	fm, err := dq3data.OpenFieldMap(ld.read("DQ3CON.MAP"))
	if err != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenFieldMap: %w", err)
	}
	if ld.err != nil {
		return nil, ld.err
	}
	g.over = &Scene{
		blk: overBLK, attr: dq3data.OpenBlockAttr(ld.read("BLKBM.DAT")), pal: pal,
		w: fm.W, h: fm.H, tileAt: fm.Tile,
	}

	// 阿里阿罕城 scene(CTY00.DAT section 0 / DQ31.BLK + BLKBM1.DAT)
	townBLK, err := dq3data.OpenBLK(ld.read("DQ31.BLK"))
	if err != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenBLK 城鎮: %w", err)
	}
	tw, err := dq3data.OpenTown(ld.read("CTY00.DAT"), 0)
	if err != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenTown 阿里阿罕: %w", err)
	}
	if ld.err != nil {
		return nil, ld.err
	}
	g.town = &Scene{
		blk: townBLK, attr: dq3data.OpenBlockAttr(ld.read("BLKBM1.DAT")), pal: pal,
		w: tw.W, h: tw.H, tileAt: tw.Tile,
		spawnX: tw.SpawnX, spawnY: tw.SpawnY,
	}
	// NPC sprite:DQ3MAN.BLS,entry_base=(b2-4)*4;同 b2 共用一份(移植 dq3_scene_load_npc_sprites)
	manBLS := ld.read("DQ3MAN.BLS")
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
		g.town.npcs = append(g.town.npcs, npcInst{x: n.X, y: n.Y, ctrl: n.Ctrl, b4: n.B4, spr: spr})
	}

	// 對話 + 命令窗:字型 D3TXT00.FON(常駐)+ 阿里阿罕 bank 1(D3TXT01.TXT,section dlg_bank=1)
	g.dlg.tx = dq3data.LoadText(ld.read("D3TXT00.FON"), ld.read("D3TXT01.TXT"))
	g.cmd.tx = g.dlg.tx // 命令窗標籤 glyph 也走同一字型
	g.hero = dq3data.LoadCharSprite(ld.read("DQ3MST.BLS"), 0)

	// 戰鬥:怪物數值(D3MNS.DAT)+ sprite(DQ3MNS.SHP)+ 怪物色盤(MNSBK.PAL)
	mons, merr := dq3data.OpenMonsters(ld.read("D3MNS.DAT"))
	if merr != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenMonsters: %w", merr)
	}
	g.battle.mons = mons
	g.battle.shp = ld.read("DQ3MNS.SHP")
	g.battle.mpal = dq3data.DecodePalette(ld.read("MNSBK.PAL"), 256)
	g.battle.tx = g.dlg.tx
	if ld.err != nil {
		return nil, ld.err
	}

	g.cur = g.over
	g.px, g.py = g.over.w/2, g.over.h/2 // 地表起點(暫用中心)
	applyDebugEnv(g)

	// MT-32 音樂(music fs 為 nil → 靜音降級)
	g.music = gaudio.NewMusic(music)
	if g.inTown {
		g.music.Play(trackCastle)
	} else {
		g.music.Play(trackField)
	}
	g.frame = ebiten.NewImage(ScreenW, ScreenH)
	g.renderFrame() // 首幀
	log.Printf("地表 %d×%d / 阿里阿罕 %d×%d(spawn %d,%d、NPC %d)— 方向鍵走、Space 對話/選定、Esc 取消/出城、Enter 進城",
		g.over.w, g.over.h, g.town.w, g.town.h, tw.SpawnX, tw.SpawnY, len(g.town.npcs))
	return g, nil
}

// applyDebugEnv:headless 截圖用的除錯環境變數(真機/正常執行無設 → 全 no-op)。
func applyDebugEnv(g *Game) {
	if os.Getenv("DQ3_START") == "town" {
		g.cur, g.inTown = g.town, true
		g.px, g.py = g.town.spawnX, g.town.spawnY
	}
	if r := os.Getenv("DQ3_DLG"); r != "" {
		var rec int
		if _, e := fmt.Sscanf(r, "%d", &rec); e == nil {
			g.dlg.Open(rec)
		}
	}
	if os.Getenv("DQ3_CMD") != "" {
		g.cmd.Open()
	}
	if os.Getenv("DQ3_TOUCH") != "" {
		g.input.touch.everTouched = true
	}
	if r := os.Getenv("DQ3_BATTLE"); r != "" { // debug:起手開一場戰鬥(截圖驗證)
		var id int
		if _, e := fmt.Sscanf(r, "%d", &id); e == nil {
			g.battle.start(id, 1)
		}
	}
}
