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
| **1 / 3** | 0x9de9 | **條件檢查**:掃隊伍 `[0x4f15]+0x3a` 8 格裝備找某物(如鑰匙);type3 設 `[0x726]=1` 續鏈 |
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

### 待 RE(收尾)

- section 事件表的**完整 entry 格式**(type0 給道具的 item id 欄、type2 warp 表 0x4ea0 逐欄)。
- 寫死座標劇情事件清單(0x9f7f 內的多個座標分支)。
- 把上述抽成內建資料(像 dq3_exedata)供 remake,不依賴 DQ3.EXE。

## remake 現況(此子系統)

- 對話**引擎**可用(`dq3_dialogue`:真繁中劇本、分頁、控制碼);`game` 模式 Enter 以
  `dq3x_event`(內建事件表 byte0)當觸發閘 → 起對話(demo 循環記錄)。
- 但「哪個物件 → 哪句對話」的**事件派發**待上述 RE 校正後才能 faithful implement。
