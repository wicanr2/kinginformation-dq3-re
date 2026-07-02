//go:build android || ios

package mobile

import (
	"embed"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2/mobile"
	"github.com/wicanr2/dq3_remake_ebitan/game"
)

// assets/ 於綁定前放入原版素材(DQ3.PAL/DQ3.BLK/CTY00.DAT/... + mt32/track_NN.ogg)。
// 版權素材 gitignore 不入庫;//go:embed 於 ebitenmobile bind 時打包進 .aar。
//
//go:embed assets
var assetsFS embed.FS

func init() {
	assets, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		log.Fatalf("mobile assets: %v", err)
	}
	music, _ := fs.Sub(assetsFS, "assets/mt32") // 無 mt32 → 靜音降級(gaudio 容忍)

	g, err := game.NewGame(assets, music)
	if err != nil {
		log.Fatalf("mobile NewGame: %v", err)
	}
	mobile.SetGame(g)
}

// Dummy 供 gomobile 產生的膠水碼引用(bind 需套件至少匯出一個符號)。
func Dummy() {}
