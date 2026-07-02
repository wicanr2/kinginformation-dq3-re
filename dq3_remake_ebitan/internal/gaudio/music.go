// Package gaudio 是 MT-32 預錄 OGG 音樂播放(Ebiten audio/vorbis,純 Go 無原生庫)。
// 對應 C dq3_audio 的 MT-32 後端:場景 → 軌號 → track_NN.ogg 無限循環。
// 所有錯誤(無檔 / 解碼失敗 / 無音訊裝置)一律非致命 → 靜音降級,遊戲照跑。
package gaudio

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/wicanr2/dq3_remake_ebitan/internal/opl2"
)

const sampleRate = 44100

// Music 管理當前播放的單軌(場景切換時換軌)。
type Music struct {
	ctx     *audio.Context
	player  *audio.Player
	cur     int
	fsys    fs.FS // 放 track_NN.ogg 的檔案系統(桌面=os.DirFS、行動=embed.FS)
	enabled bool
	vol     float64
	sfx     [][]byte       // 預轉的音效(16-bit LE stereo)
	mbg     []byte         // SB-FM 音樂源(MBG.MCX;非 nil → 用 OPL2 FM 取代 MT-32 OGG)
	fmCache map[int][]byte // 已渲染的 FM 軌(stereo bytes)
}

// SetMBG 啟用 SB-FM 音樂(OPL2 合成 MBG.MCX,取代 MT-32 OGG)。
func (m *Music) SetMBG(mbg []byte) {
	if !m.enabled || len(mbg) == 0 {
		return
	}
	m.mbg = mbg
	m.fmCache = map[int][]byte{}
}

func monoToStereo(pcm []int16) []byte {
	b := make([]byte, len(pcm)*4)
	for j, s := range pcm {
		lo, hi := byte(uint16(s)), byte(uint16(s)>>8)
		o := j * 4
		b[o], b[o+1], b[o+2], b[o+3] = lo, hi, lo, hi
	}
	return b
}

// NewMusic:fsys = 放 track_NN.ogg 的檔案系統(nil → 停用音樂)。桌面傳 os.DirFS(dir),行動傳 embed 子樹。
func NewMusic(fsys fs.FS) (m *Music) {
	m = &Music{fsys: fsys, cur: -1, vol: 0.5}
	if fsys == nil {
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
	if m == nil || !m.enabled || track < 0 {
		return
	}
	if m.cur == track && m.player != nil && m.player.IsPlaying() {
		return
	}
	var loop *audio.InfiniteLoop
	if m.mbg != nil { // SB-FM:OPL2 合成 MBG 軌(快取)
		b, ok := m.fmCache[track]
		if !ok {
			b = monoToStereo(opl2.RenderMBGTrack(m.mbg, track, sampleRate, 90))
			m.fmCache[track] = b
		}
		if len(b) == 0 {
			return
		}
		loop = audio.NewInfiniteLoop(bytes.NewReader(b), int64(len(b)))
	} else { // MT-32 預錄 OGG
		data, err := fs.ReadFile(m.fsys, fmt.Sprintf("track_%02d.ogg", track))
		if err != nil {
			return
		}
		s, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return
		}
		loop = audio.NewInfiniteLoop(s, s.Length())
	}
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

// SetSFX 預轉 VOC 音效(mono s16 @ sampleRate)為 Ebiten 位元組(16-bit LE stereo)。
func (m *Music) SetSFX(banks [][]int16) {
	if !m.enabled {
		return
	}
	m.sfx = make([][]byte, len(banks))
	for i, pcm := range banks {
		b := make([]byte, len(pcm)*4) // mono → stereo,每 sample 4 byte
		for j, s := range pcm {
			lo, hi := byte(uint16(s)), byte(uint16(s)>>8)
			o := j * 4
			b[o], b[o+1], b[o+2], b[o+3] = lo, hi, lo, hi // L, R
		}
		m.sfx[i] = b
	}
}

// PlaySFX 一次性播放第 i 個音效(fire-and-forget;非致命)。
func (m *Music) PlaySFX(i int) {
	if !m.enabled || i < 0 || i >= len(m.sfx) {
		return
	}
	defer func() { recover() }()
	p := m.ctx.NewPlayerFromBytes(m.sfx[i])
	p.Play() // Ebiten 保活至播完
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
