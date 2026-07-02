#!/usr/bin/env bash
# Android 綁定:把 game/ 套件經 mobile/ 綁成 .aar(給 Android Studio 引用打包 APK/AAB)。
#
# 前置(本機需 Android SDK + NDK,不在 docker 內做,因需 NDK toolchain):
#   1. 安裝 Android SDK + NDK,設環境:
#        export ANDROID_HOME=$HOME/Android/Sdk
#        export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<version>
#   2. 安裝 ebitenmobile:
#        go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
#   3. 放入版權素材(gitignore 不入庫):
#        cp <assets_raw>/*  mobile/assets/
#        mkdir -p mobile/assets/mt32 && cp <work/mt32>/track_*.ogg mobile/assets/mt32/
#
# 產出 dq3.aar 後:在 Android Studio 建 app、引用 dq3.aar、放一個 EbitenView(GLSurfaceView)
# 的 Activity,即可打包 APK/AAB。觸控 UI(input.go/touch.go)在 Go 端已內建。
set -euo pipefail

OUT="${1:-dq3.aar}"
PKG="com.wicanr2.dq3"

command -v ebitenmobile >/dev/null || {
  echo "找不到 ebitenmobile — 先跑:go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest" >&2
  exit 1
}
[ -e mobile/assets/DQ3.PAL ] || {
  echo "mobile/assets/ 尚未放入原版素材(見 mobile/assets/PLACE_ASSETS_HERE.txt)" >&2
  exit 1
}

echo "ebitenmobile bind → $OUT(package $PKG)…"
ebitenmobile bind -target android -javapkg "$PKG" -o "$OUT" ./mobile
echo "完成:$OUT — 匯入 Android Studio 引用後打包 APK/AAB。"
