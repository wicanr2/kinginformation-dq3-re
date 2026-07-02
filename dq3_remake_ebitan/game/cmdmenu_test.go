package game

import "testing"

// 命令窗游標導覽:2 欄 × 3 列繞回(對拍 dq3_cmdmenu_input 的方向邏輯)。
func TestCmdMenuCursor(t *testing.T) {
	var m CmdMenu
	m.Open()
	if m.cursor != 0 {
		t.Fatalf("初始 cursor=%d,應 0", m.cursor)
	}
	// 下:0(對話)→2(狀況)→4(裝備)→0 繞回(3 列)
	for _, want := range []int{2, 4, 0} {
		m.move(0)
		if m.cursor != want {
			t.Fatalf("下移後 cursor=%d,應 %d", m.cursor, want)
		}
	}
	// 上:0→4→2→0(反向繞回)
	for _, want := range []int{4, 2, 0} {
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
