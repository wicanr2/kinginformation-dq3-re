package opl2

import "testing"

func TestOPL2Synth(t *testing.T) {
	o := New(44100)
	// channel0 一個 2-op FM 音色:mod off=0x00、car off=0x03
	o.Write(0x20, 0x01) // mod: mult1
	o.Write(0x23, 0x01) // car: mult1
	o.Write(0x40, 0x10) // mod TL(中量調變)
	o.Write(0x43, 0x00) // car TL(全音量)
	o.Write(0x60, 0xf0) // mod AR15/DR0
	o.Write(0x63, 0xf0) // car AR15/DR0
	o.Write(0x80, 0x0f) // mod SL0/RR15
	o.Write(0x83, 0x0f) // car SL0/RR15
	o.Write(0xC0, 0x06) // fb3/FM
	// 設頻率 + key-on:A4≈440Hz。fnum/block
	o.Write(0xA0, 0x81) // fnum low
	o.Write(0xB0, 0x2e) // block4 + key-on + fnum high
	buf := make([]int16, 4410)
	o.Render(buf, len(buf))
	nonZero, peak := 0, int16(0)
	for _, s := range buf {
		if s != 0 {
			nonZero++
		}
		if s > peak {
			peak = s
		}
	}
	if nonZero == 0 || peak < 100 {
		t.Fatalf("OPL2 合成應有波形,得 nonZero=%d peak=%d", nonZero, peak)
	}
	t.Logf("OPL2 合成 ✓:%d/%d 非零、峰值 %d", nonZero, len(buf), peak)
}
