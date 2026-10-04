package game

import (
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
	"testing"
)

// 正常原版195..205的按欄順序，含左欄底部跨到右欄頂部。
func TestCmdMenuCursor(t *testing.T) {
	var m CmdMenu
	p, err := gamepack.BuiltinDQ3()
	if err != nil {
		t.Fatal(err)
	}
	if err = m.configure(p.Interface.FieldCommandMenu); err != nil {
		t.Fatal(err)
	}
	m.Open()
	if m.cursor != 0 {
		t.Fatalf("初始 cursor=%d,應 0", m.cursor)
	}
	// 下：對話、狀況、裝備、咒文、道具、調查、對話。
	for _, want := range []int{2, 4, 1, 3, 5, 0} {
		m.move(0)
		if m.cursor != want {
			t.Fatalf("下移後 cursor=%d,應 %d", m.cursor, want)
		}
	}
	for _, want := range []int{5, 3, 1, 4, 2, 0} {
		m.move(1)
		if m.cursor != want {
			t.Fatalf("上移後 cursor=%d,應 %d", m.cursor, want)
		}
	}
	// 左/右:欄切換(0↔1)
	m.move(3) // 右
	if m.cursor != 1 {
		t.Fatalf("右移後 cursor=%d,應 1(咒文)", m.cursor)
	}
	m.move(2) // 左
	if m.cursor != 0 {
		t.Fatalf("左移後 cursor=%d,應 0(對話)", m.cursor)
	}
	// 列 + 欄組合:下到第 2 列再右 → cursor=3(道具)
	m.move(0)
	m.move(3)
	if m.cursor != 3 {
		t.Fatalf("(下+右)後 cursor=%d,應 3(道具)", m.cursor)
	}
}
