#!/usr/bin/env bash
# 主線串接驗證(#4):單一進程依杜勝利攻略順序走完主線旗標流,
# 驗證「進度階段」單調推進 START→…→DESCEND,且真實 gate(登船/彩虹合成/下降)依序生效。
# 與 playthrough_check.sh(各系統孤立斷言)互補:此處證明「主線一條龍」串得起來。
# 用法: tools/mainline_check.sh [assets_dir] [dq3_remake_bin]
set -u
ASSETS="${1:-/repo/assets_raw}"
BIN="${2:-/work/build/dq3_remake}"

# 主線 debug chain(一個進程內依序套用):
#   建隊 → 開場/前四里程碑(prog:0..4,代表盜賊鑰匙/魔法球/羅馬利亞/達瑪線)
#   → 取船(ship,真 gate)→ 彩虹水滴合成(event 0x53,真 gate)→ 下降(descent,真 gate)
# 每步後 prog 印出當前階段,據此驗證單調推進。
CHAIN="party;prog:0;prog;prog:1;prog;prog:2;prog;prog:3;prog;prog:4;prog;ship;prog;event:0x53;prog;descent;prog"

OUT=$(DQ3_DEBUG="$CHAIN" DQ3_DUMP=/tmp/mainline.ppm timeout 30 "$BIN" "$ASSETS" game 2>&1)

echo "== DQ3 主線串接驗證(START→DESCEND 一條龍)=="

# 1) 進度階段序列單調遞增 1→8
STAGES=$(echo "$OUT" | grep -oE "進度階段 [0-9]+/9" | sed 's#.* ##; s#/9##' | head -8 | tr '\n' ' ')
echo "  進度階段序列: $STAGES"
EXPECT="1 2 3 4 5 6 7 8 "
fail=0
if [ "$STAGES" = "$EXPECT" ]; then echo "  [PASS] 階段單調推進 1→8(START→下降)"; else
    echo "  [FAIL] 階段序列非預期(want: $EXPECT)"; fail=$((fail+1)); fi

# 2) 三個真實 gate 依序在 log 出現
gate() { if echo "$OUT" | grep -qE "$2"; then echo "  [PASS] 真 gate:$1"; else echo "  [FAIL] 真 gate:$1 未觸發"; fail=$((fail+1)); fi; }
gate "取船(登船)"        "登船 → 可跨海|取得船"
gate "彩虹水滴合成(83)"   "event 83 → result=|彩虹水滴"
gate "下降アレフガルド"    "scripted_event 86 下降"

# 3) 終局階段 = 下降(8/9),距 ZOMA(終盤)一步
if echo "$OUT" | tail -5 | grep -qE "進度階段 8/9 = 下降"; then
    echo "  [PASS] 主線推進到下降アレフガルド(8/9,剩索瑪終戰)"; else
    echo "  [FAIL] 未推進到下降階段"; fail=$((fail+1)); fi

echo "== 結果:$([ $fail -eq 0 ] && echo 主線串接成立 || echo "FAIL=$fail") =="
[ "$fail" -eq 0 ]
