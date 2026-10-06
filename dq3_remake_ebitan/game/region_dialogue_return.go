package game

import (
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"io/fs"
)

type regionDialogueReturnState struct {
	event        *gamepack.RegionDialogueReturnEvent
	phase, ticks int
}

func validateRegionDialogueReturnSources(assets fs.FS, pack *gamepack.Pack) error {
	for _, e := range pack.Events.RegionDialogueReturnEvents {
		if assets == nil {
			return fmt.Errorf("region dialogue return assets are nil")
		}
		raw, err := fs.ReadFile(assets, ctyFile(e.CTYRaw))
		if err != nil {
			return err
		}
		if e.Section >= len(raw)/2 {
			return fmt.Errorf("region dialogue return section outside source")
		}
		town, err := dq3data.OpenTown(raw, e.Section, false)
		if err != nil {
			return err
		}
		if e.ActorRecordRaw >= len(town.NPCs) {
			return fmt.Errorf("region dialogue return actor outside source")
		}
		found := false
		for _, h := range town.HiMap {
			subid := int(h & 31)
			if subid > 0 && subid <= len(town.SpecialHandlers) && town.SpecialHandlers[subid-1] == e.HandlerRaw {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("region dialogue return handler absent from source")
		}
		d, ok := pack.TextDefinition(e.TextID)
		if !ok || d.Source.Record == nil {
			return fmt.Errorf("region dialogue return text source missing")
		}
		raw, err = fs.ReadFile(assets, d.Source.File)
		if err != nil {
			return err
		}
		native := dq3data.LoadText(nil, raw).Record(*d.Source.Record)
		if len(native) != len(d.GlyphCodes) {
			return fmt.Errorf("region dialogue return text length differs")
		}
		for i, c := range native {
			if int(c) != d.GlyphCodes[i] {
				return fmt.Errorf("region dialogue return text differs")
			}
		}
	}
	return nil
}

func (g *Game) regionDialogueReturnAtPlayer() *gamepack.RegionDialogueReturnEvent {
	if g.pack == nil || !g.inTown || g.cur == nil {
		return nil
	}
	index := g.py*g.cur.w + g.px
	if index < 0 || index >= len(g.cur.hiMap) {
		return nil
	}
	subid := int(g.cur.hiMap[index] & 31)
	if subid == 0 || subid > len(g.cur.specialHandlers) {
		return nil
	}
	for i := range g.pack.Events.RegionDialogueReturnEvents {
		e := &g.pack.Events.RegionDialogueReturnEvents[i]
		if g.curCty == e.CTYRaw && g.cur.sec == e.Section && g.cur.specialHandlers[subid-1] == e.HandlerRaw {
			return e
		}
	}
	return nil
}

// Forced escort movement selects a reviewed event without dispatching it.
// Ordinary tiles preserve the selection until a successful normal input step.
func (g *Game) deferRegionDialogueReturnAtPlayer() {
	if e := g.regionDialogueReturnAtPlayer(); e != nil {
		g.deferredRegionDialogueReturnID = e.ID
	}
}

func (g *Game) tryRegionDialogueReturn() (bool, error) {
	if g.pack == nil || !g.inTown || g.cur == nil || g.dlg.open || g.regionDialogueReturn != nil {
		return false, nil
	}
	selected := g.regionDialogueReturnAtPlayer()
	index := g.py*g.cur.w + g.px
	if selected == nil && index >= 0 && index < len(g.cur.hiMap) && g.cur.hiMap[index]&31 == 0 {
		for i := range g.pack.Events.RegionDialogueReturnEvents {
			e := &g.pack.Events.RegionDialogueReturnEvents[i]
			if e.ID == g.deferredRegionDialogueReturnID && g.curCty == e.CTYRaw && g.cur.sec == e.Section {
				selected = e
				break
			}
		}
	}
	// Native dispatch clears the selection before the handler checks its gate.
	g.deferredRegionDialogueReturnID = ""
	if selected == nil || !g.storyFlag(selected.RequiredFlagRaw) {
		return false, nil
	}
	for j := range g.cur.npcs {
		if g.cur.npcs[j].recordIndex == selected.ActorRecordRaw {
			g.cur.npcs[j].facing = selected.ActorFacing
			g.regionDialogueReturn = &regionDialogueReturnState{event: selected}
			return true, nil
		}
	}
	return false, fmt.Errorf("region dialogue return actor absent")
}

func (g *Game) validateDeferredRegionDialogueReturnSave(s saveState) error {
	if s.DeferredRegionDialogueReturnID == "" {
		return nil
	}
	if g.pack == nil || s.PackID == "" || !s.InTown || s.OpeningHomeAwait {
		return fmt.Errorf("deferred region return save checkpoint invalid")
	}
	for _, e := range g.pack.Events.RegionDialogueReturnEvents {
		if e.ID != s.DeferredRegionDialogueReturnID {
			continue
		}
		flag := e.RequiredFlagRaw
		if s.Cty != e.CTYRaw || s.Section != e.Section || flag/8 >= len(s.StoryBits) || s.StoryBits[flag/8]&(128>>uint(flag%8)) == 0 {
			return fmt.Errorf("deferred region return save scene or gate differs")
		}
		return nil
	}
	return fmt.Errorf("deferred region return save event unknown")
}

// Pending return events own input until the native caller's final move. The
// forced step checks collision but does not dispatch another tile event.
func (g *Game) stepRegionDialogueReturn() (bool, error) {
	s := g.regionDialogueReturn
	if s == nil {
		return false, nil
	}
	e := s.event
	switch s.phase {
	case 0:
		s.ticks++
		if s.ticks < e.TurnHoldFrames {
			return true, nil
		}
		p, ok := g.pack.OpeningPrelude()
		if !ok || p.ID != e.PresentationID {
			return true, fmt.Errorf("region dialogue return presentation absent")
		}
		frame, fok := g.pack.TextGlyphCodes(p.FrameTextID)
		codes, cok := g.pack.TextGlyphCodes(e.TextID)
		if !fok || !cok || !g.dlg.openRecord(codes) {
			return true, fmt.Errorf("region dialogue return text absent")
		}
		g.dlg.prelude, g.dlg.preludeFrame, g.dlg.layout, g.dlg.shadow = p, frame, p.Window, &e.Shadow
		s.phase, s.ticks = 1, 0
	case 1:
		g.dlg.Tick()
		if !g.dlg.open {
			g.dlg.open = true
			s.phase, s.ticks = 2, 0
		}
	case 2:
		s.ticks++
		if s.ticks < e.ReturnHoldFrames {
			return true, nil
		}
		g.dlg.open = false
		g.clearOpeningPresentation()
		dx, dy := dirDelta(e.ReturnDirection)
		g.facing = e.ReturnDirection
		if g.tryMove(g.px+dx, g.py+dy) {
			g.recordPartyTrail()
			g.px, g.py = g.px+dx, g.py+dy
		}
		g.regionDialogueReturn = nil
	}
	return true, nil
}

// The completed stationary escort actor remains needed while a return event's
// native gate is set. Reconstruct it from the existing pack checkpoint when
// the ordinary scene loader filters its original home visibility flag.
// This is remake save recovery, not an original save-format parity claim.
func (g *Game) loadRegionDialogueReturnActor(s saveState) (*npcInst, error) {
	if g.pack == nil || !s.InTown || s.OpeningHomeAwait {
		return nil, nil
	}
	escort, ok := g.pack.OpeningEscort()
	if !ok || escort.ArrivalLeaderRecord == nil || len(escort.ArrivalFrames) == 0 || s.Cty != escort.Destination.CTY || s.Section != escort.Destination.Section {
		return nil, nil
	}
	flag := func(raw int) bool {
		return raw >= 0 && raw/8 < len(s.StoryBits) && s.StoryBits[raw/8]&(128>>uint(raw%8)) != 0
	}
	for _, raw := range escort.SetStoryFlags {
		if !flag(raw) {
			return nil, nil
		}
	}
	for _, raw := range escort.ClearStoryFlags {
		if flag(raw) {
			return nil, nil
		}
	}
	needed := false
	for _, e := range g.pack.Events.RegionDialogueReturnEvents {
		if e.CTYRaw == s.Cty && e.Section == s.Section && e.ActorRecordRaw == *escort.ArrivalLeaderRecord && flag(e.RequiredFlagRaw) {
			needed = true
			break
		}
	}
	if !needed {
		return nil, nil
	}
	if s.Cty < 0 || s.Cty >= len(mapBlkNum) {
		return nil, fmt.Errorf("saved return actor scene unavailable")
	}
	scene, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS, s.Cty, mapBlkNum[s.Cty], s.Section, s.DNPhase, nil)
	if err != nil {
		return nil, err
	}
	last := escort.ArrivalFrames[len(escort.ArrivalFrames)-1]
	for _, n := range scene.npcs {
		if n.recordIndex == *escort.ArrivalLeaderRecord {
			motion := g.npcMotion()
			if motion == nil || n.ctrl&motion.MoveMask != 0 || last.Leader.X < 0 || last.Leader.Y < 0 || last.Leader.X >= scene.w || last.Leader.Y >= scene.h {
				return nil, fmt.Errorf("saved return actor lacks a stationary checkpoint")
			}
			n.x, n.y, n.facing = last.Leader.X, last.Leader.Y, last.LeaderFacing
			return &n, nil
		}
	}
	return nil, fmt.Errorf("saved return actor source missing")
}
