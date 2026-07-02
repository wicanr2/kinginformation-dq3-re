package dq3data

// 遊戲文字容器(D3TXTnn.TXT)+ 字型(D3TXT00.FON)。移植自 C dq3_text.c / dq3_font.c(docs/02/03)。
// 容器:檔首 16-bit LE 指標表(ptr[0]=表大小);記錄 i = data[ptr[i]..ptr[i+1]],
// 為一串 2-byte LE 值,0xffff 終止。值 <1476 = FON 第 N 個 16×16 字模;>=0xffed = 控制/插值碼。

// 文字控制碼(值 = C include/dq3_text.h)。
const (
	GlyphPx  = 16
	TxtNL    = 0xfffe // 換行
	TxtNL2   = 0xfffd
	TxtPage  = 0xfffc // 換頁
	TxtEnd   = 0xffff // 記錄結束
	txtVarLo = 0xffed // >= 此值 = 控制/插值占位(渲染為空白)
	GlyphMax = 1476   // >= 此值不畫(非字模)
)

// Text 是一個文字檔(共用 fon 字型)。
type Text struct {
	fon, txt []byte
	NRecords int
}

func le16(d []byte, o int) uint16 {
	if o < 0 || o+1 >= len(d) {
		return 0
	}
	return uint16(d[o]) | uint16(d[o+1])<<8
}

func le32(d []byte, o int) uint32 {
	if o < 0 || o+3 >= len(d) {
		return 0
	}
	return uint32(d[o]) | uint32(d[o+1])<<8 | uint32(d[o+2])<<16 | uint32(d[o+3])<<24
}

// LoadText:fon = D3TXT00.FON、txt = D3TXTnn.TXT。移植 dq3_text_load。
func LoadText(fon, txt []byte) *Text {
	nr := int(le16(txt, 0))/2 - 1 // 指標表項數 -1(末項=結尾哨兵)
	if nr < 0 {
		nr = 0
	}
	return &Text{fon: fon, txt: txt, NRecords: nr}
}

// Record 取記錄 rec 的 2-byte 值序列(不含 0xffff 終止);越界回 nil。移植 dq3_text_record。
func (t *Text) Record(rec int) []uint16 {
	if t == nil || rec < 0 || rec >= t.NRecords {
		return nil
	}
	start := int(le16(t.txt, rec*2))
	end := int(le16(t.txt, (rec+1)*2))
	if start > len(t.txt) || end > len(t.txt) || end < start {
		return nil
	}
	var out []uint16
	for o := start; o+1 < end; o += 2 {
		v := le16(t.txt, o)
		if v == TxtEnd {
			break
		}
		out = append(out, v)
	}
	return out
}

// Glyph 解 FON 第 idx 個 16×16 字模(每字 32B,2B/列 MSB-first,無 header)。移植 dq3_font_glyph。
// 回 (bitmap, ok);ok=false 表 idx 越界。
func (t *Text) Glyph(idx int) ([16][16]byte, bool) {
	var g [16][16]byte
	base := idx * 32
	if idx < 0 || base+32 > len(t.fon) {
		return g, false
	}
	for y := 0; y < 16; y++ {
		row := uint16(t.fon[base+y*2])<<8 | uint16(t.fon[base+y*2+1])
		for x := 0; x < 16; x++ {
			g[y][x] = byte((row >> (15 - x)) & 1)
		}
	}
	return g, true
}
