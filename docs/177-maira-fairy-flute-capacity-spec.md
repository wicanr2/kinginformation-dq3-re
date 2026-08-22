# 177 — 瑪依拉妖精之笛容量交易規格（2026-08-22）

## 玩家阻塞點

太陽之石容量交易通過後，正式主線飛抵 CTY81。調查妖精之笛寶箱得到
`party item=false, present=true`；這是下一個真實玩家 blocker，與 type1 共同 writer
在全隊滿格時的失敗語意一致。

## RE 證據

- `events.json` 的 `dq3:event.maira_fairy_flute`：CTY81 section0 event0，type1、item
  `0x77`、present flag `0x9b`；證據記錄 raw tuple `{81,0,0,1,119,155}`。
- [`docs/re-log-722-state-machine.md`](re-log-722-state-machine.md)：妖精之笛 `0x77` 來源為
  CTY81 `(22,20)` type1 寶箱；後續在 CTY82 使用但不消耗，並取得精靈的守護 `0x74`。
- [`docs/31`](31-event-system.md)、[`docs/35`](35-script-format.md) 與 [`docs/161`](161-common-party-item-grant-writer-spec.md)：
  type1 場景物品由 `sub_18C0F` 呼叫共同全隊八格 writer，成功後才清 present flag。
- 本輪 fresh IDA Pro 9.4＋IDAPython sidecar：
  `/tmp/dq3-yellow-orb-ida-20260822/party-item-grant-callers.json`；輸入 DQ3.EXE
  115,282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`；位址為
  IDA linear，保留原始定位、bytes 與推論等級。

推論等級：寶箱 selector、共同 writer 與後續道具用途為 **confirmed**；本輪滿格結果為
**confirmed runtime**。未取得原版同狀態畫格，維持既有 V 等級。

## Production trace 規格

1. 從太陽之石合法狀態經正式 CTY transition、拉米亞飛行與入口抵達 CTY81。
2. 調查前經正式道具選單預留一格：藥草優先，其次只允許 ITEM.DAT 證明為裝備且未穿戴
   的備品。
3. 不丟劇情物品、不擴大容量、不直接改 inventory／flag。
4. 正常調查 type1 event；成功條件為全隊持有 `0x77` 且 flag `0x9b` clear，不限勇者。
5. 保留原持有者位置，後續 CTY82 妖精之笛 consumer 必須掃全隊並完成精靈守護交易。

第一次越過 CTY81 後，CTY82 runtime 得到 `guard=true, hero flute=false`。檢查
`usePackItemEffect` 證實 `consume:false` 路徑沒有呼叫 `consumeSelectedItem`；笛子是由寶箱
writer 合法寫在同伴，並在原持有者欄位保留。舊斷言只查勇者，已訂正為同時以
`hasPartyItem` 驗收妖精之笛與精靈守護；不得把這個 owner 差異誤記成原版會消耗笛子。

## 驗收界線

- 完整 boot production trace 越過 CTY81 並以正常路徑抵達下一阻塞點。
- 個人物品仍受 pack 八格限制；無安全可丟物時失敗即關閉。
- 本切片不修改 production 寶箱、妖精之笛效果、精靈守護或彩虹合成。
