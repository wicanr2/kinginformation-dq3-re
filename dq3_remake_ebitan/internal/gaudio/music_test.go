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
