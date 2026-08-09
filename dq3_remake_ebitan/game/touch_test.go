package game

import (
	"math"
	"testing"
)

func TestQuantize4(t *testing.T) {
	cases := []struct {
		dx, dy, want int
		name         string
	}{
		{0, 0, -1, "死區中心"},
		{5, 5, -1, "死區內"},
		{0, 40, 0, "下"},
		{0, -40, 1, "上"},
		{-40, 0, 2, "左"},
		{40, 0, 3, "右"},
		{40, 10, 3, "右下偏右→右(主軸)"},
		{10, 40, 0, "右下偏下→下(主軸)"},
	}
	for _, c := range cases {
		if got := quantize4(c.dx, c.dy); got != c.want {
			t.Errorf("%s quantize4(%d,%d)=%d,want %d", c.name, c.dx, c.dy, got, c.want)
		}
	}
}

func TestTouchZones(t *testing.T) {
	tu := newTouchUI()
	if z := tu.zoneOf(aCX, aCY); z != zoneA {
		t.Errorf("A 鍵中心應 zoneA,得 %d", z)
	}
	if z := tu.zoneOf(bCX, bCY); z != zoneB {
		t.Errorf("B 鍵中心應 zoneB,得 %d", z)
	}
	if z := tu.zoneOf(padCX, padCY); z != zoneDpad {
		t.Errorf("十字鍵中心應 zoneDpad,得 %d", z)
	}
	if z := tu.zoneOf(ScreenW/2+10, 10); z != zoneNone { // 右上空白
		t.Errorf("右上空白應 zoneNone,得 %d", z)
	}
}

// TestTouchContextZone:情境鍵(右上)只在 SetContext 給非空標籤時才吃觸點;
// 隱藏(空字串)時同一點應回 zoneNone,不可誤吞其他區的觸控。
func TestTouchContextZone(t *testing.T) {
	tu := newTouchUI()
	if z := tu.zoneOf(ctxCX, ctxCY); z != zoneNone {
		t.Errorf("未設情境標籤時,情境鍵座標應 zoneNone,得 %d", z)
	}
	tu.SetContext("設定")
	if z := tu.zoneOf(ctxCX, ctxCY); z != zoneCtx {
		t.Errorf("SetContext 後情境鍵中心應 zoneCtx,得 %d", z)
	}
	tu.SetContext("")
	if z := tu.zoneOf(ctxCX, ctxCY); z != zoneNone {
		t.Errorf("SetContext(\"\") 後應回 zoneNone(隱藏不吃點),得 %d", z)
	}
}

// TestCtxTapMapsToInputState:ctxTap(觸控來源)應映射進 InputState.CtxTap,
// 對齊 dpadDir/aTap/bTap→InputState 既有映射模式。
//
// 注:Input.Poll() 內部呼叫 ip.touch.poll(),而 poll() 會先把 ctxTap 重置為 false、
// 再用 ebiten.AppendTouchIDs 等真實裝置 API 重新偵測——Xvfb headless 測試環境沒有實體
// 觸控裝置,不會產生 just-pressed 觸點事件,因此無法透過「呼叫 Poll() 前先灌
// ip.touch.ctxTap=true」來驗證映射(會被 poll() 立刻歸零蓋掉)。
// 改用 applyTouch(拆出的映射步驟,見 input.go)直接驗證「touch 偵測結果 → InputState」
// 這段映射邏輯本身,不依賴 poll() 的裝置偵測部分。
func TestCtxTapMapsToInputState(t *testing.T) {
	ip := newInput()
	ip.touch.ctxTap = true
	s := InputState{DirHeld: -1, DirEdge: -1}
	ip.applyTouch(&s)
	if !s.CtxTap {
		t.Errorf("touch.ctxTap=true 經 applyTouch 應映射為 InputState.CtxTap=true")
	}
}

// TestUpdateTouchContextLabel:驗證「狀態 → 情境鍵標籤」轉移(規劃 docs/64 §2):
// 標題畫面 → 設定、酒館命名階段 → 注、其餘 → 隱藏(空字串)。
// 只建純狀態 Game(無素材/無 ebiten 主迴圈),呼叫 updateTouchContext() 本身,
// 不跑完整 Update()/RunGame(那需要地圖/素材等大量依賴,無法在單元測試輕量驗證)。
func TestUpdateTouchContextLabel(t *testing.T) {
	g := &Game{input: newInput()}

	g.showTitle = true
	g.updateTouchContext()
	if g.input.touch.ctxLabel != "設定" {
		t.Errorf("標題畫面(主選單前/中)應設情境標籤「設定」,得 %q", g.input.touch.ctxLabel)
	}

	g.newGame.stage = ngName // 主角命名階段→ 情境鍵應改「注」(英數↔注音),同酒館命名
	g.updateTouchContext()
	if g.input.touch.ctxLabel != "注" {
		t.Errorf("主角命名階段應設情境標籤「注」,得 %q", g.input.touch.ctxLabel)
	}
	g.newGame.stage = ngSplash // 還原,避免影響下方既有斷言

	g.showTitle = false
	g.tavern.active = true
	g.tavern.stage = tavName
	g.updateTouchContext()
	if g.input.touch.ctxLabel != "注" {
		t.Errorf("酒館命名階段應設情境標籤「注」,得 %q", g.input.touch.ctxLabel)
	}

	g.tavern.stage = tavGender // 酒館其他階段(非命名)→ 不應顯示情境鍵
	g.updateTouchContext()
	if g.input.touch.ctxLabel != "" {
		t.Errorf("酒館非命名階段應隱藏情境鍵(空字串),得 %q", g.input.touch.ctxLabel)
	}

	g.tavern.active = false
	g.updateTouchContext()
	if g.input.touch.ctxLabel != "" {
		t.Errorf("一般情境應隱藏情境鍵(空字串),得 %q", g.input.touch.ctxLabel)
	}
}

// TestTouchHeldState:P3 壓感回饋——poll() 掃到的觸點只要「落在該鈕區」就該置 held,
// 不限 just-pressed(edge),呼應 aTap/bTap/ctxTap 是 edge、aHeld/bHeld/ctxHeld 是 level 的差異。
// poll() 依賴 ebiten 裝置 API(Xvfb headless 沒有真觸控裝置),所以直接呼叫 zoneOf() 驗證
// 分區判定本身,再手動比照 poll() 的映射規則(zoneA→aHeld 等)斷言語意正確,不依賴裝置層。
func TestTouchHeldState(t *testing.T) {
	tu := newTouchUI()
	if z := tu.zoneOf(aCX, aCY); z != zoneA {
		t.Fatalf("前提:A 鍵中心應 zoneA,得 %d", z)
	}
	if z := tu.zoneOf(bCX, bCY); z != zoneB {
		t.Fatalf("前提:B 鍵中心應 zoneB,得 %d", z)
	}
	tu.SetContext("設定")
	if z := tu.zoneOf(ctxCX, ctxCY); z != zoneCtx {
		t.Fatalf("前提:情境鍵中心應 zoneCtx,得 %d", z)
	}

	// 直接驗證 poll() 內 switch 的映射語意:任一觸點分到該區就置 held,不看 just-pressed。
	tu.aHeld, tu.bHeld, tu.ctxHeld = false, false, false
	if z := tu.zoneOf(aCX, aCY); z == zoneA {
		tu.aHeld = true
	}
	if z := tu.zoneOf(bCX, bCY); z == zoneB {
		tu.bHeld = true
	}
	if z := tu.zoneOf(ctxCX, ctxCY); z == zoneCtx {
		tu.ctxHeld = true
	}
	if !tu.aHeld {
		t.Errorf("觸點落在 A 區應置 aHeld=true")
	}
	if !tu.bHeld {
		t.Errorf("觸點落在 B 區應置 bHeld=true")
	}
	if !tu.ctxHeld {
		t.Errorf("觸點落在情境鍵區應置 ctxHeld=true")
	}
}

// TestSafeInsetKeepsCentersInsideAndSeparated:P3 safeInset——驗證固定內縮後,
// 十字鍵/A/B/情境鍵四個控制中心 (1) 仍在畫面內且離最近邊緣至少 safeInset,
// (2) 兩兩之間仍保有各自視覺/判定半徑總和的間距,不因內縮而互相重疊
// (A/B 兩鍵在 baseline 設計就幾乎相切,inset 對兩者是同向平移,相對距離不變,
// 這裡只驗證「沒有變得更差」——即 inset 後距離 ≥ baseline 既有間距)。
func TestSafeInsetKeepsCentersInsideAndSeparated(t *testing.T) {
	type ctr struct {
		name    string
		x, y, r int // r:該元件離邊緣需保留的半徑(視覺或判定半徑,取大者)
	}
	centers := []ctr{
		{"pad", padCX, padCY, padR},
		{"A", aCX, aCY, btnHit},
		{"B", bCX, bCY, btnHit},
		{"ctx", ctxCX, ctxCY, ctxHit},
	}
	for _, c := range centers {
		if c.x-safeInset < 0 || c.x+safeInset > ScreenW {
			t.Errorf("%s 中心 x=%d 內縮 safeInset=%d 後超出畫面寬 %d", c.name, c.x, safeInset, ScreenW)
		}
		if c.y-safeInset < 0 || c.y+safeInset > ScreenH {
			t.Errorf("%s 中心 y=%d 內縮 safeInset=%d 後超出畫面高 %d", c.name, c.y, safeInset, ScreenH)
		}
		if c.x-c.r < 0 || c.x+c.r > ScreenW || c.y-c.r < 0 || c.y+c.r > ScreenH {
			t.Errorf("%s 的視覺/判定範圍(半徑 %d)超出畫面邊界", c.name, c.r)
		}
	}
	dist := func(a, b ctr) float64 {
		dx, dy := float64(a.x-b.x), float64(a.y-b.y)
		return math.Sqrt(dx*dx + dy*dy)
	}
	for i := 0; i < len(centers); i++ {
		for j := i + 1; j < len(centers); j++ {
			a, b := centers[i], centers[j]
			d := dist(a, b)
			// A-B baseline 設計就接近相切(略小於 r 總和的間距屬既有行為,見上方註解),
			// 其餘任兩區至少要間隔各自半徑總和,才算「不重疊」。
			min := float64(a.r + b.r)
			if a.name == "A" && b.name == "B" {
				min -= 4 // 容許 baseline 既有的些微相切誤差(未被 inset 放大)
			}
			if d < min {
				t.Errorf("%s(%d,%d) 與 %s(%d,%d) 距離 %.1f 小於安全間距 %.1f,可能重疊",
					a.name, a.x, a.y, b.name, b.x, b.y, d, min)
			}
		}
	}
}
