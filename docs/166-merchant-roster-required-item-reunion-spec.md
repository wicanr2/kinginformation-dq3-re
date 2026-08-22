# 建城後必要道具持有者立即復隊規格（2026-08-22）

## 玩家阻塞

正式主線在紅寶珠後，為登錄建城商人而於露易達酒館寄放一名不持祭壇寶珠的同伴。
商人交給建城老人後，隊伍由四人降為三人，但原同伴仍在酒館名冊。薩滿奧薩取得拉之鏡
後需要使用暗黑神燈 `0x5f` 切換夜晚；runtime trace 證實該道具由被寄放的同伴保管，現役
隊伍因而無法使用。

這不是 production 引擎遺失道具，也不是建城事件吞掉物品。`docs/144` 的 IDA Pro 9.4
證據已閉合 `handler40`：建城 consumer 只搬運被交付商人的八個個人物品槽到共用保管處，
原先寄放的同伴與其個人物品仍留在酒館 roster。僅排除「持寶珠者」的寄放策略，不能保證
所有後續必要道具都在現役隊伍，因此舊策略不充分。

## 證據契約

- 原始輸入：`assets_raw/DQ3.EXE`，115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 主要逆向證據：`docs/144-merchant-roster-orb-custody-spec.md`；IDA Pro 9.4＋IDAPython，
  原始 `sub_15AA2`（IDA linear `0x15aa2..0x15bba`）及 `sub_16CE2`
  （IDA linear `0x16ce2..0x16d04`）。
- 位址基準：`IDA linear = logical + 0x10000`；`file = logical + 0x1370`。
- 推論等級：**confirmed**，因 active-party item reader、shared-storage writer、角色移除與
  runtime roster custody 已形成 writer→state→consumer 閉環。
- 本切片不產生新原版語意，不需重跑全域 RE；只把既有已證實的個人物品保管語意套用到
  正式玩家策略。

## Remake／trace 規格

1. 建城商人交付並完成 save/load 後，先由正常城鎮出口離開。
2. 使用正式魯拉回已造訪的 CTY0，再由地表正常踏入城鎮。
3. 透過露易達「找同伴參加」選單招回唯一的 `roster[0]`；不得直接修改
   `companions`、`roster`、道具欄、MP 或座標。
4. 復隊後必須證明：隊伍恢復四人、roster 清空、暗黑神燈 `0x5f` 回到 active party。
5. 再由正常出口與魯拉返回 CTY38，銜接既有船運／薩滿奧薩流程。
6. 終盤既有復隊段保留失敗即關閉（fail-closed）條件，但 roster 已空時不得重複操作。
7. 完整 production trace 必須抵達暗黑神燈 consumer，且之後繼續至下一個真實玩家阻塞；
   direct state injection 不算驗收。

## 非目標

- 不把暗黑神燈複製或強制搬到勇者背包。
- 不改建城 handler、酒館資料結構或個人物品上限。
- 不以「合理」預設讓引擎在找不到 owner 時自動使用 roster 道具；未在現役隊伍即應失敗。
