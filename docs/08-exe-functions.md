# DQ3.EXE 函式圖、啟動鏈與主迴圈(RE → C)

對主程式碼段 seg 0 做**遞迴反組譯**(recursive-descent)的成果。線性反組譯會被夾在程式碼
中的資料(跳表、字串、座標表)帶偏、對齊錯位;遞迴反組譯只跟著 `call` / `jmp` / `jcc`
真正可達的指令走,因此能正確切出函式邊界。

工具:`tools/re_recdis.py`(docker 內跑 capstone)。位址空間 = **seg 0 邏輯 offset**
(= file offset − HDR,HDR = 0x1370)。entry 的 seg0 off = `CS:IP` = `0x9299`(file 0xa609)。

```
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_recdis.py"          # 摘要
           # python tools/re_recdis.py json docs/data/exe_funcs.json   # 完整函式圖
```

## 概況

- 從 entry + 已知入口遞迴,**切出 280 個函式、11,461 條可達指令、涵蓋 ≈35KB** 的 seg0 主碼。
- 完整函式圖(起點/大小/caller/callee/遠呼叫/資料與檔名池 xref)在
  [`docs/data/exe_funcs.json`](data/exe_funcs.json)。
- 每個素材檔名都對映到一個獨立的載入函式(下表),與 docs/06 的檔名池一致並收斂。

## 啟動鏈(更正 docs/06 的線性視角)

docs/06 由線性掃描推測「entry→0xa660→0x82bb→0xa753」。遞迴反組譯後鏈路如下
(near call 的 16-bit 目標會 wrap,例如 entry 第一指令 capstone 顯示 `call 0x102c4`,
實際 = `0x02c4`):

```
entry  seg0:9299  (file 0xa609)
  ├─ call 0x02c4                    ; 早期環境設定
  ├─ lcall 0fe1/100a/1104 …         ; runtime init(DOS/鍵盤)
  ├─ mov ds, 0x14dd                 ; 設主資料段 DS
  ├─ lcall 1104:132 (ah=0x19)       ; 進入主初始化前的鍵盤/中斷掛載
  ├─ call 0x92f0   ◄── 遊戲主體      ; 此 call 回來後即結束程式
  └─ mov ah,4Ch; int 21h            ; 程式結束(entry 末端)

sub_92f0  seg0:92f0  (file 0xa660) = 遊戲主體
  ├─ call 0x934f                    ; 影片/檔案初始化(VGA A000、載入 item/d3mns/setup/字型)
  ├─ lcall 1209/1053/0fe1 …         ; 視訊模式、檔案系統、計時
  ├─ call 0x9387                    ; 圖形初始化(out 0x3ce GFX 暫存器、調色盤、call load_blk)
  ├─ call 0x0000                    ; C runtime startup(見下)
  ├─ call 0x6f4b   (file 0x82bb)    ; 亂數種子 + 玩家名稱表(讀 BIOS 0040:006C tick)
  └─ call 0x93e3   ◄── 主迴圈        ; 此 call 回來後 sub_92f0 ret → entry → 結束
```

`sub_0000`(file 0x1370,seg0 起點)= 典型 large-model **C runtime startup**:設旗標
`[0x72a]`、`call 0xf4e3`(環境/argv 設定)、`call 0x14da`、`call 0x30`(主初始化:在
`[0x4f2f]`/`[0x4f31]`/`[0x4f48]`… 設預設座標/視窗常數)。印證目標語言為 large-model C。

`sub_6f4b`(file 0x82bb,docs 點名的大函式)實際是**亂數種子 + 名稱初始化**:讀 BIOS
資料區 `0000:046C`(系統 tick 計數器)當種子,以 `[0x8c0]+al*3` 查名稱表填 `[0x9ec..0x9ee]`。

## 主迴圈:`sub_93e3`(seg0:93e3,file 0xa753)

這是遊戲主迴圈,符合「鍵盤輸入 + 狀態機 jump table + 繪製 + 回圈」全部特徵。它是
`sub_92f0` 的最後一個 call,return 後整個程式即結束。唯一同時讀寫玩家座標 `[0x2572]`(X)、
`[0x2574]`(Y)與方向 `[0x26ad]` 的核心函式。

```
sub_93e3:
  [0x2572]=0; [0x2574]=0; [0x26ad]=0          ; 重置玩家 X/Y/朝向
  lcall 10b6:166                               ; 滑鼠/雜項初始化
  call 0x255b                                  ; 繪場景(地圖捲動繪製)
  call 0x19b8                                  ; 玩家精靈/動畫更新
  lcall 10b6:19c
loop(0x9405):                                  ◄── 主迴圈頂
  [0x2572]=0; [0x2574]=0; [0x4f1f]=0xff        ; 清本幀位移/方向 = 無
  lcall 1104:8a            → AH = 鍵盤 scancode  ; 讀鍵盤
  cmp ah,0; je draw                            ; 無輸入 → 直接繪製
  switch(ah):
    0x50 (↓): inc [0x2574]; [0x26ad]=0; [0x4f1f]=0
    0x4b (←): dec [0x2572]; [0x26ad]=2; [0x4f1f]=1
    0x48 (↑): dec [0x2574]; [0x26ad]=4; [0x4f1f]=2
    0x4d (→): inc [0x2572]; [0x26ad]=6; [0x4f1f]=3
    0x3e     : lcall 11dd:c            ; 選單/特殊鍵
    default  : 14-entry 鍵碼表查找 (cx=0xe)
               bx=0; while([bx+0x19]!=ah) bx++ ;  ; scancode 表 @ [0x19]
               if([bx*2+0x25]!=0) call [bx*2+0x25] ; far 函式指標表 @ [0x25] = 狀態機 dispatch
draw(0x94b0):
  call 0x255b              ; 繪場景(地圖捲動)
  call 0x19b8             ; 玩家精靈/動畫 + 碰撞(用 [0x4f1f] 方向走 jump table [bx+0x255a])
  call 0x9530             ; 狀態/HUD 更新(test [0x4f46]&4 …)
  call 0x991d             ; 子系統更新
  call 0xee23             ; 計時/節拍(frame wait)
  jmp loop                 ◄── 回圈頂
```

要點:
- **輸入**:`lcall 1104:8a` 取鍵盤 scancode(seg 0x1104 = 鍵盤驅動,docs/06)。
- **方向碼雙寫**:`[0x26ad]`(0/2/4/6 = 下/左/上/右,8 朝向系統)供精靈方向,
  `[0x4f1f]`(0/1/2/3)供 `sub_19b8` 的移動 jump table。
- **狀態機**:方向鍵之外的鍵走 14-entry **跳表**(scancode 表在偏移 0x19、far 函式指標表
  在偏移 0x25),`call word ptr [bx+0x25]` 即指令選單/動作派發。
- **繪製管線**:`0x255b`(場景)→ `0x19b8`(玩家+碰撞)→ `0x9530`(HUD)→ `0x991d` →
  `0xee23`(節拍)。`sub_19b8` 依 `[0x4f2d]`(0=地表 / 否則城鎮)決定捲動方式,並用
  `call word ptr [bx+0x255a]` 依方向呼叫四個移動/碰撞函式之一。
- 緊接 `sub_93e3` 之後的 `sub_94c3` 是一個變體:不讀鍵盤、改以 `[0x4f1f]` 重播上次方向,
  作「自動移動一步」(碰撞解算後續步 / 腳本移動)用。

### 主迴圈會踩到的關鍵資料變數(主資料段 DS=0x14dd)

| DS off | 含義(推定) |
|---|---|
| `0x2572` | 玩家本幀 X 位移 / 座標累加 |
| `0x2574` | 玩家本幀 Y 位移 / 座標累加 |
| `0x26ad` | 玩家朝向(0/2/4/6 = 下/左/上/右) |
| `0x4f1f` | 本幀移動方向(0/1/2/3,0xff=無) |
| `0x4f2d` | 場景旗標(0=地表 overworld / 否則城鎮),也決定 BLK 檔選擇(docs/06) |
| `0x4f46` | 狀態旗標(`sub_9530` test bit 2) |
| `0x256c` | 地圖索引(BLK 載入器查號用,docs/06) |

## 素材載入函式表(檔名池 → 載入函式)

遞迴反組譯把每個檔名池條目(DS:0x3d 起,docs/06)各自收斂到一個載入函式:

| 載入函式(seg0) | 檔案 | 備註 |
|---|---|---|
| `sub_155e` | dragon0.dat | 龍/敵資料 |
| `sub_1778` | player.dat | 玩家資料 |
| `sub_17cb` | item.dat | 道具 |
| `sub_17ee` | d3mns.dat | 選單資料 |
| `sub_1811` | d3txt00.fon / d3txt00.txt | 對話字型 + 文字(成對載入) |
| `sub_1865` | d3txt00.txt | 文字(換頁/換章重載) |
| `sub_2c19` | dq3con.map / dq3und.map | 地表 / 地下大地圖 |
| `sub_3036` | cty00.dat | 城鎮(就地改號 cty01…) |
| `sub_997d` | setup.dat | 設定 |
| `sub_b19e` | dq3mns.shp | 選單 sprite |
| `sub_c6e5` | packbg.scr | 大背景(3.7MB,串流) |
| `sub_eaf4`(load_blk) | dq31.blk / dq3.blk / blkbm1.dat | BLK tile 載入器(docs/06) |
| `sub_ec9e` | dq3mst.bls / dq3man.bls | BLS 音樂 |
| `sub_ecc0` | blkbm.dat | tile 屬性/bitmap |
| `sub_ecdc` | dq3.pal / mnsbk.pal | 調色盤 |

## 熱點函式(in-degree top,待逐一命名)

| 函式 | size | callers | 推定角色 |
|---|---|---|---|
| `sub_f604` | 14 | 24 | 小工具(thunk) |
| `sub_f590` | 116 | 22 | 常用繪製/座標換算 |
| `sub_e6c9` | 30 | 21 | 常用工具 |
| `sub_c642` | 49 | 19 | 常用工具 |
| `sub_9834` | 14 | 14 | 小工具 |
| `sub_ed39` | 114 | 11 | 載入相關 |
| `sub_eaf4` | 170 | 6 | BLK 載入(已命名) |

最大函式:`sub_99dc`(size 3320,但只有 89 指令 → 函式體內含跨大段資料的 jmp,size 被
高估,內嵌資料/跳表;非真正 3KB 程式碼)、`sub_ac05`(961B,321 指令,23 次遠呼叫,
密集 runtime 互動)。

## 反編譯產出(re/)

- [`re/exe.h`](../re/exe.h) — runtime API far-call 原型、主資料段變數位址、函式宣告。
- [`re/startup.c`](../re/startup.c) — entry / C-runtime / 初始化骨架。
- [`re/mainloop.c`](../re/mainloop.c) — 主迴圈骨架(輸入 → 移動 → 狀態機 → 繪製 → 節拍)。

C 風格為 **real-mode large-model**:跨段資源以 far 指標表示,runtime 以 far call 表達。
求結構正確、可審閱,非位元精準重編譯版(逐函式對 DOSBox 驗證為後續步驟)。

## 下一步

1. 命名 14-entry 狀態機跳表的各 far 函式(指令選單/戰鬥/對話入口)。
2. 反編譯 `sub_255b`(場景繪製)與 `sub_19b8`(玩家移動+碰撞)細節 → 解 BLKATT 屬性語意。
3. 把 in-degree 熱點 thunk(`sub_f604`/`sub_f590`/`sub_e6c9`)命名為共用工具,降低噪音。
4. 逐函式反編譯,對 DOSBox 原版逐步驗證行為。
