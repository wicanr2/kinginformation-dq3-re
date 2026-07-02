package dq3data

// 角色 sprite(DQ3MST.BLS 主角 / DQ3MAN.BLS NPC)。移植自 C dq3_sprite.c(docs/27)。
// 檔:6B 表頭 + entry × 960(stride）。每 entry(一方向)含 2 個 walk sub-frame(各 480B:
// 384B 像素 + 96B mask@+0x180)。每 frame=32×24,4-plane plane-major(plane 順序 3,2,1,0);
// mask bit=1→透明。角色 = entry_base 起的 4 連續 entry(4 方向)×2 walk = 8 frame。
const (
	CharW, CharH = 32, 24
	CharDirs     = 4
	CharWalk     = 2
	CharFrames   = CharDirs * CharWalk // 8:方向×2 + walk
	blsStride    = 960
	blsSubframe  = 480
	blsMaskOff   = 0x180
	blsBody0     = 6
)

// Frame 是一張角色 frame:indexed 像素 + 不透明遮罩。
type Frame struct {
	Px     [CharH][CharW]uint8 // 0..15,對 palette
	Opaque [CharH][CharW]bool  // true=畫、false=透明
}

// CharSprite 是一個角色的 8 frame。
type CharSprite struct {
	Frames [CharFrames]Frame
}

func decodeFrame(d []byte, base int) Frame {
	var f Frame
	order := [4]int{3, 2, 1, 0}
	const wb = 4
	seg := 4 * CharH // 96 bytes/plane
	for s := 0; s < 4; s++ {
		b0 := base + s*seg
		pl := order[s]
		for r := 0; r < CharH; r++ {
			for b := 0; b < wb; b++ {
				o := b0 + r*wb + b
				var v byte
				if o >= 0 && o < len(d) {
					v = d[o]
				}
				for bit := 0; bit < 8; bit++ {
					if v&(0x80>>bit) != 0 {
						f.Px[r][b*8+bit] |= 1 << pl
					}
				}
			}
		}
	}
	mb := base + blsMaskOff
	for r := 0; r < CharH; r++ {
		for b := 0; b < wb; b++ {
			o := mb + r*wb + b
			v := byte(0xff) // 越界視為全透明
			if o >= 0 && o < len(d) {
				v = d[o]
			}
			for bit := 0; bit < 8; bit++ {
				f.Opaque[r][b*8+bit] = v&(0x80>>bit) == 0
			}
		}
	}
	return f
}

// LoadCharSprite 解一個角色的 8 frame。移植 dq3_charsprite_load。
func LoadCharSprite(d []byte, entryBase int) *CharSprite {
	cs := &CharSprite{}
	for fr := 0; fr < CharDirs; fr++ {
		slot := blsBody0 + (entryBase+fr)*blsStride
		for w := 0; w < CharWalk; w++ {
			cs.Frames[fr*CharWalk+w] = decodeFrame(d, slot+w*blsSubframe)
		}
	}
	return cs
}
