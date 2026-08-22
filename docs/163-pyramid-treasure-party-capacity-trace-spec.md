# 163 — 金字塔寶箱前的隊伍容量整理規格（2026-08-22）

## 問題

`docs/157`／`docs/161` 接回原版每人八格與共用取得物品 writer 後，正式主線在金字塔
魔法鑰匙寶箱正確停住：四名角色的裝備加未裝備物品都恰好占滿八格，寶箱 present flag
維持 set，沒有憑空覆蓋或越界新增。舊 trace 曾依賴勇者無上限 slice，並未真正驗證原版
背包容量下的玩家路線。

## 既有原版證據與推論等級

- `docs/157`：IDA Pro 9.4 已在 `sub_16856`（linear `0x16856`）證實依隊長至同伴搜尋
  第一個 `0x00ff` 空格；每人八格，全隊滿格時 `DS:0726=1` 且不寫入。**confirmed**。
- `docs/161`：寶箱／劇情取得物與戰鬥掉落共用同一 writer。**confirmed**。
- `docs/87`：CTY13 sec2 subid1 是 `{type=1,item=0x56,p2=0x73}`，正式調查取得魔法
  鑰匙，成功後 clear present flag。**confirmed**。
- 本輪 replay 的四人使用量均為八格；三名同伴仍持有多件藥草 `0x41`。這是 remake
  runtime observation，不提升任何新原版語意。

本切片沒有新的未知 executable 行為，因此不重新開啟 IDA 分析；直接消費上述已留存輸入
hash、位址、bytes、writer／consumer 閉環。若未來要實作原版「攜帶物太多」的完整姓名插值
訊息，仍須另行閉合 `docs/157` 所列 unknown，不能由本規格猜補。

## 規格

1. 引擎維持全隊滿格時不取得、不清 present flag；不得為了讓 campaign 通過而放寬容量。
2. production trace 在不可逆的新物品取得前檢查全隊是否至少有一格；若全滿，只能經正式
   命令窗、owner selector 與 rec421「丟掉」移除一件可消耗的藥草。第一個適用點是魔法
   鑰匙，下一個是波魯多加國王授予信件。
3. 不得直接改 inventory slice、清空裝備、注入鑰匙或修改寶箱旗標。
4. 整理後再以正式「調查」取得 `0x56`，並驗證物品在隊伍中、present flag clear 及
   save/load round-trip。

## 驗收

- 正式丟棄前後該持有者恰少一件 `0x41`，其餘角色與裝備不變。
- 取得後所有角色 `itemCount <= personal_inventory_slots`。
- `TestOpeningProductionInputTrace` 跨過魔法鑰匙 checkpoint，下一個失敗才列為 current
  campaign blocker。

## 重播追加發現

魔法鑰匙與依席斯／祠堂魔法門已通過後，全隊又因正常掉落填滿。波魯多加國王介紹結束時，
`grantPartyItem` 因無空格而保留 quest flag 且不授信；舊 trace 的 `traceCloseDialogue` 因而
持續重試。這不是狀態機或旗標錯誤。現將同一正式整理策略抽成 trace closure，在兩個已知
新物品取得點重用；引擎仍維持滿格不寫、不推進。

後續同一 replay 在巴哈拉達救援後又由正常戰鬥掉落填滿全隊；古布達的黑胡椒 choice
成功分支依 `docs/89`／`docs/161` 同樣必須先由共用 writer 找到空格。正式 trace 因此在
與 reward NPC 交談前套用相同整理策略，不改 `0x36` availability flag、選項或獎勵資料。

八頭大蛇後在 CTY38 補給八瓶聖水時，全隊初始只有六格；剩餘藥草只能再騰一格。固定八瓶
是 trace 的玩家策略上限，不是原版資料契約。大量補給現在先正式取消商店、逐次從實際
owner 丟棄可證實的藥草，再重新和同一設施 NPC 交易；若仍不足，就明確記錄容量限制並買到
`min(策略上限, 合法容量)`。後續路線若資源不足，必須在實際港口再正常補給，不能直接刪
slice、猜丟其他道具或放寬八格容量。不可逆劇情 grant 仍要求完整騰出所需格數。
