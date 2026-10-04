package game

import (
	"os"
	"strings"
	"testing"
)

func TestNewGameRejectsItemStorageMetadataMismatch(t *testing.T) {
	pack := loadTestPack(t)
	// Keep initial armor unchanged. A different archive record must still be
	// checked by the real production boot, not merely by an isolated validator.
	*pack.Characters.ItemStorage.Items[0].EquipmentPart = 1
	g, err := NewGameWithPack(os.DirFS(spineAssetsDir(t)), nil, pack)
	if err == nil || g != nil || !strings.Contains(err.Error(), "item storage") {
		t.Fatalf("production boot accepted mismatched item metadata: game=%p err=%v", g, err)
	}
}
