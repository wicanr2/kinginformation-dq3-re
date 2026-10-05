# 野外輔助咒文 rec166–171 runtime 規格

狀態：2026-08-23，D3／E2。本文只閉合原版 field caster 合法但先前尚未接線的
`rec166–171`；`rec161–165` 治療與 `rec172–179` 工具咒文見 `docs/78-field-spell-re.md`。

## 輸入與工具

- `DQ3.EXE`：115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4，16-bit DOS linear address；原檔唯讀掛載。
- 可重生 sidecar：`tools/ida_dump_field_support.py`。IDA database 與執行輸出不入 Git。

## 已證實的閉環

| record | descriptor | handler | 已證實效果 |
|---:|---|---:|---|
| 166 | `03 00 0b` | `0x1cca8` | `sub_1469F(DX=0x40)`：存活單體清除 poison bit。 |
| 167 | `06 00 0b` | `0x1ccb2` | `sub_1469F(DX=0x10)`：存活單體清除 persistent paralysis bit。 |
| 168 | `03 00 11` | `0x1ccbc` | 全隊清除 `+0x38 bit0x20`；該位元是 battle sleep，不跨戰鬥持久化，因此正常 field 路徑只完成 MP 交易。 |
| 169 | `0a 00 0b` | `0x1cd04` → `0x1ceae` | 死亡單體；`rng(100) >= 50` 成功，成功回滿 HP。 |
| 170 | `14 00 0b` | `0x1cd0e` → `0x1ceae` | 死亡單體；跳過亂數，必定成功，回復 `maxHP >> 1`。 |
| 171 | `12 00 12` | `0x1cd18` | 找第一件 equipped cursed item，清除裝備標記並保留物品。 |

推論等級均為 `confirmed`：descriptor、handler、writer／consumer 與正式選單交易已閉合。
`rec169` 特別保留原始比較方向；雖然 `0–49` 與 `50–99` 都是 50%，不可把機率等價
誤寫成相同 RNG 邊界。

## Remake 接線

- game pack 使用有限 primitive：`cure_condition`、`wake_party`、`revive`、
  `remove_curse`；MP、scope、formula、condition 與成功下限均在 `spells.json`。
- `rec166/167/169/170/171` 走正式單體目標頁；`rec168` 直接完成施法。
- 解除詛咒在 Go 的 inventory／equipment 分離模型中等價為：裝備欄清空、同一 raw item
  放回該角色物品清單，不憑空刪除物品。
- 未新增 DQ3 專屬 production Go table 或文字 fallback。

## 停止線

這批不宣稱逐 frame／PCM wall-clock V3。音訊硬體時序依 wiki／平台規格近似，不再深挖
DAC／PIT。完成條件是正常 field spell 選單、MP 交易、目標 gate 與狀態副作用的 D3／E2。

## 2026-10-05 rec171 物品 word 勘誤

上表「第一件 equipped cursed item」及分離背包的接線敘述已推翻，原文保留。
同一輸入 EXE 的 IDA Pro 9.4 linear `0x1cd34` 原始 bytes `a90040` 只判斷詛咒位，
沒有要求穿戴位。linear `0x1cd5d` 的 `8124ff00` 保留低位物品碼，清除全部高位狀態，
不是只清穿戴標記。物品仍留在原物理格，不搬到另一份集合。

此兩項 writer 語意為限定 `confirmed`，沿用 `work/field-support-ida.json` 的完整
輸入 hash 與 linear 位址基準。原始證據、A READY 及實作入口見
[docs/188](188-opening-escort-to-castle-spec.md)。特殊狀態副作用沿既有契約，本勘誤不擴大
正常原版施法路線或動畫 V3 的驗收聲明。
