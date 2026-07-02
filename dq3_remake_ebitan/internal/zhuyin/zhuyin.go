// Package zhuyin 是注音輸入法(組字 + 同音候選挑字)。移植 dq3_zhuyin.c + dq3_zhuyin_table.c。
// 注音盤 42 格:聲母 0-20(glyph 65+)、韻母 21-33(86+)、介音 34-36(99+)、聲調 37-41。
// key=((聲母*4+介音)*14+韻母)*5+聲調 → 二分查 pool 得同音字模 glyph。
package zhuyin

const (
	Cells   = 42 // 21 聲母 + 13 韻母 + 3 介音 + 5 聲調
	Cols    = 9  // 注音盤欄數(對齊 C DQ3_ZH_COLS)
	MaxCand = 32 // 候選上限
)

// CellGlyph 回注音盤第 cell 格的顯示 glyph(聲調 37-40 有符號、41 一聲無 → -1)。移植 cell_glyph。
func CellGlyph(cell int) int {
	switch {
	case cell < 21:
		return 65 + cell // 聲母
	case cell < 34:
		return 86 + (cell - 21) // 韻母
	case cell < 37:
		return 99 + (cell - 34) // 介音
	}
	switch cell { // 聲調
	case 37:
		return 103
	case 38:
		return 104
	case 39:
		return 105
	case 40:
		return 102
	}
	return -1 // 41 = 一聲(無符號)
}

// Lookup 查同音字模 glyph(移植 dq3_zhuyin_lookup;二分查 packed key)。
func Lookup(sh, ji, yu, tone int) []int {
	if tone < 0 {
		tone = 0
	}
	key := ((sh*4+ji)*14+yu)*5 + tone
	lo, hi := 0, nbucket-1
	for lo <= hi {
		mid := (lo + hi) / 2
		switch {
		case zhKey[mid] == key:
			a, b := zhOff[mid], zhOff[mid+1]
			out := make([]int, 0, b-a)
			for i := a; i < b; i++ {
				out = append(out, zhPool[i])
			}
			return out
		case zhKey[mid] < key:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return nil
}

// Composer 是注音組字狀態機(注音盤 → 候選)。Sh/Ji/Yu 供組字列顯示(0=未選)。
type Composer struct {
	Sh, Ji, Yu, tone int
	Cursor           int   // 注音盤 / 候選游標
	Cand             []int // 候選字模 glyph
	Pick             bool  // true=候選挑字階段
}

// Init 重設組字。
func (c *Composer) Init() { *c = Composer{tone: -1} }

// Move:游標移動(dir facing 碼)。注音盤 2D(9 欄,row/col 各自環繞 + clamp);候選線性。移植 dq3_zhuyin_input 導航。
func (c *Composer) Move(dir int) {
	if c.Pick { // 候選:2D(COLS 欄)
		n := len(c.Cand)
		if n == 0 {
			return
		}
		switch dir {
		case 2: // ← 左
			if c.Cursor > 0 {
				c.Cursor--
			} else {
				c.Cursor = n - 1
			}
		case 3: // → 右
			c.Cursor = (c.Cursor + 1) % n
		case 1: // ↑ 上
			if c.Cursor >= Cols {
				c.Cursor -= Cols
			}
		case 0: // ↓ 下
			if c.Cursor+Cols < n {
				c.Cursor += Cols
			}
		}
		return
	}
	// 注音盤:2D row/col 環繞(對齊 C compose 導航)
	rows := (Cells + Cols - 1) / Cols
	row, col := c.Cursor/Cols, c.Cursor%Cols
	switch dir {
	case 1: // 上
		row = (row + rows - 1) % rows
	case 0: // 下
		row = (row + 1) % rows
	case 2: // 左
		col = (col + Cols - 1) % Cols
	case 3: // 右
		col = (col + 1) % Cols
	}
	nc := row*Cols + col
	if nc >= Cells {
		nc = Cells - 1
	}
	c.Cursor = nc
}

// Confirm:選定當前格。回 (glyph, picked):picked=true 表挑定一個字(glyph)。移植 compose_select + PICK。
func (c *Composer) Confirm() (int, bool) {
	if c.Pick { // 候選挑字
		if c.Cursor >= 0 && c.Cursor < len(c.Cand) {
			g := c.Cand[c.Cursor]
			c.Init()
			return g, true
		}
		return -1, false
	}
	cell := c.Cursor
	switch {
	case cell < 21:
		c.Sh = cell + 1
	case cell < 34:
		c.Yu = cell - 21 + 1
	case cell < 37:
		c.Ji = cell - 34 + 1
	default: // 聲調 → 查表
		tone := 0
		switch cell {
		case 37:
			tone = 1
		case 38:
			tone = 2
		case 39:
			tone = 3
		case 40:
			tone = 4
		}
		c.tone = tone
		c.Cand = Lookup(c.Sh, c.Ji, c.Yu, tone)
		if len(c.Cand) > MaxCand {
			c.Cand = c.Cand[:MaxCand]
		}
		if len(c.Cand) > 0 {
			c.Pick, c.Cursor = true, 0
		} else {
			c.tone = -1 // 查無 → 重選
		}
	}
	return -1, false
}

// Cancel:候選階段退回注音盤;組字階段回 false(交上層關閉)。
func (c *Composer) Cancel() bool {
	if c.Pick {
		c.Pick, c.Cursor = false, 0
		return true
	}
	return false
}
