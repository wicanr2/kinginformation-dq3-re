package dq3data

import "fmt"

// EncounterTables 是 DQ3.EXE 內的地表 region map 與 32×4 遭遇子表。
// file offsets = DGROUP file base 0x16140 + logical 0x4966 / 0x4a56（docs/39）。
type EncounterTables struct {
	regions [256]byte
	slots   [32][4]EncounterSlot
}

type EncounterSlot struct {
	Threshold  int
	Background int
	Candidates []int
}

const (
	encounterRegionFile = 0x16140 + 0x4966
	encounterTableFile  = 0x16140 + 0x4a56
)

func OpenEncounterTables(exe []byte) (*EncounterTables, error) {
	const tableSize = 32 * 4 * 8
	if len(exe) < encounterTableFile+tableSize || len(exe) < encounterRegionFile+256 {
		return nil, fmt.Errorf("DQ3.EXE 遭遇表截斷: %d bytes", len(exe))
	}
	t := &EncounterTables{}
	copy(t.regions[:], exe[encounterRegionFile:encounterRegionFile+256])
	for region := 0; region < 32; region++ {
		for sub := 0; sub < 4; sub++ {
			o := encounterTableFile + region*0x20 + sub*8
			s := &t.slots[region][sub]
			s.Threshold = int(exe[o+1])
			s.Background = int(exe[o+2])
			for _, id := range exe[o+4 : o+8] {
				if id != 0xff {
					s.Candidates = append(s.Candidates, int(id))
				}
			}
		}
	}
	return t, nil
}

// Region 對齊 file 0xbb45：上層以 16×16 cell 查表；下層 raw Y>=0x12c 走固定區。
func (t *EncounterTables) Region(x, y int) int {
	if t == nil {
		return 0
	}
	if y >= 0x12c {
		if x >= 0x7a {
			return 0x12
		}
		return 0x11
	}
	cell := ((x >> 4) & 0x0f) + (y & 0xf0)
	return int(t.regions[cell&0xff])
}

func (t *EncounterTables) Slot(region, sub int) EncounterSlot {
	if t == nil || region < 0 || region >= len(t.slots) {
		return EncounterSlot{}
	}
	return t.slots[region][sub&3]
}
