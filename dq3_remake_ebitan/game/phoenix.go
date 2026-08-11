package game

// 拉米亞事件資料由 DQ3.EXE file 0x7425..0x7527 與 CTY70.DAT 交叉確認。
// 舊 C remake 將神殿誤標 CTY82 並加了召喚快捷鍵；原版是 CTY70 復活後停在地表，
// 玩家走上 sprite 搭乘、飛行中按一般確認鍵降落。
const (
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

// tryPhoenixAltar 對齊 CTY70 type0 event 與 EXE file 0x7425..0x74a8。
// 六座祭壇接受背包中最低 code 的寶珠 66..6b；清 8f..94 表示祭壇已完成，
// 清 12c..131 控制該顆寶珠的可見狀態。
func (g *Game) tryPhoenixAltar(x, y int) bool {
	e, ok := g.phoenixRuntimeEvent()
	if !ok || !g.inTown || g.cur == nil || g.curCty != e.NPC.CTYRaw || currentSceneSection(g.cur) != e.NPC.Section {
		return false
	}
	ev, _, ok := g.cur.tileEvent(x, y)
	if !ok || ev[0] != 0 {
		return false
	}
	flag := int(ev[2])
	altarIndex := -1
	for i, altarFlag := range e.PhoenixAltarFlagRaw {
		if flag == altarFlag {
			altarIndex = i
			break
		}
	}
	if altarIndex < 0 {
		return false
	}
	if !g.storyFlag(flag) {
		return true
	}
	item, owner := -1, -1 // 0=勇者，1..=目前隊伍中的同伴
	for code := e.PhoenixOrbItemFirstRaw; code <= e.PhoenixOrbItemLastRaw; code++ {
		if g.hasItem(code) {
			item, owner = code, 0
			break
		}
		for i, member := range g.companions {
			if containsInt(member.Inventory, code) {
				item, owner = code, i+1
				break
			}
		}
		if item >= 0 {
			break
		}
	}
	if item < 0 {
		return true
	}
	if owner == 0 {
		g.removeItems(item, 1)
	} else {
		member := g.companions[owner-1]
		for i, code := range member.Inventory {
			if code == item {
				member.Inventory = append(member.Inventory[:i], member.Inventory[i+1:]...)
				break
			}
		}
	}
	g.setStoryFlag(e.PhoenixVisualFlagBaseRaw+item-e.PhoenixOrbItemFirstRaw, false)
	g.setStoryFlag(flag, false)
	g.dlg.Open(e.PhoenixAltarDialogueRaw)
	g.setDlgVarNumItem(item)
	return true
}

func (g *Game) placedPhoenixOrbs() []int {
	e, ok := g.phoenixRuntimeEvent()
	if !ok {
		return nil
	}
	var out []int
	for item := e.PhoenixOrbItemFirstRaw; item <= e.PhoenixOrbItemLastRaw; item++ {
		if !g.storyFlag(e.PhoenixVisualFlagBaseRaw + item - e.PhoenixOrbItemFirstRaw) {
			out = append(out, item)
		}
	}
	return out
}

// talkPhoenixGuardian 對齊 handler69(file 0x74a9)：未復活時先說 rec92，
// 部分寶珠逐顆 rec93；六顆齊備則進復活動畫 transaction。復活後為 rec94。
func (g *Game) talkPhoenixGuardian() {
	e, ok := g.phoenixRuntimeEvent()
	if !ok {
		return
	}
	if !g.storyFlag(e.PhoenixUnrevivedFlagRaw) {
		g.dlg.Open(e.PhoenixRevivedDialogueRaw)
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
	e, ok := g.phoenixRuntimeEvent()
	if !ok {
		g.phoenixStage = 0
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
	g.dlg.Open(e.PhoenixOrbDialogueRaw)
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
	e, ok := g.phoenixRuntimeEvent()
	if !ok || e.Animation == nil || g.phoenixStage != 3 {
		return
	}
	g.phoenixAnimTick++
	if g.phoenixAnimTick < e.Animation.FrameTicks {
		return
	}
	g.phoenixAnimTick = 0
	g.phoenixAnim++
	if g.phoenixAnim < len(e.Animation.TilesRaw) {
		return
	}
	g.applyStoryFlagLists(e)
	g.phoenixMapCellWritten = true
	// handler69 sets the transient flag while it writes the map cell and
	// rebuilds the NPC list; it clears the flag before returning.
	g.reloadStoryFlagScene(e)
	for _, flag := range e.SetStoryFlagsRaw {
		if flag == e.PhoenixTransientFlagRaw {
			g.setStoryFlag(flag, false)
		}
	}
	g.reloadStoryFlagScene(e)
	g.phoenixOwned, g.phoenixAboard = true, false
	g.phoenixX, g.phoenixY = e.ParkPosition.X, e.ParkPosition.Y
	g.phoenixStage = 0
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
	e, ok := g.phoenixRuntimeEvent()
	if !ok || e.Animation == nil || g.cur == nil || g.curCty != e.NPC.CTYRaw || currentSceneSection(g.cur) != e.NPC.Section {
		return
	}
	for i, xy := range e.PhoenixAltarTiles {
		if i >= len(e.PhoenixAltarFlagRaw) || g.storyFlag(e.PhoenixAltarFlagRaw[i]) {
			continue
		}
		x, y := xy.X, xy.Y
		if x >= camX && x < camX+ViewCols && y >= camY && y < camY+ViewRows {
			blitTile(g.rgba, (x-camX)*TileW, (y-camY)*TileH,
				g.cur.blk.Tile(e.PhoenixAltarVisualTileRaw), g.cur.pal)
		}
	}
	pos := e.Animation.Position
	if g.phoenixStage == 3 && pos.X >= camX && pos.X < camX+ViewCols &&
		pos.Y >= camY && pos.Y < camY+ViewRows {
		frame := g.phoenixAnim
		if frame < 0 {
			frame = 0
		} else if frame >= len(e.Animation.TilesRaw) {
			frame = len(e.Animation.TilesRaw) - 1
		}
		blitTile(g.rgba, (pos.X-camX)*TileW, (pos.Y-camY)*TileH,
			g.cur.blk.Tile(e.Animation.TilesRaw[frame]), g.cur.pal)
	}
}
