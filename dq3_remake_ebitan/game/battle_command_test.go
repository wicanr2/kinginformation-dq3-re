package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	dosrng "github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

func TestBattleCommandMenusMatchTXT441To444(t *testing.T) {
	b := &Battle{
		heroHP: 10,
		spells: []int{121},
		companions: []*battleActor{
			{hp: 10},
			{hp: 10, spells: []int{130}},
		},
	}
	cases := []struct {
		actor int
		want  []int
	}{
		{0, []int{bcWar, bcSpell, bcFlee, bcItem}}, // rec443
		{1, []int{bcWar, bcDef, bcItem}},           // rec444
		{2, []int{bcWar, bcSpell, bcDef, bcItem}},  // rec441
	}
	for _, tc := range cases {
		if got := b.commandMenu(tc.actor); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("actor%d menu=%v, want %v", tc.actor, got, tc.want)
		}
	}

	b.heroHP = 0
	if got := b.commandMenu(1); !reflect.DeepEqual(got, []int{bcWar, bcFlee, bcDef, bcItem}) {
		t.Errorf("隊長死亡時首位可動同伴應使用 rec442, got %v", got)
	}
}

func TestBattleCollectsEveryMemberCommandBeforeResolving(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 20, curHP: 200, maxHP: 200, atk: 30, def: 30, agi: 30}
	comps := []*battleActor{
		{hp: 100, maxHP: 100, atk: 20, def: 20, agi: 20},
		{hp: 80, maxHP: 80, mp: 20, maxMP: 20, atk: 10, def: 10, agi: 10, spells: []int{121}},
	}
	if !b.startGroup(5, 1, 9, hero, comps) {
		t.Fatal("開戰失敗")
	}
	enemyHP := b.enemies[0].hp

	chooseAttack := func() {
		b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 戰鬥
		if b.phase != phTargetEnemy {
			t.Fatalf("戰鬥後應進敵方目標選擇，phase=%d", b.phase)
		}
		b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 目標
	}
	chooseAttack() // 隊長
	if b.commandActor != 1 || b.enemies[0].hp != enemyHP {
		t.Fatalf("隊長下令後只應前進到同伴1，actor=%d enemyHP=%d→%d",
			b.commandActor, enemyHP, b.enemies[0].hp)
	}
	chooseAttack() // 同伴1
	if b.commandActor != 2 || b.enemies[0].hp != enemyHP {
		t.Fatalf("同伴1下令後仍不可先執行，actor=%d enemyHP=%d→%d",
			b.commandActor, enemyHP, b.enemies[0].hp)
	}
	chooseAttack() // 同伴2，才結算
	if b.enemies[0].hp >= enemyHP {
		t.Fatalf("收完三人命令後才應執行回合，enemyHP=%d→%d", enemyHP, b.enemies[0].hp)
	}
	if b.phase != phMessage {
		t.Fatalf("回合結算後應停在訊息階段，got %d", b.phase)
	}
}

func TestBattleSkipsIncapacitatedMemberDuringCommandCollection(t *testing.T) {
	b := &Battle{
		heroHP: 10,
		companions: []*battleActor{
			{hp: 0},
			{hp: 10, status: statusParalysis},
			{hp: 10},
		},
		commands: make([]battleCommand, 4),
	}
	if !b.seekCommandActor(0) || b.commandActor != 0 {
		t.Fatal("應先停在存活隊長")
	}
	if !b.seekCommandActor(1) || b.commandActor != 3 {
		t.Fatalf("死亡與異常隊員應跳過，got actor%d", b.commandActor)
	}
}

func TestBattleAllIncapacitatedPartyAdvancesEnemyRound(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	companion := &battleActor{
		hp: 1, maxHP: 10, def: 0, agi: 1, status: statusParalysis,
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 1, curHP: 1, maxHP: 1, def: 0, agi: 1}
	if !b.startGroup(101, 1, 9, hero, []*battleActor{companion}) {
		t.Fatal("開戰失敗")
	}
	b.heroHP = 0
	b.enemies[0].atk, b.enemies[0].agi = 999, 999
	b.phase, b.msg, b.messageResume = phMessage, "上一回合", phEnd

	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phMessage || b.result != 2 || companion.hp != 0 {
		t.Fatalf("全員無法下令時應直接結算敵方回合並判全滅：phase=%d result=%d companionHP=%d actor=%d",
			b.phase, b.result, companion.hp, b.commandActor)
	}
}

func TestBattleSpeedOrderInterleavesEnemyAndParty(t *testing.T) {
	newDuel := func(t *testing.T, heroAgi, enemyAgi int) *Battle {
		t.Helper()
		mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
		if err != nil {
			t.Fatal(err)
		}
		b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
		hero := heroParams{level: 1, curHP: 1, maxHP: 1, atk: 999, def: 0, agi: heroAgi}
		if !b.startGroup(101, 1, 33, hero, nil) {
			t.Fatal("開戰失敗")
		}
		b.enemies[0].hp, b.enemies[0].max = 1, 1
		b.enemies[0].atk, b.enemies[0].agi = 999, enemyAgi
		b.commands[0] = battleCommand{kind: bcWar}
		return b
	}

	fastEnemy := newDuel(t, 1, 1000)
	fastEnemy.resolveRound()
	if fastEnemy.result != 2 {
		t.Fatalf("高速敵應先擊倒隊長，result=%d heroHP=%d enemyHP=%d",
			fastEnemy.result, fastEnemy.heroHP, fastEnemy.enemies[0].hp)
	}

	fastHero := newDuel(t, 1000, 1)
	fastHero.resolveRound()
	if fastHero.result != 1 {
		t.Fatalf("高速隊長應先擊倒敵人，result=%d heroHP=%d enemyHP=%d",
			fastHero.result, fastHero.heroHP, fastHero.enemies[0].hp)
	}
}

func TestBattleRoundMessagesPlayInActionThenResultOrder(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 20, curHP: 200, maxHP: 200, atk: 999, def: 30, agi: 999}
	if !b.startGroup(5, 1, 9, hero, nil) {
		t.Fatal("開戰失敗")
	}
	b.enemies[0].hp, b.enemies[0].max, b.enemies[0].agi = 1, 1, 1
	b.commands[0] = battleCommand{kind: bcWar, target: 0}
	b.resolveRound()

	if b.result != 1 || b.phase != phMessage || b.msg == "" {
		t.Fatalf("結算後應先顯示行動訊息: result=%d phase=%d msg=%q", b.result, b.phase, b.msg)
	}
	first := b.msg
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phMessage || b.msg == first || b.msg == "" {
		t.Fatalf("第一次確認應前進到勝利訊息: phase=%d first=%q msg=%q", b.phase, first, b.msg)
	}
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phEnd {
		t.Fatalf("所有訊息確認完才可進 phEnd，got %d", b.phase)
	}
}

func TestBattleAttackCanSelectSecondEnemy(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 20, curHP: 999, maxHP: 999, atk: 80, def: 999, agi: 999}
	if !b.startGroup(101, 3, 7, hero, nil) {
		t.Fatal("開戰失敗")
	}
	for i := range b.enemies {
		b.enemies[i].hp, b.enemies[i].max = 500, 500
		b.enemies[i].status = statusParalysis
	}
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phTargetEnemy {
		t.Fatalf("戰鬥指令後應選敵，phase=%d", b.phase)
	}
	b.input(InputState{DirHeld: -1, DirEdge: 3}) // 右：第二隻
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.enemies[0].hp != 500 || b.enemies[1].hp >= 500 || b.enemies[2].hp != 500 {
		t.Fatalf("應只攻擊第二隻，HP=[%d %d %d]",
			b.enemies[0].hp, b.enemies[1].hp, b.enemies[2].hp)
	}
}

func TestBattleCompanionSpellSelectsAllyAndUsesOwnMP(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{level: 10, curHP: 100, maxHP: 100, atk: 20, def: 999, agi: 999}
	priest := &battleActor{
		class: 3, level: 10, hp: 20, maxHP: 100, mp: 20, maxMP: 20,
		atk: 10, def: 999, agi: 999, spells: []int{161},
	}
	if !b.startGroup(101, 1, 5, hero, []*battleActor{priest}) {
		t.Fatal("開戰失敗")
	}
	b.enemies[0].status = statusParalysis

	b.input(InputState{DirHeld: -1, DirEdge: 0}) // 隊長：逃跑
	b.input(InputState{DirHeld: -1, DirEdge: 0}) // 隊長：防禦
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.commandActor != 1 {
		t.Fatalf("隊長防禦後應輪到僧侶，actor=%d", b.commandActor)
	}
	b.input(InputState{DirHeld: -1, DirEdge: 0}) // 僧侶：咒文
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 荷依米
	if b.phase != phTargetAlly {
		t.Fatalf("荷依米應開我方目標選擇，phase=%d", b.phase)
	}
	b.input(InputState{DirHeld: -1, DirEdge: 3})                 // 右：僧侶自己
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 結算
	if b.heroHP != 100 || priest.hp <= 20 {
		t.Fatalf("荷依米應只回復僧侶，hero=%d priest=%d", b.heroHP, priest.hp)
	}
	if priest.mp != 17 || b.heroMP != 0 {
		t.Fatalf("應扣僧侶自己的3 MP，heroMP=%d priestMP=%d", b.heroMP, priest.mp)
	}
}

func TestBattleTargetCancelReturnsToOrigin(t *testing.T) {
	b := &Battle{
		heroHP:   10,
		enemies:  []enemyUnit{{hp: 10}},
		commands: emptyCommands(1),
	}
	b.beginTarget(battleCommand{kind: bcWar, target: -1}, phCommand)
	b.input(InputState{Cancel: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phCommand || b.commandActor != 0 {
		t.Fatalf("取消敵方目標應回同一角色命令，phase=%d actor=%d", b.phase, b.commandActor)
	}
}

func TestBattleStatusWakeThresholdMatchesOriginalBranches(t *testing.T) {
	for _, tc := range []struct {
		roll            int
		wantPlayerWakes bool
		wantEnemyWakes  bool
	}{
		{99, true, true},
		{100, true, false},
		{101, false, false},
	} {
		if got := playerStatusWakes(tc.roll); got != tc.wantPlayerWakes {
			t.Errorf("玩家 roll=%d wake=%v，want %v", tc.roll, got, tc.wantPlayerWakes)
		}
		if got := enemyStatusWakes(tc.roll); got != tc.wantEnemyWakes {
			t.Errorf("怪物 roll=%d wake=%v，want %v", tc.roll, got, tc.wantEnemyWakes)
		}
	}
}

func TestBattleActorWakeConsumesCurrentTurn(t *testing.T) {
	b := &Battle{
		rng:            dosrng.New(0),
		heroHP:         10,
		heroStatus:     statusParalysis,
		statusWokeText: "醒來",
	}
	// 找到會命中原版玩家醒來分支的 seed；只用 pure threshold 挑測試輸入，
	// production 仍由同一 DOS RNG 決定。
	for seed := 0; seed <= 0xffff; seed++ {
		r := dosrng.New(uint16(seed))
		if playerStatusWakes(r.Next(256)) {
			b.rng = dosrng.New(uint16(seed))
			break
		}
	}
	if !b.consumeActorSleep(0) {
		t.Fatal("睡眠角色本回合應被狀態 consumer 消耗")
	}
	if b.heroStatus&statusParalysis != 0 {
		t.Fatal("命中醒來分支後應清除玩家睡眠／混亂狀態")
	}
	if b.msg != "醒來" {
		t.Fatalf("醒來訊息=%q", b.msg)
	}
}
