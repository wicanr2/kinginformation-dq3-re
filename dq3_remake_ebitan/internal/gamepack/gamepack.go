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

type DialogueRecords struct {
	Intro   int `json:"intro"`
	Apology int `json:"apology"`
	Accept  int `json:"accept"`
	Reject  int `json:"reject"`
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
	DialogueRecords DialogueRecords  `json:"dialogue_records"`
	Formation       BattleFormation  `json:"formation"`
	ClearFlagRaw    int              `json:"clear_flag_raw"`
	TreasureGates   []TreasureGate   `json:"treasure_gates"`
	Evidence        Evidence         `json:"evidence"`
}

type Events struct {
	SchemaVersion       string               `json:"schema_version"`
	BossSurrenderEvents []BossSurrenderEvent `json:"boss_surrender_events"`
}

type Pack struct {
	Manifest    Manifest
	Facilities  Facilities
	Events      Events
	services    map[string]*ServiceDefinition
	bossEvents  map[string]*BossSurrenderEvent
	contentHash string
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
	canonical, err := json.Marshal(struct {
		Manifest   Manifest   `json:"manifest"`
		Facilities Facilities `json:"facilities"`
		Events     Events     `json:"events"`
	}{p.Manifest, p.Facilities, p.Events})
	if err != nil {
		return nil, fmt.Errorf("canonicalize pack: %w", err)
	}
	p.contentHash = fmt.Sprintf("sha256:%x", sha256.Sum256(canonical))
	return &p, nil
}

func (p *Pack) validateEvents() error {
	if p.Events.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema_version %q", p.Events.SchemaVersion)
	}
	if p.Events.BossSurrenderEvents == nil {
		return errors.New("boss_surrender_events must be present")
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
		for _, rec := range []int{e.DialogueRecords.Intro, e.DialogueRecords.Apology,
			e.DialogueRecords.Accept, e.DialogueRecords.Reject} {
			if rec < 0 {
				return fmt.Errorf("%s: dialogue records must not be negative", e.ID)
			}
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

func (p *Pack) ID() string             { return p.Manifest.PackID }
func (p *Pack) Schema() string         { return p.Manifest.SchemaVersion }
func (p *Pack) ContentHash() string    { return p.contentHash }
func (p *Pack) ContentVersion() string { return p.Manifest.ContentVersion }
