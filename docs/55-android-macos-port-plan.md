# Android 與 macOS 移植規劃

> 狀態:規劃(尚未實作)。目標把現有 `dq3_remake/`(C99 + SDL2)從 Linux/Windows 再延伸到
> **macOS** 與 **Android**。本文用第一性原理盤點「已具備什麼、缺什麼、怎麼補」,分階段、附風險與工作量。

## 0. 可攜性基準(現況)

重製從一開始就把可攜性當設計目標,移植不是從零開始:

- **語言/依賴**:純 C99 + 單一外部依賴 **SDL2**;無平台專屬 API 散落各處,DOS runtime 語意全收斂在
  `dq3_runtime.c`(SDL2 shim:window/renderer/texture、輸入 scancode、計時)。
- **進入點**:已用 `SDL_MAIN_HANDLED` + `SDL_SetMainReady()` 自管 `main()`(為 Windows 打包做的),
  跨平台 entry 一致 —— 但 **Android 例外**(見 §2.1)。
- **建置**:CMake 已宣告 `Linux / Windows(MinGW/MSVC)/ macOS` 目標;`find_package(SDL2)` + `SDL2::SDL2`。
- **檔案存取**:素材讀取集中在 `dq3_load_file()`(單一 choke point),另有少數散落 `fopen`
  (`dq3_config` 設定、`dq3_save` 存檔、`dq3_combat` 讀 ITEM.DAT)。共 13 個 .c 直接用 `fopen`。
  → **macOS** 沿用 POSIX `fopen` 即可;**Android** 因素材在 APK 內(非檔案系統路徑)需處理(見 §2.2)。
- **輸入**:目前是鍵盤 scancode(方向 + Enter/Space/ESC + C/B/T/U)。**桌面平台直接可用**;
  **Android 無實體鍵盤**,需虛擬方向鍵 + 按鈕(見 §2.3)。
- **畫面**:邏輯解析 640×350 indexed → SDL streaming texture → renderer 放大。SDL 處理縮放,各平台一致。

---

## 1. macOS 移植(優先;工作量最小)

SDL2 原生支援 macOS(Cocoa 後端),程式碼幾乎不用改;工作集中在**打包成 `.app` bundle** 與**簽章/隔離**。

### 1.1 建置
- 編譯器 clang;SDL2 取自 Homebrew(`brew install sdl2`)或官方 `SDL2.framework`。
- **Universal binary(Intel + Apple Silicon)**:兩種作法 ——
  - 一次編兩架構:`-DCMAKE_OSX_ARCHITECTURES="x86_64;arm64"`(SDL2 須為 universal,官方 framework 是)。
  - 或各編一份再 `lipo -create`。
- CMakeLists 已相容;只需在 macOS 上 configure。

### 1.2 `.app` bundle 結構
```
Dragon Fighter III.app/Contents/
  MacOS/dq3_remake            # 主執行檔
  Resources/                  # 圖示 dq3.icns;assets_raw(full 版)或留空(slim 版,使用者自備)
  Frameworks/SDL2.framework   # 內嵌 SDL2,用 @rpath/install_name_tool 指向 bundle
  Info.plist                  # CFBundleExecutable / Icon / 最低系統版本(macOS 11.0)
```
- SDL2 連結方式:內嵌 `SDL2.framework` 到 `Contents/Frameworks/`,`install_name_tool -change` +
  `-rpath @executable_path/../Frameworks` 讓 .app 自包含(不要求使用者裝 SDL2)。
- 素材路徑:`dq3_remake` 接 `argv[1]`;wrapper 用 `Contents/Resources/assets_raw`(full)或
  從 .app 旁/`~/Documents` 找(slim)。存檔/設定寫 `~/Library/Application Support/dq3_remake/`(避免寫進唯讀 bundle)。

### 1.3 簽章 / Gatekeeper
未簽章 .app 預設被擋。兩條路(對齊作者 IJ 專案 `package_macos_local.sh` 的慣例):
- **不簽章 + 文件化解隔離**:`使用說明.txt` 教使用者右鍵「打開」或 `xattr -dr com.apple.quarantine <app>`。
- **正式簽章 + 公證**:需 Apple Developer 帳號(codesign + notarytool + staple)。對個人保存非必要。

### 1.4 此環境限制 / CI
本機是 Linux,**無 macOS 工具鏈**。比照 IJ 專案:用 **GitHub Actions `macos-latest` runner** 建 universal `.app`
(artifact),再用本機腳本(仿 `package_macos_local.sh`)折入素材 → `.app` + `使用說明.txt` + `.tar.gz`
(用 tar 不用 zip,保留 bundle 權限/symlink)。

### 1.5 工作量 / 風險
- 工作量:**小**(1–2 天,主要在 .app 打包腳本 + CI workflow + Info.plist/icns)。
- 風險:低。最大坑是 SDL2 framework 的 rpath 重寫與簽章;均有成熟解法。

---

## 2. Android 移植(工作量較大)

SDL2 官方支援 Android,但行動平台與桌面差異大(無鍵盤、素材在 APK、生命週期),需要真正的移植工。

### 2.1 架構:SDLActivity + NDK
SDL2 Android 範本:Java `SDLActivity` 啟動 → JNI 載入原生 `libmain.so` → 呼叫 `SDL_main`。
- 把 `dq3_remake` 編成 **shared lib `libmain.so`**(NDK + CMake `externalNativeBuild`,可重用現有 CMakeLists)。
- ⚠ **進入點衝突**:桌面用 `SDL_MAIN_HANDLED` 自管 main;**Android 需讓 SDLActivity 呼叫 `SDL_main`**。
  解法:`dq3_runtime.c` 的 `SDL_MAIN_HANDLED` 與 `main` 改成 `#ifndef __ANDROID__` 守衛
  (Android 下 `main` 命名為 `SDL_main`,由 SDL 的 `SDL_main.h` 巨集處理)。**這是必改的第一處。**
- ABI:`arm64-v8a`(主)、`armeabi-v7a`(舊機,選配)、`x86_64`(模擬器)。

### 2.2 素材存取(最關鍵的移植工)
桌面 `fopen("assets_raw/XXX")` 在 Android 行不通 —— 素材在 APK 內、非檔案系統路徑。兩條路:
- **(建議)首次啟動解壓**:把素材放 APK `assets/`,首啟用 SDL_RWops 從 AAsset 複製到
  `SDL_AndroidGetInternalStoragePath()`,之後 `dq3_load_file`/`fopen` 全部接上「內部儲存路徑前綴」即可
  —— **改動小**(集中在 `dq3_load_file` + 給 fopen 路徑加 base dir;13 個 fopen 點多數讀同一批檔)。
- (純淨)全面改 SDL_RWops:把 13 處 `fopen` 換成 `SDL_RWFromFile`(Android 自動走 AAsset)。較乾淨但改動面大、
  且存檔/設定仍要寫內部儲存。→ 折衷:**讀**走 RWops、**寫**(存檔/設定/dq3.cfg)走內部儲存路徑。
- **版權**:素材不公開隨 APK 散布。slim Android 版要設計「使用者把合法素材放 `/sdcard/Android/data/<pkg>/files/` 或
  app 內匯入」的流程(比桌面把 assets_raw 放旁邊麻煩,需一個簡單匯入 UI 或說明)。

### 2.3 觸控輸入(必做的新 UI)
目前是鍵盤 scancode。Android 無實體鍵盤,需**螢幕虛擬控制**:
- 半透明方向鍵(左下)+ A(Enter/確定)/B(取消)/選單(C 鍵=指令窗)按鈕(右下)。
- 實作:SDL touch 事件 → 命中測試對應到既有 scancode(`0x48/0x50/0x4b/0x4d`、`0x1c`、`0x39`、`0x2e`…),
  **餵進既有 `dq3_poll_scancode` 的同一條路徑** → 遊戲邏輯零改動。新增一個 `dq3_touch.c` 疊在 runtime 之上。
- 橫向鎖定(640×350 是寬螢幕);`AndroidManifest.xml` 設 `screenOrientation=landscape`。

### 2.4 生命週期 / 其他
- Android 會暫停/恢復 app(`SDL_APP_WILLENTERBACKGROUND` 等):需在事件中存檔/釋放 texture、恢復時重建。
- 視窗大小由系統決定:沿用現有 renderer 縮放(letterbox)。
- 觸覺/音訊:目前無音訊(SDL audio 未用),Android 不額外處理。

### 2.5 工作量 / 風險
- 工作量:**中–大**(1–2 週)。拆解:進入點守衛(0.5d)、素材解壓/路徑抽象(2–3d)、觸控 UI(3–4d)、
  Gradle/NDK 專案骨架 + 簽章 APK(2d)、生命週期(1d)、實機測試/調觸控手感(數天)。
- 風險:中。主要在觸控手感(RPG 方向操作)與素材匯入流程的 UX;技術上 SDL2 都有支援、無硬阻塞。

---

## 3. 建議順序與「現在可先做」

1. **先 macOS**(CP 值最高):程式碼幾乎不動,產出第三個桌面平台。先加 GitHub Actions macOS workflow + `.app` 打包腳本。
2. **再 Android**:先做**進入點 `#ifndef __ANDROID__` 守衛**與**素材路徑 base-dir 抽象**(這兩項在 Linux 上就能改、
   不破壞現有桌面建置、game_tester 可續綠),把「移植友善度」先墊高;觸控 UI 與 Gradle 專案待有 Android SDK/NDK 環境再上。

> 可立即在本 repo 進行、零回歸的前置工作(不需 macOS/Android 環境):
> - `dq3_runtime.c` 的 `SDL_MAIN_HANDLED`/`main` 加 `#ifndef __ANDROID__` 守衛。
> - 把素材與存檔/設定路徑收斂成「base dir + 相對名」介面(目前已大半集中在 `dq3_load_file`),
>   桌面 base dir = `argv[1]`,未來 Android base dir = 內部儲存。
> - 這些改完桌面行為不變(base dir 仍是 assets_raw),但 Android 接起來時改動面大幅縮小。

## 4. 參考
- 桌面打包既有實作:`tools/package.sh`(Linux tar)、`tools/package_win.sh`(Windows/mingw)、
  `tools/package_appimage.sh`(Linux AppImage)。macOS/Android 打包腳本待新增,風格對齊之。
- 作者 IJ 專案(ScummVM 系)的 `package_macos_local.sh`:.app + 使用說明.txt + tar.gz、quarantine 解法,
  macOS 打包慣例可直接借鏡。
