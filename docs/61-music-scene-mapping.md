# 配樂對應:哪個場景 / 地圖播哪首音樂(2026-06-28)

> 精訊版 DQ3 自己的音樂(SB FM / OPL2，從 `MBG.MCX` RE 抽出 18 軌，見 [docs/57](57-cmf-audio-re.md))。
> 本文記錄 remake **場景 → 音樂類型 → MBG.MCX 軌號**的對應,以及城鎮地圖逐 CTY 的配樂分類(T1 配曲缺口修補的成果)。
> 世界地圖(標 CTY 號)見 [`docs/maps/world_con_cty.png`](maps/world_con_cty.png)。

## 場景 → 音樂類型 → 軌號

| 場景 | 音樂類型 | MBG.MCX 軌 | 觸發條件(code) |
|---|---|---|---|
| 標題 | TITLE | 17 | 標題畫面 |
| 地表(上層) | FIELD | 0 | overworld，`layer==0` |
| 地下層(アレフガルド) | DUNGEON | 5 | overworld，`layer==1` |
| 城鎮 | TOWN | 2 | `in_town` 且非城堡/迷宮 |
| **城堡** | **CASTLE** | **3** | `in_town` 且 CTY ∈ 城堡清單 |
| **迷宮(洞窟/塔/金字塔)** | **DUNGEON** | **5** | `in_town` 且 BLK 2/4/5 或覆蓋清單 |
| 船 | SHIP | 9 | 乘船中 |
| **一般戰鬥** | **BATTLE** | **18(EBG)** | 遇敵且怪 id < 106 |
| **boss 戰** | **BOSS** | **19(EBG)** | 遇敵且怪 id ≥ 106(大魔人/巴拉摩斯/索瑪/父親/九頭龍…) |
| 結局 | ENDING | 16 | 結局序列 `end_seq >= 0` |

- **戰鬥音樂在獨立檔 `EBG.MCX`**(非 MBG.MCX!見下節)。軌 0-17=MBG、18-23=EBG(串接成單一軌號空間)。
- **戰鬥有專屬音樂**:`dq3_battlescene.c` 進戰鬥時 `id≥106 ? BOSS : BATTLE`——一般遇敵與 boss 戰各一首,來源都是 EBG。
- 對應實作:overworld/城鎮在 `main.c` 主迴圈每幀 dispatch(`cty_music_kind()` + ENDING gate);軌號表 `g_scene_track[]`(`dq3_audio.c`)。

## ★ 戰鬥音樂在獨立檔 EBG.MCX(2026-06-28 發現)

精訊 DQ3 有**兩個音樂封包**,EXE 都引用(`mbg.mcx` / `ebg.mcx` 字串):

| 檔 | 大小 | 軌數 | 內容 |
|---|---|---|---|
| `MBG.MCX` | 127 KB | 18 | 主音樂(標題/地表/城鎮/城堡/迷宮/船/結局) |
| `EBG.MCX` | 3.3 KB | 6 | **戰鬥音樂**(戰鬥/boss/勝利/升級… 短曲) |

- **格式相同**(CMF 變體,header `28 00 .. 60 00 .. 01 01`),同一 parser 可解。
- **先前 RE 漏看**:之前把 `MBG.MCX` 當「音樂」全部,沒意識到戰鬥曲在 `EBG.MCX` → 整條 pipeline
  (extract / SB / MT-32 render)只處理 MBG 18 軌,**EBG 從未抽取/render**。2026-06-28 補上。
- **remake 接法**:`dq3_audio` init 把 EBG.MCX 串接進 MBG buffer 之後(軌 18-23);
  MT-32 側 `work/mt32/track_18..23.ogg`(munt render,`export_music_mt32.sh` 已含 EBG)。
- **EBG 內 battle vs boss 確切軌**(18=battle / 19=boss)為合理推測,聽感可調(同 MBG track→曲目 未逐曲驗證問題)。

## 城鎮地圖逐 CTY 配樂分類

原版按**地圖 tileset(BLK 1–5)**分音樂:迷宮(洞窟/塔/金字塔)用迷宮 tileset → 迷宮曲;城鎮/城堡用 town tileset。
規則(`cty_music_kind()`):

- **DUNGEON 迷宮曲** = BLK ∈ {2, 4, 5}(洞窟/塔/金字塔/索瑪城)+ 覆蓋清單 `{33 沙漠之洞}`(BLK1 漏網的洞窟)。
- **CASTLE 城堡曲** = 王城清單 `{0, 2, 6, 37, 73, 76}`(名含「城」+ 有國王;DQ3 城堡整城播城堡曲)。
- **TOWN 城鎮曲** = 其餘(村/町/神殿/祠堂)。

全 89 CTY:**6 城堡 / 32 迷宮 / 62 城鎮**(註:含未命名 CTY,以 BLK 分類;下表只列已命名者供核對)。

| CTY | 名稱 | BLK | 配樂 |
|---|---|---|---|
| 0 | 阿里阿罕城 | 1 | 🏰 CASTLE |
| 1 | 雷貝村 | 1 | 🏘️ TOWN |
| 2 | 羅馬利亞城 | 1 | 🏰 CASTLE |
| 3 | 卡薩布村 | 1 | 🏘️ TOWN |
| 4 | 諾阿尼魯村 | 1 | 🏘️ TOWN |
| 5 | 精靈之村 | 1 | 🏘️ TOWN |
| 6 | 阿莎拉慕城 | 1 | 🏰 CASTLE |
| 7 | 岬之洞窟 | 2 | ⛰️ DUNGEON |
| 8 | 拿吉米之塔 | 4 | ⛰️ DUNGEON |
| 10 | 香巴尼塔 | 4 | ⛰️ DUNGEON |
| 11 | 精靈之村西南洞窟 | 2 | ⛰️ DUNGEON |
| 13 | 金字塔 | 5 | ⛰️ DUNGEON |
| 15 | 巴哈拉達鎮 | 1 | 🏘️ TOWN |
| 18 | 加爾那之塔 | 4 | ⛰️ DUNGEON |
| 19 | 洞穴(八頭大蛇) | 2 | ⛰️ DUNGEON |
| 20 | 提頓村 | 1 | 🏘️ TOWN |
| 21 | 日邦格村落 | 3 | 🏘️ TOWN |
| 33 | 沙漠之洞 | 1 | ⛰️ DUNGEON(覆蓋清單) |
| 37 | 達魯波多加城 | 1 | 🏰 CASTLE |
| 47 | 朗錫爾村 | 1 | 🏘️ TOWN |
| 49 | 達瑪神殿 | 1 | 🏘️ TOWN |
| 73 | 依席斯城 | 3 | 🏰 CASTLE |
| 76 | 貝亞城 | 3 | 🏰 CASTLE |

## track 結構特徵分析(供聽感確認 #2;2026-06-28)

無 DOSBox 不能逐曲試聽,改用**時長/事件數**當結構訊號做合理指派(標題短、地表長循環、jingle 極短)。
MT-32 render 一輪時長:

| 軌 | 時長 | 現指派 | 結構判讀 |
|---|---|---|---|
| 17 | 141.8s | TITLE | 最長;當標題偏長(疑開場/長主題),**最該聽感確認** |
| 00 | 82.4s | FIELD | 長循環,合地表 ✓ |
| 12 | 71.9s | — | 長,未用(疑地表變體/重要主題) |
| 16 | 61.4s | ENDING | 長,合結局 ✓ |
| 05 | 51.5s | DUNGEON | |
| 01 | 49.9s | — | 未用 |
| 13 | 44.9s | — | 未用 |
| 02 | 36.0s | TOWN | |
| 09 | 34.6s | SHIP | |
| 03 | 28.4s | CASTLE | |
| 04/06/07/08/10/14/15 | 22–40s | — | 未用(城鎮/迷宮/設施變體候選) |
| **18** | **11.0s** | **BATTLE**(EBG)| 戰鬥 loop ✓ |
| **19** | **15.5s** | **BOSS**(EBG)| EBG 最長最華麗 → boss 合理 ✓ |
| 20/21 | 11.1/11.6s | —(EBG)| 戰鬥變體候選 |
| 22 | 5.4s | —(EBG)| 極短 = **jingle**(疑勝利/升級)|
| 23 | 8.6s | —(EBG)| 短 = jingle(疑全滅/升級)|

- **結構支持的**:EBG battle=18 / boss=19(loop vs jingle 分界清楚)、FIELD=00、ENDING=16。
- **最需聽感確認的**:TITLE=17(141s 偏長)、TOWN/CASTLE/DUNGEON/SHIP 的確切軌。
- **取得確認最快路徑**:render 各軌試聽(`work/mt32/track_NN.ogg`),玩過原版者一聽即可認曲、回填正確軌號。

## 驗證狀態(誠實標註)

- ✅ **場景 → 音樂類型 dispatch 正確**:每個場景播對「類別」(城堡/迷宮/城鎮/戰鬥/boss/結局各有曲),
  game_tester 93/93 無回歸。配曲缺口(CASTLE/ENDING/DUNGEON 0 觸發)已修。
- ⚠ **軌號(`g_scene_track[]`)未對原版逐曲驗證**:18 軌指派(TITLE=17/FIELD=0/…)是初步合理對應;
  「哪一軌是哪首曲」的精確對應仍是 **T1 #2** 待辦(RE 線索:`_sbfm_play_music` 呼叫端 → 全域 `[ds:0x08]` track 指標
  → MBG.MCX 偏移表;見 WORKLIST Tier 1)。本環境無 DOSBox 不可逐曲試聽,需靜態 trace 或聽感辨識。
- ⚠ **城堡清單 best-effort**:BLK tileset 不分城堡 vs 城鎮,城堡靠「名含『城』+ 有國王」清單;
  玩過原版者可按聽感增刪 `cty_music_kind()` 的 `CASTLE[]` / `DUNGEON_EXTRA[]`。

## 靜態 RE 分析過程(留給後人:我們怎麼推出來的)

> 環境限制:此環境**無 DOSBox runtime**(記憶 `dq3-no-dosbox-debugger`),不能跑原版逐曲聽/逐場景看。
> 所以一切走**靜態反組譯 + 資料反推**,不靠動態觀察。方法如下。

### A. CTY → 音樂類型分類(配曲缺口怎麼修的)

1. **先找有沒有 per-CTY 音樂表**:`grep` `src/dq3_exedata.c`(從 EXE 抽出的所有表)——
   只有 `growth/thresh/terrain/event/warp/cty_loc/map_blknum`,**沒有**獨立音樂表。
   → 結論:原版不是查「音樂 id 表」,而是**按地圖 tileset(BLK)衍生音樂**。
2. **拿 BLK tileset 當分類器**:`dq3x_map_blknum[89]`(每 CTY 的 DQ3<n>.BLK,值 1–5,從 EXE file 0x0a04 抽出)。
3. **用已知 CTY 名反推每個 BLK 是什麼**(交叉 `docs/maps/cty-names.md` + `docs/32`):
   - BLK 4 → 全是塔(拿吉米/香巴尼/加爾那/Kol 塔)+ 索瑪城 ⇒ **塔/dungeon tileset**。
   - BLK 5 → 金字塔 + 洞窟(甘達特巢穴)⇒ **洞窟/迷宮 tileset**。
   - BLK 2 → 全是洞窟(岬之洞窟/八頭大蛇洞穴…)⇒ **洞窟 tileset**。
   - BLK 1/3 → 村/町/城(雷貝村/羅馬利亞城/依席斯城…)⇒ **城鎮/城堡 tileset**。
   → 乾淨規則:**BLK {2,4,5} = 迷宮曲;BLK {1,3} = 城鎮/城堡曲**。
4. **抓規則漏網**:再 `grep` 名含「洞/窟/穴/塔/金字塔」者對 BLK,發現 **CTY33 沙漠之洞是 BLK1**(洞窟卻落在城鎮 tileset)
   → 補進 `DUNGEON_EXTRA` 覆蓋清單。
5. **城堡 vs 城鎮**:BLK 不分這兩者(同 town tileset)。DQ3 城堡整城播城堡曲 → 用「名含『城』+ 有國王(王座 NPC)」
   清單 `{0,2,6,37,73,76}`。此為 best-effort(無音源可驗,留聽感調整口)。

### B. track → 曲目對應(T1 #2,RE recon — 方法與目前進度)

目標:`g_scene_track[]` 把場景對到 MBG.MCX 第幾軌,是初步猜測;要靜態還原原版「哪場景播哪軌」。追法:

1. **定位播放函式**:`tools/omf_lib.py` 解 `SBCM.LIB`(OMF 物件庫)→ 取 `_sbfm_play_music` 的 byte signature
   → 在 `DQ3.EXE` byte-match → 命中 file **0x140ee**(段:off `0x129c:0x3be`)。見 [docs/57](57-cmf-audio-re.md)。
2. **找呼叫端**(掃 `9A` far-call 指向 0x140ee):只 **2 處**(file 0x11aaf / 0x11d2f),
   且都在同一段「音樂驅動包裝」內。
3. **反組譯呼叫端看 track 從哪來**:兩處都 `push ds; push si; lcall play_music`,其中 `si = [ds:0x08]`
   (caller2 是 `[bp+...]` 同義結構)——即**播的是全域音樂控制塊 `[ds:0x08]` 的 track 資料指標**(tempo 在 `[ds:0x0c]`)。
   → 這 2 個呼叫端是**集中式「播放當前已載入的軌」**,本身不選曲。
4. **下一步(未完)**:真正「場景 → track index」在**更上層**——誰把 `[ds:0x08]` 設成
   `MBG_base + 前導dword偏移表[idx]`。要再往上追設定 `[0x08]` 的寫入點(1–2 跳),看各場景上下文傳的 `idx`。
   追到後就能把 `g_scene_track[]` 從「猜測」改成「對原版驗證」。

> 教訓:① 沒有獨立音樂表時,**tileset 就是原版的隱性分類器**,用已知地名反推 tileset 語意比硬追選曲碼快。
> ② 選曲是「集中播放包裝 + 全域控制塊指標」的間接結構,直接呼叫端只播指標、不選曲——要往上追「誰設指標」。
> 與 `62-static-provenance-trace`(反追溯源)、`64-re-screenshot-oracle` 同源:撞動態牆改走靜態反推。
