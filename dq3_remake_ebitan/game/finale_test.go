package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 這些 raw selector 只供原版 identity／input trace tests；production runtime
// 由 events.json.story_flag_runtime_events 提供同一份資料。
const (
	ctyRadatomeCastle  = 80
	radatomeThroneSec  = 1
	radatomeKingHandle = 74
	radatomeEndingRec  = 48
)

// 光之珠弱化:索瑪(0x7c)持光之珠 → 開戰時敵攻/3、防/4、旗標清;無光之珠 → 全力。
func TestZomaLightOrbWeaken(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	st, ok := mons.Stat(0x7c)
	if !ok {
		t.Skip("無索瑪數值")
	}
	full := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hp := heroParams{level: 40, curHP: 400, maxHP: 400, atk: 200, def: 100, agi: 60}

	// 無光之珠 → 全力
	full.lightOrb = false
	if !full.start(0x7c, 1, hp, nil) {
		t.Fatal("索瑪開戰失敗")
	}
	if full.enemies[0].atk != int(st.Atk) || full.enemies[0].def != int(st.Def) {
		t.Errorf("無光之珠應全力:atk=%d/%d def=%d/%d", full.enemies[0].atk, st.Atk, full.enemies[0].def, st.Def)
	}

	// 持光之珠 → 弱化 atk/3 def/4,旗標清
	weak := &Battle{mons: mons, shp: full.shp}
	weak.lightOrb = true
	weak.start(0x7c, 1, hp, nil)
	if weak.enemies[0].atk != int(st.Atk)/3 || weak.enemies[0].def != int(st.Def)/4 {
		t.Errorf("光之珠弱化錯:atk=%d(want %d) def=%d(want %d)",
			weak.enemies[0].atk, int(st.Atk)/3, weak.enemies[0].def, int(st.Def)/4)
	}
	if weak.lightOrb {
		t.Error("光之珠用畢應清旗標")
	}
}

// runFinale:破索瑪 → 設 ZOMA 里程碑並回下層地表；尚未由國王冊封。
func TestRunFinale(t *testing.T) {
	g := r4Game(t)
	// 先把前 8 個里程碑設好,驗破索瑪後進度到 9
	for _, ms := range []int{msStart, msThiefKey, msMagicBal, msRomaly, msDhama, msShip, msRainbow, msDescend} {
		g.progressSet(ms)
	}
	g.setStoryFlag(0xe1, true)
	g.runFinale()
	if !g.cleared || g.lotoBlessed {
		t.Error("破索瑪應 cleared，但見國王前不可預先取得洛特封號")
	}
	if !g.progressDone(msZoma) {
		t.Error("破索瑪應設 ZOMA 里程碑")
	}
	if s := g.progressStage(); s != msCount {
		t.Errorf("破關後進度階段 %d, want %d(9/9)", s, msCount)
	}
	if g.flags[0x217] || g.storyFlag(0xe1) {
		t.Error("索瑪勝利應 CLEAR e1，且國王前不可設冊封旗標 0x217")
	}
	if g.inTown || g.layer != 1 || g.cur != g.under ||
		g.px != ctyLoc[ctyZomaCastle][0] || g.py != ctyLoc[ctyZomaCastle][1] {
		t.Fatalf("應回索瑪城外下層地表：town=%v layer=%d @(%d,%d)",
			g.inTown, g.layer, g.px, g.py)
	}
	if g.endSeq >= 0 || g.dlg.open {
		t.Fatal("索瑪勝利後不可立刻播放 ending")
	}
}

func TestZomaAftermathRadatomeRoute(t *testing.T) {
	g := r4Game(t)
	g.runFinale()
	g.inventory = []int{0x43} // 蓋美拉翅膀；影片同樣由索瑪城外傳送回拉達多姆
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if !g.inTown || g.layer != 1 || g.curCty != 79 {
		t.Fatalf("下層 ReturnTown 應抵達 CTY79 拉達多姆外城：town=%v layer=%d cty=%d",
			g.inTown, g.layer, g.curCty)
	}

	outside, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 79, mapBlkNum[79], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	dc, ds, dx, dy, ok := outside.tileTransition(44, 5)
	if !ok || dc != ctyRadatomeCastle || ds != 0 || dx != 11 || dy != 30 {
		t.Fatalf("CTY79→CTY80 城門資料錯：ok=%v %d.%d@(%d,%d)", ok, dc, ds, dx, dy)
	}
	castle, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyRadatomeCastle, mapBlkNum[ctyRadatomeCastle], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	dc, ds, dx, dy, ok = castle.tileTransition(15, 8)
	if !ok || dc != ctyRadatomeCastle || ds != radatomeThroneSec || dx != 13 || dy != 22 {
		t.Fatalf("CTY80 sec0→王座 sec1 資料錯：ok=%v %d.%d@(%d,%d)", ok, dc, ds, dx, dy)
	}
}

// 索瑪連戰原版 formations：怨靈0x7a→殭屍0x7b→rec72→索瑪0x7c。
func TestZomaSeq(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{flags: map[int]bool{}, endSeq: -1}
	g.battle.mons = mons
	g.battle.shp = asset(t, "DQ3MNS.SHP")
	g.dlg.tx = dq3data.LoadText(asset(t, "D3TXT00.FON"), asset(t, "D3TXT07.TXT"))

	g.startZomaSeq()
	if !g.battle.active {
		t.Fatal("zomaseq 應開第一場 boss")
	}
	// 驅動連戰；索瑪前必須停在原版 rec72，而非直接開戰。
	seen := []int{g.battle.monID}
	for i := 0; i < 2; i++ {
		g.battle.active = false
		g.advanceBossQueue()
		if g.battle.active {
			seen = append(seen, g.battle.monID)
		}
	}
	if len(seen) != 2 || seen[0] != 0x7a || seen[1] != 0x7b {
		t.Fatalf("索瑪前 formation 錯: %v", seen)
	}
	if !g.zomaIntro || !g.dlg.open {
		t.Fatal("殭屍後應先顯示索瑪 rec72")
	}
	g.dlg.open = false
	g.advanceZomaIntro()
	if !g.battle.active || g.battle.monID != 0x7c {
		t.Fatalf("rec72 關閉後應開索瑪0x7c，active=%v id=%02x", g.battle.active, g.battle.monID)
	}
}

// 結局捲動:破索瑪後回 CTY80 sec1 自然與國王交談，才開 ENDTXT。
func TestEndingScroll(t *testing.T) {
	g := r4Game(t)
	end := g.endText
	if end == nil || end.NRecords == 0 {
		t.Skip("無 ENDTXT")
	}
	g.runFinale()
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyRadatomeCastle,
		mapBlkNum[ctyRadatomeCastle], radatomeThroneSec, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, ctyRadatomeCastle
	g.dlg.tx = sc.dlgText
	var king *npcInst
	for i := range sc.npcs {
		if sc.npcs[i].b4 == radatomeKingHandle {
			king = &sc.npcs[i]
			break
		}
	}
	if king == nil {
		t.Fatal("CTY80 sec1 找不到原版 handler74 國王")
	}
	g.px, g.py, g.facing = king.x, king.y+1, 1
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if g.endingKingStage != 1 || !g.dlg.open ||
		!reflect.DeepEqual(g.dlg.buf, sc.dlgText.Record(radatomeEndingRec)) {
		t.Fatal("索瑪後和國王交談應走 handler74/sub_1E713 rec48")
	}
	for g.endingKingStage == 1 {
		testStep(t, g, InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if !g.lotoBlessed || !g.flags[0x217] || g.endSeq != 0 || !g.dlg.open || g.dlg.tx != end {
		t.Fatalf("國王冊封後才應開結局：loto=%v flag=%v endSeq=%d open=%v",
			g.lotoBlessed, g.flags[0x217], g.endSeq, g.dlg.open)
	}
	// 模擬逐段推進:每段翻完關閉 → 下一段,直到 endSeq 回 -1(播完)。
	steps := 0
	for g.endSeq >= 0 && steps < end.NRecords+2 {
		g.dlg.open = false // 該段翻完關閉
		g.endSeq++         // = Update 的自動推進
		if g.endSeq < end.NRecords {
			g.dlg.Open(g.endSeq)
		} else {
			g.endSeq = -1
		}
		steps++
	}
	if g.endSeq != -1 {
		t.Errorf("結局應播完(endSeq=-1),實 %d", g.endSeq)
	}
	if steps != end.NRecords {
		t.Errorf("推進步數 %d != ENDTXT 段數 %d", steps, end.NRecords)
	}
	if len(g.endingPix) != ScreenW*ScreenH || len(g.endingPal) == 0 {
		t.Fatalf("TIT3.P 終盤 pack asset 未載入：pix=%d pal=%d", len(g.endingPix), len(g.endingPal))
	}
	g.renderFrame()
	for i, idx := range g.endingPix {
		want := dq3data.Color{}
		if int(idx) < len(g.endingPal) {
			want = g.endingPal[idx]
		}
		o := i * 4
		if g.rgba[o] != want.R || g.rgba[o+1] != want.G || g.rgba[o+2] != want.B || g.rgba[o+3] != 255 {
			t.Fatalf("THE END renderer pixel %d = %v,%v,%v,%v want %v,%v,%v,255",
				i, g.rgba[o], g.rgba[o+1], g.rgba[o+2], g.rgba[o+3], want.R, want.G, want.B)
		}
	}
}
