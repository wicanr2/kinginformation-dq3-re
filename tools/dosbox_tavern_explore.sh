#!/usr/bin/env bash
# 用已驗證的 name-skip(space 選格,Down×3 Right×8 序列)進遊戲,
# 走出起始屋 → 阿里阿罕鎮 → 找露依達酒場/登錄所。有界,每幾步截圖。
set -u
OUT=/work/dosbox/tavern/explore; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/x.log 2>&1 & sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/db.log 2>&1 & DB=$!; sleep 14
W="$(xdotool search --name DOSBox|head -1)"
xdotool windowfocus "$W" 2>/dev/null; xdotool windowactivate "$W" 2>/dev/null
shot(){ import -window root "$OUT/x_$1.png" 2>/dev/null; echo "shot $1 $(stat -c %s "$OUT/x_$1.png" 2>/dev/null)"; }
k(){ xdotool windowfocus "$W" 2>/dev/null; xdotool key --window "$W" --clearmodifiers "$1"; sleep "${2:-0.35}"; }
rep(){ local key=$1 n=$2; for ((i=0;i<n;i++)); do k "$key" 0.28; done; }

# 開場→姓名(raw0)
for n in $(seq 1 7); do k Return 1.4; done
k Return 1.4; k Return 1.4
# name-skip: TOFN→英數→打字→TOFN→完成 (全 space 選格)
rep Down 3; rep Right 8; k space 0.6   # TOFN
k space 0.6                             # 英數
rep Up 3; rep Left 8; k space 0.6       # 打 '0'
rep Down 3; rep Right 8; k space 0.6    # TOFN
rep Down 4; k space 0.8                 # 完成
# 性別:男性(預設▶) → 推開場對話
k Return 1.0; for n in $(seq 1 14); do k Return 0.8; done
shot 00_house
# 走出起始屋:往下到底牆缺口(門),出鎮
rep Down 8; shot 01_dn8
rep Down 6; shot 02_dn14
k Return 0.6; shot 03_afterdoor
# 出鎮後在城鎮戶外:四處走,找建築門。每方向走截圖
rep Down 6; shot 04_town_dn
rep Right 8; shot 05_town_rt
rep Up 8; shot 06_town_up
rep Left 8; shot 07_town_lf
rep Up 8; shot 08_town_up2
rep Right 12; shot 09_town_rt2
rep Down 10; shot 10_town_dn2
rep Left 12; shot 11_town_lf2
sleep 0.4; kill $DB 2>/dev/null
echo done
