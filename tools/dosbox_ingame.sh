#!/usr/bin/env bash
# 進階場景:從標題 -> 主選單 -> 完成姓名輸入 -> 進入遊戲 (城鎮/世界圖/戰鬥)。
# 在 dq3-dosbox 容器內執行。所有等待有上界。
set -u
SCENARIO="${1:-ingame}"
OUT=/work/dosbox
mkdir -p "$OUT"
export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 &
sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 &
DB=$!
sleep 6

WIN=""
fw(){ WIN="$(xdotool search --name DOSBox 2>/dev/null | head -1)"; }
shot(){ import -window root "$OUT/${SCENARIO}_$1.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/${SCENARIO}_$1.png" 2>/dev/null; }; echo "shot $1"; }
k(){ fw; if [ -n "$WIN" ]; then xdotool key --window "$WIN" "$1"; else xdotool key "$1"; fi; }

# 1) 推進開場動畫到標題 (多次 Return,有界)
for n in $(seq 1 6); do k Return; sleep 1.5; done
shot 01_title
# 2) 標題 -> 主選單 (再按 Return)
k Return; sleep 1.5; shot 02_menu
# 3) 主選單「遊戲開始」(預設選中) -> Return 進姓名輸入
k Return; sleep 1.5; shot 03_name
# 4) 姓名輸入:選一個注音字 (Return 選格),再移動到「完成」並 Return
k Return; sleep 0.5            # 選目前游標的注音
shot 04_name_typed
# 嘗試移到右側功能欄「完成」:多次 Right 跨到功能欄,Down 到完成,Return
for n in 1 2 3 4 5; do k Right; sleep 0.3; done
for n in 1 2 3; do k Down; sleep 0.3; done
k Return; sleep 1.5; shot 05_after_name
# 部分版本姓名完成後接 Return 確認 / 進開場劇情
for n in 1 2 3 4 5 6; do k Return; sleep 1.3; shot "06_${n}_story"; done
# 5) 進入可操作後:走動嘗試城鎮 / 世界圖
for d in Down Down Down Right Right Up Up Left Left; do k "$d"; sleep 0.5; done
shot 10_walk
for d in Down Down Down Down Down Down; do k "$d"; sleep 0.5; done
shot 11_walk2
# 6) 停在地表,等待步數遭遇觸發戰鬥 (反覆走動,有界)
for n in $(seq 1 24); do
  k Down; sleep 0.4; k Right; sleep 0.4
  case $n in 8|16|24) shot "12_${n}_explore";; esac
done
shot 20_final

sleep 1
kill $DB 2>/dev/null
echo "=== captures ==="; ls -la "$OUT/${SCENARIO}"_*.png 2>/dev/null
