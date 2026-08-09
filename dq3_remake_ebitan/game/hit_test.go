package game

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// TestHitListAt:hitList.at 命中/未命中/邊界(半開區間 [x,x+w))。
func TestHitListAt(t *testing.T) {
	var h hitList
	h.add(10, 20, 30, 15, 7) // box: x 10..40, y 20..35, idx 7

	if got := h.at(10, 20); got != 7 { // 左上角(閉)
		t.Errorf("左上角應命中 idx7,得 %d", got)
	}
	if got := h.at(39, 34); got != 7 { // 右下內側
		t.Errorf("右下內側應命中 idx7,得 %d", got)
	}
	if got := h.at(40, 20); got != -1 { // x 上界(開)→ 不命中
		t.Errorf("x 上界應不命中,得 %d", got)
	}
	if got := h.at(10, 35); got != -1 { // y 上界(開)→ 不命中
		t.Errorf("y 上界應不命中,得 %d", got)
	}
	if got := h.at(9, 20); got != -1 { // 左側外
		t.Errorf("左側外應不命中,得 %d", got)
	}
	if got := h.at(0, 0); got != -1 {
		t.Errorf("原點(無 box 覆蓋)應不命中,得 %d", got)
	}

	h.reset()
	if got := h.at(10, 20); got != -1 {
		t.Errorf("reset 後應無 box,得 %d", got)
	}
}

// newHitTestRGBA:640×350 RGBA 緩衝(供各選單 draw() 用,不需真素材)。
func newHitTestRGBA() []byte { return make([]byte, ScreenW*ScreenH*4) }

func installBattleDrawPack(t *testing.T, b *Battle) {
	t.Helper()
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	command, ok := p.BattleCommandLayout()
	if !ok {
		t.Fatal("builtin pack missing battle command layout")
	}
	enemy, ok := p.BattleEnemyLayout()
	if !ok {
		t.Fatal("builtin pack missing battle enemy layout")
	}
	labels, ok := p.BattleCommandLabels()
	if !ok {
		t.Fatal("builtin pack missing battle command labels")
	}
	b.setCommandLayout(command)
	b.setEnemyLayout(enemy)
	b.setCommandLabels(labels)
}

// TestCmdMenuDrawHits:命令窗 draw 後應有 6 個 box;點第 3 格(index2「狀況」所在列)回 index2。
// 幾何取自 cmdmenu.go draw():cx=x+c*cw、cy=y+rh+r*rh,cw=5*GlyphPx=80、rh=18,呼叫點 (x,y)=(48,32)。
func TestCmdMenuDrawHits(t *testing.T) {
	m := &CmdMenu{tx: &dq3data.Text{}}
	m.Open()
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	m.draw(rgba, white, yellow, 48, 32)

	if len(m.hits) != cmdCount {
		t.Fatalf("draw 後 hits 應有 %d 筆,得 %d", cmdCount, len(m.hits))
	}
	// index2(狀況,r=1,c=0):cx=48,cy=32+18+18=68
	if idx := m.hits.at(50, 70); idx != cmdStatus {
		t.Errorf("點 (50,70) 應命中 index%d(狀況),得 %d", cmdStatus, idx)
	}
	// index5(調查,r=2,c=1):cx=48+80=128,cy=32+18+36=86
	if idx := m.hits.at(130, 90); idx != cmdExamine {
		t.Errorf("點 (130,90) 應命中 index%d(調查),得 %d", cmdExamine, idx)
	}
}

// TestCmdMenuTapSelectsAndConfirms:透過 Game.Update 的命令窗分支,點格應等同「游標移過去 + A 確定」——
// 用「對話」格(index0,面前無 NPC)驗證選定後窗口關閉(selectCommand 一律關窗)。
func TestCmdMenuTapSelectsAndConfirms(t *testing.T) {
	m := &CmdMenu{tx: &dq3data.Text{}}
	m.Open()
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	m.draw(rgba, white, yellow, 48, 32)

	if idx := m.hits.at(130, 90); idx != cmdExamine {
		t.Fatalf("前置條件:點 (130,90) 應先命中 cmdExamine,得 %d", idx)
	}
	m.cursor = 0 // 確認點選會覆蓋既有游標,而非只靠方向鍵累加
	confirm := false
	if idx := m.hits.at(130, 90); idx >= 0 {
		m.cursor, confirm = idx, true
	}
	if m.cursor != cmdExamine || !confirm {
		t.Errorf("點格應移游標到 %d 並視同確定,得 cursor=%d confirm=%v", cmdExamine, m.cursor, confirm)
	}
}

// TestTavernAlnumGridHits:酒館英數格盤(niCells=38 格,9 欄)draw 後 hits 筆數與命中格對齊 draw() 幾何。
func TestTavernAlnumGridHits(t *testing.T) {
	tv := &Tavern{tx: &dq3data.Text{}, active: true, stage: tavName}
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	tv.draw(rgba, white)

	if len(tv.ni.hits) != niCells {
		t.Fatalf("英數格盤 draw 後 hits 應有 %d 筆,得 %d", niCells, len(tv.ni.hits))
	}
	// cell 10('A',col1,row1):x=64+1*40=104,y=96+1*22=118
	if idx := tv.ni.hits.at(106, 120); idx != 10 {
		t.Errorf("點 (106,120) 應命中 cell10('A'),得 %d", idx)
	}
	// 點格後應可直接經 input() 附加該字元
	if _, closed := tv.input(InputState{DirEdge: -1, Tapped: true, TapX: 106, TapY: 120}); closed {
		t.Fatal("點英數格不應關閉酒館")
	}
	if len(tv.ni.nameBuf) != 1 || tv.ni.nameBuf[0] != niCellGlyph(10) {
		t.Errorf("點 cell10 應附加對應字元到 nameBuf,得 %v", tv.ni.nameBuf)
	}
	if tv.ni.cursor != 10 {
		t.Errorf("點格應把游標移到該格,得 cursor=%d", tv.ni.cursor)
	}
}

// TestZhuyinBoardTapMovesAndSelects:注音盤(42 格)點格應移游標到該格並等同確定(選一個聲/介/韻/調)。
func TestZhuyinBoardTapMovesAndSelects(t *testing.T) {
	tv := &Tavern{tx: &dq3data.Text{}, active: true, stage: tavName}
	tv.ni.nameZhu = true
	tv.ni.zh.Init()
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	tv.draw(rgba, white)

	if len(tv.ni.hits) == 0 {
		t.Fatal("注音盤 draw 後應有 hits")
	}
	// cell0(聲母 ㄅ):x=64,y=96
	idx := tv.ni.hits.at(66, 98)
	if idx != 0 {
		t.Fatalf("點 (66,98) 應命中 cell0,得 %d", idx)
	}
	tv.input(InputState{DirEdge: -1, Tapped: true, TapX: 66, TapY: 98})
	if tv.ni.zh.Sh != 1 {
		t.Errorf("點注音盤 cell0(聲母)應選定 Sh=1,得 %d", tv.ni.zh.Sh)
	}
}

// TestNewGameMenuDrawHits:標題主選單(遊戲開始/載入進度)draw 後 hits 對齊幾何;點「載入進度」
// 應移游標並等同確定(經 newGameInput 觸發 g.Load(),FLOW-GAP A2 的 P2 直接點選路徑)。
// 幾何取自 newgame.go draw():bx=190,by=140,bw=200;row i 的 x=bx+60=250,y=by+40+i*22;
// hits.add(x-18,y-3,120,18,i)。
func TestNewGameMenuDrawHits(t *testing.T) {
	g := &Game{input: newInput()}
	g.dlg.tx = &dq3data.Text{}
	g.showTitle = true
	g.newGame.stage = ngMenu
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	g.newGame.draw(rgba, g.dlg.tx, white, yellow)

	if len(g.newGame.hits) != ngOptCount {
		t.Fatalf("主選單 draw 後 hits 應有 %d 筆,得 %d", ngOptCount, len(g.newGame.hits))
	}
	// index1(載入進度):x=190+60=250,y=140+40+22=202 → box (232,199,120,18)
	idx := g.newGame.hits.at(240, 205)
	if idx != ngOptLoad {
		t.Fatalf("點 (240,205) 應命中 index%d(載入進度),得 %d", ngOptLoad, idx)
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "menu-tap.json")) // 無存檔,Load() 靜默略過
	g.newGameInput(InputState{DirEdge: -1, Tapped: true, TapX: 240, TapY: 205})
	if g.newGame.cursor != ngOptLoad {
		t.Errorf("點列應移游標到 ngOptLoad,得 %d", g.newGame.cursor)
	}
	if g.showTitle {
		t.Error("點「載入進度」應等同確定,離開標題流程")
	}
}

// TestSettingsDrawHitsAndTap:設定選單 6 列 draw 後 hits 對齊 rowy;點列應移游標並等同切值(對齊右/Enter)。
func TestSettingsDrawHitsAndTap(t *testing.T) {
	g := &Game{cfg: config.Default()}
	g.settings.tx = &dq3data.Text{}
	g.settings.Open()
	rgba := newHitTestRGBA()
	g.settings.draw(rgba, g.cfg)

	if len(g.settings.hits) != setRowCount {
		t.Fatalf("設定選單 draw 後 hits 應有 %d 筆,得 %d", setRowCount, len(g.settings.hits))
	}
	// row=setRowVolume 的 rowy = by+114 = 48+114 = 162(bx=150,by=48)
	wantRow := setRowVolume
	rowy := [setRowCount]int{48 + 50, 48 + 82, 48 + 114, 48 + 146, 48 + 178, 48 + 210}
	px, py := 150+20, rowy[wantRow]+2
	if idx := g.settings.hits.at(px, py); idx != wantRow {
		t.Fatalf("點 (%d,%d) 應命中 row%d,得 %d", px, py, wantRow, idx)
	}
	before := g.cfg.MusicVolume
	g.settingsInput(InputState{DirEdge: -1, Tapped: true, TapX: px, TapY: py})
	if g.settings.cursor != wantRow {
		t.Errorf("點列應移游標到 row%d,得 %d", wantRow, g.settings.cursor)
	}
	if g.cfg.MusicVolume != before+10 && !(before == 100 && g.cfg.MusicVolume == 0) {
		t.Errorf("點列應等同切值(音量 +10),得 %d(原 %d)", g.cfg.MusicVolume, before)
	}
}

// TestShopDrawHitsAndTap:商店貨架 draw 後 hits 筆數對齊 codes;點列應移游標並等同 A 購買(回 buyCode)。
func TestShopDrawHitsAndTap(t *testing.T) {
	items, err := dq3data.OpenItems(asset(t, "ITEM.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	s := &Shop{tx: &dq3data.Text{}, items: items}
	s.open([]int{5, 8, 13})
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	s.draw(rgba, 999, white)

	if len(s.hits) != 3 {
		t.Fatalf("3 品貨架 draw 後 hits 應有 3 筆,得 %d", len(s.hits))
	}
	// row1(index1):y=40+22=62
	if idx := s.hits.at(30, 64); idx != 1 {
		t.Fatalf("點 (30,64) 應命中 index1,得 %d", idx)
	}
	buy, closed := s.input(InputState{DirEdge: -1, Tapped: true, TapX: 30, TapY: 64})
	if closed {
		t.Fatal("點貨架列不應關閉商店")
	}
	if buy != 8 {
		t.Errorf("點 index1 應等同購買 codes[1]=8,得 buyCode=%d", buy)
	}
	if s.cursor != 1 {
		t.Errorf("點列應移游標到 index1,得 %d", s.cursor)
	}
}

// TestBattleCommandDrawHits:無咒文的首位可行動者使用 rec442 四項指令窗。
// 只驗 hit-box 幾何與命中 index(不經 b.input() 觸發 execTurn——那需要 b.mons 等完整開戰狀態,
// 屬 TestPlaythroughBattle 的整合覆蓋範圍,這裡只測 P2 直接點選這條路徑本身)。
func TestBattleCommandDrawHits(t *testing.T) {
	b := &Battle{tx: &dq3data.Text{}, phase: phCommand}
	installBattleDrawPack(t, b)
	rgba := newHitTestRGBA()
	b.draw(rgba, nil)

	if len(b.hits) != 4 {
		t.Fatalf("無咒文首位角色應有4項指令,得 %d", len(b.hits))
	}
	// index bcItem(3):pack x=144,y=248 → row y=248+16+3*16=312
	tapIdx := b.hits.at(150, 314)
	if tapIdx != bcItem {
		t.Fatalf("點 (130,310) 應命中 index%d(道具),得 %d", bcItem, tapIdx)
	}
	confirm := false
	if tapIdx >= 0 {
		b.cursor, confirm = tapIdx, true
	}
	if b.cursor != bcItem || !confirm {
		t.Errorf("點格應移游標到 %d 並視同確定,得 cursor=%d confirm=%v", bcItem, b.cursor, confirm)
	}
}

// TestBattleSpellDrawHits:咒文子選單(2 咒)draw 後 hits 筆數對齊 spells,點格移 spellCursor。
func TestBattleSpellDrawHits(t *testing.T) {
	b := &Battle{tx: &dq3data.Text{}, phase: phSpell, spells: []int{121, 130}}
	installBattleDrawPack(t, b)
	rgba := newHitTestRGBA()
	b.draw(rgba, nil)

	if len(b.hits) != 2 {
		t.Fatalf("2 咒 draw 後 hits 應有 2 筆,得 %d", len(b.hits))
	}
	// index1:y=248+16+16=280
	if idx := b.hits.at(150, 282); idx != 1 {
		t.Fatalf("點 (150,282) 應命中 index1,得 %d", idx)
	}
}

func TestBattleSpellDrawScrollsToLateSpell(t *testing.T) {
	b := &Battle{
		tx: &dq3data.Text{}, phase: phSpell,
		spells:      []int{121, 122, 123, 124, 125, 126, 180},
		spellCursor: 6,
	}
	installBattleDrawPack(t, b)
	rgba := newHitTestRGBA()
	b.draw(rgba, nil)

	if len(b.hits) != 5 {
		t.Fatalf("捲動咒文窗應保持5個可點列，得 %d", len(b.hits))
	}
	if got := b.hits[4].idx; got != 6 {
		t.Fatalf("最後可見列應指向 rec180 的原陣列 index6，得 %d", got)
	}
}

func TestBattleTargetDrawHits(t *testing.T) {
	emptySprite := &dq3data.MonsterSprite{
		W: 16, H: 16,
		Px: make([][]uint8, 16), Opaque: make([][]bool, 16),
	}
	for i := 0; i < 16; i++ {
		emptySprite.Px[i] = make([]uint8, 16)
		emptySprite.Opaque[i] = make([]bool, 16)
	}
	b := &Battle{
		tx: &dq3data.Text{}, spr: emptySprite, phase: phTargetEnemy,
		heroHP: 10, enemies: []enemyUnit{{hp: 10}, {hp: 10}, {hp: 10}},
		targetCursor: 1,
	}
	installBattleDrawPack(t, b)
	b.draw(newHitTestRGBA(), nil)
	if len(b.hits) != 3 {
		t.Fatalf("三敵目標畫面應有3個怪物 hit-box，得 %d", len(b.hits))
	}
	// 第二隻中心約 x=320、y=232。
	if idx := b.hits.at(320, 228); idx != 1 {
		t.Fatalf("點第二隻怪應命中 target position1，得 %d", idx)
	}

	b.phase = phTargetAlly
	b.companions = []*battleActor{{hp: 10}}
	b.targetCursor = 1
	b.draw(newHitTestRGBA(), nil)
	if len(b.hits) != 2 {
		t.Fatalf("隊長+同伴目標畫面應有2個狀態欄 hit-box，得 %d", len(b.hits))
	}
	if idx := b.hits.at(240, 20); idx != 1 {
		t.Fatalf("點同伴狀態欄應命中 target position1，得 %d", idx)
	}
}

// TestPanelItemsDrawHits:道具清單(panel.go drawItems)draw 後 hits 筆數對齊 inventory,點列移 panelCursor。
func TestPanelItemsDrawHits(t *testing.T) {
	items, err := dq3data.OpenItems(asset(t, "ITEM.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{}
	g.dlg.tx = &dq3data.Text{}
	g.shop.items = items
	g.inventory = []int{5, 8, 13}
	rgba := newHitTestRGBA()
	white := dq3data.Color{R: 248, G: 248, B: 248}
	g.drawItems(rgba, white)

	if len(g.panelHits) != 3 {
		t.Fatalf("3 品道具清單 draw 後 hits 應有 3 筆,得 %d", len(g.panelHits))
	}
	// index2:y=56+2*22=100
	if idx := g.panelHits.at(50, 102); idx != 2 {
		t.Fatalf("點 (50,102) 應命中 index2,得 %d", idx)
	}
}
