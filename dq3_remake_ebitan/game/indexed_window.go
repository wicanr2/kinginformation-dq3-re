package game

import (
	"fmt"
	"image"
	"strconv"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 索引色視窗只實作有限的原始字模／陰影／XOR primitive；所有版面與色號來自 pack。
// 入口與 READY 證據：docs/113；欄位契約：docs/84。
type indexedNewGameRenderer struct {
	geo        gamepack.NewGameGeometry
	style      *gamepack.NewGameRasterLayout
	windows    map[string]gamepack.RawNewGameWindow
	texts      map[string][]uint16
	background []byte
	palette    []dq3data.Color
	pixels     []byte
}

func newIndexedNewGameRenderer(pack *gamepack.Pack, geo gamepack.NewGameGeometry, tx *dq3data.Text, background []byte, palette []dq3data.Color) (*indexedNewGameRenderer, error) {
	if geo.Raster == nil || tx == nil || len(background) != ScreenW*ScreenH || len(palette) != 16 {
		return nil, fmt.Errorf("missing validated layout, font or indexed background")
	}
	r := &indexedNewGameRenderer{geo: geo, style: geo.Raster, windows: map[string]gamepack.RawNewGameWindow{}, texts: map[string][]uint16{},
		background: append([]byte(nil), background...), palette: append([]dq3data.Color(nil), palette...), pixels: make([]byte, len(background))}
	for _, index := range background {
		if int(index) >= len(palette) {
			return nil, fmt.Errorf("background palette index outside palette")
		}
	}
	for _, w := range geo.RawWindows {
		r.windows[w.ID] = w
	}
	for _, c := range r.style.PaletteOverrides {
		r.palette[c.Index] = dq3data.Color{R: c.RGB[0], G: c.RGB[1], B: c.RGB[2]}
	}
	for _, id := range []string{r.style.Menu.TextID, r.style.Header.TextID, r.style.Mode.TextID, r.style.Gender.TextID,
		r.style.Ability.TextID, r.style.ConfirmationPrompt.TextID, r.style.ConfirmationChoice.TextID, r.style.ZhuyinTextID, r.style.AlnumTextID} {
		codes, ok := pack.TextGlyphCodes(id)
		if !ok {
			return nil, fmt.Errorf("missing raster text %q", id)
		}
		for _, code := range codes {
			if code != dq3data.TxtNL {
				if _, ok := tx.Glyph(int(code)); !ok {
					return nil, fmt.Errorf("raster text %q glyph %d missing", id, code)
				}
			}
		}
		r.texts[id] = codes
	}
	labels, ok := pack.NewGameLabels()
	if !ok || len(labels.ChoiceCursor) != 1 {
		return nil, fmt.Errorf("missing selection cursor glyph")
	}
	if _, ok := tx.Glyph(labels.ChoiceCursor[0]); !ok {
		return nil, fmt.Errorf("selection cursor glyph missing")
	}
	if r.style.EquipmentMarkerGlyph == nil {
		return nil, fmt.Errorf("missing equipment marker glyph")
	}
	if _, ok := tx.Glyph(*r.style.EquipmentMarkerGlyph); !ok {
		return nil, fmt.Errorf("equipment marker glyph missing")
	}
	// 狀態機容量沿用既有常數；幾何來自 pack，進入正式 UI 前驗證最遠寫入。
	inside := func(x, y, width, height int) bool {
		return x >= 0 && y >= 0 && x+width <= ScreenW && y+height <= ScreenH
	}
	if !inside(geo.NameText.X+niNameMax*geo.NameText.StepX, geo.NameText.Y, r.style.CursorWidth, r.style.CursorHeight) ||
		!inside(geo.NameComposition.X+2*geo.NameComposition.StepX, geo.NameComposition.Y, dq3data.GlyphPx, dq3data.GlyphPx) {
		return nil, fmt.Errorf("name or composition output outside canvas")
	}
	for _, rows := range []struct {
		anchor gamepack.GeometryAnchor
		hit    gamepack.GeometryRect
		count  int
	}{
		{r.style.MenuCursor, r.style.MenuHit, ngOptCount},
		{r.style.FunctionCursor, r.style.FunctionHit, len(labels.FunctionZhuyin)},
		{r.style.GenderCursor, r.style.GenderHit, 2},
	} {
		last := rows.count - 1
		if !inside(rows.anchor.X, rows.anchor.Y+last*rows.anchor.StepY, dq3data.GlyphPx, dq3data.GlyphPx) ||
			!inside(rows.hit.X, rows.hit.Y+last*rows.anchor.StepY, rows.hit.Width, rows.hit.Height) {
			return nil, fmt.Errorf("selection cursor or touch rows outside canvas")
		}
	}
	return r, nil
}

func (r *indexedNewGameRenderer) opaqueGlyph(tx *dq3data.Text, x, y, code int) {
	bitmap, ok := tx.Glyph(code)
	if !ok {
		return // 靜態字模已在 bootstrap 驗證；動態姓名沿用既有字模有效範圍。
	}
	for row := 0; row < dq3data.GlyphPx; row++ {
		for col := 0; col < dq3data.GlyphPx; col++ {
			index := byte(0)
			if bitmap[row][col] != 0 {
				index = byte(*r.style.FontIndex)
			}
			r.pixels[(y+row)*ScreenW+x+col] = index
		}
	}
}

func (r *indexedNewGameRenderer) text(tx *dq3data.Text, id string, origin image.Point) {
	x, y := origin.X, origin.Y
	for _, code := range r.texts[id] {
		if code == dq3data.TxtNL {
			x, y = origin.X, y+dq3data.GlyphPx
		} else {
			r.opaqueGlyph(tx, x, y, int(code))
			x += dq3data.GlyphPx
		}
	}
}

func (r *indexedNewGameRenderer) shadow(rect image.Rectangle) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for wordX := rect.Min.X; wordX < rect.Max.X; wordX += 16 {
			// EGA word 讀完後 latch 留在第二 byte；兩次寫回共用它。
			var latched [8]byte
			copy(latched[:], r.pixels[y*ScreenW+wordX+8:y*ScreenW+wordX+16])
			for px := 0; px < 16; px++ {
				index := latched[px%8]
				if (px+y-rect.Min.Y)%2 == 0 {
					index = 0
				}
				r.pixels[y*ScreenW+wordX+px] = index
			}
		}
	}
}

func (r *indexedNewGameRenderer) xor(rect image.Rectangle, mask byte, checker bool) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			if !checker || (x-rect.Min.X+y-rect.Min.Y)%2 == 1 {
				r.pixels[y*ScreenW+x] ^= mask
			}
		}
	}
}

func (r *indexedNewGameRenderer) window(tx *dq3data.Text, ref gamepack.RasterWindowRef) {
	w := r.windows[ref.RawWindowID]
	x, y, width, height := w.X*8, w.Y, w.Width*8, w.Height
	offset := r.style.ShadowOffset
	r.shadow(image.Rect(x+offset.X, y+offset.Y, x+offset.X+width, y+offset.Y+height))
	r.text(tx, ref.TextID, image.Pt(x, y))
	r.frame(ref)
}

// 活動外框的高亮是可逆 XOR；新視窗先撤銷前一個外框，姓名等後畫字模也一併切換。
func (r *indexedNewGameRenderer) frame(ref gamepack.RasterWindowRef) {
	w := r.windows[ref.RawWindowID]
	x, y, width, height := w.X*8, w.Y, w.Width*8, w.Height
	if w.Flags&2 != 0 {
		bw, bh := r.style.FrameBandWidth, r.style.FrameBandHeight
		for _, rect := range []image.Rectangle{
			image.Rect(x, y, x+width, y+bh), image.Rect(x, y+height-bh, x+width, y+height),
			image.Rect(x, y+bh, x+bw, y+height-bh), image.Rect(x+width-bw, y+bh, x+width, y+height-bh),
		} {
			r.xor(rect, byte(*r.style.FrameXOR), true)
		}
	}
}

func (r *indexedNewGameRenderer) cursor(x, y int) {
	r.xor(image.Rect(x, y, x+r.style.CursorWidth, y+r.style.CursorHeight), byte(*r.style.CursorXOR), false)
}

func (r *indexedNewGameRenderer) draw(rgba []byte, tx *dq3data.Text, nf *NewGameFlow) {
	copy(r.pixels, r.background)
	s := r.style
	if nf.stage == ngMenu {
		r.window(tx, s.Menu)
		r.opaqueGlyph(tx, s.MenuCursor.X, s.MenuCursor.Y+nf.cursor*s.MenuCursor.StepY, nf.labels.ChoiceCursor[0])
		nf.hits.reset()
		for row := 0; row < ngOptCount; row++ {
			h := s.MenuHit
			nf.hits.add(h.X, h.Y+row*s.MenuCursor.StepY, h.Width, h.Height, row)
		}
	} else if nf.stage == ngGender {
		r.window(tx, s.Gender)
		r.opaqueGlyph(tx, s.GenderCursor.X, s.GenderCursor.Y+nf.gs.cursor*s.GenderCursor.StepY, nf.labels.ChoiceCursor[0])
		nf.gs.hits.reset()
		for row := 0; row < 2; row++ {
			h := s.GenderHit
			nf.gs.hits.add(h.X, h.Y+row*s.GenderCursor.StepY, h.Width, h.Height, row)
		}
	} else if nf.stage == ngReview || nf.stage == ngConfirm {
		r.ability(tx, nf)
	} else {
		ni, geo := &nf.ni, r.geo
		r.window(tx, s.Header)
		textID := s.AlnumTextID
		if ni.nameZhu {
			textID = s.ZhuyinTextID
		}
		r.text(tx, textID, image.Pt(s.GridOrigin.X, s.GridOrigin.Y))
		for i, code := range ni.nameBuf {
			r.opaqueGlyph(tx, geo.NameText.X+i*geo.NameText.StepX, geo.NameText.Y, code)
		}
		r.cursor(geo.NameText.X+ni.namePos*geo.NameText.StepX, geo.NameText.Y)
		if ni.functionFocus {
			r.opaqueGlyph(tx, s.FunctionCursor.X, s.FunctionCursor.Y+ni.functionCursor*s.FunctionCursor.StepY, nf.labels.ChoiceCursor[0])
		} else {
			r.cursor(geo.NameGrid.X+(ni.cursor%geo.NameGrid.Columns)*geo.NameGrid.StepX,
				geo.NameGrid.Y+(ni.cursor/geo.NameGrid.Columns)*geo.NameGrid.StepY)
		}
		if ni.nameZhu && !ni.functionFocus {
			r.window(tx, s.Mode)
			x := geo.NameComposition.X
			ni.eachCompositionGlyph(func(code int) {
				r.opaqueGlyph(tx, x, geo.NameComposition.Y, code)
				x += geo.NameComposition.StepX
			})
		}
		ni.hits.reset()
		for raw := 0; raw < geo.NameGrid.Columns*geo.NameGrid.Rows; raw++ {
			ni.hits.add(geo.NameGrid.X+(raw%geo.NameGrid.Columns)*geo.NameGrid.StepX,
				geo.NameGrid.Y+(raw/geo.NameGrid.Columns)*geo.NameGrid.StepY, geo.NameGrid.StepX, geo.NameGrid.StepY, raw)
		}
		for row := range nf.labels.FunctionZhuyin {
			h := s.FunctionHit
			ni.hits.add(h.X, h.Y+row*s.FunctionCursor.StepY, h.Width, h.Height, niTapFunction(row))
		}
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}

func (r *indexedNewGameRenderer) number(tx *dq3data.Text, field gamepack.NumberField, value int) {
	if field.Digits <= 0 || value < 0 {
		return
	}
	digits := strconv.Itoa(value)
	for i, digit := range digits {
		r.opaqueGlyph(tx, field.X+(field.Digits-len(digits)+i)*dq3data.GlyphPx, field.Y, int(digit-'0'))
	}
}

func (r *indexedNewGameRenderer) ability(tx *dq3data.Text, nf *NewGameFlow) {
	s, geo := r.style, r.geo
	r.window(tx, s.Ability)
	glyphs := func(anchor gamepack.GeometryAnchor, codes []int) {
		for i, code := range codes {
			r.opaqueGlyph(tx, anchor.X+i*dq3data.GlyphPx, anchor.Y, code)
		}
	}
	r.opaqueGlyph(tx, s.EquipmentMarker.X, s.EquipmentMarker.Y, *s.EquipmentMarkerGlyph)
	glyphs(geo.StatsCloth, nf.labels.Cloth)
	glyphs(geo.StatsHero, nf.labels.Hero)
	sex := nf.labels.Male
	if nf.previewGender != 0 {
		sex = nf.labels.Female
	}
	glyphs(geo.StatsSexValue, sex)
	glyphs(geo.StatsName, nf.ni.nameBuf)
	l := geo.Stats
	for _, v := range []struct {
		field gamepack.NumberField
		value int
	}{
		{l.Level.Value, nf.previewLevel}, {l.HP.Value, nf.previewHP}, {l.MP.Value, nf.previewMP},
		{l.Strength.Value, int(nf.preview[stats.STR])}, {l.Agility.Value, int(nf.preview[stats.AGI])},
		{l.Vitality.Value, int(nf.preview[stats.VIT])}, {l.Intelligence.Value, int(nf.preview[stats.INT])},
		{l.Luck.Value, int(nf.preview[stats.LUCK])}, {l.MaxHP.Value, int(nf.preview[stats.HP])},
		{l.MaxMP.Value, int(nf.preview[stats.MP])}, {l.Attack.Value, int(nf.preview[stats.STR])},
		{l.Defense.Value, nf.previewDef}, {l.Experience.Value, nf.previewExp},
	} {
		r.number(tx, v.field, v.value)
	}
	if nf.stage == ngConfirm {
		r.frame(s.Ability)
		r.window(tx, s.ConfirmationPrompt)
		r.frame(s.ConfirmationPrompt)
		r.window(tx, s.ConfirmationChoice)
		a := geo.StatsChoiceCursor
		r.opaqueGlyph(tx, a.X, a.Y+nf.confirmCursor*a.StepY, nf.labels.ChoiceCursor[0])
	}
}
