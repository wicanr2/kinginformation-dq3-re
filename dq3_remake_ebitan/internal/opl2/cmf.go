package opl2

import "math"

// CMF 變體軌 parser + 排程器(移植 dq3_cmf.c)。MBG.MCX 18 軌 OPL2 FM 音樂。

type cmfEvent struct{ st, d1, d2, dl byte }

// 8 個 2-op FM 音色 patch(mod/car 各 0x20/40/60/80/E0 + C0)。自 dq3_cmf g_patch。
var patches = [8][11]byte{
	{0x21, 0x1a, 0xd5, 0x53, 0x00, 0x21, 0x07, 0xd5, 0x53, 0x00, 0x0a}, // brass lead
	{0x21, 0x10, 0xf2, 0x74, 0x00, 0x21, 0x00, 0xf2, 0x76, 0x00, 0x08}, // bass
	{0x31, 0x00, 0xf0, 0x90, 0x00, 0x31, 0x00, 0xf0, 0x90, 0x00, 0x0f}, // organ(add)
	{0x61, 0x1c, 0x52, 0x24, 0x01, 0x61, 0x08, 0x42, 0x16, 0x01, 0x0c}, // string pad
	{0x21, 0x2a, 0xf0, 0x53, 0x00, 0x21, 0x12, 0xf0, 0x53, 0x00, 0x0e}, // flute/soft
	{0x01, 0x16, 0xf8, 0x77, 0x00, 0x11, 0x06, 0xf8, 0x88, 0x00, 0x0a}, // pluck/harp
	{0x31, 0x1f, 0xd2, 0x53, 0x02, 0x21, 0x09, 0xd2, 0x55, 0x02, 0x0a}, // reed/clar
	{0x01, 0x1a, 0xf5, 0x42, 0x01, 0x12, 0x05, 0xf5, 0x52, 0x01, 0x0a}, // bell/chime
}

func (o *OPL2) applyPatch(c, idx int) {
	p := &patches[idx%8]
	m, car := modOff[c], modOff[c]+3
	o.Write(0x20+m, int(p[0]))
	o.Write(0x40+m, int(p[1]))
	o.Write(0x60+m, int(p[2]))
	o.Write(0x80+m, int(p[3]))
	o.Write(0xe0+m, int(p[4]))
	o.Write(0x20+car, int(p[5]))
	o.Write(0x40+car, int(p[6]))
	o.Write(0x60+car, int(p[7]))
	o.Write(0x80+car, int(p[8]))
	o.Write(0xe0+car, int(p[9]))
	o.Write(0xc0+c, int(p[10]))
}

// CMF 是一首 MBG 軌的播放狀態。
type CMF struct {
	opl         *OPL2
	sr          int
	sampPerTick float64
	ev          []cmfEvent
	cursor      int
	sampToNext  float64
	finished    bool
	curFnum     [9]int
	curBlock    [9]int
}

func le32b(d []byte, o int) int {
	if o+3 >= len(d) {
		return 0
	}
	return int(d[o]) | int(d[o+1])<<8 | int(d[o+2])<<16 | int(d[o+3])<<24
}

// MBGTracks 解 MBG.MCX 的軌偏移表(遞增 dword,同 VOC/audio)。回各軌 data slice。
func MBGTracks(mbg []byte) [][]byte {
	var offs []int
	for i := 0; i+4 <= len(mbg); i += 4 {
		v := le32b(mbg, i)
		if len(offs) > 0 && v <= offs[len(offs)-1] {
			break
		}
		if v >= len(mbg) || v < (len(offs)+1)*4 {
			break
		}
		offs = append(offs, v)
	}
	out := make([][]byte, len(offs))
	for k := range offs {
		end := len(mbg)
		if k+1 < len(offs) {
			end = offs[k+1]
		}
		out[k] = mbg[offs[k]:end]
	}
	return out
}

// parseEvents:MIDI-like 事件流(running status + 尾隨 delta)。移植 dq3_cmf parse_events。
func parseEvents(t []byte) []cmfEvent {
	if len(t) < 8 {
		return nil
	}
	base := int(t[2]) | int(t[3])<<8
	if base <= 0 || base >= len(t) {
		base = int(t[6]) | int(t[7])<<8
	}
	if base <= 0 || base >= len(t) {
		base = 0x60
	}
	p := base
	for p < len(t) && t[p] < 0x80 {
		p++
	}
	var ev []cmfEvent
	status := -1
	for p < len(t) {
		b := int(t[p])
		if b >= 0x80 {
			status = b
			p++
			if status == 0xff || status == 0xf0 {
				break
			}
		}
		if status < 0 {
			break
		}
		hi := status & 0xf0
		switch hi {
		case 0x80, 0x90, 0xb0, 0xa0, 0xe0:
			if p+3 > len(t) {
				return ev
			}
			ev = append(ev, cmfEvent{byte(status), t[p], t[p+1], t[p+2]})
			p += 3
		case 0xc0, 0xd0:
			if p+2 > len(t) {
				return ev
			}
			ev = append(ev, cmfEvent{byte(status), t[p], 0, t[p+1]})
			p += 2
		default:
			return ev
		}
	}
	return ev
}

// LoadCMF 從一段 MBG 軌建 CMF(tick tempo 自 word@6)。移植 dq3_cmf_load。
func LoadCMF(track []byte, opl *OPL2, sampleRate int, tickMs float64) *CMF {
	tms := tickMs
	if tms <= 0 {
		tms = 10.4
	}
	if len(track) >= 8 {
		if td := int(track[6]) | int(track[7])<<8; td > 0 {
			reload := 1193180.0 / float64(td)
			t := 1000.0 * reload / 1193182.0
			if t >= 2.0 && t <= 50.0 {
				tms = t
			}
		}
	}
	ev := parseEvents(track)
	if len(ev) == 0 {
		return nil
	}
	c := &CMF{opl: opl, sr: sampleRate, sampPerTick: tms * 0.001 * float64(sampleRate), ev: ev}
	opl.Write(0x01, 0x20) // waveform select enable
	for ch := 0; ch < 9; ch++ {
		opl.applyPatch(ch, 0)
	}
	return c
}

func noteToFB(note int) (fnum, block int) {
	freq := 440.0 * math.Pow(2, float64(note-69)/12.0)
	base := freq * 1048576.0 / 49716.0
	b := 0
	for base >= 1024.0 && b < 7 {
		base /= 2.0
		b++
	}
	fnum = int(base + 0.5)
	if fnum > 1023 {
		fnum = 1023
	}
	return fnum, b
}

func (c *CMF) applyEvent(e cmfEvent) {
	hi := int(e.st) & 0xf0
	ch := int(e.st) & 0x0f
	if ch >= 9 {
		return
	}
	switch {
	case hi == 0x90 && e.d2 > 0: // note-on
		fnum, block := noteToFB(int(e.d1))
		c.curFnum[ch], c.curBlock[ch] = fnum, block
		c.opl.Write(0xa0+ch, fnum&0xff)
		c.opl.Write(0xb0+ch, ((block&7)<<2)|((fnum>>8)&3)|0x20)
	case hi == 0x80 || (hi == 0x90 && e.d2 == 0): // note-off
		fnum, block := c.curFnum[ch], c.curBlock[ch]
		c.opl.Write(0xb0+ch, ((block&7)<<2)|((fnum>>8)&3))
	case hi == 0xc0: // program change
		c.applyPatch(ch, int(e.d1))
	}
}

func (c *CMF) applyPatch(ch, idx int) { c.opl.applyPatch(ch, idx) }

// Render 合成 n 個 mono s16 到 out;loop=true 循環。回是否未結束。移植 dq3_cmf_render。
func (c *CMF) Render(out []int16, n int, loop bool) bool {
	done := 0
	for done < n {
		if c.sampToNext <= 0 {
			if c.cursor >= len(c.ev) {
				if loop {
					c.cursor = 0
				} else {
					c.finished = true
					break
				}
			}
			if c.cursor < len(c.ev) {
				e := c.ev[c.cursor]
				c.cursor++
				c.applyEvent(e)
				c.sampToNext += float64(e.dl) * c.sampPerTick
			}
			continue
		}
		chunk := int(c.sampToNext)
		if chunk > n-done {
			chunk = n - done
		}
		if chunk < 1 {
			chunk = 1
		}
		c.opl.Render(out[done:], chunk)
		done += chunk
		c.sampToNext -= float64(chunk)
	}
	for i := done; i < n; i++ { // 結束補靜音
		out[i] = 0
	}
	return !c.finished
}

// EventCount 回事件數。
func (c *CMF) EventCount() int { return len(c.ev) }

// RenderAll 渲染整軌一遍(loop=false)→ mono s16(上限 maxSec 秒防爆)。供 gaudio 一次算好再 InfiniteLoop。
func (c *CMF) RenderAll(maxSec int) []int16 {
	chunk := make([]int16, 4410)
	var out []int16
	for len(out) < maxSec*c.sr {
		alive := c.Render(chunk, len(chunk), false)
		out = append(out, chunk...)
		if !alive {
			break
		}
	}
	return out
}

// RenderMBGTrack 便捷:MBG 全檔 + 軌號 → 該軌 mono s16(sampleRate,上限 maxSec)。
func RenderMBGTrack(mbg []byte, track, sampleRate, maxSec int) []int16 {
	tracks := MBGTracks(mbg)
	if track < 0 || track >= len(tracks) {
		return nil
	}
	cmf := LoadCMF(tracks[track], New(sampleRate), sampleRate, 0)
	if cmf == nil {
		return nil
	}
	return cmf.RenderAll(maxSec)
}
