package gamepack

import (
	"fmt"
	"reflect"
)

type HomePoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}
type HomeRect struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}
type HomeNumber struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Digits int `json:"digits"`
}

func requiredHome(raw []byte, target any) error {
	t := reflect.TypeOf(target).Elem()
	keys := make([]string, t.NumField())
	for i := range keys {
		keys[i] = t.Field(i).Tag.Get("json")
	}
	return decodeOpeningObject(raw, target, keys)
}
func (p *HomePoint) UnmarshalJSON(b []byte) error {
	type plain HomePoint
	return requiredHome(b, (*plain)(p))
}
func (p *HomeRect) UnmarshalJSON(b []byte) error {
	type plain HomeRect
	return requiredHome(b, (*plain)(p))
}
func (p *HomeNumber) UnmarshalJSON(b []byte) error {
	type plain HomeNumber
	return requiredHome(b, (*plain)(p))
}

type OpeningHome struct {
	LeaderRecord       *int                `json:"leader_record"`
	ApproachTiles      []HomePoint         `json:"approach_tiles"`
	ApproachFlag       int                 `json:"approach_flag"`
	ApproachSubid      int                 `json:"approach_subid"`
	ApproachHandlerRaw int                 `json:"approach_handler_raw"`
	ApproachFrame      OpeningArrivalFrame `json:"approach_frame"`
	Picture            HomePicture         `json:"picture"`
	Evidence           Evidence            `json:"evidence"`
}

func (p *OpeningHome) UnmarshalJSON(b []byte) error {
	type plain OpeningHome
	return requiredHome(b, (*plain)(p))
}

type HomePicture struct {
	AssetKey         string        `json:"asset_key"`
	PaletteAssetKey  string        `json:"palette_asset_key"`
	PaletteBank      int           `json:"palette_bank"`
	PaletteOverrides []RasterColor `json:"palette_overrides"`
	BackgroundTile   int           `json:"background_tile"`
	RandomAdd        uint16        `json:"random_add"`
	RandomRotate     int           `json:"random_rotate"`
	RandomMask       int           `json:"random_mask"`
	TripletSize      int           `json:"triplet_size"`
	Questions        [][]int       `json:"questions"`
	OptionCount      int           `json:"option_count"`
	AcceptPolicy     string        `json:"accept_policy"`
	Window           HomeRect      `json:"window"`
	TopTextID        string        `json:"top_text_id"`
	BodyTextID       string        `json:"body_text_id"`
	BottomTextID     string        `json:"bottom_text_id"`
	BodyRows         int           `json:"body_rows"`
	FontIndex        int           `json:"font_index"`
	BlankGlyph       int           `json:"blank_glyph"`
	ShadowOffset     HomePoint     `json:"shadow_offset"`
	Boxes            []HomeRect    `json:"boxes"`
	OptionOrigin     HomePoint     `json:"option_origin"`
	SelectedOrigin   HomePoint     `json:"selected_origin"`
	StepX            int           `json:"step_x"`
	Cursor           HomeRect      `json:"cursor"`
	CursorXOR        int           `json:"cursor_xor"`
	QuestionField    HomeNumber    `json:"question_field"`
	PreviewValues    []int         `json:"preview_values"`
	PreviewField     HomeNumber    `json:"preview_field"`
	PreviewStepY     int           `json:"preview_step_y"`
	Evidence         Evidence      `json:"evidence"`
}

func (p *HomePicture) UnmarshalJSON(b []byte) error {
	type plain HomePicture
	return requiredHome(b, (*plain)(p))
}
func validHomeRect(r HomeRect) bool {
	return r.X >= 0 && r.Y >= 0 && r.Width > 0 && r.Height > 0 && r.X+r.Width <= 640 && r.Y+r.Height <= 350
}
func validHomeNumber(n HomeNumber) bool {
	return n.Digits > 0 && n.Digits <= 5 && validHomeRect(HomeRect{n.X, n.Y, n.Digits * 16, 16})
}
func (p *Pack) validateOpeningHome(e *OpeningEscort) error {
	h := e.Home
	if h == nil || h.LeaderRecord == nil || *h.LeaderRecord < 0 || len(h.ApproachTiles) == 0 || h.ApproachFlag < 0 || h.ApproachFlag > 65535 || h.ApproachSubid <= 0 || h.ApproachSubid > 31 || h.ApproachHandlerRaw < 0 {
		return fmt.Errorf("opening home trigger invalid")
	}
	if h.Evidence.Level != "D3" || h.Picture.Evidence.Level != "D3" {
		return fmt.Errorf("opening home requires D3 evidence")
	}
	for _, v := range []Evidence{h.Evidence, h.Picture.Evidence} {
		if err := validateEvidence(v); err != nil {
			return err
		}
	}
	for i, f := range e.Frames {
		if f.LeaderFacing < 0 || f.LeaderFacing > 3 {
			return fmt.Errorf("opening home facing invalid")
		}
		if i > 0 {
			prev := e.Frames[i-1]
			if f.Player != prev.Player || absTileDifference(f.Leader.X, prev.Leader.X)+absTileDifference(f.Leader.Y, prev.Leader.Y) > 1 {
				return fmt.Errorf("opening home sequence not adjacent or player moved")
			}
		}
	}
	f := h.ApproachFrame
	last := e.Frames[len(e.Frames)-1]
	if f.Player.X < 0 || f.Player.Y < 0 || f.LeaderFacing < 0 || f.LeaderFacing > 3 || f.HoldFrames <= 0 || absTileDifference(last.Leader.X, f.Leader.X)+absTileDifference(last.Leader.Y, f.Leader.Y) != 1 {
		return fmt.Errorf("opening home approach frame invalid")
	}
	for _, t := range h.ApproachTiles {
		if t.X < 0 || t.Y < 0 || t.X != f.Player.X || t.Y != f.Player.Y {
			return fmt.Errorf("opening home approach tile invalid")
		}
	}
	s := h.Picture
	if _, ok := p.Asset(s.AssetKey); !ok {
		return fmt.Errorf("opening home picture asset missing")
	}
	if _, ok := p.Asset(s.PaletteAssetKey); !ok {
		return fmt.Errorf("opening home palette missing")
	}
	if s.PaletteBank < 0 || s.RandomRotate < 1 || s.RandomRotate > 15 || s.RandomMask < 1 || s.RandomMask > 65535 || s.TripletSize < 1 || s.OptionCount != 2*s.TripletSize || len(s.Questions) < 1 || len(s.Questions) > s.RandomMask+1 || s.AcceptPolicy != "unchecked" || s.FontIndex < 0 || s.FontIndex > 15 || s.CursorXOR < 0 || s.CursorXOR > 15 {
		return fmt.Errorf("opening home picture rules invalid")
	}
	rounds := len(s.Questions[0])
	if rounds < 1 || rounds > 16 {
		return fmt.Errorf("opening home picture rounds invalid")
	}
	for _, q := range s.Questions {
		if len(q) != rounds {
			return fmt.Errorf("opening home question shape invalid")
		}
		for _, v := range q {
			if v < 0 {
				return fmt.Errorf("opening home question reference invalid")
			}
		}
	}
	if !validHomeRect(s.Window) || s.Window.Width%16 != 0 || s.BodyRows < 1 || s.Window.Height != (s.BodyRows+2)*16 || s.ShadowOffset.X < 0 || s.ShadowOffset.Y < 0 || s.Window.X+s.Window.Width+s.ShadowOffset.X > 640 || s.Window.Y+s.Window.Height+s.ShadowOffset.Y > 350 || s.StepX <= 0 || s.StepX%8 != 0 || !validHomeRect(s.Cursor) || !validHomeNumber(s.QuestionField) || !validHomeNumber(s.PreviewField) || s.PreviewStepY < 16 || s.PreviewField.Y+(rounds+len(s.PreviewValues)-1)*s.PreviewStepY+16 > 350 {
		return fmt.Errorf("opening home picture geometry invalid")
	}
	for _, r := range s.Boxes {
		if !validHomeRect(r) {
			return fmt.Errorf("opening home box invalid")
		}
	}
	for _, a := range []struct {
		point HomePoint
		count int
	}{{s.OptionOrigin, s.OptionCount}, {s.SelectedOrigin, rounds}} {
		if !validHomeRect(HomeRect{a.point.X, a.point.Y, (a.count-1)*s.StepX + 32, 24}) {
			return fmt.Errorf("opening home image geometry invalid")
		}
	}
	if s.Cursor.X+(s.OptionCount-1)*s.StepX+s.Cursor.Width > 640 {
		return fmt.Errorf("opening home cursor bounds invalid")
	}
	for _, v := range s.PreviewValues {
		if v < 0 || v > 65535 {
			return fmt.Errorf("opening home preview invalid")
		}
	}
	seen := map[int]bool{}
	for _, o := range s.PaletteOverrides {
		if o.Index < 0 || o.Index > 15 || len(o.RGB) != 3 || seen[o.Index] {
			return fmt.Errorf("opening home palette override invalid")
		}
		seen[o.Index] = true
	}
	if s.BackgroundTile < 0 || s.BlankGlyph < 0 {
		return fmt.Errorf("opening home background tile or glyph invalid")
	}
	return nil
}
func (p *Pack) validateOpeningHomeRefs() error {
	if p.Interface.OpeningEscort == nil {
		return nil
	}
	s := p.Interface.OpeningEscort.Home.Picture
	for _, id := range []string{s.TopTextID, s.BodyTextID, s.BottomTextID} {
		codes, ok := p.TextGlyphCodes(id)
		if !ok || len(codes) != s.Window.Width/16 {
			return fmt.Errorf("opening home frame text reference invalid: %s", id)
		}
	}
	return nil
}
