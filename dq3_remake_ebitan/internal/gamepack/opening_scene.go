package gamepack

import (
	"errors"
	"fmt"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

func (p *Pack) validateOpeningScenePresentation() error {
	e, prelude, escort := p.Interface.OpeningScenePresentation, p.Interface.OpeningPrelude, p.Interface.OpeningEscort
	if e == nil {
		if prelude != nil {
			return errors.New("opening scene presentation is required with opening prelude")
		}
		return nil
	}
	if prelude == nil || escort == nil || e.ID == "" || e.PresentationID != prelude.ID ||
		e.CTY < 0 || e.Section < 0 || e.CTY != escort.CTY || e.Section != escort.Section ||
		len(e.TextIDs) == 0 || len(e.TextIDs) > 64 {
		return errors.New("opening scene presentation reference is invalid")
	}
	c, s, w := e.Camera, e.Shadow, prelude.Window
	if c.Mode != "player_anchor" || c.AnchorX < 0 || c.AnchorX >= 20 || c.AnchorY < 0 || c.AnchorY >= 15 ||
		c.ExteriorTile < 0 || c.ExteriorTile > 255 {
		return errors.New("opening scene camera is invalid")
	}
	if s.Mode != "vga_word_latch_and" || s.OffsetX < 0 || s.OffsetX%8 != 0 || s.OffsetY < 0 ||
		w.X+s.OffsetX+w.Width > 640 || w.Y+s.OffsetY+w.Height > 350 {
		return errors.New("opening scene shadow is invalid")
	}
	seen := map[string]bool{}
	for _, id := range e.TextIDs {
		if id == "" || seen[id] || id == prelude.TextID {
			return errors.New("opening scene text sequence is invalid")
		}
		seen[id] = true
	}
	for _, evidence := range []Evidence{e.Evidence, c.Evidence, s.Evidence} {
		if evidence.Level != "D3" {
			return errors.New("opening scene presentation requires D3 evidence")
		}
		if err := validateEvidence(evidence); err != nil {
			return fmt.Errorf("opening scene evidence: %w", err)
		}
	}
	return nil
}

func (p *Pack) validateOpeningSceneRefs() error {
	e := p.Interface.OpeningScenePresentation
	if e == nil {
		return nil
	}
	for _, id := range e.TextIDs {
		codes, ok := p.TextGlyphCodes(id)
		if !ok || len(codes) == 0 {
			return fmt.Errorf("opening scene text %q is missing", id)
		}
		for i := 0; i < len(codes); i++ {
			v := codes[i]
			switch {
			case v < dq3data.GlyphMax, v == dq3data.TxtNL, v == dq3data.TxtNL2, v == dq3data.TxtPage:
			case dq3data.IsVarInsert(v):
				i += p.Interface.OpeningPrelude.VariableCodeWords - 1
				if i >= len(codes) {
					return errors.New("opening scene variable control is incomplete")
				}
			default:
				return errors.New("opening scene text control is unsupported")
			}
		}
	}
	return nil
}

func (p *Pack) OpeningScenePresentation() (*OpeningScenePresentation, bool) {
	if p == nil || p.Interface.OpeningScenePresentation == nil {
		return nil, false
	}
	return p.Interface.OpeningScenePresentation, true
}
