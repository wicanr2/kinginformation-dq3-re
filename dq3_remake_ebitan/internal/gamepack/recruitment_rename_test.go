package gamepack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRecruitmentRenameOriginalDataParity(t *testing.T) {
	p, e := BuiltinDQ3()
	if e != nil {
		t.Fatal(e)
	}
	r := p.Interface.RecruitmentSelection.ViewRename
	if r.InputAction != "rename" || r.TargetScope != "singleton_party_leader" || r.RequireNonemptyName == nil || !*r.RequireNonemptyName {
		t.Fatal("finite rename primitive differs")
	}
	b, e := os.ReadFile(filepath.Join("..", "..", "..", "assets_raw", "DQ3.EXE"))
	if e != nil {
		t.Fatal(e)
	}
	if len(b) != 115282 || fmt.Sprintf("%x", sha256.Sum256(b)) != "5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c" {
		t.Fatal("original identity differs")
	}
	// IDA linear 的 file 基準是 linear−EC90；保留原始位址與 bytes。
	for _, row := range []struct {
		linear int
		bytes  string
	}{
		{0x10686, "80fc25"}, {0x1068b, "e81100"}, {0x106a8, "e8b481"},
		{0x1885f, "a07750"}, {0x18869, "c70622070100"}, {0x106ba, "8bbf154f"},
		{0x106c3, "803e260701"}, {0x106ce, "8d360e27"}, {0x106d2, "83c703"},
		{0x106d5, "b91200"}, {0x106d8, "f3a4"}, {0x10d84, "833e0a2701"},
	} {
		expected, e := hex.DecodeString(row.bytes)
		if e != nil {
			t.Fatal(e)
		}
		off := row.linear - 0xec90
		if fmt.Sprintf("%x", b[off:off+len(expected)]) != row.bytes {
			t.Fatalf("original IDA linear%x / file%x differs", row.linear, off)
		}
	}
	t.Log("pack canonical", p.ContentHash())
}

func TestRecruitmentRenameRejectsBrokenContract(t *testing.T) {
	for _, name := range []string{"missing", "action", "scope", "missing_nonempty", "empty_allowed", "unreviewed", "evidence"} {
		t.Run(name, func(t *testing.T) {
			p, e := BuiltinDQ3()
			if e != nil {
				t.Fatal(e)
			}
			r := p.Interface.RecruitmentSelection.ViewRename
			switch name {
			case "missing":
				p.Interface.RecruitmentSelection.ViewRename = nil
			case "action":
				r.InputAction = "unknown"
			case "scope":
				r.TargetScope = "viewed_roster"
			case "missing_nonempty":
				r.RequireNonemptyName = nil
			case "empty_allowed":
				v := false
				r.RequireNonemptyName = &v
			case "unreviewed":
				r.Evidence.Level = "D1"
			case "evidence":
				r.Evidence.Source = ""
			}
			if e = p.validateRecruitmentSelection(); e == nil {
				t.Fatal("broken rename contract accepted")
			}
		})
	}
}
