package game

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

// 存檔(冒險之書):持久化主角進度 + 位置。Go port 自有格式(非 C 存檔二進位相容),
// 因 remake 是重表達;只求本機讀寫一致(round-trip)。教會/記錄點觸發存檔。
type saveState struct {
	HeroExp       uint32       `json:"exp"`
	HeroHP        int          `json:"hp"`
	HeroMP        int          `json:"mp"`
	HeroStat      stats.Values `json:"stats,omitempty"`
	HeroGold      int          `json:"gold"`
	HeroName      []int        `json:"heroname,omitempty"` // 主角姓名(glyph index;newgame.go 命名創建)
	HeroGender    int          `json:"herogender"`         // 0=男 1=女
	Inventory     []int        `json:"inv"`
	Equip         [4]int       `json:"eq"`
	Comps         []compSav    `json:"comps"`
	Roster        []compSav    `json:"roster,omitempty"` // 酒場名冊(未必在隊伍中的角色;見 recruit.go)
	Flags         []int        `json:"flags"`
	ShipOwned     bool         `json:"shipowned"`
	PhoenixOwned  bool         `json:"phoenixowned,omitempty"`
	PhoenixAboard bool         `json:"phoenixaboard,omitempty"`
	PhoenixX      int          `json:"phoenixx,omitempty"`
	PhoenixY      int          `json:"phoenixy,omitempty"`
	ShipX         int          `json:"shipx"`
	ShipY         int          `json:"shipy"`
	PX            int          `json:"px"`
	PY            int          `json:"py"`
	InTown        bool         `json:"town"`
	StoryBits     []byte       `json:"storybits,omitempty"`
	DNPhase       int          `json:"dnphase,omitempty"`
	DNStep        int          `json:"dnstep,omitempty"`
	Cty           int          `json:"cty,omitempty"`
	Section       int          `json:"section,omitempty"`
	Layer         int          `json:"layer,omitempty"`
}

// compSav 是一名同伴的存檔資料。
type compSav struct {
	Name                        []int `json:"name,omitempty"`
	Class, Gender               int
	Exp                         uint32
	Stats                       stats.Values `json:"stats,omitempty"`
	CurHP, CurMP                int
	Weapon, Armor, Shield, Head int
}

func encodeSave(s saveState) ([]byte, error) { return json.Marshal(s) }

func decodeSave(b []byte) (saveState, error) {
	var s saveState
	err := json.Unmarshal(b, &s)
	return s, err
}

func (g *Game) snapshot() saveState {
	s := saveState{
		HeroExp: g.heroExp, HeroHP: g.heroHP, HeroMP: g.heroMP, HeroStat: g.heroStat, HeroGold: g.heroGold,
		HeroName: append([]int(nil), g.heroName...), HeroGender: g.heroGender,
		Inventory: append([]int(nil), g.inventory...),
		Equip:     g.equip,
		Comps:     compsToSav(g.companions),
		Roster:    compsToSav(g.roster),
		Flags:     flagsToSav(g.flags),
		ShipOwned: g.shipOwned, ShipX: g.shipX, ShipY: g.shipY,
		PhoenixOwned: g.phoenixOwned, PhoenixAboard: g.phoenixAboard,
		PhoenixX: g.phoenixX, PhoenixY: g.phoenixY,
		PX: g.px, PY: g.py, InTown: g.inTown,
		StoryBits: append([]byte(nil), g.storyBits[:]...),
		DNPhase:   g.dnPhase, DNStep: g.dnStep, Cty: g.curCty, Layer: g.layer,
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
			Weapon: m.Weapon, Armor: m.Armor, Shield: m.Shield, Head: m.Head}
	}
	return out
}

func (g *Game) restore(s saveState) {
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
	g.flags = map[int]bool{}
	for _, k := range s.Flags {
		g.flags[k] = true
	}
	if len(s.Comps) > 0 { // 還原同伴
		g.companions = make([]*Member, len(s.Comps))
		for i, c := range s.Comps {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head}
			if len(m.Name) == 0 && c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			m.ensureStats()
			g.companions[i] = m
		}
	}
	if len(s.Roster) > 0 { // 還原名冊(未入隊角色)
		g.roster = make([]*Member, len(s.Roster))
		for i, c := range s.Roster {
			m := &Member{Name: append([]int(nil), c.Name...), Class: c.Class, Gender: c.Gender,
				Exp: c.Exp, Stats: c.Stats, CurHP: c.CurHP, CurMP: c.CurMP,
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head}
			if len(m.Name) == 0 && c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			m.ensureStats()
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
	g.dnPhase, g.dnStep = s.DNPhase&3, s.DNStep
	g.layer, g.curCty = s.Layer, s.Cty
	g.px, g.py, g.inTown = s.PX, s.PY, s.InTown
	if s.InTown {
		g.restoreTownScene(s.Cty, s.Section)
	} else {
		g.cur = g.overworldScene()
		g.curCty = -1
	}
	g.applyDaynightPalette()
	g.dlg.heroName = append([]int(nil), g.heroName...)
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
	g.restore(s)
	return nil
}
