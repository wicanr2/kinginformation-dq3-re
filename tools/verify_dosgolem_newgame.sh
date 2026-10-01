#!/usr/bin/env bash
# 主機僅做 Docker／Git 控制；探測、建置、測試與輸出全部在容器內。
# 用法：bash tools/verify_dosgolem_newgame.sh [dosgolem 來源目錄] [--prototype]
# 正式比較目前會回報 Issue #4 已知差異；--prototype 僅驗證 DRAFT。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${1:-/home/anr2/cht/dosgolem}"
MODE="${2:-production}"
case "$MODE" in production|--prototype) ;; *) echo '模式須為 production 或 --prototype' >&2; exit 1;; esac
for path in "$ROOT" "$SOURCE" "$ROOT/assets_raw" "$ROOT/work" "$ROOT/work/dosgolem-opening" "$ROOT/work/.gocache-test" "$ROOT/work/.gopath-test"; do
  test -d "$path" || { echo "目錄不存在：$path" >&2; exit 1; }
done
test -f "$ROOT/assets_raw/DQ3.EXE"
test -f "$SOURCE/internal/dos/files.go"
REVISION="$(git -C "$SOURCE" rev-parse HEAD)"
test "$REVISION" = 2f44a68ebfc54b28fb15dd4a34510b0b04a5415d || { echo 'dosgolem 版本須重新審查' >&2; exit 1; }
test -z "$(git -C "$SOURCE" status --porcelain --untracked-files=no)" || { echo 'dosgolem 已有未提交修改' >&2; exit 1; }
NAME="dq3-dosgolem-newgame-$$"
cleanup() { docker rm -f "$NAME" >/dev/null 2>&1 || true; }
trap cleanup EXIT
timeout 240s docker run --rm --name "$NAME" --network none \
  --memory 4g --cpus 2 --pids-limit 192 -u "$(id -u):$(id -g)" \
  -v "$ROOT:/repo:ro" -v "$ROOT/work:/work" -v "$SOURCE:/dosgolem:ro" \
  -w /repo/dq3_remake_ebitan -e DQ3_NEWGAME_VERIFY_MODE="$MODE" \
  -e GOCACHE=/work/.gocache-test -e GOPATH=/work/.gopath-test \
  -e GOMAXPROCS=2 -e GOMEMLIMIT=2GiB -e GOPROXY=off -e GOSUMDB=off \
  -e DQ3_ASSETS=/repo/assets_raw -e DQ3_DOSGOLEM_NEWGAME_DIR=/work/dosgolem-opening \
  dq3-ebiten-test:20260822-r1 bash -c '
    set -euo pipefail
    for target in /work "$GOCACHE" "$GOPATH" "$DQ3_DOSGOLEM_NEWGAME_DIR"; do
      test "$(stat -c %u "$target")" = "$(id -u)"
      test "$(stat -c %g "$target")" = "$(id -g)"
    done
    python3 /repo/tools/dosgolem_newgame_probe.py
    go test -p 2 -c -o /tmp/dq3-game.test ./game
    Xvfb :88 -screen 0 640x350x24 -nolisten tcp >/tmp/xvfb.log 2>&1 & xvfb_pid=$!
    trap '\''kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true'\'' EXIT
    export DISPLAY=:88
    for n in $(seq 1 50); do
      test -S /tmp/.X11-unix/X88 && break
      kill -0 "$xvfb_pid"
      sleep 0.1
    done
    test -S /tmp/.X11-unix/X88
    if test "$DQ3_NEWGAME_VERIFY_MODE" = --prototype; then
      export DQ3_DOSGOLEM_WINDOW_PROTOTYPE=1
      comparison=TestDosgolemNewGameWindowPrototype
    else
      comparison=TestDosgolemNewGameMenuAndNameComparison
    fi
    timeout 30s /tmp/dq3-game.test -test.v -test.run "^${comparison}$"
  '
