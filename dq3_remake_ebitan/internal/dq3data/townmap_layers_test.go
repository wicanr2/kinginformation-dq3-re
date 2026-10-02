package dq3data

import (
	"encoding/binary"
	"testing"
)

func TestTownRejectsTruncatedLayerHeader(t *testing.T) {
	raw := make([]byte, 0x16)
	binary.LittleEndian.PutUint16(raw[2:], 1)
	binary.LittleEndian.PutUint16(raw[4:], 1)
	binary.LittleEndian.PutUint16(raw[0x0e:], 2)
	if _, err := OpenTown(raw, 0, false); err == nil {
		t.Fatal("accepted header without second layer tile")
	}
}
