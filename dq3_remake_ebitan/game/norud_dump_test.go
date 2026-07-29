package game

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// TestDumpNorudGuidedPassage 產出諾魯德密道事件的 runtime 畫面證據。
// 正式玩家輸入路徑由 TestOpeningProductionInputTrace 鎖定；本測試只在明確要求時
// 重播同一個已驗證的事件交易，並以真正 CTY62 素材與 renderFrame 產圖。
func TestDumpNorudGuidedPassage(t *testing.T) {
	if os.Getenv("DQ3_DUMP_NORUD") == "" {
		t.Skip("設 DQ3_DUMP_NORUD=1 才執行諾魯德視覺驗證 dump")
	}
	out := os.Getenv("NORUD_OUT")
	if out == "" {
		t.Fatal("需設 NORUD_OUT")
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	g, event, guide := norudTestGame(t)
	g.inventory = append(g.inventory, event.RequiredItemRawID)

	dump := func(name string) {
		t.Helper()
		g.renderFrame()
		img := image.NewRGBA(image.Rect(0, 0, ScreenW, ScreenH))
		for i := 0; i < ScreenW*ScreenH; i++ {
			o := i * 4
			img.SetRGBA(i%ScreenW, i/ScreenW,
				color.RGBA{R: g.rgba[o], G: g.rgba[o+1], B: g.rgba[o+2], A: 255})
		}
		path := filepath.Join(out, name+".png")
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			_ = f.Close()
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
		t.Logf("%s", path)
	}

	if !g.talkGuidedPassage(guide) {
		t.Fatal("諾魯德正式 selector 未接手")
	}
	closeGuidedDialogue(g)
	if g.guidedPassageStage != guidedPassageAcceptance {
		t.Fatalf("stage=%d, want acceptance", g.guidedPassageStage)
	}
	g.dlg.Advance() // rec88 第二頁明確顯示「波魯多加王的信」
	if !g.dlg.open {
		t.Fatal("rec88 王信頁不應在第一個 Advance 後關閉")
	}
	dump("norud_letter_acceptance")

	for g.dlg.open {
		g.dlg.Advance()
	}
	g.advanceGuidedPassageDialogue()
	runGuidedAnimation(t, g, guidedPassageAwaitTrigger)
	g.px, g.py = event.CompletionTrigger.Tiles[0].X-1,
		event.CompletionTrigger.Tiles[0].Y+1
	dump("norud_guide_waiting")

	trigger := event.CompletionTrigger.Tiles[0]
	g.px, g.py = trigger.X, trigger.Y
	if !g.tryGuidedPassageTrigger() {
		t.Fatal("諾魯德 handler57 trigger 未開啟 rec90")
	}
	dump("norud_passage_trigger")

	closeGuidedDialogue(g)
	runGuidedAnimation(t, g, guidedPassageIdle)
	dump("norud_passage_complete")
}
