package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// 招募入口依 docs/188 的有限 READY 播放問候並顯示資料包選單。
// 登錄只加入 roster；既有加入／分離狀態機搬移 roster 與 companions。
// 入隊後文字、分離與查看的原版對拍尚未完成，不以舊 C 實作作 oracle。
const (
	rcMenu  = 0 // 主選單:0 找同伴參加(→rcJoin)/1 與同伴分離(→rcLeave)/2 觀看名單(→rcView)
	rcJoin  = 1 // 從 roster(未入隊)選一名 → 入隊(companions,滿則擋掉,不頂替)
	rcLeave = 2 // 從 companions 選一名 → 移回 roster
	rcView  = 3 // 觀看未入隊名冊；清單與取消依 docs/188 有限 READY。
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
	join                       *gamepack.RecruitmentJoin
	backdrop                   []byte
	joinedName, leaderName     []int
	musicFrames                int
	musicStream                []byte
	contract                   *gamepack.RecruitmentEntry
	selection                  *gamepack.RecruitmentSelection
	texts                      map[string][]uint16
	dialogue                   Dialogue
	greetingIndex, cursorGlyph int
	raster                     *indexedNewGameRenderer
	viewFlow                   *NewGameFlow
	viewSpellNames             []string
	renameFlow                 *NewGameFlow
	tx                         *dq3data.Text
	active                     bool
	stage                      int
	afterText                  int
	rosterCapacity             int
	cursor                     int
	menuHits                   hitList
	listHits                   hitList
}

func (rc *Recruit) open() {
	rc.reset()
	if rc.contract == nil {
		return
	}
	rc.active, rc.stage, rc.cursor = true, rcGreeting, 0
	rc.greetingIndex = 0
	rc.startGreeting()
}

// tavernCreate:酒館 2F 登錄所 modal 的輸入 glue(掛在 g.tavern 上)。建角**只登錄 roster,
// 不自動入隊**(見檔頭說明)。抽成獨立方法,方便單元測試不必經過完整 Update()/ebiten 輸入輪詢。
func (g *Game) tavernCreate(in InputState) {
	g.tavern.occupied = len(g.roster) + len(g.companions)
	if m, _ := g.tavern.input(in, &g.prng); m != nil && g.tavern.contract != nil &&
		len(g.roster)+len(g.companions) < g.tavern.contract.RosterCapacity {
		g.roster = append(g.roster, m)
	}
}

// recruitInput:酒場招募 modal 的輸入處理(掛在 g.recruit 上)。
func (g *Game) recruitInput(in InputState) {
	rc := &g.recruit
	if rc.stage == rcViewClose && in.Rename {
		g.startRecruitmentRename()
		return
	}
	if rc.stage == rcViewRename {
		g.recruitmentRenameInput(in)
		return
	}
	if rc.stage == rcViewAbility || rc.stage == rcViewClose || rc.stage == rcViewSpells {
		rc.viewInput(in)
		return
	}
	if rc.stage >= rcJoinedText && rc.stage <= rcMusicWait {
		g.recruitJoinInput(in)
		return
	}
	if rc.stage == rcGreeting {
		rc.greetingInput(in)
		return
	}
	if rc.stage == rcText || rc.stage == rcAgain || rc.stage == rcFinalWait {
		rc.selectionInput(in)
		return
	}
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
		confirm := in.Confirm || in.Enter
		if tapIdx >= 0 {
			rc.cursor, confirm = tapIdx, true
		}
		switch {
		case in.Cancel:
			rc.active = false
		case confirm:
			if rc.contract == nil || rc.cursor < 0 || rc.cursor >= len(rc.contract.OptionActions) {
				return
			}
			switch rc.contract.OptionActions[rc.cursor] {
			case gamepack.RecruitJoin:
				if rc.selection == nil {
					return
				}
				next := rcJoin
				if len(g.roster) == 0 {
					next = rcEmptyJoinHint
				}
				rc.startSelectionText(rc.selection.PromptTextID, next)
			case gamepack.RecruitLeave:
				if len(g.companions) == 0 && rc.selection != nil {
					rc.startSelectionText(rc.selection.EmptyLeaveTextID, rcEmptySelectionReturn)
					rc.dialogue.heroName = append([]int(nil), g.heroName...)
					return
				}
				rc.stage, rc.cursor = rcLeave, 0
			case gamepack.RecruitView:
				if len(g.roster) == 0 && rc.selection != nil {
					rc.startSelectionText(rc.selection.EmptyViewTextID, rcEmptySelectionReturn)
					return
				}
				rc.stage, rc.cursor = rcView, 0
			}
		case in.DirEdge == 0 || in.DirEdge == 3:
			if rc.contract != nil {
				rc.cursor = (rc.cursor + 1) % len(rc.contract.OptionActions)
			}
		case in.DirEdge == 1 || in.DirEdge == 2:
			if rc.contract != nil {
				n := len(rc.contract.OptionActions)
				rc.cursor = (rc.cursor + n - 1) % n
			}
		}
	case rcJoin:
		if in.Cancel && rc.selection != nil {
			rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
			return
		}
		g.recruitPick(in, tapIdx, len(g.roster), func(i int) {
			if rc.join == nil || len(g.companions) >= rcPartyMax-1 { // 未有契約或已滿時保持名冊。
				return
			}
			m := g.roster[i]
			g.roster = append(g.roster[:i], g.roster[i+1:]...)
			g.companions = append(g.companions, m)
			g.resetPartyTrail()
			g.startRecruitmentJoin(m.Name)
		})
	case rcLeave:
		g.recruitPick(in, tapIdx, len(g.companions), func(i int) {
			m := g.companions[i]
			g.companions = append(g.companions[:i], g.companions[i+1:]...)
			g.roster = append(g.roster, m)
			g.resetPartyTrail()
		})
	case rcView:
		if in.Cancel && rc.selection != nil {
			rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
			return
		}
		g.recruitPick(in, tapIdx, len(g.roster), func(i int) {
			g.startRecruitmentView(g.roster[i])
		})
	}
}

// recruitPick:rcJoin/rcLeave 共用的清單導覽 + 確定邏輯。n=清單目前長度,onConfirm(i) 執行實際
// 搬移(呼叫端自行決定滿員擋不擋)。Cancel 回主選單;清單空時方向鍵/確定不動作(避免除以 0)。
func (g *Game) recruitPick(in InputState, tapIdx, n int, onConfirm func(i int)) {
	rc := &g.recruit
	confirm := in.Confirm || in.Enter
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
		stage := rc.stage
		onConfirm(rc.cursor)
		if stage != rcView && rc.cursor >= n-1 { // 搬移後清單少 1,游標夾回合法範圍
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
	if rc.join != nil && rc.join.RetainCallerBackdrop {
		if len(rc.backdrop) == 0 {
			rc.backdrop = append([]byte(nil), rgba...)
		} else {
			copy(rgba, rc.backdrop)
		}
	}
	if rc.stage == rcGreeting || rc.stage == rcMenu || rc.stage == rcText || rc.stage == rcAgain || rc.stage == rcFinalWait || rc.stage >= rcJoinedText && rc.stage <= rcMusicWait {
		rc.drawEntry(rgba, white)
		return
	}
	if rc.stage == rcJoin || rc.stage == rcView {
		g.drawRecruitSelection(rgba, white)
		return
	}
	if rc.stage == rcViewAbility || rc.stage == rcViewClose || rc.stage == rcViewRename || rc.stage == rcViewSpells {
		rc.drawView(rgba, white)
		return
	}
	fillBox(rgba, 40, 40, ScreenW-80, ScreenH-120, white)
	yellow := dq3data.Color{R: 255, G: 224, B: 32}
	switch rc.stage {
	case rcJoin:
		g.drawRecruitList(rgba, white, yellow, g.roster)
	case rcLeave:
		g.drawRecruitList(rgba, white, yellow, g.companions)
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
