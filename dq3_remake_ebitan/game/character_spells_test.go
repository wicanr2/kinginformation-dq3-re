package game

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func characterSpellsViewGame(t *testing.T) *Game {
	g := recruitmentViewGame(t)
	// 直接UI元件fixture未走正式出生與進城；兩個欄位缺漏會啟動
	// Load的舊存檔相容路徑。正常209包另從冷啟動驗證。
	g.rollHeroLevelOne()
	g.addVisitedTown(g.openingEscort.CTY)
	// 元件fixture，正常209包另驗；使用pack第三職業與既有出生規則。
	g.roster[0] = newLevelOneMember([]int{0}, g.tavern.contract.ClassOptions[2].ClassRaw, 0, &g.prng, g.tavern.equipment)
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	if g.recruit.stage != rcViewAbility || g.recruit.viewFlow == nil {
		t.Fatal("spell-bearing member did not open ability")
	}
	return g
}

func TestCharacterSpellsSeparateViewWaitsAndState(t *testing.T) {
	for _, in := range []InputState{{DirEdge: -1, AnyKeyEdge: true}, {DirEdge: -1, Cancel: true}, {DirEdge: -1, Rename: true}, {DirEdge: -1, Tapped: true}} {
		g := characterSpellsViewGame(t)
		before := g.snapshot()
		random := g.prng
		for n := 0; n < 5; n++ {
			g.recruitInput(InputState{DirHeld: 0, DirEdge: -1})
		}
		if g.recruit.stage != rcViewAbility {
			t.Fatal("held key consumed wait")
		}
		g.recruitInput(in)
		if g.recruit.stage != rcViewSpells || len(g.recruit.viewSpellNames) == 0 || g.recruit.renameFlow != nil {
			t.Fatal("first key skipped spell wait")
		}
		g.recruitInput(in)
		if g.recruit.stage != rcViewClose || g.recruit.renameFlow != nil || !reflect.DeepEqual(before, g.snapshot()) || g.prng != random {
			t.Fatal("spell close skipped independent caller wait or changed state")
		}
		g.recruitInput(InputState{DirEdge: -1, Cancel: true})
		if g.recruit.viewFlow != nil || len(g.recruit.viewSpellNames) != 0 {
			t.Fatal("close retained view UI")
		}
	}
}

func TestCharacterSpellsDynamicRowsDoNotMutatePack(t *testing.T) {
	g := characterSpellsViewGame(t)
	r := g.recruit.raster
	s := r.spells
	original := s.RawWindow
	for _, count := range []int{1, 4, 5, len(s.Catalog)} {
		ids := make([]string, count)
		for i := range ids {
			ids[i] = s.Catalog[i].TextID
		}
		g.renderFrame()
		r.characterSpells(g.recruit.tx, ids)
		rows := (count + s.Columns - 1) / s.Columns
		if r.windows[s.RawWindow.ID].Height != (rows+s.ExtraRows)*s.Names.StepY || s.RawWindow != original {
			t.Fatal("dynamic rows modified pack or height")
		}
	}
}

func TestCharacterSpellsLoadValidatesBeforeClearing(t *testing.T) {
	g := characterSpellsViewGame(t)
	g.recruitInput(InputState{DirEdge: -1, Enter: true})
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "spells.json"))
	if e := g.Save(); e != nil {
		t.Fatal(e)
	}
	valid, e := os.ReadFile(savePath())
	if e != nil {
		t.Fatal(e)
	}
	before := g.snapshot()
	flow := g.recruit.viewFlow
	ids := append([]string(nil), g.recruit.viewSpellNames...)
	random := g.prng
	bad := before
	bad.PackSchema = "invalid"
	raw, e := encodeSave(bad)
	if e != nil {
		t.Fatal(e)
	}
	if e = os.WriteFile(savePath(), raw, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e == nil || g.recruit.stage != rcViewSpells || g.recruit.viewFlow != flow || !reflect.DeepEqual(ids, g.recruit.viewSpellNames) || !reflect.DeepEqual(before, g.snapshot()) || g.prng != random {
		t.Fatal("bad Load changed spell page or state")
	}
	if e = os.WriteFile(savePath(), valid, 0644); e != nil {
		t.Fatal(e)
	}
	if e = g.Load(); e != nil {
		t.Fatal(e)
	}
	if g.recruit.active || g.recruit.viewFlow != nil || len(g.recruit.viewSpellNames) != 0 || !reflect.DeepEqual(before, g.snapshot()) || g.prng != random {
		a, b := reflect.ValueOf(before), reflect.ValueOf(g.snapshot())
		for i := 0; i < a.NumField(); i++ {
			if !reflect.DeepEqual(a.Field(i).Interface(), b.Field(i).Interface()) {
				t.Logf("changed %s: %v -> %v", a.Type().Field(i).Name, a.Field(i).Interface(), b.Field(i).Interface())
			}
		}
		t.Logf("UI active=%v view=%v spells=%v RNG changed=%v", g.recruit.active, g.recruit.viewFlow != nil, g.recruit.viewSpellNames, g.prng != random)
		t.Fatal("valid Load did not clear only UI")
	}
}

func TestCharacterSpellsRegistrationWaitDoesNotCommit(t *testing.T) {
	g := characterSpellsViewGame(t)
	tv := &g.tavern
	if e := tv.installRaster(g.newGame.raster, g.worldPal); e != nil {
		t.Fatal(e)
	}
	tv.active, tv.stage, tv.candidate = true, tavReview, g.roster[0]
	before := *tv.candidate
	random := g.prng
	g.tavernCreate(InputState{DirEdge: -1, Cancel: true})
	if tv.stage != tavSpells || len(g.roster) != 1 {
		t.Fatal("first key skipped spell wait or committed")
	}
	g.tavernCreate(InputState{DirEdge: -1, Tapped: true})
	if tv.stage != tavAccept || tv.candidate == nil || !reflect.DeepEqual(before, *tv.candidate) || len(g.roster) != 1 || g.prng != random {
		t.Fatal("spell wait cancelled, committed or changed candidate")
	}
}
