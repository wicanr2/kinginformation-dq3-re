package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// noVar 是插值 var(varItem/varNum)未設時的哨兵值 → 渲染為空白一格。
const noVar = -1

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
	tx        *dq3data.Text
	itemNames *dq3data.Text // 道具名表(D3TXT00.TXT,rec=code+1)。固定,獨立於 tx 目前切換的城鎮 bank;Game 初始化時同步 = g.shop.nameText。
	buf       []uint16
	pos       int
	open      bool

	// 插值 var context(docs/42 §四;比照 C dq3_text_set_var_*/dq3_text_clear_vars)。
	varItem    int   // VAR_ITEM 插值:道具 code(noVar=未設 → 空白)
	varNum     int   // VAR_NUM 插值:數值(noVar=未設 → 空白)
	varNumItem int   // VAR_NUM 在原版 mode=1 時取 item-name record（祭壇 rec93/96）
	heroName   []int // VAR_ENT/VAR0/VAR7/VAR_IDX 插值:主角名 glyph 序列(Game 於創角/讀檔後同步;空=空白)
}

// Open 從當前 bank 取記錄 rec 進視窗。移植 dq3_dialogue_open。
// 每次開新對話重置 varItem/varNum(比照 C dq3_text_clear_vars);heroName 是持久身分,不隨對話重置。
// 呼叫端若該訊息含 VAR_ITEM/VAR_NUM,須在 Open 後、下一次 renderFrame 前呼叫 setDlgVarItem/setDlgVarNum。
func (d *Dialogue) Open(rec int) bool {
	b := d.tx.Record(rec)
	return d.openRecord(b)
}

// OpenFrom 從另一個文字 bank 取 record，但維持目前城鎮 bank 作為後續一般對話來源。
// 字模常駐共用 D3TXT00.FON，因此 buffer 可由全域 D3TXT00.TXT 提供而不必切換 tx。
func (d *Dialogue) OpenFrom(tx *dq3data.Text, rec int) bool {
	if tx == nil {
		return false
	}
	return d.openRecord(tx.Record(rec))
}

func (d *Dialogue) openRecord(b []uint16) bool {
	if b == nil {
		return false
	}
	d.buf, d.pos, d.open = b, 0, true
	d.varItem, d.varNum, d.varNumItem = noVar, noVar, noVar
	return true
}

// varGlyphs 依插值控制碼種類回目前應插入的字模序列;var 未設或無資料 → nil(渲染空白一格)。
func (d *Dialogue) varGlyphs(code uint16) []int {
	switch code {
	case dq3data.TxtVarNum:
		if d.varNumItem >= 0 {
			return itemNameGlyphs(d.itemNames, d.varNumItem)
		}
		if d.varNum < 0 {
			return nil
		}
		return digitGlyphs(d.varNum)
	case dq3data.TxtVarItem:
		if d.varItem < 0 {
			return nil
		}
		return itemNameGlyphs(d.itemNames, d.varItem)
	case dq3data.TxtVarEnt, dq3data.TxtVar0, dq3data.TxtVar7, dq3data.TxtVarIdx:
		return d.heroName
	}
	return nil
}

// scanPage:從 pos 掃到本頁結尾,回 (下一頁起點, 是否已到記錄結尾)。移植 scan_page。
// 動態插值控制碼(0xfffb/fffa/fff9/fff6/fff5/ffed)消耗控制碼本身 +1 參數 word,展開的字模序列
// (道具名/數值/主角名,未設則視為 1 格空白)一樣走欄位/換行計數,避免展開後溢出視窗。
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
		case dq3data.IsVarInsert(v):
			n := len(d.varGlyphs(v))
			i++
			if i < len(d.buf) { // 消耗 +1 參數 word(不算欄位)
				i++
			}
			if n == 0 {
				n = 1 // 未設 var → 空白一格(維持版面,不可 0 寬)
			}
			for j := 0; j < n; j++ {
				col++
				if col >= dlgCols {
					col, line = 0, line+1
					if line >= dlgLines {
						return i, false
					}
				}
			}
		case v >= 0xffed: // 未知保留控制碼(不在六個已知插值碼內)→ 空白一格,舊行為 fallback
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

// setDlgVarItem/setDlgVarNum:設定即將渲染的對話插值 context(VAR_ITEM/VAR_NUM)。
// 呼叫時機:g.dlg.Open(rec) 之後(Open 會重置為 noVar)、renderFrame() 之前 —— 給物/金額事件路徑
// 在開「得到 X」/含金額對話時呼叫,讓該對話記錄若含 0xfffa/0xfff9 能正確展開實字。
func (g *Game) setDlgVarItem(code int) { g.dlg.varItem = code }
func (g *Game) setDlgVarNum(n int)     { g.dlg.varNum = n }
func (g *Game) setDlgVarNumItem(code int) {
	g.dlg.varNumItem = code
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
	for i := d.pos; i < end && i < len(d.buf); {
		v := d.buf[i]
		switch {
		case v == dq3data.TxtPage:
			break loop
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			col, line = 0, line+1
			i++
			if line >= dlgLines {
				break loop
			}
		case dq3data.IsVarInsert(v):
			glyphs := d.varGlyphs(v)
			i++
			if i < len(d.buf) { // 消耗 +1 參數 word(不畫)
				i++
			}
			n := len(glyphs)
			if n == 0 {
				n = 1 // 未設 var → 空白一格
			}
			for j := 0; j < n; j++ {
				if j < len(glyphs) {
					d.drawGlyph(rgba, x0+col*dq3data.GlyphPx, y0+line*dq3data.GlyphPx, glyphs[j], white)
				}
				col++
				if col >= dlgCols {
					col, line = 0, line+1
					if line >= dlgLines {
						break loop
					}
				}
			}
		case v >= 0xffed: // 未知保留控制碼 → 空白一格
			col++
			i++
			if col >= dlgCols {
				col, line = 0, line+1
			}
		default:
			if v < dq3data.GlyphMax {
				d.drawGlyph(rgba, x0+col*dq3data.GlyphPx, y0+line*dq3data.GlyphPx, int(v), white)
			}
			col++
			i++
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
