# 115 — CTY 遭遇旗標與正式步數計數器：IDA 9.4 勘誤與接線

> 2026-08-10。這份文件只修正 section `+0x11` 的 polarity，並記錄它如何進入 Go
> runtime；不把仍未閉合的強制遭遇、完整 encounter pack JSON 或戰鬥 V3 誤宣稱完成。

## 1. 問題與輸入

先前 `docs/13`、`docs/34`、`docs/35` 把 `DS:[0xd77]` 的極性寫反，導致 Go 只在地表
遇敵，CTY section 的原始值也沒有保留。以原始 `assets_raw/DQ3.EXE`（115,282 bytes，
SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）的唯讀副本，
在一次性 Docker 內使用 IDA Pro `9.4.0.260610` 重建 listing。IDA linear address、logical
與 file offset 分開記錄：

| 證據 | IDA linear | logical | file | 結論 |
|---|---:|---:|---:|---|
| `sub_130CF` 讀 section header `+0x10/+0x11` | `0x130cf..0x131b8` | `0x30cf..0x31b8` | `0x443f..0x4528` | `al=[di]`、`ah=[di+1]`；`DS:[0xd77]=ah` |
| `mov ds:0D77h, ah` | `0x131b2` | `0x31b2` | `0x4522` | `+0x11` 是 raw 遭遇 gate |
| `sub_1BD97` CTY gate | `0x1bd97..0x1bdde` | `0xbd97..0xbdde` | `0xd107..0xd14e` | `4F2D==1` 時比較 `0xd77` |
| `cmp [0xd77],0; jz return` | `0x1bda1` | `0xbda1` | `0xd111` | **0 是安全；非 0 才繼續** |
| `dec [0x52f4]` | `0x1bda8` | `0xbda8` | `0xd118` | 每個 eligible move 遞減 |
| 初始計數 writer `sub_1E6A3` | `0x1e6a3` | `0xe6a3` | `0xf9d3` | `rng(0x10)+4`，範圍 4..19 |

這是 `writer → state → consumer` 的閉合證據；CTY 名稱本身不是 polarity 證據。
例如 CTY00 sec0 的 raw `+0x11=0` 與原版城鎮無隨機遇敵相符；parser 仍保留每一 section
的原始 byte，不把它壓成由名稱推導的 bool。

## 2. Go 對映

- `internal/dq3data.Town.EncounterFlag` 保存 CTY header `+0x11`。
- `game.Scene.encounterFlag` 在每次 section load／transition／save-load rebuild 時保留 raw。
- `Game.encounterStep` 對映 `DS:[0x52f4]`，save JSON 以 optional `encounter_step` 保存，
  舊檔缺欄位時採第一次 eligible move 建立計數。
- `Game.advanceEncounterStep` 只由正式玩家移動後呼叫：地表一律 eligible，CTY 僅
  `encounterFlag != 0`；安全 CTY 與飛行不推進。
- 計數歸零後一般遭遇重擲 `rng(0x12)` 至少 10；非白天相位沿用原版 raw `[0x526c]=0`
  的 `-2` 節奏，再交給既有原始 region／候選怪／群量 loader。聖水仍 fail-safe 地
  遞減而不開啟遭遇。

## 3. 驗證

- `internal/dq3data` 測試確認 CTY00、CTY14 sec0/sec1 的 `+0x11` raw byte 未被 parser
  改寫或遺失。
- `game` component test 確認 `0` 安全、非零 CTY 可遇敵、地表不受 CTY flag、飛行禁遇敵，
  並鎖定初始 4..19 與日／非日 10..17／8..15 的計數範圍。
- 未把完整主線長測重新跑一遍；本輪依使用者要求只做受影響的針對性驗證。既有
  `TestOpeningProductionInputTrace` 的 2026-08-09 通過結果仍是 campaign E3 證據，
  不因本文件把所有畫面／音效提升為 V3。

## 4. 保留的不確定性

原版 `0x4f46 & 0x1000` 的強制遭遇短計數，以及完整 `encounters.json` game-pack 遷移，
仍是後續切片；本輪沒有用假資料填入。戰鬥框線、動畫、訊息 timing、SFX 與同狀態 V3 對拍
也仍依 `docs/74` 保持未完成狀態。
