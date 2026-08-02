package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func trackedWorldObjectTestGame(t *testing.T) (*Game, string) {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	objects := g.pack.TrackedWorldObjects()
	if len(objects) != 1 {
		t.Fatalf("tracked_world_objects 數量=%d，want 1", len(objects))
	}
	obj := objects[0]
	event, ok := g.pack.ChoiceItemExchangeEvent("dq3:event.greenland_transform_staff_exchange")
	if !ok || event.ActivateWorldObject.ObjectID != obj.ID {
		t.Fatal("交換事件未連接幽靈船世界物件")
	}
	if !g.activateTrackedWorldObject(event.ActivateWorldObject) {
		t.Fatal("無法以 pack activation 建立幽靈船座標")
	}
	g.worldState |= uint16(obj.ActiveWorldStateMask)
	return g, obj.ID
}

func TestTrackedWorldObjectLocatorUsesOriginalDirectionAndDistanceRecords(t *testing.T) {
	g, objectID := trackedWorldObjectTestGame(t)
	obj, _ := g.pack.TrackedWorldObject(objectID)
	p := g.trackedWorldPositions[objectID]
	g.inTown, g.layer, g.px, g.py = false, p.Layer, p.X+12, p.Y-7

	if !g.useTrackedWorldObjectLocator(obj.TrackerItemRawID) || !g.dlg.open {
		t.Fatal("船員骨頭未開原版前導記錄")
	}
	want, _ := g.pack.TextGlyphCodes(obj.TrackerDialogueTextIDs.Preamble)
	if !reflect.DeepEqual(g.dlg.buf, want) {
		t.Fatal("船員骨頭前導文字未由 pack 載入")
	}
	g.dlg.open = false
	g.advanceTrackedWorldObjectLocator()
	want, _ = g.pack.TextGlyphCodes(obj.TrackerDialogueTextIDs.West)
	if !reflect.DeepEqual(g.dlg.buf, want) ||
		!reflect.DeepEqual(g.dlg.varGlyph[dq3data.TxtVarItem], digitGlyphs(12)) {
		t.Fatalf("X 距離記錄錯：buf=%v var=%v", g.dlg.buf, g.dlg.varGlyph)
	}
	g.dlg.open = false
	g.advanceTrackedWorldObjectLocator()
	want, _ = g.pack.TextGlyphCodes(obj.TrackerDialogueTextIDs.South)
	if !reflect.DeepEqual(g.dlg.buf, want) ||
		!reflect.DeepEqual(g.dlg.varGlyph[dq3data.TxtVarItem], digitGlyphs(7)) {
		t.Fatalf("Y 距離記錄錯：buf=%v var=%v", g.dlg.buf, g.dlg.varGlyph)
	}
}

func TestTrackedWorldObjectEntryAndLoveMemoryTreasure(t *testing.T) {
	g, objectID := trackedWorldObjectTestGame(t)
	obj, _ := g.pack.TrackedWorldObject(objectID)
	p := g.trackedWorldPositions[objectID]
	g.inTown, g.layer, g.curCty = false, p.Layer, -1
	g.cur, g.px, g.py = g.overworldScene(), p.X, p.Y
	g.shipOwned, g.shipAboard = true, true
	if !g.tryEnterTrackedWorldObject() || !g.inTown || g.curCty != obj.EntranceCTYRaw ||
		g.cur == nil || g.cur.sec != obj.EntranceSection {
		t.Fatalf("幽靈船動態入口錯：town=%v cty=%d sec=%v", g.inTown, g.curCty, g.cur)
	}

	var treasureX, treasureY = -1, -1
	section, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, obj.EntranceCTYRaw,
		mapBlkNum[obj.EntranceCTYRaw], 1,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatalf("載入幽靈船 section1：%v", err)
	}
	for y := 0; y < section.h; y++ {
		for x := 0; x < section.w; x++ {
			_, subid, ok := section.tileEvent(x, y)
			if ok && subid == 0 {
				treasureX, treasureY = x, y
			}
		}
	}
	if treasureX != 18 || treasureY != 55 {
		t.Fatalf("愛的回憶事件座標=(%d,%d)，want (18,55)", treasureX, treasureY)
	}
	g.town, g.cur, g.curCty, g.inTown = section, section, obj.EntranceCTYRaw, true
	g.px, g.py = treasureX, treasureY
	g.examine()
	if !g.hasItem(0x64) || g.storyFlag(0x8b) {
		t.Fatalf("愛的回憶寶箱交易錯：item=%v present=%v", g.hasItem(0x64), g.storyFlag(0x8b))
	}
}

func TestTrackedWorldObjectRelocationMatchesOriginalWindowGate(t *testing.T) {
	g, objectID := trackedWorldObjectTestGame(t)
	obj, _ := g.pack.TrackedWorldObject(objectID)
	candidate := obj.Relocation.Candidates[9]
	g.worldObjectBufferX, g.worldObjectBufferY, g.worldObjectBufferValid = 60, 140, true
	for seed := 0; seed <= 0xffff; seed++ {
		g.prng.Seed(uint16(seed))
		if g.prng.Next(len(obj.Relocation.Candidates)) == 9 {
			g.prng.Seed(uint16(seed))
			break
		}
	}
	g.rebuildTrackedWorldObjects()
	want := trackedWorldPosition{X: candidate.X, Y: candidate.Y, Layer: candidate.Layer}
	if got := g.trackedWorldPositions[objectID]; got != want {
		t.Fatalf("視窗內候選未取代視窗外座標：got=%+v want=%+v", got, want)
	}
	state := g.prng.State()
	g.rebuildTrackedWorldObjects()
	if got := g.trackedWorldPositions[objectID]; got != want || g.prng.State() != state {
		t.Fatalf("物件已在 80x80 視窗內仍被重抽：got=%+v state=%#x/%#x", got, g.prng.State(), state)
	}
}
