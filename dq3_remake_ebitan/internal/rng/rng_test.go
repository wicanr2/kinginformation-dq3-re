package rng

import "testing"

func TestDOSRng(t *testing.T) {
	// 手算對拍:seed=1 → 首 Next(256):s=1+0x9018=0x9019、rol3=0x80cc、%256=0xcc=204。
	r := New(1)
	if got := r.Next(256); got != 204 {
		t.Fatalf("seed1 首 Next(256) 應 204(0xcc),得 %d", got)
	}
	// n=0 → 0
	if New(1).Next(0) != 0 {
		t.Fatal("Next(0) 應 0")
	}
	// 決定性:同種子同序列
	a, b := New(42), New(42)
	for i := 0; i < 20; i++ {
		if a.Next(100) != b.Next(100) {
			t.Fatal("同種子序列不一致")
		}
	}
	// 範圍
	rr := New(7)
	for i := 0; i < 1000; i++ {
		if v := rr.Next(6); v < 0 || v >= 6 {
			t.Fatalf("Next(6) 越界:%d", v)
		}
	}
	t.Log("DOS RNG 對拍 ✓(seed1→204、決定性、範圍)")
}
