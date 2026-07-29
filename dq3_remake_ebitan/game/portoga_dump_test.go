package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// TestDumpPortogaShipChain:視覺核驗——取船鏈全流程(古布達給黑胡椒 → 波魯多加國王授船
// → 地表停泊船 → 正式登船與航行)。用真實地圖/真實 NPC 資料(非手造 Scene),證明機制在真實素材
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
	events := g.pack.StagedVehicleExchangeEvents()
	if len(events) != 1 {
		t.Fatalf("staged vehicle exchange events=%d, want 1", len(events))
	}
	event := events[0]
	rescueEvents := g.pack.HostageRescueEvents()
	if len(rescueEvents) != 1 {
		t.Fatalf("hostage rescue events=%d, want 1", len(rescueEvents))
	}
	rescue := rescueEvents[0]
	for _, flag := range rescue.CompletionSetFlagsRaw {
		g.setStoryFlag(flag, true)
	}
	g.setStoryFlag(rescue.RewardAvailableFlagRaw, true)

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
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 15, mapBlkNum[15], pepperSec, 0, nil)
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
	g.talkHostageRescue(&sc.npcs[pepperNi])
	dump("pepper_gubda_dialogue") // 古布達致謝與黑胡椒提議
	g.dlg.open = false
	g.advanceHostageRescueDialogue()
	g.hostageRescueChoiceInput(InputState{Confirm: true, DirEdge: -1})
	dump("pepper_gubda_received")
	if !g.hasItem(event.RequiredItemRawID) {
		t.Fatal("視覺 dump 前置:應已取得黑胡椒(古布達 gate 應已放行)")
	}

	// 國王 selector、CTY、座標、handler 與交易內容全部取自 pack。
	kingSec, kingNi := findScriptedNpc(t, g, event.NPC.CTYRaw, event.NPC.HandlerRaw)
	sc2, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, event.NPC.CTYRaw,
		mapBlkNum[event.NPC.CTYRaw], kingSec, 0, nil)
	if err != nil {
		t.Fatalf("載入 CTY%d sec%d: %v", event.NPC.CTYRaw, kingSec, err)
	}
	if sc2.dlgText != nil {
		g.dlg.tx = sc2.dlgText
	}
	g.town, g.cur, g.curCty = sc2, sc2, event.NPC.CTYRaw
	kn := sc2.npcs[kingNi]
	g.px, g.py, g.facing = kn.x, kn.y+1, 1
	g.dlg.open, g.noticeTimer = false, 0 // 清上一步殘留對話/通知(乾淨的 before 截圖)
	dump("king_portoga_before")
	dump("cty37")
	g.selectCommand(cmdTalk)
	dump("portoga_king_letter_intro")
	g.dlg.open = false
	g.advanceStagedVehicleExchangeDialogue()
	dump("portoga_king_letter_item")
	g.noticeTimer = 0
	dump("portoga_king_letter_grant")
	g.dlg.open = false
	g.advanceStagedVehicleExchangeDialogue()

	// 舊文件引用的 wait 圖改成原版 first-visit 交易完成後、尚無黑胡椒的 rec26。
	for i, item := range g.inventory {
		if item == event.RequiredItemRawID {
			g.inventory = append(g.inventory[:i], g.inventory[i+1:]...)
			break
		}
	}
	g.selectCommand(cmdTalk)
	dump("portoga_king_wait")
	g.dlg.open = false
	g.inventory = append(g.inventory, event.RequiredItemRawID)
	g.selectCommand(cmdTalk)
	dump("portoga_king_ship")
	dump("portoga_getship")
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

	move := func(dir int) {
		t.Helper()
		for g.cd > 0 {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatal(err)
			}
		}
		if g.battle.active {
			t.Fatal("取船視覺 fixture 的港口兩格路徑不應進入戰鬥")
		}
		if err := g.step(InputState{DirHeld: dir, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
	move(0) // 港口陸格。
	move(0) // 走上 pack 宣告的停泊船格。
	if !g.shipAboard || g.px != g.shipX || g.py != g.shipY {
		t.Fatalf("視覺 dump 未正式登船：aboard=%v player=(%d,%d) ship=(%d,%d)",
			g.shipAboard, g.px, g.py, g.shipX, g.shipY)
	}
	dump("ship_first_boarded")

	sailDir := -1
	for dir := 0; dir < 4; dir++ {
		dx, dy := dirDelta(dir)
		nx, ny := g.px+dx, g.py+dy
		if nx >= 0 && ny >= 0 && nx < g.cur.w && ny < g.cur.h &&
			g.cur.attr.Raw(g.cur.tileIdx(nx, ny))&0x20 != 0 {
			sailDir = dir
			break
		}
	}
	if sailDir < 0 {
		t.Fatal("停泊船周圍沒有可航水格")
	}
	beforeX, beforeY := g.px, g.py
	move(sailDir)
	if !g.shipAboard || g.px == beforeX && g.py == beforeY {
		t.Fatalf("視覺 dump 未完成首次航行：aboard=%v from=(%d,%d) to=(%d,%d)",
			g.shipAboard, beforeX, beforeY, g.px, g.py)
	}
	dump("ship_first_sailing")
}

// findScriptedNpc:掃 cty 的 sections 0..29,找符合條件的 sub2 scripted NPC,回 (section, npc index)。
// wantB4 依 pack 或原始資料提供的 byte4 匹配。
// 找不到直接 Fatal(dump test 用,前置條件不成立就該失敗,不靜默略過)。
func findScriptedNpc(t *testing.T, g *Game, cty, wantB4 int) (int, int) {
	t.Helper()
	blkn := 1
	if cty >= 0 && cty < len(mapBlkNum) {
		blkn = mapBlkNum[cty]
	}
	for sec := 0; sec < 30; sec++ {
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, cty, blkn, sec, 0, nil)
		if err != nil {
			continue
		}
		for i, n := range sc.npcs {
			if (n.ctrl>>3)&7 != 2 {
				continue
			}
			if n.b4 == wantB4 {
				return sec, i
			}
		}
	}
	t.Fatalf("CTY%d 找不到符合條件的 scripted NPC(b4=%d)", cty, wantB4)
	return 0, 0
}
