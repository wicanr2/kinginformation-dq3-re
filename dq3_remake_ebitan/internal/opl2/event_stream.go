package opl2

import "fmt"

// StreamEvent stores the MIDI delta BEFORE its event. EOT is retained.
type StreamEvent struct {
	Delta                uint32
	Status, Data1, Data2 byte
}

// ParseEventStream accepts a bounded, reviewed MIDI event range. It never guesses
// a legacy header or accepts a truncated stream. Supported meta events: EOT only.
func ParseEventStream(data []byte) ([]StreamEvent, uint64, error) {
	var events []StreamEvent
	var ticks uint64
	var running byte
	pos := 0
	fail := func() ([]StreamEvent, uint64, error) {
		return nil, 0, fmt.Errorf("invalid bounded MIDI stream at byte %d", pos)
	}
	for pos < len(data) {
		var delta uint32
		complete := false
		for i := 0; i < 4; i++ {
			if pos >= len(data) {
				return fail()
			}
			b := data[pos]
			pos++
			delta = delta<<7 | uint32(b&0x7f)
			if b&0x80 == 0 {
				complete = true
				break
			}
		}
		if !complete || pos >= len(data) {
			return fail()
		}
		ticks += uint64(delta)
		st := data[pos]
		if st >= 0x80 {
			pos++
			running = st
		} else {
			st = running
		}
		if st == 0xff {
			if pos+2 != len(data) || data[pos] != 0x2f || data[pos+1] != 0 {
				return fail()
			}
			events = append(events, StreamEvent{Delta: delta, Status: st})
			return events, ticks, nil
		}
		if st < 0x80 || st >= 0xf0 {
			return fail()
		}
		n := 2
		if st&0xf0 == 0xc0 || st&0xf0 == 0xd0 {
			n = 1
		}
		if pos+n > len(data) {
			return fail()
		}
		e := StreamEvent{Delta: delta, Status: st, Data1: data[pos]}
		if n == 2 {
			e.Data2 = data[pos+1]
		}
		if e.Data1 >= 0x80 || e.Data2 >= 0x80 {
			return fail()
		}
		events = append(events, e)
		pos += n
	}
	return fail()
}

// RenderEventStream preserves event order and rational cumulative timing. The
// existing OPL2 patches remain a synthesis approximation, not original FM parity.
func RenderEventStream(data []byte, sampleRate int, divisor, reference int64) ([]int16, error) {
	events, ticks, err := ParseEventStream(data)
	if err != nil {
		return nil, err
	}
	if sampleRate <= 0 || sampleRate > 192000 || divisor <= 0 || divisor > 1000000 || reference <= 0 || reference > 1000000000 || ticks > 1000000 || int64(ticks)*divisor/reference > 3600 {
		return nil, fmt.Errorf("invalid event stream clock")
	}
	c := &CMF{opl: New(sampleRate), sr: sampleRate}
	c.opl.Write(0x01, 0x20)
	for ch := 0; ch < 9; ch++ {
		c.opl.applyPatch(ch, 0)
	}
	var out []int16
	var time uint64
	for _, e := range events {
		time += uint64(e.Delta)
		end := int((int64(time)*divisor*int64(sampleRate) + reference - 1) / reference)
		n := end - len(out)
		if n > 0 {
			start := len(out)
			out = append(out, make([]int16, n)...)
			c.opl.Render(out[start:], n)
		}
		if e.Status != 0xff {
			c.applyEvent(cmfEvent{st: e.Status, d1: e.Data1, d2: e.Data2})
		}
	}
	return out, nil
}
