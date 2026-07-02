package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// 設定選單(modal)。忠實移植 dq3_remake/src/main.c config_modal 的 6 列 + ←/→/Enter 切值 +
// ESC 關閉行為;版面座標(bx/by/rowy)沿用 C 版同一份 640×350 座標系。
//
// Go port 對「切了但沒真的接」的兩列誠實標記為 no-op / 需重啟(不假裝有效,對齊任務硬規則):
//   - RNG(row0):Go port 戰鬥恆用 DOS 忠實亂數,這列顯示但切換不改變行為。
//   - 音源(row3):MT-32 OGG ↔ SB-FM OPL2 是啟動時決定的播放路徑(NewGame 依 DQ3_FM 環境變數
//     選其一),執行期熱切換沒有對應機制;這列只顯示目前值,切換不即時套用。
const (
	setRowRNG        = iota
	setRowMusic      // 音樂 開/關 → 即時套用 g.music.SetEnabled
	setRowVolume     // 音量 0..100 → 即時套用 g.music.SetVolume
	setRowAudio      // 音源(no-op,顯示用;需重啟)
	setRowCombatInfo // 戰鬥資訊(敵 HP 顯示)→ 存 cfg,下場戰鬥生效
	setRowHurtFx     // 受傷特效時長 ms → 存 cfg,下場戰鬥生效
	setRowCount
)

// Settings 是設定選單的 modal 狀態。
type Settings struct {
	tx     *dq3data.Text // 標籤字型(共用 g.dlg.tx;Glyph() 只依賴 FON,與哪個 txt bank 無關)
	open   bool
	cursor int
}

// Open 開啟設定選單,游標歸零。
func (s *Settings) Open() { s.open, s.cursor = true, 0 }

// settingsInput 處理一幀輸入:↑/↓ 移列、←/→/Enter 切值、ESC(Cancel)關閉。
// 值變更即時套用(音樂開關/音量 → g.music;其餘寫回 g.cfg,下場戰鬥開戰時讀取)。
func (g *Game) settingsInput(in InputState) {
	s := &g.settings
	switch {
	case in.Cancel:
		s.open = false
	case in.DirEdge == 1: // 上
		s.cursor = (s.cursor + setRowCount - 1) % setRowCount
	case in.DirEdge == 0: // 下
		s.cursor = (s.cursor + 1) % setRowCount
	case in.DirEdge == 2: // 左
		g.applySettingChange(s.cursor, -1)
	case in.DirEdge == 3 || in.Confirm: // 右 / Enter
		g.applySettingChange(s.cursor, 1)
	}
}

// applySettingChange:切第 row 列的值(dir=+1/-1),對齊 C config_modal 的每列切值邏輯。
func (g *Game) applySettingChange(row, dir int) {
	switch row {
	case setRowRNG:
		// no-op:Go port 戰鬥恆用 DOS 忠實亂數(見 internal/rng),沒有 real 模式可切。
	case setRowMusic:
		g.cfg.MusicEnabled = !g.cfg.MusicEnabled
		g.music.SetEnabled(g.cfg.MusicEnabled)
	case setRowVolume:
		v := g.cfg.MusicVolume + dir*10
		if v > 100 {
			v = 0
		} else if v < 0 {
			v = 100
		}
		g.cfg.MusicVolume = v
		g.music.SetVolume(v)
	case setRowAudio:
		// no-op(顯示用):MT-32/SB-FM 為啟動時決定的播放路徑,執行期不熱切換;誠實不假裝生效。
	case setRowCombatInfo:
		g.cfg.CombatInfo = !g.cfg.CombatInfo
	case setRowHurtFx:
		v := g.cfg.CombatHurtFx
		if dir > 0 {
			if v <= 0 {
				v = 100
			} else {
				v += 200
			}
			if v > 1900 {
				v = 0
			}
		} else {
			if v <= 0 {
				v = 1900
			} else {
				v -= 200
			}
			if v < 100 {
				v = 0
			}
		}
		g.cfg.CombatHurtFx = v
	}
}

// 標籤 / 值 glyph 常數,原封移植自 C config_modal 的靜態陣列(字碼取自共用 D3TXT00.FON,
// 與 Go/C 兩版共用同一份素材,數值相同)。設/版兩字用自建字形(customglyph.go)。
var (
	setLRng  = [4]int{482, 427, 1397, 1157} // 亂數模式
	setVReal = [2]int{577, 279}             // 真實
	setLMus  = [2]int{313, 682}             // 音樂
	setLVol  = [2]int{313, 280}             // 音量
	setLAud  = [2]int{313, 1428}            // 音源
	setVFM   = [2]int{20, 27}               // FM(SB FM)
	setVMT   = [4]int{27, 34, 3, 2}         // MT32
	setLInfo = [2]int{478, 670}             // 敵情
	setLFx   = [2]int{658, 619}             // 震動(受傷特效)
	setHint  = [2]int{750, 312}             // 返回
)

const (
	setVOn  = 488 // 開
	setVOff = 683 // 關
	setVDos = 555 // 原版(第一字;第二字用自建字形 CGBan=版)
)

// draw 畫設定選單(黑底白框 + 標題「設定」+ 6 列標籤/值 + 目前列游標)。
func (s *Settings) draw(rgba []byte, cfg config.Config) {
	if !s.open {
		return
	}
	white := dq3data.Color{R: 248, G: 248, B: 248}
	yellow := dq3data.Color{R: 248, G: 216, B: 40}
	const gpx = dq3data.GlyphPx
	bx, by, bw, bh := 150, 48, 340, 276
	fillBox(rgba, bx, by, bw, bh, white)

	// 標題「設定」(設=自建字 0、定=官方字 1022)
	drawCustomGlyph(rgba, dq3data.CGShe, bx+150, by+14, yellow)
	drawGlyph(rgba, s.tx, bx+150+gpx, by+14, 1022, yellow)

	rowy := [setRowCount]int{by + 50, by + 82, by + 114, by + 146, by + 178, by + 210}
	labelX, valueX := bx+40, bx+40+6*gpx
	for r := 0; r < setRowCount; r++ {
		if r == s.cursor { // ► 游標
			drawGlyph(rgba, s.tx, bx+18, rowy[r], 11, yellow)
		}
	}
	// row0 亂數模式:標籤 4 字 + 值(DOS=原+自建「版」/ REAL=真實)
	for i, gl := range setLRng {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowRNG], gl, white)
	}
	if cfg.RngMode == config.RngReal {
		for i, gl := range setVReal {
			drawGlyph(rgba, s.tx, valueX+i*gpx, rowy[setRowRNG], gl, yellow)
		}
	} else {
		drawGlyph(rgba, s.tx, valueX, rowy[setRowRNG], setVDos, yellow)
		drawCustomGlyph(rgba, dq3data.CGBan, valueX+gpx, rowy[setRowRNG], yellow)
	}
	// row1 音樂 開/關
	for i, gl := range setLMus {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowMusic], gl, white)
	}
	drawOnOff(rgba, s.tx, valueX, rowy[setRowMusic], cfg.MusicEnabled, yellow)
	// row2 音量 數字
	for i, gl := range setLVol {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowVolume], gl, white)
	}
	drawNumber(rgba, s.tx, valueX, rowy[setRowVolume], cfg.MusicVolume, yellow)
	// row3 音源 SB-FM / MT-32
	for i, gl := range setLAud {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowAudio], gl, white)
	}
	if cfg.AudioBackend == config.AudioMT32 {
		for i, gl := range setVMT {
			drawGlyph(rgba, s.tx, valueX+i*gpx, rowy[setRowAudio], gl, yellow)
		}
	} else {
		for i, gl := range setVFM {
			drawGlyph(rgba, s.tx, valueX+i*gpx, rowy[setRowAudio], gl, yellow)
		}
	}
	// row4 敵情(戰鬥資訊)開/關
	for i, gl := range setLInfo {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowCombatInfo], gl, white)
	}
	drawOnOff(rgba, s.tx, valueX, rowy[setRowCombatInfo], cfg.CombatInfo, yellow)
	// row5 受傷特效:時長 ms 或 關
	for i, gl := range setLFx {
		drawGlyph(rgba, s.tx, labelX+i*gpx, rowy[setRowHurtFx], gl, white)
	}
	if cfg.CombatHurtFx > 0 {
		drawNumber(rgba, s.tx, valueX, rowy[setRowHurtFx], cfg.CombatHurtFx, yellow)
	} else {
		drawGlyph(rgba, s.tx, valueX, rowy[setRowHurtFx], setVOff, yellow)
	}
	// 底部提示「返回」
	for i, gl := range setHint {
		drawGlyph(rgba, s.tx, bx+140+i*gpx, by+248, gl, white)
	}
}

// drawOnOff:畫單一 glyph 的 開/關 值。
func drawOnOff(rgba []byte, tx *dq3data.Text, x, y int, on bool, fg dq3data.Color) {
	v := setVOff
	if on {
		v = setVOn
	}
	drawGlyph(rgba, tx, x, y, v, fg)
}

// drawCustomGlyph:畫自建字形(dq3data.CustomGlyph,設/版),沿用 drawGlyph 同一套點陣繪製邏輯。
func drawCustomGlyph(rgba []byte, idx, px, py int, fg dq3data.Color) {
	g, ok := dq3data.CustomGlyph(idx)
	if !ok {
		return
	}
	for r := 0; r < 16; r++ {
		for c := 0; c < 16; c++ {
			if g[r][c] != 0 {
				putPx(rgba, px+c, py+r, fg)
			}
		}
	}
}
