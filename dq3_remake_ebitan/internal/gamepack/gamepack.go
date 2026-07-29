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
	SchemaVersion = "0.1.8"
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

// ItemUseEffect 把版本專屬道具 ID 綁定至有限的引擎 primitive。
// JSON 只提供選擇器與參數，不接受可執行運算式。
type ItemUseEffect struct {
	ID                 string   `json:"id"`
	Kind               string   `json:"kind"`
	ItemRawID          int      `json:"item_raw_id"`
	EffectID           string   `json:"effect_id"`
	LocationKind       string   `json:"location_kind"`
	DayNightPhase      int      `json:"day_night_phase"`
	ResetDayNightSteps bool     `json:"reset_day_night_steps"`
	Consume            bool     `json:"consume"`
	Evidence           Evidence `json:"evidence"`
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

type Events struct {
	SchemaVersion               string                       `json:"schema_version"`
	ItemActions                 ItemActions                  `json:"item_actions"`
	BossSurrenderEvents         []BossSurrenderEvent         `json:"boss_surrender_events"`
	TemporaryRoleEvents         []TemporaryRoleEvent         `json:"temporary_role_events"`
	QuestItemChainEvents        []QuestItemChainEvent        `json:"quest_item_chain_events"`
	TreasureEvents              []TreasureEvent              `json:"treasure_events"`
	ItemUseEffects              []ItemUseEffect              `json:"item_use_effects"`
	TwoStepFloorSwitchGates     []TwoStepFloorSwitchGate     `json:"two_step_floor_switch_gates"`
	StagedVehicleExchangeEvents []StagedVehicleExchangeEvent `json:"staged_vehicle_exchange_events"`
	GuidedPassageEvents         []GuidedPassageEvent         `json:"guided_passage_events"`
	HostageRescueEvents         []HostageRescueEvent         `json:"hostage_rescue_events"`
	ReclassEvents               []ReclassEvent               `json:"reclass_events"`
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
	Manifest         Manifest
	Facilities       Facilities
	Events           Events
	Characters       Characters
	Texts            Texts
	services         map[string]*ServiceDefinition
	bossEvents       map[string]*BossSurrenderEvent
	roleEvents       map[string]*TemporaryRoleEvent
	questItemEvents  map[string]*QuestItemChainEvent
	treasureEvents   map[string]*TreasureEvent
	itemUseEffects   map[int]*ItemUseEffect
	sequenceGates    map[string]*TwoStepFloorSwitchGate
	vehicleExchanges map[string]*StagedVehicleExchangeEvent
	guidedPassages   map[string]*GuidedPassageEvent
	hostageRescues   map[string]*HostageRescueEvent
	reclassEvents    map[string]*ReclassEvent
	charDefaults     map[string]*CharacterDefault
	texts            map[string]*TextDefinition
	contentHash      string
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
		case "battle_message":
			if d.Source.Kind != "legacy_record" || d.Source.File == "" || d.Source.Record == nil {
				return fmt.Errorf("%s: battle_message requires legacy_record source", d.ID)
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
	return nil
}

func validScriptedNPC(s ScriptedNPCSelector) bool {
	return s.CTYRaw >= 0 && s.CTYRaw <= 255 && s.Section >= 0 &&
		s.Tile.X >= 0 && s.Tile.Y >= 0 && s.HandlerRaw >= 0 && s.HandlerRaw <= 255
}

func validTreasureSelector(t QuestTreasureSelector) bool {
	return t.CTYRaw >= 0 && t.CTYRaw <= 255 && t.Section >= 0 &&
		t.TileSubID >= 0 && t.TileSubID <= 31 &&
		(t.EventTypeRaw == 1 || t.EventTypeRaw == 3) &&
		t.ItemRawID >= 0 && t.ItemRawID <= 255 &&
		t.PresentFlag >= 0 && t.PresentFlag < 512
}

func (p *Pack) validateEvents() error {
	if p.Events.ItemActions.PersonalInventorySlots < 1 ||
		p.Events.ItemActions.PersonalInventorySlots > 32 {
		return errors.New("item_actions.personal_inventory_slots must be 1..32")
	}
	if err := validateEvidence(p.Events.ItemActions.Evidence); err != nil {
		return fmt.Errorf("item_actions evidence: %w", err)
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
	if p.Events.QuestItemChainEvents == nil {
		return errors.New("quest_item_chain_events must be present")
	}
	if p.Events.TreasureEvents == nil {
		return errors.New("treasure_events must be present")
	}
	if p.Events.ItemUseEffects == nil {
		return errors.New("item_use_effects must be present")
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
		if e.EffectID != "force_day_night_phase" ||
			e.LocationKind != "overworld" ||
			e.DayNightPhase < 0 || e.DayNightPhase > 3 ||
			!e.ResetDayNightSteps {
			return fmt.Errorf("%s: invalid force_day_night_phase configuration", e.ID)
		}
		if err := validateEvidence(e.Evidence); err != nil {
			return fmt.Errorf("%s evidence: %w", e.ID, err)
		}
		itemUseIDs[e.ID] = true
		p.itemUseEffects[e.ItemRawID] = e
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
				int(raw[1]) != formation.BackgroundRaw ||
				int(raw[2]) != formation.PageRaw {
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

func (p *Pack) QuestItemChainEvents() []QuestItemChainEvent {
	return append([]QuestItemChainEvent(nil), p.Events.QuestItemChainEvents...)
}

func (p *Pack) QuestItemChainEvent(id string) (*QuestItemChainEvent, bool) {
	e, ok := p.questItemEvents[id]
	return e, ok
}

func (p *Pack) TreasureEvents() []TreasureEvent {
	return append([]TreasureEvent(nil), p.Events.TreasureEvents...)
}

func (p *Pack) TreasureEvent(id string) (*TreasureEvent, bool) {
	e, ok := p.treasureEvents[id]
	return e, ok
}

func (p *Pack) ItemUseEffects() []ItemUseEffect {
	return append([]ItemUseEffect(nil), p.Events.ItemUseEffects...)
}

func (p *Pack) ItemUseEffectByRawID(itemRawID int) (*ItemUseEffect, bool) {
	e, ok := p.itemUseEffects[itemRawID]
	return e, ok
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
