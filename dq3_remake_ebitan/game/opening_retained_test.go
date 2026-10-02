package game

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestRetainedRowsPauseAfterScrollAndFinalHold(t *testing.T) {
	d := Dialogue{layout: gamepack.WindowLayout{Columns: 1, LinesPerPage: 2, GlyphHoldFrames: 1}, prelude: &gamepack.OpeningPrelude{
		GlyphStepX: 24, VariableCodeWords: 1, ReturnMode: "automatic_after_reveal",
		TextFlow: gamepack.RetainedTextFlow{Mode: "retained_rows", ScrollStepPixels: 8, ScrollSteps: 2, ScrollHoldFrames: 2}}}
	d.openRecord([]uint16{1, dq3data.TxtNL, 2, dq3data.TxtPage, 3})
	d.Tick()
	d.Tick()
	d.Tick()
	if d.waitingForConfirm() || len(d.retained.ops) != 3 {
		t.Fatal("最後一行的FFFC未先捲動")
	}
	d.Advance()
	if d.pos != 0 {
		t.Fatal("未到等待點就接受確認")
	}
	for i := 0; i < 4; i++ {
		d.Tick()
	}
	if !d.waitingForConfirm() || len(d.retained.ops) != 4 {
		t.Fatal("捲動操作期間的等待未守住")
	}
	d.Advance()
	d.Tick()
	if d.retained.row != 1 || len(d.retained.ops) != 5 || d.retained.ops[4].y != dq3data.GlyphPx {
		t.Fatal("確認清掉前文或改變最後一行游標")
	}
	d.Tick()
	if !d.retained.scrolling || !d.open {
		t.Fatal("EOF未先捲動")
	}
	for i := 0; i < 3; i++ {
		d.Tick()
		d.Advance()
		if !d.open {
			t.Fatal("提前確認跳過最後捲動hold")
		}
	}
	d.Tick()
	if d.open {
		t.Fatal("捲動完成仍多等確認")
	}
}

func TestRetainedRowsUsesEncodedNewlines(t *testing.T) {
	d := Dialogue{layout: gamepack.WindowLayout{Columns: 1, LinesPerPage: 2, GlyphHoldFrames: 1}, prelude: &gamepack.OpeningPrelude{
		GlyphStepX: 24, VariableCodeWords: 1, ReturnMode: "automatic_after_reveal", TextFlow: gamepack.RetainedTextFlow{Mode: "retained_rows", ScrollStepPixels: 8, ScrollSteps: 2, ScrollHoldFrames: 1}}}
	d.openRecord([]uint16{1, 2, 3})
	for i := 0; i < 3; i++ {
		d.Tick()
	}
	for i, op := range d.retained.ops {
		if op.y != 0 || op.x != i*d.prelude.GlyphStepX {
			t.Fatal("引擎依columns猜測換行")
		}
	}
	d.Tick()
	if d.open {
		t.Fatal("未滿最後一行的EOF多捲動或等待確認")
	}
	d.openRecord([]uint16{4})
	if d.retained != nil {
		t.Fatal("新record保留上一段畫布操作")
	}
}

// 從同一冷啟動收據重播，不寫入座標、文字游標或場景狀態。
func TestDosgolemOpeningRetainedRowsComparison(t *testing.T) {
	dir := os.Getenv("DQ3_DOSGOLEM_NEWGAME_DIR")
	if os.Getenv("DQ3_DOSGOLEM_RETAINED_COMPARE") != "1" || dir == "" {
		t.Skip("需明確啟用 dosgolem 生日續頁與四次捲動收據")
	}
	var receipt struct {
		Scenario       string `json:"scenario"`
		OriginalSHA    string `json:"original_sha256"`
		OriginalSize   int    `json:"original_size"`
		Revision       string `json:"upstream_revision_observed"`
		StateInjection *bool  `json:"game_state_injection"`
		Seed           struct {
			Value      string `json:"seed"`
			Configured bool   `json:"configured_before_execution"`
			OnlyRNG    bool   `json:"only_rng_state_modified"`
			OtherState *bool  `json:"other_gameplay_state_injection"`
		} `json:"test_rng_seed_control"`
		Inputs []struct {
			Scan string `json:"scan"`
		} `json:"player_input"`
		Artifacts []struct {
			Path string `json:"path"`
			SHA  string `json:"sha256"`
			Size int    `json:"size"`
		} `json:"artifacts"`
	}
	prefix := "issue4-birthday-pages"
	raw, err := os.ReadFile(filepath.Join(dir, prefix+"-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Scenario != "birthday_continue" || receipt.OriginalSize != 115282 || receipt.OriginalSHA != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" || receipt.Revision != "2f44a68ebfc54b28fb15dd4a34510b0b04a5415d" || receipt.StateInjection == nil || *receipt.StateInjection || receipt.Seed.Value != "0x1357" || !receipt.Seed.Configured || !receipt.Seed.OnlyRNG || receipt.Seed.OtherState == nil || *receipt.Seed.OtherState {
		t.Fatal("原版收據身份、種子或未注入前提不符")
	}
	want := []string{"0x1c", "0x1c", "0x48", "0x4b", "0x1c", "0x1c", "0x4d", "0x50", "0x1c", "0x48", "0x4b", "0x1c", "0x48", "0x1c", "0x1c", "0x1c", "0x1c", "0x1c", "0x1c"}
	if len(receipt.Inputs) != len(want) {
		t.Fatal("原版19次輸入不符")
	}
	for i, input := range receipt.Inputs {
		if input.Scan != want[i] {
			t.Fatalf("原版第%d次輸入不符", i+1)
		}
	}
	indexed := map[string]bool{}
	for _, a := range receipt.Artifacts {
		if filepath.Base(a.Path) != a.Path {
			t.Fatal("收據路徑越界")
		}
		b, err := os.ReadFile(filepath.Join(dir, a.Path))
		if err != nil {
			t.Fatal(err)
		}
		if len(b) != a.Size || fmt.Sprintf("%x", sha256.Sum256(b)) != a.SHA {
			t.Fatalf("收據雜湊不符：%s", a.Path)
		}
		indexed[a.Path] = true
	}
	t.Setenv("DQ3_SAVE", filepath.Join(t.TempDir(), "retained-save.json"))
	g, err := newProductionTraceGame(os.DirFS(spineAssetsDir(t)))
	if err != nil {
		t.Fatal(err)
	}
	g.prng.Seed(0x1357)
	for _, input := range receipt.Inputs[:17] {
		in := InputState{DirHeld: -1, DirEdge: -1}
		switch input.Scan {
		case "0x1c":
			in.Confirm = true
		case "0x48":
			in.DirEdge = 1
		case "0x50":
			in.DirEdge = 0
		case "0x4b":
			in.DirEdge = 2
		case "0x4d":
			in.DirEdge = 3
		}
		if err := g.step(in); err != nil {
			t.Fatal(err)
		}
	}
	idle := InputState{DirHeld: -1, DirEdge: -1}
	wait := func(done func() bool) {
		t.Helper()
		for i := 0; i < 2000 && !done(); i++ {
			if err := g.step(idle); err != nil {
				t.Fatal(err)
			}
		}
		if !done() {
			t.Fatal("正常輸入等待超過上限")
		}
	}
	compare := func(name string) {
		t.Helper()
		original := prefix + "-" + name + ".png"
		if !indexed[original] {
			t.Fatalf("原版圖片未登記：%s", name)
		}
		g.renderFrame()
		compareDosgolemRasterFrame(t, g, dir, original, prefix+"-remake-"+name+".png", name)
	}
	wait(func() bool { return g.dlg.waitingForConfirm() })
	compare("stable")
	if err := g.step(InputState{DirHeld: -1, DirEdge: -1, Confirm: true}); err != nil {
		t.Fatal(err)
	}
	wait(func() bool { f := g.dlg.retained; return f != nil && f.word == len(g.dlg.buf) && len(f.glyphs) == 0 })
	compare("birthday-tail-before-eof")
	for phase := 1; phase <= g.dlg.prelude.TextFlow.ScrollSteps; phase++ {
		wait(func() bool {
			n := 0
			if g.dlg.retained != nil {
				for _, op := range g.dlg.retained.ops {
					if op.scroll {
						n++
					}
				}
			}
			return n == phase
		})
		if !g.dlg.open || g.openingIdx != 0 || g.inTown || g.prng.State() != 0x356d {
			t.Fatal("捲動中提早交易出生或消耗亂數")
		}
		compare(fmt.Sprintf("birthday-scroll-%d", phase))
	}
	wait(func() bool { return g.openingIdx == 1 && g.dlg.waitingForConfirm() })
	if !reflect.DeepEqual(g.dlg.buf, g.dlg.tx.Record(83)) || g.prng.State() != 0x356d || g.dayNightClock() != g.dayNightCycle.InitialClock {
		t.Fatal("生日返回後未自然抵達既有下一節點")
	}
	t.Log("17次創角輸入與第18次確認，6張完整RGB；捲動期間無出生交易或RNG消耗；EOF無額外確認")
}
