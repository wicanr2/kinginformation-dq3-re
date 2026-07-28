# FLOW-GAP:原版流程 Oracle × ebitan(dq3_remake_ebitan)差距清單

> **狀態:完成(A-E 段 + 總結已全數寫入)**
> 對照基準:`docs/66-original-flow-oracle.md`(原版 56 步 + §1 開場 + §3 已知 bug)。
> ebitan = `dq3_remake_ebitan/`(Go/Ebiten 版,音樂/戰鬥/地圖引擎已有一定基礎但流程與原版不同)。
> C spec 指標 = `dq3_remake/src/main.c`(既有 C RE 版,行號供交叉核對,不代表 ebitan 有對應實作)。
> 修法:`port-C`=把 main.c 已驗證邏輯搬過去、`new-RE`=兩邊都缺需重新 RE、`wire`=資料已在但未接線、`ok`=已對齊。

## 表頭

`GAP# | oracle 步驟 | 原版行為 | ebitan 現況(檔:行 或 grep 0 命中) | C 有無(行號) | 嚴重度 A/B/C | 修法`

---

## A. 開場序列(對照 §1)

| GAP# | oracle 步驟 | 原版行為 | ebitan 現況 | C 有無 | 嚴重度 | 修法 |
|---|---|---|---|---|---|---|
| A1 | §1-1 開機 cutscene | 單色巨龍圖 cutscene → 標題 | grep 0 命中 "cutscene";僅 `showTitle`+`titlePix`(TITG.P 標題本身)game.go:1313 | 未查 | C | new-RE |
| A2 | §1-2 主選單(遊戲開始/載入進度) | 「▶遊戲開始 / 載入進度」選項 | 無選單;A/Enter 直接開始,`_ = g.Load()`(game.go:1259)自動續玩取代選單分支 | 未查 | C | new-RE/wire |
| A3 | §1-3/4 新遊戲注音創角(**主角**) | 5×9 注音鍵盤+英數,必填姓名 | grep 0 命中 "heroName/主角命名/protagonist";`zhuyin.Composer`+英數格盤元件已存在但只接在 `tavern.go`(招募同伴用),主角無命名 UI 入口 | main.c 有 `#include "dq3_nameinput.h"`(main.c:29),未深查是否已接主角 | B | wire(元件現成,缺主角專用 instance) |
| A4 | §1-5 性別選擇(**主角**) | 命名後選單「▶男性/女性」 | `tavGender` 僅供 tavern 招募同伴;`Member.Gender` 欄位存在但 `NewGame` 未見對主角設定性別分支 | 未查 | B | wire |
| A5 | §1-6 開場劇情(16 歲生日/母親進城見國王) | 黑底對話框旁白 | grep 0 命中 "開場/生日/母親";`progressSet(msStart)`(game.go:1294)只設旗標,無劇情文字播放 | grep 0 命中「生日/king_audience」于 main.c | C | new-RE(對話文本待抽 D3TXT) |
| A6 | §1-7 遊戲開始出生點 | 阿里阿罕(CTY00)故鄉室內房間 | `g.px,g.py = g.over.w/2, g.over.h/2`(game.go:1290)= **地表中心**,非城鎮室內 | 未查 main.c 出生點對應 | **A** | port-C/new-RE |
| A7 | §1-8 隊伍組成(開場 1 人,酒場另募≤3 人) | 主角單人起步,玩家在酒場逐一創角招募 | `startingCompanions(0)`(member.go)**硬編預建**戰士/僧侶/魔法使 3 人 demo 隊,略過玩家選擇;`tavern.go` 招募 UI 完整(職業→注音/英數命名→性別)但正式入口未接,僅 `T` 鍵 dev 捷徑(game.go:487 註記「正式入口=LUIDA 櫃台」尚未做) | main.c 有 `tavern_modal` 正式接線(main.c:1497 酒館/1885 二次呼叫) | **B** | port-C(NPC 觸發)+ wire(移除 demo 隊、改玩家招募) |
| A8 | §1-9 初始金錢/裝備 | 「國王給了金錢和裝備」(具體數字 oracle 本身待核對,見 oracle §4-1) | 硬編 `heroGold=120`、`equip=[3,0x21]`(銅劍+皮甲冑)game.go:1290-1291,無出處註記 | 未查 | C | 待人工核(oracle 本身未定案,無法判 ebitan 對錯) |
| A9 | §1-10 酒場/登錄所精確位置 | oracle 自陳待核對 | 不評(不屬 ebitan 差距) | — | — | 不適用 |
| A10 | §1b attract 職業巡禮演出 | 標題循環播放 5+ 職業立繪+能力條捲動 | grep 0 命中 "attract";素材已定位(docs/67,TIT*.P 20 檔)但無播放邏輯 | 未查 | C | new-RE(素材就緒,缺實作) |

**A 段小結**:最嚴重是 A6(出生點錯位,一開局就偏離原版)與 A7(主角隊伍創建被 demo 資料取代,酒場招募 UI 做好了但沒接线口)。A3/A4(主角命名/性別)是明顯的忠實度缺口但不阻斷遊玩。

---

## B. 主線前段(步 1–20,對照 §2)

> 交叉驗證方法:用 python 對 `game/treasure.go`(121 筆 baked 表)按 item code 反查是否收錄
> oracle 步驟所需道具(見下方「治具查核」)。`scriptedTable`(NPC 給物,byte4+cty 索引)只有 **10 筆**,
> ebitan 與 C(`dq3_scripted.c`)筆數相同 —— 不是 ebitan 特有落後,是兩邊都尚淺;真正落差在
> main.c 有 `region 0x722` runner 事件系統(grep 命中 main.c)承接大部分「NPC 對話/條件給物」邏輯,
> ebitan **完全沒有**這層(grep "runner"/"region722"/"0x722" 0 命中)。

| GAP# | oracle 步驟 | 原版行為 | ebitan 現況(檔:行 或 grep 0 命中) | C 有無 | 嚴重度 | 修法 |
|---|---|---|---|---|---|---|
| B1 | 步1 王城謁見+聞父失蹤+酒場創角 | NPC 對話揭露父親歐里狄加失蹤劇情 | 對話文字系統(`D3TXT`)已載,但王城謁見**觸發**(examine/NPC 對話鏈)未見獨立實作;見 A5/A7 | main.c 有 0x722 系統可能承接,未逐一核 | B | new-RE/port-C |
| B2 | 步2 拿吉米之塔取盜賊鑰匙(0x55) | 4F 老人巴可達給鑰匙 | treasure.go **0 命中** item=0x55(85);scriptedTable(10 筆)未逐一核對是否含此列 | main.c 命中 0x55 相關邏輯(50 筆道具碼命中之一,未逐一分離) | **A**(門戶道具,無此無法開紅門外全部門) | port-C/new-RE |
| B3 | 步3 城內開門(持鑰匙) | `keyTier()` 依 0x55/0x56/0x57 開對應鎖門 | **已有**:`g.keyTier()`(game.go)+ `tryOpenFacingDoor`(game.go)機制通用,鑰匙判定邏輯完整 | — | ok(機制在,缺 B2 供應鑰匙則空轉) | ok |
| B4 | 步4 雷貝鎮取魔法球(0x58) | 2F 老人給球 | treasure.go **0 命中** item=0x58(88) | main.c 命中(同上) | **A** | port-C/new-RE |
| B5 | 步5 誘人洞窟通行+見羅馬利亞王 | 魔法球破牆 + 王命找金皇冠 | `enterTownCty(cty=2)` 進城自動 `progressSet(msRomaly)`(game.go:673-675)——**無條件檢查**是否持 0x58,也無國王對話文字,純進城即觸發 | main.c `dq3_progress_set(&flags, DQ3_MS_ROMALY)` 有等效行(main.c:1638-1639),需查是否也無條件 | C(簡化:少了持球前置判定與王對話) | port-C(若C有gate)/new-RE(對話) |
| B6 | 步6 香巴尼塔擊敗甘達特取金皇冠(0x33) | boss 戰(打甘達特怪 26 才放行寶箱,main.c RE 已確認 gate) | 金皇冠 **treasure.go 有收錄** `{10,5,0,1,97}`(cty10 sec5 sub0 flag97,item 0x33)但為**靜態寶箱**,無擊敗甘達特(monster 26)的 boss 戰前置;examine() 全函式(game.go:342-420)只處理寶箱/warp/門,**無任何座標觸發的怪物戰**(除 Zoma 神殿連戰,見 D 段) | main.c 有「打甘達特怪 26 才放行」gate(oracle 出處欄引註) | **B**(寶箱可拿但缺 boss 戰本體與 gate) | port-C |
| B7 | 步7-8 諾阿尼魯全村沉睡 + 精靈之村線索 | NPC 對話揭露支線 | 對話系統存在,但沉睡/催眠**狀態**顯示與觸發鏈未見(itemuse.go 註解明寫「解狀態(無狀態系統)」) | 未查 | C | new-RE |
| B8 | 步9 地底湖洞窟取紅寶石(0x59) | B4F 寶箱 | treasure.go **有收錄** `{11,3,0,3,47}`(item 0x59) | — | ok(寶箱端) | ok |
| B9 | 步10 精靈之村還寶石換覺醒粉(0x59→0x5a) | 交還紅寶石,NPC 給覺醒粉 | grep "0x59\|0x5a" 僅 `itemuse.go` 的 **USE** 端(Awaken kind)有 0x5a;**無**任何程式碼把 0x59 換成 0x5a 的 NPC 交換邏輯(0 命中轉換) | 未查main.c 轉換點 | **B**(玩家背包永遠不會拿到 0x5a,除非直接塞測資) | port-C/new-RE |
| B10 | 步11 覺醒粉解諾阿尼魯催眠 | USE 0x5a 於 CTY4 | **已接**:`itemuse.go` case `Awaken`,`g.curCty==4` 才生效,設 flag 0x31(itemuse.go:39-46);oracle RE 佐證 main.c 同款邏輯 | main.c 已驗證(oracle B-11 引註) | ok | ok |
| B11 | 步12 阿莎拉慕→沙漠祠堂→依席斯城線索 | NPC 對話鏈 | 未見獨立實作(對話系統通用,無此鏈證據) | 未查 | C | new-RE |
| B12 | 步13 依席斯隱藏通道取流星手環(0x49) | 寶箱(需先找到隱藏通道) | treasure.go **有收錄**(3 筆候選之一,cty 座標含 32=依席斯附近);隱藏通道**開啟機制**(牆體變路)未見獨立驗證 | 未查 | C(寶箱端 ok,通道觸發待核) | 待人工核 |
| B13 | 步14 金字塔按鈕開門取魔法鑰匙(0x56) | 依序按鈕(先東後西)開石門+寶箱 | treasure.go **有收錄** `{13,2,1,1,115}`;**按鈕順序判定邏輯**(GH 機關)grep 0 命中 "button/按鈕/機關" | 未查 | B(寶箱端 ok,機關 gate 未見) | 待人工核 |
| B14 | 步15 依席斯紅門+羅馬利亞祠堂取祈禱戒指(0x48,第一枚) | 持魔法鑰匙開門 | treasure.go 0x48 有 **7 筆**(多處掉落點,含 cty0 疑似 demo 資料混入),鑰匙門機制見 B3 | — | ok(鑰匙+寶箱機制通用) | ok |
| B15 | 步16 波魯多加王:給國王的信(0x5b)+ 承認勇者身分 | 首訪國王對話+給信 | treasure.go **0 命中** 0x5b(NPC 給物非寶箱);`enterTownCty` 對波魯多加(需查 cty id)無 milestone 分支(只有 cty2/49 兩個 case) | main.c 有 flag 0x215 gate(oracle B-16 引註) | **A**(國王對話與信件是後續東方支線的閘門) | port-C/new-RE |
| B16 | 步17 諾魯特收信通隱密通道 | 檢查持 0x5b | grep 0 命中 "諾魯特/0x5b 檢查" 於 ebitan;無對應 NPC handler | main.c 有 handler 0x5daf 檢查 0x5b(oracle 引註) | **A** | port-C |
| B17 | 步18-19 巴哈拉達鎮線索+救古布達/達妮亞+二戰甘達特 | 對話+牢房開關+ boss 戰 | 無牢房開關邏輯、無二戰甘達特座標觸發(同 B6 boss 觸發空缺) | boss-trigger-points.md 高信度 RE(main.c 對應) | **A** | port-C |
| B18 | 步20 回胡椒店取黑胡椒(0x5c) | flag 0x211(救人完成)才給 | treasure.go **0 命中** 0x5c;flag 0x211 grep 0 命中於 ebitan | main.c 有 quest-items gate | **A** | port-C |

**B 段小結**:模式很清楚 —— **靜態寶箱**(B3/B8/B10/B12-14/B14 鑰匙門)大多已由 `treasure.go`(baked 121 筆)+ `keyTier()` 通用機制覆蓋(ok 或簡化);但 **NPC 給物 / boss 戰觸發 / 對話 gate**(B2/B4/B6/B9/B15-18)幾乎全部缺席 —— 根因是 ebitan 完全沒有 main.c 已具備的 `region 0x722` runner 事件系統,只有 121 筆靜態寶箱表 + 10 筆極簡 scripted 表可用。前段 20 步中 **6 項判 A(阻斷/靠 debug)、5 項判 B、多項判 C**。

---

## C. 主線中段(步 21–40)

| GAP# | oracle 步驟 | 原版行為 | ebitan 現況 | C 有無 | 嚴重度 | 修法 |
|---|---|---|---|---|---|---|
| C1 | 步21 交黑胡椒換船+取領悟之書(0x4a) | NPC 對話換船;塔內寶箱書 | **船完全靠 debug**:`shipOwned` 唯一賦值處是 `DQ3_SHIP` 環境變數(game.go:1353-1357),無正常玩法路徑取得;領悟之書 treasure.go **有收錄**`{18,1,0,1,183}` | main.c `DQ3_MS_SHIP` 有正常流程設定(main.c:1618/1139/1143) | **A**(核心里程碑,無船則後續一半地圖到不了) | port-C |
| C2 | 步22 提頓村白天取黑暗之燈(0x5f) | 寶箱 | treasure.go **有收錄** `{20,1,0,1,57}`;晝夜判定見系統層 E | — | ok(寶箱端) | ok |
| C3 | 步23 八頭大蛇洞窟擊敗蛇取草薙大劍(0x14)+女鬼面具(0x38) | boss 戰勝出寶箱 | 草薙大劍 **treasure.go 0 命中**(疑為 boss 戰勝利獎勵非寶箱);女鬼面具 **有收錄**`{19,0,0,1,138}`;boss 觸發座標(35,12)無座標式怪物戰(同 B6 模式) | boss-trigger-points.md CTY19 高信度 | **A** | port-C(boss 觸發)+ 待人工核(0x14 給予方式) |
| C4 | 步24 日邦格村無姬大人(八頭大蛇)二戰取紫寶珠(0x69) | 傳送房間可選戰/不戰 | 紫寶珠 treasure.go **有收錄**`{21,0,0,1,137}`(疑為寶箱形式而非戰勝獎勵,與原版「戰勝取得」語意不符,待核) | 未查 | 待人工核 | 待人工核 |
| C5 | 步25 朗錫爾道具店買隱身草(0x5d) | 商店購買 | treasure.go **0 命中**(合理,非寶箱);`shopItemPool`/貨架資料未逐一核對是否含 0x5d | 未查 | C(商店機制通用存在,商品清單未核) | 待人工核 |
| C6 | 步26 隱身進耶進貝亞+倉庫番推三石取乾渴壺(0x5e) | USE 隱身草隱形 + 推石解謎 | grep "隱身/invisib" **0 命中**——無隱身效果;grep "sokoban/倉庫番/推石" **0 命中**——無推箱解謎;乾渴壺本體 treasure.go **有收錄**`{76,0,0/1,1,58/59}`(2 筆,疑為解謎後才可拿的寶箱,但無解謎邏輯則永遠可直接撿) | main.c 有完整實作:`g_sokoban_solved`(main.c:348/1280/1387/1847/2010-2025) | **A**(整個機制未搬,寶箱變成無條件可拿=忠實度反向錯誤) | port-C |
| C7 | 步27 最終鑰匙祠堂 USE 乾渴壺吸海取最終鑰匙(0x57) | 四島礁 USE 顯現祠堂 | **已接**:`itemuse.go` case `Drain`,`!g.inTown` 才生效,設 flag 0x33(itemuse.go:52-56);最終鑰匙 treasure.go 收錄`{40,0,0,1,71}` | main.c 有等效 examine gate(main.c:1847 `ep2==0x5e`) | ok(USE 端)/簡化(缺乾渴壺本身取得前置 C6) | port-C(補 C6 前置) |
| C8 | 步28 提頓村夜晚開牢門取綠寶珠(0x66) | 夜晚+最終鑰匙 → 犯人給珠 | treasure.go **0 命中** 0x66(NPC 給物,非寶箱);夜晚判定見系統層 E(晝夜系統缺席) | main.c 有 CTY20 夜 gate(oracle 引註) | **A** | port-C(含晝夜系統) |
| C9 | 步29 朗錫爾神殿神父挑戰(flag 0x13) | 持最終鑰匙開鐵門+對話挑戰 | grep "0x13\b" **0 命中**於 ebitan | main.c dispatch 全解(main.c CTY47 handler) | **A** | port-C |
| C10 | 步30 勇氣洞窟單人取藍寶珠(0x67) | 需 flag 0x13,強行忽略警告 | 藍寶珠 treasure.go **有收錄** `{23,2,0,1,173}`(靜態寶箱,無單人/flag 前置檢查) | — | B(寶箱端 ok,前置 gate 缺) | port-C |
| C11 | 步31 亞布之塔取回音之笛(0x60) | 3F 寶箱 | treasure.go **有收錄** `{26,2,1,1,176}` | — | ok | ok |
| C12 | 步32 海盜村密道取紅寶珠(0x68) | 移石現密道+藏寶室 | treasure.go **有收錄** `{27,1,0,1,63}`(靜態,移石機關未見獨立驗證) | — | C(寶箱端 ok,機關待核) | 待人工核 |
| C13 | 步33/37/41 新城鎮商人建城支線(需商人職業同伴) | NPC 檢查隊伍 class==6 | grep "class.*6/商人.*建城/newTown" **0 命中**——**整條支線不存在**;`startingCompanions` 固定戰士/僧侶/魔法使,無商人可用 | main.c 老人 handler 0x5aba 全解(oracle B-6 引註) | **A**(職業系統雖有 8 職業但流程未接) | port-C |
| C14 | 步34 斯吾村水井調查取雷神之杖(0x0a) | examine 井 | treasure.go **有收錄** `{38,0,0,1,149}` | — | ok | ok |
| C15 | 步35 沙曼歐莎城救真國王+拉之鏡(0x61) | 最終鑰匙開牢門+南方洞窟 B3F 寶箱 | 拉之鏡 treasure.go **有收錄** `{24,2,0,1,159}`(靜態);救國王對話鏈未見 | — | C(寶箱端 ok,劇情鏈待核) | 待人工核 |
| C16 | 步35b 夜晚照假國王現形怪力魔(0x89)取變身杖(0x62) | boss 戰(持拉之鏡+夜晚觸發) | 變身杖 treasure.go **有收錄 4 筆**(cty0/13/25/80,靜態寶箱形式);**無**「持鏡+夜晚 examine 觸發變身怪 boss」邏輯(同 B6/C3 模式,examine() 無怪物 dispatch) | — | **A**(boss 觸發缺) | port-C |
| C17 | 步36 新城鎮已建成(承接 C13) | 有商店可用 | 依附 C13,C13 缺則此步無意義 | — | A(連鎖) | port-C |
| C18 | 步37 雪地草原老人交變身杖換船員之骨(0x63) | NPC 交換 | treasure.go **0 命中** 0x63(NPC 給物) | — | **A** | port-C |
| C19 | 步38 幽靈船 USE 船員之骨尋路取愛的回憶(0x64) | 反覆 USE 依方向指示 | `itemuse.KindOf` **無** 0x63 case(grep 確認 itemuse.go 只有 Awaken/Gaia/Drain/FairyFlute/Rainbow 5 種位置相關 USE);愛的回憶 treasure.go 收錄(靜態,`{0,5,2,1,41}`+`{36,1,0,1,139}`) | — | **A**(USE 導航邏輯整條缺) | port-C |
| C20 | 步39 牢獄祠堂持愛的回憶解幽靈詛咒取蓋亞之劍(0x0f) | NPC 對話相認 | 蓋亞之劍 treasure.go 收錄 `{55,0,0,1,72}`;相認對話鏈未見 | — | C(寶箱端 ok) | 待人工核 |
| C21 | 步40 USE 蓋亞之劍開火山取銀寶珠(0x6b) | 位置相關 USE | **已接**:`itemuse.go` case `Gaia`,`!g.inTown` 才生效,設 flag 0x32(itemuse.go:47-51);銀寶珠(0x6b)treasure.go **0 命中**(疑為 examine 後直接給予,非寶箱表) | — | 部分 ok(USE 端)/待人工核(珠子給予路徑) | 待人工核 |

**C 段小結**:延續 B 段模式 —— 位置相關 **USE 型**劇情道具(乾渴壺/蓋亞之劍等 5 種)全部已在 `itemuse.go` 接好且與 oracle RE 對齊(ok);**靜態寶箱**(領悟之書/黑暗之燈/女鬼面具/藍寶珠/回音之笛/紅寶珠/拉之鏡/變身杖/愛的回憶/蓋亞之劍等)幾乎全在 `treasure.go` 121 筆表裡(ok 或簡化);但凡涉及 **NPC 交換給物**(船/船員之骨/綠寶珠)或 **座標觸發 boss 戰**(八頭大蛇/怪力魔/無姬大人)或 **多步解謎**(倉庫番/推石機關)的步驟,**全數缺席**,其中「取船」(C1)與「幽靈船尋路」(C19)是本段最嚴重的兩處斷點——前者卡住地圖過半可達性,後者卡住整條 39-40 道具鏈。

---

## D. 主線後段(步 41–56)

| GAP# | oracle 步驟 | 原版行為 | ebitan 現況 | C 有無 | 嚴重度 | 修法 |
|---|---|---|---|---|---|---|
| D1 | 步41 新城鎮商人被囚取黃寶珠(0x6a) | 依附 C13(新城鎮支線) | 黃寶珠 treasure.go **有收錄**`{83,0,2,1,74}`(靜態);支線本體同 C13 缺席 | — | A(連鎖 C13) | port-C |
| D2 | 步42 不死鳥祠堂集六珠復活拉米亞坐騎 | 祭壇 examine,六珠(0x66-0x6b)判定+對話 | **2026-07-28 R-3 已完成：CTY70 六祭壇、handler69、動畫、走上搭乘、Confirm 降落、save**（見 docs/75） | EXE `0x7425..0x7599` | ✅ 事件 E3 | 完成 |
| D3 | 步43 乘坐騎見龍之女王取光之珠(0x65) | 依附 D2 | 光之珠 treasure.go **0 命中**(NPC 給物);依附 D2 缺席 | — | A(連鎖) | port-C |
| D4 | 步44 巴拉摩斯城取魔神之斧(0x19)+祈禱戒指(第二枚)+擊敗巴拉摩斯 | 寶箱+boss(須飛行抵達,打兩次) | 魔神之斧 treasure.go **有收錄**`{66,0,2,1,153}`(靜態,寶箱端 ok);boss 座標觸發同 B6/C3/C16 模式**缺席**;無飛行前置(依附 D2) | walkthrough-flow-audit 引 main.c token | A(boss+飛行連鎖) | port-C |
| D5 | 步45 阿里阿罕回報+索瑪聲音+flag 0x213 | 王城對話 | grep "0x213" **0 命中**於 ebitan | — | A | port-C/new-RE |
| D6 | 步46 蓋亞那洞穴跳坑下降愛列夫加特 | 需 flag 0x213,跳坑觸發 `do_descent` | `descend()` **已有函式**(game.go:804-813)但觸發僅 `U` 鍵 dev 捷徑(game.go:493-497 明註「正式觸發=劇情事件 86」未接),且無 flag 0x213 前置檢查 | — | **B**(函式已在,缺正式觸發+前置 gate) | wire/port-C |
| D7 | 步47 拉達多姆城取太陽之石(0x72) | 廚房密道 2F 寶箱 | treasure.go **有收錄**`{80,2,0,1,194}` | — | ok | ok |
| D8 | 步48 洛特洞窟取勇者之盾(0x40)+生命果實(0x4f) | B2-B3F 寶箱 | 兩者 treasure.go **均有收錄**(0x40:1 筆,0x4f:8 筆散佈全圖) | — | ok | ok |
| D9 | 步49 加萊祠堂密門取銀豎琴(0x71) | 密門+地下室寶箱 | treasure.go **有收錄**`{78,1,0,1,154}`(密門機關未見獨立驗證) | — | C(寶箱端 ok) | 待人工核 |
| D10 | 步50 達姆杜拉養羊圍欄取歐里空金屬(0x6d) | examine | treasure.go **有收錄**`{84,0,0,1,195}` | — | ok | ok |
| D11 | 步51 利姆達爾旅館取生命戒指(0x70) | 寶箱 | treasure.go **有收錄**`{86,0,0,1,212}` | — | ok | ok |
| D12 | 步52 瑪依拉:賣歐里空金屬(22500)買王者之劍(35000)取妖精之笛(0x77) | 商店賣+買 | 妖精之笛寶箱 ok(`{81,0,0,1,155}`);**商店無賣出功能**(grep "sell/Sell" 全專案 **0 命中**,`Shop.input` 只回 `buyCode`,無賣出分支);王者之劍(0x1c)treasure.go/scriptedTable **均 0 命中** | walkthrough-flow-audit A-4「main.c:1247 transform NPC 已驗證」 | **A**(賣店機制整個不存在,王者之劍拿不到) | port-C |
| D13 | 步53 魯比斯之塔 USE 妖精之笛解咒取精靈的守護(0x74) | 位置相關 USE | **已接**:`itemuse.go` case `FairyFlute`,`g.curCty==82` 才生效,直接 append 0x74(itemuse.go:57-65) | oracle 對齊 `DQ3_USE_FAIRYFLUTE` | ok | ok |
| D14 | 步54 精靈祠堂交精靈的守護換雲雨之杖(0x73) | NPC 交換 | treasure.go/itemuse **均 0 命中** 0x73(NPC 給物) | walkthrough-flow-audit B-6「main.c 已解」 | **A** | port-C |
| D15 | 步55 神聖祠堂合成彩虹水滴(0x75) | 持雲雨之杖+太陽之石 examine 合成(原版 bug 已知,remake 應修正) | **已接且已修正**:`synth.go` `synthRainbowAtShrine()`,`g.curCty==shrineCty` 才觸發,消耗太陽之石+雲雨之杖→給 0x75(非原版 bug 的 0x6b),設 `msRainbow` | 對齊 oracle §3 #2(remake 已修) | ok(邏輯本身),但**前置 D14 缺**(拿不到雲雨之杖則此步無法觸發) | port-C(補 D14) |
| D16 | 步56a 龍王城彩虹架橋+擊敗六大魔人(怪106×6) | USE 彩虹水滴架橋+boss連戰 | `itemuse.go` case `Rainbow` 已接位置相關 USE(itemuse.go:66-74);**六大魔人連戰本體有實作**(`startZomaSeq`/`bossQueue`,game.go:919-943)但**唯一觸發入口是 `Z` 鍵 dev 捷徑**(game.go:498),無座標 examine 觸發 | — | **A**(邏輯做了,正式入口沒接) | wire |
| D17 | 步56b 龍王城密道目睹父親戰死(過場) | 隱藏樓梯+過場劇情 | grep "歐里狄加/0x80/0x81" **0 命中**於 ebitan;oracle §3 #3 記錄原版此處**本身會當機**(sprite 缺失),C 已修 3-byte patch,ebitan 未見對應處理 | docs/18 反組譯定位 + C patch | B(原版已知 bug 的修正版本未搬) | port-C |
| D18 | 步56c 寶物室開寶箱群 | 多件散裝寶箱 | 龍王城(cty 待核)相關寶箱是否落在 treasure.go 121 筆內未逐一核對(過場依附 D16/D17,連鎖受阻) | — | 待人工核 | 待人工核 |
| D19 | 步56d 索瑪神殿三連戰(怨靈122→殭屍123→索瑪124,光之珠弱化) | 依序戰 | **已接**:`bossQueue` 序列含 122/123/124(game.go:927),`startBossBattle` 對 0x7c 特判 `lightOrb` 弱化(game.go:899-908,battle.go:136-139);同 D16,入口靠 `Z` 鍵 | walkthrough-flow-audit C-9 對齊 main.c 序列 | 邏輯 ok / **入口 A**(靠 debug) | wire |
| D20 | 步56e 回拉達多姆見國王冊封「勇者洛特」+結局捲動 | 國王對話+結局文字 | **已接**:`runFinale()`(game.go:953-965)設 `msZoma`+`cleared`+`lotoBlessed`+ ENDTXT 結局捲動(`endSeq`);對話觸發仍依附 D19(只能靠 Z 鍵鏈路走到) | main.c `dq3_progress_set(&flags, DQ3_MS_ZOMA)`(main.c:572) | 邏輯 ok / 入口依附 D16-19 | wire |

**D 段小結**:後段最大斷點是 **D2(不死鳥坐騎機制完全不存在)**——它是步 43-44 的飛行前置,少了它整個後段地圖到不了,連鎖拖垮 D3/D4;第二大斷點是 **D12(商店無賣出功能)**,王者之劍拿不到但屬支線裝備非阻斷。反而 **終盤 boss 連戰邏輯本身相當完整**(D16/D19/D20 三者的判定/弱化/結局文字都已移植對齊 main.c),只是缺一個「正式從龍王城/索瑪神殿走進去」的入口(目前全靠 `Z` 鍵),屬於「邏輯做完,只差最後一段接線」的 wire 類問題,修復成本應遠低於 D2。

---

## E. 系統層(晝夜/遭遇/存檔/戰鬥/結局等橫切面)

| GAP# | 項目 | 原版行為 | ebitan 現況 | C 有無 | 嚴重度 | 修法 |
|---|---|---|---|---|---|---|
| E1 | 晝夜系統 | 全域晝夜相位,影響提頓村寶箱(步22 白天)/牢門(步28 夜)/沙曼歐莎(步35b 夜)等多處 gate | grep "night/晝夜" **僅有 1 處註解**(itemuse.go:8/76:「晝夜→不消耗」),**無晝夜狀態變數本身**、無日夜循環計時、`DarkLamp`(0x5f 黑暗之燈)case 直接 no-op | day-night-system.md(oracle 待核 §4-8 引註)顯示原版有此系統 | **A**(牽連 C2/C8/C16 等多步 gate 判定全失效或被繞過) | new-RE/port-C |
| E2 | 遭遇區表(regional encounter table) | 依地形/城鎮分區各有專屬怪物池 | `startEncounter()`(game.go:877-895)**固定單一弱怪池** `[]int{5,0,1,3,4,6,7,8}`,`rand.Perm` 隨機選,**全地圖同一套**,不分區域/城鎮/迷宮 | 未查 | **B**(戰鬥能玩但難度曲線/怪物種類與地點脫鉤,長尾忠實度) | new-RE/port-C |
| E3 | runner/region `0x722` 事件系統 | 57 個 setter,承接大部分「NPC 對話 gate/條件給物」邏輯(`re-log-722-state-machine.md`) | grep "runner/region722/0x722" **ebitan 0 命中**;main.c 有對應(`grep -l` 命中 main.c) | 高信度(main.c 已含) | **A**(是 B/C/D 段大量 A/B 判定的共同根因) | port-C(整套系統搬遷,非逐步驟零星修) |
| E4 | 四人縱列(caterpillar)隊伍繪製 | 主角+最多 3 同伴一列跟隨行走(docs/36 §1b 8:00 影格) | `renderFrame()`(game.go:1080-1103)**只畫 `g.hero` 單一 sprite**,`g.companions` 只用於戰鬥計算(`buildCompanionActors`)與存檔,地表/城鎮渲染完全不畫同伴 | 未查 | **B**(視覺忠實度,不阻斷) | new-RE/port-C |
| E5 | 王城謁見畫面(對照必查項) | 見 A5/A8 | 同 A5/A8 | — | 見 A | — |
| E6 | boss 觸發點(examine 格) | 全掃見 `boss-trigger-points.md`,多處座標 examine → 開戰 | `examine()`(game.go:342-380)**只處理寶箱(type 1/3)/type-2 warp/鎖門**,**無任何「examine 座標 → 開戰 monID」分支** | 高信度(main.c 對應多處) | **A**(系統性缺口,見 B6/C3/C16/D4 個案) | port-C |
| E7 | 六珠/不死鳥questline | 見 D2 | 同 D2 | — | 見 D | — |
| E8 | 船/下降/彩虹自然觸發 | 三處均應由劇情事件(給黑胡椒/跳坑/examine 合成)自然觸發,非玩家手動 debug | 三者現況:船(C1)= 100% debug-only;下降(D6)= 函式在但入口 debug-only;彩虹合成(D15)= **已正確走 examine 觸發**(非 debug) | — | 混合(A/B/ok 各一) | 見對應 GAP |
| E9 | 倉庫番(sokoban) | 見 C6 | 同 C6 | main.c 有完整 `g_sokoban_solved` 狀態機 | 見 C6 | port-C |
| E10 | 隱身草 | 見 C5/C6 | USE 效果(`itemuse.KindOf`)**無 0x5d case**(grep 確認 5 個位置相關 kind 不含隱身),隱身狀態本身也不存在 | — | **A** | new-RE/port-C |
| E11 | 拉之鏡鏈 | 見 C15/C16 | 寶箱端 ok,boss 觸發缺(同 E6) | — | 見 C16 | port-C |
| E12 | 變身杖鏈 | 見 C16/C18/C19;變身杖(0x62)取得 →交換船員之骨(0x63)→USE 尋路 | 三步串全斷:C16(boss 觸發缺)、C18(NPC 交換缺)、C19(USE 導航邏輯缺,`itemuse.KindOf` 無 0x63 case) | — | **A** | port-C |
| E13 | THE END 結局畫面 | 索瑪戰後回城冊封+結局捲動文字 | **邏輯已接**(`runFinale`/`endSeq`/`ENDTXT.TXT` 載入,game.go:953-971,1261),文字內容從原版 `ENDTXT.TXT` 讀出非新編;唯一問題是入口鏈路(依附 D16-D20,目前只能靠 `Z` 鍵走到) | 對齊 main.c `DQ3_MS_ZOMA`(main.c:572) | ok(邏輯與素材)/A(入口) | wire |
| E14 | 存檔系統(冒險之書) | 教會/記錄點存檔 | `save.go` 完整實作(137 行),JSON round-trip,`facChurch`/`facRecord` 觸發 `g.Save()`(facility.go 對應 case),欄位涵蓋 exp/hp/mp/gold/inventory/equip/companions/flags/ship/位置 | 非二進位相容(remake 自有格式,設計如此非缺陷) | ok | ok |
| E15 | 商店賣出功能 | 部分道具可變賣(如步52 賣歐里空金屬) | `Shop.input`(facility.go)**只回 `buyCode`,無賣出分支**,grep "sell/Sell" 全 repo 0 命中 | — | **A**(見 D12) | port-C |

**E 段小結**:系統層清楚指出兩個「一次修好、多處連帶回血」的高槓桿缺口——**E3(region 0x722 事件系統)**與 **E6(examine 座標式 boss 觸發)**。這兩者一旦補上,B/C/D 段裡標 A 的十餘個「NPC 給物」「boss 戰觸發」項目會有共同底層可以承接,不必逐步驟各自新寫。E1(晝夜)是第三個高槓桿缺口。相對地,E14(存檔)、E8 部分(彩虹合成)、D13/D15/D19/D20 等已展現 ebitan 在「有明確 RE 佐證的單點機制」上其實做得相當扎實——落差集中在「需要跨系統串接」的長鏈條劇情。

---

## 總結(回報用)

> 統計基準:A/B/C/D 四段共 69 列(逐 oracle 步驟/子步驟展開),E 段另計 9 筆系統層獨立項
> (E5/E7/E8/E9/E11/E12 為純交叉引用,已算入對應 A-D 段列,不重複計);扣除 A9(不適用,oracle 自身待核不評)。
> **GAP 總數 = 78**(含已對齊 ok 項,便於掌握全貌;純缺口數 = 78 − 15(ok) − 3(待人工核) − 1(不適用) = 59)。

### 嚴重度分布

| 嚴重度 | 數量 | 說明 |
|---|---|---|
| **A**(開場即遇/阻斷,或完全靠 debug 鍵) | **33** | 集中在:NPC 給物鏈(盜賊鑰匙/魔法球/國王的信/黑胡椒/船員之骨/雲雨之杖等)、座標式 boss 觸發(甘達特/八頭大蛇/怪力魔/巴拉摩斯等全數缺席)、三個核心里程碑(取船/不死鳥坐騎/晝夜系統)、終盤入口(六大魔人~索瑪連戰邏輯已做但只能靠 `Z` 鍵) |
| **B**(主線 gate 錯或靠 debug 鍵,較輕) | **12** | 主角命名/性別 UI 缺、酒場正式入口未接(僅 `T` 鍵)、寶箱前置 gate 缺(有寶箱但無 boss/機關前置)、遭遇區未分區、caterpillar 隊伍未繪製、龍王城父親戰死 bug-fix 未搬 |
| **C**(忠實度長尾) | **14** | attract 演出、開場劇情文字、對話鏈類支線(未阻斷主線可達性者)、寶箱端已 ok 但機關/前置待核 |
| ok(已對齊) | 15 | 靜態寶箱(`treasure.go` 121 筆覆蓋良好)、5 種位置相關 USE 道具、鑰匙開門機制、存檔系統、終盤三連戰/弱化判定與結局文字**邏輯本身**、彩虹合成 bug 修正 |
| 待人工核(卡關超時跳過) | 3 | C4(無姬大人戰勝/寶箱形式待核)、C21(銀寶珠給予路徑)、D18(龍王城寶物室清單) |
| 不適用 | 1 | A9(oracle 自身待核對項,非 ebitan 差距) |

### 修法分布(概估,以「主要應採修法」歸類,個案常混合)

| 修法 | 概估數量 | 代表案例 |
|---|---|---|
| **port-C** | ~40 | main.c 已有等效邏輯(`region 0x722`、`tavern_modal`、`g_sokoban_solved`、`DQ3_MS_*` gate、boss token)可直接移植邏輯到 Go,是最大宗 |
| **new-RE** | ~10 | main.c 也未見或未深查的部份(晝夜系統本體、attract 素材播放、開場劇情文本、六珠/不死鳥機制起源) |
| **wire** | ~8 | 邏輯已在 ebitan 但入口未接(酒場招募、下降事件、六大魔人~索瑪連戰、結局觸發、主角命名元件复用) |
| ok | 15 | 見上 |
| 待人工核/不適用 | 4 | 見上 |

### 意外發現

1. **兩套「已完成 vs 完全缺席」的清楚分野**:凡是「靜態座標→給一件道具」(寶箱)或「位置相關 USE 道具」,ebitan 幾乎都已對齊 C/RE 佐證;凡是「NPC 對話交換給物」或「examine 觸發 boss 戰」,ebitan **系統性地完全沒有**(`examine()` 函式只有 3 種分支:寶箱/warp/開門,無怪物 dispatch)。這不是零散遺漏,是整個 `region 0x722` runner 事件系統(main.c 已具備)在 Go 版尚未起步。
2. **終盤連戰(六大魔人→怨靈→殭屍→索瑪)邏輯完成度出乎意料地高**:`bossQueue`/`startZomaSeq`/`advanceBossQueue`/`lightOrb` 弱化判定/`runFinale`/ENDTXT 結局捲動全部對齊 RE 佐證,唯一缺的是「從龍王城正常走進去」這個入口——這是全專案裡 wire 成本最低、報酬最高的一塊。
3. **主角(勇者本人)沒有命名/性別選擇**,但招募同伴的酒場 UI(職業→注音/英數命名→性別)反而做得很完整——元件方向恰好與 oracle 的「誰該被命名」相反(oracle:主角必填命名,同伴後續招募;ebitan:同伴可命名,主角被跳過)。
4. **商店只能買不能賣**(`Shop.input` 無賣出分支),卡住步 52 王者之劍取得鏈,且和道具值錢/變賣類的忠實度都有關,屬於一次修好可能連帶惠及未列出的其他變賣情境。
5. **不死鳥/六珠坐騎機制完全不存在**(非部分缺,是 0 起點),是後段地圖可達性的唯一飛行手段,牽連步 43-44 全部卡死——本專案迄今對「上半段(阿里阿罕~波魯多加)」的忠實度明顯高於「下半段(飛行後地圖)」。

### 未能判定項(需人工核對,已在表中標「待人工核」)

- C4:無姬大人(八頭大蛇二戰)是否為「戰勝取得紫寶珠」或被 ebitan 簡化成靜態寶箱,語意落差待核。
- C21:銀寶珠(0x6b)不在 `treasure.go`,推測經 examine 直接給予但未找到給予程式碼路徑,待深查。
- D18:龍王城寶物室多件散裝寶箱是否落在 `treasure.go` 121 筆內,因連鎖依附 D16/D17(入口靠 debug)而低優先,未逐一核對。
- B12(附屬):金字塔按鈕機關(先東後西)判定邏輯位置未定位,只確認寶箱端存在。

檔案位置:`/home/anr2/dq3/docs/data/flow-gap-ebitan.md`(本檔)。未 commit。
