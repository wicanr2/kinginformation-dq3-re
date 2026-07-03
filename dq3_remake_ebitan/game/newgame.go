package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 標題主選單 + 主角創建(FLOW-GAP A2/A3/A4,docs/36 §開場流程 + docs/15 命名子系統):
// 標題(壓 A/Enter)→ 主選單「▶遊戲開始 / 載入進度」→ 遊戲開始 → 主角注音/英數命名(必填,
// 完成放行條件=非空,對齊 docs/15「[0x270a]>=1 才放行」)→ 性別 → 開始新遊戲。
// 命名/性別 widget 與酒館招募共用 NameInput/GenderSelect(nameinput.go),不重複兩份邏輯;
// 差異只在完成放行 gate:主角創建強制非空(下方 ngName case),酒館允許空名回退職業名。
// 出生點暫沿用現行地表中心(家室內場景待 C-2)、紅披風勇者立繪暫略(TODO C-7)。
const (
	ngSplash = iota // 標題背景(壓任意鍵開主選單)
	ngMenu          // 主選單:遊戲開始 / 載入進度
	ngName          // 主角命名
	ngGender        // 主角性別
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
	stage  int
	cursor int // 主選單游標(0/1)
	hits   hitList
	ni     NameInput
	gs     GenderSelect
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
			g.heroGender = gender
			g.showTitle = false // 收尾:交還一般 Update()/renderFrame() 流程
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
	}
}
