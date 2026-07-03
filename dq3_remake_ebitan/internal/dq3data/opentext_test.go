package dq3data

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// loadUnicodeMap 讀 docs/data/glyph_unicode_map.json → glyph index → unicode(注意有形近誤標)。
func loadUnicodeMap(t *testing.T) map[uint16]string {
	p := os.Getenv("GLYPH_MAP")
	if p == "" {
		p = "../../../docs/data/glyph_unicode_map.json"
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Skipf("無 glyph map:%v", err)
	}
	var sm map[string]string
	if err := json.Unmarshal(raw, &sm); err != nil {
		t.Fatal(err)
	}
	m := map[uint16]string{}
	for k, v := range sm {
		if n, err := strconv.Atoi(k); err == nil {
			m[uint16(n)] = v
		}
	}
	return m
}

// TestFindOpeningText:在各 D3TXT bank 搜開場旁白(DOSBox 實測拿到的字串),
// 定位開場劇情的 rec 號(C-2 需要)。DQ3_FIND_OPEN=1 才跑。
// glyph→unicode 用 docs/data/glyph_unicode_map.json(注意:map 有形近誤標,結果需交叉驗證)。
func TestFindOpeningText(t *testing.T) {
	if os.Getenv("DQ3_FIND_OPEN") == "" {
		t.Skip("設 DQ3_FIND_OPEN=1 才執行")
	}
	dir := os.Getenv("DQ3_ASSETS")
	if dir == "" {
		dir = "../../assets_raw"
	}
	umap := loadUnicodeMap(t)
	fon := mustRead(t, filepath.Join(dir, "D3TXT00.FON"))

	// 開場相關關鍵字(實測旁白 + 攻略)
	needles := []string{"重要", "一天", "國王", "晉見", "走", "生日", "母", "十六", "王城"}

	banks, _ := filepath.Glob(filepath.Join(dir, "D3TXT0?.TXT"))
	for _, bank := range banks {
		txt := LoadText(fon, mustRead(t, bank))
		if txt == nil {
			continue
		}
		bname := filepath.Base(bank)
		for rec := 0; rec < txt.NRecords; rec++ {
			s := recToStr(txt.Record(rec), umap)
			for _, n := range needles {
				if strings.Contains(s, n) {
					t.Logf("%s rec%d 含「%s」: %s", bname, rec, n, clip(s, 40))
					break
				}
			}
		}
	}
}

func recToStr(rec []uint16, umap map[uint16]string) string {
	var b strings.Builder
	for _, g := range rec {
		if u, ok := umap[g]; ok {
			b.WriteString(u)
		} else {
			b.WriteString("·")
		}
	}
	return b.String()
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}
