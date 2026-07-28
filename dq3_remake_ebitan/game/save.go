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
	HeroExp    uint32       `json:"exp"`
	HeroHP     int          `json:"hp"`
	HeroMP     int          `json:"mp"`
	HeroStat   stats.Values `json:"stats,omitempty"`
	HeroGold   int          `json:"gold"`
	HeroName   []int        `json:"heroname,omitempty"` // 主角姓名(glyph index;newgame.go 命名創建)
	HeroGender int          `json:"herogender"`         // 0=男 1=女
	Inventory  []int        `json:"inv"`
	Equip      [4]int       `json:"eq"`
	Comps      []compSav    `json:"comps"`
	Roster     []compSav    `json:"roster,omitempty"` // 酒場名冊(未必在隊伍中的角色;見 recruit.go)
	Flags      []int        `json:"flags"`
	ShipOwned  bool         `json:"shipowned"`
	ShipX      int          `json:"shipx"`
	ShipY      int          `json:"shipy"`
	PX         int          `json:"px"`
	PY         int          `json:"py"`
	InTown     bool         `json:"town"`
}

// compSav 是一名同伴的存檔資料。
type compSav struct {
	Class, Gender               int
	Exp                         uint32
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
	return saveState{
		HeroExp: g.heroExp, HeroHP: g.heroHP, HeroMP: g.heroMP, HeroStat: g.heroStat, HeroGold: g.heroGold,
		HeroName: append([]int(nil), g.heroName...), HeroGender: g.heroGender,
		Inventory: append([]int(nil), g.inventory...),
		Equip:     g.equip,
		Comps:     compsToSav(g.companions),
		Roster:    compsToSav(g.roster),
		Flags:     flagsToSav(g.flags),
		ShipOwned: g.shipOwned, ShipX: g.shipX, ShipY: g.shipY,
		PX: g.px, PY: g.py, InTown: g.inTown,
	}
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
		out[i] = compSav{Class: m.Class, Gender: m.Gender, Exp: m.Exp, CurHP: m.CurHP, CurMP: m.CurMP,
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
			m := &Member{Class: c.Class, Gender: c.Gender, Exp: c.Exp, CurHP: c.CurHP, CurMP: c.CurMP,
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head}
			if c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class]
			}
			g.companions[i] = m
		}
	}
	if len(s.Roster) > 0 { // 還原名冊(未入隊角色;姓名如同 companions 一樣不隨存檔保留,見下方註)
		g.roster = make([]*Member, len(s.Roster))
		for i, c := range s.Roster {
			m := &Member{Class: c.Class, Gender: c.Gender, Exp: c.Exp, CurHP: c.CurHP, CurMP: c.CurMP,
				Weapon: c.Weapon, Armor: c.Armor, Shield: c.Shield, Head: c.Head}
			if c.Class >= 0 && c.Class < 8 {
				m.Name = classNames[c.Class] // compSav 未存自訂姓名 glyph(同 companions 既有限制),回退職業名
			}
			g.roster[i] = m
		}
	}
	g.shipOwned, g.shipX, g.shipY = s.ShipOwned, s.ShipX, s.ShipY
	g.px, g.py, g.inTown = s.PX, s.PY, s.InTown
	if s.InTown {
		g.cur = g.town
	} else {
		g.cur = g.over
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
