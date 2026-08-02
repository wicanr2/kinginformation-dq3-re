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
	SchemaVersion = "0.1.19"
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

// WindowLayout 是版本專屬的固定視窗幾何。引擎只依此通用契約繪製，
// 不保存任何 DQ3 專屬座標或尺寸。
type WindowLayout struct {
	ID           string   `json:"id"`
	X            int      `json:"x"`
	Y            int      `json:"y"`
	Width        int      `json:"width"`
	Height       int      `json:"height"`
	TextInsetX   int      `json:"text_inset_x"`
	TextInsetY   int      `json:"text_inset_y"`
	Columns      int      `json:"columns"`
	LinesPerPage int      `json:"lines_per_page"`
	Evidence     Evidence `json:"evidence"`
}

type Interface struct {
	SchemaVersion string       `json:"schema_version"`
	Dialogue      WindowLayout `json:"dialogue"`
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
	Success string `json:"success"`
	After   string `json:"after"`
}

// NPCItemRewardEvent describes a finite scripted-NPC transaction. The engine
// searches the party's personal inventories in party order and only clears
// PresentFlagRaw after the item was actually stored.
type NPCItemRewardEvent struct {
	ID              string               `json:"id"`
	Kind            string               `json:"kind"`
	NPC             ScriptedNPCSelector  `json:"npc"`
	PresentFlagRaw  int                  `json:"present_flag_raw"`
	GrantedItemRaw  int                  `json:"granted_item_raw"`
	DialogueTextIDs NPCItemRewardTextIDs `json:"dialogue_text_ids"`
	Evidence        Evidence             `json:"evidence"`
}

// ItemUseEffect 把版本專屬道具 ID 綁定至有限的引擎 primitive。
// JSON 只提供選擇器與參數，不接受可執行運算式。
type ItemUseEffect struct {
	ID                      string          `json:"id"`
	Kind                    string          `json:"kind"`
	ItemRawID               int             `json:"item_raw_id"`
	EffectID                string          `json:"effect_id"`
	LocationKind            string          `json:"location_kind"`
	DayNightPhase           int             `json:"day_night_phase"`
	ResetDayNightSteps      bool            `json:"reset_day_night_steps"`
	StepCount               int             `json:"step_count"`
	Consume                 bool            `json:"consume"`
	RequiredLayer           int             `json:"required_layer,omitempty"`
	RequiredVehicle         string          `json:"required_vehicle,omitempty"`
	UseTile                 *TileCoordinate `json:"use_tile,omitempty"`
	RequiredStoryFlagRaw    int             `json:"required_story_flag_raw,omitempty"`
	RequiredStoryFlagSet    bool            `json:"required_story_flag_set,omitempty"`
	SetWorldStateMask       int             `json:"set_world_state_mask,omitempty"`
	ClearStoryFlagsRaw      []int           `json:"clear_story_flags_raw,omitempty"`
	MapPatch                *WorldMapPatch  `json:"map_patch,omitempty"`
	AnimationPaletteModeRaw int             `json:"animation_palette_mode_raw,omitempty"`
	AnimationCycles         int             `json:"animation_cycles,omitempty"`
	DoorKeyTier             int             `json:"door_key_tier,omitempty"`
	Evidence                Evidence        `json:"evidence"`
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
	SuccessWorldPosition VehicleWorldPosition      `json:"success_world_position"`
	SetWorldStateMaskRaw int                       `json:"set_world_state_mask_raw"`
	ClearStoryFlagsRaw   []int                     `json:"clear_story_flags_raw"`
	DialogueTextIDs      ChoiceItemExchangeTextIDs `json:"dialogue_text_ids"`
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

type Events struct {
	SchemaVersion               string                           `json:"schema_version"`
	ItemActions                 ItemActions                      `json:"item_actions"`
	RuraNavigation              RuraNavigation                   `json:"rura_navigation"`
	WorldEntranceVariants       []WorldEntranceVariant           `json:"world_entrance_variants"`
	SettlementFounderEvents     []SettlementFounderEvent         `json:"settlement_founder_events"`
	SettlementFounderFollowups  []SettlementFounderFollowupEvent `json:"settlement_founder_followups"`
	NPCPushRule                 NPCPushRule                      `json:"npc_push_rule"`
	BossSurrenderEvents         []BossSurrenderEvent             `json:"boss_surrender_events"`
	TemporaryRoleEvents         []TemporaryRoleEvent             `json:"temporary_role_events"`
	TemporarySoloChallenges     []TemporarySoloChallengeEvent    `json:"temporary_solo_challenges"`
	QuestItemChainEvents        []QuestItemChainEvent            `json:"quest_item_chain_events"`
	ChoiceItemExchangeEvents    []ChoiceItemExchangeEvent        `json:"choice_item_exchange_events"`
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
	Manifest                   Manifest
	Facilities                 Facilities
	Interface                  Interface
	Events                     Events
	Characters                 Characters
	Texts                      Texts
	services                   map[string]*ServiceDefinition
	bossEvents                 map[string]*BossSurrenderEvent
	roleEvents                 map[string]*TemporaryRoleEvent
	soloChallenges             map[string]*TemporarySoloChallengeEvent
	questItemEvents            map[string]*QuestItemChainEvent
	choiceItemExchangeEvents   map[string]*ChoiceItemExchangeEvent
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
	if err := p.validateEventTextRefs(); err != nil {
		return nil, fmt.Errorf("%s: %w", eventsPath, err)
	}
	canonical, err := json.Marshal(struct {
		Manifest   Manifest   `json:"manifest"`
		Facilities Facilities `json:"facilities"`
		Interface  Interface  `json:"interface"`
		Events     Events     `json:"events"`
		Characters Characters `json:"characters"`
		Texts      Texts      `json:"texts"`
	}{p.Manifest, p.Facilities, p.Interface, p.Events, p.Characters, p.Texts})
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
		int(raw[1]) != formation.BackgroundRaw || int(raw[2]) != formation.PageRaw {
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
			e.SuccessWorldPosition.X < 0 || e.SuccessWorldPosition.Y < 0 ||
			e.SuccessWorldPosition.Layer < 0 || e.SuccessWorldPosition.Layer > 1 ||
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
				e.RequiredVehicle != "ship" || e.UseTile == nil ||
				e.UseTile.X < 0 || e.UseTile.Y < 0 ||
				e.RequiredStoryFlagRaw < 0 || e.RequiredStoryFlagRaw >= 512 ||
				e.RequiredStoryFlagSet || e.SetWorldStateMask <= 0 ||
				e.SetWorldStateMask > 0xffff || e.Consume ||
				patch == nil || patch.Layer != e.RequiredLayer ||
				patch.Origin.X < 0 || patch.Origin.Y < 0 ||
				patch.Width <= 0 || patch.Height <= 0 ||
				len(patch.TilesRaw) != patch.Width*patch.Height ||
				e.AnimationPaletteModeRaw < 0 || e.AnimationPaletteModeRaw > 255 ||
				e.AnimationCycles <= 0 {
				return fmt.Errorf("%s: invalid reveal_world_map_patch configuration", e.ID)
			}
			for _, tile := range patch.TilesRaw {
				if tile < 0 || tile > 255 {
					return fmt.Errorf("%s: map patch tile out of range", e.ID)
				}
			}
			foundClear := false
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
				e.MapPatch != nil {
				return fmt.Errorf("%s: invalid open_facing_locked_door configuration", e.ID)
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
