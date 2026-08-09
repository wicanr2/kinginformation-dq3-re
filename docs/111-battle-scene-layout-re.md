# 戰鬥場景帶與游標字模：非破壞性資料契約

> 2026-08-10；本文件只附加原版定位與推論等級，不改寫 `DQ3.EXE` 或 IDA
> database。這一批把已可回查的場景帶、怪物基線、游標字模與共用 frame 色彩契約
> 移入 game pack；逐動作動畫／停頓／音效仍是未閉合的 V3 工作。

## 輸入與證據

| 欄位 | 值 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`（SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`）、`assets_raw/D3TXT00.FON`（SHA-256 `c19e1ca03c6c15916d934f3338ac4215290a5fc3d0d8e57c6976226241e40b02`） |
| 主要工具／位址基準 | IDA Pro `9.4.0.260610`（Docker 內唯讀；`/home/anr2/ida_94_official/dist`）；DQ3.EXE DOS loaded linear／file 與 D3TXT00.FON glyph index 分開記錄 |
| 對應文件 | [`docs/13-exe-battle.md`](13-exe-battle.md)、[`docs/02-font-format.md`](02-font-format.md) |
| 推論等級 | `D2`：靜態 renderer 座標／字模索引與玩家可見戰鬥畫面相符；尚無逐欄位 writer sidecar 可升為 D3 |

## pack 欄位

`interface.json.battle_scene` 現在保存：

| 欄位 | 值 | 語意 | 證據／限制 |
|---|---:|---|---|
| `field_y0` | `80` | 戰鬥天空／草地場景帶起點 | `docs/13` 的原版戰鬥 renderer 座標系；D2 |
| `field_y1` | `246` | 場景帶結束位置 | 同上；D2 |
| `ground_y` | `232` | 怪物站立基線與 fallback 草地分界 | 同上；D2 |
| `cursor_glyph` | `0x77`（119） | 目標／命令選擇游標字模 | `D3TXT00.FON` 字模表將 `0x77` 定為 `★`；D2 |

`NewGameWithPack` 缺少 `battle_scene` 會拒絕啟動。`battle.go` 不再保留這四個
DQ3 專屬值；直接 `Battle` fixture 若未安裝 layout 只停止繪製，不會自行猜測座標。

## 未閉合項目

- `0x031a`、`0x0b11`、`0x0912` 是 `win_rect` 第一 word 的 raw flags／備份槽，不是
  frame style id；共用 `sub_1fd30/sub_1fdb1` 的可見結果已以 `frame` JSON 契約保存
  為連續 1px 白框／黑色內部（`strong`，完整逐 input trace 仍未逐 modal 閉合）。
- spell／target 其他文字基線、數字基線、逐動作動畫、停頓與 battle SFX cue 尚未由
  `入口 → writer → consumer → 玩家可見效果` 閉合。
- 因此本欄位資料化是 engine/data 邊界切片，不等同完整戰鬥 parity 或 release 完成。
