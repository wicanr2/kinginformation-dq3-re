package config

import (
	"strings"
	"testing"
)

// 預設值對拍 dq3_config_default。
func TestDefault(t *testing.T) {
	c := Default()
	if c.RngMode != RngDOS || !c.MusicEnabled || c.MusicVolume != 70 ||
		c.AudioBackend != AudioMT32 || c.CombatInfo || c.CombatHurtFx != 500 {
		t.Errorf("預設值不符:%+v", c)
	}
}

// Load:key=value 覆蓋 + 夾範圍 + 註解/空行略過 + combat_hurt_fx 舊布林相容。
func TestLoad(t *testing.T) {
	in := `# 註解
rng = real
music = 0
music_vol = 250
audio = sb
combat_info = y
combat_hurt_fx = 1
`
	c, got := Load(strings.NewReader(in))
	if got < 5 {
		t.Errorf("覆蓋鍵數 %d, 預期 ≥5", got)
	}
	if c.RngMode != RngReal {
		t.Error("rng=real 未生效")
	}
	if c.MusicEnabled {
		t.Error("music=0 未關")
	}
	if c.MusicVolume != 100 { // 250 → 夾到 100
		t.Errorf("music_vol 夾範圍錯:%d, want 100", c.MusicVolume)
	}
	if c.AudioBackend != AudioSB {
		t.Error("audio=sb 未生效")
	}
	if !c.CombatInfo {
		t.Error("combat_info=y 未開")
	}
	if c.CombatHurtFx != 500 { // 舊布林 1 → 500ms
		t.Errorf("combat_hurt_fx=1 應轉 500,實 %d", c.CombatHurtFx)
	}
}

// Save→Load round-trip:序列化再讀回應相等。
func TestSaveLoadRoundTrip(t *testing.T) {
	c := Config{RngMode: RngReal, MusicEnabled: false, MusicVolume: 42,
		AudioBackend: AudioSB, CombatInfo: false, CombatHurtFx: 1200}
	var b strings.Builder
	if err := Save(c, &b); err != nil {
		t.Fatal(err)
	}
	got, _ := Load(strings.NewReader(b.String()))
	if got != c {
		t.Errorf("round-trip 不符:\n寫 %+v\n讀 %+v", c, got)
	}
}
