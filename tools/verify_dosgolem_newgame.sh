#!/usr/bin/env bash
# 主機僅做 Docker／Git 控制；探測、建置、測試與輸出全部在容器內。
# 用法：bash tools/verify_dosgolem_newgame.sh [dosgolem 來源目錄] [--prototype|--navigation|--creation|--opening|--birthday-pages|--mother-entry-original|--mother-finish-original|--mother-home-original|--mother-home-contract-original|--mother-home-navigation-original|--mother-home-animation-original]
# 預設重生主選單／初始命名並比較正式畫面；--prototype 僅驗證歷史 DRAFT。
# --navigation 重生兩條命名收據，驗證六次方向與四次功能輸入的狀態及完整畫布。
# --creation 重生固定種子創角，驗證命名／性別、能力交易與等待／確認完整畫面。
# --opening 從冷啟動延伸第17次接受角色，驗證同批創角與黑底生日首頁；續頁／母親仍待閉合。
# --birthday-pages 延伸兩次生日續頁及四次捲動；生日文字必須通過，房間未閉合前仍有正式紅測試。
# --mother-entry-original 只重生原版37次正常輸入的母親入口收據，不宣稱remake對拍通過。
# --mother-finish-original 延伸至城門確認、後三步與旗標返回，仍只重生原版。
# --mother-home-original 重生家中人物、選圖原始資料與正常接近，逐項驗證原版收據。
# --mother-home-contract-original 另讀取獨立圖像初始化的自然BIOS時鐘。
# --mother-home-navigation-original 驗證Escape選定當前選項與四方向環繞，仍只重生原版。
# --mother-home-animation-original 核對完整啟動共用動畫計數及NPC取圖，仍只重生原版。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE="${1:-/home/anr2/cht/dosgolem}"
MODE="${2:-production}"
case "$MODE" in production|--prototype|--navigation|--creation|--opening|--birthday-pages|--mother-entry-original|--mother-finish-original|--mother-home-original|--mother-home-contract-original|--mother-home-navigation-original|--mother-home-animation-original) ;; *) echo '未知驗證模式；請依檔首列出的模式選擇' >&2; exit 1;; esac
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
test "$MODE" != --birthday-pages || LIMIT=420
test "$MODE" != --mother-entry-original || LIMIT=720
test "$MODE" != --mother-finish-original || LIMIT=840
test "$MODE" != --mother-home-original || LIMIT=720
test "$MODE" != --mother-home-contract-original || LIMIT=840
test "$MODE" != --mother-home-navigation-original || LIMIT=900
test "$MODE" != --mother-home-animation-original || LIMIT=900
timeout "${LIMIT}s" docker run --rm --name "$NAME" --network none \
  --memory 4g --cpus 2 --pids-limit 192 -u "$(id -u):$(id -g)" \
  -v "$ROOT:/repo:ro" -v "$ROOT/work:/work" -v "$SOURCE:/dosgolem:ro" \
  -w /repo/dq3_remake_ebitan -e DQ3_NEWGAME_VERIFY_MODE="$MODE" \
  -e DQ3_DOSGOLEM_SOURCE_HOST_PATH="$SOURCE" \
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
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-home-animation-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_home_animation python3 /repo/tools/dosgolem_newgame_probe.py
      python3 /repo/tools/verify_dosgolem_npc_animation.py --receipt /work/dosgolem-opening/issue4-home-animation-receipt.json --output /work/issue4-home-animation-source-audit.json
      exit 0
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-home-contract-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_home_contract python3 /repo/tools/dosgolem_newgame_probe.py
      python3 /repo/tools/verify_dosgolem_home_entry.py --receipt /work/dosgolem-opening/issue4-home-contract-receipt.json --output /work/issue4-home-contract-source-audit.json
      exit 0
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-home-navigation-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_home_navigation python3 /repo/tools/dosgolem_newgame_probe.py
      python3 /repo/tools/verify_dosgolem_home_entry.py --receipt /work/dosgolem-opening/issue4-home-navigation-receipt.json --output /work/issue4-home-navigation-source-audit.json
      exit 0
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-home-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_home_entry python3 /repo/tools/dosgolem_newgame_probe.py
      python3 /repo/tools/verify_dosgolem_home_entry.py --output /work/issue4-home-source-evidence-receipt.json
      echo "原版家中人物、選圖及正式接近收據已核對；尚未執行remake比較。"
      exit 0
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-entry-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_approach python3 /repo/tools/dosgolem_newgame_probe.py
      python3 -c "import json; from pathlib import Path; d=json.loads(Path(\"/work/dosgolem-opening/issue4-mother-approach-receipt.json\").read_text()); assert any(\"ida_linear=1010b \" in line for line in d[\"mother_entry_events\"]), \"尚未自然抵達母親入口\""
      echo "原版母親入口收據已重生；尚未執行remake比較。"
      exit 0
    elif test "$DQ3_NEWGAME_VERIFY_MODE" = --mother-finish-original; then
      DQ3_NEWGAME_PROBE_SCENARIO=mother_finish python3 /repo/tools/dosgolem_newgame_probe.py
      python3 -c "import json; from pathlib import Path; d=json.loads(Path(\"/work/dosgolem-opening/issue4-mother-finish-receipt.json\").read_text()); assert len(d[\"player_input\"]) == 38; assert len(d[\"actual_irq1_events\"]) == 76; assert not d[\"game_state_injection\"]; assert len(d[\"artifacts\"]) == len({a[\"path\"] for a in d[\"artifacts\"]}); assert any(\"ida_linear=1020a \" in line and \"player_x=21 player_y=17\" in line for line in d[\"mother_entry_events\"]), \"尚未自然完成母親流程\""
      echo "原版母親完整返回收據已重生；尚未執行remake比較。"
      exit 0
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
