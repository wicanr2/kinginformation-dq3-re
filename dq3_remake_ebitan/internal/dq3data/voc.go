package dq3data

// 數位音效(FVOC.VCX / NVOC.VCX)解碼(移植 dq3_voc.c,docs/57)。
// VCX bank:開頭遞增 dword 偏移表(第一筆為 cue 0 的 file offset)→ 每筆指向一段 Creative VOC sound-data block
// (byte0=0x01 type、byte1-3=length、byte4=freq div(rate=1e6/(256-div))、byte5=codec(0=8-bit unsigned PCM))。
// 每音效重取樣到 targetRate 的 mono s16。

// SFX 是一段已重取樣的 mono s16 音效。Source* 保留 VCX/VOC 原始
// block 的取樣契約，讓播放器可以用原始 sample duration 排程，而不把
// 44100Hz 重取樣長度誤當成 DOS driver 的證據。
type SFX struct {
	PCM           []int16
	FileOffset    int
	BlockLength   int
	RateDiv       int
	SourceRate    int
	SourceSamples int
	Codec         byte
}

// SourceDurationNanos returns the duration implied by the original VOC
// sample count/rate.  Zero means the block was not a supported PCM payload.
func (s SFX) SourceDurationNanos() int64 {
	if s.SourceRate <= 0 || s.SourceSamples <= 0 {
		return 0
	}
	return int64(s.SourceSamples) * 1_000_000_000 / int64(s.SourceRate)
}

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
		out[k] = decodeVOCBlock(data[offs[k]:end], targetRate, offs[k])
	}
	return out
}

func decodeVOCBlock(p []byte, targetRate, fileOffset int) SFX {
	if len(p) < 6 || p[0] != 0x01 { // 只處理 type-1 sound data
		return SFX{FileOffset: fileOffset}
	}
	blen := int(p[1]) | int(p[2])<<8 | int(p[3])<<16
	div := int(p[4])
	meta := SFX{FileOffset: fileOffset, BlockLength: blen, RateDiv: div, Codec: p[5]}
	if p[5] != 0 { // 只支援 8-bit unsigned PCM
		return meta
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
		return meta
	}
	if 6+srcN > len(p) {
		srcN = len(p) - 6
	}
	if srcN <= 0 || targetRate <= 0 {
		return meta
	}
	pcm := p[6:]
	dstN := srcN * targetRate / srcRate
	if dstN <= 0 {
		return meta
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
	meta.PCM = s
	meta.SourceRate = srcRate
	meta.SourceSamples = srcN
	return meta
}
