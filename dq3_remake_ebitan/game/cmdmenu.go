package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"io/fs"
)

func validateFieldCommandSources(assets fs.FS, p *gamepack.Pack, tx *dq3data.Text) error {
	s := p.Interface.FieldCommandMenu
	if s == nil || tx == nil {
		return fmt.Errorf("command sources missing")
	}
	d, ok := p.TextDefinition(s.TextID)
	if !ok || d.Source.Record == nil {
		return fmt.Errorf("command text source missing")
	}
	raw, err := fs.ReadFile(assets, d.Source.File)
	if err != nil {
		return err
	}
	codes := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
	if len(codes) != len(d.GlyphCodes) {
		return fmt.Errorf("command source shape differs")
	}
	for i, c := range codes {
		if int(c) != d.GlyphCodes[i] {
			return fmt.Errorf("command source text differs")
		}
		if c != dq3data.TxtNL {
			if _, ok := tx.Glyph(int(c)); !ok {
				return fmt.Errorf("command source glyph missing")
			}
		}
	}
	if _, ok := tx.Glyph(s.CursorGlyph); !ok {
		return fmt.Errorf("command cursor glyph missing")
	}
	return nil
}

// Command roles are shared engine semantics; pack entries supply native order.
const (
	cmdTalk    = iota // 對話
	cmdSpell          // 咒文
	cmdStatus         // 狀況
	cmdItem           // 道具
	cmdEquip          // 裝備
	cmdExamine        // 調查
	cmdCount
)

// CmdMenu preserves a semantic cursor; native order comes from the pack.
type CmdMenu struct {
	tx          *dq3data.Text
	cursor      int
	open        bool
	labels      [cmdCount][2]int
	title       [2]int
	labelsReady bool
	hits        hitList // 6 格可點區塊,draw() 重建(P2 直接點選)
	contract    *gamepack.FieldCommandMenu
	order       []int
}

// setLabels 安裝 versioned game pack 的玩家可見字模。直接 CmdMenu fixture
// 可以不安裝；production bootstrap 缺資料時會在 NewGameWithPack fail closed。
func (m *CmdMenu) setLabels(labels gamepack.FieldCommandLabels) {
	m.title = [2]int{labels.Title.PrimaryGlyph, labels.Title.SecondaryGlyph}
	m.labels = [cmdCount][2]int{
		{labels.Talk.PrimaryGlyph, labels.Talk.SecondaryGlyph},
		{labels.Spell.PrimaryGlyph, labels.Spell.SecondaryGlyph},
		{labels.Status.PrimaryGlyph, labels.Status.SecondaryGlyph},
		{labels.Item.PrimaryGlyph, labels.Item.SecondaryGlyph},
		{labels.Equip.PrimaryGlyph, labels.Equip.SecondaryGlyph},
		{labels.Examine.PrimaryGlyph, labels.Examine.SecondaryGlyph},
	}
	m.labelsReady = true
}

func (m *CmdMenu) configure(s *gamepack.FieldCommandMenu) error {
	if s == nil || len(s.Entries) != cmdCount {
		return fmt.Errorf("command menu contract missing")
	}
	roles := map[string]int{"talk": cmdTalk, "spell": cmdSpell, "status": cmdStatus, "item": cmdItem, "equip": cmdEquip, "examine": cmdExamine}
	m.order = nil
	for _, e := range s.Entries {
		role, ok := roles[e.Command]
		if !ok {
			return fmt.Errorf("command role unknown")
		}
		m.order = append(m.order, role)
	}
	m.contract = s
	return nil
}

func (m *CmdMenu) Open() {
	if len(m.order) == 0 {
		return
	}
	m.cursor, m.open = m.order[0], true
}

func (m *CmdMenu) nativeCursor() int {
	for i, role := range m.order {
		if role == m.cursor {
			return i
		}
	}
	return -1
}

// move consumes the native linear two-column selector. Facing codes are shared.
func (m *CmdMenu) move(dir int) {
	i := m.nativeCursor()
	n := len(m.order)
	if i < 0 || n == 0 {
		return
	}
	switch dir {
	case 1: // 上
		i = (i + n - 1) % n
	case 0: // 下
		i = (i + 1) % n
	case 2, 3: // 左 / 右(2 欄繞回)
		i = (i + n/2) % n
	}
	m.cursor = m.order[i]
}

// drawFieldCommandMenu uses the same indexed window and party HUD consumers
// as the accepted native source. No state or animation phase is overridden.
func (g *Game) drawFieldCommandMenu() {
	if g.cmd.open {
		g.drawNativeFieldCommandMenu()
	}
}
func (g *Game) drawNativeFieldCommandMenu() {
	m := &g.cmd
	s := m.contract
	if s == nil || g.cur == nil || g.newGame.raster == nil {
		return
	}
	g.drawFieldIdleStatus()
	r := *g.newGame.raster
	r.pixels = make([]byte, ScreenW*ScreenH)
	r.palette = append([]dq3data.Color(nil), g.cur.pal...)
	r.backgroundPalette = append([]dq3data.Color(nil), g.cur.pal...)
	r.palette[s.FontIndex] = g.fieldIdleForeground()
	r.windows = map[string]gamepack.RawNewGameWindow{s.RawWindow.ID: s.RawWindow}
	codes, ok := g.pack.TextGlyphCodes(s.TextID)
	if !ok {
		return
	}
	r.texts = map[string][]uint16{s.TextID: codes}
	if !r.captureBackground(g.rgba) {
		return
	}
	r.window(m.tx, gamepack.RasterWindowRef{RawWindowID: s.RawWindow.ID, TextID: s.TextID})
	m.hits.reset()
	half := len(s.Entries) / 2
	hitWidth := s.Entries[half].X - s.Entries[0].X
	for i, e := range s.Entries {
		if m.order[i] == m.cursor {
			r.opaqueGlyph(m.tx, e.X, e.Y, s.CursorGlyph)
		}
		m.hits.add(e.X, e.Y, hitWidth, dq3data.GlyphPx, m.order[i])
	}
	drawIndexedPCX(g.rgba, r.pixels, r.palette)
}
