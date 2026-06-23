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

## CTY 檔結構

- `word[0]` = section 數;`word[1..]` = 各 section 偏移(`0xffff`=空)。
- section 分兩類:
  - **地圖 section**:大 offset,header 的 `+0x0e` 是有效 layout_ptr(如 CTY00 section0=0xf50→layout)。
  - **資料 section**:小 offset,直接是 NPC/事件記錄,`+0x0e` 非 layout_ptr。
- 城鎮的入口 section 指向地圖 section;`[0x256a]` 決定哪個。

## remake 校正(dq3_town.c)

修正三處對齊 RE:
1. **tiles = layout + 4**(原誤 +8 → 整張圖位移 2 格)。
2. **spawn = section_base + 0x13/0x14**(原誤讀 layout+4/+6)。
3. layout oob 檢查改 `lay+4`。
CTY00 修正後渲染為乾淨房間(紅毯+黃磚牆+床+寶箱),驗證 tile base 正確。

## 待續:洞窟/迷宮格式

89 個 CTY 中 ~50 個(城鎮)以上述格式正確載入;**~31 個(洞窟/小祠堂/迷宮,如 CTY03/40/89/93)
所有 section 皆小 offset、無地圖 section header → 用不同的地圖格式**(非城鎮 section/layout)。
需另解洞窟/迷宮的 map 編碼(下一步 RE)。`[0x256a]` 入口 section 來源:地表入城在 file 0x39cf
(=0 重置)、warp/門進入在 0x4916/0xdbc7(由進入資料設)。
