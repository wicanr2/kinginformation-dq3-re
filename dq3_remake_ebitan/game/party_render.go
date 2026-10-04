package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// resetPartyTrail anchors every follower at the current player tile.  Scene
// changes, teleports and party transactions must call this so an old map's
// history can never leak into the new scene.
func (g *Game) resetPartyTrail() {
	entry := partyTrailEntry{x: g.px, y: g.py, facing: g.facing & 3}
	for i := range g.partyTrail {
		g.partyTrail[i] = entry
	}
	g.partyTrailLen = 0
}

// recordPartyTrail stores the player tile before a successful legal move.
func (g *Game) recordPartyTrail() {
	for i := len(g.partyTrail) - 1; i > 0; i-- {
		g.partyTrail[i] = g.partyTrail[i-1]
	}
	g.partyTrail[0] = partyTrailEntry{x: g.px, y: g.py, facing: g.facing & 3}
	if g.partyTrailLen < len(g.partyTrail) {
		g.partyTrailLen++
	}
}

// partySprite resolves the versioned class/gender mapping and caches the
// decoded BLS frames. Missing pack data is a deliberate no-render result.
func (g *Game) partySprite(classRaw, genderRaw int) *dq3data.CharSprite {
	if g.pack == nil || len(g.mstBLS) == 0 {
		return nil
	}
	asset, entry, ok := g.pack.PartySprite(classRaw, genderRaw)
	if !ok || asset == "" {
		return nil
	}
	if g.partySpriteCache == nil {
		g.partySpriteCache = make(map[[2]int]*dq3data.CharSprite)
	}
	key := [2]int{classRaw, genderRaw}
	if sprite := g.partySpriteCache[key]; sprite != nil {
		return sprite
	}
	sprite := dq3data.LoadCharSprite(g.mstBLS, entry)
	g.partySpriteCache[key] = sprite
	return sprite
}

func (g *Game) heroPartySprite() *dq3data.CharSprite {
	if sprite := g.partySprite(0, g.heroGender); sprite != nil {
		return sprite
	}
	return g.hero
}

type partyHUDActor struct {
	name   []int
	hp, mp int
	level  int
	class  int
	alive  bool
}

func (g *Game) partyHUDActors() []partyHUDActor {
	level, _, _, _, _ := g.heroStats()
	actors := []partyHUDActor{{
		name: append([]int(nil), g.heroName...), hp: g.heroHP, mp: g.heroMP,
		level: level, class: 0, alive: g.heroHP > 0,
	}}
	for _, member := range g.companions {
		if member == nil {
			continue
		}
		actors = append(actors, partyHUDActor{
			name: append([]int(nil), member.Name...), hp: member.CurHP,
			mp: member.CurMP, level: member.Level(), class: member.Class, alive: member.Alive(),
		})
	}
	return actors
}

func (g *Game) followerOrder() []int {
	order := make([]int, 0, len(g.companions))
	for _, alive := range []bool{true, false} {
		for actor := 0; actor <= len(g.companions); actor++ {
			if actor == g.partyLeader || g.actorAlive(actor) != alive {
				continue
			}
			order = append(order, actor)
		}
	}
	return order
}

func drawPartyCoffin(rgba []byte, x, y int, white dq3data.Color) {
	for row := 5; row < dq3data.CharH-2; row++ {
		for col := 9; col < dq3data.CharW-9; col++ {
			edge := row == 5 || row == dq3data.CharH-3 || col == 9 || col == dq3data.CharW-10
			crossV := col == dq3data.CharW/2 && row >= 8 && row < 16
			crossH := row == 10 && col >= dq3data.CharW/2-3 && col <= dq3data.CharW/2+3
			if edge || crossV || crossH {
				putPx(rgba, x+col, y+row, white)
			}
		}
	}
}

// drawPartyFollowers draws the original one-step-per-tile caterpillar. It is
// deliberately skipped while mounted or invisible, matching the vehicle and
// field-effect render paths that suppress the party occupant sprites.
func (g *Game) drawPartyFollowers(rgba []byte, camX, camY int, pal []dq3data.Color) {
	if len(g.companions) == 0 || g.remoaru != 0 || g.shipAboard || g.phoenixAboard {
		return
	}
	for slot, actor := range g.followerOrder() {
		trail := partyTrailEntry{x: g.px, y: g.py, facing: g.facing & 3}
		if g.partyTrailLen > slot {
			trail = g.partyTrail[slot]
		}
		dx, dy := (trail.x-camX)*TileW, (trail.y-camY)*TileH
		if !g.actorAlive(actor) {
			coffinColor := dq3data.Color{R: 255, G: 255, B: 255}
			if len(pal) > 15 {
				coffinColor = pal[15]
			}
			drawPartyCoffin(rgba, dx, dy, coffinColor)
			continue
		}
		var sprite *dq3data.CharSprite
		if actor == 0 {
			sprite = g.heroRole
			if sprite == nil {
				sprite = g.heroPartySprite()
			}
		} else {
			member := g.companions[actor-1]
			sprite = g.partySprite(member.Class, member.Gender)
		}
		if sprite == nil {
			continue
		}
		frame := (trail.facing&3)*dq3data.CharWalk + g.walk
		if frame < 0 || frame >= len(sprite.Frames) {
			frame = 0
		}
		blitSprite(rgba, dx, dy, sprite.Frames[frame], pal)
	}
}
