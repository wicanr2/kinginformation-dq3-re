package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"github.com/wicanr2/dq3_remake_ebitan/internal/itemstore"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 存檔(冒險之書):持久化主角進度 + 位置。Go port 自有格式(非 C 存檔二進位相容),
// 因 remake 是重表達;只求本機讀寫一致(round-trip)。教會/記錄點觸發存檔。
const saveFormatVersion = 2

type saveState struct {
	FormatVersion                  int             `json:"save_version"`
	Items                          json.RawMessage `json:"items"`
	itemStore                      itemstore.Store
	DeferredRegionDialogueReturnID string                   `json:"deferred_region_dialogue_return_id,omitempty"`
	OpeningHomeAwait               bool                     `json:"opening_home_await,omitempty"`
	PackID                         string                   `json:"pack_id,omitempty"`
	PackSchema                     string                   `json:"pack_schema,omitempty"`
	PackContentHash                string                   `json:"pack_content_hash,omitempty"`
	HeroExp                        uint32                   `json:"exp"`
	HeroHP                         int                      `json:"hp"`
	HeroMP                         int                      `json:"mp"`
	HeroConditions                 conditionSet             `json:"conditions,omitempty"`
	ParalysisSteps                 int                      `json:"paralysis_steps,omitempty"`
	HeroStat                       stats.Values             `json:"stats,omitempty"`
	HeroGold                       int                      `json:"gold"`
	HeroName                       []int                    `json:"heroname,omitempty"` // 主角姓名(glyph index;newgame.go 命名創建)
	HeroGender                     int                      `json:"herogender"`         // 0=男 1=女
	Comps                          []compSav                `json:"comps"`
	SoloChallengeActive            bool                     `json:"solo_challenge_active,omitempty"`
	SoloChallengeEventID           string                   `json:"solo_challenge_event_id,omitempty"`
	SoloChallengeCompanions        []compSav                `json:"solo_challenge_companions,omitempty"`
	Roster                         []compSav                `json:"roster,omitempty"` // 酒場名冊(未必在隊伍中的角色;見 recruit.go)
	SharedStorage                  []int                    `json:"shared_storage,omitempty"`
	SettlementFounder              *compSav                 `json:"settlement_founder,omitempty"`
	Flags                          []int                    `json:"flags"`
	ShipOwned                      bool                     `json:"shipowned"`
	PhoenixOwned                   bool                     `json:"phoenixowned,omitempty"`
	PhoenixAboard                  bool                     `json:"phoenixaboard,omitempty"`
	PhoenixX                       int                      `json:"phoenixx,omitempty"`
	PhoenixY                       int                      `json:"phoenixy,omitempty"`
	PhoenixMapCellWritten          bool                     `json:"phoenix_map_cell_written,omitempty"`
	ShipX                          int                      `json:"shipx"`
	ShipY                          int                      `json:"shipy"`
	EncounterStep                  int                      `json:"encounter_step,omitempty"`
	PX                             int                      `json:"px"`
	PY                             int                      `json:"py"`
	OverPX                         int                      `json:"over_px,omitempty"`
	OverPY                         int                      `json:"over_py,omitempty"`
	OverworldPosV2                 bool                     `json:"overworld_position_v2,omitempty"`
	InTown                         bool                     `json:"town"`
	StoryBits                      []byte                   `json:"storybits,omitempty"`
	WorldState                     uint16                   `json:"worldstate,omitempty"`
	TrackedWorldObjects            []trackedWorldObjectSave `json:"tracked_world_objects,omitempty"`
	DNPhase                        int                      `json:"dnphase,omitempty"`
	DNStep                         int                      `json:"dnstep,omitempty"`
	Cty                            int                      `json:"cty,omitempty"`
	Section                        int                      `json:"section,omitempty"`
	Layer                          int                      `json:"layer,omitempty"`
	VisitedTowns                   []townVisit              `json:"visited_towns,omitempty"`
	Respawn                        *respawnSave             `json:"respawn,omitempty"`
	PartyLeader                    int                      `json:"party_leader,omitempty"`
}

type respawnSave struct {
	PX      int  `json:"px"`
	PY      int  `json:"py"`
	OverPX  int  `json:"over_px"`
	OverPY  int  `json:"over_py"`
	InTown  bool `json:"in_town"`
	Cty     int  `json:"cty"`
	Section int  `json:"section"`
	Layer   int  `json:"layer"`
}

func respawnToSave(r respawnPoint) *respawnSave {
	if !r.Valid {
		return nil
	}
	return &respawnSave{
		PX: r.PX, PY: r.PY, OverPX: r.OverPX, OverPY: r.OverPY,
		InTown: r.InTown, Cty: r.Cty, Section: r.Section, Layer: r.Layer,
	}
}

func respawnFromSave(r *respawnSave) respawnPoint {
	if r == nil {
		return respawnPoint{}
	}
	return respawnPoint{
		Valid: true, PX: r.PX, PY: r.PY, OverPX: r.OverPX, OverPY: r.OverPY,
		InTown: r.InTown, Cty: r.Cty, Section: r.Section, Layer: r.Layer,
	}
}

type trackedWorldObjectSave struct {
	ID    string `json:"id"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Layer int    `json:"layer"`
}

// compSav 是一名同伴的存檔資料。
type compSav struct {
	Items         json.RawMessage `json:"items"`
	itemStore     itemstore.Store
	Name          []int `json:"name,omitempty"`
	Class, Gender int
	Exp           uint32
	Stats         stats.Values `json:"stats,omitempty"`
	CurHP, CurMP  int
	LearnedSpells []int        `json:"learned_spells,omitempty"`
	Conditions    conditionSet `json:"conditions,omitempty"`
}

func encodeSave(s saveState) ([]byte, error) { return json.Marshal(s) }

// strictSaveJSON validates exact struct keys, duplicates and every nested owner.
func strictSaveJSON(raw []byte, t reflect.Type) error {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == reflect.TypeOf(json.RawMessage{}) {
		return nil
	}
	if t.Kind() == reflect.Struct {
		dec := json.NewDecoder(bytes.NewReader(raw))
		token, err := dec.Token()
		if err != nil || token != json.Delim('{') {
			return fmt.Errorf("save object required")
		}
		fields := map[string]reflect.Type{}
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if f.PkgPath != "" {
				continue
			}
			name := strings.Split(f.Tag.Get("json"), ",")[0]
			if name == "-" {
				continue
			}
			if name == "" {
				name = f.Name
			}
			fields[name] = f.Type
		}
		seen := map[string]bool{}
		for dec.More() {
			token, err := dec.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			ft, exists := fields[key]
			if !ok || !exists || seen[key] {
				return fmt.Errorf("unknown or duplicate save field %q", key)
			}
			seen[key] = true
			var value json.RawMessage
			if err := dec.Decode(&value); err != nil {
				return err
			}
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				if ft.Kind() != reflect.Pointer && ft.Kind() != reflect.Slice {
					return fmt.Errorf("null save field %q", key)
				}
				continue
			}
			if err := strictSaveJSON(value, ft); err != nil {
				return fmt.Errorf("%s: %w", key, err)
			}
		}
		if token, err := dec.Token(); err != nil || token != json.Delim('}') {
			return fmt.Errorf("unfinished save object")
		}
		var trailing any
		if dec.Decode(&trailing) != io.EOF {
			return fmt.Errorf("trailing save data")
		}
	} else if (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) && t.Elem().Kind() == reflect.Struct {
		var rows []json.RawMessage
		if err := json.Unmarshal(raw, &rows); err != nil {
			return err
		}
		for _, row := range rows {
			if err := strictSaveJSON(row, t.Elem()); err != nil {
				return err
			}
		}
	}
	return nil
}
func decodeSave(b []byte) (saveState, error) {
	var s saveState
	if err := strictSaveJSON(b, reflect.TypeOf(s)); err != nil {
		return s, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return s, err
	}
	if s.FormatVersion != saveFormatVersion || s.PackID == "" || s.PackSchema == "" || s.PackContentHash == "" {
		return s, fmt.Errorf("unsupported save format or missing pack identity")
	}
	return s, nil
}
func encodedItemStore(store itemstore.Store) json.RawMessage { raw, _ := store.Encode(); return raw }
func (g *Game) validateSavedItemStores(s *saveState) error {
	if g.pack == nil {
		return fmt.Errorf("save requires a validated game pack")
	}
	decode := func(raw json.RawMessage, destination *itemstore.Store) error {
		v, err := g.pack.DecodeItemStore(raw)
		if err != nil {
			return err
		}
		eq := map[int]bool{}
		for _, e := range v.Worn() {
			if e.Part < 0 || eq[e.Part] {
				return fmt.Errorf("invalid saved equipment reference")
			}
			eq[e.Part] = true
		}
		*destination = v
		return nil
	}
	if err := decode(s.Items, &s.itemStore); err != nil {
		return fmt.Errorf("hero items: %w", err)
	}
	for _, group := range [][]compSav{s.Comps, s.Roster, s.SoloChallengeCompanions} {
		for i := range group {
			if err := decode(group[i].Items, &group[i].itemStore); err != nil {
				return fmt.Errorf("member %d items: %w", i, err)
			}
		}
	}
	if s.SettlementFounder != nil {
		if err := decode(s.SettlementFounder.Items, &s.SettlementFounder.itemStore); err != nil {
			return fmt.Errorf("founder items: %w", err)
		}
	}
	return nil
}

func (g *Game) snapshot() saveState {
	var packID, packSchema, packHash string
	if g.pack != nil { // 裸 Game 單元測試／舊 migration fixture 沒有 loader；production 永遠非 nil。
		packID, packSchema, packHash = g.pack.ID(), g.pack.Schema(), g.pack.ContentHash()
	}
	var founder *compSav
	if g.settlementFounder != nil {
		saved := compsToSav([]*Member{g.settlementFounder})[0]
		founder = &saved
	}
	s := saveState{
		FormatVersion: saveFormatVersion, Items: encodedItemStore(g.items), itemStore: g.items.Clone(),
		DeferredRegionDialogueReturnID: g.deferredRegionDialogueReturnID,
		PackID:                         packID, PackSchema: packSchema, PackContentHash: packHash, OpeningHomeAwait: g.homeAwait,
		HeroExp: g.heroExp, HeroHP: g.heroHP, HeroMP: g.heroMP, HeroConditions: g.heroConditions,
		ParalysisSteps: g.paralysisSteps,
		HeroStat:       g.heroStat, HeroGold: g.heroGold,
		HeroName: append([]int(nil), g.heroName...), HeroGender: g.heroGender,
		Comps:                   compsToSav(g.companions),
		SoloChallengeActive:     g.soloChallengeActive,
		SoloChallengeEventID:    g.soloChallengeEventID,
		SoloChallengeCompanions: compsToSav(g.soloChallengeCompanions),
		Roster:                  compsToSav(g.roster),
		SharedStorage:           append([]int(nil), g.sharedStorage...),
		SettlementFounder:       founder,
		Flags:                   flagsToSav(g.flags),
		ShipOwned:               g.shipOwned, ShipX: g.shipX, ShipY: g.shipY,
		PhoenixOwned: g.phoenixOwned, PhoenixAboard: g.phoenixAboard,
		PhoenixX: g.phoenixX, PhoenixY: g.phoenixY, PhoenixMapCellWritten: g.phoenixMapCellWritten,
		EncounterStep: g.encounterStep,
		PX:            g.px, PY: g.py, OverPX: g.overPx, OverPY: g.overPy,
		OverworldPosV2: true, InTown: g.inTown,
		StoryBits:           append([]byte(nil), g.storyBits[:]...),
		WorldState:          g.worldState,
		TrackedWorldObjects: trackedWorldObjectsToSave(g.trackedWorldPositions),
		DNPhase:             g.dnPhase, DNStep: g.dnStep, Cty: g.curCty, Layer: g.layer,
		VisitedTowns: append([]townVisit(nil), g.visitedTowns...),
		Respawn:      respawnToSave(g.respawn), PartyLeader: g.partyLeader,
	}
	if g.inTown && g.cur != nil {
		s.Section = g.cur.sec
	}
	return s
}

func trackedWorldObjectsToSave(in map[string]trackedWorldPosition) []trackedWorldObjectSave {
	out := make([]trackedWorldObjectSave, 0, len(in))
	for id, p := range in {
		out = append(out, trackedWorldObjectSave{ID: id, X: p.X, Y: p.Y, Layer: p.Layer})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func flagsToSav(f map[int]bool) []int {
	var out []int
	for k, v := range f {
		if v {
			out = append(out, k)
		}
	}
	return out
}

func compsToSav(ms []*Member) []compSav {
	out := make([]compSav, len(ms))
	for i, m := range ms {
		m.ensureStats()
		out[i] = compSav{Name: append([]int(nil), m.Name...), Class: m.Class, Gender: m.Gender,
			Exp: m.Exp, Stats: m.Stats, CurHP: m.CurHP, CurMP: m.CurMP,
			Items: encodedItemStore(m.Items), itemStore: m.Items.Clone(),
			LearnedSpells: append([]int(nil), m.LearnedSpells...), Conditions: m.Conditions}
	}
	return out
}

func (g *Game) restore(s saveState) error {
	if s.FormatVersion != saveFormatVersion || g.pack == nil || s.PackID != g.pack.ID() || s.PackSchema != g.pack.Schema() || s.PackContentHash != g.pack.ContentHash() {
		return fmt.Errorf("unsupported save format or game pack mismatch")
	}
	if err := g.validateSavedItemStores(&s); err != nil {
		return err
	}

	g.tavern.reset()
	if g.recruit.active && (g.recruit.stage == rcFinishText || g.recruit.stage == rcMusicWait) {
		g.finishRecruitmentMusic()
	}
	g.recruit.reset()
	if g.fieldSaveLoad.active && g.fieldSaveLoad.stage == fsSoundWait {
		if a, ok := g.music.(backgroundAudio); ok {
			a.ResumeBackground()
		}
	}
	g.fieldSaveLoad.reset()
	g.deferredRegionDialogueReturnID = s.DeferredRegionDialogueReturnID
	if g.regionDialogueReturn != nil {
		g.regionDialogueReturn = nil
		g.dlg.open = false
		g.clearOpeningPresentation()
	}
	if g.regionDialogueReward != nil {
		g.regionDialogueReward = nil
		g.dlg.open = false
		g.clearOpeningPresentation()
	}
	g.fieldIdle = fieldIdleState{}
	g.itemGivePrompt = nil
	g.fieldMessagePrompt = nil
	g.fieldEquipment = fieldEquipmentState{}
	// Go 冒險之書不保存暫態場景咒文 timer；restore 必須清掉同一 Game instance
	// 載入前的透明效果。原版 save 是否序列化這類 timer 仍需獨立 RE。
	g.remoaru = 0
	g.toramana, g.hazardGuard = false, false
	g.heroExp, g.heroHP, g.heroMP, g.heroStat, g.heroGold =
		s.HeroExp, s.HeroHP, s.HeroMP, s.HeroStat, s.HeroGold
	g.heroConditions = s.HeroConditions
	g.paralysisSteps = s.ParalysisSteps
	g.heroName, g.heroGender = append([]int(nil), s.HeroName...), s.HeroGender
	g.heroInit = true
	g.ensureHeroStats() // 舊版 JSON 沒有 stats 欄時，以當前等級 target 安全升級
	if g.heroHP > int(g.heroStat[stats.HP]) {
		g.heroHP = int(g.heroStat[stats.HP])
	}
	if g.heroMP > int(g.heroStat[stats.MP]) {
		g.heroMP = int(g.heroStat[stats.MP])
	}
	g.items = s.itemStore.Clone()
	g.flags = map[int]bool{}
	for _, k := range s.Flags {
		g.flags[k] = true
	}
	g.companions = nil
	if len(s.Comps) > 0 { // 還原同伴
		g.companions = make([]*Member, len(s.Comps))
		for i, c := range s.Comps {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Items:         c.itemStore.Clone(),
				LearnedSpells: append([]int(nil), c.LearnedSpells...), Conditions: c.Conditions}
			if len(m.Name) == 0 && c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			m.ensureStats()
			m.syncLearnedSpells()
			g.companions[i] = m
		}
	}
	g.soloChallengeActive = s.SoloChallengeActive
	g.soloChallengeEventID = s.SoloChallengeEventID
	g.soloChallengeCompanions = nil
	if len(s.SoloChallengeCompanions) > 0 {
		g.soloChallengeCompanions = make([]*Member, len(s.SoloChallengeCompanions))
		for i, c := range s.SoloChallengeCompanions {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Items: c.itemStore.Clone(), LearnedSpells: append([]int(nil), c.LearnedSpells...),
				Conditions: c.Conditions}
			m.ensureStats()
			m.syncLearnedSpells()
			g.soloChallengeCompanions[i] = m
		}
	}
	if g.soloChallengeActive {
		if g.pack == nil {
			g.soloChallengeActive, g.soloChallengeEventID, g.soloChallengeCompanions = false, "", nil
		} else if _, ok := g.pack.TemporarySoloChallenge(g.soloChallengeEventID); !ok {
			// A pack-bound save must never restore an unknown partial transaction.
			g.soloChallengeActive, g.soloChallengeEventID, g.soloChallengeCompanions = false, "", nil
		}
	}
	g.roster = nil
	if len(s.Roster) > 0 { // 還原名冊(未入隊角色)
		g.roster = make([]*Member, len(s.Roster))
		for i, c := range s.Roster {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Items:         c.itemStore.Clone(),
				LearnedSpells: append([]int(nil), c.LearnedSpells...), Conditions: c.Conditions}
			if len(m.Name) == 0 && c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			m.ensureStats()
			m.syncLearnedSpells()
			g.roster[i] = m
		}
	}
	g.sharedStorage = append([]int(nil), s.SharedStorage...)
	g.settlementFounder = nil
	if c := s.SettlementFounder; c != nil {
		m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
			Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
			Items: c.itemStore.Clone(), LearnedSpells: append([]int(nil), c.LearnedSpells...),
			Conditions: c.Conditions}
		m.ensureStats()
		m.syncLearnedSpells()
		g.settlementFounder = m
	}
	g.shipOwned, g.shipX, g.shipY = s.ShipOwned, s.ShipX, s.ShipY
	g.phoenixOwned, g.phoenixAboard = s.PhoenixOwned, s.PhoenixAboard
	g.phoenixX, g.phoenixY = s.PhoenixX, s.PhoenixY
	g.phoenixMapCellWritten = s.PhoenixMapCellWritten
	g.encounterStep = s.EncounterStep
	g.initStoryBits()
	if len(s.StoryBits) > 0 {
		copy(g.storyBits[:], s.StoryBits) // 32-byte 舊檔保留前 256 flags；高位沿用原版初值
	}
	g.syncTemporaryRoleVisual() // 角色外觀由 pack event 的 active flag 推導，不保存第二份狀態
	g.worldState = s.WorldState
	g.worldObjectBufferValid = false
	g.coordinateItemGateID = ""
	g.coordinateItemGateStage = coordinateGateIdle
	g.coordinateForcedSteps = 0
	g.coordinateForcedDir = 0
	g.trackedWorldPositions = map[string]trackedWorldPosition{}
	for _, saved := range s.TrackedWorldObjects {
		if g.pack == nil {
			continue
		}
		obj, ok := g.pack.TrackedWorldObject(saved.ID)
		if !ok || obj.Layer != saved.Layer || saved.X < 0 || saved.Y < 0 {
			continue
		}
		g.trackedWorldPositions[saved.ID] = trackedWorldPosition{X: saved.X, Y: saved.Y, Layer: saved.Layer}
	}
	// 0.1.19 的短暫錯誤實作把動態物件座標寫進 over_px/over_py。若 world-state
	// 已啟用但新欄位尚不存在，僅由 pack 中明確的 activation 記錄遷移；不猜預設。
	if len(g.trackedWorldPositions) == 0 && g.pack != nil {
		for _, event := range g.pack.ChoiceItemExchangeEvents() {
			if g.worldState&uint16(event.SetWorldStateMaskRaw) != 0 {
				g.activateTrackedWorldObject(event.ActivateWorldObject)
			}
		}
	}
	if g.flags[0x35] { // 舊 remake 存檔：自造 flag0x35 遷移至原版 [0x4f44] bit0x40。
		g.worldState |= worldStateRainbowBridge
		delete(g.flags, 0x35)
	}
	g.dnPhase, g.dnStep = s.DNPhase&3, s.DNStep
	g.respawn = respawnFromSave(s.Respawn)
	g.partyLeader = s.PartyLeader
	if g.partyLeader < 0 || g.partyLeader > len(g.companions) {
		g.partyLeader = 0
	}
	g.layer, g.curCty = s.Layer, s.Cty
	switch {
	case s.OverworldPosV2:
		g.overPx, g.overPy = s.OverPX, s.OverPY
	case s.InTown && s.Cty >= 0 && s.Cty < len(ctyLoc) &&
		ctyLoc[s.Cty][2] == s.Layer:
		// 舊 Go 存檔沒有 remembered-world 欄位。只能由已保存的 CTY/layer
		// 回推該城原版 cty_loc；不可沿用 NewGame 的阿里阿罕預設。
		g.overPx, g.overPy = ctyLoc[s.Cty][0], ctyLoc[s.Cty][1]
	default:
		// 地表舊檔的玩家座標本身就是 remembered-world 座標。
		g.overPx, g.overPy = s.PX, s.PY
	}
	g.visitedTowns = nil
	for _, v := range s.VisitedTowns {
		g.addVisitedTown(v.Cty) // 驗證並按 EXE table 正規化舊存檔順序。
	}
	g.px, g.py, g.inTown = s.PX, s.PY, s.InTown
	g.resetPartyTrail()
	if s.InTown {
		g.restoreTownScene(s.Cty, s.Section)
		g.applyPhoenixMapCell()
	} else {
		g.cur = g.overworldScene()
		g.curCty = -1
	}
	// 舊版存檔沒有 VisitedTowns。至少依可證明的起點與目前所在城鎮補遷移，
	// 不猜測玩家曾去過哪些其他城鎮。
	if len(g.visitedTowns) == 0 {
		g.addVisitedTown(0)
		if g.inTown {
			g.rememberTown()
		}
	}
	g.homeSelection.active = false
	g.openingIdx = -1
	g.openingEscortPhase = -1
	g.openingEscortIndex = -1
	g.homeAwait = false
	if s.OpeningHomeAwait && g.openingEscort != nil && g.inTown && g.curCty == g.openingEscort.CTY && g.cur.sec == g.openingEscort.Section {
		last := g.openingEscort.Frames[len(g.openingEscort.Frames)-1]
		for i, n := range g.cur.npcs {
			if n.recordIndex == *g.openingEscort.Home.LeaderRecord {
				g.openingEscortNPC = i
				g.cur.npcs[i].x = last.Leader.X
				g.cur.npcs[i].y = last.Leader.Y
				g.cur.npcs[i].facing = last.LeaderFacing
				g.homeAwait = true
				break
			}
		}
	}
	g.applyRainbowBridge()
	g.applyPackWorldMapPatches()
	g.applyDaynightPalette()
	if !g.respawn.Valid { // 舊 Go 存檔：loaded position 是唯一可證明的記錄點。
		g.respawn = g.currentRespawnPoint()
	}
	g.selectLivingPartyLeader()
	g.dlg.heroName = append([]int(nil), g.heroName...)
	return nil
}

// restoreTownScene 以存檔的 CTY/section/daynight 重建 NPC 狀態；舊存檔的零值自然落在
// 阿里阿罕 section0。測試或極早初始化若尚無 assets，才退回既有 town 指標。
func (g *Game) restoreTownScene(cty, section int) {
	if cty < 0 || cty >= 100 || g.assets == nil {
		g.cur = g.town
		return
	}
	blkn := 1
	if cty < len(mapBlkNum) {
		blkn = mapBlkNum[cty]
	}
	ns, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, cty, blkn, section, g.dnPhase, g.storyFlag)
	if err != nil {
		g.cur = g.town
		return
	}
	g.town, g.cur, g.curCty = ns, ns, cty
	g.resetPartyTrail()
	if g.towns == nil {
		g.towns = map[int]*Scene{}
	}
	g.towns[cty] = ns
	if ns.dlgText != nil {
		g.dlg.tx = ns.dlgText
	}
}

// savePath:DQ3_SAVE 指定的檔;空 → CWD/dq3save.json。
func savePath() string {
	if p := os.Getenv("DQ3_SAVE"); p != "" {
		return p
	}
	return "dq3save.json"
}

// Save 寫存檔。
func (g *Game) Save() error {
	return g.saveTo(savePath())
}

func (g *Game) saveTo(path string) error {
	if g.regionDialogueReturn != nil {
		return fmt.Errorf("region dialogue return is not at a save checkpoint")
	}
	if g.homeSelection.active || (g.homeAwait && g.openingEscortPhase == 5) {
		return fmt.Errorf("opening home transaction is not at a save checkpoint")
	}
	point := g.currentRespawnPoint()
	s := g.snapshot()
	s.Respawn = respawnToSave(point)
	if s.FormatVersion != saveFormatVersion || s.PackID == "" {
		return fmt.Errorf("save requires current format and pack identity")
	}
	if err := g.validateSavedItemStores(&s); err != nil {
		return err
	}
	b, err := encodeSave(s)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return err
	}
	g.respawn = point
	return nil
}

// Load 讀存檔(不存在 → 靜默略過,回 nil)。
func (g *Game) Load() error {
	return g.loadFrom(savePath(), true)
}

func (g *Game) loadFrom(path string, ignoreMissing bool) error {
	return g.loadSnapshotFile(path, ignoreMissing, nil)
}

func (g *Game) loadSnapshotFile(path string, ignoreMissing bool, fieldClock *gamepack.FieldSaveLoad) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if ignoreMissing && os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s, err := decodeSave(b)
	if err != nil {
		return err
	}
	// The format gate rejects every old save, including absent pack metadata.
	if g.pack == nil || s.PackID != g.pack.ID() || s.PackSchema != g.pack.Schema() ||
		s.PackContentHash != g.pack.ContentHash() {
		if g.pack == nil {
			return fmt.Errorf("save requires game pack %s/%s/%s but current Game has none",
				s.PackID, s.PackSchema, s.PackContentHash)
		}
		return fmt.Errorf("save game pack mismatch: save=%s/%s/%s current=%s/%s/%s",
			s.PackID, s.PackSchema, s.PackContentHash,
			g.pack.ID(), g.pack.Schema(), g.pack.ContentHash())
	}
	if err := g.validateSavedItemStores(&s); err != nil {
		return err
	}
	if s.OpeningHomeAwait {
		e := g.openingEscort
		if e == nil || s.PackID == "" || !s.InTown || s.Cty != e.CTY || s.Section != e.Section || s.PX < 0 || s.PY < 0 {
			return fmt.Errorf("opening home save checkpoint invalid")
		}
		flag := e.Home.ApproachFlag
		if flag/8 >= len(s.StoryBits) || s.StoryBits[flag/8]&(128>>uint(flag%8)) == 0 {
			return fmt.Errorf("opening home save gate missing")
		}
		if s.Cty >= len(mapBlkNum) {
			return fmt.Errorf("opening home save scene unavailable")
		}
		scene, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, s.Cty, mapBlkNum[s.Cty], s.Section, s.DNPhase, func(flag int) bool {
			return flag >= 0 && flag/8 < len(s.StoryBits) && s.StoryBits[flag/8]&(128>>uint(flag%8)) != 0
		})
		if err != nil {
			return fmt.Errorf("opening home save scene: %w", err)
		}
		last := e.Frames[len(e.Frames)-1]
		if s.PX >= scene.w || s.PY >= scene.h || last.Leader.X >= scene.w || last.Leader.Y >= scene.h {
			return fmt.Errorf("opening home save position outside scene")
		}
		found := false
		for _, actor := range scene.npcs {
			if actor.recordIndex == *e.Home.LeaderRecord {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("opening home save actor missing")
		}
	}
	if fieldClock != nil {
		clock, ok := fieldClock.LoadClock(s.Layer)
		phaseTicks := g.dayNightCycle.ClockTicks / 4
		if !ok || phaseTicks <= 0 || clock < 0 || clock >= g.dayNightCycle.ClockTicks {
			return fmt.Errorf("field load clock rule unavailable for layer %d", s.Layer)
		}
		s.DNPhase, s.DNStep = clock/phaseTicks, clock%phaseTicks
	}
	if err := g.validateDeferredRegionDialogueReturnSave(s); err != nil {
		return err
	}
	returnActor, err := g.loadRegionDialogueReturnActor(s)
	if err != nil {
		return err
	}
	if s.DeferredRegionDialogueReturnID != "" && returnActor == nil {
		return fmt.Errorf("deferred region return save lacks its completed escort actor")
	}
	if err := g.restore(s); err != nil {
		return err
	}
	if returnActor != nil && g.cur != nil {
		found := false
		for _, n := range g.cur.npcs {
			if n.recordIndex == returnActor.recordIndex {
				found = true
				break
			}
		}
		if !found {
			index := 0
			for index < len(g.cur.npcs) && g.cur.npcs[index].recordIndex < returnActor.recordIndex {
				index++
			}
			g.cur.npcs = append(g.cur.npcs, npcInst{})
			copy(g.cur.npcs[index+1:], g.cur.npcs[index:])
			g.cur.npcs[index] = *returnActor
		}
	}
	return nil
}
