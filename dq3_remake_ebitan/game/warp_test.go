package game

import "testing"

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

// owPortalResolve:非 portal 點回 -1;portal 點無旗標 → 預設變體;set 對應旗標 → 該城。對拍 dq3_owportal_resolve。
func TestOwPortalResolve(t *testing.T) {
	// 非 portal 點
	if got := owPortalResolve(1, 1, nil); got != -1 {
		t.Errorf("非 portal (1,1) 回 %d, want -1", got)
	}
	// (210,64) 無旗標 → 預設 83
	if got := owPortalResolve(210, 64, map[int]bool{}); got != 83 {
		t.Errorf("(210,64) 無旗標回 %d, want 83", got)
	}
	// (210,64) set flag 0x47 → 59(鏈中;0x23 未 set)
	if got := owPortalResolve(210, 64, map[int]bool{0x47: true}); got != 59 {
		t.Errorf("(210,64) flag0x47 回 %d, want 59", got)
	}
	// (210,64) set 0x23(鏈首)→ 58(優先)
	if got := owPortalResolve(210, 64, map[int]bool{0x23: true, 0x47: true}); got != 58 {
		t.Errorf("(210,64) flag0x23 優先回 %d, want 58", got)
	}
	// (82,165) 無旗標 → 47
	if got := owPortalResolve(82, 165, map[int]bool{}); got != 47 {
		t.Errorf("(82,165) 無旗標回 %d, want 47", got)
	}
	// (82,165) set 0x13 → 75
	if got := owPortalResolve(82, 165, map[int]bool{0x13: true}); got != 75 {
		t.Errorf("(82,165) flag0x13 回 %d, want 75", got)
	}
}
