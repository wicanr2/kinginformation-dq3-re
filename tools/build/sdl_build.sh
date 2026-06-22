#!/usr/bin/env bash
# 在 docker(SDL2 dev + cmake)內編 re/sdl 並做 headless 解碼驗證。
# 不污染 host:全在容器內;掛載 repo 唯讀程式碼 + assets。
#   tools/build/sdl_build.sh            # build + 用 TITG.P 跑 headless dump
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

docker run --rm \
  -e SDL_VIDEODRIVER=dummy -e SDL_AUDIODRIVER=dummy \
  -v "$ROOT":/repo dq3-sdl bash -lc '
    set -e
    mkdir -p /build && cd /build
    cmake -S /repo/re/sdl -B /build >/tmp/cmake.log 2>&1 || { tail -20 /tmp/cmake.log; exit 1; }
    cmake --build /build -j 2>/tmp/make.log || { tail -40 /tmp/make.log; exit 1; }
    echo "=== build OK: $(ls -la /build/dq3_sdl | awk "{print \$5}") bytes ==="
    # headless 解碼驗證:dump PPM(不開視窗)
    DQ3_DUMP=/tmp/titg.ppm /build/dq3_sdl /repo/assets_raw TITG.P
    ls -la /tmp/titg.ppm
    head -c 20 /tmp/titg.ppm | tr -d "\0"
    echo
    cp /tmp/titg.ppm /repo/tools/build/titg.ppm 2>/dev/null || true
  '
echo "PPM dumped to tools/build/titg.ppm (gitignored 大檔; 僅本機檢視)"
