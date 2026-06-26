#!/usr/bin/env bash
# dq3_remake 完整 game-tester 驗證套件(交付前 gate)。
# 整合:單元測試 ×16 + playthrough_check(系統孤立)+ mainline_check(主線一條龍)
#       + 全 CTY 載入零崩潰掃描 + 新內容(索瑪二階段/結局捲動/sub2 給物 NPC)。
# 全綠 = 可打包交付。容器內執行(SDL dummy);用法見 tools/run_game_tester.sh。
set -u
ASSETS="${1:-/repo/assets_raw}"
BIN="${2:-/build/dq3_remake}"
BUILD="$(dirname "$BIN")"
export SDL_VIDEODRIVER=dummy SDL_AUDIODRIVER=dummy
PASS=0; FAIL=0
ok(){ echo "  [PASS] $1"; PASS=$((PASS+1)); }
ng(){ echo "  [FAIL] $1"; FAIL=$((FAIL+1)); }

echo "######## 1. 單元測試(數值/戰鬥/系統)########"
for t in stats combat monster battle inventory door npc roster menu dialogue \
         nameinput tavern zhuyin save config; do
  b="$BUILD/dq3_${t}_test"
  [ -x "$b" ] || { ng "單元 $t(缺 binary)"; continue; }
  if timeout 40 "$b" "$ASSETS" >/tmp/ut_$t.log 2>&1; then ok "單元 $t"; else ng "單元 $t"; tail -3 /tmp/ut_$t.log|sed 's/^/      /'; fi
done

echo "######## 2. playthrough_check(各系統孤立斷言)########"
if bash /repo/tools/playthrough_check.sh "$ASSETS" "$BIN" >/tmp/pt.log 2>&1; then
  n=$(grep -c '\[PASS\]' /tmp/pt.log); ok "playthrough_check($n 項全綠)"
else ng "playthrough_check"; grep '\[FAIL\]' /tmp/pt.log|sed 's/^/      /'; fi

echo "######## 3. mainline_check(主線一條龍 → 9/9 破關)########"
if bash /repo/tools/mainline_check.sh "$ASSETS" "$BIN" >/tmp/ml.log 2>&1; then
  ok "mainline_check(主線推進 9/9)"
else ng "mainline_check"; grep '\[FAIL\]' /tmp/ml.log|sed 's/^/      /'; fi

echo "######## 4. 全 CTY 城鎮載入零崩潰掃描 ########"
crash=0; loaded=0
for f in "$ASSETS"/CTY*.DAT; do
  [ -e "$f" ] || continue
  num=$(basename "$f" | sed 's/CTY0*\([0-9]*\)\.DAT/\1/'); [ -z "$num" ] && num=0
  out=$(DQ3_DEBUG="warp:$num:5:5" DQ3_INPUT="q" DQ3_DUMP=/tmp/c.ppm timeout 20 "$BIN" "$ASSETS" game 2>&1)
  rc=$?
  if [ $rc -ge 128 ]; then echo "      CTY$num 崩潰(rc=$rc)"; crash=$((crash+1));
  elif echo "$out" | grep -q "warp → CTY$num"; then loaded=$((loaded+1)); fi
done
if [ $crash -eq 0 ]; then ok "全 CTY 載入零崩潰($loaded 城成功 warp)"; else ng "$crash 城崩潰"; fi

echo "######## 5. 新內容(索瑪二階段 / 結局捲動 / sub2 給物)########"
# 索瑪二階段:持光之珠 → 弱化訊息
o=$(DQ3_DEBUG="party;item:0x65;zoma" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FFFF" timeout 30 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "驅散索瑪的黑暗結界" && ok "索瑪二階段(光之珠弱化)" || ng "索瑪二階段"
# 結局捲動:finale → THE END
EE=$(printf 'e%.0s' $(seq 1 80))
o=$(DQ3_DEBUG="party;finale" DQ3_INPUT="$EE" timeout 30 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "THE END" && ok "結局捲動(ENDTXT 全段 → THE END)" || ng "結局捲動"
# sub2 給物 NPC:水槍 / 光之珠
o=$(DQ3_DEBUG="warp:22:4:4:1" DQ3_INPUT="le" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=31:獲得道具 0x4b" && ok "sub2 給物 31(水槍)" || ng "sub2 給物 31"
o=$(DQ3_DEBUG="warp:67:14:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=52:獲得道具 0x65" && ok "sub2 給物 52(光之珠)" || ng "sub2 給物 52"
o=$(DQ3_DEBUG="warp:16:33:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=84:獲得道具 0x10" && ok "sub2 給物 84(誘惑之劍)" || ng "sub2 給物 84"
o=$(DQ3_DEBUG="warp:15:5:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=25:獲得道具 0x5c" && ok "sub2 給物 25(黑胡椒)" || ng "sub2 給物 25"
o=$(DQ3_DEBUG="warp:64:16:11:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=49:獲得道具 0x6b" && ok "sub2 給物 49(銀寶珠)" || ng "sub2 給物 49"
# 檢查型 NPC(require_item):持船渡海反應(byte4=16 精靈女王已改 transform,見第 9 段)
o=$(DQ3_DEBUG="item:0x5b;warp:62:38:6:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=50:持有 0x5b" && ok "sub2 檢查 50(持船→渡海反應)" || ng "sub2 檢查 50 持物"
# 提頓村牢房犯人給綠寶珠(青衫攻略 runner 事件,remake 簡化 talk 即給)
o=$(DQ3_DEBUG="warp:20:16:3:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得綠寶珠 0x66" && ok "提頓村牢房給綠寶珠(0x66)" || ng "提頓村綠寶珠"

echo "######## 8. 不死鳥拉米亞飛行坐騎 ########"
o=$(DQ3_DEBUG="phoenix" DQ3_INPUT="eesrrrrdddd" DQ3_DUMP=/tmp/ph.ppm timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "不死鳥.*復活" && echo "$o" | grep -q "末幀.*in_town=0" \
  && ok "不死鳥復活+飛行(飛越地形/不進城,留 overworld)" || ng "不死鳥飛行"
o=$(DQ3_DEBUG="phoenix" DQ3_INPUT="eesy" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "不死鳥降落於" && ok "不死鳥降落於可走格" || ng "不死鳥降落"
# 集六珠(0x66-0x6b 綠藍紅紫黃銀)→ 拉米亞自動復活 → 起飛
o=$(DQ3_DEBUG="item:0x66;item:0x67;item:0x68;item:0x69;item:0x6a;item:0x6b" DQ3_INPUT="eesy" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "六珠齊備.*拉米亞復活" && echo "$o" | grep -q "起飛" && ok "六珠齊備→拉米亞復活→起飛" || ng "六珠復活"
o=$(DQ3_DEBUG="item:0x66;item:0x67;item:0x68;item:0x69;item:0x6a" DQ3_INPUT="ee" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "拉米亞復活" && ng "五珠不應復活" || ok "五珠未集齊(不復活)"
# 六珠來源:藍/紅/黃寶箱(既有寶箱系統)
o=$(DQ3_DEBUG="warp:23:24:14:2" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x67" && ok "藍寶珠寶箱 CTY23(勇氣洞窟)" || ng "藍寶珠"
o=$(DQ3_DEBUG="warp:27:1:12:1" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x68" && ok "紅寶珠寶箱 CTY27(海盜村)" || ng "紅寶珠"
o=$(DQ3_DEBUG="warp:83:4:2:0" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x6a" && ok "黃寶珠寶箱 CTY83(新城鎮)" || ng "黃寶珠"

echo "######## 9. 夢幻紅寶石鏈(杜勝利 Ch9-11)########"
o=$(DQ3_DEBUG="warp:11:21:20:3" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x59" && ok "地底湖洞窟 CTY11 夢幻紅寶石寶箱" || ng "夢幻紅寶石"
o=$(DQ3_DEBUG="item:0x59;warp:5:17:8:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=16:換得道具 0x5a" && ok "精靈女王:紅寶石→換覺醒粉(transform)" || ng "精靈女王 transform"
o=$(DQ3_DEBUG="party;item:0x5a;warp:4:5:5:0;use:0x5a" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "諾阿尼魯村.*催眠詛咒解除" && ok "諾阿尼魯:用覺醒粉解全村催眠" || ng "諾阿尼魯解催眠"

echo "######## 7. boss 劇情事件(甘達特 / 八頭大蛇)########"
# 甘達特(26)boss token:開戰(HP551)
o=$(DQ3_DEBUG="party;boss:26:0x33" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "遭遇.*HP551" && ok "boss token 甘達特(26 HP551)" || ng "boss token 甘達特"
# 八頭大蛇(75)sprite 解碼 + 開戰(HP801,曾因 sprite W=416>384 被擋)
o=$(DQ3_DEBUG="party;boss:75" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "遭遇.*HP801" && ok "boss 八頭大蛇(75 HP801,sprite W416 可繪)" || ng "八頭大蛇 sprite/戰"
# in-game 八頭大蛇觸發(CTY19 byte4=45 NPC)
o=$(DQ3_DEBUG="party;warp:19:35:13:1" DQ3_INPUT="ue" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "遭遇.*HP801" && ok "in-game 八頭大蛇觸發(CTY19)" || ng "in-game 八頭大蛇觸發"

echo "######## 6. 存檔/讀檔 roundtrip(整合)########"
SV=/tmp/gt_save.dat; rm -f "$SV"
DQ3_SAVE="$SV" DQ3_DEBUG="party;gold:5000;item:0x55;warp:2:10:10" DQ3_INPUT="que" timeout 20 "$BIN" "$ASSETS" game >/tmp/sv1.log 2>&1
if [ -s "$SV" ] && grep -q "自動存檔" /tmp/sv1.log; then
  o=$(DQ3_SAVE="$SV" DQ3_LOAD=1 DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
  echo "$o" | grep -q "讀檔續玩.*CTY2 (10,10)" && ok "存檔→讀檔 roundtrip(狀態回復)" || ng "讀檔狀態不符"
else ng "存檔未產生"; fi

echo "================================================"
echo "== game-tester 結果:PASS=$PASS FAIL=$FAIL =="
[ "$FAIL" -eq 0 ] && echo "== ✅ 全綠 — 可打包交付 ==" || echo "== ❌ 有失敗,不可交付 =="
[ "$FAIL" -eq 0 ]
