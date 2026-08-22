# 戰鬥道具 selector／owner／handler runtime 規格

> 狀態：2026-08-23。工具為 IDA Pro 9.4；輸入 `DQ3.EXE` 115282 bytes，
> SHA-256 `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；
> `ITEM.DAT` 896 bytes，SHA-256
> `7f3142de688ccca50fe888854b59ceb81b406b4c8eec038e719f10fda66e7f5d`。
> 本文位址未特別註明時是 IDA linear；原始名稱與相對運算元均保留。

## 已證實的垂直鏈

- `sub_1C1D8` 的 command 3 分支呼叫 `sub_1B836`。
- `sub_1B836` 呼叫 `sub_19834` 取得目前 actor record，從 `+0x3a` 掃描八個
  word；低 byte `0xff` 是空格。選定的原始格序寫入 `DGROUP 0x0d50`，目標寫入
  `DGROUP 0x0d4b`。這證實 owner 是目前角色，不是全隊共享藥草池。
- `DGROUP 0x3637` 的敵方 target 表是
  `05 0a 0c 10 14 17 18 1a 1c 46 ff`；`DGROUP 0x3642` 的我方 target 表是
  `41 42 44 45 4c ff`。其他項目以 actor 自身為預設目標。
- `sub_1B3F3` 讀每位 actor 的 command byte；action 3 進 `sub_1B95C`。
  後者重取 owner 與選定 slot，檢查 `ITEM.DAT` 載入區 `DGROUP 0x1f9` 的
  record `+4` bit `0x10`，再由 `DGROUP 0x3652`／`0x366a` 的 near function
  pointer 分派。pointer 必須以 callsite 的 CS offset 解讀，不能錯當 EXE load-base
  relative address。
- `sub_14CF9` 再次由 actor record `+0x3a` 找到選定 item，將該 word 寫成
  `0x00ff`。因此一般消耗品是「先消耗，再執行效果」；效果沒有命中也不能把物品補回。

## 已接入的 handler

| item | 原始 handler | runtime 結論 | 等級 |
|---|---|---|---|
| `0x41` | `sub_13DB9 → sub_14CF9 → sub_146EB` | 選定我方；先消耗；HP rejection sampling `30..39`，上限最大 HP | confirmed |
| `0x42` | `sub_13DC3 → sub_1469F` | 選定我方；先消耗；清原版 status mask `0x40` | confirmed |
| `0x43` | `sub_13DD0 → sub_14A8C` | 自身；先消耗；進共同返回 transition 並設 `DS:0x072e` | confirmed |
| `0x44` | `sub_13DDD` | 選定我方；先消耗；戰鬥分支只顯示 record `0x269`，不改 field repel | confirmed |
| `0x45` | `sub_13E15 → sub_1469F` | 選定我方；先消耗；清原版 status mask `0x10` | confirmed |
| `0x48` | `sub_13E40 → sub_1473F` | 自身；MP rejection sampling `22..31`；RNG byte `≤0x40` 才移除戒指 | confirmed |
| `0x4c` | `sub_13E82 → sub_1469F` | 選定我方；先消耗；清死亡 mask `0x80`、回最大 HP | confirmed |

上述版本專屬值置於 `battle.json.items[]`。Go 只保存 selector、有限 effect primitive、
逐 owner slot transaction 與戰後寫回。現行舊版 `heroHerbs/usedHerbs` 計數不再是正式
production transaction。

## 邊界與停止線

`0x05/0x0a/0x0c/0x10/0x14/0x17/0x18/0x1a/0x1c/0x46` 會進共同 spell-action
consumer；其 action index 的玩家可見 spell mapping 要與下一個「怪物 action」切片共用閉合，
不能由武器名稱或其他 DQ 版本猜。這些 item 仍會出現在原版八格 selector；在 mapping 完成前，
remake 以缺 definition 失敗即關閉，不把它們誤當藥草或假傷害。

可重生匯出腳本是 `tools/ida_dump_battle_item.py`。腳本只在一次性 IDA database 對已由
near pointer 證實的入口補建函式邊界，sidecar 明列
`analysis_added_from_confirmed_near_pointer`；原始 EXE、原始 symbol 與位址均未改寫。
