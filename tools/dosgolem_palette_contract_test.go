package dos

import (
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

// BIOS 公開契約：RBIL INT 10/AX=1013h，以及 DOSBox Staging
// src/ints/int10_pal.cpp 的 INT10_SelectDACPage。檢驗兩種頁面模式
// 與色號到 DAC 的實際映射，不能只檢查服務不再回報 unknown。
func TestDACColorPageContract(t *testing.T) {
	m, d := newTest(t)
	m.SetVideoMode(0x10)
	m.VGA.SetPal(11, 0x3b)
	selectPage := func(mode, page uint8, want uint8) {
		t.Helper()
		m.CPU.R[cpu.BX] = uint16(mode) << 8
		call(m, d, 0x10, 0x1013)
		m.CPU.R[cpu.BX] = uint16(page)<<8 | 1
		call(m, d, 0x10, 0x1013)
		if got := m.VGA.DACIndex(11); got != want {
			t.Fatalf("mode=%d page=%d: DAC=%#x want=%#x", mode, page, got, want)
		}
	}
	selectPage(1, 5, 0x5b)
	selectPage(1, 15, 0xfb)
	selectPage(0, 2, 0xbb)
	selectPage(0, 3, 0xfb)
	selectPage(1, 0, 0x0b)
	if len(d.UnimplementedReport()) != 0 {
		t.Fatal(d.UnimplementedReport())
	}
	if m.VGA.Pal(11) != 0x3b {
		t.Fatal("色盤頁面服務不應覆寫屬性色盤")
	}
}
