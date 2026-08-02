package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

type trackedWorldPosition struct {
	X, Y, Layer int
}

func (g *Game) activateTrackedWorldObject(a gamepack.WorldObjectActivation) bool {
	if g.pack == nil {
		return false
	}
	obj, ok := g.pack.TrackedWorldObject(a.ObjectID)
	if !ok || obj.Layer != a.Position.Layer {
		return false
	}
	if g.trackedWorldPositions == nil {
		g.trackedWorldPositions = map[string]trackedWorldPosition{}
	}
	g.trackedWorldPositions[obj.ID] = trackedWorldPosition{
		X: a.Position.X, Y: a.Position.Y, Layer: a.Position.Layer,
	}
	g.ensureTrackedWorldObjectWindow(obj.Relocation)
	return true
}

func (g *Game) ensureTrackedWorldObjectWindow(r gamepack.TrackedWorldObjectRelocation) {
	if g.worldObjectBufferValid || r.WindowWidth <= 0 || r.WindowHeight <= 0 {
		return
	}
	x, y := g.px, g.py
	if g.inTown {
		x, y = g.overPx, g.overPy
	}
	maxX, maxY := 0, 0
	if sc := g.overworldScene(); sc != nil {
		maxX, maxY = max0(sc.w-r.WindowWidth), max0(sc.h-r.WindowHeight)
	}
	g.worldObjectBufferX = clampi(x-r.OriginOffset, 0, maxX)
	g.worldObjectBufferY = clampi(y-r.OriginOffset, 0, maxY)
	g.worldObjectBufferValid = true
}

func (g *Game) trackedPositionInWindow(p trackedWorldPosition, r gamepack.TrackedWorldObjectRelocation) bool {
	return g.worldObjectBufferValid && p.X >= g.worldObjectBufferX && p.Y >= g.worldObjectBufferY &&
		p.X < g.worldObjectBufferX+r.WindowWidth && p.Y < g.worldObjectBufferY+r.WindowHeight
}

// rebuildTrackedWorldObjects mirrors IDA linear 0x12d94..0x12dde. An object
// already in the active 80x80 buffer stays put. Otherwise exactly one of the
// 18 original candidates is drawn; its persistent coordinate changes only if
// that candidate lies in the current buffer.
func (g *Game) rebuildTrackedWorldObjects() {
	if g.pack == nil {
		return
	}
	for _, obj := range g.pack.TrackedWorldObjects() {
		p, active := g.trackedWorldObjectPosition(&obj)
		if !active {
			continue
		}
		g.ensureTrackedWorldObjectWindow(obj.Relocation)
		if g.trackedPositionInWindow(p, obj.Relocation) || len(obj.Relocation.Candidates) == 0 {
			continue
		}
		candidate := obj.Relocation.Candidates[g.prng.Next(len(obj.Relocation.Candidates))]
		next := trackedWorldPosition{X: candidate.X, Y: candidate.Y, Layer: candidate.Layer}
		if g.trackedPositionInWindow(next, obj.Relocation) {
			g.trackedWorldPositions[obj.ID] = next
		}
	}
}

func (g *Game) updateTrackedWorldObjectWindow(direction int) {
	if g.pack == nil || g.inTown {
		return
	}
	objects := g.pack.TrackedWorldObjects()
	if len(objects) == 0 {
		return
	}
	r := objects[0].Relocation
	g.ensureTrackedWorldObjectWindow(r)
	dx, dy := dirDelta(direction)
	offsetX, offsetY := g.px-g.worldObjectBufferX, g.py-g.worldObjectBufferY
	rebuild := dx != 0 && (offsetX < r.HorizontalKeepMin || offsetX > r.HorizontalKeepMax) ||
		dy != 0 && (offsetY < r.VerticalKeepMin || offsetY > r.VerticalKeepMax)
	if !rebuild {
		return
	}
	maxX, maxY := max0(g.cur.w-r.WindowWidth), max0(g.cur.h-r.WindowHeight)
	if dx != 0 {
		g.worldObjectBufferX = clampi(g.px-r.OriginOffset, 0, maxX)
	} else {
		g.worldObjectBufferY = clampi(g.py-r.OriginOffset, 0, maxY)
	}
	g.rebuildTrackedWorldObjects()
}

func (g *Game) trackedWorldObjectPosition(obj *gamepack.TrackedWorldObject) (trackedWorldPosition, bool) {
	if obj == nil || g.worldState&uint16(obj.ActiveWorldStateMask) == 0 {
		return trackedWorldPosition{}, false
	}
	p, ok := g.trackedWorldPositions[obj.ID]
	return p, ok && p.Layer == obj.Layer
}

func (g *Game) tryEnterTrackedWorldObject() bool {
	if g.inTown || g.pack == nil {
		return false
	}
	for _, packed := range g.pack.TrackedWorldObjects() {
		obj := packed
		p, active := g.trackedWorldObjectPosition(&obj)
		if !active || p.Layer != g.layer || p.X != g.px || p.Y != g.py {
			continue
		}
		if obj.EntryVehicleMode == "disembark_ship_if_aboard" && g.shipAboard {
			g.shipX, g.shipY, g.shipAboard = g.px, g.py, false
		}
		g.enterTownCty(obj.EntranceCTYRaw)
		if g.inTown && g.cur != nil && obj.EntranceSection != 0 {
			g.restoreTownScene(obj.EntranceCTYRaw, obj.EntranceSection)
		}
		return g.inTown && g.curCty == obj.EntranceCTYRaw
	}
	return false
}

func (g *Game) drawTrackedWorldObjects(camX, camY int) {
	if g.inTown || g.pack == nil || g.cur == nil {
		return
	}
	for _, obj := range g.pack.TrackedWorldObjects() {
		p, active := g.trackedWorldObjectPosition(&obj)
		if !active || p.Layer != g.layer || p.X < camX || p.X >= camX+ViewCols ||
			p.Y < camY || p.Y >= camY+ViewRows || len(obj.TileRawByRNGLow2) == 0 {
			continue
		}
		frame := int(g.prng.State() & 3)
		blitTile(g.rgba, (p.X-camX)*TileW, (p.Y-camY)*TileH,
			g.cur.blk.Tile(obj.TileRawByRNGLow2[frame]), g.cur.pal)
	}
}

func (g *Game) useTrackedWorldObjectLocator(code int) bool {
	if g.pack == nil {
		return false
	}
	obj, ok := g.pack.TrackedWorldObjectByTrackerItem(code)
	if !ok {
		return false
	}
	p, active := g.trackedWorldObjectPosition(obj)
	if !active || g.layer != p.Layer || !g.openPackText(obj.TrackerDialogueTextIDs.Preamble) {
		return true
	}
	g.trackedWorldLocatorID = obj.ID
	g.trackedWorldLocatorStage = 1
	return true
}

func (g *Game) advanceTrackedWorldObjectLocator() {
	if g.trackedWorldLocatorStage == 0 || g.pack == nil {
		return
	}
	obj, ok := g.pack.TrackedWorldObject(g.trackedWorldLocatorID)
	p, active := g.trackedWorldObjectPosition(obj)
	if !ok || !active {
		g.trackedWorldLocatorStage, g.trackedWorldLocatorID = 0, ""
		return
	}
	switch g.trackedWorldLocatorStage {
	case 1:
		textID, distance := obj.TrackerDialogueTextIDs.East, p.X-g.px
		if distance < 0 {
			textID, distance = obj.TrackerDialogueTextIDs.West, -distance
		}
		if g.openPackText(textID) {
			g.setDlgVarGlyphs(dq3data.TxtVarItem, digitGlyphs(distance))
			g.trackedWorldLocatorStage = 2
		}
	case 2:
		textID, distance := obj.TrackerDialogueTextIDs.South, p.Y-g.py
		if distance < 0 {
			textID, distance = obj.TrackerDialogueTextIDs.North, -distance
		}
		if g.openPackText(textID) {
			g.setDlgVarGlyphs(dq3data.TxtVarItem, digitGlyphs(distance))
			g.trackedWorldLocatorStage = 3
		}
	case 3:
		g.trackedWorldLocatorStage, g.trackedWorldLocatorID = 0, ""
	}
}
