package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpSpellEffects:視覺核驗(規則63/64:留一份可肉眼核對的畫面)—— 一場戰鬥中玩家依序施放
// 拜基魯多(151,我方物攻↑)、拉里荷(144,敵陷入睡眠),dump 施放前/後畫面 + log 詠唱訊息(b.msg;
// 引擎目前無戰鬥文字 overlay,訊息走 stderr/test log,同既有 dump test 慣例只驗「畫面 state」)。
// 平常 skip;設 DQ3_DUMP_SPELLFX=1(+ 可選 SPELLFX_OUT,預設 _spellfx_dump/,已 gitignore)才跑。
func TestDumpSpellEffects(t *testing.T) {
	if os.Getenv("DQ3_DUMP_SPELLFX") == "" {
		t.Skip("設 DQ3_DUMP_SPELLFX=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("SPELLFX_OUT")
	if out == "" {
		out = "_spellfx_dump"
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	t.Setenv("DQ3_BATTLE", "5") // 史萊姆,單隻(HP低,不易誤殺打斷觀察窗)
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	if !g.battle.active || len(g.battle.enemies) == 0 {
		t.Fatal("DQ3_BATTLE=5 應開戰")
	}
	g.battle.heroMP, g.battle.heroMaxMP = 99, 99 // 確保 MP 足以連續施放(dump 用,略過等級習得限制)
	g.battle.enemies[0].hp, g.battle.enemies[0].max = 999, 999

	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped(msg=%q)", name, g.battle.msg)
	}
	dump("00_open")

	g.battle.execSpell(151) // 拜基魯多:我方物攻↑
	if g.battle.heroAtkPct != 200 {
		t.Fatalf("拜基魯多後 heroAtkPct 應為200,得 %d", g.battle.heroAtkPct)
	}
	dump("01_after_baikiruto_heroAtkPct200")

	g.battle.execSpell(144) // 拉里荷:敵陷入睡眠
	if g.battle.enemies[0].status&statusParalysis == 0 {
		t.Fatal("拉里荷後敵應陷入睡眠")
	}
	dump("02_after_rariho_enemy_asleep")

	g.battle.cursor = bcWar // 物攻(buff 生效中,傷害應明顯高於未 buff)
	before := g.battle.enemies[0].hp
	g.battle.execTurn()
	dmg := before - g.battle.enemies[0].hp
	t.Logf("buff 後物攻傷害 = %d(heroAtkPct=%d)", dmg, g.battle.heroAtkPct)
	dump("03_after_buffed_attack")

	t.Logf("咒文效果 dump ✓:存於 %s(00開場/01拜基魯多後/02拉里荷後/03buff攻擊後)", out)
}
