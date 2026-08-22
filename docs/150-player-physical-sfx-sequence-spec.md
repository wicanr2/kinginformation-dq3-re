# 玩家一般物理攻擊 VOC 序列契約

> 狀態：2026-08-22。原版 call-site／順序為 `confirmed`；remake 已達 D3／E2。
> 本切片只接玩家一般物理攻擊正常分支的 cue 6 → cue 11；敵人攻擊、共同傷害、
> critical／特殊 effect tail 仍分開保留，不以相同 cue 數值合併語意。

## 問題與停止線

`docs/149` 已接玩家／敵人逃跑 cue，但現行玩家普通攻擊只顯示傷害結果，沒有原版
`record 0x14a` 的攻擊訊息，也沒有兩段 VOC completion wait。直接把 cue 6 全域命名為
「攻擊音效」仍是錯的：`sub_1B7B0` 的 effect tail 也使用 cue 6，其上游條件與玩家可見
語意尚未閉合。本規格只為 `sub_1B4F6:0x1b568` 的正常分支建立穩定角色。

敵人入口 `sub_1AC05` 的 cue 4、共同傷害 `sub_1ACCE` 的 cue 1／4／9 雖有 raw call-site，
但其 mode 分支、message record、前一 cue completion 與死亡／狀態副作用尚未完成有限
語意分割；本輪不把它們寫入 production JSON。

## 輸入、工具與非破壞性輸出

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `assets_raw/FVOC.VCX`：72,531 bytes；SHA-256
  `47be27d2244e314ca6f2a21d3fa61a95d4e823f9085ed3d21808ac25d66b39c5`。
- `assets_raw/NVOC.VCX`：143,156 bytes；SHA-256
  `53c710f356b728e97287bfb64af7f64cf2067b28f83a9dfe32f99a6eb61f4156`。
- Docker 內 IDA Pro 9.4／IDAPython；匯出器 `tools/ida_dump_physical_sfx.py`。
- gitignored sidecar：`work/physical-sfx-20260822.json`。
- 主段換算：`logical = IDA linear - 0x10000`；`file = logical + 0x1370`。
- 匯出保留原始 `sub_*`、bytes、xref、CFG 與每筆推論等級；Hex-Rays 在此 16-bit
  input 無可用 pseudocode，sidecar 明列 `unknown`，不以攤平文字取代 IDA CFG。

## 原版垂直鏈

玩家 action entry `sub_1B3F3` 讀 party action type `1`，在 IDA `0x1b417` 呼叫
`sub_1B4F6`。死亡／睡眠 gate 之後，正常物理分支為：

```text
IDA 0x1b568  mov bp,6
IDA 0x1b56b  call sub_20770
IDA 0x1b570  mov di,0x14a
IDA 0x1b573  call sub_21414       ; 顯示「{角色}的攻擊!」
IDA 0x1b578  call sub_208E2       ; 等 cue 6 完成
IDA 0x1b57d  mov bp,0x0b
IDA 0x1b580  call sub_20770
IDA 0x1b585  call sub_208E2       ; 等 cue 11 完成
IDA 0x1b58a  call sub_19834
IDA 0x1b5a2  call sub_1ACCE       ; 共同傷害／miss／death 分支
```

換算範圍為 logical `0xb568..0xb5a2`、file `0xc8d8..0xc912`。`sub_1B3F3` 對
`sub_1B4F6` 的唯一直接 caller、`sub_1B4F6` 對 `sub_1ACCE` 的 caller 與兩個
`sub_208E2` consumer 都存在於 IDA database；因此 call-site、順序與 completion wait
均為 `confirmed`。

`D3TXT00.TXT` record 330／`0x14a` 的 raw glyph stream 是
`[0xfffb,149,623,624,57]`，玩家可見值為「{角色}的攻擊!」。它與 record 331／`0x14b`
文字相同但插值控制碼不同；玩家路徑必須保存 record 330，不可用敵人 record 331 代替。

## 原始 VOC descriptor

現行 desktop assets 使用 `FVOC.VCX`；其 raw descriptor 為：

| cue | file offset | block length | rate byte | source rate | samples | source duration |
|---:|---:|---:|---:|---:|---:|---:|
| 6 | `0x215e` | 1224 | 107 | 6711 Hz | 1222 | 182,089,107 ns |
| 11 | `0x7ad9` | 1167 | 107 | 6711 Hz | 1165 | 173,595,589 ns |

這些 descriptor／source duration 為 `confirmed`。`NVOC.VCX` 另有同 raw cue index，
但原版 bank selector 的玩家性別／模式語意尚未閉合；實際波形名稱、Sound Blaster DMA
wall-clock 與 44.1 kHz host 重取樣結果仍為 `unknown`，故不宣稱 V3。

## Remake 資料與狀態機契約

- `battle.json.sound_cues.player_physical` 保存有序 `steps`：cue 6、cue 11；兩步皆
  `wait_for_completion=true` 並各自帶同一原版鏈的 D3 evidence。
- Go engine 只知道「訊息綁定一個有序音效序列」。raw cue、數量與順序不得寫成 DQ3
  production 常數。
- 玩家 leader 與 companion 的正常 physical action 都先排入 record 330 的
  `actor_attack` 訊息，再排既有 damage／miss 結果。
- 訊息真正顯示時才播放 step 0；duration 到期後自動播放下一 step。所有 step 結束前
  Confirm 都被忽略，且同一 step 不可重播。
- headless 無音訊後端只保存 duration、不配置 PCM；仍須得到相同序列與等待幀數。

## 驗收

- IDA sidecar：輸入雜湊、IDA 9.4、原始位址、bytes、CFG、xref、推論等級皆存在。
- pack：schema／reference validation，cue 6／11、順序、wait 與 EXE／FVOC raw parity。
- component：fake SFX 驗證先播放 6、等待 11 幀，再播放 11、等待 11 幀；完成前 Confirm
  無效，完成後才可進 damage message。
- player path：leader／companion 都經正式 command action 排出 attack → damage；完整
  production trace 不退步。
- 當時停止線：本切片不接 cue 1／4／9、敵人 attack、critical effect tail、NVOC selector
  或 V3 wall-clock。cue1／4 與敵方 attack 後續已由 `docs/151` 獨立閉合；cue9、critical、
  NVOC selector 與 V3 wall-clock 仍維持停止線。

## 實作與回歸結果

- 此切片初次落地時為 game-pack schema `0.1.39`／DQ3 content `0.1.44`；後續
  `docs/151` 切片當時升為 schema `0.1.40`／content `0.1.45`。工作樹現行版本以根 README
  與 manifest 為準；`player_physical.steps` 仍是唯一
  cue 6／11 production owner。`texts.json` 的 record330 consumer 也由錯誤的 enemy 標註
  訂正為本玩家正常物理分支。
- 原始 `DQ3.EXE` 三個 instruction anchor、D3TXT00 record330 raw glyph stream，以及
  `FVOC.VCX` cue6／11 descriptor parity tests 通過。
- fake SFX component 證實 cue6 wait 11 frame，才播放 cue11 並再 wait 11 frame；兩段完成前
  Confirm 不前進。leader／companion action queue 都是 `actor_attack → actor_damage`。
- Docker＋Xvfb 完整 `game` 回歸通過（151.152 秒）；正式新遊戲 `InputState` trace 另行由
  標題重播至 `THE END` 通過（107.303 秒）。完整 `internal/...` 與隔離 desktop build 亦通過。
- 這些結果只證明有限 E2 與現行 campaign E3 未退步；不改變本文的 V3／unknown 停止線。
