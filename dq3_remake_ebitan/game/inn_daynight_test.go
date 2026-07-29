package game

import "testing"

func TestInnSuccessReturnsToDayButFailedPaymentDoesNot(t *testing.T) {
	g := &Game{
		curCty:   0,
		heroHP:   1,
		heroGold: 2,
		heroInit: true,
		dnPhase:  2,
		dnStep:   47,
	}
	g.ensureHeroStats()
	g.openFacility(0) // 阿里阿罕旅店：2G
	if g.dnPhase != 0 || g.dnStep != 0 {
		t.Fatalf("住宿成功應依 EXE [0x526c]/[0x251d] 回白天零時刻：phase=%d step=%d",
			g.dnPhase, g.dnStep)
	}

	g.heroGold, g.dnPhase, g.dnStep = 1, 2, 47
	g.openFacility(0)
	if g.dnPhase != 2 || g.dnStep != 47 {
		t.Fatalf("金錢不足不得改日夜：phase=%d step=%d", g.dnPhase, g.dnStep)
	}
}
