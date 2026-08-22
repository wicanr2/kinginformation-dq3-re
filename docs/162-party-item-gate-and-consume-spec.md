# 162 — 劇情道具的隊伍 gate 與一次消耗規格（2026-08-22）

## 問題

`docs/161` 依原版共用 writer 讓新取得物品落入隊伍第一個空位。production 仍有門、NPC
交換、引導通道、祭壇及合成以 hero-only `hasItem/removeItems` 查詢；同伴合法持有時會被
誤判缺物，或消耗錯角色。

## 證據與範圍

- `docs/158` confirmed：`DS:062D` 保存物品持有者，rec421 使用／給予／丟棄均操作該角色。
- `docs/161` confirmed：一般劇情 grant 與戰鬥掉落共享 `sub_16856`，物品可合法落入同伴。
- 各具名事件的 required item、消耗與旗標交易已分別由 `docs/75`、`89`、`93`、`98`、
  `103` 等垂直切片閉合；本規格只修共同 ownership，不改 raw item、座標或旗標。
- 原始輸入與 IDA 9.4 位址契約沿用 `docs/161`。沒有逐一證實的泛用 helper 維持 fail closed，
  不由本規格新增事件。

推論等級：隊伍可持有、owner-local 使用及共用 grant 為 **confirmed**；具名事件是否消耗
沿用各自 D3 文件。搜尋順序採隊長至同伴，與 `sub_16856`／現行 `grantPartyItem` 一致。

## 規格

1. 已閉合的 required-item gate 使用 `hasPartyItem`。
2. 消耗一件時依隊長至同伴找到第一個匹配 slot，只刪一次；不得先刪勇者後又刪同伴。
3. 鑰匙 tier、六珠祭壇、太陽之石＋雲雨之杖合成、引導通道及有限 NPC 交換均接受合法
   同伴持有者。
4. 原野 rec421 的確切選定 owner 仍用 `consumeSelectedItem`；事件 gate 的隊伍搜尋不可取代
   玩家在道具面板的明確選擇。
5. component tests 鎖同伴持有 gate、一次消耗與不重複刪除；正式 trace 不得依賴搬回勇者。

## 實作訂正

第一次接線只把 `talkQuestItemChain` 的 required-item gate 改成 `hasPartyItem`，但成功分支仍
只掃 `g.inventory`。結果是隊友持有夢幻紅寶石時會顯示成功文字，卻未執行原格
`0x59 → 0x5a`。現行實作統一呼叫 `replacePartyItem`，依隊長至同伴尋找第一個匹配 slot；
元件測試另鎖定隊友原格替換，避免只驗證主角背包而再次產生假通過。

同一輪正式 trace 取得魔法鑰匙後又發現驗收 helper `traceOpenReachableDoor` 仍只用
`hasItem` 搜尋鑰匙，與 production `keyTier()` 的全隊 gate 不一致。helper 現改以
`hasPartyItem` 選擇最低合法 tier，再由 `traceUseInventoryItem` 的 owner selector 正式使用；
開門後亦以全隊持有驗證鑰匙沒有被消耗。這只修正玩家輸入重播，不改原版門 tier 或物品資料。

乾渴壺與最終鑰匙的 production assertion 也仍以 `hasItem` 假設勇者持有；共用 writer
接回後，present flag 已正確清除但物品可能在同伴。取得、save/load 與乾渴壺不消耗使用
的驗收均改成 `hasPartyItem`；正式使用仍由 `traceUseInventoryItem` 找實際 owner，不把
全隊 gate 混成主角 ownership。
