# 戰鬥場景帶與游標字模：非破壞性資料契約

> 2026-08-09；本文件只附加原版定位與推論等級，不改寫 `DQ3.EXE` 或 IDA
> database。這一批只把已可回查的場景帶、怪物基線與游標字模移入 game pack；框線
> pattern、逐動作動畫／停頓／音效仍是未閉合的 V3 工作。

## 輸入與證據

| 欄位 | 值 |
|---|---|
| 原始輸入 | `assets_raw/DQ3.EXE`、`assets_raw/D3TXT00.FON` |
| 主要工具 | IDA Pro 9.4（Docker 內唯讀；`/home/anr2/ida_94_official/dist`）與原版戰鬥座標文件 |
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

- `sub_1f590` 使用的 frame style id（目前已知 raw `0x031a`、`0x0b11`、`0x0912`）
  與 EGA pattern 尚未建立可重生 sidecar，不能把目前白色單線框宣稱為逐像素 V3。
- spell／target 其他文字基線、數字基線、逐動作動畫、停頓與 battle SFX cue 尚未由
  `入口 → writer → consumer → 玩家可見效果` 閉合。
- 因此本欄位資料化是 engine/data 邊界切片，不等同完整戰鬥 parity 或 release 完成。
