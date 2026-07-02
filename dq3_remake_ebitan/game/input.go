package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// InputState 是「一幀的抽象動作」——遊戲邏輯只依賴這個,不知也不管來源(鍵盤 / 觸控)。
// 對齊 docs/63:先抽象再綁定,讓桌面/WASM/手機共用同一套邏輯。
type InputState struct {
	DirHeld  int  // 按住的方向(走動),-1 無;0下 1上 2左 3右
	DirEdge  int  // 剛按下的方向(選單導覽),-1 無
	Confirm  bool // A:確定 / 對話 / 選定(edge)
	Cancel   bool // B:取消 / 出城(edge)
	Enter    bool // 進城(鍵盤 Enter,debug;edge)
	Toggle   bool // Tab:英數 ↔ 注音 切換(酒館命名;edge)
	Settings bool // S:開設定選單(標題畫面;edge)
	CtxTap   bool // 觸控情境鍵剛點(edge);意義隨情境由呼叫端解讀(設定/切換…)
}

// Input 合流鍵盤 + 觸控兩個來源。
type Input struct {
	touch        *TouchUI
	prevTouchDir int
}

func newInput() *Input { return &Input{touch: newTouchUI(), prevTouchDir: -1} }

// Poll:讀這一幀的鍵盤 + 觸控,OR 合併成抽象 InputState。
func (ip *Input) Poll() InputState {
	s := InputState{DirHeld: -1, DirEdge: -1}
	// 鍵盤
	if d := keyDirHeld(); d >= 0 {
		s.DirHeld = d
	}
	if d := dpadEdge(); d >= 0 {
		s.DirEdge = d
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.Confirm = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.Cancel = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		s.Enter = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		s.Toggle = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.Settings = true
	}
	// 觸控(蓋過/補上鍵盤):先偵測(poll,吃真實裝置),再映射進 InputState(applyTouch)。
	ip.touch.poll()
	ip.applyTouch(&s)
	return s
}

// applyTouch:把這一幀觸控偵測結果(dpadDir/aTap/bTap/ctxTap)映射進 InputState。
// 拆出 Poll() 是為了讓映射邏輯本身可測——poll() 依賴 ebiten 真實觸控裝置 API,
// headless 測試環境模擬不了 just-pressed 事件,但映射規則(欄位 → InputState)可以。
func (ip *Input) applyTouch(s *InputState) {
	if ip.touch.dpadDir >= 0 {
		s.DirHeld = ip.touch.dpadDir
		if ip.touch.dpadDir != ip.prevTouchDir { // 觸控方向變化 = 選單導覽 edge
			s.DirEdge = ip.touch.dpadDir
		}
	}
	ip.prevTouchDir = ip.touch.dpadDir
	if ip.touch.aTap {
		s.Confirm = true
	}
	if ip.touch.bTap {
		s.Cancel = true
	}
	if ip.touch.ctxTap {
		s.CtxTap = true
	}
}

func keyDirHeld() int {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		return 0
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		return 1
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		return 2
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		return 3
	}
	return -1
}

// dirDelta:facing 碼 → tile 位移。0下 1上 2左 3右。
func dirDelta(dir int) (int, int) {
	switch dir {
	case 0:
		return 0, 1
	case 1:
		return 0, -1
	case 2:
		return -1, 0
	case 3:
		return 1, 0
	}
	return 0, 0
}
