// Package gamepack loads versioned, data-driven game definitions shared by the
// DQ1/DQ2/DQ3 Ebitengine core. It intentionally rejects unknown JSON fields:
// a misspelled original-game parameter must never silently become a default.
package gamepack

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

const (
	SchemaVersion = "0.1.32"
	EngineAPI     = ">=0.1.0 <0.2.0"
	ReviveService = "common:service.revive"
)

//go:embed packs/dq3_cht
var builtin embed.FS

type Manifest struct {
	SchemaVersion  string              `json:"schema_version"`
	PackID         string              `json:"pack_id"`
	Game           string              `json:"game"`
	Edition        string              `json:"edition"`
	ContentVersion string              `json:"content_version"`
	EngineAPI      string              `json:"engine_api"`
	TitleTextID    string              `json:"title_text_id"`
	EntryEventID   string              `json:"entry_event_id"`
	SaveNamespace  string              `json:"save_namespace"`
	Capabilities   []string            `json:"capabilities"`
	Data           map[string]string   `json:"data"`
	Assets         map[string]AssetRef `json:"assets"`
}

type AssetRef struct {
	Path            string `json:"path"`
	Size            int64  `json:"size"`
	SHA256          string `json:"sha256"`
	AlternateSize   int64  `json:"alternate_size,omitempty"`
	AlternateSHA256 string `json:"alternate_sha256,omitempty"`
}

// AudioDefinition 是 versioned game pack 的音樂／音效 cue 表。引擎只查穩定
// cue ID，不把某一代遊戲的原始軌號寫進共用 renderer。
type AudioDefinition struct {
	SchemaVersion string              `json:"schema_version"`
	Cues          map[string]AudioCue `json:"cues"`
}

type AudioCue struct {
	Kind     string   `json:"kind"`
	Track    int      `json:"track"`
	Loop     bool     `json:"loop"`
	Priority int      `json:"priority"`
	Evidence Evidence `json:"evidence"`
}

type Evidence struct {
	Level        string `json:"level"`
	SourceKind   string `json:"source_kind"`
	Source       string `json:"source"`
	AddressSpace string `json:"address_space"`
	Address      string `json:"address"`
	Consumer     string `json:"consumer"`
	Doc          string `json:"doc"`
	Note         string `json:"note,omitempty"`
}

// BattleSourceFile records the exact legacy input used to derive the E2 battle
// contract.  It is metadata only; the engine still loads pack-owned assets and
// never guesses missing raw data.
type BattleSourceFile struct {
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
}

// BattleRawBlob preserves a non-destructive byte range together with its
// original address and evidence.  Unknown fields remain bytes rather than
// becoming speculative production names.
type BattleRawBlob struct {
	AddressSpace string   `json:"address_space"`
	Address      string   `json:"address"`
	FileOffset   string   `json:"file_offset,omitempty"`
	LengthBytes  int      `json:"length_bytes"`
	RawSHA256    string   `json:"raw_sha256"`
	RawHex       string   `json:"raw_hex"`
	Evidence     Evidence `json:"evidence"`
}

type BattleFixedFormationRecord struct {
	DGroup   string   `json:"dgroup"`
	File     string   `json:"file"`
	Count    int      `json:"count"`
	RawHex   string   `json:"raw_hex"`
	Caller   string   `json:"caller"`
	Evidence Evidence `json:"evidence"`
}

// BattleFormationPosition preserves the confirmed raw-to-EGA projection used
// by the original enemy loader/drawer.  These are data-pack values rather
// than renderer constants: the shared engine only applies the primitive.
type BattleFormationPosition struct {
	OriginRaw            int      `json:"origin_raw"`
	FirstMemberOffsetRaw int      `json:"first_member_offset_raw"`
	WeightStepMultiplier int      `json:"weight_step_multiplier"`
	EGAStrideBytes       int      `json:"ega_stride_bytes"`
	EGABottomRaw         int      `json:"ega_bottom_raw"`
	PixelsPerRawByte     int      `json:"pixels_per_raw_byte"`
	Evidence             Evidence `json:"evidence"`
}

// BattleBackgroundSelector is the non-executable selector copied from a
// legacy formation record.  Its values are interpreted only through the
// enclosing pack's checked archive and palette contract.
type BattleBackgroundSelector struct {
	PageRaw        int `json:"page_raw"`
	PaletteBankRaw int `json:"palette_bank_raw"`
}

// BattleBackgroundData describes the original planar background archive as a
// versioned asset contract.  The engine decodes this generic layout but does
// not carry a DQ3 page, palette bank, filename, or byte stride fallback.
type BattleBackgroundData struct {
	ArchiveAsset          string                   `json:"archive_asset"`
	PaletteAsset          string                   `json:"palette_asset"`
	PageStrideBytes       int                      `json:"page_stride_bytes"`
	FieldOffsetBytes      int                      `json:"field_offset_bytes"`
	FieldChunkBytes       int                      `json:"field_chunk_bytes"`
	WidthPixels           int                      `json:"width_pixels"`
	FieldRows             int                      `json:"field_rows"`
	PlaneCount            int                      `json:"plane_count"`
	PageCount             int                      `json:"page_count"`
	PaletteBankCount      int                      `json:"palette_bank_count"`
	PaletteEntriesPerBank int                      `json:"palette_entries_per_bank"`
	DefaultSelector       BattleBackgroundSelector `json:"default_selector"`
	Evidence              Evidence                 `json:"evidence"`
}

type BattleEncounterData struct {
	RegionLookup      BattleRawBlob                `json:"region_lookup"`
	CandidateTable    BattleRawBlob                `json:"candidate_table"`
	RowStrideBytes    int                          `json:"row_stride_bytes"`
	RegionStrideBytes int                          `json:"region_stride_bytes"`
	UsedRegionIDs     []int                        `json:"used_region_ids"`
	FixedRecords      []BattleFixedFormationRecord `json:"fixed_records"`
	FormationPosition BattleFormationPosition      `json:"formation_position"`
	Evidence          Evidence                     `json:"evidence"`
}

type BattleResistanceData struct {
	EffectThresholds BattleRawBlob     `json:"effect_thresholds"`
	ClassThresholds  []int             `json:"class_thresholds"`
	DescriptorTable  BattleRawBlob     `json:"descriptor_table"`
	EffectNamesZH    map[string]string `json:"effect_names_zh"`
	Evidence         Evidence          `json:"evidence"`
}

// BattlePackData is the E2 production contract for the statically confirmed
// battle inputs.  It deliberately excludes unproven action-frame and player
// timing semantics; those remain in the RE sidecar as unknown/V3.
type BattlePackData struct {
	SchemaVersion string               `json:"schema_version"`
	SourceFiles   []BattleSourceFile   `json:"source_files"`
	Background    BattleBackgroundData `json:"background"`
	Encounter     BattleEncounterData  `json:"encounter"`
	Resistance    BattleResistanceData `json:"resistance"`
	Evidence      Evidence             `json:"evidence"`
}

type Pricing struct {
	FormulaID string `json:"formula_id"`
	LevelCap  int    `json:"level_cap"`
	CostsGold []int  `json:"costs_gold"`
}

type ServiceDefinition struct {
	ID       string   `json:"id"`
	Pricing  Pricing  `json:"pricing"`
	Evidence Evidence `json:"evidence"`
}

type Facilities struct {
	SchemaVersion      string              `json:"schema_version"`
	ServiceDefinitions []ServiceDefinition `json:"service_definitions"`
}

// FrameStyle 是版本專屬的視窗前景／內部色彩與邊線遮罩契約。原版 EGA 的
// 邊線 primitive 仍由引擎提供；pack 只選擇已註冊的 pattern 與玩家可見
// RGB 結果，避免 renderer 偷藏某一版的黑底／白框預設或任意 JSON 程式碼。
type FrameStyle struct {
	ID              string    `json:"id"`
	BorderPattern   string    `json:"border_pattern"`
	BorderRGB       [3]uint8  `json:"border_rgb"`
	BorderAccentRGB *[3]uint8 `json:"border_accent_rgb,omitempty"`
	InteriorRGB     [3]uint8  `json:"interior_rgb"`
	Evidence        Evidence  `json:"evidence"`
}

// WindowLayout 是版本專屬的固定視窗幾何。引擎只依此通用契約繪製，
// 不保存任何 DQ3 專屬座標、尺寸或框線色彩。
type WindowLayout struct {
	ID              string      `json:"id"`
	X               int         `json:"x"`
	Y               int         `json:"y"`
	Width           int         `json:"width"`
	Height          int         `json:"height"`
	TextInsetX      int         `json:"text_inset_x"`
	TextInsetY      int         `json:"text_inset_y"`
	Columns         int         `json:"columns"`
	LinesPerPage    int         `json:"lines_per_page"`
	HPLabelGlyph    *int        `json:"hp_label_glyph,omitempty"`
	MPLabelGlyph    *int        `json:"mp_label_glyph,omitempty"`
	LevelLabelGlyph *int        `json:"level_label_glyph,omitempty"`
	Frame           *FrameStyle `json:"frame,omitempty"`
	Evidence        Evidence    `json:"evidence"`
}

// PartyHUDLayout extends the generic window with pack-owned column contents.
// The engine does not assume where names, labels, values or class glyphs sit.
type PartyHUDLayout struct {
	WindowLayout
	NameInsetX    int     `json:"name_inset_x"`
	ValueInsetX   int     `json:"value_inset_x"`
	NameMaxGlyphs int     `json:"name_max_glyphs"`
	ClassGlyphs   [][]int `json:"class_glyphs"`
}

// BattlePanelLayout extends the common window geometry with the small set of
// data-owned offsets and row sizing rules used by the original battle HUD.
// HeightMode is deliberately a named primitive rather than an expression: the
// engine may apply a registered rule, but a pack still owns its row/base values.
type BattlePanelLayout struct {
	WindowLayout
	HeightMode           string `json:"height_mode"`
	BaseRows             int    `json:"base_rows"`
	RowHeight            int    `json:"row_height"`
	CursorInsetX         int    `json:"cursor_inset_x"`
	RowInsetY            int    `json:"row_inset_y"`
	LabelInsetX          int    `json:"label_inset_x"`
	SecondaryLabelInsetX int    `json:"secondary_label_inset_x"`
	ValueInsetX          int    `json:"value_inset_x"`
	NameInsetX           int    `json:"name_inset_x"`
	CountInsetX          int    `json:"count_inset_x"`
	// CountInsetY is relative to the name row's TextInsetY.  Keeping the
	// relation in the pack makes the enemy count baseline version-owned
	// without duplicating a DQ3-specific Y coordinate in the renderer.
	CountInsetY *int `json:"count_inset_y,omitempty"`
}

// BattleSceneLayout 是版本專屬戰鬥場景帶與游標字模。引擎只套用這些
// 已由原版證據閉合的幾何與字模，不把某一版的 y 座標或 raw glyph 留在
// 共用 renderer。
type BattleSceneLayout struct {
	ID          string   `json:"id"`
	FieldY0     int      `json:"field_y0"`
	FieldY1     int      `json:"field_y1"`
	GroundY     int      `json:"ground_y"`
	CursorGlyph int      `json:"cursor_glyph"`
	Evidence    Evidence `json:"evidence"`
}

// NewGameLabels 是開場主選單、姓名 rec451–456、性別與能力確認畫面使用的版本專屬字模。
// 欄位名稱固定，不承載可執行規則；共同 Evidence 套用到每一個原始 glyph
// 陣列，讓來源、位址基準與推論等級不會隨 renderer 遺失。
type NewGameLabels struct {
	Title        []int `json:"title"`
	Start        []int `json:"start"`
	Load         []int `json:"load"`
	Male         []int `json:"male"`
	Female       []int `json:"female"`
	Level        []int `json:"level"`
	HP           []int `json:"hp"`
	MP           []int `json:"mp"`
	Strength     []int `json:"strength"`
	Agility      []int `json:"agility"`
	Vitality     []int `json:"vitality"`
	Intelligence []int `json:"intelligence"`
	Luck         []int `json:"luck"`
	MaxHP        []int `json:"max_hp"`
	MaxMP        []int `json:"max_mp"`
	Attack       []int `json:"attack"`
	Defense      []int `json:"defense"`
	Experience   []int `json:"experience"`
	Sex          []int `json:"sex"`
	Hero         []int `json:"hero"`
	Cloth        []int `json:"cloth"`
	Prompt       []int `json:"prompt"`
	Yes          []int `json:"yes"`
	No           []int `json:"no"`
	// ChoiceCursor 是創角確認視窗的游標 glyph；不能借用戰鬥場景的目標游標。
	ChoiceCursor    []int    `json:"choice_cursor"`
	Backspace       []int    `json:"backspace"`
	OK              []int    `json:"ok"`
	NameTitle       []int    `json:"name_title"`
	NameLeftArrow   []int    `json:"name_left_arrow"`
	NameRightArrow  []int    `json:"name_right_arrow"`
	NameZhuyinGrid  []int    `json:"name_zhuyin_grid"`
	NameAlnumGrid   []int    `json:"name_alnum_grid"`
	FunctionZhuyin  [][]int  `json:"function_zhuyin"`
	FunctionAlnum   [][]int  `json:"function_alnum"`
	InputModeZhuyin []int    `json:"input_mode_zhuyin"`
	Evidence        Evidence `json:"evidence"`
}

// FrameEdgeWidths 是版本專屬 framed rectangle 的個別邊寬覆寫。0 表示使用
// 已註冊 frame primitive 的預設寬度；正值只調整該邊，讓相鄰 panel 能以同一
// 個共用 renderer 表達原版的共享 seam，而不在 Go 寫入版本座標。
type FrameEdgeWidths struct {
	Left   int `json:"left,omitempty"`
	Right  int `json:"right,omitempty"`
	Top    int `json:"top,omitempty"`
	Bottom int `json:"bottom,omitempty"`
}

// GeometryRect 是版本專屬的像素矩形。座標以 640×350 邏輯畫布為基準；
// 原始 DOS window 結構另保存在 NewGameGeometry.RawWindows，避免把像素
// 對拍值誤當成 EXE 的原始欄位。
type GeometryRect struct {
	X               int             `json:"x"`
	Y               int             `json:"y"`
	Width           int             `json:"width"`
	Height          int             `json:"height"`
	FrameEdgeWidths FrameEdgeWidths `json:"frame_edge_widths,omitempty"`
}

// GeometryAnchor 是固定文字起點與可選列／欄步距。
type GeometryAnchor struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	StepX int `json:"step_x,omitempty"`
	StepY int `json:"step_y,omitempty"`
}

// NumberField 描述固定寬度十進位欄。X/Y 是最高位數的起點；Digits 不足時由
// renderer 右對齊，對應原版 BP 每 digit 前進兩個 EGA bytes 的行為。
type NumberField struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Digits int `json:"digits"`
}

// NewGameStatField 把固定 label glyph stream 與動態數字欄分開保存。原版
// record407 的 leading blank／冒號不可從字串長度推導，故 label 是完整的
// 原始 glyph stream，數字仍使用固定 5/8 位 field。
type NewGameStatField struct {
	Label GeometryAnchor `json:"label"`
	Value NumberField    `json:"value"`
}

// NewGameStatLayout 是創角確認的角色 record 欄位布局。欄位名是跨版本 RPG
// 屬性角色；字模、像素位置、decimal width 均由版本 pack 提供。
type NewGameStatLayout struct {
	Level        NewGameStatField `json:"level"`
	HP           NewGameStatField `json:"hp"`
	MP           NewGameStatField `json:"mp"`
	Strength     NewGameStatField `json:"strength"`
	Agility      NewGameStatField `json:"agility"`
	Vitality     NewGameStatField `json:"vitality"`
	Intelligence NewGameStatField `json:"intelligence"`
	Luck         NewGameStatField `json:"luck"`
	MaxHP        NewGameStatField `json:"max_hp"`
	MaxMP        NewGameStatField `json:"max_mp"`
	Attack       NewGameStatField `json:"attack"`
	Defense      NewGameStatField `json:"defense"`
	Experience   NewGameStatField `json:"experience"`
}

// GeometryGrid 是姓名盤等規則格盤的像素起點與步距。
type GeometryGrid struct {
	X       int `json:"x"`
	Y       int `json:"y"`
	Columns int `json:"columns"`
	Rows    int `json:"rows"`
	StepX   int `json:"step_x"`
	StepY   int `json:"step_y"`
}

// PatternFill 是版本專屬的矩形底圖覆蓋。Pattern 只能引用引擎已註冊的
// primitive；RGB、矩形與證據仍由 game pack 擁有。這讓 EGA 選項背景能
// 保留「命中的 plane 寫入、未命中的 plane 留下底圖」的可見結果，而不把
// 任意排版程式碼塞進 JSON。
type PatternFill struct {
	ID       string       `json:"id"`
	Pattern  string       `json:"pattern"`
	Rect     GeometryRect `json:"rect"`
	RGB      [3]uint8     `json:"rgb"`
	Evidence Evidence     `json:"evidence"`
}

// RawNewGameWindow 保留 IDA linear address 空間中的原始 window 欄位；
// Pixel 是同一結構經 EGA writer／實機截圖閉合後的 renderer 矩形。
type RawNewGameWindow struct {
	ID      string `json:"id"`
	Flags   int    `json:"flags"`
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Width   int    `json:"width"`
	Height  int    `json:"height"`
	Address string `json:"address"`
}

// NewGameWindowBackdrop 將一個原始 EGA window 結構連到引擎已註冊的
// 背景 primitive。原始 x／width 保持 EGA byte 單位，pixel 投影由共用引擎
// 執行；pack 只選擇原始 window 與有證據的 primitive，不能嵌入任意繪圖程式。
type NewGameWindowBackdrop struct {
	ID                   string   `json:"id"`
	RawWindowID          string   `json:"raw_window_id"`
	Primitive            string   `json:"primitive"`
	DrawOrder            *int     `json:"draw_order"`
	TerminalRightPixels  *int     `json:"terminal_right_px"`
	TerminalBottomPixels *int     `json:"terminal_bottom_px"`
	Evidence             Evidence `json:"evidence"`
}

// NewGameGeometry 是開場主選單、姓名盤、性別與能力確認畫面的資料契約。
// 它只描述幾何與固定 anchor，不承載流程程式碼；缺欄位時 production
// bootstrap 必須 fail closed，不能回退到 Go 內的 DQ3 座標。
type NewGameGeometry struct {
	ID                    string                  `json:"id"`
	Frame                 *FrameStyle             `json:"frame,omitempty"`
	Menu                  GeometryRect            `json:"menu"`
	MenuTitle             GeometryAnchor          `json:"menu_title"`
	MenuOptions           GeometryAnchor          `json:"menu_options"`
	NamePanel             GeometryRect            `json:"name_panel"`
	NameTitle             GeometryAnchor          `json:"name_title"`
	NameLeftArrow         GeometryAnchor          `json:"name_left_arrow"`
	NameRightArrow        GeometryAnchor          `json:"name_right_arrow"`
	NameText              GeometryAnchor          `json:"name_text"`
	NameGrid              GeometryGrid            `json:"name_grid"`
	NameFunctionPanel     GeometryRect            `json:"name_function_panel"`
	NameFunction          GeometryAnchor          `json:"name_function"`
	NameModePanel         GeometryRect            `json:"name_mode_panel"`
	NameMode              GeometryAnchor          `json:"name_mode"`
	GenderPanel           GeometryRect            `json:"gender_panel"`
	Gender                GeometryAnchor          `json:"gender"`
	ConfirmPrompt         GeometryRect            `json:"confirm_prompt"`
	ConfirmChoice         GeometryRect            `json:"confirm_choice"`
	ConfirmChoiceContent  GeometryRect            `json:"confirm_choice_content"`
	ConfirmChoiceFrame    *FrameStyle             `json:"confirm_choice_frame,omitempty"`
	ConfirmChoiceBackdrop *PatternFill            `json:"confirm_choice_backdrop,omitempty"`
	StatsLeft             GeometryRect            `json:"stats_left"`
	StatsEquipment        GeometryRect            `json:"stats_equipment"`
	StatsRight            GeometryRect            `json:"stats_right"`
	StatsName             GeometryAnchor          `json:"stats_name"`
	StatsHero             GeometryAnchor          `json:"stats_hero"`
	StatsSex              GeometryAnchor          `json:"stats_sex"`
	StatsSexValue         GeometryAnchor          `json:"stats_sex_value"`
	StatsCloth            GeometryAnchor          `json:"stats_cloth"`
	Stats                 *NewGameStatLayout      `json:"stats,omitempty"`
	StatsPrompt           GeometryAnchor          `json:"stats_prompt"`
	StatsChoice           GeometryAnchor          `json:"stats_choice"`
	StatsChoiceCursor     GeometryAnchor          `json:"stats_choice_cursor"`
	RawWindows            []RawNewGameWindow      `json:"raw_windows"`
	WindowBackdrops       []NewGameWindowBackdrop `json:"window_backdrops"`
	Evidence              Evidence                `json:"evidence"`
}

// Entries returns copies so callers cannot mutate the validated pack through
// a renderer fixture.
func (l *NewGameLabels) Entries() map[string][]int {
	if l == nil {
		return nil
	}
	return map[string][]int{
		"title": append([]int(nil), l.Title...), "start": append([]int(nil), l.Start...),
		"load": append([]int(nil), l.Load...), "male": append([]int(nil), l.Male...),
		"female": append([]int(nil), l.Female...), "level": append([]int(nil), l.Level...),
		"hp": append([]int(nil), l.HP...), "mp": append([]int(nil), l.MP...),
		"strength": append([]int(nil), l.Strength...), "agility": append([]int(nil), l.Agility...),
		"vitality": append([]int(nil), l.Vitality...), "intelligence": append([]int(nil), l.Intelligence...),
		"luck": append([]int(nil), l.Luck...), "max_hp": append([]int(nil), l.MaxHP...),
		"max_mp": append([]int(nil), l.MaxMP...), "attack": append([]int(nil), l.Attack...),
		"defense": append([]int(nil), l.Defense...), "experience": append([]int(nil), l.Experience...),
		"sex": append([]int(nil), l.Sex...), "hero": append([]int(nil), l.Hero...),
		"cloth": append([]int(nil), l.Cloth...), "prompt": append([]int(nil), l.Prompt...),
		"yes": append([]int(nil), l.Yes...), "no": append([]int(nil), l.No...),
		"choice_cursor": append([]int(nil), l.ChoiceCursor...),
		"backspace":     append([]int(nil), l.Backspace...), "ok": append([]int(nil), l.OK...),
		"name_title":        append([]int(nil), l.NameTitle...),
		"name_left_arrow":   append([]int(nil), l.NameLeftArrow...),
		"name_right_arrow":  append([]int(nil), l.NameRightArrow...),
		"name_zhuyin_grid":  append([]int(nil), l.NameZhuyinGrid...),
		"name_alnum_grid":   append([]int(nil), l.NameAlnumGrid...),
		"input_mode_zhuyin": append([]int(nil), l.InputModeZhuyin...),
	}
}

// BattleCommandLabel 是一個原版雙 glyph 指令標籤。指令的流程語意由引擎
// 共用；實際字模仍由 versioned game pack 提供，避免把某一版的中文字留在 Go。
type BattleCommandLabel struct {
	PrimaryGlyph   int      `json:"primary_glyph"`
	SecondaryGlyph int      `json:"secondary_glyph"`
	Evidence       Evidence `json:"evidence"`
}

// BattleCommandLabels 保存固定的跨版本指令角色對映。欄位刻意使用具名
// primitive，不接受任意 JSON code，避免資料包偷偷承載可執行邏輯。
type BattleCommandLabels struct {
	War    BattleCommandLabel `json:"war"`
	Flee   BattleCommandLabel `json:"flee"`
	Defend BattleCommandLabel `json:"defend"`
	Item   BattleCommandLabel `json:"item"`
	Spell  BattleCommandLabel `json:"spell"`
}

// FieldCommandLabels 是地表命令窗的固定雙字模標籤。命令的輸入語意由
// 共用引擎保留；實際字模仍由 versioned game pack 提供。
type FieldCommandLabels struct {
	Title   BattleCommandLabel `json:"title"`
	Talk    BattleCommandLabel `json:"talk"`
	Spell   BattleCommandLabel `json:"spell"`
	Status  BattleCommandLabel `json:"status"`
	Item    BattleCommandLabel `json:"item"`
	Equip   BattleCommandLabel `json:"equip"`
	Examine BattleCommandLabel `json:"examine"`
}

// Entries returns a copy keyed by stable field-command roles.
func (l *FieldCommandLabels) Entries() map[string]BattleCommandLabel {
	if l == nil {
		return nil
	}
	return map[string]BattleCommandLabel{
		"title": l.Title, "talk": l.Talk, "spell": l.Spell,
		"status": l.Status, "item": l.Item, "equip": l.Equip,
		"examine": l.Examine,
	}
}

// FieldStatusLayout 是地表「狀況→看各人的狀況」詳細面板的資料契約。
// 原版先以 record 407 畫出 22×12 字格，再由角色 record 的欄位填入數值；
// 引擎只負責套用這個跨版本可重用的欄位 primitive，所有座標、文字與職業／性別
// 字模仍由 versioned game pack 提供。
type FieldStatusLayout struct {
	ID              string         `json:"id"`
	Window          WindowLayout   `json:"window"`
	TextID          string         `json:"text_id"`
	HeroClassGlyphs []int          `json:"hero_class_glyphs"`
	MaleGlyphs      []int          `json:"male_glyphs"`
	FemaleGlyphs    []int          `json:"female_glyphs"`
	Name            GeometryAnchor `json:"name"`
	Class           GeometryAnchor `json:"class"`
	Sex             GeometryAnchor `json:"sex"`
	Level           GeometryAnchor `json:"level"`
	CurrentHP       GeometryAnchor `json:"current_hp"`
	CurrentMP       GeometryAnchor `json:"current_mp"`
	Strength        GeometryAnchor `json:"strength"`
	Agility         GeometryAnchor `json:"agility"`
	Vitality        GeometryAnchor `json:"vitality"`
	Intelligence    GeometryAnchor `json:"intelligence"`
	Luck            GeometryAnchor `json:"luck"`
	MaxHP           GeometryAnchor `json:"max_hp"`
	MaxMP           GeometryAnchor `json:"max_mp"`
	Attack          GeometryAnchor `json:"attack"`
	Defense         GeometryAnchor `json:"defense"`
	Experience      GeometryAnchor `json:"experience"`
	Evidence        Evidence       `json:"evidence"`
}

// Entries returns a copy keyed by stable engine command roles.
func (l *BattleCommandLabels) Entries() map[string]BattleCommandLabel {
	if l == nil {
		return nil
	}
	return map[string]BattleCommandLabel{
		"war": l.War, "flee": l.Flee, "defend": l.Defend,
		"item": l.Item, "spell": l.Spell,
	}
}

// AttractFrame 是標題閒置後輪播的一張版本專屬全螢幕 PCX。引擎只依穩定的
// asset key 與資料化停留幀數輪播，不在 Go 內知道職業名稱或檔名。
type AttractFrame struct {
	AssetKey   string   `json:"asset_key"`
	HoldFrames int      `json:"hold_frames"`
	Evidence   Evidence `json:"evidence"`
}

// AttractSequence 描述原版標題後的職業巡禮。start_delay_frames 與 hold_frames
// 都是 pack 設定，缺資料時不啟動 attract（fail closed）。
type AttractSequence struct {
	ID               string         `json:"id"`
	StartDelayFrames int            `json:"start_delay_frames"`
	Frames           []AttractFrame `json:"frames"`
	Evidence         Evidence       `json:"evidence"`
}

// OpeningSequence 描述開機後、標題 splash 前的版本專屬全螢幕 PCX 過場。
// 引擎只知道可跳過的 frame sequence；素材、順序、停留幀數與證據由 game pack 提供。
type OpeningSequence struct {
	ID          string         `json:"id"`
	SkipOnInput bool           `json:"skip_on_input"`
	Frames      []AttractFrame `json:"frames"`
	Evidence    Evidence       `json:"evidence"`
}

// RawScreenAsset describes a version-owned planar screen whose bytes and
// palette are separate. The engine only knows the format primitive; the pack
// supplies the asset key, dimensions, palette and evidence.
type RawScreenAsset struct {
	ID       string     `json:"id"`
	AssetKey string     `json:"asset_key"`
	Width    int        `json:"width"`
	Height   int        `json:"height"`
	Format   string     `json:"format"`
	Palette  [][3]uint8 `json:"palette"`
	Evidence Evidence   `json:"evidence"`
}

type Interface struct {
	SchemaVersion       string               `json:"schema_version"`
	Dialogue            WindowLayout         `json:"dialogue"`
	BattleMessage       WindowLayout         `json:"battle_message,omitempty"`
	BattleCommand       BattlePanelLayout    `json:"battle_command,omitempty"`
	BattleEnemy         BattlePanelLayout    `json:"battle_enemy,omitempty"`
	BattleScene         *BattleSceneLayout   `json:"battle_scene,omitempty"`
	NewGameLabels       *NewGameLabels       `json:"new_game_labels,omitempty"`
	NewGameGeometry     *NewGameGeometry     `json:"new_game_geometry,omitempty"`
	BattleCommandLabels *BattleCommandLabels `json:"battle_command_labels,omitempty"`
	FieldCommandLabels  *FieldCommandLabels  `json:"field_command_labels,omitempty"`
	FieldStatus         *FieldStatusLayout   `json:"field_status,omitempty"`
	PartyHUD            PartyHUDLayout       `json:"party_hud,omitempty"`
	Opening             *OpeningSequence     `json:"opening,omitempty"`
	Attract             *AttractSequence     `json:"attract,omitempty"`
	NewGameConfirmation *RawScreenAsset      `json:"new_game_confirmation,omitempty"`
	BattleTexts         *BattleTextRefs      `json:"battle_texts,omitempty"`
}

// BattleTextRefs assigns stable engine roles to version-owned D3TXT records.
// The role names are cross-version semantics; only the referenced text IDs
// and their original glyph/control words belong to a game pack.
type BattleTextRefs struct {
	StatusSleeping   string `json:"status_sleeping"`
	StatusWoke       string `json:"status_woke"`
	NoSpell          string `json:"no_spell"`
	MPInsufficient   string `json:"mp_insufficient"`
	ActorFled        string `json:"actor_fled"`
	ActorHealed      string `json:"actor_healed"`
	ActorMissed      string `json:"actor_missed"`
	ActorDamage      string `json:"actor_damage"`
	ActorAttack      string `json:"actor_attack"`
	Victory          string `json:"victory"`
	ExpReward        string `json:"exp_reward"`
	GoldReward       string `json:"gold_reward"`
	SpellSealed      string `json:"spell_sealed"`
	EnemySleep       string `json:"enemy_sleep"`
	BuffAttack       string `json:"buff_attack"`
	BuffDefense      string `json:"buff_defense"`
	EnemySealed      string `json:"enemy_sealed"`
	EnemyConfused    string `json:"enemy_confused"`
	CurePoison       string `json:"cure_poison"`
	CureStatus       string `json:"cure_status"`
	SpellNoEffect    string `json:"spell_no_effect"`
	ItemEmpty        string `json:"item_empty"`
	PalpunteBlowAway string `json:"palpunte_blow_away"`
	PalpunteDrainMP  string `json:"palpunte_drain_mp"`
	PartyDefeated    string `json:"party_defeated"`
}

// IDs returns a copy of the role → pack text ID mapping for the engine
// bootstrap. Empty role references are intentionally retained so startup can
// fail closed with the role name rather than silently selecting a fallback.
func (r *BattleTextRefs) IDs() map[string]string {
	if r == nil {
		return nil
	}
	return map[string]string{
		"status_sleeping":    r.StatusSleeping,
		"status_woke":        r.StatusWoke,
		"no_spell":           r.NoSpell,
		"mp_insufficient":    r.MPInsufficient,
		"actor_fled":         r.ActorFled,
		"actor_healed":       r.ActorHealed,
		"actor_missed":       r.ActorMissed,
		"actor_damage":       r.ActorDamage,
		"actor_attack":       r.ActorAttack,
		"victory":            r.Victory,
		"exp_reward":         r.ExpReward,
		"gold_reward":        r.GoldReward,
		"spell_sealed":       r.SpellSealed,
		"enemy_sleep":        r.EnemySleep,
		"buff_attack":        r.BuffAttack,
		"buff_defense":       r.BuffDefense,
		"enemy_sealed":       r.EnemySealed,
		"enemy_confused":     r.EnemyConfused,
		"cure_poison":        r.CurePoison,
		"cure_status":        r.CureStatus,
		"spell_no_effect":    r.SpellNoEffect,
		"item_empty":         r.ItemEmpty,
		"palpunte_blow_away": r.PalpunteBlowAway,
		"palpunte_drain_mp":  r.PalpunteDrainMP,
		"party_defeated":     r.PartyDefeated,
	}
}

type TileCoordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type SceneTileTrigger struct {
	Kind      string           `json:"kind"`
	CTYRaw    int              `json:"cty_raw"`
	Section   int              `json:"section"`
	TileSubID int              `json:"tile_subid"`
	Tiles     []TileCoordinate `json:"tiles"`
}

type DialogueTextIDs struct {
	Intro     string `json:"intro"`
	Apology   string `json:"apology"`
	Accept    string `json:"accept"`
	Reject    string `json:"reject"`
	ChoiceYes string `json:"choice_yes"`
	ChoiceNo  string `json:"choice_no"`
}

type FormationGroup struct {
	MonsterRawID int `json:"monster_raw_id"`
	Count        int `json:"count"`
}

type BattleFormation struct {
	Background  BattleBackgroundSelector `json:"background"`
	Groups      []FormationGroup         `json:"groups"`
	RawBytesHex string                   `json:"raw_bytes_hex"`
}

type TreasureGate struct {
	CTYRaw       int            `json:"cty_raw"`
	Section      int            `json:"section"`
	Tile         TileCoordinate `json:"tile"`
	ItemRawID    int            `json:"item_raw_id"`
	WhileFlagSet int            `json:"while_flag_set"`
}

// BossSurrenderEvent is a finite engine primitive shared by editions that use
// intro -> battle -> apology -> repeatable yes/no surrender. It contains data,
// never Go symbols or executable expressions.
type BossSurrenderEvent struct {
	ID              string           `json:"id"`
	Kind            string           `json:"kind"`
	Trigger         SceneTileTrigger `json:"trigger"`
	PresenceFlagRaw int              `json:"presence_flag_raw"`
	DialogueTextIDs DialogueTextIDs  `json:"dialogue_text_ids"`
	Formation       BattleFormation  `json:"formation"`
	ClearFlagRaw    int              `json:"clear_flag_raw"`
	TreasureGates   []TreasureGate   `json:"treasure_gates"`
	Evidence        Evidence         `json:"evidence"`
}

// ScriptedNPCSelector identifies one original scripted-NPC consumer. The
// selector is pack data; the engine only understands the generic tuple.
type ScriptedNPCSelector struct {
	CTYRaw     int            `json:"cty_raw"`
	Section    int            `json:"section"`
	Tile       TileCoordinate `json:"tile"`
	HandlerRaw int            `json:"handler_raw"`
}

// StoryFlagMapCell is the raw map-cell mutation performed by a finite
// story-flag handler. The value remains an original tile byte; the engine
// does not infer a replacement from a screenshot or a C remake.
type StoryFlagMapCell struct {
	CTYRaw   int            `json:"cty_raw"`
	Section  int            `json:"section"`
	Tile     TileCoordinate `json:"tile"`
	ValueRaw int            `json:"value_raw"`
}

// StoryFlagAnimation is a bounded presentation primitive. It carries only
// the confirmed frame bytes and tick cadence; it cannot contain executable
// JSON or an arbitrary renderer expression.
type StoryFlagAnimation struct {
	Position   TileCoordinate `json:"position"`
	TilesRaw   []int          `json:"tiles_raw"`
	FrameTicks int            `json:"frame_ticks"`
}

// StoryFlagRuntimeEvent describes one evidence-backed original handler whose
// state transaction is part of the production game path. Different editions
// may provide different entries while the engine executes only these finite
// primitives.
type StoryFlagRuntimeEvent struct {
	ID                        string                `json:"id"`
	Kind                      string                `json:"kind"`
	HandlerRaw                int                   `json:"handler_raw"`
	NPC                       ScriptedNPCSelector   `json:"npc"`
	SetStoryFlagsRaw          []int                 `json:"set_story_flags_raw"`
	ClearStoryFlagsRaw        []int                 `json:"clear_story_flags_raw"`
	DialogueRecordRaw         int                   `json:"dialogue_record_raw,omitempty"`
	DialogueSelectorRaw       int                   `json:"dialogue_selector_raw,omitempty"`
	SceneReload               bool                  `json:"scene_reload,omitempty"`
	DayNightPhase             int                   `json:"day_night_phase,omitempty"`
	ResetDayNightSteps        bool                  `json:"reset_day_night_steps,omitempty"`
	RebuildWorldScenes        bool                  `json:"rebuild_world_scenes,omitempty"`
	MapCell                   *StoryFlagMapCell     `json:"map_cell,omitempty"`
	Animation                 *StoryFlagAnimation   `json:"animation,omitempty"`
	ParkPosition              *VehicleWorldPosition `json:"park_position,omitempty"`
	PhoenixUnrevivedFlagRaw   int                   `json:"phoenix_unrevived_flag_raw,omitempty"`
	PhoenixTransientFlagRaw   int                   `json:"phoenix_transient_flag_raw,omitempty"`
	PhoenixRevivedDialogueRaw int                   `json:"phoenix_revived_dialogue_raw,omitempty"`
	PhoenixAltarDialogueRaw   int                   `json:"phoenix_altar_dialogue_raw,omitempty"`
	PhoenixOrbDialogueRaw     int                   `json:"phoenix_orb_dialogue_raw,omitempty"`
	PhoenixAltarVisualTileRaw int                   `json:"phoenix_altar_visual_tile_raw,omitempty"`
	PhoenixAltarFlagRaw       []int                 `json:"phoenix_altar_flag_raw,omitempty"`
	PhoenixVisualFlagBaseRaw  int                   `json:"phoenix_visual_flag_base_raw,omitempty"`
	PhoenixOrbItemFirstRaw    int                   `json:"phoenix_orb_item_first_raw,omitempty"`
	PhoenixOrbItemLastRaw     int                   `json:"phoenix_orb_item_last_raw,omitempty"`
	PhoenixAltarTiles         []TileCoordinate      `json:"phoenix_altar_tiles,omitempty"`
	Evidence                  Evidence              `json:"evidence"`
}

type TemporaryRoleTextIDs struct {
	QuestGreeting     string `json:"quest_greeting"`
	QuestRequest      string `json:"quest_request"`
	ReturnPraise      string `json:"return_praise"`
	ForcedOffer       string `json:"forced_offer"`
	ForcedReject      string `json:"forced_reject"`
	Accept            string `json:"accept"`
	LaterIntro        string `json:"later_intro"`
	LaterOffer        string `json:"later_offer"`
	LaterReject       string `json:"later_reject"`
	RestoreIntro      string `json:"restore_intro"`
	ContinueRole      string `json:"continue_role"`
	RestoreReconsider string `json:"restore_reconsider"`
	RestoreReject     string `json:"restore_reject"`
	RestoreAccept     string `json:"restore_accept"`
	ChoiceYes         string `json:"choice_yes"`
	ChoiceNo          string `json:"choice_no"`
}

type GenderedSpriteEntryBases struct {
	Male   int `json:"male"`
	Female int `json:"female"`
}

// TemporaryRoleEvent is the shared finite primitive for a required-item
// return followed by a temporary player role and a separate restore NPC.
// It contains original-edition data only, never executable expressions.
type TemporaryRoleEvent struct {
	ID                string                   `json:"id"`
	Kind              string                   `json:"kind"`
	OfferNPC          ScriptedNPCSelector      `json:"offer_npc"`
	RestoreNPC        ScriptedNPCSelector      `json:"restore_npc"`
	PendingFlagRaw    int                      `json:"pending_flag_raw"`
	RequiredItemRawID int                      `json:"required_item_raw_id"`
	NormalRoleFlagRaw int                      `json:"normal_role_flag_raw"`
	ActiveRoleFlagRaw int                      `json:"active_role_flag_raw"`
	RoleSpriteEntries GenderedSpriteEntryBases `json:"role_sprite_entry_bases"`
	DialogueTextIDs   TemporaryRoleTextIDs     `json:"dialogue_text_ids"`
	Evidence          Evidence                 `json:"evidence"`
}

type TemporarySoloChallengeTextIDs struct {
	Prompt    string `json:"prompt"`
	Accept    string `json:"accept"`
	Reject    string `json:"reject"`
	Completed string `json:"completed"`
	Return    string `json:"return"`
	CaveBack  string `json:"cave_back"`
	CaveEmpty string `json:"cave_empty"`
	ChoiceYes string `json:"choice_yes"`
	ChoiceNo  string `json:"choice_no"`
}

type ScriptedHandlerText struct {
	CTYRaw     int    `json:"cty_raw"`
	Section    int    `json:"section"`
	HandlerRaw int    `json:"handler_raw"`
	TextID     string `json:"text_id"`
}

// TemporarySoloChallengeEvent describes an original event that temporarily
// removes every companion, transfers the hero to a solo world entrance, and
// restores the exact companions through a distinct return NPC. The mode mask
// is retained as raw evidence; the engine uses a named finite state rather
// than exposing arbitrary bit operations from JSON.
type TemporarySoloChallengeEvent struct {
	ID                    string                        `json:"id"`
	Kind                  string                        `json:"kind"`
	EntryNPC              ScriptedNPCSelector           `json:"entry_npc"`
	ReturnNPC             ScriptedNPCSelector           `json:"return_npc"`
	CompletedFlagRaw      int                           `json:"completed_flag_raw"`
	ModeMaskRaw           int                           `json:"mode_mask_raw"`
	SoloWorldPosition     VehicleWorldPosition          `json:"solo_world_position"`
	ReturnTownDestination SceneDestination              `json:"return_town_destination"`
	CaveMessages          []ScriptedHandlerText         `json:"cave_messages"`
	DialogueTextIDs       TemporarySoloChallengeTextIDs `json:"dialogue_text_ids"`
	Evidence              Evidence                      `json:"evidence"`
}

// QuestTreasureSelector identifies one original CTY examine event. The
// one-shot flag uses the original story-bit convention: set means available,
// clear means already collected.
type QuestTreasureSelector struct {
	CTYRaw       int `json:"cty_raw"`
	Section      int `json:"section"`
	TileSubID    int `json:"tile_subid"`
	EventTypeRaw int `json:"event_type_raw"`
	ItemRawID    int `json:"item_raw_id"`
	PresentFlag  int `json:"present_flag_raw"`
}

// TreasureEvent describes one standalone original examine-event reward. It is
// intentionally data-only; collection behavior remains a fixed engine rule.
type TreasureEvent struct {
	ID       string                `json:"id"`
	Kind     string                `json:"kind"`
	Treasure QuestTreasureSelector `json:"treasure"`
	Evidence Evidence              `json:"evidence"`
}

type NPCItemRewardTextIDs struct {
	Before  string `json:"before,omitempty"`
	Success string `json:"success"`
	After   string `json:"after"`
}

// NPCItemRewardEvent describes a finite scripted-NPC transaction. The engine
// searches the party's personal inventories in party order and only clears
// PresentFlagRaw after the item was actually stored.
type NPCItemRewardEvent struct {
	ID                 string               `json:"id"`
	Kind               string               `json:"kind"`
	NPC                ScriptedNPCSelector  `json:"npc"`
	PresentFlagRaw     int                  `json:"present_flag_raw"`
	GrantedItemRaw     int                  `json:"granted_item_raw"`
	RequiredItemRawID  *int                 `json:"required_item_raw_id,omitempty"`
	ReplaceRequired    bool                 `json:"replace_required,omitempty"`
	ClearStoryFlagsRaw []int                `json:"clear_story_flags_raw,omitempty"`
	DialogueTextIDs    NPCItemRewardTextIDs `json:"dialogue_text_ids"`
	Evidence           Evidence             `json:"evidence"`
}

// ItemUseEffect 把版本專屬道具 ID 綁定至有限的引擎 primitive。
// JSON 只提供選擇器與參數，不接受可執行運算式。
type ItemUseEffect struct {
	ID                   string          `json:"id"`
	Kind                 string          `json:"kind"`
	ItemRawID            int             `json:"item_raw_id"`
	EffectID             string          `json:"effect_id"`
	LocationKind         string          `json:"location_kind"`
	DayNightPhase        int             `json:"day_night_phase"`
	ResetDayNightSteps   bool            `json:"reset_day_night_steps"`
	StepCount            int             `json:"step_count"`
	Consume              bool            `json:"consume"`
	RequiredLayer        int             `json:"required_layer,omitempty"`
	RequiredCTYRaw       int             `json:"required_cty_raw,omitempty"`
	RequiredSection      int             `json:"required_section,omitempty"`
	RequiredVehicle      string          `json:"required_vehicle,omitempty"`
	UseTile              *TileCoordinate `json:"use_tile,omitempty"`
	RequiredStoryFlagRaw int             `json:"required_story_flag_raw,omitempty"`
	RequiredStoryFlagSet bool            `json:"required_story_flag_set,omitempty"`
	SetWorldStateMask    int             `json:"set_world_state_mask,omitempty"`
	ClearStoryFlagsRaw   []int           `json:"clear_story_flags_raw,omitempty"`
	GrantItemRawID       int             `json:"grant_item_raw_id,omitempty"`
	SetStoryFlagsRaw     []int           `json:"set_story_flags_raw,omitempty"`
	// MapPatch is the persistent world-state consumer table reapplied after
	// save/load. It is distinct from DirectMapPatch because some original
	// handlers first draw a transient viewport table and later reload a
	// different persistent table (Gaia sword: DGROUP 3B96 vs 3A54).
	MapPatch                *WorldMapPatch `json:"map_patch,omitempty"`
	DirectMapPatch          *WorldMapPatch `json:"direct_map_patch,omitempty"`
	AnimationPaletteModeRaw int            `json:"animation_palette_mode_raw,omitempty"`
	AnimationCycles         int            `json:"animation_cycles,omitempty"`
	DoorKeyTier             int            `json:"door_key_tier,omitempty"`
	Evidence                Evidence       `json:"evidence"`
}

// WorldMapPatch is a finite row-major tile replacement copied from an
// original executable/data table after a validated item-use transaction.
type WorldMapPatch struct {
	Layer    int            `json:"layer"`
	Origin   TileCoordinate `json:"origin"`
	Width    int            `json:"width"`
	Height   int            `json:"height"`
	TilesRaw []int          `json:"tiles_raw"`
}

// TrackingGuardEvent describes a guard that mirrors the player's coordinate
// on a finite trigger strip unless a named temporary state suppresses it.
type TrackingGuardEvent struct {
	ID             string              `json:"id"`
	Kind           string              `json:"kind"`
	CTYRaw         int                 `json:"cty_raw"`
	Section        int                 `json:"section"`
	TriggerTiles   []FloorSwitch       `json:"trigger_tiles"`
	Guard          ScriptedNPCSelector `json:"guard"`
	TrackAxis      string              `json:"track_axis"`
	BypassEffectID string              `json:"bypass_effect_id"`
	Evidence       Evidence            `json:"evidence"`
}

// PushPuzzleNPC identifies one original NPC slot checked by a finite puzzle
// completion handler. InitialTile and CtrlRaw are parity anchors; Slot keeps
// the identity stable after the NPC has moved away from its initial tile.
type PushPuzzleNPC struct {
	Slot        int            `json:"slot"`
	InitialTile TileCoordinate `json:"initial_tile"`
	CtrlRaw     int            `json:"ctrl_raw"`
}

// NPCPushRule describes the original-format runtime control-bit gate shared by
// every pushable NPC in this pack. Puzzle events only describe completion.
type NPCPushRule struct {
	CtrlMask          int      `json:"ctrl_mask"`
	CtrlValue         int      `json:"ctrl_value"`
	BlockedByEffectID string   `json:"blocked_by_effect_id"`
	Evidence          Evidence `json:"evidence"`
}

// PushPuzzleEvent describes a finite scene puzzle whose original handler
// checks selected NPC slots on a named axis and clears one story flag.
type PushPuzzleEvent struct {
	ID                string          `json:"id"`
	Kind              string          `json:"kind"`
	CTYRaw            int             `json:"cty_raw"`
	Section           int             `json:"section"`
	TriggerTiles      []FloorSwitch   `json:"trigger_tiles"`
	NPCs              []PushPuzzleNPC `json:"npcs"`
	CompletionAxis    string          `json:"completion_axis"`
	CompletionValue   int             `json:"completion_value"`
	ClearStoryFlagRaw int             `json:"clear_story_flag_raw"`
	Evidence          Evidence        `json:"evidence"`
}

type ItemExchange struct {
	NPC               ScriptedNPCSelector `json:"npc"`
	RequiredItemRawID int                 `json:"required_item_raw_id"`
	GrantedItemRawID  int                 `json:"granted_item_raw_id"`
	BeforeTextID      string              `json:"before_text_id"`
	SuccessTextID     string              `json:"success_text_id"`
}

// LocationItemUse is a finite, declarative location gate. It deliberately
// cannot contain expressions or callbacks.
type LocationItemUse struct {
	ItemRawID     int    `json:"item_raw_id"`
	LocationKind  string `json:"location_kind"`
	CTYRaw        int    `json:"cty_raw"`
	SetFlagsRaw   []int  `json:"set_flags_raw"`
	ClearFlagsRaw []int  `json:"clear_flags_raw"`
	SuccessTextID string `json:"success_text_id"`
	ReloadScene   bool   `json:"reload_scene"`
}

// QuestItemChainEvent is the shared finite primitive for
// treasure -> NPC item exchange -> location-gated item use. All selectors,
// item IDs, flags and text references are pack data.
type QuestItemChainEvent struct {
	ID       string                `json:"id"`
	Kind     string                `json:"kind"`
	Treasure QuestTreasureSelector `json:"treasure"`
	Exchange ItemExchange          `json:"exchange"`
	Use      LocationItemUse       `json:"use"`
	Evidence Evidence              `json:"evidence"`
}

// ChoiceItemExchangeEvent is a finite two-branch conversation followed by an
// in-place inventory replacement and declarative world-state transaction.
// It cannot embed callbacks or edition-specific engine logic.
type ChoiceItemExchangeTextIDs struct {
	Introduction string `json:"introduction"`
	Interested   string `json:"interested"`
	Hint         string `json:"hint"`
	Offer        string `json:"offer"`
	Success      string `json:"success"`
	After        string `json:"after"`
	Reject       string `json:"reject"`
	ChoiceYes    string `json:"choice_yes"`
	ChoiceNo     string `json:"choice_no"`
}

type ChoiceItemExchangeEvent struct {
	ID                   string                    `json:"id"`
	Kind                 string                    `json:"kind"`
	NPC                  ScriptedNPCSelector       `json:"npc"`
	AvailableFlagRaw     int                       `json:"available_flag_raw"`
	RequiredItemRawID    int                       `json:"required_item_raw_id"`
	GrantedItemRawID     int                       `json:"granted_item_raw_id"`
	ActivateWorldObject  WorldObjectActivation     `json:"activate_world_object"`
	SetWorldStateMaskRaw int                       `json:"set_world_state_mask_raw"`
	ClearStoryFlagsRaw   []int                     `json:"clear_story_flags_raw"`
	DialogueTextIDs      ChoiceItemExchangeTextIDs `json:"dialogue_text_ids"`
	Evidence             Evidence                  `json:"evidence"`
}

type WorldObjectActivation struct {
	ObjectID string               `json:"object_id"`
	Position VehicleWorldPosition `json:"position"`
}

type TrackedWorldObjectTextIDs struct {
	Preamble string `json:"preamble"`
	West     string `json:"west"`
	East     string `json:"east"`
	North    string `json:"north"`
	South    string `json:"south"`
}

type TrackedWorldObjectRelocation struct {
	WindowWidth       int                    `json:"window_width"`
	WindowHeight      int                    `json:"window_height"`
	OriginOffset      int                    `json:"origin_offset"`
	HorizontalKeepMin int                    `json:"horizontal_keep_min"`
	HorizontalKeepMax int                    `json:"horizontal_keep_max"`
	VerticalKeepMin   int                    `json:"vertical_keep_min"`
	VerticalKeepMax   int                    `json:"vertical_keep_max"`
	Candidates        []VehicleWorldPosition `json:"candidates"`
}

// TrackedWorldObject is a finite movable overworld entrance with an original
// item-based direction/distance locator. Runtime coordinates are save data;
// all edition-specific selectors, tiles and visible text remain in the pack.
type TrackedWorldObject struct {
	ID                     string                       `json:"id"`
	Kind                   string                       `json:"kind"`
	ActiveWorldStateMask   int                          `json:"active_world_state_mask_raw"`
	Layer                  int                          `json:"layer"`
	EntranceCTYRaw         int                          `json:"entrance_cty_raw"`
	EntranceSection        int                          `json:"entrance_section"`
	EntryVehicleMode       string                       `json:"entry_vehicle_mode"`
	TileRawByRNGLow2       []int                        `json:"tile_raw_by_rng_low2"`
	TrackerItemRawID       int                          `json:"tracker_item_raw_id"`
	TrackerDialogueTextIDs TrackedWorldObjectTextIDs    `json:"tracker_dialogue_text_ids"`
	Relocation             TrackedWorldObjectRelocation `json:"relocation"`
	Evidence               Evidence                     `json:"evidence"`
}

type CoordinateItemGateTextIDs struct {
	Approach string `json:"approach"`
	Success  string `json:"success"`
}

// CoordinateItemGateEvent is an automatic overworld event at one original
// coordinate. It either clears one story flag when the required item exists,
// or performs a finite forced movement after the approach text.
type CoordinateItemGateEvent struct {
	ID                   string                    `json:"id"`
	Kind                 string                    `json:"kind"`
	Coordinate           VehicleWorldPosition      `json:"coordinate"`
	ActiveStoryFlagRaw   int                       `json:"active_story_flag_raw"`
	RequiredItemRawID    int                       `json:"required_item_raw_id"`
	ClearStoryFlagRaw    int                       `json:"clear_story_flag_raw"`
	ConsumeRequiredItem  bool                      `json:"consume_required_item"`
	FailureDirection     int                       `json:"failure_direction"`
	FailureSteps         int                       `json:"failure_steps"`
	FailureDayNightSteps int                       `json:"failure_day_night_steps"`
	DialogueTextIDs      CoordinateItemGateTextIDs `json:"dialogue_text_ids"`
	Evidence             Evidence                  `json:"evidence"`
}

// FloorSwitch identifies one walk-on switch in a scene. HandlerRaw is retained
// as an original-format parity anchor; the engine selects by the validated
// scene/coordinate tuple and never interprets the raw handler number.
type FloorSwitch struct {
	Tile       TileCoordinate `json:"tile"`
	TileSubID  int            `json:"tile_subid"`
	HandlerRaw int            `json:"handler_raw"`
}

type SceneDestination struct {
	CTYRaw  int `json:"cty_raw"`
	Section int `json:"section"`
	X       int `json:"x"`
	Y       int `json:"y"`
}

type SequenceGateTextIDs struct {
	Prompt    string `json:"prompt"`
	Pressed   string `json:"pressed"`
	Trap      string `json:"trap"`
	Success   string `json:"success"`
	ChoiceYes string `json:"choice_yes"`
	ChoiceNo  string `json:"choice_no"`
}

// TwoStepFloorSwitchGate is a finite engine primitive: the first configured
// floor switch arms the gate, then one configured completion switch succeeds;
// another non-completion switch traps the party. It contains data only.
type TwoStepFloorSwitchGate struct {
	ID                     string                `json:"id"`
	Kind                   string                `json:"kind"`
	CTYRaw                 int                   `json:"cty_raw"`
	Section                int                   `json:"section"`
	Switches               []FloorSwitch         `json:"switches"`
	CompletionTileSubID    int                   `json:"completion_tile_subid"`
	ClearFlagRaw           int                   `json:"clear_flag_raw"`
	TrapTransitionSubIDRaw int                   `json:"trap_transition_subid_raw"`
	TrapDestination        SceneDestination      `json:"trap_destination"`
	UnlockedTreasure       QuestTreasureSelector `json:"unlocked_treasure"`
	DialogueTextIDs        SequenceGateTextIDs   `json:"dialogue_text_ids"`
	SuccessSFXRaw          int                   `json:"success_sfx_raw"`
	Evidence               Evidence              `json:"evidence"`
}

type VehicleWorldPosition struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Layer int `json:"layer"`
}

type VehicleGrant struct {
	Kind          string               `json:"kind"`
	WorldPosition VehicleWorldPosition `json:"world_position"`
	ClearFlagsRaw []int                `json:"clear_flags_raw"`
}

type StagedVehicleExchangeTextIDs struct {
	QuestIntro string `json:"quest_intro"`
	ItemGrant  string `json:"item_grant"`
	NeedItem   string `json:"need_item"`
	Success    string `json:"success"`
	After      string `json:"after"`
}

// StagedVehicleExchangeEvent is a finite shared primitive:
// first visit grants one quest item, later a required item is exchanged for a
// vehicle. Edition-specific selectors, flags, item IDs, text and coordinates
// remain pack data.
type StagedVehicleExchangeEvent struct {
	ID                       string                       `json:"id"`
	Kind                     string                       `json:"kind"`
	NPC                      ScriptedNPCSelector          `json:"npc"`
	QuestPresentFlagRaw      int                          `json:"quest_present_flag_raw"`
	GrantedItemRawID         int                          `json:"granted_item_raw_id"`
	ExchangeAvailableFlagRaw int                          `json:"exchange_available_flag_raw"`
	RequiredItemRawID        int                          `json:"required_item_raw_id"`
	Vehicle                  VehicleGrant                 `json:"vehicle"`
	DialogueTextIDs          StagedVehicleExchangeTextIDs `json:"dialogue_text_ids"`
	Evidence                 Evidence                     `json:"evidence"`
}

type GuidedPassageTextIDs struct {
	Introduction string `json:"introduction"`
	Acceptance   string `json:"acceptance"`
	GuideReady   string `json:"guide_ready"`
	After        string `json:"after"`
}

// GuidedPassageEvent is a finite shared primitive for an NPC who checks a
// required item, walks a fixed route, lets the player follow under normal
// input, then completes on a scene-tile trigger. Paths are declarative cardinal
// waypoints; they cannot contain engine code.
type GuidedPassageEvent struct {
	ID                      string               `json:"id"`
	Kind                    string               `json:"kind"`
	GuideNPC                ScriptedNPCSelector  `json:"guide_npc"`
	CompletedNPC            ScriptedNPCSelector  `json:"completed_npc"`
	InteractionTile         TileCoordinate       `json:"interaction_tile"`
	CompletionTrigger       SceneTileTrigger     `json:"completion_trigger"`
	SceneHandlerRaw         int                  `json:"scene_handler_raw"`
	RequiredItemRawID       int                  `json:"required_item_raw_id"`
	ConsumeRequiredItem     bool                 `json:"consume_required_item"`
	GuidePresentFlagRaw     int                  `json:"guide_present_flag_raw"`
	AnimationPendingFlagRaw int                  `json:"animation_pending_flag_raw"`
	CompletedFlagRaw        int                  `json:"completed_flag_raw"`
	GuidePath               []TileCoordinate     `json:"guide_path"`
	GuideReturnPath         []TileCoordinate     `json:"guide_return_path"`
	GuideEndFacing          string               `json:"guide_end_facing"`
	GuideReturnEndFacing    string               `json:"guide_return_end_facing"`
	StepFrames              int                  `json:"step_frames"`
	GuideMovementRawHex     string               `json:"guide_movement_raw_hex"`
	GuideReturnRawHex       []string             `json:"guide_return_raw_hex"`
	DialogueTextIDs         GuidedPassageTextIDs `json:"dialogue_text_ids"`
	Evidence                Evidence             `json:"evidence"`
}

type HostageRescueTextIDs struct {
	GuardQuestion  string `json:"guard_question"`
	GuardJoin      string `json:"guard_join"`
	GuardFight     string `json:"guard_fight"`
	RescueCry      string `json:"rescue_cry"`
	SwitchPrompt   string `json:"switch_prompt"`
	SwitchPressed  string `json:"switch_pressed"`
	Reunion        string `json:"reunion"`
	ReturnHome     string `json:"return_home"`
	BossArrival    string `json:"boss_arrival"`
	BossChallenge  string `json:"boss_challenge"`
	BossDefeated   string `json:"boss_defeated"`
	BossApology    string `json:"boss_apology"`
	BossReject     string `json:"boss_reject"`
	BossFarewell   string `json:"boss_farewell"`
	RewardOffer    string `json:"reward_offer"`
	RewardDecline  string `json:"reward_decline"`
	RewardReceived string `json:"reward_received"`
	ChoiceYes      string `json:"choice_yes"`
	ChoiceNo       string `json:"choice_no"`
}

type ScriptedMovement struct {
	NPC         ScriptedNPCSelector `json:"npc"`
	Path        []TileCoordinate    `json:"path"`
	EndFacing   string              `json:"end_facing"`
	RawBytesHex string              `json:"raw_bytes_hex"`
}

// HostageRescueEvent is a deliberately finite primitive for the common
// guard -> captive switch -> returning boss -> surrender -> reward-NPC flow.
// Edition-specific selectors, formations, flags, movement and text remain
// declarative pack data; JSON cannot provide callbacks or executable code.
type HostageRescueEvent struct {
	ID                      string                `json:"id"`
	Kind                    string                `json:"kind"`
	GuardNPCs               []ScriptedNPCSelector `json:"guard_npcs"`
	GuardPresentFlagRaw     int                   `json:"guard_present_flag_raw"`
	GuardFormation          BattleFormation       `json:"guard_formation"`
	ApproachTrigger         SceneTileTrigger      `json:"approach_trigger"`
	ApproachPendingFlagRaw  int                   `json:"approach_pending_flag_raw"`
	ApproachMovement        ScriptedMovement      `json:"approach_movement"`
	SwitchTrigger           SceneTileTrigger      `json:"switch_trigger"`
	CaptivePresentFlagRaw   int                   `json:"captive_present_flag_raw"`
	CaptiveMovements        []ScriptedMovement    `json:"captive_movements"`
	BossNPCs                []ScriptedNPCSelector `json:"boss_npcs"`
	BossPresentFlagRaw      int                   `json:"boss_present_flag_raw"`
	BossFormation           BattleFormation       `json:"boss_formation"`
	CompletionSetFlagsRaw   []int                 `json:"completion_set_flags_raw"`
	CompletionClearFlagsRaw []int                 `json:"completion_clear_flags_raw"`
	RewardNPC               ScriptedNPCSelector   `json:"reward_npc"`
	RewardAvailableFlagRaw  int                   `json:"reward_available_flag_raw"`
	RewardItemRawID         int                   `json:"reward_item_raw_id"`
	StepFrames              int                   `json:"step_frames"`
	DialogueTextIDs         HostageRescueTextIDs  `json:"dialogue_text_ids"`
	Evidence                Evidence              `json:"evidence"`
}

type ReclassTextIDs struct {
	Introduction      string `json:"introduction"`
	Decline           string `json:"decline"`
	ChooseMember      string `json:"choose_member"`
	HeroForbidden     string `json:"hero_forbidden"`
	LevelInsufficient string `json:"level_insufficient"`
	ChooseClass       string `json:"choose_class"`
	SameClass         string `json:"same_class"`
	ConfirmClass      string `json:"confirm_class"`
	ConfirmLevelReset string `json:"confirm_level_reset"`
	Success           string `json:"success"`
	Farewell          string `json:"farewell"`
	ClassMenuBasic    string `json:"class_menu_basic"`
	ClassMenuSage     string `json:"class_menu_sage"`
	ChoiceYes         string `json:"choice_yes"`
	ChoiceNo          string `json:"choice_no"`
}

// ReclassEvent is a finite class-change transaction. The engine owns the
// named original behavior; edition-specific selectors, raw classes, level,
// catalyst item and text remain pack data.
type ReclassEvent struct {
	ID                      string              `json:"id"`
	Kind                    string              `json:"kind"`
	NPC                     ScriptedNPCSelector `json:"npc"`
	MinimumLevel            int                 `json:"minimum_level"`
	ForbiddenSourceClasses  []int               `json:"forbidden_source_classes_raw"`
	BasicTargetClasses      []int               `json:"basic_target_classes_raw"`
	AdvancedTargetClass     int                 `json:"advanced_target_class_raw"`
	AdvancedFreeSourceClass int                 `json:"advanced_free_source_class_raw"`
	AdvancedRequiredItemRaw int                 `json:"advanced_required_item_raw_id"`
	EffectID                string              `json:"effect_id"`
	PersonalInventorySlots  int                 `json:"personal_inventory_slots"`
	ClassNameTextIDsRaw     map[string]string   `json:"class_name_text_ids_raw"`
	DialogueTextIDs         ReclassTextIDs      `json:"dialogue_text_ids"`
	Evidence                Evidence            `json:"evidence"`
}

type StagedBossTextIDs struct {
	FirstPost       string `json:"first_post"`
	SecondOffer     string `json:"second_offer"`
	SecondAccept    string `json:"second_accept"`
	SecondChallenge string `json:"second_challenge"`
	SecondVictory   string `json:"second_victory"`
	ChoiceYes       string `json:"choice_yes"`
	ChoiceNo        string `json:"choice_no"`
}

// StagedBossEvent 是第一戰、原版旗標／自動移動、第二個 NPC 選擇、第二戰及寶箱的
// 有限共用 primitive。版本專屬 formation、selector、旗標、路徑與文字都由 pack 提供。
type StagedBossEvent struct {
	ID                           string              `json:"id"`
	Kind                         string              `json:"kind"`
	FirstNPC                     ScriptedNPCSelector `json:"first_npc"`
	FirstPresenceFlagRaw         int                 `json:"first_presence_flag_raw"`
	FirstFormation               BattleFormation     `json:"first_formation"`
	FirstDropPolicy              string              `json:"first_drop_policy"`
	FirstNPCPostBattleDirections []string            `json:"first_npc_post_battle_directions"`
	FirstVictoryClearFlagsRaw    []int               `json:"first_victory_clear_flags_raw"`
	FirstVictorySetFlagsRaw      []int               `json:"first_victory_set_flags_raw"`
	FirstPostBattleDirections    []string            `json:"first_post_battle_directions"`
	SecondNPC                    ScriptedNPCSelector `json:"second_npc"`
	SecondRequiredFlagRaw        int                 `json:"second_required_flag_raw"`
	SecondFormation              BattleFormation     `json:"second_formation"`
	SecondDropPolicy             string              `json:"second_drop_policy"`
	SecondVictoryClearFlagsRaw   []int               `json:"second_victory_clear_flags_raw"`
	SecondVictorySetFlagsRaw     []int               `json:"second_victory_set_flags_raw"`
	TreasureGates                []TreasureGate      `json:"treasure_gates"`
	StepFrames                   int                 `json:"step_frames"`
	DialogueTextIDs              StagedBossTextIDs   `json:"dialogue_text_ids"`
	Evidence                     Evidence            `json:"evidence"`
}

type ItemActionTextIDs struct {
	Use  string `json:"use"`
	Give string `json:"give"`
	Drop string `json:"drop"`
}

type ItemActions struct {
	PersonalInventorySlots int               `json:"personal_inventory_slots"`
	TextIDs                ItemActionTextIDs `json:"text_ids"`
	Evidence               Evidence          `json:"evidence"`
}

// RuraDestination describes one original teleport-table entry. ArrivalOffset
// is applied to the CTY world location; ShipWorldPosition is the normalized
// per-layer position written when the party owns the ship. Raw IDs and text
// records remain pack data rather than engine constants.
type RuraDestination struct {
	CTYRaw            int                  `json:"cty_raw"`
	NameRecordRaw     int                  `json:"name_record_raw"`
	ArrivalOffset     TileCoordinate       `json:"arrival_offset"`
	ShipWorldPosition VehicleWorldPosition `json:"ship_world_position"`
}

type RuraNavigation struct {
	Destinations []RuraDestination `json:"destinations"`
	Evidence     Evidence          `json:"evidence"`
}

// WorldEntranceVariantBranch is one ordered story-flag decision for a world
// coordinate shared by multiple scene variants.
type WorldEntranceVariantBranch struct {
	FlagRaw     int  `json:"flag_raw"`
	WhenFlagSet bool `json:"when_flag_set"`
	CTYRaw      int  `json:"cty_raw"`
}

// WorldEntranceVariant keeps edition-specific coordinates, raw flags and CTY
// records out of the engine. Branch order mirrors the original executable.
type WorldEntranceVariant struct {
	ID            string                       `json:"id"`
	WorldTile     TileCoordinate               `json:"world_tile"`
	Layer         int                          `json:"layer"`
	Branches      []WorldEntranceVariantBranch `json:"branches"`
	DefaultCTYRaw int                          `json:"default_cty_raw"`
	Evidence      Evidence                     `json:"evidence"`
}

// SettlementFounderTextIDs names the original blocking dialogue sequence used
// by a party-member handoff event. The engine only interprets the sequence;
// every player-visible glyph remains pack data.
type SettlementFounderTextIDs struct {
	Introduction string `json:"introduction"`
	NoEligible   string `json:"no_eligible"`
	FirstOffer   string `json:"first_offer"`
	Reject       string `json:"reject"`
	FinalOffer   string `json:"final_offer"`
	Accepted     string `json:"accepted"`
	After        string `json:"after"`
	StorageFull  string `json:"storage_full"`
	ChoiceYes    string `json:"choice_yes"`
	ChoiceNo     string `json:"choice_no"`
}

// SettlementFounderEvent describes an original NPC transaction that removes
// the first matching active companion, archives that founder separately and
// transfers all carried items into shared storage.
type SettlementFounderEvent struct {
	ID                        string                   `json:"id"`
	Kind                      string                   `json:"kind"`
	NPC                       ScriptedNPCSelector      `json:"npc"`
	RequiredClassRaw          int                      `json:"required_class_raw"`
	RequireAlive              bool                     `json:"require_alive"`
	CompletionFlagRaw         int                      `json:"completion_flag_raw"`
	FounderVisibilityFlagsRaw []int                    `json:"founder_visibility_flags_by_gender_raw"`
	SharedStorageCapacity     int                      `json:"shared_storage_capacity"`
	DialogueTextIDs           SettlementFounderTextIDs `json:"dialogue_text_ids"`
	Evidence                  Evidence                 `json:"evidence"`
}

// SettlementFounderFollowupEvent describes a later NPC conversation that
// inserts the archived founder name and selects an ordered text sequence from
// original story-flag state. It cannot mutate flags or inventory.
type SettlementFounderFollowupEvent struct {
	ID                    string              `json:"id"`
	Kind                  string              `json:"kind"`
	NPC                   ScriptedNPCSelector `json:"npc"`
	RequiredFlagsSetRaw   []int               `json:"required_flags_set_raw"`
	RequiredFlagsClearRaw []int               `json:"required_flags_clear_raw"`
	DialogueTextIDs       []string            `json:"dialogue_text_ids"`
	Evidence              Evidence            `json:"evidence"`
}

// ConditionalFacilityEvent describes a scripted NPC that becomes a normal
// facility only when the original player coordinate matches a narrow gate.
// The fallback dialogue and facility selector are pack data; the engine only
// executes this finite conditional-facility primitive.
type ConditionalFacilityEvent struct {
	ID              string              `json:"id"`
	Kind            string              `json:"kind"`
	NPC             ScriptedNPCSelector `json:"npc"`
	RequiredPlayerY int                 `json:"required_player_y"`
	FacilityK       int                 `json:"facility_k"`
	FallbackTextID  string              `json:"fallback_text_id"`
	Evidence        Evidence            `json:"evidence"`
}

type Events struct {
	SchemaVersion               string                           `json:"schema_version"`
	ItemActions                 ItemActions                      `json:"item_actions"`
	RuraNavigation              RuraNavigation                   `json:"rura_navigation"`
	WorldEntranceVariants       []WorldEntranceVariant           `json:"world_entrance_variants"`
	SettlementFounderEvents     []SettlementFounderEvent         `json:"settlement_founder_events"`
	SettlementFounderFollowups  []SettlementFounderFollowupEvent `json:"settlement_founder_followups"`
	ConditionalFacilityEvents   []ConditionalFacilityEvent       `json:"conditional_facility_events"`
	NPCPushRule                 NPCPushRule                      `json:"npc_push_rule"`
	BossSurrenderEvents         []BossSurrenderEvent             `json:"boss_surrender_events"`
	TemporaryRoleEvents         []TemporaryRoleEvent             `json:"temporary_role_events"`
	TemporarySoloChallenges     []TemporarySoloChallengeEvent    `json:"temporary_solo_challenges"`
	QuestItemChainEvents        []QuestItemChainEvent            `json:"quest_item_chain_events"`
	ChoiceItemExchangeEvents    []ChoiceItemExchangeEvent        `json:"choice_item_exchange_events"`
	TrackedWorldObjects         []TrackedWorldObject             `json:"tracked_world_objects"`
	CoordinateItemGateEvents    []CoordinateItemGateEvent        `json:"coordinate_item_gate_events"`
	TreasureEvents              []TreasureEvent                  `json:"treasure_events"`
	NPCItemRewardEvents         []NPCItemRewardEvent             `json:"npc_item_reward_events"`
	ItemUseEffects              []ItemUseEffect                  `json:"item_use_effects"`
	TrackingGuardEvents         []TrackingGuardEvent             `json:"tracking_guard_events"`
	PushPuzzleEvents            []PushPuzzleEvent                `json:"push_puzzle_events"`
	TwoStepFloorSwitchGates     []TwoStepFloorSwitchGate         `json:"two_step_floor_switch_gates"`
	StagedVehicleExchangeEvents []StagedVehicleExchangeEvent     `json:"staged_vehicle_exchange_events"`
	GuidedPassageEvents         []GuidedPassageEvent             `json:"guided_passage_events"`
	HostageRescueEvents         []HostageRescueEvent             `json:"hostage_rescue_events"`
	ReclassEvents               []ReclassEvent                   `json:"reclass_events"`
	StagedBossEvents            []StagedBossEvent                `json:"staged_boss_events"`
	StoryFlagRuntimeEvents      []StoryFlagRuntimeEvent          `json:"story_flag_runtime_events"`
}

type TextSource struct {
	Kind   string `json:"kind"`
	File   string `json:"file,omitempty"`
	Record *int   `json:"record,omitempty"`
}

type TextLayout struct {
	Kind         string `json:"kind"`
	Columns      int    `json:"columns,omitempty"`
	LinesPerPage int    `json:"lines_per_page,omitempty"`
}

// TextDefinition keeps both the editable Unicode transcription and canonical
// glyph/control words used by the original bitmap-font renderer.
type TextDefinition struct {
	ID         string     `json:"id"`
	Value      string     `json:"value"`
	GlyphCodes []int      `json:"glyph_codes"`
	Layout     TextLayout `json:"layout"`
	Source     TextSource `json:"source"`
	Evidence   Evidence   `json:"evidence"`
}

type Texts struct {
	SchemaVersion string           `json:"schema_version"`
	Definitions   []TextDefinition `json:"definitions"`
}

type EquipmentSlots struct {
	Weapon *int `json:"weapon"`
	Armor  *int `json:"armor"`
	Shield *int `json:"shield"`
	Head   *int `json:"head"`
}

type CharacterDefault struct {
	ID        string         `json:"id"`
	Equipment EquipmentSlots `json:"equipment"`
	Evidence  Evidence       `json:"evidence"`
}

// CharacterDefaultRefs assigns engine-semantic roles to pack-owned defaults.
// The referenced IDs remain pack-specific; the engine never hard-codes them.
type CharacterDefaultRefs struct {
	NewGamePlayer         string `json:"new_game_player"`
	RegisteredPartyMember string `json:"registered_party_member,omitempty"`
}

// PartySpriteEntry maps one pack-owned party class/gender pair to a sprite
// entry in the pack-owned character asset.  The engine never derives this
// mapping from a DQ3 convention; a future pack may use a different asset or
// ordering.
type PartySpriteEntry struct {
	ClassRaw  int      `json:"class_raw"`
	GenderRaw int      `json:"gender_raw"`
	EntryBase int      `json:"entry_base"`
	Evidence  Evidence `json:"evidence"`
}

type PartySpriteDefinition struct {
	Asset   string             `json:"asset"`
	Entries []PartySpriteEntry `json:"entries"`
}

type Characters struct {
	SchemaVersion string                `json:"schema_version"`
	DefaultRefs   CharacterDefaultRefs  `json:"default_refs"`
	Defaults      []CharacterDefault    `json:"defaults"`
	PartySprite   PartySpriteDefinition `json:"party_sprite,omitempty"`
}

type Pack struct {
	Manifest                   Manifest
	Facilities                 Facilities
	Interface                  Interface
	Events                     Events
	Characters                 Characters
	Texts                      Texts
	Audio                      AudioDefinition
	Battle                     BattlePackData
	services                   map[string]*ServiceDefinition
	bossEvents                 map[string]*BossSurrenderEvent
	roleEvents                 map[string]*TemporaryRoleEvent
	soloChallenges             map[string]*TemporarySoloChallengeEvent
	questItemEvents            map[string]*QuestItemChainEvent
	choiceItemExchangeEvents   map[string]*ChoiceItemExchangeEvent
	trackedWorldObjects        map[string]*TrackedWorldObject
	coordinateItemGateEvents   map[string]*CoordinateItemGateEvent
	treasureEvents             map[string]*TreasureEvent
	itemUseEffects             map[int]*ItemUseEffect
	sequenceGates              map[string]*TwoStepFloorSwitchGate
	vehicleExchanges           map[string]*StagedVehicleExchangeEvent
	guidedPassages             map[string]*GuidedPassageEvent
	hostageRescues             map[string]*HostageRescueEvent
	reclassEvents              map[string]*ReclassEvent
	settlementFounders         map[string]*SettlementFounderEvent
	settlementFounderFollowups map[string]*SettlementFounderFollowupEvent
	stagedBossEvents           map[string]*StagedBossEvent
	charDefaults               map[string]*CharacterDefault
	texts                      map[string]*TextDefinition
	battlePresent              bool
	contentHash                string
}

func decodeStrict(fsys fs.FS, name string, dst any) error {
	raw, err := fs.ReadFile(fsys, name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("decode %s: %w", name, err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("decode %s: trailing JSON value", name)
		}
		return fmt.Errorf("decode %s trailing data: %w", name, err)
	}
	return nil
}

func cleanRelative(name, field string) (string, error) {
	if name == "" || strings.HasPrefix(name, "/") || path.Clean(name) != name ||
		name == ".." || strings.HasPrefix(name, "../") {
		return "", fmt.Errorf("%s must be a clean pack-relative path: %q", field, name)
	}
	return name, nil
}

// Load reads and fully validates one game pack. It returns no partial pack.
func Load(fsys fs.FS) (*Pack, error) {
	var p Pack
	if err := decodeStrict(fsys, "manifest.json", &p.Manifest); err != nil {
		return nil, err
	}
	if err := p.validateManifest(); err != nil {
		return nil, fmt.Errorf("manifest.json: %w", err)
	}
	facilitiesPath, ok := p.Manifest.Data["facilities"]
	if !ok {
		return nil, errors.New("manifest.json: data.facilities is required")
	}
	var err error
	if facilitiesPath, err = cleanRelative(facilitiesPath, "data.facilities"); err != nil {
		return nil, err
	}
	if err := decodeStrict(fsys, facilitiesPath, &p.Facilities); err != nil {
		return nil, err
	}
	if err := p.validateFacilities(); err != nil {
		return nil, fmt.Errorf("%s: %w", facilitiesPath, err)
	}
	interfacePath, ok := p.Manifest.Data["interface"]
	if !ok {
		return nil, errors.New("manifest.json: data.interface is required")
	}
	if interfacePath, err = cleanRelative(interfacePath, "data.interface"); err != nil {
		return nil, err
	}
	if err := decodeStrict(fsys, interfacePath, &p.Interface); err != nil {
		return nil, err
	}
	if err := p.validateInterface(); err != nil {
		return nil, fmt.Errorf("%s: %w", interfacePath, err)
	}
	eventsPath, ok := p.Manifest.Data["events"]
	if !ok {
		return nil, errors.New("manifest.json: data.events is required")
	}
	if eventsPath, err = cleanRelative(eventsPath, "data.events"); err != nil {
		return nil, err
	}
	if err := decodeStrict(fsys, eventsPath, &p.Events); err != nil {
		return nil, err
	}
	if err := p.validateEvents(); err != nil {
		return nil, fmt.Errorf("%s: %w", eventsPath, err)
	}
	charactersPath, ok := p.Manifest.Data["characters"]
	if !ok {
		return nil, errors.New("manifest.json: data.characters is required")
	}
	if charactersPath, err = cleanRelative(charactersPath, "data.characters"); err != nil {
		return nil, err
	}
	if err := decodeStrict(fsys, charactersPath, &p.Characters); err != nil {
		return nil, err
	}
	if err := p.validateCharacters(); err != nil {
		return nil, fmt.Errorf("%s: %w", charactersPath, err)
	}
	textsPath, ok := p.Manifest.Data["texts"]
	if !ok {
		return nil, errors.New("manifest.json: data.texts is required")
	}
	if textsPath, err = cleanRelative(textsPath, "data.texts"); err != nil {
		return nil, err
	}
	if err := decodeStrict(fsys, textsPath, &p.Texts); err != nil {
		return nil, err
	}
	if err := p.validateTexts(); err != nil {
		return nil, fmt.Errorf("%s: %w", textsPath, err)
	}
	if err := p.validateBattleTextRefs(); err != nil {
		return nil, fmt.Errorf("%s: %w", interfacePath, err)
	}
	if audioPath, ok := p.Manifest.Data["audio"]; ok {
		if audioPath, err = cleanRelative(audioPath, "data.audio"); err != nil {
			return nil, err
		}
		if err := decodeStrict(fsys, audioPath, &p.Audio); err != nil {
			return nil, err
		}
		if err := p.validateAudio(); err != nil {
			return nil, fmt.Errorf("%s: %w", audioPath, err)
		}
	}
	if battlePath, ok := p.Manifest.Data["battle"]; ok {
		if battlePath, err = cleanRelative(battlePath, "data.battle"); err != nil {
			return nil, err
		}
		if err := decodeStrict(fsys, battlePath, &p.Battle); err != nil {
			return nil, err
		}
		if err := p.validateBattlePack(); err != nil {
			return nil, fmt.Errorf("%s: %w", battlePath, err)
		}
		p.battlePresent = true
		if err := p.validateBattleFormationSelectors(); err != nil {
			return nil, fmt.Errorf("%s: %w", battlePath, err)
		}
	}
	if err := p.validateEventTextRefs(); err != nil {
		return nil, fmt.Errorf("%s: %w", eventsPath, err)
	}
	canonical, err := json.Marshal(struct {
		Manifest   Manifest        `json:"manifest"`
		Facilities Facilities      `json:"facilities"`
		Interface  Interface       `json:"interface"`
		Events     Events          `json:"events"`
		Characters Characters      `json:"characters"`
		Texts      Texts           `json:"texts"`
		Audio      AudioDefinition `json:"audio"`
		Battle     BattlePackData  `json:"battle"`
	}{p.Manifest, p.Facilities, p.Interface, p.Events, p.Characters, p.Texts, p.Audio, p.Battle})
	if err != nil {
		return nil, fmt.Errorf("canonicalize pack: %w", err)
	}
	p.contentHash = fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	return &p, nil
}

func (p *Pack) validateInterface() error {
	if p.Interface.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Interface.SchemaVersion)
	}
	w := p.Interface.Dialogue
	if w.ID == "" || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Height <= 0 ||
		w.TextInsetX < 0 || w.TextInsetY < 0 ||
		w.TextInsetX*2 >= w.Width || w.TextInsetY*2 >= w.Height ||
		w.Columns <= 0 || w.Columns > 100 || w.LinesPerPage <= 0 || w.LinesPerPage > 20 {
		return errors.New("dialogue window layout is invalid")
	}
	if err := validateEvidence(w.Evidence); err != nil {
		return fmt.Errorf("dialogue evidence: %w", err)
	}
	if err := validateFrameStyle("dialogue", w.Frame); err != nil {
		return err
	}
	if m := p.Interface.BattleMessage; m.ID != "" {
		if m.X < 0 || m.Y < 0 || m.Width <= 0 || m.Height <= 0 ||
			m.TextInsetX < 0 || m.TextInsetY < 0 ||
			m.TextInsetX*2 >= m.Width || m.TextInsetY*2 >= m.Height ||
			m.Columns <= 0 || m.Columns > 100 || m.LinesPerPage <= 0 || m.LinesPerPage > 20 {
			return errors.New("battle message window layout is invalid")
		}
		if err := validateEvidence(m.Evidence); err != nil {
			return fmt.Errorf("battle message evidence: %w", err)
		}
		if err := validateFrameStyle("battle message", m.Frame); err != nil {
			return err
		}
	}
	if c := p.Interface.BattleCommand; c.ID != "" {
		if err := validateBattlePanelLayout("battle command", c, false); err != nil {
			return err
		}
	}
	if e := p.Interface.BattleEnemy; e.ID != "" {
		if err := validateBattlePanelLayout("battle enemy", e, true); err != nil {
			return err
		}
	}
	if s := p.Interface.BattleScene; s != nil {
		if s.ID == "" || s.FieldY0 < 0 || s.FieldY1 <= s.FieldY0 ||
			s.FieldY1 > 4096 || s.GroundY < s.FieldY0 || s.GroundY > s.FieldY1 ||
			s.CursorGlyph < 0 || s.CursorGlyph > 0xffff {
			return errors.New("battle scene layout is invalid")
		}
		if err := validateEvidence(s.Evidence); err != nil {
			return fmt.Errorf("battle scene evidence: %w", err)
		}
	}
	if labels := p.Interface.NewGameLabels; labels != nil {
		for role, glyphs := range labels.Entries() {
			if len(glyphs) == 0 {
				return fmt.Errorf("new-game label %s is empty", role)
			}
			for _, glyph := range glyphs {
				if glyph < 0 || glyph > 0xffff {
					return fmt.Errorf("new-game label %s glyph out of range", role)
				}
			}
		}
		for mode, rows := range map[string][][]int{
			"function_zhuyin": labels.FunctionZhuyin,
			"function_alnum":  labels.FunctionAlnum,
		} {
			if len(rows) != 5 {
				return fmt.Errorf("new-game label %s must have five rows", mode)
			}
			for row, glyphs := range rows {
				if len(glyphs) == 0 {
					return fmt.Errorf("new-game label %s[%d] is empty", mode, row)
				}
				for _, glyph := range glyphs {
					if glyph < 0 || glyph > 0xffff {
						return fmt.Errorf("new-game label %s[%d] glyph out of range", mode, row)
					}
				}
			}
		}
		for mode, glyphs := range map[string][]int{
			"name_zhuyin_grid": labels.NameZhuyinGrid,
			"name_alnum_grid":  labels.NameAlnumGrid,
		} {
			if len(glyphs) != 45 {
				return fmt.Errorf("new-game label %s must have 45 cells", mode)
			}
		}
		if err := validateEvidence(labels.Evidence); err != nil {
			return fmt.Errorf("new-game labels evidence: %w", err)
		}
	}
	if geometry := p.Interface.NewGameGeometry; geometry != nil {
		if err := validateNewGameGeometry(*geometry); err != nil {
			return err
		}
	}
	if labels := p.Interface.BattleCommandLabels; labels != nil {
		for role, label := range labels.Entries() {
			if label.PrimaryGlyph < 0 || label.SecondaryGlyph < 0 {
				return fmt.Errorf("battle command label %s has negative glyph", role)
			}
			if err := validateEvidence(label.Evidence); err != nil {
				return fmt.Errorf("battle command label %s evidence: %w", role, err)
			}
		}
	}
	if labels := p.Interface.FieldCommandLabels; labels != nil {
		for role, label := range labels.Entries() {
			if label.PrimaryGlyph < 0 || label.SecondaryGlyph < 0 {
				return fmt.Errorf("field command label %s has negative glyph", role)
			}
			if err := validateEvidence(label.Evidence); err != nil {
				return fmt.Errorf("field command label %s evidence: %w", role, err)
			}
		}
	}
	if status := p.Interface.FieldStatus; status != nil {
		if err := validateFieldStatusLayout(*status); err != nil {
			return err
		}
	}
	if h := p.Interface.PartyHUD; h.ID != "" {
		if h.X < 0 || h.Y < 0 || h.Width <= 0 || h.Height <= 0 ||
			h.TextInsetX < 0 || h.TextInsetY < 0 ||
			h.TextInsetX*2 >= h.Width || h.TextInsetY*2 >= h.Height ||
			h.Columns <= 0 || h.Columns > 16 || h.LinesPerPage <= 0 || h.LinesPerPage > 8 {
			return errors.New("party HUD window layout is invalid")
		}
		if h.HPLabelGlyph == nil || h.MPLabelGlyph == nil {
			return errors.New("party HUD HP/MP label glyphs are required")
		}
		for name, glyph := range map[string]*int{
			"hp": h.HPLabelGlyph, "mp": h.MPLabelGlyph,
		} {
			if *glyph < 0 || *glyph > 0xffff {
				return fmt.Errorf("party HUD %s label glyph out of range", name)
			}
		}
		columnWidth := (h.Width - h.TextInsetX*2) / h.Columns
		if h.NameInsetX < 0 || h.ValueInsetX < 0 || h.NameMaxGlyphs <= 0 ||
			h.NameInsetX+h.NameMaxGlyphs*16 > columnWidth || h.ValueInsetX >= columnWidth ||
			len(h.ClassGlyphs) == 0 {
			return errors.New("party HUD column content layout is invalid")
		}
		for class, glyphs := range h.ClassGlyphs {
			if len(glyphs) == 0 || len(glyphs) > h.NameMaxGlyphs {
				return fmt.Errorf("party HUD class %d glyph count is invalid", class)
			}
			for _, glyph := range glyphs {
				if glyph < 0 || glyph > 0xffff {
					return fmt.Errorf("party HUD class %d glyph out of range", class)
				}
			}
		}
		if err := validateEvidence(h.Evidence); err != nil {
			return fmt.Errorf("party HUD evidence: %w", err)
		}
		if err := validateFrameStyle("party HUD", h.Frame); err != nil {
			return err
		}
	}
	if s := p.Interface.NewGameConfirmation; s != nil {
		if s.ID == "" || s.AssetKey == "" || s.Width <= 0 || s.Height <= 0 ||
			s.Width%8 != 0 || s.Format != "row_interleaved_4bpp" {
			return errors.New("new-game confirmation raw screen is invalid")
		}
		if _, ok := p.Manifest.Assets[s.AssetKey]; !ok {
			return fmt.Errorf("new-game confirmation: unknown asset key %q", s.AssetKey)
		}
		if len(s.Palette) != 16 {
			return fmt.Errorf("new-game confirmation palette has %d colors, want 16", len(s.Palette))
		}
		if err := validateEvidence(s.Evidence); err != nil {
			return fmt.Errorf("new-game confirmation evidence: %w", err)
		}
	}
	if a := p.Interface.Attract; a != nil {
		if a.ID == "" || a.StartDelayFrames < 0 || len(a.Frames) == 0 {
			return errors.New("attract sequence is invalid")
		}
		if err := validateEvidence(a.Evidence); err != nil {
			return fmt.Errorf("attract evidence: %w", err)
		}
		seenAssets := make(map[string]bool, len(a.Frames))
		for i, frame := range a.Frames {
			if frame.AssetKey == "" || frame.HoldFrames <= 0 {
				return fmt.Errorf("attract.frames[%d]: asset_key and positive hold_frames are required", i)
			}
			if seenAssets[frame.AssetKey] {
				return fmt.Errorf("attract.frames[%d]: duplicate asset_key %q", i, frame.AssetKey)
			}
			if _, ok := p.Manifest.Assets[frame.AssetKey]; !ok {
				return fmt.Errorf("attract.frames[%d]: unknown asset key %q", i, frame.AssetKey)
			}
			if err := validateEvidence(frame.Evidence); err != nil {
				return fmt.Errorf("attract.frames[%d] evidence: %w", i, err)
			}
			seenAssets[frame.AssetKey] = true
		}
	}
	if o := p.Interface.Opening; o != nil {
		if o.ID == "" || len(o.Frames) == 0 {
			return errors.New("opening sequence is invalid")
		}
		if err := validateEvidence(o.Evidence); err != nil {
			return fmt.Errorf("opening evidence: %w", err)
		}
		seenAssets := make(map[string]bool, len(o.Frames))
		for i, frame := range o.Frames {
			if frame.AssetKey == "" || frame.HoldFrames <= 0 {
				return fmt.Errorf("opening.frames[%d]: asset_key and positive hold_frames are required", i)
			}
			if seenAssets[frame.AssetKey] {
				return fmt.Errorf("opening.frames[%d]: duplicate asset_key %q", i, frame.AssetKey)
			}
			if _, ok := p.Manifest.Assets[frame.AssetKey]; !ok {
				return fmt.Errorf("opening.frames[%d]: unknown asset key %q", i, frame.AssetKey)
			}
			if err := validateEvidence(frame.Evidence); err != nil {
				return fmt.Errorf("opening.frames[%d] evidence: %w", i, err)
			}
			seenAssets[frame.AssetKey] = true
		}
	}
	return nil
}

func validateGeometryRect(name string, r GeometryRect) error {
	if r.X < 0 || r.Y < 0 || r.Width <= 0 || r.Height <= 0 ||
		r.X+r.Width > 4096 || r.Y+r.Height > 4096 {
		return fmt.Errorf("new-game geometry %s rect is invalid", name)
	}
	for edge, edgeWidth := range map[string]struct {
		width int
		limit int
	}{
		"left":   {r.FrameEdgeWidths.Left, r.Width},
		"right":  {r.FrameEdgeWidths.Right, r.Width},
		"top":    {r.FrameEdgeWidths.Top, r.Height},
		"bottom": {r.FrameEdgeWidths.Bottom, r.Height},
	} {
		if edgeWidth.width < 0 || edgeWidth.width > edgeWidth.limit {
			return fmt.Errorf("new-game geometry %s frame_edge_widths.%s is invalid", name, edge)
		}
	}
	return nil
}

func validatePatternFill(name string, f *PatternFill) error {
	if f == nil {
		return fmt.Errorf("%s is required", name)
	}
	if f.ID == "" {
		return fmt.Errorf("%s id is required", name)
	}
	if f.Pattern != "solid" && f.Pattern != "checkerboard_1px" {
		return fmt.Errorf("%s pattern %q is unsupported", name, f.Pattern)
	}
	if err := validateGeometryRect(name+".rect", f.Rect); err != nil {
		return err
	}
	if err := validateEvidence(f.Evidence); err != nil {
		return fmt.Errorf("%s evidence: %w", name, err)
	}
	return nil
}

func validateGeometryAnchor(name string, a GeometryAnchor) error {
	if a.X < 0 || a.Y < 0 || a.X > 4096 || a.Y > 4096 ||
		a.StepX < 0 || a.StepY < 0 || a.StepX > 4096 || a.StepY > 4096 {
		return fmt.Errorf("new-game geometry %s anchor is invalid", name)
	}
	return nil
}

func validateNumberField(name string, f NumberField) error {
	if f.X < 0 || f.Y < 0 || f.X > 4096 || f.Y > 4096 || f.Digits <= 0 || f.Digits > 16 {
		return fmt.Errorf("new-game geometry %s number field is invalid", name)
	}
	return nil
}

func validateNewGameStatLayout(s NewGameStatLayout) error {
	for name, field := range map[string]NewGameStatField{
		"level": s.Level, "hp": s.HP, "mp": s.MP,
		"strength": s.Strength, "agility": s.Agility, "vitality": s.Vitality,
		"intelligence": s.Intelligence, "luck": s.Luck, "max_hp": s.MaxHP,
		"max_mp": s.MaxMP, "attack": s.Attack, "defense": s.Defense,
		"experience": s.Experience,
	} {
		if err := validateGeometryAnchor("stats."+name+".label", field.Label); err != nil {
			return err
		}
		if err := validateNumberField("stats."+name+".value", field.Value); err != nil {
			return err
		}
	}
	return nil
}

func validateGeometryGrid(name string, g GeometryGrid) error {
	if g.X < 0 || g.Y < 0 || g.X > 4096 || g.Y > 4096 ||
		g.Columns <= 0 || g.Columns > 64 || g.Rows <= 0 || g.Rows > 64 ||
		g.StepX <= 0 || g.StepY <= 0 || g.StepX > 4096 || g.StepY > 4096 {
		return fmt.Errorf("new-game geometry %s grid is invalid", name)
	}
	return nil
}

func validateNewGameGeometry(g NewGameGeometry) error {
	if g.ID == "" {
		return errors.New("new-game geometry id is required")
	}
	if g.Frame == nil {
		return errors.New("new-game geometry frame is required")
	}
	if err := validateFrameStyle("new-game geometry", g.Frame); err != nil {
		return err
	}
	if err := validateFrameStyle("new-game geometry confirm_choice_frame", g.ConfirmChoiceFrame); err != nil {
		return err
	}
	for name, rect := range map[string]GeometryRect{
		"menu": g.Menu, "name_panel": g.NamePanel,
		"name_function_panel": g.NameFunctionPanel, "name_mode_panel": g.NameModePanel,
		"gender_panel": g.GenderPanel, "confirm_prompt": g.ConfirmPrompt,
		"confirm_choice": g.ConfirmChoice, "confirm_choice_content": g.ConfirmChoiceContent,
		"stats_left":      g.StatsLeft,
		"stats_equipment": g.StatsEquipment, "stats_right": g.StatsRight,
	} {
		if err := validateGeometryRect(name, rect); err != nil {
			return err
		}
	}
	if err := validatePatternFill("new-game geometry confirm_choice_backdrop", g.ConfirmChoiceBackdrop); err != nil {
		return err
	}
	if g.ConfirmChoiceFrame == nil {
		return errors.New("new-game geometry confirm_choice_frame is required")
	}
	backdrop := g.ConfirmChoiceBackdrop.Rect
	content := g.ConfirmChoiceContent
	if content.X < backdrop.X || content.Y < backdrop.Y ||
		content.X+content.Width > backdrop.X+backdrop.Width ||
		content.Y+content.Height > backdrop.Y+backdrop.Height {
		return errors.New("new-game geometry confirm_choice_content must fit backdrop")
	}
	if content.X < g.ConfirmChoice.X || content.Y < g.ConfirmChoice.Y ||
		content.X+content.Width > g.ConfirmChoice.X+g.ConfirmChoice.Width ||
		content.Y+content.Height > g.ConfirmChoice.Y+g.ConfirmChoice.Height {
		return errors.New("new-game geometry confirm_choice_content must fit frame")
	}
	for name, anchor := range map[string]GeometryAnchor{
		"menu_title": g.MenuTitle, "menu_options": g.MenuOptions,
		"name_title": g.NameTitle, "name_left_arrow": g.NameLeftArrow,
		"name_right_arrow": g.NameRightArrow, "name_text": g.NameText,
		"name_function": g.NameFunction,
		"name_mode":     g.NameMode, "gender": g.Gender,
		"stats_name": g.StatsName, "stats_hero": g.StatsHero,
		"stats_sex": g.StatsSex, "stats_sex_value": g.StatsSexValue,
		"stats_cloth":  g.StatsCloth,
		"stats_prompt": g.StatsPrompt,
		"stats_choice": g.StatsChoice, "stats_choice_cursor": g.StatsChoiceCursor,
	} {
		if err := validateGeometryAnchor(name, anchor); err != nil {
			return err
		}
	}
	if err := validateGeometryGrid("name_grid", g.NameGrid); err != nil {
		return err
	}
	if g.Stats == nil {
		return errors.New("new-game geometry stats are required")
	}
	if err := validateNewGameStatLayout(*g.Stats); err != nil {
		return err
	}
	if g.NameGrid.Columns != 9 || g.NameGrid.Rows != 5 {
		return errors.New("new-game geometry name_grid must be 9x5")
	}
	if len(g.RawWindows) == 0 {
		return errors.New("new-game geometry raw_windows are required")
	}
	rawByID := make(map[string]RawNewGameWindow, len(g.RawWindows))
	for i, raw := range g.RawWindows {
		if raw.ID == "" || raw.Address == "" || raw.Flags < 0 || raw.Flags > 0xff ||
			raw.X < 0 || raw.Y < 0 || raw.Width <= 0 || raw.Height <= 0 {
			return fmt.Errorf("new-game geometry raw_windows[%d] is invalid", i)
		}
		if _, exists := rawByID[raw.ID]; exists {
			return fmt.Errorf("new-game geometry raw_windows[%d] duplicates id %q", i, raw.ID)
		}
		rawByID[raw.ID] = raw
	}
	if len(g.WindowBackdrops) == 0 {
		return errors.New("new-game geometry window_backdrops are required")
	}
	seenBackdropID := make(map[string]bool, len(g.WindowBackdrops))
	seenRawWindow := make(map[string]bool, len(g.WindowBackdrops))
	for i, backdrop := range g.WindowBackdrops {
		if backdrop.ID == "" || backdrop.RawWindowID == "" {
			return fmt.Errorf("new-game geometry window_backdrops[%d] id and raw_window_id are required", i)
		}
		if seenBackdropID[backdrop.ID] {
			return fmt.Errorf("new-game geometry window_backdrops[%d] duplicates id %q", i, backdrop.ID)
		}
		if seenRawWindow[backdrop.RawWindowID] {
			return fmt.Errorf("new-game geometry window_backdrops[%d] duplicates raw_window_id %q", i, backdrop.RawWindowID)
		}
		raw, ok := rawByID[backdrop.RawWindowID]
		if !ok {
			return fmt.Errorf("new-game geometry window_backdrops[%d] references unknown raw_window_id %q", i, backdrop.RawWindowID)
		}
		if raw.X > 511 || raw.Width > 512 || raw.X+raw.Width > 512 || raw.Y+raw.Height > 4096 {
			return fmt.Errorf("new-game geometry window_backdrops[%d] raw window %q is outside EGA bounds", i, raw.ID)
		}
		if backdrop.Primitive != "ega_window_black_backdrop" {
			return fmt.Errorf("new-game geometry window_backdrops[%d] primitive %q is unsupported", i, backdrop.Primitive)
		}
		if backdrop.DrawOrder == nil || *backdrop.DrawOrder < 0 || *backdrop.DrawOrder > 8 {
			return fmt.Errorf("new-game geometry window_backdrops[%d] draw_order is invalid", i)
		}
		if backdrop.TerminalRightPixels == nil || backdrop.TerminalBottomPixels == nil {
			return fmt.Errorf("new-game geometry window_backdrops[%d] terminal projection is required", i)
		}
		if *backdrop.TerminalRightPixels < 0 || *backdrop.TerminalRightPixels > 8 ||
			*backdrop.TerminalBottomPixels < 0 || *backdrop.TerminalBottomPixels > 8 {
			return fmt.Errorf("new-game geometry window_backdrops[%d] terminal projection is outside one EGA cell", i)
		}
		if err := validateEvidence(backdrop.Evidence); err != nil {
			return fmt.Errorf("new-game geometry window_backdrops[%d] evidence: %w", i, err)
		}
		seenBackdropID[backdrop.ID] = true
		seenRawWindow[backdrop.RawWindowID] = true
	}
	if err := validateEvidence(g.Evidence); err != nil {
		return fmt.Errorf("new-game geometry evidence: %w", err)
	}
	return nil
}

func validateBattlePanelLayout(name string, p BattlePanelLayout, requireCountY bool) error {
	w := p.WindowLayout
	if w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Height <= 0 ||
		w.TextInsetX < 0 || w.TextInsetY < 0 ||
		w.TextInsetX*2 >= w.Width || w.TextInsetY*2 >= w.Height ||
		w.Columns <= 0 || w.Columns > 100 || w.LinesPerPage <= 0 || w.LinesPerPage > 20 {
		return fmt.Errorf("%s window layout is invalid", name)
	}
	if p.HeightMode != "rows_plus_base" || p.BaseRows < 0 || p.RowHeight <= 0 || p.RowHeight > 128 {
		return fmt.Errorf("%s height rule is invalid", name)
	}
	if requireCountY && p.CountInsetY == nil {
		return fmt.Errorf("%s count_inset_y is required", name)
	}
	values := map[string]int{
		"cursor_inset_x": p.CursorInsetX, "row_inset_y": p.RowInsetY,
		"label_inset_x": p.LabelInsetX, "secondary_label_inset_x": p.SecondaryLabelInsetX,
		"value_inset_x": p.ValueInsetX,
		"name_inset_x":  p.NameInsetX, "count_inset_x": p.CountInsetX,
	}
	if p.CountInsetY != nil {
		values["count_inset_y"] = *p.CountInsetY
	}
	for field, value := range values {
		if value < 0 {
			return fmt.Errorf("%s %s must be non-negative", name, field)
		}
	}
	if err := validateEvidence(w.Evidence); err != nil {
		return fmt.Errorf("%s evidence: %w", name, err)
	}
	if err := validateFrameStyle(name, w.Frame); err != nil {
		return err
	}
	return nil
}

func validateFieldStatusLayout(s FieldStatusLayout) error {
	if s.ID == "" || s.TextID == "" {
		return errors.New("field status id and text_id are required")
	}
	w := s.Window
	if w.ID == "" || w.X < 0 || w.Y < 0 || w.Width <= 0 || w.Height <= 0 ||
		w.TextInsetX < 0 || w.TextInsetY < 0 ||
		w.TextInsetX*2 >= w.Width || w.TextInsetY*2 >= w.Height ||
		w.Columns <= 0 || w.Columns > 100 || w.LinesPerPage <= 0 || w.LinesPerPage > 20 {
		return errors.New("field status window layout is invalid")
	}
	for name, anchor := range map[string]GeometryAnchor{
		"name": s.Name, "class": s.Class, "sex": s.Sex, "level": s.Level,
		"current_hp": s.CurrentHP, "current_mp": s.CurrentMP,
		"strength": s.Strength, "agility": s.Agility, "vitality": s.Vitality,
		"intelligence": s.Intelligence, "luck": s.Luck, "max_hp": s.MaxHP,
		"max_mp": s.MaxMP, "attack": s.Attack, "defense": s.Defense,
		"experience": s.Experience,
	} {
		if err := validateGeometryAnchor("field_status."+name, anchor); err != nil {
			return err
		}
	}
	for name, glyphs := range map[string][]int{
		"hero_class_glyphs": s.HeroClassGlyphs,
		"male_glyphs":       s.MaleGlyphs,
		"female_glyphs":     s.FemaleGlyphs,
	} {
		if len(glyphs) == 0 {
			return fmt.Errorf("field status %s must not be empty", name)
		}
		for _, glyph := range glyphs {
			if glyph < 0 || glyph > 0xffff {
				return fmt.Errorf("field status %s glyph out of range", name)
			}
		}
	}
	if err := validateEvidence(w.Evidence); err != nil {
		return fmt.Errorf("field status window evidence: %w", err)
	}
	if err := validateFrameStyle("field status", w.Frame); err != nil {
		return err
	}
	if err := validateEvidence(s.Evidence); err != nil {
		return fmt.Errorf("field status evidence: %w", err)
	}
	return nil
}

func validateFrameStyle(name string, f *FrameStyle) error {
	if f == nil {
		return nil
	}
	if f.ID == "" {
		return fmt.Errorf("%s frame id is required", name)
	}
	switch f.BorderPattern {
	case "solid", "solid_2px", "beveled_2px", "checkerboard_1px":
		// Named engine primitives only; packs cannot supply arbitrary code.
	case "checkerboard_frame_2px":
		if f.BorderAccentRGB == nil {
			return fmt.Errorf("%s frame border_accent_rgb is required for checkerboard_frame_2px", name)
		}
	default:
		return fmt.Errorf("%s frame border_pattern %q is unsupported", name, f.BorderPattern)
	}
	if err := validateEvidence(f.Evidence); err != nil {
		return fmt.Errorf("%s frame evidence: %w", name, err)
	}
	return nil
}

func (p *Pack) validateCharacters() error {
	if p.Characters.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Characters.SchemaVersion)
	}
	if p.Characters.Defaults == nil {
		return errors.New("defaults must be present")
	}
	p.charDefaults = make(map[string]*CharacterDefault, len(p.Characters.Defaults))
	for i := range p.Characters.Defaults {
		d := &p.Characters.Defaults[i]
		if d.ID == "" {
			return fmt.Errorf("defaults[%d].id is required", i)
		}
		if _, exists := p.charDefaults[d.ID]; exists {
			return fmt.Errorf("duplicate character default id %q", d.ID)
		}
		for name, code := range map[string]*int{
			"weapon": d.Equipment.Weapon, "armor": d.Equipment.Armor,
			"shield": d.Equipment.Shield, "head": d.Equipment.Head,
		} {
			if code != nil && (*code < 0 || *code > 127) {
				return fmt.Errorf("%s equipment.%s out of range", d.ID, name)
			}
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", d.ID, err)
		}
		p.charDefaults[d.ID] = d
	}
	if p.Characters.DefaultRefs.NewGamePlayer == "" {
		return errors.New("default_refs.new_game_player is required")
	}
	if p.charDefaults[p.Characters.DefaultRefs.NewGamePlayer] == nil {
		return fmt.Errorf("default_refs.new_game_player references unknown default %q",
			p.Characters.DefaultRefs.NewGamePlayer)
	}
	hasParty := false
	for _, capability := range p.Manifest.Capabilities {
		hasParty = hasParty || capability == "party"
	}
	if hasParty && p.Characters.DefaultRefs.RegisteredPartyMember == "" {
		return errors.New("party capability requires default_refs.registered_party_member")
	}
	if ref := p.Characters.DefaultRefs.RegisteredPartyMember; ref != "" && p.charDefaults[ref] == nil {
		return fmt.Errorf("default_refs.registered_party_member references unknown default %q", ref)
	}
	if hasParty {
		if p.Characters.PartySprite.Asset == "" || len(p.Characters.PartySprite.Entries) == 0 {
			return errors.New("party capability requires characters.party_sprite")
		}
	}
	seenSprite := make(map[[2]int]bool, len(p.Characters.PartySprite.Entries))
	for i, entry := range p.Characters.PartySprite.Entries {
		if entry.ClassRaw < 0 || entry.GenderRaw < 0 || entry.GenderRaw > 1 || entry.EntryBase < 0 {
			return fmt.Errorf("party_sprite.entries[%d] has invalid class/gender/entry", i)
		}
		key := [2]int{entry.ClassRaw, entry.GenderRaw}
		if seenSprite[key] {
			return fmt.Errorf("duplicate party sprite class/gender %d/%d", entry.ClassRaw, entry.GenderRaw)
		}
		seenSprite[key] = true
		if err := validateEvidence(entry.Evidence); err != nil {
			return fmt.Errorf("party_sprite.entries[%d] evidence: %w", i, err)
		}
	}
	return nil
}

func (p *Pack) validateTexts() error {
	if p.Texts.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Texts.SchemaVersion)
	}
	if p.Texts.Definitions == nil {
		return errors.New("definitions must be present")
	}
	p.texts = make(map[string]*TextDefinition, len(p.Texts.Definitions))
	for i := range p.Texts.Definitions {
		d := &p.Texts.Definitions[i]
		if d.ID == "" || d.Value == "" || len(d.GlyphCodes) == 0 {
			return fmt.Errorf("definitions[%d]: id, value and glyph_codes are required", i)
		}
		if _, exists := p.texts[d.ID]; exists {
			return fmt.Errorf("duplicate text id %q", d.ID)
		}
		for _, code := range d.GlyphCodes {
			if code < 0 || code > 0xffff {
				return fmt.Errorf("%s: glyph code %d out of range", d.ID, code)
			}
		}
		switch d.Layout.Kind {
		case "dialogue":
			if d.Layout.Columns < 1 || d.Layout.LinesPerPage < 1 ||
				d.Source.Kind != "legacy_record" || d.Source.File == "" || d.Source.Record == nil {
				return fmt.Errorf("%s: incomplete dialogue layout/source", d.ID)
			}
		case "menu_label":
			if d.Source.Kind != "glyph_map" {
				return fmt.Errorf("%s: menu_label requires glyph_map source", d.ID)
			}
		case "menu_record":
			if d.Source.Kind != "legacy_record" || d.Source.File == "" || d.Source.Record == nil {
				return fmt.Errorf("%s: menu_record requires legacy_record source", d.ID)
			}
		case "battle_message":
			if d.Source.Kind != "legacy_record" || d.Source.File == "" || d.Source.Record == nil {
				return fmt.Errorf("%s: battle_message requires legacy_record source", d.ID)
			}
		case "field_status":
			if d.Layout.Columns < 1 || d.Layout.LinesPerPage < 1 ||
				d.Source.Kind != "legacy_record" || d.Source.File == "" || d.Source.Record == nil {
				return fmt.Errorf("%s: incomplete field_status layout/source", d.ID)
			}
		default:
			return fmt.Errorf("%s: unsupported layout kind %q", d.ID, d.Layout.Kind)
		}
		if err := validateEvidence(d.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", d.ID, err)
		}
		p.texts[d.ID] = d
	}
	return nil
}

// validateBattleTextRefs is intentionally separate from validateInterface:
// interface.json is decoded before texts.json, while the cross-reference can
// only be checked after the text index exists. Packs that omit the optional
// field remain loadable as data fixtures, but NewGameWithPack rejects them
// before a production battle can start.
func (p *Pack) validateBattleTextRefs() error {
	refs := p.Interface.BattleTexts
	if refs == nil {
		return nil
	}
	for role, id := range refs.IDs() {
		if id == "" {
			return fmt.Errorf("battle_texts.%s is required", role)
		}
		if _, ok := p.texts[id]; !ok {
			return fmt.Errorf("battle_texts.%s references unknown text %q", role, id)
		}
	}
	return nil
}

func (p *Pack) validateEventTextRefs() error {
	for field, id := range map[string]string{
		"use":  p.Events.ItemActions.TextIDs.Use,
		"give": p.Events.ItemActions.TextIDs.Give,
		"drop": p.Events.ItemActions.TextIDs.Drop,
	} {
		if id == "" || p.texts[id] == nil {
			return fmt.Errorf("item_actions.text_ids.%s references unknown text %q", field, id)
		}
	}
	for _, event := range p.Events.BossSurrenderEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"intro": refs.Intro, "apology": refs.Apology, "accept": refs.Accept,
			"reject": refs.Reject, "choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.SettlementFounderEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"introduction": refs.Introduction, "no_eligible": refs.NoEligible,
			"first_offer": refs.FirstOffer, "reject": refs.Reject,
			"final_offer": refs.FinalOffer, "accepted": refs.Accepted,
			"after": refs.After, "storage_full": refs.StorageFull,
			"choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.SettlementFounderFollowups {
		if len(event.DialogueTextIDs) == 0 {
			return fmt.Errorf("%s dialogue_text_ids must not be empty", event.ID)
		}
		for i, id := range event.DialogueTextIDs {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids[%d] references unknown text %q",
					event.ID, i, id)
			}
		}
	}
	for _, event := range p.Events.ConditionalFacilityEvents {
		if event.FallbackTextID == "" || p.texts[event.FallbackTextID] == nil {
			return fmt.Errorf("%s fallback_text_id references unknown text %q",
				event.ID, event.FallbackTextID)
		}
	}
	for _, event := range p.Events.TemporaryRoleEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"quest_greeting": refs.QuestGreeting, "quest_request": refs.QuestRequest,
			"return_praise": refs.ReturnPraise, "forced_offer": refs.ForcedOffer,
			"forced_reject": refs.ForcedReject, "accept": refs.Accept,
			"later_intro": refs.LaterIntro, "later_offer": refs.LaterOffer,
			"later_reject": refs.LaterReject, "restore_intro": refs.RestoreIntro,
			"continue_role": refs.ContinueRole, "restore_reconsider": refs.RestoreReconsider,
			"restore_reject": refs.RestoreReject, "restore_accept": refs.RestoreAccept,
			"choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.TemporarySoloChallenges {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"prompt": refs.Prompt, "accept": refs.Accept, "reject": refs.Reject,
			"completed": refs.Completed, "return": refs.Return,
			"cave_back": refs.CaveBack, "cave_empty": refs.CaveEmpty,
			"choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
		for i, message := range event.CaveMessages {
			if message.TextID == "" || p.texts[message.TextID] == nil {
				return fmt.Errorf("%s cave_messages[%d].text_id references unknown text %q",
					event.ID, i, message.TextID)
			}
		}
	}
	for _, event := range p.Events.QuestItemChainEvents {
		for field, id := range map[string]string{
			"exchange.before_text_id":  event.Exchange.BeforeTextID,
			"exchange.success_text_id": event.Exchange.SuccessTextID,
			"use.success_text_id":      event.Use.SuccessTextID,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s %s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.ChoiceItemExchangeEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"introduction": refs.Introduction, "interested": refs.Interested,
			"hint": refs.Hint, "offer": refs.Offer, "success": refs.Success,
			"after": refs.After, "reject": refs.Reject,
			"choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, object := range p.Events.TrackedWorldObjects {
		refs := object.TrackerDialogueTextIDs
		for field, id := range map[string]string{
			"preamble": refs.Preamble, "west": refs.West, "east": refs.East,
			"north": refs.North, "south": refs.South,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s tracker_dialogue_text_ids.%s references unknown text %q",
					object.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.CoordinateItemGateEvents {
		for field, id := range map[string]string{
			"approach": event.DialogueTextIDs.Approach,
			"success":  event.DialogueTextIDs.Success,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.NPCItemRewardEvents {
		for field, id := range map[string]string{
			"success": event.DialogueTextIDs.Success,
			"after":   event.DialogueTextIDs.After,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.TwoStepFloorSwitchGates {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"prompt": refs.Prompt, "pressed": refs.Pressed, "trap": refs.Trap,
			"success": refs.Success, "choice_yes": refs.ChoiceYes, "choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.StagedVehicleExchangeEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"quest_intro": refs.QuestIntro, "item_grant": refs.ItemGrant,
			"need_item": refs.NeedItem, "success": refs.Success, "after": refs.After,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.GuidedPassageEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"introduction": refs.Introduction, "acceptance": refs.Acceptance,
			"guide_ready": refs.GuideReady, "after": refs.After,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.HostageRescueEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"guard_question": refs.GuardQuestion, "guard_join": refs.GuardJoin,
			"guard_fight": refs.GuardFight, "rescue_cry": refs.RescueCry,
			"switch_prompt": refs.SwitchPrompt, "switch_pressed": refs.SwitchPressed,
			"reunion": refs.Reunion, "return_home": refs.ReturnHome,
			"boss_arrival": refs.BossArrival, "boss_challenge": refs.BossChallenge,
			"boss_defeated": refs.BossDefeated, "boss_apology": refs.BossApology,
			"boss_reject": refs.BossReject, "boss_farewell": refs.BossFarewell,
			"reward_offer": refs.RewardOffer, "reward_decline": refs.RewardDecline,
			"reward_received": refs.RewardReceived, "choice_yes": refs.ChoiceYes,
			"choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	for _, event := range p.Events.ReclassEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"introduction": refs.Introduction, "decline": refs.Decline,
			"choose_member": refs.ChooseMember, "hero_forbidden": refs.HeroForbidden,
			"level_insufficient": refs.LevelInsufficient, "choose_class": refs.ChooseClass,
			"same_class": refs.SameClass, "confirm_class": refs.ConfirmClass,
			"confirm_level_reset": refs.ConfirmLevelReset, "success": refs.Success,
			"farewell": refs.Farewell, "class_menu_basic": refs.ClassMenuBasic,
			"class_menu_sage": refs.ClassMenuSage, "choice_yes": refs.ChoiceYes,
			"choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
		for key, id := range event.ClassNameTextIDsRaw {
			class, err := strconv.Atoi(key)
			if err != nil || class < 0 || class > 255 {
				return fmt.Errorf("%s class_name_text_ids_raw has invalid raw class key %q",
					event.ID, key)
			}
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s class_name_text_ids_raw.%s references unknown text %q",
					event.ID, key, id)
			}
		}
	}
	for _, event := range p.Events.StagedBossEvents {
		refs := event.DialogueTextIDs
		for field, id := range map[string]string{
			"first_post": refs.FirstPost, "second_offer": refs.SecondOffer,
			"second_accept": refs.SecondAccept, "second_challenge": refs.SecondChallenge,
			"second_victory": refs.SecondVictory, "choice_yes": refs.ChoiceYes,
			"choice_no": refs.ChoiceNo,
		} {
			if id == "" || p.texts[id] == nil {
				return fmt.Errorf("%s dialogue_text_ids.%s references unknown text %q",
					event.ID, field, id)
			}
		}
	}
	return nil
}

func validScriptedNPC(s ScriptedNPCSelector) bool {
	return s.CTYRaw >= 0 && s.CTYRaw <= 255 && s.Section >= 0 &&
		s.Tile.X >= 0 && s.Tile.Y >= 0 && s.HandlerRaw >= 0 && s.HandlerRaw <= 255
}

func validateStoryFlagList(eventID, name string, flags []int, seen map[int]string) error {
	if len(flags) == 0 {
		return fmt.Errorf("%s: %s must be present", eventID, name)
	}
	for _, flag := range flags {
		if flag < 0 || flag >= 512 {
			return fmt.Errorf("%s: %s flag %d out of range", eventID, name, flag)
		}
		if previous := seen[flag]; previous != "" {
			return fmt.Errorf("%s: flag %d appears in both %s and %s", eventID, flag, previous, name)
		}
		seen[flag] = name
	}
	return nil
}

func validateStoryFlagRuntimeEvent(e StoryFlagRuntimeEvent) error {
	if e.ID == "" || e.HandlerRaw < 0 || e.HandlerRaw > 255 || !validScriptedNPC(e.NPC) {
		return errors.New("invalid id, handler or NPC selector")
	}
	seen := map[int]string{}
	if err := validateStoryFlagList(e.ID, "set_story_flags_raw", e.SetStoryFlagsRaw, seen); err != nil {
		return err
	}
	if err := validateStoryFlagList(e.ID, "clear_story_flags_raw", e.ClearStoryFlagsRaw, seen); err != nil {
		return err
	}
	if e.DialogueRecordRaw < 0 {
		return fmt.Errorf("%s: dialogue_record_raw must be non-negative", e.ID)
	}
	if e.DialogueSelectorRaw < 0 || e.DialogueSelectorRaw > 0xffff {
		return fmt.Errorf("%s: dialogue_selector_raw out of range", e.ID)
	}
	if e.ResetDayNightSteps && (e.DayNightPhase < 0 || e.DayNightPhase > 3) {
		return fmt.Errorf("%s: day_night_phase out of range", e.ID)
	}
	if e.MapCell != nil {
		if e.MapCell.CTYRaw < 0 || e.MapCell.CTYRaw > 255 || e.MapCell.Section < 0 ||
			e.MapCell.Tile.X < 0 || e.MapCell.Tile.Y < 0 || e.MapCell.ValueRaw < 0 || e.MapCell.ValueRaw > 255 {
			return fmt.Errorf("%s: invalid map_cell", e.ID)
		}
	}
	if e.Animation != nil {
		if e.Animation.Position.X < 0 || e.Animation.Position.Y < 0 ||
			e.Animation.FrameTicks < 1 || e.Animation.FrameTicks > 120 || len(e.Animation.TilesRaw) == 0 {
			return fmt.Errorf("%s: invalid animation", e.ID)
		}
		for _, tile := range e.Animation.TilesRaw {
			if tile < 0 || tile > 255 {
				return fmt.Errorf("%s: animation tile out of range", e.ID)
			}
		}
	}
	if e.ParkPosition != nil && (e.ParkPosition.X < 0 || e.ParkPosition.Y < 0 ||
		e.ParkPosition.Layer < 0 || e.ParkPosition.Layer > 1) {
		return fmt.Errorf("%s: invalid park_position", e.ID)
	}
	switch e.Kind {
	case "post_zoma_ending":
		if !e.SceneReload || e.DialogueRecordRaw == 0 || e.MapCell != nil || e.Animation != nil {
			return fmt.Errorf("%s: invalid post_zoma_ending primitive", e.ID)
		}
	case "phoenix_revival":
		if e.MapCell == nil || e.Animation == nil || e.ParkPosition == nil ||
			e.PhoenixUnrevivedFlagRaw < 0 || e.PhoenixUnrevivedFlagRaw >= 512 ||
			e.PhoenixTransientFlagRaw < 0 || e.PhoenixTransientFlagRaw >= 512 ||
			e.PhoenixRevivedDialogueRaw <= 0 || e.PhoenixAltarDialogueRaw <= 0 || e.PhoenixOrbDialogueRaw <= 0 ||
			e.PhoenixAltarVisualTileRaw < 0 || e.PhoenixAltarVisualTileRaw > 255 ||
			len(e.PhoenixAltarFlagRaw) != 6 || len(e.PhoenixAltarTiles) != 6 ||
			e.PhoenixVisualFlagBaseRaw < 0 || e.PhoenixVisualFlagBaseRaw >= 512 ||
			e.PhoenixOrbItemFirstRaw < 0 || e.PhoenixOrbItemLastRaw < e.PhoenixOrbItemFirstRaw ||
			e.PhoenixOrbItemLastRaw > 255 || e.MapCell.CTYRaw != e.NPC.CTYRaw ||
			e.MapCell.Section != e.NPC.Section {
			return fmt.Errorf("%s: invalid phoenix_revival primitive", e.ID)
		}
		for _, flag := range e.PhoenixAltarFlagRaw {
			if flag < 0 || flag >= 512 {
				return fmt.Errorf("%s: phoenix altar flag out of range", e.ID)
			}
		}
		for _, tile := range e.PhoenixAltarTiles {
			if tile.X < 0 || tile.Y < 0 {
				return fmt.Errorf("%s: phoenix altar tile out of range", e.ID)
			}
		}
	case "boss_aftermath":
		if !e.ResetDayNightSteps || !e.RebuildWorldScenes || e.Animation != nil || e.MapCell != nil {
			return fmt.Errorf("%s: invalid boss_aftermath primitive", e.ID)
		}
	case "conditional_story_flag":
		if !e.SceneReload || e.DialogueRecordRaw == 0 || e.Animation != nil || e.MapCell != nil {
			return fmt.Errorf("%s: invalid conditional_story_flag primitive", e.ID)
		}
	default:
		return fmt.Errorf("%s: unsupported story flag event kind %q", e.ID, e.Kind)
	}
	return validateEvidence(e.Evidence)
}

func validTreasureSelector(t QuestTreasureSelector) bool {
	return t.CTYRaw >= 0 && t.CTYRaw <= 255 && t.Section >= 0 &&
		t.TileSubID >= 0 && t.TileSubID <= 31 &&
		(t.EventTypeRaw == 1 || t.EventTypeRaw == 3) &&
		t.ItemRawID >= 0 && t.ItemRawID <= 255 &&
		t.PresentFlag >= 0 && t.PresentFlag < 512
}

func validatePackFormation(eventID, name string, formation BattleFormation) error {
	raw, err := hex.DecodeString(formation.RawBytesHex)
	if err != nil || len(formation.Groups) == 0 ||
		len(raw) != 3+len(formation.Groups)*2 ||
		int(raw[0]) != len(formation.Groups) ||
		int(raw[1]) != formation.Background.PageRaw || int(raw[2]) != formation.Background.PaletteBankRaw {
		return fmt.Errorf("%s: invalid %s", eventID, name)
	}
	for i, group := range formation.Groups {
		if group.MonsterRawID < 0 || group.MonsterRawID > 255 || group.Count < 1 ||
			int(raw[3+i*2]) != group.MonsterRawID || int(raw[4+i*2]) != group.Count {
			return fmt.Errorf("%s: invalid %s group %d", eventID, name, i)
		}
	}
	return nil
}

// fixedBattleFormation 將原版固定編隊的 raw record 解成共用的有限 formation
// primitive。page／palette 與 group 都逐 byte 保留 raw parity；這不是以怪物
// 名稱或場景猜測背景。
func fixedBattleFormation(record BattleFixedFormationRecord) (BattleFormation, error) {
	if record.Count <= 0 {
		return BattleFormation{}, errors.New("count must be positive")
	}
	raw, err := hex.DecodeString(record.RawHex)
	if err != nil {
		return BattleFormation{}, fmt.Errorf("raw_hex is invalid: %w", err)
	}
	if len(raw) != 3+record.Count*2 {
		return BattleFormation{}, fmt.Errorf("raw length=%d want %d", len(raw), 3+record.Count*2)
	}
	if int(raw[0]) != record.Count {
		return BattleFormation{}, fmt.Errorf("raw count=%d want %d", raw[0], record.Count)
	}
	formation := BattleFormation{
		Background: BattleBackgroundSelector{
			PageRaw:        int(raw[1]),
			PaletteBankRaw: int(raw[2]),
		},
		Groups:      make([]FormationGroup, record.Count),
		RawBytesHex: record.RawHex,
	}
	for i := range formation.Groups {
		formation.Groups[i] = FormationGroup{
			MonsterRawID: int(raw[3+i*2]),
			Count:        int(raw[4+i*2]),
		}
		if formation.Groups[i].Count < 1 {
			return BattleFormation{}, fmt.Errorf("group %d count must be positive", i)
		}
	}
	return formation, nil
}

func (p *Pack) validateEvents() error {
	if p.Events.ItemActions.PersonalInventorySlots < 1 ||
		p.Events.ItemActions.PersonalInventorySlots > 32 {
		return errors.New("item_actions.personal_inventory_slots must be 1..32")
	}
	if err := validateEvidence(p.Events.ItemActions.Evidence); err != nil {
		return fmt.Errorf("item_actions evidence: %w", err)
	}
	if p.Events.RuraNavigation.Destinations == nil {
		return errors.New("rura_navigation.destinations must be present")
	}
	if err := validateEvidence(p.Events.RuraNavigation.Evidence); err != nil {
		return fmt.Errorf("rura_navigation evidence: %w", err)
	}
	ruraCTYs := map[int]bool{}
	for i, d := range p.Events.RuraNavigation.Destinations {
		if d.CTYRaw < 0 || d.CTYRaw > 255 || ruraCTYs[d.CTYRaw] ||
			d.NameRecordRaw < 0 || d.ArrivalOffset.X < -255 || d.ArrivalOffset.X > 255 ||
			d.ArrivalOffset.Y < -255 || d.ArrivalOffset.Y > 255 ||
			d.ShipWorldPosition.X < 0 || d.ShipWorldPosition.X > 1023 ||
			d.ShipWorldPosition.Y < 0 || d.ShipWorldPosition.Y > 1023 ||
			d.ShipWorldPosition.Layer < 0 || d.ShipWorldPosition.Layer > 1 {
			return fmt.Errorf("rura_navigation.destinations[%d]: invalid destination", i)
		}
		ruraCTYs[d.CTYRaw] = true
	}
	if p.Events.WorldEntranceVariants == nil {
		return errors.New("world_entrance_variants must be present")
	}
	if p.Events.SettlementFounderEvents == nil {
		return errors.New("settlement_founder_events must be present")
	}
	if p.Events.SettlementFounderFollowups == nil {
		return errors.New("settlement_founder_followups must be present")
	}
	if p.Events.ConditionalFacilityEvents == nil {
		return errors.New("conditional_facility_events must be present")
	}
	conditionalFacilityIDs := map[string]bool{}
	for i, event := range p.Events.ConditionalFacilityEvents {
		if event.ID == "" || event.Kind != "conditional_facility" ||
			!validScriptedNPC(event.NPC) || event.RequiredPlayerY < 0 ||
			event.RequiredPlayerY > 255 || event.FacilityK < 0 || event.FacilityK > 255 ||
			event.FallbackTextID == "" {
			return fmt.Errorf("conditional_facility_events[%d]: invalid event", i)
		}
		if conditionalFacilityIDs[event.ID] {
			return fmt.Errorf("duplicate conditional facility event id %q", event.ID)
		}
		if err := validateEvidence(event.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", event.ID, err)
		}
		conditionalFacilityIDs[event.ID] = true
	}
	entranceIDs := map[string]bool{}
	entranceTiles := map[[3]int]string{}
	for i, entrance := range p.Events.WorldEntranceVariants {
		key := [3]int{entrance.WorldTile.X, entrance.WorldTile.Y, entrance.Layer}
		if entrance.ID == "" || entranceIDs[entrance.ID] ||
			entrance.WorldTile.X < 0 || entrance.WorldTile.X > 1023 ||
			entrance.WorldTile.Y < 0 || entrance.WorldTile.Y > 1023 ||
			entrance.Layer < 0 || entrance.Layer > 1 ||
			entrance.DefaultCTYRaw < 0 || entrance.DefaultCTYRaw > 255 ||
			len(entrance.Branches) == 0 {
			return fmt.Errorf("world_entrance_variants[%d]: invalid entrance", i)
		}
		if previous := entranceTiles[key]; previous != "" {
			return fmt.Errorf("%s: world tile duplicates %s", entrance.ID, previous)
		}
		seenFlags := map[int]bool{}
		seenCTYs := map[int]bool{entrance.DefaultCTYRaw: true}
		for j, branch := range entrance.Branches {
			if branch.FlagRaw < 0 || branch.FlagRaw >= 512 || seenFlags[branch.FlagRaw] ||
				branch.CTYRaw < 0 || branch.CTYRaw > 255 || seenCTYs[branch.CTYRaw] {
				return fmt.Errorf("%s: invalid or duplicate branch %d", entrance.ID, j)
			}
			seenFlags[branch.FlagRaw] = true
			seenCTYs[branch.CTYRaw] = true
		}
		if err := validateEvidence(entrance.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", entrance.ID, err)
		}
		entranceIDs[entrance.ID] = true
		entranceTiles[key] = entrance.ID
	}
	pushRule := p.Events.NPCPushRule
	if pushRule.CtrlMask <= 0 || pushRule.CtrlMask > 255 ||
		pushRule.CtrlValue < 0 || pushRule.CtrlValue > 255 ||
		pushRule.CtrlValue & ^pushRule.CtrlMask != 0 ||
		pushRule.BlockedByEffectID != "temporary_invisibility" {
		return errors.New("npc_push_rule: invalid rule")
	}
	if err := validateEvidence(pushRule.Evidence); err != nil {
		return fmt.Errorf("npc_push_rule evidence: %w", err)
	}
	if p.Events.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Events.SchemaVersion)
	}
	if p.Events.BossSurrenderEvents == nil {
		return errors.New("boss_surrender_events must be present")
	}
	if p.Events.TemporaryRoleEvents == nil {
		return errors.New("temporary_role_events must be present")
	}
	if p.Events.TemporarySoloChallenges == nil {
		return errors.New("temporary_solo_challenges must be present")
	}
	if p.Events.QuestItemChainEvents == nil {
		return errors.New("quest_item_chain_events must be present")
	}
	if p.Events.ChoiceItemExchangeEvents == nil {
		return errors.New("choice_item_exchange_events must be present")
	}
	if p.Events.TrackedWorldObjects == nil {
		return errors.New("tracked_world_objects must be present")
	}
	if p.Events.CoordinateItemGateEvents == nil {
		return errors.New("coordinate_item_gate_events must be present")
	}
	if p.Events.TreasureEvents == nil {
		return errors.New("treasure_events must be present")
	}
	if p.Events.NPCItemRewardEvents == nil {
		return errors.New("npc_item_reward_events must be present")
	}
	if p.Events.ItemUseEffects == nil {
		return errors.New("item_use_effects must be present")
	}
	if p.Events.TrackingGuardEvents == nil {
		return errors.New("tracking_guard_events must be present")
	}
	if p.Events.PushPuzzleEvents == nil {
		return errors.New("push_puzzle_events must be present")
	}
	if p.Events.TwoStepFloorSwitchGates == nil {
		return errors.New("two_step_floor_switch_gates must be present")
	}
	if p.Events.StagedVehicleExchangeEvents == nil {
		return errors.New("staged_vehicle_exchange_events must be present")
	}
	if p.Events.GuidedPassageEvents == nil {
		return errors.New("guided_passage_events must be present")
	}
	if p.Events.HostageRescueEvents == nil {
		return errors.New("hostage_rescue_events must be present")
	}
	if p.Events.ReclassEvents == nil {
		return errors.New("reclass_events must be present")
	}
	if p.Events.StagedBossEvents == nil {
		return errors.New("staged_boss_events must be present")
	}
	if p.Events.StoryFlagRuntimeEvents == nil {
		return errors.New("story_flag_runtime_events must be present")
	}
	storyFlagEventIDs := map[string]bool{}
	for i, event := range p.Events.StoryFlagRuntimeEvents {
		if storyFlagEventIDs[event.ID] {
			return fmt.Errorf("story_flag_runtime_events[%d]: duplicate id %q", i, event.ID)
		}
		if err := validateStoryFlagRuntimeEvent(event); err != nil {
			return fmt.Errorf("story_flag_runtime_events[%d]: %w", i, err)
		}
		storyFlagEventIDs[event.ID] = true
	}
	p.settlementFounders = make(map[string]*SettlementFounderEvent,
		len(p.Events.SettlementFounderEvents))
	for i := range p.Events.SettlementFounderEvents {
		e := &p.Events.SettlementFounderEvents[i]
		if e.ID == "" || e.Kind != "settlement_founder" || !validScriptedNPC(e.NPC) ||
			e.RequiredClassRaw < 0 || e.RequiredClassRaw > 255 || !e.RequireAlive ||
			e.CompletionFlagRaw < 0 || e.CompletionFlagRaw >= 512 ||
			len(e.FounderVisibilityFlagsRaw) != 2 ||
			e.SharedStorageCapacity < 1 || e.SharedStorageCapacity > 255 {
			return fmt.Errorf("settlement_founder_events[%d]: invalid event", i)
		}
		if _, exists := p.settlementFounders[e.ID]; exists {
			return fmt.Errorf("duplicate settlement founder event id %q", e.ID)
		}
		for gender, flag := range e.FounderVisibilityFlagsRaw {
			if flag < 0 || flag >= 512 {
				return fmt.Errorf("%s: invalid founder visibility flag for gender%d", e.ID, gender)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.settlementFounders[e.ID] = e
	}
	p.settlementFounderFollowups = make(map[string]*SettlementFounderFollowupEvent,
		len(p.Events.SettlementFounderFollowups))
	for i := range p.Events.SettlementFounderFollowups {
		e := &p.Events.SettlementFounderFollowups[i]
		if e.ID == "" || e.Kind != "settlement_founder_followup" ||
			!validScriptedNPC(e.NPC) || len(e.DialogueTextIDs) == 0 {
			return fmt.Errorf("settlement_founder_followups[%d]: invalid event", i)
		}
		if _, exists := p.settlementFounderFollowups[e.ID]; exists {
			return fmt.Errorf("duplicate settlement founder followup id %q", e.ID)
		}
		flags := map[int]bool{}
		for _, flag := range e.RequiredFlagsSetRaw {
			if flag < 0 || flag >= 512 || flags[flag] {
				return fmt.Errorf("%s: invalid required set flag %#x", e.ID, flag)
			}
			flags[flag] = true
		}
		for _, flag := range e.RequiredFlagsClearRaw {
			if flag < 0 || flag >= 512 || flags[flag] {
				return fmt.Errorf("%s: invalid or contradictory required clear flag %#x", e.ID, flag)
			}
			flags[flag] = true
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.settlementFounderFollowups[e.ID] = e
	}
	p.bossEvents = make(map[string]*BossSurrenderEvent, len(p.Events.BossSurrenderEvents))
	for i := range p.Events.BossSurrenderEvents {
		e := &p.Events.BossSurrenderEvents[i]
		if e.ID == "" || e.Kind != "boss_surrender" {
			return fmt.Errorf("boss_surrender_events[%d]: id and kind=boss_surrender are required", i)
		}
		if _, exists := p.bossEvents[e.ID]; exists {
			return fmt.Errorf("duplicate boss surrender event id %q", e.ID)
		}
		if e.Trigger.Kind != "scene_tile_subid" || e.Trigger.CTYRaw < 0 ||
			e.Trigger.Section < 0 || e.Trigger.TileSubID < 0 || e.Trigger.TileSubID > 31 ||
			len(e.Trigger.Tiles) == 0 {
			return fmt.Errorf("%s: invalid scene_tile_subid trigger", e.ID)
		}
		if e.PresenceFlagRaw < 0 || e.PresenceFlagRaw > 255 ||
			e.ClearFlagRaw != e.PresenceFlagRaw {
			return fmt.Errorf("%s: invalid or inconsistent presence/clear flag", e.ID)
		}
		if e.Formation.Background.PageRaw < 0 || e.Formation.Background.PaletteBankRaw < 0 ||
			len(e.Formation.Groups) == 0 || e.Formation.RawBytesHex == "" {
			return fmt.Errorf("%s: incomplete formation", e.ID)
		}
		raw, err := hex.DecodeString(e.Formation.RawBytesHex)
		if err != nil {
			return fmt.Errorf("%s: invalid formation raw_bytes_hex: %w", e.ID, err)
		}
		if len(raw) != 3+len(e.Formation.Groups)*2 || int(raw[0]) != len(e.Formation.Groups) ||
			int(raw[1]) != e.Formation.Background.PageRaw ||
			int(raw[2]) != e.Formation.Background.PaletteBankRaw {
			return fmt.Errorf("%s: formation fields do not match raw_bytes_hex header", e.ID)
		}
		for j, group := range e.Formation.Groups {
			if group.MonsterRawID < 0 || group.MonsterRawID > 255 || group.Count < 1 {
				return fmt.Errorf("%s: invalid formation group %d", e.ID, j)
			}
			if int(raw[3+j*2]) != group.MonsterRawID || int(raw[4+j*2]) != group.Count {
				return fmt.Errorf("%s: formation group %d does not match raw_bytes_hex", e.ID, j)
			}
		}
		for j, gate := range e.TreasureGates {
			if gate.CTYRaw < 0 || gate.Section < 0 || gate.ItemRawID < 0 ||
				gate.ItemRawID > 255 || gate.WhileFlagSet < 0 || gate.WhileFlagSet > 255 {
				return fmt.Errorf("%s: invalid treasure gate %d", e.ID, j)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.bossEvents[e.ID] = e
	}
	p.roleEvents = make(map[string]*TemporaryRoleEvent, len(p.Events.TemporaryRoleEvents))
	for i := range p.Events.TemporaryRoleEvents {
		e := &p.Events.TemporaryRoleEvents[i]
		if e.ID == "" || e.Kind != "temporary_role" {
			return fmt.Errorf("temporary_role_events[%d]: id and kind=temporary_role are required", i)
		}
		if _, exists := p.roleEvents[e.ID]; exists {
			return fmt.Errorf("duplicate temporary role event id %q", e.ID)
		}
		if !validScriptedNPC(e.OfferNPC) || !validScriptedNPC(e.RestoreNPC) ||
			e.OfferNPC == e.RestoreNPC {
			return fmt.Errorf("%s: invalid or identical scripted NPC selectors", e.ID)
		}
		for name, value := range map[string]int{
			"pending_flag_raw":     e.PendingFlagRaw,
			"normal_role_flag_raw": e.NormalRoleFlagRaw,
			"active_role_flag_raw": e.ActiveRoleFlagRaw,
		} {
			if value < 0 || value >= 512 {
				return fmt.Errorf("%s: %s out of range", e.ID, name)
			}
		}
		if e.PendingFlagRaw == e.NormalRoleFlagRaw ||
			e.PendingFlagRaw == e.ActiveRoleFlagRaw ||
			e.NormalRoleFlagRaw == e.ActiveRoleFlagRaw {
			return fmt.Errorf("%s: role flags must be distinct", e.ID)
		}
		if e.RequiredItemRawID < 0 || e.RequiredItemRawID > 255 ||
			e.RoleSpriteEntries.Male < 0 || e.RoleSpriteEntries.Female < 0 {
			return fmt.Errorf("%s: invalid item or role sprite entry", e.ID)
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.roleEvents[e.ID] = e
	}
	p.soloChallenges = make(map[string]*TemporarySoloChallengeEvent, len(p.Events.TemporarySoloChallenges))
	for i := range p.Events.TemporarySoloChallenges {
		e := &p.Events.TemporarySoloChallenges[i]
		if e.ID == "" || e.Kind != "temporary_solo_challenge" {
			return fmt.Errorf("temporary_solo_challenges[%d]: id and kind=temporary_solo_challenge are required", i)
		}
		if _, exists := p.soloChallenges[e.ID]; exists {
			return fmt.Errorf("duplicate temporary solo challenge id %q", e.ID)
		}
		if !validScriptedNPC(e.EntryNPC) || !validScriptedNPC(e.ReturnNPC) || e.EntryNPC == e.ReturnNPC {
			return fmt.Errorf("%s: invalid or identical scripted NPC selectors", e.ID)
		}
		if e.CompletedFlagRaw < 0 || e.CompletedFlagRaw >= 512 || e.ModeMaskRaw <= 0 || e.ModeMaskRaw > 255 ||
			e.SoloWorldPosition.X < 0 || e.SoloWorldPosition.Y < 0 || e.SoloWorldPosition.Layer < 0 || e.SoloWorldPosition.Layer > 1 ||
			e.ReturnTownDestination.CTYRaw < 0 || e.ReturnTownDestination.CTYRaw > 255 ||
			e.ReturnTownDestination.Section < 0 || e.ReturnTownDestination.X < 0 || e.ReturnTownDestination.Y < 0 {
			return fmt.Errorf("%s: invalid flag, mode mask or destination", e.ID)
		}
		textIDs := []string{e.DialogueTextIDs.Prompt, e.DialogueTextIDs.Accept, e.DialogueTextIDs.Reject,
			e.DialogueTextIDs.Completed, e.DialogueTextIDs.Return, e.DialogueTextIDs.CaveBack,
			e.DialogueTextIDs.CaveEmpty, e.DialogueTextIDs.ChoiceYes, e.DialogueTextIDs.ChoiceNo}
		for _, id := range textIDs {
			if id == "" {
				return fmt.Errorf("%s: every dialogue text id is required", e.ID)
			}
		}
		if len(e.CaveMessages) == 0 {
			return fmt.Errorf("%s: cave_messages must be present", e.ID)
		}
		for _, message := range e.CaveMessages {
			if message.CTYRaw < 0 || message.CTYRaw > 255 || message.Section < 0 ||
				message.HandlerRaw < 0 || message.HandlerRaw > 255 || message.TextID == "" {
				return fmt.Errorf("%s: invalid cave message selector", e.ID)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.soloChallenges[e.ID] = e
	}
	p.questItemEvents = make(map[string]*QuestItemChainEvent, len(p.Events.QuestItemChainEvents))
	for i := range p.Events.QuestItemChainEvents {
		e := &p.Events.QuestItemChainEvents[i]
		if e.ID == "" || e.Kind != "quest_item_chain" {
			return fmt.Errorf("quest_item_chain_events[%d]: id and kind=quest_item_chain are required", i)
		}
		if _, exists := p.questItemEvents[e.ID]; exists {
			return fmt.Errorf("duplicate quest item chain event id %q", e.ID)
		}
		t := e.Treasure
		if !validTreasureSelector(t) {
			return fmt.Errorf("%s: invalid treasure selector", e.ID)
		}
		x := e.Exchange
		if !validScriptedNPC(x.NPC) ||
			x.RequiredItemRawID < 0 || x.RequiredItemRawID > 255 ||
			x.GrantedItemRawID < 0 || x.GrantedItemRawID > 255 ||
			x.RequiredItemRawID == x.GrantedItemRawID ||
			x.RequiredItemRawID != t.ItemRawID {
			return fmt.Errorf("%s: invalid or disconnected item exchange", e.ID)
		}
		u := e.Use
		if u.ItemRawID != x.GrantedItemRawID || u.LocationKind != "town" ||
			u.CTYRaw < 0 || u.CTYRaw > 255 || u.SuccessTextID == "" ||
			!u.ReloadScene || u.SetFlagsRaw == nil || u.ClearFlagsRaw == nil ||
			len(u.SetFlagsRaw) == 0 || len(u.ClearFlagsRaw) == 0 {
			return fmt.Errorf("%s: invalid or disconnected location item use", e.ID)
		}
		seen := map[int]string{}
		for kind, flags := range map[string][]int{
			"set_flags_raw": u.SetFlagsRaw, "clear_flags_raw": u.ClearFlagsRaw,
		} {
			for _, flag := range flags {
				if flag < 0 || flag >= 512 {
					return fmt.Errorf("%s: %s flag %d out of range", e.ID, kind, flag)
				}
				if previous := seen[flag]; previous != "" {
					return fmt.Errorf("%s: flag %d appears in both %s and %s",
						e.ID, flag, previous, kind)
				}
				seen[flag] = kind
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.questItemEvents[e.ID] = e
	}
	p.choiceItemExchangeEvents = make(map[string]*ChoiceItemExchangeEvent, len(p.Events.ChoiceItemExchangeEvents))
	choiceExchangeSelectors := map[[4]int]string{}
	for i := range p.Events.ChoiceItemExchangeEvents {
		e := &p.Events.ChoiceItemExchangeEvents[i]
		selector := [4]int{e.NPC.CTYRaw, e.NPC.Section, e.NPC.Tile.X, e.NPC.Tile.Y}
		if e.ID == "" || e.Kind != "choice_item_exchange" || !validScriptedNPC(e.NPC) ||
			e.AvailableFlagRaw < 0 || e.AvailableFlagRaw >= 512 ||
			e.RequiredItemRawID < 0 || e.RequiredItemRawID > 255 ||
			e.GrantedItemRawID < 0 || e.GrantedItemRawID > 255 ||
			e.RequiredItemRawID == e.GrantedItemRawID ||
			e.ActivateWorldObject.ObjectID == "" || e.ActivateWorldObject.Position.X < 0 ||
			e.ActivateWorldObject.Position.Y < 0 ||
			e.ActivateWorldObject.Position.Layer < 0 || e.ActivateWorldObject.Position.Layer > 1 ||
			e.SetWorldStateMaskRaw <= 0 || e.SetWorldStateMaskRaw > 0xffff ||
			e.ClearStoryFlagsRaw == nil || len(e.ClearStoryFlagsRaw) == 0 {
			return fmt.Errorf("choice_item_exchange_events[%d]: invalid event", i)
		}
		if _, exists := p.choiceItemExchangeEvents[e.ID]; exists {
			return fmt.Errorf("duplicate choice item exchange event id %q", e.ID)
		}
		if previous := choiceExchangeSelectors[selector]; previous != "" {
			return fmt.Errorf("%s: NPC selector duplicates %s", e.ID, previous)
		}
		for _, flag := range e.ClearStoryFlagsRaw {
			if flag < 0 || flag >= 512 {
				return fmt.Errorf("%s: clear story flag %d out of range", e.ID, flag)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.choiceItemExchangeEvents[e.ID] = e
		choiceExchangeSelectors[selector] = e.ID
	}
	p.coordinateItemGateEvents = make(map[string]*CoordinateItemGateEvent, len(p.Events.CoordinateItemGateEvents))
	coordinateSelectors := map[VehicleWorldPosition]string{}
	for i := range p.Events.CoordinateItemGateEvents {
		e := &p.Events.CoordinateItemGateEvents[i]
		if e.ID == "" || e.Kind != "coordinate_item_gate" ||
			e.Coordinate.X < 0 || e.Coordinate.Y < 0 || e.Coordinate.Layer < 0 || e.Coordinate.Layer > 1 ||
			e.ActiveStoryFlagRaw < 0 || e.ActiveStoryFlagRaw >= 512 ||
			e.RequiredItemRawID < 0 || e.RequiredItemRawID > 255 ||
			e.ClearStoryFlagRaw != e.ActiveStoryFlagRaw || e.ConsumeRequiredItem ||
			e.FailureDirection < 0 || e.FailureDirection > 3 || e.FailureSteps <= 0 ||
			e.FailureDayNightSteps < 0 || e.DialogueTextIDs.Approach == "" ||
			e.DialogueTextIDs.Success == "" {
			return fmt.Errorf("coordinate_item_gate_events[%d]: invalid event", i)
		}
		if p.coordinateItemGateEvents[e.ID] != nil || coordinateSelectors[e.Coordinate] != "" {
			return fmt.Errorf("coordinate_item_gate_events[%d]: duplicate id or coordinate", i)
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.coordinateItemGateEvents[e.ID] = e
		coordinateSelectors[e.Coordinate] = e.ID
	}
	p.trackedWorldObjects = make(map[string]*TrackedWorldObject, len(p.Events.TrackedWorldObjects))
	trackerItems := map[int]string{}
	for i := range p.Events.TrackedWorldObjects {
		e := &p.Events.TrackedWorldObjects[i]
		texts := e.TrackerDialogueTextIDs
		if e.ID == "" || e.Kind != "tracked_world_entrance" ||
			e.ActiveWorldStateMask <= 0 || e.ActiveWorldStateMask > 0xffff ||
			e.Layer < 0 || e.Layer > 1 || e.EntranceCTYRaw < 0 || e.EntranceCTYRaw > 255 ||
			e.EntranceSection < 0 || e.EntryVehicleMode != "disembark_ship_if_aboard" ||
			len(e.TileRawByRNGLow2) != 4 ||
			e.TrackerItemRawID < 0 || e.TrackerItemRawID > 255 ||
			texts.Preamble == "" || texts.West == "" || texts.East == "" ||
			texts.North == "" || texts.South == "" {
			return fmt.Errorf("tracked_world_objects[%d]: invalid object", i)
		}
		r := e.Relocation
		if r.WindowWidth != 80 || r.WindowHeight != 80 || r.OriginOffset != 40 ||
			r.HorizontalKeepMin != 10 || r.HorizontalKeepMax != 69 ||
			r.VerticalKeepMin != 8 || r.VerticalKeepMax != 71 || len(r.Candidates) != 18 {
			return fmt.Errorf("%s: invalid relocation window or candidate count", e.ID)
		}
		for _, candidate := range r.Candidates {
			if candidate.X < 0 || candidate.Y < 0 || candidate.Layer != e.Layer {
				return fmt.Errorf("%s: invalid relocation candidate", e.ID)
			}
		}
		if _, exists := p.trackedWorldObjects[e.ID]; exists || trackerItems[e.TrackerItemRawID] != "" {
			return fmt.Errorf("tracked_world_objects[%d]: duplicate id or tracker item", i)
		}
		for _, tile := range e.TileRawByRNGLow2 {
			if tile < 0 || tile > 255 {
				return fmt.Errorf("%s: tile_raw_by_rng_low2 out of range", e.ID)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.trackedWorldObjects[e.ID] = e
		trackerItems[e.TrackerItemRawID] = e.ID
	}
	for i := range p.Events.ChoiceItemExchangeEvents {
		e := &p.Events.ChoiceItemExchangeEvents[i]
		obj := p.trackedWorldObjects[e.ActivateWorldObject.ObjectID]
		if obj == nil || obj.Layer != e.ActivateWorldObject.Position.Layer ||
			obj.ActiveWorldStateMask != e.SetWorldStateMaskRaw {
			return fmt.Errorf("%s: disconnected tracked world object activation", e.ID)
		}
	}
	p.treasureEvents = make(map[string]*TreasureEvent, len(p.Events.TreasureEvents))
	treasureSelectors := map[[3]int]string{}
	for i := range p.Events.TreasureEvents {
		e := &p.Events.TreasureEvents[i]
		if e.ID == "" || e.Kind != "treasure" {
			return fmt.Errorf("treasure_events[%d]: id and kind=treasure are required", i)
		}
		if _, exists := p.treasureEvents[e.ID]; exists {
			return fmt.Errorf("duplicate treasure event id %q", e.ID)
		}
		if !validTreasureSelector(e.Treasure) {
			return fmt.Errorf("%s: invalid treasure selector", e.ID)
		}
		key := [3]int{e.Treasure.CTYRaw, e.Treasure.Section, e.Treasure.TileSubID}
		if previous := treasureSelectors[key]; previous != "" {
			return fmt.Errorf("%s: treasure selector duplicates %s", e.ID, previous)
		}
		treasureSelectors[key] = e.ID
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.treasureEvents[e.ID] = e
	}
	npcRewardIDs := map[string]bool{}
	npcRewardSelectors := map[[4]int]string{}
	for i := range p.Events.NPCItemRewardEvents {
		e := &p.Events.NPCItemRewardEvents[i]
		selector := [4]int{e.NPC.CTYRaw, e.NPC.Section, e.NPC.Tile.X, e.NPC.Tile.Y}
		if e.ID == "" || e.Kind != "npc_item_reward" || npcRewardIDs[e.ID] ||
			!validScriptedNPC(e.NPC) || e.PresentFlagRaw < 0 || e.PresentFlagRaw >= 512 ||
			e.GrantedItemRaw < 0 || e.GrantedItemRaw > 255 {
			return fmt.Errorf("npc_item_reward_events[%d]: invalid event", i)
		}
		if previous := npcRewardSelectors[selector]; previous != "" {
			return fmt.Errorf("%s: NPC selector duplicates %s", e.ID, previous)
		}
		if e.RequiredItemRawID != nil && (*e.RequiredItemRawID < 0 || *e.RequiredItemRawID > 255) {
			return fmt.Errorf("%s: required_item_raw_id out of range", e.ID)
		}
		if e.RequiredItemRawID == nil && e.ReplaceRequired {
			return fmt.Errorf("%s: replace_required requires required_item_raw_id", e.ID)
		}
		for _, flag := range e.ClearStoryFlagsRaw {
			if flag < 0 || flag >= 512 {
				return fmt.Errorf("%s: clear story flag out of range", e.ID)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		npcRewardIDs[e.ID] = true
		npcRewardSelectors[selector] = e.ID
	}
	p.itemUseEffects = make(map[int]*ItemUseEffect, len(p.Events.ItemUseEffects))
	itemUseIDs := make(map[string]bool, len(p.Events.ItemUseEffects))
	for i := range p.Events.ItemUseEffects {
		e := &p.Events.ItemUseEffects[i]
		if e.ID == "" || e.Kind != "item_use" {
			return fmt.Errorf("item_use_effects[%d]: id and kind=item_use are required", i)
		}
		if itemUseIDs[e.ID] {
			return fmt.Errorf("duplicate item use effect id %q", e.ID)
		}
		if e.ItemRawID < 0 || e.ItemRawID > 255 {
			return fmt.Errorf("%s: item_raw_id out of range", e.ID)
		}
		if previous := p.itemUseEffects[e.ItemRawID]; previous != nil {
			return fmt.Errorf("%s: item_raw_id duplicates %s", e.ID, previous.ID)
		}
		switch e.EffectID {
		case "force_day_night_phase":
			if e.LocationKind != "overworld" ||
				e.DayNightPhase < 0 || e.DayNightPhase > 3 ||
				!e.ResetDayNightSteps || e.StepCount != 0 {
				return fmt.Errorf("%s: invalid force_day_night_phase configuration", e.ID)
			}
		case "temporary_invisibility":
			if e.LocationKind != "any" || e.DayNightPhase != 0 ||
				e.ResetDayNightSteps || e.StepCount <= 0 || !e.Consume {
				return fmt.Errorf("%s: invalid temporary_invisibility configuration", e.ID)
			}
		case "reveal_world_map_patch":
			patch := e.MapPatch
			if e.LocationKind != "overworld" || e.RequiredLayer < 0 ||
				(e.RequiredVehicle != "" && e.RequiredVehicle != "ship") || e.UseTile == nil ||
				e.UseTile.X < 0 || e.UseTile.Y < 0 ||
				e.RequiredStoryFlagRaw < -1 || e.RequiredStoryFlagRaw >= 512 ||
				e.RequiredStoryFlagSet || e.SetWorldStateMask <= 0 ||
				e.SetWorldStateMask > 0xffff || e.Consume ||
				patch == nil || patch.Layer != e.RequiredLayer ||
				patch.Origin.X < 0 || patch.Origin.Y < 0 ||
				patch.Width <= 0 || patch.Height <= 0 ||
				len(patch.TilesRaw) != patch.Width*patch.Height ||
				(e.DirectMapPatch != nil && (e.DirectMapPatch.Layer != e.RequiredLayer ||
					e.DirectMapPatch.Origin.X < 0 || e.DirectMapPatch.Origin.Y < 0 ||
					e.DirectMapPatch.Width <= 0 || e.DirectMapPatch.Height <= 0 ||
					len(e.DirectMapPatch.TilesRaw) != e.DirectMapPatch.Width*e.DirectMapPatch.Height)) ||
				e.AnimationPaletteModeRaw < -1 || e.AnimationPaletteModeRaw > 255 ||
				e.AnimationCycles <= 0 {
				return fmt.Errorf("%s: invalid reveal_world_map_patch configuration", e.ID)
			}
			for _, tile := range patch.TilesRaw {
				if tile < 0 || tile > 255 {
					return fmt.Errorf("%s: map patch tile out of range", e.ID)
				}
			}
			if e.DirectMapPatch != nil {
				for _, tile := range e.DirectMapPatch.TilesRaw {
					if tile < 0 || tile > 255 {
						return fmt.Errorf("%s: direct map patch tile out of range", e.ID)
					}
				}
			}
			foundClear := e.RequiredStoryFlagRaw < 0
			for _, flag := range e.ClearStoryFlagsRaw {
				if flag < 0 || flag >= 512 {
					return fmt.Errorf("%s: clear story flag out of range", e.ID)
				}
				foundClear = foundClear || flag == e.RequiredStoryFlagRaw
			}
			if !foundClear {
				return fmt.Errorf("%s: required story flag must be in clear_story_flags_raw", e.ID)
			}
		case "open_facing_locked_door":
			if e.LocationKind != "town" || e.DoorKeyTier < 1 || e.DoorKeyTier > 3 ||
				e.Consume || e.DayNightPhase != 0 || e.ResetDayNightSteps ||
				e.StepCount != 0 || e.RequiredVehicle != "" || e.UseTile != nil ||
				e.MapPatch != nil || e.DirectMapPatch != nil {
				return fmt.Errorf("%s: invalid open_facing_locked_door configuration", e.ID)
			}
		case "grant_quest_item":
			if e.LocationKind != "town" || e.RequiredCTYRaw < 0 || e.RequiredCTYRaw > 255 ||
				e.RequiredSection < 0 || e.GrantItemRawID < 0 || e.GrantItemRawID > 255 ||
				e.Consume || e.RequiredLayer != 0 || e.RequiredVehicle != "" ||
				e.UseTile != nil || e.MapPatch != nil || e.DirectMapPatch != nil ||
				e.SetWorldStateMask != 0 || len(e.ClearStoryFlagsRaw) != 0 {
				return fmt.Errorf("%s: invalid grant_quest_item configuration", e.ID)
			}
			for _, flag := range e.SetStoryFlagsRaw {
				if flag < 0 || flag >= 512 {
					return fmt.Errorf("%s: set story flag out of range", e.ID)
				}
			}
		default:
			return fmt.Errorf("%s: unknown item use effect %q", e.ID, e.EffectID)
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		itemUseIDs[e.ID] = true
		p.itemUseEffects[e.ItemRawID] = e
	}
	trackingIDs := make(map[string]bool, len(p.Events.TrackingGuardEvents))
	for i := range p.Events.TrackingGuardEvents {
		e := &p.Events.TrackingGuardEvents[i]
		if e.ID == "" || e.Kind != "tracking_guard" || trackingIDs[e.ID] ||
			e.CTYRaw < 0 || e.CTYRaw > 255 || e.Section < 0 ||
			!validScriptedNPC(e.Guard) || e.Guard.CTYRaw != e.CTYRaw ||
			e.Guard.Section != e.Section || e.TrackAxis != "horizontal" ||
			e.BypassEffectID != "temporary_invisibility" ||
			len(e.TriggerTiles) == 0 {
			return fmt.Errorf("tracking_guard_events[%d]: invalid event", i)
		}
		seenTiles := map[[2]int]bool{}
		for _, trigger := range e.TriggerTiles {
			key := [2]int{trigger.Tile.X, trigger.Tile.Y}
			if trigger.Tile.X < 0 || trigger.Tile.Y < 0 || trigger.TileSubID < 1 ||
				trigger.TileSubID > 31 || trigger.HandlerRaw < 0 || trigger.HandlerRaw > 255 ||
				seenTiles[key] {
				return fmt.Errorf("%s: invalid or duplicate trigger tile", e.ID)
			}
			seenTiles[key] = true
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		trackingIDs[e.ID] = true
	}
	pushPuzzleIDs := make(map[string]bool, len(p.Events.PushPuzzleEvents))
	for i := range p.Events.PushPuzzleEvents {
		e := &p.Events.PushPuzzleEvents[i]
		if e.ID == "" || e.Kind != "push_puzzle" || pushPuzzleIDs[e.ID] ||
			e.CTYRaw < 0 || e.CTYRaw > 255 || e.Section < 0 ||
			len(e.TriggerTiles) == 0 || len(e.NPCs) == 0 ||
			e.CompletionAxis != "y" || e.CompletionValue < 0 ||
			e.ClearStoryFlagRaw <= 0 || e.ClearStoryFlagRaw > 511 {
			return fmt.Errorf("push_puzzle_events[%d]: invalid event", i)
		}
		seenTiles := map[[2]int]bool{}
		for _, trigger := range e.TriggerTiles {
			key := [2]int{trigger.Tile.X, trigger.Tile.Y}
			if trigger.Tile.X < 0 || trigger.Tile.Y < 0 || trigger.TileSubID < 1 ||
				trigger.TileSubID > 31 || trigger.HandlerRaw < 0 || trigger.HandlerRaw > 255 ||
				seenTiles[key] {
				return fmt.Errorf("%s: invalid or duplicate trigger tile", e.ID)
			}
			seenTiles[key] = true
		}
		seenSlots := map[int]bool{}
		for _, npc := range e.NPCs {
			if npc.Slot < 0 || seenSlots[npc.Slot] || npc.InitialTile.X < 0 || npc.InitialTile.Y < 0 ||
				npc.CtrlRaw < 0 || npc.CtrlRaw > 255 ||
				npc.CtrlRaw&pushRule.CtrlMask != pushRule.CtrlValue {
				return fmt.Errorf("%s: invalid or duplicate puzzle NPC", e.ID)
			}
			seenSlots[npc.Slot] = true
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		pushPuzzleIDs[e.ID] = true
	}
	p.sequenceGates = make(map[string]*TwoStepFloorSwitchGate, len(p.Events.TwoStepFloorSwitchGates))
	for i := range p.Events.TwoStepFloorSwitchGates {
		e := &p.Events.TwoStepFloorSwitchGates[i]
		if e.ID == "" || e.Kind != "two_step_floor_switch_gate" {
			return fmt.Errorf("two_step_floor_switch_gates[%d]: id and kind=two_step_floor_switch_gate are required", i)
		}
		if _, exists := p.sequenceGates[e.ID]; exists {
			return fmt.Errorf("duplicate two-step floor switch gate id %q", e.ID)
		}
		if e.CTYRaw < 0 || e.CTYRaw > 255 || e.Section < 0 ||
			len(e.Switches) < 2 || e.CompletionTileSubID < 0 || e.CompletionTileSubID > 31 ||
			e.ClearFlagRaw < 0 || e.ClearFlagRaw >= 512 ||
			e.TrapTransitionSubIDRaw < 0 || e.TrapTransitionSubIDRaw > 31 ||
			e.SuccessSFXRaw < 0 || e.SuccessSFXRaw > 255 {
			return fmt.Errorf("%s: invalid gate header", e.ID)
		}
		seenTiles, handlersBySubID, hasCompletion := map[TileCoordinate]bool{}, map[int]int{}, false
		for j, sw := range e.Switches {
			if sw.Tile.X < 0 || sw.Tile.Y < 0 || sw.TileSubID < 0 || sw.TileSubID > 31 ||
				sw.HandlerRaw < 0 || sw.HandlerRaw > 255 || seenTiles[sw.Tile] {
				return fmt.Errorf("%s: invalid or duplicate switch %d", e.ID, j)
			}
			if handler, exists := handlersBySubID[sw.TileSubID]; exists && handler != sw.HandlerRaw {
				return fmt.Errorf("%s: tile_subid %d maps to inconsistent handlers", e.ID, sw.TileSubID)
			}
			seenTiles[sw.Tile] = true
			handlersBySubID[sw.TileSubID] = sw.HandlerRaw
			hasCompletion = hasCompletion || sw.TileSubID == e.CompletionTileSubID
		}
		if !hasCompletion {
			return fmt.Errorf("%s: completion_tile_subid is not a configured switch", e.ID)
		}
		d := e.TrapDestination
		if d.CTYRaw < 0 || d.CTYRaw > 255 || d.Section < 0 || d.X < 0 || d.Y < 0 {
			return fmt.Errorf("%s: invalid trap destination", e.ID)
		}
		t := e.UnlockedTreasure
		if t.CTYRaw != e.CTYRaw || t.Section != e.Section ||
			t.TileSubID < 0 || t.TileSubID > 31 ||
			(t.EventTypeRaw != 1 && t.EventTypeRaw != 3) ||
			t.ItemRawID < 0 || t.ItemRawID > 255 ||
			t.PresentFlag < 0 || t.PresentFlag >= 512 {
			return fmt.Errorf("%s: invalid or disconnected unlocked treasure", e.ID)
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.sequenceGates[e.ID] = e
	}
	p.vehicleExchanges = make(map[string]*StagedVehicleExchangeEvent,
		len(p.Events.StagedVehicleExchangeEvents))
	for i := range p.Events.StagedVehicleExchangeEvents {
		e := &p.Events.StagedVehicleExchangeEvents[i]
		if e.ID == "" || e.Kind != "staged_vehicle_exchange" {
			return fmt.Errorf("staged_vehicle_exchange_events[%d]: id and kind=staged_vehicle_exchange are required", i)
		}
		if _, exists := p.vehicleExchanges[e.ID]; exists {
			return fmt.Errorf("duplicate staged vehicle exchange event id %q", e.ID)
		}
		if !validScriptedNPC(e.NPC) ||
			e.QuestPresentFlagRaw < 0 || e.QuestPresentFlagRaw >= 512 ||
			e.ExchangeAvailableFlagRaw < 0 || e.ExchangeAvailableFlagRaw >= 512 ||
			e.QuestPresentFlagRaw == e.ExchangeAvailableFlagRaw ||
			e.GrantedItemRawID < 0 || e.GrantedItemRawID > 255 ||
			e.RequiredItemRawID < 0 || e.RequiredItemRawID > 255 ||
			e.GrantedItemRawID == e.RequiredItemRawID {
			return fmt.Errorf("%s: invalid selector, flags or item IDs", e.ID)
		}
		v := e.Vehicle
		if v.Kind != "ship" || v.WorldPosition.X < 0 || v.WorldPosition.Y < 0 ||
			v.WorldPosition.Layer != 0 ||
			len(v.ClearFlagsRaw) == 0 {
			return fmt.Errorf("%s: invalid vehicle grant", e.ID)
		}
		seenFlags := map[int]bool{}
		for _, flag := range v.ClearFlagsRaw {
			if flag < 0 || flag >= 512 || seenFlags[flag] {
				return fmt.Errorf("%s: invalid or duplicate vehicle clear flag %d", e.ID, flag)
			}
			seenFlags[flag] = true
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.vehicleExchanges[e.ID] = e
	}
	p.guidedPassages = make(map[string]*GuidedPassageEvent,
		len(p.Events.GuidedPassageEvents))
	for i := range p.Events.GuidedPassageEvents {
		e := &p.Events.GuidedPassageEvents[i]
		if e.ID == "" || e.Kind != "guided_passage" {
			return fmt.Errorf("guided_passage_events[%d]: id and kind=guided_passage are required", i)
		}
		if _, exists := p.guidedPassages[e.ID]; exists {
			return fmt.Errorf("duplicate guided passage event id %q", e.ID)
		}
		if !validScriptedNPC(e.GuideNPC) || !validScriptedNPC(e.CompletedNPC) ||
			e.GuideNPC.CTYRaw != e.CompletedNPC.CTYRaw ||
			e.GuideNPC.Section != e.CompletedNPC.Section ||
			e.GuideNPC.HandlerRaw != e.CompletedNPC.HandlerRaw ||
			e.GuideNPC == e.CompletedNPC ||
			e.SceneHandlerRaw < 0 || e.SceneHandlerRaw > 255 ||
			e.RequiredItemRawID < 0 || e.RequiredItemRawID > 255 ||
			e.StepFrames < 1 || e.StepFrames > 60 {
			return fmt.Errorf("%s: invalid selectors, handler, item or step_frames", e.ID)
		}
		flags := []int{
			e.GuidePresentFlagRaw,
			e.AnimationPendingFlagRaw,
			e.CompletedFlagRaw,
		}
		seenFlags := map[int]bool{}
		for _, flag := range flags {
			if flag < 0 || flag >= 512 || seenFlags[flag] {
				return fmt.Errorf("%s: invalid or duplicate passage flag %d", e.ID, flag)
			}
			seenFlags[flag] = true
		}
		interactionDX := e.InteractionTile.X - e.GuideNPC.Tile.X
		if interactionDX < 0 {
			interactionDX = -interactionDX
		}
		interactionDY := e.InteractionTile.Y - e.GuideNPC.Tile.Y
		if interactionDY < 0 {
			interactionDY = -interactionDY
		}
		if e.InteractionTile.X < 0 || e.InteractionTile.Y < 0 ||
			interactionDX+interactionDY != 1 ||
			e.CompletionTrigger.Kind != "scene_tile_subid" ||
			e.CompletionTrigger.CTYRaw != e.GuideNPC.CTYRaw ||
			e.CompletionTrigger.Section != e.GuideNPC.Section ||
			e.CompletionTrigger.TileSubID < 1 || e.CompletionTrigger.TileSubID > 31 ||
			len(e.CompletionTrigger.Tiles) == 0 ||
			len(e.GuidePath) < 2 || len(e.GuideReturnPath) < 2 ||
			e.GuidePath[0] != e.GuideNPC.Tile ||
			e.GuideReturnPath[0] != e.GuidePath[len(e.GuidePath)-1] {
			return fmt.Errorf("%s: disconnected interaction, trigger or guide path", e.ID)
		}
		validatePath := func(name string, points []TileCoordinate) error {
			for j, point := range points {
				if point.X < 0 || point.Y < 0 {
					return fmt.Errorf("%s: %s waypoint %d out of range", e.ID, name, j)
				}
				if j == 0 {
					continue
				}
				prev := points[j-1]
				if prev == point || prev.X != point.X && prev.Y != point.Y {
					return fmt.Errorf("%s: %s waypoint %d is not a cardinal segment", e.ID, name, j)
				}
			}
			return nil
		}
		if err := validatePath("guide_path", e.GuidePath); err != nil {
			return err
		}
		if err := validatePath("guide_return_path", e.GuideReturnPath); err != nil {
			return err
		}
		validFacing := func(facing string) bool {
			switch facing {
			case "down", "up", "left", "right":
				return true
			default:
				return false
			}
		}
		if !validFacing(e.GuideEndFacing) || !validFacing(e.GuideReturnEndFacing) {
			return fmt.Errorf("%s: guide final facings must be down/up/left/right", e.ID)
		}
		for j, tile := range e.CompletionTrigger.Tiles {
			if tile.X < 0 || tile.Y < 0 {
				return fmt.Errorf("%s: completion trigger tile %d out of range", e.ID, j)
			}
		}
		if e.GuideMovementRawHex == "" || len(e.GuideReturnRawHex) == 0 {
			return fmt.Errorf("%s: movement raw anchors are required", e.ID)
		}
		if _, err := hex.DecodeString(e.GuideMovementRawHex); err != nil {
			return fmt.Errorf("%s: invalid guide_movement_raw_hex: %w", e.ID, err)
		}
		for j, rawHex := range e.GuideReturnRawHex {
			if rawHex == "" {
				return fmt.Errorf("%s: empty guide_return_raw_hex[%d]", e.ID, j)
			}
			if _, err := hex.DecodeString(rawHex); err != nil {
				return fmt.Errorf("%s: invalid guide_return_raw_hex[%d]: %w", e.ID, j, err)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.guidedPassages[e.ID] = e
	}
	p.hostageRescues = make(map[string]*HostageRescueEvent,
		len(p.Events.HostageRescueEvents))
	for i := range p.Events.HostageRescueEvents {
		e := &p.Events.HostageRescueEvents[i]
		if e.ID == "" || e.Kind != "hostage_rescue" {
			return fmt.Errorf("hostage_rescue_events[%d]: id and kind=hostage_rescue are required", i)
		}
		if _, exists := p.hostageRescues[e.ID]; exists {
			return fmt.Errorf("duplicate hostage rescue event id %q", e.ID)
		}
		if len(e.GuardNPCs) == 0 || len(e.BossNPCs) == 0 ||
			!validScriptedNPC(e.RewardNPC) || e.StepFrames < 1 || e.StepFrames > 60 ||
			e.RewardItemRawID < 0 || e.RewardItemRawID > 255 {
			return fmt.Errorf("%s: invalid NPC selectors, step_frames or reward item", e.ID)
		}
		for _, selectors := range [][]ScriptedNPCSelector{e.GuardNPCs, e.BossNPCs} {
			for _, selector := range selectors {
				if !validScriptedNPC(selector) {
					return fmt.Errorf("%s: invalid scripted NPC selector", e.ID)
				}
			}
		}
		stageFlags := []int{
			e.GuardPresentFlagRaw, e.ApproachPendingFlagRaw,
			e.CaptivePresentFlagRaw, e.BossPresentFlagRaw,
			e.RewardAvailableFlagRaw,
		}
		if len(e.CompletionSetFlagsRaw) == 0 || len(e.CompletionClearFlagsRaw) == 0 {
			return fmt.Errorf("%s: completion flag transactions must be present", e.ID)
		}
		seenFlags := map[int]bool{}
		for _, flag := range stageFlags {
			if flag < 0 || flag >= 512 || seenFlags[flag] {
				return fmt.Errorf("%s: invalid or duplicate hostage rescue flag %d", e.ID, flag)
			}
			seenFlags[flag] = true
		}
		seenSetFlags := map[int]bool{}
		for _, flag := range e.CompletionSetFlagsRaw {
			if flag < 0 || flag >= 512 || seenFlags[flag] || seenSetFlags[flag] {
				return fmt.Errorf("%s: invalid or duplicate hostage rescue flag %d", e.ID, flag)
			}
			seenSetFlags[flag] = true
		}
		seenClearFlags := map[int]bool{}
		for _, flag := range e.CompletionClearFlagsRaw {
			if flag < 0 || flag >= 512 || seenSetFlags[flag] || seenClearFlags[flag] {
				return fmt.Errorf("%s: invalid or duplicate hostage rescue flag %d", e.ID, flag)
			}
			seenClearFlags[flag] = true
		}
		validateTrigger := func(name string, trigger SceneTileTrigger) error {
			if trigger.Kind != "scene_tile_subid" || trigger.CTYRaw < 0 ||
				trigger.Section < 0 || trigger.TileSubID < 1 ||
				trigger.TileSubID > 31 || len(trigger.Tiles) == 0 {
				return fmt.Errorf("%s: invalid %s", e.ID, name)
			}
			seenTiles := map[TileCoordinate]bool{}
			for j, tile := range trigger.Tiles {
				if tile.X < 0 || tile.Y < 0 || seenTiles[tile] {
					return fmt.Errorf("%s: invalid or duplicate %s tile %d", e.ID, name, j)
				}
				seenTiles[tile] = true
			}
			return nil
		}
		if err := validateTrigger("approach_trigger", e.ApproachTrigger); err != nil {
			return err
		}
		if err := validateTrigger("switch_trigger", e.SwitchTrigger); err != nil {
			return err
		}
		if e.ApproachTrigger.CTYRaw != e.SwitchTrigger.CTYRaw ||
			e.ApproachTrigger.Section != e.SwitchTrigger.Section {
			return fmt.Errorf("%s: rescue triggers must share one scene", e.ID)
		}
		for _, selectors := range [][]ScriptedNPCSelector{e.GuardNPCs, e.BossNPCs} {
			for _, selector := range selectors {
				if selector.CTYRaw != e.ApproachTrigger.CTYRaw ||
					selector.Section != e.ApproachTrigger.Section {
					return fmt.Errorf("%s: guard and boss selectors must share trigger scene", e.ID)
				}
			}
		}
		validateFormation := func(name string, formation BattleFormation) error {
			raw, err := hex.DecodeString(formation.RawBytesHex)
			if err != nil || len(formation.Groups) == 0 ||
				len(raw) != 3+len(formation.Groups)*2 ||
				int(raw[0]) != len(formation.Groups) ||
				int(raw[1]) != formation.Background.PageRaw ||
				int(raw[2]) != formation.Background.PaletteBankRaw {
				return fmt.Errorf("%s: invalid %s", e.ID, name)
			}
			for j, group := range formation.Groups {
				if group.MonsterRawID < 0 || group.MonsterRawID > 255 ||
					group.Count < 1 || int(raw[3+j*2]) != group.MonsterRawID ||
					int(raw[4+j*2]) != group.Count {
					return fmt.Errorf("%s: invalid %s group %d", e.ID, name, j)
				}
			}
			return nil
		}
		if err := validateFormation("guard_formation", e.GuardFormation); err != nil {
			return err
		}
		if err := validateFormation("boss_formation", e.BossFormation); err != nil {
			return err
		}
		validateMovement := func(name string, movement ScriptedMovement) error {
			if !validScriptedNPC(movement.NPC) || len(movement.Path) < 2 ||
				movement.RawBytesHex == "" ||
				movement.NPC.CTYRaw != e.ApproachTrigger.CTYRaw ||
				movement.NPC.Section != e.ApproachTrigger.Section {
				return fmt.Errorf("%s: invalid %s", e.ID, name)
			}
			if _, err := hex.DecodeString(movement.RawBytesHex); err != nil {
				return fmt.Errorf("%s: invalid %s raw bytes", e.ID, name)
			}
			for j, point := range movement.Path {
				if point.X < 0 || point.Y < 0 {
					return fmt.Errorf("%s: %s waypoint %d out of range", e.ID, name, j)
				}
				if j == 0 {
					continue
				}
				prev := movement.Path[j-1]
				if prev == point || prev.X != point.X && prev.Y != point.Y {
					return fmt.Errorf("%s: %s waypoint %d is not a cardinal segment", e.ID, name, j)
				}
			}
			if _, ok := map[string]bool{"down": true, "up": true, "left": true, "right": true}[movement.EndFacing]; !ok {
				return fmt.Errorf("%s: invalid %s end facing", e.ID, name)
			}
			return nil
		}
		if err := validateMovement("approach_movement", e.ApproachMovement); err != nil {
			return err
		}
		if len(e.CaptiveMovements) == 0 {
			return fmt.Errorf("%s: captive_movements must be present", e.ID)
		}
		for j, movement := range e.CaptiveMovements {
			if err := validateMovement(fmt.Sprintf("captive_movements[%d]", j), movement); err != nil {
				return err
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.hostageRescues[e.ID] = e
	}
	p.reclassEvents = make(map[string]*ReclassEvent, len(p.Events.ReclassEvents))
	for i := range p.Events.ReclassEvents {
		e := &p.Events.ReclassEvents[i]
		if e.ID == "" || e.Kind != "reclass" {
			return fmt.Errorf("reclass_events[%d]: id and kind=reclass are required", i)
		}
		if _, exists := p.reclassEvents[e.ID]; exists {
			return fmt.Errorf("duplicate reclass event id %q", e.ID)
		}
		if !validScriptedNPC(e.NPC) || e.MinimumLevel < 1 || e.MinimumLevel > 255 ||
			e.AdvancedTargetClass < 0 || e.AdvancedTargetClass > 255 ||
			e.AdvancedFreeSourceClass < 0 || e.AdvancedFreeSourceClass > 255 ||
			e.AdvancedRequiredItemRaw < 0 || e.AdvancedRequiredItemRaw > 255 ||
			e.EffectID != "common:effect.reclass_original" ||
			e.PersonalInventorySlots < 1 || e.PersonalInventorySlots > 32 ||
			len(e.ClassNameTextIDsRaw) == 0 ||
			len(e.ForbiddenSourceClasses) == 0 || len(e.BasicTargetClasses) == 0 {
			return fmt.Errorf("%s: invalid selector, gate, effect or inventory size", e.ID)
		}
		seenClasses := map[int]string{}
		for kind, classes := range map[string][]int{
			"forbidden_source_classes_raw": e.ForbiddenSourceClasses,
			"basic_target_classes_raw":     e.BasicTargetClasses,
		} {
			for _, class := range classes {
				if class < 0 || class > 255 || seenClasses[class] != "" {
					return fmt.Errorf("%s: invalid or duplicate class %d in %s",
						e.ID, class, kind)
				}
				seenClasses[class] = kind
			}
		}
		if seenClasses[e.AdvancedTargetClass] != "" ||
			seenClasses[e.AdvancedFreeSourceClass] != "" {
			return fmt.Errorf("%s: advanced classes must be distinct from configured basic/forbidden classes", e.ID)
		}
		requiredClassNames := append([]int(nil), e.ForbiddenSourceClasses...)
		requiredClassNames = append(requiredClassNames, e.BasicTargetClasses...)
		requiredClassNames = append(requiredClassNames,
			e.AdvancedTargetClass, e.AdvancedFreeSourceClass)
		for _, class := range requiredClassNames {
			if e.ClassNameTextIDsRaw[strconv.Itoa(class)] == "" {
				return fmt.Errorf("%s: class_name_text_ids_raw missing configured class %d",
					e.ID, class)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.reclassEvents[e.ID] = e
	}
	p.stagedBossEvents = make(map[string]*StagedBossEvent, len(p.Events.StagedBossEvents))
	for i := range p.Events.StagedBossEvents {
		e := &p.Events.StagedBossEvents[i]
		if e.ID == "" || e.Kind != "staged_boss" {
			return fmt.Errorf("staged_boss_events[%d]: id and kind=staged_boss are required", i)
		}
		if _, exists := p.stagedBossEvents[e.ID]; exists {
			return fmt.Errorf("duplicate staged boss event id %q", e.ID)
		}
		if !validScriptedNPC(e.FirstNPC) || !validScriptedNPC(e.SecondNPC) ||
			e.FirstPresenceFlagRaw < 0 || e.FirstPresenceFlagRaw >= 512 ||
			e.SecondRequiredFlagRaw < 0 || e.SecondRequiredFlagRaw >= 512 ||
			e.StepFrames < 1 || e.StepFrames > 120 || len(e.FirstNPCPostBattleDirections) == 0 ||
			len(e.FirstPostBattleDirections) == 0 {
			return fmt.Errorf("%s: invalid trigger, selector, flag or movement timing", e.ID)
		}
		if err := validatePackFormation(e.ID, "first_formation", e.FirstFormation); err != nil {
			return err
		}
		if err := validatePackFormation(e.ID, "second_formation", e.SecondFormation); err != nil {
			return err
		}
		if e.FirstDropPolicy != "original" || e.SecondDropPolicy != "suppress" {
			return fmt.Errorf("%s: unsupported staged drop policy", e.ID)
		}
		for j, direction := range append(append([]string(nil), e.FirstNPCPostBattleDirections...), e.FirstPostBattleDirections...) {
			if _, ok := map[string]bool{"down": true, "up": true, "left": true, "right": true}[direction]; !ok {
				return fmt.Errorf("%s: invalid post-battle direction %d", e.ID, j)
			}
		}
		for name, flags := range map[string][]int{
			"first_victory_clear_flags_raw":  e.FirstVictoryClearFlagsRaw,
			"first_victory_set_flags_raw":    e.FirstVictorySetFlagsRaw,
			"second_victory_clear_flags_raw": e.SecondVictoryClearFlagsRaw,
			"second_victory_set_flags_raw":   e.SecondVictorySetFlagsRaw,
		} {
			if len(flags) == 0 {
				return fmt.Errorf("%s: %s must be present", e.ID, name)
			}
			seen := map[int]bool{}
			for _, flag := range flags {
				if flag < 0 || flag >= 512 || seen[flag] {
					return fmt.Errorf("%s: invalid or duplicate flag %d in %s", e.ID, flag, name)
				}
				seen[flag] = true
			}
		}
		for j, gate := range e.TreasureGates {
			if gate.CTYRaw < 0 || gate.CTYRaw > 255 || gate.Section < 0 ||
				gate.Tile.X < 0 || gate.Tile.Y < 0 || gate.ItemRawID < 0 ||
				gate.ItemRawID > 255 || gate.WhileFlagSet < 0 || gate.WhileFlagSet >= 512 {
				return fmt.Errorf("%s: invalid treasure gate %d", e.ID, j)
			}
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		p.stagedBossEvents[e.ID] = e
	}
	return nil
}

// BuiltinDQ3 returns the canonical repository DQ3 pack.
func BuiltinDQ3() (*Pack, error) {
	sub, err := fs.Sub(builtin, "packs/dq3_cht")
	if err != nil {
		return nil, fmt.Errorf("builtin dq3 pack: %w", err)
	}
	return Load(sub)
}

func (p *Pack) validateManifest() error {
	m := &p.Manifest
	switch {
	case m.SchemaVersion != SchemaVersion:
		return fmt.Errorf("unsupported schema_version %q (want %q)", m.SchemaVersion, SchemaVersion)
	case m.EngineAPI != EngineAPI:
		return fmt.Errorf("unsupported engine_api %q (want %q)", m.EngineAPI, EngineAPI)
	case m.PackID == "":
		return errors.New("pack_id is required")
	case m.Game != "dq1" && m.Game != "dq2" && m.Game != "dq3":
		return fmt.Errorf("unsupported game %q", m.Game)
	case m.Edition == "" || m.ContentVersion == "" || m.TitleTextID == "" ||
		m.EntryEventID == "" || m.SaveNamespace == "":
		return errors.New("edition, content_version, title_text_id, entry_event_id and save_namespace are required")
	case m.Capabilities == nil:
		return errors.New("capabilities must be present (use [] when empty)")
	case m.Data == nil:
		return errors.New("data must be present")
	case m.Assets == nil:
		return errors.New("assets must be present (use {} when empty)")
	}
	seen := make(map[string]bool, len(m.Capabilities))
	for _, capability := range m.Capabilities {
		if capability == "" || seen[capability] {
			return fmt.Errorf("invalid or duplicate capability %q", capability)
		}
		seen[capability] = true
	}
	for key, name := range m.Data {
		if key == "" {
			return errors.New("data contains an empty key")
		}
		if _, err := cleanRelative(name, "data."+key); err != nil {
			return err
		}
	}
	for key, asset := range m.Assets {
		if key == "" {
			return errors.New("assets contains an empty key")
		}
		if _, err := cleanRelative(asset.Path, "assets."+key+".path"); err != nil {
			return err
		}
		if asset.Size < 0 {
			return fmt.Errorf("assets.%s.size must not be negative", key)
		}
		if asset.AlternateSize < 0 {
			return fmt.Errorf("assets.%s.alternate_size must not be negative", key)
		}
		if asset.AlternateSHA256 != "" && len(asset.AlternateSHA256) != sha256.Size*2 {
			return fmt.Errorf("assets.%s.alternate_sha256 must be 64 hex characters", key)
		}
	}
	return nil
}

func (p *Pack) validateAudio() error {
	if p.Audio.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Audio.SchemaVersion)
	}
	if p.Audio.Cues == nil {
		return errors.New("cues must be present (use {} when empty)")
	}
	for id, cue := range p.Audio.Cues {
		if id == "" {
			return errors.New("cues contains an empty key")
		}
		if cue.Kind != "music" {
			return fmt.Errorf("audio cue %q has unsupported kind %q", id, cue.Kind)
		}
		if cue.Track < 0 || cue.Track > 255 {
			return fmt.Errorf("audio cue %q has invalid track %d", id, cue.Track)
		}
		if cue.Priority < 0 {
			return fmt.Errorf("audio cue %q priority must not be negative", id)
		}
		if err := validateEvidence(cue.Evidence); err != nil {
			return fmt.Errorf("audio cue %q evidence: %w", id, err)
		}
	}
	return nil
}

func validateBattleSourceFiles(files []BattleSourceFile) error {
	if files == nil {
		return errors.New("source_files must be present")
	}
	seen := make(map[string]bool, len(files))
	for i, file := range files {
		if file.Path == "" || file.Size <= 0 || len(file.SHA256) != sha256.Size*2 {
			return fmt.Errorf("source_files[%d] is incomplete", i)
		}
		if _, err := hex.DecodeString(file.SHA256); err != nil {
			return fmt.Errorf("source_files[%d] sha256 is invalid: %w", i, err)
		}
		if seen[file.Path] {
			return fmt.Errorf("duplicate source file %q", file.Path)
		}
		seen[file.Path] = true
	}
	return nil
}

func validateBattleRawBlob(name string, blob BattleRawBlob) error {
	if blob.AddressSpace == "" || blob.Address == "" || blob.LengthBytes <= 0 ||
		len(blob.RawSHA256) != sha256.Size*2 {
		return fmt.Errorf("%s has incomplete address/length/hash", name)
	}
	raw, err := hex.DecodeString(blob.RawHex)
	if err != nil {
		return fmt.Errorf("%s raw_hex is invalid: %w", name, err)
	}
	if len(raw) != blob.LengthBytes {
		return fmt.Errorf("%s raw_hex length=%d want %d", name, len(raw), blob.LengthBytes)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(raw)); got != blob.RawSHA256 {
		return fmt.Errorf("%s raw_sha256=%s want %s", name, blob.RawSHA256, got)
	}
	if err := validateEvidence(blob.Evidence); err != nil {
		return fmt.Errorf("%s evidence: %w", name, err)
	}
	return nil
}

func validBattleBackgroundSelector(selector BattleBackgroundSelector, background BattleBackgroundData) bool {
	return selector.PageRaw >= 0 && selector.PageRaw < background.PageCount &&
		selector.PaletteBankRaw >= 0 && selector.PaletteBankRaw < background.PaletteBankCount
}

func (p *Pack) validateBattleBackground(background BattleBackgroundData) error {
	if background.ArchiveAsset == "" || background.PaletteAsset == "" ||
		background.PageStrideBytes <= 0 || background.FieldOffsetBytes < 0 ||
		background.FieldChunkBytes <= 0 || background.WidthPixels <= 0 ||
		background.WidthPixels%8 != 0 || background.FieldRows <= 0 ||
		background.PlaneCount < 1 || background.PlaneCount > 8 ||
		background.PageCount <= 0 || background.PaletteBankCount <= 0 ||
		background.PaletteEntriesPerBank <= 0 || background.PaletteEntriesPerBank > 256 ||
		background.PaletteBankCount*background.PaletteEntriesPerBank > 256 {
		return errors.New("battle background contract is incomplete")
	}
	if _, ok := p.Manifest.Assets[background.ArchiveAsset]; !ok {
		return fmt.Errorf("battle background archive asset %q is absent from manifest", background.ArchiveAsset)
	}
	if _, ok := p.Manifest.Assets[background.PaletteAsset]; !ok {
		return fmt.Errorf("battle background palette asset %q is absent from manifest", background.PaletteAsset)
	}
	wantChunk := background.WidthPixels / 8 * background.FieldRows * background.PlaneCount
	if background.FieldChunkBytes != wantChunk ||
		background.FieldOffsetBytes+background.FieldChunkBytes > background.PageStrideBytes {
		return fmt.Errorf("battle background planar layout is invalid: chunk=%d want=%d offset=%d stride=%d",
			background.FieldChunkBytes, wantChunk, background.FieldOffsetBytes, background.PageStrideBytes)
	}
	if !validBattleBackgroundSelector(background.DefaultSelector, background) {
		return errors.New("battle background default selector is outside the declared archive/palette range")
	}
	if err := validateEvidence(background.Evidence); err != nil {
		return fmt.Errorf("battle background evidence: %w", err)
	}
	return nil
}

func (p *Pack) validateBattlePack() error {
	b := p.Battle
	if b.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", b.SchemaVersion)
	}
	if err := validateEvidence(b.Evidence); err != nil {
		return fmt.Errorf("battle evidence: %w", err)
	}
	if err := validateBattleSourceFiles(b.SourceFiles); err != nil {
		return err
	}
	if err := p.validateBattleBackground(b.Background); err != nil {
		return err
	}
	if err := validateBattleRawBlob("encounter.region_lookup", b.Encounter.RegionLookup); err != nil {
		return err
	}
	if err := validateBattleRawBlob("encounter.candidate_table", b.Encounter.CandidateTable); err != nil {
		return err
	}
	if b.Encounter.RowStrideBytes != 8 || b.Encounter.RegionStrideBytes != 0x20 {
		return errors.New("encounter raw strides do not match confirmed contract")
	}
	if len(b.Encounter.UsedRegionIDs) != 27 {
		return fmt.Errorf("encounter used_region_ids=%d want 27", len(b.Encounter.UsedRegionIDs))
	}
	for i, id := range b.Encounter.UsedRegionIDs {
		if id != i+1 {
			return fmt.Errorf("encounter used_region_ids[%d]=%d want %d", i, id, i+1)
		}
	}
	if len(b.Encounter.FixedRecords) == 0 {
		return errors.New("encounter fixed_records must not be empty")
	}
	for i, record := range b.Encounter.FixedRecords {
		if record.DGroup == "" || record.File == "" || record.Caller == "" || record.Count <= 0 {
			return fmt.Errorf("encounter fixed_records[%d] is incomplete", i)
		}
		formation, err := fixedBattleFormation(record)
		if err != nil {
			return fmt.Errorf("encounter fixed_records[%d]: %w", i, err)
		}
		if !validBattleBackgroundSelector(formation.Background, b.Background) {
			return fmt.Errorf("encounter fixed_records[%d] background selector page=%d palette_bank=%d is outside declared range",
				i, formation.Background.PageRaw, formation.Background.PaletteBankRaw)
		}
		if err := validateEvidence(record.Evidence); err != nil {
			return fmt.Errorf("encounter fixed_records[%d] evidence: %w", i, err)
		}
	}
	position := b.Encounter.FormationPosition
	if position.OriginRaw != 0x26 || position.FirstMemberOffsetRaw != 2 || position.WeightStepMultiplier != 2 ||
		position.EGAStrideBytes != 0x54 || position.EGABottomRaw != 0x102 ||
		position.PixelsPerRawByte != 8 {
		return fmt.Errorf("encounter formation_position=%+v does not match confirmed raw projection", position)
	}
	if err := validateEvidence(position.Evidence); err != nil {
		return fmt.Errorf("encounter formation_position evidence: %w", err)
	}
	if err := validateBattleRawBlob("resistance.effect_thresholds", b.Resistance.EffectThresholds); err != nil {
		return err
	}
	if len(b.Resistance.ClassThresholds) != 4 ||
		!reflectIntSliceEqual(b.Resistance.ClassThresholds, []int{0, 68, 180, 255}) {
		return fmt.Errorf("resistance class_thresholds=%v want [0 68 180 255]", b.Resistance.ClassThresholds)
	}
	if err := validateBattleRawBlob("resistance.descriptor_table", b.Resistance.DescriptorTable); err != nil {
		return err
	}
	if len(b.Resistance.EffectNamesZH) != 60 {
		return fmt.Errorf("resistance effect_names_zh=%d want 60", len(b.Resistance.EffectNamesZH))
	}
	for i := 0; i < 60; i++ {
		if b.Resistance.EffectNamesZH[strconv.Itoa(i)] == "" {
			return fmt.Errorf("resistance effect_names_zh missing %d", i)
		}
	}
	if err := validateEvidence(b.Resistance.Evidence); err != nil {
		return fmt.Errorf("resistance evidence: %w", err)
	}
	return nil
}

// validateBattleFormationSelectors runs after both events and the battle
// archive contract have loaded.  Earlier event validation only proves that
// bytes 1/2 match raw_bytes_hex; this pass proves their references fit the
// declared page and palette-bank ranges without inventing a fallback.
func (p *Pack) validateBattleFormationSelectors() error {
	if !p.battlePresent {
		return errors.New("battle background contract is unavailable")
	}
	background := p.Battle.Background
	check := func(eventID, name string, formation BattleFormation) error {
		if !validBattleBackgroundSelector(formation.Background, background) {
			return fmt.Errorf("%s: %s background selector page=%d palette_bank=%d is outside declared range",
				eventID, name, formation.Background.PageRaw, formation.Background.PaletteBankRaw)
		}
		return nil
	}
	for _, event := range p.Events.BossSurrenderEvents {
		if err := check(event.ID, "formation", event.Formation); err != nil {
			return err
		}
	}
	for _, event := range p.Events.HostageRescueEvents {
		if err := check(event.ID, "guard_formation", event.GuardFormation); err != nil {
			return err
		}
		if err := check(event.ID, "boss_formation", event.BossFormation); err != nil {
			return err
		}
	}
	for _, event := range p.Events.StagedBossEvents {
		if err := check(event.ID, "first_formation", event.FirstFormation); err != nil {
			return err
		}
		if err := check(event.ID, "second_formation", event.SecondFormation); err != nil {
			return err
		}
	}
	return nil
}

func reflectIntSliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func validateEvidence(e Evidence) error {
	if e.Level != "D1" && e.Level != "D2" && e.Level != "D3" {
		return fmt.Errorf("invalid evidence level %q", e.Level)
	}
	switch e.SourceKind {
	case "exe", "data_file", "dosbox", "video", "manual":
	default:
		return fmt.Errorf("invalid source_kind %q", e.SourceKind)
	}
	switch e.AddressSpace {
	case "file", "logical", "linear", "dgroup", "record", "none":
	default:
		return fmt.Errorf("invalid address_space %q", e.AddressSpace)
	}
	if e.Source == "" || e.Doc == "" {
		return errors.New("evidence source and doc are required")
	}
	if e.AddressSpace != "none" && e.Address == "" {
		return errors.New("evidence address is required unless address_space is none")
	}
	if e.Level == "D3" && e.Consumer == "" {
		return errors.New("D3 evidence requires consumer")
	}
	return nil
}

func (p *Pack) validateFacilities() error {
	if p.Facilities.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Facilities.SchemaVersion)
	}
	if p.Facilities.ServiceDefinitions == nil {
		return errors.New("service_definitions must be present")
	}
	p.services = make(map[string]*ServiceDefinition, len(p.Facilities.ServiceDefinitions))
	for i := range p.Facilities.ServiceDefinitions {
		s := &p.Facilities.ServiceDefinitions[i]
		if s.ID == "" {
			return fmt.Errorf("service_definitions[%d].id is required", i)
		}
		if _, exists := p.services[s.ID]; exists {
			return fmt.Errorf("duplicate service id %q", s.ID)
		}
		if s.Pricing.FormulaID != "common:formula.level_table" {
			return fmt.Errorf("%s: unsupported pricing formula %q", s.ID, s.Pricing.FormulaID)
		}
		if s.Pricing.LevelCap < 1 || len(s.Pricing.CostsGold) != s.Pricing.LevelCap {
			return fmt.Errorf("%s: level_cap %d must equal costs_gold length %d",
				s.ID, s.Pricing.LevelCap, len(s.Pricing.CostsGold))
		}
		for level, cost := range s.Pricing.CostsGold {
			if cost < 0 {
				return fmt.Errorf("%s: negative level %d cost", s.ID, level+1)
			}
		}
		if err := validateEvidence(s.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", s.ID, err)
		}
		p.services[s.ID] = s
	}
	if _, ok := p.services[ReviveService]; !ok {
		return fmt.Errorf("required service %q is missing", ReviveService)
	}
	return nil
}

// ReviveCost returns the original per-level revival cost. Levels outside the
// table are clamped exactly like the original DQ3 consumer.
func (p *Pack) ReviveCost(level int) int {
	s := p.services[ReviveService]
	if level < 1 {
		level = 1
	}
	if level > s.Pricing.LevelCap {
		level = s.Pricing.LevelCap
	}
	return s.Pricing.CostsGold[level-1]
}

func (p *Pack) BossSurrenderEvents() []BossSurrenderEvent {
	return append([]BossSurrenderEvent(nil), p.Events.BossSurrenderEvents...)
}

func (p *Pack) BossSurrenderEvent(id string) (*BossSurrenderEvent, bool) {
	e, ok := p.bossEvents[id]
	return e, ok
}

func (p *Pack) TemporaryRoleEvents() []TemporaryRoleEvent {
	return append([]TemporaryRoleEvent(nil), p.Events.TemporaryRoleEvents...)
}

func (p *Pack) TemporaryRoleEvent(id string) (*TemporaryRoleEvent, bool) {
	e, ok := p.roleEvents[id]
	return e, ok
}

func (p *Pack) TemporarySoloChallenges() []TemporarySoloChallengeEvent {
	return append([]TemporarySoloChallengeEvent(nil), p.Events.TemporarySoloChallenges...)
}

func (p *Pack) TemporarySoloChallenge(id string) (*TemporarySoloChallengeEvent, bool) {
	e, ok := p.soloChallenges[id]
	return e, ok
}

func (p *Pack) QuestItemChainEvents() []QuestItemChainEvent {
	return append([]QuestItemChainEvent(nil), p.Events.QuestItemChainEvents...)
}

func (p *Pack) QuestItemChainEvent(id string) (*QuestItemChainEvent, bool) {
	e, ok := p.questItemEvents[id]
	return e, ok
}

func (p *Pack) ChoiceItemExchangeEvents() []ChoiceItemExchangeEvent {
	return append([]ChoiceItemExchangeEvent(nil), p.Events.ChoiceItemExchangeEvents...)
}

func (p *Pack) ChoiceItemExchangeEvent(id string) (*ChoiceItemExchangeEvent, bool) {
	e, ok := p.choiceItemExchangeEvents[id]
	return e, ok
}

func (p *Pack) TrackedWorldObjects() []TrackedWorldObject {
	return append([]TrackedWorldObject(nil), p.Events.TrackedWorldObjects...)
}

func (p *Pack) TrackedWorldObject(id string) (*TrackedWorldObject, bool) {
	e, ok := p.trackedWorldObjects[id]
	return e, ok
}

func (p *Pack) TrackedWorldObjectByTrackerItem(rawID int) (*TrackedWorldObject, bool) {
	for i := range p.Events.TrackedWorldObjects {
		e := &p.Events.TrackedWorldObjects[i]
		if e.TrackerItemRawID == rawID {
			return e, true
		}
	}
	return nil, false
}

func (p *Pack) CoordinateItemGateEvents() []CoordinateItemGateEvent {
	return append([]CoordinateItemGateEvent(nil), p.Events.CoordinateItemGateEvents...)
}

func (p *Pack) CoordinateItemGateEvent(id string) (*CoordinateItemGateEvent, bool) {
	e, ok := p.coordinateItemGateEvents[id]
	return e, ok
}

func (p *Pack) CoordinateItemGateAt(x, y, layer int) (*CoordinateItemGateEvent, bool) {
	for i := range p.Events.CoordinateItemGateEvents {
		e := &p.Events.CoordinateItemGateEvents[i]
		if e.Coordinate == (VehicleWorldPosition{X: x, Y: y, Layer: layer}) {
			return e, true
		}
	}
	return nil, false
}

func (p *Pack) TreasureEvents() []TreasureEvent {
	return append([]TreasureEvent(nil), p.Events.TreasureEvents...)
}

func (p *Pack) TreasureEvent(id string) (*TreasureEvent, bool) {
	e, ok := p.treasureEvents[id]
	return e, ok
}

func (p *Pack) NPCItemRewardEvents() []NPCItemRewardEvent {
	return append([]NPCItemRewardEvent(nil), p.Events.NPCItemRewardEvents...)
}

func (p *Pack) NPCItemRewardEvent(id string) (*NPCItemRewardEvent, bool) {
	for i := range p.Events.NPCItemRewardEvents {
		if p.Events.NPCItemRewardEvents[i].ID == id {
			event := p.Events.NPCItemRewardEvents[i]
			return &event, true
		}
	}
	return nil, false
}

func (p *Pack) RuraDestinations() []RuraDestination {
	return append([]RuraDestination(nil), p.Events.RuraNavigation.Destinations...)
}

func (p *Pack) WorldEntranceVariants() []WorldEntranceVariant {
	result := append([]WorldEntranceVariant(nil), p.Events.WorldEntranceVariants...)
	for i := range result {
		result[i].Branches = append([]WorldEntranceVariantBranch(nil), result[i].Branches...)
	}
	return result
}

func (p *Pack) SettlementFounderEvents() []SettlementFounderEvent {
	result := append([]SettlementFounderEvent(nil), p.Events.SettlementFounderEvents...)
	for i := range result {
		result[i].FounderVisibilityFlagsRaw = append([]int(nil),
			result[i].FounderVisibilityFlagsRaw...)
	}
	return result
}

func (p *Pack) SettlementFounderEvent(id string) (*SettlementFounderEvent, bool) {
	e, ok := p.settlementFounders[id]
	return e, ok
}

func (p *Pack) SettlementFounderFollowups() []SettlementFounderFollowupEvent {
	result := append([]SettlementFounderFollowupEvent(nil), p.Events.SettlementFounderFollowups...)
	for i := range result {
		result[i].RequiredFlagsSetRaw = append([]int(nil), result[i].RequiredFlagsSetRaw...)
		result[i].RequiredFlagsClearRaw = append([]int(nil), result[i].RequiredFlagsClearRaw...)
		result[i].DialogueTextIDs = append([]string(nil), result[i].DialogueTextIDs...)
	}
	return result
}

func (p *Pack) ConditionalFacilityEvents() []ConditionalFacilityEvent {
	return append([]ConditionalFacilityEvent(nil), p.Events.ConditionalFacilityEvents...)
}

func (p *Pack) ConditionalFacilityEvent(id string) (*ConditionalFacilityEvent, bool) {
	for i := range p.Events.ConditionalFacilityEvents {
		if p.Events.ConditionalFacilityEvents[i].ID == id {
			event := p.Events.ConditionalFacilityEvents[i]
			return &event, true
		}
	}
	return nil, false
}

func (p *Pack) ItemUseEffects() []ItemUseEffect {
	return append([]ItemUseEffect(nil), p.Events.ItemUseEffects...)
}

func (p *Pack) TrackingGuardEvents() []TrackingGuardEvent {
	return append([]TrackingGuardEvent(nil), p.Events.TrackingGuardEvents...)
}

func (p *Pack) PushPuzzleEvents() []PushPuzzleEvent {
	return append([]PushPuzzleEvent(nil), p.Events.PushPuzzleEvents...)
}

func (p *Pack) NPCPushRule() NPCPushRule {
	return p.Events.NPCPushRule
}

func (p *Pack) ItemUseEffectByRawID(itemRawID int) (*ItemUseEffect, bool) {
	e, ok := p.itemUseEffects[itemRawID]
	return e, ok
}

func (p *Pack) DialogueWindowLayout() WindowLayout {
	return p.Interface.Dialogue
}

// BattleMessageWindowLayout returns the pack-owned message window used while
// a battle action or result is being displayed. Production callers must check
// the boolean instead of inventing a version-specific rectangle.
func (p *Pack) BattleMessageWindowLayout() (WindowLayout, bool) {
	if p.Interface.BattleMessage.ID == "" {
		return WindowLayout{}, false
	}
	return p.Interface.BattleMessage, true
}

// BattleCommandLayout returns the pack-owned command/selection panel rule.
func (p *Pack) BattleCommandLayout() (BattlePanelLayout, bool) {
	if p.Interface.BattleCommand.ID == "" {
		return BattlePanelLayout{}, false
	}
	return p.Interface.BattleCommand, true
}

// BattleEnemyLayout returns the pack-owned enemy-name/count panel rule.
func (p *Pack) BattleEnemyLayout() (BattlePanelLayout, bool) {
	if p.Interface.BattleEnemy.ID == "" {
		return BattlePanelLayout{}, false
	}
	return p.Interface.BattleEnemy, true
}

// BattleSceneLayout returns the pack-owned battle field band and cursor glyph.
// Production bootstrap must require it before a battle can be entered.
func (p *Pack) BattleSceneLayout() (BattleSceneLayout, bool) {
	if p == nil || p.Interface.BattleScene == nil || p.Interface.BattleScene.ID == "" {
		return BattleSceneLayout{}, false
	}
	return *p.Interface.BattleScene, true
}

// NewGameLabels returns the pack-owned glyph streams used by the title,
// creation, name-input and gender screens. Production bootstrap must require
// this contract before opening the title flow; lightweight fixtures may omit it.
func (p *Pack) NewGameLabels() (NewGameLabels, bool) {
	if p == nil || p.Interface.NewGameLabels == nil {
		return NewGameLabels{}, false
	}
	return *p.Interface.NewGameLabels, true
}

// NewGameGeometry returns the pack-owned pixel anchors and raw window index
// used by the opening creation screens. Production bootstrap must require it.
func (p *Pack) NewGameGeometry() (NewGameGeometry, bool) {
	if p == nil || p.Interface.NewGameGeometry == nil {
		return NewGameGeometry{}, false
	}
	return *p.Interface.NewGameGeometry, true
}

// BattleCommandLabels returns the pack-owned glyphs for the five standard
// command roles. Production bootstrap must require this contract before a
// battle can be entered; lightweight fixtures may omit it deliberately.
func (p *Pack) BattleCommandLabels() (BattleCommandLabels, bool) {
	if p == nil || p.Interface.BattleCommandLabels == nil {
		return BattleCommandLabels{}, false
	}
	return *p.Interface.BattleCommandLabels, true
}

// FieldCommandLabels returns the pack-owned glyphs for the field command
// menu. Production bootstrap must require this contract before opening the
// menu; lightweight fixtures may omit it deliberately.
func (p *Pack) FieldCommandLabels() (FieldCommandLabels, bool) {
	if p == nil || p.Interface.FieldCommandLabels == nil {
		return FieldCommandLabels{}, false
	}
	return *p.Interface.FieldCommandLabels, true
}

// FieldStatusLayout returns the pack-owned detailed status panel. Production
// bootstrap must require it before opening the field status command; fixtures
// may omit it deliberately when they do not exercise that screen.
func (p *Pack) FieldStatusLayout() (FieldStatusLayout, bool) {
	if p == nil || p.Interface.FieldStatus == nil || p.Interface.FieldStatus.ID == "" {
		return FieldStatusLayout{}, false
	}
	return *p.Interface.FieldStatus, true
}

// PartyHUDLayout is absent for packs that do not expose a longitudinal party
// HUD.  Callers must check the boolean instead of inventing a geometry.
func (p *Pack) PartyHUDLayout() (PartyHUDLayout, bool) {
	if p.Interface.PartyHUD.ID == "" {
		return PartyHUDLayout{}, false
	}
	return p.Interface.PartyHUD, true
}

func (p *Pack) TwoStepFloorSwitchGates() []TwoStepFloorSwitchGate {
	return append([]TwoStepFloorSwitchGate(nil), p.Events.TwoStepFloorSwitchGates...)
}

func (p *Pack) TwoStepFloorSwitchGate(id string) (*TwoStepFloorSwitchGate, bool) {
	e, ok := p.sequenceGates[id]
	return e, ok
}

func (p *Pack) StagedVehicleExchangeEvents() []StagedVehicleExchangeEvent {
	return append([]StagedVehicleExchangeEvent(nil), p.Events.StagedVehicleExchangeEvents...)
}

func (p *Pack) StagedVehicleExchangeEvent(id string) (*StagedVehicleExchangeEvent, bool) {
	e, ok := p.vehicleExchanges[id]
	return e, ok
}

func (p *Pack) GuidedPassageEvents() []GuidedPassageEvent {
	return append([]GuidedPassageEvent(nil), p.Events.GuidedPassageEvents...)
}

func (p *Pack) GuidedPassageEvent(id string) (*GuidedPassageEvent, bool) {
	e, ok := p.guidedPassages[id]
	return e, ok
}

func (p *Pack) HostageRescueEvents() []HostageRescueEvent {
	return append([]HostageRescueEvent(nil), p.Events.HostageRescueEvents...)
}

func (p *Pack) HostageRescueEvent(id string) (*HostageRescueEvent, bool) {
	e, ok := p.hostageRescues[id]
	return e, ok
}

func (p *Pack) ReclassEvents() []ReclassEvent {
	return append([]ReclassEvent(nil), p.Events.ReclassEvents...)
}

func (p *Pack) ItemActions() ItemActions {
	return p.Events.ItemActions
}

func (p *Pack) ReclassEvent(id string) (*ReclassEvent, bool) {
	e, ok := p.reclassEvents[id]
	return e, ok
}

func (p *Pack) StagedBossEvents() []StagedBossEvent {
	return append([]StagedBossEvent(nil), p.Events.StagedBossEvents...)
}

func (p *Pack) StagedBossEvent(id string) (*StagedBossEvent, bool) {
	e, ok := p.stagedBossEvents[id]
	return e, ok
}

// StoryFlagRuntimeEvents returns a copy of the finite, evidence-backed
// handler contracts. Callers must select by kind and verify the NPC/scene
// selector before applying a transaction.
func (p *Pack) StoryFlagRuntimeEvents() []StoryFlagRuntimeEvent {
	if p == nil {
		return nil
	}
	return append([]StoryFlagRuntimeEvent(nil), p.Events.StoryFlagRuntimeEvents...)
}

func (p *Pack) StoryFlagRuntimeEventKind(kind string) (*StoryFlagRuntimeEvent, bool) {
	if p == nil {
		return nil, false
	}
	for i := range p.Events.StoryFlagRuntimeEvents {
		if p.Events.StoryFlagRuntimeEvents[i].Kind == kind {
			event := p.Events.StoryFlagRuntimeEvents[i]
			return &event, true
		}
	}
	return nil, false
}

// CharacterEquipment resolves one pack character default into engine slots.
// Missing JSON slots become the engine's explicit -1 empty sentinel.
func (p *Pack) CharacterEquipment(id string) ([4]int, bool) {
	d, ok := p.charDefaults[id]
	out := [4]int{-1, -1, -1, -1}
	if !ok {
		return out, false
	}
	for i, code := range []*int{
		d.Equipment.Weapon, d.Equipment.Armor, d.Equipment.Shield, d.Equipment.Head,
	} {
		if code != nil {
			out[i] = *code
		}
	}
	return out, true
}

func (p *Pack) NewGamePlayerEquipment() ([4]int, bool) {
	return p.CharacterEquipment(p.Characters.DefaultRefs.NewGamePlayer)
}

func (p *Pack) RegisteredPartyMemberEquipment() ([4]int, bool) {
	ref := p.Characters.DefaultRefs.RegisteredPartyMember
	if ref == "" {
		return [4]int{-1, -1, -1, -1}, false
	}
	return p.CharacterEquipment(ref)
}

// PartySprite resolves a pack-defined class/gender pair to its asset and BLS
// entry base.  The engine does not apply a version-specific arithmetic rule.
func (p *Pack) PartySprite(classRaw, genderRaw int) (string, int, bool) {
	for _, entry := range p.Characters.PartySprite.Entries {
		if entry.ClassRaw == classRaw && entry.GenderRaw == genderRaw {
			return p.Characters.PartySprite.Asset, entry.EntryBase, true
		}
	}
	return "", 0, false
}

func (p *Pack) TextGlyphCodes(id string) ([]uint16, bool) {
	d, ok := p.texts[id]
	if !ok {
		return nil, false
	}
	out := make([]uint16, len(d.GlyphCodes))
	for i, code := range d.GlyphCodes {
		out[i] = uint16(code)
	}
	return out, true
}

func (p *Pack) TextDefinition(id string) (*TextDefinition, bool) {
	d, ok := p.texts[id]
	return d, ok
}

// BattleTextRefs returns the optional battle role contract. A missing
// contract is deliberately distinguishable from an empty one so production
// bootstrap can fail closed while lightweight data-fixture tests remain able
// to exercise unrelated pack validators.
func (p *Pack) BattleTextRefs() (*BattleTextRefs, bool) {
	if p == nil || p.Interface.BattleTexts == nil {
		return nil, false
	}
	return p.Interface.BattleTexts, true
}

// Asset returns a pack-owned logical asset reference. The engine remains
// responsible for reading the referenced bytes from its external asset FS.
func (p *Pack) Asset(key string) (AssetRef, bool) {
	if p == nil {
		return AssetRef{}, false
	}
	a, ok := p.Manifest.Assets[key]
	return a, ok
}

func (p *Pack) AssetPath(key string) (string, bool) {
	a, ok := p.Asset(key)
	return a.Path, ok
}

// AttractSequence returns the pack-owned title attract contract. A nil result
// means this edition has no verified attract sequence and the engine must keep
// the normal title splash path.
func (p *Pack) AttractSequence() (*AttractSequence, bool) {
	if p == nil || p.Interface.Attract == nil {
		return nil, false
	}
	return p.Interface.Attract, true
}

// OpeningSequence returns the pack-owned boot presentation. A nil result means
// this edition has no declared opening cutscene and the caller may start at
// the normal title splash.
func (p *Pack) OpeningSequence() (*OpeningSequence, bool) {
	if p == nil || p.Interface.Opening == nil {
		return nil, false
	}
	return p.Interface.Opening, true
}

// NewGameConfirmation returns the optional versioned raw screen used behind
// the new-game ability confirmation modal. Missing data means a black fallback
// remains available; malformed declared data is rejected during pack load.
func (p *Pack) NewGameConfirmation() (*RawScreenAsset, bool) {
	if p == nil || p.Interface.NewGameConfirmation == nil {
		return nil, false
	}
	return p.Interface.NewGameConfirmation, true
}

// AudioTrack resolves a stable cue ID to the pack's original track number.
func (p *Pack) AudioTrack(id string) (int, bool) {
	if p == nil {
		return 0, false
	}
	cue, ok := p.Audio.Cues[id]
	if !ok {
		return 0, false
	}
	return cue.Track, true
}

// BattlePack returns the statically confirmed E2 battle contract.  A missing
// contract is distinguishable from an empty value so production bootstrap can
// fail closed instead of silently reading a DQ3-specific Go fallback.
func (p *Pack) BattlePack() (BattlePackData, bool) {
	if p == nil || !p.battlePresent {
		return BattlePackData{}, false
	}
	return p.Battle, true
}

// BattleBackground returns the pack-owned archive, planar layout and default
// selector.  Production callers must check the boolean and fail closed.
func (p *Pack) BattleBackground() (BattleBackgroundData, bool) {
	if p == nil || !p.battlePresent || p.Battle.Background.Evidence.Level == "" {
		return BattleBackgroundData{}, false
	}
	return p.Battle.Background, true
}

// BattleEncounterRaw decodes the two raw encounter blobs after strict pack
// validation.  It is intentionally a narrow accessor for the shared decoder.
func (p *Pack) BattleEncounterRaw() ([]byte, []byte, bool) {
	if p == nil || !p.battlePresent {
		return nil, nil, false
	}
	region, err := hex.DecodeString(p.Battle.Encounter.RegionLookup.RawHex)
	if err != nil {
		return nil, nil, false
	}
	candidates, err := hex.DecodeString(p.Battle.Encounter.CandidateTable.RawHex)
	if err != nil {
		return nil, nil, false
	}
	return region, candidates, true
}

// FixedBattleFormationForSingleMonster 找到唯一的原版固定單敵編隊。它只消費
// pack 已驗證的 raw formation；若同一怪物有多筆固定 record 或完全沒有 record，
// 則 fail-closed，呼叫端不得退回草地預設背景。
func (p *Pack) FixedBattleFormationForSingleMonster(monsterRawID int) (BattleFormation, bool) {
	if p == nil || !p.battlePresent || monsterRawID < 0 || monsterRawID > 255 {
		return BattleFormation{}, false
	}
	var found *BattleFormation
	for _, record := range p.Battle.Encounter.FixedRecords {
		formation, err := fixedBattleFormation(record)
		if err != nil {
			return BattleFormation{}, false
		}
		if len(formation.Groups) != 1 || formation.Groups[0].Count != 1 ||
			formation.Groups[0].MonsterRawID != monsterRawID {
			continue
		}
		if found != nil || !validBattleBackgroundSelector(formation.Background, p.Battle.Background) {
			return BattleFormation{}, false
		}
		copy := formation
		found = &copy
	}
	if found == nil {
		return BattleFormation{}, false
	}
	return *found, true
}

// BattleFormationPosition returns the pack-owned raw position projection.
// Production callers must validate the boolean instead of inventing a
// version-specific horizontal or vertical fallback.
func (p *Pack) BattleFormationPosition() (BattleFormationPosition, bool) {
	if p == nil || !p.battlePresent || p.Battle.Encounter.FormationPosition.Evidence.Level == "" {
		return BattleFormationPosition{}, false
	}
	return p.Battle.Encounter.FormationPosition, true
}

func (p *Pack) ID() string             { return p.Manifest.PackID }
func (p *Pack) Schema() string         { return p.Manifest.SchemaVersion }
func (p *Pack) ContentHash() string    { return p.contentHash }
func (p *Pack) ContentVersion() string { return p.Manifest.ContentVersion }
