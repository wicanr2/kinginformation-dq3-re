package dq3data

import "testing"

func TestEncounterTablesFromOriginalEXE(t *testing.T) {
	tab, err := OpenEncounterTables(findAsset(t, "DQ3.EXE"))
	if err != nil {
		t.Fatal(err)
	}
	if got := tab.Region(153, 174); got != 1 {
		t.Fatalf("阿里阿罕周邊 region=%d, want 1", got)
	}
	s := tab.Slot(1, 0)
	want := []int{5, 6, 5, 10}
	if s.Threshold != 1 || s.Background != 0 || len(s.Candidates) != len(want) {
		t.Fatalf("region1/sub0 解碼錯: %+v", s)
	}
	for i := range want {
		if s.Candidates[i] != want[i] {
			t.Fatalf("region1/sub0 candidates=%v, want %v", s.Candidates, want)
		}
	}
	if got := tab.Slot(2, 0); got.Threshold != 99 || got.Background != 0 {
		t.Fatalf("region2/sub0 未對齊 dec region: %+v", got)
	}
	if got := tab.Slot(0, 0); len(got.Candidates) != 0 {
		t.Fatalf("raw region 0 應 fail-closed: %+v", got)
	}
}
