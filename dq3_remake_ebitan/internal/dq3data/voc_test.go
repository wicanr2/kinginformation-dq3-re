package dq3data

import "testing"

func TestDecodeVOCBank(t *testing.T) {
	raw := findAsset(t, "FVOC.VCX")
	bank := DecodeVOCBank(raw, 44100)
	if len(bank) == 0 {
		t.Fatal("FVOC.VCX 解出 0 音效")
	}
	// 至少一個音效有 PCM,且波形非全靜音
	nonEmpty, hasWave := 0, false
	for _, s := range bank {
		if len(s.PCM) > 0 {
			nonEmpty++
			for _, v := range s.PCM {
				if v != 0 {
					hasWave = true
					break
				}
			}
		}
	}
	if nonEmpty == 0 {
		t.Fatal("無任何非空音效(疑解碼有誤)")
	}
	if !hasWave {
		t.Fatal("音效全靜音(疑解碼有誤)")
	}
	// BP=1 以 VCX offset table 的第二個 dword 為起點；第一個 dword
	// 是 cue 0 的 file offset。這鎖定 cue index 與原始 block
	// rate/sample count，避免只測「能播」卻錯一格。
	cue1 := bank[1]
	if cue1.FileOffset != 0x4e6 || cue1.BlockLength != 1163 || cue1.RateDiv != 102 ||
		cue1.SourceRate != 6493 || cue1.SourceSamples != 1161 {
		t.Fatalf("FVOC cue1 metadata=%+v, want offset=0x4e6 block=1163 div=102 rate=6493 samples=1161", cue1)
	}
	if got := cue1.SourceDurationNanos(); got <= 0 || got/1_000_000 != 178 {
		t.Fatalf("FVOC cue1 source duration=%d ns, want about 178 ms", got)
	}
	t.Logf("FVOC.VCX 解出 %d 音效(%d 非空,有波形)✓", len(bank), nonEmpty)
}
