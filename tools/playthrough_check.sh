#!/usr/bin/env bash
# headless 主線里程碑驗證(debug 定位 + 腳本輸入 + stderr 里程碑斷言)。
# 用法: tools/playthrough_check.sh [assets_dir] [dq3_remake_bin]
# 依賴 docs/46 的 DQ3_DEBUG / DQ3_INPUT / DQ3_DUMP。不用真的玩到那段就驗證各系統串接。
set -u
ASSETS="${1:-/repo/assets_raw}"
BIN="${2:-/work/build/dq3_remake}"
pass=0; fail=0
chk() {  # chk "名稱" "DQ3_DEBUG" "DQ3_INPUT" "期望 stderr 正則"
    local name="$1" dbg="$2" inp="$3" want="$4"
    local out; out=$(DQ3_DEBUG="$dbg" DQ3_INPUT="$inp" DQ3_DUMP=/tmp/pt.ppm \
        timeout 25 "$BIN" "$ASSETS" game 2>&1)
    if echo "$out" | grep -qE "$want"; then echo "  [PASS] $name"; pass=$((pass+1))
    else echo "  [FAIL] $name (want /$want/)"; echo "$out" | grep -E "playthrough|DEBUG|錯|Segmentation" | head -2; fail=$((fail+1)); fi
}
echo "== DQ3 主線里程碑驗證 =="
chk "全城載入(warp CTY13 迷宮)"  "warp:13:5:5"        "."    "playthrough 末幀"
chk "話す NPC(阿里阿罕對話)"      "warp:0:37:15"       "e"    "話す NPC@.*bank=D3TXT01"
chk "武器/防具店(走到店員開店)"   "warp:0:27:32"       "e"    "設施:武器/防具店.*品項"
chk "建測試隊"                     "party"              "."    "party → 名冊4 隊伍4"
chk "下降至下層 overworld"         "descent"            "dd"   "scripted_event 86 下降.*下層"
chk "彩虹水滴合成事件(83)"        "event:0x53"         "."    "event 83 → result=117"
chk "進羅馬利亞 + 金錢"            "gold:5000;warp:2:20:10" "." "warp → CTY2"
chk "進度旗標流(取船里程碑)"      "prog:5;prog"        "."    "進度階段 6/9 = 取船"
chk "事件自動推進進度(下降→8)"   "prog:6;descent;prog" "."   "進度階段 8/9 = 下降"
echo "== 結果:PASS=$pass FAIL=$fail =="
[ "$fail" -eq 0 ]
