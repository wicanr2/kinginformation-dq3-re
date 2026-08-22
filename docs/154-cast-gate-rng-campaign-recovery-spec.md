# Cast gate RNG 勘誤後的正式練級策略（2026-08-22）

## 問題與證據邊界

[`docs/153`](153-monster-action-target-rng-order-re.md) 已依 IDA Pro 9.4 sidecar 訂正正常
怪物行動 RNG 為 `cast gate → target → action bit`。訂正後的第一次完整
`TestOpeningProductionInputTrace` 在羅馬利亞附近由 Lv15 練至 Lv25 的正式玩家
路徑中，遇到 monster15×3 全滅：

- 勇者 Lv15：`0/119 HP`、`MP=1`；
- 三名同伴皆死亡；
- 先前 CTY38→CTY70 monster77 是 gate／target 錯序版本的歷史結果。

該失敗不證明 monster15 資料或原版練級區錯誤。現行 trace 只以
`cty == 0` 或 `weakestLevel < 20 && monID >= 20` 判斷逃跑；它把 raw monster ID
當成戰力排序，與同檔已有「ID 不是強度」的安全判斷矛盾。這是測試玩家
策略缺口，不是 production 怪物近似的授權。

## 規格

1. 練級中的每場隨機戰仍由正式戰鬥命令處理，不改 HP、EXP、RNG、怪物或座標。
2. 是否作戰使用現有 `traceCanSafelyFinishEncounter` 的實際數值判斷：速度、攻擊、
   守備、敵人 HP，以及已閉合的持久狀態 action。不再比較 raw monster ID。
3. 判定不安全時必須從戰鬥選單送出「逃跑」輸入；判定安全才用普通
   攻擊取得 EXP。逃跑失敗與敵方行動仍由 production battle 處理。若逃跑失敗後
   有隊員低於 75% HP，非保留 MP 的練級戰必須先透過正式回復咒文或藥草救治，
   不可無限重試逃跑直到全滅。
4. 保留進城教會復活、旅店回復與正式出入城交易；不用 direct-entry 代替。
5. 羅馬利亞外 region 在當前隊伍 Lv15 可出現 monster15／16 三隻編隊，不得再被
   trace 稱為低危區。玩家已正式造訪 CTY0／CTY2 並恢復 MP；因此以正式魯拉
   回 CTY0 的已驗證低階 region 練級，完成後再正式魯拉回 CTY2 補給。這是可重播
   的玩家路線，不改 visited list、MP、座標或 region table。
6. 練級 helper 的城外休整門檻必須與戰鬥內正式回復策略一致：任一隊員
   低於或等於 75% HP 就返回教會／旅店。舊的 50% 門檻會讓隊伍在低 MP、無藥草、
   只剩略高於半血時踏入下一戰；monster6 全滅的 raw mask 為空，不支持新增怪物特殊行動。
7. 練級會正常推進 day/night clock。CTY1 道具店 NPC 有原版日夜 gate；離開 CTY0
   練級區前若非白天，要以現有 `traceWaitForDayNearCty` 的正式移動等待，不直接
   寫 clock／phase，也不在夜間冒稱 facility 缺失。
8. Lv25 練級完成不代表保留了旅行 MP。若下一步是魯拉，先以正式入城、
   旅店、出城交易回復；不改 MP 也不讓魯拉 helper 繞過原版成本。
9. 先重播完整 campaign。若出現新的第一個失敗，記錄其實際狀態後另開窄切片；
   不回退 `docs/153` 的 RNG 訂正。

## 驗收

- `traceCanSafelyFinishEncounter` 既有低數值高 ID 與不安全編隊 component tests 通過。
- monster17 bit38 鎖定 `gate → target → bit → per-member` RNG state。
- 完整 `TestOpeningProductionInputTrace` 必須由標題、正式 `InputState` 到達 `THE END`，
  才能重新恢復 current campaign E3。
