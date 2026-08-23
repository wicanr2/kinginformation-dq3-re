// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(桌面進入點)。
// 遊戲邏輯在 game/ 套件(桌面與 Android/mobile 共用);此檔只做「桌面組裝 + 開窗」。
package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/dq3_remake_ebitan/game"
	"github.com/wicanr2/dq3_remake_ebitan/internal/gamepack"
)

func assetsDir() string {
	if d := os.Getenv("DQ3_ASSETS"); d != "" {
		return d
	}
	return "assets_raw"
}

func main() {
	assetPath := assetsDir()
	assets := os.DirFS(assetPath)
	var music fs.FS // DQ3_MT32 明確覆蓋；完整版預設讀 assets_raw/mt32，缺少才回 SB-FM MBG.MCX。
	if d := os.Getenv("DQ3_MT32"); d != "" {
		music = os.DirFS(d)
	} else if _, err := os.Stat(filepath.Join(assetPath, "mt32", "track_00.ogg")); err == nil {
		music = os.DirFS(filepath.Join(assetPath, "mt32"))
	}

	var g *game.Game
	var err error
	if dir := os.Getenv("DQ_GAME_PACK"); dir != "" {
		pack, loadErr := gamepack.Load(os.DirFS(dir))
		if loadErr != nil {
			log.Fatalf("載入 DQ_GAME_PACK %q 失敗: %v", dir, loadErr)
		}
		g, err = game.NewGameWithPack(assets, music, pack)
	} else {
		g, err = game.NewGame(assets, music)
	}
	if err != nil {
		log.Fatalf("NewGame 失敗:%v(設 DQ3_ASSETS 指向原版素材夾)", err)
	}
	g.StartOpeningCutscene()

	ebiten.SetWindowSize(game.ScreenW*2, game.ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (overworld + アリアハン)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
