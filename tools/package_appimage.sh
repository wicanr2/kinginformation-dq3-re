#!/usr/bin/env bash
# 把 dq3_remake(Linux x86_64)打包成單檔 AppImage(slim:引擎 + 依賴庫,不含版權素材)。
# 作法參考作者另一專案 indiana-jones-...-cht 的 package_appimage.sh,改適配 SDL2 應用。
# 玩家把原版 assets_raw 放在 .AppImage 旁(或當參數傳)即可遊玩。
# 不污染 host:全在 dq3-remake 容器內;appimagetool 用 --appimage-extract-and-run(免 FUSE)。
# 用法(host):tools/package_appimage.sh
#   先決:scratchpad 內有 appimagetool(腳本會自動下載到 work/.tools/)。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"; cd "$ROOT"
TOOLDIR=work/.tools; mkdir -p "$TOOLDIR" dist/linux
APPIMAGETOOL="$TOOLDIR/appimagetool"
if [ ! -x "$APPIMAGETOOL" ]; then
  echo "== 下載 appimagetool =="
  wget -q --timeout=60 -O "$APPIMAGETOOL" \
    https://github.com/AppImage/appimagetool/releases/download/continuous/appimagetool-x86_64.AppImage
  chmod +x "$APPIMAGETOOL"
fi
if [ ! -f "$TOOLDIR/runtime-x86_64" ]; then
  echo "== 下載 AppImage runtime =="
  wget -q --timeout=60 -O "$TOOLDIR/runtime-x86_64" \
    https://github.com/AppImage/type2-runtime/releases/download/continuous/runtime-x86_64
fi

HOST_UID="$(id -u)"; HOST_GID="$(id -g)"
FULL="${DQ3_PACK_FULL:-0}"   # 1 = full 包(內嵌原版 assets_raw + mt32,個人本機驗證用,gitignore 不散布)
docker run --rm -e HOST_UID="$HOST_UID" -e HOST_GID="$HOST_GID" -e FULL="$FULL" \
  -v "$ROOT":/repo -v dq3build:/build dq3-remake bash -lc '
  set -e
  (apt-get update && apt-get install -y --no-install-recommends file) >/dev/null 2>&1 || true
  command -v file >/dev/null || { echo "   缺 file 指令(apt 失敗)"; exit 3; }
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/dev/null 2>&1
  cmake --build /build -j --target dq3_remake >/dev/null 2>&1
  echo "   build OK: $(stat -c %s /build/dq3_remake) bytes"

  APP=/tmp/DQ3Remake.AppDir
  rm -rf "$APP"; mkdir -p "$APP/usr/bin" "$APP/usr/lib"
  cp /build/dq3_remake "$APP/usr/bin/"

  # ldd 複製 SDL2 + 其相依 .so(排除 glibc/核心,靠宿主提供)
  ldd /build/dq3_remake | awk "/=>/{print \$3}" | grep -E "\.so" \
    | grep -viE "/(libc|libm|libpthread|libdl|librt|ld-linux|libstdc\+\+|libgcc_s)\." \
    | while read -r so; do [ -f "$so" ] && cp -Ln "$so" "$APP/usr/lib/" 2>/dev/null || true; done
  echo "   打包 .so:$(ls "$APP/usr/lib" | wc -l) 個"

  # FULL 包:內嵌原版素材 + MT-32 OGG(個人本機驗證,gitignore 不散布)
  if [ "$FULL" = "1" ]; then
    mkdir -p "$APP/usr/share/dq3_remake/assets/mt32"
    cp -a /repo/assets_raw/. "$APP/usr/share/dq3_remake/assets/" 2>/dev/null || true
    [ -d /repo/work/mt32 ] && cp /repo/work/mt32/*.ogg "$APP/usr/share/dq3_remake/assets/mt32/" 2>/dev/null || true
    echo "   FULL:內嵌素材 $(ls "$APP/usr/share/dq3_remake/assets" | wc -l) 檔 + MT-32 $(ls "$APP/usr/share/dq3_remake/assets/mt32"/*.ogg 2>/dev/null | wc -l) 軌"
  fi

  # AppRun:設 LD_LIBRARY_PATH + 定位 assets_raw(內嵌 / 參數 / .AppImage 旁 / CWD)
  cat > "$APP/AppRun" <<"RUN"
#!/bin/bash
HERE="$(dirname "$(readlink -f "$0")")"
export LD_LIBRARY_PATH="$HERE/usr/lib:${LD_LIBRARY_PATH:-}"
BIN="$HERE/usr/bin/dq3_remake"
has_assets(){ [ -f "$1/DQ3CON.MAP" ] && [ -f "$1/ITEM.DAT" ]; }
# 0) FULL 包內嵌素材(最優先;無參數時直接開)
if [ $# -eq 0 ] && has_assets "$HERE/usr/share/dq3_remake/assets"; then
  export DQ3_MT32_DIR="$HERE/usr/share/dq3_remake/assets/mt32"
  exec "$BIN" "$HERE/usr/share/dq3_remake/assets" game
fi
# 1) 明確指定素材夾
if [ $# -ge 1 ] && [ -d "$1" ]; then exec "$BIN" "$@"; fi
# 2) .AppImage 旁 / 當前目錄的 assets_raw
IMGDIR="$(dirname "$(readlink -f "${APPIMAGE:-$0}")")"
for b in "$IMGDIR/assets_raw" "$PWD/assets_raw" "$IMGDIR" "$PWD"; do
  has_assets "$b" && exec "$BIN" "$b" game
done
echo "找不到原版素材(需 DQ3CON.MAP + ITEM.DAT)。" >&2
echo "用法:把 assets_raw 放在 .AppImage 旁,或:  ./dq3_remake-x86_64.AppImage /path/to/assets_raw game" >&2
exit 1
RUN
  chmod +x "$APP/AppRun" "$APP/usr/bin/dq3_remake"

  # .desktop + 自製圖示(避免內嵌版權標題畫面像素)
  cat > "$APP/dq3-remake.desktop" <<DESK
[Desktop Entry]
Type=Application
Name=Dragon Fighter III (DQ3 Remake)
Comment=精訊 DQ3 反組譯 C/SDL2 跨平台重製
Exec=AppRun
Icon=dq3-remake
Categories=Game;RolePlaying;
Terminal=false
DESK
  if command -v convert >/dev/null 2>&1; then
    convert -size 256x256 xc:"#101830" -fill "#f0c000" -gravity center \
      -pointsize 92 -annotate +0-26 "DF" -pointsize 60 -annotate +0+58 "III" \
      "$APP/dq3-remake.png" 2>/dev/null || : > "$APP/dq3-remake.png"
  else : > "$APP/dq3-remake.png"; fi
  cp "$APP/dq3-remake.png" "$APP/.DirIcon" 2>/dev/null || true

  RTFLAG=""; [ -f /repo/work/.tools/runtime-x86_64 ] && RTFLAG="--runtime-file /repo/work/.tools/runtime-x86_64"
  ARCH=x86_64 /repo/work/.tools/appimagetool --appimage-extract-and-run $RTFLAG \
    "$APP" /repo/dist/linux/dq3_remake-x86_64.AppImage 2>&1 | tail -2
  chown -R "${HOST_UID}:${HOST_GID}" /repo/dist/linux
'

# 使用說明.txt(放 .AppImage 旁)
DOC="dist/linux/使用說明.txt"
cat > "$DOC" <<'TXT'
Dragon Fighter III(精訊 DQ3)C/SDL2 重製版 — Linux x86_64 AppImage
==================================================================

這是什麼
--------
精訊資訊 1993 年未發售的 DQ3 中文版,經反組譯還原成 C/SDL2 跨平台重製。
本 AppImage 是 slim 版:含引擎 + 執行所需的函式庫,「不含」原版遊戲素材。

怎麼玩
------
1. 給執行權限(只需一次):
       chmod +x dq3_remake-x86_64.AppImage
2. 把你合法持有的原版素材夾 assets_raw(內含 DQ3CON.MAP、ITEM.DAT、TITG.P…)
   放在這個 .AppImage 旁邊,然後直接執行:
       ./dq3_remake-x86_64.AppImage
   或明確指定素材路徑:
       ./dq3_remake-x86_64.AppImage /path/to/assets_raw game

若提示缺少 FUSE
---------------
改用:  ./dq3_remake-x86_64.AppImage --appimage-extract-and-run

備註:本 AppImage 在較新 glibc 環境建置,於更舊的發行版可能無法執行;
那類系統請改用原始碼自行編譯(docker 內建置,見 repo)。
TXT

ls -lh dist/linux/dq3_remake-x86_64.AppImage 2>/dev/null && echo "== ✅ AppImage -> dist/linux/dq3_remake-x86_64.AppImage ==" || { echo "AppImage build FAILED"; exit 1; }
