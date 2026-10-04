package opl2

import "testing"

func TestEventStreamUsesDeltaBeforeEventAndRunningStatus(t *testing.T) {
	data := []byte{0, 0xc0, 1, 0, 0x90, 60, 100, 0x81, 0, 60, 0, 2, 0xff, 0x2f, 0}
	events, ticks, e := ParseEventStream(data)
	if e != nil || len(events) != 4 || ticks != 130 || events[2].Delta != 128 || events[2].Status != 0x90 {
		t.Fatal(events, ticks, e)
	}
	pcm, e := RenderEventStream(data, 44100, 12428, 1193180)
	if e != nil || len(pcm) != int((int64(130)*12428*44100+1193180-1)/1193180) {
		t.Fatal("cumulative duration", len(pcm), e)
	}
	nonzero := false
	for _, s := range pcm {
		if s != 0 {
			nonzero = true
			break
		}
	}
	if !nonzero {
		t.Fatal("note interval must produce samples")
	}
}

func TestEventStreamRejectsMalformedData(t *testing.T) {
	for _, data := range [][]byte{nil, {0}, {0, 60, 100}, {0, 0x90, 60}, {0, 0x90, 60, 128}, {0, 0xff, 1, 0}, {0, 0xff, 0x2f, 0, 0}, {0, 0xf0, 0}, {0x81, 0x81, 0x81, 0x81, 0}} {
		if _, _, e := ParseEventStream(data); e == nil {
			t.Fatalf("malformed stream accepted: %x", data)
		}
	}
}
