# 巴拉摩斯戰回復資源策略（2026-08-22）

> 歷史 blocker：巴拉摩斯回復策略已閉合，正式 trace 已由標題抵達 `THE END`；下述
> 「新第一 blocker」只保存當時前沿，現況以 `docs/74` 最新 checkpoint 為準。

## 前一切片的重播結論

`docs/145` 的正式教會逐人解毒已通過既有 church component tests。完整 trace 在巴拉摩斯
入場前證明三名同伴全滿 HP、均存活且 class5／class4 都會 `rec175`；因此先前的毒步死亡
缺口已閉合。

新第一 blocker 是 monster `0x79` 巴拉摩斯戰內全滅：

- Boss 尚餘 `724/1200` HP；勇者與三名同伴 HP 全為 0。
- 勇者 Lv37；同伴約 Lv31–34，進戰前 HP 為 `214／202／223`。
- class5／class3 有正式回復咒文與充足 MP；class5 另須為戰後兩次多拉瑪那保留 4 MP。
- 現行 trace 的 `bestHealSpell` 無論敵人都選「MP 最低」的回復咒文。這是為下層連戰節流的
  測試策略，不是原版戰鬥規則；對單場高傷 Boss 會反覆使用低回復量咒文，無法覆蓋傷害。

證據：`work/opening-trace-poison-cure.log`。推論等級：全滅與剩餘 HP 為 **confirmed runtime**；
低回復量策略是資源不足的主要原因為 **strong inference**，需以重播驗證。

## 既有 RE／runtime 契約

- `docs/78-field-spell-re.md`：原版 `file 0xd4d9` 逐名收集命令，再依敏捷混排行動；正常輸入
  已接到每名角色自己的咒文、MP 與目標 phase。
- `traceResolveBattle` 只能選 production 戰鬥選單；不得直接改 HP／MP、Boss HP、RNG、回合、
  resistance 或行動結果。
- `availableMP` 對 monster `0x79` 的多拉瑪那 caster 扣除 4 MP 預算，這個戰後路線 gate
  保持不變。

## 正式追蹤規格

1. 保留一般遭遇與下層連戰的「最低 MP 回復」策略。
2. 只對 monster `0x79`，從該 actor 已習得且可負擔的 `spell.Heal` 中選 `Base` 最高者；可負擔
   判定仍使用扣除戰後 4 MP 預算後的 `availableMP`。
3. 回復目標、命令、咒文與 ally target 全部仍由 production InputState 選擇；不新增產品邏輯、
   JSON 數值或 Boss 特例到 engine。
4. component test 應鎖定「最低 MP」與「最高 Base」兩種選擇政策不混用；完整 trace 仲裁
   monster `0x79` 是否可勝且至少一名 `rec175` caster 存活離場。
5. 若仍全滅，下一切片才評估正式裝備／防禦／多 healer 目標分配；不得一次加入多個未驗證
   策略而失去因果辨識。

## 第一次重播訂正

最高 Base 策略的 component test 通過，但完整 trace 仍在完全相同的 Boss HP `724/1200`
全滅；因此「低回復量是主要原因」的 strong inference 已被結果否定。保留 selector 作為
monster `0x79` 的合法玩家策略，但它不是目前 blocker 的充分解。下一步先記錄每個 command
actor 的當前全隊 HP、選定 command、`wantedSpell` 與 `healTarget`，再判斷命令優先序、首回合
傷害或多人重複目標；未取得 ledger 前不新增防禦／裝備近似。

## 命令 ledger 與根因訂正

逐 actor ledger 證明 HP 降至 `69/202`、`116/223` 甚至更低時，`healTarget` 仍永遠是 `-1`；
回復咒文根本沒有進策略。程式直接原因是 `finalBoss` 下界錯設 `0x7a`：

- `usableHerbs` 只有 `finalBoss` 才非零。
- `finalBoss && canCastHeal` 才會掃低於 75% HP 的 ally。
- 同一 switch 內卻明確有 `g.battle.monID == 0x79` 的巴拉摩斯封咒、保留多拉瑪那與回復策略。

因此那些 `0x79` 策略在舊分類下不可達，這是 **confirmed code/runtime contradiction**。修正只把
deterministic trace 的終盤 Boss 範圍改成 `0x79..0x7c`；不更改 production battle、Boss 資料、
傷害、AI、RNG 或咒文。暫時逐命令 log 在取得證據後移除，精確輸出保留於
`work/opening-trace-baramos-command-ledger.log`。

## 最終重播結果

把終盤 Boss 下界訂正為 `0x79` 後，回復／藥草與既有封咒策略首次在巴拉摩斯戰正常可達；
完整 `TestOpeningProductionInputTrace` 已在 Docker＋Xvfb 從標題通過巴拉摩斯、下層世界、
索瑪與 `THE END`（188.68 秒，非 SKIP）。`TestTraceSelectHealSpellPolicy` 同批通過。

本切片只修 deterministic production-input trace 的合法玩家決策，不改任何 Boss 數值、AI、
傷害公式、RNG、game-pack 或 engine 戰鬥規則。主線 campaign 恢復 E3；逐動作 timing、動畫、
音效與同狀態畫面仍沿用各自 V 級，不由此推升為 V3。
