# 三平台桌面發佈與推廣片（2026-08-10）

本輪依需求只封裝三種桌面格式：Linux AppImage、Windows ZIP、macOS ZIP。Android 與
WASM 不列入本輪發佈目標，也沒有把未完成的行動版／瀏覽器版當成支援宣稱。

所有建置、封裝、媒體檢查均在一次性 Docker 容器內完成；原版素材仍由使用者以
`DQ3_ASSETS` 或 `assets_raw/` 提供，沒有被複製進公開發佈包。受保護的
`android/app/libs/`、`tmp_dump.go` 與 scratch／IDA 檔案沒有加入 Git。

## 發佈檔

| 平台 | 檔案 | 大小（bytes） | SHA-256 |
| --- | --- | ---: | --- |
| Linux x86_64 | `dist/dq3-20260809-linux-amd64/dq3-remake-20260809-x86_64.AppImage` | 6,853,824 | `349f51ad5a8ec44a9ac2aea06a180c9ce823b4a2f91a64851857a4745347425d` |
| Windows x86_64 | `dist/dq3-20260809-windows-x86_64.zip` | 6,520,162 | `30a8009780a853fda3ea2f3ac832e265349377e30a963bed3e0ba2e26177b8a3` |
| macOS Intel | `dist/dq3-20260809-macos-x86_64.zip` | 3,176,816 | `cf2f812d4388187e00fd3abea8fd5119b8f2f95526ce85eaf2b06ca5e8500e9f` |
| macOS Apple Silicon | `dist/dq3-20260809-macos-arm64.zip` | 2,972,806 | `856dc0be1ec524a21337c0fe373c251bc339490e12530d1165f919b4122ec9fa` |

macOS ZIP 內含 `Dragon Quest III.app/Contents/Info.plist` 與對應架構的可執行檔。桌面
外部 game pack 仍可在不重新編譯的情況下替換，但需通過 schema、reference、content hash
及資產路徑驗證。

## 推廣片

`dist/dq3-promo-20260810.mp4` 是由正式 renderer 產出的 runtime PNG 剪輯成 1280×700、
30 fps、H.264 的 41.97 秒無外加字幕推廣片，SHA-256 為
`75aef3ca0b066e529221e29f62396f187765e5cb00130a3a26ef7f02a15f5e2e`，大小 2,105,569
bytes。片中畫面包含標題、創角、開場、地表、八頭大蛇、巴哈拉達、怪力魔、索瑪與
`THE END`；戰鬥圖採 `combat_info=0`，沒有把 remake 診斷用的敵人 HP 飄字當成原版 UI。

## 驗證界線

- `TestOpeningProductionInputTrace`：Docker＋Xvfb，正式 `InputState` 從新遊戲到 `THE END`，
  311.44 秒通過。
- 其餘 `game` 測試（排除上述長 campaign trace）：Docker＋Xvfb，29.797 秒通過。
- 完整長測試套件依使用者指示不再重跑；這不取代上述正式主線 trace，也不把尚未 V3 的
  戰鬥逐動作 timing、動畫、音效、抗性與掉落宣稱為完成。

## 重建入口

macOS `.app`／ZIP 由 `tools/package_macos_bundle.sh` 產生；交叉編譯使用
`u5cht/osxcross:latest` 的 MacOSX 15.5 SDK、Go 1.24.13、IDA／原版資料未進產物。
Linux AppImage 與 Windows ZIP 使用既有專案 Docker image 建置；再次產出後應重新記錄
檔案大小與 SHA-256，不可沿用舊雜湊。
