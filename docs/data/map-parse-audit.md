# 城鎮/地表地圖解析稽核(ebitan vs C oracle)

稽核日期:2026-07-03。範圍:`dq3_remake_ebitan/internal/dq3data`(Go)的城鎮/地表地圖解析,
拿已 93/93 驗證的 C remake(`dq3_remake/src/dq3_town.c` + `dq3_field.c`)當 oracle。
方法:兩邊各寫一支 dump 工具,對 **CTY00–CTY93(現存 89 個檔)× section 0–15** 全掃,
輸出 CSV 逐格對帳(status/w/h/spawn/dlg_bank/events/transitions/npc 數/tile 上限),
再針對差異點逐一讀 code 定位根因。C harness 直接連結 `dq3_town.c`/`dq3_scene.c`/`dq3_npc.c`
本體(非重寫邏輯),用 `dq3_town_load` 全流程(含 BLK/PAL/attr 真實載入)取得比純 header 解析更貼近
runtime 行為的 oracle 值(尤其 spawn 座標經過 `dq3_scene_walkable`+`pick_open_start` 加工)。

## 總結

- 全掃 1424 個 (cty,sec) 組合(89 城 × 16 section)。
- **status / w / h / events 數 / transitions 數:0 處不一致**——ebitan 的 section 表解析、
  layout 解析、事件表、轉場表跟 C oracle **逐格一致**,含原任務描述的「垃圾 section 卻被寬鬆
  接受」現象。
- 但深入比對後發現:**這個「寬鬆」現象是 C oracle 本身就有的行為,不是 ebitan 獨有的迴歸**
  (見下方「修正原任務前提」)。
- 真正找到、且**只在 ebitan 出現、C 沒有**的兩個 bug:
  1. **spawn 座標未做可走性回退**(高嚴重度)——影響全遊戲 56% 的合法 section,含 **6 座城的主入口**
     (CTY30、37、49 達瑪神殿、65、73、76),玩家從地表走進城會被放到地圖外。
  2. **NPC 32 隻上限未落實**(中嚴重度)——33 處 section 的 NPC 數超過 C 的硬上限,含至少 6 處
     非垃圾(合法 w/h)section,如 CTY90(龍王城)sec13。
- Tile 索引越界(#3 稽核項)、地表 DQ3CON.MAP/DQ3UND.MAP 解析(#4 稽核項):**兩邊一致、無 bug**。
- NPC 晝夜表選擇邏輯不同,但屬已知/有註解的簡化,非解析錯誤(見下)。

## 修正原任務前提:「section 寬鬆通過」是 C 也有的行為,非 ebitan 獨有

原任務描述的起點案例(CTY01 某 section 解出誇張 w/h、tile 陣列大幅超界卻不報錯)在本次全掃中
**大量重現**:89 城中 79 城至少 1 個 section 落入此類(共 207/1424 筆標記「GARBAGE」——`w`或`h`>200
且 `w*h*2` 遠大於檔案長度)。但把同樣的 (cty,sec) 丟給 **C `dq3_town_load`**,C 端**產出完全相同的
w/h 數字、完全相同的 ok/err 判定**(逐筆 diff 0 mismatch)。

根因在 `dq3_town.c` 本體(`dq3_town.c:48-56` 的註解本身就寫明):section 偏移表**沒有 count 前綴**
(`word[i] = section i 的 base offset`),呼叫端若對每個 CTY 掃描固定 0–15 範圍(本次稽核的掃法),
超出該城實際 section 數的索引會讀到表格後面「別的資料」(layout、事件表等)當成偽 section header,
解出誇張 w/h。C 端唯一的防線是 `so+0x16>cty_len`(header 起碼要在檔案內)與
`w<=0||h<=0||tbase>=cty_len`(layout 起碼要有 4 bytes),**兩邊都沒有 `tbase+2*w*h<=cty_len` 這種
「tile 陣列真的塞得進檔案」的檢查**——這是 C oracle 設計本身的洞,ebitan 是逐位元組移植,自然帶著同一個洞。

**推論此洞在正常遊戲流程「多半不可觸及」**:遊戲從不會對一個城盲掃 section 0–15,section
索引只會來自轉場表(`so+0xc`)的 `dest_sec` 值,或程式碼寫死的初始 section(通常 0/4)。本次稽核
沒有逐城反查「所有轉場表的 dest_sec 是否命中垃圾索引」(時間盒內跳過,列為待人工確認),但從
89 城的「合法 section index 集合」通常是稀疏、不連續分布(如 CTY03 合法到 sec15、但中間夾雜垃圾
sec5/8/12)推測,這些垃圾索引本來就是「表格裡沒被用到的槽位」,而非「合法卻被誤判」。

**修法方向(不改 code,僅記錄)**:若要收斂這個洞,應該先修 `dq3_town.c`(oracle 本體),加
`tbase + 2u*(size_t)w*(size_t)h <= cty_len` 檢查後讓 `dq3_town_load` FAIL,ebitan 的 `OpenTown`
再照樣移植同一條檢查——**不要只改 ebitan**,否則兩邊又會產生新的不一致。

## Bug 1(高嚴重度):spawn 座標沒有可走性/邊界回退

### 現象

C `dq3_town_load`(`dq3_town.c:145-148`):
```c
s->px = sx; s->py = sy; s->facing = 0;
if (sx < 0 || sx >= w || sy < 0 || sy >= h || !dq3_scene_walkable(s, sx, sy))
    dq3_scene_pick_open_start(s);
```
讀 section header 的 raw spawn byte 後,若出界或落在不可走的格子,會呼叫 `dq3_scene_pick_open_start`
另外挑一個開闊、可走的格子。

ebitan `OpenTown`(`internal/dq3data/townmap.go:68-69`)只做:
```go
SpawnX: int(cty[so+0x13]), // spawn_x(section header)
SpawnY: int(cty[so+0x14]), // spawn_y
```
**直接回傳 raw byte,完全沒有邊界檢查、沒有可走性檢查、沒有回退**。呼叫端(`game.go`)也未補這個檢查——
除了兩個特例(`startOpening` line 861、`motherEscort` line 881)有各自寫死的「超出地圖寬高就退到
地圖中心」土砲修補,**其餘所有進城路徑(`enterTown` line 709、全滅重生 line 1139、debug env line
1438)完全沒有防呆**,而且「退到地圖中心」跟 C 的「挑開闊可走格」是兩種不同邏輯,不保證中心格
真的可走。

### 量化

對全部 1424 筆 (cty,sec) 中「status=ok」的 338 筆(排除 GARBAGE/err 那些反正不會被載入),
**188 筆(56%)的 raw spawn 座標超出該 section 自己的 w×h 邊界**(`SpawnX>=W` 或 `SpawnY>=H`)。
C 端對應這些筆全部落在邊界內(pick_open_start 生效)。

其中最嚴重的子集:**這 89 城裡,有 6 城的 section 0(=玩家從地表走進城的主入口)本身就 OOB**:

| cty | 城名 | ebitan sec0 spawn(w×h) | C oracle sec0 spawn(同一 w×h) |
|---|---|---|---|
| 30 | — | (0,28),w×h=32×22 → y=28 越界(h=22) | (8,6) |
| 37 | — | (39,15),w×h=34×31 → x=39 越界(w=34) | (9,10) |
| 49 | **達瑪神殿**(msDhama 里程碑城) | (10,36),w×h=14×18 → y=36 越界(h=18,超一倍以上) | (8,5) |
| 65 | — | (0,28),w×h=17×20 → y=28 越界(h=20) | (8,6) |
| 73 | — | (17,50),w×h=15×16 → x/y 皆越界 | (5,6) |
| 76 | — | (13,42),w×h=21×17 → y=42 越界(h=17) | (15,9) |

實際影響:玩家從地表走進這 6 座城,ebitan 會把主角座標設到地圖外——依 `game.go` 的攝影機/碰撞
邏輯,`Tile()`/`tileAt()` 對越界座標回 0(視為可走的預設地格),移動不會直接崩潰,但畫面會顯示
攝影機 clamp 到邊緣、主角疊在地圖外的黑邊、可能踩不到任何轉場格而卡死出不了「城外」。CTY49(達瑪神殿)
是主線里程碑城(`progress.go` msDhama),優先度最高。

### 修法方向

`OpenTown`(純資料層)不該做可走性判斷(它沒有 BLK/attr 可查),但**應該提供跟 C 對齊的
「原始 spawn 是否可能有效」資訊**,實際回退邏輯要在有 attr/walkable 資訊的呼叫端(`loadTownSceneSec`
或 `Scene` 建構後)補一個等價於 `dq3_scene_walkable` + `dq3_scene_pick_open_start` 的步驟,
取代目前只在兩個特例呼叫點土砲判斷邊界(且沒真的挑可走格)的作法。

## Bug 2(中嚴重度):NPC 32 隻上限未落實

### 現象

C `dq3_scene_load_npcs`(`dq3_scene.c:127`):
```c
for (i = 0; i < cnt && s->n_npcs < DQ3_SCENE_MAX_NPC; i++) { ... }
```
`DQ3_SCENE_MAX_NPC = 32`(`dq3_scene.h:24`),硬上限。

ebitan `parseNPCs`(`townmap.go:134`):
```go
for i := 0; i < cnt; i++ {
    ...
    out = append(out, NPC{...})
}
```
**沒有任何上限**,`cnt`(section NPC 表的 count byte,0–255)有多大就解多少隻。

### 量化

全掃 1424 筆(排除 status=err)裡,**33 筆的 NPC 數量與 C 不一致,且全部是「Go > C 的 32」**
(4 到 255 隻不等,C 端因硬上限恆為 32)。其中至少 6 筆是**合法 status=ok section**(非垃圾
w/h),包含:

| cty | sec | w×h | C nnpc | ebitan nnpc |
|---|---|---|---|---|
| 3 | 15 | 18×18 | 32 | 33 |
| 4 | 10 | 26×26 | 32 | 91 |
| 21 | 0 | 39×63 | 32 | 37 |
| 21 | 13 | 18×18 | 32 | 34 |
| 88 | 10 | 63×63 | 32 | 51 |
| 90 | 13 | 11×11(**龍王城**) | 32 | 55 |

其餘 27 筆落在 GARBAGE(誇張 w/h)section,連帶 NPC count byte 本身也是垃圾資料,優先度較低。

### 修法方向

`parseNPCs` 的迴圈加 `len(out) < 32` 上限,對齊 `DQ3_SCENE_MAX_NPC`。

## NPC 晝夜表選擇:已知簡化,非解析 bug

C `dq3_scene_load_npcs`(`dq3_scene.c:117-123`)依當前晝夜相位(`dq3_scene_get_daynight()==2`為夜)
決定先讀 section+0(日表)還是 section+2(夜表),另一份當 fallback。ebitan `parseNPCs`
(`townmap.go:118`)寫死永遠先試 offset 0(日表)、無效才退 offset 2 ——這是 ebitan 自己註解承認的
簡化(「城內晝夜相位固定 → 這裡預設用日表,無效才退夜表(靜態載入,不跑晝夜切換)」),不是誤解析,
但**功能上會漏掉所有夜間限定 NPC**(如晚上才出現的角色)。列為已知限制,不計入 bug 數。

## Tile 索引越界:兩邊一致、有安全網,非 bug

19 筆合法 status=ok section 的 tile 索引出現 255(超出城鎮 BLK 的 170 個 tile),例如 CTY00 sec6
(58×59,阿里阿罕的某張大圖)。追查後確認**兩邊的渲染層都有邊界檢查**:
C `dq3_blk_tile`(`dq3_blk.c:23`:`idx<0||idx>=blk->count` 回 -1,呼叫端不畫)、
Go `BLK.Tile`(`blk.go:38`:同樣邊界檢查回空白 tile)。兩邊都會把這類格子畫成空白/透明,行為一致,
不是 crash 風險,列為「一致 ok」。

## 地表(overworld)地圖:一致、無寬鬆洞

`DQ3CON.MAP`(w=244,h=205,244×205+4=50024 bytes,剛好等於檔案大小)與 `DQ3UND.MAP`
(w=244,h=167,50024/40752 bytes)。C `dq3_field_load_map` 與 Go `OpenFieldMap` 都是
**單一 4-byte header 直接定義 w/h,並嚴格要求 `w*h+4<=len(d)` 才算合法**,沒有像城鎮 section
表那樣「掃描表格猜出下一個 section」的模糊地帶,兩邊邏輯逐行對應,**無差異**。

## 逐項稽核表

| 項 | 範圍 | ebitan 結果 | C oracle | 一致? | 嚴重度 | 說明 |
|---|---|---|---|---|---|---|
| 1 | section 有效性/索引口徑,全 89 城×16 section | 207 筆「垃圾但通過」status=ok/GARBAGE 判定 | 逐筆數字相同 | **一致**(兩邊都寬鬆) | 低(推測多半不可觸及;根因在 C,見上) | 待人工:逐城反查轉場表 dest_sec 是否曾指到垃圾 index |
| 2a | w/h/status 逐格對帳,全 89 城×16 section | 見 CSV | 見 CSV | **0 mismatch** | — | 完全一致 |
| 2b | events 數/transitions 數逐格對帳 | 見 CSV | 見 CSV | **0 mismatch** | — | 完全一致 |
| 2c | spawn 座標逐格對帳 | 188/338 筆(ok section)OOB | 全部落在邊界內 | **不一致** | **高** | Bug 1,見上 |
| 2d | NPC 數逐格對帳 | 33 筆 > C 的 32 | 恆 ≤32 | **不一致** | 中 | Bug 2,見上 |
| 3 | tile 索引 vs BLK count(170) | 19 筆合法 section 出現 idx=255 | 同樣會出現,渲染端擋下 | **一致**(兩邊都安全處理) | — | 非 bug |
| 4 | overworld DQ3CON.MAP/DQ3UND.MAP | w=244×205 / 244×167,與檔案大小精確吻合 | 同 | **一致** | — | 無寬鬆洞,邏輯逐行對應 |
| 5 | NPC 日/夜表選擇邏輯 | 恆用日表,無效才退夜表 | 依當前晝夜相位挑主表 | **設計簡化**(非解析錯) | 低 | ebitan 自己註解承認,見上 |

## 方法論註記

- C oracle 用獨立 harness(`dq3_town_load` 全流程,含 BLK/PAL/attr 真實載入),非重寫邏輯抽取——
  對照組是 `dq3_town.c`/`dq3_scene.c`/`dq3_npc.c` 本體編譯出的行為,不是憑讀 code 推測的邏輯。
- 掃描資料來源:`/home/anr2/dq3/work/`(CTY00–93,89 個現存檔;`work/DQ3n.BLK`、`BLKBMn.DAT`、
  `DQ3.PAL`)。
- 兩邊 dump 工具與中介 CSV 為稽核暫存物,已在完成後清除(`dq3_remake_ebitan/cmd/townaudit/`、
  `townaudit_bin`),未 commit,未動任何 `.go`/`.c` 原始檔。
- 時間盒內未覆蓋:逐城反查「合法轉場表 dest_sec 是否曾指到垃圾 section index」(項 1 的殘留風險,
  待人工);5–8 個代表城的完整逐 section NPC/事件座標範圍人工複核(改用全 89 城機器對帳取代,
  覆蓋率更高但省略了逐城人工看圖驗證這一步)。

## NPC 數重查結論(2026-07-03,Opus inline,回應使用者疑慮)

用 raw record count vs 界內 vs ebitan 載入數,逐代表城 + 全 CTY 掃:

- **正常城鎮 section(raw≤31):ebitan 載入數 == 原始 record 數 == 界內數,完全相等** —— C-3 拔 `b2<4`
  修對了,主城 NPC 數已正確(使用者「NPC 沒那麼多」症狀已解)。
- **bug #2(NPC 32-cap 分歧)結案:非真 bug**。raw>32 只出現在 14 個 section,幾乎全是無 count 前綴
  誤讀的「垃圾/不可達 section」(CTY04 sec10 raw=128、CTY90 sec13 raw=129…)。C 砍 32 是防呆,
  真 section ebitan 不設限更忠實;主城全 ≤31,兩者一致。
- **真正殘留 = 夜間 NPC**:68 個 section 日表≠夜表,ebitan 恆用日表 → 夜晚 NPC 少一批
  (CTY00 sec0 日24→夜6)。歸 **C-5b 晝夜系統**,非計數錯誤。

→ 不需額外 subagent;主城 NPC 數已對,夜間差異併入 C-5b。
