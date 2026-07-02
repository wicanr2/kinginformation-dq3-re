// Package config 是可攜設定(key=value,移植 dq3_config.c)。純邏輯:預設值 + parse/serialize。
// rng 恆用 DOS 忠實模式;其餘(音樂/音量/音源/戰鬥資訊/受傷特效)對應 Go 端執行期選項。
package config

import (
	"bufio"
	"io"
	"strconv"
	"strings"
)

// RngMode:亂數模式(DOS 忠實 / real)。Go port 預設 DOS。
type RngMode int

const (
	RngDOS  RngMode = iota // 忠實 DOS 16-bit 亂數
	RngReal                // 平台 real 亂數
)

// AudioBackend:音源後端。
type AudioBackend int

const (
	AudioSB   AudioBackend = 0 // SB FM(OPL2)
	AudioMT32 AudioBackend = 1 // MT-32(無音檔時退回 SB FM)
)

// Config 是一組可攜設定(對齊 dq3_config)。
type Config struct {
	RngMode      RngMode
	MusicEnabled bool
	MusicVolume  int // 0..100
	AudioBackend AudioBackend
	CombatInfo   bool // 戰鬥顯示敵人 HP + 預計動作
	CombatHurtFx int  // 受傷特效時長 ms(0=關;預設 500;上限 2000)
}

// Default 回預設設定(移植 dq3_config_default)。
func Default() Config {
	return Config{
		RngMode:      RngDOS,
		MusicEnabled: true,
		MusicVolume:  70,
		AudioBackend: AudioMT32,
		CombatInfo:   true,
		CombatHurtFx: 500,
	}
}

func trim(s string) string { return strings.Trim(s, " \t\r\n") }

// Load 從 key=value 文字流讀入設定(在 Default 基礎上覆蓋)。回覆蓋的鍵數。移植 dq3_config_load。
func Load(r io.Reader) (Config, int) {
	c := Default()
	got := 0
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		s := trim(sc.Text())
		if s == "" || s[0] == '#' {
			continue
		}
		eq := strings.IndexByte(s, '=')
		if eq < 0 {
			continue
		}
		key, val := trim(s[:eq]), trim(s[eq+1:])
		if val == "" {
			continue
		}
		switch key {
		case "rng":
			if val[0] == 'r' || val[0] == 'R' {
				c.RngMode = RngReal
			} else {
				c.RngMode = RngDOS
			}
			got++
		case "music":
			c.MusicEnabled = boolVal(val)
			got++
		case "music_vol":
			c.MusicVolume = clamp(atoi(val), 0, 100)
			got++
		case "audio":
			if val[0] == 'm' || val[0] == 'M' {
				c.AudioBackend = AudioMT32
			} else {
				c.AudioBackend = AudioSB
			}
			got++
		case "combat_info":
			c.CombatInfo = boolVal(val)
		case "combat_hurt_fx":
			v := atoi(val)
			if v == 1 { // 舊布林 1 → 500ms(向後相容)
				v = 500
			}
			c.CombatHurtFx = clamp(v, 0, 2000)
			got++
		}
	}
	return c, got
}

// Save 以 key=value 序列化設定到 w(移植 dq3_config_save)。
func Save(c Config, w io.Writer) error {
	rng := "dos"
	if c.RngMode == RngReal {
		rng = "real"
	}
	audio := "sb"
	if c.AudioBackend == AudioMT32 {
		audio = "mt32"
	}
	_, err := io.WriteString(w,
		"# DQ3 remake 設定(可攜;取代 env)\n"+
			"rng="+rng+"\n"+
			"music="+b2s(c.MusicEnabled)+"\n"+
			"music_vol="+strconv.Itoa(c.MusicVolume)+"\n"+
			"audio="+audio+"\n"+
			"combat_info="+b2s(c.CombatInfo)+"\n"+
			"combat_hurt_fx="+strconv.Itoa(c.CombatHurtFx)+"\n")
	return err
}

func boolVal(val string) bool {
	switch val[0] {
	case '1', 'o', 'O', 'y', 'Y':
		return true
	}
	return false
}

func b2s(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
