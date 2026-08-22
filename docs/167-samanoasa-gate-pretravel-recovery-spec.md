# 薩滿奧薩關卡前正式恢復規格（2026-08-22）

## 玩家阻塞

建城後立即復隊使正式 campaign 能帶回暗黑神燈，但第一次抵達 CTY43、穿入 CTY44 後，
隊伍在走向 CTY44 北側地表出口途中全滅。精確 runtime 狀態為 CTY44 sec0 `(8,4)`、
`heroHP=0`、死亡＋中毒、`defeatDialogueStage=1`；先前只顯示 `dlg=true` 的路由錯誤訊息，
把產品已正確開啟的全滅 modal 誤看成 transition／NPC 阻塞。

## 既有 RE 證據

- 原始輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- `docs/155-field-poison-defeat-respawn-spec.md` 已用 IDA Pro 9.4＋IDAPython 閉合：
  - IDA linear `0x196AF..0x196C6`：地形／毒傷、HP 歸零與死亡 writer；
  - IDA linear `0x19621..0x19646`：存活人數、全滅 record361 與復歸 caller；
  - IDA linear `0x1C819..0x1C821`：復歸只恢復第一名角色。
- 位址基準：seg0 logical = IDA linear − `0x10000`；file = logical + `0x1370`。
- 推論等級：**confirmed**。本切片沒有新原版語意，不重跑全域 RE；只修正玩家補給策略。
- `docs/103-samanoasa-mirror-production-trace.md` 已證實 CTY43→CTY44→地表→CTY24 是正式路線，
  不得因全滅而更改 transition table 或繞過關卡。

## 正式玩家策略

1. 第一次由 CTY42 旅人之門抵達 CTY43 後，先在 CTY43 的正式旅店住宿，恢復存活者 HP／MP。
2. 由正常邊界離場，使用正式魯拉回已造訪的 CTY2。
3. 正常入城後經教會逐員復活、解毒，再住宿補滿；不得直接改 HP、condition 或位置。
4. 正常離城並以魯拉返回已造訪的 CTY43，再從正式入口進城。
5. 由原始跨 CTY portal 進 CTY44，沿既有路線前往 CTY24。
6. 導航 helper 的失敗訊息保留 `defeatDialogueStage`、HP 與 condition，避免日後再把全滅
   modal 誤報為地圖不連通。

這是 deterministic 驗收路線，不宣稱原版唯一攻略；不修改 encounter、毒傷、旅店、教會、
魯拉或敗北 consumer。
