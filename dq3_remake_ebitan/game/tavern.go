package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/zhuyin"
)

// 露易達酒館招募(移植 dq3_tavern + dq3_nameinput 英數命名):職業 → 命名(英數格盤)→ 性別 → 建隊員。
// 注音(zhuyin)輸入為後續;此處只做英數(0-9 / A-Z)。DQ3_PARTY_MAX=4(隊長+3)。
const (
	tavClass  = 0 // 選職業
	tavName   = 1 // 命名(英數格盤)
	tavGender = 2 // 選性別

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

type Tavern struct {
	tx      *dq3data.Text
	active  bool
	stage   int
	cursor  int
	pendCls int
	nameBuf []int
	nameZhu bool // true=注音模式(Tab 切換)
	zh      zhuyin.Composer
	hits    hitList // 目前畫面(職業/英數格/注音盤/候選/性別)的可點區塊,draw() 重建
}

func (tv *Tavern) open() { tv.active, tv.stage, tv.cursor, tv.nameBuf = true, tavClass, 0, nil }

// input:回 (recruited *Member, closed)。觸控直接點選(P2):tapIdx 命中 tv.hits(上一幀 draw()
// 建的、對齊當前畫面的格子)→ 游標移過去 + 等同 A 確定,十字鍵 + A 保底並存。
func (tv *Tavern) input(in InputState) (*Member, bool) {
	tapIdx := -1
	if in.Tapped {
		tapIdx = tv.hits.at(in.TapX, in.TapY)
	}
	switch tv.stage {
	case tavClass:
		confirm := in.Confirm
		if tapIdx >= 0 {
			tv.cursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			tv.active = false
			return nil, true
		case confirm:
			tv.pendCls, tv.stage, tv.cursor, tv.nameBuf = tv.cursor, tavName, 0, nil
		case in.DirEdge == 0:
			tv.cursor = (tv.cursor + 1) % 8
		case in.DirEdge == 1:
			tv.cursor = (tv.cursor + 7) % 8
		}
	case tavName:
		if in.Toggle || in.CtxTap { // Tab:英數 ↔ 注音(觸控:情境鍵)
			tv.nameZhu = !tv.nameZhu
			if tv.nameZhu {
				tv.zh.Init()
			} else {
				tv.cursor = 0
			}
			return nil, false
		}
		if tv.nameZhu { // 注音組字(盤格 或 候選格,視 tv.zh.Pick 而定;hits 對齊 draw() 當前那組)
			confirm := in.Confirm
			if tapIdx >= 0 {
				tv.zh.Cursor, confirm = tapIdx, true
			}
			switch {
			case in.Cancel:
				if !tv.zh.Cancel() { // 非候選階段 → 退英數
					tv.nameZhu, tv.cursor = false, 0
				}
			case confirm:
				if g, picked := tv.zh.Confirm(); picked && len(tv.nameBuf) < niNameMax {
					tv.nameBuf = append(tv.nameBuf, g)
				}
			case in.DirEdge >= 0:
				tv.zh.Move(in.DirEdge)
			}
			return nil, false
		}
		confirm := in.Confirm
		if tapIdx >= 0 {
			tv.cursor, confirm = tapIdx, true
		}
		switch { // 英數格盤
		case in.Cancel:
			tv.stage, tv.cursor = tavClass, tv.pendCls
		case confirm:
			switch tv.cursor {
			case niCellBS: // 退格
				if len(tv.nameBuf) > 0 {
					tv.nameBuf = tv.nameBuf[:len(tv.nameBuf)-1]
				}
			case niCellOK: // 完成 → 性別
				tv.stage, tv.cursor = tavGender, 0
			default: // 字元 → 附加
				if len(tv.nameBuf) < niNameMax {
					tv.nameBuf = append(tv.nameBuf, niCellGlyph(tv.cursor))
				}
			}
		case in.DirEdge == 3:
			tv.cursor = (tv.cursor + 1) % niCells
		case in.DirEdge == 2:
			tv.cursor = (tv.cursor + niCells - 1) % niCells
		case in.DirEdge == 0:
			tv.cursor = (tv.cursor + niCols) % niCells
		case in.DirEdge == 1:
			tv.cursor = (tv.cursor + niCells - niCols) % niCells
		}
	case tavGender:
		confirm := in.Confirm
		if tapIdx >= 0 {
			tv.cursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			tv.stage, tv.cursor = tavName, niCellOK
		case confirm:
			name := tv.nameBuf
			if len(name) == 0 { // 未命名 → 用職業名
				name = classNames[tv.pendCls]
			}
			m := newMember(name, tv.pendCls, tv.cursor, 0)
			tv.active = false
			return m, true
		case in.DirEdge == 0 || in.DirEdge == 1:
			tv.cursor ^= 1
		}
	}
	return nil, false
}

func (tv *Tavern) draw(rgba []byte, white dq3data.Color) {
	if !tv.active {
		return
	}
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	cur := func(x, y int, sel bool) {
		if sel {
			drawGlyph(rgba, tv.tx, x-18, y, curGlyph, yellow)
		}
	}
	switch tv.stage {
	case tavClass:
		tv.hits.reset()
		for i := 0; i < 8; i++ {
			y := 56 + i*22
			cur(80, y, i == tv.cursor)
			for j, g := range classNames[i] {
				drawGlyph(rgba, tv.tx, 80+j*16, y, g, white)
			}
			tv.hits.add(62, y-3, 200, 18, i)
		}
	case tavName:
		// 已輸入名
		for i, g := range tv.nameBuf {
			drawGlyph(rgba, tv.tx, 80+i*16, 56, g, yellow)
		}
		if tv.nameZhu { // 注音模式:組字列 + (候選 或 注音盤)
			// 組字列:已選 聲母(64+Sh)/ 介音(98+Ji)/ 韻母(85+Yu)
			cx := 200
			if tv.zh.Sh != 0 {
				drawGlyph(rgba, tv.tx, cx, 56, 64+tv.zh.Sh, yellow)
				cx += 16
			}
			if tv.zh.Ji != 0 {
				drawGlyph(rgba, tv.tx, cx, 56, 98+tv.zh.Ji, yellow)
				cx += 16
			}
			if tv.zh.Yu != 0 {
				drawGlyph(rgba, tv.tx, cx, 56, 85+tv.zh.Yu, yellow)
			}
			tv.hits.reset()
			if tv.zh.Pick { // 候選窗(9 欄)
				for i, g := range tv.zh.Cand {
					x, y := 64+(i%zhuyin.Cols)*24, 96+(i/zhuyin.Cols)*22
					cur(x, y, i == tv.zh.Cursor)
					drawGlyph(rgba, tv.tx, x, y, g, white)
					tv.hits.add(x-4, y-3, 22, 18, i)
				}
			} else { // 注音盤(9 欄,對齊 C)
				for c := 0; c < zhuyin.Cells; c++ {
					x, y := 64+(c%zhuyin.Cols)*24, 96+(c/zhuyin.Cols)*22
					cur(x, y, c == tv.zh.Cursor)
					if g := zhuyin.CellGlyph(c); g >= 0 {
						drawGlyph(rgba, tv.tx, x, y, g, white)
					} else { // 一聲:橫線(對齊 C)
						for px := 2; px < 12; px++ {
							putPx(rgba, x+px, y+8, white)
						}
					}
					tv.hits.add(x-4, y-3, 22, 18, c)
				}
			}
			return
		}
		// 英數格盤(9 欄)+ ← / OK
		tv.hits.reset()
		for c := 0; c < niCells; c++ {
			col, row := c%niCols, c/niCols
			x, y := 64+col*40, 96+row*22
			if c == tv.cursor {
				cur(x, y, true)
			}
			if g := niCellGlyph(c); g >= 0 {
				drawGlyph(rgba, tv.tx, x, y, g, white)
			} else if c == niCellBS {
				drawGlyph(rgba, tv.tx, x, y, 11, white) // ← 近似
			} else {
				drawGlyph(rgba, tv.tx, x, y, 665, white) // OK≈「狀」借字(近似)
			}
			tv.hits.add(x-4, y-3, 36, 18, c)
		}
	case tavGender:
		for i, g := range classNames[tv.pendCls] { // 顯已選職業
			drawGlyph(rgba, tv.tx, 80+i*16, 56, g, white)
		}
		labels := [2][]int{{272}, {234}} // 男 / 女
		tv.hits.reset()
		for i := 0; i < 2; i++ {
			y := 100 + i*22
			cur(80, y, i == tv.cursor)
			for j, g := range labels[i] {
				drawGlyph(rgba, tv.tx, 80+j*16, y, g, white)
			}
			tv.hits.add(62, y-3, 120, 18, i)
		}
	}
}
