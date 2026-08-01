package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 存檔(冒險之書):持久化主角進度 + 位置。Go port 自有格式(非 C 存檔二進位相容),
// 因 remake 是重表達;只求本機讀寫一致(round-trip)。教會/記錄點觸發存檔。
type saveState struct {
	PackID                  string       `json:"pack_id,omitempty"`
	PackSchema              string       `json:"pack_schema,omitempty"`
	PackContentHash         string       `json:"pack_content_hash,omitempty"`
	HeroExp                 uint32       `json:"exp"`
	HeroHP                  int          `json:"hp"`
	HeroMP                  int          `json:"mp"`
	HeroStat                stats.Values `json:"stats,omitempty"`
	HeroGold                int          `json:"gold"`
	HeroName                []int        `json:"heroname,omitempty"` // 主角姓名(glyph index;newgame.go 命名創建)
	HeroGender              int          `json:"herogender"`         // 0=男 1=女
	Inventory               []int        `json:"inv"`
	Equip                   [4]int       `json:"eq"`
	EquipmentV2             bool         `json:"equipment_v2,omitempty"` // true:-1=空；舊檔以 0=空
	Comps                   []compSav    `json:"comps"`
	SoloChallengeActive     bool         `json:"solo_challenge_active,omitempty"`
	SoloChallengeEventID    string       `json:"solo_challenge_event_id,omitempty"`
	SoloChallengeCompanions []compSav    `json:"solo_challenge_companions,omitempty"`
	Roster                  []compSav    `json:"roster,omitempty"` // 酒場名冊(未必在隊伍中的角色;見 recruit.go)
	Flags                   []int        `json:"flags"`
	ShipOwned               bool         `json:"shipowned"`
	PhoenixOwned            bool         `json:"phoenixowned,omitempty"`
	PhoenixAboard           bool         `json:"phoenixaboard,omitempty"`
	PhoenixX                int          `json:"phoenixx,omitempty"`
	PhoenixY                int          `json:"phoenixy,omitempty"`
	ShipX                   int          `json:"shipx"`
	ShipY                   int          `json:"shipy"`
	PX                      int          `json:"px"`
	PY                      int          `json:"py"`
	OverPX                  int          `json:"over_px,omitempty"`
	OverPY                  int          `json:"over_py,omitempty"`
	OverworldPosV2          bool         `json:"overworld_position_v2,omitempty"`
	InTown                  bool         `json:"town"`
	StoryBits               []byte       `json:"storybits,omitempty"`
	WorldState              uint16       `json:"worldstate,omitempty"`
	DNPhase                 int          `json:"dnphase,omitempty"`
	DNStep                  int          `json:"dnstep,omitempty"`
	Cty                     int          `json:"cty,omitempty"`
	Section                 int          `json:"section,omitempty"`
	Layer                   int          `json:"layer,omitempty"`
	VisitedTowns            []townVisit  `json:"visited_towns,omitempty"`
}

// compSav 是一名同伴的存檔資料。
type compSav struct {
	Name                        []int `json:"name,omitempty"`
	Class, Gender               int
	Exp                         uint32
	Stats                       stats.Values `json:"stats,omitempty"`
	CurHP, CurMP                int
	Weapon, Armor, Shield, Head int
	Inventory                   []int `json:"inventory,omitempty"`
	LearnedSpells               []int `json:"learned_spells,omitempty"`
}

func encodeSave(s saveState) ([]byte, error) { return json.Marshal(s) }

func decodeSave(b []byte) (saveState, error) {
	var s saveState
	err := json.Unmarshal(b, &s)
	return s, err
}

func (g *Game) snapshot() saveState {
	var packID, packSchema, packHash string
	if g.pack != nil { // 裸 Game 單元測試／舊 migration fixture 沒有 loader；production 永遠非 nil。
		packID, packSchema, packHash = g.pack.ID(), g.pack.Schema(), g.pack.ContentHash()
	}
	s := saveState{
		PackID: packID, PackSchema: packSchema, PackContentHash: packHash,
		HeroExp: g.heroExp, HeroHP: g.heroHP, HeroMP: g.heroMP, HeroStat: g.heroStat, HeroGold: g.heroGold,
		HeroName: append([]int(nil), g.heroName...), HeroGender: g.heroGender,
		Inventory:               append([]int(nil), g.inventory...),
		Equip:                   g.equip,
		EquipmentV2:             true,
		Comps:                   compsToSav(g.companions),
		SoloChallengeActive:     g.soloChallengeActive,
		SoloChallengeEventID:    g.soloChallengeEventID,
		SoloChallengeCompanions: compsToSav(g.soloChallengeCompanions),
		Roster:                  compsToSav(g.roster),
		Flags:                   flagsToSav(g.flags),
		ShipOwned:               g.shipOwned, ShipX: g.shipX, ShipY: g.shipY,
		PhoenixOwned: g.phoenixOwned, PhoenixAboard: g.phoenixAboard,
		PhoenixX: g.phoenixX, PhoenixY: g.phoenixY,
		PX: g.px, PY: g.py, OverPX: g.overPx, OverPY: g.overPy,
		OverworldPosV2: true, InTown: g.inTown,
		StoryBits:  append([]byte(nil), g.storyBits[:]...),
		WorldState: g.worldState,
		DNPhase:    g.dnPhase, DNStep: g.dnStep, Cty: g.curCty, Layer: g.layer,
		VisitedTowns: append([]townVisit(nil), g.visitedTowns...),
	}
	if g.inTown && g.cur != nil {
		s.Section = g.cur.sec
	}
	return s
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
			Weapon: m.Weapon, Armor: m.Armor, Shield: m.Shield, Head: m.Head,
			Inventory:     append([]int(nil), m.Inventory...),
			LearnedSpells: append([]int(nil), m.LearnedSpells...)}
	}
	return out
}

func (g *Game) restore(s saveState) {
	// Go 冒險之書不保存暫態場景咒文 timer；restore 必須清掉同一 Game instance
	// 載入前的透明效果。原版 save 是否序列化這類 timer 仍需獨立 RE。
	g.remoaru = 0
	g.toramana, g.hazardGuard = false, false
	g.heroExp, g.heroHP, g.heroMP, g.heroStat, g.heroGold =
		s.HeroExp, s.HeroHP, s.HeroMP, s.HeroStat, s.HeroGold
	g.heroName, g.heroGender = append([]int(nil), s.HeroName...), s.HeroGender
	g.heroInit = true
	g.ensureHeroStats() // 舊版 JSON 沒有 stats 欄時，以當前等級 target 安全升級
	if g.heroHP > int(g.heroStat[stats.HP]) {
		g.heroHP = int(g.heroStat[stats.HP])
	}
	if g.heroMP > int(g.heroStat[stats.MP]) {
		g.heroMP = int(g.heroStat[stats.MP])
	}
	g.inventory = append([]int(nil), s.Inventory...)
	g.equip = s.Equip
	if !s.EquipmentV2 {
		migrateLegacyEmptyEquipment(&g.equip)
	}
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
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head,
				Inventory:     append([]int(nil), c.Inventory...),
				LearnedSpells: append([]int(nil), c.LearnedSpells...)}
			if !s.EquipmentV2 {
				migrateLegacyMemberEquipment(m)
			}
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
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head,
				Inventory: append([]int(nil), c.Inventory...), LearnedSpells: append([]int(nil), c.LearnedSpells...)}
			if !s.EquipmentV2 {
				migrateLegacyMemberEquipment(m)
			}
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
	if len(s.Roster) > 0 { // 還原名冊(未入隊角色)
		g.roster = make([]*Member, len(s.Roster))
		for i, c := range s.Roster {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head,
				Inventory:     append([]int(nil), c.Inventory...),
				LearnedSpells: append([]int(nil), c.LearnedSpells...)}
			if !s.EquipmentV2 {
				migrateLegacyMemberEquipment(m)
			}
			if len(m.Name) == 0 && c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			m.ensureStats()
			m.syncLearnedSpells()
			g.roster[i] = m
		}
	}
	g.shipOwned, g.shipX, g.shipY = s.ShipOwned, s.ShipX, s.ShipY
	g.phoenixOwned, g.phoenixAboard = s.PhoenixOwned, s.PhoenixAboard
	g.phoenixX, g.phoenixY = s.PhoenixX, s.PhoenixY
	g.initStoryBits()
	if len(s.StoryBits) > 0 {
		copy(g.storyBits[:], s.StoryBits) // 32-byte 舊檔保留前 256 flags；高位沿用原版初值
	}
	g.syncTemporaryRoleVisual() // 角色外觀由 pack event 的 active flag 推導，不保存第二份狀態
	g.worldState = s.WorldState
	if g.flags[0x35] { // 舊 remake 存檔：自造 flag0x35 遷移至原版 [0x4f44] bit0x40。
		g.worldState |= worldStateRainbowBridge
		delete(g.flags, 0x35)
	}
	g.dnPhase, g.dnStep = s.DNPhase&3, s.DNStep
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
	if s.InTown {
		g.restoreTownScene(s.Cty, s.Section)
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
	g.applyRainbowBridge()
	g.applyPackWorldMapPatches()
	g.applyDaynightPalette()
	g.dlg.heroName = append([]int(nil), g.heroName...)
}

func migrateLegacyEmptyEquipment(eq *[4]int) {
	for i, code := range eq {
		if code == 0 {
			eq[i] = -1
		}
	}
}

func migrateLegacyMemberEquipment(m *Member) {
	eq := [4]int{m.Weapon, m.Armor, m.Shield, m.Head}
	migrateLegacyEmptyEquipment(&eq)
	m.Weapon, m.Armor, m.Shield, m.Head = eq[0], eq[1], eq[2], eq[3]
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
	b, err := encodeSave(g.snapshot())
	if err != nil {
		return err
	}
	_ = os.MkdirAll(filepath.Dir(savePath()), 0o755)
	return os.WriteFile(savePath(), b, 0o644)
}

// Load 讀存檔(不存在 → 靜默略過,回 nil)。
func (g *Game) Load() error {
	b, err := os.ReadFile(savePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	s, err := decodeSave(b)
	if err != nil {
		return err
	}
	// 舊存檔沒有 pack metadata，沿既有 migration 路徑接受；新格式則不得把 DQ1/DQ2、
	// modified override 或不同 schema 的狀態套進目前 DQ3 pack。
	if s.PackID != "" && (g.pack == nil || s.PackID != g.pack.ID() || s.PackSchema != g.pack.Schema() ||
		s.PackContentHash != g.pack.ContentHash()) {
		if g.pack == nil {
			return fmt.Errorf("save requires game pack %s/%s/%s but current Game has none",
				s.PackID, s.PackSchema, s.PackContentHash)
		}
		return fmt.Errorf("save game pack mismatch: save=%s/%s/%s current=%s/%s/%s",
			s.PackID, s.PackSchema, s.PackContentHash,
			g.pack.ID(), g.pack.Schema(), g.pack.ContentHash())
	}
	g.restore(s)
	return nil
}
