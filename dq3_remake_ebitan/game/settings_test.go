package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/config"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gaudio"
)

// 建一個純狀態 Game(無素材/無 ebiten image),只測 settings modal 的 input/狀態轉移。
func newSettingsTestGame() *Game {
	return &Game{cfg: config.Default(), music: gaudio.NewMusic(nil)}
}

// 開選單 → 游標在音樂列切 ON→OFF → g.cfg.MusicEnabled 變 false(且 SetEnabled nil-safe 不 panic)。
func TestSettingsToggleMusic(t *testing.T) {
	g := newSettingsTestGame()
	if !g.cfg.MusicEnabled {
		t.Fatal("預設應音樂 ON")
	}
	g.settings.Open()
	if !g.settings.open {
		t.Fatal("Open 後應為開啟狀態")
	}
	g.settings.cursor = setRowMusic
	g.settingsInput(InputState{DirEdge: -1, Confirm: true}) // Enter 切值
	if g.cfg.MusicEnabled {
		t.Error("切一次後 MusicEnabled 應變 false")
	}
	g.settingsInput(InputState{DirEdge: -1, Confirm: true})
	if !g.cfg.MusicEnabled {
		t.Error("再切一次應變回 true")
	}
}

// 音量列:→ +10、← -10,並驗夾範圍循環(對齊 C:v>100→0,v<0→100)。
func TestSettingsVolumeWrap(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open()
	g.settings.cursor = setRowVolume
	g.cfg.MusicVolume = 100
	g.settingsInput(InputState{DirEdge: 3}) // 右:+10 → 超界歸零
	if g.cfg.MusicVolume != 0 {
		t.Errorf("音量 100+10 應繞回 0,實 %d", g.cfg.MusicVolume)
	}
	g.settingsInput(InputState{DirEdge: 2}) // 左:-10 → 低界繞回 100
	if g.cfg.MusicVolume != 100 {
		t.Errorf("音量 0-10 應繞回 100,實 %d", g.cfg.MusicVolume)
	}
}

// 戰鬥資訊列(敵情)ON→OFF 切換 g.cfg.CombatInfo。
func TestSettingsToggleCombatInfo(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open()
	g.settings.cursor = setRowCombatInfo
	before := g.cfg.CombatInfo
	g.settingsInput(InputState{DirEdge: -1, Confirm: true})
	if g.cfg.CombatInfo == before {
		t.Error("CombatInfo 應切換")
	}
}

// 受傷特效列:200ms 間隔循環,0(關)↔100..1900。
func TestSettingsHurtFxCycle(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open()
	g.settings.cursor = setRowHurtFx
	g.cfg.CombatHurtFx = 0
	g.settingsInput(InputState{DirEdge: 3}) // 右:關→100
	if g.cfg.CombatHurtFx != 100 {
		t.Errorf("受傷特效 關+1 應為 100,實 %d", g.cfg.CombatHurtFx)
	}
	g.cfg.CombatHurtFx = 1900
	g.settingsInput(InputState{DirEdge: 3}) // 右:1900+200 越界 → 0
	if g.cfg.CombatHurtFx != 0 {
		t.Errorf("受傷特效 1900+1 應繞回 0(關),實 %d", g.cfg.CombatHurtFx)
	}
}

// RNG(row0)/音源(row3)為誠實 no-op:切值不應動到 cfg 對應欄位。
func TestSettingsRngAudioNoOp(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open()
	wantRng, wantAudio := g.cfg.RngMode, g.cfg.AudioBackend
	g.settings.cursor = setRowRNG
	g.settingsInput(InputState{DirEdge: -1, Confirm: true})
	g.settings.cursor = setRowAudio
	g.settingsInput(InputState{DirEdge: -1, Confirm: true})
	if g.cfg.RngMode != wantRng {
		t.Error("RNG 列應 no-op,不應改變 cfg.RngMode")
	}
	if g.cfg.AudioBackend != wantAudio {
		t.Error("音源列應 no-op(需重啟),不應改變 cfg.AudioBackend")
	}
}

// ESC(Cancel)關閉選單。
func TestSettingsCancelCloses(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open()
	if !g.settings.open {
		t.Fatal("應已開啟")
	}
	g.settingsInput(InputState{DirEdge: -1, Cancel: true})
	if g.settings.open {
		t.Error("Cancel 後應關閉")
	}
}

// 上/下方向鍵在 6 列間繞回移動游標。
func TestSettingsCursorWrap(t *testing.T) {
	g := newSettingsTestGame()
	g.settings.Open() // cursor=0
	g.settingsInput(InputState{DirEdge: 1})
	if g.settings.cursor != setRowCount-1 {
		t.Errorf("cursor=0 按上應繞到最後一列 %d,實 %d", setRowCount-1, g.settings.cursor)
	}
	g.settingsInput(InputState{DirEdge: 0})
	if g.settings.cursor != 0 {
		t.Errorf("再按下應繞回 0,實 %d", g.settings.cursor)
	}
}
