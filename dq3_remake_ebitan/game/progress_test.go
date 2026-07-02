package game

import "testing"

// 進度階段:連續已完成里程碑數,第一個未完成處停(對拍 dq3_progress_stage)。
func TestProgressStage(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	if s := g.progressStage(); s != 0 {
		t.Errorf("空旗標階段 %d, want 0", s)
	}
	// 設 START → 階段 1
	g.progressSet(msStart)
	if s := g.progressStage(); s != 1 {
		t.Errorf("START 後階段 %d, want 1", s)
	}
	// 連設 THIEF_KEY/MAGIC_BALL → 階段 3
	g.progressSet(msThiefKey)
	g.progressSet(msMagicBal)
	if s := g.progressStage(); s != 3 {
		t.Errorf("連續 3 里程碑後階段 %d, want 3", s)
	}
	// 跳著設 DHAMA(略過 ROMALY)→ 階段仍停在 3(ROMALY 未完成)
	g.progressSet(msDhama)
	if s := g.progressStage(); s != 3 {
		t.Errorf("略過 ROMALY 設 DHAMA,階段 %d, want 3(第一個未完成處停)", s)
	}
	// 補上 ROMALY → 連續到 DHAMA(階段 5)
	g.progressSet(msRomaly)
	if s := g.progressStage(); s != 5 {
		t.Errorf("補 ROMALY 後階段 %d, want 5", s)
	}
}

// eff_flag:RAINBOW/DESCEND 里程碑對應既有 EXE 旗標(不是自身 id)。
func TestProgressEffFlag(t *testing.T) {
	g := &Game{flags: map[int]bool{}}
	g.progressSet(msDescend)
	if !g.flags[flagDescended] {
		t.Error("DESCEND 應設 flagDescended(0x13a)")
	}
	if g.flags[msDescend] {
		t.Error("DESCEND 不應設自身 id(應走 eff_flag)")
	}
	if !g.progressDone(msDescend) {
		t.Error("progressDone(DESCEND) 應由 flagDescended 反查為真")
	}
	g.progressSet(msRainbow)
	if !g.flags[flagRainbowSynthed] || !g.progressDone(msRainbow) {
		t.Error("RAINBOW 應設/反查 flagRainbowSynthed(0x139)")
	}
}

// progressName:階段名夾範圍。
func TestProgressName(t *testing.T) {
	if progressName(0) != "未開始" || progressName(msCount) != "索瑪(終盤)" {
		t.Error("階段名邊界錯")
	}
	if progressName(-5) != "未開始" || progressName(999) != "索瑪(終盤)" {
		t.Error("階段名未夾範圍")
	}
}
