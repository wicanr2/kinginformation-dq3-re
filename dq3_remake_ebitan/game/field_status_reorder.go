package game

import "github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"

// Source and READY scope: docs/188. This branch only presents the reviewed
// single-member response; it never performs a speculative party transaction.
func (g *Game) beginSingleMemberReorderPrompt() {
	if g.pack == nil || g.pack.Interface.FieldStatusMenu == nil || len(g.companions) != 0 {
		return
	}
	s := g.pack.Interface.FieldStatusMenu.Reorder
	if s == nil || s.Presentation == nil {
		return
	}
	p := s.Presentation
	codes, ok := g.pack.TextGlyphCodes(s.SingleMemberTextID)
	if !ok {
		return
	}
	frame, ok := g.pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return
	}
	d := Dialogue{tx: g.cmd.tx, layout: p.Window}
	if !d.openRecord(codes) {
		return
	}
	d.prelude = &gamepack.OpeningPrelude{
		Window: p.Window, GlyphStepX: p.GlyphStepX, VariableCodeWords: p.VariableCodeWords,
		ReturnMode: "confirm", TextFlow: s.TextFlow, WaitIndicator: s.WaitIndicator,
		ForegroundRGB: p.ForegroundRGB, BackdropRGB: p.BackdropRGB,
	}
	d.preludeFrame, d.shadow = frame, &p.Shadow
	g.renderFrame()
	g.statusReorderPrompt = &fieldItemPromptState{background: append([]byte(nil), g.rgba...), dialogue: d}
	g.panel, g.cmd.open = panelNone, false
}

func (g *Game) stepStatusReorderPrompt(in InputState) {
	s := g.statusReorderPrompt
	waiting := s.dialogue.waitingForConfirm()
	s.dialogue.Tick()
	if waiting && (in.AnyKeyEdge || in.Confirm || in.Cancel || in.Enter || in.DirEdge >= 0 || in.Tapped) {
		s.dialogue.Advance()
		if !s.dialogue.open {
			g.statusReorderPrompt = nil
			g.panelActor, g.panelCursor = -1, 0
		}
	}
}
