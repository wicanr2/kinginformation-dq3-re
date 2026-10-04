package gamepack

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRecruitmentViewAbilityOriginalBindingParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	exe, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "DQ3.EXE"))
	if e != nil {
		t.Fatal(e)
	}
	if fmt.Sprintf("%x", sha256.Sum256(exe)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original EXE identity differs")
	}
	// IDA9.4 linear → MZ file。由既有能力窗引用確認View與創角使用同一typed幾何。
	r := p.Interface.NewGameGeometry.Raster
	for _, w := range p.Interface.NewGameGeometry.RawWindows {
		if w.ID != r.Ability.RawWindowID {
			continue
		}
		address, e := strconv.ParseInt(strings.TrimPrefix(w.Address, "linear:"), 0, 32)
		if e != nil || address != 0x24dd0+0x3da8 {
			t.Fatal("view ability window differs from original caller")
		}
	}
	for _, row := range []struct {
		linear int
		hex    string
	}{
		{0x10677, "8d36a83d"}, {0x1067b, "e812ef"}, {0x1067e, "e8cd7c"},
		{0x10681, "9aac000411"}, {0x10686, "80fc25"}, {0x1068e, "8d36a83d"}, {0x10692, "e86fef"},
		{0x18498, "9adb000411"}, {0x1849d, "e80100"},
		{0x184ac, "8a442e02442f02443002443132e43d00007501c3"},
	} {
		want, e := hex.DecodeString(row.hex)
		if e != nil || !bytes.Equal(exe[row.linear-0xec90:row.linear-0xec90+len(want)], want) {
			t.Fatalf("original view consumer bytes differ at IDA linear %#x", row.linear)
		}
	}
}
