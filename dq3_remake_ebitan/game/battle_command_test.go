package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	dosrng "github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

type fakeBattleSFX struct {
	played   []int
	duration map[int]int64
}

func (f *fakeBattleSFX) PlaySFX(cue int) { f.played = append(f.played, cue) }
func (f *fakeBattleSFX) SFXDurationNanos(cue int) int64 {
	return f.duration[cue]
}

func TestBattleFleeSFXWaitGate(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := pack.BattleSoundCues()
	fake := &fakeBattleSFX{duration: map[int]int64{
		13: 524362986,
		21: 233506233,
	}}
	b := &Battle{
		phase: phMessage, result: 3,
		texts: map[string]battleTextDefinition{
			battleTextActorFled: {value: battleTextActorFled},
		},
	}
	b.setSoundCues(cues, fake)
	b.emitTextWithSFX(battleTextActorFled, cues.PlayerFled)
	if !reflect.DeepEqual(fake.played, []int{13}) || b.messageWaitFrames != 32 {
		t.Fatalf("player flee SFX=%v wait=%d, want cue13/wait32", fake.played, b.messageWaitFrames)
	}
	for i := 0; i < 32; i++ {
		b.input(InputState{Confirm: true})
		if b.phase != phMessage {
			t.Fatalf("wait frame %d advanced phase=%d", i, b.phase)
		}
	}
	b.input(InputState{Confirm: true})
	if b.phase != phEnd {
		t.Fatalf("completed wait should allow Confirm to reach phEnd, got %d", b.phase)
	}

	b.phase, b.result = phMessage, 0
	b.emitTextWithSFX(battleTextActorFled, cues.EnemyFled)
	if !reflect.DeepEqual(fake.played, []int{13, 21}) || b.messageWaitFrames != 15 {
		t.Fatalf("enemy flee SFX=%v wait=%d, want cue21/wait15", fake.played, b.messageWaitFrames)
	}
}

func TestBattlePlayerPhysicalSFXSequenceWaitGate(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := pack.BattleSoundCues()
	fake := &fakeBattleSFX{duration: map[int]int64{
		6:  182089107,
		11: 173595589,
	}}
	b := &Battle{
		phase: phMessage, result: 3,
		texts: map[string]battleTextDefinition{
			battleTextActorAttack: {value: battleTextActorAttack},
		},
	}
	b.setSoundCues(cues, fake)
	b.emitTextWithSFXSequence(battleTextActorAttack, cues.PlayerPhysical.Steps, []int{1})
	if !reflect.DeepEqual(fake.played, []int{6}) || b.messageWaitFrames != 11 {
		t.Fatalf("first physical SFX=%v wait=%d, want cue6/wait11", fake.played, b.messageWaitFrames)
	}
	for i := 0; i < 11; i++ {
		b.input(InputState{Confirm: true})
		if b.phase != phMessage {
			t.Fatalf("first cue wait frame %d advanced phase=%d", i, b.phase)
		}
	}
	if !reflect.DeepEqual(fake.played, []int{6, 11}) || b.messageWaitFrames != 11 {
		t.Fatalf("second physical SFX=%v wait=%d, want cue11/wait11", fake.played, b.messageWaitFrames)
	}
	for i := 0; i < 11; i++ {
		b.input(InputState{Confirm: true})
		if b.phase != phMessage {
			t.Fatalf("second cue wait frame %d advanced phase=%d", i, b.phase)
		}
	}
	b.input(InputState{Confirm: true})
	if b.phase != phEnd {
		t.Fatalf("completed physical sequence should allow Confirm to reach phEnd, got %d", b.phase)
	}
}

func TestBattleCommonPhysicalSFXWaitGates(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := pack.BattleSoundCues()
	fake := &fakeBattleSFX{duration: map[int]int64{
		1: 178807947,
		3: 232404127,
		4: 177575850,
	}}
	b := &Battle{phase: phMessage, result: 3, texts: map[string]battleTextDefinition{}}
	b.setSoundCues(cues, fake)
	for _, tc := range []struct {
		role string
		cue  gamepack.BattleSoundCue
		want int
	}{
		{battleTextEnemyAttack, cues.EnemyPhysicalAttack, 11},
		{battleTextActorDamage, cues.PhysicalHit, 11},
		{battleTextActorMissed, cues.PhysicalMiss, 14},
	} {
		b.phase, b.result = phMessage, 3
		b.emitTextWithSFX(tc.role, tc.cue)
		if b.messageWaitFrames != tc.want {
			t.Fatalf("%s wait=%d, want %d", tc.role, b.messageWaitFrames, tc.want)
		}
		for i := 0; i < tc.want; i++ {
			b.input(InputState{Confirm: true})
			if b.phase != phMessage {
				t.Fatalf("%s wait frame %d advanced phase=%d", tc.role, i, b.phase)
			}
		}
		b.input(InputState{Confirm: true})
		if b.phase != phEnd {
			t.Fatalf("%s completed wait did not reach phEnd: phase=%d", tc.role, b.phase)
		}
	}
	if !reflect.DeepEqual(fake.played, []int{4, 1, 3}) {
		t.Fatalf("common physical SFX order=%v, want [4 1 3]", fake.played)
	}
}

func TestBattlePlayerPhysicalCommandsQueueAttackBeforeDamage(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	newBattle := func() *Battle {
		return &Battle{
			heroHP: 20, heroAtk: 20, actionTarget: 0, actionActor: 0, cursor: bcWar,
			enemies:    []enemyUnit{{hp: 999, def: 1}},
			companions: []*battleActor{{hp: 20, atk: 20, name: []int{3}}},
			rng:        dosrng.New(1), resolving: true,
			soundCues: pack.BattleSoundCues(),
			texts: map[string]battleTextDefinition{
				battleTextActorAttack: {value: battleTextActorAttack},
				battleTextActorDamage: {value: battleTextActorDamage},
			},
		}
	}
	leader := newBattle()
	leader.execTurn()
	if len(leader.messageQueue) < 2 || leader.messageQueue[0].role != battleTextActorAttack ||
		leader.messageQueue[1].role != battleTextActorDamage ||
		!reflect.DeepEqual(leader.messageQueue[0].sfx, leader.soundCues.PlayerPhysical.Steps) ||
		!reflect.DeepEqual(leader.messageQueue[1].sfx, []gamepack.BattleSoundCue{leader.soundCues.PhysicalHit}) {
		t.Fatalf("leader physical messages=%+v", leader.messageQueue)
	}
	companion := newBattle()
	companion.execCompanionCommand(0, battleCommand{kind: bcWar, target: 0})
	if len(companion.messageQueue) < 2 || companion.messageQueue[0].role != battleTextActorAttack ||
		companion.messageQueue[1].role != battleTextActorDamage ||
		!reflect.DeepEqual(companion.messageQueue[0].sfx, companion.soundCues.PlayerPhysical.Steps) ||
		!reflect.DeepEqual(companion.messageQueue[1].sfx, []gamepack.BattleSoundCue{companion.soundCues.PhysicalHit}) {
		t.Fatalf("companion physical messages=%+v", companion.messageQueue)
	}
}

func TestBattleCommonPhysicalResultQueuesOriginalRolesAndCues(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	cues := pack.BattleSoundCues()
	newBattle := func(heroHP, enemyHP int, seed uint16) *Battle {
		return &Battle{
			heroHP: heroHP, heroMax: heroHP, heroDef: 1, heroAtk: 99,
			actionTarget: 0, actionActor: 0, cursor: bcWar,
			enemies: []enemyUnit{
				{hp: enemyHP, max: enemyHP, atk: 99, def: 1},
				{hp: 999, max: 999, atk: 1, def: 1},
			},
			rng: dosrng.New(seed), resolving: true, soundCues: cues,
			texts: map[string]battleTextDefinition{},
		}
	}

	enemyHit := newBattle(999, 999, 1)
	enemyHit.enemyAction(0, dq3data.MonsterAI{}, false)
	if len(enemyHit.messageQueue) < 2 ||
		enemyHit.messageQueue[0].role != battleTextEnemyAttack ||
		enemyHit.messageQueue[1].role != battleTextPartyDamage ||
		!reflect.DeepEqual(enemyHit.messageQueue[0].sfx, []gamepack.BattleSoundCue{cues.EnemyPhysicalAttack}) ||
		!reflect.DeepEqual(enemyHit.messageQueue[1].sfx, []gamepack.BattleSoundCue{cues.PhysicalHit}) {
		t.Fatalf("enemy physical hit messages=%+v", enemyHit.messageQueue)
	}

	enemyDeath := newBattle(1, 999, 1)
	enemyDeath.enemyAction(0, dq3data.MonsterAI{}, false)
	if len(enemyDeath.messageQueue) < 3 || enemyDeath.messageQueue[2].role != battleTextActorDied {
		t.Fatalf("enemy lethal physical messages=%+v", enemyDeath.messageQueue)
	}
	if countMessageRole(enemyDeath.messageQueue, battleTextActorDied) != 1 {
		t.Fatalf("enemy lethal physical death count=%d", countMessageRole(enemyDeath.messageQueue, battleTextActorDied))
	}

	playerDeath := newBattle(999, 1, 1)
	playerDeath.execTurn()
	if len(playerDeath.messageQueue) < 3 || playerDeath.messageQueue[0].role != battleTextActorAttack ||
		playerDeath.messageQueue[1].role != battleTextActorDamage ||
		playerDeath.messageQueue[2].role != battleTextEnemyDefeated {
		t.Fatalf("player lethal physical messages=%+v", playerDeath.messageQueue)
	}
	if countMessageRole(playerDeath.messageQueue, battleTextEnemyDefeated) != 1 {
		t.Fatalf("player lethal physical defeat count=%d", countMessageRole(playerDeath.messageQueue, battleTextEnemyDefeated))
	}

	findMiss := func(enemy bool) *Battle {
		for seed := uint16(0); ; seed++ {
			b := newBattle(999, 999, seed)
			if enemy {
				b.enemies[0].status |= statusBlind
				b.enemyAction(0, dq3data.MonsterAI{}, false)
			} else {
				b.partyBlind = true
				b.execTurn()
			}
			if len(b.messageQueue) >= 2 && b.messageQueue[1].role == battleTextActorMissed {
				return b
			}
			if seed == 0xffff {
				t.Fatal("no deterministic blind-miss seed found")
			}
		}
	}
	for side, miss := range map[string]*Battle{"player": findMiss(false), "enemy": findMiss(true)} {
		if !reflect.DeepEqual(miss.messageQueue[1].sfx, []gamepack.BattleSoundCue{cues.PhysicalMiss}) {
			t.Fatalf("%s physical miss messages=%+v", side, miss.messageQueue)
		}
		wantAttack := battleTextActorAttack
		wantSFX := cues.PlayerPhysical.Steps
		if side == "enemy" {
			wantAttack = battleTextEnemyAttack
			wantSFX = []gamepack.BattleSoundCue{cues.EnemyPhysicalAttack}
		}
		if miss.messageQueue[0].role != wantAttack || !reflect.DeepEqual(miss.messageQueue[0].sfx, wantSFX) {
			t.Fatalf("%s physical miss attack message=%+v", side, miss.messageQueue[0])
		}
	}
}

func countMessageRole(messages []battleMessage, role string) int {
	n := 0
	for _, message := range messages {
		if message.role == role {
			n++
		}
	}
	return n
}

func TestPackMonsterPoisonActionAppliesToEveryEligiblePartyMember(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{
		heroHP:     10,
		companions: []*battleActor{{hp: 10, name: []int{1}}},
		enemies:    []enemyUnit{{hp: 10}},
		rng:        dosrng.New(0),
		resolving:  true,
	}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	seed := uint16(0)
	for ; seed < 0xffff; seed++ {
		b.rng.Seed(seed)
		if b.roll() <= 100 && b.roll() <= 100 {
			break
		}
	}
	b.rng.Seed(seed)
	if !b.executeMonsterAction(0, 38, -1) {
		t.Fatal("pack mask bit38 should consume poison-gas action")
	}
	if b.heroStatus&statusPoison == 0 || b.companions[0].status&statusPoison == 0 {
		t.Fatalf("poison action status hero=%#x companion=%#x", b.heroStatus, b.companions[0].status)
	}
	if len(b.messageQueue) != 3 || b.messageQueue[0].role != battleTextPoisonGas ||
		b.messageQueue[1].role != battleTextActorPoisoned || b.messageQueue[2].role != battleTextActorPoisoned {
		t.Fatalf("poison action message queue=%+v", b.messageQueue)
	}
}

func TestPackMonsterSpecialPhysicalQueuesOriginalSequenceAndParalyzesSurvivor(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{
		heroHP: 100, heroMax: 100, heroDef: 999,
		enemies: []enemyUnit{{hp: 10, atk: 44}},
		rng:     dosrng.New(0), resolving: true, enemyAtkPct: 100,
		soundCues: pack.BattleSoundCues(), texts: map[string]battleTextDefinition{},
	}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	if !b.executeMonsterAction(0, 3, -1) {
		t.Fatal("pack mask bit3 應執行特殊物理動作")
	}
	if b.heroHP < 68 || b.heroHP > 78 || b.heroStatus&statusPersistentParalysis == 0 {
		t.Fatalf("特殊物理結果 hp=%d status=%#x；應忽略 999 守備並使存活者麻痺", b.heroHP, b.heroStatus)
	}
	wantRoles := []string{battleTextEnemyAttack, battleTextFatalStrike, battleTextPartyDamage, battleTextActorParalyzed}
	if len(b.messageQueue) != len(wantRoles) {
		t.Fatalf("特殊物理訊息=%+v", b.messageQueue)
	}
	for i, role := range wantRoles {
		if b.messageQueue[i].role != role {
			t.Fatalf("訊息 %d role=%q，預期 %q", i, b.messageQueue[i].role, role)
		}
	}
	if !reflect.DeepEqual(b.messageQueue[0].sfx, []gamepack.BattleSoundCue{b.soundCues.EnemyPhysicalAttack}) ||
		!reflect.DeepEqual(b.messageQueue[1].sfx, []gamepack.BattleSoundCue{b.soundCues.FatalStrike}) ||
		len(b.messageQueue[2].sfx) != 0 {
		t.Fatalf("特殊物理 cue 順序=%+v", b.messageQueue)
	}
}

func TestPackMonsterSpecialPhysicalDeathDoesNotParalyze(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{heroHP: 1, heroMax: 1, enemies: []enemyUnit{{hp: 10, atk: 44}},
		rng: dosrng.New(0), resolving: true, enemyAtkPct: 100,
		soundCues: pack.BattleSoundCues(), texts: map[string]battleTextDefinition{}}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	if !b.executeMonsterAction(0, 3, -1) {
		t.Fatal("pack mask bit3 未執行")
	}
	if b.heroHP != 0 || b.heroStatus&statusPersistentParalysis != 0 ||
		countMessageRole(b.messageQueue, battleTextActorParalyzed) != 0 ||
		countMessageRole(b.messageQueue, battleTextActorDied) != 1 {
		t.Fatalf("致死特殊物理狀態／訊息錯誤：hp=%d status=%#x queue=%+v",
			b.heroHP, b.heroStatus, b.messageQueue)
	}
}

func TestPersistentParalysisSkipsCommandAndRoundAction(t *testing.T) {
	commandGate := &Battle{
		heroHP: 20, heroStatus: statusPersistentParalysis,
		companions: []*battleActor{{hp: 20, agi: 10}},
	}
	if !commandGate.seekCommandActor(0) || commandGate.commandActor != 1 {
		t.Fatalf("麻痺勇者未被 command gate 略過：actor=%d", commandGate.commandActor)
	}
	round := &Battle{
		heroHP: 20, heroStatus: statusPersistentParalysis,
		commands: []battleCommand{{kind: bcDef, target: -1}}, rng: dosrng.New(0),
	}
	round.resolveRound()
	if round.defending {
		t.Fatal("麻痺勇者在 action queue 中仍執行防禦")
	}
}

func TestPackMonsterPoisonActionSkipsDeadAndAlreadyPoisonedWithoutRoll(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{
		heroHP: 10, heroStatus: statusPoison,
		companions: []*battleActor{{hp: 0}}, enemies: []enemyUnit{{hp: 10}}, rng: dosrng.New(0),
	}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	b.rng.Seed(7)
	before := b.rng.State()
	if !b.executeMonsterAction(0, 38, -1) || b.rng.State() != before {
		t.Fatalf("ineligible targets must consume action without status rolls: before=%#x after=%#x",
			before, b.rng.State())
	}
}

func TestPackMonsterSleepBreathAppliesToEveryEligiblePartyMember(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{
		heroHP: 10,
		companions: []*battleActor{
			{hp: 10, name: []int{1}},
			{hp: 0, name: []int{2}},
			{hp: 10, status: statusParalysis, name: []int{3}},
		},
		enemies: []enemyUnit{{hp: 10}}, rng: dosrng.New(0), resolving: true,
	}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	seed := uint16(0)
	for ; seed < 0xffff; seed++ {
		b.rng.Seed(seed)
		if b.roll() <= 180 && b.roll() <= 180 {
			break
		}
	}
	b.rng.Seed(seed)
	if !b.executeMonsterAction(0, 41, -1) {
		t.Fatal("pack mask bit41 should consume sleep-breath action")
	}
	if b.heroStatus&statusParalysis == 0 {
		t.Fatalf("eligible hero should sleep, status=%#x", b.heroStatus)
	}
	if b.companions[0].status&statusParalysis == 0 || b.companions[1].status != 0 ||
		b.companions[2].status&statusParalysis == 0 {
		t.Fatalf("sleep eligibility mismatch: %+v", b.companions)
	}
	if len(b.messageQueue) != 3 || b.messageQueue[0].role != battleTextSleepBreath ||
		b.messageQueue[1].role != battleTextEnemySleep || b.messageQueue[2].role != battleTextEnemySleep {
		t.Fatalf("sleep action message queue=%+v", b.messageQueue)
	}
}

func TestOriginalMonsterSpellConsumesPerEnemyMP(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{
		mons: mons, rng: dosrng.New(1), heroHP: 9999, heroLevel: 0,
		enemies: []enemyUnit{{monID: 46, hp: 40, mp: 10}}, resolving: true,
	}
	ai := dq3data.MonsterAI{CastProb: 255, SpellMask: [6]uint8{0, 0x40, 0, 0, 0, 0}}
	for n := 0; n < 3; n++ {
		b.enemyAction(0, ai, true)
	}
	if b.enemies[0].mp != 1 {
		t.Fatalf("monster46 three cost3 spells should leave MP1, got %d", b.enemies[0].mp)
	}
	b.messageQueue = nil
	before := b.heroHP
	for seed := uint16(0); ; seed++ {
		b.rng.Seed(seed)
		if b.roll() < 255 {
			b.rng.Seed(seed)
			break
		}
	}
	b.enemyAction(0, ai, true)
	if b.enemies[0].mp != 1 || b.heroHP != before {
		t.Fatalf("insufficient monster MP must consume action without damage: mp=%d hp=%d→%d",
			b.enemies[0].mp, before, b.heroHP)
	}
	if len(b.messageQueue) != 1 || b.messageQueue[0].role != battleTextMPInsufficient {
		t.Fatalf("monster MP-short message=%+v", b.messageQueue)
	}
}

func TestOriginalMonster17MaskRunsPoisonActionInsteadOfPhysicalFallback(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	ai, ok := mons.AI(17)
	if !ok || ai.SpellMask != [6]uint8{0, 0, 0, 0, 2, 0} {
		t.Fatalf("monster17 AI=%+v ok=%v", ai, ok)
	}
	seed := uint16(0)
	var wantState uint16
	for ; seed < 0xffff; seed++ {
		r := dosrng.New(seed)
		if r.Next(256) < int(ai.CastProb) {
			r.Next(2) // sub_1AB83：cast gate 成功後、bit selection 前抽存活隊員。
			r.Next(1) // only one set bit, but the original RNG call still advances.
			if r.Next(256) <= 100 && r.Next(256) <= 100 {
				wantState = r.State()
				break
			}
		}
	}
	b := &Battle{
		mons: mons, rng: dosrng.New(seed), heroHP: 20, heroLevel: 0,
		companions: []*battleActor{{hp: 20}},
	}
	b.enemies = []enemyUnit{{monID: 17, hp: 20, atk: 255}}
	b.setMonsterActions(pack.MonsterActionDefinitions())
	b.enemyAction(0, ai, true)
	if got := b.rng.State(); got != wantState {
		t.Fatalf("poison all-party action must retain gate→target→bit→per-member RNG order: got %#x want %#x",
			got, wantState)
	}
	if b.heroHP != 20 || b.companions[0].hp != 20 ||
		b.heroStatus&statusPoison == 0 || b.companions[0].status&statusPoison == 0 {
		t.Fatalf("monster17 should poison without physical damage: heroHP=%d compHP=%d status=%#x/%#x",
			b.heroHP, b.companions[0].hp, b.heroStatus, b.companions[0].status)
	}
}

func TestBattleEndPersistsPoisonToHeroAndCompanion(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	member := newMember([]int{1}, 3, 0, 0)
	g := &Game{pack: pack, companions: []*Member{member}, flags: map[int]bool{}}
	g.battle.heroHP, g.battle.heroMP, g.battle.heroStatus = 10, 2, statusPoison
	g.battle.companions = []*battleActor{{hp: 9, mp: 1, status: statusPoison}}
	g.onBattleEnd()
	if g.heroConditions&conditionPoison == 0 || member.Conditions&conditionPoison == 0 {
		t.Fatalf("battle poison not persisted hero=%#x companion=%#x",
			g.heroConditions, member.Conditions)
	}
}

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
	wantQueuedRoles := []string{
		battleTextActorDamage,
		battleTextEnemyDefeated,
		battleTextVictory,
		battleTextExpReward,
		battleTextGoldReward,
	}
	gotQueuedRoles := make([]string, len(b.messageQueue))
	for i := range b.messageQueue {
		gotQueuedRoles[i] = b.messageQueue[i].role
	}
	if !reflect.DeepEqual(gotQueuedRoles, wantQueuedRoles) {
		t.Fatalf("勝利訊息佇列=%v，want %v", gotQueuedRoles, wantQueuedRoles)
	}
	first := b.msg
	b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	if b.phase != phMessage || b.msg == first || b.msg == "" {
		t.Fatalf("第一次確認應由攻擊訊息前進到傷害訊息: phase=%d first=%q msg=%q", b.phase, first, b.msg)
	}
	for i := 0; i < len(wantQueuedRoles) && b.phase == phMessage; i++ {
		b.input(InputState{Confirm: true, DirHeld: -1, DirEdge: -1})
	}
	if b.phase != phEnd {
		t.Fatalf("所有勝利／經驗／金錢訊息確認完才可進 phEnd，got %d", b.phase)
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
