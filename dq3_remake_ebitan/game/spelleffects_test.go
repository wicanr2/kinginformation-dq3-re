package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// W3(docs/72 A1)測試:玩家狀態/buff/debuff 咒文(拜基魯多/史卡拉/拉里荷/瑪荷頓/瑪努莎/解咒)+
// per-battle 修正狀態 reset + 咒文選單可選到。用 id101(castProb0/fleeRate0,對齊
// multibattle_test.go 既有慣例)避免敵方 AI 自身施咒/逃跑污染觀察窗,除非該測試就是要測敵方鏡像。

// TestBaikirutoDoublesPartyAtk:拜基魯多(151)施放後 heroAtkPct 應變 200(×2),且套用在物攻傷害上
// (刻意選 atk<2*def,使「未 buff 傷害含會心一擊上限」與「buff 後傷害下限」兩區間不重疊,見注解算式)。
func TestBaikirutoDoublesPartyAtk(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	// atk40/def30:未buff diff10,base5,variance上限2,含會心(×2)上限 14。
	// buff後 atk80/def30:diff50,base25(roll0下限,不含會心已 25 > 14)。兩區間不重疊,可判定。
	for seed := int64(1); seed <= 20; seed++ {
		b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
		hero := heroParams{level: 21, curHP: 999, maxHP: 999, atk: 40, def: 99, agi: 99, mp: 99, maxMP: 99}
		if !b.startGroup(101, 1, seed, hero, nil) {
			t.Fatal("開戰失敗")
		}
		b.enemies[0].def = 30
		b.enemies[0].hp, b.enemies[0].max = 9999, 9999 // 避免中途被打死影響觀察
		b.execSpell(151)                               // 拜基魯多
		if b.heroAtkPct != 200 {
			t.Fatalf("heroAtkPct 應變 200(×2),得 %d(seed%d)", b.heroAtkPct, seed)
		}
		before := b.enemies[0].hp
		b.cursor = bcWar
		b.execTurn()
		dmg := before - b.enemies[0].hp
		if dmg < 25 || dmg > 200 {
			t.Fatalf("拜基魯多後物攻傷害應落 25..200(buffed 下限25),得 %d(seed%d)", dmg, seed)
		}
	}
}

// TestSukaraBuffsPartyDef:史卡拉(154)施放後 heroDefPct 應變 150(上限300)、扣正確 MP,
// 且套用後我方統計期望受傷應下降(300 個 seed 平均值比較,heroDefPct 只是 +50% 非乘 2,
// 用統計平均而非固定區間判定,避免變異數重疊導致偽陽性)。
func TestSukaraBuffsPartyDef(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	measure := func(defPct int, trials int) float64 {
		total := 0
		for seed := int64(1); seed <= int64(trials); seed++ {
			b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
			hero := heroParams{level: 4, curHP: 9999, maxHP: 9999, atk: 1, def: 40, agi: 1}
			if !b.startGroup(101, 1, seed, hero, nil) {
				t.Fatal("開戰失敗")
			}
			b.enemies[0].atk = 100 // 固定敵攻,排除素材本身數值差異
			b.heroDefPct = defPct
			before := b.heroHP
			b.enemyTurn()
			total += before - b.heroHP
		}
		return float64(total) / float64(trials)
	}
	base := measure(100, 300)
	buffed := measure(150, 300)
	if buffed >= base {
		t.Fatalf("heroDefPct 150%% 應降低受傷期望值,base平均=%.2f buffed平均=%.2f", base, buffed)
	}
	t.Logf("史卡拉守備 buff ✓:100%% avg=%.2f → 150%% avg=%.2f", base, buffed)

	// 施放本身:heroDefPct 應變 150，MP 依 EXE DS:0x37c3 扣除。
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 4, curHP: 999, maxHP: 999, atk: 1, def: 40, agi: 1, mp: 99, maxMP: 99}
	if !b.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	mp0 := b.heroMP
	b.execSpell(154)
	if b.heroDefPct != 150 {
		t.Fatalf("史卡拉施放後 heroDefPct 應為150,得 %d", b.heroDefPct)
	}
	if want := mp0 - 3; b.heroMP != want {
		t.Fatalf("史卡拉應依 EXE 扣3 MP,得 heroMP=%d(want %d)", b.heroMP, want)
	}
}

// TestRariHoSleepSkipsEnemyTurn:拉里荷(144)施放後,目標敵 status 應標記睡眠(statusParalysis),
// 且該敵接下來的 enemyTurn 不行動(heroHP 不變)。對齊 C 版行為(dq3_battlescene.c 全檔搜尋無中途
// 自動清除麻痺/睡眠):睡眠會持續整場戰鬥,不會自動甦醒,直到解咒/戰鬥結束才清除——故第二次
// enemyTurn 仍應持續無法行動。
func TestRariHoSleepSkipsEnemyTurn(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	// 明確讓施法者快於敵人；原版已改為敵我敏捷混排，低敏施法者可能在咒文生效前先挨打。
	hero := heroParams{level: 9, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 999, mp: 99, maxMP: 99}
	if !b.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.execSpell(144) // 拉里荷(此呼叫內建 afterLeaderAction 也會跑一次 enemyTurn;敵剛中睡眠應同回合即不行動)
	if b.enemies[0].status&statusParalysis == 0 {
		t.Fatal("拉里荷施放後,敵應陷入睡眠(status statusParalysis)")
	}
	if b.heroHP != 999 {
		t.Fatalf("敵剛陷入睡眠的當回合不應行動,heroHP 應不變,得 %d", b.heroHP)
	}
	before := b.heroHP
	b.enemyTurn() // 再跑一次:睡眠應持續(無自動甦醒)
	if b.heroHP != before {
		t.Fatalf("睡眠應持續整場(對齊 C 無中途清除),敵仍不應行動,heroHP 變 %d→%d", before, b.heroHP)
	}
}

// TestMahotonCastSetsEnemySealed:瑪荷頓(156)施放後,目標敵 status 應標記 statusSealed。
func TestMahotonCastSetsEnemySealed(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 13, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1, mp: 99, maxMP: 99}
	if !b.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.execSpell(156) // 瑪荷頓
	if b.enemies[0].status&statusSealed == 0 {
		t.Fatal("瑪荷頓施放後,敵應標記 statusSealed")
	}
}

// TestSealedEnemyCannotCast:敵鏡像測試(item4,「敵方施放」)—— mon43(素材實測 castProb=255、
// 已知咒 rec=[152 美達巴尼, 158 瑪努莎],docs 調查用 cmd/dumpai 一次性核實,無 flee 干擾)。
// 未封咒時,幾乎每次 CastProb roll 都會嘗試施法(255/256);玩家用瑪荷頓封住該敵後,不論 RNG
// 種子為何,statusSealed 檢查都在擲骰「是否施法」之前短路擋下,故應 100% 不再命中 152/158。
func TestSealedEnemyCannotCast(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	ai, ok := mons.AI(43)
	if !ok || ai.CastProb < 200 || len(spellRecsOf(ai)) == 0 {
		t.Skip("mon43 AI 資料與預期(castProb≈255,已知152/158)不符,素材可能已變,跳過")
	}

	newBattle := func(seed int64) *Battle {
		b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
		hero := heroParams{level: 1, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1}
		if !b.startGroup(43, 1, seed, hero, nil) {
			t.Fatal("開戰失敗(mon43)")
		}
		return b
	}

	// 對照組:未封咒,50 個 seed 內應至少命中一次(否則測試前提本身有誤,先自我核實)。
	landed := false
	for seed := int64(1); seed <= 50; seed++ {
		b := newBattle(seed)
		b.enemyTurn()
		if b.heroStatus&statusParalysis != 0 || b.partyBlind {
			landed = true
			break
		}
	}
	if !landed {
		t.Fatal("前提核實失敗:mon43(castProb≈255)未封咒時 50 個 seed 內應命中152/158其一,一次都沒有,懷疑素材/邏輯已變")
	}

	// 實驗組:封咒後,30 個 seed 皆不應命中(statusSealed 短路擋在擲骰之前,無 RNG 依賴)。
	for seed := int64(1); seed <= 30; seed++ {
		b := newBattle(seed)
		b.enemies[0].status |= statusSealed
		b.enemyTurn()
		if b.heroStatus&statusParalysis != 0 || b.partyBlind {
			t.Fatalf("敵已被瑪荷頓封咒,不應再命中152/158(seed%d):heroStatus=%d partyBlind=%v",
				seed, b.heroStatus, b.partyBlind)
		}
	}
}

// TestManusaRaisesEnemyMissRate:瑪努莎(158)使目標敵陷入幻惑後,該敵物攻應~50% 失手
// (對齊 C `roll255()<128`)。統計 200 次:同一場戰鬥連續呼叫 enemyTurn()(而非逐次用新 seed 開新
// 戰鬥)——DOS 忠實 RNG(internal/rng)是簡單 add+rotate,鄰近 seed 在只推進 1~2 步時輸出會高度
// 相關(已實測驗證:seed1..200 起手第 2 個 roll 幾乎恆定 >=128),必須讓同一串 RNG 狀態連續推進
// 足夠多步才有真正的統計獨立性,同 TestEnemyTurnEachAliveActs 用「單一 seed 內多次行動」取樣的慣例。
func TestManusaRaisesEnemyMissRate(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 1, curHP: 9999, maxHP: 9999, atk: 1, def: 0, agi: 1}
	if !b.startGroup(101, 1, 12345, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.enemies[0].status |= statusBlind // 瑪努莎已命中(施放本身另測 TestManusaCastSetsEnemyBlind)
	trials := 400
	misses := 0
	for i := 0; i < trials; i++ {
		b.heroHP = 9999 // 每次行動前重置,避免累積傷害耗盡目標影響觀察(不重開戰鬥 → RNG 連續推進)
		b.enemyTurn()
		if b.heroHP == 9999 {
			misses++
		}
	}
	if misses < trials*30/100 || misses > trials*70/100 {
		t.Fatalf("幻惑敵物攻 miss 率應~50%%,得 %d/%d", misses, trials)
	}
	t.Logf("瑪努莎幻惑 miss 率 ✓:%d/%d(期望~50%%)", misses, trials)
}

// TestManusaCastSetsEnemyBlind:瑪努莎(158)施放後,目標敵 status 應標記 statusBlind。
func TestManusaCastSetsEnemyBlind(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 7, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1, mp: 99, maxMP: 99}
	if !b.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.execSpell(158) // 瑪努莎
	if b.enemies[0].status&statusBlind == 0 {
		t.Fatal("瑪努莎施放後,敵應標記 statusBlind")
	}
}

// TestCureStatusAndIncapacitatedCaster:失能者不可能替自己施咒解除失能；基阿里仍可由未失能者
// 清除自己的中毒。替其他隊員選目標另由 party-target UI 測試覆蓋。
func TestCureStatusClearsParalysis(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	hero := heroParams{level: 16, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1, mp: 99, maxMP: 99}

	b1 := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	if !b1.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b1.heroStatus |= statusParalysis
	b1.execSpell(167) // 基阿里克
	if b1.heroStatus&statusParalysis == 0 {
		t.Fatal("失能者應被行動佇列跳過，不能替自己施咒解除失能")
	}

	b2 := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	if !b2.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b2.heroStatus |= statusPoison
	b2.execSpell(166) // 基阿里
	if b2.heroStatus&statusPoison != 0 {
		t.Fatal("基阿里後 heroStatus 應清除中毒位元")
	}
}

// TestBattleModsResetEachStart:per-battle 修正狀態(heroAtkPct 等)每場 startGroup 都應歸零
// (對齊 C reset_battle_mods();不可跨戰殘留)。
func TestBattleModsResetEachStart(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 1, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1}
	if !b.startGroup(101, 1, 1, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.heroAtkPct, b.heroDefPct, b.enemyAtkPct, b.enemyDefPct = 400, 300, 400, 300
	b.partyBlind, b.partySealed = true, true
	b.heroStatus = statusParalysis

	if !b.startGroup(101, 1, 2, hero, nil) { // 重開下一場
		t.Fatal("重開戰失敗")
	}
	if b.heroAtkPct != 100 || b.heroDefPct != 100 || b.enemyAtkPct != 100 || b.enemyDefPct != 100 {
		t.Fatalf("新戰應歸零百分比修正(基準100),得 atk%d def%d eatk%d edef%d",
			b.heroAtkPct, b.heroDefPct, b.enemyAtkPct, b.enemyDefPct)
	}
	if b.partyBlind || b.partySealed || b.heroStatus != 0 {
		t.Fatalf("新戰應清除旗標/狀態,得 partyBlind=%v partySealed=%v heroStatus=%d",
			b.partyBlind, b.partySealed, b.heroStatus)
	}
}

// spellRecsOf:測試用小工具,把 AI SpellMask 轉成 rec 清單(供前置核實用,不進正式產品碼)。
func spellRecsOf(ai dq3data.MonsterAI) []int {
	var recs []int
	for b := 0; b < 48; b++ {
		if ai.SpellMask[b/8]&(0x80>>(b%8)) != 0 {
			recs = append(recs, b) // 佔位,實際 rec 轉換見 internal/spell.MonsterSpellRec;此處只需知道「非空」
		}
	}
	return recs
}
