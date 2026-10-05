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
	FrameRows    int            `json:"frame_rows"`
	RowStep      int            `json:"row_step"`
	Name         GeometryAnchor `json:"name"`
	Cursor       GeometryAnchor `json:"cursor"`
	Worn         GeometryAnchor `json:"worn"`
	ActionCursor GeometryAnchor `json:"action_cursor"`
	CursorGlyph  int            `json:"cursor_glyph"`
	WornGlyph    int            `json:"worn_glyph"`
	MarkerMask   *int           `json:"marker_mask"`
	Evidence     Evidence       `json:"evidence"`
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
	return nil
}
func (p *Pack) ItemDropAllowed(entry itemstore.Entry) bool {
	if p == nil || p.Characters.ItemStorage == nil {
		return false
	}
	s := p.Characters.ItemStorage
	if s.DropBlockedMask == nil || entry.Code < 0 || entry.Code >= len(s.Items) || s.Items[entry.Code].DropForbidden == nil {
		return false
	}
	return entry.Word&uint16(*s.DropBlockedMask) == 0 && !*s.Items[entry.Code].DropForbidden
}
