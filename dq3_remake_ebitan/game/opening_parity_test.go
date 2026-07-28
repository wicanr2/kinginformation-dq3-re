package game

import (
	"os"
	"path/filepath"
	"testing"
)

// TestOriginalNewGameInitialState 固定 DQ3.EXE 的開局資料，不讓「可遊玩預設值」
// 再次取代原版資料。來源：
//   - file 0x13a0..0x1431：地表(0x99,0xae)、CTY00 sec4、室內(5,5)
//   - file 0x1c01..0x1c4e：一人隊伍、等級1、唯一初裝 0x801e(布衣)
func TestOriginalNewGameInitialState(t *testing.T) {
	dir := spineAssetsDir(t)
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "opening-save.json"))

	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	if g.px != aliahanWorldX || g.py != aliahanWorldY {
		t.Errorf("開局記憶地表座標應=(%#x,%#x),得=(%#x,%#x)",
			aliahanWorldX, aliahanWorldY, g.px, g.py)
	}
	if g.overPx != aliahanWorldX || g.overPy != aliahanWorldY {
		t.Errorf("返回地表座標應=(%#x,%#x),得=(%#x,%#x)",
			aliahanWorldX, aliahanWorldY, g.overPx, g.overPy)
	}
	if g.equip != [4]int{0, 0x1e} {
		t.Errorf("開局應只裝備布衣 0x1e,得 %#v", g.equip)
	}
	if len(g.companions) != 0 {
		t.Errorf("原版開局 party size=1，不應預塞同伴：%d", len(g.companions))
	}
	if g.progressDone(msStart) || g.progressStage() != 0 {
		t.Errorf("出生時尚未見國王，msStart 不應完成：stage=%d", g.progressStage())
	}

	g.startOpening()
	if !g.inTown || g.curCty != 0 || g.cur == nil || g.cur.sec != 4 {
		sec := -1
		if g.cur != nil {
			sec = g.cur.sec
		}
		t.Errorf("開場應在 CTY00 sec4：inTown=%v cty=%d sec=%d",
			g.inTown, g.curCty, sec)
	}
	if g.px != openingHomeX || g.py != openingHomeY {
		t.Errorf("家中出生點應=(%d,%d),得=(%d,%d)",
			openingHomeX, openingHomeY, g.px, g.py)
	}
}

// TestOriginalOpeningEventTransactions 固定開場 runner handler54/56 的狀態交易：
// 母親帶到家門外後切 flag50→17；首次見王得到精確六件物品與 50G，再切 17→18。
func TestOriginalOpeningEventTransactions(t *testing.T) {
	dir := spineAssetsDir(t)
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "opening-events-save.json"))

	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	wantSeq := []int{82, 83, 81, 80}
	if len(openingSeq) != len(wantSeq) {
		t.Fatalf("開場對白序列長度=%d，應為 %d", len(openingSeq), len(wantSeq))
	}
	for i := range wantSeq {
		if openingSeq[i] != wantSeq[i] {
			t.Fatalf("開場對白[%d]=%d，應為 %d", i, openingSeq[i], wantSeq[i])
		}
	}

	g.motherEscort()
	if g.curCty != 0 || g.cur == nil || g.cur.sec != 0 || g.px != 8 || g.py != 38 {
		t.Fatalf("母親帶路落點應 CTY00 sec0@(8,38)，得 cty=%d sec=%v @(%d,%d)",
			g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	if !g.storyFlag(0x17) || g.storyFlag(0x50) {
		t.Fatalf("母親事件後應 set flag17/clear flag50：17=%v 50=%v",
			g.storyFlag(0x17), g.storyFlag(0x50))
	}

	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatalf("載入阿里阿罕王座: %v", err)
	}
	g.cur, g.town, g.curCty, g.inTown = throne, throne, ctyAliahanCastle, true
	g.px, g.py, g.facing = aliahanKingX, aliahanKingY+1, 1
	if !g.tryOpeningRegionEvent() { // production runner 路徑：踏入國王正前方 region
		t.Fatal("國王正前方應命中 opening runner handler56")
	}
	if g.heroGold != 50 {
		t.Errorf("首次謁見應得 50G，得 %d", g.heroGold)
	}
	if len(g.inventory) != len(aliahanKingRewardItems) {
		t.Fatalf("首次謁見應得 %d 件，得 %v", len(aliahanKingRewardItems), g.inventory)
	}
	for i, want := range aliahanKingRewardItems {
		if g.inventory[i] != want {
			t.Errorf("國王獎勵[%d]=%#x，應為 %#x", i, g.inventory[i], want)
		}
	}
	if g.storyFlag(0x17) || !g.storyFlag(0x18) || !g.progressDone(msStart) {
		t.Errorf("謁見後應 clear17/set18/msStart：17=%v 18=%v ms=%v",
			g.storyFlag(0x17), g.storyFlag(0x18), g.progressDone(msStart))
	}

	// 再呼叫不可重複領取。
	g.talkAliahanKing()
	if g.heroGold != 50 || len(g.inventory) != len(aliahanKingRewardItems) {
		t.Errorf("國王獎勵必須一次性：gold=%d inventory=%v", g.heroGold, g.inventory)
	}
}

func sceneSection(s *Scene) int {
	if s == nil {
		return -1
	}
	return s.sec
}
