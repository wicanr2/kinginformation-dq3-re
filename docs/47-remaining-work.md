# 剩餘工作盤點(往「完整還原 + 能破關」)

> 階段性盤點:remake 已能做什麼、離「能破關」還差什麼,依優先序。驗證口見 docs/46。

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
2. **劇情旗標推進**:remake 目前不靠 gameplay 推進故事旗標 → 多數 flag-gated 事件不會自然發生。
   需設計「事件 → set_flag → 解鎖下一步」的旗標流(部分已有:#2 合成設 0x139)。
3. **船**:取船劇情(波魯多加胡椒換船)+ 船移動(跨海 tile)+ 船狀態存檔。地表 gate 2 的核心。
4. **終盤**:大魔王索瑪戰 + 結局序列(對話 bank D3TXT07 已有索瑪台詞)。

> 現實路徑:A1/A2 是同一塊(旗標驅動的 scripted-event VM)。**最高槓桿但最難純靜態**;
> debug 口已能「跳過劇情直接到任一狀態」驗證 —— 可先用 debug 口把各段接成可玩,再補真實觸發。

### B. 完整還原(faithful,非破關必須)

5. **狀態效果**(中毒/麻痺/詛咒)→ 教會解毒/解詛咒、戰鬥毒傷。
6. **NPC 互動完整化**:cmd-menu「話す」走 byte4 對話(目前部分路徑用舊 demo);子型2 條件對話的
   旗標分支(目前只顯示主對話)。
7. **overworld 旗標 portal 全表**:目前接 3 個寫死座標(docs/45 §3.2),完整需抽 0x396e 全部分支。
8. **scripted warp 全接**:§5b 的 8 個 0xd1f9 warp(洞穴→城)NPC 觸發 + dest≥100 overworld 回程。
9. **道具效果**:藥草/聖水/翅膀等使用效果(目前道具欄可列、未實作使用)。

### C. Polish

10. 寶箱開過 tile 翻面(取後外觀變空)· 旅社/教會精確收費公式 · 同城多攤逐攤化(資料已備)。
11. 戰鬥逃跑/道具指令完整 · 咒文全效果 · 轉職(達瑪)實際換職。

## 單一最大 blocker

**劇情旗標驅動的 scripted-event 觸發系統(A1/A2)**——它是「把已做好的各系統串成一條可破關主線」
的連接組織。handler 都在,缺觸發。純靜態已到極限(`[0x722]` 無 setter);**務實做法**:
用 debug 口逐段驗證 + 設計 remake 自有的旗標流(不必 1:1 還原 runner VM),讓主線跑得通。

## 驗證方法(已就緒)

- 單元測試:stats/combat/battle/dialogue/npc/door/inventory/roster/menu/save(全綠)。
- DEBUG 口 + DQ3_DUMP:任意狀態 headless 截圖(docs/46)。全 89 城 load 掃描零崩潰。
- 待補:headless **腳本化 playthrough**(連續輸入注入)驗證主線串接。
