package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// opening_cutscene.go implements the shared boot-presentation state machine.
// The frame list, timing and asset identity are pack data; this file does not
// know DQ3 filenames or story text.

func (g *Game) openingReady() bool {
	return g.openingSeq != nil && len(g.openingSeq.Frames) > 0 &&
		len(g.openingPix) == len(g.openingSeq.Frames) &&
		g.openingIndex >= 0 && g.openingIndex < len(g.openingPix) &&
		len(g.openingPix[g.openingIndex]) == ScreenW*ScreenH
}

// StartOpeningCutscene is the normal desktop/mobile boot entry. Tests that
// start directly at a title fixture can omit it; production main calls it
// immediately after NewGame and before the first Ebiten frame.
func (g *Game) StartOpeningCutscene() {
	if !g.showTitle || !g.openingReady() {
		return
	}
	g.openingActive = true
	g.openingIndex = 0
	g.openingFrame = 0
}

func (g *Game) stopOpeningCutscene() {
	g.openingActive = false
	g.openingIndex = 0
	g.openingFrame = 0
}

func (g *Game) advanceOpeningCutscene() {
	if !g.openingReady() {
		g.stopOpeningCutscene()
		return
	}
	frame := g.openingSeq.Frames[g.openingIndex]
	g.openingFrame++
	if g.openingFrame < frame.HoldFrames {
		return
	}
	g.openingFrame = 0
	g.openingIndex++
	if g.openingIndex >= len(g.openingSeq.Frames) {
		g.stopOpeningCutscene()
	}
}

func (g *Game) openingInput(in InputState) bool {
	if !g.openingActive {
		return false
	}
	if g.openingSeq != nil && g.openingSeq.SkipOnInput && titleInputInterruptsAttract(in) {
		g.stopOpeningCutscene()
		return false // process the same key as title splash input
	}
	g.advanceOpeningCutscene()
	return true
}

// openingTick 使用累積有理數取樣，避免每個位置分別取整而漂移。
func (g *Game) openingTick() int {
	timing := g.openingSeq.Frames[g.openingIndex].Timing
	if timing == nil {
		return 0
	}
	return int(int64(g.openingFrame) * int64(timing.RateNumerator) / (60 * int64(timing.RateDenominator)))
}

func (g *Game) openingPixels() []uint8 {
	frame := g.openingSeq.Frames[g.openingIndex]
	background := g.openingPix[g.openingIndex]
	if frame.Overlay == nil {
		return background
	}
	o := frame.Overlay
	tick := g.openingTick() - len(frame.Timing.FadeDeductions)*frame.Timing.FadeInStepTicks
	if tick < 0 {
		return background
	}
	position := tick / o.StepTicks
	if position >= o.Positions() {
		position = o.Positions() - 1
	}
	baseY := o.StartY + position*o.StepY
	copy(g.openingComposite, background)
	for i, sprite := range g.openingSprites[g.openingIndex] {
		placement := o.Sprites[i]
		for y := 0; y < sprite.Height; y++ {
			sy := baseY + placement.Y + y
			if sy < 0 || sy >= ScreenH {
				continue
			}
			for x := 0; x < sprite.Width; x++ {
				value := sprite.Pixels[y*sprite.Width+x]
				if value == 0 && placement.TransparentZero {
					continue
				}
				g.openingComposite[sy*ScreenW+placement.X+x] = value
			}
		}
	}
	return g.openingComposite
}

func (g *Game) openingPalette() []dq3data.Color {
	frame := g.openingSeq.Frames[g.openingIndex]
	if frame.Timing == nil {
		return g.openingPal[g.openingIndex]
	}
	t := frame.Timing
	tick := g.openingTick()
	inTicks := len(t.FadeDeductions) * t.FadeInStepTicks
	animationTicks := 0
	if frame.Overlay != nil {
		animationTicks = frame.Overlay.Positions() * frame.Overlay.StepTicks
	}
	outStart := inTicks + animationTicks + t.HoldTicks
	deduction := 0
	if tick < inTicks {
		deduction = t.FadeDeductions[tick/t.FadeInStepTicks]
	} else if tick >= outStart {
		step := (tick - outStart) / t.FadeOutStepTicks
		if step >= len(t.FadeDeductions) {
			step = len(t.FadeDeductions) - 1
		}
		deduction = t.FadeDeductions[len(t.FadeDeductions)-1-step]
	}
	palette := make([]dq3data.Color, len(g.openingPal[g.openingIndex]))
	// VGA 六位元 DAC：先以原版 consumer 的整數比例量化，再扣色階。
	// 顯示展開採 dosgolem 相同的高位複製；這是顯示近似，不代表類比硬體波形。
	component := func(c uint8) uint8 {
		v := int(c)*63/255 - deduction
		if v < 0 {
			v = 0
		}
		return uint8((v << 2) | (v >> 4))
	}
	for i, c := range g.openingPal[g.openingIndex] {
		palette[i] = dq3data.Color{R: component(c.R), G: component(c.G), B: component(c.B)}
	}
	return palette
}
