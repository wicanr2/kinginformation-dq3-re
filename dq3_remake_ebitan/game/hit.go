package game

// hit.go:選單直接點選共用 hit-box 工具(P2,docs/64 §2)。
// 單一真相原則:幾何只在各選單的 draw() 裡算一次,順手記進 hitList;
// input 端只比對記好的 box,不重算幾何。座標用螢幕邏輯座標(640×350),與 InputState.TapX/TapY 同系。
type hitBox struct{ x, y, w, h, idx int }

type hitList []hitBox

// reset:每次 draw() 開頭呼叫,清空上一幀留下的 box(避免舊幾何殘留誤命中)。
func (h *hitList) reset() { *h = (*h)[:0] }

// add:記一個可點區塊(x,y,w,h 為螢幕邏輯座標矩形;idx 為選中後對應的選項/格 index)。
func (h *hitList) add(x, y, w, ht, idx int) { *h = append(*h, hitBox{x, y, w, ht, idx}) }

// at:回觸點 (px,py) 命中的第一個 box 的 idx;無命中回 -1。半開區間 [x,x+w) × [y,y+h)。
func (h hitList) at(px, py int) int {
	for _, b := range h {
		if px >= b.x && px < b.x+b.w && py >= b.y && py < b.y+b.h {
			return b.idx
		}
	}
	return -1
}
