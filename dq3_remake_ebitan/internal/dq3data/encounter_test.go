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
	want := []int{8, 9, 11, 7}
	if s.Threshold != 99 || s.Background != 0 || len(s.Candidates) != len(want) {
		t.Fatalf("region1/sub0 解碼錯: %+v", s)
	}
	for i := range want {
		if s.Candidates[i] != want[i] {
			t.Fatalf("region1/sub0 candidates=%v, want %v", s.Candidates, want)
		}
	}
}
