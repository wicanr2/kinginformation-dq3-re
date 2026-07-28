package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// 酒場「找同伴參加 / 與同伴分離 / 觀看名單」(1F,對照 dq3_remake/src/dq3_tavern.c 的
// DQ3_TAV_ROSTER 畫面 + dq3_roster.c 的 dq3_party_add/remove)。對齊原版兩段式
// (docs/36 rec527-550):
//   - 2F 冒險者登錄所(Tavern,tavern.go)創角只登錄到 roster,不自動入隊。
//   - 1F 酒場(本檔)才把 roster 內角色拉進/移出 g.companions(隊伍上限 rcPartyMax=主角+3)。
//
// 注意:C 版 dq3_tavern.c 的 DQ3_TAV_GENDER 分支建完角色後立刻呼叫 dq3_party_add
// 自動入隊——那是 C demo 的簡化,**不照抄**;本檔刻意把「登錄」與「入隊」分成兩個獨立動作。
//
// 主選單三項直接取 D3TXT00 rec529 解出的原版 glyph，不再使用 0/1/2 佔位。
const (
	rcMenu  = 0 // 主選單:0 找同伴參加(→rcJoin)/1 與同伴分離(→rcLeave)/2 觀看名單(→rcView)
	rcJoin  = 1 // 從 roster(未入隊)選一名 → 入隊(companions,滿則擋掉,不頂替)
	rcLeave = 2 // 從 companions 選一名 → 移回 roster
	rcView  = 3 // 觀看名單(roster+companions 唯讀,任意鍵/點擊回主選單)
)

const (
	rcRosterMax = 11 // 原版 sub_1ce4 掃角色槽 1..0x0b（主角固定槽 0）
	rcPartyMax  = 4  // 隊伍:主角 + 3 同伴(對齊 C DQ3_PARTY_MAX;companions 只存 3 名同伴)
)

// Recruit 只持有 UI 狀態(游標/hits);實際 roster/companions 搬移在 Game 方法
// (recruitInput/drawRecruit,本檔)操作 g.roster / g.companions,理由同 panel.go 的
// drawStatus/equipSelected 模式——選單需要同時讀寫兩份 Game 層清單,不適合像 Tavern
// 那樣做成不持有 *Game 的純 UI struct。
type Recruit struct {
	tx       *dq3data.Text
	active   bool
	stage    int
	cursor   int
	menuHits hitList
	listHits hitList
}

func (rc *Recruit) open() { rc.active, rc.stage, rc.cursor = true, rcMenu, 0 }

// tavernCreate:酒館 2F 登錄所 modal 的輸入 glue(掛在 g.tavern 上)。建角**只登錄 roster,
// 不自動入隊**(見檔頭說明)。抽成獨立方法,方便單元測試不必經過完整 Update()/ebiten 輸入輪詢。
func (g *Game) tavernCreate(in InputState) {
	if m, _ := g.tavern.input(in, &g.prng); m != nil &&
		len(g.roster)+len(g.companions) < rcRosterMax {
		g.roster = append(g.roster, m)
	}
}

// recruitInput:酒場招募 modal 的輸入處理(掛在 g.recruit 上)。
func (g *Game) recruitInput(in InputState) {
	rc := &g.recruit
	tapIdx := -1
	if in.Tapped {
		if rc.stage == rcMenu {
			tapIdx = rc.menuHits.at(in.TapX, in.TapY)
		} else {
			tapIdx = rc.listHits.at(in.TapX, in.TapY)
		}
	}
	switch rc.stage {
	case rcMenu:
		confirm := in.Confirm
		if tapIdx >= 0 {
			rc.cursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			rc.active = false
		case confirm:
			switch rc.cursor {
			case 0:
				rc.stage, rc.cursor = rcJoin, 0
			case 1:
				rc.stage, rc.cursor = rcLeave, 0
			case 2:
				rc.stage, rc.cursor = rcView, 0
			}
		case in.DirEdge == 0:
			rc.cursor = (rc.cursor + 1) % 3
		case in.DirEdge == 1:
			rc.cursor = (rc.cursor + 2) % 3
		}
	case rcJoin:
		g.recruitPick(in, tapIdx, len(g.roster), func(i int) {
			if len(g.companions) >= rcPartyMax-1 { // 隊伍已滿(主角+3)→ 擋掉,不頂替、不丟隊友
				return
			}
			m := g.roster[i]
			g.roster = append(g.roster[:i], g.roster[i+1:]...)
			g.companions = append(g.companions, m)
		})
	case rcLeave:
		g.recruitPick(in, tapIdx, len(g.companions), func(i int) {
			m := g.companions[i]
			g.companions = append(g.companions[:i], g.companions[i+1:]...)
			g.roster = append(g.roster, m)
		})
	case rcView:
		if in.Cancel || in.Confirm || tapIdx >= 0 {
			rc.stage, rc.cursor = rcMenu, 0
		}
	}
}

// recruitPick:rcJoin/rcLeave 共用的清單導覽 + 確定邏輯。n=清單目前長度,onConfirm(i) 執行實際
// 搬移(呼叫端自行決定滿員擋不擋)。Cancel 回主選單;清單空時方向鍵/確定不動作(避免除以 0)。
func (g *Game) recruitPick(in InputState, tapIdx, n int, onConfirm func(i int)) {
	rc := &g.recruit
	confirm := in.Confirm
	if tapIdx >= 0 {
		rc.cursor, confirm = tapIdx, true
	}
	switch {
	case in.Cancel:
		rc.stage, rc.cursor = rcMenu, 0
	case confirm && n > 0:
		if rc.cursor < 0 || rc.cursor >= n {
			rc.cursor = 0
		}
		onConfirm(rc.cursor)
		if rc.cursor >= n-1 { // 搬移後清單少 1,游標夾回合法範圍
			rc.cursor = max0(n - 2)
		}
	case in.DirEdge == 0 && n > 0:
		rc.cursor = (rc.cursor + 1) % n
	case in.DirEdge == 1 && n > 0:
		rc.cursor = (rc.cursor + n - 1) % n
	}
}

// drawRecruit:酒場招募 modal 繪製,依 stage 分派。
func (g *Game) drawRecruit(rgba []byte, white dq3data.Color) {
	rc := &g.recruit
	if !rc.active {
		return
	}
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	switch rc.stage {
	case rcMenu:
		rc.menuHits.reset()
		labels := [3][]int{
			{769, 601, 602, 770, 368}, // 找同伴參加
			{762, 601, 602, 764, 502}, // 與同伴分離
			{771, 668, 692, 772},      // 觀看名單
		}
		for i := 0; i < 3; i++ {
			y := 56 + i*22
			if i == rc.cursor {
				drawGlyph(rgba, rc.tx, 80-18, y, curGlyph, yellow)
			}
			for j, gi := range labels[i] {
				drawGlyph(rgba, rc.tx, 80+j*dq3data.GlyphPx, y, gi, white)
			}
			rc.menuHits.add(62, y-3, 200, 18, i)
		}
	case rcJoin:
		g.drawRecruitList(rgba, white, yellow, g.roster)
	case rcLeave:
		g.drawRecruitList(rgba, white, yellow, g.companions)
	case rcView:
		all := append(append([]*Member{}, g.roster...), g.companions...)
		g.drawRecruitList(rgba, white, yellow, all)
	}
}

// drawRecruitList:名冊/隊伍清單共用繪製 —— 姓名 glyph + 職業名(classNames)+ 等級數字,
// 游標行反白(對照 C render_roster 的欄位順序)。
func (g *Game) drawRecruitList(rgba []byte, white, yellow dq3data.Color, list []*Member) {
	rc := &g.recruit
	rc.listHits.reset()
	const gp = dq3data.GlyphPx
	for i, m := range list {
		if i >= 12 { // 畫面容量上限(320px 高扣邊框約 12 列)
			break
		}
		y := 56 + i*18
		col := white
		if i == rc.cursor {
			drawGlyph(rgba, rc.tx, 44, y, curGlyph, yellow)
			col = yellow
		}
		cx := 64
		for _, gi := range m.Name {
			drawGlyph(rgba, rc.tx, cx, y, gi, col)
			cx += gp
		}
		cx += 8
		for _, gi := range classNames[m.Class] {
			drawGlyph(rgba, rc.tx, cx, y, gi, col)
			cx += gp
		}
		cx += 8
		drawNumber(rgba, rc.tx, cx, y, m.Level(), col)
		rc.listHits.add(44, y-3, ScreenW-80-8, 18, i)
	}
}
