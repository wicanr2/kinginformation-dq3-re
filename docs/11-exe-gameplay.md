# DQ3.EXE 核心玩法函式:場景繪製 / 移動碰撞 / HUD(RE → C)

接續 docs/08(主迴圈骨架)與 docs/04(地圖格式),把主迴圈每幀呼叫的三個核心
玩法函式逐一反編譯,並由場景繪製/碰撞邏輯**解開 docs/04 留下的 CTY 城鎮佈局謎團**。

工具:`tools/re_disasm.py`(capstone,docker 內,給 file offset)、函式圖
`docs/data/exe_funcs.json`。位址空間 = seg0 邏輯 offset(file off = 0x1370 + off)。
主資料段 DS=0x14dd(file base 0x16140)。反編譯 C 在
`re/render.c`(場景/HUD)、`re/player.c`(移動碰撞)、宣告在 `re/gameplay.h`。

## 命名澄清:sub_255b 與 sub_19b8 是「兩個場景處理常式」

docs/08 / exe.h 把 `sub_255b` 命名 `draw_scene`、`sub_19b8` 命名 `update_player`,
但反組譯後兩者**結構逐行鏡像**,差別只在第一行的場景 gate 與方向 jump table 位址:

| 函式 | file off | gate(`[0x4f2d]`)| 方向 jump table | 角色 |
|---|---|---|---|---|
| `sub_255b` | 0x38cb | `==0`(SCENE_OVERWORLD)| `[bx*2 + 0x2562]` | **地表**場景:移動+捲動+繪製 |
| `sub_19b8` | 0x2d28 | `==1`(SCENE_TOWN)| `[bx*2 + 0x255a]` | **城鎮**場景:移動+碰撞+捲動 |

主迴圈每幀**兩者都呼叫**;只有與目前 `[0x4f2d]` 相符的那個會實際工作,另一個第一行
gate 不符即 `ret`。`[0x4f2d]` 的值域:0=地表、1=城鎮、2=地表→城鎮過場、3=城鎮→地表過場
(過場是「本幀切換」,由對應場景函式末段收尾)。

### 方向 jump table 內容(已從 DS 資料段讀出)

```
城鎮 @DS:0x255a (dir 下/左/上/右): 0x1ba2  0x1c26  0x1c99  0x1d09
地表 @DS:0x2562 (dir 下/左/上/右): 0x2880  0x2904  0x2977  0x29e7
```

每個 handler 是該方向的「捲動 + 局部重繪」常式:設定 blit 矩形
(`[0x4f25]`=螢幕 X、`[0x4f27]`=Y、`[0x4f21]`=寬 0x14=20 tile、`[0x4f23]`=高)與目的
VRAM 偏移 `[0x4f09]`(雙緩衝翻頁,於 0x8200 wrap),再呼叫 tile-row blitter
(`sub_30cf` 系列)貼出新捲入的一排 tile。

## 場景繪製 `sub_255b` / `sub_19b8`(`re/render.c` / `re/player.c`)

每幀流程(以城鎮版 `sub_19b8` 為例,地表版同構):

1. gate:`[0x4f2d]!=1` → `ret`。
2. 有移動方向(`[0x4f1f]!=0xff`):
   - `call 0x2dda`(`player_move_collide`):算目標 tile + 碰撞;**被擋時函式內把
     `[0x4f1f]=0xff`** 取消本幀移動。
   - 若移動仍成立:設捲動狀態 `[0x2784]`(0xff→2,事後→1),依方向走 jump table
     呼叫捲動 handler,`call 0xf003`/`0xf8b4` 收尾。
   - warp 待處理(`[0x257b]==1`)→ `lcall 10b6:385` / `0x3685` / `0x2cdb` / `10b6:3a0`。
   - `[0x4f46]&1`(踩事件 tile)→ `sub_2f01`(`scene_event_poll`)。
3. 無移動 / 被擋:整屏重繪(`sub_32be`,城鎮版)。
4. 過場(`[0x4f2d]==3`):暫切回 1 重載資源 → 切 0(地表)→ 重開檔案/音效/鍵盤
   (`lcall 1053:240`/`1053:dc`/`call 0xa9`/重開地表 + `lcall 1104:123`)。

## 玩家移動 + 碰撞 `sub_2dda`(`re/player.c::player_move_collide`,file 0x414a)

城鎮路徑的碰撞核心。**這裡解開了 BLKATT/blkbm 屬性的位元語意。**

1. 目標 tile = 玩家 tile + 本幀位移:`tx=[0x4f33]+[0x2572]`、`ty=[0x4f35]+[0x2574]`。
   - `[0x4f33]/[0x4f35]` = 玩家當前 **tile 格座標**(非螢幕像素);
     `[0x2572]/[0x2574]` = 本幀 ±1 位移(主迴圈方向鍵寫入)。
2. 邊界檢查:`0<=tx<[0xb28]` 且 `0<=ty<[0xb2a]`(寬高)。出界 → 設過場 `[0x4f2d]=3`、
   `[0x2576]=[0xb2d]`(邊界 tile),套用位移後離開。
3. 讀目標 tile:`si = [0xb26] + (ty*width+tx)*2`;`tile_word = seg_cty:si`(u16);
   **`tile_index = 低 byte`**。
4. 查屬性表:`attr = ((u16*)0x308e)[tile_index]`(`tile_attr_tab`,由 `load_blk` 載入
   `blkbmN.dat` 到 DS:0x308e);存 `[0x2577]`。
5. 依屬性 `attr` 分支決定可否移動:

| 屬性條件 | 行為 |
|---|---|
| `attr_hi & 0x20` | **可走且觸發**:先走過去,呼叫 tile 互動;失敗則回退並播撞牆音 |
| `attr & 0x0001` | **阻擋(牆)**:不移動,`[0x2875]=1` 撞牆音,`[0x4f1f]=0xff` |
| `attr & 0xe000` | **特殊地形/事件**:`[0x4f46]|=1`(觸發旗標);再依 `attr_hi & 0x1f` 細分(門/階梯/觸發點),設 `[0x4f46]|=0x800`、事件碼 `[0x258c]` |
| 其他 | **可走**:記 `[0x2576]=index`;比較 `attr_hi & 0xc0` 與 `[0x2579]`(角落屬性),變化 → 設 warp 旗標 `[0x257b]=1`;套用位移 `[0x4f33]+=dx; [0x4f35]+=dy` |

→ **docs/04 的「BLKATT 每 tile 2 byte mask」由此對齊到具體位元**:
`bit0`=阻擋牆、`bits13-15(0xe000)`=事件地形、`bit5(0x20)`=可走且觸發、
`bits6-7(0xc0)`=角落/出口類別。注意:遊戲執行期讀的屬性表在 `blkbmN.dat`
(載入到 DS:0x308e),`BLKATT.DAT` 是其原始碼匯出對應表。

## CTY 城鎮佈局格式(謎團已解,印證 docs/04)

`sub_30cf`(file 0x443f,`re/render.c::cty_select_section`)是解碼入口。CTY 整檔由
`load_cty` 載入 `seg_cty`(DS:0x2536)後**就地解析,不另解壓**。

### 檔案結構(完整)

```
+0x00  u16  n          section 數(docs/04 已解;觀察 2 / 4 / 12 …)
+0x02  n×u16           section 偏移表;0xffff = 空 section
... 各 section(起點 = section_off,以下偏移皆相對 section_off):
  +0x00 .. 0x0d        section 標頭(14 bytes;旗標/雜項,部分未逐欄定名)
  +0x0e  u16           layout_ptr(版面資料指標,相對 section_off)
  +0x10  u16           旗標 word(bit15 → 設場景旗標 [0x4f46] bit1)
  +0x15  4 bytes       出口/warp 參數([0xb56/0xb57/0xb58/0xd71])
... 版面資料(起點 = section_off + layout_ptr):
  +0x00  u16  width    城鎮寬(tile)  → [0xb28]
  +0x02  u16  height   城鎮高(tile)  → [0xb2a]
  +0x04  u8   spawn_x  進城起始 tile X → [0x4f33]
  +0x05  u8
  +0x06  u8   spawn_y  進城起始 tile Y → [0x4f35]
  +0x07  u8
  +0x08  u16[height][width]   tile 陣列(row-major)
         每格 u16:低 byte = BLK tile index、高 byte = 屬性(0xc0 位另存 [0x2579])
```

選 section 由 `[0x256a]`(section index)決定:`sect_off = ((u16*)cty_base)[idx]`。

### 為何 docs/04 解不開

docs/04 把某段直接當「壓縮命令串 / RLE」去解,run 數加總對不上 `w×h`。實際上:
**tile 佈局不是壓縮的**,是每格一個 u16(index+屬性);先前漏了「要先經 section
偏移表 → +0x0e layout_ptr → 才到達 w/h/tiles」這層**間接定址**,且 tile 是 u16
(不是 u8),所以把段當平面 u8 陣列掃 16/20/24/32 寬當然湊不出可辨識畫面。

### 驗證(對真實 CTY 檔)

依上述程式路徑解碼 `CTY00.DAT`:n=12,section 0 偏移 0xf50 → layout_ptr 0x49 →
版面 `w=16 h=12`,tile_base 0xf9d,`16×12×2=384` bytes 正好落在檔內。tile 陣列周邊
是牆 index(24/26)、內部是地板 index(4),典型 DQ 室內房間。`tools/re_gameplay_cty.py`
用 `DQ31.BLK` 把 section 0 貼成圖(`work/cty00_sect0_verify.png`):呈現磚牆外框、
紅地毯房間、床、寶箱與走廊 —— 完整、可辨識的建築室內,**佈局解碼正確**。

| 檔 | n | sect0 偏移 | layout_ptr | w×h | tile_base | 檔內合身 |
|---|---|---|---|---|---|---|
| CTY00 | 12 | 0xf50 | 0x49 | 16×12 | 0xf9d | ✓(384B) |
| CTY01 | 4 | 0x970 | 0x3f | 9×14 | 0x9b3 | ✓(252B) |

(spawn 欄位 byte 值 15/22、0/28 似為另一單位,精細語意待後續對 DOSBox 確認;
不影響佈局解碼。CTY 一檔含多 section = 同城多樓層/室內場景。)

## HUD / 事件 dispatcher `sub_9530`(`re/render.c::update_hud`,file 0xa8a0)

名為 HUD,實為**移動後的事件處理 + NPC 更新 + 狀態列重繪**綜合 pass:

1. `[0x4f46]&4`(redraw 請求)→ `lcall 11dd:c`,清 bit2。
2. gate:無移動(`[0x4f1f]==0xff`)/ `[0x4f3b]<2` / `[0x4f44]&0x10`(對話鎖定)→ `ret`。
3. 清狀態列暫存 `[0x662..0x665]=0`(4 bytes)。
4. 逐位元 dispatch 事件旗標 `bp=[0x4f46]`:

| bit | 行為 |
|---|---|
| `0x800` | `sub_aa42` 踩到觸發 tile,`ret` |
| `0x20`/`0x40` | 倒數 `[0x52f6]`;歸 0 → `sub_aa7e` |
| `0x8000`/`0x4000` | 倒數 `[0x52f7]`;歸 0 → `sub_aaca` |
| `0x200` | 倒數 `[0x52f8]`;歸 0 → `sub_ab80` |
| `4` | `lcall 11dd:c`;清 bit2;`[0x13]|=1` |

   另:`[0x50b7]&4` → 倒數 `[0x504e]`,`<=0` → `call 0x5474`。
5. tile 動畫:`[0x2577]` 高 nibble!=0 → `sub_ab18`(動畫步進),否則清 `[0x4f46]` bit10。
6. NPC 迴圈:`[0x5077]` 隻,逐隻 `sub_a9be`(讀 `[bx+0x4f15]` 實體表、依精靈狀態
   `[si+0x38]` 更新);若有變更(`[0x24b4]!=0`)→ 提交 + 經文字繪製器
   `lcall 111b:264` 重繪角色名/狀態列。

## 反編譯產出與殘留

- `re/gameplay.h` — 型別(`cty_layout_hdr`)、場景旗標/方向常數、全域 DS offset、函式宣告。
- `re/render.c` — `draw_scene`(sub_255b)、`cty_select_section`(sub_30cf)、`update_hud`(sub_9530)。
- `re/player.c` — `update_player`(sub_19b8)、`player_move_collide`(sub_2dda)、`scene_event_poll`(sub_2f01)。
- `tools/re_gameplay_cty.py` — CTY section 佈局解碼 + 貼圖驗證(產 `work/cty00_sect0_verify.png`)。

殘留(下一階段):
- 地表(overworld)碰撞 `sub_3ac8` 區段反組譯對齊有偏移,僅取結構;細節待 DOSBox 對照。
- CTY section 標頭 +0x00..0x0d 的逐欄語意(目前只定名 +0x0e/0x10/0x15)。
- spawn 欄位單位(tile vs 半格 vs 像素)待執行期驗證。
- 14-entry 主迴圈狀態機 handler(指令選單/戰鬥/對話)由另一路反編譯,本文只引用不展開。
