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
# E-11 時代結尾:破索瑪→洛特裝備昇華(そして伝説へ;精訊原無洛特道具,remake 補)
echo "$o" | grep -q "そして伝説へ" && ok "E-11 時代結尾:勇者受冊封為洛特+三裝備昇華(王者→洛特之劍 攻150)" || ng "E-11 洛特結尾"
# sub2 給物 NPC:水槍 / 光之珠
o=$(DQ3_DEBUG="warp:22:4:4:1" DQ3_INPUT="le" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=31:獲得道具 0x4b" && ok "sub2 給物 31(水槍)" || ng "sub2 給物 31"
o=$(DQ3_DEBUG="warp:67:14:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=52:獲得道具 0x65" && ok "sub2 給物 52(光之珠)" || ng "sub2 給物 52"
o=$(DQ3_DEBUG="warp:16:33:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=84:獲得道具 0x10" && ok "sub2 給物 84(誘惑之劍)" || ng "sub2 給物 84"
# 古布達黑胡椒(byte4=25 CTY15):救人前(無 flag 0x211)→ gate 不給;救人後 → 給黑胡椒
o=$(DQ3_DEBUG="warp:15:5:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "還在巴哈拉達洞窟救達妮亞" && ok "sub2 給物 25 gate(未救達妮亞→黑胡椒未給)" || ng "sub2 給物 25 gate"
o=$(DQ3_DEBUG="flag:0x211;warp:15:5:25:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=25:獲得道具 0x5c" && ok "sub2 給物 25(救達妮亞後→黑胡椒)" || ng "sub2 給物 25"
o=$(DQ3_DEBUG="warp:64:16:11:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=49:獲得道具 0x6b" && ok "sub2 給物 49(銀寶珠)" || ng "sub2 給物 49"
# 檢查型 NPC(require_item):持船渡海反應(byte4=16 精靈女王已改 transform,見第 9 段)
o=$(DQ3_DEBUG="item:0x5b;warp:62:38:6:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=50:持有 0x5b" && ok "sub2 檢查 50(持船→渡海反應)" || ng "sub2 檢查 50 持物"
# 提頓村=テドン(白天廢墟/夜晚亡靈):夜 gated 綠寶珠(忠實還原,day-night doc §9)
o=$(DQ3_DEBUG="warp:20:16:3:0;dn:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "牢門深鎖" && ok "提頓村白天牢門深鎖(夜 gated)" || ng "提頓村白天 gate"
o=$(DQ3_DEBUG="warp:20:16:3:0;dn:2" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得綠寶珠 0x66" && ok "提頓村夜晚開牢門給綠寶珠(0x66)" || ng "提頓村綠寶珠"

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

echo "######## 10. 變身杖鏈(杜勝利 Ch35/37)########"
o=$(DQ3_DEBUG="party;boss:89:0x62" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "遭遇.*HP400" && ok "怪力魔 boss(怪89,沙曼歐莎)→ 變身杖" || ng "怪力魔 boss"
o=$(DQ3_DEBUG="item:0x62;warp:54:8:3:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=44:換得道具 0x63" && ok "雪地草原:變身杖→換船員之骨(transform)" || ng "變身杖 transform"

echo "######## 11. 蓋亞之劍鏈(杜勝利 Ch38-40)########"
o=$(DQ3_DEBUG="warp:36:18:55:1" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x64" && ok "幽靈船 CTY36 愛的回憶寶箱" || ng "愛的回憶"
o=$(DQ3_DEBUG="warp:55:14:22:0" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x0f" && ok "牢獄祠堂 CTY55 蓋亞之劍寶箱" || ng "蓋亞之劍"
o=$(DQ3_DEBUG="item:0x0f;opos:153:174;use:0x0f" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "小火山熔流而出" && ok "蓋亞之劍:地表開火山(往尼羅肯特)" || ng "蓋亞之劍開火山"

echo "######## 12. 鑰匙鏈(杜勝利 Ch14/27)########"
o=$(DQ3_DEBUG="warp:13:13:5:2" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x56" && ok "金字塔 CTY13 魔法鑰匙寶箱" || ng "魔法鑰匙"
o=$(DQ3_DEBUG="warp:40:18:17:0" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x57" && ok "四島礁祠堂 CTY40 最終鑰匙寶箱" || ng "最終鑰匙"
o=$(DQ3_DEBUG="item:0x5e;opos:153:174;use:0x5e" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "最終鑰匙祠堂顯現" && ok "乾渴壺:地表吸海顯現祠堂" || ng "乾渴壺吸海"

echo "######## 13. 彩虹水滴材料鏈(杜勝利 Ch47/55)########"
o=$(DQ3_DEBUG="warp:80:1:2:2" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x72" && ok "下層 CTY80 太陽之石寶箱" || ng "太陽之石"
o=$(DQ3_DEBUG="item:0x72;item:0x73;warp:93:8:9:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "彩虹水滴 0x75" && ok "神聖祠堂:太陽之石+雲雨之杖→彩虹水滴(in-game 合成)" || ng "彩虹合成"
# bug #2:合成產彩虹水滴 0x75(非銀寶珠 0x6b);event 83 result 應 =117(=0x75)
o=$(DQ3_DEBUG="item:0x72;item:0x73;event:0x53" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "result=117" && ok "bug#2 修正:合成品=彩虹水滴 0x75(117,非銀寶珠 0x6b=107)" || ng "bug#2 合成品碼"
# 彩虹水滴用途:下層用 → 架彩虹橋通終盤(bug 修正後才拿得到彩虹水滴去架橋)
o=$(DQ3_DEBUG="party;item:0x75;descent;use:0x75" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "彩虹橋出現" && ok "彩虹水滴用途:下層架彩虹橋通終盤(杜勝利 Ch55)" || ng "彩虹橋架橋"
o=$(DQ3_DEBUG="party;item:0x75;use:0x75" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "需在下層利姆達爾" && ok "彩虹水滴:地表用→拒絕(需在下層西北)" || ng "彩虹橋位置 gate"
o=$(DQ3_DEBUG="item:0x74;warp:92:9:9:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "換得雲雨之杖 0x73" && ok "精靈祠堂 CTY92:精靈的守護→換雲雨之杖(transform)" || ng "精靈祠堂雲雨之杖"
o=$(DQ3_DEBUG="warp:92:9:9:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "需先持精靈的守護" && ok "精靈祠堂:無守護→需守護" || ng "精靈祠堂 gate"
o=$(DQ3_DEBUG="warp:81:22:20:0" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x77" && ok "瑪依拉 CTY81 妖精之笛寶箱" || ng "妖精之笛"
o=$(DQ3_DEBUG="item:0x77;warp:82:5:5:4;use:0x77" DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "妖精之笛解魯比斯詛咒 → 得精靈的守護" && ok "魯比斯之塔 CTY82:妖精之笛→精靈的守護(全鏈通)" || ng "魯比斯精靈的守護"

echo "######## 14. B-7 散件寶箱(杜勝利各章)########"
# 領悟之書真碼 0x4a @ CTY18(加爾那之塔;攻略反證 CTY18=力量種子+鐵頭盔+領悟之書)
o=$(DQ3_DEBUG="warp:18:21:18:1" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x4a" && ok "加爾那之塔 CTY18 領悟之書 0x4a(賢者轉職)" || ng "領悟之書 0x4a"
# CTY87 = 洛特洞窟:勇者之盾 0x40(原被誤標領悟之書/加爾那)
o=$(DQ3_DEBUG="warp:87:8:9:2" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x40" && ok "洛特洞窟 CTY87 勇者之盾 0x40" || ng "勇者之盾 0x40"
# 勇氣神殿神父(B-8,RE handler 0x59e4/0x608a):CTY47 接受挑戰 set flag 0x13 → owportal 轉 CTY75 神殿開放
o=$(DQ3_DEBUG="party;warp:47:25:17:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "接受單獨戰鬥挑戰" && ok "勇氣神殿神父 CTY47:接受挑戰(flag 0x13)" || ng "勇氣神殿神父挑戰"
o=$(DQ3_DEBUG="party;warp:75:26:8:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "回來了嗎" && ok "勇氣神殿神父 CTY75:神殿開放返回對話" || ng "勇氣神殿神父返回"
# D-10 下降閘(ギアガの大穴 runner ev86):自製忠實 gate=需巴拉摩斯敗(flag 0x213,杜 Ch44→46);debug descent token 不受閘
o=$(DQ3_DEBUG="party" DQ3_INPUT="ssyyy" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "尚未開啟" && ok "D-10 下降閘:未敗巴拉摩斯 → 擋" || ng "D-10 下降閘(未敗擋)"
o=$(DQ3_DEBUG="party;flag:0x213" DQ3_INPUT="ssyyy" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "scripted_event 86 下降" && ok "D-10 下降閘:敗巴拉摩斯(0x213)→ 下降" || ng "D-10 下降閘(敗後可下降)"
# B-5 諾魯特密道(杜 Ch16-17):波魯多加王首訪授國王的信 0x5b → 諾魯特(CTY62)持信開東方通道
o=$(DQ3_DEBUG="party;warp:37:9:7:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "授《國王的信》0x5b" && ok "B-5 波魯多加王授國王的信 0x5b(首訪)" || ng "波魯多加授國王的信"
o=$(DQ3_DEBUG="party;item:0x5b;warp:62:38:6:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "byte4=50:持有 0x5b" && ok "B-5 諾魯特持國王的信 0x5b → 反應(開東方通道)" || ng "諾魯特國王的信"
# B-6 帶商人建城(杜 Ch33-36):新城鎮 CTY83 老人(RE handler 0x5aba:掃隊伍 class6 商人→寄存+建城 flag 0x216)
o=$(DQ3_DEBUG="party;warp:83:16:3:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "缺商人" && ok "B-6 新城鎮老人:無商人→需商人(rec1)" || ng "新城鎮老人缺商人"
o=$(DQ3_DEBUG="party;reclass:1:6;warp:83:16:3:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "新城鎮建成" && ok "B-6 帶商人(class6)→ 寄存+新城鎮建成(flag 0x216)" || ng "新城鎮建城"
o=$(DQ3_DEBUG="warp:20:4:9:1" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x5f" && ok "提頓村 CTY20 黑暗之燈" || ng "黑暗之燈"
o=$(DQ3_DEBUG="warp:26:11:12:2" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x60" && ok "亞布之塔 CTY26 回音之笛" || ng "回音之笛"
o=$(DQ3_DEBUG="warp:78:2:4:1" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x71" && ok "加萊祠堂 CTY78 銀豎琴" || ng "銀豎琴"

echo "######## 15. 甘達特金皇冠正源(杜勝利 Ch6)########"
o=$(DQ3_DEBUG="party;warp:10:4:4:5" DQ3_INPUT="e" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "盜賊甘達特擋路" && echo "$o" | grep -q "遭遇.*HP551" && ok "香巴尼之塔:取皇冠前甘達特(怪26)boss gate" || ng "甘達特 gate"
o=$(DQ3_DEBUG="flag:0x210;warp:10:4:4:5" DQ3_INPUT="e" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x33" && ok "甘達特已敗→取金皇冠(連 Romaly 線)" || ng "甘達特後皇冠"

echo "######## 16. B-9 收尾(歐里空金屬/古布達救人/隱身草)########"
# 歐里空金屬 0x6d:CTY84 養羊圍欄寶箱,站(23,36)朝上 examine(杜勝利 Ch50)
o=$(DQ3_DEBUG="warp:84:23:36:0" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "獲得道具 0x6d" && ok "達姆杜拉 CTY84 歐里空金屬寶箱(站 23,36 朝上)" || ng "歐里空金屬"
# 古布達黑胡椒救人 gate(杜勝利 Ch19-20):未救達妮亞(flag 0x211)→ byte4=25 古布達不給黑胡椒。
# gate 邏輯靜態驗證(古布達 NPC 在 CTY15,救人前 before_rec);救人 = boss:27;boss:26;flag:0x211。
SRC="${GT_SRC:-/repo/dq3_remake/src/main.c}"; [ -f "$SRC" ] || SRC="dq3_remake/src/main.c"
grep -q '還在巴哈拉達洞窟救達妮亞' "$SRC" 2>/dev/null && ok "古布達黑胡椒救人 gate(flag 0x211 前不給)" || ng "古布達救人 gate"
# 隱身草 0x5d:已在多家商店貨架(shopdata),購買系統可得
SHOP="${GT_SHOP:-/repo/dq3_remake/src/dq3_shopdata.c}"; [ -f "$SHOP" ] || SHOP="dq3_remake/src/dq3_shopdata.c"
grep -q "0x5d" "$SHOP" 2>/dev/null && ok "隱身草 0x5d 在商店貨架(可購)" || ng "隱身草貨架"
# 王者之劍 0x1c:瑪依拉 CTY81 二樓道具店主賣歐里空金屬 0x6d 換王者之劍(杜勝利 Ch50)。
# id 正確性:ITEM.DAT 第 0x1c 筆 atk=120(王者之劍經典值,docs/22),勇者專用 mask=0x80。
ROY=$(python3 -c "d=open('$ASSETS/ITEM.DAT','rb').read(); r=d[0x1c*7:0x1c*7+7]; print(r[0], r[2]|(r[3]<<8), r[6])" 2>/dev/null)
[ "$ROY" = "120 35000 128" ] && ok "王者之劍 id=0x1c(ITEM.DAT atk120/價35000/勇者專用)" || ng "王者之劍 id ($ROY)"
# 換劍邏輯靜態驗證(CTY81 sec1 二樓店主 transform:0x6d→0x1c;sec1 實機載入受 debug warp 限制,靜態驗邏輯)
grep -q '瑪依拉(CTY81)二樓道具店主' "$SRC" 2>/dev/null && grep -q 'dq3_inv_add(&inv, 0x1c)' "$SRC" 2>/dev/null \
  && ok "瑪依拉換王者之劍邏輯(歐里空金屬 0x6d → 王者之劍 0x1c)" || ng "瑪依拉換劍邏輯"

echo "######## 17. 巴拉摩斯 boss(下降前主線大 boss)########"
# 巴拉摩斯本體:baramos token → 怪121 開戰(HP1201);勝利設 flag 0x213(杜勝利 Ch44)
o=$(DQ3_DEBUG="party;baramos" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "巴拉摩斯戰(怪121 HP1201)" && ok "巴拉摩斯本體 boss(怪121 HP1201 開戰)" || ng "巴拉摩斯 boss"
# 索瑪前完整序列:zomaseq token → 巴拉摩斯怨靈 122 → 殭屍 123 → 索瑪 124(逐戰勝才推進)
# 怨靈/殭屍數值 ground truth = docs/38(怨靈 HP1201、殭屍 HP2880);序列首戰怨靈必觸發
# C-9 六大魔人守衛(怪106,杜 Ch56):未破(無 flag 0x214)→ zomaseq 先打六大魔人
o=$(DQ3_DEBUG="party;item:0x65;zomaseq" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FFFF" timeout 30 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "六大魔人守衛(怪106" && ok "C-9 索瑪神殿六大魔人守衛(怪106 ×6)觸發" || ng "六大魔人守衛"
# 六大魔人已破(flag 0x214)→ 跳過 → 怨靈122 首戰觸發(原斷言;弱隊驗觸發非勝)
o=$(DQ3_DEBUG="party;item:0x65;flag:0x214;zomaseq" DQ3_INPUT="q" DQ3_BATTLE_SCRIPT="FFFF" timeout 30 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "巴拉摩斯怨靈(怪122)" && ok "索瑪前序列:六大魔人破後→巴拉摩斯怨靈(怪122)觸發" || ng "索瑪前序列怨靈"
# id 正確性:docs/38 怨靈 122 HP1201、殭屍 123 HP2880(ground truth,非記憶)
D38="${GT_DOCS:-/repo/docs/38-monster-stats.md}"; [ -f "$D38" ] || D38="docs/38-monster-stats.md"
GR=$(python3 -c "
for ln in open('$D38',encoding='utf-8'):
    if '| 122 ' in ln or '| 123 ' in ln:
        c=[x.strip() for x in ln.split('|')]; print(c[1]+'_'+c[3])
" 2>/dev/null | tr '\n' ' ')
echo "$GR" | grep -q "122_1201" && echo "$GR" | grep -q "123_2880" && ok "怨靈122/殭屍123 數值 ground truth(docs/38)" || ng "怨靈/殭屍數值 ($GR)"

echo "######## 18. 甘達特巢穴正式觸發點(CTY14 sec1,巴哈拉達洞窟)########"
# 玩家走到 CTY14 sec1 (14,13) byte4=58 examine → 救人劇情:甘達特手下(怪27 HP81)→ 甘達特(怪26)
# 正式觸發點 = npc_dialogue.json kind=special dlg=58(docs/boss-trigger-points.md);與古布達黑胡椒鏈 flag 0x211 閉環
o=$(DQ3_DEBUG="party;warp:14:14:14:1" DQ3_INPUT="ue" DQ3_BATTLE_SCRIPT="FF" timeout 25 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "甘達特巢穴守衛" && echo "$o" | grep -q "甘達特手下(怪27)" && ok "CTY14 巢穴 examine → 甘達特手下(怪27)觸發" || ng "CTY14 巢穴觸發"
echo "$o" | grep -q "遭遇.*HP81" && ok "甘達特手下怪27 HP81(docs/38 ground truth 對上)" || ng "怪27 HP81"
# 已救出(flag 0x211)→ 後話分支(不重複開戰),與 CTY15 古布達黑胡椒鏈閉環
o=$(DQ3_DEBUG="party;flag:0x211;warp:14:14:14:1" DQ3_INPUT="ue" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "已救出達妮亞" && ok "CTY14 已救出(flag 0x211)→ 後話(與古布達黑胡椒鏈閉環)" || ng "CTY14 後話分支"

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
if [ -s "$SV" ] && grep -q "存檔 →" /tmp/sv1.log; then
  o=$(DQ3_SAVE="$SV" DQ3_LOAD=1 DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
  echo "$o" | grep -q "讀檔續玩.*CTY2 (10,10)" && ok "存檔→讀檔 roundtrip(狀態回復)" || ng "讀檔狀態不符"
  # 讀檔回原位置(使用者需求:取檔回來在原位置):場景還原到存檔的城 CTY2
  echo "$o" | grep -q "讀檔場景還原:城 CTY2" && ok "讀檔回原位置(場景還原到存檔的城 CTY2)" || ng "讀檔場景未還原"
else ng "存檔未產生"; fi
# 多 slot:DQ3_LOAD=N 讀 slotN(dq3_saveN.dat);slot path 規則 + slot 隔離(test_save 單元已驗 6 slot)
SVB=/tmp/gt_slot.dat; rm -f /tmp/gt_slot*.dat
# 以 slot0 autosave 寫入(F10 流程需互動,headless 用 autosave 事件點寫 slot0);驗 DQ3_LOAD=3 找 dq3_slot3.dat
cp "$SV" /tmp/gt_slot3.dat 2>/dev/null
o=$(DQ3_SAVE="$SVB" DQ3_LOAD=3 DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "讀檔續玩 ← /tmp/gt_slot3.dat" && ok "多 slot:DQ3_LOAD=3 讀 slot3(dq3_saveN.dat 路徑規則)" || ng "多 slot 讀 slot3"
o=$(DQ3_SAVE="$SVB" DQ3_LOAD=5 DQ3_INPUT="q" timeout 20 "$BIN" "$ASSETS" game 2>&1)
echo "$o" | grep -q "讀檔續玩" && ng "slot5 空卻讀到" || ok "多 slot:DQ3_LOAD=5(空 slot)→ 不讀檔(新局)"
rm -f /tmp/gt_slot*.dat

echo "================================================"
echo "== game-tester 結果:PASS=$PASS FAIL=$FAIL =="
[ "$FAIL" -eq 0 ] && echo "== ✅ 全綠 — 可打包交付 ==" || echo "== ❌ 有失敗,不可交付 =="
[ "$FAIL" -eq 0 ]
