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
chk "登船 + 跨海航行"             "opos:149:179;ship:150:179" "rrr" "登船 → 可跨海"
chk "船上岸(船停原水格)"         "opos:150:179;ship"  "l"    "上岸 → 船停泊 \(150,179\)"
chk "藥草治療(封頂於 maxHP)"     "party;hurt:5;item:0x41;use:0x41" "." "藥草:隊員0 回復 HP"
chk "聖水驅弱敵"                   "party;item:0x44;use:0x44" "."  "聖水:弱敵迴避 64 步"
chk "蓋美拉翅膀回鎮"              "warp:2:20:10;item:0x43;use:0x43" "." "蓋美拉翅膀:回到地表"
chk "索瑪終戰破關(結局)"        "prog:6;descent;finale;prog" "." "進度階段 9/9"
chk "取船:無胡椒對國王(等胡椒)"  "warp:37:9:7"        "ue"   "波魯多加國王:等黑胡椒"
chk "取船:獻黑胡椒換船"          "item:0x5c;warp:37:9:7" "ue" "獻黑胡椒給波魯多加國王 → 得船"
chk "胡椒補洞:CTY15 道具店進貨"   "gold:5000;warp:15:11:16" "ue" "本城道具店補進黑胡椒"
chk "盜賊鑰匙:拿吉米老人(給+rec52)" "warp:8:9:10:3"   "ue"   "獲得盜賊的鑰匙.*rec52"
chk "拿吉米條件對話:已有鑰匙後話"  "item:0x55;warp:8:9:10:3" "ue" "已給過鑰匙 → 後話 rec53"
chk "魔法玉:雷貝 2F 老人(需鑰匙)" "item:0x55;warp:1:3:4:1" "ue" "雷貝村 2F 老人.*獲得魔法玉"
chk "羅馬利亞:持皇冠還國王"       "item:0x33;warp:2:7:3:1" "ue" "羅馬利亞國王:歸還金皇冠"
chk "達瑪神殿:進城達成 DHAMA"     "warp:49:5:5"        "."    "抵達達瑪神殿 → 轉職開放"
chk "達瑪轉職:隊員換職(Lv1減半)" "party;reclass:1:3"  "."    "達瑪轉職:隊員1 職業 1→3"
chk "狀態:驅毒草解中毒"           "party;status:0:1;item:0x42;use:0x42" "." "驅毒草:隊員0 解除中毒"
chk "狀態:滿月草解麻痺"           "party;status:0:2;item:0x45;use:0x45" "." "滿月草:隊員0 解除麻痺"
chk "野外道具選單:選藥草使用"     "party;hurt:5;item:0x41;warp:0:27:30" "cdrede" "野外つかう:藥草"
chk "達瑪轉職選單:選隊員+職業"    "party;dhama"        "dededdde" "達瑪轉職:隊員"
chk "戰鬥毒傷(中毒進戰鬥)"        "party;status:0:1;zoma" "FFFF" "中毒,受 3 毒傷"
chk "野外選單:蓋美拉翅膀回地表"   "party;warp:2:10:10;item:0x43" "cdredde" "蓋美拉翅膀:回到地表"
chk "迷宮樓梯層間轉場(拿吉米塔)"  "warp:8:15:38"       "u"    "門/階梯:→ CTY8 section5"
echo "== 結果:PASS=$pass FAIL=$fail =="
[ "$fail" -eq 0 ]
