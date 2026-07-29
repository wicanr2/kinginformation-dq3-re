package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpGarunaSatoriBook 產出加爾那之塔領悟之書取得前後的現行 runtime 畫面。
// 正式玩家可達性由 TestOpeningProductionInputTrace 負責；本測試只固定同一原版場景與
// event tile，供 deterministic PNG 與目視核對。平常 skip。
func TestDumpGarunaSatoriBook(t *testing.T) {
	if os.Getenv("DQ3_DUMP_GARUNA") == "" {
		t.Skip("設 DQ3_DUMP_GARUNA=1 才執行")
	}
	out := os.Getenv("GARUNA_OUT")
	if out == "" {
		t.Fatal("需設 GARUNA_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	assets := os.Getenv("DQ3_ASSETS")
	if assets == "" {
		assets = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(assets), nil)
	if err != nil {
		t.Fatal(err)
	}
	g.showTitle = false
	events := g.pack.TreasureEvents()
	if len(events) != 1 {
		t.Fatalf("treasure events=%d，want 1", len(events))
	}
	tr := events[0].Treasure
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		tr.CTYRaw, mapBlkNum[tr.CTYRaw], tr.Section, g.dnPhase, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, tr.CTYRaw, true
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
	}

	found := false
	eventX, eventY := -1, -1
	for y := 0; y < sc.h && !found; y++ {
		for x := 0; x < sc.w && !found; x++ {
			ev, subid, ok := sc.tileEvent(x, y)
			if !ok || subid != tr.TileSubID || ev[0] != tr.EventTypeRaw ||
				ev[1] != tr.ItemRawID || ev[2] != tr.PresentFlag {
				continue
			}
			for face := 0; face < 4; face++ {
				dx, dy := dirDelta(face)
				sx, sy := x-dx, y-dy
				if sx < 0 || sy < 0 || sx >= sc.w || sy >= sc.h || sc.Blocked(sx, sy) {
					continue
				}
				g.px, g.py, g.facing = sx, sy, face
				eventX, eventY = x, y
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("找不到領悟之書 event tile 的操作格")
	}

	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW,
				color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
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

	dump("garuna_satori_before")
	beforeTile := g.cur.tileIdx(eventX, eventY)
	g.selectCommand(cmdExamine)
	if !g.hasItem(tr.ItemRawID) || g.storyFlag(tr.PresentFlag) {
		t.Fatalf("取得領悟之書失敗：item=%v flag=%v",
			g.hasItem(tr.ItemRawID), g.storyFlag(tr.PresentFlag))
	}
	if got := g.cur.tileIdx(eventX, eventY); got != (beforeTile+1)&0xff {
		t.Fatalf("取得後寶箱 tile=%#x，want %#x", got, (beforeTile+1)&0xff)
	}
	if _, _, ok := g.cur.tileEvent(eventX, eventY); ok {
		t.Fatal("取得後寶箱 event subid 應立即清除")
	}
	dump("garuna_satori_obtained")
}
