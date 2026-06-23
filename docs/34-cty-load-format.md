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

## section/CTY 轉場機制(門/階梯)— RE 編目

踩到門/階梯 → 轉場到另一 section 或 CTY。鏈路:

1. **type-2 事件**(section 事件表,docs/31)→ handler `0x9f32`:
   `warp[param] = [param*3 + 0x4ea0]` = `{dest, X, Y}`(3 byte)→ 組成結構 `{?, cur_map, dest, X, Y}@0x2521`
   → `call 0xd1f9`(轉場執行器)。
2. **warp 表 0x4ea0**(113 個非零):多數成對(門兩端,X/Y 互換)。`dest` 多為**目標 section**
   (intra-CTY/迷宮樓層,值 1..57+);warp[35+] 的大 dest(226/242)+ 大 X(127/168)疑為**跨 CTY/地表**。
3. **轉場執行器 0xd1f9**:多 sub(0xd2a5 setup / 0xd966 判路徑 / 0xdc36 / 淡出),最後讀**待轉變數組**
   設目的並重載:
   ```
   [0x4f52]=dest CTY  [0x4f50]=dest section  [0x4f4c/4e]=spawn(X,Y)  [0x4f48/4a]=地表位置
   → lcall 0x1053:0xdc 載入(同 load_cty 路徑)
   ```
   `0x2803` 另存「當前位置」到該變數組(供返回/再進)。

## 待實作(scope:大)

- **多路徑轉場 VM** + **門編碼分歧**:section 事件表 type-2(顯式門事件)、warp 表 {dest,X,Y}、
  城鎮外圍 section0 的建築入口(sb+8 事件表=0xffff → 推測 tile-based,踩建築 tile 進對應 section)。
- remake 務實路徑:先支援「section 事件表 type-2 → warp[param] → 載入 dest section/CTY + spawn」
  (顯式門),再補城鎮外圍的 tile-based 建築入口。屬 階段③ 互動完整化的一塊。

## section header 完整欄位(0x443f 解析,已解碼)

`section_base = CTY[[0x256a]*2]`,header 欄位(相對 section_base):

| off | 內容 |
|---|---|
| +0 / +2 | **NPC/物件清單指標**(由 `[0x526c]` 選:Y<0x12c 用 +0,否則 +2)→ `{count, records}` |
| +8 | **事件表指標**(type-2 門/調べる;0xffff=無。僅 21 CTY 有)|
| +0xe | **layout 指標** → `{w(2), h(2), tiles…}`,tiles 在 +4 |
| +0x10/+0x11 | 地圖旗標 / `[0xd77]` **遭遇安全旗標**(0=可遇敵、非0=安全;A-4 洞窟遇敵 gate)|
| +0x13/+0x14 | 玩家 spawn X / Y |
| +0x15..0x18 | 地圖參數;`[0xd71]` = 地圖 id |

NPC/物件記錄 = **7 byte**:`{X, Y, b2, b3, b4, flags, b6}`(位置 = Y*寬+X;flags bit→[0x4f70] 屬性表決定顯示);
複製到 `[0xb66]` 8-byte 槽。

## 門 dispatch:type-2(21 CTY)已解,其餘 68 CTY 待動態

- **21 CTY**:建築門/階梯 = section 事件表 type-2 → warp[param]={dest section,spawn} → 已實作(A-2 核心)。
- **68 多 section CTY**(含 CTY00/02 等大城):**事件表=0xffff**,門不走 type-2;subid tile(1-4)存在但無事件表 dispatch。
  靜態追不出建築門→section 的對映(section 0 header 又與 offset 表重疊)。
  **建議動態**:DOSBox 在轉場執行器 `0xd1f9`(或 `lcall 0x1053:0xdc` 載入點)下中斷點,於城鎮中走進
  一棟建築,讀當下 `[0x256a]`(目標 section)+ 觸發 tile 座標,即可反推門→section 對映規則。
