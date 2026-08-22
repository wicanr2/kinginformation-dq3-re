# CTY23 單人試煉回程：中毒補給與 monster bit32 窄證據（2026-08-22）

> 歷史 blocker：本切片已完成，後續正式 trace 已由標題抵達 `THE END`；現況以 `docs/74`
> 最新 checkpoint 為準，不得把下述中途失敗重新列為 current work。

## 問題

poison E2 接線後，完整 `TestOpeningProductionInputTrace` 在取得藍寶珠、save/load 後的
CTY23→CTY75 單人回程全滅。不能因舊 checkpoint 曾通過就關閉，也不能直接清 condition、
改 encounter counter 或降低怪物數值；先判斷是 remake 規則差異、補給缺口或 trace 尚未
操作正式 consumer。

## 輸入與工具

- 原始輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4＋IDAPython，Docker image `ida-pro-9.4-ver3:py312-x11-v2`。
- 原始 EXE 唯讀；IDA database 在容器 tmpfs。可重生匯出器為
  `tools/ida_dump_monster_bit32.py`，sidecar 為 `work/monster-bit32-ida.json`。
- 位址契約：IDA linear = logical + `0x10000`；main load image file = logical + `0x1370`。
  匯出保留原始名稱、bytes、xref 與推論等級，不 rename 或回寫 database／EXE。

## 證據 ledger

### confirmed：全滅前的 remake 玩家狀態

同一正式輸入 trace 可重現至 monster 1、七隻群組全滅；失敗輸出顯示：

- hero Lv32、`0/251 HP`；
- `heroConditions=0x1`，即 `conditionPoison`；
- 背包含多個 raw item `0x42`；
- `repel=0`，仍有一瓶聖水。

因此「沒有驅毒補給」已被反證。`docs/138` 已由原版 dispatcher／writer／consumer 證實
`0x42` 是可選隊員的驅毒草，而且確認後先消耗，再判斷並清除 poison。現行 runtime 已有
同一正式 modal；缺口是 production trace 在 save/load 後沒有操作它。

### confirmed：monster bit32 不是未 remap 的 raw spell ID

IDA `sub_199DC` 的 DGROUP `0x3930` 30-byte remap 表中，bit32 使用 index14，raw byte
`0x28`；再由 DGROUP `0x394e` handler table 的 IDA linear `0x28746` 讀 word `0xa1b4`，
進 logical `0xa1b4`／file `0xb524`。

該 handler window 讀 `DGROUP 0x37c4 + action*3` 的 base、`+0x37c5` 的 flags，經目標
準備後對 runtime target `+0x07` 增加回復量並以 `+0x09` 封頂，最後顯示 rec `0x151`。
這支持 monster 1 的 bit32 是回復 action，而不是毒氣或本輪中毒 writer。

推論邊界：remap、handler pointer、base/flags reader 與 HP capped writer 為
**confirmed**；完整敵方回復目標選擇語意尚未逐 caller 閉合，因此不在本輪修改 production。

## Remake／驗收 spec

1. 不改 monster 1 數值、群量、encounter RNG、逃跑率或 action handler。
2. 不替玩家新增道具；只使用先前正式商店流程已存在的 `0x42`。
3. CTY23 正常移動或戰鬥結束後，只要存活主角的持久 poison 已設且背包有驅毒草，
   production trace 必須在下一步移動前走：
   命令窗 → 道具 → `0x42` → 使用 → 選主角 → 關閉原版文字。
4. 使用後必須由 runtime consumer 清 poison 並消耗一件；沒有 poison 時不額外操作。
   藍寶珠 save/load 後仍重查同一條件，確保持久化不會繞過 consumer。
5. 接著才由正式 section transition、地表方向輸入與隨機戰鬥返回 CTY75；完整 trace 回綠
   前仍不得恢復現行 campaign E3 聲明。

## 停止線

- 本文件不宣稱 monster bit32 的所有目標選擇／訊息時序達 V3。
- 原版 per-character 八格 inventory 與現行 global inventory 的架構差距仍以 `docs/138`
  為準，不因這條 trace 使用成功而消失。
- 若解毒後仍出現新的第一個玩家 blocker，另開窄 ledger；不得把本 spec 擴成全域遭遇 RE。
