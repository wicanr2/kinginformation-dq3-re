package game

// 拉米亞事件資料由 DQ3.EXE file 0x7425..0x7527 與 CTY70.DAT 交叉確認。
// 舊 C remake 將神殿誤標 CTY82 並加了召喚快捷鍵；原版是 CTY70 復活後停在地表，
// 玩家走上 sprite 搭乘、飛行中按一般確認鍵降落。
const (
	ctyTedon           = 20
	tedonOrbNPCX       = 16
	tedonOrbNPCY       = 2
	tedonOrbNPCHandler = 35
	flagTedonOrbReady  = 0x3e

	ctyPhoenixShrine       = 70
	phoenixGuardianHandler = 69
	flagPhoenixUnrevived   = 0x8e
	phoenixAltarFlagFirst  = 0x8f
	phoenixAltarFlagLast   = 0x94
	phoenixVisualFlagBase  = 0x12c
	phoenixEggX            = 8
	phoenixEggY            = 8
	phoenixParkX           = 0x30
	phoenixParkY           = 0xbd

	itemGreenOrb  = 0x66
	itemSilverOrb = 0x6b
)

var phoenixAltarTop = [6][2]int{
	{8, 11}, {4, 9}, {4, 5}, {8, 3}, {12, 5}, {12, 9},
}

// talkTedonOrb 對齊 EXE file 0x6caa：夜間亡魂 handler35，
// flag3e 尚設時 rec38 並交付綠寶珠，之後清旗標；否則 rec39。
func (g *Game) talkTedonOrb() {
	if !g.storyFlag(flagTedonOrbReady) {
		g.dlg.Open(39)
		return
	}
	g.inventory = append(g.inventory, itemGreenOrb)
	g.noticeCode, g.noticeTimer = itemGreenOrb, 120
	g.setStoryFlag(flagTedonOrbReady, false)
	g.dlg.Open(38)
}

// tryPhoenixAltar 對齊 CTY70 type0 event 與 EXE file 0x7425..0x74a8。
// 六座祭壇接受背包中最低 code 的寶珠 66..6b；清 8f..94 表示祭壇已完成，
// 清 12c..131 控制該顆寶珠的可見狀態。
func (g *Game) tryPhoenixAltar(x, y int) bool {
	if !g.inTown || g.cur == nil || g.curCty != ctyPhoenixShrine || g.cur.sec != 0 {
		return false
	}
	ev, _, ok := g.cur.tileEvent(x, y)
	if !ok || ev[0] != 0 {
		return false
	}
	flag := int(ev[2])
	if flag < phoenixAltarFlagFirst || flag > phoenixAltarFlagLast {
		return false
	}
	if !g.storyFlag(flag) {
		return true
	}
	item := -1
	for code := itemGreenOrb; code <= itemSilverOrb; code++ {
		if g.hasItem(code) {
			item = code
			break
		}
	}
	if item < 0 {
		return true
	}
	g.removeItems(item, 1)
	g.setStoryFlag(phoenixVisualFlagBase+item-itemGreenOrb, false)
	g.setStoryFlag(flag, false)
	g.dlg.Open(96)
	g.setDlgVarNumItem(item)
	return true
}

func (g *Game) placedPhoenixOrbs() []int {
	var out []int
	for item := itemGreenOrb; item <= itemSilverOrb; item++ {
		if !g.storyFlag(phoenixVisualFlagBase + item - itemGreenOrb) {
			out = append(out, item)
		}
	}
	return out
}

// talkPhoenixGuardian 對齊 handler69(file 0x74a9)：未復活時先說 rec92，
// 部分寶珠逐顆 rec93；六顆齊備則進復活動畫 transaction。復活後為 rec94。
func (g *Game) talkPhoenixGuardian() {
	if !g.storyFlag(flagPhoenixUnrevived) {
		g.dlg.Open(94)
		return
	}
	g.phoenixList = g.placedPhoenixOrbs()
	g.phoenixListPos = 0
	g.phoenixStage = 1
	g.dlg.Open(92)
}

func (g *Game) advancePhoenixEvent() {
	if g.phoenixStage == 0 || g.dlg.open {
		return
	}
	switch g.phoenixStage {
	case 1:
		if len(g.phoenixList) == 6 {
			g.revivePhoenix()
			return
		}
		if len(g.phoenixList) == 0 {
			g.phoenixStage = 0
			return
		}
		g.phoenixStage = 2
	case 2:
		g.phoenixListPos++
	}
	if g.phoenixListPos >= len(g.phoenixList) {
		g.phoenixStage = 0
		return
	}
	item := g.phoenixList[g.phoenixListPos]
	g.dlg.Open(93)
	g.setDlgVarNumItem(item)
}

func (g *Game) revivePhoenix() {
	g.phoenixStage = 3
	g.phoenixAnim, g.phoenixAnimTick = 0, 0
	g.phoenixList = nil
}

// advancePhoenixAnimation 對齊 file 0x7528..0x753d：中央 (8,8) 依序畫
// tile 0x7b..0x80。原版是同步 delay；Ebiten 以每 8 update 一幀重現且封鎖輸入。
func (g *Game) advancePhoenixAnimation() {
	if g.phoenixStage != 3 {
		return
	}
	g.phoenixAnimTick++
	if g.phoenixAnimTick < 8 {
		return
	}
	g.phoenixAnimTick = 0
	g.phoenixAnim++
	if g.phoenixAnim < 6 {
		return
	}
	g.setStoryFlag(flagPhoenixUnrevived, false)
	g.setStoryFlag(0x11, false) // file 0x7593：只在原版重載過場中暫設，return 前清除
	g.phoenixOwned, g.phoenixAboard = true, false
	g.phoenixX, g.phoenixY = phoenixParkX, phoenixParkY
	g.phoenixStage = 0
	g.reloadTownDaynight()
}

// tryLandPhoenix 對齊 EXE file 0x434f：attr&0x64 非零不可降落。
func (g *Game) tryLandPhoenix() bool {
	if !g.phoenixAboard || g.inTown || g.cur == nil {
		return false
	}
	if g.cur.attr.Raw(g.cur.tileIdx(g.px, g.py))&0x64 != 0 {
		return false
	}
	g.phoenixAboard = false
	g.phoenixOwned = true
	g.phoenixX, g.phoenixY = g.px, g.py
	return true
}

// drawPhoenixAltars 重現 CTY70 的條件圖塊。祭壇平時由場景畫空盆；
// 旗標 8f..94 清除（已放珠）後，把空盆 tile 0x94 覆成紅珠 tile 0x95。
// 0x7b..0x80 是復活時中央蛋的六幀動畫，不能誤拿 0x7b 當祭壇寶珠。
func (g *Game) drawPhoenixAltars(camX, camY int) {
	if g.cur == nil || g.curCty != ctyPhoenixShrine || g.cur.sec != 0 {
		return
	}
	for i, xy := range phoenixAltarTop {
		if g.storyFlag(phoenixAltarFlagFirst + i) {
			continue
		}
		x, y := xy[0], xy[1]
		if x >= camX && x < camX+ViewCols && y >= camY && y < camY+ViewRows {
			blitTile(g.rgba, (x-camX)*TileW, (y-camY)*TileH, g.cur.blk.Tile(0x95), g.cur.pal)
		}
	}
	if g.phoenixStage == 3 && phoenixEggX >= camX && phoenixEggX < camX+ViewCols &&
		phoenixEggY >= camY && phoenixEggY < camY+ViewRows {
		frame := g.phoenixAnim
		if frame < 0 {
			frame = 0
		} else if frame > 5 {
			frame = 5
		}
		blitTile(g.rgba, (phoenixEggX-camX)*TileW, (phoenixEggY-camY)*TileH,
			g.cur.blk.Tile(0x7b+frame), g.cur.pal)
	}
}
