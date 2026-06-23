# load_cty 載入邏輯與 CTY 地圖格式(深挖 RE)

座標 → 事件 → 載入城鎮/洞窟 的完整鏈,逐段反組譯確認;並校正 remake 的 `dq3_town.c`。

## 鏈路

1. **座標 → CTY 號**:地表走到入口 → `load_cty` 入口查表(file 0x43b7,DGROUP `0x748`,
   100×4 `{X,b1,Y(u16)}`)→ 命中 entry index = CTY 號 → `[0x256c]`(docs/13/32)。
2. **開檔 + 載 BLK**:`load_cty`(file 0x43d1)開 `CTYNN.DAT` → 讀進段 `[0x2536]` → `load_blk`
   (file 0xfe64)依 `map_blknum`(DGROUP `0x0a04`,`[0x256c]` 索引)選 `DQ3?.BLK` tile 圖庫。
3. **section/layout 解析**(file 0x443f):
   ```
   section_base = CTY[ [0x256a] * 2 ]          ; [0x256a] = 入口 section(每進入點不同,非固定)
   layout_ptr   = CTY[ section_base + 0x0e ]    ; 相對 section_base
   layout(絕對) = section_base + layout_ptr
     width  = layout[0]   height = layout[2]
     tiles  = layout + 4                         ; ★ 緊接 w,h 之後
   spawn_X = CTY[ section_base + 0x13 ]          ; ★ 在 section header,不在 layout
   spawn_Y = CTY[ section_base + 0x14 ]
   ```
   tile = u16:低 byte=BLK index、高 byte=屬性/事件 subid(docs/31)。

## CTY 檔結構(★ 無 count 前綴)

- **`word[i]` = section i 的 base offset**(`word[0]`=section 0,`0xffff`=空)。**沒有** count 欄位。
- `section_base = CTY[ [0x256a]*2 ]`(= `word[[0x256a]]`)。
- **地表入城 `[0x256a]=0` → section 0 = 從地表進入看到的那層**(城鎮外圍 / 洞窟入口室)。
  其餘 section(1,2…)= 室內房間 / 子區。
- 每個 section 自帶地圖 header(`+0x0e`=layout_ptr、`+0x13/0x14`=spawn);**城鎮與洞窟同格式**,
  差別只在大小與 section 數(城鎮多房間、洞窟常 1-2 section)。
- 例:CTY00 section0=word[0]=0xc→42×43(阿里阿罕全鎮);CTY93 section0=word[0]=2→17×17(合成祠堂)。

## remake 校正(dq3_town.c)— 三個錯,修正後 89/89 全載入

1. **section i = `word[i]`**(原誤把 `word[0]` 當 count、從 `word[1]` 起算 `2+2*section` →
   **漏掉 section 0(地表入口層)、整個差一格**,城鎮被渲成室內房間、洞窟/祠堂全 oob)。← 主因
2. **tiles = layout + 4**(原誤 +8 → 圖位移 2 格)。
3. **spawn = section_base + 0x13/0x14**(原誤讀 layout+4/+6)。

修正後實測:**89 個 CTY 全部正確載入**(原 50)。CTY00=阿里阿罕全鎮、CTY93=17×17 合成祠堂
(紅地板+柱列邊框+中央祭壇,正確渲染);headless game 進 CTY93 → #2 合成 result=0x75。

## `[0x256a]` 入口 section 來源

- 地表入城(0x2c37)→ `[0x256a]=0`(section 0)。
- 新遊戲阿里阿罕(0x13be)→ `[0x256a]=4`。
- warp/門(0x4916/0xdbc7)→ 由進入資料設(門/階梯指定目的 section)。
- 旗標分支地點(0x396e)→ 依劇情旗標選不同 CTY(隨進度變的地點)。
