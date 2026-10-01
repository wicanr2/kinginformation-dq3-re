package game

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// 此對拍只比較原版 4-bit 色號與 renderer 使用資料包色盤後的輸出。
// 原版擷取可能位於淡入／淡出階段，不把色盤正規化當成 RGB 或 timing parity。
func TestOpeningCutsceneDosgolemReceipt(t *testing.T) {
	dir := os.Getenv("DQ3_DOSGOLEM_OPENING_DIR")
	if dir == "" {
		t.Skip("需由 tools/verify_dosgolem_opening.sh 重生原版收據")
	}
	raw, err := os.ReadFile(filepath.Join(dir, "receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var receipt struct {
		Method         string `json:"method"`
		OriginalSHA256 string `json:"original_sha256"`
		StateInjection bool   `json:"state_injection"`
		RNGComparison  bool   `json:"rng_comparison"`
		Input          []any  `json:"input"`
		Frames         []struct {
			Asset        string `json:"asset"`
			AssetSize    int64  `json:"asset_size"`
			AssetSHA256  string `json:"asset_sha256"`
			Step         uint64 `json:"step"`
			Width        int    `json:"width"`
			Height       int    `json:"height"`
			PixelsFile   string `json:"pixels_file"`
			PixelsSHA256 string `json:"pixels_sha256"`
		} `json:"frames"`
		Observations []struct {
			Asset        string `json:"asset"`
			Animation    bool   `json:"animation"`
			Y            int    `json:"y"`
			Step         uint64 `json:"step"`
			Ticks        uint64 `json:"irq0_ticks"`
			PixelsFile   string `json:"pixels_file"`
			PixelsSHA256 string `json:"pixels_sha256"`
			PNGFile      string `json:"png_file"`
			PNGSHA256    string `json:"png_sha256"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(raw, &receipt); err != nil {
		t.Fatal(err)
	}
	if receipt.Method != "dosgolem-natural-boot-planar" || receipt.StateInjection ||
		receipt.RNGComparison || len(receipt.Input) != 0 {
		t.Fatal("收據不是已審查的無輸入原版自然開機")
	}
	assets := spineAssetsDir(t)
	exe, err := os.ReadFile(filepath.Join(assets, "DQ3.EXE"))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(exe)) != receipt.OriginalSHA256 {
		t.Fatalf("原版輸入 hash 不符或讀取失敗：%v", err)
	}
	g, err := NewGame(os.DirFS(assets), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(g.openingSeq.Frames) != len(receipt.Frames) {
		t.Fatalf("remake 開場 %d 張，原版 %d 張", len(g.openingSeq.Frames), len(receipt.Frames))
	}
	g.StartOpeningCutscene() // 與 main.go 相同的正式 bootstrap。
	total := 0
	for i, original := range receipt.Frames[:len(receipt.Frames)-1] {
		frame := g.openingSeq.Frames[i]
		ref, ok := g.pack.Asset(frame.AssetKey)
		if !ok || ref.Path != original.Asset || ref.Size != original.AssetSize || ref.SHA256 != original.AssetSHA256 {
			t.Fatalf("第 %d 張素材順序／完整性不符：%+v，原版=%+v", i, ref, original)
		}
		if !g.openingActive || g.openingIndex != i || original.Width != ScreenW || original.Height != ScreenH ||
			(i > 0 && original.Step <= receipt.Frames[i-1].Step) {
			t.Fatalf("第 %d 張開場狀態或原版步數不符", i)
		}
		if filepath.Base(original.PixelsFile) != original.PixelsFile {
			t.Fatal("收據色號路徑必須是單一檔名")
		}
		pixels, err := os.ReadFile(filepath.Join(dir, original.PixelsFile))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(pixels)) != original.PixelsSHA256 {
			t.Fatalf("第 %d 張色號收據讀取／hash 失敗：%v", i, err)
		}
		if !bytes.Equal(pixels, g.openingPix[i]) {
			t.Fatalf("第 %d 張原版色號與 remake 不符", i)
		}
		if frame.Timing != nil {
			target := len(frame.Timing.FadeDeductions) * frame.Timing.FadeInStepTicks
			for g.openingActive && g.openingTick() < target {
				if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
					t.Fatal(err)
				}
			}
		}
		g.renderFrame()
		palette := g.openingPalette()
		for p, index := range pixels {
			if int(index) >= len(g.openingPal[i]) {
				t.Fatalf("原版色號 %d 越界", index)
			}
			c := palette[index]
			o := p * 4
			if g.rgba[o] != c.R || g.rgba[o+1] != c.G || g.rgba[o+2] != c.B || g.rgba[o+3] != 255 {
				t.Fatalf("第 %d 張 production renderer 像素 %d 不符", i, p)
			}
		}
		total += len(pixels)
		t.Logf("原版 %s 第 %d 步：%d 個色號與 production renderer 通過（RGB 色盤相位未比較）", ref.Path, original.Step, len(pixels))
		for g.openingActive && g.openingIndex == i {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatal(err)
			}
		}
	}
	// 第六幕固定步數可能位於繪圖中途；正式驗收改比原版完整的全部翻頁，
	// 保留原有固定步數擷取作診斷，不裁切或挑選動畫位置。
	logo := g.openingSeq.Frames[g.openingIndex]
	if logo.Overlay == nil || logo.Timing == nil {
		t.Fatal("第六幕缺動畫契約")
	}
	logoRef, ok := g.pack.Asset(logo.AssetKey)
	lastOriginal := receipt.Frames[len(receipt.Frames)-1]
	if !ok || logoRef.Path != lastOriginal.Asset || logoRef.Size != lastOriginal.AssetSize || logoRef.SHA256 != lastOriginal.AssetSHA256 {
		t.Fatal("第六幕素材身份不符")
	}
	pose := 0
	var lastStep, lastTick uint64
	for _, original := range receipt.Observations {
		if !original.Animation {
			continue
		}
		if original.Asset != logoRef.Path || pose >= logo.Overlay.Positions() || original.Y != logo.Overlay.StartY+pose*logo.Overlay.StepY ||
			(pose > 0 && (original.Step <= lastStep || original.Ticks-lastTick != uint64(logo.Overlay.StepTicks))) {
			t.Fatal("原版動畫位置、順序或等待 tick 不符")
		}
		target := len(logo.Timing.FadeDeductions)*logo.Timing.FadeInStepTicks + pose*logo.Overlay.StepTicks
		for g.openingActive && g.openingTick() < target {
			if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
				t.Fatal(err)
			}
		}
		if !g.openingActive || (g.openingTick()-len(logo.Timing.FadeDeductions)*logo.Timing.FadeInStepTicks)/logo.Overlay.StepTicks != pose {
			t.Fatal("正式輸入路徑遺漏動畫位置")
		}
		if filepath.Base(original.PixelsFile) != original.PixelsFile || filepath.Base(original.PNGFile) != original.PNGFile {
			t.Fatal("收據路徑越界")
		}
		pixels, err := os.ReadFile(filepath.Join(dir, original.PixelsFile))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(pixels)) != original.PixelsSHA256 || !bytes.Equal(pixels, g.openingPixels()) {
			t.Fatalf("動畫 y=%d 色號不符：%v", original.Y, err)
		}
		pngBytes, err := os.ReadFile(filepath.Join(dir, original.PNGFile))
		if err != nil || fmt.Sprintf("%x", sha256.Sum256(pngBytes)) != original.PNGSHA256 {
			t.Fatal("原版 PNG hash 不符")
		}
		img, err := png.Decode(bytes.NewReader(pngBytes))
		if err != nil {
			t.Fatal(err)
		}
		if img.Bounds().Dx() != ScreenW || img.Bounds().Dy() != ScreenH {
			t.Fatal("動畫 PNG 尺寸不符")
		}
		g.renderFrame()
		for p := range pixels {
			r, gc, b, a := img.At(p%ScreenW, p/ScreenW).RGBA()
			o := p * 4
			if g.rgba[o] != uint8(r>>8) || g.rgba[o+1] != uint8(gc>>8) || g.rgba[o+2] != uint8(b>>8) || a != 65535 {
				t.Fatalf("動畫 y=%d 像素 %d RGB 不符", original.Y, p)
			}
		}
		total += len(pixels)
		pose++
		lastStep = original.Step
		lastTick = original.Ticks
	}
	if pose != logo.Overlay.Positions() {
		t.Fatalf("原版動畫收據只有 %d 個位置", pose)
	}
	t.Logf("%d 個原版自然翻頁位置：色號與 production renderer RGB 全部通過", pose)
	for g.openingActive {
		if err := g.step(InputState{DirHeld: -1, DirEdge: -1}); err != nil {
			t.Fatal(err)
		}
	}
	if g.openingActive || !g.showTitle {
		t.Fatal("六張無輸入播放後未交回標題")
	}
	if err := g.step(InputState{Confirm: true, DirHeld: -1, DirEdge: -1}); err != nil {
		t.Fatal(err)
	}
	if g.newGame.stage != ngMenu {
		t.Fatal("播放完成後的正式 Confirm 未進入主選單")
	}
	result := map[string]any{
		"passed": true, "scope": "前五幕色號與正規化 renderer；第六幕全部129位置的色號及RGB；不含淡入淡出RGB或實機wall-clock",
		"oracle_receipt_sha256": fmt.Sprintf("%x", sha256.Sum256(raw)),
		"pack_content_hash":     g.pack.ContentHash(), "compared_pixels": total,
	}
	encoded, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "comparison.json"), append(encoded, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
