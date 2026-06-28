# dq3_remake 打包 / 發行 WORKLIST

> 把 `dq3_remake`(C99 + SDL2)打包成各平台可直接執行的完整遊戲。
> 移植層的第一性原理分析見 [`../docs/55-android-macos-port-plan.md`](../docs/55-android-macos-port-plan.md);
> 本檔是**可執行 checklist**(打包面),逐項做、做完打勾。
>
> **性質**:個人 / 研究用完整包(素材 = 使用者合法持有的遊戲資料,gitignore 不入庫)。
> **版權閘**:remake 需原版 `DQ3.EXE` 在手才啟動 → 各包是否內含 `DQ3.EXE` 由使用者自決(個人包可含)。

## 共通前置(所有平台都要對齊)

- [ ] **素材清單固定**:`assets/`(原版檔:DQ3.EXE、MBG.MCX、FVOC/NVOC.VCX、D3TXT*、BLK、地圖…)
      + `assets/mt32/track_NN.ogg`(MT-32 音源,18 軌)。列一份「完整包必備檔」清單。
- [ ] **音訊相依分平台處理**:
      - SB FM(OPL2)= 純 C,自包含,零依賴。
      - **MT-32 = 播 OGG → 需 libvorbis**(`vorbisfile`/`vorbis`/`ogg`)。各平台要**打包對應動態庫**,
        否則 CMake 偵測不到 → 退回 WAV(需另放 `track_NN.wav`,較大)或退回 SB。決策:桌面包 libvorbis;Android 另議(見下)。
- [ ] **資產讀取路徑**:`dq3_load_file()` 是 choke point;確認各平台素材夾相對 exe 的路徑一致
      (或用 `DQ3_ASSETS` / 啟動參數)。散落的 `fopen`(config/save/ITEM.DAT)用可寫路徑(見各平台)。
- [ ] **存檔 / 設定可寫路徑**:`dq3.cfg`、`dq3_save*.dat` 不能寫進唯讀的 app bundle / AppImage → 導到各平台使用者資料夾。

## 1. Android 版本規劃

> 深度分析見 docs/55 §2;此處列打包工作。最重的三塊:APK 素材讀取、觸控、音訊庫交叉編譯。

- [ ] **專案骨架**:SDL2 官方 Android 專案模板(`android-project/`)+ Gradle + NDK;`dq3_remake` 原始碼接 `app/jni/src`。
- [ ] **進入點**:Android 走 SDL 的 `SDLActivity` → `SDL_main`;確認 `SDL_MAIN_HANDLED` 在 Android 的處理(docs/55 §2.1)。
- [ ] **素材在 APK 內**:素材打進 `app/src/main/assets/`;`dq3_load_file()` 改用 `SDL_RWFromFile`(SDL 在 Android 自動走 AAssetManager)。
      全面把直接 `fopen`(13 個 .c)的**讀取**改走 SDL_RWops;**寫入**(save/config)導到 `SDL_AndroidGetInternalStoragePath()`。
- [ ] **觸控 UI**:虛擬方向鍵 + A/B + 選單鍵(對映現有 scancode);F1/F5/F9/F10/M 改成畫面按鈕或長按選單(docs/55 §2.3)。
- [ ] **音訊**:SDL audio 在 Android OK;**libvorbis 需交叉編譯**(arm64-v8a / armeabi-v7a / x86_64 三 ABI),
      或評估改用單檔 `stb_vorbis`(免動態庫、跨 ABI 簡單)當 OGG 解碼。決定後接 `DQ3_HAVE_VORBIS` 路徑。
- [ ] **版權閘**:`DQ3.EXE` 放 APK assets(個人包);指紋驗證照舊。
- [ ] **輸出**:`./gradlew assembleRelease` → 簽章 APK / AAB;列最低 API level + 測試裝置。
- [ ] **驗收**:實機跑地表/戰鬥/音樂(SB + 若有 MT-32);存讀檔在 internal storage OK。

## 2. Windows(含全部 DLL)

> 已有 `tools/package_win.sh` + `dq3-remake-mingw` image,延伸即可。

- [ ] **交叉編譯**:用 `dq3-remake-mingw`(MinGW-w64)build;`SDL_MAIN_HANDLED` 已處理 WinMain。
- [ ] **打包全部 DLL**(關鍵:不可漏):
      - `SDL2.dll`
      - **OGG/MT-32**:`libvorbisfile-3.dll`、`libvorbis-0.dll`、`libogg-0.dll`(MinGW build 才有 OGG)
      - MinGW runtime:`libwinpthread-1.dll`、`libgcc_s_*.dll`、`libstdc++-6.dll`(視連結而定;`ldd`/`objdump -p` 查相依)
- [ ] **更新 `tools/package_win.sh`**:把上述 vorbis/ogg DLL 一併複製進包;自動 `objdump -p exe | grep dll` 列相依做檢查。
- [ ] **素材 + mt32/**:`assets/` + `assets/mt32/*.ogg` + (個人包)`DQ3.EXE` 放 exe 同層。
- [ ] **可寫路徑**:存檔/設定導 `%APPDATA%\dq3_remake\`(或 exe 同層,個人包可)。
- [ ] **輸出**:`dq3_remake_win64.zip`(exe + 所有 DLL + assets);選配 NSIS/Inno 安裝檔。
- [ ] **驗收**:乾淨 Windows(無開發環境)解壓即跑;音源能切 MT-32(DLL 齊)。

## 3. macOS(GitHub Action 編譯 → 抓回重打包,含遊戲)

> macOS 不易在 Linux 上 build → 用 GitHub Actions 的 macOS runner 編,artifact 抓回本地組 `.app`。

- [ ] **GitHub Action**:`.github/workflows/macos.yml` —— macos-latest runner,`brew install sdl2 libvorbis`,
      `cmake + make`,上傳 `dq3_remake` binary 當 artifact(universal 可選 `-arch x86_64 -arch arm64`)。
- [ ] **抓回 artifact**:`gh run download` 取編好的 binary。
- [ ] **本地組 `.app` bundle**(含遊戲):
      - `DQ3Remake.app/Contents/MacOS/dq3_remake`
      - `Contents/Resources/assets/`(素材 + mt32/ + 個人包 DQ3.EXE)
      - `Contents/Frameworks/`:SDL2 + libvorbis/vorbisfile/ogg `.dylib`;`install_name_tool` / `dylibbundler` 改 rpath
- [ ] **可寫路徑**:存檔/設定導 `~/Library/Application Support/dq3_remake/`。
- [ ] **輸出**:`.app` → `.dmg`(`hdiutil`);簽章 / notarize 個人用可略(會跳 Gatekeeper 警告,右鍵開啟)。
- [ ] **驗收**:另一台 Mac 雙擊跑;音源 MT-32 OK(dylib bundled)。

## 4. AppImage(完整遊戲,Linux)

- [ ] **AppDir**:`linuxdeploy` 收 binary + SDL2 + libvorbis/vorbisfile/ogg `.so` 進 `AppDir/usr/`。
- [ ] **素材打包**:`assets/`(+ mt32/ + 個人包 DQ3.EXE)放 `AppDir/usr/share/dq3_remake/`;
      AppRun 設 `DQ3_ASSETS` 指向它;存檔/設定導 `$XDG_DATA_HOME/dq3_remake`。
- [ ] **打包**:`linuxdeploy --appdir AppDir --output appimage`(或 `appimagetool`);桌面圖示 + `.desktop`。
- [ ] **輸出**:`DQ3Remake-x86_64.AppImage`,`chmod +x` 即跑。
- [ ] **驗收**:乾淨 Linux(無 SDL/vorbis 安裝)直接跑;MT-32 OK(libs bundled)。

## 5. dev-setup(開發者環境)

- [ ] **文件**:一份 `dq3_remake/DEV-SETUP.md` —— 各 target 怎麼 build。
- [ ] **Docker images**:`dq3-remake`(Linux/SDL + 加 `libvorbis-dev`)、`dq3-remake-mingw`(Windows)、`munt-smf2wav`(MT-32 render)。
      → 評估把 `libvorbis-dev` 烤進 `dq3-remake` image(目前每次 build 臨時 apt,flaky)。
- [ ] **建置指令**:Linux(`cmake -B build && cmake --build build`)、Windows(mingw image)、
      macOS(GH Action)、Android(gradlew)、AppImage(linuxdeploy)各一段。
- [ ] **素材生成**:RE 工具產素材(地圖/sprite/字庫)+ `tools/export_music_mt32.sh`(MT-32 OGG)+ `tools/export_music.sh`(SB)。
- [ ] **測試**:`ctest`/單測執行檔(目前 17 個)在 CI 跑;headless(SDL dummy)煙霧測試。
- [ ] **CI**:GitHub Actions matrix(Linux/Windows/macOS)build + test + 上傳各平台 artifact。

## 跨項決策(先定再動手)

- [ ] **素材與 DQ3.EXE 是否內含**:個人完整包含(方便);若日後對外則改「使用者自備原版檔」模式。
- [ ] **MT-32 OGG 解碼策略統一**:桌面用 libvorbis 動態庫;Android 傾向 `stb_vorbis` 單檔(免三 ABI 交叉編譯庫)。
- [ ] **可寫路徑統一**:抽一個 `dq3_user_dir()`(各平台回對應的 config/save 夾),取代散落 `fopen` 的相對路徑。
