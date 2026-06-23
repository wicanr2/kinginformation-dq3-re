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

## remake 現況(此子系統)

- 對話**引擎**可用(`dq3_dialogue`:真繁中劇本、分頁、控制碼);`game` 模式 Enter 以
  `dq3x_event`(內建事件表 byte0)當觸發閘 → 起對話(demo 循環記錄)。
- 但「哪個物件 → 哪句對話」的**事件派發**待上述 RE 校正後才能 faithful implement。
