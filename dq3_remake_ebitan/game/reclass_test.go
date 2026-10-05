package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func testReclassEvent(t *testing.T) (*gamepack.Pack, *gamepack.ReclassEvent) {
	t.Helper()
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("BuiltinDQ3: %v", err)
	}
	event, ok := pack.ReclassEvent("dq3:event.dhama_reclass")
	if !ok {
		t.Fatal("missing dq3:event.dhama_reclass")
	}
	return pack, event
}

func reclassMember(class, level int) *Member {
	m := newMember([]int{700}, class, 0, stats.ExpForLevel(class, level))
	m.Stats = stats.Values{
		stats.STR: 21, stats.VIT: 23, stats.AGI: 25,
		stats.HP: 101, stats.MP: 81, stats.INT: 27, stats.LUCK: 29,
	}
	m.CurHP, m.CurMP = 99, 79
	setTestGear(&m.Items, 0, 1)
	setTestGear(&m.Items, 1, 30)
	setTestGear(&m.Items, 2, 0x39)
	setTestGear(&m.Items, 3, 0x32)
	m.syncLearnedSpells()
	return m
}

func TestReclassOriginalGatesAreMemberLocal(t *testing.T) {
	pack, event := testReclassEvent(t)
	mage := reclassMember(4, 20)
	g := &Game{
		pack: pack, companions: []*Member{mage},
		reclassMember: 1, reclassTarget: event.AdvancedTargetClass,
	}
	if containsInt(g.reclassTargets(event, mage), event.AdvancedTargetClass) {
		t.Fatal("魔法使未親自持有領悟之書時，不得選賢者")
	}
	setTestInventory(&g.items, []int{event.AdvancedRequiredItemRaw})
	if containsInt(g.reclassTargets(event, mage), event.AdvancedTargetClass) {
		t.Fatal("全隊／主角背包中的領悟之書不得替同伴開啟賢者")
	}
	setTestInventory(&mage.Items, []int{event.AdvancedRequiredItemRaw})
	if !containsInt(g.reclassTargets(event, mage), event.AdvancedTargetClass) {
		t.Fatal("該隊員親自持有領悟之書時應開啟賢者")
	}

	jester := reclassMember(event.AdvancedFreeSourceClass, 20)
	if !containsInt(g.reclassTargets(event, jester), event.AdvancedTargetClass) {
		t.Fatal("遊玩者不需領悟之書即可選賢者")
	}

	underLevel := reclassMember(4, 19)
	before := *underLevel
	g.companions, g.reclassMember = []*Member{underLevel}, 1
	if g.applyReclass(event) {
		t.Fatal("Lv19 不得轉職")
	}
	if !reflect.DeepEqual(before, *underLevel) {
		t.Fatal("gate 失敗不得消耗道具或修改角色")
	}
	g.reclassMember = 0
	if g.applyReclass(event) {
		t.Fatal("勇者不得轉職")
	}
}

func TestReclassOriginalTransactionPreservesVITAndSpells(t *testing.T) {
	pack, event := testReclassEvent(t)
	m := reclassMember(4, 20)
	oldKnown := m.AllSpells()
	if !containsInt(oldKnown, 131) {
		t.Fatal("fixture 魔法使 Lv20 應已學會 rec131")
	}
	setTestInventory(&m.Items, []int{event.AdvancedRequiredItemRaw, 9})
	g := &Game{
		pack: pack, companions: []*Member{m},
		reclassMember: 1, reclassTarget: event.AdvancedTargetClass,
	}
	if !g.applyReclass(event) {
		t.Fatal("持書 Lv20 魔法使轉賢者應成功")
	}
	if m.Class != 5 || m.Level() != 1 || m.Exp != 0 {
		t.Fatalf("職業／等級／EXP transaction 錯誤：class%d lv%d exp%d",
			m.Class, m.Level(), m.Exp)
	}
	if m.CurHP != 49 || m.CurMP != 39 ||
		m.Stats[stats.STR] != 10 || m.Stats[stats.AGI] != 12 ||
		m.Stats[stats.HP] != 50 || m.Stats[stats.MP] != 40 ||
		m.Stats[stats.INT] != 13 || m.Stats[stats.LUCK] != 14 {
		t.Fatalf("原版減半欄位錯誤：hp/mp=%d/%d stats=%v", m.CurHP, m.CurMP, m.Stats)
	}
	if m.Stats[stats.VIT] != 23 {
		t.Fatalf("原版不減半 VIT，got %d", m.Stats[stats.VIT])
	}
	if m.Items.Equipment()[0] != -1 || m.Items.Equipment()[1] != -1 || m.Items.Equipment()[2] != -1 || m.Items.Equipment()[3] != -1 {
		t.Fatalf("轉賢者後必須全卸裝：%d/%d/%d/%d",
			m.Items.Equipment()[0], m.Items.Equipment()[1], m.Items.Equipment()[2], m.Items.Equipment()[3])
	}
	if !reflect.DeepEqual(testInventory(m.Items), []int{9, 1, 30, 0x39, 0x32}) {
		t.Fatalf("只消耗領悟之書並保留卸下裝備，got %v", testInventory(m.Items))
	}
	for _, rec := range []int{131, 161} {
		if !containsInt(m.AllSpells(), rec) {
			t.Fatalf("轉職後應保留舊咒文並學會賢者 Lv1 咒文 rec%d，got %v",
				rec, m.AllSpells())
		}
	}
	if !spell.BattleUsable(131) || !containsInt(m.Spells(), 131) {
		t.Fatal("保留的戰鬥咒文仍應出現在戰鬥選單")
	}
	g.shop.items = loadTestItems(t)
	g.panelActor, g.panelCursor = 1, 2 // 卸回個人物品欄的布衣 0x1e
	g.equipSelected()
	if m.Items.Equipment()[1] != 30 || containsInt(testInventory(m.Items), 30) {
		t.Fatalf("轉職卸裝後應能由同伴個人物品重新裝備：armor=%d inv=%v",
			m.Items.Equipment()[1], testInventory(m.Items))
	}
}

func TestReclassPersonalStateSaveRoundTrip(t *testing.T) {
	m := reclassMember(5, 1)
	setTestInventory(&m.Items, []int{9, 30, 31})
	m.LearnedSpells = []int{121, 131, 161}
	g := &Game{pack: loadTestPack(t), items: testItemStore(nil, [4]int{-1, -1, -1, -1}), companions: []*Member{m}}
	s := g.snapshot()
	raw, err := encodeSave(s)
	if err != nil {
		t.Fatal(err)
	}
	got, err := decodeSave(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.validateSavedItemStores(&got); err != nil {
		t.Fatal(err)
	}
	if len(got.Comps) != 1 ||
		!reflect.DeepEqual(testInventory(got.Comps[0].itemStore), testInventory(m.Items)) ||
		!reflect.DeepEqual(got.Comps[0].LearnedSpells, m.LearnedSpells) {
		t.Fatalf("個人物品／永久咒文存檔 round-trip 失敗：%+v", got.Comps)
	}
}

func TestGiveItemUsesOriginalPersonalEightSlotCapacity(t *testing.T) {
	pack, _ := testReclassEvent(t)
	m := reclassMember(4, 20)
	setTestGear(&m.Items, 0, 1)
	setTestGear(&m.Items, 1, 30)
	setTestGear(&m.Items, 2, -1)
	setTestGear(&m.Items, 3, -1)
	setTestInventory(&m.Items, []int{2, 3, 4, 5, 6, 7})
	g := &Game{pack: pack, companions: []*Member{m}, itemSelected: 0, panelActor: 0, items: testItemStore([]int{0x4a}, [4]int{-1, -1, -1, -1})}
	if g.giveSelectedItem(1) {
		t.Fatal("個人物品八格已滿時，給予必須失敗")
	}
	if !reflect.DeepEqual(testInventory(g.items), []int{0x4a}) || len(testInventory(m.Items)) != 6 {
		t.Fatal("給予失敗不得消耗或複製道具")
	}
	setTestInventory(&m.Items, testInventory(m.Items)[:5])
	if !g.giveSelectedItem(1) {
		t.Fatal("個人物品有一格空位時，給予應成功")
	}
	if len(testInventory(g.items)) != 0 || !containsInt(testInventory(m.Items), 0x4a) {
		t.Fatalf("給予交易錯誤：hero=%v member=%v", testInventory(g.items), testInventory(m.Items))
	}
}
