# 戰鬥單次行動佇列規格（Boss repeat-N 負證據閉合）

> 狀態：2026-08-22，D3／E2。本文只閉合精訊版 `DQ3.EXE` 的正式戰鬥
> action queue 結構；不把攻略、其他 DQ3 版本或舊 C remake 的 Boss 行為套進來。

## 問題與先前矛盾

`docs/123` 已把 `sub_1C34F` 的 queue 與 `sub_1C08B` 的 consumer 寫成「每個活躍
actor 每回合一次」，但後段仍把 Boss repeat-N 寫成 `unknown`。兩句不能同時作為現行
規格：若原版另有 Boss repeat caller、函式指標或資料欄位 consumer，就必須找出；若沒有，
remake 也不應保留一個暗示待補的虛構功能。

## 輸入、工具與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/D3MNS.DAT`：5,330 bytes；SHA-256
  `48bf3ce78c425239363761c13c708a6e04f8daa628cd575f7e049bc0bc7bb4ed`。
  這是專案現行 130-record 資產，末兩筆是外部回補；本結論只由 `DQ3.EXE` control flow
  得出，不能拿回補 record 反推原版機制。
- 工具：Docker 內 IDA Pro 9.4，`tools/ida_dump_battle_action_queue.py`。
- 非破壞性 sidecar：`work/boss-repeat-queue-20260822.json`（gitignored）。腳本不 rename、
  不寫 comment、不修改 database 或輸入檔；保留原始函式名、bytes、xref 與推論等級。
- 位址：本文的 `IDA` 是 linear；`logical = IDA - 0x10000`；主段
  `file = logical + 0x1370`。

## 已證實的 caller → writer → queue → consumer

### 正式玩家戰鬥

1. `sub_1BDDF`（IDA `0x1bddf`）於 `0x1be28` 直接呼叫 `sub_1C08B`。
2. `sub_1C08B`（IDA `0x1c08b`）於 `0x1c0ad` 每回合呼叫一次
   `sub_1C34F`。
3. `sub_1C34F`（IDA `0x1c34f`）先把 `DGROUP:0x2515` 清零：
   - 玩家 loop 對每個存活 party record 只在 `0x1c38c` 寫一筆 2-byte queue record，
     並在 `0x1c3a8` 將 count 加一；
   - 敵人 loop 對每個存活 active enemy record 只在 `0x1c3c8` 寫一筆，並在
     `0x1c3ed` 將 count 加一；
   - `0x1c3f7..0x1c424` 只排序既有 2-byte records，不新增或複製 actor。
4. `sub_1C08B` 以同一個 `DGROUP:0x2515` 作 `CX`：每筆先在 `0x1c0ca` 讀 actor code，
   玩家分支於 `0x1c0d8` 呼叫一次 `sub_1B3F3`，敵人分支於 `0x1c0de` 呼叫一次
   `sub_1A973`；`0x1c0e6..0x1c104` 只將 `SI += 2` 並以 `loop` 進下一筆。
5. `sub_1A973` 與 `sub_1B3F3` 都在完成或跳過一個 actor action 後直接 `retn`；兩者內部
   沒有依 monster id／Boss flag 回跳自身的 repeat loop。

因此，正式回合的 cardinality 是：

```text
queue entries = 本回合建立時存活的 party actors + 存活的 active enemies
每一 queue entry 最多呼叫一個 actor consumer 一次
```

角色在排隊後死亡、睡眠、逃跑或戰鬥已結束可以讓該次 action 被跳過，但不會產生額外 action。

### alternate 12-round runner

IDA 未把 `0x190d7` 所在區段建成函式，因此 sidecar 額外保留
`0x19074..0x19130` raw range，不替它虛構函式名。它在 `0x190e3` 呼叫
`sub_1C34F`，以 `CX=DGROUP:0x2515` 逐筆於 `0x190f4` 呼叫一次 `sub_1A973`；
`0x19105..0x19110` 增加 `DGROUP:0x26fc` 並最多重建 12 個**回合**。這是回合 repeat，
不是同一 actor 在同一回合 repeat-N。

### 第三個 queue-builder caller

`sub_18E88` 於 `0x18ea7` 呼叫 `sub_1C34F`，但 `0x18ec0..0x18ed7` 只逐筆把 queue
actor code 對映到 `DGROUP:0x35f2`，沒有呼叫任一 actor consumer。它不構成第三條戰鬥
action runner。

### consumer 完整性檢查

- IDA 對 `sub_1A973` 的全部直接 code xref 只有 `0x190f4`、`0x1c0de`。
- IDA 對 `sub_1B3F3` 的全部直接 code xref 只有 `0x1c0d8`。
- 原始 EXE 搜尋兩個 logical offset 的 little-endian word 均為零命中；沒有可疑的靜態
  函式指標 table。這不能一般化證明「程式沒有任何間接呼叫」，但與已閉合的 queue runner、
  caller graph 及 actor consumer 邊界共同構成足夠的 D3 負證據。

## 結論與 remake 契約

- **已證實（confirmed）：**精訊版正式戰鬥以及已知 alternate runner 都是每個 queue
  entry 每回合最多一次 action；Boss 沒有另一個 repeat-N 路徑。
- **已推翻：**「Boss repeat-N 的條件／數量仍待從 D3MNS 未命名欄位補入」是過期斷言。
  造成它的原因是早期把「尚未找到」保守寫成 `unknown`，後來雖已閉合所有 consumer caller，
  卻沒有同步把負證據提升為現行結論。
- **禁止實作：**不得新增 `boss_repeat_n`、Boss ID 特例、未證實的 D3MNS repeat 欄位或
  同 actor 重複排隊。未來只有另一份明確版本 binary 或新的原版動態證據推翻本結論時，才以
  獨立版本契約重開。
- Go 引擎應讓 `resolveRound` 明確由「每個存活 actor 各建立一筆」的 helper 產生 queue；
  測試必須檢查混合 party／敵人、死亡成員排除及 actor identity 不重複。這是具名狀態機
  behavior，不是可調的 DQ3 數值，因此不新增 JSON 欄位。

## 驗收與界線

- component：正式 `resolveRound` 使用的 queue builder 對每個存活 actor 恰好一筆；死者零筆。
- existing gameplay：多敵 `enemyTurn` 仍逐隻一次；完整正式 campaign trace 不退步。
- 等級：control flow 與 remake queue cardinality 為 D3／E2；本題不需要 DOSBox 動畫對拍。
- 不在本規格內：逐動作 SHP frame、PCM 人耳時序、同狀態戰鬥畫面 V3。
