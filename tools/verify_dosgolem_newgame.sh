#!/usr/bin/env bash
# 主機僅做 Docker／Git 控制；探測、建置、測試與輸出全部在容器內。
# 用法：bash tools/verify_dosgolem_newgame.sh [dosgolem 來源目錄] [--prototype|--navigation|--creation|--opening|--birthday-pages]
# 預設重生主選單／初始命名並比較正式畫面；--prototype 僅驗證歷史 DRAFT。
# --navigation 重生兩條命名收據，驗證六次方向與四次功能輸入的狀態及完整畫布。
# --creation 重生固定種子創角，驗證命名／性別、能力交易與等待／確認完整畫面。
# --opening 從冷啟動延伸第17次接受角色，驗證同批創角與黑底生日首頁；續頁／母親仍待閉合。
# --birthday-pages 延伸兩次生日續頁及四次捲動；生日文字必須通過，房間未閉合前仍有正式紅測試。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${1:-/home/anr2/cht/dosgolem}"
MODE="${2:-production}"
case "$MODE" in production|--prototype|--navigation|--creation|--opening|--birthday-pages) ;; *) echo '模式須為 production、--prototype、--navigation、--creation、--opening 或 --birthday-pages' >&2; exit 1;; esac
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
LIMIT=240
test "$MODE" != --navigation || LIMIT=480
test "$MODE" != --birthday-pages || LIMIT=300
timeout "${LIMIT}s" docker run --rm --name "$NAME" --network none \
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
    if test "$DQ3_NEWGAME_VERIFY_MODE" = --navigation; then
      for scenario in name_navigation name_function_mode; do
        DQ3_NEWGAME_PROBE_SCENARIO="$scenario" python3 /repo/tools/dosgolem_newgame_probe.py
      done
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --creation; then
      DQ3_NEWGAME_PROBE_SCENARIO=name_creation python3 /repo/tools/dosgolem_newgame_probe.py
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --opening; then
      DQ3_NEWGAME_PROBE_SCENARIO=opening_accept python3 /repo/tools/dosgolem_newgame_probe.py
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --birthday-pages; then
      DQ3_NEWGAME_PROBE_SCENARIO=birthday_continue python3 /repo/tools/dosgolem_newgame_probe.py
    else
      python3 /repo/tools/dosgolem_newgame_probe.py
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
    test -S /tmp/.X11-unix/X88
    if test "$DQ3_NEWGAME_VERIFY_MODE" = --prototype; then
      export DQ3_DOSGOLEM_WINDOW_PROTOTYPE=1
      comparison=TestDosgolemNewGameWindowPrototype
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --navigation; then
      comparison=TestDosgolemNameInputNavigationComparison
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --creation; then
      export DQ3_DOSGOLEM_CREATION_COMPARE=1
      comparison=TestDosgolemNewGameCreationComparison
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --opening; then
      export DQ3_DOSGOLEM_CREATION_COMPARE=1 DQ3_DOSGOLEM_OPENING_COMPARE=1
      comparison="TestDosgolem(NewGameCreation|OpeningAcceptance)Comparison"
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --birthday-pages; then
      export DQ3_DOSGOLEM_CREATION_COMPARE=1 DQ3_DOSGOLEM_BIRTHDAY_COMPARE=1 DQ3_DOSGOLEM_RETAINED_COMPARE=1
      comparison="TestDosgolem(NewGameCreation|OpeningAcceptance|OpeningRetainedRows)Comparison"
    else
      comparison=TestDosgolemNewGameMenuAndNameComparison
    fi
    timeout 30s /tmp/dq3-game.test -test.v -test.run "^${comparison}$"
  '
