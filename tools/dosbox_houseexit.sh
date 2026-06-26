#!/usr/bin/env bash
# 從起始屋找門出鎮:密集嘗試走到牆邊各門 tile。每段截圖判斷是否轉場(畫面尺寸/草地)。
set -u
OUT=/work/dosbox/tavern/exit; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/x.log 2>&1 & sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/db.log 2>&1 & DB=$!; sleep 14
W="$(xdotool search --name DOSBox|head -1)"
xdotool windowfocus "$W" 2>/dev/null; xdotool windowactivate "$W" 2>/dev/null
shot(){ import -window root "$OUT/e_$1.png" 2>/dev/null; echo "shot $1 $(stat -c %s "$OUT/e_$1.png" 2>/dev/null)"; }
k(){ xdotool windowfocus "$W" 2>/dev/null; xdotool key --window "$W" --clearmodifiers "$1"; sleep "${2:-0.32}"; }
rep(){ local key=$1 n=$2; for ((i=0;i<n;i++)); do k "$key" 0.26; done; }
# name-skip
for n in $(seq 1 7); do k Return 1.4; done; k Return 1.4; k Return 1.4
rep Down 3; rep Right 8; k space 0.6; k space 0.6
rep Up 3; rep Left 8; k space 0.6
rep Down 3; rep Right 8; k space 0.6; rep Down 4; k space 0.8
k Return 1.0; for n in $(seq 1 14); do k Return 0.8; done
shot 00_house
# 系統掃門:走到最下排,沿著 X 掃每一格嘗試出門(門在下牆某格)
rep Up 10; rep Left 10; shot 01_topleft      # 回左上角定位
# 從左上,逐列往下,每列從左掃到右,踩到門就轉場
for row in 1 2 3 4 5 6 7 8 9 10 11; do
  rep Down 1
  rep Right 16   # 掃整列
  rep Left 16
done
shot 02_swept
# 多按一輪嘗試
rep Down 12; rep Down 12; k Return 0.5; shot 03_force_dn
rep Left 16; k Return 0.5; shot 04_force_lf
rep Up 16; k Return 0.5; shot 05_force_up
rep Right 16; k Return 0.5; shot 06_force_rt
sleep 0.4; kill $DB 2>/dev/null
echo done
