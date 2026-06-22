# 地圖 / Tile 素材格式與還原

精訊版 DQ3 的地圖素材由四類檔案構成:調色盤(PAL)、tile 圖庫(BLK)、tile 屬性(BLKATT)、地圖佈局(MAP / CTY)。本文記錄已還原的格式與工具。

render 成果在 `docs/maps/`:
- `world_con.png` / `world_con_full.png` — 地表世界地圖(continent)
- `world_und.png` / `world_und_full.png` — 地下世界地圖(underground)
- `sheet_DQ3.png` — 地表 tile 圖庫(162 tile)
- `sheet_DQ31.png`~`sheet_DQ33.png` — 城鎮 / 室內 tile 圖庫(各 170 tile)

工具:`tools/tile_lib.py`(共用解碼)、`tools/tile_sheet.py`(tile 圖庫)、`tools/map_render.py`(世界地圖)。
一律用 `tools/dockrun.sh tools/xxx.py ...` 在 docker uv venv 內執行。

## DQ3.PAL — VGA 調色盤

| 欄位 | 值 |
|---|---|
| 大小 | 240 bytes = 80 色 |
| 每色 | 3 bytes,順序 R, G, B |
| 值域 | 0–63(VGA 6-bit DAC) |

轉 8-bit RGB:`v8 = (v6 << 2) | (v6 >> 4)`(等效 *255/63,亮部更準)。

`MNSBK.PAL`(480 bytes = 160 色)為怪物 / 背景用調色盤,結構相同但內容不同,**不適用**於地表 / 城鎮 tile。

### 海面 palette cycle(重要)

地表的海面 fill tile(index 88)使用 palette 索引 2 與 10。靜態 `DQ3.PAL` 把這兩格存成沙色
(idx 2 = `(255,219,158)`、idx 10 = `(158,117,0)`),但遊戲執行時由 EXE 改寫 DAC,
顯示成藍色(典型 DQ palette animation / 海浪流動效果)。海岸線 tile(如 index 2、75)
另外用 idx 5 = `(0,85,223)` 真藍,因此即使套靜態 palette 海岸也正確。

`map_render.py` 預設把 idx 2/10 改藍(可讀性);設環境變數 `OCEAN_FIX=0` 可看原始 palette。
精確的 runtime palette 仍在 `DQ3.EXE` 內(尚未抽出),未影響地圖可讀性。

## BLK — tile 圖庫

| 檔案 | 大小 | 內容 |
|---|---|---|
| `DQ3.BLK` | 62214 | 地表 tile(162 個):草原、海、山、森林、沙漠、城鎮 / 城堡外觀、橋、船 |
| `DQ31.BLK`~`DQ35.BLK` | 65286 | 城鎮 / 室內 tile(各 170 個):磚牆、寶箱、床、桌椅、火爐、樓梯、石地板、雕像 |

### Header(6 bytes,little-endian)

| offset | 型別 | 意義 |
|---|---|---|
| 0 | u16 | 每 plane 每 row 的 byte 數 = 4(→ 寬 32 px) |
| 2 | u16 | 高 = 24(row 數) |
| 4 | u16 | tile 數量(count) |

### Body

`count` 個 tile,每 tile **384 bytes**,格式 = **32×24 像素、4-bit planar(平面)**:

- 4 個 plane 連續排列,每 plane 96 bytes(24 row × 4 byte/row)。
- 每 byte = 8 像素的 1 個 bit-plane,MSB 先(bit 0x80 = 最左像素)。
- 像素色號 = `(plane0_bit) | (plane1_bit<<1) | (plane2_bit<<2) | (plane3_bit<<3)`,值域 0–15。

驗證:`(65286-6)/170 = 384`、`4 byte/row × 8 × 24 row = 32×24 px`、`32×24×4bit/8 = 384`。
此為 4 個其他排列假設中唯一 render 出可辨識地形的解(見 git 歷史的 probe)。

## DQ3CON.MAP / DQ3UND.MAP — 世界地圖佈局

| 檔案 | 大小 | 尺寸 |
|---|---|---|
| `DQ3CON.MAP` | 50024 | 244 × 205 tile(地表) |
| `DQ3UND.MAP` | 40752 | 244 × 167 tile(地下) |

### Header(4 bytes,little-endian)

| offset | 型別 | 意義 |
|---|---|---|
| 0 | u16 | 寬(tile 數) |
| 2 | u16 | 高(tile 數) |

### Body

`width × height` 個 byte,**每 byte 1 個 tile index**(row-major,由上到下、由左到右),
index 對應 `DQ3.BLK` 的 tile。地表用到 107 種 tile、最大 index 160(< 162);
最常見 index 88 = 海面(佔 30208 格)。index 0 為地圖邊界外的填充(地下世界四周可見黑邊)。

驗證:`244×205+4 = 50024` ✓、`244×167+4 = 40752` ✓。

## BLKATT.DAT — tile 屬性表(部分解析)

明碼 ASM 片段,開頭 `blk_att_tab`,內容為一系列 `db <8-bit binary>b, <8-bit binary>b` 行
(如 `db 00000000b,00000000b`)。每 tile 對應 2 個 bit-mask byte,編碼**可通行 / 阻擋 /
觸發**等屬性(每個 bit 一種屬性旗標)。這是原始碼匯出的資料表,逐位元語意需對照 EXE 中
讀取此表的邏輯(尚未對齊到具體旗標),暫記為「每 tile 2 byte 屬性 mask」。

## CTY00~87.DAT — 城鎮 / 場景(已解)

CTY 為**多段容器(container)**。先前以為 tile 佈局是 RLE/壓縮而卡關;經反組譯 `DQ3.EXE`
的 CTY 繪製邏輯(`sub_30cf`,見 [`docs/11-exe-gameplay.md`](11-exe-gameplay.md))確認**並非壓縮**——
卡關是因漏了 section 偏移表的**間接定址**,且把 tile 當 u8(實為 u16)。正確結構:

```
+0      u16 n                      ; section 數
+2      n × u16 section_off        ; 各 section 偏移; 0xffff = 空
section_off + 0x0e:  u16 layout_ptr        ; 版面資料指標 (相對 section_off)
section_off + layout_ptr:
        u16 width
        u16 height
        4 bytes spawn              ; 進入點 / 出生座標
        width × height × u16 tile  ; 低 byte = BLK tile index, 高 byte = 屬性
```

每格 tile 為 **u16**:低位元組 = `DQ31~35.BLK`(室內)或 `DQ3.BLK`(戶外段)的 tile index,
高位元組 = 屬性(見下 BLKATT)。section 標頭 `+0x00..0x0d` 其餘欄位語意、spawn 單位待 DOSBox 對照。

**驗證**:`tools/re_gameplay_cty.py`(單段)、`tools/map_cty_gallery.py`(整鎮)依此格式 + `DQ3.PAL` +
對應 BLK 貼出城鎮。CTY00(阿里阿罕,5 段)render 為 [`maps/cty00_town.png`](maps/cty00_town.png):
室內房間(床/寶箱/紅地毯/磚牆)、城堡王座廳與庭院綠地皆清晰可辨,佈局解碼正確。

> 註:同一 CTY 的不同 section 可能用不同 BLK 圖庫(室內 `DQ31~35.BLK` vs 戶外 `DQ3.BLK`);
> 正確 BLK 由 section 標頭欄位選定(精確欄位待定,gallery 工具以合理性檢查過濾)。

### BLKATT.DAT — tile 屬性(已解語意)

`blk_att_tab` 匯出表,每 tile 的 u16 屬性(執行期載入自 `blkbmN.dat` 到 DS:0x308e):
`bit0`=阻擋牆、`0xe000`=事件地形、`bit5(0x20)`=可走且觸發、`bits6-7(0xc0)`=角落/出口。
詳見 [`docs/11-exe-gameplay.md`](11-exe-gameplay.md)。

## 還原方法摘要

1. **palette**:`load_pal('DQ3.PAL')`,6-bit → 8-bit。
2. **tile**:`read_blk` 取 header + tile body;`decode_tile` 做 4-plane planar → 32×24 色號陣列。
3. **世界地圖**:`read_map` 取寬高 + index 陣列,逐格貼對應 tile。
4. **城鎮**:容器偏移表已可拆段;tile 佈局命令串待續解。
