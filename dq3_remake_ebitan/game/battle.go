package game

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/dq3_remake_ebitan/internal/battle"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 戰鬥場景(功能核心版):怪物 sprite + 敵/我 HP + 戰う/逃げる 指令 + 回合結算(用 internal/battle 公式)。
// 移植 dq3_battlescene 的核心迴圈;咒文/道具/防禦/AI 預測/升級為後續增量。
const (
	bcWar  = 0 // 戰う
	bcDef  = 1 // 防御
	bcItem = 2 // 道具(藥草)
	bcFlee = 3 // 逃げる
	bcN    = 4
)

// 戰鬥指令 glyph 標籤(對齊 C dq3_battlescene CMD_*)。
var battleCmdLabels = [bcN][2]int{{107, 207}, {203, 204}, {402, 1354}, {629, 630}} // 戰/防/道具/逃

const herbCode = 0x41   // 藥草 item id
const herbHeal = 30     // DQ3_HERB_HEAL

type battlePhase int

const (
	phCommand battlePhase = iota // 等玩家下指令
	phMessage                    // 顯示訊息(A 推進)
	phEnd                        // 勝/敗/逃(A 關閉戰鬥)
)

// Battle 是一場戰鬥的狀態。
type Battle struct {
	mons *dq3data.Monsters
	shp  []byte
	mpal []dq3data.Color // MNSBK.PAL(怪物色盤)
	tx   *dq3data.Text

	active   bool
	monID    int
	spr      *dq3data.MonsterSprite
	enemyHP  int
	heroHP   int
	heroMax  int
	heroAtk  int
	heroDef  int
	heroAgi  int
	cursor    int
	msg       string
	phase     battlePhase
	result    int // 0 進行中、1 勝、2 敗、3 逃
	gotExp    int
	gotGold   int
	rng       *rand.Rand
	flashCol  int // >0:受擊閃光殘餘幀
	defending bool
	heroHerbs int // 開戰時持有藥草數
	usedHerbs int // 本戰用掉藥草數(戰後從背包扣)
}

func (b *Battle) roll() int { return b.rng.Intn(256) }

// heroParams 是開戰時由 Game 傳入的主角當前數值(等級推導,見 game.heroStats)。
type heroParams struct {
	level, curHP, maxHP, atk, def, agi int
	herbs                              int // 持有藥草數(戰鬥中可用)
}

// start 開一場戰鬥(monID + 主角數值)。跳過空 sprite 的怪。
func (b *Battle) start(monID int, seed int64, hp heroParams) bool {
	spr, err := dq3data.DecodeMonsterSprite(b.shp, monID)
	if err != nil {
		return false
	}
	st, ok := b.mons.Stat(monID)
	if !ok {
		return false
	}
	b.rng = rand.New(rand.NewSource(seed))
	b.monID, b.spr = monID, spr
	b.enemyHP = int(st.HPBase) + b.rng.Intn(int(st.HPRand)+1)
	b.heroMax, b.heroHP = hp.maxHP, hp.curHP
	b.heroAtk, b.heroDef, b.heroAgi = hp.atk, hp.def, hp.agi
	b.heroHerbs, b.usedHerbs, b.defending = hp.herbs, 0, false
	b.cursor, b.phase, b.result = 0, phCommand, 0
	b.msg, b.gotExp, b.gotGold = "", 0, 0
	b.active = true
	return true
}

// input 處理一幀輸入,回傳戰鬥是否結束(關閉)。
func (b *Battle) input(in InputState) (closed bool) {
	switch b.phase {
	case phCommand:
		switch {
		case in.Confirm:
			b.execTurn()
		case in.DirEdge == 0: // 下
			b.cursor = (b.cursor + 1) % bcN
		case in.DirEdge == 1: // 上
			b.cursor = (b.cursor + bcN - 1) % bcN
		}
	case phMessage:
		if in.Confirm {
			b.phase = phCommand
			b.msg = ""
			if b.result != 0 {
				b.phase = phEnd
			}
		}
	case phEnd:
		if in.Confirm {
			b.active = false
			return true
		}
	}
	return false
}

// execTurn 執行玩家指令 + 敵方回合,結算。
func (b *Battle) execTurn() {
	st, _ := b.mons.Stat(b.monID)
	b.defending = false
	switch b.cursor {
	case bcFlee: // 逃げる
		if battle.FleeOK(b.heroAgi, int(st.FleeResist), b.roll()) {
			b.msg, b.result, b.phase = "逃走成功", 3, phMessage
			return
		}
		b.msg = "逃走失敗"
	case bcDef: // 防御(本回合受傷減半)
		b.defending = true
		b.msg = "防御"
	case bcItem: // 道具:用藥草回 HP
		if b.usedHerbs < b.heroHerbs && b.heroHP < b.heroMax {
			heal := herbHeal
			if b.heroHP+heal > b.heroMax {
				heal = b.heroMax - b.heroHP
			}
			b.heroHP += heal
			b.usedHerbs++
			b.msg = fmt.Sprintf("回復 %d", heal)
		} else {
			b.msg = "沒有藥草"
		}
	default: // 戰う:主角攻擊
		crit := 0
		if b.roll() < 8 { // ~1/32 會心
			crit = 1
		}
		dmg := battle.PhysDamage(b.heroAtk, int(st.Def), b.roll(), crit)
		b.enemyHP -= dmg
		if b.enemyHP < 0 {
			b.enemyHP = 0
		}
		b.msg = fmt.Sprintf("給予 %d 傷害", dmg)
		b.flashCol = 4
		if b.enemyHP == 0 { // 勝
			b.gotExp, b.gotGold = int(st.Exp), int(st.Gold)
			b.msg, b.result, b.phase = fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold), 1, phMessage
			return
		}
	}
	// 敵方反擊(防御則減半)
	edmg := battle.PhysDamage(int(st.Atk), b.heroDef, b.roll(), 0)
	if b.defending {
		edmg /= 2
	}
	b.heroHP -= edmg
	if b.heroHP <= 0 {
		b.heroHP = 0
		b.msg, b.result, b.phase = "全滅…", 2, phMessage
		return
	}
	b.phase = phMessage
}

// draw 畫整個戰鬥畫面到 rgba。scenePal = 場景色盤(文字/框);怪物用 b.mpal。
func (b *Battle) draw(rgba []byte, scenePal []dq3data.Color) {
	white := dq3data.Color{R: 248, G: 248, B: 248}
	if len(scenePal) > 15 {
		white = scenePal[15]
	}
	red := dq3data.Color{R: 224, G: 48, B: 48}
	green := dq3data.Color{R: 40, G: 120, B: 40}
	sky := dq3data.Color{R: 32, G: 40, B: 96}

	// 背景:上天空、下綠地(地平線 y~200)
	for y := 0; y < ScreenH; y++ {
		c := sky
		if y >= 200 {
			c = green
		}
		for x := 0; x < ScreenW; x++ {
			putPx(rgba, x, y, c)
		}
	}
	// 怪物 sprite(置中,站綠地上;受擊閃紅)
	if b.spr != nil && b.result != 1 { // 打倒後不畫
		ox := (ScreenW - b.spr.W) / 2
		oy := 200 - b.spr.H
		flash := b.flashCol > 0
		for r := 0; r < b.spr.H; r++ {
			for x := 0; x < b.spr.W; x++ {
				if !b.spr.Opaque[r][x] {
					continue
				}
				idx := int(b.spr.Px[r][x])
				c := dq3data.Color{R: 200, G: 200, B: 200}
				if idx < len(b.mpal) {
					c = b.mpal[idx]
				}
				if flash {
					c = red
				}
				putPx(rgba, ox+x, oy+r, c)
			}
		}
		if b.flashCol > 0 {
			b.flashCol--
		}
		// 敵 HP:H + 數字(怪上方)
		drawGlyph(rgba, b.tx, ScreenW/2-24, oy-20, 22, white) // glyph 22 = 'H'
		drawNumber(rgba, b.tx, ScreenW/2-4, oy-20, b.enemyHP, white)
	}
	// 下方框:左指令(戰う/逃げる)、右我方 HP
	fillBox(rgba, 16, 244, ScreenW-32, 96, white)
	if b.phase == phCommand {
		for i := 0; i < bcN; i++ {
			y := 258 + i*22
			if i == b.cursor {
				drawGlyph(rgba, b.tx, 32, y, 11, dq3data.Color{R: 255, G: 224, B: 32}) // ► 游標
			}
			drawGlyph(rgba, b.tx, 52, y, battleCmdLabels[i][0], white)
			drawGlyph(rgba, b.tx, 68, y, battleCmdLabels[i][1], white)
		}
	}
	// 我方 HP(右下):H<現>（滿血白、瀕死紅）
	hpCol := white
	if b.heroHP*4 < b.heroMax {
		hpCol = red
	}
	drawGlyph(rgba, b.tx, ScreenW-140, 258, 22, white) // H
	drawNumber(rgba, b.tx, ScreenW-118, 258, b.heroHP, hpCol)
	// 訊息(ASCII/數字用 drawText;中文訊息用 glyph 需另查,這裡以數字+H 為主,文字訊息暫存 b.msg 供 log)
	if b.msg != "" {
		drawASCII(rgba, b.tx, 32, 300, b.msg, white)
	}
}

// drawNumber 畫十進位數(digit glyph 0..9)。
func drawNumber(rgba []byte, tx *dq3data.Text, x, y, val int, fg dq3data.Color) {
	if val < 0 {
		val = 0
	}
	s := fmt.Sprintf("%d", val)
	for i := 0; i < len(s); i++ {
		drawGlyph(rgba, tx, x+i*dq3data.GlyphPx, y, int(s[i]-'0'), fg)
	}
}

// drawASCII 畫數字部分(非數字字元跳過;中文訊息 glyph 對照另補)。
func drawASCII(rgba []byte, tx *dq3data.Text, x, y int, s string, fg dq3data.Color) {
	col := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			drawGlyph(rgba, tx, x+col*dq3data.GlyphPx, y, int(c-'0'), fg)
			col++
		} else if c == ' ' {
			col++
		}
	}
}
