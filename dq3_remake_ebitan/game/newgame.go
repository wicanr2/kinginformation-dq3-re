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

// drawNewGameWindowBackdrops replays one pack-selected draw layer of legacy
// EGA window backdrops. Raw X/Width remain byte-addressed in the pack; the
// shared renderer performs the registered projection and knows no DQ3
// coordinates or window identities.
func drawNewGameWindowBackdrops(rgba []byte, geo *gamepack.NewGameGeometry, drawOrder int) {
	if geo == nil || len(geo.WindowBackdrops) == 0 {
		return
	}
	rawByID := make(map[string]gamepack.RawNewGameWindow, len(geo.RawWindows))
	for _, raw := range geo.RawWindows {
		rawByID[raw.ID] = raw
	}
	black := dq3data.Color{}
	for _, backdrop := range geo.WindowBackdrops {
		order := 0
		if backdrop.DrawOrder != nil {
			order = *backdrop.DrawOrder
		}
		if order != drawOrder {
			continue
		}
		raw, ok := rawByID[backdrop.RawWindowID]
		if !ok {
			// Production pack validation rejects this path. Direct fixtures may
			// intentionally be partial, so preserve the already drawn background.
			continue
		}
		switch backdrop.Primitive {
		case "ega_window_black_backdrop":
			// The pack makes any terminal EGA cell explicit: raw coordinates stay
			// byte-addressed, while the generic primitive applies only the
			// validated visible projection. Partial direct fixtures retain their
			// raw rectangle; production validation rejects omitted projections.
			right, bottom := 0, 0
			if backdrop.TerminalRightPixels != nil {
				right = *backdrop.TerminalRightPixels
			}
			if backdrop.TerminalBottomPixels != nil {
				bottom = *backdrop.TerminalBottomPixels
			}
			x, width := raw.X*8, raw.Width*8+right
			for y := raw.Y; y < raw.Y+raw.Height+bottom; y++ {
				for px := x; px < x+width; px++ {
					putPx(rgba, px, y, black)
				}
			}
		}
	}
}

func drawNGRow(rgba []byte, tx *dq3data.Text, field gamepack.NewGameStatField, lab []int, val int, c dq3data.Color) {
	for i, gl := range lab {
		drawGlyph(rgba, tx, field.Label.X+i*dq3data.GlyphPx, field.Label.Y, gl, c)
	}
	drawNGNumber(rgba, tx, field.Value, val, c)
}

// drawNGNumber 重現 sub_21929/sub_21822 的固定寬度 decimal field：原版先走
// 高位、跳過 leading zero，仍逐格前進，最後一位固定落在欄位右端。
func drawNGNumber(rgba []byte, tx *dq3data.Text, field gamepack.NumberField, val int, c dq3data.Color) {
	if field.Digits <= 0 || val < 0 {
		return
	}
	divisor := 1
	for i := 1; i < field.Digits; i++ {
		divisor *= 10
	}
	started := false
	for i := 0; i < field.Digits; i++ {
		digit := val / divisor
		val %= divisor
		if digit != 0 || started || i == field.Digits-1 {
			drawGlyph(rgba, tx, field.X+i*dq3data.GlyphPx, field.Y, digit, c)
			started = true
		}
		if divisor > 1 {
			divisor /= 10
		}
	}
}

// drawNewGameStats 重現原版 1c54→sub_96be 的能力確認資訊。裝備衍生值依
// sub_9521：攻=力量、守=耐力+布衣防。
func drawNewGameStats(rgba []byte, tx *dq3data.Text, nf *NewGameFlow, white, yellow dq3data.Color, geo *gamepack.NewGameGeometry) {
	if nf.labels == nil {
		return
	}
	if geo.Stats == nil {
		return
	}
	// 原版能力 panel 的分割幾何與 confirm_choice 較大背景矩形都由 pack
	// 提供。raw EGA backdrop 已先清為黑色；這三個可見框再依原始量測覆蓋。
	fillGeometryFramedRect(rgba, geo.Frame, geo.StatsLeft, white)
	fillGeometryFramedRect(rgba, geo.Frame, geo.StatsEquipment, white)
	fillGeometryFramedRect(rgba, geo.Frame, geo.StatsRight, white)

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
	sexGlyphs := nf.labels.Male
	if nf.previewGender == 1 {
		sexGlyphs = nf.labels.Female
	}
	for i, gl := range sexGlyphs {
		drawGlyph(rgba, tx, geo.StatsSexValue.X+i*dq3data.GlyphPx, geo.StatsSexValue.Y, gl, textColor)
	}
	statLayout := geo.Stats
	drawNGRow(rgba, tx, statLayout.Level, nf.labels.Level, 1, textColor)
	drawNGRow(rgba, tx, statLayout.HP, nf.labels.HP, int(nf.preview[stats.HP]), textColor)
	drawNGRow(rgba, tx, statLayout.MP, nf.labels.MP, int(nf.preview[stats.MP]), textColor)
	for i, gl := range nf.labels.Cloth {
		drawGlyph(rgba, tx, geo.StatsCloth.X+i*dq3data.GlyphPx, geo.StatsCloth.Y, gl, textColor)
	}

	// IDA linear 0x10854→0x1834e 會由同一角色 record 依序寫入十列：
	// 四項基本能力，接著是運氣／最大 HP／最大 MP／攻防／經驗。每列的 leading
	// blank、中文標籤及冒號均在 pack label glyph stream，不能由字長推導。
	drawNGRow(rgba, tx, statLayout.Strength, nf.labels.Strength, int(nf.preview[stats.STR]), textColor)
	drawNGRow(rgba, tx, statLayout.Agility, nf.labels.Agility, int(nf.preview[stats.AGI]), textColor)
	drawNGRow(rgba, tx, statLayout.Vitality, nf.labels.Vitality, int(nf.preview[stats.VIT]), textColor)
	drawNGRow(rgba, tx, statLayout.Intelligence, nf.labels.Intelligence, int(nf.preview[stats.INT]), textColor)
	drawNGRow(rgba, tx, statLayout.Luck, nf.labels.Luck, int(nf.preview[stats.LUCK]), textColor)
	drawNGRow(rgba, tx, statLayout.MaxHP, nf.labels.MaxHP, int(nf.preview[stats.HP]), textColor)
	drawNGRow(rgba, tx, statLayout.MaxMP, nf.labels.MaxMP, int(nf.preview[stats.MP]), textColor)
	drawNGRow(rgba, tx, statLayout.Attack, nf.labels.Attack, int(nf.preview[stats.STR]), textColor)
	drawNGRow(rgba, tx, statLayout.Defense, nf.labels.Defense, nf.previewDef, textColor)
	drawNGRow(rgba, tx, statLayout.Experience, nf.labels.Experience, 0, textColor)

	// sub_10854 在 record407 的數值／標籤後才畫確認 prompt 與選項 window；
	// 因此 choice backdrop 必須蓋住右欄前四列的一部分，不能讓 stat glyph
	// 反向浮到藍黑底圖上。這個順序由同一正式輸入原版畫面與 IDA caller
	// `sub_1834e → sub_1f590` 交叉確認。
	drawNewGameWindowBackdrops(rgba, geo, 1)
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
	for i, gl := range nf.labels.Prompt {
		drawGlyph(rgba, tx, geo.StatsPrompt.X+i*dq3data.GlyphPx, geo.StatsPrompt.Y, gl, textColor)
	}
	for i, label := range [][]int{nf.labels.Yes, nf.labels.No} {
		y := geo.StatsChoice.Y + i*geo.StatsChoice.StepY
		if i == nf.confirmCursor {
			drawGlyph(rgba, tx, geo.StatsChoiceCursor.X, y, nf.labels.ChoiceCursor[0], cursorColor)
		}
		for j, gl := range label {
			drawGlyph(rgba, tx, geo.StatsChoice.X+j*dq3data.GlyphPx, y, gl, textColor)
		}
	}
}
