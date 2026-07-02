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
	if full.enemyAtk != int(st.Atk) || full.enemyDef != int(st.Def) {
		t.Errorf("無光之珠應全力:atk=%d/%d def=%d/%d", full.enemyAtk, st.Atk, full.enemyDef, st.Def)
	}

	// 持光之珠 → 弱化 atk/3 def/4,旗標清
	weak := &Battle{mons: mons, shp: full.shp}
	weak.lightOrb = true
	weak.start(0x7c, 1, hp, nil)
	if weak.enemyAtk != int(st.Atk)/3 || weak.enemyDef != int(st.Def)/4 {
		t.Errorf("光之珠弱化錯:atk=%d(want %d) def=%d(want %d)",
			weak.enemyAtk, int(st.Atk)/3, weak.enemyDef, int(st.Def)/4)
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

// 索瑪連戰:startZomaSeq → 逐場勝出推進(六大魔人→怨靈→殭屍→索瑪),最終破索瑪 + 設守衛旗標。
func TestZomaSeq(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{flags: map[int]bool{}, endSeq: -1}
	g.battle.mons = mons
	g.battle.shp = asset(t, "DQ3MNS.SHP")

	g.startZomaSeq()
	if !g.battle.active {
		t.Fatal("zomaseq 應開第一場 boss")
	}
	// 驅動連戰:每場模擬勝出 → 推進佇列,直到佇列空。
	seen := []int{g.battle.monID}
	for i := 0; i < 30 && g.battle.active; i++ {
		if g.battle.monID == 0x7c { // 索瑪 → 破關結局
			g.runFinale()
		}
		g.battle.active = false
		if len(g.bossQueue) > 0 {
			g.advanceBossQueue()
			if g.battle.active {
				seen = append(seen, g.battle.monID)
			}
		}
	}
	if !g.cleared {
		t.Errorf("連戰未打到破索瑪,序列=%v", seen)
	}
	if !g.flags[0x214] {
		t.Error("六大魔人破後應設守衛旗標 0x214")
	}
	if seen[len(seen)-1] != 0x7c {
		t.Errorf("連戰最後一場應為索瑪 0x7c,實 %v", seen)
	}

	// flag 0x214 已設時再跑 → 不含六大魔人守衛
	g2 := &Game{flags: map[int]bool{0x214: true}, endSeq: -1}
	g2.battle.mons, g2.battle.shp = mons, g.battle.shp
	g2.startZomaSeq()
	if g2.battle.active && g2.battle.monID == 106 {
		t.Error("守衛旗標已設,不應再打六大魔人")
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
