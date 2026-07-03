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
	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gaudio"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 場景配樂軌號(對齊 C dq3_audio g_scene_track):地表 FIELD=6、阿里阿罕城 CASTLE=1。
const (
	trackField  = 6
	trackCastle = 1
)

const (
	ScreenW, ScreenH = 640, 350
	TileW, TileH     = 32, 24
	ViewCols         = ScreenW / TileW               // 20(對齊 C VIEW_COLS)
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
	dlgText        *dq3data.Text      // 該城對話 bank(D3TXT0<bank>.TXT;地表 nil)
	hiMap          []byte             // 每格高 byte(事件 subid)
	events         [][3]int           // section 事件表 {type,param,p2}
	transitions    [][4]int           // section 轉場表 {destCty,destSec,x,y}
	sec            int                // section 號
	npcs           []npcInst
	override       map[int]int // 執行期 tile 覆蓋(開門後 door→通行 tile;key=y*w+x)
	npcRng         rng.RNG     // NPC 遊走用確定性 RNG(依 section 種子)
	night          bool        // 此 scene 用黑夜 NPC 表載入(晝夜 cache 判斷用;overworld 恆 false)
}

// tileIdx:回 (x,y) 的 BLK tile 索引,先查執行期覆蓋(開門)再退靜態 tileAt。
func (sc *Scene) tileIdx(x, y int) int {
	if sc.override != nil {
		if v, ok := sc.override[y*sc.w+x]; ok {
			return v
		}
	}
	return sc.tileAt(x, y)
}

// doorTier:回 (tx,ty) 門所需鑰匙等級(attr&0xc0 → (a>>6)&3);非門回 0。移植 dq3_scene_door_tier。
func (sc *Scene) doorTier(tx, ty int) int {
	if tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return 0
	}
	a := sc.attr.Raw(sc.tileIdx(tx, ty))
	if a&0x00c0 == 0 {
		return 0
	}
	return int(a>>6) & 3
}

// openDoor:開 (tx,ty) 門 → tile 覆蓋成 hiMap&0x1f(通行 tile),清事件位留轉場位。移植 dq3_scene_open_door。
func (sc *Scene) openDoor(tx, ty int) bool {
	if sc.doorTier(tx, ty) == 0 {
		return false
	}
	i := ty*sc.w + tx
	var hi byte
	if sc.hiMap != nil {
		hi = sc.hiMap[i]
	}
	if sc.override == nil {
		sc.override = map[int]int{}
	}
	sc.override[i] = int(hi & 0x1f) // door → 通行 tile
	if sc.hiMap != nil {
		sc.hiMap[i] = hi & 0xe0 // 清事件位、留轉場位(0x4985)
	}
	return true
}

// tileTransition:回 (tx,ty) 的轉場(attr&0xe000 轉場格 → hiMap 低5bit subid → transitions[subid])。
// 回 (destCty, destSec, dx, dy, ok)。移植 dq3_scene_tile_transition。
func (sc *Scene) tileTransition(tx, ty int) (int, int, int, int, bool) {
	if sc.hiMap == nil || len(sc.transitions) == 0 || tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return 0, 0, 0, 0, false
	}
	if sc.attr.Raw(sc.tileIdx(tx, ty))&0xe000 == 0 { // 非轉場格
		return 0, 0, 0, 0, false
	}
	subid := int(sc.hiMap[ty*sc.w+tx] & 0x1f)
	if subid >= len(sc.transitions) {
		return 0, 0, 0, 0, false
	}
	t := sc.transitions[subid]
	return t[0], t[1], t[2], t[3], true
}

// tileEvent:回 (tx,ty) 的事件(attr&8 事件格 → hiMap 低5bit subid → events[subid])。回 (event, subid, ok)。
func (sc *Scene) tileEvent(tx, ty int) ([3]int, int, bool) {
	if sc.hiMap == nil || len(sc.events) == 0 || tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return [3]int{}, 0, false
	}
	i := ty*sc.w + tx
	if sc.attr.Raw(sc.tileIdx(tx, ty))&0x0008 == 0 { // 非事件格
		return [3]int{}, 0, false
	}
	subid := int(sc.hiMap[i] & 0x1f)
	if subid >= len(sc.events) {
		return [3]int{}, 0, false
	}
	return sc.events[subid], subid, true
}

// Blocked:出界 / BLKBM 屬性 bit0=1 / 有 NPC 佔格 → 不可走(移植 dq3_scene_walkable + npc_at)。
func (sc *Scene) Blocked(x, y int) bool {
	if x < 0 || y < 0 || x >= sc.w || y >= sc.h {
		return true
	}
	if sc.attr.Blocked(sc.tileIdx(x, y)) {
		return true
	}
	return sc.npcAt(x, y) >= 0 // NPC 擋路(對齊 dq3_scene_npc_at)
}

// startPos:城鎮進入位置。header spawn 越界或不可走時,退回 pick_open_start
// (掃 9×9 窗可走鄰格最多的格),移植 dq3_scene_pick_open_start——修 6 個城 sec0 spawn 出界
// (CTY30/37/49達瑪/65/73/76)導致進城掉地圖外的 bug(docs/data/map-parse-audit.md #1)。
func (sc *Scene) startPos() (int, int) {
	walk := func(x, y int) bool {
		return x >= 0 && y >= 0 && x < sc.w && y < sc.h && !sc.attr.Blocked(sc.tileIdx(x, y))
	}
	if walk(sc.spawnX, sc.spawnY) { // header spawn 有效 → 直接用
		return sc.spawnX, sc.spawnY
	}
	bx, by, best := sc.w/2, sc.h/2, -1
	for y := 4; y < sc.h-4; y++ {
		for x := 4; x < sc.w-4; x++ {
			if !walk(x, y) {
				continue
			}
			open := 0
			for wy := -4; wy <= 4; wy++ {
				for wx := -4; wx <= 4; wx++ {
					if walk(x+wx, y+wy) {
						open++
					}
				}
			}
			if open > best {
				best, bx, by = open, x, y
			}
		}
	}
	return bx, by
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
	over, town     *Scene // 地表 / 目前城鎮
	under          *Scene // 下層地表(DQ3UND.MAP,懶載;共用地表 BLK/attr/pal)
	layer          int    // 目前地表層:0=地面 1=下層(城鎮進出以此決定回哪層)
	cur            *Scene
	inTown         bool
	curCty         int   // 目前所在 CTY 號(-1=地表)
	dnPhase        int   // 晝夜相位:0白天 1黃昏 2黑夜 3黎明(僅地表走動推進;城內固定)
	dnStep         int   // 地表步數計數器(每 dnPhaseSteps 步推進一相位)
	assets         fs.FS // 素材(懶載其他城鎮)
	worldPal       []dq3data.Color
	manBLS         []byte         // NPC sprite 來源
	towns          map[int]*Scene // 已載入城鎮快取(cty→Scene)
	overPx, overPy int            // 記住進城前的地表座標(Esc 回來用)
	hero           *dq3data.CharSprite
	px, py         int // 主角在 cur 內的 tile 座標
	facing         int // 0..3
	walk           int // 0/1 走路動畫相位
	cd, anim       int
	dlg            Dialogue  // 對話視窗
	cmd            CmdMenu   // 野外命令窗
	battle         Battle    // 戰鬥場景
	shop           Shop      // 商店(武防/道具店)
	tavern         Tavern    // 露易達酒館(招募)
	panel          panelKind // 資訊面板(狀況/道具/裝備)
	panelCursor    int       // 裝備面板游標
	panelHits      hitList   // 道具/裝備清單可點區塊(drawItems/drawEquip 重建;panelStatus 無列表不使用)
	inventory      []int     // 持有道具 id
	music          *gaudio.Music
	input          *Input // 抽象輸入(鍵盤 + 觸控)
	showTitle      bool   // 標題畫面(含主選單/主角創建流程進行中;false=已進入一般遊戲)
	titlePix       []uint8
	titlePal       []dq3data.Color
	cfg            config.Config // 可攜設定(RNG/音樂/音量/音源/戰鬥資訊/受傷特效);NewGame 用 config.Default() 初始化
	settings       Settings      // 設定選單 modal(標題畫面按 S 開)
	newGame        NewGameFlow   // 標題主選單 + 主角命名/性別創建 modal(FLOW-GAP A2/A3/A4)
	heroName       []int         // 主角姓名(glyph index,注音/英數命名輸入結果;空=尚未創建/debug 略過)
	heroGender     int           // 主角性別(0=男 1=女)
	// 主角進度(勇者 class0):累積經驗 → 等級 → 屬性(heroStats);HP 跨戰鬥持久
	heroExp        uint32
	heroGold       int
	heroHP         int
	heroMP         int
	heroInit       bool
	equip          [4]int       // 裝備槽:0 武器 1 鎧 2 盾 3 兜(item code;0=空)
	companions     []*Member    // 同伴(隊長=hero*;酒館招募,現為預建示範)
	flags          map[int]bool // 一次性旗標(寶箱/事件已取;remake 里程碑,如 0x211/0x213)
	storyBits      [32]byte     // 原版 [0x4f70] 256-flag story-flag 陣列(NPC 可見性;初值見 initStoryBits)
	noticeCode     int          // 取得道具通知(item code;-1=無)
	noticeTimer    int
	shipOwned      bool          // 已取得船(波魯多加胡椒換船,milestone SHIP)
	shipAboard     bool          // 目前在船上
	shipX, shipY   int           // 船停泊位置(地表)
	repel          int           // 聖水驅敵剩餘步數(>0 期間地表不遇弱敵,每步遞減)
	prng           rng.RNG       // 祈禱之戒損壞判定用 RNG
	lotoBlessed    bool          // 破索瑪後洛特冊封(勇者裝備昇華為傳說的洛特裝備)
	cleared        bool          // 已破關(索瑪擊破)
	endText        *dq3data.Text // ENDTXT.TXT 結局文本
	endSeq         int           // 結局捲動段號(-1=未進行;0..n=當前段)
	openingIdx     int           // 開場旁白序號(-1=未進行;0..n=當前句)
	bossQueue      []int         // boss 連戰佇列(索瑪神殿:六大魔人→怨靈→殭屍→索瑪;或 examine 觸發的座標 boss 鏈)
	pendingTrigger *bossTrigger  // 目前由 examine 觸發、正在跑的座標 boss 鏈(非 nil 期間 onBattleEnd 需結算);完成或中斷即清 nil
	frame          *ebiten.Image
	rgba           []byte
}

// 場景配樂軌(對齊 C g_scene_track):BATTLE=14、TOWN=2、DUNGEON=3。
const (
	trackBattle  = 14
	trackTown    = 2
	trackDungeon = 3
)

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
	switch cmd {
	case cmdTalk:
		fx, fy := frontTile(g.px, g.py, g.facing)
		if idx := g.cur.npcAt(fx, fy); idx >= 0 {
			n := &g.cur.npcs[idx]
			switch sub := (n.ctrl >> 3) & 7; {
			case sub <= 1: // 對話
				g.dlg.Open(n.b4)
			case sub == 2: // scripted NPC(給物/劇情)
				if g.curCty == ctyPortoga && n.x == portogaKingX && n.y == portogaKingY {
					// 波魯多加國王(main.c:1607 位置特例,先於 scripted 表判定)
					g.talkPortogaKing()
				} else {
					g.scriptedTalk(n.b4)
				}
			case sub >= 3: // 設施(店/宿/教會)
				g.openFacility(n.b4)
			}
		}
	case cmdStatus: // 狀況
		g.panel = panelStatus
	case cmdItem: // 道具
		g.panel, g.panelCursor = panelItem, 0
	case cmdEquip: // 裝備
		g.panel, g.panelCursor = panelEquip, 0
	case cmdExamine: // 調査:檢查面向格 → 寶箱/隱藏物(一次性旗標)
		g.examine()
	}
}

func (g *Game) hasItem(code int) bool { return g.countItem(code) > 0 }

// shipTileFor:我方 facing(0下1上2左3右)→ 船 BLK tile(移植 dq3_ship_tile_for_facing)。
func shipTileFor(facing int) int {
	switch facing {
	case 1:
		return 55 // 上 UP
	case 2:
		return 53 // 左 LEFT
	case 3:
		return 57 // 右 RIGHT
	}
	return 51 // 下 DOWN
}

// tryMove:判定能否移動到 (nx,ny)(含船系統)。移植 dq3_ship_input + 碰撞。
func (g *Game) tryMove(nx, ny int) bool {
	if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h {
		return false
	}
	if g.shipAboard { // 在船上:海(attr&0x20)可航;走上可走陸地 → 下船
		if g.cur.attr.Raw(g.cur.tileIdx(nx, ny))&0x20 != 0 {
			return true // 航行至水格
		}
		if !g.cur.Blocked(nx, ny) { // 上岸下船(船留在原水格)
			g.shipAboard = false
			g.shipX, g.shipY = g.px, g.py
			return true
		}
		return false
	}
	// 步行:走到船格 → 上船
	if g.shipOwned && !g.inTown && nx == g.shipX && ny == g.shipY {
		g.shipAboard = true
		return true
	}
	return !g.cur.Blocked(nx, ny)
}

// scriptedTalk:對 scripted NPC(sub2)—— 檢查型/給予型(前置道具→給物+旗標+對白)。移植 dq3_scripted。
func (g *Game) scriptedTalk(byte4 int) {
	s := scriptedFor(byte4, g.curCty)
	if s == nil {
		return
	}
	give, prereq, require, consume, milestone := s[2], s[3], s[4], s[5], s[6]
	beforeRec, giveRec, afterRec := s[7], s[8], s[9]
	if byte4 == 25 && g.curCty == 15 && !g.flags[0x211] {
		// 古布達黑胡椒救人 gate(main.c:1736):CTY15 byte4=25 需先救出達妮亞
		// (flag 0x211,甘達特巢穴 boss 鏈勝利設,bosstrigger.go)才給黑胡椒。
		g.dlg.Open(beforeRec)
		return
	}
	if require != 255 { // 檢查型:持物→give_rec、否則→before_rec(不消耗)
		if g.hasItem(require) {
			g.dlg.Open(giveRec)
		} else {
			g.dlg.Open(beforeRec)
		}
		return
	}
	if milestone != 0 && g.flags[milestone] { // 已給過 → 後話
		g.dlg.Open(afterRec)
		return
	}
	if prereq != 255 && !g.hasItem(prereq) { // 缺前置道具
		g.dlg.Open(beforeRec)
		return
	}
	if give != 255 && g.hasItem(give) { // 已持有 give_item → 不重給(對齊 C dq3_inv_find<0;修 milestone==0 列無限複製)
		g.dlg.Open(afterRec)
		return
	}
	if give != 255 { // 給物 + 通知
		g.inventory = append(g.inventory, give)
		g.noticeCode, g.noticeTimer = give, 120
	}
	if consume == 1 && prereq != 255 {
		g.removeItems(prereq, 1)
	}
	if milestone != 0 {
		g.flags[milestone] = true
	}
	g.dlg.Open(giveRec)
	if give != 255 { // giveRec 若含 VAR_ITEM(0xfffa)→ 填實際道具名(Open 已重置 var,須在其後設)
		g.setDlgVarItem(give)
	}
}

// 取船劇情常數(main.c:57-90 DQ3_PORTOGA_*/DQ3_ITEM_PEPPER/DQ3_PORTOGA_REC_*,座標/rec/碼原封對齊)。
const (
	ctyPortoga     = 37   // DQ3_PORTOGA_CTY
	portogaKingX   = 9    // DQ3_PORTOGA_KING_X
	portogaKingY   = 6    // DQ3_PORTOGA_KING_Y
	portogaHarborX = 25   // DQ3_PORTOGA_HARBOR_X(港口海格,城南鄰)
	portogaHarborY = 73   // DQ3_PORTOGA_HARBOR_Y
	itemPepper     = 0x5c // DQ3_ITEM_PEPPER(黑胡椒,獻國王換船)
	recPortogaWait = 26   // DQ3_PORTOGA_REC_WAIT「怎麼了?我在等黑胡椒。」
	recPortogaGot  = 28   // DQ3_PORTOGA_REC_GOT「胡椒太好吃了…好想睡。」
)

// talkPortogaKing:波魯多加國王(CTY37 (9,6))授船特例,移植 main.c:1607-1631。
// 已取船 → 仍顯示「吃胡椒」後話;持黑胡椒 → 獻上換船(消耗胡椒、停泊港口、設 msShip);
// 否則 → 「等胡椒」對白。國王的信(0x5b/flag 0x215)支線不在取船鏈範圍內,未移植。
func (g *Game) talkPortogaKing() {
	switch {
	case g.shipOwned:
		g.dlg.Open(recPortogaGot)
	case g.hasItem(itemPepper):
		g.removeItems(itemPepper, 1)
		g.shipOwned, g.shipAboard = true, false
		g.shipX, g.shipY = portogaHarborX, portogaHarborY
		g.progressSet(msShip)
		g.dlg.Open(recPortogaGot)
	default:
		g.dlg.Open(recPortogaWait)
	}
}

// examine:調查面向格(或腳下)的事件 → 座標 boss 觸發點 → 開戰;否則寶箱/warp。
// 移植 dq3_treasure + main.c 座標 boss 特例(八頭大蛇/甘達特巢穴,bosstrigger.go)。
func (g *Game) examine() {
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	checkBoss := func(x, y int) bool {
		t := bossTriggerAt(g.curCty, g.cur.sec, x, y)
		if t == nil || g.flags[t.doneFlag] { // 未收錄 / 已擊敗過 → 不再開戰
			return false
		}
		g.pendingTrigger = t
		g.bossQueue = append([]int(nil), t.monsters...)
		g.advanceBossQueue()
		return true
	}
	check := func(x, y int) bool {
		ev, subid, ok := g.cur.tileEvent(x, y)
		if !ok {
			return false
		}
		if ev[0] == 2 { // 事件 type-2 → scripted 傳送(warps[param])
			return g.warpTo(ev[1])
		}
		t := treasureFor(g.curCty, g.cur.sec, subid)
		if t == nil {
			return false
		}
		flag := t[5]
		if g.flags[flag] { // 已取
			return true
		}
		g.flags[flag] = true
		if t[3] == 1 || t[3] == 3 { // 寶箱給道具
			g.inventory = append(g.inventory, t[4])
			g.noticeCode, g.noticeTimer = t[4], 120
		}
		return true
	}
	if g.inTown && g.curCty == shrineCty && g.synthRainbowAtShrine() { // 合成祠堂:太陽之石+雲雨之杖→彩虹水滴
		return
	}
	fx, fy := frontTile(g.px, g.py, g.facing)
	if g.tryOpenFacingDoor(fx, fy) { // 面向鎖門 + 鑰匙足夠 → 開門(移植 dq3_scene_try_open_facing_door)
		return
	}
	if checkBoss(fx, fy) || checkBoss(g.px, g.py) { // 座標 boss 觸發點優先於一般寶箱/warp
		return
	}
	if !check(fx, fy) {
		check(g.px, g.py) // 也試腳下
	}
}

// keyTier:背包最高階鑰匙等級(道具碼 0x55/0x56/0x57 → 1/2/3;tier=code-0x54)。移植 dq3_inv_key_tier。
func (g *Game) keyTier() int {
	best := 0
	for _, code := range g.inventory {
		if code >= 0x55 && code <= 0x57 {
			if t := code - 0x54; t > best {
				best = t
			}
		}
	}
	return best
}

// tryOpenFacingDoor:面向格是鎖門且鑰匙等級足夠 → 開門。移植 dq3_scene_try_open_facing_door。
func (g *Game) tryOpenFacingDoor(fx, fy int) bool {
	need := g.cur.doorTier(fx, fy)
	if need == 0 || g.keyTier() < need {
		return false
	}
	return g.cur.openDoor(fx, fy)
}

// openFacility:面向設施 NPC(byte4=設施索引 k)→ 依當前 CTY 開對應設施(全城設施表)。
func (g *Game) openFacility(k int) {
	f := facilityForCty(g.curCty, k)
	if f == nil {
		return
	}
	switch f.typ {
	case facInn: // 旅社:扣費(inn_cost×人數,單人)+ 治滿 HP/MP
		_, maxHP, _, _, _ := g.heroStats()
		if g.heroGold >= f.innCost {
			g.heroGold -= f.innCost
			g.heroHP, g.heroMP, g.heroInit = maxHP, g.heroMaxMP(), true
		}
	case facWeapon, facItem: // 商店:開貨架(品項自全城品項池 itemOff..+count)
		lo, hi := f.itemOff, f.itemOff+f.count
		if lo >= 0 && hi <= len(shopItemPool) {
			g.shop.open(append([]int(nil), shopItemPool[lo:hi]...))
		}
	case facChurch, facRecord: // 教會/記錄點:存檔(冒險之書)
		_ = g.Save()
	}
}

// updateTouchContext:依當前狀態設觸控情境鍵標籤(空=隱藏)。拆成獨立方法方便單元測試
// (不必跑完整 Update()/RunGame,只驗證「狀態 → 標籤」這條轉移邏輯本身)。
func (g *Game) updateTouchContext() {
	switch {
	case g.showTitle && g.newGame.stage == ngName:
		g.input.touch.SetContext("注") // 主角命名(英數↔注音切換,同酒館命名)
	case g.showTitle:
		g.input.touch.SetContext("設定")
	case g.tavern.active && g.tavern.stage == tavName:
		g.input.touch.SetContext("注")
	default:
		g.input.touch.SetContext("")
	}
}

func (g *Game) Update() error {
	g.updateTouchContext() // 必須在 Poll() 之前,讓這幀 poll 用得到新標籤
	in := g.input.Poll()   // 抽象輸入(鍵盤 + 觸控合流)
	moved := false

	// 標題畫面:主選單(遊戲開始/載入進度)→ 主角命名(必填)→ 性別 → 開始新遊戲(newgame.go)。
	// S/CtxTap 開設定選單(疊在標題上,ESC/Cancel 關閉回標題)。
	if g.showTitle {
		if g.settings.open {
			g.settingsInput(in)
			g.renderFrame()
			return nil
		}
		g.newGameInput(in)
		g.renderFrame()
		return nil
	}

	// 戰鬥 modal:結束時寫回主角結果 + 回地表音樂
	if g.battle.active {
		if g.battle.input(in) {
			won := g.battle.result == 1
			g.onBattleEnd()
			if won && len(g.bossQueue) > 0 { // boss 連戰:勝 → 接下一場
				g.advanceBossQueue()
			} else if g.endSeq < 0 { // 一般戰後回地表曲(結局進行中則不覆蓋)
				g.music.Play(trackField)
			}
		}
		g.renderFrame()
		return nil
	}
	// 酒館 modal:職業→性別→招募加入隊伍(最多 3 同伴,滿則替換最舊)
	if g.tavern.active {
		if m, _ := g.tavern.input(in); m != nil {
			if len(g.companions) < 3 {
				g.companions = append(g.companions, m)
			} else {
				g.companions = append(g.companions[1:], m)
			}
		}
		g.renderFrame()
		return nil
	}
	// T 鍵:阿里阿罕酒館招募(對齊 C dev 捷徑;正式入口=LUIDA 櫃台)
	if g.inTown && inpututil.IsKeyJustPressed(ebiten.KeyT) {
		g.tavern.open()
		g.renderFrame()
		return nil
	}
	// U 鍵:降到下層地表(對齊 C dev 捷徑;正式觸發=劇情事件 86)。在地表時可用
	if !g.inTown && g.layer == 0 && inpututil.IsKeyJustPressed(ebiten.KeyU) {
		g.descend()
		return nil
	}
	// Z 鍵:索瑪神殿最終連戰(六大魔人→怨靈→殭屍→索瑪;對齊 C dev 捷徑 zomaseq)。持光之珠弱化索瑪
	if !g.battle.active && inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		g.startZomaSeq()
		g.renderFrame()
		return nil
	}
	// 資訊面板 modal:狀況/道具 = B/A 關;裝備 = 方向選 + A 裝上 + B 關
	// 點列(P2,道具/裝備清單)= 游標移過去 + 等同 A 使用/裝上
	if g.panel != panelNone {
		tapIdx := -1
		if in.Tapped {
			tapIdx = g.panelHits.at(in.TapX, in.TapY)
		}
		switch {
		case in.Cancel:
			g.panel = panelNone
		case g.panel == panelEquip:
			confirm := in.Confirm
			if tapIdx >= 0 && len(g.inventory) > 0 {
				g.panelCursor, confirm = tapIdx, true
			}
			switch {
			case confirm:
				g.equipSelected()
			case in.DirEdge == 0 && len(g.inventory) > 0: // 下
				g.panelCursor = (g.panelCursor + 1) % len(g.inventory)
			case in.DirEdge == 1 && len(g.inventory) > 0: // 上
				g.panelCursor = (g.panelCursor + len(g.inventory) - 1) % len(g.inventory)
			}
		case g.panel == panelItem: // 道具:方向選 + A 使用(藥草/翼/聖水/祈禱之戒)
			confirm := in.Confirm
			if tapIdx >= 0 && len(g.inventory) > 0 {
				g.panelCursor, confirm = tapIdx, true
			}
			switch {
			case confirm:
				g.useSelectedItem()
			case in.DirEdge == 0 && len(g.inventory) > 0: // 下
				g.panelCursor = (g.panelCursor + 1) % len(g.inventory)
			case in.DirEdge == 1 && len(g.inventory) > 0: // 上
				g.panelCursor = (g.panelCursor + len(g.inventory) - 1) % len(g.inventory)
			}
		case in.Confirm:
			g.panel = panelNone
		}
		g.renderFrame()
		return nil
	}
	// 商店 modal:方向選、A 買、B 關
	if g.shop.active {
		if buy, _ := g.shop.input(in); buy >= 0 {
			if price := g.shop.items.Price(buy); g.heroGold >= price {
				g.heroGold -= price
				g.inventory = append(g.inventory, buy)
			}
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
	// 結局捲動:一段 ENDTXT 翻完關閉後,自動接下一段,直到全部播完(移植 main.c end_seq 推進)。
	if g.endSeq >= 0 {
		g.endSeq++
		if g.endText != nil && g.endSeq < g.endText.NRecords {
			g.dlg.Open(g.endSeq)
		} else {
			g.endSeq = -1 // 結局播完
		}
		g.renderFrame()
		return nil
	}
	// 開場旁白:一句翻完 → 下一句(rec82→83,母親帶去見國王,docs/66 §1d)→ 播完母親帶出門到城鎮。
	if g.openingIdx >= 0 {
		g.openingIdx++
		if g.openingIdx < len(openingSeq) {
			g.dlg.Open(openingSeq[g.openingIdx])
		} else {
			g.openingIdx = -1
			g.motherEscort() // 母親帶你去見國王 → 出家門到阿里阿罕城鎮(sec0)
		}
		g.renderFrame()
		return nil
	}
	// 命令窗 modal:方向移游標、A 選定、B 關窗;點格(P2)= 游標移過去 + 等同 A 選定
	if g.cmd.open {
		confirm := in.Confirm
		if in.Tapped {
			if idx := g.cmd.hits.at(in.TapX, in.TapY); idx >= 0 {
				g.cmd.cursor, confirm = idx, true
			}
		}
		switch {
		case in.Cancel:
			g.cmd.open = false
		case confirm:
			g.music.PlaySFX(0) // 選單確定音效(VOC)
			g.selectCommand(g.cmd.cursor)
		default:
			if in.DirEdge >= 0 {
				g.cmd.move(in.DirEdge)
			}
		}
		g.renderFrame()
		return nil
	}
	// A:開命令窗(地表/城內皆可,DQ 風格)
	if in.Confirm {
		g.cmd.Open()
		g.renderFrame()
		return nil
	}
	if in.Enter && !g.inTown { // 鍵盤 Enter:debug 直接進阿里阿罕
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
		if g.tryMove(nx, ny) { // 碰撞/航行:陸走 BLKBM+NPC、船走海、上/下船
			g.px, g.py = nx, ny
			moved = true
		}
		g.cd = moveCooldown
	}
	if moved && g.inTown { // 城內:踩到轉場格(門/階梯/出城)→ 切 section / 跨 CTY / 出城
		g.tryTransition()
	} else if moved && !g.inTown { // 地表:踩到城鎮入口 → 進城;否則隨機遇敵(~1/16 步)
		cty := findCtyAtLayer(g.px, g.py, g.layer)               // 依目前地表層找入城
		if pc := owPortalResolve(g.px, g.py, g.flags); pc >= 0 { // 旗標條件 portal 覆蓋(同點依進度變城)
			cty = pc
		}
		if cty >= 0 {
			g.enterTownCty(cty)
			return nil
		}
		g.advanceDaynight() // 地表走一步 → 推進晝夜(每 60 步一相位;城內不推進)
		if g.repel > 0 {    // 聖水驅敵期間每步遞減、不遇弱敵(移植 #3 repel)
			g.repel--
		} else if rand.Intn(16) == 0 {
			g.startEncounter()
		}
	}
	if g.noticeTimer > 0 {
		g.noticeTimer--
	}
	g.npcTick() // 城鎮 NPC 每幀遊走(地表/戰鬥/對話 modal 已提早 return,不會跑到這)
	g.anim++
	if moved || g.anim%18 == 0 { // 走動 / 待機都擺手腳
		if g.anim%9 == 0 {
			g.walk ^= 1
		}
	}
	g.renderFrame() // ★ 在 Update 渲染 → WritePixels 於 Draw 前生效
	return nil
}

// enterTownCty:進 CTY 城鎮(懶載 + 快取)。移植 apply_scene 的城鎮載入。
func (g *Game) enterTownCty(cty int) {
	sc := g.towns[cty]
	if sc != nil && sc.night != g.isNight() {
		sc = nil // 進城相位與 cache 不同(日/夜 NPC 表不同)→ 重載對齊當下相位
	}
	if sc == nil {
		blkn := 1
		if cty >= 0 && cty < len(mapBlkNum) {
			blkn = mapBlkNum[cty]
		}
		s, err := loadTownScene(g.assets, g.worldPal, g.manBLS, cty, blkn, g.dnPhase, g.storyFlag)
		if err != nil {
			return // 載入失敗 → 留在地表
		}
		g.towns[cty] = s
		sc = s
	} else {
		sc.pal = dq3data.DarkenPalette(g.worldPal, g.dnPhase) // cache 命中(同 night):重套相位色(黃昏/黎明差異)
	}
	g.overPx, g.overPy = g.px, g.py
	g.town, g.cur, g.inTown, g.curCty = sc, sc, true, cty
	switch cty { // 進城推進主線里程碑(移植 main.c cur_cty 判定)
	case 2: // 羅馬利亞
		g.progressSet(msRomaly)
	case 49: // 達瑪神殿
		g.progressSet(msDhama)
	}
	g.px, g.py = sc.startPos() // spawn 出界/不可走 → pick_open_start(修 6 城進城掉圖外)
	if sc.dlgText != nil {     // 切到該城對話 bank(NPC 對話用正確 bank)
		g.dlg.tx = sc.dlgText
	}
	g.cd = moveCooldown
	g.music.Play(ctyMusicTrack(cty))
	g.renderFrame()
}

// dnPhaseSteps:每晝夜相位的地表步數。使用者指定「白天→黑夜=120 步、中間有黃昏」→ 每相位 60 步
// (白天→黃昏→黑夜=2×60;全 4 相位循環=240 步,對齊原版 clock 0..0xf0)。移植 main.c DN_PHASE_STEPS。
const dnPhaseSteps = 60

// isNight:目前是否黑夜(相位 2)。供夜 gate 事件查詢(提頓夜綠寶珠等)。對齊 dq3_scene.c night 判定。
func (g *Game) isNight() bool { return g.dnPhase == 2 }

// storyFlag / setStoryFlag:原版 [0x4f70] 256-flag story-flag 陣列(NPC 可見性等;SET=file 0x824f、
// CLR=0x8264、GET=0x8279)。新遊戲初值見 dq3data.NPCStoryInitFlags(flag 0-39 清、40-255 設)。
// 供 loadTownSceneSec 的 NPC 可見性過濾;獨立於 g.flags(remake 里程碑),不互相污染。
func (g *Game) storyFlag(id int) bool {
	if id < 0 || id >= 256 {
		return false
	}
	return g.storyBits[id>>3]&(0x80>>(uint(id)&7)) != 0
}

func (g *Game) setStoryFlag(id int, v bool) {
	if id < 0 || id >= 256 {
		return
	}
	m := byte(0x80 >> (uint(id) & 7))
	if v {
		g.storyBits[id>>3] |= m
	} else {
		g.storyBits[id>>3] &^= m
	}
}

// initStoryBits:新遊戲重置 [0x4f70] 旗標陣列為原版靜態映像初值(docs/71)。
func (g *Game) initStoryBits() { g.storyBits = dq3data.NPCStoryInitFlags }

// advanceDaynight:地表走一步 → 步數計數;每 dnPhaseSteps 步推進一相位並重套地表 palette。
// 只在地表(overworld)呼叫;城內相位固定(使用者確認)。
func (g *Game) advanceDaynight() {
	g.dnStep++
	if g.dnStep >= dnPhaseSteps {
		g.dnStep = 0
		g.dnPhase = (g.dnPhase + 1) & 3
		g.applyDaynightPalette()
	}
}

// setDaynight:強制設相位(拉那魯達/黑暗之燈用)。重套 palette;若在城內,重載當前 section
// 的 NPC(日/夜表不同,對齊 main.c:988 reload_town_daynight)。
func (g *Game) setDaynight(phase int) {
	g.dnPhase = phase & 3
	g.dnStep = 0
	g.applyDaynightPalette()
	g.reloadTownDaynight()
}

// toggleDaynight:白天↔黑夜切換(拉那魯達語意)。非黑夜 → 黑夜;黑夜 → 白天。
func (g *Game) toggleDaynight() {
	if g.isNight() {
		g.setDaynight(0)
	} else {
		g.setDaynight(2)
	}
}

// applyDaynightPalette:依當前相位由日中 base 色盤(worldPal,不變)重算調暗色盤,套到現行地表/
// 城鎮 scene。base 恆為日色 → 可反覆套用不累積。
func (g *Game) applyDaynightPalette() {
	dp := dq3data.DarkenPalette(g.worldPal, g.dnPhase)
	if g.over != nil {
		g.over.pal = dp
	}
	if g.under != nil {
		g.under.pal = dp
	}
	if g.inTown && g.town != nil {
		g.town.pal = dp
	}
}

// reloadTownDaynight:城內切晝夜時重載當前 section 的 NPC(日/夜表不同),保留座標。
// 對齊 main.c:988。不在城內 / 無效 CTY → 跳過。載入失敗 → 維持原 scene。
func (g *Game) reloadTownDaynight() {
	if !g.inTown || g.curCty < 0 || g.town == nil {
		return
	}
	blkn := 1
	if g.curCty < len(mapBlkNum) {
		blkn = mapBlkNum[g.curCty]
	}
	ns, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, g.curCty, blkn, g.town.sec, g.dnPhase, g.storyFlag)
	if err != nil {
		return
	}
	g.town, g.cur = ns, ns
	g.towns[g.curCty] = ns
	if g.px >= ns.w { // 座標保留;越界夾回(section 尺寸理應相同,防護)
		g.px = ns.w - 1
	}
	if g.py >= ns.h {
		g.py = ns.h - 1
	}
	if ns.dlgText != nil {
		g.dlg.tx = ns.dlgText
	}
	g.renderFrame()
}

// tryTransition:城內踩到轉場格(門/階梯/出城)→ 依 section+0xc 轉場表切換。移植 main.c 轉場消費。
// destSec 0xff/0xfe → 出城回地表(下層 DQ3UND 本 port 未載,退地表);其餘 → 換 section(同城)或跨 CTY。
// 玩家一律置於目的 (dx,dy);dx/dy=0 表沿用既有位置。
func (g *Game) tryTransition() {
	if !g.inTown {
		return
	}
	dcty, dsec, dx, dy, ok := g.cur.tileTransition(g.px, g.py)
	if !ok {
		return
	}
	if dsec == 0xff || dsec == 0xfe { // 出城 → 地表(0xff=地面、0xfe=下層 DQ3UND)
		g.layer = 0
		if dsec == 0xfe && g.loadUnder() != nil {
			g.layer = 1
		}
		g.exitTown()
		if dx != 0 && dx < g.cur.w {
			g.px = dx
		}
		if dy != 0 && dy < g.cur.h {
			g.py = dy
		}
		g.overPx, g.overPy = g.px, g.py
		g.renderFrame()
		return
	}
	cross := dcty != g.curCty && dcty < 100 // 跨 CTY vs 同城換 section
	ncty := g.curCty
	if cross {
		ncty = dcty
	}
	blkn := 1
	if ncty >= 0 && ncty < len(mapBlkNum) {
		blkn = mapBlkNum[ncty]
	}
	ns, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ncty, blkn, dsec, g.dnPhase, g.storyFlag)
	if err != nil {
		g.exitTown() // 載入失敗 → 回地表保底(對齊 C)
		return
	}
	g.town, g.cur, g.curCty = ns, ns, ncty
	if dx < ns.w {
		g.px = dx
	}
	if dy < ns.h {
		g.py = dy
	}
	if ns.dlgText != nil { // 切到目的城對話 bank
		g.dlg.tx = ns.dlgText
	}
	if cross { // 跨 CTY → 換配樂
		g.music.Play(ctyMusicTrack(ncty))
	}
	g.cd = moveCooldown
	g.renderFrame()
}

// warpTo:event type-2 觸發 → warps[param]={destCty,X,Y} 切場景。移植 dq3_warp_get + main.c 消費。
// 目的 CTY 無效(≥100 或載入失敗)→ 不傳送(對齊 C town_load 失敗保底)。
func (g *Game) warpTo(param int) bool {
	dcty, dx, dy, ok := warpGet(param)
	if !ok || dcty < 0 || dcty >= 100 {
		return false
	}
	blkn := 1
	if dcty < len(mapBlkNum) {
		blkn = mapBlkNum[dcty]
	}
	ns, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, dcty, blkn, 0, g.dnPhase, g.storyFlag)
	if err != nil {
		return false
	}
	if !g.inTown {
		g.overPx, g.overPy = g.px, g.py
	}
	g.town, g.cur, g.inTown, g.curCty = ns, ns, true, dcty
	if dx < ns.w {
		g.px = dx
	}
	if dy < ns.h {
		g.py = dy
	}
	if ns.dlgText != nil {
		g.dlg.tx = ns.dlgText
	}
	g.music.Play(ctyMusicTrack(dcty))
	g.renderFrame()
	return true
}

// loadUnder:懶載下層地表(DQ3UND.MAP;共用地面 DQ3.BLK/BLKBM/pal)。失敗回 nil。移植 dq3_field_load_map。
func (g *Game) loadUnder() *Scene {
	if g.under != nil {
		return g.under
	}
	raw, _ := fs.ReadFile(g.assets, "DQ3UND.MAP")
	fm, err := dq3data.OpenFieldMap(raw)
	if err != nil {
		return nil
	}
	g.under = &Scene{blk: g.over.blk, attr: g.over.attr, pal: g.over.pal, w: fm.W, h: fm.H, tileAt: fm.Tile}
	return g.under
}

// overworldScene:目前地表層對應的 Scene(下層載入失敗則退地面)。
func (g *Game) overworldScene() *Scene {
	if g.layer == 1 {
		if u := g.loadUnder(); u != nil {
			return u
		}
		g.layer = 0
	}
	return g.over
}

// descend:降到下層地表(移植 do_descent;原版由劇情事件 86 觸發,此處供 debug/U 鍵)。
// 置玩家於下層城 CTY77 入口附近(84,68)。
func (g *Game) descend() {
	u := g.loadUnder()
	if u == nil {
		return
	}
	g.layer, g.cur, g.inTown, g.curCty = 1, u, false, -1
	g.px, g.py = 84, 68
	g.progressSet(msDescend) // 下降里程碑(對 flagDescended)
	g.music.Play(trackField)
	g.renderFrame()
}

// openingSeq:開場旁白 rec(D3TXT01,阿里阿罕 bank):16歲生日引子 → 帶你去見國王(docs/66 §1d,DOSBox 實測逐字吻合)。
var openingSeq = []int{82, 83}

// startOpening:新遊戲創角完成 → 進主角家(CTY00 sec4 室內)+ 播開場旁白。移植原版開場(FLOW-GAP A5/A6)。
// 對齊 U1:原版開場是家室內 + 母親旁白,非地表中心。
func (g *Game) startOpening() {
	if g.assets == nil { // 無素材的裸 Game(單元測試純狀態驗證)→ 只走狀態轉移,不載場景
		return
	}
	home, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 4, g.dnPhase, g.storyFlag)
	if err != nil { // 載入失敗 → 退回地表中心(不卡死;g.over 可能 nil,防護)
		if g.over != nil {
			g.px, g.py = g.over.w/2, g.over.h/2
		}
		return
	}
	g.town, g.cur, g.inTown, g.curCty = home, home, true, 0
	g.px, g.py = home.spawnX, home.spawnY
	if g.px >= home.w || g.py >= home.h { // sec4 header spawn 越界 → 放中央可走格
		g.px, g.py = home.w/2, home.h/2
	}
	// dlg.tx 維持 NewGame 初始的 D3TXT01(開場旁白 rec82/83 在此 bank)
	g.openingIdx = 0
	g.dlg.Open(openingSeq[0])
	g.music.Play(ctyMusicTrack(0))
	g.renderFrame()
}

// motherEscort:開場旁白播完 → 母親帶你出家門到阿里阿罕城鎮外圍(sec0)。
// sec4 transition[0] 目的地 = sec0 家門 (8,38)。原版是對母親對話護送,此處簡化為旁白後自動出門
// (語意=母親帶去見國王;TODO 精修:改對母親 NPC 對話才護送,見 docs/66 §1c 實測)。
func (g *Game) motherEscort() {
	sec0, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 0, g.dnPhase, g.storyFlag)
	if err != nil {
		return
	}
	g.town, g.cur, g.curCty, g.inTown = sec0, sec0, 0, true
	g.px, g.py = 8, 38 // sec4 transition[0] dest(家門→城鎮外圍)
	if g.px >= sec0.w || g.py >= sec0.h {
		g.px, g.py = sec0.spawnX, sec0.spawnY
	}
	if sec0.dlgText != nil {
		g.dlg.tx = sec0.dlgText // 切到城鎮對話 bank
	}
	g.renderFrame()
}

// enterTown:進阿里阿罕(CTY0;debug/後備)。
func (g *Game) enterTown() { g.enterTownCty(0) }

func (g *Game) exitTown() {
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1 // 回目前地表層(地面/下層)
	g.px, g.py = g.overPx, g.overPy
	g.cd = moveCooldown
	g.music.Play(trackField) // 地表(FIELD 曲)
	g.renderFrame()
}

// ctyMusicTrack:CTY → 配樂軌(移植 cty_music_kind:CASTLE 清單→1、BLK2/4/5 迷宮→3、其餘城鎮→2)。
func ctyMusicTrack(cty int) int {
	for _, c := range []int{0, 2, 6, 37, 73, 76} { // 城堡(王城)
		if cty == c {
			return trackCastle
		}
	}
	if cty >= 0 && cty < len(mapBlkNum) {
		if b := mapBlkNum[cty]; b == 2 || b == 4 || b == 5 {
			return trackDungeon
		}
	}
	return trackTown
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

// heroStats 由累積經驗推導主角(勇者 class0)當前等級與戰鬥數值(移植 dq3_stats 成長表)。
// 攻 = 力量 + 武器攻;防 = 耐力/2 + 鎧/盾/兜防(裝備槽,item DB)。
func (g *Game) heroStats() (level, maxHP, atk, def, agi int) {
	level = stats.LevelForExp(0, g.heroExp)
	maxHP = int(stats.GrowthTarget(0, stats.HP, level))
	atk = int(stats.GrowthTarget(0, stats.STR, level))
	def = int(stats.GrowthTarget(0, stats.VIT, level)) / 2
	agi = int(stats.GrowthTarget(0, stats.AGI, level))
	if g.shop.items != nil { // 加裝備加成
		atk += g.shop.items.Attack(g.equip[0])
		for s := 1; s < 4; s++ {
			def += g.shop.items.Defense(g.equip[s])
		}
	}
	return
}

// heroMaxMP / heroSpells:由等級推導最大 MP 與已學可施放咒文(勇者系)。
func (g *Game) heroMaxMP() int {
	return int(stats.GrowthTarget(0, stats.MP, stats.LevelForExp(0, g.heroExp)))
}
func (g *Game) heroSpells() []int { return spell.HeroKnown(stats.LevelForExp(0, g.heroExp)) }

// startEncounter:從阿里阿罕周邊弱怪池挑一隻有 sprite 的,開戰 + 切戰鬥音樂。
func (g *Game) startEncounter() {
	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, maxMP, true
	}
	hp := heroParams{level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells()}
	comps := g.buildCompanionActors()
	pool := []int{5, 0, 1, 3, 4, 6, 7, 8} // 低階怪(id5=史萊姆…);start 會跳過空 sprite
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	for _, i := range rand.Perm(len(pool)) {
		if g.battle.start(pool[i], int64(g.anim)*2654+1, hp, comps) {
			g.music.Play(trackBattle)
			return
		}
	}
}

// startBossBattle:對指定怪 monID 開一場戰鬥(索瑪等 boss)。索瑪(0x7c)依背包光之珠設弱化旗標。
// 移植 main.c zoma 觸發 + dq3_battlescene_set_light_orb。
func (g *Game) startBossBattle(monID int) bool {
	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, maxMP, true
	}
	hp := heroParams{level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells()}
	g.battle.lightOrb = monID == 0x7c && g.hasItem(0x65) // 持光之珠 → 索瑪二階段弱化
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	if g.battle.start(monID, int64(g.anim)*2654+1, hp, g.buildCompanionActors()) {
		g.music.Play(trackBattle)
		return true
	}
	return false
}

// startZomaSeq:索瑪神殿最終連戰序列(移植 main.c zomaseq)。
// 六大魔人守衛(怪106 ×6,gate flag 0x214 已破則跳過)→ 巴拉摩斯怨靈122 → 殭屍123 → 索瑪124。
// 勝一場自動接下一場(advanceBossQueue);敗/逃則中斷。
func (g *Game) startZomaSeq() {
	q := []int{}
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	if !g.flags[0x214] { // 六大魔人守衛未破 → 六連戰
		for i := 0; i < 6; i++ {
			q = append(q, 106)
		}
	}
	q = append(q, 122, 123, 124) // 怨靈 → 殭屍 → 索瑪
	g.bossQueue = q
	g.advanceBossQueue()
}

// advanceBossQueue:取佇列下一個 boss 開戰;sprite 缺失(start 失敗)則跳過續取。佇列空則不動作。
// 六大魔人全破(pop 到第一個非106)→ 設 flag 0x214(隱藏樓梯現,重來不再打守衛)。
func (g *Game) advanceBossQueue() {
	for len(g.bossQueue) > 0 {
		id := g.bossQueue[0]
		g.bossQueue = g.bossQueue[1:]
		if id != 106 && !g.flags[0x214] { // 離開守衛段 → 標記守衛已破
			g.flags[0x214] = true
		}
		if g.startBossBattle(id) {
			return
		}
		// start 失敗(無 sprite)→ 跳過此 boss,續取下一個
	}
}

// runFinale:打倒索瑪 → 破關。設 ZOMA 里程碑 + 洛特冊封(勇者裝備昇華)。移植 run_finale。
func (g *Game) runFinale() {
	g.progressSet(msZoma)
	g.cleared = true
	g.lotoBlessed = true // そして伝説へ…:王者之劍/光之鎧甲/勇者之盾 → 洛特之劍/鎧/盾
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	g.flags[0x217] = true // 洛特冊封旗標(對齊 C)
	// 啟動 ENDTXT 結局捲動(首段;逐段推進見 Update)。
	if g.endText != nil && g.endText.NRecords > 0 {
		g.endSeq = 0
		g.dlg.tx = g.endText
		g.dlg.Open(0)
	}
}

// countItem:背包內某 item id 的數量。
func (g *Game) countItem(code int) int {
	n := 0
	for _, c := range g.inventory {
		if c == code {
			n++
		}
	}
	return n
}

// removeItems:從背包移除最多 n 個某 item id。
func (g *Game) removeItems(code, n int) {
	out := g.inventory[:0]
	for _, c := range g.inventory {
		if c == code && n > 0 {
			n--
			continue
		}
		out = append(out, c)
	}
	g.inventory = out
}

// buildCompanionActors:把同伴 Member 轉成戰鬥即時 actor(數值由職業成長表 + 裝備推導)。
func (g *Game) buildCompanionActors() []*battleActor {
	var out []*battleActor
	for _, m := range g.companions {
		out = append(out, &battleActor{
			name: classNames[m.Class], class: m.Class, level: m.Level(),
			hp: m.CurHP, maxHP: m.MaxHP(), mp: m.CurMP, maxMP: m.MaxMP(),
			atk: m.Atk(g.shop.items), def: m.Def(g.shop.items),
		})
	}
	return out
}

// onBattleEnd:戰鬥結束後把結果寫回全隊(HP/MP 持久、藥草扣除、勝利全隊加 exp/gold、升級全補、敗北回城復活)。
func (g *Game) onBattleEnd() {
	g.heroHP, g.heroMP = g.battle.heroHP, g.battle.heroMP
	for i, c := range g.battle.companions { // 同步同伴 HP/MP
		if i < len(g.companions) {
			g.companions[i].CurHP, g.companions[i].CurMP = c.hp, c.mp
		}
	}
	if g.battle.usedHerbs > 0 { // 扣掉戰鬥中用掉的藥草
		g.removeItems(herbCode, g.battle.usedHerbs)
	}
	if g.battle.result == 1 { // 勝:全隊加 exp/gold
		oldLv := stats.LevelForExp(0, g.heroExp)
		g.heroExp += uint32(g.battle.gotExp)
		g.heroGold += g.battle.gotGold
		if stats.LevelForExp(0, g.heroExp) > oldLv { // 隊長升級 → 全回復
			_, maxHP, _, _, _ := g.heroStats()
			g.heroHP, g.heroMP = maxHP, g.heroMaxMP()
		}
		for _, m := range g.companions { // 同伴同得 exp,升級全補
			oc := m.Level()
			m.Exp += uint32(g.battle.gotExp)
			if m.Level() > oc {
				m.fullHeal()
			}
		}
		if g.battle.monID == 0x7c { // 打倒大魔王索瑪 → 破關結局
			g.runFinale()
		}
	}
	if g.pendingTrigger != nil { // examine 觸發的座標 boss 鏈結算(bosstrigger.go)
		switch {
		case g.battle.result == 1 && len(g.bossQueue) == 0: // 鏈全勝(最後一場也贏)→ 給獎勵 + 設旗標
			t := g.pendingTrigger
			if len(t.rewardItems) > 0 {
				g.inventory = append(g.inventory, t.rewardItems...)
				g.noticeCode, g.noticeTimer = t.rewardItems[len(t.rewardItems)-1], 120
			}
			if g.flags == nil {
				g.flags = map[int]bool{}
			}
			g.flags[t.doneFlag] = true
			g.pendingTrigger = nil
		case g.battle.result != 1 && len(g.bossQueue) == 0: // 鏈中斷(敗/逃)→ 清空,允許重新 examine 觸發
			g.pendingTrigger = nil
		}
		// 鏈中段勝出(bossQueue 仍非空)→ 不動作,交給 Update() 的 advanceBossQueue 續打下一場
	}
	partyAllDown := g.heroHP <= 0
	for _, m := range g.companions {
		if m.CurHP > 0 {
			partyAllDown = false
		}
	}
	if partyAllDown { // 全滅:回阿里阿罕、全隊滿血復活(教會復活之簡化)
		_, maxHP, _, _, _ := g.heroStats()
		g.heroHP, g.heroMP = maxHP, g.heroMaxMP()
		for _, m := range g.companions {
			m.fullHeal()
		}
		g.cur, g.inTown = g.town, true
		g.px, g.py = g.town.spawnX, g.town.spawnY
	}
}

// renderFrame:戰鬥時畫戰鬥場景;否則畫地圖 viewport(攝影機 clamp)+ NPC + 主角 → g.frame。
func (g *Game) renderFrame() {
	if g.frame == nil { // 尚未初始化(如 NewGame 中途 debug 呼叫)→ 略過
		return
	}
	if g.showTitle && g.titlePix != nil { // 標題畫面(PCX indexed → palette)
		for i, idx := range g.titlePix {
			var c dq3data.Color
			if int(idx) < len(g.titlePal) {
				c = g.titlePal[idx]
			}
			o := i * 4
			g.rgba[o], g.rgba[o+1], g.rgba[o+2], g.rgba[o+3] = c.R, c.G, c.B, 255
		}
		white := dq3data.Color{R: 255, G: 255, B: 255}
		if len(g.titlePal) > 15 {
			white = g.titlePal[15]
		}
		yellow := dq3data.Color{R: 255, G: 224, B: 32}
		g.newGame.draw(g.rgba, g.dlg.tx, white, yellow) // 主選單/命名/性別(ngSplash 無疊繪)
		g.settings.draw(g.rgba, g.cfg)                  // 設定選單(可疊其上;未開時 no-op)
		g.frame.WritePixels(g.rgba)
		return
	}
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
			tile := sc.blk.Tile(sc.tileIdx(camX+cx, camY+cy))
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
	// 船(地表):停泊船 tile;主角在船上 → 畫船 tile 取代主角
	if !g.inTown && g.shipOwned && !g.shipAboard &&
		g.shipX >= camX && g.shipX < camX+ViewCols && g.shipY >= camY && g.shipY < camY+ViewRows {
		blitTile(g.rgba, (g.shipX-camX)*TileW, (g.shipY-camY)*TileH, sc.blk.Tile(shipTileFor(0)), sc.pal)
	}
	if !g.inTown && g.shipAboard { // 在船上 → 船 tile 取代主角 sprite
		blitTile(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH, sc.blk.Tile(shipTileFor(g.facing)), sc.pal)
	} else {
		blitSprite(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH,
			g.hero.Frames[g.facing*dq3data.CharWalk+g.walk], sc.pal)
	}
	white := dq3data.Color{R: 255, G: 255, B: 255}
	if len(sc.pal) > 15 {
		white = sc.pal[15]
	}
	if g.cmd.open { // 命令窗(左上)+ 左下 HP/MP/等級小窗(oracle:verify_open_21_after_space.png)
		yellow := dq3data.Color{R: 255, G: 224, B: 32}
		g.cmd.draw(g.rgba, white, yellow, 48, 32)
		g.drawCmdStatus(g.rgba, white)
	}
	if g.dlg.open { // 對話框
		g.dlg.draw(g.rgba, white)
	}
	if g.shop.active { // 商店
		g.shop.draw(g.rgba, g.heroGold, white)
	}
	if g.noticeTimer > 0 && g.noticeCode >= 0 { // 取得道具通知(品名)
		fillBox(g.rgba, 24, 244, ScreenW-48, 40, white)
		g.shop.drawItemName(g.rgba, 40, 256, g.noticeCode, white)
	}
	switch g.panel { // 資訊面板
	case panelStatus:
		g.drawStatus(g.rgba, white)
	case panelItem:
		g.drawItems(g.rgba, white)
	case panelEquip:
		g.drawEquip(g.rgba, white)
	}
	if g.tavern.active { // 酒館招募
		g.tavern.draw(g.rgba, white)
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
	g := &Game{rgba: make([]byte, ScreenW*ScreenH*4), input: newInput(), cfg: config.Default()}
	g.initStoryBits() // [0x4f70] NPC 可見性旗標初值(必須在預載 town0 前;零值=全清=全隱藏)

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

	// 城鎮:懶載素材(NPC sprite 來源 + fs);阿里阿罕(CTY0)預載並快取。通用載入見 loadTownScene。
	manBLS := ld.read("DQ3MAN.BLS")
	if ld.err != nil {
		return nil, ld.err
	}
	g.assets, g.worldPal, g.manBLS = assets, pal, manBLS
	g.towns = map[int]*Scene{}
	g.curCty = -1
	town0, terr := loadTownScene(assets, pal, manBLS, 0, 1, 0, g.storyFlag) // 阿里阿罕(開局白天)
	if terr != nil {
		return nil, fmt.Errorf("loadTown 阿里阿罕: %w", terr)
	}
	g.town, g.towns[0] = town0, town0

	// 對話 + 命令窗:字型 D3TXT00.FON(常駐)+ 阿里阿罕 bank 1(D3TXT01.TXT,section dlg_bank=1)
	fon := ld.read("D3TXT00.FON")
	g.dlg.tx = dq3data.LoadText(fon, ld.read("D3TXT01.TXT"))
	g.cmd.tx = g.dlg.tx                                             // 命令窗標籤 glyph 也走同一字型
	g.settings.tx = g.dlg.tx                                        // 設定選單標籤 glyph 也走同一字型
	g.shop.nameText = dq3data.LoadText(fon, ld.read("D3TXT00.TXT")) // 品名 = D3TXT00 rec=code+1
	g.dlg.itemNames = g.shop.nameText                               // 對話插值 VAR_ITEM 用同一品名表(W1)
	g.dlg.heroName = g.heroName                                     // 對話插值 VAR_NAME 用主角名(創角/讀檔後另行同步)
	g.endText = dq3data.LoadText(fon, ld.read("ENDTXT.TXT"))        // 結局文本(破索瑪後捲動)
	g.endSeq = -1                                                   // 結局未進行
	g.openingIdx = -1                                               // 開場旁白未進行
	g.battle.nameText = g.shop.nameText                             // 咒文名同名表
	g.tavern.tx = g.dlg.tx                                          // 酒館 glyph
	g.hero = dq3data.LoadCharSprite(ld.read("DQ3MST.BLS"), 0)

	// 戰鬥:怪物數值(D3MNS.DAT)+ sprite(DQ3MNS.SHP)+ 怪物色盤(MNSBK.PAL)
	mons, merr := dq3data.OpenMonsters(ld.read("D3MNS.DAT"))
	if merr != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenMonsters: %w", merr)
	}
	g.battle.mons = mons
	g.battle.shp = ld.read("DQ3MNS.SHP")
	g.battle.scr = ld.read("PACKBG.SCR") // 戰鬥背景
	g.battle.mpal = dq3data.DecodePalette(ld.read("MNSBK.PAL"), 256)
	g.battle.tx = g.dlg.tx

	// 設施:道具/裝備資料(ITEM.DAT)供商店價格
	items, ierr := dq3data.OpenItems(ld.read("ITEM.DAT"))
	if ierr != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenItems: %w", ierr)
	}
	g.shop.items = items
	g.shop.tx = g.dlg.tx
	if ld.err != nil {
		return nil, ld.err
	}

	g.cur = g.over
	g.px, g.py = g.over.w/2, g.over.h/2  // 地表起點(暫用中心)
	g.heroGold = 120                     // 初始金(新遊戲勇者)
	g.equip = [4]int{3, 0x21}            // 初始裝備:銅劍 + 皮甲冑
	g.companions = startingCompanions(0) // 示範隊伍(戰士/僧侶/魔法使);酒館招募之後接
	g.flags = map[int]bool{}
	g.initStoryBits()      // 新遊戲重置 [0x4f70] NPC 可見性旗標
	g.progressSet(msStart) // 開場里程碑(阿里阿罕:見國王、酒場建隊)
	g.noticeCode = -1
	g.prng.Seed(0x1357)                                 // 祈禱之戒損壞判定 RNG(對齊 C apply_item_use)
	g.music = gaudio.NewMusic(music)                    // MT-32 音樂(music fs 為 nil → 靜音降級)
	if sfxRaw := ld.read("FVOC.VCX"); len(sfxRaw) > 0 { // 數位音效(VOC)
		bank := dq3data.DecodeVOCBank(sfxRaw, 44100)
		pcm := make([][]int16, len(bank))
		for i, s := range bank {
			pcm[i] = s.PCM
		}
		g.music.SetSFX(pcm)
	}
	if os.Getenv("DQ3_FM") != "" { // SB-FM 音樂(OPL2 合成 MBG.MCX,取代 MT-32 OGG)
		if mbg := ld.read("MBG.MCX"); len(mbg) > 0 {
			g.music.SetMBG(mbg)
		}
	}
	if tit := ld.read("TITG.P"); len(tit) > 0 { // 標題畫面(PCX)
		if pix, pal, w, h, e := dq3data.DecodePCX(tit); e == nil && w == ScreenW && h == ScreenH {
			g.titlePix, g.titlePal, g.showTitle = pix, pal, true
		}
	}
	g.frame = ebiten.NewImage(ScreenW, ScreenH)
	// 注意:不再於此自動續玩——存檔改由標題主選單「載入進度」明確觸發(newgame.go newGameInput
	// ngOptLoad 分支),對齊 docs/36 開場流程(遊戲開始 = 全新主角,不會被舊存檔悄悄蓋掉)。
	applyDebugEnv(g) // debug 環境變數可覆蓋(此時 music/frame 已就緒)
	// 起始配樂(依最終場景;debug 進城已自行換軌)
	if !g.inTown {
		g.music.Play(trackField)
	}
	g.renderFrame() // 首幀
	log.Printf("地表 %d×%d / 阿里阿罕 %d×%d(spawn %d,%d、NPC %d)— 走到城鎮入口進城、Space 命令窗、Esc 出城",
		g.over.w, g.over.h, g.town.w, g.town.h, g.town.spawnX, g.town.spawnY, len(g.town.npcs))
	return g, nil
}

// applyDebugEnv:headless 截圖用的除錯環境變數(真機/正常執行無設 → 全 no-op)。
func applyDebugEnv(g *Game) {
	// 任何場景/戰鬥 debug hook → 略過標題(截圖直接到目標畫面)
	for _, k := range []string{"DQ3_START", "DQ3_BATTLE", "DQ3_SHOP", "DQ3_TAVERN", "DQ3_ENTER", "DQ3_DLG", "DQ3_CMD"} {
		if os.Getenv(k) != "" {
			g.showTitle = false
		}
	}
	if os.Getenv("DQ3_START") == "town" {
		g.cur, g.inTown, g.curCty = g.town, true, 0 // 阿里阿罕
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
	if os.Getenv("DQ3_SHIP") != "" { // debug:給船,停在主角旁(驗證船系統)
		g.shipOwned = true
		g.shipX, g.shipY = g.px+1, g.py
		g.progressSet(msShip) // 取船里程碑
	}
	if r := os.Getenv("DQ3_EXP"); r != "" { // debug:設起始經驗(截圖驗證等級/咒文)
		var e int
		if _, err := fmt.Sscanf(r, "%d", &e); err == nil {
			g.heroExp = uint32(e)
		}
	}
	if r := os.Getenv("DQ3_ENTER"); r != "" { // debug:起手進指定 CTY 城(截圖驗證通用載入)
		var cty int
		if _, e := fmt.Sscanf(r, "%d", &cty); e == nil {
			g.enterTownCty(cty)
		}
	}
	if os.Getenv("DQ3_SHOP") != "" { // debug:起手開武防店(截圖驗證)
		g.openFacility(1)
	}
	if v := os.Getenv("DQ3_TAVERN"); v != "" { // debug:起手開酒館(截圖驗證)
		g.tavern.open()
		if v == "name" { // 直接跳命名格盤
			g.tavern.stage, g.tavern.ni.cursor, g.tavern.ni.nameBuf = tavName, 0, []int{15, 16, 17}
		}
		if v == "zhu" { // 直接跳注音盤
			g.tavern.stage, g.tavern.ni.nameZhu = tavName, true
			g.tavern.ni.zh.Init()
		}
	}
	if r := os.Getenv("DQ3_BATTLE"); r != "" { // debug:起手開一場戰鬥(截圖驗證)
		var id int
		if _, e := fmt.Sscanf(r, "%d", &id); e == nil {
			level, maxHP, atk, def, agi := g.heroStats()
			mp := g.heroMaxMP()
			g.heroHP, g.heroMP = maxHP, mp
			g.battle.showInfo = g.cfg.CombatInfo
			g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
			g.battle.start(id, 1, heroParams{level: level, curHP: maxHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
				mp: mp, maxMP: mp, spells: g.heroSpells()}, g.buildCompanionActors())
			if os.Getenv("DQ3_SPELLMENU") != "" && len(g.battle.spells) > 0 {
				g.battle.cursor, g.battle.phase = bcSpell, phSpell // 開咒文子選單
			}
		}
	}
}
