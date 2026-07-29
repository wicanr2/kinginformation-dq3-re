// dq3_remake_ebitan — 精訊 DQ3 remake 的 Go/Ebiten port(桌面進入點)。
// 遊戲邏輯在 game/ 套件(桌面與 Android/mobile 共用);此檔只做「桌面組裝 + 開窗」。
package main

import (
	"io/fs"
	"log"
	"os"

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
	assets := os.DirFS(assetsDir())
	var music fs.FS // DQ3_MT32 指向 track_NN.ogg 目錄;未設 → 靜音
	if d := os.Getenv("DQ3_MT32"); d != "" {
		music = os.DirFS(d)
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

	ebiten.SetWindowSize(game.ScreenW*2, game.ScreenH*2)
	ebiten.SetWindowTitle("Dragon Fighter III — Ebiten port (overworld + アリアハン)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
