# dq3_remake 開發 / 建置環境(各 target)

> 一切在 docker 內建、不污染 host(專案紀律)。素材(原版 `assets_raw/`)使用者合法持有、gitignore 不散布。
> 需原版 `DQ3.EXE` 在手才啟動(版權閘)。

## Docker images(建置環境)

| image | 用途 | 建法 |
|---|---|---|
| `dq3-remake` | Linux/SDL2 建置 + 測試 + AppImage(含 `libvorbis-dev` → MT-32 OGG)| `docker build -t dq3-remake dq3_remake/scripts` |
| `dq3-remake-mingw` | Windows x64 交叉編譯(mingw-w64 + SDL2)| `docker build -t dq3-remake-mingw -f dq3_remake/scripts/Dockerfile.mingw dq3_remake/scripts` |
| `munt-smf2wav` | MT-32 音樂 render(需自有 Roland ROM)| `docker build -t munt-smf2wav -f tools/Dockerfile.munt tools/` |

## 各平台建置

### Linux(原生 / AppImage)
```bash
bash dq3_remake/scripts/build.sh          # docker 內 cmake build + 全單測 + headless dump 驗證
bash tools/package_appimage.sh            # → dist/linux/dq3_remake-x86_64.AppImage(bundle SDL2 + libvorbis)
```

### Windows x64(交叉編譯)
```bash
bash tools/package_win.sh                 # mingw 交叉編 → work/dq3_remake_win64.zip(exe + SDL2.dll + [vorbis DLL] + assets)
# wine 驗證:wine work/dist-win/bin/run.bat
```
> MT-32 需 `libvorbisfile-3.dll`/`libvorbis-0.dll`/`libogg-0.dll` 隨包(mingw vorbis);無則 remake 退回 SB FM。

### macOS(GitHub Actions,無 Mac 開發機)
- `.github/workflows/macos.yml` 在 `macos-14`(Apple Silicon)runner 上 `brew install sdl2 libvorbis` + cmake build,
  上傳 `dq3_remake` binary + dylib 當 artifact;本地用 `gh run download` 抓回組 `.app`/`.dmg`(`dylibbundler`)。
- 參考 skill `mac-app-cross-pack`(不用 Mac 也能 ship universal `.app`/`.dmg` 的完整雷區)。

### Android
- 見 `docs/55` + `docs/62`(★ 評估:改 Go/Ebiten 重寫的 Android 難度對比,可能比 SDL2/NDK 路徑好走)。

## 音樂資產(MT-32)
```bash
bash tools/export_music_mt32.sh   # MBG 18 + EBG 6 = 24 軌 → work/mt32/track_NN.ogg(需 work/music/mt32rom/ 有 Roland ROM)
```
打包時把 `work/mt32/*.ogg` 複製到 `<assets>/mt32/`(gitignore,個人包)。

## CI
- `.github/workflows/ci.yml`:Linux + Windows(mingw)matrix,build + `ctest`(單測)+ headless 煙霧測試 + 上傳 artifact。
- `.github/workflows/macos.yml`:macOS build → artifact。

## 測試 / 驗收 gate
```bash
docker run --rm -v "$PWD":/repo -v dq3build:/build dq3-remake bash -lc \
  'cmake -S /repo/dq3_remake -B /build && cmake --build /build -j && bash /repo/tools/game_tester.sh /repo/assets_raw /build/dq3_remake'
# 交付 gate:game_tester 93/93 全綠
```
