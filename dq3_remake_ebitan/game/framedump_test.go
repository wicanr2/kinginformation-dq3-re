package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// TestDumpNewGameScreens:把新遊戲流程各畫面(主選單/命名/性別/能力確認)dump 成 PNG,
// 供與 docs/36_shots 原版實機截圖肉眼對照(glyph 形近誤標高風險 → 視覺驗證,規則 63)。
// 平常 skip;設 DQ3_DUMP_NG=1 + NG_OUT=<gitignored 夾> 才跑。
func TestDumpNewGameScreens(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NG") == "" {
		t.Skip("設 DQ3_DUMP_NG=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("NG_OUT")
	if out == "" {
		t.Fatal("需設 NG_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(out + "/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped", name)
	}
	// 依新遊戲狀態機逐畫面 dump(直接設 state,不跑 ebiten 主迴圈)
	g.newGame.stage = ngMenu
	dump("ng_menu")
	g.newGame.stage = ngName
	dump("ng_name")
	g.newGame.stage = ngGender
	dump("ng_gender")
	g.newGame.ni.nameBuf = []int{15} // 英數「A」
	g.heroName = append([]int(nil), g.newGame.ni.nameBuf...)
	g.rollHeroLevelOne()
	g.newGame.preview = g.heroStat
	g.newGame.previewDef = int(g.heroStat[stats.VIT]) + g.shop.items.Defense(g.equip[1])
	g.newGame.stage = ngConfirm
	dump("ng_confirm")

	// C-2:模擬創角完成 → 進主角家(sec4)+ 原版開場 runner 序列。
	g.showTitle = false
	g.startOpening()
	dump("opening_home_rec82") // 家室內 + 旁白第一句(rec82)
	g.dlg.Open(83)
	dump("opening_home_rec83")
	g.dlg.Open(81)
	dump("opening_home_rec81")

	// handler54:母親帶出門 → sec0 (8,38) → rec80。
	g.motherEscort()
	g.dlg.Open(openingMotherDirectionsRec)
	dump("opening_town_rec80")

	// handler56:走到阿里阿罕王座正前方 → rec78 + 50G/六件裝備。
	throne, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctyAliahanCastle, mapBlkNum[ctyAliahanCastle], aliahanThroneSection,
		g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty = throne, throne, ctyAliahanCastle
	g.px, g.py, g.facing = aliahanKingX, aliahanKingY+1, 1
	if !g.tryOpeningRegionEvent() {
		t.Fatal("王座 opening region 未觸發")
	}
	dump("opening_king_rec78")
	t.Logf("開場:謁見後 cty=%d sec=%d @(%d,%d), gold=%d items=%v",
		g.curCty, g.cur.sec, g.px, g.py, g.heroGold, g.inventory)

	// CTY00 原版 sub2 handler 身分接線後的兩個可達酒場畫面。
	luida, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 0, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = luida, luida, 0, true
	g.px, g.py, g.facing = 2, 18, 1 // 隔 row17 櫃台面向 b4=1 露依達
	g.dlg.open = false
	g.recruit.open()
	dump("luida_recruit")
	g.recruit.active = false

	registry, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 0, mapBlkNum[0], 2, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town = registry, registry
	g.px, g.py, g.facing = 2, 5, 1 // 隔 row4 櫃台面向 b4=3 登錄員
	g.tavern.open()
	dump("adventurer_registry")
	g.tavern.active = false

	// R-2:沙曼歐莎夜間在原版精確座標使用拉之鏡 → rec97 揭露 → 怪力魔89。
	g.dnPhase = 2
	g.setStoryFlag(0x42, true)
	g.setStoryFlag(0x10, false)
	samanosa, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		ctySamanosa, mapBlkNum[ctySamanosa], mirrorSection, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = samanosa, samanosa, ctySamanosa, true
	g.px, g.py, g.inventory = mirrorX, mirrorY, []int{0x61}
	g.dlg.tx = samanosa.dlgText
	g.useMirror()
	dump("samanosa_mirror_reveal")

	g.dlg.open, g.mirrorStage = false, 2
	g.advanceMirrorEvent()
	if !g.battle.active || g.battle.monID != monsterBossTroll {
		t.Fatal("拉之鏡揭露後未進怪力魔89戰")
	}
	dump("samanosa_boss_troll")
}
