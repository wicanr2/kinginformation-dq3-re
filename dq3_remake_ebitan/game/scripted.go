package game

// scripted NPC(sub2)給物/對話表(自 dq3_scripted baked,給予型子集)。
// 欄位:{byte4,cty,give_item,prereq_item,require_item,consume_prereq,milestone,before_rec,give_rec,after_rec}。
// 255=DQ3_SC_NOITEM。cty=0xff(255)=不限。
var scriptedTable = [][10]int{
	{12, 8, 85, 255, 255, 0, 513, 54, 52, 53},
	{7, 1, 88, 85, 255, 0, 514, 65, 60, 60},
	{31, 22, 75, 255, 255, 0, 0, 45, 46, 45},
	{52, 67, 101, 255, 255, 0, 0, 71, 71, 71},
	{84, 16, 16, 255, 255, 0, 0, 98, 99, 98},
	{25, 15, 92, 255, 255, 0, 0, 119, 118, 120},
	{49, 64, 107, 255, 255, 0, 0, 85, 84, 85},
	{44, 54, 99, 98, 255, 1, 0, 56, 59, 60},
	{50, 255, 255, 255, 91, 0, 0, 89, 87, 87},
}

// scriptedFor:依 byte4(+cty)查 scripted NPC。移植 dq3_scripted_get。
func scriptedFor(byte4, cty int) *[10]int {
	for i := range scriptedTable {
		s := &scriptedTable[i]
		if s[0] == byte4 && (s[1] == 255 || s[1] == cty) {
			return s
		}
	}
	return nil
}
