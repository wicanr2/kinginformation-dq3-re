# DQ3.EXE 素材載入函式(RE → C)

把 DQ3.EXE seg0 內的素材載入函式反編譯成可讀 C(`re/loaders.c`、`re/assets.h`),
並與已解的素材格式(docs/02 字型、docs/03 文字、docs/04 地圖)雙向印證。

座標系:seg0 程式碼 file off = `0x1370`(HDR)+ seg0_off;主資料段 DS=`0x14dd`
(file base `0x16140`),DS:offset → file = `0x16140 + off`。

## 通用載入樣式(real-mode DOS int 21h)

所有 loader 共用:
```
AH=3Dh 開檔(DS:DX=檔名字串, AL=0 唯讀) → BX=handle;CF=1 表失敗
AH=3Fh 讀  (BX=handle, CX=bytes, DS:DX=buffer);CX=0xffff = 讀到 EOF
AH=3Eh 關
```
帶序號檔名(`cty00`/`dq3?.blk`/`blkbmN`/`d3txtNN`/`dragonN`)是**模板**:載入時就地
改寫檔名字串裡的數字位元組產生實際檔名,再開檔。大型素材讀進 runtime 分配的**遠段**
(`mov ds,[segvar]` 後 `int 21h`);小型固定素材直接讀進主資料段欄位。

## 檔名字串池 ↔ open site ↔ loader 對照

檔名池在 DS:0x003d 起 23 筆(docs/06)。逐一掃 `mov ah,0x3d`(開檔,20 處)回溯各
loader,並用池內字串確認對應檔案:

| seg0 (file) | pool off | 檔案 | 載入目標 | 讀法 |
|---|---|---|---|---|
| `eaf4` (0xfe64) | 0xa8/0x9f | dq3.blk / dq3?.blk | 6B 表頭→DS:0x2859;body→seg_blk | 表頭 6B + EOF |
| `eb9e` (0xff0e) | — | (.bls 分頁讀子程式) | seg_blspage:SI | seek + 0x3c0B |
| `3061` (0x43d1) | 0x5e | ctyNN.dat | seg_cty,後接 load_blk | 整檔 EOF |
| `1811` (0x2b81) | 0x68 | d3txt00.fon | seg_fon(DS:0x252c) | 整檔 EOF |
| `1837` (0x2ba7) | 0x74 | d3txt00.txt | seg_txt(DS:0x252e) | 整檔 EOF |
| `186c` (0x2bdc) | 0x74 | d3txtNN.txt | seg_txt:0x4e20 | 整檔 EOF |
| `17cb` (0x2b3b) | 0x80 | item.dat | DS:0x1f9 | 整檔 EOF |
| `1778` (0x2ae8) | 0x47 | player.dat(讀) | DS:0x128 | 200B(0xc8) |
| `17ae` (0x2b1e) | 0x47 | player.dat(寫) | DS:0x128 | 3Ch 建檔+200B |
| `155e` (0x28ce) | 0x52 | dragonN.dat | DS:0x4f29 | (0x57a5-0x4f29)B |
| `2c19` (0x3f89) | 0x89/0x94 | dq3con/und.map | 只快取 handle | 不讀(後續分頁) |
| `eca0` (0x1000e) | 0xc5/0xd0 | dq3mst/man.bls | 只快取 handle | 不讀(分頁) |
| `ecc0` (0x10030) | 0xbb | blkbm.dat | DS:0x2eea | 整檔 EOF |
| `ecdc` (0x1004c) | 0xe6/0xee | dq3.pal / mnsbk.pal | DS:0x3232 / 0x3352 | 240B / 480B |
| `b19e` (0xc50e) | 0x102 | dq3mns.shp | 只快取 handle | 不讀(分頁) |
| `c6e5` (0xda55) | 0x10d | packbg.scr | seg_packbg | 分頁 seek + 0x6e00B |
| `997d` (0xaced) | 0x3d | setup.dat | 6B→DS:0x63a 旗標 | 6B |

> 修正 docs/06:0xaced 開的是 **setup.dat**(pool 0x3d),非 CHINA.FON;EXE 本體未見
> 開 `CHINA.FON`(母字庫應由字型烘製流程外部使用,遊戲執行期只用 D3TXT00.FON)。

## 與素材格式的雙向印證

| loader 行為 | 印證的格式(docs) |
|---|---|
| BLK 固定讀 `CX=6` 表頭 → `row_bytes/height/count` | docs/04 BLK 表頭 = 6 bytes ✓ |
| dq3.pal 讀 `0xf0`=240B、mnsbk.pal 讀 `0x1e0`=480B | docs/04 PAL=80色×3、MNSBK=160色×3 ✓ |
| d3txt00.fon / .txt 直接整檔載入、txt 用 glyph index 引 fon | docs/02/03 TXT 以 2B glyph index 引 fon ✓ |
| is_town_flag 0/1/2/3 決定 dq3.blk vs dq3?.blk | docs/04 地表 vs 城鎮 tile 圖庫切換 ✓ |
| packbg.scr 用 0x3ce/0x3c4 VGA planar 暫存器解碼 | docs/04 BLK/SHP 同 4-bit planar 家族 ✓ |
| player.dat 讀 200B、開檔失敗清零=新遊戲 | (存檔格式,200B 固定;待解語意) |

## BLK 載入函式(seg0:eaf4)細節

完整流程已反編譯為 `load_blk()`:

1. `is_town_flag`(DS:0x4f2d)非 0 → 城鎮:`fn_blk_town[3]` 改成
   `(map_blknum[cur_map_idx]-1)+'1'`,開 `dq3?.blk`;否則開地表 `dq3.blk`。
2. 開檔失敗 → 回 -1(`mov al,0xff; ret`)。
3. 讀 6B 表頭 → `blk_header_raw`(DS:0x2859)。
4. `dos_read_far(seg_blk,0,0xffff)`:tile body 讀進遠段 `seg_blk`(DS:[0x2532])到 EOF。
5. `is_town_flag!=1`:額外載兩塊固定 BLKBM page(`load_blk_page` 索引 0x12/0x13;
   子程式 0xff0e 從 `dq3man.bls` 依 page 算 32-bit 檔位、seek、讀 0x3c0B)。
6. `is_town_flag!=0`(城鎮):改寫並開 `blkbmN.dat`,讀進 DS:0x308e。

`load_blk_page`(0xff0e)偏移算式:`base = 6 + 0x3c0*bp`、`pos = ax*0x0f00 + base`,
AH=42h LSEEK 後讀 0x3c0(960)B 進 `seg_blspage`(DS:[0x2534])。

## CTY → BLK 連動(seg0:3061)

`load_cty(map_idx)`:設 `is_town_flag=1`、`cur_map_idx=map_idx`,把 map_idx 拆兩位
十進位改寫 `cty00` 的 `00`(DS:0x61/0x62),整檔讀進 `seg_cty`(DS:[0x2536]),關檔後
**立即呼叫 `load_blk()`** ——城鎮 tile 圖庫隨城鎮切換重載。這解釋了 docs/04 中「CTY 佈局
需配合對應 DQ3N.BLK」的成對關係。

## 殘留 / 待解

- `seg_*` 遠段值由 startup 以 `mov [segvar],ax` 寫入(runtime 堆積分配),非靜態常數;
  C 端以具名 `uint16_t` 抽象,實際段值待 startup 反編譯(docs/08,他人負責)補。
- player.dat 200B、dragon save `(0x57a5-0x4f29)`B 的**內部欄位語意**未解(僅確認大小與位置)。
- `.bls`(dq3mst/man/lin)、`dq3mns.shp`、`*.map` 為**分頁 seek 讀**:本文解出開檔/快取
  handle 與單頁讀子程式;完整分頁索引(page→file offset 表)待後續。
- packbg.scr 解交錯後段(0xdab0 起的 0x3c4 plane 輪寫)未逐行轉 C,僅標出 planar 性質。
- `mbg.mcx`/`ebg.mcx`/`dq3lin.bls`/`d3mns.dat`/`first.scr`/`TITA.P` 等檔在池中但本輪
  未對到明確 open site(可能由 SETUP/BAS.EXE 或尚未掃到的路徑載入),待補。

## 工具

- `tools/re_loaders.py` — `disasm`/`seg0`/`findopen`/`xref` 子命令(capstone)。
- `tools/dockrun_cap.sh` — docker uv venv 跑 capstone 反組譯(不污染 host)。
  用法:`tools/dockrun_cap.sh tools/re_loaders.py findopen`。
