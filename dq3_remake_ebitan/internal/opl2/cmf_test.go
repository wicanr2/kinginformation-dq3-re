package opl2

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMBGPlayback(t *testing.T) {
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../../assets_raw"
	}
	mbg, err := os.ReadFile(filepath.Join(dir, "MBG.MCX"))
	if err != nil {
		t.Skipf("無 MBG.MCX:%v", err)
	}
	tracks := MBGTracks(mbg)
	if len(tracks) < 10 { // 18 軌
		t.Fatalf("MBG.MCX 應多軌,得 %d", len(tracks))
	}
	t.Logf("MBG.MCX 解出 %d 軌", len(tracks))
	// 播 track2(城鎮曲)→ 事件數 > 0、render 非靜音
	opl := New(44100)
	cmf := LoadCMF(tracks[2], opl, 44100, 0)
	if cmf == nil || cmf.EventCount() == 0 {
		t.Fatal("track2 載入/解事件失敗")
	}
	buf := make([]int16, 44100) // 1 秒
	cmf.Render(buf, len(buf), true)
	nonZero, peak := 0, int16(0)
	for _, s := range buf {
		if s != 0 {
			nonZero++
		}
		if s > peak {
			peak = s
		}
	}
	if nonZero == 0 {
		t.Fatal("track2 render 全靜音(疑事件/合成有誤)")
	}
	t.Logf("track2 ✓:%d 事件、1s render %d/%d 非零、峰值 %d", cmf.EventCount(), nonZero, len(buf), peak)
}
