package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpMerchantRevolution 繪製真正的 CTY83 革命場景、牢中建城者對話與黃寶珠結果。
// 這是選擇性視覺 fixture；從商人交付自然完成中間兩條主線 gate 仍由 production trace 驗收。
func TestDumpMerchantRevolution(t *testing.T) {
	if os.Getenv("DQ3_DUMP_MERCHANT_REVOLUTION") == "" {
		t.Skip("設 DQ3_DUMP_MERCHANT_REVOLUTION=1 才產出商人城革命核對圖")
	}
	out := os.Getenv("MERCHANT_REVOLUTION_OUT")
	if out == "" {
		t.Fatal("需設 MERCHANT_REVOLUTION_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	orb, ok := g.pack.TreasureEvent("dq3:event.merchant_town_yellow_orb")
	if !ok {
		t.Fatal("缺商人城黃寶珠事件")
	}
	g.settlementFounder = newMember([]int{0}, 6, 0, 0)
	g.setStoryFlag(0x15, true)
	g.setStoryFlag(orb.Treasure.PresentFlag, true)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		orb.Treasure.CTYRaw, mapBlkNum[orb.Treasure.CTYRaw], orb.Treasure.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.inTown, g.curCty = false, true, orb.Treasure.CTYRaw
	g.town, g.cur, g.dlg.tx = sc, sc, sc.dlgText

	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	g.px, g.py = 3, 23
	dump("merchant_revolution_cty83")
	npcIndex := g.cur.npcAt(2, 23)
	if npcIndex < 0 || !g.talkSettlementFounderFollowup(&g.cur.npcs[npcIndex]) {
		t.Fatal("CTY83 找不到原版 handler48 建城者")
	}
	dump("merchant_revolution_imprisoned")
	g.dlg.open = false
	g.advanceSettlementFounderFollowup()
	dump("merchant_revolution_seat_hint")

	g.dlg.open = false
	g.px, g.py = 4, 2
	if !g.collectPackTreasure(orb.Treasure) {
		t.Fatal("黃寶珠 transaction 失敗")
	}
	g.cur.revealEventTile(g.px, g.py)
	dump("merchant_revolution_yellow_orb_obtained")
}
