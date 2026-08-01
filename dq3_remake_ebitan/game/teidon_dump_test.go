package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// TestDumpTeidonDarkLamp fixes the current Ebitengine CTY20 treasure result
// and the pack-driven overworld night result as runtime PNGs. Natural
// reachability remains covered by TestOpeningProductionInputTrace.
func TestDumpTeidonDarkLamp(t *testing.T) {
	if os.Getenv("DQ3_DUMP_TEIDON") == "" {
		t.Skip("設 DQ3_DUMP_TEIDON=1 才執行")
	}
	out := os.Getenv("TEIDON_OUT")
	if out == "" {
		t.Fatal("需設 TEIDON_OUT")
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
	event, ok := g.pack.TreasureEvent("dq3:event.teidon_dark_lamp")
	if !ok {
		t.Fatal("缺 dq3:event.teidon_dark_lamp")
	}
	tr := event.Treasure
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		tr.CTYRaw, mapBlkNum[tr.CTYRaw], tr.Section, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, tr.CTYRaw, true
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
	}

	found := false
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
				if sx < 0 || sy < 0 || sx >= sc.w || sy >= sc.h ||
					sc.Blocked(sx, sy) {
					continue
				}
				g.px, g.py, g.facing = sx, sy, face
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("找不到黑暗燈 event tile 的操作格")
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

	g.selectCommand(cmdExamine)
	if !g.hasItem(tr.ItemRawID) || g.storyFlag(tr.PresentFlag) {
		t.Fatalf("取得黑暗燈失敗：item=%v flag=%v",
			g.hasItem(tr.ItemRawID), g.storyFlag(tr.PresentFlag))
	}
	dump("teidon_dark_lamp_obtained")

	g.cur, g.town, g.inTown, g.curCty = g.overworldScene(), nil, false, -1
	g.px, g.py = ctyLoc[tr.CTYRaw][0]-1, ctyLoc[tr.CTYRaw][1]
	g.dnPhase, g.dnStep = 0, 47
	g.panel, g.panelCursor = panelItem, 0
	g.noticeTimer = 0
	g.useSelectedItem()
	if g.dnPhase != 2 || g.dnStep != 0 || !g.hasItem(tr.ItemRawID) {
		t.Fatalf("黑暗燈夜景 fixture transaction 錯：phase=%d step=%d item=%v",
			g.dnPhase, g.dnStep, g.hasItem(tr.ItemRawID))
	}
	g.panel = panelNone // 對齊 production trace 使用後按 B 關閉道具清單
	dump("teidon_dark_lamp_night")
}

// TestDumpTeidonFinalKeyGreenOrb records the current night CTY20 runtime
// before/after formal final-key use and while handler35's reward dialogue is
// visible. The full player route remains covered by TestOpeningProductionInputTrace.
func TestDumpTeidonFinalKeyGreenOrb(t *testing.T) {
	if os.Getenv("DQ3_DUMP_TEIDON") == "" {
		t.Skip("設 DQ3_DUMP_TEIDON=1 才執行")
	}
	out := os.Getenv("TEIDON_OUT")
	if out == "" {
		t.Fatal("需設 TEIDON_OUT")
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
	var reward *gamepack.NPCItemRewardEvent
	for _, event := range g.pack.NPCItemRewardEvents() {
		if event.ID == "dq3:event.teidon_green_orb" {
			e := event
			reward = &e
			break
		}
	}
	if reward == nil {
		t.Fatal("缺 dq3:event.teidon_green_orb")
	}
	g.setStoryFlag(reward.PresentFlagRaw, true)
	g.dnPhase, g.dnStep = 2, 0
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		reward.NPC.CTYRaw, mapBlkNum[reward.NPC.CTYRaw], reward.NPC.Section, 2, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, reward.NPC.CTYRaw, true
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
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

	g.px, g.py, g.facing = 17, 5, 1
	dump("teidon_final_key_door_closed")
	g.inventory = []int{0x57}
	g.panel, g.panelCursor = panelItem, 0
	g.useSelectedItem()
	if tier := g.cur.doorTier(17, 4); tier != 0 || !g.hasItem(0x57) {
		t.Fatalf("最終鑰匙畫面 fixture 交易錯：tier=%d item=%v", tier, g.hasItem(0x57))
	}
	dump("teidon_final_key_door_open")

	g.px, g.py, g.facing = reward.NPC.Tile.X, reward.NPC.Tile.Y+1, 1
	g.selectCommand(cmdTalk)
	if !g.dlg.open || g.storyFlag(reward.PresentFlagRaw) {
		t.Fatalf("綠色寶珠畫面 fixture 交易錯：dialogue=%v flag=%v",
			g.dlg.open, g.storyFlag(reward.PresentFlagRaw))
	}
	// 取得道具的短暫 notice 會蓋住對話框第一頁；dump 要固定通知
	// 倒數結束後仍開啟的 production 對話狀態，不修改交易或文字。
	g.noticeTimer = 0
	dump("teidon_green_orb_dialogue")
}
