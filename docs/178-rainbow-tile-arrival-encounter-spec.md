# 178 — 彩虹水滴目標格遭遇結算規格（2026-08-22）

## 玩家阻塞點

太陽之石、妖精之笛、精靈守護、雲雨之杖與彩虹合成都已由正式主線通過。船抵達
彩虹水滴使用座標 `(127,117)` 後，`traceUseInventoryItem` 沒有開到道具面板，最終得到
`panel=none, cursor=0`。

runtime 檢查顯示：抵達目標格的同一步耗盡遭遇計數器並開啟戰鬥。舊船路 helper 的迴圈
條件只看「尚未抵達」，因此在座標命中後立即返回，把接下來的命令窗輸入送入戰鬥 modal。
這是 production-input trace 的時序缺口，不是彩虹水滴、道具 owner 或 UI renderer 錯誤。

## 已讀取證據與界線

- [`docs/76`](76-r5-endgame-realignment.md)：IDA `loc_14243` 只接受下層 `(127,117)`，成功
  後改寫 `(126,117)` 橋 tile；不可換成直接座標注入。
- [`docs/31`](31-event-system.md)：彩虹合成 handler 與太陽之石／雲雨之杖 transaction 已閉合。
- 正式 trace 的 `traceSailToWorldCoordinate` 在每個尚未抵達的迴圈起點會處理戰鬥，但舊版
  抵達後沒有 post-arrival battle gate；本輪失敗為 **confirmed runtime**。
- 本輪 IDA Pro 9.4＋IDAPython 已以 DQ3.EXE SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c` 複核共同物品
  writer；本規格不據此猜測遭遇 RNG 的未解細節。逐步 encounter clock 的原版精確值不由
  本切片升格為 confirmed。

## Production trace 規格

1. 船路仍逐格送正式方向輸入，並只在原始可航水域規劃。
2. 座標命中後若 `battle.active`，必須先呼叫既有正式戰鬥輸入策略完成該場遭遇；不得清
   battle flag、改 RNG、改遭遇計數器或略過戰鬥。
3. 戰鬥結束後重新確認玩家仍在目標座標，才可返回 caller。
4. caller 再經正常命令窗與道具 owner selector 使用彩虹水滴。
5. 這個 post-arrival gate 屬船路 helper 的一般時序契約；沒有戰鬥時不改變任何行為。

## 驗收

- 完整 boot production trace 在 `(127,117)` 能先結算到達步遭遇，再正常使用彩虹水滴。
- 成功後 world state 與橋 tile 必須由 production item consumer 改寫，不能由測試直接設定。
- 本切片不修改 production 遭遇、戰鬥、道具或地圖規則。
