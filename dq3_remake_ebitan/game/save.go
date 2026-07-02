package game

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// 存檔(冒險之書):持久化主角進度 + 位置。Go port 自有格式(非 C 存檔二進位相容),
// 因 remake 是重表達;只求本機讀寫一致(round-trip)。教會/記錄點觸發存檔。
type saveState struct {
	HeroExp   uint32 `json:"exp"`
	HeroHP    int    `json:"hp"`
	HeroGold  int    `json:"gold"`
	Inventory []int  `json:"inv"`
	Equip     [4]int `json:"eq"`
	PX        int    `json:"px"`
	PY        int    `json:"py"`
	InTown    bool   `json:"town"`
}

func encodeSave(s saveState) ([]byte, error) { return json.Marshal(s) }

func decodeSave(b []byte) (saveState, error) {
	var s saveState
	err := json.Unmarshal(b, &s)
	return s, err
}

func (g *Game) snapshot() saveState {
	return saveState{
		HeroExp: g.heroExp, HeroHP: g.heroHP, HeroGold: g.heroGold,
		Inventory: append([]int(nil), g.inventory...),
		Equip:     g.equip,
		PX:        g.px, PY: g.py, InTown: g.inTown,
	}
}

func (g *Game) restore(s saveState) {
	g.heroExp, g.heroHP, g.heroGold = s.HeroExp, s.HeroHP, s.HeroGold
	g.heroInit = true
	g.inventory = append([]int(nil), s.Inventory...)
	g.equip = s.Equip
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
