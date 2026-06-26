# 事件 / 物件 / 互動子系統 RE 目錄(重新校正中)

> 本子系統(物件互動、對話派發、NPC、寶箱/門/旗標、合成事件如 #2)是 remake 邁向
> 「能破關」的關鍵,但也是**現有反組譯最弱、誤標最多**的一塊。本目錄以**重新逐函式
> 反組譯**為準,逐步校正並編目,**確認對了才在 remake implement**(避免照錯註解硬做)。

## 已抓出並修正的誤標(review 成果)

| 函式 / 項目 | 原(錯)註解 | 重新 RE 後 |
|---|---|---|
| `sub_fa39`(file 0xfa39)| map_obj_walk「走訪地圖物件」| **RNG**:`rol [0xb5a]×3` + `div n` → 回 `rng % n` |
| `sub_ba71`(file 0xba71)| 泛用 `event_run` 跑事件/對話 | **傷害類事件 handler**:扣實體 HP `[si+0x16]`、訊息 di=0x14d/0x165、`dec [0x24b4]` |
| packbg 戰鬥背景 | 「page 0x13」| 地形選頁 `page=[0xd73]+terrain*8`(docs/13)|

## 已確證的結構

### 前方物件 `[0x2511]` 的計算(在大函式 sub_99dc 內,file 0xad6b 起)

```
front_tile = (取自 sub @0xae46 的 dh)
if front_tile < 0x12:
    obj = front_tile;  if (front_tile >= 0x10) obj += 2      ; 0x10/0x11 特例
else:
    obj = [ (front_tile - 0x12) + 0x3930 ]                    ; ★ 次級物件表 @0x3930(tile→物件)
[0x2511] = obj
```

### 物件事件項 `[obj*3 + 0x37c4]`(3 byte)語意

- **byte0**(`[obj*3+0x37c4]`)= 事件碼(`0xff` = 無事件;confirm worker 0xb025 讀此)。
- **byte1**(`[obj*3+0x37c5]`)= 旗標;**bits 0x18**:`==8` → handler `sub @0xafdb`;否則 → 事件
  worker `sub @0xb025`(含 confirm_worker 0x9cd6)。
- byte2 = 參數(待定;與 `[0x2593]` event_param 有關)。

### 事件 worker(sub @0xb025 / confirm_worker 0x9cd6)

- 對 `[0x5077]` 個實體/物件迴圈(`[0x259c]` 為迭代器);讀前方物件事件碼,依碼派發到
  各 handler(其一 `sub_ba71` = 傷害事件;`sub_bae8` = 另一分支;`sub_ba41` = 「沒有任何事」)。
- 事件碼 → 具體 handler 的**完整派發表尚未逐碼校出**(多 handler)。

## 重要觀察:戰鬥與野外**共用碼交織**

`sub_99dc`(3320 byte)同時處理**戰鬥目標選擇**(`[0x24a9]` 選定目標、`[0xd75]&2` 戰鬥側別、
`sub @0xae46` 用 RNG 隨機選 set-bit 目標)**與野外物件互動**。`[0x37c4]` 事件表在兩種情境
都被查。所以不能把任一路徑單獨當「NPC 對話」——必須先把**戰鬥 / 野外 / 物件三條路徑拆乾淨**,
這是此子系統 RE 慢的主因。

## 待 RE(實作 NPC/事件前的前置)

1. **拆 sub_99dc 的三條路徑**(戰鬥目標 / 野外移動 / 物件互動),確認哪條是野外「調べる/話す」。
2. **次級物件表 `0x3930`** 的內容與語意(tile→物件索引)。
3. **事件碼 → handler 派發表**(逐碼:對話 / 寶箱 / 門 / 階梯 / 旗標 / 合成 #2)。
4. **NPC 實體資料**:位置 / sprite / 對話記錄 link 的來源(CTY 內?實體表 `[0x4f15]`?)。
5. **對話記錄選擇**:事件 → D3TXT record id 的對應(目前 remake demo 是循環,非派發)。

## ★ 決定性更正:`[0x37c4]` + confirm_worker 是**戰鬥/怪物事件**,非野外 NPC 對話

逐路徑追到底並**用文字記錄內容反證**:confirm_worker / sub_ba71 路徑顯示的記錄是

| rec | 內容(經 glyph→unicode 解)|
|---|---|
| 0x14d | `{V}受到{V}點的傷害。` |
| 0x160 | `{V}出現了!` |
| 0x165 | `{V}被打倒了!` |
| 0x161 | `怪物們被打倒了!` |

**全是戰鬥訊息**。⇒ `sub_ba71` = 戰鬥「受傷/擊倒」handler;`[obj*3+0x37c4]` 在此情境是
**怪物/戰鬥事件表**(配合 sub_99dc 的戰鬥目標路徑 `[0xd75]&2`/`[0x24a9]`),**不是**野外 NPC
對話表。**docs/12 把「野外 examine/話す」與「戰鬥事件」混為一談**——照它實作 NPC 對話會是錯的。

### 野外 NPC 對話的正確路徑(待 RE)

野外「話す」是**指令 idx6 `cmd_talk`(sub_5112 → sub_16dd)**:列出可對話對象(NPC 實體),
選定後顯示該 NPC 的對話。對話來源是 **NPC 實體結構的對話 id**(實體表 `[0x4f15]` / CTY 內),
**不是 `[0x37c4]`**。下一步應 RE:
1. `sub_16dd`(話す 列對象)的 NPC 實體欄位(stride 0x14;docs/12 提到 `si[0x12]`=屬性、
   `si[0x13]`=名字串 id `di=0x216+si[0x13]-1`)。
2. NPC 實體的**對話記錄 id** 欄位 + 在 CTY/實體表的位置/sprite。

> remake 教訓:`game` 模式目前用 `dq3x_event`(= 戰鬥事件表)當野外對話觸發閘 —— **語意錯**,
> 待找到正確的 NPC 實體對話路徑後改正。對話「引擎」(dq3_dialogue)本身對、可留用。

## ★★ 真正的野外/城鎮事件系統(逐層追到底,前述 [0x37c4] 結論作廢)

### 位址陷阱(先前一直追錯的根因)

`tools/re_disasm.py` 印的是 **file offset**;`exe_funcs.json` / `sub_XXXX` 名用 **logical(seg0)**。
`file = logical + 0x1370`。examine_chain(file 0x8fc0)`call 0x9cd6` 指 **file 0x9cd6**
(=logical 0x8966),**不是** sub_9cd6(logical 0x9cd6 = file 0xb046)。先前把兩者混用,
才一路追進**戰鬥**碼(file 0xb046 區)、誤判 [0x37c4]。**真正的野外 Enter worker 在 file 0x9cd6。**

### 野外 Enter 鏈(file 0x8fc0 examine_chain)

`call (file)0x9cd6` worker → 若 `[0x726]`(action result)非 0 續 `0x60c6` → `0x5b2d`(逐段推進)。

### worker(file 0x9cd6):依場景分派,事件編在 tile 高 byte

```
if [0x4f2d]==0 (地表): call (file)0x9cf6   ; 地表 examine
else (城鎮):           call (file)0x9f7f   ; 城鎮 examine(含寫死座標劇情)
```

地表 examine(file 0x9cf6):
```
讀玩家 tile 的 u16(seg_cty [0x2536] + [0xb26] + (Y*[0xb28]+X)*2)
attr = [tile_low*2 + 0x308e];  if (attr & 8)==0 → 無事件
subid = tile_high & 0x1f                         ; ★ 事件 = tile 高 byte 低 5 bit
ev_tbl = [ [0xb24]+8 ] + [0xb24]                  ; section 事件表(指標在 section+8)
si = ev_tbl + subid*4 + 1
type = [si];  param = [si+1](u16);  p2 = [si+3]   ; ★ 4-byte 事件項
依 type 分派:
```

| type | handler(file)| 動作 |
|---|---|---|
| **0** | 0x9ec1 | **調べる/寶箱**:param==0 → rec 0x108「在自己的腳邊調查‥」+ 0xf3「可惜是空的。」;否則給道具 |
| **1 / 3** | 0x9de9 | **給道具進隊伍背包**:掃隊伍 8 格找 `0xff` 空格存入 param(滿了印 rec 0x13a);type3 設 `[0x726]=1` 續鏈。⚠ **非**「找鑰匙的條件檢查」(舊註誤),全段已反組譯確認,見 docs/35 |
| **2** | 0x9f32 | **傳送 / 門 / 階梯**:目的地 = `[param*3 + 0x4ea0]`(地圖號 + 座標)→ 切場景 |
| else | 0x9f71 | 給道具 `[0x24fc]` + 訊息 rec 0x1e4 |

城鎮 examine(file 0x9f7f):另有**寫死座標的劇情事件**(如 `[0x4f2f]==0x6d && [0x4f31]==0x3d`
→ `[0x2593]=0x4c` + `seq_msg`(file 0x7c0c)逐段對話)。

### 對 remake 的意義(全部要的都在這)

- **NPC/物件對話、寶箱、門、條件**:走 tile 高 byte 事件 subid → section 事件表 → type 分派。
- **#2 彩虹水滴**:是 type 0/default 的**給道具**事件,param=道具 id;修正 = 產出 0x75(docs/18/22)。
- **世界串接(城鎮出入口)**:type 2 的 **warp 表 `0x4ea0`**(3 byte:地圖+座標)= 門/階梯目的地。
- **CTY section header**:`+8` = 事件表指標;另 `[di+1]` = 安全旗標 `[0xd77]`(docs/13)。
  → docs/04/11 未定名的 section header 欄位,逐一補上。

### #2 彩虹水滴合成(file 0x77ce,已確認本體)

祠堂劇情事件本體反組譯確認:

```
0x77ce: call 0x6380
0x77d1: mov word [0x2593], 0x72   ; 找太陽之石
0x77d7: call 0x7c0c               ; party_has_item([0x2593]) → si 指向該格(非 seq_msg)
0x77da: mov word [si], 0xff       ; 消耗太陽之石(該格清空)
0x77de: mov word [0x2593], 0x73   ; 找雲雨之杖
0x77e4: call 0x7c0c               ; party_has_item → si 指向雲雨之杖格
0x77e7: mov word [si], 0x6b       ; ★ 寫合成品 0x6b(BUG;應為 0x75 彩虹水滴)
0x77eb: mov bx, 0x139             ; 旗標 id
0x77ee: call 0x8264               ; set_flag(0x139)「已合成」
0x77f1: ret
```

> 註(file/logical 陷阱):此處 `call 0x7c0c` 是 **file 0x7c0c** =
> **`party_has_item([0x2593]) → si`**(掃隊伍 8 格道具,docs/18 已正解)。先前 docs/31
> 把它標成 seq_msg 是錯的(已更正)。⚠ 與 `re/commands` 的 **logical** `sub_7c0c`(= file
> 0x8f9c,另一個函式)不同位址,勿混。

remake:`dq3_synth_shrine_examine`(`dq3_inventory.c`)鏡射條件與副作用 —
未設旗標 0x139 才觸發、持兩材料 → 合成(修正產 0x75)+ 設旗標 0x139。

#### 呼叫鏈(已解出:scripted-event 派發)

`0x77ce` 既非同段 near-call(全檔 E8 掃 0 命中)、亦無 far-call(全檔 9A 掃 0)、
指標表也查無(word `0x77ce` / far 指標 0 命中)。原因:**0x77ce 是 handler `file 0x776c`
的成功尾段(fall-through)**,而 0x776c 由 **scripted-event 跳表** 間接派發:

```
runner  file 0xabb2:
  mov bx, [0x722]        ; event id(資料驅動)
  dec bx ; shl bx,1
  call word ptr [bx+0x3baa]   ; → handler 跳表(DGROUP DS 0x3baa)
跳表 DS 0x3baa:  idx0=phys0x988d … idx82=phys0x63fc(=file 0x776c 合成)…
∴ 彩虹水滴合成 = scripted-event id 83(0x53)  [index 82 = id-1]
```

- **event id 為資料驅動**:`[0x722]` 由 `[0x631]` 等載入,且 `[0x631]↔[0x722]` 互為
  save/restore(事件巢狀,file 0x841d/0x846b…),非源頭;全檔無 `mov [0x722],0x53`。

#### 派發鏈邊界(靜態追蹤止於此 — 已窮舉)

runner `0xabb2` 與跳表 `DS 0x3baa` 是 scripted-event 系統的核心,但:

- runner 0xabb2 **零靜態呼叫者**:near `E8`、far `9A`、near/far `jmp`、far 指標 data
  **全 0 命中**。掃描器以滿地的 `lcall 0x10b6,0x166`(56 個 far call)反查驗證無誤。
- 跳表 `DS 0x3baa` **只在 0xabe1(runner 內)被 `call [bx+0x3baa]` 用一次**;表 base 無
  其他 immediate / lea / disp 參照。
- 整個 handler 群(`0x72ab…0x776c…`)+ 表 + runner **無任何可靜態解析的進入點**。

→ (當時)結論:runner 由執行期計算的指標 / 事件腳本 VM 進入,純靜態追不到祠堂座標。

> **★ 後續更正(2026-06,純靜態解決,未用 DOSBox)**:上述「純靜態追不到 / 需動態」結論**已被推翻**。
> - 祠堂座標:CTY93 section0 NPC 清單唯一一隻祭司 @ **(8,8)**,三證據(CTY 資料 + 劇本「神聖的祠堂」+
>   地理 docs/32)靜態定位(見 `dq3_remake/include/dq3_inventory.h` DQ3_SHRINE_NPC_X/Y)。
> - runner `[0x722]` 進入點:靜態重新分析找到 **57 個 explicit setter**(`mov word [0x722],N`)+ 座標→region
>   hit-test,完整解出派發(`mov bx,[0x722]; dec bx; shl bx,1; call [bx+0x3baa]`),sub2 跳表 = runner 表
>   偏移 5 格。整套劇情觸發機器全靜態解完 —— 見 `docs/re-log-722-state-machine.md`、`tools/decode_sub2_struct.py`。
>
> 下方「動態分析」段保留為當時的死路推理紀錄(教訓:撞牆別退回 DOSBox,先換 query 把 setter 掃出來)。

**(當時擬的)下一步(動態分析)**:DOSBox debugger 在 runner phys `0x9842`(或 handler
`file 0x776c`)下中斷點 → 觸發祠堂 → 讀當下地圖 id `[0x256c]` + 座標 `[0x4f2f]/[0x4f31]`
+ 呼叫堆疊。**(實際未走此路:靜態已解,見上更正。)**

### 訊息 control-code 格式(字串印表器 `0x111b:0x264` = file 0x12784)

訊息字串 = **16-bit word 陣列**;非負 = 字元字模碼,負值 = control code(分派表 0x127ba+):

| word | 語意 | handler |
|---|---|---|
| `0xffff` (-1) | 字串結束 | — |
| `0xfffe`/`0xfffd` (-2/-3) | 換行 / 進下一行 | 0x1280a |
| `0xfffc` (-4) | 捲動 + 等待按鍵 | 0x12a33(poll [5]) |
| `0xfffb` (-5) | 插入變數子字串(依 [0x259c]) +1字參數 | 0x129c1 |
| `0xfffa` (-6) | lcall 0x111b:0x6f0(插入) +1字 | — |
| `0xfff9` (-7) | 插入索引字串(依 [0x2591]) +1字 | 0x12a97 |
| `0xfff6` (-10) | 插入變數子字串 variant 7 +1字 | 0x129c1 |
| `0xfff5` (-11) | 插入變數子字串 variant 0 +1字 | 0x129c1 |
| `0xffed` (-19) | 插入索引字串(依 [0x249d]/[0x249f]) +1字 | 0x12ad6 |

全為**文字格式 / 變數插入**,**無「run scripted-event」碼** → 確認祠堂事件不經訊息系統。

remake:`dq3_scripted_event_run(id, …)`(`dq3_inventory.c`)鏡射 runner 的 id→handler
跳表(目前僅 id 83 → 合成);main.c 城鎮調べる發 event 83(暫代真實地圖資料來源)。

### 待 RE(收尾)

- **發 event 83 的 CTY 地圖事件 / 觸發點**(祠堂城鎮 + 座標)→ **改用動態分析**(上述
  DOSBox 中斷點),靜態已窮舉。
- section 事件表的**完整 entry 格式**(type0 給道具的 item id 欄、type2 warp 表 0x4ea0 逐欄)。
- 寫死座標劇情事件清單(0x9f7f 內的多個座標分支)。
- 把上述抽成內建資料(像 dq3_exedata)供 remake,不依賴 DQ3.EXE。

## remake 現況(此子系統)

- 對話**引擎**可用(`dq3_dialogue`:真繁中劇本、分頁、控制碼);`game` 模式 Enter 以
  `dq3x_event`(內建事件表 byte0)當觸發閘 → 起對話(demo 循環記錄)。
- 但「哪個物件 → 哪句對話」的**事件派發**待上述 RE 校正後才能 faithful implement。
