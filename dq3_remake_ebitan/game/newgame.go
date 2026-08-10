package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 標題主選單 + 主角創建(docs/36 §開場流程 + docs/15 命名子系統):
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

// NewGameFlow 是標題主選單 + 主角創建的 modal 狀態機。
type NewGameFlow struct {
	stage         int
	cursor        int // 主選單游標(0/1)
	hits          hitList
	ni            NameInput
	gs            GenderSelect
	labels        *gamepack.NewGameLabels
	geometry      *gamepack.NewGameGeometry
	confirmCursor int // 0=是、1=否
	preview       stats.Values
	previewDef    int
	previewGender int
}

// setLabels 安裝標題／創角流程共用的 game-pack 字模資料。Pack 在 bootstrap
// 後視為唯讀；各 widget 只保留這個受驗證的指標，不複製版本專屬常數。
func (nf *NewGameFlow) setLabels(labels gamepack.NewGameLabels) {
	nf.labels = &labels
	nf.ni.setLabels(nf.labels)
	nf.gs.setLabels(nf.labels)
}

// setGeometry 安裝 pack 擁有的創角像素幾何；production bootstrap 會在
// NewGameWithPack 中 fail closed，直接 UI fixture 則需明確安裝同一份 pack。
func (nf *NewGameFlow) setGeometry(geometry gamepack.NewGameGeometry) {
	nf.geometry = &geometry
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
	if nf.geometry == nil {
		nf.hits.reset()
		return
	}
	geo := nf.geometry
	switch nf.stage {
	case ngMenu:
		if nf.labels == nil {
			nf.hits.reset()
			return
		}
		fillGeometryBox(rgba, geo.Frame, geo.Menu.X, geo.Menu.Y, geo.Menu.Width, geo.Menu.Height, white)
		for i, g := range nf.labels.Title {
			drawGlyph(rgba, tx, geo.MenuTitle.X+i*geo.MenuTitle.StepX, geo.MenuTitle.Y+i*geo.MenuTitle.StepY, g, white)
		}
		nf.hits.reset()
		for i := 0; i < ngOptCount; i++ {
			x := geo.MenuOptions.X
			y := geo.MenuOptions.Y + i*geo.MenuOptions.StepY
			if i == nf.cursor {
				drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
			}
			label := nf.labels.Start
			if i == ngOptLoad {
				label = nf.labels.Load
			}
			for j, g := range label {
				drawGlyph(rgba, tx, x+j*16, y, g, white)
			}
			nf.hits.add(x-18, y-3, geo.Menu.Width-24, 18, i)
		}
	case ngName:
		fillGeometryBox(rgba, geo.Frame, geo.NamePanel.X, geo.NamePanel.Y, geo.NamePanel.Width, geo.NamePanel.Height, white)
		fillGeometryBox(rgba, geo.Frame, geo.NameFunctionPanel.X, geo.NameFunctionPanel.Y, geo.NameFunctionPanel.Width, geo.NameFunctionPanel.Height, white)
		fillGeometryBox(rgba, geo.Frame, geo.NameModePanel.X, geo.NameModePanel.Y, geo.NameModePanel.Width, geo.NameModePanel.Height, white)
		nf.ni.draw(rgba, tx, white, yellow, *geo)
	case ngGender:
		fillGeometryBox(rgba, geo.Frame, geo.GenderPanel.X, geo.GenderPanel.Y, geo.GenderPanel.Width, geo.GenderPanel.Height, white)
		for i, g := range nf.ni.nameBuf { // 顯示已輸入主角名
			drawGlyph(rgba, tx, geo.NameText.X+i*16, geo.NameText.Y, g, white)
		}
		nf.gs.draw(rgba, tx, white, yellow, *geo)
	case ngConfirm:
		drawNewGameStats(rgba, tx, nf, white, yellow, geo)
	}
}

func drawNGRow(rgba []byte, tx *dq3data.Text, x, y int, lab []int, val int, c dq3data.Color) {
	for i, gl := range lab {
		drawGlyph(rgba, tx, x+i*dq3data.GlyphPx, y, gl, c)
	}
	drawNumber(rgba, tx, x+(len(lab)+1)*dq3data.GlyphPx, y, val, c)
}

// drawNewGameStats 重現原版 1c54→sub_96be 的能力確認資訊。裝備衍生值依
// sub_9521：攻=力量、守=耐力+布衣防。
func drawNewGameStats(rgba []byte, tx *dq3data.Text, nf *NewGameFlow, white, yellow dq3data.Color, geo *gamepack.NewGameGeometry) {
	if nf.labels == nil {
		return
	}
	// 原版能力 panel 的幾何與 confirm_choice 較大背景矩形都由 pack 提供。
	// backdrop 必須在文字前寫入，讓未命中的黑色 phase 保留底圖，且不會
	// 把能力欄位的 glyph 誤當成背景的一部分。
	fillGeometryBox(rgba, geo.Frame, geo.StatsLeft.X, geo.StatsLeft.Y, geo.StatsLeft.Width, geo.StatsLeft.Height, white)
	fillGeometryBox(rgba, geo.Frame, geo.StatsEquipment.X, geo.StatsEquipment.Y, geo.StatsEquipment.Width, geo.StatsEquipment.Height, white)
	fillGeometryBox(rgba, geo.Frame, geo.StatsRight.X, geo.StatsRight.Y, geo.StatsRight.Width, geo.StatsRight.Height, white)
	fillGeometryBox(rgba, geo.Frame, geo.ConfirmPrompt.X, geo.ConfirmPrompt.Y, geo.ConfirmPrompt.Width, geo.ConfirmPrompt.Height, white)
	if geo.ConfirmChoiceBackdrop != nil {
		fillPatternRect(rgba, geo.ConfirmChoiceBackdrop)
		fillGeometryRect(rgba, geo.ConfirmChoiceFrame, geo.ConfirmChoiceContent, white)
		strokeGeometryBox(rgba, geo.ConfirmChoiceFrame, geo.ConfirmChoice.X, geo.ConfirmChoice.Y,
			geo.ConfirmChoice.Width, geo.ConfirmChoice.Height, white)
	} else {
		// 只保留給未安裝 production pack 的 direct fixture；正式 validator
		// 會拒絕缺少 confirm_choice_backdrop 的資料。
		fillGeometryBox(rgba, geo.Frame, geo.ConfirmChoice.X, geo.ConfirmChoice.Y, geo.ConfirmChoice.Width, geo.ConfirmChoice.Height, white)
	}

	// 能力確認頁的前景與游標跟 frame 共用原版 lavender palette；其他
	// 新遊戲頁仍由呼叫端 white/yellow 控制，避免把未閉合的全域 palette
	// 結論外推到姓名盤或性別頁。
	textColor, cursorColor := white, yellow
	choiceFrame := geo.ConfirmChoiceFrame
	if choiceFrame == nil {
		choiceFrame = geo.Frame
	}
	if choiceFrame != nil {
		textColor = frameColor(choiceFrame.BorderRGB)
		cursorColor = textColor
	}

	for i, gl := range nf.ni.nameBuf {
		drawGlyph(rgba, tx, geo.StatsName.X+i*dq3data.GlyphPx, geo.StatsName.Y, gl, textColor)
	}
	for i, gl := range nf.labels.Hero {
		drawGlyph(rgba, tx, geo.StatsHero.X+i*dq3data.GlyphPx, geo.StatsHero.Y, gl, textColor)
	}
	for i, gl := range nf.labels.Sex {
		drawGlyph(rgba, tx, geo.StatsSex.X+i*dq3data.GlyphPx, geo.StatsSex.Y, gl, textColor)
	}
	sexGlyph := nf.labels.Male[0]
	if nf.previewGender == 1 {
		sexGlyph = nf.labels.Female[0]
	}
	drawGlyph(rgba, tx, geo.StatsSexValue.X, geo.StatsSexValue.Y, sexGlyph, textColor)
	drawNGRow(rgba, tx, geo.StatsLeftRows.X, geo.StatsLeftRows.Y, nf.labels.Level, 1, textColor)
	drawNGRow(rgba, tx, geo.StatsLeftRows.X, geo.StatsLeftRows.Y+geo.StatsLeftRows.StepY, nf.labels.HP, int(nf.preview[stats.HP]), textColor)
	drawNGRow(rgba, tx, geo.StatsLeftRows.X, geo.StatsLeftRows.Y+2*geo.StatsLeftRows.StepY, nf.labels.MP, int(nf.preview[stats.MP]), textColor)
	for i, gl := range nf.labels.Cloth {
		drawGlyph(rgba, tx, geo.StatsCloth.X+i*dq3data.GlyphPx, geo.StatsCloth.Y, gl, textColor)
	}

	x := geo.StatsRightRows.X
	// 原版能力確認右欄不是「速度／HP／MP」：record 407 的這個面板
	// 依序顯示「運氣點數／最大HP／最大MP／攻擊力／守備力／經驗」。
	// `agility` 仍保留給詳細狀況窗的 pack 對映，但不能拿來填此六列。
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y, nf.labels.Luck, int(nf.preview[stats.LUCK]), textColor)
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y+geo.StatsRightRows.StepY, nf.labels.MaxHP, int(nf.preview[stats.HP]), textColor)
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y+2*geo.StatsRightRows.StepY, nf.labels.MaxMP, int(nf.preview[stats.MP]), textColor)
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y+3*geo.StatsRightRows.StepY, nf.labels.Attack, int(nf.preview[stats.STR]), textColor)
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y+4*geo.StatsRightRows.StepY, nf.labels.Defense, nf.previewDef, textColor)
	drawNGRow(rgba, tx, x, geo.StatsRightRows.Y+5*geo.StatsRightRows.StepY, nf.labels.Experience, 0, textColor)
	for i, gl := range nf.labels.Prompt {
		drawGlyph(rgba, tx, geo.StatsPrompt.X+i*dq3data.GlyphPx, geo.StatsPrompt.Y, gl, textColor)
	}
	for i, label := range [][]int{nf.labels.Yes, nf.labels.No} {
		y := geo.StatsChoice.Y + i*geo.StatsChoice.StepY
		if i == nf.confirmCursor {
			drawGlyph(rgba, tx, geo.StatsChoiceCursor.X, y, curGlyph, cursorColor)
		}
		for j, gl := range label {
			drawGlyph(rgba, tx, geo.StatsChoice.X+j*dq3data.GlyphPx, y, gl, textColor)
		}
	}
}
