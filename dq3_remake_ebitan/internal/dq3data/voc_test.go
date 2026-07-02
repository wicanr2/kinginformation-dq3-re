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
	t.Logf("FVOC.VCX 解出 %d 音效(%d 非空,有波形)✓", len(bank), nonEmpty)
}
