package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// TestDumpNpcDensity:C-3 視覺核驗——移除 b2<4 continue 後,城鎮 NPC 密度是否明顯回升。
// section 選用同時滿足「OpenTown 不報錯」+「tile 陣列真的落在檔案大小內(w*h*2 不超出
// len(cty),排除 layout 誤解出巨大 w/h 但實際讀 0 補滿的偽陽性)」+「b2<4 熱點高」的兩筆:
// CTY03 sec15(18×18、33 NPC、25 個 b2<4)、CTY90 sec13(11×11、55 NPC、26 個 b2<4)。
// (docs/68 原標「CTY01 sec5/CTY03 sec4」用的是另一支未存檔的離線統計腳本,索引口徑不同、
// 且未做「tile 陣列落在檔案內」的嚴格檢查——CTY01 sec9 這類看似有效但 w=257,h=3079 遠超
// 檔案大小的「section」流出去會變成幾乎全讀 0,dump 出來只有主角站在空地上,已排除。)
// dump 成 PNG 供肉眼比對 DOSBox 原版截圖(NPC 密度/位置,不比對 b2<4 fallback sprite 本身
// ——該值原版 undefined,見 docs/68 修正)。平常 skip;DQ3_DUMP_NPC=1 + NPC_OUT=<gitignored 夾> 才跑。
func TestDumpNpcDensity(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NPC") == "" {
		t.Skip("設 DQ3_DUMP_NPC=1 才執行(視覺驗證 dump 工具)")
	}
	out := os.Getenv("NPC_OUT")
	if out == "" {
		t.Fatal("需設 NPC_OUT")
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
	g.showTitle = false // 跳過標題畫面分支,renderFrame() 才會走到城鎮繪製路徑
	dump := func(name string, cty, section int) {
		blkn := 1
		if cty >= 0 && cty < len(mapBlkNum) {
			blkn = mapBlkNum[cty]
		}
		sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, cty, blkn, section)
		if err != nil {
			t.Fatalf("%s: load CTY%02d sec%d: %v", name, cty, section, err)
		}
		g.town, g.cur, g.inTown, g.curCty = sc, sc, true, cty
		// 鏡頭置中在 NPC 群聚點(非 spawn),讓 dump 盡量框到最多 NPC 供核對數量。
		if len(sc.npcs) > 0 {
			sx, sy := 0, 0
			for _, n := range sc.npcs {
				sx += n.x
				sy += n.y
			}
			g.px, g.py = sx/len(sc.npcs), sy/len(sc.npcs)
		} else {
			g.px, g.py = sc.spawnX, sc.spawnY
		}
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
		t.Logf("%s: CTY%02d sec%d NPC 總數=%d(dump 視窗內可能少於此數)", name, cty, section, len(sc.npcs))
	}
	dump("cty03_sec15", 3, 15)
	dump("cty90_sec13", 90, 13)
}
