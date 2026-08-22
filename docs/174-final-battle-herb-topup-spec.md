# 174 — 終盤連戰藥草補給規格（2026-08-22）

## 玩家阻塞點

黃寶珠與六寶珠流程閉合後，正式主線在羅馬利亞終盤補給失敗：商店合法買入七株
藥草，但舊 trace 要求「本次額外買入八株」，因此在尚未進入巴拉摩斯城前中止。

這個斷言沒有把隊伍原先持有的藥草算入目標，也忽略原版每名角色的八格容量。
它是測試策略錯誤，不是商店 writer 或遊戲容量不足。

## 已讀取證據

- [`docs/156`](156-shop-personal-inventory-capacity-spec.md)：IDA Pro 9.4 已證實商店先由
  玩家選定角色，再掃該角色八格；只有成功寫入才扣款，滿格不自動轉送。
- [`docs/161`](161-common-party-item-grant-writer-spec.md)：其他新物品 writer 同樣受每人
  八格限制；正式 trace 不可新增無界背包。
- [`docs/146`](146-baramos-healing-resource-strategy-spec.md)：終盤勝利的決定條件是正式
  回復咒文／藥草策略可達，不要求每次補給額外買入固定數量。歷史完整 trace 曾以正常
  InputState 通過巴拉摩斯至 `THE END`。
- 本輪 fresh IDA Pro 9.4／IDAPython 複核的共同 writer sidecar：
  `/tmp/dq3-yellow-orb-ida-20260822/party-item-grant-callers.json`；輸入
  `assets_raw/DQ3.EXE` 115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`，位址為
  IDA linear。此 sidecar 保留原始名稱、call bytes 與推論等級，沒有修改 binary。

## 正式 trace 規格

1. 終盤補給目標是「全隊目前合計最多補到八株」，不是額外購買八株。
2. 使用既有 `topUpHerb(8)`：先計算 `countPartyItem`，只購買差額；若合法空格少於差額，
   只買可容納數量並留下 runtime ledger。
3. 購買仍經商品選擇、角色選擇、容量檢查與扣款的 production InputState；不直接增加物品。
4. 不丟劇情物品、不擴大容量、不注入戰鬥藥草或 HP／MP。
5. 若現有合法補給仍不足以通過戰鬥，應以第一個實際戰敗狀態另立規格，不把商店正常
   滿格行為改成產品缺陷。

## 驗收

- 完整 boot production trace 必須越過補給點；全隊任何角色仍不得超過 pack 容量。
- 後續戰鬥結果決定八株上限是否充分；單獨通過補給不宣稱終盤已完成。
- 本切片不改 production engine、game-pack、Boss 數值、AI、傷害、咒文或 RNG。
