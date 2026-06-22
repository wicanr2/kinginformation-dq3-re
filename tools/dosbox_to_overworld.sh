#!/usr/bin/env bash
# 推過注音姓名輸入 → 進遊戲(地表/城鎮),截主角 sprite。
# 完成機制(docs/15 + 反組譯 ni_alnum/ni_dispatch 0x2087/0x21fe/0x22b2):
#   grid raw=row*9+col(5列9欄),cell=col*5+row;Up/Down=±9,L/R=±1,Space 選格。
#   英數模式 cell43=切焦點功能列、cell44=直接設完成旗標(bit3);名字非空即定案進遊戲。
# 確定性序列:注音起點 raw0 → 切英數 → 打 1 字 → raw44 完成。等待有上界。
set -u
OUT=/work/dosbox; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 & sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 & DB=$!
sleep 6
WIN=""; fw(){ WIN="$(xdotool search --name DOSBox 2>/dev/null|head -1)"; }
shot(){ import -window root "$OUT/ow_$1.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/ow_$1.png" 2>/dev/null; }; echo "shot $1"; }
k(){ fw; for a in "$@"; do if [ -n "$WIN" ]; then xdotool key --window "$WIN" "$a"; else xdotool key "$a"; fi; sleep 0.30; done; }

# 1) 標題 → 姓名輸入(注音模式,游標 raw0)
for n in 1 2 3 4; do k Return; sleep 1.3; done
shot 01_name

# 2) 切英數:raw0 → raw35(cell43=切英數):Down×3 Right×8,space
k Down Down Down Right Right Right Right Right Right Right Right
k space; sleep 0.6; shot 02_alnum

# 3) 移到字元格 raw0('0'):從 raw35 Up×3 → raw8,Left×8 → raw0;space 打 1 字
k Up Up Up Left Left Left Left Left Left Left Left
k space; sleep 0.5; shot 03_typed

# 4) 完成:raw0 → raw44(cell44=完成):Down×4 Right×8,space
k Down Down Down Down Right Right Right Right Right Right Right Right
k space; sleep 1.2; shot 04_finish

# 5) 進開場 / 角色建立 / 遊戲:多按 Return 推進並截圖
for n in 1 2 3 4 5 6 7 8 9 10; do k Return; sleep 1.1; shot "05_${n}"; done

# 6) 嘗試走動(若已可操作)
for d in Down Down Right Right Up Up Left Left; do k "$d"; sleep 0.4; done
shot 20_walk
for d in Down Down Down Down Right Right Right Right; do k "$d"; sleep 0.4; done
shot 21_walk2

sleep 1; kill $DB 2>/dev/null
echo "=== shots ==="; ls -la "$OUT"/ow_*.png 2>/dev/null | tail -25