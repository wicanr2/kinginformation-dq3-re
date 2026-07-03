package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// 晝夜相位步數推進:地表每走一步 dnStep++,每 60 步推進一相位,240 步一循環(白天→黃昏→黑夜→黎明)。
// 對齊 C main.c DN_PHASE_STEPS=60 / 4 相位。純狀態(不需素材;applyDaynightPalette 對 nil scene 有防護)。
func TestDaynightStepAdvance(t *testing.T) {
	g := &Game{}
	if g.dnPhase != 0 || g.isNight() {
		t.Fatal("初始應為白天(相位 0)")
	}
	step := func(n int) {
		for i := 0; i < n; i++ {
			g.advanceDaynight()
		}
	}
	// 59 步:仍白天;第 60 步:進黃昏
	step(59)
	if g.dnPhase != 0 {
		t.Fatalf("59 步後應仍白天,得相位 %d", g.dnPhase)
	}
	step(1)
	if g.dnPhase != 1 {
		t.Fatalf("60 步後應為黃昏(1),得 %d", g.dnPhase)
	}
	// 再 60 步(累計 120)→ 黑夜
	step(60)
	if g.dnPhase != 2 || !g.isNight() {
		t.Fatalf("120 步後應為黑夜(2)且 isNight,得相位 %d isNight=%v", g.dnPhase, g.isNight())
	}
	// 再 60(180)→ 黎明;再 60(240)→ 回白天(循環)
	step(60)
	if g.dnPhase != 3 {
		t.Fatalf("180 步後應為黎明(3),得 %d", g.dnPhase)
	}
	step(60)
	if g.dnPhase != 0 || g.dnStep != 0 {
		t.Fatalf("240 步後應回白天(0)且 step 歸零,得相位 %d step %d", g.dnPhase, g.dnStep)
	}
	t.Log("步數推進:60步/相位、240步循環、120步=黑夜 ✓")
}

// setDaynight / toggleDaynight(拉那魯達/黑暗之燈用):強制相位,城外不重載(nil scene 防護)。
func TestDaynightForce(t *testing.T) {
	g := &Game{}
	g.setDaynight(2) // 黑暗之燈:強制黑夜
	if g.dnPhase != 2 || !g.isNight() || g.dnStep != 0 {
		t.Fatalf("setDaynight(2) 應黑夜且 step 歸零,得相位 %d isNight=%v step %d", g.dnPhase, g.isNight(), g.dnStep)
	}
	g.toggleDaynight() // 拉那魯達:黑夜→白天
	if g.dnPhase != 0 || g.isNight() {
		t.Fatalf("toggle 自黑夜應回白天,得相位 %d", g.dnPhase)
	}
	g.toggleDaynight() // 白天→黑夜
	if g.dnPhase != 2 || !g.isNight() {
		t.Fatalf("toggle 自白天應到黑夜,得相位 %d", g.dnPhase)
	}
	t.Log("setDaynight/toggle 正確 ✓")
}

// 視覺對照 dump:阿里阿罕(CTY00 sec0)白天 vs 黑夜兩張 PNG,供肉眼確認夜間 NPC 變少 + 畫面變暗。
// 平常 skip;設 DQ3_DUMP_DN=1 + DN_OUT=<gitignored 夾> 才跑。
func TestDumpDayNight(t *testing.T) {
	if os.Getenv("DQ3_DUMP_DN") == "" {
		t.Skip("設 DQ3_DUMP_DN=1 才執行(晝夜視覺對照 dump)")
	}
	out := os.Getenv("DN_OUT")
	if out == "" {
		t.Fatal("需設 DN_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	g, err := NewGame(os.DirFS(dir), nil)
	if err != nil {
		t.Fatal(err)
	}
	dump := func(name string) {
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.Set(i%ScreenW, i/ScreenW, color.RGBA{g.rgba[o], g.rgba[o+1], g.rgba[o+2], 255})
		}
		f, err := os.Create(out + "/" + name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s dumped", name)
	}
	// 白天進阿里阿罕(sec0):24 NPC、原色
	g.showTitle = false // 關標題畫面,才渲染城鎮(否則 renderFrame 畫 title)
	g.dnPhase, g.dnStep = 0, 0
	g.enterTownCty(0)
	dayN := len(g.town.npcs)
	dump("aliahan_day")
	// 切黑夜(黑暗之燈語意):重載 NPC(→6)+ palette 調暗
	g.setDaynight(2)
	nightN := len(g.town.npcs)
	dump("aliahan_night")
	t.Logf("阿里阿罕 sec0:白天 NPC=%d、黑夜 NPC=%d", dayN, nightN)
	if nightN >= dayN {
		t.Fatalf("黑夜 NPC(%d)未少於白天(%d)", nightN, dayN)
	}
}
