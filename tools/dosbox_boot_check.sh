#!/usr/bin/env bash
# 最小檢查:啟 DOSBox,等久一點,確認 window 存在 + title 有畫面。
set -u
OUT=/work/dosbox/tavern; mkdir -p "$OUT"; export DISPLAY=:99
Xvfb :99 -screen 0 1024x768x24 -ac >/tmp/xvfb.log 2>&1 & XV=$!; sleep 2
dosbox -conf /work/tools/dosbox.conf >/tmp/dosbox.log 2>&1 & DB=$!; sleep 12
WIN="$(xdotool search --name DOSBox 2>/dev/null|head -1)"
echo "WIN=[$WIN]"
import -window root "$OUT/boot_check.png" 2>/dev/null || { xwd -root -silent|convert xwd:- "$OUT/boot_check.png"; }
echo "boot shot size: $(stat -c %s "$OUT/boot_check.png")"
echo "=== dosbox.log ==="; cat /tmp/dosbox.log | tail -15
sleep 0.3; kill $DB 2>/dev/null; kill $XV 2>/dev/null
