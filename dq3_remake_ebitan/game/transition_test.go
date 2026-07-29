package game

import (
	"os"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 多 section 轉場整合:阿里阿罕(CTY0)section0 掃出轉場格 → tileTransition 回目的 →
// 對非出城目的(destSec<0xfe)實際 loadTownSceneSec 載得起來。驗「內容廣度」的跨 section 路徑。
func TestSectionTransition(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(dir + "/CTY00.DAT"); err != nil {
		t.Skipf("無素材:%v", err)
	}
	assets := os.DirFS(dir)
	rd := func(n string) []byte { d, _ := os.ReadFile(dir + "/" + n); return d }
	pal := dq3data.DecodePalette(rd("DQ3.PAL"), 256)
	manBLS := rd("DQ3MAN.BLS")

	sc0, err := loadTownScene(assets, pal, manBLS, 0, 1, 0, nil)
	if err != nil {
		t.Fatalf("載阿里阿罕 sec0:%v", err)
	}
	if len(sc0.transitions) == 0 {
		t.Fatal("阿里阿罕 sec0 無轉場表(疑 section+0xc 未接進 Scene)")
	}

	// 掃全圖找轉場格,驗 tileTransition 命中 + 目的合理;非出城目的實際載一次。
	hit, loadedDest := 0, 0
	destSeen := map[int]bool{}
	for y := 0; y < sc0.h; y++ {
		for x := 0; x < sc0.w; x++ {
			dcty, dsec, dx, dy, ok := sc0.tileTransition(x, y)
			if !ok {
				continue
			}
			hit++
			if dsec == 0xff || dsec == 0xfe { // 出城:座標為地表 spawn,不在此載
				continue
			}
			if dsec >= 16 || dx > 255 || dy > 255 {
				t.Fatalf("轉場格 (%d,%d) 目的異常:cty=%d sec=%d (%d,%d)", x, y, dcty, dsec, dx, dy)
			}
			ncty := 0
			if dcty != 0 && dcty < 100 {
				ncty = dcty
			}
			key := ncty*100 + dsec
			if destSeen[key] {
				continue
			}
			destSeen[key] = true
			blkn := 1
			if ncty < len(mapBlkNum) {
				blkn = mapBlkNum[ncty]
			}
			ns, err := loadTownSceneSec(assets, pal, manBLS, ncty, blkn, dsec, 0, nil)
			if err != nil {
				t.Fatalf("轉場目的 CTY%d sec%d 載入失敗:%v", ncty, dsec, err)
			}
			if ns.sec != dsec || ns.w < 1 || ns.h < 1 {
				t.Fatalf("目的 section 壞:sec=%d %d×%d", ns.sec, ns.w, ns.h)
			}
			loadedDest++
		}
	}
	if hit == 0 {
		t.Fatal("阿里阿罕 sec0 全圖找不到任何轉場格(attr&0xe000)")
	}
	t.Logf("阿里阿罕 sec0:轉場格 %d 個、載入 %d 個不同目的 section ✓", hit, loadedDest)
}

// 門/鑰匙:掃阿里阿罕各 section 找鎖門格(attr&0xc0),驗 doorTier 1..3 + openDoor 後變通行 tile。
func TestDoorOpen(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(dir + "/CTY00.DAT"); err != nil {
		t.Skipf("無素材:%v", err)
	}
	assets := os.DirFS(dir)
	rd := func(n string) []byte { d, _ := os.ReadFile(dir + "/" + n); return d }
	pal := dq3data.DecodePalette(rd("DQ3.PAL"), 256)
	manBLS := rd("DQ3MAN.BLS")

	doors, opened := 0, 0
	for sec := 0; sec < 12; sec++ {
		sc, err := loadTownSceneSec(assets, pal, manBLS, 0, 1, sec, 0, nil)
		if err != nil {
			continue
		}
		for y := 0; y < sc.h; y++ {
			for x := 0; x < sc.w; x++ {
				tier := sc.doorTier(x, y)
				if tier == 0 {
					continue
				}
				doors++
				if tier < 1 || tier > 3 {
					t.Fatalf("sec%d 門 (%d,%d) 等級異常:%d", sec, x, y, tier)
				}
				before := sc.tileIdx(x, y)
				if !sc.openDoor(x, y) {
					t.Fatalf("sec%d 門 (%d,%d) 開啟失敗", sec, x, y)
				}
				if sc.doorTier(x, y) != 0 { // 開後不再是門
					t.Fatalf("sec%d 門 (%d,%d) 開後仍是門", sec, x, y)
				}
				if sc.tileIdx(x, y) == before && sc.attr.Blocked(before) {
					t.Fatalf("sec%d 門 (%d,%d) 開後 tile 未變且仍阻擋", sec, x, y)
				}
				opened++
			}
		}
	}
	if doors == 0 {
		t.Skip("阿里阿罕各 section 無鎖門(attr&0xc0)—— 門系統邏輯已驗於其他城")
	}
	t.Logf("阿里阿罕:鎖門 %d 個、開啟驗證 %d 個 ✓", doors, opened)
}

// 原版 file 0x49d1 跨 CTY consumer 會以目的 CTY 查 cty_loc，更新出城所回的地表座標。
// CTY30 sec2→CTY31 sec1 是前段正常流程的真實跨 CTY transition。
func TestCrossCTYTransitionUpdatesRememberedWorldPosition(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	if _, err := os.Stat(dir + "/CTY30.DAT"); err != nil {
		t.Skipf("無素材:%v", err)
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, 30, mapBlkNum[30], 2, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.inTown, g.curCty = sc, sc, true, 30
	g.px, g.py = 4, 36
	g.overPx, g.overPy = ctyLoc[30][0], ctyLoc[30][1]
	g.tryTransition()
	if !g.inTown || g.curCty != 31 || sceneSection(g.cur) != 1 || g.px != 2 || g.py != 4 {
		t.Fatalf("CTY30→31 transition 錯：town=%v cty=%d sec=%d @(%d,%d)",
			g.inTown, g.curCty, sceneSection(g.cur), g.px, g.py)
	}
	if g.overPx != ctyLoc[31][0] || g.overPy != ctyLoc[31][1] {
		t.Fatalf("跨 CTY 未依目的 cty_loc 更新 remembered world：got(%d,%d) want(%d,%d)",
			g.overPx, g.overPy, ctyLoc[31][0], ctyLoc[31][1])
	}
}
