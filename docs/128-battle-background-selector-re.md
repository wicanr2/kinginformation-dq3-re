# 128 — 固定編隊戰鬥背景 selector 的 IDA 閉環

> 日期：2026-08-12。狀態：日邦格八頭大蛇兩場固定編隊的背景 page／palette bank 已接入
> `dq3_cht` game pack，達 D2 靜態資料閉環與正式路徑 V2；不是全 terrain 背景系統或 V3
> 同狀態逐像素宣告。

## 範圍與輸入

本切片只處理原版固定編隊 record 的 header byte 1／2：它們如何進入原版背景 archive 與
palette selector，以及如何由共用 renderer 消費。沒有反推 generic 地形的完整 selector，也沒有
把洞窟／沙漠名稱寫進 Go。

| 輸入 | 大小 | SHA-256 |
|---|---:|---|
| `assets_raw/DQ3.EXE` | 115,282 | `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` |
| `assets_raw/PACKBG.SCR` | 3,738,880 | `4e226cd23c38bbbce9db974da559933a2e84669b3abb1559ab9755c85ed800d0` |
| `assets_raw/MNSBK.PAL` | 480 | `e2985503b3eb5256bcaf44f2f1b0a0b1ab124e66873c2de210e0d25ffd75110d` |

分析在一次性 Docker 容器內以 IDA Pro 9.4 進行；原檔唯讀，工作副本、`.i64`、listing 與
script 輸出均只存在 `/tmp`，結束後清除。以下 `seg000:xxxx` 是 **IDA segmented listing**
位址；`DS:xxxx` 是 DGROUP offset；`file` 是 DOS EXE 檔案位移，三者不混用。

## 原始資料與 writer → consumer

| 固定編隊 | 原始位置 | raw bytes | IDA writer／consumer | 推論等級 |
|---|---|---|---|---|
| 日邦格第一戰 | `file 0x1b010` = `DS:0x4ed0` | `01 23 05 4b 01` | `seg000:5CB6` 以 `lea si, ds:4ED0h` 呼叫固定 battle runner；`seg000:BF35` 複製 `raw[1]`→`DS:0x0d71`、`raw[2]`→`DS:0x0d73` | 已證實（confirmed） |
| 日邦格第二戰 | `file 0x1b015` = `DS:0x4ed5` | `01 1a 05 4b 01` | `seg000:5979` 以 `lea si, ds:4ED5h` 呼叫同一 runner，並在前面設定原版掉落抑制 bit | 已證實（confirmed） |

`seg000:BF35` 的原始指令是：

```text
seg000:BF5B  mov al, [si+2]
seg000:BF60  mov ds:0D73h, ax
seg000:BF63  mov al, [si+1]
seg000:BF66  mov ds:0D71h, ax
```

因此兩個 byte 的順序不是舊 Go 欄位名稱所暗示的「background／page」；在本版本資料契約中保留
為 `page_raw`、`palette_bank_raw`，並要求與 `raw_bytes_hex[1]`／`[2]` 逐 byte 相等。

`seg000:C688` 的原始 consumer 再將 `DS:0x0d73 * 0x30` 加到載入的 `MNSBK.PAL` buffer
`DS:0x3352`，存為 `DS:0x25d1`。480-byte palette 可恰好分為 10 個 `0x30`-byte bank，每個 bank
有 16 組 RGB triplet。當 `DS:0x4f2d != 0` 時，同一函式直接將 `DS:0x0d71` 放入 `DI` 並呼叫
`seg000:C6E5`。

`seg000:C6E5` 執行：

```text
mov ax, 3D80h
mul di
add dx, di
int 21h / AH=42h       ; seek = DI * 0x13d80
...
read 0x6e00 bytes
...
read 0xcf80 bytes
```

故 `PACKBG.SCR` 是 46 個 `0x13d80`-byte page，不是 132 個可直接以 `0x6e00` stride 取用的
畫面。第二段 `0xcf80 = 640/8 × 166 × 4`，是目前戰鬥 scene band 使用的 640×166、四 plane、
row-interleaved field；第一段 `0x6e00` 仍按原 loader 讀取，但本切片不替它臆測玩家可見語意。

`seg000:0030` 與 `seg000:C7D9` 都將 `DS:0x4f2d` 寫為 1，且固定 runner
`seg000:BE89 → seg000:BFD1 → seg000:C688` 與兩個 handler 的 call chain 已由 IDA 直接列出。
「兩個日邦格事件在原版實際走 `DS:0x0d71` 分支」由這個靜態鏈、raw page 解碼結果及連續原版
影片畫格共同支持，標為強推論（strong）；沒有把 `0x4f2d` 的全域模式語意泛化到所有 encounter。

## 落地的資料契約

`data/battle.json.background` 現保存 archive asset key、page stride、field offset／長度、surface
幾何、page／bank 上限及既有草地 baseline；loader 驗證 asset hash、幾何乘積與 selector 範圍。
事件固定編隊則只帶資料：

```json
"background": { "page_raw": 35, "palette_bank_raw": 5 }
```

共同 Go decoder 接收 pack 提供的 `PackBGFormat`；renderer 依同一 selector 取 indexed pixels，並在
畫色時加上 pack 的 palette bank offset。任何缺 asset、page、bank、stride 或不一致 raw header
都會 fail-closed；沒有 DQ3 特例、`page22` 常數或「洞窟／沙漠」文字 fallback。

## V2 驗收

Docker＋Xvfb 以正式 `InputState` 從標題重播到 `THE END` 通過（67.24 秒）。設定
`DQ3_PRODUCTION_DUMP_OROCHI` 後，只有正常路徑抵達兩戰時才寫出 settled PNG：

| phase | SHA-256 | 結果 |
|---|---|---|
| `orochi_first_command` | `401edfda8ac55eab64466a01fd9e102635e17ab27305edbbdcfe0654c7e4b140` | 洞窟 page 35、bank 5 |
| `orochi_first_message` | `a5423245f902218d3c9a47b2228ebb33324fdad886c09f3f50a31f96860195a1` | 洞窟 page 35、bank 5 |
| `orochi_first_end` | `4ca5d74c298be75ff09219a4c2f4d288e4a87ed1ef8e759c9605a704c4915aa9` | 洞窟 page 35、bank 5 |
| `orochi_second_command` | `e3c604904e2b62431a9a98b9a7709e9f2ac605c6f938a7c4e6700ddd5ec70618` | 沙漠 page 26、bank 5 |
| `orochi_second_message` | `5cb82634be859c7e51fc94b2d93700221c431c4c75095a658db5dfc76b0eee66` | 沙漠 page 26、bank 5 |
| `orochi_second_end` | `ebbf7e2e19a21175c5278306e41b742c742841bd43f6971e313a3b4a0227d990` | 沙漠 page 26、bank 5 |

原版連續影片畫格 `f000745`（第一戰仍為洞窟）與 `f000768`（第二戰沙漠）可確認玩家可見
背景種類，因此此兩場達 V2。隊伍數值、按鍵節奏與動態 phase 並非同一份原版錄製，不能升格為
V3；完整前後差異保留在 [`docs/127`](127-orochi-v2-production-compare.md)。

## 第一性原理：要不要做完整背景系統？

| 問題 | 結論 | 理由 |
|---|---|---|
| 主線是否因這件事不可完成？ | 否 | campaign E3 的事件、掉落、旗標、轉場與 save/load 在背景錯誤前已可到 `THE END`。 |
| faithful remake 是否需要這個窄修正？ | 是 | 兩場必經 Boss 的背景從天空／草地變為明確的洞窟／沙漠，是高顯著、玩家立即可見的設定錯配。 |
| 現在是否要逆完所有 terrain→page 變化？ | 否 | 那是另一條 generic selector 資料流；不改變這兩場 fixed formation 的 correct output，也沒有足夠證據可安全填 JSON。 |

因此本輪只完成最小充分 primitive：**原始固定編隊 selector → validated archive/palette →
正式 renderer**。generic 地形背景目前保留已存在、已由草地 reference 支持的 `0/0` baseline；它
不是「所有地形已還原」的聲明。若未來取得新的 player-visible discrepancy，才以該場景的
writer → selector → archive → renderer 鏈另開窄任務。

2026-08-12 後續盤點已將相同 raw contract 接到怪力魔、巴拉摩斯與索瑪三連戰的單頭目入口；
它們沒有新增 generic terrain 推論，且缺同狀態原版畫格，僅標 D2／runtime V1。詳見
[`docs/129`](129-required-boss-backgrounds-re.md)。
