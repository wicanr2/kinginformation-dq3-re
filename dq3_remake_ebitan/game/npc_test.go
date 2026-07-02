package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// walkableAttr:tile idx 0 → attr 0(可走)的最小 BlockAttr(測試用)。
func walkableAttr() *dq3data.BlockAttr { return &dq3data.BlockAttr{A: []uint16{0}} }

// NPC 遊走:靜止 NPC(無 MOVE bit)不動;可動 NPC 多幀後會離開起點;牆/界外/近玩家不穿。
func TestNpcWander(t *testing.T) {
	// 空曠 10×10 地圖(全可走 tile 0,attr 全 0),玩家遠在角落。
	w, h := 10, 10
	sc := &Scene{
		w: w, h: h,
		tileAt: func(x, y int) int { return 0 },
		attr:   walkableAttr(),
	}
	sc.npcRng.Seed(1)
	g := &Game{cur: sc, inTown: true, px: 0, py: 0}

	// 靜止 NPC(ctrl 無 MOVE bit):3000 幀不動。
	sc.npcs = []npcInst{{x: 5, y: 5, ctrl: 0}}
	for i := 0; i < 3000; i++ {
		g.npcTick()
	}
	if sc.npcs[0].x != 5 || sc.npcs[0].y != 5 {
		t.Errorf("靜止 NPC 竟移動到 (%d,%d)", sc.npcs[0].x, sc.npcs[0].y)
	}

	// 可動 NPC(MOVE bit set):多幀後應曾離開起點,且永遠在界內。
	sc.npcs = []npcInst{{x: 5, y: 5, ctrl: npcMoveBit}}
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
	w, h := 10, 10
	sc := &Scene{w: w, h: h, tileAt: func(x, y int) int { return 0 }, attr: walkableAttr()}
	sc.npcRng.Seed(7)
	// 玩家在 (5,5),NPC 在 (5,7) 朝上(dir2);距離 Y=2 <3 → 不應往上踏進 (5,6)。
	g := &Game{cur: sc, inTown: true, px: 5, py: 5}
	sc.npcs = []npcInst{{x: 5, y: 7, ctrl: npcMoveBit | 2}} // dir=2(上)
	for i := 0; i < 2000; i++ {
		g.npcTick()
		if sc.npcs[0].y <= 6 && sc.npcs[0].x == 5 {
			t.Fatalf("NPC 越過近玩家閘,踏到 (%d,%d)", sc.npcs[0].x, sc.npcs[0].y)
		}
	}
}
