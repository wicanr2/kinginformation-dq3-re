# 123 — 戰鬥／日夜靜態反組譯閉環

> 2026-08-10。這份文件只記錄原版 `DQ3.EXE` 的靜態證據，不把一串可讀的
> 反編譯結果當成玩家可見 parity。`confirmed` 表示原始 bytes、caller／writer／consumer
> 已閉合；`strong` 表示多條靜態證據已一致但尚缺同狀態實機畫面；`hypothesis` 表示待驗假說；
> `unknown` 表示不能從目前輸入安全推出。production JSON 不得以本文件的
> `strong`／`hypothesis`／`unknown` 欄位填入合理預設。

> [!IMPORTANT]
> **歷史／已訂正。** 本文件保留 2026-08-10～11 的調查順序，因此前半部若寫「五個
> story flag runtime unknown／missing」、campaign 尚未重播完成或 formation 仍缺接線，
> 都是當時的 checkpoint，不是現行待辦。這些結論已由本文件「runtime 接線更新」與
> 「未閉合項目的 IDA 9.4 收斂」兩節訂正；目前完成度以
> [`docs/74`](74-ebiten-remake-completion-plan.md) 最新 checkpoint 為準。固定編隊／必經頭目
> 背景見 [`docs/128`](128-battle-background-selector-re.md) 與
> [`docs/129`](129-required-boss-backgrounds-re.md)，地表四人 HUD 見
> [`docs/130`](130-party-field-hud-re.md)。這些後續結果仍只按各文件標示為 E2、V1 或
> near-state V2，不構成全遊戲 V3。

## 輸入、工具與位址口徑

本輪只讀取下列原始輸入，未在原位置解包、patch 或覆寫：

| 輸入 | 大小／雜湊 |
|---|---|
| `assets_raw/DQ3.EXE` | 115,282 bytes；SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| `assets_raw/D3MNS.DAT` | SHA-256 `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed` |
| `assets_raw/DQ3.PAL` | SHA-256 `178a9ca809a33108d8e2a6430796eae8909d98e514fe543fc1923139110389c4` |
| `assets_raw/MNSBK.PAL` | SHA-256 `e2985503b3eb5256bcaf44f2f1b0a0b1ab124e66873c2de210e0d25ffd75110d` |
| `assets_raw/CTY00.DAT` | SHA-256 `ac8427c5fafcad4e29246dd3c2796c476bb5ad93e53dd7c7127a6b2faa31a836` |
| `assets_raw/DQ3MNS.SHP` | 1,377,210 bytes；SHA-256 `bad40e552343141b75191a5a9576adddb7a16ddf1378fe139eba4c4e15ba8bfd` |
| `assets_raw/FVOC.VCX` | 72,531 bytes；SHA-256 `47be27d2244e314ca6f2a21d3fa61a95d4e823f9085ed3d21808ac25d66b39c5` |
| `assets_raw/NVOC.VCX` | 143,156 bytes；SHA-256 `53c710f356b728e97287bfb64af7f64cf2067b28f83a9dfe32f99a6eb61f4156` |

IDA Pro 9.4（Hex-Rays `9.4.0.260610`）在一次性 Docker 容器內建立唯讀輸入的工作副本；
`.i64`／`.asm` 位於 `/tmp/dq3-ida-static`，未加入 Git。IDA listing 的線性位址範圍為
`0x10000..0x2aee0`。主程式碼段採：

```text
IDA linear = logical + 0x10000
file       = logical + 0x1370
DGROUP     = file 0x16140 + DGROUP logical offset
```

分段 runtime（`seg006`、`seg014`、`seg015` 等）保留 IDA linear 與 segment 名稱，不能把
其數值誤當成主段 file offset。所有結論仍保留原始函式名、原始 DGROUP 位址及資料範圍。

## 一、戰鬥正式回合與逐動作順序

### 1. 玩家可達的主鏈

| 原始函式 | 位址（IDA／logical／file） | 靜態閉環 |
|---|---|---|
| `sub_1BDDF` | `0x1bddf`／`0xbddf`／`0xd14f` | 進入正式戰鬥，建立遭遇、清理狀態、繪製背景／隊伍／敵人，呼叫回合 runner |
| `sub_1C08B` | `0x1c08b`／`0xc08b`／`0xd3fb` | 收集玩家命令、建立 action queue、逐筆呼叫 actor consumer；戰況未結束時回到下一回合 |
| `sub_1C34F` | `0x1c34f`／`0xc34f`／`0xd6bf` | 將存活玩家與敵人寫入 `DGROUP:0x0cf6`，依速度與 RNG 排序 |
| `sub_1A973` | `0x1a973`／`0xa973`／`0xbce3` | 讀一個 actor entry，處理死亡、失能、睡眠、逃跑、施咒／物理分派，再返回 queue |
| `sub_1C425` | `0x1c425`／`0xc425`／`0xd795` | 戰後分配經驗／金錢、檢查掉落 gate、顯示獎勵，清理戰鬥畫面 |

另有一個由事件表／分段跳轉進入的 AI-only runner（IDA 只保留標籤，不替它虛構函式名）：
`loc_190D7`（IDA linear `0x190d7`／logical `0x90d7`／primary file `0xa447`）同樣先呼叫
`sub_1C34F`，再逐一呼叫 `sub_1A973`；一輪結束後遞增 `DGROUP:0x26fc`，最多重建
`0x0c` 輪。這條路徑補證「固定 12 輪上限」而非 boss 多次行動：每輪仍是一個 queue
entry 一次，沒有在同一輪把同一 actor 重複呼叫 N 次；其事件入口與玩家可見用途維持
`strong`，不能把它誤命名成一般戰鬥或 boss repeat handler。

`sub_1C08B` 對 action entry 做 `actor code <= 0x1e` 與 `sub_1A973`／
`sub_1B3F3` 的分派；對每個尚存的 queue entry 只直接呼叫一次 actor consumer。新的回合
再由 `sub_1C34F` 重建 queue。上述 `loc_190D7` 也只是在回合外重建 queue，並未改變
entry 內的呼叫次數。這是「每個活躍 actor 每回合一次」的 `confirmed` 結論；目前沒有
找到「boss 依資料欄位把同一 actor 在同一回合再呼叫 N 次」的外層迴圈。

因此：

- 一般速度排序與 action order：`confirmed`。
- 一回合內死亡／睡眠／逃跑會跳過本次行動：`confirmed`。
- 「boss 多次行動」的數量、條件、間隔與特殊 cue：`unknown`。不能把 `D3MNS.DAT` 的
  未命名欄位、攻略描述或舊 C remake 的 boss 特例升格成 production 設定。

### 2. action queue 的原始資料流

`sub_1C34F` 的每筆 queue record 為兩個位元組：

1. 玩家：從 party member `+0x24` 取速度；通常右移一位，若 member `+0x22` bit2
   成立則保留原尺度；呼叫 RNG，再將結果寫入 queue `+1`。
2. 敵人：從 active enemy record `+0x0b`（特殊 battle mode 另讀 `+0x07`）取速度，
   右移一位後加 RNG，寫入相同 queue。
3. queue record `+0` 保留 actor index；以 `[si+1]` 與 `[si+3]` 比較並交換整個 2-byte
   record，重複 pass 直到排序完成。

速度欄位的「除二」與 mode bit、RNG 加值不是 renderer 行為，而是行動次序的原版資料流；
若日後把它資料化，必須以 `queue_speed_shift`／`rng_range` 等具名欄位保存並附同一輸入
hash，不能以「角色敏捷越高越先」的文字近似。

### 3. 敵人 action 的死亡／睡眠／逃跑／分派

`sub_1A973`（`file 0xbce3`）的閉合順序如下：

| 順序 | 原始讀寫 | 結果／等級 |
|---|---|---|
| 1 | actor index `-1`，stride `0x16`，HP `active+0x2334` | HP<=0 或 status bit0 時直接返回；`confirmed` |
| 2 | status `0x233f` bit7，RNG 與 `0x64` | 失敗時清 bit7 並寫醒來訊息；成功時保留睡眠並跳過本輪；`confirmed` |
| 3 | D3MNS record `DS:0x0d78 + monster_id*0x29` 的 `+0x17`、`+0x18` | 以 party strength 與 RNG 判斷逃跑；成功清 HP、減 group／total count、呼叫 `sub_1B23F` 消除 sprite；`confirmed` |
| 4 | record `+0x0d` 與 RNG | 小於機率欄位進 `sub_199DC`（從 `+0x0e..+0x13` 48-bit spell mask 選咒文），否則進 `sub_1AC05`（物理／特殊）；`confirmed` |
| 5 | `sub_18222`、HP 上限 `active+0x36` | 行動後結算回復／傷害並封頂；`confirmed` |

`+0x08`、`+0x09`、`+0x0b` 由 `sub_1AB2C` 複製到 active enemy action record 的
`+0x0b`、`+0x0c`、`+0x0e`；其 production 語意仍依既有 D3MNS schema 的證據等級，
不可因欄位順序看似攻擊／防禦／速度就重新命名。

## 二、戰鬥 timing、動畫與音效

### 1. 硬體 tick 與文字／動作停頓

| 函式 | 原始行為 | 等級 |
|---|---|---|
| `sub_1C673`（`logical 0xc673`） | 清 `DS:0x0005`，等待硬體 tick `DS:0x0005 >= 6` | `confirmed` |
| `sub_1C62A`（`logical 0xc62a`） | 等待 tick `>3` 後呼叫 `sub_1C642` | `confirmed` |
| `sub_1C642`（`logical 0xc642`） | `259b=4-cl`；以 `sub_211B6` 每列寫 20 個 glyph、每列 `+0x10`，共 `cl` 列 | `confirmed` |
| `sub_1C59B`（`logical 0xc59b`） | 畫敵名／sprite 後清 tick，等待 `DS:0x0005 == 6` 才返回 | `confirmed` |
| `sub_1B3C3`（`logical 0xb3c3`） | 六次呼叫 `sub_1B220`；偶數輪等待 tick `>=2`，奇數輪等待 `>=1` | `confirmed` |

`sub_1C673`／`sub_1C642` 是 actor entry 前的固定演出與訊息繪製節點；它們不能直接
等同於「每種攻擊動畫長度」。真正的 cue-to-frame 對拍仍需要 DOSBox 同狀態 frame trace，
故逐動作玩家可見 timing 目前為 `strong`，不是 V3。

### 2. SHP sprite 的遮罩、重畫與 frame 結論

- `sub_1B16F`／`sub_1B19E` 依 monster id 從 SHP offset table 載入一份 sprite payload。
- `sub_1B1FE` 以 active enemy record `+0x03` 作畫面位置，從 `DS:0x2477` 取該 monster
  的 sprite pointer，再依序呼叫 `sub_1B31A` 與 `sub_1B2AF`。
- `sub_1B31A`／`sub_1B37C` 先走 EGA 四 plane 的 RLE AND-mask；`0xc0` 高位決定重複／
  透明／literal，不能用 palette 色號 0 當透明。
- `sub_1B2AF` 寫入四 plane 顏色；兩個繪製階段前後呼叫 `sub_20CC6`，由 `sub_20E5E`
  將待更新範圍同步到畫面。
- 同一個 monster id 的正式戰鬥 draw path 沒有讀取「frame index」或第二份 SHP frame；
  `sub_1B3C3` 是同一 payload 的六次遮罩重畫／閃爍時序，不是已證實的多幀角色動畫。

結論：SHP 載入、透明遮罩、位置、重畫 timing 為 `confirmed`；「每個攻擊／咒文的不同
動畫 frame」在目前 EXE 靜態資料中沒有閉合，標 `unknown`，不可在 JSON 產生 `animation_id`
或由動作名稱推導 sprite frame。

### 3. VOC 音效呼叫鏈

原版戰鬥的音效入口不是 `sub_1683A`。`sub_1683A` 只清 `DS:0x0005` 並等待 tick
等於 7；它是一般 timing primitive。音效的靜態鏈為：

```text
battle caller sets BP cue
  → sub_20770 (seg004, IDA linear 0x20770)
  → DS:0x253e segment, [BP << 2] 取得資料指標
  → sub_22CF5 (Sound Blaster command 6)
  → sub_22F60 → sub_236B4 (seg014 driver dispatch)
  → VOC bank playback
  → sub_208E2 waits driver state before the next action/message
```

`sub_208A7` 依 `word_25B45` 選擇 `FVOC.VCX` 或 `NVOC.VCX`，讀入 `word_2730e` 指向的
資料段；這是外部 VOC 音效 bank 的 `confirmed` 證據。`sub_20770` 對 `BP < 0x1e` 走
四位元組資料指標表，`BP >= 0x1e` 另走 `sub_20950` 的長資料路徑；兩者不可混成同一
「音效編號表」。

戰鬥函式中可直接回查到的 cue 數值（只記 call-site，不替 cue 命名）：

| call-site | BP cue | 原始後續 |
|---|---:|---|
| `sub_1A973` 逃跑成功 | `0x15` | 播放後印 `rec 0x15b`，再等待 `sub_208E2` |
| `sub_1AC05` 物理／特殊入口 | `0x04` | 先播放，再印 `rec 0x14b` |
| `sub_1ACCE` 傷害／受擊分支 | `0x01`、`0x04`、`0x09` | 依分支在 `sub_1B3C3`／文字前後等待 |
| `sub_1B7B0` 效果收尾 | `0x06` | `sub_208E2` 後進 resistance／效果表 |
| `sub_1D86D`、`sub_1D881` | `0x08`、`0x10` | 先播 cue，再以 `sub_1D895` 做效果畫面 |
| `sub_1D8D6` | `0x01` | `sub_1B3C3` 後等待 driver 結束 |

這些值與 `FVOC.VCX`／`NVOC.VCX` 指標表的 writer／consumer 已 `confirmed`；cue 的
「劍擊／咒文／受傷」中文名稱、實際 PCM 波形與同狀態玩家可見的聲音時長仍需將 VOC
資料索引與實機錄音逐項對拍，標 `strong`／V3 gap。不可因 cue 數值連續或攻略名稱
自行填 `audio.json` 的新語意。

另外以 VCX 檔案開頭的 little-endian 32-bit offset table 交叉讀取 raw block。以下只記錄
不帶語意的可回查 descriptor（cue 是 `BP` 原值，offset 是 VCX file offset；`block_type=1`
與三-byte length 直接取自 payload header）：

| bank | cue 1 | cue 4 | cue 6 | cue 8 | cue 9 | cue 16 | cue 21 |
|---|---|---|---|---|---|---|---|
| `FVOC.VCX` | `0x4e6/len1163/rate102` | `0x18a2/1155/102` | `0x215e/1224/107` | `0x3947/4998/102` | `0x4cd2/2978/107` | `0xba81/4456/105` | `0x11146/2568/165` |
| `NVOC.VCX` | `0x57d/718/87` | `0x1037/5177/107` | `0x28d7/467/107` | `0x3dcb/4998/102` | `0x5156/3579/107` | `0x12de7/8924/100` | `0x1f973/1996/145` |

這些長度／rate byte 仍不是中文音效名稱，也不等於播放器已在 Ebitengine 中逐樣本
同速重現；它們把「cue→VCX raw block」的靜態鏈補成 `confirmed`，PCM 解碼、硬體取樣
率換算與玩家可見停頓仍列 V3。

## 三、抗性與 D3MNS packed 欄位

`sub_1D7FB`（`IDA linear 0x1d7fb`／`logical 0xd7fb`／`file 0xeb6b`）的資料流完全閉合：

1. 輸入 `BX=monster_id`，乘 `0x29`，以 `DS:0x0d78 + id*0x29 + 0x19` 為 packed resistance
   起點。
2. 輸入 `AL=effect/spell id`，在 `DS:0x36e5` 逐項比較 threshold；相等時先 `inc BX`。
3. 以 `BX & 3` 選 bit pair，`BX >>= 2` 選 `+0x19..+0x1d` 的 byte，旋轉後 `and AL,3` 得
   resistance class `0..3`。
4. 以 `DS:0x36f5[class]` 轉成 RNG 成功門檻並返回 caller。

原始 table 的可回查片段（DGROUP file base `0x16140`）：

```text
DS:0x36e5  file 0x19825:
03 06 09 0d 10 12 16 17 19 1b 1c 1d 1f 23 25 26 ...

DS:0x36f5  file 0x19835:
00 44 b4 ff 11 37 21 37 91 37 99 37 31 37 49 37 ...
```

### 2026-08-11 抗性 decoder 勘誤

前一版 `internal/dq3data.Monsters.SpellChance` 曾把這段誤寫成 first-`>=` bounds；
那不是原始 `cmp`／`inc BX` 的控制流。以下保留 raw table 與舊 parser 被推翻的原因，
不把既有錯誤測試結果當成證據。

`sub_1D7FB` 的比較是 `AL < table[BX]` 停止、相等先 `inc BX`；因此對原版 60 筆
effect／spell id（`0..59`）的實際類別區間是：

```text
0..2→0, 3..5→1, 6..8→2, 9..12→3, 13..15→4, 16..17→5,
18..21→6, 22→7, 23..24→8, 25..26→9, 27→10, 28→11,
29..30→12, 31..34→13, 35..36→14, 37→15, 38..59→17
```

類別 16 因原始連續的 `0x00` sentinel 為空；`68`／`0xb4`／`0xff` 之後的類別 18／19
只在超出目前 60 筆 effect descriptor 的 id 才會觸及。這個 `0` sentinel 不能被 Go
decoder 的 bounds 正規化刪掉；本輪已以 `id=2/38/59` 及怪物 121 的 packed bytes
加入 parity test。

因此 class 0／1／2／3 的門檻為 `0/0x44/0xb4/0xff`（0/68/180/255），是
`confirmed`；呼叫端的比較也要保留，不能只存一個抽象「成功率」：

- `sub_1D2D3`：`AL > threshold` 跳到無效／抗性路徑，故效果分支是 `AL <= threshold`。
- `sub_1D338`／`sub_1D3B7`：`AL < threshold` 進 `sub_1D421` 效果 handler，`AL >= threshold`
  走抗性訊息／無效分支。

這些 raw comparator 是 `confirmed`；`+0x19..+0x1d` 的每一個 bit pair 對應哪個中文
咒文名稱，必須按 effect descriptor 與玩家可見結果逐項建立 sidecar。在該對映未完成前，
不得把 packed bytes 展開成未附證據的 `fire_resist`／`sleep_resist` 等 Go 欄位。

## 四、遭遇、formation、位置與 boss 特例

### 1. 隨機遭遇與群組建構

`sub_1A7D5`（`logical 0xa7d5`／`file 0xbb45`）在非強制事件地表：

- 以座標高／低 nibble 查 `DS:0x4966` 區域表，從 `region*0x20` 及 RNG 選八個候選
  monster id（候選基址 `DS:0x29826`）。
- 將 `word_272e3` 設為 `0x26`（38 點 encounter budget），以 D3MNS `record +0x28`
  權重遞減；每次把 `(monster_id,count)` 寫入 `byte_270f1`，直到 budget 耗盡或沒有候選。
- `sub_1BF35` 將 formation record 的 group count、header `raw[1]`／`raw[2]` selector 及 sprite id
  複製到 `DS:0x231f`、`0x2321`、`0x0d71`／`0x0d73`，再呼叫 `sub_1AAA1`／`sub_1AB2C`
  逐隻建立 stride `0x16` 的 active enemy action record。

候選表、權重、38 點 budget、group count 與 fixed-formation raw selector 已是 `confirmed`。固定
編隊 selector 的 archive page／palette consumer 已另於 [`docs/128`](128-battle-background-selector-re.md)
閉合；它不代表 generic terrain selector 已知。每個
formation 的 `active enemy +0x03` 位置欄位被 `sub_1B1FE` 作為畫面座標消費，故「位置
不是 renderer 自己排版」也是 `confirmed`；完整 pack 內每個 formation 的位置 byte、
混合群組與原始 table record 仍需逐筆 sidecar，標 `strong`。

### 2. boss 多次行動的結論

正式回合只有 `sub_1C34F` 建 queue → `sub_1A973`／`sub_1B3F3` 一次 → queue 下一筆的
閉環。固定 boss formation、連戰（例如多場 boss queue）與「一場後接下一場」是不同於
「同一 boss 每回合多行動」的機制；不能把連戰 queue 當成多次行動。若未來在另一個
caller 找到 repeat count，必須附 `writer→state→consumer→可見時序`，否則維持 `unknown`。

## 五、戰後經驗、金錢與掉落

`sub_1C425`（`file 0xd795`）的勝利後資料流為：

1. 以 party count `BP` 將 `DS:0x24f6:0x24f8` 做 32-bit 除法，將經驗／金錢分配到
   存活隊員的 record `+0x32/+0x34`；`sub_1BCC4` 讀 D3MNS `+0x21` 與 `+0x23`。
2. `DS:0x2518 bit1` 成立時直接跳過掉落段（scripted／第二戰抑制 gate）。
3. 否則取第一個 formation monster `DS:0x2321`，乘 `0x29`：
   - `DS:0x2518 bit0` 未設時，讀 D3MNS `record +0x25`，呼叫 RNG，`AL > threshold`
     則沒有掉落；
   - 通過或 bit0 已設時，讀 `record +0x26` 作 item raw id，寫 `DS:0x2591/0x2593`，
     顯示 `rec 0x176`／`0x10b`，必要時等待玩家確認。
4. 最後呼叫 `sub_1D94C`／`sub_2111B`／`sub_1F604` 收尾。

`+0x21` 經驗、`+0x23` 金錢、`+0x25` 掉落門檻、`+0x26` 掉落 item raw id 的
writer／consumer 已 `confirmed`。比較指令明確是 `AL > [record+0x25]` 才跳過，因此成功
條件是 **`AL <= +0x25` 的 inclusive RNG threshold**；`0xff` 為所有 8-bit RNG 結果
都通過。這個 raw gate 可以安全記錄為 `drop_threshold_inclusive`，剩下的只是 UI／JSON
是否要另提供「百分比」等正規化顯示名稱，不能反向改變原始比較或用猜值填補資料。

## 六、日夜 clock 與精確 palette

### 1. clock writer 與 gate

`sub_1EE23`（`IDA linear 0x1ee23`／`logical 0xee23`／`file 0x10193`，caller
`sub_193E3+0xd9`）是每次有效 overworld 更新的 clock writer：

```text
if [DS:0x4f1f] == 0xff: return
inc word [DS:0x251f]                  ; frame/tick bookkeeping
if [DS:0x4f2d] == 1: return            ; battle/scene gate
if [DS:0x5051] == 1: return            ; transition gate
inc word [DS:0x251d]                  ; day/night clock
if clock == 0x0078: byte [DS:0x526c] = 0
if clock == 0x00f0: clock = 0; byte [DS:0x526c] = 1
if (clock % 0x14) != 0: return
call sub_1EE76
```

這是有效 overworld clock 的 `confirmed` 數值：`0..0xef` 共 240 tick；到 `0x78`
（120）選夜表，到 `0xf0`（240）歸零並選日表。戰鬥／轉場 gate 會阻止 clock 增加，
不是每一個 render frame 都累加。

`DS:0x526c` 的 consumer 在 `sub_131BD`：值為 1 取 section header `+0`（日 NPC 表），
值為 0 取 `+2`（夜 NPC 表）。因此 Go 現行四 phase 的近似並不能拿來回寫原版的
`[0x526c]` 語意；原版這個欄位是二值 table selector。

### 2. palette bank 選擇與硬體上傳

`sub_1EE76`（`logical 0xee76`／`file 0x101e6`）及 `sub_1EE9B`
（`logical 0xee9b`／`file 0x1020b`）都做：

```text
phase_index = [DS:0x25c5 + (clock / 0x14)]
palette_ptr = DS:0x3232 + phase_index * 0x30
```

`sub_1EE76` 接著呼叫 `sub_20A3A`（BIOS `int 10h, AX=0x1012`）將 16 色、`0x30` bytes
的 bank 上傳 DAC；`sub_1EE9B` 只更新 pointer，供場景載入 consumer。`DS:0x25c5`
的 12-byte 原始索引表（file `0x18705`）為：

```text
01 00 00 00 01 02 03 04 04 04 03 02
```

`sub_1ECDC` 將 `DQ3.PAL` 的 `0xf0` bytes 載入 `DS:0x3232`，將 `MNSBK.PAL` 的
`0x1e0` bytes 載入 `DS:0x3352`；`sub_1C688` 另以 `DS:0x3352 + [DS:0x0d73]*0x30`
選戰鬥背景 bank。clock→bank index、palette buffer、DAC upload、戰鬥 background 選擇
均為 `confirmed`；原版畫面每一個 phase 的 RGB／EGA register 與現行 runtime PNG 的
逐像素比對仍是 V3 工作，不能用 `DarkenPalette` 近似宣稱 palette parity。

### 3. 日夜與 NPC flag 的共同 consumer

`sub_131BD`（`logical 0x31bd`／`file 0x452d`）的每筆 NPC record 為 7 bytes；
record byte5 作為完整 story flag id：

```text
id = byte5
mask = 0x80 ror (id & 7)
byte_index = id >> 3
if test byte [DS:0x4f70 + byte_index], mask == 0: skip NPC
else: load NPC
```

這裡的 `id` 是完整 byte，不可只取低三位。`[DS:0x4f70]` 的 SET／CLEAR／GET API 由：

| 函式 | 位址（logical／file） | 行為 |
|---|---|---|
| `sub_16EDF` | `0x6edf`／`0x824f` | `id>>3`、`0x80 ror(id&7)` 後 OR，SET |
| `sub_16EF4` | `0x6ef4`／`0x8264` | 同一 mask 後 AND 反相，CLEAR |
| `sub_16F09` | `0x6f09`／`0x8279` | 同一 mask TEST，GET 0／1 |

所以原版的通用靜態 state chain 是：

```text
literal／data-driven writer → sub_16EDF/16EF4 → DGROUP [0x4f70]
→ sub_16F09 或 sub_131BD → NPC／事件／場景可見結果
```

這個 chain 證實 flag state 的 storage 與 generic consumers；若某一旗標沒有 NPC record，
不能反過來宣稱它「沒有 consumer」，必須再查 `sub_16F09` 的 caller、event table 與
直接 bit 操作。`docs/71` 的 21 個 literal writer 如下。

## 七、21 個條件 flag 的原版 writer ledger（歷史 runtime 欄已訂正）

> [!WARNING]
> 下表的原版 SET 位址與 writer 證據仍有效，但 `runtime` 欄凍結於接線前狀態。表中
> `unknown`／`missing` 已被本文件後段「runtime 接線更新（現行 E2）」推翻；保留原文是為了
> 追溯當時為何需要補接線，不得據此重新開啟已完成的 flag JSON／engine 工作。

下表的 `SET file` 是 `mov bx, literal` 到 `call sub_16EDF` 的原始 file offset；
同一 flag 的多個位置表示不同事件階段或重試路徑。`原版 writer` 全部是 `confirmed`；
`runtime` 欄只表示目前 Go pack 是否已接，不把缺少 runtime 接線誤寫成原版沒有 writer。

| flag | 原版 SET（file） | IDA writer function | 靜態 consumer／runtime邊界 |
|---:|---|---|---|
| `0x0d` | `0xfab0` | `sub_1E713` | 原版 writer 已證實；事件／直接 GET 尚未閉合，runtime `unknown` |
| `0x0e` | `0x7830` | `sub_16494` | writer 後呼叫 `sub_131BD`；runtime 只接部分 Ortega 流程 |
| `0x10` | `0x56c8` | `sub_14312` | item／鏡事件後 reload NPC table；runtime `confirmed` |
| `0x11` | `0x7563` | `sub_16139` | writer 後呼叫 `sub_131BD`；runtime clear-only，SET 事件 `unknown` |
| `0x14` | `0x736a` | `sub_15F5D` | rescue／boss 事件 writer；pack primitive 已有，V3 仍待 |
| `0x16` | `0x716a` | `sub_15DAF` | guided passage writer；pack primitive 已有 |
| `0x17` | `0x156b` | `sub_1010B` | opening／mother escort writer；正式 trace 已到達 |
| `0x18` | `0x162a` | `sub_1024C` | opening／throne writer；NPC loader consumer generic |
| `0x19` | `0x5622`, `0x5673`, `0x719c` | `sub_14104` 等 | dark-lamp／visibility 多階段 writer；runtime 已接主要路徑 |
| `0x1a` | `0x760a` | `sub_1622A` | 原版 writer 已證實；巴拉摩斯戰後世界／NPC 可見性交易，runtime `missing` |
| `0x1b` | `0x7610` | `sub_1622A` | 同一巴拉摩斯戰後交易，不能以彩虹合成旗標代替 |
| `0x1c` | `0x7232` | `sub_15E6C` | guided passage completion；pack 已接 |
| `0x1f` | `0x6d4a` | `sub_15979` | staged boss gate；第一／第二戰 runtime 已接主要交易 |
| `0x20` | `0x7053` | `sub_15CB6` | staged boss post-state；runtime 仍需完整 V3 |
| `0x21` | `0x6c8e` | `sub_15904` | mirror／Samanosa state；runtime 已接成功 clear |
| `0x22` | `0x5735` | `sub_14312` | mirror success visibility；runtime 已接 |
| `0x23` | `0x6ea3`, `0x6f1d` | `sub_15AA2` 等 | merchant settlement／founder；runtime 已接主要 path |
| `0x24` | `0x6c88` | `sub_15904` | 原版 writer 已證實；事件語意與 runtime `unknown` |
| `0x25` | `0x73f0` | `sub_16004` | Baharata rescue post-state；pack 有 raw completion flag |
| `0x26` | `0x5336` | `sub_13C6D` | Noaniel wake／quest item；runtime 已接 |
| `0x27` | `0x6688` | `sub_1527E` | Romaly temporary role；runtime 已接 |

### flag ledger 的限制與下一個證據閘門

- 以上完成「原版 literal writer → `[0x4f70]` API」的靜態閉合；`0x0d`、`0x11`、`0x24`、
  `0x1a`、`0x1b` 的**事件語意／runtime production writer**仍不可猜。
- `sub_131BD` 的 generic NPC consumer 已確認，但不是每個 flag 都必然掛在 NPC record；
  其餘旗標需以 `sub_16F09` caller、CTY／事件 table 及玩家可見副作用逐筆補 sidecar。
- `0x15` 是 handler40 性別／建城者 visibility flag，不屬上述 21 個 literal 清單；
  `0x13` 勇氣試煉 completed 目前仍無 confirmed writer，也不得塞入本表。
- 只有補上 caller→writer→state→consumer→同狀態畫面／聲音，才能把 runtime 欄由
  `strong` 升到 D3／E3。靜態 writer 本身不代表 remake 已完成該事件。

## 八、當時可宣稱的完成度與剩餘工作（歷史／已訂正）

> [!WARNING]
> 本節是接線前的靜態研究結算。五個 story flag、formation 與 campaign 的現行結果請直接讀
> 本文件後段的 E2 runtime 更新、IDA 收斂，以及 [`docs/74`](74-ebiten-remake-completion-plan.md)
> 最新 checkpoint；本節只保存舊證據邊界，不再是 worklist。

### 靜態分析已完成（E1／D2）

1. 正式戰鬥的 queue、速度排序、逐 actor 分派、死亡／睡眠／逃跑／咒文／物理分支。
2. SHP 的 sprite payload、RLE AND-mask、四 plane draw、位置欄位與六輪重畫時序。
3. Sound Blaster VOC bank 載入、`BP` cue 指標、driver dispatch 與等待點。
4. D3MNS packed resistance 的 bit pair 解碼及 `0/68/180/255` 門檻。
5. formation 候選／權重／38 點 budget、固定編隊 raw selector、掉落 gate 與 item raw id。
6. 日夜 `0..239` clock、120/240 邊界、二值 day selector、12 段 palette index、DQ3／MNSBK
   palette buffer 及 DAC upload。
7. 21 個條件 flag 的原版 SET literal ledger 與 `[0x4f70]` state API；generic NPC filter。

### 不能由本輪靜態結果宣稱完成

- 同狀態 DOSBox 的逐動作 frame／音效錄音、每個 cue 的實際波形與可見動畫長度（V3）。
- 每一個 formation 的完整 pack 位置與混合群組 parity（部分 `strong`）。
- boss 同回合多次行動（`unknown`，目前正式鏈只證實每 actor 一次）。
- `+0x25` 的正規化顯示名稱與所有 scripted drop 抑制組合（inclusive raw gate 已證實；
  `DGROUP 0x2518 bit1` 抑制及 bit0 強制路徑仍需逐場景 parity）。
- 21 flag 中 runtime 尚缺的事件（尤其 `0x0d`、`0x11`、`0x24`、`0x1a`、`0x1b`），以及
  已接事件的同狀態 NPC／palette／音效 V3。

上述 raw table 已另存為帶輸入 hash、IDA 位址基準、consumer、推論等級的非 production
sidecar（`docs/data/dq3-static-battle-daynight-evidence.json`），並已通過原始檔與 palette
index parity。下一批若要寫入 game-pack，仍須逐欄新增 schema／reference validation 與
原始 EXE／DAT parity test；未知值一律 fail closed，不因這份靜態 ledger 已完成而加入 Go
常數或文字。

## 九、2026-08-11 靜態補證：formation、效果名稱與五個 flag consumer

本節是追加勘誤，不覆寫前面的時間序列。輸入仍為同一份 `DQ3.EXE`（SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`），只在 Docker 內以
IDA Pro 9.4 listing／原始資料唯讀追查。

### 1. formation raw sidecar 已補齊

`sub_1A7D5` 先從 `DGROUP:0x4966`（file `0x1aaa6`）取得區域值，再執行
`dec bx; shl bx,5`；因此候選基址 `DS:0x29826` 的第一列對應區域值 `1`，不是區域值
`0`。本輪確認原始區域表 256 bytes 的最大有效值為 `0x1b`，所以使用區域 `1..27`、每區
4 列、每列 8 bytes，實際使用 `0x360` bytes（`DS:0x4a56`／file `0x1ab96`）。所有
108 列 raw hex 已放入 sidecar `formation.encounter_candidate_table_raw`；候選 8 bytes
仍保持未命名，不能把欄位看成座標或怪物數量。

固定事件編隊也已逐筆保存 `DS:0x4eb5..0x4f03`（file `0x1aff5..0x1b03f`）的 13 個
raw record，包含 count byte、兩個未命名 header byte、`count` 組 `{monster_id,count}`。
`sub_1BF35` 的 loader 與各 caller 已列在 sidecar；原始 byte 不以「背景／頁面」等推測
名稱覆蓋。`sub_1AAA1`／`sub_1AAD5`／`sub_1AB2C` 的位置資料流則是：

```text
word_272E3 = 0x26
  → 固定／隨機群組先扣 D3MNS +0x28 × count
  → sub_1AAD5 先加 2
  → sub_1AB2C 寫 active +0x03，再加 2 × D3MNS +0x28
  → sub_1B1FE 以 active +0x03 算 EGA destination
```

這使「位置由 active record 提供」達到 `strong` writer→consumer 閉環；`+0x03` 的精確
螢幕座標單位與每一列的視覺排列仍是 `unknown`，不能用 Ebitengine 的目前排版反推。

### 2. 抗性 descriptor 的中文名稱只做 record 對映

`sub_1CB3C` 以 `spell/effect id = DS:0x2591 - 0x79`，乘 3 後讀 `DS:0x37c3`；
`sub_1D6C7` 也直接以相同 3-byte descriptor 的 first byte 作 MP／消耗比較。故
`DS:0x37c3`（file `0x19903`）的 60 筆 raw descriptor 與 `D3TXT00.TXT` record
`0x79..0xb4` 的中文名稱對映已是 `confirmed`。sidecar 保存 entry count、raw hex、
`mp_raw/base_raw/flags_raw` 契約、文字檔 hash 與 60 個名稱。

這只證實「id→原始中文 record」；它沒有證實名稱所代表的抗性元素、狀態或玩家可見
效果。因此 `+0x19..+0x1d` packed pair 仍只輸出 resistance class／threshold category，
不得在 production JSON 產生未經 consumer 對拍的 `火焰抗性`、`睡眠抗性` 等欄位。

### 3. 五個 story flag 的靜態事件與 generic NPC consumer（歷史 runtime 欄已訂正）

> [!WARNING]
> 下表在靜態 deep dive 當下尚未接入 Go，故保留 `runtime unknown` 的原始判定；後續
> `events.json.story_flag_runtime_events` 與正式 command／battle tests 已完成有限 E2
> transaction。原始 handler、SET／CLEAR 與 NPC consumer 仍是有效證據，但舊 runtime 欄
> 不能再用來判定現行程式缺功能。

事件跳表 `seg016:off_28984` 由 `sub_196D2` 以「index + 1」呼叫。五個旗標的 writer、
事件副作用與 generic consumer 已閉合如下；「production runtime」欄仍不代表 Go 已接線：

| flag | handler | 靜態 writer／事件副作用 | generic consumer | 等級 |
|---|---|---|---|---|
| `0x0d` | handler 74 `sub_16346` | GET `0x0e1/0x136` → `sub_1E713`；CLEAR `0x0f5`、SET `0x0d`，場景重載與多段 map 動畫 | `sub_131BD` 讀 CTY record byte5；`sub_134AB` 寫 byte6、`sub_11971` 重畫 | writer／generic consumer `confirmed`；runtime `unknown` |
| `0x11` | handler 69 `sub_16139` | GET `0x8e` 與 `0x8f..0x94` → map cell 寫 `0x80`；SET 後 `sub_134AB → sub_131BD → sub_11971 → sub_12C9C`，返回前 CLEAR | 同一 full-byte5 filter；CTY70 有代表 record | transient chain `confirmed`；runtime `unknown` |
| `0x1a` | handler 70 `sub_1622A` | 固定編隊 `DS:0x4edf` 戰鬥後 reset party、`0x526c=1`、clock=0、palette reload，再 SET `0x1a`／`0x1b`、重建 DQ3BLK／DQ3UND | full-byte5 filter；CTY25 有代表 record | writer／generic consumer `confirmed`；runtime `unknown` |
| `0x1b` | handler 70 `sub_1622A` | 與 `0x1a` 同一 post-battle world reset 交易，兩旗標同時 SET | full-byte5 filter；CTY00 有代表 record | writer／generic consumer `confirmed`；runtime `unknown` |
| `0x24` | handler 33 `sub_15904` | `sub_15002`、顯示 rec `0x0bff`、CLEAR `0x3d`、SET `0x24`／`0x21`、`sub_15010` 關閉 | full-byte5 filter；CTY43 有代表 record | writer／generic consumer `confirmed`；runtime `unknown` |

對 `assets_raw/CTY*.DAT` 的同一格式唯讀掃描共 89 檔／768 筆 NPC，完整 byte5 命中數為
`0x0d:26`、`0x11:1`、`0x1a:6`、`0x1b:25`、`0x24:4`；代表 raw record、檔案 hash、
section／offset 均在 sidecar `story_flags_static_deep_dive.npc_record_scan`。這項結果
把「generic consumer 確實有資料可讀」與「哪個劇情入口應設定它」分開：目前未找到五個
literal 的直接 `sub_16F09` caller 可安全命名，故不能把 handler 推論成 Go production
事件或把 runtime 欄升為完成。

### 2026-08-11 攻略對讀勘誤：`0x1a/0x1b` 不是龍之女王

攻略主線 step 43 的龍之女王／光之珠事件位於 `CTY67` handler52；step 44 才進入
`CTY65` 巴拉摩斯 boss room，step 45 返回阿里阿罕。原始 `CTY65.DAT` sec0 `(8,3)` 的
sub2 `byte4=70` 與固定編隊 `DS:0x4edf` 對上 `sub_1622A`，其戰後同時 SET `0x1a`／`0x1b`
並重建日夜／世界地圖。CTY00／CTY25 的 full-byte NPC records 及「打倒巴拉摩斯」等
文字，進一步閉合這兩旗標的 post-Baramos consumer。

本文件較早版本把 handler70 寫成「龍之女王中繼」是歷史誤判，原因是把 CTY65 其他 NPC
的 rec69–71 台詞套到 sub2 handler；該結論已推翻但保留於 Git 歷史與
`docs/data/special-events-audit.md` 勘誤段。現行等級為：原版 writer／generic consumer
`confirmed`，Go production writer／完整玩家可見流程 `unknown`。

### 4. 本輪對使用者問題的界線

- boss 正式 queue 與 alternate runner 均未找到同一 actor 在同一回合的 repeat-N 迴圈；
  靜態負證據已補強，但 boss 多次行動的實際條件／數量仍 `unknown`。
- SHP loader 以 monster id 讀單一 payload，`sub_1B31A` 依 header 尺寸消費一次，沒有找到
  action frame index；六次 `sub_1B3C3` 是同 payload 的遮罩重畫。逐動作 frame 仍無法由
  目前靜態資料命名。
- VOC 的 `BP → DS:0x253e → sub_22CF5(command 6) → sub_236B4 → sub_208E2`、兩個
  VCX bank 與 raw block descriptor 已 `confirmed`；中文 cue 名稱、PCM 解碼後波形與
  玩家可見停頓不是反組譯可單獨推出的 V3，仍列待對拍。

sidecar 已同步保存上述 raw bytes、來源 hash、IDA 位址與每筆推論等級；production
game-pack、boss repeat、逐動作動畫／音效與五個缺失事件仍遵守 fail-closed，不因本輪
靜態 consumer 閉合而宣稱 remake 完成。

## 十、E2 pack integration（2026-08-11）

使用者確認採用 E2／D2 快速完成標準後，已把本文件中可直接由 raw parity 驗證的最小戰鬥
契約複製到 `dq3_remake_ebitan/internal/gamepack/packs/dq3_cht/data/battle.json`：
`DS:0x4966` 區域表、`DS:0x4a56` 的 27×4 候選列、13 筆固定編隊、`DS:0x36e5` threshold、
`DS:0x37c3` descriptor 與 60 個 D3TXT00 中文 record 對映。loader 會檢查 raw 長度／SHA-256、
stride、區域索引、固定 record 與名稱 cardinality；production 建立 encounter decoder 時
只使用 pack raw，不再直接讀取 `DQ3.EXE` 的版本專屬遭遇表。

`EncounterTables.Slot` 同步修正原版 `dec region; shl 5` 的 off-by-one，raw region `1..27`
現在對應 Go slot `0..26`；region `0`／越界值 fail-closed。這是 E2/D2 資料與 loader 閉環，
不改變 boss repeat-N、逐動作 frame、PCM 播放／停頓、精確螢幕座標或五個缺失旗標的
`unknown`／V3 狀態。

### 2026-08-11 E2 headless closure

在一次性 Docker＋Xvfb 中以同一個 versioned pack 重播正式 `InputState`：pack encounter
decoder、巴拉摩斯事件戰鬥、薩曼奧薩鏡像事件、save/load identity、混合 formation、逐 actor
敵方回合與掉落／獎勵 component tests 均通過；`go test ./internal/...` 亦全綠。這是 E2／D2
的 engine/data／交易驗收，不是 DOSBox 或 V3 畫面／聲音驗收。boss repeat-N、SHP action frame、
PCM 波形／停頓、精確 formation 座標與五個 story flag 的 remake player consumer 仍保留
`unknown`／fail-closed。

完整 `TestOpeningProductionInputTrace` 另作非 E2 回歸觀察時，在 CTY7 入口後的 CTY7→CTY8
section route fixture 停在 `(7,6)`；這是**歷史／已訂正**的中途觀察。後續正式輸入修正已由
新遊戲重播至 `THE END`，現行 campaign 判定以 [`docs/74`](74-ebiten-remake-completion-plan.md)
最新 checkpoint 為準；保留本段只為說明當時為何不能提前宣稱 E3。

## 十一、2026-08-11 IDA 9.4 formation／action／PCM 補證與 E2 接線

本節是對前述 `unknown` 項目的追加勘誤，不覆寫時間序列。輸入仍是
`assets_raw/DQ3.EXE`（115282 bytes，SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）與同一份
`D3MNS.DAT`、`DQ3MNS.SHP`、`FVOC.VCX`、`NVOC.VCX`；IDA Pro 9.4.0.260610 在一次性
Docker 容器讀取 `/tmp/dq3-ida-static/DQ3.EXE.i64` 工作副本，保留 IDA linear、logical
及 primary code file 位址口徑。原始檔沒有被 patch、改名或覆寫。

### 1. formation raw 位置已閉合，並遷入 pack

IDA raw range `0x1aaa1..0x1aad4`（logical `0xaaa1..0xaad4`、file
`0xbe11..0xbe44`）與 `0x1aad5..0x1ab2b`（file `0xbe45..0xbe9b`）顯示：

```text
sub_1AAA1: word_272E3 = 0x26
           對每個群組扣除 D3MNS[monster].+0x28 × count
sub_1AAD5: word_272E3 += 2
           逐群組、逐個體呼叫 sub_1AB2C
sub_1AB2C: active +0x03 = word_272E3
           word_272E3 += 2 × D3MNS[monster].+0x28
sub_1B31A: EGA destination = 0x54 × (0x102 - sprite_height)
                              + active +0x03
```

`sub_1AB2C` 的 raw range `0x1ab2c..0x1ab7e`（file `0xbe9c..0xbeee`）也確認每個
active record 逐個體擲 HP、複製速度／攻防／咒文欄位，再遞增上述 raw 位置；這不是把
同組 HP 或橫向座標套成代表值。`battle.json` 現保存 `origin_raw=0x26`、
`first_member_offset_raw=2`、`weight_step_multiplier=2`、`ega_stride_bytes=0x54`、
`ega_bottom_raw=0x102`、`pixels_per_raw_byte=8`，`game/battle.go` 依 raw EGA destination
轉成 renderer 像素。這一段是 `confirmed` writer→consumer；從 EGA byte 到目前像素面的
轉換標為 `strong` 靜態投影，仍不宣稱同狀態 DOSBox V3 截圖。

### 2. boss repeat-N 與逐動作 SHP frame 仍沒有可實作的原版值

- 正式 `sub_1C08B`（IDA `0x1c08b`／logical `0xc08b`／file `0xd3fb`）對每筆
  `sub_1C34F` queue entry 只呼叫一次 `sub_1A973` 或 `sub_1B3F3`；下一回合才重建 queue。
- alternate `loc_190D7`（IDA `0x190d7`／logical `0x90d7`／file `0xa447`）雖最多重建
  `0x0c` 輪，每輪仍每個 entry 一次，沒有同一 actor 同回合的 repeat-N 外層迴圈。
- `sub_1B16F`／`sub_1B19E` 依 monster id 讀一份 SHP payload；`sub_1B3C3` 的六次
  `sub_1B220` 是同一 payload 的遮罩重畫與 `[2,1]` tick 等待，不是已證實的 action
  frame index。

因此 `boss_repeat_n`、boss 特殊條件、逐動作 SHP frame 與每幀可見停頓仍為 `unknown`。
不能用連戰 12 輪、D3MNS 未命名欄位或六次重畫猜成 `repeat=1`／`repeat=N`；E2 保持
fail-closed，這些未知不阻塞已完成的 raw formation integration。

### 3. PCM raw timing metadata 已可重建，但 host wall-clock 仍未知

`sub_20770 → DS:0x253e[(BP<<2)] → sub_22CF5(command 6) → sub_236B4 → sub_208E2`
的 cue／driver chain 與兩個 VCX bank 已 `confirmed`。decoder 現保留每個 type-1 block
的 file offset、payload length、rate byte、codec、source rate、source sample count
及 `source_duration_nanos = samples × 1e9 / floor(1000000/(256-rate_byte))`；本輪
觀測的 cue `1/4/6/8/9/16/21` 完整 metadata 見 sidecar 的
`battle.audio.raw_block_descriptors`。`NVOC` cue 21 的正確位址是 `0x1f973`（不是舊文件
誤寫的 `0x1f913`）。

`gaudio` 已接收 decoder duration sidecar，但 `PlaySFX` 仍是非阻塞的 fire-and-forget；
沒有把 44100 Hz 重取樣長度、Sound Blaster DMA 完成或玩家可見停頓冒稱為硬體精確 timing。
故 raw sample timing 為 `confirmed` 可重建資料，cue 語意／實際波形／host wall-clock 與
逐動作動畫同步仍是 `unknown`／V3 gap。

### 4. 本輪可宣稱的 gate

- `go test ./internal/... -count=1`：Docker 內通過，包含 pack formation schema、VOC
  source metadata 與原始 parity。
- `DQ3_ASSETS=/repo/assets_raw` 的 Xvfb targeted game tests：混合編隊、原始 weight
  stride、逐敵行動、全滅與獎勵通過；測試仍是 remake 內部 deterministic trace，不是 DOSBox
  oracle。
- 因此目前狀態是「formation raw E2／D2 已接線；PCM source metadata E2／D2 已保留；
  boss repeat-N、逐動作 frame、host PCM wall-clock 仍 unknown」，不可在 README／release
寫成完整戰鬥 V3 parity。

## 5. 2026-08-11 runtime 接線更新（現行 E2）

本文件前述「五個 story flag 的 Go production writer 仍 unknown」是接線前的時間序列；
本節記錄目前程式與 pack 的收斂結果。`events.json` schema 已升至 `0.1.29`，每筆保留
原始 handler／selector／set-clear flag／IDA source，engine 只執行固定 primitive：

| handler | 正式入口 | transaction | 驗收 |
|---|---|---|---|
| 74 `sub_16346` | CTY80 sec1 國王冊封 | clear `0xf5`、set `0x0d`、scene reload | `TestPostZomaEndingStoryFlagWriter` |
| 69 `sub_16139` | CTY70 六珠祭壇→復活動畫 | map cell `0x80`、transient set/clear `0x11`、停放 `(0x30,0xbd)`、save replay | `TestPhoenixRuntimeMapMutationAndSaveRoundTrip`、既有正式輸入 trace |
| 70 `sub_1622A` | CTY65 `(8,3)` 固定編隊勝利 | clear `0x29`、set `0x1a/0x1b`、day phase／step reset、palette/world rebuild | `TestBaramosAftermathStoryFlagWriter`、既有 Baramos trace |
| 33 `sub_15904` | CTY44 `(13,27)` 正式「話す」 | `DI=0x0bff` 換算 D3TXT record 71、clear `0x3d`、set `0x24/0x21`、scene reload | `TestSamanosaConditionalStoryFlagWriter` |

`TestStoryFlagRuntimePackContracts` 另鎖定四筆 JSON 的 D3／IDA evidence 與 handler raw。這
是 engine/data／save 的 E2 閉環，不是同狀態 DOSBox V3；原版完整後續 route、精確 palette
register、逐幀 SHP／PCM wall-clock 與 boss repeat-N 仍維持本文件前述 `unknown`，不可在
README 或 release 描述成全戰鬥 V3 parity。

## 十二、2026-08-12 未閉合項目的 IDA 9.4 收斂

本節不覆寫前述時間序列；它以新的、一次性 Docker IDA 9.4 audit 檢查目前仍可能影響
remake 決策的項目，將「可由靜態資料閉合」和「靜態分析無法誠實保證」分開。原始
`assets_raw/DQ3.EXE` 保持唯讀；工作副本、`.i64`、`.asm`、IDC 與輸出都在 `/tmp`，未加入
Git，工作結束後刪除。輸入仍是 115,282 bytes、SHA-256
`5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；位址均為 IDA DOS
loaded linear address。

### 1. queue／repeat-N 的窄範圍交叉檢查

IDA 的 direct CREF 結果為：

```text
sub_1C34F (queue builder) ← 0x18ea7, 0x190e3, 0x1c0ad
sub_1A973 (actor consumer) ← 0x190f4, 0x1c0de
sub_1B3F3 (alternate consumer) ← 0x1c0d8
```

`0x1c0c5..0x1c104` 以 `CX=DS:0x2515` 逐筆取 2-byte queue record、依 actor code 呼叫
一次 `sub_1A973` 或 `sub_1B3F3`，再 `SI+=2`；下一回合才回到 `sub_1C08B` 重建 queue。
`0x190e3..0x19110` 同樣逐筆呼叫 `sub_1A973`，一輪後才遞增 `DS:0x26fc`，最多 `0x0c` 輪。
第三個 queue-builder caller `0x18ea7` 只將已排序 queue 的結果寫入 `DS:0x35f2`；它沒有
direct CREF 到 actor consumer，不能被命名為 boss repeat handler。

這使「一般戰鬥每個 queue entry 一次」為 `confirmed`，也使「目前沒有找到資料驅動的同一
actor repeat-N writer→consumer」成為 `strong` 的負面靜態證據。它**不**證明任何未找到的
其他版本／自修改路徑永遠不存在；因此不能把未知欄位寫成 `boss_repeat_n=1`，也不應在
remake 加入猜測性的 boss 多動作機制。現有的一次 queue action 是原始正式路徑的已證實預設。

### 2. 其餘項目的停止條件

| 項目 | 靜態最終結果 | remake 決策 |
|---|---|---|
| formation raw position | `sub_1AAA1 → sub_1AAD5 → sub_1AB2C → sub_1B31A` 的 raw formula、EGA stride／bottom 及 pack integration 已 `confirmed`／E2 | 不再把「缺每 formation sidecar」列為 blocker；像素排列仍只屬 V3 視覺驗收。 |
| 抗性中文名稱 | effect id → D3TXT00 record 中文名稱與 packed class／threshold consumer 已 `confirmed` | 只保留 raw descriptor／class；不可把咒文名稱猜成元素或 `fire_resist` 等 production 欄位。 |
| `0x0d`／`0x11`／`0x1a`／`0x1b`／`0x24` | handler、SET/CLEAR、generic NPC consumer 與四筆有限 runtime transaction 已 E2 接線 | 早期「runtime unknown/missing」是歷史狀態，不重開已閉合 JSON／engine 切片；同狀態演出仍是 V3。 |
| 日夜 clock／palette bank | `0..0xef`、`0x78/0xf0`、12-byte bank index、DAC upload 已 `confirmed` | 不用 `DarkenPalette` 的合理值回寫原版；精確可見 RGB／transition 需同狀態 V3，不是再掃一次 EXE 可解。 |
| SHP 動作 frame | `sub_1B3C3` 固定六次 `sub_1B220`，tick `[2,1]`；未見 action/frame selector | 不新增 `animation_id` 或逐動作 frame JSON；這是 `unknown`，需要原版 frame trace 才能升級。 |
| PCM cue／timing | cue dispatch、VOC raw sample count／source duration、`sub_208E2` wait 已 `confirmed` | 不把 source duration 等同 host wall-clock、PCM 輸出波形或可見停頓；需錄音／frame trace。 |

### 3. 結論與下一個可證偽輸入

原版未閉合研究在「能由目前 EXE／DAT／VOC 靜態資料決定的範圍」已收斂：沒有尚待猜填的
production 設定欄位，也沒有理由再以廣泛反組譯重跑同一批資料。剩下的是觀測性問題，不能
由更長的 listing 變成真相：boss 特例的實際玩家路徑、逐 frame EGA/SHP 狀態、PCM 裝置輸出
與 wall-clock、日夜每個 palette transition。

因此後續只有取得可重播的原版同操作 frame／音訊 trace，或發現新的具體 caller／資料檔時，
才重新開啟對應的窄任務。沒有新輸入時，production JSON 維持 fail-closed，README／release
只可稱「靜態 E1/E2 與特定 checkpoint V3」，不可稱全戰鬥或全日夜 V3。
