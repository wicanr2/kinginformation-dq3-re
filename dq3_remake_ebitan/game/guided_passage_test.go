package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func norudTestGame(t *testing.T) (*Game, gamepack.GuidedPassageEvent, *npcInst) {
	t.Helper()
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	events := g.pack.GuidedPassageEvents()
	if len(events) != 2 {
		t.Fatalf("guided passage events=%d, want CTY62/63 variants", len(events))
	}
	event := events[0]
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		event.GuideNPC.CTYRaw, mapBlkNum[event.GuideNPC.CTYRaw],
		event.GuideNPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	g.inTown, g.curCty = true, event.GuideNPC.CTYRaw
	g.town, g.cur = sc, sc
	g.dlg.tx = sc.dlgText
	g.px, g.py = event.InteractionTile.X, event.InteractionTile.Y
	g.facing = 1
	for i := range sc.npcs {
		n := &sc.npcs[i]
		if n.x == event.GuideNPC.Tile.X && n.y == event.GuideNPC.Tile.Y &&
			n.b4 == event.GuideNPC.HandlerRaw {
			return g, event, n
		}
	}
	t.Fatalf("CTY%d sec%d 找不到原版 guide NPC", event.GuideNPC.CTYRaw, event.GuideNPC.Section)
	return nil, gamepack.GuidedPassageEvent{}, nil
}

func closeGuidedDialogue(g *Game) {
	g.dlg.open = false
	g.advanceGuidedPassageDialogue()
}

func runGuidedAnimation(t *testing.T, g *Game, wantStage int) {
	t.Helper()
	for frames := 0; frames < 2000 && g.guidedPassageStage != wantStage; frames++ {
		if !g.guidedPassageAnimating() {
			break
		}
		g.advanceGuidedPassageAnimation()
	}
	if g.guidedPassageStage != wantStage {
		t.Fatalf("guided passage stage=%d, want %d", g.guidedPassageStage, wantStage)
	}
}

func TestNorudGuidedPassageTransaction(t *testing.T) {
	g, event, guide := norudTestGame(t)

	if !g.talkGuidedPassage(guide) || g.guidedPassageStage != guidedPassageIntroduction ||
		!reflect.DeepEqual(g.dlg.buf, packText(t, g, event.DialogueTextIDs.Introduction)) {
		t.Fatal("正式 guide selector 應先開 D3TXT04 rec87")
	}
	closeGuidedDialogue(g)
	if g.guidedPassageStage != guidedPassageIdle ||
		!g.storyFlag(event.GuidePresentFlagRaw) ||
		g.storyFlag(event.AnimationPendingFlagRaw) ||
		g.storyFlag(event.CompletedFlagRaw) {
		t.Fatal("無國王信時 rec87 關閉後必須 fail closed，不能移動或改旗標")
	}

	g.inventory = append(g.inventory, event.RequiredItemRawID)
	if !g.talkGuidedPassage(guide) {
		t.Fatal("持信時 guide selector 未接手")
	}
	closeGuidedDialogue(g)
	if g.guidedPassageStage != guidedPassageAcceptance ||
		!reflect.DeepEqual(g.dlg.buf, packText(t, g, event.DialogueTextIDs.Acceptance)) {
		t.Fatal("持信時 rec87 後應接原版 rec88")
	}
	closeGuidedDialogue(g)
	runGuidedAnimation(t, g, guidedPassageAwaitTrigger)
	if !g.storyFlag(event.AnimationPendingFlagRaw) || g.dlg.open {
		t.Fatal("引路者抵達後應 set transient flag 並交回玩家控制，不能提前顯示 rec90")
	}
	if guide.facing != 1 {
		t.Fatalf("引路者第一段 mode=2 結束後應朝上，got facing=%d", guide.facing)
	}

	trigger := event.CompletionTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryGuidedPassageTrigger() ||
		g.guidedPassageStage != guidedPassageCompletionDialogue ||
		!reflect.DeepEqual(g.dlg.buf, packText(t, g, event.DialogueTextIDs.GuideReady)) {
		t.Fatal("玩家抵達 special tile 後應由 handler57 顯示 rec90")
	}
	closeGuidedDialogue(g)
	runGuidedAnimation(t, g, guidedPassageIdle)
	if !g.hasItem(event.RequiredItemRawID) {
		t.Fatal("原版 inventory search 不消耗國王的信")
	}
	if g.storyFlag(event.AnimationPendingFlagRaw) ||
		g.storyFlag(event.GuidePresentFlagRaw) ||
		!g.storyFlag(event.CompletedFlagRaw) {
		t.Fatal("完成時應 clear 0x16/0x4c 並 set 0x1c")
	}
	if g.px != trigger.X || g.py != trigger.Y {
		t.Fatalf("handler57 後玩家應留在 trigger=(%d,%d)，got (%d,%d)",
			trigger.X, trigger.Y, g.px, g.py)
	}
	if g.cur.npcAt(event.GuideNPC.Tile.X, event.GuideNPC.Tile.Y) >= 0 {
		t.Fatal("完成後初始 guide NPC 應由 flag0x4c 隱藏")
	}
	if g.cur.npcAt(event.CompletedNPC.Tile.X, event.CompletedNPC.Tile.Y) < 0 {
		t.Fatal("完成後 endpoint NPC 應由 flag0x1c 顯示")
	}

	state, err := encodeSave(g.snapshot())
	if err != nil {
		t.Fatal(err)
	}
	saved, err := decodeSave(state)
	if err != nil {
		t.Fatal(err)
	}
	g2, _, _ := norudTestGame(t)
	g2.restore(saved)
	if !g2.hasItem(event.RequiredItemRawID) ||
		g2.storyFlag(event.GuidePresentFlagRaw) ||
		!g2.storyFlag(event.CompletedFlagRaw) ||
		g2.px != trigger.X || g2.py != trigger.Y {
		t.Fatal("密道完成狀態、信件與落點未通過 save-state round trip")
	}
}

func TestNorudGuidedPassageMatchesOriginalEXECTYAndText(t *testing.T) {
	dir := spineAssetsDir(t)
	read := func(name string) []byte {
		t.Helper()
		raw, err := os.ReadFile(dir + "/" + name)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	events := pack.GuidedPassageEvents()
	if len(events) != 2 {
		t.Fatalf("guided passage variants=%d, want 2", len(events))
	}
	exe := read("DQ3.EXE")
	handler50 := []byte{
		0x9a, 0x64, 0x02, 0x1b, 0x11, 0xe9, 0x49, 0xf2,
		0xbf, 0x0f, 0x0c, 0x9a, 0x64, 0x02, 0x1b, 0x11,
		0xc7, 0x06, 0x93, 0x25, 0x5b, 0x00,
	}
	if !bytes.Equal(exe[0x712f:0x712f+len(handler50)], handler50) {
		t.Fatal("DQ3.EXE handler50 flag/item/text anchor mismatch")
	}
	handler57 := []byte{
		0xbb, 0x16, 0x00, 0xe8, 0x97, 0x10, 0x3c, 0x01,
		0x74, 0x01, 0xc3, 0xe8, 0x88, 0xf1, 0xc6, 0x06,
	}
	if !bytes.Equal(exe[0x71dc:0x71dc+len(handler57)], handler57) {
		t.Fatal("DQ3.EXE scene handler57 transient-flag anchor mismatch")
	}
	dgroupBase := 0x16140
	for _, event := range events {
		if event.GuideEndFacing != "up" || event.GuideReturnEndFacing != "up" {
			t.Fatalf("%s mode=2 final facings=%q/%q, want up/up",
				event.ID, event.GuideEndFacing, event.GuideReturnEndFacing)
		}
		guideRaw := exe[dgroupBase+0x3d03 : dgroupBase+0x3d03+22]
		if got := bytesToHex(guideRaw); got != event.GuideMovementRawHex {
			t.Fatalf("%s guide script=%s, JSON=%s", event.ID, got, event.GuideMovementRawHex)
		}
		wantReturn := [][]byte{
			exe[dgroupBase+0x3d19 : dgroupBase+0x3d19+7],
			exe[dgroupBase+0x3d20 : dgroupBase+0x3d20+7],
		}
		for i := range wantReturn {
			if got := bytesToHex(wantReturn[i]); got != event.GuideReturnRawHex[i] {
				t.Fatalf("%s guide return script%d=%s, JSON=%s",
					event.ID, i, got, event.GuideReturnRawHex[i])
			}
		}
	}

	cty62, cty63 := read("CTY62.DAT"), read("CTY63.DAT")
	if len(cty62) != len(cty63) {
		t.Fatal("CTY62/63 passage variants should have equal record size")
	}
	left, right := append([]byte(nil), cty62...), append([]byte(nil), cty63...)
	for _, off := range []int{18, 21, 22} {
		left[off], right[off] = 0, 0
	}
	if !bytes.Equal(left, right) {
		t.Fatal("CTY62/63 should differ only in map flags and side-specific spawn")
	}
	town, err := dq3data.OpenTown(cty62, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	town63, err := dq3data.OpenTown(cty63, 0, false)
	if err != nil {
		t.Fatal(err)
	}
	wantTransitions := [][4]int{{62, 0xff, 81, 81}, {62, 0xff, 84, 81}}
	if !reflect.DeepEqual(town.Transitions, wantTransitions) ||
		!reflect.DeepEqual(town63.Transitions, wantTransitions) ||
		town.SpawnX != 3 || town.SpawnY != 8 || town.MapFlags != 1 ||
		town63.SpawnX != 94 || town63.SpawnY != 10 || town63.MapFlags != 0 {
		t.Fatalf("CTY62/63 雙側入口資料不符：left=%+v right=%+v transitions=%v/%v",
			[3]int{town.SpawnX, town.SpawnY, int(town.MapFlags)},
			[3]int{town63.SpawnX, town63.SpawnY, int(town63.MapFlags)},
			town.Transitions, town63.Transitions)
	}
	if !reflect.DeepEqual(town.SpecialHandlers, []int{57}) {
		t.Fatalf("CTY62 special handlers=%v, want [57]", town.SpecialHandlers)
	}
	trigger := events[0].CompletionTrigger
	if len(trigger.Tiles) != 1 || trigger.Tiles[0] != (gamepack.TileCoordinate{X: 50, Y: 15}) {
		t.Fatalf("CTY62 handler57 trigger=%+v, want (50,15)", trigger.Tiles)
	}
	i := trigger.Tiles[0].Y*town.W + trigger.Tiles[0].X
	if int(town.HiMap[i]&0x1f) != trigger.TileSubID {
		t.Fatalf("CTY62 trigger subid=%d, JSON=%d", town.HiMap[i]&0x1f, trigger.TileSubID)
	}

	txt := read("D3TXT04.TXT")
	for rec, id := range map[int]string{
		87: events[0].DialogueTextIDs.Introduction,
		88: events[0].DialogueTextIDs.Acceptance,
		89: events[0].DialogueTextIDs.After,
		90: events[0].DialogueTextIDs.GuideReady,
	} {
		want := textRecordWords(t, txt, rec)
		got, ok := pack.TextGlyphCodes(id)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("D3TXT04 rec%d != pack text %s", rec, id)
		}
	}
}

func bytesToHex(raw []byte) string {
	const digits = "0123456789abcdef"
	out := make([]byte, len(raw)*2)
	for i, b := range raw {
		out[i*2], out[i*2+1] = digits[b>>4], digits[b&15]
	}
	return string(out)
}

func textRecordWords(t *testing.T, raw []byte, rec int) []uint16 {
	t.Helper()
	start := int(binary.LittleEndian.Uint16(raw[rec*2:]))
	end := int(binary.LittleEndian.Uint16(raw[(rec+1)*2:]))
	out := make([]uint16, 0, (end-start)/2)
	for off := start; off < end; off += 2 {
		out = append(out, binary.LittleEndian.Uint16(raw[off:]))
	}
	return out
}
