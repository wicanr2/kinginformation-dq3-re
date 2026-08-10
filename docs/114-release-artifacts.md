# 三平台桌面發佈與推廣片（2026-08-10）

> 本文件前半保留上一輪「不含遊戲檔」包的歷史記錄；本輪同時產出的 patch／本機完整版與
> 最新 checksum 以 [`docs/122`](122-release-20260810.md) 為準。下方的 311.44 秒主線數字是
> 先前紀錄，不是本輪重新執行的長測試結果。

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260810-linux-amd64/dq3-remake-20260810-x86_64.AppImage` | 7,305,720 | `7bc259a4726fd1b4b576bf252e39f231774d51fbdbccbbb938b2d246027bbca1` |
| Windows x86_64 | `dist/dq3-20260810-windows-x86_64.zip` | 6,449,967 | `120ad746d43335ff44cf33cbf90ddcfa8efcf9bec467a0865de07ba95b5de58b` |
| macOS Intel | `dist/dq3-20260810-macos-x86_64.zip` | 3,195,298 | `299d2447f0be14abd66efeaa3b1179d5fd96334ef14fdb4e78ad3c78398dccd3` |
| macOS Apple Silicon | `dist/dq3-20260810-macos-arm64.zip` | 2,990,159 | `6dc87df1088459cf70b7e50c193ffcf543f174f29f43884bdd787ed933f3ab2d` |

Linux 檔案是使用專案固定的 `work/.tools/runtime-x86_64` 與 `appimagetool` 產出的
Type 2 AppImage，已在無 FUSE 的 Docker 內以 `--appimage-extract` 解包驗證；解包後的
內嵌 ELF 路徑為 `squashfs-root/usr/bin/dq3-remake`，其 SHA-256 與同一目錄保留的
`dq3-remake-linux-amd64` 直接 ELF 相同（13,137,096 bytes，
`a562cdf69a4ec525948b5e2b799305ebc45fc63d9c16c27d7cb1b9a53077dc3f`）。macOS ZIP 內含
`Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面外部 game pack 仍可在
不重新編譯的情況下替換，但需通過 schema、reference、content hash 及資產路徑驗證。
本次 macOS 內層執行檔為 Mach-O `x86_64`（8,425,464 bytes，SHA-256
`ce58c02b42078934849f5163d46e02b92426d36bc7e26966d5ac09935e3326e6`）與 `arm64`
（8,070,770 bytes，SHA-256
`1fcbea2d589d8759b99ae1fa190d5f913ddf2d6823b66d483610e99427ce4889`）；Windows PE
執行檔為 12,711,936 bytes，SHA-256
`31a4e604258ee89535356739dadc12e3f655acda57942a7de7eb509904eaeb76`。Linux 直接 ELF、
Windows PE 與兩個 Mach-O 的架構由 Docker 內 `objdump`／`otool`／`lipo` 核對。

## 推廣片

`dist/dq3-promo-20260810.mp4` 由 Docker 內新版開機過場 runtime dump、版控 Ebiten runtime
PNG（含本輪修正後的 `dq3_remake_ebitan/docs/ng_confirm.png`）剪輯成 1280×700、30 fps、
H.264 的 45.000 秒無音軌推廣片；`ffprobe` 核對 1350 幀，SHA-256 為
`454c3d83013e0be4f52c9e928c58c442532bec082bd348ae684c40f9f6cc3492`，大小 2,004,184 bytes。
片頭先展示五張 `opening` 開機卡，再依序展示創角／家中／王城、地表與戰鬥、日邦格、商人／
黃金球／解謎、不死鳥、下世界與 `THE END`；沒有把診斷用文字或舊 C/SDL 錄製幀當成目前
Go/Ebitengine 畫面。

## 驗證界線

- `TestOpeningProductionInputTrace`：Docker＋Xvfb，正式 `InputState` 從新遊戲到 `THE END`，
  311.44 秒通過。
- 本輪狀況窗切片：`go test ./internal/...` 全數通過；Docker＋Xvfb 的
  `go test ./game -run "Test(Panel|OpeningParity|CmdMenuDrawHits|Equipment|Field)"` 通過，
  並完成桌面交叉編譯。
- 本輪開場欄位／過場切片：Docker＋Xvfb 的
  `go test ./game -run "Test(OpeningCutscene|NewGame|DumpNewGameScreens|DumpOpeningCutscene)"`
  通過；`TestDumpOpeningCutscene` 產出五張 runtime PNG，並以同一容器重產
  `dq3_remake_ebitan/docs/ng_confirm.png`。完整長測試仍依使用者指示不重跑。
- 本輪 CTY 遭遇閘門：`go test ./internal/dq3data -run "Test(OpenTown.*Encounter|Encounter)"`
  與 `go test . -run "Test(Save|Encounter)"` 均在 Docker＋Xvfb 通過。
- 新版 AppImage 使用 `--appimage-extract-and-run` 搭配合法唯讀 `assets_raw` 啟動 smoke，
  Linux ELF 亦在 Docker＋Xvfb 啟動 5 秒後正常受控結束；無 `NewGame`／panic／fatal。
  完整長測試套件依使用者指示不再重跑；這不取代上述正式主線 trace，也不把尚未 V3 的
  戰鬥逐動作 timing、動畫、音效、抗性與掉落宣稱為完成。

## 重建入口

macOS `.app`／ZIP 由 `tools/package_macos_bundle.sh`（`RELEASE_DATE=20260810`）產生；
交叉編譯使用 `u5cht/osxcross:latest` 的 MacOSX SDK、Go 1.24.13，IDA／原版資料未進
產物。Linux Go 執行檔使用 `dq3-ebitan-test:latest`、Go module cache 與 Docker 內
Ebitengine；AppImage 以 `u5cht/appimage:latest`／固定 `work/.tools/appimagetool` 建立。
Windows ZIP 使用明確版本的 `dq3-ebiten-windows:20260810`（由
`dq3_remake_ebitan/Dockerfile.windows` 建立，Go 1.26.5＋Debian bookworm MinGW-w64）；本輪
以目前 Go/Ebitengine 原始碼重新編譯四個桌面執行檔。AppImage 內嵌
執行檔 SHA-256 為 `a562cdf69a4ec525948b5e2b799305ebc45fc63d9c16c27d7cb1b9a53077dc3f`，
與旁置 Linux ELF 相同。所有發佈包雜湊均為本輪重建後重新計算，不沿用舊版數值。
