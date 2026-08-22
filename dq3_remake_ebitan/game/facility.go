package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 城鎮設施（type 0=旅社、1=武防店、2=道具店、3=教會、4=記錄點）。
// NPC 子型 (ctrl>>3)&7 >=3 才是 consumer，byte4 是 block index；shopdata.go 保存 raw
// block inventory，孤立 block 或 section 不吻合的 row 不等於玩家可用設施。
const (
	facInn = iota
	facWeapon
	facItem
	facChurch
	facRecord
)

type facility struct {
	k, typ, count, itemOff, innCost int
}

type churchStage int

const (
	churchService churchStage = iota
	churchTarget
	churchConfirm
)

// Church 是教會三項服務 modal。原版 handler 0x83d8 的服務順序為解毒／解詛咒／復活；
// 三者均由 pack 定價，解詛咒的 ITEM metadata→equipped word→教會 consumer 見 docs/147。
type Church struct {
	active         bool
	stage          churchStage
	serviceCursor  int
	targetCursor   int
	confirmCursor  int // 0=是、1=否
	pendingTarget  int
	pendingService int
	pendingCost    int
	msg            string
	tx             *dq3data.Text
}

// Shop 是開啟中的商店狀態(武防/道具店)。
type Shop struct {
	items        *dq3data.Items
	tx           *dq3data.Text
	nameText     *dq3data.Text // 品名:D3TXT00.TXT record = code+1
	active       bool
	codes        []int // 貨架品項 id
	cursor       int
	msg          string
	hits         hitList // 貨架列可點區塊,draw() 重建(P2 直接點選)
	targeting    bool
	pendingCode  int
	targetCursor int
	targetHits   hitList
}

// itemNameGlyphs 取道具 code 的品名 glyph 序列(D3TXT00 rec=code+1;跳過控制/插值碼)。
// drawItemName 與對話插值 VAR_ITEM(dialogue.go varGlyphs)共用同一套查表,別各自重寫。
func itemNameGlyphs(t *dq3data.Text, code int) []int {
	if t == nil {
		return nil
	}
	rec := t.Record(code + 1)
	if rec == nil {
		return nil
	}
	out := make([]int, 0, len(rec))
	for _, v := range rec {
		if v >= dq3data.GlyphMax { // 控制/插值碼跳過
			continue
		}
		out = append(out, int(v))
	}
	return out
}

// drawItemName 畫道具名(D3TXT00 rec=code+1 的 glyph 序列)到 (x,y)。
func (s *Shop) drawItemName(rgba []byte, x, y, code int, fg dq3data.Color) {
	if s.nameText == nil {
		return
	}
	for i, v := range itemNameGlyphs(s.nameText, code) {
		drawGlyph(rgba, s.nameText, x+i*dq3data.GlyphPx, y, v, fg)
	}
}

func (s *Shop) open(codes []int) {
	s.codes, s.cursor, s.active, s.msg = codes, 0, true, ""
	s.targeting, s.pendingCode, s.targetCursor = false, -1, 0
}

// input 處理商店輸入;回傳 (buyCode, closed):buyCode>=0 表要買該 id。
// 點貨架列(P2)= 游標移過去 + 等同 A 購買。
func (s *Shop) input(in InputState) (buyCode int, closed bool) {
	buyCode = -1
	confirm := in.Confirm
	if in.Tapped {
		if idx := s.hits.at(in.TapX, in.TapY); idx >= 0 && idx < len(s.codes) {
			s.cursor, confirm = idx, true
		}
	}
	switch {
	case in.Cancel:
		s.active = false
		return -1, true
	case confirm:
		if s.cursor >= 0 && s.cursor < len(s.codes) {
			buyCode = s.codes[s.cursor]
		}
	case in.DirEdge == 0: // 下
		s.cursor = (s.cursor + 1) % len(s.codes)
	case in.DirEdge == 1: // 上
		s.cursor = (s.cursor + len(s.codes) - 1) % len(s.codes)
	}
	return buyCode, false
}

// draw 畫商店:貨架(品項 id + 價)+ 游標 + 持有金。品名 glyph 之後接(先顯示 id/價)。
func (s *Shop) draw(rgba []byte, gold int, white dq3data.Color) {
	fillBox(rgba, 24, 24, ScreenW-48, ScreenH-96, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	s.hits.reset()
	for i, code := range s.codes {
		y := 40 + i*22
		if i == s.cursor {
			drawGlyph(rgba, s.tx, 40, y, 11, yellow) // ► 游標
		}
		// 品名(D3TXT00 rec=code+1)
		s.drawItemName(rgba, 64, y, code, white)
		// 價(右欄)
		drawNumber(rgba, s.tx, 380, y, s.items.Price(code), white)
		s.hits.add(24, y-3, ScreenW-48, 20, i)
	}
	// 持有金(底部)
	drawNumber(rgba, s.tx, 64, ScreenH-84, gold, yellow)
}
