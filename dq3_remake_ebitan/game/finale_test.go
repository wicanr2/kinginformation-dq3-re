package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
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

// runFinale:破索瑪 → 設 ZOMA 里程碑(進度 9/9)+ 洛特冊封 + cleared。
func TestRunFinale(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	// 先把前 8 個里程碑設好,驗破索瑪後進度到 9
	for _, ms := range []int{msStart, msThiefKey, msMagicBal, msRomaly, msDhama, msShip, msRainbow, msDescend} {
		g.progressSet(ms)
	}
	g.runFinale()
	if !g.cleared || !g.lotoBlessed {
		t.Error("破索瑪應設 cleared + lotoBlessed")
	}
	if !g.progressDone(msZoma) {
		t.Error("破索瑪應設 ZOMA 里程碑")
	}
	if s := g.progressStage(); s != msCount {
		t.Errorf("破關後進度階段 %d, want %d(9/9)", s, msCount)
	}
	if !g.flags[0x217] {
		t.Error("應設洛特冊封旗標 0x217")
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

// 結局捲動:破索瑪 → 開 ENDTXT 首段;逐段推進(模擬 Update 的 end_seq 邏輯)直到播完。
func TestEndingScroll(t *testing.T) {
	fon := asset(t, "D3TXT00.FON")
	end := dq3data.LoadText(fon, asset(t, "ENDTXT.TXT"))
	if end == nil || end.NRecords == 0 {
		t.Skip("無 ENDTXT")
	}
	g := &Game{flags: map[int]bool{}, endText: end, endSeq: -1}
	g.runFinale()
	if g.endSeq != 0 || !g.dlg.open || g.dlg.tx != end {
		t.Fatalf("破索瑪應開結局首段:endSeq=%d open=%v", g.endSeq, g.dlg.open)
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
}
