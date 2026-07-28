package game

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// newGameTestGame:純狀態 Game(無素材/無 ebiten 主迴圈),只驅動 newGameInput()/newGame 狀態機。
func newGameTestGame() *Game {
	return &Game{input: newInput()}
}

// TestNewGameFlowStageTransitions:標題主選單 → 命名 → 性別 → 能力確認 → 收尾。
func TestNewGameFlowStageTransitions(t *testing.T) {
	g := newGameTestGame()
	if g.newGame.stage != ngSplash {
		t.Fatalf("初始應在 ngSplash,得 %d", g.newGame.stage)
	}
	g.showTitle = true

	g.newGameInput(InputState{DirEdge: -1, Confirm: true}) // 標題壓 A → 開主選單
	if g.newGame.stage != ngMenu {
		t.Fatalf("壓 A 後應進主選單 ngMenu,得 %d", g.newGame.stage)
	}
	if g.newGame.cursor != ngOptStart {
		t.Fatalf("主選單預設游標應在「遊戲開始」(0),得 %d", g.newGame.cursor)
	}

	g.newGameInput(InputState{DirEdge: -1, Confirm: true}) // 選「遊戲開始」
	if g.newGame.stage != ngName {
		t.Fatalf("選遊戲開始後應進命名 ngName,得 %d", g.newGame.stage)
	}

	// 預設=注音模式(原版 docs/36 04);本測試驅動英數格盤 → 先 Toggle 切英數
	g.newGameInput(InputState{DirEdge: -1, Toggle: true})
	if g.newGame.ni.nameZhu {
		t.Fatal("Toggle 後應為英數模式")
	}
	// 英數格盤:cell10='A' → 附加 1 字,再選 OK(niCellOK)完成(非空 → 放行)
	g.newGame.ni.cursor = 10
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if len(g.newGame.ni.nameBuf) != 1 || g.newGame.ni.nameBuf[0] != niCellGlyph(10) {
		t.Fatalf("應附加 1 字到 nameBuf,得 %v", g.newGame.ni.nameBuf)
	}
	g.newGame.ni.cursor = niCellOK
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngGender {
		t.Fatalf("命名完成(非空)後應進性別 ngGender,得 %d", g.newGame.stage)
	}
	if !g.showTitle {
		t.Fatal("性別選擇前仍應在標題流程內(showTitle=true)")
	}

	g.newGame.gs.cursor = 1 // 女性
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngConfirm || !g.showTitle {
		t.Fatalf("性別選定後應先進能力確認，stage=%d title=%v", g.newGame.stage, g.showTitle)
	}
	if g.heroStat[stats.MP] != 9 || g.heroStat[stats.HP] < 9 || g.heroStat[stats.HP] > 15 {
		t.Fatalf("原版 Lv1 能力 transaction 錯：%v", g.heroStat)
	}
	g.newGameInput(InputState{DirEdge: -1, Confirm: true}) // 「是」
	if g.showTitle {
		t.Fatal("能力確認選是後應離開標題流程(showTitle=false)")
	}
	if g.heroGender != 1 {
		t.Errorf("heroGender 應=1(女性),得 %d", g.heroGender)
	}
	if len(g.heroName) != 1 || g.heroName[0] != niCellGlyph(10) {
		t.Errorf("heroGender 選定後應把命名結果寫回 heroName,得 %v", g.heroName)
	}
}

func TestNewGameStatsRejectRestartsCreation(t *testing.T) {
	g := newGameTestGame()
	g.showTitle = true
	g.prng.Seed(1)
	g.newGame.stage = ngGender
	g.newGame.ni.nameBuf = []int{15}
	g.newGame.gs.Init()
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	first := g.heroStat
	g.newGameInput(InputState{DirEdge: 0}) // 移到「否」
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngName {
		t.Fatalf("能力確認選否應回整個創角命名，得 stage=%d", g.newGame.stage)
	}
	if len(g.newGame.ni.nameBuf) != 0 {
		t.Fatalf("重開創角應清姓名，得 %v", g.newGame.ni.nameBuf)
	}
	if first == (stats.Values{}) {
		t.Fatal("進能力確認前應已擲出 Lv1 record")
	}
}

func TestHeroLevelUpAddsGrowthWithoutFullHeal(t *testing.T) {
	g := newGameTestGame()
	g.prng.Seed(0x1357)
	g.rollHeroLevelOne()
	oldMax := int(g.heroStat[stats.HP])
	g.heroHP, g.heroMP = 1, 0 // 受傷且耗盡 MP
	g.battle.heroHP, g.battle.heroMP = 1, 0
	g.battle.result = 1
	g.battle.gotExp = int(stats.ExpForLevel(0, 2))
	g.onBattleEnd()
	if stats.LevelForExp(0, g.heroExp) != 2 || int(g.heroStat[stats.HP]) <= oldMax {
		t.Fatalf("升級未更新持久能力：lv=%d oldHP=%d new=%v",
			stats.LevelForExp(0, g.heroExp), oldMax, g.heroStat)
	}
	if g.heroHP >= int(g.heroStat[stats.HP]) {
		t.Fatalf("原版升級只加 HP 增量，不應把受傷角色全回復：cur=%d max=%d",
			g.heroHP, g.heroStat[stats.HP])
	}
}

// TestNewGameFlowNameRequiredGate:名字空時選「完成」不應放行(對齊 docs/15「[0x270a]>=1 才放行」),
// 應留在 ngName 且仍在標題流程內——這是 FLOW-GAP A3 的核心驗收點(主角命名必填)。
func TestNewGameFlowNameRequiredGate(t *testing.T) {
	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage = ngName
	g.newGame.ni.Init()
	g.newGameInput(InputState{DirEdge: -1, Toggle: true}) // 預設注音 → 切英數驅動格盤
	g.newGame.ni.cursor = niCellOK                        // 空名字直接選「完成」

	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngName {
		t.Fatalf("名字空時選完成不應放行,應仍在 ngName,得 %d", g.newGame.stage)
	}
	if !g.showTitle {
		t.Fatal("名字空時不應離開標題流程")
	}

	// 補一個字後再選完成 → 才放行
	g.newGame.ni.cursor = 0 // 字元格'0'
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	g.newGame.ni.cursor = niCellOK
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngGender {
		t.Fatalf("補字後名字非空,選完成應放行進性別,得 stage=%d", g.newGame.stage)
	}
}

// TestNewGameFlowCancelBackToMenu:命名畫面(英數格盤)按 Cancel 應退回主選單(對齊酒館既有
// 「英數層 Cancel = 回上一畫面」語意,共用元件 NameInput.input 的 canceled 分支)。
func TestNewGameFlowCancelBackToMenu(t *testing.T) {
	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage = ngName
	g.newGame.ni.Init()
	g.newGameInput(InputState{DirEdge: -1, Toggle: true}) // 預設注音 → 切英數(英數層 Cancel 才回上一畫面)

	g.newGameInput(InputState{DirEdge: -1, Cancel: true})
	if g.newGame.stage != ngMenu {
		t.Fatalf("命名畫面按 Cancel 應回主選單,得 %d", g.newGame.stage)
	}
}

// TestNewGameFlowGenderCancelBackToName:性別畫面按 Cancel 應退回命名畫面(可修改),
// 對齊酒館既有 tavGender Cancel 語意(回命名、游標停 OK 格)。
func TestNewGameFlowGenderCancelBackToName(t *testing.T) {
	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage = ngGender
	g.newGame.ni.nameBuf = []int{15}
	g.newGame.gs.Init()

	g.newGameInput(InputState{DirEdge: -1, Cancel: true})
	if g.newGame.stage != ngName {
		t.Fatalf("性別畫面按 Cancel 應回命名 ngName,得 %d", g.newGame.stage)
	}
	if len(g.newGame.ni.nameBuf) != 1 {
		t.Fatalf("回命名畫面不應清掉已輸入的名字,得 %v", g.newGame.ni.nameBuf)
	}
}

// TestNewGameFlowZhuyinTogglePutsCellIntoComposer:主角命名畫面內 Toggle(注音↔英數)與注音盤
// 選格,應如實落到共用 NameInput 的 zh.Composer(驗證酒館/主角創建共用同一份 widget,行為一致)。
func TestNewGameFlowZhuyinTogglePutsCellIntoComposer(t *testing.T) {
	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage = ngName
	g.newGame.ni.Init()

	if !g.newGame.ni.nameZhu {
		t.Fatal("預設應為注音模式(原版 docs/36 04:新遊戲命名即注音盤)")
	}
	g.newGame.ni.zh.Cursor = 0 // 聲母 cell0(ㄅ)
	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.ni.zh.Sh != 1 {
		t.Errorf("選聲母 cell0 應設 zh.Sh=1,得 %d", g.newGame.ni.zh.Sh)
	}

	g.newGameInput(InputState{DirEdge: -1, Toggle: true}) // 切英數
	if g.newGame.ni.nameZhu {
		t.Fatal("Toggle 應切到英數模式")
	}
}

// TestNewGameFlowLoadOption:主選單選「載入進度」應直接 g.Load() 讀存檔(含主角名/性別)並離開
// 標題流程,不經過命名/性別畫面——對齊 docs/36「▶遊戲開始/載入進度」兩條分岔路。
func TestNewGameFlowLoadOption(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "ng-save.json")
	t.Setenv("DQ3_SAVE", savePath)

	src := &Game{heroExp: 999, heroGold: 42, heroName: []int{15, 16, 17}, heroGender: 1}
	if err := src.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage, g.newGame.cursor = ngMenu, ngOptLoad

	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.showTitle {
		t.Fatal("選「載入進度」後應離開標題流程")
	}
	if g.heroExp != 999 || g.heroGold != 42 {
		t.Fatalf("應載入存檔內容,得 exp=%d gold=%d", g.heroExp, g.heroGold)
	}
	if len(g.heroName) != 3 || g.heroName[0] != 15 || g.heroGender != 1 {
		t.Fatalf("應載入主角名/性別,得 name=%v gender=%d", g.heroName, g.heroGender)
	}
}

// TestNewGameFlowStartOptionIgnoresStaleSave:主選單選「遊戲開始」不應被舊存檔悄悄蓋掉
// (對齊移除 NewGame 自動續玩後的新語意:遊戲開始 = 全新主角,heroName 應仍是選單前的初始值)。
func TestNewGameFlowStartOptionIgnoresStaleSave(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DQ3_SAVE", filepath.Join(dir, "stale.json"))
	stale := &Game{heroExp: 12345, heroName: []int{1, 2, 3}}
	if err := stale.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	g := newGameTestGame()
	g.showTitle = true
	g.newGame.stage, g.newGame.cursor = ngMenu, ngOptStart

	g.newGameInput(InputState{DirEdge: -1, Confirm: true})
	if g.newGame.stage != ngName {
		t.Fatalf("選遊戲開始應進命名畫面,得 stage=%d", g.newGame.stage)
	}
	if g.heroExp == 12345 {
		t.Fatal("選「遊戲開始」不應載入舊存檔(heroExp 不該被悄悄改成存檔值)")
	}
}

// TestSaveRoundTripHeroNameGender:heroName/heroGender 經 Save()→Load() round-trip 一致
// (存檔 JSON round-trip,對齊既有 TestSaveRoundTrip 的驗收模式,補上本批新增欄位)。
func TestSaveRoundTripHeroNameGender(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("DQ3_SAVE", filepath.Join(dir, "rt.json"))

	src := &Game{heroName: []int{113, 689, 488, 711}, heroGender: 1, heroGold: 7,
		heroStat: stats.Values{8, 4, 4, 12, 9, 7, 7}, heroHP: 5, heroMP: 3}
	if err := src.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	dst := &Game{}
	if err := dst.Load(); err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(dst.heroName) != len(src.heroName) {
		t.Fatalf("heroName round-trip 長度不符:存 %v 讀 %v", src.heroName, dst.heroName)
	}
	for i, g := range src.heroName {
		if dst.heroName[i] != g {
			t.Fatalf("heroName round-trip 內容不符:存 %v 讀 %v", src.heroName, dst.heroName)
		}
	}
	if dst.heroGender != src.heroGender {
		t.Fatalf("heroGender round-trip 不符:存 %d 讀 %d", src.heroGender, dst.heroGender)
	}
	if dst.heroStat != src.heroStat || dst.heroHP != 5 || dst.heroMP != 3 {
		t.Fatalf("持久能力/目前 HP MP round-trip 不符:存 %v/%d/%d 讀 %v/%d/%d",
			src.heroStat, src.heroHP, src.heroMP, dst.heroStat, dst.heroHP, dst.heroMP)
	}
	t.Logf("heroName/heroGender round-trip ✓:name=%v gender=%d", dst.heroName, dst.heroGender)
}

func TestSaveRoundTripPreservesMemberNameAndStats(t *testing.T) {
	src := &Game{
		companions: []*Member{{
			Name: []int{101, 202}, Class: 3, Gender: 1, Exp: 42,
			Stats: stats.Values{8, 7, 6, 20, 17, 9, 10}, CurHP: 13, CurMP: 11, Armor: 0x25,
		}},
		roster: []*Member{{
			Name: []int{303, 404}, Class: 4, Gender: 0,
			Stats: stats.Values{5, 4, 7, 12, 18, 10, 9}, CurHP: 12, CurMP: 18, Armor: 0x25,
		}},
	}
	b, err := encodeSave(src.snapshot())
	if err != nil {
		t.Fatalf("encodeSave: %v", err)
	}
	s, err := decodeSave(b)
	if err != nil {
		t.Fatalf("decodeSave: %v", err)
	}
	dst := &Game{}
	dst.restore(s)
	if !reflect.DeepEqual(dst.companions[0].Name, src.companions[0].Name) ||
		dst.companions[0].Stats != src.companions[0].Stats ||
		!reflect.DeepEqual(dst.roster[0].Name, src.roster[0].Name) ||
		dst.roster[0].Stats != src.roster[0].Stats {
		t.Fatalf("同伴存檔往返遺失自訂姓名/持久能力：src=%+v/%+v dst=%+v/%+v",
			src.companions[0], src.roster[0], dst.companions[0], dst.roster[0])
	}
}
