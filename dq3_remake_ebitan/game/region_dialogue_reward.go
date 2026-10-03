package game

import (
	"fmt"
	"io/fs"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func validateRegionDialogueRewardSources(assets fs.FS, pack *gamepack.Pack) error {
	for _, e := range pack.Events.RegionDialogueRewardEvents {
		if assets == nil {
			return fmt.Errorf("region dialogue reward assets are nil")
		}
		raw, err := fs.ReadFile(assets, ctyFile(e.CTYRaw))
		if err != nil {
			return fmt.Errorf("region dialogue reward scene: %w", err)
		}
		if e.Section >= len(raw)/2 {
			return fmt.Errorf("region dialogue reward section outside source")
		}
		town, err := dq3data.OpenTown(raw, e.Section, false)
		if err != nil {
			return err
		}
		if e.Tile.X >= town.W || e.Tile.Y >= town.H {
			return fmt.Errorf("region dialogue reward tile outside source")
		}
		subid := int(town.HiMap[e.Tile.Y*town.W+e.Tile.X] & 31)
		if subid == 0 || subid > len(town.SpecialHandlers) || town.SpecialHandlers[subid-1] != e.HandlerRaw {
			return fmt.Errorf("region dialogue reward handler differs from source")
		}
		d, ok := pack.TextDefinition(e.TextID)
		if !ok || d.Source.Record == nil {
			return fmt.Errorf("region dialogue reward text source missing")
		}
		raw, err = fs.ReadFile(assets, d.Source.File)
		if err != nil {
			return err
		}
		native := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if len(native) != len(d.GlyphCodes) {
			return fmt.Errorf("region dialogue reward text length differs from source")
		}
		for i, c := range native {
			if int(c) != d.GlyphCodes[i] {
				return fmt.Errorf("region dialogue reward text differs from source")
			}
		}
	}
	return nil
}

func (g *Game) tryRegionDialogueReward() bool {
	if g.pack == nil || !g.inTown || g.cur == nil || g.dlg.open || g.regionDialogueReward != nil {
		return false
	}
	for i := range g.pack.Events.RegionDialogueRewardEvents {
		e := &g.pack.Events.RegionDialogueRewardEvents[i]
		if g.curCty != e.CTYRaw || g.cur.sec != e.Section || g.px != e.Tile.X || g.py != e.Tile.Y ||
			!g.storyFlag(e.RequiredFlagRaw) || g.progressDone(e.ProgressFlagRaw) {
			continue
		}
		p, ok := g.pack.OpeningPrelude()
		if !ok || p.ID != e.PresentationID {
			return false
		}
		frame, ok := g.pack.TextGlyphCodes(p.FrameTextID)
		if !ok {
			return false
		}
		codes, ok := g.pack.TextGlyphCodes(e.TextID)
		if !ok || !g.dlg.openRecord(codes) {
			return false
		}
		g.dlg.prelude, g.dlg.preludeFrame, g.dlg.layout, g.dlg.shadow = p, frame, p.Window, &e.Shadow
		g.regionDialogueReward = e
		return true
	}
	return false
}

// The native caller continues after each grant, including failed grants. Its
// rewards and flag writes follow the blocking text reader's return (docs/188).
func (g *Game) completeRegionDialogueReward() bool {
	e := g.regionDialogueReward
	if e == nil || g.dlg.open {
		return false
	}
	g.regionDialogueReward = nil
	for _, raw := range e.ItemRawIDs {
		g.grantPartyItem(raw)
	}
	g.heroGold += e.Gold
	g.setStoryFlag(e.ClearFlagRaw, false)
	g.setStoryFlag(e.SetFlagRaw, true)
	g.progressSet(e.ProgressFlagRaw)
	g.clearOpeningPresentation()
	return true
}
