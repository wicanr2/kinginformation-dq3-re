#!/usr/bin/env bash
# A′ 可行性測試:進遊戲(阿里阿罕)→ 往各方向走一段 → 逐步截圖,找 section 轉場。
set -u
OUT=/work/dosbox; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/x.log 2>&1 & XV=$!; sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/d.log 2>&1 & DB=$!; sleep 9
WIN=""; fw(){ WIN="$(xdotool search --name DOSBox 2>/dev/null|head -1)"; }
k(){ fw; [ -n "$WIN" ] && { xdotool windowfocus "$WIN" 2>/dev/null; xdotool windowactivate "$WIN" 2>/dev/null; }; for a in "$@"; do xdotool key --clearmodifiers "$a"; sleep 0.35; done; }
shot(){ import -window root "$OUT/walk_$1.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/walk_$1.png" 2>/dev/null; }; }
# 名字輸入(同 oracle)
for n in 1 2 3 4; do k Return; sleep 1.3; done
k Down Down Down Right Right Right Right Right Right Right Right; k space; sleep 0.5
k Up Up Up Left Left Left Left Left Left Left Left; k space; sleep 0.4
k Down Down Down Down Right Right Right Right Right Right Right Right; k space; sleep 1.2
for n in $(seq 1 12); do k Return; sleep 1.0; done
shot 00_ingame
# 沿各方向走,逐步截圖找轉場
i=1
for dir in Down Down Down Down Down Down Up Up Up Up Left Left Left Left Right Right Right Right Down Down Down Down; do
  k "$dir"; sleep 0.4
  shot "$(printf '%02d' $i)_${dir}"; i=$((i+1))
done
sleep 0.5; kill $DB 2>/dev/null; kill $XV 2>/dev/null
echo "done; shots:"; ls "$OUT"/walk_*.png 2>/dev/null | wc -l
