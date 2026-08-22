# 159 — 隊伍個人物品補給的正式原野使用規格（2026-08-22）

## 問題與證據邊界

`docs/156..158` 將商店、戰鬥掉落及 rec421 原野道具改回每名角色八格。最新正式 trace
在巴哈拉塔買八瓶聖水時，商店會依容量把物品分散到同伴；但
`traceAdventureTravelToCty` 仍以 `countItem/hasItem` 只查勇者。它因此把「勇者沒有聖水」
誤判成「全隊沒有聖水」，沒有走已接線的 owner selector，最後在 CTY16 後遇 monster51
全滅。失敗訊息中的 `holyWater=0` 也只印勇者數量，不能證明隊伍庫存為零。

本切片不新增原版規則，也不修改怪物、encounter、RNG、聖水效果或容量；只讓正式 trace
消費已由 IDA Pro 9.4 證實的持有者模型。

## 原版與既有 RE 證據

- 輸入 `assets_raw/DQ3.EXE`：115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- 工具：IDA Pro 9.4＋IDAPython；位址基準與 exporter 見 `docs/158`／
  `tools/ida_dump_field_item_owner_actions.py`。
- `docs/158` confirmed：多人 rec421 先由 party selector 選 owner，`DS:062D` 保存持有者，
  `sub_13919` 從該角色 `+0x3a` 八格取物；因此同伴持有的可用道具本來就能由正常玩家使用。
- `docs/152` confirmed：monster51 的 action3、傷害與持久麻痺；目前沒有新證據推翻該規則。

證據等級：原版 owner selector 與同伴物品使用入口為 **confirmed**；「最新全滅是 trace
庫存查詢不對稱」由購買 writer、實際 party inventory 與 helper source data flow 直接閉合，
屬 remake **confirmed**。航線採何種最佳補給數量仍只是測試玩家策略，不冒稱原版 AI。

## Remake／trace 規格

1. 高危航線判斷是否仍有聖水時必須使用 `countPartyItem`，不能只看勇者。
2. 有任何角色持有聖水時，必須呼叫正式 `traceUseInventoryItem`；該 helper 經 rec421 owner
   selector 選實際持有者，不可直接移除同伴 slice。
3. 特黑洛斯、保留魯拉 MP 與逃跑 fallback 的既有順序不變。
4. 失敗診斷同時列出 party 聖水總數；不得再以 hero-only 數量寫成「已耗盡」。
5. 八頭大蛇第一戰後前往 CTY22 教會／旅店的固定回復航段屬明確高危路線；若只剩最後
   一瓶聖水，該呼叫仍須以正式 rec421 使用，不得為未定義的更後路段保留而在當下全滅。

## 驗收

- component：勇者無聖水、同伴有聖水時，正式 owner selector 能使用並只消耗同伴欄位。
- production：從標題重播；至少必須越過目前 CTY16／monster51 blocker。若出現下一個失敗，
  記錄第一個玩家狀態並另立 spec，不修改怪物或 RNG。

## 重播勘誤

首次接線已越過 CTY16、加爾那之塔、達瑪與八頭大蛇第一戰。下一個失敗是第一戰後返回
CTY22 時仍持有一瓶聖水，但該航段未要求 `explicitRepel`，helper 因歷史「保留最後一瓶」
策略跳過使用而遭 monster51×2 全滅。這證明第五條必須在該固定回復航段啟用；不改共用
預設，也不把聖水直接消耗。
