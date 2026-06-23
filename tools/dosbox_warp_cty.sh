#!/usr/bin/env bash
# A′:在 dq3-dosbox 容器內,跑名字輸入 → 新遊戲(EXE 已 patch [0x256c]=候選CTY)→ 截圖內裝。
# /game = patched 遊戲目錄(含已改 DQ3.EXE);/work = repo。輸出 /work/dosbox/warp_*.png。
set -u
TAG="${1:-cty}"
OUT=/work/dosbox; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 & XV=$!; sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 & DB=$!; sleep 9
WIN=""; fw(){ WIN="$(xdotool search --name DOSBox 2>/dev/null|head -1)"; }
shot(){ import -window root "$OUT/warp_${TAG}_$1.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/warp_${TAG}_$1.png" 2>/dev/null; }; }
k(){ fw; for a in "$@"; do [ -n "$WIN" ] && { xdotool windowfocus "$WIN" 2>/dev/null; xdotool windowactivate "$WIN" 2>/dev/null; }; xdotool key --clearmodifiers "$a"; sleep 0.35; done; }
# 標題→姓名輸入
for n in 1 2 3 4; do k Return; sleep 1.3; done
# 注音→英數(cell43):Down×3 Right×8 space
k Down Down Down Right Right Right Right Right Right Right Right; k space; sleep 0.5
# 字元格 raw0 打1字:Up×3 Left×8 space
k Up Up Up Left Left Left Left Left Left Left Left; k space; sleep 0.4
# 完成(cell44):Down×4 Right×8 space
k Down Down Down Down Right Right Right Right Right Right Right Right; k space; sleep 1.2
# 推進角色建立/開場 → 進遊戲(patch 後 = 候選 CTY 內裝)
for n in $(seq 1 12); do k Return; sleep 1.0; done
shot 10_ingame
for d in Up Up Down Down Left Right; do k "$d"; sleep 0.4; done; shot 11_moved
k Return; sleep 0.8; shot 12_examine
sleep 0.5; kill $DB 2>/dev/null; kill $XV 2>/dev/null
echo "done $TAG"; tail -3 /tmp/dosbox.log
