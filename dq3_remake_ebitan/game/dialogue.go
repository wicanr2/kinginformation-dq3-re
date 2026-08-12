package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// noVar 是插值 var(varItem/varNum)未設時的哨兵值 → 渲染為空白一格。
const noVar = -1

// Dialogue 是一個對話視窗狀態機:Open(rec) → 逐頁 Advance() → 關閉。
type Dialogue struct {
	tx        *dq3data.Text
	itemNames *dq3data.Text // 道具名表(D3TXT00.TXT,rec=code+1)。固定,獨立於 tx 目前切換的城鎮 bank;Game 初始化時同步 = g.shop.nameText。
	layout    gamepack.WindowLayout
	buf       []uint16
	pos       int
	open      bool

	// 插值 var context(docs/42 §四;比照 C dq3_text_set_var_*/dq3_text_clear_vars)。
	varItem    int              // VAR_ITEM 插值:道具 code(noVar=未設 → 空白)
	varNum     int              // VAR_NUM 插值:數值(noVar=未設 → 空白)
	varNumItem int              // VAR_NUM 在原版 mode=1 時取 item-name record（祭壇 rec93/96）
	heroName   []int            // VAR_ENT/VAR0/VAR7/VAR_IDX 插值:主角名 glyph 序列(Game 於創角/讀檔後同步;空=空白)
	varGlyph   map[uint16][]int // 事件可逐控制碼指定隊員／職業等原版變數；未指定才走既有語意
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
	d.varGlyph = nil
	return true
}

// openPackText resolves a stable game-pack text ID to canonical glyph/control
// words. Missing references fail closed; production code never substitutes a
// Go string or a legacy record number.
func (g *Game) openPackText(id string) bool {
	if g.pack == nil {
		return false
	}
	codes, ok := g.pack.TextGlyphCodes(id)
	if !ok {
		return false
	}
	return g.dlg.openRecord(codes)
}

// varGlyphs 依插值控制碼種類回目前應插入的字模序列;var 未設或無資料 → nil(渲染空白一格)。
func (d *Dialogue) varGlyphs(code uint16) []int {
	if glyphs, ok := d.varGlyph[code]; ok {
		return glyphs
	}
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
			if line >= d.layout.LinesPerPage {
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
				if col >= d.layout.Columns {
					col, line = 0, line+1
					if line >= d.layout.LinesPerPage {
						return i, false
					}
				}
			}
		case v >= 0xffed: // 未知保留控制碼(不在六個已知插值碼內)→ 空白一格,舊行為 fallback
			i, col = i+1, col+1
			if col >= d.layout.Columns {
				col, line = 0, line+1
			}
			if line >= d.layout.LinesPerPage {
				return i, false
			}
		default:
			i, col = i+1, col+1
			if col >= d.layout.Columns {
				col, line = 0, line+1
				if line >= d.layout.LinesPerPage {
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

// setDlgVarGlyphs binds one original control-code family to an explicit glyph
// sequence. Reclass records use VAR_ENT for the selected member and VAR_NUM
// for the class name in the same record, so collapsing all variables to the
// hero name would render the original text incorrectly.
func (g *Game) setDlgVarGlyphs(code uint16, glyphs []int) {
	if g.dlg.varGlyph == nil {
		g.dlg.varGlyph = make(map[uint16][]int)
	}
	g.dlg.varGlyph[code] = append([]int(nil), glyphs...)
}

// draw 把當前頁畫進 rgba(黑底 + 白框 + 白字)。移植 dq3_dialogue_render。
func (d *Dialogue) draw(rgba []byte, white dq3data.Color) {
	if !d.open {
		return
	}
	w := d.layout
	dark := dq3data.Color{R: 0, G: 0, B: 0}
	for r := 0; r < w.Height; r++ {
		for c := 0; c < w.Width; c++ {
			col := dark
			if r == 0 || r == w.Height-1 || c == 0 || c == w.Width-1 { // 白框
				col = white
			}
			putPx(rgba, w.X+c, w.Y+r, col)
		}
	}
	end, _ := d.scanPage(d.pos)
	col, line := 0, 0
	x0, y0 := w.X+w.TextInsetX, w.Y+w.TextInsetY
loop:
	for i := d.pos; i < end && i < len(d.buf); {
		v := d.buf[i]
		switch {
		case v == dq3data.TxtPage:
			break loop
		case v == dq3data.TxtNL || v == dq3data.TxtNL2:
			col, line = 0, line+1
			i++
			if line >= w.LinesPerPage {
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
				if col >= w.Columns {
					col, line = 0, line+1
					if line >= w.LinesPerPage {
						break loop
					}
				}
			}
		case v >= 0xffed: // 未知保留控制碼 → 空白一格
			col++
			i++
			if col >= w.Columns {
				col, line = 0, line+1
			}
		default:
			if v < dq3data.GlyphMax {
				d.drawGlyph(rgba, x0+col*dq3data.GlyphPx, y0+line*dq3data.GlyphPx, int(v), white)
			}
			col++
			i++
			if col >= w.Columns {
				col, line = 0, line+1
				if line >= w.LinesPerPage {
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

// fillBoxStyle 以 pack 提供的內部色、框線色與已註冊 pattern 繪製視窗。
// pattern 是共用 engine primitive 的名稱，不是可執行的 JSON code；顏色與
// 幾何仍屬於 game pack。checkerboard_1px 保留 EGA writer 未寫入的底圖像素。
func fillBoxStyle(rgba []byte, x, y, w, h int, interior, border dq3data.Color, pattern string) {
	fillBoxStyleWithEdgeWidths(rgba, x, y, w, h, interior, border, pattern, gamepack.FrameEdgeWidths{})
}

// fillBoxStyleWithEdgeWidths keeps the named frame primitive in the engine
// while letting a game pack narrow one side of a shared seam. Zero widths use
// the primitive default; no coordinates or version-specific branch enter the
// renderer.
func fillBoxStyleWithEdgeWidths(rgba []byte, x, y, w, h int, interior, border dq3data.Color, pattern string, widths gamepack.FrameEdgeWidths) {
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			col := interior
			if frameEdgeWithWidths(pattern, c, r, w, h, widths) {
				if !frameBorderPixel(pattern, c, r, w, h) {
					continue
				}
				col = border
			}
			putPx(rgba, x+c, y+r, col)
		}
	}
}

// strokeBoxStyle 只畫外框，不改寫內部底圖。能力確認的
// confirm_choice 先覆蓋一個較大的藍黑棋盤，再由這個 primitive 把
// lavender 外框疊回去；若直接呼叫 fillBoxStyle，會把藍色奇偶 phase
// 重新填成黑色，破壞原版的 plane/latch 可見結果。
func strokeBoxStyle(rgba []byte, x, y, w, h int, border dq3data.Color, pattern string) {
	if w <= 0 || h <= 0 {
		return
	}
	for r := 0; r < h; r++ {
		for c := 0; c < w; c++ {
			if !frameEdge(pattern, c, r, w, h) {
				continue
			}
			if frameBorderPixel(pattern, c, r, w, h) {
				putPx(rgba, x+c, y+r, border)
			}
		}
	}
}

// frameEdge returns the pixels owned by a named frame primitive. The original
// creation panel has a two-pixel bevel: its outermost four corners retain the
// underlying screen, while the inner edge remains continuous. confirm_choice
// has its own checkerboard_frame_2px primitive and does not come through this
// path.
func frameEdge(pattern string, x, y, w, h int) bool {
	return frameEdgeWithWidths(pattern, x, y, w, h, gamepack.FrameEdgeWidths{})
}

func frameEdgeWithWidths(pattern string, x, y, w, h int, widths gamepack.FrameEdgeWidths) bool {
	left, right, top, bottom := 1, 1, 1, 1
	if pattern == "solid_2px" || pattern == "beveled_2px" {
		left, right, top, bottom = 2, 2, 2, 2
	}
	if widths.Left > 0 {
		left = widths.Left
	}
	if widths.Right > 0 {
		right = widths.Right
	}
	if widths.Top > 0 {
		top = widths.Top
	}
	if widths.Bottom > 0 {
		bottom = widths.Bottom
	}
	return x < left || y < top || x >= w-right || y >= h-bottom
}

func strokeFrameStyle(rgba []byte, frame *gamepack.FrameStyle, x, y, w, h int, fallback dq3data.Color) {
	if frame == nil {
		strokeBoxStyle(rgba, x, y, w, h, fallback, "solid")
		return
	}
	if frame.BorderPattern == "checkerboard_frame_2px" && frame.BorderAccentRGB != nil {
		strokeCheckerboardFrame2px(rgba, x, y, w, h,
			frameColor(frame.BorderRGB), frameColor(*frame.BorderAccentRGB))
		return
	}
	strokeBoxStyle(rgba, x, y, w, h, frameColor(frame.BorderRGB), frame.BorderPattern)
}

// strokeCheckerboardFrame2px 是原版 confirm_choice 的已命名 EGA frame
// primitive。它不是任意 JSON code：兩個外框 band 使用 border／accent 交錯色，
// 兩側各兩欄，四個角依 writer 的 edge phase 留在正確的 band 上。
func strokeCheckerboardFrame2px(rgba []byte, x, y, w, h int, border, accent dq3data.Color) {
	if w < 4 || h < 4 {
		return
	}
	alt := func(n int, borderFirst bool) dq3data.Color {
		if (n&1 == 1) == borderFirst {
			return border
		}
		return accent
	}
	// Vertical bands first; horizontal bands then own the corner phase.
	for r := 1; r < h-1; r++ {
		putPx(rgba, x, y+r, alt(r, true))
		putPx(rgba, x+1, y+r, alt(r, false))
		putPx(rgba, x+w-2, y+r, alt(r, true))
		putPx(rgba, x+w-1, y+r, alt(r, false))
	}
	for c := 1; c < w-1; c++ {
		putPx(rgba, x+c, y, alt(c, true))
		putPx(rgba, x+c, y+h-1, alt(c, false))
	}
	for c := 0; c < w; c++ {
		putPx(rgba, x+c, y+1, alt(c, false))
		putPx(rgba, x+c, y+h-2, alt(c, true))
	}
	// The top writer leaves the outermost corner on the underlying backdrop;
	// its inner right corner is the border phase before the vertical writer's
	// first interior row.  Keep this explicit edge phase instead of letting the
	// horizontal alternating sequence wrap one pixel early.
	putPx(rgba, x+w-2, y, border)
	putPx(rgba, x, y+h-1, border)
}

func frameBorderPixel(pattern string, x, y, w, h int) bool {
	switch pattern {
	case "solid", "solid_2px", "beveled_2px":
		if pattern == "beveled_2px" &&
			(x == 0 || x == w-1) && (y == 0 || y == h-1) {
			return false
		}
		return true
	case "checkerboard_1px":
		// sub_1fdb1 starts each EGA edge with bh=0xaa; the visible logical
		// mask is the local odd checker phase, leaving the underlying pixel.
		return (x+y)&1 == 1
	default:
		// Invalid patterns are rejected by the pack validator. Keep direct
		// fixtures visually closed rather than inventing a new pattern.
		return false
	}
}

// fillBox 保留給沒有安裝 pack 的輕量測試 fixture；正式畫面一律透過
// fillPackBox 取得 JSON 的 frame 顏色。
func fillBox(rgba []byte, x, y, w, h int, border dq3data.Color) {
	fillBoxStyle(rgba, x, y, w, h, dq3data.Color{}, border, "solid")
}

func frameColor(rgb [3]uint8) dq3data.Color {
	return dq3data.Color{R: rgb[0], G: rgb[1], B: rgb[2]}
}

// fillGeometryBox 套用 new_game_geometry 的版本專屬 frame。直接 UI fixture
// 未安裝 frame 時仍保留導覽用 fallback；正式 pack 由 validator 先拒絕缺欄位。
func fillGeometryBox(rgba []byte, frame *gamepack.FrameStyle, x, y, w, h int, fallback dq3data.Color) {
	if frame == nil {
		fillBox(rgba, x, y, w, h, fallback)
		return
	}
	fillBoxStyle(rgba, x, y, w, h,
		frameColor(frame.InteriorRGB), frameColor(frame.BorderRGB), frame.BorderPattern)
}

// fillGeometryFramedRect is the data-shaped companion to fillGeometryBox.
// It applies an optional per-edge width carried by GeometryRect; callers that
// only have scalar geometry keep using fillGeometryBox.
func fillGeometryFramedRect(rgba []byte, frame *gamepack.FrameStyle, rect gamepack.GeometryRect, fallback dq3data.Color) {
	if frame == nil {
		fillBox(rgba, rect.X, rect.Y, rect.Width, rect.Height, fallback)
		return
	}
	fillBoxStyleWithEdgeWidths(rgba, rect.X, rect.Y, rect.Width, rect.Height,
		frameColor(frame.InteriorRGB), frameColor(frame.BorderRGB), frame.BorderPattern, rect.FrameEdgeWidths)
}

// strokeGeometryBox 套用 pack frame 的外框但保留內部像素。正式 pack 會
// 先驗證 frame；nil 僅供 direct fixture 導覽，不是 production fallback。
func strokeGeometryBox(rgba []byte, frame *gamepack.FrameStyle, x, y, w, h int, fallback dq3data.Color) {
	strokeFrameStyle(rgba, frame, x, y, w, h, fallback)
}

// fillGeometryRect 只填 pack 指定的內容矩形。confirm_choice 的原版結果是
// 「外圍藍黑棋盤 + 中央 80x32 黑色內容區 + lavender EGA frame」，因此
// 不能把整個 frame interior 當成黑色而抹掉內容區外的 checker phase。
func fillGeometryRect(rgba []byte, frame *gamepack.FrameStyle, rect gamepack.GeometryRect, fallback dq3data.Color) {
	interior := fallback
	if frame != nil {
		interior = frameColor(frame.InteriorRGB)
	}
	for r := 0; r < rect.Height; r++ {
		for c := 0; c < rect.Width; c++ {
			putPx(rgba, rect.X+c, rect.Y+r, interior)
		}
	}
}

// fillPatternRect 寫入已註冊 pattern 命中的 phase，未命中的 phase 保留
// 呼叫前的像素。這是 EGA 選項背景需要的非破壞性 plane-mask 語意，不把
// DQ3 的座標或顏色硬寫在共用引擎。
func fillPatternRect(rgba []byte, fill *gamepack.PatternFill) {
	if fill == nil {
		return
	}
	for r := 0; r < fill.Rect.Height; r++ {
		for c := 0; c < fill.Rect.Width; c++ {
			if !frameBorderPixel(fill.Pattern, c, r, fill.Rect.Width, fill.Rect.Height) {
				continue
			}
			putPx(rgba, fill.Rect.X+c, fill.Rect.Y+r, frameColor(fill.RGB))
		}
	}
}

// fillPackBox 套用版本專屬 frame。direct unit fixture 若未安裝 frame，
// 退回既有黑底白框只供測試導覽；NewGameWithPack 會先拒絕此資料缺口。
func fillPackBox(rgba []byte, layout gamepack.WindowLayout, x, y, w, h int) {
	if layout.Frame == nil {
		fillBox(rgba, x, y, w, h, dq3data.Color{R: 255, G: 255, B: 255})
		return
	}
	fillBoxStyle(rgba, x, y, w, h,
		frameColor(layout.Frame.InteriorRGB), frameColor(layout.Frame.BorderRGB), layout.Frame.BorderPattern)
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
