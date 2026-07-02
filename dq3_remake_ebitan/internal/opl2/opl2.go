// Package opl2 是精簡 OPL2(YM3812)FM 合成核心(移植 dq3_opl2.c)。
// 相位累積 + 正弦表(4 波形)+ 線性 ADSR(指數曲線)+ 2-op FM/additive。非 cycle-accurate。
package opl2

import "math"

const (
	nch       = 9
	sinLen    = 2048
	oplNative = 49716.0 // OPL2 取樣率 3.579545MHz/72
)

const (
	envOff = iota
	envAtk
	envDec
	envSus
	envRel
)

// MULT 暫存器 0..15 → 倍頻 ×2。
var multX2 = [16]int{1, 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 20, 24, 24, 30, 30}

// 9 channel modulator operator 暫存器 offset(carrier = +3)。
var modOff = [nch]int{0x00, 0x01, 0x02, 0x08, 0x09, 0x0A, 0x10, 0x11, 0x12}

type op struct {
	mult          int
	tlAmp         float64
	ar, dr, sl, rr int
	wave          int
	ksr, egt      int
	phase, phinc  float64
	state         int
	env, last     float64
}

type ch struct {
	fnum, block, keyon, fb, con int
	op                          [2]op
}

// OPL2 是合成器狀態。
type OPL2 struct {
	sr     int
	reg    [256]byte
	chn    [nch]ch
	sintab [4][sinLen]float64
	wse    int
}

// New 建 OPL2(sampleRate;<=0 → 44100)。
func New(sampleRate int) *OPL2 {
	o := &OPL2{sr: 44100}
	if sampleRate > 0 {
		o.sr = sampleRate
	}
	for i := 0; i < sinLen; i++ {
		t := float64(i) / sinLen
		s := math.Sin(2 * math.Pi * t)
		o.sintab[0][i] = s // 全正弦
		if s > 0 {         // 半波整流
			o.sintab[1][i] = s
		}
		o.sintab[2][i] = math.Abs(s) // 全波整流
		if t < 0.25 || (t >= 0.5 && t < 0.75) {
			o.sintab[3][i] = math.Abs(s) // 鋸狀脈衝
		}
	}
	o.Reset()
	return o
}

// Reset 清狀態。
func (o *OPL2) Reset() {
	for i := range o.reg {
		o.reg[i] = 0
	}
	o.wse = 0
	for c := 0; c < nch; c++ {
		o.chn[c] = ch{}
		for k := 0; k < 2; k++ {
			o.chn[c].op[k].tlAmp = 1.0
			o.chn[c].op[k].mult = 2
		}
	}
}

func off2chop(off int) (c, opi int, ok bool) {
	for c := 0; c < nch; c++ {
		if off == modOff[c] {
			return c, 0, true
		}
		if off == modOff[c]+3 {
			return c, 1, true
		}
	}
	return 0, 0, false
}

func tlToAmp(tl int) float64 { return math.Pow(10, -(float64(tl)*0.75)/20) }
func slToAmp(sl int) float64 {
	if sl >= 15 {
		return 0
	}
	return math.Pow(10, -(float64(sl)*3)/20)
}

func (o *OPL2) updatePhinc(c, k int) {
	cc := &o.chn[c]
	op := &cc.op[k]
	base := float64(cc.fnum) * float64(int(1)<<uint(cc.block)) * (oplNative / 1048576.0)
	f := base * (float64(multX2[op.mult&15]) / 2.0)
	op.phinc = f / float64(o.sr)
}

// Write 寫暫存器(YM3812 規格)。移植 dq3_opl2_write。
func (o *OPL2) Write(reg, val int) {
	reg &= 0xff
	val &= 0xff
	o.reg[reg] = byte(val)
	if reg == 0x01 {
		o.wse = (val >> 5) & 1
		return
	}
	hi, lo := reg&0xf0, reg&0x0f
	switch hi {
	case 0x20, 0x40, 0x60, 0x80, 0xE0:
		c, k, ok := off2chop(lo)
		if !ok {
			return
		}
		op := &o.chn[c].op[k]
		switch hi {
		case 0x20:
			op.mult = val & 0x0f
			op.egt = (val >> 5) & 1
			op.ksr = (val >> 4) & 1
			o.updatePhinc(c, k)
		case 0x40:
			op.tlAmp = tlToAmp(val & 0x3f)
		case 0x60:
			op.ar = (val >> 4) & 0x0f
			op.dr = val & 0x0f
		case 0x80:
			op.sl = (val >> 4) & 0x0f
			op.rr = val & 0x0f
		case 0xE0:
			if o.wse != 0 {
				op.wave = val & 0x03
			} else {
				op.wave = 0
			}
		}
		return
	}
	switch {
	case hi == 0xA0 && lo < nch:
		o.chn[lo].fnum = (o.chn[lo].fnum & 0x300) | val
		o.updatePhinc(lo, 0)
		o.updatePhinc(lo, 1)
	case hi == 0xB0 && lo < nch:
		c := lo
		o.chn[c].fnum = (o.chn[c].fnum & 0xff) | ((val & 0x03) << 8)
		o.chn[c].block = (val >> 2) & 0x07
		newkey := (val >> 5) & 1
		o.updatePhinc(c, 0)
		o.updatePhinc(c, 1)
		if newkey != 0 && o.chn[c].keyon == 0 { // key-on → 兩 op attack
			for k := 0; k < 2; k++ {
				o.chn[c].op[k].state = envAtk
				o.chn[c].op[k].phase = 0
				o.chn[c].op[k].env = 0
			}
		} else if newkey == 0 && o.chn[c].keyon != 0 { // key-off → release
			for k := 0; k < 2; k++ {
				o.chn[c].op[k].state = envRel
			}
		}
		o.chn[c].keyon = newkey
	case hi == 0xC0 && lo < nch:
		o.chn[lo].fb = (val >> 1) & 0x07
		o.chn[lo].con = val & 1
	}
}

func (o *OPL2) atkStep(r int) float64 {
	if r == 0 {
		return 0
	}
	ms := 8.0 * math.Pow(0.5, float64(r))
	if ms < 3.0 {
		ms = 3.0
	}
	return 1.0 / (ms*0.001*float64(o.sr) + 1.0)
}
func (o *OPL2) decStep(r int) float64 {
	if r == 0 {
		return 0
	}
	ms := 80.0 * math.Pow(0.5, float64(r))
	if ms < 6.0 {
		ms = 6.0
	}
	return 1.0 / (ms*0.001*float64(o.sr) + 1.0)
}

func (o *OPL2) opEval(op *op, pmod float64) float64 {
	switch op.state { // envelope
	case envAtk:
		op.env += (1.0 - op.env) * o.atkStep(op.ar)
		if op.env >= 0.999 || op.ar == 0 {
			op.env = 1.0
			op.state = envDec
		}
	case envDec:
		tgt := slToAmp(op.sl)
		op.env += (tgt - op.env) * o.decStep(op.dr) * 2.0
		if op.env <= tgt+0.001 {
			op.env = tgt
			op.state = envSus
		}
	case envSus:
	case envRel:
		op.env -= op.env * o.decStep(op.rr)
		if op.env < 0.0003 {
			op.env = 0
			op.state = envOff
		}
	default:
		return 0
	}
	ph := op.phase + pmod
	ph -= math.Floor(ph)
	idx := int(ph * sinLen)
	if idx >= sinLen {
		idx = sinLen - 1
	}
	wt := &o.sintab[op.wave&3]
	frac := ph*sinLen - float64(idx)
	s := wt[idx] + (wt[(idx+1)&(sinLen-1)]-wt[idx])*frac
	out := s * op.env * op.tlAmp
	op.phase += op.phinc
	if op.phase >= 1.0 {
		op.phase -= math.Floor(op.phase)
	}
	op.last = out
	return out
}

// Render 合成 nsamples 個 mono s16 到 out。移植 dq3_opl2_render。
func (o *OPL2) Render(out []int16, nsamples int) {
	for n := 0; n < nsamples && n < len(out); n++ {
		mix := 0.0
		for c := 0; c < nch; c++ {
			cc := &o.chn[c]
			if cc.op[0].state == envOff && cc.op[1].state == envOff {
				continue
			}
			var fbmod float64
			if cc.fb != 0 {
				fbmod = cc.op[0].last * (float64(cc.fb) / 8.0) * 0.5
			}
			modout := o.opEval(&cc.op[0], fbmod)
			if cc.con != 0 { // additive
				mix += modout + o.opEval(&cc.op[1], 0)
			} else { // FM
				mix += o.opEval(&cc.op[1], modout*0.5)
			}
		}
		mix *= 0.16
		mix = math.Tanh(mix)
		out[n] = int16(mix * 28000.0)
	}
}
