package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestDumpChoiceItemExchange(t *testing.T) {
	if os.Getenv("DQ3_DUMP_CHOICE_ITEM_EXCHANGE") == "" {
		t.Skip("設 DQ3_DUMP_CHOICE_ITEM_EXCHANGE=1 才產出 CTY54 核對圖")
	}
	out := os.Getenv("CHOICE_ITEM_EXCHANGE_OUT")
	if out == "" {
		t.Fatal("需設 CHOICE_ITEM_EXCHANGE_OUT")
	}
	g, event, _ := choiceItemExchangeTestGame(t)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, event.NPC.CTYRaw,
		mapBlkNum[event.NPC.CTYRaw], event.NPC.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.town, g.cur, g.dlg.tx = sc, sc, sc.dlgText
	g.showTitle = false
	g.px, g.py, g.facing = event.NPC.Tile.X, event.NPC.Tile.Y+1, 1
	g.inventory = append(g.inventory, event.RequiredItemRawID)
	g.setStoryFlag(event.AvailableFlagRaw, true)
	n := &g.cur.npcs[g.cur.npcAt(event.NPC.Tile.X, event.NPC.Tile.Y)]
	if !g.talkChoiceItemExchange(n) {
		t.Fatal("無法開啟 CTY54 交換事件")
	}
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
	dump("greenland_transform_staff_offer")
	g.dlg.open = false
	g.advanceChoiceItemExchangeDialogue()
	dump("greenland_transform_staff_choice")
}
