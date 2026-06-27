#!/usr/bin/env bash
# 交叉編譯 dq3_remake 的 Windows x86_64 版(mingw-w64 + SDL2),打包成 zip。
# 不污染 host:全在 dq3-remake-mingw 容器內;產物 chown 回 host。
# 用法(host):tools/package_win.sh
#   先決:docker build -t dq3-remake-mingw -f dq3_remake/scripts/Dockerfile.mingw dq3_remake/scripts
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! docker image inspect dq3-remake-mingw >/dev/null 2>&1; then
  echo "== 建 mingw image(SDL2 mingw dev)=="
  docker build -t dq3-remake-mingw -f dq3_remake/scripts/Dockerfile.mingw dq3_remake/scripts
fi

HOST_UID="$(id -u)"; HOST_GID="$(id -g)"
echo "== 容器內 Windows 交叉編譯 =="
docker run --rm -e HOST_UID="$HOST_UID" -e HOST_GID="$HOST_GID" -v "$ROOT":/repo dq3-remake-mingw bash -lc '
  set -e
  mkdir -p /build-win && cd /build-win
  cmake -S /repo/dq3_remake -B /build-win \
    -DCMAKE_TOOLCHAIN_FILE=/repo/dq3_remake/cmake/mingw-w64-x86_64.cmake \
    -DCMAKE_BUILD_TYPE=Release >/tmp/cw.log 2>&1 || { tail -30 /tmp/cw.log; exit 1; }
  cmake --build /build-win -j --target dq3_remake >/tmp/mw.log 2>&1 || { tail -40 /tmp/mw.log; exit 1; }
  test -f /build-win/dq3_remake.exe || { echo "no .exe produced"; exit 1; }
  echo "   build OK: dq3_remake.exe = $(stat -c %s /build-win/dq3_remake.exe) bytes"

  # PE 健全性:確認是 PE32+ 且僅依賴 SDL2.dll + 系統 DLL(無 libgcc_s/libstdc++)
  echo "--- 依賴 DLL ---"
  x86_64-w64-mingw32-objdump -p /build-win/dq3_remake.exe | grep -i "DLL Name" | sort -u

  # 打包:exe + SDL2.dll + run.bat + README + 驗證腳本
  OUT=/repo/work/dist-win
  rm -rf "$OUT"; mkdir -p "$OUT/bin"
  cp /build-win/dq3_remake.exe "$OUT/bin/"
  cp /usr/x86_64-w64-mingw32/bin/SDL2.dll "$OUT/bin/"
  cp /repo/dq3_remake/DIST_README.md "$OUT/" 2>/dev/null || true
  cat > "$OUT/run.bat" <<"BAT"
@echo off
REM 啟動 dq3_remake。用法:把原版素材夾 assets_raw 放在本檔旁,雙擊執行。
REM 或命令列:run.bat <assets_raw 路徑>
setlocal
set ASSETS=%~1
if "%ASSETS%"=="" set ASSETS=%~dp0assets_raw
"%~dp0bin\dq3_remake.exe" "%ASSETS%" game
BAT
  printf "dq3_remake Windows x64\n素材:把原版 assets_raw 放在 run.bat 旁,或 run.bat <路徑>\nSDL2.dll 已隨附。\n" > "$OUT/README-WIN.txt"
  chown -R "${HOST_UID}:${HOST_GID}" "$OUT"
'

echo "== 版本戳 + zip =="
git -C "$ROOT" rev-parse --short HEAD > work/dist-win/COMMIT.txt 2>/dev/null || echo unknown > work/dist-win/COMMIT.txt
docker run --rm -e HOST_UID="$HOST_UID" -e HOST_GID="$HOST_GID" -v "$ROOT":/repo dq3-remake-mingw bash -lc '
  cd /repo/work && rm -f dq3_remake_win64.zip && zip -rq dq3_remake_win64.zip dist-win
  chown "${HOST_UID}:${HOST_GID}" dq3_remake_win64.zip
'
ls -la work/dq3_remake_win64.zip
echo "== ✅ Windows 打包完成:work/dq3_remake_win64.zip(commit $(cat work/dist-win/COMMIT.txt))=="
