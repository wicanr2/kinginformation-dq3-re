package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 對話視窗(底部,DQ 風格,4 行 16×16 字模)。幾何/分頁 1:1 移植 C dq3_dialogue.c。
const (
	dlgLines   = 4
	winX, winY = 24, 244
	winW, winH = ScreenW - 48, 96
	dlgPad     = 12
	dlgCols    = (winW - 2*dlgPad) / dq3data.GlyphPx // 35
)

// Dialogue 是一個對話視窗狀態機:Open(rec) → 逐頁 Advance() → 關閉。
type Dialogue struct {
	tx   *dq3data.Text
	buf  []uint16
	pos  int
	open bool
}

// Open 從當前 bank 取記錄 rec 進視窗。移植 dq3_dialogue_open。
func (d *Dialogue) Open(rec int) bool {
	b := d.tx.Record(rec)
	if b == nil {
		return false
	}
	d.buf, d.pos, d.open = b, 0, true
	return true
}

// scanPage:從 pos 掃到本頁結尾,回 (下一頁起點, 是否已到記錄結尾)。移植 scan_page。
func (d *Dialogue) scanPage(pos int) (int, bool) {
	col, line, i := 0, 0, pos
	for i < len(d.buf) {
		v := d.buf[i]
		switch {
		case v == dq3data.TxtPage:
			return i + 1, false
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			col, line, i = 0, line+1, i+1
			if line >= dlgLines {
				return i, false
			}
		case v >= 0xffed: // 插值占位 → 空白一格
			i, col = i+1, col+1
			if col >= dlgCols {
				col, line = 0, line+1
			}
			if line >= dlgLines {
				return i, false
			}
		default:
			i, col = i+1, col+1
			if col >= dlgCols {
				col, line = 0, line+1
				if line >= dlgLines {
					return i, false
				}
			}
		}
	}
	return i, true // eof
}

// Advance:推進到下一頁;已到結尾則關閉。移植 dq3_dialogue_advance。
func (d *Dialogue) Advance() {
	if !d.open {
		return
	}
	next, eof := d.scanPage(d.pos)
	if eof {
		d.open = false
		return
	}
	d.pos = next
}

// draw 把當前頁畫進 rgba(黑底 + 白框 + 白字)。移植 dq3_dialogue_render。
func (d *Dialogue) draw(rgba []byte, white dq3data.Color) {
	if !d.open {
		return
	}
	dark := dq3data.Color{R: 0, G: 0, B: 0}
	for r := 0; r < winH; r++ {
		for c := 0; c < winW; c++ {
			col := dark
			if r == 0 || r == winH-1 || c == 0 || c == winW-1 { // 白框
				col = white
			}
			putPx(rgba, winX+c, winY+r, col)
		}
	}
	end, _ := d.scanPage(d.pos)
	col, line := 0, 0
	x0, y0 := winX+dlgPad, winY+dlgPad
loop:
	for i := d.pos; i < end && i < len(d.buf); i++ {
		v := d.buf[i]
		switch {
		case v == dq3data.TxtPage:
			break loop
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			col, line = 0, line+1
			if line >= dlgLines {
				break loop
			}
		case v >= 0xffed: // 插值占位 → 空白
			col++
			if col >= dlgCols {
				col, line = 0, line+1
			}
		default:
			if v < dq3data.GlyphMax {
				d.drawGlyph(rgba, x0+col*dq3data.GlyphPx, y0+line*dq3data.GlyphPx, int(v), white)
			}
			col++
			if col >= dlgCols {
				col, line = 0, line+1
				if line >= dlgLines {
					break loop
				}
			}
		}
	}
}

func (d *Dialogue) drawGlyph(rgba []byte, px, py, idx int, fg dq3data.Color) {
	drawGlyph(rgba, d.tx, px, py, idx, fg)
}

// drawGlyph:把 FON 第 idx 個 16×16 字模畫到 (px,py),set bit → fg。選單/HUD 共用。
func drawGlyph(rgba []byte, tx *dq3data.Text, px, py, idx int, fg dq3data.Color) {
	g, ok := tx.Glyph(idx)
	if !ok {
		return
	}
	for r := 0; r < 16; r++ {
		for c := 0; c < 16; c++ {
			if g[r][c] != 0 {
				putPx(rgba, px+c, py+r, fg)
			}
		}
	}
}

// fillBox:黑底 + 白框(選單/視窗共用)。
func fillBox(rgba []byte, x, y, w, h int, border dq3data.Color) {
	dark := dq3data.Color{R: 0, G: 0, B: 0}
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			col := dark
			if r == 0 || r == h-1 || c == 0 || c == w-1 {
				col = border
			}
			putPx(rgba, x+c, y+r, col)
		}
	}
}

func putPx(rgba []byte, x, y int, c dq3data.Color) {
	if x < 0 || x >= ScreenW || y < 0 || y >= ScreenH {
		return
	}
	o := (y*ScreenW + x) * 4
	rgba[o], rgba[o+1], rgba[o+2], rgba[o+3] = c.R, c.G, c.B, 255
}

// frontTile:主角面向的前一格(facing 0下 1上 2左 3右,對齊 Update 的方向鍵映射)。
func frontTile(x, y, facing int) (int, int) {
	switch facing {
	case 0:
		return x, y + 1
	case 1:
		return x, y - 1
	case 2:
		return x - 1, y
	case 3:
		return x + 1, y
	}
	return x, y
}
