package game

import (
	"bytes"
	"testing"

	"github.com/wicanr2/dq3_remake_ebitan/internal/dq3data"
)

// TestRequiredFixedBossBackgroundsUsePackSelectors 驗證必經單頭目戰的垂直
// runtime 鏈。原版 record 由 pack 選出；renderer 必須解出其精確 page 與
// palette bank，不得退回通用草地背景。
func TestRequiredFixedBossBackgroundsUsePackSelectors(t *testing.T) {
	g := r4Game(t)
	if !g.battle.backgroundReady {
		t.Fatal("正式遊戲未載入 battle background pack contract")
	}
	for _, monsterRawID := range []int{0x59, 0x79, 0x7a, 0x7b, 0x7c} {
		formation, ok := g.pack.FixedBattleFormationForSingleMonster(monsterRawID)
		if !ok {
			t.Fatalf("monster %#x 缺少唯一的 pack fixed formation", monsterRawID)
		}
		if !g.startBossBattle(monsterRawID) || g.battle.bg == nil {
			t.Fatalf("monster %#x 未以正式 fixed formation 啟動背景", monsterRawID)
		}
		if g.battle.bgPaletteBank != formation.Background.PaletteBankRaw {
			t.Fatalf("monster %#x palette bank=%d want pack=%d", monsterRawID,
				g.battle.bgPaletteBank, formation.Background.PaletteBankRaw)
		}
		want, ok := dq3data.DecodePackBG(g.battle.scr, formation.Background.PageRaw, dq3data.PackBGFormat{
			PageStrideBytes:  g.battle.background.PageStrideBytes,
			FieldOffsetBytes: g.battle.background.FieldOffsetBytes,
			FieldChunkBytes:  g.battle.background.FieldChunkBytes,
			WidthPixels:      g.battle.background.WidthPixels,
			FieldRows:        g.battle.background.FieldRows,
			PlaneCount:       g.battle.background.PlaneCount,
		})
		if !ok || !bytes.Equal(g.battle.bg.Pixels, want.Pixels) {
			t.Fatalf("monster %#x background page does not match pack selector", monsterRawID)
		}
		g.battle.active = false
	}
}
