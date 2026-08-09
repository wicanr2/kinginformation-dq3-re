# 三平台桌面發佈與推廣片（2026-08-10）

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260810-linux-amd64/dq3-remake-20260810-x86_64.AppImage` | 7,305,720 | `5e5a8af3cf2c41462874d5e564f6af5f6334a0d6dccdc329eb783b16db5902a9` |
| Windows x86_64 | `dist/dq3-20260810-windows-x86_64.zip` | 6,420,395 | `f02a8c74608dbd199057adf81d419c6f27a68d6f270aa61fd2d4f41aa4690c9c` |
| macOS Intel | `dist/dq3-20260810-macos-x86_64.zip` | 6,558,546 | `71f8d159a528862f6b6431772ad7fb86236bde7222cddf875b896a7958861476` |
| macOS Apple Silicon | `dist/dq3-20260810-macos-arm64.zip` | 6,104,807 | `b4cd7247a9f525f755bd3bcb529f5bee2a264b9c8f8b3390eac9d8d1c91f58f0` |

Linux 檔案是使用專案固定的 `work/.tools/runtime-x86_64` 與 `appimagetool` 產出的
Type 2 AppImage，已在無 FUSE 的 Docker 內以 `--appimage-extract` 解包驗證；解包後的
內嵌 ELF 路徑為 `squashfs-root/usr/bin/dq3-remake`，其 SHA-256 與同一目錄保留的
`dq3-remake-linux-amd64` 直接 ELF 相同（13,228,568 bytes，
`c7a3dad9d71b3c6dbff3d541f89d121c9d2d3484bb4b771f51df1449dd1b90f1`）。macOS ZIP 內含
`Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面外部 game pack 仍可在
不重新編譯的情況下替換，但需通過 schema、reference、content hash 及資產路徑驗證。
本次 macOS 內層執行檔為 Mach-O `x86_64`（12,404,648 bytes，SHA-256
`e764517c54331a8394becd5d3f04540167264232b9a9cccef034123f65767910`）與 `arm64`
（11,816,114 bytes，SHA-256
`aac3381571210869e1f63103b1c50b0c64952870725c779c4d67526d0ae3c630`）；Windows PE
執行檔為 12,638,720 bytes，SHA-256
`81c908fff9666b7549ec8c71cdfe7b02aae91dd3b6ce88b6678b7723116ad0a2`。

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
本輪敵人數量列基線資料化後，以目前 Go/Ebitengine 原始碼重新編譯四個桌面執行檔；AppImage 內嵌
執行檔 SHA-256 為 `c7a3dad9d71b3c6dbff3d541f89d121c9d2d3484bb4b771f51df1449dd1b90f1`，
與旁置 Linux ELF 相同。所有發佈包雜湊均為本輪重建後重新計算，不沿用舊版數值。
