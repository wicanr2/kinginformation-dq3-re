#!/usr/bin/env bash
# 在 docker(SDL2 + cmake)內編 dq3_remake,並做兩種 headless 驗證:
#   (1) dummy video driver:跑載入+解碼,dump PPM(不開視窗)。
#   (2) Xvfb:跑完整視窗 present 路徑,用 import 截真實視窗 PNG。
# 不污染 host:全在容器內;掛載 repo。
#   dq3_remake/scripts/build.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

# 確保 image 存在(不存在則建)
if ! docker image inspect dq3-remake >/dev/null 2>&1; then
  docker build -t dq3-remake "$ROOT/dq3_remake/scripts"
fi

docker run --rm -v "$ROOT":/repo dq3-remake bash -lc '
  set -e
  mkdir -p /build && cd /build
  cmake -S /repo/dq3_remake -B /build -DCMAKE_BUILD_TYPE=Release >/tmp/cmake.log 2>&1 \
    || { tail -30 /tmp/cmake.log; exit 1; }
  cmake --build /build -j >/tmp/make.log 2>&1 || { tail -50 /tmp/make.log; exit 1; }
  echo "=== build OK: $(stat -c %s /build/dq3_remake) bytes ==="

  # (1) dummy:解碼驗證 → PPM
  SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy DQ3_DUMP=/tmp/titg.ppm \
    /build/dq3_remake /repo/assets_raw TITG.P
  pnmtopng /tmp/titg.ppm > /repo/dq3_remake/titg.png 2>/dev/null || \
    cp /tmp/titg.ppm /repo/dq3_remake/titg.ppm
  echo "=== headless dump OK ==="

  # (2) Xvfb:完整視窗路徑 + 截圖(timeout 包住,有界,不掛起)
  export DISPLAY=:99
  Xvfb :99 -screen 0 1280x720x24 >/tmp/xvfb.log 2>&1 &
  XPID=$!
  sleep 1
  ( timeout 6 /build/dq3_remake /repo/assets_raw TITG.P >/tmp/run.log 2>&1 ) &
  RPID=$!
  sleep 3
  import -window root /repo/dq3_remake/titg_window.png 2>/tmp/import.log || \
    echo "(import 截圖略過: $(cat /tmp/import.log))"
  kill $RPID 2>/dev/null || true
  kill $XPID 2>/dev/null || true
  echo "=== Xvfb window screenshot done ==="
'
echo "輸出: dq3_remake/titg.png(解碼)、dq3_remake/titg_window.png(視窗截圖) — gitignored"
