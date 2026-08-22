package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	dosrng "github.com/wicanr2/dq3_remake_ebitan/internal/rng"
)

func battleItemTestDefinitions(t *testing.T) []gamepack.BattleItemDefinition {
	t.Helper()
	p := loadTestPack(t)
	return p.BattleItemDefinitions()
}

func TestBattleItemUsesSelectedOwnerSlotAndTarget(t *testing.T) {
	b := &Battle{
		heroHP: 10, heroMax: 100, rng: dosrng.New(1),
		heroItems: []battleItemSlot{{rawID: 0x55}, {rawID: 0x41}},
		companions: []*battleActor{{hp: 5, maxHP: 80,
			items: []battleItemSlot{{rawID: 0x42}}}},
	}
	b.setBattleItems(battleItemTestDefinitions(t))
	b.execBattleItem(0, battleCommand{kind: bcItem, itemRaw: 0x41, itemSlot: 1, target: 1})
	if b.companions[0].hp < 35 || b.companions[0].hp > 44 {
		t.Fatalf("藥草應依原版 30..39 回復選定同伴：hp=%d", b.companions[0].hp)
	}
	if got := unequippedBattleItemIDs(b.heroItems); !reflect.DeepEqual(got, []int{0x55}) {
		t.Fatalf("只可消耗下令者的選定格位：%v", got)
	}
	if got := unequippedBattleItemIDs(b.companions[0].items); !reflect.DeepEqual(got, []int{0x42}) {
		t.Fatalf("不得誤改目標的物品欄：%v", got)
	}
}

func TestBattleItemMenuPreservesPerActorEightSlotSource(t *testing.T) {
	b := &Battle{
		phase: phCommand, commandActor: 1,
		companions: []*battleActor{{hp: 10, maxHP: 10,
			items: []battleItemSlot{{rawID: 0x42}, {rawID: 0x45}}}},
	}
	b.setBattleItems(battleItemTestDefinitions(t))
	for i, kind := range b.commandMenu(1) {
		if kind == bcItem {
			b.cursor = i
		}
	}
	b.input(InputState{Confirm: true})
	if b.phase != phItem || len(b.actorItems(1)) != 2 {
		t.Fatalf("道具命令應開目前 actor 的個人物品選單：phase=%d items=%v", b.phase, b.actorItems(1))
	}
	b.input(InputState{DirEdge: 0})
	b.input(InputState{Confirm: true})
	if b.pending.itemRaw != 0x45 || b.pending.itemSlot != 1 || b.phase != phTargetAlly {
		t.Fatalf("選定格位／target route 錯：pending=%+v phase=%d", b.pending, b.phase)
	}
}

func TestOnBattleEndPersistsEachActorsRemainingPersonalItems(t *testing.T) {
	g := bossTestGame(t)
	g.companions = []*Member{newMember([]int{1}, 1, 0, 0)}
	g.inventory = []int{0x41, 0x55}
	g.companions[0].Inventory = []int{0x42, 0x45}
	g.battle.itemsReady = true
	g.battle.heroItems = []battleItemSlot{{rawID: 0x55}, {rawID: g.equip[0], equipped: true}}
	g.battle.companions = []*battleActor{{items: []battleItemSlot{{rawID: 0x45}}, hp: 1}}
	g.battle.result, g.battle.gotDrop = 0, -1
	g.onBattleEnd()
	if !reflect.DeepEqual(g.inventory, []int{0x55}) ||
		!reflect.DeepEqual(g.companions[0].Inventory, []int{0x45}) {
		t.Fatalf("戰後須逐 owner 寫回：hero=%v companion=%v", g.inventory, g.companions[0].Inventory)
	}
}

func TestBattlePrayerRingUsesExactRangeAndConditionalBreak(t *testing.T) {
	b := &Battle{heroMP: 0, heroMaxMP: 100, rng: dosrng.New(1),
		heroItems: []battleItemSlot{{rawID: 0x48}}}
	b.setBattleItems(battleItemTestDefinitions(t))
	b.execBattleItem(0, battleCommand{kind: bcItem, itemRaw: 0x48, itemSlot: 0, target: 0})
	if b.heroMP < 22 || b.heroMP > 31 {
		t.Fatalf("祈禱之戒 MP 應落在原版 22..31：%d", b.heroMP)
	}
	if len(b.heroItems) > 1 {
		t.Fatalf("條件損壞不可複製道具：%v", b.heroItems)
	}
}

func TestBattleChimeraWingSetsReturnTransaction(t *testing.T) {
	b := &Battle{heroHP: 1, heroMax: 1, rng: dosrng.New(1),
		heroItems: []battleItemSlot{{rawID: 0x43}}}
	b.setBattleItems(battleItemTestDefinitions(t))
	b.execBattleItem(0, battleCommand{kind: bcItem, itemRaw: 0x43, itemSlot: 0, target: 0})
	if b.result != 3 || !b.returnTown || len(b.heroItems) != 0 {
		t.Fatalf("蓋美拉翅膀應先消耗並交付戰後回城交易：result=%d return=%v items=%v",
			b.result, b.returnTown, b.heroItems)
	}
}
