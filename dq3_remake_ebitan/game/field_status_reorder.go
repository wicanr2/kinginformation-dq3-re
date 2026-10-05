package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// Source and READY scope: docs/188. This branch only presents the reviewed
// single-member response; it never performs a speculative party transaction.
func (g *Game) beginSingleMemberReorderPrompt() {
	if g.pack == nil || g.pack.Interface.FieldStatusMenu == nil || len(g.companions) != 0 {
		return
	}
	s := g.pack.Interface.FieldStatusMenu.Reorder
	if s == nil {
		return
	}
	g.renderFrame()
	g.beginFieldMessagePrompt(s.SingleMemberTextID, s.Presentation, s.TextFlow, s.WaitIndicator, nil)
}

func (g *Game) beginFieldMessagePrompt(textID string, p *gamepack.FieldItemPrompt, flow gamepack.RetainedTextFlow, indicator gamepack.OpeningWaitIndicator, variables map[uint16][]int) bool {
	return g.beginFieldMessageRecords([]string{textID}, p, flow, indicator, variables)
}

// Each reviewed record starts on a new line, as in the native text consumer.
func (g *Game) beginFieldMessageRecords(textIDs []string, p *gamepack.FieldItemPrompt, flow gamepack.RetainedTextFlow, indicator gamepack.OpeningWaitIndicator, variables map[uint16][]int) bool {
	if g.pack == nil || p == nil || len(textIDs) == 0 {
		return false
	}
	var codes []uint16
	for i, id := range textIDs {
		record, ok := g.pack.TextGlyphCodes(id)
		if !ok || len(record) == 0 {
			return false
		}
		if i > 0 {
			codes = append(codes, dq3data.TxtNL)
		}
		codes = append(codes, record...)
	}
	frame, ok := g.pack.TextGlyphCodes(p.FrameTextID)
	if !ok {
		return false
	}
	d := Dialogue{tx: g.cmd.tx, layout: p.Window}
	if !d.openRecord(codes) {
		return false
	}
	d.varGlyph = variables
	column := 0
	for _, code := range codes {
		if code == dq3data.TxtNL || code == dq3data.TxtNL2 || code == dq3data.TxtPage {
			column = 0
			continue
		}
		if dq3data.IsVarInsert(code) {
			column += len(d.varGlyphs(code))
		} else {
			column++
		}
		if column <= 0 || p.Window.TextInsetX+(column-1)*p.GlyphStepX+dq3data.GlyphPx > p.Window.Width-p.Window.TextInsetX {
			return false
		}
	}
	d.prelude = &gamepack.OpeningPrelude{
		Window: p.Window, GlyphStepX: p.GlyphStepX, VariableCodeWords: p.VariableCodeWords,
		ReturnMode: "confirm", TextFlow: flow, WaitIndicator: indicator,
		ForegroundRGB: p.ForegroundRGB, BackdropRGB: p.BackdropRGB,
	}
	d.preludeFrame, d.shadow = frame, &p.Shadow
	g.fieldMessagePrompt = &fieldItemPromptState{background: append([]byte(nil), g.rgba...), dialogue: d}
	g.panel, g.cmd.open = panelNone, false
	return true
}

func (g *Game) stepFieldMessagePrompt(in InputState) {
	s := g.fieldMessagePrompt
	waiting := s.dialogue.waitingForConfirm()
	s.dialogue.Tick()
	if waiting && (in.AnyKeyEdge || in.Confirm || in.Cancel || in.Enter || in.DirEdge >= 0 || in.Tapped) {
		s.dialogue.Advance()
		if !s.dialogue.open {
			g.fieldMessagePrompt = nil
			g.panelActor, g.panelCursor = -1, 0
		}
	}
}
