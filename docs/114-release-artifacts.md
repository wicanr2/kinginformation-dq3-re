# 三平台桌面發佈與推廣片（2026-08-10）

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260810-linux-amd64/dq3-remake-20260810-x86_64.AppImage` | 4,159,992 | `ceaf0f380d5a02cd998d514a79b08ee443e9d8afeeb77551463569ec67f519a6` |
| Windows x86_64 | `dist/dq3-20260810-windows-x86_64.zip` | 3,375,981 | `3eb7b68b5993caf208adec626520e9064bd574afc8148c09824cf42bcf6794f7` |
| macOS Intel | `dist/dq3-20260810-macos-x86_64.zip` | 3,180,415 | `00c0407031bd95eafc7003a4b0fd7a5a467b3e8e57859952e1fdad5cbb62f93b` |
| macOS Apple Silicon | `dist/dq3-20260810-macos-arm64.zip` | 2,976,041 | `aa5399c0b4b79e5afca32df363334fc29c729c63dd30586e36613db6eea00aa9` |

Linux 檔案是使用專案固定的 `work/.tools/runtime-x86_64` 與 `appimagetool` 產出的
Type 2 AppImage，已在無 FUSE 的 Docker 內以 `--appimage-extract` 解包驗證；同一目錄
另保留 `dq3-remake-linux-amd64` 直接 ELF 供診斷。macOS ZIP 內含
`Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面外部 game pack 仍可在
不重新編譯的情況下替換，但需通過 schema、reference、content hash 及資產路徑驗證。

## 推廣片

`dist/dq3-promo-20260810.mp4` 由版控的 Ebiten runtime 截圖 `docs/img/*.png` 剪輯成
1280×700、30 fps、H.264 的 45.434 秒無音軌推廣片，SHA-256 為
`d801efbbf6d57dd627308059b5420dfe35f72c647cf0b60ff44a6d0cde015089`，大小 3,501,430
bytes。片中依序展示開場城鎮、城鎮行走、船、日邦格八頭大蛇兩戰、商人事件、黃金球、
耶進貝亞解謎、黑暗之燈、勇氣洞窟、最終鑰匙與提頓門；沒有把診斷用文字或舊 C/SDL
錄製幀當成目前 Go/Ebitengine 畫面。

## 驗證界線

- `TestOpeningProductionInputTrace`：Docker＋Xvfb，正式 `InputState` 從新遊戲到 `THE END`，
  311.44 秒通過。
- 其餘 `game` 測試（排除上述長 campaign trace）：Docker＋Xvfb，29.797 秒通過。
- 本輪 CTY 遭遇閘門：`go test ./internal/dq3data -run "Test(OpenTown.*Encounter|Encounter)"`
  與 `go test . -run "Test(Save|Encounter)"` 均在 Docker＋Xvfb 通過。
- 完整長測試套件依使用者指示不再重跑；這不取代上述正式主線 trace，也不把尚未 V3 的
  戰鬥逐動作 timing、動畫、音效、抗性與掉落宣稱為完成。

## 重建入口

macOS `.app`／ZIP 由 `tools/package_macos_bundle.sh`（`RELEASE_DATE=20260810`）產生；
交叉編譯使用 `u5cht/osxcross:latest` 的 MacOSX 15.5 SDK、Go 1.24.13，IDA／原版資料未進
產物。Linux AppImage 與 Windows ZIP 使用 `pto2-remake-build:latest`、Go 1.26.5 建置；
本輪框線資料化後以目前 Go/Ebitengine 原始碼重新編譯四個桌面執行檔；AppImage 內嵌
執行檔 SHA-256 為 `969a2ea135a8d42a1e8c4180c1e7f828f98b75a5254a6411ddcae7775e738dac`，
與旁置 Linux ELF 相同。所有發佈包雜湊均為本輪重建後重新計算，不沿用舊版數值。
