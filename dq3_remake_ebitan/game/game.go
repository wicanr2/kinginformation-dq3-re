// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(評估見 ../docs/62)。
// Phase 3+:載入真實地表(DQ3CON.MAP+DQ3.BLK)+ 阿里阿罕城(CTY00.DAT+DQ31.BLK),全用移植的 Go parser。
// 方向鍵走動、viewport 捲動、主角 masked sprite + 走路動畫、BLKBM 碰撞(不可穿山/海)。
package game

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gaudio"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

const (
	audioCueEnding  = "ending"
	audioCueTitle   = "title"
	audioCueField   = "field"
	audioCueCastle  = "castle"
	audioCueTown    = "town"
	audioCueDungeon = "dungeon"
	audioCueBattle  = "battle"
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
	facing   int
	walk     int
	spr      *dq3data.CharSprite
}

// partyTrailEntry is one player position recorded immediately before a legal
// grid move.  Followers consume the same history in party order, matching the
// original caterpillar train instead of teleporting or inventing a path.
type partyTrailEntry struct {
	x, y   int
	facing int
}

// Scene 是一張可走動的地圖(地表或城鎮),統一 render + 碰撞邏輯。
type Scene struct {
	blk             *dq3data.BLK
	attr            *dq3data.BlockAttr
	pal             []dq3data.Color
	w, h            int
	tileAt          func(x, y int) int // (x,y) → BLK tile 索引
	spawnX, spawnY  int                // 城鎮進入 spawn(地表不用)
	mapFlags        byte               // CTY section header +0x10；原版野外咒文場景 gate
	dlgText         *dq3data.Text      // 該城對話 bank(D3TXT0<bank>.TXT;地表 nil)
	hiMap           []byte             // 每格高 byte(事件 subid)
	specialHandlers []int              // CTY section+4 scripted handler raw IDs
	events          [][3]int           // section 事件表 {type,param,p2}
	transitions     [][4]int           // section 轉場表 {destCty,destSec,x,y}
	plainTransition bool               // 此 section 允許 attr==0 + 非零 hiMap subid 的隱形轉場
	sec             int                // section 號
	npcs            []npcInst
	override        map[int]int // 執行期 tile 覆蓋(開門後 door→通行 tile;key=y*w+x)
	npcRng          rng.RNG     // NPC 遊走用確定性 RNG(依 section 種子)
	night           bool        // 此 scene 用黑夜 NPC 表載入(晝夜 cache 判斷用;overworld 恆 false)
}

// currentSceneSection 是 production code 使用的 section 讀取器；測試檔可另以
// sceneSection 便利斷言，但正式 build 不得依賴 `_test.go` 的 helper。
func currentSceneSection(s *Scene) int {
	if s == nil {
		return -1
	}
	return s.sec
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

// openDoor:開 (tx,ty) 門及相鄰 8 格的連續門片。原版 file 0x14977 先改中心，
// 再由 sub_14a49 掃 3×3 其餘格；每格用 hiMap&0x1f 作通行 tile，清事件位留轉場位。
func (sc *Scene) openDoor(tx, ty int) bool {
	if sc.doorTier(tx, ty) == 0 {
		return false
	}
	for y := ty - 1; y <= ty+1; y++ {
		for x := tx - 1; x <= tx+1; x++ {
			sc.openDoorTile(x, y)
		}
	}
	return true
}

func (sc *Scene) openDoorTile(tx, ty int) {
	if sc.doorTier(tx, ty) == 0 {
		return
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
}

// revealEventTile:原版 sub_18C01 的「發現隱藏樓梯」地圖變更。
// 目前格的低 byte tile +1，並清除 hiMap 低 5 bit 的事件 subid；高 3 bit
// 仍保留。變更只存在於當前 scene，永久狀態由事件的 story flag 負責。
func (sc *Scene) revealEventTile(tx, ty int) bool {
	if tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return false
	}
	i := ty*sc.w + tx
	if sc.override == nil {
		sc.override = map[int]int{}
	}
	sc.override[i] = (sc.tileIdx(tx, ty) + 1) & 0xff
	if sc.hiMap != nil {
		sc.hiMap[i] &= 0xe0
	}
	return true
}

// applyClearedEventTiles 還原 DQ3.EXE file 0x46df..0x4747 的載圖後事件掃描：
// 每個 attr&8 事件格以 hiMap subid 找 section event 的 p2 story flag；flag 已清時，
// low tile +1 並清 hiMap 低 5 bit。這不是魔法球特例，寶箱／隱藏地形等也共用此規則。
func (sc *Scene) applyClearedEventTiles(flagSet func(int) bool) {
	if flagSet == nil || sc.attr == nil || sc.hiMap == nil || len(sc.events) == 0 {
		return
	}
	for y := 0; y < sc.h; y++ {
		for x := 0; x < sc.w; x++ {
			i := y*sc.w + x
			if sc.attr.Raw(sc.tileIdx(x, y))&0x0008 == 0 {
				continue
			}
			subid := int(sc.hiMap[i] & 0x1f)
			if subid >= len(sc.events) || flagSet(sc.events[subid][2]) {
				continue
			}
			if sc.override == nil {
				sc.override = map[int]int{}
			}
			sc.override[i] = (sc.tileIdx(x, y) + 1) & 0xff
			sc.hiMap[i] &= 0xe0
		}
	}
}

// tileTransition:回 (tx,ty) 的轉場(attr&0xe000 轉場格 → hiMap 低5bit subid → transitions[subid])。
// 回 (destCty, destSec, dx, dy, ok)。移植 dq3_scene_tile_transition。
func (sc *Scene) tileTransition(tx, ty int) (int, int, int, int, bool) {
	if sc.hiMap == nil || len(sc.transitions) == 0 || tx < 0 || ty < 0 || tx >= sc.w || ty >= sc.h {
		return 0, 0, 0, 0, false
	}
	subid := int(sc.hiMap[ty*sc.w+tx] & 0x1f)
	// 多數門/階梯由 BLKBM attr&0xe000 標記；少數已由原版流程確認的 section
	// 允許 attr==0 的普通可走 tile 以非零 hiMap subid 選轉場。
	a := sc.attr.Raw(sc.tileIdx(tx, ty))
	if a&0xe000 == 0 && !(sc.plainTransition && a == 0 && subid > 0) {
		return 0, 0, 0, 0, false
	}
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
	over, town                *Scene // 地表 / 目前城鎮
	under                     *Scene // 下層地表(DQ3UND.MAP,懶載;共用地表 BLK/attr/pal)
	layer                     int    // 目前地表層:0=地面 1=下層(城鎮進出以此決定回哪層)
	cur                       *Scene
	inTown                    bool
	curCty                    int   // 目前所在 CTY 號(-1=地表)
	dnPhase                   int   // 晝夜相位:0白天 1黃昏 2黑夜 3黎明(僅地表走動推進;城內固定)
	dnStep                    int   // 地表步數計數器(每 dnPhaseSteps 步推進一相位)
	assets                    fs.FS // 素材(懶載其他城鎮)
	encounters                *dq3data.EncounterTables
	worldPal                  []dq3data.Color
	manBLS                    []byte         // NPC sprite 來源
	towns                     map[int]*Scene // 已載入城鎮快取(cty→Scene)
	overPx, overPy            int            // 記住進城前的地表座標(Esc 回來用)
	worldEntranceGrace        bool           // 城鎮→地表後第一個成功步伐不重新觸發同列兩格入口（strong；見 docs/103）
	hero                      *dq3data.CharSprite
	heroRole                  *dq3data.CharSprite // 由 game-pack temporary_role active flag 推導；不另存檔
	mstBLS                    []byte              // party sprite 來源(DQ3MST.BLS；由 pack asset 指定)
	partySpriteCache          map[[2]int]*dq3data.CharSprite
	partyTrail                [8]partyTrailEntry
	partyTrailLen             int
	px, py                    int // 主角在 cur 內的 tile 座標
	facing                    int // 0..3
	walk                      int // 0/1 走路動畫相位
	cd, anim                  int
	dlg                       Dialogue       // 對話視窗
	cmd                       CmdMenu        // 野外命令窗
	battle                    Battle         // 戰鬥場景
	shop                      Shop           // 商店(武防/道具店)
	church                    Church         // 教會服務
	tavern                    Tavern         // 露易達酒館 2F 冒險者登錄所(創角→僅登錄名冊 roster,不自動入隊)
	recruit                   Recruit        // 露易達酒館 1F 酒場(找同伴參加/與同伴分離/觀看名單;roster↔companions)
	panel                     panelKind      // 資訊面板(狀況/道具/裝備)
	panelCursor               int            // 裝備面板游標
	panelActor                int            // 裝備對象：-1=先選隊員、0=主角、1..=companions
	panelHits                 hitList        // 道具/裝備清單可點區塊(drawItems/drawEquip 重建;panelStatus 無列表不使用)
	itemActionStage           int            // 0=道具清單、1=使用/給予/丟掉、2=給予對象
	itemActionCursor          int            // 動作或給予對象游標
	itemSelected              int            // 已選 g.inventory index
	reclassEventID            string         // game-pack reclass event；空字串=無進行中的轉職
	reclassStage              int            // 達瑪轉職對話／選擇狀態機
	reclassCursor             int            // Yes/No、隊員或職業選單游標
	reclassMember             int            // 0=主角、1..=companions；只有非勇者可交易
	reclassTarget             int            // pack raw class；未選=-1
	settlementFounderStage    int            // game-pack 建城者交付事件階段
	settlementFounderCursor   int            // 兩次 Yes/No 游標
	settlementFounderEventID  string         // active settlement_founder event
	settlementFounderMember   int            // companions index；未選=-1
	settlementFounderOverflow int            // 預存所滿時原版逐道具訊息的剩餘次數
	settlementFounder         *Member        // 已離隊、不可由酒館重新招募的建城者
	settlementFounderFollowup []string       // 後續 NPC 對話尚待顯示的 game-pack text ID
	sharedStorage             []int          // 原版共用預存所；包含建城者交出的裝備與道具
	fieldSpell                FieldSpellMenu // 野外咒文／魯拉目的地 modal
	visitedTowns              []townVisit    // 魯拉可選的已造訪城鎮（存檔持久化）
	inventory                 []int          // 持有道具 id
	music                     *gaudio.Music
	input                     *Input // 抽象輸入(鍵盤 + 觸控)
	showTitle                 bool   // 標題畫面(含主選單/主角創建流程進行中;false=已進入一般遊戲)
	titlePix                  []uint8
	titlePal                  []dq3data.Color
	newGameConfirmPix         []uint8
	newGameConfirmPal         []dq3data.Color
	attractPix                [][]uint8
	attractPal                [][]dq3data.Color
	attractSeq                *gamepack.AttractSequence
	attractFrame              int
	attractIndex              int
	attractActive             bool
	titleIdleFrames           int
	endingPix                 []uint8 // pack asset ending_image(TIT3.P)
	endingPal                 []dq3data.Color
	cfg                       config.Config  // 可攜設定(RNG/音樂/音量/音源/戰鬥資訊/受傷特效);NewGame 用 config.Default() 初始化
	pack                      *gamepack.Pack // versioned 精訊版 game pack；遊戲設定不得再散落為 Go table
	settings                  Settings       // 設定選單 modal(標題畫面按 S 開)
	newGame                   NewGameFlow    // 標題主選單 + 主角命名/性別創建 modal（正式流程已閉合；幾何仍待 V3）
	heroName                  []int          // 主角姓名(glyph index,注音/英數命名輸入結果;空=尚未創建/debug 略過)
	heroGender                int            // 主角性別(0=男 1=女)
	// 主角進度(勇者 class0)。heroStat 是原版角色 record 的七個持久能力欄；
	// 創角/升級時由 sub_ed3c 相同的 RNG transaction 修改，不再每幀由等級公式重算。
	heroExp                 uint32
	heroGold                int
	heroHP                  int
	heroMP                  int
	heroStat                stats.Values
	heroInit                bool
	equip                   [4]int                          // 裝備槽:0 武器 1 鎧 2 盾 3 兜(item code;0=空)
	companions              []*Member                       // 現役隊伍同伴(隊長=hero*,最多 3;經 recruit.go「找同伴參加」從 roster 拉入)
	roster                  []*Member                       // 冒險者名冊(酒場 2F 登錄所創角→僅入此;未必在隊伍中,見 docs/36 rec527-550)
	flags                   map[int]bool                    // remake 暫存旗標；原版劇情旗標一律使用 storyBits
	storyBits               [64]byte                        // 原版 [0x4f70] flag 陣列；祭壇用到 0x131，非舊誤判的 256-bit 上限
	worldState              uint16                          // 原版 [0x4f44] 世界狀態；bit0x20=蓋亞火山、bit0x40=彩虹橋
	trackedWorldPositions   map[string]trackedWorldPosition // game-pack 動態世界物件座標；與玩家返回座標分離
	worldObjectBufferX      int                             // 原版 80x80 world buffer 左上角（scratch，不進存檔）
	worldObjectBufferY      int
	worldObjectBufferValid  bool
	coordinateItemGateID    string // 自動地表道具 gate 的進行中事件（scratch）
	coordinateItemGateStage int
	coordinateForcedSteps   int
	coordinateForcedDir     int
	noticeCode              int // 取得道具通知(item code;-1=無)
	noticeTimer             int
	shipOwned               bool          // 已取得船(波魯多加胡椒換船,milestone SHIP)
	shipAboard              bool          // 目前在船上
	shipX, shipY            int           // 船停泊位置(地表)
	repel                   int           // 聖水驅敵剩餘步數(>0 期間地表不遇弱敵,每步遞減)
	remoaru                 int           // レムオル透明剩餘有效移動步數；原版 [0x52f7]，施放設 25
	toramana                bool          // 多拉瑪那原版 world flag 0x2000；遇高傷害地板後轉 hazardGuard
	hazardGuard             bool          // 原版 world flag 0x0400；只在同一段連續高傷害地板維持
	prng                    rng.RNG       // 祈禱之戒損壞判定用 RNG
	lotoBlessed             bool          // 破索瑪後洛特冊封(勇者裝備昇華為傳說的洛特裝備)
	cleared                 bool          // 已破關(索瑪擊破)
	endText                 *dq3data.Text // ENDTXT.TXT 結局文本
	endSeq                  int           // 結局捲動段號(-1=未進行;0..n=當前段)
	openingIdx              int           // 開場旁白序號(-1=未進行;0..n=當前句)
	bossQueue               []int         // boss 連戰佇列(索瑪終盤:怨靈→殭屍→索瑪；或 examine 觸發的座標 boss 鏈)
	pendingTrigger          *bossTrigger  // 目前由 examine 觸發、正在跑的座標 boss 鏈(非 nil 期間 onBattleEnd 需結算);完成或中斷即清 nil
	bossIntro               bool          // 座標 boss 的 preRec 尚在顯示；關閉後才正式開戰
	bossSurrenderStage      int           // game-pack boss_surrender primitive 的目前階段
	bossSurrenderCursor     int           // 求饒 Yes/No 游標
	bossSurrenderEventID    string        // active pack event；空字串表示無事件
	stagedBossStage         int           // game-pack staged_boss primitive 階段
	stagedBossCursor        int           // 第二階段 Yes/No 游標
	stagedBossEventID       string        // active staged boss pack event
	stagedBossMoveIndex     int           // 第一戰後自動移動方向索引
	stagedBossMoveTick      int           // 第一戰後自動移動 frame 節流
	stagedBossMovingNPC     bool          // true 時先重播原版 NPC slot0 路徑，再交易旗標並移動玩家
	zomaIntro               bool          // 索瑪 rec72 尚在顯示；關閉後才開 monster0x7c
	ortegaStage             int           // CTY90 sec4 歐魯迪卡事件：1=rec69，2=rec70
	endingKingStage         int           // CTY80 sec1 handler74：1=rec48 冊封對白，關閉後才開始 ending
	baramosReturn           int           // 1=阿里阿罕國王 rec98，2=索瑪 rec99；關閉後開放 CTY72
	temporaryRoleStage      int           // game-pack temporary_role primitive 的目前階段
	temporaryRoleCursor     int           // temporary_role Yes/No 游標
	temporaryRoleEventID    string        // active pack event；空字串表示無事件
	soloChallengeActive     bool          // 原版 DS:0x4f46 bit0x80 的具名暫時單人模式
	soloChallengeCompanions []*Member     // 試煉期間離隊、返回時原樣復隊；必須進存檔
	soloChallengeStage      int           // game-pack temporary_solo_challenge 對話階段
	soloChallengeCursor     int           // 試煉 Yes/No 游標
	soloChallengeEventID    string        // active pack event；試煉期間保持以支援存讀檔
	sequenceGateStage       int           // game-pack two_step_floor_switch_gate 的目前階段
	sequenceGateCursor      int           // 序列開關 Yes/No 游標
	sequenceGateEventID     string        // active pack event；空字串表示無事件
	sequenceGateSubID       int           // 本次踩到的 pack-owned switch subid
	sequenceGateArmed       bool          // 原版暫存 scratch；不進存檔

	choiceItemExchangeStage   int    // game-pack choice_item_exchange 對話／交換階段
	choiceItemExchangeCursor  int    // 兩組 Yes/No 共用游標
	choiceItemExchangeEventID string // active choice_item_exchange event
	trackedWorldLocatorStage  int    // tracker item 的三段原版記錄播放階段
	trackedWorldLocatorID     string // active tracked_world_object id

	vehicleExchangeStage   int    // game-pack staged_vehicle_exchange 的目前階段
	vehicleExchangeEventID string // active pack event；空字串表示無事件
	guidedPassageStage     int    // game-pack guided_passage primitive 階段
	guidedPassageEventID   string // active pack event；空字串表示無事件
	guidedPassageNPC       int    // 引路 NPC runtime index
	guidedPassageWaypoint  int    // 目前目標 waypoint
	guidedPassageTick      int    // waypoint 間的 frame 節流
	hostageRescueStage     int    // game-pack hostage_rescue primitive 階段
	hostageRescueCursor    int    // Yes/No 游標
	hostageRescueEventID   string // active pack event；空字串表示無事件
	hostageRescueMovement  int    // captive_movements 目前索引
	hostageRescueWaypoint  int    // movement 目前 waypoint
	hostageRescueTick      int    // movement frame 節流
	hostageRescueNPC       int    // movement runtime NPC index
	mirrorStage            int    // 沙曼歐莎拉之鏡事件:1=rec97 2=rec98 3=怪力魔戰鬥
	phoenix                *dq3data.CharSprite
	phoenixOwned           bool
	phoenixAboard          bool
	phoenixX               int
	phoenixY               int
	phoenixStage           int   // 1=守護神 rec92 後續；2=逐顆 rec93；3=中央蛋六幀復活動畫
	phoenixList            []int // 已放上祭壇、待 rec93 列出的寶珠
	phoenixListPos         int
	phoenixAnim            int // 復活動畫 tile 0x7b..0x80 的目前 frame
	phoenixAnimTick        int
	frame                  *ebiten.Image
	rgba                   []byte
}

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
		if idx := g.facingNPC(); idx >= 0 {
			n := &g.cur.npcs[idx]
			// Pack-owned NPC transactions may use the original runner/subtype
			// value as a selector even when the map record is sub0/sub1. Check
			// this finite data primitive before the generic dialogue branch.
			if g.talkNPCItemReward(n) {
				return
			}
			if g.talkConditionalFacility(n) {
				return
			}
			switch sub := (n.ctrl >> 3) & 7; {
			case sub <= 1: // 對話
				g.dlg.Open(n.b4)
			case sub == 2: // scripted NPC(給物/劇情)
				if g.curCty == ctyPhoenixShrine && n.b4 == phoenixGuardianHandler {
					g.talkPhoenixGuardian()
				} else if g.talkZomaFinal(n) {
					// CTY90 sec5 handler80：自然交談入口。
				} else if g.talkEndingKing(n) {
					// CTY80 sec1 handler74：索瑪後回拉達多姆，由國王冊封才進 ending。
				} else if g.talkStagedBoss(n) {
					// game-pack 第一戰→移動→選擇→第二戰→寶箱 primitive。
				} else if g.talkBossSpecial(n) {
					// CTY65 巴拉摩斯等 sub2 固定物件型 boss。
				} else if g.openAliahanSpecialNPC(n) {
					// CTY00 b4=1/3/4 是 EXE sub2 jump table 的設施 handler，
					// 不屬於泛用 scriptedTable。
				} else if g.talkTemporaryRole(n) {
					// game-pack required-item → temporary-role → restore primitive。
				} else if g.talkTemporarySoloChallenge(n) {
					// game-pack party removal → solo entrance → exact-party restore primitive。
				} else if g.talkQuestItemChain(n) {
					// game-pack treasure → item exchange → location-use primitive。
				} else if g.talkChoiceItemExchange(n) {
					// game-pack 雙分支詢問 → 道具交換 → 世界狀態交易 primitive。
				} else if g.talkStagedVehicleExchange(n) {
					// game-pack quest item → required item → vehicle primitive。
				} else if g.talkGuidedPassage(n) {
					// game-pack required item → guide walk → scene trigger primitive。
				} else if g.talkHostageRescue(n) {
					// game-pack guard → captive switch → boss → reward primitive。
				} else if g.talkReclass(n) {
					// game-pack NPC → gate → class-change transaction primitive。
				} else if g.talkSettlementFounder(n) {
					// game-pack active companion → founder + shared-storage transaction。
				} else if g.talkSettlementFounderFollowup(n) {
					// game-pack founder identity → flag-gated followup dialogue sequence。
				} else {
					g.scriptedTalk(n.b4)
				}
			case sub >= 3: // 設施(店/宿/教會)
				g.openFacility(n.b4)
			}
		}
	case cmdStatus: // 狀況
		g.panel = panelStatus
	case cmdSpell:
		g.openFieldSpellMenu()
	case cmdItem: // 道具
		g.panel, g.panelCursor = panelItem, 0
		g.itemActionStage, g.itemActionCursor, g.itemSelected = itemActionList, 0, -1
	case cmdEquip: // 裝備
		g.panel, g.panelCursor, g.panelActor = panelEquip, 0, -1
	case cmdExamine: // 調査:檢查面向格 → 寶箱/隱藏物(一次性旗標)
		g.examine()
	}
}

// facingNPC 找面向的交談對象。通常是前一格；若前一格是不可走櫃台，
// 再查櫃台後一格。CTY00 sec2 登錄員位於(2,3)，玩家只能站(2,5)
// 隔著 row4 櫃台交談，故不能把「永遠只查一格」當成原版語意。
func (g *Game) facingNPC() int {
	if g.cur == nil {
		return -1
	}
	dx, dy := dirDelta(g.facing)
	x1, y1 := g.px+dx, g.py+dy
	if idx := g.cur.npcAt(x1, y1); idx >= 0 {
		return idx
	}
	if x1 >= 0 && y1 >= 0 && x1 < g.cur.w && y1 < g.cur.h &&
		g.cur.attr.Blocked(g.cur.tileIdx(x1, y1)) {
		return g.cur.npcAt(x1+dx, y1+dy)
	}
	return -1
}

// openAliahanSpecialNPC 對齊 DQ3.EXE sub2 jump table：
// b4=1 file0x6482→file0x16dd（露依達酒場 rec527/528）；
// b4=3 file0x649f→file0x1a4c（冒險者登錄所 rec550）；
// b4=4 file0x64a3→file0x7c50（預存所 rec561，尚待獨立實作）。
func (g *Game) openAliahanSpecialNPC(n *npcInst) bool {
	if g.curCty != 0 || g.cur == nil {
		return false
	}
	switch {
	case g.cur.sec == 0 && n.b4 == 1:
		g.recruit.open()
		return true
	case g.cur.sec == 2 && n.b4 == 3:
		g.tavern.open()
		return true
	case g.cur.sec == 0 && n.b4 == 4:
		g.dlg.OpenFrom(g.shop.nameText, 561) // 全域 D3TXT00 rec561「預存所」。
		return true
	}
	return false
}

func (g *Game) hasItem(code int) bool { return g.countItem(code) > 0 }

// hasPartyItem 對齊原版需要掃描整個隊伍物品記錄的劇情消費者；
// hasItem 仍保留「主角背包」語意，避免一般道具選單與隊伍交易混用。
func (g *Game) hasPartyItem(code int) bool {
	if g.hasItem(code) {
		return true
	}
	for _, member := range g.companions {
		if containsInt(member.Inventory, code) {
			return true
		}
	}
	return false
}

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

// normalizeSurfaceTarget 套用原版上世界地圖的接縫座標。
// DQ3.EXE sub_12758 並非普通的模數運算：跨越 X 時寫入 242/2，跨越 Y 時寫入
// 204/2，以保留兩格捲動邊界。下世界使用不同的合併座標範圍；其載入端與消費端
// 尚未各自閉合前，超界一律採失敗即關閉。
func (g *Game) normalizeSurfaceTarget(nx, ny int) (int, int, bool) {
	if g.cur == nil {
		return nx, ny, false
	}
	if g.inTown || g.layer != 0 {
		return nx, ny, nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h
	}
	switch {
	case nx < 0:
		nx = g.cur.w - 2
	case nx >= g.cur.w:
		nx = 2
	}
	switch {
	case ny < 0:
		ny = g.cur.h - 1
	case ny >= g.cur.h:
		ny = 2
	}
	return nx, ny, nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h
}

// tryMove:判定能否移動到 (nx,ny)(含船系統)。移植 dq3_ship_input + 碰撞。
func (g *Game) tryMove(nx, ny int) bool {
	if nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h {
		return false
	}
	if g.phoenixAboard { // 原版 mode2：飛行中無視地形，只受地圖邊界限制
		return true
	}
	if g.phoenixOwned && !g.inTown && nx == g.phoenixX && ny == g.phoenixY {
		g.phoenixAboard = true // 走上 field tile 0xfd 搭乘，不使用 remake 自造快捷鍵
		g.shipAboard = false
		return true
	}
	if g.shipAboard { // 原版 mode1：attr bit1 clear=可航；bit1 set 再以 bit0 分上岸/阻擋
		if g.cur.attr.Raw(g.cur.tileIdx(nx, ny))&0x0002 == 0 {
			return true // 航行至海面或河道格
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
	if !g.cur.Blocked(nx, ny) {
		return true
	}
	return g.tryPushableNPC(nx, ny)
}

// scriptedTalk:對 scripted NPC(sub2)—— 檢查型/給予型(前置道具→給物+旗標+對白)。移植 dq3_scripted。
func (g *Game) scriptedTalk(byte4 int) {
	s := scriptedFor(byte4, g.curCty)
	if s == nil {
		return
	}
	give, prereq, require, consume, milestone := s[2], s[3], s[4], s[5], s[6]
	beforeRec, giveRec, afterRec := s[7], s[8], s[9]
	if byte4 == dragonQueenHandler && g.curCty == ctyDragonQueen {
		g.talkDragonQueen(s)
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

const (
	ctyDragonQueen     = 67
	dragonQueenHandler = 52
	itemLightOrb       = 0x65
)

// talkDragonQueen 還原 DQ3.EXE sub_15E02：取得光之珠後，
// CLEAR story flag0x4e 並 SET story flag0x19。舊泛用表只給道具，漏了兩筆設定。
func (g *Game) talkDragonQueen(s *[10]int) {
	if g.storyFlag(0x19) || g.hasItem(itemLightOrb) {
		g.dlg.Open(s[9])
		return
	}
	g.inventory = append(g.inventory, itemLightOrb)
	g.noticeCode, g.noticeTimer = itemLightOrb, 120
	g.setStoryFlag(0x4e, false)
	g.setStoryFlag(0x19, true)
	g.dlg.Open(s[8])
	g.setDlgVarItem(itemLightOrb)
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
		if t.preRec >= 0 && g.dlg.Open(t.preRec) {
			g.bossIntro = true
		} else {
			g.advanceBossQueue()
		}
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
		if ev[0] == 4 { // sub_18C01：清事件 flag、tile+1，顯示全域 rec484「發現一座隱藏樓梯」
			if !g.storyFlag(ev[2]) {
				return true
			}
			g.setStoryFlag(ev[2], false)
			g.cur.revealEventTile(x, y)
			g.dlg.OpenFrom(g.shop.nameText, 484)
			return true
		}
		presentBefore := g.storyFlag(ev[2])
		if g.collectQuestTreasure(g.curCty, g.cur.sec, subid) {
			// 原版載圖掃描與當下取得寶箱都以「present flag 清除」將 low tile
			// 加一並移除 event subid；不能等到重新進場才顯示已開啟外觀。
			if presentBefore && !g.storyFlag(ev[2]) {
				g.cur.revealEventTile(x, y)
			}
			return true
		}
		t := treasureFor(g.curCty, g.cur.sec, subid)
		if t == nil {
			return false
		}
		// CTY10 sec5 的金皇冠位於甘達特強制戰後方。原版 handler14 勝利並接受
		// 求饒後才 clear story flag0x2e；事件尚在時不得直接從寶箱繞過 boss。
		if g.treasureBlockedByPack(g.curCty, g.cur.sec, x, y, t[4]) {
			return true
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
	if g.tryPhoenixAltar(fx, fy) || g.tryPhoenixAltar(g.px, g.py) {
		return
	}
	if checkBoss(fx, fy) || checkBoss(g.px, g.py) { // 座標 boss 觸發點優先於一般寶箱/warp
		return
	}
	if !check(fx, fy) {
		check(g.px, g.py) // 也試腳下
	}
}

// talkBossSpecial:sub2 固定物件型 boss。巴拉摩斯是 CTY65 (8,3),handler70，
// 由「話す」觸發；既有地圖固定格 boss 仍可走 examine()。
func (g *Game) talkBossSpecial(n *npcInst) bool {
	if g.cur == nil || n == nil {
		return false
	}
	t := bossTriggerAt(g.curCty, g.cur.sec, n.x, n.y)
	if t == nil || g.flags[t.doneFlag] {
		return false
	}
	g.pendingTrigger = t
	g.bossQueue = append([]int(nil), t.monsters...)
	if t.preRec >= 0 && g.dlg.Open(t.preRec) {
		g.bossIntro = true
	} else {
		g.advanceBossQueue()
	}
	return true
}

func (g *Game) advanceBossDialogue() {
	if !g.bossIntro || g.pendingTrigger == nil {
		return
	}
	g.bossIntro = false
	g.advanceBossQueue()
}

// keyTier returns the highest door tier selected by a held pack-owned key.
// Path planning may use it, but opening still requires formally using the
// selected item through the item-action menu.
func (g *Game) keyTier() int {
	if g.pack == nil {
		return 0
	}
	best := 0
	for _, effect := range g.pack.ItemUseEffects() {
		if effect.EffectID == "open_facing_locked_door" &&
			g.hasItem(effect.ItemRawID) && effect.DoorKeyTier > best {
			best = effect.DoorKeyTier
		}
	}
	return best
}

// openFacility:面向設施 NPC(byte4=設施索引 k)→ 依當前 CTY 開對應設施(全城設施表)。
func (g *Game) openFacility(k int) {
	f := facilityForCty(g.curCty, k)
	if f == nil {
		return
	}
	switch f.typ {
	case facInn: // 旅社：單價×存活人數；付清後治滿存活成員 HP/MP，死亡者不復活。
		_, maxHP, _, _, _ := g.heroStats()
		alive := 0
		if g.heroHP > 0 {
			alive++
		}
		for _, m := range g.companions {
			if m.Alive() {
				alive++
			}
		}
		cost := f.innCost * alive
		if alive > 0 && g.heroGold >= cost {
			g.heroGold -= cost
			if g.heroHP > 0 {
				g.heroHP, g.heroMP, g.heroInit = maxHP, g.heroMaxMP(), true
			}
			for _, m := range g.companions {
				if m.Alive() {
					m.fullHeal()
				}
			}
			// DQ3.EXE file 0x876a..0x8778：住宿成功後 [0x526c]=1（白天）、
			// [0x251d]=0（時刻歸零）。付款失敗不得改變日夜。
			g.setDaynight(0)
		}
	case facWeapon, facItem: // 商店:開貨架(品項自全城品項池 itemOff..+count)
		lo, hi := f.itemOff, f.itemOff+f.count
		if lo >= 0 && hi <= len(shopItemPool) {
			g.shop.open(append([]int(nil), shopItemPool[lo:hi]...))
		}
	case facChurch:
		g.church.open(g.shop.nameText)
	case facRecord: // 記錄點:存檔(冒險之書)
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
	g.updateTouchContext()        // 必須在 Poll() 之前,讓這幀 poll 用得到新標籤
	return g.step(g.input.Poll()) // 抽象輸入(鍵盤 + 觸控合流)
}

// step 是 production Update 在裝置輸入 Poll() 之後的完整遊戲邏輯。
// 拆出 InputState 注入點，讓玩家流程回歸可重播同一條正式狀態機；
// 測試不得再用直接呼叫事件函式冒充完整 playthrough。
func (g *Game) step(in InputState) error {
	moved := false

	// 標題畫面:主選單→主角命名→性別→能力確認→開始新遊戲(newgame.go)。
	// S/CtxTap 開設定選單(疊在標題上,ESC/Cancel 關閉回標題)。
	if g.showTitle {
		if g.attractActive {
			if titleInputInterruptsAttract(in) {
				g.stopAttract()
				g.newGameInput(in)
			} else {
				g.advanceAttract()
			}
			g.renderFrame()
			return nil
		}
		if g.settings.open {
			g.settingsInput(in)
			g.renderFrame()
			return nil
		}
		if g.newGame.stage == ngSplash {
			if titleInputInterruptsAttract(in) {
				g.titleIdleFrames = 0
			} else if g.attractSeq != nil {
				g.titleIdleFrames++
				if g.titleIdleFrames >= g.attractSeq.StartDelayFrames {
					g.startAttract()
				}
			}
		} else {
			g.titleIdleFrames = 0
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
				g.playAudioCue(audioCueField)
			}
		}
		g.renderFrame()
		return nil
	}
	// 酒館 2F 登錄所 modal:職業→姓名→性別→建角。**只登錄到 roster,不自動入隊**
	// (docs/36 rec527-550:登錄所與 1F 酒場「找同伴參加」是兩段式;C dq3_tavern.c 建完角色
	// 自動 dq3_party_add 是該 demo 的簡化,這裡刻意不照抄,避免滿隊時頂替丟隊友)。
	if g.tavern.active {
		g.tavernCreate(in)
		g.renderFrame()
		return nil
	}
	// 酒場 1F 招募 modal:找同伴參加/與同伴分離/觀看名單(roster↔companions,見 recruit.go)
	if g.recruit.active {
		g.recruitInput(in)
		g.renderFrame()
		return nil
	}
	if g.fieldSpell.active {
		g.fieldSpellInput(in)
		g.renderFrame()
		return nil
	}
	if g.church.active {
		g.churchInput(in)
		g.renderFrame()
		return nil
	}
	if g.bossSurrenderStage == bossSurrenderChoice {
		g.bossSurrenderChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.stagedBossStage == stagedBossSecondChoice {
		g.stagedBossChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.temporaryRoleChoosing() {
		g.temporaryRoleChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.soloChallengeStage == soloChallengeChoice {
		g.soloChallengeChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.sequenceGateChoosing() {
		g.sequenceGateChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.choiceItemExchangeChoosing() {
		g.choiceItemExchangeChoiceInput(in)
		g.renderFrame()
		return nil
	}
	switch g.hostageRescueStage {
	case hostageRescueGuardChoice, hostageRescueSwitchChoice,
		hostageRescueBossChoice, hostageRescueRewardChoice:
		g.hostageRescueChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.reclassChoosing() {
		g.reclassChoiceInput(in)
		g.renderFrame()
		return nil
	}
	if g.settlementFounderChoosing() {
		g.settlementFounderChoiceInput(in)
		g.renderFrame()
		return nil
	}
	// 資訊面板 modal:狀況/道具 = B/A 關；裝備先選隊員，再選背包裝備。
	// 點列(P2,道具/裝備清單)= 游標移過去 + 等同 A 使用/裝上
	if g.panel != panelNone {
		tapIdx := -1
		if in.Tapped {
			tapIdx = g.panelHits.at(in.TapX, in.TapY)
		}
		switch {
		case in.Cancel && g.panel == panelEquip && g.panelActor >= 0:
			g.panelActor, g.panelCursor = -1, 0
		case in.Cancel && g.panel == panelItem && g.itemActionStage == itemActionTarget:
			g.itemActionStage, g.itemActionCursor = itemActionMenu, 0
		case in.Cancel && g.panel == panelItem && g.itemActionStage == itemActionMenu:
			g.itemActionStage, g.itemActionCursor, g.itemSelected = itemActionList, 0, -1
		case in.Cancel:
			g.panel = panelNone
		case g.panel == panelEquip:
			confirm := in.Confirm
			n := g.equipCandidateCount(g.panelActor)
			if g.panelActor < 0 {
				n = 1 + len(g.companions)
			}
			if tapIdx >= 0 && n > 0 {
				g.panelCursor, confirm = tapIdx, true
			}
			switch {
			case confirm:
				if g.panelActor < 0 {
					g.panelActor, g.panelCursor = g.panelCursor, 0
				} else {
					g.equipSelected()
				}
			case in.DirEdge == 0 && n > 0: // 下
				g.panelCursor = (g.panelCursor + 1) % n
			case in.DirEdge == 1 && n > 0: // 上
				g.panelCursor = (g.panelCursor + n - 1) % n
			}
		case g.panel == panelItem:
			confirm := in.Confirm
			if g.itemActionStage == itemActionList && tapIdx >= 0 && len(g.inventory) > 0 {
				g.panelCursor, confirm = tapIdx, true
			}
			switch g.itemActionStage {
			case itemActionList:
				switch {
				case confirm && len(g.inventory) > 0:
					g.itemSelected = g.panelCursor
					g.itemActionStage, g.itemActionCursor = itemActionMenu, 0
				case in.DirEdge == 0 && len(g.inventory) > 0:
					g.panelCursor = (g.panelCursor + 1) % len(g.inventory)
				case in.DirEdge == 1 && len(g.inventory) > 0:
					g.panelCursor = (g.panelCursor + len(g.inventory) - 1) % len(g.inventory)
				}
			case itemActionMenu:
				switch {
				case confirm && g.itemActionCursor == 0:
					g.panelCursor = g.itemSelected
					g.itemActionStage = itemActionList
					g.useSelectedItem()
					g.itemSelected = -1
				case confirm && g.itemActionCursor == 1 && len(g.companions) > 0:
					g.itemActionStage, g.itemActionCursor = itemActionTarget, 0
				case confirm && g.itemActionCursor == 2:
					g.dropSelectedItem()
				case in.DirEdge == 0:
					g.itemActionCursor = (g.itemActionCursor + 1) % 3
				case in.DirEdge == 1:
					g.itemActionCursor = (g.itemActionCursor + 2) % 3
				}
			case itemActionTarget:
				switch {
				case confirm && len(g.companions) > 0:
					if g.giveSelectedItem(g.itemActionCursor) {
						g.itemActionStage, g.itemActionCursor, g.itemSelected =
							itemActionList, 0, -1
					}
				case in.DirEdge == 0 && len(g.companions) > 0:
					g.itemActionCursor = (g.itemActionCursor + 1) % len(g.companions)
				case in.DirEdge == 1 && len(g.companions) > 0:
					g.itemActionCursor = (g.itemActionCursor + len(g.companions) - 1) % len(g.companions)
				}
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
			if !g.dlg.open {
				g.advanceBossDialogue()
				g.advanceZomaIntro()
				g.advanceOrtegaEvent()
				g.advanceEndingKing()
				g.advanceBaramosReturn()
				g.advanceTemporaryRoleDialogue()
				g.advanceSoloChallengeDialogue()
				g.advanceMirrorEvent()
				g.advancePhoenixEvent()
				g.advanceBossSurrenderDialogue()
				g.advanceSequenceGateDialogue()
				g.advanceChoiceItemExchangeDialogue()
				g.advanceTrackedWorldObjectLocator()
				g.advanceCoordinateItemGateDialogue()
				g.advanceStagedVehicleExchangeDialogue()
				g.advanceGuidedPassageDialogue()
				g.advanceHostageRescueDialogue()
				g.advanceReclassDialogue()
				g.advanceSettlementFounderDialogue()
				g.advanceSettlementFounderFollowup()
				g.advanceStagedBossDialogue()
			}
		}
		g.renderFrame()
		return nil
	}
	if g.stagedBossStage == stagedBossFirstMoving {
		g.advanceStagedBossMovement()
		g.renderFrame()
		return nil
	}
	if g.coordinateForcedSteps > 0 {
		g.advanceCoordinateItemGateMovement()
		g.renderFrame()
		return nil
	}
	if g.phoenixStage == 3 {
		g.advancePhoenixAnimation()
		g.renderFrame()
		return nil
	}
	if g.guidedPassageAnimating() {
		g.advanceGuidedPassageAnimation()
		g.renderFrame()
		return nil
	}
	if g.hostageRescueAnimating() {
		g.advanceHostageRescueAnimation()
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
	// 開場演出:rec82→83→81 在家中；母親帶到家門外後播 rec80，才交回操作。
	// DQ3.EXE file 0x140a..0x147a 與 runner handler54(file 0x147b)。
	if g.openingIdx >= 0 {
		g.openingIdx++
		if g.openingIdx < len(openingSeq) {
			if openingSeq[g.openingIdx] == openingMotherDirectionsRec {
				g.motherEscort()
			}
			g.dlg.Open(openingSeq[g.openingIdx])
		} else {
			g.openingIdx = -1
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
		if g.phoenixAboard {
			g.tryLandPhoenix()
			g.renderFrame()
			return nil
		}
		g.cmd.Open()
		g.renderFrame()
		return nil
	}
	// 移動(方向 held + grid cooldown)
	if g.cd > 0 {
		g.cd--
	} else if in.DirHeld >= 0 {
		g.facing = in.DirHeld // 撞牆也轉向(對齊原版:面向先變,可走才移動)
		dx, dy := dirDelta(in.DirHeld)
		nx, ny := g.px+dx, g.py+dy
		// CTY section0 的可走邊界開口就是正常出城。部分原始 spawn 在邊界
		// (阿里阿罕 0,28)，部分在開口內一格(雷貝 1,10)；限制「必須正好等於
		// spawn」會令後者走到邊界後永遠無法離城。
		if g.inTown && g.cur.sec == 0 &&
			(nx < 0 || ny < 0 || nx >= g.cur.w || ny >= g.cur.h) {
			g.exitTown()
			g.renderFrame()
			return nil
		}
		if !g.inTown {
			var valid bool
			nx, ny, valid = g.normalizeSurfaceTarget(nx, ny)
			if !valid {
				g.cd = moveCooldown
				g.renderFrame()
				return nil
			}
		}
		if g.tryMove(nx, ny) { // 碰撞/航行:陸走 BLKBM+NPC、船走海、上/下船
			g.recordPartyTrail()
			g.px, g.py = nx, ny
			moved = true
		}
		g.cd = moveCooldown
	}
	if moved {
		g.advanceFieldSpellStep()
		g.applyHazardStep()
	}
	if moved && g.inTown { // 城內:踩到轉場格(門/階梯/出城)→ 切 section / 跨 CTY / 出城
		g.tryTransition()
		g.tryTrackingGuardEvent()
		g.tryPushPuzzleCompletionEvent()
		g.trySequenceGateEvent()
		g.tryOpeningRegionEvent()
		g.tryBossSurrenderEvent()
		g.tryGuidedPassageTrigger()
		g.tryHostageRescueTrigger()
		g.tryBaramosReturnEvent()
		g.tryOrtegaEvent()
	} else if moved && !g.inTown && !g.phoenixAboard { // 地表:飛行時不進城、不推晝夜、不遇敵
		g.updateTrackedWorldObjectWindow(g.facing)
		if g.tryCoordinateItemGateEvent() {
			g.renderFrame()
			return nil
		}
		skipEntrance := g.worldEntranceGrace
		g.worldEntranceGrace = false
		if !skipEntrance {
			cty := findCtyAtLayer(g.px, g.py, g.layer) // 依目前地表層找入城
			cty = g.resolveSoloChallengeWorldEntrance(cty)
			if g.layer == 0 && g.px >= 54 && g.px <= 55 && g.py == 129 {
				if g.storyFlag(0x4d) {
					cty = 71 // 巴拉摩斯前：封閉中的蓋亞洞窟
				} else {
					cty = 72 // 索瑪現身後：地震破壞蓋亞洞窟封印
				}
			}
			if pc := worldEntranceResolve(g.px, g.py, g.layer,
				g.pack.WorldEntranceVariants(), g.storyFlag); pc >= 0 {
				cty = pc
			}
			if cty >= 0 {
				g.enterTownCty(cty)
				return nil
			}
			if g.tryEnterTrackedWorldObject() {
				return nil
			}
		}
		g.advanceDaynight() // 地表走一步 → 推進晝夜(每 60 步一相位;城內不推進)
		if g.repel > 0 {    // 聖水驅敵期間每步遞減、不遇弱敵(移植 #3 repel)
			g.repel--
		} else if g.prng.Next(16) == 0 {
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
	case 17: // 達瑪神殿（原版地表 lookup 實際選到 CTY17）
		g.progressSet(msDhama)
	}
	g.px, g.py = sc.startPos() // spawn 出界/不可走 → pick_open_start(修 6 城進城掉圖外)
	g.resetPartyTrail()
	if sc.dlgText != nil { // 切到該城對話 bank(NPC 對話用正確 bank)
		g.dlg.tx = sc.dlgText
	}
	g.cd = moveCooldown
	g.playSceneMusic(cty)
	g.rememberTown()
	g.renderFrame()
}

// dnPhaseSteps:每晝夜相位的地表步數。使用者指定「白天→黑夜=120 步、中間有黃昏」→ 每相位 60 步
// (白天→黃昏→黑夜=2×60;全 4 相位循環=240 步,對齊原版 clock 0..0xf0)。移植 main.c DN_PHASE_STEPS。
const dnPhaseSteps = 60

// isNight:目前是否黑夜(相位 2)。供夜 gate 事件查詢(提頓夜綠寶珠等)。對齊 dq3_scene.c night 判定。
func (g *Game) isNight() bool { return g.dnPhase == 2 }

// storyFlag / setStoryFlag:原版 [0x4f70] story-flag 陣列(NPC 可見性等;SET=file 0x824f、
// CLR=0x8264、GET=0x8279)。祭壇會存取 flag 0x12c..0x131，故不能截成舊誤判的 256 flags。
// 供 loadTownSceneSec 的 NPC 可見性過濾;獨立於 g.flags(remake 里程碑),不互相污染。
func (g *Game) storyFlag(id int) bool {
	if id < 0 || id >= len(g.storyBits)*8 {
		return false
	}
	return g.storyBits[id>>3]&(0x80>>(uint(id)&7)) != 0
}

func (g *Game) setStoryFlag(id int, v bool) {
	if id < 0 || id >= len(g.storyBits)*8 {
		return
	}
	m := byte(0x80 >> (uint(id) & 7))
	if (g.storyBits[id>>3]&m != 0) == v {
		return
	}
	if v {
		g.storyBits[id>>3] |= m
	} else {
		g.storyBits[id>>3] &^= m
	}
	// NPC 可見性在 loadTownSceneSec 時依 story flag 過濾；已載入的其他 CTY
	// 若不失效，離開再回來仍會沿用事件前的 NPC 清單。當前 Scene 指標繼續供
	// 本幀事件／動畫使用，下次進城或 section 載入才按新旗標重建。
	clear(g.towns)
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
	g.resetPartyTrail()
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
	// CTY90 的橋東側 (19,29) 同時是通往 sec5 的 transition。原版
	// handler79 要求玩家先站在橋東側，再由右往左踏入 (18,29)；
	// e0 尚在且向右跨上橋時必須暫緩 transition，否則玩家永遠無法
	// 以正式輸入站到事件的起始格。事件完成後 e0 會清除，向右離橋
	// 才恢復原始 sec4→sec5 轉場。
	if g.curCty == ctyZomaCastle && g.cur != nil && g.cur.sec == ortegaSection &&
		g.px == ortegaTriggerX+1 && g.py == ortegaTriggerY &&
		g.facing == 3 && g.storyFlag(ortegaFightFlag) {
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
			g.progressSet(msDescend) // CTY77 的原版 0xfe 出口：自然進入愛列夫加特
		}
		g.exitTown()
		g.worldEntranceGrace = true
		if dx != 0 && dx < g.cur.w {
			g.px = dx
		}
		if dy != 0 && dy < g.cur.h {
			g.py = dy
		}
		g.overPx, g.overPy = g.px, g.py
		// The original vehicle state stores only ownership/mounted status; the
		// descent entrance has no separate persistent lower-world sprite
		// coordinate. Re-anchor an owned, parked Phoenix at the real landing
		// tile so the next formal Confirm can board it without a state injection.
		if dsec == 0xfe && g.phoenixOwned && !g.phoenixAboard {
			g.phoenixX, g.phoenixY = g.px, g.py
		}
		g.renderFrame()
		return
	}
	cross := dcty != g.curCty && dcty < 100 // 跨 CTY vs 同城換 section
	ncty := g.curCty
	if cross {
		ncty = dcty
		// DQ3.EXE file 0x49d1..0x49f2：跨 CTY 時先以新 CTY 索引查
		// DGROUP 0x748 cty_loc，寫回 remembered world X/Y，再載目的 CTY。
		// 否則 CTY30→31 後由 section0 出洞會錯誤回到 CTY30 的地表入口。
		if loc := ctyLoc[ncty]; loc[2] == g.layer {
			g.overPx, g.overPy = loc[0], loc[1]
		}
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
	g.resetPartyTrail()
	if ns.dlgText != nil { // 切到目的城對話 bank
		g.dlg.tx = ns.dlgText
	}
	if cross { // 跨 CTY → 換配樂
		g.playSceneMusic(ncty)
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
	g.resetPartyTrail()
	if ns.dlgText != nil {
		g.dlg.tx = ns.dlgText
	}
	g.playSceneMusic(dcty)
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
	g.applyRainbowBridge()
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

// descend:舊 C remake debug helper，直接降到下層地表。
// 原版正式流程已更正為 CTY72→CTY77→0xfe（tryTransition），production 不呼叫此函式。
func (g *Game) descend() {
	u := g.loadUnder()
	if u == nil {
		return
	}
	g.layer, g.cur, g.inTown, g.curCty = 1, u, false, -1
	g.px, g.py = 84, 68
	g.resetPartyTrail()
	g.progressSet(msDescend) // 下降里程碑(對 flagDescended)
	g.playAudioCue(audioCueField)
	g.renderFrame()
}

// openingSeq 由 DQ3.EXE file 0x140a..0x147a 與 handler54(file 0x147b)定錨：
// 生日旁白 → 母親邀請 → 主角回答 → 到家門外指向王城。
const openingMotherDirectionsRec = 80

var openingSeq = []int{82, 83, 81, openingMotherDirectionsRec}

// 原版新遊戲初始化常數(DQ3.EXE file 0x13a0..0x1431):
// remembered overworld=(0x99,0xae)、CTY00 sec4、室內位置=(5,5)。
// section header 的 spawn(10,10)是門轉場目的地，不是新遊戲出生點。
const (
	openingHomeX  = 5
	openingHomeY  = 5
	aliahanWorldX = 0x99
	aliahanWorldY = 0xae
)

// startOpening:新遊戲創角完成 → 進主角家(CTY00 sec4 室內)+ 播開場旁白；正式 opening
// production trace 已閉合，逐格演出與畫面 parity 仍屬 V3 工作。
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
	g.px, g.py = openingHomeX, openingHomeY
	if g.px >= home.w || g.py >= home.h { // 資產版本不符時 fail-safe；合法 CTY00 sec4 必為 13×13
		g.px, g.py = home.startPos()
	}
	g.resetPartyTrail()
	g.rememberTown() // 原版初始進度已包含阿里阿罕；確保首次習得魯拉時可選。
	// dlg.tx 維持 NewGame 初始的 D3TXT01(開場旁白 rec82/83 在此 bank)
	g.openingIdx = 0
	g.dlg.Open(openingSeq[0])
	g.playSceneMusic(0)
	g.renderFrame()
}

// motherEscort:開場 rec81 後 → 母親帶到阿里阿罕家門外(sec0 8,38)。
// 原版 runner handler54(file 0x147b)檢查初始 flag0x50，演出後 set flag0x17、
// clear flag0x50，再播 rec80；本函式保存相同可見性狀態與落點。
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
	g.resetPartyTrail()
	if sec0.dlgText != nil {
		g.dlg.tx = sec0.dlgText // 切到城鎮對話 bank
	}
	g.setStoryFlag(0x17, true)
	g.setStoryFlag(0x50, false)
	g.renderFrame()
}

const (
	ctyAliahanCastle     = 25
	aliahanThroneSection = 1
	aliahanKingX         = 9
	aliahanKingY         = 6
	recAliahanKingStart  = 78
)

var aliahanKingRewardItems = [...]int{0x00, 0x01, 0x01, 0x03, 0x1f, 0x1f}

// talkAliahanKing 還原首次謁見獎勵。DQ3.EXE handler56(file 0x15bc..0x1633):
// rec78 → GIVE 00,01,01,03,1f,1f → 加 0x32(50)G → clear flag17/set flag18。
// 這批是國王明說「給同伴的武器及防具」，不等同主角出生裝備。
func (g *Game) talkAliahanKing() {
	if !g.storyFlag(0x17) || g.progressDone(msStart) {
		return
	}
	g.inventory = append(g.inventory, aliahanKingRewardItems[:]...)
	g.heroGold += 0x32
	g.setStoryFlag(0x17, false)
	g.setStoryFlag(0x18, true)
	g.progressSet(msStart)
	g.dlg.Open(recAliahanKingStart)
}

// tryOpeningRegionEvent:阿里阿罕王座不是一般可見 NPC 對話，而是 runner region
// handler56。玩家走到國王正前方 (9,7) 即觸發首次謁見。
func (g *Game) tryOpeningRegionEvent() bool {
	if !g.inTown || g.cur == nil || g.curCty != ctyAliahanCastle ||
		g.cur.sec != aliahanThroneSection || g.px != aliahanKingX || g.py != aliahanKingY+1 ||
		!g.storyFlag(0x17) || g.progressDone(msStart) {
		return false
	}
	g.talkAliahanKing()
	return true
}

// tryBaramosReturnEvent:巴拉摩斯勝利後回阿里阿罕王座正前方，
// 原版 sub2 handler5 依 flag0x29 播 rec98→rec99，最後 CLR flag0x4d。
func (g *Game) tryBaramosReturnEvent() bool {
	if !g.inTown || g.cur == nil || g.curCty != ctyAliahanCastle ||
		g.cur.sec != aliahanThroneSection || g.px != aliahanKingX || g.py != aliahanKingY+1 ||
		g.storyFlag(0x29) || !g.storyFlag(0x4d) || g.baramosReturn != 0 {
		return false
	}
	if !g.dlg.Open(98) {
		return false
	}
	g.baramosReturn = 1
	return true
}

func (g *Game) advanceBaramosReturn() {
	switch g.baramosReturn {
	case 1:
		if g.dlg.Open(99) {
			g.baramosReturn = 2
		}
	case 2:
		g.baramosReturn = 0
		g.setStoryFlag(0x4d, false)
		// CTY71/72 同一地表座標、依此旗標選版本；清 cache 避免沿用舊場景。
		delete(g.towns, 71)
		delete(g.towns, 72)
	}
}

func (g *Game) exitTown() {
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1 // 回目前地表層(地面/下層)
	g.px, g.py = g.overPx, g.overPy
	g.resetPartyTrail()
	// 從船上世界物件回到地表時，出口可能與停泊船是同一座標。此時恢復
	// vehicle mode1；否則玩家會被放在不可步行的水格且無法重新踏上船。
	if g.shipOwned && g.layer == 0 && g.px == g.shipX && g.py == g.shipY {
		g.shipAboard = true
	}
	g.cd = moveCooldown
	g.playAudioCue(audioCueField)
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

// heroStats 讀主角持久能力欄並重算裝備衍生值。DQ3.EXE sub_9521 證實：
// 攻 = 力量 + 武器攻；守 = 耐力 + 防具防（不是舊 remake 猜的耐力/2）。
func (g *Game) heroStats() (level, maxHP, atk, def, agi int) {
	level = stats.LevelForExp(0, g.heroExp)
	g.ensureHeroStats()
	maxHP = int(g.heroStat[stats.HP])
	atk = int(g.heroStat[stats.STR])
	def = int(g.heroStat[stats.VIT])
	agi = int(g.heroStat[stats.AGI])
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
	g.ensureHeroStats()
	return int(g.heroStat[stats.MP])
}
func (g *Game) heroSpells() []int { return spell.HeroKnown(stats.LevelForExp(0, g.heroExp)) }

// ensureHeroStats 只供舊存檔/debug/精簡單元測試向後相容。正常新遊戲一定在
// ngConfirm 前由 rollHeroLevelOne 建立隨機 Lv1 record。
func (g *Game) ensureHeroStats() {
	if g.heroStat != (stats.Values{}) {
		return
	}
	level := stats.LevelForExp(0, g.heroExp)
	// 舊存檔沒有持久能力欄；以每級 target 建出安全狀態，之後的升級才走原版 RNG。
	for k := stats.StatKind(0); k < stats.StatCount; k++ {
		g.heroStat[k] = stats.GrowthTarget(0, k, level)
	}
}

func (g *Game) rollHeroLevelOne() {
	g.heroExp = 0
	g.heroStat = stats.InitValues(0, &g.prng)
	g.heroHP = int(g.heroStat[stats.HP])
	g.heroMP = int(g.heroStat[stats.MP])
	g.heroInit = true
}

// startEncounter:依目前座標查原版 region/subtable/candidates，再依出現權重算群量。
func (g *Game) startEncounter() {
	level, maxHP, atk, def, agi := g.heroStats()
	maxMP := g.heroMaxMP()
	if !g.heroInit {
		g.heroHP, g.heroMP, g.heroInit = maxHP, maxMP, true
	}
	hp := heroParams{level: level, curHP: g.heroHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
		herbs: g.countPartyItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells()}
	comps := g.buildCompanionActors()
	region := g.encounters.Region(g.px, g.py)
	slot := g.encounters.Slot(region, g.prng.Next(4))
	if len(slot.Candidates) == 0 {
		return
	}
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	start := g.prng.Next(len(slot.Candidates))
	for i := 0; i < len(slot.Candidates); i++ {
		monID := slot.Candidates[(start+i)%len(slot.Candidates)]
		count := g.encounterGroupCount(monID, slot.Threshold)
		if g.battle.startGroup(monID, count, int64(g.anim)*2654+1, hp, comps) {
			g.playAudioCue(audioCueBattle)
			return
		}
	}
}

// 遭遇群量演算法(docs/13 encounter_build_group,sub_a86d):戰鬥點數預算 38、
// 候選怪 D3MNS +0x28 出現權重,maxN=floor(38/weight)、實際隻數 1..maxN 擲亂數、群上限 8。
// W2(docs/72 A3)範圍:單種同種群戰;混群(不同種同場)登記待辦,未做。
const (
	encounterBudget   = 38 // [0x2513] 戰鬥點數預算
	encounterGroupCap = 8  // [0x231f] 群隻數上限
)

// spawnGroupMax 依出現權重算群量上限:floor(budget/weight),夾在 [1, encounterGroupCap];
// weight<=0(非群戰候選/無資料)→ 1。抽成純函式方便邊界值單元測試(權重極小 clip 到群上限、
// 權重過大 floor 為 0 需保底 1)。
func spawnGroupMax(weight int) int {
	if weight <= 0 {
		return 1
	}
	maxN := encounterBudget / weight
	if maxN < 1 {
		maxN = 1
	}
	if maxN > encounterGroupCap {
		maxN = encounterGroupCap
	}
	return maxN
}

// encounterGroupCount 依 monID 的出現權重(D3MNS +0x28)算本場實際隻數(1..maxN 均勻擲值)。
// 無怪物資料 → 單隻。用 g.prng(已於 NewGame 設種子的通用確定性 RNG,同祈禱之戒判定共用一條狀態)
// 擲值,不占用戰鬥自身 RNG(那個要等 startGroup 決定隻數後才依 seed 開新序列)。
func (g *Game) encounterGroupCount(monID int, threshold ...int) int {
	if g.battle.mons == nil {
		return 1
	}
	st, ok := g.battle.mons.Stat(monID)
	if !ok {
		return 1
	}
	// 原版子表 byte1 是「強制單隻」門檻；raw 百分比低於門檻時直接生成 1 隻。
	if len(threshold) > 0 && g.prng.Next(100) < threshold[0] {
		return 1
	}
	return 1 + g.prng.Next(spawnGroupMax(int(st.SpawnWeight)))
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
		herbs: g.countPartyItem(herbCode), mp: g.heroMP, maxMP: maxMP, spells: g.heroSpells()}
	g.battle.lightOrb = monID == 0x7c && g.hasPartyItem(itemLightOrb) // 隊伍持光之珠 → 索瑪二階段弱化
	g.battle.showInfo = g.cfg.CombatInfo
	g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
	if g.battle.start(monID, int64(g.anim)*2654+1, hp, g.buildCompanionActors()) {
		g.playAudioCue(audioCueBattle)
		return true
	}
	return false
}

const (
	ctyZomaCastle    = 90
	zomaFinalSection = 5
	zomaHandler      = 80
	zomaDialogueRec  = 72
	ortegaSection    = 4
	ortegaTriggerX   = 18
	ortegaTriggerY   = 29
	ortegaFightFlag  = 0xe0
	ortegaDyingFlag  = 0x0e
	ortegaFightRec   = 69
	ortegaDyingRec   = 70
)

// tryOrtegaEvent:CTY90 sec4 橋上由右往左踏入 (18,29) 時自動開始歐魯迪卡事件。
// handler79(sub_16494) 要求 flag e0，先顯示 D3TXT07 rec69；原版接著播放
// formation 0x80 的固定演出，並非玩家可下指令的通常戰鬥。
func (g *Game) tryOrtegaEvent() bool {
	if !g.inTown || g.curCty != ctyZomaCastle || g.cur == nil || g.cur.sec != ortegaSection ||
		g.px != ortegaTriggerX || g.py != ortegaTriggerY || g.ortegaStage != 0 ||
		!g.storyFlag(ortegaFightFlag) || g.facing != 2 {
		return false
	}
	if !g.dlg.Open(ortegaFightRec) {
		return false
	}
	g.ortegaStage = 1
	return true
}

// advanceOrtegaEvent:對齊 handler79→handler90 的可見 NPC 旗標交易。
// rec69 關閉後 CLEAR e0、SET 0e，重載地圖令橋上的戰鬥 sprite 換成垂死歐魯迪卡；
// rec70 關閉後 CLEAR 0e 並移除最後 sprite。兩段都不開玩家戰鬥 UI。
func (g *Game) advanceOrtegaEvent() {
	switch g.ortegaStage {
	case 1:
		g.setStoryFlag(ortegaFightFlag, false)
		g.setStoryFlag(ortegaDyingFlag, true)
		g.reloadTownDaynight()
		if g.dlg.Open(ortegaDyingRec) {
			g.ortegaStage = 2
		} else {
			g.setStoryFlag(ortegaDyingFlag, false)
			g.reloadTownDaynight()
			g.ortegaStage = 0
		}
	case 2:
		g.setStoryFlag(ortegaDyingFlag, false)
		g.reloadTownDaynight()
		g.ortegaStage = 0
	}
}

// talkZomaFinal 是 CTY90 section5 的 handler80 正式入口。六個 monster106 是城內
// 一般隨機遭遇，不屬於此序列；原版 formation 是 0x7a、0x7b、0x7c。
func (g *Game) talkZomaFinal(n *npcInst) bool {
	if g.curCty != ctyZomaCastle || g.cur == nil || g.cur.sec != zomaFinalSection ||
		n == nil || n.b4 != zomaHandler || g.cleared {
		return false
	}
	g.startZomaSeq()
	return true
}

// startZomaSeq:索瑪終盤 formation 序列：巴拉摩斯怨靈122 → 殭屍123 → 索瑪124。
func (g *Game) startZomaSeq() {
	g.bossQueue = []int{0x7a, 0x7b, 0x7c}
	g.advanceBossQueue()
}

// advanceBossQueue:取下一個 formation 開戰；索瑪本體先播放 D3TXT07 rec72。
func (g *Game) advanceBossQueue() {
	for len(g.bossQueue) > 0 {
		id := g.bossQueue[0]
		g.bossQueue = g.bossQueue[1:]
		if id == 0x7c && g.dlg.Open(zomaDialogueRec) {
			g.zomaIntro = true
			return
		}
		if g.startBossBattle(id) {
			return
		}
		// start 失敗(無 sprite)→ 跳過此 boss,續取下一個
	}
}

func (g *Game) advanceZomaIntro() {
	if !g.zomaIntro {
		return
	}
	g.zomaIntro = false
	g.startBossBattle(0x7c)
}

// runFinale:打倒索瑪後的原版 aftermath。handler80/sub_164CD 先 CLEAR flag e1，
// 把玩家送回愛列夫加特地表；此時尚未冊封、也尚未播放 ending。玩家必須自行走回
// 拉達多姆 CTY79→CTY80 sec1，和 handler74 國王交談。
func (g *Game) runFinale() {
	g.progressSet(msZoma)
	g.cleared = true
	g.setStoryFlag(0xe1, false)
	g.lotoBlessed = false
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	delete(g.flags, 0x217)
	g.endSeq = -1
	g.dlg.open = false
	g.dlg.tx = nil
	g.layer, g.inTown, g.curCty = 1, false, -1
	if under := g.loadUnder(); under != nil {
		g.under, g.cur = under, under
	}
	// CTY90 索瑪城位於下層 (110,73)；原版勝利後回到城外入口。下一次玩家
	// 移動才會跑地表進城判定，因此保留原始入口格，不需人工偏移到鄰接山地。
	g.px, g.py = ctyLoc[ctyZomaCastle][0], ctyLoc[ctyZomaCastle][1]
	g.overPx, g.overPy = g.px, g.py
	if g.music != nil {
		g.playAudioCue(audioCueField)
	}
	g.renderFrame()
}

const (
	ctyRadatomeCastle  = 80
	radatomeThroneSec  = 1
	radatomeKingHandle = 74
	radatomeEndingRec  = 48
)

// talkEndingKing:IDA handler74/sub_16346 的 post-Zoma 分支。flag e1 已被索瑪
// 勝利清除時，國王不走平時 rec22/23，而是進 sub_1E713 的 rec48 冊封序列。
func (g *Game) talkEndingKing(n *npcInst) bool {
	if !g.cleared || g.lotoBlessed || g.storyFlag(0xe1) || g.curCty != ctyRadatomeCastle ||
		g.cur == nil || g.cur.sec != radatomeThroneSec || n == nil || n.b4 != radatomeKingHandle {
		return false
	}
	if !g.dlg.Open(radatomeEndingRec) {
		return false
	}
	g.endingKingStage = 1
	return true
}

func (g *Game) advanceEndingKing() {
	if g.endingKingStage != 1 {
		return
	}
	g.endingKingStage = 0
	g.lotoBlessed = true // そして伝説へ…：國王在此正式授予「勇者洛特」封號
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	g.flags[0x217] = true
	g.playAudioCue(audioCueEnding)
	// sub_1E713 後半是固定 ending 演出；remake 以已解出的 ENDTXT 分段播放。
	if g.endText != nil && g.endText.NRecords > 0 {
		g.endSeq = 0
		g.dlg.tx = g.endText
		g.dlg.Open(0)
	}
}

func (g *Game) playAudioCue(id string) {
	if g.music == nil || g.pack == nil {
		return
	}
	if track, ok := g.pack.AudioTrack(id); ok {
		g.music.Play(track)
	}
}

// playSceneMusic 只保留跨版本的場景分類；實際軌號由 versioned audio.json 提供。
func (g *Game) playSceneMusic(cty int) {
	cue := audioCueTown
	for _, castle := range []int{0, 2, 6, 37, 73, 76} {
		if cty == castle {
			cue = audioCueCastle
			break
		}
	}
	if cue == audioCueTown && cty >= 0 && cty < len(mapBlkNum) {
		if b := mapBlkNum[cty]; b == 2 || b == 4 || b == 5 {
			cue = audioCueDungeon
		}
	}
	g.playAudioCue(cue)
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

// countPartyItem:戰鬥可用的消耗品由整個隊伍的個人道具欄共同提供；
// 一般地圖道具選單仍以主角背包為入口，避免把兩種查詢語意混在一起。
func (g *Game) countPartyItem(code int) int {
	n := g.countItem(code)
	for _, m := range g.companions {
		for _, c := range m.Inventory {
			if c == code {
				n++
			}
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

// removePartyItems:依主角→同伴順序消耗隊伍共享的戰鬥消耗品。
func (g *Game) removePartyItems(code, n int) {
	if n <= 0 {
		return
	}
	g.removeItems(code, n)
	for _, m := range g.companions {
		if n <= 0 {
			break
		}
		out := m.Inventory[:0]
		for _, c := range m.Inventory {
			if c == code && n > 0 {
				n--
				continue
			}
			out = append(out, c)
		}
		m.Inventory = out
	}
}

// buildCompanionActors:把同伴 Member 轉成戰鬥即時 actor(數值由職業成長表 + 裝備推導)。
func (g *Game) buildCompanionActors() []*battleActor {
	var out []*battleActor
	for _, m := range g.companions {
		out = append(out, &battleActor{
			name: classNames[m.Class], class: m.Class, level: m.Level(),
			hp: m.CurHP, maxHP: m.MaxHP(), mp: m.CurMP, maxMP: m.MaxMP(),
			atk: m.Atk(g.shop.items), def: m.Def(g.shop.items), agi: m.Agi(),
			spells: m.Spells(),
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
		g.removePartyItems(herbCode, g.battle.usedHerbs)
	}
	if g.battle.result == 1 { // 勝:全隊加 exp/gold
		oldLv := stats.LevelForExp(0, g.heroExp)
		g.heroExp += uint32(g.battle.gotExp)
		g.heroGold += g.battle.gotGold
		newLv := stats.LevelForExp(0, g.heroExp)
		for lv := oldLv + 1; lv <= newLv; lv++ {
			// sub_ed3c：每升一級逐欄擲成長；最大 HP/MP 的增量同時加到目前值，
			// 並非舊實作的「升級全回復」。
			add := stats.GrowLevel(0, lv, &g.heroStat, &g.prng)
			g.heroHP += int(add[stats.HP])
			g.heroMP += int(add[stats.MP])
		}
		for _, m := range g.companions { // 同伴同得 exp，逐級擲持久成長欄
			oc := m.Level()
			m.Exp += uint32(g.battle.gotExp)
			for lv := oc + 1; lv <= m.Level(); lv++ {
				add := stats.GrowLevel(m.Class, lv, &m.Stats, &g.prng)
				m.CurHP += int(add[stats.HP])
				m.CurMP += int(add[stats.MP])
			}
		}
		if g.battle.monID == 0x7c { // 打倒大魔王索瑪 → 破關結局
			g.runFinale()
		}
		if g.battle.gotDrop >= 0 {
			g.inventory = append(g.inventory, g.battle.gotDrop)
			g.noticeCode, g.noticeTimer = g.battle.gotDrop, 120
		}
	}
	g.settleMirrorBattle()
	g.settleBossSurrenderBattle()
	g.settleHostageRescueBattle()
	g.settleStagedBossBattle()
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
			if t.clearStoryFlag != 0 {
				g.setStoryFlag(t.clearStoryFlag, false)
			}
			if t.postRec >= 0 {
				g.dlg.Open(t.postRec)
			}
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
		if g.attractActive && g.attractReady() {
			drawIndexedPCX(g.rgba, g.attractPix[g.attractIndex], g.attractPal[g.attractIndex])
			g.frame.WritePixels(g.rgba)
			return
		}
		// 原版在離開 title splash、進入主選單／姓名／性別／能力確認後，
		// 皆切到 FIRST.SCR 的黑底雙立繪；不能把城堡標題 PCX 繼續
		// 疊在創角視窗後方。ngSplash 才保留 title_image。
		if g.newGame.stage != ngSplash {
			if len(g.newGameConfirmPix) == ScreenW*ScreenH && len(g.newGameConfirmPal) == 16 {
				drawIndexedPCX(g.rgba, g.newGameConfirmPix, g.newGameConfirmPal)
			} else {
				for i := 0; i < ScreenW*ScreenH; i++ {
					o := i * 4
					g.rgba[o], g.rgba[o+1], g.rgba[o+2], g.rgba[o+3] = 0, 0, 0, 255
				}
			}
		} else {
			drawIndexedPCX(g.rgba, g.titlePix, g.titlePal)
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
	if g.lotoBlessed && g.endSeq < 0 && !g.dlg.open && g.endingPix != nil {
		// 結局文字全部關閉後，原版切到 TIT3.P 的固定片尾畫面；
		// 這個分支只使用 pack asset，不以 DQ3 專屬座標或文字重畫。
		drawIndexedPCX(g.rgba, g.endingPix, g.endingPal)
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
	g.drawPhoenixAltars(camX, camY)
	g.drawTrackedWorldObjects(camX, camY)
	for i := range sc.npcs { // NPC(在視窗內才畫;靜態朝下 frame)
		n := &sc.npcs[i]
		if n.spr == nil || n.x < camX || n.x >= camX+ViewCols || n.y < camY || n.y >= camY+ViewRows {
			continue
		}
		frame := n.facing*dq3data.CharWalk + n.walk
		if frame < 0 || frame >= len(n.spr.Frames) {
			frame = 0
		}
		blitSprite(g.rgba, (n.x-camX)*TileW, (n.y-camY)*TileH, n.spr.Frames[frame], sc.pal)
	}
	// 船(地表):停泊船 tile;主角在船上 → 畫船 tile 取代主角
	if !g.inTown && g.shipOwned && !g.shipAboard &&
		g.shipX >= camX && g.shipX < camX+ViewCols && g.shipY >= camY && g.shipY < camY+ViewRows {
		blitTile(g.rgba, (g.shipX-camX)*TileW, (g.shipY-camY)*TileH, sc.blk.Tile(shipTileFor(0)), sc.pal)
	}
	if !g.inTown && g.phoenixOwned && !g.phoenixAboard &&
		g.phoenixX >= camX && g.phoenixX < camX+ViewCols && g.phoenixY >= camY && g.phoenixY < camY+ViewRows {
		blitTile(g.rgba, (g.phoenixX-camX)*TileW, (g.phoenixY-camY)*TileH,
			sc.blk.Tile(0xfd), sc.pal) // 原版 world renderer 的 vehicle bit2 代換 tile
	}
	if g.remoaru == 0 { // 原版 world flag 0xc000 會略過 party／vehicle occupant render paths。
		if !g.inTown && g.phoenixAboard {
			blitSprite(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH,
				g.phoenix.Frames[g.facing*dq3data.CharWalk+g.walk], sc.pal)
		} else if !g.inTown && g.shipAboard { // 在船上 → 船 tile 取代主角 sprite
			blitTile(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH, sc.blk.Tile(shipTileFor(g.facing)), sc.pal)
		} else {
			hero := g.heroSceneSprite()
			if hero != nil {
				frame := g.facing*dq3data.CharWalk + g.walk
				if frame < 0 || frame >= len(hero.Frames) {
					frame = 0
				}
				blitSprite(g.rgba, (g.px-camX)*TileW, (g.py-camY)*TileH,
					hero.Frames[frame], sc.pal)
			}
			g.drawPartyFollowers(g.rgba, camX, camY, sc.pal)
		}
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
	g.drawChurch(g.rgba, white)
	g.drawBossSurrenderChoice(g.rgba, white)
	g.drawStagedBossChoice(g.rgba, white)
	g.drawTemporaryRoleChoice(g.rgba, white)
	g.drawSoloChallengeChoice(g.rgba, white)
	g.drawSequenceGateChoice(g.rgba, white)
	g.drawChoiceItemExchangeChoice(g.rgba, white)
	g.drawHostageRescueChoice(g.rgba, white)
	g.drawReclass(g.rgba, white)
	g.drawSettlementFounderChoice(g.rgba, white)
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
	if g.tavern.active { // 酒館 2F 登錄所(創角)
		g.tavern.draw(g.rgba, white)
	}
	if g.recruit.active { // 酒場 1F 招募(roster↔companions)
		g.drawRecruit(g.rgba, white)
	}
	g.drawFieldSpell(g.rgba, white)
	g.input.touch.draw(g.rgba) // 觸控控制疊在最上層(有觸控過才顯示)
	g.frame.WritePixels(g.rgba)
}

func drawIndexedPCX(rgba []byte, pix []uint8, pal []dq3data.Color) {
	for i, idx := range pix {
		var c dq3data.Color
		if int(idx) < len(pal) {
			c = pal[idx]
		}
		o := i * 4
		if o+3 >= len(rgba) {
			break
		}
		rgba[o], rgba[o+1], rgba[o+2], rgba[o+3] = c.R, c.G, c.B, 255
	}
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
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		return nil, fmt.Errorf("load builtin dq3 game pack: %w", err)
	}
	return NewGameWithPack(assets, music, pack)
}

// NewGameWithPack 使用已嚴格驗證的外部 game pack 建立遊戲。桌面入口可藉此修改 JSON
// 而不重新編譯；測試與 mobile 預設仍走 NewGame 的 canonical embedded dq3_cht pack。
func NewGameWithPack(assets fs.FS, music fs.FS, pack *gamepack.Pack) (*Game, error) {
	if pack == nil {
		return nil, fmt.Errorf("game pack is nil")
	}
	battleRefs, ok := pack.BattleTextRefs()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_texts")
	}
	battleMessageLayout, ok := pack.BattleMessageWindowLayout()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_message")
	}
	battleCommandLayout, ok := pack.BattleCommandLayout()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_command")
	}
	battleEnemyLayout, ok := pack.BattleEnemyLayout()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_enemy")
	}
	battleSceneLayout, ok := pack.BattleSceneLayout()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_scene")
	}
	battleCommandLabels, ok := pack.BattleCommandLabels()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.battle_command_labels")
	}
	fieldCommandLabels, ok := pack.FieldCommandLabels()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.field_command_labels")
	}
	newGameLabels, ok := pack.NewGameLabels()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.new_game_labels")
	}
	newGameGeometry, ok := pack.NewGameGeometry()
	if !ok {
		return nil, fmt.Errorf("game pack missing interface.new_game_geometry")
	}
	battleDefs := make(map[string]battleTextDefinition, len(battleRefs.IDs()))
	for role, id := range battleRefs.IDs() {
		if id == "" {
			return nil, fmt.Errorf("game pack battle_texts.%s is empty", role)
		}
		def, defined := pack.TextDefinition(id)
		if !defined || def == nil {
			return nil, fmt.Errorf("game pack battle_texts.%s references unknown text %q", role, id)
		}
		codes, defined := pack.TextGlyphCodes(id)
		if !defined || len(codes) == 0 {
			return nil, fmt.Errorf("game pack battle_texts.%s has no glyph stream", role)
		}
		battleDefs[role] = battleTextDefinition{
			value: def.Value, glyph: codes,
			columns: def.Layout.Columns, lines: def.Layout.LinesPerPage,
		}
	}
	heroEquipment, ok := pack.NewGamePlayerEquipment()
	if !ok {
		return nil, fmt.Errorf("game pack missing new-game hero equipment")
	}
	memberEquipment, ok := pack.RegisteredPartyMemberEquipment()
	if !ok {
		return nil, fmt.Errorf("game pack missing registered-member equipment")
	}
	partySpriteAsset, heroSpriteEntry, ok := pack.PartySprite(0, 0)
	if !ok || partySpriteAsset == "" {
		return nil, fmt.Errorf("game pack missing male hero party sprite")
	}
	ld := &loader{fsys: assets}
	pal := dq3data.DecodePalette(ld.read("DQ3.PAL"), 256) // 城鎮/地表共用(tile 只用色 0..15)
	g := &Game{
		rgba: make([]byte, ScreenW*ScreenH*4), input: newInput(), cfg: config.Default(), pack: pack,
	}
	g.dlg.layout = pack.DialogueWindowLayout()
	g.battle.setTextDefinitions(battleDefs)
	g.battle.setMessageLayout(battleMessageLayout)
	g.battle.setCommandLayout(battleCommandLayout)
	g.battle.setEnemyLayout(battleEnemyLayout)
	g.battle.setSceneLayout(battleSceneLayout)
	g.battle.setCommandLabels(battleCommandLabels)
	g.cmd.setLabels(fieldCommandLabels)
	g.newGame.setLabels(newGameLabels)
	g.newGame.setGeometry(newGameGeometry)
	g.tavern.setLabels(newGameLabels)
	g.tavern.setGeometry(newGameGeometry)
	g.tavern.equipment = memberEquipment
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
	mstBLS := ld.read(partySpriteAsset)
	if ld.err != nil {
		return nil, ld.err
	}
	g.assets, g.worldPal, g.manBLS, g.mstBLS = assets, pal, manBLS, mstBLS
	g.partySpriteCache = make(map[[2]int]*dq3data.CharSprite)
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
	g.recruit.tx = g.dlg.tx                                         // 酒場招募 glyph
	g.hero = dq3data.LoadCharSprite(mstBLS, heroSpriteEntry)
	g.phoenix = dq3data.LoadCharSprite(manBLS, 176) // CTY70 egg b2=48 → (48-4)*4；原版拉米亞 8-frame sprite

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
	encounters, eerr := dq3data.OpenEncounterTables(ld.read("DQ3.EXE"))
	if eerr != nil && ld.err == nil {
		return nil, fmt.Errorf("OpenEncounterTables: %w", eerr)
	}
	g.encounters = encounters

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
	// DQ3.EXE file 0x13a0..0x13b2：新遊戲記住的阿里阿罕地表座標。
	// 真正畫面會在創角完成後由 startOpening 切到 CTY00 sec4@(5,5)。
	g.px, g.py = aliahanWorldX, aliahanWorldY
	g.overPx, g.overPy = aliahanWorldX, aliahanWorldY
	g.heroGold = 0 // 國王的 50G 屬後續謁見 transaction，不在出生時預給
	g.equip = heroEquipment
	g.companions = nil // file 0x1c4e：[0x722]=1，開局只有主角
	g.flags = map[int]bool{}
	g.initStoryBits() // 新遊戲重置 [0x4f70] NPC 可見性旗標
	// msStart 代表「已見國王且已能在酒場建隊」，不可在出生時提前完成。
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
	loadPCXAsset := func(key string) ([]uint8, []dq3data.Color, error) {
		ref, ok := pack.Asset(key)
		if !ok {
			return nil, nil, nil
		}
		raw := ld.read(ref.Path)
		if ld.err != nil {
			return nil, nil, ld.err
		}
		if ref.Size > 0 && int64(len(raw)) != ref.Size {
			return nil, nil, fmt.Errorf("game pack asset %q size mismatch: got %d want %d", key, len(raw), ref.Size)
		}
		if ref.SHA256 != "" {
			got := fmt.Sprintf("%x", sha256.Sum256(raw))
			if got != ref.SHA256 {
				return nil, nil, fmt.Errorf("game pack asset %q sha256 mismatch: got %s want %s", key, got, ref.SHA256)
			}
		}
		pix, pal, w, h, err := dq3data.DecodePCX(raw)
		if err != nil {
			return nil, nil, fmt.Errorf("decode game pack asset %q (%s): %w", key, ref.Path, err)
		}
		if w != ScreenW || h != ScreenH {
			return nil, nil, fmt.Errorf("game pack asset %q has size %dx%d, want %dx%d", key, w, h, ScreenW, ScreenH)
		}
		return pix, pal, nil
	}
	loadRawScreenAsset := func(spec *gamepack.RawScreenAsset) ([]uint8, []dq3data.Color, error) {
		if spec == nil {
			return nil, nil, nil
		}
		ref, ok := pack.Asset(spec.AssetKey)
		if !ok {
			return nil, nil, fmt.Errorf("game pack raw screen %q has no manifest asset", spec.ID)
		}
		raw := ld.read(ref.Path)
		if ld.err != nil {
			return nil, nil, ld.err
		}
		if ref.Size > 0 && int64(len(raw)) != ref.Size {
			return nil, nil, fmt.Errorf("game pack asset %q size mismatch: got %d want %d", spec.AssetKey, len(raw), ref.Size)
		}
		if ref.SHA256 != "" {
			got := fmt.Sprintf("%x", sha256.Sum256(raw))
			if got != ref.SHA256 {
				return nil, nil, fmt.Errorf("game pack asset %q sha256 mismatch: got %s want %s", spec.AssetKey, got, ref.SHA256)
			}
		}
		if spec.Format != "row_interleaved_4bpp" {
			return nil, nil, fmt.Errorf("game pack raw screen %q has unsupported format %q", spec.ID, spec.Format)
		}
		pix, err := dq3data.DecodeRowInterleaved4BPP(raw, spec.Width, spec.Height)
		if err != nil {
			return nil, nil, fmt.Errorf("decode game pack asset %q (%s): %w", spec.AssetKey, ref.Path, err)
		}
		if spec.Width != ScreenW || spec.Height != ScreenH {
			return nil, nil, fmt.Errorf("game pack raw screen %q has size %dx%d, want %dx%d", spec.ID, spec.Width, spec.Height, ScreenW, ScreenH)
		}
		pal := make([]dq3data.Color, len(spec.Palette))
		for i, rgb := range spec.Palette {
			pal[i] = dq3data.Color{R: rgb[0], G: rgb[1], B: rgb[2]}
		}
		return pix, pal, nil
	}
	var assetErr error
	if g.titlePix, g.titlePal, assetErr = loadPCXAsset("title_image"); assetErr != nil {
		return nil, assetErr
	}
	if spec, ok := pack.NewGameConfirmation(); ok {
		if g.newGameConfirmPix, g.newGameConfirmPal, assetErr = loadRawScreenAsset(spec); assetErr != nil {
			return nil, assetErr
		}
	}
	g.showTitle = g.titlePix != nil
	if seq, ok := pack.AttractSequence(); ok {
		g.attractSeq = seq
		g.attractPix = make([][]uint8, len(seq.Frames))
		g.attractPal = make([][]dq3data.Color, len(seq.Frames))
		for i, frame := range seq.Frames {
			g.attractPix[i], g.attractPal[i], assetErr = loadPCXAsset(frame.AssetKey)
			if assetErr != nil {
				return nil, assetErr
			}
			if g.attractPix[i] == nil || len(g.attractPal[i]) == 0 {
				return nil, fmt.Errorf("game pack attract frame %q is unavailable", frame.AssetKey)
			}
		}
	}
	if g.endingPix, g.endingPal, assetErr = loadPCXAsset("ending_image"); assetErr != nil {
		return nil, assetErr
	}
	g.frame = ebiten.NewImage(ScreenW, ScreenH)
	// 注意:不再於此自動續玩——存檔改由標題主選單「載入進度」明確觸發(newgame.go newGameInput
	// ngOptLoad 分支),對齊 docs/36 開場流程(遊戲開始 = 全新主角,不會被舊存檔悄悄蓋掉)。
	applyDebugEnv(g) // debug 環境變數可覆蓋(此時 music/frame 已就緒)
	// 起始配樂(依最終場景;debug 進城已自行換軌)
	if g.showTitle {
		g.playAudioCue(audioCueTitle)
	} else if !g.inTown {
		g.playAudioCue(audioCueField)
	}
	g.renderFrame() // 首幀
	log.Printf("地表 %d×%d / 阿里阿罕 %d×%d(spawn %d,%d、NPC %d)— 城鎮依正式入口與邊界轉場進出、Space 命令窗",
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
	if r := os.Getenv("DQ3_BATTLE"); r != "" { // debug:起手開一場戰鬥(截圖驗證);DQ3_BATTLE_N=群量(預設1)
		var id int
		if _, e := fmt.Sscanf(r, "%d", &id); e == nil {
			n := 1
			if nr := os.Getenv("DQ3_BATTLE_N"); nr != "" {
				fmt.Sscanf(nr, "%d", &n)
			}
			level, maxHP, atk, def, agi := g.heroStats()
			mp := g.heroMaxMP()
			g.heroHP, g.heroMP = maxHP, mp
			g.battle.showInfo = g.cfg.CombatInfo
			g.battle.hurtFxFrames = hurtFxFrames(g.cfg.CombatHurtFx)
			g.battle.startGroup(id, n, 1, heroParams{level: level, curHP: maxHP, maxHP: maxHP, atk: atk, def: def, agi: agi,
				mp: mp, maxMP: mp, spells: g.heroSpells()}, g.buildCompanionActors())
			if os.Getenv("DQ3_SPELLMENU") != "" && len(g.battle.spells) > 0 {
				g.battle.cursor, g.battle.phase = bcSpell, phSpell // 開咒文子選單
			}
		}
	}
}
