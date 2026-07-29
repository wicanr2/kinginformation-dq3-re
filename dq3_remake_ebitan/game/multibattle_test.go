package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// W2(docs/72 A3)測試:多敵戰鬥陣列化 + 怪物群權重隻數。

// TestFirstAliveEnemyTargeting:firstAliveEnemy 回組內第一個存活(hp>0 且未逃走)敵的 index;
// 逐步死亡/逃走應依序退到下一個;全滅/全逃回 -1。
func TestFirstAliveEnemyTargeting(t *testing.T) {
	b := &Battle{enemies: []enemyUnit{{hp: 0}, {hp: 5}, {hp: 3}}}
	if tgt := b.firstAliveEnemy(); tgt != 1 {
		t.Fatalf("首存活應是 index1,得 %d", tgt)
	}
	b.enemies[1].fled = true
	if tgt := b.firstAliveEnemy(); tgt != 2 {
		t.Fatalf("index1 逃走後應退到 index2,得 %d", tgt)
	}
	b.enemies[2].hp = 0
	if tgt := b.firstAliveEnemy(); tgt != -1 {
		t.Fatalf("全滅/全逃應回 -1,得 %d", tgt)
	}
}

// TestGroupBattleAllMustDie:群戰(id4×3)全滅才判勝(對齊 C alive_enemy==0);
// 經驗/金錢按擊殺隻數累加(對齊 apply_victory_exp:(en-g_fled)*ms.exp/gold)。
// 用 id4(fleeRate=0,見 docs 調查:此批怪的唯一已知咒 rec 無 Def 表 → 必落回物攻)避免
// 逃走/施咒讓擊殺數與經驗判定變得不確定。
func TestGroupBattleAllMustDie(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 30, curHP: 999, maxHP: 999, atk: 200, def: 100, agi: 80}
	if !b.startGroup(4, 3, 777, hero, nil) {
		t.Fatal("開群戰失敗(id4×3)")
	}
	if len(b.enemies) != 3 {
		t.Fatalf("應開 3 隻,得 %d", len(b.enemies))
	}
	st, _ := mons.Stat(4)
	turns := 0
	for b.result == 0 && turns < 100 {
		alive := 0
		for _, e := range b.enemies {
			if e.hp > 0 {
				alive++
			}
		}
		b.cursor = bcWar
		b.execTurn()
		turns++
		if alive > 1 && b.result == 1 {
			t.Fatalf("戰前仍有 %d 隻存活卻判勝(第 %d 回合)", alive, turns)
		}
	}
	if b.result != 1 {
		t.Fatalf("強勇者應能全滅 id4×3,得 result=%d(%d 回合)", b.result, turns)
	}
	for i, e := range b.enemies {
		if e.hp != 0 || e.fled {
			t.Errorf("勝利時敵%d 應為擊殺(hp0且未逃),得 hp=%d fled=%v", i, e.hp, e.fled)
		}
	}
	if b.gotExp != 3*int(st.Exp) || b.gotGold != 3*int(st.Gold) {
		t.Fatalf("3 隻全殺應得 3×單隻 EXP/G,得 EXP%d(want %d) G%d(want %d)",
			b.gotExp, 3*int(st.Exp), b.gotGold, 3*int(st.Gold))
	}
	t.Logf("群戰全滅 ✓:%d 回合、EXP%d G%d(3×單隻)", turns, b.gotExp, b.gotGold)
}

// TestShanpaneOriginalMixedFormation 鎖定 DQ3.EXE DGROUP 0x4eb5 的香巴尼塔編隊：
// 4 groups、背景 0x18、page 1、甘達特26×1 + 手下27×1×3。這同時證明 remake
// 不再把混合編隊錯畫／錯算成「第一隻怪 ×4」。
func TestShanpaneOriginalMixedFormation(t *testing.T) {
	exe := asset(t, "DQ3.EXE")
	const formationFile = 0x16140 + 0x4eb5
	wantRaw := []byte{4, 0x18, 1, 26, 1, 27, 1, 27, 1, 27, 1}
	if got := exe[formationFile : formationFile+len(wantRaw)]; string(got) != string(wantRaw) {
		t.Fatalf("香巴尼塔 formation bytes=% x, want % x", got, wantRaw)
	}

	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 30, curHP: 999, maxHP: 999, atk: 999, def: 999, agi: 999}
	groups := []enemyGroup{{monID: 26, count: 1}, {monID: 27, count: 1},
		{monID: 27, count: 1}, {monID: 27, count: 1}}
	if !b.startFormation(groups, 1, hero, nil) {
		t.Fatal("香巴尼塔混合編隊啟動失敗")
	}
	wantIDs := []int{26, 27, 27, 27}
	if len(b.enemies) != len(wantIDs) {
		t.Fatalf("敵數=%d, want %d", len(b.enemies), len(wantIDs))
	}
	for i, id := range wantIDs {
		if b.enemies[i].monID != id || b.enemies[i].spr == nil {
			t.Fatalf("敵%d=%+v, want monID%d + sprite", i, b.enemies[i], id)
		}
	}
	if b.enemies[0].atk == b.enemies[1].atk || b.enemies[0].max == b.enemies[1].max {
		t.Fatalf("甘達特與手下不應共用能力/HP：boss=%+v henchman=%+v", b.enemies[0], b.enemies[1])
	}
	for i := range b.enemies {
		b.enemies[i].hp = 0
	}
	b.finishVictory()
	if b.gotExp != 2200+3*80 || b.gotGold != 0 {
		t.Fatalf("混合編隊獎勵 EXP%d G%d, want EXP2440 G0", b.gotExp, b.gotGold)
	}
}

// TestEnemyTurnEachAliveActs:enemyTurn 逐隻存活敵各行動一次(對齊 C do_turn 敵方段 for 迴圈),
// 非只有組內第一隻打。用 id101×3(實測 fleeRate=0 不逃、castProb=0 不施咒 → 每隻必物攻)+
// 高 HP 零防禦隊長(單一目標,無同伴→ 隨機選目標退化為必中隊長),驗一次 enemyTurn() 造成的
// 總傷害落在「3 隻各物攻一次」的期望區間(PhysDamage(atk29,def0,*)=14+variance(0..6),
// 3 隻合計 42..60);若只有 1 隻行動,傷害會落在 14..20,可清楚區分。
func TestEnemyTurnEachAliveActs(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 1, curHP: 999, maxHP: 999, atk: 1, def: 0, agi: 1}
	if !b.startGroup(101, 3, 55, hero, nil) {
		t.Fatal("開群戰失敗(id101×3)")
	}
	before := b.heroHP
	b.enemyTurn()
	lost := before - b.heroHP
	if lost < 42 || lost > 60 {
		t.Fatalf("3 隻應各物攻一次(合計傷害應落 42..60),得 %d(僅 1 隻行動的話應落 14..20)", lost)
	}
	if b.result != 0 {
		t.Fatalf("我方 HP 充足不應判敗,得 result=%d", b.result)
	}
	t.Logf("敵方回合逐隻行動 ✓:3 隻合計傷害 %d(區間 42..60)", lost)
}

// TestSpellTargetSingleVsGroup:直接呼叫相容入口時，單體咒(美拉121)預設第一個存活敵；
// 正常輸入的指定目標另由 battle_command_test 覆蓋。群體咒(吉拉124)波及全部存活敵(對齊
// TG_ENEMY1 vs 其餘 target 分支)。用 id101(實測 castProb=0)避免 execSpell 內建的
// afterLeaderAction→enemyTurn() 順帶觸發敵方自身咒文(如自我回復)污染觀察窗。
func TestSpellTargetSingleVsGroup(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	hero := heroParams{level: 10, curHP: 999, maxHP: 999, atk: 1, def: 1, agi: 1, mp: 99, maxMP: 99}

	b1 := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	if !b1.startGroup(101, 3, 42, hero, nil) {
		t.Fatal("開群戰失敗(單體咒案例)")
	}
	hp0 := [3]int{b1.enemies[0].hp, b1.enemies[1].hp, b1.enemies[2].hp}
	b1.execSpell(121) // 美拉:單體
	if b1.enemies[0].hp >= hp0[0] {
		t.Errorf("單體咒應傷第一個存活敵,hp 未降:%d→%d", hp0[0], b1.enemies[0].hp)
	}
	if b1.enemies[1].hp != hp0[1] || b1.enemies[2].hp != hp0[2] {
		t.Errorf("單體咒不應波及其他敵,得 [%d %d](原 [%d %d])",
			b1.enemies[1].hp, b1.enemies[2].hp, hp0[1], hp0[2])
	}

	b2 := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	if !b2.startGroup(101, 3, 42, hero, nil) {
		t.Fatal("開群戰失敗(群體咒案例)")
	}
	hp1 := [3]int{b2.enemies[0].hp, b2.enemies[1].hp, b2.enemies[2].hp}
	b2.execSpell(124) // 吉拉:群體
	for i := 0; i < 3; i++ {
		if b2.enemies[i].hp >= hp1[i] {
			t.Errorf("群體咒應波及敵%d,hp 未降:%d→%d", i, hp1[i], b2.enemies[i].hp)
		}
	}
	t.Logf("咒文目標範圍 ✓:單體咒只傷首敵、群體咒波及全部")
}

// TestSpawnGroupMax:出現權重→群量上限(docs/13:floor(budget38/weight),夾 [1,8])。
func TestSpawnGroupMax(t *testing.T) {
	cases := []struct{ weight, want int }{
		{8, 4},   // floor(38/8)=4(對齊實測 id0 spawn_weight=8)
		{4, 8},   // floor(38/4)=9 → clip 群上限8(對齊 id5 史萊姆 spawn_weight=4,docs game3.png 史萊姆×6 案例的量級)
		{1, 8},   // floor(38/1)=38 → clip 8
		{100, 1}, // floor(38/100)=0 → 保底 1
		{0, 1},   // 非群戰候選 → 1
		{-5, 1},  // 異常負值 → 1
	}
	for _, c := range cases {
		if got := spawnGroupMax(c.weight); got != c.want {
			t.Errorf("spawnGroupMax(%d)=%d, want %d", c.weight, got, c.want)
		}
	}
}

// TestEncounterGroupCount:g.encounterGroupCount 依真實 D3MNS 權重擲值,落在 [1,maxN] 範圍內、
// 多次擲值應覆蓋多種隻數(非恆定同一值,證明真的有走亂數分支,非誤寫死 1)。
func TestEncounterGroupCount(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{}
	g.battle.mons = mons
	g.prng.Seed(9999)

	st, ok := mons.Stat(0) // id0(阿里阿罕弱怪池成員),實測 spawn_weight=8 → maxN=4
	if !ok {
		t.Skip("無 id0 怪物資料")
	}
	maxN := spawnGroupMax(int(st.SpawnWeight))
	seen := map[int]bool{}
	for i := 0; i < 200; i++ {
		n := g.encounterGroupCount(0)
		if n < 1 || n > maxN {
			t.Fatalf("weight=%d maxN=%d,擲出越界隻數 %d", st.SpawnWeight, maxN, n)
		}
		seen[n] = true
	}
	if maxN > 1 && len(seen) < 2 {
		t.Errorf("200 次擲值應覆蓋 1..%d 多種隻數,只見 %v(疑似亂數分支未走到)", maxN, seen)
	}
	t.Logf("遭遇群量擲值 ✓:weight=%d → maxN=%d,200 次覆蓋 %v", st.SpawnWeight, maxN, seen)

	// 無效 monID(越界,mons.Stat 查不到)→ 固定 1
	if n := g.encounterGroupCount(9999); n != 1 {
		t.Errorf("無效 monID 應回 1,得 %d", n)
	}
	// mons 未設 → 固定 1
	g2 := &Game{}
	if n := g2.encounterGroupCount(0); n != 1 {
		t.Errorf("battle.mons 為 nil 應回 1,得 %d", n)
	}
}

func TestEncounterSubtableThresholdForcesSingleEnemy(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	g := &Game{}
	g.battle.mons = mons
	g.prng.Seed(1)
	for i := 0; i < 64; i++ {
		if got := g.encounterGroupCount(6, 100); got != 1 {
			t.Fatalf("threshold100 必須強制單隻，got %d", got)
		}
	}
}

// TestDumpMultiEnemyBattle:視覺核驗——多隻同種怪橫排的戰鬥畫面 dump 成 PNG(規則 63/64:
// 靜態/單元測試之外留一份可肉眼核對的畫面,證明 N 隻 sprite 真的畫出來)。平常 skip;
// 設 DQ3_DUMP_MULTI=1 + MULTI_OUT=<gitignored 夾> 才跑。用 DQ3_BATTLE/DQ3_BATTLE_N debug hook
// (game.go applyDebugEnv)開一場史萊姆(id5)×6 戰鬥,對齊 docs/13「game3.png 史萊姆×6」案例。
func TestDumpMultiEnemyBattle(t *testing.T) {
	if os.Getenv("DQ3_DUMP_MULTI") == "" {
		t.Skip("設 DQ3_DUMP_MULTI=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("MULTI_OUT")
	if out == "" {
		t.Fatal("需設 MULTI_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	t.Setenv("DQ3_BATTLE", "5")   // 史萊姆
	t.Setenv("DQ3_BATTLE_N", "6") // ×6(對齊 game3.png 案例)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !g.battle.active || len(g.battle.enemies) != 6 {
		t.Fatalf("DQ3_BATTLE=5 DQ3_BATTLE_N=6 應開 6 隻史萊姆群戰,得 active=%v 隻數=%d",
			g.battle.active, len(g.battle.enemies))
	}

	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped", name)
	}
	dump("multi_slime_x6_open")

	// 打贏一隻(cursor=戰う 反覆攻擊組內第一個存活敵)後再 dump 一張,驗證死亡的少畫、存活的仍橫排。
	for i := 0; i < 200 && g.battle.result == 0; i++ {
		alive := 0
		for _, e := range g.battle.enemies {
			if e.hp > 0 {
				alive++
			}
		}
		if alive <= 3 { // 打到剩 3 隻就停,留下「部分死亡」的畫面可對照
			break
		}
		g.battle.cursor = bcWar
		g.battle.execTurn()
		if g.battle.phase == phMessage && g.battle.result == 0 {
			g.battle.phase = phCommand // 模擬玩家按 A 推進訊息,繼續下一回合
		}
	}
	dump("multi_slime_partial_kill")
	t.Logf("多敵戰鬥 dump ✓:開場 6 隻、部分擊殺後畫面,存 %s", out)
}
