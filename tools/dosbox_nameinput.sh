#!/usr/bin/env bash
# 推過注音姓名輸入 (依 re/nameinput.c 精確 raw-index 語意) -> 進城鎮/世界圖/觸發戰鬥。
# 在 dq3-dosbox 容器內跑。所有等待有上界,無 sentinel 輪詢。
#
# grid 是 1-D ring raw[0..44]:Up=-9 Down=+9 Left=-1 Right=+1,皆 mod45。
# cell = col*5+row,其中 row=raw/9,col=raw%9。
#  - TOFN(切功能列)=cell43=col8,row3 => raw=3*9+8=35。
#  - alnum cell0(左上字元格)=raw0。
# 起始游標 = raw0(實測 ni_00 紅框在 ㄅ)。
# 功能列(win_nav 5列,起始 row1 英數):Up/Down 移列,Space 選;完成=row5 => Down×4。
set -u
OUT=/work/dosbox; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/x.log 2>&1 &
sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/db.log 2>&1 &
DB=$!; sleep 6
W="$(xdotool search --name DOSBox|head -1)"
xdotool windowfocus "$W" 2>/dev/null; xdotool windowactivate "$W" 2>/dev/null
shot(){ import -window root "$OUT/ni_$1.png" 2>/dev/null; echo "shot $1"; }
k(){ xdotool windowfocus "$W" 2>/dev/null; xdotool key --window "$W" --clearmodifiers "$1"; sleep "${2:-0.4}"; }
rep(){ local key=$1 n=$2; for ((i=0;i<n;i++)); do k "$key" 0.3; done; }

# === 開場 -> 標題 -> 主選單 -> 遊戲開始 -> 姓名輸入 ===
for n in $(seq 1 7); do k Return 1.4; done
k Return 1.4    # 標題->主選單
k Return 1.4    # 遊戲開始->姓名輸入 (游標 raw0=ㄅ)
shot 00_name_start

# === 從 raw0 -> raw35(TOFN cell43):Down×3 Right×8 ===
rep Down 3; rep Right 8; shot 01_at_tofn
k space 0.6     # 選 TOFN -> 進功能列 (row1 英數)
shot 02_in_fnlist
k space 0.6     # 選 row1 英數 -> 切英數鍵盤
shot 03_alnum_mode

# === 英數模式:游標仍 raw35;-> raw0(左上字元格):Up×3 Left×8 ===
rep Up 3; rep Left 8; shot 04_alnum_topleft
k space 0.6     # 選 cell0 字元 -> 名字 +1
shot 05_typed_one

# === 再從 raw0 -> raw35(TOFN) -> 功能列 -> 完成(row5) ===
rep Down 3; rep Right 8; shot 06_at_tofn2
k space 0.6; shot 07_in_fnlist2     # 進功能列 (row1)
rep Down 4; shot 08_at_done          # row1 -> row5 完成
k space 0.8; shot 09_after_done      # 選完成 -> 名字非空 -> 離開
k Return 1.0; shot 10_post1
k Return 1.0; shot 11_post2
k Return 1.0; shot 12_post3

# === 進遊戲後:推過開場劇情對話 (狂按 Return 清對話,有界) ===
for n in $(seq 1 14); do k Return 0.9; done
shot 20_town_a
# 系統性找城鎮出口 -> 城外 -> 世界圖。每方向長走 + Return 清沿途對話。
for round in 1 2 3 4 5; do
  rep Down 14;  k Return 0.5; shot "2${round}_dn"
  rep Right 14; k Return 0.5; shot "2${round}_rt"
  rep Up 14;    k Return 0.5; shot "2${round}_up"
  rep Left 14;  k Return 0.5; shot "2${round}_lf"
done
shot 30_after_explore
# === 若到世界圖:草地走動觸發隨機遇敵 -> 戰鬥 ===
for n in $(seq 1 60); do
  k Down 0.2; k Up 0.2
  case $n in 15|30|45|60) shot "35_walk_$n";; esac
done
shot 40_final

sleep 1; kill $DB 2>/dev/null
echo "=== shots ==="; ls "$OUT"/ni_*.png 2>/dev/null
