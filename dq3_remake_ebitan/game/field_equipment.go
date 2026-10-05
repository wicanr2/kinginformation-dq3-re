package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
	"io/fs"
)

func validateFieldEquipmentSources(assets fs.FS, p *gamepack.Pack, tx *dq3data.Text) error {
	if p == nil || p.Interface.FieldEquipment == nil || tx == nil {
		return fmt.Errorf("field equipment sources missing")
	}
	s := p.Interface.FieldEquipment
	ids := []string{s.RowTextID, s.FooterTextID, s.NoneTextID, s.PreviewTextID}
	for _, part := range s.Parts {
		ids = append(ids, part.HeaderTextID)
	}
	for _, id := range ids {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Record == nil {
			return fmt.Errorf("equipment text reference missing")
		}
		raw, err := fs.ReadFile(assets, d.Source.File)
		if err != nil {
			return err
		}
		codes := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if len(codes) != len(d.GlyphCodes) {
			return fmt.Errorf("equipment text shape differs")
		}
		for i, c := range codes {
			if int(c) != d.GlyphCodes[i] {
				return fmt.Errorf("equipment text source differs")
			}
			if c < dq3data.GlyphMax {
				if _, ok := tx.Glyph(int(c)); !ok {
					return fmt.Errorf("equipment glyph missing")
				}
			}
		}
	}
	for _, glyph := range []int{s.CursorGlyph, s.WornGlyph} {
		if _, ok := tx.Glyph(glyph); !ok {
			return fmt.Errorf("equipment marker glyph missing")
		}
	}
	return nil
}

// Only the reviewed healthy, eligible, uncursed single-owner route uses this
// modal. Other existing equipment routes retain their previous behavior.
type fieldEquipmentState struct {
	active    bool
	partIndex int
}

func (g *Game) beginNativeEquipment() bool {
	if g.pack == nil || g.pack.Interface.FieldEquipment == nil || len(g.companions) != 0 || g.heroHP <= 0 || g.heroConditions != 0 {
		return false
	}
	s := g.pack.Interface.FieldEquipment
	for _, entry := range g.items.Entries() {
		if entry.Part < 0 {
			continue
		}
		m := g.pack.Characters.ItemStorage.Items[entry.Code]
		if entry.Word&uint16(*g.pack.Characters.ItemStorage.Encoding.CurseMask) != 0 || *m.CursedWhenWorn || !s.CanEquip(entry.Code, g.equipActorClass(0), g.heroGender) {
			return false
		}
	}
	g.fieldEquipment = fieldEquipmentState{active: true}
	g.panel, g.panelActor, g.panelCursor = panelEquip, 0, 0
	return true
}
func (g *Game) nativeEquipmentEntries() []itemstore.Entry {
	s := g.pack.Interface.FieldEquipment
	if !g.fieldEquipment.active || g.fieldEquipment.partIndex >= len(s.Parts) {
		return nil
	}
	var out []itemstore.Entry
	for _, entry := range g.items.Entries() {
		if entry.Part == s.Parts[g.fieldEquipment.partIndex].Part {
			out = append(out, entry)
		}
	}
	return out
}
func (g *Game) stepNativeEquipment(in InputState, tap int) {
	s := g.pack.Interface.FieldEquipment
	entries := g.nativeEquipmentEntries()
	count := len(entries) + 1
	confirm := in.Confirm || in.Enter
	if tap >= 0 && tap < count {
		g.panelCursor, confirm = tap, true
	}
	advance := in.Cancel
	if !advance && confirm {
		if g.panelCursor == len(entries) {
			advance = g.items.UnwearPart(s.Parts[g.fieldEquipment.partIndex].Part)
		} else if g.panelCursor >= 0 && g.panelCursor < len(entries) {
			advance = g.items.Wear(entries[g.panelCursor].Position)
		}
	}
	if advance {
		g.fieldEquipment.partIndex++
		g.panelCursor = 0
		if g.fieldEquipment.partIndex >= len(s.Parts) {
			g.fieldEquipment = fieldEquipmentState{}
			g.panel, g.panelActor, g.cmd.open = panelNone, -1, false
		}
		return
	}
	switch in.DirEdge {
	case 0:
		g.panelCursor = (g.panelCursor + 1) % count
	case 1:
		g.panelCursor = (g.panelCursor + count - 1) % count
	}
}

func (g *Game) drawNativeEquipment(rgba []byte) {
	if g.pack == nil || g.cur == nil || g.newGame.raster == nil || g.cmd.contract == nil {
		return
	}
	s := g.pack.Interface.FieldEquipment
	entries := g.nativeEquipmentEntries()
	g.drawNativeFieldCommandMenu()
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	r.palette[g.cmd.contract.FontIndex] = g.fieldIdleForeground()
	w, pw := s.RawWindow, s.PreviewWindow
	w.Height = (len(entries) + s.FrameRows) * s.RowStep
	pw.Y = w.Y + w.Height
	r.windows = map[string]gamepack.RawNewGameWindow{w.ID: w, pw.ID: pw, g.cmd.contract.RawWindow.ID: g.cmd.contract.RawWindow}
	headerID := s.Parts[g.fieldEquipment.partIndex].HeaderTextID
	header, _ := g.pack.TextGlyphCodes(headerID)
	row, _ := g.pack.TextGlyphCodes(s.RowTextID)
	footer, _ := g.pack.TextGlyphCodes(s.FooterTextID)
	preview, _ := g.pack.TextGlyphCodes(s.PreviewTextID)
	none, _ := g.pack.TextGlyphCodes(s.NoneTextID)
	frame := append([]uint16(nil), header...)
	for i := 0; i <= len(entries); i++ {
		frame = append(frame, dq3data.TxtNL)
		frame = append(frame, row...)
	}
	frame = append(frame, dq3data.TxtNL)
	frame = append(frame, footer...)
	r.texts = map[string][]uint16{headerID: frame, s.PreviewTextID: preview}
	if !r.captureBackground(rgba) {
		return
	}
	r.frame(gamepack.RasterWindowRef{RawWindowID: g.cmd.contract.RawWindow.ID})
	r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: w.ID, TextID: headerID})
	g.panelHits.reset()
	for i := 0; i <= len(entries); i++ {
		var name []int
		if i < len(entries) {
			entry := entries[i]
			name = itemNameGlyphs(g.shop.nameText, entry.Code)
			if g.items.IsWorn(entry) {
				r.opaqueGlyph(g.cmd.tx, s.Worn.X, s.Worn.Y+i*s.RowStep, s.WornGlyph)
			}
		} else {
			for _, c := range none {
				name = append(name, int(c))
			}
		}
		for j, c := range name {
			r.opaqueGlyph(g.cmd.tx, s.Name.X+j*dq3data.GlyphPx, s.Name.Y+i*s.RowStep, c)
		}
		g.panelHits.add(s.Cursor.X, s.Cursor.Y+i*s.RowStep, (w.X+w.Width)*8-s.Cursor.X, s.RowStep, i)
	}
	r.opaqueGlyph(g.cmd.tx, s.Cursor.X, s.Cursor.Y+g.panelCursor*s.RowStep, s.CursorGlyph)
	if g.panelCursor >= 0 && g.panelCursor < len(entries) {
		r.window(g.cmd.tx, gamepack.RasterWindowRef{RawWindowID: pw.ID, TextID: s.PreviewTextID})
		a, d := s.Attack, s.Defense
		a.Y += pw.Y
		d.Y += pw.Y
		r.number(g.cmd.tx, a, g.shop.items.Attack(entries[g.panelCursor].Code))
		r.number(g.cmd.tx, d, g.shop.items.Defense(entries[g.panelCursor].Code))
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}
