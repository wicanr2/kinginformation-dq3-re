package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

// 露易達酒館招募(移植 dq3_tavern + dq3_nameinput):職業 → 命名(NameInput,共用 widget)→
// 性別(GenderSelect,共用 widget)→ 建隊員。DQ3_PARTY_MAX=4(隊長+3)。
// 命名/性別 widget 與主角創建(newgame.go)共用同一份實作(見 nameinput.go),不重複兩份邏輯;
// 差異只在「命名完成時是否強制非空才放行」——酒館允許空名回退職業名(見 tavGender case),
// 主角創建則必填(newgame.go 的 ngName 分支)。
const (
	tavClass  = 0 // 選職業
	tavName   = 1 // 命名(NameInput:英數/注音格盤)
	tavGender = 2 // 選性別(GenderSelect)
)

type Tavern struct {
	tx        *dq3data.Text
	active    bool
	stage     int
	cursor    int // 職業選單游標(tavClass 專用)
	pendCls   int
	ni        NameInput
	gs        GenderSelect
	classHits hitList // 職業選單(tavClass)可點區塊,draw() 重建
	equipment [4]int  // game-pack registered_member 初始四槽；-1=空
}

func (tv *Tavern) open() {
	tv.active, tv.stage, tv.cursor, tv.pendCls = true, tavClass, 0, 0
	tv.ni = NameInput{}
	tv.gs = GenderSelect{}
}

// input:回 (recruited *Member, closed)。觸控直接點選(P2):tapIdx 命中對應階段的 hits(上一幀
// draw() 建的、對齊當前畫面的格子)→ 游標移過去 + 等同 A 確定,十字鍵 + A 保底並存。
func (tv *Tavern) input(in InputState, rs ...*rng.RNG) (*Member, bool) {
	// 正式 Game 路徑一定傳共享 RNG；無參數只保留給純 UI 單測。
	r := rng.New(0x1357)
	if len(rs) > 0 && rs[0] != nil {
		r = rs[0]
	}
	tapIdx := -1
	if in.Tapped {
		switch tv.stage {
		case tavClass:
			tapIdx = tv.classHits.at(in.TapX, in.TapY)
		case tavName:
			tapIdx = tv.ni.hits.at(in.TapX, in.TapY)
		case tavGender:
			tapIdx = tv.gs.hits.at(in.TapX, in.TapY)
		}
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
			tv.pendCls, tv.stage = tv.cursor, tavName
			tv.ni.Init()
		case in.DirEdge == 0:
			tv.cursor = (tv.cursor + 1) % 8
		case in.DirEdge == 1:
			tv.cursor = (tv.cursor + 7) % 8
		}
	case tavName:
		confirmed, canceled := tv.ni.input(in, tapIdx)
		switch {
		case canceled:
			tv.stage, tv.cursor = tavClass, tv.pendCls
		case confirmed: // 酒館允許空名(下方用職業名回退),完成即放行 → 性別
			tv.gs.Init()
			tv.stage = tavGender
		}
	case tavGender:
		gender, confirmed := tv.gs.input(in, tapIdx)
		switch {
		case in.Cancel:
			tv.stage, tv.ni.cursor = tavName, niCellOK
		case confirmed:
			name := tv.ni.nameBuf
			if len(name) == 0 { // 未命名 → 用職業名
				name = classNames[tv.pendCls]
			}
			m := newLevelOneMember(name, tv.pendCls, gender, r, tv.equipment)
			tv.active = false
			return m, true
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
	switch tv.stage {
	case tavClass:
		tv.classHits.reset()
		for i := 0; i < 8; i++ {
			y := 56 + i*22
			if i == tv.cursor {
				drawGlyph(rgba, tv.tx, 80-18, y, curGlyph, yellow)
			}
			for j, g := range classNames[i] {
				drawGlyph(rgba, tv.tx, 80+j*16, y, g, white)
			}
			tv.classHits.add(62, y-3, 200, 18, i)
		}
	case tavName:
		tv.ni.draw(rgba, tv.tx, white, yellow, 80, 56)
	case tavGender:
		for i, g := range classNames[tv.pendCls] { // 顯已選職業
			drawGlyph(rgba, tv.tx, 80+i*16, 56, g, white)
		}
		tv.gs.draw(rgba, tv.tx, white, yellow, 80, 100, 22)
	}
}
