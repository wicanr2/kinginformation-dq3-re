# 三平台桌面發佈與推廣片（2026-08-10）

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260810-linux-amd64/dq3-remake-20260810-x86_64.AppImage` | 8,530,424 | `d14dd35952e88176d896526af1fdfa1ce8ec3fdea675b66c40ada5f0414d1b0b` |
| Windows x86_64 | `dist/dq3-20260810-windows-x86_64.zip` | 6,434,554 | `c8251e29e2226e163c4e207b2148d7ba0badac8cf242b82468d3b71042d388de` |
| macOS Intel | `dist/dq3-20260810-macos-x86_64.zip` | 6,569,854 | `de8960074643f8c08312f15c6186d256872fff653de8eb012445bcc1f5f7d4c9` |
| macOS Apple Silicon | `dist/dq3-20260810-macos-arm64.zip` | 6,112,271 | `3fb1efd48587eb2f63f5637ab13b1191089adb69c81d42df8a76776a1d559e89` |

Linux 檔案是使用專案固定的 `work/.tools/runtime-x86_64` 與 `appimagetool` 產出的
Type 2 AppImage，已在無 FUSE 的 Docker 內以 `--appimage-extract` 解包驗證；解包後的
內嵌 ELF 路徑為 `squashfs-root/usr/bin/dq3-remake`，其 SHA-256 與同一目錄保留的
`dq3-remake-linux-amd64` 直接 ELF 相同（13,261,768 bytes，
`7cdaf1e7627a7370fca9cc43e12f5f8cfb971e30d5eb7179de9469f1789a75a8`）。macOS ZIP 內含
`Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面外部 game pack 仍可在
不重新編譯的情況下替換，但需通過 schema、reference、content hash 及資產路徑驗證。
本次 macOS 內層執行檔為 Mach-O `x86_64`（12,433,704 bytes，SHA-256
`26818c18a565be481971cd7f4352ce39b5e536099455ca054d0b3413785e06c1`）與 `arm64`
（11,849,602 bytes，SHA-256
`887d883a2a619f2b849434ec02a056d3b8ebc295d445bdaa33dc94cd08cd362e`）；Windows PE
執行檔為 12,674,048 bytes，SHA-256
`a2af3ff359f8020fb1e3b873a2714b31a168063810b37d3f2b86a1b8a2065af7`。

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
- 本輪狀況窗切片：`go test ./internal/...` 全數通過；Docker＋Xvfb 的
  `go test ./game -run "Test(Panel|OpeningParity|CmdMenuDrawHits|Equipment|Field)"` 通過，
  並完成桌面交叉編譯。
- 本輪 CTY 遭遇閘門：`go test ./internal/dq3data -run "Test(OpenTown.*Encounter|Encounter)"`
  與 `go test . -run "Test(Save|Encounter)"` 均在 Docker＋Xvfb 通過。
- 完整長測試套件依使用者指示不再重跑；這不取代上述正式主線 trace，也不把尚未 V3 的
  戰鬥逐動作 timing、動畫、音效、抗性與掉落宣稱為完成。

## 重建入口

macOS `.app`／ZIP 由 `tools/package_macos_bundle.sh`（`RELEASE_DATE=20260810`）產生；
交叉編譯使用 `u5cht/osxcross:latest` 的 MacOSX 15.5 SDK、Go 1.24.13，IDA／原版資料未進
產物。Linux AppImage 使用 `pto2-remake-build:latest`、Go 1.26.5；Windows ZIP 使用明確
版本的 `dq3-ebiten-windows:20260810`（由 `dq3_remake_ebitan/Dockerfile.windows` 建立，
Go 1.26.5＋Debian bookworm MinGW-w64）；本輪
詳細狀況窗資料化後，以目前 Go/Ebitengine 原始碼重新編譯四個桌面執行檔。AppImage 內嵌
執行檔 SHA-256 為 `7cdaf1e7627a7370fca9cc43e12f5f8cfb971e30d5eb7179de9469f1789a75a8`，
與旁置 Linux ELF 相同。所有發佈包雜湊均為本輪重建後重新計算，不沿用舊版數值。
