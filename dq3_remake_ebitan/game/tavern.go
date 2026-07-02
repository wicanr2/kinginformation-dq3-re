package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 露易達酒館招募(簡化:職業→性別→建 lv1 隊員加入隊伍)。移植 dq3_tavern 流程,
// 略過注音姓名輸入(zhuyin/nameinput 為後續),名字暫用職業名。DQ3_PARTY_MAX=4(隊長+3)。
type Tavern struct {
	tx       *dq3data.Text
	active   bool
	stage    int // 0=選職業、1=選性別
	cursor   int
	pendCls  int
	genders  [2][]int // 性別 glyph:男/女
}

func (tv *Tavern) open() { tv.active, tv.stage, tv.cursor = true, 0, 0 }

// input:方向移游標、A 選定推進、B 取消。回 (recruited *Member, closed)。
func (tv *Tavern) input(in InputState) (*Member, bool) {
	switch {
	case in.Cancel:
		if tv.stage == 1 {
			tv.stage, tv.cursor = 0, tv.pendCls // 退回職業
			return nil, false
		}
		tv.active = false
		return nil, true
	case in.Confirm:
		if tv.stage == 0 {
			tv.pendCls, tv.stage, tv.cursor = tv.cursor, 1, 0
			return nil, false
		}
		// 性別選定 → 建 lv1 隊員
		m := newMember(classNames[tv.pendCls], tv.pendCls, tv.cursor, 0)
		tv.active = false
		return m, true
	case in.DirEdge == 0: // 下
		tv.cursor = (tv.cursor + 1) % tv.optCount()
	case in.DirEdge == 1: // 上
		tv.cursor = (tv.cursor + tv.optCount() - 1) % tv.optCount()
	}
	return nil, false
}

func (tv *Tavern) optCount() int {
	if tv.stage == 0 {
		return 8
	}
	return 2
}

func (tv *Tavern) draw(rgba []byte, white dq3data.Color) {
	if !tv.active {
		return
	}
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	drawName := func(x, y int, gs []int, cur bool) {
		if cur {
			drawGlyph(rgba, tv.tx, x-20, y, curGlyph, yellow)
		}
		for i, g := range gs {
			drawGlyph(rgba, tv.tx, x+i*16, y, g, white)
		}
	}
	if tv.stage == 0 { // 8 職業
		for i := 0; i < 8; i++ {
			drawName(80, 56+i*22, classNames[i], i == tv.cursor)
		}
	} else { // 性別:男(272,674 借用)/女(234,674)—— 簡以職業已選 + 男/女 glyph
		labels := [2][]int{{272}, {234}} // 男 / 女(glyph 近似)
		drawName(80, 56, classNames[tv.pendCls], false)
		for i := 0; i < 2; i++ {
			drawName(80, 100+i*22, labels[i], i == tv.cursor)
		}
	}
}
