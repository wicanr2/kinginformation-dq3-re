# dq3_remake_ebitan

精訊版《勇者鬥惡龍 III》的 Go／Ebitengine remake，也是本 repo 唯一現行產品線。

目前可執行且已有多個系統與主線事件切片，但尚未完成從新遊戲、只用正式玩家輸入抵達
THE END 的全程驗收。現況與工作順序以
[`docs/74-ebiten-remake-completion-plan.md`](../docs/74-ebiten-remake-completion-plan.md)
為準；不要從本文件的歷史清單推算完成度。

## 建置

原版素材不納入 repo。請自行提供合法持有的素材至 `assets_raw/`，或用
`DQ3_ASSETS` 指向素材目錄。

```bash
bash dq3_remake_ebitan/build.sh
```

桌面執行：

```bash
cd dq3_remake_ebitan
DQ3_ASSETS=/path/to/assets_raw go run .
```

遊戲設定採 versioned game pack。未指定時使用執行檔內建、經測試的 canonical
`dq3_cht`；開發時可改外部 JSON 而不重新編譯：

```bash
DQ3_ASSETS=/path/to/assets_raw \
DQ_GAME_PACK=/path/to/dq3_cht \
go run .
```

外部 pack 會嚴格檢查 schema、未知欄位、ID、範圍與 evidence；格式見
[`docs/84-game-pack-json-contract.md`](../docs/84-game-pack-json-contract.md)。
套用不同 content hash 的新格式存檔不會被靜默載入。

圖形測試需在 Xvfb 下建立 `game.test`，並從 `dq3_remake_ebitan/game/` 執行，否則
測試相對路徑不會找到 repo 的 `assets_raw/`。素材缺失造成的 `SKIP` 不算對拍通過。

Android 綁定入口在 `mobile/`，建置腳本是 `build-android.sh`；它與桌面版共用
`game/` core。實機 APK／AAB、觸控打磨與完整流程仍需另行驗收。

## 目錄角色

- `game/`：場景、事件、戰鬥、選單、存檔與 renderer。
- `internal/`：原始資料解析、公式、音訊及其他可獨立測試的模組。
- `mobile/`：Android／iOS 綁定骨架。
- `docs/`：由 production renderer 產生的檢查圖片。

## 證據規則

`dq3_remake/`、舊文件及既有 Go 程式只能作研究線索。實作參數或流程前，應以
`DQ3.EXE`、原始資料、DOSBox 同狀態實機，以及本機完整影片交叉驗證。每個功能應分別記錄：

- E：正常玩家輸入能否抵達；
- V：畫面／操作與原版的對拍程度；
- D：反組譯資料流的證據程度。

單元測試通過、內部 handler 可直接呼叫或單張截圖相似，都不能單獨標成「完成」。
