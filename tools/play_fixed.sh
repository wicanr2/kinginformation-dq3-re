#!/usr/bin/env bash
# 一鍵啟動「精訊 DQ3 修正版(binary patch 對照組)」於 DOSBox。
# 修正:#1 巴拉摩斯打不死 / #2 彩虹水滴 / #3 五頭龍·歐里狄加當機 / #4 勇者MP / #7a 隼劍雙擊。
# 不污染 host:優先用 host 既有 dosbox;沒有則用 docker dq3-dosbox(需 X11)。
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GAME="$ROOT/work/dq3_fixed_game"

if [ ! -f "$GAME/DQ3.EXE" ]; then
  echo "修正版遊戲包不存在。先建立:"
  echo "  bash tools/dockrun.sh tools/build_fixed_version.py"
  exit 1
fi

echo "=== 精訊 DQ3 修正版 ==="
echo "遊戲目錄: $GAME"

if command -v dosbox >/dev/null 2>&1; then
  echo "[host dosbox] 啟動中…(視窗開啟後即可遊玩;存檔在遊戲內進行)"
  exec dosbox -c "mount c \"$GAME\"" -c "c:" -c "dq3"
elif command -v dosbox-x >/dev/null 2>&1; then
  echo "[host dosbox-x] 啟動中…"
  exec dosbox-x -c "mount c \"$GAME\"" -c "c:" -c "dq3"
else
  echo "[docker dq3-dosbox] host 無 dosbox,改用容器(需要 X11 顯示)。"
  command -v xhost >/dev/null 2>&1 && xhost +local:docker >/dev/null 2>&1 || true
  exec docker run --rm \
    -e DISPLAY="${DISPLAY:-:0}" \
    -v /tmp/.X11-unix:/tmp/.X11-unix \
    -v "$ROOT":/work \
    -v "$GAME":/game \
    dq3-dosbox dosbox -conf /work/tools/dosbox.conf
fi
