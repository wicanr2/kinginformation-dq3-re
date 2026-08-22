# 165 — 單人勇氣試煉前的主角空格保留規格（2026-08-22）

## 問題

蘭西爾 handler37 暫時移除三名同伴後，CTY23 藍寶珠只能寫入仍在 active party 的勇者。
現行正式路線進試煉時勇者裝備加未裝備物已滿八格，因此 `grantPartyItem` 正確不寫、寶箱
present flag `0xad` 保持 set。舊 trace 在接回個人容量前未暴露這個玩家資源前置條件。

## 證據與推論等級

- `docs/100`：handler37 保存 active party count、強制單人並在 handler62 還原；CTY23 event0
  是 item `0x67`／present flag `0xad`。**confirmed**。
- `docs/157`／`docs/161`：取得物只搜尋 active party 的每人八格；全滿不寫、不清 flag。
  **confirmed**。
- `docs/158`：試煉前仍為四人時，可經 rec421 把勇者未裝備物交給任一有空格同伴。
  **confirmed**。
- `beginSoloChallenge` 保存的是完整 companion records，復隊時原樣還原；這是 remake
  typed-state contract，另由既有 save/load tests 覆蓋。

本切片不新增 executable 語意；直接使用上述 IDA Pro 9.4 已閉合的原始位址、bytes 與
writer／consumer 證據，不猜 flag `0x13` writer。

## 規格

1. 接受試煉前，勇者至少保留一個個人物品空格。
2. 勇者已滿時，找有空格的同伴；若全隊滿格，只能先從同伴正式丟一株藥草。
3. 若同伴沒有空格，優先正式使用一瓶已購聖水，既提供洞窟防遭遇效果也釋放實際 owner
   一格；再經 owner selector／rec421／target selector，把勇者第一件不是三把鑰匙的
   未裝備物移交給該同伴。不得移走單人洞窟仍需使用的鑰匙。
4. 不直接改 inventory、暫存隊伍、寶箱旗標或容量；不猜丟棄其他 item 的合法性。
5. 進入單人後由正式調查取得 `0x67`，並鎖定途中 save/load、返回 handler62 與復隊。
6. 預期的離隊同伴快照必須在正式背包重分配後建立；整理前快照不能用來比較 handler37
   保存的 records。

## 驗收

- 接受前 `actorItemCount(hero) < personal_inventory_slots`。
- 取得後勇者持有 `0x67`、flag `0xad` clear，離隊同伴仍保持完整。
- 復隊後四人皆不超過個人容量，並可繼續下一個正式主線節點。
