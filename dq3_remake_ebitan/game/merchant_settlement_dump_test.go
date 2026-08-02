package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpMerchantSettlement 繪製真正的 CTY58 場景與 game pack 對話狀態。
// 完整的玩家可達交易由 TestOpeningProductionInputTrace 覆蓋；此選擇性測試負責
// 產出檢閱用 PNG。
func TestDumpMerchantSettlement(t *testing.T) {
	if os.Getenv("DQ3_DUMP_MERCHANT_SETTLEMENT") == "" {
		t.Skip("設 DQ3_DUMP_MERCHANT_SETTLEMENT=1 才產出商人城核對圖")
	}
	out := os.Getenv("MERCHANT_SETTLEMENT_OUT")
	if out == "" {
		t.Fatal("需設 MERCHANT_SETTLEMENT_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, err := NewGame(os.DirFS(spineAssetsDir(t)), nil)
	if err != nil {
		t.Fatal(err)
	}
	event, ok := g.pack.SettlementFounderEvent("dq3:event.merchant_settlement_founder")
	if !ok {
		t.Fatal("缺少商人城交付事件")
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, event.NPC.CTYRaw,
		mapBlkNum[event.NPC.CTYRaw], event.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle, g.inTown, g.curCty = false, true, event.NPC.CTYRaw
	g.town, g.cur = sc, sc
	g.dlg.tx = sc.dlgText
	g.companions = []*Member{newLevelOneMember([]int{0}, event.RequiredClassRaw, 0,
		&g.prng, g.tavern.equipment)}
	g.px, g.py, g.facing = event.NPC.Tile.X, event.NPC.Tile.Y+1, 1

	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{
				R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255,
			})
		}
		f, err := os.Create(filepath.Join(out, name+".png"))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}

	dump("merchant_settlement_cty58_before")
	n := g.cur.npcs[g.cur.npcAt(event.NPC.Tile.X, event.NPC.Tile.Y)]
	if !g.talkSettlementFounder(&n) {
		t.Fatal("無法開啟商人城原版對話")
	}
	dump("merchant_settlement_introduction")
	g.dlg.open = false
	g.advanceSettlementFounderDialogue()
	dump("merchant_settlement_first_offer")
}
