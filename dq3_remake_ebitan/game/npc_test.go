package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// walkableAttr:tile idx 0 → attr 0(可走)的最小 BlockAttr(測試用)。
func walkableAttr() *dq3data.BlockAttr { return &dq3data.BlockAttr{A: []uint16{0}} }

func npcTestPack(t *testing.T) *gamepack.Pack {
	t.Helper()
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	return p
}

// NPC 遊走:靜止 NPC(無 MOVE bit)不動;可動 NPC 多幀後會離開起點;牆/界外/近玩家不穿。
func TestNpcWander(t *testing.T) {
	p := npcTestPack(t)
	// 空曠 10×10 地圖(全可走 tile 0,attr 全 0),玩家遠在角落。
	w, h := 10, 10
	sc := &Scene{
		w: w, h: h,
		hiMap:  make([]byte, w*h),
		tileAt: func(x, y int) int { return 0 },
		attr:   walkableAttr(),
	}
	sc.npcRng.Seed(1)
	g := &Game{cur: sc, inTown: true, pack: p, px: 0, py: 0}

	// 靜止 NPC(ctrl 無 MOVE bit):3000 幀不動。
	sc.npcs = []npcInst{{x: 5, y: 5, ctrl: 0}}
	for i := 0; i < 3000; i++ {
		g.npcTick()
	}
	if sc.npcs[0].x != 5 || sc.npcs[0].y != 5 {
		t.Errorf("靜止 NPC 竟移動到 (%d,%d)", sc.npcs[0].x, sc.npcs[0].y)
	}

	// 可動 NPC(MOVE bit set):多幀後應曾離開起點,且永遠在界內。
	sc.npcs = []npcInst{{x: 5, y: 5, ctrl: p.Characters.NPCMotion.MoveMask}}
	moved := false
	for i := 0; i < 3000; i++ {
		g.npcTick()
		n := sc.npcs[0]
		if n.x < 0 || n.x >= w || n.y < 0 || n.y >= h {
			t.Fatalf("NPC 走出界:(%d,%d)", n.x, n.y)
		}
		if n.x != 5 || n.y != 5 {
			moved = true
		}
	}
	if !moved {
		t.Error("可動 NPC 3000 幀都沒動過(遊走未生效)")
	}
}

// 近玩家閘:NPC 緊鄰玩家(距離 <3)沿該軸不走(不擠進玩家 2 格內)。
func TestNpcNearPlayerGate(t *testing.T) {
	p := npcTestPack(t)
	w, h := 10, 10
	sc := &Scene{w: w, h: h, hiMap: make([]byte, w*h), tileAt: func(x, y int) int { return 0 }, attr: walkableAttr()}
	sc.npcRng.Seed(7)
	// 玩家在 (5,5),NPC 在 (5,7) 朝上(dir2);距離 Y=2 <3 → 不應往上踏進 (5,6)。
	g := &Game{cur: sc, inTown: true, pack: p, px: 5, py: 5}
	sc.npcs = []npcInst{{x: 5, y: 7, ctrl: p.Characters.NPCMotion.MoveMask | 2}} // dir=2(上)
	for i := 0; i < 2000; i++ {
		g.npcTick()
		if sc.npcs[0].y <= 6 && sc.npcs[0].x == 5 {
			t.Fatalf("NPC 越過近玩家閘,踏到 (%d,%d)", sc.npcs[0].x, sc.npcs[0].y)
		}
	}
}

func TestNpcFacingAndIdleAnimation(t *testing.T) {
	p := npcTestPack(t)
	if got := npcCtrlFacing(1); got != 2 {
		t.Fatalf("ctrl 左應轉 renderer 左，got %d", got)
	}
	if got := npcCtrlFacing(2); got != 1 {
		t.Fatalf("ctrl 上應轉 renderer 上，got %d", got)
	}
	sc := &Scene{w: 10, h: 10, hiMap: make([]byte, 100), tileAt: func(x, y int) int { return 0 }, attr: walkableAttr()}
	g := &Game{cur: sc, inTown: true, pack: p, px: 0, py: 0}
	sc.npcs = []npcInst{{x: 5, y: 5, ctrl: 0}}
	for i := 0; i < 18; i++ {
		g.npcTick()
	}
	if sc.npcs[0].walk != 1 {
		t.Fatal("靜止 NPC 經 18 幀應切換角色步行幀")
	}
	n := &sc.npcs[0]
	n.ctrl = p.Characters.NPCMotion.MoveMask | 3
	g.npcTryStep(0)
	if n.facing != 3 {
		t.Fatalf("往右移動應面向右，got %d", n.facing)
	}
}

func TestNPCMotionTerrainAndMissingMapGate(t *testing.T) {
	for _, attribute := range []uint16{0, 2, 0x20, 0x40, 0x80, 0x100} {
		p := npcTestPack(t)
		sc := &Scene{w: 10, h: 10, hiMap: make([]byte, 100), tileAt: func(x, y int) int { return 0 }, attr: &dq3data.BlockAttr{A: []uint16{attribute}}, npcs: []npcInst{{x: 5, y: 5, ctrl: p.Characters.NPCMotion.MoveMask | 3}}}
		g := &Game{cur: sc, inTown: true, pack: p, px: 0, py: 0}
		got := g.npcTryStep(0)
		if got != (attribute&0xff == 0) {
			t.Fatalf("original AL terrain gate differs for %04x", attribute)
		}
		if !got && (sc.npcs[0].x != 5 || sc.npcs[0].y != 5) {
			t.Fatal("failed terrain transaction changed position")
		}
	}
	p := npcTestPack(t)
	sc := &Scene{w: 10, h: 10, tileAt: func(x, y int) int { return 0 }, attr: walkableAttr(), npcs: []npcInst{{x: 5, y: 5, ctrl: p.Characters.NPCMotion.MoveMask}}}
	sc.npcRng.Seed(1)
	g := &Game{cur: sc, inTown: true, pack: p, px: 0, py: 0}
	g.npcTick()
	if sc.npcRng.State() != 1 || sc.npcs[0].x != 5 || sc.npcs[0].y != 5 {
		t.Fatal("missing typed map did not stop movement")
	}
}
