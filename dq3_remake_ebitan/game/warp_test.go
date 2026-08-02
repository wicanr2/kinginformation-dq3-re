package game

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
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

func TestCTY42WorldExitUsesOriginalDestinationWithoutFacingNudge(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = filepath.Join("..", "..", "assets_raw")
	}
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Skipf("original %s unavailable: %v", name, err)
		}
		return b
	}
	exe, cty := read("DQ3.EXE"), read("CTY42.DAT")
	section := int(binary.LittleEndian.Uint16(cty[0:2]))
	transitionTable := section + int(binary.LittleEndian.Uint16(cty[section+0x0c:section+0x0e]))
	if section != 0x02 || transitionTable != 0x2b ||
		!bytes.Equal(cty[transitionTable+8:transitionTable+12], []byte{42, 0xff, 213, 123}) {
		t.Fatalf("CTY42 世界出口 raw drift：section=%#x table=%#x raw=%x",
			section, transitionTable, cty[transitionTable+8:transitionTable+12])
	}
	// IDA linear 0x13689 / file 0x49f9：重新取 transition X/Y；兩者皆零才回退
	// remembered coordinates，否則直接寫 DGROUP 0x4f2f/0x4f31。此段不讀 facing。
	if raw := exe[0x49f9:0x4a41]; !bytes.Equal(raw, []byte{
		0x06, 0xa1, 0x36, 0x25, 0x8e, 0xc0, 0x8b, 0x16, 0x24, 0x0b,
		0xbf, 0x0c, 0x00, 0x03, 0xfa, 0x26, 0x8b, 0x35, 0x03, 0xf2,
		0x8a, 0x1e, 0x55, 0x0b, 0x80, 0xe3, 0x1f, 0x32, 0xff, 0xd1,
		0xe3, 0xd1, 0xe3, 0x03, 0xf3, 0x26, 0x8a, 0x44, 0x02, 0x26,
		0x8a, 0x5c, 0x03, 0x07, 0x32, 0xe4, 0x32, 0xff, 0x3d, 0x00,
		0x00, 0x75, 0x0c, 0x83, 0xfb, 0x00, 0x75, 0x07, 0xa1, 0x53,
		0x50, 0x8b, 0x1e, 0x55, 0x50, 0xa3, 0x2f, 0x4f, 0x89, 0x1e,
		0x31, 0x4f,
	}) {
		t.Fatalf("DQ3.EXE transition destination consumer file0x49f9 drifted: %x", raw)
	}

	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	sc, err := loadTownSceneSec(g.assets, g.worldPal, g.manBLS,
		42, mapBlkNum[42], 0, 0, g.storyFlag)
	if err != nil {
		t.Fatal(err)
	}
	g.cur, g.town, g.curCty, g.inTown = sc, sc, 42, true
	g.showTitle = false
	g.px, g.py, g.facing = 7, 12, 0
	g.tryTransition()
	if g.inTown || g.curCty != -1 || g.px != 213 || g.py != 123 ||
		g.overPx != 213 || g.overPy != 123 {
		t.Fatalf("CTY42 世界出口 transaction 錯：town=%v cty=%d pos=(%d,%d) remembered=(%d,%d)",
			g.inTown, g.curCty, g.px, g.py, g.overPx, g.overPy)
	}
	if !g.worldEntranceGrace {
		t.Fatal("CTY42 離場未建立一次性的 world entrance grace")
	}
	for frames := 0; g.cd > 0 && frames < 100; frames++ {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
	if g.cd > 0 {
		t.Fatalf("CTY42 離場 cooldown 100 frame 後仍為 %d", g.cd)
	}
	if err := g.step(InputState{DirHeld: 2, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.inTown || g.curCty != -1 || g.px != 212 || g.py != 123 || g.worldEntranceGrace {
		t.Fatalf("CTY42 離場首步未在不重入的情況下消耗 grace：town=%v cty=%d pos=(%d,%d) grace=%v",
			g.inTown, g.curCty, g.px, g.py, g.worldEntranceGrace)
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
