package game

import (
	"strings"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func palpunteBattle(t *testing.T, monID, count int) *Battle {
	t.Helper()
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}
	hero := heroParams{
		level: 40, curHP: 20, maxHP: 200, atk: 1, def: 999, agi: 999,
		mp: 100, maxMP: 100, spells: []int{180},
	}
	if !b.startGroup(monID, count, 1, hero, []*battleActor{
		{hp: 30, maxHP: 200, mp: 20, maxMP: 100},
	}) {
		t.Fatalf("開戰失敗(mon%d)", monID)
	}
	return b
}

func TestPalpunteTableHasOnlyFiveEffectiveSlots(t *testing.T) {
	effective := map[int]bool{0: true, 2: true, 8: true, 11: true, 12: true}
	for effect := 0; effect < 16; effect++ {
		if effective[effect] {
			continue
		}
		b := palpunteBattle(t, 5, 1)
		hp, mp := b.heroHP, b.heroMP
		b.execPalpunte(effect)
		if b.heroHP != hp || b.heroMP != mp || !strings.Contains(b.msg, "沒有效") {
			t.Fatalf("table[%d] 應為 NULL/無效：hp=%d mp=%d msg=%q", effect, b.heroHP, b.heroMP, b.msg)
		}
	}
}

func TestPalpunteHealSlotsRestoreEveryLivingMember(t *testing.T) {
	for _, effect := range []int{2, 8} {
		b := palpunteBattle(t, 5, 1)
		b.execPalpunte(effect)
		if b.heroHP < 105 || b.heroHP > 114 {
			t.Fatalf("effect%d 隊長應回復85..94，HP=%d", effect, b.heroHP)
		}
		if got := b.companions[0].hp; got < 115 || got > 124 {
			t.Fatalf("effect%d 同伴應獨立回復85..94，HP=%d", effect, got)
		}
	}
}

func TestPalpunteBlowAwayGivesNoKillCredit(t *testing.T) {
	// 找到能命中史萊姆 packed chance=255 的 deterministic RNG 狀態。
	var b *Battle
	for seed := int64(1); seed < 1000; seed++ {
		candidate := palpunteBattle(t, 5, 1)
		candidate.rng.Seed(uint16(seed))
		candidate.execPalpunte(11)
		if candidate.enemies[0].fled {
			b = candidate
			break
		}
	}
	if b == nil {
		t.Fatal("未找到可命中吹走效果的 seed")
	}
	if b.enemies[0].hp != 0 || !b.enemies[0].fled || b.killedCount() != 0 {
		t.Fatalf("吹走應移除敵人且不計擊殺：%+v killed=%d", b.enemies[0], b.killedCount())
	}
}

func TestPalpunteSleepCanAffectBothSides(t *testing.T) {
	var b *Battle
	for seed := int64(1); seed < 1000; seed++ {
		candidate := palpunteBattle(t, 5, 1)
		candidate.rng.Seed(uint16(seed))
		candidate.execPalpunte(0)
		if candidate.enemies[0].status&statusParalysis != 0 &&
			candidate.heroStatus&statusParalysis != 0 {
			b = candidate
			break
		}
	}
	if b == nil {
		t.Fatal("未找到敵我雙方皆睡眠的 deterministic seed")
	}
	if b.companions[0].hp <= 0 {
		t.Fatal("測試前提：同伴應存活")
	}
}

func TestPalpunteDrainTransfersEnemyMPWithoutMaxCap(t *testing.T) {
	// mon2 的原始 +0x04 MP=12、Palpunte packed chance=255。
	var b *Battle
	for seed := int64(1); seed < 1000; seed++ {
		candidate := palpunteBattle(t, 2, 1)
		candidate.heroMP, candidate.heroMaxMP = 100, 100
		candidate.rng.Seed(uint16(seed))
		candidate.execPalpunte(12)
		if candidate.enemies[0].mp < 12 {
			b = candidate
			break
		}
	}
	if b == nil {
		t.Fatal("未找到成功吸取 MP 的 deterministic seed")
	}
	drained := 12 - b.enemies[0].mp
	if drained < 1 || drained > 9 {
		t.Fatalf("原版 min(10,MP) RNG 將0改1，應吸1..9，得 %d", drained)
	}
	if b.heroMP != 100+drained {
		t.Fatalf("吸取量應直接加到施法者且不封頂：heroMP=%d drained=%d", b.heroMP, drained)
	}
}

func TestExecPalpunteCostsTwentyMP(t *testing.T) {
	b := palpunteBattle(t, 5, 1)
	b.heroHP, b.heroMax = 200, 200 // 無論抽到哪格，避免治療分支改變斷言焦點
	b.execSpell(180)
	if b.heroMP != 80 {
		t.Fatalf("帕魯朋特應扣20 MP，得 %d", b.heroMP)
	}
}

func TestCompanionPalpunteUsesCompanionMP(t *testing.T) {
	b := palpunteBattle(t, 5, 1)
	b.heroMP = 77
	b.companions[0].mp = 100
	b.enemies[0].mp = 0 // 即使抽到吸 MP，也不改變施法者結果
	b.actionActor = 1
	b.resolving = true
	b.execSpell(180)
	b.resolving = false
	if b.heroMP != 77 || b.companions[0].mp != 80 {
		t.Fatalf("同伴施放巴魯朋特應只扣同伴20 MP，hero=%d companion=%d",
			b.heroMP, b.companions[0].mp)
	}
}
