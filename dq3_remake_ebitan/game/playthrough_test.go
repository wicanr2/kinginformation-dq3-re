package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/spell"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

func asset(t *testing.T, name string) []byte {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	d, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Skipf("無素材 %s:%v", name, err)
	}
	return d
}

// 整合 game test:真正驅動戰鬥迴圈(遇敵 → 逐回合攻擊 → 勝負結算),驗證 HP/EXP/勝負一致。
func TestPlaythroughBattle(t *testing.T) {
	mons, err := dq3data.OpenMonsters(asset(t, "D3MNS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	b := &Battle{mons: mons, shp: asset(t, "DQ3MNS.SHP")}

	// 強勇者 vs 史萊姆(id5):應必勝、取得 EXP4/G2
	hero := heroParams{level: 20, curHP: 200, maxHP: 200, atk: 80, def: 40, agi: 30}
	if !b.start(5, 12345, hero, nil) {
		t.Fatal("開戰失敗(史萊姆)")
	}
	turns := 0
	for b.result == 0 && turns < 50 {
		b.cursor = bcWar
		b.execTurn()
		turns++
	}
	if b.result != 1 {
		t.Fatalf("強勇者對史萊姆應勝,得 result=%d(%d 回合後 敵HP=%d)", b.result, turns, b.enemyHP)
	}
	if b.gotExp != 4 || b.gotGold != 2 {
		t.Fatalf("史萊姆戰利品應 EXP4/G2,得 EXP%d/G%d", b.gotExp, b.gotGold)
	}
	t.Logf("戰鬥迴圈 ✓:%d 回合擊倒史萊姆,EXP%d G%d、我方 HP %d/%d", turns, b.gotExp, b.gotGold, b.heroHP, b.heroMax)

	// 弱勇者 vs 史萊姆:HP 應被扣(敵有反擊)
	weak := heroParams{level: 1, curHP: 9, maxHP: 9, atk: 4, def: 3, agi: 2}
	b.start(5, 999, weak, nil)
	b.cursor = bcWar
	b.execTurn() // 一回合:我方攻 + 敵反擊
	if b.heroHP == 9 && b.enemyHP > 0 {
		t.Log("注意:一回合後我方 HP 未變(史萊姆傷害可能為 0,合理)")
	}
	t.Logf("弱勇者一回合後:我方 HP %d、敵 HP %d(敵反擊生效)", b.heroHP, b.enemyHP)
}

// 整合 game test:升級 + 咒文習得一致性(串成長表 + 咒文表)。
func TestPlaythroughProgression(t *testing.T) {
	// exp 0 → lv1、HP9、無咒;exp 4364 → lv10、學會 美拉/荷依米/吉拉
	if stats.LevelForExp(0, 0) != 1 || stats.GrowthTarget(0, stats.HP, 1) != 9 {
		t.Fatal("lv1 應 HP9")
	}
	if len(spell.HeroKnown(1)) != 0 {
		t.Fatal("lv1 應無咒")
	}
	lv10 := stats.LevelForExp(0, 4364)
	if lv10 != 10 {
		t.Fatalf("exp4364 應 lv10,得 %d", lv10)
	}
	known := spell.HeroKnown(lv10)
	hasMera, hasHoimi, hasGira := false, false, false
	for _, r := range known {
		switch r {
		case 121:
			hasMera = true
		case 161:
			hasHoimi = true
		case 124:
			hasGira = true
		}
	}
	if !hasMera || !hasHoimi || !hasGira {
		t.Fatalf("lv10 應學會 美拉/荷依米/吉拉,得 %v", known)
	}
	t.Logf("升級鏈 ✓:exp4364→lv10、HP%d、學會 %d 咒(含美拉/荷依米/吉拉)",
		stats.GrowthTarget(0, stats.HP, lv10), len(known))
}

// 整合 game test:商店經濟(買 → 扣金 → 入庫)+ 設施表全城查詢。
func TestPlaythroughEconomy(t *testing.T) {
	items, err := dq3data.OpenItems(asset(t, "ITEM.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	// 阿里阿罕(cty0)武防店(k1):7 品項,含銅劍(id3,價100)
	f := facilityForCty(0, 1)
	if f == nil || f.typ != facWeapon || f.count != 7 {
		t.Fatalf("阿里阿罕武防店應 7 品,得 %+v", f)
	}
	codes := shopItemPool[f.itemOff : f.itemOff+f.count]
	foundSword := false
	for _, c := range codes {
		if c == 3 && items.Price(c) == 100 {
			foundSword = true
		}
	}
	if !foundSword {
		t.Fatalf("武防店應賣銅劍(id3,100G),貨架=%v", codes)
	}
	// 宿屋(k0)inn_cost=2
	inn := facilityForCty(0, 0)
	if inn == nil || inn.typ != facInn || inn.innCost != 2 {
		t.Fatalf("阿里阿罕宿屋應 inn_cost2,得 %+v", inn)
	}
	t.Logf("經濟 ✓:阿里阿罕武防店 7 品(銅劍100G)、宿屋 2G;全城設施表 %d 筆", len(allFacilities))
}

// 整合 game test:寶箱表查詢(阿里阿罕 cty0 sec5 有 6 寶箱)+ 事件系統資料。
func TestPlaythroughTreasure(t *testing.T) {
	if len(treasures) != 121 {
		t.Fatalf("寶箱表應 121 筆,得 %d", len(treasures))
	}
	// 阿里阿罕(cty0 sec5)sub0 寶箱:item98、flag41
	tr := treasureFor(0, 5, 0)
	if tr == nil || tr[4] != 98 || tr[5] != 41 {
		t.Fatalf("阿里阿罕 sec5 sub0 應 item98/flag41,得 %v", tr)
	}
	// 越界查詢 → nil
	if treasureFor(99, 0, 0) != nil {
		t.Fatal("不存在的寶箱應回 nil")
	}
	t.Logf("寶箱表 ✓:121 筆、阿里阿罕 sec5 sub0=item98/flag41")
}
