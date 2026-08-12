package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 虛擬觸控控制(Android/行動):左下浮動十字鍵 + 右下 A/B。實作 docs/63 規劃。
// 半透明疊在畫面上;多點觸控(十字鍵 + 按鈕同時)。座標用 Ebiten 邏輯座標(640×350)。
//
// P3 打磨(docs/64 §4):
//  1. 壓感回饋 —— aHeld/bHeld/ctxHeld(見 TouchUI 欄位 + poll())驅動 draw() 提高 alpha/放大。
//  2. safeInset —— 下面用「原始邊距 + safeInset」重算各元件中心,固定內縮避開瀏海/手勢列/圓角。
//     真正的 per-device 安全區(瀏海實際挖孔座標)要靠 mobile binding 餵 insets,屬 P4 真機。
//  3. DeviceScaleFactor —— 特意不接。在「固定 640×350 邏輯畫布 + Ebiten letterbox」模型下,
//     Ebiten 已把整個邏輯畫布依視窗大小等比縮放到實體螢幕,遊戲端看不到「這台裝置實際多少
//     物理 px/mm」;ebiten.Monitor().DeviceScaleFactor() 只反映 backing-store 像素密度
//     (points→physical px 的比例),不等於「這個邏輯半徑在這支手機上有幾 mm」——沒有實體螢幕
//     尺寸/PPI 就算讀到 scale factor 也無法推回「按鈕是否 ≥ 9mm」,硬套只會是猜的常數,不算
//     有意義的 best-effort。P4 真機驗證時直接量測/回報實體按鈕尺寸比空談 scale factor 準確。
const (
	safeInset = 12 // 邊緣安全內縮(粗略保護;真實安全區見上方 P3-2 說明)

	padMarginX, padMarginY = 92, 84 // 十字鍵閒置中心:原始(未 inset)左邊距 / 下邊距
	padCX, padCY           = padMarginX + safeInset, ScreenH - padMarginY - safeInset
	padR                   = 58 // 十字鍵作用半徑
	touchDeadzone          = 14 // 死區(內回無方向)

	aMarginX, aMarginY = 70, 100
	aCX, aCY           = ScreenW - aMarginX - safeInset, ScreenH - aMarginY - safeInset
	bMarginX, bMarginY = 148, 64
	bCX, bCY           = ScreenW - bMarginX - safeInset, ScreenH - bMarginY - safeInset
	btnR               = 32
	btnHit             = 44 // 按鈕觸控判定半徑(略大於視覺;P1/P2 已驗證,不動)

	ctxMarginX, ctxMarginY = 56, 44 // 情境鍵中心(右上,與左下十字鍵/右下 A・B 不重疊)
	ctxCX, ctxCY           = ScreenW - ctxMarginX - safeInset, ctxMarginY + safeInset
	ctxR                   = 28
	ctxHit                 = 40
)

// touchContext 是跨版本的觸控情境，不攜帶任何版本專屬可見文字。
// 文字與 glyph 仍由 game-pack／遊戲畫面提供；這裡只需要決定情境鍵是否存在及其動作。
type touchContext uint8

const (
	touchContextNone touchContext = iota
	touchContextSettings
	touchContextToggle
)

type touchZone int

const (
	zoneNone touchZone = iota
	zoneDpad
	zoneA
	zoneB
	zoneCtx
)

// TouchUI 是一幀的觸控輸入結果 + 繪製狀態。
type TouchUI struct {
	dpadDir               int // 這幀十字鍵方向(-1 無,0下 1上 2左 3右)
	aTap, bTap            bool
	aHeld, bHeld, ctxHeld bool                      // 這幀是否有觸點壓在該鈕區(不限 just-pressed;P3 壓感回饋用)
	ctx                   touchContext              // 情境鍵情境;none=隱藏(由 Game 每幀依狀態設)
	ctxTap                bool                      // 這幀情境鍵剛點(edge)
	tapX, tapY            int                       // 落在 zoneNone 的 just-pressed 觸點座標(選單直接點選,P2)
	tapEdge               bool                      // 這幀有 zoneNone 剛點(edge);每幀 reset
	origin                map[ebiten.TouchID][2]int // 每個十字鍵觸點的浮動中心
	everTouched           bool                      // 有過觸控才畫控制(桌面預設不擾)
}

// SetContext:Game 每幀依當前狀態設跨版本情境;none = 隱藏(不吃觸點)。
func (t *TouchUI) SetContext(ctx touchContext) { t.ctx = ctx }

func newTouchUI() *TouchUI { return &TouchUI{dpadDir: -1, origin: map[ebiten.TouchID][2]int{}} }

func (t *TouchUI) zoneOf(x, y int) touchZone {
	if t.ctx != touchContextNone && (x-ctxCX)*(x-ctxCX)+(y-ctxCY)*(y-ctxCY) <= ctxHit*ctxHit {
		return zoneCtx
	}
	if (x-aCX)*(x-aCX)+(y-aCY)*(y-aCY) <= btnHit*btnHit {
		return zoneA
	}
	if (x-bCX)*(x-bCX)+(y-bCY)*(y-bCY) <= btnHit*btnHit {
		return zoneB
	}
	if x < ScreenW/2 && y > ScreenH/2 { // 左下半 → 十字鍵
		return zoneDpad
	}
	return zoneNone
}

// poll:掃所有觸點,分區,產生這幀的方向(held)+ A/B(edge)。多點觸控。
func (t *TouchUI) poll() {
	t.dpadDir, t.aTap, t.bTap, t.ctxTap, t.tapEdge = -1, false, false, false, false
	t.aHeld, t.bHeld, t.ctxHeld = false, false, false
	ids := ebiten.AppendTouchIDs(nil)
	if len(ids) > 0 {
		t.everTouched = true
	}
	just := map[ebiten.TouchID]bool{}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		just[id] = true
	}
	live := map[ebiten.TouchID]bool{}
	for _, id := range ids {
		live[id] = true
	}
	for id := range t.origin { // 放開的十字鍵觸點清掉浮動中心
		if !live[id] {
			delete(t.origin, id)
		}
	}
	for _, id := range ids {
		x, y := ebiten.TouchPosition(id)
		switch t.zoneOf(x, y) {
		case zoneDpad:
			o, ok := t.origin[id]
			if !ok {
				o = [2]int{x, y}
				t.origin[id] = o // 首次觸該區 → 定浮動中心
			}
			if d := quantize4(x-o[0], y-o[1]); d >= 0 {
				t.dpadDir = d
			}
		case zoneA:
			t.aHeld = true
			if just[id] {
				t.aTap = true
			}
		case zoneB:
			t.bHeld = true
			if just[id] {
				t.bTap = true
			}
		case zoneCtx:
			t.ctxHeld = true
			if just[id] {
				t.ctxTap = true
			}
		case zoneNone: // 選單直接點選(P2):非控制區的 just-pressed 觸點 → 記座標,交上層對 hitList
			if just[id] {
				t.tapX, t.tapY, t.tapEdge = x, y, true
			}
		}
	}
}

// quantize4:位移向量 → 4 向(主軸決定;死區內回 -1)。facing 碼 0下 1上 2左 3右。
func quantize4(dx, dy int) int {
	if dx*dx+dy*dy < touchDeadzone*touchDeadzone {
		return -1
	}
	if abs(dx) > abs(dy) {
		if dx < 0 {
			return 2
		}
		return 3
	}
	if dy < 0 {
		return 1
	}
	return 0
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// draw:半透明疊繪十字鍵 + A/B(有觸控過才畫,免擾桌面鍵盤玩家)。
func (t *TouchUI) draw(rgba []byte) {
	if !t.everTouched {
		return
	}
	base := dq3data.Color{R: 230, G: 230, B: 240}
	hot := dq3data.Color{R: 255, G: 224, B: 32}
	// 十字鍵:底環 + 4 方向點(active 方向高亮)
	fillCircle(rgba, padCX, padCY, padR, base, 0.18)
	dirPt := [4][2]int{{padCX, padCY + 34}, {padCX, padCY - 34}, {padCX - 34, padCY}, {padCX + 34, padCY}}
	for d := 0; d < 4; d++ {
		col, a := base, 0.4
		if t.dpadDir == d {
			col, a = hot, 0.85
		}
		fillCircle(rgba, dirPt[d][0], dirPt[d][1], 13, col, a)
	}
	// A / B(壓感回饋:held 時提高 alpha + 微放大,呼應十字鍵方向高亮風格)
	aAlpha, aRad := 0.4, btnR
	if t.aHeld {
		aAlpha, aRad = 0.8, btnR+4
	}
	fillCircle(rgba, aCX, aCY, aRad, base, aAlpha)
	bAlpha, bRad := 0.32, btnR-6
	if t.bHeld {
		bAlpha, bRad = 0.8, btnR-6+4
	}
	fillCircle(rgba, bCX, bCY, bRad, base, bAlpha)
	// 情境鍵(右上,語意隨情境變義;以通用圖示區分設定／切換,不硬寫版本文字)
	if t.ctx != touchContextNone {
		amber := dq3data.Color{R: 239, G: 154, B: 60}
		ctxAlpha, ctxRad := 0.55, ctxR
		if t.ctxHeld {
			ctxAlpha, ctxRad = 0.85, ctxR+4
		}
		fillCircle(rgba, ctxCX, ctxCY, ctxRad, amber, ctxAlpha)
		drawContextIcon(rgba, t.ctx, ctxCX, ctxCY)
	}
}

// drawContextIcon 只畫跨版本的幾何圖示，不把「設定／注」等文字硬寫進共用 Go renderer。
func drawContextIcon(rgba []byte, ctx touchContext, cx, cy int) {
	ink := dq3data.Color{R: 42, G: 42, B: 52}
	switch ctx {
	case touchContextSettings:
		// 八個齒點 + 中央孔：低解析度下仍比純色圓容易辨識。
		fillCircle(rgba, cx, cy, 5, ink, 0.78)
		for _, p := range [][2]int{{0, -9}, {0, 9}, {-9, 0}, {9, 0}, {-6, -6}, {6, -6}, {-6, 6}, {6, 6}} {
			fillCircle(rgba, cx+p[0], cy+p[1], 2, ink, 0.78)
		}
		fillCircle(rgba, cx, cy, 2, dq3data.Color{R: 239, G: 154, B: 60}, 0.9)
	case touchContextToggle:
		// 兩條相反方向箭頭，代表英數／注音層切換；不依賴任何字型 glyph。
		drawTouchSegment(rgba, cx-10, cy-4, cx+8, cy-4, ink, 0.86)
		drawTouchSegment(rgba, cx+10, cy+4, cx-8, cy+4, ink, 0.86)
		for _, p := range [][2]int{{cx + 8, cy - 7}, {cx + 8, cy - 1}, {cx - 8, cy + 1}, {cx - 8, cy + 7}} {
			fillCircle(rgba, p[0], p[1], 2, ink, 0.86)
		}
	}
}

func drawTouchSegment(rgba []byte, x0, y0, x1, y1 int, c dq3data.Color, a float64) {
	dx, dy := x1-x0, y1-y0
	n := abs(dx)
	if abs(dy) > n {
		n = abs(dy)
	}
	if n == 0 {
		blendPx(rgba, x0, y0, c, a)
		return
	}
	for i := 0; i <= n; i++ {
		x := x0 + dx*i/n
		y := y0 + dy*i/n
		blendPx(rgba, x, y, c, a)
		blendPx(rgba, x+1, y, c, a)
	}
}

// fillCircle:以 alpha 混合畫實心圓。
func fillCircle(rgba []byte, cx, cy, r int, c dq3data.Color, a float64) {
	for y := -r; y <= r; y++ {
		for x := -r; x <= r; x++ {
			if x*x+y*y <= r*r {
				blendPx(rgba, cx+x, cy+y, c, a)
			}
		}
	}
}

func blendPx(rgba []byte, x, y int, c dq3data.Color, a float64) {
	if x < 0 || x >= ScreenW || y < 0 || y >= ScreenH {
		return
	}
	o := (y*ScreenW + x) * 4
	inv := 1 - a
	rgba[o] = uint8(float64(rgba[o])*inv + float64(c.R)*a)
	rgba[o+1] = uint8(float64(rgba[o+1])*inv + float64(c.G)*a)
	rgba[o+2] = uint8(float64(rgba[o+2])*inv + float64(c.B)*a)
	rgba[o+3] = 255
}
