package game

import (
	"bytes"
	"encoding/json"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"testing"
)

func TestOrderedItemsAllSaveOwnersPreservePhysicalWords(t *testing.T) {
	g := bossTestGame(t)
	words := []uint16{0x801e, 0xff, 0xc01b, 0, 0x2041, 0xff, 0x4a, 0xff}
	store, err := g.pack.ItemStoreFromWords(words)
	if err != nil {
		t.Fatal(err)
	}
	makeMember := func() *Member { m := newMember([]int{1}, 1, 0, 0); m.Items = store.Clone(); return m }
	g.items = store.Clone()
	g.companions = []*Member{makeMember()}
	g.roster = []*Member{makeMember()}
	g.soloChallengeCompanions = []*Member{makeMember()}
	g.settlementFounder = makeMember()
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "all-owners.json"))
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*itemstore.Store{&g.items, &g.companions[0].Items, &g.roster[0].Items, &g.soloChallengeCompanions[0].Items, &g.settlementFounder.Items} {
		s.Remove(2)
	}
	if err := g.Load(); err != nil {
		t.Fatal(err)
	}
	for _, s := range []*itemstore.Store{&g.items, &g.companions[0].Items, &g.roster[0].Items, &g.soloChallengeCompanions[0].Items, &g.settlementFounder.Items} {
		if !reflect.DeepEqual(s.Words(), words) {
			t.Fatalf("owner lost physical words: %04x", s.Words())
		}
	}
	g.companions[0].Items.Remove(2)
	if !reflect.DeepEqual(g.items.Words(), words) || !reflect.DeepEqual(g.roster[0].Items.Words(), words) {
		t.Fatal("loaded owners share writable words")
	}
	raw, err := os.ReadFile(savePath())
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"inv":`, `"eq":`, `"equipment_v2":`, `"Weapon":`, `"Inventory":`} {
		if bytes.Contains(raw, []byte(key)) {
			t.Fatal("legacy owner truth serialized", key)
		}
	}
}

func TestOrderedItemSaveRejectsOldAndCorruptOwnersBeforeMutation(t *testing.T) {
	g := bossTestGame(t)
	m := newMember([]int{1}, 1, 0, 0)
	g.companions = []*Member{m}
	g.roster = []*Member{m}
	g.soloChallengeCompanions = []*Member{m}
	g.settlementFounder = m
	raw, err := encodeSave(g.snapshot())
	if err != nil {
		t.Fatal(err)
	}
	var valid map[string]json.RawMessage
	if json.Unmarshal(raw, &valid) != nil {
		t.Fatal("fixture snapshot")
	}
	cases := map[string][]byte{"unversioned": []byte(`{"inv":[],"eq":[0,30,0,0]}`), "old_version": bytes.Replace(raw, []byte(`"save_version":2`), []byte(`"save_version":1`), 1), "duplicate_version": append([]byte(`{"save_version":2,`), raw[1:]...), "case_version": bytes.Replace(raw, []byte(`"save_version"`), []byte(`"Save_version"`), 1), "trailing": append(append([]byte(nil), raw...), []byte(` {}`)...)}
	for _, field := range []string{"pack_id", "pack_schema", "pack_content_hash", "items"} {
		copyMap := map[string]json.RawMessage{}
		for k, v := range valid {
			copyMap[k] = v
		}
		delete(copyMap, field)
		bad, _ := json.Marshal(copyMap)
		cases["missing_"+field] = bad
	}
	for _, field := range []string{"items", "comps", "roster", "solo_challenge_companions", "settlement_founder"} {
		copyMap := map[string]json.RawMessage{}
		for k, v := range valid {
			copyMap[k] = v
		}
		broken := json.RawMessage(`{"storage_version":1,"words":[255]}`)
		if field == "items" {
			copyMap[field] = broken
		} else if field == "settlement_founder" {
			var member map[string]json.RawMessage
			json.Unmarshal(copyMap[field], &member)
			member["items"] = broken
			copyMap[field], _ = json.Marshal(member)
		} else {
			var members []map[string]json.RawMessage
			json.Unmarshal(copyMap[field], &members)
			members[0]["items"] = broken
			copyMap[field], _ = json.Marshal(members)
		}
		bad, _ := json.Marshal(copyMap)
		cases["broken_"+field] = bad
	}
	before, seed := g.snapshot(), g.prng.State()
	path := filepath.Join(t.TempDir(), "bad-save.json")
	for name, bad := range cases {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, bad, 0600); err != nil {
				t.Fatal(err)
			}
			if err := g.loadFrom(path, false); err == nil {
				t.Fatal("invalid save accepted")
			}
			if !equalFieldSave(before, g.snapshot()) || seed != g.prng.State() {
				t.Fatal("invalid save mutated game before full validation")
			}
		})
	}
}

func TestOrderedItemDropRejectsFlagsMetadataAndDeadOwner(t *testing.T) {
	g := bossTestGame(t)
	words := []uint16{0x801e, 0x201f, 0x401f, 0x1d, 0x41, 0xff, 0xff, 0xff}
	store, err := g.pack.ItemStoreFromWords(words)
	if err != nil {
		t.Fatal(err)
	}
	g.items = store
	g.panelActor = 0
	for _, position := range []int{0, 1, 2, 3} {
		g.itemSelected = position
		if g.dropSelectedItem() {
			t.Fatal("blocked drop accepted", position)
		}
		if !reflect.DeepEqual(g.items.Words(), words) {
			t.Fatal("blocked drop changed words")
		}
	}
	g.heroHP = 0
	g.itemSelected = 4
	if g.dropSelectedItem() {
		t.Fatal("dead owner dropped item")
	}
	g.heroHP = 1
	if !g.dropSelectedItem() {
		t.Fatal("legal selected drop rejected")
	}
	words[4] = 0xff
	if !reflect.DeepEqual(g.items.Words(), words) {
		t.Fatal("drop moved another physical word")
	}
}

func TestReclassConditionalItemsMatchBoundedOriginalWriter(t *testing.T) {
	pack, event := testReclassEvent(t)
	before := []uint16{0x801e, 0xc04a, 0x204a, 0, 0xff, 0xa003, 0x404a, 0x401f}
	for _, target := range append(append([]int(nil), event.BasicTargetClasses...), event.AdvancedTargetClass) {
		t.Run(string(rune('0'+target)), func(t *testing.T) {
			m := reclassMember(event.AdvancedFreeSourceClass, 20)
			if m.Class == target {
				m.Class = 4
			}
			store, err := pack.ItemStoreFromWords(before)
			if err != nil {
				t.Fatal(err)
			}
			m.Items = store
			m.Exp = stats.ExpForLevel(m.Class, 20)
			g := &Game{pack: pack, companions: []*Member{m}, reclassMember: 1, reclassTarget: target}
			if !g.applyReclass(event) {
				t.Fatal("reviewed transition rejected")
			}
			want := append([]uint16(nil), before...)
			if target == event.AdvancedTargetClass {
				for i, w := range want {
					if w&0xff == uint16(event.AdvancedRequiredItemRaw) {
						want[i] = 0xff
					} else {
						want[i] &= 0xff
					}
				}
			}
			if !reflect.DeepEqual(m.Items.Words(), want) {
				t.Fatalf("conditional writer differs: %04x/%04x", m.Items.Words(), want)
			}
		})
	}
}

// Historical fixtures install an explicit validated test contract. These
// helpers never enter the runtime and are not compatibility owner fields.
var testItemPackOnce sync.Once
var testItemPack *gamepack.Pack

func fixtureItemPack() *gamepack.Pack {
	testItemPackOnce.Do(func() {
		var err error
		testItemPack, err = gamepack.BuiltinDQ3()
		if err != nil {
			panic(err)
		}
	})
	return testItemPack
}
func testItemStore(inventory []int, equipment [4]int) itemstore.Store {
	store, ok := fixtureItemPack().NewGamePlayerItems()
	if !ok {
		panic("fixture contract")
	}
	for _, entry := range store.Entries() {
		store.Remove(entry.Position)
	}
	for _, code := range inventory {
		if _, ok := store.Add(code); !ok {
			panic("fixture inventory exceeds validated contract")
		}
	}
	for _, code := range equipment {
		if code >= 0 {
			position, ok := store.Add(code)
			if !ok || !store.Wear(position) {
				panic("fixture equipment exceeds validated contract")
			}
		}
	}
	return store
}
func testInventory(store itemstore.Store) []int {
	var out []int
	for _, entry := range store.Inventory() {
		out = append(out, entry.Code)
	}
	return out
}
func testEquipmentCodes(store itemstore.Store) []int {
	eq := store.Equipment()
	return append([]int(nil), eq[:]...)
}
func setTestInventory(store *itemstore.Store, codes []int) {
	*store = testItemStore(codes, store.Equipment())
}
func setTestEquipment(store *itemstore.Store, equipment [4]int) {
	*store = testItemStore(testInventory(*store), equipment)
}
func setTestGear(store *itemstore.Store, part, code int) {
	eq := store.Equipment()
	eq[part] = code
	setTestEquipment(store, eq)
}
func testBattleItems(entries []battleItemSlot) itemstore.Store {
	store := testItemStore(nil, [4]int{-1, -1, -1, -1})
	for _, entry := range entries {
		position, ok := store.Add(entry.rawID)
		if !ok {
			panic("fixture battle items")
		}
		if entry.equipped && !store.Wear(position) {
			panic("fixture battle wear")
		}
	}
	return store
}
func unequippedBattleItemIDs(store itemstore.Store) []int { return testInventory(store) }

func newMember(name []int, class, gender int, exp uint32) *Member {
	m := &Member{Name: name, Class: class, Gender: gender, Exp: exp, Items: testItemStore(nil, [4]int{-1, -1, -1, -1})}
	m.ensureStats()
	m.syncLearnedSpells()
	m.fullHeal()
	return m
}
func startingCompanions(exp uint32) []*Member {
	return []*Member{newMember(classNames[1], 1, 0, exp), newMember(classNames[3], 3, 0, exp), newMember(classNames[4], 4, 1, exp)}
}
