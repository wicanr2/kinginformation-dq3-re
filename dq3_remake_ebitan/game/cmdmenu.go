package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

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

// 每指令 2-glyph 標籤(對齊 C LABEL[]):對話/咒文/狀況/道具/裝備/調查。
var cmdLabels = [cmdCount][2]int{
	{443, 449}, {429, 430}, {665, 666}, {402, 237}, {197, 409}, {544, 545},
}

// CmdMenu 是命令窗狀態:cursor 0..5(列優先 r*2+c)。
type CmdMenu struct {
	tx     *dq3data.Text
	cursor int
	open   bool
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
	drawGlyph(rgba, m.tx, x+gpx, y, 274, fg)              // 命
	drawGlyph(rgba, m.tx, x+2*gpx, y, 664, fg)            // 令
	for i := 0; i < cmdCount; i++ {
		r, c := i>>1, i&1
		cx, cy := x+c*cw, y+rh+r*rh
		if i == m.cursor {
			drawGlyph(rgba, m.tx, cx, cy, 11, curfg) // glyph 11 = ►
		}
		drawGlyph(rgba, m.tx, cx+gpx, cy, cmdLabels[i][0], fg)
		drawGlyph(rgba, m.tx, cx+2*gpx, cy, cmdLabels[i][1], fg)
	}
}
