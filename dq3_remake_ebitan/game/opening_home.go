package game

import (
	"encoding/binary"
	"fmt"
	"image"
	"math/bits"
	"time"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// The finite selector owns a separate generator. Tests seed it before input;
// ordinary sessions use a platform clock approximation of the BIOS tick word.
type homePictureSelection struct {
	active                         bool
	stage, round, cursor, question int
	choices, expected              []int
	options                        [][]int
	frames                         []dq3data.Frame
	palette                        []dq3data.Color
	seed                           *uint16
}

func (g *Game) loadHomePicture(read func(string) []byte) error {
	s := g.openingEscort.Home.Picture
	raw := read(s.AssetKey)
	if len(raw) < 6 {
		return fmt.Errorf("opening picture header missing")
	}
	width, height, count := int(binary.LittleEndian.Uint16(raw)), int(binary.LittleEndian.Uint16(raw[2:])), int(binary.LittleEndian.Uint16(raw[4:]))
	stride := width * height * 5
	if width != dq3data.CharW/8 || height != dq3data.CharH || count <= 0 || len(raw) != 6+count*stride || count%s.TripletSize != 0 {
		return fmt.Errorf("opening picture archive shape invalid")
	}
	if count > s.RandomMask+1 {
		return fmt.Errorf("opening picture random range invalid")
	}
	for _, q := range s.Questions {
		for _, id := range q {
			if id >= count {
				return fmt.Errorf("opening picture reference outside archive")
			}
		}
	}
	g.homeSelection.frames = make([]dq3data.Frame, count)
	for id := 0; id < count; id++ {
		f := &g.homeSelection.frames[id]
		base := 6 + id*stride
		planeBytes := width * height
		for y := 0; y < height; y++ {
			for x := 0; x < width*8; x++ {
				off := y*width + x/8
				bit := byte(128 >> uint(x%8))
				for plane := 0; plane < 4; plane++ {
					if raw[base+plane*planeBytes+off]&bit != 0 {
						f.Px[y][x] |= 1 << uint(3-plane)
					}
				}
				f.Opaque[y][x] = raw[base+4*planeBytes+off]&bit == 0
			}
		}
	}
	pal := dq3data.DecodePalette(read(s.PaletteAssetKey), 256)
	start := s.PaletteBank * 16
	if len(pal) < start+16 {
		return fmt.Errorf("opening picture palette bank missing")
	}
	g.homeSelection.palette = append([]dq3data.Color(nil), pal[start:start+16]...)
	for _, c := range s.PaletteOverrides {
		g.homeSelection.palette[c.Index] = dq3data.Color{R: c.RGB[0], G: c.RGB[1], B: c.RGB[2]}
	}
	if _, ok := g.dlg.tx.Glyph(s.BlankGlyph); !ok {
		return fmt.Errorf("opening picture blank glyph missing")
	}
	for _, id := range []string{s.TopTextID, s.BodyTextID, s.BottomTextID} {
		codes, ok := g.pack.TextGlyphCodes(id)
		if !ok {
			return fmt.Errorf("opening picture text missing")
		}
		for _, c := range codes {
			if _, ok := g.dlg.tx.Glyph(int(c)); !ok {
				return fmt.Errorf("opening picture glyph missing")
			}
		}
	}
	return nil
}
func (g *Game) startHomePicture() error {
	if g.openingEscort == nil || g.openingEscort.Home == nil || len(g.homeSelection.frames) == 0 {
		return fmt.Errorf("opening picture unavailable")
	}
	h := &g.homeSelection
	s := g.openingEscort.Home.Picture
	if g.cur == nil || g.cur.blk == nil || s.BackgroundTile >= g.cur.blk.Count {
		return fmt.Errorf("opening picture background outside archive")
	}
	seed := uint16(time.Now().UnixMilli() * 182065 / 10000000)
	if h.seed != nil {
		seed = *h.seed
	}
	draw := func() uint16 { seed = bits.RotateLeft16(seed+s.RandomAdd, s.RandomRotate); return seed }
	sample := func(limit int) (int, error) {
		for i := 0; i < 65536; i++ {
			v := int(draw()) & s.RandomMask
			if v < limit {
				return v, nil
			}
		}
		return 0, fmt.Errorf("opening picture generator rejected all candidates")
	}
	question, err := sample(len(s.Questions))
	if err != nil {
		return err
	}
	targets := s.Questions[question]
	options := make([][]int, len(targets))
	expected := make([]int, len(targets))
	for round, target := range targets {
		randomFirst := draw()&1 != 0
		random, err := sample(len(h.frames))
		if err != nil {
			return err
		}
		bases := []int{target / s.TripletSize * s.TripletSize, random / s.TripletSize * s.TripletSize}
		if randomFirst {
			bases[0], bases[1] = bases[1], bases[0]
		}
		for _, base := range bases {
			for j := 0; j < s.TripletSize; j++ {
				options[round] = append(options[round], base+j)
			}
		}
		for i, id := range options[round] {
			if id == target {
				expected[round] = i + 1
				break
			}
		}
	}
	h.question, h.options, h.expected = question, options, expected
	h.choices = make([]int, len(targets))
	h.stage, h.round, h.cursor, h.active = 0, 0, 0, true
	return nil
}
func (g *Game) stepHomePicture(in InputState) {
	h := &g.homeSelection
	s := g.openingEscort.Home.Picture
	confirm := in.Confirm || in.Enter
	acknowledge := in.AnyKeyEdge || confirm || in.Cancel || in.DirEdge >= 0 || in.Toggle || in.Settings || in.Help
	switch h.stage {
	case 0:
		if acknowledge {
			h.stage = 1
		}
	case 1:
		if confirm || in.Cancel {
			h.choices[h.round] = h.cursor + 1
			h.round++
			h.cursor = 0
			if h.round == len(h.options) {
				h.stage = 2
			}
			return
		}
		if in.DirEdge >= 0 {
			delta := 1
			if in.DirEdge == 1 || in.DirEdge == 2 {
				delta = -1
			}
			h.cursor = (h.cursor + delta + s.OptionCount) % s.OptionCount
		}
	case 2:
		if acknowledge {
			h.active = false
			g.homeAwait = true
			g.openingEscortPhase = 4
			g.cd = 0
		}
	}
}
func (g *Game) tryHomeApproach() bool {
	e := g.openingEscort
	if !g.homeAwait || e == nil || !g.inTown || g.cur == nil || g.curCty != e.CTY || g.cur.sec != e.Section {
		return false
	}
	h := e.Home
	if h.ApproachFrame.Leader.X >= g.cur.w || h.ApproachFrame.Leader.Y >= g.cur.h {
		return false
	}
	if !g.storyFlag(h.ApproachFlag) || g.cur.hiMap == nil || int(g.cur.hiMap[g.py*g.cur.w+g.px]&31) != h.ApproachSubid || h.ApproachSubid > len(g.cur.specialHandlers) || g.cur.specialHandlers[h.ApproachSubid-1] != h.ApproachHandlerRaw {
		return false
	}
	match := false
	for _, p := range h.ApproachTiles {
		if g.px == p.X && g.py == p.Y {
			match = true
		}
	}
	if !match {
		return false
	}
	last := e.Frames[len(e.Frames)-1]
	idx := -1
	for i, n := range g.cur.npcs {
		if n.recordIndex == *h.LeaderRecord && n.x == last.Leader.X && n.y == last.Leader.Y {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	g.openingEscortNPC = idx
	g.openingEscortPhase = 5
	g.openingEscortTick = 0
	g.applyOpeningEscortFrame(h.ApproachFrame)
	return true
}
func (g *Game) drawHomePicture() {
	h := &g.homeSelection
	if !h.active {
		return
	}
	s := g.openingEscort.Home.Picture
	// Reuse indexed glyph, number and latch primitives with pack-owned geometry.
	font := s.FontIndex
	r := indexedNewGameRenderer{pixels: make([]byte, ScreenW*ScreenH), palette: h.palette, texts: map[string][]uint16{}}
	r.style = &gamepack.NewGameRasterLayout{FontIndex: &font}
	for _, id := range []string{s.TopTextID, s.BodyTextID, s.BottomTextID} {
		r.texts[id], _ = g.pack.TextGlyphCodes(id)
	}
	base := g.worldPal[s.PaletteBank*16 : s.PaletteBank*16+16]
	colors := map[dq3data.Color]byte{}
	for i, c := range base {
		colors[c] = byte(i)
	}
	for i := range r.pixels {
		off := i * 4
		r.pixels[i] = colors[dq3data.Color{R: g.rgba[off], G: g.rgba[off+1], B: g.rgba[off+2]}]
	}
	w := s.Window
	r.shadow(image.Rect(w.X+s.ShadowOffset.X, w.Y+s.ShadowOffset.Y, w.X+w.Width+s.ShadowOffset.X, w.Y+w.Height+s.ShadowOffset.Y))
	r.text(g.dlg.tx, s.TopTextID, image.Pt(w.X, w.Y))
	for i := 0; i < s.BodyRows; i++ {
		r.text(g.dlg.tx, s.BodyTextID, image.Pt(w.X, w.Y+(i+1)*16))
	}
	r.text(g.dlg.tx, s.BottomTextID, image.Pt(w.X, w.Y+(s.BodyRows+1)*16))
	outline := func(b gamepack.HomeRect, mask byte) {
		r.xor(image.Rect(b.X, b.Y, b.X+b.Width, b.Y+1), mask, false)
		r.xor(image.Rect(b.X, b.Y+b.Height-1, b.X+b.Width, b.Y+b.Height), mask, false)
		r.xor(image.Rect(b.X, b.Y+1, b.X+1, b.Y+b.Height-1), mask, false)
		r.xor(image.Rect(b.X+b.Width-1, b.Y+1, b.X+b.Width, b.Y+b.Height-1), mask, false)
	}
	for _, b := range s.Boxes {
		outline(b, byte(s.FontIndex))
	}
	number := func(field gamepack.HomeNumber, value int) {
		for i := 0; i < field.Digits; i++ {
			r.opaqueGlyph(g.dlg.tx, field.X+i*16, field.Y, s.BlankGlyph)
		}
		r.number(g.dlg.tx, gamepack.NumberField{X: field.X, Y: field.Y, Digits: field.Digits}, value)
	}
	picture := func(id, x, y int) {
		f := h.frames[id]
		background := g.cur.blk.Tile(s.BackgroundTile)
		for cy := 0; cy < dq3data.CharH; cy++ {
			for cx := 0; cx < dq3data.CharW; cx++ {
				c := f.Px[cy][cx]
				if !f.Opaque[cy][cx] {
					c |= background[cy][cx]
				}
				r.pixels[(y+cy)*ScreenW+x+cx] = c
			}
		}
	}
	if h.stage == 0 {
		values := append(append([]int(nil), h.expected...), s.PreviewValues...)
		for i, v := range values {
			f := s.PreviewField
			f.Y += i * s.PreviewStepY
			number(f, v)
		}
	} else {
		number(s.QuestionField, h.question)
		round := h.round
		if round == len(h.options) {
			round--
		}
		for i, id := range h.options[round] {
			picture(id, s.OptionOrigin.X+i*s.StepX, s.OptionOrigin.Y)
		}
		if h.stage == 1 {
			b := s.Cursor
			b.X += h.cursor * s.StepX
			outline(b, byte(s.CursorXOR))
		}
		for i, choice := range h.choices {
			if choice > 0 {
				picture(h.options[i][choice-1], s.SelectedOrigin.X+i*s.StepX, s.SelectedOrigin.Y)
			}
		}
	}
	drawIndexedPCX(g.rgba, r.pixels, r.palette)
}
