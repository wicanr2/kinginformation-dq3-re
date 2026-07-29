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
	"strings"
)

const (
	SchemaVersion = "0.1.0"
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
	Path   string `json:"path"`
	Size   int64  `json:"size"`
	SHA256 string `json:"sha256"`
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
	BackgroundRaw int              `json:"background_raw"`
	PageRaw       int              `json:"page_raw"`
	Groups        []FormationGroup `json:"groups"`
	RawBytesHex   string           `json:"raw_bytes_hex"`
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

type Events struct {
	SchemaVersion       string               `json:"schema_version"`
	BossSurrenderEvents []BossSurrenderEvent `json:"boss_surrender_events"`
	TemporaryRoleEvents []TemporaryRoleEvent `json:"temporary_role_events"`
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

type Characters struct {
	SchemaVersion string               `json:"schema_version"`
	DefaultRefs   CharacterDefaultRefs `json:"default_refs"`
	Defaults      []CharacterDefault   `json:"defaults"`
}

type Pack struct {
	Manifest     Manifest
	Facilities   Facilities
	Events       Events
	Characters   Characters
	Texts        Texts
	services     map[string]*ServiceDefinition
	bossEvents   map[string]*BossSurrenderEvent
	roleEvents   map[string]*TemporaryRoleEvent
	charDefaults map[string]*CharacterDefault
	texts        map[string]*TextDefinition
	contentHash  string
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
	if err := p.validateEventTextRefs(); err != nil {
		return nil, fmt.Errorf("%s: %w", eventsPath, err)
	}
	canonical, err := json.Marshal(struct {
		Manifest   Manifest   `json:"manifest"`
		Facilities Facilities `json:"facilities"`
		Events     Events     `json:"events"`
		Characters Characters `json:"characters"`
		Texts      Texts      `json:"texts"`
	}{p.Manifest, p.Facilities, p.Events, p.Characters, p.Texts})
	if err != nil {
		return nil, fmt.Errorf("canonicalize pack: %w", err)
	}
	p.contentHash = fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	return &p, nil
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

func (p *Pack) validateEventTextRefs() error {
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
	return nil
}

func validScriptedNPC(s ScriptedNPCSelector) bool {
	return s.CTYRaw >= 0 && s.CTYRaw <= 255 && s.Section >= 0 &&
		s.Tile.X >= 0 && s.Tile.Y >= 0 && s.HandlerRaw >= 0 && s.HandlerRaw <= 255
}

func (p *Pack) validateEvents() error {
	if p.Events.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Events.SchemaVersion)
	}
	if p.Events.BossSurrenderEvents == nil {
		return errors.New("boss_surrender_events must be present")
	}
	if p.Events.TemporaryRoleEvents == nil {
		return errors.New("temporary_role_events must be present")
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
		if e.Formation.BackgroundRaw < 0 || e.Formation.PageRaw < 0 ||
			len(e.Formation.Groups) == 0 || e.Formation.RawBytesHex == "" {
			return fmt.Errorf("%s: incomplete formation", e.ID)
		}
		raw, err := hex.DecodeString(e.Formation.RawBytesHex)
		if err != nil {
			return fmt.Errorf("%s: invalid formation raw_bytes_hex: %w", e.ID, err)
		}
		if len(raw) != 3+len(e.Formation.Groups)*2 || int(raw[0]) != len(e.Formation.Groups) ||
			int(raw[1]) != e.Formation.BackgroundRaw || int(raw[2]) != e.Formation.PageRaw {
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
	}
	return nil
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
	case "file", "logical", "dgroup", "record", "none":
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

func (p *Pack) ID() string             { return p.Manifest.PackID }
func (p *Pack) Schema() string         { return p.Manifest.SchemaVersion }
func (p *Pack) ContentHash() string    { return p.contentHash }
func (p *Pack) ContentVersion() string { return p.Manifest.ContentVersion }
