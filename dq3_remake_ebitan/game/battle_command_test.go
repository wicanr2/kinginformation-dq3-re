package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
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

	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 隊長：戰鬥
	if b.commandActor != 1 || b.enemies[0].hp != enemyHP {
		t.Fatalf("隊長下令後只應前進到同伴1，actor=%d enemyHP=%d→%d",
			b.commandActor, enemyHP, b.enemies[0].hp)
	}
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 同伴1：戰鬥
	if b.commandActor != 2 || b.enemies[0].hp != enemyHP {
		t.Fatalf("同伴1下令後仍不可先執行，actor=%d enemyHP=%d→%d",
			b.commandActor, enemyHP, b.enemies[0].hp)
	}
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}) // 同伴2：戰鬥，才結算
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
