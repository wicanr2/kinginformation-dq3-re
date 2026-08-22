# 怪物 action runtime 規格（2026-08-23）

## 結論

原始 `D3MNS.DAT` 130 筆怪物資料實際使用 39 個 mask bit。現行 game pack 已逐 bit 提供
有限的 action definition，`Battle` 不再以歷史 `MonsterSpellRec` 順序表猜效果；測試會掃過
全部原始怪物 mask，任何實際使用但未定義的 bit 都失敗。

這個切片完成的是靜態 handler→runtime primitive 的 **D2／E2** 閉合，不宣稱戰鬥文字、
抗性、反射的逐訊息動畫或 RNG rejection sampling 已達 V3。玩家可見 timing 依既定政策為
V3-approx；PCM/DAC/PIT wall-clock 直接採公開硬體規格，不重開遊戲 driver 深挖。

## 證據契約

- `DQ3.EXE`：115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4，image `ida-pro-9.4-ver3:py312-x11-v4`；原始檔唯讀，database 僅在
  一次性容器 `/tmp`。
- exporter：`tools/ida_dump_monster_action_handlers.py`；sidecar
  `work/monster-actions-ida.json`（gitignored）。
- 位址：IDA linear；本文件的 `0x19xxx`／`0x1axxx` 均保留 IDA linear 位址。
- `sub_19AD6` 依 mask 選 bit；`sub_199DC` 對 bit16/17 加 2，bit18..47 再查
  `DGROUP 0x3930`；action `0x14..0x3b` 由 `DGROUP 0x394e` near-pointer table 分派。
- exporter 另列所有 `[si+38h]`、`[di+12h]`、`[si+4dh]` 與 active enemy 高位狀態的
  間接 operand。這些是候選索引，不是假 xref；每個 production 語意仍需 writer 與 consumer。

## 已接入的有限 primitive

| mask bit | action | runtime primitive | 原版 consumer 摘要 |
|---:|---:|---|---|
| 0,1,2,4,8..13 | 0,1,2,4,8..13 | `descriptor_damage` | `sub_19CB5` 先做每人 `rng<0xdc` gate，再依 3-byte descriptor 扣 MP、單體／全體傷害 |
| 3 | 3 | `special_physical_condition` | 無視守備物攻；存活者寫持久麻痺 |
| 16,17 | 18,19 | `instant_death` | 每人 `rng<0xdc` 後，`base=0xff` 進單體／全體死亡 writer |
| 18 | 20 | `sacrifice_death` | 同一 `rng<0xdc` 全體死亡 consumer 後，`sub_19D35` 將施術怪 HP 歸零 |
| 19,38,41 | 23,50,53 | `apply_condition` | 全隊睡眠、毒、睡眠吐息，各自保留原始比較邊界 |
| 22,24 | 26,28 | `scale_party_stat` | 攻擊／防禦第一次半減、第二次四分之一 |
| 23 | 27 | `banish_companion` | `sub_19EC0` 排除勇者，縮短現役隊伍；戰後搬回酒館 roster |
| 25,44 | 29,56 | `drain_mp` | 1..10 轉移給施術怪；3..9 只消除目標 MP |
| 26,29,31 | 31,35,37 | `apply_actor_status` | 混亂改打我方、扣 MP 後咒文無效、HP 回復無效 |
| 28 | 34 | `scale_enemy_stat` | 敵全體敏捷加半，active +0x12 bit0x200 防止重複 |
| 30 | 36 | `reflect_enemy` | active enemy 高位狀態由反射 consumer 讀取 |
| 32..35 | 40..43 | `heal_enemy` | 單體三階與敵全體回復 descriptor |
| 36,37 | 48,49 | `revive_enemy` | 一個已移除敵人以半 HP／全 HP 回場 |
| 39,40,42,47 | 51,52,54,59 | `breath_damage` | 共用 `sub_1A38C` 的 6..13、29..40、80..99 band |
| 45 | 57 | `summon_clone` | 同種敵人增援，受八單位上限約束 |
| 46 | 58 | `summon_related` | `sub_1A51A` 讀 D3MNS +0x20 關聯 raw id 後建 active enemy |

## 狀態 consumer 閉環

- `+0x38 bit0x400`：writer `sub_1A026`；consumer `sub_1B4F6` 忽略原命令、重抽存活我方
  目標並做物理攻擊，故命名為 battle confusion，等級 `confirmed`。
- `+0x38 bit0x200`：writer `sub_1A0FD` 只標記死亡角色；consumer `sub_1CF29` 已先扣 MP，
  再顯示無效並結束咒文，故 runtime primitive 為 `spell_null`，等級 `confirmed`。
- `+0x38 bit0x100`：writer `sub_1A15A`；consumer `sub_1D192` 在 HP recovery 前直接走無效
  分支，故 runtime primitive 為 `heal_block`，等級 `confirmed`。
- active enemy `0xc000`：writer `sub_1A146`；`sub_19BF7`／`sub_1D433` 旋轉並消耗反射
  狀態。現行 renderer／訊息為近似，不能由此宣稱逐訊息 V3。

## 驗收與停止線

- `TestEveryUsedD3MNSActionBitHasPackDefinition`：掃 130 筆 raw mask，鎖定 39/39 definitions。
- 怪物 MP、群傷、吸 MP、攻防削弱、banish→roster、敏捷一次 gate 皆有 component test。
- 未列入原始 D3MNS 的 action 不為了「表格完整」實作；未來只有新玩家路徑證明會用到時
  才重新開啟。
- 本切片不把動畫 frame、每一筆文字 record、PCM wall-clock 或全抗性矩陣冒稱為 exact。
