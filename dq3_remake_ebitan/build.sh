#!/usr/bin/env bash
# 在 docker golang 內建 dq3_remake_ebitan(不污染 host)。
# - 純 Go 資料解析(internal/dq3data)go test 對拍真實素材。
# - Ebiten shell compile-check(需 GL/X11/ALSA dev libs;桌面實際跑要有顯示器)。
# 用法:dq3_remake_ebitan/build.sh
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE=dq3-ebiten-test

docker build -q -t "$IMAGE" -f "$ROOT/dq3_remake_ebitan/Dockerfile.test" "$ROOT/dq3_remake_ebitan" >/dev/null
docker run --rm -v "$ROOT":/repo -w /repo/dq3_remake_ebitan \
  -e DQ3_ASSETS=/repo/assets_raw "$IMAGE" bash -c '
  set -euo pipefail
  go mod download

  echo "== 資料解析單測(對拍真實素材)=="
  go test ./internal/... -count=1

  echo "== 遊戲邏輯/真實素材回歸 =="
  xvfb-run -a go test ./game -count=1 -timeout=180s

  echo "== Ebiten shell compile-check =="
  go build -buildvcs=false -o /tmp/dq3ebitan . && echo "BUILD OK: $(stat -c %s /tmp/dq3ebitan) bytes"
'
