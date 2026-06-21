# 資產 Inventory 與格式偵察

原版 `dq3.zip` 共 194 個檔、約 8.19 MB（日期戳 2021-12-11，內含 `.jsdos/` 表示曾打包為 js-dos 瀏覽器 DOS 模擬）。以下依類型整理，並記錄初步格式判讀。判讀為偵察階段推測，後續迭代會逐一驗證。

## 執行檔與工具鏈（.EXE）

| 檔案 | 大小 | 判讀 |
|---|---|---|
| `DQ3.EXE` | 115,282 | **主程式**。檔頭 `4D 5A 50` = `MZP` = Borland Pascal 7.0 DPMI 保護模式執行檔。RE 主目標。|
| `SETUP.EXE` | 4,542 | 設定程式（音效卡 / 顯示設定）|
| `PAS.EXE` | 3,989 | 精訊自製小工具（疑似 pack / 資料處理）|
| `FIN.EXE` | 2,656 | 同上 |
| `BAS.EXE` | 2,203 | 同上 |

`DQ3.ORI`（95,188）疑為 `DQ3.EXE` 的未加殼 / 早期版本，對反組譯有對照價值。

## 字型（.FON）— 精訊中文技術核心

| 檔案 | 大小 | 判讀 |
|---|---|---|
| `CHINA.FON` | 192,200 | **中文字型**。開頭為 22-entry 32-bit offset table（0x00–0x57），其後為 22 個字模 section。section 0（0x58 起）含一張 16-bit 子表，疑為字碼→字模索引；多個 section 大小為 72 的倍數（72 bytes = 24×24 點陣），研判為 24×24 中文點陣字。|
| `D3TXT00.FON` | 47,232 | ASCII / 半形字型，字模直接排列（開頭 `07 f0 0c 18…` 即 glyph row）|
| `SETTXT.FON` | 6,368 | SETUP 用字型 |

`*.CF1` / `*.CF2`（各 39 / 5 bytes，對應 `D3TXT00` 與 `SETTXT`）疑為字型設定 / 編碼對照小檔。

## 文字腳本（.TXT）

`D3TXT00.TXT`～`D3TXT11.TXT`（12 檔，約 73 KB），研判為**遊戲對話 / 劇情文字**，需搭配 `CHINA.FON` 字碼索引解碼。未發售中文版的中文劇情挖掘重點。

## 地圖（.BLK / .DAT / .MAP）

| 檔案 | 大小 | 判讀 |
|---|---|---|
| `DQ3.BLK` / `DQ31~35.BLK` | 62~65 KB | 地圖 block tile 資料。共同檔頭 `04 00 18 00 …`（疑為 4×24 或尺寸欄位）|
| `BLKATT.DAT` | 2,791 | tile 屬性表，明碼開頭 `blk_att_tab`（可通行 / 觸發等屬性）|
| `CTYxx.DAT` | ~113 檔 | 城鎮 / 場景資料（CTY = city），編號 00~87 |
| `BLKBMx.DAT` | 340~400 | block bitmap，疑為 tile 點陣 |
| `DQ3CON.MAP` / `DQ3UND.MAP` | 50K / 40K | 地表（CON＝continent？）與地下（UND＝underground）地圖 |
| `*.BLS` | 46~222 KB | block set / 場景集合（待驗）|

## 圖檔 / 螢幕（.SCR / .SHP / .PIC / .PAL）

| 檔案 | 大小 | 判讀 |
|---|---|---|
| `PACKBG.SCR` | 3,738,880 | 封包背景圖集（PACK + BG）。開頭 `FF FF…`，疑 RLE 或多圖封包 |
| `FIRST.SCR` | 112,000 | 開場畫面 |
| `DQ3MNS.SHP` | 1,377,210 | sprite / shape 集（MNS＝monsters？怪物圖）。開頭為 32-bit offset table |
| `DFP.PIC` | 24,435 | 單張圖，開頭為 offset table |
| `DQ3.PAL` / `MNSBK.PAL` | 240 / 480 | VGA 調色盤，值域 0–0x3F（6-bit DAC）。240/3 = 80 色、480/3 = 160 色 |

## 音樂 / 音效（.LIB / .MCX / .VCX / .MAP）

| 檔案 | 大小 | 判讀 |
|---|---|---|
| `SBCM.LIB` | 35,840 | Creative Sound Blaster 音樂驅動（字串：`Copyright (c) Creative Technology Pte Ltd., 1989-1990`）|
| `SBCM.MAP` | 1,422 | 對應 link map |
| `MBG.MCX` / `EBG.MCX` | 127K / 3K | 背景音樂（MCX，疑 CMF / 自訂格式）|
| `FVOC.VCX` / `NVOC.VCX` | 72K / 143K | 語音 / 音效（VOC 衍生）|

## 其他

`*.RPB` / `*.RPL` / `*.BRP`（DFP 系列）與 `DFP.LST`、`DFP.PIC` 同組，疑為某子系統（DFP＝？）的資料 + 清單 + 圖。`DRAGONx.DAT`、`GRADE.DAT`、`DOGDT.DAT`、`TITx.P`（20 檔 .P）等待後續分類。

## 待驗證清單（下一輪起逐一攻破）

1. `CHINA.FON` — render 24×24 字模確認格式、建字碼索引 → **進行中**
2. `D3TXT*.TXT` — 搭配字型解碼中文劇情
3. `DQ3*.BLK` + `BLKATT.DAT` — 還原地圖
4. `DQ3.EXE` — 剝 DPMI loader、反組譯 Pascal
