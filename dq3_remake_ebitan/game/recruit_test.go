package game

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/rng"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// createViaTavern:跑完整酒館 2F 登錄所流程(職業→命名→性別)產一名成員,回傳 tv.input 的結果。
// 命名跳過注音輸入,直接選英數格盤 OK 格(允許空名 → 回退職業名,tavern.go tavGender 分支邏輯)。
func createViaTavern(tv *Tavern) *Member {
	r := rng.New(0x1357)
	tv.open()
	tv.input(InputState{DirEdge: -1, Confirm: true}, r) // 選職業(游標預設0)→ 進命名
	tv.ni.nameZhu = false
	tv.ni.cursor = niCellOK
	tv.input(InputState{DirEdge: -1, Confirm: true}, r) // 完成命名(空名)→ 進性別
	m, _ := tv.input(InputState{DirEdge: -1, Confirm: true}, r)
	return m
}

// TestTavernCreateGoesToRosterNotCompanions:酒館 2F 登錄所創角(g.tavernCreate glue)只登錄
// roster,不觸碰既有 companions——這是本次修的核心 bug(舊碼建完角色立即 append 進 companions,
// 滿 3 人時 `append(companions[1:], m)` 頂替最舊成員)。建 4 名角色後 roster 應有 4 筆,
// companions 維持原本 3 名不變(逐一比對 pointer,確認第 0 位「最舊」成員沒被頂替/丟失)。
func TestTavernCreateGoesToRosterNotCompanions(t *testing.T) {
	g := &Game{companions: startingCompanions(0)}
	orig := append([]*Member(nil), g.companions...)

	var tv Tavern
	for i := 0; i < 4; i++ {
		m := createViaTavern(&tv)
		if m == nil {
			t.Fatalf("第 %d 名創角應成功回傳 Member", i+1)
		}
		if len(g.roster) < rcRosterMax { // 對齊 g.tavernCreate 的 glue 邏輯(見 TestGameTavernCreateGoesToRoster)
			g.roster = append(g.roster, m)
		}
	}

	if len(g.roster) != 4 {
		t.Fatalf("roster 應有 4 筆(創 4 名角色),得 %d", len(g.roster))
	}
	if len(g.companions) != 3 {
		t.Fatalf("companions 不應被創角動作變動,應維持 3 名,得 %d", len(g.companions))
	}
	for i, m := range orig {
		if g.companions[i] != m {
			t.Errorf("companions[%d] 被創角動作改動(疑似頂替/丟隊友 bug 重現):原 %p 現 %p", i, m, g.companions[i])
		}
	}
}

// TestGameTavernCreateGoesToRoster:直接經 g.tavern(而非獨立 Tavern 值)+ g.tavernCreate 走一遍,
// 對齊 game.go Update() 實際呼叫路徑(g.tavern.active 分支)。
func TestGameTavernCreateGoesToRoster(t *testing.T) {
	g := &Game{companions: startingCompanions(0)}
	origLen := len(g.companions)

	g.tavern.open()
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true}) // 選職業 → 命名
	g.tavern.ni.nameZhu = false
	g.tavern.ni.cursor = niCellOK
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true}) // 完成命名(空名)→ 性別
	g.tavernCreate(InputState{DirEdge: -1, Confirm: true}) // 選性別 → 建角,glue 應寫進 roster

	if len(g.roster) != 1 {
		t.Fatalf("g.tavernCreate 建角後 roster 應有 1 筆,得 %d", len(g.roster))
	}
	if len(g.companions) != origLen {
		t.Fatalf("g.tavernCreate 不應動 companions,應維持 %d,得 %d", origLen, len(g.companions))
	}
}

// TestRecruitJoinMovesRosterMemberIntoParty:酒場「找同伴參加」(rcMenu cursor0 → rcJoin → 選一名
// 確定)應把該角色從 roster 移進 companions(不重複、不遺失)。
func TestRecruitJoinMovesRosterMemberIntoParty(t *testing.T) {
	a := newMember(classNames[1], 1, 0, 0)
	b := newMember(classNames[2], 2, 0, 0)
	g := &Game{roster: []*Member{a, b}}

	g.recruit.open()
	g.recruitInput(InputState{DirEdge: -1, Confirm: true}) // 主選單 cursor0「找同伴參加」→ rcJoin
	if g.recruit.stage != rcJoin {
		t.Fatalf("選主選單第 0 項應進 rcJoin,得 stage=%d", g.recruit.stage)
	}
	g.recruit.cursor = 0
	g.recruitInput(InputState{DirEdge: -1, Confirm: true}) // 選 roster[0]=a → 入隊

	if len(g.companions) != 1 || g.companions[0] != a {
		t.Fatalf("companions 應含 a(入隊),得 %+v", g.companions)
	}
	if len(g.roster) != 1 || g.roster[0] != b {
		t.Fatalf("roster 應只剩 b,得 %+v", g.roster)
	}
}

// TestRecruitJoinBlockedWhenPartyFull:隊伍已滿(3 同伴,對齊 DQ3_PARTY_MAX=主角+3)時「找同伴參加」
// 應擋掉,不頂替 / 不丟失既有隊友(對比修正前 bug:append(companions[1:], m))。
func TestRecruitJoinBlockedWhenPartyFull(t *testing.T) {
	full := startingCompanions(0) // 3 名,已達上限
	extra := newMember(classNames[6], 6, 1, 0)
	g := &Game{companions: full, roster: []*Member{extra}}
	origFirst := g.companions[0]

	g.recruit.stage, g.recruit.cursor = rcJoin, 0
	g.recruitInput(InputState{DirEdge: -1, Confirm: true})

	if len(g.companions) != 3 {
		t.Fatalf("隊伍已滿時 companions 長度應維持 3,得 %d", len(g.companions))
	}
	if g.companions[0] != origFirst {
		t.Errorf("隊伍已滿時不應頂替第 0 位既有隊友(舊 bug 重現):原 %p 現 %p", origFirst, g.companions[0])
	}
	if len(g.roster) != 1 || g.roster[0] != extra {
		t.Fatalf("擋掉時 roster 應維持不變(extra 仍在名冊),得 %+v", g.roster)
	}
}

// TestRecruitLeaveMovesCompanionBackToRoster:「與同伴分離」把選中的隊員移回 roster,不刪除。
func TestRecruitLeaveMovesCompanionBackToRoster(t *testing.T) {
	comps := startingCompanions(0)
	g := &Game{companions: comps}
	target := comps[1]

	g.recruit.stage, g.recruit.cursor = rcLeave, 1
	g.recruitInput(InputState{DirEdge: -1, Confirm: true})

	if len(g.companions) != 2 {
		t.Fatalf("分離後 companions 應剩 2 名,得 %d", len(g.companions))
	}
	for _, m := range g.companions {
		if m == target {
			t.Fatalf("分離目標不應仍在 companions 內")
		}
	}
	if len(g.roster) != 1 || g.roster[0] != target {
		t.Fatalf("分離的隊員應回到 roster(不刪除),得 %+v", g.roster)
	}
}

// TestRecruitViewReturnsToMenu:「觀看名單」任意鍵/確定應回主選單,且不改動 roster/companions(唯讀)。
func TestRecruitViewReturnsToMenu(t *testing.T) {
	g := &Game{
		roster:     []*Member{newMember(classNames[3], 3, 0, 0)},
		companions: []*Member{newMember(classNames[4], 4, 1, 0)},
	}
	g.recruit.stage, g.recruit.cursor = rcView, 0
	g.recruitInput(InputState{DirEdge: -1, Confirm: true})

	if g.recruit.stage != rcMenu {
		t.Fatalf("觀看名單確定後應回主選單,得 stage=%d", g.recruit.stage)
	}
	if len(g.roster) != 1 || len(g.companions) != 1 {
		t.Fatalf("觀看名單是唯讀,不應改動 roster/companions,得 roster=%d companions=%d",
			len(g.roster), len(g.companions))
	}
}

// TestRecruitCancelFromSubmenuReturnsToMenuWithoutMoving:子選單(rcJoin/rcLeave)按 Cancel 應回主選單,
// 且不搬動任何成員。
func TestRecruitCancelFromSubmenuReturnsToMenuWithoutMoving(t *testing.T) {
	g := &Game{roster: []*Member{newMember(classNames[5], 5, 0, 0)}}
	g.recruit.stage, g.recruit.cursor = rcJoin, 0
	g.recruitInput(InputState{DirEdge: -1, Cancel: true})

	if g.recruit.stage != rcMenu {
		t.Fatalf("子選單 Cancel 應回主選單,得 stage=%d", g.recruit.stage)
	}
	if len(g.roster) != 1 || len(g.companions) != 0 {
		t.Fatalf("Cancel 不應搬動成員,得 roster=%d companions=%d", len(g.roster), len(g.companions))
	}
}

func TestRegisteredMemberUsesOriginalLevelOneTransaction(t *testing.T) {
	r := rng.New(0x1357)
	m := newLevelOneMember([]int{1, 2, 3}, 3, 1, r, [4]int{-1, 0x1e, -1, -1})
	if m.Armor != 0x1e || m.Weapon != -1 || m.Shield != -1 || m.Head != -1 {
		t.Fatalf("原版 sub_1c94 新登錄角色應只穿布衣 0x1e，得 W/A/S/H=%#x/%#x/%#x/%#x",
			m.Weapon, m.Armor, m.Shield, m.Head)
	}
	if m.Stats == (stats.Values{}) || m.CurHP != int(m.Stats[stats.HP]) ||
		m.CurMP != int(m.Stats[stats.MP]) {
		t.Fatalf("Lv1 七欄/目前 HP MP 未由同一 transaction 建立：%v HP%d MP%d",
			m.Stats, m.CurHP, m.CurMP)
	}
}

func TestAliahanSpecialNPCDispatchMatchesEXEHandlers(t *testing.T) {
	dir := spineAssetsDir(t)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatalf("NewGame: %v", err)
	}
	check := func(sec, x, y, wantB4 int, verify func()) {
		t.Helper()
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], sec, 0, nil)
		if err != nil {
			t.Fatalf("load CTY00 sec%d: %v", sec, err)
		}
		idx := sc.npcAt(x, y)
		if idx < 0 {
			t.Fatalf("CTY00 sec%d @(%d,%d) 找不到原版特殊 NPC", sec, x, y)
		}
		n := sc.npcs[idx]
		if (n.ctrl>>3)&7 != 2 || n.b4 != wantB4 {
			t.Fatalf("特殊 NPC identity 錯：sub%d b4=%d", (n.ctrl>>3)&7, n.b4)
		}
		g.curCty, g.cur, g.town = 0, sc, sc
		g.openAliahanSpecialNPC(&sc.npcs[idx])
		verify()
		g.tavern.active, g.recruit.active, g.dlg.open = false, false, false
	}
	check(0, 2, 16, 1, func() {
		if !g.recruit.active {
			t.Fatal("b4=1 file0x6482 應開露依達酒場")
		}
	})
	check(2, 2, 3, 3, func() {
		if !g.tavern.active {
			t.Fatal("b4=3 file0x649f 應開冒險者登錄所")
		}
	})
	check(0, 8, 21, 4, func() {
		if !g.dlg.open || !reflect.DeepEqual(g.dlg.buf, g.shop.nameText.Record(561)) {
			t.Fatal("b4=4 file0x64a3 應顯示預存所 rec561，不可誤開酒場")
		}
	})
}
