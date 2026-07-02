#!/usr/bin/env bash
# 在 docker golang 內建 dq3_remake_ebitan(不污染 host)。
# - 純 Go 資料解析(internal/dq3data)go test 對拍真實素材。
# - Ebiten shell compile-check(需 GL/X11/ALSA dev libs;桌面實際跑要有顯示器)。
# 用法:dq3_remake_ebitan/build.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

docker run --rm -v "$ROOT":/repo -w /repo/dq3_remake_ebitan \
  -e DQ3_ASSETS=/repo/assets_raw golang:1.24-bookworm bash -c '
  set -e
  echo "== 資料解析單測(對拍真實素材)=="
  go test ./internal/... -v 2>&1 | tail -15

  echo "== Ebiten 相依(GL/X11/ALSA dev)=="
  apt-get update >/dev/null 2>&1
  apt-get install -y --no-install-recommends \
    libgl1-mesa-dev libx11-dev libxrandr-dev libxcursor-dev libxi-dev \
    libxinerama-dev libxxf86vm-dev libasound2-dev pkg-config >/dev/null 2>&1
  go get github.com/hajimehoshi/ebiten/v2 github.com/hajimehoshi/ebiten/v2/vector 2>&1 | tail -3
  go mod tidy 2>&1 | tail -2

  echo "== Ebiten shell compile-check =="
  go build -buildvcs=false -o /tmp/dq3ebitan . && echo "BUILD OK: $(stat -c %s /tmp/dq3ebitan) bytes"
'
