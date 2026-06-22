# DQ3.EXE Call Map（RE → C 第一步:函式結構與 runtime API）

對主程式碼段(seg 0,file 0x1370..0x11370)線性反組譯後的呼叫普查(`tools/re_census.py`)。
目標語言已定為 **C**(可重編譯);此為反編譯的骨架。

## Runtime library:13 個段(手寫組語驅動)

主程式以 `lcall seg:off` 遠呼叫低段 runtime(共 1,030 次遠呼叫):

| segment | file off | 被呼叫次數 | 推測角色 |
|---|---|---|---|
| `0x111b` | 0x12520 | 585 | **繪圖/文字輸出**(含核心 0264,見下)|
| `0x1053` | 0x118a0 | 183 | 多功能(檔案/資料?,sub-fn 多)|
| `0x10b6` | 0x11ed0 | 122 | (待定)|
| `0x109c` | 0x11d30 | 59 | (待定)|
| `0x1104` | 0x123b0 | 50 | **鍵盤**(8042 port 直驅,已反組譯確認)|
| `0x0fe1` | 0x11180 | 14 | runtime 起點 |
| `0x11c4`/`0x11dd`/`0x1209`/`0x100a`/`0x129c`/`0x11d3`/`0x1032` | … | 1–4 | 雜項 |

中斷:`int 21h`(DOS,83×)、`int 33h`(**滑鼠**,13×)→ 遊戲支援滑鼠。

## 已識別函式

### `111b:0264`(file 0x12784)= 文字/對話繪製器 — 489 次(全程式最熱)

反組譯顯示它:

- 由 `[0x3e70]`/`[0x3e72]` 取繪製座標,x+2、y+0x10;讀計數器 `[0x259b]`×0x10 累加 y → **每行 16px 推進**(對應 16×16 字格)。
- 從資料段 `[0x252e]` 取一個 word `bx`(字碼),對控制碼分支:

| bx 值 | 語意 | 對應文字格式(docs/03) |
|---|---|---|
| `-1` (0xffff) | 終止 | 記錄結束 |
| `-2`/`-3` (0xfffe/fffd) | 換行 | `0xfffe` 換行 |
| `-4` (0xfffc) | 換頁 | `0xfffc` 換頁 |
| `-5`/`-6`/`-7` (0xfffb/fffa/fff9) | 動態插值 | 主角名/道具/金額占位 |

→ **EXE 程式碼這端完整印證了文字 agent 從 `D3TXT*.TXT` 資料推出的控制碼語意**(雙向驗證)。
此函式即遊戲把 `D3TXT*.TXT` 的字碼序列逐字繪到畫面的核心常式。

### `seg 0x111b` = VGA planar 繪圖驅動(585 次,最熱段)

段內密集 `out dx, ax` 至 VGA 暫存器;blitter primitive(111b:0006 起)為:

```asm
mov dl, 0xc4              ; DX = 0x3c4 = VGA Sequencer port
mov ds, [0x252c]         ; DS = tile/sprite 來源段
mov ax, 0x0802           ; AL=2 (Map Mask reg), AH=8 (plane 3 mask)
out dx, ax               ; 選寫入平面
  mov cx, 0x10           ; 16 列
  lodsw                  ; 取一字來源 (16px)
  and word es:[di], ax   ; AND-mask 寫入 video memory (sprite 透明遮罩)
  add di, 0x54           ; 下一列, pitch = 84
  loop
shr ah, 1; jnz 上面      ; 4 個 plane 輪流 (8→4→2→1)
```

**程式碼↔資料雙向印證**:`out 0x3c4` Map Mask + 4 plane 輪寫 = **4-bit planar** 圖形,
正是地圖 agent 解出的 BLK/SHP tile 像素格式(docs/04)。`and es:[di]` 為 sprite 透明遮罩,
row pitch = 0x54(84)。此段即遊戲所有畫面輸出(tile/sprite/字模)的底層,文字繪製器
`111b:0264` 亦在其中。

## 主程式函式熱點(seg 0)

近呼叫熱點(2,202 次近呼叫,過濾線性掃到資料的誤判):
`5002`(81×)、`6ef4`(71×)、`6f09`(67×)、`f604`(58×)、`9834`(55×)、`5023`(43×)、
`e6c9`(41×)、`683a`(38×)、`6edf`(38×)、`f590`(34×)、`c642`(34×)… 這些是主迴圈
/繪製/輸入處理的核心函式,待逐一命名。

## 素材載入

### 檔名字串池:DS:0x003d(主資料段 DS=0x14dd)

啟動碼 `mov ax,0x14dd; mov ds,ax` 設定主資料段(file base 0x16140)。23 個素材檔名以
null 結尾連續排放於 DS:0x003d 起:

```
DS:003d setup.dat   0047 player.dat  0052 dragon0.dat 005e cty00.dat
0068 d3txt00.fon    0074 d3txt00.txt 0080 item.dat    0089 dq3con.map
0094 dq3und.map     009f dq31.blk    00a8 dq3.blk     00b0 blkbm1.dat
00bb blkbm.dat      00c5 dq3mst.bls  00d0 dq3man.bls  00db dq3lin.bls
00e6 dq3.pal        00ee mnsbk.pal   00f8 d3mns.dat   0102 dq3mns.shp
010d packbg.scr     0118 mbg.mcx     0120 ebg.mcx
```

帶序號的檔名(`cty00`/`dq31`/`blkbm1`)是**模板**,載入時就地改寫數字位元組產生 `cty01`、
`dq32`、`blkbm2`…(見下)。

### BLK tile 載入函式(file 0xfe64 = seg0:eaf4)

完整的開檔→讀表頭→讀資料→關檔序列,並由旗標選地表/城鎮 block 檔:

```asm
cmp word [0x4f2d], 0        ; 0 = 地表(dq3.blk), 否則城鎮(dq3?.blk)
je  use_overworld
lea dx, [0x9f]             ; DX = "dq31.blk" 模板
mov bx, [0x256c]           ; 地圖索引
mov al, [bx+0xa04]         ; 查號碼
dec al; add al, 0x31       ; → ASCII 數字
mov [0xa2], al             ; 改寫檔名第 4 字 '?'(0x9f+3)→ dq31/dq32/...
jmp open
use_overworld:
lea dx, [0xa8]             ; DX = "dq3.blk"
open:
mov ah, 0x3d; mov al, 0; int 21h   ; DOS 開檔 (DS:DX=檔名), BX=handle
mov bx, ax
lea dx, [0x2859]; mov cx, 6; mov ah, 0x3f; int 21h   ; 讀 6-byte 表頭到 DS:0x2859
mov ds, [0x2532]; xor dx,dx; mov cx, 0xffff; mov ah, 0x3f; int 21h  ; 讀 tile 資料到遠段
mov ah, 0x3e; int 21h      ; 關檔
```

**程式碼↔資料雙向印證**:讀表頭固定 `CX=6` → 確認 BLK 表頭為 **6 bytes**(`<u16 4><u16 24><u16 count>`,
即地圖 agent 在 docs/04 解出的格式)。其餘素材沿用同一「DOS AH=3Dh 開檔 → AH=3Fh 讀 → AH=3Eh 關」樣式,
檔名模板就地改號的技巧通用(`cty00`+index、`dq3?`+index、`blkbm`+index)。

## 下一步(RE → C)

1. 反組譯並命名各 runtime 段函式 → 建 runtime API 標頭(keyboard/VGA/mouse/file/sound)。
2. 從引用素材檔名的資料位址回溯 xref,定位**素材載入函式**(開檔→讀入→解碼)。
3. 找主迴圈(由 `int 33h`、文字繪製器、輸入函式的呼叫者回溯)。
4. 逐函式反編譯為 C,以 dosbox 對照原版驗證行為。
5. 由場景繪製邏輯一併解開素材謎團:CTY 城鎮佈局、BLKATT 屬性、海面 palette cycling。

工具:`tools/re_disasm.py`(單點反組譯)、`tools/re_census.py`(呼叫普查)。
