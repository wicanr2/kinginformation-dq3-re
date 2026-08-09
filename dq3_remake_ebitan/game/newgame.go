package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 標題主選單 + 主角創建(FLOW-GAP A2/A3/A4,docs/36 §開場流程 + docs/15 命名子系統):
// 標題(壓 A/Enter)→ 主選單「▶遊戲開始 / 載入進度」→ 遊戲開始 → 主角注音/英數命名(必填,
// 完成放行條件=非空,對齊 docs/15「[0x270a]>=1 才放行」)→ 性別 →
// Lv1 能力生成/「這個人可以嗎？」→ 開始新遊戲。
// 命名/性別 widget 與酒館招募共用 NameInput/GenderSelect(nameinput.go),不重複兩份邏輯;
// 差異只在完成放行 gate:主角創建強制非空(下方 ngName case),酒館允許空名回退職業名。
// 出生點由 startOpening 載入 CTY00 sec4；能力確認背景由 game-pack 的 raw screen 提供，
// stat panel 本身仍是共用 renderer，逐像素幾何對拍另列 V3 工作。
const (
	ngSplash  = iota // 標題背景(壓任意鍵開主選單)
	ngMenu           // 主選單:遊戲開始 / 載入進度
	ngName           // 主角命名
	ngGender         // 主角性別
	ngConfirm        // 原版 sub_1c01..1c84：能力確認（否→整個創角重來）
)

const (
	ngOptStart = iota // 遊戲開始
	ngOptLoad         // 載入進度
	ngOptCount
)

// 主選單標籤 glyph(D3TXT00.FON,經 docs/data/glyph_unicode_map.json 驗證,對齊 docs/36 實機截圖
// 「勇者鬥惡龍 / ▶遊戲開始 / 載入進度」):標題 5 字、選項各 4 字。
var (
	ngTitleLabel = []int{106, 187, 207, 710, 179} // 勇者鬥惡龍
	ngOptLabels  = [ngOptCount][]int{
		{113, 689, 488, 711}, // 遊戲開始
		{712, 519, 696, 283}, // 載入進度
	}
)

// NewGameFlow 是標題主選單 + 主角創建的 modal 狀態機。
type NewGameFlow struct {
	stage         int
	cursor        int // 主選單游標(0/1)
	hits          hitList
	ni            NameInput
	gs            GenderSelect
	confirmCursor int // 0=是、1=否
	preview       stats.Values
	previewDef    int
	previewGender int
}

// openMenu:標題壓鍵後開主選單,游標歸零。
func (nf *NewGameFlow) openMenu() { nf.stage, nf.cursor = ngMenu, 0 }

// newGameInput:一幀輸入(僅 g.showTitle==true 且 g.settings.open==false 時呼叫)。
// 依 nf.stage 分派;選定「遊戲開始」進命名、命名完成(非空)進性別、性別選定寫回主角 → 收尾
// (g.showTitle=false,交還一般 Update() 流程,此批先落在現行出生點)。
func (g *Game) newGameInput(in InputState) {
	nf := &g.newGame
	switch nf.stage {
	case ngSplash:
		if in.Settings || in.CtxTap {
			g.settings.Open()
		} else if in.Confirm || in.Enter {
			nf.openMenu()
		}
	case ngMenu:
		tapIdx := -1
		if in.Tapped {
			tapIdx = nf.hits.at(in.TapX, in.TapY)
		}
		confirm := in.Confirm || in.Enter
		if tapIdx >= 0 {
			nf.cursor, confirm = tapIdx, true
		}
		switch {
		case in.Settings || in.CtxTap:
			g.settings.Open()
		case confirm && nf.cursor == ngOptStart:
			nf.ni.Init()
			nf.stage = ngName
		case confirm && nf.cursor == ngOptLoad:
			_ = g.Load() // 載入進度:讀存檔(對齊 docs/36「載入進度」;無存檔則靜默略過,回一般流程)
			g.titleIdleFrames = 0
			g.showTitle = false
		case in.DirEdge == 0 || in.DirEdge == 1:
			nf.cursor ^= 1
		}
	case ngName:
		tapIdx := -1
		if in.Tapped {
			tapIdx = nf.ni.hits.at(in.TapX, in.TapY)
		}
		confirmed, canceled := nf.ni.input(in, tapIdx)
		switch {
		case canceled:
			nf.stage = ngMenu
		case confirmed && len(nf.ni.nameBuf) == 0:
			// 名字空 → 完成無效(docs/15「[0x270a]>=1 才放行」),留在命名畫面重輸
		case confirmed:
			nf.gs.Init()
			nf.stage = ngGender
		}
	case ngGender:
		tapIdx := -1
		if in.Tapped {
			tapIdx = nf.gs.hits.at(in.TapX, in.TapY)
		}
		gender, confirmed := nf.gs.input(in, tapIdx)
		switch {
		case in.Cancel: // 回命名(可修改),游標停在 OK 格(對齊酒館既有 Cancel 語意)
			nf.stage, nf.ni.cursor = ngName, niCellOK
		case confirmed:
			g.heroName = append([]int(nil), nf.ni.nameBuf...)
			g.dlg.heroName = g.heroName // 對話插值 VAR_NAME 同步(創角後 heroName 重新 assign,須重同步)
			g.heroGender = gender
			g.rollHeroLevelOne()
			nf.preview = g.heroStat
			nf.previewDef = int(g.heroStat[stats.VIT])
			nf.previewGender = gender
			if g.shop.items != nil {
				nf.previewDef += g.shop.items.Defense(g.equip[1])
			}
			nf.stage, nf.confirmCursor = ngConfirm, 0
		}
	case ngConfirm:
		if in.DirEdge == 0 || in.DirEdge == 1 {
			nf.confirmCursor ^= 1
		}
		if in.Cancel || (in.Confirm || in.Enter) && nf.confirmCursor == 1 {
			// DQ3.EXE 1c76..1c84：Cancel 或選否都跳回 1bca，姓名/性別重新輸入。
			nf.ni.Init()
			nf.stage = ngName
			return
		}
		if (in.Confirm || in.Enter) && nf.confirmCursor == 0 {
			g.titleIdleFrames = 0
			g.showTitle = false
			g.startOpening()
		}
	}
}

// draw:標題子畫面(ngSplash 無額外疊繪,底圖由 renderFrame 畫;其餘三階段疊 modal 框)。
func (nf *NewGameFlow) draw(rgba []byte, tx *dq3data.Text, white, yellow dq3data.Color) {
	switch nf.stage {
	case ngMenu:
		bx, by, bw := 190, 140, 200
		fillBox(rgba, bx, by, bw, 90, white)
		for i, g := range ngTitleLabel {
			drawGlyph(rgba, tx, bx+30+i*16, by+10, g, white)
		}
		nf.hits.reset()
		for i := 0; i < ngOptCount; i++ {
			y := by + 40 + i*22
			x := bx + 60
			if i == nf.cursor {
				drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
			}
			for j, g := range ngOptLabels[i] {
				drawGlyph(rgba, tx, x+j*16, y, g, white)
			}
			nf.hits.add(x-18, y-3, 120, 18, i)
		}
	case ngName:
		fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
		nf.ni.draw(rgba, tx, white, yellow, 80, 56)
	case ngGender:
		fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
		for i, g := range nf.ni.nameBuf { // 顯示已輸入主角名
			drawGlyph(rgba, tx, 80+i*16, 56, g, white)
		}
		nf.gs.draw(rgba, tx, white, yellow, 80, 100, 22)
	case ngConfirm:
		drawNewGameStats(rgba, tx, nf, white, yellow)
	}
}

var (
	ngLabLevel = []int{419, 420}
	ngLabHP    = []int{22, 30}
	ngLabMP    = []int{27, 30}
	ngLabAGI   = []int{282, 283}
	ngLabATK   = []int{623, 624}
	ngLabDEF   = []int{340, 409}
	ngLabEXP   = []int{525, 526}
	ngLabSex   = []int{674, 417}
	ngHero     = []int{106, 187}
	ngCloth    = []int{190, 149, 191, 192}
	ngPrompt   = []int{398, 546, 194, 229, 456, 534} // 這個人可以嗎
	ngYesNo    = [2][]int{{399}, {678}}              // 是 / 否
)

func drawNGRow(rgba []byte, tx *dq3data.Text, x, y int, lab []int, val int, c dq3data.Color) {
	for i, gl := range lab {
		drawGlyph(rgba, tx, x+i*dq3data.GlyphPx, y, gl, c)
	}
	drawNumber(rgba, tx, x+(len(lab)+1)*dq3data.GlyphPx, y, val, c)
}

// drawNewGameStats 重現原版 1c54→sub_96be 的能力確認資訊。裝備衍生值依
// sub_9521：攻=力量、守=耐力+布衣防。
func drawNewGameStats(rgba []byte, tx *dq3data.Text, nf *NewGameFlow, white, yellow dq3data.Color) {
	const rowH = 16
	// 原版 v3_01_afterstats.png 的四個 panel 幾何（640×350 native）。
	fillBox(rgba, 128, 42, 116, 76, white)
	fillBox(rgba, 128, 116, 116, 67, white)
	fillBox(rgba, 243, 42, 181, 141, white)
	fillBox(rgba, 294, 16, 116, 29, white)
	fillBox(rgba, 294, 48, 79, 45, white)

	for i, gl := range nf.ni.nameBuf {
		drawGlyph(rgba, tx, 170+i*dq3data.GlyphPx, 50, gl, white)
	}
	for i, gl := range ngHero {
		drawGlyph(rgba, tx, 170+i*dq3data.GlyphPx, 66, gl, white)
	}
	for i, gl := range ngLabSex {
		drawGlyph(rgba, tx, 145+i*dq3data.GlyphPx, 82, gl, white)
	}
	sexGlyph := 775
	if nf.previewGender == 1 {
		sexGlyph = 234
	}
	drawGlyph(rgba, tx, 196, 82, sexGlyph, white)
	drawNGRow(rgba, tx, 145, 98, ngLabLevel, 1, white)
	drawNGRow(rgba, tx, 145, 114, ngLabHP, int(nf.preview[stats.HP]), white)
	drawNGRow(rgba, tx, 145, 130, ngLabMP, int(nf.preview[stats.MP]), white)
	for i, gl := range ngCloth {
		drawGlyph(rgba, tx, 158+i*dq3data.GlyphPx, 142, gl, white)
	}

	x := 274
	drawNGRow(rgba, tx, x, 92, ngLabAGI, int(nf.preview[stats.AGI]), white)
	drawNGRow(rgba, tx, x, 92+rowH, ngLabHP, int(nf.preview[stats.HP]), white)
	drawNGRow(rgba, tx, x, 92+2*rowH, ngLabMP, int(nf.preview[stats.MP]), white)
	drawNGRow(rgba, tx, x, 92+3*rowH, ngLabATK, int(nf.preview[stats.STR]), white)
	drawNGRow(rgba, tx, x, 92+4*rowH, ngLabDEF, nf.previewDef, white)
	drawNGRow(rgba, tx, x, 92+5*rowH, ngLabEXP, 0, white)
	for i, gl := range ngPrompt {
		drawGlyph(rgba, tx, 304+i*dq3data.GlyphPx, 22, gl, white)
	}
	for i := range ngYesNo {
		y := 54 + i*16
		if i == nf.confirmCursor {
			drawGlyph(rgba, tx, 304, y, curGlyph, yellow)
		}
		for j, gl := range ngYesNo[i] {
			drawGlyph(rgba, tx, 332+j*dq3data.GlyphPx, y, gl, white)
		}
	}
}
