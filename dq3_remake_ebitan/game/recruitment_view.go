package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/stats"
)

const (
	rcViewAbility = rcMusicWait + 1 + iota
	rcViewClose
	rcViewRename
	rcViewSpells
)

// docs/188 的有限 READY。未知咒文、異常與其他裝備保持清單，
// 不借用創角 cloth renderer 猜出缺少的角色頁。
func (g *Game) startRecruitmentView(m *Member) {
	rc, tv := &g.recruit, &g.tavern
	if m == nil || rc.selection == nil || tv.labels == nil || rc.raster == nil || g.shop.items == nil ||
		len(m.Name) == 0 || m.Gender < 0 || m.Gender >= len(rc.selection.GenderTextIDs) ||
		m.Conditions != 0 || m.Stats == (stats.Values{}) ||
		m.Weapon >= 0 || [4]int{m.Weapon, m.Armor, m.Shield, m.Head} != tv.equipment {
		return
	}
	spellNames, ok := rc.raster.spells.OrderedTextIDs(m.LearnedSpells)
	if !ok {
		return
	}
	labels := *tv.labels
	labels.Hero = nil
	for _, option := range rc.selection.ClassOptions {
		if option.ClassRaw == m.Class {
			for _, code := range rc.texts[option.TextID] {
				labels.Hero = append(labels.Hero, int(code))
			}
			break
		}
	}
	if len(labels.Hero) == 0 {
		return
	}
	// 衍生值在副本上計算，不能讓舊 ensureStats 相容路徑回寫名冊。
	member := *m
	rc.viewFlow = &NewGameFlow{
		stage: ngReview, labels: &labels,
		ni:      NameInput{nameBuf: append([]int(nil), m.Name...)},
		preview: m.Stats, previewGender: m.Gender, previewLevel: m.Level(),
		previewHP: m.CurHP, previewMP: m.CurMP, previewExp: int(m.Exp),
		previewDef: member.Def(g.shop.items),
	}
	rc.stage = rcViewAbility
	rc.viewSpellNames = spellNames
}

func (rc *Recruit) viewInput(in InputState) {
	if rc.viewFlow == nil || rc.selection == nil {
		return
	}
	knownKey := in.Confirm || in.Enter || in.Cancel || in.DirEdge >= 0 || in.Rename
	if rc.stage == rcViewSpells {
		if knownKey || in.AnyKeyEdge || in.Tapped {
			rc.stage = rcViewClose
		}
		return
	}
	if rc.stage == rcViewAbility {
		if knownKey || in.AnyKeyEdge || in.Tapped {
			rc.stage = rcViewClose
			if len(rc.viewSpellNames) > 0 {
				rc.stage = rcViewSpells
			}
		}
		return // 第一次輸入不能同時消費第二個讀鍵。
	}
	// 具名改名由 Game 分派；其他沒有映射的鍵不猜成關閉。
	if knownKey || in.Tapped {
		rc.viewFlow = nil
		rc.viewSpellNames = nil
		rc.startSelectionText(rc.selection.AgainTextID, rcAgain)
	}
}

func (rc *Recruit) drawView(rgba []byte, white dq3data.Color) {
	if rc.viewFlow == nil || rc.raster == nil {
		return
	}
	rc.drawEntry(rgba, white)
	r := rc.raster
	if !r.captureBackground(rgba) {
		return
	}
	r.ability(rc.tx, rc.viewFlow)
	if rc.stage == rcViewSpells {
		r.characterSpells(rc.tx, rc.viewSpellNames)
	}
	if rc.stage == rcViewRename && rc.renameFlow != nil {
		r.drawContent(rc.tx, rc.renameFlow)
	}
	drawIndexedPCX(rgba, r.pixels, r.palette)
}
