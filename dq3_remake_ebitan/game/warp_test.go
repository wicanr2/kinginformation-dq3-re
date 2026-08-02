package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

// warpGet:有效項回目的、無效項(全0/越界)回 ok=false。對拍 dq3_warp_get。
func TestWarpGet(t *testing.T) {
	// warps[0]={2,40,1}
	if c, x, y, ok := warpGet(0); !ok || c != 2 || x != 40 || y != 1 {
		t.Errorf("warp[0]=(%d,%d,%d,%v), want (2,40,1,true)", c, x, y, ok)
	}
	// warps[8]={26,1,27}
	if c, x, y, ok := warpGet(8); !ok || c != 26 || x != 1 || y != 27 {
		t.Errorf("warp[8]=(%d,%d,%d,%v), want (26,1,27,true)", c, x, y, ok)
	}
	// warps[43]={0,0,0} → 無效
	if _, _, _, ok := warpGet(43); ok {
		t.Error("warp[43] 全0 應 ok=false")
	}
	// 越界
	if _, _, _, ok := warpGet(200); ok {
		t.Error("warp[200] 越界應 ok=false")
	}
	if _, _, _, ok := warpGet(-1); ok {
		t.Error("warp[-1] 應 ok=false")
	}
}

// worldEntranceResolve 對拍 DQ3.EXE file 0x396e..0x39f1 的 ordered flag chain。
func TestWorldEntranceResolve(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("載入內建 game pack: %v", err)
	}
	variants := pack.WorldEntranceVariants()
	resolve := func(x, y, layer int, flags map[int]bool) int {
		return worldEntranceResolve(x, y, layer, variants,
			func(flag int) bool { return flags[flag] })
	}
	if got := resolve(1, 1, 0, nil); got != -1 {
		t.Errorf("非入口 (1,1) 回 %d, want -1", got)
	}
	if got := resolve(210, 64, 0, map[int]bool{}); got != 58 {
		t.Errorf("(210,64) flag0x23 clear 回 %d, want 58", got)
	}
	allGates := map[int]bool{0x23: true, 0x47: true, 0x42: true, 0x48: true}
	if got := resolve(210, 64, 0, allGates); got != 59 {
		t.Errorf("(210,64) flag0x47 set 回 %d, want 59", got)
	}
	if got := resolve(210, 64, 0, map[int]bool{0x23: true, 0x42: true, 0x48: true}); got != 60 {
		t.Errorf("(210,64) flag0x47 clear/0x42 set 回 %d, want 60", got)
	}
	if got := resolve(210, 64, 0, map[int]bool{0x23: true, 0x48: true}); got != 61 {
		t.Errorf("(210,64) flag0x42 clear/0x48 set 回 %d, want 61", got)
	}
	if got := resolve(210, 64, 0, map[int]bool{0x23: true}); got != 83 {
		t.Errorf("(210,64) build gates clear 回 %d, want 83", got)
	}
	if got := resolve(210, 64, 1, map[int]bool{}); got != -1 {
		t.Errorf("(210,64) layer1 回 %d, want -1", got)
	}
	if got := resolve(54, 129, 0, map[int]bool{0x4d: true}); got != 71 {
		t.Errorf("(54,129) flag0x4d set 回 %d, want 71", got)
	}
	if got := resolve(54, 129, 0, map[int]bool{}); got != 72 {
		t.Errorf("(54,129) flag0x4d clear 回 %d, want 72", got)
	}
	// 勇氣試煉由獨立 mode state 控制，不能混進 story-flag 入口表。
	if got := resolve(82, 165, 0, map[int]bool{0x13: true}); got != -1 {
		t.Errorf("(82,165) 不應有入口變體，回 %d, want -1", got)
	}
}

func TestMerchantSettlementEntranceUsesOriginalNewGameStoryBits(t *testing.T) {
	pack, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatalf("載入內建 game pack: %v", err)
	}
	g := &Game{pack: pack}
	g.initStoryBits()
	resolve := func() int {
		return worldEntranceResolve(210, 64, 0, pack.WorldEntranceVariants(), g.storyFlag)
	}
	if got := resolve(); got != 58 {
		t.Fatalf("新遊戲原始 story bits 選 CTY%d, want CTY58", got)
	}
	g.setStoryFlag(0x23, true)
	if got := resolve(); got != 59 {
		t.Fatalf("交付商人 set flag0x23 後選 CTY%d, want CTY59", got)
	}
}
