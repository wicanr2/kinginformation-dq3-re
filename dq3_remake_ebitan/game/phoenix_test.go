package game

import (
	"os"
	"testing"
)

func phoenixEventGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyPhoenixShrine, mapBlkNum[ctyPhoenixShrine], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyPhoenixShrine
	g.dlg.tx = sc.dlgText
	return g
}

func testStep(t *testing.T, g *Game, in InputState) {
	t.Helper()
	if in.DirHeld == 0 && in.DirEdge == 0 && !in.Confirm && !in.Cancel {
		in.DirHeld, in.DirEdge = -1, -1
	}
	if err := g.step(in); err != nil {
		t.Fatal(err)
	}
}

func chooseExamine(t *testing.T, g *Game) {
	t.Helper()
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	testStep(t, g, InputState{DirHeld: -1, DirEdge: 2}) // talk(0) → spell(1)
	testStep(t, g, InputState{DirHeld: -1, DirEdge: 1}) // spell(1) → examine(5)
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
}

func TestPhoenixShrineOriginalDataIdentity(t *testing.T) {
	g := phoenixEventGame(t)
	if g.cur.w != 17 || g.cur.h != 20 || g.cur.spawnX != 8 || g.cur.spawnY != 19 {
		t.Fatalf("CTY70 sec0 尺寸/spawn 錯: %dx%d @%d,%d",
			g.cur.w, g.cur.h, g.cur.spawnX, g.cur.spawnY)
	}
	for i, xy := range phoenixAltarTop {
		ev, _, ok := g.cur.tileEvent(xy[0], xy[1])
		want := phoenixAltarFlagFirst + i
		if !ok || ev[0] != 0 || int(ev[2]) != want {
			t.Fatalf("祭壇%d @%v event 錯: ok=%v ev=%v want flag=%x", i, xy, ok, ev, want)
		}
	}
	for _, xy := range [][2]int{{7, 10}, {9, 10}} {
		idx := g.cur.npcAt(xy[0], xy[1])
		if idx < 0 || (g.cur.npcs[idx].ctrl>>3)&7 != 2 || g.cur.npcs[idx].b4 != phoenixGuardianHandler {
			t.Fatalf("守護神 @%v 身分錯: idx=%d", xy, idx)
		}
	}
	unfiltered, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyPhoenixShrine, mapBlkNum[ctyPhoenixShrine], 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	idx := unfiltered.npcAt(phoenixEggX, phoenixEggY)
	if idx < 0 || unfiltered.npcs[idx].spr == nil {
		t.Fatalf("條件式拉米亞 sprite @(%d,%d) 未存在原始 NPC 表, idx=%d", phoenixEggX, phoenixEggY, idx)
	}
}

func TestTedonGreenOrbOriginalNightNPC(t *testing.T) {
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.dnPhase = 2
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyTedon, mapBlkNum[ctyTedon], 0, 2, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty, g.dlg.tx = sc, sc, true, ctyTedon, sc.dlgText
	idx := sc.npcAt(tedonOrbNPCX, tedonOrbNPCY)
	if idx < 0 || (sc.npcs[idx].ctrl>>3)&7 != 2 || sc.npcs[idx].b4 != tedonOrbNPCHandler {
		t.Fatalf("提頓夜間綠寶珠 NPC 錯: idx=%d", idx)
	}
	g.px, g.py, g.facing = tedonOrbNPCX, tedonOrbNPCY+1, 1
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.hasItem(itemGreenOrb) || g.storyFlag(flagTedonOrbReady) || !g.dlg.open {
		t.Fatalf("production 對話未依 handler35 給綠寶珠: inv=%v flag3e=%v dlg=%v",
			g.inventory, g.storyFlag(flagTedonOrbReady), g.dlg.open)
	}
}

func TestPhoenixProductionInputTrace(t *testing.T) {
	g := phoenixEventGame(t)
	g.inventory = []int{0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b}

	// 第一座祭壇完整走 production 命令窗→調查；其餘祭壇呼叫同一 examine dispatch，
	// checkpoint 只省略祭壇之間的無關步行。
	g.px, g.py, g.facing = 8, 12, 1
	chooseExamine(t, g)
	if g.hasItem(0x66) || g.storyFlag(0x8f) || g.storyFlag(0x12c) || !g.dlg.open {
		t.Fatalf("第一祭壇 transaction 錯: inv=%v f8f=%v f12c=%v dlg=%v",
			g.inventory, g.storyFlag(0x8f), g.storyFlag(0x12c), g.dlg.open)
	}
	closeCurrentDialogue(t, g)
	for i := 1; i < 6; i++ {
		xy := phoenixAltarTop[i]
		g.px, g.py, g.facing = xy[0], xy[1]+1, 1
		g.examine()
		if g.storyFlag(phoenixAltarFlagFirst + i) {
			t.Fatalf("祭壇%d 未清旗標%x", i, phoenixAltarFlagFirst+i)
		}
		closeCurrentDialogue(t, g)
	}
	if len(g.inventory) != 0 {
		t.Fatalf("六顆寶珠應全消耗，剩 %v", g.inventory)
	}

	// 和原版影片相同：對守護神談話才復活，並非放下最後一顆立即復活。
	g.px, g.py, g.facing = 7, 11, 1
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.phoenixOwned || g.phoenixStage != 1 || !g.dlg.open {
		t.Fatalf("rec92 前置錯: owned=%v stage=%d dlg=%v", g.phoenixOwned, g.phoenixStage, g.dlg.open)
	}
	closeCurrentDialogue(t, g)
	for i := 0; g.phoenixStage == 3 && i < 64; i++ {
		testStep(t, g, InputState{DirHeld: -1, DirEdge: -1})
	}
	if !g.phoenixOwned || g.phoenixAboard || g.storyFlag(flagPhoenixUnrevived) ||
		g.phoenixX != phoenixParkX || g.phoenixY != phoenixParkY {
		t.Fatalf("復活 transaction 錯: owned=%v aboard=%v flag8e=%v park=%d,%d",
			g.phoenixOwned, g.phoenixAboard, g.storyFlag(flagPhoenixUnrevived), g.phoenixX, g.phoenixY)
	}
	if g.storyFlag(0x11) || g.cur.npcAt(phoenixEggX, phoenixEggY) >= 0 {
		t.Fatal("原版 flag11 僅是復活重載過場暫態，事件結束應清除中央條件 sprite")
	}

	// 原版：離開神殿後走上 (48,189) 搭乘；飛行跨越阻擋地形且不進城。
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1
	g.px, g.py, g.cd = phoenixParkX, phoenixParkY+1, 0
	testStep(t, g, InputState{DirHeld: 1, DirEdge: -1})
	if !g.phoenixAboard || g.px != phoenixParkX || g.py != phoenixParkY {
		t.Fatalf("走上拉米亞未搭乘: aboard=%v @%d,%d", g.phoenixAboard, g.px, g.py)
	}
	g.cd = 0
	testStep(t, g, InputState{DirHeld: 3, DirEdge: -1})
	if g.px != phoenixParkX+1 {
		t.Fatalf("飛行 mode2 應無視地形移動，x=%d", g.px)
	}
}

func TestPhoenixGuardianListsDepositedOrbName(t *testing.T) {
	g := phoenixEventGame(t)
	g.setStoryFlag(phoenixAltarFlagFirst, false)
	g.setStoryFlag(phoenixVisualFlagBase, false)
	g.talkPhoenixGuardian()
	for i := 0; g.dlg.open && g.phoenixStage == 1 && i < 16; i++ {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if !g.dlg.open || g.phoenixStage != 2 || g.dlg.varNumItem != itemGreenOrb {
		t.Fatalf("部分放珠後 rec93 應以 VAR_NUM mode1 列綠寶珠: dlg=%v stage=%d item=%x",
			g.dlg.open, g.phoenixStage, g.dlg.varNumItem)
	}
	got := g.dlg.varGlyphs(0xfff9)
	want := itemNameGlyphs(g.dlg.itemNames, itemGreenOrb)
	if len(got) == 0 || len(got) != len(want) {
		t.Fatalf("VAR_NUM 道具名展開錯: got=%v want=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("VAR_NUM 道具名 glyph[%d]=%d want %d", i, got[i], want[i])
		}
	}
}

func TestPhoenixLandAndSaveCompatibility(t *testing.T) {
	g := phoenixEventGame(t)
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1
	g.phoenixOwned, g.phoenixAboard = true, true

	blocked := false
	for y := 0; y < g.cur.h && !blocked; y++ {
		for x := 0; x < g.cur.w; x++ {
			if g.cur.attr.Raw(g.cur.tileIdx(x, y))&0x64 != 0 {
				g.px, g.py, blocked = x, y, true
				break
			}
		}
	}
	if !blocked {
		t.Fatal("地表找不到原版禁止降落地形")
	}
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if !g.phoenixAboard || g.cmd.open {
		t.Fatalf("attr&0x64 地形不可降落且不應開命令窗: aboard=%v cmd=%v", g.phoenixAboard, g.cmd.open)
	}

	found := false
	for y := 0; y < g.cur.h && !found; y++ {
		for x := 0; x < g.cur.w; x++ {
			if g.cur.attr.Raw(g.cur.tileIdx(x, y))&0x64 == 0 {
				g.px, g.py, found = x, y, true
				break
			}
		}
	}
	if !found {
		t.Fatal("地表找不到原版允許降落地形")
	}
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.phoenixAboard || g.phoenixX != g.px || g.phoenixY != g.py || g.cmd.open {
		t.Fatalf("Confirm 降落錯: aboard=%v park=%d,%d cmd=%v",
			g.phoenixAboard, g.phoenixX, g.phoenixY, g.cmd.open)
	}

	s := g.snapshot()
	var old [32]byte
	copy(old[:], s.StoryBits[:32])
	s.StoryBits = old[:]
	r := phoenixEventGame(t)
	r.restore(s)
	if !r.phoenixOwned || r.phoenixAboard || r.phoenixX != g.phoenixX || r.phoenixY != g.phoenixY {
		t.Fatalf("拉米亞存檔 roundtrip 錯: owned=%v aboard=%v park=%d,%d",
			r.phoenixOwned, r.phoenixAboard, r.phoenixX, r.phoenixY)
	}
	if r.storyBits[31] != old[31] || r.storyBits[32] != 0xff {
		t.Fatalf("32-byte 舊檔相容錯: byte31=%02x byte32=%02x", r.storyBits[31], r.storyBits[32])
	}
}
