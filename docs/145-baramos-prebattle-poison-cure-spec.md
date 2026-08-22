# 巴拉摩斯戰前隊伍解毒策略（2026-08-22）

> 歷史 blocker：教會逐人解毒與後續路線已閉合，正式 trace 已由標題抵達 `THE END`；
> 下述失敗狀態只保存研究時間線，現況以 `docs/74` 最新 checkpoint 為準。

## 問題與直接 runtime 證據

完整 production-input trace 已在巴拉摩斯前由正式酒場招回第四人。進入 CTY66 時三名同伴
都存活，class5 Lv32 與 class4 Lv34 的永久咒文均包含 `rec175` 多拉瑪那；第一次施放可通過。
失敗發生在巴拉摩斯戰後由 CTY65 返回 CTY66 sec0、位置 `(14,12)` 的後續傷害地板路線：

- 勇者 HP `301`，場景 choices 仍有勇者的 `172／173／176`。
- 三名同伴 HP 全為 `0`，三者 `Conditions=1`，即現行持久狀態的 `conditionPoison`。
- class5／class4 的 `LearnedSpells` 仍含 `175`，所以不是習得或 save/roster 遺失。
- `fieldSpellCaster` 依正式規則排除死亡 caster，因此後續清單不列 `175`。

證據檔：`work/opening-trace-re175-state.log`（gitignored）。失敗當下的隊員／名冊摘要由
`traceUseFieldSpell` 自動輸出，不靠人工推測。

## 既有 RE 契約

- `docs/137-church-poison-condition-re.md`：教會 service0 的 poison condition reader、固定費用與
  清除 consumer 已閉合。
- `docs/78-field-spell-re.md`：`sub_1CB3C` 的 field caster 先檢查存活施法者，再依 descriptor
  扣 MP／派發 effect；`rec175` 是多拉瑪那，MP 2。
- `game/fieldspell.go` 的移動 consumer 只在成功移動後套 poison field damage；旅店已由
  `facility.go` 契約明定只恢復存活角色 HP／MP，不清 poison，也不復活。

推論等級：死亡 caster→`rec175` 不出現為 **confirmed**；三人未解毒是 **confirmed**；三人
究竟有多少 HP 由巴拉摩斯戰鬥、毒步與地形傷害各自扣除仍是 **unknown**。因此不能把「毒是
唯一死因」寫成事實，但在有教會的正式整備點離開中毒狀態，是玩家合法且必要的資源策略。

## Remake／正式追蹤規格

1. 新增通用 `traceCurePartyPoisonAtChurch`：只在主角或 active companions 有 poison 時開啟
   正式教會 NPC，逐一走 service0→target→確認，核對每筆 pack 固定費用與狀態清除。
2. helper 不得直接改 `Conditions`、HP、金錢、座標或 NPC 狀態；沒有中毒者時不得開教會。
3. 在巴拉摩斯前羅馬利亞 CTY2 的「正式酒場復隊→教會復活→旅店」鏈中，於住宿前完成全隊
   解毒；住宿只負責補滿存活者 HP／MP。
4. 完整 trace 必須證明：第一次及巴拉摩斯戰後需要時，仍有存活 caster 可由正式命令窗施放
   `rec175`，並繼續到下一個 production checkpoint。
5. 若解毒後仍全數死亡，保留本規格與 runtime ledger，再另開巴拉摩斯戰術／裝備窄切片；
   不以延長多拉瑪那、無視死亡 caster 或直接復活近似。

## 重播結果

教會解毒後，巴拉摩斯入場前三名同伴已全滿 HP，第一次 `rec175` 正常可用；但 Boss 戰內仍
全滅，當時 monster `0x79` 尚餘 `724/1200` HP。故 poison 是先前戰後死亡的重要資源缺口，
但不是本次 Boss 全滅的唯一原因；戰鬥策略移交 `docs/146-baramos-healing-resource-strategy-spec.md`。

## 驗收

- component：混合中毒／未中毒隊伍逐人清除、扣款總額正確、未中毒 no-op。
- production：`TestOpeningProductionInputTrace` 不可 SKIP，至少越過巴拉摩斯戰後 CTY66
  傷害地板；最終仍以從標題到 THE END 為 E3 gate。
