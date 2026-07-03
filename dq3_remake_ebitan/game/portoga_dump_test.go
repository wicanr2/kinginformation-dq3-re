package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// TestDumpPortogaShipChain:視覺核驗——取船鏈全流程(古布達給黑胡椒 → 波魯多加國王授船
// → 主角出海站上停泊的船)。用真實地圖/真實 NPC 資料(非手造 Scene),證明機制在真實素材
// + 真實 renderFrame 下確實 live(規則 63/64)。平常 skip;DQ3_DUMP_SHIP=1 + SHIP_OUT=
// <gitignored 夾> 才跑。
func TestDumpPortogaShipChain(t *testing.T) {
	if os.Getenv("DQ3_DUMP_SHIP") == "" {
		t.Skip("設 DQ3_DUMP_SHIP=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("SHIP_OUT")
	if out == "" {
		t.Fatal("需設 SHIP_OUT")
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
	g.showTitle = false
	if g.flags == nil {
		g.flags = map[int]bool{}
	}
	g.flags[0x211] = true // 前置:已救出達妮亞(C-4a 批次,甘達特巢穴 boss 鏈勝利)

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

	// 找古布達(CTY15 byte4=25 scripted NPC)所在 section。
	pepperSec, pepperNi := findScriptedNpc(t, g, 15, 25)
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 15, mapBlkNum[15], pepperSec)
	if err != nil {
		t.Fatalf("載入 CTY15 sec%d: %v", pepperSec, err)
	}
	if sc.dlgText != nil {
		g.dlg.tx = sc.dlgText
	}
	g.town, g.cur, g.inTown, g.curCty = sc, sc, true, 15
	n := sc.npcs[pepperNi]
	g.px, g.py, g.facing = n.x, n.y+1, 1 // 站南邊面向北 → frontTile=(n.x,n.y)
	g.dlg.open, g.noticeTimer = false, 0 // 清殘留對話/通知(乾淨的 before 截圖)
	dump("pepper_gubda_before")          // 已救人,對話前
	g.scriptedTalk(n.b4)
	dump("pepper_gubda_dialogue") // 給黑胡椒對白
	if !g.hasItem(itemPepper) {
		t.Fatal("視覺 dump 前置:應已取得黑胡椒(古布達 gate 應已放行)")
	}

	// 找波魯多加國王(CTY37 位置特例 (9,6))所在 section。
	kingSec, kingNi := findScriptedNpc(t, g, ctyPortoga, -1) // -1=改用座標比對(國王非 scriptedTable 項)
	sc2, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, ctyPortoga, mapBlkNum[ctyPortoga], kingSec)
	if err != nil {
		t.Fatalf("載入 CTY%d sec%d: %v", ctyPortoga, kingSec, err)
	}
	if sc2.dlgText != nil {
		g.dlg.tx = sc2.dlgText
	}
	g.town, g.cur, g.curCty = sc2, sc2, ctyPortoga
	kn := sc2.npcs[kingNi]
	g.px, g.py, g.facing = kn.x, kn.y+1, 1
	g.dlg.open, g.noticeTimer = false, 0 // 清上一步殘留對話/通知(乾淨的 before 截圖)
	dump("king_portoga_before")          // 授船前(持黑胡椒站國王前)
	g.selectCommand(cmdTalk)
	dump("king_portoga_ship") // 授船對白(shipOwned 應已 true)
	if !g.shipOwned {
		t.Fatal("視覺 dump:對話後應已取船")
	}

	// 出城至地表(exitTown 換回真正的地表 Scene,不能只清 inTown flag),鏡頭置中在
	// 港口船格附近,證明船 sprite 真的畫在停泊座標上。
	g.dlg.open, g.noticeTimer = false, 0 // 清對話框,乾淨看船 sprite
	g.exitTown()
	g.px, g.py = g.shipX, g.shipY-2
	t.Logf("船停泊座標=(%d,%d)", g.shipX, g.shipY)
	dump("ship_on_overworld") // 地表可見停泊的船
}

// findScriptedNpc:掃 cty 的 sections 0..29,找符合條件的 sub2 scripted NPC,回 (section, npc index)。
// wantB4>=0 → 依 byte4 匹配(scriptedTable 項);wantB4<0 → 改用波魯多加國王座標特例比對。
// 找不到直接 Fatal(dump test 用,前置條件不成立就該失敗,不靜默略過)。
func findScriptedNpc(t *testing.T, g *Game, cty, wantB4 int) (int, int) {
	t.Helper()
	blkn := 1
	if cty >= 0 && cty < len(mapBlkNum) {
		blkn = mapBlkNum[cty]
	}
	for sec := 0; sec < 30; sec++ {
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, cty, blkn, sec)
		if err != nil {
			continue
		}
		for i, n := range sc.npcs {
			if (n.ctrl>>3)&7 != 2 {
				continue
			}
			if wantB4 >= 0 && n.b4 == wantB4 {
				return sec, i
			}
			if wantB4 < 0 && n.x == portogaKingX && n.y == portogaKingY {
				return sec, i
			}
		}
	}
	t.Fatalf("CTY%d 找不到符合條件的 scripted NPC(b4=%d)", cty, wantB4)
	return 0, 0
}
