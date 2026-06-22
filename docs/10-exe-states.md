# DQ3.EXE 主迴圈狀態機:鍵碼跳表與指令 handler(RE → C)

主迴圈 `sub_93e3`(seg0:0x93e3,file 0xa753)讀鍵盤 scancode 後,方向鍵直接改玩家座標,
其餘鍵走一張 **scancode → near 函式指標** 的跳表派發到各遊戲狀態 / 動作。本文件記錄這張
跳表的實際內容、各 handler 的位址與語意,以及對應的 C 反編譯(`re/states.c` / `re/states.h`)。

工具:`tools/re_states_dis.py`(docker capstone 內跑,讀資料段表 + 反組譯各 handler)。

```
docker run --rm -v "$PWD":/work -w /work ghcr.io/astral-sh/uv:python3.12-bookworm-slim \
  bash -c "uv venv -q /tmp/venv && . /tmp/venv/bin/activate && uv pip install -q capstone \
           && python tools/re_states_dis.py"            # 印跳表 + 6 個 handler
           # python tools/re_states_dis.py extra        # 印選單子指令與輔助函式
           # python tools/re_states_dis.py 0xNNNN C     # 任意 file offset 反組譯 C 條
```

## 跳表派發機制(更正 docs/08 的「14-entry far 指標表」描述)

主迴圈分派碼(file 0xa801):

```asm
mov cx, 0xe                 ; 固定掃 14 格(迴圈上界,非真實表長)
mov bx, 0
loop_find:
  cmp byte ptr [bx+0x19], ah ; scancode 表 @ DS:0x19(位元組陣列)
  je  found
  inc bx
  loop loop_find             ; 找不到 → 不做事
  jmp draw
found:
  shl bx, 1                  ; index*2
  cmp word ptr [bx+0x25], 0  ; 指標表 @ DS:0x25(每格 1 word)
  je  draw                   ; 0 = 已註冊但無 handler(吃掉按鍵不做事)
  call word ptr [bx+0x25]    ; ★ near call(不是 far),目標在 seg0 內
```

兩處要點修正先前文件:

1. **指標表是 near(`call word ptr`,2 bytes/entry,seg0 內偏移),不是 far(seg:off)**。
   far call 會是 `call dword ptr`(4 bytes)。所有 handler 都在主程式碼段 seg0。
2. **真實表長 = 11 項 + `0xff` 終止碼**,不是 14。`cx=0xe`(14)只是迴圈上界;
   scancode 表第 12 格(idx 11)是 `0xff`(永不等於任何真實 scancode AH),其後是檔名字串池。
   未命中時這個 0xff 與後續垃圾位元組讓掃描安全落空。

### scancode 表(DS:0x19)與 handler 指標表(DS:0x25)實際內容

從主資料段 DS=0x14dd(file base 0x16140)讀出:

| idx | scancode | 鍵(DOS Set-1) | handler near off | file off | 語意 |
|-----|----------|----------------|------------------|----------|------|
| 0 | 0x10 | Q  | 0x9842 | 0xabb2 | **開指令選單**(field command window) |
| 1 | 0x1f | S  | 0x9842 | 0xabb2 | 同上(指令選單,Q/S/F1 三鍵共用) |
| 2 | 0x39 | Space | 0x7c83 | 0x8ff3 | **快速「調べる/前方互動」**(滑鼠+視窗) |
| 3 | 0x1c | Enter | 0x7c43 | 0x8fb3 | **確認 / 對話推進 / 前方動作**(分段執行) |
| 4 | 0x3b | F1 | 0x9842 | 0xabb2 | 同 idx0(指令選單) |
| 5 | 0x3e | F4 | 0x0000 | — | 已註冊、**無 handler**(吃掉按鍵) |
| 6 | 0x3f | F5 | 0x13a9 | 0x2719 | **開狀態 / 訊息資訊視窗**(顯示模式,[0x1f0]=0) |
| 7 | 0x40 | F6 | 0x14da | 0x284a | **存檔 / 載入槽選擇**(int21 開檔,[0x1f0]=1) |
| 8 | 0x41 | F7 | 0x0000 | — | 已註冊、無 handler |
| 9 | 0x42 | F8 | 0x0000 | — | 已註冊、無 handler |
| 10 | 0x43 | F9 | 0xe915 | 0xfc85 | handler 僅 `ret`(no-op stub) |
| 11 | 0xff | — | — | — | 終止碼(掃描止於此) |

> scancode 對鍵的對照採 IBM PC/XT Set-1。Q(0x10)/ S(0x1f)/ F1(0x3b)三鍵指向同一個指令選單
> handler,是當年「多鍵綁同一動作」的常見設計(鍵盤 + 不同提示)。

方向鍵與選單鍵不走此表,直接在主迴圈內處理:↓0x50 / ←0x4b / ↑0x48 / →0x4d 改座標+朝向;
`0x3e`(此處指 AH=0x3e 的另一路徑,主迴圈 switch 內)`lcall 11dd:0c` 為選單/特殊入口
(見 mainloop.c,屬主迴圈本體,非本表)。

## 各 handler 語意(反組譯佐證)

### h_9842 — 指令選單(Q / S / F1) file 0xabb2

DQ 風格「指令視窗」開關 + 二級派發。流程:

```
lcall 10b6:166                 ; 滑鼠 off / 畫面凍結
[0x72a] = 1                    ; 進入 modal(選單忙碌旗標)
lcall 10b6:19c                 ; 滑鼠 on
window @ [0x424c]:  call 0x10900(建視窗/游標)→ cx=5; call 0x10ae9(選單導航,5 列)
if [0x726] != 1:               ; [0x726] = 結果碼(1=取消/關閉)
    bx = [0x722] - 1           ; [0x722] = 選中的指令序號(1-based)
    call word ptr [bx*2 + 0x3baa]   ; ★ 二級跳表:指令動作派發
close window: call 0x10974
[0x72a] = 0                    ; 離開 modal
```

二級跳表 **DS:0x3baa**(near 指標,`[0x722]-1` 索引)= 12 個 field 指令動作:

| sub-idx | near off | file off | 推定指令(field command) |
|---------|----------|----------|--------------------------|
| 0 | 0x988d | 0xabfd | 旗標/開關類動作(test [0xf]&0x18 → 設 [0x286d],file I/O) |
| 1 | 0x98b9 | 0xac29 | 同群組(清 [0x286d],file I/O) |
| 2 | 0x98cd | 0xac3d | test [0xf]&1 → [0x2784]=0,lcall 10b6:71/3a0 |
| 3 | 0x98e6 | 0xac56 | test [0xf]&1 → [0x2784]=1,lcall 10b6:385 |
| 4 | 0x9900 | 0xac70 | 重入主流程(lcall 1104:123 + jmp 0xa62c 重啟) |
| 5 | 0x504e | 0x63be | **道具 / 裝備清單**(逐項以文字繪製器列出,空欄判斷) |
| 6 | 0x5112 | 0x6482 | call 0x16dd(子動作) |
| 7 | 0x5116 | 0x6486 | 旗標查詢 + 文字輸出 |
| 8 | 0x512f | 0x649f | call 0x1a4c(子動作) |
| 9 | 0x5133 | 0x64a3 | call 0x7c50(子動作) |
| 10 | 0x5137 | 0x64a7 | 條件動作(查旗標 0x29/0x4d → 文字 + call 0x2719 開資訊視窗) |
| 11 | 0x51e9 | 0x6559 | 條件動作(查旗標 → 文字 + call 0x2719) |

> sub-idx 5(0x504e)反組譯可見「逐欄位讀 [0x507f] 結構、空欄印不同字串 ID(0xbca/0xbcb…)、
> 以文字繪製器 `lcall 111b:264` 輸出」,是典型的**道具 / 裝備 / 隊伍清單**列表畫面。
> 二級跳表多數 entry 以 `lcall 111b:264`(文字繪製器,見 docs/06)輸出訊息,以 `call 0x8279`
> (旗標查詢,回 al=0/1)/ `call 0x8264`(旗標設定)讀寫遊戲旗標。

### h_7c83 — 快速互動 / 調べる(Space) file 0x8ff3

```
if [0x4f3b] >= 2: call 0x434f; ret        ; 特定模式(如戰鬥/過場)下走別的分支
lcall 10b6:166                            ; 凍結畫面
int 33h ax=3 → 存滑鼠座標 [0x2852]/[0x2854]
[0x72a]=1; lcall 10b6:19c                  ; modal on
[0x2875]=1; lcall 1053:240                 ; 檔案/資源操作
建立資訊框 [0x3e9c..]:座標常數(0x13/0xee)+ [0x5077](隊伍人數)算高度
call 0x8279(bx=0x27 旗標)分支 → call 0x9592(畫隊伍/狀態列)
lcall 1104:123(等待鍵)→ 關框 call 0x10974
int 33h ax=4 還原滑鼠座標 → [0x72a]=0
```

語意:**前方/原地快速互動**(以滑鼠座標 + 隊伍狀態框呈現),是 Space 的「看一下 / 調べる」捷徑。
與指令選單的「調べる」動作相關但走獨立捷徑。

### h_7c43 — 確認 / 對話推進(Enter) file 0x8fb3

```
if [0x4f3b] >= 2: call 0x434f; ret         ; 同 Space,特定模式改走別處
lcall 10b6:166
[0x1f7]=1                                   ; 進入「動作執行中」旗標
call 0x9cd6                                 ; 主要動作 worker(讀 [0x2511] 前方 tile/物件,
                                            ;   查 [bx+0x37c4] 物件表,推進對話/事件)
if [0x726] != 0:  call 0x60c6               ; 分段:第一段成功才做下一段
   if [0x726] != 0: call 0x5b2d             ; 再下一段(多段對話 / 事件鏈)
lcall 1104:123                              ; 等待鍵
[0x1f7]=0                                    ; 離開
```

語意:**Enter = 確認 / 與前方對象互動 / 推進多段對話事件**。`[0x726]` 為每段的結果碼,
非 0 才接續下一段,構成「按 Enter 逐段推進」的對話/事件流。

### h_13a9 — 資訊 / 狀態視窗(F5) file 0x2719

```
[0x72a]=1
if ax != 0: 開子視窗 [0x3e6e]; [0x259b]=0(文字行計數歸零)
call 0xec59                                 ; 讀資料(int21 AH=42/3F,從 [0x2534] 段載入區塊)
di=0xfd; lcall 111b:264                      ; 文字繪製器:畫訊息 ID 0xfd
call 0x9ac
if [0x722]==1: di=0xfa; lcall 111b:264; call 0x2793   ; 額外資訊頁
di=0xfc; lcall 111b:264                      ; 收尾訊息
lcall 1104:db                                ; 等待
關框 [0x3e6e]; [0x72a]=0
```

`call 0x2793` 內含 `rep movsb` 把 [0x507f+3..] 結構(18+3 bytes)複製到 [0x128+(idx*0x14)],
並 call 0x27f4 把 camera 座標 [0x4f2f..0x4f35] 備份到 [0x4f48..0x4f4e](開窗前存座標、關窗還原)。
與 F6 共用開窗常式 0x29d0(建 [0x4f21..0x4f27] 視窗矩形)。`[0x1f0]=0` 標示「顯示模式」。

語意:**F5 = 開狀態 / 隊伍資訊視窗(唯讀顯示)**。

### h_14da — 存檔 / 載入槽(F6) file 0x284a

```
lcall 10b6:166; [0x72a]=1; lcall 10b6:19c
[0x1f0]=1                                    ; 「編輯/選槽模式」(對比 F5 的 0)
call 0x29d0                                  ; 與 F5 共用:建視窗矩形
if [0x726]==2: ret                           ; 取消
畫框 [0x3fba](文字 ID 0x3fba)
if [0x726]!=1:                               ; 確認選槽
    掃 [0x128] 槽位陣列找空槽 → [0x722]=槽號(+1)
    if [bx+0x13a]!=0: call 0x28ce(寫檔)+ call 0x2902(填存檔緩衝)
[0x72a]=0
```

存檔核心 `0x28ce`:

```asm
ax = [0x722]; dec al; add al, 0x30      ; 槽號 → ASCII 數字 '0'..'9'
mov [0x58], al                          ; 改寫檔名第 (0x58-0x3d=0x1b) 位元組 = 存檔檔名模板的數字位
... compute length [0x57a5]-[0x4f29] ...
lea dx,[0x52]; ah=0x3d; al=0; int 21h    ; DOS 開檔(DS:DX=存檔名)
ah=0x3f; int 21h                         ; 讀(載入)
ah=0x3e; int 21h                         ; 關檔
```

`0x2902`:把 [bx+0x4f15] 隊伍 / 進度資料搬到 [bx+0x613] 緩衝(存檔內容組裝),
以 [0x5077](隊伍人數)為迴圈次數。

語意:**F6 = 存檔 / 載入槽選擇**;以 [0x722]-1 的槽號改寫存檔檔名模板的數字位(同 docs/06
「檔名模板就地改號」技巧:`cty00`+index、`dq3?`+index),再 DOS 開/讀/關。

### h_e915 — F9 no-op stub file 0xfc85

```asm
00fc85: c3   ret
```

handler 入口即 `ret`,**不做任何事**。F9 在此版被註冊到表中但停用(可能為 debug / 未實作功能保留)。

### 無 handler 的已註冊鍵:F4(0x3e)/ F7(0x41)/ F8(0x42)

指標表對應格為 0,主迴圈 `cmp word ptr [bx+0x25],0; je` 直接跳過 → **按鍵被吃掉但不做事**。
與「未在表中的鍵」差別:這些鍵命中 scancode 表(不會落到掃描末端),但無動作。推測為
保留鍵位 / 已停用功能。

## 主迴圈會踩到的狀態變數(主資料段 DS=0x14dd)補充

| DS off | 含義(本文件新增 / 佐證) |
|--------|--------------------------|
| `0x19` | scancode 跳表(11 項 + 0xff 終止,位元組陣列) |
| `0x25` | handler near 指標表(每格 1 word,0=無 handler) |
| `0x3baa` | 指令選單二級派發跳表(12 項 near 指標,[0x722]-1 索引) |
| `0x722` | 選單 / 槽位選中序號(1-based) |
| `0x726` | 動作結果碼 / 取消碼(1 或 2 = 取消;Enter 鏈逐段判 ≠0 續行) |
| `0x72a` | modal / 選單忙碌旗標(進選單=1,離開=0;繪製管線會檢查) |
| `0x1f0` | 資訊視窗模式:F5=0(唯讀顯示)/ F6=1(選槽 / 編輯) |
| `0x1f7` | Enter 動作執行中旗標 |
| `0x5077` | 隊伍人數(多處作迴圈次數 / 視窗高度計算) |
| `0x2511` | 玩家前方 tile / 物件索引(Enter worker 查 [bx+0x37c4] 物件表) |
| `0x4f3b` | 模式旗標(≥2 時 Space/Enter 改走 0x434f 分支) |

## 反編譯產出

- [`re/states.h`](../re/states.h) — 跳表型別、handler / 二級派發宣告、本文件用到的狀態變數位址。
- [`re/states.c`](../re/states.c) — 6 個 handler + 指令選單二級派發的可讀 C(real-mode large-model
  風格;far runtime 以 exe.h wrapper 表達;核心遊戲函式 forward-declare,不在此反編譯)。

C 風格同專案慣例:求結構正確、可審閱,非位元精準;逐函式對 DOSBox 驗證為後續步驟。

## 殘留 / 待釐清

- 二級跳表 sub-idx 6/8/9(call 0x16dd / 0x1a4c / 0x7c50)的具體指令名稱(話す/呪文/装備?)
  需追入各子函式 + 對照 D3TXT 訊息 ID 才能精確命名;目前以「指令動作」標記。
- `[0x726]` 結果碼的完整值域(目前見 0/1/2)與各值語意需更多 caller 佐證。
- F9 stub 是否在其他版本啟用、F4/F7/F8 原規劃功能,需比對其他 build。
