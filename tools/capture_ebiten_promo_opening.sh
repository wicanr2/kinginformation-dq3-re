#!/usr/bin/env bash
# 在 Docker／Xvfb 中錄下正式 v0.1.34 AppRun 的開場動態；不在主機啟動遊戲。
set -euo pipefail

if [ ! -f /.dockerenv ]; then
  echo "錯誤：動態錄影只能在 Docker 容器內執行。" >&2
  exit 2
fi

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

APP="${DQ3_PROMO_APP:-dist/v0.1.34/full/linux-amd64/AppDir/AppRun}"
OUT="${DQ3_PROMO_CAPTURE_OUT:-dist-all/v0.1.34/promo/source/opening_runtime.mp4}"
DISPLAY_ID="${DQ3_PROMO_DISPLAY:-:97}"
LOG="${DQ3_PROMO_CAPTURE_LOG:-/tmp/dq3-promo-game.log}"

for command_name in Xvfb ffmpeg ffprobe xdotool; do
  command -v "$command_name" >/dev/null 2>&1 || {
    echo "錯誤：容器缺少 $command_name。" >&2
    exit 2
  }
done
test -x "$APP" || { echo "錯誤：找不到現行 AppRun $APP。" >&2; exit 2; }

mkdir -p "$(dirname "$OUT")"
runtime_home="$(mktemp -d /tmp/dq3-promo-home-XXXXXX)"
XVFB_PID=""
GAME_PID=""
FFMPEG_PID=""
cleanup() {
  [ -z "$FFMPEG_PID" ] || kill "$FFMPEG_PID" 2>/dev/null || true
  [ -z "$GAME_PID" ] || kill "$GAME_PID" 2>/dev/null || true
  [ -z "$XVFB_PID" ] || kill "$XVFB_PID" 2>/dev/null || true
  rm -rf "$runtime_home"
}
trap cleanup EXIT INT TERM

Xvfb "$DISPLAY_ID" -screen 0 1280x720x24 -nolisten tcp -noreset >"$runtime_home/xvfb.log" 2>&1 &
XVFB_PID=$!
export DISPLAY="$DISPLAY_ID"
export HOME="$runtime_home"
export XDG_CONFIG_HOME="$runtime_home/config"
export XDG_DATA_HOME="$runtime_home/data"
export XDG_CACHE_HOME="$runtime_home/cache"
mkdir -p "$XDG_CONFIG_HOME" "$XDG_DATA_HOME" "$XDG_CACHE_HOME"

for _ in $(seq 1 50); do
  xdotool getmouselocation >/dev/null 2>&1 && break
  sleep 0.1
done
xdotool getmouselocation >/dev/null 2>&1 || { echo "錯誤：Xvfb 未就緒。" >&2; exit 1; }

"$APP" >"$LOG" 2>&1 &
GAME_PID=$!
window_id="$(xdotool search --sync --onlyvisible --name 'Dragon Fighter III' | head -1)"
test -n "$window_id" || { echo "錯誤：找不到 Ebitengine 視窗。" >&2; exit 1; }

ffmpeg -y -loglevel error -f x11grab -draw_mouse 0 -framerate 30 -video_size 1280x720 \
  -i "${DISPLAY}+0,0" -t 12 -an -c:v libx264 -preset veryfast -crf 18 \
  -pix_fmt yuv420p -movflags +faststart "$OUT" &
FFMPEG_PID=$!

# 先讓原版職業巡禮動起來，再以正式鍵盤輸入進標題選單與創角入口。
sleep 5
xdotool key --window "$window_id" Return
sleep 2
xdotool key --window "$window_id" Return

wait "$FFMPEG_PID"
FFMPEG_PID=""
test -s "$OUT" || { echo "錯誤：未產生動態錄影。" >&2; exit 1; }

unique_frames="$(ffmpeg -v error -i "$OUT" -vf fps=1 -f framemd5 - 2>/dev/null \
  | awk -F', ' '!/^#/ { print $6 }' | sort -u | wc -l)"
[ "$unique_frames" -ge 3 ] || {
  echo "錯誤：動態錄影只有 $unique_frames 個不同的每秒取樣畫格。" >&2
  exit 1
}
ffmpeg -v error -i "$OUT" -map 0:v:0 -f null -
echo "通過：現行 Ebitengine 開場動態錄影，12 秒，抽樣不同畫格 $unique_frames。"
