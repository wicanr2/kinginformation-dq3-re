package dq3data

// 數位音效(FVOC.VCX / NVOC.VCX)解碼(移植 dq3_voc.c,docs/57)。
// VCX bank:開頭遞增 dword 偏移表(首筆=表大小)→ 每筆指向一段 Creative VOC sound-data block
// (byte0=0x01 type、byte1-3=length、byte4=freq div(rate=1e6/(256-div))、byte5=codec(0=8-bit unsigned PCM))。
// 每音效重取樣到 targetRate 的 mono s16。

// SFX 是一段已重取樣的 mono s16 音效。
type SFX struct{ PCM []int16 }

// DecodeVOCBank 解一個 VCX bank(全部重取樣到 targetRate)。回各音效(失敗者 PCM 為 nil)。
func DecodeVOCBank(data []byte, targetRate int) []SFX {
	var offs []int
	for i := 0; i+4 <= len(data); i += 4 {
		v := int(le32(data, i))
		if len(offs) > 0 && v <= offs[len(offs)-1] {
			break
		}
		if v >= len(data) || v < (len(offs)+1)*4 {
			break
		}
		offs = append(offs, v)
	}
	if len(offs) == 0 {
		return nil
	}
	out := make([]SFX, len(offs))
	for k := range offs {
		end := len(data)
		if k+1 < len(offs) {
			end = offs[k+1]
		}
		out[k] = decodeVOCBlock(data[offs[k]:end], targetRate)
	}
	return out
}

func decodeVOCBlock(p []byte, targetRate int) SFX {
	if len(p) < 6 || p[0] != 0x01 { // 只處理 type-1 sound data
		return SFX{}
	}
	blen := int(p[1]) | int(p[2])<<8 | int(p[3])<<16
	div := int(p[4])
	if p[5] != 0 { // 只支援 8-bit unsigned PCM
		return SFX{}
	}
	srcRate := 11025
	if div < 256 {
		srcRate = 1000000 / (256 - div)
	}
	if srcRate <= 0 {
		srcRate = 8000
	}
	srcN := blen - 2
	if srcN <= 0 {
		return SFX{}
	}
	if 6+srcN > len(p) {
		srcN = len(p) - 6
	}
	pcm := p[6:]
	dstN := srcN * targetRate / srcRate
	if dstN <= 0 {
		return SFX{}
	}
	s := make([]int16, dstN)
	for i := 0; i < dstN; i++ {
		si := i * srcRate / targetRate
		u := 128
		if si < srcN {
			u = int(pcm[si])
		}
		s[i] = int16((u - 128) * 200) // 8-bit unsigned(中心128)→ s16(留 headroom)
	}
	return SFX{PCM: s}
}
