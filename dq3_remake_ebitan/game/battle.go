package game

import (
	"fmt"
	"math/rand"

	"github.com/wicanr2/dq3_remake_ebitan/internal/battle"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
)

// 戰鬥場景(功能核心版):怪物 sprite + 敵/我 HP + 戰う/逃げる 指令 + 回合結算(用 internal/battle 公式)。
// 移植 dq3_battlescene 的核心迴圈;咒文/道具/防禦/AI 預測/升級為後續增量。
// 指令順序對齊 C dq3_battlescene col1[]:戰鬥/逃跑/防禦/道具/咒文。
const (
	bcWar   = 0 // 戰う
	bcFlee  = 1 // 逃げる
	bcDef   = 2 // 防御
	bcItem  = 3 // 道具(藥草)
	bcSpell = 4 // 咒文
	bcN     = 5
)

// 戰鬥指令 glyph 標籤(對齊 C CMD_WAR/FLEE/DEF/ITEM/SPELL)。
var battleCmdLabels = [bcN][2]int{{107, 207}, {629, 630}, {203, 204}, {402, 1354}, {429, 430}}

const herbCode = 0x41   // 藥草 item id
const herbHeal = 30     // DQ3_HERB_HEAL

type battlePhase int

const (
	phCommand battlePhase = iota // 等玩家下指令
	phSpell                      // 選咒文
	phMessage                    // 顯示訊息(A 推進)
	phEnd                        // 勝/敗/逃(A 關閉戰鬥)
)

// Battle 是一場戰鬥的狀態。
type Battle struct {
	mons *dq3data.Monsters
	shp  []byte
	mpal     []dq3data.Color // MNSBK.PAL(怪物色盤)
	tx       *dq3data.Text
	nameText *dq3data.Text // D3TXT00 名表(咒文名 rec)

	active   bool
	monID    int
	spr      *dq3data.MonsterSprite
	enemyHP  int
	enemyMax int
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
	heroHerbs   int   // 開戰時持有藥草數
	usedHerbs   int   // 本戰用掉藥草數(戰後從背包扣)
	heroLevel   int   // 我方等級(敵 AI 逃跑門檻用)
	heroMP      int   // 目前 MP(施咒消耗)
	heroMaxMP   int   //
	spells      []int // 已學可施放咒文 rec
	spellCursor int
}

func (b *Battle) roll() int { return b.rng.Intn(256) }

// heroParams 是開戰時由 Game 傳入的主角當前數值(等級推導,見 game.heroStats)。
type heroParams struct {
	level, curHP, maxHP, atk, def, agi int
	herbs                              int   // 持有藥草數(戰鬥中可用)
	mp, maxMP                          int   // 目前/最大 MP
	spells                             []int // 已學可施放咒文 rec
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
	b.enemyMax = b.enemyHP
	b.heroMax, b.heroHP = hp.maxHP, hp.curHP
	b.heroAtk, b.heroDef, b.heroAgi, b.heroLevel = hp.atk, hp.def, hp.agi, hp.level
	b.heroHerbs, b.usedHerbs, b.defending = hp.herbs, 0, false
	b.heroMP, b.heroMaxMP, b.spells = hp.mp, hp.maxMP, hp.spells
	b.spellCursor = 0
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
			if b.cursor == bcSpell { // 咒文 → 選咒子選單
				if len(b.spells) > 0 {
					b.spellCursor, b.phase = 0, phSpell
				} else {
					b.msg, b.phase = "還沒學會咒文", phMessage
				}
			} else {
				b.execTurn()
			}
		case in.DirEdge == 0: // 下
			b.cursor = (b.cursor + 1) % bcN
		case in.DirEdge == 1: // 上
			b.cursor = (b.cursor + bcN - 1) % bcN
		}
	case phSpell:
		switch {
		case in.Cancel:
			b.phase = phCommand
		case in.Confirm:
			b.execSpell(b.spells[b.spellCursor])
		case in.DirEdge == 0:
			b.spellCursor = (b.spellCursor + 1) % len(b.spells)
		case in.DirEdge == 1:
			b.spellCursor = (b.spellCursor + len(b.spells) - 1) % len(b.spells)
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
	b.enemyTurn()
}

// execSpell 施放咒文 rec(DMG 傷敵 / HEAL 補己),消耗 MP,再敵方回合。
func (b *Battle) execSpell(rec int) {
	def, ok := spell.GetDef(rec)
	if !ok {
		b.phase = phCommand
		return
	}
	if b.heroMP < def.MP {
		b.msg, b.phase = "MP 不足", phMessage
		return
	}
	b.heroMP -= def.MP
	b.defending = false
	val := spell.CastValue(def.Base, b.roll())
	if def.Kind == spell.Heal {
		if b.heroHP+val > b.heroMax {
			val = b.heroMax - b.heroHP
		}
		b.heroHP += val
		b.msg = fmt.Sprintf("回復 %d", val)
	} else { // Dmg
		b.enemyHP -= val
		if b.enemyHP < 0 {
			b.enemyHP = 0
		}
		b.flashCol = 4
		b.msg = fmt.Sprintf("咒文 %d 傷害", val)
		if b.enemyHP == 0 {
			st, _ := b.mons.Stat(b.monID)
			b.gotExp, b.gotGold = int(st.Exp), int(st.Gold)
			b.msg, b.result, b.phase = fmt.Sprintf("打倒! EXP%d G%d", b.gotExp, b.gotGold), 1, phMessage
			return
		}
	}
	b.enemyTurn()
}

// enemyTurn:敵方回合(AI 逃跑判定 → 否則物理反擊,防御減半)。玩家行動後呼叫。
func (b *Battle) enemyTurn() {
	st, _ := b.mons.Stat(b.monID)
	if ai, ok := b.mons.AI(b.monID); ok && ai.FleeRate > 0 &&
		b.heroLevel >= int(ai.FleeThresh) && b.roll() <= int(ai.FleeRate) {
		b.msg, b.result, b.phase = "敵人逃走了", 3, phMessage
		return
	}
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

// 版面常數(1:1 對齊 C dq3_battlescene render,references/game3.png)。
const (
	fieldY0, fieldY1, groundY = 80, 246, 232 // 場景帶(天空+綠地)/ 綠地平線
	curGlyph                  = 0x77         // ►/★ 游標 glyph(對齊 C)
)

// drawName:畫 D3TXT00 記錄 rec 的 glyph 序列(咒名/敵名;非 code+1)。
func (b *Battle) drawName(rgba []byte, x, y, rec int, fg dq3data.Color) {
	if b.nameText == nil {
		return
	}
	col := 0
	for _, v := range b.nameText.Record(rec) {
		if v < dq3data.GlyphMax {
			drawGlyph(rgba, b.nameText, x+col*dq3data.GlyphPx, y, int(v), fg)
			col++
		}
	}
}

// draw 畫整個戰鬥畫面到 rgba,1:1 對齊 C render():上狀態列 / 中怪群站綠地 / 下左指令 + 下右敵名。
func (b *Battle) draw(rgba []byte, scenePal []dq3data.Color) {
	white := dq3data.Color{R: 248, G: 248, B: 248}
	if len(scenePal) > 15 {
		white = scenePal[15]
	}
	red := dq3data.Color{R: 224, G: 48, B: 48}
	cyan := dq3data.Color{R: 64, G: 208, B: 208}
	black := dq3data.Color{}
	sky := dq3data.Color{R: 56, G: 120, B: 216}
	ground := dq3data.Color{R: 40, G: 128, B: 40}

	// 背景:全黑 → 場景帶(天空 y80..232 / 綠地 232..246);其餘黑(對齊 C FIELD_Y0/Y1/GROUND_Y)
	for y := 0; y < ScreenH; y++ {
		c := black
		if y >= fieldY0 && y < fieldY1 {
			if y < groundY {
				c = sky
			} else {
				c = ground
			}
		}
		for x := 0; x < ScreenW; x++ {
			putPx(rgba, x, y, c)
		}
	}
	// 怪物:站綠地平線(gy=groundY+8-h,置中);受擊閃紅;死不畫
	if b.spr != nil && b.enemyHP > 0 {
		gx := ScreenW/2 - b.spr.W/2
		gy := groundY + 8 - b.spr.H
		flash := b.flashCol > 0
		for r := 0; r < b.spr.H; r++ {
			for x := 0; x < b.spr.W; x++ {
				if !b.spr.Opaque[r][x] {
					continue
				}
				c := dq3data.Color{R: 200, G: 200, B: 200}
				if idx := int(b.spr.Px[r][x]); idx < len(b.mpal) {
					c = b.mpal[idx]
				}
				if flash {
					c = red
				}
				putPx(rgba, gx+x, gy+r, c)
			}
		}
		if b.flashCol > 0 {
			b.flashCol--
		}
		// 敵 HP(remake 增強):H + 數字,置中於怪上方
		cx := gx + b.spr.W/2
		hy := gy - 32
		if hy < fieldY0 {
			hy = fieldY0
		}
		hc := white
		if b.enemyMax > 0 && b.enemyHP*4 < b.enemyMax {
			hc = red
		}
		drawGlyph(rgba, b.tx, cx-26, hy, 22, white) // H
		drawNumber(rgba, b.tx, cx-8, hy, b.enemyHP, hc)
	}
	// 上方隊伍狀態列(X=0x13*8=152、Y=8;單勇者一欄:H+HP / M+MP / 等級)
	{
		tx0, ty0 := 0x13*8, 8
		fillBox(rgba, tx0-8, ty0-6, 88, 4*16+12, white)
		hc := white
		if b.heroHP <= 0 {
			hc = red
		}
		drawGlyph(rgba, b.tx, tx0, ty0+16, 22, hc) // H
		drawNumber(rgba, b.tx, tx0+18, ty0+16, b.heroHP, hc)
		drawGlyph(rgba, b.tx, tx0, ty0+32, 27, white) // M
		drawNumber(rgba, b.tx, tx0+18, ty0+32, b.heroMP, white)
		drawNumber(rgba, b.tx, tx0+18, ty0+48, b.heroLevel, white) // 等級
	}
	// 下方左:指令窗(120,236,150×108)—— 指令 或 咒文子選單
	{
		mx, my, mw, mh := 120, 236, 150, 108
		fillBox(rgba, mx, my, mw, mh, white)
		if b.phase == phSpell {
			for i, rec := range b.spells {
				if i >= 5 {
					break
				}
				y := my + 24 + i*16
				if i == b.spellCursor {
					drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				}
				b.drawName(rgba, mx+28, y, rec, white)
				if def, ok := spell.GetDef(rec); ok {
					drawNumber(rgba, b.tx, mx+110, y, def.MP, cyan)
				}
			}
		} else {
			for i := 0; i < bcN; i++ {
				y := my + 24 + i*16
				if i == b.cursor {
					drawGlyph(rgba, b.tx, mx+8, y, curGlyph, white)
				}
				drawGlyph(rgba, b.tx, mx+28, y, battleCmdLabels[i][0], white)
				drawGlyph(rgba, b.tx, mx+72, y, battleCmdLabels[i][1], white)
			}
		}
	}
	// 下方右:敵名框(290,262,250×36)—— 敵名 = D3TXT00 record 0x258+monID + 數量
	{
		ex, ey, ew, eh := 290, 262, 250, 36
		fillBox(rgba, ex, ey, ew, eh, white)
		b.drawName(rgba, ex+12, ey+8, 0x258+b.monID, white)
		alive := 0
		if b.enemyHP > 0 {
			alive = 1
		}
		drawNumber(rgba, b.tx, ex+ew-40, ey+8, alive, white)
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
