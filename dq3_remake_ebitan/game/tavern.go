package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 露易達酒館招募(移植 dq3_tavern + dq3_nameinput 英數命名):職業 → 命名(英數格盤)→ 性別 → 建隊員。
// 注音(zhuyin)輸入為後續;此處只做英數(0-9 / A-Z)。DQ3_PARTY_MAX=4(隊長+3)。
const (
	tavClass  = 0 // 選職業
	tavName   = 1 // 命名(英數格盤)
	tavGender = 2 // 選性別

	niChars  = 36 // 0-9 + A-Z
	niCols   = 9
	niCells  = niChars + 2 // + ← + OK
	niCellBS = niChars     // 退格
	niCellOK = niChars + 1 // 完成
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
}

func (tv *Tavern) open() { tv.active, tv.stage, tv.cursor, tv.nameBuf = true, tavClass, 0, nil }

// input:回 (recruited *Member, closed)。
func (tv *Tavern) input(in InputState) (*Member, bool) {
	switch tv.stage {
	case tavClass:
		switch {
		case in.Cancel:
			tv.active = false
			return nil, true
		case in.Confirm:
			tv.pendCls, tv.stage, tv.cursor, tv.nameBuf = tv.cursor, tavName, 0, nil
		case in.DirEdge == 0:
			tv.cursor = (tv.cursor + 1) % 8
		case in.DirEdge == 1:
			tv.cursor = (tv.cursor + 7) % 8
		}
	case tavName:
		switch {
		case in.Cancel:
			tv.stage, tv.cursor = tavClass, tv.pendCls
		case in.Confirm:
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
		switch {
		case in.Cancel:
			tv.stage, tv.cursor = tavName, niCellOK
		case in.Confirm:
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
		for i := 0; i < 8; i++ {
			cur(80, 56+i*22, i == tv.cursor)
			for j, g := range classNames[i] {
				drawGlyph(rgba, tv.tx, 80+j*16, 56+i*22, g, white)
			}
		}
	case tavName:
		// 已輸入名
		for i, g := range tv.nameBuf {
			drawGlyph(rgba, tv.tx, 80+i*16, 56, g, yellow)
		}
		// 英數格盤(9 欄)+ ← / OK
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
		}
	case tavGender:
		for i, g := range classNames[tv.pendCls] { // 顯已選職業
			drawGlyph(rgba, tv.tx, 80+i*16, 56, g, white)
		}
		labels := [2][]int{{272}, {234}} // 男 / 女
		for i := 0; i < 2; i++ {
			cur(80, 100+i*22, i == tv.cursor)
			for j, g := range labels[i] {
				drawGlyph(rgba, tv.tx, 80+j*16, 100+i*22, g, white)
			}
		}
	}
}
