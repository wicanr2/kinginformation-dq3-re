#!/usr/bin/env bash
# 主機僅做 Docker／Git 控制；實際探測、建置、測試與輸出全部在容器內。
# 用法：bash tools/verify_dosgolem_opening.sh [dosgolem 原始碼目錄]
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${1:-/home/anr2/cht/dosgolem}"
for path in "$ROOT" "$SOURCE" "$ROOT/assets_raw" "$ROOT/work" "$ROOT/work/.gocache-test" "$ROOT/work/.gopath-test"; do
  test -d "$path" || { echo "目錄不存在：$path" >&2; exit 1; }
done
test -f "$ROOT/assets_raw/DQ3.EXE"
test -f "$SOURCE/internal/dos/files.go"
REVISION="$(git -C "$SOURCE" rev-parse HEAD)"
test "$REVISION" = 2f44a68ebfc54b28fb15dd4a34510b0b04a5415d || { echo 'dosgolem 版本須重新審查' >&2; exit 1; }
test -z "$(git -C "$SOURCE" status --porcelain --untracked-files=no)" || { echo 'dosgolem 已有未提交修改' >&2; exit 1; }
NAME="dq3-dosgolem-opening-$$"
cleanup() { docker rm -f "$NAME" >/dev/null 2>&1 || true; }
trap cleanup EXIT
timeout 240s docker run --rm --name "$NAME" --network none \
  --memory 3g --cpus 2 --pids-limit 192 -u "$(id -u):$(id -g)" \
  -v "$ROOT:/repo:ro" -v "$ROOT/work:/repo/work" -v "$SOURCE:/dosgolem:ro" \
  -w /repo/dq3_remake_ebitan -e DOSGOLEM_REVISION="$REVISION" \
  -e GOCACHE=/repo/work/.gocache-test -e GOPATH=/repo/work/.gopath-test \
  -e GOMAXPROCS=2 -e GOPROXY=off -e DQ3_ASSETS=/repo/assets_raw \
  -e DQ3_DOSGOLEM_OPENING_DIR=/repo/work/dosgolem-opening \
  -e DQ3_DUMP_OPENING=1 -e OPENING_OUT=/repo/work/dosgolem-opening/remake \
  dq3-ebiten-test:20260822-r1 bash -c '
    set -euo pipefail
    test "$(stat -c %u /repo/work)" = "$(id -u)"
    test "$(stat -c %u "$GOCACHE")" = "$(id -u)"
    test "$(stat -c %u "$GOPATH")" = "$(id -u)"
    python3 /repo/tools/dosgolem_opening_probe.py
    if ! go test -p 2 ./internal/gamepack > /repo/work/dosgolem-opening/gamepack.log 2>&1; then
      cat /repo/work/dosgolem-opening/gamepack.log
      exit 1
    fi
    go test -p 2 -c -o /tmp/dq3-game.test ./game
    Xvfb :88 -screen 0 640x350x24 -nolisten tcp >/tmp/xvfb.log 2>&1 & xvfb_pid=$!
    trap '\''kill "$xvfb_pid" 2>/dev/null || true; wait "$xvfb_pid" 2>/dev/null || true'\'' EXIT
    export DISPLAY=:88
    for n in $(seq 1 50); do
      test -S /tmp/.X11-unix/X88 && break
      kill -0 "$xvfb_pid"
      sleep 0.1
    done
    if ! timeout 60s /tmp/dq3-game.test -test.v \
      -test.run "^Test(OpeningCutscene.*|DumpOpeningCutscene|SaveRoundTripHeroNameGender|SaveRecordsAndChecksGamePackIdentity)$" \
      > /repo/work/dosgolem-opening/remake.log 2>&1; then
      cat /repo/work/dosgolem-opening/remake.log
      exit 1
    fi
    cat /repo/work/dosgolem-opening/remake.log
    find /repo/work/dosgolem-opening -user root -o -type d -name "*.md"
  '
