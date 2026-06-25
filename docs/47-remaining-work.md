# 剩餘工作盤點(往「完整還原 + 能破關」)

> 階段性盤點:remake 已能做什麼、離「能破關」還差什麼,依優先序。驗證口見 docs/46。

## ★ 前提:這是「未發售的早期 build」

精訊版 DQ3 是**未上市的早期版本**。實證觀察:部分劇情對話/事件**內容不全或未接好**——
例如下降事件(event 86,ギアガの大穴)的 **handler 本體存在**(0x783d 反組譯到),但**觸發那塊
追不到靜態 setter**(`[0x722]` 無 `mov imm`);搜尋「跳進坑」之類提示 NPC 對白也找不到乾淨的一句。
合理推斷:**這個 build 可能根本還沒把部分劇情觸發/文本串上去**(非我沒找到)。

⇒ 對「完整還原」的意義:**不必硬復刻一個原版可能都還沒做完的觸發**。務實做法 = 用 debug 口
(docs/46)把各段串成可玩可驗 + 設計 remake 自有的旗標流補上缺口,而非 1:1 還原半成品 runner VM。

## 已完成(RE + remake 落地)

| 系統 | 狀態 |
|---|---|
| 素材萃取 | 字型 / tile / sprite / palette / 怪物 / ITEM.DAT / 咒文 / 雙 overworld 地圖,全解 |
| 戰鬥 | 互動戰鬥 + 怪物 AI + 咒文 + 裝備加成 + **持久 HP/MP** + 升級(7 bug 全修)|
| 創角 | 露依達酒場 + 8 職業 + 注音/英數姓名輸入 + 隊伍編成 |
| 城鎮 | **全 89 城載入零崩潰**(debug 掃描驗證)+ NPC + 逐句對話(byte4→bank) |
| 設施 | 武器/防具店 · 道具店(真實逐城貨架)· 旅社(治療+扣費)· 教會(復活)· 記錄(存檔)|
| 寶箱 | 城鎮/迷宮寶箱 + 隱藏物品(121 處)|
| 移動 | overworld(雙層:地表 DQ3CON / 下層 DQ3UND)+ 城鎮 + 轉場(門/階梯/出城/0xfe 下降)|
| 傳送 | type-2 examine warp(0x4ea0)· overworld 旗標 portal(城鎮變體)|
| 遭遇 | region-based 遇敵 |
| 存檔 | dq3_save(名冊/隊伍/道具/位置/持久 HP)|
| scripted event | 彩虹水滴合成(83)· 下降(86,場景已可重現)|
| 對話變數 | {V} 主角名/數值替換 |
| DEBUG 口 | descent/warp/party/event/flag/item/gold(headless 腳本驗證,docs/46)|

## 剩餘工作(依「能破關」優先序)

### A. 能破關的硬 blocker(劇情進度系統)

1. **scripted-event 觸發系統(runner 0xabb2 / `[0x722]`)**:多數劇情 gate(阿里阿罕首旅の扉、
   下降觸發、取船、各 boss 後事件)由劇情旗標驅動的 runner 事件觸發。**event id `[0x722]` 資料驅動、
   無靜態 setter**(docs/31/44)→ 純靜態追不到,需動態(DOSBox debugger,本環境無)或逐 event 反推。
   handler 多數已 RE(0x3baa 表),缺的是「何時/由什麼旗標觸發哪個 event」。
2. ~~**劇情旗標推進**~~ **(#1 已落地)**:remake 自有的主線旗標流已建——`dq3_progress`
   (include/src,里程碑 0x200..0x208,順序取自杜勝利攻略 START→THIEF_KEY→MAGIC_BALL→
   ROMALY→DHAMA→SHIP→RAINBOW→DESCEND→ZOMA)。RAINBOW/DESCEND 鏡射既有 EXE 旗標
   (0x139/0x13a),故**合成/下降事件直接推進進度,不需另一套狀態**。debug 口 `prog`/`prog:N`。
   **真實 NPC 觸發已全接**(8/8 里程碑不再靠 prog:N):盜賊鑰匙(拿吉米之塔CTY8 4F老人)、
   魔法玉(雷貝CTY1 2F老人,需鑰匙)、羅馬利亞(CTY2國王持皇冠歸還,皇冠在香巴尼塔CTY10)、
   達瑪(CTY49神殿進城)、取船(波魯多加CTY37國王獻胡椒)、彩虹(合成)、下降、索瑪(終戰)。
   CTY↔城名對照見 docs/maps/cty-names.md。
   > CTY 對照(world_con_cty.png):0=阿里阿罕 1=雷貝 2=羅馬利亞 3=卡薩布 4=諾阿尼魯村
   > 5=精靈之村 7=岬之洞窟 8=拿吉米之塔 10=香巴尼塔(甘達特/金皇冠)11=精靈村西南洞窟
   > 15=賣胡椒城 37=波魯多加(取船)。
3. ~~**船**~~ **(#2 已落地)**:`dq3_ship`(include/src + test_ship 17 斷言)。海 tile 辨識 RE 完成
   ——overworld attr 指紋 `(attr&1)&&(attr&0x20)`(海 0x25,山 0x07 無 0x20),docs/48。
   登船/航行/上岸狀態機 + 船上不遇敵 + 船狀態併入 dq3_save(v3)。debug 口 `ship`/`ship:X:Y`/`opos:X:Y`;
   playthrough_check 驗證跨海航行與上岸。**取船劇情已接**(docs/50):波魯多加=CTY37,國王(9,6)
   sub2 NPC,獻黑胡椒(0x5c)→ 授船 + SHIP 里程碑 + 停泊 (25,73)。**剩**:黑胡椒的取得 NPC(商人)待接。
4. ~~**終盤**~~ **(已落地)**:大魔王索瑪戰(怪 0x7c)+ 結局序列。`zoma` debug 跑真實索瑪戰
   (修了 boss 大圖 sprite:MAXW 160→384,讓索瑪 384×144 / 0x7d·0x7f 大圖可繪,仍擋 0x80/0x81 垃圾 header);
   勝利 → `run_finale` 設 ZOMA 里程碑(進度 9/9)+ ENDTXT 結局首段。`finale` debug 直接驗破關路徑。
   mainline_check 已一條龍推到 **9/9 = 全主線完成**。剩:索瑪兩階段(光之玉)+ 完整結局捲動。

> 現實路徑:A1/A2 是同一塊(旗標驅動的 scripted-event VM)。**最高槓桿但最難純靜態**;
> debug 口已能「跳過劇情直接到任一狀態」驗證 —— 可先用 debug 口把各段接成可玩,再補真實觸發。

### B. 完整還原(faithful,非破關必須)

5. ~~**狀態效果**~~ **(部分落地)**:member.status bitfield(中毒/麻痺)。中毒 overworld 走路扣 HP
   (不致死留 1);驅毒草(0x42)解毒、滿月草(0x45)解麻痺、教會解異常;debug 口 `status:slot:bit`。
   存檔 v4。**剩**:戰鬥中怪物施加狀態 + 麻痺戰鬥不能行動(戰鬥系統整合)。
6. **NPC 互動完整化**:cmd-menu「話す」走 byte4 對話(目前部分路徑用舊 demo);子型2 條件對話的
   旗標分支(目前只顯示主對話)。
7. **overworld 旗標 portal 全表**:目前接 3 個寫死座標(docs/45 §3.2),完整需抽 0x396e 全部分支。
8. **scripted warp 全接**:§5b 的 8 個 0xd1f9 warp(洞穴→城)NPC 觸發 + dest≥100 overworld 回程。
9. ~~**道具效果**~~ **(#3 已落地)**:`dq3_item_use`(+test 12 斷言)。消耗品 id 由商店價
   交叉驗證鎖定(docs/49);藥草治 HP(持久 HP,封頂/陣亡保護)、聖水驅弱敵、蓋美拉翅膀回地表
   已生效;解毒/解麻痺待狀態系統(#5)。debug 口 `use:N`/`hurt:N`;playthrough +3 斷言。
   **剩**:野外指令窗「つかう」互動選取 UI(效果引擎已備)。

### C. Polish

10. 寶箱開過 tile 翻面(取後外觀變空)· 旅社/教會精確收費公式 · 同城多攤逐攤化(資料已備)。
11. 戰鬥逃跑/道具指令完整 · 咒文全效果 · 轉職(達瑪)實際換職。

## 早期 build 的「道具來源斷鏈」(已查證 pattern)

精訊版多個劇情物**有用途/有反應 NPC,但資料裡沒接上取得管道**——逐一靜態查證:

| 道具 | 用途 | 反應 NPC | 取得來源 | 結論 |
|---|---|---|---|---|
| 黑胡椒 0x5c | 獻波魯多加國王換船 | CTY15「店裡有賣」+ 國王「等胡椒」 | 無店進貨/無寶箱 | 斷鏈 → remake 補進 CTY15 道具店(docs/50)|
| ~~盜賊鑰匙 0x55~~ **(更正:非斷鏈)** | 開 tier1 鎖門 | CTY1 rec58「拿到了嗎」| **拿吉米之塔(CTY8)4F 老人給** | 已接 NPC 觸發 |

> 查證法(rule 62 靜態反追溯):寶箱表、真實 facility 抽取、EXE immediate、對話 glyph 錨點。
> **教訓(盜賊鑰匙)**:我一度誤判鑰匙「斷鏈」,因為(a)信了不完整的 `dq3_treasures` 抽取表
> 而非遊戲即時讀的 CTY 事件;(b)沒查攻略就下結論。**正解靠攻略**:盜賊鑰匙在**拿吉米之塔
> (CTY8)4F 老人** sub2 對話給(scripted-event,非 examine、非寶箱)。⇒ 查道具來源時:
> **先查攻略**(references/dq3_bbs)、dungeon 找 world map 結構(別在城裡找塔)、sub2 scripted NPC
> 也是給道具的合法管道(treasure 抽取涵蓋不到)。

## 單一最大 blocker

**劇情旗標驅動的 scripted-event 觸發系統(A1/A2)**——它是「把已做好的各系統串成一條可破關主線」
的連接組織。handler 都在,缺觸發。純靜態已到極限(`[0x722]` 無 setter);**務實做法**:
用 debug 口逐段驗證 + 設計 remake 自有的旗標流(不必 1:1 還原 runner VM),讓主線跑得通。

## 驗證方法(已就緒)

- 單元測試:stats/combat/battle/dialogue/npc/door/inventory/roster/menu/save(全綠)。
- DEBUG 口 + DQ3_DUMP:任意狀態 headless 截圖(docs/46)。全 89 城 load 掃描零崩潰。
- **`tools/playthrough_check.sh`(14 項)**:各系統孤立斷言(載入/對話/設施/隊伍/scripted event/
  傳送/進度旗標流/船/道具)全綠。
- **`tools/mainline_check.sh`(#4)**:單一進程依攻略順序走 START→…→索瑪,驗進度階段**單調推進
  1→9** + 四真實 gate(取船/彩虹合成/下降/索瑪破關)依序生效 → **主線一條龍推到 9/9 = 能破關**。
