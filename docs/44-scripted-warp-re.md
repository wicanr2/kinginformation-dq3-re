# Scripted Warp / 轉場 / 下降系統 — 逐段反組譯記錄

> 把「世界怎麼串起來」(門/階梯/傳送/旅の扉/下降)的反組譯**逐段帶註解記錄**,避免散落遺忘。
> 位址一律 **file offset**(`tools/re_disasm.py` 印的);logical = file − 0x1370。
> 相關:docs/35(轉場/事件)、docs/42(NPC 子型)、docs/43(可達性 + 機制總結)。

## 0. DOSBox 動態分析可行性(已評估)

**不可行**做 debugger 中斷/讀記憶體:環境只有 `dq3-dosbox`(普通 dosbox)+ Xvfb + xdotool 打鍵 + 截圖,
無 dosbox-x debugger。截圖只能「確認行為」、抽不出 warp 參數,且盲打鍵走到旅の扉路徑長又脆弱。
⇒ 走純靜態 RE。

## 1. 轉場分派器 0x488f(門/階梯/出城/下降的入口)

走到 attr `0xe000` 轉場 tile → 移動後 `call 0x488f`。讀 section header `+0xc` 轉場表,
4-byte 項 `{destCTY, destSec, X, Y}`,由 tile 高 byte subid 索引。

```
; 0x488f 讀轉場項後,以 destSec(ah)分派:
08d7?  ...
088d7  cmp  ah, 0xff      ; destSec==0xff → 出地表(0x49f9)
088da  je   0x48ee
088dc  cmp  ah, 0xfe      ; ★ destSec==0xfe → 下層特殊
088df  jne  0x48f1
088e1  cmp  [0x5051], 3   ; 已是特殊層?
088e6  je   0x48ee
088e8  mov  [0x5051], 1   ; ★ 設層旗標=1(下降)
088ee  jmp  0x49f9        ; → 出 overworld handler
088f1  cmp  al, [0x256c]  ; destCTY==當前CTY?→ 同CTY換section / 否則跨CTY
```

**結論**:`0xfe` = 下降標記(設 `[0x5051]=1`),`0xff` = 普通出地表。

## 2. 出 overworld handler 0x49f9(+ Y+300 下降)

```
049f9  push es
049fa  mov  ax,[0x2536]   ; seg_cty
049fd  mov  es,ax
049ff  mov  dx,[0xb24]    ; section_base
04a03  mov  di,0xc
04a06  add  di,dx
04a08  mov  si,es:[di]    ; 轉場表 ptr
...    (以 subid 取轉場項 si)
04a1c  mov  al,es:[si+2]  ; spawn X
04a20  mov  bl,es:[si+3]  ; spawn Y
04a29  cmp  ax,0          ; spawn(0,0)?
04a31  jne  0x4a3a
04a33  mov  ax,[0x5053]   ; (0,0)→用存的地表位置
04a36  mov  bx,[0x5055]
04a3a  mov  [0x4f2f],ax   ; overworld X
04a3d  mov  [0x4f31],bx   ; overworld Y
04a41  cmp  [0x5051],1    ; ★ 層旗標==1 或
04a46  je   0x4a4f
04a48  cmp  [0x5051],3    ;    ==3 →
04a4d  jne  0x4a55
04a4f  add  [0x4f31],0x12c ; ★★ Y += 300 = 切到下層 overworld
04a55  ...
```

**結論**:下層 = overworld Y+300 同座標空間;`[0x5051]` 層旗標決定是否 +300。

## 3. 層由 Y 反推地圖檔(0x3fac)+ 兩圖檔常開(0x3f89)

```
; 0x3f89:開兩個 overworld 地圖檔,檔柄常駐
03f91  lea dx,[0x89]      ; "dq3con.map"(小寫,DGROUP 0x89)
03f99  int 0x21 (AH=3D open) → [0x1f1]=con 檔柄
03f9e  lea dx,[0x94]      ; "dq3und.map"(DGROUP 0x94)
03fa6  int 0x21 → [0x1f3]=und 檔柄
; 0x3fac:依玩家 Y 選讀哪個檔
03fac  mov bx,[0x2570]    ; 玩家 Y
03fb0  cmp bx,0x12c       ; ≥300?
03fb4  jl  0x3fc1         ; <300 → con(地表)
03fb6  sub bx,0x12c       ; ≥300 → Y−300
03fba  mov bp,[0x1f3]     ;        用 und(下層)檔柄
```

> 0x37f5 是**依 Y 同步層旗標**(`if Y≥300 [0x5051]=1 else 0`),非觸發。

## 4. type-2 examine 傳送 0x9f32(★ 已接 remake,已驗證)

section `+8` examine 事件 type-2 → warp 表 0x4ea0:

```
09f32  mov bx,dx          ; dx=param
09f38  shl bx,1; add bx,dx ; bx=param*3
09f3c  mov ax,[0xd71]     ; 當前 map id
09f3f  mov [0x2522],al    ; struct.cur_map
09f42  mov al,[bx+0x4ea0] ; ★ warp[param].dest
09f46  mov [0x2523],al    ; struct.dest
09f49  mov al,[bx+0x4ea1] ; warp[param].X
09f4d  mov [0x2524],al    ; struct.X
09f50  mov al,[bx+0x4ea2] ; warp[param].Y
09f54  mov [0x2525],al    ; struct.Y
09f57  lea si,[0x2521]    ; struct @0x2521 = {?,cur_map,dest,X,Y}
09f5b  call 0xd1f9        ; 執行轉場
```

**warp 表 0x4ea0**:3-byte `{dest, X, Y}`,113 非空。dest 小值=CTY 號(2,5,8,9),
dest≥100=overworld 回程(分派在 0xd1f9 深處,未解)。已抽 `dq3_warp.{c,h}`、remake examine
type-2 接 dest<100 載 CTY 於 (X,Y)。實證:CTY13洞→CTY2(40,1)、CTY24/26/66→CTY5、CTY82/87/88→CTY8。

## 5. 子型2 NPC = 0x3bb4 跳表(含 0xd1f9 warp NPC)

NPC 互動子型 `(byte3>>3)&7`;子型2 → handler 0x6355 → `byte4*2` 索引 `DS 0x3bb4` 跳表(docs/42)。
其中 **byte4 33-36 的 handler 呼叫 0xd1f9 warp**。

**全 CTY 掃描:0xd1f9 warp NPC(子型2 byte4 33-36)位置**:
- CTY21 sec0 (18,7) byte4=36
- CTY44 sec0 (13,27) byte4=33
- CTY44 sec3 (2,3) byte4=34

⇒ **阿里阿罕(CTY0)沒有 warp NPC**(CTY0 子型2 NPC byte4 都是 0-4 = 條件對話)。
故**阿里阿罕首個旅の扉不是子型2 warp NPC**。

> ★ **更正(假陽性)**:先前 sub2scan 以「0x120 byte 窗內出現 0xd1f9」判 byte4 33-36 = warp,**錯**。
> 逐段對齊後,0x5900-0x5990(byte4 32-34 handler)是**畫面效果 cutscene**(`lcall 0x1053,0x240` blit +
> `call 0xfd36` 逐行繪 + sti/cli 抖動迴圈 + `clear_flag(0x51)`),**不呼叫 0xd1f9**。0xd1f9 是窗尾溢出的
> 鄰近函式。docs/42 的「byte4 33-36 = warp」已一併更正。

## 5b. 真正的 scripted warp:call 0xd1f9 + 固定 struct(已解出目的地)

全檔 `call 0xd1f9`(E8)共 8 處(除 type-2 handler 0x9f5b 外);各以 `lea si,[固定 addr]` 帶
5-byte struct `{b0, cur_map, dest, X, Y}` warp。寫死座標 location 事件(同祠堂 0x9f7f 模式:
先 `[0x256c]==某CTY && [0x4f33/35]==某XY` 才觸發)。**解出目的地**:

| call@ | struct@ | cur_map(源)| dest(目的 CTY)| 落點 X,Y |
|---|---|---|---|---|
| 0x56ea | 0x4ee4 | 31 | 6 | 89,1 |
| 0x680b | 0x4eb5 | 24 | **1 雷貝** | 26,1 |
| 0x6d27 | 0x4ed5 | 26 | 5 | 75,1 |
| 0x7035 | 0x4ed0 | 35 | 5 | 75,1 |
| 0x71bf | 0x4eda | 12 | 5 | 67,1 |
| 0x7285 | 0x4ec0 | 24 | **1 雷貝** | 57,4 |
| 0x737e | 0x4ec5 | 24 | **1 雷貝** | 56,1 |
| 0x75a7 | 0x4edf | 39 | 6 | 121,1 |

**觀察**:dest 全是 1/5/6(CTY 號),洞穴/迷宮 → 城的固定連接,由 `cur_map` 閘。
**無 dest=2(羅馬利亞)、無 dest≥100(overworld)、無下層** → **阿里阿罕首旅の扉 + 首次下降不在這 8 個裡**。
(範例 0x680b:flag 0x2e set → CTY24 某處 warp 到雷貝(26,1)。)

> struct 位址(0x4eb5-0x4ee4)落在 warp 表 0x4ea0 區內(0x4ea0+0x15..),與 3-byte 表項資料重疊
> (緊湊打包);讀法以 5-byte struct 為準(0xd1f9 介面)。

## 5c. 呼叫鏈解出:warp handler = runner 0xabb2 / 子型2 同一張表

> ★ 又一個 **file/logical 混用更正**(rule 62):§5 我「disasm byte4=33 handler @0x5904」用 0x5904
> 當 **file** offset 是錯的 —— 跳表值是 **logical**,真 handler 在 **file = logical + 0x1370 = 0x6c74**。
> 所以那些 handler **確實是 warp**(我之前看錯位址才以為是 cutscene;§5 cutscene 更正的是另一段)。

**關鍵結構**:`0x3baa`(runner 0xabb2 event-id 跳表)與 `0x3bb4`(子型2 NPC 0x6355 byte4 跳表)是
**同一張 handler 表、偏移 5 項**(0x3bb4 = 0x3baa + 10 byte)。⇒ **子型2 byte4 = runner event_id − 5**。

8 個 scripted warp handler(§5b)的入口(logical)= `0x4312, 0x5482, 0x5904, 0x5cc1, 0x5e45,
0x5ee0, 0x6004, 0x622a`。其中 5 個落在 runner 表:

| handler(logical)| runner event id | 子型2 byte4 | warp 目的(§5b)|
|---|---|---|---|
| 0x5904 | 39(? 近 38)| 33 | dest 5/6(CTY44 byte4=33)|
| 0x5e45 | 58 | 53 | dest 5 |
| 0x5ee0 | 63 | 58 | dest 5 |
| 0x6004 | 66 | 61 | dest 1 雷貝 |
| 0x622a | 75 | 70 | dest 6 |

**結論**:scripted warp 由 ① **子型2 NPC**(byte4∈{33,53,58,61,70})或 ② **runner 0xabb2**(event_id 觸發)
派發。「event_id 從哪來」對 NPC 路徑 = **NPC byte4**(已解);對純 runner 路徑 = `[0x722]`(docs/31:
仍資料驅動、無靜態 setter,最難)。⇒ **NPC 觸發的 scripted warp 現在完全靜態可解**(byte4 → handler →
固定 struct 目的)。剩純 runner 觸發(無 NPC)的事件待動態。

## 7. overworld 寫死座標 → 旗標條件 CTY 進入(0x39f2 區,dispatcher 0x39cb)

掃 overworld 位置 `[0x4f2f]`(X)/`[0x4f31]`(Y)與常數比較,找到一組**寫死座標事件**(地表走到特定格
觸發),dispatcher **0x39cb**:

```
039cb  mov [0x256c], bp   ; ★ [0x256c] = 目的 CTY 號 = bp
039cf  mov [0x256a], 0    ; section 0
039d5  call 0x4378        ; 載入該 CTY(load_cty 路徑)
```

⇒ **地表特定座標 → 載入 bp 指定的 CTY**(進城/迷宮),且 **bp 由故事旗標選**(同位置依進度載不同 CTY)。
逐段讀出的事件(owX,owY,旗標,→CTY):

| owX | owY | 旗標(test_flag)| → 目的 CTY | 層 |
|---|---|---|---|---|
| 82 | 165 | 0x13 | CTY75 / CTY47 | 地表 |
| =[0x5053] | =[0x5055] | — | CTY36 | **下層** |
| 76 | 54 | 0x35 | (劇情:`[0x2593]=0x64` 物品事件 + rec 0x255/0x256,非進城)| — |
| 54 | 129 | 0x4d | CTY71 | 地表 |
| (flag 0x42)| | 0x42 | CTY60 | 地表 |
| (flag 0x48)| | 0x48 | CTY61 | 地表 |
| (else) | | — | **CTY83** | **下層** |

> ★ **更正**:先前說「含進下層 CTY 36/83」**有誤** —— 查 cty_loc:**CTY83=(210,64) map=0 是地表**
> (與 CTY58 同位置)、CTY36 map=0xff 是純迷宮。所以這機制是**同一 overworld 點依進度載不同城**
> 的**城鎮變體**(經典 DQ:城被毀/重建),**不是下降**(目的多為地表城)。完整 portal 表見 docs/45 §3.2。
> 另:`0x055ba` 有下層座標(127, 373=Y≥300)的寫死事件(下層 overworld 內的劇情點)。

> remake 落地:`find_cty_at` 目前只用 cty_loc 表(0x748)。要完整需另加「overworld 寫死座標 → 旗標條件
> CTY」這層(0x39f2 區的事件表 + 旗標閘)。屬待補。

## 8. 世界連接機制總表(收斂後)— 已找齊大部分

逐段 RE 後,把世界串起來的機制已大致找齊:

| 機制 | RE 段 | 觸發 | remake |
|---|---|---|---|
| `+0xc` 門/階梯/跨CTY/出城 | §1/2 | 走到轉場 tile | ✓ |
| `0xfe` 下降(Y+300)| §1/2 | 下層 CTY 的 0xfe 出口 | ✓ dual-layer |
| type-2 examine warp(0x4ea0)| §4 | Enter 事件 tile | ✓ |
| scripted warp(call 0xd1f9 + struct)| §5b/5c | 子型2 NPC byte4∈{33,53,58,61,70} / runner | ✗ 待接 |
| **overworld 座標 → 旗標條件 CTY** | §7 | 走到地表特定格 | ✗ 待接(含進下層 36/83)|

**地表↔下層連接仍未定位**(§7 是地表城變體,非下降;更正見上)。下層(map=1 的 15 城)要到下層 overworld
才進得去,而首次下降(地表→下層 overworld)仍未找到具體觸發。

### ★★ 首次下降 = runner event 86 / handler file 0x783d(已定位)

逐 `[0x5051]` setter 追:`[0x5051]=3` 在 **file 0x78ef**,在函式 **0x783d(logical 0x64cd)** 內。
該函式 = **runner 0xabb2 跳表 idx 85(event id 86)= 子型2 0x3bb4 idx 80(byte4 80)**。內容(劇情 cutscene):

```
0783d  di=0xc00; 印對話 rec 0x48        ; (D3TXT07 bank rec0x48 = 索瑪台詞 → 終盤)
0784e  [0x2518]|=0x42
07853  lea si,[0x4ef0]; call 0xd1f9      ; warp struct@0x4ef0 = {cur_map45 → CTY8 (124,1)}
07882  call 0x481b / 0x452d              ; 重載 section
0789d  清 [0x506a] 隊伍...
078d7  縮放畫面效果迴圈([0x251d])        ; ギアガの大穴 掉落動畫
078e3  [0x526c]=1
078ef  [0x5051]=3                        ; ★ 啟用下層(下次出場 Y+=300)
```

⇒ **首次下降是劇情 scripted 事件(runner event 86)**:顯示對話 → warp → 掉落動畫 → 設下層旗標。
**無 NPC 放置 byte4=80**(掃全 CTY 0 個)→ 純 runner 觸發(`[0x722]=86` 由劇情進度設,docs/31:
無靜態 setter)。rec 0x48 在終盤 bank = 索瑪 → 對應 **post-バラモス / アレフガルド 入口**(ギアガの大穴)。

> 使用者直覺(「有 NPC 說跳進坑」)方向對:此事件確實**先顯示對話再 warp**;只是觸發是 runner
> 而非可走 NPC。「跳進坑」的提示對白可能是另一條 hint NPC(子型0/1 對話)。

### remake 落地路徑

機制全解 ⇒ 可比照 #2 彩虹水滴(scripted_event 83)做:加 **scripted_event 86 = 下降**
(切 `layer=1` + 置玩家於下層 spawn),由劇情旗標或 debug 觸發 —— 把 dev 鍵 U 換成真實事件。
剩「event 86 的劇情觸發旗標」需 runner/動態(本環境做不到),但**事件本體已可在 remake 重現**。

### 待接進 remake(具體)
1. **§5b scripted warp**:抽「子型2 byte4 → warp struct 目的(dest,X,Y)」表 → 子型2 warp NPC 載 dest CTY。
2. **§7 overworld 座標事件**:抽 (owX,owY,旗標,destCTY) 表 → field 走到該格(旗標通過)→ load_cty。
   ← **這層含地表→下層(CTY36/83),接上即解鎖下層世界結構**。

## remake 落地對照

| 機制 | RE | remake |
|---|---|---|
| +0xc 門/階梯/跨CTY | §1 | ✓ 轉場 handler |
| 0xfe 下降(Y+300)| §1/2 | ✓ dual-layer field_under |
| 層由 Y 選圖檔 | §3 | ✓ 雙 field（raw 座標等價）|
| type-2 examine warp | §4 | ✓ dq3_warp + examine 接入 |
| 子型2 warp NPC | §5 | ✗ 待解 warp 目的 |
| 阿里阿罕旅の扉/下降 | §6 | ✗ dev 鍵 U 代 |
