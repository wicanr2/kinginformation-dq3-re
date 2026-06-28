# 城鎮日夜 NPC 機制 RE + 晚期 gameplay 還原(2026-06-28)

> 這批工作的核心是一個 RE 突破:**城鎮 NPC 的白天/黑夜差異不是靠可見性旗標,而是每個 section
> 內建兩份完整 NPC 清單,由日夜狀態選載哪一份。** 本文記錄找到它的過程、機制、實作與驗證,並附帶
> 同批的傳送持久化等還原。系統層的晝夜(palette / 步數驅動 / 手動切換)見
> [`docs/data/day-night-system.md`](data/day-night-system.md);本文補上其中「NPC」那一格的底層機制。

## 1. 問題

使用者看 demo 影片後指出:原版城鎮**白天(到黃昏都算)NPC 走動,黑夜時有的睡覺、有的隱藏,
有的白天隱藏的夜間才現**。remake 當時日夜**只改 palette**,NPC 行為日夜完全相同。問:有沒有還原?
答:沒有,需要 RE。

## 2. RE 過程(含兩個錯誤轉折)

### 2.1 第一個錯:位址基準

`docs/35` 把 NPC 載入標在 `0x457c`、旗標表標在 `[0x4f70]`。我先用本專案慣例
`file = logical + 0x1370` 去反組譯 → 落在一段查 `[0x4f33]` 的圖形/緩衝碼,**不是** NPC 載入。
試 window 1(`+0x1370+0x10000`)→ 落在音訊驅動。兩個都不對。

關鍵更正:**`docs/35` 的位址其實是 file offset**(不是 logical)。證據 ——
`[0x4f70]`(operand bytes `70 4f`)第一個引用就落在 **file 0x457e**,恰好 ≈ doc 的 `0x457c`。
換句話說 doc 標的就是檔內偏移。校正基準後一切對上。

### 2.2 找到可見性檢查(但這只是劇情層)

file 0x4560 是 NPC 載入迴圈,核心:

```asm
mov  bl, es:[si+5]          ; NPC 槽 byte5 = 旗標 id(整個 byte)
mov  cl, 7
and  cl, bl                 ; 低 3 bit → bit 位置
mov  ah, 0x80
ror  ah, cl                 ; 0x80 ror (id&7) = 該 bit 的 mask
shr  bx, 1 / 1 / 1          ; id >> 3 = byte index
test byte [bx+0x4f70], ah   ; 測 [0x4f70] bit array
je   skip_npc               ; 旗標清 → 不載入這隻 NPC
... 把 7-byte 記錄複製進 NPC 槽(載入)...
```

即 NPC 的**個別可見性** = 槽 byte5 當旗標 id,查 `[0x4f70]` story-flag bit array,清則不載入。
這是**劇情層**(某 NPC 在某進度才出現),一度誤以為日夜也走這條 —— 但 8-bit id 與日夜無直接關係,
而且找不到「日夜推進去 set/clear 哪個旗標」。卡住。

### 2.3 真正的機制:section 內兩份 NPC 清單

往迴圈**上文**看(file 0x452d),謎底揭曉:

```asm
cmp  byte [0x526c], 1       ; [0x526c] = 日夜選擇器
jne  +5                     ;   ≠1 → di=2
mov  di, 0                  ;   ==1 → di=0
...
mov  bx, [0xb24]            ; bx = 當前 section base
add  di, bx                 ; di = base + (0 或 2)
mov  ax, [0x2536]; mov es,ax
mov  si, es:[di]            ; 讀 base+0 或 base+2 的 word = NPC 清單指標
cmp  si, -1; je end         ; 0xffff = 無清單
add  si, bx                 ; 清單位置 = base + 指標
mov  cl, es:[si]            ; 進 2.2 的載入迴圈({count, 7-byte 記錄×count})
```

**每個 section 開頭兩個 word 是兩份 NPC 清單指標**(相對 section base):

| 偏移 | `[0x526c]` 值 | 內容 |
|---|---|---|
| `base+0` (word0) | `==1` → di=0 | 一份清單 |
| `base+2` (word2) | `≠1` → di=2 | 另一份清單 |

`[0x526c]` 就是日夜狀態。可見性旗標(2.2)在**每份清單內部各自再過濾**(劇情)。兩層獨立。

### 2.4 哪份是白天、哪份是黑夜:讓資料說話

`[0x526c]` 的語意(0/1 哪個是日)直接看真實 `CTY*.DAT` 資料最快。dump 各城 section0 的兩份清單 count:

| 城鎮 sec0 | word0 NPC 數 | word2 NPC 數 |
|---|---|---|
| CTY00 | 24 | 6 |
| CTY01 | 11 | 4 |
| CTY02 | 31 | 10 |
| CTY03 | 12 | 4 |

word0 一律比 word2 多 3–5 倍。城鎮白天人多、夜間多數人回屋睡覺 → **word0 = 白天表、word2 = 黑夜表**。
對回 2.3:`[0x526c]==1`(→word0)= 白天。再核對日夜寫入點:

- file 0x101b8(每步時刻推進):`inc [0x251d]; cmp [0x251d],0x78; → mov [0x526c],0` —— 時刻計數器
  `[0x251d]` 每步加,到 0x78 設夜(0)。
- file 0x78e3(黃昏漸暗動畫):逐步 `sub [0x251d],0x14` 重繪,結束 `mov [0x526c],1`(回白天)。
- file 0xe1ab(拉那魯達 toggle):`[0x526c]` 在 0↔1 間切。

時刻只在 overworld 前進(`[0x4f2d]` gate 跳過城內推進)→ **城內相位固定**。

## 3. 機制總結

```
[0x251d] 時刻計數器(0..0xf0 循環,每 overworld 步 +1,城內不前進)
   │  衍生
   ▼
[0x526c] 日(=1→word0)/夜(=0→word2)
   │  進城 / 切換時選清單
   ▼
section base+0 (白天 NPC 清單)  ┐
section base+2 (黑夜 NPC 清單)  ┘ → {count, 7-byte 記錄×count}
   │  逐筆再過濾
   ▼
槽 byte5 旗標 id → [0x4f70] story-flag bit → 清則略過(個別劇情可見性)
```

「睡覺」「隱藏」「夜間才現」三種觀察全由「黑夜清單 = 另一組記錄」自然涵蓋:
夜間消失的 NPC 不在黑夜清單、夜間才現的在黑夜清單、睡在床上的 NPC 就是黑夜清單裡
**byte2(sprite id)指向睡姿圖**的記錄 —— 全是資料,不需額外狀態欄位。

## 4. remake 實作

### 4.1 依相位選清單(`dq3_scene.c` `dq3_scene_load_npcs`)

原本永遠先用 word0(白天表),word2 只當 word0==0xffff 的 fallback —— 等於**一直顯示白天 NPC**。
改成依當前相位選主表,主表無效(0xffff/越界)才退另一份(部分 section 只有一份,日夜共用):

```c
night = (dq3_scene_get_daynight() == 2);     /* 2=黑夜;白天/黃昏/黎明皆當白天 */
for (k = 0; k < 2; k++) {
    size_t off = ((night ^ k) ? 2u : 0u);    /* k=0 主表;k=1 退另一份 */
    uint16_t w = cty[so+off] | (cty[so+off+1] << 8);
    if (w == 0xffff || so + w >= cty_len) continue;
    np = w; break;
}
```

時刻只在 overworld 前進、城內相位固定 → **進城時選一次即還原日夜 NPC**,與原版一致,不需城內輪詢重載。

### 4.2 城內即時切日夜(`main.c` `reload_town_daynight()`)

拉那魯達(咒)/黑暗之燈(道具)可在城內把白天切成黑夜。切換後若不重載,夜間 NPC 不會現出(看起來像壞掉)。
加一個 local 重載:在城內時用當前 cty/section 重跑 `dq3_town_load`,**保留座標**;follower 由每幀
`apply_followers`(以 scene 指標判斷)在下一幀自動重設。

### 4.3 驗證

town 測試模式加 `DQ3_DN` env 強制相位 + `[DN]` 載入數診斷,docker headless 跑:

```
CTY00  白天 → NPC 24 隻    黑夜 → NPC 6 隻
CTY01  白天 → NPC 11 隻    黑夜 → NPC 4 隻
CTY02  白天 → NPC 31 隻    黑夜 → NPC 10 隻
CTY03  白天 → NPC 12 隻    黑夜 → NPC 4 隻
```

runtime 與 §2.4 資料表完全一致。unit tests + 全 headless dump 綠。

## 5. 同批還原:魯拉/蓋美拉傳送持久化(save v8)

- 魯拉(咒 rec172)/蓋美拉之翼(道具 0x43)選單可選「回地表 / 回去過的城鎮」。
- 「去過的城鎮」清單(`visited[16]`:CTY 號 + 最後座標 + section)原本只在記憶體 → **存讀檔不持久**。
- 修正:`dq3_save_pos` 加 `n_visited` + `visited[16]`,存檔 magic `DQ3SAVE7`→`DQ3SAVE8`
  (結構大小欄位自動擋舊檔),3 個存檔點 + 2 個讀檔點接上。
- `test_save` 加 round-trip 斷言(去過 CTY3/CTY7 → 存讀一致),17/17 綠。

## 6. 教訓

- **位址基準先驗證再反組譯**:同一專案不同來源的位址可能是 logical、file、或不同 window;
  反錯段會追進隔壁函式還以為機制不存在。用「已知 operand 值的引用落點」反推基準最快(對齊
  記憶 `re-disasm-offset-base-mismatch`)。
- **別把兩層機制混成一層**:可見性旗標(劇情)與日夜清單是獨立兩層;卡在「找日夜旗標」是因為
  把日夜硬塞進旗標層。往上文多看幾行就看到真正的 selector。
- **語意分歧讓資料當裁判**:`[0x526c]` 哪個值是白天,反組譯要再追好幾跳;dump 真實 CTY 的兩份
  清單 count 一眼定案(白天人多)。對齊本專案「反編當 oracle、資料是真相」紀律。
