# 建城前酒場寄放與寶珠持有權（2026-08-22）

> 歷史 blocker：寶珠持有權與正式復隊已由後續切片閉合，正式 trace 已由標題抵達
> `THE END`；下述失敗狀態只保存勘誤歷史，現況以 `docs/74` 最新 checkpoint 為準。

## 問題與被推翻的假說

黃寶珠後的正式 trace 發現綠寶珠 `0x66` 不在主角／現役隊伍，而在酒場 roster 的 class5
同伴背包；為重新招募該角色而返航時，殘破隊伍在 monster77×2 全滅。

先前文件把它歸因為「建城商人帶走寶珠」；這是錯誤。runtime ledger 顯示：

- roster class5 的 inventory 是 `[0x1e,0x01,0x1f,0x3a,0x66]`；綠寶珠在此。
- founder 是 class6，inventory 空；shared storage 只有 `[0x1e]`。
- 正式 trace 在紅寶珠後為登錄商人，明確以酒場 UI 寄放 `companions[0]`；該角色正是持有
  綠寶珠的 class5。這是玩家策略錯誤，不是 founder transaction。

## IDA 輸入與位址契約

- `assets_raw/DQ3.EXE`：115,282 bytes；SHA-256
  `5178fdc85021513392f6061451178121330a2a0282987c7cf4844187d9d7530c`。
- IDA Pro 9.4＋IDAPython；腳本 `tools/ida_dump_merchant_founder_inventory.py`。
- sidecar：`work/merchant-founder-inventory-ida.json`（不加入 Git）。
- 位址：`IDA linear = logical + 0x10000`；`file = logical + 0x1370`。

原始 EXE 唯讀；IDA database／工作副本只在一次性容器 `/tmp`。匯出保留原始函式、bytes、
xref、位址與推論等級，沒有 rename 或修改原始 binary。

## handler40 的 founder／物品資料流

handler40 為原始 `sub_15AA2`（IDA `0x15aa2..0x15bba`，file `0x6e12..0x6f2a`）：

1. IDA `0x15b3c..0x15b40` 由 active-party pointer table 取得選中的商人角色 record。
2. `0x15b55..0x15b6c` 將角色 record `+0x3a` 起八個 word item slot 逐一讀取；低 byte不是
   `0xff` 時呼叫 `sub_16CE2`。
3. `sub_16CE2`（IDA `0x16ce2..0x16d04`，file `0x8052..0x8074`）以 `CX=0x78` 掃
   DGROUP `0x526f`，把物品寫入第一個 `0xff`；滿載則顯示 record `0x135`。
4. `0x15b6f..0x15b7f` 將 18 bytes 商人身份複製到 DGROUP `0x5058`。
5. 後續清除 active-party slot／調整 party count，沒有把 founder 加入酒場 roster。

推論等級：**confirmed**（active-party reader→八槽 consumer→storage writer→founder identity
writer→party removal）。因此建城交易不會產生本次 roster 綠寶珠；先前斷言已推翻。

## Remake 規格

1. 不修改 production `settlement_founder`、shared storage、酒場 roster 或怪物資料。
2. 在紅寶珠後需要空出一格登錄商人時，trace 透過正式酒場「與同伴分離」清單，選擇第一名
   **未持有 `0x66..0x6b` 任一祭壇寶珠**的同伴，而非固定 `companions[0]`。
3. 若所有可寄放同伴都持珠，trace 必須失敗並記錄隊伍，不得直接搬移物品；後續需另行閉合
   正式「給予」UI。
4. 加 component test 鎖定持綠寶珠的 index0 不會被選中、無珠同伴可被選中、全員持珠時
   fail-closed。
5. 完整 trace 必須證明黃寶珠後六珠都在主角／active party，可直接前往 CTY70，不再為取回
   roster 物品返航阿里阿罕。

這是正式玩家輸入策略修正；不是把寶珠強制集中到主角，也不是繞過原版物品持有權。

## 第一次重播訂正

避開持珠同伴後，正式 trace 已不再進入返航阿里阿罕分支，證明六珠均留在主角／active
party；但巴拉摩斯外城需要正式施放 `rec175` 多拉瑪那時，現役 choices 只有
`rec172／173／176`。`rec176` 才是特黑洛斯，因此先前把 `rec175` 寫成特黑洛斯是錯誤，
已在此訂正。當時只有 `choices` 與名冊摘要，尚不足以證明 roster 成員就是缺少的施法者；
把兩者直接連結成因果是過度斷言，已撤回。必須以招募後每名角色的 class／level／HP／MP／
永久咒文 ledger 再判定。

## 第二次重播訂正

CTY38→CTY2 的教會復活分支只能恢復 active party 的死亡成員，不能把 roster 成員帶回隊伍；
先前宣稱它可解決 `rec175` 缺席也是錯誤。六珠放上 CTY70 祭壇後，寄放角色已不再持有流程
必要物品，因此正式解法是在巴拉摩斯戰前既有的羅馬利亞 CTY2 整備點：正常離城→魯拉
CTY0→正常入城→由露依達「找同伴參加」招回 `roster[0]`→正常離城→魯拉 CTY2→教會／旅店
恢復→繼續原路。這條路徑只使用 production `InputState`、已造訪目的地與酒場／設施 consumer；
若名冊為空、隊伍已滿或招募後人數／名冊不符，trace 必須失敗即關閉，不得直接修改
`companions`、`roster`、咒文或角色狀態。

## 第三次重播結果（待閉合）

上述 `roster[0]` 正式回隊已實作且 component test 通過，但完整 trace 仍在同一個
`rec175` gate 得到 `choices=[172,173,176]`。因此「只要招回 roster[0] 即可恢復多拉瑪那」
是否充分已被 runtime 結果否定。後續精確 ledger 現已收斂：進城時 class5／class4 兩名
存活同伴都具有 `rec175`；失敗實際發生在巴拉摩斯戰後，三名同伴 HP=0 且全員中毒，死亡
caster 被正式 consumer 排除。酒場持有權切片至此閉合；後續正式解毒策略移交
`docs/145-baramos-prebattle-poison-cure-spec.md`，不得再把它誤列為 roster／習得資料問題。

## 第四次訂正：復隊時點不能只由寶珠決定

後續由新遊戲正式重播到薩滿奧薩後證實：寄放時避開六顆祭壇寶珠雖能讓 CTY70 不必返航，
但被寄放者仍可能持有後續必要的暗黑神燈 `0x5f`。因此「無珠同伴可一直留到巴拉摩斯前」
是過度狹窄的玩家策略，不是原版保管語意；該策略已由 runtime consumer 反證。

現行規格改為：建城商人正式交付後隊伍立即出現空位，立刻經正常出口、魯拉、CTY0 正常
入口與露易達「找同伴參加」招回唯一 roster 成員。復隊必須同時鎖定四人隊伍、空 roster
與 active party 持有暗黑神燈，詳見 `docs/166-merchant-roster-required-item-reunion-spec.md`。
這不修改 `handler40`、不搬運或複製道具；只修正 production trace 的玩家操作時點。
