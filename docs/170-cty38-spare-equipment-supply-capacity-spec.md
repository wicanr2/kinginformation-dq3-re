# 幽靈船前 CTY38 備用裝備整理與補給容量規格（2026-08-22）

## 玩家阻塞與 runtime ledger

變化之杖交換後，四人皆達原版每人八格上限，CTY38 商店無法購入幽靈船／奧莉薇亞航線
所需補給。正式 trace 的 inventory／equipped 分離 ledger 為：

```text
inventory:
  hero  [0x55,0x1f,0x03,0x22,0x67,0x22]
  actor1[0x1e,0x14,0x69,0x5e,0x57,0x66]
  actor2[0x1e,0x45,0x68,0x61,0x42,0x63]
  actor3[0x22,0x51,0x56,0x5b,0x01,0x1f,0x3a,0x5f]
equipped:
  hero  [0x0d,0x28,-1,-1]
  actor1[0x01,0x1f,-1,-1]
  actor2[0x00,0x1e,-1,-1]
  actor3[-1,-1,-1,-1]
free=0, antidote=1, herb=0
```

`inventory` 與 `equipped` 是不同 storage；因此 inventory 中仍由 `ITEM.DAT` 分類為武器／
鎧／盾／兜的項目是未裝備備品，不是目前能力來源。大量劇情道具與六珠不能用「看似已用過」
猜測丟棄，故本策略只處理可由原始 ITEM 類別證明的備用裝備。

## 證據契約

- `docs/156-shop-personal-inventory-capacity-spec.md`：IDA Pro 9.4 已證實商店以角色八個 item
  word 作容量 gate；滿格不扣款、不寫入，推論等級 **confirmed**。
- `docs/158-field-item-owner-selection-drop-spec.md`：IDA Pro 9.4 已證實 rec421 owner 與正式
  丟棄操作，推論等級 **confirmed**。
- `docs/161-common-party-item-grant-writer-spec.md`：每人八格、裝備亦占格，推論等級
  **confirmed**。
- `ITEM.DAT` `+4/+5` 類別與 `EquipSlot` parity oracle 已由 `docs/13`／`docs/147` 使用；本切片
  只用它辨識 inventory 中的備用裝備，不替任何 raw ID 猜名稱或故事用途。

本切片沒有新的 executable 語意，不重跑 IDA；沿用上述已分級證據。玩家採用何種合法物品
管理不是原版唯一攻略，屬 deterministic production trace 策略。

## 規格

1. 在 CTY38 商店補給前先關閉 shop modal。
2. 先經正式 rec421 丟棄現有藥草；仍不足時，依隊長至同伴順序，只從各角色
   **unequipped inventory** 選 `ITEM.DAT EquipSlot>=0` 的備用裝備正式丟棄。
3. 不丟劇情道具、六珠、最終鑰匙或目前 equipped slots；不得直接 slice/delete。
4. 第一次重播以十格依序購入四瓶聖水、三份驅毒草與四份藥草，已正式通過幽靈船與
   奧莉薇亞海岬，但四份藥草不足以治療海岬後存活者；因此「十格足夠」已由 runtime
   反證，不再作現行策略。
5. 現行策略丟完所有可由 ITEM 類別證明的 unequipped 備用裝備；本航段聖水上限降為兩瓶，
   驅毒草仍補到三份，其餘容量購買藥草。兩瓶與藥草數量仍是 deterministic 策略，不冒稱
   原版固定需求；完整 replay 決定是否足夠。
6. 後續使用藥草／驅毒草均由 `traceUseInventoryItem` 選實際 owner；不得再以 hero-only
   `hasItem` 決定全隊是否還有補給。
7. 若三份驅毒草最低需求或後續玩家可見存活條件仍無法滿足，trace 失敗即關閉並另立下一
   blocker，不放寬容量或注入 HP／condition。

## 驗收

- 每次 drop、buy 都經正式 `InputState`，金錢與 owner inventory 由 production consumer 改變。
- 每名角色 `actorItemCount<=8`；equipped slots 原值不變；整理後至少十格可供本航段補給。
- 完整 campaign 通過 CTY38、幽靈船、奧莉薇亞解毒／治療路線，再以後續第一個真實 blocker
  決定是否需要調整數量。
