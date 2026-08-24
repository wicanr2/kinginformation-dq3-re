package game

import (
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func TestZhuyinWangCompatibilityAlias(t *testing.T) {
	ni := NameInput{labels: &gamepack.NewGameLabels{ZhuyinAliases: []gamepack.ZhuyinCandidateAlias{{
		Sh: 0, Ji: 2, Yu: 11, Tone: 2, Glyphs: []int{186},
	}}}}
	ni.Init()
	ni.zh.Ji, ni.zh.Yu = 2, 11
	ni.cursor = niCellToRaw(38) // ˇ
	ni.input(InputState{Confirm: true}, -1)
	if !ni.zh.Pick {
		t.Fatal("ㄨㄤˇ 應進入候選")
	}
	found := false
	for _, glyph := range ni.zh.Cand {
		found = found || glyph == 186
	}
	if !found {
		t.Fatalf("ㄨㄤˇ 相容候選應含王 glyph186，got %v", ni.zh.Cand)
	}
}
