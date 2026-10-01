package dos

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/dosgolem/internal/cpu"
)

// 此檔由驗證工具複製到 dosgolem 的可丟棄副本，不屬於遊戲引擎。
// 平台依據與工具修正範圍見 docs/196-dosgolem-opening-sequence-parity.md。
func TestPaddedASCIIZFilename(t *testing.T) {
	for _, requested := range []string{"TITA    .P  ", `C:\DQ3\TITA    .P  `, "tita    .p  "} {
		t.Run(requested, func(t *testing.T) {
			m, d := newTest(t)
			if err := os.WriteFile(filepath.Join(d.Root, "TITA.P"), []byte("original"), 0o644); err != nil {
				t.Fatal(err)
			}
			m.CPU.Seg[cpu.DS], m.CPU.R[cpu.DX] = 0x3000, 0
			m.WriteBytes(cpu.Addr(0x3000, 0), append([]byte(requested), 0))
			call(m, d, 0x21, 0x3D00)
			if m.CPU.Flags&cpu.CF != 0 {
				t.Fatalf("開檔失敗：%v", d.Missing)
			}
			h := m.CPU.R[cpu.AX]
			m.CPU.R[cpu.BX], m.CPU.R[cpu.CX], m.CPU.R[cpu.DX] = h, 8, 0x40
			call(m, d, 0x21, 0x3F00)
			if m.CPU.R[cpu.AX] != 8 {
				t.Fatalf("讀取長度 %d", m.CPU.R[cpu.AX])
			}
			for i, want := range []byte("original") {
				if got := m.Read8(cpu.Addr(0x3000, 0x40) + uint32(i)); got != want {
					t.Fatalf("位元組 %d：%x != %x", i, got, want)
				}
			}
		})
	}
}
