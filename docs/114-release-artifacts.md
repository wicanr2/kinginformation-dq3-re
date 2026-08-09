# 三平台桌面發佈與推廣片（2026-08-10）

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260810-linux-amd64/dq3-remake-20260810-x86_64.AppImage` | 8,526,328 | `4cf021a365ea440f0288af3d2e5cc34ff4b6cbf99e198e66939ae4658621941e` |
| Windows x86_64 | `dist/dq3-20260810-windows-x86_64.zip` | 6,407,190 | `e8c9901c51170228fe116256a3938453ee3b5e804656b953c4bc81b4e9e084ad` |
| macOS Intel | `dist/dq3-20260810-macos-x86_64.zip` | 6,570,252 | `9a0667a0668adc0911fffc2442928b44dc75dfe0adecd00198f4e0e7e0bbca3a` |
| macOS Apple Silicon | `dist/dq3-20260810-macos-arm64.zip` | 6,112,357 | `f6f37f2d123ba7a07aec7bc60799548ff29df4fd63bb4670a6b4f5ec973af870` |

Linux 檔案是使用專案固定的 `work/.tools/runtime-x86_64` 與 `appimagetool` 產出的
Type 2 AppImage，已在無 FUSE 的 Docker 內以 `--appimage-extract` 解包驗證；解包後的
內嵌 ELF 路徑為 `squashfs-root/usr/bin/dq3-remake`，其 SHA-256 與同一目錄保留的
`dq3-remake-linux-amd64` 直接 ELF 相同（13,270,304 bytes，
`7dca25fafa2725f9d7f454d14d88ed4b0113240256f6cdbfcc23b54c01d4bb03`）。macOS ZIP 內含
`Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面外部 game pack 仍可在
不重新編譯的情況下替換，但需通過 schema、reference、content hash 及資產路徑驗證。
本次 macOS 內層執行檔為 Mach-O `x86_64`（12,441,992 bytes，SHA-256
`c6074aa94352130d2f72e63fa22b713c747d62f27b62b22d4bfcc8b08b8786cd`）與 `arm64`
（11,849,682 bytes，SHA-256
`e4a5fcd0a0154c26d1d839baa3f9b7935b5a8b52a20917fa46c38e72fb5cb046`）；Windows PE
執行檔為 12,677,120 bytes，SHA-256
`7a1735364dcd1e93acd3f9d5404d19dcdc7029f7d53998b0e9f9933cc63ca36f`。

## 推廣片

`dist/dq3-promo-20260810.mp4` 由版控的 Ebiten runtime PNG（含本輪修正後的
`dq3_remake_ebitan/docs/ng_confirm.png`）剪輯成 1280×700、30 fps、H.264 的 45.000 秒
無音軌推廣片，SHA-256 為
`e2c784a2142f620fed1d796a3917fad3fe9524d44de2b2a8e9d8538ca50d286f`，大小 4,101,937
bytes。片中依序展示開場創角／家中／王城、地表與戰鬥、日邦格八頭大蛇兩戰、商人／黃金球／
解謎、提頓／勇氣洞窟、不死鳥、下世界與 `THE END`；沒有把診斷用文字或舊 C/SDL 錄製幀
當成目前 Go/Ebitengine 畫面。

## 驗證界線

- `TestOpeningProductionInputTrace`：Docker＋Xvfb，正式 `InputState` 從新遊戲到 `THE END`，
  311.44 秒通過。
- 本輪狀況窗切片：`go test ./internal/...` 全數通過；Docker＋Xvfb 的
  `go test ./game -run "Test(Panel|OpeningParity|CmdMenuDrawHits|Equipment|Field)"` 通過，
  並完成桌面交叉編譯。
- 本輪開場欄位切片：Docker＋Xvfb 的
  `go test ./game -run "Test(NewGame|DumpNewGameScreens)"` 通過，並以同一容器重產
  `dq3_remake_ebitan/docs/ng_confirm.png`；完整長測試仍依使用者指示不重跑。
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
執行檔 SHA-256 為 `7dca25fafa2725f9d7f454d14d88ed4b0113240256f6cdbfcc23b54c01d4bb03`，
與旁置 Linux ELF 相同。所有發佈包雜湊均為本輪重建後重新計算，不沿用舊版數值。
