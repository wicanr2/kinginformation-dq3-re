package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 虛擬觸控控制(Android/行動):左下浮動十字鍵 + 右下 A/B。實作 docs/63 規劃。
// 半透明疊在畫面上;多點觸控(十字鍵 + 按鈕同時)。座標用 Ebiten 邏輯座標(640×350)。
const (
	padCX, padCY  = 92, ScreenH - 84 // 十字鍵閒置中心(左下)
	padR          = 58               // 十字鍵作用半徑
	touchDeadzone = 14               // 死區(內回無方向)
	aCX, aCY      = ScreenW - 70, ScreenH - 100
	bCX, bCY      = ScreenW - 148, ScreenH - 64
	btnR          = 32
	btnHit        = 44               // 按鈕觸控判定半徑(略大於視覺)
	ctxCX, ctxCY  = ScreenW - 56, 44 // 情境鍵中心(右上,與左下十字鍵/右下 A・B 不重疊)
	ctxR          = 28
	ctxHit        = 40
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
	dpadDir     int // 這幀十字鍵方向(-1 無,0下 1上 2左 3右)
	aTap, bTap  bool
	ctxLabel    string                    // 情境鍵標籤;空字串=隱藏(由 Game 每幀依情境設)
	ctxTap      bool                      // 這幀情境鍵剛點(edge)
	origin      map[ebiten.TouchID][2]int // 每個十字鍵觸點的浮動中心
	everTouched bool                      // 有過觸控才畫控制(桌面預設不擾)
}

// SetContext:Game 每幀依當前狀態設情境鍵標籤;空字串 = 隱藏(不吃觸點)。
func (t *TouchUI) SetContext(label string) { t.ctxLabel = label }

func newTouchUI() *TouchUI { return &TouchUI{dpadDir: -1, origin: map[ebiten.TouchID][2]int{}} }

func (t *TouchUI) zoneOf(x, y int) touchZone {
	if t.ctxLabel != "" && (x-ctxCX)*(x-ctxCX)+(y-ctxCY)*(y-ctxCY) <= ctxHit*ctxHit {
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
	t.dpadDir, t.aTap, t.bTap, t.ctxTap = -1, false, false, false
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
			if just[id] {
				t.aTap = true
			}
		case zoneB:
			if just[id] {
				t.bTap = true
			}
		case zoneCtx:
			if just[id] {
				t.ctxTap = true
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
	// A / B
	fillCircle(rgba, aCX, aCY, btnR, base, 0.4)
	fillCircle(rgba, bCX, bCY, btnR-6, base, 0.32)
	// 情境鍵(右上,標籤隨情境變義;MVP 不畫字,琥珀色圓明顯區別 A/B)
	if t.ctxLabel != "" {
		amber := dq3data.Color{R: 239, G: 154, B: 60}
		fillCircle(rgba, ctxCX, ctxCY, ctxR, amber, 0.55)
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
