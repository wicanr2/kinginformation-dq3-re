package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func defeatTestGame(t *testing.T) *Game {
	t.Helper()
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.openingIdx = false, -1
	g.cur, g.inTown, g.curCty = g.overworldScene(), false, -1
	g.px, g.py, g.overPx, g.overPy = aliahanWorldX, aliahanWorldY, aliahanWorldX, aliahanWorldY
	g.heroExp = stats.ExpForLevel(0, 12)
	g.ensureHeroStats()
	_, maxHP, _, _, _ := g.heroStats()
	g.heroHP, g.heroMP = maxHP, 23
	return g
}

func closeDefeatDialogue(t *testing.T, g *Game) {
	t.Helper()
	for steps := 0; g.dlg.open && steps < 8; steps++ {
		g.dlg.Advance()
		if !g.dlg.open {
			g.advanceDefeatDialogue()
		}
	}
	if g.dlg.open || g.defeatDialogueStage != defeatDialogueIdle {
		t.Fatalf("敗北對話未在有界輸入內結束：open=%v stage=%d", g.dlg.open, g.defeatDialogueStage)
	}
}

func TestFieldPoisonPartialDeathPromotesLivingLeaderWithoutRespawn(t *testing.T) {
	g := defeatTestGame(t)
	g.heroHP, g.heroConditions = 1, conditionPoison
	g.companions = []*Member{{Class: 2, Gender: 1, CurHP: 12}}
	beforeX, beforeY := g.px, g.py

	g.applyHazardStep()

	if g.heroHP != 0 || g.companions[0].CurHP != 12 || g.partyLeader != 1 {
		t.Fatalf("部分死亡交易錯：hero=%d companion=%d leader=%d", g.heroHP, g.companions[0].CurHP, g.partyLeader)
	}
	if g.dlg.open || g.px != beforeX || g.py != beforeY {
		t.Fatalf("尚有存活者不得開全滅對話或傳送：dlg=%v pos=(%d,%d)", g.dlg.open, g.px, g.py)
	}
	if got, want := g.heroSceneSprite(), g.partySprite(2, 1); got != want {
		t.Fatal("死亡隊長後，地表領隊 sprite 未切換為第一名存活同伴")
	}
}

func TestFieldPoisonAllDownUsesTwoMessagesAndLastSavePoint(t *testing.T) {
	g := defeatTestGame(t)
	saveName := t.TempDir() + "/dq3save.json"
	t.Setenv("DQ3_SAVE", saveName)
	g.px, g.py, g.overPx, g.overPy = 153, 174, 153, 174
	if err := g.Save(); err != nil {
		t.Fatal(err)
	}
	wantRespawn := g.respawn
	g.px, g.py, g.overPx, g.overPy = 90, 80, 90, 80
	g.heroHP, g.heroMP, g.heroGold, g.heroConditions = 1, 23, 777, conditionPoison
	g.companions = []*Member{{Class: 1, CurHP: 1, CurMP: 9, Conditions: conditionPoison}}

	g.applyHazardStep()
	partyTextID := g.defeatTextID("party_defeated")
	partyGlyphs, _ := g.pack.TextGlyphCodes(partyTextID)
	if !g.dlg.open || g.defeatDialogueStage != defeatDialogueParty ||
		!reflect.DeepEqual(g.dlg.buf, partyGlyphs) {
		t.Fatalf("field 全滅未先開 record361：open=%v stage=%d buf=%v", g.dlg.open, g.defeatDialogueStage, g.dlg.buf)
	}

	g.dlg.Advance()
	if !g.dlg.open {
		g.advanceDefeatDialogue()
	}
	voiceID := g.defeatTextID("revival_voice")
	voiceGlyphs, _ := g.pack.TextGlyphCodes(voiceID)
	if !g.dlg.open || g.defeatDialogueStage != defeatDialogueVoice ||
		!reflect.DeepEqual(g.dlg.buf, voiceGlyphs) {
		t.Fatalf("record361 後未接 record362：open=%v stage=%d", g.dlg.open, g.defeatDialogueStage)
	}
	_, maxHP, _, _, _ := g.heroStats()
	if g.heroHP != maxHP || g.heroMP != 23 || g.heroGold != 777 || g.companions[0].CurHP != 0 {
		t.Fatalf("原版只復活第一人且不得回捲進度：hero=%d/%d mp=%d gold=%d companion=%d",
			g.heroHP, maxHP, g.heroMP, g.heroGold, g.companions[0].CurHP)
	}
	if g.px != wantRespawn.PX || g.py != wantRespawn.PY || g.overPx != wantRespawn.OverPX || g.overPy != wantRespawn.OverPY {
		t.Fatalf("未回最後 Save 位置：got=(%d,%d;%d,%d) want=%+v", g.px, g.py, g.overPx, g.overPy, wantRespawn)
	}
	closeDefeatDialogue(t, g)
}

func TestBattleAllDownContinuesAtRevivalVoiceWithoutHealingCompanions(t *testing.T) {
	g := defeatTestGame(t)
	g.respawn = g.currentRespawnPoint()
	g.heroHP = 1
	g.companions = []*Member{{Class: 1, CurHP: 1}}
	g.battle.heroHP, g.battle.heroMP, g.battle.result = 0, 7, 2
	g.battle.companions = []*battleActor{{hp: 0, mp: 3}}

	g.onBattleEnd()
	voiceGlyphs, _ := g.pack.TextGlyphCodes(g.defeatTextID("revival_voice"))
	if !g.dlg.open || g.defeatDialogueStage != defeatDialogueVoice ||
		!reflect.DeepEqual(g.dlg.buf, voiceGlyphs) {
		t.Fatalf("戰鬥 record361 後應直接接 record362：open=%v stage=%d", g.dlg.open, g.defeatDialogueStage)
	}
	_, maxHP, _, _, _ := g.heroStats()
	if g.heroHP != maxHP || g.heroMP != 7 || g.companions[0].CurHP != 0 {
		t.Fatalf("戰鬥全滅復歸錯：hero=%d/%d mp=%d companion=%d", g.heroHP, maxHP, g.heroMP, g.companions[0].CurHP)
	}
	closeDefeatDialogue(t, g)
}
