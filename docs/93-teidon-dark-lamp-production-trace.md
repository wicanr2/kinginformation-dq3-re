# 93 — 提頓《黑暗之燈》原版派發與 production trace

> 日期：2026-07-30
> 範圍：攻略步驟 22；由達瑪轉職合法 checkpoint 前往 CTY20，取得並使用
> 《黑暗之燈》，保存後繼續抵達 CTY19。
> 結論：設定資料 D3、玩家流程 E3、現行 runtime 畫面 V2；尚未宣稱原版同狀態 V3。

## 1. 先前「機制已有」仍不算完成的原因

舊 Go 實作已能把 item `0x5f` 轉成夜晚，但這個關係來自
`internal/itemuse.KindOf` 的 DQ3 專屬常數與 `switch`，不是原版 dispatcher。CTY20
寶箱也仍在 `game/treasure.go` 的 baked table。更重要的是，舊實作允許在城內使用後
重載夜間 NPC；IDA 證明原版 handler 只接受地表。

本批因此不重寫日夜 renderer，而是閉合兩條資料鏈：

```text
CTY20 sec1 event → 通用寶箱 consumer → item 0x5f / flag 0x39
道具選單 record(item+1) → item-use pointer table → handler → 晝夜 state/clock → palette
```

## 2. 寶箱原始資料

`CTY20.DAT` section 1 的 subid 0 entry 是：

```text
01 5f 00 39
```

依原版 section-event 格式：

- type `1`：給道具；
- item `0x005f`：《黑暗之燈》；
- present flag `0x39`：set 表示可取、成功後 clear；
- event tile 位於武器店二樓 `(4,9)`。

原版通用 dispatcher `file 0x9cf6..0x9e41` 由 subid 取 entry；type 1/3 consumer
`file 0x9de9` 將道具寫入角色八格物品欄，成功後清除 present flag。JSON parity test
直接重讀 CTY20 raw entry，不以舊 Go table 作 oracle。

## 3. IDA Pro 9.4：item `0x5f` 如何抵達夜晚 handler

分析在一次性 Docker 容器內使用 IDA Pro 9.4 完成。IDA EA 以載入 image base
`0x10000` 表示；本文換回專案 canonical logical/file 口徑。

### 3.1 選中道具到 pointer table

道具使用 dispatcher：

- `logical 0x3ccf`（`file 0x503f`）讀已選 record，先 `dec bx` 得 raw item ID。
- raw ID `<0x41` 走特殊裝備清單；`>=0x41` 於
  `logical 0x3d00..0x3d12` 減 `0x41`、乘 2，再呼叫
  `word ptr [bx+DGROUP 0x366a]`。
- `0x5f-0x41=0x1e`，因此讀 `DGROUP 0x36a6`。
- `DGROUP 0x36a6`（`file 0x197e6`）原始 word 是 `0x4063`，即真正 handler
  `logical 0x4063`（`file 0x53d3`）。

這條鏈也解釋為何單純搜尋 `cmp 0x5f` 會誤判：IDA 找到的唯一立即值比較位於
`logical 0x6ffa` 的隨機數 helper，與道具派發無關。

### 3.2 handler gate、writer 與 consumer

`logical 0x4063..0x40aa`：

1. 比較 `DGROUP 0x4f2d == 0`；不是地表便跳失敗路徑。
2. 比較 `DGROUP 0x526c == 1`；只有日間狀態進入 transaction，已是夜晚不重寫。
3. 寫 `DGROUP 0x526c = 0`。
4. 將晝夜 clock `DGROUP 0x251d` 每次加 `0x0a`，呼叫 palette/time consumer，
   直到大於 `0x8c`；最後固定寫成 `0x008c` 並再刷新。
5. handler 沒有清除角色 inventory slot，因此《黑暗之燈》可重複使用。

步數 consumer `logical 0xee23`（`file 0x10193`）每個合法地表步增加
`DGROUP 0x251d`；clock 到 `0x78` 時寫 `[0x526c]=0`，到 `0xf0` 時 clock 歸零並寫
`[0x526c]=1`。這使 `0x526c=1/0` 的日／夜語意及黑暗燈的可見 palette 副作用閉合。

現行引擎仍以四個 renderer phase 表示白天／黃昏／黑夜／黎明；pack 的
`day_night_phase: 2` 是 canonical「黑夜」值。原版 `0x008c` clock 與四相位 palette
逐色對拍仍屬 P7 V3 工作，不能反過來把現行近似 palette 稱為原版數值。

## 4. game-pack 與引擎邊界

schema `0.1.8` 新增有限的 `item_use_effects`：

- JSON 保存 raw item `0x5f`、地表 gate、具名 effect
  `force_day_night_phase`、目標 phase、step reset、是否消耗及 D3 evidence。
- Go 只實作具名有限 primitive，不知道《黑暗之燈》或 raw `0x5f`。
- 未知 effect／錯誤 location fail closed；城內使用不改 phase、不重載 NPC、不消耗。
- CTY20 treasure 同時遷入 `treasure_events`，並從 `game/treasure.go` 移除。
- 原始 EXE／CTY decoder 與 parity test 保留，不以 JSON 自我驗證。

## 5. 正式玩家 trace

`TestOpeningProductionInputTrace` 從既有 boot trace 繼續：

1. 正式讀取達瑪賢者轉職存檔。
2. 因 save/load 會重建關閉 tile，再用魔法鑰匙與「調查」開 CTY17 魔法門。
3. 正常出城、走到停泊船、登船、航行並靠岸進 CTY20。
4. 航程抵達非白天相位時，先正常出村，在地表以方向鍵、晝夜步數及正式遭遇等待白天，
   再重新進村；沒有直接修改 phase。
5. 經原始 transition 到 CTY20 sec1，從命令窗選「調查」取得 `0x5f`，驗
   flag `0x39` clear。
6. 回 sec0 正常出村，由 D3TXT00 rec421 選「使用」；驗 phase=黑夜、step reset、
   item 保留。
7. save/load round-trip 後仍在地表夜晚、仍持 `0x5f`、flag `0x39` 仍 clear。
8. 再次正式登船、航行並進入下一節點 CTY19，證明交易後 campaign 可繼續。

## 6. Runtime 圖與下一個 blocker

- [`img/teidon_dark_lamp_obtained.png`](img/teidon_dark_lamp_obtained.png)：
  CTY20 sec1 開箱後的現行 Ebitengine runtime。
- [`img/teidon_dark_lamp_night.png`](img/teidon_dark_lamp_night.png)：
  地表由 pack-driven effect 進入黑夜並關閉道具清單後的 runtime。

下一個連續 audit 從 CTY19 正式入口閉合八頭大蛇洞窟 B2F 第一戰、草薙大劍
`0x14`、日邦格傳送後第二戰與紫寶珠 `0x69`。既有孤立 boss handler 只能作線索，
不得代替這條 boot 起的 production trace。
