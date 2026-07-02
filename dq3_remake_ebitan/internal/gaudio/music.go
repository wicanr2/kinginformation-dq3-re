// Package gaudio 是 MT-32 預錄 OGG 音樂播放(Ebiten audio/vorbis,純 Go 無原生庫)。
// 對應 C dq3_audio 的 MT-32 後端:場景 → 軌號 → track_NN.ogg 無限循環。
// 所有錯誤(無檔 / 解碼失敗 / 無音訊裝置)一律非致命 → 靜音降級,遊戲照跑。
package gaudio

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

const sampleRate = 44100

// Music 管理當前播放的單軌(場景切換時換軌)。
type Music struct {
	ctx     *audio.Context
	player  *audio.Player
	cur     int
	dir     string
	enabled bool
	vol     float64
}

// NewMusic:dir = 放 track_NN.ogg 的目錄(空字串 → 停用音樂)。
func NewMusic(dir string) (m *Music) {
	m = &Music{dir: dir, cur: -1, vol: 0.5}
	if dir == "" {
		return m
	}
	defer func() {
		if recover() != nil { // audio 後端初始化失敗 → 靜音降級
			m.enabled = false
		}
	}()
	m.ctx = audio.NewContext(sampleRate)
	m.enabled = true
	return m
}

// Play 播放第 track 軌(track_NN.ogg,無限循環)。同軌播放中則不重啟。
func (m *Music) Play(track int) {
	if !m.enabled || track < 0 {
		return
	}
	if m.cur == track && m.player != nil && m.player.IsPlaying() {
		return
	}
	data, err := os.ReadFile(filepath.Join(m.dir, fmt.Sprintf("track_%02d.ogg", track)))
	if err != nil {
		return
	}
	s, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return
	}
	loop := audio.NewInfiniteLoop(s, s.Length())
	p, err := m.ctx.NewPlayer(loop)
	if err != nil {
		return
	}
	if m.player != nil {
		m.player.Pause()
		_ = m.player.Close()
	}
	p.SetVolume(m.vol)
	func() {
		defer func() {
			if recover() != nil {
				m.enabled = false
			}
		}()
		p.Play()
	}()
	m.player, m.cur = p, track
}

// Stop 暫停當前軌。
func (m *Music) Stop() {
	if m.player != nil {
		m.player.Pause()
	}
}

// Enabled 回音樂是否可用(有目錄 + 音訊後端 OK)。
func (m *Music) Enabled() bool { return m.enabled }

// DecodeOgg 只解碼一段 OGG(供 headless 測試驗有效性);回 (sample 長度, err)。不需 audio context。
func DecodeOgg(data []byte) (int64, error) {
	s, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return 0, err
	}
	return s.Length(), nil
}
