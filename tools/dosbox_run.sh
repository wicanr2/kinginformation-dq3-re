#!/usr/bin/env bash
# 在 dq3-dosbox 容器內執行:啟動 Xvfb + DOSBox,跑一段按鍵腳本並截圖。
# 由 host 透過 `docker run ... dq3-dosbox bash /work/tools/dosbox_run.sh <SCENARIO>` 呼叫。
# 所有等待皆有上界;無 sentinel 無限輪詢。
set -u
SCENARIO="${1:-title}"
OUT=/work/dosbox
mkdir -p "$OUT/capture"
export DISPLAY=:99

# 啟動虛擬 framebuffer (320x200 放大後約 640x400,給足 1024x768)
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 &
XVFB_PID=$!
sleep 2

# 啟動 DOSBox (背景);conf 的 autoexec 會 mount /game 並跑 dq3
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 &
DB_PID=$!
sleep 6   # 等 DOSBox 視窗 + DQ3 載入

WIN=""
find_win() { WIN="$(xdotool search --name DOSBox 2>/dev/null | head -1)"; }
shot() {  # shot <name>
  import -window root "$OUT/${SCENARIO}_$1.png" 2>>/tmp/import.log || \
    { xwd -root -silent | convert xwd:- "$OUT/${SCENARIO}_$1.png" 2>>/tmp/import.log; }
  echo "shot $1"
}
key() {  # key <keysym>
  find_win
  if [ -n "$WIN" ]; then xdotool key --window "$WIN" "$1" 2>>/tmp/xdo.log
  else xdotool key "$1" 2>>/tmp/xdo.log; fi
}

# 通用推進序列
shot 00_boot
sleep 3; shot 01_title
# 多按 Enter 推進標題/開場/對話,每步截圖
i=2
for n in $(seq 2 8); do
  key Return; sleep 1.2; shot "$(printf '%02d' $n)_enter"
done
# 走動 (城鎮/地圖)
for d in Down Down Right Right Up Left; do key "$d"; sleep 0.6; done
shot 20_moved
# 再走一段嘗試觸發世界圖/戰鬥
for n in $(seq 21 30); do
  key Down; sleep 0.5
  case $n in 24|27|30) shot "$(printf '%02d' $n)_walk";; esac
done
shot 31_final

# 收尾:用 DOSBox 內建 capture 也保險 (Ctrl-F5 已被腳本忽略,靠 import)
sleep 1
kill $DB_PID 2>/dev/null
kill $XVFB_PID 2>/dev/null
echo "=== dosbox.log tail ==="
tail -20 /tmp/dosbox.log
echo "=== captures ==="
ls -la "$OUT"/*.png 2>/dev/null
