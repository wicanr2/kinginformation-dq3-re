# 尼羅肯特銀寶珠容量預留規格（2026-08-22）

## 新 blocker

野外補血咒文接線後，從新遊戲開始的正式 trace 已跨過幽靈船與奧莉維亞海岬，
但抵達 CTY64 handler49 時得到：`item=false, present=true`。這與 `docs/107`／`docs/161`
的原版 transaction 一致：補血改用 MP 後保留了較多藥草，使全隊八格 inventory 沒有
空槽；`grantPartyItem` 正確拒絕，並沒有 engine regression。

## 規格

1. 在第一次與 CTY64 `(16,10)` handler49 交談前，玩家策略必須預留至少一個全隊
   個人物品空槽。
2. 只能經正式道具面板的 rec421「丟掉」處理已證實可丟的項目：先丟藥草；若長路徑
   已耗盡藥草，再依 `docs/170` 丟 `ITEM.DAT EquipSlot>=0` 且不在 equipped slot 的備用
   裝備。不直接修改 inventory、不丟劇情道具、不覆蓋已穿戴裝備。
3. 若上述安全集合仍沒有可丟項目，trace 應在交談前失敗並列出容量，而不是修改
   `grantPartyItem`、放寬八格上限或提前清 flag。
4. 騰位後由正常 NPC 交談；item `0x6B` 寫入與 flag `0x4B` clear 必須同時成立，
   並通過 save/load 與 repeat dialogue。

## 證據等級

- handler49、全隊八格 writer、失敗旗標與成功後 clear：`confirmed`，見重新建立的
  `docs/107` 及 IDA sidecar。
- 本次容量成因：E3 trace 的實際失敗結果；屬玩家資源策略，不改變原版規則。
- 精確丟哪一株藥草／哪件備用裝備不是原版固定事件語意，只是 deterministic 驗收策略。
