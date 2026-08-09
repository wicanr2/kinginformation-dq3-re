package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// 野外命令窗(命令):2 欄 × 3 列指令格。1:1 移植 C dq3_cmdmenu.c。
// 標籤 glyph 取自 D3TXT00 rec400(官方字串,非猜)。
const (
	cmdTalk    = iota // 對話
	cmdSpell          // 咒文
	cmdStatus         // 狀況
	cmdItem           // 道具
	cmdEquip          // 裝備
	cmdExamine        // 調查
	cmdCount
)

// CmdMenu 是命令窗狀態:cursor 0..5(列優先 r*2+c)。
type CmdMenu struct {
	tx          *dq3data.Text
	cursor      int
	open        bool
	labels      [cmdCount][2]int
	title       [2]int
	labelsReady bool
	hits        hitList // 6 格可點區塊,draw() 重建(P2 直接點選)
}

// setLabels 安裝 versioned game pack 的玩家可見字模。直接 CmdMenu fixture
// 可以不安裝；production bootstrap 缺資料時會在 NewGameWithPack fail closed。
func (m *CmdMenu) setLabels(labels gamepack.FieldCommandLabels) {
	m.title = [2]int{labels.Title.PrimaryGlyph, labels.Title.SecondaryGlyph}
	m.labels = [cmdCount][2]int{
		{labels.Talk.PrimaryGlyph, labels.Talk.SecondaryGlyph},
		{labels.Spell.PrimaryGlyph, labels.Spell.SecondaryGlyph},
		{labels.Status.PrimaryGlyph, labels.Status.SecondaryGlyph},
		{labels.Item.PrimaryGlyph, labels.Item.SecondaryGlyph},
		{labels.Equip.PrimaryGlyph, labels.Equip.SecondaryGlyph},
		{labels.Examine.PrimaryGlyph, labels.Examine.SecondaryGlyph},
	}
	m.labelsReady = true
}

func (m *CmdMenu) Open() { m.cursor, m.open = 0, true }

// move:方向鍵移游標(2 欄 × 3 列繞回)。dir 用 facing 碼(0下 1上 2左 3右)。移植 dq3_cmdmenu_input 的方向部分。
func (m *CmdMenu) move(dir int) {
	r, c := m.cursor>>1, m.cursor&1
	switch dir {
	case 1: // 上
		r = (r + 2) % 3
	case 0: // 下
		r = (r + 1) % 3
	case 2, 3: // 左 / 右(2 欄繞回)
		c ^= 1
	}
	m.cursor = r*2 + c
}

// draw:在 (x,y) 繪命令窗(黑底白框 + 「命令」+ 6 格 + ► 游標)。移植 dq3_cmdmenu_render。
func (m *CmdMenu) draw(rgba []byte, fg, curfg dq3data.Color, x, y int) {
	if !m.open {
		return
	}
	const gpx = dq3data.GlyphPx
	const cw = 5 * gpx // 每欄寬(► + 2 字 + 間距)
	const rh = 18
	fillBox(rgba, x-gpx, y-gpx/2, 2*cw+gpx, rh*4+gpx, fg) // 視窗底 + 白框
	if m.labelsReady {
		drawGlyph(rgba, m.tx, x+gpx, y, m.title[0], fg)
		drawGlyph(rgba, m.tx, x+2*gpx, y, m.title[1], fg)
	}
	m.hits.reset()
	for i := 0; i < cmdCount; i++ {
		r, c := i>>1, i&1
		cx, cy := x+c*cw, y+rh+r*rh
		if i == m.cursor {
			drawGlyph(rgba, m.tx, cx, cy, 11, curfg) // glyph 11 = ►
		}
		if m.labelsReady {
			drawGlyph(rgba, m.tx, cx+gpx, cy, m.labels[i][0], fg)
			drawGlyph(rgba, m.tx, cx+2*gpx, cy, m.labels[i][1], fg)
		}
		m.hits.add(cx, cy, cw, rh, i)
	}
}
