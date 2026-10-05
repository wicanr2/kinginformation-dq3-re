package game

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func fillActorInventoryToCapacity(t *testing.T, g *Game, actor, code int) {
	t.Helper()
	store := g.actorItemStore(actor)
	if store == nil {
		t.Fatal("fixture owner unavailable")
	}
	for len(store.Entries()) < g.pack.ItemActions().PersonalInventorySlots {
		if _, ok := store.Add(code); !ok {
			t.Fatal("fixture cannot fill owner")
		}
	}

}

func TestBattleDropUsesFirstPartyPersonalInventoryWithSpace(t *testing.T) {
	g := bossTestGame(t)
	g.companions = []*Member{newMember([]int{1}, 1, 0, 0)}
	fillActorInventoryToCapacity(t, g, 0, 0x41)
	beforeHero := append([]int(nil), testInventory(g.items)...)
	beforeGold := g.heroGold

	g.battle.active = false
	g.battle.monID = 1
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 0, 7, 0x44
	g.onBattleEnd()

	if !bytes.Equal(intSliceBytes(testInventory(g.items)), intSliceBytes(beforeHero)) {
		t.Fatalf("隊長滿格時不得修改其物品欄：before=%v after=%v", beforeHero, testInventory(g.items))
	}
	if got := testInventory(g.companions[0].Items); len(got) != 1 || got[0] != 0x44 {
		t.Fatalf("掉落未寫入第一個有空位的同伴：%v", got)
	}
	if g.heroGold != beforeGold+7 || g.noticeCode != 0x44 || g.noticeTimer == 0 {
		t.Fatalf("成功掉落結算錯：gold=%d notice=%#x/%d", g.heroGold, g.noticeCode, g.noticeTimer)
	}
}

func TestBattleDropAllPartyFullDoesNotMutateInventory(t *testing.T) {
	g := bossTestGame(t)
	g.companions = []*Member{newMember([]int{1}, 1, 0, 0)}
	fillActorInventoryToCapacity(t, g, 0, 0x41)
	fillActorInventoryToCapacity(t, g, 1, 0x41)
	beforeHero := append([]int(nil), testInventory(g.items)...)
	beforeCompanion := append([]int(nil), testInventory(g.companions[0].Items)...)
	beforeGold := g.heroGold
	g.noticeCode, g.noticeTimer = -1, 0

	g.battle.active = false
	g.battle.monID = 1
	g.battle.result, g.battle.gotExp, g.battle.gotGold, g.battle.gotDrop = 1, 0, 7, 0x44
	g.onBattleEnd()

	if !bytes.Equal(intSliceBytes(testInventory(g.items)), intSliceBytes(beforeHero)) ||
		!bytes.Equal(intSliceBytes(testInventory(g.companions[0].Items)), intSliceBytes(beforeCompanion)) {
		t.Fatalf("全隊滿格時掉落不得修改物品欄：hero=%v companion=%v",
			testInventory(g.items), testInventory(g.companions[0].Items))
	}
	if g.heroGold != beforeGold+7 || g.noticeCode != -1 || g.noticeTimer != 0 {
		t.Fatalf("滿格仍須結算金錢且不得冒稱取得物品：gold=%d notice=%#x/%d",
			g.heroGold, g.noticeCode, g.noticeTimer)
	}
	want, ok := g.pack.TextGlyphCodes("common:text.battle.drop.inventory_full")
	if !ok || !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, want) {
		t.Fatal("全隊滿格時未顯示 pack-owned record345")
	}
}

func TestBattleDropInventoryWriterMatchesOriginal(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join(spineAssetsDir(t), "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []struct {
		file int
		hex  []byte
	}{
		{0xd879, []byte{0xa0, 0x21, 0x23}},                   // first formation monster
		{0xd897, []byte{0x8a, 0x87, 0x9e, 0x0d}},             // D3MNS +0x26 item
		{0xd8a7, []byte{0xe8, 0x14, 0xa3}},                   // call sub_1684E
		{0xd8c4, []byte{0xc7, 0x06, 0x9c, 0x25, 0x01, 0x00}}, // full: DS:259C=1 (leader name)
		{0xd8ca, []byte{0xbf, 0x59, 0x01}},                   // full: record 0x159 / 345
		{0x7be8, []byte{0x80, 0xfa, 0x08}},                   // eight personal words
		{0x7bf5, []byte{0xc6, 0x06, 0x26, 0x07, 0x01}},       // all party full
		{0x7bfc, []byte{0xc6, 0x06, 0x26, 0x07, 0x00}},       // found free slot
		{0x7c05, []byte{0x89, 0x04}},                         // write item word
	} {
		if want.file < 0 || want.file+len(want.hex) > len(raw) ||
			!bytes.Equal(raw[want.file:want.file+len(want.hex)], want.hex) {
			t.Fatalf("DQ3.EXE file %#x bytes mismatch", want.file)
		}
	}
}

func intSliceBytes(values []int) []byte {
	out := make([]byte, len(values)*2)
	for i, value := range values {
		out[i*2], out[i*2+1] = byte(value), byte(value>>8)
	}
	return out
}
