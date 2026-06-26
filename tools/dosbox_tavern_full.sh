#!/usr/bin/env bash
# 完整推進:cutscene→splash→title menu(遊戲開始)→姓名輸入→完成→開場→進城。
# 慢推 Return 過開場,姓名用「英數打1字+功能列完成」機制。每步截圖,有界,無 sentinel。
set -u
OUT=/work/dosbox/tavern; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 & XV=$!; sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 & DB=$!; sleep 13
WIN=""; fw(){ WIN="$(xdotool search --name DOSBox 2>/dev/null|head -1)"; }
shot(){ import -window root "$OUT/f_$1.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/f_$1.png" 2>/dev/null; }; echo "shot $1 $(stat -c %s "$OUT/f_$1.png")"; }
k1(){ fw; [ -n "$WIN" ] && { xdotool windowfocus "$WIN" 2>/dev/null; xdotool windowactivate "$WIN" 2>/dev/null; }; xdotool key --clearmodifiers "$1"; sleep 0.40; }

shot 00_boot
# A) 過 cutscene + splash + title menu:慢推 Return,每步截圖(到姓名 grid)
for n in $(seq 1 9); do k1 Return; sleep 1.5; shot "$(printf 'a%02d' $n)"; done
# B) 姓名:切英數(grid→功能列列1→Enter)
for i in 1 2 3 4 5 6 7 8 9; do k1 Right; done; shot b1_fn_eng
k1 Return; sleep 0.6; shot b2_alnum
# C) 回 grid 打 1 字(Left 回 grid 左上,Enter 選字)
for i in 1 2 3 4 5 6 7 8 9; do k1 Left; done; shot c1_grid
k1 Return; sleep 0.6; shot c2_char
# D) 功能列→完成(列5)
for i in 1 2 3 4 5 6 7 8 9; do k1 Right; done; shot d1_fn
k1 Down; k1 Down; k1 Down; k1 Down; shot d2_finsel
k1 Return; sleep 1.6; shot d3_done
# E) 開場/性別/職業/隊伍/進城:慢推 Return
for n in $(seq 1 16); do k1 Return; sleep 1.2; shot "$(printf 'e%02d' $n)"; done
sleep 0.5; kill $DB 2>/dev/null; kill $XV 2>/dev/null
echo "=== done ==="; ls "$OUT"/f_*.png 2>/dev/null | wc -l
