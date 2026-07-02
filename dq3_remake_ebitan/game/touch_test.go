package game

import "testing"

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
		t.Errorf("標題畫面應設情境標籤「設定」,得 %q", g.input.touch.ctxLabel)
	}

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
