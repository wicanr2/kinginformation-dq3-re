package gaudio

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// 驗證 MT-32 OGG 能被 Ebiten vorbis 解碼(headless:只解碼、不建 audio context/裝置)。
func TestDecodeOgg(t *testing.T) {
	dir := os.Getenv("DQ3_MT32")
	if dir == "" {
		dir = "../../../work/mt32" // repo 內 MT-32 OGG(24 軌)
	}
	// FIELD=6、CASTLE=1(阿里阿罕)兩軌都驗
	for _, tr := range []int{6, 1} {
		p := filepath.Join(dir, fmt.Sprintf("track_%02d.ogg", tr))
		data, err := os.ReadFile(p)
		if err != nil {
			t.Skipf("無 MT-32 OGG(%s):%v", p, err)
		}
		n, err := DecodeOgg(data)
		if err != nil {
			t.Fatalf("解碼 %s 失敗:%v", p, err)
		}
		if n <= 0 {
			t.Fatalf("%s 解碼長度 %d,應 >0", p, n)
		}
		t.Logf("%s 解碼 OK,%d samples", p, n)
	}
}

// SetEnabled/SetVolume 必須 nil-safe(nil *Music 直接呼叫不 panic);SetEnabled(false)
// 後 Play/PlaySFX 不 panic(靜音降級路徑,設定選單「音樂 關」用)。
func TestSetEnabledSetVolumeNilSafe(t *testing.T) {
	var nilM *Music
	nilM.SetEnabled(false)
	nilM.SetVolume(50)
	nilM.Play(0)
	nilM.PlaySFX(0)

	m := NewMusic(nil) // 無檔案系統 → enabled=false(既有靜音降級路徑)
	m.SetEnabled(false)
	m.SetVolume(30)
	m.Play(0)
	m.PlaySFX(0)

	m.SetEnabled(true) // 重開後仍應不 panic(enabled 仍 false,因無 fsys)
	m.Play(0)
}

// SetEnabled(false) 對已初始化(enabled=true)的 Music 應立即讓 Play 變 no-op,
// 但不動 enabled 欄位本身(SFX 走獨立旗標,不受音樂開關影響)。
func TestSetEnabledGatesMusicOnly(t *testing.T) {
	m := &Music{enabled: true, vol: 0.5}
	m.SetEnabled(false)
	if !m.enabled {
		t.Error("SetEnabled(false) 不應動到 enabled(音訊後端可用旗標,SFX 共用)")
	}
	m.Play(0) // musicOff=true → no-op,不應 panic(無 ctx 也不該被呼叫到)
	m.SetEnabled(true)
	if m.musicOff {
		t.Error("SetEnabled(true) 後 musicOff 應為 false")
	}
}
