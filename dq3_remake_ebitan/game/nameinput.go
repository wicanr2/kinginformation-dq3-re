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
	niCells   = niChars + 2 // + ← + OK
	niCellBS  = niChars     // 退格
	niCellOK  = niChars + 1 // 完成
	niNameMax = 4
)

// niCellGlyph:格 index → glyph(數字 0-9→0-9、字母 A-Z→15..40);特殊格(←/OK)回 -1。移植 dq3_nameinput_cell_glyph。
func niCellGlyph(cell int) int {
	if cell < 10 {
		return cell
	}
	if cell < niChars {
		return 15 + (cell - 10)
	}
	return -1
}

// NameInput 是命名輸入 widget 狀態(英數格盤 或 注音組字,Tab 互切)。
type NameInput struct {
	cursor  int
	nameBuf []int
	nameZhu bool // true=注音模式(Tab/CtxTap 切換)
	labels  *gamepack.NewGameLabels
	zh      zhuyin.Composer
	hits    hitList // 目前畫面(英數格盤/注音盤/候選)的可點區塊,draw() 重建
}

func (ni *NameInput) setLabels(labels *gamepack.NewGameLabels) { ni.labels = labels }

// Init 重設為新一輪命名輸入(英數模式、空緩衝、游標歸零)。
// Init 重設輸入狀態。預設**注音模式**(對齊原版:docs/36 04 截圖新遊戲命名即注音盤,游標 raw0=ㄅ;
// Tab/情境鍵切英數)。
func (ni *NameInput) Init() {
	labels := ni.labels
	*ni = NameInput{nameZhu: true, labels: labels}
	ni.zh.Init()
}

// input:一幀輸入 → (confirmed, canceled)。
//   - confirmed=true:游標在 OK 格被選定(是否放行由呼叫端依 nameBuf 是否非空判斷,
//     見 docs/15「完成放行條件:選完成 且 名字非空 才離開」)。
//   - canceled=true:英數格盤按 Cancel(注音模式按 Cancel 只退回英數,不算 canceled,對齊
//     docs/15「cell43 只在注音模式成立」的模式分工)。
//
// tapIdx:觸控命中格 index(<0=無命中);呼叫端需依 ni.nameZhu 決定要對哪個 hits 算 tapIdx
//   - nameZhu 時對 ni.hits(注音盤/候選共用同一份,draw() 每次重建),否則同樣對 ni.hits(英數格盤)。
func (ni *NameInput) input(in InputState, tapIdx int) (confirmed, canceled bool) {
	if in.Toggle || in.CtxTap { // Tab/情境鍵:英數 ↔ 注音
		ni.nameZhu = !ni.nameZhu
		if ni.nameZhu {
			ni.zh.Init()
		} else {
			ni.cursor = 0
		}
		return false, false
	}
	if ni.nameZhu { // 注音組字(盤格 或 候選格,視 ni.zh.Pick 而定)
		confirm := in.Confirm
		if tapIdx >= 0 {
			ni.zh.Cursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			if !ni.zh.Cancel() { // 非候選階段 → 退英數
				ni.nameZhu, ni.cursor = false, 0
			}
		case confirm:
			if g, picked := ni.zh.Confirm(); picked && len(ni.nameBuf) < niNameMax {
				ni.nameBuf = append(ni.nameBuf, g)
			}
		case in.DirEdge >= 0:
			ni.zh.Move(in.DirEdge)
		}
		return false, false
	}
	confirm := in.Confirm
	if tapIdx >= 0 {
		ni.cursor, confirm = tapIdx, true
	}
	switch { // 英數格盤
	case in.Cancel:
		return false, true
	case confirm:
		switch ni.cursor {
		case niCellBS: // 退格
			if len(ni.nameBuf) > 0 {
				ni.nameBuf = ni.nameBuf[:len(ni.nameBuf)-1]
			}
		case niCellOK: // 完成
			return true, false
		default: // 字元 → 附加
			if len(ni.nameBuf) < niNameMax {
				ni.nameBuf = append(ni.nameBuf, niCellGlyph(ni.cursor))
			}
		}
	case in.DirEdge == 3:
		ni.cursor = (ni.cursor + 1) % niCells
	case in.DirEdge == 2:
		ni.cursor = (ni.cursor + niCells - 1) % niCells
	case in.DirEdge == 0:
		ni.cursor = (ni.cursor + niCols) % niCells
	case in.DirEdge == 1:
		ni.cursor = (ni.cursor + niCells - niCols) % niCells
	}
	return false, false
}

// draw:已輸入名(黃字,起點 x0,y0)+ 注音組字列/候選/盤 或 英數格盤(格盤起點固定在 x0-16,y0+40,
// 對齊既有酒館版面幾何,原封不動移植自 tavern.go)。
func (ni *NameInput) draw(rgba []byte, tx *dq3data.Text, white, yellow dq3data.Color, x0, y0 int) {
	cur := func(x, y int, sel bool) {
		if sel {
			drawGlyph(rgba, tx, x-18, y, curGlyph, yellow)
		}
	}
	for i, g := range ni.nameBuf {
		drawGlyph(rgba, tx, x0+i*16, y0, g, yellow)
	}
	if ni.nameZhu { // 注音模式:組字列 + (候選 或 注音盤)
		cx := x0 + 120
		if ni.zh.Sh != 0 {
			drawGlyph(rgba, tx, cx, y0, 64+ni.zh.Sh, yellow)
			cx += 16
		}
		if ni.zh.Ji != 0 {
			drawGlyph(rgba, tx, cx, y0, 98+ni.zh.Ji, yellow)
			cx += 16
		}
		if ni.zh.Yu != 0 {
			drawGlyph(rgba, tx, cx, y0, 85+ni.zh.Yu, yellow)
		}
		ni.hits.reset()
		gy0 := y0 + 40
		if ni.zh.Pick { // 候選窗(9 欄)
			for i, g := range ni.zh.Cand {
				x, y := x0-16+(i%zhuyin.Cols)*24, gy0+(i/zhuyin.Cols)*22
				cur(x, y, i == ni.zh.Cursor)
				drawGlyph(rgba, tx, x, y, g, white)
				ni.hits.add(x-4, y-3, 22, 18, i)
			}
		} else { // 注音盤(9 欄)
			for c := 0; c < zhuyin.Cells; c++ {
				x, y := x0-16+(c%zhuyin.Cols)*24, gy0+(c/zhuyin.Cols)*22
				cur(x, y, c == ni.zh.Cursor)
				if g := zhuyin.CellGlyph(c); g >= 0 {
					drawGlyph(rgba, tx, x, y, g, white)
				} else { // 一聲:橫線
					for px := 2; px < 12; px++ {
						putPx(rgba, x+px, y+8, white)
					}
				}
				ni.hits.add(x-4, y-3, 22, 18, c)
			}
		}
		return
	}
	// 英數格盤(9 欄)+ ← / OK
	ni.hits.reset()
	gy0 := y0 + 40
	for c := 0; c < niCells; c++ {
		col, row := c%niCols, c/niCols
		x, y := x0-16+col*40, gy0+row*22
		if c == ni.cursor {
			cur(x, y, true)
		}
		if g := niCellGlyph(c); g >= 0 {
			drawGlyph(rgba, tx, x, y, g, white)
		} else if c == niCellBS {
			if ni.labels != nil {
				for j, gl := range ni.labels.Backspace {
					drawGlyph(rgba, tx, x+j*dq3data.GlyphPx, y, gl, white)
				}
			}
		} else {
			if ni.labels != nil {
				for j, gl := range ni.labels.OK {
					drawGlyph(rgba, tx, x+j*dq3data.GlyphPx, y, gl, white)
				}
			}
		}
		ni.hits.add(x-4, y-3, 36, 18, c)
	}
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
func (gs *GenderSelect) draw(rgba []byte, tx *dq3data.Text, white, yellow dq3data.Color, x, y0, dy int) {
	gs.hits.reset()
	for i := 0; i < 2; i++ {
		y := y0 + i*dy
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
		gs.hits.add(x-18, y-3, 120, 18, i)
	}
}
