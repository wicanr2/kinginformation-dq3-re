package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/zhuyin"
)

// 共用「命名輸入」+「性別選單」widget(移植 dq3_nameinput,docs/15):英數格盤(0-9/A-Z + ←/OK)+
// 注音組字(zhuyin.Composer,42 格盤 + 同音候選)、Tab/情境鍵切模式。酒館招募(tavern.go)與
// 主角創建(newgame.go)共用同一份實作,不重複兩份邏輯——差異只在「OK 時是否強制非空才放行」,
// 由呼叫端依情境判斷(酒館允許空名回退職業名;主角創建必填,見 docs/15 §「完成放行條件」)。
const (
	niChars   = 36 // 0-9 + A-Z
	niCols    = 9
	niRows    = 5
	niCells   = niCols * niRows // 原版 rec452/453 的 45 格(含功能／離開符號)
	niCellFn  = 43              // 注音／英數格的 ⊙：進右側功能列
	niCellOK  = 44              // 相容舊測試名稱；原版右下角為完成／離開
	niNameMax = 4
)

// NameInput 的游標以畫面 raw index 表示：row*9+col。原版 handler
// 再把它換成欄優先的語意 cell=col*5+row，注音組字表才是聲母／韻母／聲調順序。
// 這兩個換算不能交給「看起來像 9 欄」的 renderer 自行猜測。
func niRawToCell(raw int) int {
	if raw < 0 || raw >= niCells {
		return -1
	}
	return (raw%niCols)*niRows + raw/niCols
}

func niCellToRaw(cell int) int {
	if cell < 0 || cell >= niCells {
		return -1
	}
	return (cell%niRows)*niCols + cell/niRows
}

func niMoveRaw(raw, dir int) int {
	if raw < 0 || raw >= niCells {
		raw = 0
	}
	row, col := raw/niCols, raw%niCols
	switch dir {
	case 0: // 下
		row = (row + 1) % niRows
	case 1: // 上
		row = (row + niRows - 1) % niRows
	case 2: // 左
		col = (col + niCols - 1) % niCols
	case 3: // 右
		col = (col + 1) % niCols
	}
	return row*niCols + col
}

const niTapFunctionBase = -100

func niTapFunction(idx int) int { return niTapFunctionBase - idx }

func niFunctionFromTap(tap int) (int, bool) {
	if tap > niTapFunctionBase || tap <= niTapFunctionBase-5 {
		return 0, false
	}
	return niTapFunctionBase - tap, true
}

// niCellGlyph 是測試／無 pack fixture 的相容對映；production renderer 會優先
// 取 game-pack rec453 的 name_alnum_grid，不把版本專屬 glyph 當成引擎 fallback。
func niCellGlyph(cell int) int {
	if cell < 0 || cell >= niCells {
		return -1
	}
	if cell < 10 {
		return cell
	}
	if cell <= 35 {
		return 15 + (cell - 10)
	}
	if cell <= 42 {
		return 0x71 + (cell - 35)
	}
	return -1
}

// NameInput 是命名輸入 widget 狀態(英數格盤 或 注音組字,Tab 互切)。
type NameInput struct {
	cursor         int // raw grid index(0..44);候選窗時與 zh.Cursor 同步
	nameBuf        []int
	namePos        int  // 姓名列游標(0..len(nameBuf))，功能列前進／後退使用
	nameZhu        bool // true=注音模式(Tab/CtxTap 切換)
	functionFocus  bool // true=焦點在右側五列功能列
	functionCursor int
	labels         *gamepack.NewGameLabels
	zh             zhuyin.Composer
	hits           hitList // 目前畫面(格盤/候選/功能列)的可點區塊,draw() 重建
}

func (ni *NameInput) setLabels(labels *gamepack.NewGameLabels) { ni.labels = labels }

// Init 重設為新一輪命名輸入(英數模式、空緩衝、游標歸零)。
// Init 重設輸入狀態。預設**注音模式**(對齊原版:docs/36 04 截圖新遊戲命名即注音盤,游標 raw0=ㄅ;
// Tab/情境鍵切英數)。
func (ni *NameInput) Init() {
	labels := ni.labels
	*ni = NameInput{nameZhu: true, labels: labels, functionCursor: 0}
	ni.zh.Init()
}

// input:一幀輸入 → (confirmed, canceled)。
//   - confirmed=true:原版完成格／功能列第五列被選定；是否放行由呼叫端依情境判斷名字是否為空
//     （主角必填，酒館可回退職業名；見 docs/15「完成放行條件」）。
//   - canceled=true:英數格盤按 Cancel(注音模式按 Cancel 只退回英數,不算 canceled,對齊
//     docs/15「cell43 只在注音模式成立」的模式分工)。
//
// tapIdx:觸控命中格 index(<0=無命中);呼叫端需依 ni.nameZhu 決定要對哪個 hits 算 tapIdx
//   - nameZhu 時對 ni.hits(注音盤/候選共用同一份,draw() 每次重建),否則同樣對 ni.hits(英數格盤)。
func (ni *NameInput) input(in InputState, tapIdx int) (confirmed, canceled bool) {
	confirm := in.Confirm || in.Enter
	if tapIdx != -1 {
		if fn, ok := niFunctionFromTap(tapIdx); ok {
			ni.functionFocus, ni.functionCursor = true, fn
			return ni.activateFunction(fn)
		}
		if ni.functionFocus {
			// 觸控格盤與功能列互不重疊；保留此分支只為在舊 fixture
			// 傳入 stale tap 時不把功能焦點誤改回格盤。
			return false, false
		}
		ni.cursor, confirm = tapIdx, true
	}
	if (in.Toggle || in.CtxTap) && !ni.functionFocus {
		ni.toggleMode()
		return false, false
	}
	if ni.functionFocus {
		if in.Cancel {
			ni.functionFocus = false
			return false, true
		}
		if in.DirEdge == 0 || in.DirEdge == 1 {
			if in.DirEdge == 0 {
				ni.functionCursor = (ni.functionCursor + 1) % 5
			} else {
				ni.functionCursor = (ni.functionCursor + 4) % 5
			}
		}
		if confirm {
			return ni.activateFunction(ni.functionCursor)
		}
		return false, false
	}
	if ni.nameZhu { // 注音組字(盤格或候選格,視 ni.zh.Pick 而定)
		if ni.zh.Pick {
			if tapIdx >= 0 {
				ni.zh.Cursor = tapIdx
			}
			switch {
			case in.Cancel:
				ni.zh.Cancel()
			case confirm:
				if g, picked := ni.zh.Confirm(); picked {
					ni.appendGlyph(g)
				}
			case in.DirEdge >= 0:
				ni.zh.Move(in.DirEdge)
			}
			ni.cursor = ni.zh.Cursor
			return false, false
		}
		if ni.cursor < 0 || ni.cursor >= niCells {
			ni.cursor = 0
		}
		ni.zh.Cursor = niRawToCell(ni.cursor)
		if in.Cancel {
			if !ni.zh.Cancel() { // 非候選階段 → 保留相容的英數退回
				ni.nameZhu, ni.cursor = false, 0
			}
			return false, false
		}
		if confirm {
			cell := niRawToCell(ni.cursor)
			switch cell {
			case niCellFn:
				ni.functionFocus, ni.functionCursor = true, 0
			case niCellOK:
				return true, false
			case 42: // ←：清除目前注音組字，原版重畫 rec456「輸入注音」框
				ni.zh.Init()
			default:
				if g, picked := ni.zh.Confirm(); picked {
					ni.appendGlyph(g)
				}
			}
		}
		if in.DirEdge >= 0 && !confirm {
			ni.cursor = niMoveRaw(ni.cursor, in.DirEdge)
		}
		return false, false
	}
	// 英數格盤：rec453 的 45 格中，0..42 是字元，43 進功能列，44 完成。
	if in.Cancel {
		return false, true
	}
	if confirm {
		cell := niRawToCell(ni.cursor)
		switch {
		case cell == niCellFn:
			ni.functionFocus, ni.functionCursor = true, 0
		case cell == niCellOK:
			return true, false
		default:
			if ni.labels != nil && ni.cursor >= 0 && ni.cursor < len(ni.labels.NameAlnumGrid) {
				ni.appendGlyph(ni.labels.NameAlnumGrid[ni.cursor])
			} else if g := niCellGlyph(cell); g >= 0 {
				ni.appendGlyph(g)
			}
		}
	} else if in.DirEdge >= 0 {
		ni.cursor = niMoveRaw(ni.cursor, in.DirEdge)
	}
	return false, false
}

func (ni *NameInput) toggleMode() {
	ni.nameZhu = !ni.nameZhu
	ni.functionFocus = false
	ni.cursor = 0
	if ni.nameZhu {
		ni.zh.Init()
	}
}

func (ni *NameInput) appendGlyph(g int) {
	if g < 0 || len(ni.nameBuf) >= niNameMax {
		return
	}
	if ni.namePos < 0 || ni.namePos > len(ni.nameBuf) {
		ni.namePos = len(ni.nameBuf)
	}
	if ni.namePos == len(ni.nameBuf) {
		ni.nameBuf = append(ni.nameBuf, g)
	} else {
		ni.nameBuf[ni.namePos] = g
	}
	if ni.namePos < niNameMax {
		ni.namePos++
	}
}

func (ni *NameInput) activateFunction(fn int) (confirmed, canceled bool) {
	ni.functionFocus = false
	switch fn {
	case 0: // 英數／注音
		ni.toggleMode()
	case 1: // 前進
		if ni.namePos < len(ni.nameBuf) {
			ni.namePos++
		}
	case 2: // 後退：把姓名游標往前，後續輸入覆寫該字
		if ni.namePos > 0 {
			ni.namePos--
		}
	case 3: // 取消
		return false, true
	case 4: // 完成；空名是否放行由主角／酒館呼叫端決定
		return true, false
	}
	return false, false
}

// draw:已輸入名、注音／英數格盤與 pack 擁有的像素 anchor。raw grid 的
// 9 欄／5 列與 16px 步距來自 sub_11172/sub_1123c，不能在 renderer
// 內另造一組 DQ3 預設座標。
func (ni *NameInput) draw(rgba []byte, tx *dq3data.Text, white, yellow dq3data.Color, geo gamepack.NewGameGeometry) {
	nameStep := geo.NameText.StepX
	if nameStep <= 0 {
		nameStep = dq3data.GlyphPx
	}
	grid := geo.NameGrid
	cur := func(x, y int, sel bool) {
		if sel {
			drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
		}
	}
	drawStream := func(anchor gamepack.GeometryAnchor, stream []int, color dq3data.Color) {
		step := anchor.StepX
		if step <= 0 {
			step = dq3data.GlyphPx
		}
		for i, g := range stream {
			drawGlyph(rgba, tx, anchor.X+i*step, anchor.Y, g, color)
		}
	}
	if ni.labels != nil {
		drawStream(geo.NameTitle, ni.labels.NameTitle, white)
		drawStream(geo.NameLeftArrow, ni.labels.NameLeftArrow, white)
		drawStream(geo.NameRightArrow, ni.labels.NameRightArrow, white)
	}
	for i, g := range ni.nameBuf {
		drawGlyph(rgba, tx, geo.NameText.X+i*nameStep, geo.NameText.Y, g, yellow)
	}
	// 姓名列游標的原版是 EGA writer 產生的色塊，不是戰鬥用 glyph；
	// frame pattern／逐幀閃爍尚未閉合前不以共用 curGlyph 假造它。

	// 右側五列功能列是獨立 widget；即使游標仍在格盤，也要建立觸控命中區。
	ni.hits.reset()
	if ni.labels != nil {
		rows := ni.labels.FunctionZhuyin
		if !ni.nameZhu {
			rows = ni.labels.FunctionAlnum
		}
		for row, label := range rows {
			x := geo.NameFunction.X
			y := geo.NameFunction.Y + row*geo.NameFunction.StepY
			if row == ni.functionCursor && ni.functionFocus {
				drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
			}
			for j, g := range label {
				drawGlyph(rgba, tx, x+j*dq3data.GlyphPx, y, g, white)
			}
			ni.hits.add(x-18, y-3, geo.NameFunctionPanel.Width-18, geo.NameFunction.StepY+2, niTapFunction(row))
		}
		if ni.nameZhu {
			drawStream(geo.NameMode, ni.labels.InputModeZhuyin, white)
		}
	}

	if ni.nameZhu { // 注音模式：組字列 + (候選或 45 格原版盤)
		cx := geo.NameText.X + 5*nameStep
		if ni.zh.Sh != 0 {
			if g := ni.zhComponentGlyph(ni.zh.Sh - 1); g >= 0 {
				drawGlyph(rgba, tx, cx, geo.NameText.Y, g, yellow)
			}
			cx += nameStep
		}
		if ni.zh.Ji != 0 {
			if g := ni.zhComponentGlyph(34 + ni.zh.Ji - 1); g >= 0 {
				drawGlyph(rgba, tx, cx, geo.NameText.Y, g, yellow)
			}
			cx += nameStep
		}
		if ni.zh.Yu != 0 {
			if g := ni.zhComponentGlyph(21 + ni.zh.Yu - 1); g >= 0 {
				drawGlyph(rgba, tx, cx, geo.NameText.Y, g, yellow)
			}
		}
		if ni.zh.Pick { // 候選窗(9 欄)
			ni.cursor = ni.zh.Cursor
			for i, g := range ni.zh.Cand {
				x, y := grid.X+(i%grid.Columns)*grid.StepX, grid.Y+(i/grid.Columns)*grid.StepY
				cur(x, y, i == ni.zh.Cursor)
				drawGlyph(rgba, tx, x, y, g, white)
				ni.hits.add(x-4, y-3, grid.StepX+2, grid.StepY+2, i)
			}
		} else {
			for raw := 0; raw < niCells; raw++ {
				col, row := raw%grid.Columns, raw/grid.Columns
				x, y := grid.X+col*grid.StepX, grid.Y+row*grid.StepY
				cur(x, y, raw == ni.cursor)
				if ni.labels != nil && raw < len(ni.labels.NameZhuyinGrid) {
					drawGlyph(rgba, tx, x, y, ni.labels.NameZhuyinGrid[raw], white)
				}
				ni.hits.add(x-4, y-3, grid.StepX+2, grid.StepY+2, raw)
			}
		}
		return
	}
	// 英數模式同樣是 rec453 的 45 格；特殊符號不是退格 fallback，而是原版姓名字元。
	for raw := 0; raw < niCells; raw++ {
		col, row := raw%grid.Columns, raw/grid.Columns
		x, y := grid.X+col*grid.StepX, grid.Y+row*grid.StepY
		cur(x, y, raw == ni.cursor)
		if ni.labels != nil && raw < len(ni.labels.NameAlnumGrid) {
			drawGlyph(rgba, tx, x, y, ni.labels.NameAlnumGrid[raw], white)
		} else if g := niCellGlyph(niRawToCell(raw)); g >= 0 {
			drawGlyph(rgba, tx, x, y, g, white)
		}
		ni.hits.add(x-4, y-3, grid.StepX+2, grid.StepY+2, raw)
	}
}

func (ni *NameInput) zhComponentGlyph(cell int) int {
	if raw := niCellToRaw(cell); ni.labels != nil && raw >= 0 && raw < len(ni.labels.NameZhuyinGrid) {
		return ni.labels.NameZhuyinGrid[raw]
	}
	return zhuyin.CellGlyph(cell)
}

// GenderSelect 是共用性別選單 widget(▶男性/女性 兩列)。移植 rec534/535(docs/36)。
type GenderSelect struct {
	cursor int
	hits   hitList
	labels *gamepack.NewGameLabels
}

func (gs *GenderSelect) setLabels(labels *gamepack.NewGameLabels) { gs.labels = labels }

// Init 重設游標(新一輪性別選擇固定從男性開始)。
func (gs *GenderSelect) Init() { gs.cursor = 0 }

// input:方向鍵切換(0/1 兩列環繞)、Confirm/點列 選定 → (gender, confirmed)。
func (gs *GenderSelect) input(in InputState, tapIdx int) (gender int, confirmed bool) {
	confirm := in.Confirm
	if tapIdx >= 0 {
		gs.cursor, confirm = tapIdx, true
	}
	switch {
	case confirm:
		return gs.cursor, true
	case in.DirEdge == 0 || in.DirEdge == 1:
		gs.cursor ^= 1
	}
	return 0, false
}

// draw:兩列「男性/女性」(x,y0 為第一列起點,dy 為列距),游標 ► + hits 對齊 input() 的 tapIdx。
func (gs *GenderSelect) draw(rgba []byte, tx *dq3data.Text, white, yellow dq3data.Color, geo gamepack.NewGameGeometry) {
	gs.hits.reset()
	for i := 0; i < 2; i++ {
		x, y := geo.Gender.X, geo.Gender.Y+i*geo.Gender.StepY
		if i == gs.cursor {
			drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
		}
		var label []int
		if gs.labels != nil {
			if i == 0 {
				label = gs.labels.Male
			} else {
				label = gs.labels.Female
			}
		}
		for j, g := range label {
			drawGlyph(rgba, tx, x+j*16, y, g, white)
		}
		gs.hits.add(x-18, y-3, geo.GenderPanel.Width-18, geo.Gender.StepY+2, i)
	}
}
