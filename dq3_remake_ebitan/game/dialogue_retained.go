package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"

// retainedTextFlow 重播有限字模／捲動操作，確認後保留既有畫面。
// 參數來自資料包；來源與適用範圍見 docs/188。
type retainedTextFlow struct {
	word, x, row    int
	glyphs          []int
	waiting         bool
	waitFrames      int
	scrolling       bool
	scrollRemaining int
	afterScroll     retainedTextAction
	ops             []retainedTextOp
}

type retainedTextAction uint8

const (
	retainedContinue retainedTextAction = iota
	retainedWait
	retainedReturn
)

type retainedTextOp struct {
	glyph, x, y int
	scroll      bool
}

func (d *Dialogue) usesRetainedRows() bool {
	return d.prelude != nil && d.prelude.TextFlow.Mode == "retained_rows"
}

func (d *Dialogue) tickRetainedRows() {
	if d.retained == nil {
		d.retained = &retainedTextFlow{}
	}
	f := d.retained
	if f.waiting {
		w := d.prelude.WaitIndicator
		f.waitFrames = (f.waitFrames + 1) % (w.HoldFrames(w.VisibleTicks) + w.HoldFrames(w.HiddenTicks))
		return
	}
	d.revealTick++
	hold := d.layout.GlyphHoldFrames
	if f.scrolling {
		hold = d.prelude.TextFlow.ScrollHoldFrames
	}
	if d.revealTick < hold {
		return
	}
	d.revealTick = 0
	for {
		if f.scrollRemaining > 0 {
			f.ops = append(f.ops, retainedTextOp{scroll: true})
			f.scrollRemaining--
			return
		}
		if f.scrolling {
			action := f.afterScroll
			f.scrolling = false
			f.afterScroll = retainedContinue
			if action == retainedWait {
				f.waiting = true
				f.waitFrames = 0
				return
			}
			if action == retainedReturn {
				d.open = false
				return
			}
		}
		if len(f.glyphs) > 0 {
			f.ops = append(f.ops, retainedTextOp{glyph: f.glyphs[0], x: f.x, y: f.row * dq3data.GlyphPx})
			f.glyphs = f.glyphs[1:]
			f.x += d.prelude.GlyphStepX
			d.revealCells++
			return
		}
		if f.word >= len(d.buf) {
			if f.row == d.layout.LinesPerPage-1 {
				f.scrolling = true
				f.scrollRemaining = d.prelude.TextFlow.ScrollSteps
				f.afterScroll = retainedReturn
				continue
			}
			d.open = false
			return
		}
		v := d.buf[f.word]
		f.word++
		switch {
		case v == dq3data.TxtNL || v == dq3data.TxtNL2 || v == dq3data.TxtPage:
			f.x = 0
			if f.row == d.layout.LinesPerPage-1 {
				f.scrolling = true
				f.scrollRemaining = d.prelude.TextFlow.ScrollSteps
				if v == dq3data.TxtPage {
					f.afterScroll = retainedWait
				}
				continue
			}
			f.row++
			if v == dq3data.TxtPage {
				f.waiting = true
				f.waitFrames = 0
				return
			}
		case dq3data.IsVarInsert(v):
			f.word += d.variableCodeWords() - 1
			f.glyphs = append(f.glyphs, d.varGlyphs(v)...)
		default:
			f.glyphs = append(f.glyphs, int(v))
		}
	}
}

func (d *Dialogue) drawWaitIndicator(rgba []byte, fg dq3data.Color) {
	if d.retained == nil || !d.retained.waiting {
		return
	}
	w := d.prelude.WaitIndicator
	code := w.VisibleGlyph
	if d.retained.waitFrames >= w.HoldFrames(w.VisibleTicks) {
		code = w.HiddenGlyph
	}
	glyph, ok := d.tx.Glyph(code)
	if !ok {
		return
	}
	y := d.layout.Y + d.layout.TextInsetY + d.retained.row*dq3data.GlyphPx
	bg := dq3data.Color{R: d.prelude.BackdropRGB[0], G: d.prelude.BackdropRGB[1], B: d.prelude.BackdropRGB[2]}
	for r := 0; r < dq3data.GlyphPx; r++ {
		for c := 0; c < dq3data.GlyphPx; c++ {
			color := bg
			if glyph[r][c] != 0 {
				color = fg
			}
			putPx(rgba, w.X+c, y+r, color)
		}
	}
}

func (d *Dialogue) drawRetainedRows(rgba []byte, fg dq3data.Color) {
	defer d.drawWaitIndicator(rgba, fg)
	if d.retained == nil {
		return
	}
	w := d.layout
	x, y := w.X+w.TextInsetX, w.Y+w.TextInsetY
	width, height := w.Width-2*w.TextInsetX, w.LinesPerPage*dq3data.GlyphPx
	step := d.prelude.TextFlow.ScrollStepPixels
	bg := dq3data.Color{R: d.prelude.BackdropRGB[0], G: d.prelude.BackdropRGB[1], B: d.prelude.BackdropRGB[2]}
	for _, op := range d.retained.ops {
		if op.scroll {
			for r := 0; r < height-step; r++ {
				dst := ((y+r)*ScreenW + x) * 4
				src := ((y+r+step)*ScreenW + x) * 4
				copy(rgba[dst:dst+width*4], rgba[src:src+width*4])
			}
			for r := height - step; r < height; r++ {
				for c := 0; c < width; c++ {
					putPx(rgba, x+c, y+r, bg)
				}
			}
			continue
		}
		glyph, ok := d.tx.Glyph(op.glyph)
		if !ok {
			continue
		}
		for r := 0; r < dq3data.GlyphPx; r++ {
			for c := 0; c < dq3data.GlyphPx; c++ {
				color := bg
				if glyph[r][c] != 0 {
					color = fg
				}
				putPx(rgba, x+op.x+c, y+op.y+r, color)
			}
		}
	}
}

// appendRetainedRecord 維持 caller 的原生文字行計數與已畫操作。
// EOF 未捲動時移至下一行；最後一行的 EOF 已完成捲動。
func (d *Dialogue) appendRetainedRecord(codes []uint16) {
	old := d.retained
	d.openRecord(codes)
	if old != nil {
		row := old.row
		if row < d.layout.LinesPerPage-1 {
			row++
		}
		d.retained = &retainedTextFlow{row: row, ops: old.ops}
	}
}
