# 商店賣出、售價與裝備交易 runtime 規格

狀態：2026-08-23；一般武防店／道具店賣出為 D3／E2。本文件原先保留的瑪依拉
`DGROUP 0x0b62=1` 停止線，已由後續 [`docs/184`](184-maira-kings-sword-special-shop.md)
以 handler72／flag `0x134/0x135` 完成，不再是 current 待辦。

## 證據

- `DQ3.EXE`：115282 bytes，SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4；16-bit DOS linear address；原檔唯讀、原始名稱保留。
- 可重生 exporter：`tools/ida_dump_shop_sell.py`；sidecar 不入 Git。

`0x1776f..0x1793a` 的賣出鏈已證實：

1. 選隊員後列出該角色的八個 item words；選定後保存原始 slot pointer。
2. 一般售價讀 `ITEM +2/+3`，`0x17873 shr ax,1` 保存二分之一，再次右移得四分之一，
   相加後寫 `DGROUP 0x2593`，也就是逐步整數截斷的 `price/2 + price/4`。
3. 原價 0 在 `0x1786b..0x17870` 直接拒絕，不進確認。
4. 玩家確認後，`0x1789a..0x1789e` 才把選中 item word 寫 `0x00ff`，接著
   `sub_1895C` 加錢；取消不改物品、不加錢。
5. item word 可含裝備標記，所以裝備中物品也走同一 slot transaction；remake 的分離模型
   以清 equipment slot 表示，能力值由既有即時計算自然更新。

推論等級：`confirmed`。原始 bytes parity test 鎖定售價與 remove-before-credit 序列。

## 特殊 selector 的停止線

全程式 operand audit 只找到 `DGROUP 0x0b62` writer：`0x16315` 設 1，呼叫
`sub_17034(BX=0)` 後於 `0x1632d` 清 0。只有這個 scripted facility mode 才令 raw `0x6d`
在 `0x1784c..0x1785e` 取得固定 `0x57e4`（22500）售價。pack 已保存此 override，且以
`shop_kind=special` 避免污染所有一般道具店。這是當時的正確停止線；後續入口與王者之劍
交易已由 `docs/184` 閉合，保留本段是為說明一般店切片為何沒有提前猜補。

## Remake 接線

- 售價 ratio、零價 gate 與特殊 override 位於 `facilities.json`；共用 Go 只執行有限公式。
- 正常店鋪貨架按取消進入賣出選人，再選個人物品並二次確認；取消任一頁不交易。
- 個人未裝備物品與四個裝備槽皆可出售；成功才移除並加錢。
- 玩家可見名稱仍取原始 D3TXT00 item record，沒有在 production Go 新增文字。

本切片不升格逐窗幾何 V3。瑪依拉特殊 NPC 的獨立原版與玩家輸入驗收見 `docs/184`。

## 2026-10-05：同伴裝備出售支路修正 READY

Issue #4物品寫入稽核發現，`equipActorSlots`對主角回傳實際裝備array，對同伴回傳新array副本；`sellShopItem`的裝備分支直接清該array，再加錢。元件重現同伴武器仍為1、金錢卻增加22，主角支路通過。原版bytes parity仍通過，證明舊測試未涵蓋同伴寫入；不能把主角通過外推全隊。

本輪沿既有原版規則，沒有改資料格式。有限READY修正只將裝備清除交給既有`setEquipActorSlot(actor, slot, -1)`，寫實際持有者；原始IDA linear1789A..1789E選中slot清00FF後才加錢的契約保持。零價／取消不交易、正常售價與特殊店gate不變，沒有新增raw ID、文字或game-pack fallback。

驗收以主角／同伴兩支元件、另一持有者保持、原版bytes及正常新遊戲路線為準。[正式出售trace](../dq3_remake_ebitan/game/shop_sale_input_trace_test.go)共用既有新遊戲至羅馬利亞換裝的InputState前綴，再由正常facility NPC開店、取消貨架、選同伴／穿戴盾、取消確認及真正確認。需要只清所選裝備、加一次錢、防禦更新、存讀檔與正常下一步；runtime PNG只標remake流程證據，原版商店同狀態動態畫面未知，不宣稱V3。

`traceOpeningProductionInputRoute`的nil callback維持完整THE END測試；本切片專用callback只在正常抵達換裝checkpoint後執行並結束限定測試，不注入人物、座標、物品或故事旗標。八格順序資料方案另在docs/188待使用者選擇，不因本次setter修正而確認該方案。

### 有限 CONFORMED 與驗證限制

修正前正常路線r5在預定交易檢查點失敗：盾牌code58仍穿戴，金錢41218已達預期。修正後最終r5的5頂層／2子PASS，零SKIP／OOM；取消前後持久snapshot與PRNG相同，確認只移除同伴盾、售款135只加一次，另一持有者保持，防禦85→78。新遊戲seed1357初始化一次，交易前PRNG49849保持；沒有重擲、座標或裝備注入。

存檔由目前玩家狀態寫出，新Game的Load及目前玩家Load均完整保持，正式方向輸入可走到下一格。此處engine Save／Load是交易保存驗證；F5／F6正式UI仍由既有獨立正常路線覆蓋，不能把直接Save／Load稱為本切片的F5／F6輸入對拍。

| 本機收據 | SHA-256／範圍 |
| --- | --- |
| `work/issue4-companion-sale-red-r5/validation.json` | `fef3927b1650bfa415eb9a61136820110ab9009ca0bd3045a3c1ddf21dec5558`，指定產品錯誤重現 |
| `work/issue4-companion-sale-green-r5/validation.json` | `32187900f7863d1d65ec697d9a52f3bc2c5c043beb53d9be7d0819b666c0cad9`，修正後元件／正常交易／存讀檔／下一步 |
| `work/issue4-companion-sale-green-r5/sale-input/receipt.json` | `e77e3756ebf4b8edc2501d7b5ea01db88b20854459cfa7b40631795aee4b1e97`，交易前seed、所選物品、售價及防禦變化 |

五張runtime PNG保存確認、取消、售後、讀回及下一步，沿正式renderer，不重畫原版UI。確認／取消兩張相同，目視仍為黑底空窗，未顯示清單文字；本切片不將它們列為畫面通過，原因及原版同狀態畫面仍待查，不能宣稱商店V3。早期測試只在截圖時啟用frame，底圖錯用標題；最終r5在開店前啟用持續繪圖，舊r4及之前的PNG不列入本輪畫面驗收。

驗證腳本的失敗保留：首輪API名稱錯誤，後續Go map輸出旗標順序造成snapshot假差異，只排序snapshot複本的旗標集合；新Game.Load不切換標題狀態，下一步改由目前玩家實例讀回；撞櫃台也會重設cooldown，每次換方向前用正常空白幀等候。這些修正不改production時序。remake持有權E2與正常交易路線E3有限閉合，原版商店動態同狀態與完整campaign仍未知。

完整回歸收據為`work/issue4-companion-sale-full-r2/game-receipt.json`，SHA-256 `1536a8163c745e7509a35d7d342a8a686154bfb5cb48d21bbae427cb926b928f`。完整game485頂層覆蓋、489次執行，434不同頂層／121子PASS、51選用SKIP；internal171頂層／375子、11套件PASS及4選用SKIP。正常THE END163.93秒、Linux desktop PASS；逐項重跑零OOM，1082張既有PNG及5張新出售PNG保持。必驗出售零SKIP；最終五張PNG與targeted r5逐byte相同。獨立收尾`work/issue4-sale-final-audit-r1.json`保存來源hash、原版EXE身份、root基線3213及零.md目錄。首輪r1兩worker並行時一項招募測試exit-9，cgroup記錄oom4／oom_kill1；改為逐項、同image與3GiB資源乾淨重跑，r2全部通過，沒有為環境失敗修改遊戲。
