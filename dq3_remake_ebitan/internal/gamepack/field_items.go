package gamepack

import (
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
)

// FieldItems is the reviewed native item-list contract from docs/188.
type FieldItems struct {
	ID           string           `json:"id"`
	RawWindow    RawNewGameWindow `json:"raw_window"`
	ActionWindow RawNewGameWindow `json:"action_window"`
	TextIDs      struct {
		Header     string `json:"header"`
		Row        string `json:"row"`
		Footer     string `json:"footer"`
		Actions    string `json:"actions"`
		GivePrompt string `json:"give_prompt"`
	} `json:"text_ids"`
	FrameRows        int                   `json:"frame_rows"`
	RowStep          int                   `json:"row_step"`
	Name             GeometryAnchor        `json:"name"`
	Cursor           GeometryAnchor        `json:"cursor"`
	Worn             GeometryAnchor        `json:"worn"`
	ActionCursor     GeometryAnchor        `json:"action_cursor"`
	CursorGlyph      int                   `json:"cursor_glyph"`
	WornGlyph        int                   `json:"worn_glyph"`
	MarkerMask       *int                  `json:"marker_mask"`
	GivePresentation *FieldItemPrompt      `json:"give_presentation"`
	UseNoEffect      *FieldItemUseNoEffect `json:"use_no_effect"`
	Drop             *FieldItemDrop        `json:"drop"`
	Evidence         Evidence              `json:"evidence"`
}

type FieldItemUseNoEffect struct {
	IntroTextID       string   `json:"intro_text_id"`
	ResultTextID      string   `json:"result_text_id"`
	ActorVariableCode *int     `json:"actor_variable_code"`
	ItemVariableCode  *int     `json:"item_variable_code"`
	Evidence          Evidence `json:"evidence"`
}

type FieldItemDrop struct {
	SuccessTextID     string   `json:"success_text_id"`
	BlockedTextID     string   `json:"blocked_text_id"`
	ActorVariableCode *int     `json:"actor_variable_code"`
	ItemVariableCode  *int     `json:"item_variable_code"`
	Evidence          Evidence `json:"evidence"`
}

func (s *FieldItemDrop) UnmarshalJSON(b []byte) error {
	type plain FieldItemDrop
	return requiredHome(b, (*plain)(s))
}

func (s *FieldItemUseNoEffect) UnmarshalJSON(b []byte) error {
	type plain FieldItemUseNoEffect
	return requiredHome(b, (*plain)(s))
}

// FieldItemPrompt presents a short native message over the previous UI frame.
type FieldItemPrompt struct {
	RawWindow         RawNewGameWindow `json:"raw_window"`
	Window            WindowLayout     `json:"window"`
	FrameTextID       string           `json:"frame_text_id"`
	GlyphStepX        int              `json:"glyph_step_x"`
	VariableCodeWords int              `json:"variable_code_words"`
	ForegroundRGB     []uint8          `json:"foreground_rgb"`
	BackdropRGB       []uint8          `json:"backdrop_rgb"`
	Shadow            WindowShadow     `json:"shadow"`
	Evidence          Evidence         `json:"evidence"`
}

func (s *FieldItemPrompt) UnmarshalJSON(b []byte) error {
	type plain FieldItemPrompt
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var f map[string]json.RawMessage
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	if err := decodeOpeningObject(f["raw_window"], &s.RawWindow, []string{"id", "flags", "x", "y", "width", "height", "address"}); err != nil {
		return err
	}
	return decodeOpeningObject(f["window"], &s.Window, []string{"id", "x", "y", "width", "height", "text_inset_x", "text_inset_y", "columns", "lines_per_page", "glyph_hold_frames", "glyph_timing_evidence", "evidence"})
}

func (s *FieldItems) UnmarshalJSON(b []byte) error {
	type plain FieldItems
	if err := requiredHome(b, (*plain)(s)); err != nil {
		return err
	}
	var f map[string]json.RawMessage
	if err := json.Unmarshal(b, &f); err != nil {
		return err
	}
	for name, destination := range map[string]*RawNewGameWindow{"raw_window": &s.RawWindow, "action_window": &s.ActionWindow} {
		if err := decodeOpeningObject(f[name], destination, []string{"id", "flags", "x", "y", "width", "height", "address"}); err != nil {
			return err
		}
	}
	return nil
}
func (p *Pack) validateFieldItems() error {
	s := p.Interface.FieldItems
	if s == nil || s.ID == "" || s.Evidence.Level != "D3" || s.FrameRows < 1 || s.RowStep != dq3data.GlyphPx || s.MarkerMask == nil || *s.MarkerMask <= 0 || *s.MarkerMask > 65535 || s.CursorGlyph < 0 || s.CursorGlyph >= dq3data.GlyphMax || s.WornGlyph < 0 || s.WornGlyph >= dq3data.GlyphMax {
		return fmt.Errorf("field items missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	if *s.MarkerMask&*p.Characters.ItemStorage.Encoding.CodeMask != 0 {
		return fmt.Errorf("item marker mask overlaps identity")
	}
	for _, w := range []RawNewGameWindow{s.RawWindow, s.ActionWindow} {
		if w.ID == "" || w.Address == "" || w.Flags < 0 || w.Flags > 7 || w.Flags&3 != 3 || w.X < 0 || w.Y < 0 || w.Width < 1 || w.Width%2 != 0 || w.Height < 1 || (w.X+w.Width)*8 > 640 || w.Y+w.Height > 350 {
			return fmt.Errorf("invalid field item window")
		}
	}
	maxHeight := (p.Events.ItemActions.PersonalInventorySlots + s.FrameRows) * s.RowStep
	if s.RawWindow.Y+maxHeight > 350 {
		return fmt.Errorf("field item capacity outside canvas")
	}
	for _, a := range []GeometryAnchor{s.Name, s.Cursor, s.Worn} {
		lastRowBottom := a.Y + (p.Events.ItemActions.PersonalInventorySlots-1)*s.RowStep + dq3data.GlyphPx
		if a.X < s.RawWindow.X*8 || a.X+dq3data.GlyphPx > (s.RawWindow.X+s.RawWindow.Width)*8 || a.Y < s.RawWindow.Y || lastRowBottom > s.RawWindow.Y+maxHeight {
			return fmt.Errorf("item anchor outside window")
		}
	}
	a := s.ActionCursor
	w := s.ActionWindow
	if a.X < w.X*8 || a.X+dq3data.GlyphPx > (w.X+w.Width)*8 || a.Y < w.Y || a.Y+len([3]int{})*s.RowStep > w.Y+w.Height {
		return fmt.Errorf("action anchors outside window")
	}
	for _, id := range []string{s.TextIDs.Header, s.TextIDs.Row, s.TextIDs.Footer, s.TextIDs.Actions, s.TextIDs.GivePrompt} {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("field item text reference missing")
		}
	}
	for _, id := range []string{s.TextIDs.Header, s.TextIDs.Row, s.TextIDs.Footer} {
		d, _ := p.TextDefinition(id)
		if d.Layout.Columns*dq3data.GlyphPx != s.RawWindow.Width*8 || d.Layout.LinesPerPage != 1 || len(d.GlyphCodes) != d.Layout.Columns {
			return fmt.Errorf("item frame record shape invalid")
		}
	}
	d, _ := p.TextDefinition(s.TextIDs.Actions)
	if d.Layout.Columns*dq3data.GlyphPx != s.ActionWindow.Width*8 || d.Layout.LinesPerPage*dq3data.GlyphPx != s.ActionWindow.Height || len(d.GlyphCodes) != d.Layout.Columns*d.Layout.LinesPerPage+d.Layout.LinesPerPage-1 {
		return fmt.Errorf("item action record shape invalid")
	}
	if err := p.validateFieldItemPrompt(); err != nil {
		return err
	}
	if err := p.validateFieldItemUseNoEffect(); err != nil {
		return err
	}
	return p.validateFieldItemDrop()
}

func (p *Pack) validateFieldItemDrop() error {
	s := p.Interface.FieldItems.Drop
	if s == nil || s.Evidence.Level != "D3" || s.SuccessTextID == s.BlockedTextID || s.ActorVariableCode == nil || s.ItemVariableCode == nil || *s.ActorVariableCode == *s.ItemVariableCode {
		return fmt.Errorf("item drop presentation missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	for _, code := range []*int{s.ActorVariableCode, s.ItemVariableCode} {
		if *code < dq3data.GlyphMax || *code > 65535 || !dq3data.IsVarInsert(uint16(*code)) {
			return fmt.Errorf("invalid drop variable binding")
		}
	}
	for _, id := range []string{s.SuccessTextID, s.BlockedTextID} {
		d, ok := p.TextDefinition(id)
		if !ok || d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("drop text missing or unreviewed")
		}
		counts := map[int]int{}
		line, rows := 0, 1
		w := p.Interface.FieldItems.GivePresentation
		for _, code := range d.GlyphCodes {
			switch {
			case code == *s.ActorVariableCode || code == *s.ItemVariableCode:
				if id != s.SuccessTextID {
					return fmt.Errorf("blocked drop binding unsupported")
				}
				counts[code]++
				line++
			case code == int(dq3data.TxtNL):
				if id != s.SuccessTextID || line == 0 {
					return fmt.Errorf("drop newline unsupported")
				}
				line = 0
				rows++
			case code >= 0 && code < dq3data.GlyphMax:
				line++
			default:
				return fmt.Errorf("drop control unsupported")
			}
			if rows > w.Window.LinesPerPage || w.Window.TextInsetY+rows*dq3data.GlyphPx > w.Window.Height || (line > 0 && w.Window.TextInsetX+(line-1)*w.GlyphStepX+dq3data.GlyphPx > w.Window.Width) {
				return fmt.Errorf("drop text outside prompt")
			}
		}
		if id == s.SuccessTextID && (counts[*s.ActorVariableCode] != 1 || counts[*s.ItemVariableCode] != 1) {
			return fmt.Errorf("drop binding missing or repeated")
		}
	}
	return nil
}

func (p *Pack) validateFieldItemUseNoEffect() error {
	s := p.Interface.FieldItems.UseNoEffect
	if s == nil || s.Evidence.Level != "D3" || s.ActorVariableCode == nil || s.ItemVariableCode == nil || *s.ActorVariableCode == *s.ItemVariableCode {
		return fmt.Errorf("item no-effect presentation missing or unreviewed")
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return err
	}
	for _, code := range []*int{s.ActorVariableCode, s.ItemVariableCode} {
		if *code < dq3data.GlyphMax || *code > 65535 || !dq3data.IsVarInsert(uint16(*code)) {
			return fmt.Errorf("invalid item prompt variable binding")
		}
	}
	intro, a := p.TextDefinition(s.IntroTextID)
	result, b := p.TextDefinition(s.ResultTextID)
	if !a || !b {
		return fmt.Errorf("item use text reference missing")
	}
	for _, d := range []*TextDefinition{intro, result} {
		if d.Source.Kind != "legacy_record" || d.Source.Record == nil || d.Evidence.Level != "D3" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("item use text unreviewed")
		}
	}
	counts := map[int]int{}
	for _, code := range intro.GlyphCodes {
		if code == *s.ActorVariableCode || code == *s.ItemVariableCode {
			counts[code]++
			continue
		}
		if code < 0 || code >= dq3data.GlyphMax {
			return fmt.Errorf("item introduction control unsupported")
		}
	}
	if counts[*s.ActorVariableCode] != 1 || counts[*s.ItemVariableCode] != 1 {
		return fmt.Errorf("item introduction binding missing or repeated")
	}
	w := p.Interface.FieldItems.GivePresentation
	if w.Window.LinesPerPage < 2 || w.Window.TextInsetY+2*dq3data.GlyphPx > w.Window.Height || w.Window.TextInsetX+(len(result.GlyphCodes)-1)*w.GlyphStepX+dq3data.GlyphPx > w.Window.Width {
		return fmt.Errorf("item result outside prompt")
	}
	for _, code := range result.GlyphCodes {
		if code < 0 || code >= dq3data.GlyphMax {
			return fmt.Errorf("item result control unsupported")
		}
	}
	return nil
}

func (p *Pack) validateFieldItemPrompt() error {
	s := p.Interface.FieldItems.GivePresentation
	if s == nil || s.Evidence.Level != "D3" {
		return fmt.Errorf("item give presentation missing or unreviewed")
	}
	for _, e := range []Evidence{s.Evidence, s.Window.Evidence, s.Shadow.Evidence} {
		if e.Level != "D3" {
			return fmt.Errorf("item prompt evidence unreviewed")
		}
		if err := validateEvidence(e); err != nil {
			return err
		}
	}
	w, r := s.Window, s.RawWindow
	if r.ID == "" || r.Address == "" || r.Flags != 1 || w.ID == "" || w.X != r.X*8 || w.Y != r.Y || w.Width != r.Width*8 || w.Height != r.Height || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Width%16 != 0 || w.Height <= 0 || w.Height%16 != 0 || w.TextInsetX < 0 || w.TextInsetY < 0 || w.Columns < 1 || w.LinesPerPage < 1 || w.TextInsetX+w.Columns*dq3data.GlyphPx > w.Width || w.TextInsetY+w.LinesPerPage*dq3data.GlyphPx > w.Height || w.GlyphHoldFrames < 1 || w.GlyphTiming == nil || s.GlyphStepX < dq3data.GlyphPx || s.VariableCodeWords != 1 || len(s.ForegroundRGB) != 3 || len(s.BackdropRGB) != 3 {
		return fmt.Errorf("invalid item prompt geometry or timing")
	}
	if err := validateEvidence(*w.GlyphTiming); err != nil {
		return err
	}
	if w.GlyphTiming.Level != "D2" && w.GlyphTiming.Level != "D3" {
		return fmt.Errorf("item prompt timing unreviewed")
	}
	sh := s.Shadow
	if sh.Mode != "vga_word_latch_and" || sh.OffsetX < 0 || sh.OffsetX%8 != 0 || sh.OffsetY < 0 || w.X+w.Width+sh.OffsetX > 640 || w.Y+w.Height+sh.OffsetY > 350 {
		return fmt.Errorf("item prompt shadow outside canvas")
	}
	frame, ok := p.TextDefinition(s.FrameTextID)
	if !ok || frame.Source.Kind != "legacy_record" || frame.Source.Record == nil || frame.Evidence.Level != "D3" || frame.Layout.Columns*dq3data.GlyphPx != w.Width || frame.Layout.LinesPerPage*dq3data.GlyphPx != w.Height {
		return fmt.Errorf("item prompt frame invalid")
	}
	col, row := 0, 0
	for _, c := range frame.GlyphCodes {
		if c == int(dq3data.TxtNL) {
			if col != frame.Layout.Columns {
				return fmt.Errorf("item prompt frame row invalid")
			}
			col, row = 0, row+1
			continue
		}
		if c < 0 || c >= dq3data.GlyphMax {
			return fmt.Errorf("item prompt frame control unsupported")
		}
		col++
	}
	if col != frame.Layout.Columns || row+1 != frame.Layout.LinesPerPage {
		return fmt.Errorf("item prompt frame shape invalid")
	}
	text, _ := p.TextDefinition(p.Interface.FieldItems.TextIDs.GivePrompt)
	if len(text.GlyphCodes) > w.Columns || w.TextInsetX+(len(text.GlyphCodes)-1)*s.GlyphStepX+dq3data.GlyphPx > w.Width || w.TextInsetY+dq3data.GlyphPx > w.Height {
		return fmt.Errorf("item prompt text outside canvas")
	}
	for _, c := range text.GlyphCodes {
		if c < 0 || c >= dq3data.GlyphMax {
			return fmt.Errorf("item prompt control unsupported")
		}
	}
	return nil
}
func (p *Pack) ItemDropAllowed(entry itemstore.Entry) bool {
	if p == nil || p.Characters.ItemStorage == nil {
		return false
	}
	s := p.Characters.ItemStorage
	if s.DropBlockedMask == nil || entry.Code < 0 || entry.Code >= len(s.Items) || s.Items[entry.Code].DropForbidden == nil || s.Items[entry.Code].DropHasValue == nil {
		return false
	}
	return entry.Word&uint16(*s.DropBlockedMask) == 0 && !*s.Items[entry.Code].DropForbidden && *s.Items[entry.Code].DropHasValue
}
